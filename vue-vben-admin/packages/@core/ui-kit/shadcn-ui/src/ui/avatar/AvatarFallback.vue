<script setup lang="ts">
import type { AvatarFallbackProps } from 'reka-ui';

import type { ClassType } from '@vben-core/typings';

import { cn } from '@vben-core/shared/utils';

import { AvatarFallback } from 'reka-ui';

const props = defineProps<AvatarFallbackProps & { class?: ClassType }>();
</script>

<template>
  <AvatarFallback
    v-bind="props"
    :class="
      cn(
        'flex h-full w-full items-center justify-center overflow-hidden rounded-[inherit]',
        props.class,
      )
    "
  >
    <slot></slot>
  </AvatarFallback>
</template>
