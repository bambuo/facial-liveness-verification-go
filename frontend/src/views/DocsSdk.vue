<template>
  <DocsShell slug="docs/sdk" :nav="NAV">
    <!-- JAVASCRIPT SDK -->
    <DocsSectionAnchor id="sdk" />
    <DocsH2 index="01">JavaScript SDK</DocsH2>
    <DocsLead>
      The official <code class="docs-inline-code">@identity-verification/sdk</code> wraps the REST API with TypeScript
      types, automatic error handling, and a real-time event emitter for tracking verification progress
      without manual polling.
    </DocsLead>

    <DocsCode label="Install" :code="installCode" />
    <DocsCode label="Initialize" :code="initCode" />

    <a-card class="docs-table-card" :bordered="true">
      <div class="docs-info-label">SDK Methods</div>
      <a-table
        :data="methods"
        :columns="methodColumns"
        :pagination="false"
        size="small"
        :bordered="false"
        row-key="method"
      >
        <template #method="{ record }"><code class="docs-table-code">{{ record.method }}</code></template>
        <template #returns="{ record }"><code class="docs-table-return">→ {{ record.returns }}</code></template>
      </a-table>
    </a-card>

    <h3 class="docs-h3">
      <span class="docs-eyebrow" style="margin: 0">--</span>
      Real-Time Events with <code class="docs-inline-code">watch()</code>
    </h3>
    <p class="docs-body-text">
      Instead of writing polling loops, use <code class="docs-inline-code">watch()</code> to get event-driven
      status updates. Polls at 2-second intervals by default, auto-stops on terminal states.
    </p>

    <DocsCode label="watch() usage" :code="watchCode" />

    <a-card class="docs-table-card" :bordered="true">
      <div class="docs-info-label">Event Types</div>
      <div class="docs-field-list">
        <DocsFieldRow name="status_changed" type="event" desc="Fires on any status transition (e.g. AWAITING_FRONT → AWAITING_BACK)" />
        <DocsFieldRow name="step_completed" type="event" desc="Fires when current_step increments (step 1→2, 2→3, etc.)" />
        <DocsFieldRow name="verification_complete" type="event" desc="Terminal success — status reached COMPLETE" />
        <DocsFieldRow name="verification_failed" type="event" desc="Terminal failure — status reached HARD_REJECTED" />
        <DocsFieldRow name="error" type="event" desc="Polling or network error occurred" />
      </div>
    </a-card>

    <a-card class="docs-table-card" :bordered="true">
      <div class="docs-info-label">Watch Options</div>
      <div class="docs-field-list">
        <DocsFieldRow name="interval" type="number" desc="Polling interval in ms (default: 2000)" />
        <DocsFieldRow name="maxAttempts" type="number" desc="Max poll attempts before auto-stop (default: 300 = 10 min)" />
      </div>
    </a-card>

    <DocsDivider />

    <!-- EMBED COMPONENT -->
    <DocsSectionAnchor id="embed" />
    <DocsH2 index="02">Embed Component</DocsH2>
    <DocsLead>
      Drop a complete verification UI into your app with zero frontend code.
      <code class="docs-inline-code">IdentityVerificationEmbed</code> creates an iframe pointing to the hosted verification
      page and communicates results via <code class="docs-inline-code">postMessage</code>.
    </DocsLead>

    <a-row :gutter="[16, 16]" class="docs-mode-row">
      <a-col v-for="mode in embedModes" :key="mode.title" :xs="24" :md="12">
        <a-card :bordered="true" class="docs-mode-card">
          <div class="docs-mode-head">
            <span class="docs-mode-title">{{ mode.title }}</span>
            <a-tag size="small" :color="mode.tagColor">{{ mode.tag }}</a-tag>
          </div>
          <p class="docs-mode-desc">{{ mode.desc }}</p>
          <div class="docs-mode-signals">
            <div v-for="feature in mode.features" :key="feature" class="docs-mode-signal">
              <span class="docs-check-mark">✓</span>{{ feature }}
            </div>
          </div>
        </a-card>
      </a-col>
    </a-row>

    <DocsCode label="Modal mode" :code="modalCode" />
    <DocsCode label="Inline mode" :code="inlineCode" />

    <a-card class="docs-table-card" :bordered="true">
      <div class="docs-info-label">Embed Options</div>
      <div class="docs-field-list">
        <DocsFieldRow name="mode" type="'modal' | 'inline'" desc="UI mode. Default: 'modal'" />
        <DocsFieldRow name="container" type="string | HTMLElement" desc="Target element for inline mode. CSS selector or DOM element." />
        <DocsFieldRow name="theme" type="'light' | 'dark'" desc="Color theme. Default: 'dark'" />
        <DocsFieldRow name="width" type="string" desc="Iframe width. Default: '100%'" />
        <DocsFieldRow name="height" type="string" desc="Iframe height. Default: '700px'" />
        <DocsFieldRow name="closeOnBackdropClick" type="boolean" desc="Close modal on backdrop click. Default: true" />
        <DocsFieldRow name="verificationUrl" type="string" desc="Base URL for hosted page. Default: https://verify.example.com" />
      </div>
    </a-card>

    <a-card class="docs-table-card" :bordered="true">
      <div class="docs-info-label">Callbacks</div>
      <div class="docs-field-list">
        <DocsFieldRow name="onComplete" type="(result) => void" desc="Verification succeeded. Result includes verificationId, status, finalResult." />
        <DocsFieldRow name="onError" type="(error) => void" desc="Verification failed. Error includes code and message." />
        <DocsFieldRow name="onStepChange" type="(step) => void" desc="Step changed. Step includes current, total, and status." />
        <DocsFieldRow name="onClose" type="() => void" desc="User closed the modal (modal mode only)." />
      </div>
    </a-card>

    <DocsCallout type="tip">
      The embed component communicates via <code>postMessage</code> with a
      <code>source: 'identity-embed'</code> identifier. Messages are ignored if they don't match this source, so
      it's safe to use alongside other iframes.
    </DocsCallout>
  </DocsShell>
</template>

<script setup lang="ts">
import DocsShell, {
  DocsSectionAnchor,
  DocsH2,
  DocsLead,
  DocsDivider,
  DocsCallout,
  DocsCode,
  DocsFieldRow,
  docsApiUrl,
  type DocsNavItem,
} from '../components/DocsShell.vue'

const apiUrl = docsApiUrl()

const NAV: DocsNavItem[] = [
  { id: 'sdk', label: 'JavaScript SDK', depth: 0 },
  { id: 'embed', label: 'Embed Component', depth: 0 },
]

const methods = [
  { method: 'startVerification()', returns: 'InitializeResponse', desc: 'Create a new verification session' },
  { method: 'uploadFrontDocument()', returns: 'VerificationResult', desc: 'Upload front of ID for OCR extraction' },
  { method: 'uploadBackDocument()', returns: 'VerificationResult', desc: 'Upload back of ID for cross-validation' },
  { method: 'uploadSelfie()', returns: 'VerificationResult', desc: 'Submit live capture for face match' },
  { method: 'getVerificationStatus()', returns: 'VerificationResult', desc: 'Get full session status and results' },
  { method: 'watch()', returns: 'EventEmitter', desc: 'Real-time event stream (see below)' },
  { method: 'createBatch()', returns: 'BatchJobResponse', desc: 'Create batch verification job' },
  { method: 'uploadAddressDocument()', returns: 'AddressResult', desc: 'Upload proof-of-address document' },
  { method: 'createMonitoringSchedule()', returns: 'Schedule', desc: 'Schedule re-verification' },
]

const methodColumns = [
  { title: 'Method', slotName: 'method', width: 220 },
  { title: 'Returns', slotName: 'returns', width: 190 },
  { title: 'Description', dataIndex: 'desc' },
]

const embedModes = [
  {
    title: 'Modal Mode',
    tag: 'Default',
    tagColor: 'green',
    desc: 'Full-screen overlay with backdrop. Best for trigger-based flows (e.g. "Verify Now" button).',
    features: [
      'Dark overlay backdrop',
      'Close on backdrop click or ✕ button',
      'Auto-removes on completion',
      'Blocks page scroll while open',
    ],
  },
  {
    title: 'Inline Mode',
    tag: 'Container',
    tagColor: 'arcoblue',
    desc: 'Renders inside a DOM element you specify. Best for embedding within an existing page layout.',
    features: ['Fits your container', 'Custom width/height', 'No overlay or backdrop', 'Stays in page flow'],
  },
]

const installCode = `npm install @identity-verification/sdk`

const initCode = `const { IdentityVerificationSDK } = require('@identity-verification/sdk');

const sdk = new IdentityVerificationSDK({
  apiKey: 'ik_your_api_key',
  baseURL: '${apiUrl}',     // optional, defaults to your deployment's base URL
  sandbox: false,            // set true for sandbox mode
});`

const watchCode = `const watcher = sdk.watch(verificationId);

// Listen for specific events
watcher.on('status_changed', (event) => {
  console.log('Status:', event.status);   // e.g. 'AWAITING_BACK'
});

watcher.on('step_completed', (event) => {
  console.log('Step', event.data.current_step, 'done');
});

watcher.on('verification_complete', (event) => {
  console.log('Verified!', event.data.final_result);
  // 'verified' | 'manual_review' | 'failed'
});

watcher.on('verification_failed', (event) => {
  console.log('Rejected:', event.data.rejection_reason);
});

// One-time listener
watcher.once('verification_complete', (event) => {
  redirectToSuccess(event.data.verification_id);
});

// Wildcard — catch everything
watcher.on('*', (event) => analytics.track(event.type));

// Clean up when done
watcher.destroy();`

const modalCode = `import { IdentityVerificationEmbed } from '@identity-verification/sdk';

const embed = new IdentityVerificationEmbed({
  mode: 'modal',
  theme: 'dark',
});

// Open with a session token (from your backend)
embed.open(sessionToken, {
  onComplete: (result) => {
    console.log('Verified!', result.verificationId);
    console.log('Result:', result.finalResult);
  },
  onError: (error) => {
    console.error(error.code, error.message);
  },
  onStepChange: (step) => {
    console.log('Step', step.current, 'of', step.total);
  },
  onClose: () => {
    console.log('User closed the modal');
  },
});

// Programmatically close
embed.close();`

const inlineCode = `const embed = new IdentityVerificationEmbed({
  mode: 'inline',
  container: '#verification-container',  // CSS selector or HTMLElement
  theme: 'dark',
  width: '100%',
  height: '700px',
});

embed.open(sessionToken, {
  onComplete: (result) => showSuccessMessage(result),
});`
</script>

<style>
.docs-inline-code {
  font-family: ui-monospace, Menlo, monospace;
  font-size: 12px;
  color: #0e7490;
  background: #ecfeff;
  padding: 1px 5px;
}
.docs-table-card { margin: 14px 0; background: #fbfcfe; }
.docs-table-card .arco-card-body { padding: 14px 18px; }
.docs-table-code { font-family: ui-monospace, Menlo, monospace; font-size: 12px; color: #0e7490; white-space: nowrap; }
.docs-table-return { font-family: ui-monospace, Menlo, monospace; font-size: 11.5px; color: #64748b; }
.docs-mode-row { margin: 16px 0 18px; }
.docs-mode-card { height: 100%; }
.docs-mode-card .arco-card-body { display: flex; flex-direction: column; gap: 10px; padding: 18px 20px; }
.docs-mode-head { display: flex; align-items: center; gap: 8px; }
.docs-mode-title { font-size: 14px; font-weight: 700; }
.docs-mode-desc { margin: 0; font-size: 12.5px; color: #475569; line-height: 1.6; }
.docs-mode-signals { display: flex; flex-direction: column; gap: 3px; }
.docs-mode-signal { display: flex; gap: 8px; font-size: 12px; color: #64748b; }
</style>
