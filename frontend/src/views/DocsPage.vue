<template>
  <DocsShell slug="api-docs" :nav="NAV">
    <!-- QUICK START -->
    <DocsSectionAnchor id="quick-start" />
    <DocsH2 index="01">Quick Start</DocsH2>
    <DocsLead>
      The verification flow is a 5-call sequence. Every processing step is asynchronous: submit, then poll
      GET /api/v2/verify/:id/status until the relevant field appears.
    </DocsLead>

    <DocsPipeline />

    <DocsCodeTabs v-model="tab" :js="quickStartJs" :python="quickStartPython" />

    <DocsDivider />

    <!-- AUTHENTICATION -->
    <DocsSectionAnchor id="auth" />
    <DocsH2 index="02">Authentication</DocsH2>
    <DocsLead>
      Every API request must include your API key in the X-API-Key header. You can generate keys in the
      Developer Portal.
    </DocsLead>

    <a-card class="docs-info-card" :bordered="true">
      <div class="docs-info-grid">
        <div>
          <div class="docs-info-label">Base URL</div>
          <code class="docs-info-value">{{ apiUrl }}</code>
        </div>
        <div>
          <div class="docs-info-label">Auth Header</div>
          <code class="docs-info-value">X-API-Key: ik_your_api_key</code>
        </div>
      </div>
    </a-card>

    <DocsCode label="Curl example" :code="authCurl" />

    <DocsCallout type="tip">
      Use a sandbox key during development. Sandbox mode is a property of the key (set in the Developer Portal),
      not a prefix distinction — all keys use the ik_ prefix. Sandbox uses the same OCR and face-matching
      pipeline, but counts against a separate quota and does not affect production metrics.
    </DocsCallout>

    <DocsDivider />

    <!-- VERIFICATION FLOW -->
    <DocsSectionAnchor id="flow" />
    <DocsH2 index="03">Verification Flow</DocsH2>
    <DocsLead>
      Each verification is a session with a unique ID. Steps must be completed in order, and each step unlocks
      the next. Heavy processing (OCR, cross-validation, liveness) is asynchronous: submit the data, then poll
      GET /api/v2/verify/:id/status to check progress.
    </DocsLead>

    <!-- Step 1 -->
    <DocsSectionAnchor id="step-1" />
    <DocsEndpointCard :step="1" method="POST" path="/api/v2/verify/initialize" title="Start a Verification Session">
      <p class="docs-body-text">
        Creates a new verification session for a user. Returns a verification_id that ties together all
        subsequent uploads and results. One session equals one complete identity check.
      </p>
      <div class="docs-field-list">
        <DocsFieldRow name="user_id" type="UUID string" :req="true" desc="Your unique identifier for the user being verified." />
        <DocsFieldRow name="sandbox" type="boolean" desc="Set true to use sandbox mode. Defaults to false." />
        <DocsFieldRow name="addons" type="object" desc="Optional add-on features to enable for this session." />
        <DocsFieldRow
          name="addons.aml_screening"
          type="boolean"
          desc="Enable AML/sanctions screening against OFAC, EU and UN lists. PEP screening is also available via AML_PROVIDER=pep. Runs automatically after identity verification completes."
        />
      </div>
      <DocsCodeTabs v-model="tab" :curl="step1Curl" :js="step1Js" :python="step1Python" />
      <DocsCode label="Response — HTTP 201" :code="step1Response" />
    </DocsEndpointCard>

    <!-- Step 2 -->
    <DocsSectionAnchor id="step-2" />
    <DocsEndpointCard :step="2" method="POST" path="/api/v2/verify/:id/front-document" title="Upload Front Document">
      <p class="docs-body-text">
        Upload the <strong>front face</strong> of the identity document (passport, driver's license, national
        ID). OCR extraction and image quality analysis (Gate 1) run during upload. The response includes
        <code>ocr_data</code> with extracted fields and confidence scores. If the image quality is too poor, the
        session is hard-rejected.
      </p>
      <div class="docs-field-list">
        <DocsFieldRow name="document_type" type="string" :req="true" desc="'passport' | 'drivers_license' | 'national_id' | 'other'" />
        <DocsFieldRow name="document" type="File" :req="true" desc="JPEG, PNG, WebP, or PDF. Max 10 MB." />
      </div>
      <DocsCodeTabs v-model="tab" :curl="step2Curl" :js="step2Js" :python="step2Python" />
      <DocsCode label="Response — HTTP 201" :code="step2Response" />
      <DocsCallout type="note">
        If Gate 1 rejects the document (blur, low resolution), the session status becomes
        <code>HARD_REJECTED</code> and subsequent steps will return 409. Check <code>rejection_reason</code> and
        <code>rejection_detail</code> for guidance to display to the user.
      </DocsCallout>
    </DocsEndpointCard>

    <!-- Step 3 -->
    <DocsSectionAnchor id="step-3" />
    <DocsEndpointCard
      :step="3"
      method="POST"
      path="/api/v2/verify/:id/back-document"
      title="Upload Back-of-ID"
      badge="required"
      badge-color="red"
    >
      <p class="docs-body-text">
        Upload the <strong>back face</strong> of the ID for barcode/QR scanning and cross-validation against
        the front OCR data. This is a <strong class="docs-danger">required step</strong> — it unlocks live
        capture. If the front and back do not match the same document, the verification immediately moves to
        <DocsStatusPill status="failed" /> and live capture is blocked.
      </p>

      <a-card class="docs-info-card" :bordered="true">
        <div class="docs-info-label">What cross-validation checks</div>
        <div class="docs-check-grid">
          <div v-for="check in crossChecks" :key="check" class="docs-check-item"><span class="docs-check-mark">✓</span>{{ check }}</div>
        </div>
      </a-card>

      <div class="docs-field-list">
        <DocsFieldRow name="document_type" type="string" :req="true" desc="Must match the document_type used in Step 2." />
        <DocsFieldRow name="document" type="File" :req="true" desc="JPEG, PNG, or WebP. Max 10 MB. Field name is 'document'." />
      </div>
      <DocsCodeTabs v-model="tab" :curl="step3Curl" :js="step3Js" :python="step3Python" />
      <DocsCode label="Response — HTTP 201" :code="step3Response" />
      <DocsCallout type="warning">
        Cross-validation now auto-triggers during back-document upload. Poll
        <code>GET /api/v2/verify/:id/status</code> until <code>cross_validation_results</code> is populated. If
        <code>final_result</code> is <DocsStatusPill status="failed" />, cross-validation did not pass — do not
        proceed to live capture. Check <code>failure_reason</code> for details.
      </DocsCallout>
    </DocsEndpointCard>

    <!-- Step 4 -->
    <DocsSectionAnchor id="step-4" />
    <DocsEndpointCard
      :step="4"
      method="POST"
      path="/api/v2/verify/:id/live-capture"
      title="Submit Live Capture"
      badge="final gate"
      badge-color="green"
    >
      <p class="docs-body-text">
        One endpoint, two gates, two liveness modes. Submit a <strong>live capture image</strong> plus optional
        <code>liveness_metadata</code> to unlock stronger liveness detection. <strong>Gate 4 (Liveness)</strong>
        checks that the image is from a live person, then <strong>Gate 5 (Face Match)</strong> auto-triggers to
        compare against the document photo (128-d cosine similarity, threshold 0.60).
      </p>

      <a-row :gutter="[12, 12]" class="docs-mode-row">
        <a-col v-for="mode in livenessModes" :key="mode.title" :xs="24" :md="12">
          <a-card :bordered="true" class="docs-mode-card" :style="{ borderTop: `3px solid ${mode.color}` }">
            <div class="docs-mode-head">
              <span :style="{ color: mode.color }" class="docs-mode-title">{{ mode.title }}</span>
              <a-tag size="small" :color="mode.tagColor">{{ mode.label }}</a-tag>
            </div>
            <p class="docs-mode-desc">{{ mode.desc }}</p>
            <div class="docs-mode-signals">
              <div v-for="signal in mode.signals" :key="signal" class="docs-mode-signal">
                <span :style="{ color: mode.color }">▸</span>{{ signal }}
              </div>
            </div>
            <div class="docs-mode-foot">
              <div class="docs-info-label">Security</div>
              <span :style="{ color: mode.color }" class="docs-mode-strong">{{ mode.security }}</span>
            </div>
            <div>
              <div class="docs-info-label">Best for</div>
              <span class="docs-mode-best">{{ mode.bestFor }}</span>
            </div>
          </a-card>
        </a-col>
      </a-row>

      <a-card class="docs-info-card" :bordered="true">
        <div class="docs-info-label">Client-side requirements</div>
        <div class="docs-client-grid">
          <template v-for="row in clientRequirements" :key="row[0]">
            <code class="docs-info-value">{{ row[0] }}</code>
            <span class="docs-client-cell">{{ row[1] }}</span>
            <span class="docs-client-cell">{{ row[2] }}</span>
          </template>
        </div>
      </a-card>

      <DocsCallout type="tip">
        <strong>Head turn needs zero ML dependencies on the client.</strong> Just capture JPEG frames from the
        camera using standard browser APIs (<code>getUserMedia()</code> + <code>canvas.toDataURL()</code>). The
        server handles all face detection and yaw estimation. This is what the hosted verification page and demo
        use internally.
      </DocsCallout>

      <div class="docs-field-list">
        <DocsFieldRow
          name="selfie"
          type="File"
          :req="true"
          desc="JPEG or PNG image captured live from the user's camera via getUserMedia(). Max 10 MB. Static file uploads will fail liveness."
        />
        <DocsFieldRow
          name="liveness_metadata"
          type="JSON string"
          desc="JSON-stringified challenge data for head_turn liveness. Omit for passive mode (lowest pass rate)."
        />
      </div>

      <DocsCallout type="warning">
        <strong>The selfie must be a live camera capture, not a file upload.</strong> The liveness engine runs
        anti-spoofing checks on every image — even in passive mode (8 heuristic signals including EXIF analysis,
        compression artifacts, moire detection). A static photo uploaded from disk will almost certainly fail with
        <code>LIVENESS_FAILED</code> because it lacks camera metadata and has re-compression artifacts. Always use
        <code>getUserMedia()</code> to capture directly from the device camera.
      </DocsCallout>

      <DocsCode label="liveness_metadata — head_turn" :code="livenessMetadata" />

      <DocsCodeTabs v-model="tab" :curl="passiveCurl" :js="passiveJs" :python="passivePython" />
      <DocsCodeTabs v-model="tab" :curl="headTurnCurl" :js="headTurnJs" :python="headTurnPython" />

      <DocsCode label="Response — HTTP 201" :code="step4Response" />

      <DocsCallout type="tip">
        <strong>Which mode should you use?</strong> For low-risk flows (e.g. sandbox, basic onboarding), passive
        is fine. For all production identity verification, use <strong>head turn</strong> — it provides strong
        anti-spoofing with zero client-side ML dependencies.
      </DocsCallout>
      <DocsCallout type="warning">
        If <code>liveness_metadata</code> is provided but fails validation (wrong schema, missing fields), the API
        returns <strong>HTTP 400</strong> with a <code>VALIDATION_ERROR</code> code. It does not silently fall back
        to passive mode. Always check that your metadata format matches the schema above.
      </DocsCallout>
    </DocsEndpointCard>

    <!-- Step 5 -->
    <DocsSectionAnchor id="step-5" />
    <DocsEndpointCard :step="5" method="GET" path="/api/v2/verify/:id/status" title="Get Verification Results">
      <p class="docs-body-text">
        The single source of truth for a verification session. Use this endpoint for polling at each processing
        stage and for reading the final result. Returns the full record including OCR data, cross-validation
        scores, liveness score, and face match score.
      </p>

      <a-card class="docs-info-card" :bordered="true">
        <div class="docs-info-label">Polling conditions</div>
        <a-table
          :data="pollingRows"
          :columns="pollingColumns"
          :pagination="false"
          size="small"
          :bordered="false"
          row-key="after"
        >
          <template #condition="{ record }"><code class="docs-table-code">{{ record.condition }}</code></template>
          <template #next="{ record }"><span class="docs-table-next">→ {{ record.next }}</span></template>
        </a-table>
      </a-card>

      <DocsCodeTabs v-model="tab" :curl="step5Curl" :js="step5Js" :python="step5Python" />
      <DocsCode label="Response — completed verification" :code="step5Response" />
    </DocsEndpointCard>

    <!-- Cross-validation (optional) -->
    <DocsSectionAnchor id="selfie" />
    <DocsEndpointCard
      method="POST"
      path="/api/v2/verify/:id/cross-validation"
      title="Get Cross-Validation Results"
      badge="optional"
      badge-color="gray"
    >
      <p class="docs-body-text">
        Retrieve the cached cross-validation results separately. Cross-validation runs automatically during the
        back document upload (Step 3), so this endpoint is only needed if you want to query the results
        independently.
      </p>
      <DocsCode label="Response — HTTP 200" :code="crossValidationResponse" />
    </DocsEndpointCard>

    <!-- Live-capture selfie URL (optional) -->
    <DocsSectionAnchor id="selfie-url" />
    <DocsEndpointCard
      method="GET"
      path="/api/v2/verify/:id/selfie"
      title="Get Live-Capture Selfie URL"
      badge="optional"
      badge-color="gray"
    >
      <p class="docs-body-text">
        Returns a short-lived <strong>signed URL</strong> for the verification's live-capture selfie — handy when
        you want to run a downstream face match on your own side. Scoped to the authenticated API key: a
        verification you don't own (or an unknown ID) returns <code>404</code> (never 403), so the endpoint can't
        be used to enumerate other developers' verification IDs. Returns only the selfie URL — no OCR, PII, or
        other documents.
      </p>
      <DocsCodeTabs v-model="tab" :curl="selfieCurl" :js="selfieJs" :python="selfiePython" />
      <DocsCode label="Response — HTTP 200" :code="selfieResponse" />
      <DocsCallout type="note">
        The signed URL is valid for <strong>300 seconds</strong> — long enough to fetch once, short enough to not
        be a durable link. Re-call the endpoint for a fresh URL. If no live-capture selfie is on file yet, the
        endpoint returns <strong>HTTP 404</strong> with <code>{ "selfie_url": null }</code>.
      </DocsCallout>
    </DocsEndpointCard>
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
  DocsEndpointCard,
  DocsPipeline,
  DocsStatusPill,
  docsApiUrl,
  type DocsNavItem,
} from '../components/DocsShell.vue'

const apiUrl = docsApiUrl()
const tab = ref('curl')

const NAV: DocsNavItem[] = [
  { id: 'quick-start', label: 'Quick Start', depth: 0 },
  { id: 'auth', label: 'Authentication', depth: 0 },
  { id: 'flow', label: 'Verification Flow', depth: 0 },
  { id: 'step-1', label: '1 · Start Session', depth: 1 },
  { id: 'step-2', label: '2 · Upload Front', depth: 1 },
  { id: 'step-3', label: '3 · Upload Back', depth: 1 },
  { id: 'step-4', label: '4 · Live Capture', depth: 1 },
  { id: 'step-5', label: '5 · Get Results', depth: 1 },
  { id: 'selfie', label: 'Cross-Validation', depth: 1 },
  { id: 'selfie-url', label: 'Get Selfie URL', depth: 1 },
]

const crossChecks = [
  'PDF417 / QR barcode decoding',
  'ID number consistency (front OCR ↔ barcode)',
  'Expiry date matching',
  'Issuing authority matching',
  'Photo consistency score',
  'Security feature detection',
]

const livenessModes = [
  {
    title: 'Passive',
    label: 'Basic',
    tagColor: 'orange',
    color: '#b45309',
    desc: 'Just upload the image — no extra data needed. 8 heuristic signals (entropy, pixel variance, EXIF, file size, etc.).',
    signals: [
      'Byte entropy analysis',
      'Pixel variance & texture',
      'File size heuristics',
      'EXIF metadata checks',
      'Color distribution',
      'Digital artifact detection',
      'Face detection (SSDMobileNet)',
      'Frequency domain analysis',
    ],
    security: 'Basic',
    bestFor: 'Low-risk onboarding, sandbox testing',
  },
  {
    title: 'Head Turn',
    label: 'Recommended',
    tagColor: 'green',
    color: '#15803d',
    desc: 'Client captures 5–12 timed frames while the user turns their head. Server runs face detection plus 7 weighted checks. Pass ≥ 0.70. Zero ML on the client.',
    signals: [
      'All passive checks',
      'Face present across all frames',
      'Head turn detected (≥12° yaw)',
      'Correct direction compliance',
      'Return to center verified',
      'Temporal plausibility (3–60s)',
      'Face bounding box consistency',
      'Virtual camera detection',
    ],
    security: 'Strong',
    bestFor: 'All production identity verification',
  },
]

const clientRequirements = [
  ['Passive', 'None — just upload the image file', 'None'],
  ['Head turn', 'getUserMedia() + canvas.toDataURL()', 'None — all face analysis runs server-side'],
]

const pollingColumns = [
  { title: 'After', dataIndex: 'after', width: 190 },
  { title: 'Condition', slotName: 'condition' },
  { title: 'Next', slotName: 'next', width: 260 },
]

const pollingRows = [
  { after: 'After Step 2 (front doc)', condition: 'ocr_data is not null', next: 'proceed to Step 3' },
  {
    after: 'After Step 3 (back-of-ID)',
    condition: 'cross_validation_results is not null',
    next: 'check final_result; if not "failed", proceed to Step 4',
  },
  { after: 'After Step 4 (live capture)', condition: 'final_result is not null', next: 'verification complete' },
]

const quickStartJs = `const BASE = '${apiUrl}';
const KEY  = 'your-api-key';
const headers = { 'X-API-Key': KEY };

// 1. Start session
const { verification_id } = await fetch(\`\${BASE}/api/v2/verify/initialize\`, {
  method: 'POST',
  headers: { ...headers, 'Content-Type': 'application/json' },
  body: JSON.stringify({ user_id: 'user-uuid' }),
}).then(r => r.json());

// 2. Upload front document
const fd1 = new FormData();
fd1.append('document_type', 'drivers_license');
fd1.append('document', frontFile);       // File from <input type="file">
await fetch(\`\${BASE}/api/v2/verify/\${verification_id}/front-document\`, { method: 'POST', headers, body: fd1 });

// 3. Poll until OCR finishes
let r;
do {
  await new Promise(ok => setTimeout(ok, 2000));
  r = await fetch(\`\${BASE}/api/v2/verify/\${verification_id}/status\`, { headers }).then(r => r.json());
} while (!r.ocr_data);

// 4. Upload back-of-ID (cross-validation auto-triggers)
const fd2 = new FormData();
fd2.append('document_type', 'drivers_license');
fd2.append('document', backFile);        // field name is 'document' (not 'back_of_id')
await fetch(\`\${BASE}/api/v2/verify/\${verification_id}/back-document\`, { method: 'POST', headers, body: fd2 });

// 5. Poll until cross-validation finishes
do {
  await new Promise(ok => setTimeout(ok, 2000));
  r = await fetch(\`\${BASE}/api/v2/verify/\${verification_id}/status\`, { headers }).then(r => r.json());
} while (!r.cross_validation_results);

if (r.final_result === 'failed') throw new Error('Cross-validation failed: ' + r.failure_reason);

// 6. Submit live capture (file upload)
const fd3 = new FormData();
fd3.append('selfie', captureBlob, 'capture.jpg');  // Blob from canvas.toBlob()
await fetch(\`\${BASE}/api/v2/verify/\${verification_id}/live-capture\`, {
  method: 'POST',
  headers,
  body: fd3,
});

// 7. Poll for final result
do {
  await new Promise(ok => setTimeout(ok, 2000));
  r = await fetch(\`\${BASE}/api/v2/verify/\${verification_id}/status\`, { headers }).then(r => r.json());
} while (!r.final_result);

console.log(r.final_result);               // 'verified' | 'failed' | 'manual_review'
console.log(r.face_match_results?.similarity_score);   // 0.0 – 1.0
console.log(r.ocr_data.name);              // "Jane Smith"`

const quickStartPython = `import requests, time

BASE = '${apiUrl}'
HEADERS = {'X-API-Key': 'your-api-key'}

def poll(vid, until_key=None, until_final=False):
    """Poll /status until condition is met."""
    while True:
        time.sleep(2)
        r = requests.get(f'{BASE}/api/v2/verify/{vid}/status', headers=HEADERS).json()
        if until_key and r.get(until_key): return r
        if until_final and r.get('final_result') is not None: return r

# 1. Start session
session = requests.post(f'{BASE}/api/v2/verify/initialize',
    headers={**HEADERS, 'Content-Type': 'application/json'},
    json={'user_id': 'user-uuid'}
).json()
vid = session['verification_id']

# 2. Upload front document
with open('front.jpg', 'rb') as f:
    requests.post(f'{BASE}/api/v2/verify/{vid}/front-document', headers=HEADERS,
        data={'document_type': 'drivers_license'},
        files={'document': f})

# 3. Poll until OCR is ready
r = poll(vid, until_key='ocr_data')
print('Name:', r['ocr_data']['name'])

# 4. Upload back-of-ID (cross-validation auto-triggers)
with open('back.jpg', 'rb') as f:
    requests.post(f'{BASE}/api/v2/verify/{vid}/back-document', headers=HEADERS,
        data={'document_type': 'drivers_license'},
        files={'document': f})

# 5. Poll until cross-validation finishes
r = poll(vid, until_key='cross_validation_results')
if r.get('final_result') == 'failed':
    raise Exception('Cross-validation failed: ' + r.get('failure_reason', ''))

# 6. Submit live capture (file upload)
with open('capture.jpg', 'rb') as f:
    requests.post(f'{BASE}/api/v2/verify/{vid}/live-capture', headers=HEADERS,
        files={'selfie': f})  # field name is 'selfie' in the API

# 7. Poll for final result
r = poll(vid, until_final=True)
print(r['final_result'])                  # verified / failed / manual_review
print(r.get('face_match_results', {}))    # match scores
print(r['ocr_data']['name'])              # "Jane Smith"`

const authCurl = `curl -X POST ${apiUrl}/api/v2/verify/initialize \\
  -H "X-API-Key: ik_your_api_key" \\
  -H "Content-Type: application/json" \\
  -d '{"user_id": "550e8400-e29b-41d4-a716-446655440000"}'`

const step1Curl = `curl -X POST ${apiUrl}/api/v2/verify/initialize \\
  -H "X-API-Key: your-key" \\
  -H "Content-Type: application/json" \\
  -d '{
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "addons": { "aml_screening": true }
  }'`

const step1Js = `const res = await fetch(\`${apiUrl}/api/v2/verify/initialize\`, {
  method: 'POST',
  headers: {
    'X-API-Key': 'your-key',
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    user_id: '550e8400-e29b-41d4-a716-446655440000',
    addons: { aml_screening: true },
  }),
});
const data = await res.json();`

const step1Python = `import requests

res = requests.post(
    '${apiUrl}/api/v2/verify/initialize',
    headers={ 'X-API-Key': 'your-key' },
    json={
        'user_id': '550e8400-e29b-41d4-a716-446655440000',
        'addons': { 'aml_screening': True },
    },
)
data = res.json()`

const step1Response = `{
  "success": true,
  "verification_id": "550e8400-e29b-41d4-a716-446655440001",
  "status": "AWAITING_FRONT",
  "current_step": 1,
  "total_steps": 5,
  "message": "Verification session initialized"
}`

const step2Curl = `curl -X POST ${apiUrl}/api/v2/verify/550e8400-e29b-41d4-a716-446655440001/front-document \\
  -H "X-API-Key: your-key" \\
  -F "document_type=drivers_license" \\
  -F "document=@front.jpg"`

const step2Js = `const fd = new FormData();
fd.append('document_type', 'drivers_license');
fd.append('document', frontFile); // File from <input type="file">

const res = await fetch(
  \`${apiUrl}/api/v2/verify/\${verification_id}/front-document\`,
  { method: 'POST', headers: { 'X-API-Key': 'your-key' }, body: fd },
);
const data = await res.json();`

const step2Python = `import requests

res = requests.post(
    f'${apiUrl}/api/v2/verify/{verification_id}/front-document',
    headers={ 'X-API-Key': 'your-key' },
    data={ 'document_type': 'drivers_license' },
    files={ 'document': open('front.jpg', 'rb') },
)
data = res.json()`

const step2Response = `{
  "success": true,
  "verification_id": "550e8400-e29b-41d4-a716-446655440001",
  "status": "AWAITING_BACK",
  "current_step": 2,
  "document_id": "doc_abc123",
  "message": "Front document processed successfully",

  // OCR extraction results (included in response):
  "ocr_data": {
    "full_name": "JANE SMITH",
    "date_of_birth": "1990-06-15",
    "id_number": "DL123456789",
    "expiration_date": "2030-06-15",
    "address": "123 Main St, Anytown, US",
    "confidence_scores": { "full_name": 0.97, "date_of_birth": 0.96 }
  }

  // If Gate 1 (quality) fails → status becomes "HARD_REJECTED":
  // "rejection_reason": "POOR_QUALITY",
  // "rejection_detail": "Image too blurry for OCR extraction"
}`

const step3Curl = `curl -X POST ${apiUrl}/api/v2/verify/550e8400-e29b-41d4-a716-446655440001/back-document \\
  -H "X-API-Key: your-key" \\
  -F "document_type=drivers_license" \\
  -F "document=@back.jpg"`

const step3Js = `const fd = new FormData();
fd.append('document_type', 'drivers_license');
fd.append('document', backFile); // File from <input type="file">

const res = await fetch(
  \`${apiUrl}/api/v2/verify/\${verification_id}/back-document\`,
  { method: 'POST', headers: { 'X-API-Key': 'your-key' }, body: fd },
);
const data = await res.json();`

const step3Python = `import requests

res = requests.post(
    f'${apiUrl}/api/v2/verify/{verification_id}/back-document',
    headers={ 'X-API-Key': 'your-key' },
    data={ 'document_type': 'drivers_license' },
    files={ 'document': open('back.jpg', 'rb') },
)
data = res.json()`

const step3Response = `{
  "success": true,
  "verification_id": "550e8400-e29b-41d4-a716-446655440001",
  "status": "AWAITING_LIVE",
  "current_step": 3,
  "message": "Back document processed successfully",

  // Barcode / PDF417 extraction:
  "barcode_data": {
    "first_name": "JANE",
    "last_name": "SMITH",
    "date_of_birth": "19900615",
    "document_number": "DL123456789"
  },

  // Cross-validation (front OCR vs back barcode):
  "documents_match": true,
  "cross_validation_results": {
    "verdict": "PASS",            // "PASS" or "REVIEW"
    "has_critical_failure": false,
    "score": 0.95,                // 0.0 – 1.0
    "failures": []                // e.g. ["DOB_MISMATCH", "NAME_MISMATCH"]
  }
}`

const livenessMetadata = `{
  "challenge_type": "head_turn",
  "challenge_direction": "left",
  "frames": [
    { "frame_base64": "/9j/4AAQ...", "timestamp": 1000,  "phase": "turn1_start" },
    { "frame_base64": "/9j/4BBR...", "timestamp": 4000,  "phase": "turn1_peak" },
    { "frame_base64": "/9j/4CCG...", "timestamp": 7000,  "phase": "turn1_return" },
    { "frame_base64": "/9j/4DDH...", "timestamp": 8200,  "phase": "turn_start" },
    { "frame_base64": "/9j/4EEI...", "timestamp": 11200, "phase": "turn_peak" },
    { "frame_base64": "/9j/4FFJ...", "timestamp": 14200, "phase": "turn_return" }
  ],
  "start_timestamp": 0,
  "end_timestamp": 15000
}`

const passiveCurl = `# Passive liveness (just the file, no metadata)
curl -X POST ${apiUrl}/api/v2/verify/VERIFICATION_ID/live-capture \\
  -H "X-API-Key: your-key" \\
  -F "selfie=@capture.jpg"`

const passiveJs = `// Passive liveness — just the file
const fd = new FormData();
fd.append('selfie', captureFile);

const res = await fetch(
  \`${apiUrl}/api/v2/verify/\${verification_id}/live-capture\`,
  { method: 'POST', headers: { 'X-API-Key': 'your-key' }, body: fd },
);
const data = await res.json();`

const passivePython = `# Passive liveness — just the file
import requests

res = requests.post(
    f'${apiUrl}/api/v2/verify/{verification_id}/live-capture',
    headers={'X-API-Key': 'your-key'},
    files={'selfie': open('capture.jpg', 'rb')},
)
data = res.json()`

const headTurnCurl = `# Head-turn liveness — capture frames from camera, server analyzes face
curl -X POST ${apiUrl}/api/v2/verify/VERIFICATION_ID/live-capture \\
  -H "X-API-Key: your-key" \\
  -F "selfie=@capture.jpg" \\
  -F 'liveness_metadata={
    "challenge_type": "head_turn",
    "challenge_direction": "left",
    "frames": [
      {"frame_base64":"<base64>","timestamp":1000,"phase":"turn1_start"},
      {"frame_base64":"<base64>","timestamp":4000,"phase":"turn1_peak"},
      {"frame_base64":"<base64>","timestamp":7000,"phase":"turn1_return"},
      {"frame_base64":"<base64>","timestamp":8200,"phase":"turn_start"},
      {"frame_base64":"<base64>","timestamp":11200,"phase":"turn_peak"},
      {"frame_base64":"<base64>","timestamp":14200,"phase":"turn_return"}
    ],
    "start_timestamp": 0,
    "end_timestamp": 15000
  }'`

const headTurnJs = `// Head-turn liveness — capture frames from camera, server analyzes face
const fd = new FormData();
fd.append('selfie', captureFile);
fd.append('liveness_metadata', JSON.stringify({
  challenge_type: 'head_turn',
  challenge_direction: 'left',
  frames: capturedFrames, // from canvas.toDataURL('image/jpeg', 0.7)
  // Each frame: { frame_base64, timestamp, phase }
  // Phases: turn1_start → turn1_peak → turn1_return → turn_start → turn_peak → turn_return
  start_timestamp: startTime,
  end_timestamp: Date.now(),
}));

const res = await fetch(
  \`${apiUrl}/api/v2/verify/\${verification_id}/live-capture\`,
  { method: 'POST', headers: { 'X-API-Key': 'your-key' }, body: fd },
);
const data = await res.json();`

const headTurnPython = `# Head-turn liveness — capture frames from camera, server analyzes face
import requests, json

metadata = {
    'challenge_type': 'head_turn',
    'challenge_direction': 'left',
    'frames': captured_frames,  # list of {frame_base64, timestamp, phase}
    # Phases: turn1_start → turn1_peak → turn1_return → turn_start → turn_peak → turn_return
    'start_timestamp': start_time,
    'end_timestamp': end_time,
}

res = requests.post(
    f'${apiUrl}/api/v2/verify/{verification_id}/live-capture',
    headers={'X-API-Key': 'your-key'},
    files={'selfie': open('capture.jpg', 'rb')},
    data={'liveness_metadata': json.dumps(metadata)},
)
data = res.json()`

const step4Response = `{
  "success": true,
  "verification_id": "550e8400-e29b-41d4-a716-446655440001",
  "status": "COMPLETE",
  "current_step": 5,
  "selfie_id": "selfie_abc789",
  "message": "Verification complete",

  // Face match (live capture vs document photo):
  "face_match_results": {
    "passed": true,
    "similarity_score": 0.94,     // cosine similarity 0.0 – 1.0
    "threshold_used": 0.6         // configurable confidence threshold
  },

  // Liveness detection:
  "liveness_results": {
    "liveness_passed": true,
    "liveness_score": 0.96,
    "liveness_mode": "head_turn"  // "passive" | "head_turn"
  },

  // Age estimation (risk signal, never pass/fail):
  "age_estimation": {
    "document_face_age": 33,      // estimated from document photo
    "live_face_age": 35,          // estimated from live capture
    "declared_age": 35,           // calculated from DOB
    "age_discrepancy": 0          // |live_face_age - declared_age|
  },

  // Velocity analysis (fraud detection):
  "velocity_analysis": {
    "ip_verifications_1h": 1,     // same IP in last hour
    "ip_verifications_24h": 2,    // same IP in last 24 hours
    "user_verifications_24h": 1,  // same user in last 24 hours
    "avg_step_duration_ms": 8500, // avg time between steps
    "fastest_step_ms": 5200,      // fastest single step
    "flags": [],                  // e.g. ["rapid_ip_reuse", "bot_like_timing"]
    "score": 0                    // 0-100, highest flag score
  },

  // IP geolocation risk:
  "geo_analysis": {
    "ip_country": "US",           // ISO alpha-2 from IP lookup
    "ip_region": "CA",
    "ip_city": "San Francisco",
    "document_country": "US",     // from front document OCR
    "is_tor": false,
    "is_datacenter": false,
    "flags": [],                  // e.g. ["country_mismatch", "tor_exit_node"]
    "score": 0                    // 0-100, highest flag score
  },

  // Final auto-decision:
  "final_result": "verified"      // "verified" | "manual_review" | "failed"
}`

const step5Curl = `curl -X GET ${apiUrl}/api/v2/verify/550e8400-e29b-41d4-a716-446655440001/status \\
  -H "X-API-Key: your-key"`

const step5Js = `const res = await fetch(
  \`${apiUrl}/api/v2/verify/\${verification_id}/status\`,
  { headers: { 'X-API-Key': 'your-key' } },
);
const data = await res.json();`

const step5Python = `import requests

res = requests.get(
    f'${apiUrl}/api/v2/verify/{verification_id}/status',
    headers={ 'X-API-Key': 'your-key' },
)
data = res.json()`

const crossValidationResponse = `{
  "success": true,
  "verification_id": "550e8400-e29b-41d4-a716-446655440001",
  "status": "AWAITING_LIVE",
  "documents_match": true,
  "cross_validation_results": {
    "verdict": "PASS",
    "has_critical_failure": false,
    "score": 0.95,
    "failures": []
  }
}`

const selfieCurl = `curl -X GET ${apiUrl}/api/v2/verify/550e8400-e29b-41d4-a716-446655440001/selfie \\
  -H "X-API-Key: your-key"`

const selfieJs = `const res = await fetch(
  \`${apiUrl}/api/v2/verify/\${verification_id}/selfie\`,
  { headers: { 'X-API-Key': 'your-key' } },
);
const data = await res.json();`

const selfiePython = `import requests

res = requests.get(
    f'${apiUrl}/api/v2/verify/{verification_id}/selfie',
    headers={ 'X-API-Key': 'your-key' },
)
data = res.json()`

const selfieResponse = `{
  "verification_id": "550e8400-e29b-41d4-a716-446655440001",
  "selfie_url": "https://storage.example.com/selfies/…?token=…",
  "expires_in": 300
}`

const step5Response = `{
  "success": true,
  "verification_id": "550e8400-e29b-41d4-a716-446655440001",
  "status": "COMPLETE",
  "current_step": 5,
  "total_steps": 5,
  "created_at": "2026-03-06T12:00:00Z",
  "updated_at": "2026-03-06T12:05:30Z",

  // ── Upload progress ──────────────────────────────────────────
  "front_document_uploaded": true,
  "back_document_uploaded": true,
  "live_capture_uploaded": true,

  // ── Front document (OCR) ─────────────────────────────────────
  "ocr_data": {
    "full_name":         "JANE SMITH",
    "date_of_birth":     "1990-06-15",
    "id_number":         "DL123456789",
    "expiration_date":   "2030-06-15",
    "address":           "123 Main St, Anytown, US",
    "confidence_scores": { "full_name": 0.97, "date_of_birth": 0.96 }
  },

  // ── Back-of-ID (barcode + cross-validation) ──────────────────
  "barcode_data": {
    "first_name": "JANE", "last_name": "SMITH",
    "document_number": "DL123456789"
  },
  "documents_match": true,
  "cross_validation_results": {
    "verdict": "PASS",
    "has_critical_failure": false,
    "overall_score": 0.95,
    "failures": []
  },

  // ── Live capture (liveness + face match) ─────────────────────
  "face_match_results": {
    "passed": true,
    "similarity_score": 0.94,
    "threshold_used": 0.6
  },
  "liveness_results": {
    "passed": true,
    "score": 0.96
  },

  // ── AML / Sanctions / PEP (when AML_PROVIDER is configured) ──
  "aml_screening": {
    "risk_level": "clear",           // "clear" | "potential_match" | "confirmed_match"
    "match_found": false,
    "match_count": 0,
    "lists_checked": ["us_ofac_sdn", "eu_sanctions", "un_sanctions", "opensanctions-pep"],
    "screened_at": "2026-03-06T12:05:28Z"
    // PEP matches include match_type: "pep", list_source: "opensanctions-pep"
  },

  // ── Age estimation (risk signal, never pass/fail) ─────────
  "age_estimation": {
    "document_face_age": 33,       // estimated from document photo
    "live_face_age": 35,           // estimated from live capture
    "declared_age": 35,            // calculated from DOB
    "age_discrepancy": 0           // |live_face_age - declared_age|
  },

  // ── Velocity analysis (fraud detection) ─────────────────────
  "velocity_analysis": {
    "ip_verifications_1h": 1,
    "ip_verifications_24h": 2,
    "user_verifications_24h": 1,
    "avg_step_duration_ms": 8500,
    "fastest_step_ms": 5200,
    "flags": [],
    "score": 0
  },

  // ── IP geolocation risk ────────────────────────────────────────
  "geo_analysis": {
    "ip_country": "US",
    "ip_region": "CA",
    "ip_city": "San Francisco",
    "document_country": "US",
    "is_tor": false,
    "is_datacenter": false,
    "flags": [],
    "score": 0
  },

  // ── Final decision ───────────────────────────────────────────
  "final_result": "verified",       // "verified" | "manual_review" | "failed"
  "failure_reason": null,           // set if final_result is "failed"
  "manual_review_reason": null      // set if final_result is "manual_review"
}`
</script>

<style>
.docs-info-card { margin: 14px 0; background: #fbfcfe; }
.docs-info-card .arco-card-body { padding: 14px 18px; }
.docs-info-label {
  font-family: ui-monospace, Menlo, monospace;
  font-size: 10.5px;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: #64748b;
  margin-bottom: 8px;
}
.docs-info-grid { display: flex; gap: 28px; flex-wrap: wrap; }
.docs-info-value { font-family: ui-monospace, Menlo, monospace; font-size: 13px; color: #0e7490; }
.docs-field-list { margin-bottom: 8px; }
.docs-check-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 6px 24px; }
.docs-check-item { display: flex; gap: 8px; align-items: center; font-size: 12.5px; color: #475569; }
.docs-check-mark { color: #15803d; font-size: 11px; }
.docs-danger { color: #b91c1c; }
.docs-mode-row { margin: 16px 0 12px; }
.docs-mode-card { height: 100%; }
.docs-mode-card .arco-card-body { display: flex; flex-direction: column; gap: 10px; padding: 16px 18px; }
.docs-mode-head { display: flex; align-items: center; gap: 8px; }
.docs-mode-title { font-family: ui-monospace, Menlo, monospace; font-size: 13px; font-weight: 700; }
.docs-mode-desc { margin: 0; font-size: 12.5px; color: #475569; line-height: 1.6; }
.docs-mode-signals { display: flex; flex-direction: column; gap: 2px; }
.docs-mode-signal { display: flex; gap: 6px; font-size: 12px; color: #64748b; }
.docs-mode-foot { margin-top: auto; padding-top: 8px; border-top: 1px solid #eef2f7; }
.docs-mode-strong { font-size: 12.5px; font-weight: 600; }
.docs-mode-best { font-size: 12px; color: #64748b; }
.docs-client-grid { display: grid; grid-template-columns: 90px 1fr 1fr; gap: 8px 16px; }
.docs-client-cell { font-size: 12.5px; color: #475569; line-height: 1.5; }
.docs-table-code { font-family: ui-monospace, Menlo, monospace; font-size: 11.5px; color: #0e7490; }
.docs-table-next { font-size: 12px; color: #15803d; }
@media (max-width: 767px) {
  .docs-check-grid,
  .docs-client-grid { grid-template-columns: 1fr; }
}
</style>
