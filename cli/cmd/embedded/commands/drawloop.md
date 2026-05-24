# /drawloop — WebGL SDF Visual Iteration Protocol

Rapidly iterate on GLSL shaders using headless Metal-backed WebGL screenshots + source comparison.

---

## SCREENSHOT MECHANISM

**Critical:** Headless Chrome must use Metal/ANGLE for WebGL to work. `--disable-gpu` kills WebGL.

```js
const browser = await puppeteer.launch({
  headless: 'new',
  args: [
    '--no-sandbox',
    '--use-angle=metal',          // Metal backend for macOS ARM64
    '--ignore-gpu-blocklist',
    '--enable-gpu',
  ]
});
const page = await browser.newPage();
await page.setViewport({ width: 1280, height: 720 });
await page.setCacheEnabled(false);          // bypass ES module cache
await page.goto(url, { waitUntil: 'networkidle0', timeout: 20000 });
await new Promise(r => setTimeout(r, 4000)); // wait for first render
await page.screenshot({ path: outPath });
await browser.close();
```

**Verify WebGL is running:** If the screenshot looks like a flat CSS background color (not the 3D scene), WebGL failed. Watch for `THREE.WebGLRenderer: A WebGL context could not be created` in console.

---

## ITERATION LOOP

1. Take screenshot of SDF port (`/virgo/port`) and reference (`/virgo` or `/virgo/sdf`)
2. Read both screenshots visually
3. **Describe each scene in English**, ground-truth style:
   - Sky: zenith color, horizon color, stars, moon
   - Ground: color tone, brightness, texture
   - Primary objects: tree (scale, canopy quality, trunk), sunflowers (shape, emission), fireflies (count, visibility)
   - Overall atmosphere: local vs global lighting balance
4. Identify top 3 deltas
5. Map each delta to its GLSL code location
6. Edit → screenshot → repeat

---

## GLSL DOMAIN MAP (virgo_sdf_port.mjs)

| Visual element | Function | Key parameters |
|---|---|---|
| Sky color | `skyColor()` | `zenith`, `midNight`, `horizon`, `horGlow`, star density |
| Global lighting | main render loop | `key` (moonlight), `skyL` (hemisphere), `amb` (ambient), multipliers |
| Oak canopy fill | `isOak` branch | `fillPos`, `fillAtten`, `canopyWeight`, additive vec3 |
| Sunflower glow | `isFlower` branch | emission vec3 |
| Fireflies | firefly loop | `exp(-d2*d2*160.)`, `pulse`, count (14), radius (0.20) |
| Ground colors | `groundColor()` | `gD`, `gB`, `gL` grass vecs |
| Fog | fog block | density (`0.00030`), fogCol, max clamp |

---

## SOURCE COMPARISON PROTOCOL

When visuals don't match the polygon reference:

1. Read `virgo_sdf.mjs` (original night SDF) — same lighting model, calibrated values
2. Check `virgo_ground.mjs` for polygon surface colors (reference: `grassBase = vec3(0.29,0.35,0.20)`)
3. Run diff mentally:
   - Are surface colors in the same ballpark as polygon scene source?
   - Is ambient too high (washes out) or too low (too black)?
   - Is fog mixing in too much sky color (causes tint)?
4. Fix: surface colors first, then lighting ratios, then sky last

---

## COMMON FAILURE MODES

| Symptom | Cause | Fix |
|---|---|---|
| Scene is pure black CSS background | WebGL failed, wrong Chrome flags | Add `--use-angle=metal` |
| Ground is pink/teal | Fog mixing warm sky color | Set `fogCol = vec3(0.008,0.006,0.012)` (no sky mix) |
| Ground is too bright | Pre-lighting `col*=` tint in `groundColor()` or ambient too high | Remove tint, lower `amb` |
| Sky too vivid | `horGlow` too high after ACES lift | Reduce horGlow to < 0.05 |
| Trees invisible | Surface colors too dark for lighting ratio | Raise grass/leaf vecs OR raise key |
| Fireflies invisible | `d2 < 0.05` too tight, glow too dim | Use `d2 < 0.20`, `exp(-d2*d2*160.)*2.5` |
| Oak trunk glowing green | canopy fill light applied to full tree | Add `canopyWeight = smoothstep(3.5, 6.0, lp.y)` |

---

## CALIBRATION REFERENCE

The polygon `/virgo` scene is **local-light-dominant** — global ambient near-zero, everything lit by:
- Oak canopy internal fill (warm green additive)
- Sunflower face emission (orange)
- Fireflies (golden-yellow `vec3(0.95,0.85,0.20)`)
- Moonlight is very subtle

If the ground is visible when NOT near a light source, the ambient is too high.

The original night SDF (`virgo_sdf.mjs`) calibrated values:
```glsl
key  = vec3(0.62,0.68,0.80) * diff * shad * 0.32
skyL = vec3(0.08,0.10,0.22) * max(0., n.y) * 0.55
// no explicit ambient — col *= ao handles it
```

These values are verified working. Use them as anchors when recalibrating.
