# 中间件升级服务器操作手册

> **✅ 状态：已全部执行完成（2026-09-20 ~ 09-22）**，线上稳定运行 48h+
> 适用提交：c3dc840(P0) / 23cfdcf(P1) / 7470730(P2) / c24d472(运维归档)
> 目标服务器：134.175.206.158（广州 ap-guangzhou-3，4C4G/70GB）
>
> ## 执行结果一览
>
> | 组件 | 升级前 | 升级后 | 备注 |
> |:---|:---|:---|:---|
> | MySQL | 8.0.32 (EOL) | **8.4.10 LTS** | 系统表自动升级，数据无损 |
> | RabbitMQ | 3.11.9 (EOL) | **4.3.6** | 数据目录轮换重建 |
> | 对象存储 | MinIO 2023.12.7 (AGPL) | **RustFS 1.0.0 (Apache-2.0)** | 319 物件 mc mirror 校验一致 |
> | Elasticsearch | 8.19.14 (RCE) | **9.5.3 + IK 9.5.3 官方版** | 集群 green，数据无损 |
> | Redis | redis-stack:latest | **redis:8.2.9-alpine** | 固定 tag |
> | Nginx | 1.23.3 | **1.28.0-alpine** | 动态 DNS 反代（防 502） |
> | Go 基座 | alpine:3.20 | **alpine:3.22** | 二进制无感更新 |
> | Go 工具链 | 1.26 | **1.27** + 8 依赖升级 | build/vet/test 三绿 |
>
> 稳定性验证：Maxwell 连续消费 MySQL 8.4 binlog 48h+、aurora-go 零 ERROR、全链路诊断 8/8 绿。

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

# 6. ⚠️ 必做：补设 bucket 公开读策略（mc mirror 不迁移 policy！）
#    漏掉此步 → 匿名访问图片全部 403，前端图片全裂（实际踩过的坑）
docker run --rm --network aurora-network bitnami/minio:2023.12.7 \
  bash -c "mc alias set dst http://aurora-rustfs:9000 ${MINIO_ROOT_USER} ${MINIO_ROOT_PASSWORD} && \
           mc anonymous set download dst/aurora"

# 7. 应用侧 endpoint 指向容器名时，给 RustFS 加 DNS 别名（零配置兼容）
docker network disconnect aurora-network aurora-rustfs
docker network connect --alias aurora-minio aurora-network aurora-rustfs

# 8. 起 aurora-go（endpoint 不变 http://IP:9000，RustFS 端口同 9000）
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

# 3.x → 4.x 数据目录可能不兼容，轮换保留（可回滚，勿直接 rm）
docker stop aurora-go aurora-maxwell aurora-rabbitmq
mv /opt/aurora/rabbitmq/data /opt/aurora/rabbitmq/data.old.$(date +%Y%m%d) && mkdir -p /opt/aurora/rabbitmq/data

docker compose -f docker-compose-go.yml up -d rabbitmq
# 客户端 amqp091-go 已升 v1.15.0 配套 4.x
docker compose -f docker-compose-go.yml up -d aurora-go aurora-maxwell
```

### 1.4 Elasticsearch 8.19 → 9.5.3（P3，大版本）

```bash
# 0. 前置：8.19.x 必须是最新 patch（官方跳板要求），IK 需 9.x 对应版本
#    国内 Docker Hub 不稳时用镜像源：docker.1ms.run/library/elasticsearch:9.5.3

# 1. 用 9.5.3 一次性容器安装 IK 插件并导出（TAT 通道禁外网 wget 时的好办法）
docker run -d --name ik9tmp --entrypoint bash elasticsearch:9.5.3 \
  -c "elasticsearch-plugin install --batch https://get.infini.cloud/elasticsearch/analysis-ik/9.5.3 && sleep 600"
docker cp ik9tmp:/usr/share/elasticsearch/plugins /tmp/ik9/plugins9 && docker rm -f ik9tmp

# 2. 换插件：备份旧目录移出挂载范围（放 plugins/ 内会被当成插件扫描 → duplicate plugin！）
mv /opt/aurora/elasticsearch/plugins/analysis-ik /opt/aurora/elasticsearch/analysis-ik.bak.819
cp -r /tmp/ik9/plugins9/analysis-ik /opt/aurora/elasticsearch/plugins/analysis-ik
cp -r /opt/aurora/elasticsearch/analysis-ik.bak.819/config /opt/aurora/elasticsearch/plugins/analysis-ik/config
chmod -R 755 /opt/aurora/elasticsearch/plugins/analysis-ik

# 3. 改 compose 后重建（数据目录 8.19 → 9.5 自动升级，不可逆，务必先快照）
docker compose -f docker-compose-go.yml up -d elasticsearch --force-recreate
```

> **⚠️ ES 9.x 内存红线（实测四轮实验结论）**
> | heap | 结果 |
> |:---|:---|
> | 96m/160m | ❌ GC 风暴(88条) + master 选举卡死 |
> | 112m/176m | ⚠️ 勉强（启动期 GC 50 条，实占与 192m 相同无收益） |
> | **128m/192m** | ✅ **精简线（当前定稿）** green + 搜索 143ms |
> | 384m | ✅ 推荐线（写峰值余量最足） |
> | 512m | 🟡 能跑但顶格 918MiB |
>
> 同时 `limits.memory` 必须 ≥ heap 的 2 倍（9.x 进程总占用 ≈ heap + ~400M overhead），
> 否则 cgroup OOM kill 循环（Exit 137）。当前定稿：heap 128m/192m + limits 1G。

## 二、P1+P2：Go 1.27 + 基础镜像

```bash
# 本地/CI 用 golang:1.27-alpine 构建二进制
docker build -t aurora-server:1.27 -f Dockerfile .
# 或本地装 Go 1.27.1 编译（GOTOOLCHAIN=auto 自动下载或手动装）

# 部署：替换二进制 + 重启
docker compose -f docker-compose-go.yml up -d aurora-go

# ES 9.5.3 / Redis 8.2.9-alpine / Nginx 1.28.0-alpine / Alpine 3.22 同步重启
docker compose -f docker-compose-go.yml up -d elasticsearch redis nginx

# Nginx 动态 DNS（防容器重建后反代 502）：nginx.conf http 块加
#   resolver 127.0.0.11 valid=10s ipv6=off;
# proxy_pass 变量化：set $up_ago aurora-go:8080; proxy_pass http://$up_ago;
# 线上完整参考配置见 docs/ops/nginx.conf.server
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

## 五、Maxwell × MySQL 8.4（✅ 已验证兼容）

zendesk/maxwell 对 MySQL 8.4 binlog 的兼容性**已在生产验证通过**：
升级后连续消费 48h+ 无异常，日志确认 `Database version: 8.4.10` + `Binlog connected`。

```bash
docker logs aurora-maxwell --tail 10
# 正常特征：BinaryLogClient - Connected to aurora-mysql:3306 at binlog.XXXXX
```

若未来版本升级后 maxwell 异常，备选：
1. 用 debezium/connectors 替代
2. 改用 MySQL 触发器 + Go 消费者

## 六、实际踩坑记录（2026-09-20 ~ 22 生产执行）

| # | 坑 | 现象 | 修复 |
|:--|:---|:---|:---|
| 1 | 升级前内存告急 | 可用仅 434Mi（含 ClickHouse/Kafka 等非博客容器） | 临时停非必需容器腾至 1.6Gi，完成后恢复 |
| 2 | **bucket 策略丢失** | RustFS 迁移后前端图片全 403——`mc mirror` 只迁对象不迁 policy | `mc anonymous set download dst/aurora`（**对象存储迁移 checklist 必须含 policy/CORS**） |
| 3 | 应用存储 DNS 断 | 应用 endpoint 配置指向容器名 `aurora-minio`，新容器叫 rustfs → lookup 失败 | `docker network connect --alias aurora-minio` 零配置兼容 |
| 4 | Nginx 证书 404 循环 | 1.28 重建后 `/etc/nginx/cert-pan` 空（旧容器有 compose 外挂载） | 证书复制进已挂载目录：`cp cert/* conf/cert-pan/` |
| 5 | Nginx 反代 502 | 多次重建 aurora-go 后容器 IP 变化，Nginx 缓存旧 DNS | 根治：resolver 127.0.0.11 + proxy_pass 变量化（见 docs/ops/nginx.conf.server） |
| 6 | ES 9.x duplicate plugin | 备份目录放 plugins/ 挂载内被当插件扫描 | 备份目录移出 plugins/ |
| 7 | ES 9.x 选举卡死 + Exit 137 | heap 160m GC 停顿阻塞选举；limits 640M < 进程需求 ~900M | heap 128m/192m + limits 1G（红线见 1.4 节） |
| 8 | TAT 命令限制 | >2KB/含 curl/外网 wget 被 AccessDeny；输出偶发被吞 | 长命令拆分；每步 compose 操作后必须 docker ps 复查 |
| 9 | GCM 损坏 | git push 静默 128（--version 也秒退） | CredRead 读凭据 + URL 内嵌认证绕过；长期修复 `winget install GitCredentialManager` |
