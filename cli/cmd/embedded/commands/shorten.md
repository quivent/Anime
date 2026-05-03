# /shorten

Strip zero-information comments from source code. Automated first, manual second.

Scope: $ARGUMENTS (file, directory, or blank = whole project)

## What dies
- Doc comments restating the item name (`/// Get session ID` on `fn session_id()`)
- Section banners (`// ===============`)
- Inline comments restating the next line (`// Extract DA` before `let da = extract_da()`)
- Module docs >2 lines that could be 1
- Field comments obvious from type+name (`/// The prompt` on `pub prompt: String`)

## What lives
- Comments explaining *why*, not *what*
- Formulas, weights, thresholds (`/// weight = 1.0 + variance_factor * variance`)
- Semantic info not in the name (`// backward compat with old DB`)
- Safety/correctness notes (`// clamp prevents NaN`)
- TODO/FIXME/HACK markers

## Method

### P0: Automated pass
Write/run a script targeting the scope. Pattern: single-line doc comment whose words ⊆ item name words (minus filler). Run `cargo check` (Rust) or equivalent. Zero errors required.

### P1: Manual sweep
Read each file the script touched. Kill:
- Multi-line restating docs the script missed
- Verbose inline comments on obvious code
- Module docs that can compress (17 ln → 1 ln)

### P2: Structural compress
Look for code patterns exposed by comment removal:
- Near-identical constructors → shared helper
- `&self` methods that don't use `self` → pure fns
- Verbose local var names (`accumulated_text` → `text`)
- Single-line method bodies that don't need braces

### P3: Verify
`cargo check && cargo test` (or lang equivalent). Report:
```
Strip: [scope]
Before: [N] lines across [M] files
After: [N'] lines
Delta: -[D] lines ([P]%)
Errors: 0 | Tests: [pass]/[total]
```

## Principles
- If removing a comment makes the code less clear, the code needs a better name — not a comment
- The script handles 60% of the work; manual handles the rest
- Never strip comments from files you haven't read
