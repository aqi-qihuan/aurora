package com.aurora.quartz;

import cn.hutool.core.date.LocalDateTimeUtil;
import com.alibaba.fastjson2.JSON;
import co.elastic.clients.elasticsearch.ElasticsearchClient;
import co.elastic.clients.elasticsearch.core.IndexRequest;
import co.elastic.clients.elasticsearch.core.DeleteRequest;
import com.aurora.model.dto.ArticleSearchDTO;
import com.aurora.model.dto.UserAreaDTO;
import com.aurora.entity.*;
import com.aurora.mapper.UniqueViewMapper;
import com.aurora.mapper.UserAuthMapper;
import com.aurora.service.*;
import com.aurora.util.BeanCopyUtil;
import com.aurora.util.IpUtil;
import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import lombok.extern.slf4j.Slf4j;
import org.apache.commons.lang3.StringUtils;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.HttpEntity;
import org.springframework.http.HttpHeaders;
import org.springframework.stereotype.Component;
import org.springframework.web.client.RestTemplate;

import java.time.LocalDateTime;
import java.time.temporal.ChronoUnit;
import java.util.*;
import java.util.stream.Collectors;

import static com.aurora.constant.CommonConstant.UNKNOWN;
import static com.aurora.constant.RedisConstant.*;

@Slf4j
@Component("auroraQuartz")
public class AuroraQuartz {

    @Autowired
    private RedisService redisService;

    @Autowired
    private ArticleService articleService;

    @Autowired
    private JobLogService jobLogService;

    @Autowired
    private ResourceService resourceService;

    @Autowired
    private RoleResourceService roleResourceService;

    @Autowired
    private UniqueViewMapper uniqueViewMapper;

    @Autowired
    private UserAuthMapper userAuthMapper;

    @Autowired
    private RestTemplate restTemplate;

    @Autowired(required = false)
    private ElasticsearchClient elasticsearchClient;


    @Value("${website.url}")
    private String websiteUrl;

    public void saveUniqueView() {
        try {
            Long count = redisService.sSize(UNIQUE_VISITOR);
            UniqueView uniqueView = UniqueView.builder()
                    .createTime(LocalDateTimeUtil.offset(LocalDateTime.now(), -1, ChronoUnit.DAYS))
                    .viewsCount(Optional.of(count.intValue()).orElse(0))
                    .build();
            uniqueViewMapper.insert(uniqueView);
            log.info("Unique view statistics saved: {}", count);
        } catch (Exception e) {
            log.error("Failed to save unique view statistics: {}", e.getMessage(), e);
        }
    }

    public void clear() {
        try {
            redisService.del(UNIQUE_VISITOR);
            redisService.del(VISITOR_AREA);
            log.info("Redis cache cleared successfully");
        } catch (Exception e) {
            log.error("Failed to clear Redis cache: {}", e.getMessage(), e);
        }
    }

    public void statisticalUserArea() {
        try {
            Map<String, Long> userAreaMap = userAuthMapper.selectList(new LambdaQueryWrapper<UserAuth>().select(UserAuth::getIpSource))
                    .stream()
                    .map(item -> {
                        if (Objects.nonNull(item) && StringUtils.isNotBlank(item.getIpSource())) {
                            return IpUtil.getIpProvince(item.getIpSource());
                        }
                        return UNKNOWN;
                    })
                    .collect(Collectors.groupingBy(item -> item, Collectors.counting()));
            List<UserAreaDTO> userAreaList = userAreaMap.entrySet().stream()
                    .map(item -> UserAreaDTO.builder()
                            .name(item.getKey())
                            .value(item.getValue())
                            .build())
                    .collect(Collectors.toList());
            redisService.set(USER_AREA, JSON.toJSONString(userAreaList));
            log.info("User area statistics saved successfully");
        } catch (Exception e) {
            log.error("Failed to save user area statistics: {}", e.getMessage(), e);
        }
    }

    public void baiduSeo() {
        List<Integer> ids = articleService.list().stream().map(Article::getId).collect(Collectors.toList());
        if (ids.isEmpty()) {
            log.info("No articles to submit for SEO");
            return;
        }
        
        // 构建所有 URL（每行一个）
        StringBuilder urlsBuilder = new StringBuilder();
        for (Integer id : ids) {
            String url = websiteUrl + "/articles/" + id;
            urlsBuilder.append(url).append("\n");
        }
        String urls = urlsBuilder.toString();
        
        // 设置正确的请求头（不硬编码 Content-Length）
        HttpHeaders headers = new HttpHeaders();
        headers.add("Host", "data.zz.baidu.com");
        headers.add("User-Agent", "curl/7.12.1");
        headers.add("Content-Type", "text/plain");
        
        try {
            HttpEntity<String> entity = new HttpEntity<>(urls, headers);
            // 使用百度推送 API 正确地址（需要替换为真实的 token）
            String result = restTemplate.postForObject(
                "http://data.zz.baidu.com/urls?site=" + websiteUrl + "&token=你的百度推送 token", 
                entity, 
                String.class
            );
            log.info("Baidu SEO submission result: {}", result);
        } catch (Exception e) {
            log.error("Failed to submit URLs to Baidu SEO", e);
        }
    }

    public void clearJobLogs() {
        jobLogService.cleanJobLogs();
    }

    public void importSwagger() {
        resourceService.importSwagger();
        List<Integer> resourceIds = resourceService.list().stream().map(Resource::getId).collect(Collectors.toList());
        List<RoleResource> roleResources = new ArrayList<>();
        for (Integer resourceId : resourceIds) {
            roleResources.add(RoleResource.builder()
                    .roleId(1)
                    .resourceId(resourceId)
                    .build());
        }
        roleResourceService.saveBatch(roleResources);
    }

    public void importDataIntoES() {
        if (elasticsearchClient == null) {
            log.warn("ElasticsearchClient not available, skipping ES import");
            return;
        }
        try {
            // 删除所有文档 - 通过删除索引再重建的方式
            elasticsearchClient.indices().delete(d -> d.index("article"));
        } catch (Exception e) {
            // 索引不存在时忽略错误
            log.debug("Index 'article' does not exist, will create new one");
        }
        
        List<Article> articles = articleService.list();
        for (Article article : articles) {
            ArticleSearchDTO dto = BeanCopyUtil.copyObject(article, ArticleSearchDTO.class);
            try {
                elasticsearchClient.index(i -> i
                    .index("article")
                    .id(dto.getId().toString())
                    .document(dto)
                );
            } catch (Exception e) {
                log.error("Error indexing article {}", dto.getId(), e);
            }
        }
    }
}
