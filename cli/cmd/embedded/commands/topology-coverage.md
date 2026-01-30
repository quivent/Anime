# Topology Quality Benchmark v3.0

Execute topology encoding quality analysis using database-backed metrics and precision/recall benchmarking.

## QUICK EXECUTION

### Dashboard (Recommended)
```bash
cd tools/topology-dashboard && npm run dev
```
Navigate to **Topology Analysis → Benchmark Runner**

### CLI
```bash
# Full validation
python3 tools/topology_db.py validate

# Specific checks
python3 tools/topology_db.py validate --check weights
python3 tools/topology_db.py validate --check uncalibrated
python3 tools/topology_db.py validate --check orphans
```

---

## QUALITY METRICS (v3.0)

| Component | Weight | Target | Description |
|-----------|--------|--------|-------------|
| Precision | 30% | >70% | Retrieved concepts match expected |
| Weight Distribution | 25% | 20/40/30/10 | Balanced 15/14/13/12 weights |
| Calibration Coverage | 25% | >80% | Entries with depth > 4 |
| Connectivity | 20% | >0.5 avg | Average relations per concept |

## PRECISION BENCHMARK

Tests whether queries retrieve expected concepts:

### Test Suite (7 default queries)
See `tests/benchmark_queries.json` for the current test suite. Example queries include:

| Query | Expected Concepts |
|-------|------------------|
| "Why doesn't B200 work?" | b200, liquid_cooling, thermal_plasticity |
| "thermal plasticity mechanism" | thermal_plasticity, plasticity_field |
| "identity encoding format" | identity_encoding, association_format |
| "M4 hardware selection" | m4_mac_studio, hardware |
| "lineage format failure" | lineage_format, failed, mechanical_execution |
| "substrate memory research" | neural_substrate, memory, thermal_memory |
| "topology retrieval system" | topology_retrieval, graph_traversal, topologist |

### Precision Calculation
```
Precision = |retrieved ∩ expected| / |expected|
Pass if precision >= 0.7 (70%)
```

### Running Benchmark

**Dashboard**:
1. Open Topology Analysis tab
2. Click "Benchmark Runner"
3. Click "Run Benchmark"
4. View pass rate, precision, recall, F1 scores

**CLI** (if implemented):
```bash
python3 tools/topology_db.py benchmark --test-file tests/benchmark_queries.json
```

---

## WEIGHT DISTRIBUTION CHECK

Target distribution:
- **15 (breakthrough)**: 20% of associations
- **14 (important)**: 40% of associations
- **13 (supporting)**: 30% of associations
- **12 (detail)**: 10% of associations

Dashboard shows deviation from targets and suggests corrections.

---

## CALIBRATION COVERAGE

Tracks what percentage of concept-document associations have been properly calibrated (depth > 4).

| Status | Depth Range | Meaning |
|--------|-------------|---------|
| Calibrated | 8-16 | Actually read and assessed |
| Rough | 1-4 | Path heuristic only |

Target: >80% calibrated entries

---

## INTERPRETING RESULTS

### Good Encoding (Ready to Promote)
- Quality Score: ≥85/100
- Precision: ≥70%
- Weight distribution within ±10% of targets
- Calibration: ≥80%

### Needs Work
- Precision <70%: Missing concept mappings
- Weight imbalance: Too many 15s or 12s
- Low calibration: Many depth-1 entries
- Orphans: Concepts with no documents

---

## DETAILED PROTOCOL (Manual Analysis)

### Phase 1: Database Statistics
```bash
python3 tools/topology_db.py stats
```

Output:
```
Concepts: 246 defined, 515 stubs (761 total)
Paths: 753 documents
Associations: 912 concept-document links
Relations: 1,494 concept-concept links
Calibration: 72% (depth > 4)
```

### Phase 2: Weight Distribution
```bash
python3 tools/topology_db.py validate --check weights
```

Output:
```
Weight Distribution:
  15: 184 (20.2%) - target 20% ✓
  14: 352 (38.6%) - target 40% (deviation: -1.4%)
  13: 289 (31.7%) - target 30% (deviation: +1.7%)
  12: 87 (9.5%) - target 10% ✓
```

### Phase 3: Calibration Check
```bash
python3 tools/topology_db.py validate --check uncalibrated
```

Lists entries needing calibration (depth 1-4).

### Phase 4: Orphan Detection
```bash
python3 tools/topology_db.py validate --check orphans
```

Lists concepts with no document associations (stubs).

### Phase 5: Benchmark Run
Use Dashboard Benchmark Runner or:
```bash
# Test queries against encoding
for query in "thermal plasticity" "identity encoding" "B200"; do
  echo "Query: $query"
  python3 tools/topology_db.py query "$query"
  echo "---"
done
```

---

## OUTPUT FORMAT (Dashboard)

```
╔══════════════════════════════════════════════════════════╗
║           TOPOLOGY BENCHMARK RESULTS                      ║
╠══════════════════════════════════════════════════════════╣
║ PRECISION/RECALL                                          ║
║   Pass Rate:          87.5% (7/8 queries)                ║
║   Avg Precision:      78.3%                              ║
║   Avg Recall:         82.1%                              ║
║   Avg F1:             80.1%                              ║
╠══════════════════════════════════════════════════════════╣
║ WEIGHT DISTRIBUTION                                       ║
║   15: 184 (20.2%) ✓                                      ║
║   14: 352 (38.6%) ~                                      ║
║   13: 289 (31.7%) ~                                      ║
║   12: 87 (9.5%) ✓                                        ║
║   Health: GOOD                                            ║
╠══════════════════════════════════════════════════════════╣
║ CALIBRATION                                               ║
║   Calibrated: 657 (72.1%)                                ║
║   Uncalibrated: 255 (27.9%)                              ║
║   Status: NEEDS WORK (target >80%)                       ║
╠══════════════════════════════════════════════════════════╣
║ OVERALL QUALITY:      81.2 / 100                         ║
╚══════════════════════════════════════════════════════════╝
```

---

## METHODOLOGY REFERENCE

- Full methodology: `docs/04-research/PROPER_TOPOLOGIST_PROTOCOL.md`
- Database API: `docs/02-architecture/TOPOLOGIST_DATABASE_INSTRUCTIONS.md`
- Dashboard implementation: `tools/topology-dashboard/src/api/checks/topology.ts`

---

## RELATED COMMANDS

- `/topologist` - Load and query encoding
- `/topologize` - Add/refine entries
- `/topology` - Natural language summary

---

*Benchmark version: 3.0 (database-backed, precision/recall metrics)*
