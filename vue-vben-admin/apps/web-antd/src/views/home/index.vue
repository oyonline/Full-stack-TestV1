<script lang="ts" setup>
import { computed } from 'vue';
import { useRouter } from 'vue-router';

import { preferences } from '@vben/preferences';
import { useUserStore } from '@vben/stores';

import { Button } from 'ant-design-vue';

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

function openEntry(path: string) {
  router.push(path);
}
</script>

<template>
  <div
    class="min-h-full bg-[linear-gradient(180deg,#f5f7fb_0%,#eef3f8_100%)] p-6"
  >
    <div class="mx-auto max-w-7xl space-y-6">
      <!-- 欢迎头部 -->
      <section class="app-radius-panel bg-white px-6 py-5 shadow-sm">
        <div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
          <div>
            <h1 class="text-lg font-medium text-slate-800">
              欢迎回来，{{ displayName }}
            </h1>
            <p class="mt-1 text-sm text-slate-400">{{ todayText }}</p>
          </div>
          <div class="text-sm text-slate-400">
            {{ preferences.app.name }}
          </div>
        </div>
      </section>

      <!-- 首页内容开发中占位 -->
      <section
        class="app-radius-panel flex min-h-[400px] flex-col items-center justify-center bg-white py-16 shadow-sm"
      >
        <div class="text-center">
          <!-- 开发中图标 -->
          <div
            class="mx-auto mb-6 flex size-20 items-center justify-center rounded-full bg-amber-50 text-4xl"
          >
            🚧
          </div>

          <!-- 主标题 -->
          <h2 class="text-xl font-semibold text-slate-800">
            首页功能开发中
          </h2>

          <!-- 副标题 -->
          <p class="mt-3 text-slate-500">
            预计上线时间：<span class="text-amber-600">待定</span>
          </p>

          <!-- 提示语 -->
          <p class="mx-auto mt-6 max-w-md text-sm text-slate-400">
            您可以通过左侧菜单访问系统功能，首页 Dashboard 正在规划中...
          </p>

          <!-- 快捷跳转 -->
          <div class="mt-8">
            <Button type="primary" @click="openEntry('/admin/sys-user')">
              前往用户管理
            </Button>
          </div>
        </div>
      </section>
    </div>
  </div>
</template>
