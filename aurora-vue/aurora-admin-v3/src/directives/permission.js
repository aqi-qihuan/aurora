/**
 * 自定义指令模块
 * 注册全局自定义指令
 */
import { permissionDirective, roleDirective } from '@/utils/auth'

/**
 * 注册全局指令
 * @param {App} app - Vue 应用实例
 */
export function setupDirectives(app) {
  // 注册权限指令
  app.directive('permission', permissionDirective)
  
  // 注册角色指令
  app.directive('role', roleDirective)
}

export default setupDirectives
