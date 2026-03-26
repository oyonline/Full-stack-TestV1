import type {
  AppPreferences,
  BreadcrumbPreferences,
  FooterPreferences,
  HeaderPreferences,
  NavigationPreferences,
  Preferences,
  ShortcutKeyPreferences,
  SidebarPreferences,
  TabbarPreferences,
  ThemePreferences,
  TransitionPreferences,
  WidgetPreferences,
} from '@vben/preferences';

import { requestClient } from '#/api/request';

/** 参数配置列表项（与后端 models.SysConfig 对齐） */
export interface SysConfigItem {
  id: number;
  configName: string;
  configKey: string;
  configValue: string;
  configType: string;
  isFrontend: string;
  remark: string;
  createBy?: number;
  updateBy?: number;
  createdAt?: string;
  updatedAt?: string;
}

/** 参数配置分页响应（requestClient 解包后为 data 内容） */
export interface SysConfigPageResult {
  list: SysConfigItem[];
  count: number;
  pageIndex: number;
  pageSize: number;
}

export interface AppConfigMap {
  [key: string]: string;
}

export interface SystemBrandingSettings {
  appLogo: string;
  appLogoPlaceholderColor: string;
  appName: string;
}

export interface SystemUiAppSettings {
  colorGrayMode: AppPreferences['colorGrayMode'];
  colorWeakMode: AppPreferences['colorWeakMode'];
  contentCompact: AppPreferences['contentCompact'];
  dynamicTitle: AppPreferences['dynamicTitle'];
  enableCheckUpdates: AppPreferences['enableCheckUpdates'];
  layout: AppPreferences['layout'];
  locale: AppPreferences['locale'];
  loginDescription: string;
  loginTitle: string;
  watermark: AppPreferences['watermark'];
  watermarkContent: AppPreferences['watermarkContent'];
}

export interface SystemUiPreferences {
  app: SystemUiAppSettings;
  breadcrumb: Pick<
    BreadcrumbPreferences,
    'enable' | 'hideOnlyOne' | 'showHome' | 'showIcon' | 'styleType'
  >;
  copyright: Pick<
    Preferences['copyright'],
    'companyName' | 'companySiteLink' | 'date' | 'enable' | 'icp' | 'icpLink'
  >;
  footer: Pick<FooterPreferences, 'enable' | 'fixed'>;
  header: Pick<HeaderPreferences, 'enable' | 'hidden' | 'menuAlign' | 'mode'>;
  navigation: Pick<NavigationPreferences, 'accordion' | 'split' | 'styleType'>;
  shortcutKeys: Pick<
    ShortcutKeyPreferences,
    'enable' | 'globalLockScreen' | 'globalLogout' | 'globalSearch'
  >;
  sidebar: Pick<
    SidebarPreferences,
    | 'autoActivateChild'
    | 'collapsed'
    | 'collapsedButton'
    | 'collapsedShowTitle'
    | 'draggable'
    | 'enable'
    | 'expandOnHover'
    | 'fixedButton'
    | 'width'
  >;
  tabbar: Pick<
    TabbarPreferences,
    | 'draggable'
    | 'enable'
    | 'maxCount'
    | 'middleClickToClose'
    | 'persist'
    | 'showIcon'
    | 'showMaximize'
    | 'showMore'
    | 'styleType'
    | 'visitHistory'
    | 'wheelable'
  >;
  theme: Pick<
    ThemePreferences,
    | 'builtinType'
    | 'colorPrimary'
    | 'fontSize'
    | 'mode'
    | 'radius'
    | 'semiDarkHeader'
    | 'semiDarkSidebar'
    | 'semiDarkSidebarSub'
  >;
  transition: Pick<
    TransitionPreferences,
    'enable' | 'loading' | 'name' | 'progress'
  >;
  widget: Pick<
    WidgetPreferences,
    | 'fullscreen'
    | 'globalSearch'
    | 'languageToggle'
    | 'lockScreen'
    | 'notification'
    | 'refresh'
    | 'sidebarToggle'
    | 'themeToggle'
    | 'timezone'
  >;
}

export interface SystemSettingsResponse {
  branding: SystemBrandingSettings;
  uiPreferences: SystemUiPreferences;
}

/** 参数配置列表查询参数 */
interface GetConfigPageParams {
  pageIndex?: number;
  pageSize?: number;
  configName?: string;
  configKey?: string;
  configType?: string;
  isFrontend?: string;
}

/**
 * 获取参数配置分页列表
 * 对接 go-admin GET /api/v1/config
 */
export async function getConfigPage(
  params: GetConfigPageParams = {},
): Promise<SysConfigPageResult> {
  return requestClient.get<SysConfigPageResult>('/v1/config', { params });
}

/**
 * 获取参数配置详情
 * 对接 go-admin GET /api/v1/config/:id
 */
export async function getConfigDetail(id: number): Promise<SysConfigItem> {
  return requestClient.get<SysConfigItem>(`/v1/config/${id}`);
}

/** 新增参数配置请求体 */
interface CreateConfigData {
  configName: string;
  configKey: string;
  configValue: string;
  configType?: string;
  isFrontend?: string;
  remark?: string;
}

/**
 * 新增参数配置
 * 对接 go-admin POST /api/v1/config
 */
export async function createConfig(data: CreateConfigData): Promise<number> {
  return requestClient.post('/v1/config', data);
}

/** 更新参数配置请求体 */
interface UpdateConfigData {
  configName?: string;
  configKey?: string;
  configValue?: string;
  configType?: string;
  isFrontend?: string;
  remark?: string;
}

/**
 * 更新参数配置
 * 对接 go-admin PUT /api/v1/config/:id
 */
export async function updateConfig(
  id: number,
  data: UpdateConfigData,
): Promise<number> {
  return requestClient.put(`/v1/config/${id}`, data);
}

/**
 * 删除参数配置（支持批量）
 * 对接 go-admin DELETE /api/v1/config
 */
export async function deleteConfig(ids: number[]): Promise<number[]> {
  return requestClient.delete('/v1/config', { data: { ids } });
}

/**
 * 获取前台可见配置
 * 对接 go-admin GET /api/v1/app-config
 */
export async function getAppConfigApi(): Promise<AppConfigMap> {
  return requestClient.get<AppConfigMap>('/v1/app-config');
}

/**
 * 获取参数设置页使用的配置映射
 * 对接 go-admin GET /api/v1/set-config
 */
export async function getSystemSettingsApi(): Promise<SystemSettingsResponse> {
  return requestClient.get<SystemSettingsResponse>('/v1/set-config');
}

/**
 * 批量保存参数设置页配置
 * 对接 go-admin PUT /api/v1/set-config
 */
export async function updateSystemSettingsApi(
  data: SystemSettingsResponse,
): Promise<void> {
  return requestClient.put('/v1/set-config', data);
}
