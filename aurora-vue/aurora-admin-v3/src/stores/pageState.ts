/**
 * 页面状态管理
 * 管理各页面的分页状态，支持持久化。
 */
import { defineStore } from 'pinia'
import { ref } from 'vue'

/**
 * 照片页面状态接口
 */
interface PhotoPageState {
  albumId: number
  current: number
}

/**
 * 定时任务日志页面状态接口
 */
interface QuartzLogPageState {
  jobId: number
  current: number
}

/**
 * 页面状态接口
 */
interface PageState {
  articleList: number
  category: number
  tag: number
  comment: number
  talkList: number
  user: number
  online: number
  role: number
  quartz: number
  friendLink: number
  operationLog: number
  exceptionLog: number
  quartzLog: QuartzLogPageState
  photo: PhotoPageState
}

export const usePageStateStore = defineStore('pageState', () => {
  // ==================== State ====================

  const pageState = ref<PageState>({
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
   * @param pageName - 页面名称
   * @param value - 状态值
   */
  const updatePageState = (pageName: keyof PageState, value: number | PhotoPageState | QuartzLogPageState): void => {
    ;(pageState.value as any)[pageName] = value
  }

  /**
   * 获取页面状态
   * @param pageName - 页面名称
   * @returns 状态值
   */
  const getPageState = (pageName: keyof PageState): number | PhotoPageState | QuartzLogPageState => {
    return (pageState.value as any)[pageName] || 1
  }

  /**
   * 更新照片页面状态
   * @param value - { albumId, current }
   */
  const updatePhotoPageState = (value: PhotoPageState): void => {
    pageState.value.photo = value
  }

  /**
   * 更新定时任务日志页面状态
   * @param jobId - 任务ID
   * @param current - 当前页
   */
  const updateQuartzLogState = (jobId: number, current: number): void => {
    pageState.value.quartzLog = { jobId, current }
  }

  /**
   * 重置所有页面状态
   */
  const resetAllPageState = (): void => {
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
