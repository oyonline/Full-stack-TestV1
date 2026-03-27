<script lang="ts" setup>
import { computed, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import { AuthenticationLoginExpiredModal } from '@vben/common-ui';
import { useWatermark } from '@vben/hooks';
import { message } from 'ant-design-vue';
import { CircleHelp } from '@vben/icons';
import {
  BasicLayout,
  LockScreen,
  Notification,
  UserDropdown,
} from '@vben/layouts';
import { preferences } from '@vben/preferences';
import { useAccessStore, useUserStore } from '@vben/stores';

import { $t } from '#/locales';
import { useAuthStore } from '#/store';
import LoginForm from '#/views/_core/authentication/login.vue';

const router = useRouter();
const route = useRoute();
const userStore = useUserStore();
const authStore = useAuthStore();
const accessStore = useAccessStore();
const { destroyWatermark, updateWatermark } = useWatermark();
const placeholderNotifications = ref([]);
const showDot = computed(() => false);

const profileDisplayName = computed(
  () =>
    userStore.userInfo?.realName?.trim() ||
    userStore.userInfo?.username?.trim() ||
    'User',
);

const profileDescription = computed(
  () =>
    userStore.userInfo?.username?.trim() ||
    '当前账号',
);

const profileAvatarText = computed(() => {
  const username = userStore.userInfo?.username?.trim();
  if (!username) {
    return 'U';
  }
  return username.slice(0, 1).toUpperCase();
});

const profileAvatarColor = computed(() => {
  const seed = userStore.userInfo?.username?.trim() || 'user';
  const palette = [
    '#2563eb',
    '#0f766e',
    '#b45309',
    '#7c3aed',
    '#be123c',
    '#0369a1',
    '#15803d',
    '#9333ea',
  ];
  const hash = Array.from(seed).reduce((total, char) => {
    return total + char.charCodeAt(0);
  }, 0);
  return palette[hash % palette.length];
});

async function handleOpenProfile() {
  if (route.path === '/profile') {
    message.info('已在个人中心');
    return;
  }
  await router.push('/profile');
}

const menus = computed(() => [
  {
    handler: handleOpenProfile,
    icon: 'lucide:user',
    text: $t('page.auth.profile'),
  },
  {
    handler: () => {
      message.info('问题与帮助入口建设中');
    },
    icon: CircleHelp,
    text: $t('ui.widgets.qa'),
  },
]);

async function handleLogout() {
  await authStore.logout(false);
}

watch(
  () => ({
    enable: preferences.app.watermark,
    content: preferences.app.watermarkContent,
  }),
  async ({ enable, content }) => {
    if (enable) {
      await updateWatermark({
        content:
          content ||
          `${userStore.userInfo?.username} - ${userStore.userInfo?.realName}`,
      });
    } else {
      destroyWatermark();
    }
  },
  {
    immediate: true,
  },
);
</script>

<template>
  <BasicLayout @clear-preferences-and-logout="handleLogout">
    <template #user-dropdown>
      <UserDropdown
        :avatar="userStore.userInfo?.avatar || ''"
        :avatar-background-color="profileAvatarColor"
        :avatar-text="profileAvatarText"
        :menus
        :text="profileDisplayName"
        :description="profileDescription"
        @logout="handleLogout"
      />
    </template>
    <template #notification>
      <Notification
        placeholder-only
        placeholder-title="通知中心建设中"
        placeholder-description="当前右上角提示图标仅保留通知中心入口占位，不承载真实审批待办数据。审批任务请前往“流程中心 / 我的待办”查看。"
        :dot="showDot"
        :notifications="placeholderNotifications"
      />
    </template>
    <template #extra>
      <AuthenticationLoginExpiredModal
        v-model:open="accessStore.loginExpired"
        :avatar="userStore.userInfo?.avatar"
      >
        <LoginForm />
      </AuthenticationLoginExpiredModal>
    </template>
    <template #lock-screen>
      <LockScreen :avatar="userStore.userInfo?.avatar" @to-login="handleLogout" />
    </template>
  </BasicLayout>
</template>
