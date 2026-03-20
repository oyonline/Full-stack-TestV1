# Full-stack Scaffold Skill

为 go-admin + vue-vben-admin 项目生成完整的前后端代码脚手架。

---

## 使用说明

### 基本用法

```
生成一个 [模块名] 模块，包含字段：[字段列表]
```

### 示例

```
生成一个商品管理模块，包含字段：商品名称(name)、商品编码(code)、价格(price)、库存(stock)、状态(status)、分类ID(categoryId)
```

---

## 生成规则

### 1. 命名转换

| 输入 | 模块名 | 后端 Model | 前端 API 文件 |
|------|--------|-----------|--------------|
| 商品管理 | Product | SysProduct | product.ts |
| 订单管理 | Order | SysOrder | order.ts |
| 文章分类 | ArticleCategory | SysArticleCategory | article-category.ts |

### 2. 字段类型推断

| 字段特征 | Go 类型 | TS 类型 | 前端组件 |
|---------|---------|---------|---------|
| ID/分类ID/用户ID | int | number | InputNumber |
| 名称/标题/内容 | string | string | Input |
| 描述/备注/正文 | string | string | Textarea |
| 价格/金额 | float64 | number | InputNumber |
| 状态/类型/是否 | string | string | Select |
| 排序/数量 | int | number | InputNumber |
| 时间/日期 | time.Time | string | DatePicker |

### 3. 默认字段（自动生成）

无需在输入中指定，自动包含：
- `id` - 主键
- `createBy` - 创建人
- `updateBy` - 更新人
- `createdAt` - 创建时间
- `updatedAt` - 更新时间

---

## 输出文件

### 后端文件 (go-admin/)

```
app/admin/
├── models/sys_{{table_name}}.go          # GORM 模型
├── service/dto/sys_{{table_name}}.go     # 请求/响应 DTO
├── service/sys_{{table_name}}.go         # 业务逻辑 Service
├── apis/sys_{{table_name}}.go            # API Handler
└── router/sys_{{table_name}}.go          # 路由配置
```

### 前端文件 (vue-vben-admin/)

```
apps/web-antd/src/
├── api/core/{{api_file}}.ts                      # API 封装
├── views/admin/sys-{{table_name}}/index.vue      # 管理页面
└── api/core/index.ts (更新导出)                   # 追加导出
```

---

## 代码模板

### 后端 Model

```go
package models

import "go-admin/common/models"

type Sys{{PascalName}} struct {
    models.Model
{{ModelFields}}
    models.ControlBy
    models.ModelTime
}

func (*Sys{{PascalName}}) TableName() string {
    return "sys_{{SnakeName}}"
}

func (e *Sys{{PascalName}}) Generate() models.ActiveRecord {
    o := *e
    return &o
}

func (e *Sys{{PascalName}}) GetId() interface{} {
    return e.Id
}
```

**ModelFields 模板：**
```
    {{FieldPascal}} {{GoType}} `json:"{{FieldCamel}}" gorm:"{{GormTag}}"`
```

### 后端 DTO

```go
package dto

import (
    "go-admin/app/admin/models"
    "go-admin/common/dto"
    common "go-admin/common/models"
)

type Sys{{PascalName}}GetPageReq struct {
    dto.Pagination `search:"-"`
{{SearchFields}}
}

func (m *Sys{{PascalName}}GetPageReq) GetNeedSearch() interface{} {
    return *m
}

type Sys{{PascalName}}Control struct {
    Id int `uri:"Id" comment:"编码"`
{{ControlFields}}
    common.ControlBy
}

func (s *Sys{{PascalName}}Control) Generate(model *models.Sys{{PascalName}}) {
    if s.Id != 0 {
        model.Model = common.Model{Id: s.Id}
    }
{{GenerateFields}}
}

func (s *Sys{{PascalName}}Control) GetId() interface{} {
    return s.Id
}

type Sys{{PascalName}}GetReq struct {
    Id int `uri:"id"`
}

func (s *Sys{{PascalName}}GetReq) GetId() interface{} {
    return s.Id
}

type Sys{{PascalName}}DeleteReq struct {
    Ids []int `json:"ids"`
    common.ControlBy
}

func (s *Sys{{PascalName}}DeleteReq) GetId() interface{} {
    return s.Ids
}
```

**SearchFields 模板：**
```go
    {{FieldPascal}} {{GoType}} `form:"{{FieldCamel}}" search:"type:contains;column:{{FieldSnake}};table:sys_{{SnakeName}}"`
```

**ControlFields 模板：**
```go
    {{FieldPascal}} {{GoType}} `json:"{{FieldCamel}}" comment:"{{FieldComment}}"`
```

**GenerateFields 模板：**
```go
    model.{{FieldPascal}} = s.{{FieldPascal}}
```

### 后端 Service

```go
package service

import (
    "errors"
    "go-admin/app/admin/models"
    "go-admin/app/admin/service/dto"
    cDto "go-admin/common/dto"
    "github.com/go-admin-team/go-admin-core/sdk/service"
)

type Sys{{PascalName}} struct {
    service.Service
}

func (e *Sys{{PascalName}}) GetPage(c *dto.Sys{{PascalName}}GetPageReq, list *[]models.Sys{{PascalName}}, count *int64) error {
    err := e.Orm.
        Scopes(
            cDto.MakeCondition(c.GetNeedSearch()),
            cDto.Paginate(c.GetPageSize(), c.GetPageIndex()),
        ).
        Find(list).Limit(-1).Offset(-1).
        Count(count).Error
    if err != nil {
        e.Log.Errorf("Service GetPage error: %s", err)
        return err
    }
    return nil
}

func (e *Sys{{PascalName}}) Get(d *dto.Sys{{PascalName}}GetReq, model *models.Sys{{PascalName}}) error {
    err := e.Orm.FirstOrInit(model, d.GetId()).Error
    if err != nil {
        e.Log.Errorf("db error: %s", err)
        _ = e.AddError(err)
        return err
    }
    if model.Id == 0 {
        err = errors.New("查看对象不存在或无权查看")
        e.Log.Errorf("Service Get error: %s", err)
        _ = e.AddError(err)
        return err
    }
    return nil
}

func (e *Sys{{PascalName}}) Insert(c *dto.Sys{{PascalName}}Control) error {
    var err error
    var data models.Sys{{PascalName}}
    c.Generate(&data)
    err = e.Orm.Create(&data).Error
    if err != nil {
        e.Log.Errorf("Service Insert error: %s", err)
        return err
    }
    return nil
}

func (e *Sys{{PascalName}}) Update(c *dto.Sys{{PascalName}}Control) error {
    var err error
    var model = models.Sys{{PascalName}}{}
    e.Orm.First(&model, c.GetId())
    c.Generate(&model)
    db := e.Orm.Save(&model)
    err = db.Error
    if err != nil {
        e.Log.Errorf("Service Update error: %s", err)
        return err
    }
    if db.RowsAffected == 0 {
        return errors.New("无权更新该数据")
    }
    return nil
}

func (e *Sys{{PascalName}}) Remove(d *dto.Sys{{PascalName}}DeleteReq) error {
    var err error
    var data models.Sys{{PascalName}}
    db := e.Orm.Delete(&data, d.Ids)
    if err = db.Error; err != nil {
        e.Log.Errorf("Service Remove error: %s", err)
        return err
    }
    if db.RowsAffected == 0 {
        err = errors.New("无权删除该数据")
        return err
    }
    return nil
}
```

### 后端 API

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

type Sys{{PascalName}} struct {
    api.Api
}

// GetPage 列表
// @Summary {{ModuleDesc}}列表
// @Description {{ModuleDesc}}列表
// @Tags {{ModuleDesc}}
{{SwaggerParams}}
// @Success 200 {object} response.Response{data=response.Page{list=[]models.Sys{{PascalName}}}} "{"code": 200, "data": [...]}"
// @Router /api/v1/sys-{{KebabName}} [get]
// @Security Bearer
func (e Sys{{PascalName}}) GetPage(c *gin.Context) {
    s := service.Sys{{PascalName}}{}
    req := dto.Sys{{PascalName}}GetPageReq{}
    err := e.MakeContext(c).MakeOrm().Bind(&req, binding.Form).MakeService(&s.Service).Errors
    if err != nil {
        e.Logger.Error(err)
        return
    }
    list := make([]models.Sys{{PascalName}}, 0)
    var count int64
    err = s.GetPage(&req, &list, &count)
    if err != nil {
        e.Error(500, err, "查询失败")
        return
    }
    e.PageOK(list, int(count), req.GetPageIndex(), req.GetPageSize(), "查询成功")
}

// Get 详情
// @Summary {{ModuleDesc}}详情
// @Description {{ModuleDesc}}详情
// @Tags {{ModuleDesc}}
// @Param id path string false "id"
// @Success 200 {object} response.Response{data=models.Sys{{PascalName}}} "{"code": 200, "data": [...]}"
// @Router /api/v1/sys-{{KebabName}}/{id} [get]
// @Security Bearer
func (e Sys{{PascalName}}) Get(c *gin.Context) {
    req := dto.Sys{{PascalName}}GetReq{}
    s := service.Sys{{PascalName}}{}
    err := e.MakeContext(c).MakeOrm().Bind(&req, binding.JSON, nil).MakeService(&s.Service).Errors
    if err != nil {
        e.Logger.Error(err)
        e.Error(500, err, err.Error())
        return
    }
    var object models.Sys{{PascalName}}
    err = s.Get(&req, &object)
    if err != nil {
        e.Error(500, err, err.Error())
        return
    }
    e.OK(object, "查询成功")
}

// Insert 新增
// @Summary 创建{{ModuleDesc}}
// @Description 创建{{ModuleDesc}}
// @Tags {{ModuleDesc}}
// @Accept application/json
// @Product application/json
// @Param data body dto.Sys{{PascalName}}Control true "body"
// @Success 200 {object} response.Response "{"code": 200, "message": "创建成功"}"
// @Router /api/v1/sys-{{KebabName}} [post]
// @Security Bearer
func (e Sys{{PascalName}}) Insert(c *gin.Context) {
    s := service.Sys{{PascalName}}{}
    req := dto.Sys{{PascalName}}Control{}
    err := e.MakeContext(c).MakeOrm().Bind(&req, binding.JSON).MakeService(&s.Service).Errors
    if err != nil {
        e.Logger.Error(err)
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
// @Summary 修改{{ModuleDesc}}
// @Description 修改{{ModuleDesc}}
// @Tags {{ModuleDesc}}
// @Accept application/json
// @Product application/json
// @Param data body dto.Sys{{PascalName}}Control true "body"
// @Success 200 {object} response.Response "{"code": 200, "message": "修改成功"}"
// @Router /api/v1/sys-{{KebabName}}/{id} [put]
// @Security Bearer
func (e Sys{{PascalName}}) Update(c *gin.Context) {
    s := service.Sys{{PascalName}}{}
    req := dto.Sys{{PascalName}}Control{}
    err := e.MakeContext(c).MakeOrm().Bind(&req, binding.JSON, nil).MakeService(&s.Service).Errors
    if err != nil {
        e.Logger.Error(err)
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
// @Summary 删除{{ModuleDesc}}
// @Description 删除{{ModuleDesc}}
// @Tags {{ModuleDesc}}
// @Param ids body []int false "ids"
// @Success 200 {object} response.Response "{"code": 200, "message": "删除成功"}"
// @Router /api/v1/sys-{{KebabName}} [delete]
// @Security Bearer
func (e Sys{{PascalName}}) Delete(c *gin.Context) {
    s := service.Sys{{PascalName}}{}
    req := dto.Sys{{PascalName}}DeleteReq{}
    err := e.MakeContext(c).MakeOrm().Bind(&req, binding.JSON, nil).MakeService(&s.Service).Errors
    if err != nil {
        e.Logger.Error(err)
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

### 后端 Router

```go
package router

import (
    "go-admin/app/admin/apis"
    "go-admin/common/middleware"
    "github.com/gin-gonic/gin"
    jwt "github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth"
)

func init() {
    routerCheckRole = append(routerCheckRole, registerSys{{PascalName}}Router)
}

func registerSys{{PascalName}}Router(v1 *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) {
    api := apis.Sys{{PascalName}}{}
    r := v1.Group("/sys-{{KebabName}}").
        Use(authMiddleware.MiddlewareFunc()).
        Use(middleware.AuthCheckRole())
    {
        r.GET("", api.GetPage)
        r.GET("/:id", api.Get)
        r.POST("", api.Insert)
        r.PUT("/:id", api.Update)
        r.DELETE("", api.Delete)
    }
}
```

### 前端 API

```typescript
import { requestClient } from '#/api/request';

/** {{ModuleDesc}}列表项 */
export interface {{PascalName}}Item {
  id: number;
{{ItemFields}}
  createBy?: number;
  updateBy?: number;
  createdAt?: string;
  updatedAt?: string;
}

/** {{ModuleDesc}}分页响应 */
export interface {{PascalName}}PageResult {
  list: {{PascalName}}Item[];
  count: number;
  pageIndex: number;
  pageSize: number;
}

/** 列表查询参数 */
interface Get{{PascalName}}PageParams {
  pageIndex?: number;
  pageSize?: number;
{{ParamFields}}
}

/**
 * 获取分页列表
 * 对接 go-admin GET /api/v1/sys-{{KebabName}}
 */
export async function get{{PascalName}}Page(
  params: Get{{PascalName}}PageParams = {},
): Promise<{{PascalName}}PageResult> {
  return requestClient.get<{{PascalName}}PageResult>('/v1/sys-{{KebabName}}', { params });
}

/**
 * 获取详情
 * 对接 go-admin GET /api/v1/sys-{{KebabName}}/:id
 */
export async function get{{PascalName}}Detail(id: number): Promise<{{PascalName}}Item> {
  return requestClient.get<{{PascalName}}Item>(`/v1/sys-{{KebabName}}/${id}`);
}

/** 新增请求体 */
interface Create{{PascalName}}Data {
{{CreateFields}}
}

/**
 * 新增
 * 对接 go-admin POST /api/v1/sys-{{KebabName}}
 */
export async function create{{PascalName}}(data: Create{{PascalName}}Data): Promise<number> {
  return requestClient.post('/v1/sys-{{KebabName}}', data);
}

/** 更新请求体 */
interface Update{{PascalName}}Data {
{{UpdateFields}}
}

/**
 * 更新
 * 对接 go-admin PUT /api/v1/sys-{{KebabName}}/:id
 */
export async function update{{PascalName}}(
  id: number,
  data: Update{{PascalName}}Data,
): Promise<number> {
  return requestClient.put(`/v1/sys-{{KebabName}}/${id}`, data);
}

/**
 * 删除（支持批量）
 * 对接 go-admin DELETE /api/v1/sys-{{KebabName}}
 */
export async function delete{{PascalName}}(ids: number[]): Promise<number[]> {
  return requestClient.delete('/v1/sys-{{KebabName}}', { data: { ids } });
}
```

### 前端页面

```vue
<script lang="ts" setup>
/**
 * 系统管理 - {{ModuleDesc}}
 * 列表 + 搜索 + 分页 + 新增 + 编辑 + 删除
 */
import { onMounted, reactive, ref } from 'vue';
import {
  Button,
  Form,
  FormItem,
  Input,
  InputNumber,
  Modal,
  Select,
  Table,
  Textarea,
  message,
} from 'ant-design-vue';
import type { TableColumnType } from 'ant-design-vue';

import {
  create{{PascalName}},
  delete{{PascalName}},
  get{{PascalName}}Detail,
  get{{PascalName}}Page,
  update{{PascalName}},
} from '#/api/core';
import type { {{PascalName}}Item, {{PascalName}}PageResult } from '#/api/core';

const loading = ref(false);
const tableData = ref<{{PascalName}}Item[]>([]);
const errorMsg = ref('');

const pagination = ref({
  current: 1,
  pageSize: 10,
  total: 0,
  showSizeChanger: true,
  showTotal: (total: number) => `共 ${total} 条`,
});

{{SearchState}}

async function fetchList() {
  loading.value = true;
  errorMsg.value = '';
  try {
    const params: {
      pageIndex: number;
      pageSize: number;
{{ParamTypes}}
    } = {
      pageIndex: pagination.value.current,
      pageSize: pagination.value.pageSize,
    };
{{ParamAssignments}}
    const res: {{PascalName}}PageResult = await get{{PascalName}}Page(params);
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

function onSearch() {
  pagination.value.current = 1;
  fetchList();
}

function onReset() {
{{ResetFields}}
  pagination.value.current = 1;
  fetchList();
}

function onTableChange(
  pag: { current?: number; pageSize?: number },
  _filters: unknown,
  _sorter: unknown,
) {
  if (pag.current) pagination.value.current = pag.current;
  if (pag.pageSize) pagination.value.pageSize = pag.pageSize;
  fetchList();
}

{{RenderHelpers}}

function renderEmpty(value: string | null | undefined): string {
  return value ?? '-';
}

const columns: TableColumnType[] = [
  { title: 'ID', dataIndex: 'id', key: 'id', width: 80 },
{{ColumnDefs}}
  { title: '创建时间', dataIndex: 'createdAt', key: 'createdAt', width: 160 },
  { title: '操作', key: 'action', width: 140, fixed: 'right' },
];

const addVisible = ref(false);
const addSubmitting = ref(false);
const addForm = reactive({
{{AddFormFields}}
});

function resetAddForm() {
{{ResetAddForm}}
}

function openAddModal() {
  resetAddForm();
  addVisible.value = true;
}

function validateAddForm(): { ok: boolean; message?: string } {
{{ValidateAddForm}}
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
    await create{{PascalName}}({
{{CreateFormData}}
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

const editVisible = ref(false);
const editSubmitting = ref(false);
const editLoading = ref(false);
const editId = ref<number | null>(null);
const editForm = reactive({
{{EditFormFields}}
});

async function openEditModal(record: {{PascalName}}Item) {
  editId.value = record.id;
  editLoading.value = true;
  editVisible.value = true;
  try {
    const detail = await get{{PascalName}}Detail(record.id);
{{EditFormAssignments}}
  } catch (e: unknown) {
    message.error('获取详情失败');
    editVisible.value = false;
  } finally {
    editLoading.value = false;
  }
}

function validateEditForm(): { ok: boolean; message?: string } {
{{ValidateEditForm}}
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
    await update{{PascalName}}(editId.value, {
{{UpdateFormData}}
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

function onDelete(record: {{PascalName}}Item) {
  Modal.confirm({
    title: '确认删除',
    content: `确定要删除吗？删除后不可恢复。`,
    okText: '确定',
    okType: 'danger',
    cancelText: '取消',
    async onOk() {
      try {
        await delete{{PascalName}}([record.id]);
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
    <div class="mb-4 flex items-center justify-between">
      <h2 class="text-lg font-medium">{{ModuleDesc}}</h2>
      <div class="flex gap-2">
        <Button @click="fetchList">刷新</Button>
        <Button type="primary" @click="openAddModal">新增</Button>
      </div>
    </div>

    <div class="mb-4 flex flex-wrap items-center gap-2">
{{SearchTemplate}}
      <Button type="primary" size="small" @click="onSearch">查询</Button>
      <Button size="small" @click="onReset">重置</Button>
    </div>

    <div v-if="errorMsg" class="mb-4 text-red-600">{{ errorMsg }}</div>

    <Table
      :columns="columns"
      :data-source="tableData"
      :loading="loading"
      :pagination="pagination"
      :row-key="(record: {{PascalName}}Item) => record.id"
      size="small"
      bordered
      @change="onTableChange"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'action'">
          <Button type="link" size="small" @click="openEditModal(record as {{PascalName}}Item)">编辑</Button>
          <Button type="link" size="small" danger @click="onDelete(record as {{PascalName}}Item)">删除</Button>
        </template>
      </template>
    </Table>

    <Modal v-model:open="addVisible" title="新增{{ModuleDesc}}" :confirm-loading="addSubmitting" @ok="onAddOk" @cancel="onAddCancel">
      <Form :label-col="{ span: 6 }" :wrapper-col="{ span: 16 }" class="mt-4">
{{AddFormTemplate}}
      </Form>
    </Modal>

    <Modal v-model:open="editVisible" title="编辑{{ModuleDesc}}" :confirm-loading="editSubmitting" @ok="onEditOk" @cancel="onEditCancel">
      <div v-if="editLoading" class="py-8 text-center text-gray-400">加载中...</div>
      <Form v-else :label-col="{ span: 6 }" :wrapper-col="{ span: 16 }" class="mt-4">
{{EditFormTemplate}}
      </Form>
    </Modal>
  </div>
</template>
```

---

## 模板变量替换规则

### 命名变量

| 变量 | 示例输入 | 输出 | 用途 |
|------|---------|------|------|
| `{{PascalName}}` | 商品管理 | Product | 类名、接口名 |
| `{{CamelName}}` | 商品管理 | product | 函数名、文件名 |
| `{{KebabName}}` | 商品管理 | product | URL、目录名 |
| `{{SnakeName}}` | 商品管理 | product | 数据库表名 |
| `{{ModuleDesc}}` | 商品管理 | 商品管理 | 中文描述 |

### 字段变量

| 变量 | 说明 |
|------|------|
| `{{ModelFields}}` | Go Model 字段定义列表 |
| `{{SearchFields}}` | DTO 搜索字段定义 |
| `{{ControlFields}}` | DTO 控制字段定义 |
| `{{GenerateFields}}` | Generate 方法字段映射 |
| `{{ItemFields}}` | TypeScript Item 接口字段 |
| `{{ParamFields}}` | API 查询参数字段 |
| `{{ColumnDefs}}` | 表格列定义 |
| `{{AddFormFields}}` | 新增表单字段 |
| `{{EditFormFields}}` | 编辑表单字段 |
| `{{AddFormTemplate}}` | 新增表单模板 |
| `{{EditFormTemplate}}` | 编辑表单模板 |

---

## 字段类型映射表

### Go <-> TypeScript <-> 组件

| 数据类型 | Go | TypeScript | 前端组件 |
|---------|-----|-----------|---------|
| 字符串(短) | string | string | Input |
| 字符串(长) | string | string | Textarea |
| 整数 | int | number | InputNumber |
| 小数 | float64 | number | InputNumber |
| 布尔 | string | string | Select(是/否) |
| 日期时间 | time.Time | string | DatePicker |
| 枚举 | string | string | Select |

### GORM Tag 规则

| 类型 | Tag 示例 |
|------|---------|
| 普通字符串 | `gorm:"size:128;comment:字段名"` |
| 长文本 | `gorm:"type:text;comment:内容"` |
| 状态码 | `gorm:"size:4;comment:状态"` |
| 整数 | `gorm:"comment:排序"` |

### 表单组件规则

| 字段特征 | 组件 | 属性 |
|---------|------|------|
| 包含"内容"/"描述"/"正文" | Textarea | `:rows="4"` |
| 包含"排序"/"数量"/"价格" | InputNumber | `class="w-full"` |
| 包含"状态"/"类型"/"是否" | Select | `:options="xxxOptions"` |
| 其他 | Input | `allow-clear` |

---

## 生成步骤

1. **解析输入** - 提取模块名和字段列表
2. **命名转换** - 生成 Pascal/Camel/Kebab/Snake 各种形式
3. **类型推断** - 根据字段名推断数据类型
4. **生成后端** - 依次生成 Model → DTO → Service → API → Router
5. **生成前端** - 依次生成 API → Page
6. **更新导出** - 在 `api/core/index.ts` 追加导出语句

---

## 注意事项

### 后端
1. 所有结构体必须实现 `Generate()` 和 `GetId()` 方法
2. 路由使用 `init()` 自动注册到 `routerCheckRole`
3. 使用 `binding.Form` 绑定 GET 参数，`binding.JSON` 绑定 Body
4. 删除操作接收 `ids` 数组支持批量删除

### 前端
1. 使用 `<Textarea>` 组件而非 `Input type="textarea"`
2. 搜索条件下拉需要包含 `{ value: '', label: '全部' }`
3. 编辑弹窗需要 `editLoading` 状态处理详情加载
4. 所有表单字段使用 `?.trim() ?? ''` 处理空值

### 数据库
1. 表名使用 `sys_xxx` 前缀
2. 需要手动执行数据库迁移创建表
3. 字段名使用蛇形命名(snake_case)

---

## 生成后待办

- [ ] 数据库建表 SQL
- [ ] 在 go-admin 中注册菜单
- [ ] 配置角色权限
- [ ] 联调测试
