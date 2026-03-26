<script lang="ts" setup>
import type { FormSchemaJson } from '#/api/core/gen';

import { computed, onBeforeUnmount, onMounted, ref } from 'vue';

import {
  Alert,
  Button,
  Input,
  Modal,
  Space,
  Tag,
  message,
} from 'ant-design-vue';

import AdminPageShell from '#/components/admin/page-shell.vue';

const DRAWING_ITEMS_KEY = 'drawingItems';
const DRAWING_ITEMS_VERSION_KEY = 'DRAWING_ITEMS_VERSION';
const FORM_CONF_KEY = 'formConf';
const ID_GLOBAL_KEY = 'idGlobal';
const TREE_NODE_ID_KEY = 'treeNodeId';
const FORM_GENERATOR_VERSION = '1.1';

const iframeRef = ref<HTMLIFrameElement>();
const iframeLoaded = ref(false);
const bridgeAvailable = ref(false);
const bridgeError = ref('');
const frameTimedOut = ref(false);
const importModalOpen = ref(false);
const importPayload = ref('');

const iframeSrc = computed(() => '/form-generator/');
let frameLoadTimer: null | number = null;

function clearFrameTimer() {
  if (frameLoadTimer) {
    clearTimeout(frameLoadTimer);
    frameLoadTimer = null;
  }
}

function armFrameTimeout() {
  clearFrameTimer();
  iframeLoaded.value = false;
  frameTimedOut.value = false;
  frameLoadTimer = window.setTimeout(() => {
    if (iframeLoaded.value) {
      return;
    }
    frameTimedOut.value = true;
    bridgeAvailable.value = false;
    bridgeError.value =
      '表单构建器加载超时。请确认后端 /form-generator 静态资源可访问，且浏览器没有拦截构建器依赖的外链资源。';
  }, 15000);
}

function getFrameStorage() {
  const contentWindow = iframeRef.value?.contentWindow;
  if (!contentWindow) {
    throw new Error('表单构建器尚未完成加载');
  }
  return contentWindow.localStorage;
}

function ensureBridgeAvailable() {
  try {
    const storage = getFrameStorage();
    storage.getItem(DRAWING_ITEMS_VERSION_KEY);
    bridgeAvailable.value = true;
    bridgeError.value = '';
    frameTimedOut.value = false;
    return storage;
  } catch {
    bridgeAvailable.value = false;
    bridgeError.value =
      '当前无法直接读取构建器内部草稿。请确认开发环境通过同域 /form-generator 代理访问，并检查 form-generator 外链资源是否加载成功。';
    return null;
  }
}

function normalizeSchemaPayload(payload: string): FormSchemaJson {
  const parsed = JSON.parse(payload) as FormSchemaJson;
  if (!parsed || typeof parsed !== 'object') {
    throw new Error('导入内容必须是对象');
  }
  if (!Array.isArray(parsed.drawingItems)) {
    throw new Error('schema.drawingItems 必须是数组');
  }
  if (!parsed.formConf || typeof parsed.formConf !== 'object') {
    throw new Error('schema.formConf 必须是对象');
  }
  return parsed;
}

function readSchema(): FormSchemaJson {
  const storage = ensureBridgeAvailable();
  if (!storage) {
    throw new Error(bridgeError.value || '无法连接表单构建器');
  }

  return {
    drawingItems: JSON.parse(storage.getItem(DRAWING_ITEMS_KEY) || '[]'),
    formConf: JSON.parse(storage.getItem(FORM_CONF_KEY) || '{}'),
    idGlobal: storage.getItem(ID_GLOBAL_KEY) || undefined,
    treeNodeId: storage.getItem(TREE_NODE_ID_KEY) || undefined,
    version:
      storage.getItem(DRAWING_ITEMS_VERSION_KEY) || FORM_GENERATOR_VERSION,
  };
}

function writeSchema(schema: FormSchemaJson) {
  const storage = ensureBridgeAvailable();
  if (!storage) {
    throw new Error(bridgeError.value || '无法连接表单构建器');
  }

  if (!Array.isArray(schema.drawingItems)) {
    throw new Error('schema.drawingItems 必须是数组');
  }
  if (!schema.formConf || typeof schema.formConf !== 'object') {
    throw new Error('schema.formConf 必须是对象');
  }

  storage.setItem(
    DRAWING_ITEMS_VERSION_KEY,
    schema.version || FORM_GENERATOR_VERSION,
  );
  storage.setItem(DRAWING_ITEMS_KEY, JSON.stringify(schema.drawingItems));
  storage.setItem(FORM_CONF_KEY, JSON.stringify(schema.formConf));
  storage.setItem(ID_GLOBAL_KEY, schema.idGlobal || '100');
  storage.setItem(TREE_NODE_ID_KEY, schema.treeNodeId || '100');
  iframeRef.value?.contentWindow?.location.reload();
}

function refreshBuilder() {
  armFrameTimeout();
  iframeRef.value?.contentWindow?.location.reload();
}

async function copySchema() {
  try {
    const schema = readSchema();
    await navigator.clipboard.writeText(JSON.stringify(schema, null, 2));
    message.success('表单 schema 已复制到剪贴板');
  } catch (error: unknown) {
    const err = error as { message?: string };
    message.error(err.message || '复制 schema 失败');
  }
}

function downloadSchema() {
  try {
    const schema = readSchema();
    const blob = new Blob([JSON.stringify(schema, null, 2)], {
      type: 'application/json;charset=utf-8',
    });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = 'form-schema.json';
    link.click();
    URL.revokeObjectURL(url);
    message.success('表单 schema 已开始下载');
  } catch (error: unknown) {
    const err = error as { message?: string };
    message.error(err.message || '下载 schema 失败');
  }
}

function openImportModal() {
  importPayload.value = '';
  importModalOpen.value = true;
}

function importSchema() {
  try {
    const parsed = normalizeSchemaPayload(importPayload.value);
    writeSchema(parsed);
    importModalOpen.value = false;
    message.success('schema 已导入，构建器正在刷新');
  } catch (error: unknown) {
    const err = error as { message?: string };
    message.error(err.message || '导入 schema 失败，请检查 JSON 格式');
  }
}

function handleFileImport(event: Event) {
  const input = event.target as HTMLInputElement | null;
  const file = input?.files?.[0];
  if (!file) return;

  file
    .text()
    .then((text) => {
      importPayload.value = text;
    })
    .catch(() => {
      message.error('读取 schema 文件失败');
    })
    .finally(() => {
      if (input) {
        input.value = '';
      }
    });
}

function handleFrameLoad() {
  clearFrameTimer();
  iframeLoaded.value = true;
  ensureBridgeAvailable();
}

function handleFrameError() {
  clearFrameTimer();
  iframeLoaded.value = false;
  bridgeAvailable.value = false;
  bridgeError.value =
    '表单构建器加载失败。请确认后端静态页可访问，且浏览器没有拦截外链脚本或样式。';
}

onMounted(() => {
  armFrameTimeout();
});

onBeforeUnmount(() => {
  clearFrameTimer();
});
</script>

<template>
  <AdminPageShell>
    <template #eyebrow>DEV TOOLS</template>
    <template #title>表单构建</template>
    <template #description>
      当前页通过同域 iframe 承接后端 form-generator，并额外提供 schema 导出、复制和导入能力。第一版不做数据库持久化，只围绕本地 schema 交付。
    </template>

    <template #toolbar>
      <div class="flex flex-wrap items-center gap-2">
        <Tag color="blue">Schema Bridge</Tag>
        <span class="text-sm text-slate-500">
          直接读写构建器草稿：`drawingItems`、`formConf`、`idGlobal`、`treeNodeId`。
        </span>
      </div>
      <Space wrap>
        <Button @click="refreshBuilder">刷新构建器</Button>
        <Button :disabled="!bridgeAvailable" @click="downloadSchema">
          下载 Schema
        </Button>
        <Button :disabled="!bridgeAvailable" @click="copySchema">
          复制 Schema
        </Button>
        <Button type="primary" @click="openImportModal">导入 Schema</Button>
      </Space>
    </template>

    <Alert
      v-if="bridgeError"
      type="warning"
      show-icon
      :message="bridgeError"
      class="mb-4"
    />
    <Alert
      v-else-if="iframeLoaded"
      type="success"
      show-icon
      message="构建器与 schema 桥接已就绪，可以直接导入、导出和复制。"
      class="mb-4"
    />
    <Alert
      v-else
      type="info"
      show-icon
      message="构建器加载中，完成后会自动检测 schema 桥接可用性。"
      class="mb-4"
    />

    <section
      class="app-radius-panel overflow-hidden border border-slate-200 bg-white shadow-sm"
    >
      <iframe
        ref="iframeRef"
        :src="iframeSrc"
        class="h-[calc(100vh-19rem)] min-h-[720px] w-full border-0 bg-white"
        @error="handleFrameError"
        @load="handleFrameLoad"
      />
    </section>

    <Modal
      v-model:open="importModalOpen"
      title="导入 Form Schema JSON"
      ok-text="导入"
      cancel-text="取消"
      width="760px"
      @ok="importSchema"
    >
      <div class="space-y-4">
        <Alert
          type="info"
          show-icon
          message="导入后会覆盖构建器当前草稿，并自动刷新 iframe。"
        />
        <div class="flex flex-wrap items-center gap-3">
          <label
            class="inline-flex cursor-pointer items-center text-sm text-blue-600 hover:text-blue-500"
          >
            <input
              class="hidden"
              type="file"
              accept="application/json,.json"
              @change="handleFileImport"
            />
            从 JSON 文件读取
          </label>
          <span class="text-xs text-slate-500">
            支持直接粘贴或选择之前导出的 schema 文件。
          </span>
        </div>
        <Input.TextArea
          v-model:value="importPayload"
          :auto-size="{ minRows: 16, maxRows: 22 }"
          placeholder="请粘贴完整的 FormSchemaJson"
        />
        <p class="text-xs text-slate-500">
          最少需要包含 `drawingItems` 数组和 `formConf` 对象，否则导入会被拒绝。
        </p>
      </div>
    </Modal>
  </AdminPageShell>
</template>
