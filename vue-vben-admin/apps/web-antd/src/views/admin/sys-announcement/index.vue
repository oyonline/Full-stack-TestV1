<script lang="ts" setup>
import type { TableColumnType } from 'ant-design-vue';
import type { Dayjs } from 'dayjs';

import type {
  AnnouncementItem,
  AnnouncementPageResult,
  SysDeptItem,
} from '#/api/core';

/**
 * 系统管理 - 公告管理
 * 列表 + 搜索 + 分页 + 新增 + 编辑 + 删除 + 详情查看（自动 mark-read）
 */
import { computed, h, onMounted, reactive, ref } from 'vue';

import {
  Button,
  DatePicker,
  Drawer,
  Input,
  InputNumber,
  message,
  Modal,
  Select,
  Switch,
  Table,
  Tag,
  TreeSelect,
} from 'ant-design-vue';
import dayjs from 'dayjs';

import {
  createAnnouncement,
  deleteAnnouncement,
  getAnnouncementDetail,
  getAnnouncementPage,
  getDeptListApi,
  markAnnouncementRead,
  updateAnnouncement,
} from '#/api/core';
import AdminActionButton from '#/components/admin/action-button.vue';
import AdminErrorAlert from '#/components/admin/error-alert.vue';
import AdminFilterField from '#/components/admin/filter-field.vue';
import AdminPageShell from '#/components/admin/page-shell.vue';
import { useAdminTable } from '#/composables/use-admin-table';
import { formatAdminDateTime, renderAdminEmpty } from '#/utils/admin-crud';

import AnnouncementAttachmentSection from './attachment-section.vue';
import AnnouncementCoverUpload from './cover-upload.vue';
import AnnouncementRichTextEditor from './rich-text-editor.vue';

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
  AnnouncementItem,
  {
    isTop: '' | number;
    status: '' | number;
    title: string;
  },
  {
    isTop?: number;
    status?: number;
    title?: string;
  }
>({
  createParams: (q) => ({
    title: q.title.trim() || undefined,
    status: q.status === '' ? undefined : Number(q.status),
    isTop: q.isTop === '' ? undefined : Number(q.isTop),
  }),
  createQuery: () => ({
    title: '',
    status: '' as '' | number,
    isTop: '' as '' | number,
  }),
  fallbackErrorMessage: '加载公告列表失败',
  fetcher: async (params): Promise<AnnouncementPageResult> => {
    return await getAnnouncementPage(params);
  },
});

const statusOptions = [
  { value: '' as const, label: '全部' },
  { value: 1, label: '草稿' },
  { value: 2, label: '已发布' },
  { value: 3, label: '已下线' },
];

const isTopOptions = [
  { value: '' as const, label: '全部' },
  { value: 1, label: '是' },
  { value: 0, label: '否' },
];

function renderStatusTag(status: number) {
  if (status === 1) return h(Tag, { color: 'default' }, () => '草稿');
  if (status === 2) return h(Tag, { color: 'green' }, () => '已发布');
  if (status === 3) return h(Tag, { color: 'red' }, () => '已下线');
  return String(status);
}

const columns: TableColumnType[] = [
  {
    title: '封面',
    dataIndex: 'coverImageUrl',
    key: 'coverImageUrl',
    width: 90,
    customRender: ({ text }: { text: string }) =>
      text
        ? h('img', {
            src: text,
            class: 'w-12 h-12 object-cover rounded',
            alt: 'cover',
          })
        : renderAdminEmpty(''),
  },
  {
    title: '标题',
    dataIndex: 'title',
    key: 'title',
    ellipsis: true,
    width: 240,
  },
  {
    title: '状态',
    dataIndex: 'status',
    key: 'status',
    width: 100,
    customRender: ({ text }: { text: number }) => renderStatusTag(text),
  },
  {
    title: '置顶',
    dataIndex: 'isTop',
    key: 'isTop',
    width: 80,
    customRender: ({ text }: { text: number }) =>
      text === 1
        ? h(Tag, { color: 'gold' }, () => '置顶')
        : renderAdminEmpty(''),
  },
  {
    title: '生效起始',
    dataIndex: 'publishAt',
    key: 'publishAt',
    width: 170,
    customRender: ({ text }: { text: string }) => formatAdminDateTime(text),
  },
  {
    title: '失效时间',
    dataIndex: 'expireAt',
    key: 'expireAt',
    width: 170,
    customRender: ({ text }: { text: string }) => formatAdminDateTime(text),
  },
  {
    title: '已读人数',
    dataIndex: 'readCount',
    key: 'readCount',
    width: 100,
  },
  {
    title: '操作',
    key: 'action',
    width: 200,
    fixed: 'right',
  },
];

/* -------- 部门树 -------- */
const deptTreeData = ref<unknown[]>([]);

function deptListToTree(items: SysDeptItem[]): unknown[] {
  // 后端 dept 已经是 children 嵌套结构；递归转换为 antd TreeSelect 节点
  return (items || []).map((d) => ({
    label: d.deptName,
    value: d.deptId,
    key: d.deptId,
    children: d.children ? deptListToTree(d.children) : [],
  }));
}

async function loadDeptTree() {
  try {
    const list = await getDeptListApi();
    deptTreeData.value = deptListToTree(list);
  } catch (error) {
    // 部门接口失败不阻塞列表，仅 console
    console.warn('load dept tree failed', error);
  }
}

/* -------- 表单：新增/编辑共用 -------- */
type FormState = {
  content: string;
  coverImageUrl: string;
  deptIds: number[];
  expireAt: Dayjs | undefined;
  isTop: boolean;
  publishAt: Dayjs | undefined;
  remark: string;
  status: number;
  title: string;
  topSort: number;
};

function emptyForm(): FormState {
  return {
    title: '',
    content: '',
    coverImageUrl: '',
    status: 1,
    isTop: false,
    topSort: 0,
    publishAt: undefined,
    expireAt: undefined,
    deptIds: [],
    remark: '',
  };
}

/* -------- 新增 -------- */
const addVisible = ref(false);
const addSubmitting = ref(false);
const addForm = reactive<FormState>(emptyForm());

function resetAddForm() {
  Object.assign(addForm, emptyForm());
}

function openAddModal() {
  resetAddForm();
  addVisible.value = true;
}

async function onAddOk() {
  const title = addForm.title.trim();
  if (!title) {
    message.error('请输入公告标题');
    return;
  }
  addSubmitting.value = true;
  try {
    await createAnnouncement({
      title,
      content: addForm.content,
      coverImageUrl: addForm.coverImageUrl?.trim() || '',
      status: addForm.status,
      isTop: addForm.isTop ? 1 : 0,
      topSort: addForm.topSort,
      publishAt: addForm.publishAt ? addForm.publishAt.toISOString() : null,
      expireAt: addForm.expireAt ? addForm.expireAt.toISOString() : null,
      deptIds: addForm.deptIds,
      remark: addForm.remark?.trim() ?? '',
    });
    message.success('新增成功');
    addVisible.value = false;
    void fetchList();
  } catch (error: any) {
    message.error(error?.message || '新增失败');
  } finally {
    addSubmitting.value = false;
  }
}

/* -------- 编辑 -------- */
const editVisible = ref(false);
const editSubmitting = ref(false);
const editLoading = ref(false);
const editId = ref<null | number>(null);
const editForm = reactive<FormState>(emptyForm());

async function openEditModal(record: AnnouncementItem) {
  editId.value = record.announcementId;
  editLoading.value = true;
  editVisible.value = true;
  Object.assign(editForm, emptyForm());
  try {
    const d = await getAnnouncementDetail(record.announcementId);
    editForm.title = d.title ?? '';
    editForm.content = d.content ?? '';
    editForm.coverImageUrl = d.coverImageUrl ?? '';
    editForm.status = d.status ?? 1;
    editForm.isTop = d.isTop === 1;
    editForm.topSort = d.topSort ?? 0;
    editForm.publishAt = d.publishAt ? dayjs(d.publishAt) : undefined;
    editForm.expireAt = d.expireAt ? dayjs(d.expireAt) : undefined;
    editForm.deptIds = d.deptIds ?? [];
    editForm.remark = d.remark ?? '';
  } catch (error: any) {
    message.error(error?.message || '获取公告详情失败');
    editVisible.value = false;
  } finally {
    editLoading.value = false;
  }
}

async function onEditOk() {
  if (editId.value === null) return;
  const title = editForm.title.trim();
  if (!title) {
    message.error('请输入公告标题');
    return;
  }
  editSubmitting.value = true;
  try {
    await updateAnnouncement(editId.value, {
      title,
      content: editForm.content,
      coverImageUrl: editForm.coverImageUrl?.trim() || '',
      status: editForm.status,
      isTop: editForm.isTop ? 1 : 0,
      topSort: editForm.topSort,
      publishAt: editForm.publishAt ? editForm.publishAt.toISOString() : null,
      expireAt: editForm.expireAt ? editForm.expireAt.toISOString() : null,
      deptIds: editForm.deptIds,
      remark: editForm.remark?.trim() ?? '',
    });
    message.success('保存成功');
    editVisible.value = false;
    void fetchList();
  } catch (error: any) {
    message.error(error?.message || '保存失败');
  } finally {
    editSubmitting.value = false;
  }
}

/* -------- 删除 -------- */
function onDelete(record: AnnouncementItem) {
  Modal.confirm({
    title: '确认删除',
    content: `确定要删除公告「${record.title}」吗？删除后将级联清理 scope 与已读记录，且不可恢复。`,
    okText: '确定',
    okType: 'danger',
    cancelText: '取消',
    async onOk() {
      try {
        await deleteAnnouncement([record.announcementId]);
        message.success('删除成功');
        void fetchList();
      } catch (error: any) {
        message.error(error?.message || '删除失败');
      }
    },
  });
}

/* -------- 详情抽屉（打开自动 mark-read） -------- */
const detailVisible = ref(false);
const detailLoading = ref(false);
const detailItem = ref<AnnouncementItem | null>(null);

async function openDetail(record: AnnouncementItem) {
  detailVisible.value = true;
  detailLoading.value = true;
  detailItem.value = null;
  try {
    const d = await getAnnouncementDetail(record.announcementId);
    detailItem.value = d;
    // 进入即调 mark-read（幂等）
    try {
      await markAnnouncementRead(record.announcementId);
    } catch {
      // mark-read 失败不阻塞展示
    }
    void fetchList();
  } catch (error: any) {
    message.error(error?.message || '获取公告详情失败');
    detailVisible.value = false;
  } finally {
    detailLoading.value = false;
  }
}

const editStatusOptions = [
  { value: 1, label: '草稿' },
  { value: 2, label: '已发布' },
  { value: 3, label: '已下线' },
];

const editFormReadOnly = computed(() => editLoading.value);

onMounted(() => {
  void fetchList();
  void loadDeptTree();
});
</script>

<template>
  <AdminPageShell>
    <template #eyebrow>System Admin</template>
    <template #title>公告管理</template>
    <template #description>
      统一维护公告内容、可见部门、生效时间窗口与已读追踪。富文本入库前自动做 XSS
      过滤。
    </template>
    <template #header-extra>
      <AdminActionButton
        type="primary"
        codes="admin:announcement:add"
        @click="openAddModal"
      >
        新增公告
      </AdminActionButton>
    </template>
    <template #filters>
      <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <AdminFilterField label="标题">
          <Input
            v-model:value="query.title"
            placeholder="请输入标题关键词"
            allow-clear
            @press-enter="onSearch"
          />
        </AdminFilterField>
        <AdminFilterField label="状态">
          <Select
            v-model:value="query.status"
            :options="statusOptions"
            placeholder="请选择状态"
          />
        </AdminFilterField>
        <AdminFilterField label="是否置顶">
          <Select
            v-model:value="query.isTop"
            :options="isTopOptions"
            placeholder="请选择"
          />
        </AdminFilterField>
      </div>
    </template>
    <template #filter-actions>
      <Button type="primary" @click="onSearch">查询</Button>
      <Button @click="onReset">重置</Button>
    </template>
    <template #toolbar>
      <div class="text-base font-semibold text-slate-900">公告列表</div>
    </template>

    <AdminErrorAlert :message="errorMsg" />

    <Table
      :columns="columns"
      :data-source="tableData"
      :loading="loading"
      :pagination="pagination"
      :row-key="(r: AnnouncementItem) => r.announcementId"
      :scroll="{ x: 1200 }"
      size="middle"
      @change="(pag) => onTableChange(pag)"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'action'">
          <AdminActionButton
            type="link"
            size="small"
            @click="openDetail(record as AnnouncementItem)"
          >
            查看
          </AdminActionButton>
          <AdminActionButton
            type="link"
            size="small"
            codes="admin:announcement:edit"
            @click="openEditModal(record as AnnouncementItem)"
          >
            编辑
          </AdminActionButton>
          <AdminActionButton
            type="link"
            size="small"
            danger
            codes="admin:announcement:remove"
            @click="onDelete(record as AnnouncementItem)"
          >
            删除
          </AdminActionButton>
        </template>
      </template>
    </Table>

    <!-- 新增弹窗 -->
    <Modal
      v-model:open="addVisible"
      title="新增公告"
      :confirm-loading="addSubmitting"
      :width="820"
      ok-text="保存"
      cancel-text="取消"
      @ok="onAddOk"
    >
      <div class="grid gap-4 md:grid-cols-2">
        <div class="md:col-span-2">
          <div class="mb-1 text-sm">
            标题<span class="text-red-500"> *</span>
          </div>
          <Input v-model:value="addForm.title" placeholder="请输入公告标题" />
        </div>
        <div class="md:col-span-2">
          <div class="mb-1 text-sm">封面图</div>
          <AnnouncementCoverUpload v-model="addForm.coverImageUrl" />
        </div>
        <div class="md:col-span-2">
          <div class="mb-1 text-sm">正文（入库自动 XSS 过滤）</div>
          <AnnouncementRichTextEditor v-model="addForm.content" />
        </div>
        <div>
          <div class="mb-1 text-sm">状态</div>
          <Select v-model:value="addForm.status" :options="editStatusOptions" />
        </div>
        <div>
          <div class="mb-1 text-sm">是否置顶</div>
          <Switch v-model:checked="addForm.isTop" />
          <span class="ml-3 text-sm">
            排序值
            <InputNumber
              v-model:value="addForm.topSort"
              :min="0"
              class="ml-2"
            />
          </span>
        </div>
        <div>
          <div class="mb-1 text-sm">生效起始</div>
          <DatePicker
            v-model:value="addForm.publishAt"
            show-time
            placeholder="可空 = 立即生效"
            class="w-full"
          />
        </div>
        <div>
          <div class="mb-1 text-sm">失效时间</div>
          <DatePicker
            v-model:value="addForm.expireAt"
            show-time
            placeholder="可空 = 永不过期"
            class="w-full"
          />
        </div>
        <div class="md:col-span-2">
          <div class="mb-1 text-sm">可见部门（多选）</div>
          <TreeSelect
            v-model:value="addForm.deptIds"
            :tree-data="deptTreeData as any"
            tree-checkable
            tree-default-expand-all
            multiple
            allow-clear
            placeholder="请选择可见部门（不选则任何人都看不到）"
            class="w-full"
          />
        </div>
        <div class="md:col-span-2">
          <div class="mb-1 text-sm">备注</div>
          <Input.TextArea v-model:value="addForm.remark" :rows="2" />
        </div>
      </div>
    </Modal>

    <!-- 编辑弹窗 -->
    <Modal
      v-model:open="editVisible"
      title="编辑公告"
      :confirm-loading="editSubmitting"
      :ok-button-props="{ disabled: editFormReadOnly }"
      :width="820"
      ok-text="保存"
      cancel-text="取消"
      @ok="onEditOk"
    >
      <div v-if="editLoading" class="py-8 text-center text-gray-400">
        加载详情中…
      </div>
      <div v-else class="grid gap-4 md:grid-cols-2">
        <div class="md:col-span-2">
          <div class="mb-1 text-sm">
            标题<span class="text-red-500"> *</span>
          </div>
          <Input v-model:value="editForm.title" />
        </div>
        <div class="md:col-span-2">
          <div class="mb-1 text-sm">封面图</div>
          <AnnouncementCoverUpload
            v-model="editForm.coverImageUrl"
            :business-id="editId ?? undefined"
          />
        </div>
        <div class="md:col-span-2">
          <div class="mb-1 text-sm">正文</div>
          <AnnouncementRichTextEditor v-model="editForm.content" />
        </div>
        <div>
          <div class="mb-1 text-sm">状态</div>
          <Select
            v-model:value="editForm.status"
            :options="editStatusOptions"
          />
        </div>
        <div>
          <div class="mb-1 text-sm">是否置顶</div>
          <Switch v-model:checked="editForm.isTop" />
          <span class="ml-3 text-sm">
            排序值
            <InputNumber
              v-model:value="editForm.topSort"
              :min="0"
              class="ml-2"
            />
          </span>
        </div>
        <div>
          <div class="mb-1 text-sm">生效起始</div>
          <DatePicker
            v-model:value="editForm.publishAt"
            show-time
            class="w-full"
          />
        </div>
        <div>
          <div class="mb-1 text-sm">失效时间</div>
          <DatePicker
            v-model:value="editForm.expireAt"
            show-time
            class="w-full"
          />
        </div>
        <div class="md:col-span-2">
          <div class="mb-1 text-sm">可见部门（多选）</div>
          <TreeSelect
            v-model:value="editForm.deptIds"
            :tree-data="deptTreeData as any"
            tree-checkable
            tree-default-expand-all
            multiple
            allow-clear
            class="w-full"
          />
        </div>
        <div class="md:col-span-2">
          <div class="mb-1 text-sm">备注</div>
          <Input.TextArea v-model:value="editForm.remark" :rows="2" />
        </div>
      </div>
    </Modal>

    <!-- 详情抽屉 -->
    <Drawer
      v-model:open="detailVisible"
      :title="detailItem?.title || '公告详情'"
      :width="720"
      placement="right"
    >
      <div v-if="detailLoading" class="py-8 text-center text-gray-400">
        加载中…
      </div>
      <div v-else-if="detailItem" class="space-y-4">
        <div v-if="detailItem.coverImageUrl">
          <img
            :src="detailItem.coverImageUrl"
            class="max-h-64 w-full rounded-sm object-cover"
            alt="cover"
          />
        </div>
        <div class="text-sm text-slate-500">
          状态：<component :is="renderStatusTag(detailItem.status)" />
          <span class="ml-4">已读人数：{{ detailItem.readCount ?? 0 }}</span>
          <span v-if="detailItem.publishAt" class="ml-4">
            生效：{{ formatAdminDateTime(detailItem.publishAt) }}
          </span>
          <span v-if="detailItem.expireAt" class="ml-4">
            失效：{{ formatAdminDateTime(detailItem.expireAt) }}
          </span>
        </div>
        <!-- 后端已做 XSS 过滤，前端可安全 v-html -->
        <div
          class="prose max-w-none"
          v-html="detailItem.content || '<p class=\'text-gray-400\'>无正文</p>'"
        ></div>
        <div class="border-t border-slate-200 pt-4">
          <AnnouncementAttachmentSection
            :announcement-id="detailItem.announcementId"
          />
        </div>
      </div>
    </Drawer>
  </AdminPageShell>
</template>
