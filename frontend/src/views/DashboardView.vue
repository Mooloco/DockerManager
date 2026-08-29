<template>
  <div class="page-container">
    <el-alert
      v-if="errorMsg"
      :title="errorMsg"
      type="error"
      show-icon
      :closable="false"
      class="mb-16"
    />

    <el-row :gutter="16" class="mb-16">
      <el-col :span="6">
        <el-card shadow="never" class="stat-card">
          <div class="stat-icon running"><el-icon :size="22"><CaretRight /></el-icon></div>
          <div>
            <div class="stat-value">{{ info?.containers_running ?? '-' }}</div>
            <div class="stat-label">运行中容器</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never" class="stat-card">
          <div class="stat-icon stopped"><el-icon :size="22"><VideoPause /></el-icon></div>
          <div>
            <div class="stat-value">{{ info?.containers_stopped ?? '-' }}</div>
            <div class="stat-label">已停止容器</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never" class="stat-card">
          <div class="stat-icon total"><el-icon :size="22"><Box /></el-icon></div>
          <div>
            <div class="stat-value">{{ info?.containers ?? '-' }}</div>
            <div class="stat-label">容器总数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never" class="stat-card">
          <div class="stat-icon image"><el-icon :size="22"><Picture /></el-icon></div>
          <div>
            <div class="stat-value">{{ info?.images ?? '-' }}</div>
            <div class="stat-label">镜像总数</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="16">
      <!-- Docker 引擎信息 -->
      <el-col :span="12">
        <el-card shadow="never" class="mb-16">
          <template #header>
            <div class="card-header">
              <el-icon><Ship /></el-icon>
              <span>Docker Engine</span>
            </div>
          </template>
          <el-descriptions :column="2" border>
            <el-descriptions-item label="服务端版本">{{ info?.server_version || '-' }}</el-descriptions-item>
            <el-descriptions-item label="API 版本">{{ info?.api_version || '-' }}</el-descriptions-item>
            <el-descriptions-item label="操作系统">{{ info?.operating_system || '-' }}</el-descriptions-item>
            <el-descriptions-item label="架构">{{ info?.architecture || '-' }}</el-descriptions-item>
            <el-descriptions-item label="内核版本">{{ info?.kernel_version || '-' }}</el-descriptions-item>
            <el-descriptions-item label="存储驱动">{{ info?.driver || '-' }}</el-descriptions-item>
            <el-descriptions-item label="主机名">{{ info?.name || '-' }}</el-descriptions-item>
            <el-descriptions-item label="CPU 核数">{{ info?.cpus ?? '-' }}</el-descriptions-item>
            <el-descriptions-item label="总内存">{{ formatBytes(info?.total_memory) }}</el-descriptions-item>
            <el-descriptions-item label="Root Dir" class="mono-cell">{{ info?.root_dir || '-' }}</el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>

      <!-- 容器状态分布 -->
      <el-col :span="12">
        <el-card shadow="never" class="mb-16">
          <template #header>
            <div class="card-header">
              <el-icon><Histogram /></el-icon>
              <span>容器状态分布</span>
            </div>
          </template>
          <div class="dist-list">
            <div v-for="d in distribution" :key="d.label" class="dist-row">
              <span class="dist-label">{{ d.label }}</span>
              <el-progress
                :percentage="d.pct"
                :color="d.color"
                :stroke-width="14"
                :show-text="true"
                :format="() => `${d.value}`"
              />
            </div>
          </div>
          <el-empty v-if="!info?.containers" description="当前没有容器" :image-size="60" />
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="16">
      <el-col :span="24">
        <el-card shadow="never">
          <template #header>
            <div class="card-header">
              <el-icon><Timer /></el-icon>
              <span>快速操作</span>
            </div>
          </template>
          <div class="quick-actions">
            <el-button type="primary" @click="router.push('/containers')">
              <el-icon><Box /></el-icon>&nbsp;管理容器
            </el-button>
            <el-button @click="router.push('/images')">
              <el-icon><Download /></el-icon>&nbsp;拉取镜像
            </el-button>
            <el-button @click="refresh()" :loading="loading">
              <el-icon><Refresh /></el-icon>&nbsp;刷新数据
            </el-button>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { systemApi } from '../api'
import type { SystemInfo } from '../api/types'
import { formatBytes } from '../utils/format'
import { useRefreshStore } from '../stores/refresh'

const router = useRouter()
const refreshStore = useRefreshStore()
const info = ref<SystemInfo | null>(null)
const loading = ref(false)
const errorMsg = ref('')

const distribution = computed(() => {
  const total = info.value?.containers ?? 0
  if (!total) return []
  const mk = (label: string, value: number, color: string) => ({
    label,
    value,
    pct: total ? Math.round((value / total) * 100) : 0,
    color,
  })
  return [
    mk('运行中', info.value?.containers_running ?? 0, '#67c23a'),
    mk('已暂停', info.value?.containers_paused ?? 0, '#e6a23c'),
    mk('已停止', info.value?.containers_stopped ?? 0, '#909399'),
  ]
})

async function refresh() {
  loading.value = true
  errorMsg.value = ''
  try {
    info.value = await systemApi.info()
  } catch (e: any) {
    errorMsg.value = e.message || '无法获取 Docker 信息'
  } finally {
    loading.value = false
  }
}

let timer: number | undefined
function setupTimer() {
  if (timer) window.clearInterval(timer)
  timer = window.setInterval(refresh, refreshStore.interval * 1000)
}
onMounted(() => {
  refresh()
  setupTimer()
})
onUnmounted(() => {
  if (timer) window.clearInterval(timer)
})
// 刷新频率变化时重建定时器
watch(() => refreshStore.interval, setupTimer)
</script>

<style scoped>
.mb-16 {
  margin-bottom: 16px;
}

.stat-card {
  display: flex;
  align-items: center;
}

.stat-card :deep(.el-card__body) {
  display: flex;
  align-items: center;
  gap: 16px;
  width: 100%;
}

.stat-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.stat-icon.running {
  background: rgba(103, 194, 58, 0.15);
  color: #67c23a;
}
.stat-icon.stopped {
  background: rgba(144, 147, 153, 0.15);
  color: #909399;
}
.stat-icon.total {
  background: rgba(64, 158, 255, 0.15);
  color: #409eff;
}
.stat-icon.image {
  background: rgba(230, 162, 60, 0.15);
  color: #e6a23c;
}

.stat-value {
  font-size: 26px;
  font-weight: 700;
  line-height: 1.2;
  color: var(--el-text-color-primary);
}

.stat-label {
  font-size: 13px;
  color: var(--el-text-color-secondary);
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
}

.mono-cell :deep(span) {
  font-family: Consolas, monospace;
  font-size: 0.9em;
}

.dist-list {
  display: flex;
  flex-direction: column;
  gap: 18px;
  padding: 4px 0;
}

.dist-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.dist-label {
  width: 56px;
  font-size: 13px;
  color: var(--el-text-color-secondary);
  flex-shrink: 0;
}

.dist-row .el-progress {
  flex: 1;
}

.quick-actions {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}
</style>
