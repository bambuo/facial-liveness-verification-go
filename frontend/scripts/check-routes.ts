import { chromium } from 'playwright'

const base = process.env.FRONTEND_URL || 'http://127.0.0.1:5173'
const routes = [
  '/', '/developer', '/developer/page-builder', '/demo', '/verify', '/verify-credential',
  '/user-verification', '/v/demo-slug', '/live-capture', '/verify/mobile',
  '/docs', '/docs/guides', '/docs/sdk', '/docs/features', '/docs/reference',
  '/docs/markdown', '/docs/review', '/legal', '/setup',
  '/admin/login', '/admin/verifications', '/admin/developers', '/admin/anything', '/does-not-exist',
]
const expectedAuthProbe = /status of (401|403)/
const browser = await chromium.launch({ headless: true })
const page = await browser.newPage({ viewport: { width: 1440, height: 1000 } })
let failed = 0
for (const route of routes) {
  const errors = []
  const onConsole = (message) => {
    if (message.type() !== 'error') return
    if (expectedAuthProbe.test(message.text())) return
    errors.push(message.text())
  }
  const onPageError = (error) => errors.push(error.message)
  page.on('console', onConsole)
  page.on('pageerror', onPageError)
  const response = await page.goto(`${base}${route}`, { waitUntil: 'networkidle' })
  const text = (await page.locator('body').innerText()).trim()
  const ok = response?.ok() && text.length > 50 && errors.length === 0
  if (!ok) {
    failed++
    console.log('FAIL', route, 'status=', response?.status(), 'text=', text.length, errors.slice(0, 2))
  } else {
    console.log('OK  ', route, text.length)
  }
  page.off('console', onConsole)
  page.off('pageerror', onPageError)
}
await browser.close()
if (failed) {
  console.log(`FAILED ${failed}/${routes.length}`)
  process.exit(1)
}
console.log(`Routes OK: ${routes.length}`)
