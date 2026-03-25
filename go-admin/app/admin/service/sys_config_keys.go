package service

var settingsConfigKeys = []string{
	"sys_app_name",
	"sys_app_logo",
	"sys_ui_preferences",
}

var protectedConfigKeySet = map[string]struct{}{
	"sys_app_name":       {},
	"sys_app_logo":       {},
	"sys_ui_preferences": {},
}

func isProtectedConfigKey(key string) bool {
	_, ok := protectedConfigKeySet[key]
	return ok
}
