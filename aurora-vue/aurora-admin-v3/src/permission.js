/**
 * 路由权限控制
 * 处理动态路由加载和权限验证
 */
import router from '@/router'
import { useUserStore } from '@/stores/user'
import Layout from '@/layout/index.vue'
import NProgress from 'nprogress'
import 'nprogress/nprogress.css'
import request from '@/utils/request'
import logger from '@/utils/logger'

// 配置 NProgress
NProgress.configure({
  easing: 'ease',
  speed: 500,
  showSpinner: false,
  trickleSpeed: 200,
  minimum: 0.3
})

// 白名单路由(不需要登录)
const whiteList = ['/login', '/404', '/403']

/**
 * 添加固定静态路由（不受动态路由通配符影响）
 * 必须在通配符路由之前注册
 */
const addStaticRoutes = () => {
  // 文章编辑/发布页面 + 定时任务调度日志 - 添加到 Layout 的子路由中
  router.addRoute({
    path: '/',
    component: Layout,
    children: [
      {
        path: 'articles/:id',
        name: 'ArticleEdit',
        component: () => import('@/views/article/Article.vue'),
        meta: { title: '发布文章' }
      },
      {
        path: 'quartz/log/all',
        name: 'QuartzLogAll',
        component: () => import('@/views/log/QuartzLog.vue'),
        meta: { title: '调度日志' }
      },
      {
        path: 'quartz/log/:id',
        name: 'QuartzLog',
        component: () => import('@/views/log/QuartzLog.vue'),
        meta: { title: '调度日志' }
      }
    ]
  })
}

// 标记是否已加载动态路由
let dynamicRoutesLoaded = false

/**
 * 动态导入视图组件
 * 显式导入所有需要动态加载的组件（避免 Windows 路径问题）
 */
const modules = {
  // 文章模块
  '/src/views/article/Article.vue': () => import('@/views/article/Article.vue'),
  '/src/views/article/ArticleList.vue': () => import('@/views/article/ArticleList.vue'),
  // 用户模块
  '/src/views/user/User.vue': () => import('@/views/user/User.vue'),
  '/src/views/user/Online.vue': () => import('@/views/user/Online.vue'),
  // 分类和标签
  '/src/views/category/Category.vue': () => import('@/views/category/Category.vue'),
  '/src/views/tag/Tag.vue': () => import('@/views/tag/Tag.vue'),
  // 评论和留言
  '/src/views/comment/Comment.vue': () => import('@/views/comment/Comment.vue'),
  '/src/views/talk/Talk.vue': () => import('@/views/talk/Talk.vue'),
  '/src/views/talk/TalkList.vue': () => import('@/views/talk/TalkList.vue'),
  // 友链和相册
  '/src/views/friendLink/FriendLink.vue': () => import('@/views/friendLink/FriendLink.vue'),
  '/src/views/album/Album.vue': () => import('@/views/album/Album.vue'),
  '/src/views/album/Photo.vue': () => import('@/views/album/Photo.vue'),
  // 资源和日志
  '/src/views/resource/Resource.vue': () => import('@/views/resource/Resource.vue'),
  '/src/views/log/OperationLog.vue': () => import('@/views/log/OperationLog.vue'),
  '/src/views/log/ExceptionLog.vue': () => import('@/views/log/ExceptionLog.vue'),
  // 定时任务
  '/src/views/quartz/Quartz.vue': () => import('@/views/quartz/Quartz.vue'),
  // 系统配置
  '/src/views/website/Website.vue': () => import('@/views/website/Website.vue'),
  '/src/views/setting/Setting.vue': () => import('@/views/setting/Setting.vue'),
  '/src/views/role/Role.vue': () => import('@/views/role/Role.vue'),
  '/src/views/menu/Menu.vue': () => import('@/views/menu/Menu.vue'),
  // 其他
  '/src/views/about/About.vue': () => import('@/views/about/About.vue'),
  '/src/views/album/Delete.vue': () => import('@/views/album/Delete.vue'),
  '/src/views/error/404.vue': () => import('@/views/error/404.vue'),
  '/src/views/error/403.vue': () => import('@/views/error/403.vue')
}

/**
 * 根据组件路径动态加载组件
 * @param {string} componentPath - 后端返回的组件路径
 */
const loadView = (componentPath) => {
  if (!componentPath) return null

  // 规范化路径：移除开头的斜杠，确保有.vue后缀
  let normalizedPath = componentPath.replace(/^\//, '')
  if (!normalizedPath.endsWith('.vue')) {
    normalizedPath += '.vue'
  }

  // 构建完整路径（统一使用正斜杠）
  const fullPath = `/src/views/${normalizedPath}`

  // 尝试精确匹配
  if (modules[fullPath]) {
    return modules[fullPath]
  }

  // 如果精确匹配失败，遍历查找（处理大小写问题）
  const normalizedFullPath = fullPath.toLowerCase()
  for (const [key, moduleLoader] of Object.entries(modules)) {
    if (key.toLowerCase() === normalizedFullPath) {
      return moduleLoader
    }
  }

  logger.warn(`组件不存在: ${fullPath}`)

  // 返回404组件
  const notFoundPath = '/src/views/error/404.vue'
  if (modules[notFoundPath]) {
    return modules[notFoundPath]
  }

  // 如果404组件也不存在，返回一个空组件
  logger.error('404组件不存在！')
  return () => Promise.resolve({
    default: {
      name: 'NotFound',
      template: '<div>页面未找到</div>'
    }
  })
}

/**
 * 处理菜单数据并生成路由
 */
const processMenuData = (menus) => {
  return menus.map(menu => {
    const route = { ...menu }
    
    // 处理图标
    if (route.icon) {
      route.icon = route.icon.replace('iconfont ', '')
    }
    
    // 处理组件
    if (route.component === 'Layout' || !route.component) {
      route.component = Layout
    } else {
      route.component = loadView(route.component)
    }
    
    // 处理子菜单
    if (route.children && route.children.length > 0) {
      route.children = route.children.map(child => {
        const childRoute = { ...child }
        
        // 处理图标
        if (childRoute.icon) {
          childRoute.icon = childRoute.icon.replace('iconfont ', '')
        }
        
        // 处理组件
        if (childRoute.component === 'Layout' || !childRoute.component) {
          childRoute.component = Layout
        } else {
          childRoute.component = loadView(childRoute.component)
        }
        
        return childRoute
      })
    }
    
    return route
  })
}

/**
 * 加载动态路由
 */
const loadDynamicRoutes = async (userStore) => {
  try {
    const { data } = await request.get('/admin/user/menus')
    
    if (data.flag && data.data) {
      logger.log('后端菜单数据:', data.data)
      
      // 处理菜单数据
      const processedMenus = processMenuData(data.data)
      
      // 保存菜单到 store（用于侧边栏显示）
      const menusForStore = data.data.map(menu => {
        const m = { ...menu }
        if (m.icon) {
          m.icon = m.icon.replace('iconfont ', '')
        }
        if (m.children && m.children.length > 0) {
          m.children = m.children.map(child => {
            const c = { ...child }
            if (c.icon) {
              c.icon = c.icon.replace('iconfont ', '')
            }
            return c
          })
        }
        return m
      })
      userStore.saveUserMenus(menusForStore)

      // 动态添加路由
      processedMenus.forEach(route => {
        logger.log('添加路由:', route.path)
        router.addRoute(route)
      })

      // 在通配符之前添加静态路由
      addStaticRoutes()

      // 添加 404 通配符路由（必须最后添加）
      router.addRoute({
        name: 'NotFound',
        path: '/:pathMatch(.*)*',
        redirect: '/404',
        hidden: true
      })

      dynamicRoutesLoaded = true
      logger.log('动态路由加载成功')

      return true
    } else {
      throw new Error(data.message || '获取菜单失败')
    }
  } catch (error) {
    logger.error('加载动态路由失败:', error)
    return false
  }
}

/**
 * 重置动态路由
 */
export const resetDynamicRoutes = () => {
  dynamicRoutesLoaded = false
}

// 全局前置守卫
router.beforeEach(async (to, from, next) => {
  NProgress.start()

  const userStore = useUserStore()
  const hasToken = userStore.token

  if (hasToken) {
    if (to.path === '/login') {
      // 已登录,跳转到首页
      next({ path: '/home' })
      NProgress.done()
    } else if (to.path === '/') {
      // 访问根路径，重定向到首页
      next({ path: '/home' })
      NProgress.done()
    } else {
      // 检查是否已加载动态路由
      if (!dynamicRoutesLoaded) {
        // 加载动态路由
        const success = await loadDynamicRoutes(userStore)

        if (success) {
          // 重新导航到目标路由
          next({ ...to, replace: true })
        } else {
          // 加载失败,跳转到登录页
          userStore.logout()
          next('/login')
        }
      } else {
        next()
      }
    }
  } else {
    // 未登录
    if (whiteList.includes(to.path)) {
      // 在白名单中,直接放行
      next()
    } else {
      // 不在白名单中,跳转到登录页
      next('/login')
      NProgress.done()
    }
  }
})

// 全局后置守卫
router.afterEach(() => {
  NProgress.done()
})

export default router
