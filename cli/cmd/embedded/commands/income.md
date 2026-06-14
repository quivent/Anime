---
description: Daily income operations -- proposals, tracking, pipeline
argument-hint: [proposals|track|pipeline|leads|report]
---

Daily Income Operations -- until monthly income > $8,000

**Command:** $ARGUMENTS

Based on the subcommand, execute the corresponding section below. If no argument or "proposals" is given, run the **proposals** section.

---

## proposals (default when no argument given)

Today's income actions. Run ALL of the following and present as a single dashboard.

### 1. Financial Position (context for urgency)
```bash
sqlite3 ~/.mercenary/mercenary.db "
SELECT monthly_expenses, current_reserves,
       ROUND(CAST(current_reserves AS REAL) / monthly_expenses, 1) as runway_months,
       minimum_acceptable_rate
FROM financial_config WHERE id=1;
"
```
```bash
bash ~/.mercenary/finance-tracker.sh burn-rate
```
Present runway prominently. If runway < 3 months, flag as URGENT.

### 2. Ready Proposals
```bash
ls ~/protocols/revenue/wealth/execution/ready-proposals/ 2>/dev/null
```
If any `.md` files exist, read each one's first 5 lines and list them as ready to send.

### 3. Cold Outreach Queue
```bash
ls ~/protocols/revenue/wealth/execution/cold-outreach-consulting/ 2>/dev/null
```
List companies with prepared outreach. Note which have NOT yet been sent (cross-reference with approach_log):
```bash
sqlite3 ~/.mercenary/mercenary.db "
SELECT DISTINCT company FROM approach_log
WHERE status IN ('sent', 'responded', 'meeting_scheduled', 'converted');
"
```

### 4. Upwork / Job Board Matches
```bash
sqlite3 ~/.mercenary/mercenary.db "
SELECT sc.id, sc.source, sc.title, sc.company,
       COALESCE(sc.match_tier, 'unscored') as tier,
       ROUND(sc.match_score, 1) as score,
       sc.rate_type, sc.rate_min, sc.rate_max,
       sc.url
FROM scanned_contracts sc
WHERE sc.status = 'new'
AND sc.match_tier IN ('tier_1', 'tier_2')
ORDER BY sc.match_score DESC
LIMIT 15;
"
```
Also show Upwork-specific jobs:
```bash
sqlite3 ~/.mercenary/mercenary.db "
SELECT id, title, budget_min, budget_max, budget_type,
       ROUND(relevance_score, 1) as score, url
FROM jobs
WHERE archived = 0 AND source = 'upwork'
ORDER BY relevance_score DESC
LIMIT 10;
"
```

### 5. Today's Proposal Count vs Target
```bash
sqlite3 ~/.mercenary/mercenary.db "
SELECT
  (SELECT COUNT(*) FROM approach_log
   WHERE date(sent_at) = date('now')) as sent_today,
  (SELECT COUNT(*) FROM approach_log
   WHERE date(sent_at) >= date('now', '-7 days')) as sent_this_week,
  (SELECT COUNT(*) FROM approach_log
   WHERE date(sent_at) >= date('now', '-30 days')) as sent_this_month;
"
```
```bash
sqlite3 ~/.mercenary/mercenary.db "
SELECT COUNT(*) FROM job_applications
WHERE date(applied_at) = date('now');
"
```
Target: 3-5 proposals/day, 15-25/week. Show delta from target.

### 6. Action Items
Based on the data above, generate a numbered list of the 3-5 highest-priority actions for today. Prioritize:
1. Responding to any leads that responded (status='responded' in approach_log)
2. Following up on overdue follow-ups
3. Sending proposals to top-scoring uncontacted contracts
4. Applying to matching Upwork jobs
5. Preparing new outreach for tier 1 contracts not yet approached

---

## track

Quick financial tracking. Parse the remaining arguments after "track":
- `track income <amount> <source> [note]`
- `track expense <amount> <category> [note]`
- `track balance`
- `track monthly`
- `track burn-rate`
- `track log [N]`
- `track runway <savings>`

Run the appropriate finance-tracker.sh command:
```bash
bash ~/.mercenary/finance-tracker.sh <parsed_subcommand> <parsed_args>
```

After any income or expense entry, also run balance to show the updated state:
```bash
bash ~/.mercenary/finance-tracker.sh balance
```

If "track" is given alone with no further args, run both balance and monthly:
```bash
bash ~/.mercenary/finance-tracker.sh balance
bash ~/.mercenary/finance-tracker.sh monthly
```

---

## pipeline

Show the current income pipeline. Run ALL of the following:

### 1. Pipeline Funnel
```bash
sqlite3 ~/.mercenary/mercenary.db "
SELECT
  current_state,
  COUNT(*) as count
FROM pipeline_states
GROUP BY current_state
ORDER BY CASE current_state
  WHEN 'identified' THEN 1
  WHEN 'engaged' THEN 2
  WHEN 'negotiating' THEN 3
  WHEN 'committed' THEN 4
  WHEN 'active' THEN 5
  WHEN 'completed' THEN 6
  WHEN 'terminated' THEN 7
END;
"
```

### 2. Approach Activity
```bash
sqlite3 ~/.mercenary/mercenary.db "
SELECT status, COUNT(*) as count
FROM approach_log
GROUP BY status
ORDER BY CASE status
  WHEN 'drafted' THEN 1
  WHEN 'sent' THEN 2
  WHEN 'viewed' THEN 3
  WHEN 'responded' THEN 4
  WHEN 'meeting_scheduled' THEN 5
  WHEN 'converted' THEN 6
  WHEN 'no_response' THEN 7
  WHEN 'declined' THEN 8
END;
"
```

### 3. Approach Detail (last 14 days)
```bash
sqlite3 -column -header ~/.mercenary/mercenary.db "
SELECT
  al.vector_type,
  al.decision_maker,
  al.channel,
  al.status,
  al.sent_at,
  COALESCE(c.title, sc.title, '(unknown)') as contract_title
FROM approach_log al
LEFT JOIN contracts c ON al.contract_id = c.id
LEFT JOIN scanned_contracts sc ON al.contract_id = CAST(sc.id AS TEXT)
WHERE al.sent_at >= datetime('now', '-14 days')
ORDER BY al.sent_at DESC
LIMIT 20;
"
```

### 4. Web Applications
```bash
sqlite3 -column -header ~/.mercenary/mercenary.db "
SELECT status, COUNT(*) as count
FROM web_applications
GROUP BY status;
"
```

### 5. Job Applications
```bash
sqlite3 -column -header ~/.mercenary/mercenary.db "
SELECT
  ja.outcome,
  COUNT(*) as count,
  ROUND(AVG(ja.predicted_score), 1) as avg_score
FROM job_applications ja
GROUP BY ja.outcome;
"
```

### 6. Contract Scan Pool Summary
```bash
sqlite3 ~/.mercenary/mercenary.db "
SELECT
  match_tier,
  COUNT(*) as count,
  ROUND(AVG(match_score), 1) as avg_score,
  SUM(CASE WHEN status='new' THEN 1 ELSE 0 END) as untouched
FROM scanned_contracts
WHERE match_tier IS NOT NULL
GROUP BY match_tier
ORDER BY match_tier;
"
```

### 7. Conversion Metrics
```bash
sqlite3 ~/.mercenary/mercenary.db "
SELECT
  'Approaches' as stage, COUNT(*) as count FROM approach_log
UNION ALL
SELECT 'Responded', COUNT(*) FROM approach_log WHERE status='responded'
UNION ALL
SELECT 'Meetings', COUNT(*) FROM approach_log WHERE status='meeting_scheduled'
UNION ALL
SELECT 'Converted', COUNT(*) FROM approach_log WHERE status='converted';
"
```
Calculate and display conversion rates between stages.

### 8. Revenue from Finance Tracker
```bash
bash ~/.mercenary/finance-tracker.sh balance
```

Present everything as a single pipeline dashboard with funnel visualization.

---

## leads

Show warm leads that need follow-up. Run ALL of the following:

### 1. Responded Approaches (HIGHEST PRIORITY)
```bash
sqlite3 -column -header ~/.mercenary/mercenary.db "
SELECT
  al.id,
  al.decision_maker,
  al.channel,
  al.vector_type,
  al.sent_at,
  al.response_at,
  COALESCE(c.title, sc.title, '(direct)') as target,
  al.body_preview
FROM approach_log al
LEFT JOIN contracts c ON al.contract_id = c.id
LEFT JOIN scanned_contracts sc ON al.contract_id = CAST(sc.id AS TEXT)
WHERE al.status = 'responded'
ORDER BY al.response_at DESC;
"
```
Flag these as IMMEDIATE ACTION items.

### 2. Overdue Follow-ups
```bash
sqlite3 -column -header ~/.mercenary/mercenary.db "
SELECT
  fs.id,
  fs.scheduled_for,
  fs.follow_up_type,
  al.decision_maker,
  al.channel,
  al.vector_type,
  al.status as approach_status
FROM follow_up_schedule fs
JOIN approach_log al ON fs.approach_log_id = al.id
WHERE fs.status = 'pending'
AND fs.scheduled_for <= datetime('now')
ORDER BY fs.scheduled_for ASC;
"
```

### 3. Upcoming Follow-ups (next 7 days)
```bash
sqlite3 -column -header ~/.mercenary/mercenary.db "
SELECT
  fs.id,
  fs.scheduled_for,
  fs.follow_up_type,
  al.decision_maker,
  al.channel
FROM follow_up_schedule fs
JOIN approach_log al ON fs.approach_log_id = al.id
WHERE fs.status = 'pending'
AND fs.scheduled_for > datetime('now')
AND fs.scheduled_for <= datetime('now', '+7 days')
ORDER BY fs.scheduled_for ASC;
"
```

### 4. Tier 1 Contracts Not Yet Approached
```bash
sqlite3 -column -header ~/.mercenary/mercenary.db "
SELECT
  sc.id,
  sc.source,
  sc.title,
  sc.company,
  ROUND(sc.match_score, 1) as score,
  sc.rate_type,
  sc.rate_min,
  sc.rate_max,
  sc.url
FROM scanned_contracts sc
WHERE sc.match_tier = 'tier_1'
AND sc.status = 'new'
AND sc.id NOT IN (
  SELECT CAST(contract_id AS INTEGER) FROM approach_log
  WHERE contract_id IS NOT NULL
)
ORDER BY sc.match_score DESC;
"
```

### 5. Stale Sent Approaches (sent > 5 days ago, no response)
```bash
sqlite3 -column -header ~/.mercenary/mercenary.db "
SELECT
  al.id,
  al.decision_maker,
  al.channel,
  al.sent_at,
  CAST(julianday('now') - julianday(al.sent_at) AS INTEGER) as days_since
FROM approach_log al
WHERE al.status = 'sent'
AND al.sent_at <= datetime('now', '-5 days')
ORDER BY al.sent_at ASC
LIMIT 15;
"
```
Recommend: mark as no_response or schedule a follow-up.

For each section, if results are empty, say so clearly. Then generate a prioritized action list: respond to warm leads first, handle overdue follow-ups second, approach untouched tier 1 contracts third.

---

## report

Weekly income report. Run ALL of the following:

### 1. Proposals Sent (last 7 days)
```bash
sqlite3 ~/.mercenary/mercenary.db "
SELECT
  date(sent_at) as day,
  COUNT(*) as proposals_sent
FROM approach_log
WHERE sent_at >= datetime('now', '-7 days')
GROUP BY date(sent_at)
ORDER BY day;
"
```
```bash
sqlite3 ~/.mercenary/mercenary.db "
SELECT
  COUNT(*) as total_sent,
  SUM(CASE WHEN status='responded' THEN 1 ELSE 0 END) as responses,
  SUM(CASE WHEN status='meeting_scheduled' THEN 1 ELSE 0 END) as meetings,
  SUM(CASE WHEN status='converted' THEN 1 ELSE 0 END) as conversions
FROM approach_log
WHERE sent_at >= datetime('now', '-7 days');
"
```

### 2. Job Applications (last 7 days)
```bash
sqlite3 ~/.mercenary/mercenary.db "
SELECT COUNT(*) as applications_this_week
FROM job_applications
WHERE applied_at >= datetime('now', '-7 days');
"
```

### 3. Contracts Won
```bash
sqlite3 ~/.mercenary/mercenary.db "
SELECT
  al.decision_maker,
  c.title,
  c.rate_min,
  c.rate_max,
  al.sent_at,
  al.response_at
FROM approach_log al
JOIN contracts c ON al.contract_id = c.id
WHERE al.status = 'converted'
AND al.response_at >= datetime('now', '-7 days');
"
```

### 4. Revenue Earned (this week and this month)
```bash
bash ~/.mercenary/finance-tracker.sh balance
bash ~/.mercenary/finance-tracker.sh monthly
```

### 5. Burn Rate vs Income
```bash
bash ~/.mercenary/finance-tracker.sh burn-rate
```
```bash
sqlite3 ~/.mercenary/mercenary.db "
SELECT monthly_expenses, current_reserves,
       ROUND(CAST(current_reserves AS REAL) / monthly_expenses, 1) as runway_months
FROM financial_config WHERE id=1;
"
```

### 6. Scan Activity
```bash
sqlite3 ~/.mercenary/mercenary.db "
SELECT source, COUNT(*) as new_this_week
FROM scanned_contracts
WHERE scanned_at >= datetime('now', '-7 days')
GROUP BY source
ORDER BY COUNT(*) DESC;
"
```

### 7. Weekly Scorecard
Calculate and present:
- Proposals sent vs target (21/week = 3/day)
- Response rate (responses / proposals sent)
- Conversion rate (converted / responses)
- Revenue earned this week
- Runway remaining
- Gap to $8,000/month target

Present as a clean scorecard. If income this month < $8,000, show the deficit and how many more contracts at the minimum rate ($200/hr) would close the gap.

---

## Priority

This command exists because `/wealth` is for the $1.687B pipeline.
`/income` is for the next $500.
Use `/income` until monthly income exceeds $8,000.
Then use `/wealth`.

## Data Sources

| Source | Path | Contents |
|--------|------|----------|
| Mercenary DB | `~/.mercenary/mercenary.db` | Pipeline, approaches, contracts, follow-ups |
| Finance DB | `~/.mercenary/finance.db` | Income/expense transactions |
| Finance Tracker | `~/.mercenary/finance-tracker.sh` | CLI for finance.db |
| Revenue Tracker | `~/.mercenary/revenue-tracker.sh` | Cross-stream revenue view |
| Financial Config | `mercenary.db:financial_config` | Burn rate, reserves, min rate |
| Scanned Contracts | `mercenary.db:scanned_contracts` | 2000+ scraped opportunities |
| Approach Log | `mercenary.db:approach_log` | All outreach sent |
| Follow-up Schedule | `mercenary.db:follow_up_schedule` | Pending/overdue follow-ups |
| Ready Proposals | `~/protocols/revenue/wealth/execution/ready-proposals/` | Pre-written proposals |
| Cold Outreach | `~/protocols/revenue/wealth/execution/cold-outreach-consulting/` | Company-specific approaches |
| Upwork Profiles | `~/protocols/revenue/wealth/execution/upwork-profiles.md` | 3 niche profiles |
| Rate Ladder | `~/.mercenary/rate_ladder.json` | Current billing tier |
