<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { usageApi, keyApi, type UsageLog, type APIKey } from '../api'

const logs = ref<UsageLog[]>([])
const loading = ref(false)
const keys = ref<APIKey[]>([])

const filter = ref<{ date: string; key_id: string }>({ date: '', key_id: '' })

async function load() {
  loading.value = true
  const params: { date?: string; key_id?: number } = {}
  if (filter.value.date) params.date = filter.value.date
  if (filter.value.key_id) params.key_id = Number(filter.value.key_id)
  const { data } = await usageApi.list(params)
  logs.value = data || []
  loading.value = false
}

function totalTokens() {
  return logs.value.reduce((s, r) => s + r.total_tokens, 0)
}

onMounted(async () => {
  const { data } = await keyApi.list()
  keys.value = data
  load()
})
</script>

<template>
  <div>
    <div class="page-header">
      <h2>Usage 统计</h2>
    </div>

    <el-form inline style="margin-bottom:16px">
      <el-form-item label="日期">
        <el-date-picker
          v-model="filter.date"
          type="date"
          value-format="YYYY-MM-DD"
          placeholder="全部日期"
          clearable
        />
      </el-form-item>
      <el-form-item label="API Key">
        <el-select v-model="filter.key_id" clearable placeholder="全部">
          <el-option v-for="k in keys" :key="k.id" :label="k.name || k.key" :value="k.id" />
        </el-select>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="load">查询</el-button>
      </el-form-item>
    </el-form>

    <el-table :data="logs" v-loading="loading" border stripe show-summary :summary-method="() => ['合计', '', '', '', totalTokens().toLocaleString(), '', '']">
      <el-table-column prop="date" label="日期" width="120" />
      <el-table-column prop="api_key_name" label="API Key" width="160" />
      <el-table-column prop="model" label="模型" />
      <el-table-column prop="request_count" label="请求次数" width="100" align="right" />
      <el-table-column prop="total_tokens" label="总 Token" width="120" align="right">
        <template #default="{ row }">{{ row.total_tokens.toLocaleString() }}</template>
      </el-table-column>
      <el-table-column prop="prompt_tokens" label="输入 Token" width="120" align="right">
        <template #default="{ row }">{{ row.prompt_tokens.toLocaleString() }}</template>
      </el-table-column>
      <el-table-column prop="completion_tokens" label="输出 Token" width="120" align="right">
        <template #default="{ row }">{{ row.completion_tokens.toLocaleString() }}</template>
      </el-table-column>
    </el-table>
  </div>
</template>

<style scoped>
.page-header {
  display: flex;
  align-items: center;
  margin-bottom: 16px;
}
</style>
