<template>
  <div class="verify-shell">
    <header class="verify-header"><router-link to="/" class="brand">Identity Verification</router-link><a-tag color="arcoblue">Secure session</a-tag></header>
    <main class="verify-main">
      <div class="verify-card">
        <div class="eyebrow">Identity verification</div>
        <h1>{{ verificationId ? 'Continue your verification' : 'Verify your identity' }}</h1>
        <p class="muted">You will need a government-issued ID and a selfie. It usually takes under two minutes.</p>

        <a-steps :current="stepIndex" style="margin:24px 0">
          <a-step>Start</a-step><a-step>Document</a-step><a-step>Selfie</a-step><a-step>Complete</a-step>
        </a-steps>

        <a-alert v-if="!auth.apiKey" type="warning" style="margin-bottom:16px">This demo needs an API key. Configure one in the developer portal, then return.</a-alert>
        <a-form v-if="!verificationId" :model="{}" layout="vertical">
          <a-form-item label="User ID"><a-input v-model="userID" placeholder="UUID" /></a-form-item>
          <a-form-item label="API key"><a-input-password v-model="apiKey" placeholder="ik_..." /></a-form-item>
          <a-button type="primary" long :loading="loading" @click="initialize">Start verification</a-button>
        </a-form>

        <template v-else>
          <div class="session-row"><span class="muted">Verification</span><span class="mono">{{ verificationId }}</span><a-tag :color="statusColor">{{ status }}</a-tag></div>
          <a-divider />
          <div v-if="stepIndex === 1">
            <h3>Front of your ID</h3><p class="muted">Passport, driving licence or national ID. Keep all four corners visible.</p>
            <input ref="frontInput" type="file" accept="image/jpeg,image/png" capture="environment" @change="uploadFront" />
            <a-button type="primary" long style="margin-top:14px" :loading="uploading" @click="uploadFront(null)">Upload front document</a-button>
            <a-collapse style="margin-top:16px"><a-collapse-item header="Add back of ID (optional)" key="back"><input ref="backInput" type="file" accept="image/jpeg,image/png" @change="uploadBack" /><a-button size="small" style="margin-top:10px" @click="uploadBack(null)">Upload back</a-button></a-collapse-item></a-collapse>
          </div>
          <div v-else-if="stepIndex === 2">
            <h3>Take a selfie</h3><p class="muted">Use a well-lit photo without sunglasses or a hat.</p>
            <input ref="selfieInput" type="file" accept="image/jpeg,image/png" capture="user" @change="uploadSelfie" />
            <div class="voice-box"><div><strong>Voice challenge (optional)</strong><div class="muted">Read six digits to add voice verification.</div></div><a-button size="small" @click="voiceChallenge">Get challenge</a-button></div>
            <a-alert v-if="voiceDigits" type="info" style="margin-top:10px">Read these digits: <strong class="mono">{{ voiceDigits }}</strong></a-alert>
            <div v-if="voiceDigits" class="voice-upload"><input ref="voiceInput" type="file" accept="audio/wav" @change="uploadVoice" /><a-button size="small" @click="uploadVoice(null)">Upload voice</a-button><span v-if="voiceResult" class="muted">{{ voiceResult }}</span></div>
            <a-button type="primary" long style="margin-top:14px" :loading="uploading" @click="uploadSelfie(null)">Upload selfie</a-button>
          </div>
          <div v-else>
            <h2>Verification complete</h2>
            <p class="muted">Final result: <strong>{{ finalResult || 'pending' }}</strong></p>
            <a-space><a-button type="primary" @click="refreshStatus">Refresh status</a-button><a-button @click="restart">Start over</a-button></a-space>
          </div>
        </template>
        <a-alert v-if="message" style="margin-top:16px" :type="error ? 'error' : 'success'">{{ message }}</a-alert>
      </div>
      <aside class="verify-aside"><h3>Privacy</h3><p>Documents are processed for this verification. Face embeddings are held in memory for matching and are not stored.</p><h3>Need help?</h3><p>Contact the organization that sent you this link.</p></aside>
    </main>
  </div>
</template>
<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { Message } from '@arco-design/web-vue'
import { api, apiMessage } from '../lib/api'
import { useAuthStore } from '../stores/auth'
const auth = useAuthStore()
const userID = ref(crypto.randomUUID()); const apiKey = ref(auth.apiKey); const verificationId = ref(''); const status = ref(''); const finalResult = ref(''); const message = ref(''); const error = ref(false)
const loading = ref(false); const uploading = ref(false); const voiceDigits = ref(''); const voiceResult = ref(''); const step = ref(0)
const frontInput = ref<HTMLInputElement | null>(null); const backInput = ref<HTMLInputElement | null>(null); const selfieInput = ref<HTMLInputElement | null>(null); const voiceInput = ref<HTMLInputElement | null>(null)
const stepIndex = computed(() => step.value)
const statusColor = computed(() => status.value === 'COMPLETE' ? 'green' : status.value === 'HARD_REJECTED' ? 'red' : 'arcoblue')
async function initialize() {
  if (!apiKey.value) { Message.warning('API key is required'); return }
  loading.value = true
  try {
    auth.setAPIKey(apiKey.value.trim())
    const { data } = await api.post('/v2/verify/initialize', { user_id: userID.value })
    verificationId.value = data.verification_id; status.value = data.status; auth.setSessionToken(data.session_token || '')
    message.value = 'Verification started.'; error.value = false; step.value = 1
  } catch (e) { error.value = true; message.value = apiMessage(e) } finally { loading.value = false }
}
async function uploadFront(event: Event | null) {
  const file = event ? (event.target as HTMLInputElement).files?.[0] : frontInput.value?.files?.[0]
  if (!file) { Message.warning('Choose a document file'); return }
  const data = await upload('front-document', 'document', file)
  step.value = data?.requires_back ? 1 : 2
}
async function uploadBack(event: Event | null) {
  const file = event ? (event.target as HTMLInputElement).files?.[0] : backInput.value?.files?.[0]
  if (!file) { Message.warning('Choose a back document file'); return }
  await upload('back-document', 'document', file)
  step.value = 2
}
async function uploadSelfie(event: Event | null) {
  const file = event ? (event.target as HTMLInputElement).files?.[0] : selfieInput.value?.files?.[0]
  if (!file) { Message.warning('Choose a selfie file'); return }
  await upload('live-capture', 'selfie', file)
  step.value = 3
}
async function uploadVoice(event: Event | null) {
  const input = event ? (event.target as HTMLInputElement) : (document.querySelector('input[accept="audio/wav"]') as HTMLInputElement | null)
  const file = input?.files?.[0]
  if (!file) { Message.warning('Choose a WAV recording'); return }
  try {
    const form = new FormData(); form.append('file', file)
    const { data } = await api.post(`/v2/verify/${verificationId.value}/voice-capture`, form)
    voiceResult.value = `Voice score ${data.voice_match_results?.score ?? '—'} · ${data.voice_match_results?.transcription || ''}`
    Message.success('Voice captured')
  } catch (e) { Message.error(apiMessage(e)) }
}
async function upload(path: string, field: string, file: File) {
  uploading.value = true
  try {
    const form = new FormData(); form.append(field, file)
    const { data } = await api.post(`/v2/verify/${verificationId.value}/${path}`, form)
    status.value = data.status || status.value
    if (path === 'live-capture') await finalize()
    else message.value = data.message || 'Uploaded'; error.value = false
    return data
  } catch (e) { error.value = true; message.value = apiMessage(e); return null } finally { uploading.value = false }
}
async function finalize() {
  try { const { data } = await api.post(`/v2/verify/${verificationId.value}/finalize`); status.value = data.status; finalResult.value = data.final_result || ''; message.value = data.message || 'Verification finalized'; error.value = false }
  catch (e) { error.value = true; message.value = apiMessage(e) }
}
async function voiceChallenge() { try { const { data } = await api.post(`/v2/verify/${verificationId.value}/voice-challenge`); voiceDigits.value = data.challenge_digits } catch (e) { Message.error(apiMessage(e)) } }
async function refreshStatus() { try { const { data } = await api.get(`/v2/verify/${verificationId.value}/status`); status.value = data.status; finalResult.value = data.final_result || '' } catch (e) { Message.error(apiMessage(e)) } }
function restart() { verificationId.value = ''; status.value = ''; finalResult.value = ''; message.value = ''; step.value = 0; auth.setSessionToken('') }
onMounted(() => { if (auth.apiKey) apiKey.value = auth.apiKey })
</script>
<style scoped>
.verify-shell { min-height: 100vh; background: linear-gradient(180deg, #eff6ff, #f8fafc 40%); }
.verify-header { height: 64px; display: flex; align-items: center; justify-content: space-between; padding: 0 24px; background: #fff; border-bottom: 1px solid #e2e8f0; }
.brand { font-weight: 700; color: #0f172a; text-decoration: none; }
.verify-main { max-width: 1050px; margin: 0 auto; padding: 34px 20px 60px; display: grid; grid-template-columns: minmax(0, 1fr) 280px; gap: 22px; align-items: start; }
.verify-card { background: #fff; border: 1px solid #e2e8f0; border-radius: 18px; padding: 30px; }
.verify-aside { background: #fff; border: 1px solid #e2e8f0; border-radius: 14px; padding: 20px; }
.verify-aside h3 { margin: 0 0 6px; }
.verify-aside p { color: #64748b; line-height: 1.6; margin: 0 0 16px; }
.session-row { display: flex; gap: 10px; align-items: center; flex-wrap: wrap; }
.voice-upload { display:flex; align-items:center; gap:10px; margin-top:10px; font-size:13px; }
.voice-box { display: flex; align-items: center; justify-content: space-between; gap: 12px; margin-top: 16px; padding: 14px; border: 1px dashed #cbd5e1; border-radius: 10px; }
@media (max-width: 860px) { .verify-main { grid-template-columns: 1fr; } .verify-aside { display: none; } }
</style>
