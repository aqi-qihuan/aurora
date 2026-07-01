# cmd/diag 诊断工具集

> 定位：开发期一次性诊断脚本，用于排查特定数据问题。所有文件带 `//go:build ignore` 标签，**不参与正常编译**，需手动 `go run` 执行。

---

## 工具清单

| 文件 | 用途 | 依赖 |
|:---|:---|:---|
| `unique_view_diag.go` | 一周访问趋势诊断（t_unique_view 表 + Redis 访客 Set + 查询条件修复建议） | MySQL + Redis + config |
| `unique_visitor_diag.go` | 访客数据统计诊断（Redis key 新旧格式检查 + 错误数据修复建议） | MySQL + Redis |
| `visit_stats_diag.go` | 访客数据统计诊断（修复后版本，按天 key + 数据库记录校验） | MySQL + Redis |
| `message_count_diag.go` | 评论表 type 分布统计（文章评论/留言板/其它类型计数） | MySQL |

---

## 一、unique_view_diag.go — 一周访问趋势诊断

**排查问题**：首页"一周访问趋势"图表数据为空或不准确。

**诊断内容**：
1. 查询 `t_unique_view` 表最近 20 条记录
2. 模拟实际查询条件（`create_time > ? AND create_time <= ?`）
3. 对比修复后查询条件（`>= beginOfDay AND <= endOfDay`）
4. 检查 Redis 全局访客 Set（`constant.UniqueVisitor`）的 SCard 与成员
5. 检测遗留的按天 key（旧实现残留，建议 DEL）
6. 生成完整 7 天数据（DB 历史 + Redis 今日实时）

**运行方式**：

```bash
# 需要本地存在 configs/config-local.yaml
go run cmd/diag/unique_view_diag.go
```

**输出解读**：
- 【2】与【3】结果不一致 → 查询条件边界问题，需把 `>` 改为 `>=` 并对齐到 endOfDay
- Redis 数据为 0 → 前端访客上报未写入 Redis，检查 `aurora_info_handler.go`
- 发现 `unique_visitor:YYYY-MM-DD` 按天 key → 旧实现残留，建议 `DEL`

---

## 二、unique_visitor_diag.go — 访客数据统计诊断

**排查问题**：访客数统计异常（旧格式 key 累积 / DB 出现 `views_count=0` 错误记录）。

**诊断内容**：
1. 检查 Redis 旧格式 key（`constant.UniqueVisitor` 全局 Set）的 IP 数
2. 检查最近 7 天新格式按天 key（`unique_visitor:YYYY-MM-DD`）的 IP 数
3. 查询数据库 `t_unique_view` 最近 7 天记录
4. 输出诊断结论与修复方案

**运行方式**：

```bash
go run cmd/diag/unique_visitor_diag.go
```

**修复方案**（脚本输出）：
- 方案1（推荐）：`DELETE FROM t_unique_view WHERE views_count = 0 AND create_time >= '2026-04-09';`
- 方案2：Redis `DEL unique_visitor`，从当天重新开始用按天 key

> ⚠️ 该脚本内含硬编码的 Redis/MySQL 连接信息，仅限内网调试用，**禁止提交到公开仓库或生产环境执行**。

---

## 三、visit_stats_diag.go — 访客数据统计诊断（修复后）

**定位**：`unique_visitor_diag.go` 的修复后版本，用于验证修复是否生效。

**诊断内容**：
1. 检查最近 7 天 Redis 按天 key 的 IP 数（`✅` 有数据 / `❌` 无数据）
2. 查询数据库 `t_unique_view` 最近 7 天记录，标记 `views_count=0` 的错误记录
3. 输出修复建议

**运行方式**：

```bash
go run cmd/diag/visit_stats_diag.go
```

---

## 四、message_count_diag.go — 评论表 type 分布统计

**排查问题**：首页留言板数量、文章评论数量统计异常。

**诊断内容**：
1. 统计 `t_comment` 表按 `type` 分组的记录数
2. 统计 `type IN (1, 2, 5)` 的总数
3. 输出 `type=2`（留言板）原始数据（id/content/parent/type）
4. 输出 `type=1`（文章评论）原始数据

**type 含义对照**：

| type 值 | 含义 |
|:---:|:---|
| 1 | 文章评论 |
| 2 | 留言板 |
| 5 | 其它（说说/相册等） |

**运行方式**：

```bash
go run cmd/diag/message_count_diag.go
```

---

## 安全注意事项

> ⚠️ **3 个脚本（unique_visitor_diag.go / visit_stats_diag.go / message_count_diag.go）内含硬编码的数据库密码与 IP 地址。**

- 这些脚本仅用于内网开发环境的一次性排查
- **禁止**在生产环境执行
- **禁止**将含密码的版本提交到公开仓库
- 后续维护建议：改为从 `.env` 或 `configs/config-local.yaml` 读取连接信息（参考 `unique_view_diag.go` 的实现）

---

## 添加新诊断脚本规范

1. 文件放 `cmd/diag/` 下，命名 `{主题}_diag.go`
2. **必须**加 `//go:build ignore` 首行，避免污染正常编译
3. **必须**从配置文件或环境变量读取连接信息，禁止硬编码密码
4. 脚本顶部注释说明：排查问题、运行方式、输出解读
5. 更新本文档清单
