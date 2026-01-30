import { useEffect, useRef, useState } from 'react'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import { WebLinksAddon } from '@xterm/addon-web-links'
import { useInstanceStore } from '../store/instanceStore'
import '@xterm/xterm/css/xterm.css'

export default function OllamaTerminal() {
  const terminalRef = useRef<HTMLDivElement>(null)
  const xtermRef = useRef<Terminal | null>(null)
  const fitAddonRef = useRef<FitAddon | null>(null)
  const { selectedInstance } = useInstanceStore()
  const [ollamaEndpoint, setOllamaEndpoint] = useState('')
  const [connected, setConnected] = useState(false)
  const [currentModel, setCurrentModel] = useState('llama2')
  const [chatHistory, setChatHistory] = useState<Array<{ role: string; content: string }>>([])
  const inputBufferRef = useRef('')

  useEffect(() => {
    if (selectedInstance?.ip) {
      setOllamaEndpoint(`http://${selectedInstance.ip}:11434`)
    }
  }, [selectedInstance])

  useEffect(() => {
    if (!terminalRef.current) return

    const term = new Terminal({
      cursorBlink: true,
      fontSize: 14,
      fontFamily: 'Menlo, Monaco, "Courier New", monospace',
      theme: {
        background: '#0a0a0a',
        foreground: '#e5e7eb',
        cursor: '#10b981',
        cursorAccent: '#0a0a0a',
        selectionBackground: '#10b9814d',
        black: '#1f2937',
        red: '#ef4444',
        green: '#10b981',
        yellow: '#f59e0b',
        blue: '#3b82f6',
        magenta: '#ec4899',
        cyan: '#06b6d4',
        white: '#e5e7eb',
        brightBlack: '#6b7280',
        brightRed: '#f87171',
        brightGreen: '#34d399',
        brightYellow: '#fbbf24',
        brightBlue: '#60a5fa',
        brightMagenta: '#f472b6',
        brightCyan: '#22d3ee',
        brightWhite: '#f9fafb',
      },
    })

    const fitAddon = new FitAddon()
    const webLinksAddon = new WebLinksAddon()

    term.loadAddon(fitAddon)
    term.loadAddon(webLinksAddon)
    term.open(terminalRef.current)
    fitAddon.fit()

    xtermRef.current = term
    fitAddonRef.current = fitAddon

    // Welcome message
    term.writeln('\x1b[32m╔════════════════════════════════════════════╗\x1b[0m')
    term.writeln('\x1b[32m║        Ollama Terminal Interface           ║\x1b[0m')
    term.writeln('\x1b[32m╚════════════════════════════════════════════╝\x1b[0m')
    term.writeln('')
    term.writeln('\x1b[36mConfigure your Lambda GPU instance endpoint\x1b[0m')
    term.writeln('')

    const handleResize = () => {
      fitAddon.fit()
    }
    window.addEventListener('resize', handleResize)

    term.onData((data) => {
      if (!connected) return

      // Handle special keys
      if (data === '\r') {
        // Enter key
        term.write('\r\n')
        const input = inputBufferRef.current.trim()
        inputBufferRef.current = ''

        if (input) {
          handleUserInput(input, term)
        } else {
          term.write('\x1b[32m❯\x1b[0m ')
        }
      } else if (data === '\x7F') {
        // Backspace
        if (inputBufferRef.current.length > 0) {
          inputBufferRef.current = inputBufferRef.current.slice(0, -1)
          term.write('\b \b')
        }
      } else if (data === '\x03') {
        // Ctrl+C
        term.write('^C\r\n')
        inputBufferRef.current = ''
        term.write('\x1b[32m❯\x1b[0m ')
      } else {
        // Regular character
        inputBufferRef.current += data
        term.write(data)
      }
    })

    return () => {
      window.removeEventListener('resize', handleResize)
      term.dispose()
    }
  }, [])

  const handleUserInput = async (input: string, term: Terminal) => {
    const trimmedInput = input.trim()

    // Handle special commands
    if (trimmedInput.startsWith('/')) {
      if (trimmedInput === '/clear') {
        term.clear()
        term.write('\x1b[32m❯\x1b[0m ')
        return
      } else if (trimmedInput.startsWith('/model ')) {
        const model = trimmedInput.substring(7).trim()
        setCurrentModel(model)
        term.writeln(`\x1b[33mℹ Model changed to: ${model}\x1b[0m`)
        term.write('\x1b[32m❯\x1b[0m ')
        return
      } else if (trimmedInput === '/help') {
        term.writeln('\x1b[36mAvailable commands:\x1b[0m')
        term.writeln('  /model <name>  - Change the model')
        term.writeln('  /clear         - Clear the screen')
        term.writeln('  /help          - Show this help')
        term.writeln('')
        term.write('\x1b[32m❯\x1b[0m ')
        return
      }
    }

    // Send to Ollama
    try {
      term.writeln('\x1b[36mThinking...\x1b[0m')

      const newHistory = [...chatHistory, { role: 'user', content: trimmedInput }]

      const response = await fetch(`${ollamaEndpoint}/api/chat`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          model: currentModel,
          messages: newHistory,
          stream: false,
        }),
      })

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}`)
      }

      const data = await response.json()
      const assistantMessage = data.message.content

      term.writeln(`\x1b[35m${assistantMessage}\x1b[0m`)
      term.writeln('')

      setChatHistory([...newHistory, { role: 'assistant', content: assistantMessage }])
    } catch (error) {
      const errorMsg = error instanceof Error ? error.message : String(error)
      term.writeln(`\x1b[31m✗ Error: ${errorMsg}\x1b[0m`)
      term.writeln(`\x1b[33mℹ Make sure Ollama is running at ${ollamaEndpoint}\x1b[0m`)
      term.writeln('')
    }

    term.write('\x1b[32m❯\x1b[0m ')
  }

  const handleConnect = () => {
    if (!xtermRef.current) return
    if (!ollamaEndpoint) {
      xtermRef.current.writeln('\x1b[31m✗ Please configure an endpoint first\x1b[0m')
      return
    }

    setConnected(true)
    setChatHistory([])
    xtermRef.current.clear()
    xtermRef.current.writeln(`\x1b[32m✓ Connected to Ollama at ${ollamaEndpoint}\x1b[0m`)
    xtermRef.current.writeln(`\x1b[36mℹ Model: ${currentModel}\x1b[0m`)
    xtermRef.current.writeln(`\x1b[36mℹ Type /help for commands\x1b[0m`)
    xtermRef.current.writeln('')
    xtermRef.current.write('\x1b[32m❯\x1b[0m ')
    xtermRef.current.focus()
  }

  const handleDisconnect = () => {
    setConnected(false)
    setChatHistory([])
    inputBufferRef.current = ''
    if (xtermRef.current) {
      xtermRef.current.writeln('')
      xtermRef.current.writeln('\x1b[33m✓ Disconnected from Ollama\x1b[0m')
    }
  }

  return (
    <div className="h-full flex flex-col overflow-hidden bg-gray-950">
      {/* Header */}
      <div className="border-b border-gray-800 bg-gray-900/80 backdrop-blur-md p-4">
        <div className="flex items-center justify-between mb-3">
          <div className="flex items-center gap-3">
            <span className="text-2xl">🦙</span>
            <div>
              <h2 className="text-lg font-bold text-mint-400">Ollama</h2>
              <p className="text-xs text-gray-400">
                Lambda GPU Instance • Model: {currentModel}
              </p>
            </div>
          </div>

          <div className="flex items-center gap-2">
            {connected && (
              <div className="px-3 py-1 rounded-full bg-mint-500/10 border border-mint-500/30 text-mint-400 text-xs flex items-center gap-2">
                <div className="w-2 h-2 bg-mint-500 rounded-full animate-pulse" />
                Connected
              </div>
            )}

            {!connected ? (
              <button
                onClick={handleConnect}
                disabled={!ollamaEndpoint}
                className="px-4 py-2 bg-mint-500/20 hover:bg-mint-500/30 border border-mint-500/50 rounded-lg text-mint-400 font-medium transition-all disabled:opacity-50 disabled:cursor-not-allowed text-sm"
              >
                🔌 Connect
              </button>
            ) : (
              <button
                onClick={handleDisconnect}
                className="px-4 py-2 bg-sunset-500/20 hover:bg-sunset-500/30 border border-sunset-500/50 rounded-lg text-sunset-400 font-medium transition-all text-sm"
              >
                🔌 Disconnect
              </button>
            )}
          </div>
        </div>

        {/* Endpoint Configuration */}
        {!connected && (
          <div className="flex gap-2 items-center">
            <label className="text-xs text-gray-400 whitespace-nowrap">Endpoint:</label>
            <input
              type="text"
              value={ollamaEndpoint}
              onChange={(e) => setOllamaEndpoint(e.target.value)}
              placeholder="http://instance-ip:11434"
              className="flex-1 px-3 py-2 bg-gray-800/50 border border-gray-700 rounded-lg focus:outline-none focus:border-mint-500 text-white text-sm"
            />
            <input
              type="text"
              value={currentModel}
              onChange={(e) => setCurrentModel(e.target.value)}
              placeholder="model name"
              className="w-32 px-3 py-2 bg-gray-800/50 border border-gray-700 rounded-lg focus:outline-none focus:border-mint-500 text-white text-sm"
            />
          </div>
        )}
      </div>

      {/* Terminal Container */}
      <div className="flex-1 p-4 overflow-hidden">
        <div
          ref={terminalRef}
          onClick={() => xtermRef.current?.focus()}
          className="h-full w-full rounded-lg border border-gray-800 bg-[#0a0a0a] p-2 cursor-text"
          style={{ minHeight: '400px' }}
        />
      </div>

      {/* Footer */}
      <div className="border-t border-gray-800 bg-gray-900/80 backdrop-blur-md px-4 py-2">
        <div className="flex items-center justify-between text-xs text-gray-500">
          <div className="flex items-center gap-4">
            <span>🦙 Ollama Chat</span>
            <span>•</span>
            <span>Type /help for commands</span>
          </div>
          {!connected && (
            <span className="text-gray-400">
              Configure endpoint and click Connect
            </span>
          )}
        </div>
      </div>
    </div>
  )
}
