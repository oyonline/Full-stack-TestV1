<script setup lang="ts">
import type {
  AvatarFallbackProps,
  AvatarImageProps,
  AvatarRootProps,
} from 'reka-ui';

import type { CSSProperties } from 'vue';

import type { ClassType } from '@vben-core/typings';

import { computed } from 'vue';

import { cn } from '@vben-core/shared/utils';

import { AvatarFallback, AvatarImage } from '../../ui';

import { AvatarRoot } from 'reka-ui';

interface Props extends AvatarFallbackProps, AvatarImageProps, AvatarRootProps {
  alt?: string;
  class?: ClassType;
  dot?: boolean;
  dotClass?: ClassType;
  fallbackClass?: ClassType;
  fallbackStyle?: CSSProperties;
  fit?: 'contain' | 'cover' | 'fill' | 'none' | 'scale-down';
  size?: number;
}

defineOptions({
  inheritAttrs: false,
});

const props = withDefaults(defineProps<Props>(), {
  alt: 'avatar',
  as: 'button',
  dot: false,
  dotClass: 'bg-green-500',
  fit: 'cover',
});

const imageStyle = computed<CSSProperties>(() => {
  const { fit } = props;
  if (fit) {
    return {
      display: 'block',
      height: '100%',
      objectFit: fit,
      objectPosition: 'center',
      width: '100%',
    };
  }
  return {};
});

const text = computed(() => {
  return props.alt.slice(-2).toUpperCase();
});

const wrapperClass = computed(() =>
  cn(
    'relative flex aspect-square shrink-0 items-center justify-center overflow-hidden rounded-full',
    props.class,
  ),
);

const fallbackTextStyle = computed<CSSProperties>(() => ({
  // Keep the fallback initials visually proportional to the avatar diameter.
  fontSize:
    props.size !== undefined && props.size > 0
      ? `${Math.max(12, Math.round(props.size * 0.35))}px`
      : 'clamp(0.75rem, 35cqw, 2rem)',
  fontWeight: 600,
  letterSpacing: '-0.02em',
  lineHeight: 1,
}));

const mergedFallbackStyle = computed<CSSProperties>(() => ({
  ...props.fallbackStyle,
  ...fallbackTextStyle.value,
}));

const rootStyle = computed<CSSProperties>(() => ({
  containerType: 'inline-size',
  ...(props.size !== undefined && props.size > 0
    ? {
        height: `${props.size}px`,
        width: `${props.size}px`,
      }
    : {}),
}));
</script>

<template>
  <div :class="wrapperClass" :style="rootStyle">
    <AvatarRoot :as="props.as" :as-child="props.asChild" class="flex size-full rounded-[inherit]">
      <AvatarImage :alt="alt" :src="src" :style="imageStyle" />
      <AvatarFallback :class="fallbackClass" :style="mergedFallbackStyle">
        {{ text }}
      </AvatarFallback>
    </AvatarRoot>
    <span
      v-if="dot"
      :class="dotClass"
      class="border-background absolute right-0 bottom-0 size-3 rounded-full border-2"
    >
    </span>
  </div>
</template>
