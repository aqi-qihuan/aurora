import { createRouter, createWebHistory } from 'vue-router'
import Layout from '@/layout/index.vue'
// 静态导入静态路由组件（消除动态/静态导入冲突警告）
import Home from '@/views/home/Home.vue'
import Login from '@/views/login/Login.vue'
import NotFound from '@/views/error/404.vue'
import Forbidden from '@/views/error/403.vue'

/**
 * 静态路由配置
 * 只包含基础路由，其他路由从后端动态加载
 */
export const constantRoutes = [
  {
    path: '/',
    component: Layout,
    redirect: '/home',
    children: [
      {
        path: 'home',
        name: 'Home',
        component: Home,
        meta: { title: '首页' }
      }
    ]
  },
  {
    path: '/login',
    name: 'Login',
    component: Login,
    hidden: true,
    meta: { title: '登录' }
  },
  {
    path: '/404',
    name: '404',
    component: NotFound,
    hidden: true,
    meta: { title: '404' }
  },
  {
    path: '/403',
    name: '403',
    component: Forbidden,
    hidden: true,
    meta: { title: '403' }
  }
]

/**
 * 创建路由实例
 */
const router = createRouter({
  history: createWebHistory(),
  routes: constantRoutes,
  scrollBehavior() {
    return { top: 0 }
  }
})

/**
 * 重置路由
 */
export function resetRouter() {
  const newRouter = createRouter({
    history: createWebHistory(),
    routes: constantRoutes
  })
  router.matcher = newRouter.matcher
}

export default router
