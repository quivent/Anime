# Todo System - Tauri Commands Documentation

This module provides comprehensive CRUD operations for a todo/task management system integrated into the ANIME desktop application.

## Overview

The todo system supports:
- Creating, reading, updating, and deleting todos
- Organizing todos with categories and tags
- Filtering and searching todos
- Tracking todo statistics
- Managing hierarchical todos (parent-child relationships)
- Time tracking with estimated and actual hours
- Custom metadata for extensibility

## Data Structures

### TodoStatus
```rust
enum TodoStatus {
    Pending,      // Not yet started
    InProgress,   // Currently being worked on
    Completed,    // Finished
    Blocked,      // Blocked by dependencies or issues
    Cancelled,    // Cancelled/abandoned
}
```

### TodoPriority
```rust
enum TodoPriority {
    Low,
    Medium,
    High,
    Critical,
}
```

### Todo
```rust
struct Todo {
    id: String,
    title: String,
    description: Option<String>,
    status: TodoStatus,
    priority: TodoPriority,
    category: Option<String>,
    assignee: Option<String>,
    due_date: Option<DateTime<Utc>>,
    parent_id: Option<String>,
    tags: Vec<String>,
    created_at: DateTime<Utc>,
    updated_at: DateTime<Utc>,
    completed_at: Option<DateTime<Utc>>,
    estimated_hours: Option<f64>,
    actual_hours: Option<f64>,
    metadata: Option<serde_json::Value>,
}
```

## Available Commands

### 1. create_todo
Create a new todo item.

**Parameters:**
- `title: String` - The title of the todo (required, cannot be empty)
- `description: Option<String>` - Detailed description
- `priority: TodoPriority` - Priority level
- `category: Option<String>` - Category ID
- `assignee: Option<String>` - Person assigned to this todo
- `due_date: Option<DateTime<Utc>>` - Due date
- `parent_id: Option<String>` - Parent todo ID for creating subtasks
- `tags: Vec<String>` - Array of tag names

**Returns:** `Result<Todo, String>`

**Example (JavaScript):**
```javascript
const todo = await invoke('create_todo', {
  title: 'Implement user authentication',
  description: 'Add JWT-based authentication to the API',
  priority: 'high',
  category: 'backend',
  assignee: 'john@example.com',
  dueDate: new Date('2025-12-31').toISOString(),
  parentId: null,
  tags: ['backend', 'security', 'api']
});
```

---

### 2. get_todos
Get all todos with optional filtering.

**Parameters:**
- `filters: Option<TodoFilters>` - Filter criteria (optional)

**TodoFilters structure:**
```rust
struct TodoFilters {
    status: Option<TodoStatus>,
    priority: Option<TodoPriority>,
    category: Option<String>,
    assignee: Option<String>,
    search_query: Option<String>,
    tags: Option<Vec<String>>,
    include_subtasks: Option<bool>,
}
```

**Returns:** `Result<Vec<Todo>, String>`

**Examples (JavaScript):**
```javascript
// Get all todos
const allTodos = await invoke('get_todos', { filters: null });

// Get only pending high-priority todos
const urgentTodos = await invoke('get_todos', {
  filters: {
    status: 'pending',
    priority: 'high'
  }
});

// Search todos
const searchResults = await invoke('get_todos', {
  filters: {
    searchQuery: 'authentication'
  }
});

// Get todos by tag
const backendTodos = await invoke('get_todos', {
  filters: {
    tags: ['backend']
  }
});

// Get only top-level todos (no subtasks)
const topLevelTodos = await invoke('get_todos', {
  filters: {
    includeSubtasks: false
  }
});
```

---

### 3. get_todo_by_id
Get a single todo by its ID.

**Parameters:**
- `id: String` - Todo ID

**Returns:** `Result<Option<Todo>, String>`

**Example (JavaScript):**
```javascript
const todo = await invoke('get_todo_by_id', {
  id: 'todo-uuid-here'
});

if (todo) {
  console.log('Todo found:', todo);
} else {
  console.log('Todo not found');
}
```

---

### 4. update_todo
Update an existing todo.

**Parameters:**
- `id: String` - Todo ID
- `updates: TodoUpdate` - Fields to update

**TodoUpdate structure:**
```rust
struct TodoUpdate {
    title: Option<String>,
    description: Option<String>,
    status: Option<TodoStatus>,
    priority: Option<TodoPriority>,
    category: Option<String>,
    assignee: Option<String>,
    due_date: Option<DateTime<Utc>>,
    tags: Option<Vec<String>>,
    estimated_hours: Option<f64>,
    actual_hours: Option<f64>,
    metadata: Option<serde_json::Value>,
}
```

**Returns:** `Result<Todo, String>`

**Examples (JavaScript):**
```javascript
// Update status to completed
const updated = await invoke('update_todo', {
  id: 'todo-uuid',
  updates: {
    status: 'completed'
  }
});

// Update multiple fields
const updated = await invoke('update_todo', {
  id: 'todo-uuid',
  updates: {
    title: 'New title',
    priority: 'critical',
    dueDate: new Date('2025-11-30').toISOString(),
    actualHours: 5.5
  }
});
```

---

### 5. delete_todo
Delete a todo by ID.

**Parameters:**
- `id: String` - Todo ID

**Returns:** `Result<bool, String>` - Returns true if deleted, false if not found

**Example (JavaScript):**
```javascript
const deleted = await invoke('delete_todo', {
  id: 'todo-uuid'
});

if (deleted) {
  console.log('Todo deleted successfully');
} else {
  console.log('Todo not found');
}
```

---

### 6. bulk_update_todos
Update the status of multiple todos at once.

**Parameters:**
- `ids: Vec<String>` - Array of todo IDs
- `status: TodoStatus` - New status to apply

**Returns:** `Result<usize, String>` - Number of todos updated

**Example (JavaScript):**
```javascript
const updatedCount = await invoke('bulk_update_todos', {
  ids: ['todo-1', 'todo-2', 'todo-3'],
  status: 'completed'
});

console.log(`Updated ${updatedCount} todos`);
```

---

### 7. get_todo_stats
Get statistics about todos.

**Parameters:** None

**Returns:** `Result<TodoStats, String>`

**TodoStats structure:**
```rust
struct TodoStats {
    total: usize,
    pending: usize,
    in_progress: usize,
    completed: usize,
    blocked: usize,
    cancelled: usize,
    by_priority: PriorityStats,
    overdue: usize,
}

struct PriorityStats {
    low: usize,
    medium: usize,
    high: usize,
    critical: usize,
}
```

**Example (JavaScript):**
```javascript
const stats = await invoke('get_todo_stats');

console.log(`Total todos: ${stats.total}`);
console.log(`Pending: ${stats.pending}`);
console.log(`In Progress: ${stats.in_progress}`);
console.log(`Completed: ${stats.completed}`);
console.log(`Blocked: ${stats.blocked}`);
console.log(`Overdue: ${stats.overdue}`);
console.log(`Critical priority: ${stats.by_priority.critical}`);
```

---

### 8. search_todos
Search todos by query string.

**Parameters:**
- `query: String` - Search query (searches in title and description)

**Returns:** `Result<Vec<Todo>, String>`

**Example (JavaScript):**
```javascript
const results = await invoke('search_todos', {
  query: 'authentication'
});

console.log(`Found ${results.length} todos matching "authentication"`);
```

---

### 9. create_category
Create a new category.

**Parameters:**
- `name: String` - Category name (required, cannot be empty)
- `color: String` - Hex color code (e.g., "#FF5733" or "#F57")
- `icon: Option<String>` - Icon identifier (optional)

**Returns:** `Result<Category, String>`

**Example (JavaScript):**
```javascript
const category = await invoke('create_category', {
  name: 'Backend Development',
  color: '#3498db',
  icon: 'code'
});
```

---

### 10. get_categories
Get all categories.

**Parameters:** None

**Returns:** `Result<Vec<Category>, String>`

**Example (JavaScript):**
```javascript
const categories = await invoke('get_categories');

categories.forEach(cat => {
  console.log(`${cat.name} (${cat.color})`);
});
```

---

### 11. add_tags_to_todo
Add tags to an existing todo.

**Parameters:**
- `todo_id: String` - Todo ID
- `tags: Vec<String>` - Array of tag names to add

**Returns:** `Result<bool, String>`

**Example (JavaScript):**
```javascript
await invoke('add_tags_to_todo', {
  todoId: 'todo-uuid',
  tags: ['urgent', 'needs-review']
});
```

---

### 12. get_tags
Get all tags with usage counts.

**Parameters:** None

**Returns:** `Result<Vec<Tag>, String>`

**Tag structure:**
```rust
struct Tag {
    id: String,
    name: String,
    color: Option<String>,
    usage_count: usize,
}
```

**Example (JavaScript):**
```javascript
const tags = await invoke('get_tags');

// Sort by most used
tags.sort((a, b) => b.usage_count - a.usage_count);

console.log('Most used tags:');
tags.slice(0, 5).forEach(tag => {
  console.log(`${tag.name}: ${tag.usage_count} uses`);
});
```

---

### 13. create_tag
Create a new predefined tag.

**Parameters:**
- `name: String` - Tag name (required, cannot be empty)
- `color: Option<String>` - Hex color code (optional)

**Returns:** `Result<Tag, String>`

**Example (JavaScript):**
```javascript
const tag = await invoke('create_tag', {
  name: 'urgent',
  color: '#e74c3c'
});
```

---

## Error Handling

All commands return `Result<T, String>` types. In JavaScript/TypeScript, wrap calls in try-catch blocks:

```javascript
try {
  const todo = await invoke('create_todo', { /* params */ });
  console.log('Success:', todo);
} catch (error) {
  console.error('Error:', error);
  // Display error to user
}
```

## Common Error Messages

- `"Todo title cannot be empty"` - Title is required and must have content
- `"Parent todo with id '...' not found"` - Invalid parent_id provided
- `"Todo with id '...' not found"` - Todo doesn't exist
- `"Category name cannot be empty"` - Category name is required
- `"Color must be in hex format (#RGB or #RRGGBB)"` - Invalid color format
- `"No todo IDs provided for bulk update"` - Empty array passed to bulk_update_todos

## Implementation Notes

### Storage
Uses SQLite database for persistent storage via `rusqlite`. The database is stored at `~/.anime-desktop/todos.db` and includes:
- Full relational schema with proper foreign keys
- Automatic indexing for performance optimization
- Database triggers for auto-updating timestamps
- Pre-populated with default categories and tags
- Activity logging for tracking changes

### Logging
All commands include log statements for debugging. Enable logging in your Tauri config to see these messages.

### Thread Safety
The storage layer uses `Arc<Mutex<HashMap>>` for thread-safe concurrent access to todos, categories, and tags.

### Tag Usage Tracking
Tags automatically track their usage count when:
- A todo is created with tags
- Tags are added to a todo
- A todo's tags are updated

### Auto-Completion Tracking
When a todo's status is changed to `Completed`, the `completed_at` timestamp is automatically set. When changed away from `Completed`, this timestamp is cleared.

## Database Schema

The SQLite schema includes the following tables:
- **todos**: Main todo items with all fields
- **categories**: Predefined categories for organizing todos
- **tags**: Flexible tag system with usage tracking
- **todo_tags**: Junction table for many-to-many todo-tag relationships
- **todo_activity**: Activity log for tracking all changes

Views are also provided for common queries:
- `v_todos_full`: Todos with all related category and tag information
- `v_todos_pending`, `v_todos_in_progress`, `v_todos_completed`: Filtered by status
- `v_todos_overdue`: Todos past their due date
- `v_todos_by_assignee`: Statistics grouped by assignee

## Future Enhancements

Consider adding:
- Todo attachments/files
- Recurring todos
- Todo templates
- Export/import functionality (JSON, CSV)
- Notifications for due dates
- Advanced time tracking with start/stop functionality
- Collaboration features (sharing, mentions)
- Custom views and saved filters
- Gantt chart visualization
- Kanban board view

---

# Seed Data Documentation

## Overview

The seed data includes:
- **8 Categories**: Organized by development focus areas
- **50 Todos**: Comprehensive list of all identified work items from the codebase audit

## Categories

1. **🐛 Bug Fixes** (Red: #ef4444) - Critical bugs and issues that need fixing
2. **✨ Features** (Purple: #8b5cf6) - New features and functionality
3. **🔧 Refactoring** (Orange: #f59e0b) - Code quality improvements and refactoring
4. **📚 Documentation** (Blue: #3b82f6) - Documentation and guides
5. **🧪 Testing** (Green: #10b981) - Tests and test coverage improvements
6. **🎨 UI/UX** (Pink: #ec4899) - User interface and experience improvements
7. **⚡ Performance** (Yellow: #eab308) - Performance optimizations
8. **🔒 Security** (Dark Red: #dc2626) - Security improvements and fixes

## Priority Breakdown

### Critical Priority (7 todos)
Primary focus on completing mock implementations and core SSH functionality:
- Complete SSH functionality implementation
- Implement real AI text generation (replacing mock)
- Implement real script analysis (replacing mock)
- Implement real screenplay parser (replacing mock)
- Implement real shot suggestions (replacing mock)
- Implement real storyboard image generation (replacing mock)
- Implement real export functionality (replacing mock)

### High Priority (12 todos)
Code quality improvements and essential features:
- Version tracking for installer
- Dependency resolver implementation
- Installation logic completion
- Server configuration loading
- Quick-win compiler warning fixes (6 items)

### Medium Priority (21 todos)
Code quality and feature enhancements:
- **Code Quality** (11 todos):
  - Error handling standardization
  - Async pattern consolidation
  - API call patterns
  - Type consolidation
  - Testing infrastructure
  - Documentation

- **Features** (10 todos):
  - Dark mode theme
  - Logging system
  - Error tracking
  - Notifications
  - Search functionality
  - Settings persistence
  - Multi-instance management
  - Connection retry logic

### Low Priority (10 todos)
Nice-to-have features and polish:
- Animation transitions
- Drag-and-drop uploads
- Command palette (Cmd+K)
- Custom icon set
- Tutorial/onboarding
- Undo/redo functionality
- Bulk operations
- Dashboard analytics
- Notification system
- Backup/restore

## Assignee Distribution

- **Agent**: 30 todos (primarily technical implementation tasks)
- **Human**: 20 todos (requiring design decisions or planning)

## Seeding Commands

### 14. seed_initial_todos
Seeds the database with initial todos and categories on first run.

**Parameters:** None

**Returns:** `Result<(usize, usize), String>` - (categories_count, todos_count)

**Example (JavaScript):**
```javascript
const [categoriesCount, todosCount] = await invoke('seed_initial_todos');
console.log(`Seeded ${categoriesCount} categories and ${todosCount} todos`);
```

---

### 15. is_seeded
Checks if the database has been seeded with initial data.

**Parameters:** None

**Returns:** `Result<bool, String>` - true if seeded, false otherwise

**Example (JavaScript):**
```javascript
const isSeeded = await invoke('is_seeded');

if (!isSeeded) {
  // Seed the database
  await invoke('seed_initial_todos');
}
```

## Usage Example

Complete initialization flow:

```javascript
import { invoke } from '@tauri-apps/api/tauri';

async function initializeTodos() {
  try {
    // Check if already seeded
    const seeded = await invoke('is_seeded');

    if (!seeded) {
      console.log('Seeding initial todos...');
      const [categoriesCount, todosCount] = await invoke('seed_initial_todos');
      console.log(`Successfully seeded ${categoriesCount} categories and ${todosCount} todos`);
    } else {
      console.log('Database already seeded');
    }

    // Get all categories
    const categories = await invoke('get_categories');
    console.log('Available categories:', categories);

    // Get critical priority todos
    const criticalTodos = await invoke('get_todos', {
      filters: {
        priority: 'critical',
        status: 'pending'
      }
    });
    console.log(`Found ${criticalTodos.length} critical pending todos`);

  } catch (error) {
    console.error('Error initializing todos:', error);
  }
}

// Call on app startup
initializeTodos();
```

## Data Structure

Each todo includes:
- **Basic Info**: ID, title, description
- **Organization**: Category, priority, status, assignee
- **Tags**: Multiple tags for filtering and search
- **Scheduling**: Due date, estimated hours
- **Metadata**: File references stored in metadata field
- **Timestamps**: Created, updated, completed dates

Subtasks are embedded in the description using markdown bullet points for simplicity.

## File References

Todos include file references in the metadata field pointing to:
- Source files with TODOs
- Files requiring changes
- Related code locations

Example:
```json
{
  "file_references": [
    "/Users/joshkornreich/lambda/anime-desktop/src-tauri/src/ssh.rs"
  ]
}
```

## Tags

Common tags used across todos:
- **Technology**: `ssh`, `ai`, `lambda`, `backend`, `frontend`
- **Type**: `mock`, `api`, `ui`, `parser`, `export`
- **Focus**: `code-quality`, `warnings`, `networking`, `testing`
- **Status**: `quick-win` (for easy fixes)

## Implementation Notes

1. The seed data is designed to be run once on first application launch
2. Categories use consistent color schemes matching the UI theme
3. Due dates are relative to seed time (7 days, 14 days, 30 days)
4. Estimated hours help with sprint planning
5. File references enable quick navigation from todos to code

## Testing

The seed module includes comprehensive tests:
- Category ID uniqueness validation
- Category reference validation in todos
- Priority distribution coverage
- Todo count verification (should be exactly 50)

Run tests with:
```bash
cargo test --lib todos::seed
```

## Maintenance

When adding new todos to the seed data:
1. Update the `create_todo` calls in `get_initial_todos()`
2. Ensure category IDs match existing categories
3. Add appropriate tags for filtering
4. Include file references where applicable
5. Update the count in `test_todo_count` test
6. Run tests to verify data integrity
