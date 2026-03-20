-- ========================================================
-- 修复前端页面路由配置问题
-- ========================================================

-- 修复1: 字典类型路径 (Dict -> sys-dict-type)
UPDATE sys_menu SET 
    path = '/admin/sys-dict-type',
    component = '/admin/sys-dict-type/index'
WHERE menu_name = 'Dict';

-- 修复2: 字典数据路径
UPDATE sys_menu SET 
    path = '/admin/sys-dict-data',
    component = '/admin/sys-dict-data/index'
WHERE menu_name = 'SysDictDataManage';

-- 修复3: 操作日志路径 (修正拼写 oper -> opera)
UPDATE sys_menu SET 
    path = '/admin/sys-opera-log',
    component = '/admin/sys-opera-log/index',
    menu_name = 'SysOperaLogManage'
WHERE menu_name = 'OperLog';

-- 验证修复结果
SELECT '修复后的菜单配置:' AS info;
SELECT menu_id, menu_name, title, path, component 
FROM sys_menu 
WHERE menu_name IN ('Dict', 'SysDictDataManage', 'SysOperaLogManage');
