-- 开发工具空页面修复
-- 目标：
-- 1. 系统接口复用当前 Vue 前端的接口管理页
-- 2. 表单构建复用后端静态 form-generator，并通过前端桥接页承接
-- 3. 代码生成接入当前前端最小可用生成器页
-- 4. 代码生成修改页延后，先兼容到生成器主页，避免继续落 404

START TRANSACTION;

UPDATE sys_menu
SET
  menu_name = 'Swagger',
  title = '系统接口',
  icon = 'guide',
  path = '/dev-tools/swagger',
  paths = '/0/60/61',
  menu_type = 'C',
  action = '无',
  permission = '',
  parent_id = 60,
  no_cache = 0,
  breadcrumb = '',
  component = '/admin/sys-api/index',
  sort = 1,
  visible = '0',
  is_frame = '1',
  updated_at = NOW(3)
WHERE menu_id = 61;

UPDATE sys_menu
SET
  menu_name = 'Gen',
  title = '代码生成',
  icon = 'code',
  path = '/dev-tools/gen',
  paths = '/0/60/261',
  menu_type = 'C',
  parent_id = 60,
  component = '/dev-tools/gen/index',
  sort = 2,
  visible = '0',
  is_frame = '1',
  updated_at = NOW(3)
WHERE menu_id = 261;

UPDATE sys_menu
SET
  menu_name = 'EditTable',
  title = '代码生成修改',
  icon = 'build',
  path = '/dev-tools/editTable',
  paths = '/0/60/262',
  menu_type = 'C',
  parent_id = 60,
  component = '/dev-tools/gen/edit',
  sort = 100,
  visible = '1',
  is_frame = '1',
  updated_at = NOW(3)
WHERE menu_id = 262;

UPDATE sys_menu
SET
  menu_name = 'Build',
  title = '表单构建',
  icon = 'build',
  path = '/dev-tools/build',
  paths = '/0/60/264',
  menu_type = 'C',
  parent_id = 60,
  component = '/dev-tools/build/index',
  sort = 1,
  visible = '0',
  is_frame = '1',
  updated_at = NOW(3)
WHERE menu_id = 264;

COMMIT;
