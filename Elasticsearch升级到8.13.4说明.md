# Elasticsearch 8.13.4 升级说明

## 升级日期
2026-03-07

## 升级内容

### 1. 依赖升级

#### pom.xml 变更
- 移除 `elasticsearch-rest-high-level-client:7.17.15`（已弃用）
- 保留 `elasticsearch-rest-client:8.13.4`
- 保留 `elasticsearch-java:8.13.4`
- 添加必要的 Jackson 和 Jakarta JSON 依赖

### 2. 代码重构

#### ElasticsearchConfig.java
- 从 `RestHighLevelClient` 迁移到新的 `ElasticsearchClient`
- 使用 `RestClient` 和 `RestClientTransport` 构建客户端
- 配置方式保持不变（用户名密码认证）

**新Bean定义**：
```java
@Bean(destroyMethod = "close")
public RestClient restClient() { ... }

@Bean(destroyMethod = "close")
public ElasticsearchTransport elasticsearchTransport(RestClient restClient) { ... }

@Bean(destroyMethod = "close")
public ElasticsearchClient elasticsearchClient(ElasticsearchTransport transport) { ... }
```

#### EsSearchStrategyImpl.java
- 使用新的 `co.elastic.clients.elasticsearch` API
- 查询构建方式改变：
  - `QueryBuilders` → `Query.of()`
  - `BoolQuery` → `BoolQuery.Builder`
  - `MatchQuery` → `MatchQuery.of()`
- 高亮处理方式改变：
  - `HighlightBuilder` → `SearchRequest` 中的 `highlight()` 方法
  - 高亮结果获取方式：`hit.highlight().get("fieldName")`

**主要变更**：
- 搜索请求构建：`SearchRequest.of(s -> s.index("article").query(query))`
- 查询执行：`elasticsearchClient.search(request, ArticleSearchDTO.class)`
- 结果处理：使用响应式API获取搜索结果

#### ArticleSearchDTO.java
- 移除 Spring Data ES 注解（`@Document`, `@Field`, `@Id`）
- 使用 Jackson `@JsonProperty` 注解进行字段映射
- 纯POJO类，无ES特定依赖

### 3. 配置文件变更

#### application-prod.yml
配置保持不变：
```yaml
elasticsearch:
  rest:
    uris: http://134.175.206.158:9200
    username: user_admin
    password: user_aqi125.cn
```

### 4. 部署文档更新

#### docker-compose.yml
- ES镜像版本：`elasticsearch:7.17.12` → `elasticsearch:8.13.4`
- 其他配置保持不变

#### IK分词器
- 下载地址：`elasticsearch-analysis-ik-8.13.4.zip`
- 版本要求：必须与ES版本一致

### 5. 兼容性说明

#### 保留的兼容性
- `MaxWellConsumer.java` 无需修改（使用Repository接口）
- `ElasticsearchMapper.java` 无需修改（Spring Data ES Repository自动适配）
- `ArticleSearchRepository.java` 无需修改

#### API变化
- 所有新的查询使用 `co.elastic.clients.elasticsearch` 包
- 旧的 `org.elasticsearch.index.query` 包不再使用
- 新API使用Builder模式，更类型安全

## 部署步骤

### 1. 服务器端更新
```bash
# 1. 停止当前ES容器
docker stop aurora-elasticsearch

# 2. 备份数据
docker cp aurora-elasticsearch:/usr/share/elasticsearch/data /backup/es-data

# 3. 更新docker-compose.yml中的ES镜像版本
# image: elasticsearch:8.13.4

# 4. 重新拉取镜像
docker-compose pull elasticsearch

# 5. 启动新版本ES
docker-compose up -d elasticsearch

# 6. 重新安装IK分词器
docker exec -it aurora-elasticsearch /bin/bash
./bin/elasticsearch-plugin install https://release.infinilabs.com/analysis-ik/stable/elasticsearch-analysis-ik-8.13.4.zip
exit
docker restart aurora-elasticsearch

# 7. 重新初始化密码（如果需要）
docker exec aurora-elasticsearch /usr/share/elasticsearch/bin/elasticsearch-setup-passwords auto -b
```

### 2. 后端代码部署
```bash
# 1. 打包新版本
cd aurora-springboot
mvn clean package -DskipTests

# 2. 上传jar包到服务器

# 3. 重新构建并启动容器
docker stop aurora-springboot
docker rm aurora-springboot
cd /opt/aurora/app
docker build -t aurora-springboot:latest .
docker run -d --name aurora-springboot \
  --network aurora_aurora-network \
  -p 8080:8080 \
  --restart=always \
  aurora-springboot:latest

# 4. 查看日志
docker logs -f aurora-springboot
```

## 注意事项

### 1. 数据迁移
- ES 7.x 到 8.x 通常无需特殊迁移
- 建议先在测试环境验证
- 生产环境升级前务必备份数据

### 2. API兼容性
- 新的 `ElasticsearchClient` 与旧的 `RestHighLevelClient` 不兼容
- 所有使用旧API的代码必须重写
- Repository接口保持兼容

### 3. 性能影响
- ES 8.x 性能优于 7.x
- 内存占用可能略有增加（建议保持 `-Xms96m -Xmx192m`）
- 索引格式可能需要重建（首次启动时）

### 4. 安全特性
- ES 8.x 安全特性增强
- 默认启用安全认证
- SSL/TLS配置需要额外注意

## 测试验证

### 1. 连接测试
```bash
curl -u user_admin:user_aqi125.cn http://134.175.206.158:9200
```

### 2. 索引检查
```bash
curl -u user_admin:user_aqi125.cn http://134.175.206.158:9200/_cat/indices?v
```

### 3. IK分词器测试
```bash
curl -u user_admin:user_aqi125.cn -X POST "http://134.175.206.158:9200/_analyze" \
  -H 'Content-Type: application/json' \
  -d '{
    "tokenizer":"ik_smart",
    "text":"我爱技术"
  }'
```

### 4. 搜索功能测试
- 访问博客前台搜索功能
- 测试中文分词
- 验证高亮显示

## 回滚方案

如果升级失败，可以回滚到7.17.12：

```bash
# 1. 停止ES容器
docker stop aurora-elasticsearch
docker rm aurora-elasticsearch

# 2. 恢复旧版本镜像
# 修改docker-compose.yml: image: elasticsearch:7.17.12

# 3. 启动旧版本
docker-compose up -d elasticsearch

# 4. 恢复数据（如需要）
docker cp /backup/es-data aurora-elasticsearch:/usr/share/elasticsearch/data
docker restart aurora-elasticsearch

# 5. 重新安装旧版本IK分词器
docker exec -it aurora-elasticsearch /bin/bash
./bin/elasticsearch-plugin install https://release.infinilabs.com/analysis-ik/stable/elasticsearch-analysis-ik-7.17.12.zip
exit
docker restart aurora-elasticsearch
```

## 参考资料

- [Elasticsearch 8.13.4 Release Notes](https://www.elastic.co/guide/en/elasticsearch/reference/current/release-notes-8.13.4.html)
- [Java API Client 8.x 迁移指南](https://www.elastic.co/guide/en/elasticsearch/client/java-api-client/current/migrate-hlrc.html)
- [IK分词器 GitHub](https://github.com/medcl/elasticsearch-analysis-ik)

## 总结

本次升级将Elasticsearch从7.17.12升级到8.13.4，主要变化包括：
1. ✅ 移除已弃用的RestHighLevelClient
2. ✅ 采用新的ElasticsearchClient API
3. ✅ 更新相关配置文件
4. ✅ 保持数据兼容性
5. ✅ 提升性能和安全性

升级完成后，建议：
- 监控ES性能指标
- 检查日志是否有异常
- 验证搜索功能正常
- 备份生产配置和数据
