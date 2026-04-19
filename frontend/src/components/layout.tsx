import { Outlet } from 'react-router-dom'
import { ThemeToggle } from '@/components/theme-toggle'
import { TerminalBar } from '@/components/terminal-bar'

export function Layout() {
  return (
    <div className="min-h-screen bg-background text-foreground">
      <header className="sticky top-0 z-50 border-b border-border bg-background/80 backdrop-blur-sm">
        <div className="max-w-5xl mx-auto flex items-center gap-3 px-4 py-3">
          <div className="flex-1">
            <TerminalBar />
          </div>
          <ThemeToggle />
        </div>
      </header>
      <main>
        <Outlet />
      </main>
    </div>
  )
}
