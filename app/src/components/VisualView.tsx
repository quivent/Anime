import { useEffect, useState } from 'react'
import { invoke } from '@tauri-apps/api/core'
import { listen } from '@tauri-apps/api/event'
import { ToastContainer } from './Toast'
import { useToast } from '../hooks/useToast'

// Types
interface ModelCard {
  id: string
  name: string
  model_type: string
  description: string
  capabilities: string[]
  max_duration: number
  recommended_fps: number[]
  recommended_resolutions: string[]
  installed: boolean
  size_gb: number
}

interface StylePreset {
  id: string
  name: string
  description: string
  model_type: string
  positive_prompt_suffix: string
  negative_prompt: string
  recommended_settings: VideoSettings
}

interface VideoSettings {
  duration: number
  fps: number
  resolution: string
  motion_scale: number
  style_preset: string | null
  seed: number | null
  camera_motion: CameraMotion
  guidance_scale: number
  num_inference_steps: number
}

interface CameraMotion {
  pan_x: number
  pan_y: number
  zoom: number
  rotate: number
}

interface GenerationJob {
  id: string
  model: string
  prompt: string
  negative_prompt: string | null
  input_image: string | null
  settings: VideoSettings
  status: 'queued' | 'preparing' | 'generating' | 'postprocessing' | 'completed' | 'failed' | 'cancelled'
  progress: number
  current_frame: number
  total_frames: number
  output_path: string | null
  error_message: string | null
  created_at: string
  completed_at: string | null
  estimated_time_remaining: number | null
}

type Tab = 'models' | 'generate' | 'queue' | 'gallery'

export default function VisualView() {
  const [activeTab, setActiveTab] = useState<Tab>('models')
  const [models, setModels] = useState<ModelCard[]>([])
  const [stylePresets, setStylePresets] = useState<StylePreset[]>([])
  const [jobs, setJobs] = useState<GenerationJob[]>([])
  const [selectedModel, setSelectedModel] = useState<string>('')
  const { toasts, showError, removeToast } = useToast()

  // Generation form state
  const [prompt, setPrompt] = useState('')
  const [negativePrompt, setNegativePrompt] = useState('')
  const [selectedPreset, setSelectedPreset] = useState<string>('')
  const [settings, setSettings] = useState<VideoSettings>({
    duration: 2,
    fps: 8,
    resolution: '512x512',
    motion_scale: 1.27,
    style_preset: null,
    seed: null,
    camera_motion: {
      pan_x: 0,
      pan_y: 0,
      zoom: 1,
      rotate: 0,
    },
    guidance_scale: 7.5,
    num_inference_steps: 25,
  })

  // Load initial data
  useEffect(() => {
    loadModels()
    loadStylePresets()
    loadJobs()

    // Listen for job updates
    const unlisten = listen<GenerationJob>('generation_job_update', (event) => {
      const updatedJob = event.payload
      setJobs((prevJobs) => {
        const index = prevJobs.findIndex((j) => j.id === updatedJob.id)
        if (index >= 0) {
          const newJobs = [...prevJobs]
          newJobs[index] = updatedJob
          return newJobs
        } else {
          return [updatedJob, ...prevJobs]
        }
      })
    })

    return () => {
      unlisten.then((fn) => fn())
    }
  }, [])

  const loadModels = async () => {
    try {
      const data = await invoke<ModelCard[]>('get_animation_models')
      setModels(data)
      if (data.length > 0) {
        setSelectedModel(data[0].id)
      }
    } catch (error) {

    }
  }

  const loadStylePresets = async () => {
    try {
      const data = await invoke<StylePreset[]>('get_animation_style_presets')
      setStylePresets(data)
    } catch (error) {

    }
  }

  const loadJobs = async () => {
    try {
      const data = await invoke<GenerationJob[]>('get_generation_jobs')
      setJobs(data)
    } catch (error) {

    }
  }

  const handlePresetChange = (presetId: string) => {
    setSelectedPreset(presetId)
    const preset = stylePresets.find((p) => p.id === presetId)
    if (preset) {
      setNegativePrompt(preset.negative_prompt)
      setSettings(preset.recommended_settings)
    }
  }

  const handleSubmitGeneration = async () => {
    if (!prompt.trim() || !selectedModel) {
      showError('Please enter a prompt and select a model')
      return
    }

    try {
      await invoke<string>('submit_generation_job', {
        model: selectedModel,
        prompt: prompt.trim(),
        negativePrompt: negativePrompt.trim() || null,
        inputImage: null,
        settings,
      })

      setActiveTab('queue')
      loadJobs()
    } catch (error) {
      showError('Failed to submit generation job: ' + error)
    }
  }

  const handleCancelJob = async (jobId: string) => {
    try {
      await invoke('cancel_generation_job', { jobId })
      loadJobs()
    } catch (error) {

    }
  }

  const handleRetryJob = async (jobId: string) => {
    try {
      await invoke<string>('retry_generation_job', { jobId })
      loadJobs()
    } catch (error) {

    }
  }

  const handleDeleteJob = async (jobId: string) => {
    try {
      await invoke('delete_generation_job', { jobId })
      loadJobs()
    } catch (error) {

    }
  }

  const formatTime = (seconds: number | null): string => {
    if (seconds === null || seconds === 0) return 'Done'
    if (seconds < 60) return `${seconds}s`
    const mins = Math.floor(seconds / 60)
    const secs = seconds % 60
    return `${mins}m ${secs}s`
  }

  const getStatusColor = (status: string): string => {
    switch (status) {
      case 'completed':
        return 'text-mint-400 bg-mint-500/10'
      case 'generating':
        return 'text-electric-400 bg-electric-500/10 animate-pulse'
      case 'failed':
        return 'text-sunset-400 bg-sunset-500/10'
      case 'cancelled':
        return 'text-gray-400 bg-gray-500/10'
      default:
        return 'text-neon-400 bg-neon-500/10'
    }
  }

  return (
    <div className="h-full flex flex-col overflow-hidden">
      {/* Header */}
      <div className="px-6 py-4 border-b border-gray-800">
        <h2 className="text-3xl font-bold mb-2 sakura-gradient bg-clip-text text-transparent">
          Visual & Animation Studio
        </h2>
        <p className="text-gray-400">AI-powered video generation with AnimateDiff, SVD, and more</p>
      </div>

      {/* Tab Navigation */}
      <div className="px-6 py-3 border-b border-gray-800 bg-gray-900/50">
        <div className="flex gap-2">
          <button
            onClick={() => setActiveTab('models')}
            className={`px-4 py-2 rounded-lg font-medium text-sm transition-all ${
              activeTab === 'models'
                ? 'bg-sakura-500 text-white'
                : 'bg-gray-800 text-gray-400 hover:bg-gray-700'
            }`}
          >
            Models
          </button>
          <button
            onClick={() => setActiveTab('generate')}
            className={`px-4 py-2 rounded-lg font-medium text-sm transition-all ${
              activeTab === 'generate'
                ? 'bg-sakura-500 text-white'
                : 'bg-gray-800 text-gray-400 hover:bg-gray-700'
            }`}
          >
            Generate
          </button>
          <button
            onClick={() => setActiveTab('queue')}
            className={`px-4 py-2 rounded-lg font-medium text-sm transition-all relative ${
              activeTab === 'queue'
                ? 'bg-sakura-500 text-white'
                : 'bg-gray-800 text-gray-400 hover:bg-gray-700'
            }`}
          >
            Queue
            {jobs.filter((j) => j.status === 'generating' || j.status === 'preparing').length > 0 && (
              <span className="absolute -top-1 -right-1 w-3 h-3 bg-electric-400 rounded-full animate-pulse" />
            )}
          </button>
          <button
            onClick={() => setActiveTab('gallery')}
            className={`px-4 py-2 rounded-lg font-medium text-sm transition-all ${
              activeTab === 'gallery'
                ? 'bg-sakura-500 text-white'
                : 'bg-gray-800 text-gray-400 hover:bg-gray-700'
            }`}
          >
            Gallery
          </button>
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-y-auto">
        {activeTab === 'models' && (
          <div className="p-6">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {models.map((model) => (
                <div
                  key={model.id}
                  className={`p-5 rounded-lg border-2 transition-all ${
                    selectedModel === model.id
                      ? 'border-sakura-500 bg-sakura-500/10'
                      : 'border-gray-800 bg-gray-900/50 hover:border-gray-700'
                  }`}
                >
                  <div className="flex items-start justify-between mb-3">
                    <div className="flex-1">
                      <h3 className="font-bold text-lg mb-1">{model.name}</h3>
                      <p className="text-sm text-gray-400">{model.description}</p>
                    </div>
                    {model.installed && (
                      <span className="px-2 py-1 bg-mint-500/20 text-mint-400 text-xs rounded">
                        Installed
                      </span>
                    )}
                  </div>

                  <div className="mb-3">
                    <div className="text-xs font-semibold text-gray-500 mb-1">Capabilities:</div>
                    <div className="flex flex-wrap gap-1">
                      {model.capabilities.map((cap) => (
                        <span
                          key={cap}
                          className="px-2 py-0.5 bg-gray-800 rounded text-xs text-gray-400"
                        >
                          {cap}
                        </span>
                      ))}
                    </div>
                  </div>

                  <div className="flex items-center gap-4 text-xs text-gray-500 mb-3">
                    <span>Max: {model.max_duration}s</span>
                    <span>FPS: {model.recommended_fps.join(', ')}</span>
                    <span>Size: {model.size_gb.toFixed(1)} GB</span>
                  </div>

                  <div className="flex gap-2">
                    {model.installed ? (
                      <button
                        onClick={() => setSelectedModel(model.id)}
                        className="flex-1 px-4 py-2 rounded-lg bg-sakura-500 text-white hover:bg-sakura-600 transition-colors text-sm font-medium"
                      >
                        Use Model
                      </button>
                    ) : (
                      <button className="flex-1 px-4 py-2 rounded-lg bg-neon-500 text-white hover:bg-neon-600 transition-colors text-sm font-medium">
                        Install
                      </button>
                    )}
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {activeTab === 'generate' && (
          <div className="p-6">
            <div className="max-w-4xl mx-auto">
              <div className="grid grid-cols-3 gap-6">
                {/* Main Generation Form */}
                <div className="col-span-2 space-y-4">
                  {/* Model Selection */}
                  <div>
                    <label className="block text-sm font-semibold mb-2">Model</label>
                    <select
                      value={selectedModel}
                      onChange={(e) => setSelectedModel(e.target.value)}
                      className="w-full px-4 py-2 bg-gray-800 border border-gray-700 rounded-lg focus:border-sakura-500 focus:outline-none"
                    >
                      {models.map((model) => (
                        <option key={model.id} value={model.id}>
                          {model.name} {!model.installed && '(Not Installed)'}
                        </option>
                      ))}
                    </select>
                  </div>

                  {/* Prompt */}
                  <div>
                    <label className="block text-sm font-semibold mb-2">Prompt</label>
                    <textarea
                      value={prompt}
                      onChange={(e) => setPrompt(e.target.value)}
                      placeholder="Describe the video you want to generate..."
                      className="w-full px-4 py-3 bg-gray-800 border border-gray-700 rounded-lg focus:border-sakura-500 focus:outline-none resize-none"
                      rows={4}
                    />
                  </div>

                  {/* Negative Prompt */}
                  <div>
                    <label className="block text-sm font-semibold mb-2">Negative Prompt</label>
                    <textarea
                      value={negativePrompt}
                      onChange={(e) => setNegativePrompt(e.target.value)}
                      placeholder="What to avoid in the video..."
                      className="w-full px-4 py-3 bg-gray-800 border border-gray-700 rounded-lg focus:border-sakura-500 focus:outline-none resize-none"
                      rows={2}
                    />
                  </div>

                  {/* Video Settings */}
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <label className="block text-sm font-semibold mb-2">
                        Duration (seconds): {settings.duration}
                      </label>
                      <input
                        type="range"
                        min="1"
                        max="8"
                        value={settings.duration}
                        onChange={(e) =>
                          setSettings({ ...settings, duration: parseInt(e.target.value) })
                        }
                        className="w-full"
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-semibold mb-2">FPS: {settings.fps}</label>
                      <input
                        type="range"
                        min="6"
                        max="24"
                        step="2"
                        value={settings.fps}
                        onChange={(e) => setSettings({ ...settings, fps: parseInt(e.target.value) })}
                        className="w-full"
                      />
                    </div>
                  </div>

                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <label className="block text-sm font-semibold mb-2">Resolution</label>
                      <select
                        value={settings.resolution}
                        onChange={(e) => setSettings({ ...settings, resolution: e.target.value })}
                        className="w-full px-4 py-2 bg-gray-800 border border-gray-700 rounded-lg focus:border-sakura-500 focus:outline-none"
                      >
                        <option value="512x512">512x512</option>
                        <option value="768x768">768x768</option>
                        <option value="1024x1024">1024x1024</option>
                        <option value="512x768">512x768 (Portrait)</option>
                        <option value="768x512">768x512 (Landscape)</option>
                      </select>
                    </div>
                    <div>
                      <label className="block text-sm font-semibold mb-2">
                        Motion Scale: {settings.motion_scale.toFixed(2)}
                      </label>
                      <input
                        type="range"
                        min="0"
                        max="2"
                        step="0.05"
                        value={settings.motion_scale}
                        onChange={(e) =>
                          setSettings({ ...settings, motion_scale: parseFloat(e.target.value) })
                        }
                        className="w-full"
                      />
                    </div>
                  </div>

                  {/* Camera Motion */}
                  <div>
                    <label className="block text-sm font-semibold mb-2">Camera Motion</label>
                    <div className="grid grid-cols-2 gap-3">
                      <div>
                        <label className="text-xs text-gray-400">Pan X: {settings.camera_motion.pan_x.toFixed(2)}</label>
                        <input
                          type="range"
                          min="-1"
                          max="1"
                          step="0.1"
                          value={settings.camera_motion.pan_x}
                          onChange={(e) =>
                            setSettings({
                              ...settings,
                              camera_motion: {
                                ...settings.camera_motion,
                                pan_x: parseFloat(e.target.value),
                              },
                            })
                          }
                          className="w-full"
                        />
                      </div>
                      <div>
                        <label className="text-xs text-gray-400">Pan Y: {settings.camera_motion.pan_y.toFixed(2)}</label>
                        <input
                          type="range"
                          min="-1"
                          max="1"
                          step="0.1"
                          value={settings.camera_motion.pan_y}
                          onChange={(e) =>
                            setSettings({
                              ...settings,
                              camera_motion: {
                                ...settings.camera_motion,
                                pan_y: parseFloat(e.target.value),
                              },
                            })
                          }
                          className="w-full"
                        />
                      </div>
                      <div>
                        <label className="text-xs text-gray-400">Zoom: {settings.camera_motion.zoom.toFixed(2)}</label>
                        <input
                          type="range"
                          min="0.5"
                          max="2"
                          step="0.1"
                          value={settings.camera_motion.zoom}
                          onChange={(e) =>
                            setSettings({
                              ...settings,
                              camera_motion: {
                                ...settings.camera_motion,
                                zoom: parseFloat(e.target.value),
                              },
                            })
                          }
                          className="w-full"
                        />
                      </div>
                      <div>
                        <label className="text-xs text-gray-400">Rotate: {settings.camera_motion.rotate.toFixed(0)}°</label>
                        <input
                          type="range"
                          min="-180"
                          max="180"
                          step="15"
                          value={settings.camera_motion.rotate}
                          onChange={(e) =>
                            setSettings({
                              ...settings,
                              camera_motion: {
                                ...settings.camera_motion,
                                rotate: parseFloat(e.target.value),
                              },
                            })
                          }
                          className="w-full"
                        />
                      </div>
                    </div>
                  </div>

                  {/* Advanced Settings */}
                  <details className="bg-gray-800/50 rounded-lg">
                    <summary className="px-4 py-2 cursor-pointer font-semibold text-sm">
                      Advanced Settings
                    </summary>
                    <div className="px-4 pb-4 space-y-3">
                      <div>
                        <label className="block text-sm mb-2">
                          Guidance Scale: {settings.guidance_scale.toFixed(1)}
                        </label>
                        <input
                          type="range"
                          min="1"
                          max="20"
                          step="0.5"
                          value={settings.guidance_scale}
                          onChange={(e) =>
                            setSettings({ ...settings, guidance_scale: parseFloat(e.target.value) })
                          }
                          className="w-full"
                        />
                      </div>
                      <div>
                        <label className="block text-sm mb-2">
                          Inference Steps: {settings.num_inference_steps}
                        </label>
                        <input
                          type="range"
                          min="20"
                          max="50"
                          value={settings.num_inference_steps}
                          onChange={(e) =>
                            setSettings({ ...settings, num_inference_steps: parseInt(e.target.value) })
                          }
                          className="w-full"
                        />
                      </div>
                      <div>
                        <label className="block text-sm mb-2">Seed (optional)</label>
                        <input
                          type="number"
                          value={settings.seed || ''}
                          onChange={(e) =>
                            setSettings({
                              ...settings,
                              seed: e.target.value ? parseInt(e.target.value) : null,
                            })
                          }
                          placeholder="Random"
                          className="w-full px-4 py-2 bg-gray-800 border border-gray-700 rounded-lg focus:border-sakura-500 focus:outline-none"
                        />
                      </div>
                    </div>
                  </details>

                  {/* Submit Button */}
                  <button
                    onClick={handleSubmitGeneration}
                    disabled={!prompt.trim() || !models.find((m) => m.id === selectedModel)?.installed}
                    className="w-full px-6 py-3 bg-sakura-500 text-white rounded-lg hover:bg-sakura-600 transition-colors font-semibold disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    Generate Video
                  </button>
                </div>

                {/* Style Presets Sidebar */}
                <div className="space-y-4">
                  <div>
                    <h3 className="text-sm font-semibold mb-3">Style Presets</h3>
                    <div className="space-y-2">
                      {stylePresets.map((preset) => (
                        <button
                          key={preset.id}
                          onClick={() => handlePresetChange(preset.id)}
                          className={`w-full p-3 rounded-lg border-2 text-left transition-all ${
                            selectedPreset === preset.id
                              ? 'border-electric-500 bg-electric-500/10'
                              : 'border-gray-800 bg-gray-900/50 hover:border-gray-700'
                          }`}
                        >
                          <div className="font-semibold text-sm mb-1">{preset.name}</div>
                          <div className="text-xs text-gray-400">{preset.description}</div>
                        </button>
                      ))}
                    </div>
                  </div>

                  <div className="p-4 bg-gray-800/50 rounded-lg">
                    <div className="text-xs font-semibold text-gray-400 mb-2">Input Image (optional)</div>
                    <div className="border-2 border-dashed border-gray-700 rounded-lg p-4 text-center">
                      <div className="text-gray-500 text-sm">Click to upload</div>
                      <div className="text-xs text-gray-600 mt-1">or drag & drop</div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        )}

        {activeTab === 'queue' && (
          <div className="p-6">
            <div className="space-y-3">
              {jobs.length === 0 ? (
                <div className="text-center py-12">
                  <div className="text-gray-500 text-lg mb-2">No generation jobs yet</div>
                  <button
                    onClick={() => setActiveTab('generate')}
                    className="px-4 py-2 bg-sakura-500 text-white rounded-lg hover:bg-sakura-600 transition-colors"
                  >
                    Create Your First Video
                  </button>
                </div>
              ) : (
                jobs.map((job) => (
                  <div
                    key={job.id}
                    className="p-4 bg-gray-900/50 border border-gray-800 rounded-lg hover:border-gray-700 transition-all"
                  >
                    <div className="flex items-start justify-between mb-3">
                      <div className="flex-1">
                        <div className="flex items-center gap-2 mb-1">
                          <span className={`px-2 py-0.5 rounded text-xs font-medium ${getStatusColor(job.status)}`}>
                            {job.status}
                          </span>
                          <span className="text-xs text-gray-500">
                            {models.find((m) => m.id === job.model)?.name || job.model}
                          </span>
                        </div>
                        <div className="text-sm font-medium mb-1 line-clamp-2">{job.prompt}</div>
                        <div className="text-xs text-gray-500">
                          {job.total_frames} frames ({job.settings.duration}s @ {job.settings.fps} FPS)
                        </div>
                      </div>
                    </div>

                    {/* Progress Bar */}
                    {job.status !== 'completed' && job.status !== 'failed' && job.status !== 'cancelled' && (
                      <div className="mb-3">
                        <div className="flex items-center justify-between text-xs text-gray-500 mb-1">
                          <span>
                            Frame {job.current_frame} / {job.total_frames}
                          </span>
                          <span>{formatTime(job.estimated_time_remaining)}</span>
                        </div>
                        <div className="h-2 bg-gray-800 rounded-full overflow-hidden">
                          <div
                            className="h-full bg-gradient-to-r from-sakura-500 to-electric-500 transition-all duration-300"
                            style={{ width: `${job.progress}%` }}
                          />
                        </div>
                      </div>
                    )}

                    {/* Actions */}
                    <div className="flex gap-2">
                      {job.status === 'completed' && job.output_path && (
                        <>
                          <button className="px-3 py-1.5 bg-mint-500 text-white rounded text-sm hover:bg-mint-600 transition-colors">
                            View
                          </button>
                          <button className="px-3 py-1.5 bg-electric-500 text-white rounded text-sm hover:bg-electric-600 transition-colors">
                            Download
                          </button>
                        </>
                      )}
                      {(job.status === 'generating' || job.status === 'preparing') && (
                        <button
                          onClick={() => handleCancelJob(job.id)}
                          className="px-3 py-1.5 bg-gray-700 text-gray-300 rounded text-sm hover:bg-gray-600 transition-colors"
                        >
                          Cancel
                        </button>
                      )}
                      {job.status === 'failed' && (
                        <button
                          onClick={() => handleRetryJob(job.id)}
                          className="px-3 py-1.5 bg-neon-500 text-white rounded text-sm hover:bg-neon-600 transition-colors"
                        >
                          Retry
                        </button>
                      )}
                      <button
                        onClick={() => handleDeleteJob(job.id)}
                        className="px-3 py-1.5 bg-gray-800 text-gray-400 rounded text-sm hover:bg-gray-700 transition-colors ml-auto"
                      >
                        Delete
                      </button>
                    </div>

                    {job.error_message && (
                      <div className="mt-2 p-2 bg-sunset-500/10 border border-sunset-500/30 rounded text-xs text-sunset-400">
                        {job.error_message}
                      </div>
                    )}
                  </div>
                ))
              )}
            </div>
          </div>
        )}

        {activeTab === 'gallery' && (
          <div className="p-6">
            <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
              {jobs
                .filter((j) => j.status === 'completed' && j.output_path)
                .map((job) => (
                  <div
                    key={job.id}
                    className="aspect-square bg-gray-900/50 border border-gray-800 rounded-lg overflow-hidden hover:border-sakura-500 transition-all cursor-pointer group"
                  >
                    <div className="w-full h-full flex items-center justify-center bg-gradient-to-br from-sakura-500/20 to-electric-500/20">
                      <div className="text-6xl opacity-50 group-hover:opacity-100 transition-opacity">
                        🎬
                      </div>
                    </div>
                    <div className="p-2 bg-gray-900">
                      <div className="text-xs font-medium line-clamp-1">{job.prompt}</div>
                      <div className="text-xs text-gray-500">
                        {job.settings.duration}s @ {job.settings.fps}fps
                      </div>
                    </div>
                  </div>
                ))}
            </div>

            {jobs.filter((j) => j.status === 'completed').length === 0 && (
              <div className="text-center py-12">
                <div className="text-gray-500 text-lg">No completed videos yet</div>
                <p className="text-gray-600 text-sm mt-2">Your generated videos will appear here</p>
              </div>
            )}
          </div>
        )}
      </div>

      {/* Toast Notifications */}
      <ToastContainer toasts={toasts} onClose={removeToast} />
    </div>
  )
}
