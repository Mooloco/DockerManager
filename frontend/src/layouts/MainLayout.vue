<template>
  <el-container class="layout">
    <!-- 侧边栏 -->
    <el-aside width="220px" class="sidebar">
      <div class="logo">
        <el-icon :size="22"><Ship /></el-icon>
        <span>Docker Manager</span>
      </div>
      <el-menu :default-active="activeMenu" router class="menu">
        <el-menu-item index="/dashboard">
          <el-icon><Odometer /></el-icon>
          <span>总览</span>
        </el-menu-item>
        <el-menu-item index="/containers">
          <el-icon><Box /></el-icon>
          <span>容器</span>
        </el-menu-item>
        <el-menu-item index="/projects">
          <el-icon><Files /></el-icon>
          <span>项目</span>
        </el-menu-item>
        <el-menu-item index="/images">
          <el-icon><Picture /></el-icon>
          <span>镜像</span>
        </el-menu-item>
        <el-menu-item index="/networks">
          <el-icon><Share /></el-icon>
          <span>网络</span>
        </el-menu-item>
        <el-menu-item index="/volumes">
          <el-icon><FolderOpened /></el-icon>
          <span>卷</span>
        </el-menu-item>
        <el-menu-item index="/settings">
          <el-icon><Setting /></el-icon>
          <span>设置</span>
        </el-menu-item>
      </el-menu>
    </el-aside>

    <el-container>
      <!-- 顶栏 -->
      <el-header class="header" height="56px">
        <div class="header-left">
          <span class="page-title">{{ pageTitle }}</span>
        </div>
        <div class="header-right">
          <!-- 刷新频率 -->
          <el-dropdown trigger="click" @command="onRefreshInterval">
            <el-button text class="refresh-btn" title="页面自动刷新频率">
              <el-icon :size="16"><Refresh /></el-icon>
              <span class="refresh-label">刷新 {{ refresh.interval }}s</span>
              <el-icon :size="12"><ArrowDown /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item
                  v-for="opt in REFRESH_OPTIONS"
                  :key="opt.value"
                  :command="opt.value"
                  :class="{ active: refresh.interval === opt.value }"
                >
                  {{ opt.label }}
                  <el-icon v-if="refresh.interval === opt.value" class="check-icon"><Check /></el-icon>
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>

          <!-- 暗色模式 -->
          <el-tooltip :content="theme.dark ? '切换到浅色模式' : '切换到深色模式'">
            <el-button circle text @click="theme.toggle()">
              <el-icon :size="18">
                <Moon v-if="theme.dark" />
                <Sunny v-else />
              </el-icon>
            </el-button>
          </el-tooltip>
          <el-dropdown @command="onCommand">
            <span class="user-chip">
              <el-icon><User /></el-icon>
              {{ auth.username }}
              <el-icon><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="settings">
                  <el-icon><Setting /></el-icon>修改密码
                </el-dropdown-item>
                <el-dropdown-item command="logout" divided>
                  <el-icon><SwitchButton /></el-icon>退出登录
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>

      <!-- 内容区 -->
      <el-main class="main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '../stores/auth'
import { useThemeStore } from '../stores/theme'
import { useRefreshStore, REFRESH_OPTIONS } from '../stores/refresh'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const theme = useThemeStore()
const refresh = useRefreshStore()

const activeMenu = computed(() => {
  // 容器详情页高亮"容器"
  if (route.path.startsWith('/containers')) return '/containers'
  return route.path
})

const pageTitle = computed(() => (route.meta.title as string) || 'Docker Manager')

function onRefreshInterval(sec: number) {
  refresh.setInterval(sec)
  ElMessage.success(`刷新频率已设为 ${sec} 秒`)
}

async function onCommand(cmd: string) {
  if (cmd === 'logout') {
    await auth.logout()
    router.push('/login')
    ElMessage.success('已退出登录')
  } else if (cmd === 'settings') {
    router.push('/settings')
  }
}
</script>

<style scoped>
.layout {
  height: 100%;
}

.sidebar {
  background: var(--el-bg-color);
  border-right: 1px solid var(--el-border-color-lighter);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.logo {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 0 20px;
  height: var(--dm-header-height);
  font-size: 16px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  border-bottom: 1px solid var(--el-border-color-lighter);
  flex-shrink: 0;
}

.menu {
  border-right: none;
  flex: 1;
  overflow-y: auto;
}

.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: var(--el-bg-color);
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.page-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.refresh-btn {
  display: flex;
  align-items: center;
  gap: 4px;
  color: var(--el-text-color-regular);
}

.refresh-label {
  font-size: 13px;
}

:deep(.el-dropdown-menu__item.active) {
  color: var(--el-color-primary);
  font-weight: 600;
}

.check-icon {
  margin-left: 6px;
}

.user-chip {
  display: flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
  color: var(--el-text-color-primary);
  font-size: 14px;
  outline: none;
}

.main {
  padding: 0;
  overflow: auto;
}
</style>
