import { createRouter, createWebHistory } from 'vue-router'
import Layout from '@/layout/index.vue'

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
        component: () => import('@/views/home/Home.vue'),
        meta: { title: '首页' }
      }
    ]
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/login/Login.vue'),
    hidden: true,
    meta: { title: '登录' }
  },
  {
    path: '/article/:id',
    name: 'ArticleEdit',
    component: () => import('@/views/article/Article.vue'),
    hidden: true,
    meta: { title: '编辑文章' }
  },
  {
    path: '/articles/write',
    name: 'ArticleWrite',
    component: () => import('@/views/article/Article.vue'),
    hidden: true,
    meta: { title: '写文章' }
  },
  {
    path: '/albums/:albumId',
    name: 'AlbumPhoto',
    component: () => import('@/views/album/Photo.vue'),
    hidden: true,
    meta: { title: '相册照片' }
  },
  {
    path: '/photos/delete',
    name: 'PhotoDelete',
    component: () => import('@/views/album/Delete.vue'),
    hidden: true,
    meta: { title: '照片回收站' }
  },
  {
    path: '/talks/:talkId',
    name: 'TalkEdit',
    component: () => import('@/views/talk/Talk.vue'),
    hidden: true,
    meta: { title: '编辑说说' }
  },
  {
    path: '/quartz/log/:quartzId',
    name: 'QuartzLog',
    component: () => import('@/views/log/QuartzLog.vue'),
    hidden: true,
    meta: { title: '调度日志' }
  },
  {
    path: '/404',
    name: '404',
    component: () => import('@/views/error/404.vue'),
    hidden: true,
    meta: { title: '404' }
  },
  {
    path: '/403',
    name: '403',
    component: () => import('@/views/error/403.vue'),
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

export default router
