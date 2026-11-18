<template>
  <a-layout class="developer-shell">
    <a-layout-sider class="sider" :width="232" breakpoint="lg" :collapsed-width="0">
      <div class="sider-brand"><span class="brand-mark">IV</span><span>Developer</span></div>
      <a-menu :selected-keys="selected" @menu-item-click="navigate">
        <a-menu-item key="/developer"><template #icon><icon-dashboard /></template>Overview</a-menu-item>
        <a-menu-item key="/developer/page-builder"><template #icon><icon-layout /></template>Page builder</a-menu-item>
        <a-menu-item key="/docs"><template #icon><icon-book /></template>Documentation</a-menu-item>
      </a-menu>
      <div class="sider-foot"><span>{{ auth.email || 'Developer account' }}</span><a-link @click="signOut">Sign out</a-link></div>
    </a-layout-sider>
    <a-layout>
      <a-layout-header class="topbar">
        <div><div class="topbar-title">{{ title }}</div><div class="muted topbar-sub">Manage verifications, keys and integrations</div></div>
        <div class="toolbar"><a-tag color="arcoblue">API key auth</a-tag><a-button size="small" @click="$router.push('/demo')">Open demo</a-button></div>
      </a-layout-header>
      <a-layout-content class="content"><router-view /></a-layout-content>
    </a-layout>
  </a-layout>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const selected = computed(() => [route.path.startsWith('/developer/page-builder') ? '/developer/page-builder' : route.path.startsWith('/docs') ? '/docs' : '/developer'])
const title = computed(() => String(route.meta.title || 'Developer'))
function navigate(key: string) { router.push(key) }
function signOut() { auth.logout(); router.push('/') }
</script>

<style scoped>
.developer-shell { min-height: 100vh; }
.sider { background: #0f172a; color: #e2e8f0; min-height: 100vh; }
.sider-brand { height: 64px; display: flex; align-items: center; gap: 10px; padding: 0 20px; color: #fff; font-weight: 700; border-bottom: 1px solid #1e293b; }
.brand-mark { width: 28px; height: 28px; border-radius: 7px; background: #2563eb; display: grid; place-items: center; font-size: 11px; }
.sider :deep(.arco-menu), .sider :deep(.arco-menu-item), .sider :deep(.arco-menu-inner) { background: transparent; color: #cbd5e1; }
.sider :deep(.arco-menu-item:hover) { background: #1e293b; color: #fff; }
.sider :deep(.arco-menu-item.arco-menu-selected) { background: #1e293b; color: #fff; }
.sider-foot { position: absolute; bottom: 18px; left: 0; right: 0; padding: 0 20px; display: flex; flex-direction: column; gap: 6px; color: #94a3b8; font-size: 12px; }
.topbar { height: 76px; background: #fff; border-bottom: 1px solid #e2e8f0; padding: 0 28px; display: flex; align-items: center; justify-content: space-between; }
.topbar-title { font-size: 18px; font-weight: 700; }
.topbar-sub { font-size: 12px; margin-top: 2px; }
.content { background: #f5f7fb; min-height: calc(100vh - 76px); }
</style>
