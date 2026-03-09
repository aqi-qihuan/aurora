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
@Schema(description = "标签对象")
public class TagVO {

    @Schema(name = "id", description = "标签id")
    private Integer id;

    @NotBlank(message = "标签名不能为空")
    @Schema(name = "categoryName", description = "标签名", required = true)
    private String tagName;

}
