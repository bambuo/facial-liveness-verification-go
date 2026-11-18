<template>
  <div class="page">
    <div class="card-title">
      <div><div class="eyebrow">Operations</div><h2 style="margin:6px 0 4px">Verification management</h2><p class="muted">Review outcomes, inspect evidence and resolve manual-review cases.</p></div>
      <a-button @click="load" :loading="loading">Refresh</a-button>
    </div>
    <div class="grid grid-4" style="margin-top:18px">
      <div class="panel"><div class="stat-label">Total (30d)</div><div class="stat-value">{{ totalCount() }}</div></div>
      <div class="panel"><div class="stat-label">Verified</div><div class="stat-value">{{ statusCount('verified') }}</div></div>
      <div class="panel"><div class="stat-label">Manual review</div><div class="stat-value">{{ statusCount('manual_review') }}</div></div>
      <div class="panel"><div class="stat-label">Failed</div><div class="stat-value">{{ statusCount('failed') }}</div></div>
    </div>
    <div class="panel" style="margin-top:18px">
      <div class="toolbar" style="margin-bottom:14px">
        <a-select v-model="status" placeholder="All statuses" allow-clear style="width:180px">
          <a-option value="pending">pending</a-option><a-option value="processing">processing</a-option>
          <a-option value="verified">verified</a-option><a-option value="failed">failed</a-option><a-option value="manual_review">manual_review</a-option>
        </a-select>
        <a-input v-model="developerID" placeholder="Developer UUID" style="width:300px" allow-clear />
        <a-button type="primary" @click="page=1;load()">Filter</a-button>
        <div class="spacer"></div><span class="muted">{{ total }} records</span>
      </div>
      <a-table :data="rows" :columns="columns" :loading="loading" :pagination="false" row-key="verification_id">
        <template #status="{ record }"><a-tag :color="statusColor(record.status)">{{ record.status }}</a-tag></template>
        <template #actions="{ record }"><a-button size="mini" @click="openDetail(record)">Inspect</a-button></template>
      </a-table>
      <div class="toolbar" style="justify-content:flex-end;margin-top:14px">
        <a-pagination :current="page" :page-size="limit" :total="total" @change="onPageChange" />
      </div>
    </div>

    <a-drawer v-model:visible="drawer" :width="520" title="Verification detail">
      <template v-if="detail">
        <a-descriptions :column="1" bordered size="small">
          <a-descriptions-item label="Verification ID"><span class="mono">{{ detail.verification_id }}</span></a-descriptions-item>
          <a-descriptions-item label="Status"><a-tag :color="statusColor(detail.status)">{{ detail.status }}</a-tag></a-descriptions-item>
          <a-descriptions-item label="Final result">{{ detail.final_result || '—' }}</a-descriptions-item>
          <a-descriptions-item label="Step">{{ detail.current_step }} / {{ detail.total_steps }}</a-descriptions-item>
          <a-descriptions-item label="Mode">{{ detail.verification_mode || '—' }}</a-descriptions-item>
          <a-descriptions-item label="Sandbox">{{ detail.is_sandbox ? 'Yes' : 'No' }}</a-descriptions-item>
        </a-descriptions>
        <a-divider>Review decision</a-divider>
        <a-form :model="{}" layout="vertical">
          <a-form-item label="Reason"><a-textarea v-model="reason" :rows="3" placeholder="Optional review note" /></a-form-item>
          <a-space><a-button type="primary" status="success" @click="review('approve')">Approve</a-button><a-button status="danger" @click="review('reject')">Reject</a-button><a-button @click="review('manual_review')">Keep in review</a-button></a-space>
        </a-form>
      </template>
    </a-drawer>
  </div>
</template>
<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { Message } from '@arco-design/web-vue'
import { api, apiMessage } from '../lib/api'
const rows = ref<any[]>([]); const loading = ref(false); const status = ref(''); const developerID = ref('')
const page = ref(1); const limit = ref(20); const total = ref(0); const stats = reactive<Record<string, number>>({})
const drawer = ref(false); const detail = ref<any>(null); const reason = ref('')
const columns = [
  { title: 'Verification', dataIndex: 'verification_id', ellipsis: true },
  { title: 'Status', slotName: 'status', width: 150 },
  { title: 'Final result', dataIndex: 'final_result', width: 120 },
  { title: 'Mode', dataIndex: 'verification_mode', width: 110 },
  { title: 'Sandbox', dataIndex: 'is_sandbox', width: 90 },
  { title: '', slotName: 'actions', width: 90 },
]
function statusColor(value: string) {
  if (value === 'COMPLETE' || value === 'verified') return 'green'
  if (value === 'HARD_REJECTED' || value === 'failed') return 'red'
  if (value === 'manual_review') return 'orange'
  return 'arcoblue'
}
function statusCount(key: string) { return stats[key] ?? 0 }
const totalCount = () => Object.values(stats).reduce((a, b) => a + b, 0)
async function load() {
  loading.value = true
  try {
    const params: Record<string, any> = { page: page.value, limit: limit.value }
    if (status.value) params.status = status.value
    if (developerID.value) params.developer_id = developerID.value
    const [list, summary] = await Promise.all([api.get('/admin/verifications', { params }), api.get('/admin/dashboard')])
    rows.value = list.data.verifications || []
    total.value = list.data.pagination?.total || 0
    Object.assign(stats, summary.data.stats || {})
  } catch (error) { Message.error(apiMessage(error)) } finally { loading.value = false }
}
function onPageChange(value: number) { page.value = value; load() }
async function openDetail(record: any) {
  try {
    const { data } = await api.get(`/admin/verification/${record.verification_id}`)
    detail.value = data.verification || record; drawer.value = true; reason.value = ''
  } catch (error) { Message.error(apiMessage(error)) }
}
async function review(decision: string) {
  if (!detail.value) return
  try {
    await api.put(`/admin/verification/${detail.value.verification_id}/review`, { decision, reason: reason.value })
    Message.success('Review recorded'); drawer.value = false; load()
  } catch (error) { Message.error(apiMessage(error)) }
}
onMounted(load)
</script>
