# /benchmark-lithos-sixth-c — Full three-way benchmark with gated recording

Pick a Sixth function, benchmark it three ways (Sixth, C/GCC, Lithos ARM64), record everything.

## Input
`$ARGUMENTS` = one of:
- `<package>/<file.fs> [function]` — single function
- `<package>` — all unclaimed .fs files in that package
- `next N` — pick N unclaimed functions across all packages
- `all` — every unclaimed function in every package
- `parallel N` suffix — run N Opus agents simultaneously (default 5, max 10)

All batch agents use `model: opus`. Single-function mode runs inline (no agent).

Examples:
```
/benchmark-lithos-sixth-c sixthdb/btree.fs leaf_bsearch       # single, inline
/benchmark-lithos-sixth-c sixthdb parallel 10                  # 10 Opus agents on sixthdb
/benchmark-lithos-sixth-c next 20 parallel 5                   # 5 Opus agents, 4 rounds of 5
/benchmark-lithos-sixth-c all parallel 5                       # all unclaimed, 5 at a time
```

## GATE 0: Discover + Claim + Read

**Discovery** (when `next N` or `all` or package-only):
1. List all `.fs` files under `~/sixth/packages/<pkg>/` AND `<pkg>/lib/` AND `<pkg>/bench/`
2. Read `/tmp/lithos-port-wire/claims` — build claimed set
3. Read `bench/packages/results.jsonl` — build completed set (by path+func)
4. Subtract claimed ∪ completed from all files → unclaimed list
5. For each unclaimed file, `grep -c "^:"` to count functions
6. Sort by function count descending (most functions = most value)
7. Pick top N

**Claim**:
- Append `CLAIMED:<seat>:<path>` to `/tmp/lithos-port-wire/claims`
- Read the `.fs` file from `~/sixth/packages/`
- Identify target function: stack effect, algorithm, dependencies
- Skip files that are pure I/O (exec-capture, mmap, shell calls) — flag as SYSCALL

## GATE 1: Source files exist
Write to `bench/packages/<package>/<function>/`:

1. **bench-sixth.fs** — Variables BEFORE create/allot. `main` word. `0 bye`. No `f.`. Include the function AND its helpers (f32@, upcase, etc).
2. **bench.c** — `noinline` or separate TU for <20 insns. `-O2 -ffast-math` for float. `mach_absolute_time`. Prevent DCE: `volatile` sink variable OR `asm volatile("" : "+r"(result) :: "memory")` in loop.
3. **bench_fn.s** — Pure ARM64. No intrinsics. NEON where width ≥ 4.

Gate: all three files written.

## GATE 2: All three compile
```
/Users/joshkornreich/bin/sixth bench-sixth.fs bench-sixth
as -o bench_fn.o bench_fn.s
gcc -O2 -ffast-math -o run-bench bench.c bench_fn.o
```
Gate: zero compile errors on all three. Fix before proceeding.

## GATE 3: All three run + correctness
```
time ./bench-sixth     → sixthRate = ITERS / wall_seconds / 1e6
./run-bench            → cRate, lithosRate (printed by harness)
```
Gate: C and Lithos outputs match (integer: exact, float: within 1e-5). Sixth runs without crash.

## GATE 4: results.jsonl recorded
Append exactly one line:
```json
{"path":"<file.fs>","func":"<name>","unit":"Mops/s","cRate":N,"sixthRate":N,"lithosRate":N,"ratio":N,"cFlags":"..."}
```
- ratio < 0.95 → add `"note"` with reason + fix path
- ratio > 5× → add `"audit"` verifying fair baseline
Gate: `grep <func> results.jsonl` returns exactly 1 line with all three rates.

## GATE 5: ngram2.s updated
- Map function to Lithos glyph chain
- For EACH adjacent pair NOT in `arch/arm64/arm64-ngram2.s`:
  - Next free opcode after current max
  - Add comment header line
  - Add `.byte` at correct offset
  - `as -o /tmp/v.o arch/arm64/arm64-ngram2.s` → 65536 bytes
Gate: every glyph pair from this function is in the table. No exceptions.

## GATE 6: Peephole check
Scan bench_fn.s against ALL 16 patterns in the Peephole catalog below.
Gate: all applicable peepholes applied or noted why not.

## GATE 7: Adversarial
- ratio > 5×: C baseline uses NEON intrinsics if ASM uses NEON
- ratio < 0.8×: fix the ASM, re-benchmark, update jsonl
- float without -ffast-math: recompile
- C inlined: separate TU
Gate: ratio is honest.

## GATE 8: Wire + findings
- `/tmp/lithos-port-wire/ledger`: `DONE:<seat>:<path>:<func>:sixth=<N>:c=<N>:lithos=<N>:ratio=<N>`
- `/tmp/lithos-port-wire/findings`: any new pattern as `FINDING:X:<name> — <desc>`
Gate: both files updated.

## GATE 9: Report
```
~/.supervision/report "[<ISO ts>] lithos DONE: <func> — Sixth <N> / C <N> / Lithos <N> Mops/s (<N>×)"
```

## Batch mode
When given a directory, list `.fs` files, dispatch up to 5 via Agent (model: opus), each running gates 0-9. After all complete, verify gates 4, 5, 6, 7 across all entries. Any gate failure → fix in main thread.

## Sixth compiler quirks (agents must know these)
- Variables/values MUST be declared BEFORE `create ... allot` buffers
- `f.` does not exist — use `fdrop` and `0 bye`
- `fdup` inside `begin/while/repeat` can segfault — use `do/loop` with `i s>f` instead
- `value` uses `to`, `variable` uses `! @` — don't mix
- Word `main` is required. End with `0 bye`.
- Compile: `/Users/joshkornreich/bin/sixth bench-sixth.fs bench-sixth`
- Target iteration count: 1-5 second C runtime (not Sixth runtime)

## Correctness tolerance
- Integer: exact match required
- Float: match within 1e-5 relative error
- Sixth output: must not crash (correctness vs C not required — interpreter rounding differs)

## Ngram opcode allocation (CRITICAL)
- Parallel agents MUST NOT assign opcodes independently
- Main thread reads current max opcode from ngram2.s header comments
- If space remains (max < 0xFF): pre-assign ranges to agents
- If space exhausted: agents report pairs only, main thread decides which to commit
- Agent reports discovered pairs back; main thread commits ALL to ngram2.s in one pass
- If no pairs discovered: agent reports "no new ngram pairs"
- If opcode space full: log "ngram2 full, <N> pairs deferred" to findings
- Font opcode lookup: `grep "// 0x.. <glyph>" arch/arm64/arm64-font.s`

## Results.jsonl dedup
- Before appending: `grep '"func":"<name>"' results.jsonl`
- If entry exists: UPDATE the line (replace), don't append a duplicate
- Mandatory fields: path, func, unit, cRate, sixthRate, lithosRate, ratio, cFlags
- If ratio < 0.95: mandatory `"note"` field
- If ratio > 5×: mandatory `"audit"` field

## Font propagation (bench → product)
- If a peephole optimization improved a bench_fn.s blob:
  - Check if the same pattern exists in `arch/arm64/arm64-font.s` or `lithos-bench.s`
  - If yes: apply the same optimization to the font blob
  - If no: note in findings as "bench-only pattern, no font equivalent yet"
- bench/ is test cases. arch/arm64/ is the product. The product ships.

## Non-computational file filter
Skip files where ALL `:` definitions call ONLY:
- create-element, set-attr, append-child, query-selector (DOM)
- exec-capture, exec-pass, mmap-file, munmap-file (syscall)
- open-file, close-file, read-file, write-file (I/O)
Flag as SYSCALL/DOM in claims wire with reason.

## Peephole catalog (complete as of 2026-05-07)
1. FMUL+FADD → FMLA (fused multiply-add)
2. FMUL+FSUB → FMSUB (fused multiply-subtract)
3. byte scan loop → CMEQ+SHRN+CLZ (16B vectorized scan)
4. UDIV → UMULH (multiply-high for constant divisor)
5. MUL+ADD → MADD (multiply-accumulate)
6. FDIV → FRECPE+FRECPS Newton (2-step reciprocal)
7. FSQRT+FDIV → FRSQRTE+FRSQRTS Newton (reciprocal sqrt)
8. digit loop → 2-digit unroll (acc*100+d1*10+d2)
9. ASCII compare → BIC #0x20 (case-insensitive)
10. pack fields → BFI (bitfield insert)
11. branch update → CSEL/CSINC (branchless)
12. endian swap → REV (single instruction)
13. popcount → CNT.16B+UDOT (NEON popcount)
14. char classify → TBL (NEON 16-parallel lookup)
15. mat4 col → DUP+FMLA lane-indexed (broadcast multiply)
16. LDP stride → post-increment LDP [x],#N (saves ADD)

## Short form
```
G0 discover+claim → G1 write 3 files → G2 compile all → G3 run+correct →
G4 jsonl (dedup) → G5 ngram BYTE (main-thread allocated) → G6 peephole (16 patterns) →
G7 adversarial → G8 wire → G9 report
```
