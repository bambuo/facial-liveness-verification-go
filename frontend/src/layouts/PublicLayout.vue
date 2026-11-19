<template>
  <a-layout class="public-shell">
    <a-layout-content>
      <div v-if="showBack" class="back-bar">
        <router-link to="/" class="back-link">← Back to developer portal</router-link>
      </div>
      <router-view />
    </a-layout-content>
    <a-layout-footer class="public-footer">
      <div class="footer-inner">
        <div class="footer-brand"><span class="footer-mark">IV</span><span>Identity Verification</span></div>
        <a-space wrap :size="18">
          <router-link to="/">Developer portal</router-link>
          <router-link to="/docs">Docs</router-link>
          <router-link to="/demo">Demo</router-link>
          <router-link to="/verify-credential">Verify credential</router-link>
          <a-link href="https://github.com" target="_blank">GitHub</a-link>
        </a-space>
      </div>
    </a-layout-footer>
  </a-layout>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
const route = useRoute()
const showBack = computed(() => !['/', '/developer', '/setup'].includes(route.path))
</script>

<style scoped>
.public-shell { min-height: 100vh; background: var(--paper, #f7f6f3); }
.back-bar { border-bottom: 1px solid var(--rule, #e4e2dc); padding: 10px 24px; }
.back-link { font-family: ui-monospace, Menlo, monospace; color: var(--mid, #6b6b70); font-size: 12.5px; text-decoration: none; }
.back-link:hover { color: var(--ink, #0b0b0d); }
.public-footer { background: var(--panel, #fff); border-top: 1px solid var(--rule, #e4e2dc); padding: 28px 24px; }
.footer-inner { max-width: 1280px; margin: 0 auto; display: flex; align-items: center; justify-content: space-between; gap: 18px; flex-wrap: wrap; }
.footer-brand { display: flex; align-items: center; gap: 10px; font-family: ui-monospace, Menlo, monospace; font-size: 12px; color: var(--mid, #6b6b70); }
.footer-mark { width: 24px; height: 24px; border-radius: 5px; background: var(--ink, #0b0b0d); color: var(--paper, #fff); display: grid; place-items: center; font-size: 10px; }
.footer-inner a { font-family: ui-monospace, Menlo, monospace; font-size: 12px; color: var(--mid, #6b6b70); text-decoration: none; }
.footer-inner a:hover { color: var(--ink, #0b0b0d); }
</style>
