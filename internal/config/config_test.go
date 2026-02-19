package config

import (
	"testing"
	"time"

	"github.com/spf13/cobra"
)

func TestResolvePrecedenceFlagsEnvConfig(t *testing.T) {
	t.Setenv(EnvURL, "https://env.example.com")
	t.Setenv(EnvFormat, "json")
	t.Setenv(EnvTimeout, "45s")

	cfg := &Config{Profiles: map[string]Profile{
		"default": {
			URL:      "https://config.example.com",
			APIBase:  "/api",
			Insecure: false,
			Auth: ProfileAuth{
				Token: "cfg-token",
			},
		},
	}}

	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("url", "", "")
	cmd.Flags().String("api-base", DefaultAPIBase, "")
	cmd.Flags().String("token", "", "")
	cmd.Flags().String("session", "", "")
	cmd.Flags().Bool("insecure", false, "")
	cmd.Flags().String("timeout", DefaultTimeout.String(), "")
	cmd.Flags().String("format", DefaultFormat, "")
	cmd.Flags().String("profile", DefaultProfile, "")
	cmd.Flags().Int("max-width", 0, "")

	if err := cmd.Flags().Set("url", "https://flag.example.com"); err != nil {
		t.Fatalf("set flag: %v", err)
	}
	if err := cmd.Flags().Set("token", "flag-token"); err != nil {
		t.Fatalf("set flag: %v", err)
	}

	r, err := Resolve(cmd, cfg)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}

	if r.URL != "https://flag.example.com" {
		t.Fatalf("expected URL from flag, got %q", r.URL)
	}
	if r.Token != "flag-token" {
		t.Fatalf("expected token from flag, got %q", r.Token)
	}
	if r.Format != "json" {
		t.Fatalf("expected format from env, got %q", r.Format)
	}
	if r.Timeout != 45*time.Second {
		t.Fatalf("expected timeout from env, got %s", r.Timeout)
	}
}
