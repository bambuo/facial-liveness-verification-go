<template><DocsShell eyebrow="Guide" title="Verify an identity" description="A complete API flow from initialization to final result.">
  <a-steps :current="1" style="margin:20px 0">
    <a-step>Initialize</a-step><a-step>Front document</a-step><a-step>Back document</a-step><a-step>Live capture</a-step><a-step>Finalize</a-step>
  </a-steps>
  <pre class="code-block"># 1. Initialize
POST /api/v2/verify/initialize
X-API-Key: ik_...
{"user_id":"&lt;uuid&gt;"}

# 2. Upload front
POST /api/v2/verify/{id}/front-document
X-API-Key: ik_...
multipart: document=@passport.jpg

# 3. Upload back (if required)
POST /api/v2/verify/{id}/back-document
multipart: document=@back.jpg

# 4. Capture selfie
POST /api/v2/verify/{id}/live-capture
multipart: selfie=@selfie.jpg

# 5. Finalize
POST /api/v2/verify/{id}/finalize</pre>
  <a-alert type="info" style="margin-top:18px">All decisions are deterministic. Review `status`, `current_step` and `final_result` after each step.</a-alert>
</DocsShell>
</template>
<script setup lang="ts">
import DocsShell from '../components/DocsShell.vue'
</script>
