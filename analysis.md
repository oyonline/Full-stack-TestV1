# Full-stack-TestV1 项目技术栈与架构分析

> 本文档用于指导脚手架生成，基于对 go-admin + vue-vben-admin 的深度分析

---

## 1. 项目整体架构

### 1.1 技术栈概览

| 层级 | 技术 | 版本 |
|------|------|------|
| 后端 | Go + Gin + GORM + JWT + Casbin | Go 1.24+ |
| 前端 | Vue 3 + Vite + Ant Design Vue | Vue 3.4+ |
| 状态管理 | Pinia | 2.x |
| 路由 | Vue Router | 4.x |
| HTTP | Axios (封装在 @vben/request) | - |
| 构建 | Vite | 5.x |
| 包管理 | pnpm | 8.x |

### 1.2 目录结构

```
Full-stack-TestV1/
├── go-admin/                    # 后端项目
│   ├── app/admin/
│   │   ├── apis/               # API 处理器 (Controller)
│   │   ├── models/             # 数据模型 (Model)
│   │   ├── service/            # 业务逻辑层
│   │   │   └── dto/            # 数据传输对象
│   │   └── router/             # 路由配置
│   ├── common/
│   │   ├── models/             # 通用模型基类
│   │   ├── dto/                # 通用 DTO
│   │   └── middleware/         # 中间件
│   └── config/                 # 配置文件
│
└── vue-vben-admin/             # 前端项目
    ├── apps/web-antd/src/
    │   ├── api/core/           # API 封装
    │   ├── views/admin/        # 管理页面
    │   ├── router/             # 路由配置
    │   └── store/              # Pinia Store
    └── packages/               # 共享包
        ├── @core/              # 核心框架
        ├── stores/             # 状态库
        └── types/              # TS 类型
```

---

## 2. 后端架构模式 (go-admin)

### 2.1 分层架构

```
HTTP Request
    │
    ▼
Router (路由定义 + 中间件)
    │
    ▼
API Handler (gin.Context 处理)
    │
    ▼
Service (业务逻辑)
    │
    ▼
Model (GORM 数据模型)
    │
    ▼
Database
```

### 2.2 核心文件模板

#### 2.2.1 Model (models/sys_xxx.go)

```go
package models

import "go-admin/common/models"

type SysXxx struct {
    models.Model
    Field1    string `json:"field1" gorm:"size:128;comment:字段1"`
    Field2    int    `json:"field2" gorm:"comment:字段2"`
    models.ControlBy
    models.ModelTime
}

func (*SysXxx) TableName() string {
    return "sys_xxx"
}

func (e *SysXxx) Generate() models.ActiveRecord {
    o := *e
    return &o
}

func (e *SysXxx) GetId() interface{} {
    return e.Id
}
```

**关键特征：**
- 嵌入 `models.Model` (包含 Id 主键)
- 嵌入 `models.ControlBy` (createBy/updateBy)
- 嵌入 `models.ModelTime` (createdAt/updatedAt/deletedAt)
- 实现 `TableName()`, `Generate()`, `GetId()` 方法
- GORM Tag: `gorm:"size:128;comment:注释"`
- JSON Tag: `json:"camelCase"`

#### 2.2.2 DTO (service/dto/sys_xxx.go)

```go
package dto

import (
    "go-admin/app/admin/models"
    "go-admin/common/dto"
    common "go-admin/common/models"
)

// 列表查询请求
type SysXxxGetPageReq struct {
    dto.Pagination `search:"-"`
    Field1         string `form:"field1" search:"type:contains;column:field1;table:sys_xxx"`
    Field2         string `form:"field2" search:"type:exact;column:field2;table:sys_xxx"`
}

func (m *SysXxxGetPageReq) GetNeedSearch() interface{} {
    return *m
}

// 增改请求
type SysXxxControl struct {
    Id     int    `uri:"Id" comment:"编码"`
    Field1 string `json:"field1" comment:"字段1"`
    Field2 string `json:"field2" comment:"字段2"`
    common.ControlBy
}

func (s *SysXxxControl) Generate(model *models.SysXxx) {
    if s.Id != 0 {
        model.Model = common.Model{Id: s.Id}
    }
    model.Field1 = s.Field1
    model.Field2 = s.Field2
}

func (s *SysXxxControl) GetId() interface{} {
    return s.Id
}

// 详情查询请求
type SysXxxGetReq struct {
    Id int `uri:"id"`
}

func (s *SysXxxGetReq) GetId() interface{} {
    return s.Id
}

// 删除请求
type SysXxxDeleteReq struct {
    Ids []int `json:"ids"`
    common.ControlBy
}

func (s *SysXxxDeleteReq) GetId() interface{} {
    return s.Ids
}
```

**关键特征：**
- 列表请求嵌入 `dto.Pagination` (pageIndex/pageSize)
- `GetNeedSearch()` 返回搜索条件
- `Generate()` 方法将 DTO 转换为 Model
- `GetId()` 返回主键
- 支持批量删除 (ids 数组)

#### 2.2.3 Service (service/sys_xxx.go)

```go
package service

import (
    "errors"
    "go-admin/app/admin/models"
    "go-admin/app/admin/service/dto"
    cDto "go-admin/common/dto"
    "github.com/go-admin-team/go-admin-core/sdk/service"
)

type SysXxx struct {
    service.Service
}

// GetPage 分页列表
func (e *SysXxx) GetPage(c *dto.SysXxxGetPageReq, list *[]models.SysXxx, count *int64) error {
    err := e.Orm.
        Scopes(
            cDto.MakeCondition(c.GetNeedSearch()),
            cDto.Paginate(c.GetPageSize(), c.GetPageIndex()),
        ).
        Find(list).Limit(-1).Offset(-1).
        Count(count).Error
    return err
}

// Get 详情
func (e *SysXxx) Get(d *dto.SysXxxGetReq, model *models.SysXxx) error {
    err := e.Orm.FirstOrInit(model, d.GetId()).Error
    if err != nil || model.Id == 0 {
        return errors.New("查看对象不存在或无权查看")
    }
    return nil
}

// Insert 新增
func (e *SysXxx) Insert(c *dto.SysXxxControl) error {
    var data models.SysXxx
    c.Generate(&data)
    return e.Orm.Create(&data).Error
}

// Update 修改
func (e *SysXxx) Update(c *dto.SysXxxControl) error {
    var model models.SysXxx
    e.Orm.First(&model, c.GetId())
    c.Generate(&model)
    db := e.Orm.Save(&model)
    if db.RowsAffected == 0 {
        return errors.New("无权更新该数据")
    }
    return db.Error
}

// Remove 删除
func (e *SysXxx) Remove(d *dto.SysXxxDeleteReq) error {
    var data models.SysXxx
    db := e.Orm.Delete(&data, d.Ids)
    if db.RowsAffected == 0 {
        return errors.New("无权删除该数据")
    }
    return db.Error
}
```

**关键特征：**
- 嵌入 `service.Service` (提供 Orm, Log 等)
- 使用 `e.Orm` 进行数据库操作
- 使用 GORM Scopes 实现条件查询和分页
- 返回明确的错误信息

#### 2.2.4 API Handler (apis/sys_xxx.go)

```go
package apis

import (
    "github.com/gin-gonic/gin"
    "github.com/gin-gonic/gin/binding"
    "github.com/go-admin-team/go-admin-core/sdk/api"
    "github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth/user"
    
    "go-admin/app/admin/models"
    "go-admin/app/admin/service"
    "go-admin/app/admin/service/dto"
)

type SysXxx struct {
    api.Api
}

// GetPage 列表
// @Summary 列表
// @Tags 模块名
// @Router /api/v1/sys-xxx [get]
func (e SysXxx) GetPage(c *gin.Context) {
    s := service.SysXxx{}
    req := dto.SysXxxGetPageReq{}
    err := e.MakeContext(c).MakeOrm().Bind(&req, binding.Form).MakeService(&s.Service).Errors
    if err != nil {
        e.Logger.Error(err)
        return
    }
    list := make([]models.SysXxx, 0)
    var count int64
    err = s.GetPage(&req, &list, &count)
    if err != nil {
        e.Error(500, err, "查询失败")
        return
    }
    e.PageOK(list, int(count), req.GetPageIndex(), req.GetPageSize(), "查询成功")
}

// Get 详情
// @Router /api/v1/sys-xxx/{id} [get]
func (e SysXxx) Get(c *gin.Context) {
    req := dto.SysXxxGetReq{}
    s := service.SysXxx{}
    err := e.MakeContext(c).MakeOrm().Bind(&req, binding.JSON, nil).MakeService(&s.Service).Errors
    if err != nil {
        e.Error(500, err, err.Error())
        return
    }
    var object models.SysXxx
    err = s.Get(&req, &object)
    if err != nil {
        e.Error(500, err, err.Error())
        return
    }
    e.OK(object, "查询成功")
}

// Insert 新增
// @Router /api/v1/sys-xxx [post]
func (e SysXxx) Insert(c *gin.Context) {
    s := service.SysXxx{}
    req := dto.SysXxxControl{}
    err := e.MakeContext(c).MakeOrm().Bind(&req, binding.JSON).MakeService(&s.Service).Errors
    if err != nil {
        e.Error(500, err, err.Error())
        return
    }
    req.SetCreateBy(user.GetUserId(c))
    err = s.Insert(&req)
    if err != nil {
        e.Error(500, err, "创建失败")
        return
    }
    e.OK(req.GetId(), "创建成功")
}

// Update 修改
// @Router /api/v1/sys-xxx/{id} [put]
func (e SysXxx) Update(c *gin.Context) {
    s := service.SysXxx{}
    req := dto.SysXxxControl{}
    err := e.MakeContext(c).MakeOrm().Bind(&req, binding.JSON, nil).MakeService(&s.Service).Errors
    if err != nil {
        e.Error(500, err, err.Error())
        return
    }
    req.SetUpdateBy(user.GetUserId(c))
    err = s.Update(&req)
    if err != nil {
        e.Error(500, err, "更新失败")
        return
    }
    e.OK(req.GetId(), "更新成功")
}

// Delete 删除
// @Router /api/v1/sys-xxx [delete]
func (e SysXxx) Delete(c *gin.Context) {
    s := service.SysXxx{}
    req := dto.SysXxxDeleteReq{}
    err := e.MakeContext(c).MakeOrm().Bind(&req, binding.JSON, nil).MakeService(&s.Service).Errors
    if err != nil {
        e.Error(500, err, err.Error())
        return
    }
    req.SetUpdateBy(user.GetUserId(c))
    err = s.Remove(&req)
    if err != nil {
        e.Error(500, err, "删除失败")
        return
    }
    e.OK(req.GetId(), "删除成功")
}
```

**关键特征：**
- 嵌入 `api.Api` (提供 MakeContext, MakeOrm 等)
- 使用链式调用: `e.MakeContext(c).MakeOrm().Bind(...).MakeService(...)`
- `binding.Form` 用于 GET 请求参数绑定
- `binding.JSON` 用于 POST/PUT/DELETE 请求体绑定
- 使用 `user.GetUserId(c)` 获取当前用户 ID
- 响应方法: `e.OK()`, `e.PageOK()`, `e.Error()`

#### 2.2.5 Router (router/sys_xxx.go)

```go
package router

import (
    "go-admin/app/admin/apis"
    "go-admin/common/middleware"
    "github.com/gin-gonic/gin"
    jwt "github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth"
)

func init() {
    routerCheckRole = append(routerCheckRole, registerSysXxxRouter)
}

func registerSysXxxRouter(v1 *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) {
    api := apis.SysXxx{}
    r := v1.Group("/sys-xxx").
        Use(authMiddleware.MiddlewareFunc()).
        Use(middleware.AuthCheckRole())
    {
        r.GET("", api.GetPage)       // 列表
        r.GET("/:id", api.Get)       // 详情
        r.POST("", api.Insert)       // 新增
        r.PUT("/:id", api.Update)    // 修改
        r.DELETE("", api.Delete)     // 删除
    }
}
```

**关键特征：**
- 使用 `init()` 自动注册路由
- 添加到 `routerCheckRole` 切片
- 使用 JWT 中间件: `authMiddleware.MiddlewareFunc()`
- 使用权限检查中间件: `middleware.AuthCheckRole()`
- API 前缀: `/api/v1/sys-xxx`

### 2.3 通用响应格式

```go
// 成功响应 (单对象)
{"code": 200, "data": {...}, "msg": "查询成功"}

// 成功响应 (分页列表)
{
  "code": 200,
  "data": {
    "list": [...],
    "count": 100,
    "pageIndex": 1,
    "pageSize": 10
  },
  "msg": "查询成功"
}

// 错误响应
{"code": 500, "data": null, "msg": "错误信息"}
```

---

## 3. 前端架构模式 (vue-vben-admin)

### 3.1 目录结构

```
apps/web-antd/src/
├── api/core/           # API 封装 (按模块)
│   ├── config.ts      # 参数配置 API
│   ├── user.ts        # 用户管理 API
│   └── ...
├── views/admin/       # 管理页面
│   ├── sys-config/    # 参数配置页面
│   ├── sys-user/      # 用户管理页面
│   └── ...
├── router/            # 路由配置
└── store/             # Pinia Store
```

### 3.2 API 封装模式

#### 3.2.1 API 文件 (api/core/xxx.ts)

```typescript
import { requestClient } from '#/api/request';

// 列表项接口
export interface XxxItem {
  id: number;
  field1: string;
  field2: string;
  createBy?: number;
  createdAt?: string;
}

// 分页响应接口
export interface XxxPageResult {
  list: XxxItem[];
  count: number;
  pageIndex: number;
  pageSize: number;
}

// 列表查询参数
interface GetXxxPageParams {
  pageIndex?: number;
  pageSize?: number;
  field1?: string;
  field2?: string;
}

/**
 * 获取分页列表
 * 对接 go-admin GET /api/v1/sys-xxx
 */
export async function getXxxPage(
  params: GetXxxPageParams = {},
): Promise<XxxPageResult> {
  return requestClient.get<XxxPageResult>('/v1/sys-xxx', { params });
}

/**
 * 获取详情
 * 对接 go-admin GET /api/v1/sys-xxx/:id
 */
export async function getXxxDetail(id: number): Promise<XxxItem> {
  return requestClient.get<XxxItem>(`/v1/sys-xxx/${id}`);
}

// 新增请求体
interface CreateXxxData {
  field1: string;
  field2?: string;
}

/**
 * 新增
 * 对接 go-admin POST /api/v1/sys-xxx
 */
export async function createXxx(data: CreateXxxData): Promise<number> {
  return requestClient.post('/v1/sys-xxx', data);
}

// 更新请求体
interface UpdateXxxData {
  field1?: string;
  field2?: string;
}

/**
 * 更新
 * 对接 go-admin PUT /api/v1/sys-xxx/:id
 */
export async function updateXxx(
  id: number,
  data: UpdateXxxData,
): Promise<number> {
  return requestClient.put(`/v1/sys-xxx/${id}`, data);
}

/**
 * 删除 (支持批量)
 * 对接 go-admin DELETE /api/v1/sys-xxx
 */
export async function deleteXxx(ids: number[]): Promise<number[]> {
  return requestClient.delete('/v1/sys-xxx', { data: { ids } });
}
```

**关键特征：**
- 使用 `requestClient` 发送请求 (封装自 @vben/request)
- 导出 Item 和 PageResult 类型
- 函数命名: `getXxxPage`, `getXxxDetail`, `createXxx`, `updateXxx`, `deleteXxx`
- 批量删除传 `{ ids: number[] }`
- 路径使用 `/v1/xxx` (baseURL 已配置 `/api`)

#### 3.2.2 API 导出 (api/core/index.ts)

```typescript
export * from './auth';
export * from './config';
export * from './user';
// ... 新增模块需在此导出
```

### 3.3 页面组件模式

#### 3.3.1 标准 CRUD 页面结构

```vue
<script lang="ts" setup>
/**
 * 系统管理 - XXX管理
 * 列表 + 搜索 + 分页 + 新增 + 编辑 + 删除
 */
import { onMounted, reactive, ref } from 'vue';
import {
  Button, Form, FormItem, Input, Modal, Table, message,
} from 'ant-design-vue';
import type { TableColumnType } from 'ant-design-vue';

import {
  createXxx, deleteXxx, getXxxDetail, getXxxPage, updateXxx,
} from '#/api/core';
import type { XxxItem, XxxPageResult } from '#/api/core';

// ========== 状态定义 ==========
const loading = ref(false);
const tableData = ref<XxxItem[]>([]);
const errorMsg = ref('');

// 分页
const pagination = ref({
  current: 1,
  pageSize: 10,
  total: 0,
  showSizeChanger: true,
  showTotal: (total: number) => `共 ${total} 条`,
});

// 搜索条件
const searchField1 = ref('');
const searchField2 = ref('');

// ========== 表格数据加载 ==========
async function fetchList() {
  loading.value = true;
  errorMsg.value = '';
  try {
    const params: {
      pageIndex: number;
      pageSize: number;
      field1?: string;
      field2?: string;
    } = {
      pageIndex: pagination.value.current,
      pageSize: pagination.value.pageSize,
    };
    if (searchField1.value.trim()) {
      params.field1 = searchField1.value.trim();
    }
    if (searchField2.value.trim()) {
      params.field2 = searchField2.value.trim();
    }
    const res: XxxPageResult = await getXxxPage(params);
    tableData.value = res.list || [];
    pagination.value.total = res.count || 0;
  } catch (e: unknown) {
    const err = e as { message?: string; response?: { data?: { msg?: string } } };
    errorMsg.value = err?.message || err?.response?.data?.msg || '加载失败';
    tableData.value = [];
    pagination.value.total = 0;
  } finally {
    loading.value = false;
  }
}

// ========== 搜索 ==========
function onSearch() {
  pagination.value.current = 1;
  fetchList();
}

function onReset() {
  searchField1.value = '';
  searchField2.value = '';
  pagination.value.current = 1;
  fetchList();
}

// ========== 分页变化 ==========
function onTableChange(
  pag: { current?: number; pageSize?: number },
  _filters: unknown,
  _sorter: unknown,
) {
  if (pag.current) pagination.value.current = pag.current;
  if (pag.pageSize) pagination.value.pageSize = pag.pageSize;
  fetchList();
}

// ========== 表格列定义 ==========
const columns: TableColumnType[] = [
  { title: 'ID', dataIndex: 'id', key: 'id', width: 80 },
  { title: '字段1', dataIndex: 'field1', key: 'field1', width: 140 },
  { title: '字段2', dataIndex: 'field2', key: 'field2', width: 140 },
  {
    title: '操作',
    key: 'action',
    width: 140,
    fixed: 'right',
  },
];

// ========== 新增 ==========
const addVisible = ref(false);
const addSubmitting = ref(false);
const addForm = reactive({
  field1: '',
  field2: '',
});

function resetAddForm() {
  addForm.field1 = '';
  addForm.field2 = '';
}

function openAddModal() {
  resetAddForm();
  addVisible.value = true;
}

function validateAddForm(): { ok: boolean; message?: string } {
  if (!addForm.field1?.trim()) {
    return { ok: false, message: '请输入字段1' };
  }
  return { ok: true };
}

async function onAddOk() {
  const v = validateAddForm();
  if (!v.ok) {
    message.error(v.message);
    return;
  }
  addSubmitting.value = true;
  try {
    await createXxx({
      field1: addForm.field1.trim(),
      field2: addForm.field2?.trim() ?? '',
    });
    message.success('新增成功');
    addVisible.value = false;
    fetchList();
  } catch (e: unknown) {
    const err = e as { message?: string; response?: { data?: { msg?: string } } };
    message.error(err?.message || err?.response?.data?.msg || '新增失败');
  } finally {
    addSubmitting.value = false;
  }
}

function onAddCancel() {
  addVisible.value = false;
}

// ========== 编辑 ==========
const editVisible = ref(false);
const editSubmitting = ref(false);
const editLoading = ref(false);
const editId = ref<number | null>(null);
const editForm = reactive({
  field1: '',
  field2: '',
});

async function openEditModal(record: XxxItem) {
  editId.value = record.id;
  editLoading.value = true;
  editVisible.value = true;
  try {
    const detail = await getXxxDetail(record.id);
    editForm.field1 = detail.field1 ?? '';
    editForm.field2 = detail.field2 ?? '';
  } catch (e: unknown) {
    message.error('获取详情失败');
    editVisible.value = false;
  } finally {
    editLoading.value = false;
  }
}

function validateEditForm(): { ok: boolean; message?: string } {
  if (!editForm.field1?.trim()) {
    return { ok: false, message: '请输入字段1' };
  }
  return { ok: true };
}

async function onEditOk() {
  if (editId.value === null) return;
  const v = validateEditForm();
  if (!v.ok) {
    message.error(v.message);
    return;
  }
  editSubmitting.value = true;
  try {
    await updateXxx(editId.value, {
      field1: editForm.field1.trim(),
      field2: editForm.field2?.trim() ?? '',
    });
    message.success('编辑成功');
    editVisible.value = false;
    fetchList();
  } catch (e: unknown) {
    message.error('编辑失败');
  } finally {
    editSubmitting.value = false;
  }
}

function onEditCancel() {
  editVisible.value = false;
}

// ========== 删除 ==========
function onDelete(record: XxxItem) {
  Modal.confirm({
    title: '确认删除',
    content: `确定要删除「${record.field1}」吗？`,
    okText: '确定',
    okType: 'danger',
    cancelText: '取消',
    async onOk() {
      try {
        await deleteXxx([record.id]);
        message.success('删除成功');
        fetchList();
      } catch (e: unknown) {
        message.error('删除失败');
      }
    },
  });
}

onMounted(() => {
  fetchList();
});
</script>

<template>
  <div class="p-4">
    <!-- 标题 -->
    <div class="mb-4 flex items-center justify-between">
      <h2 class="text-lg font-medium">XXX管理</h2>
      <Button type="primary" @click="openAddModal">新增</Button>
    </div>

    <!-- 搜索 -->
    <div class="mb-4 flex flex-wrap items-center gap-2">
      <Input v-model:value="searchField1" placeholder="搜索字段1" class="w-52" />
      <Button type="primary" @click="onSearch">查询</Button>
      <Button @click="onReset">重置</Button>
    </div>

    <!-- 错误提示 -->
    <div v-if="errorMsg" class="mb-4 text-red-600">{{ errorMsg }}</div>

    <!-- 表格 -->
    <Table
      :columns="columns"
      :data-source="tableData"
      :loading="loading"
      :pagination="pagination"
      :row-key="(record: XxxItem) => record.id"
      bordered
      @change="onTableChange"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'action'">
          <Button type="link" @click="openEditModal(record as XxxItem)">编辑</Button>
          <Button type="link" danger @click="onDelete(record as XxxItem)">删除</Button>
        </template>
      </template>
    </Table>

    <!-- 新增弹窗 -->
    <Modal v-model:open="addVisible" title="新增" @ok="onAddOk" @cancel="onAddCancel">
      <Form :label-col="{ span: 6 }" :wrapper-col="{ span: 16 }">
        <FormItem label="字段1" required>
          <Input v-model:value="addForm.field1" />
        </FormItem>
        <FormItem label="字段2">
          <Input v-model:value="addForm.field2" />
        </FormItem>
      </Form>
    </Modal>

    <!-- 编辑弹窗 -->
    <Modal v-model:open="editVisible" title="编辑" @ok="onEditOk" @cancel="onEditCancel">
      <div v-if="editLoading">加载中...</div>
      <Form v-else :label-col="{ span: 6 }" :wrapper-col="{ span: 16 }">
        <FormItem label="字段1" required>
          <Input v-model:value="editForm.field1" />
        </FormItem>
        <FormItem label="字段2">
          <Input v-model:value="editForm.field2" />
        </FormItem>
      </Form>
    </Modal>
  </div>
</template>
```

**关键特征：**
- 使用 `<script setup>` 语法
- 使用 `ant-design-vue` 组件
- 响应式数据使用 `ref()` 和 `reactive()`
- API 函数从 `#/api/core` 导入
- 错误处理使用统一的 try-catch 模式
- 使用 `message` 和 `Modal` 进行交互反馈

---

## 4. 前后端映射关系

### 4.1 API 路径映射

| 操作 | 前端 API 路径 | 后端 Router 路径 | HTTP 方法 |
|------|--------------|-----------------|-----------|
| 列表 | `/v1/sys-xxx` | `/api/v1/sys-xxx` | GET |
| 详情 | `/v1/sys-xxx/${id}` | `/api/v1/sys-xxx/:id` | GET |
| 新增 | `/v1/sys-xxx` | `/api/v1/sys-xxx` | POST |
| 修改 | `/v1/sys-xxx/${id}` | `/api/v1/sys-xxx/:id` | PUT |
| 删除 | `/v1/sys-xxx` | `/api/v1/sys-xxx` | DELETE |

### 4.2 字段命名映射

| 前端 | 后端 | 说明 |
|------|------|------|
| `id` | `Id` | 主键 |
| `fieldName` | `FieldName` | 字段名 (camelCase vs PascalCase) |
| `createdAt` | `CreatedAt` | 创建时间 |
| `updatedAt` | `UpdatedAt` | 更新时间 |
| `createBy` | `CreateBy` | 创建人 |
| `updateBy` | `UpdateBy` | 更新人 |

### 4.3 分页参数映射

| 前端 | 后端 | 说明 |
|------|------|------|
| `pageIndex` | `pageIndex` | 页码 (从1开始) |
| `pageSize` | `pageSize` | 每页条数 |
| `list` | `list` | 数据列表 |
| `count` | `count` | 总条数 |

---

## 5. 代码生成规则

### 5.1 命名规范

| 类型 | 规范 | 示例 |
|------|------|------|
| 模块名 | 小写 + 连字符 | `sys-config` |
| 后端 Model | PascalCase, Sys 前缀 | `SysConfig` |
| 后端 DTO | PascalCase + 后缀 | `SysConfigGetPageReq` |
| 后端 Service | PascalCase | `SysConfig` |
| 后端 API | PascalCase | `SysConfig` |
| 前端 API 文件 | camelCase | `config.ts` |
| 前端函数 | camelCase | `getConfigPage` |
| 前端页面 | 目录小写连字符 | `sys-config/index.vue` |
| 前端类型 | PascalCase | `ConfigItem` |

### 5.2 字段类型映射

| 数据库 | Go Model | TypeScript |
|--------|----------|------------|
| `int` | `int` | `number` |
| `varchar` | `string` | `string` |
| `text` | `string` | `string` |
| `datetime` | `time.Time` | `string` |
| `tinyint` | `string` | `string` |
| `decimal` | `float64` | `number` |

### 5.3 必需生成的文件清单

**后端 (5个文件):**
1. `app/admin/models/sys_xxx.go` - Model
2. `app/admin/service/dto/sys_xxx.go` - DTO
3. `app/admin/service/sys_xxx.go` - Service
4. `app/admin/apis/sys_xxx.go` - API Handler
5. `app/admin/router/sys_xxx.go` - Router

**前端 (2个文件):**
1. `apps/web-antd/src/api/core/xxx.ts` - API 封装
2. `apps/web-antd/src/views/admin/sys-xxx/index.vue` - 页面组件
3. `apps/web-antd/src/api/core/index.ts` - 更新导出 (追加)

---

## 6. 特殊处理说明

### 6.1 批量删除
- 前端: `deleteXxx(ids: number[])` 传 `{ data: { ids } }`
- 后端: `SysXxxDeleteReq` 包含 `Ids []int`

### 6.2 权限控制
- 后端路由使用 `AuthCheckRole()` 中间件
- 前端权限码固定返回 `['*:*:*']`
- 菜单权限由后端 `menurole` 接口控制

### 6.3 时间字段
- 后端: `time.Time` 类型，GORM 自动处理
- 前端: `string` 类型，直接显示

### 6.4 请求/响应处理
- 后端使用 `binding.Form` (GET) 或 `binding.JSON` (POST/PUT/DELETE)
- 前端使用 `requestClient`，baseURL 指向 `/api`
- 成功响应 code: 200

---

## 7. 文件生成依赖关系

```
生成顺序:
1. Model (models/sys_xxx.go)
   └── 依赖: common/models

2. DTO (service/dto/sys_xxx.go)
   └── 依赖: Model, common/dto, common/models

3. Service (service/sys_xxx.go)
   └── 依赖: Model, DTO, common/dto

4. API (apis/sys_xxx.go)
   └── 依赖: Model, Service, DTO, common middleware

5. Router (router/sys_xxx.go)
   └── 依赖: API

6. Frontend API (api/core/xxx.ts)
   └── 依赖: requestClient

7. Frontend Page (views/admin/sys-xxx/index.vue)
   └── 依赖: API, ant-design-vue
```

---

*分析完成时间: 基于当前代码库*
*适用项目: go-admin + vue-vben-admin 整合项目*
