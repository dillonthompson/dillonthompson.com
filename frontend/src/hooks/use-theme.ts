import { useEffect, useRef, useState } from 'react'
import { flushSync } from 'react-dom'
import { runThemeTransition } from '@/lib/theme-transition'

// 'dark' is the sunset theme (default), 'light' is the sunrise theme. The
// stored values predate the rename and are kept so returning visitors' choice
// survives.
type Theme = 'light' | 'dark'

function getInitialTheme(): Theme {
  try {
    return localStorage.getItem('theme') === 'light' ? 'light' : 'dark'
  } catch {
    return 'dark'
  }
}

// Browser toolbar colors, matching each theme's --background.
const THEME_COLOR: Record<Theme, string> = { dark: '#07162c', light: '#d9efff' }

function applyTheme(theme: Theme) {
  document.documentElement.classList.toggle('dark', theme === 'dark')
  document.querySelector('meta[name="theme-color"]')?.setAttribute('content', THEME_COLOR[theme])
  try {
    localStorage.setItem('theme', theme)
  } catch {
    // Storage can be blocked (private mode, quota); the theme still applies.
  }
}

export function useTheme() {
  const [theme, setTheme] = useState<Theme>(getInitialTheme)
  // Where the toggle is headed. `theme` only updates a frame into the
  // transition, so a quick second click must read this, not `theme`.
  const target = useRef(theme)

  useEffect(() => {
    applyTheme(theme)
  }, [theme])

  const toggleTheme = () => {
    const next: Theme = target.current === 'dark' ? 'light' : 'dark'
    target.current = next
    // Flip the DOM (and flush React, so the toggle icon updates too) inside the
    // view transition so it snapshots old -> new.
    runThemeTransition(next === 'dark', () => {
      applyTheme(next)
      flushSync(() => setTheme(next))
    })
  }

  return { theme, toggleTheme }
}
