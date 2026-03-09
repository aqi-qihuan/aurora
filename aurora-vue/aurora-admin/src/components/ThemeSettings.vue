<template>
  <el-dialog
    title="主题设置"
    :visible.sync="visible"
    width="400px"
    :before-close="handleClose"
    custom-class="theme-settings-dialog"
  >
    <div class="theme-settings">
      <!-- 主题选择 -->
      <div class="settings-section">
        <div class="section-title">
          <i class="el-icon-picture-outline" />
          主题颜色
        </div>
        <div class="theme-grid">
          <div
            v-for="theme in themes"
            :key="theme.value"
            :class="['theme-item', { active: currentTheme === theme.value }]"
            @click="setTheme(theme.value)"
          >
            <div class="theme-preview" :style="theme.previewStyle" />
            <div class="theme-name">{{ theme.name }}</div>
            <i v-if="currentTheme === theme.value" class="el-icon-check theme-check" />
          </div>
        </div>
      </div>

      <!-- 暗色模式 -->
      <div class="settings-section">
        <div class="section-title">
          <i class="el-icon-moon" />
          暗色模式
        </div>
        <div class="setting-item">
          <span class="setting-label">跟随系统</span>
          <el-switch
            v-model="followSystem"
            @change="handleFollowSystemChange"
          />
        </div>
        <div v-if="!followSystem" class="setting-item">
          <span class="setting-label">手动切换</span>
          <el-switch
            v-model="darkMode"
            active-text="暗色"
            inactive-text="亮色"
            @change="handleDarkModeChange"
          />
        </div>
      </div>

      <!-- 动画设置 -->
      <div class="settings-section">
        <div class="section-title">
          <i class="el-icon-video-play" />
          动画设置
        </div>
        <div class="setting-item">
          <span class="setting-label">启用动画</span>
          <el-switch
            v-model="animationEnabled"
            @change="handleAnimationChange"
          />
        </div>
        <div v-if="animationEnabled" class="setting-item">
          <span class="setting-label">动画速度</span>
          <el-slider
            v-model="animationSpeed"
            :min="0.5"
            :max="2"
            :step="0.1"
            :marks="speedMarks"
            @change="handleAnimationSpeedChange"
          />
        </div>
      </div>
    </div>

    <div slot="footer" class="dialog-footer">
      <el-button @click="resetSettings">重置默认</el-button>
      <el-button type="primary" @click="handleClose">确定</el-button>
    </div>
  </el-dialog>
</template>

<script>
export default {
  name: 'ThemeSettings',
  props: {
    visible: {
      type: Boolean,
      default: false
    }
  },
  data() {
    return {
      themes: [
        { value: 'light', name: '默认白', previewStyle: 'background: linear-gradient(135deg, #f5f7fa 0%, #ffffff 100%)' },
        { value: 'dark', name: '暗夜黑', previewStyle: 'background: linear-gradient(135deg, #0a0a0a 0%, #141414 100%)' },
        { value: 'blue', name: '海洋蓝', previewStyle: 'background: linear-gradient(135deg, #3b82f6 0%, #60a5fa 100%)' },
        { value: 'purple', name: '紫罗兰', previewStyle: 'background: linear-gradient(135deg, #8b5cf6 0%, #a78bfa 100%)' },
        { value: 'green', name: '翠绿青', previewStyle: 'background: linear-gradient(135deg, #10b981 0%, #34d399 100%)' }
      ],
      currentTheme: 'light',
      darkMode: false,
      followSystem: true,
      animationEnabled: true,
      animationSpeed: 1,
      speedMarks: {
        0.5: '慢',
        1: '正常',
        2: '快'
      }
    }
  },
  created() {
    this.loadSettings()
    this.setupSystemThemeListener()
  },
  methods: {
    loadSettings() {
      // 加载主题设置
      const savedTheme = localStorage.getItem('aurora-theme')
      if (savedTheme) {
        this.currentTheme = savedTheme
        this.applyTheme(savedTheme)
      }

      // 加载暗色模式设置
      const savedFollowSystem = localStorage.getItem('aurora-theme-follow-system')
      this.followSystem = savedFollowSystem === 'false' ? false : true

      const savedDarkMode = localStorage.getItem('aurora-dark-mode')
      if (savedDarkMode !== null) {
        this.darkMode = savedDarkMode === 'true'
      }

      // 如果跟随系统,立即应用系统主题
      if (this.followSystem) {
        this.applySystemTheme()
      }

      // 加载动画设置
      const savedAnimationEnabled = localStorage.getItem('aurora-animation-enabled')
      if (savedAnimationEnabled !== null) {
        this.animationEnabled = savedAnimationEnabled === 'true'
        this.applyAnimation()
      }

      const savedAnimationSpeed = localStorage.getItem('aurora-animation-speed')
      if (savedAnimationSpeed) {
        this.animationSpeed = parseFloat(savedAnimationSpeed)
        this.applyAnimationSpeed()
      }
    },
    setupSystemThemeListener() {
      // 监听系统主题变化
      if (window.matchMedia) {
        this.mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
        this.mediaQuery.addEventListener('change', this.handleSystemThemeChange)
      }
    },
    handleSystemThemeChange(e) {
      if (this.followSystem) {
        this.darkMode = e.matches
        this.applyTheme(this.currentTheme)
      }
    },
    applySystemTheme() {
      if (window.matchMedia) {
        this.darkMode = window.matchMedia('(prefers-color-scheme: dark)').matches
        this.applyTheme(this.currentTheme)
      }
    },
    setTheme(theme) {
      this.currentTheme = theme
      this.applyTheme(theme)
      localStorage.setItem('aurora-theme', theme)
    },
    applyTheme(theme) {
      const root = document.documentElement

      // 移除所有主题属性
      root.removeAttribute('data-theme')

      // 应用当前主题
      if (theme === 'dark') {
        root.setAttribute('data-theme', 'dark')
      } else if (theme !== 'light') {
        root.setAttribute('data-theme', theme)
      }

      // 如果暗色模式开启,叠加暗色主题
      if (this.darkMode && theme !== 'dark') {
        root.setAttribute('data-theme', 'dark')
      }
    },
    handleFollowSystemChange(value) {
      this.followSystem = value
      localStorage.setItem('aurora-theme-follow-system', value)

      if (value) {
        this.applySystemTheme()
      }
    },
    handleDarkModeChange(value) {
      this.darkMode = value
      localStorage.setItem('aurora-dark-mode', value)
      this.applyTheme(this.currentTheme)
    },
    handleAnimationChange(value) {
      this.animationEnabled = value
      localStorage.setItem('aurora-animation-enabled', value)
      this.applyAnimation()
    },
    applyAnimation() {
      const root = document.documentElement

      if (this.animationEnabled) {
        root.style.setProperty('--transition-base', `all ${0.3 * this.animationSpeed}s cubic-bezier(0.4, 0, 0.2, 1)`)
        root.style.setProperty('--transition-fast', `all ${0.2 * this.animationSpeed}s ease`)
        root.style.setProperty('--transition-slow', `all ${0.5 * this.animationSpeed}s cubic-bezier(0.4, 0, 0.2, 1)`)
      } else {
        root.style.setProperty('--transition-base', 'none')
        root.style.setProperty('--transition-fast', 'none')
        root.style.setProperty('--transition-slow', 'none')
      }
    },
    handleAnimationSpeedChange(value) {
      this.animationSpeed = value
      localStorage.setItem('aurora-animation-speed', value)
      this.applyAnimationSpeed()
    },
    applyAnimationSpeed() {
      if (this.animationEnabled) {
        this.applyAnimation()
      }
    },
    resetSettings() {
      this.currentTheme = 'light'
      this.darkMode = false
      this.followSystem = true
      this.animationEnabled = true
      this.animationSpeed = 1

      // 清除存储
      localStorage.removeItem('aurora-theme')
      localStorage.removeItem('aurora-dark-mode')
      localStorage.removeItem('aurora-theme-follow-system')
      localStorage.removeItem('aurora-animation-enabled')
      localStorage.removeItem('aurora-animation-speed')

      // 应用默认设置
      this.applyTheme('light')
      this.applyAnimation()

      this.$message.success('已重置为默认设置')
    },
    handleClose() {
      this.$emit('update:visible', false)
    }
  },
  beforeDestroy() {
    // 移除系统主题监听
    if (this.mediaQuery) {
      this.mediaQuery.removeEventListener('change', this.handleSystemThemeChange)
    }
  }
}
</script>

<style scoped>
.theme-settings {
  padding: 10px;
}

.settings-section {
  margin-bottom: 25px;
}

.settings-section:last-child {
  margin-bottom: 0;
}

.section-title {
  font-size: 14px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 15px;
  display: flex;
  align-items: center;
}

.section-title i {
  margin-right: 8px;
  font-size: 16px;
  color: #409eff;
}

.theme-grid {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 12px;
}

.theme-item {
  position: relative;
  cursor: pointer;
  border-radius: 8px;
  overflow: hidden;
  border: 2px solid transparent;
  transition: all 0.3s ease;
}

.theme-item:hover {
  transform: translateY(-3px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.theme-item.active {
  border-color: #409eff;
  box-shadow: 0 0 0 3px rgba(64, 158, 255, 0.1);
}

.theme-preview {
  height: 50px;
  width: 100%;
  border-radius: 6px;
}

.theme-name {
  text-align: center;
  font-size: 12px;
  color: #606266;
  margin-top: 6px;
}

.theme-check {
  position: absolute;
  top: 5px;
  right: 5px;
  font-size: 14px;
  color: #fff;
  background: #409eff;
  border-radius: 50%;
  width: 18px;
  height: 18px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.setting-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 0;
  border-bottom: 1px solid #f0f0f0;
}

.setting-item:last-child {
  border-bottom: none;
}

.setting-label {
  font-size: 14px;
  color: #606266;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}
</style>

<style>
.theme-settings-dialog .el-dialog__header {
  border-bottom: 1px solid #f0f0f0;
}

.theme-settings-dialog .el-dialog__body {
  padding: 20px;
}

.theme-settings-dialog .el-slider__marks-text {
  font-size: 12px;
}
</style>
