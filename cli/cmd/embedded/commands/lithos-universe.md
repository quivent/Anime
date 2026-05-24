# /lithos-universe — Lithos Integration Lens for Universe

Activate the Lithos integration perspective. Every rendering decision gets evaluated through the Lithos lens.

---

## FAST LOAD

Read these memory files in order:
1. `~/.claude/projects/-Users-joshkornreich-universe/memory/project_lithos_integration_map.md` — current bridge, gaps, directive
2. `~/.claude/projects/-Users-joshkornreich-universe/memory/feedback_lithos_lens.md` — the five rules
3. `~/.claude/projects/-Users-joshkornreich-universe/memory/reference_lithos_paint_api.md` — emitter API

Check live coverage: open `/lithos` in browser or run:
```bash
cd ~/lithos/packages/paint && node emit-scene.mjs --list
```

---

## COVERAGE STATUS

Check what's active:
```bash
# Which scenes have emitters?
node -e "import('~/lithos/packages/paint/lithos-emit.mjs').then(m => {
  console.log('Scenes:', Object.keys(m.SCENES));
  for (const [k,v] of Object.entries(m.SCENES)) console.log(k + ':', v.length, 'objects');
})"

# Which loaders have Lithos routing?
grep -rl 'isLithosMode' ~/universe/homes/*/
```

---

## THE FIVE RULES (always active)

1. **sceneSDF is sacred** — every SDF scene defines `float sceneSDF(vec3 p)`. Never rename it.
2. **Use 0.879** — Lipschitz-derived march step. Never heuristic values.
3. **Make objects bakeable** — positions as `{type, pos, params, boundR}` for the emitter.
4. **Extend the paint package** — all Lithos code in `~/lithos/packages/paint/`. Never fork.
5. **Keep SDFs self-contained** — pure functions of `(vec3 p, ...)`, no global state.

---

## ACTIONS

### Adding Lithos mode to a new home

1. Define scene array in `~/lithos/packages/paint/lithos-emit.mjs`:
   ```javascript
   export const CAPRICORN_SCENE = [
     { type: 'peak', pos: [0, 0, -5], params: [height], boundR: 8.0 },
     ...
   ];
   ```
   Add to `SCENES` registry.

2. Add emitter functions for new types:
   ```javascript
   function emitPeakSDF(varName, obj) { ... }
   ```
   Register in `EMITTERS` table.

3. Add terrain function if different from Virgo:
   ```javascript
   export function jsCapricornTerrainH(x, z) { ... }
   ```

4. Create shader template: `~/lithos/packages/paint/capricorn_lithos_shader.glsl`
   - Copy structure from `virgo_lithos_shader.glsl` or `taurus_lithos_shader.glsl`
   - Must have `/*LITHOS_SCENE_SDF*/` and `/*LITHOS_MARCH*/` placeholders
   - Extract GLSL from the home's `*_sdf.mjs` file

5. Create client installer: `~/lithos/packages/paint/capricorn_lithos_baked.mjs`
   - Copy from `taurus_lithos_baked.mjs`, change camera defaults and scene name

6. Update loader: `homes/capricorn/capricorn_loader.mjs`
   - Add `isLithosMode()` branch loading from paint package

7. Update serve.mjs import if new scene/terrain exported

8. Test:
   ```bash
   node ~/lithos/packages/paint/emit-scene.mjs capricorn --stats
   curl 'http://localhost:7180/lithos/emit?scene=capricorn' | node -e "process.stdin.on('data',d=>{const j=JSON.parse(d);console.log('OK:',j.visible,'objects,',j.fragmentShader.length,'chars')})"
   ```

### Evaluating any code change

Before committing, ask:
- Does this `sceneSDF` stay a pure function of `vec3 p`?
- Can the emitter bake the positions I'm adding?
- Am I using 0.879 step factor?
- Is the new SDF function portable to `.ls` notation?
- Did I add the object type to the EMITTERS registry?

### Debugging shader emission

```bash
# Emit GLSL to inspect
node ~/lithos/packages/paint/emit-scene.mjs taurus > /tmp/taurus.glsl

# Check frustum culling
node ~/lithos/packages/paint/emit-scene.mjs taurus 50 5 0 0 0 --stats
# (camera far away, facing +Z — should cull most objects)

# Verify server endpoint
curl -s 'http://localhost:7180/lithos/emit?scene=taurus' | python3 -m json.tool | head -5
```

---

## KEY FILES

| File | Role |
|------|------|
| `~/lithos/packages/paint/lithos-emit.mjs` | Emitter + scene registry + EMITTERS table |
| `~/lithos/packages/paint/emit-scene.mjs` | CLI emit tool |
| `~/lithos/packages/paint/*_lithos_shader.glsl` | Shader templates per scene |
| `~/lithos/packages/paint/*_lithos_baked.mjs` | Client installers per scene |
| `universe/serve.mjs` | Server routes for /lithos/emit, /lithos/heightmap |
| `universe/homes/params.mjs` | `isLithosMode()` detection |
| `universe/homes/*/loader.mjs` | Per-home Lithos routing |
| `universe/lithos.html` | Live coverage dashboard |

---

## ACTIVATION

When `/lithos-universe` is invoked:

1. Load the three memory files listed above.
2. Run coverage check: which scenes exist in SCENES registry, which loaders have active routing.
3. Check `/lithos` dashboard status (live stats from emitter).
4. Announce state and identify the single highest-value next integration.
5. Apply the Lithos lens to the current task.

Format:
```
LITHOS LENS active.

Scenes: virgo (15 obj), taurus (29 obj) — [N] more planned
Loaders: 3/12 active, 9/12 prepped
Coverage: [dashboard URL]
Next: [single most valuable integration action]
Lens applied to current task.
```
