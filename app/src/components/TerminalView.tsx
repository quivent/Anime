import { useEffect, useRef, useState, useCallback } from 'react'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import { WebLinksAddon } from '@xterm/addon-web-links'
import { invoke } from '@tauri-apps/api/core'
import { listen } from '@tauri-apps/api/event'
import { useInstanceStore } from '../store/instanceStore'
import type { Instance } from '../types/lambda'
import '@xterm/xterm/css/xterm.css'

interface TerminalOutput {
  data: string
}

export default function TerminalView() {
  const { selectedInstance, setSelectedInstance } = useInstanceStore()
  const terminalRef = useRef<HTMLDivElement>(null)
  const xtermRef = useRef<Terminal | null>(null)
  const fitAddonRef = useRef<FitAddon | null>(null)
  const [connected, setConnected] = useState(false)
  const [connecting, setConnecting] = useState(false)
  const [sessionId, setSessionId] = useState<string | null>(null)
  const [instances, setInstances] = useState<Instance[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // Use refs to avoid stale closure in onData callback
  const connectedRef = useRef(false)
  const sessionIdRef = useRef<string | null>(null)
  const autoConnectAttemptedRef = useRef(false)
  const autoConnectTimeoutRef = useRef<number | null>(null)

  useEffect(() => {
    loadInstances()
  }, [])

  // Define connectToInstance before it's used in auto-connect
  const connectToInstance = useCallback(async () => {
    if (!selectedInstance || !selectedInstance.ip) {
      const msg = 'Please select an active Lambda instance first!'
      setError(msg)
      return
    }

    console.log('[TerminalView] Connecting to instance:', selectedInstance.ip)
    setConnecting(true)
    setError(null)

    try {
      const sid = await invoke<string>('terminal_connect', {
        host: selectedInstance.ip,
        username: 'ubuntu',
      })
      console.log('[TerminalView] Connected with session ID:', sid)

      // Update refs IMMEDIATELY before state to avoid race conditions
      sessionIdRef.current = sid
      connectedRef.current = true

      // Now update state
      setSessionId(sid)
      setConnected(true)

      if (xtermRef.current) {
        xtermRef.current.clear()
        xtermRef.current.writeln(`\x1b[32m✓ Connected to ${selectedInstance.hostname || selectedInstance.ip}\x1b[0m`)
        xtermRef.current.writeln('')

        // Fit terminal and sync PTY size
        if (fitAddonRef.current) {
          fitAddonRef.current.fit()
          const { rows, cols } = xtermRef.current

          await invoke('terminal_resize', { sessionId: sid, rows, cols }).catch(_ => {})
        }

        // Focus terminal so user can type immediately
        xtermRef.current.focus()
        console.log('[TerminalView] Terminal focused and ready for input')
      }
    } catch (error) {
      const errorMsg = error instanceof Error ? error.message : String(error)
      console.error('[TerminalView] Connection failed:', errorMsg)
      setError(`Connection failed: ${errorMsg}`)
      if (xtermRef.current) {
        xtermRef.current.writeln(`\x1b[31m✗ Connection failed: ${errorMsg}\x1b[0m`)
      }
    } finally {
      setConnecting(false)
    }
  }, [selectedInstance, xtermRef, fitAddonRef, sessionIdRef, connectedRef])

  // Auto-connect to first available instance - simplified with proper cleanup
  useEffect(() => {
    // Only run if we haven't attempted auto-connect yet
    if (autoConnectAttemptedRef.current) {
      return
    }

    // Check all required conditions
    if (
      instances.length > 0 &&
      !connected &&
      !connecting &&
      !sessionId &&
      !selectedInstance &&
      xtermRef.current
    ) {
      const activeInstances = instances.filter(i => i.status === 'active')
      if (activeInstances.length > 0) {
        console.log('[TerminalView] Auto-connecting to first instance:', activeInstances[0])
        autoConnectAttemptedRef.current = true
        const firstInstance = activeInstances[0]
        setSelectedInstance(firstInstance)

        // Auto-connect after a brief delay - store timeout for cleanup
        autoConnectTimeoutRef.current = window.setTimeout(async () => {
          try {
            console.log('[TerminalView] Executing auto-connect...')
            await connectToInstance()
          } catch (error) {
            console.error('[TerminalView] Auto-connect failed:', error)
          }
        }, 1500)
      }
    }

    // Cleanup timeout on unmount or when dependencies change
    return () => {
      if (autoConnectTimeoutRef.current !== null) {
        clearTimeout(autoConnectTimeoutRef.current)
        autoConnectTimeoutRef.current = null
      }
    }
  }, [instances.length, connected, connecting, sessionId, selectedInstance])

  // Sync refs with state to avoid stale closures
  useEffect(() => {
    connectedRef.current = connected
    sessionIdRef.current = sessionId
  }, [connected, sessionId])

  async function loadInstances() {
    setLoading(true)
    try {
      const instancesData = await invoke<Instance[]>('lambda_list_instances')
      setInstances(instancesData)
    } catch (error) {

    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    if (!terminalRef.current) return

    // Initialize terminal
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

    // Handle window resize
    const handleResize = () => {
      fitAddon.fit()
      // Sync PTY size with visual terminal size
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

    // Handle terminal input - use refs to avoid stale closure
    term.onData((data) => {
      console.log('[TerminalView] Input received:', data.length, 'bytes, connected:', connectedRef.current, 'sessionId:', sessionIdRef.current)
      if (connectedRef.current && sessionIdRef.current) {
        invoke('terminal_input', { sessionId: sessionIdRef.current, data })
          .then(() => {
            console.log('[TerminalView] Input sent successfully')
          })
          .catch((error) => {
            console.error('[TerminalView] Input send failed:', error)
            term.writeln(`\x1b[31m✗ Input failed: ${error}\x1b[0m`)
          })
      } else {
        console.warn('[TerminalView] Input ignored - not connected. connected:', connectedRef.current, 'sessionId:', sessionIdRef.current)
      }
    })

    return () => {
      window.removeEventListener('resize', handleResize)
      term.dispose()
    }
  }, [])

  useEffect(() => {
    if (!xtermRef.current) return

    // Listen for terminal output
    const unlisten = listen<TerminalOutput>('terminal_output', (event) => {
      if (event.payload.data && xtermRef.current) {
        xtermRef.current.write(event.payload.data)
      }
    })

    return () => {
      unlisten.then((fn) => fn())
    }
  }, [])

  const disconnectFromInstance = async () => {
    if (sessionId) {
      try {
        await invoke('terminal_disconnect', { sessionId })
        setConnected(false)
        setSessionId(null)
        if (xtermRef.current) {
          xtermRef.current.writeln('')
          xtermRef.current.writeln('\x1b[33m✓ Disconnected\x1b[0m')
        }
      } catch (error) {

      }
    }
  }

  const activeInstances = instances.filter(i => i.status === 'active')

  return (
    <div className="h-full flex overflow-hidden bg-gray-950">
      {/* Sidebar with Instance List */}
      <div className="w-80 border-r border-gray-800 bg-gray-900/50 flex flex-col">
        {/* Sidebar Header */}
        <div className="border-b border-gray-800 p-4">
          <div className="flex items-center justify-between mb-3">
            <h3 className="text-sm font-semibold text-gray-300">Active Instances</h3>
            <button
              onClick={loadInstances}
              className="p-1 rounded hover:bg-gray-800 transition-colors"
              title="Refresh"
            >
              <span className="text-gray-400">🔄</span>
            </button>
          </div>
          <p className="text-xs text-gray-500">
            {activeInstances.length} instance{activeInstances.length !== 1 ? 's' : ''} available
          </p>
        </div>

        {/* Instance List */}
        <div className="flex-1 overflow-y-auto p-2">
          {loading ? (
            <div className="flex items-center justify-center h-32 text-gray-500 text-sm">
              Loading instances...
            </div>
          ) : activeInstances.length === 0 ? (
            <div className="flex flex-col items-center justify-center h-32 text-gray-500 text-sm px-4 text-center">
              <span className="text-2xl mb-2">☁️</span>
              <p>No active instances</p>
              <p className="text-xs mt-1">Launch an instance in Lambda tab</p>
            </div>
          ) : (
            <div className="space-y-2">
              {activeInstances.map((instance) => {
                const isSelected = selectedInstance?.id === instance.id
                return (
                  <button
                    key={instance.id}
                    onClick={() => setSelectedInstance(instance)}
                    className={`w-full text-left p-3 rounded-lg border transition-all ${
                      isSelected
                        ? 'border-electric-500 bg-electric-500/10 anime-glow'
                        : 'border-gray-800 bg-gray-900/50 hover:border-gray-700 hover:bg-gray-800/50'
                    }`}
                  >
                    <div className="flex items-start justify-between mb-2">
                      <div className="flex-1 min-w-0">
                        <div className="font-medium text-sm text-gray-200 truncate">
                          {instance.hostname || instance.id}
                        </div>
                        <div className="text-xs text-gray-400 truncate">
                          {instance.ip}
                        </div>
                      </div>
                      {isSelected && connected && (
                        <div className="ml-2 flex-shrink-0">
                          <div className="w-2 h-2 bg-mint-500 rounded-full animate-pulse" />
                        </div>
                      )}
                    </div>
                    <div className="text-xs text-gray-500">
                      {instance.instance_type.description}
                    </div>
                    <div className="text-xs text-gray-600 mt-1">
                      {instance.region.description}
                    </div>
                  </button>
                )
              })}
            </div>
          )}
        </div>
      </div>

      {/* Main Terminal Area */}
      <div className="flex-1 flex flex-col overflow-hidden">
        {/* Header */}
        <div className="border-b border-gray-800 bg-gray-900/80 backdrop-blur-md p-4">
          {error && (
            <div className="mb-3 p-3 bg-sunset-500/10 border border-sunset-500/30 rounded-lg text-sunset-400 text-sm flex items-center gap-2">
              <span>⚠️</span>
              <span>{error}</span>
              <button onClick={() => setError(null)} className="ml-auto text-sunset-300 hover:text-sunset-200">✕</button>
            </div>
          )}
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <span className="text-2xl">💻</span>
              <div>
                <h2 className="text-lg font-bold text-gray-200">SSH Terminal</h2>
                {selectedInstance ? (
                  <p className="text-xs text-gray-400">
                    {selectedInstance.hostname || selectedInstance.ip} • {selectedInstance.instance_type.description}
                  </p>
                ) : (
                  <p className="text-xs text-gray-400">Select an instance to connect</p>
                )}
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
                  onClick={connectToInstance}
                  disabled={!selectedInstance || connecting}
                  className="px-4 py-2 bg-electric-500/20 hover:bg-electric-500/30 border border-electric-500/50 rounded-lg text-electric-400 font-medium transition-all disabled:opacity-50 disabled:cursor-not-allowed text-sm"
                >
                  {connecting ? '⏳ Connecting...' : '🔌 Connect'}
                </button>
              ) : (
                <button
                  onClick={disconnectFromInstance}
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
          {!selectedInstance ? (
            <div className="h-full flex items-center justify-center text-gray-500">
              <div className="text-center">
                <span className="text-6xl mb-4 block">💻</span>
                <p className="text-lg font-medium text-gray-400 mb-2">No Instance Selected</p>
                <p className="text-sm">Select an active instance from the sidebar to start an SSH session</p>
              </div>
            </div>
          ) : (
            <div
              ref={terminalRef}
              onClick={() => xtermRef.current?.focus()}
              className="h-full w-full rounded-lg border border-gray-800 bg-[#0a0a0a] p-2 cursor-text"
              style={{ minHeight: '400px' }}
            />
          )}
        </div>

        {/* Footer */}
        <div className="border-t border-gray-800 bg-gray-900/80 backdrop-blur-md px-4 py-2">
          <div className="flex items-center justify-between text-xs text-gray-500">
            <div className="flex items-center gap-4">
              <span>⌨️ Full SSH terminal</span>
              <span>•</span>
              <span>Ctrl+C to interrupt</span>
              <span>•</span>
              <span>Ctrl+D to exit</span>
            </div>
            {selectedInstance && !connected && (
              <span className="text-gray-400">
                Click Connect to start SSH session
              </span>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
