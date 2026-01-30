Audit code structure with zero tolerance for bullshit. Find the giant files, the panic-happy error handling, the hardcoded paths, the silent failures.

Usage: /linus-code-audit [path] - Analyze code quality issues and generate fix plan

**PROTOCOL ENFORCEMENT: PARALLEL_TASK_ALLOWED**
Phase 2 analysis tasks can run in parallel. Everything else sequential.

**Philosophy:**
"Talk is cheap. Show me the code."
"Bad programmers worry about the code. Good programmers worry about data structures and their relationships."

We're looking for code that will bite you at 3am. Code that panics instead of handling errors. Code that lies about failures. Code that nobody can read in six months.

---

**Phase 1: Measure the Mess**

For each source file, collect:
- Line count
- Function/method count
- Cyclomatic complexity (if tooling available)
- Language

Generate summary:
```
FILE SIZE DISTRIBUTION:
- Files > 1000 lines: X (LIST THEM)
- Files > 500 lines: Y
- Files > 200 lines: Z
- Average file size: N lines

LARGEST FILES (top 10):
1. [path] - [lines] lines
...
```

Files over 500 lines are suspects. Over 1000 is a crime scene.

---

**Phase 2: Find the Sins**

**PARALLEL EXECUTION** - Run these analyses concurrently:

### 2a: Panic Points (Rust)
Find all crash-on-error patterns:
```
grep -rn "\.unwrap()" --include="*.rs"
grep -rn "\.expect(" --include="*.rs"
grep -rn "panic!" --include="*.rs"
```

Categorize:
- **Acceptable**: In tests, in truly impossible cases with comment explaining why
- **Suspect**: In production code paths
- **Unacceptable**: In error handling paths, in initialization, anywhere user input flows

### 2b: Panic Points (Python/JS/Go)
```python
# Python
raise Exception  # without context
except:          # bare except
pass             # in except block (swallowing errors)

# JavaScript
throw new Error  # without message
catch(e) {}      # empty catch

# Go
panic()
_ = err          # ignored error
```

### 2c: Hardcoded Paths
```
grep -rn "/Users/" --include="*.rs" --include="*.py" --include="*.go" --include="*.js" --include="*.ts"
grep -rn "/home/" --include="*.rs" --include="*.py" --include="*.go" --include="*.js" --include="*.ts"
grep -rn "/Volumes/" --include="*.rs" --include="*.py" --include="*.go" --include="*.js" --include="*.ts"
grep -rn "C:\\" --include="*.rs" --include="*.py" --include="*.go" --include="*.js" --include="*.ts"
```

Every hardcoded absolute path is a bug waiting for a different machine.

### 2d: Silent Failures
Look for patterns that hide errors:
```rust
// Rust
.unwrap_or_default()  // in non-obvious places
.ok()                 // discarding error info
let _ = ...           // explicitly ignoring Result

// Python
except: pass
or None               # hiding failures

// Mock data fallbacks
unwrap_or_else(|| { /* create fake data */ })
```

### 2e: God Functions
Find functions over 50 lines. List them. They're doing too much.

### 2f: TODO/FIXME Archaeology
```
grep -rn "TODO" --include="*.rs" --include="*.py" --include="*.go" --include="*.js" --include="*.ts"
grep -rn "FIXME" --include="*.rs" --include="*.py" --include="*.go" --include="*.js" --include="*.ts"
grep -rn "HACK" --include="*.rs" --include="*.py" --include="*.go" --include="*.js" --include="*.ts"
```

TODOs older than 90 days are lies. Either do them or delete them.

---

**Phase 3: Assess the Damage**

Categorize issues by severity:

**CRITICAL** - Will cause production failures:
- Panics in error handling paths
- Silent data corruption (mock fallbacks)
- Hardcoded paths that won't exist elsewhere

**HIGH** - Will cause maintenance nightmares:
- Files over 1000 lines
- Functions over 100 lines
- Bare excepts / empty catches

**MEDIUM** - Should fix but won't explode:
- Files over 500 lines
- Unwraps that could be handled
- Ancient TODOs

**LOW** - Cleanup when bored:
- Style inconsistencies
- Missing documentation
- Dead code

---

**Phase 4: Generate the Verdict**

```
CODE AUDIT: [project name]
==========================

SUMMARY:
- Total source files: X
- Total lines of code: Y
- Files over 500 lines: Z (PROBLEM)
- Panic points (unwrap/expect): N
- Hardcoded paths: M
- Silent failures: P

CRITICAL ISSUES:
1. [file:line] - [issue] - [why it's critical]
2. ...

HIGH PRIORITY:
1. [file:line] - [issue]
2. ...

GIANT FILES REQUIRING SPLIT:
1. [file] - [lines] lines
   Suggested splits:
   - [functions A,B,C] → new_module_1.rs
   - [functions D,E,F] → new_module_2.rs

PANIC POINTS REQUIRING FIX:
[file:line] - unwrap() on user input path
[file:line] - expect() in initialization without recovery

HARDCODED PATHS:
[file:line] - "/Volumes/Lexar/..." → should use config/env

SILENT FAILURES:
[file:line] - mock data fallback hides real errors

FIX PRIORITY ORDER:
1. Critical issues (will break in production)
2. Giant files (will break your sanity)
3. Panic points (will break at 3am)
4. Hardcoded paths (will break on other machines)
5. Everything else
```

---

**Language-Specific Checks:**

**Rust:**
- `clippy` warnings
- Missing `?` propagation
- `Arc<Mutex<>>` everywhere (concurrency smell)

**Python:**
- Bare `except:`
- `type: ignore` comments (hiding type errors)
- Global state

**TypeScript/JavaScript:**
- `any` type usage
- Missing null checks
- Callback hell

**Go:**
- Ignored errors (`_ = err`)
- Naked returns
- `panic()` in library code

---

**MANDATORY WORKFLOW:**
1. Create TodoWrite with all 4 phases
2. Execute Phase 1 (measurement)
3. Launch parallel Task agents for Phase 2a-2f
4. Execute Phase 3-4 sequentially
5. Output the verdict

Target: $ARGUMENTS
