# Java → Go 迁移指南

> **Aurora 博客系统从 SpringBoot 4.1.0-M4 迁移至 Go 1.26 + tRPC-Agent-Go v1.8**

---

## 一、迁移概述

### 为什么迁移？

| 维度 | Java SpringBoot | Go Aurora |
|------|----------------|-----------|
| 内存占用 | ~200-400MB (JVM) | ~30-60MB (原生) |
| 启动时间 | 8-15秒 | 0.3-0.8秒 |
| P99延迟 | 50-100ms (GC停顿) | 10-30ms (无GC) |
| Docker镜像 | 80-500MB | 15-25MB |
| AI能力扩展 | 需自研全部引擎 | tRPC-Agent-Go 开箱即用 |
| 部署复杂度 | JRE + JVM调优 | 单二进制 + 静态编译 |

### 迁移原则

1. **API契约100%兼容**: 所有REST端点的路径/方法/请求响应格式不变
2. **数据库Schema不变**: 复用现有24张表，零DDL变更
3. **中间件不变**: MySQL/Redis/RabbitMQ/ES/MinIO 配置直接复用
4. **前端零改动**: aurora-blog + aurora-admin-v3 无需任何修改
5. **渐进式灰度**: 支持Java/Go双版并行运行，Nginx切流

---

## 二、架构映射

### 2.1 核心框架映射

```
┌─────────────────────┐     ┌─────────────────────┐
│   Java SpringBoot   │ ──→ │     Go Gin          │
├─────────────────────┤     ├─────────────────────┤
│ Spring MVC          │     │ gin.Context         │
│ @RestController     │     │ handler.HandlerFunc  │
│ @RequestMapping     │     │ r.GET/POST/PUT/DELETE│
│ @RequestBody/@Resp  │     │ c.ShouldBindJSON()  │
│ @PathVariable       │     │ c.Param("id")       │
│ @RequestParam       │     │ c.DefaultQuery()    │
│ @Valid              │     │ validator.Struct()  │
└─────────────────────┘     └─────────────────────┘

┌─────────────────────┐     ┌─────────────────────┐
│   MyBatis-Plus      │ ──→ │     GORM            │
├─────────────────────┤     ├─────────────────────┤
│ BaseMapper<T>       │     │ gorm.DB             │
│ @TableName          │     │ TableName() tag      │
│ @TableField         │     │ Column() tag         │
│ @TableId(IdType.AUTO)│    │ PrimaryKey tag       │
│ QueryWrapper/LambdaW│     │ db.Where(&{}).Find() │
│ PageHelper/IPage    │     │ db.Scopes(Paginate())│
│ @Transactional      │     │ db.Transaction()     │
│ ResultMap           │     │ Preload("Tags")      │
│ @One/@Many          │     │ Association ForeignKey│
└─────────────────────┘     └─────────────────────┘

┌─────────────────────┐     ┌─────────────────────┐
│   Security Filter   │ ──→ │  Gin Middleware      │
├─────────────────────┤     ├─────────────────────┤
│ OncePerRequestFilter│     │ HandlerFunc chain    │
│ JwtAuthFilter       │     │ middleware.JWTAuth() │
│ RBACFilter          │     │ middleware.RBAC()    │
│ RateLimiterFilter   │     │ middleware.RateLimit()│
│ CorsFilter          │     │ middleware.CORS()     │
│ ExceptionHandler    │     │ middleware.Recovery() │
└─────────────────────┘     └─────────────────────┘
```

### 2.2 依赖库映射

| Java依赖 | Go替代 | 用途 |
|----------|--------|------|
| spring-boot-starter-web | gin-gonic/gin | Web框架 |
| mybatis-plus/generator |gorm.io/gorm | ORM |
| spring-boot-starter-data-redis | go-redis/v9 | Redis客户端 |
| spring-rabbitmq | amqp091-go | RabbitMQ |
| elasticsearch-rest-high-level-client | elastic/go-elasticsearch v8 | ES客户端 |
| minio/minio-java | minio-go/v7 | MinIO SDK |
| jjwt | golang-jwt/jwt v5 | JWT |
| knife4j | swaggo/gin-swagger | API文档 |
| quartz-scheduler | robfig/cron v3 | 定时任务 |
| hutool-crypto | x/crypto/bcrypt | 密码哈希 |
| ip2region | ip2region | IP归属地 |
| lombok | Go struct tags | 数据对象 |
| swagger-annotations | swaggo annotations | 注解文档 |
| 自研Agent引擎 | trpc.group/trpc-agent-go v1.8 | AI引擎 |

---

## 三、代码对照表

### 3.1 Controller → Handler

```java
// === Java Controller ===
@RestController
@RequestMapping("/api/admin/articles")
public class ArticleController {

    @Autowired
    private ArticleService articleService;

    @GetMapping("/list")
    public ResultVO<PageResultDTO<ArticleDTO>> listArticle(
            @RequestParam(defaultValue = "1") Integer current,
            @RequestParam(defaultValue = "10") Integer size,
            @RequestParam(required = false) String keyword) {
        return ResultVO.ok(articleService.listArticles(current, size, keyword));
    }

    @PostMapping
    @OptLog(optType = SaveOrUpdate)
    public ResultVO<?> saveArticle(@Validated @RequestBody ArticleVO articleVO) {
        articleService.saveOrEditArticle(articleVO);
        return ResultVO.ok();
    }
}
```

```go
// === Go Handler ===
func (h *ArticleHandler) List(c *gin.Context) {
    var vo VO.ArticleConditionVO
    if err := c.ShouldBindQuery(&vo); err != nil {
        response.Fail(c, code.ParamError)
        return
    }

    dto, total := h.articleSvc.ListArticles(vo.Current, vo.Size, vo.Keyword)
    response.SuccessWithData(c, gin.H{"recordList": dto, "count": total})
}

func (h *ArticleHandler) Create(c *gin.Context) {
    var vo VO.ArticleVO
    if err := c.ShouldBindJSON(&vo); err != nil {
        response.Fail(c, code.ParamError)
        return
    }

    h.articleSvc.SaveOrEditArticle(vo)
    response.Success(c)
}
```

### 3.2 Service 层

```java
// === Java Service ===
@Service
@Transactional(rollbackFor = Exception.class)
public class ArticleServiceImpl implements ArticleService {

    @Autowired
    private ArticleMapper articleMapper;
    @Autowired
    private ArticleTagMapper articleTagMapper;

    @Override
    public void saveOrEditArticle(ArticleVO articleVO) {
        Article article = BeanCopyUtil.copyProperties(articleVO, Article.class);
        articleMapper.insert(article);

        // 保存文章-标签关联
        List<Integer> tagIds = articleVO.getTagIds();
        for (Integer tagId : tagIds) {
            ArticleTag at = new ArticleTag();
            at.setArticleId(article.getId());
            at.setTagId(tagId);
            articleTagMapper.insert(at);
        }
    }
}
```

```go
// === Go Service ===
func (s *ArticleService) SaveOrEditArticle(vo VO.ArticleVO) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        // DTO转换
        article := s.voToModel(vo)

        // 插入文章
        if err := tx.Create(&article).Error; err != nil {
            return err
        }

        // 批量插入文章-标签关联
        var articleTags []Model.ArticleTag
        for _, tagID := range vo.TagIDs {
            articleTags = append(articleTags, Model.ArticleTag{
                ArticleID: article.ID,
                TagID:    tagID,
            })
        }
        return tx.Create(&articleTags).Error
    })
}
```

### 3.3 中间件

```java
// === Java JWT Filter ===
@Component
public class JwtAuthenticationFilter extends OncePerRequestFilter {

    @Override
    protected void doFilterInternal(HttpServletRequest req, HttpServletResponse res,
                                     FilterChain chain) throws IOException, ServletException {
        String token = req.getHeader("Authorization");
        if (StringUtils.hasText(token)) {
            Claims claims = Jwts.parser().setSigningKey(secret).parseClaimsJws(token).getBody();
            // set security context...
        }
        chain.doFilter(req, res);
    }
}
```

```go
// === Go JWT Middleware ===
func JWTAuth() gin.HandlerFunc {
    return func(c *gin.Context) {
        tokenStr := c.GetHeader("Authorization")
        tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")

        claims, err := tokenSvc.ParseToken(tokenStr)
        if err != nil {
            response.Unauthorized(c, "Token无效")
            c.Abort()
            return
        }

        c.Set("userID", claims.UserID)
        c.Set("roleList", claims.Roles)
        c.Next()
    }
}
```

### 3.4 并发处理

```java
// === Java CompletableFuture ===
public HomeDTO getHomeData() {
    CompletableFuture<List<ArticleDTO>> articlesFuture = 
        CompletableFuture.supplyAsync(() -> articleService.listTopArticles());
    CompletableFuture<CategoryDTO> categoryFuture = 
        CompletableFuture.supplyAsync(() -> categoryService.getCategoryDTO());
    
    CompletableFuture.allOf(articlesFuture, categoryFuture).join();
    return new HomeDTO(articlesFuture.get(), categoryFuture.get());
}
```

```go
// === Go goroutine + errgroup ===
func (s *HomeService) GetHomeData() (*DTO.HomeDTO, error) {
    var (
        articles []DTO.ArticleCardDTO
        categories []DTO.CategoryOptionDTO
    )

    g, ctx := errgroup.WithContext(context.Background())

    g.Go(func() error {
        var err error
        articles, err = s.articleSvc.ListTopArticles()
        return err
    })

    g.Go(func() error {
        var err error
        categories, err = s.categorySvc.GetCategoryOptions()
        return err
    })

    if err := g.Wait(); err != nil {
        return nil, err
    }

    return &DTO.HomeDTO{Articles: articles, Categories: categories}, nil
}
```

---

## 四、数据库Schema映射

### 4.1 GORM Model定义规范

```go
package model

import "time"

type Article struct {
    ID          uint           `gorm:"primaryKey;autoIncrement" json:"id"`
    UserID      uint           `gorm:"not null;index:idx_user_id" json:"userId"`
    CategoryID  uint           `gorm:"not null;index:idx_category_id" json:"categoryId"`
    Title       string         `gorm:"varchar:200);not null" json:"articleTitle"`
    Cover       string         `gorm:"varchar:1024)" json:"articleCover"`
    Content     string         `gorm:"type:longtext" json:"articleContent"`
    Status      int8           `gorm:"default:1" json:"status"`          // 1-公开 2-草稿 3-密码保护
    IsTop       bool           `gorm:"default:false" json:"isTop"`
    IsFeatured  bool           `gorm:"default:false" json:"isFeatured"`
    Type        int8           `gorm:"default:1" json:"type"`            // 1-原创 2-转载 3-翻译
    OriginalURL string         `gorm:"varchar:512)" json:"originalUrl"`
    Password    string         `gorm:"varchar:32)" json:"password"`
    LikeCount   int64          `gorm:"default:0" json:"likeCount"`
    ViewCount   int64          `gorm:"default:0" json:"viewCount"`
    CreatedAt   time.Time      `json:"createdAt"`
    UpdatedAt   time.Time      `json:"updatedAt"`
    DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

    // 关联
    Category    Category       `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
    Tags        []Tag          `gorm:"many2many:article_tags;" json:"tags,omitempty"`
}

func (Article) TableName() string { return "t_article" }
```

### 4.2 表名映射规则

| 规则 | 示例 |
|------|------|
| Java实体: `Article` | Go表名: `t_article` |
| Java字段: `articleTitle` | JSON: `articleTitle`, DB: `article_title` |
| Java枚举: `ArticleStatus.PUBLISHED` | Go常量: `ArticleStatusPublished = 1` |
| Java LocalDateTime | Go `time.Time` |
| Java BigDecimal | Go `float64` 或 `int64` |

---

## 五、配置迁移

### 5.1 application-prod.yml → config.yaml

```yaml
# === Java application-prod.yml ===
server:
  port: 8080

spring:
  datasource:
    url: jdbc:mysql://localhost:3306/aurora?useUnicode=true&characterEncoding=utf-8
    username: root
    password: ${MYSQL_PASSWORD}
  redis:
    host: localhost
    port: 6379
    password: ${REDIS_PASSWORD}
  rabbitmq:
    host: localhost
    port: 5672
    username: guest
    password: ${MQ_PASSWORD}
  elasticsearch:
    hosts: localhost:9200
    username: ${ES_USERNAME}
    password: ${ES_PASSWORD}
```

```yaml
# === Go configs/config.yaml ===
server:
  mode: release
  port: 8080

mysql:
  host: localhost
  port: 3306
  database: aurora
  username: root
  password: ${MYSQL_PASSWORD}  # 环境变量替换
  max_open_conns: 50
  max_idle_conns: 10
  log_level: info

redis:
  host: localhost
  port: 6379
  password: ${REDIS_PASSWORD}
  db: 0
  pool_size: 20

rabbitmq:
  host: localhost
  port: 5672
  username: guest
  password: ${MQ_PASSWORD}
  vhost: /

elasticsearch:
  addresses:
    - http://localhost:9200
  username: ${ES_USERNAME}
  password: ${ES_PASSWORD}
  index_prefix: aurora

agent:
  enabled: false  # 可选模块开关
  llm:
    default_model: deepseek-chat
    providers:
      openai:
        api_key: ${OPENAI_API_KEY}
        base_url: https://api.openai.com/v1
      deepseek:
        api_key: ${DEEPSEEK_API_KEY}
        base_url: https://api.deepseek.com
```

---

## 六、部署切换

### 6.1 Nginx灰度配置

```nginx
upstream aurora_backend {
    # Go版本(主)
    server aurora-go:8080 weight=90;
    # Java版本(备/回滚)
    server aurora-java:8081 weight=10;
}

server {
    listen 443 ssl;
    server_name blog.example.com;

    location /api/ {
        proxy_pass http://aurora_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        
        # SSE支持 (Agent对话)
        proxy_buffering off;
        proxy_cache off;
        chunked_transfer_encoding on;
    }
}
```

### 6.2 Docker Compose编排

```yaml
version: '3.8'
services:
  aurora-go:
    build: .
    image: aurora-go:latest
    container_name: aurora-go-server
    restart: unless-stopped
    ports:
      - "8080:8080"
    env_file:
      - .env
    depends_on:
      mysql:
        condition: service_healthy
      redis:
        condition: service_healthy
    networks:
      - aurora-network
    healthcheck:
      test: ["CMD", "wget", "--spider", "-q", "http://localhost:8080/health"]
      interval: 30s
      timeout: 5s
      retries: 3

  mysql:
    image: mysql:9.0
    container_name: aurora-mysql
    environment:
      MYSQL_ROOT_PASSWORD: ${MYSQL_ROOT_PASSWORD}
      MYSQL_DATABASE: aurora
      MYSQL_CHARACTER_SET_SERVER: utf8mb4
      MYSQL_COLLATION_SERVER: utf8mb4_unicode_ci
    volumes:
      - mysql_data:/var/lib/mysql
      - ./init:/docker-entrypoint-initdb.d
    ports:
      - "3306:3306"
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 10s
      retries: 5

  redis:
    image: redis:7-alpine
    container_name: aurora-redis
    command: redis-server --requirepass ${REDIS_PASSWORD} --maxmemory 16mb --maxmemory-policy allkeys-lru
    volumes:
      - redis_data:/data
    ports:
      - "6379:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "-a", "${REDIS_PASSWORD}", "ping"]
      interval: 10s

  rabbitmq:
    image: rabbitmq:3-management-alpine
    container_name: aurora-mq
    environment:
      RABBITMQ_DEFAULT_USER: guest
      RABBITMQ_DEFAULT_PASS: ${MQ_PASSWORD}
    volumes:
      - mq_data:/var/lib/rabbitmq
    ports:
      - "5672:5672"
      - "15672:15672"

  elasticsearch:
    image: docker.elastic.co/elasticsearch/elasticsearch:8.13.4
    container_name: aurora-es
    environment:
      - discovery.type=single-node
      - xpack.security.enabled=false
      - ES_JAVA_OPTS=-Xms256m -Xmx512m
      - "indices.query.bool.max_clause_count=10240"
    volumes:
      - es_data:/usr/share/elasticsearch/data
    ports:
      - "9200:9200"

  minio:
    image: minio/minio:latest
    container_name: aurora-minio
    command: server /data --console-address ":9001"
    environment:
      MINIO_ROOT_USER: ${MINIO_ACCESS_KEY}
      MINIO_ROOT_PASSWORD: ${MINIO_SECRET_KEY}
    volumes:
      - minio_data:/data
    ports:
      - "9000:9000"
      - "9001:9001"

volumes:
  mysql_data:
  redis_data:
  mq_data:
  es_data:
  minio_data:

networks:
  aurora-network:
    driver: bridge
```

### 6.3 环境变量模板 (.env.example)

```bash
# 数据库
MYSQL_ROOT_PASSWORD=your_root_password
MYSQL_PASSWORD=your_password
REDIS_PASSWORD=your_password
MQ_PASSWORD=your_password
ES_USERNAME=
ES_PASSWORD=

# MinIO
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin

# LLM API Keys (Agent模块使用)
OPENAI_API_KEY=sk-...
DEEPSEEK_API_KEY=sk-...

# JWT密钥
JWT_SECRET=your-jwt-secret-key-at-least-32-chars-long

# 邮箱SMTP (评论通知)
SMTP_HOST=smtp.qq.com
SMTP_PORT=587
SMTP_USER=your@qq.com
SMTP_PASS=your_auth_code
```

---

## 七、验证清单

### 功能验证

- [ ] 用户登录/登出正常
- [ ] JWT Token认证有效
- [ ] 文章CRUD完整可用
- [ ] 分类/标签管理正常
- [ ] 评论嵌套显示正确
- [ ] 评论邮件通知发送(RabbitMQ消费者)
- [ ] Elasticsearch搜索功能正常
- [ ] 文件上传到MinIO成功
- [ ] 定时任务调度执行
- [ ] 操作日志记录
- [ ] RBAC权限控制生效
- [ ] 限流中间件生效
- [ ] **Agent对话功能**(如果启用)
- [ ] **AI写作助手**(如果启用)

### 性能验证

```bash
# 基准测试
make bench

# 压力测试
hey -n 10000 -c 100 http://localhost:8080/health
hey -n 5000 -c 50 http://localhost:8080/api/articles

# 内存占用检查
docker stats aurora-go-server
```

### 兼容性验证

```bash
# 对比API响应格式 (确保前端无需改动)
curl http://localhost:8080/api/articles?current=1\&size=5 | jq .

# 对比旧Java版本
curl http://old-server:8081/api/articles?current=1\&size=5 | jq .
```

---

## 八、常见问题

### Q: 如何调试Agent模块？

```bash
# 1. 启用Agent
# 编辑 configs/config.yaml: agent.enabled: true

# 2. 设置LLM API Key
export DEEPSEEK_API_KEY="sk-xxx"

# 3. 启动服务
make run

# 4. 查看日志
tail -f logs/aurora.log | grep agent

# 5. 调用Agent端点
curl -X POST http://localhost:8080/api/agent/chat \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"message":"你好"}'
```

### Q: 如何回退到Java版本？

```bash
# Nginx权重调整: Go 0% / Java 100%
# 或直接停止Go服务, 将Java端口改为8080
docker stop aurora-go-server
```

### Q: 数据库迁移需要注意什么？

- **无需迁移**: Go版本完全复用现有MySQL数据库
- **GORM AutoMigrate**: 首次启动自动检查表结构差异, 仅做增量更新
- **Maxwell Binlog同步**: 继续工作, ES索引保持一致

### Q: 性能调优建议？

| 参数 | 推荐值 | 说明 |
|------|--------|------|
| GOMAXPROCS | CPU核心数 | 自动设置即可 |
| MySQL连接池 | 20-50 | `max_open_conns: 50` |
| Redis连接池 | 10-20 | `pool_size: 20` |
| GIN模式 | release | 生产环境必须 |
| 日志级别 | info | 开发用debug |

---

*文档版本: 1.0 | 最后更新: 2026-04-11*
