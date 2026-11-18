<template><DocsShell eyebrow="Features" title="What the platform does" description="A complete verification stack with real OCR, face and risk models.">
  <div class="grid grid-2">
    <div class="panel"><h3>Document intelligence</h3><p class="muted">PP-OCRv5 document OCR, MRZ TD1/TD2/TD3, QR/DataMatrix/PDF417, country field rules and cross-validation.</p></div>
    <div class="panel"><h3>Biometrics</h3><p class="muted">YuNet detection, SFace 1:1 match, passive liveness, head-turn actions, real EfficientNet deepfake classification.</p></div>
    <div class="panel"><h3>Risk and compliance</h3><p class="muted">Deterministic gates, manual review, audit export, configurable thresholds and compliance rulesets.</p></div>
    <div class="panel"><h3>Platform</h3><p class="muted">API keys, handoff tokens, webhooks, vault, verifiable credentials, batch processing and monitoring.</p></div>
  </div>
</DocsShell>
</template>
<script setup lang="ts">
import DocsShell from '../components/DocsShell.vue'
</script>
