import { useState } from 'react'
import ClaudeTerminal from './ClaudeTerminal'
import OllamaTerminal from './OllamaTerminal'

type Tab = 'claude' | 'ollama'

export default function OurGuysView() {
  const [activeTab, setActiveTab] = useState<Tab>('claude')

  const tabs = [
    { id: 'claude' as Tab, name: 'Claude Code', icon: '🤖', color: 'electric' },
    { id: 'ollama' as Tab, name: 'Ollama', icon: '🦙', color: 'mint' },
  ]

  return (
    <div className="h-full flex flex-col overflow-hidden bg-gray-950">
      {/* Header */}
      <div className="border-b border-gray-800 bg-gray-900/50 p-6">
        <div className="flex items-center justify-between mb-4">
          <div>
            <h2 className="text-2xl font-bold text-electric-400 flex items-center gap-3">
              <span className="text-3xl">👥</span>
              Our Guys
            </h2>
            <p className="text-gray-400 mt-1 text-sm">
              AI assistants and language models
            </p>
          </div>
        </div>

        {/* Tabs */}
        <div className="flex gap-2">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={`px-6 py-3 rounded-lg font-medium transition-all flex items-center gap-2 ${
                activeTab === tab.id
                  ? tab.color === 'electric'
                    ? 'bg-electric-500/20 border border-electric-500/50 text-electric-400 anime-glow-electric'
                    : 'bg-mint-500/20 border border-mint-500/50 text-mint-400 anime-glow-mint'
                  : 'bg-gray-800/30 border border-gray-700 text-gray-400 hover:bg-gray-800/50'
              }`}
            >
              <span className="text-xl">{tab.icon}</span>
              <span>{tab.name}</span>
            </button>
          ))}
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-hidden">
        {activeTab === 'claude' && <ClaudeTerminal />}
        {activeTab === 'ollama' && <OllamaTerminal />}
      </div>
    </div>
  )
}
