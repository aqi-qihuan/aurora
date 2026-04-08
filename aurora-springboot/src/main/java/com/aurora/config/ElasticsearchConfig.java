package com.aurora.config;

import co.elastic.clients.elasticsearch.ElasticsearchClient;
import co.elastic.clients.json.jackson.JacksonJsonpMapper;
import co.elastic.clients.transport.ElasticsearchTransport;
import co.elastic.clients.transport.rest_client.RestClientTransport;
import lombok.Data;
import org.apache.http.HttpHost;
import org.apache.http.auth.AuthScope;
import org.apache.http.auth.UsernamePasswordCredentials;
import org.apache.http.client.CredentialsProvider;
import org.apache.http.client.config.RequestConfig;
import org.apache.http.impl.client.BasicCredentialsProvider;
import org.apache.http.impl.nio.client.HttpAsyncClientBuilder;
import org.elasticsearch.client.RestClient;
import org.elasticsearch.client.RestClientBuilder;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

import java.util.concurrent.TimeUnit;

@Data
@Configuration
@ConfigurationProperties(prefix = "elasticsearch.rest")
public class ElasticsearchConfig {

    private String uris;
    private String username;
    private String password;

    @Bean
    public RestClient elasticsearchRestClient() {
        String[] uriParts = uris.replace("http://", "").split(":");
        String hostname = uriParts[0];
        int port = Integer.parseInt(uriParts[1]);

        final CredentialsProvider credentialsProvider = new BasicCredentialsProvider();
        credentialsProvider.setCredentials(AuthScope.ANY,
                new UsernamePasswordCredentials(username, password));

        // 配置超时（ES 8.17.x + SB 4.0 需要显式设置）
        // socket timeout 设置为 60 秒以支持批量同步操作
        RequestConfig requestConfig = RequestConfig.custom()
                .setConnectTimeout(10000)
                .setSocketTimeout(60000)
                .setConnectionRequestTimeout(10000)
                .build();

        RestClientBuilder builder = RestClient.builder(new HttpHost(hostname, port, "http"))
                .setHttpClientConfigCallback((HttpAsyncClientBuilder httpClientBuilder) -> {
                    httpClientBuilder.setDefaultCredentialsProvider(credentialsProvider)
                            .setDefaultRequestConfig(requestConfig)
                            .setMaxConnTotal(100)
                            .setMaxConnPerRoute(30);
                    return httpClientBuilder;
                });

        return builder.build();
    }

    @Bean
    public ElasticsearchClient elasticsearchClient(RestClient restClient) {
        JacksonJsonpMapper mapper = new JacksonJsonpMapper();
        return new ElasticsearchClient(new RestClientTransport(restClient, mapper));
    }
}
