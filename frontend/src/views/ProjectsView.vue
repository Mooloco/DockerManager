<template>
  <div class="page-container">
    <!-- 工具栏:第一行 -->
    <div class="page-toolbar">
      <el-input
        v-model="search"
        placeholder="搜索项目名"
        :prefix-icon="Search"
        clearable
        style="width: 240px"
      />
      <div class="spacer" />
      <span class="count-hint">共 {{ filtered.length }} 个项目</span>
      <el-button :loading="loading" @click="load()">
        <el-icon><Refresh /></el-icon>&nbsp;刷新
      </el-button>
      <el-button type="primary" @click="openCreate">
        <el-icon><Plus /></el-icon>&nbsp;新建项目
      </el-button>
    </div>

    <!-- 工具栏:第二行(批量操作,作用于勾选的项目) -->
    <div class="batch-bar">
      <span class="batch-label">项目操作</span>
      <el-button size="small" type="success" :disabled="!selected.length || !can('start')" @click="batchAction('start')">
        <el-icon><VideoPlay /></el-icon>&nbsp;启动
      </el-button>
      <el-button size="small" :disabled="!selected.length || !can('stop')" @click="batchAction('stop')">
        <el-icon><VideoPause /></el-icon>&nbsp;停止
      </el-button>
      <el-button size="small" type="warning" :disabled="!selected.length || !can('restart')" @click="batchAction('restart')">
        <el-icon><RefreshRight /></el-icon>&nbsp;重启
      </el-button>
      <el-button size="small" type="warning" plain :disabled="!selected.length" @click="batchRebuild">
        <el-icon><Refresh /></el-icon>&nbsp;重建
      </el-button>
      <el-button size="small" type="danger" plain :disabled="!selected.length" @click="batchRemove">
        <el-icon><Delete /></el-icon>&nbsp;删除
      </el-button>
      <span v-if="selected.length" class="batch-hint">已选 {{ selected.length }} 个项目</span>
    </div>

    <el-alert
      v-if="composeUnavailable"
      title="docker compose CLI 不可用,项目只能查看,无法执行启动/停止等操作"
      type="warning"
      show-icon
      :closable="false"
      class="mb-16"
    />

    <div class="table-card">
      <el-table
        ref="tableRef"
        :data="filtered"
        v-loading="loading"
        stripe
        row-key="name"
        @selection-change="onSelectionChange"
      >
        <el-table-column type="selection" width="42" reserve-selection />
        <el-table-column label="名称" min-width="150">
          <template #default="{ row }">
            <div class="name-cell" @click="router.push(`/projects/${row.name}`)">
              <span class="state-dot" :style="{ background: dotColor(row) }" />
              <span class="name-text">{{ row.name }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="来源" width="90">
          <template #default="{ row }">
            <el-tag :type="row.source === 'managed' ? 'primary' : 'info'" size="small" effect="plain">
              {{ row.source === 'managed' ? '本工具' : '已有' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="服务" min-width="140">
          <template #default="{ row }">
            <span class="svc-list">{{ row.services?.join(', ') || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="容器" width="110">
          <template #default="{ row }">
            <span v-if="row.has_containers">
              <span class="running">{{ row.running }}</span> / {{ row.containers }}
            </span>
            <el-tag v-else type="warning" size="small">已停止</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="compose 文件" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">
            <span class="mono dim">{{ row.compose_file || row.config_files?.join(', ') || '-' }}</span>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 新建项目 -->
    <el-dialog v-model="createVisible" title="新建 Compose 项目" width="700px" :close-on-click-modal="false">
      <el-form label-width="80px">
        <el-form-item label="项目名" required>
          <el-input v-model="createForm.name" placeholder="如 navi(仅字母数字 - _ .)" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="createForm.description" placeholder="可选" />
        </el-form-item>

        <!-- docker run 转 compose -->
        <el-form-item label="docker run">
          <div class="convert-box">
            <el-input
              v-model="createForm.dockerRun"
              placeholder="粘贴 docker run 命令,如: docker run -d --name nginx -p 8080:80 -v /data:/html nginx:latest"
              clearable
            >
              <template #append>
                <el-button :loading="converting" @click="doConvert">转换为 Compose</el-button>
              </template>
            </el-input>
            <div class="convert-tip">
              输入 docker run 命令后点击"转换为 Compose",自动生成下方的 compose YAML,可继续编辑
            </div>
          </div>
        </el-form-item>

        <el-form-item label="compose" required>
          <el-input
            v-model="createForm.yaml"
            type="textarea"
            :rows="14"
            class="mono"
            placeholder="services:
  app:
    image: nginx:latest
    ports:
      - '8080:80'"
          />
        </el-form-item>
        <el-form-item>
          <el-checkbox v-model="createForm.start">保存后立即启动 (up -d)</el-checkbox>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createVisible = false">取消</el-button>
        <el-button type="primary" :loading="creating" @click="doCreate">保存</el-button>
      </template>
    </el-dialog>

    <!-- 重建日志弹窗 -->
    <el-dialog v-model="rebuildVisible" :title="rebuildTitle" width="640px" :close-on-click-modal="false">
      <pre class="rebuild-log" v-loading="rebuilding">{{ rebuildOutput || '(执行中...)' }}</pre>
      <template #footer>
        <el-button type="primary" @click="rebuildVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { TableInstance } from 'element-plus'
import { Search, Refresh, Plus, VideoPlay, VideoPause, RefreshRight, Delete } from '@element-plus/icons-vue'
import { projectApi, composeApi } from '../api'
import type { Project } from '../api/types'

const router = useRouter()
const tableRef = ref<TableInstance>()
const loading = ref(false)
const items = ref<Project[]>([])
const search = ref('')
const composeUnavailable = ref(false)
const selected = ref<Project[]>([])

const createVisible = ref(false)
const creating = ref(false)
const converting = ref(false)
const createForm = ref({ name: '', description: '', yaml: '', dockerRun: '', start: true })

const rebuildVisible = ref(false)
const rebuilding = ref(false)
const rebuildOutput = ref('')
const rebuildTitle = ref('重建日志')

const filtered = computed(() => {
  if (!search.value) return items.value
  return items.value.filter((p) => p.name.toLowerCase().includes(search.value.toLowerCase()))
})

async function load() {
  loading.value = true
  try {
    items.value = await projectApi.list()
  } catch (e: any) {
    if (e.message?.includes('Docker')) {
      ElMessage.error(e.message)
    } else {
      composeUnavailable.value = true
    }
  } finally {
    loading.value = false
  }
}

function onSelectionChange(rows: Project[]) {
  selected.value = rows
}

/** 批量操作可用性:选中项目里有符合状态的即可 */
function can(action: string): boolean {
  if (!selected.value.length) return false
  return selected.value.some((p) => {
    switch (action) {
      case 'start':
        return p.running === 0
      case 'stop':
      case 'restart':
        return p.running > 0
      default:
        return false
    }
  })
}

async function loadAfter() {
  await load()
  tableRef.value?.clearSelection()
}

function openCreate() {
  createForm.value = { name: '', description: '', yaml: '', dockerRun: '', start: true }
  createVisible.value = true
}

/** docker run → compose YAML */
async function doConvert() {
  const cmd = createForm.value.dockerRun.trim()
  if (!cmd) {
    ElMessage.warning('请先输入 docker run 命令')
    return
  }
  converting.value = true
  try {
    const res = await composeApi.convert(cmd)
    createForm.value.yaml = res.yaml
    if (!createForm.value.name.trim()) {
      createForm.value.name = res.service
    }
    ElMessage.success('已转换为 compose YAML,可继续编辑')
  } catch (e: any) {
    ElMessage.error(e.message || '转换失败')
  } finally {
    converting.value = false
  }
}

async function doCreate() {
  const f = createForm.value
  if (!f.name.trim()) {
    ElMessage.warning('请输入项目名')
    return
  }
  if (!/^[a-zA-Z0-9._-]+$/.test(f.name.trim())) {
    ElMessage.warning('项目名只能包含字母、数字、- _ .')
    return
  }
  if (!f.yaml.trim()) {
    ElMessage.warning('请填写 compose YAML(或先转换 docker run 命令)')
    return
  }
  creating.value = true
  try {
    const res = await projectApi.create({
      name: f.name.trim(),
      yaml: f.yaml,
      description: f.description.trim(),
      start: f.start,
    })
    ElMessage.success(res.started ? `项目已创建并启动(${res.compose_file})` : '项目已保存')
    createVisible.value = false
    load()
  } catch (e: any) {
    ElMessage.error(e.message || '创建失败')
  } finally {
    creating.value = false
  }
}

/** 批量执行操作 */
async function batchAction(action: string) {
  const targets = selected.value
  if (!targets.length) return
  let ok = 0
  const failed: string[] = []
  for (const p of targets) {
    try {
      if (action === 'start') await projectApi.up(p.name)
      else if (action === 'stop') await projectApi.stop(p.name)
      else await projectApi.restart(p.name)
      ok++
    } catch {
      failed.push(p.name)
    }
  }
  const actionText: Record<string, string> = { start: '启动', stop: '停止', restart: '重启' }
  if (failed.length) {
    ElMessage.warning(`${actionText[action]}:成功 ${ok} 个,失败 ${failed.length} 个(${failed.join(', ')})`)
  } else {
    ElMessage.success(`${actionText[action]}成功:${ok} 个项目`)
  }
  loadAfter()
}

/** 批量重建:弹窗逐个展示日志 */
async function batchRebuild() {
  const targets = selected.value
  if (!targets.length) return
  rebuildVisible.value = true
  rebuilding.value = true
  rebuildOutput.value = ''
  let ok = 0
  for (const p of targets) {
    rebuildTitle.value = `重建日志 - ${p.name}`
    rebuildOutput.value += `===== 重建 ${p.name} =====\n`
    try {
      const res = await projectApi.rebuild(p.name)
      rebuildOutput.value += (res.output || '(无输出)') + '\n\n'
      ok++
    } catch (e: any) {
      rebuildOutput.value += `失败: ${e.message}\n\n`
    }
  }
  rebuilding.value = false
  ElMessage.success(`重建完成:${ok}/${targets.length} 个项目`)
  loadAfter()
}

/** 批量删除 */
async function batchRemove() {
  const targets = selected.value
  if (!targets.length) return
  const managed = targets.some((p) => p.source === 'managed')
  const names = targets.map((p) => p.name).join(', ')
  const tip = `确定删除选中的 <b>${targets.length}</b> 个项目吗?<br/><span class="mono">${names}</span><br/>将执行 compose down 删除容器与网络,${
    managed ? '<br/>本工具创建的项目会同时删除其 compose 文件' : ''
  }<br/><span style="color:#e6a23c">数据卷保留;已有项目的宿主机 compose 文件保留</span>`
  try {
    await ElMessageBox.confirm(tip, '批量删除确认', {
      type: 'warning',
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      confirmButtonClass: 'el-button--danger',
      dangerouslyUseHTMLString: true,
    })
  } catch {
    return
  }
  let ok = 0
  const failed: string[] = []
  for (const p of targets) {
    try {
      await projectApi.remove(p.name)
      ok++
    } catch {
      failed.push(p.name)
    }
  }
  if (failed.length) {
    ElMessage.warning(`删除:成功 ${ok} 个,失败 ${failed.length} 个(${failed.join(', ')})`)
  } else {
    ElMessage.success(`已删除 ${ok} 个项目`)
  }
  loadAfter()
}

function dotColor(row: Project): string {
  if (!row.has_containers) return '#909399'
  return row.running > 0 ? '#67c23a' : '#e6a23c'
}

onMounted(load)
</script>

<style scoped>
.mb-16 {
  margin-bottom: 16px;
}

.count-hint {
  color: var(--el-text-color-secondary);
  font-size: 13px;
}

/* 第二行操作栏 */
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
  color: var(--el-color-primary);
  cursor: pointer;
}

.svc-list {
  color: var(--el-text-color-regular);
  font-size: 0.9em;
}

.running {
  color: #67c23a;
  font-weight: 600;
}

.dim {
  color: var(--el-text-color-secondary);
}

.mono textarea {
  font-family: Consolas, 'Courier New', monospace;
}

/* docker run 转换区 */
.convert-box {
  width: 100%;
}

.convert-tip {
  font-size: 12px;
  color: var(--el-text-color-placeholder);
  margin-top: 4px;
}

.rebuild-log {
  background: var(--el-fill-color-light);
  border-radius: 8px;
  padding: 14px;
  font-size: 12.5px;
  line-height: 1.6;
  max-height: 400px;
  overflow: auto;
  color: var(--el-text-color-primary);
  margin: 0;
  white-space: pre-wrap;
  word-break: break-all;
}
</style>
