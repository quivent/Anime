import { useState, useEffect, useRef } from 'react'
import { invoke } from '@tauri-apps/api/core'
import type { Instance, InstanceType, SSHKey } from '../types/lambda'
import { useInstanceStore } from '../store/instanceStore'
import ServerMonitor from './ServerMonitor'

export default function LambdaView() {
  const { selectedInstance, setSelectedInstance } = useInstanceStore()
  const [apiKeySet, setApiKeySet] = useState(false)
  const [showApiKeyDialog, setShowApiKeyDialog] = useState(false)
  const [apiKey, setApiKey] = useState('')
  const [instances, setInstances] = useState<Instance[]>([])
  const [instanceTypes, setInstanceTypes] = useState<InstanceType[]>([])
  const [sshKeys, setSSHKeys] = useState<SSHKey[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [showLaunchDialog, setShowLaunchDialog] = useState(false)
  const [monitoringInstance, setMonitoringInstance] = useState<Instance | null>(null)
  const [terminatingId, setTerminatingId] = useState<string | null>(null)
  const [restartingId, setRestartingId] = useState<string | null>(null)
  const [successMessage, setSuccessMessage] = useState<string | null>(null)
  const [instanceToTerminate, setInstanceToTerminate] = useState<string | null>(null)
  const [showFinalConfirm, setShowFinalConfirm] = useState(false)
  const [holdProgress, setHoldProgress] = useState(0)
  const holdTimerRef = useRef<number | null>(null)
  const [terminateProgress, setTerminateProgress] = useState<string | null>(null)
  const pollingIntervalRef = useRef<number | null>(null)
  const [showConfigModal, setShowConfigModal] = useState(false)
  const [configInstance, setConfigInstance] = useState<Instance | null>(null)

  useEffect(() => {
    checkConnection()
  }, [])

  // Handle Escape key to close config modal
  useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && showConfigModal) {
        setShowConfigModal(false)
        setConfigInstance(null)
      }
    }

    if (showConfigModal) {
      window.addEventListener('keydown', handleEscape)
    }

    return () => {
      window.removeEventListener('keydown', handleEscape)
    }
  }, [showConfigModal])

  // Auto-refresh instances every 10 seconds to keep status updated
  useEffect(() => {
    if (apiKeySet && instances.length > 0) {
      pollingIntervalRef.current = window.setInterval(async () => {
        try {
          const instancesData = await invoke<Instance[]>('lambda_list_instances')
          setInstances(instancesData)
        } catch (err) {
          console.error('[LambdaView] Error polling instances:', err)
        }
      }, 10000)
    }

    return () => {
      if (pollingIntervalRef.current) {
        clearInterval(pollingIntervalRef.current)
      }
    }
  }, [apiKeySet, instances.length])

  const checkConnection = async () => {
    try {
      // First try to load API key from persistent storage
      const loaded = await invoke<boolean>('load_lambda_api_key')

      // Then check if we have a connection
      const connected = await invoke<boolean>('check_lambda_connection')
      setApiKeySet(connected)

      if (connected) {
        await loadData()
      } else {
        setShowApiKeyDialog(true)
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    } finally {
      setLoading(false)
    }
  }

  const loadData = async () => {
    setLoading(true)
    try {
      const [instancesData, typesData, keysData] = await Promise.all([
        invoke<Instance[]>('lambda_list_instances'),
        invoke<InstanceType[]>('lambda_list_instance_types'),
        invoke<SSHKey[]>('lambda_list_ssh_keys'),
      ])

      setInstances(instancesData)
      setInstanceTypes(typesData)
      setSSHKeys(keysData)
      setError(null)
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    } finally {
      setLoading(false)
    }
  }

  const handleSaveApiKey = async () => {
    console.log('[LambdaView] handleSaveApiKey called with key length:', apiKey.length)
    try {
      console.log('[LambdaView] Invoking set_lambda_api_key...')
      const result = await invoke('set_lambda_api_key', { apiKey })
      console.log('[LambdaView] set_lambda_api_key result:', result)
      setApiKeySet(true)
      setShowApiKeyDialog(false)
      console.log('[LambdaView] Loading data after API key set...')
      await loadData()
    } catch (err) {
      console.error('[LambdaView] Error setting API key:', err)
      setError(err instanceof Error ? err.message : String(err))
    }
  }

  const handleHoldStart = () => {
    setHoldProgress(0)
    const startTime = Date.now()
    const duration = 2000 // 2 seconds hold

    const updateProgress = () => {
      const elapsed = Date.now() - startTime
      const progress = Math.min((elapsed / duration) * 100, 100)
      setHoldProgress(progress)

      if (progress < 100) {
        holdTimerRef.current = window.requestAnimationFrame(updateProgress)
      } else {
        // Hold complete, trigger termination
        if (instanceToTerminate) {
          confirmTerminate(instanceToTerminate)
        }
      }
    }

    holdTimerRef.current = window.requestAnimationFrame(updateProgress)
  }

  const handleHoldEnd = () => {
    if (holdTimerRef.current) {
      cancelAnimationFrame(holdTimerRef.current)
      holdTimerRef.current = null
    }
    setHoldProgress(0)
  }

  const confirmTerminate = async (instanceId: string) => {
    console.log('[LambdaView] Confirming termination for instance:', instanceId)
    setTerminatingId(instanceId)
    setError(null)
    setSuccessMessage(null)
    setInstanceToTerminate(null)
    setShowFinalConfirm(false)
    setTerminateProgress('Sending termination request...')

    try {
      console.log('[LambdaView] Calling lambda_terminate_instances...')
      const result = await invoke<string[]>('lambda_terminate_instances', { instanceIds: [instanceId] })
      console.log('[LambdaView] Terminate API call successful, result:', result)

      setTerminateProgress('Instance terminating successfully')
      setSuccessMessage(`✓ Successfully terminated instance ${instanceId}`)

      // Optimistically update the instance status to 'terminating' in the local state
      setInstances(prevInstances =>
        prevInstances.map(inst =>
          inst.id === instanceId
            ? { ...inst, status: 'terminating' as any }
            : inst
        )
      )

      setTerminateProgress(null)

      // Clear success message after 3 seconds
      setTimeout(() => setSuccessMessage(null), 3000)
    } catch (err) {
      console.error('[LambdaView] Error terminating instance:', err)
      setError(`Failed to terminate: ${err instanceof Error ? err.message : String(err)}`)
      setTerminateProgress(null)
    } finally {
      console.log('[LambdaView] Clearing terminatingId')
      setTerminatingId(null)
    }
  }

  const handleRestartInstance = async (instanceId: string) => {
    setRestartingId(instanceId)
    setError(null)
    setSuccessMessage(null)

    try {
      await invoke('lambda_restart_instances', { instanceIds: [instanceId] })
      setSuccessMessage(`✓ Successfully restarted instance ${instanceId}`)
      await loadData()

      // Clear success message after 3 seconds
      setTimeout(() => setSuccessMessage(null), 3000)
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    } finally {
      setRestartingId(null)
    }
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active': return 'text-mint-400 bg-mint-500/10 border-mint-500/30'
      case 'booting': return 'text-electric-400 bg-electric-500/10 border-electric-500/30'
      case 'unhealthy': return 'text-sunset-400 bg-sunset-500/10 border-sunset-500/30'
      case 'terminated': return 'text-gray-400 bg-gray-500/10 border-gray-500/30'
      case 'terminating': return 'text-red-400 bg-red-500/10 border-red-500/30'
      default: return 'text-gray-400 bg-gray-500/10 border-gray-500/30'
    }
  }

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'active': return '●'
      case 'booting': return '◐'
      case 'unhealthy': return '⚠'
      case 'terminated': return '○'
      case 'terminating': return '⏳'
      default: return '?'
    }
  }

  if (showApiKeyDialog) {
    return (
      <div className="flex-1 flex items-center justify-center p-8">
        <div className="bg-gray-900/80 backdrop-blur-md border border-electric-500/30 rounded-xl p-8 max-w-md w-full anime-glow">
          <div className="text-center mb-6">
            <div className="text-6xl mb-4">☁️</div>
            <h2 className="text-2xl font-bold text-electric-400 mb-2">Lambda Cloud API Key</h2>
            <p className="text-gray-400 text-sm">
              Enter your Lambda API key to get started. You can find this in your Lambda Cloud dashboard.
            </p>
          </div>

          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-300 mb-2">API Key</label>
              <input
                type="password"
                value={apiKey}
                onChange={(e) => setApiKey(e.target.value)}
                placeholder="Enter your Lambda API key"
                className="w-full px-4 py-3 bg-gray-800/50 border border-gray-700 rounded-lg focus:outline-none focus:border-electric-500 text-white"
                onKeyDown={(e) => e.key === 'Enter' && handleSaveApiKey()}
              />
            </div>

            {error && (
              <div className="p-3 bg-sunset-500/10 border border-sunset-500/30 rounded-lg text-sunset-400 text-sm">
                {error}
              </div>
            )}

            <button
              onClick={handleSaveApiKey}
              disabled={!apiKey.trim()}
              className="w-full px-6 py-3 bg-electric-500/20 hover:bg-electric-500/30 border border-electric-500/50 rounded-lg text-electric-400 font-medium transition-all disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Connect to Lambda
            </button>
          </div>
        </div>
      </div>
    )
  }

  if (loading) {
    return (
      <div className="flex-1 flex items-center justify-center">
        <div className="text-center">
          <div className="text-6xl mb-4 animate-pulse">☁️</div>
          <div className="text-electric-400 font-medium">Loading Lambda data...</div>
        </div>
      </div>
    )
  }

  return (
    <div className="flex-1 overflow-auto p-8">
      <div className="max-w-7xl mx-auto space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold text-electric-400 flex items-center gap-3">
              <span className="text-4xl">☁️</span>
              Lambda Cloud
            </h1>
            <p className="text-gray-400 mt-1">Manage your GPU instances</p>
          </div>

          <div className="flex gap-3">
            <button
              onClick={() => loadData()}
              className="px-4 py-2 bg-gray-800/50 hover:bg-gray-700/50 border border-gray-700 rounded-lg text-gray-300 transition-all"
            >
              🔄 Refresh
            </button>
            <button
              onClick={() => setShowLaunchDialog(true)}
              className="px-6 py-2 bg-electric-500/20 hover:bg-electric-500/30 border border-electric-500/50 rounded-lg text-electric-400 font-medium transition-all anime-glow"
            >
              + Launch Instance
            </button>
          </div>
        </div>

        {error && (
          <div className="p-4 bg-sunset-500/10 border border-sunset-500/30 rounded-lg text-sunset-400">
            {error}
          </div>
        )}

        {successMessage && (
          <div className="p-4 bg-mint-500/10 border border-mint-500/30 rounded-lg text-mint-400">
            {successMessage}
          </div>
        )}

        {terminateProgress && (
          <div className="p-4 bg-red-500/10 border border-red-500/30 rounded-lg">
            <div className="flex items-center gap-3 mb-2">
              <div className="animate-spin text-red-400">⏳</div>
              <span className="text-red-400 font-medium">{terminateProgress}</span>
            </div>
            <div className="w-full h-2 bg-gray-800 rounded-full overflow-hidden">
              <div className="h-full bg-red-500 animate-pulse" style={{ width: '100%' }} />
            </div>
          </div>
        )}

        {/* Instances Grid */}
        <div>
          <h2 className="text-xl font-semibold text-gray-200 mb-4">Running Instances ({instances.length})</h2>

          {instances.length === 0 ? (
            <div className="bg-gray-900/50 border border-gray-800 rounded-xl p-12 text-center">
              <div className="text-6xl mb-4">🚀</div>
              <h3 className="text-xl font-semibold text-gray-300 mb-2">No instances running</h3>
              <p className="text-gray-500 mb-6">Launch your first GPU instance to get started</p>
              <button
                onClick={() => setShowLaunchDialog(true)}
                className="px-6 py-3 bg-electric-500/20 hover:bg-electric-500/30 border border-electric-500/50 rounded-lg text-electric-400 font-medium transition-all"
              >
                Launch Instance
              </button>
            </div>
          ) : (
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
              {instances.map((instance) => {
                const isSelected = selectedInstance?.id === instance.id
                return (
                <div
                  key={instance.id}
                  onClick={() => {
                    setConfigInstance(instance)
                    setShowConfigModal(true)
                  }}
                  className={`bg-gray-900/50 border rounded-xl p-6 transition-all cursor-pointer ${
                    isSelected
                      ? 'border-sakura-500 bg-sakura-500/10 anime-glow'
                      : 'border-gray-800 hover:border-electric-500/30 hover:bg-gray-800/70'
                  }`}
                >
                  <div className="flex items-start justify-between mb-4">
                    <div>
                      <h3 className="text-lg font-semibold text-gray-200">
                        {instance.name || instance.id}
                      </h3>
                      <div className="text-sm text-gray-400 mt-1">
                        {instance.instance_type.name}
                      </div>
                    </div>

                    <div className={`px-3 py-1 rounded-full text-xs font-medium border flex items-center gap-2 ${getStatusColor(instance.status)}`}>
                      <span>{getStatusIcon(instance.status)}</span>
                      <span className="capitalize">{instance.status}</span>
                    </div>
                  </div>

                  <div className="space-y-2 mb-4">
                    {instance.ip && (
                      <div className="flex items-center gap-2 text-sm">
                        <span className="text-gray-500">IP:</span>
                        <code className="px-2 py-1 bg-gray-800 rounded text-mint-400 font-mono">
                          {instance.ip}
                        </code>
                      </div>
                    )}

                    {instance.region && (
                      <div className="flex items-center gap-2 text-sm">
                        <span className="text-gray-500">Region:</span>
                        <span className="text-gray-300">{instance.region.name}</span>
                      </div>
                    )}

                    {instance.ssh_key_names.length > 0 && (
                      <div className="flex items-center gap-2 text-sm">
                        <span className="text-gray-500">SSH Keys:</span>
                        <span className="text-gray-300">{instance.ssh_key_names.join(', ')}</span>
                      </div>
                    )}
                  </div>

                  <div className="flex gap-2 flex-wrap" onClick={(e) => e.stopPropagation()}>
                    {instance.status === 'active' && instance.ip && (
                      <>
                        <button
                          onClick={(e) => {
                            e.stopPropagation()
                            setMonitoringInstance(instance)
                          }}
                          className="flex-1 min-w-[120px] px-3 py-2 bg-electric-500/20 hover:bg-electric-500/30 border border-electric-500/50 rounded-lg text-electric-400 text-sm font-medium transition-all anime-glow"
                        >
                          📊 Monitor
                        </button>
                        <button
                          onClick={(e) => {
                            e.stopPropagation()
                            navigator.clipboard.writeText(`ssh ubuntu@${instance.ip}`)
                          }}
                          className="flex-1 min-w-[120px] px-3 py-2 bg-mint-500/20 hover:bg-mint-500/30 border border-mint-500/50 rounded-lg text-mint-400 text-sm font-medium transition-all"
                        >
                          📋 Copy SSH
                        </button>
                      </>
                    )}

                    <button
                      onClick={(e) => {
                        e.stopPropagation()
                        handleRestartInstance(instance.id)
                      }}
                      disabled={instance.status !== 'active' || restartingId === instance.id}
                      className="px-3 py-2 bg-gray-800/50 hover:bg-gray-700/50 border border-gray-700 rounded-lg text-gray-300 text-sm transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                      {restartingId === instance.id ? '⏳ Restarting...' : '🔄 Restart'}
                    </button>

                    <button
                      onClick={(e) => {
                        e.stopPropagation()
                        setInstanceToTerminate(instance.id)
                      }}
                      disabled={terminatingId === instance.id}
                      className="px-3 py-2 bg-sunset-500/20 hover:bg-sunset-500/30 border border-sunset-500/50 rounded-lg text-sunset-400 text-sm font-medium transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                      {terminatingId === instance.id ? '⏳ Terminating...' : '⚠ Terminate'}
                    </button>

                    {instance.status === 'active' && (
                      <button
                        onClick={(e) => {
                          e.stopPropagation()
                          setSelectedInstance(isSelected ? null : instance)
                        }}
                        className={`w-full mt-2 px-3 py-2 border rounded-lg text-sm font-medium transition-all ${
                          isSelected
                            ? 'bg-sakura-500/30 border-sakura-500 text-sakura-300 anime-glow'
                            : 'bg-gray-800/50 hover:bg-sakura-500/20 border-gray-700 hover:border-sakura-500/50 text-gray-300 hover:text-sakura-400'
                        }`}
                      >
                        {isSelected ? '✓ Selected for Packages' : '📦 Select for Packages'}
                      </button>
                    )}
                  </div>
                </div>
              )})}
            </div>
          )}
        </div>

        {/* Instance Types */}
        {showLaunchDialog && (
          <LaunchInstanceDialog
            instanceTypes={instanceTypes}
            sshKeys={sshKeys}
            onClose={() => setShowLaunchDialog(false)}
            onLaunch={async () => {
              setShowLaunchDialog(false)
              await loadData()
            }}
          />
        )}

        {/* Server Monitor */}
        {monitoringInstance && (
          <ServerMonitor
            instance={monitoringInstance}
            onClose={() => setMonitoringInstance(null)}
          />
        )}

        {/* Instance Configuration Modal */}
        {showConfigModal && configInstance && (
          <div
            className="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50 p-4"
            onClick={() => {
              setShowConfigModal(false)
              setConfigInstance(null)
            }}
          >
            <div
              className="bg-gray-900 border border-electric-500/30 rounded-xl max-w-3xl w-full anime-glow"
              onClick={(e) => e.stopPropagation()}
            >
              <div className="p-6 border-b border-gray-800 flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <span className="text-3xl">⚙️</span>
                  <div>
                    <h2 className="text-2xl font-bold text-electric-400">Instance Configuration</h2>
                    <p className="text-sm text-gray-400 mt-1">{configInstance.hostname || configInstance.id}</p>
                  </div>
                </div>
                <button
                  onClick={() => setShowConfigModal(false)}
                  className="p-2 rounded-lg hover:bg-gray-800 transition-colors text-gray-400 hover:text-gray-200"
                >
                  ✕
                </button>
              </div>

              <div className="p-6 space-y-6 max-h-[70vh] overflow-y-auto">
                {/* Instance Details */}
                <div className="space-y-4">
                  <h3 className="text-lg font-semibold text-gray-200 flex items-center gap-2">
                    <span>📋</span>
                    Instance Details
                  </h3>
                  <div className="grid grid-cols-2 gap-4">
                    <div className="p-4 bg-gray-800/50 rounded-lg">
                      <div className="text-xs text-gray-500 mb-1">Instance ID</div>
                      <div className="font-mono text-sm text-gray-300">{configInstance.id}</div>
                    </div>
                    <div className="p-4 bg-gray-800/50 rounded-lg">
                      <div className="text-xs text-gray-500 mb-1">Status</div>
                      <div className="text-sm capitalize">{configInstance.status}</div>
                    </div>
                    <div className="p-4 bg-gray-800/50 rounded-lg">
                      <div className="text-xs text-gray-500 mb-1">Public IP</div>
                      <div className="font-mono text-sm text-mint-400">{configInstance.ip || 'N/A'}</div>
                    </div>
                    <div className="p-4 bg-gray-800/50 rounded-lg">
                      <div className="text-xs text-gray-500 mb-1">Private IP</div>
                      <div className="font-mono text-sm text-gray-400">{configInstance.private_ip || 'N/A'}</div>
                    </div>
                  </div>
                </div>

                {/* Hardware */}
                <div className="space-y-4">
                  <h3 className="text-lg font-semibold text-gray-200 flex items-center gap-2">
                    <span>🖥️</span>
                    Hardware
                  </h3>
                  <div className="grid grid-cols-2 gap-4">
                    <div className="p-4 bg-gray-800/50 rounded-lg">
                      <div className="text-xs text-gray-500 mb-1">Instance Type</div>
                      <div className="text-sm font-medium text-gray-300">{configInstance.instance_type.name}</div>
                      <div className="text-xs text-gray-500 mt-1">{configInstance.instance_type.description}</div>
                    </div>
                    <div className="p-4 bg-gray-800/50 rounded-lg">
                      <div className="text-xs text-gray-500 mb-1">Region</div>
                      <div className="text-sm font-medium text-gray-300">{configInstance.region.name}</div>
                      <div className="text-xs text-gray-500 mt-1">{configInstance.region.description}</div>
                    </div>
                  </div>
                </div>

                {/* SSH Access */}
                <div className="space-y-4">
                  <h3 className="text-lg font-semibold text-gray-200 flex items-center gap-2">
                    <span>🔑</span>
                    SSH Access
                  </h3>
                  <div className="space-y-2">
                    {configInstance.ssh_key_names.length > 0 ? (
                      configInstance.ssh_key_names.map((key) => (
                        <div key={key} className="p-3 bg-gray-800/50 rounded-lg flex items-center justify-between">
                          <div className="flex items-center gap-2">
                            <span className="text-mint-400">🔐</span>
                            <span className="text-sm text-gray-300">{key}</span>
                          </div>
                        </div>
                      ))
                    ) : (
                      <div className="p-3 bg-gray-800/50 rounded-lg text-sm text-gray-500">
                        No SSH keys attached
                      </div>
                    )}
                  </div>
                  {configInstance.ip && (
                    <div className="p-4 bg-electric-500/10 border border-electric-500/30 rounded-lg">
                      <div className="text-xs text-electric-400 mb-2">SSH Command</div>
                      <div className="font-mono text-sm text-gray-300 flex items-center justify-between">
                        <code>ssh ubuntu@{configInstance.ip}</code>
                        <button
                          onClick={() => navigator.clipboard.writeText(`ssh ubuntu@${configInstance.ip}`)}
                          className="ml-2 px-2 py-1 bg-electric-500/20 hover:bg-electric-500/30 rounded text-xs text-electric-400"
                        >
                          📋 Copy
                        </button>
                      </div>
                    </div>
                  )}
                </div>

                {/* Jupyter (if available) */}
                {configInstance.jupyter_url && (
                  <div className="space-y-4">
                    <h3 className="text-lg font-semibold text-gray-200 flex items-center gap-2">
                      <span>📓</span>
                      Jupyter Notebook
                    </h3>
                    <div className="p-4 bg-neon-500/10 border border-neon-500/30 rounded-lg space-y-2">
                      <div>
                        <div className="text-xs text-neon-400 mb-1">URL</div>
                        <a
                          href={configInstance.jupyter_url}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="text-sm text-neon-400 hover:text-neon-300 underline"
                        >
                          {configInstance.jupyter_url}
                        </a>
                      </div>
                      <div>
                        <div className="text-xs text-neon-400 mb-1">Token</div>
                        <div className="font-mono text-sm text-gray-300 flex items-center justify-between">
                          <code className="truncate">{configInstance.jupyter_token}</code>
                          <button
                            onClick={() => navigator.clipboard.writeText(configInstance.jupyter_token || '')}
                            className="ml-2 px-2 py-1 bg-neon-500/20 hover:bg-neon-500/30 rounded text-xs text-neon-400"
                          >
                            📋 Copy
                          </button>
                        </div>
                      </div>
                    </div>
                  </div>
                )}
              </div>

              <div className="p-6 border-t border-gray-800 flex gap-3">
                <button
                  onClick={() => setShowConfigModal(false)}
                  className="flex-1 px-6 py-3 bg-gray-800/50 hover:bg-gray-700/50 border border-gray-700 rounded-lg text-gray-300 transition-all"
                >
                  Close
                </button>
              </div>
            </div>
          </div>
        )}

        {/* Terminate Confirmation Dialog */}
        {instanceToTerminate && (
          <div className="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50 p-4">
            <div className="bg-gray-900 border border-sunset-500/30 rounded-xl max-w-2xl w-full anime-glow">
              <div className="p-6 border-b border-gray-800">
                <h2 className="text-2xl font-bold text-sunset-400">⚠ Confirm Termination</h2>
              </div>

              <div className="p-6">
                {!showFinalConfirm ? (
                  <>
                    <p className="text-gray-300 mb-4">
                      Are you sure you want to terminate this instance?
                    </p>
                    <div className="p-4 bg-sunset-500/10 border border-sunset-500/30 rounded-lg mb-4">
                      <p className="text-sm text-sunset-400 font-mono">{instanceToTerminate}</p>
                    </div>
                    <p className="text-gray-400 text-sm">
                      This action cannot be undone. All data on this instance will be permanently lost.
                    </p>
                  </>
                ) : (
                  <>
                    <div className="p-6 bg-red-500/10 border-2 border-red-500/50 rounded-lg mb-4">
                      <p className="text-red-400 font-bold text-lg mb-2">⛔ FINAL WARNING</p>
                      <p className="text-red-300 mb-3">
                        You are about to permanently destroy this instance. This action is IRREVERSIBLE.
                      </p>
                      <div className="p-3 bg-black/30 border border-red-500/30 rounded">
                        <p className="text-red-400 font-mono text-sm">{instanceToTerminate}</p>
                      </div>
                    </div>
                    <p className="text-gray-400 text-sm mb-4">
                      To confirm, press and hold the button below for 2 seconds.
                    </p>
                  </>
                )}
              </div>

              <div className="p-6 border-t border-gray-800 flex gap-3">
                <button
                  onClick={() => {
                    setInstanceToTerminate(null)
                    setShowFinalConfirm(false)
                    handleHoldEnd()
                  }}
                  className="flex-1 px-6 py-3 bg-gray-800/50 hover:bg-gray-700/50 border border-gray-700 rounded-lg text-gray-300 transition-all"
                >
                  Cancel
                </button>
                {!showFinalConfirm ? (
                  <button
                    onClick={() => setShowFinalConfirm(true)}
                    className="flex-1 px-6 py-3 bg-sunset-500/20 hover:bg-sunset-500/30 border border-sunset-500/50 rounded-lg text-sunset-400 font-medium transition-all"
                  >
                    ⚠ Continue to Terminate
                  </button>
                ) : (
                  <button
                    onMouseDown={handleHoldStart}
                    onMouseUp={handleHoldEnd}
                    onMouseLeave={handleHoldEnd}
                    onTouchStart={handleHoldStart}
                    onTouchEnd={handleHoldEnd}
                    disabled={terminatingId === instanceToTerminate}
                    className="flex-1 px-6 py-3 bg-red-500/20 hover:bg-red-500/30 border-2 border-red-500/50 rounded-lg text-red-400 font-bold transition-all relative overflow-hidden disabled:opacity-50"
                  >
                    <div
                      className="absolute inset-0 bg-red-500/30 transition-all"
                      style={{ width: `${holdProgress}%` }}
                    />
                    <span className="relative z-10">
                      {terminatingId === instanceToTerminate ? '⏳ Terminating...' : holdProgress > 0 ? `Hold... ${Math.round(holdProgress)}%` : '⛔ HOLD TO DESTROY'}
                    </span>
                  </button>
                )}
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

interface LaunchInstanceDialogProps {
  instanceTypes: InstanceType[]
  sshKeys: SSHKey[]
  onClose: () => void
  onLaunch: () => void
}

function LaunchInstanceDialog({ instanceTypes, sshKeys, onClose, onLaunch }: LaunchInstanceDialogProps) {
  // Generate creative instance name (three words, no numbers)
  const generateInstanceName = () => {
    const adjectives = ['swift', 'quantum', 'neural', 'cosmic', 'lightning', 'turbo', 'hyper', 'mega', 'ultra', 'prime', 'stellar', 'blazing', 'crimson', 'azure', 'golden', 'silver', 'shadow', 'radiant', 'mystic', 'arcane']
    const modifiers = ['flux', 'pulse', 'wave', 'storm', 'drift', 'surge', 'echo', 'spark', 'void', 'dawn', 'dusk', 'edge', 'peak', 'flow', 'stream', 'tide', 'wind', 'fire', 'frost', 'ember']
    const nouns = ['tensor', 'nexus', 'matrix', 'forge', 'engine', 'reactor', 'core', 'vault', 'node', 'cluster', 'phoenix', 'dragon', 'titan', 'sentinel', 'guardian', 'oracle', 'spectre', 'wraith', 'prism', 'beacon']
    const adj = adjectives[Math.floor(Math.random() * adjectives.length)]
    const mod = modifiers[Math.floor(Math.random() * modifiers.length)]
    const noun = nouns[Math.floor(Math.random() * nouns.length)]
    return `${adj}-${mod}-${noun}`
  }

  // Custom sorting: GH200 > H100 > B200 > A100 > rest
  const sortInstanceTypes = (types: InstanceType[]) => {
    const priority: { [key: string]: number } = {
      'gh200': 1,
      'h100': 2,
      'b200': 3,
      'a100': 4,
    }

    return [...types].sort((a, b) => {
      // Get priority based on GPU type
      const getPriority = (name: string) => {
        const lower = name.toLowerCase()
        for (const [key, value] of Object.entries(priority)) {
          if (lower.includes(key)) return value
        }
        return 999
      }

      const aPriority = getPriority(a.name)
      const bPriority = getPriority(b.name)

      if (aPriority !== bPriority) {
        return aPriority - bPriority
      }

      // If same priority, sort by price
      return a.price_cents_per_hour - b.price_cents_per_hour
    })
  }

  const sortedInstanceTypes = sortInstanceTypes(instanceTypes.filter(type => type.regions_with_capacity_available.length > 0))

  const [selectedType, setSelectedType] = useState('')
  const [region, setRegion] = useState('')
  const [selectedKeys, setSelectedKeys] = useState<string[]>([])
  const [instanceName, setInstanceName] = useState(generateInstanceName())
  const [quantity, setQuantity] = useState(1)
  const [launching, setLaunching] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState<string | null>(null)
  const launchingRef = useRef(false)
  const timeoutRef = useRef<number | null>(null)

  // Auto-select first available instance, region, and SSH key
  useEffect(() => {
    if (sortedInstanceTypes.length > 0 && !selectedType) {
      const firstType = sortedInstanceTypes[0]
      setSelectedType(firstType.name)
      if (firstType.regions_with_capacity_available.length > 0) {
        setRegion(firstType.regions_with_capacity_available[0].name)
      }
    }
  }, [sortedInstanceTypes, selectedType])

  useEffect(() => {
    if (sshKeys.length > 0 && selectedKeys.length === 0) {
      setSelectedKeys([sshKeys[0].name])
    }
  }, [sshKeys, selectedKeys.length])

  const handleLaunch = async () => {
    // Prevent duplicate launches
    if (launchingRef.current) {
      console.log('[LaunchInstanceDialog] Launch already in progress, ignoring duplicate call')
      return
    }

    if (!selectedType || selectedKeys.length === 0) {
      setError('Please select instance type and at least one SSH key')
      return
    }

    if (!region) {
      setError('Please select a region')
      return
    }

    launchingRef.current = true
    setLaunching(true)
    setError(null)
    setSuccess(null)

    try {
      const result = await invoke<string[]>('lambda_launch_instance', {
        instanceType: selectedType,
        region,
        sshKeys: selectedKeys,
        name: instanceName || null,
        quantity,
      })

      setSuccess(`✓ Successfully launched ${result.length} instance${result.length > 1 ? 's' : ''}! Instance IDs: ${result.join(', ')}`)
      setLaunching(false)

      // Close dialog and refresh after 2 seconds to show success message
      timeoutRef.current = window.setTimeout(() => {
        onLaunch()
      }, 2000)
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : String(err)
      setError(`Failed to launch instance: ${errorMsg}`)
      setLaunching(false)
      launchingRef.current = false
    }
  }

  // Cleanup timeout on unmount
  useEffect(() => {
    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current)
      }
    }
  }, [])

  const selectedInstanceType = sortedInstanceTypes.find(t => t.name === selectedType)

  return (
    <div className="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50 p-4">
      <div className="bg-gray-900 border border-electric-500/30 rounded-xl max-w-2xl w-full max-h-[90vh] overflow-auto anime-glow">
        <div className="p-6 border-b border-gray-800">
          <h2 className="text-2xl font-bold text-electric-400">Launch Instance</h2>
        </div>

        <div className="p-6 space-y-6">
          {/* Instance Type */}
          <div>
            {sortedInstanceTypes.length === 0 ? (
              <div className="p-6 bg-sunset-500/10 border border-sunset-500/30 rounded-lg text-center">
                <div className="text-4xl mb-2">⚠️</div>
                <p className="text-sunset-400 font-medium mb-1">No instances available</p>
                <p className="text-gray-400 text-sm">There are currently no GPU instances with available capacity. Please try again later.</p>
              </div>
            ) : (
              <div className="grid grid-cols-1 gap-3 max-h-64 overflow-y-auto">
                {sortedInstanceTypes.map((type) => (
                <button
                  key={type.name}
                  onClick={() => {
                    setSelectedType(type.name)
                    // Auto-select first available region
                    if (type.regions_with_capacity_available.length > 0) {
                      setRegion(type.regions_with_capacity_available[0].name)
                    }
                  }}
                  className={`p-4 rounded-lg border text-left transition-all ${
                    selectedType === type.name
                      ? 'bg-electric-500/20 border-electric-500/50 text-electric-400 ring-2 ring-electric-500/30'
                      : 'bg-gray-800/50 border-gray-700 text-gray-300 hover:border-electric-500/30'
                  }`}
                >
                  <div className="flex items-center justify-between mb-2">
                    <span className="font-semibold">{type.description}</span>
                    <span className="text-mint-400 font-mono text-sm">
                      ${(type.price_cents_per_hour / 100).toFixed(2)}/hr
                    </span>
                  </div>
                  <div className="text-sm text-gray-400 mb-1">
                    {type.specs.vcpus} vCPUs • {type.specs.memory_gib} GB RAM • {type.specs.storage_gib} GB Storage
                    {type.specs.gpus > 0 && ` • ${type.specs.gpus} GPU${type.specs.gpus > 1 ? 's' : ''}`}
                  </div>
                  <div className="text-xs text-mint-400/70">
                    Available in: {type.regions_with_capacity_available.map(r => r.description).join(', ')}
                  </div>
                </button>
              ))}
              </div>
            )}
          </div>

          {/* SSH Keys */}
          <div>
            <div className="flex flex-wrap gap-2">
              {sshKeys.map((key) => (
                <button
                  key={key.id}
                  type="button"
                  onClick={() => {
                    if (selectedKeys.includes(key.name)) {
                      setSelectedKeys(selectedKeys.filter(k => k !== key.name))
                    } else {
                      setSelectedKeys([...selectedKeys, key.name])
                    }
                  }}
                  className={`px-4 py-2 rounded-lg text-sm font-medium transition-all ${
                    selectedKeys.includes(key.name)
                      ? 'bg-electric-500/20 border-2 border-electric-500/50 text-electric-400'
                      : 'bg-gray-800/50 border border-gray-700 text-gray-300 hover:border-electric-500/30'
                  }`}
                >
                  {key.name}
                </button>
              ))}
            </div>
          </div>
          {/* Instance Name */}
          <div>
            <div className="flex items-center gap-2">
              <div className="flex-1 relative group">
                <div className="absolute -inset-0.5 bg-gradient-to-r from-electric-500 via-mint-400 to-sunset-500 rounded-lg blur opacity-30 group-hover:opacity-50 transition duration-300 animate-pulse"></div>
                <input
                  type="text"
                  value={instanceName}
                  onChange={(e) => setInstanceName(e.target.value)}
                  placeholder="instance-name"
                  className="relative w-full px-5 py-4 bg-gray-900 border-2 border-electric-400/40 rounded-lg focus:outline-none focus:border-electric-400 focus:ring-2 focus:ring-electric-500/50 text-transparent bg-clip-text bg-gradient-to-r from-electric-400 via-mint-400 to-electric-400 font-mono text-xl font-bold tracking-wide placeholder:text-gray-600 placeholder:bg-clip-text placeholder:bg-gradient-to-r placeholder:from-gray-600 placeholder:to-gray-600"
                  style={{
                    WebkitTextFillColor: 'transparent',
                    WebkitBackgroundClip: 'text',
                    backgroundClip: 'text'
                  }}
                />
              </div>
              <button
                onClick={() => setInstanceName(generateInstanceName())}
                className="px-5 py-4 bg-gradient-to-r from-electric-500/20 to-mint-500/20 hover:from-electric-500/30 hover:to-mint-500/30 border-2 border-electric-500/50 rounded-lg text-2xl transition-all hover:scale-110 hover:rotate-12"
              >
                🎲
              </button>
            </div>
          </div>

          {/* Region */}
          {selectedInstanceType && selectedInstanceType.regions_with_capacity_available.length > 0 && (
            <div className={`flex ${selectedInstanceType.regions_with_capacity_available.length === 1 ? 'w-full' : 'flex-wrap'} gap-2`}>
              {selectedInstanceType.regions_with_capacity_available.map((r) => (
                <button
                  key={r.name}
                  type="button"
                  onClick={() => setRegion(r.name)}
                  className={`px-4 py-3 rounded-lg text-sm font-medium transition-all ${
                    selectedInstanceType.regions_with_capacity_available.length === 1
                      ? 'flex-1'
                      : ''
                  } ${
                    region === r.name
                      ? 'bg-electric-500/20 border-2 border-electric-500/50 text-electric-400'
                      : 'bg-gray-800/50 border border-gray-700 text-gray-300 hover:border-electric-500/30'
                  }`}
                >
                  {r.description}
                </button>
              ))}
            </div>
          )}

          {launching && (
            <div className="p-4 bg-electric-500/10 border border-electric-500/30 rounded-lg">
              <div className="flex items-center gap-3 mb-2">
                <div className="animate-spin text-electric-400 text-xl">⚡</div>
                <span className="text-electric-400 font-medium">Launching instance...</span>
              </div>
              <div className="w-full h-2 bg-gray-800 rounded-full overflow-hidden">
                <div className="h-full bg-electric-500 animate-pulse" style={{ width: '100%' }} />
              </div>
            </div>
          )}

          {error && (
            <div className="p-3 bg-sunset-500/10 border border-sunset-500/30 rounded-lg text-sunset-400 text-sm">
              {error}
            </div>
          )}

          {success && (
            <div className="p-3 bg-mint-500/10 border border-mint-500/30 rounded-lg text-mint-400 text-sm">
              {success}
            </div>
          )}

          {/* Summary */}
          {selectedInstanceType && (
            <div className="p-4 bg-electric-500/10 border border-electric-500/30 rounded-lg">
              <div className="text-sm text-gray-300 space-y-1">
                <div className="flex justify-between">
                  <span>Instance Type:</span>
                  <span className="font-semibold text-electric-400">{selectedInstanceType.description}</span>
                </div>
                <div className="flex justify-between">
                  <span>Cost per hour:</span>
                  <span className="font-mono text-mint-400">${(selectedInstanceType.price_cents_per_hour / 100).toFixed(2)}</span>
                </div>
                <div className="flex justify-between">
                  <span>SSH Keys:</span>
                  <span className="text-gray-400">{selectedKeys.length} selected</span>
                </div>
              </div>
            </div>
          )}
        </div>

        <div className="p-6 border-t border-gray-800 flex gap-3">
          <button
            onClick={onClose}
            disabled={launching}
            className="flex-1 px-6 py-3 bg-gray-800/50 hover:bg-gray-700/50 border border-gray-700 rounded-lg text-gray-300 transition-all disabled:opacity-50"
          >
            Cancel
          </button>
          <button
            onClick={handleLaunch}
            disabled={!selectedType || selectedKeys.length === 0 || launching || !!success}
            className="flex-1 px-6 py-3 bg-electric-500/20 hover:bg-electric-500/30 border border-electric-500/50 rounded-lg text-electric-400 font-medium transition-all disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {launching ? '⏳ Launching...' : success ? '✓ Success!' : 'Launch Instance'}
          </button>
        </div>
      </div>
    </div>
  )
}
