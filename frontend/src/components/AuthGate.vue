<template>
  <div class="auth-wrap">
    <a-card class="auth-card">
      <div class="mono muted">identity / developer-portal</div>
      <template v-if="step === 'email'">
        <a-typography-title :heading="3">Sign in</a-typography-title>
        <a-typography-paragraph type="secondary">Use your email to receive a six-digit code.</a-typography-paragraph>
        <a-form :model="{}" layout="vertical" @submit.prevent="sendOtp">
          <a-form-item label="Email" :validate-status="emailError ? 'error' : undefined" :help="emailError">
            <a-input v-model="email" type="text" inputmode="email" placeholder="you@company.com" autofocus />
          </a-form-item>
          <a-button type="primary" long html-type="submit" :loading="loading">{{ loading ? 'Sending code…' : 'Continue with email' }}</a-button>
        </a-form>
        <a-divider>or</a-divider>
        <a-button long disabled>Continue with GitHub</a-button>
        <a-typography-paragraph type="secondary" class="foot">Self-hosted community edition.</a-typography-paragraph>
      </template>

      <template v-else>
        <a-typography-title :heading="3">Check your email</a-typography-title>
        <a-typography-paragraph type="secondary">We sent a code to {{ email }}.</a-typography-paragraph>
        <div class="otp-grid" @paste.prevent="pasteOtp">
          <a-input v-for="(digit, index) in digits" :key="index" :ref="(el) => setOtpRef(el, index)" v-model="digits[index]" :max-length="1" :disabled="loading" class="otp-input" @input="onDigit(index, $event)" @keydown.backspace="onBackspace(index, $event)" />
        </div>
        <a-space style="margin-top:16px" fill>
          <a-button @click="step = 'email'">Use different email</a-button>
          <a-button :disabled="cooldown > 0 || loading" @click="sendOtp">{{ cooldown > 0 ? `Resend in ${cooldown}s` : 'Resend code' }}</a-button>
        </a-space>
        <a-alert v-if="devCode" type="info" style="margin-top:14px">Development code: <span class="mono">{{ devCode }}</span></a-alert>
      </template>
      <a-alert v-if="error" type="error" style="margin-top:14px">{{ error }}</a-alert>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, reactive, ref } from 'vue'
import { Message } from '@arco-design/web-vue'
import { api, apiMessage } from '../lib/api'
import { useAuthStore } from '../stores/auth'

const emit = defineEmits<{ authed: [] }>()
const auth = useAuthStore()
const step = ref<'email' | 'otp'>('email')
const email = ref(auth.email)
const digits = reactive(['', '', '', '', '', ''])
const loading = ref(false)
const error = ref('')
const emailError = ref('')
const devCode = ref('')
const cooldown = ref(0)
const otpRefs: any[] = []
let timer: number | undefined
function setOtpRef(el: any, index: number) { otpRefs[index] = el }
function startCooldown() {
  cooldown.value = 30
  window.clearInterval(timer)
  timer = window.setInterval(() => { cooldown.value--; if (cooldown.value <= 0) window.clearInterval(timer) }, 1000)
}
async function sendOtp() {
  const value = email.value.trim()
  if (!value) { emailError.value = 'Please enter your email address'; return }
  if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value)) { emailError.value = 'Please enter a valid email address'; return }
  emailError.value = ''; loading.value = true; error.value = ''
  try {
    const { data } = await api.post('/auth/developer/otp/send', { email: value })
    step.value = 'otp'; startCooldown()
    if (data.code && data.self_hosted) { devCode.value = data.code; fillCode(data.code); Message.success('Development code filled'); setTimeout(() => verify(data.code), 250) }
    else { devCode.value = ''; Message.success('Code sent') }
  } catch (e) { error.value = apiMessage(e) } finally { loading.value = false }
}
function fillCode(code: string) { code.split('').slice(0, 6).forEach((digit, index) => { digits[index] = digit }) }
function onDigit(index: number, value: string) {
  value = value.replace(/\D/g, '').slice(-1)
  digits[index] = value
  if (value && index < 5) otpRefs[index + 1]?.focus()
  if (value && index === 5 && digits.every((digit) => digit.length === 1)) verify(digits.join(''))
}
function onBackspace(index: number, event: KeyboardEvent) { if (!digits[index] && index > 0) { event.preventDefault(); otpRefs[index - 1]?.focus() } }
function pasteOtp(event: ClipboardEvent) {
  const code = (event.clipboardData?.getData('text') || '').replace(/\D/g, '').slice(0, 6)
  if (!code) return
  fillCode(code)
  if (code.length === 6) verify(code); else otpRefs[code.length]?.focus()
}
async function verify(code: string) {
  loading.value = true; error.value = ''
  try {
    const { data } = await api.post('/auth/developer/otp/verify', { email: email.value, code })
    if (!data.token) throw new Error('Server did not return a session token')
    auth.setDeveloperToken(data.token, email.value)
    Message.success('Signed in')
    emit('authed')
  } catch (e) { error.value = apiMessage(e); digits.fill(''); otpRefs[0]?.focus() } finally { loading.value = false }
}
onBeforeUnmount(() => window.clearInterval(timer))
</script>

<style scoped>
.auth-wrap { min-height: 100vh; display: grid; place-items: center; padding: 24px; background: #f7f6f3; }
.auth-card { width: min(420px, 100%); }
.mono { font-family: ui-monospace, Menlo, monospace; }
.muted { color: #6b6b70; font-size: 12px; }
.otp-grid { display: flex; gap: 8px; justify-content: center; }
.otp-input { width: 44px; height: 52px; text-align: center; font-size: 22px; font-family: ui-monospace, monospace; }
.otp-input :deep(input) { text-align: center; font-size: 22px; font-weight: 600; }
.foot { margin-top: 16px; font-size: 12px; text-align: center; }
</style>
