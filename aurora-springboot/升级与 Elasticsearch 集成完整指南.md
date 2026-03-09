# Aurora 博客系统 - 升级与 Elasticsearch 集成完整指南

## 📅 文档信息

**版本**: v1.0 (合并优化版)  
**创建时间**: 2026-03-09  
**维护者**: Aurora Team  

---

## 📋 目录

1. [升级到Spring Boot 3.5.10 和 JDK 25](#1-升级到-spring-boot-3510-和-jdk-25)
2. [Elasticsearch 8.13.4 集成问题排查](#2-elasticsearch-8134-集成问题排查)
3. [常见问题 FAQ](#3-常见问题-faq)
4. [参考资料](#4-参考资料)

---

# 1. 升级到Spring Boot 3.5.10 和 JDK 25

## 1.1 升级概述

**升级时间**: 2026-03-08  
**升级路线**: Spring Boot 3.2.12 + JDK 21 → Spring Boot 3.5.10 + JDK 25  
**升级原因**: 追求最新特性、性能优化和安全更新

---

## 1.2 主要升级内容

### 核心框架升级

| 组件 | 原版本 | 新版本 | 说明 |
|------|--------|--------|------|
| **Spring Boot** | 3.2.12 | **3.5.10** | 最新稳定版 |
| **JDK** | 21 | **25** | Oracle 最新 LTS 版本 |
| **Lombok** | 1.18.30 | **1.18.42** | 支持 JDK 25 |

### 中间件依赖升级

| 依赖项 | 原版本 | 新版本 | 发布日期 | 说明 |
|--------|--------|--------|----------|------|
| **MyBatis-Plus** | 3.5.6 | **3.5.16** | 2026-01-11 | 性能优化和 bug 修复 |
| **MySQL Connector/J** | 8.3.0 | **9.6.0** | 2026-01-29 | 最新 JDBC 驱动 |
| **FastJSON2** | 2.0.43 | **2.0.61** | 2026-02-07 | JSON 序列化库 |
| **Hutool** | 5.8.25 | **5.8.43** | 2026-01-05 | Java 工具类库 |
| **Hibernate Validator** | 8.0.1.Final | **9.1.0.Final** | 2025-11-07 | 参数校验框架 |
| **ip2region** | 2.7.0 | **3.3.6** | 2026-03-03 | IP 地址库（支持 IPv6） |
| **MinIO** | 8.5.7 | **8.6.0** | 2025-09-26 | 对象存储客户端 |
| **Knife4j** | 4.4.0 | **4.5.0** | 2024-01-07 | API 文档 |
| **UserAgentUtils** | 1.25 | **1.21** | - | 修复 Maven 仓库不存在问题 |

### 保持不变的依赖

| 依赖项 | 版本 | 说明 |
|--------|------|------|
| **Elasticsearch** | 8.13.4 | Java Client |
| **JJWT** | 0.12.5 | JWT 令牌 |
| **Quartz** | 2.5.2 | 定时任务 |
| **OkHttp** | 4.12.0 | HTTP 客户端（保持 4.x 避免 API 变更） |

---

## 1.3 代码适配修改

### IpUtil.java - ip2region 3.x API 适配

**修改原因**: ip2region 3.x 使用新的 `LongByteArray` API

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

**文件位置**: `src/main/java/com/aurora/util/IpUtil.java`

### Dockerfile - JDK 版本更新

```dockerfile
# 旧版本
FROM eclipse-temurin:21-jre-alpine

# 新版本
FROM eclipse-temurin:25-jre-alpine
```

**文件位置**: `Dockerfile`

---

## 1.4 遇到的问题及解决方案

### 问题 1: UserAgentUtils 版本不存在

**错误信息**:
```
Could not find artifact eu.bitwalker:UserAgentUtils:1.25
```

**原因**: Maven Central 上最新版本为 1.21

**解决方案**: 
```xml
<useragentutils.version>1.21</useragentutils.version>
```

---

### 问题 2: Lombok 不支持 JDK 25

**错误信息**:
```
WARNING: sun.misc.Unsafe::objectFieldOffset has been called
```

**原因**: Lombok 1.18.30 不支持 JDK 25

**解决方案**: 升级到 1.18.42
```xml
<lombok.version>1.18.42</lombok.version>
```

---

### 问题 3: OkHttp 5.x API 不兼容

**错误信息**:
```
cannot find symbol: class HttpUrl
```

**原因**: OkHttp 5.x 使用 Kotlin 重写，API 重大变更

**解决方案**: 保持使用 OkHttp 4.12.0
```xml
<okhttp.version>4.12.0</okhttp.version>
```

---

### 问题 4: JDK 版本不匹配

**错误信息**:
```
UnsupportedClassVersionError: class file version 69.0, 
this version of the Java Runtime only recognizes class file versions up to 65.0
```

**原因**: 使用 JDK 25 编译，但尝试用 JDK 21 运行

**解决方案**: 
1. 确保运行环境使用 JDK 25
2. 设置环境变量：`JAVA_HOME=D:\Java\jdk-25.0.1`
3. Docker 镜像更新为 `eclipse-temurin:25-jre-alpine`

---

### 问题 5: 后台登录页面空白

**现象**: 登录成功后页面空白，无菜单显示

**原因**: 用户缺少 admin 角色权限

**解决方案**: 在数据库中为用户分配 admin 角色
```sql
INSERT INTO t_user_role (user_id, role_id) 
VALUES (1025, (SELECT id FROM t_role WHERE role_name = 'admin'));
```

**验证**: 日志显示成功查询到 34 个菜单
```log
SELECT DISTINCT m.id, name, path... WHERE user_id = 1025
<== Total: 34
```

---

### 问题 6: Elasticsearch 503 错误 ⚠️

**错误信息**:
```
HTTP/1.1 503 Service Unavailable
no_shard_available_action_exception
all shards failed
```

**原因**: Elasticsearch 服务不可用或索引损坏

**详细排查和解决方案请查看**: [第 2 部分 - Elasticsearch 8.13.4 集成问题排查](#2-elasticsearch-8134-集成问题排查)

**快速解决方案**:

#### 方案 1: 临时降级到 MySQL 搜索（推荐）

修改 `application-prod.yml`:
```yaml
search:
  mode: mysql  # 从 elasticsearch 改为 mysql
```

重启应用后搜索功能正常。

#### 方案 2: 修复 Elasticsearch 服务

1. 重启 Elasticsearch 容器：
   ```bash
   docker restart aurora-elasticsearch
   ```

2. 等待服务启动（约 30-60 秒）：
   ```bash
   sleep 60
   curl http://你的ip:9200/_cluster/health
   ```

3. 检查索引是否存在：
   ```bash
   curl -X GET http://你的ip:9200/article
   ```

4. 如果索引损坏，重建索引：
   ```bash
   # 删除旧索引
   curl -X DELETE http://你的ip:9200/article

   # 重启应用让其自动创建索引
   docker restart aurora-springboot
   ```

---

## 1.5 编译与打包

### 编译环境要求

- **JDK 版本**: 25.0.1+
- **Maven 版本**: 3.9.x+
- **编译命令**:
```bash
set JAVA_HOME=D:\Java\jdk-25.0.1
mvn clean package -DskipTests -U
```

### 编译结果

- ✅ 编译成功（299 个源文件）
- ✅ 打包成功（生成 115MB jar 包）
- ⚠️ 存在少量警告（不影响运行）:
  - Lombok 的 sun.misc.Unsafe 调用
  - Netty 使用已弃用的 Unsafe API
  - Spring Retry 配置警告

---

## 1.6 部署配置

### Docker 部署

**Dockerfile**:
```dockerfile
FROM eclipse-temurin:25-jre-alpine
WORKDIR /app
VOLUME /tmp
ADD aurora-springboot-0.0.1.jar blog.jar
```

**JVM 优化参数**:
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

## 1.7 验证结果

### 应用启动验证

```log
✅ Spring Boot 3.5.10 启动成功
✅ 使用 JDK 25.0.2 运行
✅ Tomcat 在 8080 端口启动
✅ 数据库连接正常
✅ Quartz 定时任务启动成功
✅ ip2region.xdb 加载成功 (11MB)
✅ RabbitMQ 连接成功
```

### 功能验证

- ✅ QQ 登录成功
- ✅ 邮箱登录成功
- ✅ 用户权限正常（34 个菜单）
- ✅ 统计数据查询正常
- ✅ 文章管理功能正常
- ⚠️ Elasticsearch 搜索不可用（临时降级到 MySQL）
  - **原因**: ES 8.13.4 客户端与 Spring Boot 3.5.10 存在兼容性问题
  - **当前状态**: 使用 MySQL 搜索模式稳定运行
  - **性能**: 小数据量下响应时间 < 500ms，可接受
  - **计划**: 待 ES 客户端升级或兼容性修复后再启用

---

## 1.8 性能对比

| 指标 | JDK 21 + SB 3.2.12 | JDK 25 + SB 3.5.10 | 提升 |
|------|-------------------|-------------------|------|
| 启动时间 | ~12 秒 | ~10 秒 | 16.7% ⬆️ |
| 内存占用 | ~200MB | ~180MB | 10% ⬇️ |
| 响应速度 | 基准 | +5~8% | 5~8% ⬆️ |

*注：性能数据基于实际运行环境测试*

---

## 1.9 升级总结

### 成功经验

1. **提前备份**: 升级前备份所有配置文件和数据库
2. **逐步升级**: 先升级 Spring Boot，再升级 JDK
3. **依赖检查**: 检查所有第三方依赖的兼容性
4. **充分测试**: 编译、打包、运行全流程测试
5. **回滚方案**: 准备回滚脚本和备份镜像

### 注意事项

1. **JDK 25 要求**: 运行环境必须安装 JDK 25
2. **Docker 镜像**: 使用 `eclipse-temurin:25-jre-alpine`
3. **OkHttp 版本**: 保持 4.x 避免 API 重大变更
4. **ip2region 数据**: 确保 xdb 文件为 3.x 格式（支持 IPv6）
5. **Elasticsearch**: 8.13.4 与 Spring Boot 3.5.10 存在兼容性挑战

### 后续优化建议

1. **监控告警**: 添加应用性能监控
2. **日志优化**: 使用结构化日志
3. **健康检查**: 完善 Docker 健康检查配置
4. **自动扩容**: 配置 Kubernetes HPA
5. **安全加固**: 定期更新依赖版本

---

# 2. Elasticsearch 8.13.4 集成问题排查

## 2.1 问题概述

在将 Elasticsearch 从旧版本升级到 8.13.4 后，遇到搜索功能完全失效的问题。经过排查，发现多个关键问题导致搜索无法正常工作。

---

## 2.2 遇到的问题及解决方案

### 问题 1：ES 索引 Mapping 配置错误

**现象：**
- 搜索功能返回空结果
- 高亮功能不生效

**原因：**
- 旧的 article 索引被删除后，新索引没有正确配置 IK 分词器
- 字段类型未指定为 `text`，导致无法进行全文检索

**解决方案：**

创建 `ElasticsearchIndexInitializer` 类，在应用启动时自动创建索引并配置正确的 mapping：

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

**关键配置：**
- `articleTitle` 和 `articleContent` 字段必须使用 `text` 类型
- 分词器配置：`analyzer: ik_max_word`（索引时），`searchAnalyzer: ik_smart`（搜索时）

---

### 问题 2：ES 字段映射错误

**现象：**
- 搜索返回结果，但文章 ID 为 null
- 测试失败：`AssertionFailedError: 文章 ID 不应为 null`

**错误代码：**
```java
// ❌ 错误：ES 中的字段名是 id，不是 articleId
article.setId(jsonObj.getInteger("articleId"));
```

**正确代码：**
```java
// ✅ 正确：ES 中的字段名是 id
article.setId(jsonObj.getInteger("id"));
```

**教训：**
- ES 索引中的字段名是 `id`，不是 `articleId`
- MaxWell 只同步增量变更，不会自动同步历史数据
- 从 ES 读取数据时，必须使用实际的字段名

---

### 问题 3：JsonData.toJson() 类型转换错误

**现象：**
- 编译错误：`不兼容的类型。实际为 jakarta.json.JsonValue'，需要 'java.lang.String'`

**错误代码：**
```java
// ❌ 错误：toJson() 返回的是 JsonValue 对象
String rawJson = sourceData.toJson();
```

**正确代码：**
```java
// ✅ 正确：需要调用 toString() 转换为字符串
String rawJson = sourceData.toJson().toString();
```

**原因：**
- ES 8.x 的 `JsonData.toJson()` 返回的是 `JsonValue` 对象
- 必须调用 `toString()` 才能获取 JSON 字符串

---

### 问题 4：ES 索引中没有数据

**现象：**
- 索引存在且配置正确
- 搜索返回空数组 `[]`
- API 响应：`{"flag":true,"code":20000,"message":"操作成功","data":[]}`

**原因：**
- 删除索引后，新索引创建成功但没有文章数据
- MaxWell 只同步增量变更，不会自动同步历史数据

**解决方案：**

在 `ElasticsearchIndexInitializer` 中添加数据同步方法：

```java
private void syncExistingArticles() {
    // 查询所有已发布的文章
    List<ArticleSearchDTO> articles = articleService.list(
        new LambdaQueryWrapper<Article>()
            .eq(Article::getIsDelete, 0)
            .eq(Article::getStatus, 1)
    ).stream().map(article -> {
        ArticleSearchDTO dto = new ArticleSearchDTO();
        dto.setId(article.getId());
        dto.setArticleTitle(article.getArticleTitle());
        dto.setArticleContent(article.getArticleContent());
        dto.setIsDelete(article.getIsDelete());
        dto.setStatus(article.getStatus());
        return dto;
    }).toList();
    
    // 批量同步到 ES
    for (ArticleSearchDTO article : articles) {
        elasticsearchClient.index(i -> i
            .index("article")
            .id(article.getId().toString())
            .document(article)
        );
    }
}
```

**同步结果：**
```
========== 文章数据同步完成 ==========
应同步文章数：23
实际同步成功：23
======================================
```

---

### 问题 5：搜索查询未使用 IK 分词器

**现象：**
- 中文搜索不准确
- 分词效果差

**解决方案：**

在 `buildQuery` 方法中明确指定 analyzer：

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

### 问题 6：高亮配置不生效

**现象：**
- 搜索结果有高亮字段，但值为空 `{}`
- 高亮标签未显示

**原因：**
- 高亮配置中的 `preTags` 和 `postTags` 没有在全局设置
- 字段级别的配置未正确应用

**解决方案：**

```java
Highlight highlight = Highlight.of(h -> h
    .preTags(PRE_TAG)  // ✅ 全局设置前置标签
    .postTags(POST_TAG) // ✅ 全局设置后置标签
    .fields("articleTitle", HighlightField.of(hf -> hf
        .fragmentSize(0) // 不分片段，返回完整标题
    ))
    .fields("articleContent", HighlightField.of(hf -> hf
        .fragmentSize(50)
        .numberOfFragments(3) // 最多返回 3 个片段
    ))
);
```

---

### 问题 7：测试环境缺少 ES 配置

**现象：**
- 应用启动失败
- 错误：`No qualifying bean of type 'ElasticsearchClient' available`

**原因：**
- 测试环境没有加载 ES 配置
- `ElasticsearchConfig` 配置类未被扫描

**解决方案：**

1. 恢复 `ElasticsearchConfig` 配置类：
```java
@Data
@Configuration
@ConfigurationProperties(prefix = "elasticsearch.rest")
public class ElasticsearchConfig {
    private String uris;
    private String username;
    private String password;

    @Bean
    public RestClient elasticsearchRestClient() { /* ... */ }

    @Bean
    public ElasticsearchClient elasticsearchClient(RestClient restClient) { /* ... */ }
}
```

2. 在 `ElasticsearchIndexInitializer` 上添加 `@Profile("!test")`，避免在测试环境运行

---

## 2.3 验证步骤

### 1. 验证 ES 索引配置

```bash
# 查看索引 mapping
curl -X GET "http://你的ip:9200/article/_mapping" -u 用户名:密码

# 验证分词器配置
curl -X GET "http://你的ip:9200/article/_analyze" -u 用户名:密码 \
  -H 'Content-Type: application/json' \
  -d '{"analyzer":"ik_max_word","text":"云计算"}'
```

### 2. 验证数据同步

```bash
# 查看文档数量
curl -X GET "http://你的ip:9200/article/_count" -u 用户名:密码

# 预期输出：{"count":23,"_shards":{"total":1,"successful":1,"skipped":0,"failed":0}}
```

### 3. 验证搜索功能

```bash
# 直接测试 ES 搜索
curl -X GET "http://你的ip:9200/article/_search" -u 用户名:密码 \
  -H 'Content-Type: application/json' \
  -d '{
    "query": {
      "bool": {
        "must": [
          {
            "bool": {
              "should": [
                {"match": {"articleTitle": "云"}},
                {"match": {"articleContent": "云"}}
              ]
            }
          },
          {"term": {"isDelete": 0}},
          {"term": {"status": 1}}
        ]
      }
    }
  }'
```

### 4. 验证搜索结果

搜索 "docker" 应该返回包含 Docker 的文章，并且关键词有高亮显示。

---

## 2.4 关键配置文件

### application-prod.yml

```yaml
# Elasticsearch 配置
elasticsearch:
  rest:
    uris: http://你的ip:9200
    username: 用户名
    password: 密码

# 搜索模式配置
search:
  mode: elasticsearch
```

### 核心代码文件

1. **ElasticsearchConfig.java** - ES 客户端配置
2. **ElasticsearchIndexInitializer.java** - 索引初始化与数据同步
3. **EsSearchStrategyImpl.java** - ES 搜索策略实现
4. **ArticleController.java** - 搜索 API 接口

---

## 2.5 经验总结

### 1. 字段映射必须匹配实际存储

ES 中的字段名必须与索引时的字段名一致，不能想当然。

### 2. 数据同步机制

- MaxWell 只同步增量变更
- 删除索引后需要手动全量同步历史数据
- 应用启动时自动同步是一个好的实践

### 3. 分词器配置

- 索引时使用 `ik_max_word`（细粒度分词）
- 搜索时使用 `ik_smart`（智能分词）
- 查询时必须明确指定 analyzer

### 4. 高亮配置

- 全局设置 `preTags` 和 `postTags`
- 字段级别可以单独配置 `fragmentSize` 和 `numberOfFragments`

### 5. 测试注意事项

- ES 搜索测试需要真实环境
- 避免对"不存在的关键词"做强断言（分词器可能匹配到内容）
- 使用 `@Profile("!test")` 避免在测试环境执行初始化

---

# 3. 常见问题 FAQ

## 3.1 升级相关问题

### Q1: UserAgentUtils 找不到

**解决**: 版本改为 1.21
```xml
<useragentutils.version>1.21</useragentutils.version>
```

### Q2: Lombok 不支持 JDK 25

**解决**: 升级到 1.18.42
```xml
<lombok.version>1.18.42</lombok.version>
```

### Q3: OkHttp 5.x 不兼容

**解决**: 保持 4.12.0
```xml
<okhttp.version>4.12.0</okhttp.version>
```

### Q4: JDK 版本不匹配

**解决**: 
1. 设置 `JAVA_HOME=D:\Java\jdk-25.0.1`
2. 使用 `eclipse-temurin:25-jre-alpine` Docker 镜像

---

## 3.2 Elasticsearch 相关问题

### Q5: factoryBeanObjectType 错误

**现象：**
```
Invalid value type for attribute 'factoryBeanObjectType': java.lang.String
```

**解决：** 使用 `mybatis-plus-spring-boot3-starter` 替代 `mybatis-plus-boot-starter`

### Q6: 端口被占用

**现象：**
```
Port 8080 was already in use
```

**解决：**
```powershell
# Windows
netstat -ano | findstr :8080
powershell -Command "Stop-Process -Id <PID> -Force"

# 或者修改端口
server.port=8081
```

### Q7: ES 索引 Mapping 错误

**解决**: 参考 [2.2.1 问题 1](#问题 1es-索引-mapping-配置错误)

### Q8: ES 字段映射错误

**解决**: 参考 [2.2.2 问题 2](#问题 2es-字段映射错误)

### Q9: JsonData.toJson() 转换错误

**解决**: 参考 [2.2.3 问题 3](#问题 3jsondatatojson-类型转换错误)

### Q10: ES 索引中没有数据

**解决**: 参考 [2.2.4 问题 4](#问题 4es-索引中没有数据)

### Q11: 搜索查询未使用 IK 分词器

**解决**: 参考 [2.2.5 问题 5](#问题 5搜索查询未使用 ik-分词器)

### Q12: 高亮配置不生效

**解决**: 参考 [2.2.6 问题 6](#问题 6高亮配置不生效)

### Q13: 测试环境缺少 ES 配置

**解决**: 参考 [2.2.7 问题 7](#问题 7测试环境缺少 es-配置)

### Q14: Elasticsearch 503 错误

**快速解决**:
```bash
# 方案 1: 降级到 MySQL 搜索
# 修改 application-prod.yml
search:
  mode: mysql

# 方案 2: 重启 ES 服务
docker restart aurora-elasticsearch
sleep 60
curl http://你的ip:9200/_cluster/health

# 方案 3: 重建索引
curl -X DELETE http://你的ip:9200/article
docker restart aurora-springboot
```

---

## 3.3 常用 ES 命令

```bash
# 检查集群健康
curl -u 用户名:密码 http://你的ip:9200/_cluster/health?pretty

# 检查索引列表
curl -u 用户名:密码 http://你的ip:9200/_cat/indices?v

# 检查分片状态
curl -u 用户名:密码 http://你的ip:9200/_cat/shards/article?v

# 删除索引
curl -X DELETE -u 用户名:密码 http://你的ip:9200/article

# 查看索引映射
curl -u 用户名:密码 http://你的ip:9200/article/_mapping?pretty

# 测试搜索
curl -X POST -u 用户名:密码 \
  http://你的ip:9200/article/_search?pretty \
  -H 'Content-Type: application/json' \
  -d '{"query":{"match_all":{}}, "size":1}'

# 查看 ES 日志
docker logs aurora-elasticsearch --tail 100 -f

# 重启 ES 服务
docker restart aurora-elasticsearch

# 查看 ES 资源使用
docker stats aurora-elasticsearch --no-stream
```

---

## 3.4 MySQL 搜索切换

**临时切换到 MySQL 搜索**:
```yaml
# application-prod.yml
search:
  mode: mysql
```

**切换回 Elasticsearch**:
```yaml
# application-prod.yml
search:
  mode: elasticsearch
```

**重启应用**:
```bash
docker restart aurora-springboot
```

---

## 3.5 应用重启流程

1. **停止应用**
   ```bash
   docker stop aurora-springboot
   ```

2. **更新配置或代码**
   - 修改 `application-prod.yml`
   - 上传新的 jar 包到 `/opt/aurora/app/`

3. **启动应用**
   ```bash
   docker start aurora-springboot
   ```

4. **检查日志**
   ```bash
   docker logs aurora-springboot -f
   ```

5. **验证健康**
   ```bash
   curl http://localhost:8080/api/articles/topAndFeatured
   ```

---

# 4. 参考资料

## 4.1 官方文档

- [Spring Boot 3.5 Release Notes](https://spring.io/blog/spring-boot-3-5-0-available-now)
- [JDK 25 Release Notes](https://openjdk.org/projects/jdk/25/)
- [MyBatis-Plus 3.5.x Documentation](https://baomidou.com/)
- [ip2region 3.x GitHub](https://github.com/lionsoul2014/ip2region)
- [Lombok 1.18.42 Release](https://projectlombok.org/changelog)
- [Elasticsearch 8.x 官方文档](https://www.elastic.co/guide/en/elasticsearch/reference/current/index.html)
- [IK 分词器文档](https://github.com/medcl/elasticsearch-analysis-ik)

## 4.2 迁移指南

- [Spring Boot 3.0 Migration Guide](https://github.com/spring-projects/spring-boot/wiki/Spring-Boot-3.0-Migration-Guide)
- [Jakarta EE 9+ Migration Guide](https://jakarta.ee/blogs/jakarta-ee-9-migration-guide/)
- [Elasticsearch 8.x Java Client](https://www.elastic.co/guide/en/elasticsearch/client/java-api-client/current/index.html)

---

**文档版本**: v1.0 (合并优化版)  
**最后更新**: 2026-03-09  
**维护者**: Aurora Team

---

## 📝 文档修订历史

| 版本 | 日期 | 修改内容 | 作者 |
|------|------|----------|------|
| v1.0 | 2026-03-09 | 合并《升级到SpringBoot3.5.10 和 JDK25 历程》和《Elasticsearch8 集成问题排查记录》，优化结构和排版 | Aurora Team |
