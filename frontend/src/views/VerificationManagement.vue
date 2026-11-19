<template>
  <div class="page">
    <a-page-header
      :show-back="false"
      title="Verification management"
      subtitle="Review outcomes, inspect evidence and resolve manual-review cases."
    >
      <template #extra>
        <a-button :loading="loading" @click="refresh">
          <template #icon><icon-refresh /></template>
          Refresh
        </a-button>
      </template>
    </a-page-header>

    <a-grid :cols="{ xs: 2, sm: 3, md: 5 }" :col-gap="12" :row-gap="12">
      <a-grid-item v-for="card in statCards" :key="card.key">
        <a-card size="small">
          <a-statistic :title="card.label" :value="card.value" :value-style="card.style" />
        </a-card>
      </a-grid-item>
    </a-grid>

    <a-card style="margin-top:16px">
      <a-space wrap style="margin-bottom:14px">
        <a-radio-group v-model="statusFilter" type="button" @change="onFilterChange">
          <a-radio value="all">All</a-radio>
          <a-radio value="manual_review">Manual review</a-radio>
          <a-radio value="pending">Pending</a-radio>
          <a-radio value="verified">Verified</a-radio>
          <a-radio value="failed">Failed</a-radio>
        </a-radio-group>
        <a-input v-model="search" allow-clear placeholder="Search verification or user ID" style="width:260px">
          <template #prefix><icon-search /></template>
        </a-input>
        <div class="spacer" />
        <a-typography-text type="secondary">
          {{ filteredRows.length }} of {{ total }} records
        </a-typography-text>
      </a-space>

      <a-alert v-if="error" type="error" :show-icon="true" style="margin-bottom:12px">
        <a-space>
          <span>{{ error }}</span>
          <a-link @click="load">Retry</a-link>
        </a-space>
      </a-alert>

      <a-table
        :data="filteredRows"
        :columns="columns"
        :loading="loading"
        :pagination="false"
        :bordered="{ cell: true }"
        :scroll="{ x: 1080 }"
        row-key="verification_id"
        @row-click="onRowClick"
      >
        <template #verification="{ record }">
          <a-space size="mini">
            <a-tooltip :content="record.verification_id || '—'">
              <span class="mono" style="white-space:nowrap">{{ truncateId(record.verification_id) }}</span>
            </a-tooltip>
            <a-tag v-if="record.is_sandbox" color="purple" size="small">Sandbox</a-tag>
          </a-space>
        </template>
        <template #user="{ record }">
          <a-tooltip v-if="record.user_id" :content="record.user_id">
            <span class="mono" style="white-space:nowrap">{{ truncateId(record.user_id) }}</span>
          </a-tooltip>
          <span v-else class="muted">—</span>
        </template>
        <template #status="{ record }">
          <a-tag :color="statusMeta(record).color">{{ statusMeta(record).label }}</a-tag>
        </template>
        <template #step="{ record }">
          <span class="mono" style="white-space:nowrap">{{ record.current_step ?? '—' }} / {{ record.total_steps ?? '—' }}</span>
        </template>
        <template #actions="{ record }">
          <a-space size="mini" @click.stop>
            <a-button size="mini" type="text" @click="openDetail(record)">Inspect</a-button>
            <template v-if="statusMeta(record).reviewable">
              <a-button size="mini" status="success" @click="askReview(record, 'approve')">Approve</a-button>
              <a-button size="mini" status="danger" @click="askReview(record, 'reject')">Reject</a-button>
              <a-button size="mini" status="warning" @click="askReview(record, 'override')">Override</a-button>
            </template>
          </a-space>
        </template>
      </a-table>

      <div style="display:flex;justify-content:flex-end;margin-top:14px">
        <a-pagination
          :current="page"
          :page-size="limit"
          :total="total"
          :show-total="true"
          :show-jumper="true"
          @change="onPageChange"
        />
      </div>
    </a-card>

    <a-drawer
      v-model:visible="drawerVisible"
      :width="680"
      title="Verification detail"
      :footer="false"
      unmount-on-close
    >
      <a-spin :loading="detailLoading" style="width:100%">
        <template v-if="detail">
          <a-descriptions :column="1" bordered size="small" :label-style="{ width: '180px' }">
            <a-descriptions-item label="Verification ID">
              <span class="mono">{{ detail.verification_id }}</span>
            </a-descriptions-item>
            <a-descriptions-item label="User ID">
              <span class="mono">{{ detail.user_id || '—' }}</span>
            </a-descriptions-item>
            <a-descriptions-item label="Status">
              <a-tag :color="statusMeta(detail).color">{{ statusMeta(detail).label }}</a-tag>
            </a-descriptions-item>
            <a-descriptions-item label="Final result">{{ detail.final_result || '—' }}</a-descriptions-item>
            <a-descriptions-item label="Step">{{ detail.current_step ?? '—' }} / {{ detail.total_steps ?? '—' }}</a-descriptions-item>
            <a-descriptions-item label="Mode">{{ detail.verification_mode || '—' }}</a-descriptions-item>
            <a-descriptions-item label="Sandbox">{{ detail.is_sandbox ? 'Yes' : 'No' }}</a-descriptions-item>
            <a-descriptions-item label="Created">{{ formatDateTime(detail.created_at) }}</a-descriptions-item>
            <a-descriptions-item v-if="detail.failure_reason" label="Failure reason">
              {{ detail.failure_reason }}
            </a-descriptions-item>
          </a-descriptions>

          <a-divider>Evidence</a-divider>
          <a-empty v-if="!hasEvidence" description="No documents returned for this verification" />
          <a-space v-else wrap>
            <a-image
              v-for="doc in detail.documents"
              :key="doc.id"
              :src="doc.url"
              :width="140"
              :height="96"
              fit="cover"
              :title="doc.file_name"
            />
            <a-image v-if="detail.selfie_url" :src="detail.selfie_url" :width="140" :height="96" fit="cover" title="Live capture" />
          </a-space>

          <a-divider>Linked verifications</a-divider>
          <a-empty v-if="linked.length === 0" description="No linked verifications" />
          <a-table
            v-else
            :data="linked"
            :columns="linkedColumns"
            :pagination="false"
            :bordered="{ cell: true }"
            size="small"
            row-key="verification_id"
            @row-click="onRowClick"
          >
            <template #status="{ record }">
              <a-tag :color="statusMeta(record).color">{{ statusMeta(record).label }}</a-tag>
            </template>
          </a-table>

          <a-divider>Review decision</a-divider>
          <a-form :model="{}" layout="vertical">
            <a-form-item label="Reason" extra="Optional for approval, recommended for rejection or override.">
              <a-textarea v-model="reason" :rows="3" placeholder="Why is this decision being recorded?" />
            </a-form-item>
            <a-space wrap>
              <a-button type="primary" status="success" @click="askReview(detail, 'approve')">Approve</a-button>
              <a-button status="danger" @click="askReview(detail, 'reject')">Reject</a-button>
              <a-button status="warning" @click="askReview(detail, 'override')">Override status</a-button>
              <a-button @click="askReview(detail, 'manual_review')">Keep in review</a-button>
            </a-space>
          </a-form>
        </template>
        <a-empty v-else-if="!detailLoading" description="This verification could not be loaded" />
      </a-spin>
    </a-drawer>

    <a-modal
      v-model:visible="confirmVisible"
      :title="confirmTitle"
      :width="460"
      :ok-text="confirmOkText"
      :ok-loading="submitting"
      :ok-button-props="{ status: confirmStatus }"
      cancel-text="Cancel"
      unmount-on-close
      @ok="submitReview"
    >
      <a-space direction="vertical" size="medium" style="width:100%">
        <a-typography-text type="secondary">{{ confirmDescription }}</a-typography-text>
        <a-form :model="{}" layout="vertical">
          <a-form-item label="Verification ID">
            <a-input :model-value="confirmTarget?.verification_id || ''" readonly />
          </a-form-item>
          <a-form-item v-if="confirmDecision === 'override'" label="New status">
            <a-select v-model="overrideStatus">
              <a-option value="verified">Verified</a-option>
              <a-option value="failed">Failed</a-option>
              <a-option value="manual_review">Manual review</a-option>
              <a-option value="pending">Pending</a-option>
            </a-select>
          </a-form-item>
          <a-form-item :label="confirmDecision === 'approve' ? 'Reason (optional)' : 'Reason (recommended)'">
            <a-textarea v-model="confirmReason" :rows="3" :placeholder="confirmReasonPlaceholder" />
          </a-form-item>
        </a-form>
      </a-space>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { Message } from '@arco-design/web-vue'
import { api, apiMessage } from '../lib/api'

type Decision = 'approve' | 'reject' | 'override' | 'manual_review'

const LIMIT = 25

const rows = ref<any[]>([])
const loading = ref(false)
const error = ref('')
const statusFilter = ref<'all' | 'manual_review' | 'pending' | 'verified' | 'failed'>('all')
const search = ref('')
const page = ref(1)
const limit = ref(LIMIT)
const total = ref(0)
const stats = reactive<Record<string, number>>({})

const drawerVisible = ref(false)
const detail = ref<any>(null)
const detailLoading = ref(false)
const linked = ref<any[]>([])
const reason = ref('')

const confirmVisible = ref(false)
const confirmTarget = ref<any>(null)
const confirmDecision = ref<Decision>('approve')
const confirmReason = ref('')
const overrideStatus = ref('verified')
const submitting = ref(false)

const columns = [
  { title: 'Verification ID', slotName: 'verification', width: 220 },
  { title: 'User ID', slotName: 'user', width: 150 },
  { title: 'Status', slotName: 'status', width: 120 },
  { title: 'Final result', dataIndex: 'final_result', width: 110, ellipsis: true, tooltip: true },
  { title: 'Mode', dataIndex: 'verification_mode', width: 100, ellipsis: true, tooltip: true },
  { title: 'Step', slotName: 'step', width: 90 },
  { title: 'Actions', slotName: 'actions', width: 290 },
]

const linkedColumns = [
  { title: 'Verification ID', dataIndex: 'verification_id', ellipsis: true, tooltip: true },
  { title: 'Status', slotName: 'status', width: 140 },
]

const statusConfig: Record<string, { key: string; label: string; color: string; reviewable: boolean }> = {
  verified: { key: 'verified', label: 'Verified', color: 'green', reviewable: false },
  complete: { key: 'verified', label: 'Complete', color: 'green', reviewable: false },
  failed: { key: 'failed', label: 'Failed', color: 'red', reviewable: false },
  hard_rejected: { key: 'failed', label: 'Hard rejected', color: 'red', reviewable: false },
  manual_review: { key: 'manual_review', label: 'Manual review', color: 'orange', reviewable: true },
  pending: { key: 'pending', label: 'Pending', color: 'arcoblue', reviewable: true },
  processing: { key: 'processing', label: 'Processing', color: 'cyan', reviewable: true },
  awaiting_front: { key: 'processing', label: 'Awaiting front', color: 'cyan', reviewable: true },
  awaiting_back: { key: 'processing', label: 'Awaiting back', color: 'cyan', reviewable: true },
  awaiting_live: { key: 'processing', label: 'Awaiting live capture', color: 'cyan', reviewable: true },
  awaiting_voice: { key: 'processing', label: 'Awaiting voice', color: 'cyan', reviewable: true },
  face_matching: { key: 'processing', label: 'Face matching', color: 'cyan', reviewable: true },
  front_processing: { key: 'processing', label: 'Front processing', color: 'cyan', reviewable: true },
}

function statusMeta(record: any) {
  const raw = String(record?.final_result || record?.status || '').toLowerCase()
  return statusConfig[raw] || { key: raw || 'unknown', label: record?.status || 'Unknown', color: 'gray', reviewable: false }
}

function truncateId(value?: string | null) {
  if (!value) return '—'
  return value.length > 12 ? `${value.slice(0, 6)}...${value.slice(-4)}` : value
}

function formatDateTime(value?: string | null) {
  if (!value) return '—'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '—'
  return `${date.toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' })} ${date.toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit' })}`
}

function statusCount(key: string) {
  return Number(stats[key] ?? 0)
}

const totalCount = computed(() => Object.values(stats).reduce((sum, value) => sum + Number(value || 0), 0))

const statCards = computed(() => [
  { key: 'total', label: 'Total (30d)', value: totalCount.value, style: { color: '#0f172a' } },
  { key: 'manual_review', label: 'Manual review', value: statusCount('manual_review'), style: { color: '#d97706' } },
  { key: 'pending', label: 'Pending', value: statusCount('pending'), style: { color: '#2563eb' } },
  { key: 'verified', label: 'Verified', value: statusCount('verified'), style: { color: '#15803d' } },
  { key: 'failed', label: 'Failed', value: statusCount('failed'), style: { color: '#dc2626' } },
])

const filteredRows = computed(() => {
  const query = search.value.trim().toLowerCase()
  if (!query) return rows.value
  return rows.value.filter((row) =>
    String(row.verification_id || '').toLowerCase().includes(query) ||
    String(row.user_id || '').toLowerCase().includes(query),
  )
})

const hasEvidence = computed(() => {
  if (!detail.value) return false
  const documents = Array.isArray(detail.value.documents) ? detail.value.documents : []
  return documents.length > 0 || Boolean(detail.value.selfie_url)
})

const confirmTitle = computed(() => {
  if (confirmDecision.value === 'approve') return 'Approve verification'
  if (confirmDecision.value === 'reject') return 'Reject verification'
  if (confirmDecision.value === 'override') return 'Override verification status'
  return 'Return to manual review'
})

const confirmDescription = computed(() => {
  if (confirmDecision.value === 'approve') return 'The applicant will be marked as verified and notified through the configured webhook.'
  if (confirmDecision.value === 'reject') return 'The applicant will be rejected. Record the reason so the decision stays auditable.'
  if (confirmDecision.value === 'override') return 'Force this verification to a different state. Use only when the automated decision is wrong.'
  return 'Keep this verification in the manual review queue for another reviewer.'
})

const confirmOkText = computed(() => {
  if (submitting.value) return 'Saving...'
  if (confirmDecision.value === 'approve') return 'Approve'
  if (confirmDecision.value === 'reject') return 'Reject'
  if (confirmDecision.value === 'override') return 'Override'
  return 'Keep in review'
})

const confirmStatus = computed(() => {
  if (confirmDecision.value === 'approve') return 'success'
  if (confirmDecision.value === 'reject') return 'danger'
  if (confirmDecision.value === 'override') return 'warning'
  return 'normal'
})

const confirmReasonPlaceholder = computed(() => {
  if (confirmDecision.value === 'approve') return 'Optional note for the audit trail'
  if (confirmDecision.value === 'reject') return 'Describe the evidence that led to this rejection'
  if (confirmDecision.value === 'override') return 'Explain why the automatic decision is being overridden'
  return 'Note for the next reviewer'
})

function onFilterChange() {
  page.value = 1
  load()
}

function onPageChange(value: number) {
  page.value = value
  load()
}

function onRowClick(record: any) {
  if (record?.verification_id) openDetail(record)
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    const params: Record<string, unknown> = { page: page.value, limit: limit.value }
    if (statusFilter.value !== 'all') params.status = statusFilter.value
    const [list, summary] = await Promise.all([
      api.get('/admin/verifications', { params }),
      api.get('/admin/dashboard'),
    ])
    rows.value = list.data.verifications || []
    total.value = list.data.pagination?.total || 0
    Object.keys(stats).forEach((key) => delete stats[key])
    Object.assign(stats, summary.data.stats || {})
  } catch (caught) {
    error.value = apiMessage(caught)
  } finally {
    loading.value = false
  }
}

async function refresh() {
  await load()
}

async function openDetail(record: any) {
  drawerVisible.value = true
  detailLoading.value = true
  detail.value = null
  linked.value = []
  reason.value = ''
  const id = record.verification_id
  const [verificationResult, duplicateResult] = await Promise.allSettled([
    api.get(`/admin/verification/${id}`),
    api.get(`/admin/verification/${id}/duplicates`),
  ])
  if (verificationResult.status === 'fulfilled') {
    detail.value = verificationResult.value.data.verification || record
  } else {
    Message.error(apiMessage(verificationResult.reason))
  }
  if (duplicateResult.status === 'fulfilled') {
    const list = duplicateResult.value.data?.linked_verifications
    linked.value = Array.isArray(list) ? list : []
  }
  detailLoading.value = false
}

function askReview(record: any, decision: Decision) {
  if (!record?.verification_id) return
  confirmTarget.value = record
  confirmDecision.value = decision
  confirmReason.value = decision === 'approve' ? '' : reason.value
  overrideStatus.value = 'verified'
  confirmVisible.value = true
}

function applyDecision(target: any, resolved: string) {
  if (!target) return
  target.final_result = resolved
  target.status = resolved === 'verified'
    ? 'COMPLETE'
    : resolved === 'failed'
      ? 'HARD_REJECTED'
      : resolved === 'manual_review'
        ? 'COMPLETE'
        : resolved
}

async function submitReview() {
  if (!confirmTarget.value) return
  submitting.value = true
  try {
    const body: Record<string, string> = { decision: confirmDecision.value }
    if (confirmDecision.value === 'override') body.new_status = overrideStatus.value
    if (confirmReason.value.trim()) body.reason = confirmReason.value.trim()
    await api.put(`/admin/verification/${confirmTarget.value.verification_id}/review`, body)

    const resolved = confirmDecision.value === 'approve'
      ? 'verified'
      : confirmDecision.value === 'reject'
        ? 'failed'
        : confirmDecision.value === 'override'
          ? overrideStatus.value
          : 'manual_review'

    const targetId = confirmTarget.value.verification_id
    rows.value = rows.value.map((row) => {
      if (row.verification_id !== targetId) return row
      const next = { ...row }
      applyDecision(next, resolved)
      return next
    })
    if (detail.value?.verification_id === targetId) {
      const next = { ...detail.value }
      applyDecision(next, resolved)
      detail.value = next
    }
    confirmVisible.value = false
    confirmTarget.value = null
    confirmReason.value = ''
    Message.success(`Review recorded: ${resolved.replace(/_/g, ' ')}`)
    const summary = await api.get('/admin/dashboard').catch(() => null)
    if (summary?.data?.stats) {
      Object.keys(stats).forEach((key) => delete stats[key])
      Object.assign(stats, summary.data.stats)
    }
  } catch (caught) {
    Message.error(apiMessage(caught))
  } finally {
    submitting.value = false
  }
}

onMounted(load)
</script>
