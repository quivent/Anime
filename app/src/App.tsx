import { useState } from 'react'
import Sidebar from './components/Sidebar'
import PackageGrid from './components/PackageGrid'
import WizardFlow from './components/WizardFlow'
import ServerManager from './components/ServerManager'
import LambdaView from './components/LambdaView'
import TerminalListView from './components/TerminalListView'
import SakuraBackground from './components/SakuraBackground'
import SakuraDecoration from './components/SakuraDecoration'
import ModelsView from './components/ModelsView'
import VisualView from './components/VisualView'
import WorkflowsView from './components/WorkflowsView'
import WritingView from './components/WritingView'
import AnalysisView from './components/AnalysisView'
import StoryboardsView from './components/StoryboardsView'
import TodosView from './components/TodosView'
import CoverageView from './components/CoverageView'
import ClaudeTerminal from './components/ClaudeTerminal'
import OllamaTerminal from './components/OllamaTerminal'
import { usePackageStore } from './store/packageStore'

type View = 'lambda' | 'terminal' | 'lambda-terminal' | 'packages' | 'wizard' | 'servers' | 'flows' | 'models' | 'visual' | 'writing' | 'analysis' | 'storyboards' | 'todos' | 'ourguys' | 'claude' | 'ollama'

function App() {
  const [currentView, setCurrentView] = useState<View>('lambda')
  const { installing } = usePackageStore()

  return (
    <div className="flex h-screen w-screen overflow-hidden bg-gray-950 text-gray-100">
      <SakuraBackground />

      <Sidebar currentView={currentView} onViewChange={setCurrentView} />

      <main className="flex-1 flex flex-col overflow-hidden relative">
        {/* Header */}
        <header className="border-b border-gray-800 bg-gray-900/80 backdrop-blur-md relative overflow-hidden">
          <SakuraDecoration />
          <div className="px-6 py-4 relative z-10">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div className="text-3xl">🍒</div>
                <div>
                  <h1 className="text-2xl font-bold sakura-gradient bg-clip-text text-transparent">
                    ANIME
                  </h1>
                  <p className="text-xs text-gray-400">
                    Lambda GH200 Deployment Manager
                  </p>
                </div>
              </div>

              <div className="flex items-center gap-2">
                <div className="px-3 py-1 rounded-full bg-electric-500/10 border border-electric-500/30 text-electric-400 text-xs font-mono">
                  v0.1.0
                </div>
                {installing && (
                  <div className="px-3 py-1 rounded-full bg-mint-500/10 border border-mint-500/30 text-mint-400 text-xs animate-pulse">
                    Installing...
                  </div>
                )}
              </div>
            </div>
          </div>
        </header>

        {/* Content Area */}
        <div className="flex-1 overflow-hidden">
          {currentView === 'lambda' && <LambdaView />}
          {currentView === 'terminal' && <TerminalListView />}
          {currentView === 'packages' && <PackageGrid />}
          {currentView === 'wizard' && <WizardFlow />}
          {currentView === 'servers' && <ServerManager />}
          {currentView === 'flows' && <WorkflowsView />}
          {currentView === 'models' && <ModelsView />}
          {currentView === 'visual' && <VisualView />}
          {currentView === 'writing' && <WritingView />}
          {currentView === 'analysis' && <CoverageView />}
          {currentView === 'storyboards' && <StoryboardsView />}
          {currentView === 'todos' && <TodosView />}
          {currentView === 'ourguys' && <div className="flex gap-4 p-4"><div className="flex-1"><ClaudeTerminal /></div><div className="flex-1"><OllamaTerminal /></div></div>}
          {currentView === 'claude' && <ClaudeTerminal />}
          {currentView === 'ollama' && <OllamaTerminal />}
        </div>

        {/* Footer */}
        <footer className="border-t border-gray-800 bg-gray-900/80 backdrop-blur-md px-6 py-3">
          <div className="flex items-center justify-between text-xs text-gray-500">
            <div className="flex items-center gap-4">
              <span>⚡ Powered by Tauri</span>
              <span>•</span>
              <span>🦀 Rust Backend</span>
              <span>•</span>
              <span>⚛️ React Frontend</span>
            </div>
            <div className="font-mono">
              Ready to deploy
            </div>
          </div>
        </footer>
      </main>
    </div>
  )
}

export default App
