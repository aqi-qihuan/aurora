<template>
  <div class="theme-settings">
    <el-drawer
      :visible.sync="drawerVisible"
      :size="isMobile ? '85%' : '360px'"
      :with-header="false"
      :modal-append-to-body="true"
      custom-class="theme-drawer"
      @close="handleClose"
    >
      <div class="theme-settings-panel">
        <!-- 头部 -->
        <div class="panel-header">
          <h3 class="panel-title">
            <i class="el-icon-brush" />
            主题设置
          </h3>
          <button class="close-btn" @click="drawerVisible = false">
            <i class="el-icon-close" />
          </button>
        </div>

        <!-- 内容区 -->
        <div class="panel-body">
          <!-- 主题模式 -->
          <section class="setting-section">
            <h4 class="section-title">
              <i class="el-icon-sunny" />
              主题模式
            </h4>
            <div class="theme-modes">
              <div
                v-for="mode in themeModes"
                :key="mode.value"
                class="mode-card"
                :class="{ active: currentTheme === mode.value }"
                @click="setThemeMode(mode.value)"
              >
                <div class="mode-preview" :class="mode.value">
                  <div class="preview-header" />
                  <div class="preview-sidebar" />
                  <div class="preview-content">
                    <div class="preview-card" />
                    <div class="preview-card" />
                  </div>
                </div>
                <span class="mode-label">{{ mode.label }}</span>
                <i v-if="currentTheme === mode.value" class="el-icon-check check-icon" />
              </div>
            </div>
          </section>

          <!-- 强调色 -->
          <section class="setting-section">
            <h4 class="section-title">
              <i class="el-icon-palette" />
              强调色
            </h4>
            <div class="color-palette">
              <div
                v-for="color in accentColors"
                :key="color.value"
                class="color-item"
                :class="{ active: currentAccent === color.value }"
                :style="{ backgroundColor: color.value }"
                @click="setAccentColor(color.value)"
              >
                <i v-if="currentAccent === color.value" class="el-icon-check" />
              </div>
            </div>
          </section>

          <!-- 动画设置 -->
          <section class="setting-section">
            <h4 class="section-title">
              <i class="el-icon-video-play" />
              动画效果
            </h4>
            <div class="setting-item">
              <span>启用动画</span>
              <el-switch
                v-model="animationEnabled"
                active-color="#409EFF"
                @change="toggleAnimation"
              />
            </div>
            <div v-if="animationEnabled" class="setting-item">
              <span>动画速度</span>
              <el-slider
                v-model="animationSpeed"
                :min="0.5"
                :max="2"
                :step="0.25"
                :format-tooltip="formatSpeedTooltip"
                @change="updateAnimationSpeed"
              />
            </div>
          </section>

          <!-- 布局设置 -->
          <section class="setting-section">
            <h4 class="section-title">
              <i class="el-icon-s-grid" />
              布局设置
            </h4>
            <div class="setting-item">
              <span>侧边栏折叠</span>
              <el-switch
                v-model="sidebarCollapsed"
                active-color="#409EFF"
                @change="toggleSidebar"
              />
            </div>
            <div class="setting-item">
              <span>标签页显示</span>
              <el-switch
                v-model="showTabs"
                active-color="#409EFF"
                @change="toggleTabs"
              />
            </div>
          </section>

          <!-- 重置按钮 -->
          <div class="reset-section">
            <el-button type="default" size="small" @click="resetSettings">
              <i class="el-icon-refresh-left" />
              恢复默认设置
            </el-button>
          </div>
        </div>
      </div>
    </el-drawer>
  </div>
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
      drawerVisible: this.visible,
      isMobile: false,
      currentTheme: 'light',
      currentAccent: '#2563EB',
      animationEnabled: true,
      animationSpeed: 1,
      sidebarCollapsed: false,
      showTabs: true,
      themeModes: [
        { label: '浅色', value: 'light' },
        { label: '深色', value: 'dark' },
        { label: '跟随系统', value: 'system' }
      ],
      accentColors: [
        { value: '#2563EB', label: '蓝色' },
        { value: '#10B981', label: '绿色' },
        { value: '#F59E0B', label: '橙色' },
        { value: '#EF4444', label: '红色' },
        { value: '#8B5CF6', label: '紫色' },
        { value: '#EC4899', label: '粉色' },
        { value: '#06B6D4', label: '青色' },
        { value: '#F97316', label: '琥珀' }
      ]
    }
  },
  mounted() {
    this.checkMobile()
    window.addEventListener('resize', this.checkMobile)
    this.loadSettings()
  },
  beforeDestroy() {
    window.removeEventListener('resize', this.checkMobile)
  },
  methods: {
    checkMobile() {
      this.isMobile = window.innerWidth < 768
    },
    loadSettings() {
      // 加载主题设置
      const savedTheme = localStorage.getItem('aurora-theme')
      const followSystem = localStorage.getItem('aurora-theme-follow-system') !== 'false'
      
      if (followSystem) {
        this.currentTheme = 'system'
      } else if (savedTheme === 'dark') {
        this.currentTheme = 'dark'
      } else {
        this.currentTheme = 'light'
      }

      // 加载强调色
      const savedAccent = localStorage.getItem('aurora-accent-color')
      if (savedAccent) {
        this.currentAccent = savedAccent
      }

      // 加载动画设置
      const savedAnimation = localStorage.getItem('aurora-animation-enabled')
      this.animationEnabled = savedAnimation !== 'false'

      const savedSpeed = localStorage.getItem('aurora-animation-speed')
      this.animationSpeed = savedSpeed ? parseFloat(savedSpeed) : 1

      // 加载布局设置
      const savedSidebar = localStorage.getItem('aurora-sidebar-collapsed')
      this.sidebarCollapsed = savedSidebar === 'true'

      const savedTabs = localStorage.getItem('aurora-show-tabs')
      this.showTabs = savedTabs !== 'false'
    },
    setThemeMode(mode) {
      this.currentTheme = mode
      const root = document.documentElement

      if (mode === 'system') {
        localStorage.setItem('aurora-theme-follow-system', 'true')
        const systemDark = window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches
        if (systemDark) {
          root.setAttribute('data-theme', 'dark')
          localStorage.setItem('aurora-dark-mode', 'true')
        } else {
          root.removeAttribute('data-theme')
          localStorage.setItem('aurora-dark-mode', 'false')
        }
      } else if (mode === 'dark') {
        localStorage.setItem('aurora-theme-follow-system', 'false')
        localStorage.setItem('aurora-dark-mode', 'true')
        localStorage.setItem('aurora-theme', 'dark')
        root.setAttribute('data-theme', 'dark')
      } else {
        localStorage.setItem('aurora-theme-follow-system', 'false')
        localStorage.setItem('aurora-dark-mode', 'false')
        localStorage.setItem('aurora-theme', 'light')
        root.removeAttribute('data-theme')
      }

      this.$message.success(`已切换到${this.themeModes.find(m => m.value === mode).label}模式`)
    },
    setAccentColor(color) {
      this.currentAccent = color
      localStorage.setItem('aurora-accent-color', color)
      
      // 更新 CSS 变量
      const root = document.documentElement
      root.style.setProperty('--color-primary', color)
      
      // 计算并设置相关颜色
      const lightColor = this.adjustColor(color, 20)
      const darkColor = this.adjustColor(color, -20)
      root.style.setProperty('--color-primary-light', lightColor)
      root.style.setProperty('--color-primary-dark', darkColor)
      
      this.$message.success('强调色已更新')
    },
    adjustColor(color, amount) {
      const num = parseInt(color.replace('#', ''), 16)
      const r = Math.min(255, Math.max(0, (num >> 16) + amount))
      const g = Math.min(255, Math.max(0, ((num >> 8) & 0x00FF) + amount))
      const b = Math.min(255, Math.max(0, (num & 0x00FF) + amount))
      return '#' + ((r << 16) | (g << 8) | b).toString(16).padStart(6, '0')
    },
    toggleAnimation(enabled) {
      localStorage.setItem('aurora-animation-enabled', enabled)
      const root = document.documentElement
      
      if (!enabled) {
        root.style.setProperty('--transition-base', 'none')
        root.style.setProperty('--transition-fast', 'none')
        root.style.setProperty('--transition-slow', 'none')
      } else {
        this.updateAnimationSpeed(this.animationSpeed)
      }
    },
    updateAnimationSpeed(speed) {
      localStorage.setItem('aurora-animation-speed', speed)
      const root = document.documentElement
      
      if (this.animationEnabled) {
        root.style.setProperty('--transition-base', `all ${0.3 * speed}s cubic-bezier(0.4, 0, 0.2, 1)`)
        root.style.setProperty('--transition-fast', `all ${0.2 * speed}s ease`)
        root.style.setProperty('--transition-slow', `all ${0.5 * speed}s cubic-bezier(0.4, 0, 0.2, 1)`)
      }
    },
    formatSpeedTooltip(val) {
      return val === 1 ? '正常' : val < 1 ? `慢 ${(1/val).toFixed(1)}x` : `快 ${val}x`
    },
    toggleSidebar(collapsed) {
      localStorage.setItem('aurora-sidebar-collapsed', collapsed)
      this.$store.commit('trigger')
    },
    toggleTabs(show) {
      localStorage.setItem('aurora-show-tabs', show)
      this.$emit('toggle-tabs', show)
    },
    resetSettings() {
      this.$confirm('确定要恢复所有默认设置吗？', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        // 清除所有主题设置
        localStorage.removeItem('aurora-theme')
        localStorage.removeItem('aurora-dark-mode')
        localStorage.removeItem('aurora-theme-follow-system')
        localStorage.removeItem('aurora-accent-color')
        localStorage.removeItem('aurora-animation-enabled')
        localStorage.removeItem('aurora-animation-speed')
        localStorage.removeItem('aurora-sidebar-collapsed')
        localStorage.removeItem('aurora-show-tabs')

        // 重置为默认值
        this.currentTheme = 'light'
        this.currentAccent = '#2563EB'
        this.animationEnabled = true
        this.animationSpeed = 1
        this.sidebarCollapsed = false
        this.showTabs = true

        // 应用默认设置
        const root = document.documentElement
        root.removeAttribute('data-theme')
        root.style.removeProperty('--color-primary')
        root.style.removeProperty('--color-primary-light')
        root.style.removeProperty('--color-primary-dark')
        this.updateAnimationSpeed(1)

        this.$message.success('已恢复默认设置')
        this.drawerVisible = false
        this.$emit('update:visible', false)
      }).catch(() => {})
    },
    open() {
      this.drawerVisible = true
      this.$emit('update:visible', true)
    },
    handleClose() {
      this.$emit('update:visible', false)
    }
  }
}
</script>

<style scoped>
.theme-settings-panel {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: var(--color-bg, #f5f7fa);
}

.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px 24px;
  background: var(--color-bg-card, #fff);
  border-bottom: 1px solid var(--color-border, #e4e7ed);
}

.panel-title {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: var(--color-text, #303133);
  display: flex;
  align-items: center;
  gap: 10px;
}

.panel-title i {
  color: var(--color-primary, #409eff);
}

.close-btn {
  width: 32px;
  height: 32px;
  border: none;
  background: transparent;
  border-radius: 8px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-text-secondary, #606266);
  transition: all 0.2s ease;
}

.close-btn:hover {
  background: var(--color-bg-hover, #f5f7fa);
  color: var(--color-text, #303133);
}

.panel-body {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
}

.setting-section {
  background: var(--color-bg-card, #fff);
  border-radius: 12px;
  padding: 20px;
  margin-bottom: 16px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
}

.section-title {
  margin: 0 0 16px 0;
  font-size: 14px;
  font-weight: 600;
  color: var(--color-text, #303133);
  display: flex;
  align-items: center;
  gap: 8px;
}

.section-title i {
  color: var(--color-primary, #409eff);
  font-size: 16px;
}

/* 主题模式卡片 */
.theme-modes {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
}

.mode-card {
  position: relative;
  cursor: pointer;
  border-radius: 10px;
  padding: 12px;
  border: 2px solid transparent;
  background: var(--color-bg, #f5f7fa);
  transition: all 0.2s ease;
  text-align: center;
}

.mode-card:hover {
  border-color: var(--color-primary-200, #d9ecff);
}

.mode-card.active {
  border-color: var(--color-primary, #409eff);
  background: var(--color-primary-50, #ecf5ff);
}

.mode-preview {
  width: 100%;
  height: 60px;
  border-radius: 6px;
  overflow: hidden;
  position: relative;
  margin-bottom: 8px;
  background: #f5f7fa;
}

.mode-preview.light {
  background: #fff;
}

.mode-preview.light .preview-header {
  height: 12px;
  background: #fff;
  border-bottom: 1px solid #e4e7ed;
}

.mode-preview.light .preview-sidebar {
  position: absolute;
  left: 0;
  top: 12px;
  bottom: 0;
  width: 20px;
  background: #fff;
  border-right: 1px solid #e4e7ed;
}

.mode-preview.light .preview-content {
  margin-left: 20px;
  padding: 6px;
}

.mode-preview.light .preview-card {
  height: 14px;
  background: #f5f7fa;
  border-radius: 2px;
  margin-bottom: 4px;
}

.mode-preview.dark {
  background: #1a1a1a;
}

.mode-preview.dark .preview-header {
  height: 12px;
  background: #1a1a1a;
  border-bottom: 1px solid #2b2b2c;
}

.mode-preview.dark .preview-sidebar {
  position: absolute;
  left: 0;
  top: 12px;
  bottom: 0;
  width: 20px;
  background: #1a1a1a;
  border-right: 1px solid #2b2b2c;
}

.mode-preview.dark .preview-content {
  margin-left: 20px;
  padding: 6px;
}

.mode-preview.dark .preview-card {
  height: 14px;
  background: #2b2b2c;
  border-radius: 2px;
  margin-bottom: 4px;
}

.mode-preview.system {
  background: linear-gradient(135deg, #fff 50%, #1a1a1a 50%);
}

.mode-preview.system .preview-header {
  height: 12px;
  background: transparent;
}

.mode-label {
  font-size: 12px;
  color: var(--color-text, #303133);
  font-weight: 500;
}

.check-icon {
  position: absolute;
  top: 6px;
  right: 6px;
  width: 18px;
  height: 18px;
  background: var(--color-primary, #409eff);
  color: #fff;
  border-radius: 50%;
  font-size: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* 颜色调色板 */
.color-palette {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
}

.color-item {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
  position: relative;
  margin: 0 auto;
}

.color-item:hover {
  transform: scale(1.1);
}

.color-item.active {
  box-shadow: 0 0 0 3px var(--color-bg-card, #fff), 0 0 0 5px var(--color-primary, #409eff);
}

.color-item i {
  color: #fff;
  font-size: 20px;
  font-weight: bold;
}

/* 设置项 */
.setting-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 0;
  border-bottom: 1px solid var(--color-border-light, #f5f7fa);
}

.setting-item:last-child {
  border-bottom: none;
}

.setting-item span {
  font-size: 14px;
  color: var(--color-text, #303133);
}

/* 重置按钮 */
.reset-section {
  padding: 20px;
  text-align: center;
}

.reset-section .el-button {
  width: 100%;
}

/* 滚动条样式 */
.panel-body::-webkit-scrollbar {
  width: 6px;
}

.panel-body::-webkit-scrollbar-thumb {
  border-radius: 3px;
  background-color: rgba(144, 147, 153, 0.2);
}

.panel-body::-webkit-scrollbar-thumb:hover {
  background-color: rgba(144, 147, 153, 0.4);
}

/* 移动端适配 */
@media (max-width: 768px) {
  .panel-header {
    padding: 16px 20px;
  }

  .panel-title {
    font-size: 16px;
  }

  .panel-body {
    padding: 16px;
  }

  .setting-section {
    padding: 16px;
    margin-bottom: 12px;
  }

  .theme-modes {
    gap: 8px;
  }

  .mode-card {
    padding: 8px;
  }

  .mode-preview {
    height: 50px;
  }

  .color-palette {
    gap: 8px;
  }

  .color-item {
    width: 40px;
    height: 40px;
  }
}
</style>
