import { useEffect } from 'react'

export type ToastType = 'success' | 'error' | 'warning' | 'info'

export interface ToastMessage {
  id: string
  message: string
  type: ToastType
  duration?: number
}

interface ToastProps {
  toast: ToastMessage
  onClose: (id: string) => void
}

export function Toast({ toast, onClose }: ToastProps) {
  useEffect(() => {
    const duration = toast.duration || 5000
    const timer = setTimeout(() => {
      onClose(toast.id)
    }, duration)

    return () => clearTimeout(timer)
  }, [toast.id, toast.duration, onClose])

  const getToastStyles = () => {
    switch (toast.type) {
      case 'success':
        return 'bg-mint-500/90 border-mint-400 text-white'
      case 'error':
        return 'bg-sunset-500/90 border-sunset-400 text-white'
      case 'warning':
        return 'bg-electric-500/90 border-electric-400 text-white'
      case 'info':
        return 'bg-neon-500/90 border-neon-400 text-white'
      default:
        return 'bg-gray-800/90 border-gray-700 text-white'
    }
  }

  const getIcon = () => {
    switch (toast.type) {
      case 'success':
        return '✓'
      case 'error':
        return '✗'
      case 'warning':
        return '⚠'
      case 'info':
        return 'ℹ'
      default:
        return ''
    }
  }

  return (
    <div
      className={`
        flex items-center gap-3 px-4 py-3 rounded-lg border-2 shadow-lg
        animate-slide-in-right backdrop-blur-sm
        ${getToastStyles()}
      `}
    >
      <div className="text-xl font-bold">{getIcon()}</div>
      <div className="flex-1 text-sm font-medium whitespace-pre-line">{toast.message}</div>
      <button
        onClick={() => onClose(toast.id)}
        className="text-white/80 hover:text-white transition-colors ml-2"
      >
        ✕
      </button>
    </div>
  )
}

interface ToastContainerProps {
  toasts: ToastMessage[]
  onClose: (id: string) => void
}

export function ToastContainer({ toasts, onClose }: ToastContainerProps) {
  return (
    <div className="fixed top-4 right-4 z-50 flex flex-col gap-2 max-w-md">
      {toasts.map((toast) => (
        <Toast key={toast.id} toast={toast} onClose={onClose} />
      ))}
    </div>
  )
}
