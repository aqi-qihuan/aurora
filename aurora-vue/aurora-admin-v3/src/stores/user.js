import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useAppStore } from './app'

/**
 * 用户状态管理
 * 管理用户信息、菜单等用户相关状态
 */
export const useUserStore = defineStore('user', () => {
  // ==================== State ====================
  
  // 用户信息
  const userInfo = ref(null)
  
  // 用户菜单
  const userMenus = ref([])

  // ==================== Getters ====================

  /**
   * 是否已登录
   */
  const isLoggedIn = computed(() => !!userInfo.value)
  
  /**
   * 用户 token
   */
  const token = computed(() => userInfo.value?.token || '')

  // ==================== Actions ====================

  /**
   * 用户登录
   * @param {Object} user - 用户信息
   */
  const login = (user) => {
    sessionStorage.setItem('token', user.token)
    userInfo.value = user
  }

  /**
   * 用户登出
   */
  const logout = () => {
    userInfo.value = null
    userMenus.value = []
    sessionStorage.removeItem('token')
    
    // 重置标签页
    const appStore = useAppStore()
    appStore.resetTab()
  }

  /**
   * 更新用户头像
   * @param {string} avatar - 头像 URL
   */
  const updateAvatar = (avatar) => {
    if (userInfo.value) {
      userInfo.value.avatar = avatar
    }
  }

  /**
   * 更新用户信息
   * @param {Object} user - 用户信息
   */
  const updateUserInfo = (user) => {
    if (userInfo.value) {
      userInfo.value.nickname = user.nickname
      userInfo.value.intro = user.intro
      userInfo.value.webSite = user.webSite
    }
  }

  /**
   * 保存用户菜单
   * @param {Array} menus - 菜单列表
   */
  const saveUserMenus = (menus) => {
    userMenus.value = menus
  }

  return {
    // State
    userInfo,
    userMenus,
    // Getters
    isLoggedIn,
    token,
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
