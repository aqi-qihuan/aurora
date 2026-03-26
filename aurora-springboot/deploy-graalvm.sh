# Aurora GraalVM JIT 部署指南

## 服务器信息
- IP: 你的ip
- 地域: 广州 (ap-guangzhou)

---

## 方式一：使用 docker-compose（推荐）

```bash
# 1. 上传配置到服务器
scp aurora-springboot/docker-compose-graalvm.yml root@你的ip:/opt/aurora/

# 2. 停止旧容器并启动
cd /opt/aurora
docker-compose up -d springboot

# 3. 查看日志
docker-compose logs -f springboot
```

---

## 方式二：手动 docker run

```bash
docker stop aurora-springboot && docker rm aurora-springboot

docker run -d --name aurora-springboot --network aurora-network -p 8080:8080 \
  -v /opt/aurora/app/aurora-springboot-0.0.1.jar:/app/blog.jar \
  -e SPRING_PROFILES_ACTIVE=prod \
  -e TZ=Asia/Shanghai \
  -e SPRING_DATASOURCE_URL="jdbc:mysql://aurora-mysql:3306/aurora?serverTimezone=Asia/Shanghai&allowMultiQueries=true&useSSL=false&allowPublicKeyRetrieval=true" \
  -e SPRING_DATA_REDIS_HOST=aurora-redis \
  -e SPRING_RABBITMQ_HOST=aurora-rabbitmq \
  -e ELASTICSEARCH_REST_URIS=http://aurora-elasticsearch:9200 \
  -e UPLOAD_MINIO_ENDPOINT=http://aurora-minio:9000/ \
  --restart=always ghcr.io/graalvm/jdk-community:25 \
  java -Xms64m -Xmx128m -XX:+UseSerialGC -XX:MetaspaceSize=64m -XX:MaxMetaspaceSize=128m \
       -XX:ReservedCodeCacheSize=32m -Xss256k -XX:+UseCompressedOops -XX:+UseCompressedClassPointers \
       -XX:+UseStringDeduplication -XX:+OptimizeStringConcat \
       -Djava.security.egd=file:/dev/./urandom -Dspring.jmx.enabled=false -jar /app/blog.jar
```

---

## 验证

```bash
docker-compose ps springboot
docker exec aurora-springboot java -version
curl http://localhost:8080/
docker stats --no-stream | grep springboot
```

---

## 回滚（HotSpot JVM）

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

## 实际收益（2026-03-27 实测）

| 指标 | HotSpot JVM | GraalVM JIT | 改善 |
|------|-------------|-------------|------|
| 内存占用 | 342.8 MiB | 275.9 MiB | ⬇️ **-19.5%** |
| 堆内存配置 | -Xms64m -Xmx128m | -Xms64m -Xmx128m | 一致 |
| 状态 | 正常运行 | 正常运行 | - |

**JVM 参数**：
- `-Xms64m -Xmx128m` 堆内存
- `-XX:+UseSerialGC` 串行 GC
- `-XX:MetaspaceSize=64m -XX:MaxMetaspaceSize=128m` 元空间
- `-XX:ReservedCodeCacheSize=32m` 代码缓存
- `-Xss256k` 线程栈
