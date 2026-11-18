<template>
  <div class="page-narrow">
    <div class="eyebrow">First run</div><h1>Set up your workspace</h1>
    <p class="muted">Check initialization status, then create the first developer account.</p>
    <div class="panel" style="margin-top:20px">
      <div class="card-title"><h3 style="margin:0">Deployment status</h3><a-tag :color="initialized ? 'green' : 'orange'">{{ initialized ? 'Initialized' : 'Pending' }}</a-tag></div>
      <a-descriptions :column="1" style="margin-top:16px">
        <a-descriptions-item label="API">Connected</a-descriptions-item>
        <a-descriptions-item label="Setup">/api/setup/status</a-descriptions-item>
      </a-descriptions>
      <a-space style="margin-top:16px"><a-button @click="refresh">Refresh</a-button><a-button type="primary" @click="initialize">Initialize</a-button></a-space>
    </div>
    <a-alert v-if="message" style="margin-top:16px" :type="error ? 'error' : 'info'">{{ message }}</a-alert>
  </div>
</template>
<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { api, apiMessage } from '../lib/api'
const initialized = ref(false); const message = ref(''); const error = ref(false)
async function refresh() {
  try { const { data } = await api.get('/setup/status'); initialized.value = !!data.initialized } catch (e) { error.value = true; message.value = apiMessage(e) }
}
async function initialize() {
  try { const { data } = await api.post('/setup/initialize', {}); message.value = data.message || 'Initialized'; initialized.value = true; error.value = false } catch (e) { error.value = true; message.value = apiMessage(e) }
}
onMounted(refresh)
</script>
