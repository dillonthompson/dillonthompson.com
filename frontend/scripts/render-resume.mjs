// Render frontend/scripts/resume.html to a PDF via headless Chromium.
//
// Usage (from the frontend/ directory):
//   pnpm resume:pdf
//
// Output lands at frontend/scripts/Dillon_Thompson_Resume.pdf. Copy or move
// wherever you want. Rerun any time you edit resume.html.

import puppeteer from 'puppeteer'
import { fileURLToPath } from 'node:url'
import { existsSync } from 'node:fs'
import path from 'node:path'

const scriptsDir = path.dirname(fileURLToPath(import.meta.url))
const htmlPath = path.join(scriptsDir, 'resume.html')
const outPath = path.join(scriptsDir, 'Dillon_Thompson_Resume.pdf')

if (!existsSync(htmlPath)) {
  console.error(`✗ Could not find ${htmlPath}`)
  process.exit(1)
}

const start = Date.now()
console.log(`→ Rendering ${path.basename(htmlPath)}...`)

const browser = await puppeteer.launch({
  // 'new' headless is default on recent Chromium; kept explicit for clarity.
  headless: true,
})

try {
  const page = await browser.newPage()

  // networkidle0 waits until there are no in-flight requests. Overkill for a
  // static HTML file with no external assets, but ensures any future CSS
  // includes get pulled in before we snapshot.
  await page.goto(`file://${htmlPath}`, { waitUntil: 'networkidle0' })

  // Force print-media styling so any `@media print` rules take effect. The
  // resume CSS doesn't have any today but future-us might add them.
  await page.emulateMediaType('print')

  await page.pdf({
    path: outPath,
    format: 'letter',
    printBackground: true,
    preferCSSPageSize: true,
  })
} finally {
  await browser.close()
}

const elapsed = ((Date.now() - start) / 1000).toFixed(1)
console.log(`✓ Wrote ${outPath} in ${elapsed}s`)
