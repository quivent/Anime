import { useEffect, useState } from 'react'
import { invoke } from '@tauri-apps/api/core'
import { listen } from '@tauri-apps/api/event'
import { usePackageStore } from '../store/packageStore'
import type { Package } from '../types/package'

interface InstalledModel {
  model_id: string
  name: string
  size_bytes: number
  install_path: string
  installed_at: string
  model_type: string
  loaded: boolean
}

interface ModelDownloadProgress {
  model_id: string
  status: 'pending' | 'downloading' | 'completed' | 'failed' | 'cancelled'
  bytes_downloaded: number
  total_bytes: number
  progress: number
  message: string
  download_speed?: number
}

type ModelSize = 'small' | 'medium' | 'large' | 'all'

interface ModelCategory {
  title: string
  icon: string
  color: string
  sizes: { id: ModelSize; label: string; desc: string }[]
}

const categories: ModelCategory[] = [
  {
    title: 'LLM Models',
    icon: '🤖',
    color: 'neon',
    sizes: [
      { id: 'small', label: 'Small (7-8B)', desc: 'Fast inference, 16GB VRAM' },
      { id: 'medium', label: 'Medium (14-34B)', desc: 'Balanced, 32GB VRAM' },
      { id: 'large', label: 'Large (70B+)', desc: 'Best quality, 80GB+ VRAM' },
    ],
  },
]

export default function ModelsView() {
  const { packages, setPackages } = usePackageStore()
  const [selectedSize, setSelectedSize] = useState<ModelSize>('all')
  const [installedModels, setInstalledModels] = useState<Set<string>>(new Set())
  const [downloadProgress, setDownloadProgress] = useState<Map<string, ModelDownloadProgress>>(new Map())
  const [loadingModels, setLoadingModels] = useState<Set<string>>(new Set())

  useEffect(() => {
    loadPackages()
    loadInstalledModels()

    // Listen for download progress events
    const unlisten = listen<ModelDownloadProgress>('model_download_progress', (event) => {
      const progress = event.payload
      setDownloadProgress(prev => {
        const next = new Map(prev)
        next.set(progress.model_id, progress)
        return next
      })

      // Update installed models when download completes
      if (progress.status === 'completed') {
        setInstalledModels(prev => new Set([...prev, progress.model_id]))
        // Remove from progress after a short delay
        setTimeout(() => {
          setDownloadProgress(prev => {
            const next = new Map(prev)
            next.delete(progress.model_id)
            return next
          })
        }, 3000)
      }
    })

    return () => {
      unlisten.then(fn => fn())
    }
  }, [])

  async function loadPackages() {
    try {
      const pkgs = await invoke<Package[]>('get_packages_command')
      setPackages(pkgs)
    } catch (error) {

    }
  }

  async function loadInstalledModels() {
    try {
      const models = await invoke<InstalledModel[]>('list_installed_models', {
        ollamaUrl: null
      })
      const modelIds = new Set(models.map(m => m.model_id))

      // Try to match Ollama model names to our package IDs
      const installedIds = new Set<string>()
      for (const model of models) {
        // Direct match
        if (modelIds.has(model.model_id)) {
          installedIds.add(model.model_id)
        }

        // Try to find matching package by Ollama name
        packages.forEach(pkg => {
          if (model.name.includes('llama3.3') && pkg.id.includes('llama-3_3')) {
            installedIds.add(pkg.id)
          } else if (model.name.includes('mistral') && pkg.id === 'mistral-7b') {
            installedIds.add(pkg.id)
          } else if (model.name.includes('qwen2.5') && pkg.id.includes('qwen-2_5')) {
            installedIds.add(pkg.id)
          } else if (model.name.includes('phi4') && pkg.id === 'phi-4') {
            installedIds.add(pkg.id)
          } else if (model.name.includes('mixtral') && pkg.id === 'mixtral-8x7b') {
            installedIds.add(pkg.id)
          } else if (model.name.includes('deepseek-coder') && pkg.id === 'deepseek-coder-33b') {
            installedIds.add(pkg.id)
          } else if (model.name.includes('yi') && pkg.id === 'yi-34b') {
            installedIds.add(pkg.id)
          } else if (model.name.includes('deepseek-v3') && pkg.id === 'deepseek-v3') {
            installedIds.add(pkg.id)
          }
        })
      }

      setInstalledModels(installedIds)
    } catch (error) {

    }
  }

  // Filter packages to only LLM Models category
  const modelPackages = packages.filter(p => p.category === 'LLM Models')

  // Categorize by size based on naming
  const categorizeModel = (pkg: Package): ModelSize => {
    const name = pkg.name.toLowerCase()
    if (name.includes('7b') || name.includes('8b')) return 'small'
    if (name.includes('14b') || name.includes('34b') || name.includes('33b')) return 'medium'
    if (name.includes('70b') || name.includes('72b') || name.includes('v3')) return 'large'
    return 'small'
  }

  const filteredModels = selectedSize === 'all'
    ? modelPackages
    : modelPackages.filter(p => categorizeModel(p) === selectedSize)

  const groupedModels = filteredModels.reduce((acc, pkg) => {
    const size = categorizeModel(pkg)
    if (!acc[size]) acc[size] = []
    acc[size].push(pkg)
    return acc
  }, {} as Record<ModelSize, Package[]>)

  const handleInstall = async (packageId: string) => {
    try {
      await invoke('download_model', {
        modelId: packageId,
        ollamaUrl: null
      })
    } catch (error) {

      setDownloadProgress(prev => {
        const next = new Map(prev)
        next.set(packageId, {
          model_id: packageId,
          status: 'failed',
          bytes_downloaded: 0,
          total_bytes: 0,
          progress: 0,
          message: `Failed to install: ${error}`,
        })
        return next
      })
    }
  }

  const handleUninstall = async (packageId: string) => {
    try {
      await invoke('delete_model', {
        modelId: packageId,
        modelType: null,
        ollamaUrl: null
      })
      setInstalledModels(prev => {
        const next = new Set(prev)
        next.delete(packageId)
        return next
      })
    } catch (error) {

    }
  }

  const handleLoad = async (packageId: string) => {
    try {
      setLoadingModels(prev => new Set([...prev, packageId]))
      await invoke('load_model', {
        modelId: packageId,
        ollamaUrl: null
      })
      // Show success for a moment
      setTimeout(() => {
        setLoadingModels(prev => {
          const next = new Set(prev)
          next.delete(packageId)
          return next
        })
      }, 2000)
    } catch (error) {

      setLoadingModels(prev => {
        const next = new Set(prev)
        next.delete(packageId)
        return next
      })
    }
  }

  return (
    <div className="h-full flex flex-col p-6 overflow-hidden">
      {/* Header */}
      <div className="mb-6">
        <h2 className="text-3xl font-bold mb-2 sakura-gradient bg-clip-text text-transparent">
          Model Library
        </h2>
        <p className="text-gray-400">
          Manage your LLM models and download weights
        </p>
      </div>

      {/* Size Filter */}
      <div className="flex items-center gap-2 mb-6 overflow-x-auto pb-2">
        <button
          onClick={() => setSelectedSize('all')}
          className={`px-4 py-2 rounded-lg font-medium text-sm whitespace-nowrap transition-all ${
            selectedSize === 'all'
              ? 'bg-sakura-500 text-white'
              : 'bg-gray-800 text-gray-400 hover:bg-gray-700'
          }`}
        >
          🍒 All Models
        </button>
        {categories[0].sizes.map((size) => (
          <button
            key={size.id}
            onClick={() => setSelectedSize(size.id)}
            className={`px-4 py-2 rounded-lg font-medium text-sm whitespace-nowrap transition-all ${
              selectedSize === size.id
                ? 'bg-neon-500 text-white'
                : 'bg-gray-800 text-gray-400 hover:bg-gray-700'
            }`}
          >
            {size.label}
          </button>
        ))}
      </div>

      {/* Models Grid */}
      <div className="flex-1 overflow-y-auto">
        {selectedSize === 'all' ? (
          // Show grouped by size
          <div className="space-y-8">
            {(['small', 'medium', 'large'] as ModelSize[]).map((size) => {
              const models = groupedModels[size] || []
              if (models.length === 0) return null

              const sizeInfo = categories[0].sizes.find(s => s.id === size)!

              return (
                <div key={size}>
                  <div className="mb-4">
                    <h3 className="text-xl font-bold text-neon-400 mb-1">
                      {sizeInfo.label}
                    </h3>
                    <p className="text-sm text-gray-500">{sizeInfo.desc}</p>
                  </div>
                  <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                    {models.map((model) => (
                      <ModelCard
                        key={model.id}
                        model={model}
                        installed={installedModels.has(model.id)}
                        onInstall={handleInstall}
                        onUninstall={handleUninstall}
                        onLoad={handleLoad}
                        downloadProgress={downloadProgress.get(model.id)}
                        isLoading={loadingModels.has(model.id)}
                      />
                    ))}
                  </div>
                </div>
              )
            })}
          </div>
        ) : (
          // Show filtered size only
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {filteredModels.map((model) => (
              <ModelCard
                key={model.id}
                model={model}
                installed={installedModels.has(model.id)}
                onInstall={handleInstall}
                onUninstall={handleUninstall}
                onLoad={handleLoad}
                downloadProgress={downloadProgress.get(model.id)}
                isLoading={loadingModels.has(model.id)}
              />
            ))}
          </div>
        )}
      </div>
    </div>
  )
}

interface ModelCardProps {
  model: Package
  installed: boolean
  onInstall: (id: string) => void
  onUninstall: (id: string) => void
  onLoad: (id: string) => void
  downloadProgress?: ModelDownloadProgress
  isLoading: boolean
}

function ModelCard({ model, installed, onInstall, onUninstall, onLoad, downloadProgress, isLoading }: ModelCardProps) {
  const isDownloading = downloadProgress?.status === 'downloading'
  const downloadFailed = downloadProgress?.status === 'failed'

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return '0 B'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return `${(bytes / Math.pow(k, i)).toFixed(2)} ${sizes[i]}`
  }

  const formatSpeed = (bytesPerSec?: number) => {
    if (!bytesPerSec) return ''
    return `${formatBytes(bytesPerSec)}/s`
  }
  return (
    <div
      className={`p-4 rounded-lg border-2 transition-all ${
        installed
          ? 'border-mint-500 bg-mint-500/10'
          : isDownloading
          ? 'border-neon-500 bg-neon-500/10'
          : downloadFailed
          ? 'border-red-500 bg-red-500/10'
          : 'border-gray-800 bg-gray-900/50 hover:border-gray-700'
      }`}
    >
      <div className="flex items-start justify-between mb-3">
        <div className="flex-1">
          <h3 className="font-bold text-lg mb-1">{model.name}</h3>
          <p className="text-xs text-gray-400 line-clamp-2">{model.description}</p>
        </div>
        {installed && !isDownloading && (
          <span className="text-mint-400 text-xl ml-2">✓</span>
        )}
        {isDownloading && (
          <span className="text-neon-400 text-xl ml-2 animate-pulse">⬇</span>
        )}
      </div>

      {/* Download Progress */}
      {isDownloading && downloadProgress && (
        <div className="mb-4">
          <div className="flex justify-between text-xs text-gray-400 mb-1">
            <span>{downloadProgress.message}</span>
            <span>{downloadProgress.progress.toFixed(1)}%</span>
          </div>
          <div className="w-full bg-gray-800 rounded-full h-2 overflow-hidden">
            <div
              className="bg-gradient-to-r from-neon-500 to-electric-500 h-full transition-all duration-300"
              style={{ width: `${downloadProgress.progress}%` }}
            />
          </div>
          {downloadProgress.download_speed && (
            <div className="text-xs text-gray-500 mt-1 text-right">
              {formatSpeed(downloadProgress.download_speed)}
            </div>
          )}
        </div>
      )}

      {/* Error Message */}
      {downloadFailed && downloadProgress && (
        <div className="mb-4 p-2 bg-red-500/20 rounded text-xs text-red-400">
          {downloadProgress.message}
        </div>
      )}

      <div className="flex items-center gap-3 text-xs text-gray-500 mb-4">
        <div className="flex items-center gap-1">
          <span>⏱️</span>
          <span>~{model.estimatedTime}m</span>
        </div>
        <div className="flex items-center gap-1">
          <span>💾</span>
          <span>{model.size}</span>
        </div>
      </div>

      {installed ? (
        <div className="flex gap-2">
          <button
            onClick={() => onUninstall(model.id)}
            disabled={isDownloading}
            className="flex-1 px-4 py-2 rounded-lg bg-gray-800 text-gray-400 hover:bg-gray-700 transition-colors text-sm font-medium disabled:opacity-50 disabled:cursor-not-allowed"
          >
            Uninstall
          </button>
          <button
            onClick={() => onLoad(model.id)}
            disabled={isDownloading || isLoading}
            className="flex-1 px-4 py-2 rounded-lg bg-electric-500 text-white hover:bg-electric-600 transition-colors text-sm font-medium disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {isLoading ? 'Loading...' : 'Load'}
          </button>
        </div>
      ) : (
        <button
          onClick={() => onInstall(model.id)}
          disabled={isDownloading}
          className="w-full px-4 py-2 rounded-lg bg-neon-500 text-white hover:bg-neon-600 transition-colors text-sm font-medium disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {isDownloading ? 'Downloading...' : downloadFailed ? 'Retry Download' : 'Download'}
        </button>
      )}
    </div>
  )
}
