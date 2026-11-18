<template>
  <div class="page">
    <div class="card-title"><div><div class="eyebrow">Operations</div><h2 style="margin:6px 0 4px">Developers</h2><p class="muted">Accounts connected to this deployment.</p></div><a-button @click="load" :loading="loading">Refresh</a-button></div>
    <div class="panel" style="margin-top:18px">
      <a-table :data="rows" :columns="columns" :loading="loading" :pagination="{ pageSize: 20 }" row-key="id">
        <template #status="{ record }"><a-tag :color="record.status === 'suspended' ? 'red' : 'green'">{{ record.status || 'active' }}</a-tag></template>
      </a-table>
    </div>
  </div>
</template>
<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { Message } from '@arco-design/web-vue'
import { api, apiMessage } from '../lib/api'
const rows = ref<any[]>([]); const loading = ref(false)
const columns = [
  { title: 'Name', dataIndex: 'name' }, { title: 'Email', dataIndex: 'email' },
  { title: 'Company', dataIndex: 'company' }, { title: 'Status', slotName: 'status', width: 120 },
  { title: 'Created', dataIndex: 'created_at', width: 190 },
]
async function load() {
  loading.value = true
  try { const { data } = await api.get('/admin/developers', { params: { page: 1, limit: 100 } }); rows.value = data.developers || [] }
  catch (error) { Message.error(apiMessage(error)) } finally { loading.value = false }
}
onMounted(load)
</script>
