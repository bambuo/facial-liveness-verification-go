<template>
  <div class="login-shell">
    <div class="login-card">
      <div class="eyebrow">Operations console</div>
      <h1>Admin sign in</h1>
      <p class="muted">Use the administrator credentials configured for this deployment.</p>
      <a-form :model="{}" layout="vertical" @submit.prevent="submit">
        <a-form-item label="Email"><a-input v-model="email" placeholder="admin@example.com" /></a-form-item>
        <a-form-item label="Password"><a-input-password v-model="password" placeholder="Password" /></a-form-item>
        <a-button type="primary" long html-type="submit" :loading="loading">Sign in</a-button>
      </a-form>
      <a-divider />
      <a-space><a-link @click="$router.push('/')">Back to portal</a-link><a-link @click="$router.push('/admin/verifications')">Already signed in?</a-link></a-space>
    </div>
  </div>
</template>
<script setup lang="ts">
import { ref } from 'vue'
import { Message } from '@arco-design/web-vue'
import { useRoute, useRouter } from 'vue-router'
import { api, apiMessage } from '../lib/api'
import { useAuthStore } from '../stores/auth'
const email = ref(''); const password = ref(''); const loading = ref(false)
const auth = useAuthStore(); const router = useRouter(); const route = useRoute()
async function submit() {
  loading.value = true
  try {
    const { data } = await api.post('/auth/admin/login', { email: email.value, password: password.value })
    auth.setAdminToken(data.token, data.admin?.email || email.value)
    Message.success('Signed in')
    router.replace(String(route.query.redirect || '/admin/verifications'))
  } catch (error) { Message.error(apiMessage(error)) } finally { loading.value = false }
}
</script>
<style scoped>
.login-shell { min-height: 100vh; display: grid; place-items: center; background: radial-gradient(circle at top left, #dbeafe, #f8fafc 55%); padding: 24px; }
.login-card { width: min(420px, 100%); background: #fff; border: 1px solid #e2e8f0; border-radius: 16px; padding: 32px; box-shadow: 0 20px 60px rgba(15,23,42,.08); }
.login-card h1 { margin: 8px 0; }
</style>
