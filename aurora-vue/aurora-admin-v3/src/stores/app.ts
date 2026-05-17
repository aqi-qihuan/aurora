/**
 * 应用状态管理
 * 管理主题、侧边栏、布局、标签页等应用级状态。
 */
import { defineStore } from 'pinia'
import { ref, computed, watch } from 'vue'
import { useUserStore } from './user'

/**
 * 主题模式
 */
export type ThemeMode = 'light' | 'dark' | 'system'

/**
 * 标签页接口
 */
export interface TabItem {
  name: string
  path: string
}

export const useAppStore = defineStore('app', () => {
  // ==================== State ====================

  const isDark = ref<boolean>(false)
  const sidebarOpen = ref<boolean>(true)
  const themeColor = ref<string>('#409EFF')
  const themeMode = ref<ThemeMode>('system')
  const sidebarWidth = ref<number>(210)
  const showSettings = ref<boolean>(false)
  const fixedHeader = ref<boolean>(true)
  const showTabs = ref<boolean>(true)
  const showFooter = ref<boolean>(true)
  const tabList = ref<TabItem[]>([{ name: '首页', path: '/home' }])

  // ==================== Getters ====================

  const sidebarWidthPx = computed<string>(() => `${sidebarWidth.value}px`)
  const isMobile = computed<boolean>(() => window.innerWidth < 768)
  const currentTheme = computed<'light' | 'dark'>(() => {
    if (themeMode.value === 'system') {
      return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
    }
    return themeMode.value === 'dark' ? 'dark' : 'light'
  })
  const collapse = computed<boolean>(() => !sidebarOpen.value)

  // ==================== Actions ====================

  /**
   * 切换侧边栏
   */
  const toggleSidebar = (): void => {
    sidebarOpen.value = !sidebarOpen.value
  }

  /**
   * 切换折叠状态（toggleSidebar 的别名）
   */
  const toggleCollapse = (): void => {
    sidebarOpen.value = !sidebarOpen.value
  }

  /**
   * 设置侧边栏状态
   */
  const setSidebarOpen = (open: boolean): void => {
    sidebarOpen.value = open
  }

  /**
   * 切换主题模式
   */
  const toggleTheme = (): void => {
    isDark.value = !isDark.value
    applyTheme()
  }

  /**
   * 设置主题模式
   */
  const setThemeMode = (mode: ThemeMode): void => {
    themeMode.value = mode
    applyTheme()
  }

  /**
   * 设置主题色
   */
  const setThemeColor = (color: string): void => {
    themeColor.value = color
    applyThemeColor(color)
  }

  /**
   * 应用主题到 DOM
   */
  const applyTheme = (): void => {
    const theme = currentTheme.value
    if (theme === 'dark') {
      document.documentElement.setAttribute('data-theme', 'dark')
    } else {
      document.documentElement.removeAttribute('data-theme')
    }
    localStorage.setItem('theme-mode', themeMode.value)
    localStorage.setItem('theme-color', themeColor.value)
  }

  /**
   * 应用主题色
   */
  const applyThemeColor = (color: string): void => {
    document.documentElement.style.setProperty('--el-color-primary', color)
    // 生成浅色变量
    for (let i = 1; i <= 9; i++) {
      document.documentElement.style.setProperty(
        `--el-color-primary-light-${i}`,
        mixColor(color, '#ffffff', i / 10)
      )
    }
  }

  /**
   * 混合颜色
   */
  const mixColor = (color1: string, color2: string, weight: number): string => {
    // Simplified color mixing - in production, use a proper color lib
    return color1 // placeholder
  }

  /**
   * 重置应用状态
   */
  const resetAppState = (): void => {
    isDark.value = false
    sidebarOpen.value = true
    themeColor.value = '#409EFF'
    themeMode.value = 'system'
    showSettings.value = false
  }

  /**
   * 初始化应用状态
   */
  const initAppState = (): void => {
    const savedMode = localStorage.getItem('theme-mode') as ThemeMode | null
    const savedColor = localStorage.getItem('theme-color')

    if (savedMode) {
      themeMode.value = savedMode
    }
    if (savedColor) {
      themeColor.value = savedColor
    }

    applyTheme()
    applyThemeColor(themeColor.value)
  }

  /**
   * 保存标签页
   */
  const saveTab = (route: any): void => {
    const { name, path } = route
    if (!name) return
    
    const existingTab = tabList.value.find(tab => tab.path === path)
    if (!existingTab) {
      tabList.value.push({ name: name as string, path })
    }
  }

  /**
   * 移除标签页
   */
  const removeTab = (tab: TabItem): void => {
    const index = tabList.value.findIndex(t => t.path === tab.path)
    if (index > -1) {
      tabList.value.splice(index, 1)
    }
  }

  /**
   * 关闭所有标签页
   */
  const closeAllTabs = (): void => {
    tabList.value = [{ name: '首页', path: '/home' }]
  }

  /**
   * 重置标签页（关闭所有标签页的别名）
   */
  const resetTab = (): void => {
    tabList.value = [{ name: '首页', path: '/home' }]
  }

  // 初始化
  initAppState()

  return {
    // State
    isDark,
    sidebarOpen,
    themeColor,
    themeMode,
    sidebarWidth,
    showSettings,
    fixedHeader,
    showTabs,
    showFooter,
    tabList,
    // Getters
    sidebarWidthPx,
    isMobile,
    currentTheme,
    collapse,
    // Actions
    toggleSidebar,
    toggleCollapse,
    setSidebarOpen,
    toggleTheme,
    setThemeMode,
    setThemeColor,
    applyTheme,
    resetAppState,
    initAppState,
    saveTab,
    removeTab,
    closeAllTabs,
    resetTab
  }
}, {
  persist: {
    storage: localStorage,
    paths: ['isDark', 'sidebarOpen', 'themeColor', 'themeMode', 'sidebarWidth']
  }
})
