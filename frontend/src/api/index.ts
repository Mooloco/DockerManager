import { get, post, del, request } from './client'
import type {
  SystemInfo,
  ContainerItem,
  ContainerOverview,
  ContainerStats,
  ImageItem,
  NetworkItem,
  VolumeItem,
  Project,
  ProjectDetail,
  ComposeResult,
} from './types'

// ---- 认证 ----
export const authApi = {
  login: (username: string, password: string) =>
    post<{ username: string }>('/auth/login', { username, password }),
  logout: () => post<{ logged_out: boolean }>('/auth/logout'),
  me: () => get<{ username: string }>('/auth/me'),
  changePassword: (oldPassword: string, newPassword: string) =>
    post<{ changed: boolean }>('/auth/change-password', {
      old_password: oldPassword,
      new_password: newPassword,
    }),
}

// ---- System ----
export const systemApi = {
  info: () => get<SystemInfo>('/system/info'),
  ping: () => get<{ ok: boolean }>('/system/ping'),
}

// ---- Containers ----
export const containerApi = {
  list: (all = true) => get<ContainerItem[]>(`/containers?all=${all}`),
  action: (id: string, action: string, opts: Record<string, unknown> = {}) =>
    post<{ action: string; state: string }>(`/containers/${id}/${action}`, opts),
  inspect: (id: string) => get<Record<string, unknown>>(`/containers/${id}/inspect`),
  overview: (id: string) => get<ContainerOverview>(`/containers/${id}/overview`),
  stats: (id: string) => get<ContainerStats>(`/containers/${id}/stats`),
}

// ---- Images ----
export const imageApi = {
  list: () => get<ImageItem[]>('/images'),
  remove: (id: string, force = false) => del<{ removed: boolean }>(`/images/${id}?force=${force}`),
}

// ---- Networks ----
export const networkApi = {
  list: () => get<NetworkItem[]>('/networks'),
  inspect: (id: string) => get<Record<string, unknown>>(`/networks/${id}`),
  remove: (id: string) => del<{ removed: boolean }>(`/networks/${id}`),
}

// ---- Volumes ----
export const volumeApi = {
  list: () => get<VolumeItem[]>('/volumes'),
  inspect: (name: string) => get<Record<string, unknown>>(`/volumes/${encodeURIComponent(name)}`),
  remove: (name: string, force = false) =>
    del<{ removed: boolean }>(`/volumes/${encodeURIComponent(name)}?force=${force}`),
}

// ---- Projects (Compose) ----
export const projectApi = {
  list: () => get<Project[]>('/projects'),
  get: (name: string) => get<ProjectDetail>(`/projects/${encodeURIComponent(name)}`),
  getYaml: (name: string) =>
    get<{ name: string; yaml: string; compose_file: string }>(`/projects/${encodeURIComponent(name)}/yaml`),
  create: (data: { name: string; yaml: string; description?: string; start?: boolean }) =>
    post<{ name: string; compose_file: string; started: boolean; output?: string }>('/projects', data),
  update: (name: string, yaml: string) =>
    request<{ updated: boolean }>(`/projects/${encodeURIComponent(name)}`, {
      method: 'PUT',
      body: JSON.stringify({ yaml }),
    }),
  up: (name: string) => post<ComposeResult>(`/projects/${encodeURIComponent(name)}/up`),
  stop: (name: string) => post<ComposeResult>(`/projects/${encodeURIComponent(name)}/stop`),
  down: (name: string, removeVolumes = false) =>
    post<ComposeResult>(`/projects/${encodeURIComponent(name)}/down`, { remove_volumes: removeVolumes }),
  restart: (name: string) => post<ComposeResult>(`/projects/${encodeURIComponent(name)}/restart`),
  rebuild: (name: string) => post<ComposeResult>(`/projects/${encodeURIComponent(name)}/rebuild`),
  remove: (name: string) =>
    del<{ removed: boolean; message: string; file_gone: boolean }>(`/projects/${encodeURIComponent(name)}`),
}

// ---- Compose 工具 ----
export const composeApi = {
  convert: (command: string) =>
    post<{ yaml: string; service: string }>('/compose/convert', { command }),
}
