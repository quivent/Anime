import { useState } from 'react'

interface LambdaInstance {
  id: string
  name: string
  ip: string
  status: 'running' | 'stopped' | 'pending'
  gpuType: string
  gpuCount: number
  region: string
  createdAt: string
}

type ViewMode = 'list' | 'launch'

export default function ServerManager() {
  const [viewMode, setViewMode] = useState<ViewMode>('list')
  const [instances, setInstances] = useState<LambdaInstance[]>([
    {
      id: '1',
      name: 'inference-prod-01',
      ip: '185.8.107.137',
      status: 'running',
      gpuType: 'GH200',
      gpuCount: 1,
      region: 'us-west',
      createdAt: '2025-01-15',
    },
    {
      id: '2',
      name: 'training-dev-01',
      ip: '84.32.70.4',
      status: 'running',
      gpuType: 'GH200',
      gpuCount: 4,
      region: 'us-east',
      createdAt: '2025-01-14',
    },
  ])

  // Launch new instance form state
  const [newInstanceName, setNewInstanceName] = useState('')
  const [newInstanceRegion, setNewInstanceRegion] = useState('us-west')
  const [newInstanceGpuCount, setNewInstanceGpuCount] = useState(1)
  const [newInstanceGpuType, setNewInstanceGpuType] = useState('GH200')

  // Add existing instance form state
  const [showAddModal, setShowAddModal] = useState(false)
  const [addInstanceName, setAddInstanceName] = useState('')
  const [addInstanceIP, setAddInstanceIP] = useState('')

  const handleLaunchInstance = async () => {
    try {
      // TODO: Implement actual Lambda API call
      const newInstance: LambdaInstance = {
        id: Date.now().toString(),
        name: newInstanceName,
        ip: 'pending',
        status: 'pending',
        gpuType: newInstanceGpuType,
        gpuCount: newInstanceGpuCount,
        region: newInstanceRegion,
        createdAt: new Date().toISOString().split('T')[0],
      }
      setInstances([...instances, newInstance])

      // Reset form
      setNewInstanceName('')
      setViewMode('list')

    } catch (error) {

    }
  }

  const handleAddExistingInstance = () => {
    const newInstance: LambdaInstance = {
      id: Date.now().toString(),
      name: addInstanceName,
      ip: addInstanceIP,
      status: 'running',
      gpuType: 'GH200',
      gpuCount: 1,
      region: 'unknown',
      createdAt: new Date().toISOString().split('T')[0],
    }
    setInstances([...instances, newInstance])
    setShowAddModal(false)
    setAddInstanceName('')
    setAddInstanceIP('')
  }

  const handleConnectInstance = async (_instance: LambdaInstance) => {
    try {

      // TODO: Implement SSH connection test
    } catch (error) {

    }
  }

  const handleRemoveInstance = (id: string) => {
    setInstances(instances.filter(i => i.id !== id))
  }

  return (
    <div className="h-full overflow-y-auto p-6">
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h2 className="text-2xl font-bold">Lambda Instances</h2>
          <p className="text-sm text-gray-400 mt-1">
            Manage your GH200 GPU instances
          </p>
        </div>
        <div className="flex gap-3">
          <button
            onClick={() => setShowAddModal(true)}
            className="px-4 py-2 rounded-lg border border-electric-500 text-electric-400 hover:bg-electric-500/10 transition-all"
          >
            + Add Existing
          </button>
          <button
            onClick={() => setViewMode(viewMode === 'list' ? 'launch' : 'list')}
            className="px-4 py-2 rounded-lg bg-gradient-to-r from-sakura-500 to-neon-400 hover:from-sakura-600 hover:to-neon-500 transition-all font-semibold"
          >
            {viewMode === 'list' ? '🚀 Launch New' : '← Back to List'}
          </button>
        </div>
      </div>

      {/* Instance List View */}
      {viewMode === 'list' && (
        <div className="grid grid-cols-1 gap-4">
          {instances.length === 0 ? (
            <div className="text-center py-12 text-gray-400">
              <div className="text-6xl mb-4">🖥️</div>
              <p>No instances configured</p>
              <p className="text-sm mt-2">Add an existing instance or launch a new one</p>
            </div>
          ) : (
            instances.map((instance) => (
              <div
                key={instance.id}
                className="bg-gray-900/50 backdrop-blur-sm rounded-xl p-6 border border-gray-800 hover:border-gray-700 transition-all"
              >
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <div className="flex items-center gap-3 mb-3">
                      <h3 className="text-xl font-bold">{instance.name}</h3>
                      <span
                        className={`px-2 py-1 rounded-full text-xs font-semibold ${
                          instance.status === 'running'
                            ? 'bg-mint-500/20 text-mint-400 border border-mint-500/30'
                            : instance.status === 'pending'
                            ? 'bg-sunset-500/20 text-sunset-400 border border-sunset-500/30'
                            : 'bg-gray-700 text-gray-400'
                        }`}
                      >
                        {instance.status}
                      </span>
                    </div>

                    <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
                      <div>
                        <p className="text-gray-400 text-xs">IP Address</p>
                        <p className="font-mono font-semibold">{instance.ip}</p>
                      </div>
                      <div>
                        <p className="text-gray-400 text-xs">GPUs</p>
                        <p className="font-semibold">
                          {instance.gpuCount}x {instance.gpuType}
                        </p>
                      </div>
                      <div>
                        <p className="text-gray-400 text-xs">Region</p>
                        <p className="font-semibold">{instance.region}</p>
                      </div>
                      <div>
                        <p className="text-gray-400 text-xs">Created</p>
                        <p className="font-semibold">{instance.createdAt}</p>
                      </div>
                    </div>
                  </div>

                  <div className="flex gap-2 ml-4">
                    <button
                      onClick={() => handleConnectInstance(instance)}
                      className="px-4 py-2 rounded-lg bg-electric-500/10 border border-electric-500/30 text-electric-400 hover:bg-electric-500/20 transition-all text-sm"
                    >
                      Connect
                    </button>
                    <button
                      onClick={() => handleRemoveInstance(instance.id)}
                      className="px-4 py-2 rounded-lg bg-gray-800 hover:bg-gray-700 transition-all text-sm"
                    >
                      Remove
                    </button>
                  </div>
                </div>
              </div>
            ))
          )}
        </div>
      )}

      {/* Launch New Instance View */}
      {viewMode === 'launch' && (
        <div className="max-w-2xl mx-auto">
          <div className="bg-gray-900/50 backdrop-blur-sm rounded-xl p-8 border border-gray-800">
            <h3 className="text-xl font-bold mb-6">Launch New Lambda Instance</h3>

            <div className="space-y-6">
              <div>
                <label className="block text-sm font-semibold mb-2">Instance Name</label>
                <input
                  type="text"
                  value={newInstanceName}
                  onChange={(e) => setNewInstanceName(e.target.value)}
                  placeholder="my-inference-node"
                  className="w-full px-4 py-2 rounded-lg bg-gray-800 border border-gray-700 focus:border-sakura-500 focus:outline-none"
                />
              </div>

              <div>
                <label className="block text-sm font-semibold mb-2">Region</label>
                <select
                  value={newInstanceRegion}
                  onChange={(e) => setNewInstanceRegion(e.target.value)}
                  className="w-full px-4 py-2 rounded-lg bg-gray-800 border border-gray-700 focus:border-sakura-500 focus:outline-none"
                >
                  <option value="us-west">US West</option>
                  <option value="us-east">US East</option>
                  <option value="us-central">US Central</option>
                  <option value="eu-west">EU West</option>
                  <option value="asia-pacific">Asia Pacific</option>
                </select>
              </div>

              <div>
                <label className="block text-sm font-semibold mb-2">GPU Type</label>
                <select
                  value={newInstanceGpuType}
                  onChange={(e) => setNewInstanceGpuType(e.target.value)}
                  className="w-full px-4 py-2 rounded-lg bg-gray-800 border border-gray-700 focus:border-sakura-500 focus:outline-none"
                >
                  <option value="GH200">NVIDIA GH200 (96GB)</option>
                </select>
              </div>

              <div>
                <label className="block text-sm font-semibold mb-2">GPU Count</label>
                <div className="grid grid-cols-4 gap-3">
                  {[1, 2, 4, 8].map((count) => (
                    <button
                      key={count}
                      onClick={() => setNewInstanceGpuCount(count)}
                      className={`p-3 rounded-lg border-2 transition-all ${
                        newInstanceGpuCount === count
                          ? 'border-sakura-500 bg-sakura-500/10'
                          : 'border-gray-700 hover:border-gray-600'
                      }`}
                    >
                      <div className="text-lg font-bold">{count}x</div>
                      <div className="text-xs text-gray-400">GPU</div>
                    </button>
                  ))}
                </div>
              </div>

              <div className="bg-gray-800/50 rounded-lg p-4">
                <h4 className="text-sm font-semibold mb-2 text-gray-400">Estimated Cost</h4>
                <p className="text-2xl font-bold">
                  ${(newInstanceGpuCount * 2.49).toFixed(2)}/hr
                </p>
                <p className="text-xs text-gray-400 mt-1">
                  ~${(newInstanceGpuCount * 2.49 * 730).toFixed(0)}/month
                </p>
              </div>

              <button
                onClick={handleLaunchInstance}
                disabled={!newInstanceName}
                className="w-full px-6 py-3 rounded-lg bg-gradient-to-r from-mint-500 to-electric-500 hover:from-mint-600 hover:to-electric-600 disabled:opacity-50 disabled:cursor-not-allowed transition-all font-bold"
              >
                🚀 Launch Instance
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Add Existing Instance Modal */}
      {showAddModal && (
        <div className="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50">
          <div className="bg-gray-900 rounded-xl p-8 border border-gray-800 max-w-md w-full mx-4">
            <h3 className="text-xl font-bold mb-6">Add Existing Instance</h3>

            <div className="space-y-4 mb-6">
              <div>
                <label className="block text-sm font-semibold mb-2">Instance Name</label>
                <input
                  type="text"
                  value={addInstanceName}
                  onChange={(e) => setAddInstanceName(e.target.value)}
                  placeholder="my-instance"
                  className="w-full px-4 py-2 rounded-lg bg-gray-800 border border-gray-700 focus:border-sakura-500 focus:outline-none"
                />
              </div>

              <div>
                <label className="block text-sm font-semibold mb-2">IP Address</label>
                <input
                  type="text"
                  value={addInstanceIP}
                  onChange={(e) => setAddInstanceIP(e.target.value)}
                  placeholder="185.8.107.137"
                  className="w-full px-4 py-2 rounded-lg bg-gray-800 border border-gray-700 focus:border-sakura-500 focus:outline-none"
                />
              </div>
            </div>

            <div className="flex gap-3">
              <button
                onClick={() => setShowAddModal(false)}
                className="flex-1 px-4 py-2 rounded-lg border border-gray-700 hover:border-gray-600 transition-all"
              >
                Cancel
              </button>
              <button
                onClick={handleAddExistingInstance}
                disabled={!addInstanceName || !addInstanceIP}
                className="flex-1 px-4 py-2 rounded-lg bg-gradient-to-r from-sakura-500 to-neon-400 hover:from-sakura-600 hover:to-neon-500 disabled:opacity-50 disabled:cursor-not-allowed transition-all font-semibold"
              >
                Add Instance
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
