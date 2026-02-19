package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const (
	EnvURL      = "GRAYLOGCTL_URL"
	EnvAPIBase  = "GRAYLOGCTL_API_BASE"
	EnvToken    = "GRAYLOGCTL_TOKEN"
	EnvSession  = "GRAYLOGCTL_SESSION"
	EnvInsecure = "GRAYLOGCTL_INSECURE"
	EnvTimeout  = "GRAYLOGCTL_TIMEOUT"
	EnvFormat   = "GRAYLOGCTL_FORMAT"
	EnvProfile  = "GRAYLOGCTL_PROFILE"

	DefaultProfile = "default"
	DefaultAPIBase = "/api"
	DefaultTimeout = 30 * time.Second
	DefaultFormat  = "table"
)

type Config struct {
	Profiles map[string]Profile `yaml:"profiles"`
}

type Profile struct {
	URL      string      `yaml:"url"`
	APIBase  string      `yaml:"api_base"`
	Insecure bool        `yaml:"insecure"`
	Auth     ProfileAuth `yaml:"auth"`
}

type ProfileAuth struct {
	Token   string `yaml:"token"`
	Session string `yaml:"session"`
}

type Runtime struct {
	URL      string
	APIBase  string
	Token    string
	Session  string
	Insecure bool
	Timeout  time.Duration
	Format   string
	Profile  string
	MaxWidth int
}

func ConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}
	return filepath.Join(home, ".config", "graylogctl", "config.yaml"), nil
}

func DefaultConfig() *Config {
	return &Config{Profiles: map[string]Profile{DefaultProfile: {
		URL:     "",
		APIBase: DefaultAPIBase,
		Auth:    ProfileAuth{},
	}}}
}

func LoadConfig() (*Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return DefaultConfig(), nil
		}
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}
	cfg := DefaultConfig()
	if err := yaml.Unmarshal(b, cfg); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}
	if cfg.Profiles == nil {
		cfg.Profiles = map[string]Profile{}
	}
	if _, ok := cfg.Profiles[DefaultProfile]; !ok {
		cfg.Profiles[DefaultProfile] = Profile{APIBase: DefaultAPIBase, Auth: ProfileAuth{}}
	}
	return cfg, nil
}

func SaveConfig(cfg *Config) error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}
	if cfg.Profiles == nil {
		cfg.Profiles = map[string]Profile{}
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create config dir %s: %w", dir, err)
	}
	b, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err := os.WriteFile(path, b, 0o600); err != nil {
		return fmt.Errorf("write config %s: %w", path, err)
	}
	return nil
}

func Resolve(cmd *cobra.Command, cfg *Config) (Runtime, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	profile := chooseString(cmd, "profile", EnvProfile, "", DefaultProfile)
	p := cfg.Profiles[profile]

	url := chooseString(cmd, "url", EnvURL, p.URL, "")
	apiBase := chooseString(cmd, "api-base", EnvAPIBase, p.APIBase, DefaultAPIBase)
	token := chooseString(cmd, "token", EnvToken, p.Auth.Token, "")
	session := chooseString(cmd, "session", EnvSession, p.Auth.Session, "")
	format := strings.ToLower(chooseString(cmd, "format", EnvFormat, "", DefaultFormat))
	if format != "table" && format != "json" {
		return Runtime{}, fmt.Errorf("unsupported --format %q (use table|json)", format)
	}

	insecure, err := chooseBool(cmd, "insecure", EnvInsecure, p.Insecure, false)
	if err != nil {
		return Runtime{}, err
	}
	timeout, err := chooseDuration(cmd, "timeout", EnvTimeout, DefaultTimeout)
	if err != nil {
		return Runtime{}, err
	}
	maxWidth, err := cmd.Flags().GetInt("max-width")
	if err != nil {
		return Runtime{}, fmt.Errorf("read --max-width: %w", err)
	}

	return Runtime{
		URL:      url,
		APIBase:  apiBase,
		Token:    token,
		Session:  session,
		Insecure: insecure,
		Timeout:  timeout,
		Format:   format,
		Profile:  profile,
		MaxWidth: maxWidth,
	}, nil
}

func chooseString(cmd *cobra.Command, flagName, envName, profileVal, fallback string) string {
	if cmd.Flags().Changed(flagName) {
		v, _ := cmd.Flags().GetString(flagName)
		return strings.TrimSpace(v)
	}
	if v, ok := os.LookupEnv(envName); ok {
		return strings.TrimSpace(v)
	}
	if strings.TrimSpace(profileVal) != "" {
		return strings.TrimSpace(profileVal)
	}
	return fallback
}

func chooseBool(cmd *cobra.Command, flagName, envName string, profileVal, fallback bool) (bool, error) {
	if cmd.Flags().Changed(flagName) {
		v, err := cmd.Flags().GetBool(flagName)
		if err != nil {
			return false, fmt.Errorf("read --%s: %w", flagName, err)
		}
		return v, nil
	}
	if raw, ok := os.LookupEnv(envName); ok {
		v, err := strconv.ParseBool(strings.TrimSpace(raw))
		if err != nil {
			return false, fmt.Errorf("invalid %s=%q: %w", envName, raw, err)
		}
		return v, nil
	}
	if profileVal {
		return true, nil
	}
	return fallback, nil
}

func chooseDuration(cmd *cobra.Command, flagName, envName string, fallback time.Duration) (time.Duration, error) {
	if cmd.Flags().Changed(flagName) {
		raw, err := cmd.Flags().GetString(flagName)
		if err != nil {
			return 0, fmt.Errorf("read --%s: %w", flagName, err)
		}
		v, err := time.ParseDuration(strings.TrimSpace(raw))
		if err != nil {
			return 0, fmt.Errorf("invalid --%s %q: %w", flagName, raw, err)
		}
		return v, nil
	}
	if raw, ok := os.LookupEnv(envName); ok {
		v, err := time.ParseDuration(strings.TrimSpace(raw))
		if err != nil {
			return 0, fmt.Errorf("invalid %s=%q: %w", envName, raw, err)
		}
		return v, nil
	}
	return fallback, nil
}
