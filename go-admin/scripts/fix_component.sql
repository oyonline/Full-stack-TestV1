-- 修正公告管理的组件路径
UPDATE sys_menu 
SET component = 'views/admin/sys-notice/index'
WHERE menu_name = 'SysNotice';

-- 验证
SELECT menu_id, menu_name, title, component FROM sys_menu WHERE menu_name = 'SysNotice';
