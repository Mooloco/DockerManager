/** 与后端 API 对应的类型定义 */

// ---- System ----
export interface SystemInfo {
  server_version: string
  api_version: string
  min_api_version: string
  os_type: string
  operating_system: string
  architecture: string
  kernel_version: string
  driver: string
  root_dir: string
  memory_limit: boolean
  swap_limit: boolean
  cpus: number
  total_memory: number
  name: string
  containers: number
  containers_running: number
  containers_paused: number
  containers_stopped: number
  images: number
}

// ---- Containers ----
export interface Port {
  ip: string
  private_port: number
  public_port: number
  type: string
}

export interface ContainerItem {
  id: string
  short_id: string
  name: string
  image: string
  image_id: string
  state: string
  status: string
  created: number
  ports: Port[]
  cpu_percent: number
  memory_bytes: number
  memory_percent: number
  pids: number
  net_io: [number, number]
  block_io: [number, number]
}

export interface ContainerStats {
  cpu_percent: number
  memory_bytes: number
  memory_percent: number
  memory_limit: number
  pids: number
  net_io: [number, number]
  block_io: [number, number]
}

export interface ContainerMount {
  type: string
  source?: string
  name?: string
  destination: string
  mode: string
  rw: boolean
}

export interface ContainerOverview {
  id: string
  name: string
  image: string
  created: string
  started_at: string
  finished_at: string
  status: string
  state: Record<string, unknown>
  restart_policy: string
  command: string[] | null
  entrypoint: string[] | null
  working_dir: string
  tty: boolean
  ports: Record<string, unknown>
  mounts: ContainerMount[]
  env: string[]
  labels: Record<string, string>
  network_mode: string
  restart_count: number
}

// ---- Images ----
export interface ImageItem {
  id: string
  short_id: string
  repo_tags: string[]
  repo_digests: string[]
  created: number
  size: number
  shared_size: number
  virtual_size: number
  containers: number
}

// ---- Networks ----
export interface NetworkItem {
  id: string
  short_id: string
  name: string
  driver: string
  scope: string
  attachable: boolean
  internal: boolean
  ipv6: boolean
  labels: Record<string, string>
}

// ---- Volumes ----
export interface VolumeItem {
  name: string
  driver: string
  mountpoint: string
  created_at: string
  labels: Record<string, string>
  scope: string
  usage_data?: { size: number; ref_count: number }
}

// ---- Projects (Compose) ----
export interface Project {
  name: string
  source: 'discovered' | 'managed'
  config_files: string[]
  working_dir: string
  containers: number
  running: number
  services: string[]
  has_containers: boolean
  description?: string
  compose_file?: string
}

export interface ProjectContainer {
  id: string
  name: string
  image: string
  state: string
  status: string
  ports: Port[]
  ip?: string
}

export interface NetworkMember {
  name: string
  ip: string
}

export interface ProjectNetwork {
  name: string
  driver: string
  internal: boolean
  containers: NetworkMember[]
}

export interface ProjectService {
  name: string
  containers: ProjectContainer[]
}

export interface ProjectVolume {
  type: 'volume' | 'bind' | 'tmpfs' | string
  name: string
  destination: string
  mountpoint?: string
  service: string
  rw: boolean
}

export interface ProjectDetail {
  name: string
  source: 'discovered' | 'managed'
  config_files: string[]
  working_dir: string
  containers: number
  running: number
  services: ProjectService[]
  volumes: ProjectVolume[]
  networks: ProjectNetwork[]
  has_containers: boolean
  description?: string
  compose_file?: string
}

export interface ComposeResult {
  output: string
  exit_code: number
}

// ---- WS ----
export interface WSMessage {
  type: 'log' | 'stats' | 'pull' | 'error' | 'end'
  stream?: 'stdout' | 'stderr'
  data?: unknown
  message?: string
}
