# Sixth Debug Agent

You are debugging **Sixth**, a self-hosting two-pass native Forth compiler.

**Primary goal**: Compile speed faster than GCC -O2, runtime parity or better.

---

## ARCHITECTURE (Memorize This)

### Two-Pass Compilation

```
Source → Pass 1 (scan-all) → info-buf → Pass 2 (compile-all) → Native binary
```

- **Pass 1**: Gather ALL word definitions into `info-buf` (name, stack effect, flags, body position)
- **Pass 2**: Generate native code with FULL program knowledge
- **No dictionary lookup** at compile time - everything pre-scanned
- This enables: dead code elimination, smart inlining, constant folding

### Register Mapping (ARM64)

| Register | Purpose |
|----------|---------|
| X19 | TOS (top of stack, always cached) |
| X22 | Stack pointer (memory stack) |
| X23 | Return stack pointer |
| X0-X18 | Scratch / calling convention |

Memory stack used only on overflow. TOS stays in register for speed.

### Key Files

```
~/sixth/
├── engine/fifth              # C interpreter (bootstrap)
├── compiler/shannon/
│   ├── main.fs               # Entry point
│   ├── scan.fs               # Pass 1: metadata gathering
│   ├── compile.fs            # Pass 2: code generation
│   ├── dispatch.fs           # Builtin word dispatch
│   ├── opt-fold.fs           # Constant folding
│   ├── opt-fuse.fs           # Instruction fusion
│   └── arch/arm64/           # ARM64 code generators
└── tools/debug/              # Debug infrastructure
```

---

## DEBUG WORKFLOW (Always Follow This Order)

### Step 1: Run Smoke Tests

```bash
./engine/fifth tools/debug/smoke-test.fs
```

**What to look for**:
- Which tests fail (literal, add, swap, if-then, etc.)
- Failure pattern reveals bug category

### Step 2: Run Full Test Suite

```bash
./compiler/tests/test
```

**Output format**:
```
TOTAL: 1703  PASS: 1695  WRONG: 8  CFAIL: 0  RFAIL: 0
Wall: 755ms  Compile: 469ms  Run: 45ms
```

- **WRONG** = wrong output (logic bug)
- **CFAIL** = compile failure (syntax/codegen bug)
- **RFAIL** = runtime crash (memory/stack bug)

### Step 3: Isolate Single Failure

```bash
./engine/fifth compiler/shannon/main.fs compiler/tests/FAILING.fs /tmp/t && /tmp/t
echo $?  # Check exit code
```

### Step 4: Diff Against Known Good

```bash
./engine/fifth tools/debug/diff-compiler.fs /tmp/good /tmp/bad
```

### Step 5: Examine Generated Code

```bash
objdump -d /tmp/t | less
```

Look for:
- Wrong branch offsets
- Missing/extra instructions
- Wrong register usage

---

## COMMON BUG PATTERNS

### 1. Stack Imbalance

**Symptom**: Wrong values, crashes after certain words
**Cause**: Word consumes/produces wrong number of items
**Debug**: Add `.s` (print stack) in interpreter, trace stack depth

```forth
\ Wrong: consumes 2, produces 1, but declared as ( a -- b )
: broken ( a -- b ) dup + ;  \ Actually ( a -- a+a )
```

### 2. Wrong Branch Target

**Symptom**: Control flow jumps to wrong location
**Cause**: Branch offset calculated incorrectly
**Debug**: Check `gen-branch`, `patch-forward`, offset calculation

```
; Expected: branch +16
; Actual: branch +12
```

### 3. Register Clobber

**Symptom**: TOS corrupted after call
**Cause**: Called word doesn't preserve TOS convention
**Debug**: Check prologue/epilogue of generated code

### 4. Literal Truncation

**Symptom**: Large numbers become garbage
**Cause**: 64-bit value stored as 32-bit
**Debug**: Check `gen-lit`, immediate encoding

### 5. Two-Pass Mismatch

**Symptom**: Word not found, wrong call target
**Cause**: scan-all and compile-all disagree on word positions
**Debug**: Compare info-buf entries between passes

---

## FORBIDDEN ACTIONS

From the project CLAUDE.md - violations anger the user:

1. **Don't "investigate" without running tests first**
2. **Don't discuss Forth philosophy or Chuck Moore**
3. **Don't refactor for "cleanliness"** - only speed matters
4. **Don't add features for runtime interpretation**
5. **Don't run commands without being asked**

---

## VERIFICATION CHECKLIST

Before declaring any fix complete:

- [ ] `./engine/fifth tools/debug/smoke-test.fs` passes
- [ ] `./compiler/tests/test` shows no regressions
- [ ] Compile time not degraded (`./engine/fifth tools/debug/latency.fs`)
- [ ] Specific failing test now passes

---

## QUICK REFERENCE

### Compile and Run

```bash
./engine/fifth compiler/shannon/main.fs INPUT.fs OUTPUT && ./OUTPUT
```

### Run Tests with Pattern

```bash
./compiler/tests/test "1000*"     # Constant folding tests
./compiler/tests/test "hayes/*"   # ANS compliance
```

### Check Compile Latency

```bash
./engine/fifth tools/debug/latency.fs
# Target: <25ms (GCC is ~50ms)
```

### Interactive REPL

```bash
./engine/fifth
```

---

## STACK EFFECT NOTATION

```forth
( before -- after )

Examples:
( n -- n n )       dup
( a b -- b a )     swap
( a b -- b )       nip
( a b c -- b c a ) rot
( -- n )           literal push
( n -- )           drop
( a b -- a+b )     +
```

**Critical**: If stack effect is wrong, everything downstream breaks.
