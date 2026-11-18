<template><DocsShell eyebrow="Reference" title="API reference" description="The contract is frozen at 154 registrations / 158 reachable paths. Authentication uses API keys, handoff tokens or session tokens.">
  <a-tabs default-active-key="auth">
    <a-tab-pane key="auth" title="Authentication"><pre class="code-block">X-API-Key: ik_...          # developer/server integrations
X-Handoff-Token: &lt;64 hex&gt;  # desktop → mobile handoff
X-Session-Token: &lt;64 hex&gt;  # hosted verification page
X-Service-Token: ...       # internal service calls
Authorization: Bearer ...  # admin / developer portal</pre></a-tab-pane>
    <a-tab-pane key="verify" title="Core verification"><pre class="code-block">POST /api/v2/verify/initialize
POST /api/v2/verify/{id}/front-document
POST /api/v2/verify/{id}/back-document
POST /api/v2/verify/{id}/live-capture
POST /api/v2/verify/{id}/finalize
GET  /api/v2/verify/{id}/status</pre></a-tab-pane>
    <a-tab-pane key="errors" title="Errors"><pre class="code-block">{"success":false,"error":"...","message":"..."}
HTTP 400/401/403/404/409/410/422/429/500/503</pre></a-tab-pane>
  </a-tabs>
</DocsShell>
</template>
<script setup lang="ts">
import DocsShell from '../components/DocsShell.vue'
</script>
