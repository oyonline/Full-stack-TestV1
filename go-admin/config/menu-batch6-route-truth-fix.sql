-- 菜单真相源收口
-- 目标：
-- 1. 清理字典/日志/开发工具的历史组件路径，让数据库菜单直接对齐当前正式前端页面
-- 2. 为系统接口使用正式开发工具入口 /admin/sys-api
-- 3. 将旧的隐藏接口管理页移动到内部保留路径，避免生成重复路由

START TRANSACTION;

UPDATE sys_menu
SET
  menu_name = 'Dict',
  title = '字典管理',
  icon = 'education',
  path = '/admin/dict',
  paths = '/0/2/58',
  menu_type = 'M',
  component = 'RouteView',
  visible = '0',
  updated_at = NOW(3)
WHERE menu_id = 58;

INSERT INTO sys_menu (
  menu_id, menu_name, title, icon, path, paths, menu_type, action,
  permission, parent_id, no_cache, breadcrumb, component, sort,
  visible, is_frame, created_by, updated_by, created_at, updated_at, deleted_at
)
VALUES (
  543, 'SysDictTypeManage', '字典类型', 'education', '/admin/sys-dict-type',
  '/0/2/58/543', 'C', '无', 'admin:sysDictType:list', 58, false, '',
  '/admin/sys-dict-type/index', 1, '0', '1', 0, 1, NOW(3), NOW(3), NULL
)
ON DUPLICATE KEY UPDATE
  menu_name = VALUES(menu_name),
  title = VALUES(title),
  icon = VALUES(icon),
  path = VALUES(path),
  paths = VALUES(paths),
  menu_type = VALUES(menu_type),
  action = VALUES(action),
  permission = VALUES(permission),
  parent_id = VALUES(parent_id),
  no_cache = VALUES(no_cache),
  breadcrumb = VALUES(breadcrumb),
  component = VALUES(component),
  sort = VALUES(sort),
  visible = VALUES(visible),
  is_frame = VALUES(is_frame),
  updated_by = VALUES(updated_by),
  updated_at = NOW(3),
  deleted_at = NULL;

UPDATE sys_menu
SET
  menu_name = 'Log',
  title = '日志管理',
  icon = 'log',
  path = '/log',
  paths = '/0/2/211',
  menu_type = 'M',
  component = 'RouteView',
  visible = '0',
  updated_at = NOW(3)
WHERE menu_id = 211;

UPDATE sys_menu
SET
  menu_name = 'SysOperaLogManage',
  title = '操作日志',
  path = '/admin/sys-opera-log',
  component = '/admin/sys-opera-log/index',
  paths = '/0/2/211/216',
  menu_type = 'C',
  visible = '0',
  updated_at = NOW(3)
WHERE menu_id = 216;

UPDATE sys_menu
SET
  menu_name = 'Swagger',
  title = '系统接口',
  icon = 'guide',
  path = '/admin/sys-api',
  component = '/admin/sys-api/index',
  permission = 'admin:sysApi:list',
  paths = '/0/60/61',
  parent_id = 60,
  menu_type = 'C',
  visible = '0',
  updated_at = NOW(3)
WHERE menu_id = 61;

UPDATE sys_menu
SET
  menu_name = 'Build',
  title = '表单构建',
  path = '/dev-tools/build',
  component = '/dev-tools/build/index',
  paths = '/0/60/264',
  parent_id = 60,
  menu_type = 'C',
  visible = '0',
  updated_at = NOW(3)
WHERE menu_id = 264;

UPDATE sys_menu
SET
  menu_name = 'Gen',
  title = '代码生成',
  path = '/dev-tools/gen',
  component = '/dev-tools/gen/index',
  paths = '/0/60/261',
  parent_id = 60,
  menu_type = 'C',
  visible = '0',
  updated_at = NOW(3)
WHERE menu_id = 261;

UPDATE sys_menu
SET
  menu_name = 'EditTable',
  title = '代码生成修改',
  path = '/dev-tools/editTable',
  component = '/dev-tools/gen/edit',
  paths = '/0/60/262',
  parent_id = 60,
  menu_type = 'C',
  visible = '1',
  updated_at = NOW(3)
WHERE menu_id = 262;

UPDATE sys_menu
SET
  menu_name = 'SysApiManage',
  title = '接口管理(内置)',
  path = '/__internal/sys-api',
  component = 'RouteView',
  permission = '',
  visible = '1',
  updated_at = NOW(3)
WHERE menu_id = 528;

COMMIT;
