# Topologize v5.0 - Database-First Encoding

**CURATED, NOT COMPREHENSIVE** - Only encode what a Topologist needs to navigate the project.

## MODES

- **EXPAND** (default): Add new concepts/documents to the database
- **REFINE**: Calibrate existing entries (use `/topologize refine`)
- **VALIDATE**: Check encoding health (use `/topologize validate`)

## DATABASE LOCATION

**Primary**: `topology.db` (SQLite)
**CLI Tool**: `tools/topology_db.py`
**Deprecated**: `TOPOLOGIST_ENCODING.md` files (auto-generated from database)

---

## SCORING RUBRIC

### Weight (12-15 typical)

| Weight | Meaning | Ask Yourself |
|--------|---------|--------------|
| 15 | Breakthrough/Core | Is this fundamental to the paradigm? Would project fail without? |
| 14 | Important | Key mechanism or finding? Connects multiple concepts? |
| 13 | Supporting | Useful context? Supports but doesn't define understanding? |
| 12 | Detail | Specific note that aids recall? Could understand without this? |

**Target Distribution**: 20% weight-15 / 40% weight-14 / 30% weight-13 / 10% weight-12

### Analysis Depth (1-16)

| Depth | Meaning | When to use |
|-------|---------|-------------|
| 16 | Full read + cross-referenced | Read file, compared to related files |
| 12 | Read thoroughly | Read entire file, understood content |
| 8 | Skimmed | Read key sections, understood purpose |
| 4 | Brief scan | Looked at headers, filename context |
| 1 | Path heuristic | Only inferred from path, no read |

---

## MODE: EXPAND (default)

### Phase 1: Assess Current State

```bash
python3 tools/topology_db.py stats
```

Shows:
- Concept count (defined + stubs)
- Document count
- Association count
- Calibration percentage

### Phase 2: Add New Concepts

**Add concept:**
```bash
python3 tools/topology_db.py add-concept <name> \
  --description "What this concept represents" \
  --category <category>
```

**Add document to concept:**
```bash
python3 tools/topology_db.py add-doc <concept> <path> \
  --weight 14 --depth 8
```

**Add concept relation:**
```bash
python3 tools/topology_db.py add-relation <from_concept> <to_concept> \
  --weight 14 --type relates
```

Relation types: `prerequisite`, `implements`, `relates`, `contradicts`

### Phase 3: Export for Git

```bash
# Export human-readable markdown
python3 tools/topology_db.py export --format markdown > TOPOLOGIST_ENCODING.md

# Verify export
head -50 TOPOLOGIST_ENCODING.md
```

### Phase 4: Run Benchmark

```bash
# Python benchmark
python3 tools/topology_coverage.py

# Or use Dashboard benchmark runner (visual)
cd tools/topology-dashboard && npm run tauri dev
```

Target: >70% precision, >85% quality score

---

## MODE: REFINE (`/topologize refine`)

**Purpose:** Calibrate existing entries by reading files and updating weights/depths.

### CRITICAL SAFEGUARDS

1. **NEVER delete concepts** - mark defunct instead
2. **NEVER decrease depth** without re-reading file
3. **ALWAYS preserve relationships** when updating
4. **DOCUMENT changes** for audit trail

### Phase 1: Identify Uncalibrated Entries

```bash
# Find low-depth entries (depth 1-4)
python3 tools/topology_db.py validate --check uncalibrated

# List specific concepts needing work
python3 tools/topology_db.py list --uncalibrated --limit 20
```

### Phase 2: Calibrate Entry

For each entry to refine:

1. **Read the actual file** (resolve path from database)
2. **Apply 6-dimension rubric:**
   - relevance: How useful for current work?
   - quality: How complete/well-written?
   - significance: How important to project?
   - accuracy: How correct/current?
   - recency: Superseded by later work?
   - uniqueness: Duplicate information elsewhere?

3. **Update weight if needed:**
   ```bash
   python3 tools/topology_db.py update-weight <concept> <path> <new_weight>
   ```

4. **Update depth to reflect analysis:**
   ```bash
   python3 tools/topology_db.py update-depth <concept> <path> <new_depth>
   ```

### Phase 3: Export and Verify

```bash
# Export updated encoding
python3 tools/topology_db.py export --format markdown > TOPOLOGIST_ENCODING.md

# Run validation
python3 tools/topology_db.py validate
```

---

## MODE: VALIDATE (`/topologize validate`)

```bash
# Full validation
python3 tools/topology_db.py validate

# Specific checks
python3 tools/topology_db.py validate --check orphans    # Concepts with no docs
python3 tools/topology_db.py validate --check files      # Missing files
python3 tools/topology_db.py validate --check weights    # Distribution check
python3 tools/topology_db.py validate --check uncalibrated  # Low-depth entries
```

---

## COMMON MISTAKES TO AVOID

1. **Encoding facts, not relationships**
   - BAD: `substrate, exists` (fact)
   - GOOD: `substrate → memory_mechanism` (relationship)

2. **Over-weighting**
   - Not everything is a breakthrough (15)
   - Most associations should be 13-14

3. **Forgetting depth**
   - Always include depth when adding entries
   - Depth 1 = "I didn't read this" - be honest

4. **Inconsistent terminology**
   - Pick one term and use it everywhere
   - `thermal_memory` vs `heat_memory` - pick one

5. **Editing markdown directly**
   - Always modify via database CLI
   - Markdown files are auto-generated exports

---

## DASHBOARD ALTERNATIVE

For visual encoding workflow:
```bash
cd tools/topology-dashboard && npm run tauri dev
```

Navigate to **Topology Analysis → Encoding Workflow** for:
- Visual relation editor
- Weight selector with guidelines
- Recent relations table
- Inline weight modification

---

## RELATED COMMANDS

- `/topologist` - Load and query encoding
- `/topology-coverage` - Run coverage benchmark
- `/topology` - Natural language summary
- `/defunct` - Check document status

---

*Command version: 5.0 (database-first architecture)*
