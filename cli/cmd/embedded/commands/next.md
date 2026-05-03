Show the current work queue status (read-only).

1. Show any in-progress items:
   ```
   sqlite3 -header -column /Volumes/Lexar/work/work.db "SELECT id, title, priority, started_at FROM work_items WHERE status = 'in_progress' ORDER BY started_at"
   ```

2. Show the next pending item:
   ```
   sqlite3 -header -column /Volumes/Lexar/work/work.db "SELECT id, title, description, outcome, priority FROM work_items WHERE status = 'pending' ORDER BY priority ASC, created_at ASC LIMIT 1"
   ```

3. Show any blocked items:
   ```
   sqlite3 -header -column /Volumes/Lexar/work/work.db "SELECT id, title, reason FROM work_items WHERE status = 'blocked'"
   ```

4. Show queue summary:
   ```
   sqlite3 /Volumes/Lexar/work/work.db "SELECT status, COUNT(*) as count FROM work_items GROUP BY status ORDER BY CASE status WHEN 'in_progress' THEN 0 WHEN 'pending' THEN 1 WHEN 'blocked' THEN 2 WHEN 'done' THEN 3 END"
   ```

Present this as a concise status report. Do not modify any data.
