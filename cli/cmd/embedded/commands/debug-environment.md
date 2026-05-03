---
description: Set up an optimized debugging environment for self-hosting compiler work. Reduces friction, increases iteration speed.
---

# DEBUG ENVIRONMENT OPTIMIZATION

This document describes techniques to create the optimal environment for debugging self-hosting compiler issues. The goal: **minimize time between hypothesis and test result**.

---

## 1. CODE ISOLATION SYSTEM

### Problem
The full compiler is 2900+ lines. Most bugs involve a small subset. Loading everything adds noise.

### Solution: Modular Debug Builds

Create isolated, testable modules that can be swapped in/out:

```
compiler/shannon/arch/arm64/
├── asm.fs       (245 lines) - instruction encoding
├── compile.fs   (887 lines) - main compiler logic
├── control.fs   (381 lines) - control flow
├── macho.fs     (265 lines) - binary format
├── prims.fs     (301 lines) - primitive operations
└── stack.fs     (825 lines) - stack management
```

**Action: Create debug stubs**

```bash
# Create minimal test harnesses for each module
mkdir -p compiler/shannon/debug/

# Example: test just asm.fs in isolation
cat > compiler/shannon/debug/test-asm-only.fs << 'EOF'
\ Minimal harness to test asm.fs in isolation
\ Include only what asm.fs needs
include compiler/shannon/arch/arm64/asm.fs

\ Test cases here
: test-movz  0x12345678 x0 movz-imm  code@ $D2800000 = ;
: test-add   x0 x1 x2 add-reg       code@ $8B020020 = ;

test-movz . cr
test-add . cr
EOF
```

**Patch System**

```bash
#!/bin/bash
# patch-module.sh - Swap module versions for testing
MODULE="$1"
VARIANT="$2"  # "debug", "minimal", "instrumented"

ORIG="compiler/shannon/arch/arm64/${MODULE}.fs"
PATCH="compiler/shannon/debug/${MODULE}-${VARIANT}.fs"
BACKUP="compiler/shannon/debug/${MODULE}-backup.fs"

cp "$ORIG" "$BACKUP"
cp "$PATCH" "$ORIG"
echo "Patched $MODULE with $VARIANT variant. Restore with: cp $BACKUP $ORIG"
```

---

## 2. CROSS-REFERENTIAL ANALYSIS

### Problem
Don't know which words call which. Hard to trace data flow.

### Solution: Static Analysis Tools

**Call Graph Generator**

```bash
#!/bin/bash
# gen-call-graph.sh - Extract call relationships
FILE="$1"

echo "digraph calls {"
grep -E '^: [a-zA-Z]' "$FILE" | while read -r line; do
    WORD=$(echo "$line" | sed 's/^: \([^ ]*\).*/\1/')
    # Find words called in this definition (rough heuristic)
    BODY=$(sed -n "/^: $WORD /,/^;/p" "$FILE" | tail -n +1)
    for CALLED in $(echo "$BODY" | grep -oE '\b[a-z][a-z0-9-]*\b' | sort -u); do
        echo "  \"$WORD\" -> \"$CALLED\";"
    done
done
echo "}"
```

**Usage:**
```bash
./gen-call-graph.sh compiler/shannon/arch/arm64/compile.fs > /tmp/calls.dot
dot -Tpng /tmp/calls.dot -o /tmp/calls.png
open /tmp/calls.png
```

**Dependency Matrix**

```bash
#!/bin/bash
# dependency-matrix.sh - Which file depends on which words
for file in compiler/shannon/arch/arm64/*.fs; do
    echo "=== $file ==="
    # Words defined
    echo "DEFINES:"
    grep -E '^: [a-zA-Z]' "$file" | sed 's/^: \([^ ]*\).*/  \1/' | head -20
    # Words used but not defined (external deps)
    echo "REQUIRES:"
    comm -23 \
        <(grep -oE '\b[a-z][a-z0-9-]+\b' "$file" | sort -u) \
        <(grep -E '^: [a-zA-Z]' "$file" | sed 's/^: \([^ ]*\).*/\1/' | sort -u) \
        | head -20
done
```

---

## 3. TEST OPTIMIZATION

### Problem
591 tests. Running all takes too long. Most are irrelevant to any specific bug.

### Solution: Smart Test Selection

**Test Categorization**

```bash
#!/bin/bash
# categorize-tests.sh - Group tests by what they exercise
mkdir -p /tmp/test-categories

# Literal/constant tests
grep -l 'expect:.*[0-9]' tools/arm64-tests/*.fs > /tmp/test-categories/literals.txt

# Control flow tests
grep -l -E 'if|else|then|begin|while|until' tools/arm64-tests/*.fs > /tmp/test-categories/control.txt

# Stack operation tests
grep -l -E 'dup|drop|swap|over|rot' tools/arm64-tests/*.fs > /tmp/test-categories/stack.txt

# Memory tests
grep -l -E '@|!|cells|allot' tools/arm64-tests/*.fs > /tmp/test-categories/memory.txt

# Arithmetic tests
grep -l -E '\+|\-|\*|\/|mod' tools/arm64-tests/*.fs > /tmp/test-categories/arithmetic.txt

echo "Categories created in /tmp/test-categories/"
wc -l /tmp/test-categories/*.txt
```

**Targeted Test Runner**

```bash
#!/bin/bash
# run-category.sh - Run only tests in a category
CATEGORY="$1"
TESTS=$(cat "/tmp/test-categories/${CATEGORY}.txt" 2>/dev/null)

if [ -z "$TESTS" ]; then
    echo "Unknown category: $CATEGORY"
    echo "Available: literals, control, stack, memory, arithmetic"
    exit 1
fi

echo "Running $CATEGORY tests..."
for test in $TESTS; do
    name=$(basename "$test")
    if ./engine/fifth compiler/shannon/arch/arm64/compile.fs "$test" /tmp/t 2>/dev/null && /tmp/t >/dev/null 2>&1; then
        echo "  PASS: $name"
    else
        echo "  FAIL: $name"
    fi
done
```

**Minimal Smoke Test**

```bash
#!/bin/bash
# smoke-test.sh - 10 critical tests, runs in <1 second
CRITICAL_TESTS=(
    "tools/arm64-tests/01-lit.fs"
    "tools/arm64-tests/07-dup.fs"
    "tools/arm64-tests/10-add.fs"
    "tools/arm64-tests/30-if-then.fs"
    "tools/arm64-tests/40-begin-until.fs"
    "tools/arm64-tests/50-colon-def.fs"
    "tools/arm64-tests/60-memory.fs"
)

PASS=0
FAIL=0
for test in "${CRITICAL_TESTS[@]}"; do
    if [ -f "$test" ]; then
        if ./engine/fifth compiler/shannon/arch/arm64/compile.fs "$test" /tmp/t 2>/dev/null && /tmp/t >/dev/null 2>&1; then
            ((PASS++))
        else
            echo "FAIL: $test"
            ((FAIL++))
        fi
    fi
done
echo "Smoke: $PASS pass, $FAIL fail"
```

---

## 4. COMPILATION LATENCY REDUCTION

### Problem
Even at 3ms per compile, hundreds of iterations add up.

### Solution: Eliminate All Overhead

**RAM Disk for Temp Files**

```bash
# macOS - create RAM disk
RAMDISK=$(hdiutil attach -nomount ram://2048)  # 1MB
diskutil erasevolume HFS+ "RAMDisk" $RAMDISK
export TMPDIR=/Volumes/RAMDisk

# All /tmp operations now in RAM
./engine/fifth compiler/shannon/arch/arm64/compile.fs test.fs /Volumes/RAMDisk/out
```

**Pre-Compiled Baseline**

```bash
# Build once, reuse
./engine/fifth compiler/shannon/arch/arm64/compile.fs compiler/shannon/arch/arm64/compile.fs /tmp/baseline-self
chmod +x /tmp/baseline-self

# Now compare against baseline without recompiling C version
```

**Incremental Compilation Cache**

```bash
#!/bin/bash
# cached-compile.sh - Only recompile if source changed
SOURCE="$1"
OUTPUT="$2"
CACHE_DIR="/tmp/compile-cache"
mkdir -p "$CACHE_DIR"

HASH=$(md5 -q "$SOURCE" 2>/dev/null || md5sum "$SOURCE" | cut -d' ' -f1)
CACHED="$CACHE_DIR/$HASH"

if [ -f "$CACHED" ]; then
    cp "$CACHED" "$OUTPUT"
    echo "Cache hit"
else
    ./engine/fifth compiler/shannon/arch/arm64/compile.fs "$SOURCE" "$OUTPUT"
    cp "$OUTPUT" "$CACHED"
    echo "Cache miss, compiled"
fi
```

---

## 5. LOCAL MODEL INTEGRATION

### Problem
Cloud API latency for each question. Context loading overhead.

### Solution: Local LLM for Fast Iteration

**Load Compiler Into Local Model Context**

```bash
#!/bin/bash
# prepare-context.sh - Concatenate all source for local model
cat > /tmp/compiler-context.txt << 'HEADER'
# Sixth ARM64 Compiler - Complete Source
# Use this context to answer questions about the compiler.
# The compiler generates native ARM64 code from Forth-like source.

HEADER

for file in compiler/shannon/arch/arm64/*.fs; do
    echo "### FILE: $file ###" >> /tmp/compiler-context.txt
    cat "$file" >> /tmp/compiler-context.txt
    echo "" >> /tmp/compiler-context.txt
done

echo "Context prepared: $(wc -l < /tmp/compiler-context.txt) lines"
echo "Token estimate: ~$(( $(wc -w < /tmp/compiler-context.txt) * 4 / 3 )) tokens"
```

**MLX Integration (Apple Silicon)**

```bash
# Start MLX server with compiler context pre-loaded
mlx_lm.server --model mlx-community/Llama-3.2-3B-Instruct-4bit \
    --system-prompt "$(cat /tmp/compiler-context.txt)"

# Query via API
curl -X POST http://localhost:8080/v1/chat/completions \
    -H "Content-Type: application/json" \
    -d '{
        "messages": [{"role": "user", "content": "What does gen-lit do?"}],
        "max_tokens": 200
    }'
```

**Pattern Search via Local Model**

```python
#!/usr/bin/env python3
# quick-search.py - Use local model for semantic code search
import subprocess
import json

def ask_local(question, context_file="/tmp/compiler-context.txt"):
    """Query local model about compiler code"""
    with open(context_file) as f:
        context = f.read()

    prompt = f"""Given this compiler source:

{context[:8000]}  # Truncate for small models

Question: {question}
Answer concisely:"""

    # Call local model (adjust for your setup)
    result = subprocess.run(
        ["mlx_lm.generate", "--prompt", prompt, "--max-tokens", "200"],
        capture_output=True, text=True
    )
    return result.stdout

if __name__ == "__main__":
    import sys
    print(ask_local(sys.argv[1]))
```

---

## 6. AUTOMATED BISECTION

### Problem
Manual binary search is tedious and error-prone.

### Solution: Automated Bisection Scripts

**Source Bisection**

```bash
#!/bin/bash
# bisect-source.sh - Find which lines cause failure
SOURCE="$1"
TEST="$2"

TOTAL=$(wc -l < "$SOURCE")
LOW=1
HIGH=$TOTAL

while [ $((HIGH - LOW)) -gt 5 ]; do
    MID=$(( (LOW + HIGH) / 2 ))

    # Create truncated version
    head -n $MID "$SOURCE" > /tmp/bisect-test.fs
    echo ": main 0 ; main" >> /tmp/bisect-test.fs  # Ensure valid program

    if ./engine/fifth compiler/shannon/arch/arm64/compile.fs /tmp/bisect-test.fs /tmp/t 2>/dev/null; then
        echo "Lines 1-$MID: OK"
        LOW=$MID
    else
        echo "Lines 1-$MID: FAIL"
        HIGH=$MID
    fi
done

echo "Bug introduced between lines $LOW and $HIGH"
sed -n "${LOW},${HIGH}p" "$SOURCE"
```

**Word-Level Bisection**

```bash
#!/bin/bash
# bisect-words.sh - Find which word definition causes failure
SOURCE="$1"

# Extract all word definitions
grep -n '^: ' "$SOURCE" | while IFS=: read -r line word rest; do
    WORD_NAME=$(echo "$word" | sed 's/^: \([^ ]*\).*/\1/')

    # Test with this word commented out
    sed "${line}s/^: /\\\\ DISABLED: /" "$SOURCE" > /tmp/test-without.fs

    if ./engine/fifth compiler/shannon/arch/arm64/compile.fs /tmp/test-without.fs /tmp/t 2>/dev/null; then
        echo "Without $WORD_NAME: PASS (this word may be the problem)"
    fi
done
```

---

## 7. DIFFERENTIAL INFRASTRUCTURE

### Problem
Manual diffing is slow and misses patterns.

### Solution: Pre-Built Comparison Tools

**Instruction-Level Binary Diff**

```bash
#!/bin/bash
# instr-diff.sh - Compare binaries at instruction level
GOOD="$1"
BAD="$2"

# Disassemble both
objdump -d "$GOOD" | grep -E '^\s+[0-9a-f]+:' > /tmp/good.instr
objdump -d "$BAD" | grep -E '^\s+[0-9a-f]+:' > /tmp/bad.instr

# Find first divergence
diff --side-by-side /tmp/good.instr /tmp/bad.instr | head -50

# Summary
echo "---"
echo "Good: $(wc -l < /tmp/good.instr) instructions"
echo "Bad:  $(wc -l < /tmp/bad.instr) instructions"
echo "Diff: $(diff /tmp/good.instr /tmp/bad.instr | grep -c '^[<>]') lines differ"
```

**Semantic Diff (ignores addresses)**

```bash
#!/bin/bash
# semantic-diff.sh - Compare instructions ignoring absolute addresses
GOOD="$1"
BAD="$2"

normalize() {
    objdump -d "$1" | \
    grep -E '^\s+[0-9a-f]+:' | \
    sed 's/[0-9a-f]\{8,16\}/<ADDR>/g' | \
    sed 's/0x[0-9a-f]\+/<HEX>/g'
}

normalize "$GOOD" > /tmp/good.norm
normalize "$BAD" > /tmp/bad.norm

diff /tmp/good.norm /tmp/bad.norm
```

**Visual Diff Tool**

```bash
#!/bin/bash
# vdiff.sh - Side-by-side visual comparison
GOOD="$1"
BAD="$2"

# Create colored diff
objdump -d "$GOOD" > /tmp/good.asm
objdump -d "$BAD" > /tmp/bad.asm

# Use delta or diff-so-fancy if available
if command -v delta &>/dev/null; then
    diff /tmp/good.asm /tmp/bad.asm | delta
elif command -v diff-so-fancy &>/dev/null; then
    diff /tmp/good.asm /tmp/bad.asm | diff-so-fancy
else
    diff --color=always /tmp/good.asm /tmp/bad.asm | less -R
fi
```

---

## 8. MEMORY DEBUGGING TOOLS

### Problem
Memory corruption is hard to detect by inspection.

### Solution: Instrumentation

**AddressSanitizer Build**

```bash
# Rebuild engine with ASan
cd engine
make clean
CFLAGS="-fsanitize=address -g" make
cd ..

# Now memory errors will be caught immediately
./engine/fifth compiler/shannon/arch/arm64/compile.fs test.fs /tmp/t
```

**Valgrind Wrapper**

```bash
#!/bin/bash
# valgrind-compile.sh - Run compilation under valgrind
valgrind --leak-check=full \
         --show-leak-kinds=all \
         --track-origins=yes \
         --error-exitcode=1 \
         ./engine/fifth compiler/shannon/arch/arm64/compile.fs "$1" "$2" 2>&1 | \
    tee /tmp/valgrind.log

grep -E "(Invalid|uninitialised)" /tmp/valgrind.log && echo "MEMORY ERROR DETECTED"
```

---

## 9. STATE SNAPSHOT SYSTEM

### Problem
Hard to compare state at different points in compilation.

### Solution: Checkpoint System

**Add to Compiler: State Dump**

```forth
\ debug-state.fs - Dump compiler state for comparison
: dump-state ( -- )
  ." === COMPILER STATE ===" cr
  ." code-here: " code-here @ . cr
  ." data-here: " data-here @ . cr
  ." info-count: " info-count @ . cr
  ." string-count: " string-count @ . cr
  \ Add more state variables as needed
;

\ Insert at key points
\ : compile-word ... dump-state ... ;
```

**Compare States**

```bash
# Run both compilers with state dumps
./engine/fifth compiler/shannon/arch/arm64/compile.fs test.fs /tmp/good 2>&1 | grep "===" > /tmp/state-good.txt
./self-compiled test.fs /tmp/bad 2>&1 | grep "===" > /tmp/state-bad.txt

diff /tmp/state-good.txt /tmp/state-bad.txt
```

---

## SETUP CHECKLIST

Run this to set up the debug environment:

```bash
#!/bin/bash
# setup-debug-env.sh

echo "Setting up debug environment..."

# 1. Create directories
mkdir -p compiler/shannon/debug
mkdir -p /tmp/test-categories
mkdir -p /tmp/compile-cache

# 2. Create RAM disk (macOS)
if [[ "$OSTYPE" == "darwin"* ]]; then
    RAMDISK=$(hdiutil attach -nomount ram://4096 2>/dev/null)
    if [ -n "$RAMDISK" ]; then
        diskutil erasevolume HFS+ "RAMDisk" $RAMDISK >/dev/null 2>&1
        echo "RAM disk created at /Volumes/RAMDisk"
    fi
fi

# 3. Build baseline
echo "Building baseline compiler..."
./engine/fifth compiler/shannon/arch/arm64/compile.fs compiler/shannon/arch/arm64/compile.fs /tmp/baseline-self 2>/dev/null
chmod +x /tmp/baseline-self 2>/dev/null

# 4. Prepare context for local model
echo "Preparing compiler context..."
cat compiler/shannon/arch/arm64/*.fs > /tmp/compiler-context.txt
echo "Context: $(wc -l < /tmp/compiler-context.txt) lines"

# 5. Categorize tests
echo "Categorizing tests..."
grep -l 'if\|then' tools/arm64-tests/*.fs 2>/dev/null > /tmp/test-categories/control.txt
grep -l '@\|!' tools/arm64-tests/*.fs 2>/dev/null > /tmp/test-categories/memory.txt

echo "Debug environment ready."
echo ""
echo "Available tools:"
echo "  smoke-test.sh        - Quick 10-test sanity check"
echo "  run-category.sh      - Run tests by category"
echo "  bisect-source.sh     - Binary search for failing line"
echo "  instr-diff.sh        - Compare generated code"
echo "  valgrind-compile.sh  - Memory error detection"
```

---

## QUICK REFERENCE

| Task | Command |
|------|---------|
| Smoke test | `./smoke-test.sh` |
| Test category | `./run-category.sh control` |
| Binary search source | `./bisect-source.sh compile.fs test.fs` |
| Compare binaries | `./instr-diff.sh /tmp/good /tmp/bad` |
| Memory check | `./valgrind-compile.sh test.fs /tmp/out` |
| Prepare context | `cat compiler/shannon/arch/arm64/*.fs > /tmp/ctx.txt` |
| Cache compile | `./cached-compile.sh test.fs /tmp/out` |

---

## INTEGRATION WITH /debug-self-hosting

This environment supports the methodology in `/debug-self-hosting`:

1. **Phase 1** (Symptom) → Use `dump-state` to capture exact failure point
2. **Phase 2** (Oracle) → Use pre-built baseline at `/tmp/baseline-self`
3. **Phase 3** (Binary Search) → Use `bisect-source.sh` and `bisect-words.sh`
4. **Phase 4** (Disassembly) → Use `instr-diff.sh` and `semantic-diff.sh`
5. **Phase 5** (Trace) → Add state dumps, compare with `diff`

The environment eliminates friction. The methodology provides structure. Together: fast, systematic debugging.
