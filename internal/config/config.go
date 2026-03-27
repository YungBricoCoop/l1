// SPDX-FileCopyrightText: 2026 Elwan Mayencourt <mayencourt@elwan.ch>
// SPDX-License-Identifier: MIT

package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/YungBricoCoop/l1/internal/theme"
	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/viper"
)

type Config struct {
	S3 S3Config `mapstructure:"s3" toml:"s3"`
	UI UIConfig `mapstructure:"ui" toml:"ui"`
	GI GIConfig `mapstructure:"gi" toml:"gi"`
}

type S3Config struct {
	URL           string `mapstructure:"url"            toml:"url"`
	Region        string `mapstructure:"region"         toml:"region"`
	AccessKey     string `mapstructure:"access_key"     toml:"access_key"`
	SecretKey     string `mapstructure:"secret_key"     toml:"secret_key"`
	DefaultBucket string `mapstructure:"default_bucket" toml:"default_bucket"`
}

type UIConfig struct {
	Color    bool   `mapstructure:"color"    toml:"color"`
	Progress bool   `mapstructure:"progress" toml:"progress"`
	Theme    string `mapstructure:"theme"    toml:"theme"`
}

type GIConfig struct {
	Templates []string `mapstructure:"templates" toml:"templates"`
}

const (
	keyTypeString      = "string"
	keyTypeBool        = "bool"
	keyTypeStringSlice = "string_slice"
)

func DefaultConfig() Config {
	return Config{
		S3: S3Config{
			URL:           "",
			Region:        "",
			AccessKey:     "",
			SecretKey:     "",
			DefaultBucket: "",
		},
		UI: UIConfig{
			Color:    true,
			Progress: true,
			Theme:    theme.DefaultName(),
		},
		GI: GIConfig{
			Templates: []string{},
		},
	}
}

func ResolvePath(override string) (string, error) {
	if strings.TrimSpace(override) != "" {
		return override, nil
	}
	return DefaultPath()
}

func DefaultPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config dir: %w", err)
	}

	return filepath.Join(configDir, "l1", "config.toml"), nil
}

func Load(path string) (Config, error) {
	v := viper.New()
	v.SetConfigType("toml")
	v.SetConfigFile(path)
	v.SetDefault("ui.color", true)
	v.SetDefault("ui.progress", true)
	v.SetDefault("ui.theme", theme.DefaultName())
	v.SetDefault("gi.templates", []string{})

	if err := v.ReadInConfig(); err != nil {
		if os.IsNotExist(err) {
			return Config{}, fmt.Errorf("config file not found at %s", path)
		}

		var notFoundErr viper.ConfigFileNotFoundError
		if errors.As(err, &notFoundErr) {
			return Config{}, fmt.Errorf("config file not found at %s", path)
		}

		return Config{}, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("decode config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func LoadForEdit(path string) (Config, error) {
	cfg := DefaultConfig()

	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return Config{}, fmt.Errorf("read config file: %w", err)
	}

	if len(strings.TrimSpace(string(content))) == 0 {
		return cfg, nil
	}

	decodeErr := toml.Unmarshal(content, &cfg)
	if decodeErr != nil {
		return Config{}, fmt.Errorf("decode config file: %w", decodeErr)
	}

	validateErr := cfg.Validate()
	if validateErr != nil {
		return Config{}, validateErr
	}

	return cfg, nil
}

func Save(path string, cfg Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	encoded, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config file: %w", err)
	}

	writeErr := os.WriteFile(path, encoded, 0o600)
	if writeErr != nil {
		return fmt.Errorf("write config file: %w", writeErr)
	}

	return nil
}

func InitFile(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return false, nil
	}
	if !os.IsNotExist(err) {
		return false, fmt.Errorf("stat config file: %w", err)
	}

	saveErr := Save(path, DefaultConfig())
	if saveErr != nil {
		return false, saveErr
	}

	return true, nil
}

func BackupFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("config file not found at %s", path)
		}
		return "", fmt.Errorf("read config file: %w", err)
	}

	backupPath := path + ".bkp"
	writeErr := os.WriteFile(backupPath, content, 0o600)
	if writeErr != nil {
		return "", fmt.Errorf("write backup file: %w", writeErr)
	}

	return backupPath, nil
}

func RestoreBackupFile(path string) (string, error) {
	backupPath := path + ".bkp"

	content, err := os.ReadFile(backupPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("backup file not found at %s", backupPath)
		}
		return "", fmt.Errorf("read backup file: %w", err)
	}

	mkdirErr := os.MkdirAll(filepath.Dir(path), 0o750)
	if mkdirErr != nil {
		return "", fmt.Errorf("create config directory: %w", mkdirErr)
	}

	writeErr := os.WriteFile(path, content, 0o600)
	if writeErr != nil {
		return "", fmt.Errorf("write config file: %w", writeErr)
	}

	return backupPath, nil
}

func ParseValueForKey(key, raw string) (any, error) {
	typeName, ok := configKeyType(key)
	if !ok {
		return nil, fmt.Errorf("unsupported config key %q", key)
	}

	switch typeName {
	case keyTypeString:
		return raw, nil
	case keyTypeBool:
		v, err := strconv.ParseBool(raw)
		if err != nil {
			return nil, fmt.Errorf("invalid bool for %s: %q", key, raw)
		}
		return v, nil
	case keyTypeStringSlice:
		return parseStringSlice(raw), nil
	default:
		return nil, fmt.Errorf("unsupported config type %q for key %q", typeName, key)
	}
}

func SetConfigValue(cfg *Config, key string, value any) error {
	if cfg == nil {
		return errors.New("config cannot be nil")
	}

	handled, err := setS3ConfigValue(cfg, key, value)
	if handled {
		return err
	}

	handled, err = setUIConfigValue(cfg, key, value)
	if handled {
		return err
	}

	handled, err = setGIConfigValue(cfg, key, value)
	if handled {
		return err
	}

	return fmt.Errorf("unsupported config key %q", key)
}

func setS3ConfigValue(cfg *Config, key string, value any) (bool, error) {
	switch key {
	case "s3.url":
		strValue, err := valueAsString(key, value)
		if err != nil {
			return true, err
		}
		cfg.S3.URL = strValue
	case "s3.region":
		strValue, err := valueAsString(key, value)
		if err != nil {
			return true, err
		}
		cfg.S3.Region = strValue
	case "s3.access_key":
		strValue, err := valueAsString(key, value)
		if err != nil {
			return true, err
		}
		cfg.S3.AccessKey = strValue
	case "s3.secret_key":
		strValue, err := valueAsString(key, value)
		if err != nil {
			return true, err
		}
		cfg.S3.SecretKey = strValue
	case "s3.default_bucket":
		strValue, err := valueAsString(key, value)
		if err != nil {
			return true, err
		}
		cfg.S3.DefaultBucket = strValue
	default:
		return false, nil
	}

	return true, nil
}

func setUIConfigValue(cfg *Config, key string, value any) (bool, error) {
	switch key {
	case "ui.color":
		boolValue, err := valueAsBool(key, value)
		if err != nil {
			return true, err
		}
		cfg.UI.Color = boolValue
		return true, nil
	case "ui.progress":
		boolValue, err := valueAsBool(key, value)
		if err != nil {
			return true, err
		}
		cfg.UI.Progress = boolValue
		return true, nil
	case "ui.theme":
		strValue, err := valueAsString(key, value)
		if err != nil {
			return true, err
		}

		canonicalTheme, err := theme.CanonicalName(strValue)
		if err != nil {
			return true, err
		}

		cfg.UI.Theme = canonicalTheme
		return true, nil
	default:
		return false, nil
	}
}

func setGIConfigValue(cfg *Config, key string, value any) (bool, error) {
	if key != "gi.templates" {
		return false, nil
	}

	sliceValue, err := valueAsStringSlice(key, value)
	if err != nil {
		return true, err
	}

	cfg.GI.Templates = sliceValue
	return true, nil
}

func valueAsString(key string, value any) (string, error) {
	v, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("invalid value type for %s", key)
	}
	return v, nil
}

func valueAsBool(key string, value any) (bool, error) {
	v, ok := value.(bool)
	if !ok {
		return false, fmt.Errorf("invalid value type for %s", key)
	}
	return v, nil
}

func valueAsStringSlice(key string, value any) ([]string, error) {
	v, ok := value.([]string)
	if !ok {
		return nil, fmt.Errorf("invalid value type for %s", key)
	}
	return v, nil
}

func ResolveSecretValue(value string) (string, error) {
	if value == "" {
		return "", nil
	}

	if !strings.HasPrefix(value, "env:") {
		return value, nil
	}

	envKey := strings.TrimSpace(strings.TrimPrefix(value, "env:"))
	if envKey == "" {
		return "", fmt.Errorf("invalid env secret reference %q", value)
	}

	resolved, ok := os.LookupEnv(envKey)
	if !ok {
		return "", fmt.Errorf("environment variable %s is not set", envKey)
	}

	return resolved, nil
}

func (c Config) Validate() error {
	if _, err := theme.CanonicalName(c.UI.Theme); err != nil {
		return err
	}

	if c.S3.URL != "" {
		u, err := url.Parse(c.S3.URL)
		if err != nil {
			return fmt.Errorf("invalid s3.url: %w", err)
		}
		if u.Scheme != "http" && u.Scheme != "https" {
			return fmt.Errorf("invalid s3.url scheme %q: expected http or https", u.Scheme)
		}
		if u.Host == "" {
			return errors.New("invalid s3.url: missing host")
		}
	}

	return nil
}

func (c Config) ValidateForPush(bucketOverride string) error {
	if err := c.Validate(); err != nil {
		return err
	}

	bucket := strings.TrimSpace(bucketOverride)
	if bucket == "" {
		bucket = strings.TrimSpace(c.S3.DefaultBucket)
	}
	if bucket == "" {
		return errors.New("missing bucket: set s3.default_bucket or pass --bucket")
	}

	return nil
}

func configKeyType(key string) (string, bool) {
	switch key {
	case "s3.url", "s3.region", "s3.access_key", "s3.secret_key", "s3.default_bucket":
		return keyTypeString, true
	case "ui.theme":
		return keyTypeString, true
	case "ui.color", "ui.progress":
		return keyTypeBool, true
	case "gi.templates":
		return keyTypeStringSlice, true
	default:
		return "", false
	}
}

func parseStringSlice(raw string) []string {
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\t' || r == '\n'
	})
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value == "" {
			continue
		}
		values = append(values, value)
	}

	return values
}
