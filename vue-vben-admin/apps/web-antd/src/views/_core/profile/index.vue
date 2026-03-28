<script setup lang="ts">
import { computed, ref } from 'vue';

import { Profile } from '@vben/common-ui';
import { useUserStore } from '@vben/stores';

import ProfileBase from './base-setting.vue';
import ProfileNotificationSetting from './notification-setting.vue';
import ProfilePasswordSetting from './password-setting.vue';
import ProfileSecuritySetting from './security-setting.vue';
import { resolveUserAvatar } from '#/utils/user-avatar';

const userStore = useUserStore();

const tabsValue = ref<string>('basic');
const profileAvatar = computed(() =>
  resolveUserAvatar({
    avatar: userStore.userInfo?.avatar,
    avatarColor: userStore.userInfo?.avatarColor,
    avatarType: userStore.userInfo?.avatarType,
    realName: userStore.userInfo?.realName,
    username: userStore.userInfo?.username,
  }),
);

const tabs = ref([
  {
    label: '基本设置',
    value: 'basic',
  },
  {
    label: '安全设置',
    value: 'security',
  },
  {
    label: '修改密码',
    value: 'password',
  },
  {
    label: '新消息提醒',
    value: 'notice',
  },
]);
</script>
<template>
  <Profile
    v-model:model-value="tabsValue"
    :avatar-background-color="profileAvatar.avatarBackgroundColor"
    :avatar-text="profileAvatar.avatarText"
    title="个人中心"
    :user-info="userStore.userInfo"
    :tabs="tabs"
  >
    <template #content>
      <ProfileBase v-if="tabsValue === 'basic'" />
      <ProfileSecuritySetting v-if="tabsValue === 'security'" />
      <ProfilePasswordSetting v-if="tabsValue === 'password'" />
      <ProfileNotificationSetting v-if="tabsValue === 'notice'" />
    </template>
  </Profile>
</template>
