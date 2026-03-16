<script lang="ts" setup>
/**
 * 系统管理 - 参数配置
 * 列表 + 搜索 + 分页 + 新增 + 编辑 + 删除
 */
import { onMounted, reactive, ref } from 'vue';

import {
  Button,
  Form,
  FormItem,
  Input,
  Modal,
  Select,
  Table,
  message,
} from 'ant-design-vue';
import type { TableColumnType } from 'ant-design-vue';

import {
  createConfig,
  deleteConfig,
  getConfigDetail,
  getConfigPage,
  updateConfig,
} from '#/api/core';
import type { SysConfigItem, SysConfigPageResult } from '#/api/core';

/** 表格加载状态 */
const loading = ref(false);
/** 表格数据 */
const tableData = ref<SysConfigItem[]>([]);
/** 错误提示 */
const errorMsg = ref('');

/** 分页状态 */
const pagination = ref({
  current: 1,
  pageSize: 10,
  total: 0,
  showSizeChanger: true,
  showTotal: (total: number) => `共 ${total} 条`,
});

/** 搜索：参数名称（模糊） */
const searchConfigName = ref('');
/** 搜索：参数键名（模糊） */
const searchConfigKey = ref('');
/** 搜索：是否前台（0=否 1=是） */
const searchIsFrontend = ref<string>('');

/** 是否前台下拉选项 */
const isFrontendOptions = [
  { value: '', label: '全部' },
  { value: '0', label: '否' },
  { value: '1', label: '是' },
];

/** 获取参数配置列表 */
async function fetchConfigList() {
  loading.value = true;
  errorMsg.value = '';
  try {
    const params: {
      pageIndex: number;
      pageSize: number;
      configName?: string;
      configKey?: string;
      isFrontend?: string;
    } = {
      pageIndex: pagination.value.current,
      pageSize: pagination.value.pageSize,
    };
    if (searchConfigName.value.trim()) {
      params.configName = searchConfigName.value.trim();
    }
    if (searchConfigKey.value.trim()) {
      params.configKey = searchConfigKey.value.trim();
    }
    if (searchIsFrontend.value !== '') {
      params.isFrontend = searchIsFrontend.value;
    }
    const res: SysConfigPageResult = await getConfigPage(params);
    tableData.value = res.list || [];
    pagination.value.total = res.count || 0;
  } catch (e: unknown) {
    const err = e as { message?: string; response?: { data?: { msg?: string } } };
    errorMsg.value =
      err?.message || err?.response?.data?.msg || '加载参数配置列表失败';
    tableData.value = [];
    pagination.value.total = 0;
  } finally {
    loading.value = false;
  }
}

/** 查询按钮 */
function onSearch() {
  pagination.value.current = 1;
  fetchConfigList();
}

/** 重置按钮 */
function onReset() {
  searchConfigName.value = '';
  searchConfigKey.value = '';
  searchIsFrontend.value = '';
  pagination.value.current = 1;
  fetchConfigList();
}

/** 分页变化 */
function onTableChange(
  pag: { current?: number; pageSize?: number },
  _filters: unknown,
  _sorter: unknown,
) {
  if (pag.current) pagination.value.current = pag.current;
  if (pag.pageSize) pagination.value.pageSize = pag.pageSize;
  fetchConfigList();
}

/** 是否前台渲染 */
function renderIsFrontend(val: string): string {
  if (val === '1') return '是';
  if (val === '0') return '否';
  return val || '-';
}

/** 空值渲染 */
function renderEmpty(value: string | null | undefined): string {
  return value ?? '-';
}

/** 表格列定义 */
const columns: TableColumnType[] = [
  { title: 'ID', dataIndex: 'id', key: 'id', width: 80 },
  { title: '参数名称', dataIndex: 'configName', key: 'configName', width: 140 },
  { title: '参数键名', dataIndex: 'configKey', key: 'configKey', width: 140 },
  {
    title: '参数值',
    dataIndex: 'configValue',
    key: 'configValue',
    ellipsis: true,
    customRender: ({ text }: { text: string }) => renderEmpty(text),
  },
  {
    title: '配置类型',
    dataIndex: 'configType',
    key: 'configType',
    width: 100,
    customRender: ({ text }: { text: string }) => renderEmpty(text),
  },
  {
    title: '是否前台',
    dataIndex: 'isFrontend',
    key: 'isFrontend',
    width: 90,
    customRender: ({ text }: { text: string }) => renderIsFrontend(text),
  },
  {
    title: '备注',
    dataIndex: 'remark',
    key: 'remark',
    ellipsis: true,
    customRender: ({ text }: { text: string }) => renderEmpty(text),
  },
  {
    title: '创建时间',
    dataIndex: 'createdAt',
    key: 'createdAt',
    width: 160,
    customRender: ({ text }: { text: string }) => renderEmpty(text),
  },
  {
    title: '操作',
    key: 'action',
    width: 140,
    fixed: 'right',
  },
];

/* -------- 新增参数配置 -------- */

/** 新增弹窗可见性 */
const addVisible = ref(false);
/** 新增提交中状态 */
const addSubmitting = ref(false);

/** 新增表单 */
const addForm = reactive({
  configName: '',
  configKey: '',
  configValue: '',
  configType: '',
  isFrontend: '0',
  remark: '',
});

/** 重置新增表单为默认值 */
function resetAddForm() {
  addForm.configName = '';
  addForm.configKey = '';
  addForm.configValue = '';
  addForm.configType = '';
  addForm.isFrontend = '0';
  addForm.remark = '';
}

/** 打开新增弹窗 */
function openAddModal() {
  resetAddForm();
  addVisible.value = true;
}

/** 新增表单校验 */
function validateAddForm(): { ok: boolean; message?: string } {
  const name = addForm.configName?.trim() ?? '';
  const key = addForm.configKey?.trim() ?? '';
  if (!name) {
    return { ok: false, message: '请输入参数名称' };
  }
  if (!key) {
    return { ok: false, message: '请输入参数键名' };
  }
  return { ok: true };
}

/** 确认新增 */
async function onAddOk() {
  const v = validateAddForm();
  if (!v.ok) {
    message.error(v.message);
    return;
  }
  addSubmitting.value = true;
  try {
    await createConfig({
      configName: addForm.configName.trim(),
      configKey: addForm.configKey.trim(),
      configValue: addForm.configValue?.trim() ?? '',
      configType: addForm.configType?.trim() ?? '',
      isFrontend: addForm.isFrontend,
      remark: addForm.remark?.trim() ?? '',
    });
    message.success('新增成功');
    addVisible.value = false;
    fetchConfigList();
  } catch (e: unknown) {
    const err = e as { message?: string; response?: { data?: { msg?: string } } };
    message.error(err?.message || err?.response?.data?.msg || '新增失败');
  } finally {
    addSubmitting.value = false;
  }
}

/** 取消新增 */
function onAddCancel() {
  addVisible.value = false;
}

/* -------- 编辑参数配置 -------- */

/** 编辑弹窗可见性 */
const editVisible = ref(false);
/** 编辑提交中状态 */
const editSubmitting = ref(false);
/** 编辑详情加载中状态 */
const editLoading = ref(false);
/** 当前编辑的配置 ID */
const editConfigId = ref<number | null>(null);

/** 编辑表单 */
const editForm = reactive({
  configName: '',
  configKey: '',
  configValue: '',
  configType: '',
  isFrontend: '0',
  remark: '',
});

/** 打开编辑弹窗 */
async function openEditModal(record: SysConfigItem) {
  editConfigId.value = record.id;
  editLoading.value = true;
  editVisible.value = true;
  try {
    const detail = await getConfigDetail(record.id);
    editForm.configName = detail.configName ?? '';
    editForm.configKey = detail.configKey ?? '';
    editForm.configValue = detail.configValue ?? '';
    editForm.configType = detail.configType ?? '';
    editForm.isFrontend = detail.isFrontend ?? '0';
    editForm.remark = detail.remark ?? '';
  } catch (e: unknown) {
    const err = e as { message?: string };
    message.error(err?.message || '获取参数配置详情失败');
    editVisible.value = false;
  } finally {
    editLoading.value = false;
  }
}

/** 编辑表单校验 */
function validateEditForm(): { ok: boolean; message?: string } {
  const name = editForm.configName?.trim() ?? '';
  const key = editForm.configKey?.trim() ?? '';
  if (!name) {
    return { ok: false, message: '请输入参数名称' };
  }
  if (!key) {
    return { ok: false, message: '请输入参数键名' };
  }
  return { ok: true };
}

/** 确认编辑 */
async function onEditOk() {
  if (editConfigId.value === null) return;
  const v = validateEditForm();
  if (!v.ok) {
    message.error(v.message);
    return;
  }
  editSubmitting.value = true;
  try {
    await updateConfig(editConfigId.value, {
      configName: editForm.configName.trim(),
      configKey: editForm.configKey.trim(),
      configValue: editForm.configValue?.trim() ?? '',
      configType: editForm.configType?.trim() ?? '',
      isFrontend: editForm.isFrontend,
      remark: editForm.remark?.trim() ?? '',
    });
    message.success('编辑成功');
    editVisible.value = false;
    fetchConfigList();
  } catch (e: unknown) {
    const err = e as { message?: string; response?: { data?: { msg?: string } } };
    message.error(err?.message || err?.response?.data?.msg || '编辑失败');
  } finally {
    editSubmitting.value = false;
  }
}

/** 取消编辑 */
function onEditCancel() {
  editVisible.value = false;
}

/* -------- 删除参数配置 -------- */

/** 删除参数配置 */
function onDelete(record: SysConfigItem) {
  const name = record.configName || record.configKey || `ID:${record.id}`;
  Modal.confirm({
    title: '确认删除',
    content: `确定要删除参数配置「${name}」吗？删除后不可恢复。`,
    okText: '确定',
    okType: 'danger',
    cancelText: '取消',
    async onOk() {
      try {
        await deleteConfig([record.id]);
        message.success('删除成功');
        fetchConfigList();
      } catch (e: unknown) {
        const err = e as {
          message?: string;
          response?: { data?: { msg?: string } };
        };
        message.error(err?.message || err?.response?.data?.msg || '删除失败');
      }
    },
  });
}

/** 新增/编辑共用的是否前台下拉（不含"全部"） */
const isFrontendEditOptions = [
  { value: '0', label: '否' },
  { value: '1', label: '是' },
];

onMounted(() => {
  fetchConfigList();
});
</script>

<template>
  <div class="p-4">
    <!-- 页面标题 -->
    <div class="mb-4 flex items-center justify-between">
      <h2 class="text-lg font-medium">参数配置</h2>
      <div class="flex gap-2">
        <Button @click="fetchConfigList">刷新</Button>
        <Button type="primary" @click="openAddModal">新增参数</Button>
      </div>
    </div>

    <!-- 搜索区 -->
    <div class="mb-4 flex flex-wrap items-center gap-2">
      <span class="text-sm text-gray-600">参数名称：</span>
      <Input
        v-model:value="searchConfigName"
        placeholder="请输入参数名称"
        allow-clear
        class="w-52"
        @press-enter="onSearch"
      />
      <span class="ml-2 text-sm text-gray-600">参数键名：</span>
      <Input
        v-model:value="searchConfigKey"
        placeholder="请输入参数键名"
        allow-clear
        class="w-52"
        @press-enter="onSearch"
      />
      <span class="ml-2 text-sm text-gray-600">是否前台：</span>
      <Select
        v-model:value="searchIsFrontend"
        :options="isFrontendOptions"
        class="w-28"
        placeholder="请选择"
      />
      <Button type="primary" size="small" @click="onSearch">查询</Button>
      <Button size="small" @click="onReset">重置</Button>
    </div>

    <!-- 错误提示 -->
    <div v-if="errorMsg" class="mb-4 text-red-600">{{ errorMsg }}</div>

    <!-- 表格区 -->
    <Table
      :columns="columns"
      :data-source="tableData"
      :loading="loading"
      :pagination="pagination"
      :row-key="(record: SysConfigItem) => record.id"
      size="small"
      bordered
      @change="onTableChange"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'action'">
          <Button
            type="link"
            size="small"
            @click="openEditModal(record as SysConfigItem)"
          >
            编辑
          </Button>
          <Button
            type="link"
            size="small"
            danger
            @click="onDelete(record as SysConfigItem)"
          >
            删除
          </Button>
        </template>
      </template>
    </Table>

    <!-- 新增弹窗 -->
    <Modal
      v-model:open="addVisible"
      title="新增参数"
      :confirm-loading="addSubmitting"
      ok-text="保存"
      cancel-text="取消"
      @ok="onAddOk"
      @cancel="onAddCancel"
    >
      <Form :label-col="{ span: 6 }" :wrapper-col="{ span: 16 }" class="mt-4">
        <FormItem label="参数名称" required>
          <Input
            v-model:value="addForm.configName"
            placeholder="请输入参数名称"
            allow-clear
          />
        </FormItem>
        <FormItem label="参数键名" required>
          <Input
            v-model:value="addForm.configKey"
            placeholder="请输入参数键名"
            allow-clear
          />
        </FormItem>
        <FormItem label="参数值">
          <Input
            v-model:value="addForm.configValue"
            placeholder="请输入参数值"
            allow-clear
          />
        </FormItem>
        <FormItem label="配置类型">
          <Input
            v-model:value="addForm.configType"
            placeholder="如：文本、数字等"
            allow-clear
          />
        </FormItem>
        <FormItem label="是否前台">
          <Select
            v-model:value="addForm.isFrontend"
            :options="isFrontendEditOptions"
            class="w-full"
          />
        </FormItem>
        <FormItem label="备注">
          <Input
            v-model:value="addForm.remark"
            placeholder="请输入备注"
            allow-clear
            type="textarea"
            :rows="2"
          />
        </FormItem>
      </Form>
    </Modal>

    <!-- 编辑弹窗 -->
    <Modal
      v-model:open="editVisible"
      title="编辑参数"
      :confirm-loading="editSubmitting"
      :ok-button-props="{ disabled: editLoading }"
      ok-text="保存"
      cancel-text="取消"
      @ok="onEditOk"
      @cancel="onEditCancel"
    >
      <div v-if="editLoading" class="py-8 text-center text-gray-400">
        加载详情中…
      </div>
      <Form
        v-else
        :label-col="{ span: 6 }"
        :wrapper-col="{ span: 16 }"
        class="mt-4"
      >
        <FormItem label="参数名称" required>
          <Input
            v-model:value="editForm.configName"
            placeholder="请输入参数名称"
            allow-clear
          />
        </FormItem>
        <FormItem label="参数键名" required>
          <Input
            v-model:value="editForm.configKey"
            placeholder="请输入参数键名"
            allow-clear
          />
        </FormItem>
        <FormItem label="参数值">
          <Input
            v-model:value="editForm.configValue"
            placeholder="请输入参数值"
            allow-clear
          />
        </FormItem>
        <FormItem label="配置类型">
          <Input
            v-model:value="editForm.configType"
            placeholder="如：文本、数字等"
            allow-clear
          />
        </FormItem>
        <FormItem label="是否前台">
          <Select
            v-model:value="editForm.isFrontend"
            :options="isFrontendEditOptions"
            class="w-full"
          />
        </FormItem>
        <FormItem label="备注">
          <Input
            v-model:value="editForm.remark"
            placeholder="请输入备注"
            allow-clear
            type="textarea"
            :rows="2"
          />
        </FormItem>
      </Form>
    </Modal>
  </div>
</template>
