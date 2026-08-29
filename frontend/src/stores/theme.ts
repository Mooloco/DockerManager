import { defineStore } from 'pinia'

const THEME_KEY = 'dm-theme'

/** 主题:dark / light */
export const useThemeStore = defineStore('theme', {
  state: () => ({
    dark: localStorage.getItem(THEME_KEY) !== 'light',
  }),
  actions: {
    toggle() {
      this.dark = !this.dark
      localStorage.setItem(THEME_KEY, this.dark ? 'dark' : 'light')
      this.apply()
    },
    apply() {
      document.documentElement.classList.toggle('dark', this.dark)
    },
  },
})
