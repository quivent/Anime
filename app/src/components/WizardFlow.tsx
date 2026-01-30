import { useState } from 'react'
import { invoke } from '@tauri-apps/api/core'
import { Package } from '../types/package'

type NodeType = 'inference' | 'training' | 'art-video' | 'development' | 'multimodal' | 'everything'

interface WizardStep {
  title: string
  description: string
}

const steps: WizardStep[] = [
  { title: 'Node Type', description: 'Select your deployment configuration' },
  { title: 'Packages', description: 'Choose software to install' },
  { title: 'Configuration', description: 'Set instance parameters' },
  { title: 'Review', description: 'Confirm and deploy' },
]

const nodeTypes: { id: NodeType; name: string; description: string; icon: string }[] = [
  {
    id: 'inference',
    name: 'Inference',
    description: 'Optimized for running LLM inference with vLLM',
    icon: '⚡',
  },
  {
    id: 'training',
    name: 'Training',
    description: 'High-performance training with PyTorch and DeepSpeed',
    icon: '🧠',
  },
  {
    id: 'art-video',
    name: 'Art & Video',
    description: 'Creative workloads with ComfyUI and Mochi',
    icon: '🎨',
  },
  {
    id: 'development',
    name: 'Development',
    description: 'Full development environment with all tools',
    icon: '💻',
  },
  {
    id: 'multimodal',
    name: 'Multimodal',
    description: 'Vision and audio models with LLaVA and Whisper',
    icon: '👁️',
  },
  {
    id: 'everything',
    name: 'Everything',
    description: 'Complete deployment with all packages',
    icon: '🍒',
  },
]

const recommendedPackages: Record<NodeType, string[]> = {
  inference: ['core', 'python', 'pytorch', 'vllm', 'sglang', 'aphrodite'],
  training: ['core', 'python', 'pytorch', 'deepspeed', 'llamafactory'],
  'art-video': ['core', 'python', 'pytorch', 'comfyui', 'mochi'],
  development: ['core', 'python', 'pytorch', 'comfyui', 'vllm', 'sglang'],
  multimodal: ['core', 'python', 'pytorch', 'llava-34b', 'whisper-large', 'ollama'],
  everything: ['core', 'python', 'pytorch', 'deepspeed', 'vllm', 'sglang', 'aphrodite', 'comfyui', 'mochi', 'ollama', 'llamafactory'],
}

export default function WizardFlow() {
  const [currentStep, setCurrentStep] = useState(0)
  const [selectedNodeType, setSelectedNodeType] = useState<NodeType | null>(null)
  const [selectedPackages, setSelectedPackages] = useState<Set<string>>(new Set())
  const [allPackages, setAllPackages] = useState<Package[]>([])
  const [instanceName, setInstanceName] = useState('')
  const [gpuCount, setGpuCount] = useState(1)

  // Load packages when needed
  const loadPackages = async () => {
    if (allPackages.length === 0) {
      try {
        const packages = await invoke<Package[]>('get_packages_command')
        setAllPackages(packages)
      } catch (error) {

      }
    }
  }

  const handleNodeTypeSelect = (type: NodeType) => {
    setSelectedNodeType(type)
    const recommended = new Set(recommendedPackages[type])
    setSelectedPackages(recommended)
    setCurrentStep(1)
    loadPackages()
  }

  const togglePackage = (id: string) => {
    const newSelected = new Set(selectedPackages)
    if (newSelected.has(id)) {
      newSelected.delete(id)
    } else {
      newSelected.add(id)
    }
    setSelectedPackages(newSelected)
  }

  const handleNext = () => {
    if (currentStep < steps.length - 1) {
      setCurrentStep(currentStep + 1)
    }
  }

  const handleBack = () => {
    if (currentStep > 0) {
      setCurrentStep(currentStep - 1)
    }
  }

  const handleDeploy = async () => {
    try {
      const packageArray = Array.from(selectedPackages)
      await invoke<string>('install_packages', { packageIds: packageArray })

      // TODO: Navigate to progress monitor
    } catch (error) {

    }
  }

  return (
    <div className="max-w-6xl mx-auto">
      {/* Progress Steps */}
      <div className="mb-8">
        <div className="flex items-center justify-between">
          {steps.map((step, idx) => (
            <div key={idx} className="flex items-center flex-1">
              <div className="flex flex-col items-center flex-1">
                <div
                  className={`w-10 h-10 rounded-full flex items-center justify-center font-bold transition-all ${
                    idx <= currentStep
                      ? 'bg-gradient-to-r from-sakura-500 to-neon-400 text-white'
                      : 'bg-gray-700 text-gray-400'
                  }`}
                >
                  {idx + 1}
                </div>
                <div className="mt-2 text-center">
                  <div className="text-sm font-semibold">{step.title}</div>
                  <div className="text-xs text-gray-400">{step.description}</div>
                </div>
              </div>
              {idx < steps.length - 1 && (
                <div
                  className={`h-1 flex-1 mx-4 transition-all ${
                    idx < currentStep ? 'bg-gradient-to-r from-sakura-500 to-neon-400' : 'bg-gray-700'
                  }`}
                />
              )}
            </div>
          ))}
        </div>
      </div>

      {/* Step Content */}
      <div className="bg-gray-900/50 backdrop-blur-sm rounded-xl p-8 border border-gray-800">
        {/* Step 0: Node Type Selection */}
        {currentStep === 0 && (
          <div className="space-y-6">
            <h2 className="text-3xl font-bold text-center bg-gradient-to-r from-sakura-500 to-electric-500 bg-clip-text text-transparent">
              Choose Your Node Type
            </h2>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mt-8">
              {nodeTypes.map((type) => (
                <button
                  key={type.id}
                  onClick={() => handleNodeTypeSelect(type.id)}
                  className="p-6 rounded-lg border-2 border-gray-700 hover:border-sakura-500 transition-all text-left group hover:bg-sakura-500/5"
                >
                  <div className="text-4xl mb-3">{type.icon}</div>
                  <h3 className="text-xl font-bold mb-2 group-hover:text-sakura-500 transition-colors">
                    {type.name}
                  </h3>
                  <p className="text-gray-400 text-sm">{type.description}</p>
                </button>
              ))}
            </div>
          </div>
        )}

        {/* Step 1: Package Selection */}
        {currentStep === 1 && (
          <div className="space-y-6">
            <div>
              <h2 className="text-2xl font-bold mb-2">Select Packages</h2>
              <p className="text-gray-400">
                Recommended packages for{' '}
                <span className="text-sakura-500 font-semibold">
                  {nodeTypes.find((n) => n.id === selectedNodeType)?.name}
                </span>{' '}
                are pre-selected
              </p>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-3 max-h-96 overflow-y-auto">
              {allPackages.map((pkg) => (
                <button
                  key={pkg.id}
                  onClick={() => togglePackage(pkg.id)}
                  className={`p-4 rounded-lg border-2 transition-all text-left ${
                    selectedPackages.has(pkg.id)
                      ? 'border-sakura-500 bg-sakura-500/10'
                      : 'border-gray-700 hover:border-electric-500'
                  }`}
                >
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <h3 className="font-bold">{pkg.name}</h3>
                      <p className="text-xs text-gray-400 mt-1">{pkg.description}</p>
                      <div className="flex gap-2 mt-2">
                        <span className="text-xs bg-gray-800 px-2 py-1 rounded">{pkg.size}</span>
                        <span className="text-xs bg-gray-800 px-2 py-1 rounded">
                          ~{pkg.estimatedTime}min
                        </span>
                      </div>
                    </div>
                    <div
                      className={`w-5 h-5 rounded border-2 flex items-center justify-center ${
                        selectedPackages.has(pkg.id)
                          ? 'border-sakura-500 bg-sakura-500'
                          : 'border-gray-600'
                      }`}
                    >
                      {selectedPackages.has(pkg.id) && <span className="text-white text-xs">✓</span>}
                    </div>
                  </div>
                </button>
              ))}
            </div>
          </div>
        )}

        {/* Step 2: Configuration */}
        {currentStep === 2 && (
          <div className="space-y-6">
            <h2 className="text-2xl font-bold">Instance Configuration</h2>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-semibold mb-2">Instance Name</label>
                <input
                  type="text"
                  value={instanceName}
                  onChange={(e) => setInstanceName(e.target.value)}
                  placeholder="my-lambda-instance"
                  className="w-full px-4 py-2 rounded-lg bg-gray-800 border border-gray-700 focus:border-sakura-500 focus:outline-none"
                />
              </div>
              <div>
                <label className="block text-sm font-semibold mb-2">GPU Count</label>
                <select
                  value={gpuCount}
                  onChange={(e) => setGpuCount(Number(e.target.value))}
                  className="w-full px-4 py-2 rounded-lg bg-gray-800 border border-gray-700 focus:border-sakura-500 focus:outline-none"
                >
                  <option value={1}>1 GPU (GH200)</option>
                  <option value={2}>2 GPUs (GH200)</option>
                  <option value={4}>4 GPUs (GH200)</option>
                  <option value={8}>8 GPUs (GH200)</option>
                </select>
              </div>
            </div>
          </div>
        )}

        {/* Step 3: Review */}
        {currentStep === 3 && (
          <div className="space-y-6">
            <h2 className="text-2xl font-bold">Review & Deploy</h2>
            <div className="space-y-4">
              <div className="bg-gray-800/50 rounded-lg p-4">
                <h3 className="text-sm font-semibold text-gray-400 mb-2">Node Type</h3>
                <p className="text-lg font-bold">
                  {nodeTypes.find((n) => n.id === selectedNodeType)?.name}
                </p>
              </div>
              <div className="bg-gray-800/50 rounded-lg p-4">
                <h3 className="text-sm font-semibold text-gray-400 mb-2">Instance Configuration</h3>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <p className="text-xs text-gray-400">Name</p>
                    <p className="font-semibold">{instanceName || 'Not specified'}</p>
                  </div>
                  <div>
                    <p className="text-xs text-gray-400">GPUs</p>
                    <p className="font-semibold">{gpuCount} GH200</p>
                  </div>
                </div>
              </div>
              <div className="bg-gray-800/50 rounded-lg p-4">
                <h3 className="text-sm font-semibold text-gray-400 mb-2">
                  Selected Packages ({selectedPackages.size})
                </h3>
                <div className="flex flex-wrap gap-2 mt-2">
                  {Array.from(selectedPackages).map((pkgId) => {
                    const pkg = allPackages.find((p) => p.id === pkgId)
                    return (
                      <span
                        key={pkgId}
                        className="px-3 py-1 bg-sakura-500/20 border border-sakura-500 rounded-full text-sm"
                      >
                        {pkg?.name || pkgId}
                      </span>
                    )
                  })}
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Navigation Buttons */}
        <div className="flex justify-between mt-8 pt-6 border-t border-gray-800">
          <button
            onClick={handleBack}
            disabled={currentStep === 0}
            className="px-6 py-2 rounded-lg border border-gray-700 hover:border-gray-600 disabled:opacity-50 disabled:cursor-not-allowed transition-all"
          >
            Back
          </button>
          <div className="flex gap-3">
            {currentStep < steps.length - 1 ? (
              <button
                onClick={handleNext}
                disabled={currentStep === 0 && !selectedNodeType}
                className="px-6 py-2 rounded-lg bg-gradient-to-r from-sakura-500 to-neon-400 hover:from-sakura-600 hover:to-neon-500 disabled:opacity-50 disabled:cursor-not-allowed transition-all font-semibold"
              >
                Next
              </button>
            ) : (
              <button
                onClick={handleDeploy}
                className="px-8 py-2 rounded-lg bg-gradient-to-r from-mint-500 to-electric-500 hover:from-mint-600 hover:to-electric-600 transition-all font-bold"
              >
                🚀 Deploy
              </button>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
