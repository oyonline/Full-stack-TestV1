<script lang="ts" setup>
import type { SystemSettingsResponse } from '#/api/core/config';

import { computed, onMounted, reactive, ref } from 'vue';

import { SUPPORT_LANGUAGES } from '@vben/constants';

import {
  Alert,
  Button,
  Card,
  Image,
  Input,
  InputNumber,
  Select,
  Skeleton,
  Space,
  Switch,
  Tabs,
  Tag,
  message,
} from 'ant-design-vue';

import { getSystemSettingsApi, updateSystemSettingsApi } from '#/api/core';
import {
  applySystemSettingsToRuntime,
  createDefaultSystemSettings,
} from '#/utils/system-settings';

function clonePlain<T>(value: T): T {
  return JSON.parse(JSON.stringify(value)) as T;
}

const themeModeOptions = [
  { label: '浅色', value: 'light' },
  { label: '深色', value: 'dark' },
  { label: '跟随系统', value: 'auto' },
];

const themePresetOptions = [
  { label: '默认蓝', value: 'default' },
  { label: '深蓝', value: 'deep-blue' },
  { label: '深绿', value: 'deep-green' },
  { label: '灰色', value: 'gray' },
  { label: '绿色', value: 'green' },
  { label: '中性灰', value: 'neutral' },
  { label: '橙色', value: 'orange' },
  { label: '粉色', value: 'pink' },
  { label: '玫瑰', value: 'rose' },
  { label: '天蓝', value: 'sky-blue' },
  { label: '石板灰', value: 'slate' },
  { label: '紫罗兰', value: 'violet' },
  { label: '黄色', value: 'yellow' },
  { label: '锌灰', value: 'zinc' },
  { label: '自定义', value: 'custom' },
];

const layoutOptions = [
  { label: '侧边导航', value: 'sidebar-nav' },
  { label: '侧边双列', value: 'sidebar-mixed-nav' },
  { label: '顶部导航', value: 'header-nav' },
  { label: '顶部双列', value: 'header-mixed-nav' },
  { label: '顶部+侧边', value: 'header-sidebar-nav' },
  { label: '混合导航', value: 'mixed-nav' },
  { label: '全内容', value: 'full-content' },
];

const contentCompactOptions = [
  { label: '宽版', value: 'wide' },
  { label: '紧凑', value: 'compact' },
];

const headerModeOptions = [
  { label: '固定', value: 'fixed' },
  { label: '静态', value: 'static' },
  { label: '自动', value: 'auto' },
  { label: '自动滚动', value: 'auto-scroll' },
];

const headerAlignOptions = [
  { label: '居左', value: 'start' },
  { label: '居中', value: 'center' },
  { label: '居右', value: 'end' },
];

const navigationStyleOptions = [
  { label: '圆角', value: 'rounded' },
  { label: '朴素', value: 'plain' },
];

const breadcrumbStyleOptions = [
  { label: '普通', value: 'normal' },
  { label: '背景', value: 'background' },
];

const tabbarStyleOptions = [
  { label: 'Chrome', value: 'chrome' },
  { label: 'Plain', value: 'plain' },
  { label: 'Card', value: 'card' },
  { label: 'Brisk', value: 'brisk' },
];

const transitionOptions = [
  { label: 'fade', value: 'fade' },
  { label: 'fade-slide', value: 'fade-slide' },
  { label: 'fade-up', value: 'fade-up' },
  { label: 'fade-down', value: 'fade-down' },
];

const widgetItems = [
  { key: 'globalSearch', label: '全局搜索' },
  { key: 'themeToggle', label: '主题切换' },
  { key: 'languageToggle', label: '语言切换' },
  { key: 'fullscreen', label: '全屏' },
  { key: 'notification', label: '通知' },
  { key: 'lockScreen', label: '锁屏' },
  { key: 'refresh', label: '刷新' },
  { key: 'sidebarToggle', label: '侧栏切换' },
  { key: 'timezone', label: '时区' },
] as const;

const shortcutItems = [
  { key: 'enable', label: '启用快捷键总开关' },
  { key: 'globalSearch', label: '搜索快捷键' },
  { key: 'globalLogout', label: '注销快捷键' },
  { key: 'globalLockScreen', label: '锁屏快捷键' },
] as const;

const loading = ref(false);
const saving = ref(false);
const activeTab = ref('appearance');

const settings = reactive<SystemSettingsResponse>(
  createDefaultSystemSettings(),
);
const initialSnapshot = ref<SystemSettingsResponse>(
  clonePlain(createDefaultSystemSettings()),
);

const logoPreviewUrl = computed(() => {
  const value = settings.branding.appLogo.trim();
  return /^https?:\/\//.test(value) ? value : '';
});

const logoPlaceholderColor = computed(
  () => settings.branding.appLogoPlaceholderColor.trim() || '#1d4ed8',
);

const logoPlaceholderText = computed(() => {
  const name = settings.branding.appName.trim().replace(/\s+/g, '');
  if (!name) {
    return 'S';
  }

  const asciiGroups = name.match(/[A-Za-z0-9]+/g);
  const firstAsciiGroup = asciiGroups?.[0];
  if (firstAsciiGroup?.[0]) {
    return firstAsciiGroup[0].toUpperCase();
  }

  return name[0]?.toUpperCase() || 'S';
});

const hasUnsavedChanges = computed(
  () => JSON.stringify(settings) !== JSON.stringify(initialSnapshot.value),
);

function applySettings(payload: SystemSettingsResponse) {
  const normalized = clonePlain(payload);
  Object.assign(settings.branding, normalized.branding);
  Object.assign(settings.uiPreferences.app, normalized.uiPreferences.app);
  Object.assign(
    settings.uiPreferences.breadcrumb,
    normalized.uiPreferences.breadcrumb,
  );
  Object.assign(
    settings.uiPreferences.copyright,
    normalized.uiPreferences.copyright,
  );
  Object.assign(settings.uiPreferences.footer, normalized.uiPreferences.footer);
  Object.assign(settings.uiPreferences.header, normalized.uiPreferences.header);
  Object.assign(
    settings.uiPreferences.navigation,
    normalized.uiPreferences.navigation,
  );
  Object.assign(
    settings.uiPreferences.shortcutKeys,
    normalized.uiPreferences.shortcutKeys,
  );
  Object.assign(
    settings.uiPreferences.sidebar,
    normalized.uiPreferences.sidebar,
  );
  Object.assign(settings.uiPreferences.tabbar, normalized.uiPreferences.tabbar);
  Object.assign(settings.uiPreferences.theme, normalized.uiPreferences.theme);
  Object.assign(
    settings.uiPreferences.transition,
    normalized.uiPreferences.transition,
  );
  Object.assign(settings.uiPreferences.widget, normalized.uiPreferences.widget);
  initialSnapshot.value = normalized;
}

async function loadSettings() {
  loading.value = true;
  try {
    const data = await getSystemSettingsApi();
    applySettings(data);
  } catch (error: unknown) {
    const err = error as {
      message?: string;
      response?: { data?: { msg?: string } };
    };
    message.error(
      err?.message || err?.response?.data?.msg || '加载参数设置失败',
    );
  } finally {
    loading.value = false;
  }
}

function resetToLastSaved() {
  applySettings(initialSnapshot.value);
}

function validateSettings() {
  if (!settings.branding.appName.trim()) {
    message.error('请输入系统名称');
    return false;
  }

  const logo = settings.branding.appLogo.trim();
  if (logo && !/^https?:\/\//.test(logo)) {
    message.error('系统 Logo 目前仅支持 http/https 图片地址');
    return false;
  }

  const placeholderColor = settings.branding.appLogoPlaceholderColor.trim();
  if (!/^#[0-9a-fA-F]{6}$/.test(placeholderColor)) {
    message.error('Logo 占位底色请使用 6 位十六进制颜色，例如 #1d4ed8');
    return false;
  }

  return true;
}

async function handleSave() {
  if (!validateSettings()) {
    return;
  }

  saving.value = true;
  try {
    const payload = clonePlain(settings);
    payload.branding.appName = payload.branding.appName.trim();
    payload.branding.appLogo = payload.branding.appLogo.trim();
    payload.branding.appLogoPlaceholderColor =
      payload.branding.appLogoPlaceholderColor.trim().toLowerCase();
    payload.uiPreferences.app.loginTitle =
      payload.uiPreferences.app.loginTitle.trim();
    payload.uiPreferences.app.loginDescription =
      payload.uiPreferences.app.loginDescription.trim();
    payload.uiPreferences.app.watermarkContent =
      payload.uiPreferences.app.watermarkContent.trim();
    payload.uiPreferences.copyright.companyName =
      payload.uiPreferences.copyright.companyName.trim();
    payload.uiPreferences.copyright.companySiteLink =
      payload.uiPreferences.copyright.companySiteLink.trim();
    payload.uiPreferences.copyright.icp =
      payload.uiPreferences.copyright.icp.trim();
    payload.uiPreferences.copyright.icpLink =
      payload.uiPreferences.copyright.icpLink.trim();

    await updateSystemSettingsApi(payload);
    applySettings(payload);
    applySystemSettingsToRuntime(payload);
    message.success('参数设置已保存，系统展示已同步刷新');
  } catch (error: unknown) {
    const err = error as {
      message?: string;
      response?: { data?: { msg?: string } };
    };
    message.error(err?.message || err?.response?.data?.msg || '保存失败');
  } finally {
    saving.value = false;
  }
}

onMounted(() => {
  loadSettings();
});
</script>

<template>
  <div class="min-h-full bg-slate-50 p-4 md:p-6">
    <div class="mx-auto max-w-7xl space-y-6">
      <section
        class="app-radius-panel border border-slate-200 bg-[linear-gradient(135deg,#ffffff_0%,#f8fbff_45%,#eef5ff_100%)] p-6 shadow-sm"
      >
        <div
          class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between"
        >
          <div class="max-w-3xl">
            <div class="flex items-center gap-3">
              <Tag color="blue">System Settings</Tag>
              <Tag color="gold">Global UI Preferences</Tag>
            </div>
            <h1 class="mt-4 text-2xl font-semibold text-slate-900 md:text-3xl">
              参数设置
            </h1>
            <p class="mt-3 text-sm leading-7 text-slate-600">
              这里是系统设置与界面偏好的唯一承载页。保存后会直接刷新当前运行时配置，后续登录与重进页面也会继续沿用这套全局设置。
            </p>
          </div>

          <Space wrap>
            <Button :disabled="loading || saving" @click="loadSettings"
              >刷新</Button
            >
            <Button
              :disabled="loading || saving || !hasUnsavedChanges"
              @click="resetToLastSaved"
            >
              重置
            </Button>
            <Button
              type="primary"
              :loading="saving"
              :disabled="loading || !hasUnsavedChanges"
              @click="handleSave"
            >
              保存设置
            </Button>
          </Space>
        </div>
      </section>

      <Alert
        show-icon
        type="info"
        message="全局生效说明"
        description="当前页面保存的是系统级界面设置，不再区分个人偏好。右上角偏好设置和登录页工具条默认入口都已移除，后续界面展示以这里保存的配置为准。"
      />

      <Card :bordered="false" class="app-radius-panel shadow-sm">
        <template #title>
          <div class="flex items-center justify-between">
            <span class="text-base font-semibold">系统品牌</span>
            <Tag color="cyan">全局</Tag>
          </div>
        </template>

        <Skeleton :loading="loading" active>
          <div class="grid gap-6 lg:grid-cols-[1.6fr_1fr]">
            <div class="space-y-5">
              <div>
                <label class="mb-2 block text-sm font-medium text-slate-700">
                  系统名称
                </label>
                <Input
                  v-model:value="settings.branding.appName"
                  :maxlength="64"
                  placeholder="请输入系统名称"
                />
              </div>

              <div>
                <label class="mb-2 block text-sm font-medium text-slate-700">
                  系统 Logo 地址
                </label>
                <Input
                  v-model:value="settings.branding.appLogo"
                  placeholder="https://example.com/logo.png"
                />
                <p class="mt-2 text-xs text-slate-500">
                  支持 http/https
                  图片地址，保存后会同步刷新侧栏、顶部和登录页品牌展示。
                </p>
              </div>

              <div>
                <label class="mb-2 block text-sm font-medium text-slate-700">
                  登录页标题
                </label>
                <Input
                  v-model:value="settings.uiPreferences.app.loginTitle"
                  :maxlength="80"
                  placeholder="欢迎回到管理系统"
                />
              </div>

              <div>
                <label class="mb-2 block text-sm font-medium text-slate-700">
                  登录页说明
                </label>
                <Input.TextArea
                  v-model:value="settings.uiPreferences.app.loginDescription"
                  :auto-size="{ minRows: 2, maxRows: 4 }"
                  :maxlength="160"
                  placeholder="输入一段简短的登录页说明文案，保存后会立即更新认证页展示。"
                />
              </div>

              <div>
                <label class="mb-2 block text-sm font-medium text-slate-700">
                  Logo 占位底色
                </label>
                <div class="flex flex-wrap items-center gap-3">
                  <input
                    v-model="settings.branding.appLogoPlaceholderColor"
                    type="color"
                    class="app-radius-control h-11 w-14 cursor-pointer border border-slate-200 bg-white p-1"
                  />
                  <Input
                    v-model:value="settings.branding.appLogoPlaceholderColor"
                    class="max-w-[220px]"
                    placeholder="#1d4ed8"
                  />
                  <div class="flex items-center gap-2 text-xs text-slate-500">
                    <span
                      :style="{ backgroundColor: logoPlaceholderColor }"
                      class="inline-flex size-4 rounded-full border border-slate-200"
                    ></span>
                    <span>空 Logo 时会用这个底色展示默认占位</span>
                  </div>
                </div>
              </div>
            </div>

            <div
              class="app-radius-panel border border-dashed border-slate-200 bg-slate-50 p-5"
            >
              <div class="text-sm font-medium text-slate-700">Logo 预览</div>
              <div class="mt-4 flex min-h-[180px] items-center justify-center">
                <Image
                  v-if="logoPreviewUrl"
                  :src="logoPreviewUrl"
                  :preview="false"
                  class="max-h-[140px] max-w-full object-contain"
                />
                <div
                  v-else
                  class="app-radius-panel flex w-full max-w-[240px] flex-col items-center border border-slate-200 bg-white px-6 py-7 text-center shadow-sm"
                >
                  <div
                    :style="{ backgroundColor: logoPlaceholderColor }"
                    class="app-radius-box flex size-20 items-center justify-center text-2xl font-semibold text-white shadow-[0_16px_36px_rgba(15,23,42,0.12)]"
                  >
                    {{ logoPlaceholderText }}
                  </div>
                  <div class="mt-4 text-sm font-medium text-slate-700">
                    默认 Logo 占位
                  </div>
                  <div class="mt-2 text-xs leading-6 text-slate-400">
                    当前未设置 Logo 地址，这里会先显示系统默认占位效果。
                  </div>
                </div>
              </div>
            </div>
          </div>
        </Skeleton>
      </Card>

      <Card :bordered="false" class="app-radius-panel shadow-sm">
        <Tabs v-model:active-key="activeTab">
          <Tabs.TabPane key="appearance" tab="外观">
            <div class="grid gap-6 xl:grid-cols-2">
              <Card title="主题模式" :bordered="false" class="app-radius-panel">
                <div class="space-y-4">
                  <div class="flex items-center justify-between gap-4">
                    <span class="text-sm text-slate-700">主题模式</span>
                    <Select
                      v-model:value="settings.uiPreferences.theme.mode"
                      :options="themeModeOptions"
                      class="min-w-[180px]"
                    />
                  </div>
                  <div class="flex items-center justify-between gap-4">
                    <span class="text-sm text-slate-700">主题预设</span>
                    <Select
                      v-model:value="settings.uiPreferences.theme.builtinType"
                      :options="themePresetOptions"
                      class="min-w-[180px]"
                    />
                  </div>
                  <div class="flex items-center justify-between gap-4">
                    <span class="text-sm text-slate-700">主色</span>
                    <Input
                      v-model:value="settings.uiPreferences.theme.colorPrimary"
                      class="max-w-[220px]"
                    />
                  </div>
                  <div class="flex items-center justify-between gap-4">
                    <span class="text-sm text-slate-700">圆角</span>
                    <Input
                      v-model:value="settings.uiPreferences.theme.radius"
                      class="max-w-[180px]"
                    />
                  </div>
                  <div class="flex items-center justify-between gap-4">
                    <span class="text-sm text-slate-700">字号</span>
                    <InputNumber
                      v-model:value="settings.uiPreferences.theme.fontSize"
                      :min="12"
                      :max="24"
                      class="min-w-[140px]"
                    />
                  </div>
                </div>
              </Card>

              <Card title="辅助模式" :bordered="false" class="app-radius-panel">
                <div class="space-y-4">
                  <div class="flex items-center justify-between gap-4">
                    <span class="text-sm text-slate-700">灰色模式</span>
                    <Switch
                      v-model:checked="settings.uiPreferences.app.colorGrayMode"
                    />
                  </div>
                  <div class="flex items-center justify-between gap-4">
                    <span class="text-sm text-slate-700">色弱模式</span>
                    <Switch
                      v-model:checked="settings.uiPreferences.app.colorWeakMode"
                    />
                  </div>
                  <div class="flex items-center justify-between gap-4">
                    <span class="text-sm text-slate-700">半深色头部</span>
                    <Switch
                      v-model:checked="
                        settings.uiPreferences.theme.semiDarkHeader
                      "
                    />
                  </div>
                  <div class="flex items-center justify-between gap-4">
                    <span class="text-sm text-slate-700">半深色侧栏</span>
                    <Switch
                      v-model:checked="
                        settings.uiPreferences.theme.semiDarkSidebar
                      "
                    />
                  </div>
                  <div class="flex items-center justify-between gap-4">
                    <span class="text-sm text-slate-700">半深色子菜单</span>
                    <Switch
                      v-model:checked="
                        settings.uiPreferences.theme.semiDarkSidebarSub
                      "
                    />
                  </div>
                </div>
              </Card>
            </div>
          </Tabs.TabPane>

          <Tabs.TabPane key="layout" tab="布局">
            <div class="grid gap-6 xl:grid-cols-2">
              <Card title="整体布局" :bordered="false" class="app-radius-panel">
                <div class="space-y-4">
                  <div class="flex items-center justify-between gap-4">
                    <span class="text-sm text-slate-700">布局模式</span>
                    <Select
                      v-model:value="settings.uiPreferences.app.layout"
                      :options="layoutOptions"
                      class="min-w-[220px]"
                    />
                  </div>
                  <div class="flex items-center justify-between gap-4">
                    <span class="text-sm text-slate-700">内容宽度</span>
                    <Select
                      v-model:value="settings.uiPreferences.app.contentCompact"
                      :options="contentCompactOptions"
                      class="min-w-[180px]"
                    />
                  </div>
                </div>
              </Card>

              <Card title="侧边栏" :bordered="false" class="app-radius-panel">
                <div class="space-y-4">
                  <div class="flex items-center justify-between gap-4">
                    <span class="text-sm text-slate-700">启用侧边栏</span>
                    <Switch
                      v-model:checked="settings.uiPreferences.sidebar.enable"
                    />
                  </div>
                  <div class="flex items-center justify-between gap-4">
                    <span class="text-sm text-slate-700">默认折叠</span>
                    <Switch
                      v-model:checked="settings.uiPreferences.sidebar.collapsed"
                    />
                  </div>
                  <div class="flex items-center justify-between gap-4">
                    <span class="text-sm text-slate-700">显示折叠按钮</span>
                    <Switch
                      v-model:checked="
                        settings.uiPreferences.sidebar.collapsedButton
                      "
                    />
                  </div>
                  <div class="flex items-center justify-between gap-4">
                    <span class="text-sm text-slate-700">折叠显示标题</span>
                    <Switch
                      v-model:checked="
                        settings.uiPreferences.sidebar.collapsedShowTitle
                      "
                    />
                  </div>
                  <div class="flex items-center justify-between gap-4">
                    <span class="text-sm text-slate-700">允许拖拽宽度</span>
                    <Switch
                      v-model:checked="settings.uiPreferences.sidebar.draggable"
                    />
                  </div>
                  <div class="flex items-center justify-between gap-4">
                    <span class="text-sm text-slate-700">悬停展开</span>
                    <Switch
                      v-model:checked="
                        settings.uiPreferences.sidebar.expandOnHover
                      "
                    />
                  </div>
                  <div class="flex items-center justify-between gap-4">
                    <span class="text-sm text-slate-700">固定按钮</span>
                    <Switch
                      v-model:checked="
                        settings.uiPreferences.sidebar.fixedButton
                      "
                    />
                  </div>
                  <div class="flex items-center justify-between gap-4">
                    <span class="text-sm text-slate-700">自动激活子菜单</span>
                    <Switch
                      v-model:checked="
                        settings.uiPreferences.sidebar.autoActivateChild
                      "
                    />
                  </div>
                  <div class="flex items-center justify-between gap-4">
                    <span class="text-sm text-slate-700">侧栏宽度</span>
                    <InputNumber
                      v-model:value="settings.uiPreferences.sidebar.width"
                      :min="160"
                      :max="320"
                    />
                  </div>
                </div>
              </Card>

              <Card title="头部与导航" :bordered="false" class="app-radius-panel">
                <div class="space-y-4">
                  <div class="flex items-center justify-between gap-4">
                    <span class="text-sm text-slate-700">启用头部栏</span>
                    <Switch
                      v-model:checked="settings.uiPreferences.header.enable"
                    />
                  </div>
                  <div class="flex items-center justify-between gap-4">
                    <span class="text-sm text-slate-700">隐藏头部栏</span>
                    <Switch
                      v-model:checked="settings.uiPreferences.header.hidden"
                    />
                  </div>
                  <div class="flex items-center justify-between gap-4">
                    <span class="text-sm text-slate-700">头部模式</span>
                    <Select
                      v-model:value="settings.uiPreferences.header.mode"
                      :options="headerModeOptions"
                      class="min-w-[180px]"
                    />
                  </div>
                  <div class="flex items-center justify-between gap-4">
                    <span class="text-sm text-slate-700">菜单对齐</span>
                    <Select
                      v-model:value="settings.uiPreferences.header.menuAlign"
                      :options="headerAlignOptions"
                      class="min-w-[180px]"
                    />
                  </div>
                  <div class="flex items-center justify-between gap-4">
                    <span class="text-sm text-slate-700">导航风格</span>
                    <Select
                      v-model:value="
                        settings.uiPreferences.navigation.styleType
                      "
                      :options="navigationStyleOptions"
                      class="min-w-[180px]"
                    />
                  </div>
                  <div class="flex items-center justify-between gap-4">
                    <span class="text-sm text-slate-700">导航手风琴</span>
                    <Switch
                      v-model:checked="
                        settings.uiPreferences.navigation.accordion
                      "
                    />
                  </div>
                  <div class="flex items-center justify-between gap-4">
                    <span class="text-sm text-slate-700">导航切割</span>
                    <Switch
                      v-model:checked="settings.uiPreferences.navigation.split"
                    />
                  </div>
                </div>
              </Card>

              <Card
                title="面包屑与标签栏"
                :bordered="false"
                class="app-radius-panel"
              >
                <div class="space-y-4">
                  <div class="flex items-center justify-between gap-4">
                    <span class="text-sm text-slate-700">启用面包屑</span>
                    <Switch
                      v-model:checked="settings.uiPreferences.breadcrumb.enable"
                    />
                  </div>
                  <div class="flex items-center justify-between gap-4">
                    <span class="text-sm text-slate-700">只剩一级时隐藏</span>
                    <Switch
                      v-model:checked="
                        settings.uiPreferences.breadcrumb.hideOnlyOne
                      "
                    />
                  </div>
                  <div class="flex items-center justify-between gap-4">
                    <span class="text-sm text-slate-700">显示首页</span>
                    <Switch
                      v-model:checked="
                        settings.uiPreferences.breadcrumb.showHome
                      "
                    />
                  </div>
                  <div class="flex items-center justify-between gap-4">
                    <span class="text-sm text-slate-700">显示图标</span>
                    <Switch
                      v-model:checked="
                        settings.uiPreferences.breadcrumb.showIcon
                      "
                    />
                  </div>
                  <div class="flex items-center justify-between gap-4">
                    <span class="text-sm text-slate-700">面包屑风格</span>
                    <Select
                      v-model:value="
                        settings.uiPreferences.breadcrumb.styleType
                      "
                      :options="breadcrumbStyleOptions"
                      class="min-w-[180px]"
                    />
                  </div>
                  <hr class="border-slate-200" />
                  <div class="flex items-center justify-between gap-4">
                    <span class="text-sm text-slate-700">启用标签栏</span>
                    <Switch
                      v-model:checked="settings.uiPreferences.tabbar.enable"
                    />
                  </div>
                  <div class="flex items-center justify-between gap-4">
                    <span class="text-sm text-slate-700">标签风格</span>
                    <Select
                      v-model:value="settings.uiPreferences.tabbar.styleType"
                      :options="tabbarStyleOptions"
                      class="min-w-[180px]"
                    />
                  </div>
                  <div
                    v-for="item in [
                      ['显示图标', 'showIcon'],
                      ['显示更多', 'showMore'],
                      ['显示最大化', 'showMaximize'],
                      ['持久化', 'persist'],
                      ['访问历史', 'visitHistory'],
                      ['允许拖拽', 'draggable'],
                      ['滚轮切换', 'wheelable'],
                      ['中键关闭', 'middleClickToClose'],
                    ]"
                    :key="item[1]"
                    class="flex items-center justify-between gap-4"
                  >
                    <span class="text-sm text-slate-700">{{ item[0] }}</span>
                    <Switch
                      v-model:checked="
                        settings.uiPreferences.tabbar[
                          item[1] as keyof typeof settings.uiPreferences.tabbar
                        ] as boolean
                      "
                    />
                  </div>
                  <div class="flex items-center justify-between gap-4">
                    <span class="text-sm text-slate-700">最大标签数</span>
                    <InputNumber
                      v-model:value="settings.uiPreferences.tabbar.maxCount"
                      :min="0"
                      :max="50"
                    />
                  </div>
                </div>
              </Card>

              <Card title="顶部部件" :bordered="false" class="app-radius-panel">
                <div class="grid gap-4 md:grid-cols-2">
                  <div
                    v-for="item in widgetItems"
                    :key="item.key"
                    class="app-radius-box flex items-center justify-between border border-slate-200 px-4 py-3"
                  >
                    <span class="text-sm text-slate-700">{{ item.label }}</span>
                    <Switch
                      v-model:checked="
                        settings.uiPreferences.widget[
                          item.key as keyof typeof settings.uiPreferences.widget
                        ] as boolean
                      "
                    />
                  </div>
                </div>
              </Card>

              <Card title="页脚与版权" :bordered="false" class="app-radius-panel">
                <div class="space-y-4">
                  <div class="flex items-center justify-between gap-4">
                    <span class="text-sm text-slate-700">启用页脚</span>
                    <Switch
                      v-model:checked="settings.uiPreferences.footer.enable"
                    />
                  </div>
                  <div class="flex items-center justify-between gap-4">
                    <span class="text-sm text-slate-700">固定页脚</span>
                    <Switch
                      v-model:checked="settings.uiPreferences.footer.fixed"
                    />
                  </div>
                  <div class="flex items-center justify-between gap-4">
                    <span class="text-sm text-slate-700">启用版权</span>
                    <Switch
                      v-model:checked="settings.uiPreferences.copyright.enable"
                    />
                  </div>
                  <div>
                    <label
                      class="mb-2 block text-sm font-medium text-slate-700"
                    >
                      公司名称
                    </label>
                    <Input
                      v-model:value="
                        settings.uiPreferences.copyright.companyName
                      "
                    />
                  </div>
                  <div>
                    <label
                      class="mb-2 block text-sm font-medium text-slate-700"
                    >
                      公司链接
                    </label>
                    <Input
                      v-model:value="
                        settings.uiPreferences.copyright.companySiteLink
                      "
                    />
                  </div>
                  <div class="grid gap-4 md:grid-cols-2">
                    <div>
                      <label
                        class="mb-2 block text-sm font-medium text-slate-700"
                      >
                        年份
                      </label>
                      <Input
                        v-model:value="settings.uiPreferences.copyright.date"
                      />
                    </div>
                    <div>
                      <label
                        class="mb-2 block text-sm font-medium text-slate-700"
                      >
                        ICP
                      </label>
                      <Input
                        v-model:value="settings.uiPreferences.copyright.icp"
                      />
                    </div>
                  </div>
                  <div>
                    <label
                      class="mb-2 block text-sm font-medium text-slate-700"
                    >
                      ICP 链接
                    </label>
                    <Input
                      v-model:value="settings.uiPreferences.copyright.icpLink"
                    />
                  </div>
                </div>
              </Card>
            </div>
          </Tabs.TabPane>

          <Tabs.TabPane key="general" tab="通用">
            <div class="grid gap-6 xl:grid-cols-2">
              <Card title="通用设置" :bordered="false" class="app-radius-panel">
                <div class="space-y-4">
                  <div class="flex items-center justify-between gap-4">
                    <span class="text-sm text-slate-700">语言</span>
                    <Select
                      v-model:value="settings.uiPreferences.app.locale"
                      :options="SUPPORT_LANGUAGES"
                      class="min-w-[180px]"
                    />
                  </div>
                  <div class="flex items-center justify-between gap-4">
                    <span class="text-sm text-slate-700">动态标题</span>
                    <Switch
                      v-model:checked="settings.uiPreferences.app.dynamicTitle"
                    />
                  </div>
                  <div class="flex items-center justify-between gap-4">
                    <span class="text-sm text-slate-700">检查更新</span>
                    <Switch
                      v-model:checked="
                        settings.uiPreferences.app.enableCheckUpdates
                      "
                    />
                  </div>
                  <div class="flex items-center justify-between gap-4">
                    <span class="text-sm text-slate-700">水印</span>
                    <Switch
                      v-model:checked="settings.uiPreferences.app.watermark"
                    />
                  </div>
                  <div v-if="settings.uiPreferences.app.watermark">
                    <label
                      class="mb-2 block text-sm font-medium text-slate-700"
                    >
                      水印内容
                    </label>
                    <Input
                      v-model:value="
                        settings.uiPreferences.app.watermarkContent
                      "
                      placeholder="请输入水印内容"
                    />
                  </div>
                </div>
              </Card>

              <Card title="动画与过渡" :bordered="false" class="app-radius-panel">
                <div class="space-y-4">
                  <div class="flex items-center justify-between gap-4">
                    <span class="text-sm text-slate-700">启用过渡</span>
                    <Switch
                      v-model:checked="settings.uiPreferences.transition.enable"
                    />
                  </div>
                  <div class="flex items-center justify-between gap-4">
                    <span class="text-sm text-slate-700">启用 Loading</span>
                    <Switch
                      v-model:checked="
                        settings.uiPreferences.transition.loading
                      "
                    />
                  </div>
                  <div class="flex items-center justify-between gap-4">
                    <span class="text-sm text-slate-700">启用进度条</span>
                    <Switch
                      v-model:checked="
                        settings.uiPreferences.transition.progress
                      "
                    />
                  </div>
                  <div class="flex items-center justify-between gap-4">
                    <span class="text-sm text-slate-700">过渡动画</span>
                    <Select
                      v-model:value="settings.uiPreferences.transition.name"
                      :options="transitionOptions"
                      class="min-w-[180px]"
                    />
                  </div>
                </div>
              </Card>
            </div>
          </Tabs.TabPane>

          <Tabs.TabPane key="shortcut" tab="快捷键">
            <div class="grid gap-6 xl:grid-cols-2">
              <Card title="全局快捷键" :bordered="false" class="app-radius-panel">
                <div class="space-y-4">
                  <div
                    v-for="item in shortcutItems"
                    :key="item.key"
                    class="flex items-center justify-between gap-4"
                  >
                    <span class="text-sm text-slate-700">{{ item.label }}</span>
                    <Switch
                      v-model:checked="
                        settings.uiPreferences.shortcutKeys[
                          item.key as keyof typeof settings.uiPreferences.shortcutKeys
                        ] as boolean
                      "
                    />
                  </div>
                </div>
              </Card>
            </div>
          </Tabs.TabPane>
        </Tabs>
      </Card>
    </div>
  </div>
</template>
