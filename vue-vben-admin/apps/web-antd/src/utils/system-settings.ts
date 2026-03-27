import type { Preferences } from '@vben/preferences';

import { reactive } from 'vue';
import { preferencesManager, updatePreferences } from '@vben/preferences';

import type {
  AppConfigMap,
  SystemSettingsResponse,
  SystemUiPreferences,
} from '#/api/core/config';

function clonePlain<T>(value: T): T {
  return JSON.parse(JSON.stringify(value)) as T;
}

export const systemDisplayState = reactive({
  loginDescription: '',
  loginTitle: '',
});

function createDefaultUiPreferencesFromInitial(
  initialPreferences: Preferences,
): SystemUiPreferences {
  return {
    app: {
      colorGrayMode: initialPreferences.app.colorGrayMode,
      colorWeakMode: initialPreferences.app.colorWeakMode,
      contentCompact: initialPreferences.app.contentCompact,
      dynamicTitle: initialPreferences.app.dynamicTitle,
      enableCheckUpdates: initialPreferences.app.enableCheckUpdates,
      layout: initialPreferences.app.layout,
      loginDescription: '',
      loginTitle: '',
      locale: initialPreferences.app.locale,
      watermark: initialPreferences.app.watermark,
      watermarkContent: initialPreferences.app.watermarkContent,
    },
    breadcrumb: {
      enable: initialPreferences.breadcrumb.enable,
      hideOnlyOne: initialPreferences.breadcrumb.hideOnlyOne,
      showHome: initialPreferences.breadcrumb.showHome,
      showIcon: initialPreferences.breadcrumb.showIcon,
      styleType: initialPreferences.breadcrumb.styleType,
    },
    copyright: {
      companyName: initialPreferences.copyright.companyName,
      companySiteLink: initialPreferences.copyright.companySiteLink,
      date: initialPreferences.copyright.date,
      enable: initialPreferences.copyright.enable,
      icp: initialPreferences.copyright.icp,
      icpLink: initialPreferences.copyright.icpLink,
    },
    footer: {
      enable: initialPreferences.footer.enable,
      fixed: initialPreferences.footer.fixed,
    },
    header: {
      enable: initialPreferences.header.enable,
      hidden: initialPreferences.header.hidden,
      menuAlign: initialPreferences.header.menuAlign,
      mode: initialPreferences.header.mode,
    },
    navigation: {
      accordion: initialPreferences.navigation.accordion,
      split: initialPreferences.navigation.split,
      styleType: initialPreferences.navigation.styleType,
    },
    shortcutKeys: {
      enable: initialPreferences.shortcutKeys.enable,
      globalLockScreen: initialPreferences.shortcutKeys.globalLockScreen,
      globalLogout: initialPreferences.shortcutKeys.globalLogout,
      globalSearch: initialPreferences.shortcutKeys.globalSearch,
    },
    sidebar: {
      autoActivateChild: initialPreferences.sidebar.autoActivateChild,
      collapsed: initialPreferences.sidebar.collapsed,
      collapsedButton: initialPreferences.sidebar.collapsedButton,
      collapsedShowTitle: initialPreferences.sidebar.collapsedShowTitle,
      draggable: initialPreferences.sidebar.draggable,
      enable: initialPreferences.sidebar.enable,
      expandOnHover: initialPreferences.sidebar.expandOnHover,
      fixedButton: initialPreferences.sidebar.fixedButton,
      width: initialPreferences.sidebar.width,
    },
    tabbar: {
      draggable: initialPreferences.tabbar.draggable,
      enable: initialPreferences.tabbar.enable,
      maxCount: initialPreferences.tabbar.maxCount,
      middleClickToClose: initialPreferences.tabbar.middleClickToClose,
      persist: initialPreferences.tabbar.persist,
      showIcon: initialPreferences.tabbar.showIcon,
      showMaximize: initialPreferences.tabbar.showMaximize,
      showMore: initialPreferences.tabbar.showMore,
      styleType: initialPreferences.tabbar.styleType,
      visitHistory: initialPreferences.tabbar.visitHistory,
      wheelable: initialPreferences.tabbar.wheelable,
    },
    theme: {
      builtinType: initialPreferences.theme.builtinType,
      colorPrimary: initialPreferences.theme.colorPrimary,
      fontSize: initialPreferences.theme.fontSize,
      mode: initialPreferences.theme.mode,
      radius: initialPreferences.theme.radius,
      semiDarkHeader: initialPreferences.theme.semiDarkHeader,
      semiDarkSidebar: initialPreferences.theme.semiDarkSidebar,
      semiDarkSidebarSub: initialPreferences.theme.semiDarkSidebarSub,
    },
    transition: {
      enable: initialPreferences.transition.enable,
      loading: initialPreferences.transition.loading,
      name: initialPreferences.transition.name,
      progress: initialPreferences.transition.progress,
    },
    widget: {
      fullscreen: initialPreferences.widget.fullscreen,
      globalSearch: initialPreferences.widget.globalSearch,
      languageToggle: initialPreferences.widget.languageToggle,
      lockScreen: initialPreferences.widget.lockScreen,
      notification: initialPreferences.widget.notification,
      refresh: initialPreferences.widget.refresh,
      sidebarToggle: initialPreferences.widget.sidebarToggle,
      themeToggle: initialPreferences.widget.themeToggle,
      timezone: initialPreferences.widget.timezone,
    },
  };
}

export function createDefaultSystemUiPreferences(): SystemUiPreferences {
  return createDefaultUiPreferencesFromInitial(
    preferencesManager.getInitialPreferences(),
  );
}

export function createDefaultSystemSettings(): SystemSettingsResponse {
  const initialPreferences = preferencesManager.getInitialPreferences();
  return {
    branding: {
      appLogo: '',
      appLogoPlaceholderColor: '#1d4ed8',
      appName: initialPreferences.app.name,
    },
    uiPreferences: createDefaultUiPreferencesFromInitial(initialPreferences),
  };
}

function applyCompactUiPreferencesSection(
  target: SystemUiPreferences,
  sectionKey: string,
  sectionValue: Record<string, unknown>,
) {
  switch (sectionKey) {
    case 'a':
      if ('g' in sectionValue)
        target.app.colorGrayMode = Boolean(sectionValue.g);
      if ('w' in sectionValue)
        target.app.colorWeakMode = Boolean(sectionValue.w);
      if ('c' in sectionValue) {
        target.app.contentCompact = String(
          sectionValue.c,
        ) as SystemUiPreferences['app']['contentCompact'];
      }
      if ('d' in sectionValue)
        target.app.dynamicTitle = Boolean(sectionValue.d);
      if ('u' in sectionValue) {
        target.app.enableCheckUpdates = Boolean(sectionValue.u);
      }
      if ('l' in sectionValue) {
        target.app.layout = String(
          sectionValue.l,
        ) as SystemUiPreferences['app']['layout'];
      }
      if ('t' in sectionValue) {
        target.app.loginTitle = String(sectionValue.t ?? '');
      }
      if ('y' in sectionValue) {
        target.app.loginDescription = String(sectionValue.y ?? '');
      }
      if ('o' in sectionValue) {
        target.app.locale = String(
          sectionValue.o,
        ) as SystemUiPreferences['app']['locale'];
      }
      if ('m' in sectionValue) target.app.watermark = Boolean(sectionValue.m);
      if ('x' in sectionValue) {
        target.app.watermarkContent = String(sectionValue.x ?? '');
      }
      return;
    case 'b':
      if ('e' in sectionValue)
        target.breadcrumb.enable = Boolean(sectionValue.e);
      if ('h' in sectionValue) {
        target.breadcrumb.hideOnlyOne = Boolean(sectionValue.h);
      }
      if ('m' in sectionValue)
        target.breadcrumb.showHome = Boolean(sectionValue.m);
      if ('i' in sectionValue)
        target.breadcrumb.showIcon = Boolean(sectionValue.i);
      if ('s' in sectionValue) {
        target.breadcrumb.styleType = String(
          sectionValue.s,
        ) as SystemUiPreferences['breadcrumb']['styleType'];
      }
      return;
    case 'c':
      if ('n' in sectionValue)
        target.copyright.companyName = String(sectionValue.n);
      if ('s' in sectionValue) {
        target.copyright.companySiteLink = String(sectionValue.s);
      }
      if ('d' in sectionValue) target.copyright.date = String(sectionValue.d);
      if ('e' in sectionValue)
        target.copyright.enable = Boolean(sectionValue.e);
      if ('i' in sectionValue)
        target.copyright.icp = String(sectionValue.i ?? '');
      if ('l' in sectionValue)
        target.copyright.icpLink = String(sectionValue.l ?? '');
      return;
    case 'f':
      if ('e' in sectionValue) target.footer.enable = Boolean(sectionValue.e);
      if ('f' in sectionValue) target.footer.fixed = Boolean(sectionValue.f);
      return;
    case 'h':
      if ('e' in sectionValue) target.header.enable = Boolean(sectionValue.e);
      if ('h' in sectionValue) target.header.hidden = Boolean(sectionValue.h);
      if ('a' in sectionValue) {
        target.header.menuAlign = String(
          sectionValue.a,
        ) as SystemUiPreferences['header']['menuAlign'];
      }
      if ('m' in sectionValue) {
        target.header.mode = String(
          sectionValue.m,
        ) as SystemUiPreferences['header']['mode'];
      }
      return;
    case 'n':
      if ('a' in sectionValue)
        target.navigation.accordion = Boolean(sectionValue.a);
      if ('s' in sectionValue)
        target.navigation.split = Boolean(sectionValue.s);
      if ('t' in sectionValue) {
        target.navigation.styleType = String(
          sectionValue.t,
        ) as SystemUiPreferences['navigation']['styleType'];
      }
      return;
    case 'k':
      if ('e' in sectionValue)
        target.shortcutKeys.enable = Boolean(sectionValue.e);
      if ('s' in sectionValue) {
        target.shortcutKeys.globalSearch = Boolean(sectionValue.s);
      }
      if ('o' in sectionValue) {
        target.shortcutKeys.globalLogout = Boolean(sectionValue.o);
      }
      if ('l' in sectionValue) {
        target.shortcutKeys.globalLockScreen = Boolean(sectionValue.l);
      }
      return;
    case 's':
      if ('a' in sectionValue)
        target.sidebar.autoActivateChild = Boolean(sectionValue.a);
      if ('c' in sectionValue)
        target.sidebar.collapsed = Boolean(sectionValue.c);
      if ('b' in sectionValue) {
        target.sidebar.collapsedButton = Boolean(sectionValue.b);
      }
      if ('t' in sectionValue) {
        target.sidebar.collapsedShowTitle = Boolean(sectionValue.t);
      }
      if ('d' in sectionValue)
        target.sidebar.draggable = Boolean(sectionValue.d);
      if ('e' in sectionValue) target.sidebar.enable = Boolean(sectionValue.e);
      if ('h' in sectionValue) {
        target.sidebar.expandOnHover = Boolean(sectionValue.h);
      }
      if ('f' in sectionValue)
        target.sidebar.fixedButton = Boolean(sectionValue.f);
      if ('w' in sectionValue) target.sidebar.width = Number(sectionValue.w);
      return;
    case 't':
      if ('d' in sectionValue)
        target.tabbar.draggable = Boolean(sectionValue.d);
      if ('e' in sectionValue) target.tabbar.enable = Boolean(sectionValue.e);
      if ('p' in sectionValue) target.tabbar.persist = Boolean(sectionValue.p);
      if ('v' in sectionValue) {
        target.tabbar.visitHistory = Boolean(sectionValue.v);
      }
      if ('i' in sectionValue) target.tabbar.showIcon = Boolean(sectionValue.i);
      if ('x' in sectionValue) {
        target.tabbar.showMaximize = Boolean(sectionValue.x);
      }
      if ('m' in sectionValue) target.tabbar.showMore = Boolean(sectionValue.m);
      if ('s' in sectionValue) {
        target.tabbar.styleType = String(
          sectionValue.s,
        ) as SystemUiPreferences['tabbar']['styleType'];
      }
      if ('w' in sectionValue)
        target.tabbar.wheelable = Boolean(sectionValue.w);
      if ('n' in sectionValue) target.tabbar.maxCount = Number(sectionValue.n);
      if ('c' in sectionValue) {
        target.tabbar.middleClickToClose = Boolean(sectionValue.c);
      }
      return;
    case 'm':
      if ('b' in sectionValue)
        target.theme.builtinType = String(sectionValue.b);
      if ('p' in sectionValue)
        target.theme.colorPrimary = String(sectionValue.p);
      if ('m' in sectionValue) {
        target.theme.mode = String(
          sectionValue.m,
        ) as SystemUiPreferences['theme']['mode'];
      }
      if ('r' in sectionValue) target.theme.radius = String(sectionValue.r);
      if ('f' in sectionValue) target.theme.fontSize = Number(sectionValue.f);
      if ('h' in sectionValue) {
        target.theme.semiDarkHeader = Boolean(sectionValue.h);
      }
      if ('s' in sectionValue) {
        target.theme.semiDarkSidebar = Boolean(sectionValue.s);
      }
      if ('u' in sectionValue) {
        target.theme.semiDarkSidebarSub = Boolean(sectionValue.u);
      }
      return;
    case 'r':
      if ('e' in sectionValue)
        target.transition.enable = Boolean(sectionValue.e);
      if ('l' in sectionValue)
        target.transition.loading = Boolean(sectionValue.l);
      if ('n' in sectionValue) target.transition.name = String(sectionValue.n);
      if ('p' in sectionValue) {
        target.transition.progress = Boolean(sectionValue.p);
      }
      return;
    case 'w':
      if ('f' in sectionValue)
        target.widget.fullscreen = Boolean(sectionValue.f);
      if ('s' in sectionValue) {
        target.widget.globalSearch = Boolean(sectionValue.s);
      }
      if ('l' in sectionValue) {
        target.widget.languageToggle = Boolean(sectionValue.l);
      }
      if ('k' in sectionValue)
        target.widget.lockScreen = Boolean(sectionValue.k);
      if ('n' in sectionValue) {
        target.widget.notification = Boolean(sectionValue.n);
      }
      if ('r' in sectionValue) target.widget.refresh = Boolean(sectionValue.r);
      if ('b' in sectionValue) {
        target.widget.sidebarToggle = Boolean(sectionValue.b);
      }
      if ('t' in sectionValue)
        target.widget.themeToggle = Boolean(sectionValue.t);
      if ('z' in sectionValue) target.widget.timezone = Boolean(sectionValue.z);
      return;
  }
}

export function decodeSystemUiPreferences(
  raw: string | undefined,
): SystemUiPreferences {
  const defaults = createDefaultSystemUiPreferences();
  if (!raw) {
    return defaults;
  }

  try {
    const parsed = JSON.parse(raw) as Record<string, Record<string, unknown>>;
    const next = clonePlain(defaults);
    for (const [sectionKey, sectionValue] of Object.entries(parsed)) {
      if (sectionValue && typeof sectionValue === 'object') {
        applyCompactUiPreferencesSection(next, sectionKey, sectionValue);
      }
    }
    return next;
  } catch {
    return defaults;
  }
}

export function applySystemSettingsToRuntime(
  settings: Pick<SystemSettingsResponse, 'branding' | 'uiPreferences'>,
) {
  systemDisplayState.loginTitle = String(
    settings.uiPreferences.app.loginTitle ?? '',
  ).trim();
  systemDisplayState.loginDescription = String(
    settings.uiPreferences.app.loginDescription ?? '',
  ).trim();
  updatePreferences({
    ...settings.uiPreferences,
    app: {
      ...settings.uiPreferences.app,
      name: settings.branding.appName,
    },
    logo: {
      placeholderBgColor: settings.branding.appLogoPlaceholderColor,
      source: settings.branding.appLogo,
      sourceDark: settings.branding.appLogo,
    },
  });
}

export function appConfigMapToSystemSettings(
  appConfig: AppConfigMap,
): Pick<SystemSettingsResponse, 'branding' | 'uiPreferences'> {
  return {
    branding: {
      appLogo: appConfig.sys_app_logo?.trim() || '',
      appLogoPlaceholderColor:
        appConfig.sys_app_logo_placeholder_color?.trim() || '#1d4ed8',
      appName:
        appConfig.sys_app_name?.trim() ||
        preferencesManager.getInitialPreferences().app.name,
    },
    uiPreferences: decodeSystemUiPreferences(appConfig.sys_ui_preferences),
  };
}
