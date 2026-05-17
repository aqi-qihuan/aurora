/**
 * 权限状态管理
 * 管理用户权限、角色等信息。
 */
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { getPermissions } from '@/utils/auth'

/**
 * 权限状态接口
 */
export interface PermissionState {
  permissions: string[]
  roles: string[]
  buttonPermissions: string[]
}

export const usePermissionStore = defineStore('permission', () => {
  // ==================== State ====================

  const permissions = ref<string[]>([])
  const roles = ref<string[]>([])
  const buttonPermissions = ref<string[]>([])

  // ==================== Getters ====================

  const hasPermissions = computed<boolean>(() => permissions.value.length > 0)
  const hasRoles = computed<boolean>(() => roles.value.length > 0)

  // ==================== Actions ====================

  /**
   * 设置权限列表
   */
  const setPermissions = (perms: string[]): void => {
    permissions.value = perms
  }

  /**
   * 设置角色列表
   */
  const setRoles = (roleList: string[]): void => {
    roles.value = roleList
  }

  /**
   * 设置按钮权限
   */
  const setButtonPermissions = (buttons: string[]): void => {
    buttonPermissions.value = buttons
  }

  /**
   * 初始化权限
   * 从用户菜单中提取权限
   */
  const initPermissions = (): void => {
    const perms = getPermissions()
    setPermissions(perms)
  }

  /**
   * 检查是否有权限
   */
  const checkPermission = (permission: string): boolean => {
    if (!permission) return true
    return permissions.value.includes(permission)
  }

  /**
   * 检查是否有角色
   */
  const checkRole = (role: string): boolean => {
    if (!role) return true
    return roles.value.includes(role)
  }

  /**
   * 清空权限
   */
  const clearPermissions = (): void => {
    permissions.value = []
    roles.value = []
    buttonPermissions.value = []
  }

  return {
    // State
    permissions,
    roles,
    buttonPermissions,
    // Getters
    hasPermissions,
    hasRoles,
    // Actions
    setPermissions,
    setRoles,
    setButtonPermissions,
    initPermissions,
    checkPermission,
    checkRole,
    clearPermissions
  }
}, {
  persist: {
    storage: sessionStorage,
    paths: ['permissions', 'roles', 'buttonPermissions']
  }
})
