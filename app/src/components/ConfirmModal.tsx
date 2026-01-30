interface ConfirmModalProps {
  isOpen: boolean
  title: string
  message: string
  confirmText?: string
  cancelText?: string
  onConfirm: () => void
  onCancel: () => void
  variant?: 'danger' | 'warning' | 'info'
}

export function ConfirmModal({
  isOpen,
  title,
  message,
  confirmText = 'Confirm',
  cancelText = 'Cancel',
  onConfirm,
  onCancel,
  variant = 'info'
}: ConfirmModalProps) {
  if (!isOpen) return null

  const getVariantStyles = () => {
    switch (variant) {
      case 'danger':
        return {
          headerBg: 'bg-sunset-500/10 border-sunset-500/30',
          headerText: 'text-sunset-400',
          confirmButton: 'bg-sunset-500 hover:bg-sunset-600'
        }
      case 'warning':
        return {
          headerBg: 'bg-electric-500/10 border-electric-500/30',
          headerText: 'text-electric-400',
          confirmButton: 'bg-electric-500 hover:bg-electric-600'
        }
      case 'info':
        return {
          headerBg: 'bg-sakura-500/10 border-sakura-500/30',
          headerText: 'text-sakura-400',
          confirmButton: 'bg-sakura-500 hover:bg-sakura-600'
        }
    }
  }

  const styles = getVariantStyles()

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center animate-fade-in">
      {/* Backdrop */}
      <div
        className="absolute inset-0 bg-black/60 backdrop-blur-sm"
        onClick={onCancel}
      />

      {/* Modal */}
      <div className="relative bg-gray-900 rounded-lg border-2 border-gray-800 shadow-2xl max-w-md w-full mx-4 animate-scale-in">
        {/* Header */}
        <div className={`px-6 py-4 border-b-2 ${styles.headerBg}`}>
          <h3 className={`text-lg font-bold ${styles.headerText}`}>
            {title}
          </h3>
        </div>

        {/* Content */}
        <div className="px-6 py-4">
          <p className="text-gray-300 whitespace-pre-line">
            {message}
          </p>
        </div>

        {/* Actions */}
        <div className="px-6 py-4 border-t border-gray-800 flex items-center justify-end gap-3">
          <button
            onClick={onCancel}
            className="px-4 py-2 rounded-lg bg-gray-800 text-gray-300 hover:bg-gray-700 transition-colors font-medium"
          >
            {cancelText}
          </button>
          <button
            onClick={onConfirm}
            className={`px-4 py-2 rounded-lg text-white transition-colors font-medium ${styles.confirmButton}`}
          >
            {confirmText}
          </button>
        </div>
      </div>
    </div>
  )
}
