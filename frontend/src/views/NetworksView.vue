<template>
  <div class="page-container">
    <!-- 工具栏:第一行 -->
    <div class="page-toolbar">
      <el-input v-model="search" placeholder="搜索网络名称 / ID" :prefix-icon="Search" clearable style="width: 260px" />
      <div class="spacer" />
      <span class="count-hint">共 {{ filtered.length }} 个网络</span>
      <el-button :loading="loading" @click="load()">
        <el-icon><Refresh /></el-icon>&nbsp;刷新
      </el-button>
    </div>

    <!-- 工具栏:第二行(批量操作) -->
    <div class="batch-bar">
      <span class="batch-label">网络操作</span>
      <el-button size="small" type="danger" plain :disabled="!selected.length" @click="batchRemove">
        <el-icon><Delete /></el-icon>&nbsp;删除
      </el-button>
      <span v-if="selected.length" class="batch-hint">已选 {{ selected.length }} 个网络</span>
    </div>

    <div class="table-card">
      <el-table
        ref="tableRef"
        :data="filtered"
        v-loading="loading"
        stripe
        row-key="id"
        @selection-change="onSelectionChange"
      >
        <el-table-column type="selection" width="42" reserve-selection />
        <el-table-column label="名称" min-width="180">
          <template #default="{ row }">
            <div class="name-cell" @click="router.push(`/networks/${row.id}`)">
              <span class="name-text">{{ row.name }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="驱动" prop="driver" width="110" />
        <el-table-column label="作用域" prop="scope" width="90" />
        <el-table-column label="ID" width="130">
          <template #default="{ row }"><span class="mono dim">{{ row.short_id }}</span></template>
        </el-table-column>
        <el-table-column label="特性" width="200">
          <template #default="{ row }">
            <el-tag v-if="row.internal" size="small" type="warning" class="feature-tag">internal</el-tag>
            <el-tag v-if="row.ipv6" size="small" type="info" class="feature-tag">IPv6</el-tag>
            <el-tag v-if="row.attachable" size="small" type="success" class="feature-tag">attachable</el-tag>
            <span v-if="!row.internal && !row.ipv6 && !row.attachable" class="dim">-</span>
          </template>
        </el-table-column>
      </el-table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { TableInstance } from 'element-plus'
import { Search, Refresh, Delete } from '@element-plus/icons-vue'
import { networkApi } from '../api'
import type { NetworkItem } from '../api/types'

const router = useRouter()
const tableRef = ref<TableInstance>()
const loading = ref(false)
const items = ref<NetworkItem[]>([])
const search = ref('')
const selected = ref<NetworkItem[]>([])

const filtered = computed(() => {
  if (!search.value) return items.value
  const q = search.value.toLowerCase()
  return items.value.filter((n) => n.name.toLowerCase().includes(q) || n.id.toLowerCase().includes(q))
})

async function load() {
  loading.value = true
  try {
    items.value = await networkApi.list()
  } catch (e: any) {
    ElMessage.error(e.message || '获取网络列表失败')
  } finally {
    loading.value = false
  }
}

function onSelectionChange(rows: NetworkItem[]) {
  selected.value = rows
}

/** 批量删除:逐个执行,被运行中容器引用的网络会提示并跳过 */
async function batchRemove() {
  const targets = selected.value
  if (!targets.length) return
  const names = targets.map((n) => n.name).join(', ')
  try {
    await ElMessageBox.confirm(
      `确定删除选中的 <b>${targets.length}</b> 个网络吗?<br/><span class="mono">${names}</span><br/><span style="color:#e6a23c">被运行中容器引用的网络无法删除</span>`,
      '删除网络确认',
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
  const failed: { name: string; msg: string }[] = []
  for (const n of targets) {
    try {
      await networkApi.remove(n.id)
      ok++
    } catch (e: any) {
      failed.push({ name: n.name, msg: e.message || '删除失败' })
    }
  }
  if (failed.length) {
    ElMessage.warning(`删除:成功 ${ok} 个,${failed.length} 个未删除:\n${failed.map((f) => `- ${f.name}: ${f.msg}`).join('\n')}`)
  } else {
    ElMessage.success(`已删除 ${ok} 个网络`)
  }
  tableRef.value?.clearSelection()
  load()
}

onMounted(load)
</script>

<style scoped>
.count-hint {
  color: var(--el-text-color-secondary);
  font-size: 13px;
}

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

.name-cell {
  cursor: pointer;
}

.name-text {
  font-weight: 500;
  color: var(--el-color-primary);
}

.dim {
  color: var(--el-text-color-secondary);
}

.feature-tag {
  margin-right: 4px;
}
</style>
