import { chromium } from 'playwright'

const browser = await chromium.launch({ headless: true })
const page = await browser.newPage({ viewport: { width: 1440, height: 1000 } })
await page.addInitScript(() => {
  localStorage.setItem('iv_admin_token', 'admin-test-token')
  localStorage.setItem('iv_email', 'admin@example.com')
  localStorage.setItem('iv_api_key', 'ik_test_key')
})
await page.route('**/api/admin/dashboard', (route) => route.fulfill({ json: { stats: { verified: 42, failed: 3, manual_review: 7, pending: 5 }, developer_count: 12, recent_verifications: [], system_health: { status: 'ok' } } }))
await page.route('**/api/admin/verifications**', (route) => route.fulfill({ json: { verifications: [
  { verification_id: '11111111-1111-4111-8111-111111111111', status: 'COMPLETE', final_result: 'verified', current_step: 5, total_steps: 5, verification_mode: 'full', is_sandbox: true },
  { verification_id: '22222222-2222-4222-8222-222222222222', status: 'manual_review', final_result: 'manual_review', current_step: 5, total_steps: 5, verification_mode: 'full', is_sandbox: false },
], pagination: { page: 1, limit: 20, total: 2, pages: 1 } } }))
await page.route('**/api/admin/verification/*/review', (route) => route.fulfill({ json: { verification: {}, message: 'Review recorded' } }))
await page.route('**/api/admin/verification/*', (route) => route.fulfill({ json: { verification: { verification_id: '11111111-1111-4111-8111-111111111111', status: 'manual_review', final_result: null, current_step: 5, total_steps: 5, verification_mode: 'full', is_sandbox: true } } }))
await page.route('**/api/admin/developers**', (route) => route.fulfill({ json: { developers: [
  { id: '33333333-3333-4333-8333-333333333333', name: 'Acme', email: 'dev@acme.test', company: 'Acme', status: 'active', created_at: '2026-11-01T00:00:00Z' },
], pagination: { page: 1, limit: 100, total: 1, pages: 1 } } }))
await page.route('**/api/admin/thresholds**', (route) => {
  if (route.request().method() === 'GET') return route.fulfill({ json: { success: true, data: { production: { face_matching: 0.85, liveness: 0.75, cross_validation: 0.7 }, sandbox: { face_matching: 0.8, liveness: 0.65, cross_validation: 0.7 } } } })
  return route.fulfill({ json: { success: true, data: { explanation: 'preview' } } })
})
await page.route('**/api/developer/stats', (route) => route.fulfill({ json: { total_requests: 128, successful_requests: 110, manual_review_requests: 8, failed_requests: 10, success_rate: '85.94%' } }))
await page.route('**/api/developer/api-keys', (route) => route.fulfill({ json: { api_keys: [{ id: '44444444-4444-4444-8444-444444444444', name: 'Server', key_preview: 'ik_abc...', status: 'active', created_at: '2026-11-01T00:00:00Z' }], total_keys: 1, active_keys: 1 } }))
await page.route('**/api/developer/settings/page-builder', (route) => route.fulfill({ json: { success: true, data: { config: { company_name: 'Acme', accent_color: '#2563eb', welcome_text: 'Verify with Acme' }, slug: 'acme' } } }))

const checks = []
async function visit(path, expect) {
  const errors = []
  page.on('pageerror', (err) => errors.push(err.message))
  await page.goto(`http://127.0.0.1:5173${path}`, { waitUntil: 'networkidle' })
  const body = await page.locator('body').innerText()
  const ok = body.includes(expect) && errors.length === 0
  checks.push({ path, ok, errors })
  console.log(ok ? 'OK  ' : 'FAIL', path, body.length, errors.slice(0, 2))
}
await visit('/admin/verifications', 'Verification management')
await page.screenshot({ path: '/tmp/iv-admin-verifications.png', fullPage: false })
await visit('/admin/developers', 'Developers')
await page.screenshot({ path: '/tmp/iv-admin-developers.png', fullPage: false })
await visit('/admin/thresholds', 'Verification thresholds')
await page.screenshot({ path: '/tmp/iv-admin-thresholds.png', fullPage: false })
await visit('/developer', 'API keys')
await page.screenshot({ path: '/tmp/iv-developer-authed.png', fullPage: false })
await visit('/developer/page-builder', 'Page builder')
await page.screenshot({ path: '/tmp/iv-page-builder.png', fullPage: false })
await browser.close()
const failed = checks.filter((item) => !item.ok)
if (failed.length) { console.log('FAILED', JSON.stringify(failed)); process.exit(1) }
console.log('Authenticated UI OK:', checks.length)
