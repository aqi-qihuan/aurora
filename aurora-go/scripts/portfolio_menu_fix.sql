-- ============================================================
-- 作品集菜单权限修复脚本
-- 作用：让 admin 角色能看到「作品集管理」菜单
-- 执行方式：在 MySQL 中直接运行本脚本
-- ============================================================

-- 1. 确认 t_menu 中已有作品集菜单（id=226），没有则插入
INSERT IGNORE INTO `t_menu` (`id`, `name`, `path`, `component`, `icon`, `create_time`, `update_time`, `order_num`, `parent_id`, `is_hidden`)
VALUES (226, '作品集管理', '/portfolios', 'portfolio/Portfolio.vue', 'el-icon-myimage-fill', NOW(), NULL, 6, 4, 0);

-- 2. 给 admin 角色(role_id=1) 分配 menu_id=226 权限
--    t_role_menu 没有唯一约束，先删再插避免重复
DELETE FROM `t_role_menu` WHERE `role_id` = 1 AND `menu_id` = 226;
INSERT INTO `t_role_menu` (`role_id`, `menu_id`) VALUES (1, 226);

-- 3. 验证
SELECT m.id, m.name, m.path, m.component, m.parent_id, m.order_num, m.is_hidden,
       rm.role_id AS granted_role
FROM `t_menu` m
LEFT JOIN `t_role_menu` rm ON rm.menu_id = m.id AND rm.role_id = 1
WHERE m.id = 226;
