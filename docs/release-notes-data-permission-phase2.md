# Release Notes — 数据权限 phase2 上线（C7-7）

> 草稿。覆盖 phase1 → phase2 的开关切换、面向用户/集成方的可见行为变化、回滚路径。

## 一句话

`settings.application.enabledp` 由 `false` 切换为 `true`，`sys_role.data_scope` 真正生效；之前角色编辑页的"数据范围"下拉是假功能，现在每个选项都会实际收紧后端查询范围。

## 变更亮点（changelog 风格）

### Added
- 数据权限全量启用：`enabledp=true` 默认下发到 `settings.yml` / `settings.full.yml` / `settings.sqlite.yml` / `settings.local.yml.example`（`settings.demo.yml` 一直为 true，不变）。
- announcement service 接入 `dataScope`，5 路 dataScope（`1` 全部 / `2` 自定义 / `3` 本部门 / `4` 本部门及以下 / `5` 仅本人）端到端验收通过（C7-5）。
- 业务模块接入规约：见 `PROJECT_CONVENTIONS.md` 的数据权限章节，新业务 service 必须走统一的 `dataScope` 链路，禁止绕过。

### Changed
- 角色编辑页"数据范围"下拉从 phase1 的"假功能"升级为真生效；既有角色的 `data_scope` 取值会立即影响其挂载用户的可见数据。
- 默认管理员（`dataScope=1`）行为不变：仍可看到全部数据，包括公告 / 业务列表。
- announcement / 后续接入了 `dataScope` 的业务列表查询：当用户挂载的角色 `data_scope ∈ {2,3,4,5}` 时，会按 phase2 的统一策略收紧。

### Migration
- `1778200000000_data_permission_default`（C7-1）已对存量 `sys_role.data_scope` 空值做兜底（默认填 `1` 全部数据）。
- 上线步骤：先跑迁移 → 再 restart 后端，避免空 `data_scope` 在 `enabledp=true` 下被解释成"看不到任何数据"。

### Tooling / Docs
- 角色编辑页"数据范围"下拉新增 tooltip & help text，说明 5 路语义（C7-4）。
- `docs/local-dev-handoff.md` 增补"数据权限（phase2 已启用）"段落，描述本地默认行为与 smoke 验证步骤。
- 路由策略文档：`docs/audit/c7-2-data-permission-routing.md`（C7-2）。

## 部署 / Restart 步骤

```bash
cd <repo>/go-admin
make build-and-migrate          # 必须：1778200000000_data_permission_default 兜底
./go-admin server -c config/settings.yml   # 或对应环境的 settings 文件
```

如有自定义 `config/settings.local.yml`，请确认未显式把 `enabledp` 覆盖为 `false`。

## 验收清单

- [ ] 4 份 settings 文件（`settings.yml` / `settings.full.yml` / `settings.sqlite.yml` / `settings.local.yml.example`）`application.enabledp` 一致为 `true`；`settings.demo.yml` 维持 `true`。
- [ ] `admin`（`dataScope=1`）登录后看到全部公告，行为与 phase1 一致。
- [ ] 新建 `dataScope=5` 测试角色，分配给非 admin 测试用户登录，仅看到该用户 `create_by` 的行。
- [ ] 部署 / restart 文档与本说明一致。

## 回滚

phase2 出现严重数据可见性问题时：

1. 把目标环境的 `settings.*.yml` 中 `application.enabledp` 改回 `false` 并 restart。
2. 行为退回 phase1：`dataScope` 下拉视觉上仍存在但不生效，全部用户可见全部数据。
3. 不需要回滚迁移（`1778200000000_data_permission_default` 仅做兜底填充，对 phase1 行为无副作用）。
