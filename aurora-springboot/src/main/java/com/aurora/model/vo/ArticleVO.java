package com.aurora.model.vo;


import io.swagger.v3.oas.annotations.media.Schema;
import lombok.*;

import jakarta.validation.constraints.NotBlank;
import java.util.List;

@Data
@Builder
@AllArgsConstructor
@NoArgsConstructor
@Schema(description = "文章")
public class ArticleVO {

    @Schema(name = "id", description = "文章id")
    private Integer id;

    @NotBlank(message = "文章标题不能为空")
    @Schema(name = "articleTitle", description = "文章标题", required = true)
    private String articleTitle;

    @NotBlank(message = "文章内容不能为空")
    @Schema(name = "articleContent", description = "文章内容", required = true)
    private String articleContent;

    @Schema(name = "articleAbstract", description = "文章摘要")
    private String articleAbstract;

    @Schema(name = "articleCover", description = "文章缩略图")
    private String articleCover;

    @Schema(name = "category", description = "文章分类")
    private String categoryName;

    @Schema(name = "tagNameList", description = "文章标签")
    private List<String> tagNames;

    @Schema(name = "isTop", description = "是否置顶")
    private Integer isTop;

    @Schema(name = "isFeatured", description = "是否推荐")
    private Integer isFeatured;

    @Schema(name = "status", description = "文章状态")
    private Integer status;

    @Schema(name = "type", description = "文章类型")
    private Integer type;

    @Schema(name = "originalUrl", description = "原文链接")
    private String originalUrl;

    @Schema(name = "password", description = "文章访问密码")
    private String password;
}
