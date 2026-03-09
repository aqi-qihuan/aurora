package com.aurora.model.vo;

import io.swagger.v3.oas.annotations.media.Schema;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class JobSearchVO {

    @Schema(name = "任务名称", description = "jobName", required = true)
    private String jobName;

    @Schema(name = "任务组别", description = "jobGroup", required = true)
    private String jobGroup;

    @Schema(name = "任务状态", description = "status", required = true)
    private Integer status;
}
