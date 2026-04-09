package com.aurora;

import lombok.extern.slf4j.Slf4j;
import org.mybatis.spring.annotation.MapperScan;
import org.springframework.boot.*;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.context.ConfigurableApplicationContext;
import org.springframework.context.annotation.Bean;
import org.springframework.core.env.Environment;
import org.springframework.web.client.RestTemplate;

import java.net.InetAddress;
import java.net.URI;
import java.util.LinkedHashMap;
import java.util.Map;
import java.util.concurrent.TimeUnit;

@SpringBootApplication(exclude = {
    com.baomidou.mybatisplus.autoconfigure.MybatisPlusAutoConfiguration.class
})
@MapperScan("com.aurora.mapper")
@Slf4j
public class AuroraSpringbootApplication implements ApplicationRunner {

    private final ConfigurableApplicationContext ctx;

    public AuroraSpringbootApplication(ConfigurableApplicationContext ctx) {
        this.ctx = ctx;
    }

    public static void main(String[] args) {
        long startTime = System.currentTimeMillis();

        ConfigurableApplicationContext ctx =
                SpringApplication.run(AuroraSpringbootApplication.class, args);

        long elapsed = System.currentTimeMillis() - startTime;
        printStartupSummary(ctx, elapsed);
    }

    @Bean
    public RestTemplate restTemplate() {
        return new RestTemplate();
    }

    @Override
    public void run(ApplicationArguments args) {
        diagnoseExternalServices();
    }

    /**
     * 打印启动摘要信息（ASCII 表格风格）
     */
    private static void printStartupSummary(ConfigurableApplicationContext ctx, long elapsedMs) {
        Environment env = ctx.getEnvironment();

        String appName  = env.getProperty("spring.application.name", "Aurora");
        String port     = env.getProperty("server.port", "8080");
        String context  = env.getProperty("server.servlet.context-path", "");
        String profile  = env.getActiveProfiles().length > 0
                            ? String.join(",", env.getActiveProfiles())
                            : "default";

        String hostAddress;
        try {
            hostAddress = InetAddress.getLocalHost().getHostAddress();
        } catch (Exception e) {
            hostAddress = "unknown";
        }

        String elapsedStr = elapsedMs >= 1000
                ? String.format("%.2fs", elapsedMs / 1000.0)
                : elapsedMs + "ms";

        log.info("""
                ┌─────────────────────────────────────────────────────────────┐
                │                     {} 启动完成                           │
                ├──────────────────┬────────────────────────────────────────┤
                │   profiles       │  {}                                    │
                │   port           │  {}{}                                   │
                │   address        │  http://{}:{}{}                        │
                │   elapsed        │  {}                                    │
                ├──────────────────┴────────────────────────────────────────┤
                │   Swagger UI      │  http://{}:{}{}/swagger-ui.html       │
                │   API Docs        │  http://{}:{}{}/v3/api-docs            │
                └─────────────────────────────────────────────────────────────┘
                """,
                appName, profile,
                port, context.isEmpty() ? "" : " (context: " + context + ")",
                hostAddress, port, context, elapsedStr,
                hostAddress, port, context,
                hostAddress, port, context);
    }

    /**
     * 诊断外部服务连接状态（后台异步执行，不阻塞主线程）
     * 所有地址以 application-prod.yml 配置为准，
     * 通过 Spring Environment 统一读取，自动支持环境变量覆盖。
     *
     * Aurora 中间件清单：
     *   - MySQL    → spring.datasource.url
     *   - Redis     → spring.data.redis.host / port
     *   - RabbitMQ → spring.rabbitmq.host / port
     *   - ES       → elasticsearch.rest.uris
     *   - MinIO    → upload.minio.endpoint
     */
    private void diagnoseExternalServices() {
        Environment env = ctx.getEnvironment();

        Thread.startVirtualThread(() -> {
            try {
                TimeUnit.SECONDS.sleep(3);
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
                return;
            }

            Map<String, String> results = new LinkedHashMap<>();
            results.put("MySQL",    checkDataSource(env));
            results.put("Redis",    checkRedis(env));
            results.put("RabbitMQ", checkRabbitMQ(env));
            results.put("ES",       checkElasticsearch(env));
            results.put("MinIO",    checkPropertyEndpoint(env, "upload.minio.endpoint", "http://localhost:9000"));

            log.info("┌─ 外部服务诊断 ──────────────────────────────────────────┐");
            results.forEach((service, status) ->
                    log.info("│  {} : {}",
                            String.format("%-8s", service), status));
            log.info("└───────────────────────────────────────────────────────┘");
        });
    }

    // ==================== 各中间件检测实现 ====================

    /** 从 JDBC URL 解析 MySQL host:port */
    private String checkDataSource(Environment env) {
        String jdbcUrl = env.getProperty("spring.datasource.url", "");
        try {
            URI uri = URI.create(jdbcUrl.substring(5)); // 去掉 "jdbc:"
            return ping(uri.getHost(), uri.getPort() > 0 ? uri.getPort() : 3306);
        } catch (Exception e) {
            return "\u274C PARSE FAIL \u2014 " + e.getMessage();
        }
    }

    /** 从 Redis 属性读取 */
    private String checkRedis(Environment env) {
        String host = env.getProperty("spring.data.redis.host", "localhost");
        int port = env.getProperty("spring.data.redis.port", Integer.class, 6379);
        return ping(host, port);
    }

    /** 从 RabbitMQ 属性读取 */
    private String checkRabbitMQ(Environment env) {
        String host = env.getProperty("spring.rabbitmq.host", "localhost");
        int port = env.getProperty("spring.rabbitmq.port", Integer.class, 5672);
        return ping(host, port);
    }

    /** 从 Elasticsearch URIs 属性读取 */
    private String checkElasticsearch(Environment env) {
        String uris = env.getProperty("elasticsearch.rest.uris", "");
        if (!uris.isEmpty()) {
            try {
                URI uri = URI.create(uris.split(",")[0].trim());
                return ping(uri.getHost(), uri.getPort() > 0 ? uri.getPort() : 9200);
            } catch (Exception ignored) { }
        }
        return ping("localhost", 9200);
    }

    /** 从通用 URL 类端点属性读取（MinIO 等） */
    private String checkPropertyEndpoint(Environment env, String key, String defaultUrl) {
        String urlStr = env.getProperty(key, defaultUrl);
        try {
            URI uri = URI.create(urlStr);
            int port = uri.getPort() > 0 ? uri.getPort()
                    : "https".equals(uri.getScheme()) ? 443 : 80;
            return ping(uri.getHost(), port);
        } catch (Exception e) {
            return "\u274C PARSE FAIL \u2014 " + e.getMessage();
        }
    }

    /** TCP 端口探测 */
    private String ping(String host, int port) {
        long start = System.currentTimeMillis();
        try {
            URI.create("http://" + host + ":" + port + "/").toURL().openConnection().connect();
            return "\u2705 OK  (" + (System.currentTimeMillis() - start) + "ms)";
        } catch (Exception e) {
            return "\u274C FAIL \u2014 " + e.getMessage();
        }
    }
}
