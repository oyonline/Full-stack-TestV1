<script setup lang="ts">
import type { UserInfo } from '@vben/types';

import { computed, ref, watch } from 'vue';

import { Alert, Button, Card, Descriptions, Input, Skeleton, Tag, message } from 'ant-design-vue';

import { useUserStore } from '@vben/stores';

import { updateUserProfile } from '#/api/core';
import { useAuthStore } from '#/store/auth';

type ProfileInfo = Partial<UserInfo> & {
  primaryRoleName?: string;
  roleNames?: string[];
};

const authStore = useAuthStore();
const userStore = useUserStore();

const profile = computed<ProfileInfo>(() => userStore.userInfo ?? {});
const introduction = ref('');
const saving = ref(false);

const roleNames = computed(() => {
  const roles = Array.isArray(profile.value.roleNames)
    ? profile.value.roleNames
    : Array.isArray(profile.value.roles)
      ? profile.value.roles
      : [];
  return roles.filter(Boolean);
});
const primaryRoleName = computed(
  () =>
    profile.value.primaryRoleName ||
    roleNames.value[0] ||
    '未分配主角色',
);

watch(
  () => profile.value.desc,
  (value) => {
    introduction.value = value || '';
  },
  { immediate: true },
);

async function handleSaveIntroduction() {
  if (saving.value) {
    return;
  }
  try {
    saving.value = true;
    await updateUserProfile({
      introduction: introduction.value.trim(),
    });
    await authStore.fetchUserInfo();
    message.success('个人简介已保存');
  } catch (error: any) {
    message.error(error?.message || error?.response?.data?.msg || '个人简介保存失败');
  } finally {
    saving.value = false;
  }
}
</script>

<template>
  <div class="flex flex-col gap-4">
    <Alert
      type="info"
      show-icon
      message="个人中心当前只展示用户自维护信息。角色、部门、岗位和权限由管理员统一配置，不在这里修改。"
    />

    <Card title="基本信息">
      <Skeleton :loading="!profile.username" active>
        <Descriptions :column="1" bordered size="middle">
          <Descriptions.Item label="姓名">
            {{ profile.realName || '-' }}
          </Descriptions.Item>
          <Descriptions.Item label="用户名">
            {{ profile.username || '-' }}
          </Descriptions.Item>
          <Descriptions.Item label="主角色">
            <Tag color="blue">{{ primaryRoleName }}</Tag>
          </Descriptions.Item>
          <Descriptions.Item label="全部角色">
            <div class="flex flex-wrap gap-2">
              <Tag
                v-for="roleName in roleNames"
                :key="roleName"
                color="default"
              >
                {{ roleName }}
              </Tag>
              <span v-if="roleNames.length === 0">-</span>
            </div>
          </Descriptions.Item>
          <Descriptions.Item label="个人简介">
            {{ profile.desc || '暂未设置个人简介' }}
          </Descriptions.Item>
        </Descriptions>
      </Skeleton>
    </Card>

    <Card title="个人简介">
      <div class="flex flex-col gap-3">
        <Input.TextArea
          v-model:value="introduction"
          :auto-size="{ minRows: 4, maxRows: 6 }"
          :maxlength="255"
          placeholder="请输入个人简介"
          show-count
        />
        <div>
          <Button type="primary" :loading="saving" @click="handleSaveIntroduction">
            保存个人简介
          </Button>
        </div>
      </div>
    </Card>
  </div>
</template>
