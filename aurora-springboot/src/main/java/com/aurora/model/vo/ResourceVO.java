package com.aurora.model.vo;

import io.swagger.v3.oas.annotations.media.Schema;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import jakarta.validation.constraints.NotBlank;

@Data
@Builder
@AllArgsConstructor
@NoArgsConstructor
@Schema(description = "资源")
public class ResourceVO {

    @Schema(name = "id", description = "资源id", required = true)
    private Integer id;

    @NotBlank(message = "资源名不能为空")
    @Schema(name = "resourceName", description = "资源名", required = true)
    private String resourceName;

    @Schema(name = "url", description = "资源路径", required = true)
    private String url;

    @Schema(name = "url", description = "资源路径", required = true)
    private String requestMethod;

    @Schema(name = "parentId", description = "父资源id", required = true)
    private Integer parentId;

    @Schema(name = "isAnonymous", description = "是否匿名访问", required = true)
    private Integer isAnonymous;

}
