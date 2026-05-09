<div align="center">
![Spring Boot](https://img.shields.io/badge/Spring%20Boot-4.1.0--M4-6DB33F?style=for-the-badge&logo=springboot)
![Go](https://img.shields.io/badge/Go-1.26-00ADD8?style=for-the-badge&logo=go)
![JDK](https://img.shields.io/badge/JDK-25-ED8B00?style=for-the-badge&logo=openjdk)
![Vue](https://img.shields.io/badge/Vue-3.4-4FC08D?style=for-the-badge&logo=vue.js)
![MySQL](https://img.shields.io/badge/MySQL-9.x-4479A1?style=for-the-badge&logo=mysql)
![Redis](https://img.shields.io/badge/Redis-7.x-DC382D?style=for-the-badge&logo=redis)
![Elasticsearch](https://img.shields.io/badge/ES-8.19.14-FEC514?style=for-the-badge&logo=elasticsearch)
![RabbitMQ](https://img.shields.io/badge/RabbitMQ-3.x-FF6600?style=for-the-badge&logo=rabbitmq)
![MinIO](https://img.shields.io/badge/MinIO-8.6-C72E49?style=for-the-badge&logo=minio)

<br/>

# 🌌 Aurora

**前后端分离博客系统 · 双后端架构**

*Spring Boot 4.x / Go 1.26 · JDK 25 · Vue 3 · Elasticsearch 8.x*

🚀 [快速开始](#-快速开始) · 🌐 [在线演示](#-在线地址) · 🛠️ [技术栈](#-技术栈) · 📦 [部署指南](#-部署)

</div>

---

## 📖 前言

> ⭐ 开源不易，希望大家 **Star** 支持一下！

由于本人还在上班，主语言并不是 Java，所以项目更新频率较慢，但是本项目会**长期维护**。有问题可以提 [Issue](https://github.com/zhouyqxy/aurora/issues)，也欢迎大家来共建此项目，包括但不限于：🐛 **Bug 修复**、✨ **代码优化**、🎉 **功能开发** 等。

---

## 🌐 在线地址

| 🖥️ 站点 | 🔗 链接 | 🔑 账号 |
|:--------|:--------|:--------|
| 🏠 前台 | [www.aqi125.cn](https://www.aqi125.cn) | — |
| ⚙️ 后台 | [admin.aqi125.cn](https://admin.aqi125.cn) | `test@163.com` / `123456` |

---

## 🎨 效果图

<div align="center">

<img src="https://ws.aqi125.cn/aurora/articles/a850a2955e44fb4728efba2a51590b1f.png" alt="首页展示" width="45%" /> &nbsp; <img src="https://ws.aqi125.cn/aurora/articles/d4e0269e395ae411c2d1187f0f51844a.png" alt="文章详情" width="45%" />

🏡 **首页展示** &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; 📄 **文章详情**

<br/>

<img src="https://ws.aqi125.cn/aurora/articles/46a2d83d8060fcd824ebb1c6c84f9fab.png" alt="博客列表" width="45%" /> &nbsp; <img src="https://ws.aqi125.cn/aurora/articles/864628ec3af76aa3fa33d8dea209e90b.png" alt="管理后台" width="45%" />

📋 **博客列表** &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; 🖥️ **管理后台**

</div>

---

## 🏗️ 项目结构

```
aurora-master/
├── aurora-springboot/       ☕ Java 后端（主力版本）
├── aurora-go/               🐹 Go 后端（轻量版本，可替换 Spring Boot）
├── aurora-vue/
│   ├── aurora-blog/         🏠 前台（用户端）
│   └── aurora-admin-v3/     ⚙️ 后台（管理端，Vue 3 重写）
├── aurora.sql               📄 数据库初始化脚本
└── README.md
```

### 双后端架构

Aurora 提供 **Java** 和 **Go** 两种后端实现，共享同一套前端和数据库：

| 对比项 | ☕ Spring Boot | 🐹 Go |
|:-------|:--------------|:------|
| 状态 | ✅ 主力版本，功能完整 | ✅ 已完成核心功能，可替换使用 |
| 内存占用 | ~280 MiB | ~29 MiB (**↓89.6%**) |
| 启动时间 | ~8s | ~0.3s (**↓96%**) |
| Docker 镜像 | ~180 MB (JRE) | ~5 MB (**↓97.2%**) |
| 总内存占用 | ~1,587 MiB | ~1,336 MiB (**↓15.8%**) |
| AI Agent | — | tRPC-Agent-Go v1.8 (可选) |
| 适用场景 | 功能优先、团队熟悉 Java | 资源受限、追求极致性能 |

> 💡 两个后端 API 完全兼容，前端无需修改即可切换。

---

## 🛠️ 技术栈

### ⚙️ 后端

#### ☕ Spring Boot（主力版本）

| 📦 分类 | 🛠️ 技术 | 📌 版本 | 📝 说明 |
|:--------|:--------|:--------|:--------|
| 🏗️ 基础框架 | **Spring Boot** | 4.1.0-M4 | 最新里程碑版 |
| ☕ 运行环境 | **JDK** | 25 | 最新版 |
| 🗄️ 持久化框架 | MyBatis-Plus | 3.5.16 | — |
| 🐬 数据库 | MySQL | 9.x | Connector 9.6.0 |
| 🔴 缓存中间件 | Redis Stack | 7.x | — |
| 🐇 消息中间件 | RabbitMQ | 3.x | — |
| 🔍 搜索引擎 | Elasticsearch | 8.19.14 | 原生 Java Client |
| ⏰ 任务调度 | Quartz | 6.x | — |
| 🔒 权限框架 | Spring Security | 6.x | — |
| 📚 API 文档 | SpringDoc OpenAPI | 2.8.0 | OpenAPI 3.x |
| ☁️ 对象存储 | MinIO / Aliyun OSS | 8.6.0 / 3.18.5 | 双存储支持 |
| 🔐 JWT 认证 | JJWT | 0.12.7 | — |
| 📄 JSON 处理 | FastJSON2 | 2.0.61 | — |
| 🔧 工具库 | Hutool | 5.8.44 | — |

#### 🐹 Go（轻量版本）

| 📦 分类 | 🛠️ 技术 | 📌 版本 | 📝 说明 |
|:--------|:--------|:--------|:--------|
| 🏗️ Web 框架 | Gin | 1.10 | — |
| ☕ 运行环境 | Go | 1.26 | — |
| 🗄️ ORM | GORM | 1.30 | — |
| 🐬 数据库 | MySQL | 9.x | GORM Driver |
| 🔴 缓存 | go-redis | 9.7 | — |
| 🐇 消息队列 | amqp091-go | 1.10 | — |
| 🔍 搜索引擎 | go-elasticsearch | 8.19.14 | 原生 ES 8.x Client |
| ☁️ 对象存储 | minio-go | 7.0 | — |
| 🔐 JWT | golang-jwt | 5.2 | — |
| 📅 定时任务 | robfig/cron | 3.0 | — |
| ⚙️ 配置管理 | Viper | 1.19 | — |
| 📝 日志 | Zap | 1.27 | 结构化日志 |
| 🤖 AI Agent | tRPC-Agent-Go | 1.8 | 腾讯开源，可选插件 |

### 🎭 前端

#### 🏠 前台 · aurora-blog

| 📦 分类 | 🛠️ 技术 | 📌 版本 |
|:--------|:--------|:--------|
| 🖼️ 基础框架 | **Vue 3** | 3.x |
| 🧩 UI 组件库 | Element Plus | 2.2.9 |
| 📊 状态管理 | Pinia | 2.0.14 |
| 🧭 路由组件 | Vue Router | 4.0.3 |
| 🌐 网络请求 | Axios | 0.27.2 |
| 🎨 样式框架 | Tailwind CSS | 2.x |
| 🌍 国际化 | Vue I18n | 9.1.10 |
| ✍️ 富文本编辑器 | Mavon Editor | 3.0.1 |
| 📈 图表库 | ECharts | 5.x |
| 📝 Markdown 解析 | markdown-it | 13.x |
| 🔤 代码高亮 | PrismJS | 1.28.x |
| 🖼️ SVG 图标 | vite-plugin-svg-icons | 2.0 |

#### ⚙️ 后台 · aurora-admin-v3

| 📦 分类 | 🛠️ 技术 | 📌 版本 |
|:--------|:--------|:--------|
| 🖼️ 基础框架 | **Vue 3** | 3.4.21 |
| 🧩 UI 组件库 | Element Plus | 2.5.6 |
| 📊 状态管理 | Pinia | 2.1.7 |
| 🧭 路由组件 | Vue Router | 4.3.0 |
| 🌐 网络请求 | Axios | 1.6.7 |
| ✍️ 富文本编辑器 | MdEditor V3 | 6.4.0 |
| ⚡ 构建工具 | **Vite** | 5.1.5 |
| 📈 图表库 | ECharts | 5.6.0 |
| 🧪 测试框架 | Vitest | 1.6.0 |
| 🎨 主题系统 | 深色/浅色 + 自定义主题 | — |
| 🔍 全局搜索 | Ctrl+K 快捷搜索 | — |
| 🕐 Cron 生成器 | 可视化 Cron 表达式 | — |

> 🎨 样式来源：[hexo aurora 主题](https://github.com/auroral-ui/hexo-theme-aurora)

### 🏗️ 中间件架构

```
┌──────────┐    ┌──────────────────┐    ┌──────────────┐    ┌─────────┐
│  Nginx   │───▶│ Spring Boot / Go │───▶│ Elasticsearch │    │  MinIO  │
│ (反向代理) │    │    (后端服务)      │    │   (全文检索)   │    │ (对象存储)│
└──────────┘    └────┬─────────────┘    └──────────────┘    └─────────┘
                     │
              ┌──────┼──────┐
              ▼      ▼      ▼
        ┌──────┐ ┌──────┐ ┌──────┐
        │ MySQL │ │Redis │ │  MQ  │
        │(持久化)│ │(缓存) │ │(消息) │
        └──────┘ └──────┘ └──────┘
```

---

## 📋 功能模块

### 🏠 前台功能

| 模块 | 功能 |
|:-----|:-----|
| 📄 文章 | Markdown 渲染、代码高亮、目录导航、Mermaid 图表、LaTeX 公式 |
| 🏷️ 标签 | 标签云、标签筛选 |
| 📂 分类 | 文章分类浏览 |
| 🗂️ 归档 | 按时间线归档 |
| 💬 留言板 | 话题留言、表情回复 |
| 📸 相册 | 照片墙、灯箱预览 |
| 🔗 友链 | 友情链接展示、申请 |
| 📢 说说 | 简短动态发布 |
| 🔍 搜索 | Elasticsearch 全文检索 |
| 🌙 主题 | 深色/浅色切换 |
| 🌍 国际化 | 中/英双语 |
| 📱 移动端 | 响应式布局、手势交互优化 |

### ⚙️ 后台功能

| 模块 | 功能 |
|:-----|:-----|
| 📝 文章管理 | 发布/编辑/删除、Markdown 编辑器、标签/分类 |
| 🏷️ 标签管理 | 增删改查、批量操作 |
| 📂 分类管理 | 增删改查、层级管理 |
| 💬 评论管理 | 审核/删除/搜索 |
| 🔗 友链管理 | 增删改查、上下架 |
| 📸 相册管理 | 创建相册、上传照片、回收站 |
| 📢 说说管理 | 发布/编辑/删除 |
| 👤 用户管理 | 用户列表、角色分配 |
| 🛡️ 角色权限 | RBAC 权限控制、菜单/按钮级权限 |
| 📋 菜单管理 | 动态路由、菜单配置 |
| 📊 资源管理 | 文件上传/管理 |
| ⏰ 定时任务 | Cron 可视化配置、执行日志 |
| 🌐 网站配置 | 基本信息、社交链接、功能开关 |
| 🎨 主题设置 | 深色/浅色/自定义主题、个人化配置 |
| 🔍 全局搜索 | Ctrl+K 全局搜索 |
| 📈 数据面板 | 访问统计、文章数据 |

---

## 🚀 快速开始

### ☕ Spring Boot 后端

```bash
# 1. 导入数据库
mysql -u root -p aurora < aurora.sql

# 2. 修改配置
cd aurora-springboot
# 编辑 src/main/resources/application-prod.yml，填入数据库/Redis/MQ/ES/MinIO 连接信息

# 3. 编译运行
mvn clean package -DskipTests
java -jar target/aurora-springboot-0.0.1.jar --spring.profiles.active=prod

# 或 Docker 部署
docker compose up -d
```

### 🐹 Go 后端

```bash
cd aurora-go

# 1. 安装依赖
go mod download

# 2. 配置
cp configs/config.yaml configs/config.local.yaml
# 编辑 config.local.yaml 填入中间件连接信息

# 3. 运行
go run cmd/server/main.go --config configs/config.yaml
# 或使用 Make
make run

# Docker 部署
make docker-up
```

### 🏠 前台

```bash
cd aurora-vue/aurora-blog

npm install
npm run dev      # 开发模式 → http://localhost:5173
npm run build    # 生产构建
```

### ⚙️ 后台

```bash
cd aurora-vue/aurora-admin-v3

npm install
npm run dev      # 开发模式 → http://localhost:8080
npm run build    # 生产构建
```

### 🐳 Docker Compose 一键部署

```bash
# Spring Boot 全套
cd aurora-springboot
docker compose up -d

# Go 全套
cd aurora-go
docker compose -f docker-compose-go.yml up -d
```

---

## 📦 部署

> 📖 详见项目中的 **部署文档**

---

## 📋 后续计划

- [x] 🔄 Go 版本重构
- [ ] 🤖 接入 Agent — tRPC-Agent-Go 完善
- [ ] 📦 前后端一体化 Docker Compose 编排
- [ ] 🧪 Go 后端测试覆盖率提升
- [ ] 📱 PWA 支持

---

## 💬 交流群

| 📱 社群 | 🔢 号码 |
|:--------|:--------|
| 💬 QQ 群 | **338371628** |

---

## 🙏 鸣谢

感谢 [linhaojun857](https://github.com/linhaojun857) 提供的 Aurora 原版代码。

---

<div align="center">

[![Powered by DartNode](https://dartnode.com/branding/DN-Open-Source-sm.png)](https://dartnode.com "Powered by DartNode - Free VPS for Open Source")

</div>
