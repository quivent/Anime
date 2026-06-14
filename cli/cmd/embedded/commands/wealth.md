---
description: Convergent Portfolio wealth system status and actions
argument-hint: [status|file|license|grant|product|calendar|crisis|dashboard]
---

Wealth Protocol System -- Convergent Portfolio

**Subcommand:** $ARGUMENTS

Based on the subcommand, execute the corresponding section below. If no argument or "status" is given, run the **status** section.

---

## status (default when no argument given)

Run ALL of the following queries and present results in a structured report:

### 1. Patent Inventory
```bash
sqlite3 ~/.mercenary/patents.db "
SELECT '--- INVENTION SUMMARY ---';
SELECT 'Total inventions: ' || COUNT(*) FROM inventions;
SELECT 'By status:';
SELECT '  ' || status || ': ' || COUNT(*) FROM inventions GROUP BY status;
SELECT '';
SELECT 'By domain:';
SELECT '  ' || COALESCE(domain,'(none)') || ': ' || COUNT(*) FROM inventions GROUP BY domain;
SELECT '';
SELECT 'By priority tier:';
SELECT '  Tier ' || COALESCE(priority_tier,'?') || ': ' || COUNT(*) FROM inventions GROUP BY priority_tier;
SELECT '';
SELECT '--- FILINGS ---';
SELECT 'Patent applications: ' || COUNT(*) FROM patent_applications;
SELECT '  By status: ' || COALESCE(status,'(none)') || ' = ' || COUNT(*) FROM patent_applications GROUP BY status;
SELECT '';
SELECT '--- LICENSING ---';
SELECT 'Licensing deals: ' || COUNT(*) FROM licensing_deals;
SELECT '  By status: ' || COALESCE(status,'(none)') || ' = ' || COUNT(*) || ', value = $' || COALESCE(SUM(annual_value),0) FROM licensing_deals GROUP BY status;
SELECT '';
SELECT '--- GRANTS ---';
SELECT 'Grant applications: ' || COUNT(*) FROM grants;
SELECT '  By status: ' || COALESCE(status,'(none)') || ' = ' || COUNT(*) || ', total = $' || COALESCE(SUM(amount),0) FROM grants GROUP BY status;
SELECT '';
SELECT '--- PRODUCTS ---';
SELECT 'Products: ' || COUNT(*) FROM products;
SELECT '  ' || name || ' [' || status || '] MRR=$' || COALESCE(mrr,0) || ' ARR=$' || COALESCE(arr,0) FROM products;
"
```

### 2. Crisis Triggers
```bash
cat ~/.mercenary/crisis-status.json | python3 -c "
import json, sys
d = json.load(sys.stdin)
print('Overall posture:', d['overall_posture'])
for name, t in d['triggers'].items():
    icon = {'green':'GREEN','yellow':'YELLOW','red':'RED'}.get(t['status'],'?')
    print(f'  {name}: {icon} ({t[\"level\"]}) -- {t[\"notes\"]}')
print()
print('Recommended actions:')
for a in d['recommended_actions']:
    print(f'  - {a}')
"
```

### 3. Execution Documents
```bash
echo "Execution documents:" && ls ~/protocols/revenue/wealth/execution/ | wc -l | tr -d ' ' && echo "Provisional drafts:" && ls ~/protocols/revenue/wealth/execution/provisional-inv*.md 2>/dev/null | wc -l | tr -d ' ' && echo "Licensing approaches:" && ls ~/protocols/revenue/wealth/execution/licensing-outreach/*.md 2>/dev/null | wc -l | tr -d ' '
```

### 4. Next Calendar Items
Read `~/protocols/revenue/wealth/execution/master-calendar-12mo.md` and extract the current month's section. Show only CRITICAL and IMPORTANT items (the unchecked `- [ ]` lines) for the current month and the next month.

### 5. CRM Pipeline
```bash
sqlite3 ~/.mercenary/crm.db "
SELECT 'Client pipeline:';
SELECT '  ' || stage || ': ' || COUNT(*) FROM clients GROUP BY stage;
SELECT 'Pending follow-ups: ' || COUNT(*) FROM follow_ups WHERE status='PENDING';
"
```

### 6. Rate & Revenue
```bash
python3 -c "
import json
r = json.load(open('$HOME/.mercenary/rate_ladder.json'))
print(f'Current rate: \${r[\"current_rate\"]}/hr (Tier {r[\"current_tier\"]}: {r[\"tiers\"][r[\"current_tier\"]-1][\"name\"]})')
print(f'Next tier trigger: {r[\"tiers\"][min(r[\"current_tier\"], len(r[\"tiers\"])-1)][\"trigger\"]}')
"
```

Present everything as a single unified dashboard. Use plain text tables or aligned output -- no HTML.

---

## file

Patent filing workflow. Run these queries and present results:

### Top Unfiled Inventions
```bash
sqlite3 ~/.mercenary/patents.db "
SELECT id, name, COALESCE(score,'--') as score, domain, 'Tier ' || COALESCE(priority_tier,'?')
FROM inventions
WHERE status = 'UNFILED'
ORDER BY score DESC, priority_tier ASC
LIMIT 15;
"
```

### Existing Provisional Drafts
```bash
echo "--- Drafted provisionals in execution/ ---"
ls ~/protocols/revenue/wealth/execution/provisional-inv*.md 2>/dev/null
echo ""
echo "--- Batch 2 provisionals ---"
ls ~/protocols/revenue/wealth/execution/provisionals-batch2/ 2>/dev/null
```

### Filing Checklist
Read `~/protocols/revenue/wealth/execution/filing-checklist-template.md` and display it.

### Filing Cost Estimate
```bash
sqlite3 ~/.mercenary/patents.db "
SELECT 'Unfiled inventions: ' || COUNT(*) FROM inventions WHERE status='UNFILED';
SELECT 'At \$64/provisional (micro-entity): \$' || (COUNT(*) * 64) FROM inventions WHERE status='UNFILED';
SELECT 'Filed so far: ' || COUNT(*) FROM patent_applications;
SELECT 'Total filing costs paid: \$' || COALESCE(SUM(amount),0) FROM filing_costs;
"
```

Present a prioritized filing plan: which 5 to file next based on score, and whether drafts exist.

---

## license

Licensing pipeline status. Run these queries and present results:

### Active Deals
```bash
sqlite3 ~/.mercenary/patents.db "
SELECT '--- LICENSING PIPELINE ---';
SELECT 'Total deals: ' || COUNT(*) FROM licensing_deals;
SELECT '';
SELECT 'By status:';
SELECT '  ' || status || ': ' || COUNT(*) || ' (value: \$' || COALESCE(SUM(annual_value),0) || '/yr)' FROM licensing_deals GROUP BY status;
SELECT '';
SELECT 'Total pipeline value: \$' || COALESCE(SUM(annual_value),0) || '/yr' FROM licensing_deals;
"
```

### Outreach Templates
```bash
echo "--- Available outreach approaches ---"
ls ~/protocols/revenue/wealth/execution/licensing-outreach/
echo ""
echo "--- Template library ---"
ls ~/.mercenary/templates/
```

### Auction Lots
```bash
sqlite3 ~/.mercenary/patents.db "
SELECT 'Auction lots: ' || COUNT(*) FROM auction_lots;
SELECT '  ' || id || ': ' || name || ' (' || invention_count || ' inventions, reserve \$' || COALESCE(reserve_price,0) || ') [' || status || ']' FROM auction_lots;
"
```

### Standards Submissions (FRAND)
```bash
sqlite3 ~/.mercenary/patents.db "
SELECT 'Standards submissions: ' || COUNT(*) FROM standards_submissions;
SELECT '  ' || standards_body || '/' || committee || ': ' || submission_type || ' [' || status || ']' FROM standards_submissions;
"
```

### Next Licensing Action
Read `~/protocols/revenue/wealth/execution/cross-license-nvidia-strategy.md` first 20 lines for context, and read `~/protocols/revenue/wealth/execution/frand-standards-roadmap.md` first 20 lines. Summarize the next concrete licensing action to take.

---

## grant

Government grants pipeline. Run these queries and present:

### Grant Applications
```bash
sqlite3 ~/.mercenary/patents.db "
SELECT '--- GRANTS PIPELINE ---';
SELECT 'Total applications: ' || COUNT(*) FROM grants;
SELECT '';
SELECT 'By status:';
SELECT '  ' || status || ': ' || COUNT(*) || ' (\$' || COALESCE(SUM(amount),0) || ')' FROM grants GROUP BY status;
SELECT '';
SELECT 'By agency:';
SELECT '  ' || COALESCE(agency,'?') || ': ' || COUNT(*) FROM grants GROUP BY agency;
"
```

### SAM.gov Status
Read `~/protocols/revenue/wealth/execution/sam-gov-registration.md` first 30 lines.

### SBIR Draft
Read `~/protocols/revenue/wealth/execution/sbir-nsf-phase1-draft.md` first 30 lines for the current pitch status.

### DARPA I2O
Read `~/protocols/revenue/wealth/execution/darpa-i2o-white-paper.md` first 20 lines.

### Next Grant Deadline
Extract the next grant-related deadline from `~/protocols/revenue/wealth/execution/master-calendar-12mo.md` (search for SBIR, DARPA, NSF, grant).

---

## product

Product pipeline status. Run these queries and present:

### All Products
```bash
sqlite3 ~/.mercenary/patents.db "
SELECT '--- PRODUCT PIPELINE ---';
SELECT 'Total products: ' || COUNT(*) FROM products;
SELECT '';
SELECT name || ' [' || status || ']' || ' MRR=\$' || COALESCE(mrr,0) || ' ARR=\$' || COALESCE(arr,0) || ' Customers=' || COALESCE(customers,0)
FROM products
ORDER BY id;
"
```

### Product Specs
List and summarize the product spec documents:
```bash
ls ~/protocols/revenue/wealth/execution/*-product-spec.md 2>/dev/null
ls ~/protocols/revenue/wealth/execution/*-spec.md 2>/dev/null
```

Read the first 10 lines of each product spec found to extract the product name and one-line summary.

### Revenue Model
Read `~/protocols/revenue/wealth/execution/revenue-model-5yr.md` first 40 lines for the 5-year projection.

### Next Product Milestone
Extract the next product-related action from the master calendar for the current month.

---

## calendar

Show this month's and next month's actions from the master calendar.

Read `~/protocols/revenue/wealth/execution/master-calendar-12mo.md` fully. Identify which month maps to the current date (today is provided by the system). Display:

1. **This month's** CRITICAL and IMPORTANT items (all `- [ ]` lines under those headings)
2. **Next month's** CRITICAL and IMPORTANT items
3. The cost estimate and revenue target lines for both months
4. Any items that are overdue (from prior months that should have been done by now)

Format as a clean checklist grouped by priority.

---

## crisis

Show current crisis trigger status. Run:

```bash
python3 -c "
import json, sys

d = json.load(open('$HOME/.mercenary/crisis-status.json'))
print('=' * 60)
print('CRISIS MONITOR -- Convergent Portfolio')
print('=' * 60)
print(f'Timestamp: {d[\"timestamp\"]}')
print(f'Overall posture: {d[\"overall_posture\"]}')
print()

status_icon = {'green': 'GREEN', 'yellow': 'YELLOW', 'red': 'RED'}

for name, t in d['triggers'].items():
    icon = status_icon.get(t['status'], '?')
    print(f'  [{icon}] {name.upper()}')
    print(f'         Level: {t[\"level\"]}')
    print(f'         Notes: {t[\"notes\"]}')
    print(f'         Response: {t[\"response_if_triggered\"]}')
    print(f'         Indicators:')
    for ind in t['indicators']:
        print(f'           - {ind}')
    print()

print('RECOMMENDED ACTIONS:')
for i, a in enumerate(d['recommended_actions'], 1):
    print(f'  {i}. {a}')
print()
print('=' * 60)
"
```

Also read the crisis monitor script for reference:
```bash
head -30 ~/protocols/revenue/wealth/execution/crisis-monitor.sh
```

If any trigger is YELLOW or RED, prominently flag it and show the response protocol.

---

## dashboard

Open the HTML dashboard in the default browser:

```bash
open ~/protocols/revenue/wealth/execution/convergent-dashboard.html
```

Then confirm it was opened.

---

## Data Sources Reference

When any query returns empty results, note it as "No records yet -- pre-filing phase" rather than treating it as an error. The system is in pre-crisis positioning mode.

| Source | Path | Contents |
|--------|------|----------|
| Patent DB | `~/.mercenary/patents.db` | 70 inventions, 9 tables |
| CRM DB | `~/.mercenary/crm.db` | Client pipeline, contracts, follow-ups |
| AB Tracker | `~/.mercenary/ab-tracker.db` | Proposal experiments |
| Crisis Status | `~/.mercenary/crisis-status.json` | 3 trigger monitors |
| Rate Ladder | `~/.mercenary/rate_ladder.json` | Current billing tier |
| Invariants | `~/.mercenary/invariants.json` | System constraints |
| Execution Docs | `~/protocols/revenue/wealth/execution/` | 40+ strategy documents |
| Calendar | `~/protocols/revenue/wealth/execution/master-calendar-12mo.md` | 12-month plan |
| Templates | `~/.mercenary/templates/` | Outreach email templates |
| Licensing | `~/protocols/revenue/wealth/execution/licensing-outreach/` | 7 approach strategies |
