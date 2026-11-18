<template>
  <div class="page">
    <div class="card-title">
      <div><div class="eyebrow">Developer portal</div><h1 style="margin:8px 0 4px">Build trust into your product</h1><p class="muted">Issue verifications, manage API keys and monitor outcomes from one place.</p></div>
      <a-button v-if="auth.isDeveloper" type="outline" @click="load">Refresh</a-button>
    </div>

    <a-alert v-if="!auth.isDeveloper" type="info" style="margin:20px 0">Enter an API key to unlock the dashboard. Keys are stored only in this browser.</a-alert>

    <div class="grid grid-2" style="margin-top:20px">
      <div class="panel">
        <div class="card-title"><h3 style="margin:0">{{ auth.isDeveloper ? 'Workspace connected' : 'Connect workspace' }}</h3><a-tag :color="auth.isDeveloper ? 'green' : 'gray'">{{ auth.isDeveloper ? 'Active' : 'Not connected' }}</a-tag></div>
        <a-form :model="{}" layout="vertical" style="margin-top:16px" @submit.prevent="connect">
          <a-form-item label="API key">
            <a-input-password v-model="apiKeyInput" placeholder="ik_..." allow-clear />
          </a-form-item>
          <a-space><a-button type="primary" html-type="submit" :loading="loading">Connect</a-button><a-button @click="openDemo">Try the demo</a-button></a-space>
        </a-form>
        <a-divider />
        <a-form :model="{}" layout="vertical">
          <a-form-item label="Email OTP (optional)">
            <a-input v-model="email" placeholder="you@company.com" />
          </a-form-item>
          <a-space><a-button @click="sendOTP" :loading="otpLoading">Send code</a-button><a-input v-model="otpCode" placeholder="6-digit code" style="width:140px" /><a-button @click="verifyOTP">Verify</a-button></a-space>
        </a-form>
      </div>

      <div class="panel">
        <div class="card-title"><h3 style="margin:0">Quick actions</h3></div>
        <div class="quick-grid">
          <button class="quick" @click="$router.push('/developer/page-builder')"><icon-layout /><span>Page builder</span><small>Hosted verification page</small></button>
          <button class="quick" @click="$router.push('/docs')"><icon-book /><span>API docs</span><small>Endpoints and guides</small></button>
          <button class="quick" @click="$router.push('/live-capture')"><icon-camera /><span>Live capture</span><small>Camera and liveness</small></button>
          <button class="quick" @click="$router.push('/user-verification')"><icon-user /><span>Verification flow</span><small>Document + selfie</small></button>
        </div>
      </div>
    </div>

    <div class="grid grid-4" style="margin-top:20px">
      <div class="panel"><div class="stat-label">Requests this month</div><div class="stat-value">{{ stats.total_requests ?? '—' }}</div></div>
      <div class="panel"><div class="stat-label">Verified</div><div class="stat-value">{{ stats.successful_requests ?? '—' }}</div></div>
      <div class="panel"><div class="stat-label">Manual review</div><div class="stat-value">{{ stats.manual_review_requests ?? '—' }}</div></div>
      <div class="panel"><div class="stat-label">Success rate</div><div class="stat-value">{{ stats.success_rate ?? '—' }}</div></div>
    </div>

    <div class="grid grid-2" style="margin-top:20px">
      <div class="panel">
        <div class="card-title"><h3 style="margin:0">API keys</h3><a-button size="small" @click="createKey">Create key</a-button></div>
        <a-table :data="keys" :columns="keyColumns" :pagination="false" style="margin-top:12px" />
      </div>
      <div class="panel">
        <div class="card-title"><h3 style="margin:0">Integration snippet</h3><a-tag>cURL</a-tag></div>
        <pre class="code-block">curl -X POST {{ origin }}/api/v2/verify/initialize \
  -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"user_id":"00000000-0000-0000-0000-000000000000"}'</pre>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { Message } from '@arco-design/web-vue'
import { useRouter } from 'vue-router'
import { api, apiMessage } from '../lib/api'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const router = useRouter()
const apiKeyInput = ref(auth.apiKey)
const email = ref(auth.email)
const otpCode = ref('')
const loading = ref(false)
const otpLoading = ref(false)
const stats = reactive<Record<string, any>>({})
const keys = ref<any[]>([])
const origin = window.location.origin
const keyColumns = [
  { title: 'Name', dataIndex: 'name' },
  { title: 'Prefix', dataIndex: 'key_preview' },
  { title: 'Status', dataIndex: 'status' },
  { title: 'Created', dataIndex: 'created_at' },
]

async function load() {
  if (!auth.isDeveloper) return
  loading.value = true
  try {
    const [statsRes, keyRes] = await Promise.all([api.get('/developer/stats'), api.get('/developer/api-keys')])
    Object.assign(stats, statsRes.data)
    keys.value = keyRes.data.api_keys || []
  } catch (error) { Message.error(apiMessage(error)) } finally { loading.value = false }
}
function connect() {
  if (!apiKeyInput.value.trim()) { Message.warning('Enter an API key'); return }
  auth.setAPIKey(apiKeyInput.value.trim())
  load()
}
function openDemo() { router.push('/demo') }
async function sendOTP() {
  if (!email.value) { Message.warning('Enter your email'); return }
  otpLoading.value = true
  try {
    const { data } = await api.post('/auth/developer/otp/send', { email: email.value })
    if (data.code) { otpCode.value = data.code; Message.info(`Development code: ${data.code}`) } else Message.success('Code sent')
  } catch (error) { Message.error(apiMessage(error)) } finally { otpLoading.value = false }
}
async function verifyOTP() {
  if (!otpCode.value) { Message.warning('Enter the code'); return }
  try {
    const { data } = await api.post('/auth/developer/otp/verify', { email: email.value, code: otpCode.value })
    if (data.developer) { auth.setDeveloperToken('', data.developer.email); Message.success('Email verified. Connect an API key to continue.') }
  } catch (error) { Message.error(apiMessage(error)) }
}
async function createKey() {
  try {
    const { data } = await api.post('/developer/api-key', { name: `Dashboard key ${new Date().toLocaleDateString()}`, is_sandbox: false })
    Message.success('API key created. Store it securely.')
    apiKeyInput.value = data.api_key
    auth.setAPIKey(data.api_key)
    load()
  } catch (error) { Message.error(apiMessage(error)) }
}
onMounted(load)
</script>

<style scoped>
.quick-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 12px; margin-top: 14px; }
.quick { display: flex; flex-direction: column; align-items: flex-start; gap: 4px; padding: 14px; border: 1px solid #e2e8f0; border-radius: 10px; background: #fff; text-align: left; cursor: pointer; font-size: 16px; }
.quick:hover { border-color: #93c5fd; background: #f8fbff; }
.quick span { font-weight: 600; color: #0f172a; }
.quick small { color: #64748b; font-size: 12px; }
</style>
