/**
 * 用户状态管理
 * 管理用户信息、菜单等用户相关状态。
 */
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Router } from 'vue-router'
import { useAppStore } from './app'

/**
 * 用户信息接口
 */
export interface UserInfo {
  id: number
  username: string
  nickname: string
  intro: string
  webSite: string
  avatar: string
  email: string
  role: string
  token: string
  roles?: string[]
}

/**
 * 登录参数接口
 */
export interface LoginParams {
  username: string
  password: string
  captcha?: string
}

export const useUserStore = defineStore('user', () => {
  // ==================== State ====================

  const userInfo = ref<UserInfo | null>(null)
  const userMenus = ref<MenuItem[]>([])

  // ==================== Getters ====================

  const isLoggedIn = computed<boolean>(() => !!userInfo.value)

  const token = computed<string>(() => userInfo.value?.token || '')

  const userId = computed<number | undefined>(() => userInfo.value?.id)

  const nickname = computed<string>(() => userInfo.value?.nickname || '')

  // ==================== Actions ====================

  /**
   * 用户登录
   * @param user - 用户信息
   */
  const login = (user: UserInfo): void => {
    sessionStorage.setItem('token', user.token)
    userInfo.value = user
  }

  /**
   * 用户登出
   */
  const logout = (): void => {
    userInfo.value = null
    userMenus.value = []
    sessionStorage.removeItem('token')

    // 重置标签页
    const appStore = useAppStore()
    appStore.resetAppState()

    // 跳转到登录页
    const router: Router = (window as any).$router
    router?.push({ path: '/login' })
  }

  /**
   * 更新用户头像
   * @param avatar - 头像 URL
   */
  const updateAvatar = (avatar: string): void => {
    if (userInfo.value) {
      userInfo.value.avatar = avatar
    }
  }

  /**
   * 更新用户信息
   * @param user - 用户信息
   */
  const updateUserInfo = (user: Partial<UserInfo>): void => {
    if (userInfo.value) {
      if (user.nickname !== undefined) userInfo.value.nickname = user.nickname
      if (user.intro !== undefined) userInfo.value.intro = user.intro
      if (user.webSite !== undefined) userInfo.value.webSite = user.webSite
    }
  }

  /**
   * 保存用户菜单
   * @param menus - 菜单列表
   */
  const saveUserMenus = (menus: MenuItem[]): void => {
    userMenus.value = menus
  }

  return {
    // State
    userInfo,
    userMenus,
    // Getters
    isLoggedIn,
    token,
    userId,
    nickname,
    // Actions
    login,
    logout,
    updateAvatar,
    updateUserInfo,
    saveUserMenus
  }
}, {
  persist: {
    storage: sessionStorage,
    paths: ['userInfo', 'userMenus']
  }
})
