export type TodoStatus = 'pending' | 'in_progress' | 'completed' | 'blocked'
export type TodoPriority = 'low' | 'medium' | 'high' | 'critical' | 'red'
export type TodoCategory = 'feature' | 'bugfix' | 'refactor' | 'docs' | 'test' | 'deployment' | 'research' | 'design' | 'security' | 'ui'
export type TodoAssignee = 'human' | 'agent'

export interface Todo {
  id: string
  title: string
  description: string
  status: TodoStatus
  priority: TodoPriority
  category: TodoCategory
  assignee: TodoAssignee
  tags: string[]
  dueDate: string | null
  createdAt: string
  updatedAt: string
  completedAt: string | null
  blockedReason?: string
}

export interface TodoStats {
  total: number
  pending: number
  inProgress: number
  completed: number
  blocked: number
  overdue: number
  completionRate: number
}

export interface TodoFilters {
  status: TodoStatus | 'all'
  priority: TodoPriority | 'all'
  category: TodoCategory | 'all'
  assignee: TodoAssignee | 'all'
  searchQuery: string
}

export type TodoSortKey = 'createdAt' | 'updatedAt' | 'priority' | 'dueDate' | 'title'
export type TodoSortOrder = 'asc' | 'desc'
