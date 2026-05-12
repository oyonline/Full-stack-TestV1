<script lang="ts" setup>
import type { MenuRecordRaw } from '@vben/types';

import type { WorkflowTaskItem } from '#/api/core';

import { computed, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';

import { IconifyIcon } from '@vben/icons';
import { preferences } from '@vben/preferences';
import { useAccessStore, useUserStore } from '@vben/stores';

import { Alert, Button, Skeleton, Tag } from 'ant-design-vue';

import {
  getAppConfigApi,
  getWorkflowStartedPage,
  getWorkflowTodoPage,
} from '#/api/core';
import {
  getWorkflowStatusLabel,
  workflowStatusColors,
} from '#/views/platform/workflow/constants';

const router = useRouter();
const userStore = useUserStore();
const accessStore = useAccessStore();

const todoLoading = ref(false);
const todoList = ref<WorkflowTaskItem[]>([]);
const todoTotal = ref(0);
const startedTotal = ref(0);
const appConfigLoading = ref(false);
const appName = ref('');
const appVersion = ref(import.meta.env.VITE_APP_VERSION || '');

const todayText = computed(() =>
  new Intl.DateTimeFormat('zh-CN', { dateStyle: 'full' }).format(new Date()),
);

const greeting = computed(() => {
  const h = new Date().getHours();
  if (h < 6) return '夜深了';
  if (h < 12) return '早上好';
  if (h < 14) return '中午好';
  if (h < 18) return '下午好';
  return '晚上好';
});

const displayName = computed(
  () => userStore.userInfo?.realName || userStore.userInfo?.username || '同事',
);

const roleLabel = computed(() => {
  const info = userStore.userInfo as null | Record<string, any>;
  if (info?.primaryRoleName) return String(info.primaryRoleName);
  const roles = info?.roles;
  if (Array.isArray(roles) && roles.length > 0) return String(roles[0]);
  return '未分配角色';
});

// 模块快捷入口：来自后端菜单（accessMenus），只取顶层 + 一级子项
interface ModuleEntry {
  name: string;
  path: string;
  icon?: string;
  parentName: string;
}

const moduleEntries = computed<ModuleEntry[]>(() => {
  const result: ModuleEntry[] = [];
  for (const top of accessStore.accessMenus as MenuRecordRaw[]) {
    if (top.show === false) continue;
    const children = (top.children ?? []).filter(
      (c: MenuRecordRaw) => c.show !== false && c.path,
    );
    if (children.length === 0) {
      // 没有子项的顶层菜单：自身作为入口
      if (top.path && top.path !== '/home') {
        result.push({
          name: top.name,
          path: top.path,
          icon: typeof top.icon === 'string' ? top.icon : undefined,
          parentName: top.name,
        });
      }
      continue;
    }
    for (const child of children) {
      if (!child.path || child.path === '/home') continue;
      result.push({
        name: child.name,
        path: child.path,
        icon:
          (typeof child.icon === 'string' ? child.icon : undefined) ||
          (typeof top.icon === 'string' ? top.icon : undefined),
        parentName: top.name,
      });
    }
  }
  return result.slice(0, 12);
});

function go(path: string) {
  router.push(path);
}

function formatTime(value?: string) {
  if (!value) return '-';
  const d = new Date(value);
  if (Number.isNaN(d.getTime())) return value;
  return new Intl.DateTimeFormat('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  }).format(d);
}

async function loadTodos() {
  todoLoading.value = true;
  try {
    const [todo, started] = await Promise.all([
      getWorkflowTodoPage({ pageIndex: 1, pageSize: 5 }),
      getWorkflowStartedPage({ pageIndex: 1, pageSize: 1 }),
    ]);
    todoList.value = todo.list ?? [];
    todoTotal.value = todo.count ?? 0;
    startedTotal.value = started.count ?? 0;
  } catch {
    todoList.value = [];
    todoTotal.value = 0;
    startedTotal.value = 0;
  } finally {
    todoLoading.value = false;
  }
}

async function loadAppConfig() {
  appConfigLoading.value = true;
  try {
    const config = await getAppConfigApi();
    appName.value = config?.sys_app_name?.trim() || preferences.app.name || '';
  } catch {
    appName.value = preferences.app.name || '';
  } finally {
    appConfigLoading.value = false;
  }
}

function openTodoCenter() {
  router.push('/platform/workflow/todo');
}

function openStarted() {
  router.push('/platform/workflow/started');
}

onMounted(() => {
  loadTodos();
  loadAppConfig();
});
</script>

<template>
  <div
    class="min-h-full bg-[linear-gradient(180deg,#f5f7fb_0%,#eef3f8_100%)] p-6"
  >
    <div class="mx-auto max-w-7xl space-y-6">
      <!-- Multica Agent 验证标记 -->
      <Alert
        message="✅ Modified by Multica Agent"
        type="success"
        show-icon
        banner
        class="rounded-md"
      />
      <!-- 欢迎头部 -->
      <section class="app-radius-panel bg-white px-6 py-5 shadow-sm">
        <div
          class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between"
        >
          <div>
            <h1 class="text-lg font-medium text-slate-800">
              {{ greeting }}，{{ displayName }}
            </h1>
            <p class="mt-1 text-sm text-slate-400">
              {{ todayText }} · 当前角色 {{ roleLabel }}
            </p>
          </div>
          <div class="text-sm text-slate-400">
            {{ appName || preferences.app.name }}
          </div>
        </div>
      </section>

      <!-- 统计指标 -->
      <section class="grid grid-cols-1 gap-4 md:grid-cols-3">
        <div
          class="app-radius-panel cursor-pointer bg-white px-5 py-4 shadow-sm transition hover:shadow-sm"
          @click="openTodoCenter"
        >
          <div class="flex items-center justify-between">
            <span class="text-sm text-slate-500">我的待办</span>
            <IconifyIcon icon="lucide:inbox" class="text-lg text-slate-400" />
          </div>
          <div class="mt-3 text-2xl font-semibold text-slate-800">
            {{ todoTotal }}
          </div>
          <div class="mt-1 text-xs text-slate-400">待处理工作流任务</div>
        </div>
        <div
          class="app-radius-panel cursor-pointer bg-white px-5 py-4 shadow-sm transition hover:shadow-sm"
          @click="openStarted"
        >
          <div class="flex items-center justify-between">
            <span class="text-sm text-slate-500">我发起的</span>
            <IconifyIcon icon="lucide:send" class="text-lg text-slate-400" />
          </div>
          <div class="mt-3 text-2xl font-semibold text-slate-800">
            {{ startedTotal }}
          </div>
          <div class="mt-1 text-xs text-slate-400">我发起的工作流实例</div>
        </div>
        <div class="app-radius-panel bg-white px-5 py-4 shadow-sm">
          <div class="flex items-center justify-between">
            <span class="text-sm text-slate-500">可访问模块</span>
            <IconifyIcon
              icon="lucide:layout-grid"
              class="text-lg text-slate-400"
            />
          </div>
          <div class="mt-3 text-2xl font-semibold text-slate-800">
            {{ moduleEntries.length }}
          </div>
          <div class="mt-1 text-xs text-slate-400">基于当前角色的菜单数</div>
        </div>
      </section>

      <div class="grid grid-cols-1 gap-6 lg:grid-cols-3">
        <!-- 我的待办 -->
        <section
          class="app-radius-panel bg-white px-6 py-5 shadow-sm lg:col-span-2"
        >
          <div class="flex items-center justify-between">
            <h2 class="text-base font-medium text-slate-800">最近待办</h2>
            <Button type="link" size="small" @click="openTodoCenter">
              查看全部
            </Button>
          </div>
          <div class="mt-3">
            <Skeleton v-if="todoLoading" active :paragraph="{ rows: 4 }" />
            <div
              v-else-if="todoList.length === 0"
              class="flex h-32 items-center justify-center text-sm text-slate-400"
            >
              暂无待处理任务
            </div>
            <ul v-else class="divide-y divide-slate-100">
              <li
                v-for="item in todoList"
                :key="item.taskId"
                class="flex cursor-pointer items-center justify-between py-3 hover:bg-slate-50"
                @click="openTodoCenter"
              >
                <div class="min-w-0 flex-1">
                  <div class="flex items-center gap-2">
                    <span class="truncate text-sm text-slate-800">
                      {{ item.title || item.definitionName }}
                    </span>
                    <Tag
                      :color="workflowStatusColors[item.status] || 'default'"
                      class="m-0"
                    >
                      {{ getWorkflowStatusLabel(item.status) }}
                    </Tag>
                  </div>
                  <div class="mt-1 text-xs text-slate-400">
                    <span>{{ item.starterName || '-' }}</span>
                    <span class="mx-2">·</span>
                    <span>{{ item.currentNodeName || '-' }}</span>
                    <span class="mx-2">·</span>
                    <span>{{ formatTime(item.taskCreatedAt) }}</span>
                  </div>
                </div>
                <IconifyIcon
                  icon="lucide:chevron-right"
                  class="ml-3 shrink-0 text-slate-300"
                />
              </li>
            </ul>
          </div>
        </section>

        <!-- 系统信息 -->
        <section class="app-radius-panel bg-white px-6 py-5 shadow-sm">
          <h2 class="text-base font-medium text-slate-800">系统信息</h2>
          <Skeleton
            v-if="appConfigLoading"
            class="mt-3"
            active
            :paragraph="{ rows: 3 }"
          />
          <dl v-else class="mt-3 space-y-3 text-sm">
            <div class="flex justify-between">
              <dt class="text-slate-400">系统名称</dt>
              <dd class="text-slate-700">
                {{ appName || preferences.app.name }}
              </dd>
            </div>
            <div class="flex justify-between">
              <dt class="text-slate-400">前端版本</dt>
              <dd class="text-slate-700">{{ appVersion || '-' }}</dd>
            </div>
            <div class="flex justify-between">
              <dt class="text-slate-400">登录账号</dt>
              <dd class="text-slate-700">
                {{ userStore.userInfo?.username || '-' }}
              </dd>
            </div>
            <div class="flex justify-between">
              <dt class="text-slate-400">主题模式</dt>
              <dd class="text-slate-700">
                {{ preferences.theme.mode === 'dark' ? '深色' : '浅色' }}
              </dd>
            </div>
            <div class="flex justify-between">
              <dt class="text-slate-400">权限模式</dt>
              <dd class="text-slate-700">{{ preferences.app.accessMode }}</dd>
            </div>
          </dl>
        </section>
      </div>

      <!-- 平台模块快捷入口 -->
      <section class="app-radius-panel bg-white px-6 py-5 shadow-sm">
        <div class="flex items-center justify-between">
          <h2 class="text-base font-medium text-slate-800">平台模块</h2>
          <span class="text-xs text-slate-400">基于当前角色的可访问菜单</span>
        </div>
        <div
          v-if="moduleEntries.length === 0"
          class="mt-4 flex h-24 items-center justify-center text-sm text-slate-400"
        >
          暂无可访问模块
        </div>
        <div
          v-else
          class="mt-4 grid grid-cols-2 gap-3 sm:grid-cols-3 md:grid-cols-4"
        >
          <button
            v-for="entry in moduleEntries"
            :key="entry.path"
            type="button"
            class="group flex items-center gap-3 rounded-md border border-slate-100 bg-slate-50/60 p-3 text-left transition hover:border-blue-200 hover:bg-blue-50"
            @click="go(entry.path)"
          >
            <span
              class="flex size-9 shrink-0 items-center justify-center rounded-md bg-white text-blue-500 shadow-sm group-hover:text-blue-600"
            >
              <IconifyIcon
                :icon="entry.icon || 'lucide:square-dashed'"
                class="text-lg"
              />
            </span>
            <span class="min-w-0 flex-1">
              <span class="block truncate text-sm text-slate-700">
                {{ entry.name }}
              </span>
              <span class="block truncate text-xs text-slate-400">
                {{ entry.parentName }}
              </span>
            </span>
          </button>
        </div>
      </section>
    </div>
  </div>
</template>
