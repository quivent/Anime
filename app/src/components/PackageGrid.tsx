import { useEffect, useState } from 'react'
import { invoke } from '@tauri-apps/api/core'
import { listen } from '@tauri-apps/api/event'
import { usePackageStore } from '../store/packageStore'
import { useInstanceStore } from '../store/instanceStore'
import type { Package, Category, InstallProgress } from '../types/package'
import { CategoryEmojis, CategoryColors } from '../types/package'
import { ToastContainer } from './Toast'
import { ConfirmModal } from './ConfirmModal'
import { useToast } from '../hooks/useToast'

interface PackageInstallState {
  status: 'not_installed' | 'installing' | 'installed'
  progress: number
  message: string
}

export default function PackageGrid() {
  const { packages, selectedPackages, togglePackage, setPackages, clearSelection } = usePackageStore()
  const { selectedInstance } = useInstanceStore()
  const [filter, setFilter] = useState<Category | 'all'>('all')
  const [packageStates, setPackageStates] = useState<Map<string, PackageInstallState>>(new Map())
  const [isBatchInstalling, setIsBatchInstalling] = useState(false)
  const { toasts, showError, showSuccess, showWarning, removeToast } = useToast()

  // Confirmation modal state
  const [confirmModal, setConfirmModal] = useState<{
    isOpen: boolean
    title: string
    message: string
    onConfirm: () => void
  }>({
    isOpen: false,
    title: '',
    message: '',
    onConfirm: () => {}
  })

  useEffect(() => {
    // Load packages from Rust backend
    loadPackages()
    setupProgressListener()
  }, [])

  async function loadPackages() {
    try {
      const pkgs = await invoke<Package[]>('get_packages_command')
      setPackages(pkgs)
    } catch (error) {

    }
  }

  function setupProgressListener() {
    listen<InstallProgress>('install_progress', (event) => {
      const progress = event.payload
      setPackageStates(prev => {
        const newStates = new Map(prev)
        newStates.set(progress.packageId, {
          status: progress.status === 'completed' ? 'installed' :
                 progress.status === 'installing' ? 'installing' :
                 progress.status === 'failed' ? 'not_installed' : 'not_installed',
          progress: progress.progress,
          message: progress.message
        })
        return newStates
      })
    })
  }

  async function handleInstallPackage(packageId: string) {
    if (!selectedInstance) {
      showError('Please select an instance from Lambda Cloud first!')
      return
    }

    if (!selectedInstance.ip) {
      showError('Selected instance has no IP address!')
      return
    }

    try {
      setPackageStates(prev => {
        const newStates = new Map(prev)
        newStates.set(packageId, {
          status: 'installing',
          progress: 0,
          message: 'Starting installation...'
        })
        return newStates
      })

      // Lambda instances use 'ubuntu' as the default username
      const username = 'ubuntu'

      await invoke('install_package_remote', {
        host: selectedInstance.ip,
        username,
        packageId
      })
    } catch (error) {

      setPackageStates(prev => {
        const newStates = new Map(prev)
        newStates.set(packageId, {
          status: 'not_installed',
          progress: 0,
          message: `Failed: ${error}`
        })
        return newStates
      })
    }
  }

  async function handleBatchInstall() {
    // Validate instance selection
    if (!selectedInstance) {
      showError('Please select an instance from Lambda Cloud first!')
      return
    }

    if (!selectedInstance.ip) {
      showError('Selected instance has no IP address!')
      return
    }

    // Get all selected package IDs
    const packageIds = Array.from(selectedPackages)

    if (packageIds.length === 0) {
      showWarning('No packages selected!')
      return
    }

    // Confirm batch installation
    const packageNames = packageIds
      .map(id => packages.find(p => p.id === id)?.name)
      .filter(Boolean)
      .join(', ')

    setConfirmModal({
      isOpen: true,
      title: 'Confirm Batch Installation',
      message: `Install ${packageIds.length} package${packageIds.length > 1 ? 's' : ''} on ${selectedInstance.hostname || selectedInstance.ip}?\n\nPackages: ${packageNames}`,
      onConfirm: () => {
        setConfirmModal(prev => ({ ...prev, isOpen: false }))
        executeBatchInstall(packageIds)
      }
    })
  }

  async function executeBatchInstall(packageIds: string[]) {
    if (!selectedInstance) return

    // Set batch installing flag
    setIsBatchInstalling(true)

    // Initialize all packages to installing state
    setPackageStates(prev => {
      const newStates = new Map(prev)
      packageIds.forEach(packageId => {
        newStates.set(packageId, {
          status: 'installing',
          progress: 0,
          message: 'Queued for installation...'
        })
      })
      return newStates
    })

    // Lambda instances use 'ubuntu' as the default username
    const username = 'ubuntu'
    const host = selectedInstance.ip

    // Install packages sequentially to avoid overwhelming the server
    // and to provide better progress tracking
    let successCount = 0
    let failedPackages: string[] = []

    for (let i = 0; i < packageIds.length; i++) {
      const packageId = packageIds[i]
      const pkg = packages.find(p => p.id === packageId)
      const packageName = pkg?.name || packageId

      try {
        // Update status to show current package being installed
        setPackageStates(prev => {
          const newStates = new Map(prev)
          newStates.set(packageId, {
            status: 'installing',
            progress: 0,
            message: `Installing (${i + 1}/${packageIds.length})...`
          })
          return newStates
        })

        await invoke('install_package_remote', {
          host,
          username,
          packageId
        })

        successCount++

      } catch (error) {

        failedPackages.push(packageName)

        // Mark as failed
        setPackageStates(prev => {
          const newStates = new Map(prev)
          newStates.set(packageId, {
            status: 'not_installed',
            progress: 0,
            message: `Failed: ${error}`
          })
          return newStates
        })
      }
    }

    // Clear selection after installation attempts
    clearSelection()
    setIsBatchInstalling(false)

    // Show summary
    if (failedPackages.length === 0) {
      showSuccess(`Successfully installed all ${successCount} package${successCount > 1 ? 's' : ''}!`)
    } else if (successCount === 0) {
      showError(`Failed to install all packages:\n${failedPackages.join(', ')}`)
    } else {
      showWarning(
        `Batch installation completed:\n\n` +
        `✓ Installed: ${successCount} package${successCount > 1 ? 's' : ''}\n` +
        `✗ Failed: ${failedPackages.length} package${failedPackages.length > 1 ? 's' : ''}\n\n` +
        `Failed packages: ${failedPackages.join(', ')}`
      )
    }
  }

  function getPackageStatusIcon(packageId: string): string {
    const state = packageStates.get(packageId)
    if (!state) return '○'

    switch (state.status) {
      case 'installed': return '✓'
      case 'installing': return '⏳'
      case 'not_installed': return '○'
      default: return '○'
    }
  }

  const categories = Array.from(new Set(packages.map(p => p.category)))
  const filteredPackages = filter === 'all'
    ? packages
    : packages.filter(p => p.category === filter)

  return (
    <div className="h-full flex flex-col p-6 overflow-hidden">
      {/* Selected Instance Info */}
      {selectedInstance ? (
        <div className="mb-4 p-3 rounded-lg bg-mint-500/10 border border-mint-500/30 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-2 h-2 bg-mint-500 rounded-full animate-pulse" />
            <div>
              <div className="text-sm font-medium text-mint-400">
                Installing to: {selectedInstance.hostname || selectedInstance.ip}
              </div>
              <div className="text-xs text-gray-400">
                {selectedInstance.instance_type.description} • {selectedInstance.region.description}
              </div>
            </div>
          </div>
        </div>
      ) : (
        <div className="mb-4 p-3 rounded-lg bg-gray-800/50 border border-gray-700 flex items-center gap-3">
          <div className="w-2 h-2 bg-gray-600 rounded-full" />
          <div className="text-sm text-gray-400">
            No instance selected. Select an instance from Lambda Cloud to enable installation.
          </div>
        </div>
      )}

      {/* Category Filter */}
      <div className="flex items-center gap-2 mb-6 overflow-x-auto pb-2">
        <button
          onClick={() => setFilter('all')}
          className={`
            px-4 py-2 rounded-lg font-medium text-sm whitespace-nowrap transition-all
            ${filter === 'all'
              ? 'bg-sakura-500 text-white'
              : 'bg-gray-800 text-gray-400 hover:bg-gray-700'
            }
          `}
        >
          🍒 All Packages
        </button>
        {categories.map((category) => (
          <button
            key={category}
            onClick={() => setFilter(category)}
            className={`
              px-4 py-2 rounded-lg font-medium text-sm whitespace-nowrap transition-all
              ${filter === category
                ? `bg-${CategoryColors[category]}-500 text-white`
                : 'bg-gray-800 text-gray-400 hover:bg-gray-700'
              }
            `}
          >
            {CategoryEmojis[category]} {category}
          </button>
        ))}
      </div>

      {/* Package Grid */}
      <div className="flex-1 overflow-y-auto">
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {filteredPackages.map((pkg) => {
            const isSelected = selectedPackages.has(pkg.id)

            const statusIcon = getPackageStatusIcon(pkg.id)
            const state = packageStates.get(pkg.id)
            const isInstalling = state?.status === 'installing'
            const isInstalled = state?.status === 'installed'

            return (
              <div
                key={pkg.id}
                className={`
                  relative text-left p-4 rounded-lg border-2 transition-all duration-200
                  ${isSelected
                    ? 'border-sakura-500 bg-sakura-500/10'
                    : isInstalled
                    ? 'border-mint-500 bg-mint-500/10'
                    : isInstalling
                    ? 'border-electric-500 bg-electric-500/10 anime-glow'
                    : 'border-gray-800 bg-gray-900/50 hover:border-gray-700 hover:bg-gray-800/50'
                  }
                `}
              >
                {/* Selection checkbox */}
                <div className="absolute top-3 left-3 z-10">
                  <input
                    type="checkbox"
                    checked={isSelected}
                    onChange={() => togglePackage(pkg.id)}
                    disabled={isInstalling || isInstalled}
                    className="w-5 h-5 rounded border-2 border-gray-600 bg-gray-800 checked:bg-sakura-500 checked:border-sakura-500 cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
                  />
                </div>

                {/* Status indicator circle */}
                <div className={`absolute top-3 right-3 w-3 h-3 rounded-full ${
                  isInstalled ? 'bg-mint-500' :
                  isInstalling ? 'bg-electric-500 animate-pulse' :
                  'bg-gray-600'
                }`} />

                <div className="flex items-start justify-between mb-2">
                  <div className="flex items-center gap-2">
                    <span className="text-2xl">{CategoryEmojis[pkg.category]}</span>
                    <h3 className="font-bold text-lg">{pkg.name}</h3>
                  </div>
                  <span className={`text-2xl ${
                    isInstalled ? 'text-mint-400' :
                    isInstalling ? 'text-electric-400' :
                    'text-gray-600'
                  }`}>{statusIcon}</span>
                </div>

                <p className="text-sm text-gray-400 mb-3 line-clamp-2">
                  {pkg.description}
                </p>

                {/* Progress bar for installing packages */}
                {isInstalling && state && (
                  <div className="mb-3">
                    <div className="flex items-center justify-between text-xs text-electric-400 mb-1">
                      <span>{state.message}</span>
                      <span>{state.progress}%</span>
                    </div>
                    <div className="w-full h-1.5 bg-gray-800 rounded-full overflow-hidden">
                      <div
                        className="h-full bg-electric-500 transition-all duration-300"
                        style={{ width: `${state.progress}%` }}
                      />
                    </div>
                  </div>
                )}

                <div className="flex items-center gap-4 text-xs text-gray-500 mb-3">
                  <div className="flex items-center gap-1">
                    <span>⏱️</span>
                    <span>{pkg.estimatedTime}m</span>
                  </div>
                  <div className="flex items-center gap-1">
                    <span>💾</span>
                    <span>{pkg.size}</span>
                  </div>
                  {pkg.dependencies.length > 0 && (
                    <div className="flex items-center gap-1">
                      <span>🔗</span>
                      <span>{pkg.dependencies.length}</span>
                    </div>
                  )}
                </div>

                {/* Install button */}
                <button
                  onClick={() => handleInstallPackage(pkg.id)}
                  disabled={isInstalling || isInstalled}
                  className={`w-full py-2 px-4 rounded-lg font-medium text-sm transition-all ${
                    isInstalled
                      ? 'bg-mint-500/20 border border-mint-500/50 text-mint-400 cursor-default'
                      : isInstalling
                      ? 'bg-electric-500/20 border border-electric-500/50 text-electric-400 cursor-wait'
                      : 'bg-sakura-500 text-white hover:bg-sakura-600 anime-glow'
                  }`}
                >
                  {isInstalled ? '✓ Installed' : isInstalling ? '⏳ Installing...' : '🚀 Install'}
                </button>
              </div>
            )
          })}
        </div>
      </div>

      {/* Action Bar */}
      {selectedPackages.size > 0 && (
        <div className="mt-4 p-4 rounded-lg bg-sakura-500/10 border border-sakura-500/30">
          <div className="flex items-center justify-between">
            <div className="text-sm flex-1 mr-4">
              <div className="mb-1">
                <span className="text-sakura-400 font-bold">{selectedPackages.size}</span>
                <span className="text-gray-400"> packages selected:</span>
              </div>
              <div className="text-xs text-gray-400 flex flex-wrap gap-1">
                {Array.from(selectedPackages).map(id => {
                  const pkg = packages.find(p => p.id === id)
                  return pkg ? (
                    <span key={id} className="px-2 py-0.5 bg-sakura-500/20 rounded">
                      {pkg.name}
                    </span>
                  ) : null
                })}
              </div>
            </div>
            <button
              onClick={handleBatchInstall}
              disabled={isBatchInstalling || !selectedInstance}
              className={`px-6 py-2 rounded-lg font-medium transition-colors whitespace-nowrap ${
                isBatchInstalling
                  ? 'bg-electric-500/50 text-electric-200 cursor-wait'
                  : !selectedInstance
                  ? 'bg-gray-600 text-gray-400 cursor-not-allowed'
                  : 'bg-sakura-500 text-white hover:bg-sakura-600 anime-glow'
              }`}
            >
              {isBatchInstalling ? '⏳ Installing...' : '🚀 Install Selected'}
            </button>
          </div>
        </div>
      )}

      {/* Toast Notifications */}
      <ToastContainer toasts={toasts} onClose={removeToast} />

      {/* Confirmation Modal */}
      <ConfirmModal
        isOpen={confirmModal.isOpen}
        title={confirmModal.title}
        message={confirmModal.message}
        confirmText="Install"
        cancelText="Cancel"
        onConfirm={confirmModal.onConfirm}
        onCancel={() => setConfirmModal(prev => ({ ...prev, isOpen: false }))}
        variant="info"
      />
    </div>
  )
}
