<template>
  <DocsShell slug="docs/features" :nav="NAV">
    <!-- ANALYSIS ENGINE -->
    <DocsSectionAnchor id="analysis" />
    <DocsH2 index="01">Analysis Engine</DocsH2>
    <DocsLead>
      What the platform extracts and validates from each document and capture. Processing uses algorithmic
      rules, image forensics, and pre-trained OCR / face-detection models — no custom AI or LLMs.
    </DocsLead>

    <a-row :gutter="[12, 12]">
      <a-col v-for="column in analysisColumns" :key="column.title" :xs="24" :sm="12" :md="8">
        <a-card :bordered="true" class="docs-analysis-card">
          <div class="docs-analysis-title" :style="{ color: column.color }">{{ column.title }}</div>
          <div v-for="item in column.items" :key="item" class="docs-mode-signal">
            <span :style="{ color: column.color }">▸</span>{{ item }}
          </div>
        </a-card>
      </a-col>
    </a-row>

    <DocsDivider />

    <!-- STATUSES -->
    <DocsSectionAnchor id="statuses" />
    <DocsH2 index="02">Verification Statuses</DocsH2>
    <DocsLead>
      A verification moves through these statuses sequentially.
      <strong class="docs-success">COMPLETE</strong> and <strong class="docs-danger">HARD_REJECTED</strong> are
      terminal states.
    </DocsLead>

    <a-card class="docs-table-card" :bordered="true">
      <div v-for="row in statuses" :key="row.status" class="docs-status-line">
        <div class="docs-status-line-pill">
          <DocsStatusPill :status="row.status" />
          <span v-if="row.terminal" class="docs-terminal" :style="{ color: row.color }">TERMINAL</span>
        </div>
        <span class="docs-status-line-desc">{{ row.desc }}</span>
      </div>
    </a-card>

    <DocsDivider />

    <!-- BATCH API -->
    <DocsSectionAnchor id="batch" />
    <DocsH2 index="03">Batch Verification</DocsH2>
    <DocsLead>
      Process hundreds of verifications at once for enterprise onboarding, user migration, or periodic
      re-verification. Items are processed with controlled concurrency (5 concurrent max).
    </DocsLead>

    <DocsEndpointCard method="POST" path="/api/v2/batch/upload" title="Create Batch Job">
      <p class="docs-body-text">
        Submit an array of items for batch processing. Each item requires a
        <code class="docs-inline-code">user_id</code> and optionally includes document URLs, document type, and
        metadata. Returns a <code class="docs-inline-code">batch_id</code> for tracking.
      </p>
      <div class="docs-field-list">
        <DocsFieldRow name="items" type="array" :req="true" desc="Array of verification items. Each needs user_id; optionally front/back/selfie URLs." />
        <DocsFieldRow name="items[].user_id" type="string" :req="true" desc="Unique identifier for the user being verified." />
        <DocsFieldRow name="items[].document_type" type="string" desc="'passport' | 'drivers_license' | 'national_id'" />
        <DocsFieldRow name="items[].front_document_url" type="URL" desc="URL to the front document image." />
        <DocsFieldRow name="items[].back_document_url" type="URL" desc="URL to the back document image." />
        <DocsFieldRow name="items[].selfie_url" type="URL" desc="URL to the selfie/live capture image." />
      </div>
      <DocsCode label="Request" :code="batchRequest" />
      <DocsCode label="Response — HTTP 201" :code="batchCreateResponse" />
    </DocsEndpointCard>

    <DocsEndpointCard method="GET" path="/api/v2/batch/:id/status" title="Get Batch Progress">
      <p class="docs-body-text">
        Poll this endpoint to track batch progress. Includes item counts and a percentage.
      </p>
      <DocsCode label="Response" :code="batchStatusResponse" />
    </DocsEndpointCard>

    <DocsEndpointCard method="GET" path="/api/v2/batch/:id/results" title="Get Batch Results">
      <p class="docs-body-text">
        Download individual item results for a completed batch. Each item includes its verification ID (if
        created) and status.
      </p>
      <DocsCode label="Response" :code="batchResultsResponse" />
    </DocsEndpointCard>

    <DocsEndpointCard method="POST" path="/api/v2/batch/:id/cancel" title="Cancel Batch Job">
      <p class="docs-body-text">
        Cancel a running batch job. Already-completed items are unaffected. Remaining items are skipped.
      </p>
    </DocsEndpointCard>

    <DocsCode label="SDK usage" :code="batchSdk" />

    <DocsDivider />

    <!-- ADDRESS VERIFICATION -->
    <DocsSectionAnchor id="address" />
    <DocsH2 index="04">Address Verification</DocsH2>
    <DocsLead>
      Verify proof-of-address documents (utility bills, bank statements, tax documents) and cross-reference the
      name against the verified ID. Requires a completed identity verification first.
    </DocsLead>

    <DocsEndpointCard
      method="POST"
      path="/api/v2/verify/:id/address-document"
      title="Upload Address Document"
      badge="post-verification"
      badge-color="arcoblue"
    >
      <p class="docs-body-text">
        Upload a proof-of-address document for OCR extraction and name cross-validation against the identity
        document. The address document must show the same person's name and be recently dated.
      </p>
      <div class="docs-field-list">
        <DocsFieldRow name="document" type="File" :req="true" desc="JPEG, PNG, or PDF. Max 10 MB." />
        <DocsFieldRow name="document_type" type="string" :req="true" desc="'utility_bill' | 'bank_statement' | 'tax_document'" />
      </div>
      <DocsCode label="Request" :code="addressRequest" />
      <DocsCode label="Response — HTTP 200" :code="addressResponse" />
      <DocsCallout type="note">
        The <code>name_match_score</code> compares the name on the address document against the name extracted
        from the identity document (OCR). A score below 0.7 routes to <code>review</code> status. The
        <code>document_fresh</code> flag checks if the document is dated within the last 90 days.
      </DocsCallout>
    </DocsEndpointCard>

    <DocsEndpointCard method="GET" path="/api/v2/verify/:id/address-status" title="Get Address Status">
      <p class="docs-body-text">Query the address verification result separately.</p>
      <DocsCode label="Response" :code="addressStatusResponse" />
    </DocsEndpointCard>

    <DocsCode label="SDK usage" :code="addressSdk" />

    <DocsDivider />

    <!-- AML / SANCTIONS -->
    <DocsSectionAnchor id="aml" />
    <DocsH2 index="05">AML / Sanctions Screening</DocsH2>
    <DocsLead>
      Screen verified identities against global sanctions and watchlists (OFAC SDN, EU, UN). AML screening is an
      opt-in addon — enable it per session by passing
      <code class="docs-inline-code">addons.aml_screening: true</code> during initialization. The screening runs
      automatically as Gate 6, after all identity verification gates pass.
    </DocsLead>

    <DocsCallout type="note">
      AML screening uses OCR-extracted data (full name, date of birth, nationality) from the front document. It
      runs automatically after Gate 5 (face match) — no additional API call needed. Results appear in the
      <code>aml_screening</code> field of the status response.
    </DocsCallout>

    <a-card class="docs-table-card" :bordered="true">
      <div class="docs-info-label">How it works</div>
      <div class="docs-numbered-list">
        <template v-for="row in amlSteps" :key="row[0]">
          <span class="docs-list-number">{{ row[0] }}.</span>
          <span class="docs-list-text">{{ row[1] }}</span>
        </template>
      </div>
    </a-card>

    <a-card class="docs-table-card" :bordered="true">
      <div class="docs-info-label">Risk levels &amp; outcomes</div>
      <a-row :gutter="[12, 12]">
        <a-col v-for="row in amlRiskLevels" :key="row.level" :xs="24" :md="8">
          <div class="docs-risk-tile" :style="{ borderLeft: `3px solid ${row.color}` }">
            <div class="docs-risk-level" :style="{ color: row.color }">{{ row.level }}</div>
            <div class="docs-risk-score">score {{ row.score }}</div>
            <div class="docs-risk-outcome">{{ row.outcome }}</div>
          </div>
        </a-col>
      </a-row>
    </a-card>

    <DocsCode label="Enable AML screening" :code="amlEnable" />
    <DocsCode label="Status response — with AML results" :code="amlClearResponse" />
    <DocsCode label="Status response — potential match (manual review)" :code="amlMatchResponse" />

    <DocsCallout type="warning">
      When AML screening is not enabled (no <code>addons.aml_screening</code> flag), the
      <code>aml_screening</code> field will be <code>null</code> in the status response. AML screening is
      failure-safe — if the screening provider is unavailable, verification proceeds normally and the field is set
      to <code>null</code>.
    </DocsCallout>

    <DocsCode label="SDK usage" :code="amlSdk" />

    <DocsDivider />

    <!-- AGE ESTIMATION -->
    <DocsSectionAnchor id="age-estimation" />
    <DocsH2 index="06">Face Age Estimation</DocsH2>
    <DocsLead>
      Cross-checks apparent face age against declared date of birth. Extracts age estimates from both the
      document photo and the live capture, then flags discrepancies as a fraud signal (e.g., identity borrowing).
    </DocsLead>

    <DocsCode label="Status response — age_estimation object" :code="ageEstimation" />

    <DocsCallout type="note">
      Age estimation runs automatically during live capture. The <code>age_discrepancy</code> field shows the
      absolute difference between <code>live_face_age</code> and <code>declared_age</code>. High discrepancies
      contribute to the composite risk score.
    </DocsCallout>

    <DocsDivider />

    <!-- VELOCITY -->
    <DocsSectionAnchor id="velocity" />
    <DocsH2 index="07">Velocity / Fraud Detection</DocsH2>
    <DocsLead>
      Analyzes verification velocity during the live-capture step to detect fraud patterns such as rapid
      resubmissions, bot-like step timing, burst activity, and multi-IP abuse. Flagged sessions are routed to
      manual review.
    </DocsLead>

    <DocsCode label="Status response — velocity_analysis object" :code="velocityAnalysis" />

    <DocsCallout type="note">
      Velocity flags: <code>RAPID_IP_REUSE</code> (3+ verifications from same IP in 1 hour),
      <code>BOT_LIKE_TIMING</code> (steps completed too fast), <code>BURST_ACTIVITY</code> (sudden spike in
      verifications), <code>MULTI_IP_ABUSE</code> (same user from many IPs). Score range: 0–100, highest flag wins.
    </DocsCallout>

    <DocsDivider />

    <!-- IP GEOLOCATION -->
    <DocsSectionAnchor id="ip-geolocation" />
    <DocsH2 index="08">IP Geolocation Risk</DocsH2>
    <DocsLead>
      Detects geographic risk signals by comparing the client IP location against the document's issuing
      country. Identifies Tor exit nodes, datacenter/VPN IPs, and high-risk jurisdictions.
    </DocsLead>

    <DocsCode label="Status response — ip_geolocation object" :code="geoAnalysis" />

    <DocsCallout type="note">
      Geo risk flags: <code>COUNTRY_MISMATCH</code>, <code>TOR_EXIT_NODE</code>,
      <code>DATACENTER_IP</code>, <code>HIGH_RISK_JURISDICTION</code>. Score contributes to the composite risk
      score.
    </DocsCallout>

    <DocsDivider />

    <!-- VOICE AUTH -->
    <DocsSectionAnchor id="voice-auth" />
    <DocsH2 index="09">Voice Authentication</DocsH2>
    <DocsLead>
      Optional speaker verification step after face matching. Users speak randomly generated digits; the system
      verifies both the spoken content (ASR transcription) and the speaker identity (512-dimensional embedding
      comparison). Enable per-developer via the settings API.
    </DocsLead>

    <DocsCode label="Enable voice auth — PUT /api/developer/settings/voice-auth" :code="voiceEnable" />

    <DocsEndpointCard method="POST" path="/api/v2/verify/:id/voice-challenge" title="Request Voice Challenge">
      <p class="docs-body-text">
        Generates 6 random digits for the user to speak aloud. Challenge expires after 120 seconds. Session must be
        in <code class="docs-inline-code">AWAITING_VOICE</code> state.
      </p>
      <DocsCode label="Response — HTTP 200" :code="voiceChallengeResponse" />
    </DocsEndpointCard>

    <DocsEndpointCard method="POST" path="/api/v2/verify/:id/voice-capture" title="Submit Voice Capture">
      <p class="docs-body-text">
        Upload audio recording of the user speaking the challenge digits. Accepts WebM, WAV, or OGG formats.
        Returns voice match results and final verification status.
      </p>
      <DocsCode label="Response — HTTP 200" :code="voiceCaptureResponse" />
    </DocsEndpointCard>

    <DocsCallout type="note">
      Voice auth uses 512-dimensional speaker embeddings (CAM++ model) with cosine similarity matching.
      Thresholds: 0.55 (production), 0.50 (sandbox). All decisions are deterministic — no LLM in the gate path.
      Challenge expiry: 120 seconds. Rejection reasons: <code>VOICE_MATCH_FAILED</code>,
      <code>VOICE_CHALLENGE_FAILED</code>.
    </DocsCallout>

    <DocsDivider />

    <!-- COMPLIANCE -->
    <DocsSectionAnchor id="compliance" />
    <DocsH2 index="10">Compliance Rules</DocsH2>
    <DocsLead>
      Ship your compliance policy as code. Define rules that automatically adjust verification requirements based
      on country, document type, user age, risk score, and custom metadata. The engine evaluates them at session
      initialization — zero code changes needed.
    </DocsLead>

    <DocsCallout type="tip">
      Rules are evaluated server-side during <code>POST /api/v2/verify/initialize</code>. Your existing API
      integration works unchanged — the engine transparently adjusts the verification configuration based on
      matching rules.
    </DocsCallout>

    <a-card class="docs-table-card" :bordered="true">
      <div class="docs-info-label">How it works</div>
      <div class="docs-numbered-list">
        <template v-for="row in complianceSteps" :key="row[0]">
          <span class="docs-list-number">{{ row[0] }}.</span>
          <span class="docs-list-text">{{ row[1] }}</span>
        </template>
      </div>
    </a-card>

    <a-row :gutter="[12, 12]" class="docs-option-row">
      <a-col :xs="24" :md="12">
        <a-card :bordered="true" class="docs-analysis-card">
          <div class="docs-analysis-title" style="color: #0e7490">Condition Fields</div>
          <div v-for="field in conditionFields" :key="field.field" class="docs-field-item">
            <code class="docs-field-key">{{ field.field }}</code>
            <span>{{ field.desc }}</span>
          </div>
        </a-card>
      </a-col>
      <a-col :xs="24" :md="12">
        <a-card :bordered="true" class="docs-analysis-card">
          <div class="docs-analysis-title" style="color: #15803d">Action Fields</div>
          <div v-for="field in actionFields" :key="field.field" class="docs-field-item">
            <code class="docs-field-key">{{ field.field }}</code>
            <span>{{ field.desc }}</span>
          </div>
        </a-card>
      </a-col>
    </a-row>

    <DocsCallout type="note">
      When multiple rules match, actions merge additively. Verification mode escalates (never downgrades):
      <code>age_only &lt; document_only &lt; identity &lt; full</code>. Boolean flags like
      <code>force_manual_review</code> stick once set. Custom flags from all matching rules are collected and
      deduplicated.
    </DocsCallout>

    <DocsEndpointCard method="POST" path="/api/v2/compliance/rulesets" title="Create Ruleset">
      <p class="docs-body-text">
        Create a named group of compliance rules. Rulesets are evaluated in priority order (lower numbers first).
      </p>
      <div class="docs-field-list">
        <DocsFieldRow name="name" type="string" :req="true" desc="Human-readable ruleset name (max 200 chars)." />
        <DocsFieldRow name="description" type="string" desc="Optional description." />
        <DocsFieldRow name="is_active" type="boolean" desc="Only active rulesets are evaluated (default: true)." />
        <DocsFieldRow name="priority" type="integer" desc="Evaluation order — lower = first (default: 100)." />
      </div>
      <DocsCode label="Request" :code="rulesetRequest" />
      <DocsCode label="Response — HTTP 201" :code="rulesetResponse" />
    </DocsEndpointCard>

    <DocsEndpointCard method="POST" path="/api/v2/compliance/rulesets/:id/rules" title="Add Rule to Ruleset">
      <p class="docs-body-text">
        Add a rule with a structured condition and action. The engine validates both server-side — invalid
        conditions or actions return HTTP 400 with a descriptive error.
      </p>
      <div class="docs-field-list">
        <DocsFieldRow name="condition" type="object" :req="true" desc="Structured condition — leaf { field, op, value } or combinator { all, any, not }." />
        <DocsFieldRow name="action" type="object" :req="true" desc="Action to enforce when condition matches." />
        <DocsFieldRow name="description" type="string" desc="Human-readable rule description." />
      </div>
      <DocsCode label="Request — country + age rule" :code="ruleRequest" />
      <DocsCode label="Response — HTTP 201" :code="ruleResponse" />
    </DocsEndpointCard>

    <DocsEndpointCard method="POST" path="/api/v2/compliance/evaluate" title="Dry-Run Evaluate">
      <p class="docs-body-text">
        Test your rules against a synthetic context without creating a verification session. Returns which rules
        match and the resolved action after merging.
      </p>
      <div class="docs-field-list">
        <DocsFieldRow name="context" type="object" :req="true" desc="Test context with any combination of fields: country, document_type, user_age, verification_mode, risk_score, aml_risk_level, metadata." />
      </div>
      <DocsCode label="Request" :code="evaluateRequest" />
      <DocsCode label="Response" :code="evaluateResponse" />
    </DocsEndpointCard>

    <a-card class="docs-table-card" :bordered="true">
      <div class="docs-info-label">Operators</div>
      <div class="docs-operator-grid">
        <div v-for="row in operators" :key="row[0]" class="docs-operator-row">
          <code class="docs-operator-code">{{ row[0] }}</code>
          <span class="docs-operator-desc">{{ row[1] }}</span>
        </div>
      </div>
    </a-card>

    <a-card class="docs-table-card" :bordered="true">
      <div class="docs-info-label">Common patterns</div>
      <a-row :gutter="[12, 12]">
        <a-col v-for="pattern in commonPatterns" :key="pattern.title" :xs="24" :md="8">
          <div class="docs-risk-tile" :style="{ borderLeft: `3px solid ${pattern.color}` }">
            <div class="docs-risk-level" :style="{ color: pattern.color }">{{ pattern.title }}</div>
            <div class="docs-pattern-label">condition:</div>
            <pre class="docs-pattern-code">{{ pattern.cond }}</pre>
            <div class="docs-pattern-label">action:</div>
            <pre class="docs-pattern-code">{{ pattern.action }}</pre>
          </div>
        </a-col>
      </a-row>
    </a-card>

    <a-card class="docs-table-card" :bordered="true">
      <div class="docs-info-label">All compliance endpoints</div>
      <div v-for="endpoint in complianceEndpoints" :key="endpoint.method + endpoint.path" class="docs-endpoint-row">
        <span class="docs-method" :style="{ color: methodColor(endpoint.method) }">{{ endpoint.method }}</span>
        <code class="docs-endpoint-row-path">{{ endpoint.path }}</code>
        <span class="docs-endpoint-row-desc">{{ endpoint.desc }}</span>
      </div>
    </a-card>

    <DocsCode label="SDK usage" :code="complianceSdk" />

    <DocsCallout type="note">
      The Admin Dashboard (cloud edition) includes a visual drag-and-drop rule builder — create conditions, pick
      actions, and dry-run test without writing any code. Organization admins can open the
      <strong>Compliance Rules</strong> tab on the Verification Management page at
      <code>/admin/verifications</code>.
    </DocsCallout>

    <DocsDivider />

    <!-- MONITORING -->
    <DocsSectionAnchor id="monitoring" />
    <DocsH2 index="11">Monitoring &amp; Re-verification</DocsH2>
    <DocsLead>
      Schedule automatic re-verification reminders and track document expiry dates. Sends webhook notifications
      when documents are approaching expiry (90/60/30 days) or when re-verification is due.
    </DocsLead>

    <DocsEndpointCard method="POST" path="/api/v2/monitoring/schedules" title="Create Re-verification Schedule">
      <p class="docs-body-text">
        Schedule periodic re-verification for a user. When a schedule comes due, a
        <code class="docs-inline-code">reverification.due</code> webhook event fires.
      </p>
      <div class="docs-field-list">
        <DocsFieldRow name="user_id" type="UUID" :req="true" desc="User to schedule re-verification for." />
        <DocsFieldRow name="interval_days" type="number" :req="true" desc="Days between re-verifications (30–730)." />
        <DocsFieldRow name="verification_request_id" type="UUID" desc="Link to a specific verification." />
      </div>
      <DocsCode label="Request" :code="scheduleRequest" />
      <DocsCode label="Response — HTTP 201" :code="scheduleResponse" />
    </DocsEndpointCard>

    <DocsEndpointCard method="GET" path="/api/v2/monitoring/expiring-documents" title="Get Expiring Documents">
      <p class="docs-body-text">
        List identity documents approaching expiry across your verified users. Alerts are generated at 90, 60, and
        30-day thresholds.
      </p>
      <div class="docs-field-list">
        <DocsFieldRow name="days_ahead" type="number" desc="Look-ahead window in days (default: 90)." />
        <DocsFieldRow name="page" type="number" desc="Page number for pagination." />
        <DocsFieldRow name="limit" type="number" desc="Items per page (default: 20)." />
      </div>
      <DocsCode label="Response" :code="expiringResponse" />
    </DocsEndpointCard>

    <DocsCode label="SDK usage" :code="monitoringSdk" />

    <DocsCallout type="note">
      Document expiry detection is based on the <code>expiration_date</code> field extracted during OCR. Webhook
      events (<code>document.expiring</code>, <code>document.expired</code>, <code>reverification.due</code>) are
      delivered to your registered webhook URLs.
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
  DocsEndpointCard,
  DocsStatusPill,
  docsApiUrl,
  type DocsNavItem,
} from '../components/DocsShell.vue'

const apiUrl = docsApiUrl()

const NAV: DocsNavItem[] = [
  { id: 'analysis', label: 'Analysis Engine', depth: 0 },
  { id: 'statuses', label: 'Statuses', depth: 0 },
  { id: 'batch', label: 'Batch API', depth: 0 },
  { id: 'address', label: 'Address Verification', depth: 0 },
  { id: 'aml', label: 'AML / Sanctions', depth: 0 },
  { id: 'age-estimation', label: 'Age Estimation', depth: 0 },
  { id: 'velocity', label: 'Velocity Checks', depth: 0 },
  { id: 'ip-geolocation', label: 'IP Geolocation', depth: 0 },
  { id: 'voice-auth', label: 'Voice Auth', depth: 0 },
  { id: 'compliance', label: 'Compliance Rules', depth: 0 },
  { id: 'monitoring', label: 'Monitoring', depth: 0 },
]

const analysisColumns = [
  {
    title: 'OCR Extraction',
    color: '#0e7490',
    items: ['PaddleOCR / Tesseract engines', 'Name, DOB, ID number', 'Expiry & issue date', 'AAMVA field parsing (US DLs)', 'State DL format validation', 'Per-field confidence scores'],
  },
  {
    title: 'Document Quality',
    color: '#1d4ed8',
    items: ['Sobel edge blur detection', 'Brightness & contrast stats', 'Resolution check (≥ 800×600)', 'File size validation', 'Overall quality score', 'Auto-reject below threshold'],
  },
  {
    title: 'Cross-Validation',
    color: '#b45309',
    items: ['PDF417 / QR barcode decode', 'Levenshtein distance matching', 'Token-set name similarity', 'Front OCR ↔ back barcode check', 'Date & ID number consistency', 'Weighted field scoring'],
  },
  {
    title: 'Liveness & Face Match',
    color: '#15803d',
    items: ['EXIF metadata analysis (20%)', 'JPEG artifact detection (15%)', 'Color histogram analysis (15%)', 'Byte entropy scoring (12%)', 'Pixel variance & edge density', 'Face detection (SSDMobilenetv1)', '128-d face embeddings', 'Cosine similarity scoring'],
  },
  {
    title: 'Voice Authentication',
    color: '#7c3aed',
    items: ['Random digit challenge (6 digits)', 'ASR transcription verification', '512-d speaker embeddings (CAM++)', 'Cosine similarity matching', 'Configurable threshold (0.55)', 'Optional per-developer toggle'],
  },
]

const statuses = [
  { status: 'AWAITING_FRONT', color: '#b45309', desc: 'Session initialized. Waiting for front document upload.', terminal: false },
  { status: 'AWAITING_BACK', color: '#b45309', desc: 'Front document processed. Waiting for back document upload.', terminal: false },
  { status: 'CROSS_VALIDATING', color: '#1d4ed8', desc: 'Running cross-validation between front OCR and back barcode data.', terminal: false },
  { status: 'AWAITING_LIVE', color: '#b45309', desc: 'Cross-validation passed. Waiting for selfie upload.', terminal: false },
  { status: 'FACE_MATCHING', color: '#1d4ed8', desc: 'Running liveness detection and face matching against document photo.', terminal: false },
  { status: 'COMPLETE', color: '#15803d', desc: 'All gates passed. Check final_result for: "verified", "manual_review", or "failed".', terminal: true },
  { status: 'HARD_REJECTED', color: '#b91c1c', desc: 'Session rejected by a quality gate. Check rejection_reason for details. Subsequent steps return 409.', terminal: true },
]

const amlSteps = [
  ['1', 'Pass addons.aml_screening: true in the initialize request'],
  ['2', 'Complete identity verification (front doc → back doc → live capture)'],
  ['3', 'Gate 6 automatically screens extracted name/DOB against sanctions lists'],
  ['4', 'Results appear in aml_screening field of the status response'],
]

const amlRiskLevels = [
  { level: 'clear', score: '< 0.5', outcome: 'Verification proceeds normally', color: '#15803d' },
  { level: 'potential_match', score: '0.5 – 0.84', outcome: 'Routed to manual_review', color: '#b45309' },
  { level: 'confirmed_match', score: '>= 0.85', outcome: 'Hard reject (failed)', color: '#b91c1c' },
]

const complianceSteps = [
  ['1', 'Create rulesets with conditions and actions via API or the Admin Dashboard'],
  ['2', 'User starts verification — engine loads active rules for the developer'],
  ['3', "Each rule's condition is tested against the session context (country, age, metadata)"],
  ['4', 'Matching actions merge (most restrictive wins) and adjust the verification pipeline'],
]

const conditionFields = [
  { field: 'country', desc: 'ISO country code' },
  { field: 'document_type', desc: 'passport, drivers_license, national_id' },
  { field: 'user_age', desc: 'Calculated from DOB' },
  { field: 'risk_score', desc: 'Composite score (0.0 – 1.0)' },
  { field: 'aml_risk_level', desc: 'clear, potential_match, confirmed_match' },
  { field: 'metadata.*', desc: 'Developer-supplied key-value data' },
]

const actionFields = [
  { field: 'set_mode', desc: 'Override to age_only / document_only / identity / full' },
  { field: 'require_address', desc: 'Enable address verification step' },
  { field: 'require_liveness', desc: 'Set to passive or head_turn' },
  { field: 'require_aml', desc: 'Force AML/sanctions screening' },
  { field: 'force_manual_review', desc: 'Route to manual review' },
  { field: 'set_flag', desc: 'Attach a custom flag (e.g. high_risk)' },
]

const operators = [
  ['eq / neq', 'Equals / not equals'],
  ['in / not_in', 'Value in / not in array'],
  ['gt / gte / lt / lte', 'Numeric comparisons'],
  ['exists', 'Field present (true) or absent (false)'],
  ['contains', 'String contains substring'],
  ['all / any / not', 'Combinators — nest conditions'],
]

const commonPatterns = [
  { title: 'Country restriction', color: '#0e7490', cond: '{ "field": "country", "op": "in", "value": ["DE", "FR"] }', action: '{ "set_mode": "full", "require_aml": true }' },
  { title: 'Age gate', color: '#b45309', cond: '{ "field": "user_age", "op": "lt", "value": 18 }', action: '{ "force_manual_review": true }' },
  { title: 'Risk escalation', color: '#b91c1c', cond: '{ "field": "risk_score", "op": "gte", "value": 0.7 }', action: '{ "require_aml": true, "set_flag": "high_risk" }' },
]

const complianceEndpoints = [
  { method: 'POST', path: '/api/v2/compliance/rulesets', desc: 'Create ruleset' },
  { method: 'GET', path: '/api/v2/compliance/rulesets', desc: 'List rulesets (includes rule_count)' },
  { method: 'GET', path: '/api/v2/compliance/rulesets/:id', desc: 'Get ruleset with all rules' },
  { method: 'PUT', path: '/api/v2/compliance/rulesets/:id', desc: 'Update ruleset metadata' },
  { method: 'DELETE', path: '/api/v2/compliance/rulesets/:id', desc: 'Delete ruleset + all rules (CASCADE)' },
  { method: 'POST', path: '/api/v2/compliance/rulesets/:id/rules', desc: 'Add rule to ruleset' },
  { method: 'PUT', path: '/api/v2/compliance/rules/:id', desc: 'Update rule condition/action' },
  { method: 'DELETE', path: '/api/v2/compliance/rules/:id', desc: 'Delete rule' },
  { method: 'POST', path: '/api/v2/compliance/evaluate', desc: 'Dry-run evaluation' },
]

const METHOD_COLORS: Record<string, string> = { GET: '#15803d', POST: '#1d4ed8', PUT: '#b45309', DELETE: '#b91c1c' }
const methodColor = (method: string): string => METHOD_COLORS[method] ?? '#64748b'

const batchRequest = `curl -X POST ${apiUrl}/api/v2/batch/upload \\
  -H "X-API-Key: your-key" \\
  -H "Content-Type: application/json" \\
  -d '{
    "items": [
      { "user_id": "user-001", "document_type": "drivers_license" },
      { "user_id": "user-002", "document_type": "passport" },
      { "user_id": "user-003", "document_type": "national_id" }
    ]
  }'`

const batchCreateResponse = `{
  "success": true,
  "batch_id": "batch_abc123",
  "status": "pending",
  "total_items": 3,
  "message": "Batch job created"
}`

const batchStatusResponse = `{
  "batch_id": "batch_abc123",
  "status": "processing",
  "total_items": 3,
  "processed_items": 1,
  "succeeded_items": 1,
  "failed_items": 0,
  "progress_percentage": 33,
  "created_at": "2026-03-15T10:00:00Z",
  "completed_at": null
}`

const batchResultsResponse = `{
  "results": [
    { "item_id": "item_001", "user_id": "user-001", "status": "completed", "verification_id": "v_abc", "error": null },
    { "item_id": "item_002", "user_id": "user-002", "status": "completed", "verification_id": "v_def", "error": null },
    { "item_id": "item_003", "user_id": "user-003", "status": "failed", "verification_id": null, "error": "Invalid document URL" }
  ]
}`

const batchSdk = `// Create batch
const batch = await sdk.createBatch([
  { user_id: 'user-001', document_type: 'drivers_license' },
  { user_id: 'user-002', document_type: 'passport' },
]);

// Poll progress
const status = await sdk.getBatchStatus(batch.batch_id);
console.log(status.progress_percentage + '% complete');

// Get results when done
const results = await sdk.getBatchResults(batch.batch_id);

// Cancel if needed
await sdk.cancelBatch(batch.batch_id);`

const addressRequest = `curl -X POST ${apiUrl}/api/v2/verify/VERIFICATION_ID/address-document \\
  -H "X-API-Key: your-key" \\
  -F "document=@utility_bill.jpg" \\
  -F "document_type=utility_bill"`

const addressResponse = `{
  "success": true,
  "verification_id": "v_abc123",
  "address_verification": {
    "status": "pass",
    "score": 0.92,
    "name_match_score": 0.95,
    "address": "123 Main St, Apt 4B, New York, NY 10001",
    "document_type": "utility_bill",
    "document_fresh": true,
    "reasons": []
  }
}`

const addressStatusResponse = `{
  "verification_id": "v_abc123",
  "address_verification": {
    "status": "pass",          // "pass" | "review" | "reject"
    "score": 0.92,
    "name_match_score": 0.95,
    "address": "123 Main St, New York, NY 10001",
    "document_type": "utility_bill",
    "document_fresh": true,
    "reasons": []
  }
}`

const addressSdk = `// Upload proof-of-address document
const result = await sdk.uploadAddressDocument(
  verificationId,
  documentBuffer,     // Buffer or Blob
  'utility_bill',     // document type
  'electric-bill.pdf' // optional filename
);

console.log(result.address_verification.status);          // 'pass'
console.log(result.address_verification.address);          // extracted address
console.log(result.address_verification.name_match_score); // 0.95

// Query status separately
const status = await sdk.getAddressStatus(verificationId);`
const amlEnable = `curl -X POST ${apiUrl}/api/v2/verify/initialize \\
  -H "X-API-Key: your-key" \\
  -H "Content-Type: application/json" \\
  -d '{
    "user_id": "user-uuid",
    "addons": { "aml_screening": true }
  }'`

const amlClearResponse = `{
  "verification_id": "v_abc123",
  "status": "COMPLETE",
  "final_result": "verified",

  "aml_screening": {
    "risk_level": "clear",             // "clear" | "potential_match" | "confirmed_match"
    "match_found": false,
    "match_count": 0,
    "lists_checked": [
      "us_ofac_sdn",
      "eu_sanctions",
      "un_sanctions"
    ],
    "screened_at": "2026-03-19T10:05:28Z"
  }
}`

const amlMatchResponse = `{
  "verification_id": "v_abc123",
  "status": "COMPLETE",
  "final_result": "manual_review",
  "manual_review_reason": "AML_POTENTIAL_MATCH",

  "aml_screening": {
    "risk_level": "potential_match",
    "match_found": true,
    "match_count": 1,
    "lists_checked": ["us_ofac_sdn", "eu_sanctions", "un_sanctions"],
    "screened_at": "2026-03-19T10:05:28Z"
  }
}`

const amlSdk = `// Initialize with AML screening enabled
const session = await sdk.initialize({
  userId: 'user-uuid',
  addons: { aml_screening: true },
});

// ... complete verification steps ...

// Check AML results in final status
const status = await sdk.getStatus(session.verification_id);

if (status.aml_screening?.risk_level === 'clear') {
  console.log('No sanctions matches found');
} else if (status.aml_screening?.risk_level === 'potential_match') {
  console.log('Manual review required — potential sanctions match');
} else if (status.aml_screening?.risk_level === 'confirmed_match') {
  console.log('Verification failed — confirmed sanctions match');
}`

const ageEstimation = `{
  "age_estimation": {
    "document_face_age": 34,
    "live_face_age": 22,
    "declared_age": 35,
    "age_discrepancy": 13,
    "flag": "Age discrepancy between live capture and declared DOB"
  }
}`

const velocityAnalysis = `{
  "velocity_analysis": {
    "ip_address_hash": "a1b2c3...",
    "ip_reuse_count_1h": 3,
    "ip_reuse_count_24h": 8,
    "avg_step_duration_ms": 1200,
    "flags": ["RAPID_IP_REUSE", "BOT_LIKE_TIMING"],
    "score": 75
  }
}`

const geoAnalysis = `{
  "ip_geolocation": {
    "ip_country": "RO",
    "document_country": "US",
    "country_match": false,
    "is_tor": false,
    "is_datacenter": true,
    "geo_risk_flags": ["COUNTRY_MISMATCH", "DATACENTER_IP"],
    "geo_risk_score": 60
  }
}`

const voiceEnable = `curl -X PUT ${apiUrl}/api/developer/settings/voice-auth \\
  -H "X-API-Key: your-key" \\
  -H "Content-Type: application/json" \\
  -d '{ "enabled": true }'

// Response: { "success": true, "voice_auth_enabled": true }`

const voiceChallengeResponse = `{
  "success": true,
  "challenge_digits": "3 7 1 9 0 5",
  "expires_in": 120
}`

const voiceCaptureResponse = `{
  "success": true,
  "verification_id": "v_abc123",
  "status": "COMPLETE",
  "voice_match_results": {
    "similarity_score": 0.82,
    "passed": true,
    "threshold_used": 0.55,
    "challenge_verified": true,
    "challenge_digits": "3 7 1 9 0 5"
  }
}`
const rulesetRequest = `curl -X POST ${apiUrl}/api/v2/compliance/rulesets \\
  -H "X-API-Key: your-key" \\
  -H "Content-Type: application/json" \\
  -d '{
    "name": "EU Compliance",
    "description": "Enhanced verification for EU member states",
    "priority": 10
  }'`

const rulesetResponse = `{
  "success": true,
  "ruleset": {
    "id": "550e8400-e29b-41d4-a716-446655440001",
    "name": "EU Compliance",
    "description": "Enhanced verification for EU member states",
    "is_active": true,
    "priority": 10,
    "created_at": "2026-04-07T12:00:00Z"
  }
}`

const ruleRequest = `curl -X POST ${apiUrl}/api/v2/compliance/rulesets/RULESET_ID/rules \\
  -H "X-API-Key: your-key" \\
  -H "Content-Type: application/json" \\
  -d '{
    "condition": {
      "all": [
        { "field": "country", "op": "in", "value": ["DE", "FR", "IT", "ES"] },
        { "field": "user_age", "op": "lt", "value": 18 }
      ]
    },
    "action": {
      "set_mode": "full",
      "require_aml": true,
      "force_manual_review": true,
      "set_flag": "eu_minor"
    },
    "description": "Full verification + manual review for EU minors"
  }'`

const ruleResponse = `{
  "success": true,
  "rule": {
    "id": "660e8400-e29b-41d4-a716-446655440002",
    "condition": { "all": [{ "field": "country", "op": "in", "value": ["DE", "FR", "IT", "ES"] }, { "field": "user_age", "op": "lt", "value": 18 }] },
    "action": { "set_mode": "full", "require_aml": true, "force_manual_review": true, "set_flag": "eu_minor" },
    "description": "Full verification + manual review for EU minors",
    "created_at": "2026-04-07T12:01:00Z"
  }
}`

const evaluateRequest = `curl -X POST ${apiUrl}/api/v2/compliance/evaluate \\
  -H "X-API-Key: your-key" \\
  -H "Content-Type: application/json" \\
  -d '{
    "context": {
      "country": "DE",
      "user_age": 16,
      "document_type": "passport"
    }
  }'`

const evaluateResponse = `{
  "success": true,
  "matched_rules": 1,
  "matches": [
    {
      "ruleset": "EU Compliance",
      "rule": "Full verification + manual review for EU minors",
      "action": { "set_mode": "full", "require_aml": true, "force_manual_review": true, "set_flag": "eu_minor" }
    }
  ],
  "resolved_action": {
    "set_mode": "full",
    "require_aml": true,
    "force_manual_review": true,
    "flags": ["eu_minor"]
  }
}`

const complianceSdk = `// Create a ruleset
const { ruleset } = await fetch('/api/v2/compliance/rulesets', {
  method: 'POST',
  headers: { 'X-API-Key': apiKey, 'Content-Type': 'application/json' },
  body: JSON.stringify({ name: 'EU Compliance', priority: 10 }),
}).then(r => r.json());

// Add a rule
await fetch(\`/api/v2/compliance/rulesets/\${ruleset.id}/rules\`, {
  method: 'POST',
  headers: { 'X-API-Key': apiKey, 'Content-Type': 'application/json' },
  body: JSON.stringify({
    condition: { "field": "country", "op": "in", "value": ["DE", "FR", "IT"] },
    action: { "set_mode": "full", "require_aml": true },
    description: "Full verification for EU countries",
  }),
}).then(r => r.json());

// Dry-run test
const result = await fetch('/api/v2/compliance/evaluate', {
  method: 'POST',
  headers: { 'X-API-Key': apiKey, 'Content-Type': 'application/json' },
  body: JSON.stringify({ context: { country: 'DE', user_age: 25 } }),
}).then(r => r.json());

console.log(result.matched_rules, 'rules matched');
console.log(result.resolved_action);  // { set_mode: 'full', require_aml: true }`

const scheduleRequest = `curl -X POST ${apiUrl}/api/v2/monitoring/schedules \\
  -H "X-API-Key: your-key" \\
  -H "Content-Type: application/json" \\
  -d '{ "user_id": "user-uuid", "interval_days": 365 }'`

const scheduleResponse = `{
  "success": true,
  "schedule": {
    "id": "sched_abc123",
    "user_id": "user-uuid",
    "interval_days": 365,
    "next_verification_at": "2027-03-15T00:00:00Z",
    "last_verification_at": null,
    "status": "active",
    "created_at": "2026-03-15T10:00:00Z"
  }
}`

const expiringResponse = `{
  "alerts": [
    {
      "id": "alert_001",
      "verification_request_id": "v_abc",
      "user_id": "user-001",
      "expiry_date": "2026-06-15",
      "alert_type": "90_day",
      "webhook_sent": true,
      "created_at": "2026-03-15T00:00:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "limit": 20
}`

const monitoringSdk = `// Create re-verification schedule (annual)
const schedule = await sdk.createMonitoringSchedule({
  user_id: 'user-uuid',
  interval_days: 365,
});

// List active schedules
const schedules = await sdk.listMonitoringSchedules({ status: 'active' });

// Get documents expiring in the next 60 days
const expiring = await sdk.getExpiringDocuments({ days_ahead: 60 });
console.log(expiring.alerts.length, 'documents expiring soon');

// Cancel a schedule
await sdk.cancelMonitoringSchedule(schedule.schedule.id);`
</script>

<style>
.docs-success { color: #15803d; }
.docs-analysis-card .arco-card-body { padding: 16px 18px; }
.docs-analysis-title { font-family: ui-monospace, Menlo, monospace; font-size: 12.5px; font-weight: 600; margin-bottom: 10px; }
.docs-status-line { display: flex; gap: 16px; padding: 12px 0; border-top: 1px solid #eef2f7; align-items: flex-start; flex-wrap: wrap; }
.docs-status-line:first-child { border-top: none; }
.docs-status-line-pill { width: 160px; flex-shrink: 0; display: flex; gap: 8px; align-items: center; }
.docs-terminal { font-family: ui-monospace, Menlo, monospace; font-size: 9.5px; letter-spacing: 0.06em; }
.docs-status-line-desc { font-size: 13px; color: #475569; flex: 1; line-height: 1.6; }
.docs-risk-tile { background: #f8fafc; padding: 12px; height: 100%; }
.docs-risk-level { font-family: ui-monospace, Menlo, monospace; font-size: 12px; font-weight: 600; margin-bottom: 4px; }
.docs-risk-score { font-family: ui-monospace, Menlo, monospace; font-size: 11px; color: #94a3b8; margin-bottom: 6px; }
.docs-risk-outcome { font-size: 12px; color: #475569; line-height: 1.5; }
.docs-field-item { display: flex; gap: 8px; padding: 3px 0; font-size: 12px; color: #64748b; }
.docs-field-key { font-family: ui-monospace, Menlo, monospace; font-size: 11.5px; color: #0f172a; min-width: 120px; }
.docs-operator-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 4px 24px; }
.docs-operator-row { display: flex; gap: 8px; padding: 4px 0; }
.docs-operator-code { font-family: ui-monospace, Menlo, monospace; font-size: 11.5px; color: #0e7490; min-width: 130px; flex-shrink: 0; }
.docs-operator-desc { font-size: 12px; color: #64748b; }
.docs-pattern-label { font-family: ui-monospace, Menlo, monospace; font-size: 10.5px; color: #94a3b8; margin: 4px 0 2px; }
.docs-pattern-code {
  margin: 0 0 4px;
  font-family: ui-monospace, Menlo, monospace;
  font-size: 10.5px;
  color: #64748b;
  white-space: pre-wrap;
  word-break: break-all;
}
.docs-endpoint-row { display: flex; align-items: center; gap: 10px; padding: 5px 0; }
.docs-method { font-family: ui-monospace, Menlo, monospace; font-size: 11px; font-weight: 600; width: 52px; text-align: right; flex-shrink: 0; }
.docs-endpoint-row-path { font-family: ui-monospace, Menlo, monospace; font-size: 11.5px; color: #0f172a; flex: 1; }
.docs-endpoint-row-desc { font-size: 12px; color: #94a3b8; }
@media (max-width: 767px) {
  .docs-operator-grid { grid-template-columns: 1fr; }
}
</style>
