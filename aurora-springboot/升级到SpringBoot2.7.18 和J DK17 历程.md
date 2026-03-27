# Aurora 博客系统 - JDK 17 + Spring Boot 2.7.18 升级历程

> 从 JDK 1.8 + Spring Boot 2.3.7 升级到 JDK 17 + Spring Boot 2.7.18，同时集成 Elasticsearch 7.17.12

---

## 📅 升级时间

2026年3月4日

---

## 📋 升级概述

本次升级将 Aurora 博客系统从 JDK 1.8 + Spring Boot 2.3.7 升级到 JDK 17 + Spring Boot 2.7.18，同时集成 Elasticsearch 7.17.12 以支持全文搜索功能。

| 组件 | 升级前 | 升级后 |
|:---|:---:|:---:|
| JDK | 1.8 | **17** |
| Spring Boot | 2.3.7.RELEASE | **2.7.18** |
| Elasticsearch | - | **7.17.12** |
| MyBatis-Plus | 3.4.2 | **3.4.3.4** |
| Knife4j | - | **2.0.9** |
| Lombok | - | **1.18.24** |
| MySQL Connector | - | **8.0.33** |

---

## 🚀 升级步骤

### 1. 核心依赖版本升级 (pom.xml)

#### 1.1 基础版本配置

```xml
<properties>
    <java.version>17</java.version>
    <spring-boot.version>2.7.18</spring-boot.version>
    <elasticsearch.version>7.17.12</elasticsearch.version>
    <mybatis-plus.version>3.4.3.4</mybatis-plus.version>
    <knife4j.version>2.0.9</knife4j.version>
</properties>
```

#### 1.2 Maven 编译器配置

```xml
<plugin>
    <groupId>org.apache.maven.plugins</groupId>
    <artifactId>maven-compiler-plugin</artifactId>
    <version>3.8.1</version>
    <configuration>
        <source>17</source>
        <target>17</target>
        <annotationProcessorPaths>
            <path>
                <groupId>org.projectlombok</groupId>
                <artifactId>lombok</artifactId>
                <version>1.18.24</version>
            </path>
        </annotationProcessorPaths>
    </configuration>
</plugin>
```

#### 1.3 MySQL 依赖

```xml
<dependency>
    <groupId>mysql</groupId>
    <artifactId>mysql-connector-java</artifactId>
    <version>8.0.33</version>
    <scope>runtime</scope>
</dependency>
```

---

## 🔧 兼容性问题解决

### ❌ 问题 1: Knife4j 版本冲突

| 项目 | 内容 |
|:---|:---|
| 🔴 问题 | Knife4j 4.x 与 Spring Boot 2.7.18 不兼容，导致 NullPointerException |
| 🟢 解决 | 降级到 2.0.9 版本 |

```xml
<knife4j.version>2.0.9</knife4j.version>
```

---

### ❌ 问题 2: CORS 跨域配置

| 项目 | 内容 |
|:---|:---|
| 🔴 问题 | `allowedOrigins("*")` 与 `allowCredentials(true)` 冲突 |
| 🟢 解决 | 使用 `allowedOriginPatterns("*")` |

```java
registry.addMapping("/**")
    .allowedOriginPatterns("*")
    .allowCredentials(true)
    .allowedMethods("GET", "POST", "PUT", "DELETE", "OPTIONS")
    .maxAge(3600);
```

---

### ❌ 问题 3: CompletableFuture 类型转换

| 项目 | 内容 |
|:---|:---|
| 🔴 问题 | MyBatis-Plus 3.4.3+ 的 `selectCount()` 返回 `Long` 而非 `Integer` |
| 🔴 涉及文件 | AuroraInfoServiceImpl、ArticleServiceImpl、CategoryServiceImpl、CommentServiceImpl、MenuServiceImpl、PhotoAlbumServiceImpl、ResourceServiceImpl、RoleServiceImpl、TalkServiceImpl、TagServiceImpl |
| 🟢 解决 | 修改类型声明 |

```java
// 修改前
CompletableFuture<Integer> asyncArticleCount = CompletableFuture.supplyAsync(...);

// 修改后
CompletableFuture<Long> asyncArticleCount = CompletableFuture.supplyAsync(...);

// 使用时转换
.articleCount(asyncArticleCount.get().intValue())
```

---

### ❌ 问题 4: Lombok 注解不生效

| 项目 | 内容 |
|:---|:---|
| 🔴 问题 | 使用 JDK 25 时 Lombok 注解无法生成代码 |
| 🟢 解决 | 必须使用 JDK 17 编译和运行项目 |

---

## ⚙️ 配置文件调整

### application-prod.yml

```yaml
# JWT 配置
jwt:
  secret: aurora

# Spring MVC 路径匹配策略（Spring Boot 2.6+ 必需）
spring:
  mvc:
    pathmatch:
      matching-strategy: ant_path_matcher
```

---

## 🧪 验证与运行

### 设置 JDK 17 环境

```powershell
$env:JAVA_HOME='D:\Java\jdk-17.0.12'
```

### 编译项目

```powershell
& 'D:\Java\apache-maven-3.9.12\bin\mvn.cmd' clean compile -DskipTests
```

### 启动项目

```powershell
& 'D:\Java\apache-maven-3.9.12\bin\mvn.cmd' spring-boot:run
```

### 打包部署

```powershell
& 'D:\Java\apache-maven-3.9.12\bin\mvn.cmd' clean package -DskipTests
```

---

## ✅ 升级成果

| 成果 | 状态 |
|:---|:---:|
| ✅ 项目成功使用 JDK 17 编译 | |
| ✅ Spring Boot 2.7.18 正常运行 | |
| ✅ Elasticsearch 7.17.12 集成完成 | |
| ✅ API 文档（Knife4j）正常访问 | |
| ✅ 所有业务功能正常 | |
| ✅ Lombok 注解完全生效 | |
| ✅ CORS 跨域配置正确 | |
| ✅ 项目启动时间约 10 秒 | |

---

## ⚠️ 注意事项

| # | 注意事项 |
|:---:|:---|
| 1️⃣ | JDK 版本：必须使用 JDK 17，不要使用 JDK 21 或更高版本 |
| 2️⃣ | Knife4j 版本：必须使用 2.0.9，不要使用 4.x |
| 3️⃣ | CORS 配置：必须使用 `allowedOriginPatterns("*")` 而非 `allowedOrigins("*")` |
| 4️⃣ | 类型转换：MyBatis-Plus `selectCount()` 返回 `Long`，需要通过 `.intValue()` 转换 |

---

## 📝 修改文件清单

| # | 文件 | 说明 |
|:---:|:---|:---|
| 1️⃣ | `pom.xml` | 核心依赖和插件配置 |
| 2️⃣ | `src/main/resources/application-prod.yml` | 生产环境配置 |
| 3️⃣ | `src/main/java/com/aurora/config/WebMvcConfig.java` | CORS 配置 |
| 4️⃣ | 10个 ServiceImpl 文件 | CompletableFuture 类型调整 |

---

## 🔗 参考资料

| 文档 | 链接 |
|:---|:---|
| 📘 Spring Boot 2.7 官方文档 | [查看](https://docs.spring.io/spring-boot/docs/2.7.18/reference/html/) |
| 📝 Spring Boot 2.7 发行说明 | [查看](https://github.com/spring-projects/spring-boot/wiki/Spring-Boot-2.7-Release-Notes) |
| 📦 Lombok 与 Java 版本兼容性 | [查看](https://projectlombok.org/changelog) |
| 🔪 Knife4j 版本兼容性 | [查看](https://doc.xiaominfo.com/docs/knife4j/version) |

---

<div align="center">

**文档版本**: 1.0  
**最后更新**: 2026-03-04

</div>
