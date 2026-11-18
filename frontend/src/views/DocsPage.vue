<template>
  <div class="page">
    <div class="eyebrow">Documentation</div><h1>Build with Identity Verification</h1>
    <p class="muted" style="max-width:760px;font-size:16px;line-height:1.7">Start with the API, follow a guide, or explore the SDK. Every endpoint uses the same response conventions and error model.</p>
    <div class="grid grid-3" style="margin-top:24px">
      <router-link v-for="item in items" :key="item.to" :to="item.to" class="doc-card"><icon-book /><h3>{{ item.title }}</h3><p class="muted">{{ item.text }}</p></router-link>
    </div>
  </div>
</template>
<script setup lang="ts">
const items = [
  { to: '/docs/guides', title: 'Guides', text: 'Initialize, upload documents, capture a selfie and finalize.' },
  { to: '/docs/sdk', title: 'SDK', text: 'JavaScript, Python and React SDK conventions.' },
  { to: '/docs/features', title: 'Features', text: 'OCR, liveness, face match, vault, VC and compliance.' },
  { to: '/docs/reference', title: 'API reference', text: '154 contract endpoints and authentication.' },
  { to: '/docs/markdown', title: 'Markdown docs', text: 'Author guides in Markdown and publish them here.' },
  { to: '/docs/review', title: 'Review dashboard', text: 'Manual review workflow and evidence.' },
]
</script>
<style scoped>
.doc-card { display: block; text-decoration: none; color: inherit; background: #fff; border: 1px solid #e2e8f0; border-radius: 14px; padding: 22px; }
.doc-card:hover { border-color: #93c5fd; transform: translateY(-1px); }
.doc-card h3 { margin: 12px 0 6px; }
</style>
