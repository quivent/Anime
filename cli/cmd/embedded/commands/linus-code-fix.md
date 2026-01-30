Fix the code issues identified by audit. Split giant files, fix error handling, remove hardcoded paths, eliminate silent failures.

Usage: /linus-code-fix [path] [--critical-only|--all] - Execute code fixes

**PROTOCOL ENFORCEMENT: PARALLEL_TASK_ALLOWED**
Independent file fixes can run in parallel. Dependent changes must be sequential.

**Prerequisites:**
- Run /linus-code-audit first
- Have clear list of issues to fix
- Git commit or stash current state

**Philosophy:**
"If you can't read it, you can't fix it. If you can't fix it, it will break."

We fix things in order of how badly they'll hurt you when they fail.

---

**Phase 1: Verify Clean State**

Before touching anything:
```bash
git status  # Should be clean or intentionally dirty
git stash   # If needed
```

Create a branch if working on significant changes:
```bash
git checkout -b linus-fixes
```

Document what we're about to do in commit message or notes.

---

**Phase 2: Fix Critical Issues First**

**SEQUENTIAL** - These affect correctness, do them carefully.

### 2a: Fix Panic Points in Error Paths

Replace crash-on-error with proper handling:

```rust
// BEFORE (crashes)
let config = load_config().expect("Failed to load config");

// AFTER (propagates error)
let config = load_config()
    .map_err(|e| format!("Failed to load config: {}", e))?;

// OR (with context crate)
let config = load_config()
    .context("Failed to load config")?;
```

For initialization code that can't propagate:
```rust
// BEFORE
std::fs::create_dir_all(&path).expect("Failed to create dir");

// AFTER
if let Err(e) = std::fs::create_dir_all(&path) {
    eprintln!("Failed to create directory {}: {}", path.display(), e);
    eprintln!("Please check permissions and try again.");
    std::process::exit(1);
}
```

### 2b: Kill Silent Failures

Replace mock data fallbacks with explicit failures:

```rust
// BEFORE (lies about failure)
let weights = get_weights().unwrap_or_else(|| HashMap::new());

// AFTER (fails honestly)
let weights = get_weights()
    .ok_or("No weights available - is a model loaded?")?;
```

If a fallback is genuinely acceptable, make it LOUD:

```rust
let weights = match get_weights() {
    Some(w) => w,
    None => {
        warn!("No weights available, using empty defaults");
        HashMap::new()
    }
};
```

### 2c: Replace Hardcoded Paths

```rust
// BEFORE
const DATA_DIR: &str = "/Volumes/Lexar/data";

// AFTER
fn data_dir() -> PathBuf {
    std::env::var("APP_DATA_DIR")
        .map(PathBuf::from)
        .unwrap_or_else(|_| {
            dirs::data_local_dir()
                .unwrap_or_else(|| PathBuf::from("."))
                .join("app-name")
        })
}
```

For config files, use:
- Environment variables for deployment-specific paths
- XDG/dirs crate for platform-appropriate defaults
- Relative paths from executable when possible

---

**Phase 3: Split Giant Files**

**PARALLEL** - Independent file splits can happen concurrently.

For each file over 500 lines:

1. **Identify logical groups** - What clusters of functions belong together?
2. **Create new modules** - One concept per file
3. **Move functions** - Keep related code together
4. **Update imports** - Fix all references
5. **Re-export if needed** - Maintain public API

Example split for `commands/mod.rs` (1640 lines):

```
commands/
├── mod.rs              # Re-exports only (~50 lines)
├── generation.rs       # Generation commands
├── adapters.rs         # LoRA adapter commands
├── signals.rs          # Signal commands
├── database.rs         # DB commands
└── models.rs           # Model management
```

The `mod.rs` becomes:
```rust
mod generation;
mod adapters;
mod signals;
mod database;
mod models;

pub use generation::*;
pub use adapters::*;
pub use signals::*;
pub use database::*;
pub use models::*;
```

Each new file: 100-300 lines. If it's bigger, split further.

---

**Phase 4: Fix Remaining Issues**

**PARALLEL** - These are independent fixes.

### 4a: Fix Remaining Unwraps

For each unwrap not in tests:
- If the error is truly impossible, add a comment explaining why
- If it's just lazy, convert to proper error handling
- If it's in a closure where `?` doesn't work, use `.map_err()` and handle outside

### 4b: Clean Up Lock Handling

```rust
// BEFORE
let guard = mutex.lock().unwrap();

// AFTER
let guard = mutex.lock()
    .map_err(|e| format!("Lock poisoned: {}", e))?;

// OR if you want to recover from poisoning
let guard = mutex.lock()
    .unwrap_or_else(|e| e.into_inner());
```

### 4c: Remove Dead TODOs

For each TODO over 90 days old:
- If it's still relevant: create an issue, remove the TODO
- If it's done: remove it
- If it's never going to happen: remove it

TODOs are not a task tracker.

---

**Phase 5: Verify and Test**

After all fixes:

```bash
# Build
cargo build  # or npm build, go build, etc.

# Run tests
cargo test

# Run lints
cargo clippy -- -D warnings

# Check for new issues
/linus-code-audit [path]  # Should show improvement
```

Fix any new issues introduced by the refactoring.

---

**Phase 6: Report Results**

```
CODE FIXES COMPLETE: [project name]
===================================

CHANGES MADE:

Error Handling Fixed:
- [file:line] - unwrap → proper error propagation
- [file:line] - silent failure → explicit error
- [count] total panic points addressed

Files Split:
- [old_file] (1640 lines) → 5 files (avg 300 lines each)
  - commands/generation.rs (280 lines)
  - commands/adapters.rs (320 lines)
  - ...

Hardcoded Paths Removed:
- [file:line] - now uses APP_DATA_DIR env var
- [count] total paths fixed

Dead Code Removed:
- [list of removed TODOs/dead code]

VERIFICATION:
- Build: PASS
- Tests: PASS (X tests)
- Clippy: PASS (0 warnings)

BEFORE/AFTER:
- Panic points: X → Y
- Files over 500 lines: A → B
- Hardcoded paths: M → 0

REMAINING ISSUES (deferred):
- [any issues intentionally left for later]
```

---

**Safety Protocols:**
- Commit after each major change (split, error handling fix)
- Run tests after each change
- If tests fail, fix before continuing
- Don't change behavior - only improve structure and error handling

**Commit Message Format:**
```
refactor(module): split giant file into logical components

- Extracted X into new module
- Extracted Y into new module
- No behavior changes
```

```
fix(error-handling): replace panics with proper error propagation

- Converted X unwrap() calls to Result propagation
- Added context to error messages
- No silent failures remain
```

---

**MANDATORY WORKFLOW:**
1. Create TodoWrite with all 6 phases
2. Execute Phase 1 (verify clean state)
3. Execute Phase 2 sequentially (critical fixes)
4. Launch parallel Task agents for Phase 3 (file splits)
5. Launch parallel Task agents for Phase 4 (remaining fixes)
6. Execute Phase 5-6 sequentially (verify and report)

Target: $ARGUMENTS
