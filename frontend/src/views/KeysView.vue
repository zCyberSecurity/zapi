<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { keyApi, type APIKey } from '../api'

const keys = ref<APIKey[]>([])
const loading = ref(false)
const dialog = ref(false)
const editingId = ref<number | null>(null)

const form = ref<{
  name: string
  allowed_models_input: string
  enabled: boolean
}>({ name: '', allowed_models_input: '', enabled: true })

async function load() {
  loading.value = true
  const { data } = await keyApi.list()
  keys.value = data
  loading.value = false
}

function openAdd() {
  editingId.value = null
  form.value = { name: '', allowed_models_input: '', enabled: true }
  dialog.value = true
}

function openEdit(k: APIKey) {
  editingId.value = k.id
  let models: string[] = []
  try { models = JSON.parse(k.allowed_models || '[]') } catch {}
  form.value = { name: k.name, allowed_models_input: models.join('\n'), enabled: k.enabled }
  dialog.value = true
}

function parseModels(input: string): string[] {
  return input.split(/[\n,]+/).map(s => s.trim()).filter(Boolean)
}

async function save() {
  const models = parseModels(form.value.allowed_models_input)
  try {
    if (editingId.value) {
      await keyApi.update(editingId.value, {
        name: form.value.name,
        allowed_models: models,
        enabled: form.value.enabled,
      })
    } else {
      await keyApi.create({ name: form.value.name, allowed_models: models })
    }
    ElMessage.success('保存成功')
    dialog.value = false
    load()
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || '保存失败')
  }
}

async function remove(id: number) {
  await ElMessageBox.confirm('确认删除该 API Key？', '提示', { type: 'warning' })
  await keyApi.remove(id)
  ElMessage.success('已删除')
  load()
}

function copyKey(key: string) {
  navigator.clipboard.writeText(key)
  ElMessage.success('已复制')
}

function formatModels(raw: string): string {
  try {
    const arr = JSON.parse(raw || '[]')
    return arr.length === 0 ? '全部' : arr.join(', ')
  } catch {
    return raw || '全部'
  }
}

onMounted(load)
</script>

<template>
  <div>
    <div class="page-header">
      <h2>API Keys</h2>
      <el-button type="primary" @click="openAdd">+ 创建 Key</el-button>
    </div>

    <el-table :data="keys" v-loading="loading" border stripe>
      <el-table-column prop="id" label="ID" width="60" />
      <el-table-column prop="name" label="名称" width="160" />
      <el-table-column label="Key">
        <template #default="{ row }">
          <el-text truncated style="max-width:300px">{{ row.key }}</el-text>
          <el-button link size="small" @click="copyKey(row.key)" style="margin-left:8px">复制</el-button>
        </template>
      </el-table-column>
      <el-table-column label="可用模型">
        <template #default="{ row }">{{ formatModels(row.allowed_models) }}</template>
      </el-table-column>
      <el-table-column label="状态" width="80">
        <template #default="{ row }">
          <el-tag :type="row.enabled ? 'success' : 'info'">{{ row.enabled ? '启用' : '禁用' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="创建时间" width="180" />
      <el-table-column label="操作" width="130">
        <template #default="{ row }">
          <el-button size="small" type="primary" @click="openEdit(row)">编辑</el-button>
          <el-button size="small" type="danger" @click="remove(row.id)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialog" :title="editingId ? '编辑 Key' : '创建 API Key'" width="460px">
      <el-form :model="form" label-width="90px">
        <el-form-item label="名称">
          <el-input v-model="form.name" placeholder="如：dev-key" />
        </el-form-item>
        <el-form-item label="可用模型">
          <el-input
            v-model="form.allowed_models_input"
            type="textarea"
            :rows="4"
            placeholder="每行或逗号分隔一个 Model ID，留空表示全部可用"
          />
        </el-form-item>
        <el-form-item v-if="editingId" label="启用">
          <el-switch v-model="form.enabled" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialog = false">取消</el-button>
        <el-button type="primary" @click="save">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}
</style>
