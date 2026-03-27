# Aurora 博客系统 GraalVM JIT 迁移记录

> 将 Spring Boot 后端从 HotSpot JVM 迁移到 GraalVM JIT 的完整历程

---

## 📋 目录

- [概述](#概述)
- [迁移背景](#迁移背景)
- [迁移历程](#迁移历程)
- [迁移文件清单](#迁移文件清单)
- [部署指南](#部署指南)
- [GraalVM 特性](#graalvm-特性)
- [验证命令](#验证命令)
- [迁移总结](#迁移总结)

---

## 🌟 概述

| 项目 | 内容 |
|:---|:---|
| 📅 迁移日期 | 2026-03-27 |
| 🎯 迁移目标 | Spring Boot: HotSpot JVM → GraalVM JIT |
| 🖥️ 服务器 | 广州 Lighthouse (4C4G) |
| ☕ JDK 版本 | JDK 25 |
| 📉 内存改善 | **-19.5%** (342.8 MiB → 275.9 MiB) |

---

## 📖 迁移背景

在完成**四轮内存优化**后，Spring Boot 服务内存占用已从 **513.5 MiB** 降至 **342.8 MiB**。为进一步探索优化空间，开始研究 GraalVM JIT 模式。

### 为什么选择 GraalVM JIT？

| 优势 | 说明 |
|:---|:---|
| ✅ 零代码修改 | 完全兼容现有代码，迁移成本为 0 |
| ✅ Graal JIT 编译器 | 比 C2 编译器更高效的优化能力 |
| ✅ 性能提升 | 预期 10-20% 性能改善 |
| ✅ 内存优化 | 更高效的内存管理 |
| ✅ 向后兼容 | 与 HotSpot JVM 完全兼容 |

---

## 🚀 迁移历程

### 阶段一：内存收益评估

对比当前 HotSpot JVM 的基准内存使用情况：

```
┌─────────────────────────────────────┐
│  HotSpot JVM 内存占用: 342.8 MiB    │
│  堆内存配置: -Xms64m -Xmx128m       │
│  GC: SerialGC                       │
└─────────────────────────────────────┘
```

### 阶段二：寻找 Alpine 镜像

> **期望**: 使用更小的 Alpine 镜像进一步优化内存

**尝试方案**:
- 搜索官方 GraalVM Alpine 镜像
- 探索第三方 Alpine 构建

**结论**:
> ⚠️ **GraalVM 官方没有 Alpine 版本**，只有基于 Oracle Linux 的镜像
>
> **官方镜像列表**:
> - `ghcr.io/graalvm/jdk-community:25` ← **选用**
> - `ghcr.io/graalvm/jdk-community:25-ol9`
> - `oracle/graalvmce:25`

### 阶段三：参数对齐 HotSpot

为确保公平对比，将 GraalVM 参数设置为与 HotSpot **完全一致**:

```bash
java -Xms64m -Xmx128m \
     -XX:+UseSerialGC \
     -XX:MetaspaceSize=64m -XX:MaxMetaspaceSize=128m \
     -XX:ReservedCodeCacheSize=32m -Xss256k \
     -XX:+UseCompressedOops -XX:+UseCompressedClassPointers \
     -XX:+UseStringDeduplication -XX:+OptimizeStringConcat \
     -Djava.security.egd=file:/dev/./urandom \
     -Dspring.jmx.enabled=false \
     -jar /app/blog.jar
```

### 阶段四：实测验证

**实测结果** (2026-03-27):

| 指标 | HotSpot JVM | GraalVM JIT | 改善 |
|:---|:---:|:---:|:---:|
| **内存占用** | 342.8 MiB | 275.9 MiB | ⬇️ **-19.5%** |
| 堆内存配置 | -Xms64m -Xmx128m | -Xms64m -Xmx128m | = |
| GC | SerialGC | SerialGC | = |
| 状态 | 正常运行 | 正常运行 | ✅ |

> 💡 **结论**: 在相同参数下，GraalVM JIT 比 HotSpot JVM 节省约 **66.9 MiB** 内存

### 阶段五：集成编排

| 步骤 | 操作 |
|:---|:---|
| 1️⃣ | 将 GraalVM 集成到主 `docker-compose.yml` 统一管理 |
| 2️⃣ | 创建 `docker-compose-graalvm.yml` |
| 3️⃣ | 修复网络配置: `aurora_aurora-network` → `aurora-network` |

### 阶段六：本地配置同步

| 文件 | 说明 |
|:---|:---|
| `Dockerfile.graalvm` | GraalVM Dockerfile |
| `docker-compose-graalvm.yml` | GraalVM Compose 配置 |
| `deploy-graalvm.sh` | 部署说明文档 |
| `内存优化总结.md` | 优化总结文档 |

---

## 📦 迁移文件清单

### Dockerfile.graalvm

```dockerfile
FROM ghcr.io/graalvm/jdk-community:25

WORKDIR /app
COPY aurora-springboot-0.0.1.jar /app/blog.jar
EXPOSE 8080

ENTRYPOINT ["java", \
  "-Xms64m", "-Xmx128m", \
  "-XX:+UseSerialGC", \
  "-XX:MetaspaceSize=64m", "-XX:MaxMetaspaceSize=128m", \
  "-XX:ReservedCodeCacheSize=32m", "-Xss256k", \
  "-XX:+UseCompressedOops", "-XX:+UseCompressedClassPointers", \
  "-XX:+UseStringDeduplication", "-XX:+OptimizeStringConcat", \
  "-Djava.security.egd=file:/dev/./urandom", \
  "-Dspring.jmx.enabled=false", \
  "-Dspring.profiles.active=prod", \
  "-jar", "blog.jar"]
```

### docker-compose-graalvm.yml

```yaml
version: '3.8'

services:
  springboot:
    image: ghcr.io/graalvm/jdk-community:25
    container_name: aurora-springboot
    ports:
      - "8080:8080"
    volumes:
      - /opt/aurora/app/aurora-springboot-0.0.1.jar:/app/blog.jar
    environment:
      - SPRING_PROFILES_ACTIVE=prod
      - TZ=Asia/Shanghai
      - SPRING_DATASOURCE_URL=jdbc:mysql://aurora-mysql:3306/aurora?serverTimezone=Asia/Shanghai&allowMultiQueries=true&useSSL=false&allowPublicKeyRetrieval=true
      - SPRING_DATA_REDIS_HOST=aurora-redis
      - SPRING_RABBITMQ_HOST=aurora-rabbitmq
      - ELASTICSEARCH_REST_URIS=http://aurora-elasticsearch:9200
      - UPLOAD_MINIO_ENDPOINT=http://aurora-minio:9000/
    command: >
      java -Xms64m -Xmx128m -XX:+UseSerialGC
          -XX:MetaspaceSize=64m -XX:MaxMetaspaceSize=128m
          -XX:ReservedCodeCacheSize=32m -Xss256k
          -XX:+UseCompressedOops -XX:+UseCompressedClassPointers
          -XX:+UseStringDeduplication -XX:+OptimizeStringConcat
          -Djava.security.egd=file:/dev/./urandom
          -Dspring.jmx.enabled=false
          -jar /app/blog.jar
    networks:
      - aurora-network
    restart: always
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/"]
      interval: 30s
      timeout: 10s
      retries: 3

networks:
  aurora-network:
    external: true
```

---

## 🚢 部署指南

### 云端部署

```bash
# 1️⃣ 停止现有服务
docker stop aurora-springboot && docker rm aurora-springboot

# 2️⃣ 使用 docker-compose 启动
cd /opt/aurora
docker-compose -f docker-compose-graalvm.yml up -d springboot

# 3️⃣ 查看日志
docker-compose logs -f springboot

# 4️⃣ 验证
docker stats --no-stream | grep springboot
docker exec aurora-springboot java -version
```

### 回滚方案 (HotSpot JVM)

```bash
docker stop aurora-springboot && docker rm aurora-springboot

docker run -d --name aurora-springboot --network aurora-network -p 8080:8080 \
  -v /opt/aurora/app/aurora-springboot-0.0.1.jar:/app/blog.jar \
  -e SPRING_PROFILES_ACTIVE=prod -e TZ=Asia/Shanghai \
  -e SPRING_DATASOURCE_URL="jdbc:mysql://aurora-mysql:3306/aurora?serverTimezone=Asia/Shanghai&allowMultiQueries=true&useSSL=false&allowPublicKeyRetrieval=true" \
  -e SPRING_DATA_REDIS_HOST=aurora-redis -e SPRING_RABBITMQ_HOST=aurora-rabbitmq \
  -e ELASTICSEARCH_REST_URIS=http://aurora-elasticsearch:9200 \
  -e UPLOAD_MINIO_ENDPOINT=http://aurora-minio:9000/ \
  --restart=always tencentcloud/tencentkona25:jre-alpine \
  java -Xms64m -Xmx128m -XX:+UseSerialGC -XX:MetaspaceSize=64m -XX:MaxMetaspaceSize=128m \
       -XX:ReservedCodeCacheSize=32m -Xss256k -XX:+UseCompressedOops -XX:+UseCompressedClassPointers \
       -XX:+UseStringDeduplication -XX:+OptimizeStringConcat \
       -Djava.security.egd=file:/dev/./urandom -Dspring.jmx.enabled=false -jar /app/blog.jar
```

---

## ⚡ GraalVM 特性

### Graal JIT 编译器优势

| 特性 | 说明 |
|:---|:---|
| 🔥 更激进的内联 | 方法调用优化，减少 5-10% 开销 |
| 🚀 更好的逃逸分析 | 栈上分配优化，减少堆压力 |
| 🧹 优化的死代码消除 | 减少无效代码，提升 CPU 效率 |
| 🌍 多语言支持 | 支持 Truffle 框架，可运行多语言 |

### 注意事项

| ⚠️ 事项 | 说明 |
|:---|:---|
| ⏱️ 预热期 | GraalVM 需要运行 10-30 分钟才能体现完整优化效果 |
| 📦 镜像大小 | Oracle Linux 镜像比 Alpine 略大，但兼容性更好 |
| 🔄 回滚方案 | 保留原配置，可快速回滚到 HotSpot JVM |
| 🏭 生产部署 | 建议先在测试环境验证，再灰度发布到生产 |

---

## ✅ 验证命令

```bash
# 📊 查看容器内存使用
docker stats --no-stream | grep springboot

# ☕ 检查 JVM 版本
docker exec aurora-springboot java -version

# ❤️ 检查应用健康
curl http://localhost:8080/

# 🔧 查看 JVM 参数
docker exec aurora-springboot jcmd 1 VM.flags
```

---

## 📊 迁移总结

| 指标 | 数值 |
|:---|:---:|
| 📅 迁移日期 | 2026-03-27 |
| 🔴 HotSpot JVM 内存 | 342.8 MiB |
| 🟢 GraalVM JIT 内存 | 275.9 MiB |
| 💾 内存节省 | **66.9 MiB (-19.5%)** |
| 📝 代码修改 | **0 行** |
| 🚢 部署方式 | docker-compose 统一编排 |

### 关键决策

| # | 决策 | 说明 |
|:---:|:---|:---|
| 1 | 选择 Oracle Linux 镜像 | GraalVM 官方无 Alpine 版本，Oracle Linux 更稳定 |
| 2 | 参数完全对齐 | 确保公平对比，排除参数差异干扰 |
| 3 | 集成编排 | 统一管理，便于维护和回滚 |

### 后续优化方向

- 🎯 探索 GraalVM Native Image (静态编译) 进一步降低内存
- 🌍 评估 Truffle 框架下的多语言支持
- 📈 监控长期运行稳定性

---

<div align="center">

**文档更新时间**: 2026-03-27

</div>
