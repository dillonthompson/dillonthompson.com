import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import { usePageTitle } from '@/hooks/use-page-title'
import { MapPin, Mail } from 'lucide-react'

interface About {
  name?: string
  title?: string
  location?: string
  email?: string
  linkedin?: string
  github?: string
  summary?: string
  // Fields added by migration 00004:
  bio?: string[]
  personal?: string
  interests?: string[]
}

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

export default function AboutPage() {
  usePageTitle('About')
  const [about, setAbout] = useState<About | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetch('/api/v1/profile/about')
      .then((res) => {
        if (!res.ok) throw new Error(`HTTP ${res.status}`)
        return res.json()
      })
      .then(setAbout)
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false))
  }, [])

  return (
    <div className="max-w-3xl mx-auto px-6 py-16">
      <Button variant="ghost" size="sm" className="mb-8" render={<Link to="/" />}>
        ← Back
      </Button>

      <h1 className="text-4xl sm:text-5xl font-bold tracking-tight mb-2">About</h1>
      {about?.title && (
        <p className="text-lg text-muted-foreground mb-2">{about.title}</p>
      )}
      {about?.location && (
        <p className="flex items-center gap-1.5 text-sm text-muted-foreground mb-12">
          <MapPin size={14} />
          {about.location}
        </p>
      )}

      {loading && <p className="text-muted-foreground">Loading...</p>}
      {error && <p className="text-destructive">Failed to load: {error}</p>}

      {/* Professional bio */}
      {about?.bio && about.bio.length > 0 && (
        <section className="space-y-5 text-base text-foreground/90 leading-relaxed">
          {about.bio.map((paragraph, i) => (
            <p key={i}>{paragraph}</p>
          ))}
        </section>
      )}

      {/* Personal narrative */}
      {(about?.personal || (about?.interests && about.interests.length > 0)) && (
        <>
          <Separator className="my-12" />
          <section>
            <h2 className="text-2xl font-bold tracking-tight mb-6">Beyond the keyboard</h2>
            {about.personal && (
              <p className="text-base text-foreground/90 leading-relaxed mb-6">
                {about.personal}
              </p>
            )}
            {about.interests && about.interests.length > 0 && (
              <div>
                <h3 className="text-xs font-medium text-muted-foreground uppercase tracking-wider mb-3">
                  Things I'm into
                </h3>
                <div className="flex flex-wrap gap-1.5">
                  {about.interests.map((interest) => (
                    <Badge key={interest} variant="secondary">
                      {interest}
                    </Badge>
                  ))}
                </div>
              </div>
            )}
          </section>
        </>
      )}

      {/* Contact CTA */}
      {(about?.email || about?.github || about?.linkedin) && (
        <>
          <Separator className="my-12" />
          <section>
            <h2 className="text-2xl font-bold tracking-tight mb-2">Let's talk</h2>
            <p className="text-muted-foreground mb-6">
              I'm available for senior full-stack roles and consulting engagements.
            </p>
            <div className="flex flex-wrap gap-3">
              {about.email && (
                <Button size="lg" className="cursor-pointer gap-2" render={<a href={`mailto:${about.email}`} />}>
                  <Mail size={16} />
                  Email
                </Button>
              )}
              {about.github && (
                <Button variant="outline" size="lg" className="cursor-pointer gap-2" render={<a href={`https://${about.github}`} target="_blank" rel="noopener noreferrer" />}>
                  <GitHubIcon />
                  GitHub
                </Button>
              )}
              {about.linkedin && (
                <Button variant="outline" size="lg" className="cursor-pointer gap-2" render={<a href={`https://${about.linkedin}`} target="_blank" rel="noopener noreferrer" />}>
                  <LinkedInIcon />
                  LinkedIn
                </Button>
              )}
            </div>
          </section>
        </>
      )}
    </div>
  )
}
