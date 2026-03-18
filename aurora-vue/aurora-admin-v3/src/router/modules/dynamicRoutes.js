/**
 * 动态路由模块
 * 负责根据后端返回的菜单数据动态生成路由
 */
import Layout from '@/layout/index.vue'

/**
 * 动态导入视图组件
 * 使用 Vite 的 import.meta.glob 实现动态导入
 */
const modules = import.meta.glob('../views/**/*.vue')

/**
 * 根据组件路径动态加载组件
 * @param {string} componentPath - 组件路径,如 '/article/ArticleList.vue'
 * @returns {Function} 组件加载函数
 */
export const loadView = (componentPath) => {
  // 确保路径格式正确
  let path = componentPath
  
  // 处理不同的路径格式
  if (!path.startsWith('/')) {
    path = '/' + path
  }
  
  // 构建 Vite 动态导入路径
  const importPath = `../views${path}`
  
  // 检查是否存在对应的组件
  if (modules[importPath]) {
    return modules[importPath]
  }
  
  // 如果找不到组件,返回 404 页面
  console.warn(`Component not found: ${importPath}`)
  return () => import('@/views/error/404.vue')
}

/**
 * 生成路由配置
 * @param {Array} menus - 后端返回的菜单数据
 * @returns {Array} 路由配置数组
 */
export const generateRoutes = (menus) => {
  const routes = []
  
  menus.forEach(menu => {
    const route = {
      path: menu.path,
      name: menu.name,
      meta: {
        title: menu.name,
        icon: menu.icon,
        hidden: menu.hidden || false,
        ...menu.meta
      }
    }
    
    // 处理组件
    if (menu.component === 'Layout' || !menu.component) {
      route.component = Layout
    } else {
      route.component = loadView(menu.component)
    }
    
    // 处理重定向
    if (menu.redirect) {
      route.redirect = menu.redirect
    }
    
    // 处理子菜单
    if (menu.children && menu.children.length > 0) {
      route.children = generateRoutes(menu.children)
    }
    
    routes.push(route)
  })
  
  return routes
}

/**
 * 扁平化路由(用于权限判断)
 * @param {Array} routes - 路由配置数组
 * @returns {Array} 扁平化的路由路径数组
 */
export const flattenRoutes = (routes) => {
  const result = []
  
  routes.forEach(route => {
    if (route.path) {
      result.push(route.path)
    }
    if (route.children && route.children.length > 0) {
      result.push(...flattenRoutes(route.children))
    }
  })
  
  return result
}

/**
 * 获取所有路由路径(用于权限判断)
 * @param {Array} menus - 菜单数据
 * @returns {Set} 路由路径集合
 */
export const getAllRoutePaths = (menus) => {
  const paths = new Set()
  
  const traverse = (menuList) => {
    menuList.forEach(menu => {
      if (menu.path) {
        paths.add(menu.path)
      }
      if (menu.children && menu.children.length > 0) {
        traverse(menu.children)
      }
    })
  }
  
  traverse(menus)
  return paths
}

export default {
  loadView,
  generateRoutes,
  flattenRoutes,
  getAllRoutePaths
}
