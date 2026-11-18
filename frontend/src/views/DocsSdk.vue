<template><DocsShell eyebrow="SDK" title="Client SDKs" description="Thin clients around the REST contract; keep API keys server-side.">
  <div class="grid grid-3">
    <div class="panel"><h3>JavaScript</h3><p class="muted">Browser and Node client with API-key and session-token support.</p><pre class="code-block">const client = new VerificationClient({ apiKey })

const session = await client.initialize({ userId })
await client.uploadFront(session.id, file)</pre></div>
    <div class="panel"><h3>Python</h3><p class="muted">Server-side client for backend orchestration.</p><pre class="code-block">client = VerificationClient(api_key="ik_...")
verification = client.initialize(user_id=user.id)
client.upload_front(verification.id, "passport.jpg")</pre></div>
    <div class="panel"><h3>React</h3><p class="muted">Hosted-page components and verification status hooks.</p><pre class="code-block"><VerificationFlow
  sessionToken={sessionToken}
  onComplete={handleResult}
/></pre></div>
  </div>
  <a-alert type="warning" style="margin-top:18px">Never embed production API keys in browser bundles. Use session tokens for end-user flows.</a-alert>
</DocsShell>
</template>
<script setup lang="ts">
import DocsShell from '../components/DocsShell.vue'
</script>
