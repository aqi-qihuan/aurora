# Aurora Admin V3 - TypeScript 迁移计划

## 📋 概述

本文档详细说明了将 Aurora Admin V3 从 JavaScript 迁移到 TypeScript 的完整计划，包括迁移策略、时间规划和实施步骤。

## 🎯 迁移目标

- ✅ 完整的 TypeScript 类型支持
- ✅ 更好的 IDE 智能提示和代码补全
- ✅ 编译时错误检测，减少运行时错误
- ✅ 增强的代码可维护性和可读性
- ✅ 保持现有功能 100% 可用

## 📊 项目现状

### 当前技术栈

```json
{
  "vue": "^3.4.21",
  "vite": "^5.1.5",
  "pinia": "^2.1.7",
  "vue-router": "^4.3.0",
  "element-plus": "^2.5.6",
  "vitest": "^1.6.0"
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

#### 任务清单

- [ ] 安装 TypeScript 依赖
  ```bash
  npm install -D typescript @types/node @vue/tsconfig vue-tsc
  ```

- [ ] 创建 `tsconfig.json`
  ```json
  {
    "compilerOptions": {
      "target": "ESNext",
      "module": "ESNext",
      "moduleResolution": "bundler",
      "strict": true,
      "jsx": "preserve",
      "baseUrl": ".",
      "paths": {
        "@/*": ["./src/*"]
      }
    },
    "include": ["src/**/*.ts", "src/**/*.vue"],
    "exclude": ["node_modules", "dist"]
  }
  ```

- [ ] 创建 `src/env.d.ts`
  ```typescript
  /// <reference types="vite/client" />
  
  declare module '*.vue' {
    import type { DefineComponent } from 'vue'
    const component: DefineComponent<{}, {}, any>
    export default component
  }
  ```

- [ ] 重命名配置文件
  - `vite.config.js` → `vite.config.ts`
  - `vitest.config.js` → `vitest.config.ts`

- [ ] 更新 `package.json` scripts
  ```json
  {
    "scripts": {
      "build": "vue-tsc && vite build",
      "type-check": "vue-tsc --noEmit"
    }
  }
  ```

**完成标志**：`npm run type-check` 运行无致命错误

---

### Phase 2: 类型定义准备（2-3 天）

#### 任务清单

- [ ] 创建类型定义目录结构
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
  └── store.ts              # Store 状态类型
  ```

- [ ] 定义核心 API 类型
  ```typescript
  // src/types/api.ts
  export interface ApiResponse<T = any> {
    code: number
    message: string
    data: T
  }
  
  export interface PageParams {
    pageNum?: number
    pageSize?: number
  }
  
  export interface PageData<T> {
    list: T[]
    total: number
  }
  ```

- [ ] 定义用户模块类型
  ```typescript
  // src/types/user.ts
  export interface UserInfo {
    id: number
    username: string
    nickname: string
    avatar: string
    email?: string
    loginType: LoginType
    roles: Role[]
    isDisable: 0 | 1
    createTime: string
  }
  
  export type LoginType = 1 | 2 | 3
  
  export interface LoginForm {
    username: string
    password: string
  }
  ```

- [ ] 定义路由类型
  ```typescript
  // src/types/router.ts
  import type { RouteRecordRaw } from 'vue-router'
  
  export interface AppRouteRecordRaw extends Omit<RouteRecordRaw, 'children'> {
    hidden?: boolean
    meta?: RouteMeta
    children?: AppRouteRecordRaw[]
  }
  
  export interface RouteMeta {
    title?: string
    icon?: string
    hidden?: boolean
    keepAlive?: boolean
  }
  ```

**完成标志**：所有核心业务类型定义完成

---

### Phase 3: 工具函数迁移（2-3 天）

#### 任务清单

- [ ] 迁移 `utils/request.js` → `utils/request.ts`
  ```typescript
  import axios, { type AxiosInstance, type AxiosResponse } from 'axios'
  
  const request: AxiosInstance = axios.create({
    baseURL: '/api',
    timeout: 10000
  })
  
  export default request
  ```

- [ ] 迁移 `utils/auth.js` → `utils/auth.ts`
  ```typescript
  export function getToken(): string | null {
    return sessionStorage.getItem('token')
  }
  
  export function setToken(token: string): void {
    sessionStorage.setItem('token', token)
  }
  ```

- [ ] 迁移 `utils/logger.js` → `utils/logger.ts`

**完成标志**：所有工具函数迁移完成，类型检查通过

---

### Phase 4: Pinia Store 迁移（3-4 天）

#### 任务清单

- [ ] 迁移 `stores/user.js` → `stores/user.ts`
  ```typescript
  import { defineStore } from 'pinia'
  import type { UserState } from '@/types/store'
  
  export const useUserStore = defineStore('user', {
    state: (): UserState => ({
      userInfo: null,
      token: null,
      isLoggedIn: false
    }),
    actions: {
      async login(loginForm: LoginForm): Promise<void> {
        // ...
      }
    }
  })
  ```

- [ ] 迁移 `stores/app.js` → `stores/app.ts`
- [ ] 迁移 `stores/permission.js` → `stores/permission.ts`
- [ ] 迁移 `stores/pageState.js` → `stores/pageState.ts`

**完成标志**：所有 Store 迁移完成，类型检查通过

---

### Phase 5: Vue 组件迁移（5-7 天）

#### 迁移策略

采用**渐进式迁移**，按优先级分组：

**Group A - 核心组件（高优先级）**
1. `App.vue`
2. `layout/index.vue`
3. `login/Login.vue`
4. `home/Home.vue`

**Group B - 功能页面（中优先级）**
5. `user/User.vue`
6. `article/Article.vue`
7. `category/Category.vue`
8. `tag/Tag.vue`
9. `comment/Comment.vue`
10. `resource/Resource.vue`

**Group C - 其他页面（低优先级）**
11. `role/Role.vue`
12. `menu/Menu.vue`
13. `website/Website.vue`
14. `album/Album.vue`
15. `quartz/Quartz.vue`

#### 组件迁移示例

```typescript
// Login.vue - JavaScript 版本
<script setup>
import { ref } from 'vue'
const loginForm = ref({
  username: '',
  password: ''
})
</script>

// Login.vue - TypeScript 版本
<script setup lang="ts">
import { ref, reactive } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import type { LoginForm } from '@/types/user'

const loginFormRef = ref<FormInstance>()
const loginForm = reactive<LoginForm>({
  username: '',
  password: ''
})

const loginRules: FormRules<LoginForm> = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
}
</script>
```

#### 组件迁移检查清单

每个组件迁移后需要检查：

- [ ] `<script setup lang="ts">` 已添加
- [ ] 所有 props 定义类型
- [ ] 所有 emits 定义类型
- [ ] ref/reactive 变量添加类型注解
- [ ] 函数参数和返回值添加类型
- [ ] 无 TypeScript 编译错误
- [ ] 功能测试通过

**完成标志**：所有组件迁移完成，功能测试通过

---

### Phase 6: 路由迁移（1-2 天）

#### 任务清单

- [ ] 迁移 `router/index.js` → `router/index.ts`
  ```typescript
  import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
  import type { AppRouteRecordRaw } from '@/types/router'
  
  export const constantRoutes: AppRouteRecordRaw[] = [
    {
      path: '/login',
      name: 'Login',
      component: () => import('@/views/login/Login.vue'),
      meta: { title: '登录', hidden: true }
    }
  ]
  
  const router = createRouter({
    history: createWebHistory(),
    routes: constantRoutes as RouteRecordRaw[]
  })
  
  export default router
  ```

- [ ] 迁移 `router/modules/dynamicRoutes.js` → `router/modules/dynamicRoutes.ts`

**完成标志**：路由迁移完成，动态路由功能正常

---

### Phase 7: 测试和验证（2-3 天）

#### 任务清单

- [ ] 运行类型检查
  ```bash
  npm run type-check
  ```

- [ ] 运行单元测试
  ```bash
  npm run test:run
  ```

- [ ] 运行构建
  ```bash
  npm run build
  ```

- [ ] 功能测试
  - [ ] 用户登录/登出
  - [ ] 路由跳转和权限控制
  - [ ] 主题切换
  - [ ] 核心业务功能（用户、文章、分类等）

**完成标志**：所有测试通过，功能正常

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

### 常见问题解决方案

#### 1. 第三方库类型

```bash
# 安装类型定义
npm install -D @types/xxx
```

#### 2. Vue 组件类型

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

#### 3. Pinia Store 类型

```typescript
// 使用泛型定义 state 类型
export const useStore = defineStore('store', {
  state: (): State => ({ ... })
})
```

---

## 🎯 预期收益

### 开发体验提升

- ✅ **更好的 IDE 支持**
  - 智能提示和自动完成
  - 代码重构工具
  - 类型导航

- ✅ **编译时错误检测**
  - 减少 90% 的类型相关 bug
  - 提前发现潜在问题

- ✅ **代码可维护性**
  - 类型即文档
  - 更清晰的接口定义
  - 更好的协作体验

### 代码质量提升

- ✅ **更少的 bug**
  - 类型检查捕获常见错误
  - 空值检查
  - 类型不匹配警告

- ✅ **更易重构**
  - 类型系统保护重构过程
  - 自动更新引用

- ✅ **更好的协作**
  - 类型定义作为接口文档
  - 降低沟通成本

---

## 📚 参考资源

- [Vue 3 TypeScript 文档](https://vuejs.org/guide/typescript/overview.html)
- [Vite TypeScript 指南](https://vitejs.dev/guide/features.html#typescript)
- [Pinia TypeScript 支持](https://pinia.vuejs.org/core-concepts/#typescript)
- [Vue Router TypeScript](https://router.vuejs.org/guide/advanced/typing.html)
- [Element Plus TypeScript](https://element-plus.org/en-US/guide/typescript.html)

---

## 📝 Git 分支策略

### 分支命名

```bash
# 从 v3.0.0-js 标签创建新分支
git checkout -b typescript v3.0.0-js

# 或从当前分支创建
git checkout -b typescript
```

### 提交策略

```bash
# Phase 1 完成
git commit -m "chore: setup TypeScript configuration"

# Phase 2 完成
git commit -m "feat: add type definitions for all modules"

# Phase 3 完成
git commit -m "refactor: migrate utils to TypeScript"

# 以此类推...
```

---

**创建时间**: 2026-03-19  
**预计完成时间**: 16-24 天  
**优先级**: 中等  
**状态**: ⏳ 待开始
