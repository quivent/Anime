# /kernighan-compress

Character-level compression. Not structural (`/kernighan-refactor`). Prose editing for code.

Scope: $ARGUMENTS (file, glob, directory; no arg = project source).

## Method

Read each file as if you didn't write it. Read it aloud — the telephone test. Everything that makes you stumble, fix. One file, one pass, one commit.

## Setup

1. Count chars in scope: `find <scope> -name '*.rs' -o -name '*.svelte' -o -name '*.ts' | xargs wc -c`. Record as `BEFORE`
2. Build + test baseline
3. Clean git state for rollback

## Per-file pass

Read top to bottom. Fix what you see. No categories — clarity is the only filter.

### Comments: if the code says it, the comment goes

| Kill | Rule |
|------|------|
| `/// Get foo` on `fn foo()` | Name says it → delete |
| `// ========` banners | Blank line or nothing |
| `// increment x` above `x += 1` | Echo → delete |
| `// Added in v2.3` | Git's job |
| `// let old = ...` | Git remembers |
| 15-line `//!` restating README | One line: `//! What it does.` |
| `/// The name` on `name: String` | Field says it → delete |

Keep: *why* comments. `// SAFETY:`. Real `// TODO:`. Algorithm explanations. `#[cfg]` context. `pub` items in library crates.

### Formatting: remove air

| Pattern | Action |
|---------|--------|
| Multi-line chain fitting 100 cols | One line |
| Variable used once, names nothing | Inline |
| 3+ blank lines | Collapse to 1 |
| Blank lines between struct fields | Remove |
| Trailing whitespace, extra EOF | Strip |

Never fight the formatter. If `rustfmt` expands it back, leave it.

### Names: scope determines length

| Rule | Example |
|------|---------|
| Short scope (<20 ln) → short name | `accumulated_text` → `text` |
| Long scope (>20 ln) → keep descriptive | leave it |
| Suffix echoes type → drop | `_string`, `_vec`, `_map` |
| Prefix echoes context → drop | `kv_cache_layers` in `process_kv_cache` → `layers` |
| `get_` returning a field → drop | `get_name()` → `name()` |
| Inconsistent convention → pick one | all `_act` or none |
| Method names: what, not how | `lookup_variance_multiplier` → `variance_multiplier_for` |

Don't rename: public API. Domain vocabulary. Cross-crate. Already clear (`i`, `n`, `e`).

### Idioms: say it the short way

| Verbose | Short |
|---------|-------|
| `if let Some(x) = y { x } else { d }` | `y.unwrap_or(d)` |
| `match x { true => a, false => b }` | `if x { a } else { b }` |
| Nested `.map().filter().map()` | Combine closures |
| Same `.map_err(format!(...))` >3 places | Extract helper |

Don't touch: `unsafe`. Performance paths. Correctness changes — that's refactoring.

### After each file

1. `wc -c <file>` — record AFTER for this file
2. `cargo check` — if it breaks, `git checkout -- <file>`. If it passes:

```
compress(<filename>): [before] → [after] chars (-[n]%)
```

3. Commit: `git commit -m "compress(<filename>): [before] → [after] (-N chars)"`

## Report

```
BEFORE: [n] chars across [n] files
AFTER:  [n] chars across [n] files
DELTA:  -[n] chars (-[x.x]%)

Per-file:
  [filename]: [before] → [after] (-[n] chars, -[x.x]%)
  [filename]: [before] → [after] (-[n] chars, -[x.x]%)
  ...
  [filename]: [before] → [after] (no change — reads clean)

Build: [pass/fail]  Tests: [p]/[t]
```

## One rule

Read it as if you didn't write it. If it's hard to read, rewrite it. Don't comment bad code — rewrite it.
