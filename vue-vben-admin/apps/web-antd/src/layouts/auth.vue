<script lang="ts" setup>
import { computed } from 'vue';

import { AuthPageLayout } from '@vben/layouts';
import { preferences } from '@vben/preferences';

import { $t } from '#/locales';
import { systemDisplayState } from '#/utils/system-settings';

const appName = computed(() => preferences.app.name);
const logo = computed(() => preferences.logo.source);
const logoDark = computed(() => preferences.logo.sourceDark);
const logoPlaceholderBgColor = computed(
  () => preferences.logo.placeholderBgColor || '#1d4ed8',
);
const pageTitle = computed(
  () => systemDisplayState.loginTitle || $t('authentication.pageTitle'),
);
const pageDescription = computed(
  () =>
    systemDisplayState.loginDescription || $t('authentication.pageDesc'),
);
</script>

<template>
  <AuthPageLayout
    :app-name="appName"
    :logo="logo"
    :logo-dark="logoDark"
    :logo-placeholder-bg-color="logoPlaceholderBgColor"
    :page-description="pageDescription"
    :page-title="pageTitle"
  >
    <!-- 自定义工具栏 -->
    <!-- <template #toolbar></template> -->
  </AuthPageLayout>
</template>
