# /reindex - Add to Topology Encoding v2.0

Add new concepts, documents, or relations to the topology database.

## DEPRECATED NOTICE

This command replaces the old `EIGEN_REPO_ENCODING.md` format. All encoding changes now go through the SQLite database.

---

## USAGE

### Add New Concept
```bash
python3 tools/topology_db.py add-concept <name> \
  --description "What this concept represents" \
  --category <category>
```

Categories: `core`, `research`, `identity`, `architecture`, `experiments`, `methodology`

### Add Document to Concept
```bash
python3 tools/topology_db.py add-doc <concept> <path> \
  --weight <12-15> \
  --depth <1-16>
```

### Add Relation Between Concepts
```bash
python3 tools/topology_db.py add-relation <from_concept> <to_concept> \
  --weight <12-15> \
  --type <relation_type>
```

Relation types: `prerequisite`, `implements`, `relates`, `contradicts`

---

## WEIGHT GUIDELINES

| Weight | Meaning | Example |
|--------|---------|---------|
| 15 | Breakthrough/Core | Fundamental paradigm discovery |
| 14 | Important | Key mechanism, major finding |
| 13 | Supporting | Useful context, supporting evidence |
| 12 | Detail | Specific implementation note |

**Target Distribution**: 20%/40%/30%/10% for weights 15/14/13/12

---

## DEPTH GUIDELINES

| Depth | Meaning | When to use |
|-------|---------|-------------|
| 16 | Full read + cross-referenced | Read everything, compared to related |
| 12 | Read thoroughly | Read entire file |
| 8 | Skimmed | Read key sections |
| 4 | Brief scan | Headers and filename only |
| 1 | Path heuristic | Inferred from path, no read |

---

## WORKFLOW

When you discover content not in the encoding:

### 1. Read the File
Understand what it contains and how it relates to existing concepts.

### 2. Check Existing Concepts
```bash
python3 tools/topology_db.py query "<relevant_terms>"
```

### 3. Add or Update

**If concept exists**, add document:
```bash
python3 tools/topology_db.py add-doc existing_concept path/to/file.md --weight 14 --depth 12
```

**If concept is new**, create it:
```bash
python3 tools/topology_db.py add-concept new_concept \
  --description "Description of what this concept means" \
  --category research
```

Then add document and relations:
```bash
python3 tools/topology_db.py add-doc new_concept path/to/file.md --weight 14 --depth 12
python3 tools/topology_db.py add-relation new_concept existing_concept --weight 14 --type relates
```

### 4. Export for Git
```bash
python3 tools/topology_db.py export --format markdown > TOPOLOGIST_ENCODING.md
git add TOPOLOGIST_ENCODING.md topology.db
git commit -m "encoding: Add <concept> from <source>"
```

---

## ANTI-PATTERNS TO AVOID

1. **Encoding facts, not relationships**
   - BAD: Just adding file to random concept
   - GOOD: Adding file AND relationship to related concepts

2. **Skipping depth**
   - If you didn't read it, depth = 1
   - Be honest about calibration status

3. **Over-weighting everything as 15**
   - Most things are 13-14
   - 15 is reserved for paradigm-defining content

4. **Creating duplicate concepts**
   - Query first: `topology_db.py query "<term>"`
   - Use existing concept if semantic match exists

5. **Editing markdown files directly**
   - NEVER edit `TOPOLOGIST_ENCODING.md` manually
   - Always use database CLI, then export

---

## EXAMPLE SESSION

User found useful content in `docs/04-research/memory/NEW_FINDING.md`:

```bash
# Check if related concept exists
python3 tools/topology_db.py query "memory finding"
# → Found: thermal_memory, memory_consolidation

# Add document to existing concept
python3 tools/topology_db.py add-doc thermal_memory \
  docs/04-research/memory/NEW_FINDING.md \
  --weight 14 --depth 12

# Export
python3 tools/topology_db.py export --format markdown > TOPOLOGIST_ENCODING.md
```

---

## DASHBOARD ALTERNATIVE

For visual encoding:
```bash
cd tools/topology-dashboard && npm run tauri dev
```

Navigate to **Topology Analysis → Encoding Workflow** for:
- Visual add relation form
- Weight dropdown with descriptions
- Recent relations table
- Inline editing

---

## CURRENT STATS

Check encoding status:
```bash
python3 tools/topology_db.py stats
```

$ARGUMENTS
