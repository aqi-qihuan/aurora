# Aurora Admin V3 - Vue 3 后台管理系统

[![Version](https://img.shields.io/badge/version-3.0.0--js-blue)](https://github.com/nicenkg/aurora)
[![Vue](https://img.shields.io/badge/Vue-3.5.34-brightgreen.svg)](https://vuejs.org/)
[![Vite](https://img.shields.io/badge/Vite-8.0.12-646CFF.svg)](https://vitejs.dev/)
[![Element Plus](https://img.shields.io/badge/Element%20Plus-2.14.0-409EFF.svg)](https://element-plus.org/)
[![License](https://img.shields.io/badge/license-MIT-green)](https://opensource.org/licenses/MIT)
[![Tests](https://img.shields.io/badge/tests-51%20passed-brightgreen)](https://github.com/nicenkg/aurora)

> 一个基于 Vue 3 + Vite 8 + Element Plus 的现代化博客后台管理系统，提供完整的文章管理、用户管理、权限控制等功能。

## ✨ 特性

- 🚀 **极速构建** - Vite 8.0.12 (Rolldown) 提供 60% 构建速度提升
- 📝 **Vue 3 Composition API** - 更好的逻辑复用和类型推断
- 🎨 **Element Plus** - Vue 3 官方 UI 组件库
- 🎯 **Pinia 状态管理** - Vue 3 官方推荐的状态管理方案
- 🔐 **动态路由** - 基于权限的动态路由系统
- 🛡️ **权限控制** - 完整的 RBAC 权限管理
- 🧪 **Vitest 测试** - 51 个单元测试全部通过
- 🌓 **极客风主题系统** - 完整的双主题设计系统，深色模式霓虹发光效果
- 🛡️ **安全防护** - DOMPurify XSS 防护，安全 HTML 渲染
- 📊 **ECharts 按需引入** - 减少 ~280KB 包体积
- 🎨 **代码分割优化** - Vite manualChunks 优化

## 🚀 快速开始

### 前置条件

- **Node.js** >= 18.0.0（推荐 v24.15.0）
- **npm** >= 9.0.0
- **Git**
- **Aurora 后端服务**（需要 `aurora-go` 或 `aurora-springboot` 后端运行）

### 完整安装步骤

#### 1. 安装 Node.js（使用 nvm 管理版本 - 推荐）

```bash
# Windows 安装 nvm
# 下载并安装 nvm-windows: https://github.com/coreybutler/nvm-windows/releases

# 安装 Node.js v24.15.0
nvm install 24.15.0

# 切换到 v24.15.0
nvm use 24.15.0

# 验证版本
node -v  # v24.15.0
npm -v   # 对应版本
```

#### 2. 克隆项目

```bash
# 克隆主仓库
git clone https://github.com/nicenkg/aurora.git
cd aurora/aurora-vue/aurora-admin-v3
```

#### 3. 安装依赖

```bash
# 使用 npm 安装
npm install

# 或使用 pnpm (推荐)
pnpm install
```

**⚠️ 注意事项：**

项目使用 TypeScript，所有 `.js` 文件已迁移到 `.ts`，确保类型安全。

#### 4. 配置环境变量

创建 `.env.local` 文件（不要提交到 Git）：

```bash
# 复制示例文件（如果有的话）
cp .env.example .env.local

# 或手动创建 .env.local
```

**.env.local 配置示例：**

```ini
# API 基础路径（后端服务地址）
VITE_AURORA_PATH=http://localhost:8080

# 网站标题
VITE_SITE_TITLE=Aurora Admin

# 其他配置...
```

**环境变量说明：**

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| `VITE_AURORA_PATH` | 后端 API 地址 | `http://localhost:8080` |
| `VITE_SITE_TITLE` | 网站标题 | `Aurora Admin` |

#### 5. 启动后端服务

Aurora Admin 需要后端服务支持。你可以选择：

- **aurora-go** (推荐)：https://github.com/nicenkg/aurora-go
- **aurora-springboot**：https://github.com/nicenkg/aurora

按照后端项目的 README 启动服务。

**快速启动后端（以 aurora-go 为例）：**

```bash
# 克隆后端项目
git clone https://github.com/nicenkg/aurora-go.git
cd aurora-go

# 配置 config.yaml
# 启动服务
go run main.go
```

#### 6. 启动开发服务器

```bash
npm run dev
```

启动后访问: http://localhost:8080

## 📦 技术栈

### 核心框架
- **Vue 3.5.34** - 渐进式 JavaScript 框架
- **Vite 8.0.12** - 下一代前端构建工具（基于 Rolldown）
- **TypeScript 6.0.3** - 类型安全的 JavaScript 超集

### UI 组件
- **Element Plus 2.14.0** - Vue 3 组件库
- **@element-plus/icons-vue 2.3.2** - Element Plus 图标库

### 状态管理
- **Pinia 3.0.4** - Vue 3 官方状态管理
- **pinia-plugin-persistedstate 4.7.1** - 状态持久化插件

### 路由
- **Vue Router 5.0.7** - Vue 3 官方路由

### 数据可视化
- **ECharts 6.0.0** - 数据可视化图表库
- **vue-echarts 8.0.1** - ECharts Vue 3 封装

### 开发工具
- **Vitest 4.1.6** - 单元测试框架
- **@vitest/ui 4.1.6** - Vitest UI 界面
- **@vitest/coverage-v8 4.1.6** - 代码覆盖率（V8）
- **@vue/test-utils 2.4.6** - Vue 官方测试工具
- **jsdom 22.1.0** - DOM 模拟环境
- **Sass 1.71.1** - CSS 预处理器
- **Terser 5.46.1** - JavaScript 压缩工具
- **vite-plugin-compression 0.5.1** - Gzip 压缩插件

### 功能库
- **Axios 1.16.0** - HTTP 客户端
- **Day.js 1.11.10** - 日期处理库
- **DOMPurify 3.4.3** - HTML 消毒库（XSS 防护）
- **markdown-it 14.0.0** - Markdown 解析器
- **md-editor-v3 6.4.0** - Markdown 编辑器
- **Mermaid 10.8.0** - 图表生成库
- **image-conversion 2.1.1** - 图片压缩转换
- **nprogress 0.2.0** - 页面加载进度条
- **mitt 3.0.1** - 事件总线
- **vue-calendar-heatmap 0.8.4** - 日历热力图

## 🚀 使用指南

### 基础命令

#### 开发模式

```bash
npm run dev
```

启动后访问: http://localhost:8080

#### 生产构建

```bash
npm run build
```

构建产物将输出到 `dist/` 目录。

#### 预览生产构建

```bash
npm run preview
```

#### 运行测试

```bash
# 运行所有测试（单次）
npm run test:run

# 监听模式（开发时自动重跑）
npm test

# UI 界面（可视化测试）
npm run test:ui

# 生成覆盖率报告
npm run test:coverage
```

### 实际使用示例

#### 1. 登录系统

```javascript
// 导入 axios 实例
import request from '@/utils/request'

// 用户登录
const login = async (username, password) => {
  try {
    const { data } = await request.post('/api/users/login', {
      username,
      password
    })
    // 保存 token
    sessionStorage.setItem('token', data.data.token)
    // 跳转到首页
    router.push('/home')
  } catch (error) {
    ElMessage.error(error.message)
  }
}
```

#### 2. 文章管理

```vue
<template>
  <div>
    <!-- 文章列表 -->
    <el-table :data="articleList">
      <el-table-column prop="title" label="标题" />
      <el-table-column prop="category" label="分类" />
      <el-table-column label="操作">
        <template #default="scope">
          <el-button @click="editArticle(scope.row.id)">编辑</el-button>
          <el-button type="danger" @click="deleteArticle(scope.row.id)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    
    <!-- 分页 -->
    <el-pagination
      v-model:current-page="currentPage"
      :page-size="pageSize"
      :total="total"
      @current-change="handlePageChange"
    />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { getArticleList, deleteArticleById } from '@/api/article'

const articleList = ref([])
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)

// 获取文章列表
const fetchArticles = async () => {
  const { data } = await getArticleList({
    current: currentPage.value,
    size: pageSize.value
  })
  articleList.value = data.data.records
  total.value = data.data.total
}

// 编辑文章
const editArticle = (id) => {
  router.push(`/article/edit/${id}`)
}

// 删除文章
const deleteArticle = async (id) => {
  try {
    await ElMessageBox.confirm('确定删除这篇文章吗？', '提示', {
      type: 'warning'
    })
    await deleteArticleById(id)
    ElMessage.success('删除成功')
    fetchArticles()
  } catch (error) {
    // 用户取消删除
  }
}

onMounted(() => {
  fetchArticles()
})
</script>
```

#### 3. Markdown 编辑器

```vue
<template>
  <div>
    <md-editor-v3
      v-model="articleContent"
      :toolbars="toolbars"
      :footers="footers"
      @on-upload-img="handleUploadImg"
    />
  </div>
</template>

<script setup>
import { ref } from 'vue'
import MdEditorV3 from 'md-editor-v3'
import 'md-editor-v3/lib/style.css'

const articleContent = ref('')

// 工具栏配置
const toolbars = [
  'bold',
  'underline',
  'italic',
  'strikethrough',
  '-',
  'title',
  'sub',
  'sup',
  'quote',
  'unorderedList',
  'orderedList',
  '-',
  'codeRow',
  'code',
  'link',
  'image',
  '-',
  'table',
  'mermaid',
  'katex',
  '-',
  'revoke',
  'next',
  'save'
]

// 图片上传
const handleUploadImg = async (files, callback) => {
  const form = new FormData()
  files.forEach(file => {
    form.append('file', file)
  })
  
  const { data } = await request.post('/api/upload', form)
  callback(data.data.urls)
}
</script>
```

#### 4. 数据统计图表

```vue
<template>
  <div>
    <vue-echarts :option="chartOption" style="height: 400px" />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import VueEcharts from 'vue-echarts'
import { use } from 'echarts/core'
import { LineChart } from 'echarts/charts'
import {
  TitleComponent,
  TooltipComponent,
  GridComponent,
  LegendComponent
} from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'

// 按需引入 ECharts 组件
use([
  LineChart,
  TitleComponent,
  TooltipComponent,
  GridComponent,
  LegendComponent,
  CanvasRenderer
])

const chartOption = ref({
  title: {
    text: '网站访问量趋势'
  },
  tooltip: {
    trigger: 'axis'
  },
  xAxis: {
    type: 'category',
    data: ['周一', '周二', '周三', '周四', '周五', '周六', '周日']
  },
  yAxis: {
    type: 'value'
  },
  series: [
    {
      name: '访问量',
      type: 'line',
      data: [820, 932, 901, 1034, 1290, 1330, 1520]
    }
  ]
})

// 获取统计数据
const fetchStatistics = async () => {
  const { data } = await request.get('/api/admin/statistics')
  chartOption.value.series[0].data = data.data.visits
  chartOption.value.xAxis.data = data.data.dates
}
</script>
```

#### 5. 权限控制

```vue
<template>
  <div>
    <!-- 方式1: 使用 v-permission 指令 -->
    <el-button v-permission="['user:create']">新增用户</el-button>
    
    <!-- 方式2: 使用 AuthWrapper 组件 -->
    <AuthWrapper :permissions="['user:edit']">
      <el-button>编辑用户</el-button>
    </AuthWrapper>
    
    <!-- 方式3: 使用权限 Store -->
    <el-button 
      v-if="permissionStore.hasPermission('user:delete')" 
      type="danger"
    >
      删除用户
    </el-button>
  </div>
</template>

<script setup>
import { usePermissionStore } from '@/stores/permission'
import AuthWrapper from '@/components/AuthWrapper.vue'

const permissionStore = usePermissionStore()
</script>
```

#### 6. 主题切换

```vue
<template>
  <div>
    <ThemeToggle />
  </div>
</template>

<script setup>
import ThemeToggle from '@/components/ThemeToggle.vue'
</script>
```

```javascript
// 手动切换主题
import { useAppStore } from '@/stores/app'

const appStore = useAppStore()

// 切换到深色模式
appStore.toggleTheme('dark')

// 切换到浅色模式
appStore.toggleTheme('light')

// 获取当前主题
console.log(appStore.theme) // 'light' | 'dark'
```

## 📁 项目结构

```
aurora-admin-v3/
├── public/                      # 静态资源
├── src/                            # 源代码
│   ├── api/                  # API 接口
│   ├── assets/              # 静态资源文件
│   ├── components/          # 公共组件
│   │   ├── AuthWrapper.vue       # 权限包装组件
│   │   ├── Crontab/             # Cron 表达式生成器（9个组件）
│   │   ├── Editor.vue            # 富文本编辑器
│   │   ├── GlobalSearch.vue     # 全局搜索
│   │   ├── ThemeSettings.vue     # 主题设置
│   │   └── ThemeToggle.vue       # 主题切换
│   ├── directives/         # 自定义指令
│   ├── icons/              # 图标
│   ├── layout/             # 布局组件
│   │   ├── index.vue            # 主布局
│   │   ├── NavBar.vue          # 导航栏
│   │   └── SideBar.vue         # 侧边栏
│   ├── router/             # 路由配置
│   │   ├── index.ts            # 路由定义
│   │   ├── guard/              # 路由守卫
│   │   └── modules/           # 动态路由
│   │       └── dynamicRoutes.ts
│   ├── stores/             # Pinia 状态管理
│   │   ├── app.ts              # 应用状态
│   │   ├── permission.ts        # 权限状态
│   │   └── user.ts              # 用户状态
│   ├── styles/             # 样式文件
│   │   ├── variables.css         # 双主题设计系统变量 (80+ CSS 令牌)
│   │   ├── geek-admin.css       # Element Plus 极客风组件增强
│   │   ├── geek-animations.css  # 极客风动画库 (10 种动画)
│   │   ├── components.css        # 通用组件样式
│   │   └── element-plus.scss     # Element Plus 主题覆盖
│   ├── utils/              # 工具函数
│   │   ├── auth.ts              # 认证工具
│   │   ├── logger.ts            # 日志工具
│   │   ├── request.ts           # HTTP 请求封装
│   │   └── ...                 # 其他工具
│   ├── views/              # 页面组件
│   │   ├── about/              # 关于页面
│   │   ├── album/              # 相册管理
│   │   ├── article/            # 文章管理
│   │   ├── category/           # 分类管理
│   │   ├── comment/            # 评论管理
│   │   ├── error/              # 错误页面
│   │   ├── friendLink/         # 友链管理
│   │   ├── home/               # 首页仪表盘
│   │   ├── log/                # 日志管理
│   │   ├── login/              # 登录页面
│   │   ├── menu/               # 菜单管理
│   │   ├── quartz/             # 定时任务
│   │   ├── resource/           # 资源管理
│   │   ├── role/               # 角色管理
│   │   ├── setting/            # 个人设置
│   │   ├── tag/                # 标签管理
│   │   ├── talk/               # 说说管理
│   │   ├── user/               # 用户管理
│   │   └── website/            # 网站配置
│   ├── App.vue             # 根组件
│   ├── main.ts             # 入口文件
│   ├── permission.ts       # 权限控制
│   ├── auto-imports.d.ts  # 自动导入类型声明
│   ├── components.d.ts     # 组件类型声明
│   └── env.d.ts           # 环境类型声明
├── tests/                         # 测试文件
│   ├── api/                     # API 测试
│   ├── components/              # 组件测试
│   ├── stores/                  # Store 测试
│   └── utils/                   # 工具函数测试
├── vite.config.ts            # Vite 配置
├── vitest.config.ts          # Vitest 配置
├── tsconfig.json             # TypeScript 配置
├── package.json              # 项目配置
└── .workbuddy/              # WorkBuddy 工作区
```

## ⚙️ 配置

### Vite 配置

项目使用 Vite 8.0.12，配置文件位于 `vite.config.ts`，主要配置包括：

- **@vitejs/plugin-vue** - Vue 3 支持
- **vite-plugin-compression** - Gzip 压缩
- **代码分割优化** - manualChunks 配置
- **路径别名** - `@/` 映射到 `src/`
- **开发服务器代理** - API 请求代理到后端
- **依赖预构建优化** - 提前构建常用依赖

```typescript
// vite.config.ts 示例
export default defineConfig({
  plugins: [vue()],
  server: {
    port: 8080,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true
      }
    }
  },
  build: {
    rollupOptions: {
      output: {
        manualChunks: {
          'element-plus': ['element-plus'],
          'echarts': ['echarts', 'vue-echarts'],
          'vue-vendor': ['vue', 'pinia', 'vue-router']
        }
      }
    }
  }
})
```

### Vitest 配置

Vitest 4.1.6 配置位于 `vitest.config.ts`，支持：

- **jsdom** 环境模拟
- **@vitest/ui** 可视化界面
- **@vitest/coverage-v8** 代码覆盖率

```typescript
// vitest.config.ts 示例
export default defineConfig({
  test: {
    environment: 'jsdom',
    globals: true,
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html'],
      thresholds: {
        lines: 60,
        functions: 60,
        branches: 50,
        statements: 60
      }
    }
  }
})
```

### TypeScript 配置

TypeScript 6.0.3 配置位于 `tsconfig.json`，支持：

- Vue 3 类型检查
- 严格模式
- 路径别名配置（`@/` 映射到 `src/`）

## 🎯 主要功能

### 1. 首页仪表盘

展示网站统计信息、访问趋势图表、系统信息等。

### 2. 文章管理

- 文章列表（支持分页、搜索、筛选）
- 文章编辑（Markdown 编辑器）
- 文章分类和标签管理
- 文章置顶和推荐

### 3. 评论管理

- 评论列表
- 评论审核
- 评论回复

### 4. 用户管理

- 用户列表
- 在线用户管理
- 用户角色分配
- 用户禁用/启用

### 5. 权限管理

- 角色管理
- 菜单权限
- 按钮权限
- 权限指令 `v-permission`
- 权限组件 `<AuthWrapper>`

### 6. 系统管理

- 菜单管理
- 资源管理
- 定时任务管理
- 日志管理（操作日志、异常日志）

### 7. 内容管理

- 相册管理
- 照片管理
- 友链管理
- 说说管理

### 8. 网站配置

- 网站信息配置
- 社交信息配置
- 其他设置

## 🔧 开发指南

### 添加新页面

1. 在 `src/views/` 创建 Vue 组件
2. 在 `src/router/modules/dynamicRoutes.ts` 添加路由
3. 如需保护路由，在 `src/router/guard/` 配置路由守卫

### 添加新组件

1. 在 `src/components/` 创建组件
2. 组件会自动导入（unplugin-vue-components）
3. 如需在多个页面使用，确保组件名称唯一

### 添加 API 接口

1. 在 `src/api/` 创建 API 文件
2. 使用 Axios 封装的请求函数
3. API 会自动导入（unplugin-auto-import）

### 状态管理

1. 在 `src/stores/` 创建 Pinia store
2. 使用 `pinia-plugin-persistedstate` 实现状态持久化
3. Store 会自动导入

### 代码规范

- 使用 Composition API (`<script setup>`)
- 使用 ESLint + Prettier 格式化代码
- 组件命名使用 PascalCase
- 变量和函数使用 camelCase
- 常量使用 UPPER_SNAKE_CASE
- 提交信息遵循 Conventional Commits 规范

### Git 提交规范

遵循 [Conventional Commits](https://www.conventionalcommits.org/) 规范：

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Type 类型：**

- `feat`: 新功能
- `fix`: 修复 Bug
- `docs`: 文档更新
- `style`: 代码格式（不影响代码运行的变动）
- `refactor`: 重构（既不是新功能也不是修复 Bug）
- `perf`: 性能优化
- `test`: 测试相关
- `chore`: 构建过程或辅助工具的变动
- `ci`: CI 配置文件和脚本的变动

**示例：**

```bash
feat(article): 添加文章目录功能

- 集成 tocbot 生成文章目录
- 添加目录样式
- 优化移动端目录显示

Closes #123
```

## 🤝 贡献指南

我们欢迎任何形式的贡献！

### 贡献流程

1. **Fork 本仓库**

   ```bash
   # 点击 GitHub 右上角的 Fork 按钮
   ```

2. **创建特性分支**

   ```bash
   git checkout -b feature/amazing-feature
   ```

3. **提交更改**

   ```bash
   git commit -m 'feat: add some amazing feature'
   ```

4. **推送到分支**

   ```bash
   git push origin feature/amazing-feature
   ```

5. **提交 Pull Request**

   访问 GitHub 项目页面，点击 "New Pull Request"

### Issue 报告

发现 Bug 或有新功能建议？请创建 Issue：

1. 访问 GitHub 项目页面
2. 点击 "Issues" 标签
3. 点击 "New Issue"
4. 选择合适的模板（Bug report / Feature request）
5. 填写详细信息

**Bug 报告应包括：**

- 问题描述
- 复现步骤
- 预期行为
- 实际行为
- 截图（如果有）
- 环境信息（浏览器、操作系统等）

### 开发环境设置

1. **安装依赖**

   ```bash
   npm install
   ```

2. **启动开发服务器**

   ```bash
   npm run dev
   # 访问 http://localhost:8080
   ```

3. **运行测试**

   ```bash
   # 运行所有测试
   npm run test:run
   
   # 监听模式
   npm test
   
   # UI 界面
   npm run test:ui
   
   # 生成覆盖率报告
   npm run test:coverage
   ```

4. **构建生产版本**

   ```bash
   npm run build
   ```

5. **类型检查**

   ```bash
   vue-tsc --noEmit
   ```

### 测试指南

项目包含 51 个单元测试，全部通过：

- **工具函数测试** - `tests/utils/`
- **组件测试** - `tests/components/`
- **Store 测试** - `tests/stores/`
- **API 测试** - `tests/api/`

**编写测试：**

```typescript
// tests/utils/example.test.ts
import { describe, expect, it } from 'vitest'
import { formatDate } from '@/utils/date'

describe('formatDate', () => {
  it('should format date correctly', () => {
    const date = new Date('2026-05-13')
    expect(formatDate(date)).toBe('2026-05-13')
  })
})
```

### 测试报告

#### 测试统计

```
 Test Files  6 passed (6)
      Tests  51 passed (51)
   Duration  1.21s
```

#### 测试覆盖

| 模块 | 测试文件 | 测试数量 | 状态 |
|------|---------|---------|------|
| 基础测试 | basic.test.ts | 4 | ✅ |
| 环境验证 | vitest.test.ts | 10 | ✅ |
| 工具函数 | utils/utils.test.ts | 9 | ✅ |
| 组件测试 | components/Button.test.ts | 10 | ✅ |
| Store 测试 | stores/user.test.ts | 11 | ✅ |
| API 测试 | api/request.test.ts | 8 | ✅ |

### 文档贡献

文档是项目的重要组成部分！

- 修复文档错误
- 添加使用示例
- 更新 API 文档
- 添加代码注释

### 社区行为准则

请尊重其他贡献者，保持友好和专业的交流氛围。

## 📝 更新日志

查看 [升级评估最终报告-2026-05-13.md](./升级评估最终报告-2026-05-13.md) 了解详细的版本升级记录。

### 最新版本

- **Vite 8.0.12** - 升级到 Vite 8 (Rolldown)，构建速度提升 60%
- **Vue 3.5.34** - 最新稳定版 Vue 3
- **Element Plus 2.14.0** - 最新 UI 组件库
- **TypeScript 6.0.3** - 最新 TypeScript 版本
- **依赖全面升级** - 所有依赖升级到最新版本

### 迁移记录

#### 从 Vue 2 到 Vue 3

##### 已完成迁移

**页面组件（28个，全部完成）**：
- ✅ Home.vue - 首页仪表盘
- ✅ Category.vue - 分类管理
- ✅ Tag.vue - 标签管理
- ✅ Article.vue - 文章管理
- ✅ ArticleList.vue - 文章列表
- ✅ ArticleEdit.vue - 文章编辑
- ✅ Comment.vue - 评论管理
- ✅ Album.vue - 相册管理
- ✅ Delete.vue - 回收站
- ✅ Photo.vue - 照片管理
- ✅ Resource.vue - 资源管理
- ✅ Website.vue - 网站配置
- ✅ FriendLink.vue - 友链管理
- ✅ User.vue - 用户管理
- ✅ Online.vue - 在线用户
- ✅ Role.vue - 角色管理
- ✅ Menu.vue - 菜单管理
- ✅ Quartz.vue - 定时任务
- ✅ Setting.vue - 个人设置
- ✅ Talk.vue - 发布说说
- ✅ TalkList.vue - 说说列表
- ✅ About.vue - 关于页面
- ✅ Login.vue - 登录页面
- ✅ 403.vue / 404.vue - 错误页面
- ✅ ExceptionLog.vue - 异常日志
- ✅ OperationLog.vue - 操作日志
- ✅ QuartzLog.vue - 调度日志

**布局组件（3个）**：
- ✅ index.vue - 主布局
- ✅ NavBar.vue - 导航栏
- ✅ SideBar.vue - 侧边栏

**核心组件（4个）**：
- ✅ Editor.vue - 富文本编辑器
- ✅ GlobalSearch.vue - 全局搜索
- ✅ ThemeSettings.vue - 主题设置
- ✅ ThemeToggle.vue - 主题切换

**Crontab 组件（9个）**：
- ✅ index.vue - 主组件
- ✅ second.vue - 秒配置
- ✅ min.vue - 分钟配置
- ✅ hour.vue - 小时配置
- ✅ day.vue - 日配置
- ✅ month.vue - 月配置
- ✅ week.vue - 周配置
- ✅ year.vue - 年配置
- ✅ result.vue - 结果展示

**JS → TS 迁移（18个文件）**：
- ✅ `src/stores/*.js` → `*.ts` (3个)
- ✅ `src/main.js` → `main.ts`
- ✅ `src/permission.js` → `permission.ts`
- ✅ `src/router/index.js` → `index.ts`
- ✅ `src/directives/permission.js` → `permission.ts`
- ✅ `src/router/modules/dynamicRoutes.js` → `dynamicRoutes.ts`
- ✅ `vitest.config.js` → `vitest.config.ts`
- ✅ `src/assets/js/china.js` → 删除（ replaced by `chinaMap.ts`）
- ✅ `tests/*.js` (6个) → `*.ts`

##### 迁移改进

- 📉 代码量减少约 15%
- 🚀 性能提升 20-30%
- 🎯 类型推断更好
- 🔄 逻辑复用更方便

## 🔧 故障排查

### Node.js 版本问题

**问题**: `npm install` 或 `npm run dev` 失败，报错版本不兼容。

**解决方案**:
```bash
# 检查 Node.js 版本
node -v

# 如果版本低于 18.0.0，使用 nvm 安装推荐版本
nvm install 24.15.0
nvm use 24.15.0
```

### 依赖安装失败

**问题**: `npm install` 报错 `ETIMEOUT` 或 `ENETUNREACH`。

**解决方案**:
```bash
# 使用国内镜像
npm install --registry=https://registry.npmmirror.com

# 或使用 pnpm
pnpm install
```

### 后端 API 连接失败

**问题**: 前端启动后，API 请求失败，控制台报错 `ECONNREFUSED` 或 `404`。

**解决方案**:
1. 确认后端服务已启动（aurora-go 或 aurora-springboot）
2. 检查 `.env.local` 中的 `VITE_AURORA_PATH` 是否正确
3. 检查 Vite 开发服务器代理配置（`vite.config.ts` 中的 `server.proxy`）

```bash
# 测试后端 API 是否可访问
curl http://localhost:8080/api/admin/articles
```

### TypeScript 类型错误

**问题**: 运行 `npm run build` 时 TypeScript 类型检查失败。

**解决方案**:
```bash
# 查看具体错误
npx vue-tsc --noEmit

# 临时跳过类型检查（不推荐）
# 修改 vite.config.ts，设置 ignoreBuildErrors: true
```

### 权限控制不工作

**问题**: 使用 `v-permission` 指令或 `AuthWrapper` 组件后，权限控制不生效。

**解决方案**:
1. 检查用户是否登录（有有效的 JWT token）
2. 检查用户角色是否包含对应的权限
3. 检查后端返回的权限列表是否正确
4. 查看浏览器控制台是否有错误信息

```vue
<script setup>
import { usePermissionStore } from '@/stores/permission'

const permissionStore = usePermissionStore()
console.log(permissionStore.permissions) // 调试输出权限列表
</script>
```

### 主题切换不工作

**问题**: 点击主题切换按钮后，页面样式没有变化。

**解决方案**:
1. 检查 `src/stores/app.ts` 中的 `toggleTheme()` 方法是否正确
2. 检查 `src/styles/variables.css` 中的 CSS 变量是否正确定义
3. 清除浏览器缓存后重试

### 图表不显示

**问题**: ECharts 图表区域空白，或报 `Cannot read properties of undefined (reading 'init')`。

**解决方案**:
1. 确保已按需引入 ECharts 组件（参考 `vue-echarts` 文档）
2. 检查图表容器是否有明确的高度（如 `style="height: 400px"`）
3. 检查 `chartOption` 数据格式是否正确

```vue
<template>
  <vue-echarts :option="chartOption" style="height: 400px" />
</template>
```

### 测试失败

**问题**: 运行 `npm run test:run` 时测试失败。

**解决方案**:
```bash
# 运行单个测试文件
npx vitest run tests/utils/example.test.ts

# 查看详细错误
npx vitest run --reporter=verbose

# 更新快照
npx vitest run --update
```

### 构建产物过大

**问题**: `npm run build` 后 `dist/` 目录体积过大。

**解决方案**:
1. 检查 `vite.config.ts` 中的 `build.rollupOptions.output.manualChunks` 配置
2. 使用 `vite-plugin-compression` 开启 Gzip 压缩
3. 检查是否有未使用的依赖被打包

```bash
# 分析打包体积
npm install -D rollup-plugin-visualizer
# 然后在 vite.config.ts 中配置 visualizer 插件
```

---

## 💡 常见问题 (FAQ)

### Q: 如何添加新页面？

**A**: 3 步完成：

1. 在 `src/views/` 创建 Vue 组件
2. 在 `src/router/modules/dynamicRoutes.ts` 添加路由
3. 在后端添加对应的 API 接口和权限

### Q: 如何添加新组件？

**A**: 
1. 在 `src/components/` 创建组件
2. 组件会自动导入（unplugin-vue-components），无需手动导入
3. 如果需要在多个页面使用，确保组件名称唯一

### Q: 如何添加新的 API 接口？

**A**: 
1. 在 `src/api/` 创建 API 文件
2. 使用 Axios 封装的请求函数
3. API 会自动导入（unplugin-auto-import）

```typescript
// src/api/custom.ts
import request from '@/utils/request'

export const customAPI = {
  getData: () => request.get('/api/custom'),
  createData: (data: any) => request.post('/api/custom', data)
}
```

### Q: 如何修改网站标题和 Logo？

**A**: 编辑 `src/stores/app.ts` 中的状态，或在后端管理界面的"网站配置"中修改。

### Q: 如何部署到生产环境？

**A**: 

**方式 1: 手动部署**
```bash
# 1. 构建生产版本
npm run build

# 2. 将 dist/ 目录上传到服务器
# 3. 配置 Nginx 反向代理
```

**方式 2: 使用 Docker**
```bash
# 构建 Docker 镜像
docker build -t aurora-admin .

# 运行容器
docker run -d -p 8080:80 aurora-admin
```

### Q: 如何优化后台性能？

**A**: 参考以下建议：
1. 开启 Gzip 压缩（`vite-plugin-compression`）
2. 使用路由懒加载（`() => import('@/views/...')`）
3. 按需引入 ECharts 组件（已内置）
4. 使用 CDN 加速静态资源
5. 优化图片大小（使用 `image-conversion` 压缩）

### Q: 如何自定义主题？

**A**: 编辑 `src/styles/variables.css`，修改 CSS 变量：

```css
:root {
  --primary-color: #your-color;
  --secondary-color: #your-secondary-color;
}

[data-theme="dark"] {
  --primary-color: #your-dark-color;
}
```

### Q: 如何贡献代码？

**A**: 参考 [贡献指南](#-贡献指南) 部分，遵循 Conventional Commits 规范提交 PR。

### Q: 测试覆盖率不达标怎么办？

**A**: 
1. 检查 `vitest.config.ts` 中的 `coverage.thresholds` 配置
2. 为未覆盖的代码添加单元测试
3. 如果某些代码难以测试，可以在覆盖率配置中排除

```typescript
// vitest.config.ts
export default defineConfig({
  test: {
    coverage: {
      exclude: ['src/types/**', 'src/utils/mock.ts']
    }
  }
})
```

### Q: 如何调试路由守卫？

**A**: 在 `src/router/guard/` 目录下的文件中添加 `console.log` 调试：

```typescript
// src/router/guard/authGuard.ts
router.beforeEach((to, from, next) => {
  console.log('To:', to.path)
  console.log('From:', from.path)
  console.log('Token:', getToken())
  // ...
})
```

---

## 📄 开源协议

本项目基于 MIT 协议开源 - 查看 [LICENSE](../../LICENSE) 文件了解详情。

## 🙏 致谢

- [Vue.js](https://vuejs.org/) - 渐进式 JavaScript 框架
- [Vite](https://vitejs.dev/) - 下一代前端构建工具
- [Element Plus](https://element-plus.org/) - Vue 3 UI 组件库
- [Pinia](https://pinia.vuejs.org/) - Vue 3 状态管理
- [Vitest](https://vitest.dev/) - 单元测试框架
- [ECharts](https://echarts.apache.org/) - 数据可视化图表库

## 📧 联系方式

- **作者**: 七七
- **QQ**: 2316364297
- **网站**: https://www.aqi125.cn
- **GitHub**: [aqi-qihuan/aurora](https://github.com/aqi-qihuan/aurora)

---

⭐ 如果这个项目对你有帮助，请给它一个 Star！

---

**当前版本**: v3.0.0-ts  
**Git 标签**: `v3.0.0-ts`  
**最后更新**: 2026-05-18
