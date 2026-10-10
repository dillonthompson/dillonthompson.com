import { useEffect } from 'react'

const STYLESHEETS = ['/blog/assets/blog.css', '/blog/assets/syntax.css']

// Loads the Go-served blog stylesheets so the editor preview renders post
// bodies exactly as the public site does. The sheets are scoped to
// `.blog-scope`, so they can't restyle the rest of the app.
export function useBlogStyles() {
  useEffect(() => {
    const links = STYLESHEETS.map((href) => {
      const link = document.createElement('link')
      link.rel = 'stylesheet'
      link.href = href
      document.head.appendChild(link)
      return link
    })
    return () => links.forEach((l) => l.remove())
  }, [])
}
