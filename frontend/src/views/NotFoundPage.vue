<template>
  <div class="not-found-page">
    <a-result :status="null" class="not-found-result">
      <template #icon>
        <div class="code-404">404</div>
      </template>
      <template #title>
        <div class="eyebrow">Page Not Found</div>
      </template>
      <template #subtitle>
        The page you're looking for doesn't exist or has been moved.
      </template>
      <template #extra>
        <a-space direction="vertical" :size="24" align="center">
          <a-space :size="12" wrap>
            <a-button type="primary" @click="router.push('/')">Home</a-button>
            <a-button @click="router.push('/docs')">View Docs</a-button>
          </a-space>
          <a-space :size="24" wrap>
            <a-link class="mini-link" @click="router.push('/developer')">Developer Portal</a-link>
            <a-link class="mini-link" @click="router.push('/demo')">Live Demo</a-link>
          </a-space>
        </a-space>
      </template>
    </a-result>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'

const router = useRouter()
</script>

<style scoped>
.not-found-page {
  min-height: 100vh;
  background: #f7f6f3;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40px 24px;
}
.not-found-result { max-width: 480px; padding: 0; }
.code-404 {
  font-family: ui-monospace, Menlo, monospace;
  font-size: clamp(80px, 16vw, 128px);
  font-weight: 500;
  color: #e4e2dc;
  line-height: 1;
  letter-spacing: -.04em;
  user-select: none;
}
.not-found-result :deep(.arco-result-icon) { margin-bottom: 8px; }
.not-found-result :deep(.arco-result-title) { margin-bottom: 20px; }
.not-found-result :deep(.arco-result-subtitle) { font-size: 15px; line-height: 1.6; color: #6b6b70; }
.mini-link { font-family: ui-monospace, Menlo, monospace; font-size: 13px; color: #6b6b70; }
.mini-link:hover { color: #0b0b0d; }
</style>
