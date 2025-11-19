import { useState, useEffect } from 'react'
import { invoke } from '@tauri-apps/api/core'
import type { Instance, InstanceType, SSHKey } from '../types/lambda'

export default function LambdaView() {
  const [apiKeySet, setApiKeySet] = useState(false)
  const [showApiKeyDialog, setShowApiKeyDialog] = useState(false)
  const [apiKey, setApiKey] = useState('')
  const [instances, setInstances] = useState<Instance[]>([])
  const [instanceTypes, setInstanceTypes] = useState<InstanceType[]>([])
  const [sshKeys, setSSHKeys] = useState<SSHKey[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [showLaunchDialog, setShowLaunchDialog] = useState(false)

  useEffect(() => {
    checkConnection()
  }, [])

  const checkConnection = async () => {
    try {
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

  const handleTerminateInstance = async (instanceId: string) => {
    if (!confirm('Are you sure you want to terminate this instance?')) return

    try {
      await invoke('lambda_terminate_instances', { instanceIds: [instanceId] })
      await loadData()
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    }
  }

  const handleRestartInstance = async (instanceId: string) => {
    try {
      await invoke('lambda_restart_instances', { instanceIds: [instanceId] })
      await loadData()
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    }
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active': return 'text-mint-400 bg-mint-500/10 border-mint-500/30'
      case 'booting': return 'text-electric-400 bg-electric-500/10 border-electric-500/30'
      case 'unhealthy': return 'text-sunset-400 bg-sunset-500/10 border-sunset-500/30'
      case 'terminated': return 'text-gray-400 bg-gray-500/10 border-gray-500/30'
      default: return 'text-gray-400 bg-gray-500/10 border-gray-500/30'
    }
  }

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'active': return '●'
      case 'booting': return '◐'
      case 'unhealthy': return '⚠'
      case 'terminated': return '○'
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
              {instances.map((instance) => (
                <div
                  key={instance.id}
                  className="bg-gray-900/50 border border-gray-800 rounded-xl p-6 hover:border-electric-500/30 transition-all"
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

                  <div className="flex gap-2">
                    {instance.status === 'active' && instance.ip && (
                      <button
                        onClick={() => navigator.clipboard.writeText(`ssh ubuntu@${instance.ip}`)}
                        className="flex-1 px-3 py-2 bg-mint-500/20 hover:bg-mint-500/30 border border-mint-500/50 rounded-lg text-mint-400 text-sm font-medium transition-all"
                      >
                        📋 Copy SSH
                      </button>
                    )}

                    <button
                      onClick={() => handleRestartInstance(instance.id)}
                      disabled={instance.status !== 'active'}
                      className="px-3 py-2 bg-gray-800/50 hover:bg-gray-700/50 border border-gray-700 rounded-lg text-gray-300 text-sm transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                      🔄 Restart
                    </button>

                    <button
                      onClick={() => handleTerminateInstance(instance.id)}
                      className="px-3 py-2 bg-sunset-500/20 hover:bg-sunset-500/30 border border-sunset-500/50 rounded-lg text-sunset-400 text-sm font-medium transition-all"
                    >
                      ⚠ Terminate
                    </button>
                  </div>
                </div>
              ))}
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
  const [selectedType, setSelectedType] = useState('')
  const [region, setRegion] = useState('')
  const [selectedKeys, setSelectedKeys] = useState<string[]>([])
  const [instanceName, setInstanceName] = useState('')
  const [quantity, setQuantity] = useState(1)
  const [launching, setLaunching] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleLaunch = async () => {
    if (!selectedType || selectedKeys.length === 0) {
      setError('Please select instance type and at least one SSH key')
      return
    }

    if (!region) {
      setError('Please select a region')
      return
    }

    setLaunching(true)
    try {
      await invoke('lambda_launch_instance', {
        instanceType: selectedType,
        region,
        sshKeys: selectedKeys,
        name: instanceName || null,
        quantity,
      })
      onLaunch()
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
      setLaunching(false)
    }
  }

  const selectedInstanceType = instanceTypes.find(t => t.name === selectedType)

  return (
    <div className="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50 p-4">
      <div className="bg-gray-900 border border-electric-500/30 rounded-xl max-w-2xl w-full max-h-[90vh] overflow-auto anime-glow">
        <div className="p-6 border-b border-gray-800">
          <h2 className="text-2xl font-bold text-electric-400">Launch Instance</h2>
        </div>

        <div className="p-6 space-y-6">
          {/* Instance Type */}
          <div>
            <label className="block text-sm font-medium text-gray-300 mb-3">Instance Type</label>
            {instanceTypes.filter(type => type.regions_with_capacity_available.length > 0).length === 0 ? (
              <div className="p-6 bg-sunset-500/10 border border-sunset-500/30 rounded-lg text-center">
                <div className="text-4xl mb-2">⚠️</div>
                <p className="text-sunset-400 font-medium mb-1">No instances available</p>
                <p className="text-gray-400 text-sm">There are currently no GPU instances with available capacity. Please try again later.</p>
              </div>
            ) : (
              <div className="grid grid-cols-1 gap-3 max-h-64 overflow-y-auto">
                {instanceTypes.filter(type => type.regions_with_capacity_available.length > 0).map((type) => (
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
                      ? 'bg-electric-500/20 border-electric-500/50 text-electric-400'
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
            <label className="block text-sm font-medium text-gray-300 mb-3">SSH Keys (select at least one)</label>
            <div className="space-y-2 max-h-40 overflow-y-auto">
              {sshKeys.map((key) => (
                <label
                  key={key.id}
                  className="flex items-center gap-3 p-3 bg-gray-800/50 border border-gray-700 rounded-lg hover:border-electric-500/30 cursor-pointer transition-all"
                >
                  <input
                    type="checkbox"
                    checked={selectedKeys.includes(key.name)}
                    onChange={(e) => {
                      if (e.target.checked) {
                        setSelectedKeys([...selectedKeys, key.name])
                      } else {
                        setSelectedKeys(selectedKeys.filter(k => k !== key.name))
                      }
                    }}
                    className="w-4 h-4 text-electric-500 bg-gray-700 border-gray-600 rounded focus:ring-electric-500"
                  />
                  <span className="text-gray-300">{key.name}</span>
                </label>
              ))}
            </div>
          </div>

          {/* Instance Name */}
          <div>
            <label className="block text-sm font-medium text-gray-300 mb-2">Instance Name (optional)</label>
            <input
              type="text"
              value={instanceName}
              onChange={(e) => setInstanceName(e.target.value)}
              placeholder="my-instance"
              className="w-full px-4 py-3 bg-gray-800/50 border border-gray-700 rounded-lg focus:outline-none focus:border-electric-500 text-white"
            />
          </div>

          {/* Region */}
          {selectedInstanceType && selectedInstanceType.regions_with_capacity_available.length > 0 && (
            <div>
              <label className="block text-sm font-medium text-gray-300 mb-2">Region</label>
              <select
                value={region}
                onChange={(e) => setRegion(e.target.value)}
                className="w-full px-4 py-3 bg-gray-800/50 border border-gray-700 rounded-lg focus:outline-none focus:border-electric-500 text-white"
              >
                {selectedInstanceType.regions_with_capacity_available.map((r) => (
                  <option key={r.name} value={r.name}>{r.description}</option>
                ))}
              </select>
            </div>
          )}

          {error && (
            <div className="p-3 bg-sunset-500/10 border border-sunset-500/30 rounded-lg text-sunset-400 text-sm">
              {error}
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
            disabled={!selectedType || selectedKeys.length === 0 || launching}
            className="flex-1 px-6 py-3 bg-electric-500/20 hover:bg-electric-500/30 border border-electric-500/50 rounded-lg text-electric-400 font-medium transition-all disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {launching ? 'Launching...' : 'Launch Instance'}
          </button>
        </div>
      </div>
    </div>
  )
}
