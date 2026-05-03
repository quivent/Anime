# /pareto

Scope: $ARGUMENTS (no arg = whole project). Drive compiler warnings to zero via Pareto: 20% of sources cause 80% of noise.

## P0: Measure
Run compiler w/ warnings. Rust: `cargo check 2>&1`. TS: `npx tsc --noEmit`. Svelte: `npx svelte-check`. Adapt to stack.
Categorize each warning: type (unused import, dead code, deprecated, ambiguous re-export), source file, count per file.
Sort files by warning count desc = Pareto rank. Record baseline: total, top 5, categories.
Build+test before changes.

## P1: Elephants
Highest-concentration sources first. One rm can kill hundreds.
- 0 external callers → rm file
- `#[deprecated]`/legacy w/ replacement → verify replacement covers all, rm
- Callers but deprecated types unused → rm deprecated items only
- Glob re-export spreading warnings → explicit re-exports

Rules: grep before rm. Replacement exists → confirm coverage. Callers exist → inline/migrate first. One elephant at a time. Build+test between each.

## P2: Tail
After elephants, sweep scattered warnings. Batch by file.
- Unused import → rm `use` line
- Unused var → prefix `_` or rm if dead
- Ambiguous glob → explicit re-exports
- `#[allow(dead_code)]` truly dead → rm item. Serde field → leave
- Unnecessary `mut` → rm `mut`
- Deprecated call w/ replacement → migrate

For `#[allow(dead_code)]`: grep callers. 0 = rm. Serde = leave. Ambiguous re-exports: check conflicting names, export explicitly.
Build+test. Must pass.

## P3: Verify
Rerun compiler. Count remaining. Project warnings = must fix or justify. Dependency warnings = out of scope. If project warnings remain → back to P2.

## P4: Report
```
Baseline: [n] total, top=[file] ([n] warns [x]%)
Elephants: -[n] warns | File | Warns | Action | Lines |
Tail: -[n] warns | File | Type | Fix |
Final: [n] project / [m] dep, -[n]ln, build [pass/fail], tests [p]/[t]
```

## Vs /kernighan-refactor
Kernighan: reads code → dedup/modularize/compress → structural clarity.
Pareto: reads compiler → rank/elephants/tail → zero warnings.
Complement: Kernighan first (structure), then Pareto (noise).
