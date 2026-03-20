-- ========================================================  
-- 公告管理菜单配置
-- 执行后会添加菜单到系统管理下
-- ========================================================

-- 1. 先查找系统管理的 menu_id 作为父菜单
-- 通常系统管理的路径是 /admin 或类似
SET @parent_id = (SELECT menu_id FROM sys_menu WHERE menu_name = 'Admin' OR path = '/admin' LIMIT 1);

-- 如果没找到，使用默认值 1（通常是系统管理）
SET @parent_id = IFNULL(@parent_id, 1);

-- 2. 添加公告管理菜单（目录类型 C）
INSERT INTO `sys_menu` (
    `menu_name`, `title`, `icon`, `path`, `paths`, `menu_type`, 
    `action`, `permission`, `parent_id`, `no_cache`, 
    `breadcrumb`, `component`, `sort`, `visible`, `is_frame`,
    `create_by`, `update_by`, `created_at`, `updated_at`
) VALUES (
    'SysNotice', '公告管理', 'notification', '/sys-notice', 
    '/0/1/3/51', 'C', 
    '', 'admin:sysnotice:list', @parent_id, '0',
    '', 'admin/sys-notice/index', 10, '0', '0',
    1, 1, NOW(), NOW()
);

-- 3. 获取刚插入的菜单ID
SET @notice_menu_id = LAST_INSERT_ID();

-- 4. 添加按钮权限菜单（功能类型 F）
INSERT INTO `sys_menu` (`menu_name`, `title`, `icon`, `path`, `paths`, `menu_type`, `action`, `permission`, `parent_id`, `sort`, `visible`, `create_by`, `update_by`, `created_at`, `updated_at`) VALUES
('SysNoticeView', '公告查看', '', '', CONCAT('/0/1/3/', @parent_id, '/', @notice_menu_id), 'F', '', 'admin:sysnotice:view', @notice_menu_id, 0, '0', 1, 1, NOW(), NOW()),
('SysNoticeAdd', '公告新增', '', '', CONCAT('/0/1/3/', @parent_id, '/', @notice_menu_id), 'F', '', 'admin:sysnotice:add', @notice_menu_id, 1, '0', 1, 1, NOW(), NOW()),
('SysNoticeEdit', '公告编辑', '', '', CONCAT('/0/1/3/', @parent_id, '/', @notice_menu_id), 'F', '', 'admin:sysnotice:edit', @notice_menu_id, 2, '0', 1, 1, NOW(), NOW()),
('SysNoticeRemove', '公告删除', '', '', CONCAT('/0/1/3/', @parent_id, '/', @notice_menu_id), 'F', '', 'admin:sysnotice:remove', @notice_menu_id, 3, '0', 1, 1, NOW(), NOW());

-- 5. 给管理员角色分配权限（role_id = 1 通常是管理员）
INSERT INTO `sys_role_menu` (`role_id`, `menu_id`, `created_at`)
SELECT 1, menu_id, NOW() 
FROM sys_menu 
WHERE menu_name LIKE 'SysNotice%' 
AND menu_id NOT IN (SELECT menu_id FROM sys_role_menu WHERE role_id = 1);

-- 6. 验证插入结果
SELECT '公告管理菜单已创建：' AS info;
SELECT m.menu_id, m.menu_name, m.title, m.path, m.component, m.permission, m.menu_type
FROM sys_menu m 
WHERE m.menu_name LIKE 'SysNotice%'
ORDER BY m.menu_id;
