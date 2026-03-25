<script lang="ts" setup>
import { computed } from 'vue';
import { useRouter } from 'vue-router';

import { preferences } from '@vben/preferences';
import { useUserStore } from '@vben/stores';

import { Button, Card, Tag } from 'ant-design-vue';

const router = useRouter();
const userStore = useUserStore();

const todayText = computed(() =>
  new Intl.DateTimeFormat('zh-CN', {
    dateStyle: 'full',
  }).format(new Date()),
);

const displayName = computed(
  () => userStore.userInfo?.realName || userStore.userInfo?.username || '同事',
);

const quickEntries = [
  {
    description: '查看账号、角色和部门绑定情况',
    path: '/admin/sys-user',
    title: '用户管理',
  },
  {
    description: '校验权限字符、菜单授权和数据范围',
    path: '/admin/sys-role',
    title: '角色管理',
  },
  {
    description: '维护菜单、按钮权限和组件映射',
    path: '/admin/sys-menu',
    title: '菜单管理',
  },
  {
    description: '检查定时任务配置与执行状态',
    path: '/admin/sys-job',
    title: '定时任务',
  },
  {
    description: '观察系统资源与接口运行状态',
    path: '/admin/sys-server-monitor',
    title: '服务监控',
  },
  {
    description: '维护系统动态参数与运行开关',
    path: '/admin/sys-config/set',
    title: '参数设置',
  },
];

const statusCards = [
  {
    label: '权限模式',
    value: '后端菜单驱动',
  },
  {
    label: '开发链路',
    value: '/api 代理转发',
  },
  {
    label: '登录校验',
    value: '验证码已启用',
  },
];

function openEntry(path: string) {
  router.push(path);
}
</script>

<template>
  <div
    class="min-h-full bg-[linear-gradient(180deg,#f5f7fb_0%,#eef3f8_100%)] p-6"
  >
    <div class="mx-auto max-w-7xl space-y-6">
      <section
        class="overflow-hidden rounded-3xl border border-white/70 bg-[radial-gradient(circle_at_top_left,#1f6feb_0%,#1648c8_32%,#0f172a_100%)] p-8 text-white shadow-[0_24px_80px_rgba(15,23,42,0.18)]"
      >
        <div
          class="flex flex-col gap-6 lg:flex-row lg:items-end lg:justify-between"
        >
          <div class="max-w-3xl">
            <p
              class="text-sm font-medium uppercase tracking-[0.28em] text-white/70"
            >
              {{ preferences.app.name }}
            </p>
            <h1 class="mt-3 text-3xl font-semibold leading-tight lg:text-5xl">
              首页
            </h1>
            <p class="mt-4 text-lg text-white/85">
              {{
                displayName
              }}，欢迎回来。现在先从首页进入，再按需要跳转到系统管理各模块。
            </p>
            <p class="mt-2 text-sm text-white/70">
              {{ todayText }}
            </p>
          </div>

          <div class="grid gap-3 sm:grid-cols-3 lg:min-w-[420px]">
            <div
              v-for="card in statusCards"
              :key="card.label"
              class="rounded-2xl border border-white/15 bg-white/10 px-4 py-4 backdrop-blur"
            >
              <div class="text-xs uppercase tracking-[0.22em] text-white/60">
                {{ card.label }}
              </div>
              <div class="mt-3 text-base font-medium">
                {{ card.value }}
              </div>
            </div>
          </div>
        </div>
      </section>

      <section class="grid gap-6 lg:grid-cols-[1.6fr_1fr]">
        <Card :bordered="false" class="rounded-3xl shadow-sm">
          <template #title>
            <div class="flex items-center justify-between">
              <span class="text-base font-semibold">快捷入口</span>
              <Tag color="blue">Landing</Tag>
            </div>
          </template>

          <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
            <button
              v-for="entry in quickEntries"
              :key="entry.path"
              class="rounded-2xl border border-slate-200 bg-slate-50 px-4 py-4 text-left transition hover:-translate-y-0.5 hover:border-blue-300 hover:bg-white hover:shadow-sm"
              type="button"
              @click="openEntry(entry.path)"
            >
              <div class="text-base font-medium text-slate-900">
                {{ entry.title }}
              </div>
              <div class="mt-2 text-sm leading-6 text-slate-500">
                {{ entry.description }}
              </div>
            </button>
          </div>
        </Card>

        <div class="space-y-6">
          <Card :bordered="false" class="rounded-3xl shadow-sm">
            <template #title>
              <span class="text-base font-semibold">今日建议</span>
            </template>
            <div class="space-y-3 text-sm text-slate-600">
              <div class="rounded-2xl bg-slate-50 px-4 py-3">
                先验证登录、菜单加载和退出流程，确认首页落地后再进入业务页。
              </div>
              <div class="rounded-2xl bg-slate-50 px-4 py-3">
                优先检查用户、角色、菜单三块，它们决定大部分权限链路是否正常。
              </div>
              <div class="rounded-2xl bg-slate-50 px-4 py-3">
                如果发现接口异常，先看浏览器请求是否统一走了 <code>/api</code>。
              </div>
            </div>
          </Card>

          <Card :bordered="false" class="rounded-3xl shadow-sm">
            <template #title>
              <span class="text-base font-semibold">快速开始</span>
            </template>
            <div class="space-y-3">
              <Button
                block
                type="primary"
                @click="openEntry('/admin/sys-user')"
              >
                从用户管理开始
              </Button>
              <Button block @click="openEntry('/admin/sys-menu')">
                打开菜单管理
              </Button>
              <Button block @click="openEntry('/admin/sys-server-monitor')">
                查看服务监控
              </Button>
            </div>
          </Card>
        </div>
      </section>
    </div>
  </div>
</template>
