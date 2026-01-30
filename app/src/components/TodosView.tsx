import { useState } from 'react'
import { useTodoStore } from '../store/todoStore'
import TodoCard from './TodoCard'
import TodoModal from './TodoModal'
import type { Todo, TodoStatus, TodoSortKey } from '../types/todo'

export default function TodosView() {
  const {
    todos,
    filters,
    sortKey,
    sortOrder,
    viewMode,
    selectedTodos,
    setFilter,
    resetFilters,
    setSortKey,
    toggleSortOrder,
    setViewMode,
    getFilteredTodos,
    getTodosByStatus,
    getStats,
    selectAll,
    clearSelection,
    bulkUpdateStatus,
    bulkDelete,
    reorderTodos,
  } = useTodoStore()

  const [showModal, setShowModal] = useState(false)
  const [editingTodo, setEditingTodo] = useState<Todo | undefined>(undefined)
  const [showBulkActions, setShowBulkActions] = useState(false)
  const [draggedIndex, setDraggedIndex] = useState<number | null>(null)

  const stats = getStats()
  const filteredTodos = getFilteredTodos()

  const handleEditTodo = (todo: Todo) => {
    setEditingTodo(todo)
    setShowModal(true)
  }

  const handleCloseModal = () => {
    setShowModal(false)
    setEditingTodo(undefined)
  }

  const handleCreateNew = () => {
    setEditingTodo(undefined)
    setShowModal(true)
  }

  const handleBulkAction = (action: 'delete' | TodoStatus) => {
    const selectedIds = Array.from(selectedTodos)
    if (selectedIds.length === 0) return

    if (action === 'delete') {
      if (confirm(`Delete ${selectedIds.length} selected todos?`)) {
        bulkDelete(selectedIds)
        clearSelection()
      }
    } else {
      bulkUpdateStatus(selectedIds, action)
      clearSelection()
    }
    setShowBulkActions(false)
  }

  const handleDragStart = (index: number) => {
    setDraggedIndex(index)
  }

  const handleDragOver = (e: React.DragEvent, index: number) => {
    e.preventDefault()
    if (draggedIndex === null || draggedIndex === index) return

    reorderTodos(draggedIndex, index)
    setDraggedIndex(index)
  }

  const handleDragEnd = () => {
    setDraggedIndex(null)
  }

  // Kanban columns
  const kanbanColumns: { status: TodoStatus; label: string; color: string }[] = [
    { status: 'pending', label: 'Pending', color: 'gray' },
    { status: 'in_progress', label: 'In Progress', color: 'electric' },
    { status: 'completed', label: 'Completed', color: 'mint' },
    { status: 'blocked', label: 'Blocked', color: 'sunset' },
  ]

  return (
    <div className="flex flex-col h-full overflow-hidden">
      {/* Header */}
      <div className="p-6 border-b border-gray-800 bg-gray-900/50">
        <div className="flex items-center justify-between mb-4">
          <div>
            <h2 className="text-2xl font-bold text-electric-400 flex items-center gap-3">
              <span className="text-3xl">📋</span>
              Todos
            </h2>
            <p className="text-gray-400 mt-1 text-sm">
              Manage your tasks and track progress
            </p>
          </div>

          <button
            onClick={handleCreateNew}
            className="px-6 py-3 bg-electric-500/20 hover:bg-electric-500/30 border border-electric-500/50 rounded-lg text-electric-400 font-medium transition-all flex items-center gap-2 anime-glow-electric"
          >
            <span className="text-xl">+</span>
            <span>New Todo</span>
          </button>
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-auto p-6">
        {filteredTodos.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-full text-center">
            <div className="text-6xl mb-4">📭</div>
            <h3 className="text-xl font-semibold text-gray-300 mb-2">No todos found</h3>
            <p className="text-gray-500 mb-6">
              {todos.length === 0
                ? 'Create your first todo to get started'
                : 'Try adjusting your filters'}
            </p>
            {todos.length === 0 && (
              <button
                onClick={handleCreateNew}
                className="px-6 py-3 bg-electric-500/20 hover:bg-electric-500/30 border border-electric-500/50 rounded-lg text-electric-400 font-medium transition-all"
              >
                Create Your First Todo
              </button>
            )}
          </div>
        ) : (
          /* List View */
          <div className="space-y-3">
            {filteredTodos.map((todo, index) => (
              <div
                key={todo.id}
                draggable
                onDragStart={() => handleDragStart(index)}
                onDragOver={(e) => handleDragOver(e, index)}
                onDragEnd={handleDragEnd}
                className="cursor-move"
              >
                <TodoCard
                  todo={todo}
                  onEdit={handleEditTodo}
                  isDragging={draggedIndex === index}
                />
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Modal */}
      {showModal && <TodoModal todo={editingTodo} onClose={handleCloseModal} />}
    </div>
  )
}
