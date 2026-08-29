<template>
  <div class="page-container" v-loading="!loaded && !loadError">
    <!-- 头部信息 -->
    <div class="detail-head">
      <el-button text :icon="ArrowLeft" @click="router.push('/containers')">返回</el-button>
      <span class="name">{{ overview?.name || '容器详情' }}</span>
      <el-tag v-if="overview" :type="statusType(overview.status)" size="small">
        {{ statusText(overview.status) }}
      </el-tag>
      <div class="spacer" />
      <el-button-group>
        <el-tooltip content="启动"><el-button size="small" type="success" :icon="VideoPlay" :disabled="!canStart" @click="action('start')" /></el-tooltip>
        <el-tooltip content="停止"><el-button size="small" :icon="VideoPause" :disabled="!canStop" @click="action('stop')" /></el-tooltip>
        <el-tooltip content="重启"><el-button size="small" type="warning" :icon="RefreshRight" :disabled="!canRestart" @click="action('restart')" /></el-tooltip>
        <el-tooltip content="删除"><el-button size="small" type="danger" plain :icon="Delete" @click="remove" /></el-tooltip>
      </el-button-group>
    </div>

    <el-alert v-if="loadError" :title="loadError" type="error" show-icon :closable="false" class="mb-16" />

    <el-tabs v-model="tab" v-if="loaded">
      <!-- Overview -->
      <el-tab-pane label="Overview" name="overview">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="容器 ID">
            <span class="mono">{{ overview?.id }}</span>
          </el-descriptions-item>
          <el-descriptions-item label="名称">{{ overview?.name }}</el-descriptions-item>
          <el-descriptions-item label="镜像">{{ overview?.image }}</el-descriptions-item>
          <el-descriptions-item label="状态">{{ overview?.status }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ formatTime(overview?.created) }}</el-descriptions-item>
          <el-descriptions-item label="启动时间">{{ formatTime(overview?.started_at) }}</el-descriptions-item>
          <el-descriptions-item label="停止时间">{{ formatTime(overview?.finished_at) }}</el-descriptions-item>
          <el-descriptions-item label="重启策略">{{ overview?.restart_policy || '-' }}</el-descriptions-item>
          <el-descriptions-item label="重启次数">{{ overview?.restart_count ?? 0 }}</el-descriptions-item>
          <el-descriptions-item label="网络模式">{{ overview?.network_mode || '-' }}</el-descriptions-item>
          <el-descriptions-item label="工作目录">{{ overview?.working_dir || '-' }}</el-descriptions-item>
          <el-descriptions-item label="TTY">{{ overview?.tty ? '是' : '否' }}</el-descriptions-item>
          <el-descriptions-item label="Command" :span="2">
            <code class="mono cmd">{{ fmtCmd(overview?.command) }}</code>
          </el-descriptions-item>
          <el-descriptions-item label="Entrypoint" :span="2">
            <code class="mono cmd">{{ fmtCmd(overview?.entrypoint) }}</code>
          </el-descriptions-item>
          <el-descriptions-item label="端口映射" :span="2">
            <span class="mono">{{ fmtPortMap(overview?.ports) }}</span>
          </el-descriptions-item>
          <el-descriptions-item label="挂载" :span="2">
            <div v-if="overview?.mounts?.length" class="mount-list">
              <div v-for="(m, i) in overview.mounts" :key="i" class="mono mount-item">
                {{ m.type }}: {{ m.source || m.name }} → {{ m.destination }}
              </div>
            </div>
            <span v-else>-</span>
          </el-descriptions-item>
          <el-descriptions-item label="环境变量" :span="2">
            <el-collapse v-if="overview?.env?.length">
              <el-collapse-item title="查看全部环境变量" :name="'env'">
                <div v-for="(e, i) in overview.env" :key="i" class="mono env-item">
                  {{ maskEnv(e) }}
                </div>
              </el-collapse-item>
            </el-collapse>
            <span v-else>-</span>
          </el-descriptions-item>
        </el-descriptions>
      </el-tab-pane>

      <!-- 日志 -->
      <el-tab-pane label="日志" name="logs">
        <div class="log-toolbar">
          <el-checkbox v-model="logFollow" @change="connectLogs">实时跟随</el-checkbox>
          <el-checkbox v-model="logAutoScroll" :disabled="!logFollow">自动滚动</el-checkbox>
          <el-checkbox v-model="logStdout" @change="connectLogs">stdout</el-checkbox>
          <el-checkbox v-model="logStderr" @change="connectLogs">stderr</el-checkbox>
          <el-select v-model="logTail" style="width: 120px" @change="connectLogs">
            <el-option v-for="n in [100, 500, 1000, 5000]" :key="n" :label="`最近 ${n} 行`" :value="String(n)" />
          </el-select>
          <div class="spacer" />
          <el-tag v-if="logConnected" type="success" size="small" effect="plain">已连接</el-tag>
          <el-tag v-else type="warning" size="small" effect="plain">连接中...</el-tag>
          <el-button size="small" @click="clearLogs"><el-icon><Delete /></el-icon>&nbsp;清屏</el-button>
        </div>
        <div ref="logBox" class="log-box" :class="{ dark: true }">
          <div v-if="!logs.length" class="log-empty">暂无日志</div>
          <div v-for="(line, i) in logs" :key="i" class="log-line">
            <span class="log-stream" :class="line.stream">{{ line.stream === 'stderr' ? 'ERR' : 'OUT' }}</span>
            <span class="log-text">{{ line.data }}</span>
          </div>
        </div>
      </el-tab-pane>

      <!-- 实时监控 -->
      <el-tab-pane label="实时监控" name="stats">
        <div class="stats-grid">
          <div class="stat-panel">
            <div class="stat-title">CPU 使用率</div>
            <div class="stat-big">{{ lastStats?.cpu_percent?.toFixed(1) ?? '-' }}%</div>
            <Sparkline :points="cpuHistory" :max="100" unit="%" color="#409eff" />
          </div>
          <div class="stat-panel">
            <div class="stat-title">内存</div>
            <div class="stat-big">
              {{ formatBytes(lastStats?.memory_bytes) }}
              <span class="stat-sub">{{ lastStats?.memory_percent?.toFixed(1) ?? '-' }}%</span>
            </div>
            <Sparkline :points="memHistory" :max="memMax" unit="MB" color="#67c23a" />
          </div>
        </div>
        <el-descriptions :column="4" border class="mt-16">
          <el-descriptions-item label="PIDs">{{ lastStats?.pids ?? '-' }}</el-descriptions-item>
          <el-descriptions-item label="内存限制">{{ formatBytes(lastStats?.memory_limit) }}</el-descriptions-item>
          <el-descriptions-item label="网络 RX">{{ formatBytes(lastStats?.net_io?.[0]) }}</el-descriptions-item>
          <el-descriptions-item label="网络 TX">{{ formatBytes(lastStats?.net_io?.[1]) }}</el-descriptions-item>
          <el-descriptions-item label="磁盘读">{{ formatBytes(lastStats?.block_io?.[0]) }}</el-descriptions-item>
          <el-descriptions-item label="磁盘写">{{ formatBytes(lastStats?.block_io?.[1]) }}</el-descriptions-item>
          <el-descriptions-item label="推送间隔">2 秒</el-descriptions-item>
          <el-descriptions-item label="数据点">{{ cpuHistory.length }} 个</el-descriptions-item>
        </el-descriptions>
      </el-tab-pane>

      <!-- Inspect -->
      <el-tab-pane label="Inspect" name="inspect">
        <div class="inspect-head">
          <el-segmented v-model="inspectMode" :options="['分组视图', 'Raw JSON']" />
        </div>
        <!-- 分组视图 -->
        <div v-if="inspectMode === '分组视图'" class="inspect-groups">
          <el-descriptions v-for="g in inspectGroups" :key="g.title" :title="g.title" :column="2" border class="group-desc">
            <el-descriptions-item v-for="(v, k) in g.items" :key="k" :label="String(k)">
              <span class="mono small">{{ fmtInspectValue(v) }}</span>
            </el-descriptions-item>
          </el-descriptions>
        </div>
        <!-- Raw JSON -->
        <pre v-else class="raw-json">{{ inspectJson }}</pre>
      </el-tab-pane>

      <!-- 终端(占位) -->
      <el-tab-pane label="终端" name="terminal">
        <div class="terminal-placeholder">
          <el-icon :size="48"><Monitor /></el-icon>
          <div>终端功能未开放,后续版本提供</div>
          <el-button type="primary" plain size="small" @click="ElMessage.info('终端功能未开放,后续版本提供')">
            知道了
          </el-button>
        </div>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  ArrowLeft,
  VideoPlay,
  VideoPause,
  RefreshRight,
  Delete,
  Monitor,
} from '@element-plus/icons-vue'
import { containerApi } from '../api'
import type { ContainerOverview, ContainerStats } from '../api/types'
import { formatBytes, formatTime, statusText, statusType } from '../utils/format'
import { WSClient } from '../websocket/ws'
import { useRefreshStore } from '../stores/refresh'
import Sparkline from '../components/Sparkline.vue'

const route = useRoute()
const router = useRouter()
const containerId = route.params.id as string
const refreshStore = useRefreshStore()

const tab = ref('overview')
const loaded = ref(false)
const loadError = ref('')
const overview = ref<ContainerOverview | null>(null)
const inspect = ref<Record<string, any> | null>(null)
const inspectMode = ref('分组视图')

// ---- 操作可用性 ----
const canStart = computed(() => ['exited', 'created', 'dead'].includes(overview.value?.status || ''))
const canStop = computed(() => overview.value?.status === 'running')
const canRestart = computed(() => ['running', 'exited'].includes(overview.value?.status || ''))

// ---- 日志 ----
const logs = ref<{ stream: string; data: string }[]>([])
const logFollow = ref(true)
const logAutoScroll = ref(true)
const logStdout = ref(true)
const logStderr = ref(true)
const logTail = ref('500')
const logConnected = ref(false)
const logBox = ref<HTMLElement>()
let logWS: WSClient | null = null

// ---- Stats ----
const lastStats = ref<ContainerStats | null>(null)
const cpuHistory = ref<number[]>([])
const memHistory = ref<number[]>([])
const memMax = computed(() => {
  const limit = lastStats.value?.memory_limit || 0
  return limit > 0 ? Math.max(limit / 1024 / 1024, 1) : 100
})
let statsWS: WSClient | null = null

// ---- 初始化 ----
async function loadOverview() {
  try {
    overview.value = await containerApi.overview(containerId)
    loadError.value = ''
  } catch (e: any) {
    loadError.value = e.message || '加载容器详情失败'
  }
}

async function loadInspect() {
  try {
    inspect.value = await containerApi.inspect(containerId)
  } catch {
    // inspect 失败不阻塞页面
  }
}

onMounted(async () => {
  await Promise.all([loadOverview(), loadInspect()])
  loaded.value = true
  connectLogs()
  connectStats()
})

onUnmounted(() => {
  logWS?.close()
  statsWS?.close()
})

// ---- 容器操作 ----
async function action(a: string) {
  try {
    const res = await containerApi.action(containerId, a)
    ElMessage.success(`操作成功,当前状态:${statusText(res.state)}`)
    await loadOverview()
  } catch (e: any) {
    ElMessage.error(e.message || '操作失败')
  }
}

async function remove() {
  try {
    await ElMessageBox.confirm(`确定删除容器 <b>${overview.value?.name}</b> 吗?`, '删除确认', {
      type: 'warning',
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      confirmButtonClass: 'el-button--danger',
      dangerouslyUseHTMLString: true,
    })
  } catch {
    return
  }
  try {
    await containerApi.action(containerId, 'remove', { force: true })
    ElMessage.success('容器已删除')
    router.push('/containers')
  } catch (e: any) {
    ElMessage.error(e.message || '删除失败')
  }
}

// ---- 日志连接 ----
function connectLogs() {
  logWS?.close()
  logWS = null
  if (!logFollow) {
    logConnected.value = false
    return
  }
  if (!logStdout.value && !logStderr.value) {
    logConnected.value = false
    return
  }
  const q = new URLSearchParams({
    follow: 'true',
    tail: logTail.value,
    stdout: String(logStdout.value),
    stderr: String(logStderr.value),
  })
  logWS = new WSClient(`/api/v1/ws/containers/${containerId}/logs?${q}`, {
    onOpen: () => {
      logConnected.value = true
    },
    onMessage: (msg) => {
      if (msg.type === 'log' && msg.stream) {
        logs.value.push({ stream: msg.stream, data: String(msg.data ?? '') })
        if (logs.value.length > 3000) logs.value.splice(0, logs.value.length - 3000)
        if (logAutoScroll.value) scrollLogsToBottom()
      } else if (msg.type === 'end') {
        // 日志流结束(如容器停止):主动关闭,防止自动重连反复拉取
        logConnected.value = false
        logWS?.close()
        logWS = null
      } else if (msg.type === 'error') {
        ElMessage.warning(`日志流错误: ${msg.message}`)
        logConnected.value = false
        logWS?.close()
        logWS = null
      }
    },
    onClose: () => {
      logConnected.value = false
    },
  })
  logWS.connect()
}

function scrollLogsToBottom() {
  requestAnimationFrame(() => {
    if (logBox.value) logBox.value.scrollTop = logBox.value.scrollHeight
  })
}

function clearLogs() {
  logs.value = []
}

// ---- Stats 连接 ----
function connectStats() {
  statsWS?.close()
  // 推送间隔跟随顶栏设置的刷新频率
  statsWS = new WSClient(`/api/v1/ws/containers/${containerId}/stats?interval=${refreshStore.interval}`, {
    onMessage: (msg) => {
      if (msg.type === 'stats') {
        const st = msg.data as ContainerStats
        lastStats.value = st
        cpuHistory.value.push(st.cpu_percent)
        memHistory.value.push(st.memory_bytes / 1024 / 1024)
        if (cpuHistory.value.length > 60) cpuHistory.value.shift()
        if (memHistory.value.length > 60) memHistory.value.shift()
      } else if (msg.type === 'error') {
        ElMessage.warning(`Stats 错误: ${msg.message}`)
      }
    },
  })
  statsWS.connect()
}

// 刷新频率变化时重建 stats 连接
watch(() => refreshStore.interval, connectStats)

// ---- 格式化 ----
function fmtCmd(cmd: string[] | null | undefined): string {
  if (!cmd?.length) return '-'
  return cmd.join(' ')
}

function fmtPortMap(ports: any): string {
  if (!ports || !Object.keys(ports).length) return '-'
  const parts: string[] = []
  for (const [privatePort, bindings] of Object.entries(ports)) {
    const arr = bindings as any[]
    if (!arr?.length) {
      parts.push(`${privatePort}`)
    } else {
      for (const b of arr) {
        parts.push(`${b.HostIp || '0.0.0.0'}:${b.HostPort}->${privatePort}`)
      }
    }
  }
  return parts.join(', ')
}

function maskEnv(env: string): string {
  const [k, v] = env.split('=')
  if (v && /(PASSWORD|PASS|SECRET|TOKEN|KEY)/i.test(k)) return `${k}=******`
  return env
}

function fmtInspectValue(v: any): string {
  if (v === null || v === undefined) return '-'
  if (typeof v === 'object') return JSON.stringify(v)
  return String(v)
}

// Inspect 分组
const inspectGroups = computed(() => {
  const d = inspect.value
  if (!d) return []
  const groups: { title: string; items: Record<string, any> }[] = []

  groups.push({ title: 'Basic', items: { ID: d.Id, Name: d.Name, Image: d.Image, Created: d.Created, Driver: d.Driver } })
  groups.push({
    title: 'Config',
    items: {
      Hostname: d.Config?.Hostname,
      User: d.Config?.User,
      ExposedPorts: d.Config?.ExposedPorts,
      Env: d.Config?.Env,
      Cmd: d.Config?.Cmd,
      Entrypoint: d.Config?.Entrypoint,
      WorkingDir: d.Config?.WorkingDir,
      Tty: d.Config?.Tty,
      OpenStdin: d.Config?.OpenStdin,
    },
  })
  groups.push({
    title: 'State',
    items: {
      Status: d.State?.Status,
      Running: d.State?.Running,
      Paused: d.State?.Paused,
      Restarting: d.State?.Restarting,
      ExitCode: d.State?.ExitCode,
      StartedAt: d.State?.StartedAt,
      FinishedAt: d.State?.FinishedAt,
      Health: d.State?.Health?.Status,
    },
  })
  groups.push({
    title: 'Network',
    items: {
      NetworkMode: d.HostConfig?.NetworkMode,
      Networks: d.NetworkSettings?.Networks,
      Ports: d.NetworkSettings?.Ports,
      IPAddress: d.NetworkSettings?.Networks ? undefined : d.NetworkSettings?.IPAddress,
    },
  })
  groups.push({ title: 'Mounts', items: d.Mounts?.map((m: any) => `${m.Type}: ${m.Source || m.Name} → ${m.Destination}`) })
  groups.push({
    title: 'Host Config',
    items: {
      RestartPolicy: d.HostConfig?.RestartPolicy?.Name,
      AutoRemove: d.HostConfig?.AutoRemove,
      Privileged: d.HostConfig?.Privileged,
      CapAdd: d.HostConfig?.CapAdd,
      CapDrop: d.HostConfig?.CapDrop,
      SecurityOpt: d.HostConfig?.SecurityOpt,
    },
  })
  groups.push({
    title: 'Resources',
    items: {
      CpuShares: d.HostConfig?.CpuShares,
      CpuPeriod: d.HostConfig?.CpuPeriod,
      CpuQuota: d.HostConfig?.CpuQuota,
      CpusetCpus: d.HostConfig?.CpusetCpus,
      Memory: d.HostConfig?.Memory,
      MemorySwap: d.HostConfig?.MemorySwap,
      NanoCpus: d.HostConfig?.NanoCpus,
    },
  })
  return groups
})

const inspectJson = computed(() => JSON.stringify(inspect.value, null, 2))

// 暗色日志框
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
}

.detail-head .name {
  font-size: 17px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.spacer {
  flex: 1;
}

.cmd {
  white-space: pre-wrap;
  word-break: break-all;
}

.mount-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.mount-item,
.env-item {
  font-size: 0.85em;
  color: var(--el-text-color-regular);
}

/* 日志 */
.log-toolbar {
  display: flex;
  align-items: center;
  gap: 14px;
  margin-bottom: 10px;
  flex-wrap: wrap;
}

.log-box {
  height: 520px;
  overflow-y: auto;
  background: #1e1e1e;
  border-radius: 8px;
  padding: 12px;
  font-family: Consolas, 'Courier New', monospace;
  font-size: 12.5px;
  line-height: 1.6;
}

.log-empty {
  color: #6a737d;
  text-align: center;
  padding-top: 40px;
}

.log-line {
  display: flex;
  gap: 8px;
  white-space: pre-wrap;
  word-break: break-all;
}

.log-stream {
  flex-shrink: 0;
  font-size: 10px;
  padding: 0 4px;
  border-radius: 3px;
  align-self: flex-start;
  margin-top: 3px;
}

.log-stream.stdout {
  background: #1f6feb33;
  color: #58a6ff;
}

.log-stream.stderr {
  background: #f8514933;
  color: #ff7b72;
}

.log-text {
  color: #c9d1d9;
}

/* Stats */
.stats-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
  margin-bottom: 16px;
}

.stat-panel {
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 8px;
  padding: 16px;
}

.stat-title {
  color: var(--el-text-color-secondary);
  font-size: 13px;
  margin-bottom: 8px;
}

.stat-big {
  font-size: 28px;
  font-weight: 700;
  color: var(--el-text-color-primary);
  margin-bottom: 12px;
}

.stat-sub {
  font-size: 14px;
  font-weight: 400;
  color: var(--el-text-color-secondary);
}

.mt-16 {
  margin-top: 16px;
}

/* Inspect */
.inspect-head {
  margin-bottom: 16px;
}

.inspect-groups {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.group-desc :deep(.el-descriptions__title) {
  font-size: 14px;
  font-weight: 600;
}

.small {
  font-size: 0.85em;
  word-break: break-all;
}

.raw-json {
  background: var(--el-fill-color-light);
  border-radius: 8px;
  padding: 16px;
  font-size: 12px;
  max-height: 600px;
  overflow: auto;
  color: var(--el-text-color-primary);
  margin: 0;
}

@media (max-width: 768px) {
  .stats-grid {
    grid-template-columns: 1fr;
  }
}
</style>
