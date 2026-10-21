import { useTheme } from '@/hooks/use-theme'
import { Button } from '@/components/ui/button'

const svgProps = {
  xmlns: 'http://www.w3.org/2000/svg',
  width: 16,
  height: 16,
  viewBox: '0 0 24 24',
  fill: 'none',
  stroke: 'currentColor',
  strokeWidth: 2,
  strokeLinecap: 'round',
  strokeLinejoin: 'round',
} as const

// The icon shows where the toggle takes you: sunrise while it's sunset, and
// sunset while it's sunrise.
export function ThemeToggle() {
  const { theme, toggleTheme } = useTheme()
  const toSunrise = theme === 'dark'

  return (
    <Button
      variant="ghost"
      size="icon"
      onClick={toggleTheme}
      aria-label={toSunrise ? 'Switch to sunrise theme' : 'Switch to sunset theme'}
    >
      <svg {...svgProps}>
        <path d={toSunrise ? 'M12 2v8' : 'M12 10V2'} />
        <path d="m4.93 10.93 1.41 1.41M2 18h2M20 18h2m-.93-7.07-1.41 1.41M22 22H2" />
        <path d={toSunrise ? 'm8 6 4-4 4 4' : 'm16 6-4 4-4-4'} />
        <path d="M16 18a4 4 0 0 0-8 0" />
      </svg>
    </Button>
  )
}
