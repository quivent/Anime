#!/usr/bin/env python3
"""wan-pipeline — stateful Wan 2.2 render pipeline with SQLite memory.

Commands:
    wan render "prompt" [opts]      Submit a render, store it
    wan history [-n N]              List recent renders
    wan show <id>                   Full record + URL
    wan resume <id>                 Re-render with same seed
    wan vary <id> [-n N]            Same prompt, new seeds
    wan rate <id> <1-5> [note]      Rate a render
    wan models                      List installed Wan models
    wan presets                     List render presets
    wan stats                       DB stats
    wan tui                         Interactive TUI

DB lives at ~/.anime/wan-pipeline.db
"""
import argparse, json, os, random, signal, sqlite3, sys, time, urllib.error, urllib.parse, urllib.request, uuid
from contextlib import contextmanager
from pathlib import Path

# ── color palette (no external deps) ──
B='\033[1m'; D='\033[2m'; R='\033[0m'
G='\033[38;5;42m'; C='\033[38;5;51m'; Y='\033[38;5;220m'; P='\033[38;5;213m'; X='\033[38;5;203m'

DB_PATH = Path(os.environ.get("WAN_DB", str(Path.home() / ".anime/wan-pipeline.db")))
DB_PATH.parent.mkdir(parents=True, exist_ok=True)
COMFY_API = os.environ.get("COMFY_API", "http://127.0.0.1:8188")
# Default view base = same origin as the API. Override with COMFY_VIEW_BASE
# only when ComfyUI is fronted by a separate public host.
COMFY_VIEW_BASE = os.environ.get("COMFY_VIEW_BASE", COMFY_API)

# ── presets baked in (today's tuned config) ──
PRESETS = {
    "t2v-14b-dual-maxq": {
        "kind": "t2v_dual",
        "model_high": "wan2.2_t2v_high_noise_14B_fp8_scaled.safetensors",
        "model_low":  "wan2.2_t2v_low_noise_14B_fp8_scaled.safetensors",
        "encoder":    "umt5_xxl_fp8_e4m3fn_scaled.safetensors",
        "vae":        "wan_2.1_vae.safetensors",
        "lora_high":  None, "lora_low": None,
        "width": 1280, "height": 720, "length": 121, "fps": 24,
        "steps": 50, "switch_at": 25, "cfg": 5.0, "shift": 8.0,
        "sampler": "uni_pc", "scheduler": "simple",
        "description": "Wan 2.2 14B dual-expert max quality (no LoRA, 1280x720, 50 steps, ~5min)",
    },
    "t2v-14b-dual-fast": {
        "kind": "t2v_dual",
        "model_high": "wan2.2_t2v_high_noise_14B_fp8_scaled.safetensors",
        "model_low":  "wan2.2_t2v_low_noise_14B_fp8_scaled.safetensors",
        "encoder":    "umt5_xxl_fp8_e4m3fn_scaled.safetensors",
        "vae":        "wan_2.1_vae.safetensors",
        "lora_high":  "wan2.2_t2v_lightx2v_4steps_lora_v1.1_high_noise.safetensors",
        "lora_low":   "wan2.2_t2v_lightx2v_4steps_lora_v1.1_low_noise.safetensors",
        "width": 832, "height": 480, "length": 81, "fps": 24,
        "steps": 8, "switch_at": 4, "cfg": 1.0, "shift": 8.0,
        "sampler": "uni_pc", "scheduler": "simple",
        "description": "Wan 2.2 14B + 4-step lightx2v LoRA (832x480, 4+4 steps, ~30s)",
    },
    "ti2v-5b": {
        "kind": "t2v_single",
        "model": "wan2.2_ti2v_5B_fp16.safetensors",
        "encoder": "umt5_xxl_fp8_e4m3fn_scaled.safetensors",
        "vae":     "wan2.2_vae.safetensors",
        "width": 832, "height": 480, "length": 81, "fps": 24,
        "steps": 20, "cfg": 5.0, "shift": 8.0,
        "sampler": "uni_pc", "scheduler": "simple",
        "description": "Wan 2.2 5B TI2V (fast iteration, 832x480, 20 steps, ~12s)",
    },
}
DEFAULT_PRESET = "t2v-14b-dual-fast"
# Standard SFW negative — strips NSFW content too, plus quality degraders.
# Keep as the default for unmarked renders.
DEFAULT_NEGATIVE_SFW = (
    "blurry, low quality, deformed, text, watermark, jpeg artifacts, oversaturated, "
    "cropped, partial body, choppy, film grain, noise, granular, "
    "nudity, nsfw, explicit, sexual, suggestive"
)
# "Explicit" negative — only quality/artifact suppressors, no content gating.
# Selected via --explicit (CLI) or the explicit toggle in TUI / studio.
DEFAULT_NEGATIVE_EXPLICIT = (
    "blurry, low quality, deformed, text, watermark, jpeg artifacts, oversaturated, "
    "cropped, partial body, choppy, film grain, noise, granular"
)
# Backwards-compatible alias (any old caller importing this still works).
DEFAULT_NEGATIVE = DEFAULT_NEGATIVE_SFW

# ──────────────────────────────────────────────────────────────────
# DB
# ──────────────────────────────────────────────────────────────────
SCHEMA = """
CREATE TABLE IF NOT EXISTS renders (
    id INTEGER PRIMARY KEY,
    created_at TEXT NOT NULL,
    name TEXT,
    prompt TEXT NOT NULL,
    negative TEXT,
    preset TEXT,
    params_json TEXT NOT NULL,
    seed INTEGER NOT NULL,
    output_path TEXT,
    output_url TEXT,
    file_size INTEGER,
    render_seconds REAL,
    parent_id INTEGER REFERENCES renders(id),
    rating INTEGER,
    notes TEXT,
    status TEXT NOT NULL DEFAULT 'pending'
);
CREATE INDEX IF NOT EXISTS idx_renders_created ON renders(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_renders_rating  ON renders(rating DESC);

CREATE TABLE IF NOT EXISTS installs (
    id INTEGER PRIMARY KEY,
    component TEXT NOT NULL,
    version TEXT,
    status TEXT NOT NULL,
    started_at TEXT NOT NULL,
    finished_at TEXT,
    notes TEXT
);
"""

@contextmanager
def db():
    conn = sqlite3.connect(DB_PATH)
    conn.row_factory = sqlite3.Row
    conn.executescript(SCHEMA)
    try:
        yield conn
        conn.commit()
    finally:
        conn.close()

# ──────────────────────────────────────────────────────────────────
# Workflow construction
# ──────────────────────────────────────────────────────────────────
def build_workflow(preset_name: str, prompt: str, negative: str, seed: int, name_prefix: str):
    p = PRESETS[preset_name]
    # Pick the right empty-latent node per model:
    #   • 14B (Wan 2.1/2.2 dual-expert) uses the 8x VAE → 16-channel latent at /8 spatial.
    #     Use EmptyHunyuanLatentVideo. Wan22ImageToVideoLatent here halves the output silently.
    #   • 5B TI2V uses the new 16x VAE → 48-channel latent at /16 spatial.
    #     Wan22ImageToVideoLatent is correct.
    if p["kind"] == "t2v_dual" or p.get("model", "").startswith("wan2.1"):
        latent_node = {"class_type": "EmptyHunyuanLatentVideo",
                       "inputs": {"width": p["width"], "height": p["height"],
                                  "length": p["length"], "batch_size": 1}}
    else:
        latent_node = {"class_type": "Wan22ImageToVideoLatent",
                       "inputs": {"width": p["width"], "height": p["height"], "length": p["length"],
                                  "batch_size": 1, "vae": ["vae", 0]}}

    g = {
        "clip": {"class_type": "CLIPLoader", "inputs": {"clip_name": p["encoder"], "type": "wan", "device": "default"}},
        "vae":  {"class_type": "VAELoader",  "inputs": {"vae_name": p["vae"]}},
        "pos":  {"class_type": "CLIPTextEncode", "inputs": {"text": prompt,   "clip": ["clip", 0]}},
        "neg":  {"class_type": "CLIPTextEncode", "inputs": {"text": negative, "clip": ["clip", 0]}},
        "lat":  latent_node,
    }

    if p["kind"] == "t2v_dual":
        g["high"] = {"class_type": "UNETLoader", "inputs": {"unet_name": p["model_high"], "weight_dtype": "default"}}
        g["low"]  = {"class_type": "UNETLoader", "inputs": {"unet_name": p["model_low"],  "weight_dtype": "default"}}
        h_in, l_in = ["high", 0], ["low", 0]
        if p.get("lora_high"):
            g["lora_h"] = {"class_type": "LoraLoaderModelOnly", "inputs": {"lora_name": p["lora_high"], "strength_model": 1.0, "model": ["high", 0]}}
            h_in = ["lora_h", 0]
        if p.get("lora_low"):
            g["lora_l"] = {"class_type": "LoraLoaderModelOnly", "inputs": {"lora_name": p["lora_low"], "strength_model": 1.0, "model": ["low", 0]}}
            l_in = ["lora_l", 0]
        g["ms_h"] = {"class_type": "ModelSamplingSD3", "inputs": {"shift": p["shift"], "model": h_in}}
        g["ms_l"] = {"class_type": "ModelSamplingSD3", "inputs": {"shift": p["shift"], "model": l_in}}
        g["ks_h"] = {"class_type": "KSamplerAdvanced", "inputs": {
            "add_noise": "enable", "noise_seed": seed, "control_after_generate": "fixed",
            "steps": p["steps"], "cfg": p["cfg"], "sampler_name": p["sampler"], "scheduler": p["scheduler"],
            "start_at_step": 0, "end_at_step": p["switch_at"], "return_with_leftover_noise": "enable",
            "model": ["ms_h", 0], "positive": ["pos", 0], "negative": ["neg", 0], "latent_image": ["lat", 0]}}
        g["ks_l"] = {"class_type": "KSamplerAdvanced", "inputs": {
            "add_noise": "disable", "noise_seed": seed, "control_after_generate": "fixed",
            "steps": p["steps"], "cfg": p["cfg"], "sampler_name": p["sampler"], "scheduler": p["scheduler"],
            "start_at_step": p["switch_at"], "end_at_step": 10000, "return_with_leftover_noise": "disable",
            "model": ["ms_l", 0], "positive": ["pos", 0], "negative": ["neg", 0], "latent_image": ["ks_h", 0]}}
        g["dec"] = {"class_type": "VAEDecode",   "inputs": {"samples": ["ks_l", 0], "vae": ["vae", 0]}}
    else:  # t2v_single
        g["model"] = {"class_type": "UNETLoader", "inputs": {"unet_name": p["model"], "weight_dtype": "default"}}
        g["ms"]    = {"class_type": "ModelSamplingSD3", "inputs": {"shift": p["shift"], "model": ["model", 0]}}
        g["ks"]    = {"class_type": "KSampler", "inputs": {
            "seed": seed, "steps": p["steps"], "cfg": p["cfg"],
            "sampler_name": p["sampler"], "scheduler": p["scheduler"], "denoise": 1.0,
            "model": ["ms", 0], "positive": ["pos", 0], "negative": ["neg", 0], "latent_image": ["lat", 0]}}
        g["dec"] = {"class_type": "VAEDecode",   "inputs": {"samples": ["ks", 0], "vae": ["vae", 0]}}

    g["vid"] = {"class_type": "CreateVideo", "inputs": {"images": ["dec", 0], "fps": p["fps"]}}
    g["save"] = {"class_type": "SaveVideo", "inputs": {"video": ["vid", 0], "filename_prefix": f"wan-pipeline/{name_prefix}", "format": "auto", "codec": "auto"}}
    return g

# ──────────────────────────────────────────────────────────────────
# Submit + wait
# ──────────────────────────────────────────────────────────────────
def submit_render(graph: dict) -> str:
    payload = json.dumps({"prompt": graph, "client_id": str(uuid.uuid4())}).encode()
    req = urllib.request.Request(f"{COMFY_API}/prompt", data=payload, headers={"Content-Type": "application/json"})
    try:
        r = json.load(urllib.request.urlopen(req, timeout=15))
    except urllib.error.URLError as e:
        raise RuntimeError(
            f"ComfyUI not reachable at {COMFY_API} ({e.reason}). "
            f"Start it with: anime comfyui start"
        ) from None
    if r.get("node_errors"):
        # Surface the first node error in human-readable form, not a Python repr.
        first = next(iter(r["node_errors"].items()))
        raise RuntimeError(f"workflow rejected by ComfyUI (node {first[0]}): {first[1]}")
    return r["prompt_id"]

def wait_for(prompt_id: str, timeout: int = 1800) -> dict:
    start = time.time()
    last = None
    while True:
        elapsed = time.time() - start
        try:
            h = json.load(urllib.request.urlopen(f"{COMFY_API}/history/{prompt_id}", timeout=10))
        except urllib.error.URLError as e:
            raise RuntimeError(f"ComfyUI history unreachable at {COMFY_API} ({e.reason})") from None
        if prompt_id in h:
            entry = h[prompt_id]
            status = (entry.get("status") or {}).get("status_str", "")
            if status == "error":
                # ComfyUI failed during execution — pull the first message for the user.
                msgs = (entry.get("status") or {}).get("messages", [])
                detail = next((m[1] for m in msgs if m and m[0] == "execution_error"), msgs)
                raise RuntimeError(f"render failed in ComfyUI: {detail}")
            return entry
        msg = f"  {C}…{R} rendering {elapsed:.0f}s"
        if msg != last:
            print(msg, flush=True); last = msg
        if elapsed > timeout:
            raise TimeoutError(f"render exceeded {timeout}s")
        time.sleep(5)

def extract_outputs(history: dict):
    out = []
    for nid, o in history.get("outputs", {}).items():
        for kind in ("videos", "images", "gifs"):
            for f in o.get(kind, []):
                qs = urllib.parse.urlencode({"filename": f["filename"], "subfolder": f.get("subfolder",""), "type": f.get("type","output")})
                local = Path.home() / "ComfyUI/output" / (f.get("subfolder","") + "/" if f.get("subfolder") else "") / f["filename"]
                out.append({
                    "filename": f["filename"],
                    "url": f"{COMFY_VIEW_BASE}/api/view?{qs}",
                    "local": str(local),
                    "kind": kind,
                })
    return out

# ──────────────────────────────────────────────────────────────────
# Commands
# ──────────────────────────────────────────────────────────────────
def cmd_render(args):
    preset = args.preset
    if preset not in PRESETS:
        print(f"{X}unknown preset:{R} {preset}\n  available: {', '.join(PRESETS)}"); sys.exit(1)
    seed = args.seed if args.seed is not None else random.randint(1, 2**63-1)
    name = args.name or f"render_{int(time.time())}"
    # --negative wins; otherwise pick SFW or explicit baseline.
    if args.negative is not None:
        negative = args.negative
    elif getattr(args, "explicit", False):
        negative = DEFAULT_NEGATIVE_EXPLICIT
    else:
        negative = DEFAULT_NEGATIVE_SFW
    graph = build_workflow(preset, args.prompt, negative, seed, name)

    with db() as conn:
        cur = conn.execute("""
            INSERT INTO renders(created_at, name, prompt, negative, preset, params_json, seed, parent_id, status)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'pending')
        """, (time.strftime("%Y-%m-%d %H:%M:%S"), name, args.prompt, negative, preset,
              json.dumps(PRESETS[preset]), seed, args.parent))
        rid = cur.lastrowid

    print(f"\n{B}{P}╭─ wan-pipeline render {rid}{R}")
    print(f"{P}│{R} {B}preset{R}  {preset}")
    print(f"{P}│{R} {B}seed{R}    {seed}")
    print(f"{P}│{R} {B}prompt{R}  {args.prompt[:80]}{'...' if len(args.prompt)>80 else ''}")
    print(f"{P}╰─ submitting...{R}\n")

    # Ctrl+C while this row is in flight should mark it 'cancelled', not leave
    # 'pending' forever. We register the handler narrowly so it can't outlive
    # the render and confuse later commands.
    def _on_sigint(signum, frame):
        try:
            with db() as conn:
                conn.execute(
                    "UPDATE renders SET status='cancelled', notes=COALESCE(notes,'')||' [SIGINT]' WHERE id=? AND status='pending'",
                    (rid,),
                )
        except Exception:
            pass
        print(f"\n  {Y}✗ cancelled (id={rid}){R}\n")
        sys.exit(130)
    prev_sigint = signal.signal(signal.SIGINT, _on_sigint)

    t0 = time.time()
    try:
        pid = submit_render(graph)
        print(f"  {C}prompt_id{R} {pid}")
        result = wait_for(pid, timeout=args.timeout)
        elapsed = time.time() - t0
        outs = extract_outputs(result)
        if not outs:
            raise RuntimeError("no outputs from render")
        primary = outs[0]
        size = Path(primary["local"]).stat().st_size if Path(primary["local"]).exists() else 0
        with db() as conn:
            conn.execute("""
                UPDATE renders SET status='done', output_path=?, output_url=?, file_size=?, render_seconds=? WHERE id=?
            """, (primary["local"], primary["url"], size, elapsed, rid))
        print(f"\n  {G}✓ done in {elapsed:.0f}s ({elapsed/60:.1f} min)  ·  {size/1024/1024:.1f}MB{R}")
        print(f"  {B}url:{R}   {primary['url']}")
        print(f"  {B}local:{R} {primary['local']}")
        print(f"  {B}id:{R}    {rid}\n")
    except Exception as e:
        with db() as conn:
            conn.execute("UPDATE renders SET status='failed', notes=? WHERE id=?", (str(e), rid))
        print(f"\n  {X}✗ failed:{R} {e}\n")
        sys.exit(2)
    finally:
        signal.signal(signal.SIGINT, prev_sigint)

def cmd_history(args):
    with db() as conn:
        rows = conn.execute("""
            SELECT id, created_at, name, preset, status, seed, render_seconds, file_size, rating,
                   prompt, output_url
            FROM renders ORDER BY id DESC LIMIT ?
        """, (args.n,)).fetchall()
    if args.json:
        out = [{k: r[k] for k in r.keys()} for r in rows]
        print(json.dumps(out))
        return
    if not rows:
        print(f"{D}(no renders yet — try: wan render \"a dragon\"){R}"); return
    print(f"\n{B}  id{R}  {B}date              status   preset                seed              t       size    ★  prompt{R}")
    print(f"{D}  ─── ────────────────  ──────── ────────────────────  ────────────────  ──────  ──────  ─  ──────────────────────{R}")
    for r in rows:
        st = r["status"] or "?"
        st_c = G if st=="done" else (Y if st=="pending" else X)
        rating = ("★"*r["rating"] + "·"*(5-r["rating"])) if r["rating"] else "·····"
        sz = f"{(r['file_size'] or 0)/1024/1024:.1f}M" if r["file_size"] else "—"
        t = f"{r['render_seconds']:.0f}s" if r["render_seconds"] else "—"
        prompt_short = (r["prompt"] or "")[:60]
        print(f"  {r['id']:>3}  {r['created_at'][:16]}  {st_c}{st:<8}{R} {(r['preset'] or '?'):<20}  {r['seed']:>16}  {t:>6}  {sz:>6}  {rating}  {prompt_short}")
    print()

def cmd_show(args):
    with db() as conn:
        r = conn.execute("SELECT * FROM renders WHERE id=?", (args.id,)).fetchone()
    if not r:
        print(f"{X}not found:{R} {args.id}"); sys.exit(1)
    print(f"\n{B}{P}render #{r['id']}{R}  ({r['status']})  {r['created_at']}")
    print(f"  {B}preset{R}      {r['preset']}")
    print(f"  {B}seed{R}        {r['seed']}")
    print(f"  {B}prompt{R}      {r['prompt']}")
    if r['negative']: print(f"  {B}negative{R}    {r['negative'][:120]}")
    if r['parent_id']: print(f"  {B}parent{R}      #{r['parent_id']}")
    if r['render_seconds']: print(f"  {B}duration{R}    {r['render_seconds']:.0f}s ({r['render_seconds']/60:.1f}min)")
    if r['file_size']: print(f"  {B}size{R}        {r['file_size']/1024/1024:.1f}MB")
    if r['rating']: print(f"  {B}rating{R}      {'★'*r['rating'] + '·'*(5-r['rating'])}")
    if r['output_url']: print(f"  {B}url{R}         {r['output_url']}")
    if r['output_path']: print(f"  {B}local{R}       {r['output_path']}")
    if r['notes']: print(f"  {B}notes{R}       {r['notes']}")
    p = json.loads(r['params_json'])
    print(f"  {B}params{R}      {p['width']}x{p['height']}  ·  {p['length']}f@{p['fps']}fps  ·  {p['steps']} steps  ·  cfg {p['cfg']}  ·  shift {p['shift']}")
    print()

def cmd_resume(args):
    """Re-render with the same seed (deterministic reproduce)."""
    with db() as conn:
        r = conn.execute("SELECT * FROM renders WHERE id=?", (args.id,)).fetchone()
    if not r: print(f"{X}not found:{R} {args.id}"); sys.exit(1)
    sub = argparse.Namespace(prompt=r['prompt'], negative=r['negative'], preset=r['preset'],
                             seed=r['seed'], name=f"resume_{r['id']}", parent=r['id'], timeout=args.timeout)
    print(f"{C}resuming render #{r['id']} with seed {r['seed']}{R}")
    cmd_render(sub)

def cmd_vary(args):
    """Same prompt + preset, fresh seeds."""
    with db() as conn:
        r = conn.execute("SELECT * FROM renders WHERE id=?", (args.id,)).fetchone()
    if not r: print(f"{X}not found:{R} {args.id}"); sys.exit(1)
    for i in range(args.n):
        new_seed = random.randint(1, 2**63-1)
        sub = argparse.Namespace(prompt=r['prompt'], negative=r['negative'], preset=r['preset'],
                                 seed=new_seed, name=f"vary_{r['id']}_{i+1}", parent=r['id'], timeout=args.timeout)
        print(f"\n{P}variation {i+1}/{args.n}{R}  (parent #{r['id']})")
        cmd_render(sub)

def cmd_rate(args):
    with db() as conn:
        if not conn.execute("SELECT 1 FROM renders WHERE id=?", (args.id,)).fetchone():
            print(f"{X}not found:{R} {args.id}"); sys.exit(1)
        conn.execute("UPDATE renders SET rating=?, notes=COALESCE(?, notes) WHERE id=?", (args.rating, args.note, args.id))
    print(f"{G}rated #{args.id}: {'★'*args.rating + '·'*(5-args.rating)}{R}")

def cmd_models(args):
    root = Path.home() / "ComfyUI/models"
    if not root.exists():
        print(f"{Y}no ~/ComfyUI/models dir — run: anime install wanmodels{R}"); return
    sections = [
        ("diffusion_models", "wan*.safetensors"),
        ("text_encoders",    "*umt5*.safetensors"),
        ("vae",              "*wan*.safetensors"),
        ("loras",            "wan*.safetensors"),
    ]
    print(f"\n{B}Wan models in {root}{R}")
    any_found = False
    for sub, pattern in sections:
        d = root / sub
        files = sorted(d.glob(pattern)) if d.exists() else []
        if not files:
            print(f"  {D}{sub}/{R}  {Y}(none){R}")
            continue
        any_found = True
        print(f"  {B}{sub}/{R}")
        for f in files:
            sz = f.stat().st_size / (1024**3)
            print(f"    {f.name:<60}  {sz:>5.1f}GB")
    if not any_found:
        print(f"  {Y}no Wan models found — run: anime install wanmodels{R}")
    print()

def cmd_presets(args):
    print(f"\n{B}available presets:{R}")
    for k, v in PRESETS.items():
        marker = G+"●"+R if k==DEFAULT_PRESET else " "
        print(f"  {marker} {B}{k:<22}{R} {D}{v['description']}{R}")
    print()

def cmd_stats(args):
    with db() as conn:
        n   = conn.execute("SELECT COUNT(*) FROM renders").fetchone()[0]
        nd  = conn.execute("SELECT COUNT(*) FROM renders WHERE status='done'").fetchone()[0]
        nf  = conn.execute("SELECT COUNT(*) FROM renders WHERE status='failed'").fetchone()[0]
        tt  = conn.execute("SELECT SUM(render_seconds) FROM renders WHERE status='done'").fetchone()[0] or 0
        ts  = conn.execute("SELECT SUM(file_size) FROM renders WHERE status='done'").fetchone()[0] or 0
        rated = conn.execute("SELECT AVG(rating) FROM renders WHERE rating IS NOT NULL").fetchone()[0]
        top = conn.execute("SELECT id, substr(prompt,1,60) as p, rating FROM renders WHERE rating>=4 ORDER BY rating DESC, id DESC LIMIT 5").fetchall()
    print(f"\n{B}{P}wan-pipeline stats{R}  ({DB_PATH})")
    print(f"  total renders:    {n}  ({G}{nd} done{R}, {X}{nf} failed{R})")
    print(f"  GPU time used:    {tt/60:.1f} min   ({tt/3600:.2f} GPU-hours)")
    print(f"  disk used:        {ts/1024/1024/1024:.1f} GB")
    if rated: print(f"  avg rating:       {'★'*int(rated+0.5)}{D} ({rated:.2f}){R}")
    if top:
        print(f"  {B}top-rated:{R}")
        for r in top:
            print(f"    #{r['id']}  {'★'*r['rating']}  {r['p']}")
    print()

def cmd_tui(args):
    """Tiny inline TUI — no Bubble Tea, just curses."""
    import curses
    def draw(stdscr):
        curses.curs_set(0)
        stdscr.clear()
        stdscr.addstr(0, 2, "wan-pipeline TUI", curses.A_BOLD)
        stdscr.addstr(2, 2, "(r) render  (h) history  (s) stats  (p) presets  (q) quit", curses.A_DIM)
        stdscr.addstr(4, 2, "Run `wan render \"prompt\"` for now — full TUI coming.")
        stdscr.addstr(6, 2, "Press any key.")
        stdscr.getch()
    curses.wrapper(draw)
    print(f"{D}(rich TUI is the next iteration — current commands cover all features){R}")

# ──────────────────────────────────────────────────────────────────
# CLI parser
# ──────────────────────────────────────────────────────────────────
def main():
    ap = argparse.ArgumentParser(description="wan-pipeline — stateful Wan 2.2 render")
    sp = ap.add_subparsers(dest="cmd", required=True)

    pr = sp.add_parser("render"); pr.add_argument("prompt")
    pr.add_argument("--preset", default=DEFAULT_PRESET)
    pr.add_argument("--negative", default=None)
    pr.add_argument("--explicit", action="store_true",
                    help="Drop NSFW gating from the negative prompt (allow explicit content)")
    pr.add_argument("--seed", type=int, default=None)
    pr.add_argument("--name", default=None)
    pr.add_argument("--parent", type=int, default=None)
    pr.add_argument("--timeout", type=int, default=1800)
    pr.set_defaults(fn=cmd_render)

    ph = sp.add_parser("history"); ph.add_argument("-n", type=int, default=20); ph.add_argument("--json", action="store_true"); ph.set_defaults(fn=cmd_history)
    pw = sp.add_parser("show");    pw.add_argument("id", type=int);             pw.set_defaults(fn=cmd_show)
    ps = sp.add_parser("resume");  ps.add_argument("id", type=int); ps.add_argument("--timeout", type=int, default=1800); ps.set_defaults(fn=cmd_resume)
    pv = sp.add_parser("vary");    pv.add_argument("id", type=int); pv.add_argument("-n", type=int, default=4); pv.add_argument("--timeout", type=int, default=1800); pv.set_defaults(fn=cmd_vary)
    pt = sp.add_parser("rate");    pt.add_argument("id", type=int); pt.add_argument("rating", type=int, choices=[1,2,3,4,5]); pt.add_argument("--note", default=None); pt.set_defaults(fn=cmd_rate)
    sp.add_parser("models")  .set_defaults(fn=cmd_models)
    sp.add_parser("presets") .set_defaults(fn=cmd_presets)
    sp.add_parser("stats")   .set_defaults(fn=cmd_stats)
    sp.add_parser("tui")     .set_defaults(fn=cmd_tui)

    args = ap.parse_args()
    args.fn(args)

if __name__ == "__main__":
    main()
