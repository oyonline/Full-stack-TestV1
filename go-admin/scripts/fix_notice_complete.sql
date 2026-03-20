-- ========================================================
-- 步骤1: 诊断 - 查看当前混乱情况
-- ========================================================
SELECT '=== 步骤1: 诊断当前 SysNotice 菜单情况 ===' AS step;
SELECT menu_id, menu_name, title, path, component, menu_type, parent_id 
FROM sys_menu 
WHERE menu_name LIKE 'SysNotice%' OR title = '公告管理'
ORDER BY menu_id;

-- ========================================================
-- 步骤2: 清理 - 使用临时表删除所有 SysNotice 相关数据
-- ========================================================
SELECT '=== 步骤2: 清理混乱数据 ===' AS step;

-- 创建临时表存储要删除的菜单ID
DROP TEMPORARY TABLE IF EXISTS tmp_notice_ids;
CREATE TEMPORARY TABLE tmp_notice_ids AS
SELECT menu_id FROM sys_menu WHERE menu_name LIKE 'SysNotice%' OR title = '公告管理';

-- 查看将要删除的ID
SELECT '将要删除的菜单ID:' AS info, GROUP_CONCAT(menu_id) AS ids FROM tmp_notice_ids;

-- 删除角色权限关联
DELETE FROM sys_role_menu WHERE menu_id IN (SELECT menu_id FROM tmp_notice_ids);
SELECT CONCAT('已删除 sys_role_menu 关联: ', ROW_COUNT(), ' 条') AS result;

-- 删除按钮权限（子菜单）
DELETE FROM sys_menu WHERE parent_id IN (SELECT menu_id FROM tmp_notice_ids);
SELECT CONCAT('已删除子菜单: ', ROW_COUNT(), ' 条') AS result;

-- 删除主菜单
DELETE FROM sys_menu WHERE menu_id IN (SELECT menu_id FROM tmp_notice_ids);
SELECT CONCAT('已删除主菜单: ', ROW_COUNT(), ' 条') AS result;

-- 清理临时表
DROP TEMPORARY TABLE IF EXISTS tmp_notice_ids;

-- ========================================================
-- 步骤3: 重建 - 使用正确格式参考 SysConfig
-- ========================================================
SELECT '=== 步骤3: 重建正确结构 ===' AS step;

-- 获取系统管理的 parent_id (通常是 2)
SET @parent_id = (SELECT menu_id FROM sys_menu WHERE menu_name = 'Admin' OR title = '系统管理' LIMIT 1);
SET @parent_id = IFNULL(@parent_id, 2);
SELECT CONCAT('使用 parent_id: ', @parent_id) AS info;

-- 参考 SysConfig 格式插入公告管理菜单
-- 关键修正: component = '/admin/sys-notice/index' (带前导/，不带views前缀)
INSERT INTO sys_menu (
    menu_name, title, icon, path, paths, menu_type, 
    action, permission, parent_id, no_cache, 
    breadcrumb, component, sort, visible, is_frame
) VALUES (
    'SysNotice', '公告管理', 'notification', '/sys-notice', 
    CONCAT('/0/1/3/', @parent_id), 'C', 
    '', 'admin:sysnotice:list', @parent_id, '0',
    '', '/admin/sys-notice/index', 10, '0', '0'
);

SET @notice_menu_id = LAST_INSERT_ID();
SELECT CONCAT('创建主菜单 ID: ', @notice_menu_id) AS result;

-- 创建4个按钮权限子项
INSERT INTO sys_menu (menu_name, title, icon, path, paths, menu_type, action, permission, parent_id, sort, visible) VALUES
('SysNoticeView', '公告查看', '', '', CONCAT('/0/1/3/', @parent_id, '/', @notice_menu_id), 'F', '', 'admin:sysnotice:view', @notice_menu_id, 0, '0'),
('SysNoticeAdd', '公告新增', '', '', CONCAT('/0/1/3/', @parent_id, '/', @notice_menu_id), 'F', '', 'admin:sysnotice:add', @notice_menu_id, 1, '0'),
('SysNoticeEdit', '公告编辑', '', '', CONCAT('/0/1/3/', @parent_id, '/', @notice_menu_id), 'F', '', 'admin:sysnotice:edit', @notice_menu_id, 2, '0'),
('SysNoticeRemove', '公告删除', '', '', CONCAT('/0/1/3/', @parent_id, '/', @notice_menu_id), 'F', '', 'admin:sysnotice:remove', @notice_menu_id, 3, '0');

SELECT CONCAT('创建按钮权限: ', ROW_COUNT(), ' 个') AS result;

-- ========================================================
-- 步骤4: 验证 - 对比 SysConfig 和 SysNotice 的字段差异
-- ========================================================
SELECT '=== 步骤4: 验证修复结果 ===' AS step;

-- 对比检查
SELECT '对比 SysConfig vs SysNotice:' AS check_item;
SELECT 
    'SysConfig' as menu,
    menu_name, title, path, component, menu_type, parent_id
FROM sys_menu 
WHERE menu_name = 'SysConfigManage'
UNION ALL
SELECT 
    'SysNotice' as menu,
    menu_name, title, path, component, menu_type, parent_id
FROM sys_menu 
WHERE menu_name = 'SysNotice';

-- 确认最终结构
SELECT '最终 SysNotice 菜单结构:' AS info;
SELECT menu_id, menu_name, title, path, component, menu_type, parent_id 
FROM sys_menu 
WHERE menu_name LIKE 'SysNotice%'
ORDER BY menu_id;

-- 统计确认
SELECT '统计:' AS info, 
    COUNT(*) as total_menus,
    SUM(CASE WHEN menu_type = 'C' THEN 1 ELSE 0 END) as dir_count,
    SUM(CASE WHEN menu_type = 'F' THEN 1 ELSE 0 END) as btn_count
FROM sys_menu 
WHERE menu_name LIKE 'SysNotice%';

-- ========================================================
-- 步骤5: 授权 - 给管理员角色授权
-- ========================================================
SELECT '=== 步骤5: 给管理员角色授权 ===' AS step;

INSERT INTO sys_role_menu (role_id, menu_id)
SELECT 1, menu_id FROM sys_menu WHERE menu_name = 'SysNotice'
ON DUPLICATE KEY UPDATE menu_id = menu_id;

INSERT INTO sys_role_menu (role_id, menu_id)
SELECT 1, menu_id FROM sys_menu WHERE parent_id = @notice_menu_id
ON DUPLICATE KEY UPDATE menu_id = menu_id;

SELECT CONCAT('授权完成: ', ROW_COUNT(), ' 条') AS result;

-- 最终验证授权
SELECT '管理员角色的 SysNotice 权限:' AS info;
SELECT m.menu_id, m.menu_name, m.title, m.permission
FROM sys_menu m
JOIN sys_role_menu rm ON m.menu_id = rm.menu_id
WHERE rm.role_id = 1 AND m.menu_name LIKE 'SysNotice%'
ORDER BY m.menu_id;

SELECT '=== 修复完成 ===' AS step;
