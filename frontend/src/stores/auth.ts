import { defineStore } from 'pinia'
import { authApi } from '../api'

interface AuthState {
  username: string
  loading: boolean
}

export const useAuthStore = defineStore('auth', {
  state: (): AuthState => ({
    username: '',
    loading: false,
  }),
  getters: {
    isLoggedIn: (s) => !!s.username,
  },
  actions: {
    async login(username: string, password: string) {
      this.loading = true
      try {
        const res = await authApi.login(username, password)
        this.username = res.username
      } finally {
        this.loading = false
      }
    },
    async logout() {
      try {
        await authApi.logout()
      } catch {
        // 忽略登出失败
      }
      this.username = ''
    },
    /** 恢复会话(页面刷新后调用) */
    async restore() {
      try {
        const res = await authApi.me()
        this.username = res.username
        return true
      } catch {
        this.username = ''
        return false
      }
    },
  },
})
