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

export default function LocalTerminal() {
  const terminalRef = useRef<HTMLDivElement>(null)
  const xtermRef = useRef<Terminal | null>(null)
  const fitAddonRef = useRef<FitAddon | null>(null)
  const [sessionId, setSessionId] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)
  const sessionIdRef = useRef<string | null>(null)

  useEffect(() => {
    if (!terminalRef.current) return

    // Initialize xterm.js
    const term = new Terminal({
      cursorBlink: true,
      fontSize: 14,
      fontFamily: 'Menlo, Monaco, "Courier New", monospace',
      theme: {
        background: '#0a0a0a',
        foreground: '#e5e7eb',
        cursor: '#f472b6',
        selection: 'rgba(244, 114, 182, 0.3)',
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

    // Handle terminal input
    term.onData((data) => {
      if (sessionIdRef.current) {
        invoke('terminal_input', {
          sessionId: sessionIdRef.current,
          data,
        }).catch((err) => {
          console.error('[LocalTerminal] Failed to send input:', err)
        })
      }
    })

    // Handle window resize
    const handleResize = () => {
      if (fitAddonRef.current && xtermRef.current && sessionIdRef.current) {
        fitAddon.fit()
        const { rows, cols } = xtermRef.current
        invoke('terminal_resize', {
          sessionId: sessionIdRef.current,
          rows,
          cols,
        }).catch(() => {})
      }
    }

    window.addEventListener('resize', handleResize)

    // Auto-connect to local terminal
    connectToLocal()

    return () => {
      window.removeEventListener('resize', handleResize)
      if (sessionIdRef.current) {
        invoke('terminal_disconnect', { sessionId: sessionIdRef.current }).catch(() => {})
      }
      term.dispose()
    }
  }, [])

  // Listen for terminal output
  useEffect(() => {
    if (!sessionId) return

    const unlisten = listen<TerminalOutput>('terminal_output', (event) => {
      if (xtermRef.current) {
        xtermRef.current.write(event.payload.data)
      }
    })

    return () => {
      unlisten.then((fn) => fn())
    }
  }, [sessionId])

  const connectToLocal = async () => {
    if (!xtermRef.current) return

    try {
      // Connect to local terminal (shell)
      const sid = await invoke<string>('terminal_connect_local')

      sessionIdRef.current = sid
      setSessionId(sid)

      xtermRef.current.clear()
      xtermRef.current.focus()

      // Fit and resize
      if (fitAddonRef.current) {
        fitAddonRef.current.fit()
        const { rows, cols } = xtermRef.current
        await invoke('terminal_resize', { sessionId: sid, rows, cols }).catch(() => {})
      }
    } catch (error) {
      const errorMsg = error instanceof Error ? error.message : String(error)
      console.error('[LocalTerminal] Connection failed:', errorMsg)
      setError(`Failed to start local terminal: ${errorMsg}`)
      if (xtermRef.current) {
        xtermRef.current.writeln(`\x1b[31m✗ Failed to start terminal: ${errorMsg}\x1b[0m`)
      }
    }
  }

  return (
    <div className="flex flex-col h-full bg-gray-950">
      {/* Header */}
      <div className="px-6 py-3 border-b border-gray-800 bg-gray-900/50 backdrop-blur-md">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <span className="text-xl">💻</span>
            <div>
              <h2 className="text-sm font-semibold text-gray-200">Local Terminal</h2>
              <p className="text-xs text-gray-500">This Machine</p>
            </div>
          </div>
        </div>
      </div>

      {/* Terminal Container */}
      <div className="flex-1 p-4 overflow-hidden">
        {error && (
          <div className="mb-4 p-4 bg-red-500/10 border border-red-500/30 rounded-lg">
            <p className="text-sm text-red-400">{error}</p>
          </div>
        )}
        <div
          ref={terminalRef}
          className="w-full h-full rounded-lg overflow-hidden border border-gray-800"
          style={{ background: '#0a0a0a' }}
        />
      </div>
    </div>
  )
}
