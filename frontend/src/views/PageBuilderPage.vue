<template>
  <div class="page">
    <div class="card-title"><div><div class="eyebrow">Developer</div><h2 style="margin:6px 0 4px">Hosted page builder</h2><p class="muted">Configure the customer-facing verification page.</p></div><a-button type="primary" @click="save" :loading="saving">Save</a-button></div>
    <div class="grid grid-2" style="margin-top:18px">
      <div class="panel">
        <a-alert v-if="!auth.isDeveloper" type="warning" style="margin-bottom:14px">Connect an API key in the developer portal to load and save this page.</a-alert>
        <a-form :model="{}" layout="vertical">
          <a-form-item label="Public slug"><a-input v-model="slug" placeholder="acme-verify" /></a-form-item>
          <a-form-item label="Company name"><a-input v-model="config.company_name" /></a-form-item>
          <a-form-item label="Accent color"><a-input v-model="config.accent_color" placeholder="#2563eb" /></a-form-item>
          <a-form-item label="Logo URL"><a-input v-model="config.logo_url" /></a-form-item>
          <a-form-item label="Welcome copy"><a-textarea v-model="config.welcome_text" :rows="4" /></a-form-item>
          <a-space><a-button type="primary" @click="save" :loading="saving">Save configuration</a-button><a-button @click="load">Reset</a-button></a-space>
        </a-form>
      </div>
      <div class="panel">
        <div class="card-title"><h3 style="margin:0">Preview</h3><a-tag>{{ slug || 'no-slug' }}</a-tag></div>
        <div class="preview" :style="{ borderColor: config.accent_color || '#2563eb' }">
          <div class="preview-brand">{{ config.company_name || 'Your company' }}</div>
          <h2>{{ config.welcome_text || 'Verify your identity' }}</h2>
          <p class="muted">You will need a government-issued ID and a selfie.</p>
          <a-button type="primary" long>Start verification</a-button>
        </div>
      </div>
    </div>
    <a-alert v-if="message" style="margin-top:16px" :type="error ? 'error' : 'success'">{{ message }}</a-alert>
  </div>
</template>
<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { api, apiMessage } from '../lib/api'
import { useAuthStore } from '../stores/auth'
const auth = useAuthStore()
const slug = ref(''); const saving = ref(false); const message = ref(''); const error = ref(false)
const config = reactive<Record<string, string>>({ company_name: '', accent_color: '#2563eb', logo_url: '', welcome_text: 'Verify your identity' })
async function load() {
  if (!auth.isDeveloper) return
  try {
    const { data } = await api.get('/developer/settings/page-builder')
    const value = data.data?.config || {}
    if (typeof value === 'string') Object.assign(config, JSON.parse(value || '{}')); else Object.assign(config, value)
    slug.value = data.data?.slug || ''
  } catch (e) { error.value = true; message.value = apiMessage(e) }
}
async function save() {
  saving.value = true
  try {
    await api.put('/developer/settings/page-builder', { config })
    if (slug.value) await api.put('/developer/settings/page-builder/slug', { slug: slug.value })
    message.value = 'Page configuration saved'; error.value = false
  } catch (e) { error.value = true; message.value = apiMessage(e) } finally { saving.value = false }
}
onMounted(load)
</script>
<style scoped>
.preview { margin-top: 18px; border: 2px solid; border-radius: 16px; padding: 40px 28px; text-align: center; background: linear-gradient(180deg, #f8fbff, #fff); }
.preview-brand { font-weight: 700; margin-bottom: 20px; }
.preview h2 { margin: 0 0 8px; }
</style>
