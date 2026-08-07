// Package config loads catfu configuration from flags, env, and file.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Config is the effective runtime configuration.
type Config struct {
	DBPath      string  `mapstructure:"db"`
	BraveAPIKey string  `mapstructure:"brave_api_key"` // Brave Search plan X-Subscription-Token
	YTDLP       string  `mapstructure:"ytdlp"`
	LogLevel    string  `mapstructure:"log_level"`
	JSON        bool    `mapstructure:"json"`
	Format      string  `mapstructure:"format"`
	Quiet       bool    `mapstructure:"quiet"`
	ConfigFile  string  `mapstructure:"config"`
	SleepReq    float64 `mapstructure:"sleep_requests"`
	SleepMin    float64 `mapstructure:"sleep_interval"`
	SleepMax    float64 `mapstructure:"max_sleep_interval"`
}

// Defaults returns a Config with built-in defaults (paths resolved).
func Defaults() Config {
	return Config{
		DBPath:   defaultDBPath(),
		YTDLP:    "yt-dlp",
		LogLevel: "info",
		Format:   "table",
		SleepReq: 0.5,
		SleepMin: 1,
		SleepMax: 3,
	}
}

func defaultDBPath() string {
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, "catfu", "catfu.db")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "catfu.db"
	}
	return filepath.Join(home, ".local", "share", "catfu", "catfu.db")
}

// ConfigDir returns the XDG-style config directory for catfu.
func ConfigDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "catfu")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "catfu"
	}
	return filepath.Join(home, ".config", "catfu")
}

// InitViper sets up viper with env bindings and optional config file.
func InitViper(cfgFile string) (*viper.Viper, error) {
	v := viper.New()
	d := Defaults()
	v.SetDefault("db", d.DBPath)
	v.SetDefault("ytdlp", d.YTDLP)
	v.SetDefault("log_level", d.LogLevel)
	v.SetDefault("format", d.Format)
	v.SetDefault("json", false)
	v.SetDefault("quiet", false)
	v.SetDefault("sleep_requests", d.SleepReq)
	v.SetDefault("sleep_interval", d.SleepMin)
	v.SetDefault("max_sleep_interval", d.SleepMax)
	v.SetDefault("brave_api_key", "")

	v.SetEnvPrefix("CATFU")
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	v.AutomaticEnv()
	_ = v.BindEnv("brave_api_key", "BRAVE_API_KEY", "CATFU_BRAVE_API_KEY")
	_ = v.BindEnv("db", "CATFU_DB")
	_ = v.BindEnv("ytdlp", "CATFU_YTDLP")
	_ = v.BindEnv("log_level", "CATFU_LOG_LEVEL")

	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("read config: %w", err)
		}
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(ConfigDir())
		v.AddConfigPath(".")
		_ = v.ReadInConfig()
	}
	return v, nil
}

// Load merges viper into Config.
func Load(v *viper.Viper) (Config, error) {
	var c Config
	if err := v.Unmarshal(&c); err != nil {
		return c, err
	}
	if c.Format == "" {
		c.Format = "table"
	}
	if c.JSON {
		c.Format = "json"
	}
	if c.YTDLP == "" {
		c.YTDLP = "yt-dlp"
	}
	if c.DBPath == "" {
		c.DBPath = defaultDBPath()
	}
	return c, nil
}

// Redacted returns a copy safe for display (secrets masked).
func (c Config) Redacted() map[string]any {
	keyStatus := "(not set)"
	if c.BraveAPIKey != "" {
		keyStatus = "***"
	}
	return map[string]any{
		"db":                 c.DBPath,
		"ytdlp":              c.YTDLP,
		"log_level":          c.LogLevel,
		"format":             c.Format,
		"json":               c.JSON,
		"quiet":              c.Quiet,
		"brave_api_key":      keyStatus,
		"brave_plan":         "Search",
		"sleep_requests":     c.SleepReq,
		"sleep_interval":     c.SleepMin,
		"max_sleep_interval": c.SleepMax,
		"config_file":        c.ConfigFile,
	}
}

// EnsureDBDir creates the parent directory for the database file.
func EnsureDBDir(dbPath string) error {
	dir := filepath.Dir(dbPath)
	if dir == "" || dir == "." {
		return nil
	}
	return os.MkdirAll(dir, 0o755)
}
