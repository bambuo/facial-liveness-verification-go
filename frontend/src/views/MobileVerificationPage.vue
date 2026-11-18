<template>
  <div class="mobile-shell">
    <main class="mobile-card">
      <div class="eyebrow">Mobile verification</div>
      <h1>{{ branding?.company_name || 'Verify your identity' }}</h1>
      <p class="muted">{{ session?.source === 'vaas' ? 'Continue your verification.' : 'Complete this verification on your mobile device.' }}</p>
      <a-alert v-if="error" type="error" style="margin:16px 0">{{ message }}</a-alert>
      <a-alert v-else-if="session" type="success" style="margin:16px 0">Session loaded. Keep this page open while you complete the steps.</a-alert>
      <div v-if="session" class="session-box">
        <div><span class="muted">User</span><strong>{{ session.user_id }}</strong></div>
        <div><span class="muted">Session</span><strong class="mono">{{ shortToken }}</strong></div>
      </div>
      <a-button type="primary" long size="large" :disabled="!session" @click="continueVerification">Continue verification</a-button>
      <a-button long style="margin-top:10px" @click="$router.push('/docs')">Need help?</a-button>
    </main>
  </div>
</template>
<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { api, apiMessage } from '../lib/api'
import { useAuthStore } from '../stores/auth'
const route = useRoute(); const auth = useAuthStore()
const token = ref(String(route.query.token || '')); const session = ref<any>(null); const branding = ref<any>(null); const message = ref(''); const error = ref(false)
const shortToken = computed(() => token.value ? token.value.slice(0, 12) + '…' : '')
async function load() {
  if (!token.value) { error.value = true; message.value = 'Missing handoff token'; return }
  try {
    const { data } = await api.get(`/verify/handoff/${token.value}/session`)
    session.value = data; branding.value = data.branding; auth.setSessionToken(token.value)
  } catch (e) { error.value = true; message.value = apiMessage(e) }
}
function continueVerification() { window.location.href = `/user-verification${session.value?.verification_id ? `?verification_id=${session.value.verification_id}` : ''}` }
onMounted(load)
</script>
<style scoped>
.mobile-shell { min-height: 100vh; display: grid; place-items: center; padding: 20px; background: #eff6ff; }
.mobile-card { width: min(430px, 100%); background: #fff; border: 1px solid #dbeafe; border-radius: 20px; padding: 28px; box-shadow: 0 18px 50px rgba(37,99,235,.12); }
.mobile-card h1 { margin: 8px 0; font-size: 28px; }
.session-box { background: #f8fafc; border: 1px solid #e2e8f0; border-radius: 12px; padding: 14px; display: grid; gap: 8px; margin-bottom: 16px; }
.session-box div { display: flex; justify-content: space-between; gap: 10px; font-size: 13px; }
</style>
