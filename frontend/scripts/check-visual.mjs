import { chromium } from 'playwright'
import fs from 'node:fs'
import path from 'node:path'
import { PNG } from 'pngjs'
import pixelmatch from 'pixelmatch'

const routes = [
  ['home', '/'], ['developer', '/developer'], ['page-builder', '/developer/page-builder'],
  ['demo', '/demo'], ['credential', '/verify-credential'], ['user-verification', '/user-verification'],
  ['slug-verification', '/v/demo-slug'], ['live-capture', '/live-capture'], ['mobile', '/verify/mobile'],
  ['docs', '/docs'], ['docs-guides', '/docs/guides'], ['docs-sdk', '/docs/sdk'],
  ['docs-features', '/docs/features'], ['docs-reference', '/docs/reference'],
  ['docs-markdown', '/docs/markdown'], ['docs-review', '/docs/review'],
  ['legal', '/legal'], ['setup', '/setup'], ['operator-login', '/operator/login'],
  ['admin-login', '/admin/login'], ['admin-verifications', '/admin/verifications'],
  ['admin-developers', '/admin/developers'], ['admin-thresholds', '/admin/thresholds'],
  ['not-found', '/does-not-exist'],
]
const update = process.env.UPDATE_BASELINES === '1'
const baselineDir = path.resolve('tests/baselines')
fs.mkdirSync(baselineDir, { recursive: true })
const browser = await chromium.launch({ headless: true })
const page = await browser.newPage({ viewport: { width: 1440, height: 1000 } })
await page.addInitScript(() => {
  localStorage.setItem('iv_admin_token', 'admin-test-token')
  localStorage.setItem('iv_email', 'admin@example.com')
  localStorage.setItem('iv_api_key', 'ik_test_key')
  crypto.randomUUID = () => '00000000-0000-4000-8000-000000000000'
})
await page.addStyleTag({ content: '*,*::before,*::after{animation:none!important;transition:none!important;caret-color:transparent!important}' }).catch(() => {})
const fulfill = (json) => (route) => route.fulfill({ json })
await page.route('**/api/admin/dashboard', fulfill({ stats: { verified: 42, failed: 3, manual_review: 7, pending: 5 }, developer_count: 12, recent_verifications: [], system_health: { status: 'ok' } }))
await page.route('**/api/admin/verifications**', fulfill({ verifications: [
  { verification_id: '11111111-1111-4111-8111-111111111111', status: 'COMPLETE', final_result: 'verified', current_step: 5, total_steps: 5, verification_mode: 'full', is_sandbox: true },
  { verification_id: '22222222-2222-4222-8222-222222222222', status: 'manual_review', final_result: 'manual_review', current_step: 5, total_steps: 5, verification_mode: 'full', is_sandbox: false },
], pagination: { page: 1, limit: 20, total: 2, pages: 1 } }))
await page.route('**/api/admin/verification/*/review', fulfill({ verification: {}, message: 'Review recorded' }))
await page.route('**/api/admin/verification/*', fulfill({ verification: { verification_id: '11111111-1111-4111-8111-111111111111', status: 'manual_review', final_result: null, current_step: 5, total_steps: 5, verification_mode: 'full', is_sandbox: true } }))
await page.route('**/api/admin/developers**', fulfill({ developers: [{ id: '33333333-3333-4333-8333-333333333333', name: 'Acme', email: 'dev@acme.test', company: 'Acme', status: 'active', created_at: '2026-11-01T00:00:00Z' }], pagination: { page: 1, limit: 100, total: 1, pages: 1 } }))
await page.route('**/api/admin/thresholds**', (route) => route.request().method() === 'GET'
  ? route.fulfill({ json: { success: true, data: { production: { face_matching: 0.85, liveness: 0.75, cross_validation: 0.7 }, sandbox: { face_matching: 0.8, liveness: 0.65, cross_validation: 0.7 } } } })
  : route.fulfill({ json: { success: true, data: { explanation: 'preview' } } }))
await page.route('**/api/developer/stats', fulfill({ total_requests: 128, successful_requests: 110, manual_review_requests: 8, failed_requests: 10, success_rate: '85.94%' }))
await page.route('**/api/developer/api-keys', fulfill({ api_keys: [{ id: '44444444-4444-4444-8444-444444444444', name: 'Server', key_preview: 'ik_abc...', status: 'active', created_at: '2026-11-01T00:00:00Z' }], total_keys: 1, active_keys: 1 }))
await page.route('**/api/developer/settings/page-builder', fulfill({ success: true, data: { config: { company_name: 'Acme', accent_color: '#2563eb', welcome_text: 'Verify with Acme' }, slug: 'acme' } }))
await page.route('**/api/setup/status', fulfill({ initialized: false }))

let failed = 0
for (const [name, route] of routes) {
  const file = path.join(baselineDir, `${name}.png`)
  await page.goto(`http://127.0.0.1:5173${route}`, { waitUntil: 'networkidle' })
  await page.waitForTimeout(300)
  await page.screenshot({ path: `${file}.tmp.png`, fullPage: false })
  if (update || !fs.existsSync(file)) {
    fs.copyFileSync(`${file}.tmp.png`, file)
    console.log('BASELINE', name)
  } else {
    const actual = PNG.sync.read(fs.readFileSync(`${file}.tmp.png`))
    const expected = PNG.sync.read(fs.readFileSync(file))
    if (actual.width !== expected.width || actual.height !== expected.height) {
      failed++; console.log('FAIL', name, 'size mismatch'); continue
    }
    const diff = new PNG({ width: actual.width, height: actual.height })
    const pixels = pixelmatch(actual.data, expected.data, diff.data, actual.width, actual.height, { threshold: 0.1 })
    const ratio = pixels / (actual.width * actual.height)
    if (ratio > 0.005) {
      failed++; fs.writeFileSync(path.join(baselineDir, `${name}.diff.png`), PNG.sync.write(diff))
      console.log('FAIL', name, `${(ratio * 100).toFixed(3)}% pixels differ`)
    } else console.log('OK  ', name, `${(ratio * 100).toFixed(3)}%`)
  }
  fs.rmSync(`${file}.tmp.png`, { force: true })
}
await browser.close()
if (failed) { console.log('Visual regression failures:', failed); process.exit(1) }
console.log('Visual baseline OK')
