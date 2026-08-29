<template>
  <div class="page-container">
    <!-- 工具栏:第一行(搜索/过滤/刷新) -->
    <div class="page-toolbar">
      <el-input
        v-model="search"
        placeholder="搜索容器名称 / ID"
        :prefix-icon="Search"
        clearable
        style="width: 220px"
      />
      <el-select v-model="stateFilter" placeholder="状态过滤" clearable style="width: 130px">
        <el-option v-for="s in stateOptions" :key="s.value" :label="s.label" :value="s.value" />
      </el-select>
      <el-select v-model="allFilter" style="width: 120px">
        <el-option :value="true" label="全部容器" />
        <el-option :value="false" label="仅运行中" />
      </el-select>

      <div class="spacer" />
      <span class="count-hint">共 {{ filtered.length }} 个容器</span>
      <el-button :loading="loading" @click="load()">
        <el-icon><Refresh /></el-icon>&nbsp;刷新
      </el-button>
    </div>

    <!-- 工具栏:第二行(批量操作,仿 OpenWrt Dockerman) -->
    <div class="batch-bar">
      <span class="batch-label">批量操作</span>
      <el-button
        type="success"
        size="small"
        :disabled="!selected.length || !can('start')"
        @click="batchAction('start')"
      >
        <el-icon><VideoPlay /></el-icon>&nbsp;启动
      </el-button>
      <el-button size="small" :disabled="!selected.length || !can('stop')" @click="batchAction('stop')">
        <el-icon><VideoPause /></el-icon>&nbsp;停止
      </el-button>
      <el-button size="small" type="warning" :disabled="!selected.length || !can('restart')" @click="batchAction('restart')">
        <el-icon><RefreshRight /></el-icon>&nbsp;重启
      </el-button>
      <el-button size="small" :disabled="!selected.length || !can('pause')" @click="batchAction('pause')">
        <el-icon><VideoPause /></el-icon>&nbsp;暂停
      </el-button>
      <el-button size="small" type="success" plain :disabled="!selected.length || !can('unpause')" @click="batchAction('unpause')">
        <el-icon><VideoPlay /></el-icon>&nbsp;恢复
      </el-button>
      <el-button size="small" type="danger" :disabled="!selected.length || !can('kill')" @click="batchAction('kill')">
        <el-icon><CloseBold /></el-icon>&nbsp;强制终止
      </el-button>
      <el-button size="small" type="danger" plain :disabled="!selected.length" @click="onBatchRemove">
        <el-icon><Delete /></el-icon>&nbsp;删除
      </el-button>
      <span v-if="selected.length" class="batch-hint">已选 {{ selected.length }} 个容器</span>
    </div>

    <!-- 容器表格 -->
    <div class="table-card">
      <el-table
        ref="tableRef"
        :data="filtered"
        v-loading="loading"
        stripe
        row-key="id"
        @row-click="onRowClick"
        @selection-change="onSelectionChange"
        row-class-name="clickable-row"
      >
        <el-table-column type="selection" width="42" reserve-selection />
        <el-table-column label="名称" min-width="170">
          <template #default="{ row }">
            <div class="name-cell">
              <span class="state-dot" :style="{ background: dotColor(row.state) }" />
              <span class="name-text">{{ row.name || row.short_id }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="镜像" prop="image" min-width="160" show-overflow-tooltip />
        <el-table-column label="状态" width="96">
          <template #default="{ row }">
            <el-tag :type="statusType(row.state)" size="small">{{ statusText(row.state) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态详情" prop="status" min-width="130" show-overflow-tooltip />
        <el-table-column label="端口" min-width="150">
          <template #default="{ row }">
            <span class="mono ports">{{ formatPorts(row.ports) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="120">
          <template #default="{ row }">{{ formatRelative(row.created) }}</template>
        </el-table-column>
        <el-table-column label="ID" width="104">
          <template #default="{ row }">
            <span class="mono id-text">{{ row.short_id }}</span>
          </template>
        </el-table-column>
      </el-table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { TableInstance } from 'element-plus'
import {
  Search,
  Refresh,
  VideoPlay,
  VideoPause,
  RefreshRight,
  CloseBold,
  Delete,
} from '@element-plus/icons-vue'
import { containerApi } from '../api'
import type { ContainerItem } from '../api/types'
import { formatPorts, formatRelative, statusText, statusType } from '../utils/format'

const router = useRouter()
const tableRef = ref<TableInstance>()
const loading = ref(false)
const items = ref<ContainerItem[]>([])
const search = ref('')
const stateFilter = ref('')
const allFilter = ref(true)
const selected = ref<ContainerItem[]>([])

// 容器页自动刷新固定 1 分钟(不跟随顶栏刷新频率,避免频繁刷新打断勾选操作)
const AUTO_REFRESH_MS = 60_000

const stateOptions = [
  { value: 'running', label: '运行中' },
  { value: 'exited', label: '已退出' },
  { value: 'paused', label: '已暂停' },
  { value: 'restarting', label: '重启中' },
  { value: 'created', label: '已创建' },
  { value: 'dead', label: '异常' },
]

const filtered = computed(() => {
  let list = items.value
  if (stateFilter.value) {
    list = list.filter((c) => c.state === stateFilter.value)
  }
  if (search.value) {
    const q = search.value.toLowerCase()
    list = list.filter(
      (c) => c.name.toLowerCase().includes(q) || c.id.toLowerCase().includes(q),
    )
  }
  return list
})

function onSelectionChange(rows: ContainerItem[]) {
  selected.value = rows
}

/** 批量操作后清空选中 */
function clearSelection() {
  tableRef.value?.clearSelection()
}

/** 判断批量操作对当前选中容器是否可用(按状态) */
function can(action: string): boolean {
  if (!selected.value.length) return false
  return selected.value.some((c) => {
    switch (action) {
      case 'start':
        return ['exited', 'created', 'dead'].includes(c.state)
      case 'stop':
        return c.state === 'running'
      case 'restart':
        return ['running', 'exited'].includes(c.state)
      case 'pause':
        return c.state === 'running'
      case 'unpause':
        return c.state === 'paused'
      case 'kill':
        return ['running', 'paused'].includes(c.state)
      default:
        return false
    }
  })
}

async function load() {
  loading.value = true
  try {
    items.value = await containerApi.list(allFilter.value)
  } catch (e: any) {
    ElMessage.error(e.message || '获取容器列表失败')
  } finally {
    loading.value = false
  }
}

/** 批量操作:逐个执行并汇总结果 */
async function batchAction(action: string) {
  const targets = selected.value
  if (!targets.length) return
  let ok = 0
  const failed: string[] = []
  for (const c of targets) {
    try {
      await containerApi.action(c.id, action)
      ok++
    } catch {
      failed.push(c.name)
    }
  }
  const actionText: Record<string, string> = {
    start: '启动',
    stop: '停止',
    restart: '重启',
    pause: '暂停',
    unpause: '恢复',
    kill: '强制终止',
  }
  if (failed.length) {
    ElMessage.warning(`${actionText[action]}:成功 ${ok} 个,失败 ${failed.length} 个(${failed.join(', ')})`)
  } else {
    ElMessage.success(`${actionText[action]}成功:${ok} 个容器`)
  }
  clearSelection()
  load()
}

async function onBatchRemove() {
  const targets = selected.value
  if (!targets.length) return
  const names = targets.map((c) => c.name).join(', ')
  try {
    await ElMessageBox.confirm(
      `确定删除选中的 <b>${targets.length}</b> 个容器吗?<br/><span class="mono">${names}</span><br/>该操作不可恢复。`,
      '批量删除确认',
      {
        type: 'warning',
        confirmButtonText: '删除',
        cancelButtonText: '取消',
        confirmButtonClass: 'el-button--danger',
        dangerouslyUseHTMLString: true,
      },
    )
  } catch {
    return
  }
  let ok = 0
  const failed: string[] = []
  for (const c of targets) {
    try {
      await containerApi.action(c.id, 'remove', { force: true })
      ok++
    } catch {
      failed.push(c.name)
    }
  }
  if (failed.length) {
    ElMessage.warning(`删除:成功 ${ok} 个,失败 ${failed.length} 个(${failed.join(', ')})`)
  } else {
    ElMessage.success(`已删除 ${ok} 个容器`)
  }
  clearSelection()
  load()
}

function onRowClick(row: ContainerItem) {
  router.push(`/containers/${row.id}`)
}

function dotColor(state: string): string {
  switch (state) {
    case 'running':
      return '#67c23a'
    case 'paused':
      return '#e6a23c'
    case 'restarting':
      return '#e6a23c'
    case 'dead':
      return '#f56c6c'
    default:
      return '#909399'
  }
}

let timer: number | undefined
onMounted(() => {
  load()
  timer = window.setInterval(load, AUTO_REFRESH_MS)
})
onUnmounted(() => {
  if (timer) window.clearInterval(timer)
})
</script>

<style scoped>
.count-hint {
  color: var(--el-text-color-secondary);
  font-size: 13px;
}

/* 批量操作栏(第二行) */
.batch-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}

.batch-label {
  font-size: 13px;
  color: var(--el-text-color-secondary);
  margin-right: 4px;
}

.batch-hint {
  font-size: 13px;
  color: var(--el-color-primary);
  margin-left: 8px;
}

.clickable-row {
  cursor: pointer;
}

.name-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.state-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.name-text {
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
}

.id-text {
  color: var(--el-text-color-secondary);
}

.ports {
  color: var(--el-text-color-regular);
  font-size: 0.85em;
}
</style>
