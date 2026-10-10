import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter, Routes, Route } from 'react-router-dom'
import './index.css'
import { Layout } from '@/components/layout'
import App from './App.tsx'
import About from './pages/About.tsx'
import AdminPosts from './pages/AdminPosts.tsx'
import AdminEditor from './pages/AdminEditor.tsx'
import NotFound from './pages/NotFound.tsx'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <BrowserRouter>
      <Routes>
        <Route element={<Layout />}>
          <Route path="/" element={<App />} />
          <Route path="/about" element={<About />} />
          {/* Gated at the edge by Cloudflare Access; the API enforces auth itself. */}
          <Route path="/admin" element={<AdminPosts />} />
          <Route path="/admin/new" element={<AdminEditor />} />
          <Route path="/admin/:id" element={<AdminEditor />} />
          <Route path="*" element={<NotFound />} />
        </Route>
      </Routes>
    </BrowserRouter>
  </StrictMode>,
)
