import { create } from 'zustand'
import type { Todo, TodoStatus, TodoPriority, TodoFilters, TodoStats, TodoSortKey, TodoSortOrder } from '../types/todo'

interface TodoStore {
  todos: Todo[]
  filters: TodoFilters
  sortKey: TodoSortKey
  sortOrder: TodoSortOrder
  viewMode: 'kanban' | 'list'
  selectedTodos: Set<string>

  // Actions
  addTodo: (todo: Omit<Todo, 'id' | 'createdAt' | 'updatedAt' | 'completedAt'>) => void
  updateTodo: (id: string, updates: Partial<Todo>) => void
  deleteTodo: (id: string) => void
  updateTodoStatus: (id: string, status: TodoStatus) => void

  // Bulk actions
  bulkUpdateStatus: (ids: string[], status: TodoStatus) => void
  bulkDelete: (ids: string[]) => void

  // Reorder
  reorderTodos: (startIndex: number, endIndex: number) => void

  // Selection
  toggleSelection: (id: string) => void
  selectAll: () => void
  clearSelection: () => void

  // Filters & Sorting
  setFilter: (key: keyof TodoFilters, value: string) => void
  resetFilters: () => void
  setSortKey: (key: TodoSortKey) => void
  toggleSortOrder: () => void

  // View
  setViewMode: (mode: 'kanban' | 'list') => void

  // Computed
  getFilteredTodos: () => Todo[]
  getTodosByStatus: (status: TodoStatus) => Todo[]
  getStats: () => TodoStats
}

const defaultFilters: TodoFilters = {
  status: 'all',
  priority: 'all',
  category: 'all',
  assignee: 'all',
  searchQuery: '',
}

const mockTodos: Todo[] = [
  { id: '0', title: 'Fix duplicate terminal_connect_local function causing compilation failure', description: '', status: 'pending', priority: 'red', category: 'bugfix', assignee: 'agent', tags: [], dueDate: null, createdAt: '2025-11-20T09:59:00Z', updatedAt: '2025-11-20T09:59:00Z', completedAt: null },
  { id: '1', title: 'Fix incomplete API key save function in LambdaView', description: '', status: 'pending', priority: 'critical', category: 'bugfix', assignee: 'agent', tags: [], dueDate: null, createdAt: '2025-11-20T10:00:00Z', updatedAt: '2025-11-20T10:00:00Z', completedAt: null },
  { id: '2', title: 'Fix CPU load average type mismatch in ServerMonitor', description: '', status: 'pending', priority: 'critical', category: 'bugfix', assignee: 'agent', tags: [], dueDate: null, createdAt: '2025-11-20T10:01:00Z', updatedAt: '2025-11-20T10:01:00Z', completedAt: null },
  { id: '3', title: 'Add missing GPU driver version to GpuStatus struct', description: '', status: 'pending', priority: 'critical', category: 'bugfix', assignee: 'agent', tags: [], dueDate: null, createdAt: '2025-11-20T10:02:00Z', updatedAt: '2025-11-20T10:02:00Z', completedAt: null },
  { id: '4', title: 'Implement missing network telemetry fields', description: '', status: 'pending', priority: 'critical', category: 'bugfix', assignee: 'agent', tags: [], dueDate: null, createdAt: '2025-11-20T10:03:00Z', updatedAt: '2025-11-20T10:03:00Z', completedAt: null },
  { id: '5', title: 'Implement execute_server_command functionality', description: '', status: 'pending', priority: 'critical', category: 'feature', assignee: 'agent', tags: [], dueDate: null, createdAt: '2025-11-20T10:04:00Z', updatedAt: '2025-11-20T10:04:00Z', completedAt: null },
  { id: '6', title: 'Fix uninitialized invoke call in handleSaveApiKey', description: '', status: 'pending', priority: 'critical', category: 'bugfix', assignee: 'agent', tags: [], dueDate: null, createdAt: '2025-11-20T10:05:00Z', updatedAt: '2025-11-20T10:05:00Z', completedAt: null },
  { id: '7', title: 'Fix API key validation flow error handling', description: '', status: 'pending', priority: 'high', category: 'bugfix', assignee: 'agent', tags: [], dueDate: null, createdAt: '2025-11-20T10:06:00Z', updatedAt: '2025-11-20T10:06:00Z', completedAt: null },
  { id: '8', title: 'Add GPU specs to InstanceTypeName type', description: '', status: 'pending', priority: 'high', category: 'feature', assignee: 'agent', tags: [], dueDate: null, createdAt: '2025-11-20T10:07:00Z', updatedAt: '2025-11-20T10:07:00Z', completedAt: null },
  { id: '9', title: 'Add SSH key generation command', description: '', status: 'pending', priority: 'high', category: 'feature', assignee: 'agent', tags: [], dueDate: null, createdAt: '2025-11-20T10:08:00Z', updatedAt: '2025-11-20T10:08:00Z', completedAt: null },
  { id: '10', title: 'Add automatic SSH key selection for instance monitor', description: '', status: 'pending', priority: 'high', category: 'feature', assignee: 'agent', tags: [], dueDate: null, createdAt: '2025-11-20T10:09:00Z', updatedAt: '2025-11-20T10:09:00Z', completedAt: null },
  { id: '11', title: 'Fix InstanceStatus enum serialization mismatch', description: '', status: 'pending', priority: 'high', category: 'bugfix', assignee: 'agent', tags: [], dueDate: null, createdAt: '2025-11-20T10:10:00Z', updatedAt: '2025-11-20T10:10:00Z', completedAt: null },
  { id: '12', title: 'Add proper TypeScript types for Instance optional fields', description: '', status: 'pending', priority: 'high', category: 'refactor', assignee: 'agent', tags: [], dueDate: null, createdAt: '2025-11-20T10:11:00Z', updatedAt: '2025-11-20T10:11:00Z', completedAt: null },
  { id: '13', title: 'Create custom error types instead of generic String', description: '', status: 'pending', priority: 'high', category: 'refactor', assignee: 'agent', tags: [], dueDate: null, createdAt: '2025-11-20T10:12:00Z', updatedAt: '2025-11-20T10:12:00Z', completedAt: null },
  { id: '14', title: 'Add timestamp format validation in TypeScript', description: '', status: 'pending', priority: 'high', category: 'bugfix', assignee: 'agent', tags: [], dueDate: null, createdAt: '2025-11-20T10:13:00Z', updatedAt: '2025-11-20T10:13:00Z', completedAt: null },
  { id: '15', title: 'Add uptime_seconds field to TypeScript ProcessInfo', description: '', status: 'pending', priority: 'high', category: 'bugfix', assignee: 'agent', tags: [], dueDate: null, createdAt: '2025-11-20T10:14:00Z', updatedAt: '2025-11-20T10:14:00Z', completedAt: null },
  { id: '16', title: 'Fix terminal session cleanup on unmount', description: '', status: 'pending', priority: 'high', category: 'bugfix', assignee: 'agent', tags: [], dueDate: null, createdAt: '2025-11-20T10:15:00Z', updatedAt: '2025-11-20T10:15:00Z', completedAt: null },
  { id: '17', title: 'Clear polling interval on view change', description: '', status: 'pending', priority: 'high', category: 'bugfix', assignee: 'agent', tags: [], dueDate: null, createdAt: '2025-11-20T10:16:00Z', updatedAt: '2025-11-20T10:16:00Z', completedAt: null },
  { id: '18', title: 'Add mechanism to force-close orphaned SSH connections', description: '', status: 'pending', priority: 'high', category: 'feature', assignee: 'agent', tags: [], dueDate: null, createdAt: '2025-11-20T10:17:00Z', updatedAt: '2025-11-20T10:17:00Z', completedAt: null },
  { id: '19', title: 'Fix server status polling race condition', description: '', status: 'pending', priority: 'high', category: 'bugfix', assignee: 'agent', tags: [], dueDate: null, createdAt: '2025-11-20T10:18:00Z', updatedAt: '2025-11-20T10:18:00Z', completedAt: null },
  { id: '20', title: 'Add ComfyUI workflow execution monitoring with WebSocket', description: '', status: 'pending', priority: 'medium', category: 'feature', assignee: 'agent', tags: [], dueDate: null, createdAt: '2025-11-20T10:19:00Z', updatedAt: '2025-11-20T10:19:00Z', completedAt: null },
  { id: '21', title: 'Implement file system management UI', description: '', status: 'pending', priority: 'medium', category: 'feature', assignee: 'agent', tags: [], dueDate: null, createdAt: '2025-11-20T10:20:00Z', updatedAt: '2025-11-20T10:20:00Z', completedAt: null },
  { id: '22', title: 'Add SSH key rotation mechanism', description: '', status: 'pending', priority: 'medium', category: 'feature', assignee: 'agent', tags: [], dueDate: null, createdAt: '2025-11-20T10:21:00Z', updatedAt: '2025-11-20T10:21:00Z', completedAt: null },
  { id: '23', title: 'Implement historical GPU metrics tracking', description: '', status: 'pending', priority: 'medium', category: 'feature', assignee: 'agent', tags: [], dueDate: null, createdAt: '2025-11-20T10:22:00Z', updatedAt: '2025-11-20T10:22:00Z', completedAt: null },
  { id: '24', title: 'Add cost tracking and billing forecasts', description: '', status: 'pending', priority: 'medium', category: 'feature', assignee: 'agent', tags: [], dueDate: null, createdAt: '2025-11-20T10:23:00Z', updatedAt: '2025-11-20T10:23:00Z', completedAt: null },
  { id: '25', title: 'Implement Lambda instance snapshot functionality', description: '', status: 'pending', priority: 'medium', category: 'feature', assignee: 'agent', tags: [], dueDate: null, createdAt: '2025-11-20T10:24:00Z', updatedAt: '2025-11-20T10:24:00Z', completedAt: null },
  { id: '26', title: 'Add load balancer configuration UI', description: '', status: 'pending', priority: 'medium', category: 'feature', assignee: 'human', tags: [], dueDate: null, createdAt: '2025-11-20T10:25:00Z', updatedAt: '2025-11-20T10:25:00Z', completedAt: null },
  { id: '27', title: 'Implement VPC and security group management', description: '', status: 'pending', priority: 'medium', category: 'feature', assignee: 'human', tags: [], dueDate: null, createdAt: '2025-11-20T10:26:00Z', updatedAt: '2025-11-20T10:26:00Z', completedAt: null },
  { id: '28', title: 'Add resumable model download support', description: '', status: 'pending', priority: 'medium', category: 'feature', assignee: 'agent', tags: [], dueDate: null, createdAt: '2025-11-20T10:27:00Z', updatedAt: '2025-11-20T10:27:00Z', completedAt: null },
  { id: '29', title: 'Implement exponential backoff retry logic', description: '', status: 'pending', priority: 'medium', category: 'feature', assignee: 'agent', tags: [], dueDate: null, createdAt: '2025-11-20T10:28:00Z', updatedAt: '2025-11-20T10:28:00Z', completedAt: null },
  { id: '30', title: 'Add input validation for instance names and paths', description: '', status: 'pending', priority: 'medium', category: 'feature', assignee: 'agent', tags: [], dueDate: null, createdAt: '2025-11-20T10:29:00Z', updatedAt: '2025-11-20T10:29:00Z', completedAt: null },
  { id: '31', title: 'Encrypt Lambda API key in storage', description: '', status: 'pending', priority: 'medium', category: 'security', assignee: 'agent', tags: [], dueDate: null, createdAt: '2025-11-20T10:30:00Z', updatedAt: '2025-11-20T10:30:00Z', completedAt: null },
  { id: '32', title: 'Add network timeout handling for long operations', description: '', status: 'pending', priority: 'medium', category: 'feature', assignee: 'agent', tags: [], dueDate: null, createdAt: '2025-11-20T10:31:00Z', updatedAt: '2025-11-20T10:31:00Z', completedAt: null },
  { id: '33', title: 'Add graceful degradation when GPU info unavailable', description: '', status: 'pending', priority: 'medium', category: 'feature', assignee: 'agent', tags: [], dueDate: null, createdAt: '2025-11-20T10:32:00Z', updatedAt: '2025-11-20T10:32:00Z', completedAt: null },
  { id: '34', title: 'Implement rate limiting for Lambda API requests', description: '', status: 'pending', priority: 'medium', category: 'feature', assignee: 'agent', tags: [], dueDate: null, createdAt: '2025-11-20T10:33:00Z', updatedAt: '2025-11-20T10:33:00Z', completedAt: null },
  { id: '35', title: 'Fix stderr capture in SSH command execution', description: '', status: 'pending', priority: 'medium', category: 'bugfix', assignee: 'agent', tags: [], dueDate: null, createdAt: '2025-11-20T10:34:00Z', updatedAt: '2025-11-20T10:34:00Z', completedAt: null },
  { id: '36', title: 'Implement SSH connection pooling', description: '', status: 'pending', priority: 'medium', category: 'feature', assignee: 'agent', tags: [], dueDate: null, createdAt: '2025-11-20T10:35:00Z', updatedAt: '2025-11-20T10:35:00Z', completedAt: null },
  { id: '37', title: 'Add TLS certificate verification for Lambda API', description: '', status: 'pending', priority: 'medium', category: 'security', assignee: 'agent', tags: [], dueDate: null, createdAt: '2025-11-20T10:36:00Z', updatedAt: '2025-11-20T10:36:00Z', completedAt: null },
  { id: '38', title: 'Improve error messages with actionable details', description: '', status: 'pending', priority: 'medium', category: 'refactor', assignee: 'agent', tags: [], dueDate: null, createdAt: '2025-11-20T10:37:00Z', updatedAt: '2025-11-20T10:37:00Z', completedAt: null },
  { id: '39', title: 'Fix instance launch dialog auto-close timing', description: '', status: 'pending', priority: 'low', category: 'ui', assignee: 'agent', tags: [], dueDate: null, createdAt: '2025-11-20T10:38:00Z', updatedAt: '2025-11-20T10:38:00Z', completedAt: null },
  { id: '40', title: 'Add loading indicator to region selection', description: '', status: 'pending', priority: 'low', category: 'ui', assignee: 'agent', tags: [], dueDate: null, createdAt: '2025-11-20T10:39:00Z', updatedAt: '2025-11-20T10:39:00Z', completedAt: null },
  { id: '41', title: 'Make terminal output scrollable', description: '', status: 'pending', priority: 'low', category: 'ui', assignee: 'agent', tags: [], dueDate: null, createdAt: '2025-11-20T10:40:00Z', updatedAt: '2025-11-20T10:40:00Z', completedAt: null },
  { id: '42', title: 'Fix server monitor modal dismissal on connection failure', description: '', status: 'pending', priority: 'low', category: 'ui', assignee: 'agent', tags: [], dueDate: null, createdAt: '2025-11-20T10:41:00Z', updatedAt: '2025-11-20T10:41:00Z', completedAt: null },
  { id: '43', title: 'Add keyboard shortcuts for common actions', description: '', status: 'pending', priority: 'low', category: 'ui', assignee: 'agent', tags: [], dueDate: null, createdAt: '2025-11-20T10:42:00Z', updatedAt: '2025-11-20T10:42:00Z', completedAt: null },
  { id: '44', title: 'Fix instance status auto-refresh on view change', description: '', status: 'pending', priority: 'low', category: 'bugfix', assignee: 'agent', tags: [], dueDate: null, createdAt: '2025-11-20T10:43:00Z', updatedAt: '2025-11-20T10:43:00Z', completedAt: null },
  { id: '45', title: 'Add confirmation dialog for SSH key deletion', description: '', status: 'pending', priority: 'low', category: 'ui', assignee: 'agent', tags: [], dueDate: null, createdAt: '2025-11-20T10:44:00Z', updatedAt: '2025-11-20T10:44:00Z', completedAt: null },
  { id: '46', title: 'Make Lambda API base URL configurable', description: '', status: 'pending', priority: 'low', category: 'feature', assignee: 'agent', tags: [], dueDate: null, createdAt: '2025-11-20T10:45:00Z', updatedAt: '2025-11-20T10:45:00Z', completedAt: null },
  { id: '47', title: 'Add ComfyUI port configuration option', description: '', status: 'pending', priority: 'low', category: 'feature', assignee: 'agent', tags: [], dueDate: null, createdAt: '2025-11-20T10:46:00Z', updatedAt: '2025-11-20T10:46:00Z', completedAt: null },
  { id: '48', title: 'Support environment variables for API keys', description: '', status: 'pending', priority: 'low', category: 'feature', assignee: 'agent', tags: [], dueDate: null, createdAt: '2025-11-20T10:47:00Z', updatedAt: '2025-11-20T10:47:00Z', completedAt: null },
  { id: '49', title: 'Implement configuration file loading', description: '', status: 'pending', priority: 'low', category: 'feature', assignee: 'agent', tags: [], dueDate: null, createdAt: '2025-11-20T10:48:00Z', updatedAt: '2025-11-20T10:48:00Z', completedAt: null },
  { id: '50', title: 'Make SSH username configurable', description: '', status: 'pending', priority: 'low', category: 'feature', assignee: 'agent', tags: [], dueDate: null, createdAt: '2025-11-20T10:49:00Z', updatedAt: '2025-11-20T10:49:00Z', completedAt: null },
  { id: '51', title: 'Write unit tests for Lambda client', description: '', status: 'pending', priority: 'low', category: 'test', assignee: 'agent', tags: [], dueDate: null, createdAt: '2025-11-20T10:50:00Z', updatedAt: '2025-11-20T10:50:00Z', completedAt: null },
  { id: '52', title: 'Add error recovery tests', description: '', status: 'pending', priority: 'low', category: 'test', assignee: 'agent', tags: [], dueDate: null, createdAt: '2025-11-20T10:51:00Z', updatedAt: '2025-11-20T10:51:00Z', completedAt: null },
  { id: '53', title: 'Create integration test suite', description: '', status: 'pending', priority: 'low', category: 'test', assignee: 'agent', tags: [], dueDate: null, createdAt: '2025-11-20T10:52:00Z', updatedAt: '2025-11-20T10:52:00Z', completedAt: null },
  { id: '54', title: 'Write setup and architecture documentation', description: '', status: 'pending', priority: 'low', category: 'docs', assignee: 'human', tags: [], dueDate: null, createdAt: '2025-11-20T10:53:00Z', updatedAt: '2025-11-20T10:53:00Z', completedAt: null },
  { id: '55', title: 'Enable TypeScript strict mode', description: '', status: 'pending', priority: 'low', category: 'refactor', assignee: 'agent', tags: [], dueDate: null, createdAt: '2025-11-20T10:54:00Z', updatedAt: '2025-11-20T10:54:00Z', completedAt: null },
  { id: '56', title: 'Prevent API key exposure in error messages', description: '', status: 'pending', priority: 'low', category: 'security', assignee: 'agent', tags: [], dueDate: null, createdAt: '2025-11-20T10:55:00Z', updatedAt: '2025-11-20T10:55:00Z', completedAt: null },
]

export const useTodoStore = create<TodoStore>((set, get) => ({
  todos: mockTodos,
  filters: defaultFilters,
  sortKey: 'createdAt',
  sortOrder: 'desc',
  viewMode: 'kanban',
  selectedTodos: new Set(),

  addTodo: (todoData) => {
    const newTodo: Todo = {
      ...todoData,
      id: Date.now().toString(),
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
      completedAt: null,
    }
    set({ todos: [...get().todos, newTodo] })
  },

  updateTodo: (id, updates) => {
    set({
      todos: get().todos.map((todo) =>
        todo.id === id
          ? {
              ...todo,
              ...updates,
              updatedAt: new Date().toISOString(),
              completedAt:
                updates.status === 'completed' && todo.status !== 'completed'
                  ? new Date().toISOString()
                  : updates.status !== 'completed'
                  ? null
                  : todo.completedAt,
            }
          : todo
      ),
    })
  },

  deleteTodo: (id) => {
    set({ todos: get().todos.filter((todo) => todo.id !== id) })
    const selected = new Set(get().selectedTodos)
    selected.delete(id)
    set({ selectedTodos: selected })
  },

  updateTodoStatus: (id, status) => {
    get().updateTodo(id, { status })
  },

  bulkUpdateStatus: (ids, status) => {
    ids.forEach((id) => get().updateTodoStatus(id, status))
  },

  bulkDelete: (ids) => {
    ids.forEach((id) => get().deleteTodo(id))
  },

  reorderTodos: (startIndex, endIndex) => {
    const todos = [...get().todos]
    const [removed] = todos.splice(startIndex, 1)
    todos.splice(endIndex, 0, removed)
    set({ todos })
  },

  toggleSelection: (id) => {
    const selected = new Set(get().selectedTodos)
    if (selected.has(id)) {
      selected.delete(id)
    } else {
      selected.add(id)
    }
    set({ selectedTodos: selected })
  },

  selectAll: () => {
    const filtered = get().getFilteredTodos()
    set({ selectedTodos: new Set(filtered.map((t) => t.id)) })
  },

  clearSelection: () => {
    set({ selectedTodos: new Set() })
  },

  setFilter: (key, value) => {
    set({ filters: { ...get().filters, [key]: value } })
  },

  resetFilters: () => {
    set({ filters: defaultFilters })
  },

  setSortKey: (key) => {
    set({ sortKey: key })
  },

  toggleSortOrder: () => {
    set({ sortOrder: get().sortOrder === 'asc' ? 'desc' : 'asc' })
  },

  setViewMode: (mode) => {
    set({ viewMode: mode })
  },

  getFilteredTodos: () => {
    const { todos, filters, sortKey, sortOrder } = get()

    let filtered = todos.filter((todo) => {
      if (filters.status !== 'all' && todo.status !== filters.status) return false
      if (filters.priority !== 'all' && todo.priority !== filters.priority) return false
      if (filters.category !== 'all' && todo.category !== filters.category) return false
      if (filters.assignee !== 'all' && todo.assignee !== filters.assignee) return false
      if (filters.searchQuery) {
        const query = filters.searchQuery.toLowerCase()
        return (
          todo.title.toLowerCase().includes(query) ||
          todo.description.toLowerCase().includes(query) ||
          todo.tags.some((tag) => tag.toLowerCase().includes(query))
        )
      }
      return true
    })

    // Sort
    filtered.sort((a, b) => {
      let aVal: any = a[sortKey]
      let bVal: any = b[sortKey]

      // Handle null values
      if (aVal === null) return sortOrder === 'asc' ? 1 : -1
      if (bVal === null) return sortOrder === 'asc' ? -1 : 1

      // Priority has special ordering
      if (sortKey === 'priority') {
        const priorityOrder: Record<TodoPriority, number> = {
          red: 5,
          critical: 4,
          high: 3,
          medium: 2,
          low: 1,
        }
        aVal = priorityOrder[a.priority]
        bVal = priorityOrder[b.priority]
      }

      if (aVal < bVal) return sortOrder === 'asc' ? -1 : 1
      if (aVal > bVal) return sortOrder === 'asc' ? 1 : -1
      return 0
    })

    return filtered
  },

  getTodosByStatus: (status) => {
    return get().getFilteredTodos().filter((todo) => todo.status === status)
  },

  getStats: () => {
    const todos = get().todos
    const now = new Date()

    const stats: TodoStats = {
      total: todos.length,
      pending: todos.filter((t) => t.status === 'pending').length,
      inProgress: todos.filter((t) => t.status === 'in_progress').length,
      completed: todos.filter((t) => t.status === 'completed').length,
      blocked: todos.filter((t) => t.status === 'blocked').length,
      overdue: todos.filter((t) => {
        if (!t.dueDate || t.status === 'completed') return false
        return new Date(t.dueDate) < now
      }).length,
      completionRate: todos.length > 0
        ? Math.round((todos.filter((t) => t.status === 'completed').length / todos.length) * 100)
        : 0,
    }

    return stats
  },
}))
