<div align="center">

![Spring Boot](https://img.shields.io/badge/Spring%20Boot-4.1.0--M4-6DB33F?style=for-the-badge&logo=springboot)
![JDK](https://img.shields.io/badge/JDK-25-ED8B00?style=for-the-badge&logo=openjdk)
![Vue](https://img.shields.io/badge/Vue-3.4-4FC08D?style=for-the-badge&logo=vue.js)
![MySQL](https://img.shields.io/badge/MySQL-8.x-4479A1?style=for-the-badge&logo=mysql)
![Redis](https://img.shields.io/badge/Redis-7.x-DC382D?style=for-the-badge&logo=redis)
![Elasticsearch](https://img.shields.io/badge/ES-8.17-FEC514?style=for-the-badge&logo=elasticsearch)
![RabbitMQ](https://img.shields.io/badge/RabbitMQ-3.x-FF6600?style=for-the-badge&logo=rabbitmq)
![MinIO](https://img.shields.io/badge/MinIO-8.6-C72E49?style=for-the-badge&logo=minio)

<br/>

# 🌌 Aurora

**前后端分离博客系统**

*Spring Boot 4.x · JDK 25 · Vue 3 · Elasticsearch 8.x*

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

> 💡 轻量版后端：[aurora_Lite](https://github.com/zhouyqxy/aurora_Lite)

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

## 🛠️ 技术栈

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

#### ⚡ 后台 · aurora-admin-v3

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

> 🎨 样式来源：[hexo aurora 主题](https://github.com/auroral-ui/hexo-theme-aurora)

### ⚙️ 后端

| 📦 分类 | 🛠️ 技术 | 📌 版本 | 📝 说明 |
|:--------|:--------|:--------|:--------|
| 🏗️ 基础框架 | **Spring Boot** | 4.1.0-M4 | 最新里程碑版 |
| ☕ 运行环境 | **JDK** | 25 | 最新 LTS |
| 🗄️ 持久化框架 | MyBatis-Plus | 3.5.16 | — |
| 🐬 数据库 | MySQL | 8.x | — |
| 🔴 缓存中间件 | Redis Stack | 7.x | — |
| 🐇 消息中间件 | RabbitMQ | 3.x | — |
| 🔍 搜索引擎 | Elasticsearch | 8.17.2 | 原生 Java Client |
| ⏰ 任务调度 | Quartz | 6.x | — |
| 🔒 权限框架 | Spring Security | 6.x | — |
| 📚 API 文档 | SpringDoc OpenAPI | 2.8.0 | OpenAPI 3.x |
| ☁️ 对象存储 | MinIO / Aliyun OSS | 8.6.0 / 3.18.5 | 双存储支持 |
| 🔐 JWT 认证 | JJWT | 0.12.7 | — |
| 📄 JSON 处理 | FastJSON2 | 2.0.61 | — |
| 🔧 工具库 | Hutool | 5.8.44 | — |

### 🏗️ 中间件架构

```
┌──────────┐    ┌──────────┐    ┌──────────────┐    ┌─────────┐
│  Nginx   │───▶│  Spring  │───▶│ Elasticsearch │    │  MinIO  │
│ (反向代理) │    │   Boot   │    │   (全文检索)   │    │ (对象存储)│
└──────────┘    └────┬─────┘    └──────────────┘    └─────────┘
                     │
              ┌──────┼──────┐
              ▼      ▼      ▼
        ┌──────┐ ┌──────┐ ┌──────┐
        │ MySQL │ │Redis │ │  MQ  │
        │(持久化)│ │(缓存) │ │(消息) │
        └──────┘ └──────┘ └──────┘
```

---

## 📋 后续计划

- [ ] 🔄 Go 版本重构
- [ ] 🤖 接入 Agent — tRPC-Agent-Go 改造
- [ ] 📦 后端提供轻量化选择

---

## 🚀 快速开始

### ⚡ 一键安装

```shell
curl -sSL https://kangxianghui.top/api/Util/OnlineView/aurora_shell/aurora_install.sh -o aurora_install.sh && sh aurora_install.sh
```

> ⚠️ 适用于 **CentOS** 操作系统

### 🛠️ 手动部署

详见项目部署文档。

---

## 📦 部署

> 📖 详见项目中的 **部署文档**

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
