package service

import (
	"encoding/json"
	"errors"

	"go-admin/app/admin/models"
	"go-admin/app/admin/service/dto"

	"gorm.io/gorm"
)

const (
	systemAppNameKey                 = "sys_app_name"
	systemAppLogoKey                 = "sys_app_logo"
	systemAppLogoPlaceholderColorKey = "sys_app_logo_placeholder_color"
	systemUIPreferencesKey           = "sys_ui_preferences"

	systemUIPreferencesConfigName   = "界面偏好配置"
	systemUIPreferencesConfigType   = "Y"
	systemUIPreferencesIsFrontend   = "1"
	systemUIPreferencesConfigRemark = "系统级界面设置(JSON diff)"
)

type compactSystemSettings struct {
	App         *compactSystemAppSettings         `json:"a,omitempty"`
	Breadcrumb  *compactSystemBreadcrumbSettings  `json:"b,omitempty"`
	Copyright   *compactSystemCopyrightSettings   `json:"c,omitempty"`
	Footer      *compactSystemFooterSettings      `json:"f,omitempty"`
	Header      *compactSystemHeaderSettings      `json:"h,omitempty"`
	Navigation  *compactSystemNavigationSettings  `json:"n,omitempty"`
	ShortcutKey *compactSystemShortcutKeySettings `json:"k,omitempty"`
	Sidebar     *compactSystemSidebarSettings     `json:"s,omitempty"`
	Tabbar      *compactSystemTabbarSettings      `json:"t,omitempty"`
	Theme       *compactSystemThemeSettings       `json:"m,omitempty"`
	Transition  *compactSystemTransitionSettings  `json:"r,omitempty"`
	Widget      *compactSystemWidgetSettings      `json:"w,omitempty"`
}

type compactSystemAppSettings struct {
	ColorGrayMode     *bool   `json:"g,omitempty"`
	ColorWeakMode     *bool   `json:"w,omitempty"`
	ContentCompact    *string `json:"c,omitempty"`
	DynamicTitle      *bool   `json:"d,omitempty"`
	EnableCheckUpdate *bool   `json:"u,omitempty"`
	Layout            *string `json:"l,omitempty"`
	LoginDescription  *string `json:"y,omitempty"`
	LoginTitle        *string `json:"t,omitempty"`
	Locale            *string `json:"o,omitempty"`
	Watermark         *bool   `json:"m,omitempty"`
	WatermarkContent  *string `json:"x,omitempty"`
}

type compactSystemBreadcrumbSettings struct {
	Enable      *bool   `json:"e,omitempty"`
	HideOnlyOne *bool   `json:"h,omitempty"`
	ShowHome    *bool   `json:"m,omitempty"`
	ShowIcon    *bool   `json:"i,omitempty"`
	StyleType   *string `json:"s,omitempty"`
}

type compactSystemCopyrightSettings struct {
	CompanyName     *string `json:"n,omitempty"`
	CompanySiteLink *string `json:"s,omitempty"`
	Date            *string `json:"d,omitempty"`
	Enable          *bool   `json:"e,omitempty"`
	Icp             *string `json:"i,omitempty"`
	IcpLink         *string `json:"l,omitempty"`
}

type compactSystemFooterSettings struct {
	Enable *bool `json:"e,omitempty"`
	Fixed  *bool `json:"f,omitempty"`
}

type compactSystemHeaderSettings struct {
	Enable    *bool   `json:"e,omitempty"`
	Hidden    *bool   `json:"h,omitempty"`
	MenuAlign *string `json:"a,omitempty"`
	Mode      *string `json:"m,omitempty"`
}

type compactSystemNavigationSettings struct {
	Accordion *bool   `json:"a,omitempty"`
	Split     *bool   `json:"s,omitempty"`
	StyleType *string `json:"t,omitempty"`
}

type compactSystemShortcutKeySettings struct {
	Enable           *bool `json:"e,omitempty"`
	GlobalSearch     *bool `json:"s,omitempty"`
	GlobalLogout     *bool `json:"o,omitempty"`
	GlobalLockScreen *bool `json:"l,omitempty"`
}

type compactSystemSidebarSettings struct {
	AutoActivateChild *bool `json:"a,omitempty"`
	Collapsed         *bool `json:"c,omitempty"`
	CollapsedButton   *bool `json:"b,omitempty"`
	CollapsedShowText *bool `json:"t,omitempty"`
	Draggable         *bool `json:"d,omitempty"`
	Enable            *bool `json:"e,omitempty"`
	ExpandOnHover     *bool `json:"h,omitempty"`
	FixedButton       *bool `json:"f,omitempty"`
	Width             *int  `json:"w,omitempty"`
}

type compactSystemTabbarSettings struct {
	Draggable          *bool   `json:"d,omitempty"`
	Enable             *bool   `json:"e,omitempty"`
	Persist            *bool   `json:"p,omitempty"`
	VisitHistory       *bool   `json:"v,omitempty"`
	ShowIcon           *bool   `json:"i,omitempty"`
	ShowMaximize       *bool   `json:"x,omitempty"`
	ShowMore           *bool   `json:"m,omitempty"`
	StyleType          *string `json:"s,omitempty"`
	Wheelable          *bool   `json:"w,omitempty"`
	MaxCount           *int    `json:"n,omitempty"`
	MiddleClickToClose *bool   `json:"c,omitempty"`
}

type compactSystemThemeSettings struct {
	BuiltinType        *string `json:"b,omitempty"`
	ColorPrimary       *string `json:"p,omitempty"`
	Mode               *string `json:"m,omitempty"`
	Radius             *string `json:"r,omitempty"`
	FontSize           *int    `json:"f,omitempty"`
	SemiDarkHeader     *bool   `json:"h,omitempty"`
	SemiDarkSidebar    *bool   `json:"s,omitempty"`
	SemiDarkSidebarSub *bool   `json:"u,omitempty"`
}

type compactSystemTransitionSettings struct {
	Enable   *bool   `json:"e,omitempty"`
	Loading  *bool   `json:"l,omitempty"`
	Name     *string `json:"n,omitempty"`
	Progress *bool   `json:"p,omitempty"`
}

type compactSystemWidgetSettings struct {
	Fullscreen     *bool `json:"f,omitempty"`
	GlobalSearch   *bool `json:"s,omitempty"`
	LanguageToggle *bool `json:"l,omitempty"`
	LockScreen     *bool `json:"k,omitempty"`
	Notification   *bool `json:"n,omitempty"`
	Refresh        *bool `json:"r,omitempty"`
	SidebarToggle  *bool `json:"b,omitempty"`
	ThemeToggle    *bool `json:"t,omitempty"`
	Timezone       *bool `json:"z,omitempty"`
}

func defaultSystemUiPreferences() dto.SystemUiPreferences {
	return dto.SystemUiPreferences{
		App: dto.SystemUiAppSettings{
			ColorGrayMode:      false,
			ColorWeakMode:      false,
			ContentCompact:     "wide",
			DynamicTitle:       true,
			EnableCheckUpdates: true,
			Layout:             "sidebar-nav",
			LoginDescription:   "",
			LoginTitle:         "",
			Locale:             "zh-CN",
			Watermark:          false,
			WatermarkContent:   "",
		},
		Breadcrumb: dto.SystemUiBreadcrumbSettings{
			Enable:      true,
			HideOnlyOne: false,
			ShowHome:    false,
			ShowIcon:    true,
			StyleType:   "normal",
		},
		Copyright: dto.SystemUiCopyrightSettings{
			CompanyName:     "Vben",
			CompanySiteLink: "https://www.vben.pro",
			Date:            "2024",
			Enable:          true,
			Icp:             "",
			IcpLink:         "",
		},
		Footer: dto.SystemUiFooterSettings{
			Enable: false,
			Fixed:  false,
		},
		Header: dto.SystemUiHeaderSettings{
			Enable:    true,
			Hidden:    false,
			MenuAlign: "start",
			Mode:      "fixed",
		},
		Navigation: dto.SystemUiNavigationSettings{
			Accordion: true,
			Split:     true,
			StyleType: "rounded",
		},
		ShortcutKeys: dto.SystemUiShortcutKeySettings{
			Enable:           true,
			GlobalSearch:     true,
			GlobalLogout:     true,
			GlobalLockScreen: true,
		},
		Sidebar: dto.SystemUiSidebarSettings{
			AutoActivateChild:  false,
			Collapsed:          false,
			CollapsedButton:    true,
			CollapsedShowTitle: false,
			Draggable:          true,
			Enable:             true,
			ExpandOnHover:      true,
			FixedButton:        true,
			Width:              224,
		},
		Tabbar: dto.SystemUiTabbarSettings{
			Draggable:          true,
			Enable:             true,
			Persist:            true,
			VisitHistory:       true,
			ShowIcon:           true,
			ShowMaximize:       true,
			ShowMore:           true,
			StyleType:          "chrome",
			Wheelable:          true,
			MaxCount:           0,
			MiddleClickToClose: false,
		},
		Theme: dto.SystemUiThemeSettings{
			BuiltinType:        "default",
			ColorPrimary:       "hsl(212 100% 45%)",
			Mode:               "dark",
			Radius:             "0.5",
			FontSize:           16,
			SemiDarkHeader:     false,
			SemiDarkSidebar:    false,
			SemiDarkSidebarSub: false,
		},
		Transition: dto.SystemUiTransitionSettings{
			Enable:   true,
			Loading:  true,
			Name:     "fade-slide",
			Progress: true,
		},
		Widget: dto.SystemUiWidgetSettings{
			Fullscreen:     true,
			GlobalSearch:   true,
			LanguageToggle: true,
			LockScreen:     true,
			Notification:   true,
			Refresh:        true,
			SidebarToggle:  true,
			ThemeToggle:    true,
			Timezone:       true,
		},
	}
}

func defaultSystemSettingsPayload() dto.SystemSettingsPayload {
	return dto.SystemSettingsPayload{
		Branding: dto.SystemBrandingSettings{
			AppLogoPlaceholderColor: "#1d4ed8",
		},
		UIPreferences: defaultSystemUiPreferences(),
	}
}

func sanitizeSystemSettingsPayload(payload dto.SystemSettingsPayload) dto.SystemSettingsPayload {
	payload.UIPreferences = sanitizeSystemUiPreferences(payload.UIPreferences)
	return payload
}

func sanitizeSystemUiPreferences(input dto.SystemUiPreferences) dto.SystemUiPreferences {
	sanitized := input

	if sanitized.App.ContentCompact == "" {
		sanitized.App.ContentCompact = defaultSystemUiPreferences().App.ContentCompact
	}
	if sanitized.App.Layout == "" {
		sanitized.App.Layout = defaultSystemUiPreferences().App.Layout
	}
	if sanitized.App.Locale == "" {
		sanitized.App.Locale = defaultSystemUiPreferences().App.Locale
	}
	if sanitized.Header.MenuAlign == "" {
		sanitized.Header.MenuAlign = defaultSystemUiPreferences().Header.MenuAlign
	}
	if sanitized.Header.Mode == "" {
		sanitized.Header.Mode = defaultSystemUiPreferences().Header.Mode
	}
	if sanitized.Breadcrumb.StyleType == "" {
		sanitized.Breadcrumb.StyleType = defaultSystemUiPreferences().Breadcrumb.StyleType
	}
	if sanitized.Navigation.StyleType == "" {
		sanitized.Navigation.StyleType = defaultSystemUiPreferences().Navigation.StyleType
	}
	if sanitized.Tabbar.StyleType == "" {
		sanitized.Tabbar.StyleType = defaultSystemUiPreferences().Tabbar.StyleType
	}
	if sanitized.Theme.BuiltinType == "" {
		sanitized.Theme.BuiltinType = defaultSystemUiPreferences().Theme.BuiltinType
	}
	if sanitized.Theme.ColorPrimary == "" {
		sanitized.Theme.ColorPrimary = defaultSystemUiPreferences().Theme.ColorPrimary
	}
	if sanitized.Theme.Mode == "" {
		sanitized.Theme.Mode = defaultSystemUiPreferences().Theme.Mode
	}
	if sanitized.Theme.Radius == "" {
		sanitized.Theme.Radius = defaultSystemUiPreferences().Theme.Radius
	}
	if sanitized.Transition.Name == "" {
		sanitized.Transition.Name = defaultSystemUiPreferences().Transition.Name
	}
	if sanitized.Sidebar.Width <= 0 {
		sanitized.Sidebar.Width = defaultSystemUiPreferences().Sidebar.Width
	}
	if sanitized.Tabbar.MaxCount < 0 {
		sanitized.Tabbar.MaxCount = defaultSystemUiPreferences().Tabbar.MaxCount
	}
	if sanitized.Theme.FontSize <= 0 {
		sanitized.Theme.FontSize = defaultSystemUiPreferences().Theme.FontSize
	}

	return sanitized
}

func decodeSystemUiPreferences(raw string) (dto.SystemUiPreferences, error) {
	defaults := defaultSystemUiPreferences()
	if raw == "" {
		return defaults, nil
	}

	var compact compactSystemSettings
	if err := json.Unmarshal([]byte(raw), &compact); err != nil {
		return dto.SystemUiPreferences{}, err
	}

	if compact.App != nil {
		if compact.App.ColorGrayMode != nil {
			defaults.App.ColorGrayMode = *compact.App.ColorGrayMode
		}
		if compact.App.ColorWeakMode != nil {
			defaults.App.ColorWeakMode = *compact.App.ColorWeakMode
		}
		if compact.App.ContentCompact != nil {
			defaults.App.ContentCompact = *compact.App.ContentCompact
		}
		if compact.App.DynamicTitle != nil {
			defaults.App.DynamicTitle = *compact.App.DynamicTitle
		}
		if compact.App.EnableCheckUpdate != nil {
			defaults.App.EnableCheckUpdates = *compact.App.EnableCheckUpdate
		}
		if compact.App.Layout != nil {
			defaults.App.Layout = *compact.App.Layout
		}
		if compact.App.LoginTitle != nil {
			defaults.App.LoginTitle = *compact.App.LoginTitle
		}
		if compact.App.LoginDescription != nil {
			defaults.App.LoginDescription = *compact.App.LoginDescription
		}
		if compact.App.Locale != nil {
			defaults.App.Locale = *compact.App.Locale
		}
		if compact.App.Watermark != nil {
			defaults.App.Watermark = *compact.App.Watermark
		}
		if compact.App.WatermarkContent != nil {
			defaults.App.WatermarkContent = *compact.App.WatermarkContent
		}
	}

	if compact.Breadcrumb != nil {
		if compact.Breadcrumb.Enable != nil {
			defaults.Breadcrumb.Enable = *compact.Breadcrumb.Enable
		}
		if compact.Breadcrumb.HideOnlyOne != nil {
			defaults.Breadcrumb.HideOnlyOne = *compact.Breadcrumb.HideOnlyOne
		}
		if compact.Breadcrumb.ShowHome != nil {
			defaults.Breadcrumb.ShowHome = *compact.Breadcrumb.ShowHome
		}
		if compact.Breadcrumb.ShowIcon != nil {
			defaults.Breadcrumb.ShowIcon = *compact.Breadcrumb.ShowIcon
		}
		if compact.Breadcrumb.StyleType != nil {
			defaults.Breadcrumb.StyleType = *compact.Breadcrumb.StyleType
		}
	}

	if compact.Copyright != nil {
		if compact.Copyright.CompanyName != nil {
			defaults.Copyright.CompanyName = *compact.Copyright.CompanyName
		}
		if compact.Copyright.CompanySiteLink != nil {
			defaults.Copyright.CompanySiteLink = *compact.Copyright.CompanySiteLink
		}
		if compact.Copyright.Date != nil {
			defaults.Copyright.Date = *compact.Copyright.Date
		}
		if compact.Copyright.Enable != nil {
			defaults.Copyright.Enable = *compact.Copyright.Enable
		}
		if compact.Copyright.Icp != nil {
			defaults.Copyright.Icp = *compact.Copyright.Icp
		}
		if compact.Copyright.IcpLink != nil {
			defaults.Copyright.IcpLink = *compact.Copyright.IcpLink
		}
	}

	if compact.Footer != nil {
		if compact.Footer.Enable != nil {
			defaults.Footer.Enable = *compact.Footer.Enable
		}
		if compact.Footer.Fixed != nil {
			defaults.Footer.Fixed = *compact.Footer.Fixed
		}
	}

	if compact.Header != nil {
		if compact.Header.Enable != nil {
			defaults.Header.Enable = *compact.Header.Enable
		}
		if compact.Header.Hidden != nil {
			defaults.Header.Hidden = *compact.Header.Hidden
		}
		if compact.Header.MenuAlign != nil {
			defaults.Header.MenuAlign = *compact.Header.MenuAlign
		}
		if compact.Header.Mode != nil {
			defaults.Header.Mode = *compact.Header.Mode
		}
	}

	if compact.Navigation != nil {
		if compact.Navigation.Accordion != nil {
			defaults.Navigation.Accordion = *compact.Navigation.Accordion
		}
		if compact.Navigation.Split != nil {
			defaults.Navigation.Split = *compact.Navigation.Split
		}
		if compact.Navigation.StyleType != nil {
			defaults.Navigation.StyleType = *compact.Navigation.StyleType
		}
	}

	if compact.ShortcutKey != nil {
		if compact.ShortcutKey.Enable != nil {
			defaults.ShortcutKeys.Enable = *compact.ShortcutKey.Enable
		}
		if compact.ShortcutKey.GlobalSearch != nil {
			defaults.ShortcutKeys.GlobalSearch = *compact.ShortcutKey.GlobalSearch
		}
		if compact.ShortcutKey.GlobalLogout != nil {
			defaults.ShortcutKeys.GlobalLogout = *compact.ShortcutKey.GlobalLogout
		}
		if compact.ShortcutKey.GlobalLockScreen != nil {
			defaults.ShortcutKeys.GlobalLockScreen = *compact.ShortcutKey.GlobalLockScreen
		}
	}

	if compact.Sidebar != nil {
		if compact.Sidebar.AutoActivateChild != nil {
			defaults.Sidebar.AutoActivateChild = *compact.Sidebar.AutoActivateChild
		}
		if compact.Sidebar.Collapsed != nil {
			defaults.Sidebar.Collapsed = *compact.Sidebar.Collapsed
		}
		if compact.Sidebar.CollapsedButton != nil {
			defaults.Sidebar.CollapsedButton = *compact.Sidebar.CollapsedButton
		}
		if compact.Sidebar.CollapsedShowText != nil {
			defaults.Sidebar.CollapsedShowTitle = *compact.Sidebar.CollapsedShowText
		}
		if compact.Sidebar.Draggable != nil {
			defaults.Sidebar.Draggable = *compact.Sidebar.Draggable
		}
		if compact.Sidebar.Enable != nil {
			defaults.Sidebar.Enable = *compact.Sidebar.Enable
		}
		if compact.Sidebar.ExpandOnHover != nil {
			defaults.Sidebar.ExpandOnHover = *compact.Sidebar.ExpandOnHover
		}
		if compact.Sidebar.FixedButton != nil {
			defaults.Sidebar.FixedButton = *compact.Sidebar.FixedButton
		}
		if compact.Sidebar.Width != nil {
			defaults.Sidebar.Width = *compact.Sidebar.Width
		}
	}

	if compact.Tabbar != nil {
		if compact.Tabbar.Draggable != nil {
			defaults.Tabbar.Draggable = *compact.Tabbar.Draggable
		}
		if compact.Tabbar.Enable != nil {
			defaults.Tabbar.Enable = *compact.Tabbar.Enable
		}
		if compact.Tabbar.Persist != nil {
			defaults.Tabbar.Persist = *compact.Tabbar.Persist
		}
		if compact.Tabbar.VisitHistory != nil {
			defaults.Tabbar.VisitHistory = *compact.Tabbar.VisitHistory
		}
		if compact.Tabbar.ShowIcon != nil {
			defaults.Tabbar.ShowIcon = *compact.Tabbar.ShowIcon
		}
		if compact.Tabbar.ShowMaximize != nil {
			defaults.Tabbar.ShowMaximize = *compact.Tabbar.ShowMaximize
		}
		if compact.Tabbar.ShowMore != nil {
			defaults.Tabbar.ShowMore = *compact.Tabbar.ShowMore
		}
		if compact.Tabbar.StyleType != nil {
			defaults.Tabbar.StyleType = *compact.Tabbar.StyleType
		}
		if compact.Tabbar.Wheelable != nil {
			defaults.Tabbar.Wheelable = *compact.Tabbar.Wheelable
		}
		if compact.Tabbar.MaxCount != nil {
			defaults.Tabbar.MaxCount = *compact.Tabbar.MaxCount
		}
		if compact.Tabbar.MiddleClickToClose != nil {
			defaults.Tabbar.MiddleClickToClose = *compact.Tabbar.MiddleClickToClose
		}
	}

	if compact.Theme != nil {
		if compact.Theme.BuiltinType != nil {
			defaults.Theme.BuiltinType = *compact.Theme.BuiltinType
		}
		if compact.Theme.ColorPrimary != nil {
			defaults.Theme.ColorPrimary = *compact.Theme.ColorPrimary
		}
		if compact.Theme.Mode != nil {
			defaults.Theme.Mode = *compact.Theme.Mode
		}
		if compact.Theme.Radius != nil {
			defaults.Theme.Radius = *compact.Theme.Radius
		}
		if compact.Theme.FontSize != nil {
			defaults.Theme.FontSize = *compact.Theme.FontSize
		}
		if compact.Theme.SemiDarkHeader != nil {
			defaults.Theme.SemiDarkHeader = *compact.Theme.SemiDarkHeader
		}
		if compact.Theme.SemiDarkSidebar != nil {
			defaults.Theme.SemiDarkSidebar = *compact.Theme.SemiDarkSidebar
		}
		if compact.Theme.SemiDarkSidebarSub != nil {
			defaults.Theme.SemiDarkSidebarSub = *compact.Theme.SemiDarkSidebarSub
		}
	}

	if compact.Transition != nil {
		if compact.Transition.Enable != nil {
			defaults.Transition.Enable = *compact.Transition.Enable
		}
		if compact.Transition.Loading != nil {
			defaults.Transition.Loading = *compact.Transition.Loading
		}
		if compact.Transition.Name != nil {
			defaults.Transition.Name = *compact.Transition.Name
		}
		if compact.Transition.Progress != nil {
			defaults.Transition.Progress = *compact.Transition.Progress
		}
	}

	if compact.Widget != nil {
		if compact.Widget.Fullscreen != nil {
			defaults.Widget.Fullscreen = *compact.Widget.Fullscreen
		}
		if compact.Widget.GlobalSearch != nil {
			defaults.Widget.GlobalSearch = *compact.Widget.GlobalSearch
		}
		if compact.Widget.LanguageToggle != nil {
			defaults.Widget.LanguageToggle = *compact.Widget.LanguageToggle
		}
		if compact.Widget.LockScreen != nil {
			defaults.Widget.LockScreen = *compact.Widget.LockScreen
		}
		if compact.Widget.Notification != nil {
			defaults.Widget.Notification = *compact.Widget.Notification
		}
		if compact.Widget.Refresh != nil {
			defaults.Widget.Refresh = *compact.Widget.Refresh
		}
		if compact.Widget.SidebarToggle != nil {
			defaults.Widget.SidebarToggle = *compact.Widget.SidebarToggle
		}
		if compact.Widget.ThemeToggle != nil {
			defaults.Widget.ThemeToggle = *compact.Widget.ThemeToggle
		}
		if compact.Widget.Timezone != nil {
			defaults.Widget.Timezone = *compact.Widget.Timezone
		}
	}

	return sanitizeSystemUiPreferences(defaults), nil
}

func encodeSystemUiPreferences(preferences dto.SystemUiPreferences) (string, error) {
	defaults := defaultSystemUiPreferences()
	sanitized := sanitizeSystemUiPreferences(preferences)
	compact := compactSystemSettings{}

	if section := buildCompactAppSettings(defaults.App, sanitized.App); section != nil {
		compact.App = section
	}
	if section := buildCompactBreadcrumbSettings(defaults.Breadcrumb, sanitized.Breadcrumb); section != nil {
		compact.Breadcrumb = section
	}
	if section := buildCompactCopyrightSettings(defaults.Copyright, sanitized.Copyright); section != nil {
		compact.Copyright = section
	}
	if section := buildCompactFooterSettings(defaults.Footer, sanitized.Footer); section != nil {
		compact.Footer = section
	}
	if section := buildCompactHeaderSettings(defaults.Header, sanitized.Header); section != nil {
		compact.Header = section
	}
	if section := buildCompactNavigationSettings(defaults.Navigation, sanitized.Navigation); section != nil {
		compact.Navigation = section
	}
	if section := buildCompactShortcutKeySettings(defaults.ShortcutKeys, sanitized.ShortcutKeys); section != nil {
		compact.ShortcutKey = section
	}
	if section := buildCompactSidebarSettings(defaults.Sidebar, sanitized.Sidebar); section != nil {
		compact.Sidebar = section
	}
	if section := buildCompactTabbarSettings(defaults.Tabbar, sanitized.Tabbar); section != nil {
		compact.Tabbar = section
	}
	if section := buildCompactThemeSettings(defaults.Theme, sanitized.Theme); section != nil {
		compact.Theme = section
	}
	if section := buildCompactTransitionSettings(defaults.Transition, sanitized.Transition); section != nil {
		compact.Transition = section
	}
	if section := buildCompactWidgetSettings(defaults.Widget, sanitized.Widget); section != nil {
		compact.Widget = section
	}

	encoded, err := json.Marshal(compact)
	if err != nil {
		return "", err
	}

	if len(encoded) > 255 {
		return "", errors.New("界面设置过多，超过单项配置存储上限")
	}

	return string(encoded), nil
}

func buildCompactAppSettings(defaults dto.SystemUiAppSettings, current dto.SystemUiAppSettings) *compactSystemAppSettings {
	compact := &compactSystemAppSettings{}
	if current.ColorGrayMode != defaults.ColorGrayMode {
		compact.ColorGrayMode = boolPtr(current.ColorGrayMode)
	}
	if current.ColorWeakMode != defaults.ColorWeakMode {
		compact.ColorWeakMode = boolPtr(current.ColorWeakMode)
	}
	if current.ContentCompact != defaults.ContentCompact {
		compact.ContentCompact = stringPtr(current.ContentCompact)
	}
	if current.DynamicTitle != defaults.DynamicTitle {
		compact.DynamicTitle = boolPtr(current.DynamicTitle)
	}
	if current.EnableCheckUpdates != defaults.EnableCheckUpdates {
		compact.EnableCheckUpdate = boolPtr(current.EnableCheckUpdates)
	}
	if current.Layout != defaults.Layout {
		compact.Layout = stringPtr(current.Layout)
	}
	if current.LoginTitle != defaults.LoginTitle {
		compact.LoginTitle = stringPtr(current.LoginTitle)
	}
	if current.LoginDescription != defaults.LoginDescription {
		compact.LoginDescription = stringPtr(current.LoginDescription)
	}
	if current.Locale != defaults.Locale {
		compact.Locale = stringPtr(current.Locale)
	}
	if current.Watermark != defaults.Watermark {
		compact.Watermark = boolPtr(current.Watermark)
	}
	if current.WatermarkContent != defaults.WatermarkContent {
		compact.WatermarkContent = stringPtr(current.WatermarkContent)
	}
	if isCompactSectionEmpty(compact) {
		return nil
	}
	return compact
}

func buildCompactBreadcrumbSettings(defaults dto.SystemUiBreadcrumbSettings, current dto.SystemUiBreadcrumbSettings) *compactSystemBreadcrumbSettings {
	compact := &compactSystemBreadcrumbSettings{}
	if current.Enable != defaults.Enable {
		compact.Enable = boolPtr(current.Enable)
	}
	if current.HideOnlyOne != defaults.HideOnlyOne {
		compact.HideOnlyOne = boolPtr(current.HideOnlyOne)
	}
	if current.ShowHome != defaults.ShowHome {
		compact.ShowHome = boolPtr(current.ShowHome)
	}
	if current.ShowIcon != defaults.ShowIcon {
		compact.ShowIcon = boolPtr(current.ShowIcon)
	}
	if current.StyleType != defaults.StyleType {
		compact.StyleType = stringPtr(current.StyleType)
	}
	if isCompactSectionEmpty(compact) {
		return nil
	}
	return compact
}

func buildCompactCopyrightSettings(defaults dto.SystemUiCopyrightSettings, current dto.SystemUiCopyrightSettings) *compactSystemCopyrightSettings {
	compact := &compactSystemCopyrightSettings{}
	if current.CompanyName != defaults.CompanyName {
		compact.CompanyName = stringPtr(current.CompanyName)
	}
	if current.CompanySiteLink != defaults.CompanySiteLink {
		compact.CompanySiteLink = stringPtr(current.CompanySiteLink)
	}
	if current.Date != defaults.Date {
		compact.Date = stringPtr(current.Date)
	}
	if current.Enable != defaults.Enable {
		compact.Enable = boolPtr(current.Enable)
	}
	if current.Icp != defaults.Icp {
		compact.Icp = stringPtr(current.Icp)
	}
	if current.IcpLink != defaults.IcpLink {
		compact.IcpLink = stringPtr(current.IcpLink)
	}
	if isCompactSectionEmpty(compact) {
		return nil
	}
	return compact
}

func buildCompactFooterSettings(defaults dto.SystemUiFooterSettings, current dto.SystemUiFooterSettings) *compactSystemFooterSettings {
	compact := &compactSystemFooterSettings{}
	if current.Enable != defaults.Enable {
		compact.Enable = boolPtr(current.Enable)
	}
	if current.Fixed != defaults.Fixed {
		compact.Fixed = boolPtr(current.Fixed)
	}
	if isCompactSectionEmpty(compact) {
		return nil
	}
	return compact
}

func buildCompactHeaderSettings(defaults dto.SystemUiHeaderSettings, current dto.SystemUiHeaderSettings) *compactSystemHeaderSettings {
	compact := &compactSystemHeaderSettings{}
	if current.Enable != defaults.Enable {
		compact.Enable = boolPtr(current.Enable)
	}
	if current.Hidden != defaults.Hidden {
		compact.Hidden = boolPtr(current.Hidden)
	}
	if current.MenuAlign != defaults.MenuAlign {
		compact.MenuAlign = stringPtr(current.MenuAlign)
	}
	if current.Mode != defaults.Mode {
		compact.Mode = stringPtr(current.Mode)
	}
	if isCompactSectionEmpty(compact) {
		return nil
	}
	return compact
}

func buildCompactNavigationSettings(defaults dto.SystemUiNavigationSettings, current dto.SystemUiNavigationSettings) *compactSystemNavigationSettings {
	compact := &compactSystemNavigationSettings{}
	if current.Accordion != defaults.Accordion {
		compact.Accordion = boolPtr(current.Accordion)
	}
	if current.Split != defaults.Split {
		compact.Split = boolPtr(current.Split)
	}
	if current.StyleType != defaults.StyleType {
		compact.StyleType = stringPtr(current.StyleType)
	}
	if isCompactSectionEmpty(compact) {
		return nil
	}
	return compact
}

func buildCompactShortcutKeySettings(defaults dto.SystemUiShortcutKeySettings, current dto.SystemUiShortcutKeySettings) *compactSystemShortcutKeySettings {
	compact := &compactSystemShortcutKeySettings{}
	if current.Enable != defaults.Enable {
		compact.Enable = boolPtr(current.Enable)
	}
	if current.GlobalSearch != defaults.GlobalSearch {
		compact.GlobalSearch = boolPtr(current.GlobalSearch)
	}
	if current.GlobalLogout != defaults.GlobalLogout {
		compact.GlobalLogout = boolPtr(current.GlobalLogout)
	}
	if current.GlobalLockScreen != defaults.GlobalLockScreen {
		compact.GlobalLockScreen = boolPtr(current.GlobalLockScreen)
	}
	if isCompactSectionEmpty(compact) {
		return nil
	}
	return compact
}

func buildCompactSidebarSettings(defaults dto.SystemUiSidebarSettings, current dto.SystemUiSidebarSettings) *compactSystemSidebarSettings {
	compact := &compactSystemSidebarSettings{}
	if current.AutoActivateChild != defaults.AutoActivateChild {
		compact.AutoActivateChild = boolPtr(current.AutoActivateChild)
	}
	if current.Collapsed != defaults.Collapsed {
		compact.Collapsed = boolPtr(current.Collapsed)
	}
	if current.CollapsedButton != defaults.CollapsedButton {
		compact.CollapsedButton = boolPtr(current.CollapsedButton)
	}
	if current.CollapsedShowTitle != defaults.CollapsedShowTitle {
		compact.CollapsedShowText = boolPtr(current.CollapsedShowTitle)
	}
	if current.Draggable != defaults.Draggable {
		compact.Draggable = boolPtr(current.Draggable)
	}
	if current.Enable != defaults.Enable {
		compact.Enable = boolPtr(current.Enable)
	}
	if current.ExpandOnHover != defaults.ExpandOnHover {
		compact.ExpandOnHover = boolPtr(current.ExpandOnHover)
	}
	if current.FixedButton != defaults.FixedButton {
		compact.FixedButton = boolPtr(current.FixedButton)
	}
	if current.Width != defaults.Width {
		compact.Width = intPtr(current.Width)
	}
	if isCompactSectionEmpty(compact) {
		return nil
	}
	return compact
}

func buildCompactTabbarSettings(defaults dto.SystemUiTabbarSettings, current dto.SystemUiTabbarSettings) *compactSystemTabbarSettings {
	compact := &compactSystemTabbarSettings{}
	if current.Draggable != defaults.Draggable {
		compact.Draggable = boolPtr(current.Draggable)
	}
	if current.Enable != defaults.Enable {
		compact.Enable = boolPtr(current.Enable)
	}
	if current.Persist != defaults.Persist {
		compact.Persist = boolPtr(current.Persist)
	}
	if current.VisitHistory != defaults.VisitHistory {
		compact.VisitHistory = boolPtr(current.VisitHistory)
	}
	if current.ShowIcon != defaults.ShowIcon {
		compact.ShowIcon = boolPtr(current.ShowIcon)
	}
	if current.ShowMaximize != defaults.ShowMaximize {
		compact.ShowMaximize = boolPtr(current.ShowMaximize)
	}
	if current.ShowMore != defaults.ShowMore {
		compact.ShowMore = boolPtr(current.ShowMore)
	}
	if current.StyleType != defaults.StyleType {
		compact.StyleType = stringPtr(current.StyleType)
	}
	if current.Wheelable != defaults.Wheelable {
		compact.Wheelable = boolPtr(current.Wheelable)
	}
	if current.MaxCount != defaults.MaxCount {
		compact.MaxCount = intPtr(current.MaxCount)
	}
	if current.MiddleClickToClose != defaults.MiddleClickToClose {
		compact.MiddleClickToClose = boolPtr(current.MiddleClickToClose)
	}
	if isCompactSectionEmpty(compact) {
		return nil
	}
	return compact
}

func buildCompactThemeSettings(defaults dto.SystemUiThemeSettings, current dto.SystemUiThemeSettings) *compactSystemThemeSettings {
	compact := &compactSystemThemeSettings{}
	if current.BuiltinType != defaults.BuiltinType {
		compact.BuiltinType = stringPtr(current.BuiltinType)
	}
	if current.ColorPrimary != defaults.ColorPrimary {
		compact.ColorPrimary = stringPtr(current.ColorPrimary)
	}
	if current.Mode != defaults.Mode {
		compact.Mode = stringPtr(current.Mode)
	}
	if current.Radius != defaults.Radius {
		compact.Radius = stringPtr(current.Radius)
	}
	if current.FontSize != defaults.FontSize {
		compact.FontSize = intPtr(current.FontSize)
	}
	if current.SemiDarkHeader != defaults.SemiDarkHeader {
		compact.SemiDarkHeader = boolPtr(current.SemiDarkHeader)
	}
	if current.SemiDarkSidebar != defaults.SemiDarkSidebar {
		compact.SemiDarkSidebar = boolPtr(current.SemiDarkSidebar)
	}
	if current.SemiDarkSidebarSub != defaults.SemiDarkSidebarSub {
		compact.SemiDarkSidebarSub = boolPtr(current.SemiDarkSidebarSub)
	}
	if isCompactSectionEmpty(compact) {
		return nil
	}
	return compact
}

func buildCompactTransitionSettings(defaults dto.SystemUiTransitionSettings, current dto.SystemUiTransitionSettings) *compactSystemTransitionSettings {
	compact := &compactSystemTransitionSettings{}
	if current.Enable != defaults.Enable {
		compact.Enable = boolPtr(current.Enable)
	}
	if current.Loading != defaults.Loading {
		compact.Loading = boolPtr(current.Loading)
	}
	if current.Name != defaults.Name {
		compact.Name = stringPtr(current.Name)
	}
	if current.Progress != defaults.Progress {
		compact.Progress = boolPtr(current.Progress)
	}
	if isCompactSectionEmpty(compact) {
		return nil
	}
	return compact
}

func buildCompactWidgetSettings(defaults dto.SystemUiWidgetSettings, current dto.SystemUiWidgetSettings) *compactSystemWidgetSettings {
	compact := &compactSystemWidgetSettings{}
	if current.Fullscreen != defaults.Fullscreen {
		compact.Fullscreen = boolPtr(current.Fullscreen)
	}
	if current.GlobalSearch != defaults.GlobalSearch {
		compact.GlobalSearch = boolPtr(current.GlobalSearch)
	}
	if current.LanguageToggle != defaults.LanguageToggle {
		compact.LanguageToggle = boolPtr(current.LanguageToggle)
	}
	if current.LockScreen != defaults.LockScreen {
		compact.LockScreen = boolPtr(current.LockScreen)
	}
	if current.Notification != defaults.Notification {
		compact.Notification = boolPtr(current.Notification)
	}
	if current.Refresh != defaults.Refresh {
		compact.Refresh = boolPtr(current.Refresh)
	}
	if current.SidebarToggle != defaults.SidebarToggle {
		compact.SidebarToggle = boolPtr(current.SidebarToggle)
	}
	if current.ThemeToggle != defaults.ThemeToggle {
		compact.ThemeToggle = boolPtr(current.ThemeToggle)
	}
	if current.Timezone != defaults.Timezone {
		compact.Timezone = boolPtr(current.Timezone)
	}
	if isCompactSectionEmpty(compact) {
		return nil
	}
	return compact
}

func (e *SysConfig) getSystemConfigByKey(configKey string) (*models.SysConfig, error) {
	var model models.SysConfig
	if err := e.Orm.Where("config_key = ?", configKey).First(&model).Error; err != nil {
		return nil, err
	}
	return &model, nil
}

func (e *SysConfig) getSystemConfigValue(configKey string) (string, error) {
	var model models.SysConfig
	if err := e.Orm.Where("config_key = ?", configKey).First(&model).Error; err != nil {
		return "", err
	}
	return model.ConfigValue, nil
}

func (e *SysConfig) upsertSystemConfigValue(configKey, configName, configValue, isFrontend, remark string) error {
	var model models.SysConfig
	err := e.Orm.Where("config_key = ?", configKey).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		model = models.SysConfig{
			ConfigKey:   configKey,
			ConfigName:  configName,
			ConfigType:  systemUIPreferencesConfigType,
			ConfigValue: configValue,
			IsFrontend:  isFrontend,
			Remark:      remark,
		}
		return e.Orm.Create(&model).Error
	}
	if err != nil {
		return err
	}

	model.ConfigName = configName
	model.ConfigValue = configValue
	model.ConfigType = systemUIPreferencesConfigType
	model.IsFrontend = isFrontend
	model.Remark = remark

	return e.Orm.Save(&model).Error
}

func boolPtr(v bool) *bool { return &v }

func intPtr(v int) *int { return &v }

func stringPtr(v string) *string { return &v }

func isCompactSectionEmpty(v interface{}) bool {
	raw, err := json.Marshal(v)
	if err != nil {
		return false
	}
	return string(raw) == "{}"
}
