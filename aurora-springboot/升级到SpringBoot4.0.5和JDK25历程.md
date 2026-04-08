# Aurora 博客系统 - 升级到 Spring Boot 4.0.5 和 JDK 25 历程

<div align="center">

![Spring Boot](https://img.shields.io/badge/Spring%20Boot-4.0.5-6DB33F?style=for-the-badge&logo=spring)
![JDK](https://img.shields.io/badge/JDK-25.0.1+-F80000?style=for-the-badge&logo=openjdk&logoColor=white)
![Elasticsearch](https://img.shields.io/badge/Elasticsearch-8.17.2-00BFb3?style=for-the-badge&logo=elasticsearch)
![IK Analyzer](https://img.shields.io/badge/IK-8.17.2-orange?style=for-the-badge)
![Status](https://img.shields.io/badge/Status-Production%20Ready-success?style=for-the-badge)

</div>

<div align="center">

| 项目 | 信息 |
|:---|:---|
| 📅 **更新时间** | 2026-04-09 |
| 👤 **维护者** | 七七 |
| 🔖 **文档版本** | v2.1 |

</div>

---

## 目录

1. [升级概述](#1-升级概述)
2. [依赖版本对比](#2-依赖版本对比)
3. [代码适配修改](#3-代码适配修改)
4. [常见问题与解决方案](#4-常见问题与解决方案)
5. [编译与部署](#5-编译与部署)
6. [验证结果](#6-验证结果)
7. [升级总结](#7-升级总结)

---

# 1. 升级概述

| 属性 | 内容 |
|:---|:---|
| 📅 **升级时间** | 2026-04-08 |
| 🚀 **升级路线** | Spring Boot 3.5.10 + JDK 25 → Spring Boot 4.0.5 + JDK 25 |
| 💡 **升级原因** | 获取最新特性、内存优化（JVM 模式降低 11-17%，GraalVM Native Image 降低 71-81%） |

---

# 2. 依赖版本对比

## 2.1 升级变更

| 组件 | 原版本 | 新版本 | 说明 |
|:---|:---:|:---:|:---|
| **Spring Boot** | 3.5.10 | **4.0.5** | 基于 Spring Framework 7.0 |
| **Spring Framework** | 6.x | **7.0.6** | 核心框架重大更新 |
| **Spring Security** | 6.x | **7** | 安全框架 API 变更 |
| **Elasticsearch Client** | 8.13.4 | **8.17.2** | Java Client 升级 |
| **MyBatis-Plus** | 3.5.6/3.5.16 | **统一 3.5.16** | 修复版本不一致问题 |
| **MySQL Connector/J** | 9.6.0 | **9.6.0** | 保持不变 |
| **AOP** | spring-boot-starter-aop | **aspectjweaver 1.9.22.1** | SB 4.0 移除 AOP starter |

## 2.2 保持不变的依赖

| 依赖项 | 版本 | 说明 |
|:---|:---:|:---|
| **JDK** | 25 | 保持不变 |
| **Lombok** | 1.18.42 | 支持 JDK 25 |
| **JJWT** | 0.12.5 | JWT 令牌 |
| **Quartz** | 2.5.2 | 定时任务 |
| **OkHttp** | 4.12.0 | HTTP 客户端 |
| **FastJSON2** | 2.0.61 | JSON 序列化库 |
| **Hutool** | 5.8.43 | Java 工具类库 |

---

# 3. 代码适配修改

## 3.1 pom.xml - 依赖调整

**📁 文件位置**: `pom.xml`

```xml
<!-- ✅ Spring Boot 版本升级 -->
<spring-boot.version>4.0.5</spring-boot.version>

<!-- ✅ Elasticsearch 版本升级 -->
<elasticsearch.version>8.17.2</elasticsearch.version>

<!-- ✅ MyBatis-Plus 统一版本 -->
<mybatis-plus.version>3.5.16</mybatis-plus.version>

<!-- ❌ 移除：SB 4.0 已移除 AOP starter -->
<!-- <dependency>
    <groupId>org.springframework.boot</groupId>
    <artifactId>spring-boot-starter-aop</artifactId>
</dependency> -->

<!-- ✅ 新增：直接引入 aspectjweaver -->
<dependency>
    <groupId>org.aspectj</groupId>
    <artifactId>aspectjweaver</artifactId>
    <version>1.9.22.1</version>
</dependency>

<!-- ✅ 新增：Spring Security 7 需要 access 模块 -->
<dependency>
    <groupId>org.springframework.security</groupId>
    <artifactId>spring-security-access</artifactId>
</dependency>

<!-- ✅ 新增：MyBatis-Plus 分页插件独立模块 -->
<dependency>
    <groupId>com.baomidou</groupId>
    <artifactId>mybatis-plus-jsqlparser</artifactId>
    <version>${mybatis-plus.version}</version>
</dependency>
```

---

## 3.2 AuroraSpringbootApplication.java - 排除自动配置

**📁 文件位置**: `src/main/java/com/aurora/AuroraSpringbootApplication.java`

```java
@SpringBootApplication(exclude = {
    com.baomidou.mybatisplus.autoconfigure.MybatisPlusAutoConfiguration.class
})
@MapperScan("com.aurora.mapper")
public class AuroraSpringbootApplication {
    public static void main(String[] args) {
        SpringApplication.run(AuroraSpringbootApplication.class, args);
    }
}
```

> **📝 原因**: SB 4.0 的自动配置顺序变化，`MybatisPlusAutoConfiguration` 在 `DataSourceAutoConfiguration` 之前执行，导致找不到 DataSource。

---

## 3.3 MybatisPlusConfig.java - 手动配置 SqlSessionFactory

**📁 文件位置**: `src/main/java/com/aurora/config/MybatisPlusConfig.java`

```java
@EnableTransactionManagement
@Configuration
@MapperScan("com.aurora.mapper")
public class MybatisPlusConfig {

    @Autowired
    private DataSource dataSource;

    /**
     * 手动配置 SqlSessionFactory（SB 4.0 兼容）
     */
    @Bean
    public SqlSessionFactory sqlSessionFactory() throws Exception {
        MybatisSqlSessionFactoryBean sessionFactory = new MybatisSqlSessionFactoryBean();
        sessionFactory.setDataSource(dataSource);
        sessionFactory.setMapperLocations(new PathMatchingResourcePatternResolver()
                .getResources("classpath*:mapper/*.xml"));
        sessionFactory.setPlugins(mybatisPlusInterceptor());
        return sessionFactory.getObject();
    }

    /**
     * 配置 SqlSessionTemplate
     */
    @Bean
    public SqlSessionTemplate sqlSessionTemplate(SqlSessionFactory sqlSessionFactory) {
        return new SqlSessionTemplate(sqlSessionFactory);
    }

    /**
     * 配置事务管理器
     */
    @Bean
    public PlatformTransactionManager transactionManager() {
        return new DataSourceTransactionManager(dataSource);
    }

    @Bean
    public MybatisPlusInterceptor mybatisPlusInterceptor() {
        MybatisPlusInterceptor interceptor = new MybatisPlusInterceptor();
        PaginationInnerInterceptor paginationInterceptor = new PaginationInnerInterceptor(DbType.MYSQL);
        paginationInterceptor.setOverflow(false);
        paginationInterceptor.setMaxLimit(500L);
        interceptor.addInnerInterceptor(paginationInterceptor);
        return interceptor;
    }
}
```

---

## 3.4 WebSecurityConfig.java - Spring Security 7 API 适配

**📁 文件位置**: `src/main/java/com/aurora/config/WebSecurityConfig.java`

```java
// ❌ 旧版本（SB 3.5 / Security 6）
@Bean
public DaoAuthenticationProvider authenticationProvider() {
    DaoAuthenticationProvider provider = new DaoAuthenticationProvider();
    provider.setUserDetailsService(userDetailServiceImpl);
    provider.setPasswordEncoder(passwordEncoder());
    return provider;
}

// ✅ 新版本（SB 4.0 / Security 7）
@Bean
public DaoAuthenticationProvider authenticationProvider() {
    // Security 7 要求构造函数传入 UserDetailsService
    DaoAuthenticationProvider provider = new DaoAuthenticationProvider(userDetailServiceImpl);
    provider.setPasswordEncoder(passwordEncoder());
    return provider;
}
```

---

## 3.5 ElasticsearchConfig.java - ES Client 超时配置

**📁 文件位置**: `src/main/java/com/aurora/config/ElasticsearchConfig.java`

```java
@Bean
public RestClient elasticsearchRestClient() {
    String[] uriParts = uris.replace("http://", "").split(":");
    String hostname = uriParts[0];
    int port = Integer.parseInt(uriParts[1]);

    final CredentialsProvider credentialsProvider = new BasicCredentialsProvider();
    credentialsProvider.setCredentials(AuthScope.ANY,
            new UsernamePasswordCredentials(username, password));

    // ✅ SB 4.0 + ES 8.17.2 需要显式设置超时
    RequestConfig requestConfig = RequestConfig.custom()
            .setConnectTimeout(10000)   // 连接超时 10s
            .setSocketTimeout(60000)   // Socket 超时 60s（支持批量同步）
            .setConnectionRequestTimeout(10000)
            .build();

    RestClientBuilder builder = RestClient.builder(new HttpHost(hostname, port, "http"))
            .setHttpClientConfigCallback((HttpAsyncClientBuilder httpClientBuilder) -> {
                httpClientBuilder.setDefaultCredentialsProvider(credentialsProvider)
                        .setDefaultRequestConfig(requestConfig)
                        .setMaxConnTotal(100)      // 最大连接数
                        .setMaxConnPerRoute(30);   // 每路由最大连接数
                return httpClientBuilder;
            });

    return builder.build();
}
```

---

## 3.6 application-prod.yml - HikariCP 连接池优化

**📁 文件位置**: `src/main/resources/application-prod.yml`

```yaml
spring:
  datasource:
    hikari:
      minimum-idle: 5              # ✅ 从 10 降到 5
      maximum-pool-size: 20        # ✅ 从 100 降到 20（防止 MySQL Too many connections）
      idle-timeout: 180000
      max-lifetime: 1800000
      connection-timeout: 30000
      keepalive-time: 300000
```

---

# 4. 常见问题与解决方案

## ❌ 问题 1: spring-boot-starter-aop 不存在

| 类别 | 内容 |
|:---|:---|
| 🔴 **错误** | `Could not find artifact org.springframework.boot:spring-boot-starter-aop` |
| 🔴 **原因** | SB 4.0 移除了 AOP starter |
| 🟢 **解决** | 直接引入 `aspectjweaver` |

```xml
<dependency>
    <groupId>org.aspectj</groupId>
    <artifactId>aspectjweaver</artifactId>
    <version>1.9.22.1</version>
</dependency>
```

---

## ❌ 问题 2: MyBatis-Plus SqlSessionFactory 创建失败

| 类别 | 内容 |
|:---|:---|
| 🔴 **错误** | `Property 'sqlSessionFactory' or 'sqlSessionTemplate' are required` |
| 🔴 **原因** | SB 4.0 自动配置顺序变化，DataSource 未就绪 |
| 🟢 **解决** | 排除 `MybatisPlusAutoConfiguration`，手动配置 `SqlSessionFactory` |

> 📖 参考 [3.3 MybatisPlusConfig.java](#33-mybatisplusconfigjava---手动配置-sqlsessionfactory)

---

## ❌ 问题 3: Elasticsearch 同步超时

| 类别 | 内容 |
|:---|:---|
| 🔴 **错误** | `30,000 milliseconds timeout on connection http-outgoing-0` |
| 🔴 **原因** | ES 8.17.2 + SB 4.0 需要显式配置超时 |
| 🟢 **解决** | 在 `ElasticsearchConfig` 中设置 `RequestConfig` |

> 📖 参考 [3.5 ElasticsearchConfig.java](#35-elasticsearchconfigjava---es-client-超时配置)

---

## ❌ 问题 4: MySQL Too many connections

| 类别 | 内容 |
|:---|:---|
| 🔴 **错误** | `java.sql.SQLNonTransientConnectionException: Too many connections` |
| 🔴 **原因** | HikariCP 最大连接数 100 超过 MySQL 默认限制 |
| 🟢 **解决** | 降低 `maximum-pool-size` 到 20 |

> 📖 参考 [3.6 application-prod.yml](#36-application-prodyml---hikaricp-连接池优化)

---

## ❌ 问题 5: ES article 索引 RED 状态

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

# 5. 编译与部署

## 5.1 编译环境要求

| 环境 | 版本要求 |
|:---|:---|
| **JDK** | 25.0.1+ |
| **Maven** | 3.9.x+ |

```bash
# 设置 JDK 环境
set JAVA_HOME=D:\Java\jdk-25.0.1

# 编译项目
mvn clean package -DskipTests -U
```

---

## 5.2 JVM 优化参数

### 本地开发 / 测试环境

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

### 生产环境 Docker 部署

```bash
java -Xms48m -Xmx96m \
     -XX:+UseSerialGC \
     -XX:MetaspaceSize=128m -XX:MaxMetaspaceSize=192m \
     -XX:ReservedCodeCacheSize=24m -Xss256k \
     -XX:+UseCompressedOops \
     -XX:+UseCompressedClassPointers \
     -Djava.security.egd=file:/dev/./urandom \
     -Dspring.jmx.enabled=false \
     -Dspring.backgroundpreinitializer.ignore=true \
     -Djdk.attach.allowAttachSelf=false \
     -jar /app/blog.jar
```

### Elasticsearch 生产环境参数

```yaml
# ES_JAVA_OPTS（docker-compose.yml）
ES_JAVA_OPTS: "-Xms128m -Xmx160m"
```

---

# 6. 验证结果

## 6.1 应用启动验证

```
✅ Spring Boot 4.0.5 启动成功
✅ 使用 JDK 25 运行（eclipse-temurin:25-jre-alpine）
✅ Tomcat 在 8080 端口启动（context-path: /api）
✅ 数据库连接正常（HikariCP 最大 20 连接）
✅ Quartz 定时任务启动成功
✅ ip2region.xdb 加载成功 (11MB)
✅ RabbitMQ 连接成功
✅ Elasticsearch 8.17.2 + IK 8.17.2 连接成功
✅ 文章数据同步：28/28 成功（100%）
```

---

## 6.2 功能验证

| 功能模块 | 状态 |
|:---|:---:|
| ✅ QQ 登录 | 正常 |
| ✅ 邮箱登录 | 正常 |
| ✅ 用户权限 | 正常 |
| ✅ 文章管理 | 正常 |
| ✅ Elasticsearch 搜索 + IK 分词 | **正常**（28 篇文章全部同步） |
| ✅ 定时任务 | 正常 |

---

## 6.3 性能测试数据

### 启动时间

| 阶段 | 耗时 |
|:---|:---:|
| Spring Boot 启动到 Tomcat 就绪 | ~17 秒 |
| 完整启动（含 ES 同步 27 篇文章） | **~22.7 秒** |

**启动日志关键时间点**：

```
21:56:52.992 - 应用开始启动
21:57:14.005 - Tomcat 启动完成 (21.013 秒)
21:57:15.107 - 应用完全启动 (22.115 秒)
21:57:17.063 - ES 同步完成 (24.071 秒)
```

### 内存占用

| 指标 | 数值 |
|:---|:---:|
| **物理内存 (RSS)** | **733 MB** |
| JVM Heap - Used | 102.5 MB |
| JVM Heap - Committed | 600 MB (Young: 344MB + Old: 245.7MB) |
| Metaspace - Used | 86.2 MB |
| Metaspace - Committed | 86.9 MB |

**GC 统计**：

| 指标 | 数值 |
|:---|:---:|
| Young GC 次数 | 22 次 |
| Young GC 总耗时 | 0.143 秒 |
| Full GC 次数 | 0 次 |
| Concurrent GC 次数 | 6 次 |

### 与 SB 3.5.10 对比

| 指标 | SB 3.5.10 | SB 4.0.5 | 变化 |
|:---|:---:|:---:|:---|
| 启动时间 | ~10 秒 | ~22.7 秒 | +127% ⬆️ |
| 物理内存 | ~418 MB | **733 MB** | +75% ⬆️ |
| JVM Heap Used | ~180 MB | **102.5 MB** | -43% ⬇️ |
| ES 同步成功率 | 0/27 | **27/27** | 100% ⬆️ |

> **⚠️ 注意**：
> - SB 4.0.5 启动时间增加是因为 Spring Framework 7.0 增加了更多初始化检查和 Bean 验证
> - 物理内存增加主要是由于 JDK 25 和 Spring Framework 7.0 的基础开销增加
> - **JVM Heap 使用量反而降低了 43%**，说明 SB 4.0.5 的内存管理更高效
> - 预期经过 JIT 预热和 G1 GC 优化后，长期运行的内存占用会进一步降低
> - 如需显著降低内存，可考虑 GraalVM Native Image（预期降低 71-81%）或 GraalVM JIT（预计节省 5-15%）

---

## 6.4 生产环境部署优化

### 服务器信息

| 属性 | 内容 |
|:---|:---|
| 🖥️ **服务器** | 腾讯云轻量应用服务器（广州） |
| 🆔 **实例 ID** | 你的实例 |
| 🌐 **公网 IP** | 你的ip |
| 💾 **内存** | 3.3 GiB |
| 🌐 **域名** | www.你的域名 / admin.你的域名 / static.你的域名 |

### 问题排查与修复记录

| # | 时间 | 问题 | 根因 | 解决方案 | 状态 |
|:---:|:---:|------|------|---------|:----:|
| 1 | 23:12 | ES 启动崩溃 | `cache.max_size` 无效配置项（ES 8.x 不支持） | 从 docker-compose.yml 删除 | ✅ |
| 2 | 23:14 | IK 分词器 NPE 空指针 | IK **7.17.12** 与 ES **8.17.2** 版本不匹配 | 重装 IK **8.17.2** | ✅ |
| 3 | 23:22 | IK 字典文件缺失 | 手动下载 zip 缺少 `config/` 目录 + volume 挂载覆盖容器内正确插件 | 去掉 plugins volume 挂载，改用 `elasticsearch-plugin install` 安装 | ✅ |
| 4 | 23:23 | ES OOM 崩溃 | 64MB 堆内存不够用（ES 8.x + IK 插件需要更多内存） | 调大到 128MB → 最终 **160MB** | ✅ |
| 5 | 23:27 | 文章同步 circuit_breaking | HTTP 熔断器限制太小（108MB），大文章内容超限 | 调大堆内存 + breaker.limit 60% | ✅ |
| 6 | 24:00 | SpringBoot Metaspace OOM | 异步任务动态加载 MySQL 驱动类，Metaspace 96m 不够 | 增加 MetaspaceSize 到 128m，MaxMetaspaceSize 到 192m | ✅ |

### IK 分词器安装命令

```bash
# 进入 ES 容器安装（每次容器重建后需重新执行）
docker exec -it aurora-elasticsearch /bin/bash
./bin/elasticsearch-plugin install https://release.infinilabs.com/analysis-ik/stable/elasticsearch-analysis-ik-8.17.2.zip

# 输入 y 确认后重启 ES 激活插件
docker restart aurora-elasticsearch
```

> **⚠️ 重要**: 不要使用宿主机 plugins volume 挂载！直接用 `elasticsearch-plugin install` 安装到容器内部最可靠。

### 最终生产环境参数

#### SpringBoot JVM 参数

```bash
java -Xms48m -Xmx96m \
     -XX:+UseSerialGC \
     -XX:MetaspaceSize=128m -XX:MaxMetaspaceSize=192m \
     -XX:ReservedCodeCacheSize=24m -Xss256k \
     -XX:+UseCompressedOops \
     -XX:+UseCompressedClassPointers \
     -Djava.security.egd=file:/dev/./urandom \
     -Dspring.jmx.enabled=false \
     -Dspring.backgroundpreinitializer.ignore=true \
     -Djdk.attach.allowAttachSelf=false \
     -jar /app/blog.jar
```

#### Elasticsearch 参数

```yaml
# docker-compose.yml
environment:
  ES_JAVA_OPTS: "-Xms128m -Xmx160m"
  discovery.type: single-node
  bootstrap.memory_lock: "true"
  xpack.security.enabled: "false"
  indices.query.bool.max_clause_count: "1024"
  indices.memory.index_buffer_size: "20%"
  indices.fielddata.cache.size: "10%"
  indices.breaker.total.limit: "60%"
```

### 生产环境实测数据

#### 各服务内存占用

| 服务 | 内存占用 | 占比 | CPU |
|:---|:---|:---|:---|
| **Elasticsearch** (128/160MB + IK) | **677 MiB** | 19.9% | 0.75% |
| **SpringBoot** (48/96MB) | **285 MiB** | 8.38% | 0.32% |
| MySQL | 163 MiB | 4.9% | 0.60% |
| Maxwell | 154 MiB | 4.6% | 0.19% |
| MinIO | 135 MiB | 4.1% | 0.00% |
| RabbitMQ | 124 MiB | 3.7% | 0.14% |
| Redis | 7.4 MiB | 0.2% | 0.42% |
| Nginx | 19 MiB | 0.6% | 0.00% |
| **合计** | **~1.56 GiB** | **47%** | — |

#### 主机内存状态

```
总计: 3.3 GiB
已用: 2.5 GiB (75%)
可用: 827 MiB
Swap: 无
```

#### SpringBoot 启动性能

| 指标 | 数值 |
|:---|:---|
| 启动耗时 | **11.2 秒** |
| 文章同步成功率 | **28/28 (100%)** |
| 同步失败数 | **0** |
| API 响应时间 | **176ms** |

### 优化前后对比

| 指标 | 优化前（首次部署） | 优化后（最终） | 节省 |
|:---|:---|:---|:---|
| SpringBoot RSS | ~342 MiB | **285 MiB** | -57 MiB (-17%) |
| Elasticsearch RSS | ~697 MiB | **677 MiB** | -20 MiB (-3%) |
| ES Heap | 64 MB | **160 MB** | 调大以适配 IK 插件 |
| SB Heap | ~256 MB | **96 MB** | -62% |
| SB 启动时间 | 较慢 | **11.8 秒** | 显著提升 |
| ES 同步成功 | 0/28 (Connection refused) | **28/28 (100%)** | 完全修复 |

### 进一步优化的可行性分析

| 方案 | 可行性 | 预期收益 | 推荐度 |
|:---|:---:|:---:|:---:|
| 当前 JVM 优化（已完成） | ★★★★★ | 已节省 ~77 MiB | ✅ 已完成 |
| GraalVM JIT | ★★★★★ | 省 ~20-30 MiB（5-10%）| ★★★★☆ 可顺手做 |
| GraalVM Native Image | ★★☆☆☆ | 省 ~200+ MiB（60-70%）| ★★☆☆☆ SB 4.0 + JDK 25 生态不成熟，建议等半年 |
| 关停 Maxwell | ★★★★☆ | 省 154 MiB | ★★★☆☆ 不用 CDC 时可关 |

> **💡 结论**: 核心服务（SpringBoot + ES）已到极限，当前 285 + 677 = 962 MiB 在 3.3G 主机上运行稳定（75% 使用率）。建议保持现状。

---

# 7. 升级总结

## 🎯 成功经验

| # | 经验说明 |
|:---:|:---|
| 1️⃣ | **提前备份**：升级前备份所有配置文件和数据库 |
| 2️⃣ | **依赖检查**：SB 4.0 移除了 AOP starter，需手动引入 aspectjweaver |
| 3️⃣ | **自动配置顺序**：SB 4.0 改变了自动配置顺序，需手动配置 MyBatis-Plus |
| 4️⃣ | **ES 超时配置**：ES 8.17.2 + SB 4.0 需要显式设置超时参数 |
| 5️⃣ | **连接池优化**：降低 HikariCP 最大连接数防止 MySQL 连接耗尽 |
| 6️⃣ | **IK 分词器版本匹配**：必须使用与 ES 完全一致的 IK 版本（ES 8.17.2 → IK 8.17.2）|
| 7️⃣ | **IK 安装方式**：用 `elasticsearch-plugin install` 安装到容器内部，不要用宿主机 volume 挂载 plugins 目录 |
| 8️⃣ | **ES 堆内存规划**：ES 8.x + IK 插件至少需要 128MB 堆内存，64MB 会 OOM |
| 9️⃣ | **Metaspace 规划**：异步任务中动态加载驱动类可能导致 Metaspace 增长，建议初始 128m，最大 192m |

---

## ⚠️ 注意事项

| # | 注意事项 |
|:---:|:---|
| 1️⃣ | **Spring Security 7**：`DaoAuthenticationProvider` 构造函数需要传入 `UserDetailsService` |
| 2️⃣ | **MyBatis-Plus 分页**：需要单独引入 `mybatis-plus-jsqlparser` 依赖 |
| 3️⃣ | **ES IK 分词器**：必须在 ES 服务器上安装与 ES 版本一致的 IK 插件 |
| 4️⃣ | **JAR 文件锁定**：Windows 环境下重新构建前需停止所有 Java 进程 |
| 5️⃣ | **GraalVM Native Image**：如需进一步降低内存，可考虑迁移到 Native Image（预期降低 71-81%），但 SB 4.0 + JDK 25 生态尚未成熟，建议等半年后再尝试 |
| 6️⃣ | **GraalVM JIT**：零代码改动，改 Dockerfile 基础镜像即可，预计节省 5-15%（约 20-30 MiB）|
| 7️⃣ | **ES 无效配置项**：ES 8.x 不支持 `cache.max_size` 等旧版设置，需清理 |
| 8️⃣ | **容器重建后插件丢失**：每次 `docker compose up -d elasticsearch` 会重建容器，需重新安装 IK 插件 |
| 9️⃣ | **Metaspace OOM**：异步任务动态加载类（如 MySQL 驱动）可能导致 Metaspace 爆满，需监控并预留充足空间 |

---

## 📚 相关文档

- [升级与 Elasticsearch 集成完整指南](./升级与%20Elasticsearch%20集成完整指南.md) - Elasticsearch 8.17.2 集成问题排查与运维指南

---

<div align="center">
**文档版本**: v2.1 | **创建日期**: 2026-04-08 | **更新日期**: 2026-04-09 

![Built with ❤️](https://img.shields.io/badge/Built%20with-%E2%9D%A4%EF%B8%8F-blue?style=flat-square)

</div>
