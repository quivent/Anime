import { useMemo } from 'react'
import type { Todo, TodoStatus, TodoPriority } from '../types/todo'
import { useTodoStore } from '../store/todoStore'

interface TodoCardProps {
  todo: Todo
  onEdit: (todo: Todo) => void
  isDragging?: boolean
}

export default function TodoCard({ todo, onEdit, isDragging = false }: TodoCardProps) {
  const { toggleSelection, selectedTodos } = useTodoStore()
  const isSelected = selectedTodos.has(todo.id)

  const statusConfig: Record<TodoStatus, { label: string; color: string; bgColor: string; borderColor: string }> = {
    pending: { label: 'Pending', color: 'text-gray-400', bgColor: 'bg-gray-500/10', borderColor: 'border-gray-500/30' },
    in_progress: { label: 'In Progress', color: 'text-electric-400', bgColor: 'bg-electric-500/10', borderColor: 'border-electric-500/30' },
    completed: { label: 'Completed', color: 'text-mint-400', bgColor: 'bg-mint-500/10', borderColor: 'border-mint-500/30' },
    blocked: { label: 'Blocked', color: 'text-sunset-400', bgColor: 'bg-sunset-500/10', borderColor: 'border-sunset-500/30' },
  }

  const priorityConfig: Record<TodoPriority, { label: string; icon: string; color: string }> = {
    low: { label: 'Low', icon: '▼', color: 'text-gray-500' },
    medium: { label: 'Medium', icon: '◆', color: 'text-electric-400' },
    high: { label: 'High', icon: '▲', color: 'text-sunset-400' },
    critical: { label: 'Critical', icon: '⬆', color: 'text-sakura-400' },
    red: { label: 'RED', icon: '🔴', color: 'text-red-600' },
  }

  const categoryIcons: Record<string, string> = {
    feature: '✨',
    bugfix: '🐛',
    refactor: '♻️',
    docs: '📚',
    test: '🧪',
    deployment: '🚀',
    research: '🔬',
    design: '🎨',
    security: '🔒',
    ui: '🎨',
  }

  const isOverdue = useMemo(() => {
    if (!todo.dueDate || todo.status === 'completed') return false
    return new Date(todo.dueDate) < new Date()
  }, [todo.dueDate, todo.status])

  const formatDate = (date: string) => {
    const d = new Date(date)
    return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
  }

  return (
    <div
      className={`
        group bg-gray-800/50 border rounded-xl p-4 transition-all duration-200
        ${isDragging ? 'opacity-50 rotate-2' : ''}
        ${isSelected ? 'border-sakura-500/50 bg-sakura-500/5 anime-glow' : 'border-gray-700 hover:border-gray-600'}
        hover:bg-gray-800/70 cursor-pointer
      `}
      onClick={() => onEdit(todo)}
    >
      {/* Header */}
      <div className="flex items-start justify-between gap-3 mb-3">
        <div className="flex items-start gap-2 flex-1 min-w-0">
          <div className="flex-1 min-w-0">
            <h3 className="font-semibold text-gray-200 line-clamp-2 break-words">
              {todo.title}
            </h3>
          </div>
        </div>
      </div>

      {/* Description */}
      {todo.description && (
        <p className="text-sm text-gray-400 line-clamp-2 mb-3 break-words">
          {todo.description}
        </p>
      )}

      {/* Blocked Reason */}
      {todo.status === 'blocked' && todo.blockedReason && (
        <div className="mb-3 p-2 bg-sunset-500/10 border border-sunset-500/30 rounded text-xs text-sunset-400">
          <span className="font-semibold">Blocked: </span>
          {todo.blockedReason}
        </div>
      )}

      {/* Metadata */}
      <div className="flex flex-wrap gap-2 mb-3">
        {/* Status Badge */}
        <span
          className={`
            px-2 py-1 rounded text-xs font-medium border
            ${statusConfig[todo.status].color}
            ${statusConfig[todo.status].bgColor}
            ${statusConfig[todo.status].borderColor}
          `}
        >
          {statusConfig[todo.status].label}
        </span>

        {/* Priority */}
        <span className={`flex items-center gap-1 px-2 py-1 rounded text-xs font-medium bg-gray-700/50 ${priorityConfig[todo.priority].color}`}>
          <span>{priorityConfig[todo.priority].icon}</span>
          <span>{priorityConfig[todo.priority].label}</span>
        </span>

        {/* Category */}
        <span className="flex items-center gap-1 px-2 py-1 rounded text-xs font-medium bg-gray-700/50 text-gray-300">
          <span>{categoryIcons[todo.category] || '📋'}</span>
          <span className="capitalize">{todo.category}</span>
        </span>

        {/* Assignee */}
        <span className="px-2 py-1 rounded text-xs font-medium bg-gray-700/50 text-gray-300">
          {todo.assignee === 'agent' ? '🤖 Agent' : '👤 Human'}
        </span>
      </div>

      {/* Tags */}
      {todo.tags.length > 0 && (
        <div className="flex flex-wrap gap-1.5 mb-3">
          {todo.tags.map((tag) => (
            <span
              key={tag}
              className="px-2 py-0.5 rounded-full text-xs bg-electric-500/10 text-electric-400 border border-electric-500/20"
            >
              #{tag}
            </span>
          ))}
        </div>
      )}
    </div>
  )
}
