import { useState, useEffect, useRef } from 'react'
import { usePackageStore } from '../store/packageStore'

interface PackageInstallStatus {
  id: string
  name: string
  status: 'pending' | 'installing' | 'completed' | 'failed'
  progress: number
  logs: string[]
  startTime?: number
  endTime?: number
}

export default function InstallProgress() {
  const { selectedPackages } = usePackageStore()
  const [packageStatuses, setPackageStatuses] = useState<PackageInstallStatus[]>([])
  const [selectedPackageId, setSelectedPackageId] = useState<string | null>(null)
  const [autoScroll, setAutoScroll] = useState(true)
  // Store interval ID in a ref to ensure proper cleanup
  const installIntervalRef = useRef<number | null>(null)

  useEffect(() => {
    // Initialize package statuses from selected packages
    const statuses: PackageInstallStatus[] = Array.from(selectedPackages).map(id => ({
      id,
      name: id.charAt(0).toUpperCase() + id.slice(1),
      status: 'pending',
      progress: 0,
      logs: [],
    }))
    setPackageStatuses(statuses)

    // Simulate installation progress (remove when Rust backend is ready)
    if (statuses.length > 0) {
      simulateInstallation(statuses)
    }

    // Cleanup function to prevent memory leak
    return () => {
      if (installIntervalRef.current !== null) {
        clearInterval(installIntervalRef.current)
        installIntervalRef.current = null
      }
    }
  }, [selectedPackages])

  const simulateInstallation = (statuses: PackageInstallStatus[]) => {
    let currentIndex = 0

    // Clear any existing interval before creating a new one
    if (installIntervalRef.current !== null) {
      clearInterval(installIntervalRef.current)
    }

    installIntervalRef.current = window.setInterval(() => {
      if (currentIndex >= statuses.length) {
        if (installIntervalRef.current !== null) {
          clearInterval(installIntervalRef.current)
          installIntervalRef.current = null
        }
        return
      }

      setPackageStatuses(prev => {
        const updated = [...prev]
        const current = updated[currentIndex]

        if (current.status === 'pending') {
          current.status = 'installing'
          current.startTime = Date.now()
          current.logs.push(`[${new Date().toLocaleTimeString()}] Starting ${current.name} installation...`)
          current.logs.push(`[${new Date().toLocaleTimeString()}] Downloading package...`)
        } else if (current.status === 'installing') {
          current.progress += 10

          if (current.progress >= 50 && current.progress < 60) {
            current.logs.push(`[${new Date().toLocaleTimeString()}] Extracting files...`)
          } else if (current.progress >= 80 && current.progress < 90) {
            current.logs.push(`[${new Date().toLocaleTimeString()}] Running post-install scripts...`)
          }

          if (current.progress >= 100) {
            current.status = 'completed'
            current.endTime = Date.now()
            current.logs.push(`[${new Date().toLocaleTimeString()}] ${current.name} installed successfully!`)
            currentIndex++

            // Start next package
            if (currentIndex < statuses.length) {
              updated[currentIndex].status = 'pending'
            }
          }
        }

        return updated
      })
    }, 500)
  }

  const selectedPackage = selectedPackageId
    ? packageStatuses.find(p => p.id === selectedPackageId)
    : packageStatuses.find(p => p.status === 'installing')

  const completedCount = packageStatuses.filter(p => p.status === 'completed').length
  const failedCount = packageStatuses.filter(p => p.status === 'failed').length
  const totalCount = packageStatuses.length
  const overallProgress = totalCount > 0 ? (completedCount / totalCount) * 100 : 0

  return (
    <div className="h-full overflow-y-auto p-6">
      {/* Header */}
      <div className="mb-6">
        <div className="flex items-center justify-between mb-4">
          <div>
            <h2 className="text-2xl font-bold">Installation Progress</h2>
            <p className="text-sm text-gray-400 mt-1">
              {completedCount} of {totalCount} packages installed
            </p>
          </div>
          <div className="flex items-center gap-4">
            <div className="text-right">
              <div className="text-3xl font-bold text-mint-400">{completedCount}</div>
              <div className="text-xs text-gray-400">Completed</div>
            </div>
            {failedCount > 0 && (
              <div className="text-right">
                <div className="text-3xl font-bold text-red-400">{failedCount}</div>
                <div className="text-xs text-gray-400">Failed</div>
              </div>
            )}
          </div>
        </div>

        {/* Overall Progress Bar */}
        <div className="bg-gray-800 rounded-full h-3 overflow-hidden">
          <div
            className="h-full bg-gradient-to-r from-sakura-500 to-mint-400 transition-all duration-500"
            style={{ width: `${overallProgress}%` }}
          />
        </div>
      </div>

      {packageStatuses.length === 0 ? (
        <div className="text-center py-12 text-gray-400">
          <div className="text-6xl mb-4">📦</div>
          <p>No packages selected for installation</p>
          <p className="text-sm mt-2">Select packages from the wizard or package grid</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Package List */}
          <div className="lg:col-span-1 space-y-2">
            <h3 className="text-lg font-bold mb-3">Packages</h3>
            <div className="space-y-2 max-h-96 overflow-y-auto">
              {packageStatuses.map((pkg) => (
                <button
                  key={pkg.id}
                  onClick={() => setSelectedPackageId(pkg.id)}
                  className={`w-full p-3 rounded-lg border-2 transition-all text-left ${
                    selectedPackageId === pkg.id
                      ? 'border-sakura-500 bg-sakura-500/10'
                      : 'border-gray-800 hover:border-gray-700 bg-gray-900/50'
                  }`}
                >
                  <div className="flex items-center justify-between mb-2">
                    <span className="font-semibold">{pkg.name}</span>
                    {pkg.status === 'completed' && (
                      <span className="text-mint-400 text-xl">✓</span>
                    )}
                    {pkg.status === 'failed' && (
                      <span className="text-red-400 text-xl">✗</span>
                    )}
                    {pkg.status === 'installing' && (
                      <span className="text-electric-400 text-xl animate-spin">⟳</span>
                    )}
                  </div>
                  {pkg.status === 'installing' && (
                    <div className="bg-gray-800 rounded-full h-1.5 overflow-hidden">
                      <div
                        className="h-full bg-electric-400 transition-all duration-300"
                        style={{ width: `${pkg.progress}%` }}
                      />
                    </div>
                  )}
                  {pkg.status === 'completed' && pkg.startTime && pkg.endTime && (
                    <p className="text-xs text-gray-400 mt-1">
                      {Math.round((pkg.endTime - pkg.startTime) / 1000)}s
                    </p>
                  )}
                </button>
              ))}
            </div>
          </div>

          {/* Log Output */}
          <div className="lg:col-span-2">
            <div className="flex items-center justify-between mb-3">
              <h3 className="text-lg font-bold">
                {selectedPackage ? selectedPackage.name : 'Select a package'}
              </h3>
              <label className="flex items-center gap-2 text-sm text-gray-400 cursor-pointer">
                <input
                  type="checkbox"
                  checked={autoScroll}
                  onChange={(e) => setAutoScroll(e.target.checked)}
                  className="rounded"
                />
                Auto-scroll
              </label>
            </div>

            <div className="bg-gray-950 rounded-lg border border-gray-800 p-4 h-96 overflow-y-auto font-mono text-sm">
              {selectedPackage ? (
                selectedPackage.logs.length > 0 ? (
                  <div className="space-y-1">
                    {selectedPackage.logs.map((log, idx) => (
                      <div key={idx} className="text-gray-300">
                        {log}
                      </div>
                    ))}
                    {selectedPackage.status === 'installing' && (
                      <div className="text-electric-400 animate-pulse">
                        Installing... {selectedPackage.progress}%
                      </div>
                    )}
                  </div>
                ) : (
                  <div className="text-gray-500">No logs yet...</div>
                )
              ) : (
                <div className="text-gray-500">Select a package to view logs</div>
              )}
            </div>

            {/* Installation Stats */}
            {selectedPackage && selectedPackage.status === 'completed' && (
              <div className="mt-4 grid grid-cols-3 gap-4">
                <div className="bg-gray-900/50 rounded-lg p-3 border border-gray-800">
                  <p className="text-xs text-gray-400">Status</p>
                  <p className="text-lg font-bold text-mint-400">Completed</p>
                </div>
                <div className="bg-gray-900/50 rounded-lg p-3 border border-gray-800">
                  <p className="text-xs text-gray-400">Duration</p>
                  <p className="text-lg font-bold">
                    {selectedPackage.startTime && selectedPackage.endTime
                      ? `${Math.round((selectedPackage.endTime - selectedPackage.startTime) / 1000)}s`
                      : '-'}
                  </p>
                </div>
                <div className="bg-gray-900/50 rounded-lg p-3 border border-gray-800">
                  <p className="text-xs text-gray-400">Progress</p>
                  <p className="text-lg font-bold">100%</p>
                </div>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Completion Message */}
      {completedCount === totalCount && totalCount > 0 && (
        <div className="mt-6 bg-gradient-to-r from-mint-500/10 to-electric-500/10 border border-mint-500/30 rounded-xl p-6 text-center">
          <div className="text-5xl mb-3">🎉</div>
          <h3 className="text-2xl font-bold mb-2">Installation Complete!</h3>
          <p className="text-gray-400">
            All {totalCount} packages have been installed successfully.
          </p>
        </div>
      )}
    </div>
  )
}
