# /topology - Natural Language Topology Summary v2.0

Provide a natural language summary of the project's knowledge topology from the database.

## EXECUTION

### 1. Load Database Statistics
```bash
python3 tools/topology_db.py stats
```

### 2. Get Domain Summary
```bash
python3 tools/topology_db.py export --format summary
```

### 3. Respond Naturally

Structure your response as a knowledgeable guide who **knows** this project:

---

**What This Project Is:**
[1-2 sentences from your understanding - this is Eigen, a sparse topology neural substrate project]

**Current State:**
- Concepts: [X] defined, [Y] stubs
- Documents: [Z] indexed
- Relations: [N] concept connections
- Calibration: [%] of entries have depth > 4

**Key Concepts I Know:**
- **[Concept]**: [What it means] - Weight [X], documented in [Y] files
- **[Concept]**: [What it means] - Weight [X], documented in [Y] files
- [Continue for 5-10 key concepts]

**What Works (Validated):**
- [Key validated finding from encoding]
- [Another validated finding]

**What Failed (Rejected):**
- [What was tried and why it failed]
- [Another failure with reason]

**Critical Files:**
| Purpose | Path |
|---------|------|
| Project state | PROJECT_STATE.md |
| Hardware | docs/02-architecture/M4_TRANSITION_ARCHITECTURE.md |
| Methodology | docs/04-research/PROPER_TOPOLOGIST_PROTOCOL.md |
| Database API | docs/02-architecture/TOPOLOGIST_DATABASE_INSTRUCTIONS.md |

**How to Query:**
```bash
python3 tools/topology_db.py query "<search>"
python3 tools/topology_db.py docs <concept> --min-weight 14
```

---

## KEY PRINCIPLE

You're not searching or retrieving. You loaded the encoding. You **know** this topology. Speak from understanding, not lookup results.

When you say "I know thermal plasticity is documented in THERMAL_MEMORY.md", you're speaking as the Topologist who has internalized the knowledge graph.

## EXAMPLE RESPONSE

"This is Eigen, a sparse topology neural substrate research project. I know that:

- **B200 GPUs don't work** because liquid cooling defeats thermal plasticity - documented with weight 15 in PROJECT_STATE.md
- The hardware is now **M4 Mac Studio with 128GB** unified memory
- **Thermal plasticity** is the core learning mechanism - heat creates plasticity, cooling consolidates

The encoding contains 246 defined concepts with 1,494 relations. Key domains include:
- Identity research (encoding AI identity as topology)
- Substrate research (GPU-resident neural substrate)
- Training methodology (Socratic finetuning)

To dive deeper, query: `python3 tools/topology_db.py docs thermal_plasticity`"

---

## DATABASE LOCATION

- Primary: `topology.db` (SQLite)
- CLI: `tools/topology_db.py`
- Dashboard: `tools/topology-dashboard/`

$ARGUMENTS
