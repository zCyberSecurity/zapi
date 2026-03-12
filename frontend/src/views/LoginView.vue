<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'

const token = ref('')
const router = useRouter()

function login() {
  if (!token.value.trim()) {
    ElMessage.warning('请输入 Admin Token')
    return
  }
  localStorage.setItem('admin_token', token.value.trim())
  router.push('/')
}
</script>

<template>
  <div class="login-wrap">
    <el-card class="login-card">
      <h2 style="text-align:center;margin-bottom:24px">zAPI 管理后台</h2>
      <el-form @submit.prevent="login">
        <el-form-item>
          <el-input
            v-model="token"
            placeholder="Admin Token"
            show-password
            size="large"
            @keyup.enter="login"
          />
        </el-form-item>
        <el-button type="primary" size="large" style="width:100%" @click="login">登录</el-button>
      </el-form>
    </el-card>
  </div>
</template>

<style scoped>
.login-wrap {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100vh;
  background: #f0f2f5;
}
.login-card {
  width: 360px;
}
</style>
