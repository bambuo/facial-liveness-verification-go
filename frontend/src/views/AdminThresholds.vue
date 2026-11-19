<template>
  <div class="page">
    <a-page-header
      :show-back="false"
      title="Verification thresholds"
      subtitle="Tune production and sandbox decision boundaries."
    >
      <template #extra>
        <a-button :loading="loading" @click="load">
          <template #icon><icon-refresh /></template>
          Refresh
        </a-button>
      </template>
    </a-page-header>

    <a-alert v-if="loadError" type="error" :show-icon="true" style="margin-bottom:16px">
      <a-space>
        <span>{{ loadError }}</span>
        <a-link @click="load">Retry</a-link>
      </a-space>
    </a-alert>

    <a-row :gutter="[16, 16]">
      <a-col :xs="24" :lg="14">
        <a-card title="Admin settings">
          <template #extra><a-tag color="arcoblue">Live</a-tag></template>
          <a-form :model="{}" layout="vertical">
            <a-form-item
              label="Auto-approve threshold (face matching)"
              extra="Verifications at or above this production face-matching score are approved automatically. Range 0 – 1."
            >
              <a-input-number v-model="autoApprove" :min="0" :max="1" :step="0.01" :precision="2" style="width:100%" />
            </a-form-item>
            <a-form-item
              label="Manual review threshold (liveness)"
              extra="Liveness below this production score is routed to manual review. Range 0 – 1."
            >
              <a-input-number v-model="manualReview" :min="0" :max="1" :step="0.01" :precision="2" style="width:100%" />
            </a-form-item>
          </a-form>
          <a-space wrap>
            <a-button type="primary" :loading="saving" @click="save">Save thresholds</a-button>
            <a-button :loading="previewing" @click="preview">Preview impact</a-button>
            <a-popconfirm
              content="Reset every threshold to the system defaults?"
              ok-text="Reset"
              cancel-text="Cancel"
              @ok="reset"
            >
              <a-button status="warning" :loading="resetting">Reset defaults</a-button>
            </a-popconfirm>
            <a-link @click="load">Discard changes</a-link>
          </a-space>
        </a-card>
      </a-col>

      <a-col :xs="24" :lg="10">
        <a-card title="Preview impact">
          <template #extra><a-tag v-if="previewAt" size="small">Draft</a-tag></template>
          <a-space direction="vertical" size="medium" style="width:100%">
            <a-alert v-if="previewExplanation" type="info" :show-icon="true">{{ previewExplanation }}</a-alert>
            <a-descriptions
              v-if="previewRows.length"
              :column="1"
              size="small"
              bordered
              :data="previewRows"
            />
            <a-empty v-else description="Run Preview impact to see the resulting thresholds" />
          </a-space>
        </a-card>
      </a-col>
    </a-row>

    <a-card title="Current threshold matrix" style="margin-top:16px">
      <template #extra><a-typography-text type="secondary">{{ matrixMeta }}</a-typography-text></template>
      <a-alert type="info" :show-icon="true" style="margin-bottom:12px">
        Only face matching (auto-approve) and liveness (manual review) are writable through this deployment's
        update endpoint. The remaining scores are reported read-only.
      </a-alert>
      <a-table
        :data="matrixRows"
        :columns="matrixColumns"
        :loading="loading"
        :pagination="false"
        :bordered="{ cell: true }"
        row-key="key"
      >
        <template #setting="{ record }">
          <a-space size="mini">
            <span>{{ record.label }}</span>
            <a-tag v-if="!record.writable" size="small">read-only</a-tag>
          </a-space>
        </template>
      </a-table>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { Message } from '@arco-design/web-vue'
import { api, apiMessage } from '../lib/api'

const WRITABLE_KEYS = ['face_matching', 'liveness']
const KEY_LABELS: Record<string, string> = {
  photo_consistency: 'Photo consistency',
  face_matching: 'Face matching',
  liveness: 'Liveness',
  cross_validation: 'Cross validation',
  quality_minimum: 'Quality minimum',
  ocr_confidence: 'OCR confidence',
  pdf417_confidence: 'PDF417 confidence',
}

const matrix = reactive<{ production: Record<string, number>; sandbox: Record<string, number>; meta: Record<string, any> }>({
  production: {},
  sandbox: {},
  meta: {},
})

const loading = ref(false)
const saving = ref(false)
const previewing = ref(false)
const resetting = ref(false)
const loadError = ref('')
const autoApprove = ref(0.85)
const manualReview = ref(0.75)

const previewTarget = ref<any>(null)
const previewExplanationText = ref('')
const previewAt = ref('')

const matrixColumns = [
  { title: 'Threshold', slotName: 'setting' },
  { title: 'Production', dataIndex: 'production', width: 140 },
  { title: 'Sandbox', dataIndex: 'sandbox', width: 140 },
]

const matrixRows = computed(() => {
  const keys = Array.from(new Set([...Object.keys(matrix.production), ...Object.keys(matrix.sandbox)]))
  const ordered = [
    ...Object.keys(KEY_LABELS).filter((key) => keys.includes(key)),
    ...keys.filter((key) => !(key in KEY_LABELS)),
  ]
  return ordered.map((key) => ({
    key,
    label: KEY_LABELS[key] || key.replace(/_/g, ' '),
    production: formatScore(matrix.production[key]),
    sandbox: formatScore(matrix.sandbox[key]),
    writable: WRITABLE_KEYS.includes(key),
  }))
})

const matrixMeta = computed(() => {
  if (matrix.meta?.using_defaults) return 'Using system defaults'
  if (matrix.meta?.updated_at) return `Updated ${matrix.meta.updated_at}`
  return 'Organization overrides'
})

const previewRows = computed<{ label: string; value: string }[]>(() => {
  const source = previewTarget.value
  if (!source || typeof source !== 'object') return []
  const rows: { label: string; value: string }[] = []
  const walk = (value: any, prefix: string) => {
    if (value && typeof value === 'object' && !Array.isArray(value)) {
      Object.entries(value).forEach(([key, item]) => walk(item, prefix ? `${prefix}.${key}` : key))
      return
    }
    rows.push({ label: prefix || 'value', value: formatScore(value) })
  }
  walk(source, '')
  return rows
})

const previewExplanation = computed(() => {
  const source = previewTarget.value ? previewExplanationText.value : ''
  return typeof source === 'string' ? source : ''
})

function formatScore(value: unknown) {
  if (value === null || value === undefined || value === '') return '—'
  const numeric = typeof value === 'number'
    ? value
    : typeof value === 'string' && value.trim() !== '' && Number.isFinite(Number(value))
      ? Number(value)
      : Number.NaN
  if (!Number.isFinite(numeric)) return String(value)
  return numeric.toFixed(2)
}

function hydrateForm() {
  if (typeof matrix.production.face_matching === 'number') autoApprove.value = matrix.production.face_matching
  if (typeof matrix.production.liveness === 'number') manualReview.value = matrix.production.liveness
}

async function load() {
  loading.value = true
  loadError.value = ''
  try {
    const { data } = await api.get('/admin/thresholds')
    const payload = data.data || {}
    matrix.production = payload.production || {}
    matrix.sandbox = payload.sandbox || {}
    matrix.meta = payload.meta || {}
    hydrateForm()
  } catch (caught) {
    loadError.value = apiMessage(caught)
  } finally {
    loading.value = false
  }
}

async function save() {
  if (autoApprove.value <= manualReview.value) {
    Message.warning('The auto-approve threshold must be higher than the manual review threshold')
    return
  }
  saving.value = true
  try {
    const { data } = await api.put('/admin/thresholds', {
      auto_approve_threshold: autoApprove.value,
      manual_review_threshold: manualReview.value,
    })
    const payload = data.data || {}
    if (payload.production) matrix.production = payload.production
    if (payload.sandbox) matrix.sandbox = payload.sandbox
    if (payload.meta) matrix.meta = payload.meta
    Message.success('Thresholds saved')
    await load()
  } catch (caught) {
    Message.error(apiMessage(caught))
  } finally {
    saving.value = false
  }
}

async function preview() {
  previewing.value = true
  try {
    const { data } = await api.post('/admin/thresholds/preview', {
      auto_approve_threshold: autoApprove.value,
      manual_review_threshold: manualReview.value,
    })
    const payload = data.data || {}
    previewTarget.value = payload.preview || null
    previewExplanationText.value = typeof payload.explanation === 'string'
      ? payload.explanation
      : payload.explanation
        ? Object.entries(payload.explanation).map(([key, value]) => `${key.replace(/_/g, ' ')}: ${value}`).join(' · ')
        : ''
    previewAt.value = new Date().toISOString()
    Message.success('Preview generated')
  } catch (caught) {
    Message.error(apiMessage(caught))
  } finally {
    previewing.value = false
  }
}

async function reset() {
  resetting.value = true
  try {
    await api.post('/admin/thresholds/reset', {})
    previewTarget.value = null
    previewExplanationText.value = ''
    previewAt.value = ''
    Message.success('Thresholds reset to defaults')
    await load()
  } catch (caught) {
    Message.error(apiMessage(caught))
  } finally {
    resetting.value = false
  }
}

onMounted(load)
</script>
