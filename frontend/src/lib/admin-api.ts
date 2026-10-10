// Client for the authenticated blog admin API (/api/v1/admin/*). Access to
// these endpoints is enforced server-side (Cloudflare Access + JWT
// verification); nothing here is a security boundary.

export type PostStatus = 'draft' | 'published'

export interface PostSummary {
  id: string
  slug: string
  title: string
  description: string
  tags: string[]
  status: PostStatus
  published_at: string | null
  created_at: string
  updated_at: string
}

export interface Post extends PostSummary {
  body_md: string
}

export interface PostInput {
  /** updated_at of the version being edited; required when updating. */
  updated_at?: string
  slug: string
  title: string
  description: string
  body_md: string
  tags: string[]
  status: PostStatus
}

export class ApiError extends Error {
  status: number
  constructor(status: number, message: string, options?: ErrorOptions) {
    super(message, options)
    this.status = status
  }
}

export function isAbortError(err: unknown): boolean {
  return err instanceof DOMException && err.name === 'AbortError'
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  let res: Response
  try {
    res = await fetch(`/api/v1/admin${path}`, {
      credentials: 'same-origin',
      ...init,
      headers: init.body ? { 'Content-Type': 'application/json' } : undefined,
    })
  } catch (err) {
    // A caller-initiated abort (e.g. a newer preview request superseding this
    // one) is not a failure; let it through untouched.
    if (isAbortError(err)) throw err
    // An expired Cloudflare Access session redirects XHRs to a cross-origin
    // login page, which the browser reports as a network failure.
    throw new ApiError(0, 'Request failed. Your session may have expired — reload the page to sign in again.', { cause: err })
  }
  if (res.status === 204) return undefined as T
  const data = await res.json().catch(() => null)
  if (!res.ok) throw new ApiError(res.status, data?.error ?? `HTTP ${res.status}`)
  return data as T
}

export const adminApi = {
  me: () => request<{ email: string }>('/me'),
  list: () => request<PostSummary[]>('/posts'),
  get: (id: string) => request<Post>(`/posts/${id}`),
  create: (input: PostInput) => request<Post>('/posts', { method: 'POST', body: JSON.stringify(input) }),
  update: (id: string, input: PostInput) => request<Post>(`/posts/${id}`, { method: 'PUT', body: JSON.stringify(input) }),
  remove: (id: string) => request<void>(`/posts/${id}`, { method: 'DELETE' }),
  preview: (bodyMd: string, signal?: AbortSignal) =>
    request<{ html: string; reading_minutes: number }>('/preview', {
      method: 'POST',
      body: JSON.stringify({ body_md: bodyMd }),
      signal,
    }),
}
