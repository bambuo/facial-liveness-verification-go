<template>
  <a-layout class="admin-shell">
    <a-layout-sider class="sider" :width="232" breakpoint="lg" :collapsed-width="0">
      <div class="sider-brand"><span class="brand-mark">IV</span><span>Admin</span></div>
      <a-menu :selected-keys="selected" @menu-item-click="navigate">
        <a-menu-item key="/admin/verifications"><template #icon><icon-safe /></template>Verifications</a-menu-item>
        <a-menu-item key="/admin/developers"><template #icon><icon-user-group /></template>Developers</a-menu-item>
        <a-menu-item key="/admin/thresholds"><template #icon><icon-settings /></template>Thresholds</a-menu-item>
      </a-menu>
      <div class="sider-foot"><span>{{ auth.email || 'Administrator' }}</span><a-link @click="signOut">Sign out</a-link></div>
    </a-layout-sider>
    <a-layout>
      <a-layout-header class="topbar">
        <div><div class="topbar-title">{{ title }}</div><div class="muted topbar-sub">Operations console</div></div>
        <div class="toolbar"><a-tag color="green">Admin</a-tag><a-button size="small" @click="$router.push('/docs')">Docs</a-button></div>
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
const selected = computed(() => [route.path])
const title = computed(() => String(route.meta.title || 'Admin'))
function navigate(key: string) { router.push(key) }
function signOut() { auth.logout(); router.push('/admin/login') }
</script>

<style scoped>
.admin-shell { min-height: 100vh; }
.sider { background: #111827; color: #e5e7eb; min-height: 100vh; }
.sider-brand { height: 64px; display: flex; align-items: center; gap: 10px; padding: 0 20px; color: #fff; font-weight: 700; border-bottom: 1px solid #1f2937; }
.brand-mark { width: 28px; height: 28px; border-radius: 7px; background: #059669; display: grid; place-items: center; font-size: 11px; }
.sider :deep(.arco-menu), .sider :deep(.arco-menu-item), .sider :deep(.arco-menu-inner) { background: transparent; color: #d1d5db; }
.sider :deep(.arco-menu-item:hover) { background: #1f2937; color: #fff; }
.sider :deep(.arco-menu-item.arco-menu-selected) { background: #1f2937; color: #fff; }
.sider-foot { position: absolute; bottom: 18px; left: 0; right: 0; padding: 0 20px; display: flex; flex-direction: column; gap: 6px; color: #9ca3af; font-size: 12px; }
.topbar { height: 76px; background: #fff; border-bottom: 1px solid #e5e7eb; padding: 0 28px; display: flex; align-items: center; justify-content: space-between; }
.topbar-title { font-size: 18px; font-weight: 700; }
.topbar-sub { font-size: 12px; margin-top: 2px; }
.content { background: #f5f7fb; min-height: calc(100vh - 76px); }
</style>
