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
            f"ComfyUI not reachable at {COMFY_API} ({e.reason}).\n"
            f"  Start it:     anime comfyui start\n"
            f"  Check status: curl {COMFY_API}/system_stats\n"
            f"  Custom URL:   COMFY_API=http://host:port anime wan render ..."
        ) from None
    if r.get("node_errors"):
        first_key, first_val = next(iter(r["node_errors"].items()))
        # Try to extract the actual error string from ComfyUI's nested format
        if isinstance(first_val, dict):
            errs = first_val.get("errors", [])
            msg = errs[0].get("message", str(first_val)) if errs else str(first_val)
        else:
            msg = str(first_val)
        raise RuntimeError(
            f"ComfyUI rejected the workflow (node: {first_key}).\n"
            f"  Error:  {msg}\n"
            f"  This usually means a model file is missing or misnamed.\n"
            f"  Check:  anime wan models\n"
            f"  Fix:    anime install wanmodels"
        )
    return r["prompt_id"]

_SPINNER = ["⠋","⠙","⠹","⠸","⠼","⠴","⠦","⠧","⠇","⠏"]

def wait_for(prompt_id: str, timeout: int = 1800) -> dict:
    start = time.time()
    tick = 0
    # Track progress by polling the ComfyUI queue for position/running state
    while True:
        elapsed = time.time() - start
        try:
            h = json.load(urllib.request.urlopen(f"{COMFY_API}/history/{prompt_id}", timeout=10))
        except urllib.error.URLError as e:
            raise RuntimeError(
                f"ComfyUI unreachable at {COMFY_API} ({e.reason}).\n"
                f"  Start it:  anime comfyui start\n"
                f"  Check it:  curl {COMFY_API}/system_stats"
            ) from None
        if prompt_id in h:
            entry = h[prompt_id]
            status = (entry.get("status") or {}).get("status_str", "")
            if status == "error":
                msgs = (entry.get("status") or {}).get("messages", [])
                detail = next((m[1] for m in msgs if m and m[0] == "execution_error"), msgs)
                raise RuntimeError(f"render failed in ComfyUI: {detail}")
            # Clear the spinner line
            print(f"\r{' '*60}\r", end="", flush=True)
            return entry

        # In-place spinner with elapsed time
        spin = _SPINNER[tick % len(_SPINNER)]
        mins, secs = divmod(int(elapsed), 60)
        if mins > 0:
            elapsed_str = f"{mins}m {secs:02d}s"
        else:
            elapsed_str = f"{secs}s"
        print(f"\r  {P}{spin}{R} Rendering... {D}(elapsed: {elapsed_str}){R}  ", end="", flush=True)
        tick += 1

        if elapsed > timeout:
            print()
            raise TimeoutError(
                f"Render exceeded {timeout}s timeout.\n"
                f"  The render may still be running in ComfyUI.\n"
                f"  Check:  {COMFY_API}/queue\n"
                f"  Increase timeout:  --timeout {timeout*2}"
            )
        time.sleep(2)

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
def _estimate_render_time(preset_name: str) -> str:
    """Return a human-friendly time estimate for a preset."""
    p = PRESETS[preset_name]
    steps = p.get("steps", 20)
    res = p.get("width", 832) * p.get("height", 480)
    # Rough heuristic: maxq (1280x720, 50 steps) ~ 5min; fast (832x480, 8 steps) ~ 30s
    if steps >= 40 and res >= 900000:
        return "~5 minutes on H100"
    elif steps <= 10:
        return "~30 seconds on H100"
    else:
        return "~1-2 minutes on H100"


def cmd_render(args):
    preset = args.preset
    if preset not in PRESETS:
        print(f"\n  {X}Unknown preset:{R} {preset}")
        print(f"  {D}Available presets:{R}")
        for k in PRESETS:
            marker = f"{G}*{R}" if k == DEFAULT_PRESET else " "
            print(f"    {marker} {k}")
        print(f"\n  {D}Use: anime wan render \"prompt\" --preset <name>{R}\n")
        sys.exit(1)
    seed = args.seed if args.seed is not None else random.randint(1, 2**63-1)
    name = args.name or f"render_{int(time.time())}"
    # --negative wins; otherwise pick SFW or explicit baseline.
    if args.negative is not None:
        negative = args.negative
    elif getattr(args, "explicit", False):
        negative = DEFAULT_NEGATIVE_EXPLICIT
    else:
        negative = DEFAULT_NEGATIVE_SFW

    # Confirmation for expensive presets (maxq = 50 steps at full resolution)
    p_info = PRESETS[preset]
    is_expensive = p_info.get("steps", 0) >= 40
    if is_expensive and not getattr(args, "yes", False) and sys.stdin.isatty():
        est = _estimate_render_time(preset)
        dims = f"{p_info['width']}x{p_info['height']}, {p_info['steps']} steps"
        print(f"\n  {Y}This is a high-quality render ({dims}).{R}")
        print(f"  {D}Estimated time: {est}{R}")
        try:
            resp = input(f"  Continue? [Y/n] ").strip().lower()
        except (EOFError, KeyboardInterrupt):
            print(f"\n  {D}Cancelled.{R}\n")
            sys.exit(0)
        if resp and resp not in ("y", "yes", ""):
            print(f"  {D}Cancelled.{R}\n")
            sys.exit(0)

    graph = build_workflow(preset, args.prompt, negative, seed, name)

    with db() as conn:
        cur = conn.execute("""
            INSERT INTO renders(created_at, name, prompt, negative, preset, params_json, seed, parent_id, status)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'pending')
        """, (time.strftime("%Y-%m-%d %H:%M:%S"), name, args.prompt, negative, preset,
              json.dumps(PRESETS[preset]), seed, args.parent))
        rid = cur.lastrowid

    p = PRESETS[preset]
    dims = f"{p['width']}x{p['height']}"
    est = _estimate_render_time(preset)
    print(f"\n{B}{P}╭─ render #{rid}{R}")
    print(f"{P}│{R}  {B}preset{R}   {preset} {D}({dims}, {p['steps']} steps){R}")
    print(f"{P}│{R}  {B}seed{R}     {seed}")
    print(f"{P}│{R}  {B}prompt{R}   {args.prompt[:80]}{'...' if len(args.prompt)>80 else ''}")
    print(f"{P}│{R}  {B}est{R}      {est}")
    print(f"{P}╰─{R}\n")

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
        result = wait_for(pid, timeout=args.timeout)
        elapsed = time.time() - t0
        outs = extract_outputs(result)
        if not outs:
            raise RuntimeError(
                "Render completed but produced no output files.\n"
                "  This can happen if the SaveVideo node is misconfigured.\n"
                "  Check ComfyUI logs for details."
            )
        primary = outs[0]
        size = Path(primary["local"]).stat().st_size if Path(primary["local"]).exists() else 0
        with db() as conn:
            conn.execute("""
                UPDATE renders SET status='done', output_path=?, output_url=?, file_size=?, render_seconds=? WHERE id=?
            """, (primary["local"], primary["url"], size, elapsed, rid))

        # Success card
        secs = int(elapsed)
        if secs >= 60:
            dur = f"{secs//60}m {secs%60}s"
        else:
            dur = f"{secs}s"
        sz_mb = size / 1024 / 1024
        print(f"\n  {G}Done{R}  {dur}  {D}·{R}  {sz_mb:.1f} MB")
        print(f"  {B}url{R}    {primary['url']}")
        print(f"  {B}local{R}  {primary['local']}")
        print(f"  {B}id{R}     {rid}")
        print(f"\n  {D}Next:  anime wan show {rid}    anime wan vary {rid}    anime wan rate {rid} 5{R}\n")
    except Exception as e:
        with db() as conn:
            conn.execute("UPDATE renders SET status='failed', notes=? WHERE id=?", (str(e), rid))
        print(f"\n  {X}Failed:{R} {e}\n")
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
        print(f"\n  {D}No renders yet.{R}")
        print(f"  {D}Get started:{R}  anime wan render \"a dragon breathing fire\"\n")
        return

    # Clean, aligned table
    print()
    print(f"  {C}{B}{'ID':>4}  {'Date':<12}  {'Status':<9}  {'Time':>5}  {'Size':>6}  {'Rating':<5}  Prompt{R}")
    print(f"  {D}{'─'*4}  {'─'*12}  {'─'*9}  {'─'*5}  {'─'*6}  {'─'*5}  {'─'*30}{R}")
    for r in rows:
        st = r["status"] or "?"
        st_c = G if st == "done" else (Y if st == "pending" else X)
        rating_val = r["rating"]
        if rating_val:
            rating = f"{Y}{'★'*rating_val}{D}{'·'*(5-rating_val)}{R}"
        else:
            rating = f"{D}·····{R}"
        sz = f"{(r['file_size'] or 0)/1024/1024:.1f}M" if r["file_size"] else f"{D}  —{R} "
        if r["render_seconds"]:
            secs = int(r["render_seconds"])
            t = f"{secs//60}m{secs%60:02d}" if secs >= 60 else f"{secs}s"
        else:
            t = f"{D} —{R} "
        prompt_short = (r["prompt"] or "")[:40]
        date_short = (r["created_at"] or "")[:10]  # just YYYY-MM-DD
        print(f"  {B}{r['id']:>4}{R}  {date_short:<12}  {st_c}{st:<9}{R}  {t:>5}  {sz:>6}  {rating}  {prompt_short}")
    total = len(rows)
    print(f"\n  {D}Showing {total} render{'s' if total != 1 else ''}.{R}", end="")
    if total == args.n:
        print(f"  {D}More: anime wan history -n {args.n * 2}{R}")
    else:
        print()
    print()

def cmd_show(args):
    with db() as conn:
        r = conn.execute("SELECT * FROM renders WHERE id=?", (args.id,)).fetchone()
    if not r:
        print(f"\n  {X}Render #{args.id} not found.{R}")
        print(f"  {D}List renders: anime wan history{R}\n")
        sys.exit(1)

    st = r['status'] or "?"
    st_c = G if st == "done" else (Y if st == "pending" else X)
    p = json.loads(r['params_json'])

    # Card layout with box-drawing characters
    print()
    print(f"  {P}╭─────────────────────────────────────────────────────────────╮{R}")
    print(f"  {P}│{R}  {B}Render #{r['id']}{R}   {st_c}{st}{R}   {D}{r['created_at']}{R}")
    print(f"  {P}├─────────────────────────────────────────────────────────────┤{R}")
    print(f"  {P}│{R}")
    print(f"  {P}│{R}  {C}prompt{R}     {r['prompt']}")
    if r['negative']:
        neg_short = r['negative'][:80] + ('...' if len(r['negative']) > 80 else '')
        print(f"  {P}│{R}  {D}negative{R}   {D}{neg_short}{R}")
    print(f"  {P}│{R}")
    print(f"  {P}│{R}  {C}preset{R}     {r['preset']}")
    print(f"  {P}│{R}  {C}seed{R}       {r['seed']}")
    print(f"  {P}│{R}  {C}params{R}     {p['width']}x{p['height']}  {D}·{R}  {p['length']}f @ {p['fps']}fps  {D}·{R}  {p['steps']} steps  {D}·{R}  cfg {p['cfg']}")
    if r['parent_id']:
        print(f"  {P}│{R}  {C}parent{R}     #{r['parent_id']}")
    if r['render_seconds']:
        secs = int(r['render_seconds'])
        if secs >= 60:
            dur = f"{secs//60}m {secs%60}s"
        else:
            dur = f"{secs}s"
        print(f"  {P}│{R}  {C}duration{R}   {dur}")
    if r['file_size']:
        print(f"  {P}│{R}  {C}size{R}       {r['file_size']/1024/1024:.1f} MB")
    if r['rating']:
        stars = f"{Y}{'★'*r['rating']}{D}{'·'*(5-r['rating'])}{R}"
        print(f"  {P}│{R}  {C}rating{R}     {stars}")
    print(f"  {P}│{R}")
    if r['output_url']:
        print(f"  {P}│{R}  {G}url{R}        {r['output_url']}")
    if r['output_path']:
        print(f"  {P}│{R}  {D}local{R}      {r['output_path']}")
    if r['notes']:
        print(f"  {P}│{R}  {D}notes{R}      {r['notes']}")
    print(f"  {P}│{R}")
    print(f"  {P}╰─────────────────────────────────────────────────────────────╯{R}")

    # Contextual next actions
    if st == "done":
        print(f"  {D}Next:  anime wan vary {r['id']}    anime wan rate {r['id']} 5{R}")
    elif st == "failed":
        print(f"  {D}Next:  anime wan resume {r['id']}  (retry with same seed){R}")
    print()

def cmd_resume(args):
    """Re-render with the same seed (deterministic reproduce)."""
    with db() as conn:
        r = conn.execute("SELECT * FROM renders WHERE id=?", (args.id,)).fetchone()
    if not r:
        print(f"\n  {X}Render #{args.id} not found.{R}")
        print(f"  {D}List renders: anime wan history{R}\n")
        sys.exit(1)
    sub = argparse.Namespace(prompt=r['prompt'], negative=r['negative'], preset=r['preset'],
                             seed=r['seed'], name=f"resume_{r['id']}", parent=r['id'],
                             timeout=args.timeout, yes=True)
    print(f"\n  {C}Resuming render #{r['id']}{R} {D}(same seed: {r['seed']}){R}")
    cmd_render(sub)

def cmd_vary(args):
    """Same prompt + preset, fresh seeds."""
    with db() as conn:
        r = conn.execute("SELECT * FROM renders WHERE id=?", (args.id,)).fetchone()
    if not r:
        print(f"\n  {X}Render #{args.id} not found.{R}")
        print(f"  {D}List renders: anime wan history{R}\n")
        sys.exit(1)
    est = _estimate_render_time(r['preset']) if r['preset'] in PRESETS else "unknown"
    print(f"\n  {P}Generating {args.n} variation{'s' if args.n != 1 else ''} of render #{r['id']}{R}")
    print(f"  {D}prompt: {r['prompt'][:60]}{'...' if len(r['prompt'])>60 else ''}{R}")
    print(f"  {D}est per render: {est}{R}")
    for i in range(args.n):
        new_seed = random.randint(1, 2**63-1)
        sub = argparse.Namespace(prompt=r['prompt'], negative=r['negative'], preset=r['preset'],
                                 seed=new_seed, name=f"vary_{r['id']}_{i+1}", parent=r['id'],
                                 timeout=args.timeout, yes=True)
        print(f"\n  {P}[{i+1}/{args.n}]{R} variation  {D}(seed: {new_seed}){R}")
        cmd_render(sub)

def cmd_rate(args):
    with db() as conn:
        row = conn.execute("SELECT id, prompt FROM renders WHERE id=?", (args.id,)).fetchone()
        if not row:
            print(f"\n  {X}Render #{args.id} not found.{R}")
            print(f"  {D}List renders: anime wan history{R}\n")
            sys.exit(1)
        conn.execute("UPDATE renders SET rating=?, notes=COALESCE(?, notes) WHERE id=?", (args.rating, args.note, args.id))
    stars = f"{Y}{'★'*args.rating}{D}{'·'*(5-args.rating)}{R}"
    prompt_short = (row['prompt'] or "")[:50]
    print(f"\n  {G}Rated #{args.id}{R}  {stars}")
    print(f"  {D}{prompt_short}{R}\n")

def cmd_models(args):
    root = Path.home() / "ComfyUI/models"
    if not root.exists():
        print(f"\n  {Y}ComfyUI models directory not found.{R}")
        print(f"  {D}Expected: ~/ComfyUI/models{R}")
        print(f"  {D}Install:  anime install wan{R}\n")
        return
    sections = [
        ("diffusion_models", "wan*.safetensors", "Diffusion models (the brains)"),
        ("text_encoders",    "*umt5*.safetensors", "Text encoders (prompt understanding)"),
        ("vae",              "*wan*.safetensors", "VAE (latent decoder)"),
        ("loras",            "wan*.safetensors", "LoRA adapters (speed/style)"),
    ]
    print(f"\n  {B}Wan models{R}  {D}{root}{R}\n")
    any_found = False
    total_gb = 0.0
    for sub, pattern, label in sections:
        d = root / sub
        files = sorted(d.glob(pattern)) if d.exists() else []
        if not files:
            print(f"  {D}  {label:<40}  (none){R}")
            continue
        any_found = True
        print(f"  {C}{label}{R}")
        for f in files:
            sz = f.stat().st_size / (1024**3)
            total_gb += sz
            print(f"    {f.name:<55}  {G}{sz:>5.1f} GB{R}")
        print()
    if not any_found:
        print(f"  {Y}No Wan models found.{R}")
        print(f"  {D}Install: anime install wanmodels{R}")
    else:
        print(f"  {D}Total: {total_gb:.1f} GB{R}")
    print()

def cmd_presets(args):
    print(f"\n  {B}Render presets{R}\n")
    for k, v in PRESETS.items():
        if k == DEFAULT_PRESET:
            marker = f"{G}*{R}"
            name_style = f"{B}{G}{k}{R}"
        else:
            marker = " "
            name_style = f"{B}{k}{R}"
        dims = f"{v['width']}x{v['height']}"
        est = _estimate_render_time(k)
        print(f"  {marker} {name_style}")
        print(f"      {D}{v['description']}{R}")
        print(f"      {D}{dims}  ·  {v['steps']} steps  ·  {est}{R}")
        print()
    print(f"  {D}{G}*{R} {D}= default.  Override: anime wan render \"prompt\" --preset <name>{R}\n")

def cmd_stats(args):
    with db() as conn:
        n   = conn.execute("SELECT COUNT(*) FROM renders").fetchone()[0]
        nd  = conn.execute("SELECT COUNT(*) FROM renders WHERE status='done'").fetchone()[0]
        nf  = conn.execute("SELECT COUNT(*) FROM renders WHERE status='failed'").fetchone()[0]
        np_ = conn.execute("SELECT COUNT(*) FROM renders WHERE status='pending'").fetchone()[0]
        nc  = conn.execute("SELECT COUNT(*) FROM renders WHERE status='cancelled'").fetchone()[0]
        tt  = conn.execute("SELECT SUM(render_seconds) FROM renders WHERE status='done'").fetchone()[0] or 0
        ts  = conn.execute("SELECT SUM(file_size) FROM renders WHERE status='done'").fetchone()[0] or 0
        avg_t = conn.execute("SELECT AVG(render_seconds) FROM renders WHERE status='done'").fetchone()[0] or 0
        rated = conn.execute("SELECT AVG(rating) FROM renders WHERE rating IS NOT NULL").fetchone()[0]
        nr  = conn.execute("SELECT COUNT(*) FROM renders WHERE rating IS NOT NULL").fetchone()[0]
        top = conn.execute("SELECT id, substr(prompt,1,50) as p, rating FROM renders WHERE rating>=4 ORDER BY rating DESC, id DESC LIMIT 5").fetchall()
        recent = conn.execute("SELECT id, substr(prompt,1,50) as p, status, render_seconds FROM renders ORDER BY id DESC LIMIT 3").fetchall()

    if n == 0:
        print(f"\n  {D}No renders yet.{R}")
        print(f"  {D}Get started:{R}  anime wan render \"a dragon breathing fire\"\n")
        return

    # Dashboard layout
    print()
    print(f"  {P}╭─────────────────────────────────────────────────────────╮{R}")
    print(f"  {P}│{R}  {B}wan-pipeline dashboard{R}                                  {P}│{R}")
    print(f"  {P}├─────────────────────────────────────────────────────────┤{R}")
    print(f"  {P}│{R}                                                         {P}│{R}")
    # Renders
    success_rate = f"{nd/n*100:.0f}%" if n > 0 else "—"
    print(f"  {P}│{R}  {C}Renders{R}         {B}{n}{R} total                               {P}│{R}")
    print(f"  {P}│{R}                  {G}{nd} done{R}  {X}{nf} failed{R}  {Y}{np_} pending{R}  {D}{nc} cancelled{R}")
    print(f"  {P}│{R}                  {D}success rate: {success_rate}{R}")
    print(f"  {P}│{R}                                                         {P}│{R}")
    # Time
    gpu_hrs = tt / 3600
    avg_secs = int(avg_t)
    avg_str = f"{avg_secs//60}m {avg_secs%60}s" if avg_secs >= 60 else f"{avg_secs}s"
    print(f"  {P}│{R}  {C}GPU time{R}        {B}{gpu_hrs:.1f}{R} hours ({tt/60:.0f} min total)        {P}│{R}")
    print(f"  {P}│{R}                  {D}avg per render: {avg_str}{R}")
    print(f"  {P}│{R}                                                         {P}│{R}")
    # Storage
    gb = ts / 1024 / 1024 / 1024
    print(f"  {P}│{R}  {C}Storage{R}         {B}{gb:.1f}{R} GB                                {P}│{R}")
    print(f"  {P}│{R}                                                         {P}│{R}")
    # Rating
    if rated:
        stars = f"{Y}{'★'*int(rated+0.5)}{D}{'·'*(5-int(rated+0.5))}{R}"
        print(f"  {P}│{R}  {C}Rating{R}          {stars} {D}({rated:.1f} avg, {nr} rated){R}")
    else:
        print(f"  {P}│{R}  {C}Rating{R}          {D}no ratings yet — try: anime wan rate <id> 5{R}")
    print(f"  {P}│{R}                                                         {P}│{R}")
    print(f"  {P}╰─────────────────────────────────────────────────────────╯{R}")

    if top:
        print(f"\n  {B}Top rated{R}")
        for r in top:
            print(f"    {Y}{'★'*r['rating']}{R}  #{r['id']:<4}  {r['p']}")

    if recent:
        print(f"\n  {B}Recent{R}")
        for r in recent:
            st = r["status"] or "?"
            st_c = G if st == "done" else (Y if st == "pending" else X)
            t_str = ""
            if r["render_seconds"]:
                s = int(r["render_seconds"])
                t_str = f"  {D}{s//60}m{s%60:02d}s{R}" if s >= 60 else f"  {D}{s}s{R}"
            print(f"    #{r['id']:<4}  {st_c}{st:<9}{R}{t_str}  {r['p']}")
    print(f"\n  {D}db: {DB_PATH}{R}\n")

def cmd_tui(args):
    """Tiny inline TUI — no Bubble Tea, just curses."""
    import curses
    def draw(stdscr):
        curses.curs_set(0)
        stdscr.clear()
        stdscr.addstr(0, 2, "wan-pipeline TUI", curses.A_BOLD)
        stdscr.addstr(2, 2, "(r) render  (h) history  (s) stats  (p) presets  (q) quit", curses.A_DIM)
        stdscr.addstr(4, 2, "Run `anime wan render \"prompt\"` from the CLI for now.")
        stdscr.addstr(5, 2, "Or: `anime wan studio` for the full web UI.")
        stdscr.addstr(7, 2, "Press any key to exit.")
        stdscr.getch()
    curses.wrapper(draw)
    print(f"\n  {D}The native TUI uses the Go Bubble Tea interface.{R}")
    print(f"  {D}Run: anime wan tui  (from the compiled CLI){R}\n")

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
    pr.add_argument("--yes", "-y", action="store_true",
                    help="Skip confirmation prompts (for scripting)")
    pr.set_defaults(fn=cmd_render)

    ph = sp.add_parser("history"); ph.add_argument("-n", type=int, default=10); ph.add_argument("--json", action="store_true"); ph.set_defaults(fn=cmd_history)
    pw = sp.add_parser("show");    pw.add_argument("id", type=int);             pw.set_defaults(fn=cmd_show)
    ps = sp.add_parser("resume");  ps.add_argument("id", type=int); ps.add_argument("--timeout", type=int, default=1800); ps.set_defaults(fn=cmd_resume)
    pv = sp.add_parser("vary");    pv.add_argument("id", type=int); pv.add_argument("-n", type=int, default=3); pv.add_argument("--timeout", type=int, default=1800); pv.set_defaults(fn=cmd_vary)
    pt = sp.add_parser("rate");    pt.add_argument("id", type=int); pt.add_argument("rating", type=int, choices=[1,2,3,4,5]); pt.add_argument("--note", default=None); pt.set_defaults(fn=cmd_rate)
    sp.add_parser("models")  .set_defaults(fn=cmd_models)
    sp.add_parser("presets") .set_defaults(fn=cmd_presets)
    sp.add_parser("stats")   .set_defaults(fn=cmd_stats)
    sp.add_parser("tui")     .set_defaults(fn=cmd_tui)

    args = ap.parse_args()
    args.fn(args)

if __name__ == "__main__":
    main()
