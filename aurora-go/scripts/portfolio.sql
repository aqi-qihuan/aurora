-- ============================================================
-- Aurora 作品集（Portfolio）模块 — 数据库初始化脚本
-- 适用：aurora-go / aurora-springboot 双后端
-- 用法：mysql -u root -p aurora < scripts/portfolio.sql
-- ============================================================

-- 1. 作品集表（缓存 GitHub 仓库快照）
DROP TABLE IF EXISTS `t_portfolio`;
CREATE TABLE `t_portfolio` (
  `id`                int          NOT NULL AUTO_INCREMENT COMMENT '主键',
  `repo_id`           bigint       NOT NULL COMMENT 'GitHub 仓库 ID（唯一）',
  `name`              varchar(128) NOT NULL COMMENT '仓库名',
  `full_name`         varchar(255) DEFAULT NULL COMMENT 'owner/repo',
  `description`       varchar(500) DEFAULT NULL COMMENT '仓库描述',
  `html_url`          varchar(500) NOT NULL COMMENT '仓库地址',
  `homepage`          varchar(500) DEFAULT NULL COMMENT '演示地址（homepage）',
  `language`          varchar(64)  DEFAULT NULL COMMENT '主语言',
  `stargazers_count`  int          DEFAULT 0  COMMENT 'star 数',
  `forks_count`       int          DEFAULT 0  COMMENT 'fork 数',
  `topics`            text         DEFAULT NULL COMMENT '话题标签，JSON 数组',
  `repo_created_at`   datetime     DEFAULT NULL COMMENT '仓库创建时间',
  `repo_updated_at`   datetime     DEFAULT NULL COMMENT '仓库更新时间',
  `cover`             varchar(500) DEFAULT NULL COMMENT '自定义封面（后台覆盖）',
  `sort`              int          DEFAULT 0  COMMENT '排序权重，越大越靠前',
  `is_featured`       tinyint(1)   DEFAULT 0  COMMENT '是否首页置顶展示 0否 1是',
  `is_visible`        tinyint(1)   DEFAULT 1  COMMENT '是否展示 0隐藏 1展示',
  `create_time`       datetime     DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time`       datetime     DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_repo_id` (`repo_id`),
  KEY `idx_visible_featured` (`is_visible`, `is_featured`, `sort`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='GitHub 作品集';

-- 2. 注册定时任务（每天 03:00 同步 GitHub 仓库）
--    invoke_target 对应 aurora-go 的 TaskRegistry["auroraQuartz.syncGitHubRepos"]
--    misfire_policy=3(立即执行) concurrent=1(禁止并发) status=1(运行中)
INSERT INTO `t_job` (`job_name`, `job_group`, `invoke_target`, `cron_expression`, `misfire_policy`, `concurrent`, `status`, `remark`)
VALUES ('GitHub作品集同步', '默认', 'auroraQuartz.syncGitHubRepos', '0 0 3 * * ?', 3, 1, 1, '每天03:00同步GitHub仓库到t_portfolio');

-- 3. 后台菜单注册 — 作品集管理（放在「系统管理」id=4 下作为子菜单）
--    component 字段对应 aurora-admin-v3 的 src/views/portfolio/Portfolio.vue
--    order_num=6 排在「网站管理」(order_num=1) 之后
INSERT INTO `t_menu` (`name`, `path`, `component`, `icon`, `order_num`, `is_hidden`, `parent_id`)
VALUES ('作品集管理', '/portfolios', 'portfolio/Portfolio.vue', 'el-icon-myimage-fill', 6, 0, 4);

-- 4. 给管理员角色（默认 role_id=1）授权新菜单（若使用 RBAC 菜单角色关联）
-- INSERT INTO `t_role_menu` (`role_id`, `menu_id`) VALUES (1, (SELECT LAST_INSERT_ID()));
