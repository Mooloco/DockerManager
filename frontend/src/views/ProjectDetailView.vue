<template>
  <div class="page-container" v-loading="!loaded && !loadError">
    <!-- 头部 -->
    <div class="detail-head">
      <el-button text :icon="ArrowLeft" @click="router.push('/projects')">返回</el-button>
      <span class="name">{{ detail?.name || '项目详情' }}</span>
      <el-tag v-if="detail" :type="detail.source === 'managed' ? 'primary' : 'info'" size="small" effect="plain">
        {{ detail.source === 'managed' ? '本工具创建' : '已有项目' }}
      </el-tag>
      <el-tag v-if="detail" :type="detail.has_containers && detail.running > 0 ? 'success' : 'warning'" size="small">
        {{ detail.has_containers ? (detail.running > 0 ? `运行中 ${detail.running}/${detail.containers}` : '全部停止') : '已停止' }}
      </el-tag>
      <div class="spacer" />
      <el-button-group>
        <el-button size="small" type="success" :disabled="!canUp" @click="action('up')">
          <el-icon><VideoPlay /></el-icon>&nbsp;启动
        </el-button>
        <el-button size="small" :disabled="!canDown" @click="action('stop')">
          <el-icon><VideoPause /></el-icon>&nbsp;停止
        </el-button>
        <el-button size="small" type="warning" :disabled="!canDown" @click="action('restart')">
          <el-icon><RefreshRight /></el-icon>&nbsp;重启
        </el-button>
        <el-button size="small" type="warning" plain @click="doRebuild">
          <el-icon><Refresh /></el-icon>&nbsp;重建
        </el-button>
        <el-button size="small" @click="openYamlEdit">
          <el-icon><EditPen /></el-icon>&nbsp;编辑 compose
        </el-button>
      </el-button-group>
    </div>

    <el-alert v-if="loadError" :title="loadError" type="error" show-icon :closable="false" class="mb-16" />

    <template v-if="detail">
      <!-- 基本信息 -->
      <el-card shadow="never" class="mb-16">
        <template #header>
          <div class="card-header"><el-icon><InfoFilled /></el-icon><span>基本信息</span></div>
        </template>
        <el-descriptions :column="2" border>
          <el-descriptions-item label="项目名">{{ detail.name }}</el-descriptions-item>
          <el-descriptions-item label="工作目录">{{ detail.working_dir || '-' }}</el-descriptions-item>
          <el-descriptions-item label="compose 文件" :span="2">
            <div v-for="f in detail.config_files" :key="f" class="mono file-line">{{ f }}</div>
            <span v-if="!detail.config_files?.length">-</span>
          </el-descriptions-item>
        </el-descriptions>
      </el-card>

      <!-- 服务列表 -->
      <el-card shadow="never" class="mb-16">
        <template #header>
          <div class="card-header"><el-icon><Box /></el-icon><span>服务</span></div>
        </template>
        <div v-if="detail.services?.length">
          <div v-for="svc in detail.services" :key="svc.name" class="svc-block">
            <div class="svc-name">{{ svc.name }}</div>
            <div class="svc-containers">
              <div v-for="c in svc.containers" :key="c.id" class="svc-container">
                <span class="state-dot" :style="{ background: dotColor(c.state) }" />
                <span class="cname clickable" @click="router.push(`/containers/${c.id}`)">{{ c.name }}</span>
                <span class="cimg mono dim">{{ c.image }}</span>
                <el-tag :type="c.state === 'running' ? 'success' : 'info'" size="small">{{ c.state }}</el-tag>
                <el-button
                  size="small"
                  :type="c.state === 'running' ? 'default' : 'success'"
                  @click="toggleContainer(c)"
                >
                  {{ c.state === 'running' ? '停止' : '启动' }}
                </el-button>
              </div>
            </div>
          </div>
        </div>
        <el-empty v-else description="该项目当前没有容器(可能已停止)" :image-size="60" />
      </el-card>

      <!-- 卷信息 -->
      <el-card shadow="never" class="mb-16">
        <template #header>
          <div class="card-header"><el-icon><FolderOpened /></el-icon><span>卷 / 挂载</span></div>
        </template>
        <el-table :data="detail.volumes" size="small" stripe v-if="detail.volumes?.length">
          <el-table-column label="类型" width="90">
            <template #default="{ row }">
              <el-tag :type="row.type === 'bind' ? 'warning' : 'primary'" size="small">
                {{ row.type === 'bind' ? 'bind 挂载' : row.type }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="名称 / 源路径" min-width="200" show-overflow-tooltip>
            <template #default="{ row }"><span class="mono">{{ row.name }}</span></template>
          </el-table-column>
          <el-table-column label="容器内路径" min-width="160">
            <template #default="{ row }"><span class="mono dim">{{ row.destination }}</span></template>
          </el-table-column>
          <el-table-column label="宿主机位置" min-width="200" show-overflow-tooltip>
            <template #default="{ row }">
              <span v-if="row.mountpoint" class="mono dim">{{ row.mountpoint }}</span>
              <span v-else class="dim">-</span>
            </template>
          </el-table-column>
          <el-table-column label="服务" width="120">
            <template #default="{ row }">{{ row.service }}</template>
          </el-table-column>
          <el-table-column label="读写" width="70">
            <template #default="{ row }">
              <el-tag :type="row.rw ? 'success' : 'info'" size="small">{{ row.rw ? 'rw' : 'ro' }}</el-tag>
            </template>
          </el-table-column>
        </el-table>
        <el-empty v-else description="该项目没有卷或挂载" :image-size="60" />
      </el-card>

      <!-- 网络 / 端口 -->
      <el-card shadow="never" class="mb-16">
        <template #header>
          <div class="card-header"><el-icon><Share /></el-icon><span>网络 / 端口</span></div>
        </template>

        <!-- 网络 -->
        <div v-if="detail.networks?.length" class="net-list">
          <div v-for="net in detail.networks" :key="net.name" class="net-item">
            <span class="mono net-name">{{ net.name }}</span>
            <el-tag v-if="net.internal" size="small" type="warning" class="net-tag">internal</el-tag>
            <el-tag size="small" type="info" effect="plain" class="net-tag">{{ net.driver || '-' }}</el-tag>
            <span class="net-members mono dim">
              {{ net.containers.map((c) => (c.ip ? `${c.name}(${c.ip})` : c.name)).join('  ') }}
            </span>
          </div>
        </div>
        <el-empty v-if="!detail.networks?.length" description="无网络信息" :image-size="50" />

        <!-- 端口映射:宿主机 → 容器,可点击 -->
        <template v-if="portMappings.length">
          <el-divider>端口映射</el-divider>
          <div class="port-list">
            <div v-for="(m, i) in portMappings" :key="i" class="port-map">
              <span class="port-container">{{ m.container }}</span>
              <a :href="m.hostUrl" target="_blank" rel="noopener" class="port-link">{{ m.hostDisplay }}</a>
              <el-icon class="port-arrow" :size="14"><Right /></el-icon>
              <a
                v-if="m.containerUrl"
                :href="m.containerUrl"
                target="_blank"
                rel="noopener"
                class="port-link"
              >{{ m.containerDisplay }}</a>
              <span v-else class="port-link dim">{{ m.containerDisplay }}</span>
            </div>
          </div>
        </template>
      </el-card>

      <!-- YAML 展示 -->
      <el-card shadow="never">
        <template #header>
          <div class="card-header">
            <el-icon><Document /></el-icon>
            <span>compose YAML</span>
            <el-button size="small" class="yaml-edit-btn" @click="openYamlEdit">
              <el-icon><EditPen /></el-icon>&nbsp;编辑
            </el-button>
          </div>
        </template>
        <pre class="yaml-view">{{ yamlContent || '(加载中...)' }}</pre>
      </el-card>
    </template>

    <!-- 编辑 YAML 对话框 -->
    <el-dialog v-model="yamlDialog" :title="`编辑 compose - ${detail?.name || ''}`" width="720px" :close-on-click-modal="false">
      <div class="yaml-file-info mono" v-if="yamlFile">{{ yamlFile }}</div>
      <el-input v-model="yamlEditContent" type="textarea" :rows="18" class="mono yaml-editor" />
      <template #footer>
        <el-button @click="yamlDialog = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveYaml">保存</el-button>
      </template>
    </el-dialog>

    <!-- 重建日志弹窗 -->
    <el-dialog v-model="rebuildVisible" title="重建日志" width="680px" :close-on-click-modal="false">
      <pre class="rebuild-log" v-loading="rebuilding">{{ rebuildOutput || '(执行中...)' }}</pre>
      <template #footer>
        <el-button type="primary" @click="rebuildVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  ArrowLeft,
  VideoPlay,
  VideoPause,
  RefreshRight,
  Refresh,
  EditPen,
  InfoFilled,
  Box,
  FolderOpened,
  Document,
  Share,
  Right,
} from '@element-plus/icons-vue'
import { projectApi, containerApi } from '../api'
import type { ProjectDetail, ProjectContainer } from '../api/types'

const route = useRoute()
const router = useRouter()
const projectName = route.params.name as string

const detail = ref<ProjectDetail | null>(null)
const yamlContent = ref('')
const yamlFile = ref('')
const yamlDialog = ref(false)
const yamlEditContent = ref('')
const saving = ref(false)
const loaded = ref(false)
const loadError = ref('')

const canUp = computed(() => {
  if (!detail.value) return false
  return detail.value.running === 0
})
const canDown = computed(() => !!detail.value?.has_containers && detail.value.running > 0)

/** 聚合所有服务的端口映射(宿主机 → 容器,链接化,IPv4/IPv6 双绑定去重) */
const portMappings = computed(() => {
  const hostname = window.location.hostname
  const out: {
    container: string
    hostDisplay: string
    hostUrl: string
    containerDisplay: string
    containerUrl?: string
  }[] = []
  const seen = new Set<string>()
  for (const svc of detail.value?.services || []) {
    for (const c of svc.containers) {
      for (const p of c.ports || []) {
        if (!p.public_port) continue
        // 同一容器同一容器端口的 IPv4(0.0.0.0)与 IPv6(::)双绑定只保留一条
        const key = `${c.name}|${p.private_port}`
        if (seen.has(key)) continue
        seen.add(key)
        // 宿主机地址:绑定 0.0.0.0/:: 时用当前访问的管理页面地址
        const hostIp = p.ip && p.ip !== '0.0.0.0' && p.ip !== '::' ? p.ip : hostname
        const containerIp = c.ip || ''
        out.push({
          container: c.name,
          hostDisplay: `${hostIp}:${p.public_port}`,
          hostUrl: `http://${hostIp}:${p.public_port}/`,
          containerDisplay: containerIp ? `${containerIp}:${p.private_port}` : `${p.private_port}`,
          containerUrl: containerIp ? `http://${containerIp}:${p.private_port}/` : undefined,
        })
      }
    }
  }
  return out
})

async function loadDetail() {
  try {
    detail.value = await projectApi.get(projectName)
    loadError.value = ''
  } catch (e: any) {
    loadError.value = e.message || '加载项目详情失败'
  }
}

async function loadYaml() {
  try {
    const res = await projectApi.getYaml(projectName)
    yamlContent.value = res.yaml
    yamlFile.value = res.compose_file
  } catch {
    yamlContent.value = '# 无法读取 compose 文件内容\n# 容器部署时需挂载宿主目录,或该项目由外部管理'
  }
}

onMounted(async () => {
  await Promise.all([loadDetail(), loadYaml()])
  loaded.value = true
})

async function action(a: string) {
  try {
    let res
    if (a === 'up') res = await projectApi.up(projectName)
    else if (a === 'stop') res = await projectApi.stop(projectName)
    else res = await projectApi.restart(projectName)
    if (res.output) console.log('compose 输出:', res.output)
    ElMessage.success(`项目已${a === 'up' ? '启动' : a === 'stop' ? '停止(容器保留)' : '重启'}`)
    await Promise.all([loadDetail(), loadYaml()])
  } catch (e: any) {
    ElMessage.error(e.message || '操作失败')
    loadDetail()
  }
}

/** 重建:弹窗展示 down + up 完整日志 */
const rebuildVisible = ref(false)
const rebuilding = ref(false)
const rebuildOutput = ref('')

async function doRebuild() {
  rebuildVisible.value = true
  rebuilding.value = true
  rebuildOutput.value = ''
  try {
    const res = await projectApi.rebuild(projectName)
    rebuildOutput.value = res.output || '(无输出)'
    ElMessage.success('项目重建完成')
    await Promise.all([loadDetail(), loadYaml()])
  } catch (e: any) {
    rebuildOutput.value = `执行失败: ${e.message}`
    ElMessage.error(e.message || '重建失败')
  } finally {
    rebuilding.value = false
  }
}

async function toggleContainer(c: ProjectContainer) {
  try {
    const target = c.state === 'running' ? 'stop' : 'start'
    await containerApi.action(c.id, target)
    ElMessage.success(`${c.name} ${target === 'stop' ? '已停止' : '已启动'}`)
    loadDetail()
  } catch (e: any) {
    ElMessage.error(e.message || '操作失败')
  }
}

function openYamlEdit() {
  yamlEditContent.value = yamlContent.value
  yamlDialog.value = true
}

async function saveYaml() {
  if (!yamlEditContent.value.trim()) {
    ElMessage.warning('YAML 不能为空')
    return
  }
  saving.value = true
  try {
    await projectApi.update(projectName, yamlEditContent.value)
    ElMessage.success('已保存,执行"启动"使改动生效')
    yamlContent.value = yamlEditContent.value
    yamlDialog.value = false
  } catch (e: any) {
    ElMessage.error(e.message || '保存失败')
  } finally {
    saving.value = false
  }
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
</script>

<style scoped>
.mb-16 {
  margin-bottom: 16px;
}

.detail-head {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}

.detail-head .name {
  font-size: 17px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.spacer {
  flex: 1;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
}

.file-line {
  font-size: 0.85em;
  color: var(--el-text-color-regular);
}

.svc-block {
  margin-bottom: 16px;
}

.svc-block:last-child {
  margin-bottom: 0;
}

.svc-name {
  font-weight: 600;
  font-size: 14px;
  margin-bottom: 8px;
  color: var(--el-text-color-primary);
}

.svc-containers {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.svc-container {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 12px;
  background: var(--el-fill-color-lighter);
  border-radius: 6px;
  flex-wrap: wrap;
}

.state-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.cname {
  font-weight: 500;
}

.clickable {
  cursor: pointer;
  color: var(--el-color-primary);
}

.cimg {
  font-size: 0.8em;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.yaml-view {
  background: var(--el-fill-color-light);
  border-radius: 8px;
  padding: 14px;
  font-size: 12.5px;
  line-height: 1.6;
  max-height: 420px;
  overflow: auto;
  color: var(--el-text-color-primary);
  margin: 0;
}

.yaml-edit-btn {
  margin-left: auto;
}

.yaml-file-info {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-bottom: 8px;
}

.yaml-editor :deep(textarea) {
  font-family: Consolas, 'Courier New', monospace;
  font-size: 12.5px;
}

/* 网络 */
.net-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.net-item {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.net-name {
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.net-members {
  font-size: 12px;
  margin-left: 4px;
}

.net-tag {
  margin-right: 2px;
}

/* 端口映射 */
.port-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.port-map {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 12px;
  background: var(--el-fill-color-lighter);
  border-radius: 6px;
  flex-wrap: wrap;
}

.port-container {
  font-size: 13px;
  color: var(--el-text-color-regular);
  min-width: 110px;
}

.port-link {
  font-family: Consolas, 'Courier New', monospace;
  font-size: 13px;
  color: var(--el-color-primary);
  text-decoration: none;
}

.port-link:hover {
  text-decoration: underline;
}

.port-link.dim {
  color: var(--el-text-color-secondary);
  cursor: default;
}

.port-arrow {
  color: var(--el-text-color-secondary);
}

/* 重建日志 */
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
