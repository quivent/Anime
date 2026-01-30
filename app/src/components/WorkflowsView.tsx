import { useState, useEffect, useRef } from 'react'
import { invoke } from '@tauri-apps/api/core'
import type { ComfyUIWorkflow, WorkflowExecution, ComfyUIStatus, WorkflowExecutionOutput } from '../types/workflow'

type CategoryFilter = 'all' | 'image' | 'video' | 'upscaling' | 'custom'

export default function WorkflowsView() {
  const [workflows, setWorkflows] = useState<ComfyUIWorkflow[]>([])
  const [selectedWorkflow, setSelectedWorkflow] = useState<ComfyUIWorkflow | null>(null)
  const [executions, setExecutions] = useState<WorkflowExecution[]>([])
  const [status, setStatus] = useState<ComfyUIStatus | null>(null)
  const [selectedCategory, setSelectedCategory] = useState<CategoryFilter>('all')
  const [parameters, setParameters] = useState<Record<string, any>>({})
  const [executing, setExecuting] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [showExecutionPanel, setShowExecutionPanel] = useState(false)
  const [connectionChecked, setConnectionChecked] = useState(false)
  const pollingIntervalRef = useRef<number | null>(null)

  useEffect(() => {
    loadWorkflows()
    checkConnection()
  }, [])

  // Poll for status updates
  useEffect(() => {
    if (connectionChecked && status?.connected) {
      pollingIntervalRef.current = window.setInterval(() => {
        checkConnection()
        updateExecutions()
      }, 3000)
    }

    return () => {
      if (pollingIntervalRef.current) {
        clearInterval(pollingIntervalRef.current)
      }
    }
  }, [connectionChecked, status?.connected])

  async function loadWorkflows() {
    try {
      const wfs = await invoke<ComfyUIWorkflow[]>('comfyui_list_workflows')
      setWorkflows(wfs)
    } catch (err) {

      setError(`Failed to load workflows: ${err}`)
    }
  }

  async function checkConnection() {
    try {
      const connected = await invoke<boolean>('comfyui_check_connection')
      if (connected) {
        const st = await invoke<ComfyUIStatus>('comfyui_get_status')
        setStatus(st)
      } else {
        setStatus({
          connected: false,
          queue_remaining: 0,
          queue_running: 0,
          system_stats: undefined
        })
      }
      setConnectionChecked(true)
    } catch (err) {

      setStatus({
        connected: false,
        queue_remaining: 0,
        queue_running: 0,
        system_stats: undefined
      })
      setConnectionChecked(true)
    }
  }

  async function updateExecutions() {
    try {
      const execs = await invoke<WorkflowExecution[]>('comfyui_list_executions')
      setExecutions(execs)

      // Update individual execution statuses
      for (const exec of execs) {
        if (exec.status === 'queued' || exec.status === 'running') {
          try {
            const updated = await invoke<WorkflowExecution>('comfyui_get_execution', {
              executionId: exec.id
            })
            setExecutions(prev => prev.map(e => e.id === updated.id ? updated : e))
          } catch (err) {

          }
        }
      }
    } catch (err) {

    }
  }

  async function executeWorkflow() {
    if (!selectedWorkflow) return

    setExecuting(true)
    setError(null)

    try {
      const execution = await invoke<WorkflowExecution>('comfyui_execute_workflow', {
        workflowId: selectedWorkflow.id,
        parameters
      })

      setExecutions(prev => [execution, ...prev])
      setShowExecutionPanel(true)
      setExecuting(false)

      // Start monitoring
      const monitorInterval = setInterval(async () => {
        try {
          const updated = await invoke<WorkflowExecution>('comfyui_get_execution', {
            executionId: execution.id
          })

          setExecutions(prev => prev.map(e => e.id === updated.id ? updated : e))

          if (updated.status === 'completed' || updated.status === 'failed') {
            clearInterval(monitorInterval)
          }
        } catch (err) {

          clearInterval(monitorInterval)
        }
      }, 1000)
    } catch (err) {

      setError(`Failed to execute: ${err}`)
      setExecuting(false)
    }
  }

  async function cancelExecution(executionId: string) {
    try {
      await invoke('comfyui_cancel_execution', { executionId })
      await updateExecutions()
    } catch (err) {

    }
  }

  async function interruptAll() {
    try {
      await invoke('comfyui_interrupt')
      await checkConnection()
    } catch (err) {

    }
  }

  const filteredWorkflows = selectedCategory === 'all'
    ? workflows
    : workflows.filter(w => w.category === selectedCategory)

  const categories = [
    { id: 'all' as CategoryFilter, label: 'All', icon: '🎨' },
    { id: 'image' as CategoryFilter, label: 'Images', icon: '🖼️' },
    { id: 'video' as CategoryFilter, label: 'Videos', icon: '🎬' },
    { id: 'upscaling' as CategoryFilter, label: 'Upscaling', icon: '⬆️' },
    { id: 'custom' as CategoryFilter, label: 'Custom', icon: '⚙️' },
  ]

  if (!connectionChecked) {
    return (
      <div className="h-full flex items-center justify-center">
        <div className="text-center">
          <div className="text-6xl mb-4 animate-pulse">🎨</div>
          <div className="text-gray-400">Checking ComfyUI connection...</div>
        </div>
      </div>
    )
  }

  if (!status?.connected) {
    return (
      <div className="h-full flex items-center justify-center p-8">
        <div className="bg-gray-900/80 backdrop-blur-md border border-sunset-500/30 rounded-xl p-8 max-w-md w-full">
          <div className="text-center mb-6">
            <div className="text-6xl mb-4">⚠️</div>
            <h2 className="text-2xl font-bold text-sunset-400 mb-2">ComfyUI Not Connected</h2>
            <p className="text-gray-400 text-sm">
              ComfyUI is not running or not accessible at localhost:8188
            </p>
          </div>

          <div className="space-y-4">
            <div className="p-4 bg-gray-800/50 rounded-lg">
              <h3 className="text-sm font-semibold text-gray-300 mb-2">To connect:</h3>
              <ol className="text-xs text-gray-400 space-y-1 list-decimal list-inside">
                <li>Make sure ComfyUI is installed on your server</li>
                <li>Start ComfyUI (usually runs on port 8188)</li>
                <li>Ensure the server is accessible from this machine</li>
              </ol>
            </div>

            <button
              onClick={checkConnection}
              className="w-full px-6 py-3 bg-electric-500/20 hover:bg-electric-500/30 border border-electric-500/50 rounded-lg text-electric-400 font-medium transition-all"
            >
              🔄 Retry Connection
            </button>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="h-full flex">
      {/* Main Panel */}
      <div className="flex-1 flex flex-col p-6 overflow-hidden">
        {/* Header */}
        <div className="mb-6">
          <div className="flex items-center justify-between mb-2">
            <h2 className="text-3xl font-bold sakura-gradient bg-clip-text text-transparent">
              ComfyUI Workflows
            </h2>
            <div className="flex items-center gap-3">
              {status.queue_running > 0 && (
                <button
                  onClick={interruptAll}
                  className="px-4 py-2 bg-sunset-500/20 hover:bg-sunset-500/30 border border-sunset-500/50 rounded-lg text-sunset-400 text-sm font-medium transition-all"
                >
                  ⏸️ Interrupt
                </button>
              )}
              <button
                onClick={() => setShowExecutionPanel(!showExecutionPanel)}
                className={`px-4 py-2 rounded-lg text-sm font-medium transition-all ${
                  showExecutionPanel
                    ? 'bg-electric-500 text-white'
                    : 'bg-gray-800 text-gray-400 hover:bg-gray-700'
                }`}
              >
                📊 Executions ({executions.length})
              </button>
            </div>
          </div>
          <div className="flex items-center gap-4 text-sm">
            <div className={`flex items-center gap-2 px-3 py-1 rounded-full ${
              status.connected ? 'bg-mint-500/10 border border-mint-500/30 text-mint-400' : 'bg-gray-800 text-gray-400'
            }`}>
              <span className={`w-2 h-2 rounded-full ${status.connected ? 'bg-mint-400 animate-pulse' : 'bg-gray-400'}`}></span>
              <span>{status.connected ? 'Connected' : 'Disconnected'}</span>
            </div>
            {status.connected && (
              <>
                <div className="text-gray-400">
                  Queue: {status.queue_running} running, {status.queue_remaining} pending
                </div>
                {status.system_stats?.devices && status.system_stats.devices.length > 0 && (
                  <div className="text-gray-400">
                    GPU: {status.system_stats.devices[0].name}
                    <span className="ml-2 text-electric-400">
                      {((status.system_stats.devices[0].vram_free / status.system_stats.devices[0].vram_total) * 100).toFixed(0)}% free
                    </span>
                  </div>
                )}
              </>
            )}
          </div>
        </div>

        {/* Category Filter */}
        <div className="flex items-center gap-2 mb-6 overflow-x-auto pb-2">
          {categories.map((cat) => (
            <button
              key={cat.id}
              onClick={() => setSelectedCategory(cat.id)}
              className={`px-4 py-2 rounded-lg font-medium text-sm whitespace-nowrap transition-all ${
                selectedCategory === cat.id
                  ? 'bg-sakura-500 text-white'
                  : 'bg-gray-800 text-gray-400 hover:bg-gray-700'
              }`}
            >
              {cat.icon} {cat.label}
            </button>
          ))}
        </div>

        {/* Error Display */}
        {error && (
          <div className="mb-4 p-4 bg-sunset-500/10 border border-sunset-500/30 rounded-lg text-sunset-400">
            {error}
          </div>
        )}

        {/* Workflow Grid */}
        <div className="flex-1 overflow-y-auto">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {filteredWorkflows.map((workflow) => (
              <WorkflowCard
                key={workflow.id}
                workflow={workflow}
                selected={selectedWorkflow?.id === workflow.id}
                onSelect={() => {
                  setSelectedWorkflow(workflow)
                  // Initialize parameters with defaults
                  const defaultParams: Record<string, any> = {}
                  workflow.parameters.forEach(param => {
                    if (param.default_value !== undefined) {
                      defaultParams[param.id] = param.default_value
                    }
                  })
                  setParameters(defaultParams)
                }}
              />
            ))}
          </div>

          {filteredWorkflows.length === 0 && (
            <div className="text-center py-12">
              <p className="text-gray-500 text-lg">No workflows found in this category</p>
            </div>
          )}
        </div>
      </div>

      {/* Execution/Parameter Panel */}
      {(selectedWorkflow || showExecutionPanel) && (
        <div className="w-96 border-l border-gray-800 bg-gray-900/50 overflow-hidden flex flex-col">
          {selectedWorkflow && !showExecutionPanel ? (
            /* Parameter Panel */
            <div className="flex-1 flex flex-col">
              <div className="p-4 border-b border-gray-800">
                <div className="flex items-center justify-between mb-2">
                  <h3 className="font-bold text-lg">{selectedWorkflow.name}</h3>
                  <button
                    onClick={() => setSelectedWorkflow(null)}
                    className="text-gray-400 hover:text-gray-200"
                  >
                    ✕
                  </button>
                </div>
                <p className="text-sm text-gray-400">{selectedWorkflow.description}</p>
              </div>

              <div className="flex-1 overflow-y-auto p-4 space-y-4">
                {selectedWorkflow.parameters.map((param) => (
                  <ParameterInput
                    key={param.id}
                    parameter={param}
                    value={parameters[param.id]}
                    onChange={(value) => setParameters(prev => ({ ...prev, [param.id]: value }))}
                  />
                ))}
              </div>

              <div className="p-4 border-t border-gray-800">
                <button
                  onClick={executeWorkflow}
                  disabled={executing}
                  className="w-full px-6 py-3 bg-electric-500 text-white hover:bg-electric-600 rounded-lg font-medium transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {executing ? '⏳ Executing...' : '▶️ Execute Workflow'}
                </button>
              </div>
            </div>
          ) : (
            /* Executions Panel */
            <div className="flex-1 flex flex-col">
              <div className="p-4 border-b border-gray-800">
                <div className="flex items-center justify-between">
                  <h3 className="font-bold text-lg">Executions</h3>
                  <button
                    onClick={() => setShowExecutionPanel(false)}
                    className="text-gray-400 hover:text-gray-200"
                  >
                    ✕
                  </button>
                </div>
              </div>

              <div className="flex-1 overflow-y-auto p-4 space-y-3">
                {executions.length === 0 ? (
                  <div className="text-center py-8 text-gray-500">
                    No executions yet
                  </div>
                ) : (
                  executions.map((execution) => (
                    <ExecutionCard
                      key={execution.id}
                      execution={execution}
                      onCancel={() => cancelExecution(execution.id)}
                    />
                  ))
                )}
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  )
}

interface WorkflowCardProps {
  workflow: ComfyUIWorkflow
  selected: boolean
  onSelect: () => void
}

function WorkflowCard({ workflow, selected, onSelect }: WorkflowCardProps) {
  return (
    <div
      onClick={onSelect}
      className={`p-4 rounded-lg border-2 cursor-pointer transition-all ${
        selected
          ? 'border-sakura-500 bg-sakura-500/10 anime-glow'
          : 'border-gray-800 bg-gray-900/50 hover:border-gray-700'
      }`}
    >
      <div className="flex items-start justify-between mb-3">
        <div className="flex items-center gap-3">
          <div className="text-3xl">{workflow.icon}</div>
          <div>
            <h3 className="font-bold text-lg">{workflow.name}</h3>
            <p className="text-xs text-gray-400 line-clamp-2">{workflow.description}</p>
          </div>
        </div>
      </div>

      <div className="flex items-center gap-3 text-xs text-gray-500">
        <span className="px-2 py-1 bg-gray-800 rounded capitalize">{workflow.category}</span>
        <span>{workflow.parameters.length} parameters</span>
      </div>
    </div>
  )
}

interface ParameterInputProps {
  parameter: any
  value: any
  onChange: (value: any) => void
}

function ParameterInput({ parameter, value, onChange }: ParameterInputProps) {
  if (parameter.type === 'text') {
    return (
      <div>
        <label className="block text-sm font-medium text-gray-300 mb-2">
          {parameter.name}
          {parameter.required && <span className="text-sunset-400 ml-1">*</span>}
        </label>
        <textarea
          value={value || ''}
          onChange={(e) => onChange(e.target.value)}
          placeholder={parameter.description}
          rows={3}
          className="w-full px-3 py-2 bg-gray-800 border border-gray-700 rounded-lg text-white text-sm focus:outline-none focus:border-electric-500"
        />
      </div>
    )
  }

  if (parameter.type === 'number') {
    return (
      <div>
        <label className="block text-sm font-medium text-gray-300 mb-2">
          {parameter.name}
          {parameter.required && <span className="text-sunset-400 ml-1">*</span>}
        </label>
        <input
          type="number"
          value={value || ''}
          onChange={(e) => onChange(parseFloat(e.target.value))}
          min={parameter.min}
          max={parameter.max}
          placeholder={parameter.description}
          className="w-full px-3 py-2 bg-gray-800 border border-gray-700 rounded-lg text-white text-sm focus:outline-none focus:border-electric-500"
        />
        {(parameter.min !== undefined || parameter.max !== undefined) && (
          <div className="text-xs text-gray-500 mt-1">
            Range: {parameter.min ?? '−∞'} to {parameter.max ?? '∞'}
          </div>
        )}
      </div>
    )
  }

  if (parameter.type === 'select') {
    return (
      <div>
        <label className="block text-sm font-medium text-gray-300 mb-2">
          {parameter.name}
          {parameter.required && <span className="text-sunset-400 ml-1">*</span>}
        </label>
        <select
          value={value || ''}
          onChange={(e) => onChange(e.target.value)}
          className="w-full px-3 py-2 bg-gray-800 border border-gray-700 rounded-lg text-white text-sm focus:outline-none focus:border-electric-500"
        >
          {parameter.options?.map((option: string) => (
            <option key={option} value={option}>
              {option}
            </option>
          ))}
        </select>
      </div>
    )
  }

  if (parameter.type === 'checkbox') {
    return (
      <div className="flex items-center gap-2">
        <input
          type="checkbox"
          checked={value || false}
          onChange={(e) => onChange(e.target.checked)}
          className="w-4 h-4 bg-gray-800 border border-gray-700 rounded"
        />
        <label className="text-sm text-gray-300">{parameter.name}</label>
      </div>
    )
  }

  return null
}

interface ExecutionCardProps {
  execution: WorkflowExecution
  onCancel: () => void
}

function ExecutionCard({ execution, onCancel }: ExecutionCardProps) {
  const statusColors = {
    queued: 'text-gray-400 bg-gray-500/10',
    running: 'text-electric-400 bg-electric-500/10',
    completed: 'text-mint-400 bg-mint-500/10',
    failed: 'text-sunset-400 bg-sunset-500/10',
  }

  const statusIcons = {
    queued: '⏳',
    running: '⚡',
    completed: '✓',
    failed: '✗',
  }

  return (
    <div className="p-3 rounded-lg bg-gray-800/50 border border-gray-700">
      <div className="flex items-start justify-between mb-2">
        <div className="flex-1">
          <div className="flex items-center gap-2 mb-1">
            <span className={`px-2 py-0.5 rounded text-xs font-medium ${statusColors[execution.status as keyof typeof statusColors]}`}>
              {statusIcons[execution.status as keyof typeof statusIcons]} {execution.status}
            </span>
            {execution.queue_position !== undefined && (
              <span className="text-xs text-gray-500">
                #{execution.queue_position}
              </span>
            )}
          </div>
          <div className="text-xs text-gray-400">
            {execution.workflow_id}
          </div>
        </div>
        {(execution.status === 'queued' || execution.status === 'running') && (
          <button
            onClick={onCancel}
            className="text-sunset-400 hover:text-sunset-300 text-xs"
          >
            Cancel
          </button>
        )}
      </div>

      {(execution.status === 'running' || execution.status === 'queued') && (
        <div className="mb-2">
          <div className="w-full bg-gray-700 rounded-full h-1.5 overflow-hidden">
            <div
              className="bg-gradient-to-r from-electric-500 to-mint-500 h-full transition-all duration-300"
              style={{ width: `${execution.progress}%` }}
            />
          </div>
          <div className="text-xs text-gray-500 mt-1 text-right">
            {execution.progress.toFixed(0)}%
          </div>
        </div>
      )}

      {execution.error && (
        <div className="text-xs text-sunset-400 mb-2">
          {execution.error}
        </div>
      )}

      {execution.outputs && execution.outputs.length > 0 && (
        <div className="mt-2 space-y-2">
          {execution.outputs.map((output, idx) => (
            <OutputPreview key={idx} output={output} />
          ))}
        </div>
      )}
    </div>
  )
}

interface OutputPreviewProps {
  output: WorkflowExecutionOutput
}

function OutputPreview({ output }: OutputPreviewProps) {
  return (
    <div className="relative group">
      <img
        src={output.url}
        alt={output.filename}
        className="w-full rounded-lg border border-gray-700"
      />
      <div className="absolute inset-0 bg-black/50 opacity-0 group-hover:opacity-100 transition-opacity rounded-lg flex items-center justify-center">
        <a
          href={output.url}
          target="_blank"
          rel="noopener noreferrer"
          className="px-3 py-1 bg-electric-500 text-white rounded text-sm"
        >
          View Full
        </a>
      </div>
      <div className="text-xs text-gray-500 mt-1 truncate">{output.filename}</div>
    </div>
  )
}
