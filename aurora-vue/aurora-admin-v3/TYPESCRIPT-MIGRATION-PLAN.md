# Aurora Admin V3 - TypeScript 迁移计划

## 📋 概述

将 Aurora Admin V3 项目从 JavaScript 迁移到 TypeScript，提升代码质量、类型安全和开发体验。

## 🎯 迁移目标

- ✅ 完整的 TypeScript 类型支持
- ✅ 更好的 IDE 智能提示
- ✅ 编译时错误检测
- ✅ 增强的代码可维护性
- ✅ 保持现有功能 100% 可用

## 📊 项目现状分析

### 当前技术栈

```json
{
  "vue": "^3.4.21",
  "vite": "^5.1.5",
  "pinia": "^2.1.7",
  "vue-router": "^4.3.0",
  "element-plus": "^2.5.6"
}
```

### 文件统计

| 文件类型 | 数量 | 预计工作量 |
|---------|------|-----------|
| `.vue` 文件 | 43+ | 高 |
| `.js` 文件 | 16+ | 中 |
| 配置文件 | 5+ | 低 |
| **总计** | **64+** | **高** |

## 🔄 迁移阶段

### Phase 1: 基础配置（1-2 天）

#### 1.1 安装 TypeScript 依赖

```bash
npm install -D typescript @types/node
npm install -D @vue/tsconfig
npm install -D vue-tsc
```

#### 1.2 创建 TypeScript 配置文件

**tsconfig.json**
```json
{
  "compilerOptions": {
    "target": "ESNext",
    "useDefineForClassFields": true,
    "module": "ESNext",
    "lib": ["ESNext", "DOM", "DOM.Iterable"],
    "skipLibCheck": true,
    
    /* Bundler mode */
    "moduleResolution": "bundler",
    "allowImportingTsExtensions": true,
    "resolveJsonModule": true,
    "isolatedModules": true,
    "noEmit": true,
    "jsx": "preserve",
    
    /* Linting */
    "strict": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noFallthroughCasesInSwitch": true,
    
    /* Path mapping */
    "baseUrl": ".",
    "paths": {
      "@/*": ["./src/*"]
    },
    
    /* Type definitions */
    "types": ["vite/client", "node", "element-plus/global"]
  },
  "include": [
    "src/**/*.ts",
    "src/**/*.d.ts",
    "src/**/*.tsx",
    "src/**/*.vue"
  ],
  "exclude": ["node_modules", "dist"]
}
```

**tsconfig.node.json**
```json
{
  "compilerOptions": {
    "composite": true,
    "module": "ESNext",
    "moduleResolution": "bundler",
    "allowSyntheticDefaultImports": true,
    "strict": true
  },
  "include": ["vite.config.ts", "vitest.config.ts"]
}
```

#### 1.3 创建类型定义文件

**src/env.d.ts**
```typescript
/// <reference types="vite/client" />

declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<{}, {}, any>
  export default component
}

interface ImportMetaEnv {
  readonly VITE_APP_TITLE: string
  readonly VITE_API_BASE_URL: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
```

#### 1.4 更新 vite.config.js → vite.config.ts

```typescript
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'
import type { UserConfig } from 'vite'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src')
    }
  }
} as UserConfig)
```

#### 1.5 更新 package.json scripts

```json
{
  "scripts": {
    "dev": "vite",
    "build": "vue-tsc && vite build",
    "preview": "vite preview",
    "type-check": "vue-tsc --noEmit"
  }
}
```

---

### Phase 2: 类型定义准备（2-3 天）

#### 2.1 创建类型定义目录结构

```
src/types/
├── index.ts              # 统一导出
├── api.ts                # API 响应类型
├── user.ts               # 用户相关类型
├── article.ts            # 文章相关类型
├── category.ts           # 分类相关类型
├── tag.ts                # 标签相关类型
├── comment.ts            # 评论相关类型
├── resource.ts           # 资源相关类型
├── menu.ts               # 菜单相关类型
├── role.ts               # 角色相关类型
├── quartz.ts             # 定时任务类型
├── router.ts             # 路由相关类型
├── store.ts              # Store 状态类型
└── global.d.ts           # 全局类型声明
```

#### 2.2 核心 API 类型定义

**src/types/api.ts**
```typescript
// 通用 API 响应类型
export interface ApiResponse<T = any> {
  code: number
  message: string
  data: T
}

// 分页请求参数
export interface PageParams {
  pageNum?: number
  pageSize?: number
}

// 分页响应数据
export interface PageData<T> {
  list: T[]
  total: number
  pageNum: number
  pageSize: number
}

// 列表 API 响应
export type ListResponse<T> = ApiResponse<PageData<T>>
```

#### 2.3 用户模块类型定义

**src/types/user.ts**
```typescript
// 用户信息
export interface UserInfo {
  id: number
  username: string
  nickname: string
  avatar: string
  email?: string
  loginType: LoginType
  roles: Role[]
  isDisable: 0 | 1
  ipAddress?: string
  ipSource?: string
  createTime: string
  updateTime?: string
}

// 登录类型
export type LoginType = 1 | 2 | 3  // 1: 用户名 2: QQ 3: 微博

export interface LoginTypeOption {
  type: LoginType
  desc: string
  icon: string
}

// 用户查询参数
export interface UserQuery extends PageParams {
  keywords?: string
  loginType?: LoginType
}

// 用户登录表单
export interface LoginForm {
  username: string
  password: string
}
```

#### 2.4 文章模块类型定义

**src/types/article.ts**
```typescript
// 文章信息
export interface Article {
  id: number
  title: string
  content: string
  summary?: string
  cover?: string
  categoryId: number
  categoryName?: string
  tags: Tag[]
  isTop: 0 | 1
  isDelete: 0 | 1
  status: ArticleStatus
  viewCount: number
  likeCount: number
  commentCount: number
  createTime: string
  updateTime?: string
}

// 文章状态
export type ArticleStatus = 1 | 2 | 3  // 1: 公开 2: 私密 3: 草稿

// 文章查询参数
export interface ArticleQuery extends PageParams {
  keywords?: string
  categoryId?: number
  tagId?: number
  status?: ArticleStatus
  isDelete?: 0 | 1
}

// 文章表单
export interface ArticleForm {
  id?: number
  title: string
  content: string
  summary?: string
  cover?: string
  categoryId: number
  tagIds: number[]
  isTop: 0 | 1
  status: ArticleStatus
}
```

#### 2.5 分类和标签类型

**src/types/category.ts**
```typescript
export interface Category {
  id: number
  categoryName: string
  parentId?: number
  children?: Category[]
  articleCount?: number
  createTime: string
}

export interface CategoryForm {
  id?: number
  categoryName: string
  parentId?: number
}
```

**src/types/tag.ts**
```typescript
export interface Tag {
  id: number
  tagName: string
  articleCount?: number
  createTime: string
}

export interface TagForm {
  id?: number
  tagName: string
}
```

#### 2.6 菜单和角色类型

**src/types/menu.ts**
```typescript
export interface Menu {
  id: number
  menuName: string
  path?: string
  component?: string
  icon?: string
  parentId: number
  orderNum: number
  isHidden: 0 | 1
  isDisable: 0 | 1
  children?: Menu[]
  createTime: string
}

export interface MenuForm {
  id?: number
  menuName: string
  path?: string
  component?: string
  icon?: string
  parentId: number
  orderNum: number
  isHidden: 0 | 1
  isDisable: 0 | 1
}
```

**src/types/role.ts**
```typescript
export interface Role {
  id: number
  roleName: string
  roleLabel: string
  menus?: number[]
  isDisable: 0 | 1
  createTime: string
}

export interface RoleForm {
  id?: number
  roleName: string
  roleLabel: string
  menuIds: number[]
  isDisable: 0 | 1
}
```

#### 2.7 路由类型定义

**src/types/router.ts**
```typescript
import type { RouteRecordRaw } from 'vue-router'

export interface AppRouteRecordRaw extends Omit<RouteRecordRaw, 'children'> {
  hidden?: boolean
  alwaysShow?: boolean
  meta?: RouteMeta
  children?: AppRouteRecordRaw[]
}

export interface RouteMeta {
  title?: string
  icon?: string
  hidden?: boolean
  keepAlive?: boolean
  requiresAuth?: boolean
  roles?: string[]
}

export type AppRoute = AppRouteRecordRaw
```

#### 2.8 Store 类型定义

**src/types/store.ts**
```typescript
// 用户 Store 状态
export interface UserState {
  userInfo: UserInfo | null
  token: string | null
  isLoggedIn: boolean
}

// 应用 Store 状态
export interface AppState {
  sidebar: boolean
  device: 'desktop' | 'mobile'
  theme: 'light' | 'dark'
}

// 权限 Store 状态
export interface PermissionState {
  routes: AppRouteRecordRaw[]
  dynamicRoutes: AppRouteRecordRaw[]
  permissions: string[]
}
```

---

### Phase 3: 工具函数迁移（2-3 天）

#### 3.1 工具函数迁移优先级

| 文件 | 优先级 | 复杂度 |
|------|--------|--------|
| utils/request.js | 高 | 中 |
| utils/auth.js | 高 | 低 |
| utils/logger.js | 中 | 低 |
| stores/*.js | 高 | 高 |
| router/*.js | 高 | 高 |

#### 3.2 request.js → request.ts 示例

```typescript
import axios, { type AxiosInstance, type AxiosRequestConfig, type AxiosResponse } from 'axios'
import { ElMessage } from 'element-plus'
import router from '@/router'

// API 响应数据结构
interface ApiResponse<T = any> {
  code: number
  message: string
  data: T
}

// 创建 axios 实例
const request: AxiosInstance = axios.create({
  baseURL: '/api',
  timeout: 10000
})

// 请求拦截器
request.interceptors.request.use(
  (config: AxiosRequestConfig) => {
    const token = sessionStorage.getItem('token')
    if (token && config.headers) {
      config.headers['Authorization'] = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
request.interceptors.response.use(
  (response: AxiosResponse<ApiResponse>) => {
    const { data } = response
    
    if (data.code !== 20000 && data.code !== 200) {
      ElMessage.error(data.message || '操作失败')
      
      if (data.code === 40001) {
        sessionStorage.removeItem('token')
        router.push({ path: '/login' })
      }
    }
    
    return response
  },
  (error) => {
    const message = error.response?.data?.message || error.message || '请求失败'
    ElMessage.error(message)
    return Promise.reject(error)
  }
)

export default request
```

#### 3.3 auth.js → auth.ts 示例

```typescript
// Token 存储 key
const TOKEN_KEY = 'token'
const USER_INFO_KEY = 'userInfo'

// 获取 Token
export function getToken(): string | null {
  return sessionStorage.getItem(TOKEN_KEY)
}

// 设置 Token
export function setToken(token: string): void {
  sessionStorage.setItem(TOKEN_KEY, token)
}

// 移除 Token
export function removeToken(): void {
  sessionStorage.removeItem(TOKEN_KEY)
}

// 获取用户信息
export function getUserInfo<T = any>(): T | null {
  const userInfo = sessionStorage.getItem(USER_INFO_KEY)
  return userInfo ? JSON.parse(userInfo) : null
}

// 设置用户信息
export function setUserInfo<T>(userInfo: T): void {
  sessionStorage.setItem(USER_INFO_KEY, JSON.stringify(userInfo))
}

// 移除用户信息
export function removeUserInfo(): void {
  sessionStorage.removeItem(USER_INFO_KEY)
}

// 清除所有认证信息
export function clearAuth(): void {
  removeToken()
  removeUserInfo()
}
```

---

### Phase 4: Pinia Store 迁移（3-4 天）

#### 4.1 user.js → user.ts 示例

```typescript
import { defineStore } from 'pinia'
import type { UserState } from '@/types/store'
import type { UserInfo, LoginForm } from '@/types/user'
import { getToken, setToken, removeToken, setUserInfo, removeUserInfo } from '@/utils/auth'
import request from '@/utils/request'

export const useUserStore = defineStore('user', {
  state: (): UserState => ({
    userInfo: null,
    token: getToken(),
    isLoggedIn: false
  }),
  
  getters: {
    userName: (state): string => state.userInfo?.nickname || 'Guest',
    isAdmin: (state): boolean => state.userInfo?.roles?.some(r => r.roleLabel === 'admin') ?? false
  },
  
  actions: {
    // 登录
    async login(loginForm: LoginForm): Promise<void> {
      const { data } = await request.post('/login', loginForm)
      this.token = data.data.token
      setToken(data.data.token)
    },
    
    // 获取用户信息
    async getUserInfo(): Promise<UserInfo> {
      const { data } = await request.get('/user/info')
      this.userInfo = data.data
      this.isLoggedIn = true
      setUserInfo(data.data)
      return data.data
    },
    
    // 登出
    async logout(): Promise<void> {
      await request.post('/logout')
      this.$reset()
      removeToken()
      removeUserInfo()
    },
    
    // 重置状态
    $reset(): void {
      this.userInfo = null
      this.token = null
      this.isLoggedIn = false
    }
  }
})
```

#### 4.2 app.js → app.ts 示例

```typescript
import { defineStore } from 'pinia'
import type { AppState } from '@/types/store'

export const useAppStore = defineStore('app', {
  state: (): AppState => ({
    sidebar: true,
    device: 'desktop',
    theme: 'light'
  }),
  
  getters: {
    isMobile: (state): boolean => state.device === 'mobile'
  },
  
  actions: {
    toggleSidebar(): void {
      this.sidebar = !this.sidebar
    },
    
    closeSidebar(): void {
      this.sidebar = false
    },
    
    toggleDevice(device: 'desktop' | 'mobile'): void {
      this.device = device
    },
    
    setTheme(theme: 'light' | 'dark'): void {
      this.theme = theme
      document.documentElement.setAttribute('data-theme', theme)
    }
  }
})
```

---

### Phase 5: Vue 组件迁移（5-7 天）

#### 5.1 迁移策略

**渐进式迁移**：每个组件单独迁移，确保功能不受影响

#### 5.2 组件迁移顺序

**优先级分组**：

**Group A - 核心组件（高优先级）**
1. `App.vue` - 应用根组件
2. `layout/index.vue` - 布局组件
3. `login/Login.vue` - 登录页面
4. `home/Home.vue` - 首页

**Group B - 功能页面（中优先级）**
5. `user/User.vue` - 用户管理
6. `article/Article.vue` - 文章管理
7. `category/Category.vue` - 分类管理
8. `tag/Tag.vue` - 标签管理
9. `comment/Comment.vue` - 评论管理
10. `resource/Resource.vue` - 资源管理

**Group C - 其他页面（低优先级）**
11. `role/Role.vue` - 角色管理
12. `menu/Menu.vue` - 菜单管理
13. `website/Website.vue` - 网站配置
14. `album/Album.vue` - 相册管理
15. `quartz/Quartz.vue` - 定时任务

#### 5.3 Vue 组件迁移示例

**Login.vue (JavaScript)**
```javascript
<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const userStore = useUserStore()

const loginForm = ref({
  username: '',
  password: ''
})

const handleLogin = async () => {
  await userStore.login(loginForm.value)
  router.push('/')
}
</script>
```

**Login.vue (TypeScript)**
```typescript
<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/stores/user'
import type { FormInstance, FormRules } from 'element-plus'
import type { LoginForm } from '@/types/user'

const router = useRouter()
const userStore = useUserStore()

// 表单引用
const loginFormRef = ref<FormInstance>()

// 登录表单数据
const loginForm = reactive<LoginForm>({
  username: '',
  password: ''
})

// 表单验证规则
const loginRules: FormRules<LoginForm> = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能小于 6 位', trigger: 'blur' }
  ]
}

// 登录加载状态
const loading = ref(false)

// 处理登录
const handleLogin = async (): Promise<void> => {
  if (!loginFormRef.value) return
  
  await loginFormRef.value.validate(async (valid) => {
    if (!valid) return
    
    loading.value = true
    try {
      await userStore.login(loginForm)
      ElMessage.success('登录成功')
      router.push('/')
    } catch (error) {
      console.error('登录失败:', error)
    } finally {
      loading.value = false
    }
  })
}
</script>
```

#### 5.4 组件迁移检查清单

每个组件迁移后需要检查：

- [ ] `<script setup lang="ts">` 已添加
- [ ] 所有 props 定义类型
- [ ] 所有 emits 定义类型
- [ ] ref/reactive 变量添加类型注解
- [ ] 函数参数和返回值添加类型
- [ ] API 调用使用正确的类型
- [ ] 模板中的变量类型正确
- [ ] 无 TypeScript 编译错误
- [ ] 功能测试通过

---

### Phase 6: 路由迁移（1-2 天）

#### 6.1 router/index.js → router/index.ts

```typescript
import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import type { AppRouteRecordRaw } from '@/types/router'

// 静态路由
export const constantRoutes: AppRouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/login/Login.vue'),
    meta: { title: '登录', hidden: true }
  },
  {
    path: '/404',
    name: '404',
    component: () => import('@/views/error/404.vue'),
    meta: { title: '404', hidden: true }
  }
]

// 动态路由（根据权限动态添加）
export const asyncRoutes: AppRouteRecordRaw[] = [
  // 根据后端返回的菜单动态生成
]

const router = createRouter({
  history: createWebHistory(),
  routes: constantRoutes as RouteRecordRaw[]
})

export default router
```

#### 6.2 dynamicRoutes.js → dynamicRoutes.ts

```typescript
import type { RouteRecordRaw } from 'vue-router'
import type { AppRouteRecordRaw, RouteMeta } from '@/types/router'
import type { Menu } from '@/types/menu'

/**
 * 将后端菜单转换为路由
 */
export function generateRoutes(menus: Menu[]): AppRouteRecordRaw[] {
  const routes: AppRouteRecordRaw[] = []
  
  menus.forEach(menu => {
    const route: AppRouteRecordRaw = {
      path: menu.path || '',
      name: menu.menuName,
      component: loadComponent(menu.component),
      meta: {
        title: menu.menuName,
        icon: menu.icon,
        hidden: menu.isHidden === 1
      } as RouteMeta,
      children: menu.children ? generateRoutes(menu.children) : undefined
    }
    
    routes.push(route)
  })
  
  return routes
}

/**
 * 动态加载组件
 */
function loadComponent(component?: string) {
  if (!component) return undefined
  
  return () => import(`@/views/${component}.vue`)
}

/**
 * 重置路由
 */
export function resetRouter(): void {
  const newRouter = createRouter({
    history: createWebHistory(),
    routes: constantRoutes as RouteRecordRaw[]
  })
  router.matcher = newRouter.matcher
}
```

---

### Phase 7: 测试和验证（2-3 天）

#### 7.1 类型检查

```bash
# 运行类型检查
npm run type-check

# 构建时自动检查
npm run build
```

#### 7.2 测试验证

```bash
# 运行单元测试
npm run test:run

# 运行 E2E 测试（如有）
npm run test:e2e
```

#### 7.3 功能测试清单

**核心功能测试**：
- [ ] 用户登录/登出
- [ ] 路由跳转和权限控制
- [ ] 主题切换
- [ ] 侧边栏展开/收起

**业务功能测试**：
- [ ] 用户管理（列表、编辑、禁用）
- [ ] 文章管理（增删改查）
- [ ] 分类管理
- [ ] 标签管理
- [ ] 评论管理
- [ ] 资源管理
- [ ] 角色管理
- [ ] 菜单管理
- [ ] 网站配置

---

## 📈 迁移进度跟踪

### 总体进度

| 阶段 | 任务 | 预计时间 | 状态 |
|------|------|---------|------|
| Phase 1 | 基础配置 | 1-2 天 | ⏳ 待开始 |
| Phase 2 | 类型定义准备 | 2-3 天 | ⏳ 待开始 |
| Phase 3 | 工具函数迁移 | 2-3 天 | ⏳ 待开始 |
| Phase 4 | Pinia Store 迁移 | 3-4 天 | ⏳ 待开始 |
| Phase 5 | Vue 组件迁移 | 5-7 天 | ⏳ 待开始 |
| Phase 6 | 路由迁移 | 1-2 天 | ⏳ 待开始 |
| Phase 7 | 测试和验证 | 2-3 天 | ⏳ 待开始 |
| **总计** | **完整迁移** | **16-24 天** | **⏳ 待开始** |

### 文件迁移进度

| 类型 | 总数 | 已完成 | 进行中 | 待开始 | 完成率 |
|------|------|--------|--------|--------|--------|
| 配置文件 | 5 | 0 | 0 | 5 | 0% |
| 类型定义 | 12 | 0 | 0 | 12 | 0% |
| 工具函数 | 5 | 0 | 0 | 5 | 0% |
| Store | 4 | 0 | 0 | 4 | 0% |
| 路由 | 3 | 0 | 0 | 3 | 0% |
| Vue 组件 | 43+ | 0 | 0 | 43+ | 0% |

---

## ⚠️ 注意事项

### 迁移原则

1. **渐进式迁移**
   - 不要一次性修改所有文件
   - 每个模块独立迁移，确保功能正常
   - 保持代码可运行状态

2. **类型安全优先**
   - 避免 `any` 类型
   - 使用严格模式 `strict: true`
   - 为所有 API 响应定义类型

3. **保持兼容性**
   - 迁移过程中保持功能可用
   - 确保向后兼容
   - 保留原有代码风格

### 常见问题

1. **第三方库类型**
   ```bash
   # 安装类型定义
   npm install -D @types/xxx
   ```

2. **Vue 组件类型**
   ```typescript
   // 使用 defineProps 和 defineEmits
   interface Props {
     title: string
     count?: number
   }
   
   const props = defineProps<Props>()
   const emit = defineEmits<{
     (e: 'update', value: number): void
   }>()
   ```

3. **Pinia Store 类型**
   ```typescript
   // 使用泛型定义 state 类型
   export const useStore = defineStore('store', {
     state: (): State => ({ ... })
   })
   ```

---

## 🎯 预期收益

### 开发体验提升

- ✅ **更好的 IDE 支持**：智能提示、自动完成、重构工具
- ✅ **编译时错误检测**：减少运行时错误
- ✅ **类型安全**：避免类型相关的 bug
- ✅ **代码可维护性**：类型即文档

### 代码质量提升

- ✅ **更少的 bug**：类型检查捕获常见错误
- ✅ **更易重构**：类型系统保护重构过程
- ✅ **更好的协作**：类型定义作为接口文档
- ✅ **更高的测试覆盖率**：类型驱动测试编写

---

## 📚 参考资源

- [Vue 3 TypeScript 文档](https://vuejs.org/guide/typescript/overview.html)
- [Vite TypeScript 指南](https://vitejs.dev/guide/features.html#typescript)
- [Pinia TypeScript 支持](https://pinia.vuejs.org/core-concepts/#typescript)
- [Vue Router TypeScript](https://router.vuejs.org/guide/advanced/typing.html)
- [Element Plus TypeScript](https://element-plus.org/en-US/guide/typescript.html)

---

**创建时间**: 2026-03-19  
**预计完成时间**: 16-24 天  
**优先级**: 中等  
**状态**: ⏳ 待开始
