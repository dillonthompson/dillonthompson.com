// Animated sunrise/sunset theme swap, built on the View Transitions API.
//
// The `dark` class is the sunset theme, no class is sunrise. The sun travels an
// arc east to west (left to right): it rises from the left horizon to the top
// center for the sunrise, and continues from the top center down to the right
// horizon for the sunset. The theme change is a circular wash centred on the
// sun — the new theme grows out of it going to sunrise, the old theme shrinks
// into it going to sunset. Browsers without View Transitions (or users who
// prefer reduced motion) just get the instant swap.
//
// internal/handlers/blogassets/theme.js has a vanilla copy for the blog — keep
// them in sync.

const DURATION_MS = 1300
// Keyframes are interpolated linearly, so sample the eased arc about every 8ms
// (finer than a 120Hz frame) or the circle visibly steps as it follows the sun.
const SAMPLES = Math.ceil(DURATION_MS / 8)
const SUN_ID = 'theme-sun'

// The transition currently owning the root classes. A superseded transition's
// cleanup must not strip the classes the newer one just set.
let active: object | null = null

const easeInOut = (t: number) => (t < 0.5 ? 4 * t * t * t : 1 - Math.pow(-2 * t + 2, 3) / 2)

export function runThemeTransition(toSunset: boolean, apply: () => void) {
  const root = document.documentElement
  const reduceMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches
  if (reduceMotion || typeof document.startViewTransition !== 'function') {
    apply()
    return
  }

  const token = {}
  active = token

  // The sun only exists in the new state; the browser snapshots it as its own
  // layer (view-transition-name) above the page, and we animate that layer.
  let sun: HTMLDivElement | undefined
  const transition = document.startViewTransition(() => {
    // Scope the transition CSS to this animation so it can't touch any other
    // view transition.
    root.classList.add('theme-anim')
    root.classList.toggle('theme-to-sunset', toSunset)
    apply()
    sun = document.createElement('div')
    sun.id = SUN_ID
    sun.setAttribute('aria-hidden', 'true')
    document.body.appendChild(sun)
  })

  // fill: 'forwards' animations on <html> outlive the transition and would keep
  // stacking onto the next one (getting choppier each toggle), so cancel them.
  const animations: Animation[] = []
  const cleanup = () => {
    animations.forEach((a) => a.cancel())
    sun?.remove()
    if (active === token) {
      active = null
      root.classList.remove('theme-anim', 'theme-to-sunset')
    }
  }

  transition.ready
    .then(() => {
      const w = window.innerWidth
      const h = window.innerHeight
      const size = sun?.offsetWidth ?? 224

      // Sun centre along a half-ellipse over the horizon (bottom edge). Sunrise
      // covers the left quarter-arc, sunset the right one.
      const rx = w * 0.4
      const ry = h * 0.7
      const sunAt = (p: number) => {
        const phi = ((toSunset ? 0.5 : 0) + p * 0.5) * Math.PI
        return { x: w / 2 - rx * Math.cos(phi), y: h - ry * Math.sin(phi) + size * 0.2 }
      }
      // Radius needed for a circle at (x, y) to cover the viewport.
      const reach = (x: number, y: number) =>
        Math.ceil(Math.hypot(Math.max(x, w - x), Math.max(y, h - y)))

      const sunFrames: Keyframe[] = []
      const clipFrames: Keyframe[] = []
      for (let i = 0; i <= SAMPLES; i++) {
        const p = easeInOut(i / SAMPLES)
        const { x, y } = sunAt(p)
        sunFrames.push({ transform: `translate(${x - size / 2}px, ${y - size / 2}px)` })
        // Sunrise: wash grows 0 -> full. Sunset: old theme shrinks full -> 0.
        const r = reach(x, y) * (toSunset ? 1 - p : p)
        clipFrames.push({ clipPath: `circle(${r}px at ${x}px ${y}px)` })
      }

      // fill: 'forwards' holds the final state until teardown, otherwise the
      // layers snap back for a frame (flicker).
      const opts = { duration: DURATION_MS, easing: 'linear', fill: 'forwards' } as const
      animations.push(
        root.animate(clipFrames, {
          ...opts,
          pseudoElement: toSunset ? '::view-transition-old(root)' : '::view-transition-new(root)',
        }),
        root.animate(sunFrames, { ...opts, pseudoElement: `::view-transition-group(${SUN_ID})` }),
        // Fade the sun in/out so it never pops on or off at full strength.
        root.animate(
          { opacity: [0, 1, 1, 0], offset: [0, 0.15, 0.6, 1] },
          { ...opts, pseudoElement: `::view-transition-image-pair(${SUN_ID})` },
        ),
      )
    })
    .catch(cleanup)

  transition.finished.then(cleanup, cleanup)
}
