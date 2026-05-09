# Migrations Changelog

逐条记录 `go-admin/cmd/migrate/migration/version/*.go` 中需要补文档的 4 条业务 migration。
范围对齐 [EPO-59](#) 文档检查报告 A3 节;不覆盖框架自带的初始化 migration（`1599190683659_tables.go` / `1653638869132_migrate.go`)以及 SKU/SPU、announcement、avatar、platform-core 等已在各自模块文档中描述的 migration（见 `docs/sku-module-guide.md`、`docs/avatar-system.md` 等)。

## 框架前置(共性)

- 所有 migration 都通过 `migration.Migrate.SetVersion(filename[:13], fn)` 注册;`Migrate()` 调度时先 `SELECT ... FROM sys_migration WHERE version = ?`,**已记录的版本会整段跳过**(`go-admin/cmd/migrate/migration/init.go:46-55`)。
- 因此"二次 migrate 的预期 diff"在**框架层默认是空**:同一台库上重跑 `go run cmd/migrate/main.go` 不会再次执行 `_177xxx...` 函数,只在 `sys_migration` 上看到原有行,无新增。
- 下文每条仍然单独标注**逻辑层是否幂等**(假设有人手工删除 `sys_migration` 行强制重跑,或在另一台库上从头再来)。
- 全部使用 `db.Transaction(...)` 包裹;失败回滚,不会留下半应用状态。

按时间正序排列。

---

## 1774900000000 — `drop_finance_subsystem`

- **Timestamp / 版本号**:`1774900000000`(文件 `1774900000000_drop_finance_subsystem.go`,提交 `7e5dae5`)
- **目的**:配合 `f5b3312` 的 finance 模块代码删除,清理 MySQL 中 finance 子系统的所有残留(11 张业务表 + sys_api / sys_menu / sys_role_menu / sys_menu_api_rule / sys_casbin_rule / module_registry 中的相关条目)。
- **影响表**:
  - **DROP**(11 张):`cost_center_info`、`cost_center_info_change`、`cost_center_related_customer`、`cost_budget_version`、`cost_budget_version_detail`、`budget_fee_category`、`budget_fee_category_details`、`allocation_rule_settings`、`allocation_rule_settings_dept`、`fee_request_log`、`fee_request`
  - **DELETE**(条件命中):`sys_api`(`path LIKE '/api/v1/cost-center-info%'` 等 10 条前缀)、`sys_menu`(`path / permission / component` 命中 finance / cost-center / cost-budget / budget-fee / allocation-rule / fee-request)、`sys_role_menu`(子查询级联)、`sys_menu_api_rule`(子查询级联)、`sys_casbin_rule`(`ptype='p' AND v1 LIKE '<finance api>%'`)、`module_registry`(`module_key='finance-budget'`)
  - **新增 / 修改 schema**:无(纯清理)
- **回滚要点**:
  - **逻辑层幂等**:✅。`Migrator().DropTable` 内部走 `DROP TABLE IF EXISTS`;所有 DELETE 均为条件 LIKE / 等值删除,二次执行匹配 0 行无副作用。
  - **手工回滚**:不可逆。表结构与 finance 路由 / 菜单 seed 已随代码一起删除,没有保留 down 迁移。如需恢复,必须从 `f5b3312^` 之前的代码 + `1774600000000_platform_core` 之前的库快照中 restore;线上务必先备份。
  - **风险**:`sys_menu` 命中条件用 LIKE,若有非 finance 模块的菜单 `path` 包含 `/finance/` 子串(目前未发现)会被误删。CI 如发现新菜单匹配此前缀需要 review。
  - **二次 migrate 的预期 diff**:框架层跳过,空 diff;若强制再跑,逻辑层匹配 0 行,空 diff。

---

## 1775200000000 — `sys_user_feishu_fields`

- **Timestamp / 版本号**:`1775200000000`(文件 `1775200000000_sys_user_feishu_fields.go`,提交 `68ad6cc`)
- **目的**:补齐运行时 model `app/admin/models/sys_user.go` 已声明、但历史 migration 缺失的 5 个飞书字段。fresh DB 创建用户时会因 `Unknown column 'open_id'` 失败,本 migration 把这 5 列加到 `sys_user`。
- **影响表**:
  - **ALTER**:`sys_user` 新增 5 列
    - `open_id` `varchar(55)` — 飞书用户应用 ID
    - `job_title` `varchar(55)` — 飞书用户职务
    - `open_department_id` `varchar(55)` — 飞书系统部门 ID
    - `open_department_ids` `varchar(255)` — 飞书系统多部门 ID
    - `cn_name` `varchar(25)` — 飞书中文名
- **回滚要点**:
  - **逻辑层幂等**:✅。每列 AddColumn 前先 `Migrator().HasColumn` 判断,已存在则跳过。
  - **手工回滚**:`ALTER TABLE sys_user DROP COLUMN open_id, DROP COLUMN job_title, DROP COLUMN open_department_id, DROP COLUMN open_department_ids, DROP COLUMN cn_name;`,然后 `DELETE FROM sys_migration WHERE version='1775200000000';`。
  - **数据丢失**:回滚会丢失这些列上的飞书集成数据;若线上已有飞书登录用户,务必先导出 backup。
  - **二次 migrate 的预期 diff**:框架层跳过,空 diff;强制再跑时 `HasColumn` 返回 true,逻辑跳过,空 diff。

---

## 1775300000000 — `remove_dict_data_page`

- **Timestamp / 版本号**:`1775300000000`(文件 `1775300000000_remove_dict_data_page.go`,提交 `c987616`)
- **目的**:收口"字典数据"作为独立菜单页的历史方案。`menu_id=59` (`SysDictDataManage`) 已无对应前端 view,是字典菜单"假已完成"的根因;删除 59 与配套查询按钮 240,并把增删改三个 F-按钮(241/242/243)从 59 子级迁到 543('字典类型') 之下。
- **影响表**:
  - **DELETE**:`sys_menu`(menu_id ∈ {59, 240})、`sys_role_menu`(menu_id ∈ {59, 240})、`sys_menu_api_rule`(sys_menu_menu_id ∈ {59, 240})
  - **UPDATE**:`sys_menu`(menu_id ∈ {241, 242, 243} → `parent_id=543`,`paths='/0/2/58/543/<id>'`)
  - **不动**:`sys_dict` 后端表与 `/api/v1/dict/data` 路由保留(给 `字典类型` 页内嵌 tab 复用);`sys_casbin_rule` 不动(V1 是 API path,不绑定 menu_id)
- **回滚要点**:
  - **逻辑层幂等**:✅。DELETE 用固定 menu_id,二次执行匹配 0 行;UPDATE 重设同样的 parent_id / paths,值相同。
  - **手工回滚**:必须从备份恢复 menu_id=59、240 两行的原始记录,以及 241/242/243 的原 `parent_id` / `paths`(原父级是 59,原 `paths` 是 `/0/2/58/59/<id>`)。无法通过 SQL 重建,因为业务侧默认 fresh DB seed 已经不再插入 59 / 240。回滚后还需 `DELETE FROM sys_migration WHERE version='1775300000000';`。
  - **二次 migrate 的预期 diff**:框架层跳过,空 diff;强制再跑时,DELETE 0 行 + UPDATE 重写同值,空 diff。
  - **风险**:依赖固定 menu_id 命中;若 fresh DB seed 中这些 ID 被重新分配或被复用为别的菜单,会误伤。fresh DB 走 `db.sql` seed 后 ID 与本 migration 假设一致,需保持。

---

## 1775400000000 — `cleanup_orphan_menus`

- **Timestamp / 版本号**:`1775400000000`(文件 `1775400000000_cleanup_orphan_menus.go`,提交 `80b7765`)
- **目的**:清理 `sys_menu` 中真实 orphan 菜单。原 spec 列出 7 个候选(262/61/211/460/471/269/537),经前端 `views/` 目录交叉比对后,确认只有 `menu_id=471` (`JobLog → /schedule/log`) 是真叶子 orphan(整个 `vue-vben-admin/apps/web-antd/src/views/schedule/` 目录不存在)。其余 6 个对应的视图 / Layout 都存在,不动。
- **影响表**:
  - **DELETE**:`sys_menu`(menu_id=471)、`sys_role_menu`(menu_id=471)、`sys_menu_api_rule`(sys_menu_menu_id=471)
  - **不动**:父级 459 (`定时任务` Schedule Layout) 与兄弟 460 (`ScheduleManage`) 全部保留;`sys_casbin_rule` / `sys_api` 不动。
  - **方法论**:沿用 1775300000000 的"先清 m2m 再删本体";471 是叶子,无 reparent 步骤。
- **回滚要点**:
  - **逻辑层幂等**:✅。固定 menu_id=471,DELETE 匹配 0 行无副作用。
  - **手工回滚**:从备份 restore `sys_menu` 中 menu_id=471 的原行(component / path / permission / parent_id=459 / paths 等),以及 `sys_role_menu` / `sys_menu_api_rule` 的关联记录;然后 `DELETE FROM sys_migration WHERE version='1775400000000';`。但因为对应前端 view 不存在,即使恢复也无法点开,只是回到"挂着但 404"的旧状态——通常没有回滚必要。
  - **二次 migrate 的预期 diff**:框架层跳过,空 diff;强制再跑时,DELETE 0 行,空 diff。
  - **风险**:若未来重新引入 `views/schedule/` 并复用 menu_id=471,本 migration 不会再次触发(已记录在 `sys_migration`),需要写新的 migration 重建。
