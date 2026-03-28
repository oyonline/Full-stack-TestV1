<script setup lang="ts">
import type { UserInfo } from '@vben/types';

import { computed, onBeforeUnmount, ref, watch } from 'vue';

import { VCropper } from '@vben/common-ui';
import { Alert, Button, Card, Descriptions, Input, Modal, Skeleton, Tag, message } from 'ant-design-vue';

import { useUserStore } from '@vben/stores';

import UserAvatar from '#/components/user-avatar.vue';
import { updateUserProfile, uploadUserAvatar } from '#/api/core';
import { useAuthStore } from '#/store/auth';
import {
  USER_AVATAR_PRESET_COLORS,
  buildAvatarAcceptAttribute,
  isAvatarFileTypeSupported,
  normalizeAvatarColor,
  resolveUserAvatar,
} from '#/utils/user-avatar';

type ProfileInfo = Partial<UserInfo> & {
  primaryRoleName?: string;
  roleNames?: string[];
};

const authStore = useAuthStore();
const userStore = useUserStore();

const profile = computed<ProfileInfo>(() => userStore.userInfo ?? {});
const introduction = ref('');
const avatarType = ref<'image' | 'letter'>('letter');
const avatarColor = ref('');
const avatarImage = ref('');
const saving = ref(false);
const uploadingAvatar = ref(false);
const avatarCropVisible = ref(false);
const avatarCropSource = ref('');
const avatarCropperRef = ref<InstanceType<typeof VCropper> | null>(null);
const avatarInputRef = ref<HTMLInputElement | null>(null);
const localPreviewObjectUrl = ref('');

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
const previewAvatar = computed(() =>
  resolveUserAvatar({
    avatar: avatarImage.value,
    avatarColor: avatarColor.value,
    avatarType: avatarType.value,
    realName: profile.value.realName,
    username: profile.value.username,
  }),
);
const hasUploadedAvatar = computed(() => Boolean((avatarImage.value || '').trim()));
const avatarAccept = buildAvatarAcceptAttribute();

watch(
  () => [
    profile.value.avatar,
    profile.value.avatarColor,
    profile.value.avatarType,
    profile.value.desc,
    profile.value.realName,
    profile.value.username,
  ],
  () => {
    const syncedAvatar = resolveUserAvatar({
      avatar: profile.value.avatar,
      avatarColor: profile.value.avatarColor,
      avatarType: profile.value.avatarType,
      realName: profile.value.realName,
      username: profile.value.username,
    });
    introduction.value = profile.value.desc || '';
    avatarImage.value = profile.value.avatar || '';
    avatarColor.value = normalizeAvatarColor(profile.value.avatarColor) || '';
    avatarType.value = syncedAvatar.mode;
  },
  { immediate: true },
);

function clearLocalPreviewObjectUrl() {
  if (!localPreviewObjectUrl.value) {
    return;
  }
  URL.revokeObjectURL(localPreviewObjectUrl.value);
  localPreviewObjectUrl.value = '';
}

function setLocalAvatarPreview(url: string) {
  clearLocalPreviewObjectUrl();
  localPreviewObjectUrl.value = url;
  avatarImage.value = url;
}

function handleUseLetterAvatar() {
  avatarType.value = 'letter';
}

function handleUseImageAvatar() {
  if (!hasUploadedAvatar.value) {
    message.info('请先上传图片头像');
    return;
  }
  avatarType.value = 'image';
}

function handleSelectPresetColor(color: string) {
  avatarType.value = 'letter';
  avatarColor.value = color;
}

function handleCustomColorInput(event: Event) {
  const target = event.target as HTMLInputElement | null;
  avatarType.value = 'letter';
  avatarColor.value = normalizeAvatarColor(target?.value) || '';
}

function openAvatarFilePicker() {
  avatarInputRef.value?.click();
}

function resetAvatarInput() {
  if (avatarInputRef.value) {
    avatarInputRef.value.value = '';
  }
}

function readFileAsDataUrl(file: File) {
  return new Promise<string>((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve(String(reader.result || ''));
    reader.onerror = () => reject(new Error('头像文件读取失败'));
    reader.readAsDataURL(file);
  });
}

async function handleAvatarFileChange(event: Event) {
  const input = event.target as HTMLInputElement | null;
  const file = input?.files?.[0];
  resetAvatarInput();
  if (!file) {
    return;
  }
  if (!isAvatarFileTypeSupported(file)) {
    message.error('头像仅支持 JPG、PNG、WebP 格式');
    return;
  }
  try {
    avatarCropSource.value = await readFileAsDataUrl(file);
    avatarCropVisible.value = true;
  } catch (error: any) {
    message.error(error?.message || '头像文件读取失败');
  }
}

async function handleConfirmAvatarCrop() {
  const cropper = avatarCropperRef.value;
  if (!cropper) {
    message.error('头像裁剪器尚未准备好');
    return;
  }

  try {
    uploadingAvatar.value = true;
    const blob = await cropper.getCropImage(
      'image/jpeg',
      0.85,
      'blob',
      200,
      200,
    );
    if (!(blob instanceof Blob)) {
      throw new Error('头像裁剪失败');
    }
    if (blob.size > 2 * 1024 * 1024) {
      throw new Error('头像文件大小不能超过 2MB');
    }

    setLocalAvatarPreview(URL.createObjectURL(blob));
    avatarType.value = 'image';

    const uploadFile = new File([blob], 'avatar.jpg', {
      type: 'image/jpeg',
    });
    const data = await uploadUserAvatar(uploadFile, 'avatar.jpg');
    clearLocalPreviewObjectUrl();
    avatarImage.value = data.avatar || '';
    avatarType.value = 'image';
    avatarCropVisible.value = false;
    avatarCropSource.value = '';
    await authStore.fetchUserInfo();
    message.success('头像已更新');
  } catch (error: any) {
    message.error(error?.message || error?.response?.data?.msg || '头像上传失败');
  } finally {
    uploadingAvatar.value = false;
  }
}

function handleCancelAvatarCrop() {
  avatarCropVisible.value = false;
  avatarCropSource.value = '';
}

async function handleSaveProfile() {
  if (saving.value) {
    return;
  }
  try {
    saving.value = true;
    await updateUserProfile({
      introduction: introduction.value.trim(),
      avatarColor: avatarColor.value,
      avatarType: avatarType.value,
    });
    await authStore.fetchUserInfo();
    message.success('个人资料已保存');
  } catch (error: any) {
    message.error(error?.message || error?.response?.data?.msg || '个人资料保存失败');
  } finally {
    saving.value = false;
  }
}

onBeforeUnmount(() => {
  clearLocalPreviewObjectUrl();
});
</script>

<template>
  <div class="flex flex-col gap-4">
    <Alert
      type="info"
      show-icon
      message="个人中心当前只展示用户自维护信息。角色、部门、岗位和权限由管理员统一配置，不在这里修改；头像默认走前端字母头像，可随时切换为自定义图片。"
    />

    <Card title="基本信息">
      <Skeleton :loading="!profile.username" active>
        <Descriptions :column="1" bordered size="middle">
          <Descriptions.Item label="姓名">
            {{ profile.realName || '-' }}
          </Descriptions.Item>
          <Descriptions.Item label="登录账号">
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

    <Card title="头像设置">
      <div class="grid gap-6 lg:grid-cols-[240px_minmax(0,1fr)]">
        <div class="rounded-2xl border border-slate-200 bg-slate-50 p-6">
          <div class="flex flex-col items-center gap-4">
            <UserAvatar
              :avatar="previewAvatar.avatar"
              :avatar-color="previewAvatar.avatarBackgroundColor"
              :avatar-type="avatarType"
              :real-name="profile.realName"
              :username="profile.username"
              class="size-24"
            />
            <div class="text-center">
              <div class="text-sm font-medium text-slate-900">
                {{ profile.realName || profile.username || '当前用户' }}
              </div>
              <div class="mt-1 text-xs text-slate-500">
                当前模式：{{ avatarType === 'image' ? '图片头像' : '字母头像' }}
              </div>
              <div class="mt-1 text-xs text-slate-500">
                背景色：{{ avatarColor || previewAvatar.avatarBackgroundColor }}
              </div>
            </div>
          </div>
        </div>

        <div class="flex flex-col gap-5">
          <div>
            <div class="text-sm font-medium text-slate-900">头像模式</div>
            <div class="mt-3 flex flex-wrap gap-3">
              <Button
                :type="avatarType === 'letter' ? 'primary' : 'default'"
                @click="handleUseLetterAvatar"
              >
                使用字母头像
              </Button>
              <Button
                :disabled="!hasUploadedAvatar"
                :type="avatarType === 'image' ? 'primary' : 'default'"
                @click="handleUseImageAvatar"
              >
                使用图片头像
              </Button>
              <Button :loading="uploadingAvatar" @click="openAvatarFilePicker">
                上传图片头像
              </Button>
            </div>
            <div class="mt-2 text-xs leading-6 text-slate-500">
              支持 JPG、PNG、WebP，前端会先裁剪压缩，再上传到本地开发目录
              <code class="mx-1">static/upload/avatar/&lt;userId&gt;/</code>。
            </div>
          </div>

          <div>
            <div class="text-sm font-medium text-slate-900">字母头像背景色</div>
            <div class="mt-3 flex flex-wrap gap-3">
              <button
                v-for="color in USER_AVATAR_PRESET_COLORS"
                :key="color"
                :aria-label="`头像颜色 ${color}`"
                :class="[
                  'h-9 w-9 rounded-full border-2 transition',
                  (avatarColor || previewAvatar.avatarBackgroundColor) === color
                    ? 'scale-110 border-slate-900'
                    : 'border-white shadow-sm',
                ]"
                :style="{ backgroundColor: color }"
                type="button"
                @click="handleSelectPresetColor(color)"
              />
              <label class="flex items-center gap-2 rounded-full border border-slate-200 px-3 py-2 text-xs text-slate-600">
                自定义
                <input
                  :value="avatarColor || previewAvatar.avatarBackgroundColor"
                  class="h-8 w-8 cursor-pointer rounded border-none bg-transparent p-0"
                  type="color"
                  @input="handleCustomColorInput"
                />
              </label>
            </div>
            <div class="mt-2 text-xs text-slate-500">
              未选择自定义颜色时，会按登录账号稳定映射默认主题色。
            </div>
          </div>
        </div>
      </div>
      <input
        ref="avatarInputRef"
        :accept="avatarAccept"
        class="hidden"
        type="file"
        @change="handleAvatarFileChange"
      />
    </Card>

    <Card title="个人资料">
      <div class="flex flex-col gap-3">
        <Input.TextArea
          v-model:value="introduction"
          :auto-size="{ minRows: 4, maxRows: 6 }"
          :maxlength="255"
          placeholder="请输入个人简介"
          show-count
        />
        <div>
          <Button type="primary" :loading="saving" @click="handleSaveProfile">
            保存资料
          </Button>
        </div>
      </div>
    </Card>

    <Modal
      :confirm-loading="uploadingAvatar"
      :open="avatarCropVisible"
      cancel-text="取消"
      ok-text="裁剪并上传"
      title="裁剪头像"
      width="560px"
      @cancel="handleCancelAvatarCrop"
      @ok="handleConfirmAvatarCrop"
    >
      <div class="flex flex-col gap-3">
        <div class="text-xs leading-6 text-slate-500">
          建议保留头像主体在中央区域，系统会输出 200 x 200 的方形头像用于导航栏、个人中心和列表页统一展示。
        </div>
        <VCropper
          v-if="avatarCropSource"
          ref="avatarCropperRef"
          :img="avatarCropSource"
          aspect-ratio="1:1"
        />
      </div>
    </Modal>
  </div>
</template>
