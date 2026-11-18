import { chromium } from 'playwright'

const routes = [
  '/', '/developer', '/developer/page-builder', '/demo', '/verify', '/verify-credential',
  '/user-verification', '/v/demo-slug', '/live-capture', '/verify/mobile',
  '/docs', '/docs/guides', '/docs/sdk', '/docs/features', '/docs/reference',
  '/docs/markdown', '/docs/review', '/legal', '/setup', '/operator/login',
  '/admin/login', '/admin/verifications', '/admin/developers', '/admin/thresholds',
  '/admin/anything', '/does-not-exist',
]
const browser = await chromium.launch({ headless: true })
const page = await browser.newPage({ viewport: { width: 1440, height: 1000 } })
let failed = 0
for (const route of routes) {
  const errors = []
  const onConsole = (msg) => { if (msg.type() === 'error') errors.push(msg.text()) }
  const onPageError = (err) => errors.push(err.message)
  page.on('console', onConsole); page.on('pageerror', onPageError)
  const response = await page.goto(`http://127.0.0.1:5173${route}`, { waitUntil: 'networkidle' })
  const text = (await page.locator('body').innerText()).trim()
  const ok = response?.ok() && text.length > 60 && errors.length === 0
  if (!ok) { failed++; console.log('FAIL', route, 'status=', response?.status(), 'text=', text.length, 'url=', page.url(), 'errors=', errors.slice(0, 3)) }
  else console.log('OK  ', route, text.length)
  page.off('console', onConsole); page.off('pageerror', onPageError)
}
await browser.close()
if (failed) { console.log(`FAILED ${failed}/${routes.length}`); process.exit(1) }
console.log(`Routes OK: ${routes.length}`)
