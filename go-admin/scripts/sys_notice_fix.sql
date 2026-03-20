INSERT INTO sys_menu (menu_name, title, icon, path, paths, menu_type, action, permission, parent_id, no_cache, breadcrumb, component, sort, visible, is_frame) 
VALUES ('SysNotice', '公告管理', 'notification', '/sys-notice', '/0/1/3/2', 'C', '', 'admin:sysnotice:list', 2, '0', '', 'admin/sys-notice/index', 10, '0', '0');

SET @notice_menu_id = LAST_INSERT_ID();

INSERT INTO sys_menu (menu_name, title, icon, path, paths, menu_type, action, permission, parent_id, sort, visible) VALUES
('SysNoticeView', '公告查看', '', '', '/0/1/3/2', 'F', '', 'admin:sysnotice:view', @notice_menu_id, 0, '0'),
('SysNoticeAdd', '公告新增', '', '', '/0/1/3/2', 'F', '', 'admin:sysnotice:add', @notice_menu_id, 1, '0'),
('SysNoticeEdit', '公告编辑', '', '', '/0/1/3/2', 'F', '', 'admin:sysnotice:edit', @notice_menu_id, 2, '0'),
('SysNoticeRemove', '公告删除', '', '', '/0/1/3/2', 'F', '', 'admin:sysnotice:remove', @notice_menu_id, 3, '0');

INSERT INTO sys_role_menu (role_id, menu_id) VALUES (1, @notice_menu_id);
INSERT INTO sys_role_menu (role_id, menu_id) SELECT 1, menu_id FROM sys_menu WHERE parent_id = @notice_menu_id;

SELECT '公告管理菜单创建成功！' as result;
SELECT menu_id, menu_name, title, path, component FROM sys_menu WHERE menu_name LIKE 'SysNotice%';
