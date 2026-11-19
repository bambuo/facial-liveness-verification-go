<template>
  <DocsShell slug="review-dashboard" :nav="NAV">
    <!-- OVERVIEW -->
    <DocsSectionAnchor id="overview" />
    <DocsH2 index="01">Review Dashboard</DocsH2>
    <DocsLead>
      The Review Dashboard lets you and your team review, approve, or reject identity verifications without
      writing any code. It's the fastest way to go live — integrate the API, then use the dashboard for manual
      review decisions while you build your automation.
    </DocsLead>

    <a-row :gutter="[12, 12]" class="docs-feature-row">
      <a-col v-for="card in highlights" :key="card.title" :xs="24" :sm="12" :md="6">
        <a-card class="docs-resource" :bordered="true">
          <div class="docs-resource-icon">{{ card.icon }}</div>
          <div class="docs-resource-title">{{ card.title }}</div>
          <div class="docs-resource-desc">{{ card.desc }}</div>
        </a-card>
      </a-col>
    </a-row>

    <DocsCallout type="tip">
      The Review Dashboard is designed for the transition period between "API works" and "fully automated."
      Many teams use it permanently for <code>manual_review</code> cases that need a human decision.
    </DocsCallout>

    <DocsDivider />

    <!-- ACCESS -->
    <DocsSectionAnchor id="access" />
    <DocsH2 index="02">Getting Access</DocsH2>
    <DocsLead>
      Developers invite team members from the Developer Portal. Team members access the dashboard via
      passwordless OTP — no passwords to manage.
    </DocsLead>

    <DocsSectionAnchor id="roles" />
    <h3 class="docs-h3"><span class="docs-number">1.</span> Roles</h3>
    <p class="docs-body-text docs-body-narrow">
      There are two roles for the Review Dashboard. Both are scoped to the developer's verifications and sign in
      via OTP.
    </p>

    <a-row :gutter="[12, 12]" class="docs-role-row">
      <a-col :xs="24" :md="12">
        <a-card :bordered="true" class="docs-role-card">
          <div class="docs-mode-head">
            <a-tag size="small" color="purple">Admin</a-tag>
            <span class="docs-role-title">Organization Admin</span>
          </div>
          <p class="docs-role-desc">
            Approve, reject, <strong>override</strong> verifications. Access analytics, manage GDPR erasure
            requests, and manage the reviewer team.
          </p>
        </a-card>
      </a-col>
      <a-col :xs="24" :md="12">
        <a-card :bordered="true" class="docs-role-card">
          <div class="docs-mode-head">
            <a-tag size="small">Reviewer</a-tag>
            <span class="docs-role-title">Reviewer</span>
          </div>
          <p class="docs-role-desc">
            Approve or reject verifications. Cannot override, access analytics, or manage GDPR requests.
          </p>
        </a-card>
      </a-col>
    </a-row>

    <DocsCallout type="note">
      Developers never see the Review Dashboard directly. They manage their API keys in the Developer Portal and
      invite Organization Admins to handle day-to-day verification review.
    </DocsCallout>

    <DocsSectionAnchor id="reviewers" />
    <h3 class="docs-h3"><span class="docs-number">2.</span> Invite Team Members</h3>
    <p class="docs-body-text docs-body-narrow">
      From the Developer Portal, invite team members as Organization Admins or Reviewers:
    </p>

    <div class="docs-step-chips">
      <div v-for="step in inviteSteps" :key="step.step" class="docs-step-chip">
        <span class="docs-step">{{ step.step }}</span>
        <span>{{ step.text }}</span>
      </div>
    </div>

    <DocsCallout type="warning">
      All team members can only see verifications belonging to <strong>your</strong> developer account. They
      cannot access API keys, billing, or other developers' data. Each person gets their own OTP-based login.
    </DocsCallout>

    <DocsSectionAnchor id="reviewer-login" />
    <h3 class="docs-h3"><span class="docs-number">3.</span> OTP Login</h3>
    <p class="docs-body-text docs-body-narrow">
      Both Organization Admins and Reviewers access the dashboard at <code class="docs-inline-code">/admin/login</code>.
      They enter their email, receive an OTP code, and are taken directly to the Review Dashboard. No passwords,
      no account creation required. A role badge in the header shows the user's access level.
    </p>

    <a-card :bordered="true" class="docs-otp-card">
      <div class="docs-otp-label">OTP LOGIN FLOW</div>
      <div v-for="(step, index) in otpSteps" :key="step" class="docs-otp-step">
        <span class="docs-otp-dot" :class="{ 'docs-otp-dot-done': index === 3 }" />
        <span>{{ step }}</span>
      </div>
    </a-card>

    <DocsDivider />

    <!-- DASHBOARD -->
    <DocsSectionAnchor id="dashboard" />
    <DocsH2 index="03">The Dashboard</DocsH2>
    <DocsLead>
      The dashboard is a single-page view with stats, a filterable table, and an expandable detail panel —
      everything you need to review verifications quickly.
    </DocsLead>

    <DocsSectionAnchor id="stats" />
    <h3 class="docs-h3">Stats Bar</h3>
    <p class="docs-body-text docs-body-narrow">
      At the top of the page, five stat cards show real-time counts:
    </p>

    <a-row :gutter="[10, 10]" class="docs-stats-row">
      <a-col v-for="stat in stats" :key="stat.label" :span="24 / 5">
        <a-card :bordered="true" class="docs-stat-card" :style="{ borderTop: `3px solid ${stat.color}` }">
          <div class="docs-stat-label">{{ stat.label.toUpperCase() }}</div>
          <div class="docs-stat-value" :style="{ color: stat.color }">{{ stat.value }}</div>
        </a-card>
      </a-col>
    </a-row>

    <DocsSectionAnchor id="table" />
    <h3 class="docs-h3">Verification Table</h3>
    <p class="docs-body-text docs-body-narrow">
      The main table shows all verifications with sortable columns. Click any row to expand the detail panel.
    </p>

    <a-card class="docs-table-card" :bordered="true">
      <a-table
        :data="verificationRows"
        :columns="verificationColumns"
        :pagination="false"
        size="small"
        :bordered="false"
        row-key="id"
      >
        <template #id="{ record }"><code class="docs-table-code">{{ record.id }}</code></template>
        <template #user="{ record }"><code class="docs-table-code">{{ record.user }}</code></template>
        <template #status="{ record }"><DocsStatusPill :status="record.status" /></template>
        <template #actions="{ record }">
          <a-link @click="onView(record.id)">View</a-link>
        </template>
      </a-table>
    </a-card>

    <DocsSectionAnchor id="filters" />
    <h3 class="docs-h3">Filters &amp; Search</h3>
    <p class="docs-body-text docs-body-narrow">
      Use the status tabs above the table to filter by verification state:
    </p>

    <a-tabs default-active-key="all" size="small" class="docs-filter-tabs">
      <a-tab-pane key="all" title="All" />
      <a-tab-pane key="manual" title="Manual Review" />
      <a-tab-pane key="verified" title="Verified" />
      <a-tab-pane key="failed" title="Failed" />
      <a-tab-pane key="pending" title="Pending" />
    </a-tabs>

    <DocsSectionAnchor id="detail" />
    <h3 class="docs-h3">Detail Panel</h3>
    <p class="docs-body-text docs-body-narrow">
      Clicking a row expands the detail panel showing everything about the verification:
    </p>

    <a-row :gutter="[10, 10]">
      <a-col v-for="panel in detailPanels" :key="panel.title" :xs="24" :sm="12" :md="8">
        <a-card :bordered="true" class="docs-detail-card">
          <div class="docs-detail-title">{{ panel.title }}</div>
          <div class="docs-detail-desc">{{ panel.desc }}</div>
        </a-card>
      </a-col>
    </a-row>

    <DocsDivider />

    <!-- ACTIONS -->
    <DocsSectionAnchor id="actions" />
    <DocsH2 index="04">Review Actions</DocsH2>
    <DocsLead>
      Each verification can be approved, rejected, or overridden. All actions require a confirmation dialog to
      prevent accidental decisions.
    </DocsLead>

    <DocsSectionAnchor id="approve" />
    <a-card :bordered="true" class="docs-action-card">
      <div class="docs-action-head">
        <span class="docs-action-dot" style="background: #15803d" />
        <span class="docs-action-title">Approve</span>
      </div>
      <div class="docs-action-body">
        Sets the verification status to <DocsStatusPill status="verified" />. The decision is final and triggers
        a <code class="docs-inline-code">verification.verified</code> webhook event to the developer's registered
        webhook URL.
      </div>
    </a-card>

    <DocsSectionAnchor id="reject" />
    <a-card :bordered="true" class="docs-action-card">
      <div class="docs-action-head">
        <span class="docs-action-dot" style="background: #b91c1c" />
        <span class="docs-action-title">Reject</span>
      </div>
      <div class="docs-action-body">
        Sets the verification status to <DocsStatusPill status="failed" />. You can provide an optional reason
        which is stored and included in the webhook payload. Triggers a
        <code class="docs-inline-code">verification.failed</code> webhook event.
      </div>
    </a-card>

    <DocsSectionAnchor id="override" />
    <a-card :bordered="true" class="docs-action-card">
      <div class="docs-action-head">
        <span class="docs-action-dot" style="background: #b45309" />
        <span class="docs-action-title">Override</span>
        <a-tag size="small" color="purple">Admin Only</a-tag>
      </div>
      <div class="docs-action-body">
        Lets you set any status — useful for edge cases. For example, overriding a
        <DocsStatusPill status="failed" /> verification to <DocsStatusPill status="verified" /> when you've
        manually confirmed the person's identity through other means. A reason is required for overrides. This
        action is only available to <strong>Organization Admins</strong> — regular Reviewers can only approve or
        reject.
      </div>
    </a-card>

    <DocsCallout type="warning">
      All review actions are logged in the audit trail. Webhook notifications fire immediately after a decision,
      so your application can react in real time.
    </DocsCallout>

    <DocsDivider />

    <!-- STATUS REFERENCE -->
    <DocsSectionAnchor id="statuses" />
    <DocsH2 index="05">Status Reference</DocsH2>
    <DocsLead>
      Every verification moves through these states. The dashboard color-codes each status for quick visual
      scanning.
    </DocsLead>

    <a-card class="docs-table-card" :bordered="true">
      <div v-for="row in statusReference" :key="row.status" class="docs-status-ref-row">
        <div class="docs-status-ref-pill"><DocsStatusPill :status="row.status" /></div>
        <div class="docs-status-ref-desc">{{ row.desc }}</div>
      </div>
    </a-card>

    <DocsDivider />

    <!-- WEBHOOKS -->
    <DocsSectionAnchor id="webhooks" />
    <DocsH2 index="06">Webhooks</DocsH2>
    <DocsLead>
      When you approve, reject, or override a verification from the dashboard, a webhook is fired to your
      registered URL — the same events as automated decisions.
    </DocsLead>

    <a-card class="docs-table-card" :bordered="true">
      <a-table
        :data="webhookRows"
        :columns="webhookColumns"
        :pagination="false"
        size="small"
        :bordered="false"
        row-key="event"
      >
        <template #event="{ record }"><code class="docs-table-code">{{ record.event }}</code></template>
        <template #payload="{ record }"><code class="docs-table-payload">{{ record.payload }}</code></template>
      </a-table>
    </a-card>

    <DocsCallout type="tip">
      Set up your webhook URL in the <router-link to="/developer">Developer Portal</router-link> under Settings.
      Webhooks retry up to 3 times on failure. See the <router-link to="/docs">API Docs</router-link> for the
      full webhook payload schema.
    </DocsCallout>
  </DocsShell>
</template>

<script setup lang="ts">
import { Message } from '@arco-design/web-vue'
import DocsShell, {
  DocsSectionAnchor,
  DocsH2,
  DocsLead,
  DocsDivider,
  DocsCallout,
  DocsStatusPill,
  type DocsNavItem,
} from '../components/DocsShell.vue'

const NAV: DocsNavItem[] = [
  { id: 'overview', label: 'Overview', depth: 0 },
  { id: 'access', label: 'Getting Access', depth: 0 },
  { id: 'roles', label: 'Roles', depth: 1 },
  { id: 'reviewers', label: 'Invite Team Members', depth: 1 },
  { id: 'reviewer-login', label: 'OTP Login', depth: 1 },
  { id: 'dashboard', label: 'The Dashboard', depth: 0 },
  { id: 'stats', label: 'Stats Bar', depth: 1 },
  { id: 'table', label: 'Verification Table', depth: 1 },
  { id: 'filters', label: 'Filters & Search', depth: 1 },
  { id: 'detail', label: 'Detail Panel', depth: 1 },
  { id: 'actions', label: 'Review Actions', depth: 0 },
  { id: 'approve', label: 'Approve', depth: 1 },
  { id: 'reject', label: 'Reject', depth: 1 },
  { id: 'override', label: 'Override', depth: 1 },
  { id: 'statuses', label: 'Status Reference', depth: 0 },
  { id: 'webhooks', label: 'Webhooks', depth: 0 },
]

const highlights = [
  { icon: '🔍', title: 'Review Verifications', desc: 'View document images, OCR data, face match scores, and cross-validation results' },
  { icon: '✅', title: 'Approve or Reject', desc: 'Make manual decisions with confirmation — results pushed via webhooks instantly' },
  { icon: '👥', title: 'Invite Reviewers', desc: 'Give team members access with passwordless OTP login — no shared credentials' },
  { icon: '📊', title: 'Live Stats', desc: 'See total, pending, verified, and failed counts at a glance' },
]

const inviteSteps = [
  { step: '1', text: 'Open Developer Portal → Settings (gear icon)' },
  { step: '2', text: 'Scroll to "Team Management"' },
  { step: '3', text: 'Enter email, name, select role → Invite' },
]

const otpSteps = [
  'Reviewer navigates to /admin/login',
  'Enters their invited email address',
  'Receives 6-digit OTP via email',
  'Enters code → redirected to /admin/verifications',
]

const stats = [
  { label: 'Total', value: '—', color: '#0f172a' },
  { label: 'Review', value: '—', color: '#b45309' },
  { label: 'Pending', value: '—', color: '#0e7490' },
  { label: 'Verified', value: '—', color: '#15803d' },
  { label: 'Failed', value: '—', color: '#b91c1c' },
]

const verificationRows = [
  { id: 'a3f8...c2d1', user: '9e7b...4a01', status: 'manual_review', doc: 'drivers_license', date: '2 min ago' },
  { id: 'b1e2...d4f5', user: '3c8a...7b02', status: 'verified', doc: 'passport', date: '15 min ago' },
  { id: 'c9d0...e6a3', user: '5f2d...1c03', status: 'failed', doc: 'national_id', date: '1 hour ago' },
]

const verificationColumns = [
  { title: 'ID', slotName: 'id', width: 140 },
  { title: 'User ID', slotName: 'user', width: 140 },
  { title: 'Status', slotName: 'status', width: 150 },
  { title: 'Doc type', dataIndex: 'doc', width: 140 },
  { title: 'Created', dataIndex: 'date', width: 120 },
  { title: 'Actions', slotName: 'actions', width: 90 },
]

const detailPanels = [
  { title: 'Document Images', desc: 'Front and back photos of the uploaded ID document' },
  { title: 'OCR Data', desc: 'Extracted name, date of birth, document number, expiry' },
  { title: 'Cross-Validation', desc: 'Front OCR vs back barcode consistency scores' },
  { title: 'Face Match', desc: 'Similarity score between document photo and live capture' },
  { title: 'Gate Analysis', desc: 'Score bars for OCR, cross-validation, liveness, deepfake, and face match gates' },
  { title: 'Risk Assessment', desc: 'Overall risk score, risk level, contributing factors, and AML screening results' },
]

const statusReference = [
  { status: 'pending', desc: 'Verification created, waiting for document upload or processing to complete.' },
  { status: 'manual_review', desc: 'Automated checks flagged something — a human needs to review and decide.' },
  { status: 'verified', desc: 'Identity confirmed. Either by automated pipeline or manual approval.' },
  { status: 'failed', desc: 'Verification failed. Document unreadable, face mismatch, or manually rejected.' },
]

const webhookRows = [
  { event: 'verification.verified', trigger: 'Approve', payload: 'verification_id, status, reviewed_by, reviewed_at' },
  { event: 'verification.failed', trigger: 'Reject', payload: 'verification_id, status, reason, reviewed_by, reviewed_at' },
  { event: 'verification.status_changed', trigger: 'Override', payload: 'verification_id, old_status, new_status, reason, reviewed_by' },
]

const webhookColumns = [
  { title: 'Event', slotName: 'event', width: 220 },
  { title: 'Trigger', dataIndex: 'trigger', width: 110 },
  { title: 'Payload includes', slotName: 'payload' },
]

function onView(id: string) {
  Message.info(`View ${id}`)
}
</script>

<style>
.docs-feature-row { margin-bottom: 8px; }
.docs-body-narrow { max-width: 680px; }
.docs-number { color: #0e7490; font-family: ui-monospace, Menlo, monospace; font-size: 13px; font-weight: 400; }
.docs-role-row { margin-bottom: 20px; }
.docs-role-card .arco-card-body { padding: 18px; }
.docs-role-title { font-size: 14px; font-weight: 600; }
.docs-role-desc { margin: 0; font-size: 12px; color: #64748b; line-height: 1.55; }
.docs-step-chips { display: flex; gap: 8px; flex-wrap: wrap; margin-bottom: 20px; }
.docs-step-chip {
  display: flex;
  align-items: center;
  gap: 10px;
  background: #fff;
  border: 1px solid var(--rule, #e4e2dc);
  padding: 10px 14px;
  flex: 1 1 180px;
  font-size: 12.5px;
}
.docs-otp-card { max-width: 500px; margin-bottom: 8px; }
.docs-otp-card .arco-card-body { padding: 18px 20px; }
.docs-otp-label { font-family: ui-monospace, Menlo, monospace; font-size: 11px; letter-spacing: 0.06em; color: #94a3b8; margin-bottom: 10px; }
.docs-otp-step { display: flex; align-items: center; gap: 10px; padding: 5px 0; font-size: 12.5px; }
.docs-otp-dot { width: 6px; height: 6px; border-radius: 50%; background: #0e7490; flex-shrink: 0; }
.docs-otp-dot-done { background: #15803d; }
.docs-stats-row { margin-bottom: 20px; }
.docs-stat-card .arco-card-body { padding: 12px 14px; }
.docs-stat-label { font-family: ui-monospace, Menlo, monospace; font-size: 10px; color: #64748b; letter-spacing: 0.06em; margin-bottom: 4px; }
.docs-stat-value { font-family: ui-monospace, Menlo, monospace; font-size: 17px; font-weight: 600; }
.docs-filter-tabs { margin-bottom: 18px; }
.docs-detail-card .arco-card-body { padding: 12px 14px; }
.docs-detail-title { font-size: 13px; font-weight: 600; margin-bottom: 4px; }
.docs-detail-desc { font-size: 12px; color: #64748b; line-height: 1.55; }
.docs-action-card { margin-bottom: 14px; }
.docs-action-card .arco-card-body { padding: 18px 20px; }
.docs-action-head { display: flex; align-items: center; gap: 10px; margin-bottom: 8px; }
.docs-action-dot { width: 10px; height: 10px; border-radius: 50%; }
.docs-action-title { font-size: 14.5px; font-weight: 600; }
.docs-action-body { font-size: 13px; color: #475569; line-height: 1.7; }
.docs-status-ref-row { display: flex; gap: 16px; padding: 12px 0; border-top: 1px solid #eef2f7; align-items: flex-start; }
.docs-status-ref-row:first-child { border-top: none; }
.docs-status-ref-pill { width: 130px; flex-shrink: 0; }
.docs-status-ref-desc { font-size: 13px; color: #475569; line-height: 1.65; }
</style>
