-- 后端菜单接入：Batch3 五个页面（字典类型、字典数据、登录日志、操作日志、接口管理）
-- 用途：使前端 web-antd 左侧菜单能正常访问上述页面（component 与前端 views 路径一致）
-- 执行方式：在已初始化过的库中执行本文件，仅更新 sys_menu 表数据，不涉及角色/权限表
-- 适用：MySQL / PostgreSQL（若为 PG，将 true/false 改为 t/f 或 1/0 视库而定）

-- 1. 字典类型（原 menu_id=58 为「字典管理」，改为「字典类型」并指向 admin/sys-dict-type/index）
UPDATE sys_menu SET
  menu_name = 'SysDictType',
  title = '字典类型',
  path = '/admin/sys-dict-type',
  paths = '/0/2/58',
  component = 'admin/sys-dict-type/index'
WHERE menu_id = 58;

-- 2. 字典数据（原 menu_id=59，改为独立页路径 admin/sys-dict-data/index）
UPDATE sys_menu SET
  menu_name = 'SysDictData',
  title = '字典数据',
  path = '/admin/sys-dict-data',
  paths = '/0/2/59',
  component = 'admin/sys-dict-data/index'
WHERE menu_id = 59;

-- 3. 登录日志（menu_id=212，仅确保 component 与前端一致）
UPDATE sys_menu SET
  component = 'admin/sys-login-log/index'
WHERE menu_id = 212;

-- 4. 操作日志（menu_id=216，路径与前端 sys-opera-log 一致）
UPDATE sys_menu SET
  path = '/admin/sys-opera-log',
  paths = '/0/2/211/216',
  component = 'admin/sys-opera-log/index'
WHERE menu_id = 216;

-- 5. 接口管理（menu_id=528，修正 paths 的 parent 为 2，component 与前端一致）
UPDATE sys_menu SET
  paths = '/0/2/528',
  component = 'admin/sys-api/index'
WHERE menu_id = 528;
