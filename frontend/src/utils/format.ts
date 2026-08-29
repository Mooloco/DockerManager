/** 格式化工具 */

/** 时间戳/时间字符串 → 本地时间字符串 */
export function formatTime(v: string | number | undefined | null): string {
  if (v === undefined || v === null || v === '' || v === '0001-01-01T00:00:00Z') return '-'
  const d = typeof v === 'number' ? new Date(v * 1000) : new Date(v)
  if (isNaN(d.getTime())) return String(v)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

/** 相对时间(多久之前) */
export function formatRelative(v: number | string): string {
  const t = typeof v === 'number' ? v * 1000 : new Date(v).getTime()
  if (isNaN(t) || t <= 0) return '-'
  const diff = Date.now() - t
  const min = 60_000
  const hour = 60 * min
  const day = 24 * hour
  if (diff < min) return '刚刚'
  if (diff < hour) return `${Math.floor(diff / min)} 分钟前`
  if (diff < day) return `${Math.floor(diff / hour)} 小时前`
  if (diff < 30 * day) return `${Math.floor(diff / day)} 天前`
  return formatTime(v)
}

/** 字节数 → 可读大小 */
export function formatBytes(bytes: number | undefined | null): string {
  if (bytes === undefined || bytes === null || bytes < 0) return '-'
  if (bytes === 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB', 'PB']
  const i = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1)
  const val = bytes / Math.pow(1024, i)
  return `${val >= 100 ? val.toFixed(0) : val.toFixed(1)} ${units[i]}`
}

/** 端口列表 → 字符串(80/tcp 或 0.0.0.0:8080->80/tcp) */
export function formatPorts(ports: { ip?: string; private_port: number; public_port?: number; type: string }[] | null | undefined): string {
  if (!ports || ports.length === 0) return '-'
  return ports
    .map((p) => {
      if (p.public_port) {
        const ip = p.ip && p.ip !== '0.0.0.0' && p.ip !== '::' ? `${p.ip}:` : ''
        return `${ip}${p.public_port}->${p.private_port}/${p.type}`
      }
      return `${p.private_port}/${p.type}`
    })
    .join(', ')
}

/** 容器状态 → 状态徽章类型 */
export function statusType(state: string): 'success' | 'warning' | 'danger' | 'info' {
  switch (state) {
    case 'running':
      return 'success'
    case 'paused':
    case 'restarting':
      return 'warning'
    case 'dead':
      return 'danger'
    default:
      return 'info'
  }
}

/** 状态文本(英文原始值 → 中文) */
export function statusText(state: string): string {
  const map: Record<string, string> = {
    running: '运行中',
    exited: '已退出',
    paused: '已暂停',
    restarting: '重启中',
    created: '已创建',
    dead: '异常',
    removed: '已删除',
  }
  return map[state] || state
}
