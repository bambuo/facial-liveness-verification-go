<template>
  <div class="dev-portal">
    <AuthGate v-if="!auth.isDeveloper" />

    <div v-else class="portal-grid">
      <aside class="side">
        <div class="side-brand"><span class="brand-mark"><i /><i /><i /><i /></span><span>developer<em> / portal</em></span></div>
        <div class="side-group">
          <div class="side-label">Developers</div>
          <a-menu :selected-keys="[activeSection]" @menu-item-click="scrollToSection">
            <a-menu-item key="api-keys"><template #icon><IconLock /></template>API keys</a-menu-item>
            <a-menu-item key="webhooks"><template #icon><IconThunderbolt /></template>Webhooks</a-menu-item>
            <a-menu-item key="build"><template #icon><icon-storage /></template>Build</a-menu-item>
          </a-menu>
        </div>
        <div class="side-group">
          <div class="side-label">Account</div>
          <a-menu @menu-item-click="onAccountMenu">
            <a-menu-item key="settings"><template #icon><icon-settings /></template>Settings</a-menu-item>
            <a-menu-item key="signout"><template #icon><icon-export /></template>Sign out</a-menu-item>
          </a-menu>
        </div>
        <div class="side-foot">
          <div class="row"><span>Status</span><span class="ok">● Operational</span></div>
          <div class="row"><span>Region</span><span>iad-1</span></div>
          <div class="row"><span>Plan</span><span>starter</span></div>
        </div>
      </aside>

      <main class="main">
        <header class="topbar">
          <div class="crumbs"><span>Console</span><span>/</span><span>Developers</span><span>/</span><span class="here">{{ activeSection === 'webhooks' ? 'Webhooks' : 'API keys' }}</span></div>
          <div class="spacer" />
          <a-button size="mini" @click="load">↻ Refresh</a-button>
          <a-button size="mini" @click="$router.push('/docs')">Docs</a-button>
        </header>

        <div class="portal-page">
          <a-page-header :show-back="false" title="API keys" subtitle="Create credentials, monitor usage and keep integrations healthy.">
            <template #extra><a-button type="primary" @click="createKey">Create key</a-button></template>
          </a-page-header>

          <div ref="apiKeysAnchor" data-section="api-keys">
            <a-row :gutter="[12, 12]">
              <a-col :xs="12" :md="6"><a-card size="small"><a-statistic title="Requests this month" :value="stats.total_requests ?? 0" /></a-card></a-col>
              <a-col :xs="12" :md="6"><a-card size="small"><a-statistic title="Verified" :value="stats.successful_requests ?? 0" /></a-card></a-col>
              <a-col :xs="12" :md="6"><a-card size="small"><a-statistic title="Manual review" :value="stats.manual_review_requests ?? 0" /></a-card></a-col>
              <a-col :xs="12" :md="6"><a-card size="small"><a-typography-text type="secondary">Success rate</a-typography-text><a-typography-title :heading="3" style="margin:6px 0 0">{{ stats.success_rate ?? '—' }}</a-typography-title></a-card></a-col>
            </a-row>
            <a-card title="API keys" style="margin-top:12px">
              <a-table :data="keys" :columns="columns" :pagination="false" row-key="id">
                <template #status="{ record }"><a-tag :color="record.status === 'active' ? 'green' : 'gray'">{{ record.status }}</a-tag></template>
                <template #actions="{ record }"><a-button size="mini" status="danger" @click="revoke(record)">Revoke</a-button></template>
              </a-table>
            </a-card>
          </div>

          <div ref="webhooksAnchor" data-section="webhooks" style="margin-top:22px">
            <a-card title="Webhooks">
              <template #extra><a-button size="small" @click="showWebhook = true">Add endpoint</a-button></template>
              <a-table :data="webhooks" :columns="webhookColumns" :pagination="false" row-key="id">
                <template #active="{ record }"><a-tag :color="record.is_active ? 'green' : 'gray'">{{ record.is_active ? 'active' : 'disabled' }}</a-tag></template>
              </a-table>
              <a-empty v-if="webhooks.length === 0" description="No webhook endpoints" />
            </a-card>
          </div>

          <div ref="buildAnchor" data-section="build" style="margin-top:22px">
            <a-card>
              <a-space direction="vertical" size="small">
                <a-tag color="arcoblue">Page builder</a-tag>
                <a-typography-title :heading="4" style="margin:0">Host a verification page for your customers</a-typography-title>
                <a-typography-text type="secondary">Configure branding and a public slug, then share the link.</a-typography-text>
                <a-button type="primary" @click="$router.push('/developer/page-builder')">Open page builder</a-button>
              </a-space>
            </a-card>
          </div>

          <a-card title="Integration snippet" style="margin-top:22px">
            <pre class="code-block">curl -X POST {{ origin }}/api/v2/verify/initialize \
  -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"user_id":"00000000-0000-0000-0000-000000000000"}'</pre>
          </a-card>
        </div>
      </main>
    </div>

    <a-modal v-model:visible="showWebhook" title="Add webhook endpoint" @ok="addWebhook">
      <a-form :model="{}" layout="vertical">
        <a-form-item label="URL"><a-input v-model="webhookURL" placeholder="https://example.com/hooks/verification" /></a-form-item>
        <a-form-item label="Secret (optional)"><a-input-password v-model="webhookSecret" placeholder="At least 10 characters" /></a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { Message } from '@arco-design/web-vue'
import { api, apiMessage } from '../lib/api'
import { useAuthStore } from '../stores/auth'
import AuthGate from '../components/AuthGate.vue'

const auth = useAuthStore()
const apiKeyInput = ref(auth.apiKey)
const email = ref(auth.email)
const otpCode = ref('')
const loading = ref(false)
const otpLoading = ref(false)
const stats = reactive<Record<string, any>>({})
const keys = ref<any[]>([])
const webhooks = ref<any[]>([])
const activeSection = ref('api-keys')
const apiKeysAnchor = ref<HTMLElement | null>(null)
const webhooksAnchor = ref<HTMLElement | null>(null)
const buildAnchor = ref<HTMLElement | null>(null)
const showWebhook = ref(false)
const webhookURL = ref('')
const webhookSecret = ref('')
const origin = window.location.origin
const columns = [
  { title: 'Name', dataIndex: 'name' }, { title: 'Prefix', dataIndex: 'key_preview' },
  { title: 'Status', slotName: 'status', width: 110 }, { title: 'Last used', dataIndex: 'last_used_at' },
  { title: '', slotName: 'actions', width: 90 },
]
const webhookColumns = [
  { title: 'URL', dataIndex: 'url', ellipsis: true }, { title: 'Sandbox', dataIndex: 'is_sandbox', width: 90 },
  { title: 'Status', slotName: 'active', width: 110 }, { title: 'Created', dataIndex: 'created_at' },
]
async function load() {
  if (!auth.isDeveloper) return
  loading.value = true
  try {
    const [statsRes, keysRes, hooksRes] = await Promise.all([api.get('/developer/stats'), api.get('/developer/api-keys'), api.get('/developer/webhooks')])
    Object.assign(stats, statsRes.data)
    keys.value = keysRes.data.api_keys || []
    webhooks.value = hooksRes.data.webhooks || []
  } catch (error) { Message.error(apiMessage(error)) } finally { loading.value = false }
}
function connect() { if (!apiKeyInput.value.trim()) { Message.warning('Enter an API key'); return }; auth.setAPIKey(apiKeyInput.value.trim()); load() }
async function sendOTP() {
  if (!email.value) { Message.warning('Enter your email'); return }
  otpLoading.value = true
  try { const { data } = await api.post('/auth/developer/otp/send', { email: email.value }); if (data.code) { otpCode.value = data.code; Message.info(`Development code: ${data.code}`) } else Message.success('Code sent') }
  catch (error) { Message.error(apiMessage(error)) } finally { otpLoading.value = false }
}
async function verifyOTP() {
  if (!otpCode.value) { Message.warning('Enter the code'); return }
  try { const { data } = await api.post('/auth/developer/otp/verify', { email: email.value, code: otpCode.value }); if (data.developer) { auth.setDeveloperToken('', data.developer.email); Message.success('Email verified. Connect an API key to continue.') } }
  catch (error) { Message.error(apiMessage(error)) }
}
async function createKey() {
  try { const { data } = await api.post('/developer/api-key', { name: `Dashboard key ${new Date().toLocaleDateString()}`, is_sandbox: false }); Message.success('API key created. Store it securely.'); auth.setAPIKey(data.api_key); apiKeyInput.value = data.api_key; load() }
  catch (error) { Message.error(apiMessage(error)) }
}
async function revoke(record: any) {
  try { await api.delete(`/developer/api-key/${record.id}`); Message.success('API key revoked'); load() }
  catch (error) { Message.error(apiMessage(error)) }
}
async function addWebhook() {
  try { await api.post('/developer/webhooks', { url: webhookURL.value, is_sandbox: false, secret_token: webhookSecret.value }); Message.success('Webhook added'); showWebhook.value = false; load() }
  catch (error) { Message.error(apiMessage(error)) }
}
function scrollToSection(key: string) {
  const target = key === 'webhooks' ? webhooksAnchor.value : key === 'build' ? buildAnchor.value : apiKeysAnchor.value
  target?.scrollIntoView({ behavior: 'smooth', block: 'start' })
  activeSection.value = key
}
function onAccountMenu(key: string) {
  if (key === 'signout') { auth.logout(); return }
  if (key === 'settings') Message.info('Settings panel is available from the Page Builder and admin console.')
}
const sectionObserver = ref<IntersectionObserver | null>(null)
onMounted(() => {
  load()
  nextTick(() => {
    if (!window.IntersectionObserver) return
    sectionObserver.value = new IntersectionObserver((entries) => {
      const visible = entries.filter((entry) => entry.isIntersecting)
      if (!visible.length) return
      const winner = visible.reduce((a, b) => a.intersectionRatio > b.intersectionRatio ? a : b)
      const key = winner.target.getAttribute('data-section')
      if (key) activeSection.value = key
    }, { rootMargin: '-25% 0px -55% 0px', threshold: [0, .25, .5, .75, 1] })
    ;[apiKeysAnchor.value, webhooksAnchor.value, buildAnchor.value].forEach((target) => target && sectionObserver.value?.observe(target))
  })
})
onBeforeUnmount(() => sectionObserver.value?.disconnect())
</script>

<style scoped>
.dev-portal { min-height: 100vh; background: #f7f6f3; color: #0b0b0d; }
.mono { font-family: ui-monospace, Menlo, monospace; }
.auth-wrap { min-height: 100vh; display: grid; place-items: center; padding: 24px; background: #f7f6f3; }
.auth-card { width: min(420px, 100%); }
.auth-card h1 { margin: 8px 0; }
.small { font-size: 12px; margin-top: 16px; }
.portal-grid { display: grid; grid-template-columns: 220px minmax(0, 1fr); min-height: 100vh; max-width: 1320px; margin: 0 auto; }
.side { border-right: 1px solid #e4e2dc; position: sticky; top: 0; height: 100vh; display: flex; flex-direction: column; }
.side-brand { display: flex; align-items: center; gap: 10px; padding: 18px; border-bottom: 1px solid #e4e2dc; font-family: ui-monospace, monospace; font-size: 13px; }
.brand-mark { width: 18px; height: 18px; display: grid; grid-template-columns: 1fr 1fr; grid-template-rows: 1fr 1fr; }
.brand-mark i:nth-child(1), .brand-mark i:nth-child(4) { background: #0b0b0d; }
.brand-mark i:nth-child(2), .brand-mark i:nth-child(3) { background: #00d4d4; }
.side-brand em { font-style: normal; color: #6b6b70; font-weight: 400; }
.side-group { padding: 14px 8px 4px; }
.side-label { font-family: ui-monospace, monospace; font-size: 10px; letter-spacing: .08em; text-transform: uppercase; color: #9a9a9f; padding: 0 12px 6px; }
.side :deep(.arco-menu) { background: transparent; }
.side :deep(.arco-menu-item) { font-family: ui-monospace, monospace; font-size: 12.5px; color: #6b6b70; }
.side :deep(.arco-menu-item.arco-menu-selected) { color: #0b0b0d; background: #fff; border-left: 2px solid #00d4d4; }
.side-foot { margin-top: auto; padding: 14px 18px; border-top: 1px solid #e4e2dc; font-family: ui-monospace, monospace; font-size: 11px; color: #6b6b70; }
.side-foot .row { display: flex; justify-content: space-between; padding: 3px 0; }
.side-foot .ok { color: #00a6a6; }
.main { min-width: 0; }
.topbar { height: 48px; border-bottom: 1px solid #e4e2dc; display: flex; align-items: center; gap: 12px; padding: 0 24px; position: sticky; top: 0; background: #f7f6f3; z-index: 30; }
.crumbs { display: flex; gap: 6px; align-items: center; font-family: ui-monospace, monospace; font-size: 12px; color: #6b6b70; }
.crumbs .here { color: #0b0b0d; }
.spacer { flex: 1; }
.portal-page { padding: 28px 32px 80px; max-width: 1280px; margin: 0 auto; }
.code-block { margin: 0; }
@media (max-width: 760px) { .portal-grid { grid-template-columns: 1fr; } .side { position: static; height: auto; } .portal-page { padding: 20px 16px 60px; } }
</style>
