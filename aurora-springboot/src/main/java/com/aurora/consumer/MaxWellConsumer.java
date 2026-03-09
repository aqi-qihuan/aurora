package com.aurora.consumer;

import co.elastic.clients.elasticsearch.ElasticsearchClient;
import com.alibaba.fastjson2.JSON;
import com.aurora.model.dto.ArticleSearchDTO;
import com.aurora.model.dto.MaxwellDataDTO;
import com.aurora.entity.Article;
import com.aurora.util.BeanCopyUtil;
import lombok.extern.slf4j.Slf4j;
import org.springframework.amqp.rabbit.annotation.RabbitHandler;
import org.springframework.amqp.rabbit.annotation.RabbitListener;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import static com.aurora.constant.RabbitMQConstant.MAXWELL_QUEUE;

@Slf4j
@Component
@RabbitListener(queues = MAXWELL_QUEUE)
public class MaxWellConsumer {

    @Autowired(required = false)
    private ElasticsearchClient elasticsearchClient;

    @RabbitHandler
    public void process(byte[] data) {
        if (elasticsearchClient == null) {
            log.warn("ElasticsearchClient not available, skipping ES save");
            return;
        }
        try {
            MaxwellDataDTO maxwellDataDTO = JSON.parseObject(new String(data), MaxwellDataDTO.class);
            Article article = JSON.parseObject(JSON.toJSONString(maxwellDataDTO.getData()), Article.class);
            switch (maxwellDataDTO.getType()) {
                case "insert":
                case "update":
                    ArticleSearchDTO dto = BeanCopyUtil.copyObject(article, ArticleSearchDTO.class);
                    elasticsearchClient.index(i -> i
                        .index("article")
                        .id(dto.getId().toString())
                        .document(dto)
                    );
                    break;
                case "delete":
                    elasticsearchClient.delete(d -> d
                        .index("article")
                        .id(article.getId().toString())
                    );
                    break;
                default:
                    break;
            }
        } catch (Exception e) {
            log.error("Error processing ES operation", e);
        }
    }
}