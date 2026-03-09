package com.aurora.strategy.impl;

import co.elastic.clients.elasticsearch.ElasticsearchClient;
import co.elastic.clients.elasticsearch._types.query_dsl.BoolQuery;
import co.elastic.clients.elasticsearch._types.query_dsl.Query;
import co.elastic.clients.elasticsearch.core.SearchRequest;
import co.elastic.clients.elasticsearch.core.SearchResponse;
import co.elastic.clients.elasticsearch.core.search.Highlight;
import co.elastic.clients.elasticsearch.core.search.HighlightField;
import com.aurora.model.dto.ArticleSearchDTO;
import com.aurora.strategy.SearchStrategy;
import com.baomidou.mybatisplus.core.toolkit.CollectionUtils;
import com.baomidou.mybatisplus.core.toolkit.StringUtils;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.List;
import java.util.stream.Collectors;

import static com.aurora.constant.CommonConstant.*;
import static com.aurora.enums.ArticleStatusEnum.PUBLIC;

@Slf4j
@Service("esSearchStrategyImpl")
public class EsSearchStrategyImpl implements SearchStrategy {

    @Autowired
    private ElasticsearchClient elasticsearchClient;

    @Override
    public List<ArticleSearchDTO> searchArticle(String keywords) {
        if (StringUtils.isBlank(keywords)) {
            return new ArrayList<>();
        }
        return search(buildQuery(keywords));
    }

    private SearchRequest buildQuery(String keywords) {
        // 构建布尔查询
        BoolQuery boolQuery = BoolQuery.of(b -> b
            .must(m -> m.bool(bb -> bb
                .should(s -> s.match(t -> t.field("articleTitle").query(keywords).analyzer("ik_max_word")))
                .should(s -> s.match(t -> t.field("articleContent").query(keywords).analyzer("ik_max_word")))
            ))
            .must(m -> m.term(t -> t.field("isDelete").value(FALSE)))
            .must(m -> m.term(t -> t.field("status").value(PUBLIC.getStatus())))
        );

        // 构建高亮配置 - 明确指定 analyzer
        Highlight highlight = Highlight.of(h -> h
            .preTags(PRE_TAG)
            .postTags(POST_TAG)
            .fields("articleTitle", HighlightField.of(hf -> hf
                .fragmentSize(0) // 不分片段，返回完整标题
            ))
            .fields("articleContent", HighlightField.of(hf -> hf
                .fragmentSize(50)
                .numberOfFragments(3) // 最多返回 3 个片段
            ))
        );

        return SearchRequest.of(s -> s
            .index("article")
            .query(Query.of(q -> q.bool(boolQuery)))
            .highlight(highlight)
        );
    }

    private List<ArticleSearchDTO> search(SearchRequest searchRequest) {
        try {
            // 使用 JsonData 作为泛型类型，避免自动反序列化
            SearchResponse<co.elastic.clients.json.JsonData> searchResponse =
                elasticsearchClient.search(searchRequest, co.elastic.clients.json.JsonData.class);

            return searchResponse.hits().hits().stream().map(hit -> {
                // 从原始 JSON 中提取数据
                com.alibaba.fastjson2.JSONObject jsonObj;
                try {
                    co.elastic.clients.json.JsonData sourceData = hit.source();
                    if (sourceData == null) {
                        log.warn("Hit source is null");
                        return null;
                    }

                    // 使用 toJson() 方法获取原始 JSON 字符串
                    String rawJson = sourceData.toJson().toString();
                    log.debug("Raw JSON from ES: {}", rawJson);
                    jsonObj = com.alibaba.fastjson2.JSON.parseObject(rawJson);
                } catch (Exception e) {
                    log.error("Failed to parse ES response: {}", e.getMessage(), e);
                    return null;
                }

                if (jsonObj == null) {
                    return null;
                }

                ArticleSearchDTO article = new ArticleSearchDTO();
                // 注意：ES 中的字段名是 id，不是 articleId
                article.setId(jsonObj.getInteger("id"));
                article.setArticleTitle(jsonObj.getString("articleTitle"));
                article.setArticleContent(jsonObj.getString("articleContent"));
                article.setIsDelete(jsonObj.getInteger("isDelete"));
                article.setStatus(jsonObj.getInteger("status"));

                // 处理高亮字段
                if (hit.highlight() != null) {
                    List<String> titleHighlights = hit.highlight().get("articleTitle");
                    if (CollectionUtils.isNotEmpty(titleHighlights)) {
                        article.setArticleTitle(titleHighlights.get(0));
                    }

                    List<String> contentHighlights = hit.highlight().get("articleContent");
                    if (CollectionUtils.isNotEmpty(contentHighlights)) {
                        // 取最后一个高亮片段
                        article.setArticleContent(contentHighlights.get(contentHighlights.size() - 1));
                    }
                }

                return article;
            }).filter(obj -> obj != null)
              .collect(Collectors.toList());
        } catch (Exception e) {
            log.error("Elasticsearch search error: {}", e.getMessage(), e);
        }
        return new ArrayList<>();
    }

}
