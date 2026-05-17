/// <reference types="vite/client" />

declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<{}, {}, any>
  export default component
}

// 模块声明（JS → TS 迁移期间使用）
declare module '@/stores/user' {
  import type { UserInfo } from './types/user'
  import { defineStore } from 'pinia'
  import type { Ref } from 'vue'

  export const useUserStore: () => {
    userInfo: Ref<UserInfo | null>
    userMenus: Ref<MenuItem[]>
    isLoggedIn: Readonly<Ref<boolean>>
    token: Readonly<Ref<string>>
    login: (user: UserInfo) => void
    logout: () => void
    updateAvatar: (avatar: string) => void
    updateUserInfo: (user: Partial<UserInfo>) => void
    saveUserMenus: (menus: MenuItem[]) => void
  }
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
  }
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
  export interface RouteMeta {
    title?: string
    hidden?: boolean
    requiresAuth?: boolean
    roles?: string[]
    noAuth?: boolean
  }
}

declare module '@/router' {
  import type { Router } from 'vue-router'
  const router: Router
  export default router
}

declare module '@/utils/request' {
  import type { AxiosInstance } from 'axios'
  import type { ApiResponse } from './types/api'
  const request: AxiosInstance
  export default request
  export interface ApiResponse<T = unknown> {
    code: number
    message: string
    data: T
  }
}
