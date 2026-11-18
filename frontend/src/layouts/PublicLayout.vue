<template>
  <a-layout class="public-shell">
    <a-layout-header class="public-header">
      <div class="header-inner">
        <router-link to="/" class="brand"><span class="brand-mark">IV</span><span>Identity Verification</span></router-link>
        <nav class="public-nav">
          <router-link to="/demo">Demo</router-link>
          <router-link to="/docs">Docs</router-link>
          <router-link to="/legal">Legal</router-link>
        </nav>
        <div class="toolbar">
          <a-button type="text" @click="$router.push('/admin/login')">Admin</a-button>
          <a-button type="primary" @click="$router.push('/developer')">Developer portal</a-button>
        </div>
      </div>
    </a-layout-header>
    <a-layout-content><router-view /></a-layout-content>
    <a-layout-footer class="public-footer">
      <div class="footer-inner"><span>Identity Verification Platform</span><span class="muted">API-first · Self-hosted · Privacy by design</span></div>
    </a-layout-footer>
  </a-layout>
</template>

<style scoped>
.public-shell { min-height: 100vh; background: #f8fafc; }
.public-header { background: #fff; border-bottom: 1px solid #e2e8f0; height: 64px; line-height: 64px; padding: 0 24px; }
.header-inner { max-width: 1180px; margin: 0 auto; display: flex; align-items: center; gap: 28px; }
.brand { display: flex; align-items: center; gap: 10px; color: #0f172a; font-weight: 700; text-decoration: none; }
.brand-mark { width: 30px; height: 30px; border-radius: 8px; background: #0f172a; color: #fff; display: grid; place-items: center; font-size: 12px; }
.public-nav { display: flex; gap: 22px; flex: 1; }
.public-nav a { color: #475569; text-decoration: none; font-size: 14px; }
.public-nav a.router-link-active { color: #2563eb; font-weight: 600; }
.public-footer { background: #fff; border-top: 1px solid #e2e8f0; padding: 26px 24px; }
.footer-inner { max-width: 1180px; margin: 0 auto; display: flex; justify-content: space-between; gap: 12px; color: #64748b; font-size: 13px; }
@media (max-width: 760px) { .public-nav { display: none; } .header-inner { gap: 12px; } .footer-inner { flex-direction: column; } }
</style>
