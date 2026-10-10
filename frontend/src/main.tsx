import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { createBrowserRouter, RouterProvider } from 'react-router-dom'
import './index.css'
import { Layout } from '@/components/layout'
import App from './App.tsx'
import About from './pages/About.tsx'
import AdminPosts from './pages/AdminPosts.tsx'
import AdminEditor from './pages/AdminEditor.tsx'
import NotFound from './pages/NotFound.tsx'

// A data router (rather than <BrowserRouter>) is required for useBlocker, which
// the admin editor uses to stop in-app navigation from discarding unsaved work.
const router = createBrowserRouter([
  {
    element: <Layout />,
    children: [
      { path: '/', element: <App /> },
      { path: '/about', element: <About /> },
      // Gated at the edge by Cloudflare Access; the API enforces auth itself.
      { path: '/admin', element: <AdminPosts /> },
      { path: '/admin/new', element: <AdminEditor /> },
      { path: '/admin/:id', element: <AdminEditor /> },
      { path: '*', element: <NotFound /> },
    ],
  },
])

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <RouterProvider router={router} />
  </StrictMode>,
)
