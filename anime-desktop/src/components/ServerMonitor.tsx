import { useState, useEffect, useRef } from 'react'
import { invoke } from '@tauri-apps/api/core'
import { Store } from '@tauri-apps/plugin-store'
import type { Instance } from '../types/lambda'
import type { ServerStatus } from '../types/server'

interface SshKeyInfo {
  path: string
  name: string
  key_type: string
  is_valid: boolean
}

interface ServerMonitorProps {
  instance: Instance
  onClose: () => void
}

export default function ServerMonitor({ instance, onClose }: ServerMonitorProps) {
  const [status, setStatus] = useState<ServerStatus | null>(null)
  const [isConnecting, setIsConnecting] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [sshKeyPath, setSshKeyPath] = useState('')
  const [showKeyDialog, setShowKeyDialog] = useState(true)
  const [availableKeys, setAvailableKeys] = useState<SshKeyInfo[]>([])
  const [isLoadingKeys, setIsLoadingKeys] = useState(true)
  const [useCustomPath, setUseCustomPath] = useState(false)
  const [customKeyPath, setCustomKeyPath] = useState('')
  const intervalRef = useRef<number | null>(null)
  const storeRef = useRef<Store | null>(null)
  const isFetchingRef = useRef(false)
  // Track if component is mounted to prevent state updates after unmount
  const isMountedRef = useRef(true)

  useEffect(() => {
    // Initialize store and load SSH keys
    const initStore = async () => {
      try {
        storeRef.current = await Store.load('settings.json')

        // Load last used key
        const lastUsedKey = await storeRef.current.get<string>('lastUsedSshKey')
        if (lastUsedKey) {
          setSshKeyPath(lastUsedKey)
        }
      } catch (err) {

      }
    }

    const loadKeys = async () => {
      try {
        const keys = await invoke<SshKeyInfo[]>('find_ssh_keys')
        setAvailableKeys(keys)

        // Auto-select first valid key if no last-used key
        if (!sshKeyPath && keys.length > 0) {
          const firstValidKey = keys.find(k => k.is_valid)
          if (firstValidKey) {
            setSshKeyPath(firstValidKey.path)
          }
        }
      } catch (err) {

        setError(err instanceof Error ? err.message : String(err))
      } finally {
        setIsLoadingKeys(false)
      }
    }

    initStore()
    loadKeys()

    // Cleanup function to prevent memory leaks
    return () => {
      // Set mounted flag to false to prevent state updates on unmounted component
      isMountedRef.current = false

      if (intervalRef.current) {
        clearInterval(intervalRef.current)
        intervalRef.current = null
      }
      // Properly handle async cleanup
      handleDisconnect().catch(_ => {})
    }
  }, [])

  const handleConnect = async () => {
    const keyPath = useCustomPath ? customKeyPath : sshKeyPath

    if (!keyPath.trim()) {
      setError('Please select or enter the path to your SSH private key')
      return
    }

    if (!instance.ip) {
      setError('Instance does not have an IP address')
      return
    }

    setIsConnecting(true)
    setError(null)

    try {
      // Validate the key first
      await invoke<boolean>('validate_ssh_key', { keyPath })

      await invoke('connect_to_server', {
        instanceId: instance.id,
        host: instance.ip,
        username: 'ubuntu',
        privateKeyPath: keyPath,
      })

      // Save last used key to store
      if (storeRef.current) {
        await storeRef.current.set('lastUsedSshKey', keyPath)
        await storeRef.current.save()
      }

      setShowKeyDialog(false)

      // Start polling for status - first fetch, then start interval
      await fetchStatus()
      intervalRef.current = window.setInterval(fetchStatus, 3000) // Poll every 3 seconds
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
      setIsConnecting(false)
    }
  }

  const handleDisconnect = async () => {
    if (intervalRef.current) {
      clearInterval(intervalRef.current)
    }

    try {
      await invoke('disconnect_from_server', {
        instanceId: instance.id,
      })
    } catch (err) {

    }
  }

  const fetchStatus = async () => {
    // Prevent overlapping requests
    if (isFetchingRef.current) {
      return
    }

    isFetchingRef.current = true
    try {
      const serverStatus = await invoke<ServerStatus>('get_server_status', {
        instanceId: instance.id,
      })
      // Only update state if component is still mounted (prevents memory leak)
      if (isMountedRef.current) {
        setStatus(serverStatus)
        setError(null)
        setIsConnecting(false)
      }
    } catch (err) {
      // Only update state if component is still mounted (prevents memory leak)
      if (isMountedRef.current) {
        setError(err instanceof Error ? err.message : String(err))
        setIsConnecting(false)
      }
    } finally {
      isFetchingRef.current = false
    }
  }

  const formatUptime = (seconds: number) => {
    const days = Math.floor(seconds / 86400)
    const hours = Math.floor((seconds % 86400) / 3600)
    const mins = Math.floor((seconds % 3600) / 60)

    if (days > 0) return `${days}d ${hours}h ${mins}m`
    if (hours > 0) return `${hours}h ${mins}m`
    return `${mins}m`
  }

  const formatBytes = (bytes: number) => {
    const gb = bytes / 1024 / 1024 / 1024
    if (gb >= 1) return `${gb.toFixed(2)} GB`
    const mb = bytes / 1024 / 1024
    return `${mb.toFixed(2)} MB`
  }

  const getUsageColor = (percent: number) => {
    if (percent >= 90) return 'text-sunset-400 bg-sunset-500/20 border-sunset-500/30'
    if (percent >= 70) return 'text-electric-400 bg-electric-500/20 border-electric-500/30'
    return 'text-mint-400 bg-mint-500/20 border-mint-500/30'
  }

  const getTempColor = (temp: number) => {
    if (temp >= 80) return 'text-sunset-400'
    if (temp >= 70) return 'text-electric-400'
    return 'text-mint-400'
  }

  if (showKeyDialog) {
    return (
      <div className="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50 p-4">
        <div className="bg-gray-900 border border-electric-500/30 rounded-xl max-w-lg w-full anime-glow">
          <div className="p-6 border-b border-gray-800">
            <h2 className="text-2xl font-bold text-electric-400">Connect to Server</h2>
            <p className="text-gray-400 mt-1">
              {instance.name || instance.id} • {instance.ip}
            </p>
          </div>

          <div className="p-6 space-y-4">
            {isLoadingKeys ? (
              <div className="flex items-center justify-center py-8">
                <div className="text-gray-400">Scanning for SSH keys...</div>
              </div>
            ) : (
              <>
                {!useCustomPath && availableKeys.length > 0 ? (
                  <div>
                    <label className="block text-sm font-medium text-gray-300 mb-2">
                      Select SSH Private Key
                    </label>
                    <select
                      value={sshKeyPath}
                      onChange={(e) => setSshKeyPath(e.target.value)}
                      className="w-full px-4 py-3 bg-gray-800/50 border border-gray-700 rounded-lg focus:outline-none focus:border-electric-500 text-white text-sm"
                      autoFocus
                    >
                      <option value="">-- Select a key --</option>
                      {availableKeys.map((key) => (
                        <option key={key.path} value={key.path}>
                          {key.name} ({key.key_type}) {!key.is_valid && ' - Invalid permissions'}
                        </option>
                      ))}
                    </select>
                    <div className="flex items-center justify-between mt-2">
                      <p className="text-xs text-gray-500">
                        {availableKeys.length} key(s) found in ~/.ssh
                      </p>
                      <button
                        onClick={() => setUseCustomPath(true)}
                        className="text-xs text-electric-400 hover:text-electric-300 transition-colors"
                      >
                        Use custom path
                      </button>
                    </div>
                  </div>
                ) : (
                  <div>
                    <label className="block text-sm font-medium text-gray-300 mb-2">
                      SSH Private Key Path
                    </label>
                    <input
                      type="text"
                      value={useCustomPath ? customKeyPath : sshKeyPath}
                      onChange={(e) => useCustomPath ? setCustomKeyPath(e.target.value) : setSshKeyPath(e.target.value)}
                      placeholder="~/.ssh/id_rsa"
                      className="w-full px-4 py-3 bg-gray-800/50 border border-gray-700 rounded-lg focus:outline-none focus:border-electric-500 text-white font-mono text-sm"
                      onKeyDown={(e) => e.key === 'Enter' && handleConnect()}
                      autoFocus
                    />
                    {availableKeys.length > 0 && useCustomPath && (
                      <button
                        onClick={() => {
                          setUseCustomPath(false)
                          setCustomKeyPath('')
                        }}
                        className="text-xs text-electric-400 hover:text-electric-300 transition-colors mt-2"
                      >
                        Choose from detected keys
                      </button>
                    )}
                  </div>
                )}

                {instance.ssh_key_names && instance.ssh_key_names.length > 0 && (
                  <div className="p-3 bg-mint-500/10 border border-mint-500/30 rounded-lg">
                    <p className="text-xs text-mint-400">
                      Instance expects key: {instance.ssh_key_names.join(', ')}
                    </p>
                  </div>
                )}

                {error && (
                  <div className="p-3 bg-sunset-500/10 border border-sunset-500/30 rounded-lg text-sunset-400 text-sm">
                    {error}
                  </div>
                )}
              </>
            )}
          </div>

          <div className="p-6 border-t border-gray-800 flex gap-3">
            <button
              onClick={onClose}
              className="flex-1 px-6 py-3 bg-gray-800/50 hover:bg-gray-700/50 border border-gray-700 rounded-lg text-gray-300 transition-all"
            >
              Cancel
            </button>
            <button
              onClick={handleConnect}
              disabled={isLoadingKeys || (useCustomPath ? !customKeyPath.trim() : !sshKeyPath.trim()) || isConnecting}
              className="flex-1 px-6 py-3 bg-electric-500/20 hover:bg-electric-500/30 border border-electric-500/50 rounded-lg text-electric-400 font-medium transition-all disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {isConnecting ? 'Connecting...' : 'Connect'}
            </button>
          </div>
        </div>
      </div>
    )
  }

  if (isConnecting || !status) {
    return (
      <div className="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50 p-4">
        <div className="bg-gray-900/80 border border-electric-500/30 rounded-xl p-8 text-center anime-glow">
          <div className="text-6xl mb-4 animate-pulse">📊</div>
          <div className="text-electric-400 font-medium">Connecting to server...</div>
        </div>
      </div>
    )
  }

  return (
    <div className="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50 p-4">
      <div className="bg-gray-900 border border-electric-500/30 rounded-xl max-w-6xl w-full max-h-[90vh] overflow-auto anime-glow">
        {/* Header */}
        <div className="p-6 border-b border-gray-800 flex items-center justify-between">
          <div>
            <h2 className="text-2xl font-bold text-electric-400 flex items-center gap-3">
              <span className="text-3xl">📊</span>
              Server Monitor
            </h2>
            <p className="text-gray-400 mt-1">
              {status.hostname} • Uptime: {formatUptime(status.uptime_seconds)}
            </p>
          </div>
          <button
            onClick={onClose}
            className="px-4 py-2 bg-gray-800/50 hover:bg-gray-700/50 border border-gray-700 rounded-lg text-gray-300 transition-all"
          >
            Close
          </button>
        </div>

        <div className="p-6 space-y-6">
          {error && (
            <div className="p-4 bg-sunset-500/10 border border-sunset-500/30 rounded-lg text-sunset-400">
              {error}
            </div>
          )}

          {/* CPU & Memory Grid */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {/* CPU */}
            <div className="bg-gray-800/50 border border-gray-700 rounded-xl p-6">
              <h3 className="text-lg font-semibold text-gray-200 mb-4 flex items-center gap-2">
                <span>🔧</span> CPU
              </h3>

              <div className="space-y-4">
                <div>
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-sm text-gray-400">Usage</span>
                    <span className={`text-lg font-bold ${status.cpu.usage_percent >= 80 ? 'text-sunset-400' : 'text-mint-400'}`}>
                      {status.cpu.usage_percent.toFixed(1)}%
                    </span>
                  </div>
                  <div className="w-full bg-gray-700 rounded-full h-3 overflow-hidden">
                    <div
                      className={`h-full transition-all duration-300 ${status.cpu.usage_percent >= 80 ? 'bg-sunset-500' : 'bg-mint-500'}`}
                      style={{ width: `${Math.min(status.cpu.usage_percent, 100)}%` }}
                    />
                  </div>
                </div>

                <div className="grid grid-cols-2 gap-4 text-sm">
                  <div>
                    <div className="text-gray-500">Cores</div>
                    <div className="text-gray-200 font-mono">{status.cpu.cores}</div>
                  </div>
                  {status.cpu.temperature !== null && (
                    <div>
                      <div className="text-gray-500">Temp</div>
                      <div className={`font-mono font-bold ${getTempColor(status.cpu.temperature)}`}>
                        {status.cpu.temperature.toFixed(1)}°C
                      </div>
                    </div>
                  )}
                </div>

                <div>
                  <div className="text-gray-500 text-sm mb-2">Load Average</div>
                  <div className="flex gap-4 text-sm font-mono">
                    <div>
                      <span className="text-gray-500">1m:</span>{' '}
                      <span className="text-gray-200">{status.cpu.load_average_1m.toFixed(2)}</span>
                    </div>
                    <div>
                      <span className="text-gray-500">5m:</span>{' '}
                      <span className="text-gray-200">{status.cpu.load_average_5m.toFixed(2)}</span>
                    </div>
                    <div>
                      <span className="text-gray-500">15m:</span>{' '}
                      <span className="text-gray-200">{status.cpu.load_average_15m.toFixed(2)}</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            {/* Memory */}
            <div className="bg-gray-800/50 border border-gray-700 rounded-xl p-6">
              <h3 className="text-lg font-semibold text-gray-200 mb-4 flex items-center gap-2">
                <span>💾</span> Memory
              </h3>

              <div className="space-y-4">
                <div>
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-sm text-gray-400">Usage</span>
                    <span className={`text-lg font-bold ${status.memory.usage_percent >= 80 ? 'text-sunset-400' : 'text-mint-400'}`}>
                      {status.memory.usage_percent.toFixed(1)}%
                    </span>
                  </div>
                  <div className="w-full bg-gray-700 rounded-full h-3 overflow-hidden">
                    <div
                      className={`h-full transition-all duration-300 ${status.memory.usage_percent >= 80 ? 'bg-sunset-500' : 'bg-mint-500'}`}
                      style={{ width: `${Math.min(status.memory.usage_percent, 100)}%` }}
                    />
                  </div>
                </div>

                <div className="grid grid-cols-3 gap-4 text-sm">
                  <div>
                    <div className="text-gray-500">Total</div>
                    <div className="text-gray-200 font-mono">{status.memory.total_gb.toFixed(1)} GB</div>
                  </div>
                  <div>
                    <div className="text-gray-500">Used</div>
                    <div className="text-gray-200 font-mono">{status.memory.used_gb.toFixed(1)} GB</div>
                  </div>
                  <div>
                    <div className="text-gray-500">Available</div>
                    <div className="text-gray-200 font-mono">{status.memory.available_gb.toFixed(1)} GB</div>
                  </div>
                </div>

                {status.memory.swap_total_gb > 0 && (
                  <div>
                    <div className="text-gray-500 text-sm mb-2">Swap</div>
                    <div className="flex gap-4 text-sm font-mono">
                      <div>
                        <span className="text-gray-500">Total:</span>{' '}
                        <span className="text-gray-200">{status.memory.swap_total_gb.toFixed(1)} GB</span>
                      </div>
                      <div>
                        <span className="text-gray-500">Used:</span>{' '}
                        <span className="text-gray-200">{status.memory.swap_used_gb.toFixed(1)} GB</span>
                      </div>
                    </div>
                  </div>
                )}
              </div>
            </div>
          </div>

          {/* GPU */}
          {status.gpu && status.gpu.gpus.length > 0 && (
            <div className="bg-gray-800/50 border border-gray-700 rounded-xl p-6">
              <h3 className="text-lg font-semibold text-gray-200 mb-4 flex items-center gap-2">
                <span>🎮</span> GPU ({status.gpu.gpus.length} {status.gpu.gpus.length === 1 ? 'GPU' : 'GPUs'}) • Driver: {status.gpu.driver_version}
              </h3>

              <div className="space-y-4">
                {status.gpu.gpus.map((gpu) => (
                  <div key={gpu.id} className="bg-gray-900/50 border border-gray-700/50 rounded-lg p-4">
                    <div className="flex items-center justify-between mb-3">
                      <div>
                        <div className="font-semibold text-gray-200">GPU {gpu.id}: {gpu.name}</div>
                        <div className="text-xs text-gray-500 mt-1">
                          Memory: {gpu.memory_used_gb.toFixed(1)} / {gpu.memory_total_gb.toFixed(1)} GB
                        </div>
                      </div>
                      <div className="text-right">
                        <div className={`text-lg font-bold ${getTempColor(gpu.temperature)}`}>
                          {gpu.temperature.toFixed(1)}°C
                        </div>
                        <div className="text-xs text-gray-500">
                          {gpu.power_draw_watts.toFixed(0)}W / {gpu.power_limit_watts.toFixed(0)}W
                        </div>
                      </div>
                    </div>

                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <div className="flex items-center justify-between mb-1">
                          <span className="text-xs text-gray-500">GPU Utilization</span>
                          <span className="text-sm font-bold text-mint-400">{gpu.utilization_percent.toFixed(0)}%</span>
                        </div>
                        <div className="w-full bg-gray-700 rounded-full h-2 overflow-hidden">
                          <div
                            className="h-full bg-mint-500 transition-all duration-300"
                            style={{ width: `${Math.min(gpu.utilization_percent, 100)}%` }}
                          />
                        </div>
                      </div>

                      <div>
                        <div className="flex items-center justify-between mb-1">
                          <span className="text-xs text-gray-500">Memory Usage</span>
                          <span className="text-sm font-bold text-electric-400">{gpu.memory_usage_percent.toFixed(0)}%</span>
                        </div>
                        <div className="w-full bg-gray-700 rounded-full h-2 overflow-hidden">
                          <div
                            className="h-full bg-electric-500 transition-all duration-300"
                            style={{ width: `${Math.min(gpu.memory_usage_percent, 100)}%` }}
                          />
                        </div>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Disk & Network */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {/* Disk */}
            <div className="bg-gray-800/50 border border-gray-700 rounded-xl p-6">
              <h3 className="text-lg font-semibold text-gray-200 mb-4 flex items-center gap-2">
                <span>💿</span> Disk
              </h3>

              <div className="space-y-4">
                <div>
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-sm text-gray-400">Usage</span>
                    <span className={`text-lg font-bold ${status.disk.usage_percent >= 80 ? 'text-sunset-400' : 'text-mint-400'}`}>
                      {status.disk.usage_percent.toFixed(1)}%
                    </span>
                  </div>
                  <div className="w-full bg-gray-700 rounded-full h-3 overflow-hidden">
                    <div
                      className={`h-full transition-all duration-300 ${status.disk.usage_percent >= 80 ? 'bg-sunset-500' : 'bg-mint-500'}`}
                      style={{ width: `${Math.min(status.disk.usage_percent, 100)}%` }}
                    />
                  </div>
                </div>

                <div className="grid grid-cols-3 gap-4 text-sm">
                  <div>
                    <div className="text-gray-500">Total</div>
                    <div className="text-gray-200 font-mono">{status.disk.total_gb.toFixed(1)} GB</div>
                  </div>
                  <div>
                    <div className="text-gray-500">Used</div>
                    <div className="text-gray-200 font-mono">{status.disk.used_gb.toFixed(1)} GB</div>
                  </div>
                  <div>
                    <div className="text-gray-500">Available</div>
                    <div className="text-gray-200 font-mono">{status.disk.available_gb.toFixed(1)} GB</div>
                  </div>
                </div>
              </div>
            </div>

            {/* Network */}
            <div className="bg-gray-800/50 border border-gray-700 rounded-xl p-6">
              <h3 className="text-lg font-semibold text-gray-200 mb-4 flex items-center gap-2">
                <span>🌐</span> Network
              </h3>

              <div className="grid grid-cols-2 gap-4 text-sm">
                <div>
                  <div className="text-gray-500 mb-1">Sent</div>
                  <div className="text-gray-200 font-mono text-lg">{formatBytes(status.network.bytes_sent)}</div>
                  <div className="text-gray-500 text-xs mt-1">{status.network.packets_sent.toLocaleString()} packets</div>
                </div>
                <div>
                  <div className="text-gray-500 mb-1">Received</div>
                  <div className="text-gray-200 font-mono text-lg">{formatBytes(status.network.bytes_received)}</div>
                  <div className="text-gray-500 text-xs mt-1">{status.network.packets_received.toLocaleString()} packets</div>
                </div>
              </div>
            </div>
          </div>

          {/* Processes */}
          {status.processes.length > 0 && (
            <div className="bg-gray-800/50 border border-gray-700 rounded-xl p-6">
              <h3 className="text-lg font-semibold text-gray-200 mb-4 flex items-center gap-2">
                <span>⚙️</span> Top Processes ({status.processes.length})
              </h3>

              <div className="overflow-x-auto">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="border-b border-gray-700">
                      <th className="text-left py-2 px-3 text-gray-400 font-medium">PID</th>
                      <th className="text-left py-2 px-3 text-gray-400 font-medium">Name</th>
                      <th className="text-right py-2 px-3 text-gray-400 font-medium">CPU %</th>
                      <th className="text-right py-2 px-3 text-gray-400 font-medium">Memory</th>
                      <th className="text-left py-2 px-3 text-gray-400 font-medium">Status</th>
                    </tr>
                  </thead>
                  <tbody>
                    {status.processes.map((proc) => (
                      <tr key={proc.pid} className="border-b border-gray-800/50 hover:bg-gray-700/30">
                        <td className="py-2 px-3 text-gray-300 font-mono">{proc.pid}</td>
                        <td className="py-2 px-3 text-gray-200 font-medium">{proc.name}</td>
                        <td className="py-2 px-3 text-right">
                          <span className={`font-mono ${proc.cpu_percent > 50 ? 'text-sunset-400' : 'text-gray-300'}`}>
                            {proc.cpu_percent.toFixed(1)}%
                          </span>
                        </td>
                        <td className="py-2 px-3 text-right text-gray-300 font-mono">
                          {proc.memory_mb.toFixed(0)} MB
                        </td>
                        <td className="py-2 px-3">
                          <span className={`px-2 py-1 rounded text-xs ${getUsageColor(proc.cpu_percent)} border`}>
                            {proc.status}
                          </span>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}

          {/* Last Updated */}
          <div className="text-center text-xs text-gray-500">
            Last updated: {new Date(status.timestamp).toLocaleTimeString()} • Auto-refreshing every 3 seconds
          </div>
        </div>
      </div>
    </div>
  )
}
