<script lang="ts" setup>
import type { TableColumnType } from 'ant-design-vue';

import type {
  CreateSkuData,
  SkuBrandItem,
  SkuCategoryTreeNode,
  SkuItem,
  SpuItem,
  SpuPageResult,
} from '#/api/core';

/**
 * 产品中心 - SPU 主管理页
 * 列表 + 类目/品牌/状态 过滤 + 新增/编辑（富文本+主图+详情多图+SKU 子表）+ 提交审核 + 详情查看
 */
import { computed, h, onMounted, reactive, ref } from 'vue';

import {
  Button,
  Drawer,
  Input,
  InputNumber,
  message,
  Modal,
  Select,
  Table,
  Tag,
  TreeSelect,
} from 'ant-design-vue';

import {
  createSku,
  createSpu,
  deleteSku,
  deleteSpu,
  getSkuBrandPage,
  getSkuCategoryTree,
  getSkuPage,
  getSpuDetail,
  getSpuPage,
  submitSpu,
  updateSku,
  updateSpu,
} from '#/api/core';
import AdminActionButton from '#/components/admin/action-button.vue';
import AdminErrorAlert from '#/components/admin/error-alert.vue';
import AdminFilterField from '#/components/admin/filter-field.vue';
import AdminPageShell from '#/components/admin/page-shell.vue';
import { useAdminTable } from '#/composables/use-admin-table';
import { formatAdminDateTime, renderAdminEmpty } from '#/utils/admin-crud';

import AnnouncementCoverUpload from '../sys-announcement/cover-upload.vue';
import AnnouncementRichTextEditor from '../sys-announcement/rich-text-editor.vue';
import SpuDetailImages from './detail-images.vue';

/* -------- 状态映射 -------- */
const SPU_STATUS = {
  draft: 1,
  reviewing: 2,
  approved: 3,
  rejected: 4,
  offline: 5,
} as const;

const statusFilterOptions = [
  { value: '' as const, label: '全部' },
  { value: SPU_STATUS.draft, label: '草稿' },
  { value: SPU_STATUS.reviewing, label: '待审核' },
  { value: SPU_STATUS.approved, label: '已通过' },
  { value: SPU_STATUS.rejected, label: '已驳回' },
  { value: SPU_STATUS.offline, label: '已下架' },
];

const editStatusOptions = [
  { value: SPU_STATUS.draft, label: '草稿' },
  { value: SPU_STATUS.reviewing, label: '待审核' },
  { value: SPU_STATUS.approved, label: '已通过' },
  { value: SPU_STATUS.rejected, label: '已驳回' },
  { value: SPU_STATUS.offline, label: '已下架' },
];

function renderStatusTag(status: number) {
  if (status === SPU_STATUS.draft)
    return h(Tag, { color: 'default' }, () => '草稿');
  if (status === SPU_STATUS.reviewing)
    return h(Tag, { color: 'blue' }, () => '待审核');
  if (status === SPU_STATUS.approved)
    return h(Tag, { color: 'green' }, () => '已通过');
  if (status === SPU_STATUS.rejected)
    return h(Tag, { color: 'red' }, () => '已驳回');
  if (status === SPU_STATUS.offline)
    return h(Tag, { color: 'orange' }, () => '已下架');
  return String(status);
}

/* -------- 列表 -------- */
const {
  errorMsg,
  fetchList,
  loading,
  onReset,
  onSearch,
  onTableChange,
  pagination,
  query,
  tableData,
} = useAdminTable<
  SpuItem,
  {
    brandId: '' | number;
    categoryId: '' | number;
    spuName: string;
    status: '' | number;
  },
  {
    brandId?: number;
    categoryId?: number;
    spuName?: string;
    status?: number;
  }
>({
  createParams: (q) => ({
    spuName: q.spuName.trim() || undefined,
    status: q.status === '' ? undefined : Number(q.status),
    categoryId: q.categoryId === '' ? undefined : Number(q.categoryId),
    brandId: q.brandId === '' ? undefined : Number(q.brandId),
  }),
  createQuery: () => ({
    spuName: '',
    status: '' as '' | number,
    categoryId: '' as '' | number,
    brandId: '' as '' | number,
  }),
  fallbackErrorMessage: '加载 SPU 列表失败',
  fetcher: async (params): Promise<SpuPageResult> => {
    return await getSpuPage(params);
  },
});

/* -------- 类目树 / 品牌下拉 -------- */
const categoryTreeData = ref<unknown[]>([]);
const brandOptions = ref<{ label: string; value: number }[]>([]);
const categoryNameMap = ref<Map<number, string>>(new Map());
const brandNameMap = ref<Map<number, string>>(new Map());

function categoryTreeToSelectNodes(
  nodes: SkuCategoryTreeNode[],
): unknown[] {
  return (nodes || []).map((n) => {
    categoryNameMap.value.set(n.categoryId, n.categoryName);
    const children = n.children ? categoryTreeToSelectNodes(n.children) : [];
    const isLeaf = !n.children || n.children.length === 0;
    return {
      label: n.categoryName,
      value: n.categoryId,
      key: n.categoryId,
      // 仅叶子可选
      selectable: isLeaf,
      disabled: !isLeaf,
      children,
    };
  });
}

async function loadCategoryTree() {
  try {
    const tree = await getSkuCategoryTree({ status: 2 });
    categoryNameMap.value = new Map();
    categoryTreeData.value = categoryTreeToSelectNodes(tree);
  } catch (error) {
    console.warn('load category tree failed', error);
  }
}

async function loadBrandOptions() {
  try {
    const result = await getSkuBrandPage({
      pageIndex: 1,
      pageSize: 200,
      status: 2,
    });
    const list = result.list || [];
    brandNameMap.value = new Map(
      list.map((b: SkuBrandItem) => [b.brandId, b.brandName]),
    );
    brandOptions.value = list.map((b) => ({
      label: b.brandName,
      value: b.brandId,
    }));
  } catch (error) {
    console.warn('load brand options failed', error);
  }
}

/* -------- 列定义 -------- */
const columns: TableColumnType[] = [
  {
    title: '主图',
    dataIndex: 'mainImageUrl',
    key: 'mainImageUrl',
    width: 80,
    customRender: ({ text }: { text: string }) =>
      text
        ? h('img', {
            src: text,
            class: 'w-12 h-12 object-cover rounded',
            alt: 'main',
          })
        : renderAdminEmpty(''),
  },
  {
    title: 'SPU 名称',
    dataIndex: 'spuName',
    key: 'spuName',
    ellipsis: true,
    width: 200,
  },
  {
    title: 'SPU 编码',
    dataIndex: 'spuCode',
    key: 'spuCode',
    width: 140,
  },
  {
    title: '类目',
    dataIndex: 'categoryId',
    key: 'categoryId',
    width: 140,
    customRender: ({ text }: { text: number }) =>
      renderAdminEmpty(categoryNameMap.value.get(text) || ''),
  },
  {
    title: '品牌',
    dataIndex: 'brandId',
    key: 'brandId',
    width: 120,
    customRender: ({ text }: { text: number }) =>
      renderAdminEmpty(brandNameMap.value.get(text) || ''),
  },
  {
    title: '状态',
    dataIndex: 'status',
    key: 'status',
    width: 100,
    customRender: ({ text }: { text: number }) => renderStatusTag(text),
  },
  {
    title: '创建人',
    dataIndex: 'creatorId',
    key: 'creatorId',
    width: 80,
    customRender: ({ text }: { text: number }) =>
      renderAdminEmpty(text ? String(text) : ''),
  },
  {
    title: '更新时间',
    dataIndex: 'updatedAt',
    key: 'updatedAt',
    width: 170,
    customRender: ({ text }: { text: string }) => formatAdminDateTime(text),
  },
  {
    title: '操作',
    key: 'action',
    width: 240,
    fixed: 'right',
  },
];

/* -------- 表单：新增/编辑共用 -------- */
type SkuRow = {
  // 用于新增行的本地标识；负数代表本地未保存
  _localId: number;
  price: number;
  skuCode: string;
  skuId?: number;
  skuName: string;
  spec: string;
  status: number;
  unit: string;
};

let nextLocalId = -1;

function emptySkuRow(): SkuRow {
  nextLocalId -= 1;
  return {
    _localId: nextLocalId,
    skuCode: '',
    skuName: '',
    spec: '',
    unit: '',
    price: 0,
    status: 2,
  };
}

type FormState = {
  brandId: number | undefined;
  categoryId: number | undefined;
  description: string;
  detailImages: string[];
  mainImageUrl: string;
  skuRows: SkuRow[];
  spuCode: string;
  spuName: string;
  status: number;
};

function emptyForm(): FormState {
  return {
    spuName: '',
    spuCode: '',
    categoryId: undefined,
    brandId: undefined,
    description: '',
    mainImageUrl: '',
    detailImages: [],
    skuRows: [emptySkuRow()],
    status: SPU_STATUS.draft,
  };
}

function parseDetailImages(raw?: string): string[] {
  if (!raw) return [];
  try {
    const arr = JSON.parse(raw);
    return Array.isArray(arr) ? arr.filter((s) => typeof s === 'string') : [];
  } catch {
    return [];
  }
}

/* -------- 编辑抽屉（新增 + 编辑统一） -------- */
const editorOpen = ref(false);
const editorMode = ref<'create' | 'edit'>('create');
const editorSubmitting = ref(false);
const editorLoading = ref(false);
const editorSpuId = ref<null | number>(null);
const editorForm = reactive<FormState>(emptyForm());

const isReviewing = computed(
  () => editorMode.value === 'edit' && editorForm.status === SPU_STATUS.reviewing,
);

function openCreate() {
  editorMode.value = 'create';
  editorSpuId.value = null;
  Object.assign(editorForm, emptyForm());
  editorOpen.value = true;
}

async function openEdit(record: SpuItem) {
  editorMode.value = 'edit';
  editorSpuId.value = record.spuId;
  Object.assign(editorForm, emptyForm());
  editorOpen.value = true;
  editorLoading.value = true;
  try {
    const detail = await getSpuDetail(record.spuId);
    editorForm.spuName = detail.spuName ?? '';
    editorForm.spuCode = detail.spuCode ?? '';
    editorForm.categoryId = detail.categoryId ?? undefined;
    editorForm.brandId = detail.brandId ?? undefined;
    editorForm.description = detail.description ?? '';
    editorForm.mainImageUrl = detail.mainImageUrl ?? '';
    editorForm.detailImages = parseDetailImages(detail.detailImages);
    editorForm.status = detail.status ?? SPU_STATUS.draft;

    // 加载该 SPU 关联的 SKU 列表
    const skuPage = await getSkuPage({
      spuId: record.spuId,
      pageIndex: 1,
      pageSize: 200,
    });
    const rows: SkuRow[] = (skuPage.list || []).map((s: SkuItem) => ({
      _localId: s.skuId,
      skuId: s.skuId,
      skuCode: s.skuCode ?? '',
      skuName: s.skuName ?? '',
      spec: s.spec ?? '',
      unit: s.unit ?? '',
      price: Number(s.price ?? 0),
      status: s.status ?? 2,
    }));
    editorForm.skuRows = rows.length > 0 ? rows : [emptySkuRow()];
  } catch (error: any) {
    message.error(error?.message || '获取 SPU 详情失败');
    editorOpen.value = false;
  } finally {
    editorLoading.value = false;
  }
}

function addSkuRow() {
  editorForm.skuRows.push(emptySkuRow());
}

function removeSkuRow(localId: number) {
  if (editorForm.skuRows.length <= 1) {
    // 至少保留 1 行
    editorForm.skuRows = [emptySkuRow()];
    return;
  }
  editorForm.skuRows = editorForm.skuRows.filter((r) => r._localId !== localId);
}

function validateForm(): { ok: boolean; message?: string } {
  if (!editorForm.spuName.trim()) {
    return { ok: false, message: '请输入 SPU 名称' };
  }
  if (!editorForm.spuCode.trim()) {
    return { ok: false, message: '请输入 SPU 编码' };
  }
  if (!editorForm.categoryId) {
    return { ok: false, message: '请选择类目（叶子节点）' };
  }
  if (!editorForm.brandId) {
    return { ok: false, message: '请选择品牌' };
  }
  // SKU 行：忽略全空行；其余行需 skuCode 非空
  const meaningfulRows = editorForm.skuRows.filter(
    (r) => r.skuCode.trim() || r.skuName.trim() || r.skuId !== undefined,
  );
  for (const r of meaningfulRows) {
    if (!r.skuCode.trim()) {
      return { ok: false, message: `存在 SKU 行缺少编码（${r.skuName || ''}）` };
    }
  }
  return { ok: true };
}

async function persistSkuRows(spuId: number) {
  const wantedKeys = new Set<number>();
  for (const r of editorForm.skuRows) {
    const code = r.skuCode.trim();
    if (!code && !r.skuName.trim() && r.skuId === undefined) {
      // 空行跳过
      continue;
    }
    if (!code) continue;
    const payload: CreateSkuData = {
      spuId,
      skuCode: code,
      skuName: r.skuName.trim() || undefined,
      spec: r.spec.trim() || undefined,
      unit: r.unit.trim() || undefined,
      price: Number(r.price ?? 0),
      status: r.status ?? 2,
    };
    if (r.skuId !== undefined) {
      await updateSku(r.skuId, payload);
      wantedKeys.add(r.skuId);
    } else {
      await createSku(payload);
    }
  }
  // 删除：编辑模式下，原有但本次未保留的 SKU
  if (editorMode.value === 'edit') {
    const existingPage = await getSkuPage({
      spuId,
      pageIndex: 1,
      pageSize: 500,
    });
    const toDelete = (existingPage.list || [])
      .map((s) => s.skuId)
      .filter((id) => !wantedKeys.has(id));
    if (toDelete.length > 0) {
      await deleteSku(toDelete);
    }
  }
}

async function onEditorOk() {
  const v = validateForm();
  if (!v.ok) {
    message.error(v.message);
    return;
  }
  editorSubmitting.value = true;
  try {
    const payload = {
      spuName: editorForm.spuName.trim(),
      spuCode: editorForm.spuCode.trim(),
      categoryId: editorForm.categoryId,
      brandId: editorForm.brandId,
      description: editorForm.description ?? '',
      mainImageUrl: editorForm.mainImageUrl?.trim() || '',
      detailImages: JSON.stringify(editorForm.detailImages || []),
      status: editorForm.status ?? SPU_STATUS.draft,
    };
    let spuId: number;
    if (editorMode.value === 'create') {
      spuId = await createSpu(payload);
    } else {
      if (editorSpuId.value === null) {
        throw new Error('缺少 SPU ID');
      }
      await updateSpu(editorSpuId.value, payload);
      spuId = editorSpuId.value;
    }
    await persistSkuRows(spuId);
    message.success('保存成功');
    editorOpen.value = false;
    void fetchList();
  } catch (error: any) {
    message.error(error?.message || '保存失败');
  } finally {
    editorSubmitting.value = false;
  }
}

/* -------- 提交审核 -------- */
async function onSubmitReview(record: SpuItem) {
  Modal.confirm({
    title: '提交审核',
    content: `确定要提交「${record.spuName}」进入审核吗？提交后不可修改。`,
    okText: '提交',
    cancelText: '取消',
    async onOk() {
      try {
        await submitSpu(record.spuId, {});
        message.success('已提交审核');
        void fetchList();
      } catch (error: any) {
        message.error(error?.message || '提交审核失败');
      }
    },
  });
}

/* -------- 删除 -------- */
function onDelete(record: SpuItem) {
  Modal.confirm({
    title: '确认删除',
    content: `确定要删除 SPU「${record.spuName}」吗？关联 SKU 也会被删除，且不可恢复。`,
    okText: '确定',
    okType: 'danger',
    cancelText: '取消',
    async onOk() {
      try {
        await deleteSpu([record.spuId]);
        message.success('删除成功');
        void fetchList();
      } catch (error: any) {
        message.error(error?.message || '删除失败');
      }
    },
  });
}

/* -------- 详情查看 -------- */
const detailOpen = ref(false);
const detailLoading = ref(false);
const detailItem = ref<null | SpuItem>(null);
const detailImages = ref<string[]>([]);
const detailSkus = ref<SkuItem[]>([]);

async function openDetail(record: SpuItem) {
  detailOpen.value = true;
  detailLoading.value = true;
  detailItem.value = null;
  detailImages.value = [];
  detailSkus.value = [];
  try {
    const d = await getSpuDetail(record.spuId);
    detailItem.value = d;
    detailImages.value = parseDetailImages(d.detailImages);
    const skuPage = await getSkuPage({
      spuId: record.spuId,
      pageIndex: 1,
      pageSize: 200,
    });
    detailSkus.value = skuPage.list || [];
  } catch (error: any) {
    message.error(error?.message || '获取 SPU 详情失败');
    detailOpen.value = false;
  } finally {
    detailLoading.value = false;
  }
}

const detailSkuColumns: TableColumnType[] = [
  { title: 'SKU 编码', dataIndex: 'skuCode', key: 'skuCode', width: 140 },
  { title: 'SKU 名称', dataIndex: 'skuName', key: 'skuName', ellipsis: true },
  { title: '规格', dataIndex: 'spec', key: 'spec', width: 120 },
  { title: '单位', dataIndex: 'unit', key: 'unit', width: 80 },
  { title: '价格', dataIndex: 'price', key: 'price', width: 100 },
];

onMounted(() => {
  void loadCategoryTree();
  void loadBrandOptions().then(() => fetchList());
});
</script>

<template>
  <AdminPageShell>
    <template #eyebrow>Product Center</template>
    <template #title>SPU 管理</template>
    <template #description>
      统一维护 SPU 主信息、详情图、富文本描述与子 SKU。提交审核后由流程中心驱动审批。
    </template>
    <template #header-extra>
      <AdminActionButton
        type="primary"
        codes="admin:spu:add"
        @click="openCreate"
      >
        新增 SPU
      </AdminActionButton>
    </template>
    <template #filters>
      <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <AdminFilterField label="SPU 名称">
          <Input
            v-model:value="query.spuName"
            placeholder="请输入名称关键词"
            allow-clear
            @press-enter="onSearch"
          />
        </AdminFilterField>
        <AdminFilterField label="状态">
          <Select
            v-model:value="query.status"
            :options="statusFilterOptions"
            placeholder="请选择状态"
          />
        </AdminFilterField>
        <AdminFilterField label="类目">
          <TreeSelect
            v-model:value="query.categoryId"
            :tree-data="categoryTreeData as any"
            tree-default-expand-all
            allow-clear
            placeholder="请选择类目（叶子）"
            class="w-full"
          />
        </AdminFilterField>
        <AdminFilterField label="品牌">
          <Select
            v-model:value="query.brandId"
            :options="brandOptions"
            placeholder="请选择品牌"
            allow-clear
            show-search
            :filter-option="
              (input: string, option: any) =>
                String(option?.label ?? '').toLowerCase().includes(input.toLowerCase())
            "
          />
        </AdminFilterField>
      </div>
    </template>
    <template #filter-actions>
      <Button type="primary" @click="onSearch">查询</Button>
      <Button @click="onReset">重置</Button>
    </template>
    <template #toolbar>
      <div class="text-base font-semibold text-slate-900">SPU 列表</div>
    </template>

    <AdminErrorAlert :message="errorMsg" />

    <Table
      :columns="columns"
      :data-source="tableData"
      :loading="loading"
      :pagination="pagination"
      :row-key="(r: SpuItem) => r.spuId"
      :scroll="{ x: 1280 }"
      size="middle"
      @change="(pag) => onTableChange(pag)"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'action'">
          <AdminActionButton
            type="link"
            size="small"
            @click="openDetail(record as SpuItem)"
          >
            查看
          </AdminActionButton>
          <AdminActionButton
            v-if="
              (record as SpuItem).status === SPU_STATUS.draft ||
              (record as SpuItem).status === SPU_STATUS.rejected
            "
            type="link"
            size="small"
            codes="admin:spu:edit"
            @click="openEdit(record as SpuItem)"
          >
            编辑
          </AdminActionButton>
          <AdminActionButton
            v-if="
              (record as SpuItem).status === SPU_STATUS.draft ||
              (record as SpuItem).status === SPU_STATUS.rejected
            "
            type="link"
            size="small"
            codes="admin:spu:edit"
            @click="onSubmitReview(record as SpuItem)"
          >
            提交审核
          </AdminActionButton>
          <AdminActionButton
            v-if="(record as SpuItem).status !== SPU_STATUS.reviewing"
            type="link"
            size="small"
            danger
            codes="admin:spu:remove"
            @click="onDelete(record as SpuItem)"
          >
            删除
          </AdminActionButton>
        </template>
      </template>
    </Table>

    <!-- 编辑/新增 抽屉 -->
    <Drawer
      v-model:open="editorOpen"
      :title="editorMode === 'create' ? '新增 SPU' : '编辑 SPU'"
      :width="960"
      placement="right"
      :body-style="{ paddingBottom: '80px' }"
    >
      <div v-if="editorLoading" class="py-8 text-center text-gray-400">
        加载中…
      </div>
      <div v-else class="space-y-5">
        <div v-if="isReviewing" class="rounded-sm bg-amber-50 p-3 text-sm text-amber-700">
          当前 SPU 处于待审核状态，仅可查看，不能编辑。
        </div>
        <div class="grid gap-4 md:grid-cols-2">
          <div>
            <div class="mb-1 text-sm">
              SPU 名称<span class="text-red-500"> *</span>
            </div>
            <Input
              v-model:value="editorForm.spuName"
              placeholder="请输入 SPU 名称"
              :disabled="isReviewing"
            />
          </div>
          <div>
            <div class="mb-1 text-sm">
              SPU 编码<span class="text-red-500"> *</span>
            </div>
            <Input
              v-model:value="editorForm.spuCode"
              placeholder="请输入 SPU 编码"
              :disabled="editorMode === 'edit'"
            />
          </div>
          <div>
            <div class="mb-1 text-sm">
              类目<span class="text-red-500"> *</span>
            </div>
            <TreeSelect
              v-model:value="editorForm.categoryId"
              :tree-data="categoryTreeData as any"
              tree-default-expand-all
              placeholder="请选择类目（叶子）"
              class="w-full"
              allow-clear
              :disabled="isReviewing"
            />
          </div>
          <div>
            <div class="mb-1 text-sm">
              品牌<span class="text-red-500"> *</span>
            </div>
            <Select
              v-model:value="editorForm.brandId"
              :options="brandOptions"
              placeholder="请选择品牌"
              show-search
              :filter-option="
                (input: string, option: any) =>
                  String(option?.label ?? '').toLowerCase().includes(input.toLowerCase())
              "
              class="w-full"
              :disabled="isReviewing"
            />
          </div>
          <div>
            <div class="mb-1 text-sm">状态</div>
            <Select
              v-model:value="editorForm.status"
              :options="editStatusOptions"
              :disabled="isReviewing"
            />
          </div>
        </div>

        <div>
          <div class="mb-1 text-sm">主图</div>
          <AnnouncementCoverUpload
            v-model="editorForm.mainImageUrl"
            :business-id="editorSpuId ?? undefined"
            business-type="spu-main-image"
            hint="建议 800×800，jpg/png/webp"
            :disabled="isReviewing"
          />
        </div>

        <div>
          <div class="mb-1 text-sm">详情图（最多 3 张）</div>
          <SpuDetailImages
            v-model="editorForm.detailImages"
            :business-id="editorSpuId ?? undefined"
            :max="3"
            :disabled="isReviewing"
          />
        </div>

        <div>
          <div class="mb-1 text-sm">详情（富文本，入库自动 XSS 过滤）</div>
          <AnnouncementRichTextEditor
            v-model="editorForm.description"
            upload-business-type="spu-inline"
            :height="280"
            placeholder="请输入 SPU 详情…"
          />
        </div>

        <div>
          <div class="mb-2 flex items-center justify-between">
            <div class="text-sm font-medium">SKU 子表（{{ editorForm.skuRows.length }}）</div>
            <Button
              size="small"
              :disabled="isReviewing"
              @click="addSkuRow"
            >
              + 新增 SKU
            </Button>
          </div>
          <div class="overflow-x-auto rounded-sm border border-slate-200">
            <table class="w-full min-w-[760px] text-sm">
              <thead class="bg-slate-50 text-slate-600">
                <tr>
                  <th class="px-2 py-2 text-left font-medium">SKU 编码 *</th>
                  <th class="px-2 py-2 text-left font-medium">SKU 名称</th>
                  <th class="px-2 py-2 text-left font-medium">规格</th>
                  <th class="px-2 py-2 text-left font-medium">单位</th>
                  <th class="px-2 py-2 text-left font-medium">价格</th>
                  <th class="px-2 py-2 text-left font-medium" style="width: 64px">
                    操作
                  </th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="row in editorForm.skuRows"
                  :key="row._localId"
                  class="border-t border-slate-100"
                >
                  <td class="px-2 py-1.5">
                    <Input
                      v-model:value="row.skuCode"
                      size="small"
                      placeholder="编码"
                      :disabled="isReviewing"
                    />
                  </td>
                  <td class="px-2 py-1.5">
                    <Input
                      v-model:value="row.skuName"
                      size="small"
                      placeholder="名称"
                      :disabled="isReviewing"
                    />
                  </td>
                  <td class="px-2 py-1.5">
                    <Input
                      v-model:value="row.spec"
                      size="small"
                      placeholder="规格"
                      :disabled="isReviewing"
                    />
                  </td>
                  <td class="px-2 py-1.5">
                    <Input
                      v-model:value="row.unit"
                      size="small"
                      placeholder="单位"
                      :disabled="isReviewing"
                    />
                  </td>
                  <td class="px-2 py-1.5">
                    <InputNumber
                      v-model:value="row.price"
                      size="small"
                      :min="0"
                      :step="0.01"
                      class="w-full"
                      :disabled="isReviewing"
                    />
                  </td>
                  <td class="px-2 py-1.5">
                    <Button
                      type="link"
                      danger
                      size="small"
                      :disabled="isReviewing"
                      @click="removeSkuRow(row._localId)"
                    >
                      删除
                    </Button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <template #footer>
        <div class="flex justify-end gap-2">
          <Button @click="editorOpen = false">取消</Button>
          <Button
            type="primary"
            :loading="editorSubmitting"
            :disabled="isReviewing"
            @click="onEditorOk"
          >
            保存
          </Button>
        </div>
      </template>
    </Drawer>

    <!-- 详情查看抽屉 -->
    <Drawer
      v-model:open="detailOpen"
      :title="detailItem?.spuName || 'SPU 详情'"
      :width="800"
      placement="right"
    >
      <div v-if="detailLoading" class="py-8 text-center text-gray-400">
        加载中…
      </div>
      <div v-else-if="detailItem" class="space-y-5">
        <div class="text-sm text-slate-500">
          状态：<component :is="renderStatusTag(detailItem.status)" />
          <span class="ml-4">编码：{{ detailItem.spuCode }}</span>
          <span class="ml-4">
            类目：{{ categoryNameMap.get(detailItem.categoryId) || '-' }}
          </span>
          <span class="ml-4">
            品牌：{{ brandNameMap.get(detailItem.brandId) || '-' }}
          </span>
        </div>
        <div v-if="detailItem.mainImageUrl">
          <div class="mb-1 text-sm font-medium">主图</div>
          <img
            :src="detailItem.mainImageUrl"
            class="max-h-72 rounded-sm object-contain"
            alt="main"
          />
        </div>
        <div v-if="detailImages.length > 0">
          <div class="mb-1 text-sm font-medium">详情图</div>
          <div class="flex flex-wrap gap-2">
            <img
              v-for="(url, idx) in detailImages"
              :key="`${url}-${idx}`"
              :src="url"
              class="h-28 w-28 rounded-sm object-cover"
              alt="detail"
            />
          </div>
        </div>
        <div>
          <div class="mb-1 text-sm font-medium">详情</div>
          <div
            class="prose max-w-none"
            v-html="detailItem.description || '<p class=\'text-gray-400\'>无详情</p>'"
          ></div>
        </div>
        <div>
          <div class="mb-2 text-sm font-medium">SKU 列表（{{ detailSkus.length }}）</div>
          <Table
            :columns="detailSkuColumns"
            :data-source="detailSkus"
            :pagination="false"
            :row-key="(r: SkuItem) => r.skuId"
            size="small"
          />
        </div>
      </div>
    </Drawer>
  </AdminPageShell>
</template>
