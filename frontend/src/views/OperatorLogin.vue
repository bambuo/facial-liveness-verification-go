<template>
  <div class="page-narrow" style="padding-top:72px">
    <div class="panel">
      <div class="eyebrow">Service operator</div>
      <h1 style="margin:8px 0">Operator access</h1>
      <p class="muted">Service operators use an `isk_` key. Paste it below; it is kept only in this browser.</p>
      <a-form :model="{}" layout="vertical" style="margin-top:20px">
        <a-form-item label="Service API key"><a-input-password v-model="key" placeholder="isk_..." /></a-form-item>
        <a-button type="primary" @click="save">Continue</a-button>
      </a-form>
    </div>
  </div>
</template>
<script setup lang="ts">
import { ref } from 'vue'
import { Message } from '@arco-design/web-vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
const key = ref(''); const auth = useAuthStore(); const router = useRouter()
function save() { if (!key.value) { Message.warning('Enter a service key'); return }; auth.setAPIKey(key.value.trim()); Message.success('Service key saved'); router.push('/developer') }
</script>
