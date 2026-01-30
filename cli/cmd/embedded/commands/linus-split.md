Split a giant file into logical modules. One concept per file. No file over 500 lines.

Usage: /linus-split [file_path] - Analyze and split an oversized file

**PROTOCOL ENFORCEMENT: SEQUENTIAL_TODO_REQUIRED**
This is a surgical operation. Sequential execution only.

**Philosophy:**
"If a file is over 500 lines, it's doing too much."

Giant files are where bugs hide. Split them and suddenly you can see what's happening.

---

**Phase 1: Analyze the Beast**

Read the entire file. For each function/struct/class:
- Name
- Line count
- What it does (one sentence)
- What it depends on
- What depends on it

Build a dependency graph. Find the clusters.

Output:
```
FILE ANALYSIS: [filename]
=========================

Size: X lines
Functions: Y
Structs/Classes: Z

CLUSTERS IDENTIFIED:

Cluster 1: [name/theme]
- function_a (50 lines) - does X
- function_b (30 lines) - does Y
- struct_c - data for X and Y
Dependencies: [external deps]
Dependents: [what uses this cluster]

Cluster 2: [name/theme]
...

ORPHANS (don't clearly belong):
- helper_function (10 lines)

CIRCULAR DEPENDENCIES:
- [any problematic cycles]
```

---

**Phase 2: Plan the Split**

For each cluster, propose a new file:
```
SPLIT PLAN:

[old_file.ext] (X lines) → N files

1. [new_module_a.ext] (~Y lines)
   Contains: [list of functions/structs]
   Purpose: [what this module does]

2. [new_module_b.ext] (~Z lines)
   Contains: [list of functions/structs]
   Purpose: [what this module does]

3. [remaining in old_file.ext] (~W lines)
   Contains: [what stays - usually re-exports]
   Purpose: Module root, re-exports public API

DEPENDENCY RESOLUTION:
- [function_x] needs [struct_y] from different cluster
  Solution: [how to handle - shared types module, or reorganize]

PUBLIC API PRESERVATION:
- All current public items remain accessible via same path
- mod.rs/index.ts re-exports as needed
```

Verify:
- No circular dependencies between new modules
- Every file under 500 lines (ideally under 300)
- Public API unchanged

---

**Phase 3: Execute the Split**

For each new module:

1. Create the new file with header comment:
```rust
//! [Module name]
//!
//! [One line description]
//!
//! Extracted from [original_file] during refactoring.
```

2. Move the identified functions/structs
3. Add necessary imports
4. Mark items `pub` as needed for cross-module access

For the original file:
1. Remove moved code
2. Add `mod` declarations for new modules
3. Add `pub use` re-exports to preserve public API

---

**Phase 4: Fix the Imports**

After moving code, imports will be broken. Fix them:

1. Find all files that import from the original module
2. Update import paths if they've changed
3. Verify: `cargo check` / `tsc` / `go build`

If using re-exports properly, external imports shouldn't need to change.

---

**Phase 5: Verify the Surgery**

```bash
# Build
cargo build  # or equivalent

# Test
cargo test

# Verify no regressions
# Run any integration tests
```

Check:
- All tests pass
- No new warnings
- Public API unchanged (same imports work)

---

**Phase 6: Report**

```
FILE SPLIT COMPLETE: [filename]
===============================

BEFORE:
- 1 file, X lines

AFTER:
- N files:
  - [new_module_a.ext]: Y lines
  - [new_module_b.ext]: Z lines
  - [original.ext]: W lines (re-exports only)

VERIFICATION:
- Build: PASS
- Tests: PASS
- Public API: UNCHANGED

DEPENDENCY GRAPH:
[simple ASCII diagram showing new module relationships]

RECOMMENDED COMMIT:
```
refactor([module]): split [filename] into logical components

Extracted:
- [new_module_a]: [purpose]
- [new_module_b]: [purpose]

No behavior changes. Public API preserved.
```
```

---

**Common Patterns:**

**Rust:**
```
// mod.rs becomes:
mod handlers;
mod types;
mod utils;

pub use handlers::*;
pub use types::*;
pub use utils::*;
```

**TypeScript:**
```typescript
// index.ts becomes:
export * from './handlers';
export * from './types';
export * from './utils';
```

**Python:**
```python
# __init__.py becomes:
from .handlers import *
from .types import *
from .utils import *
```

**Go:**
```go
// Keep in same package, just different files
// Go handles this naturally - just split the files
```

---

**Splitting Heuristics:**

1. **By domain** - User handling, data processing, network ops
2. **By layer** - Types, handlers, utils
3. **By dependency** - Things that depend on each other stay together
4. **By volatility** - Things that change together stay together

When in doubt: if you have to scroll to see both ends of related code, it should probably be closer together (same file) or explicitly separate (different files with clear interface).

---

**MANDATORY WORKFLOW:**
1. Create TodoWrite with all 6 phases
2. Execute each phase sequentially
3. Do not proceed to Phase 3 until Phase 2 plan is approved
4. Verify build after Phase 4
5. Output final report

Target file: $ARGUMENTS
