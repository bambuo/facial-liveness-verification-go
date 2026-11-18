<template><DocsShell eyebrow="Operations" title="Review dashboard" description="Manual review keeps edge cases auditable without weakening automated gates.">
  <div class="grid grid-2">
    <div class="panel"><h3>When review is required</h3><p class="muted">Low-confidence OCR, cross-validation review verdicts, face-match skipped reasons or policy overrides.</p></div>
    <div class="panel"><h3>Reviewer actions</h3><p class="muted">Approve, reject or keep in review with a reason. Every action is written to the audit trail.</p></div>
  </div>
  <a-button type="primary" style="margin-top:18px" @click="$router.push('/admin/verifications')">Open review queue</a-button>
</DocsShell>
</template>
<script setup lang="ts">
import DocsShell from '../components/DocsShell.vue'
</script>
