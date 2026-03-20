-- ========================================================
-- 清理并重建公告管理菜单（修复版）
-- ========================================================

-- 步骤1：先找出需要删除的菜单ID，存入临时表
DROP TEMPORARY TABLE IF EXISTS tmp_notice_menu_ids;
CREATE TEMPORARY TABLE tmp_notice_menu_ids AS
SELECT menu_id FROM sys_menu WHERE menu_name LIKE 'SysNotice%' OR title = '公告管理';

-- 步骤2：删除角色权限关联
DELETE FROM sys_role_menu WHERE menu_id IN (SELECT menu_id FROM tmp_notice_menu_ids);

-- 步骤3：删除按钮权限（子菜单）
DELETE FROM sys_menu WHERE parent_id IN (SELECT menu_id FROM tmp_notice_menu_ids);

-- 步骤4：删除主菜单
DELETE FROM sys_menu WHERE menu_id IN (SELECT menu_id FROM tmp_notice_menu_ids);

-- 步骤5：清理临时表
DROP TEMPORARY TABLE IF EXISTS tmp_notice_menu_ids;

-- 步骤6：重新创建（确保 parent_id = 2 是系统管理）
INSERT INTO sys_menu (menu_name, title, icon, path, paths, menu_type, action, permission, parent_id, no_cache, breadcrumb, component, sort, visible, is_frame) 
VALUES ('SysNotice', '公告管理', 'notification', '/sys-notice', '/0/1/3/2', 'C', '', 'admin:sysnotice:list', 2, '0', '', 'views/admin/sys-notice/index', 10, '0', '0');

SET @notice_menu_id = LAST_INSERT_ID();

-- 步骤7：创建按钮权限
INSERT INTO sys_menu (menu_name, title, icon, path, paths, menu_type, action, permission, parent_id, sort, visible) VALUES
('SysNoticeView', '公告查看', '', '', '/0/1/3/2', 'F', '', 'admin:sysnotice:view', @notice_menu_id, 0, '0'),
('SysNoticeAdd', '公告新增', '', '', '/0/1/3/2', 'F', '', 'admin:sysnotice:add', @notice_menu_id, 1, '0'),
('SysNoticeEdit', '公告编辑', '', '', '/0/1/3/2', 'F', '', 'admin:sysnotice:edit', @notice_menu_id, 2, '0'),
('SysNoticeRemove', '公告删除', '', '', '/0/1/3/2', 'F', '', 'admin:sysnotice:remove', @notice_menu_id, 3, '0');

-- 步骤8：给管理员角色授权
INSERT INTO sys_role_menu (role_id, menu_id) VALUES (1, @notice_menu_id);
INSERT INTO sys_role_menu (role_id, menu_id) SELECT 1, menu_id FROM sys_menu WHERE parent_id = @notice_menu_id;

-- 验证结果
SELECT '修复完成！当前公告管理菜单：' AS info;
SELECT menu_id, menu_name, title, path, component, parent_id, menu_type
FROM sys_menu 
WHERE menu_name LIKE 'SysNotice%' OR title = '公告管理';
