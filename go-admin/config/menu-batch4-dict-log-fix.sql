-- 字典管理 / 日志管理 空页面修复
-- 目标：
-- 1. 字典管理改为父级分组菜单，并补齐「字典类型」子菜单
-- 2. 字典数据改挂到字典管理分组下，并对齐到当前前端页面组件
-- 3. 日志管理父级改为布局分组，操作日志统一收口到 sys-opera-log
-- 4. 脚本可重复执行，适用于已初始化过的 MySQL 库

START TRANSACTION;

UPDATE sys_menu
SET
  menu_name = 'Dict',
  title = '字典管理',
  icon = 'education',
  path = '/admin/dict',
  paths = '/0/2/58',
  menu_type = 'M',
  action = '无',
  permission = '',
  parent_id = 2,
  no_cache = 0,
  breadcrumb = '',
  component = 'RouteView',
  sort = 60,
  visible = '0',
  is_frame = '1',
  updated_at = NOW(3)
WHERE menu_id = 58;

INSERT INTO sys_menu (
  menu_id, menu_name, title, icon, path, paths, menu_type, action,
  permission, parent_id, no_cache, breadcrumb, component, sort, visible,
  is_frame, create_by, update_by, created_at, updated_at, deleted_at
)
SELECT
  543, 'SysDictTypeManage', '字典类型', 'education', '/admin/sys-dict-type',
  '/0/2/58/543', 'C', '无', 'admin:sysDictType:list', 58, 0, '',
  '/admin/sys-dict-type/index', 1, '0', '1', 0, 1, NOW(3), NOW(3), NULL
FROM DUAL
WHERE NOT EXISTS (
  SELECT 1 FROM sys_menu WHERE menu_id = 543
);

UPDATE sys_menu
SET
  menu_name = 'SysDictTypeManage',
  title = '字典类型',
  icon = 'education',
  path = '/admin/sys-dict-type',
  paths = '/0/2/58/543',
  menu_type = 'C',
  action = '无',
  permission = 'admin:sysDictType:list',
  parent_id = 58,
  no_cache = 0,
  breadcrumb = '',
  component = '/admin/sys-dict-type/index',
  sort = 1,
  visible = '0',
  is_frame = '1',
  updated_at = NOW(3)
WHERE menu_id = 543;

UPDATE sys_menu
SET
  menu_name = 'SysDictDataManage',
  title = '字典数据',
  icon = 'education',
  path = '/admin/dict/data/:dictId',
  paths = '/0/2/58/59',
  menu_type = 'C',
  action = '无',
  permission = 'admin:sysDictData:list',
  parent_id = 58,
  no_cache = 0,
  breadcrumb = '',
  component = '/admin/sys-dict-data/index',
  sort = 100,
  visible = '1',
  is_frame = '1',
  updated_at = NOW(3)
WHERE menu_id = 59;

UPDATE sys_menu SET parent_id = 543, paths = '/0/2/58/543/236', updated_at = NOW(3) WHERE menu_id = 236;
UPDATE sys_menu SET parent_id = 543, paths = '/0/2/58/543/237', updated_at = NOW(3) WHERE menu_id = 237;
UPDATE sys_menu SET parent_id = 543, paths = '/0/2/58/543/238', updated_at = NOW(3) WHERE menu_id = 238;
UPDATE sys_menu SET parent_id = 543, paths = '/0/2/58/543/239', updated_at = NOW(3) WHERE menu_id = 239;

UPDATE sys_menu SET parent_id = 59, paths = '/0/2/58/59/240', updated_at = NOW(3) WHERE menu_id = 240;
UPDATE sys_menu SET parent_id = 59, paths = '/0/2/58/59/241', updated_at = NOW(3) WHERE menu_id = 241;
UPDATE sys_menu SET parent_id = 59, paths = '/0/2/58/59/242', updated_at = NOW(3) WHERE menu_id = 242;
UPDATE sys_menu SET parent_id = 59, paths = '/0/2/58/59/243', updated_at = NOW(3) WHERE menu_id = 243;

UPDATE sys_menu
SET
  menu_name = 'Log',
  title = '日志管理',
  icon = 'log',
  path = '/log',
  paths = '/0/2/211',
  menu_type = 'M',
  action = '无',
  permission = '',
  parent_id = 2,
  no_cache = 0,
  breadcrumb = '',
  component = 'RouteView',
  sort = 80,
  visible = '0',
  is_frame = '1',
  updated_at = NOW(3)
WHERE menu_id = 211;

UPDATE sys_menu
SET
  component = '/admin/sys-login-log/index',
  updated_at = NOW(3)
WHERE menu_id = 212;

UPDATE sys_menu
SET
  menu_name = 'SysOperaLogManage',
  title = '操作日志',
  icon = 'skill',
  path = '/admin/sys-opera-log',
  paths = '/0/2/211/216',
  menu_type = 'C',
  permission = 'admin:sysOperLog:list',
  parent_id = 211,
  component = '/admin/sys-opera-log/index',
  visible = '0',
  is_frame = '1',
  updated_at = NOW(3)
WHERE menu_id = 216;

COMMIT;
