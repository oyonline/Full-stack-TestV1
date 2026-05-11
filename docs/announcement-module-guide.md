# 公告模块使用指南

本指南面向需要使用、维护或扩展公告模块（admin/sys-announcement）的工程师，
配合 `docs/sku-module-guide.md` 看：公告与 SKU 是当前两个完整接入「平台附件」
能力的业务模块,二者上传链路同源(`/api/v1/platform/attachments/upload`)。

覆盖四件事:

1. **数据模型** —— 三张业务表的列、索引、约束
2. **REST 路由 + 权限码** —— 端点、Casbin 权限点、菜单按钮 ID
3. **wangeditor v5 上传链路** —— 富文本图片从浏览器到落盘的完整时序
4. **附件复用与垃圾回收** —— 当前落地形态(含已知缺口)

> 本模块的 schema、菜单、API 元数据、桥接行全部由
> `1778160000000_announcement.go` 一次性 AutoMigrate + INSERT,迁移幂等。

---

## 1. 数据模型

迁移文件: `go-admin/cmd/migrate/migration/version/1778160000000_announcement.go`
GORM 模型(运行时):
- `go-admin/app/admin/models/announcement.go`
- `go-admin/app/admin/models/announcement_scope.go`
- `go-admin/app/admin/models/announcement_read_log.go`

> 迁移目录下的 `migration/models/announcement.go` 是迁移期使用的"无业务方法"
> 影子模型,字段定义与运行时一致。两份模型必须同步改。

### 1.1 `announcement` —— 公告主表

公告本体。富文本正文整段写在 `content`(`mediumtext`),封面图与状态等元数据
也在这张表。

| 列 | 类型 / 约束 | 说明 |
|----|-----------|----|
| `announcement_id` | bigint, PK, AUTO_INCREMENT | 主键 |
| `title` | varchar(200), NOT NULL | 标题 |
| `content` | mediumtext | 富文本 HTML(经 `service.SanitizeContent` XSS 过滤后入库) |
| `cover_image_url` | varchar(512) | 封面图 URL,通常是 `/static/uploadfile/attachment/...` |
| `status` | tinyint, NOT NULL, default 1 | 1=草稿 / 2=已发布 / 3=已下线,常量见 `models/announcement.go:9-13` |
| `is_top` | tinyint, NOT NULL, default 0 | 0/1,是否置顶 |
| `top_sort` | int, NOT NULL, default 0 | 置顶组内排序值,`is_top DESC, top_sort DESC` |
| `publish_at` | datetime, NULL | 生效起始,`OnlyValid=1` 时按它过滤 |
| `expire_at` | datetime, NULL | 失效时间,同上 |
| `creator_id` | int, INDEX | 创建人(独立于 `create_by`,目前两者由 service 一起写) |
| `remark` | varchar(255) | 备注 |
| `create_by` / `update_by` | int | `common.ControlBy` 嵌入,用于 `data_scope` 过滤 |
| `created_at` / `updated_at` / `deleted_at` | datetime | `common.ModelTime` 嵌入,软删除 |

**索引**:GORM AutoMigrate 仅显式建了 `creator_id` 索引。`status`、`is_top`、
`publish_at` 没有单列索引。当前 MVP 数据量下 `is_top DESC, top_sort DESC, COALESCE(publish_at, created_at) DESC`
排序走全表是可接受的;数据量上来后建议补复合索引,见末节"已知缺口"。

**外键**:无显式外键(项目惯例)。`creator_id` / `create_by` 在应用层关联到
`sys_user`,但数据库层不强制。

### 1.2 `announcement_scope` —— 部门可见范围

公告 ↔ 部门的多对多桥接表,驱动「部门可见性」过滤。一个公告可关联 0~N 个部门;
0 个表示"全部门可见"(由 `service.Announcement.GetPage` 处理)。

| 列 | 类型 / 约束 | 说明 |
|----|-----------|----|
| `announcement_id` | bigint, PK | 复合主键之一 |
| `dept_id` | int, PK | 复合主键之一 |

**约束**:复合主键天然去重。GORM 没显式建 FK。`service.Announcement.Insert/Update`
里有应用层去重(`seen` map)。

**写时机**:
- `service.Announcement.Insert` —— 新建公告时一次性 batch insert(`announcement.go:266-269`)
- `service.Announcement.Update` —— 仅当请求显式带 `DeptIds` 时,先 `Delete` 后 `Create` 重建(`announcement.go:305-332`),"显式 nil" 表示"不动 scope"
- `service.Announcement.Remove` —— 公告级联删除时同步 `Delete`(`announcement.go:358-359`)

### 1.3 `announcement_read_log` —— 已读记录

每个用户对每个公告的"首次已读时间",`MarkRead` 接口写入,详情/列表派生
`isRead` / `readCount` 都靠它聚合。

| 列 | 类型 / 约束 | 说明 |
|----|-----------|----|
| `user_id` | int, PK | 复合主键之一 |
| `announcement_id` | bigint, PK | 复合主键之一 |
| `read_at` | datetime | 首次标记已读的时间戳 |

**幂等**:`service.Announcement.MarkRead` 先 `Count` 再 `Create`,并在 `Create`
失败时按"复合主键冲突 = 并发写入,视为成功"处理(`announcement.go:404-413`)。

**级联清理**:仅在 `service.Announcement.Remove` 中按 `announcement_id IN ?`
批量删除(`announcement.go:362-364`)。删除用户时不级联(应用未实现)。

---

## 2. REST 路由与权限码

### 2.1 路由表

注册入口: `go-admin/app/admin/router/announcement.go`(`/api/v1/announcement` 路由组,
`AuthCheckRole` + `PermissionAction` 中间件)。

| Method | Path | Handler | 用途 |
|--------|------|---------|----|
| GET    | `/api/v1/announcement`          | `apis.Announcement.GetPage`  | 列表/搜索,支持 `onlyValid` / `onlyVisible` |
| GET    | `/api/v1/announcement/:id`      | `apis.Announcement.Get`      | 详情(派生 `deptIds` / `isRead` / `readCount`) |
| POST   | `/api/v1/announcement`          | `apis.Announcement.Insert`   | 新建,正文走 `SanitizeContent` |
| PUT    | `/api/v1/announcement/:id`      | `apis.Announcement.Update`   | 编辑(可选重建 scope) |
| DELETE | `/api/v1/announcement`          | `apis.Announcement.Delete`   | 批量删除,体内带 `ids` |
| POST   | `/api/v1/announcement/:id/read` | `apis.Announcement.MarkRead` | 当前用户标记已读(幂等) |

前端 SDK 一一对应: `vue-vben-admin/apps/web-antd/src/api/core/announcement.ts`。

### 2.2 Casbin 权限点 (rolekey × path × action)

权限通过 `sys_menu_api_rule` 把 `sys_menu`(挂权限码)与 `sys_api`(挂 path+action)
桥接,Casbin 实际比对的是 `(rolekey, path, action)`。迁移时插入的 6 行 `sys_api`:

| sys_api.path                     | action | 关联 sys_menu.permission        | 业务含义 |
|----------------------------------|--------|---------------------------------|----|
| `/api/v1/announcement`           | GET    | `admin:announcement:list`       | 列表 → 父菜单 |
| `/api/v1/announcement/:id`       | GET    | `admin:announcement:list`       | 详情 → 父菜单 |
| `/api/v1/announcement`           | POST   | `admin:announcement:add`        | 新建 → "新增公告" 按钮 |
| `/api/v1/announcement/:id`       | PUT    | `admin:announcement:edit`       | 编辑 → "修改公告" 按钮 |
| `/api/v1/announcement`           | DELETE | `admin:announcement:remove`     | 删除 → "删除公告" 按钮 |
| `/api/v1/announcement/:id/read`  | POST   | `admin:announcement:list`       | 标记已读 → 父菜单(只要能进详情就能调) |

挂菜单结构:

- 父菜单 `AnnouncementManage` —— `permission=admin:announcement:list`、
  `path=/admin/sys-announcement`、`component=admin/sys-announcement/index`、
  父菜单 ID = `sys_menu` 表中"系统管理"(ParentId=2)
- 4 个 F 类按钮:`add` / `edit` / `remove` / `read`(标记已读权限按钮虽建了,
  但当前桥接里 `MarkRead` 实际挂在父菜单的 `list` 权限上,见迁移
  `1778160000000_announcement.go:119`)

### 2.3 数据权限(data_scope)

走通用的 `actions.Permission(tableName, p)` GORM Scope。生效路径:

- `GetPage` / `Get` / `Update` / `Remove` / `MarkRead` 都把 Permission 作为
  必经 Scope 应用到 `announcement` 主表上
- `data_scope=1`(全部):看全部公告
- `data_scope=3/4`(本部门 / 本部门及以下):仅看 `create_by` 在对应部门下的公告
- `data_scope=5`(仅本人):仅看自己创建的公告

> 详情/更新/删除/已读的「先按 Permission 过滤再操作」是越权防护:`data_scope=5`
> 用户拿到任意 `id` 也无法越权操作他人公告(返回"公告不存在")。
> 见 `service/announcement.go:190-200`、`286-298`、`342-356`、`378-392`。

`announcement_scope` 是「部门可见性」过滤(读侧 OnlyVisible),与 data_scope
正交,二者都满足才出现在结果中。

---

## 3. wangeditor v5 上传链路

公告模块有三处上传入口,全部复用平台附件接口:

| 入口 | 组件 | businessType | businessId | moduleKey |
|------|------|--------------|------------|-----------|
| 富文本内嵌图 | `rich-text-editor.vue` | `announcement-inline` | `'0'`(常量) | `admin` |
| 公告封面 | `cover-upload.vue` | `announcement-cover` | `'0'`(新建)或公告 ID | `admin` |
| 普通附件 | `attachment-section.vue` | `announcement` | 公告 ID | `admin` |

### 3.1 内嵌图(wangeditor)端到端时序

源码:
- 前端: `vue-vben-admin/apps/web-antd/src/views/admin/sys-announcement/rich-text-editor.vue`
- 客户端 SDK: `vue-vben-admin/apps/web-antd/src/api/core/attachment.ts`(`uploadPlatformAttachment`)
- 后端路由: `go-admin/app/platform/router/attachment.go`
- 后端 handler: `go-admin/app/platform/apis/attachment.go`(`Attachment.Upload`)
- 后端 service: `go-admin/app/platform/service/attachment.go`(`Attachment.Upload`)

```
┌─────────────────────────────────────────────────────────────────────┐
│ 1. 用户在 wangeditor 工具栏点 "图片",选择本地文件                    │
│    rich-text-editor.vue 的 MENU_CONF.uploadImage.customUpload 接管  │
│    (默认上传链路被禁用,完全自定义)                                  │
└─────────────────────────────────────────────────────────────────────┘
                                ↓
┌─────────────────────────────────────────────────────────────────────┐
│ 2. customUpload 调 uploadPlatformAttachment({                        │
│      file, businessType: 'announcement-inline',                     │
│      businessId: '0', moduleKey: 'admin' })                         │
│    SDK 拼 multipart/form-data,fetch POST /api/v1/platform/          │
│    attachments/upload,带 Bearer token                               │
│    (注:此处用原生 fetch,不走 requestClient,避免拦截器重排           │
│     FormData)                                                        │
└─────────────────────────────────────────────────────────────────────┘
                                ↓
┌─────────────────────────────────────────────────────────────────────┐
│ 3. Gin 路由命中 platform/router/attachment.go 中的                   │
│    /v1/platform/attachments/upload (POST),AuthCheckRole 中间件      │
│    校验 (rolekey, path, action) 通过 Casbin                         │
└─────────────────────────────────────────────────────────────────────┘
                                ↓
┌─────────────────────────────────────────────────────────────────────┐
│ 4. apis.Attachment.Upload 解析 form 字段(req)+ FormFile("file"),   │
│    调 service.Attachment.Upload                                      │
└─────────────────────────────────────────────────────────────────────┘
                                ↓
┌─────────────────────────────────────────────────────────────────────┐
│ 5. service.Attachment.Upload (platform/service/attachment.go:57):    │
│    a. EnsureModuleEnabled(req.ModuleKey) —— 模块未启用直接 403       │
│    b. 计算落盘目录:                                                  │
│       static/uploadfile/attachment/<YYYYMM>/<uuid><ext>             │
│    c. utils.IsNotExistMkDir 建月份目录                               │
│    d. c.SaveUploadedFile(fileHeader, storagePath) —— 写盘            │
│    e. 在 att_file 表插入一行,storage_type='local'                    │
│    f. 写 audit log (platform.attachment.upload)                     │
│    返回 *AttachmentFile (含 storagePath)                             │
└─────────────────────────────────────────────────────────────────────┘
                                ↓
┌─────────────────────────────────────────────────────────────────────┐
│ 6. 前端把 result.storagePath 拼成 `/${storagePath}` 作为图片 URL    │
│    塞回 wangeditor: insertFn(url, file.name, url)                   │
│    最终 HTML 里产生 <img src="/static/uploadfile/attachment/.../    │
│    <uuid>.png"> 这样的相对 URL                                       │
└─────────────────────────────────────────────────────────────────────┘
                                ↓
┌─────────────────────────────────────────────────────────────────────┐
│ 7. 用户点保存:整段 HTML 走 POST /api/v1/announcement                │
│    service.Announcement.Insert 调 SanitizeContent(bluemonday        │
│    UGCPolicy),允许 <img src/alt/width/height>,阻断 script/iframe/  │
│    on* 事件,写入 announcement.content                               │
└─────────────────────────────────────────────────────────────────────┘
```

### 3.2 关键约束

- **后端只支持本地落盘**: `service/attachment.go:24` 的 `attachmentBasePath = "static/uploadfile/attachment"`
  + `dto/attachment.go:9` 的 `AttachmentStorageLocal = "local"`。
  当前没有任何 OSS/S3 实现路径——`StorageType` 字段保留了未来切换的扩展点,
  但 service 层硬编码了 local 分支。
- **静态服务**: `/static/uploadfile/...` 由 Gin 静态目录中间件直接吐出文件,
  不再过 Casbin。任何拿到 URL 的人都能 GET——所以后端**不要**依赖 `att_file`
  表上的 `module_key/business_type/business_id` 做对外鉴权,这只是元数据。
- **前端大小限制**: rich-text-editor 限 10MB / `image/*`(`rich-text-editor.vue:49-50`),
  后端没有显式 size cap,继承 Gin 默认 multipart 上限(32MB)。
- **文件命名**: `<uuid><ext>`(`service/attachment.go:72`)。原始文件名只在
  `att_file.file_name` 与下载时的 `Content-Disposition` 中保留。

### 3.3 临时附件清理策略

采用 **保存时 rebind + 离线 cron 兜底** 的双保险策略(实现见 [EPO-84](mention://issue/0c3e84ff-4903-441f-8ca4-9e8081c1d16a)):

1. **rebind(主路径)**: `service.Announcement.Insert` / `Update` 提交事务前,
   调用 `rebindAnnouncementAttachments(tx, announcementId, content, coverImageUrl)`。
   该函数扫描 `content` 中的 `<img src>` 和 `coverImageUrl`,
   把 `storage_path` 命中且 `business_id='0'` 的 `att_file` 行批量改写为公告 ID。
   - 仅改写 `module_key='admin' AND business_type IN ('announcement-inline','announcement-cover')` 的**临时行**,
     非临时行不受影响。
   - `NormalizeStoragePath` 严格限定前缀 `static/uploadfile/attachment/`,
     外部 URL 或相对路径不会误命中。

2. **cron 兜底**: `AnnouncementAttachmentGC` job 按 `sys_job` 配置的 cron 表达式运行,
   删除 `business_id='0' AND created_at < NOW()-24h` 的临时附件。
   覆盖「用户取消编辑 / 关闭浏览器」等 rebind 无法触及的场景。

参见 `app/admin/service/announcement_attachments.go` 与 `app/jobs/announcement_attachment_gc.go`。

---

## 4. 附件复用与垃圾回收

### 4.1 当前的复用模型

`att_file` 是**单条记录 / 单文件**模型,不是"逻辑附件 + 物理文件"二张表的引用计数模型。
也就是说,**每一次上传都新建一个 UUID 文件 + 一行 `att_file`**。
"跨公告复用同一附件"在当前实现下没有内置支持:

- 附件区(`attachment-section.vue`):用户在公告 A 上传后再在公告 B 上传同一文件,
  会产生两份物理文件、两行 `att_file`(`business_id` 各为 A、B 的 ID)。
- 富文本内嵌图:公告 A 富文本里的图片 URL 可以被复制粘贴到公告 B 的富文本中
  (浏览器层面就是一个 `<img src>`),但这个"复用"只是 URL 字符串复用,
  `att_file` 仍然只对应原本的那次上传(`business_id='0'`),没有 ref-count 概念。

### 4.2 垃圾回收

**当前实现:两级 GC**(实现见 [EPO-84](mention://issue/0c3e84ff-4903-441f-8ca4-9e8081c1d16a))。

#### 级联删除(主路径)
删除公告时,`service.Announcement.Remove` 在事务内调用:
```go
removeAnnouncementAttachments(tx, announcementIds)
```
- 按 `module_key='admin' AND business_type IN ('announcement','announcement-inline','announcement-cover')`
  查出所有关联 `att_file` 行的 `storage_path`;
- 事务内 `DELETE` 这些记录;
- 返回的 `storagePaths` 在事务提交后由调用方 `os.Remove` 做物理删除(best-effort,
  失败不影响主事务,留 cron 兜底)。

#### 离线 cron 兜底
`AnnouncementAttachmentGC` 两阶段清理:

1. **临时孤儿**: `business_id='0'` 且 `created_at < NOW()-24h`。
2. **编辑替换孤儿**: 逐公告扫描,若某 `att_file` 行的 `storage_path`
   已不在该公告 `content` / `cover_image_url` 中,则视为编辑时换图/删图产生的残留,
   一并清理。

默认 `dry_run=true`,运维需通过环境变量或配置显式切 `dry_run=false` 才执行真删。

此外,用户仍可在前端附件列表单条删除 → `DELETE /api/v1/platform/attachments/{id}`
→ `service.Attachment.Delete`(事务内删 `att_file` 行 + `os.Remove`)。
该路径权限校验不变(仅上传人本人或 admin 角色可删)。

### 4.3 与实现不一致 / 已知缺口

> 本节是 EPO-64 文档化过程中发现的、**与任务描述里"应有"行为存在差距**的事项。
> 按 EPO-64 边界**不在本任务内修复**,需开 issue 跟进。

- **TODO: 与实现不一致 — 需开 issue 修** —— 任务描述要求"附件复用机制:跨公告复用
  同一附件如何记 ref-count、如何垃圾回收"。当前 `att_file` 没有 ref-count 列,
  也没有"逻辑附件 + 物理文件"分表,跨公告复用 = 重复上传重复落盘。建议方案:
  a) 增加 `file_hash` (sha256) + `ref_count` 列,按 hash 去重物理文件;
  b) 或拆 `att_logical` / `att_blob` 两张表,逻辑层维护 ref。
- ✅ **已实现** —— 临时附件清理:保存时 rebind + 离线 cron 兜底(见 §3.3 / §4.2)。
  实现 commit: `42568e9`, hotfix: `c948388`。
- ✅ **已实现** —— 删除公告级联清附件:`Remove` 事务内调用 `removeAnnouncementAttachments`
  删 `att_file` 行,commit 后 best-effort 物理删(见 §4.2)。
  实现 commit: `42568e9`。
- **TODO: 与实现不一致 — 需开 issue 修** —— 迁移在桥接 `MarkRead` 时挂在父
  菜单(`list` 权限),而不是迁移里同时建好的"标记公告已读"按钮(`admin:announcement:read`)
  上(`1778160000000_announcement.go:119`)。当前所有能进列表的人都能调 read,
  `read` 按钮事实上没生效。如果产品需要更细粒度,需要把桥接换到 `buttons[3].MenuId`。
- **次级建议** —— `announcement` 表只有 `creator_id` 索引,列表页排序键
  `(is_top, top_sort, publish_at, created_at)` 没有覆盖索引。MVP 数据量下不影响,
  数据量上来再补 `idx_announcement_top_publish (is_top, top_sort, publish_at)`。

---

## 5. 验收 & 排查清单

| 想确认的事 | 看哪 |
|-----------|----|
| 用户拿到的列表是不是按数据权限过滤了 | `service/announcement.go:54-67` 的 `actions.Permission` Scope |
| 已读/未读为什么不准 | `announcement_read_log` 表 `(user_id, announcement_id)` 是否有行 |
| 富文本图片 404 | `att_file` 里查 `storage_path`,然后 `ls static/uploadfile/attachment/<YYYYMM>/<uuid>.<ext>` |
| 谁、什么时候上传了哪个附件 | `att_file.uploader_id/uploader_name/created_at` + audit `platform.attachment.upload` |
| 公告新建/修改/删除审计 | audit method = `admin.announcement.{insert,update,delete,markRead}`,target type = `AuditCategoryAnnouncement` |
| 数据权限漏过 | 所有写路径(`Update` / `Remove` / `MarkRead`)都先按 `actions.Permission` 拉一次,过不到就视为不存在。绕开它需要直接走 SQL |
