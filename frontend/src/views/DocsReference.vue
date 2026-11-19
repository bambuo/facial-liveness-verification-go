<template>
  <DocsShell slug="docs/reference" :nav="NAV">
    <!-- RATE LIMITS -->
    <DocsSectionAnchor id="rate-limits" />
    <DocsH2 index="01">Rate Limits & Status Codes</DocsH2>

    <a-row :gutter="[16, 16]">
      <a-col :xs="24" :md="12">
        <a-card class="docs-table-card" :bordered="true">
          <div class="docs-info-label">Rate Limits</div>
          <div v-for="row in rateLimits" :key="row.label" class="docs-rate-row">
            <div>
              <div class="docs-rate-label">{{ row.label }}</div>
              <div class="docs-rate-note">{{ row.note }}</div>
            </div>
            <code class="docs-rate-value">{{ row.value }}</code>
          </div>
        </a-card>
      </a-col>
      <a-col :xs="24" :md="12">
        <a-card class="docs-table-card" :bordered="true">
          <div class="docs-info-label">HTTP Status Codes</div>
          <div v-for="row in statusCodes" :key="row.code" class="docs-status-row">
            <code class="docs-status-code" :style="{ color: row.color }">{{ row.code }}</code>
            <span class="docs-status-desc">{{ row.desc }}</span>
          </div>
        </a-card>
      </a-col>
    </a-row>

    <DocsDivider />

    <!-- CHANGELOG -->
    <DocsSectionAnchor id="changelog" />
    <DocsH2 index="02">Changelog</DocsH2>
    <DocsLead>
      Notable changes to the Identity Verification API. This project uses
      <a href="https://semver.org" target="_blank" rel="noopener noreferrer">Semantic Versioning</a>.
    </DocsLead>

    <a-card v-for="release in releases" :key="release.version" class="docs-release" :bordered="true">
      <template #title>
        <div class="docs-release-head">
          <a-tag size="small" color="cyan">v{{ release.version }}</a-tag>
          <span class="docs-release-date">{{ release.date }}</span>
          <a-tag v-if="release.version === '1.8.3'" size="small" color="green">latest</a-tag>
        </div>
      </template>

      <div v-if="release.added.length" class="docs-release-group">
        <div class="docs-release-label docs-release-added"><span class="docs-dot" />Added</div>
        <div v-for="item in release.added" :key="item" class="docs-release-item docs-release-item-added">{{ item }}</div>
      </div>
      <div v-if="release.fixed.length" class="docs-release-group">
        <div class="docs-release-label docs-release-fixed"><span class="docs-dot" />Fixed</div>
        <div v-for="item in release.fixed" :key="item" class="docs-release-item docs-release-item-fixed">{{ item }}</div>
      </div>
      <div v-if="release.changed.length" class="docs-release-group">
        <div class="docs-release-label docs-release-changed"><span class="docs-dot" />Changed</div>
        <div v-for="item in release.changed" :key="item" class="docs-release-item docs-release-item-changed">{{ item }}</div>
      </div>
    </a-card>

    <DocsDivider />

    <!-- SUPPORT -->
    <DocsSectionAnchor id="support" />
    <DocsH2 index="03">Support &amp; Resources</DocsH2>

    <a-row :gutter="[12, 12]">
      <a-col v-for="resource in resources" :key="resource.title" :xs="24" :sm="12" :md="8">
        <a-card class="docs-resource" :bordered="true">
          <div class="docs-resource-icon">{{ resource.icon }}</div>
          <div class="docs-resource-title">{{ resource.title }}</div>
          <div class="docs-resource-desc">{{ resource.desc }}</div>
          <a-link
            :href="resource.href"
            :target="resource.href.startsWith('http') || resource.href.startsWith('mailto') ? '_blank' : undefined"
            class="docs-resource-cta"
          >{{ resource.cta }}</a-link>
        </a-card>
      </a-col>
    </a-row>
  </DocsShell>
</template>

<script setup lang="ts">
import DocsShell, {
  DocsSectionAnchor,
  DocsH2,
  DocsLead,
  DocsDivider,
  type DocsNavItem,
} from '../components/DocsShell.vue'

const NAV: DocsNavItem[] = [
  { id: 'rate-limits', label: 'Rate Limits', depth: 0 },
  { id: 'changelog', label: 'Changelog', depth: 0 },
  { id: 'support', label: 'Support', depth: 0 },
]

const rateLimits = [
  { label: 'Per developer key', value: '1,000 req / hour', note: 'cloud edition — sandbox + production combined' },
  { label: 'Per user', value: '5 verifications / hour', note: 'cloud edition — across all developer keys' },
  { label: 'Self-hosted', value: 'None', note: 'rate limiting disabled by default' },
  { label: 'Enterprise', value: 'Custom', note: 'contact sales' },
]

const statusCodes = [
  { code: '200 / 201', color: '#15803d', desc: 'Success' },
  { code: '400', color: '#b45309', desc: 'Bad request — validation error' },
  { code: '401', color: '#b91c1c', desc: 'Unauthorized — invalid or missing API key' },
  { code: '404', color: '#b91c1c', desc: 'Verification not found' },
  { code: '409', color: '#c2410c', desc: 'Conflict — session hard-rejected by a gate' },
  { code: '429', color: '#c2410c', desc: 'Rate limit exceeded' },
  { code: '500', color: '#b91c1c', desc: 'Internal server error' },
]

const releases = [
  {
    version: '1.8.3',
    date: '2026-04-09',
    added: [],
    fixed: [
      'Compliance auth recursion bug — authenticateComplianceRequest self-recursed on the X-API-Key branch, which would have stack-overflowed any API key request (latent because the previous UI only ever hit the JWT branch)',
    ],
    changed: [
      'Compliance ruleset UI moved from Developer Portal to Admin Dashboard (Verification Management → Compliance Rules tab, gated to organization admins on cloud edition)',
      'Compliance endpoints now accept exactly two auth paths: X-API-Key (developer SDK/automation) OR organization-admin reviewer session cookie. Regular reviewers and platform admins are rejected',
      'CSRF enforced on /api/v2/compliance routes (no-ops for X-API-Key callers, enforced for cookie-auth)',
    ],
  },
  {
    version: '1.8.2',
    date: '2026-04-02',
    added: [
      'Verification page branding — white-label the hosted verification page with a custom logo, accent color, and company name',
      'GET/PUT /api/developer/settings/branding — configure branding settings',
      'POST /api/developer/branding/logo — upload branding logo (JPEG/PNG, max 2 MB)',
      'GET /api/v2/verify/page-config?api_key=... — public endpoint returning developer branding (cached 5 min)',
      'Live preview panel in Developer Portal Settings modal',
    ],
    fixed: [],
    changed: [],
  },
  {
    version: '1.8.1',
    date: '2026-04-02',
    added: [
      'Custom verification flows — verification_mode now supports document_only (front → back → cross-validation) and identity (front → liveness → face match) presets',
      'Endpoint guards: back-document returns 400 for identity/age_only flows; live-capture returns 400 for document_only/age_only flows',
    ],
    fixed: [],
    changed: [],
  },
  {
    version: '1.8.0',
    date: '2026-04-02',
    added: [
      'Role-based access control — verification_reviewers table now has a role column (admin or reviewer), enabling Organization Admins with elevated privileges',
      'Organization Admin role — access analytics, GDPR data deletion, and override verification decisions scoped to their developer',
      'Role-aware reviewer invitations: POST /api/developer/reviewers/invite accepts optional role parameter',
      'requireOrgAdminOrPlatformAdmin middleware — gates analytics, GDPR delete, and override endpoints',
      'Team setup banner + role badges in Developer Portal Settings',
    ],
    fixed: [],
    changed: [
      'Developer escalation removed — POST /api/auth/admin/escalate now returns 410 Gone',
      'Analytics endpoints opened to org admins (scoped by developer_id)',
      'Verification override restricted to org admins and platform admins (regular reviewers get 403)',
      'GET /api/admin/developers restricted to platform admins only',
    ],
  },
  {
    version: '1.7.2',
    date: '2026-03-30',
    added: [
      'AML screening auto-trigger — runs automatically on all non-sandbox verifications when AML_PROVIDER is configured (no longer requires per-session addons.aml_screening)',
      'Multi-provider AML screening — AML_PROVIDER supports comma-separated providers (e.g. opensanctions,offline), matches deduplicated, highest risk level wins',
      'AML result persistence in aml_screenings table with full match details (listed name, list source, score, match type)',
      'Address cross-validation — front OCR address compared against back barcode address as a supplementary signal',
      'Developer aml_enabled column (opt-out)',
    ],
    fixed: [],
    changed: ['Risk scoring weights rebalanced to include aml_screening factor (weight 0.10)'],
  },
  {
    version: '1.7.0',
    date: '2026-03-27',
    added: [
      'Reviewer invitation system — developers invite external reviewers with OTP-based access scoped to their verifications',
      'Reviewer auth: POST /api/auth/reviewer/otp/send, POST /api/auth/reviewer/otp/verify',
      'Reviewer management: POST /api/developer/reviewers/invite, GET /api/developer/reviewers, DELETE /api/developer/reviewers/:id',
      'Reviewer management UI in Developer Portal settings (invite, list, revoke, copy login link)',
    ],
    fixed: [],
    changed: [
      'Admin verification endpoints scope queries by developer_id for reviewer tokens',
      'Reviewers cannot use override decision (admin-only)',
    ],
  },
  {
    version: '1.6.0',
    date: '2026-03-26',
    added: [
      'Batch verification processing — full pipeline (OCR, cross-validation, quality gates) without live capture',
      'Admin status override with new_status field on PUT /api/admin/verification/:id/review',
      'Webhook forwarding on admin review actions (approve, reject, override)',
      'Verification Management page at /admin/verifications — stats, filterable table, detail view, review actions',
    ],
    fixed: [],
    changed: ['Batch items that fail quality gates correctly marked as failed with rejection reason'],
  },
  {
    version: '1.5.0',
    date: '2026-03-24',
    added: [
      'Community edition first-run setup wizard (auto-detects zero-developer state)',
      'Mobile handoff link endpoint: PATCH /api/verify/handoff/:token/link',
      'Engine Worker microservice — standalone container for ML verification (OCR, face detection, liveness, deepfake)',
    ],
    fixed: ['Mobile handoff desktop notification — exponential backoff retry, 30-min session timeout'],
    changed: [
      'Core API image reduced from ~2GB to ~250MB — ML dependencies isolated in engine container (~1.5GB)',
      'Docker Compose architecture: postgres + engine + api + frontend (4 containers)',
    ],
  },
  {
    version: '1.2.0',
    date: '2026-03-20',
    added: [],
    fixed: [],
    changed: [
      'Liveness system: removed dead MediaPipe code path, renamed MultiFrame → HeadTurn',
      'Malformed liveness_metadata now returns HTTP 400 (VALIDATION_ERROR) instead of silently falling back to passive mode',
      'Removed legacy multi_frame_color challenge type alias — only head_turn is accepted',
      'color_sequence field is now optional (clients no longer need to send it)',
    ],
  },
  {
    version: '1.1.0',
    date: '2026-03-19',
    added: [
      'Visual authenticity checks — FFT analysis, color distribution, zone validation, deepfake detection',
      'Webhook resend endpoint (POST /api/developer/webhooks/:id/deliveries/:did/resend)',
      'Per-API-key scoping for webhook endpoints',
      'Developer-configurable LLM fallback for OCR with date disambiguation',
      'Account deletion endpoint (DELETE /api/developer/account)',
      'Email OTP + GitHub OAuth authentication (replaced insecure password login)',
      'Webhook delivery logs endpoint (GET /api/developer/webhooks/:id/deliveries)',
      'Webhook test endpoint with timeout handling',
      'AML/sanctions screening opt-in addon',
      "US driver's license format validator",
      '/health endpoint for Railway health checks + /api/health for API consumers',
    ],
    fixed: [
      'NULL events column silently filtering out webhook deliveries',
      'Per-provider metrics now derived from results JSONB instead of session-level aggregates',
      'Missing webhook_deliveries table migration',
      'OTP security hardening — atomic verify, timing-safe comparison, fail-closed',
      'Trailing AAMVA field markers stripped from space-separated DLN',
      'CORS origins always include production domains',
    ],
    changed: [],
  },
  {
    version: '1.0.0',
    date: '2025-12-01',
    added: [
      'Initial release — document OCR, face matching, verification pipeline',
      'RESTful API for verification workflows',
      'API key management system',
      'Webhook notification system with HMAC-SHA256 signing',
      'Sandbox environment for developer testing',
      'Rate limiting and abuse protection',
    ],
    fixed: [],
    changed: [],
  },
]

const resources = [
  { icon: '🔑', title: 'Developer Portal', desc: 'Get API keys, view usage stats and analytics', href: '/developer', cta: 'Open Portal →' },
  { icon: '🔍', title: 'Review Dashboard', desc: 'Review, approve, and manage verifications with your team', href: '/docs/review', cta: 'View Guide →' },
  { icon: '🎮', title: 'Live Demo', desc: 'Try the full verification flow with a sandbox key', href: '/demo', cta: 'Open Demo →' },
  { icon: '📦', title: 'JavaScript SDK', desc: 'TypeScript SDK with real-time events and embed component', href: 'https://www.npmjs.com/package/@identity-verification/sdk', cta: 'npm install @identity-verification/sdk' },
  { icon: '🔧', title: 'GitHub', desc: 'Source code, examples, and issue tracker', href: 'https://github.com/team-identity/identity-community', cta: 'View on GitHub →' },
  { icon: '✉️', title: 'Email Support', desc: 'Technical support and integration help', href: 'mailto:support@example.com', cta: 'support@example.com' },
]
</script>

<style>
.docs-rate-row {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  padding: 8px 0;
  border-top: 1px solid #eef2f7;
}
.docs-rate-label { font-size: 13px; color: #0f172a; }
.docs-rate-note { font-size: 11.5px; color: #94a3b8; }
.docs-rate-value { font-family: ui-monospace, Menlo, monospace; font-size: 12.5px; color: #0e7490; align-self: center; white-space: nowrap; }
.docs-status-row { display: flex; gap: 14px; align-items: center; padding: 7px 0; border-top: 1px solid #eef2f7; }
.docs-status-code { font-family: ui-monospace, Menlo, monospace; font-size: 12px; width: 72px; flex-shrink: 0; }
.docs-status-desc { font-size: 12.5px; color: #475569; }
.docs-release { margin-bottom: 18px; }
.docs-release .arco-card-header { padding: 12px 20px; }
.docs-release .arco-card-body { padding: 16px 20px; display: flex; flex-direction: column; gap: 16px; }
.docs-release-head { display: flex; align-items: center; gap: 10px; flex-wrap: wrap; }
.docs-release-date { font-family: ui-monospace, Menlo, monospace; font-size: 12px; color: #94a3b8; }
.docs-release-group { display: flex; flex-direction: column; gap: 4px; }
.docs-release-label {
  display: flex;
  align-items: center;
  gap: 6px;
  font-family: ui-monospace, Menlo, monospace;
  font-size: 10.5px;
  font-weight: 600;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  margin-bottom: 4px;
}
.docs-dot { width: 6px; height: 6px; border-radius: 50%; background: currentColor; }
.docs-release-added { color: #15803d; }
.docs-release-fixed { color: #b45309; }
.docs-release-changed { color: #1d4ed8; }
.docs-release-item { font-size: 12.5px; color: #475569; line-height: 1.65; padding: 3px 0 3px 12px; }
.docs-release-item-added { border-left: 2px solid rgba(21, 128, 61, 0.25); }
.docs-release-item-fixed { border-left: 2px solid rgba(180, 83, 9, 0.25); }
.docs-release-item-changed { border-left: 2px solid rgba(29, 78, 216, 0.25); }
.docs-resource { height: 100%; }
.docs-resource .arco-card-body { display: flex; flex-direction: column; gap: 6px; padding: 18px; }
.docs-resource-icon { font-size: 22px; }
.docs-resource-title { font-size: 14px; font-weight: 600; }
.docs-resource-desc { font-size: 12px; color: #64748b; line-height: 1.55; flex: 1; }
.docs-resource-cta { font-size: 12.5px; font-weight: 600; }
</style>
