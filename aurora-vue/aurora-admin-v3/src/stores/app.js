import { defineStore } from 'pinia'
import { ref } from 'vue'

/**
 * 应用状态管理
 * 管理侧边栏折叠、标签页等全局 UI 状态
 */
export const useAppStore = defineStore('app', () => {
  // ==================== State ====================
  
  // 侧边栏折叠状态
  const collapse = ref(false)
  
  // 标签页列表
  const tabList = ref([
    { path: '/home', name: '首页' }
  ])

  // ==================== Actions ====================

  /**
   * 切换侧边栏折叠状态
   */
  const toggleCollapse = () => {
    collapse.value = !collapse.value
  }

  /**
   * 添加标签页
   * @param {Object} route - 路由对象
   */
  const saveTab = (route) => {
    // 检查是否已存在
    const exists = tabList.value.some(tab => tab.path === route.path)

    if (!exists && route.path !== '/login') {
      tabList.value.push({
        path: route.path,
        name: route.meta?.title || route.name || '未命名'
      })
    }
  }

  /**
   * 移除标签页
   * @param {Object} tab - 标签页对象
   */
  const removeTab = (tab) => {
    const index = tabList.value.findIndex(t => t.path === tab.path)
    if (index > -1) {
      tabList.value.splice(index, 1)
    }
  }

  /**
   * 重置标签页列表
   */
  const resetTab = () => {
    tabList.value = [{ path: '/home', name: '首页' }]
  }

  return {
    // State
    collapse,
    tabList,
    // Actions
    toggleCollapse,
    saveTab,
    removeTab,
    resetTab
  }
}, {
  persist: {
    storage: sessionStorage,
    paths: ['collapse', 'tabList']
  }
})
