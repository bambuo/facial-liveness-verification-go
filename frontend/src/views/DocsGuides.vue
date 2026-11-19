<template>
  <DocsShell slug="docs/guides" :nav="NAV">
    <!-- INTEGRATION OPTIONS -->
    <DocsSectionAnchor id="integration" />
    <DocsH2 index="01">Integration Options</DocsH2>
    <DocsLead>Three ways to add identity verification to your product — from zero-code to full control.</DocsLead>

    <a-card :bordered="true" class="docs-hosted-card">
      <div class="docs-hosted-title">Hosted Verification Page</div>
      <p class="docs-hosted-desc">
        Identity Verification hosts a complete verification UI at the following URL. This is the page used by all three
        integration methods below — redirect, iframe, or SDK embed.
      </p>
      <code class="docs-hosted-url">{{ siteUrl }}/user-verification</code>
    </a-card>

    <a-row :gutter="[16, 16]" class="docs-option-row">
      <a-col v-for="option in integrationOptions" :key="option.title" :xs="24" :md="8">
        <a-card :bordered="true" class="docs-mode-card">
          <div class="docs-mode-head">
            <span class="docs-option-emoji">{{ option.emoji }}</span>
            <span class="docs-option-title">{{ option.title }}</span>
            <a-tag size="small" :color="option.tagColor">{{ option.tag }}</a-tag>
          </div>
          <p class="docs-mode-desc">{{ option.desc }}</p>
          <div class="docs-mode-signals">
            <div v-for="feature in option.features" :key="feature" class="docs-mode-signal">
              <span class="docs-check-mark">✓</span>{{ feature }}
            </div>
          </div>
        </a-card>
      </a-col>
    </a-row>

    <p class="docs-body-text">
      All three methods use the same hosted page URL with query parameters. For fully custom UI, see the
      <router-link to="/docs" class="docs-inline-link">REST API</router-link> section.
    </p>

    <DocsCallout type="warning">
      <strong>Never put your API key in the browser.</strong> Initialize the verification session on your backend
      with <code>POST /api/v2/verify/initialize</code>, then pass the returned <code>verification_url</code> to
      the user. The URL contains a single-use <code>session</code> token scoped to one verification and valid for
      one hour — exposing it is bounded; exposing your <code>ik_*</code> key would let anyone create or bypass
      verifications under your account.
    </DocsCallout>

    <DocsCode
      label="Step 1: Initialize on your backend (server-side, with the API key)"
      :code="step1Code"
    />

    <a-card class="docs-table-card" :bordered="true">
      <div class="docs-info-label">Optional query parameters (append to verification_url)</div>
      <div class="docs-field-list">
        <DocsFieldRow name="theme" type="'light' | 'dark'" desc="UI color theme. Defaults to dark." />
        <DocsFieldRow
          name="address_verif"
          type="'true'"
          desc="When set to 'true', adds an optional proof-of-address step after identity verification. Users can upload a utility bill, bank statement, or tax document. The name is cross-referenced against the verified ID."
        />
      </div>
    </a-card>

    <DocsCode label="Step 2 — Option 1: Redirect (link or window.location)" :code="redirectCode" />

    <a-card class="docs-table-card" :bordered="true">
      <div class="docs-info-label">Redirect callback parameters</div>
      <p class="docs-callback-desc">
        When verification completes, the user is redirected back to your
        <code class="docs-inline-code">redirect_url</code> with these query parameters appended:
      </p>
      <div class="docs-field-list">
        <DocsFieldRow name="verification_id" type="UUID string" :req="true" desc="The verification session ID. Use this to fetch full results via GET /api/v2/verify/:id/status." />
        <DocsFieldRow name="status" type="string" :req="true" desc="Result status: 'verified', 'failed', or 'manual_review'. If manual_review, poll the status endpoint or listen for a webhook to get the final decision." />
        <DocsFieldRow name="user_id" type="string" :req="true" desc="The user_id you passed when starting verification." />
      </div>
    </a-card>

    <DocsCode label="Example: handling the redirect callback" :code="callbackCode" />

    <DocsCallout type="note">
      <strong>Extracted ID fields are in <code>ocr_data</code></strong> on the status response —
      <code>full_name</code>, <code>date_of_birth</code>, <code>document_number</code>, <code>nationality</code>,
      <code>address</code>, and more, with confidence scores. Available as soon as the front document is
      processed, regardless of whether the user has uploaded the back document or completed liveness. Use it to
      cross-check against user-supplied form data (e.g. ensure entered name matches OCR).
    </DocsCallout>

    <DocsCode label="Step 2 — Option 2: Iframe embed (HTML)" :code="iframeCode" />
    <DocsCode label="Step 2 — Option 3: SDK embed (see SDK &amp; Embed docs)" :code="sdkCode" />

    <DocsDivider />

    <!-- GUIDES & TUTORIALS -->
    <DocsSectionAnchor id="guides" />
    <DocsH2 index="02">Guides &amp; Tutorials</DocsH2>
    <DocsLead>
      Step-by-step walkthroughs for common integration patterns. Each guide includes production-ready code with
      proper error handling.
    </DocsLead>

    <!-- Guide 1: End-to-End Tutorial -->
    <DocsSectionAnchor id="guide-e2e" />
    <h3 class="docs-h3"><span class="docs-eyebrow" style="margin: 0">--</span> End-to-End Verification Tutorial</h3>
    <p class="docs-body-text">
      A complete walkthrough with proper error handling, exponential backoff polling, and all 5 verification
      steps. Unlike the Quick Start (which is minimal), this guide shows production-ready patterns.
    </p>

    <DocsCodeTabs v-model="tab" :js="e2eJs" :python="e2ePython" />

    <DocsCallout type="tip">
      For event-driven updates without manual polling, use the SDK's <code>watch()</code> method — see the
      <router-link to="/docs/sdk" class="docs-inline-link">JavaScript SDK</router-link> section.
    </DocsCallout>
    <DocsCallout type="warning">
      <strong>Liveness detection is mandatory.</strong> The live capture endpoint runs Gate 4 (liveness) before
      Gate 5 (face match). We recommend the <strong>head turn</strong> mode — it requires no client-side ML
      libraries, just <code>getUserMedia()</code> + canvas frame capture. Omitting
      <code>liveness_metadata</code> falls back to passive mode (lowest pass rate). Check
      <code>liveness_results.liveness_mode</code> in the response to confirm which mode was used. See
      <router-link to="/docs#step-4" class="docs-inline-link">Step 4</router-link> for details and metadata
      shapes.
    </DocsCallout>

    <!-- Guide 2: Mobile Handoff -->
    <DocsSectionAnchor id="guide-mobile" />
    <h3 class="docs-h3"><span class="docs-eyebrow" style="margin: 0">--</span> Mobile Handoff</h3>
    <p class="docs-body-text">
      Let users start verification on desktop and continue on their phone. Your backend creates a handoff session
      with a short-lived token, which you present as a QR code or deep link. The mobile browser opens the hosted
      verification page and camera access works natively.
    </p>

    <a-card class="docs-table-card" :bordered="true">
      <div class="docs-info-label">How handoff works</div>
      <div class="docs-numbered-list">
        <template v-for="row in handoffSteps" :key="row[0]">
          <span class="docs-list-number">{{ row[0] }}.</span>
          <span class="docs-list-text">{{ row[1] }}</span>
        </template>
      </div>
    </a-card>

    <DocsCodeTabs v-model="tab" :curl="handoffCurl" :js="handoffJs" :python="handoffPython" />

    <DocsCode label="Response — handoff/create (HTTP 201)" :code="handoffResponse" />

    <a-card class="docs-table-card" :bordered="true">
      <div class="docs-info-label">Best practices</div>
      <div class="docs-ticks">
        <div v-for="tip in handoffTips" :key="tip" class="docs-tick"><span class="docs-check-mark">✓</span>{{ tip }}</div>
      </div>
    </a-card>

    <!-- Guide 3: Building Custom UI -->
    <DocsSectionAnchor id="guide-custom-ui" />
    <h3 class="docs-h3"><span class="docs-eyebrow" style="margin: 0">--</span> Building a Custom Verification UI</h3>
    <p class="docs-body-text">
      Patterns and best practices for building your own verification UI on top of the REST API.
    </p>

    <h4 class="docs-h4">Polling Strategies</h4>
    <a-row :gutter="[12, 12]" class="docs-option-row">
      <a-col v-for="strategy in pollingStrategies" :key="strategy.title" :xs="24" :md="8">
        <a-card :bordered="true" class="docs-strategy-card">
          <div class="docs-strategy-title" :style="{ color: strategy.color }">{{ strategy.title }}</div>
          <div class="docs-mode-signal"><span class="docs-plus">+</span>{{ strategy.pros }}</div>
          <div class="docs-mode-signal"><span class="docs-minus">-</span>{{ strategy.cons }}</div>
          <code class="docs-strategy-code">{{ strategy.code }}</code>
        </a-card>
      </a-col>
    </a-row>

    <h4 class="docs-h4">Progress Indicators</h4>
    <p class="docs-body-text">
      Map the <code class="docs-inline-code">status</code> field to user-friendly step labels:
    </p>
    <a-card class="docs-table-card" :bordered="true">
      <div v-for="row in progressRows" :key="row.status" class="docs-progress-row">
        <code class="docs-table-code">{{ row.status }}</code>
        <span class="docs-progress-label">{{ row.label }}</span>
        <code class="docs-progress-step">{{ row.step }}</code>
      </div>
    </a-card>

    <h4 class="docs-h4">Error Recovery</h4>
    <a-card class="docs-table-card" :bordered="true">
      <div v-for="row in errorRows" :key="row.code" class="docs-error-row">
        <code class="docs-error-code" :style="{ color: row.color }">{{ row.code }}</code>
        <span class="docs-error-action">{{ row.action }}</span>
      </div>
    </a-card>

    <h4 class="docs-h4">Displaying OCR Results</h4>
    <DocsCode label="Parse OCR data from status response" :code="ocrCode" />

    <DocsCallout type="tip">
      For teams that want a pre-built UI, use the
      <router-link to="/docs/sdk#embed" class="docs-inline-link">Embed Component</router-link> — it provides a
      complete verification experience you can drop into your app with zero frontend code.
    </DocsCallout>

    <DocsDivider />

    <!-- SELF-HOSTING GUIDE -->
    <DocsSectionAnchor id="self-hosting" />
    <DocsH2 index="03">Self-Hosting Guide</DocsH2>
    <DocsLead>
      Deploy Identity Verification Community Edition on your own server with Docker Compose. Full control over your data and
      infrastructure.
    </DocsLead>

    <DocsSectionAnchor id="sh-prerequisites" />
    <h3 class="docs-h3"><span class="docs-eyebrow" style="margin: 0">--</span> Prerequisites</h3>
    <p class="docs-body-text">
      A clean Linux server (Debian 12+, Ubuntu 22.04+) with root access and a domain name pointed to the
      server's IP address.
    </p>

    <a-card class="docs-table-card" :bordered="true">
      <div class="docs-info-label">Required software</div>
      <div v-for="dep in requiredSoftware" :key="dep.name" class="docs-dep-row">
        <code class="docs-dep-name">{{ dep.name }}</code>
        <span class="docs-dep-desc">{{ dep.desc }}</span>
      </div>
    </a-card>

    <DocsCode label="Install dependencies on Debian / Ubuntu" :code="installDepsCode" />

    <DocsSectionAnchor id="sh-install" />
    <h3 class="docs-h3"><span class="docs-eyebrow" style="margin: 0">--</span> Install</h3>
    <p class="docs-body-text">Three install options, from quickest to most control.</p>

    <a-row :gutter="[16, 16]" class="docs-option-row">
      <a-col v-for="option in installOptions" :key="option.title" :xs="24" :md="8">
        <a-card :bordered="true" class="docs-mode-card">
          <div class="docs-mode-head">
            <span class="docs-option-emoji">{{ option.emoji }}</span>
            <span class="docs-option-title">{{ option.title }}</span>
            <a-tag size="small" :color="option.tagColor">{{ option.tag }}</a-tag>
          </div>
          <p class="docs-mode-desc">{{ option.desc }}</p>
        </a-card>
      </a-col>
    </a-row>

    <DocsCode label="Option 1: One-line install" :code="installOneLiner" />
    <DocsCode label="Option 2: Clone &amp; install (recommended)" :code="installClone" />
    <DocsCode label="Option 3: Build from source" :code="installSource" />

    <p class="docs-body-text">The installer will interactively prompt for:</p>
    <a-card class="docs-table-card" :bordered="true">
      <div class="docs-ticks">
        <div v-for="item in installerPrompts" :key="item" class="docs-tick"><span class="docs-diamond">◆</span>{{ item }}</div>
      </div>
    </a-card>

    <DocsCallout type="tip">
      After installation, verify all containers are healthy with <code>docker ps</code> and check the API with
      <code>curl -s http://localhost:3001/health | jq</code>.
    </DocsCallout>

    <DocsSectionAnchor id="sh-external-db" />
    <h3 class="docs-h3"><span class="docs-eyebrow" style="margin: 0">--</span> External Database (BYOD)</h3>
    <p class="docs-body-text">
      If using an external database provider (Railway, Supabase, RDS, etc.), keep these in mind:
    </p>

    <a-card class="docs-table-card" :bordered="true">
      <div v-for="row in externalDbIssues" :key="row.issue" class="docs-error-row">
        <code class="docs-issue-code">{{ row.issue }}</code>
        <span class="docs-error-action">{{ row.fix }}</span>
      </div>
    </a-card>

    <DocsCode label=".env — external database configuration" :code="envCode" />

    <DocsCallout type="warning">
      After changing <code>.env</code>, you must recreate the container — not just restart it. Use
      <code>docker compose up -d --force-recreate api</code> to pick up new environment variables.
    </DocsCallout>

    <DocsSectionAnchor id="sh-commands" />
    <h3 class="docs-h3"><span class="docs-eyebrow" style="margin: 0">--</span> Useful Commands</h3>

    <a-card class="docs-table-card" :bordered="true">
      <div v-for="row in usefulCommands" :key="row.cmd" class="docs-error-row">
        <code class="docs-command-code">{{ row.cmd }}</code>
        <span class="docs-error-action">{{ row.desc }}</span>
      </div>
    </a-card>
  </DocsShell>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import DocsShell, {
  DocsSectionAnchor,
  DocsH2,
  DocsLead,
  DocsDivider,
  DocsCallout,
  DocsCode,
  DocsCodeTabs,
  DocsFieldRow,
  docsApiUrl,
  docsSiteUrl,
  type DocsNavItem,
} from '../components/DocsShell.vue'

const apiUrl = docsApiUrl()
const siteUrl = docsSiteUrl()
const tab = ref('curl')

const NAV: DocsNavItem[] = [
  { id: 'integration', label: 'Integration Options', depth: 0 },
  { id: 'guides', label: 'Guides & Tutorials', depth: 0 },
  { id: 'guide-e2e', label: 'End-to-End Tutorial', depth: 1 },
  { id: 'guide-mobile', label: 'Mobile Handoff', depth: 1 },
  { id: 'guide-custom-ui', label: 'Building Custom UI', depth: 1 },
  { id: 'self-hosting', label: 'Self-Hosting', depth: 0 },
  { id: 'sh-prerequisites', label: 'Prerequisites', depth: 1 },
  { id: 'sh-install', label: 'Install', depth: 1 },
  { id: 'sh-external-db', label: 'External Database', depth: 1 },
  { id: 'sh-commands', label: 'Useful Commands', depth: 1 },
]

const integrationOptions = [
  {
    emoji: '🔗',
    title: 'Redirect',
    tag: 'Easiest',
    tagColor: 'green',
    desc: 'Send users to the hosted page with a link or redirect. After verification, they are sent back to your redirect_url.',
    features: [
      'Zero frontend code',
      'Works on any platform',
      'User leaves your site temporarily',
      'Best for: server-rendered apps, email flows',
    ],
  },
  {
    emoji: '🖼️',
    title: 'Iframe Embed',
    tag: 'No SDK needed',
    tagColor: 'arcoblue',
    desc: 'Embed the hosted page inside your app using a standard HTML iframe. Users never leave your site.',
    features: ['One line of HTML', 'Users stay on your domain', 'Requires allow="camera"', 'Best for: quick inline embed'],
  },
  {
    emoji: '⚡',
    title: 'SDK Embed',
    tag: 'Recommended',
    tagColor: 'cyan',
    desc: 'Use the @identity-verification/sdk IdentityVerificationEmbed component for modal or inline mode with event callbacks.',
    features: [
      'Modal or inline mode',
      'onComplete / onError callbacks',
      'postMessage communication',
      'Best for: SPAs, React, Vue',
    ],
  },
]

const handoffSteps = [
  ['1', 'Your backend calls POST /api/verify/handoff/create with api_key and user_id in the request body'],
  ['2', 'API returns a token (valid for 10 minutes) and an expires_at timestamp'],
  ['3', 'Build a verification URL using the token and display it as a QR code or deep link'],
  ['4', 'User opens the URL on their phone — the hosted page handles all steps'],
  ['5', 'Poll GET /api/verify/handoff/:token/status from your backend until status is "completed"'],
]

const handoffTips = [
  'HTTPS is required for camera access on mobile browsers',
  "Test on real devices — emulators don't support camera APIs",
  'Use a QR library (qrcode, qrcode.react) rather than generating images manually',
  'Handle token expiry (10 min) — show a "Regenerate" button if the user is slow',
  'Set redirect_url to bring the user back to your app after verification',
]

const pollingStrategies = [
  { title: 'Fixed Interval', color: '#b45309', pros: 'Simple to implement', cons: 'Wastes requests if processing is slow', code: 'setInterval(() => poll(), 2000)' },
  { title: 'Exponential Backoff', color: '#15803d', pros: 'Efficient — reduces load over time', cons: 'Slightly slower on fast responses', code: 'delay = min(delay * 1.5, 10s)' },
  { title: 'SDK watch()', color: '#0e7490', pros: 'Event-driven, auto-cleanup', cons: 'Requires @identity-verification/sdk', code: "sdk.watch(id).on('done', cb)" },
]

const progressRows = [
  { status: 'AWAITING_FRONT', label: 'Upload your ID (front)', step: '1/4' },
  { status: 'AWAITING_BACK', label: 'Upload your ID (back)', step: '2/4' },
  { status: 'CROSS_VALIDATING', label: 'Verifying document...', step: '2/4' },
  { status: 'AWAITING_LIVE', label: 'Take a photo of yourself', step: '3/4' },
  { status: 'FACE_MATCHING', label: 'Verifying identity...', step: '3/4' },
  { status: 'COMPLETE', label: 'Verification complete!', step: '4/4' },
  { status: 'HARD_REJECTED', label: 'Verification failed', step: '—' },
]

const errorRows = [
  { code: '409 Conflict', action: 'Session was rejected by a gate. Start a new verification session.', color: '#c2410c' },
  { code: '429 Too Many Requests', action: 'Back off and retry after the Retry-After header value.', color: '#b45309' },
  { code: '400 Bad Request', action: 'Fix the input (wrong file format, missing fields) and retry the same step.', color: '#b45309' },
  { code: 'Network Error', action: 'Retry with exponential backoff (max 3 retries). The API is idempotent for uploads.', color: '#b91c1c' },
]

const requiredSoftware = [
  { name: 'Git', desc: 'For cloning the repository' },
  { name: 'Docker', desc: 'Container runtime (v20+)' },
  { name: 'Docker Compose', desc: 'Multi-container orchestration (V2, included with Docker)' },
  { name: 'curl', desc: 'For downloading the install script' },
]

const installOptions = [
  {
    emoji: '⚡',
    title: 'One-line Install',
    tag: 'Quickest',
    tagColor: 'green',
    desc: 'Downloads and runs the install script directly. Pulls pre-built Docker images from the registry.',
  },
  {
    emoji: '📋',
    title: 'Clone & Install',
    tag: 'Recommended',
    tagColor: 'cyan',
    desc: 'Clone the repo first so you can inspect the script before running it. Still pulls pre-built images.',
  },
  {
    emoji: '🔧',
    title: 'Build from Source',
    tag: 'Full control',
    tagColor: 'orange',
    desc: 'Clone the repo and build Docker images locally from source. Useful for custom modifications.',
  },
]

const installerPrompts = [
  'Domain name — for HTTPS via Caddy reverse proxy',
  'Database — bundled PostgreSQL or bring your own (external URL)',
  'Storage — local filesystem or S3-compatible object storage',
  'Auto-updates — optional Watchtower sidecar for automatic container updates',
]

const externalDbIssues = [
  {
    issue: 'Internal hostnames',
    fix: 'Use the public connection string, not internal/private hostnames (e.g., use mainline.proxy.rlwy.net instead of postgres.railway.internal)',
  },
  {
    issue: 'SSL certificate errors',
    fix: 'Set DATABASE_SSL_REJECT_UNAUTHORIZED=false in .env if your provider uses self-signed certificates',
  },
  {
    issue: 'Env changes not applied',
    fix: 'After editing .env, recreate containers with docker compose up -d --force-recreate api (restart alone reuses old env)',
  },
]

const usefulCommands = [
  { cmd: 'docker compose logs -f api', desc: 'Stream API logs in real-time' },
  { cmd: 'docker compose down && docker compose up -d', desc: 'Restart the entire stack' },
  { cmd: 'docker compose up -d --force-recreate api', desc: 'Recreate API container (after .env changes)' },
  { cmd: 'bash update.sh', desc: 'Pull latest images and update' },
  { cmd: 'bash uninstall.sh', desc: 'Remove Identity Verification and clean up' },
]

const step1Code = `// Server-side — your API key never touches the browser
const init = await fetch('${apiUrl}/api/v2/verify/initialize', {
  method: 'POST',
  headers: {
    'X-API-Key': 'ik_your_api_key',
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    user_id: 'user-123',
    redirect_url: 'https://yourapp.com/done',
  }),
});
const { verification_id, verification_url, session_token } = await init.json();
// verification_url looks like:
//   ${siteUrl}/user-verification?session=<64-char-hex-token>`

const redirectCode = `// Pass verification_url to your frontend, then redirect.
// The ?session= token in the URL scopes this verification — no API key needed.
window.location.href = verification_url + '&theme=dark';`

const callbackCode = `// On your redirect_url page (e.g. https://yourapp.com/done)
const params = new URLSearchParams(window.location.search);
const verificationId = params.get('verification_id');
const status = params.get('status');   // 'verified', 'failed', or 'manual_review'
const userId = params.get('user_id');

if (status === 'verified' && verificationId) {
  // Fetch full results from your backend — includes ocr_data, cross_validation_results,
  // face_match_results, liveness_results, risk_score, age_verification, and more.
  // See "Polling Status" in the End-to-End Tutorial below for the full response shape.
  const res = await fetch('/api/verification-result?id=' + encodeURIComponent(verificationId));
  const data = await res.json();
  // data.ocr_data.full_name, data.ocr_data.date_of_birth, etc.
  // Update your user record, grant access, etc.
} else if (status === 'manual_review') {
  // Verification needs human review — poll GET /api/v2/verify/:id/status
  // or wait for a webhook to get the final decision
}`

const iframeCode = `<!-- Render verification_url from your backend into the iframe src.
     The ?session= token in the URL scopes this verification — never put
     your API key in an iframe src, it would be visible in DevTools. -->
<iframe
  src="<%= verification_url %>&theme=dark"
  width="100%" height="700" frameborder="0"
  allow="camera; microphone"
  style="border: none; border-radius: 8px;"
></iframe>`

const sdkCode = `import { IdentityVerificationEmbed } from '@identity-verification/sdk';

const embed = new IdentityVerificationEmbed({ mode: 'modal', theme: 'dark' });
embed.open(sessionToken, {
  onComplete: (result) => console.log(result.finalResult),
  onError: (error) => console.error(error.message),
});`
const e2eJs = `// Reusable polling helper with exponential backoff
async function pollStatus(verificationId, apiKey, baseUrl, {
  condition, // (data) => boolean — stop polling when true
  maxAttempts = 60,
  initialDelay = 1000,
  maxDelay = 10000,
} = {}) {
  let delay = initialDelay;
  for (let i = 0; i < maxAttempts; i++) {
    await new Promise(ok => setTimeout(ok, delay));
    const res = await fetch(
      \`\${baseUrl}/api/v2/verify/\${verificationId}/status\`,
      { headers: { 'X-API-Key': apiKey } },
    );
    if (!res.ok) throw new Error(\`Poll failed: \${res.status}\`);
    const data = await res.json();
    if (data.final_result === 'failed') {
      throw new Error(\`Verification failed: \${data.failure_reason}\`);
    }
    if (condition(data)) return data;
    delay = Math.min(delay * 1.5, maxDelay); // exponential backoff
  }
  throw new Error('Polling timed out');
}

// Full verification flow
const BASE = '${apiUrl}';
const KEY  = 'ik_your_api_key';
const headers = { 'X-API-Key': KEY };

// Step 1: Initialize
const init = await fetch(\`\${BASE}/api/v2/verify/initialize\`, {
  method: 'POST',
  headers: { ...headers, 'Content-Type': 'application/json' },
  body: JSON.stringify({ user_id: 'user-uuid' }),
}).then(r => r.json());
const vid = init.verification_id;

// Step 2: Upload front document
const fd1 = new FormData();
fd1.append('document_type', 'drivers_license');
fd1.append('document', frontFile);
const frontRes = await fetch(\`\${BASE}/api/v2/verify/\${vid}/front-document\`, {
  method: 'POST', headers, body: fd1,
}).then(r => r.json());
if (!frontRes.success) throw new Error(frontRes.message);

// Poll until OCR is ready
const ocrResult = await pollStatus(vid, KEY, BASE, {
  condition: (d) => d.ocr_data != null,
});
console.log('OCR:', ocrResult.ocr_data.full_name);

// Step 3: Upload back document (cross-validation auto-triggers)
const fd2 = new FormData();
fd2.append('document_type', 'drivers_license');
fd2.append('document', backFile);
await fetch(\`\${BASE}/api/v2/verify/\${vid}/back-document\`, {
  method: 'POST', headers, body: fd2,
}).then(r => r.json());

// Poll until cross-validation completes
const crossResult = await pollStatus(vid, KEY, BASE, {
  condition: (d) => d.cross_validation_results != null,
});
console.log('Cross-val:', crossResult.cross_validation_results.verdict);

// Step 4: Submit live capture with liveness data
// Head turn challenge — no client ML needed, just camera frames.
// Capture frames from getUserMedia() via canvas at different time points.
const fd3 = new FormData();
fd3.append('selfie', captureBlob, 'capture.jpg');
fd3.append('liveness_metadata', JSON.stringify({
  challenge_type: 'head_turn',
  challenge_direction: 'left',
  frames: capturedFrames, // array from canvas.toDataURL('image/jpeg')
  // Each frame: { frame_base64, timestamp, phase }
  start_timestamp: startTime,
  end_timestamp: Date.now(),
}));
const liveRes = await fetch(\`\${BASE}/api/v2/verify/\${vid}/live-capture\`, {
  method: 'POST', headers, body: fd3,
}).then(r => r.json());

// Verify which liveness mode was actually used
console.log('Liveness mode:', liveRes.liveness_results?.liveness_mode);

// Poll for final result
const final = await pollStatus(vid, KEY, BASE, {
  condition: (d) => d.final_result != null,
});
console.log('Result:', final.final_result);
console.log('Face match:', final.face_match_results?.similarity_score);
console.log('Liveness:', final.liveness_results?.liveness_passed);`
const e2ePython = `import requests, time

BASE = '${apiUrl}'
KEY  = 'ik_your_api_key'
HEADERS = {'X-API-Key': KEY}

# Reusable polling helper with exponential backoff
def poll_status(vid, *, condition, max_attempts=60,
                initial_delay=1.0, max_delay=10.0):
    delay = initial_delay
    for _ in range(max_attempts):
        time.sleep(delay)
        r = requests.get(
            f'{BASE}/api/v2/verify/{vid}/status', headers=HEADERS
        ).json()
        if r.get('final_result') == 'failed':
            raise Exception(f"Verification failed: {r.get('failure_reason')}")
        if condition(r):
            return r
        delay = min(delay * 1.5, max_delay)
    raise TimeoutError('Polling timed out')

# Full verification flow

# Step 1: Initialize
init = requests.post(f'{BASE}/api/v2/verify/initialize',
    headers={**HEADERS, 'Content-Type': 'application/json'},
    json={'user_id': 'user-uuid'},
).json()
vid = init['verification_id']

# Step 2: Upload front document
with open('front.jpg', 'rb') as f:
    requests.post(f'{BASE}/api/v2/verify/{vid}/front-document',
        headers=HEADERS,
        data={'document_type': 'drivers_license'},
        files={'document': f},
    )

# Poll until OCR is ready
ocr = poll_status(vid, condition=lambda d: d.get('ocr_data'))
print('OCR:', ocr['ocr_data']['full_name'])

# Step 3: Upload back document
with open('back.jpg', 'rb') as f:
    requests.post(f'{BASE}/api/v2/verify/{vid}/back-document',
        headers=HEADERS,
        data={'document_type': 'drivers_license'},
        files={'document': f},
    )

# Poll until cross-validation completes
cross = poll_status(vid, condition=lambda d: d.get('cross_validation_results'))
print('Cross-val:', cross['cross_validation_results']['verdict'])

# Step 4: Submit live capture with liveness data
# Head turn challenge — no client ML needed, just camera frames.
# captured_frames: list of dicts from canvas.toDataURL() on the frontend.
import json

liveness_metadata = {
    'challenge_type': 'head_turn',
    'challenge_direction': 'left',
    'frames': captured_frames,  # list of {frame_base64, timestamp, phase}
    'start_timestamp': start_time,
    'end_timestamp': end_time,
}

with open('capture.jpg', 'rb') as f:
    live_res = requests.post(f'{BASE}/api/v2/verify/{vid}/live-capture',
        headers=HEADERS,
        files={'selfie': f},
        data={'liveness_metadata': json.dumps(liveness_metadata)},
    ).json()

# Verify which liveness mode was actually used
print('Liveness mode:', live_res.get('liveness_results', {}).get('liveness_mode'))

# Poll for final result
final = poll_status(vid, condition=lambda d: d.get('final_result'))
print('Result:', final['final_result'])
print('Face match:', final.get('face_match_results', {}).get('similarity_score'))
print('Liveness:', final.get('liveness_results', {}).get('liveness_passed'))`

const handoffCurl = `# Create handoff session (api_key in body, not header)
curl -X POST ${apiUrl}/api/verify/handoff/create \\
  -H "Content-Type: application/json" \\
  -d '{
    "api_key": "ik_your_api_key",
    "user_id": "user-uuid"
  }'

# Poll status using the token
curl -X GET ${apiUrl}/api/verify/handoff/TOKEN/status`

const handoffJs = `// Create handoff session from your backend
const handoff = await fetch(\`${apiUrl}/api/verify/handoff/create\`, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    api_key: 'ik_your_api_key',
    user_id: 'user-uuid',
  }),
}).then(r => r.json());

// Build the verification URL with the token
const verifyUrl = \`${siteUrl}/user-verification?token=\${handoff.token}\`;

// Display QR code (use any QR library: qrcode, qrcode.react, etc.)
showQrCode(verifyUrl);

// Poll for completion using the token
const pollHandoff = setInterval(async () => {
  const res = await fetch(
    \`${apiUrl}/api/verify/handoff/\${handoff.token}/status\`,
  );
  if (res.status === 410) {
    clearInterval(pollHandoff);
    console.log('Session expired — regenerate token');
    return;
  }
  const status = await res.json();
  if (status.status === 'completed') {
    clearInterval(pollHandoff);
    console.log('Verification complete:', status.result);
  }
}, 3000);`

const handoffPython = `import requests, time

BASE = '${apiUrl}'

# Create handoff session (api_key in body, not header)
handoff = requests.post(f'{BASE}/api/verify/handoff/create',
    json={
        'api_key': 'ik_your_api_key',
        'user_id': 'user-uuid',
    },
).json()

# Build the verification URL with the token
verify_url = f'${siteUrl}/user-verification?token={handoff["token"]}'
print('Scan this URL:', verify_url)

# Poll for completion using the token
while True:
    time.sleep(3)
    res = requests.get(f'{BASE}/api/verify/handoff/{handoff["token"]}/status')
    if res.status_code == 410:
        print('Session expired — regenerate token')
        break
    status = res.json()
    if status['status'] == 'completed':
        print('Done! Result:', status.get('result'))
        break`

const handoffResponse = `{
  "token": "a1b2c3d4e5f6...64-char-hex-token",
  "expires_at": "2026-03-20T10:10:00Z"
}`
const ocrCode = `const status = await fetch(
  \`${apiUrl}/api/v2/verify/\${vid}/status\`,
  { headers: { 'X-API-Key': 'your-key' } },
).then(r => r.json());

if (status.ocr_data) {
  const { full_name, date_of_birth, id_number, confidence_scores } = status.ocr_data;

  // Highlight low-confidence fields for manual review
  const LOW_CONFIDENCE = 0.8;
  Object.entries(confidence_scores).forEach(([field, score]) => {
    if (score < LOW_CONFIDENCE) {
      console.warn(\`Low confidence for \${field}: \${score}\`);
      // Show a yellow highlight or "please verify" prompt in your UI
    }
  });
}`

const installDepsCode = `apt-get update && apt-get install -y git curl

# Install Docker
curl -fsSL https://get.docker.com | sh

# Verify
docker --version
docker compose version`

const installOneLiner = `curl -fsSL https://raw.githubusercontent.com/team-identity/identity-community/main/install.sh | bash`

const installClone = `git clone https://github.com/team-identity/identity-community.git
cd identity-community
bash install.sh`

const installSource = `git clone https://github.com/team-identity/identity-community.git
cd identity-community
bash install.sh --build`

const envCode = `DATABASE_URL=postgresql://user:password@your-host.example.com:5432/dbname
DATABASE_SSL=true
DATABASE_SSL_REJECT_UNAUTHORIZED=false`
</script>

<style>
.docs-hosted-card {
  margin-bottom: 24px;
  border-left: 3px solid #0e7490 !important;
}
.docs-hosted-card .arco-card-body { padding: 16px 20px; }
.docs-hosted-title { font-size: 13.5px; font-weight: 600; margin-bottom: 6px; }
.docs-hosted-desc { margin: 0 0 8px; font-size: 12.5px; color: #64748b; line-height: 1.65; }
.docs-hosted-url {
  display: block;
  padding: 8px 12px;
  background: #f8fafc;
  border: 1px solid var(--rule, #e4e2dc);
  font-family: ui-monospace, Menlo, monospace;
  font-size: 12.5px;
  color: #0e7490;
}
.docs-option-row { margin-bottom: 20px; }
.docs-option-emoji { font-size: 20px; }
.docs-option-title { font-size: 14px; font-weight: 700; }
.docs-inline-link { color: #0e7490; font-family: ui-monospace, Menlo, monospace; font-size: 12.5px; }
.docs-callback-desc { margin: 0 0 10px; font-size: 12.5px; color: #64748b; line-height: 1.65; }
.docs-numbered-list { display: grid; grid-template-columns: auto 1fr; gap: 6px 16px; }
.docs-list-number { font-family: ui-monospace, Menlo, monospace; font-size: 12px; font-weight: 600; color: #0e7490; }
.docs-list-text { font-size: 12.5px; color: #475569; line-height: 1.6; }
.docs-ticks { display: flex; flex-direction: column; gap: 6px; }
.docs-tick { display: flex; gap: 8px; align-items: flex-start; font-size: 12.5px; color: #475569; line-height: 1.6; }
.docs-diamond { color: #0e7490; font-size: 10px; margin-top: 3px; }
.docs-strategy-card .arco-card-body { padding: 16px 18px; display: flex; flex-direction: column; gap: 4px; }
.docs-strategy-title { font-family: ui-monospace, Menlo, monospace; font-size: 12.5px; font-weight: 600; margin-bottom: 4px; }
.docs-plus { color: #15803d; }
.docs-minus { color: #b91c1c; }
.docs-strategy-code { margin-top: 6px; font-family: ui-monospace, Menlo, monospace; font-size: 11px; color: #94a3b8; }
.docs-progress-row { display: flex; gap: 12px; align-items: center; padding: 8px 0; border-top: 1px solid #eef2f7; }
.docs-progress-row:first-child { border-top: none; }
.docs-progress-row .docs-table-code { width: 170px; flex-shrink: 0; }
.docs-progress-label { flex: 1; font-size: 13px; color: #0f172a; }
.docs-progress-step { font-family: ui-monospace, Menlo, monospace; font-size: 11.5px; color: #94a3b8; }
.docs-error-row { display: flex; gap: 12px; align-items: flex-start; padding: 10px 0; border-top: 1px solid #eef2f7; }
.docs-error-row:first-child { border-top: none; }
.docs-error-code { font-family: ui-monospace, Menlo, monospace; font-size: 12px; width: 180px; flex-shrink: 0; }
.docs-issue-code { font-family: ui-monospace, Menlo, monospace; font-size: 12px; width: 190px; flex-shrink: 0; color: #b45309; }
.docs-command-code { font-family: ui-monospace, Menlo, monospace; font-size: 12px; width: 380px; flex-shrink: 0; color: #0e7490; }
.docs-error-action { font-size: 12.5px; color: #475569; line-height: 1.6; flex: 1; }
.docs-dep-row { display: flex; gap: 12px; align-items: center; padding: 8px 0; border-top: 1px solid #eef2f7; }
.docs-dep-row:first-child { border-top: none; }
.docs-dep-name { font-family: ui-monospace, Menlo, monospace; font-size: 12px; color: #0e7490; width: 150px; flex-shrink: 0; }
.docs-dep-desc { font-size: 12.5px; color: #475569; }
@media (max-width: 767px) {
  .docs-error-code,
  .docs-issue-code,
  .docs-command-code,
  .docs-dep-name { width: auto; }
  .docs-error-row,
  .docs-dep-row { flex-direction: column; gap: 4px; }
}
</style>
