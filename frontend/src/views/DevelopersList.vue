<template>
  <div class="page">
    <a-page-header
      :show-back="false"
      title="Developers"
      subtitle="Accounts connected to this deployment."
    >
      <template #extra>
        <a-space>
          <a-button :loading="loading" @click="load">
            <template #icon><icon-refresh /></template>
            Refresh
          </a-button>
        </a-space>
      </template>
    </a-page-header>

    <a-grid :cols="{ xs: 2, sm: 2, md: 4 }" :col-gap="12" :row-gap="12">
      <a-grid-item v-for="card in statCards" :key="card.key">
        <a-card size="small">
          <a-statistic :title="card.label" :value="card.value" :value-style="card.style" />
        </a-card>
      </a-grid-item>
    </a-grid>

    <a-card style="margin-top:16px">
      <a-space wrap style="margin-bottom:14px">
        <a-input
          v-model="search"
          allow-clear
          placeholder="Search name, email or company"
          style="width:320px"
        >
          <template #prefix><icon-search /></template>
        </a-input>
        <div class="spacer" />
        <a-typography-text type="secondary">
          {{ filteredRows.length }} of {{ total }} developers
        </a-typography-text>
      </a-space>

      <a-alert v-if="error" type="error" :show-icon="true" style="margin-bottom:12px">
        <a-space>
          <span>{{ error }}</span>
          <a-link @click="load">Retry</a-link>
        </a-space>
      </a-alert>

      <a-table
        v-model:expanded-keys="expandedKeys"
        :data="filteredRows"
        :columns="columns"
        :loading="loading"
        :pagination="false"
        :bordered="{ cell: true }"
        :expandable="{ width: 36 }"
        :scroll="{ x: 980 }"
        row-key="id"
        @row-click="onRowClick"
      >
        <template #name="{ record }">
          <span>{{ record.name || record.full_name || '—' }}</span>
        </template>
        <template #email="{ record }">
          <span class="mono">{{ record.email || '—' }}</span>
        </template>
        <template #company="{ record }">
          <span>{{ record.company || '—' }}</span>
        </template>
        <template #status="{ record }">
          <a-tag :color="record.status === 'suspended' ? 'red' : 'green'">
            {{ record.status === 'suspended' ? 'Suspended' : 'Active' }}
          </a-tag>
        </template>
        <template #created="{ record }">
          <span>{{ formatDate(record.created_at) }}</span>
        </template>
        <template #actions="{ record }">
          <a-button size="mini" type="text" @click.stop="toggleRow(record)">
            {{ isExpanded(record) ? 'Hide' : 'Details' }}
          </a-button>
        </template>
        <template #expand-row="{ record }">
          <div style="padding:4px 0 8px">
            <div class="stat-label" style="margin-bottom:8px">Developer details</div>
            <a-descriptions :column="2" size="small" bordered :label-style="{ width: '130px' }">
              <a-descriptions-item label="Developer ID">
                <span class="mono">{{ record.id }}</span>
              </a-descriptions-item>
              <a-descriptions-item label="Email">
                <span class="mono">{{ record.email || '—' }}</span>
              </a-descriptions-item>
              <a-descriptions-item label="Name">{{ record.name || record.full_name || '—' }}</a-descriptions-item>
              <a-descriptions-item label="Company">{{ record.company || '—' }}</a-descriptions-item>
              <a-descriptions-item label="Status">
                <a-tag :color="record.status === 'suspended' ? 'red' : 'green'">
                  {{ record.status === 'suspended' ? 'Suspended' : 'Active' }}
                </a-tag>
              </a-descriptions-item>
              <a-descriptions-item label="Signed up">{{ formatDate(record.created_at) }}</a-descriptions-item>
            </a-descriptions>

            <div class="stat-label" style="margin:16px 0 8px">Account status</div>
            <a-space wrap>
              <a-typography-text type="secondary">
                Suspending blocks new verification calls signed with this developer's API keys.
              </a-typography-text>
              <a-tooltip content="Developer status changes are not exposed to admin tokens in this deployment.">
                <a-button type="outline" status="danger" @click="notifySuspend">
                  {{ record.status === 'suspended' ? 'Reactivate developer' : 'Suspend developer' }}
                </a-button>
              </a-tooltip>
            </a-space>
          </div>
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
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { Message } from '@arco-design/web-vue'
import { api, apiMessage } from '../lib/api'

const LIMIT = 25

const rows = ref<any[]>([])
const loading = ref(false)
const error = ref('')
const search = ref('')
const page = ref(1)
const limit = ref(LIMIT)
const total = ref(0)
const expandedKeys = ref<string[]>([])

const columns = [
  { title: 'Name', slotName: 'name', width: 220 },
  { title: 'Email', slotName: 'email', width: 240 },
  { title: 'Company', slotName: 'company', width: 200 },
  { title: 'Status', slotName: 'status', width: 130 },
  { title: 'Signed up', slotName: 'created', width: 160 },
  { title: 'Actions', slotName: 'actions', width: 100, fixed: 'right' as const },
]

const filteredRows = computed(() => {
  const query = search.value.trim().toLowerCase()
  if (!query) return rows.value
  return rows.value.filter((row) =>
    String(row.name || row.full_name || '').toLowerCase().includes(query) ||
    String(row.email || '').toLowerCase().includes(query) ||
    String(row.company || '').toLowerCase().includes(query),
  )
})

const statCards = computed(() => {
  const active = rows.value.filter((row) => row.status !== 'suspended').length
  const suspended = rows.value.filter((row) => row.status === 'suspended').length
  return [
    { key: 'total', label: 'Total developers', value: total.value, style: { color: '#0f172a' } },
    { key: 'active', label: 'Active', value: active, style: { color: '#15803d' } },
    { key: 'suspended', label: 'Suspended', value: suspended, style: { color: '#dc2626' } },
    { key: 'listed', label: 'Listed on page', value: filteredRows.value.length, style: { color: '#2563eb' } },
  ]
})

function formatDate(value?: string | null) {
  if (!value) return '—'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '—'
  return date.toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' })
}

function isExpanded(record: any) {
  return expandedKeys.value.includes(String(record.id))
}

function toggleRow(record: any) {
  const id = String(record.id)
  expandedKeys.value = expandedKeys.value.includes(id) ? [] : [id]
}

function onRowClick(record: any) {
  toggleRow(record)
}

function onPageChange(value: number) {
  page.value = value
  load()
}

function notifySuspend() {
  Message.info('Developer status changes need the service-token admin endpoint, which this console does not use yet.')
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    const { data } = await api.get('/admin/developers', { params: { page: page.value, limit: limit.value } })
    rows.value = data.developers || []
    total.value = data.pagination?.total || 0
  } catch (caught) {
    error.value = apiMessage(caught)
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>
