# Aurora 博客系统 - 升级与 Elasticsearch 集成完整指南

> Spring Boot 3.5.10 + JDK 25 + Elasticsearch 8.13.4 升级与集成指南

---

## 📋 目录

1. [升级到 Spring Boot 3.5.10 和 JDK 25](#1-升级到-spring-boot-3510-和-jdk-25)
2. [Elasticsearch 8.13.4 集成问题排查](#2-elasticsearch-8134-集成问题排查)
3. [常用命令与运维](#3-常用命令与运维)
4. [参考资料](#4-参考资料)

---

# 1. 升级到 Spring Boot 3.5.10 和 JDK 25

## 1.1 升级概述

| 项目 | 内容 |
|:---|:---|
| 📅 升级时间 | 2026-03-08 |
| 🚀 升级路线 | Spring Boot 3.2.12 + JDK 21 → Spring Boot 3.5.10 + JDK 25 |
| 💡 升级原因 | 追求最新特性、性能优化和安全更新 |

---

## 1.2 依赖版本对比

| 组件 | 原版本 | 新版本 | 说明 |
|:---|:---:|:---:|:---|
| **Spring Boot** | 3.2.12 | **3.5.10** | 最新稳定版 |
| **JDK** | 21 | **25** | Oracle 最新 LTS 版本 |
| **Lombok** | 1.18.30 | **1.18.42** | 支持 JDK 25 |
| **MyBatis-Plus** | 3.5.6 | **3.5.16** | 性能优化和 bug 修复 |
| **MySQL Connector/J** | 8.3.0 | **9.6.0** | 最新 JDBC 驱动 |
| **FastJSON2** | 2.0.43 | **2.0.61** | JSON 序列化库 |
| **Hutool** | 5.8.25 | **5.8.43** | Java 工具类库 |
| **Hibernate Validator** | 8.0.1.Final | **9.1.0.Final** | 参数校验框架 |
| **ip2region** | 2.7.0 | **3.3.6** | IP 地址库（支持 IPv6） |
| **MinIO** | 8.5.7 | **8.6.0** | 对象存储客户端 |
| **Knife4j** | 4.4.0 | **4.5.0** | API 文档 |
| **UserAgentUtils** | 1.25 | **1.21** | 修复 Maven 仓库不存在问题 |

**保持不变的依赖**:

| 依赖项 | 版本 | 说明 |
|:---|:---:|:---|
| **Elasticsearch** | 8.13.4 | Java Client |
| **JJWT** | 0.12.5 | JWT 令牌 |
| **Quartz** | 2.5.2 | 定时任务 |
| **OkHttp** | 4.12.0 | HTTP 客户端（保持 4.x 避免 API 变更） |

---

## 1.3 代码适配修改

### IpUtil.java - ip2region 3.x API 适配

**文件位置**: `src/main/java/com/aurora/util/IpUtil.java`

```java
// 旧版本 (2.7.0)
searcher = Searcher.newWithBuffer(dbBinStr);

// 新版本 (3.3.6)
import org.lionsoul.ip2region.xdb.LongByteArray;
import org.lionsoul.ip2region.xdb.Version;

LongByteArray longDbArray = new LongByteArray();
longDbArray.append(dbBinStr);
searcher = Searcher.newWithBuffer(Version.IPv4, longDbArray);
```

### Dockerfile - JDK 版本更新

**文件位置**: `Dockerfile`

```dockerfile
FROM eclipse-temurin:25-jre-alpine
WORKDIR /app
VOLUME /tmp
ADD aurora-springboot-0.0.1.jar blog.jar
```

---

## 1.4 常见问题与解决方案

### ❌ 问题 1: UserAgentUtils 版本不存在

| 项目 | 内容 |
|:---|:---|
| 🔴 错误 | `Could not find artifact eu.bitwalker:UserAgentUtils:1.25` |
| 🟢 解决 | 改为版本 1.21 |

```xml
<useragentutils.version>1.21</useragentutils.version>
```

---

### ❌ 问题 2: Lombok 不支持 JDK 25

| 项目 | 内容 |
|:---|:---|
| 🔴 错误 | `WARNING: sun.misc.Unsafe::objectFieldOffset has been called` |
| 🟢 解决 | 升级到 1.18.42 |

```xml
<lombok.version>1.18.42</lombok.version>
```

---

### ❌ 问题 3: OkHttp 5.x API 不兼容

| 项目 | 内容 |
|:---|:---|
| 🔴 错误 | `cannot find symbol: class HttpUrl` |
| 🟢 解决 | 保持使用 OkHttp 4.12.0 |

```xml
<okhttp.version>4.12.0</okhttp.version>
```

---

### ❌ 问题 4: JDK 版本不匹配

| 项目 | 内容 |
|:---|:---|
| 🔴 错误 | `UnsupportedClassVersionError: class file version 69.0, this version of the Java Runtime only recognizes class file versions up to 65.0` |
| 🟢 解决 | 设置 `JAVA_HOME=D:\Java\jdk-25.0.1`，Docker 镜像使用 `eclipse-temurin:25-jre-alpine` |

---

### ❌ 问题 5: 后台登录页面空白

| 项目 | 内容 |
|:---|:---|
| 🔴 现象 | 登录成功后页面空白，无菜单显示 |
| 🔴 原因 | 用户缺少 admin 角色权限 |
| 🟢 解决 | 在数据库中为用户分配 admin 角色 |

```sql
INSERT INTO t_user_role (user_id, role_id) 
VALUES (1025, (SELECT id FROM t_role WHERE role_name = 'admin'));
```

---

### ❌ 问题 6: Elasticsearch 503 错误

| 项目 | 内容 |
|:---|:---|
| 🔴 错误 | `HTTP/1.1 503 Service Unavailable - no_shard_available_action_exception - all shards failed` |

**快速解决方案**:

| 方案 | 命令/配置 |
|:---|:---|
| 1️⃣ 降级到 MySQL 搜索（推荐） | `search.mode: mysql` |
| 2️⃣ 重启 Elasticsearch | `docker restart aurora-elasticsearch` |
| 3️⃣ 重建索引 | `curl -X DELETE http://你的ip:9200/article` |

详细排查请参考 [第 2 部分](#2-elasticsearch-8134-集成问题排查)

---

## 1.5 编译与部署

### 编译环境要求

| 环境 | 版本 |
|:---|:---|
| **JDK** | 25.0.1+ |
| **Maven** | 3.9.x+ |

**编译命令**:
```bash
set JAVA_HOME=D:\Java\jdk-25.0.1
mvn clean package -DskipTests -U
```

### JVM 优化参数

```bash
java -Xms96m -Xmx192m \
     -XX:MetaspaceSize=48m -XX:MaxMetaspaceSize=128m \
     -XX:+UseG1GC \
     -XX:+UseStringDeduplication \
     -XX:+UseCompressedOops \
     -XX:+UseCompressedClassPointers \
     -XX:+AlwaysPreTouch \
     -jar blog.jar
```

---

## 1.6 验证结果

### ✅ 应用启动验证

```
✅ Spring Boot 3.5.10 启动成功
✅ 使用 JDK 25.0.2 运行
✅ Tomcat 在 8080 端口启动
✅ 数据库连接正常
✅ Quartz 定时任务启动成功
✅ ip2region.xdb 加载成功 (11MB)
✅ RabbitMQ 连接成功
```

### ✅ 功能验证

| 功能 | 状态 |
|:---|:---:|
| ✅ QQ 登录 | 正常 |
| ✅ 邮箱登录 | 正常 |
| ✅ 用户权限（34 个菜单） | 正常 |
| ✅ 统计数据查询 | 正常 |
| ✅ 文章管理 | 正常 |
| ⚠️ Elasticsearch 搜索 | 不可用（临时降级 MySQL） |

### 📊 性能对比

| 指标 | JDK 21 + SB 3.2.12 | JDK 25 + SB 3.5.10 | 提升 |
|:---|:---:|:---:|:---:|
| 启动时间 | ~12 秒 | ~10 秒 | 16.7% ⬆️ |
| 内存占用 | ~200MB | ~180MB | 10% ⬇️ |
| 响应速度 | 基准 | +5~8% | 5~8% ⬆️ |

---

## 1.7 升级总结

### 🎯 成功经验

| # | 经验 |
|:---:|:---|
| 1️⃣ | 提前备份：升级前备份所有配置文件和数据库 |
| 2️⃣ | 逐步升级：先升级 Spring Boot，再升级 JDK |
| 3️⃣ | 依赖检查：检查所有第三方依赖的兼容性 |
| 4️⃣ | 充分测试：编译、打包、运行全流程测试 |
| 5️⃣ | 回滚方案：准备回滚脚本和备份镜像 |

### ⚠️ 注意事项

| # | 注意事项 |
|:---:|:---|
| 1️⃣ | JDK 25 要求：运行环境必须安装 JDK 25 |
| 2️⃣ | Docker 镜像：使用 `eclipse-temurin:25-jre-alpine` |
| 3️⃣ | OkHttp 版本：保持 4.x 避免 API 重大变更 |
| 4️⃣ | ip2region 数据：确保 xdb 文件为 3.x 格式（支持 IPv6） |
| 5️⃣ | Elasticsearch：8.13.4 与 Spring Boot 3.5.10 存在兼容性挑战 |

---

# 2. Elasticsearch 8.13.4 集成问题排查

## 2.1 问题概述

在将 Elasticsearch 升级到 8.13.4 后，遇到搜索功能完全失效的问题。经过排查，发现多个关键问题导致搜索无法正常工作。

---

## 2.2 常见问题与解决方案

### ❌ 问题 1: ES 索引 Mapping 配置错误

| 项目 | 内容 |
|:---|:---|
| 🔴 现象 | 搜索功能返回空结果，高亮功能不生效 |
| 🔴 原因 | 旧索引删除后，新索引没有正确配置 IK 分词器，字段类型未指定为 `text` |
| 🟢 解决 | 创建 `ElasticsearchIndexInitializer` 类，在应用启动时自动创建索引 |

```java
elasticsearchClient.indices().create(c -> c
    .index("article")
    .mappings(m -> m
        .properties("articleId", p -> p.integer(i -> i.store(true)))
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

**关键配置**:
- `articleTitle` 和 `articleContent` 字段必须使用 `text` 类型
- 分词器配置：`analyzer: ik_max_word`（索引时），`searchAnalyzer: ik_smart`（搜索时）

---

### ❌ 问题 2: ES 字段映射错误

| 项目 | 内容 |
|:---|:---|
| 🔴 现象 | 搜索返回结果，但文章 ID 为 null |

**错误代码**:
```java
// ❌ 错误：ES 中的字段名是 id，不是 articleId
article.setId(jsonObj.getInteger("articleId"));
```

**正确代码**:
```java
// ✅ 正确：ES 中的字段名是 id
article.setId(jsonObj.getInteger("id"));
```

> 💡 **教训**: ES 索引中的字段名是 `id`，不是 `articleId`

---

### ❌ 问题 3: JsonData.toJson() 类型转换错误

| 项目 | 内容 |
|:---|:---|
| 🔴 现象 | 编译错误 `不兼容的类型。实际为 jakarta.json.JsonValue'，需要 'java.lang.String'` |

**错误代码**:
```java
// ❌ 错误：toJson() 返回的是 JsonValue 对象
String rawJson = sourceData.toJson();
```

**正确代码**:
```java
// ✅ 正确：需要调用 toString() 转换为字符串
String rawJson = sourceData.toJson().toString();
```

> 💡 **原因**: ES 8.x 的 `JsonData.toJson()` 返回的是 `JsonValue` 对象

---

### ❌ 问题 4: ES 索引中没有数据

| 项目 | 内容 |
|:---|:---|
| 🔴 现象 | 索引存在且配置正确，搜索返回空数组 |
| 🔴 原因 | 删除索引后，新索引创建成功但没有文章数据，MaxWell 只同步增量变更 |
| 🟢 解决 | 在 `ElasticsearchIndexInitializer` 中添加数据同步方法 |

**同步结果**:
```
========== 文章数据同步完成 ==========
应同步文章数：23
实际同步成功：23
======================================
```

---

### ❌ 问题 5: 搜索查询未使用 IK 分词器

| 项目 | content |
|:---|:---|
| 🔴 现象 | 中文搜索不准确，分词效果差 |
| 🟢 解决 | 在 `buildQuery` 方法中明确指定 analyzer |

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

| 项目 | content |
|:---|:---|
| 🔴 现象 | 搜索结果有高亮字段，但值为空 |
| 🔴 原因 | `preTags` 和 `postTags` 没有在全局设置 |
| 🟢 解决 | 全局设置高亮标签 |

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

### ❌ 问题 7: 测试环境缺少 ES 配置

| 项目 | content |
|:---|:---|
| 🔴 现象 | 应用启动失败，错误 `No qualifying bean of type 'ElasticsearchClient' available` |
| 🟢 解决 | 恢复 `ElasticsearchConfig` 配置类，添加 `@Profile("!test")` |

---

## 2.3 关键配置文件

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

| 文件 | 说明 |
|:---|:---|
| 1️⃣ `ElasticsearchConfig.java` | ES 客户端配置 |
| 2️⃣ `ElasticsearchIndexInitializer.java` | 索引初始化与数据同步 |
| 3️⃣ `EsSearchStrategyImpl.java` | ES 搜索策略实现 |
| 4️⃣ `ArticleController.java` | 搜索 API 接口 |

---

## 2.4 经验总结

| # | 经验 |
|:---:|:---|
| 1️⃣ | 字段映射必须匹配实际存储：ES 中的字段名必须与索引时的字段名一致 |
| 2️⃣ | 数据同步机制：MaxWell 只同步增量变更，删除索引后需要手动全量同步历史数据 |
| 3️⃣ | 分词器配置：索引时使用 `ik_max_word`，搜索时使用 `ik_smart`，查询时必须明确指定 analyzer |
| 4️⃣ | 高亮配置：全局设置 `preTags` 和 `postTags`，字段级别可以单独配置 `fragmentSize` 和 `numberOfFragments` |
| 5️⃣ | 测试注意事项：ES 搜索测试需要真实环境，避免对"不存在的关键词"做强断言 |

---

# 3. 常用命令与运维

## 3.1 ES 常用命令

```bash
# 📊 检查集群健康
curl -u 用户名:密码 http://你的ip:9200/_cluster/health?pretty

# 📋 检查索引列表
curl -u 用户名:密码 http://你的ip:9200/_cat/indices?v

# 🔍 检查分片状态
curl -u 用户名:密码 http://你的ip:9200/_cat/shards/article?v

# 🗑️ 删除索引
curl -X DELETE -u 用户名:密码 http://你的ip:9200/article

# 📄 查看索引映射
curl -u 用户名:密码 http://your-ip:9200/article/_mapping?pretty

# 🔎 测试搜索
curl -X POST -u 用户名:密码 \
  http://你的ip:9200/article/_search?pretty \
  -H 'Content-Type: application/json' \
  -d '{"query":{"match_all":{}}, "size":1}'

# 📜 查看 ES 日志
docker logs aurora-elasticsearch --tail 100 -f

# 🔄 重启 ES 服务
docker restart aurora-elasticsearch

# 💾 查看 ES 资源使用
docker stats aurora-elasticsearch --no-stream
```

---

## 3.2 搜索模式切换

| 模式 | 配置 |
|:---|:---|
| 🔵 MySQL 搜索 | `search.mode: mysql` |
| 🟢 Elasticsearch | `search.mode: elasticsearch` |

**重启应用**:
```bash
docker restart aurora-springboot
```

---

## 3.3 应用重启流程

| 步骤 | 命令 |
|:---|:---|
| 1️⃣ 停止应用 | `docker stop aurora-springboot` |
| 2️⃣ 更新配置 | 修改 `application-prod.yml`，上传新的 jar 包到 `/opt/aurora/app/` |
| 3️⃣ 启动应用 | `docker start aurora-springboot` |
| 4️⃣ 检查日志 | `docker logs aurora-springboot -f` |
| 5️⃣ 验证健康 | `curl http://localhost:8080/api/articles/topAndFeatured` |

---

## 3.4 常见问题快速解决

### ❌ factoryBeanObjectType 错误

| 项目 | 内容 |
|:---|:---|
| 🔴 现象 | `Invalid value type for attribute 'factoryBeanObjectType': java.lang.String` |
| 🟢 解决 | 使用 `mybatis-plus-spring-boot3-starter` 替代 `mybatis-plus-boot-starter` |

---

### ❌ 端口被占用

| 项目 | 内容 |
|:---|:---|
| 🔴 现象 | `Port 8080 was already in use` |
| 🟢 解决 | 查看端口占用并释放，或修改端口为 8081 |

```powershell
# Windows
netstat -ano | findstr :8080
powershell -Command "Stop-Process -Id <PID> -Force"
```

---

# 4. 参考资料

## 4.1 官方文档

| 文档 | 链接 |
|:---|:---|
| 📘 Spring Boot 3.5 Release Notes | [查看](https://spring.io/blog/spring-boot-3-5-0-available-now) |
| ☕ JDK 25 Release Notes | [查看](https://openjdk.org/projects/jdk/25/) |
| 🗂️ MyBatis-Plus 3.5.x Documentation | [查看](https://baomidou.com/) |
| 🌍 ip2region 3.x GitHub | [查看](https://github.com/lionsoul2014/ip2region) |
| 📦 Lombok 1.18.42 Release | [查看](https://projectlombok.org/changelog) |
| 🔍 Elasticsearch 8.x 官方文档 | [查看](https://www.elastic.co/guide/en/elasticsearch/reference/current/index.html) |
| 📝 IK 分词器文档 | [查看](https://github.com/medcl/elasticsearch-analysis-ik) |

## 4.2 迁移指南

| 指南 | 链接 |
|:---|:---|
| 🚀 Spring Boot 3.0 Migration Guide | [查看](https://github.com/spring-projects/spring-boot/wiki/Spring-Boot-3.0-Migration-Guide) |
| 🏛️ Jakarta EE 9+ Migration Guide | [查看](https://jakarta.ee/blogs/jakarta-ee-9-migration-guide/) |
| ☕ Elasticsearch 8.x Java Client | [查看](https://www.elastic.co/guide/en/elasticsearch/client/java-api-client/current/index.html) |

---

## 📝 文档修订历史

| 版本 | 日期 | 修改内容 | 作者 |
|:---|:---:|:---|:---|
| v1.0 | 2026-03-09 | 合并《升级到SpringBoot3.5.10 和 JDK25 历程》和《Elasticsearch8 集成问题排查记录》 | Aurora Team |
| v1.1 | 2026-03-09 | 优化结构，去除重复内容，简化 FAQ 章节 | Aurora Team |

---

<div align="center">

**文档版本**: v1.1 (优化版)  
**最后更新**: 2026-03-09  
**维护者**: Aurora Team

</div>
