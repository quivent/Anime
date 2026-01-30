import { useState, useEffect } from 'react'
import type { Todo } from '../types/todo'
import { useTodoStore } from '../store/todoStore'

interface TodoModalProps {
  todo?: Todo
  onClose: () => void
}

export default function TodoModal({ todo, onClose }: TodoModalProps) {
  const { addTodo, updateTodo } = useTodoStore()
  const isEditing = !!todo

  const [formData, setFormData] = useState({
    title: todo?.title || '',
    description: todo?.description || '',
  })

  const [errors, setErrors] = useState<Record<string, string>>({})

  useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        onClose()
      }
    }
    window.addEventListener('keydown', handleEscape)
    return () => window.removeEventListener('keydown', handleEscape)
  }, [onClose])

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()

    // Validation
    const newErrors: Record<string, string> = {}
    if (!formData.title.trim()) {
      newErrors.title = 'Title is required'
    }
    if (formData.title.length > 200) {
      newErrors.title = 'Title must be less than 200 characters'
    }
    if (formData.description.length > 2000) {
      newErrors.description = 'Description must be less than 2000 characters'
    }

    if (Object.keys(newErrors).length > 0) {
      setErrors(newErrors)
      return
    }

    if (isEditing && todo) {
      updateTodo(todo.id, formData)
    } else {
      addTodo(formData)
    }

    onClose()
  }

  return (
    <div className="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50 p-4 animate-fade-in">
      <div className="bg-gray-900 border border-electric-500/30 rounded-xl max-w-2xl w-full max-h-[90vh] overflow-auto anime-glow animate-scale-in">
        {/* Header */}
        <div className="p-6 border-b border-gray-800">
          <h2 className="text-2xl font-bold text-electric-400">
            {isEditing ? 'Edit Todo' : 'Create New Todo'}
          </h2>
          <p className="text-gray-400 mt-1 text-sm">
            {isEditing ? 'Update the details below' : 'Fill in the details to create a new todo'}
          </p>
        </div>

        {/* Form */}
        <form onSubmit={handleSubmit} className="p-6 space-y-5">
          {/* Title */}
          <div>
            <label className="block text-sm font-medium text-gray-300 mb-2">
              Title <span className="text-sunset-400">*</span>
            </label>
            <input
              type="text"
              value={formData.title}
              onChange={(e) => setFormData({ ...formData, title: e.target.value })}
              placeholder="Enter a descriptive title..."
              className={`w-full px-4 py-3 bg-gray-800/50 border rounded-lg focus:outline-none focus:border-electric-500 text-white text-sm ${
                errors.title ? 'border-sunset-500' : 'border-gray-700'
              }`}
              autoFocus
            />
            {errors.title && <p className="mt-1 text-xs text-sunset-400">{errors.title}</p>}
          </div>

          {/* Description */}
          <div>
            <label className="block text-sm font-medium text-gray-300 mb-2">
              Description
            </label>
            <textarea
              value={formData.description}
              onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              placeholder="Add detailed description..."
              rows={4}
              className={`w-full px-4 py-3 bg-gray-800/50 border rounded-lg focus:outline-none focus:border-electric-500 text-white text-sm resize-none ${
                errors.description ? 'border-sunset-500' : 'border-gray-700'
              }`}
            />
            {errors.description && <p className="mt-1 text-xs text-sunset-400">{errors.description}</p>}
          </div>
        </form>

        {/* Footer */}
        <div className="p-6 border-t border-gray-800 flex gap-3">
          <button
            type="button"
            onClick={onClose}
            className="flex-1 px-6 py-3 bg-gray-800/50 hover:bg-gray-700/50 border border-gray-700 rounded-lg text-gray-300 transition-all"
          >
            Cancel
          </button>
          <button
            type="submit"
            onClick={handleSubmit}
            className="flex-1 px-6 py-3 bg-electric-500/20 hover:bg-electric-500/30 border border-electric-500/50 rounded-lg text-electric-400 font-medium transition-all"
          >
            {isEditing ? 'Update Todo' : 'Create Todo'}
          </button>
        </div>
      </div>
    </div>
  )
}
