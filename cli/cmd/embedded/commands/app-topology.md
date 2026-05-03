# /app-topology - Show What the App Agent Loaded

Read-only diagnostic. Shows exactly what `/app-agent` picked up from `projects.db` for the current project.

---

## Purpose

After running `/app-agent`, you may want to verify what was actually loaded — which constraints are active, what navigation was picked up, whether the exploration cache is fresh. This command answers that question with a single query and formatted output.

## Execution

**One Bash call. One database open.**

```bash
CWD=$(pwd)
HASH=$(git rev-parse HEAD 2>/dev/null || echo "no-git")
sqlite3 -separator '	' ~/.claude/db/projects.db << ENDSQL
.headers off

SELECT '=== PROJECT ===' as tag;
SELECT name, domain, sensitivity, purpose FROM projects WHERE path = '$CWD';

SELECT '=== CONSTRAINTS ===' as tag;
SELECT severity, type, content FROM constraints
  WHERE project_id = (SELECT id FROM projects WHERE path = '$CWD')
  ORDER BY CASE severity WHEN 'absolute' THEN 1 WHEN 'strong' THEN 2 ELSE 3 END;

SELECT '=== NAVIGATION ===' as tag;
SELECT importance, category, path, description FROM navigation
  WHERE project_id = (SELECT id FROM projects WHERE path = '$CWD')
  ORDER BY CASE importance WHEN 'critical' THEN 1 WHEN 'high' THEN 2 ELSE 3 END;

SELECT '=== COMMANDS ===' as tag;
SELECT name, command FROM commands
  WHERE project_id = (SELECT id FROM projects WHERE path = '$CWD');

SELECT '=== GLOSSARY ===' as tag;
SELECT term, definition FROM glossary
  WHERE project_id = (SELECT id FROM projects WHERE path = '$CWD');

SELECT '=== CONVENTIONS ===' as tag;
SELECT category, pattern FROM conventions
  WHERE project_id = (SELECT id FROM projects WHERE path = '$CWD');

SELECT '=== CACHE ===' as tag;
SELECT git_hash, total_loc, file_count,
       CASE WHEN git_hash = '$HASH' THEN 'FRESH' ELSE 'STALE' END as freshness
  FROM exploration_cache
  WHERE project_id = (SELECT id FROM projects WHERE path = '$CWD');

SELECT '=== COUNTS ===' as tag;
SELECT
  (SELECT count(*) FROM constraints WHERE project_id = p.id) as constraints,
  (SELECT count(*) FROM navigation WHERE project_id = p.id) as navigation,
  (SELECT count(*) FROM verification WHERE project_id = p.id) as verification,
  (SELECT count(*) FROM commands WHERE project_id = p.id) as commands,
  (SELECT count(*) FROM conventions WHERE project_id = p.id) as conventions,
  (SELECT count(*) FROM glossary WHERE project_id = p.id) as glossary,
  (SELECT count(*) FROM integrations WHERE project_id = p.id) as integrations,
  (SELECT count(*) FROM personas WHERE project_id = p.id) as personas
FROM projects p WHERE p.path = '$CWD';
ENDSQL
```

## Output Format

Parse the tagged output and render:

```
╭─────────────────────────────────────────────────────────────╮
│  APP-AGENT TOPOLOGY                                         │
│  What was loaded from projects.db                           │
╰─────────────────────────────────────────────────────────────╯

Project: [name] ([domain], [sensitivity])
Cache: [FRESH|STALE] — [total_loc] LOC, [file_count] files

ACTIVE CONSTRAINTS ([n] total)
  ABSOLUTE:
    [severity] [type]: [content]
    ...
  STRONG:
    [type]: [content]
    ...
  SUGGESTED:
    [type]: [content]
    ...

NAVIGATION ([n] locations)
  critical:
    [category] → [path] — [description]
  high:
    [category] → [path] — [description]
  normal:
    [category] → [path] — [description]

COMMANDS ([n])
  [name] → [command]
  ...

GLOSSARY ([n] terms)
  [term]: [definition]
  ...

CONVENTIONS ([n] patterns)
  [category]: [pattern]
  ...

Summary: [constraints] constraints, [navigation] nav, [commands] cmds,
         [glossary] terms, [conventions] patterns, [verification] checks,
         [integrations] integrations, [personas] personas
```

## If No Project Found

```
No project encoding found for: [cwd]
Run /project-encode first, then /app-agent.
```

## Notes

- This is read-only. It changes nothing.
- Use this to verify encoding quality after `/project-encode`.
- Use this to confirm what an agent session is operating with.
- If the cache shows STALE, consider re-running `/project-encode`.
