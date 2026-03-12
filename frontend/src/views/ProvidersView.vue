<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { providerApi, modelApi, type Provider, type ProviderModel } from '../api'

// ── Providers ──────────────────────────────────────────────────────────────
const providers = ref<Provider[]>([])
const providerLoading = ref(false)
const providerDialog = ref(false)
const providerForm = ref<Partial<Provider>>({})
const editingProvider = ref<number | null>(null)

async function loadProviders() {
  providerLoading.value = true
  const { data } = await providerApi.list()
  providers.value = data
  providerLoading.value = false
}

function openAddProvider() {
  editingProvider.value = null
  providerForm.value = { api_type: 'openai', enabled: true }
  providerDialog.value = true
}

function openEditProvider(p: Provider) {
  editingProvider.value = p.id
  providerForm.value = { ...p }
  providerDialog.value = true
}

async function saveProvider() {
  try {
    if (editingProvider.value) {
      await providerApi.update(editingProvider.value, providerForm.value)
    } else {
      await providerApi.create(providerForm.value)
    }
    ElMessage.success('保存成功')
    providerDialog.value = false
    loadProviders()
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || '保存失败')
  }
}

async function deleteProvider(id: number) {
  await ElMessageBox.confirm('确认删除该 Provider？', '提示', { type: 'warning' })
  await providerApi.remove(id)
  ElMessage.success('已删除')
  loadProviders()
}

// ── Models ─────────────────────────────────────────────────────────────────
const modelDialog = ref(false)
const modelForm = ref<Partial<ProviderModel>>({})
const models = ref<ProviderModel[]>([])
const currentProviderId = ref<number>(0)
const currentProviderName = ref('')
const editingModel = ref<number | null>(null)

async function openModels(p: Provider) {
  currentProviderId.value = p.id
  currentProviderName.value = p.name
  const { data } = await modelApi.list(p.id)
  models.value = data
  modelDialog.value = true
}

function openAddModel() {
  editingModel.value = null
  modelForm.value = { enabled: true }
}

function openEditModel(m: ProviderModel) {
  editingModel.value = m.id
  modelForm.value = { ...m }
}

async function saveModel() {
  try {
    if (editingModel.value) {
      await modelApi.update(editingModel.value, modelForm.value)
    } else {
      await modelApi.add(currentProviderId.value, modelForm.value)
    }
    ElMessage.success('保存成功')
    editingModel.value = null
    modelForm.value = {}
    const { data } = await modelApi.list(currentProviderId.value)
    models.value = data
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || '保存失败')
  }
}

async function deleteModel(id: number) {
  await ElMessageBox.confirm('确认删除该模型？', '提示', { type: 'warning' })
  await modelApi.remove(id)
  ElMessage.success('已删除')
  const { data } = await modelApi.list(currentProviderId.value)
  models.value = data
}

onMounted(loadProviders)
</script>

<template>
  <div>
    <div class="page-header">
      <h2>Providers</h2>
      <el-button type="primary" @click="openAddProvider">+ 添加 Provider</el-button>
    </div>

    <el-table :data="providers" v-loading="providerLoading" border stripe>
      <el-table-column prop="id" label="ID" width="60" />
      <el-table-column prop="name" label="名称" />
      <el-table-column prop="base_url" label="Base URL" />
      <el-table-column prop="api_type" label="类型" width="100" />
      <el-table-column label="状态" width="80">
        <template #default="{ row }">
          <el-tag :type="row.enabled ? 'success' : 'info'">{{ row.enabled ? '启用' : '禁用' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200">
        <template #default="{ row }">
          <el-button size="small" @click="openModels(row)">模型</el-button>
          <el-button size="small" type="primary" @click="openEditProvider(row)">编辑</el-button>
          <el-button size="small" type="danger" @click="deleteProvider(row.id)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- Provider Dialog -->
    <el-dialog v-model="providerDialog" :title="editingProvider ? '编辑 Provider' : '添加 Provider'" width="500px">
      <el-form :model="providerForm" label-width="100px">
        <el-form-item label="名称">
          <el-input v-model="providerForm.name" />
        </el-form-item>
        <el-form-item label="Base URL">
          <el-input v-model="providerForm.base_url" placeholder="https://api.example.com/v1" />
        </el-form-item>
        <el-form-item label="API Key">
          <el-input v-model="providerForm.api_key" show-password />
        </el-form-item>
        <el-form-item label="API 类型">
          <el-select v-model="providerForm.api_type">
            <el-option label="OpenAI 兼容" value="openai" />
            <el-option label="Anthropic 原生" value="anthropic" />
          </el-select>
        </el-form-item>
        <el-form-item label="启用">
          <el-switch v-model="providerForm.enabled" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="providerDialog = false">取消</el-button>
        <el-button type="primary" @click="saveProvider">保存</el-button>
      </template>
    </el-dialog>

    <!-- Models Dialog -->
    <el-dialog v-model="modelDialog" :title="`${currentProviderName} — 模型列表`" width="700px">
      <!-- Add/Edit inline form -->
      <el-form :model="modelForm" inline style="margin-bottom:16px">
        <el-form-item label="Model ID">
          <el-input v-model="modelForm.model_id" placeholder="gpt-4o" style="width:140px" />
        </el-form-item>
        <el-form-item label="上游 ID">
          <el-input v-model="modelForm.provider_model_id" placeholder="同上可留空" style="width:140px" />
        </el-form-item>
        <el-form-item label="显示名">
          <el-input v-model="modelForm.display_name" style="width:120px" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="saveModel">{{ editingModel ? '保存' : '添加' }}</el-button>
          <el-button v-if="editingModel" @click="editingModel = null; modelForm = {}">取消</el-button>
        </el-form-item>
      </el-form>

      <el-table :data="models" border stripe size="small">
        <el-table-column prop="model_id" label="Model ID" />
        <el-table-column prop="provider_model_id" label="上游 ID" />
        <el-table-column prop="display_name" label="显示名" />
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.enabled ? 'success' : 'info'" size="small">{{ row.enabled ? '启用' : '禁用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="130">
          <template #default="{ row }">
            <el-button size="small" @click="openEditModel(row)">编辑</el-button>
            <el-button size="small" type="danger" @click="deleteModel(row.id)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
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
