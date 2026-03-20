<template>
  <div class="theme-settings">
    <el-drawer
      v-model="drawerVisible"
      :size="isMobile ? '85%' : '360px'"
      :with-header="false"
      :modal="true"
      :append-to-body="true"
      custom-class="theme-drawer"
      @close="handleClose"
    >
      <div class="theme-settings-panel">
        <div class="panel-header">
          <h3 class="panel-title">
            <el-icon :size="18"><Brush /></el-icon>
            主题设置
          </h3>
          <button class="close-btn" @click="drawerVisible = false">
            <el-icon :size="18"><Close /></el-icon>
          </button>
        </div>

        <div class="panel-body">
          <section class="setting-section">
            <h4 class="section-title">
              <el-icon><Sunny /></el-icon>
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
                <el-icon v-if="currentTheme === mode.value" class="check-icon"><Check /></el-icon>
              </div>
            </div>
          </section>

          <section class="setting-section">
            <h4 class="section-title">
              <el-icon><Opportunity /></el-icon>
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
                <el-icon v-if="currentAccent === color.value"><Check /></el-icon>
              </div>
            </div>
          </section>

          <section class="setting-section">
            <h4 class="section-title">
              <el-icon><VideoPlay /></el-icon>
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

          <section class="setting-section">
            <h4 class="section-title">
              <el-icon><Grid /></el-icon>
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

          <div class="reset-section">
            <el-button type="default" size="small" @click="resetSettings">
              <el-icon><RefreshLeft /></el-icon>
              恢复默认设置
            </el-button>
          </div>
        </div>
      </div>
    </el-drawer>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Brush, Close, Check, Sunny, Opportunity, VideoPlay,
  Grid, RefreshLeft
} from '@element-plus/icons-vue'

const props = defineProps({
  visible: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['update:visible', 'toggle-tabs'])

const drawerVisible = ref(props.visible)
const isMobile = ref(false)
const currentTheme = ref('light')
const currentAccent = ref('#2563EB')
const animationEnabled = ref(true)
const animationSpeed = ref(1)
const sidebarCollapsed = ref(false)
const showTabs = ref(true)

const themeModes = [
  { label: '浅白色', value: 'light', description: '清新明亮的浅白色调' },
  { label: '极客风', value: 'dark', description: '深色极客风格主题' },
  { label: '跟随系统', value: 'system', description: '自动跟随系统设置' }
]

const accentColors = [
  { value: '#2563EB', label: '蓝色' },
  { value: '#10B981', label: '绿色' },
  { value: '#F59E0B', label: '橙色' },
  { value: '#EF4444', label: '红色' },
  { value: '#8B5CF6', label: '紫色' },
  { value: '#EC4899', label: '粉色' },
  { value: '#06B6D4', label: '青色' },
  { value: '#F97316', label: '琥珀' }
]

const checkMobile = () => {
  isMobile.value = window.innerWidth < 768
}

const loadSettings = () => {
  const savedTheme = localStorage.getItem('aurora-theme')
  const followSystem = localStorage.getItem('aurora-theme-follow-system') !== 'false'
  const darkMode = localStorage.getItem('aurora-dark-mode')

  if (followSystem) {
    currentTheme.value = 'system'
  } else if (savedTheme === 'dark') {
    currentTheme.value = 'dark'
  } else {
    currentTheme.value = 'light'
  }

  // 应用主题到 DOM
  const root = document.documentElement
  if (currentTheme.value === 'system') {
    const systemDark = window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches
    if (systemDark) {
      root.setAttribute('data-theme', 'dark')
    } else {
      root.removeAttribute('data-theme')
    }
  } else if (currentTheme.value === 'dark') {
    root.setAttribute('data-theme', 'dark')
  } else {
    root.removeAttribute('data-theme')
  }

  const savedAccent = localStorage.getItem('aurora-accent-color')
  if (savedAccent) {
    currentAccent.value = savedAccent
    // 应用强调色
    root.style.setProperty('--color-primary', savedAccent)
    const lightColor = adjustColor(savedAccent, 20)
    const darkColor = adjustColor(savedAccent, -20)
    root.style.setProperty('--color-primary-light', lightColor)
    root.style.setProperty('--color-primary-dark', darkColor)
  }

  const savedAnimation = localStorage.getItem('aurora-animation-enabled')
  animationEnabled.value = savedAnimation !== 'false'

  const savedSpeed = localStorage.getItem('aurora-animation-speed')
  animationSpeed.value = savedSpeed ? parseFloat(savedSpeed) : 1

  // 应用动画速度
  if (animationEnabled.value) {
    updateAnimationSpeed(animationSpeed.value)
  } else {
    const root = document.documentElement
    root.style.setProperty('--transition-base', 'none')
    root.style.setProperty('--transition-fast', 'none')
    root.style.setProperty('--transition-slow', 'none')
  }

  const savedSidebar = localStorage.getItem('aurora-sidebar-collapsed')
  sidebarCollapsed.value = savedSidebar === 'true'

  const savedTabs = localStorage.getItem('aurora-show-tabs')
  showTabs.value = savedTabs !== 'false'
  emit('toggle-tabs', showTabs.value)
}

const setThemeMode = (mode) => {
  currentTheme.value = mode
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

  ElMessage.success(`已切换到${themeModes.find(m => m.value === mode).label}模式`)
}

const setAccentColor = (color) => {
  currentAccent.value = color
  localStorage.setItem('aurora-accent-color', color)

  const root = document.documentElement
  root.style.setProperty('--color-primary', color)

  const lightColor = adjustColor(color, 20)
  const darkColor = adjustColor(color, -20)
  root.style.setProperty('--color-primary-light', lightColor)
  root.style.setProperty('--color-primary-dark', darkColor)

  ElMessage.success('强调色已更新')
}

const adjustColor = (color, amount) => {
  const num = parseInt(color.replace('#', ''), 16)
  const r = Math.min(255, Math.max(0, (num >> 16) + amount))
  const g = Math.min(255, Math.max(0, ((num >> 8) & 0x00FF) + amount))
  const b = Math.min(255, Math.max(0, (num & 0x00FF) + amount))
  return '#' + ((r << 16) | (g << 8) | b).toString(16).padStart(6, '0')
}

const toggleAnimation = (enabled) => {
  localStorage.setItem('aurora-animation-enabled', enabled)
  const root = document.documentElement

  if (!enabled) {
    root.style.setProperty('--transition-base', 'none')
    root.style.setProperty('--transition-fast', 'none')
    root.style.setProperty('--transition-slow', 'none')
  } else {
    updateAnimationSpeed(animationSpeed.value)
  }
}

const updateAnimationSpeed = (speed) => {
  localStorage.setItem('aurora-animation-speed', speed)
  const root = document.documentElement

  if (animationEnabled.value) {
    root.style.setProperty('--transition-base', `all ${0.3 * speed}s cubic-bezier(0.4, 0, 0.2, 1)`)
    root.style.setProperty('--transition-fast', `all ${0.2 * speed}s ease`)
    root.style.setProperty('--transition-slow', `all ${0.5 * speed}s cubic-bezier(0.4, 0, 0.2, 1)`)
  }
}

const formatSpeedTooltip = (val) => {
  return val === 1 ? '正常' : val < 1 ? `慢 ${(1/val).toFixed(1)}x` : `快 ${val}x`
}

const toggleSidebar = (collapsed) => {
  localStorage.setItem('aurora-sidebar-collapsed', collapsed)
}

const toggleTabs = (show) => {
  localStorage.setItem('aurora-show-tabs', show)
  emit('toggle-tabs', show)
}

const resetSettings = () => {
  ElMessageBox.confirm('确定要恢复所有默认设置吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(() => {
    localStorage.removeItem('aurora-theme')
    localStorage.removeItem('aurora-dark-mode')
    localStorage.removeItem('aurora-theme-follow-system')
    localStorage.removeItem('aurora-accent-color')
    localStorage.removeItem('aurora-animation-enabled')
    localStorage.removeItem('aurora-animation-speed')
    localStorage.removeItem('aurora-sidebar-collapsed')
    localStorage.removeItem('aurora-show-tabs')

    currentTheme.value = 'light'
    currentAccent.value = '#2563EB'
    animationEnabled.value = true
    animationSpeed.value = 1
    sidebarCollapsed.value = false
    showTabs.value = true

    const root = document.documentElement
    root.removeAttribute('data-theme')
    root.style.removeProperty('--color-primary')
    root.style.removeProperty('--color-primary-light')
    root.style.removeProperty('--color-primary-dark')
    updateAnimationSpeed(1)

    ElMessage.success('已恢复默认设置')
    drawerVisible.value = false
    emit('update:visible', false)
  }).catch(() => {})
}

const open = () => {
  drawerVisible.value = true
  emit('update:visible', true)
}

const handleClose = () => {
  emit('update:visible', false)
}

defineExpose({
  open
})

onMounted(() => {
  checkMobile()
  window.addEventListener('resize', checkMobile)
  loadSettings()
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', checkMobile)
})
</script>

<style scoped>
.theme-settings-panel {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: var(--bg-base);
  color: var(--text-primary);
  transition: background-color 0.3s ease, color 0.3s ease;
}

.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px 24px;
  background: var(--bg-elevated);
  border-bottom: 1px solid var(--border-default);
  transition: background-color 0.3s ease, border-color 0.3s ease;
}

.panel-title {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
  display: flex;
  align-items: center;
  gap: 10px;
}

.panel-title .el-icon {
  color: var(--primary);
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
  color: var(--text-secondary);
  transition: all 0.2s ease;
}

.close-btn:hover {
  background: var(--bg-surface);
  color: var(--text-primary);
}

.panel-body {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
}

.setting-section {
  background: var(--bg-elevated);
  border-radius: 12px;
  padding: 20px;
  margin-bottom: 16px;
  border: 1px solid var(--border-light);
  box-shadow: var(--shadow-sm);
  transition: background-color 0.3s ease, border-color 0.3s ease;
}

.section-title {
  margin: 0 0 16px 0;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  display: flex;
  align-items: center;
  gap: 8px;
}

.section-title .el-icon {
  color: var(--primary);
  font-size: 16px;
}

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
  border: 2px solid var(--border-default);
  background: var(--bg-base);
  transition: all 0.25s ease;
  text-align: center;
}

.mode-card:hover {
  border-color: var(--primary);
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
}

.mode-card.active {
  border-color: var(--primary);
  background: var(--primary-light);
  box-shadow: 0 0 0 1px var(--primary);
}

.mode-preview {
  width: 100%;
  height: 60px;
  border-radius: 6px;
  overflow: hidden;
  position: relative;
  margin-bottom: 8px;
}

/* 浅白色主题预览 */
.mode-preview.light {
  background: #FFFFFF;
  border: 1px solid #E2E8F0;
}

.mode-preview.light .preview-header {
  height: 12px;
  background: #F8FAFC;
  border-bottom: 1px solid #E2E8F0;
}

.mode-preview.light .preview-sidebar {
  position: absolute;
  left: 0;
  top: 12px;
  bottom: 0;
  width: 20px;
  background: #F8FAFC;
  border-right: 1px solid #E2E8F0;
}

.mode-preview.light .preview-content {
  margin-left: 20px;
  padding: 6px;
}

.mode-preview.light .preview-card {
  height: 14px;
  background: #F1F5F9;
  border-radius: 2px;
  margin-bottom: 4px;
  border: 1px solid #E2E8F0;
}

/* 极客风主题预览 */
.mode-preview.dark {
  background: #0F172A;
  border: 1px solid #3B82F6;
  box-shadow: 0 0 8px rgba(59, 130, 246, 0.3);
}

.mode-preview.dark .preview-header {
  height: 12px;
  background: #1B2336;
  border-bottom: 1px solid #3B82F6;
  box-shadow: 0 0 4px rgba(0, 212, 255, 0.2);
}

.mode-preview.dark .preview-sidebar {
  position: absolute;
  left: 0;
  top: 12px;
  bottom: 0;
  width: 20px;
  background: linear-gradient(180deg, #1B2336 0%, #0F172A 100%);
  border-right: 1px solid #3B82F6;
}

.mode-preview.dark .preview-content {
  margin-left: 20px;
  padding: 6px;
}

.mode-preview.dark .preview-card {
  height: 14px;
  background: #272F42;
  border-radius: 2px;
  margin-bottom: 4px;
  border: 1px solid #475569;
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
  color: var(--text-primary);
  font-weight: 500;
  margin-top: 4px;
}

.check-icon {
  position: absolute;
  top: 6px;
  right: 6px;
  width: 18px;
  height: 18px;
  background: var(--primary);
  color: #fff;
  border-radius: 50%;
  font-size: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
}

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
  transition: all 0.25s ease;
  position: relative;
  margin: 0 auto;
  border: 2px solid transparent;
}

.color-item:hover {
  transform: scale(1.1);
  box-shadow: var(--shadow-md);
}

.color-item.active {
  border-color: var(--text-primary);
  box-shadow: 0 0 0 3px var(--bg-elevated), 0 0 0 5px var(--primary);
}

.color-item .el-icon {
  color: #fff;
  font-size: 20px;
  font-weight: bold;
}

.setting-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 0;
  border-bottom: 1px solid var(--border-light);
}

.setting-item:last-child {
  border-bottom: none;
}

.setting-item span {
  font-size: 14px;
  color: var(--text-primary);
}

.reset-section {
  padding: 20px;
  text-align: center;
}

.reset-section .el-button {
  width: 100%;
}

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
