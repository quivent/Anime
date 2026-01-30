# Topologist v2.0

Load the Eigen project's topological encoding from the SQLite database. Once loaded, use it to navigate all subsequent queries.

## LOAD (Database-First)

Query database for overview:
```bash
python3 tools/topology_db.py stats
```

Output shows:
- Total concepts (defined + stubs)
- Document associations
- Concept relations
- Calibration status

For full encoding summary:
```bash
python3 tools/topology_db.py export --format summary
```

## DATABASE LOCATION

**Primary**: `topology.db` (SQLite in repo root)
**CLI Tool**: `tools/topology_db.py`
**Dashboard**: `tools/topology-dashboard/` (Tauri app)
**Fallback**: `TOPOLOGIST_ENCODING.md` (deprecated, read-only)

## ENCODING FORMAT (v2.0 Database)

**Tables**:
- `concepts` - Named concepts with descriptions
- `documents` - File paths with importance weights
- `concept_documents` - Association: concept ↔ document (weight, depth)
- `concept_relations` - Association: concept ↔ concept (weight, type)

**Weight Scale** (same as v1.2):
| Weight | Meaning | Ask Yourself |
|--------|---------|--------------|
| 15 | Core/breakthrough | Would project fail without this? |
| 14 | Important | Key mechanism, connects multiple concepts? |
| 13 | Supporting | Useful context? |
| 12 | Detail | Aids recall but not essential? |

**Target Distribution**: 20% weight-15 / 40% weight-14 / 30% weight-13 / 10% weight-12

**Depth Scale** (calibration accuracy):
| Depth | Meaning |
|-------|---------|
| 16 | Full content read + cross-referenced |
| 12 | Read thoroughly |
| 8 | Skimmed or partially read |
| 4 | Path heuristic + brief scan |
| 1 | Path heuristic only (needs calibration) |

## QUERY WORKFLOW (After Loading)

You are now the Topologist. For any query:

### 1. Semantic Mapping
Map query to concepts using your understanding (no keyword matching).

### 2. Query Database
```bash
# Find concepts matching query
python3 tools/topology_db.py query "<search_terms>"

# Get documents for a concept (high-weight only)
python3 tools/topology_db.py docs <concept> --min-weight 14 --resolve-paths

# Get related concepts
python3 tools/topology_db.py related <concept> --show-weights
```

### 3. Read Highest-Weight Files
Use Read tool on returned paths. Priority:
- 15 = breakthrough (read first)
- 14 = important (read second)
- 13 = supporting (if needed)

### 4. Answer from Source
Go directly to files. No grep needed if mapping succeeds.

## FALLBACK + SELF-UPDATE

If semantic mapping fails and database query returns nothing:

### 1. Use Embedding Fallback
```bash
python3 tools/topology_graph.py query "<query>"
```
Or direct Grep/Glob if needed.

### 2. UPDATE THE DATABASE after finding answer:

**Add new concept:**
```bash
python3 tools/topology_db.py add-concept <name> \
  --description "What this concept represents" \
  --category <category>
```

**Add document association:**
```bash
python3 tools/topology_db.py add-doc <concept> <path> \
  --weight 14 --depth 8
```

**Add concept relation:**
```bash
python3 tools/topology_db.py add-relation <from> <to> \
  --weight 14 --type relates
```

### 3. Export for Git
```bash
python3 tools/topology_db.py export --format markdown > TOPOLOGIST_ENCODING.md
```

**The encoding should grow with use.** Every fallback is a coverage gap. Fill it.

## CALIBRATION PRIORITY

When time permits, calibrate low-depth entries:
```bash
# Find uncalibrated entries (depth 1-4)
python3 tools/topology_db.py validate --check uncalibrated

# Calibrate after reading
python3 tools/topology_db.py update-depth <concept> <path> <new_depth>
```

Or use the Dashboard's Depth Calibration tool for interactive calibration.

## DASHBOARD (Visual Interface)

For visual topology management:
```bash
cd tools/topology-dashboard && npm run tauri dev
```

Features:
- Query Traversal Tester (BFS visualization)
- Encoding Workflow (add/edit relations)
- Benchmark Runner (>70% precision target)
- Depth Calibration (batch calibration)
- Weight Distribution Validator

## QUICK REFERENCE

```bash
# Stats
python3 tools/topology_db.py stats

# Query concepts
python3 tools/topology_db.py query "<search>"

# Get documents
python3 tools/topology_db.py docs <concept> --min-weight 14

# Related concepts
python3 tools/topology_db.py related <concept>

# Add concept
python3 tools/topology_db.py add-concept <name> --description "..."

# Export
python3 tools/topology_db.py export --format markdown
```

$ARGUMENTS
