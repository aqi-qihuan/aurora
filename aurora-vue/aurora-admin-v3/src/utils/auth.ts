/**
 * 权限工具模块
 * 提供权限验证、角色判断等功能。
 */
import { useUserStore } from '@/stores/user'
import type { Directive } from 'vue'

/**
 * 菜单项接口
 */
export interface MenuItem {
  id?: number
  name: string
  path: string
  component?: string
  icon?: string
  children?: MenuItem[]
  hidden?: boolean
  meta?: RouteMeta
}

/**
 * 路由元信息接口
 */
export interface RouteMeta {
  title?: string
  hidden?: boolean
  requiresAuth?: boolean
  roles?: string[]
  noAuth?: boolean
}

/**
 * 检查用户是否有指定权限
 * @param permission - 权限标识或权限列表
 * @returns {boolean}
 */
export function hasPermission(permission: string | string[]): boolean {
  const userStore = useUserStore()
  const userMenus = userStore.userMenus.value

  if (!userMenus || userMenus.length === 0) {
    return false
  }

  // 如果是数组,检查是否包含任一权限
  if (Array.isArray(permission)) {
    return permission.some(p => checkPermissionInMenus(p, userMenus))
  }

  // 单个权限检查
  return checkPermissionInMenus(permission, userMenus)
}

/**
 * 在菜单中检查权限
 * @param permission - 权限标识
 * @param menus - 菜单列表
 * @returns {boolean}
 */
function checkPermissionInMenus(permission: string, menus: MenuItem[]): boolean {
  for (const menu of menus) {
    // 检查当前菜单
    if (menu.path === permission || menu.name === permission) {
      return true
    }

    // 递归检查子菜单
    if (menu.children && menu.children.length > 0) {
      if (checkPermissionInMenus(permission, menu.children)) {
        return true
      }
    }
  }

  return false
}

/**
 * 检查用户是否有指定角色
 * @param role - 角色标识或角色列表
 * @returns {boolean}
 */
export function hasRole(role: string | string[]): boolean {
  const userStore = useUserStore()
  const userInfo = userStore.userInfo.value

  if (!userInfo || !userInfo.roles) {
    return false
  }

  // 如果是数组,检查是否包含任一角色
  if (Array.isArray(role)) {
    return role.some(r => userInfo.roles!.includes(r))
  }

  // 单个角色检查
  return userInfo.roles.includes(role)
}

/**
 * 检查用户是否为管理员
 * @returns {boolean}
 */
export function isAdmin(): boolean {
  return hasRole('admin') || hasRole('ROLE_ADMIN')
}

/**
 * 获取用户所有权限列表
 * @returns {Array<string>}
 */
export function getPermissions(): string[] {
  const userStore = useUserStore()
  const userMenus = userStore.userMenus.value
  const permissions: string[] = []

  const traverseMenus = (menus: MenuItem[]): void => {
    menus.forEach(menu => {
      if (menu.path) {
        permissions.push(menu.path)
      }
      if (menu.name) {
        permissions.push(menu.name)
      }
      if (menu.children && menu.children.length > 0) {
        traverseMenus(menu.children)
      }
    })
  }

  if (userMenus && userMenus.length > 0) {
    traverseMenus(userMenus)
  }

  return [...new Set(permissions)]
}

/**
 * 检查路由是否有访问权限
 * @param route - 路由对象
 * @returns {boolean}
 */
export function checkRoutePermission(route: { path?: string; meta?: RouteMeta }): boolean {
  // 如果路由标记为不需要权限,直接放行
  if (route.meta?.noAuth) {
    return true
  }

  // 检查路由路径权限
  if (route.path) {
    return hasPermission(route.path)
  }

  // 默认放行
  return true
}

/**
 * 权限指令
 * 用于在模板中控制元素显示
 *
 * 使用示例:
 * <el-button v-permission="'/users'">用户管理</el-button>
 * <el-button v-permission="['/users', '/roles']">用户或角色管理</el-button>
 */
export const permissionDirective: Directive = {
  mounted(el: HTMLElement, binding: { value: string | string[] }) {
    const { value } = binding

    if (value) {
      const hasAuth = hasPermission(value)

      if (!hasAuth) {
        // 没有权限,移除元素
        el.parentNode?.removeChild(el)
      }
    }
  }
}

/**
 * 角色指令
 * 用于在模板中根据角色控制元素显示
 *
 * 使用示例:
 * <el-button v-role="'admin'">管理员按钮</el-button>
 * <el-button v-role="['admin', 'editor']">管理员或编辑按钮</el-button>
 */
export const roleDirective: Directive = {
  mounted(el: HTMLElement, binding: { value: string | string[] }) {
    const { value } = binding

    if (value) {
      const hasAuth = hasRole(value)

      if (!hasAuth) {
        // 没有角色,移除元素
        el.parentNode?.removeChild(el)
      }
    }
  }
}

/**
 * 获取认证请求头
 */
export const getAuthHeaders = (): Record<string, string> => ({
  Authorization: 'Bearer ' + sessionStorage.getItem('token')
})

export default {
  hasPermission,
  hasRole,
  isAdmin,
  getPermissions,
  checkRoutePermission,
  permissionDirective,
  roleDirective,
  getAuthHeaders
}
