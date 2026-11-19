<template>
  <div class="operator-shell">
    <a-card class="operator-card">
      <!-- Eyebrow -->
      <div class="mono-eyebrow">identity / service-operator</div>

      <!-- Inline error (shared across all steps) -->
      <a-alert v-if="error" type="error" :show-icon="false" class="operator-error">{{ error }}</a-alert>

      <!-- Step: email -->
      <template v-if="step === 'email'">
        <a-typography-title :heading="3" class="step-title" style="margin:0 0 8px">Operator sign in</a-typography-title>
        <a-typography-text type="secondary" class="step-copy">Enter your email to receive a verification code</a-typography-text>
        <a-form :model="formModel" layout="vertical" class="step-form" @submit="sendCode">
          <a-form-item label="Email">
            <a-input
              v-model="email"
              type="text"
              :input-attrs="{ autofocus: true, inputmode: 'email', autocomplete: 'email' }"
              placeholder="operator@company.com"
              :disabled="loading"
            />
          </a-form-item>
          <a-button type="primary" long html-type="submit" :loading="loading" :disabled="loading">
            {{ loading ? 'Sending code...' : 'Send verification code' }}
          </a-button>
        </a-form>
      </template>

      <!-- Step: OTP -->
      <template v-else-if="step === 'otp'">
        <a-typography-title :heading="3" class="step-title" style="margin:0 0 8px">Check your email</a-typography-title>
        <a-typography-text type="secondary" class="step-copy">
          We sent a 6-digit code to <span class="ink">{{ email }}</span>
        </a-typography-text>
        <a-form :model="formModel" layout="vertical" class="step-form" @submit="verifyCode">
          <a-form-item label="Verification code">
            <a-input
              ref="otpRef"
              v-model="otp"
              class="otp-input"
              :max-length="6"
              :input-attrs="{ inputmode: 'numeric', autocomplete: 'one-time-code' }"
              placeholder="000000"
              :disabled="loading"
              @input="onOtpInput"
            />
          </a-form-item>
          <a-button type="primary" long html-type="submit" :loading="loading" :disabled="loading || otp.length !== 6">
            {{ loading ? 'Verifying...' : 'Verify & sign in' }}
          </a-button>
          <a-button type="text" class="inline-back" :disabled="loading" @click="backToEmail">← Use a different email</a-button>
        </a-form>
      </template>

      <!-- Step: key picker -->
      <template v-else>
        <a-typography-title :heading="3" class="step-title" style="margin:0 0 8px">Select a service key</a-typography-title>
        <a-typography-text type="secondary" class="step-copy">Multiple service keys are linked to this account. Choose one to continue.</a-typography-text>
        <a-form :model="formModel" layout="vertical" class="step-form" @submit="selectKey">
          <a-radio-group v-model="selectedKeyId" direction="vertical" class="key-list">
            <a-radio v-for="key in keys" :key="key.api_key_id" :value="key.api_key_id" :disabled="loading" class="key-option">
              <span class="key-label">{{ keyLabel(key) }}</span>
              <span class="key-meta">{{ key.service_product || '—' }} · {{ key.service_environment || 'unknown' }}</span>
            </a-radio>
          </a-radio-group>
          <a-button type="primary" long html-type="submit" :loading="loading" :disabled="loading || !selectedKeyId">
            {{ loading ? 'Signing in...' : 'Continue' }}
          </a-button>
        </a-form>
      </template>

      <!-- Footer -->
      <div class="operator-foot">
        <a-link class="foot-link" @click="router.push('/')">← Back to Portal</a-link>
      </div>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { nextTick, ref } from 'vue'
import { Message } from '@arco-design/web-vue'
import { useRouter } from 'vue-router'
import { api, apiMessage } from '../lib/api'
import { useAuthStore } from '../stores/auth'

type Step = 'email' | 'otp' | 'select'

interface OperatorKeyOption {
  api_key_id: string
  service_product: string | null
  service_environment: string | null
  service_label: string | null
}

const auth = useAuthStore()
const router = useRouter()

const step = ref<Step>('email')
const email = ref('')
const otp = ref('')
const selectionToken = ref('')
const keys = ref<OperatorKeyOption[]>([])
const selectedKeyId = ref('')
const error = ref('')
const loading = ref(false)
const otpRef = ref<any>(null)
const formModel = {}

function keyLabel(key: OperatorKeyOption) {
  return key.service_label ?? key.service_product ?? 'Unnamed service'
}

function onOtpInput(value: string) {
  otp.value = String(value ?? '').replace(/\D/g, '').slice(0, 6)
}

function backToEmail() {
  step.value = 'email'
  otp.value = ''
  error.value = ''
}

async function focusOtp() {
  await nextTick()
  otpRef.value?.focus?.()
}

// Step 1: send OTP
async function sendCode() {
  const value = email.value.trim()
  if (!value) {
    error.value = 'Enter your email'
    return
  }
  loading.value = true
  error.value = ''
  try {
    // Community backend exposes the shared email OTP routes; operator scope is carried by the service keys on the account
    await api.post('/auth/developer/otp/send', { email: value })
    email.value = value
    otp.value = ''
    step.value = 'otp'
  } catch (e) {
    error.value = apiMessage(e)
    return
  } finally {
    loading.value = false
  }
  await focusOtp()
}

// Step 2: verify OTP — single key signs in, multiple keys open the picker
async function verifyCode() {
  loading.value = true
  error.value = ''
  try {
    const { data } = await api.post('/auth/developer/otp/verify', { email: email.value, code: otp.value })
    if (data?.selection_token) {
      selectionToken.value = String(data.selection_token)
      keys.value = Array.isArray(data.keys) ? (data.keys as OperatorKeyOption[]) : []
      selectedKeyId.value = keys.value[0]?.api_key_id ?? ''
      step.value = 'select'
      return
    }
    if (!data?.token) {
      // No active service key for this account
      error.value = 'No service access for this account'
      step.value = 'email'
      return
    }
    auth.setDeveloperToken(String(data.token), email.value)
    Message.success('Signed in')
    router.replace('/developer')
  } catch (e) {
    const status = (e as { response?: { status?: number } })?.response?.status
    error.value = apiMessage(e)
    // 401 covers both "no active key" and invalid/expired OTP
    if (status === 401) step.value = 'email'
  } finally {
    loading.value = false
  }
}

// Step 3: select a key (when multiple keys exist)
async function selectKey() {
  if (!selectedKeyId.value) return
  loading.value = true
  error.value = ''
  try {
    const { data } = await api.post('/auth/service-operator/otp/select', {
      selection_token: selectionToken.value,
      api_key_id: selectedKeyId.value,
    })
    if (data?.token) auth.setDeveloperToken(String(data.token), email.value)
    Message.success('Signed in')
    router.replace('/developer')
  } catch (e) {
    error.value = apiMessage(e)
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.operator-shell {
  min-height: calc(100vh - 120px);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
  background: #f7f6f3;
}
.operator-card { width: 100%; max-width: 400px; border-radius: 12px; }
.operator-card :deep(.arco-card-body) { padding: 40px; }
.operator-card :deep(.arco-form-item-label-col > .arco-form-item-label) {
  font-family: ui-monospace, Menlo, monospace;
  font-size: 11px;
  font-weight: 500;
  letter-spacing: .05em;
  text-transform: uppercase;
  color: #6b6b70;
}
.mono-eyebrow { font-family: ui-monospace, Menlo, monospace; font-size: 11px; letter-spacing: .08em; color: #6b6b70; margin-bottom: 24px; }
.operator-error { margin-bottom: 20px; }
.step-title { font-size: 22px; font-family: ui-monospace, Menlo, monospace; font-weight: 600; color: #0b0b0d; }
.step-copy { display: block; font-size: 14px; margin-bottom: 24px; }
.step-copy .ink { color: #0b0b0d; }
.step-form { display: flex; flex-direction: column; }
.step-form :deep(.arco-form-item) { margin-bottom: 16px; }
.otp-input :deep(.arco-input) { text-align: center; font-size: 24px; font-family: ui-monospace, Menlo, monospace; letter-spacing: .2em; }
.inline-back { align-self: flex-start; padding: 0; font-size: 13px; color: #6b6b70; }
.inline-back:hover { color: #0b0b0d; }
.key-list { display: flex; flex-direction: column; gap: 8px; margin-bottom: 16px; }
.key-list :deep(.arco-radio) { align-items: flex-start; padding: 12px 14px; border: 1px solid #e4e2dc; border-radius: 8px; background: #fff; }
.key-list :deep(.arco-radio-checked) { border-color: rgba(0, 212, 212, .45); background: rgba(0, 240, 255, .08); }
.key-list :deep(.arco-radio-icon) { margin: 3px 10px 0 0; }
.key-label { display: block; font-size: 14px; font-weight: 600; color: #0b0b0d; }
.key-meta { display: block; margin-top: 2px; font-family: ui-monospace, Menlo, monospace; font-size: 12px; color: #6b6b70; }
.operator-foot { margin-top: 24px; text-align: center; }
.foot-link { font-family: ui-monospace, Menlo, monospace; font-size: 13px; color: #6b6b70; }
.foot-link:hover { color: #0b0b0d; }
</style>
