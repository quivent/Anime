-- ANIME Desktop Todo System Database Schema
-- SQLite database for comprehensive task management
-- Supports both human and agent assignees with rich metadata

-- Categories table for organizing todos
CREATE TABLE IF NOT EXISTS categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    color TEXT NOT NULL DEFAULT '#3B82F6', -- Hex color code
    icon TEXT, -- Icon identifier (emoji or icon name)
    description TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Tags table for flexible labeling
CREATE TABLE IF NOT EXISTS tags (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    color TEXT DEFAULT '#6B7280',
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Main todos table
CREATE TABLE IF NOT EXISTS todos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    description TEXT,
    status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending', 'in_progress', 'completed', 'blocked')),
    priority TEXT NOT NULL DEFAULT 'medium' CHECK(priority IN ('low', 'medium', 'high', 'critical')),
    assignee TEXT NOT NULL DEFAULT 'human' CHECK(assignee IN ('human', 'agent')),
    category_id INTEGER,
    parent_id INTEGER, -- For subtasks/hierarchical organization
    order_index INTEGER DEFAULT 0, -- For custom ordering within a parent/category
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    completed_at TEXT, -- Timestamp when status changed to completed
    due_date TEXT, -- ISO 8601 datetime
    estimated_hours REAL, -- Estimated effort in hours
    actual_hours REAL, -- Actual time spent
    metadata TEXT, -- JSON blob for additional flexible data
    FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL,
    FOREIGN KEY (parent_id) REFERENCES todos(id) ON DELETE CASCADE
);

-- Junction table for many-to-many relationship between todos and tags
CREATE TABLE IF NOT EXISTS todo_tags (
    todo_id INTEGER NOT NULL,
    tag_id INTEGER NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    PRIMARY KEY (todo_id, tag_id),
    FOREIGN KEY (todo_id) REFERENCES todos(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);

-- Activity log for tracking changes to todos
CREATE TABLE IF NOT EXISTS todo_activity (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    todo_id INTEGER NOT NULL,
    action TEXT NOT NULL, -- created, updated, status_changed, assigned, commented
    old_value TEXT, -- JSON snapshot of previous state
    new_value TEXT, -- JSON snapshot of new state
    actor TEXT NOT NULL DEFAULT 'user', -- user, agent, system
    comment TEXT, -- Optional comment/note
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (todo_id) REFERENCES todos(id) ON DELETE CASCADE
);

-- Indexes for performance optimization

-- Todos indexes
CREATE INDEX IF NOT EXISTS idx_todos_status ON todos(status);
CREATE INDEX IF NOT EXISTS idx_todos_priority ON todos(priority);
CREATE INDEX IF NOT EXISTS idx_todos_assignee ON todos(assignee);
CREATE INDEX IF NOT EXISTS idx_todos_category_id ON todos(category_id);
CREATE INDEX IF NOT EXISTS idx_todos_parent_id ON todos(parent_id);
CREATE INDEX IF NOT EXISTS idx_todos_due_date ON todos(due_date);
CREATE INDEX IF NOT EXISTS idx_todos_created_at ON todos(created_at);
CREATE INDEX IF NOT EXISTS idx_todos_updated_at ON todos(updated_at);
CREATE INDEX IF NOT EXISTS idx_todos_status_priority ON todos(status, priority);
CREATE INDEX IF NOT EXISTS idx_todos_assignee_status ON todos(assignee, status);

-- Todo tags indexes
CREATE INDEX IF NOT EXISTS idx_todo_tags_todo_id ON todo_tags(todo_id);
CREATE INDEX IF NOT EXISTS idx_todo_tags_tag_id ON todo_tags(tag_id);

-- Activity indexes
CREATE INDEX IF NOT EXISTS idx_todo_activity_todo_id ON todo_activity(todo_id);
CREATE INDEX IF NOT EXISTS idx_todo_activity_created_at ON todo_activity(created_at);
CREATE INDEX IF NOT EXISTS idx_todo_activity_action ON todo_activity(action);

-- Categories index
CREATE INDEX IF NOT EXISTS idx_categories_name ON categories(name);

-- Tags index
CREATE INDEX IF NOT EXISTS idx_tags_name ON tags(name);

-- Triggers for automatic updated_at timestamp
CREATE TRIGGER IF NOT EXISTS update_todos_updated_at
    AFTER UPDATE ON todos
    FOR EACH ROW
BEGIN
    UPDATE todos SET updated_at = datetime('now') WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS update_categories_updated_at
    AFTER UPDATE ON categories
    FOR EACH ROW
BEGIN
    UPDATE categories SET updated_at = datetime('now') WHERE id = NEW.id;
END;

-- Trigger to automatically set completed_at when status changes to completed
CREATE TRIGGER IF NOT EXISTS set_todo_completed_at
    AFTER UPDATE OF status ON todos
    FOR EACH ROW
    WHEN NEW.status = 'completed' AND OLD.status != 'completed'
BEGIN
    UPDATE todos SET completed_at = datetime('now') WHERE id = NEW.id;
END;

-- Trigger to log status changes to activity log
CREATE TRIGGER IF NOT EXISTS log_todo_status_change
    AFTER UPDATE OF status ON todos
    FOR EACH ROW
    WHEN NEW.status != OLD.status
BEGIN
    INSERT INTO todo_activity (todo_id, action, old_value, new_value)
    VALUES (NEW.id, 'status_changed', OLD.status, NEW.status);
END;

-- Default categories
INSERT OR IGNORE INTO categories (name, color, icon, description) VALUES
    ('Development', '#3B82F6', '💻', 'Software development tasks'),
    ('Design', '#8B5CF6', '🎨', 'Design and UI/UX tasks'),
    ('Documentation', '#10B981', '📝', 'Documentation and writing'),
    ('Testing', '#F59E0B', '🧪', 'Testing and QA tasks'),
    ('Deployment', '#EF4444', '🚀', 'Deployment and DevOps'),
    ('Research', '#6366F1', '🔬', 'Research and exploration'),
    ('Bug', '#DC2626', '🐛', 'Bug fixes and issues'),
    ('Feature', '#059669', '✨', 'New features and enhancements'),
    ('General', '#6B7280', '📋', 'General tasks');

-- Default tags
INSERT OR IGNORE INTO tags (name, color) VALUES
    ('urgent', '#EF4444'),
    ('backend', '#3B82F6'),
    ('frontend', '#8B5CF6'),
    ('database', '#10B981'),
    ('api', '#F59E0B'),
    ('ui', '#EC4899'),
    ('performance', '#F97316'),
    ('security', '#DC2626'),
    ('refactor', '#6366F1'),
    ('enhancement', '#14B8A6');

-- Views for common queries

-- View for todos with all related information
CREATE VIEW IF NOT EXISTS v_todos_full AS
SELECT
    t.id,
    t.title,
    t.description,
    t.status,
    t.priority,
    t.assignee,
    t.parent_id,
    t.order_index,
    t.created_at,
    t.updated_at,
    t.completed_at,
    t.due_date,
    t.estimated_hours,
    t.actual_hours,
    t.metadata,
    c.name as category_name,
    c.color as category_color,
    c.icon as category_icon,
    GROUP_CONCAT(tag.name, ',') as tags
FROM todos t
LEFT JOIN categories c ON t.category_id = c.id
LEFT JOIN todo_tags tt ON t.id = tt.todo_id
LEFT JOIN tags tag ON tt.tag_id = tag.id
GROUP BY t.id;

-- View for pending todos
CREATE VIEW IF NOT EXISTS v_todos_pending AS
SELECT * FROM v_todos_full
WHERE status = 'pending'
ORDER BY priority DESC, created_at ASC;

-- View for in-progress todos
CREATE VIEW IF NOT EXISTS v_todos_in_progress AS
SELECT * FROM v_todos_full
WHERE status = 'in_progress'
ORDER BY priority DESC, updated_at DESC;

-- View for completed todos
CREATE VIEW IF NOT EXISTS v_todos_completed AS
SELECT * FROM v_todos_full
WHERE status = 'completed'
ORDER BY completed_at DESC;

-- View for overdue todos
CREATE VIEW IF NOT EXISTS v_todos_overdue AS
SELECT * FROM v_todos_full
WHERE status != 'completed'
  AND due_date IS NOT NULL
  AND due_date < datetime('now')
ORDER BY due_date ASC;

-- View for todos by assignee
CREATE VIEW IF NOT EXISTS v_todos_by_assignee AS
SELECT
    assignee,
    status,
    COUNT(*) as count,
    SUM(CASE WHEN priority = 'critical' THEN 1 ELSE 0 END) as critical_count,
    SUM(CASE WHEN priority = 'high' THEN 1 ELSE 0 END) as high_count,
    SUM(CASE WHEN priority = 'medium' THEN 1 ELSE 0 END) as medium_count,
    SUM(CASE WHEN priority = 'low' THEN 1 ELSE 0 END) as low_count
FROM todos
GROUP BY assignee, status;
