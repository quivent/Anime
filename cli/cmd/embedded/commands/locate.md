# /locate - Document Discovery Protocol v2.0

Find documents by topic using topology-first search (database → embedding → grep fallback).

## PROTOCOL (in order)

### 1. Database Query (Primary)

```bash
# Search concepts semantically
python3 tools/topology_db.py query "$ARGUMENTS"
```

If concepts found:
```bash
# Get documents for matched concepts
python3 tools/topology_db.py docs <concept> --min-weight 13 --resolve-paths
```

### 2. Embedding Fallback

If database returns no matches:
```bash
python3 tools/topology_graph.py query "$ARGUMENTS"
```

### 3. Grep Fallback (Last Resort)

If embedding also fails:
```bash
# Search key locations
grep -r -i "$ARGUMENTS" docs/04-research/ --include="*.md" | head -20
grep -r -i "$ARGUMENTS" identity/repository/ --include="*.md" | head -10
```

### 4. Update Encoding

**Critical**: If grep finds relevant content not in database:
```bash
python3 tools/topology_db.py add-concept <discovered_concept> --description "..."
python3 tools/topology_db.py add-doc <concept> <found_path> --weight 14
```

---

## RESPONSE FORMAT

**Query**: $ARGUMENTS

**Database Results**:
| Concept | Weight | Documents |
|---------|--------|-----------|
| [matched] | [wt] | [paths] |

**Recommended Reading**:
1. **[highest-weight file]** - [why relevant]
2. **[second file]** - [why relevant]

**Action**: [which file to start with]

---

## QUICK REFERENCE INDEX

For common topics, these are authoritative:

| Topic | Primary Document |
|-------|-----------------|
| Project state | `PROJECT_STATE.md` |
| Database instructions | `docs/02-architecture/TOPOLOGIST_DATABASE_INSTRUCTIONS.md` |
| Topology methodology | `docs/04-research/PROPER_TOPOLOGIST_PROTOCOL.md` |
| Hardware architecture | `docs/02-architecture/M4_TRANSITION_ARCHITECTURE.md` |
| Identity encoding | `docs/04-research/identity/ENCODING_V1.3.md` |
| Thermal plasticity | `docs/04-research/memory/THERMAL_MEMORY.md` |
| Experiments | `experiments/docs/EXPERIMENT_INDEX.md` |
| Defunct registry | `identity/repository/DEFUNCT_INDEX.md` |

---

## SEARCH PRIORITY ORDER

1. `PROJECT_STATE.md` - Current authoritative state
2. `identity/repository/` - Encoding and topology specs
3. `docs/04-research/` - Research findings
4. `docs/02-architecture/` - Architecture decisions
5. `experiments/docs/` - Experiment results

---

## EXAMPLE USAGE

```
/locate thermal plasticity
→ Database: thermal_plasticity, heat_memory, substrate_dynamics
→ Primary: docs/04-research/memory/THERMAL_MEMORY.md (weight 15)

/locate encoding format
→ Database: identity_encoding, encoding_format, topology_format
→ Primary: docs/02-architecture/TOPOLOGIST_DATABASE_INSTRUCTIONS.md

/locate why B200 failed
→ Database: b200_rejected, thermal_plasticity, liquid_cooling
→ Primary: PROJECT_STATE.md (search for B200 section)
```

---

## KEY PRINCIPLE

You're not just searching—you're navigating a knowledge topology. The database maps concepts to documents. Use semantic understanding first, grep last.

If you find yourself using grep repeatedly for the same topic, **add it to the encoding**.

$ARGUMENTS
