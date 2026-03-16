# 后端菜单接入（Batch3 五页）

本文档说明「字典类型、字典数据、登录日志、操作日志、接口管理」五个菜单在后端的接入方式，**仅做菜单数据配置，未改前端代码，未做联调**。

---

## 1. 实际改动的文件

| 文件 | 改动类型 |
|------|----------|
| `go-admin/config/menu-batch3-web-antd.sql` | **新增**：可执行 SQL，用于更新 `sys_menu` 表中上述 5 个菜单项的 path / component / paths 等 |
| `docs/menu-batch3-backend-access.md` | **新增**：本说明文档 |

**未改动**：前端代码、角色/权限逻辑、其他后端业务代码。

---

## 2. 每个菜单项的 path / name / component / parentId / sort / icon / permission

以下为执行 SQL 后，五个菜单在 `sys_menu` 表中的关键字段（与前端左侧菜单、路由一致）：

| 菜单名称 | path | menu_name (name) | component | parent_id | sort | icon | permission |
|----------|------|------------------|-----------|------------|------|------|------------|
| 字典类型 | `/admin/sys-dict-type` | SysDictType | admin/sys-dict-type/index | 2 | 60 | education | admin:sysDictType:list |
| 字典数据 | `/admin/sys-dict-data` | SysDictData | admin/sys-dict-data/index | 2 | 100 | education | admin:sysDictData:list |
| 登录日志 | `/admin/sys-login-log` | SysLoginLogManage | admin/sys-login-log/index | 211 | 1 | logininfor | admin:sysLoginLog:list |
| 操作日志 | `/admin/sys-opera-log` | OperLog | admin/sys-opera-log/index | 211 | 1 | skill | admin:sysOperLog:list |
| 接口管理 | `/admin/sys-api` | SysApiManage | admin/sys-api/index | 2 | 0 | api-doc | admin:sysApi:list |

说明：

- **parent_id**：2 = 系统管理；211 = 日志管理（其下为登录日志、操作日志）。
- **paths**：58/59 为 `/0/2/58`、`/0/2/59`；212/216 为 `/0/2/211/212`、`/0/2/211/216`；528 为 `/0/2/528`。
- **menu_type**：均为 `C`（菜单）。
- **visible**：`0` 显示、`1` 隐藏；上述五项均为显示。

---

## 3. 可执行 SQL（种子数据方式）

系统通过**初始化/种子 SQL** 维护菜单时，直接执行以下文件即可完成本批菜单接入：

**文件路径**：`go-admin/config/menu-batch3-web-antd.sql`

**执行方式示例**（在项目根目录或 go-admin 目录下）：

```bash
# MySQL
mysql -u用户名 -p 数据库名 < go-admin/config/menu-batch3-web-antd.sql

# 或登录 MySQL 后
source /path/to/go-admin/config/menu-batch3-web-antd.sql;
```

**作用**：仅对 `sys_menu` 表做 5 条 `UPDATE`（menu_id = 58, 59, 212, 216, 528），不插入新行、不改角色/权限表。若库中已有这些 menu_id（与当前 go-admin 默认 `config/db.sql` 一致），执行后即可在「系统管理」/「日志管理」下看到上述 5 个菜单并正确打开对应前端页。

---

## 4. 若通过后台界面录入（完整录入字段表）

若采用**后台菜单管理界面**逐条录入，可按下表填写（仅列出与路由/前端 component 相关的必填与建议字段）：

| 字段 | 字典类型 | 字典数据 | 登录日志 | 操作日志 | 接口管理 |
|------|----------|----------|----------|----------|----------|
| **菜单名称 (menuName)** | SysDictType | SysDictData | SysLoginLogManage | OperLog | SysApiManage |
| **显示名称 (title)** | 字典类型 | 字典数据 | 登录日志 | 操作日志 | 接口管理 |
| **路由 path** | /admin/sys-dict-type | /admin/sys-dict-data | /admin/sys-login-log | /admin/sys-opera-log | /admin/sys-api |
| **组件 component** | admin/sys-dict-type/index | admin/sys-dict-data/index | admin/sys-login-log/index | admin/sys-opera-log/index | admin/sys-api/index |
| **上级菜单 parentId** | 2（系统管理） | 2（系统管理） | 211（日志管理） | 211（日志管理） | 2（系统管理） |
| **排序 sort** | 60 | 100 | 1 | 1 | 0 |
| **图标 icon** | education | education | logininfor | skill | api-doc |
| **权限标识 permission** | admin:sysDictType:list | admin:sysDictData:list | admin:sysLoginLog:list | admin:sysOperLog:list | admin:sysApi:list |
| **菜单类型 menuType** | C | C | C | C | C |
| **是否可见 visible** | 0 | 0 | 0 | 0 | 0 |
| **是否缓存 noCache** | 否/false | 否/false | 否/false | 否/false | 否/false |
| **paths（若有）** | /0/2/58 | /0/2/59 | /0/2/211/212 | /0/2/211/216 | /0/2/528 |

录入后需保证前端的 `import.meta.glob('../views/**/*.vue')` 能解析到上述 component 路径（即存在 `views/admin/sys-dict-type/index.vue` 等文件），否则路由会落回 not-found。

---

## 5. 小结

- **改动范围**：仅新增 1 个 SQL 文件 + 1 个说明文档；未改前端、未改权限/角色逻辑。
- **执行后效果**：后端菜单数据中，上述 5 个菜单的 path、component 与前端路由和视图路径一致，登录后左侧菜单可正常跳转到对应页面（需角色已分配这些菜单）。
- **本轮未做**：前端代码修改、联调、接口联调。
