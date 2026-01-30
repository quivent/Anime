type View = 'lambda' | 'terminal' | 'lambda-terminal' | 'packages' | 'wizard' | 'servers' | 'flows' | 'models' | 'visual' | 'writing' | 'analysis' | 'storyboards' | 'todos' | 'claude' | 'ollama'

interface SidebarProps {
  currentView: View
  onViewChange: (view: View) => void
}

export default function Sidebar({ currentView, onViewChange }: SidebarProps) {
  const navItems = [
    { id: 'lambda', label: 'Lambda', icon: '☁️', color: 'electric', group: 'cloud' },
    { id: 'terminal', label: 'Terminals', icon: '💻', color: 'mint', group: 'cloud' },
    { id: 'packages', label: 'Packages', icon: '📦', color: 'sakura', group: 'setup' },
    { id: 'wizard', label: 'Wizard', icon: '🔮', color: 'neon', group: 'setup' },
    { id: 'todos', label: 'Todos', icon: '📋', color: 'electric', group: 'setup' },
  ]

  const creativeItems = [
    { id: 'flows', label: 'Workflows', icon: '🌊', color: 'electric', group: 'creative' },
    { id: 'models', label: 'Models', icon: '🤖', color: 'neon', group: 'creative' },
    { id: 'visual', label: 'Visual/Animation', icon: '🎨', color: 'sakura', group: 'creative' },
    { id: 'writing', label: 'Writing', icon: '✍️', color: 'mint', group: 'creative' },
    { id: 'analysis', label: 'Coverage', icon: '🔍', color: 'sunset', group: 'creative' },
    { id: 'storyboards', label: 'Storyboards', icon: '🎬', color: 'neon', group: 'creative' },
  ]

  return (
    <aside className="w-64 border-r border-gray-800 bg-gray-900/50 backdrop-blur-md flex flex-col">
      <nav className="flex-1 p-4 space-y-4 overflow-y-auto">
        {/* Cloud Section */}
        <div className="space-y-2">
          <h3 className="text-xs font-semibold text-gray-500 uppercase tracking-wider px-2">Cloud</h3>
          {navItems.filter(item => item.group === 'cloud').map((item) => {
            const isActive = currentView === item.id
            return (
              <button
                key={item.id}
                onClick={() => onViewChange(item.id as View)}
                className={`
                  w-full flex items-center gap-3 px-4 py-3 rounded-lg
                  transition-all duration-200
                  ${isActive
                    ? `bg-${item.color}-500/20 border-${item.color}-500/50 text-${item.color}-400 border anime-glow`
                    : 'hover:bg-gray-800/50 text-gray-400 hover:text-gray-300'
                  }
                `}
              >
                <span className="text-xl">{item.icon}</span>
                <span className="font-medium">{item.label}</span>
              </button>
            )
          })}
        </div>

        {/* Setup Section */}
        <div className="space-y-2">
          <h3 className="text-xs font-semibold text-gray-500 uppercase tracking-wider px-2">Setup</h3>
          {navItems.filter(item => item.group === 'setup').map((item) => {
            const isActive = currentView === item.id
            return (
              <button
                key={item.id}
                onClick={() => onViewChange(item.id as View)}
                className={`
                  w-full flex items-center gap-3 px-4 py-3 rounded-lg
                  transition-all duration-200
                  ${isActive
                    ? `bg-${item.color}-500/20 border-${item.color}-500/50 text-${item.color}-400 border anime-glow`
                    : 'hover:bg-gray-800/50 text-gray-400 hover:text-gray-300'
                  }
                `}
              >
                <span className="text-xl">{item.icon}</span>
                <span className="font-medium">{item.label}</span>
              </button>
            )
          })}
        </div>

        {/* Creative Tools Section */}
        <div className="space-y-2">
          <h3 className="text-xs font-semibold text-gray-500 uppercase tracking-wider px-2">Creative Tools</h3>
          {creativeItems.map((item) => {
            const isActive = currentView === item.id
            return (
              <button
                key={item.id}
                onClick={() => onViewChange(item.id as View)}
                className={`
                  w-full flex items-center gap-3 px-4 py-3 rounded-lg
                  transition-all duration-200
                  ${isActive
                    ? `bg-${item.color}-500/20 border-${item.color}-500/50 text-${item.color}-400 border anime-glow`
                    : 'hover:bg-gray-800/50 text-gray-400 hover:text-gray-300'
                  }
                `}
              >
                <span className="text-xl">{item.icon}</span>
                <span className="font-medium text-sm">{item.label}</span>
              </button>
            )
          })}
        </div>

        {/* Our Guys Section */}
        <div className="space-y-2">
          <h3 className="text-xs font-semibold text-gray-500 uppercase tracking-wider px-2">Our Guys</h3>
          <button
            onClick={() => onViewChange('claude')}
            className={`
              w-full flex items-center gap-3 px-4 py-3 rounded-lg
              transition-all duration-200
              ${currentView === 'claude'
                ? 'bg-electric-500/20 border-electric-500/50 text-electric-400 border anime-glow'
                : 'hover:bg-gray-800/50 text-gray-400 hover:text-gray-300'
              }
            `}
          >
            <span className="text-xl">🤖</span>
            <span className="font-medium">Claude</span>
          </button>
          <button
            onClick={() => onViewChange('ollama')}
            className={`
              w-full flex items-center gap-3 px-4 py-3 rounded-lg
              transition-all duration-200
              ${currentView === 'ollama'
                ? 'bg-mint-500/20 border-mint-500/50 text-mint-400 border anime-glow'
                : 'hover:bg-gray-800/50 text-gray-400 hover:text-gray-300'
              }
            `}
          >
            <span className="text-xl">🦙</span>
            <span className="font-medium">Ollama</span>
          </button>
        </div>
      </nav>

      <div className="p-4 border-t border-gray-800">
        <div className="text-xs text-gray-500 space-y-1">
          <div className="flex items-center justify-between">
            <span>Status</span>
            <span className="text-mint-400">●  Online</span>
          </div>
          <div className="flex items-center justify-between">
            <span>Theme</span>
            <span className="text-sakura-400">Sakura</span>
          </div>
        </div>
      </div>
    </aside>
  )
}
