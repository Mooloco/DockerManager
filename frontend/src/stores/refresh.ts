import { defineStore } from 'pinia'

const KEY = 'dm-refresh-interval'

/** 页面自动刷新频率(秒) */
export const useRefreshStore = defineStore('refresh', {
  state: () => ({
    interval: Number(localStorage.getItem(KEY)) || 5,
  }),
  actions: {
    setInterval(sec: number) {
      this.interval = sec
      localStorage.setItem(KEY, String(sec))
    },
  },
})

/** 可选刷新频率 */
export const REFRESH_OPTIONS = [
  { value: 1, label: '1 秒' },
  { value: 2, label: '2 秒' },
  { value: 3, label: '3 秒' },
  { value: 5, label: '5 秒' },
  { value: 7, label: '7 秒' },
]
