package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/dsantic/graylog-cli/internal/config"
	"github.com/dsantic/graylog-cli/internal/graylog"
)

type App struct {
	v       *viper.Viper
	cfg     *config.Config
	runtime config.Runtime
}

func NewRootCmd() *cobra.Command {
	app := &App{v: viper.New()}

	cmd := &cobra.Command{
		Use:           "graylogctl",
		Short:         "Graylog v6 CLI",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := config.LoadConfig()
			if err != nil {
				return err
			}
			r, err := config.Resolve(cmd, cfg)
			if err != nil {
				return err
			}
			app.cfg = cfg
			app.runtime = r
			return nil
		},
	}

	cmd.PersistentFlags().String("url", "", "Graylog base URL")
	cmd.PersistentFlags().String("api-base", config.DefaultAPIBase, "Graylog API base path")
	cmd.PersistentFlags().String("token", "", "Graylog access token")
	cmd.PersistentFlags().String("session", "", "Graylog session id")
	cmd.PersistentFlags().Bool("insecure", false, "Skip TLS certificate verification")
	cmd.PersistentFlags().String("timeout", config.DefaultTimeout.String(), "HTTP timeout (Go duration, e.g. 30s)")
	cmd.PersistentFlags().String("format", config.DefaultFormat, "Output format: table|json")
	cmd.PersistentFlags().String("profile", config.DefaultProfile, "Config profile name")
	cmd.PersistentFlags().Int("max-width", 0, "Maximum table cell width (0 disables truncation)")

	app.bindEnv("url", config.EnvURL)
	app.bindEnv("api-base", config.EnvAPIBase)
	app.bindEnv("token", config.EnvToken)
	app.bindEnv("session", config.EnvSession)
	app.bindEnv("insecure", config.EnvInsecure)
	app.bindEnv("timeout", config.EnvTimeout)
	app.bindEnv("format", config.EnvFormat)
	app.bindEnv("profile", config.EnvProfile)

	cmd.AddCommand(
		app.newAuthCmd(),
		app.newClusterCmd(),
		app.newSystemCmd(),
		app.newNodesCmd(),
		app.newIndicesCmd(),
		app.newSearchCmd(),
	)

	return cmd
}

func (a *App) bindEnv(key, env string) {
	a.v.SetEnvPrefix("GRAYLOGCTL")
	a.v.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	a.v.BindEnv(key, env)
}

func (a *App) client() (*graylog.Client, error) {
	return graylog.NewClient(graylog.ClientConfig{
		BaseURL:  a.runtime.URL,
		APIBase:  a.runtime.APIBase,
		Token:    a.runtime.Token,
		Session:  a.runtime.Session,
		Insecure: a.runtime.Insecure,
		Timeout:  a.runtime.Timeout,
	})
}

func (a *App) mustAuth() error {
	if a.runtime.Token == "" && a.runtime.Session == "" {
		return fmt.Errorf("no auth configured; set --token or --session, or run graylogctl auth login")
	}
	return nil
}

func Execute() {
	if err := NewRootCmd().ExecuteContext(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
