<template>
  <div class="page-narrow">
    <div class="eyebrow">Verifiable credential</div><h1>Credential verification</h1>
    <p class="muted">Paste a JWT-VC or open a link containing `?jwt=` to check its current status.</p>
    <div class="panel" style="margin-top:18px">
      <a-textarea v-model="jwt" :rows="6" placeholder="eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9..." />
      <a-space style="margin-top:12px"><a-button type="primary" @click="check" :loading="loading">Check credential</a-button><a-button @click="clear">Clear</a-button></a-space>
    </div>
    <a-alert v-if="message" style="margin-top:16px" :type="error ? 'error' : 'success'">{{ message }}</a-alert>
    <div v-if="claims" class="panel" style="margin-top:16px">
      <div class="card-title"><h3 style="margin:0">Credential</h3><a-tag :color="active ? 'green' : 'red'">{{ active ? 'Active' : reason || 'Inactive' }}</a-tag></div>
      <a-descriptions :column="1" bordered size="small" style="margin-top:14px">
        <a-descriptions-item label="Issuer">{{ claims.iss || '—' }}</a-descriptions-item>
        <a-descriptions-item label="JTI">{{ claims.jti || '—' }}</a-descriptions-item>
        <a-descriptions-item label="Subject">{{ JSON.stringify(claims.vc?.credentialSubject || {}) }}</a-descriptions-item>
        <a-descriptions-item label="Expires">{{ claims.exp ? new Date(claims.exp * 1000).toLocaleString() : '—' }}</a-descriptions-item>
      </a-descriptions>
    </div>
  </div>
</template>
<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { Message } from '@arco-design/web-vue'
import { useRoute } from 'vue-router'
import { api, apiMessage } from '../lib/api'
const jwt = ref(''); const claims = ref<any>(null); const active = ref(false); const reason = ref(''); const loading = ref(false); const message = ref(''); const error = ref(false)
const route = useRoute()
function decode(token: string) {
  const part = token.split('.')[1]; if (!part) throw new Error('Invalid JWT')
  const normalized = part.replace(/-/g, '+').replace(/_/g, '/').padEnd(Math.ceil(part.length / 4) * 4, '=')
  return JSON.parse(decodeURIComponent(escape(atob(normalized))))
}
async function check() {
  if (!jwt.value.trim()) { Message.warning('Paste a credential'); return }
  loading.value = true; error.value = false
  try {
    claims.value = decode(jwt.value.trim())
    const { data } = await api.get(`/v2/credentials/${encodeURIComponent(claims.value.jti)}/status`)
    active.value = !!data.active; reason.value = data.reason || ''
    message.value = active.value ? 'Credential is active.' : `Credential is not active: ${reason.value || 'unknown'}`
  } catch (e) { error.value = true; message.value = apiMessage(e) } finally { loading.value = false }
}
function clear() { jwt.value = ''; claims.value = null; message.value = '' }
onMounted(() => { const value = route.query.jwt; if (typeof value === 'string') { jwt.value = value; check() } })
</script>
