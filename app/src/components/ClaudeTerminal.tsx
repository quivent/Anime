import { useEffect, useRef, useState } from 'react'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import { WebLinksAddon } from '@xterm/addon-web-links'
import { invoke } from '@tauri-apps/api/core'
import { listen } from '@tauri-apps/api/event'
import '@xterm/xterm/css/xterm.css'

interface TerminalOutput {
  data: string
}

export default function ClaudeTerminal() {
  const terminalRef = useRef<HTMLDivElement>(null)
  const xtermRef = useRef<Terminal | null>(null)
  const fitAddonRef = useRef<FitAddon | null>(null)
  const [sessionId, setSessionId] = useState<string | null>(null)
  const [connected, setConnected] = useState(false)
  const sessionIdRef = useRef<string | null>(null)
  const connectedRef = useRef(false)

  useEffect(() => {
    if (!terminalRef.current) return

    const term = new Terminal({
      cursorBlink: true,
      fontSize: 14,
      fontFamily: 'Menlo, Monaco, "Courier New", monospace',
      theme: {
        background: '#0a0a0a',
        foreground: '#e5e7eb',
        cursor: '#ec4899',
        cursorAccent: '#0a0a0a',
        selectionBackground: '#ec48994d',
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
    term.writeln('\x1b[35m╔════════════════════════════════════════════╗\x1b[0m')
    term.writeln('\x1b[35m║      Claude Code Terminal Interface        ║\x1b[0m')
    term.writeln('\x1b[35m╚════════════════════════════════════════════╝\x1b[0m')
    term.writeln('')
    term.writeln('\x1b[36mClick "Connect" to start a Claude Code session\x1b[0m')
    term.writeln('')

    const handleResize = () => {
      fitAddon.fit()
      if (sessionIdRef.current && connectedRef.current) {
        const { rows, cols } = term
        invoke('terminal_resize', {
          sessionId: sessionIdRef.current,
          rows,
          cols
        }).catch(() => {})
      }
    }
    window.addEventListener('resize', handleResize)

    term.onData((data) => {
      if (connectedRef.current && sessionIdRef.current) {
        invoke('terminal_input', { sessionId: sessionIdRef.current, data }).catch(() => {})
      }
    })

    return () => {
      window.removeEventListener('resize', handleResize)
      term.dispose()
    }
  }, [])

  useEffect(() => {
    if (!xtermRef.current) return

    const unlisten = listen<TerminalOutput>('terminal_output', (event) => {
      if (event.payload.data && xtermRef.current) {
        xtermRef.current.write(event.payload.data)
      }
    })

    return () => {
      unlisten.then((fn) => fn())
    }
  }, [])

  useEffect(() => {
    connectedRef.current = connected
    sessionIdRef.current = sessionId
  }, [connected, sessionId])

  const handleConnect = async () => {
    if (!xtermRef.current) return

    try {
      // Start a local shell session
      const sid = await invoke<string>('terminal_connect_local', {})

      sessionIdRef.current = sid
      connectedRef.current = true
      setSessionId(sid)
      setConnected(true)

      xtermRef.current.clear()
      xtermRef.current.writeln('\x1b[32m✓ Local shell session started\x1b[0m')
      xtermRef.current.writeln('\x1b[36mℹ Starting Claude Code...\x1b[0m')
      xtermRef.current.writeln('')

      if (fitAddonRef.current) {
        fitAddonRef.current.fit()
        const { rows, cols } = xtermRef.current
        await invoke('terminal_resize', { sessionId: sid, rows, cols }).catch(() => {})
      }

      // Send claude command
      setTimeout(() => {
        if (sessionIdRef.current) {
          invoke('terminal_input', {
            sessionId: sessionIdRef.current,
            data: 'claude\n'
          }).catch(() => {})
        }
      }, 500)

      xtermRef.current.focus()
    } catch (error) {
      const errorMsg = error instanceof Error ? error.message : String(error)
      if (xtermRef.current) {
        xtermRef.current.writeln(`\x1b[31m✗ Connection failed: ${errorMsg}\x1b[0m`)
        xtermRef.current.writeln(`\x1b[33mℹ Make sure you have the 'claude' command installed\x1b[0m`)
      }
    }
  }

  const handleDisconnect = async () => {
    if (sessionId) {
      try {
        await invoke('terminal_disconnect', { sessionId })
        setConnected(false)
        setSessionId(null)
        if (xtermRef.current) {
          xtermRef.current.writeln('')
          xtermRef.current.writeln('\x1b[33m✓ Disconnected from Claude Code\x1b[0m')
        }
      } catch (error) {
        // Ignore disconnect errors
      }
    }
  }

  return (
    <div className="h-full flex flex-col overflow-hidden bg-gray-950">
      {/* Header */}
      <div className="border-b border-gray-800 bg-gray-900/80 backdrop-blur-md p-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <span className="text-2xl">🤖</span>
            <div>
              <h2 className="text-lg font-bold text-electric-400">Claude Code</h2>
              <p className="text-xs text-gray-400">
                Interactive AI coding assistant
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
                className="px-4 py-2 bg-electric-500/20 hover:bg-electric-500/30 border border-electric-500/50 rounded-lg text-electric-400 font-medium transition-all text-sm"
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
            <span>⌨️ Claude Code CLI</span>
            <span>•</span>
            <span>Type naturally to interact with Claude</span>
          </div>
          {!connected && (
            <span className="text-gray-400">
              Click Connect to start
            </span>
          )}
        </div>
      </div>
    </div>
  )
}
