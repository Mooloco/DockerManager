<template>
  <div class="page-container">
    <div class="page-toolbar">
      <el-input v-model="search" placeholder="搜索卷名称" :prefix-icon="Search" clearable style="width: 260px" />
      <div class="spacer" />
      <span class="count-hint">共 {{ filtered.length }} 个卷</span>
      <el-button :loading="loading" @click="load()">
        <el-icon><Refresh /></el-icon>&nbsp;刷新
      </el-button>
    </div>

    <div class="table-card">
      <el-table :data="filtered" v-loading="loading" stripe>
        <el-table-column label="名称" prop="name" min-width="180" />
        <el-table-column label="驱动" prop="driver" width="110" />
        <el-table-column label="挂载点" min-width="220">
          <template #default="{ row }"><span class="mono dim">{{ row.mountpoint }}</span></template>
        </el-table-column>
        <el-table-column label="创建时间" width="160">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="被引用" width="90">
          <template #default="{ row }">
            <el-tag v-if="row.usage_data?.ref_count" type="warning" size="small">{{ row.usage_data.ref_count }}</el-tag>
            <span v-else class="dim">0</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="90" fixed="right">
          <template #default="{ row }">
            <el-tooltip content="删除卷">
              <el-button size="small" type="danger" plain :icon="Delete" @click="onRemove(row)" />
            </el-tooltip>
          </template>
        </el-table-column>
      </el-table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Refresh, Delete } from '@element-plus/icons-vue'
import { volumeApi } from '../api'
import type { VolumeItem } from '../api/types'
import { formatTime } from '../utils/format'

const loading = ref(false)
const items = ref<VolumeItem[]>([])
const search = ref('')

const filtered = computed(() => {
  if (!search.value) return items.value
  const q = search.value.toLowerCase()
  return items.value.filter((v) => v.name.toLowerCase().includes(q))
})

async function load() {
  loading.value = true
  try {
    items.value = await volumeApi.list()
  } catch (e: any) {
    ElMessage.error(e.message || '获取卷列表失败')
  } finally {
    loading.value = false
  }
}

async function onRemove(row: VolumeItem) {
  const inUse = (row.usage_data?.ref_count ?? 0) > 0
  try {
    await ElMessageBox.confirm(
      `确定删除卷 <b>${row.name}</b> 吗?<br/>删除后卷数据将丢失!` +
        (inUse ? '<br/><span style="color:#e6a23c">该卷正在被 ' + row.usage_data?.ref_count + ' 个容器使用</span>' : ''),
      '删除卷确认',
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
  try {
    await volumeApi.remove(row.name, true)
    ElMessage.success('卷已删除')
    load()
  } catch (e: any) {
    ElMessage.error(e.message || '删除失败')
  }
}

onMounted(load)
</script>

<style scoped>
.count-hint {
  color: var(--el-text-color-secondary);
  font-size: 13px;
}

.dim {
  color: var(--el-text-color-secondary);
}
</style>
