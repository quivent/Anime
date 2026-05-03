---
description: Structured methodology for debugging self-hosting compiler issues. Prevents aimless wandering.
---

# SELF-HOSTING DEBUG PROTOCOL

**You are now operating under a strict debugging methodology. Random exploration is FORBIDDEN.**

## The Core Truth

Self-hosting bugs have ONE cause: **the compiler generates wrong code for some construct that appears in the compiler itself.**

The methodology is: **Differential Testing + Binary Search**

Nothing else. No "investigating." No "exploring." No "let me check."

---

## PHASE 1: Establish the Symptom (5 minutes MAX)

Before ANY debugging, answer these THREE questions:

1. **What exactly happens?**
   - Crash (with what error?)
   - Wrong output (what output, what expected?)
   - Hang (at what point?)
   - Compile error (what error?)

2. **At what stage?**
   - Compiling sixth.fs with the self-compiled compiler?
   - Running the self-compiled binary?
   - Compiling test code with the self-compiled compiler?
   - Something else?

3. **State the symptom in ONE sentence.**

```
SYMPTOM: ________________________________________________
```

**IF YOU CANNOT WRITE ONE SENTENCE, YOU DON'T UNDERSTAND THE PROBLEM. STOP AND CLARIFY.**

### Phase 1 Checkpoint

Use `TodoWrite` to record:
```
[ ] Symptom established: [one sentence]
[ ] Failure stage identified: [stage]
[ ] Oracle comparison set up
```

---

## PHASE 2: Create the Comparison Oracle

You have TWO compilers. Use BOTH for every test.

| Compiler | Command | Status |
|----------|---------|--------|
| **C-compiler** (oracle) | `./engine/fifth compiler/sixth.fs INPUT OUTPUT` | Known Good |
| **Self-compiler** (suspect) | `./OUTPUT_FROM_ABOVE INPUT OUTPUT2` | Under Test |

**Every single test runs through BOTH compilers.**
**Divergence = Bug Location.**

### Helper Script (create if needed)

```bash
#!/bin/bash
# compare-compilers.sh - Run same input through both compilers
INPUT="$1"
C_OUT="/tmp/c-compiled"
SELF_OUT="/tmp/self-compiled"

# C-compiled compiler
./engine/fifth compiler/sixth.fs "$INPUT" "$C_OUT" 2>/dev/null && "$C_OUT"
C_EXIT=$?
C_OUTPUT=$("$C_OUT" 2>&1)

# Self-compiled compiler
./self-compiled "$INPUT" "$SELF_OUT" 2>/dev/null && "$SELF_OUT"
SELF_EXIT=$?
SELF_OUTPUT=$("$SELF_OUT" 2>&1)

echo "=== C-COMPILER ==="
echo "Exit: $C_EXIT"
echo "Output: $C_OUTPUT"
echo ""
echo "=== SELF-COMPILER ==="
echo "Exit: $SELF_EXIT"
echo "Output: $SELF_OUTPUT"
echo ""
if [ "$C_EXIT" = "$SELF_EXIT" ] && [ "$C_OUTPUT" = "$SELF_OUTPUT" ]; then
    echo "MATCH"
else
    echo "DIVERGENCE DETECTED"
fi
```

---

## PHASE 3: Binary Search for Failing Construct

**DO NOT debug the whole compiler. Find the SMALLEST failing program.**

### Step 3.1: Establish Baseline

```bash
# Does the trivial program work?
echo ': main 42 ; main' > /tmp/trivial.fs
./compare-compilers.sh /tmp/trivial.fs
```

**If trivial fails:** The bug is in basic code generation. Simplify further.
**If trivial passes:** The bug is triggered by something more complex. Proceed to 3.2.

### Step 3.2: Binary Search the Compiler Source

If full sixth.fs fails but simple programs work:

1. **Split sixth.fs in half** (by line count or by logical section)
2. **Test the first half** (add a minimal main if needed)
3. **If passes:** Bug is in second half
4. **If fails:** Bug is in first half
5. **Recurse** until you find THE ONE WORD that triggers the bug

### Step 3.3: Document the Minimal Case

```
MINIMAL FAILING CASE:
─────────────────────
[paste 1-10 lines of code here]
```

### Phase 3 Checkpoint

Use `TodoWrite` to record:
```
[ ] Trivial program tested
[ ] Binary search completed
[ ] Minimal reproducer: [paste code or file path]
```

---

## PHASE 4: Differential Disassembly

You now have a minimal failing case. Compare the generated machine code.

```bash
# Compile same source with both compilers
./engine/fifth compiler/sixth.fs minimal.fs /tmp/good
./self-compiled minimal.fs /tmp/bad

# Extract and compare code sections
objdump -d /tmp/good > /tmp/good.asm
objdump -d /tmp/bad > /tmp/bad.asm

# Find first divergence
diff /tmp/good.asm /tmp/bad.asm | head -100
```

### What You're Looking For

| Divergence Type | Likely Cause |
|-----------------|--------------|
| Different instruction | Wrong opcode generation |
| Different immediate | Wrong constant encoding |
| Different register | Wrong register allocation |
| Missing instruction | Dropped operation |
| Extra instruction | Spurious generation |
| Different address | Wrong offset calculation |

### Phase 4 Checkpoint

```
FIRST DIVERGENCE:
─────────────────
Good: [instruction from good.asm]
Bad:  [instruction from bad.asm]
Location: [address or function name]
```

---

## PHASE 5: Trace the Divergence

Now you know WHAT diverges. Find WHY.

### Step 5.1: Identify the Code Generator Path

Which word/construct in the minimal case generates the divergent code?

### Step 5.2: Add Targeted Debug Output

Add `.` or `emit` statements ONLY in the specific code generation path that produces the divergent instruction. NOT everywhere.

```forth
\ Example: if gen-lit is suspect
: gen-lit ( n -- )
  ." gen-lit: " dup . cr   \ ADD THIS
  ... rest of gen-lit ... ;
```

### Step 5.3: Compare Traces

```bash
# Run both compilers on minimal case, capture debug output
./engine/fifth compiler/sixth.fs minimal.fs /tmp/good 2>&1 | tee /tmp/trace-good.txt
./self-compiled minimal.fs /tmp/bad 2>&1 | tee /tmp/trace-bad.txt

diff /tmp/trace-good.txt /tmp/trace-bad.txt
```

### Step 5.4: Identify Root Cause

The trace divergence tells you exactly which value went wrong and when.

---

## FORBIDDEN ACTIONS

| DO NOT | DO INSTEAD |
|--------|------------|
| "Let me investigate the whole compiler" | Reduce to 3-line reproducer |
| Run tests hoping one reveals something | Binary search for first failure |
| Read code trying to spot the bug | Diff actual generated machine code |
| Debug memory corruption by inspection | Use AddressSanitizer or Valgrind |
| Spend hours "exploring" | 15-minute timeboxes with specific hypotheses |
| Say "let me check if..." | State hypothesis, predict outcome, test |
| Modify code without a hypothesis | Every change must test a specific theory |

---

## THE 15-MINUTE RULE

**Every 15 minutes, you MUST answer:**

1. What specific hypothesis am I testing?
2. What evidence will confirm or refute it?
3. What did I learn in the last 15 minutes?

**If you cannot answer these: STOP. You are wandering.**

Update your TodoWrite with current hypothesis and findings every 15 minutes.

---

## COMPLETION CRITERIA

You are DONE when you can state:

```
ROOT CAUSE:
───────────
The word [WORD] generates [WRONG THING] instead of [RIGHT THING]
because [SPECIFIC REASON].

FIX:
────
Change [SPECIFIC LOCATION] from [OLD] to [NEW].

VERIFICATION:
─────────────
After fix, minimal case produces identical output from both compilers.
```

---

## BEGIN NOW

1. State your current symptom in ONE sentence
2. Set up the comparison oracle
3. Begin binary search

**No wandering. No exploring. Hypothesis → Test → Learn → Repeat.**
