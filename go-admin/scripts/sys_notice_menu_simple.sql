-- ========================================================  
-- 公告管理菜单配置（简化版）
-- ========================================================

-- 步骤：
-- 1. 先查询系统管理的 menu_id：
--    SELECT menu_id, title FROM sys_menu WHERE menu_name = 'Admin';
-- 2. 把下面的 @parent_id 替换为实际的 menu_id（通常是 51 或 52）
-- 3. 执行此 SQL

-- 假设系统管理的 menu_id 是 51，如果不对请修改
SET @parent_id = 51;

-- 插入公告管理菜单
INSERT INTO `sys_menu` (
    `menu_name`, `title`, `icon`, `path`, `paths`, `menu_type`, 
    `action`, `permission`, `parent_id`, `no_cache`, 
    `breadcrumb`, `component`, `sort`, `visible`, `is_frame`,
    `create_by`, `update_by`, `created_at`, `updated_at`
) VALUES (
    'SysNotice', '公告管理', 'notification', '/sys-notice', 
    '/0/1/3/51/9999', 'C', 
    '', 'admin:sysnotice:list', @parent_id, '0',
    '', 'admin/sys-notice/index', 10, '0', '0',
    1, 1, NOW(), NOW()
);

-- 获取新菜单ID
SET @notice_menu_id = LAST_INSERT_ID();

-- 插入按钮权限
INSERT INTO `sys_menu` (`menu_name`, `title`, `icon`, `path`, `paths`, `menu_type`, `action`, `permission`, `parent_id`, `sort`, `visible`, `create_by`, `update_by`, `created_at`, `updated_at`) VALUES
('SysNoticeView', '公告查看', '', '', '/0/1/3/51/9999/1', 'F', '', 'admin:sysnotice:view', @notice_menu_id, 0, '0', 1, 1, NOW(), NOW()),
('SysNoticeAdd', '公告新增', '', '', '/0/1/3/51/9999/2', 'F', '', 'admin:sysnotice:add', @notice_menu_id, 1, '0', 1, 1, NOW(), NOW()),
('SysNoticeEdit', '公告编辑', '', '', '/0/1/3/51/9999/3', 'F', '', 'admin:sysnotice:edit', @notice_menu_id, 2, '0', 1, 1, NOW(), NOW()),
('SysNoticeRemove', '公告删除', '', '', '/0/1/3/51/9999/4', 'F', '', 'admin:sysnotice:remove', @notice_menu_id, 3, '0', 1, 1, NOW(), NOW());

-- 给管理员角色授权
INSERT INTO `sys_role_menu` (`role_id`, `menu_id`, `created_at`)
SELECT 1, menu_id, NOW() 
FROM sys_menu 
WHERE menu_name LIKE 'SysNotice%' 
AND menu_id NOT IN (SELECT menu_id FROM sys_role_menu WHERE role_id = 1);

-- 验证
SELECT '公告管理菜单已创建！' AS result;
SELECT menu_id, menu_name, title, path, component FROM sys_menu WHERE menu_name LIKE 'SysNotice%';
