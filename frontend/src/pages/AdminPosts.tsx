import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { usePageTitle } from '@/hooks/use-page-title'
import { adminApi, type PostSummary } from '@/lib/admin-api'

function formatDate(iso: string | null): string {
  if (!iso) return '—'
  return new Date(iso).toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' })
}

export default function AdminPostsPage() {
  usePageTitle('Admin · Posts')
  const [posts, setPosts] = useState<PostSummary[] | null>(null)
  const [email, setEmail] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    adminApi.list().then(setPosts).catch((err) => setError(err.message))
    adminApi.me().then((me) => setEmail(me.email)).catch(() => {})
  }, [])

  return (
    <div className="max-w-3xl mx-auto px-6 py-16">
      <div className="flex items-end justify-between gap-4 mb-8">
        <div>
          <h1 className="text-4xl font-bold tracking-tight">Posts</h1>
          {email && <p className="text-sm text-muted-foreground mt-1">Signed in as {email}</p>}
        </div>
        <Button className="cursor-pointer" nativeButton={false} render={<Link to="/admin/new" />}>
          New post
        </Button>
      </div>

      {error && <p className="text-destructive">{error}</p>}
      {!posts && !error && <p className="text-muted-foreground">Loading...</p>}
      {posts && posts.length === 0 && (
        <p className="text-muted-foreground">No posts yet. Write the first one.</p>
      )}

      {posts && posts.length > 0 && (
        <ul className="divide-y divide-border border-y border-border">
          {posts.map((post) => (
            <li key={post.id} className="flex items-center gap-4 py-4">
              <div className="min-w-0 flex-1">
                <Link to={`/admin/${post.id}`} className="font-medium hover:underline truncate block">
                  {post.title}
                </Link>
                <p className="text-xs text-muted-foreground mt-0.5">
                  /blog/{post.slug} · updated {formatDate(post.updated_at)}
                  {post.published_at && ` · published ${formatDate(post.published_at)}`}
                </p>
              </div>
              <Badge variant={post.status === 'published' ? 'default' : 'secondary'}>{post.status}</Badge>
              {post.status === 'published' && (
                // Plain anchor: /blog is rendered by the Go API, not the SPA router.
                <a href={`/blog/${post.slug}`} className="text-sm text-muted-foreground hover:text-foreground">
                  View
                </a>
              )}
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
