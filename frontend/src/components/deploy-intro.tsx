import { useState, useEffect, useCallback } from 'react'
import { Button } from '@/components/ui/button'

interface Step {
  loading: string
  done: string
  duration: number
}

// All quotes from shows/games — grouped by progression feel

// Step 1: Starting up / setting the scene
const step1Pool = [
  { loading: '"Cool. Cool cool cool."', done: 'Cool' },                                        // Community
  { loading: '"Mathematical!"', done: 'Algebraic' },                                           // Adventure Time
  { loading: '"What time is it?"', done: 'Adventure Time' },                                   // Adventure Time
]

// Step 2: Building / working through it
const step2Pool = [
  { loading: '"Sucking at something is the first step..."', done: 'to being sorta good' },          // Adventure Time — Jake
  { loading: '"Homies help homies. Always."', done: 'Always' },                                    // Adventure Time — Finn
  { loading: '"I need a weapon."', done: 'Acquired' },                                             // Halo — Master Chief
]

// Step 3: Getting closer / refining
const step3Pool = [
  { loading: '"Sometimes science is more art than science."', done: 'A lot of people don\'t get that' },  // Rick and Morty
  { loading: '"The clue was the friends we made along the way."', done: 'That\'s the whole clue' },       // Rick and Morty
  { loading: '"The boss has been summoned..."', done: 'Preparing defenses' },                              // Valheim
]

// Step 4: Ready / arrival
const step4Pool = [
  { loading: '"To live is to risk it all."', done: 'Otherwise you\'re just an inert chunk' },       // Rick and Morty
  { loading: '"Welcome to Greendale. You\'re already accepted."', done: 'Go Human Beings' },        // Community
  { loading: '"Six seasons and a movie!"', done: '#andamovie' },                                    // Community
  { loading: '"Odin welcomes you..."', done: 'The portal is open' },                                // Valheim
]

function pickOne<T>(pool: T[]): T {
  return pool[Math.floor(Math.random() * pool.length)]
}

function buildSteps(): Step[] {
  const durations = [1700, 1900, 1500, 2000]
  return [step1Pool, step2Pool, step3Pool, step4Pool].map((pool, i) => ({
    ...pickOne(pool),
    duration: durations[i],
  }))
}

const steps: Step[] = buildSteps()

const spinnerFrames = ['⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏']

function Spinner() {
  const [frame, setFrame] = useState(0)
  useEffect(() => {
    const id = setInterval(() => setFrame((f) => (f + 1) % spinnerFrames.length), 80)
    return () => clearInterval(id)
  }, [])
  return <span className="text-muted-foreground">{spinnerFrames[frame]}</span>
}

interface DeployIntroProps {
  onComplete: () => void
}

export function DeployIntro({ onComplete }: DeployIntroProps) {
  const [currentStep, setCurrentStep] = useState(0)
  const [stepDone, setStepDone] = useState(false)
  const [allDone, setAllDone] = useState(false)
  const [showButton, setShowButton] = useState(false)
  const [elapsed, setElapsed] = useState(0)
  const [exiting, setExiting] = useState(false)

  // Build timer
  useEffect(() => {
    if (allDone) return
    const id = setInterval(() => setElapsed((e) => e + 0.1), 100)
    return () => clearInterval(id)
  }, [allDone])

  // Step sequencer
  useEffect(() => {
    if (currentStep >= steps.length) return

    setStepDone(false)
    const timer = setTimeout(() => {
      setStepDone(true)
      setTimeout(() => {
        if (currentStep < steps.length - 1) {
          setCurrentStep((s) => s + 1)
        } else {
          setAllDone(true)
          setTimeout(() => setShowButton(true), 500)
        }
      }, 200)
    }, steps[currentStep].duration)

    return () => clearTimeout(timer)
  }, [currentStep])

  const handleEnter = useCallback(() => {
    setExiting(true)
    setTimeout(onComplete, 500)
  }, [onComplete])

  return (
    <div className={`fixed inset-0 z-[100] bg-background flex items-center justify-center transition-opacity duration-500 ${exiting ? 'opacity-0' : 'opacity-100'}`}>
      <div className="w-full max-w-lg px-6">
        {/* Welcome message */}
        <div className="mb-8">
          <h1 className="text-2xl font-bold tracking-tight mb-1">dillonthompson.com</h1>
          <p className="text-sm text-muted-foreground">Building your experience.</p>
        </div>

        {/* Build log */}
        <div className="font-mono text-sm space-y-2 mb-8">
          {steps.map((step, i) => {
            if (i > currentStep) return null
            const isDone = i < currentStep || (i === currentStep && stepDone)
            return (
              <div key={i} className={`flex items-start gap-2 transition-opacity duration-300 ${i > currentStep ? 'opacity-0' : 'opacity-100'}`}>
                {isDone ? (
                  <span className="text-green-500">✓</span>
                ) : (
                  <Spinner />
                )}
                <span className={isDone ? 'text-muted-foreground' : 'text-foreground'}>
                  {isDone ? step.done : step.loading}
                </span>
              </div>
            )
          })}
        </div>

        {/* Live indicator */}
        {allDone && (
          <div className="animate-in fade-in duration-500">
            {/* Status line */}
            <div className="flex items-center justify-center gap-3 py-6">
              <div className="w-3 h-3 rounded-full bg-green-500 shadow-[0_0_12px_rgba(34,197,94,0.6)]" />
              <span className="font-mono text-base text-foreground font-medium">
                dillonthompson.com is live
              </span>
            </div>

            {/* Build time */}
            <div className="text-center mb-6">
              <span className="font-mono text-xs text-muted-foreground/50">
                Build time: {elapsed.toFixed(1)}s
              </span>
            </div>

            {/* Enter button with radiating color rings */}
            {showButton && (
              <div className="flex justify-center animate-in fade-in slide-in-from-bottom-4 duration-500">
                <Button
                  size="lg"
                  onClick={handleEnter}
                  className="font-mono gap-2 cursor-pointer border-2 animate-[border-glow_6s_ease-in-out_infinite]"
                >
                  Visit Site
                  <span className="text-xs opacity-60">→</span>
                </Button>
              </div>
            )}
          </div>
        )}

        {/* Elapsed timer in corner */}
        {!allDone && (
          <div className="fixed bottom-4 right-4 font-mono text-xs text-muted-foreground/30">
            {elapsed.toFixed(1)}s
          </div>
        )}
      </div>
    </div>
  )
}
