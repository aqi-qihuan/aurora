# Aurora Blog - 现代化个人博客系统

[![Vue](https://img.shields.io/badge/Vue-3.5.34-brightgreen.svg)](https://vuejs.org/)
[![Vite](https://img.shields.io/badge/Vite-8.0.12-646CFF.svg)](https://vitejs.dev/)
[![Element Plus](https://img.shields.io/badge/Element%20Plus-2.14.0-409EFF.svg)](https://element-plus.org/)
[![TypeScript](https://img.shields.io/badge/TypeScript-6.0.3-3178C6.svg)](https://www.typescriptlang.org/)
[![Tailwind CSS](https://img.shields.io/badge/Tailwind%20CSS-3.4.19-06B6D4.svg)](https://tailwindcss.com/)

Aurora Blog 是一个基于 Vue 3 + Vite 8 + Element Plus 的现代化个人博客系统，支持 Markdown 渲染、国际化、代码高亮、数学公式等丰富功能。

## ✨ 特性

- 🚀 **极速构建** - Vite 8.0.12 (Rolldown) 提供 60% 构建速度提升
- 📝 **Markdown 增强** - 支持 Mermaid 图表、KaTeX 数学公式、Emoji、脚注等
- 🎨 **现代化 UI** - Element Plus 组件库 + Tailwind CSS 样式
- 🌍 **国际化** - 内置 Vue I18n 多语言支持
- 💾 **状态持久化** - Pinia 状态管理 + 持久化插件
- 🖼️ **图片预览** - 支持图片懒加载和全屏预览
- 📱 **响应式设计** - 完美适配桌面端和移动端
- 🔍 **无限滚动** - 文章列表无限滚动加载
- 💬 **评论系统** - 完整的文章评论和回复功能
- 🏷️ **标签归档** - 文章分类、标签、归档功能

## 🛠️ 技术栈

### 核心框架
- **Vue 3.5.34** - 渐进式 JavaScript 框架
- **Vite 8.0.12** - 下一代前端构建工具（基于 Rolldown）
- **TypeScript 6.0.3** - 类型安全的 JavaScript 超集

### UI 组件
- **Element Plus 2.14.0** - Vue 3 组件库
- **Tailwind CSS 3.4.19** - 实用优先的 CSS 框架
- **@element-plus/icons-vue 2.3.2** - Element Plus 图标集

### 状态管理
- **Pinia 3.0.4** - Vue 3 官方状态管理
- **pinia-plugin-persistedstate 4.7.1** - 状态持久化插件

### 路由
- **Vue Router 5.0.6** - Vue 3 官方路由

### Markdown 支持
- **markdown-it 14.1.1** - Markdown 解析器
- **markdown-it-mermaid 1.1.0** - Mermaid 图表支持
- **markdown-it-katex 4.0.1** - KaTeX 数学公式
- **markdown-it-emoji 3.0.0** - Emoji 表情
- **markdown-it-container 4.0.0** - 自定义容器
- **markdown-it-footnote 4.0.0** - 脚注
- **markdown-it-abbr 2.0.0** - 缩写
- **markdown-it-ins 4.0.0** - 插入线
- **markdown-it-mark 4.0.0** - 标记
- **markdown-it-sub 2.0.0** - 下标
- **markdown-it-sup 2.0.0** - 上标
- **mavon-editor 3.0.1** - Markdown 编辑器

### 代码高亮
- **Prism.js 1.30.0** - 代码语法高亮
- **vite-plugin-prismjs 0.0.11** - Vite Prism 集成

### 其他功能
- **Axios 1.15.1** - HTTP 客户端
- **vue-i18n 11.4.2** - 国际化
- **tocbot 4.36.6** - 文章目录生成
- **nprogress 0.2.0** - 页面加载进度条
- **vue3-img-preview 1.1.16** - 图片预览
- **vue-avatar-cropper 6.1.1** - 头像裁剪
- **vue3-lazy 1.0.0-alpha.1** - 图片懒加载
- **vue3-infinite-scroll-better 2.2.0** - 无限滚动
- **vue3-click-away 1.2.4** - 点击外部指令

## 📦 安装

### 环境要求

- **Node.js** >= 22.0.0（推荐 v24.15.0）
- **npm** >= 9.0.0

### 使用 nvm 管理 Node.js 版本（推荐）

```bash
# 安装 nvm (Windows)
# 下载并安装 nvm-windows: https://github.com/coreybutler/nvm-windows/releases

# 安装 Node.js v24.15.0
nvm install 24.15.0

# 切换到 v24.15.0
nvm use 24.15.0

# 验证版本
node -v  # v24.15.0
npm -v   # 对应版本
```

### 克隆项目

```bash
git clone https://github.com/nicepkg/aurora.git
cd aurora/aurora-vue/aurora-blog
```

### 安装依赖

```bash
# 使用 npm 安装
npm install

# 或使用 pnpm (推荐)
pnpm install
```

#### 注意事项

1. **Postinstall 脚本**: 项目包含 `patch-cytoscape.js` 脚本，会在 `npm install` 后自动运行，用于修复 Cytoscape.js 在 Vite 8 中的兼容性问题。

2. **可选依赖**: `@rolldown/binding-win32-x64-msvc` 是 Vite 8 (Rolldown) 的 native binding，Windows 系统会自动安装。

## 🚀 使用

### 开发模式

```bash
npm run dev
```

启动后访问: http://localhost:5173

### 生产构建

```bash
npm run build
```

构建产物将输出到 `dist/` 目录。

### 预览生产构建

```bash
npm run preview
```

### 部署

```bash
npm run deploy
```

使用 `deploy.js` 脚本进行自动化部署。

## 📁 项目结构

```
aurora-blog/
├── public/                      # 静态资源
│   ├── favicon.ico            # 网站图标
│   ├── aurora-logo.html      # Logo 页面
│   └── http-client.env.json  # HTTP 客户端环境配置
├── src/                       # 源代码
│   ├── api/                  # API 接口
│   ├── assets/              # 项目资源文件
│   ├── components/          # 通用组件
│   │   ├── ArticleCard/    # 文章卡片组件
│   │   ├── Comment/        # 评论组件
│   │   ├── Feature/        # 功能特性组件
│   │   └── ...            # 其他组件
│   ├── config/             # 配置文件
│   ├── icons/              # SVG 图标
│   ├── locales/            # 国际化文件
│   ├── plugins/            # 插件配置
│   ├── router/             # 路由配置
│   │   ├── guard/         # 路由守卫
│   │   └── index.ts       # 路由定义
│   ├── stores/             # Pinia 状态管理
│   ├── styles/             # 全局样式
│   ├── utils/              # 工具函数
│   ├── views/              # 页面视图
│   │   ├── Home.vue       # 首页
│   │   ├── Article.vue    # 文章详情
│   │   ├── ArticleList.vue # 文章列表
│   │   ├── Archives.vue   # 归档页面
│   │   ├── Tags.vue       # 标签页面
│   │   ├── About.vue      # 关于页面
│   │   ├── FriendLink.vue # 友链页面
│   │   ├── Photos.vue     # 相册页面
│   │   ├── Talk.vue       # 说说页面
│   │   ├── TalkList.vue   # 说说列表
│   │   ├── Message.vue    # 留言板
│   │   └── 404.vue       # 404 页面
│   ├── App.vue             # 根组件
│   ├── main.ts             # 入口文件
│   ├── auto-imports.d.ts  # 自动导入类型声明
│   ├── components.d.ts     # 组件类型声明
│   ├── env.d.ts           # 环境类型声明
│   └── shims-vue.d.ts    # Vue 类型声明
├── index.html               # HTML 模板
├── package.json             # 项目配置
├── vite.config.ts          # Vite 配置
├── tsconfig.json           # TypeScript 配置
├── tailwind.config.cjs     # Tailwind CSS 配置
├── postcss.config.cjs      # PostCSS 配置
├── patch-cytoscape.js      # Cytoscape 补丁脚本
├── deploy.js               # 部署脚本
└── .workbuddy/            # WorkBuddy 工作区
```

## ⚙️ 配置

### Vite 配置

项目使用 Vite 8.0.12，配置文件位于 `vite.config.ts`，主要配置包括：

- **@vitejs/plugin-vue** - Vue 3 支持
- **vite-plugin-prismjs** - Prism.js 代码高亮
- **vite-plugin-svg-icons** - SVG 图标集成
- **unplugin-auto-import** - API 自动导入
- **unplugin-vue-components** - 组件自动导入

### TypeScript 配置

TypeScript 6.0.3 配置位于 `tsconfig.json`，支持：

- Vue 3 类型检查
- 严格模式
- 路径别名配置

### Tailwind CSS 配置

Tailwind CSS 3.4.19 配置位于 `tailwind.config.cjs`，可自定义主题、颜色和字体。

### 国际化配置

国际化文件位于 `src/locales/`，支持多语言切换。

## 🎯 主要功能

### 1. 首页

展示博客统计信息、文章列表、网站配置等。

### 2. 文章系统

- 文章列表（支持分页和无限滚动）
- 文章详情（Markdown 渲染、代码高亮、目录导航）
- 文章归档（按日期分类）
- 标签系统

### 3. Markdown 支持

- **Mermaid 图表** - 流程图、时序图、类图等
- **KaTeX 数学公式** - 行内和块级公式
- **Emoji 表情** - 丰富的 Emoji 支持
- **代码高亮** - Prism.js 语法高亮
- **自定义容器** - Info、Warning、Danger 等
- **脚注** - 文章脚注支持

### 4. 评论系统

- 评论列表
- 回复功能
- 评论点赞

### 5. 相册系统

- 相册列表
- 照片预览

### 6. 说说系统

- 说说列表
- 说说详情

### 7. 其他功能

- 友链页面
- 关于页面
- 留言板
- 搜索功能

## 🔧 开发指南

### 添加新页面

1. 在 `src/views/` 创建 Vue 组件
2. 在 `src/router/index.ts` 添加路由
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

## 🤝 贡献指南

我们欢迎任何形式的贡献！

### 贡献流程

1. **Fork 项目**

   ```bash
   # 点击 GitHub 右上角的 Fork 按钮
   ```

2. **创建特性分支**

   ```bash
   git checkout -b feature/AmazingFeature
   ```

3. **提交更改**

   ```bash
   git commit -m 'Add some AmazingFeature'
   ```

4. **推送到分支**

   ```bash
   git push origin feature/AmazingFeature
   ```

5. **打开 Pull Request**

   访问 GitHub 项目页面，点击 "New Pull Request"

### 代码规范

- 使用 TypeScript 编写代码
- 遵循 Vue 3 风格指南
- 使用 Prettier 格式化代码（配置文件：`.prettierrc`）
- 组件名称使用 PascalCase
- 变量和函数使用 camelCase
- 常量使用 UPPER_SNAKE_CASE

### 提交规范

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
- 环境信息（浏览器、操作系统、Node.js 版本等）

### 开发环境设置

1. **安装依赖**

   ```bash
   npm install
   ```

2. **启动开发服务器**

   ```bash
   npm run dev
   ```

3. **运行类型检查**

   ```bash
   vue-tsc --noEmit
   ```

4. **构建生产版本**

   ```bash
   npm run build
   ```

### 测试指南

目前项目尚未集成自动化测试，计划中的测试框架：

- **Vitest** - 单元测试
- **@vue/test-utils** - 组件测试
- **Playwright** - E2E 测试

### 文档贡献

文档是项目的重要组成部分！

- 修复文档错误
- 添加使用示例
- 翻译文档
- 添加 API 文档

### 社区行为准则

请阅读并遵守我们的 [行为准则](CODE_OF_CONDUCT.md)（如果有的话）。

## 📝 更新日志

查看 [升级评估报告.md](./升级评估报告.md) 了解详细的版本升级记录。

### 最新版本

- **Vite 8.0.12** - 升级到 Vite 8 (Rolldown)，构建速度提升 60%
- **Vue 3.5.34** - 最新稳定版 Vue 3
- **Element Plus 2.14.0** - 最新 UI 组件库
- **TypeScript 6.0.3** - 最新 TypeScript 版本
- **依赖全面升级** - 所有依赖升级到最新版本

## 📄 开源协议

本项目采用 MIT 协议开源 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🙏 致谢

- [Vue.js](https://vuejs.org/) - 渐进式 JavaScript 框架
- [Vite](https://vitejs.dev/) - 下一代前端构建工具
- [Element Plus](https://element-plus.org/) - Vue 3 UI 组件库
- [Pinia](https://pinia.vuejs.org/) - Vue 3 状态管理
- [markdown-it](https://github.com/markdown-it/markdown-it) - Markdown 解析器
- [Prism.js](https://prismjs.com/) - 代码高亮库

## 📧 联系方式

- **作者**: 七七
- **QQ**: 2316364297
- **网站**: https://www.aqi125.cn
- **GitHub**: [aqi-qihuan/aurora: 基于SpringBoot4.1.X+Vue3开发的个人博客系统](https://github.com/aqi-qihuan/aurora))

---

⭐ 如果这个项目对你有帮助，请给它一个 Star！
