<template>
  <div id="app">
    <router-view />
  </div>
</template>

<script>
import { generaMenu } from '@/assets/js/menu'
export default {
  created() {
    if (this.$store.state.userInfo != null) {
      generaMenu()
    }
    this.axios.post('/api/report')

    // 初始化主题和动画设置
    this.initTheme()
    this.initAnimation()
  },
  methods: {
    initTheme() {
      const root = document.documentElement

      // 加载主题设置
      const savedTheme = localStorage.getItem('aurora-theme')
      if (savedTheme) {
        root.setAttribute('data-theme', savedTheme === 'dark' ? 'dark' : savedTheme)
      }

      // 加载暗色模式设置
      const savedFollowSystem = localStorage.getItem('aurora-theme-follow-system')
      const followSystem = savedFollowSystem === 'false' ? false : true

      if (followSystem) {
        const darkMode = window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches
        localStorage.setItem('aurora-dark-mode', darkMode)
        if (darkMode) {
          root.setAttribute('data-theme', 'dark')
        }
      } else {
        const savedDarkMode = localStorage.getItem('aurora-dark-mode')
        if (savedDarkMode === 'true') {
          root.setAttribute('data-theme', 'dark')
        }
      }
    },
    initAnimation() {
      const root = document.documentElement

      // 加载动画设置
      const savedAnimationEnabled = localStorage.getItem('aurora-animation-enabled')
      if (savedAnimationEnabled === 'false') {
        root.style.setProperty('--transition-base', 'none')
        root.style.setProperty('--transition-fast', 'none')
        root.style.setProperty('--transition-slow', 'none')
      } else {
        const savedAnimationSpeed = localStorage.getItem('aurora-animation-speed')
        const speed = savedAnimationSpeed ? parseFloat(savedAnimationSpeed) : 1
        root.style.setProperty('--transition-base', `all ${0.3 * speed}s cubic-bezier(0.4, 0, 0.2, 1)`)
        root.style.setProperty('--transition-fast', `all ${0.2 * speed}s ease`)
        root.style.setProperty('--transition-slow', `all ${0.5 * speed}s cubic-bezier(0.4, 0, 0.2, 1)`)
      }
    }
  }
}
</script>
