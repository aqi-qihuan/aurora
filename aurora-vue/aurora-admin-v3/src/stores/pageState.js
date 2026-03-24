import { defineStore } from 'pinia'
import { ref } from 'vue'

/**
 * 页面状态管理
 * 管理各页面的分页状态，支持持久化
 */
export const usePageStateStore = defineStore('pageState', () => {
  // ==================== State ====================
  
  /**
   * 各页面的分页状态
   * 使用单一对象管理所有页面状态，避免重复代码
   */
  const pageState = ref({
    articleList: 1,
    category: 1,
    tag: 1,
    comment: 1,
    talkList: 1,
    user: 1,
    online: 1,
    role: 1,
    quartz: 1,
    friendLink: 1,
    operationLog: 1,
    exceptionLog: 1,
    quartzLog: {
      jobId: -1,
      current: 1
    },
    photo: {
      albumId: -1,
      current: 1
    }
  })

  // ==================== Actions ====================

  /**
   * 更新页面状态（通用方法）
   * @param {string} pageName - 页面名称
   * @param {number|Object} value - 状态值
   */
  const updatePageState = (pageName, value) => {
    if (pageName in pageState.value) {
      pageState.value[pageName] = value
    }
  }

  /**
   * 获取页面状态
   * @param {string} pageName - 页面名称
   * @returns {number|Object} 状态值
   */
  const getPageState = (pageName) => {
    return pageState.value[pageName] || 1
  }

  /**
   * 更新照片页面状态
   * @param {Object} value - { albumId, current }
   */
  const updatePhotoPageState = (value) => {
    pageState.value.photo = value
  }

  /**
   * 更新定时任务日志页面状态
   * @param {number} jobId - 任务ID
   * @param {number} current - 当前页
   */
  const updateQuartzLogState = (jobId, current) => {
    pageState.value.quartzLog = { jobId, current }
  }

  /**
   * 重置所有页面状态
   */
  const resetAllPageState = () => {
    pageState.value = {
      articleList: 1,
      category: 1,
      tag: 1,
      comment: 1,
      talkList: 1,
      user: 1,
      online: 1,
      role: 1,
      quartz: 1,
      friendLink: 1,
      operationLog: 1,
      exceptionLog: 1,
      quartzLog: { jobId: -1, current: 1 },
      photo: { albumId: -1, current: 1 }
    }
  }

  return {
    // State
    pageState,
    // Actions
    updatePageState,
    getPageState,
    resetAllPageState,
    updatePhotoPageState,
    updateQuartzLogState
  }
}, {
  persist: {
    storage: sessionStorage,
    paths: ['pageState']
  }
})
