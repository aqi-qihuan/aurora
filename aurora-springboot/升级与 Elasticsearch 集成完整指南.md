# Aurora 博客系统 - 升级与 Elasticsearch 集成完整指南

<div align="center">

![Elasticsearch](https://img.shields.io/badge/Elasticsearch-8.17.2-00BFb3?style=flat-square&logo=elasticsearch)
![Spring Boot](https://img.shields.io/badge/Spring%20Boot-4.0+-6DB33F?style=flat-square&logo=spring)
![Version](https://img.shields.io/badge/Version-v2.0-blue?style=flat-square)
![Last Updated](https://img.shields.io/badge/Last%20Updated-2026--04--08-orange?style=flat-square)

</div>

---

## 目录

1. [Elasticsearch 8.17.2 集成问题排查](#1-elasticsearch-8172-集成问题排查)
   - [问题概述](#11-问题概述)
   - [常见问题与解决方案](#12-常见问题与解决方案)
   - [关键配置文件](#13-关键配置文件)
   - [经验总结](#14-经验总结)
2. [常用命令与运维](#2-常用命令与运维)
   - [ES 常用命令](#21-es-常用命令)
   - [搜索模式切换](#22-搜索模式切换)
   - [应用重启流程](#23-应用重启流程)
   - [常见问题快速解决](#24-常见问题快速解决)
3. [参考资料](#3-参考资料)

---

# 1. Elasticsearch 8.17.2 集成问题排查

## 1.1 问题概述

在将 Elasticsearch 升级到 8.17.2 后，遇到搜索功能完全失效的问题。经过排查，发现多个关键问题导致搜索无法正常工作。

> **📌 核心问题**：索引配置、字段映射、数据同步、分词器、超时设置

---

## 1.2 常见问题与解决方案

### ❌ 问题 1: ES 索引 Mapping 配置错误

| 类别 | 内容 |
|:---|:---|
| 🔴 **现象** | 搜索功能返回空结果，高亮功能不生效 |
| 🔴 **原因** | 旧索引删除后，新索引没有正确配置 IK 分词器，字段类型未指定为 `text` |
| 🟢 **解决** | 创建 `ElasticsearchIndexInitializer` 类，在应用启动时自动创建索引 |

```java
elasticsearchClient.indices().create(c -> c
    .index("article")
    .mappings(m -> m
        .properties("id", p -> p.integer(i -> i.store(true)))
        .properties("articleTitle", p -> p.text(t -> t
            .analyzer("ik_max_word")
            .searchAnalyzer("ik_smart")
            .store(true)
        ))
        .properties("articleContent", p -> p.text(t -> t
            .analyzer("ik_max_word")
            .searchAnalyzer("ik_smart")
            .store(true)
        ))
        .properties("isDelete", p -> p.integer(i -> i.store(true)))
        .properties("status", p -> p.integer(i -> i.store(true)))
    )
);
```

> **🔑 关键配置**：
> - `articleTitle` 和 `articleContent` 字段必须使用 `text` 类型
> - 分词器配置：`analyzer: ik_max_word`（索引时），`searchAnalyzer: ik_smart`（搜索时）

---

### ❌ 问题 2: ES 字段映射错误

| 类别 | 内容 |
|:---|:---|
| 🔴 **现象** | 搜索返回结果，但文章 ID 为 null |

**错误代码** ❌：

```java
// ❌ 错误：ES 中的字段名是 id，不是 articleId
article.setId(jsonObj.getInteger("articleId"));
```

**正确代码** ✅：

```java
// ✅ 正确：ES 中的字段名是 id
article.setId(jsonObj.getInteger("id"));
```

> **💡 教训**：ES 索引中的字段名是 `id`，不是 `articleId`

---

### ❌ 问题 3: JsonData.toJson() 类型转换错误

| 类别 | 内容 |
|:---|:---|
| 🔴 **现象** | 编译错误 `不兼容的类型。实际为 jakarta.json.JsonValue'，需要 'java.lang.String'` |

**错误代码** ❌：

```java
// ❌ 错误：toJson() 返回的是 JsonValue 对象
String rawJson = sourceData.toJson();
```

**正确代码** ✅：

```java
// ✅ 正确：需要调用 toString() 转换为字符串
String rawJson = sourceData.toJson().toString();
```

> **💡 原因**：ES 8.x 的 `JsonData.toJson()` 返回的是 `JsonValue` 对象

---

### ❌ 问题 4: ES 索引中没有数据

| 类别 | 内容 |
|:---|:---|
| 🔴 **现象** | 索引存在且配置正确，搜索返回空数组 |
| 🔴 **原因** | 删除索引后，新索引创建成功但没有文章数据，MaxWell 只同步增量变更 |
| 🟢 **解决** | 在 `ElasticsearchIndexInitializer` 中添加数据同步方法 |

**同步结果**：

```
========== 文章数据同步完成 ==========
应同步文章数：27
实际同步成功：27
同步失败：0
======================================
```

---

### ❌ 问题 5: 搜索查询未使用 IK 分词器

| 类别 | 内容 |
|:---|:---|
| 🔴 **现象** | 中文搜索不准确，分词效果差 |
| 🟢 **解决** | 在 `buildQuery` 方法中明确指定 analyzer |

```java
BoolQuery boolQuery = BoolQuery.of(b -> b
    .must(m -> m.bool(bb -> bb
        .should(s -> s.match(t -> t
            .field("articleTitle")
            .query(keywords)
            .analyzer("ik_max_word")  // ✅ 明确指定分词器
        ))
        .should(s -> s.match(t -> t
            .field("articleContent")
            .query(keywords)
            .analyzer("ik_max_word")  // ✅ 明确指定分词器
        ))
    ))
    .must(m -> m.term(t -> t.field("isDelete").value(FALSE)))
    .must(m -> m.term(t -> t.field("status").value(PUBLIC.getStatus())))
);
```

---

### ❌ 问题 6: 高亮配置不生效

| 类别 | 内容 |
|:---|:---|
| 🔴 **现象** | 搜索结果有高亮字段，但值为空 |
| 🔴 **原因** | `preTags` 和 `postTags` 没有在全局设置 |
| 🟢 **解决** | 全局设置高亮标签 |

```java
Highlight highlight = Highlight.of(h -> h
    .preTags(PRE_TAG)  // ✅ 全局设置前置标签
    .postTags(POST_TAG) // ✅ 全局设置后置标签
    .fields("articleTitle", HighlightField.of(hf -> hf
        .fragmentSize(0)
    ))
    .fields("articleContent", HighlightField.of(hf -> hf
        .fragmentSize(50)
        .numberOfFragments(3)
    ))
);
```

---

### ❌ 问题 7: ES Client 超时配置（SB 4.0 + ES 8.17.2）

| 类别 | 内容 |
|:---|:---|
| 🔴 **现象** | `30,000 milliseconds timeout on connection http-outgoing-0` |
| 🔴 **原因** | SB 4.0 + ES 8.17.2 需要显式配置超时参数 |
| 🟢 **解决** | 在 `ElasticsearchConfig` 中设置 `RequestConfig` |

```java
RequestConfig requestConfig = RequestConfig.custom()
        .setConnectTimeout(10000)   // 连接超时 10s
        .setSocketTimeout(60000)    // Socket 超时 60s（支持批量同步）
        .setConnectionRequestTimeout(10000)
        .build();

RestClientBuilder builder = RestClient.builder(new HttpHost(hostname, port, "http"))
        .setHttpClientConfigCallback((HttpAsyncClientBuilder httpClientBuilder) -> {
            httpClientBuilder.setDefaultCredentialsProvider(credentialsProvider)
                    .setDefaultRequestConfig(requestConfig)
                    .setMaxConnTotal(100)
                    .setMaxConnPerRoute(30);
            return httpClientBuilder;
        });
```

---

### ❌ 问题 8: ES article 索引 RED 状态

| 类别 | 内容 |
|:---|:---|
| 🔴 **现象** | `primary shard is not active`，同步失败 |
| 🔴 **原因** | IK 分词器插件未安装，索引 mapping 无法解析 |
| 🟢 **解决** | 在 ES 服务器安装 IK 分词器并重建索引 |

```bash
# 1. 安装 IK 分词器（ES 8.17.2）
elasticsearch-plugin install https://github.com/infinilabs/analysis-ik/releases/download/v8.17.2/elasticsearch-analysis-ik-8.17.2.zip

# 2. 重启 ES
docker restart aurora-elasticsearch

# 3. 删除损坏的索引
curl -X DELETE "http://你的ip:9200/article"

# 4. 重启应用，自动重建索引并同步数据
```

---

## 1.3 关键配置文件

### application-prod.yml

```yaml
elasticsearch:
  rest:
    uris: http://你的ip:9200
    username: 用户名
    password: 密码

search:
  mode: elasticsearch
```

### 核心代码文件

| 序号 | 文件 | 说明 |
|:---:|:---|:---|
| 1️⃣ | `ElasticsearchConfig.java` | ES 客户端配置（SB 4.0 需显式设置超时） |
| 2️⃣ | `ElasticsearchIndexInitializer.java` | 索引初始化与数据同步 |
| 3️⃣ | `EsSearchStrategyImpl.java` | ES 搜索策略实现 |
| 4️⃣ | `ArticleController.java` | 搜索 API 接口 |

---

## 1.4 经验总结

| # | 经验说明 |
|:---:|:---|
| 1️⃣ | **字段映射必须匹配**：ES 中的字段名必须与索引时的字段名一致 |
| 2️⃣ | **数据同步机制**：MaxWell 只同步增量变更，删除索引后需要手动全量同步历史数据 |
| 3️⃣ | **分词器配置**：索引时使用 `ik_max_word`，搜索时使用 `ik_smart`，查询时必须明确指定 analyzer |
| 4️⃣ | **高亮配置**：全局设置 `preTags` 和 `postTags`，字段级别可以单独配置 `fragmentSize` 和 `numberOfFragments` |
| 5️⃣ | **SB 4.0 + ES 8.17.2**：必须显式配置超时参数，否则批量同步会失败 |
| 6️⃣ | **IK 分词器版本**：必须与 ES 版本严格一致（如 ES 8.17.2 对应 IK 8.17.2） |

---

# 2. 常用命令与运维

## 2.1 ES 常用命令

### 集群与索引管理

```bash
# 📊 检查集群健康
curl -u 用户名:密码 http://你的ip:9200/_cluster/health?pretty

# 📋 检查索引列表
curl -u 用户名:密码 http://你的ip:9200/_cat/indices?v

# 🔍 检查分片状态
curl -u 用户名:密码 http://你的ip:9200/_cat/shards/article?v
```

### 数据操作

```bash
# 🗑️ 删除索引
curl -X DELETE -u 用户名:密码 http://你的ip:9200/article

# 📄 查看索引映射
curl -u 用户名:密码 http://your-ip:9200/article/_mapping?pretty

# 🔎 测试搜索
curl -X POST -u 用户名:密码 \
  http://你的ip:9200/article/_search?pretty \
  -H 'Content-Type: application/json' \
  -d '{"query":{"match_all":{}}, "size":1}'
```

### 容器运维

```bash
# 📜 查看 ES 日志
docker logs aurora-elasticsearch --tail 100 -f

# 🔄 重启 ES 服务
docker restart aurora-elasticsearch

# 💾 查看 ES 资源使用
docker stats aurora-elasticsearch --no-stream
```

---

## 2.2 搜索模式切换

| 模式 | 配置值 | 说明 |
|:---|:---|:---|
| 🔵 MySQL 搜索 | `search.mode: mysql` | 使用 MySQL LIKE 查询 |
| 🟢 Elasticsearch | `search.mode: elasticsearch` | 使用 ES 全文搜索 |

**重启应用**：

```bash
docker restart aurora-springboot
```

---

## 2.3 应用重启流程

| 步骤 | 操作 | 命令 |
|:---:|:---|:---|
| 1️⃣ | 停止应用 | `docker stop aurora-springboot` |
| 2️⃣ | 更新配置 | 修改 `application-prod.yml`，上传新的 jar 包到 `/opt/aurora/app/` |
| 3️⃣ | 启动应用 | `docker start aurora-springboot` |
| 4️⃣ | 检查日志 | `docker logs aurora-springboot -f` |
| 5️⃣ | 验证健康 | `curl http://localhost:8080/api/articles/topAndFeatured` |

---

## 2.4 常见问题快速解决

### ❌ factoryBeanObjectType 错误

| 类别 | 内容 |
|:---|:---|
| 🔴 **现象** | `Invalid value value type for attribute 'factoryBeanObjectType': java.lang.String` |
| 🟢 **解决** | 使用 `mybatis-plus-spring-boot3-starter` 替代 `mybatis-plus-boot-starter` |

---

### ❌ 端口被占用

| 类别 | 内容 |
|:---|:---|
| 🔴 **现象** | `Port 8080 was already in use` |
| 🟢 **解决** | 查看端口占用并释放，或修改端口为 8081 |

**Windows 解决方案**：

```powershell
# 查看端口占用
netstat -ano | findstr :8080

# 终止占用进程
powershell -Command "Stop-Process -Id <PID> -Force"
```

---

# 3. 参考资料

## 3.1 官方文档

| 文档类型 | 文档名称 | 链接 |
|:---|:---|:---|
| 🔍 | Elasticsearch 8.17.x 官方文档 | [查看](https://www.elastic.co/guide/en/elasticsearch/reference/8.17/index.html) |
| 📝 | IK 分词器文档 | [查看](https://github.com/medcl/elasticsearch-analysis-ik) |
| ☕ | Elasticsearch 8.x Java Client | [查看](https://www.elastic.co/guide/en/elasticsearch/client/java-api-client/current/index.html) |

## 3.2 相关文档

- [升级到 Spring Boot 4.0.5 和 JDK 25 历程](./升级到SpringBoot4.0.5和JDK25历程.md) - Spring Boot 4.0.5 升级详细指南

---

## 文档修订历史

| 版本 | 日期 | 修改内容 | 作者 |
|:---|:---:|:---|:---|
| v1.0 | 2026-03-09 | 初始版本：Elasticsearch 8 集成问题排查记录 | Aurora Team |
| v2.0 | 2026-04-08 | 更新 ES 到 8.17.2，新增 SB 4.0 兼容性说明 | Aurora Team |

---

<div align="center">

**文档版本**: v2.0 (ES 8.17.2) | **最后更新**: 2026-04-08 | **维护者**: Aurora Team

![Built with ❤️](https://img.shields.io/badge/Built%20with-%E2%9D%A4%EF%B8%8F-blue?style=flat-square)

</div>
