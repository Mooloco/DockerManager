<template>
  <div class="page-container">
    <!-- 工具栏 -->
    <div class="page-toolbar">
      <el-input
        v-model="search"
        placeholder="搜索镜像名 / ID"
        :prefix-icon="Search"
        clearable
        style="width: 260px"
      />
      <div class="spacer" />
      <span class="count-hint">共 {{ filtered.length }} 个镜像</span>
      <el-button :loading="loading" @click="load()">
        <el-icon><Refresh /></el-icon>&nbsp;刷新
      </el-button>
      <el-button type="primary" @click="pullDialog = true">
        <el-icon><Download /></el-icon>&nbsp;拉取镜像
      </el-button>
    </div>

    <!-- 镜像表格 -->
    <div class="table-card">
      <el-table :data="filtered" v-loading="loading" stripe>
        <el-table-column label="仓库:标签" min-width="240">
          <template #default="{ row }">
            <div v-if="row.repo_tags?.length" class="tags">
              <el-tag v-for="t in row.repo_tags" :key="t" size="small" effect="plain" class="tag">{{ t }}</el-tag>
            </div>
            <span v-else class="mono dim">&lt;none&gt;:&lt;none&gt;</span>
          </template>
        </el-table-column>
        <el-table-column label="ID" width="130">
          <template #default="{ row }"><span class="mono dim">{{ row.short_id }}</span></template>
        </el-table-column>
        <el-table-column label="大小" width="110" sortable prop="size">
          <template #default="{ row }">{{ formatBytes(row.size) }}</template>
        </el-table-column>
        <el-table-column label="创建时间" width="150">
          <template #default="{ row }">{{ formatRelative(row.created) }}</template>
        </el-table-column>
        <el-table-column label="被容器引用" width="100">
          <template #default="{ row }">
            <el-tag v-if="row.containers > 0" type="warning" size="small">{{ row.containers }}</el-tag>
            <span v-else class="dim">0</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="90" fixed="right">
          <template #default="{ row }">
            <el-tooltip content="删除镜像">
              <el-button size="small" type="danger" plain :icon="Delete" @click="onRemove(row)" />
            </el-tooltip>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 拉取镜像对话框 -->
    <el-dialog v-model="pullDialog" title="拉取镜像" width="560px" :close-on-click-modal="false" @closed="resetPull">
      <el-input
        v-model="pullRef"
        placeholder="例如: nginx:latest 或 registry.example.com/app:v1"
        size="large"
        @keyup.enter="startPull"
      />
      <div v-if="pulling || pullDone" class="pull-progress">
        <div v-for="(layer, idx) in pullLayers" :key="idx" class="layer">
          <div class="layer-head">
            <span class="mono layer-id">{{ layer.id || layer.status }}</span>
            <span class="layer-status" :class="{ done: layer.done, err: layer.error }">
              {{ layer.error ? '失败' : layer.done ? '完成' : layer.progress || layer.status || '等待中' }}
            </span>
          </div>
          <el-progress
            v-if="layer.total > 0 && !layer.done"
            :percentage="Math.min(Math.round((layer.current / layer.total) * 100), 100)"
            :stroke-width="6"
            :show-text="false"
            :status="layer.error ? 'exception' : undefined"
          />
        </div>
        <el-alert v-if="pullError" :title="pullError" type="error" :closable="false" show-icon />
      </div>
      <template #footer>
        <el-button @click="pullDialog = false" :disabled="pulling">关闭</el-button>
        <el-button v-if="pullDone" type="success" @click="pullDialog = false">
          <el-icon><CircleCheck /></el-icon>&nbsp;完成
        </el-button>
        <el-button v-else-if="!pulling" type="primary" :disabled="!pullRef" @click="startPull">
          <el-icon><Download /></el-icon>&nbsp;拉取
        </el-button>
        <el-button v-else type="danger" @click="stopPull">
          <el-icon><VideoPause /></el-icon>&nbsp;停止
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Refresh, Download, Delete, VideoPause, CircleCheck } from '@element-plus/icons-vue'
import { imageApi } from '../api'
import type { ImageItem } from '../api/types'
import { formatBytes, formatRelative } from '../utils/format'
import { WSClient } from '../websocket/ws'

const loading = ref(false)
const items = ref<ImageItem[]>([])
const search = ref('')

// Pull 状态
const pullDialog = ref(false)
const pullRef = ref('')
const pulling = ref(false)
const pullDone = ref(false)
const pullError = ref('')
interface PullLayer {
  id: string
  status: string
  progress: string
  current: number
  total: number
  done: boolean
  error: boolean
}
const pullLayers = ref<PullLayer[]>([])
let pullWS: WSClient | null = null

const filtered = computed(() => {
  if (!search.value) return items.value
  const q = search.value.toLowerCase()
  return items.value.filter((im) => {
    const tags = im.repo_tags.join(' ')
    return tags.toLowerCase().includes(q) || im.id.toLowerCase().includes(q)
  })
})

async function load() {
  loading.value = true
  try {
    items.value = await imageApi.list()
  } catch (e: any) {
    ElMessage.error(e.message || '获取镜像列表失败')
  } finally {
    loading.value = false
  }
}

async function onRemove(row: ImageItem) {
  const name = row.repo_tags?.[0] || row.short_id
  try {
    await ElMessageBox.confirm(
      `确定删除镜像 <b>${name}</b> 吗?${row.containers > 0 ? '<br/><span style="color:#e6a23c">该镜像正被 ' + row.containers + ' 个容器引用</span>' : ''}`,
      '删除镜像确认',
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
    await imageApi.remove(row.id, true)
    ElMessage.success('镜像已删除')
    load()
  } catch (e: any) {
    ElMessage.error(e.message || '删除失败')
  }
}

// ---- Pull ----
/** 校验镜像引用格式:Docker 官方规范(reference.ParseNormalizedNamed) */
function validateImageRef(ref: string): boolean {
  const s = ref.trim()
  if (!s) return false
  // 不允许空白、引号、反斜杠、管道等危险字符
  if (/[\s"'\\<>|]/.test(s)) return false
  // digest 形式: name@sha256:64位hex
  if (/^[a-z0-9][a-z0-9._/-]*@sha256:[a-f0-9]{64}$/.test(s)) return true
  // name[:tag] 形式(支持 registry/name)
  return /^[a-z0-9][a-z0-9._/-]*(:[a-zA-Z0-9._-]+)?$/.test(s)
}

function startPull() {
  const ref = pullRef.value.trim()
  if (!ref) {
    ElMessage.warning('请输入镜像名称')
    return
  }
  if (!validateImageRef(ref)) {
    ElMessage.warning('输入错误,请重新输入')
    pullRef.value = ''
    // 聚焦输入框
    requestAnimationFrame(() => {
      const input = document.querySelector('.el-dialog input') as HTMLInputElement | null
      input?.focus()
    })
    return
  }
  pulling.value = true
  pullDone.value = false
  pullError.value = ''
  pullLayers.value = []

  pullWS = new WSClient(`/api/v1/ws/images/pull?ref=${encodeURIComponent(ref)}`, {
    onMessage: (msg) => {
      if (msg.type === 'pull') {
        const d = msg.data as any
        if (d.error) {
          pullError.value = d.error
        } else {
          upsertLayer(d)
        }
      } else if (msg.type === 'end') {
        // 拉取结束:主动关闭连接,防止 WSClient 自动重连导致重复拉取
        pulling.value = false
        pullDone.value = true
        ElMessage.success('镜像拉取完成')
        load()
        pullWS?.close()
        pullWS = null
      } else if (msg.type === 'error') {
        pulling.value = false
        pullError.value = msg.message || '拉取失败'
        pullWS?.close()
        pullWS = null
      }
    },
    onClose: () => {
      pulling.value = false
    },
    onError: (e) => {
      pullError.value = e
      pulling.value = false
    },
  })
  pullWS.connect()
}

function upsertLayer(d: any) {
  // 分层进度:以 layer id 或 status 为 key
  const key = d.id || d.status
  let layer = pullLayers.value.find((l) => l.id === key)
  if (!layer) {
    layer = { id: key, status: d.status || '', progress: '', current: 0, total: 0, done: false, error: false }
    pullLayers.value.push(layer)
  }
  layer.status = d.status || layer.status
  layer.progress = d.progress || ''
  if (d.progressDetail?.total) {
    layer.current = d.progressDetail.current || 0
    layer.total = d.progressDetail.total
  }
  // 完成标记:状态含 "Downloaded"/"Pull complete"/"Digest"/"Status"
  if (/Downloaded|Pull complete|Digest:|Status:|Pulling fs layer/.test(d.status || '')) {
    if (d.status === 'Pulling fs layer') {
      // 准备阶段,不算完成
    } else {
      layer.done = true
    }
  }
  if (d.status === 'Downloading' && d.progressDetail?.current === d.progressDetail?.total) {
    layer.done = true
  }
}

function stopPull() {
  pullWS?.close()
  pullWS = null
  pulling.value = false
}

function resetPull() {
  pullWS?.close()
  pullWS = null
  pullRef.value = ''
  pulling.value = false
  pullDone.value = false
  pullError.value = ''
  pullLayers.value = []
}

onMounted(load)
</script>

<style scoped>
.count-hint {
  color: var(--el-text-color-secondary);
  font-size: 13px;
}

.tags {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.tag {
  max-width: 220px;
  overflow: hidden;
  text-overflow: ellipsis;
}

.dim {
  color: var(--el-text-color-secondary);
}

.pull-progress {
  margin-top: 16px;
  max-height: 320px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.layer-head {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 4px;
}

.layer-id {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.layer-status {
  font-size: 12px;
  color: var(--el-text-color-regular);
  flex-shrink: 0;
}

.layer-status.done {
  color: #67c23a;
}

.layer-status.err {
  color: #f56c6c;
}
</style>
