<template>
  <div v-if="checkingAuth" class="auth-check">
    <a-space direction="vertical" align="center" :size="14">
      <a-spin :size="28" />
      <span class="mono-muted">Authenticating...</span>
    </a-space>
  </div>

  <div v-else class="auth-page">
    <div class="auth-shell">
      <!-- Logo + header -->
      <div class="auth-head">
        <div class="logo"><span class="brand-mark"><i /><i /><i /><i /></span><span>identity</span></div>
        <div class="eyebrow">identity / review-dashboard</div>
        <a-typography-title :heading="3" class="auth-title" style="margin:0 0 8px">Verification Management</a-typography-title>
        <a-typography-text type="secondary" class="auth-sub">
          {{ step === 'email' ? 'Enter your reviewer email to sign in' : `Enter the code sent to ${email}` }}
        </a-typography-text>
      </div>

      <!-- Login card -->
      <a-card class="auth-card">
        <a-alert v-if="error" type="error" :show-icon="false" class="auth-error">{{ error }}</a-alert>

        <a-form v-if="step === 'email'" :model="formModel" layout="vertical" @submit="sendOtp">
          <a-form-item label="Email Address">
            <a-input
              v-model="email"
              type="text"
              :input-attrs="{ autofocus: true, inputmode: 'email', autocomplete: 'email' }"
              placeholder="reviewer@company.com"
              :disabled="loading"
            />
          </a-form-item>
          <a-button type="primary" long html-type="submit" :loading="loading" :disabled="loading">
            {{ loading ? 'Sending code...' : 'Send verification code' }}
          </a-button>
        </a-form>

        <a-form v-else :model="formModel" layout="vertical" @submit="verifyOtp">
          <a-alert v-if="selfHostedCode" type="info" class="self-hosted">
            <template #title>Self-hosted mode — your code:</template>
            <div class="self-hosted-code">{{ selfHostedCode }}</div>
          </a-alert>

          <a-form-item label="Verification Code">
            <a-input
              ref="otpInput"
              v-model="otpCode"
              class="otp-input"
              :max-length="6"
              :input-attrs="{ inputmode: 'numeric', autocomplete: 'one-time-code' }"
              placeholder="000000"
              :disabled="loading"
              @input="onOtpInput"
            />
          </a-form-item>

          <a-button type="primary" long html-type="submit" :loading="loading" :disabled="loading || otpCode.length !== 6">
            {{ loading ? 'Verifying...' : 'Verify & Sign In' }}
          </a-button>
          <a-button long class="secondary" :disabled="loading" @click="resetEmail">Use a different email</a-button>
        </a-form>
      </a-card>

      <!-- Footer link -->
      <div class="auth-foot">
        <a-link class="foot-link" @click="router.push('/')">← Back to Portal</a-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { nextTick, onMounted, ref } from 'vue'
import { Message } from '@arco-design/web-vue'
import { useRouter } from 'vue-router'
import { api, apiMessage } from '../lib/api'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const router = useRouter()

const email = ref('')
const otpCode = ref('')
const step = ref<'email' | 'otp'>('email')
const loading = ref(false)
const error = ref('')
const selfHostedCode = ref<string | null>(null)
const checkingAuth = ref(true)
const otpInput = ref<any>(null)
const formModel = {}

// Check existing admin session — no escalation, OTP-only
onMounted(async () => {
  if (!auth.isAdmin) {
    checkingAuth.value = false
    return
  }
  try {
    await api.get('/admin/dashboard')
    router.replace('/admin/verifications')
  } catch {
    checkingAuth.value = false
  }
})

function onOtpInput(value: string) {
  otpCode.value = String(value ?? '').replace(/\D/g, '').slice(0, 6)
}

async function sendOtp() {
  const value = email.value.trim()
  if (!value) {
    error.value = 'Enter your reviewer email'
    return
  }
  loading.value = true
  error.value = ''
  selfHostedCode.value = null
  try {
    const { data } = await api.post('/auth/reviewer/otp/send', { email: value })
    email.value = value
    if (data?.self_hosted && data?.code) selfHostedCode.value = String(data.code)
    step.value = 'otp'
  } catch (e) {
    error.value = apiMessage(e)
    return
  } finally {
    loading.value = false
  }
  // Focus the code field once the OTP step is rendered and no longer disabled
  await nextTick()
  otpInput.value?.focus?.()
}

async function verifyOtp() {
  loading.value = true
  error.value = ''
  try {
    const { data } = await api.post('/auth/reviewer/otp/verify', { email: email.value, code: otpCode.value })
    if (!data?.token) {
      error.value = 'Login failed: invalid response from server'
      return
    }
    // Token is kept in the admin session store (Bearer auth) — same scope as the cookie in the reference
    auth.setAdminToken(String(data.token), email.value)
    Message.success('Signed in')
    router.replace('/admin/verifications')
  } catch (e) {
    error.value = apiMessage(e)
  } finally {
    loading.value = false
  }
}

function resetEmail() {
  step.value = 'email'
  otpCode.value = ''
  error.value = ''
  selfHostedCode.value = null
}
</script>

<style scoped>
.auth-check { min-height: 100vh; display: flex; align-items: center; justify-content: center; background: #f7f6f3; padding: 24px; }
.mono-muted { font-family: ui-monospace, Menlo, monospace; font-size: 13px; letter-spacing: .04em; color: #6b6b70; }
.auth-page { min-height: 100vh; background: #f7f6f3; display: flex; align-items: center; justify-content: center; padding: 40px 24px; }
.auth-shell { width: 100%; max-width: 420px; }
.auth-head { text-align: center; margin-bottom: 32px; }
.logo { display: flex; align-items: center; justify-content: center; gap: 8px; margin-bottom: 24px; font-family: ui-monospace, Menlo, monospace; font-size: 14px; color: #0b0b0d; }
.brand-mark { width: 18px; height: 18px; display: grid; grid-template-columns: 1fr 1fr; grid-template-rows: 1fr 1fr; }
.brand-mark i:nth-child(1), .brand-mark i:nth-child(4) { background: #0b0b0d; }
.brand-mark i:nth-child(2), .brand-mark i:nth-child(3) { background: #00d4d4; }
.auth-head .eyebrow { margin-bottom: 8px; }
.auth-title { font-family: ui-monospace, Menlo, monospace; font-weight: 600; color: #0b0b0d; }
.auth-sub { font-size: 14px; }
.auth-card { border-radius: 0; }
.auth-card :deep(.arco-card-body) { padding: 32px; }
.auth-card :deep(.arco-form-item-label-col > .arco-form-item-label) {
  font-family: ui-monospace, Menlo, monospace;
  font-size: 11px;
  font-weight: 500;
  letter-spacing: .05em;
  text-transform: uppercase;
  color: #6b6b70;
}
.auth-error { margin-bottom: 20px; }
.self-hosted { margin-bottom: 20px; }
.self-hosted-code { text-align: center; font-family: ui-monospace, Menlo, monospace; font-size: 24px; font-weight: 700; letter-spacing: .15em; color: #007a7a; }
.otp-input :deep(.arco-input) { text-align: center; font-size: 24px; font-family: ui-monospace, Menlo, monospace; letter-spacing: .2em; }
.secondary { margin-top: 12px; }
.auth-foot { text-align: center; margin-top: 24px; }
.foot-link { font-family: ui-monospace, Menlo, monospace; font-size: 13px; color: #6b6b70; }
.foot-link:hover { color: #0b0b0d; }
</style>
