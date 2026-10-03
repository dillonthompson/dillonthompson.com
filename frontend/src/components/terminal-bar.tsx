import { useState, useRef, useEffect, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'

const COMMANDS: Record<string, { description: string; action: 'navigate' | 'response'; value: string }> = {
  help:    { description: 'List available commands', action: 'response', value: '' },
  home:    { description: 'Go home', action: 'navigate', value: '/' },
  about:   { description: 'About me', action: 'navigate', value: '/about' },
  contact: { description: 'Get in touch', action: 'response', value: 'Email me: dj.thompson715@gmail.com' },
  whoami:  { description: '???', action: 'response', value: 'Dillon Thompson — Senior Full Stack Engineer & Technical Consultant' },
  clear:   { description: 'Clear output', action: 'response', value: '' },
}

interface TerminalLine {
  type: 'input' | 'output' | 'error'
  text: string
}

function helpText(): string {
  const lines = Object.entries(COMMANDS).map(
    ([cmd, { description }]) => `  ${cmd.padEnd(12)} ${description}`
  )
  return 'Available commands (type one and hit enter):\n' + lines.join('\n')
}

export function TerminalBar() {
  const [input, setInput] = useState('')
  const [expanded, setExpanded] = useState(false)
  const [history, setHistory] = useState<TerminalLine[]>([])
  const [cmdHistory, setCmdHistory] = useState<string[]>([])
  const [historyIdx, setHistoryIdx] = useState(-1)
  const inputRef = useRef<HTMLInputElement>(null)
  const outputRef = useRef<HTMLDivElement>(null)
  // Tracks whether we've auto-displayed help yet. We only do this once per
  // session — re-showing it after `clear` or after the user closes and reopens
  // would feel intrusive.
  const helpShown = useRef(false)
  const navigate = useNavigate()

  const scrollToBottom = useCallback(() => {
    if (outputRef.current) {
      outputRef.current.scrollTop = outputRef.current.scrollHeight
    }
  }, [])

  useEffect(scrollToBottom, [history, scrollToBottom])

  const executeCommand = useCallback((raw: string) => {
    const trimmed = raw.trim().toLowerCase()
    if (!trimmed) return

    const newHistory = [...history, { type: 'input' as const, text: `> ${trimmed}` }]
    setCmdHistory((prev) => [trimmed, ...prev])
    setHistoryIdx(-1)

    if (trimmed === 'clear') {
      setHistory([])
      return
    }

    if (trimmed === 'help') {
      setHistory([...newHistory, { type: 'output', text: helpText() }])
      return
    }

    if (trimmed === 'sudo hire dillon') {
      setHistory([...newHistory, { type: 'output', text: "Great choice. Let's talk → dj.thompson715@gmail.com" }])
      return
    }

    if (trimmed === 'neofetch') {
      const info = [
        '  dillon@portfolio',
        '  ─────────────────',
        '  OS:     macOS / Linux',
        '  Stack:  Go, React, Postgres',
        '  Editor: VS Code + Claude Code',
        '  Shell:  fish',
        '  Theme:  dark',
      ].join('\n')
      setHistory([...newHistory, { type: 'output', text: info }])
      return
    }

    const cmd = COMMANDS[trimmed]
    if (!cmd) {
      setHistory([...newHistory, { type: 'error', text: `command not found: ${trimmed}. Type 'help' for available commands.` }])
      return
    }

    if (cmd.action === 'navigate') {
      setHistory([...newHistory, { type: 'output', text: `→ navigating to ${cmd.value}` }])
      setTimeout(() => {
        navigate(cmd.value)
        setExpanded(false)
      }, 300)
      return
    }

    if (cmd.value) {
      setHistory([...newHistory, { type: 'output', text: cmd.value }])
    }
  }, [history, navigate])

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      executeCommand(input)
      setInput('')
    } else if (e.key === 'Escape') {
      setExpanded(false)
      inputRef.current?.blur()
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      if (cmdHistory.length > 0) {
        const next = Math.min(historyIdx + 1, cmdHistory.length - 1)
        setHistoryIdx(next)
        setInput(cmdHistory[next])
      }
    } else if (e.key === 'ArrowDown') {
      e.preventDefault()
      if (historyIdx > 0) {
        const next = historyIdx - 1
        setHistoryIdx(next)
        setInput(cmdHistory[next])
      } else {
        setHistoryIdx(-1)
        setInput('')
      }
    } else if (e.key === 'Tab') {
      e.preventDefault()
      const match = Object.keys(COMMANDS).find((cmd) => cmd.startsWith(input.toLowerCase()))
      if (match) setInput(match)
    }
  }

  // Global Cmd+K / Ctrl+K handler
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault()
        setExpanded(true)
        setTimeout(() => inputRef.current?.focus(), 50)
      }
    }
    window.addEventListener('keydown', handler)
    return () => window.removeEventListener('keydown', handler)
  }, [])

  // Surface the help listing the first time the user opens the terminal so the
  // available commands are obvious. Subsequent opens keep whatever's there —
  // including an intentionally empty pane after `clear` — so it doesn't feel
  // chatty.
  useEffect(() => {
    if (expanded && !helpShown.current) {
      helpShown.current = true
      setHistory((prev) => [...prev, { type: 'output', text: helpText() }])
    }
  }, [expanded])

  return (
    <div className="w-full max-w-2xl mx-auto">
      {/* Input bar */}
      <div
        className="flex items-center gap-2 bg-muted/50 border border-border rounded-lg px-3 py-2 font-mono text-sm cursor-text transition-colors hover:border-foreground/20 focus-within:border-foreground/30"
        onClick={() => { setExpanded(true); inputRef.current?.focus() }}
      >
        <span className="text-muted-foreground select-none shrink-0">~$</span>
        <input
          ref={inputRef}
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={handleKeyDown}
          onFocus={() => setExpanded(true)}
          placeholder="type a command..."
          className="flex-1 bg-transparent outline-none text-foreground placeholder:text-muted-foreground/50 caret-foreground"
          spellCheck={false}
          autoComplete="off"
        />
        <kbd className="hidden sm:inline text-xs text-muted-foreground/40 border border-border/50 rounded px-1.5 py-0.5">
          ⌘K
        </kbd>
      </div>

      {/* Expanded output */}
      {expanded && history.length > 0 && (
        <div
          ref={outputRef}
          className="mt-1 bg-muted/50 border border-border rounded-lg p-3 font-mono text-xs max-h-64 overflow-y-auto animate-in fade-in slide-in-from-top-2 duration-200"
        >
          {history.map((line, i) => (
            <div
              key={i}
              className={
                line.type === 'input' ? 'text-muted-foreground' :
                line.type === 'error' ? 'text-destructive' :
                'text-foreground'
              }
            >
              <pre className="whitespace-pre-wrap">{line.text}</pre>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
