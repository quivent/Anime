tidy - Quick repository cleanup without deep analysis

Usage: Fast, safe cleanup without full caretaker protocol. Removes obvious clutter, normalizes simple issues, respects everything else.

**Sequential Tidy Protocol:**

🔍 **Phase 1: Detect & Respect**
- Auto-detect project type
- Load `.gitignore`, `.tidyignore`
- Identify safe vs risky operations
- Establish quick baseline

🧹 **Phase 2: Clean (Safe Only)**
Default targets:
- `**/*.tmp`, `**/*.bak`, `**/*.swp`, `**/.DS_Store`
- `__pycache__/`, `*.pyc`, `.pytest_cache/`
- `node_modules/.cache/`, `.next/cache/`
- `target/debug/incremental/` (Rust incremental)
- Empty directories
- Duplicate files (exact match only)

📊 **Phase 3: Report**
- Show what was cleaned
- Report space reclaimed
- Suggest further cleanup if warranted

**Integration Patterns:**

```bash
# Quick safe cleanup
tidy .

# Preview only
tidy . --dry-run

# Include build artifacts
tidy . --artifacts

# Custom patterns
tidy . --include "*.log" --older-than 7d

# Remove exact duplicates (keeps newest)
tidy . --duplicates

# Verbose output showing each file
tidy . --verbose
```

**Options:**
- `--dry-run`: Preview only, no changes
- `--artifacts`: Include stale build directories
- `--duplicates`: Remove exact duplicates (keeps newest)
- `--include <pattern>`: Additional patterns to clean
- `--older-than <duration>`: Age filter (e.g., 7d, 30d)
- `--verbose`: Show each file being processed

**Tidy vs Caretaker:**
| tidy | caretaker |
|------|-----------|
| Fast, no analysis | Deep understanding first |
| Safe operations only | Can restructure |
| No confirmation needed | Interactive by default |
| Clutter removal | Full organization |
| Seconds to run | Minutes to hours |

**When to Use Which:**
- **Use tidy for**: Quick cleanups, CI hygiene, pre-commit tidying
- **Use caretaker for**: Meaningful restructuring, repo rehabilitation

**Quality Standards:**
- ✅ **Speed**: Completes in seconds
- ✅ **Safety**: Only obviously safe operations
- ✅ **Respect**: Honors ignore patterns
- ✅ **Reporting**: Clear summary of actions

Target: $ARGUMENTS

The tidy command delivers fast, safe repository cleanup for everyday use - removing obvious clutter without the deep analysis of a full caretaker pass.
