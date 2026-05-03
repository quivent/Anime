Review the work queue ordering and suggest reprioritization.

1. Load all non-done work items:
   ```
   sqlite3 -header -column /Volumes/Lexar/work/work.db "
     SELECT id, title, description, outcome, reason, status, priority, created_at
     FROM work_items
     WHERE status != 'done'
     ORDER BY priority ASC, created_at ASC
   "
   ```

2. Analyze the items for:
   - Priority ordering: are the most impactful items first?
   - Dependencies: should any items be blocked by others?
   - Redundancy: are any items duplicates or subsets of others?
   - Missing items: based on the title/description patterns, are there obvious gaps?

3. Present a suggested reprioritization with reasoning. Show:
   - Current order vs. suggested order
   - Which items should be reprioritized and why
   - Any items that should be marked as blocked

4. If the user confirms, apply the changes:
   ```
   sqlite3 /Volumes/Lexar/work/work.db "UPDATE work_items SET priority = <new_priority> WHERE id = <ID>"
   ```

Only modify priorities if the user explicitly confirms. Present suggestions first.
