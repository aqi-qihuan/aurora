package com.aurora.config;

import co.elastic.clients.elasticsearch.ElasticsearchClient;
import com.aurora.entity.Article;
import com.aurora.model.dto.ArticleSearchDTO;
import com.aurora.service.ArticleService;
import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.baomidou.mybatisplus.core.toolkit.CollectionUtils;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.ApplicationArguments;
import org.springframework.boot.ApplicationRunner;
import org.springframework.context.annotation.Profile;
import org.springframework.stereotype.Component;

import java.util.List;

/**
 * Elasticsearch 索引初始化器
 * 自动创建 article 索引并同步已有文章数据
 * 仅在非测试环境下运行
 */
@Slf4j
@Component
@Profile("!test")
public class ElasticsearchIndexInitializer implements ApplicationRunner {

    @Autowired
    private ElasticsearchClient elasticsearchClient;

    @Autowired
    private ArticleService articleService;

    @Override
    public void run(ApplicationArguments args) throws Exception {
        // 检查并创建 article 索引
        initArticleIndex();
        // 同步已有文章数据
        syncExistingArticles();
    }

    /**
     * 初始化 article 索引
     */
    private void initArticleIndex() {
        try {
            // 检查索引是否存在
            boolean exists = elasticsearchClient.indices().exists(b -> b.index("article")).value();
            
            if (exists) {
                log.info("article 索引已存在");
                return;
            }

            log.info("article 索引不存在，开始创建...");
            
            // 创建索引并设置 mapping
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
            
            log.info("article 索引创建成功，已配置 ik_max_word 分词器");
            log.info("========== ES 索引配置信息 ==========");
            log.info("articleTitle: analyzer=ik_max_word, searchAnalyzer=ik_smart");
            log.info("articleContent: analyzer=ik_max_word, searchAnalyzer=ik_smart");
            log.info("====================================");
            
        } catch (Exception e) {
            log.error("初始化 article 索引失败：{}", e.getMessage(), e);
        }
    }

    /**
     * 同步已有的文章数据到 ES
     */
    private void syncExistingArticles() {
        try {
            log.info("开始同步文章数据到 Elasticsearch...");
            
            // 查询所有已发布的文章
            List<ArticleSearchDTO> articles = articleService.list(
                new LambdaQueryWrapper<Article>()
                    .eq(Article::getIsDelete, 0)
                    .eq(Article::getStatus, 1)
            ).stream().map(article -> {
                ArticleSearchDTO dto = new ArticleSearchDTO();
                // 注意：ES 中的字段名是 id，不是 articleId
                dto.setId(article.getId());
                dto.setArticleTitle(article.getArticleTitle());
                dto.setArticleContent(article.getArticleContent());
                dto.setIsDelete(article.getIsDelete());
                dto.setStatus(article.getStatus());
                return dto;
            }).toList();
            
            if (CollectionUtils.isEmpty(articles)) {
                log.info("没有需要同步的文章数据");
                return;
            }
            
            // 批量同步到 ES
            int successCount = 0;
            for (ArticleSearchDTO article : articles) {
                try {
                    elasticsearchClient.index(i -> i
                        .index("article")
                        .id(article.getId().toString())
                        .document(article)
                    );
                    successCount++;
                } catch (Exception e) {
                    log.error("同步文章 ID={} 失败：{}", article.getId(), e.getMessage());
                }
            }
            
            log.info("========== 文章数据同步完成 ==========");
            log.info("应同步文章数：{}", articles.size());
            log.info("实际同步成功：{}", successCount);
            log.info("======================================");
            
        } catch (Exception e) {
            log.error("同步文章数据失败：{}", e.getMessage(), e);
        }
    }
}
