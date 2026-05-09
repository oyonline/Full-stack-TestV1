<script setup lang="ts">
import { computed, ref } from 'vue';

import { extractInitial, pickForegroundColor } from '@vben-core/shared/utils';

interface Props {
  /**
   * @zh_CN 是否收起文本
   */
  collapsed?: boolean;
  /**
   * @zh_CN 图片加载失败时是否切换到 fallback 占位
   */
  fallbackOnError?: boolean;
  /**
   * @zh_CN Logo 图片适应方式
   */
  fit?: 'contain' | 'cover' | 'fill' | 'none' | 'scale-down';
  /**
   * @zh_CN Logo 跳转地址
   */
  href?: string;
  /**
   * @zh_CN Logo 图片大小
   */
  logoSize?: number;
  /**
   * @zh_CN 占位 Logo 底色(来自 sys_app_logo_placeholder_color)
   */
  placeholderBgColor?: string;
  /**
   * @zh_CN Logo 图标
   */
  src?: string;
  /**
   * @zh_CN 暗色主题 Logo 图标 (可选，若不设置则使用 src)
   */
  srcDark?: string;
  /**
   * @zh_CN 系统名称，用于 fallback 首字提取
   */
  systemName?: string;
  /**
   * @zh_CN Logo 文本
   */
  text: string;
  /**
   * @zh_CN Logo 主题
   */
  theme?: string;
}

defineOptions({
  name: 'VbenLogo',
});

const props = withDefaults(defineProps<Props>(), {
  collapsed: false,
  fallbackOnError: true,
  href: 'javascript:void 0',
  logoSize: 32,
  placeholderBgColor: '#1F2937',
  src: '',
  srcDark: '',
  systemName: '',
  theme: 'light',
  fit: 'cover',
});

const imgFailed = ref(false);

const logoSrc = computed(() => {
  if (props.theme === 'dark' && props.srcDark) {
    return props.srcDark;
  }
  return props.src;
});

const showFallback = computed(
  () => imgFailed.value && props.fallbackOnError,
);

const fallbackInitial = computed(() =>
  extractInitial(props.systemName || props.text),
);

const fallbackFgColor = computed(() =>
  pickForegroundColor(props.placeholderBgColor),
);
</script>

<template>
  <div :class="theme" class="flex h-full items-center text-lg">
    <a
      :class="$attrs.class"
      :href="href"
      class="flex h-full items-center gap-2 overflow-hidden px-3 text-lg leading-normal transition-all duration-500"
    >
      <img
        v-if="logoSrc && !showFallback"
        :alt="text"
        :src="logoSrc"
        :style="{
          width: `${logoSize}px`,
          height: `${logoSize}px`,
          objectFit: fit,
        }"
        class="relative rounded-none bg-transparent"
        @error="imgFailed = true"
      />
      <div
        v-else
        :style="{
          width: `${logoSize}px`,
          height: `${logoSize}px`,
          backgroundColor: placeholderBgColor,
          color: fallbackFgColor,
        }"
        class="flex items-center justify-center rounded-lg text-sm font-semibold shadow-sm"
      >
        {{ fallbackInitial }}
      </div>
      <template v-if="!collapsed">
        <slot name="text">
          <span class="text-foreground truncate font-semibold text-nowrap">
            {{ text }}
          </span>
        </slot>
      </template>
    </a>
  </div>
</template>
