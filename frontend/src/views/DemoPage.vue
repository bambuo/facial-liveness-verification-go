<template>
  <div class="page">
    <section class="hero">
      <div class="eyebrow">Live demo</div>
      <h1>Identity verification in minutes.</h1>
      <p>Initialize a verification, collect a document and selfie, then receive a deterministic result with review controls.</p>
      <a-space><a-button type="primary" size="large" @click="$router.push('/user-verification')">Start verification</a-button><a-button size="large" @click="$router.push('/docs/sdk')">View SDK</a-button></a-space>
    </section>
    <div class="grid grid-3">
      <div class="panel"><icon-file /><h3>1. Document</h3><p class="muted">Upload a passport, driving licence or national ID. OCR and MRZ extraction run server-side.</p></div>
      <div class="panel"><icon-camera /><h3>2. Selfie</h3><p class="muted">Capture a live selfie. Liveness, face match and deepfake checks are evaluated with real models.</p></div>
      <div class="panel"><icon-check-circle /><h3>3. Result</h3><p class="muted">Receive a deterministic status and final result; manual review remains available.</p></div>
    </div>
    <div class="panel" style="margin-top:20px">
      <div class="card-title"><h3 style="margin:0">API quick start</h3><a-tag>REST</a-tag></div>
      <pre class="code-block">POST /api/v2/verify/initialize
X-API-Key: ik_...
Content-Type: application/json

{"user_id":"00000000-0000-0000-0000-000000000000"}

→ 201 {"verification_id":"...","session_token":"...","status":"AWAITING_FRONT"}</pre>
    </div>
  </div>
</template>
<style scoped>
.hero { padding-top: 30px; }
.panel h3 { margin: 12px 0 6px; }
</style>
