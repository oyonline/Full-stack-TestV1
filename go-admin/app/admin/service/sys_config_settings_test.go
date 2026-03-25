package service

import (
	"testing"

	"go-admin/app/admin/service/dto"
)

func TestEncodeDecodeSystemUIPreferences(t *testing.T) {
	input := defaultSystemUiPreferences()
	input.App.Layout = "header-nav"
	input.App.Locale = "en-US"
	input.Theme.Mode = "light"
	input.Theme.BuiltinType = "green"
	input.Sidebar.Width = 200
	input.Widget.Timezone = false

	encoded, err := encodeSystemUiPreferences(input)
	if err != nil {
		t.Fatalf("encodeSystemUiPreferences returned error: %v", err)
	}

	decoded, err := decodeSystemUiPreferences(encoded)
	if err != nil {
		t.Fatalf("decodeSystemUiPreferences returned error: %v", err)
	}

	if decoded.App.Layout != input.App.Layout {
		t.Fatalf("expected layout %q, got %q", input.App.Layout, decoded.App.Layout)
	}
	if decoded.App.Locale != input.App.Locale {
		t.Fatalf("expected locale %q, got %q", input.App.Locale, decoded.App.Locale)
	}
	if decoded.Theme.Mode != input.Theme.Mode {
		t.Fatalf("expected theme mode %q, got %q", input.Theme.Mode, decoded.Theme.Mode)
	}
	if decoded.Theme.BuiltinType != input.Theme.BuiltinType {
		t.Fatalf("expected builtin theme %q, got %q", input.Theme.BuiltinType, decoded.Theme.BuiltinType)
	}
	if decoded.Sidebar.Width != input.Sidebar.Width {
		t.Fatalf("expected sidebar width %d, got %d", input.Sidebar.Width, decoded.Sidebar.Width)
	}
	if decoded.Widget.Timezone != input.Widget.Timezone {
		t.Fatalf("expected timezone %v, got %v", input.Widget.Timezone, decoded.Widget.Timezone)
	}
}

func TestSanitizeSystemSettingsPayload(t *testing.T) {
	payload := dto.SystemSettingsPayload{
		Branding: dto.SystemBrandingSettings{
			AppLogo: "https://example.com/logo.png",
			AppName: "Console",
		},
		UIPreferences: dto.SystemUiPreferences{
			App:          dto.SystemUiAppSettings{},
			Breadcrumb:   dto.SystemUiBreadcrumbSettings{},
			Copyright:    dto.SystemUiCopyrightSettings{},
			Footer:       dto.SystemUiFooterSettings{},
			Header:       dto.SystemUiHeaderSettings{},
			Navigation:   dto.SystemUiNavigationSettings{},
			ShortcutKeys: dto.SystemUiShortcutKeySettings{},
			Sidebar:      dto.SystemUiSidebarSettings{},
			Tabbar:       dto.SystemUiTabbarSettings{},
			Theme:        dto.SystemUiThemeSettings{},
			Transition:   dto.SystemUiTransitionSettings{},
			Widget:       dto.SystemUiWidgetSettings{},
		},
	}

	sanitized := sanitizeSystemSettingsPayload(payload)

	if sanitized.UIPreferences.App.Layout == "" {
		t.Fatal("expected sanitized app layout to fall back to default")
	}
	if sanitized.UIPreferences.Theme.Mode == "" {
		t.Fatal("expected sanitized theme mode to fall back to default")
	}
	if sanitized.UIPreferences.Sidebar.Width <= 0 {
		t.Fatal("expected sanitized sidebar width to fall back to default")
	}
}
