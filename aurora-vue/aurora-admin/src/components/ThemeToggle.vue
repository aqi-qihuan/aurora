<template>
  <div class="theme-toggle-wrapper">
    <el-tooltip
      :content="isDark ? '切换到浅色模式' : '切换到深色模式'"
      placement="bottom"
      effect="light"
    >
      <button
        class="theme-toggle-btn"
        :class="{ 'is-dark': isDark }"
        @click="toggleTheme"
        aria-label="切换主题"
      >
        <transition name="icon-fade" mode="out-in">
          <i v-if="isDark" key="moon" class="el-icon-moon" style="font-size: 20px;"></i>
          <i v-else key="sunny" class="el-icon-sunny" style="font-size: 20px;"></i>
        </transition>
      </button>
    </el-tooltip>
  </div>
</template>

<script>
export default {
  name: 'ThemeToggle',
  data() {
    return {
      isDark: false
    }
  },
  methods: {
    checkTheme() {
      const root = document.documentElement
      const savedDarkMode = localStorage.getItem('aurora-dark-mode')
      const followSystem = localStorage.getItem('aurora-theme-follow-system') !== 'false'

      if (followSystem) {
        const systemDark = window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches
        this.isDark = systemDark
      } else {
        this.isDark = savedDarkMode === 'true'
      }

      if (this.isDark) {
        root.setAttribute('data-theme', 'dark')
      } else {
        root.removeAttribute('data-theme')
      }
    },
    toggleTheme() {
      const root = document.documentElement
      this.isDark = !this.isDark

      if (this.isDark) {
        root.setAttribute('data-theme', 'dark')
        localStorage.setItem('aurora-dark-mode', 'true')
      } else {
        root.removeAttribute('data-theme')
        localStorage.setItem('aurora-dark-mode', 'false')
      }

      localStorage.setItem('aurora-theme-follow-system', 'false')

      this.$emit('theme-change', this.isDark)
    },
    handleSystemThemeChange(e) {
      const followSystem = localStorage.getItem('aurora-theme-follow-system') !== 'false'
      if (followSystem) {
        this.isDark = e.matches
        const root = document.documentElement
        if (this.isDark) {
          root.setAttribute('data-theme', 'dark')
        } else {
          root.removeAttribute('data-theme')
        }
      }
    }
  },
  mounted() {
    this.checkTheme()
    if (window.matchMedia) {
      const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
      mediaQuery.addEventListener('change', this.handleSystemThemeChange)
    }
  },
  beforeDestroy() {
    if (window.matchMedia) {
      const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
      mediaQuery.removeEventListener('change', this.handleSystemThemeChange)
    }
  }
}
</script>

<style scoped>
.theme-toggle-wrapper {
  display: inline-flex;
  align-items: center;
}

.theme-toggle-btn {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  border: none;
  background: #f5f7fa;
  color: #606266;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  position: relative;
  overflow: hidden;
}

.theme-toggle-btn:hover {
  background: rgba(64, 158, 255, 0.1);
  color: #409eff;
  transform: scale(1.05);
}

.theme-toggle-btn.is-dark {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #fff;
}

.theme-toggle-btn.is-dark:hover {
  background: linear-gradient(135deg, #764ba2 0%, #667eea 100%);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
}

.theme-toggle-btn i {
  transition: transform 0.3s ease;
}

.theme-toggle-btn:hover i {
  transform: rotate(15deg);
}

.icon-fade-enter-active,
.icon-fade-leave-active {
  transition: all 0.2s ease;
}

.icon-fade-enter {
  opacity: 0;
  transform: scale(0.5) rotate(-45deg);
}

.icon-fade-leave-to {
  opacity: 0;
  transform: scale(0.5) rotate(45deg);
}

@media (max-width: 768px) {
  .theme-toggle-btn {
    width: 36px;
    height: 36px;
    font-size: 18px;
  }
}
</style>
