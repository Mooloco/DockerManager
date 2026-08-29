<template>
  <div class="login-page">
    <el-card class="login-card" shadow="always">
      <div class="login-head">
        <el-icon :size="32" color="var(--el-color-primary)"><Ship /></el-icon>
        <h1>Docker Manager</h1>
        <p>Docker 单节点 Web 管理控制台</p>
      </div>

      <el-form @submit.prevent="onSubmit" :model="form" size="large">
        <el-form-item>
          <el-input v-model="form.username" placeholder="用户名" :prefix-icon="User" autofocus />
        </el-form-item>
        <el-form-item>
          <el-input
            v-model="form.password"
            type="password"
            placeholder="密码"
            :prefix-icon="Lock"
            show-password
            @keyup.enter="onSubmit"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" native-type="submit" class="login-btn" :loading="auth.loading">
            登 录
          </el-button>
        </el-form-item>
        <el-alert
          v-if="errorMsg"
          :title="errorMsg"
          type="error"
          :closable="false"
          show-icon
        />
      </el-form>

      <div class="login-foot">
        初始密码通过环境变量 <code>DM_ADMIN_PASSWORD</code> 设置
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { User, Lock } from '@element-plus/icons-vue'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()

const form = reactive({ username: '', password: '' })
const errorMsg = ref('')

onMounted(() => {
  if (route.query.expired) {
    errorMsg.value = '登录已过期,请重新登录'
  }
})

async function onSubmit() {
  errorMsg.value = ''
  if (!form.username || !form.password) {
    errorMsg.value = '请输入用户名和密码'
    return
  }
  try {
    await auth.login(form.username, form.password)
    ElMessage.success('登录成功')
    const redirect = (route.query.redirect as string) || '/dashboard'
    router.push(redirect)
  } catch (e: any) {
    errorMsg.value = e.message || '登录失败'
  }
}
</script>

<style scoped>
.login-page {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, var(--el-bg-color-page) 0%, var(--el-fill-color-light) 100%);
}

.login-card {
  width: 380px;
  padding: 8px 8px 0;
}

.login-head {
  text-align: center;
  margin-bottom: 24px;
}

.login-head h1 {
  margin: 12px 0 4px;
  font-size: 22px;
  color: var(--el-text-color-primary);
}

.login-head p {
  margin: 0;
  color: var(--el-text-color-secondary);
  font-size: 13px;
}

.login-btn {
  width: 100%;
}

.login-foot {
  margin-top: 8px;
  text-align: center;
  color: var(--el-text-color-placeholder);
  font-size: 12px;
}

.login-foot code {
  background: var(--el-fill-color-light);
  padding: 1px 6px;
  border-radius: 4px;
}
</style>
