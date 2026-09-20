# 中间件升级服务器操作手册

> 适用提交：c3dc840(P0) / 23cfdcf(P1) / 7470730(P2)
> 目标：在 134.175.206.158 完成中间件升级 + Go 1.27 部署

## 〇、前置准备（必做）

```bash
# 1. 全量快照（腾讯云轻量服务器先打快照）
# 2. MySQL 全量备份
docker exec aurora-mysql mysqldump -uroot -p${AURORA_MYSQL_PASSWORD} \
  --single-transaction --routines --triggers aurora > /opt/aurora/backup/aurora_$(date +%Y%m%d).sql
# 3. MinIO 数据导出（迁移到 RustFS 需要）
docker run --rm --network aurora-network \
  -v /opt/aurora/backup:/backup \
  bitnami/minio:2023.12.7 \
  mc alias set src http://aurora-minio:9000 ${MINIO_ROOT_USER} ${MINIO_ROOT_PASSWORD} && \
  mc mirror src/aurora /backup/minio_aurora/ --overwrite
# 4. ES 快照（可选，博客索引可全量重建）
curl -X PUT "localhost:9200/_snapshot/backup" -H 'Content-Type: application/json' \
  -d '{"type":"fs","settings":{"location":"/usr/share/elasticsearch/backup"}}'
```

## 一、P0：中间件断供/EOL 抢救

### 1.1 MinIO → RustFS（数据迁移）

```bash
# 1. 停 aurora-go（避免上传写入）
docker stop aurora-go

# 2. 起新 RustFS（新数据目录）
mkdir -p /opt/aurora/rustfs/data && chown -R 10001:10001 /opt/aurora/rustfs
cd /opt/aurora/app && git pull  # 拉取 compose 改动
docker compose -f docker-compose-go.yml up -d rustfs

# 3. 建 bucket（RustFS 无 bitnami 的 MINIO_DEFAULT_BUCKETS）
docker run --rm --network aurora-network bitnami/minio:2023.12.7 \
  mc alias set dst http://aurora-rustfs:9000 ${MINIO_ROOT_USER} ${MINIO_ROOT_PASSWORD} && \
  docker run --rm --network aurora-network bitnami/minio:2023.12.7 \
  mc mb dst/aurora

# 4. 灌数据（S3 API 层拷贝，磁盘格式不兼容不能换挂卷）
docker run --rm --network aurora-network \
  -v /opt/aurora/backup/minio_aurora:/data \
  bitnami/minio:2023.12.7 \
  bash -c "mc alias set src http://aurora-minio:9000 ${MINIO_ROOT_USER} ${MINIO_ROOT_PASSWORD} 2>/dev/null; \
           mc alias set dst http://aurora-rustfs:9000 ${MINIO_ROOT_USER} ${MINIO_ROOT_PASSWORD}; \
           mc mirror src/aurora dst/aurora --overwrite"

# 5. 验证对象数一致后停老 MinIO
docker stop aurora-minio && docker rm aurora-minio

# 6. 起 aurora-go（endpoint 不变 http://IP:9000，RustFS 端口同 9000）
docker compose -f docker-compose-go.yml up -d aurora-go
```

### 1.2 MySQL 8.0.32 → 8.4.10 LTS

```bash
# 1. 停服务
docker stop aurora-go aurora-maxwell

# 2. 升级（官方支持 8.0→8.4 in-place）
docker compose -f docker-compose-go.yml up -d mysql
# 首次启动会自动升级系统表，观察日志
docker logs -f aurora-mysql  # 等待 "ready for connections"

# 3. 检查 mysql_native_password 用户（8.4 默认禁用）
docker exec aurora-mysql mysql -uroot -p${AURORA_MYSQL_PASSWORD} \
  -e "SELECT user,host,plugin FROM mysql.user WHERE plugin='mysql_native_password';"
# 若有，迁移：ALTER USER 'root'@'%' IDENTIFIED WITH caching_sha2_password BY 'password';
# aurora-go 用 go-sql-driver v1.10 已支持 caching_sha2_password，无需改代码

# 4. 起回依赖服务
docker compose -f docker-compose-go.yml up -d aurora-go aurora-maxwell
```

### 1.3 RabbitMQ 3.11 → 4.3.6

```bash
# 博客消息瞬态（maxwell 同步 + 异步任务），停机升级
docker stop aurora-go aurora-maxwell

# 3.x → 4.x 数据目录可能不兼容，清空重建（博客无持久化消息依赖）
docker stop aurora-rabbitmq && docker rm aurora-rabbitmq
rm -rf /opt/aurora/rabbitmq/data/*  # 保留目录清数据

docker compose -f docker-compose-go.yml up -d rabbitmq
# 客户端 amqp091-go 已升 v1.15.0 配套 4.x
docker compose -f docker-compose-go.yml up -d aurora-go aurora-maxwell
```

## 二、P1+P2：Go 1.27 + 基础镜像

```bash
# 本地/CI 用 golang:1.27-alpine 构建二进制
docker build -t aurora-server:1.27 -f Dockerfile .
# 或本地装 Go 1.27.1 编译（GOTOOLCHAIN=auto 自动下载或手动装）

# 部署：替换二进制 + 重启
docker compose -f docker-compose-go.yml up -d aurora-go

# ES 8.19.21 / Redis 8.2.9-alpine / Nginx 1.28.0-alpine / Alpine 3.22 同步重启
docker compose -f docker-compose-go.yml up -d elasticsearch redis nginx
```

## 三、验证清单

```bash
# 1. aurora-go 启动横幅全绿
docker logs aurora-go 2>&1 | grep -E "MySQL|Redis|RabbitMQ|Elasticsearch|Storage|Email"

# 2. 接口冒烟
curl -s http://localhost:8080/health
curl -s http://localhost:8080/api/articles?page=1&page_size=1

# 3. 上传测试（验证 RustFS）
curl -X POST http://localhost:8080/api/upload -F "file=@test.jpg"
# 浏览器访问返回的图片 URL

# 4. ES 搜索
curl -s "http://localhost:9200/aurora_articles/_count"
# 5. RabbitMQ 管理台 http://IP:15672
```

## 四、回滚预案

| 故障 | 回滚动作 |
|:---|:---|
| RustFS 异常 | 起回老 MinIO 容器（数据卷 /minio/data 未删），停 RustFS |
| MySQL 8.4 不兼容 | 用备份 SQL 恢复到全新 8.0 容器 |
| RabbitMQ 4.x 客户端报错 | 回退 amqp091-go v1.12.0 + rabbitmq:3.13.7-management |
| Go 1.27 编译失败 | go.mod 回退 go 1.26，Dockerfile 回 golang:1.26-alpine |
| ES 升级失败 | 删索引重灌（数据源 MySQL，maxwell/同步任务重建）|

**天然安全垫**：aurora-go 自带 DeleteIndex + 全量重灌能力，
ES 最坏情况直接删索引重建（数据源在 MySQL）。

## 五、Maxwell 风险（MySQL 8.4）

zendesk/maxwell 维护基本停滞，1.41.x 对 MySQL 8.4 binlog
兼容性需重点验证。若 maxwell 异常，备选：
1. 用 debezium/connectors 替代
2. 改用 MySQL 触发器 + Go 消费者
3. 降级 MySQL 到 8.0.46（8.0 最后版本，仍 EOL 但 maxwell 稳定）

**验证命令**：升级后观察 maxwell 日志是否正常消费 binlog
```bash
docker logs -f aurora-maxwell --tail 50
```
