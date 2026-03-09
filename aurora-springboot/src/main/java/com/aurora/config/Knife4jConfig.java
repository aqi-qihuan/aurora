package com.aurora.config;

import io.swagger.v3.oas.models.OpenAPI;
import io.swagger.v3.oas.models.info.Contact;
import io.swagger.v3.oas.models.info.Info;
import io.swagger.v3.oas.models.security.SecurityRequirement;
import io.swagger.v3.oas.models.servers.Server;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

import java.util.Collections;
import java.util.List;

@Configuration
public class Knife4jConfig {

    @Bean
    public OpenAPI openAPI() {
        return new OpenAPI()
                .info(new Info()
                        .title("Aurora API 文档")
                        .description("Aurora 博客系统后端 API")
                        .contact(new Contact()
                                .name("七七")
                                .email("2316364297@qq.com"))
                        .version("1.0"))
                .servers(List.of(new Server()
                        .url("https://www.aqi125.cn")
                        .description("Production Server")))
                .security(Collections.singletonList(new SecurityRequirement().addList("Authorization")));
    }

}
