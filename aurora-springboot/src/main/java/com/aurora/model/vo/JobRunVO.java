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
public class JobRunVO {

    @Schema(name = "id", description = "任务 id", required = true)
    private Integer id;

    @Schema(name = "jobGroup", description = "任务组别", required = true)
    private String jobGroup;
}
