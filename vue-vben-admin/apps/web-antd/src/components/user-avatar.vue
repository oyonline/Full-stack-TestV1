<script setup lang="ts">
import type { CSSProperties } from 'vue';

import { computed } from 'vue';

import { VbenAvatar } from '@vben/common-ui';

import { resolveUserAvatar } from '#/utils/user-avatar';

type ClassType = Array<object | string> | object | string;

interface Props {
  avatar?: string;
  avatarColor?: string;
  avatarType?: string;
  class?: ClassType;
  dot?: boolean;
  dotClass?: ClassType;
  fit?: 'contain' | 'cover' | 'fill' | 'none' | 'scale-down';
  realName?: string;
  size?: number;
  username?: string;
}

defineOptions({
  name: 'UserAvatar',
});

const props = withDefaults(defineProps<Props>(), {
  avatar: '',
  avatarColor: '',
  avatarType: '',
  class: '',
  dot: false,
  dotClass: '',
  fit: 'cover',
  realName: '',
  size: undefined,
  username: '',
});

const resolvedAvatar = computed(() =>
  resolveUserAvatar({
    avatar: props.avatar,
    avatarColor: props.avatarColor,
    avatarType: props.avatarType,
    realName: props.realName,
    username: props.username,
  }),
);

const fallbackStyle = computed<CSSProperties>(() => ({
  backgroundColor: resolvedAvatar.value.avatarBackgroundColor,
  color: '#ffffff',
}));
</script>

<template>
  <VbenAvatar
    :alt="resolvedAvatar.avatarText"
    :class="props.class"
    :dot="props.dot"
    :dot-class="props.dotClass"
    :fallback-style="fallbackStyle"
    :fit="props.fit"
    :size="props.size"
    :src="resolvedAvatar.avatar"
  />
</template>
