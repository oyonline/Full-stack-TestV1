INSERT INTO sys_menu (
  menu_id, menu_name, title, icon, path, paths, menu_type, action, permission,
  parent_id, no_cache, breadcrumb, component, sort, visible, is_frame,
  create_by, created_at, update_by, updated_at, deleted_at
) VALUES
  (680, 'WorkflowCenter', '流程中心', 'ant-design:deployment-unit-outlined', '/platform/workflow', '/0/537/680', 'M', '', '', 537, false, '', 'RouteView', 20, '0', 1, 1, '2026-03-27 00:00:00.000', 1, '2026-03-27 00:00:00.000', NULL),
  (681, 'WorkflowTodo', '我的待办', 'ant-design:unordered-list-outlined', '/platform/workflow/todo', '/0/537/680/681', 'C', '', 'platform:workflow:todo:list', 680, false, '', '/platform/workflow/todo/index', 10, '0', 1, 1, '2026-03-27 00:00:00.000', 1, '2026-03-27 00:00:00.000', NULL),
  (682, 'WorkflowStarted', '我发起的', 'ant-design:send-outlined', '/platform/workflow/started', '/0/537/680/682', 'C', '', 'platform:workflow:started:list', 680, false, '', '/platform/workflow/started/index', 20, '0', 1, 1, '2026-03-27 00:00:00.000', 1, '2026-03-27 00:00:00.000', NULL)
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
  update_by = VALUES(update_by),
  updated_at = VALUES(updated_at),
  deleted_at = NULL;

INSERT IGNORE INTO sys_role_menu (role_id, menu_id) VALUES
  (2, 680),
  (2, 681),
  (2, 682),
  (3, 680),
  (3, 681),
  (3, 682);
