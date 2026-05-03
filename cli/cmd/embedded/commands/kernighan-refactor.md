# /kernighan-refactor

Scope: $ARGUMENTS (no arg = whole project). Strict order: dedup→modularize→compress. Each enables the next.

## P0: Audit
Scan target for: repeats (identical blocks), overloaded fns (>1 job, >50ln), dead code (0 callers, `#[allow(dead_code)]`, deprecated), magic values.
Count lines scanned. Categorize: dupes, overloaded, dead, noise. Estimate savings honestly.
Build+test baseline before changes.

## P1: Dedup
Extract shared abstractions from repeated code:
- N fns varying 1-2 params → macro/generic/HOF
- Repeated validation → shared helper
- Copy-paste HTTP → generic w/ type params
- Same conditional N places → fn returning decision

Rules: abstraction must be smaller than sum of dupes. More lines added than removed = don't extract. Name for what it does. Place near callers, not utils grab-bag.
Build+test. Must pass before P2.

## P2: Modularize
Split overloaded fns into single-responsibility:
- Fn >60ln w/ distinct phases → extract each
- Setup+work+cleanup → extract setup/cleanup
- Inline block w/ own locals → name it
- Fn described with "and" → split at "and"

Rules: only split if extracted fn has clear name. Don't split 1-5ln. Don't split if extraction needs 6+ params. Parent fn reads as TOC after. Net lines ~0; win = locality.
Build+test. Must pass before P3.

## P3: Compress
Rm everything that doesn't earn its place:
- `#[deprecated]`/`@deprecated` w/ 0 callers → rm
- `#[allow(dead_code)]` never ref'd → rm
- Speculative future structs/impls → rm
- Trivial wrappers (body = single call) → inline+rm
- Design notes in code → move to issues
- `_legacy` shims w/ no importers → rm

Rules: grep before rm. Callers exist = not dead (test-only counts; inline at test site). `#[allow(dead_code)]` w/ active roadmap comment → ask first. `// TODO:` comment → ask first.
Build+test. Must pass.

## P4: Report
```
Audit: [n]ln scanned, [n] dupes, [n] overloaded, [n] dead
Dedup: -[n]ln | File | What | Savings |
Modularize: ~0 net | File | What |
Compress: -[n]ln | File | What | Savings |
Total: -[n]ln ([x]%), [n] warnings eliminated, build [pass/fail], tests [p]/[t]
```

## Order
Dedup→Modularize→Compress. Can't modularize before dedup (split same dupe into 2 places). Can't compress before modularize (dead code hides in overloaded fns). Verify between each.
