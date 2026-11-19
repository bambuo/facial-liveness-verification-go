<template>
  <!-- Loading state -->
  <div v-if="state === 'loading'" class="setup-screen">
    <span class="status-line">Checking setup status...</span>
  </div>

  <!-- Error state — API unreachable -->
  <div v-else-if="state === 'error'" class="setup-screen">
    <a-result status="error" title="Could not reach the API server" subtitle="Make sure the backend is running and try again.">
      <template #extra>
        <a-button @click="retry">Retry</a-button>
      </template>
    </a-result>
  </div>

  <!-- Success state — show API key + next steps -->
  <div v-else-if="state === 'success' && result" class="setup-success">
    <a-result status="success" title="You're all set" :subtitle="`Welcome, ${result.developer?.name ?? ''}. Your developer account is ready.`" />

    <a-card v-if="result.api_key" class="api-key-card">
      <div class="api-key-label">Your API Key — save it now, it won't be shown again</div>
      <div class="api-key-row">
        <code class="api-key-value">{{ result.api_key.key }}</code>
        <a-button type="primary" class="api-key-copy" @click="copyKey(result.api_key.key)">
          <template #icon><icon-copy /></template>
          {{ copied ? 'Copied' : 'Copy' }}
        </a-button>
      </div>
      <div class="api-key-type">Type: {{ result.api_key.is_sandbox ? 'Sandbox' : 'Production' }} • Name: {{ result.api_key.name }}</div>
    </a-card>

    <a-card class="next-steps">
      <div class="next-steps-title">Next steps</div>
      <div class="next-step">1. Copy the API key above and store it securely</div>
      <div class="next-step">
        2. Check the <a-link @click="router.push('/docs')">documentation</a-link> for integration guides
      </div>
      <div class="next-step">
        3. Try the <a-link @click="router.push('/demo')">interactive demo</a-link> to test verification
      </div>
    </a-card>

    <a-button long class="dashboard-button" @click="router.replace('/')">Go to Dashboard</a-button>
  </div>

  <!-- Form state — welcome + create account -->
  <div v-else class="setup-screen setup-screen-top">
    <div class="setup-form">
      <div class="setup-head">
        <div class="logo"><span class="brand-mark"><i /><i /><i /><i /></span><span>identity</span></div>
        <a-typography-title :heading="3" class="setup-title" style="margin:0 0 8px">Welcome to Identity Verification</a-typography-title>
        <a-typography-text type="secondary" class="setup-desc">Create your admin developer account to get started. This will generate your first API key.</a-typography-text>
      </div>

      <a-form :model="formModel" layout="vertical" @submit="submitSetup">
        <a-card class="setup-card">
          <a-form-item label="Name *">
            <a-input
              v-model="name"
              :input-attrs="{ autofocus: true, autocomplete: 'name' }"
              placeholder="Your name"
              :disabled="loading"
            />
          </a-form-item>
          <a-form-item label="Email *">
            <a-input
              v-model="email"
              type="text"
              :input-attrs="{ inputmode: 'email', autocomplete: 'email' }"
              placeholder="you@example.com"
              :disabled="loading"
            />
          </a-form-item>
          <a-form-item>
            <template #label>Company <span class="optional">(optional)</span></template>
            <a-input v-model="company" placeholder="Your company" :disabled="loading" />
          </a-form-item>
          <a-button type="primary" long html-type="submit" :loading="loading" :disabled="loading || !name.trim() || !email.trim()">
            {{ loading ? 'Creating account...' : 'Create Account & Generate API Key' }}
          </a-button>
        </a-card>
      </a-form>

      <div class="setup-note">This is a one-time setup for your self-hosted instance.</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { Message } from '@arco-design/web-vue'
import { useRouter } from 'vue-router'
import { api, apiMessage } from '../lib/api'
import { useAuthStore } from '../stores/auth'

type SetupState = 'loading' | 'form' | 'success' | 'error'

interface SetupResult {
  token?: string
  developer: { id: string; email: string; name: string; company: string | null }
  api_key?: { key: string; id: string; name: string; is_sandbox: boolean; created_at: string }
}

const auth = useAuthStore()
const router = useRouter()

const state = ref<SetupState>('loading')
const name = ref('')
const email = ref('')
const company = ref('')
const loading = ref(false)
const result = ref<SetupResult | null>(null)
const copied = ref(false)
const formModel = {}

onMounted(checkStatus)

async function checkStatus() {
  try {
    const { data } = await api.get('/setup/status')
    // Contract field is needs_setup; the current core returns initialized
    const needsSetup = typeof data?.needs_setup === 'boolean' ? data.needs_setup : !data?.initialized
    if (needsSetup) {
      state.value = 'form'
    } else {
      router.replace('/')
    }
  } catch {
    state.value = 'error'
  }
}

function retry() {
  window.location.reload()
}

async function submitSetup() {
  if (!name.value.trim() || !email.value.trim()) return
  loading.value = true
  try {
    const { data } = await api.post(
      '/setup/initialize',
      {
        name: name.value.trim(),
        email: email.value.trim(),
        company: company.value.trim() || undefined,
        locale: 'en',
      },
      { headers: { 'Accept-Language': 'en' } },
    )
    if (data?.token) auth.setDeveloperToken(String(data.token), email.value.trim())
    result.value = data as SetupResult
    state.value = 'success'
    Message.success('Account created!')
  } catch (e) {
    Message.error(apiMessage(e))
  } finally {
    loading.value = false
  }
}

async function copyKey(key: string) {
  try {
    await navigator.clipboard.writeText(key)
    copied.value = true
    Message.success('API key copied')
    setTimeout(() => { copied.value = false }, 2000)
  } catch {
    Message.error('Failed to copy')
  }
}
</script>

<style scoped>
.setup-screen { min-height: 100vh; background: #f7f6f3; display: flex; align-items: center; justify-content: center; padding: 24px; }
.setup-screen-top { align-items: flex-start; padding-top: 64px; }
.status-line { font-family: ui-monospace, Menlo, monospace; font-size: 13px; letter-spacing: .04em; color: #6b6b70; }
.setup-form { width: 100%; max-width: 440px; }
.setup-head { text-align: center; margin-bottom: 40px; }
.logo { display: flex; align-items: center; justify-content: center; gap: 8px; margin-bottom: 20px; font-family: ui-monospace, Menlo, monospace; font-size: 14px; color: #0b0b0d; }
.brand-mark { width: 18px; height: 18px; display: grid; grid-template-columns: 1fr 1fr; grid-template-rows: 1fr 1fr; }
.brand-mark i:nth-child(1), .brand-mark i:nth-child(4) { background: #0b0b0d; }
.brand-mark i:nth-child(2), .brand-mark i:nth-child(3) { background: #00d4d4; }
.setup-title { font-size: 22px; font-family: ui-monospace, Menlo, monospace; font-weight: 500; color: #0b0b0d; }
.setup-desc { font-size: 14px; line-height: 1.6; }
.setup-card :deep(.arco-card-body) { padding: 24px; }
.setup-card :deep(.arco-form-item-label-col > .arco-form-item-label) {
  font-family: ui-monospace, Menlo, monospace;
  font-size: 11px;
  font-weight: 500;
  letter-spacing: .04em;
  text-transform: uppercase;
  color: #6b6b70;
}
.optional { color: #9a9a9f; text-transform: none; letter-spacing: 0; }
.setup-note { text-align: center; margin-top: 16px; font-size: 12px; color: #9a9a9f; }
.setup-success { max-width: 520px; margin: 0 auto; padding: 56px 24px 80px; background: #f7f6f3; min-height: 100vh; }
.setup-success :deep(.arco-result) { padding: 0 0 32px; background: transparent; }
.api-key-card { border-color: #00d4d4; margin-bottom: 24px; }
.api-key-label { font-family: ui-monospace, Menlo, monospace; font-size: 11px; font-weight: 500; letter-spacing: .04em; text-transform: uppercase; color: #007a7a; margin-bottom: 4px; }
.api-key-row { display: flex; align-items: center; gap: 8px; margin-top: 8px; }
.api-key-value { flex: 1; min-width: 0; padding: 8px 12px; border: 1px solid #e4e2dc; background: #fff; font-family: ui-monospace, Menlo, monospace; font-size: 13px; color: #0b0b0d; word-break: break-all; }
.api-key-copy { flex-shrink: 0; }
.api-key-type { margin-top: 8px; font-family: ui-monospace, Menlo, monospace; font-size: 11px; color: #6b6b70; }
.next-steps { margin-bottom: 24px; }
.next-steps-title { font-size: 13px; font-weight: 600; margin-bottom: 12px; color: #0b0b0d; }
.next-step { font-size: 13px; line-height: 1.8; color: #6b6b70; }
.dashboard-button { width: 100%; }
</style>
