<script lang="ts" setup>
import { computed, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import {
  Alert,
  Button,
  Card,
  Input,
  InputNumber,
  Select,
  Switch,
  Table,
  Tag,
  message,
} from 'ant-design-vue';
import type { TableColumnType } from 'ant-design-vue';
import type { SelectValue } from 'ant-design-vue/es/select';

import { getSysTableDetail, previewGeneratedCode, updateSysTable } from '#/api/core';
import type {
  PreviewResult,
  SysTableColumnItem,
  SysTableDetailResult,
  UpdateSysTablePayload,
} from '#/api/core';
import AdminFilterField from '#/components/admin/filter-field.vue';
import AdminPageShell from '#/components/admin/page-shell.vue';
import GenPreviewModal from '#/views/dev-tools/gen/components/preview-modal.vue';
import {
  ACTIONS_OPTIONS,
  AUTH_OPTIONS,
  BOOLEAN_STRING_OPTIONS,
  DATA_SCOPE_OPTIONS,
  HTML_TYPE_OPTIONS,
  LOGICAL_DELETE_OPTIONS,
  QUERY_TYPE_OPTIONS,
  TEMPLATE_CATEGORY_OPTIONS,
} from '#/views/dev-tools/gen/constants';

type SwitchCheckedValue = boolean | number | string;

const route = useRoute();
const router = useRouter();

const loading = ref(false);
const saving = ref(false);
const errorMsg = ref('');
const form = ref<UpdateSysTablePayload | null>(null);

const previewVisible = ref(false);
const previewLoading = ref(false);
const previewTitle = ref('');
const previewData = ref<PreviewResult>({});

const tableId = computed(() => {
  const raw = route.query.tableId;
  if (Array.isArray(raw)) return Number(raw[0] ?? 0);
  return Number(raw ?? 0);
});

const columns = computed<SysTableColumnItem[]>(() => form.value?.columns ?? []);

const fieldTableColumns: TableColumnType[] = [
  { title: '字段名', dataIndex: 'columnName', key: 'columnName', width: 160, fixed: 'left' },
  { title: '注释', dataIndex: 'columnComment', key: 'columnComment', width: 180 },
  { title: '类型', dataIndex: 'columnType', key: 'columnType', width: 140 },
  { title: 'Go类型', dataIndex: 'goType', key: 'goType', width: 140 },
  { title: 'Go字段', dataIndex: 'goField', key: 'goField', width: 160 },
  { title: 'JSON字段', dataIndex: 'jsonField', key: 'jsonField', width: 160 },
  { title: '主键', dataIndex: 'isPk', key: 'isPk', width: 100 },
  { title: '必填', dataIndex: 'isRequired', key: 'isRequired', width: 100 },
  { title: '新增', dataIndex: 'isInsert', key: 'isInsert', width: 100 },
  { title: '编辑', dataIndex: 'isEdit', key: 'isEdit', width: 100 },
  { title: '列表', dataIndex: 'isList', key: 'isList', width: 100 },
  { title: '查询', dataIndex: 'isQuery', key: 'isQuery', width: 100 },
  { title: '查询方式', dataIndex: 'queryType', key: 'queryType', width: 140 },
  { title: '表单类型', dataIndex: 'htmlType', key: 'htmlType', width: 150 },
  { title: '字典类型', dataIndex: 'dictType', key: 'dictType', width: 150 },
  { title: '排序', dataIndex: 'sort', key: 'sort', width: 100 },
  { title: '备注', dataIndex: 'remark', key: 'remark', width: 220 },
];

function toBooleanFlag(value?: string, fallback = false) {
  if (value === '1') return true;
  if (value === '0') return false;
  return fallback;
}

function normalizeColumn(column: SysTableColumnItem): SysTableColumnItem {
  return {
    ...column,
    columnComment: column.columnComment ?? '',
    columnName: column.columnName ?? '',
    columnType: column.columnType ?? '',
    dictType: column.dictType ?? '',
    goField: column.goField ?? '',
    goType: column.goType ?? 'string',
    htmlType: column.htmlType ?? 'input',
    jsonField: column.jsonField ?? '',
    queryType: column.queryType ?? 'EQ',
    remark: column.remark ?? '',
    sort: typeof column.sort === 'number' ? column.sort : 0,
    isPk: column.isPk ?? (column.pk ? '1' : '0'),
    isRequired: column.isRequired ?? (column.required ? '1' : '0'),
    isInsert: column.isInsert ?? (column.insert ? '1' : '0'),
    isEdit: column.isEdit ?? (column.edit ? '1' : '0'),
    isList: column.isList ?? '1',
    isQuery: column.isQuery ?? (column.query ? '1' : '0'),
    pk: column.pk ?? toBooleanFlag(column.isPk),
    required: column.required ?? toBooleanFlag(column.isRequired),
    insert: column.insert ?? toBooleanFlag(column.isInsert),
    edit: column.edit ?? toBooleanFlag(column.isEdit),
    query: column.query ?? toBooleanFlag(column.isQuery),
    increment: column.increment ?? toBooleanFlag(column.isIncrement),
    isIncrement: column.isIncrement ?? (column.increment ? '1' : '0'),
  };
}

function normalizePayload(detail: SysTableDetailResult): UpdateSysTablePayload {
  const info = detail.info ?? ({} as UpdateSysTablePayload);
  return {
    ...info,
    tableId: info.tableId,
    businessName: info.businessName ?? '',
    className: info.className ?? '',
    columns: (detail.list ?? []).map(normalizeColumn),
    crud: info.crud ?? true,
    functionAuthor: info.functionAuthor ?? '',
    functionName: info.functionName ?? '',
    isActions: typeof info.isActions === 'number' ? info.isActions : 2,
    isAuth: typeof info.isAuth === 'number' ? info.isAuth : 1,
    isDataScope: typeof info.isDataScope === 'number' ? info.isDataScope : 1,
    isLogicalDelete: info.isLogicalDelete ?? (info.logicalDelete ? '1' : '0'),
    logicalDelete: info.logicalDelete ?? toBooleanFlag(info.isLogicalDelete),
    logicalDeleteColumn: info.logicalDeleteColumn ?? 'is_del',
    moduleFrontName: info.moduleFrontName ?? '',
    moduleName: info.moduleName ?? '',
    options: info.options ?? '',
    packageName: info.packageName ?? 'admin',
    remark: info.remark ?? '',
    tableComment: info.tableComment ?? '',
    tableName: info.tableName ?? '',
    tplCategory: info.tplCategory ?? 'crud',
    tree: info.tree ?? false,
    treeCode: info.treeCode ?? '',
    treeName: info.treeName ?? '',
    treeParentCode: info.treeParentCode ?? '',
  };
}

function setColumnStringFlag(column: SysTableColumnItem, key: 'isPk' | 'isRequired' | 'isInsert' | 'isEdit' | 'isQuery', value: string) {
  column[key] = value;
  if (key === 'isPk') column.pk = value === '1';
  if (key === 'isRequired') column.required = value === '1';
  if (key === 'isInsert') column.insert = value === '1';
  if (key === 'isEdit') column.edit = value === '1';
  if (key === 'isQuery') column.query = value === '1';
}

function handleTreeChange(checked: SwitchCheckedValue) {
  if (!form.value) return;
  const nextChecked = checked === true;
  form.value.tree = nextChecked;
  form.value.tplCategory = nextChecked ? 'tree' : 'crud';
}

function handleLogicalDeleteChange(checked: SwitchCheckedValue) {
  if (!form.value) return;
  const nextChecked = checked === true;
  form.value.logicalDelete = nextChecked;
  form.value.isLogicalDelete = nextChecked ? '1' : '0';
  if (!nextChecked) {
    form.value.logicalDeleteColumn = '';
  } else if (!form.value.logicalDeleteColumn) {
    form.value.logicalDeleteColumn = 'is_del';
  }
}

function handleColumnFlagChange(
  column: SysTableColumnItem,
  key: 'isPk' | 'isRequired' | 'isInsert' | 'isEdit' | 'isQuery',
  value: SelectValue,
) {
  setColumnStringFlag(column, key, String(value ?? '0'));
}

function buildPayload(): UpdateSysTablePayload | null {
  if (!form.value) return null;
  return {
    ...form.value,
    columns: columns.value.map((column) => ({
      ...column,
      pk: column.isPk === '1',
      required: column.isRequired === '1',
      insert: column.isInsert === '1',
      edit: column.isEdit === '1',
      query: column.isQuery === '1',
      increment: column.isIncrement === '1',
    })),
    logicalDelete: form.value.isLogicalDelete === '1',
  };
}

function validateForm(): string | null {
  if (!form.value) return '未找到表配置';

  const root = form.value;
  const requiredFields: Array<[string, string | undefined]> = [
    ['表备注', root.tableComment],
    ['类名', root.className],
    ['包名', root.packageName],
    ['模块名', root.moduleName],
    ['业务名', root.businessName],
    ['功能名称', root.functionName],
  ];

  for (const [label, value] of requiredFields) {
    if (!String(value ?? '').trim()) return `请输入${label}`;
  }

  if (root.tree) {
    if (!String(root.treeCode ?? '').trim()) return '树表模式下必须填写树编码字段';
    if (!String(root.treeParentCode ?? '').trim()) return '树表模式下必须填写父级字段';
    if (!String(root.treeName ?? '').trim()) return '树表模式下必须填写树名称字段';
  }

  for (const column of root.columns) {
    if (!String(column.columnName ?? '').trim()) return '字段名不能为空';
    if (!String(column.goField ?? '').trim()) return `字段 ${column.columnName} 的 Go 字段不能为空`;
    if (!String(column.jsonField ?? '').trim()) return `字段 ${column.columnName} 的 JSON 字段不能为空`;
    if (!String(column.htmlType ?? '').trim()) return `字段 ${column.columnName} 的表单类型不能为空`;
    if (typeof column.sort !== 'number') return `字段 ${column.columnName} 的排序不能为空`;
  }

  return null;
}

async function loadDetail() {
  if (!tableId.value) {
    errorMsg.value = '缺少 tableId 参数';
    form.value = null;
    return;
  }
  loading.value = true;
  errorMsg.value = '';
  try {
    const detail = await getSysTableDetail(tableId.value);
    form.value = normalizePayload(detail);
  } catch (error: unknown) {
    const err = error as {
      message?: string;
      response?: { data?: { msg?: string } };
    };
    errorMsg.value = err?.message || err?.response?.data?.msg || '加载表配置失败';
    form.value = null;
  } finally {
    loading.value = false;
  }
}

async function saveConfig(openPreviewAfterSave = false) {
  const validationError = validateForm();
  if (validationError) {
    message.error(validationError);
    return;
  }
  const payload = buildPayload();
  if (!payload) return;

  saving.value = true;
  try {
    await updateSysTable(payload);
    message.success('保存成功');
    if (openPreviewAfterSave) {
      await openPreview();
    } else {
      await loadDetail();
    }
  } catch (error: unknown) {
    const err = error as {
      message?: string;
      response?: { data?: { msg?: string } };
    };
    message.error(err?.message || err?.response?.data?.msg || '保存失败');
  } finally {
    saving.value = false;
  }
}

async function openPreview() {
  if (!tableId.value) return;
  previewVisible.value = true;
  previewLoading.value = true;
  previewTitle.value = `${form.value?.tableComment || form.value?.tableName || '代码生成'} - 模板预览`;
  previewData.value = {};
  try {
    previewData.value = await previewGeneratedCode(tableId.value);
  } catch (error: unknown) {
    const err = error as {
      message?: string;
      response?: { data?: { msg?: string } };
    };
    message.error(err?.message || err?.response?.data?.msg || '获取预览失败');
    previewVisible.value = false;
  } finally {
    previewLoading.value = false;
  }
}

function goBack() {
  router.push('/dev-tools/gen');
}

watch(
  () => route.query.tableId,
  () => {
    loadDetail();
  },
  { immediate: true },
);
</script>

<template>
  <AdminPageShell header-mode="compact">
    <template #eyebrow>DEV TOOLS</template>
    <template #title>代码生成配置</template>
    <template #description>
      编辑表级配置和字段级配置，保存后可直接预览模板输出。本轮不开放真正的写项目/写菜单动作。
    </template>

    <template #header-extra>
      <Button @click="goBack">返回列表</Button>
      <Button :loading="saving" @click="saveConfig(false)">保存配置</Button>
      <Button type="primary" :loading="saving" @click="saveConfig(true)">
        保存并预览
      </Button>
    </template>

    <Alert
      v-if="errorMsg"
      show-icon
      type="error"
      :message="errorMsg"
      class="mb-4"
    />

    <div v-if="form" class="space-y-4">
      <Card :bordered="false" class="app-radius-panel shadow-sm">
        <template #title>
          <div class="flex items-center gap-2">
            <span>基础信息</span>
            <Tag color="blue">{{ form.tableName || '-' }}</Tag>
          </div>
        </template>
        <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          <AdminFilterField label="表名">
            <Input :value="form.tableName" disabled />
          </AdminFilterField>
          <AdminFilterField label="表备注">
            <Input v-model:value="form.tableComment" placeholder="请输入表备注" />
          </AdminFilterField>
          <AdminFilterField label="类名">
            <Input v-model:value="form.className" placeholder="请输入类名" />
          </AdminFilterField>
          <AdminFilterField label="包名">
            <Input v-model:value="form.packageName" placeholder="请输入包名" />
          </AdminFilterField>
          <AdminFilterField label="模块名">
            <Input v-model:value="form.moduleName" placeholder="请输入模块名" />
          </AdminFilterField>
          <AdminFilterField label="前端模块名">
            <Input
              v-model:value="form.moduleFrontName"
              placeholder="请输入前端模块名"
            />
          </AdminFilterField>
          <AdminFilterField label="业务名">
            <Input v-model:value="form.businessName" placeholder="请输入业务名" />
          </AdminFilterField>
          <AdminFilterField label="功能名称">
            <Input v-model:value="form.functionName" placeholder="请输入功能名称" />
          </AdminFilterField>
          <AdminFilterField label="作者">
            <Input
              v-model:value="form.functionAuthor"
              placeholder="请输入作者"
            />
          </AdminFilterField>
          <AdminFilterField label="模板类别">
            <Select
              v-model:value="form.tplCategory"
              :options="TEMPLATE_CATEGORY_OPTIONS"
            />
          </AdminFilterField>
          <AdminFilterField label="备注" :span="2">
            <Input
              v-model:value="form.remark"
              placeholder="请输入备注"
            />
          </AdminFilterField>
        </div>
      </Card>

      <Card :bordered="false" class="app-radius-panel shadow-sm">
        <template #title>生成策略</template>
        <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          <AdminFilterField label="启用 CRUD">
            <div class="flex h-10 items-center">
              <Switch v-model:checked="form.crud" />
            </div>
          </AdminFilterField>
          <AdminFilterField label="树表模式">
            <div class="flex h-10 items-center">
              <Switch :checked="form.tree" @update:checked="handleTreeChange" />
            </div>
          </AdminFilterField>
          <AdminFilterField label="数据权限">
            <Select
              v-model:value="form.isDataScope"
              :options="DATA_SCOPE_OPTIONS"
            />
          </AdminFilterField>
          <AdminFilterField label="动作模板">
            <Select
              v-model:value="form.isActions"
              :options="ACTIONS_OPTIONS"
            />
          </AdminFilterField>
          <AdminFilterField label="权限校验">
            <Select
              v-model:value="form.isAuth"
              :options="AUTH_OPTIONS"
            />
          </AdminFilterField>
          <AdminFilterField label="逻辑删除开关">
            <Select
              v-model:value="form.isLogicalDelete"
              :options="LOGICAL_DELETE_OPTIONS"
              @change="handleLogicalDeleteChange($event === '1')"
            />
          </AdminFilterField>
          <AdminFilterField label="逻辑删除布尔">
            <div class="flex h-10 items-center">
              <Switch
                :checked="form.logicalDelete"
                @update:checked="handleLogicalDeleteChange"
              />
            </div>
          </AdminFilterField>
          <AdminFilterField label="逻辑删除字段">
            <Input
              v-model:value="form.logicalDeleteColumn"
              placeholder="如 is_del"
            />
          </AdminFilterField>
        </div>
      </Card>

      <Card v-if="form.tree" :bordered="false" class="app-radius-panel shadow-sm">
        <template #title>树表配置</template>
        <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          <AdminFilterField label="树编码字段">
            <Input v-model:value="form.treeCode" placeholder="如 dept_id" />
          </AdminFilterField>
          <AdminFilterField label="父级字段">
            <Input
              v-model:value="form.treeParentCode"
              placeholder="如 parent_id"
            />
          </AdminFilterField>
          <AdminFilterField label="名称字段">
            <Input v-model:value="form.treeName" placeholder="如 dept_name" />
          </AdminFilterField>
        </div>
      </Card>

      <Card :bordered="false" class="app-radius-panel shadow-sm">
        <template #title>
          <div class="flex items-center justify-between gap-3">
            <span>字段配置</span>
            <span class="text-sm font-normal text-slate-500">
              当前共 {{ columns.length }} 个字段，支持行内编辑后整体保存
            </span>
          </div>
        </template>
        <Table
          row-key="columnId"
          :columns="fieldTableColumns"
          :data-source="columns"
          :loading="loading"
          :pagination="false"
          :scroll="{ x: 2600, y: 620 }"
          size="small"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'columnComment'">
              <Input
                v-model:value="(record as SysTableColumnItem).columnComment"
                size="small"
              />
            </template>
            <template v-else-if="column.key === 'goType'">
              <Input
                v-model:value="(record as SysTableColumnItem).goType"
                size="small"
              />
            </template>
            <template v-else-if="column.key === 'goField'">
              <Input
                v-model:value="(record as SysTableColumnItem).goField"
                size="small"
              />
            </template>
            <template v-else-if="column.key === 'jsonField'">
              <Input
                v-model:value="(record as SysTableColumnItem).jsonField"
                size="small"
              />
            </template>
            <template v-else-if="column.key === 'isPk'">
              <Select
                size="small"
                :value="(record as SysTableColumnItem).isPk"
                :options="BOOLEAN_STRING_OPTIONS"
                @change="handleColumnFlagChange(record as SysTableColumnItem, 'isPk', $event)"
              />
            </template>
            <template v-else-if="column.key === 'isRequired'">
              <Select
                size="small"
                :value="(record as SysTableColumnItem).isRequired"
                :options="BOOLEAN_STRING_OPTIONS"
                @change="handleColumnFlagChange(record as SysTableColumnItem, 'isRequired', $event)"
              />
            </template>
            <template v-else-if="column.key === 'isInsert'">
              <Select
                size="small"
                :value="(record as SysTableColumnItem).isInsert"
                :options="BOOLEAN_STRING_OPTIONS"
                @change="handleColumnFlagChange(record as SysTableColumnItem, 'isInsert', $event)"
              />
            </template>
            <template v-else-if="column.key === 'isEdit'">
              <Select
                size="small"
                :value="(record as SysTableColumnItem).isEdit"
                :options="BOOLEAN_STRING_OPTIONS"
                @change="handleColumnFlagChange(record as SysTableColumnItem, 'isEdit', $event)"
              />
            </template>
            <template v-else-if="column.key === 'isList'">
              <Select
                v-model:value="(record as SysTableColumnItem).isList"
                size="small"
                :options="BOOLEAN_STRING_OPTIONS"
              />
            </template>
            <template v-else-if="column.key === 'isQuery'">
              <Select
                size="small"
                :value="(record as SysTableColumnItem).isQuery"
                :options="BOOLEAN_STRING_OPTIONS"
                @change="handleColumnFlagChange(record as SysTableColumnItem, 'isQuery', $event)"
              />
            </template>
            <template v-else-if="column.key === 'queryType'">
              <Select
                v-model:value="(record as SysTableColumnItem).queryType"
                size="small"
                :options="QUERY_TYPE_OPTIONS"
              />
            </template>
            <template v-else-if="column.key === 'htmlType'">
              <Select
                v-model:value="(record as SysTableColumnItem).htmlType"
                size="small"
                :options="HTML_TYPE_OPTIONS"
              />
            </template>
            <template v-else-if="column.key === 'dictType'">
              <Input
                v-model:value="(record as SysTableColumnItem).dictType"
                size="small"
              />
            </template>
            <template v-else-if="column.key === 'sort'">
              <InputNumber
                v-model:value="(record as SysTableColumnItem).sort"
                size="small"
                class="w-full"
                :min="0"
              />
            </template>
            <template v-else-if="column.key === 'remark'">
              <Input
                v-model:value="(record as SysTableColumnItem).remark"
                size="small"
              />
            </template>
          </template>
        </Table>
      </Card>
    </div>

    <div v-else-if="loading" class="py-20 text-center text-slate-500">
      正在加载表配置...
    </div>

    <div v-else class="py-20 text-center text-slate-500">
      未找到可编辑的表配置
    </div>

    <GenPreviewModal
      v-model:open="previewVisible"
      :loading="previewLoading"
      :preview-data="previewData"
      :title="previewTitle"
    />
  </AdminPageShell>
</template>
type SwitchCheckedValue = boolean | string;
