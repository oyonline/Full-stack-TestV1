package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	sdkConfig "github.com/go-admin-team/go-admin-core/sdk/config"
	"gopkg.in/yaml.v3"
)

type localSettingsFile struct {
	Settings localSettings `yaml:"settings"`
}

type localSettings struct {
	Application *localApplication `yaml:"application"`
	Jwt         *localJWT         `yaml:"jwt"`
	Database    *localDatabase    `yaml:"database"`
	Gen         *localGen         `yaml:"gen"`
}

type localApplication struct {
	Host string `yaml:"host"`
	Mode string `yaml:"mode"`
	Name string `yaml:"name"`
	Port int64  `yaml:"port"`
}

type localJWT struct {
	Secret  string `yaml:"secret"`
	Timeout int64  `yaml:"timeout"`
}

type localDatabase struct {
	Driver string `yaml:"driver"`
	Source string `yaml:"source"`
}

type localGen struct {
	FrontPath string `yaml:"frontpath"`
}

// ApplyLocalOverrides returns a callback that mutates the loaded config
// before the database/storage initialization callbacks run.
func ApplyLocalOverrides(configPath string) func() {
	return func() {
		applyLocalConfigFile(configPath)
		applyEnvOverrides()
	}
}

func applyLocalConfigFile(configPath string) {
	localPath := buildLocalConfigPath(configPath)
	data, err := os.ReadFile(localPath)
	if err != nil {
		return
	}

	var cfg localSettingsFile
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return
	}

	if app := cfg.Settings.Application; app != nil {
		if app.Host != "" {
			sdkConfig.ApplicationConfig.Host = app.Host
		}
		if app.Mode != "" {
			sdkConfig.ApplicationConfig.Mode = app.Mode
		}
		if app.Name != "" {
			sdkConfig.ApplicationConfig.Name = app.Name
		}
		if app.Port != 0 {
			sdkConfig.ApplicationConfig.Port = app.Port
		}
	}

	if jwt := cfg.Settings.Jwt; jwt != nil {
		if jwt.Secret != "" {
			sdkConfig.JwtConfig.Secret = jwt.Secret
		}
		if jwt.Timeout != 0 {
			sdkConfig.JwtConfig.Timeout = jwt.Timeout
		}
	}

	if db := cfg.Settings.Database; db != nil {
		overrideDatabase(db.Driver, db.Source)
	}

	if gen := cfg.Settings.Gen; gen != nil {
		if gen.FrontPath != "" {
			sdkConfig.GenConfig.FrontPath = gen.FrontPath
		}
	}
}

func applyEnvOverrides() {
	if value := strings.TrimSpace(os.Getenv("GO_ADMIN_APP_HOST")); value != "" {
		sdkConfig.ApplicationConfig.Host = value
	}
	if value := strings.TrimSpace(os.Getenv("GO_ADMIN_APP_MODE")); value != "" {
		sdkConfig.ApplicationConfig.Mode = value
	}
	if value := strings.TrimSpace(os.Getenv("GO_ADMIN_APP_NAME")); value != "" {
		sdkConfig.ApplicationConfig.Name = value
	}
	if value := strings.TrimSpace(os.Getenv("GO_ADMIN_APP_PORT")); value != "" {
		if port, err := strconv.ParseInt(value, 10, 64); err == nil {
			sdkConfig.ApplicationConfig.Port = port
		}
	}
	if value := strings.TrimSpace(os.Getenv("GO_ADMIN_JWT_SECRET")); value != "" {
		sdkConfig.JwtConfig.Secret = value
	}
	if value := strings.TrimSpace(os.Getenv("GO_ADMIN_JWT_TIMEOUT")); value != "" {
		if timeout, err := strconv.ParseInt(value, 10, 64); err == nil {
			sdkConfig.JwtConfig.Timeout = timeout
		}
	}
	if value := strings.TrimSpace(os.Getenv("GO_ADMIN_DB_DRIVER")); value != "" {
		overrideDatabase(value, "")
	}
	if value := strings.TrimSpace(os.Getenv("GO_ADMIN_DB_SOURCE")); value != "" {
		overrideDatabase("", value)
	}
	if value := strings.TrimSpace(os.Getenv("GO_ADMIN_GEN_FRONTPATH")); value != "" {
		sdkConfig.GenConfig.FrontPath = value
	}
}

func overrideDatabase(driver string, source string) {
	if driver != "" {
		sdkConfig.DatabaseConfig.Driver = driver
	}
	if source != "" {
		sdkConfig.DatabaseConfig.Source = source
	}

	if db, ok := sdkConfig.DatabasesConfig["*"]; ok && db != nil {
		if driver != "" {
			db.Driver = driver
		}
		if source != "" {
			db.Source = source
		}
	}
}

func buildLocalConfigPath(configPath string) string {
	ext := filepath.Ext(configPath)
	if ext == "" {
		return configPath + ".local"
	}
	return strings.TrimSuffix(configPath, ext) + ".local" + ext
}
