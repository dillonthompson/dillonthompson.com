import { useCallback, useEffect, useRef, useState } from 'react'
import { Link, useBlocker, useNavigate, useParams } from 'react-router-dom'
import { Button } from '@/components/ui/button'
import { useBlogStyles } from '@/hooks/use-blog-styles'
import { usePageTitle } from '@/hooks/use-page-title'
import { adminApi, isAbortError, type PostInput, type PostStatus } from '@/lib/admin-api'

const fieldClass =
  'w-full rounded-lg border border-border bg-background px-3 py-2 text-sm outline-none transition-colors focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50'

function slugify(title: string): string {
  return title
    .toLowerCase()
    .normalize('NFKD')
    .replace(/[̀-ͯ]/g, '')
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
    .slice(0, 100)
}

function parseTags(raw: string): string[] {
  return raw.split(',').map((t) => t.trim()).filter(Boolean)
}

const emptyForm = { title: '', slug: '', description: '', tags: '', status: 'draft' as PostStatus, body: '' }
type Form = typeof emptyForm

// Keyed by post id so navigating /admin/new → /admin/:id, or between two posts,
// remounts the editor with fresh state instead of leaking the previous post's
// form (and any in-flight load) into the next one.
export default function AdminEditorPage() {
  const { id } = useParams<{ id: string }>()
  return <AdminEditor key={id ?? 'new'} id={id} />
}

function AdminEditor({ id }: { id?: string }) {
  const isNew = !id
  usePageTitle(isNew ? 'Admin · New post' : 'Admin · Edit post')
  useBlogStyles()
  const navigate = useNavigate()

  const [form, setForm] = useState<Form>(emptyForm)
  const [slugTouched, setSlugTouched] = useState(false)
  const [loading, setLoading] = useState(!isNew)
  const [saving, setSaving] = useState(false)
  const [dirty, setDirty] = useState(false)
  // updated_at of the version we loaded or last saved; sent back on save so the
  // server can refuse to overwrite a newer version from another tab.
  const [loadedUpdatedAt, setLoadedUpdatedAt] = useState<string | null>(null)
  const [message, setMessage] = useState<{ kind: 'ok' | 'error'; text: string } | null>(null)
  const [previewHtml, setPreviewHtml] = useState('')
  const [previewError, setPreviewError] = useState<string | null>(null)
  const [readingMinutes, setReadingMinutes] = useState(1)

  // The router's blocker callback must read the *current* dirtiness, and must be
  // clearable synchronously right before a post-save navigation.
  const dirtyRef = useRef(false)
  const markDirty = (value: boolean) => {
    dirtyRef.current = value
    setDirty(value)
  }

  const update = <K extends keyof Form>(key: K, value: Form[K]) => {
    setForm((f) => ({ ...f, [key]: value }))
    markDirty(true)
    setMessage(null)
  }

  // Load an existing post.
  useEffect(() => {
    if (!id) return
    let cancelled = false
    adminApi
      .get(id)
      .then((post) => {
        if (cancelled) return
        setForm({
          title: post.title,
          slug: post.slug,
          description: post.description,
          tags: post.tags.join(', '),
          status: post.status,
          body: post.body_md,
        })
        setLoadedUpdatedAt(post.updated_at)
        setSlugTouched(true)
      })
      .catch((err) => !cancelled && setMessage({ kind: 'error', text: err.message }))
      .finally(() => !cancelled && setLoading(false))
    return () => {
      cancelled = true
    }
  }, [id])

  // Debounced server-side preview. Rendering happens in the Go API with the same
  // sanitizing pipeline as the public site, so the HTML injected below is the
  // sanitized output, never raw Markdown input.
  useEffect(() => {
    const controller = new AbortController()
    const timer = setTimeout(() => {
      adminApi
        .preview(form.body, controller.signal)
        .then((res) => {
          setPreviewHtml(res.html)
          setReadingMinutes(res.reading_minutes)
          setPreviewError(null)
        })
        .catch((err) => {
          if (!isAbortError(err)) setPreviewError(err.message)
        })
    }, 500)
    return () => {
      clearTimeout(timer)
      controller.abort()
    }
  }, [form.body])

  // Unsaved work survives neither a tab close/reload (beforeunload) nor in-app
  // navigation: the "← All posts" link, the terminal bar's `home`/`about`, and
  // the browser Back button all go through the router, where the blocker asks first.
  useEffect(() => {
    if (!dirty) return
    const handler = (e: BeforeUnloadEvent) => e.preventDefault()
    window.addEventListener('beforeunload', handler)
    return () => window.removeEventListener('beforeunload', handler)
  }, [dirty])

  const blocker = useBlocker(
    ({ currentLocation, nextLocation }) => dirtyRef.current && currentLocation.pathname !== nextLocation.pathname,
  )
  useEffect(() => {
    if (blocker.state !== 'blocked') return
    if (window.confirm('You have unsaved changes. Leave this page and discard them?')) {
      blocker.proceed()
    } else {
      blocker.reset()
    }
  }, [blocker])

  const save = useCallback(async () => {
    if (saving) return
    const input: PostInput = {
      slug: form.slug,
      title: form.title,
      description: form.description,
      body_md: form.body,
      tags: parseTags(form.tags),
      status: form.status,
    }
    setSaving(true)
    setMessage(null)
    try {
      if (isNew) {
        const created = await adminApi.create(input)
        dirtyRef.current = false // clear synchronously so the blocker lets us through
        setDirty(false)
        navigate(`/admin/${created.id}`, { replace: true })
      } else {
        if (!loadedUpdatedAt) throw new Error('Post has not finished loading yet')
        const updated = await adminApi.update(id!, { ...input, updated_at: loadedUpdatedAt })
        setForm((f) => ({ ...f, slug: updated.slug, tags: updated.tags.join(', '), status: updated.status }))
        setLoadedUpdatedAt(updated.updated_at)
        markDirty(false)
        setMessage({ kind: 'ok', text: 'Saved. Public pages can take up to ~2 minutes to reflect changes, including unpublishing.' })
      }
    } catch (err) {
      setMessage({ kind: 'error', text: err instanceof Error ? err.message : 'Save failed' })
    } finally {
      setSaving(false)
    }
  }, [form, id, isNew, loadedUpdatedAt, navigate, saving])

  // Cmd/Ctrl+S saves. A ref keeps the listener stable while `save` changes.
  const saveRef = useRef(save)
  saveRef.current = save
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 's') {
        e.preventDefault()
        void saveRef.current()
      }
    }
    window.addEventListener('keydown', handler)
    return () => window.removeEventListener('keydown', handler)
  }, [])

  const remove = async () => {
    if (!id || !window.confirm(`Delete "${form.title}"? This can't be undone.`)) return
    try {
      await adminApi.remove(id)
      dirtyRef.current = false
      setDirty(false)
      navigate('/admin', { replace: true })
    } catch (err) {
      setMessage({ kind: 'error', text: err instanceof Error ? err.message : 'Delete failed' })
    }
  }

  if (loading) {
    return <div className="max-w-6xl mx-auto px-6 py-16 text-muted-foreground">Loading...</div>
  }

  return (
    <div className="max-w-6xl mx-auto px-6 py-10">
      <div className="flex items-center justify-between gap-4 mb-6">
        <Button variant="ghost" size="sm" nativeButton={false} render={<Link to="/admin" />}>
          ← All posts
        </Button>
        <div className="flex items-center gap-3">
          {message && (
            <span className={`text-sm ${message.kind === 'error' ? 'text-destructive' : 'text-muted-foreground'}`} role="status">
              {message.text}
            </span>
          )}
          {!message && dirty && <span className="text-sm text-muted-foreground">Unsaved changes</span>}
          {!isNew && form.status === 'published' && (
            <a href={`/blog/${form.slug}`} className="text-sm text-muted-foreground hover:text-foreground">
              View
            </a>
          )}
          {!isNew && (
            <Button variant="destructive" size="sm" className="cursor-pointer" onClick={remove}>
              Delete
            </Button>
          )}
          <Button className="cursor-pointer" onClick={() => void save()} disabled={saving}>
            {saving ? 'Saving...' : 'Save'}
          </Button>
        </div>
      </div>

      <div className="grid gap-3 sm:grid-cols-2 mb-3">
        <input
          className={`${fieldClass} sm:col-span-2 text-base font-medium`}
          placeholder="Title"
          value={form.title}
          maxLength={200}
          onChange={(e) => {
            update('title', e.target.value)
            if (isNew && !slugTouched) setForm((f) => ({ ...f, title: e.target.value, slug: slugify(e.target.value) }))
          }}
        />
        <input
          className={`${fieldClass} font-mono`}
          placeholder="slug-for-the-url"
          value={form.slug}
          maxLength={100}
          onChange={(e) => {
            setSlugTouched(true)
            update('slug', e.target.value)
          }}
          aria-label="Slug"
        />
        <select
          className={fieldClass}
          value={form.status}
          onChange={(e) => update('status', e.target.value as PostStatus)}
          aria-label="Status"
        >
          <option value="draft">Draft</option>
          <option value="published">Published</option>
        </select>
        <input
          className={`${fieldClass} sm:col-span-2`}
          placeholder="Short description (used in the post list, search results and link previews)"
          value={form.description}
          maxLength={300}
          onChange={(e) => update('description', e.target.value)}
          aria-label="Description"
        />
        <input
          className={`${fieldClass} sm:col-span-2`}
          placeholder="Tags, comma separated (lowercase letters, numbers, hyphens)"
          value={form.tags}
          onChange={(e) => update('tags', e.target.value)}
          aria-label="Tags"
        />
      </div>

      <div className="grid gap-4 lg:grid-cols-2">
        <textarea
          className={`${fieldClass} font-mono min-h-[28rem] lg:min-h-[36rem] resize-y leading-relaxed`}
          placeholder="Write in Markdown..."
          value={form.body}
          onChange={(e) => update('body', e.target.value)}
          spellCheck
          aria-label="Body (Markdown)"
        />
        <div className="rounded-lg border border-border p-5 min-h-[28rem] lg:min-h-[36rem] overflow-auto">
          <p className="text-xs uppercase tracking-wider text-muted-foreground mb-4">
            Preview · {readingMinutes} min read
          </p>
          {previewError ? (
            <p className="text-destructive text-sm">{previewError}</p>
          ) : (
            <div className="blog-scope">
              <div className="blog-prose" dangerouslySetInnerHTML={{ __html: previewHtml }} />
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
