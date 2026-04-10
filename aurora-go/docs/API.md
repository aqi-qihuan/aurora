# Aurora Go - API Reference

> **OpenAPI 3.0 规范** | 版本 1.0 | 基础路径: `http://localhost:8080/api`

---

## 目录

1. [认证 (Auth)](#认证-auth)
2. [文章 (Articles)](#文章-articles)
3. [分类标签 (Category & Tag)](#分类标签-category--tag)
4. [评论 (Comments)](#评论-comments)
5. [用户 (Users)](#用户-users)
6. [相册 (Photos & Albums)](#相册-photos--albums)
7. [友链 (Friend Links)](#友链-friend-links)
8. [说说 (Talks)](#说说-talks)
9. [文件上传 (Upload)](#文件上传-upload)
10. [网站配置 (Website Config)](#网站配置-website-config)
11. [定时任务 (Jobs)](#定时任务-jobs)
12. [日志 (Logs)](#日志-logs)
13. [AI Agent (AI智能体)](#ai-agent-ai智能体)
14. [数据统计 (Statistics)](#数据统计-statistics)
15. [错误码 (Error Codes)](#错误码-error-codes)

---

## 全局说明

### 认证方式

```
Authorization: Bearer <JWT_TOKEN>
```

### 统一响应格式

**成功响应:**
```json
{
  "code": 200,
  "message": "操作成功",
  "data": { ... }
}
```

**分页响应:**
```json
{
  "code": 200,
  "data": {
    "recordList": [...],
    "count": 100
  }
}
```

**错误响应:**
```json
{
  "code": 500,
  "message": "错误描述",
  "data": null
}
```

### 通用查询参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `current` | int | 1 | 页码 |
| `size` | int | 10 | 每页条数 |
| `keyword` | string | - | 搜索关键词 |

---

## 认证 (Auth)

### POST /auth/login - 用户登录

登录并返回 JWT Token。

**请求体:**
```json
{
  "username": "admin",
  "password": "123456"
}
```

**响应:**
```json
{
  "code": 200,
  "data": {
    "accessToken": "eyJhbGciOiJIUzI1NiIs...",
    "user": {
      "id": 1,
      "username": "admin",
      "nickname": "管理员",
      "avatar": "https://...",
      "roleList": ["admin"]
    }
  }
}
```

---

### POST /auth/register - 用户注册

注册新账号。

**请求体:**
```json
{
  "email": "user@example.com",
  "code": "123456"  // 邮箱验证码
}
```

---

### GET /auth/userinfo - 获取当前用户信息

需要: `Authorization: Bearer <token>`

**响应:**
```json
{
  "code": 200,
  "data": {
    "id": 1,
    "username": "admin",
    "nickname": "管理员",
    "avatar": "https://...",
    "website": "https://example.com",
    "intro": "个人简介"
  }
}
```

---

### POST /auth/logout - 登出

需要: `Authorization: Bearer <token>`

将 Token 加入 Redis 黑名单。

---

### POST /auth/password - 修改密码

需要: `Authorization: Bearer <token>`

**请求体:**
```json
{
  "oldPassword": "旧密码",
  "newPassword": "新密码"
}
```

---

### POST /auth/oauth/qq - QQ OAuth登录

QQ OAuth 回调接口。

**请求体:**
```json
{
  "accessToken": "QQ_ACCESS_TOKEN",
  "openId": "QQ_OPEN_ID"
}
```

---

## 文章 ( Articles)

### GET /articles - 前台文章列表

前台展示用，支持分页和搜索。

**Query参数:**

| 参数 | 类型 | 说明 |
|------|------|------|
| `current` | int | 页码 |
| `size` | int | 每页数 |
| `categoryId` | int | 分类ID筛选 |
| `tagId` | int | 标签ID筛选 |
| `keyword` | string | 标题关键词搜索 |

**响应:**
```json
{
  "code": 200,
  "data": {
    "recordList": [
      {
        "id": 1,
        "articleTitle": "Go语言入门指南",
        "articleCover": "https://...",
        "articleContent": "...",
        "categoryName": "技术",
        "tagName": ["Go", "后端"],
        "viewCount": 1234,
        "likeCount": 56,
        "createTime": "2026-04-11T00:00:00Z"
      }
    ],
    "count": 100
  }
}
```

---

### GET /articles/:id - 文章详情

获取单篇文章完整内容。

**Path参数:** `id` - 文章ID

---

### GET /articles/topAndFeatured - 置顶和推荐文章

获取置顶和推荐的文章列表。

**响应:**
```json
{
  "code": 200,
  "data": [
    {
      "id": 1,
      "articleTitle": "热门文章标题",
      "articleCover": "https://...",
      "isTop": true,
      "isFeatured": true
    }
  ]
}
```

---

### GET /articles/archives - 文章归档

按年月归档的文章列表。

---

### GET /articles/search - 搜索文章

支持 ES 和 MySQL 双模式搜索。

**Query参数:**

| 参数 | 类型 | 说明 |
|------|------|------|
| `keyword` | string | **必填** - 搜索关键词 |
| `mode` | string | 搜索模式: `es`(默认) / `mysql` |

**响应 (ES模式):**
```json
{
  "code": 200,
  "data": {
    "recordList": [...],
    "count": 50
  },
  "searchMode": "es"
}
```

---

### POST /admin/articles - 发布文章 (后台)

需要: `JWT + article:write 权限`

**请求体:**
```json
{
  "articleTitle": "文章标题",
  "articleContent": "# Markdown 内容",
  "articleCover": "https://cover.jpg",
  "categoryId": 1,
  "tagName": ["Go", "Gin"],
  "status": 1,           // 1-公开, 2-草稿, 3-密码保护
  "isTop": false,
  "isFeatured": false,
  "type": 1,             // 1-原创, 2-转载, 3-翻译
  "originalUrl": "",     // 转载来源URL(当type=2时必填),
  "password": ""         // 密码保护密码(当status=3时)
}
```

---

### PUT /admin/articles/:id - 更新文章

同发布，传入文章ID。

---

### DELETE /admin/articles/:id - 删除文章

逻辑删除（移入回收站）。

---

### PUT /admin/articles/top - 置顶/取消置顶

**请求体:** `{ "id": 1, "isTop": true }`

---

### PUT /admin/articles/featured - 推荐/取消推荐

**请求体:** `{ "id": 1, "isFeatured": true }`

---

### GET /admin/articles/recycle - 回收站文章

获取已删除文章列表。

---

### PUT /admin/articles/restore/:id - 恢复文章

从回收站恢复。

---

### DELETE /admin/articles/delete/:id - 永久删除

物理删除文章。

---

### POST /admin/articles/export - 导出文章

导出为 Markdown 文件。

**请求体:** `{ "id": 1 }`

---

### POST /admin/articles/import - 导入文章

导入 Markdown 文件创建文章。

**请求:** multipart/form-data, 字段: file

---

## 分类标签 (Category & Tag)

### GET /categories - 分类列表

返回所有分类及其文章数量。

**响应:**
```json
{
  "code": 200,
  "data": [
    {
      "id": 1,
      "categoryName": "技术",
      "articleCount": 42
    }
  ]
}
```

---

### POST /admin/categories - 新增分类

**请求体:** `{ "categoryName": "新分类" }`

---

### PUT /admin/categories - 更新分类

**请求体:** `{ "id": 1, "categoryName": "更新名称" }`

---

### DELETE /admin/categories/:id - 删除分类

---

### GET /tags - 标签列表

---

### POST /admin/tags - 新增标签

**请求体:** `{ "tagName": "Go" }`

---

### PUT /admin/tags - 更新标签

**请求体:** `{ "id": 1, "tagName": "Golang" }`

---

### DELETE /admin/tags/:id - 删除标签

---

## 评论 (Comments)

### GET /comments/:type/:resourceId - 评论列表

**Path参数:**
- `type`: 评论类型 (1=文章, 2=说说, 3=友链, 4=关于, 5=留言)
- `resourceId`: 资源ID

**Query参数:** `current`, `size`

**响应:**
```json
{
  "code": 200,
  "data": {
    "recordList": [
      {
        "id": 100,
        "commentContent": "很棒的文章！",
        "replyUserId": null,
        "replyUserNickname": null,
        "nickname": "访客",
        "avatar": "https://...",
        "likeCount": 5,
        "createTime": "2026-04-11T00:00:00Z",
        "replies": [...]   // 子回复
      }
    ],
    "count": 20
  }
}
```

---

### POST /comments/:type/:resourceId - 发表评论

**请求体:**
```json
{
  "commentContent": "评论内容",
  "replyUserId": null       // 回复的父评论用户ID(可选)
}
```

---

### DELETE /admin/comments/:id - 删除评论 (后台)

---

### GET /admin/comments/review - 待审核评论列表

---

### PUT /admin/comments/review/:id - 审核评论

**请求体:** `{ "isReview": true }`

---

## 用户 (Users)

### GET /admin/users - 用户列表 (后台管理)

需要: `JWT + user:list 权限`

---

### PUT /admin/users/status - 禁用/启用用户

**请求体:** `{ "id": 1, "isDisable": true }`

---

### PUT /admin/users/password - 重置用户密码

**请求体:** `{ "id": 1, "password": "newPassword123" }`

---

### GET /users/info/:id - 用户公开信息

---

### PUT /users/info - 更新个人信息

需要: `JWT`

**请求体:** `{ "nickname": "昵称", "intro": "简介", ... }`

---

### PUT /users/avatar - 更新头像

**请求:** multipart/form-data, 字段: file

---

### PUT /users/role - 切换角色

需要: `JWT`

**请求体:** `{ "roleId": 2 }`

---

## 相册 (Photos & Albums)

### GET /photos/albums - 相册列表

**响应:**
```json
{
  "code": 200,
  "data": [
    {
      "id": 1,
      "albumName": "旅行照片",
      "albumCover": "https://...",
      "photoCount": 25,
      "isPublished": true,
      "description": "2026春游记录"
    }
  ]
}
```

---

### POST /admin/photos/albums - 新建相册

**请求体:**
```json
{
  "albumName": "新相册",
  "description": "相册描述",
  "isPublished": true
}
```

---

### PUT /admin/photos/albums - 更新相册

---

### DELETE /admin/photos/albums/:id - 删除相册

---

### GET /photos/albums/:id - 相册详情(含照片列表)

---

### POST /admin/photos/albums/upload - 上传相册封面

**请求:** multipart/form-data, 字段: albumId, file

---

### GET /photos/albums/delete - 已删除相册(回收站)

---

### PUT /admin/photos/albums/restore/:id - 恢复相册

---

### DELETE /admin/photos/albums/delete/:id - 永久删除相册

---

## 友链 (Friend Links)

### GET /links - 前台友链列表

仅返回已审核通过的友链。

**响应:**
```json
{
  "code": 200,
  "data": [
    {
      "id": 1,
      "linkName": "示例博客",
      "linkAvatar": "https://...",
      "linkAddress": "https://example.com",
      "linkIntro": "技术博客"
    }
  ]
}
```

---

### POST /admin/links - 申请友链

**请求体:**
```json
{
  "linkName": "我的博客",
  "linkAvatar": "https://...",
  "linkAddress": "https://myblog.com",
  "linkIntro": "关于我"
}
```

---

### PUT /admin links - 更新友链

---

### DELETE /admin/links/:id - 删除友链

---

### GET /admin/links/review - 待审核友链

---

### PUT /admin/links/review/:id - 审核友链

**请求体:** `{ "isReview": true }`

---

## 说说 (Talks)

### GET /talks/max - 最新说说

获取最新一条说说。

**响应:**
```json
{
  "code": 200,
  "data": {
    "id": 1,
    "talkContent": "今天天气真好 ☀️",
    "images": ["https://...", "https://..."],
    "likeCount": 0,
    "createTime": "2026-04-11T00:00:00Z"
  }
}
```

---

### GET /talks - 说说列表 (前台)

---

### POST /admin/talks - 发布说说

**请求体:**
```json
{
  "talkContent": "说说内容",
  "images": ["url1", "url2"],
  "isTop": false,
  "status": 1
}
```

---

### PUT /admin/talks - 更新说说

---

### DELETE /admin/talks/:id - 删除说说

---

## 文件上传 ( Upload)

### POST /upload/file - 通用文件上传

**请求:** multipart/form-data, 字段: file

**响应:**
```json
{
  "code": 200,
  "data": {
    "url": "https://minio.example.com/aurora/uploads/xxx.jpg",
    "filename": "photo.jpg",
    "size": 102400
  }
}
```

支持的图片格式: jpg/jpeg/png/gif/webp/svg, 最大 5MB

---

### POST /upload/images - 图片批量上传

**请求:** multipart/form-data, 字段: files[] (多文件)

---

### POST /upload/config/images - 配置类图片上传

用于头像、Logo、Favicon、收款码等配置图片。

---

## 网站配置 (Website Config)

### GET /admin/website/config - 获取网站配置

**响应:**
```json
{
  "code": 200,
  "data": {
    "config": {
      "authorAvatar": "https://...",
      "logo": "https://...",
      "favicon": "https://...",
      "name": "Aurora Blog",
      "summary": "AI驱动的下一代博客系统",
      "notice": "公告内容",
      "footer": "备案信息"
    },
    "socialInfo": {
      "github": "https://github.com/...",
      "gitee": "https://gitee.com/...",
      "wechat": "wx_code.png"
    },
    "otherConfig": {
      "touristAvatar": "https://...",
      "emailNoticeEnabled": true,
      "commentCheckEnabled": true,
      "rewardEnabled": true
    }
  }
}
```

---

### PUT /admin/website/config - 更新网站配置

---

### GET /about - 关于页面信息

---

## 定时任务 ( Jobs)

### GET /jobs - 任务列表

**响应:**
```json
{
  "code": 200,
  "data": [
    {
      "id": 1,
      "jobName": "每日UV统计",
      "groupName": "default",
      "cronExpression": "0 0 0 * * ?",
      "status": 0,          // 0-正常, 1-暂停
      "param": "{}",
      "remark": "统计独立访客数"
    }
  ]
}
```

---

### POST /jobs - 新建任务

**请求体:**
```json
{
  "jobName": "新任务",
  "groupName": "default",
  "cronExpression": "0 0 2 * * ?",
  "param": "{}",
  "remark": "任务描述"
}
```

---

### PUT /jobs - 更新任务

---

### DELETE /jobs/:id - 删除任务

---

### PUT /jobs/status - 启停任务

**请求体:** `{ "id": 1, "status": 0 }`

---

### POST /jobs/run - 手动执行一次

**请求体:** `{ "id": 1 }`

---

### GET /jobLogs/:jobId - 任务调度日志

**Query参数:** `current`, `size`

**响应:**
```json
{
  "code": 200,
  "data": {
    "recordList": [
      {
        "id": 1,
        "jobId": 1,
        "jobMessage": "执行成功",
        "status": 1,       // 0-失败, 1-成功
        "duration": 1523,  // 执行耗时(ms)
        "createTime": "2026-04-11T02:00:00Z"
      }
    ],
    "count": 365
  }
}
```

---

### DELETE /jobLogs/:ids - 批量删除日志

---

## 日志 ( Logs)

### GET /operation/logs - 操作日志

**Query参数:** `current`, `size`, `optModule`(模块), `startDate`, `endDate`

---

### GET /exception/logs - 异常日志

**Query参数:** `current`, `size`, `startDate`, `endDate`

---

### DELETE /operation/logs - 清空操作日志

---

### DELETE /exception/logs - 清空异常日志

---

## AI Agent (AI智能体)

> **可选模块** | 需要: `agent.enabled=true` + `JWT Token`

所有 Agent 端点前缀: `/api/agent/*`

---

### POST /api/agent/chat - AI 对话 (SSE流式)

与 AI 进行多轮对话，支持 SSE 实时流式输出。

**请求头:** `Accept: text/event-stream`

**请求体:**
```json
{
  "sessionId": "sess_abc123",   // 会话ID(可选, 首次对话自动创建)
  "message": "帮我写一篇关于Go语言的博客文章",
  "model": "deepseek-chat",      // 可选模型: gpt-4o/claude-3.5/deepseek-chat/qwen-max
  "temperature": 0.7,            // 可选: 0.0~1.0
  "stream": true                 // 是否流式输出(默认true)
}
```

**SSE响应流:**
```
event: message
data: {"type":"token","content":"Go语言"}

event: message
data: {"type":"token","content":"是Google开发"}

event: message
data: {"type":"done","sessionId":"sess_abc123","tokenUsage":{"prompt":150,"completion":320,"total":470}}
```

**非SSE同步响应(stream=false):**
```json
{
  "code": 200,
  "data": {
    "sessionId": "sess_abc123",
    "reply": "完整的AI回答内容...",
    "tokenUsage": { "prompt": 150, "completion": 320, "total": 470 }
  }
}
```

---

### POST /api/agent/write - AI 写作助手

提供6种写作辅助能力。

**请求体:**
```json
{
  "operation": "generate",       // 操作类型:
                                // generate  - 生成全文
                                // continue  - 续写
                                // polish    - 润色改写
                                // summarize - 提取摘要
                                // keywords  - 推荐关键词
                                // seo       - SEO优化建议
  "topic": "Go语言并发编程实战",
  "content": "现有文本内容...",  // continue/polish 时必填
  "style": "professional",       // 风格: professional/casual/technical/creative
  "length": "medium",            // 长度: short(~300字)/medium(~800字)/long(~1500字+)
  "language": "zh-cn"            // 语言
}
```

**响应:**
```json
{
  "code": 200,
  "data": {
    "result": "生成的Markdown内容...",
    "keywords": ["Go", "goroutine", "channel", "并发"],
    "wordCount": 850,
    "readingTime": "3 min",
    "seoScore": 92
  }
}
```

---

### POST /api/agent/search - AI 语义搜索

基于 RAG 的增强语义搜索，理解自然语言意图。

**请求体:**
```json
{
  "query": "Go语言如何实现高性能并发？",
  "useRag": true,              // 是否启用RAG增强检索(默认true)
  "topK": 5,                   // 返回结果数量
  "mode": "hybrid"             // 搜索模式: hybrid(混合)/semantic(纯语义)/keyword(关键词)
}
```

**响应:**
```json
{
  "code": 200,
  "data": {
    "answer": "Go语言通过goroutine和channel机制实现了高性能并发...",
    "sources": [
      {
        "id": 15,
        "title": "Go并发编程实战指南",
        "snippet": "...高亮匹配片段...",
        "relevance": 0.95,
        "url": "/articles/15"
      }
    ],
    "relatedQuestions": [
      "goroutine vs thread有什么区别？",
      "channel缓冲区大小如何选择？"
    ],
    "searchMetadata": {
      "mode": "hybrid",
      "ragEnabled": true,
      "totalResults": 23,
      "searchTimeMs": 156
    }
  }
}
```

---

### POST /api/agent/analyze - 数据分析与洞察

AI驱动的博客运营数据分析。

**请求体:**
```json
{
  "analysisType": "overview",   // 分析类型:
                                // overview   - 总览仪表盘
                                // traffic    - 流量趋势分析
                                // content    - 内容表现分析
                                // audience   - 受众画像分析
                                // prediction - 趋势预测
                                // suggestion - 运营建议
  "dateRange": {
    "start": "2026-03-01",
    "end": "2026-04-11"
  },
  "includeRecommendations": true
}
```

**响应 (overview):**
```json
{
  "code": 200,
  "data": {
    "summary": {
      "totalPV": 125000,
      "totalUV": 45000,
      "avgStayTime": "3m 24s",
      "bounceRate": 38.5,
      "newArticleCount": 12
    },
    "charts": {
      "trafficTrend": [{ "date": "04-01", "pv": 3500, "uv": 1200 }],
      "topArticles": [{ "id": 1, "title": "...", "views": 5200 }],
      "geoDistribution": [{ "region": "广东", "count": 8500 }],
      "deviceDistribution": [{ "device": "Desktop", "percent": 62 }]
    },
    "insights": [
      "本周PV环比增长15.3%，主要来自搜索引擎流量",
      "\"Go语言入门\"系列文章持续贡献40%以上访问量",
      "移动端占比提升至38%, 建议继续优化移动端体验"
    ],
    "recommendations": [
      "建议增加\"Docker容器化\"相关内容, 该领域搜索量上升30%",
      "每周二、四下午14:00-16:00为访问高峰, 建议此时发布新文章"
    ]
  }
}
```

---

### GET /api/agent/sessions - 会话列表

获取历史对话会话列表。

**Query参数:** `limit`(默认20), `offset`(默认0)

**响应:**
```json
{
  "code": 200,
  "data": [
    {
      "id": "sess_abc123",
      "title": "Go语言博客写作",
      "createdAt": "2026-04-11T03:00:00Z",
      "messageCount": 8,
      "lastActiveAt": "2026-04-11T04:20:00Z"
    }
  ]
}
```

---

### DELETE /api/agent/sessions/:id - 删除会话

删除指定会话及其消息历史。

---

### GET /api/agent/stats - Agent 使用统计

**响应:**
```json
{
  "code": 200,
  "data": {
    "chatSessions": 156,
    "chatMessages": 2890,
    "writeOperations": 89,
    "searchQueries": 445,
    "analyzeReports": 34,
    "totalTokensConsumed": 1250000,
    "dailyStats": [
      { "date": "04-10", "chats": 23, "writes": 5, "searches": 31 }
    ]
  }
}
```

---

### GET /api/agent/models - 可用模型列表

**响应:**
```json
{
  "code": 200,
  "data": [
    { "id": "gpt-4o", "name": "GPT-4o", "provider": "OpenAI", "maxTokens": 128000 },
    { "id": "claude-3.5-sonnet", "name": "Claude 3.5 Sonnet", "provider": "Anthropic", "maxTokens": 200000 },
    { "id": "deepseek-chat", "name": "DeepSeek Chat", "provider": "DeepSeek", "maxTokens": 65536 },
    { "id": "qwen-max", "name": "通义千问 Max", "provider": "Alibaba", "maxTokens": 32768 }
  ]
}
```

---

### PUT /api/agent/config - 更新Agent配置

**请求体:**
```json
{
  "defaultModel": "deepseek-chat",
  "temperature": 0.7,
  "maxTokens": 2048,
  "ragEnabled": true,
  "autoModerate": true,
  "streamDefault": true
}
```

---

## 数据统计 ( Statistics)

### GET /admin/info/back - 后台统计数据

聚合首页仪表盘数据。

**响应:**
```json
{
  "code": 200,
  "data": {
    "userCount": 1234,
    "articleCount": 567,
    "categoryCount": 12,
    "tagCount": 89,
    "todayPv": 1234,
    "todayUv": 567,
    "messageCount": 345,
    "pendingComments": 12,
    "pendingLinks": 5
  }
}
```

---

### GET /home/info - 前台首页数据

---

### GET /stats/views - 浏览量排行 (Redis ZSet)

**Query参数:** `limit`(默认10)

---

### GET /stats/unique-view - 独立访客统计

**Query参数:** `startDate`, `endDate`

---

### GET /stats/area - 地域分布 (GeoIP)

---

### GET /stats/today - 今日实时统计

---

## 错误码 (Error Codes)

| 错误码 | 说明 | HTTP状态码 |
|--------|------|-----------|
| 200 | 操作成功 | 200 |
| 400 | 请求参数错误 | 400 |
| 401 | 未认证/Token无效 | 401 |
| 403 | 无权限访问 | 403 |
| 404 | 资源不存在 | 404 |
| 409 | 资源冲突(如重复) | 409 |
| 429 | 请求频率超限 | 429 |
| 500 | 服务器内部错误 | 500 |
| 503 | 服务不可用(Agent降级) | 503 |

### 业务错误码

| 错误码 | 说明 |
|--------|------|
| 1001 | 用户名或密码错误 |
| 1002 | 用户已被禁用 |
| 1003 | Token已过期 |
| 1004 | Token无效 |
| 1101 | 用户已存在 |
| 1102 | 验证码错误或过期 |
| 2001 | 文章不存在 |
| 2002 | 文章无权限操作 |
| 3001 | 分类不存在 |
| 3002 | 分类下有文章, 无法删除 |
| 4001 | 敏感词检测未通过 |
| 9001 | Agent服务不可用 |
| 9002 | LLM调用失败 |
| 9003 | RAG检索无结果 |
| 9999 | 系统内部错误 |

---

*文档版本: 1.0.0 | 最后更新: 2026-04-11*
