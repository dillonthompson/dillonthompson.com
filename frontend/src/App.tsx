import { useEffect, useState, useRef } from 'react'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import { DeployIntro } from '@/components/deploy-intro'
import { usePageTitle } from '@/hooks/use-page-title'
import { MapPin } from 'lucide-react'

function GitHubIcon({ size = 16 }: { size?: number }) {
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="currentColor">
      <path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0024 12c0-6.63-5.37-12-12-12z" />
    </svg>
  )
}

function LinkedInIcon({ size = 16 }: { size?: number }) {
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="currentColor">
      <path d="M20.447 20.452h-3.554v-5.569c0-1.328-.027-3.037-1.852-3.037-1.853 0-2.136 1.445-2.136 2.939v5.667H9.351V9h3.414v1.561h.046c.477-.9 1.637-1.85 3.37-1.85 3.601 0 4.267 2.37 4.267 5.455v6.286zM5.337 7.433a2.062 2.062 0 01-2.063-2.065 2.064 2.064 0 112.063 2.065zm1.782 13.019H3.555V9h3.564v11.452zM22.225 0H1.771C.792 0 0 .774 0 1.729v20.542C0 23.227.792 24 1.771 24h20.451C23.2 24 24 23.227 24 22.271V1.729C24 .774 23.2 0 22.222 0h.003z" />
    </svg>
  )
}

interface Profile {
  about?: {
    name: string
    title: string
    summary: string
    location: string
    email: string
    linkedin: string
    github: string
  }
  skills?: {
    languages: string[]
    frameworks: string[]
    tools_devops: string[]
  }
  education?: {
    degree: string
    school: string
    graduated: string
    gpa: string
    honors: string[]
    certifications?: { name: string; issuer: string; year: string }[]
  }
}

interface Experience {
  id: string
  company: string
  role: string
  start_date: string
  end_date: string | null
  description: string[]
  tech_stack: string[]
  security_highlights: string[]
  metadata: Record<string, unknown>
}

function formatDuration(startDate: string, endDate: string | null): string {
  const startYear = parseInt(startDate.split('-')[0], 10)
  if (!endDate) return `${startYear}–Present`
  const endYear = parseInt(endDate.split('-')[0], 10)
  return startYear === endYear ? `${startYear}` : `${startYear}–${endYear}`
}

interface ExpandCardProps {
  title: string
  teaser: string
  isOpen: boolean
  onToggle: () => void
  children: React.ReactNode
}

function ExpandCard({ title, teaser, isOpen, onToggle, children }: ExpandCardProps) {
  const contentRef = useRef<HTMLDivElement>(null)
  const [height, setHeight] = useState(0)

  useEffect(() => {
    if (contentRef.current) {
      setHeight(contentRef.current.scrollHeight)
    }
  }, [isOpen, children])

  return (
    <Card
      className={`cursor-pointer transition-all duration-300 ${isOpen ? 'ring-1 ring-foreground/10 scale-[1.01]' : 'hover:ring-1 hover:ring-foreground/5'}`}
      onClick={onToggle}
    >
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle className="text-lg">{title}</CardTitle>
            {!isOpen && <CardDescription className="mt-1">{teaser}</CardDescription>}
          </div>
          <span className={`text-muted-foreground transition-transform duration-300 ${isOpen ? 'rotate-45' : ''}`}>
            +
          </span>
        </div>
      </CardHeader>
      <div
        className="overflow-hidden transition-all duration-400 ease-in-out"
        style={{ maxHeight: isOpen ? height : 0 }}
      >
        <div ref={contentRef}>
          <CardContent className="pt-0 pb-2">
            {children}
          </CardContent>
        </div>
      </div>
    </Card>
  )
}

const ALWAYS_SHOW_INTRO = import.meta.env.VITE_ALWAYS_SHOW_INTRO === 'true'

function App() {
  usePageTitle()
  const [showIntro, setShowIntro] = useState(() =>
    ALWAYS_SHOW_INTRO || !sessionStorage.getItem('intro_seen')
  )
  const [profile, setProfile] = useState<Profile>({})
  const [experiences, setExperiences] = useState<Experience[]>([])
  const [openCard, setOpenCard] = useState<string | null>(null)

  useEffect(() => {
    fetch('/api/v1/profile')
      .then((r) => r.ok ? r.json() : {})
      .then(setProfile)
      .catch(() => {})

    fetch('/api/v1/experience')
      .then((r) => r.ok ? r.json() : [])
      .then(setExperiences)
      .catch(() => {})
  }, [])

  const about = profile.about
  const skills = profile.skills
  const education = profile.education

  const toggle = (key: string) => setOpenCard((prev) => prev === key ? null : key)

  if (showIntro) {
    return (
      <DeployIntro onComplete={() => {
        sessionStorage.setItem('intro_seen', '1')
        setShowIntro(false)
      }} />
    )
  }

  return (
    <div className="max-w-3xl mx-auto px-6">
      {/* Hero */}
      <section className="py-20 sm:py-28 text-center">
        <h1 className="text-4xl sm:text-5xl font-bold tracking-tight mb-3">
          {about?.name ?? 'Dillon Thompson'}
        </h1>
        <p className="text-lg text-muted-foreground mb-4">
          {about?.title ?? 'Senior Full Stack Engineer & Technical Consultant'}
        </p>
        <p className="text-base text-muted-foreground/80 mb-8 max-w-xl mx-auto">
          I build secure, high-scale systems and take products from 0 to 1.
        </p>
        <div className="flex flex-wrap gap-3 justify-center">
          {about?.email && (
            <Button size="lg" className="cursor-pointer" render={<a href={`mailto:${about.email}`} />}>
              Get in touch
            </Button>
          )}
          <Button
            variant="outline"
            size="lg"
            className="cursor-pointer"
            onClick={() => {
              setOpenCard('experience')
              document.getElementById('cards')?.scrollIntoView({ behavior: 'smooth' })
            }}
          >
            See my work
          </Button>
        </div>
        <div className="flex flex-wrap gap-3 justify-center mt-6">
          {about?.location && (
            <span className="flex items-center gap-1.5 text-sm text-muted-foreground px-3 py-1.5">
              <MapPin size={14} />
              {about.location}
            </span>
          )}
          {about?.github && (
            <Button variant="outline" size="sm" className="cursor-pointer gap-2" render={<a href={`https://${about.github}`} target="_blank" rel="noopener noreferrer" />}>
              <GitHubIcon />
              GitHub
            </Button>
          )}
          {about?.linkedin && (
            <Button variant="outline" size="sm" className="cursor-pointer gap-2" render={<a href={`https://${about.linkedin}`} target="_blank" rel="noopener noreferrer" />}>
              <LinkedInIcon />
              LinkedIn
            </Button>
          )}
        </div>
      </section>

      <Separator />

      {/* Expandable cards */}
      <section id="cards" className="py-16 space-y-4">
        {/* Experience */}
        <ExpandCard
          title="Experience"
          teaser="8 years building production systems at scale"
          isOpen={openCard === 'experience'}
          onToggle={() => toggle('experience')}
        >
          <div className="space-y-6 mt-2">
            {experiences.map((exp) => (
              <div key={exp.id} className="border-l-2 border-border pl-4">
                <div className="flex flex-col sm:flex-row sm:items-baseline sm:justify-between gap-1 mb-2">
                  <div>
                    <p className="font-medium text-foreground">{exp.role}</p>
                    <p className="text-sm text-muted-foreground">{exp.company}</p>
                  </div>
                  <Badge variant="outline" className="font-mono text-xs w-fit shrink-0">
                    {formatDuration(exp.start_date, exp.end_date)}
                  </Badge>
                </div>
                <ul className="space-y-1 mb-3">
                  {exp.description.map((item, i) => (
                    <li key={i} className="text-sm text-muted-foreground pl-4 relative before:content-['•'] before:absolute before:left-0 before:text-muted-foreground/40">
                      {item}
                    </li>
                  ))}
                </ul>
                <div className="flex flex-wrap gap-1.5">
                  {exp.tech_stack.map((tech) => (
                    <Badge key={tech} variant="secondary">{tech}</Badge>
                  ))}
                </div>
              </div>
            ))}
          </div>
        </ExpandCard>

        {/* Skills */}
        {skills && (
          <ExpandCard
            title="Skills"
            teaser="Go, React, Kubernetes, and more"
            isOpen={openCard === 'skills'}
            onToggle={() => toggle('skills')}
          >
            <div className="space-y-5 mt-2">
              <div>
                <h3 className="text-xs font-medium text-muted-foreground uppercase tracking-wider mb-2">Languages</h3>
                <div className="flex flex-wrap gap-1.5">
                  {skills.languages.map((s) => <Badge key={s} variant="secondary">{s}</Badge>)}
                </div>
              </div>
              <Separator />
              <div>
                <h3 className="text-xs font-medium text-muted-foreground uppercase tracking-wider mb-2">Frameworks</h3>
                <div className="flex flex-wrap gap-1.5">
                  {skills.frameworks.map((s) => <Badge key={s} variant="secondary">{s}</Badge>)}
                </div>
              </div>
              <Separator />
              <div>
                <h3 className="text-xs font-medium text-muted-foreground uppercase tracking-wider mb-2">Tools & DevOps</h3>
                <div className="flex flex-wrap gap-1.5">
                  {skills.tools_devops.map((s) => <Badge key={s} variant="secondary">{s}</Badge>)}
                </div>
              </div>
            </div>
          </ExpandCard>
        )}

        {/* Education */}
        {education && (
          <ExpandCard
            title="Education"
            teaser={`${education.degree} — ${education.school}`}
            isOpen={openCard === 'education'}
            onToggle={() => toggle('education')}
          >
            <div className="mt-2 space-y-4">
              <div>
                <p className="font-medium text-foreground">{education.degree}</p>
                <p className="text-sm text-muted-foreground mb-2">{education.school}</p>
                <div className="flex flex-wrap gap-4 text-sm text-muted-foreground">
                  <span>{education.graduated}</span>
                  <span>GPA: {education.gpa}</span>
                  {education.honors.map((h) => (
                    <Badge key={h} variant="outline">{h}</Badge>
                  ))}
                </div>
              </div>
              {education.certifications && education.certifications.length > 0 && (
                <>
                  <Separator />
                  <div>
                    <h3 className="text-xs font-medium text-muted-foreground uppercase tracking-wider mb-2">Certifications</h3>
                    <div className="space-y-2">
                      {education.certifications.map((cert) => (
                        <div key={cert.name} className="flex flex-col sm:flex-row sm:items-baseline sm:justify-between gap-1">
                          <p className="text-sm text-foreground">{cert.name}</p>
                          <span className="text-xs text-muted-foreground shrink-0">{cert.issuer}, {cert.year}</span>
                        </div>
                      ))}
                    </div>
                  </div>
                </>
              )}
            </div>
          </ExpandCard>
        )}
      </section>

      {/* Footer */}
      <footer className="py-8 border-t border-border text-center text-xs text-muted-foreground">
        &copy; {new Date().getFullYear()} Dillon Thompson
      </footer>
    </div>
  )
}

export default App
