# Todo System Quick Start Guide

## Installation

Already installed! The todo system is integrated into the ANIME desktop application.

## Database Location

```
~/.anime-desktop/todos.db
```

## Quick Examples

### Creating a Todo

```javascript
import { invoke } from '@tauri-apps/api/tauri';

const todo = await invoke('create_todo', {
    title: 'My first todo',
    description: 'This is a test todo',
    priority: 'medium',  // low | medium | high | critical
    category: 'Development',
    assignee: 'human',  // human | agent
    dueDate: new Date('2025-12-31').toISOString(),
    parentId: null,  // For subtasks
    tags: ['backend', 'urgent']
});

console.log('Created todo:', todo.id);
```

### Getting All Todos

```javascript
// Get all todos
const allTodos = await invoke('get_todos', { filters: null });

// Get only pending todos
const pending = await invoke('get_todos', {
    filters: { status: 'pending' }
});

// Get high-priority todos
const urgent = await invoke('get_todos', {
    filters: { priority: 'high' }
});

// Get todos for a specific category
const devTodos = await invoke('get_todos', {
    filters: { category: 'Development' }
});
```

### Updating a Todo

```javascript
// Mark as completed
await invoke('update_todo', {
    id: 'todo-id-here',
    updates: { status: 'completed' }
});

// Change priority
await invoke('update_todo', {
    id: 'todo-id-here',
    updates: { priority: 'critical' }
});

// Add actual hours worked
await invoke('update_todo', {
    id: 'todo-id-here',
    updates: { actualHours: 3.5 }
});
```

### Searching Todos

```javascript
const results = await invoke('search_todos', {
    query: 'authentication'
});
```

### Getting Statistics

```javascript
const stats = await invoke('get_todo_stats');

console.log(`Total: ${stats.total}`);
console.log(`Pending: ${stats.pending}`);
console.log(`In Progress: ${stats.in_progress}`);
console.log(`Completed: ${stats.completed}`);
console.log(`Blocked: ${stats.blocked}`);
console.log(`Overdue: ${stats.overdue}`);
console.log(`Critical: ${stats.by_priority.critical}`);
```

### Working with Categories

```javascript
// Get all categories
const categories = await invoke('get_categories');

// Create a new category
const category = await invoke('create_category', {
    name: 'Marketing',
    color: '#FF6B6B',
    icon: '📢'
});
```

### Working with Tags

```javascript
// Get all tags
const tags = await invoke('get_tags');

// Create a new tag
const tag = await invoke('create_tag', {
    name: 'needs-review',
    color: '#FFA500'
});

// Add tags to a todo
await invoke('add_tags_to_todo', {
    todoId: 'todo-id-here',
    tags: ['urgent', 'needs-review']
});
```

### Bulk Operations

```javascript
// Mark multiple todos as completed
const updatedCount = await invoke('bulk_update_todos', {
    ids: ['todo-1', 'todo-2', 'todo-3'],
    status: 'completed'
});

console.log(`Updated ${updatedCount} todos`);
```

### Creating Subtasks

```javascript
// Create parent todo
const parent = await invoke('create_todo', {
    title: 'Build authentication system',
    priority: 'high',
    category: 'Development',
    assignee: 'human',
    tags: ['backend']
});

// Create subtasks
const subtask1 = await invoke('create_todo', {
    title: 'Design database schema',
    priority: 'high',
    category: 'Development',
    assignee: 'human',
    parentId: parent.id,  // Link to parent
    tags: ['database']
});

const subtask2 = await invoke('create_todo', {
    title: 'Implement JWT tokens',
    priority: 'high',
    category: 'Development',
    assignee: 'agent',
    parentId: parent.id,  // Link to parent
    tags: ['security']
});
```

## Status Values

- `pending` - Not yet started
- `in_progress` - Currently being worked on
- `completed` - Finished
- `blocked` - Blocked by dependencies or issues
- `cancelled` - Cancelled/abandoned

## Priority Values

- `low` - Low priority
- `medium` - Medium priority
- `high` - High priority
- `critical` - Critical/urgent

## Assignee Values

- `human` - Assigned to a human
- `agent` - Assigned to an AI agent

## Default Categories

The system comes pre-populated with these categories:

1. **Development** (💻 #3B82F6) - Software development tasks
2. **Design** (🎨 #8B5CF6) - Design and UI/UX tasks
3. **Documentation** (📝 #10B981) - Documentation and writing
4. **Testing** (🧪 #F59E0B) - Testing and QA tasks
5. **Deployment** (🚀 #EF4444) - Deployment and DevOps
6. **Research** (🔬 #6366F1) - Research and exploration
7. **Bug** (🐛 #DC2626) - Bug fixes and issues
8. **Feature** (✨ #059669) - New features and enhancements
9. **General** (📋 #6B7280) - General tasks

## Default Tags

Pre-populated tags:
- urgent, backend, frontend, database, api, ui, performance, security, refactor, enhancement

## Error Handling

Always wrap Tauri commands in try-catch:

```javascript
try {
    const todo = await invoke('create_todo', { /* ... */ });
    console.log('Success!', todo);
} catch (error) {
    console.error('Error:', error);
    // Show error message to user
}
```

## Common Patterns

### Today's Todos

```javascript
const today = new Date();
today.setHours(0, 0, 0, 0);
const tomorrow = new Date(today);
tomorrow.setDate(tomorrow.getDate() + 1);

// Get all todos (SQLite date filtering happens in backend)
const allTodos = await invoke('get_todos', { filters: null });

// Filter in frontend for today's due date
const todaysTodos = allTodos.filter(todo => {
    if (!todo.due_date) return false;
    const dueDate = new Date(todo.due_date);
    return dueDate >= today && dueDate < tomorrow;
});
```

### Overdue Todos

```javascript
const allTodos = await invoke('get_todos', { filters: null });
const now = new Date();

const overdue = allTodos.filter(todo => {
    if (!todo.due_date) return false;
    if (todo.status === 'completed' || todo.status === 'cancelled') return false;
    return new Date(todo.due_date) < now;
});
```

### Progress Tracking

```javascript
const stats = await invoke('get_todo_stats');

const totalTasks = stats.pending + stats.in_progress + stats.completed + stats.blocked;
const completedPercentage = (stats.completed / totalTasks) * 100;

console.log(`Progress: ${completedPercentage.toFixed(1)}%`);
```

### Agent vs Human Tasks

```javascript
// Get all agent tasks
const agentTodos = await invoke('get_todos', {
    filters: { assignee: 'agent' }
});

// Get all human tasks
const humanTodos = await invoke('get_todos', {
    filters: { assignee: 'human' }
});
```

## TypeScript Types

```typescript
interface Todo {
    id: string;
    title: string;
    description?: string;
    status: 'pending' | 'in_progress' | 'completed' | 'blocked' | 'cancelled';
    priority: 'low' | 'medium' | 'high' | 'critical';
    category?: string;
    assignee?: 'human' | 'agent';
    due_date?: string;  // ISO 8601
    parent_id?: string;
    tags: string[];
    created_at: string;  // ISO 8601
    updated_at: string;  // ISO 8601
    completed_at?: string;  // ISO 8601
    estimated_hours?: number;
    actual_hours?: number;
    metadata?: any;  // JSON
}

interface TodoFilters {
    status?: 'pending' | 'in_progress' | 'completed' | 'blocked' | 'cancelled';
    priority?: 'low' | 'medium' | 'high' | 'critical';
    category?: string;
    assignee?: 'human' | 'agent';
    search_query?: string;
    tags?: string[];
    include_subtasks?: boolean;
}

interface TodoStats {
    total: number;
    pending: number;
    in_progress: number;
    completed: number;
    blocked: number;
    cancelled: number;
    by_priority: {
        low: number;
        medium: number;
        high: number;
        critical: number;
    };
    overdue: number;
}

interface Category {
    id: string;
    name: string;
    color: string;
    icon?: string;
    created_at: string;
}

interface Tag {
    id: string;
    name: string;
    color?: string;
    usage_count: number;
}
```

## React Example Component

```tsx
import { useEffect, useState } from 'react';
import { invoke } from '@tauri-apps/api/tauri';

function TodoList() {
    const [todos, setTodos] = useState([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        loadTodos();
    }, []);

    async function loadTodos() {
        try {
            const data = await invoke('get_todos', { filters: null });
            setTodos(data);
        } catch (error) {
            console.error('Failed to load todos:', error);
        } finally {
            setLoading(false);
        }
    }

    async function toggleComplete(todo) {
        try {
            const newStatus = todo.status === 'completed' ? 'pending' : 'completed';
            await invoke('update_todo', {
                id: todo.id,
                updates: { status: newStatus }
            });
            await loadTodos();
        } catch (error) {
            console.error('Failed to update todo:', error);
        }
    }

    if (loading) return <div>Loading...</div>;

    return (
        <div>
            <h2>Todos ({todos.length})</h2>
            <ul>
                {todos.map(todo => (
                    <li key={todo.id}>
                        <input
                            type="checkbox"
                            checked={todo.status === 'completed'}
                            onChange={() => toggleComplete(todo)}
                        />
                        <span style={{
                            textDecoration: todo.status === 'completed' ? 'line-through' : 'none'
                        }}>
                            {todo.title}
                        </span>
                        <span className={`priority-${todo.priority}`}>
                            {todo.priority}
                        </span>
                    </li>
                ))}
            </ul>
        </div>
    );
}
```

## Troubleshooting

### Database not found
The database is automatically created on first run. If you see errors, check that `~/.anime-desktop/` directory has write permissions.

### Lock errors
SQLite locks the database during writes. If you see lock errors, the system will retry automatically. Increase timeout if needed.

### Corrupted database
```bash
sqlite3 ~/.anime-desktop/todos.db "PRAGMA integrity_check;"
```

## Documentation

For complete API documentation, see:
- `src/todos/README.md` - Complete API reference
- `schema/README.md` - Database schema details
- `TODOS_IMPLEMENTATION.md` - Implementation overview

---

**Happy task tracking!** 🚀
