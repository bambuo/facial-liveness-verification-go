<template>
  <div class="page">
    <div class="card-title"><div><div class="eyebrow">Operations</div><h2 style="margin:6px 0 4px">Verification thresholds</h2><p class="muted">Tune production and sandbox decision boundaries.</p></div><a-button @click="load" :loading="loading">Refresh</a-button></div>
    <div class="grid grid-2" style="margin-top:18px">
      <div class="panel" v-for="env in ['production','sandbox']" :key="env">
        <div class="card-title"><h3 style="margin:0">{{ env }}</h3><a-tag>{{ env === 'production' ? 'Live' : 'Test' }}</a-tag></div>
        <a-form :model="{}" layout="vertical" style="margin-top:14px">
          <a-form-item v-for="(value, key) in data[env]" :key="key" :label="String(key)">
            <a-input-number v-model="data[env][key]" :min="0" :max="1" :step="0.01" style="width:100%" />
          </a-form-item>
        </a-form>
      </div>
    </div>
    <div class="panel" style="margin-top:18px">
      <div class="toolbar"><a-button type="primary" @click="save" :loading="saving">Save thresholds</a-button><a-button @click="preview">Preview impact</a-button><a-button status="warning" @click="reset">Reset defaults</a-button></div>
      <a-alert v-if="message" style="margin-top:14px" :type="error ? 'error' : 'success'">{{ message }}</a-alert>
    </div>
  </div>
</template>
<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { Message } from '@arco-design/web-vue'
import { api, apiMessage } from '../lib/api'
const data = reactive<Record<string, Record<string, number>>>({ production: {}, sandbox: {} })
const loading = ref(false); const saving = ref(false); const message = ref(''); const error = ref(false)
async function load() {
  loading.value = true
  try { const { data: res } = await api.get('/admin/thresholds'); data.production = res.data.production || {}; data.sandbox = res.data.sandbox || {} }
  catch (e) { error.value = true; message.value = apiMessage(e) } finally { loading.value = false }
}
async function save() {
  saving.value = true
  try {
    await api.put('/admin/thresholds', { auto_approve_threshold: data.production.face_matching ?? 0.85, manual_review_threshold: data.production.liveness ?? 0.75 })
    message.value = 'Thresholds saved'; error.value = false; load()
  } catch (e) { error.value = true; message.value = apiMessage(e) } finally { saving.value = false }
}
async function preview() { const { data: res } = await api.post('/admin/thresholds/preview', { ...data.production }); message.value = res.data?.explanation || 'Preview generated'; error.value = false }
async function reset() { await api.post('/admin/thresholds/reset', {}); message.value = 'Reset to defaults'; error.value = false; load() }
onMounted(load)
</script>
