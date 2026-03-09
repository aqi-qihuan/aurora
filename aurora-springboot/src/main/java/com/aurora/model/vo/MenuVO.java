package com.aurora.model.vo;

import io.swagger.v3.oas.annotations.media.Schema;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;

@Data
@AllArgsConstructor
@NoArgsConstructor
@Builder
@Schema(description = "菜单")
public class MenuVO {

    @Schema(name = "id", description = "菜单id")
    private Integer id;

    @NotBlank(message = "菜单名不能为空")
    @Schema(name = "name", description = "菜单名")
    private String name;

    @NotBlank(message = "菜单icon不能为空")
    @Schema(name = "icon", description = "菜单icon")
    private String icon;

    @NotBlank(message = "路径不能为空")
    @Schema(name = "path", description = "路径")
    private String path;

    @NotBlank(message = "组件不能为空")
    @Schema(name = "component", description = "组件")
    private String component;

    @NotNull(message = "排序不能为空")
    @Schema(name = "orderNum", description = "排序")
    private Integer orderNum;

    @Schema(name = "parentId", description = "父id")
    private Integer parentId;

    @Schema(name = "isHidden", description = "是否隐藏")
    private Integer isHidden;

}
