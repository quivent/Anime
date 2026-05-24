# /port-sixth — Three-way Sixth → Lithos ARM64 port

Port Sixth functions with mandatory three-way benchmarks AND mandatory ngram recording.
No shortcuts. No skipped steps. No "candidates." A byte or it didn't happen.

## Input

`$ARGUMENTS` = `<package>/<file.fs> [function]` or batch path.

## Protocol (sequential gates, parallel phases)

### GATE 0: Claim
- Check `/tmp/lithos-port-wire/claims` — skip if claimed
- Append claim. Read `.fs`. Identify target function(s).

### PHASE 1: Write (3 files, parallel)

1. **bench-sixth.fs** — MANDATORY. Variables BEFORE create/allot. Word `main`, end `0 bye`. No `f.`. Compile: `/Users/joshkornreich/bin/sixth bench-sixth.fs bench-sixth`. Debug until it works. NEVER SKIP.

2. **bench.c** — C harness. `noinline` or separate TU for <20-insn functions. `-O2 -ffast-math` for float. `mach_absolute_time`. `asm volatile` barriers. Print Mops/s.

3. **bench_fn.s** — Pure ARM64. No intrinsics, no libm. NEON where width ≥ 4. Post-increment. Callee-save only if needed.

Directory: `bench/packages/<package>/<function>/`

### GATE 1: All three compile
```
time /Users/joshkornreich/bin/sixth bench-sixth.fs bench-sixth
time as -o bench_fn.o bench_fn.s
time gcc -O2 -ffast-math -o run-bench bench.c bench_fn.o
```
Gate fails if ANY fails. Fix before proceeding.

### PHASE 2: Benchmark
```
time ./bench-sixth     → Sixth Mops/s (from wall clock)
./run-bench            → C and Lithos Mops/s
```

### GATE 2: Correctness
C and Lithos outputs must match. Fix bench_fn.s if not.

### PHASE 3: Record (ALL FOUR mandatory, parallel)

1. **results.jsonl** — Append:
   ```json
   {"path":"<file.fs>","func":"<name>","unit":"Mops/s","cRate":N,"sixthRate":N,"lithosRate":N,"ratio":N,"cFlags":"..."}
   ```
   ratio < 0.95 → add `"note":"<reason>. Fix: <path>"`

2. **ngram2.s** — THE TABLE, NOT A COMMENT FILE.
   - Map function's glyph chain to adjacent pairs
   - For EACH pair not in `arch/arm64/arm64-ngram2.s`:
     - Assign next free opcode after current max
     - Add header comment: `//   [0xNN][0xNN] = 0xNN   glyph glyph → name  [tag]`
     - Add `.byte` at correct offset in data section
     - Verify: `as -o /tmp/v.o arch/arm64/arm64-ngram2.s && size -m /tmp/v.o` = 65536
   - A fusion discovered but not written here IS LOST.

3. **findings** — `/tmp/lithos-port-wire/findings`:
   ```
   FINDING:X:<pattern> — <description>. <func>: <ratio>×.
   ```

4. **ledger** — `/tmp/lithos-port-wire/ledger`:
   ```
   DONE:<seat>:<path>:<func>:ratio=<N>
   ```

### GATE 3: Verify recording
- `grep <func> bench/packages/results.jsonl` → exactly 1 line
- If ratio > 1.5×: `grep` ngram2.s for the glyph pair → MUST exist
- `as` ngram2.s → 65536 bytes

### PHASE 4: Adversarial audit
- ratio > 5×: verify C isn't unfairly scalar vs NEON ASM. Add `"audit"` field.
- ratio < 0.8×: ASM is wrong. Fix the ASM, don't document the loss.
- ratio 0.8–0.95×: add `"note"` with root cause and fix path
- Float without `-ffast-math`: recompile. Lithos has no IEEE ordering guarantee.
- C inlined despite noinline: move to separate TU (`gcc -c c_fn.c -o c_fn.o`)

### PHASE 5: Report
```
~/.supervision/report "[<ISO ts>] lithos DONE: <func> — Sixth <N> / C <N> / Lithos <N> Mops/s (<N>×)"
```

## Batch mode

List all `.fs` files in package. Dispatch up to 5 via Agent (model: opus), each running full protocol. After all complete, run GATE 3 on every entry.

## Rules (inviolable)

- Sixth benchmark: NEVER optional
- ngram2 byte: NEVER optional — not a candidate, not a comment, A BYTE
- `-ffast-math` for ALL float benchmarks
- Separate TU for functions under 20 instructions
- bench/ = test cases. arch/arm64/ = product. Fixes flow to font.
- Peephole patterns from findings: reuse CMEQ+SHRN+CLZ, UMULH div10, LDP post-inc, BIC upcase, BFI pack, CSEL branchless, 2-digit unroll
