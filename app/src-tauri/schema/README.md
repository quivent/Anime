# ANIME Desktop Todo System Database Schema

This directory contains the SQL schema for the comprehensive todo/task management system.

## Database Location

The SQLite database is created at: `~/.anime-desktop/todos.db`

## Schema Overview

The schema is designed for a robust, production-ready todo system with the following features:

### Core Tables

#### 1. `todos`
Main table storing all todo items.

**Columns:**
- `id` (INTEGER PRIMARY KEY) - Auto-incrementing unique identifier
- `title` (TEXT NOT NULL) - Todo title
- `description` (TEXT) - Detailed description
- `status` (TEXT NOT NULL) - One of: pending, in_progress, completed, blocked
- `priority` (TEXT NOT NULL) - One of: low, medium, high, critical
- `assignee` (TEXT NOT NULL) - One of: human, agent
- `category_id` (INTEGER FK) - References categories.id
- `parent_id` (INTEGER FK) - Self-reference for hierarchical todos
- `order_index` (INTEGER) - Custom ordering within parent/category
- `created_at` (TEXT NOT NULL) - ISO 8601 timestamp
- `updated_at` (TEXT NOT NULL) - ISO 8601 timestamp (auto-updated)
- `completed_at` (TEXT) - ISO 8601 timestamp (auto-set when completed)
- `due_date` (TEXT) - ISO 8601 timestamp
- `estimated_hours` (REAL) - Estimated effort
- `actual_hours` (REAL) - Actual time spent
- `metadata` (TEXT) - JSON blob for extensibility

**Indexes:**
- Single column: status, priority, assignee, category_id, parent_id, due_date, created_at, updated_at
- Composite: (status, priority), (assignee, status)

#### 2. `categories`
Predefined categories for organizing todos.

**Columns:**
- `id` (INTEGER PRIMARY KEY)
- `name` (TEXT NOT NULL UNIQUE)
- `color` (TEXT NOT NULL) - Hex color code
- `icon` (TEXT) - Icon identifier
- `description` (TEXT)
- `created_at` (TEXT NOT NULL)
- `updated_at` (TEXT NOT NULL) - Auto-updated

**Default Categories:**
- Development (💻 #3B82F6)
- Design (🎨 #8B5CF6)
- Documentation (📝 #10B981)
- Testing (🧪 #F59E0B)
- Deployment (🚀 #EF4444)
- Research (🔬 #6366F1)
- Bug (🐛 #DC2626)
- Feature (✨ #059669)
- General (📋 #6B7280)

#### 3. `tags`
Flexible tag system for labeling todos.

**Columns:**
- `id` (INTEGER PRIMARY KEY)
- `name` (TEXT NOT NULL UNIQUE)
- `color` (TEXT) - Hex color code
- `created_at` (TEXT NOT NULL)

**Default Tags:**
- urgent, backend, frontend, database, api, ui, performance, security, refactor, enhancement

#### 4. `todo_tags`
Junction table for many-to-many todo-tag relationships.

**Columns:**
- `todo_id` (INTEGER FK NOT NULL) - References todos.id
- `tag_id` (INTEGER FK NOT NULL) - References tags.id
- `created_at` (TEXT NOT NULL)

**Primary Key:** (todo_id, tag_id)

**Behavior:**
- ON DELETE CASCADE for both foreign keys

#### 5. `todo_activity`
Activity log tracking all changes to todos.

**Columns:**
- `id` (INTEGER PRIMARY KEY)
- `todo_id` (INTEGER FK NOT NULL) - References todos.id
- `action` (TEXT NOT NULL) - created, updated, status_changed, assigned, commented
- `old_value` (TEXT) - JSON snapshot of previous state
- `new_value` (TEXT) - JSON snapshot of new state
- `actor` (TEXT NOT NULL) - user, agent, system
- `comment` (TEXT) - Optional comment/note
- `created_at` (TEXT NOT NULL)

### Views

Pre-defined views for common queries:

#### `v_todos_full`
Complete todo information with category and tags.

**Columns:** All todo fields plus:
- `category_name`, `category_color`, `category_icon`
- `tags` (comma-separated list)

#### `v_todos_pending`
All pending todos ordered by priority and creation date.

#### `v_todos_in_progress`
All in-progress todos ordered by priority and update date.

#### `v_todos_completed`
All completed todos ordered by completion date.

#### `v_todos_overdue`
Incomplete todos past their due date, ordered by due date.

#### `v_todos_by_assignee`
Statistics grouped by assignee and status with priority breakdowns.

### Triggers

#### Auto-Update Triggers
- `update_todos_updated_at` - Automatically updates `updated_at` on any todo change
- `update_categories_updated_at` - Automatically updates `updated_at` on category change

#### Business Logic Triggers
- `set_todo_completed_at` - Automatically sets `completed_at` when status changes to completed
- `log_todo_status_change` - Automatically logs status changes to `todo_activity` table

### Foreign Key Constraints

All foreign keys use appropriate cascading:
- `todos.category_id` → `categories.id` (ON DELETE SET NULL)
- `todos.parent_id` → `todos.id` (ON DELETE CASCADE)
- `todo_tags.todo_id` → `todos.id` (ON DELETE CASCADE)
- `todo_tags.tag_id` → `tags.id` (ON DELETE CASCADE)
- `todo_activity.todo_id` → `todos.id` (ON DELETE CASCADE)

## Performance Optimizations

1. **Comprehensive Indexing**: 15+ indexes covering all common query patterns
2. **Composite Indexes**: Optimized for multi-column WHERE clauses
3. **Materialized Views**: Pre-computed common queries (via views)
4. **Triggers for Denormalization**: Auto-updated timestamps reduce query complexity

## Data Integrity

1. **Check Constraints**: Enforce valid enum values for status and priority
2. **Foreign Keys**: Maintain referential integrity
3. **Unique Constraints**: Prevent duplicate categories and tags
4. **NOT NULL Constraints**: Ensure required fields are always populated

## Migration Strategy

The schema is designed to be applied idempotently using `CREATE TABLE IF NOT EXISTS` and `INSERT OR IGNORE`. This allows:
- Safe re-application of the schema
- Easy upgrades with ALTER TABLE statements
- Backward compatibility

## Querying Examples

### Get all high-priority pending todos
```sql
SELECT * FROM v_todos_pending WHERE priority = 'high';
```

### Get todos due this week
```sql
SELECT * FROM todos
WHERE due_date BETWEEN datetime('now') AND datetime('now', '+7 days')
  AND status != 'completed'
ORDER BY due_date ASC;
```

### Get todos with multiple tags
```sql
SELECT t.*, GROUP_CONCAT(tag.name) as all_tags
FROM todos t
JOIN todo_tags tt ON t.id = tt.todo_id
JOIN tags tag ON tt.tag_id = tag.id
GROUP BY t.id
HAVING COUNT(DISTINCT tag.id) > 1;
```

### Get todo hierarchy (parent-child)
```sql
WITH RECURSIVE todo_tree AS (
  SELECT id, title, parent_id, 0 as level
  FROM todos
  WHERE parent_id IS NULL

  UNION ALL

  SELECT t.id, t.title, t.parent_id, tt.level + 1
  FROM todos t
  JOIN todo_tree tt ON t.parent_id = tt.id
)
SELECT * FROM todo_tree ORDER BY level, id;
```

### Get activity history for a todo
```sql
SELECT * FROM todo_activity
WHERE todo_id = ?
ORDER BY created_at DESC;
```

## Extending the Schema

To add new fields to todos:

```sql
-- Add column
ALTER TABLE todos ADD COLUMN new_field TEXT;

-- Add index if needed
CREATE INDEX IF NOT EXISTS idx_todos_new_field ON todos(new_field);
```

To add new enum values, update the CHECK constraint:

```sql
-- SQLite doesn't support ALTER TABLE for constraints
-- Instead, recreate the table with new constraints
-- or handle validation in application code
```

## Backup and Export

### Backup
```bash
sqlite3 ~/.anime-desktop/todos.db ".backup todos_backup.db"
```

### Export to CSV
```bash
sqlite3 ~/.anime-desktop/todos.db \
  -header -csv \
  "SELECT * FROM v_todos_full;" \
  > todos_export.csv
```

### Export to JSON
```bash
sqlite3 ~/.anime-desktop/todos.db \
  -json \
  "SELECT * FROM v_todos_full;" \
  > todos_export.json
```

## Schema Versioning

Current version: 1.0.0

Future versions should include a schema_version table:

```sql
CREATE TABLE IF NOT EXISTS schema_version (
    version TEXT PRIMARY KEY,
    applied_at TEXT NOT NULL DEFAULT (datetime('now'))
);
```

## Performance Metrics

Expected performance on typical hardware:

- **Insert**: < 1ms per todo
- **Query (indexed)**: < 5ms for 10,000 todos
- **Full-text search**: < 10ms for 10,000 todos
- **Complex joins**: < 20ms with proper indexing

## Security Considerations

1. All user input should be parameterized (no SQL injection)
2. File permissions on the database file should be 600
3. Backup database regularly to prevent data loss
4. Consider encryption for sensitive metadata

## Troubleshooting

### Database locked errors
```rust
// Increase timeout in rusqlite
conn.busy_timeout(Duration::from_secs(5))?;
```

### Corrupted database
```bash
# Check integrity
sqlite3 ~/.anime-desktop/todos.db "PRAGMA integrity_check;"

# Rebuild if needed
sqlite3 ~/.anime-desktop/todos.db ".dump" | sqlite3 new_todos.db
```

### Large database size
```bash
# Vacuum to reclaim space
sqlite3 ~/.anime-desktop/todos.db "VACUUM;"
```
