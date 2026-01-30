import { useState, useEffect } from 'react'
import { useInstanceStore } from '../store/instanceStore'
import type { Instance } from '../types/lambda'
import LocalTerminal from './LocalTerminal'
import TerminalView from './TerminalView'

export default function TerminalListView() {
  const [instances, setInstances] = useState<Instance[]>([])
  const [activeTerminal, setActiveTerminal] = useState<string>('local')
  const { setSelectedInstance } = useInstanceStore()

  useEffect(() => {
    loadInstances()
  }, [])

  const loadInstances = async () => {
    try {
      const { invoke } = await import('@tauri-apps/api/core')
      const result = await invoke<{ instances: Instance[] }>('list_instances')
      setInstances(result.instances || [])
    } catch (error) {
      console.error('[TerminalListView] Failed to load instances:', error)
    }
  }

  const handleTerminalSelect = (terminalId: string, instance?: Instance) => {
    setActiveTerminal(terminalId)
    if (instance) {
      setSelectedInstance(instance)
    }
  }

  return (
    <div className="flex h-full">
      {/* Terminal List Sidebar */}
      <div className="w-64 border-r border-gray-800 bg-gray-900/50 backdrop-blur-md overflow-y-auto">
        <div className="p-4 space-y-2">
          <h3 className="text-xs font-semibold text-gray-500 uppercase tracking-wider px-2 mb-3">
            Terminals
          </h3>

          {/* Local Machine Terminal */}
          <button
            onClick={() => handleTerminalSelect('local')}
            className={`
              w-full flex items-center gap-3 px-4 py-3 rounded-lg
              transition-all duration-200
              ${activeTerminal === 'local'
                ? 'bg-mint-500/20 border-mint-500/50 text-mint-400 border anime-glow'
                : 'hover:bg-gray-800/50 text-gray-400 hover:text-gray-300'
              }
            `}
          >
            <span className="text-xl">💻</span>
            <div className="flex-1 text-left">
              <div className="font-medium">This Machine</div>
              <div className="text-xs opacity-70">Local Terminal</div>
            </div>
          </button>

          {/* Lambda Instance Terminals */}
          {instances.filter(i => i.status === 'active').map((instance) => (
            <button
              key={instance.id}
              onClick={() => handleTerminalSelect(instance.id, instance)}
              className={`
                w-full flex items-center gap-3 px-4 py-3 rounded-lg
                transition-all duration-200
                ${activeTerminal === instance.id
                  ? 'bg-electric-500/20 border-electric-500/50 text-electric-400 border anime-glow'
                  : 'hover:bg-gray-800/50 text-gray-400 hover:text-gray-300'
                }
              `}
            >
              <span className="text-xl">🖥️</span>
              <div className="flex-1 text-left">
                <div className="font-medium">{instance.hostname || instance.name}</div>
                <div className="text-xs opacity-70">{instance.ip}</div>
              </div>
            </button>
          ))}
        </div>
      </div>

      {/* Terminal Display Area */}
      <div className="flex-1">
        {activeTerminal === 'local' ? (
          <LocalTerminal />
        ) : (
          <TerminalView key={activeTerminal} />
        )}
      </div>
    </div>
  )
}
