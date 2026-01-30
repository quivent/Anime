import { useState, useRef } from 'react'
import { invoke } from '@tauri-apps/api/core'
import { open } from '@tauri-apps/plugin-dialog'
import type {
  StoryboardProject,
  StoryboardScene,
  StoryboardPanel,
  ScriptParseResult,
  ShotSuggestion,
} from '../types/creative'

export default function StoryboardsView() {
  const [projects, setProjects] = useState<StoryboardProject[]>([])
  const [currentProject, setCurrentProject] = useState<StoryboardProject | null>(null)
  const [script, setScript] = useState('')
  const [isParsing, setIsParsing] = useState(false)
  const [selectedScene, setSelectedScene] = useState<StoryboardScene | null>(null)
  const [showNewProjectDialog, setShowNewProjectDialog] = useState(false)
  const [newProjectName, setNewProjectName] = useState('')
  const [generatingPanelId, setGeneratingPanelId] = useState<string | null>(null)
  const canvasRef = useRef<HTMLCanvasElement>(null)

  const createProject = async () => {
    if (!newProjectName.trim()) return

    try {
      const project = await invoke<StoryboardProject>('create_storyboard_project', {
        name: newProjectName,
      })
      setProjects([...projects, project])
      setCurrentProject(project)
      setNewProjectName('')
      setShowNewProjectDialog(false)
    } catch (error) {

    }
  }

  const loadScriptFromFile = async () => {
    try {
      const selected = await open({
        multiple: false,
        filters: [{
          name: 'Script Files',
          extensions: ['txt', 'fountain', 'pdf']
        }]
      })

      if (selected && typeof selected === 'string') {
        const content = await invoke<string>('read_file', { path: selected })
        setScript(content)
      }
    } catch (error) {

    }
  }

  const parseScript = async () => {
    if (!script.trim() || !currentProject) return

    setIsParsing(true)

    try {
      const result = await invoke<ScriptParseResult>('parse_script', {
        script,
      })

      const updatedProject: StoryboardProject = {
        ...currentProject,
        script,
        scenes: result.scenes,
      }

      setCurrentProject(updatedProject)
      setProjects(projects.map(p => p.id === updatedProject.id ? updatedProject : p))
    } catch (error) {

    } finally {
      setIsParsing(false)
    }
  }

  const generateShotSuggestions = async (sceneId: string) => {
    try {
      const suggestions = await invoke<ShotSuggestion[]>('generate_shot_suggestions', {
        projectId: currentProject?.id,
        sceneId,
      })

      // Create panels from suggestions
      const newPanels: StoryboardPanel[] = suggestions.map((sugg, idx) => ({
        id: `${sceneId}-panel-${idx}`,
        scene_id: sceneId,
        panel_number: idx + 1,
        shot_type: sugg.type,
        composition: sugg.composition,
        description: sugg.description,
        notes: `Camera: ${sugg.camera_angle}, Lighting: ${sugg.lighting}`,
      }))

      if (currentProject) {
        const updatedProject: StoryboardProject = {
          ...currentProject,
          panels: [...currentProject.panels, ...newPanels],
        }
        setCurrentProject(updatedProject)
        setProjects(projects.map(p => p.id === updatedProject.id ? updatedProject : p))
      }
    } catch (error) {

    }
  }

  const generatePanelImage = async (panel: StoryboardPanel) => {
    setGeneratingPanelId(panel.id)

    try {
      const imageUrl = await invoke<string>('generate_storyboard_image', {
        description: panel.description,
        shotType: panel.shot_type,
        composition: panel.composition,
      })

      if (currentProject) {
        const updatedPanels = currentProject.panels.map(p =>
          p.id === panel.id ? { ...p, image_url: imageUrl } : p
        )
        const updatedProject: StoryboardProject = {
          ...currentProject,
          panels: updatedPanels,
        }
        setCurrentProject(updatedProject)
        setProjects(projects.map(p => p.id === updatedProject.id ? updatedProject : p))
      }
    } catch (error) {

    } finally {
      setGeneratingPanelId(null)
    }
  }

  const exportStoryboard = async () => {
    if (!currentProject) return

    try {
      await invoke('export_storyboard', {
        projectId: currentProject.id,
        format: 'pdf',
      })
    } catch (error) {

    }
  }

  const deletePanel = (panelId: string) => {
    if (!currentProject) return

    const updatedProject: StoryboardProject = {
      ...currentProject,
      panels: currentProject.panels.filter(p => p.id !== panelId),
    }
    setCurrentProject(updatedProject)
    setProjects(projects.map(p => p.id === updatedProject.id ? updatedProject : p))
  }

  const scenePanels = currentProject && selectedScene
    ? currentProject.panels.filter(p => p.scene_id === selectedScene.id)
    : []

  return (
    <div className="h-full flex overflow-hidden">
      {/* Mock Feature Warning Banner */}
      <div className="absolute top-0 left-0 right-0 bg-sunset-500/20 border-b border-sunset-500/50 px-4 py-2 z-10">
        <div className="flex items-center gap-2 text-sm">
          <span className="text-sunset-400">⚠️</span>
          <span className="text-sunset-300 font-medium">Development Mode:</span>
          <span className="text-gray-300">Script parsing, shot suggestions, and image generation are mock features</span>
        </div>
      </div>

      {/* Sidebar - Projects & Scenes */}
      <div className="w-64 border-r border-gray-800 bg-gray-900/50 flex flex-col mt-10">
        <div className="p-4 border-b border-gray-800">
          <button
            onClick={() => setShowNewProjectDialog(true)}
            className="w-full px-4 py-2 bg-electric-500/20 hover:bg-electric-500/30 border border-electric-500/50 rounded-lg text-electric-400 font-medium transition-all"
          >
            + New Project
          </button>
        </div>

        {currentProject && (
          <>
            <div className="p-4 border-b border-gray-800">
              <h3 className="font-bold text-gray-200 mb-2">{currentProject.name}</h3>
              <div className="text-xs text-gray-500">
                {currentProject.scenes.length} scenes • {currentProject.panels.length} panels
              </div>
            </div>

            <div className="flex-1 overflow-y-auto p-4 space-y-2">
              <h4 className="text-xs font-semibold text-gray-500 uppercase mb-2">Scenes</h4>
              {currentProject.scenes.map(scene => (
                <div
                  key={scene.id}
                  className={`p-3 rounded-lg cursor-pointer transition-all ${
                    selectedScene?.id === scene.id
                      ? 'bg-electric-500/20 border border-electric-500/50'
                      : 'bg-gray-800/50 hover:bg-gray-800 border border-transparent'
                  }`}
                  onClick={() => setSelectedScene(scene)}
                >
                  <div className="flex items-center gap-2 mb-1">
                    <span className="text-xs font-mono text-gray-500">#{scene.scene_number}</span>
                    <span className="font-medium text-gray-200 text-sm truncate">{scene.title}</span>
                  </div>
                  <div className="text-xs text-gray-500 line-clamp-2">
                    {scene.description}
                  </div>
                  {selectedScene?.id === scene.id && (
                    <button
                      onClick={(e) => {
                        e.stopPropagation()
                        generateShotSuggestions(scene.id)
                      }}
                      className="mt-2 w-full px-2 py-1 bg-mint-500/20 hover:bg-mint-500/30 border border-mint-500/50 rounded text-xs text-mint-400"
                    >
                      + Generate Shots
                    </button>
                  )}
                </div>
              ))}
            </div>
          </>
        )}
      </div>

      {/* Main Content */}
      <div className="flex-1 flex flex-col overflow-hidden mt-10">
        {currentProject ? (
          <>
            {/* Toolbar */}
            <div className="border-b border-gray-800 bg-gray-900/80 backdrop-blur-md p-4">
              <div className="flex items-center justify-between mb-4">
                <h2 className="text-2xl font-bold text-electric-400">Storyboard Editor</h2>
                <div className="flex gap-2">
                  <button
                    onClick={loadScriptFromFile}
                    className="px-4 py-2 bg-gray-800/50 hover:bg-gray-700/50 border border-gray-700 rounded-lg text-gray-300 transition-all text-sm"
                  >
                    📁 Load Script
                  </button>
                  <button
                    onClick={exportStoryboard}
                    className="px-4 py-2 bg-mint-500/20 hover:bg-mint-500/30 border border-mint-500/50 rounded-lg text-mint-400 font-medium transition-all"
                  >
                    📄 Export PDF
                  </button>
                </div>
              </div>

              {/* Script Input */}
              {currentProject.scenes.length === 0 && (
                <div className="space-y-2">
                  {/* Warning Notice */}
                  <div className="p-3 bg-sunset-500/10 border border-sunset-500/30 rounded-lg">
                    <div className="flex items-start gap-2 text-xs">
                      <span className="text-sunset-400">⚠️</span>
                      <div className="text-gray-300">
                        <span className="font-semibold text-sunset-300">Note:</span> Script parser returns hardcoded scenes. Real screenplay parsing not yet implemented.
                      </div>
                    </div>
                  </div>

                  <div className="flex gap-2">
                    <textarea
                      value={script}
                      onChange={(e) => setScript(e.target.value)}
                      placeholder="Paste your script here or load from file..."
                      className="flex-1 h-24 px-4 py-3 bg-gray-800/50 border border-gray-700 rounded-lg text-gray-200 focus:border-electric-500 focus:outline-none resize-none"
                    />
                    <button
                      onClick={parseScript}
                      disabled={!script.trim() || isParsing}
                      className="px-6 bg-electric-500/20 hover:bg-electric-500/30 border border-electric-500/50 rounded-lg text-electric-400 font-medium transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                      {isParsing ? (
                        <span className="flex items-center gap-2">
                          <span className="animate-spin">⚙️</span>
                          Parsing...
                        </span>
                      ) : (
                        '🔍 Parse Script'
                      )}
                    </button>
                  </div>
                </div>
              )}
            </div>

            {/* Panels Grid */}
            {selectedScene ? (
              <div className="flex-1 overflow-y-auto p-6">
                <div className="mb-4">
                  <h3 className="text-xl font-bold text-gray-200 mb-1">
                    Scene {selectedScene.scene_number}: {selectedScene.title}
                  </h3>
                  <p className="text-gray-400 text-sm">{selectedScene.description}</p>
                  {selectedScene.location && (
                    <div className="mt-2 text-xs text-gray-500">
                      📍 {selectedScene.location} {selectedScene.time && `• ⏰ ${selectedScene.time}`}
                    </div>
                  )}
                </div>

                <div className="grid grid-cols-2 lg:grid-cols-3 gap-4">
                  {scenePanels.map(panel => (
                    <div
                      key={panel.id}
                      className="bg-gray-800/50 border border-gray-700 rounded-lg overflow-hidden hover:border-electric-500/30 transition-all"
                    >
                      {/* Panel Image */}
                      <div className="aspect-video bg-gray-900 relative">
                        {panel.image_url ? (
                          <img
                            src={panel.image_url}
                            alt={`Panel ${panel.panel_number}`}
                            className="w-full h-full object-cover"
                          />
                        ) : (
                          <div className="flex items-center justify-center h-full">
                            {generatingPanelId === panel.id ? (
                              <div className="text-center">
                                <div className="animate-spin text-4xl mb-2">⚙️</div>
                                <div className="text-xs text-gray-500">Generating...</div>
                              </div>
                            ) : (
                              <div className="text-center">
                                <button
                                  onClick={() => generatePanelImage(panel)}
                                  className="px-4 py-2 bg-electric-500/20 hover:bg-electric-500/30 border border-electric-500/50 rounded-lg text-electric-400 text-sm"
                                >
                                  🎨 Generate Image
                                </button>
                                <div className="mt-2 text-xs text-sunset-400">
                                  ⚠️ Placeholder only
                                </div>
                              </div>
                            )}
                          </div>
                        )}
                        <div className="absolute top-2 left-2 px-2 py-1 bg-black/70 rounded text-xs font-mono text-gray-300">
                          {panel.panel_number}
                        </div>
                      </div>

                      {/* Panel Info */}
                      <div className="p-3 space-y-2">
                        <div className="flex items-center gap-2">
                          <span className="px-2 py-1 bg-electric-500/10 border border-electric-500/30 rounded text-xs text-electric-400">
                            {panel.shot_type}
                          </span>
                        </div>
                        <div className="text-sm text-gray-300">{panel.description}</div>
                        {panel.dialogue && (
                          <div className="p-2 bg-gray-900/50 rounded text-xs text-gray-400 italic">
                            "{panel.dialogue}"
                          </div>
                        )}
                        {panel.notes && (
                          <div className="text-xs text-gray-500">{panel.notes}</div>
                        )}
                        <button
                          onClick={() => deletePanel(panel.id)}
                          className="text-xs text-sunset-400 hover:text-sunset-300"
                        >
                          Delete Panel
                        </button>
                      </div>
                    </div>
                  ))}
                </div>

                {scenePanels.length === 0 && (
                  <div className="text-center py-12">
                    <div className="text-4xl mb-2">🎬</div>
                    <p className="text-gray-500">No panels yet</p>
                    <p className="text-gray-600 text-sm mt-1">
                      Click "Generate Shots" in the sidebar to create panels
                    </p>
                  </div>
                )}
              </div>
            ) : (
              <div className="flex-1 flex items-center justify-center">
                <div className="text-center">
                  <div className="text-6xl mb-4">🎬</div>
                  <h3 className="text-2xl font-bold text-gray-300 mb-2">Select a Scene</h3>
                  <p className="text-gray-500">Choose a scene from the sidebar to view and edit panels</p>
                </div>
              </div>
            )}
          </>
        ) : (
          <div className="flex-1 flex items-center justify-center">
            <div className="text-center">
              <div className="text-6xl mb-4">🎬</div>
              <h3 className="text-2xl font-bold text-gray-300 mb-2">No Project Selected</h3>
              <p className="text-gray-500 mb-6">Create a new storyboard project to get started</p>
              <button
                onClick={() => setShowNewProjectDialog(true)}
                className="px-6 py-3 bg-electric-500/20 hover:bg-electric-500/30 border border-electric-500/50 rounded-lg text-electric-400 font-medium transition-all"
              >
                + Create New Project
              </button>
            </div>
          </div>
        )}
      </div>

      {/* Canvas for annotations (hidden initially) */}
      <canvas ref={canvasRef} className="hidden" />

      {/* New Project Dialog */}
      {showNewProjectDialog && (
        <div className="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50 p-4">
          <div className="bg-gray-900 border border-electric-500/30 rounded-xl max-w-md w-full anime-glow">
            <div className="p-6 border-b border-gray-800">
              <h2 className="text-2xl font-bold text-electric-400">Create Storyboard Project</h2>
            </div>

            <div className="p-6 space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-300 mb-2">Project Name</label>
                <input
                  type="text"
                  value={newProjectName}
                  onChange={(e) => setNewProjectName(e.target.value)}
                  placeholder="Enter project name..."
                  className="w-full px-4 py-3 bg-gray-800/50 border border-gray-700 rounded-lg focus:outline-none focus:border-electric-500 text-white"
                  onKeyDown={(e) => e.key === 'Enter' && createProject()}
                  autoFocus
                />
              </div>
            </div>

            <div className="p-6 border-t border-gray-800 flex gap-3">
              <button
                onClick={() => {
                  setShowNewProjectDialog(false)
                  setNewProjectName('')
                }}
                className="flex-1 px-6 py-3 bg-gray-800/50 hover:bg-gray-700/50 border border-gray-700 rounded-lg text-gray-300 transition-all"
              >
                Cancel
              </button>
              <button
                onClick={createProject}
                disabled={!newProjectName.trim()}
                className="flex-1 px-6 py-3 bg-electric-500/20 hover:bg-electric-500/30 border border-electric-500/50 rounded-lg text-electric-400 font-medium transition-all disabled:opacity-50 disabled:cursor-not-allowed"
              >
                Create
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
