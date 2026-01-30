interface ParsingStep {
  id: string
  name: string
  status: 'pending' | 'in_progress' | 'completed' | 'error'
  progress?: number
  details?: string
}

interface ParsingProgressProps {
  steps: ParsingStep[]
  currentFile?: string
}

export default function ParsingProgress({ steps, currentFile }: ParsingProgressProps) {
  const getStatusIcon = (status: ParsingStep['status']) => {
    switch (status) {
      case 'completed':
        return '✓'
      case 'in_progress':
        return '⚡'
      case 'error':
        return '✗'
      default:
        return '○'
    }
  }

  const getStatusColor = (status: ParsingStep['status']) => {
    switch (status) {
      case 'completed':
        return 'text-mint-400 border-mint-500/50 bg-mint-500/10'
      case 'in_progress':
        return 'text-electric-400 border-electric-500/50 bg-electric-500/10 animate-pulse'
      case 'error':
        return 'text-sunset-400 border-sunset-500/50 bg-sunset-500/10'
      default:
        return 'text-gray-500 border-gray-700 bg-gray-800/30'
    }
  }

  return (
    <div className="flex flex-col h-full p-8">
      <div className="max-w-3xl mx-auto w-full">
        {/* Header */}
        <div className="text-center mb-8">
          <div className="text-6xl mb-4 animate-bounce">🤖</div>
          <h2 className="text-2xl font-bold text-gray-200 mb-2">Parsing Your Script</h2>
          {currentFile && (
            <p className="text-gray-400">{currentFile}</p>
          )}
        </div>

        {/* Progress Steps */}
        <div className="space-y-4">
          {steps.map((step, index) => (
            <div
              key={step.id}
              className={`
                border rounded-xl p-4 transition-all duration-300
                ${getStatusColor(step.status)}
              `}
            >
              <div className="flex items-center justify-between mb-2">
                <div className="flex items-center gap-3">
                  <span className="text-2xl">{getStatusIcon(step.status)}</span>
                  <div>
                    <h3 className="font-semibold text-gray-200">{step.name}</h3>
                    {step.details && (
                      <p className="text-xs text-gray-400 mt-1">{step.details}</p>
                    )}
                  </div>
                </div>
                {step.progress !== undefined && (
                  <span className="text-sm font-mono">{step.progress}%</span>
                )}
              </div>

              {step.status === 'in_progress' && step.progress !== undefined && (
                <div className="h-2 bg-gray-700 rounded-full overflow-hidden">
                  <div
                    className="h-full bg-gradient-to-r from-electric-500 to-sakura-500 transition-all duration-500"
                    style={{ width: `${step.progress}%` }}
                  />
                </div>
              )}
            </div>
          ))}
        </div>

        {/* Overall Progress */}
        <div className="mt-8 p-4 bg-gray-800/50 border border-gray-700 rounded-xl">
          <div className="flex items-center justify-between text-sm text-gray-400">
            <span>Overall Progress</span>
            <span>
              {steps.filter(s => s.status === 'completed').length} / {steps.length} steps completed
            </span>
          </div>
        </div>
      </div>
    </div>
  )
}
