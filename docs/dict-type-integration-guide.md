# 字典类型页联调执行说明

## 【本次改动的大目标】

围绕「字典类型」页面做一次真实联调执行前说明：不修改代码，仅输出如何进入页面、需确认的配置、接口清单、操作顺序、通过标准与风险点，直接服务于人工联调。

---

## 1. 页面进入方式

- **页面文件路径（已确认）**  
  `vue-vben-admin/apps/web-antd/src/views/admin/sys-dict-type/index.vue`

- **路由命中方式（已确认）**  
  - 本项目采用**后端菜单动态生成路由**，无静态路由指向该页面。  
  - 流程：登录后 `generateAccess()` 调用 `getAllMenusApi()` 拉取菜单；`access.ts` 中 `mapSysMenuToRoute()` 将菜单节点的 `component` 与 `import.meta.glob('../views/**/*.vue')` 的 key 做标准化匹配（`normalizeViewPath`），匹配到 `admin/sys-dict-type/index` 即加载上述页面组件。  
  - 路由配置文件：`vue-vben-admin/apps/web-antd/src/router/access.ts`（菜单转路由逻辑在此）。

- **后端菜单需满足的条件（已确认）**  
  - 执行 `go-admin/config/menu-batch3-web-antd.sql` 后，`menu_id = 58` 的菜单项为：  
    - `path = '/admin/sys-dict-type'`  
    - `component = 'admin/sys-dict-type/index'`  
  - 若未执行该 SQL，或该菜单未分配给当前登录角色，左侧菜单不会出现「字典类型」入口，也无法通过菜单进入该页面。

- **是否可从静态路由直接进入**  
  - **不能**。仓库内未检索到指向 `/admin/sys-dict-type` 的静态路由模块；必须依赖后端返回的菜单中包含上述 path 与 component，且当前角色有该菜单权限，才能从侧栏点击进入。

---

## 2. 本地联调前需要确认的配置

- **前端启动命令（已确认）**  
  - 在**仓库根目录** `vue-vben-admin/`：`pnpm dev:antd`（定义在 `vue-vben-admin/package.json`，内部执行 `pnpm -F @vben/web-antd run dev`）。  
  - 在**应用目录** `vue-vben-admin/apps/web-antd/`：`pnpm dev`（对应 `package.json` 的 `"dev": "pnpm vite --mode development"`）。  
  - 开发环境端口：`.env.development` 中 `VITE_PORT=5666`，默认访问为 `http://localhost:5666`。

- **请求基础配置（已确认）**  
  - 请求客户端创建：`vue-vben-admin/apps/web-antd/src/api/request.ts`。  
  - `apiURL` 来源：`const { apiURL } = useAppConfig(import.meta.env, import.meta.env.PROD)`；`useAppConfig` 在 `packages/effects/hooks/src/use-app-config.ts` 中，开发态直接取 `env.VITE_GLOB_API_URL`。  
  - 开发环境下该值来自：`vue-vben-admin/apps/web-antd/.env.development`，内容为 `VITE_GLOB_API_URL=/api`。  
  - 故**开发时 baseURL = `/api`**，请求路径为相对路径（如 `/v1/dict/type`），最终请求 URL = **baseURL + path = `/api/v1/dict/type`**（浏览器地址栏可见为同源，如 `http://localhost:5666/api/v1/dict/type`）。

- **代理配置（已确认）**  
  - 文件：`vue-vben-admin/apps/web-antd/vite.config.mts`。  
  - 配置：`server.proxy['/api']` → `target: 'http://localhost:10086'`，`changeOrigin: true`，`ws: true`。  
  - 含义：浏览器请求 `/api/xxx` 会被代理到 `http://localhost:10086/api/xxx`。  
  - **需人工确认**：go-admin 默认配置为 `config/settings.yml` 中 `port: 8000`。若后端实际运行在 8000，需将 proxy 的 `target` 改为 `http://localhost:8000`，或把后端改为监听 10086，否则联调时接口会连错端口。

- **多环境**  
  - 开发看：`vue-vben-admin/apps/web-antd/.env.development`（`VITE_GLOB_API_URL=/api`）。  
  - 生产看：`vue-vben-admin/apps/web-antd/.env.production`（当前为 mock 地址，联调不涉及）。

---

## 3. 接口清单

以下均为**字典类型页**实际用到的接口（仅本页；`getDictTypeAll` / `getDictTypeOptionSelect` 仅在「字典数据」页使用，此处不列）。

### 3.1 列表接口

- **请求方法**：GET  
- **前端调用函数名**：`getDictTypePage`  
- **API 文件位置**：`vue-vben-admin/apps/web-antd/src/api/core/dict.ts`（约 38–42 行）  
- **请求路径**：`/v1/dict/type`（相对 baseURL，实际为 `/api/v1/dict/type`）  
- **关键请求参数**（query）：  
  - `pageIndex`（number，必带，来自 `pagination.value.current`）  
  - `pageSize`（number，必带，来自 `pagination.value.pageSize`）  
  - `dictName`（string，可选，有输入时带）  
  - `dictType`（string，可选，有输入时带）  
  - `status`（number，可选，筛选时 1 或 2）  
- **前端预期响应关键结构**：`SysDictTypePageResult`：`{ list: SysDictTypeItem[], count: number, pageIndex?: number, pageSize?: number }`。页面逻辑只**必须**使用 `res.list` 和 `res.count`（见 `sys-dict-type/index.vue` 86–87 行）；`pageIndex`/`pageSize` 可选。  
- **触发时机**：页面挂载（onMounted）、点击「查询」、点击「重置」、分页/每页条数变化、新增/编辑/删除成功后刷新。

### 3.2 详情接口

- **请求方法**：GET  
- **前端调用函数名**：`getDictTypeDetail`  
- **API 文件位置**：`vue-vben-admin/apps/web-antd/src/api/core/dict.ts`（约 48–50 行）  
- **请求路径**：`/v1/dict/type/:id`（如 `/api/v1/dict/type/1`）  
- **关键请求参数**：路径参数 `id`（number）。  
- **前端预期响应关键结构**：单条 `SysDictTypeItem`：至少包含 `id, dictName, dictType, status, remark`（及可选 `createdAt` 等），用于填充编辑表单。  
- **触发时机**：点击表格行「编辑」时，在打开编辑弹窗前调用（`openEditModal(record)` 内，约 246–257 行）。

### 3.3 新增接口

- **请求方法**：POST  
- **前端调用函数名**：`createDictType`  
- **API 文件位置**：`vue-vben-admin/apps/web-antd/src/api/core/dict.ts`（约 91–93 行）  
- **请求路径**：`/v1/dict/type`  
- **关键请求体**：`CreateDictTypeData`：`{ dictName: string, dictType: string, status?: number, remark?: string }`。页面提交时带齐四项（status 默认 2，remark 可为空字符串）。  
- **前端预期响应**：注释为 `Promise<number>`，成功解包后可为 id 或任意；页面不依赖返回值，仅根据成功则关弹窗并 `fetchList()`。  
- **触发时机**：新增弹窗中点击「保存」且校验通过（`onAddOk`，约 202–225 行）。

### 3.4 编辑接口

- **请求方法**：PUT  
- **前端调用函数名**：`updateDictType`  
- **API 文件位置**：`vue-vben-admin/apps/web-antd/src/api/core/dict.ts`（约 107–112 行）  
- **请求路径**：`/v1/dict/type/:id`  
- **关键请求体**：`UpdateDictTypeData`：`{ dictName?, dictType?, status?, remark? }`。页面提交当前表单全部字段。  
- **前端预期响应**：注释为 `Promise<number>`，页面不依赖返回值，成功则关弹窗并 `fetchList()`。  
- **触发时机**：编辑弹窗中点击「保存」且校验通过（`onEditOk`，约 268–292 行）。

### 3.5 删除接口

- **请求方法**：DELETE  
- **前端调用函数名**：`deleteDictType`  
- **API 文件位置**：`vue-vben-admin/apps/web-antd/src/api/core/dict.ts`（约 118–120 行）  
- **请求路径**：`/v1/dict/type`（无路径参数）  
- **关键请求体**：`{ ids: number[] }`（如单条删除为 `{ ids: [record.id] }`）。  
- **前端预期响应**：注释为 `Promise<number[]>`，页面不依赖返回值，成功则 `fetchList()`。  
- **触发时机**：点击表格行「删除」并在确认框中点「确定」（`onDelete` 内 onOk，约 309–314 行）。

---

## 4. 建议联调顺序

1. **先开页面**  
   - 启动前端（见上文），登录后确认左侧是否有「字典类型」菜单；若无，检查后端菜单数据是否已执行 Batch3 SQL 且角色有权限。  
   - 点击进入，确认能打开 `sys-dict-type/index.vue` 对应页面（标题「字典类型」、有筛选区与表格）。

2. **再看列表**  
   - 打开开发者工具 Network，筛选 XHR/Fetch。  
   - 观察是否发出 `GET /api/v1/dict/type?pageIndex=1&pageSize=10`（无筛选时）。  
   - 重点看：状态码 200、响应体为 `{ code: 200, data: { list, count, ... } }`，且 `data.list` 为数组、`data.count` 为数字；表格有数据或空表、无报错。

3. **再试筛选**  
   - 输入字典名称/字典类型、选择状态，点击「查询」。  
   - 观察请求 query 是否带上 `dictName`/`dictType`/`status`，以及返回 list/count 是否符合筛选预期。  
   - 点击「重置」后应恢复无筛选参数的请求并刷新列表。

4. **再试分页**  
   - 改页码或每页条数，观察请求中 `pageIndex`/`pageSize` 变化及返回 list/count 是否对应当前页。

5. **再试新增**  
   - 点击「新增」，填写必填项（字典名称、字典类型）及状态、备注，点「保存」。  
   - 观察 `POST /api/v1/dict/type`，body 为 `{ dictName, dictType, status, remark }`；成功则弹窗关闭且列表刷新，且新记录出现（或列表接口再次被调用）。

6. **再试编辑**  
   - 点击某行「编辑」，确认先发 `GET /api/v1/dict/type/:id` 再打开表单；修改后点「保存」。  
   - 观察 `PUT /api/v1/dict/type/:id` 及 body；成功则弹窗关闭、列表刷新且该行数据更新。

7. **再试删除**  
   - 点击某行「删除」，确认框中点「确定」。  
   - 观察 `DELETE /api/v1/dict/type`，请求体为 `{ ids: [id] }`；成功则列表刷新且该行消失。

---

## 5. 通过标准

- **列表联通通过**：  
  - 进入页面后自动发起 GET `/api/v1/dict/type`，返回 200，且 `data.list` 为数组、`data.count` 为数字；表格正常展示或展示空状态，无控制台/接口报错。

- **新增通过**：  
  - POST `/api/v1/dict/type` 返回 200（或业务约定成功码）；弹窗关闭、列表刷新；再次请求列表时能看到新记录（或至少不报错）。

- **编辑通过**：  
  - GET `/api/v1/dict/type/:id` 返回 200 且数据正确回填表单；PUT `/api/v1/dict/type/:id` 返回 200；弹窗关闭、列表刷新且该行展示为修改后内容。

- **删除通过**：  
  - DELETE `/api/v1/dict/type` 携带 body `{ ids: [id] }` 返回 200；列表刷新后该行消失或列表接口再次被调用且数据一致。

---

## 6. 风险点

1. **代理 target 与后端端口不一致**  
   - 现状：`vite.config.mts` 中 proxy 指向 `http://localhost:10086`，go-admin 默认 `settings.yml` 为 8000。  
   - 风险：请求被代理到错误端口，列表/详情/增删改均无法联通。  
   - 建议：联调前确认后端实际端口，将 proxy `target` 改为该端口，或后端改为 10086。

2. **后端返回结构与前端解包不一致**  
   - 现状：`request.ts` 使用 `defaultResponseInterceptor`，`dataField: 'data'`，`successCode: (code) => code === 0 || code === 200`；`requestClient` 配置 `responseReturn: 'data'`，即业务代码拿到的是 `response.data`（即后端 body 里的 `data` 字段）。  
   - 风险：若后端返回为 `{ list, count }` 直接作为 body 而不包在 `data` 下，或 code 非 0/200，前端会拿错数据或走错误分支。  
   - 建议：确认 go-admin 字典类型相关接口统一为 `{ code: 200, data: { list, count, ... } }` 形式。

3. **DELETE 使用 body 是否被后端接受**  
   - 现状：前端 `deleteDictType` 使用 `requestClient.delete('/v1/dict/type', { data: { ids } })`，即 DELETE 带 body。  
   - 风险：部分网关或框架对 DELETE body 支持不好或默认不解析，导致后端收不到 `ids`。  
   - 建议：联调时若删除一直失败，先在后端或抓包确认是否收到 body；若后端改为 query 或 path 传 id，再考虑改前端（本轮仅说明风险，不改代码）。

4. **分页/筛选参数名与后端不一致**  
   - 现状：前端固定使用 `pageIndex`、`pageSize` 及 `dictName`、`dictType`、`status`（与 dto 注释一致）。  
   - 风险：若 go-admin 实际使用 `page`/`pageNum`/`limit` 等，或 form tag 与前端不一致，会导致列表一直空或分页错乱。  
   - 建议：对照后端 `SysDictTypeGetPageReq` 与路由绑定方式，确认 query 参数名一致。

5. **状态值语义**  
   - 现状：前端 1=停用、2=启用（见 `statusOptions` 与 `statusEditOptions`）。  
   - 风险：若后端 1=启用、2=停用，则展示与业务含义相反。  
   - 建议：与后端约定一致，或在联调时用实际数据验证展示是否正确。

---

## 7. 下一步最建议做什么

**只做一件事**：在本地先确认「代理 target 与后端端口一致」：若 go-admin 跑在 8000，则把 `vue-vben-admin/apps/web-antd/vite.config.mts` 里 proxy `/api` 的 `target` 改为 `http://localhost:8000`（或把后端改为 10086），然后启动前端与后端，从菜单进入字典类型页，看列表 GET 是否 200 且表格有数据或空表无报错；再按第 4 节顺序依次试筛选、分页、新增、编辑、删除。
