SET @parent_id = (SELECT menu_id FROM sys_menu WHERE title = '系统管理' LIMIT 1);
SET @parent_id = IFNULL(@parent_id, 1);

INSERT INTO sys_menu (menu_name, title, icon, path, paths, menu_type, action, permission, parent_id, no_cache, breadcrumb, component, sort, visible, is_frame, create_by, update_by, created_at, updated_at) 
VALUES ('SysNotice', '公告管理', 'notification', '/sys-notice', CONCAT('/0/1/3/', @parent_id), 'C', '', 'admin:sysnotice:list', @parent_id, '0', '', 'admin/sys-notice/index', 10, '0', '0', 1, 1, NOW(), NOW());

SET @notice_menu_id = LAST_INSERT_ID();

INSERT INTO sys_menu (menu_name, title, icon, path, paths, menu_type, action, permission, parent_id, sort, visible, create_by, update_by, created_at, updated_at) VALUES
('SysNoticeView', '公告查看', '', '', CONCAT('/0/1/3/', @parent_id, '/', @notice_menu_id), 'F', '', 'admin:sysnotice:view', @notice_menu_id, 0, '0', 1, 1, NOW(), NOW()),
('SysNoticeAdd', '公告新增', '', '', CONCAT('/0/1/3/', @parent_id, '/', @notice_menu_id), 'F', '', 'admin:sysnotice:add', @notice_menu_id, 1, '0', 1, 1, NOW(), NOW()),
('SysNoticeEdit', '公告编辑', '', '', CONCAT('/0/1/3/', @parent_id, '/', @notice_menu_id), 'F', '', 'admin:sysnotice:edit', @notice_menu_id, 2, '0', 1, 1, NOW(), NOW()),
('SysNoticeRemove', '公告删除', '', '', CONCAT('/0/1/3/', @parent_id, '/', @notice_menu_id), 'F', '', 'admin:sysnotice:remove', @notice_menu_id, 3, '0', 1, 1, NOW(), NOW());

INSERT INTO sys_role_menu (role_id, menu_id, created_at)
SELECT 1, menu_id, NOW() FROM sys_menu WHERE menu_name LIKE 'SysNotice%' AND menu_id NOT IN (SELECT menu_id FROM sys_role_menu WHERE role_id = 1);

SELECT '菜单创建成功' as result, menu_id, menu_name, title FROM sys_menu WHERE menu_name LIKE 'SysNotice%';
