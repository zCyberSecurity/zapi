<script setup lang="ts">
import { useRouter } from 'vue-router'

const router = useRouter()

function logout() {
  localStorage.removeItem('admin_token')
  router.push('/login')
}
</script>

<template>
  <el-container style="height: 100vh">
    <el-aside width="200px" style="background: #001529">
      <div class="logo">zAPI</div>
      <el-menu
        router
        background-color="#001529"
        text-color="#ccc"
        active-text-color="#fff"
        :default-active="$route.path"
      >
        <el-menu-item index="/providers">
          <el-icon><Connection /></el-icon>
          <span>Providers</span>
        </el-menu-item>
        <el-menu-item index="/keys">
          <el-icon><Key /></el-icon>
          <span>API Keys</span>
        </el-menu-item>
        <el-menu-item index="/usage">
          <el-icon><DataLine /></el-icon>
          <span>Usage</span>
        </el-menu-item>
        <el-menu-item index="/test">
          <el-icon><ChatDotRound /></el-icon>
          <span>Test</span>
        </el-menu-item>
      </el-menu>
    </el-aside>

    <el-container>
      <el-header style="display:flex;align-items:center;justify-content:flex-end;border-bottom:1px solid #eee">
        <el-button link @click="logout">退出登录</el-button>
      </el-header>
      <el-main>
        <RouterView />
      </el-main>
    </el-container>
  </el-container>
</template>

<style scoped>
.logo {
  color: #fff;
  font-size: 20px;
  font-weight: bold;
  text-align: center;
  padding: 20px 0;
  letter-spacing: 2px;
}
</style>
