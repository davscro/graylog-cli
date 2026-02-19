package cli

import (
	"fmt"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/dsantic/graylog-cli/internal/config"
	"github.com/dsantic/graylog-cli/internal/graylog"
	"github.com/dsantic/graylog-cli/internal/output"
)

func (a *App) newAuthCmd() *cobra.Command {
	authCmd := &cobra.Command{Use: "auth", Short: "Authentication commands"}
	authCmd.AddCommand(a.newAuthLoginCmd(), a.newAuthWhoAmICmd(), a.newAuthLogoutCmd())
	return authCmd
}

func (a *App) newAuthLoginCmd() *cobra.Command {
	var user, pass string
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Create a Graylog session token",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if a.runtime.URL == "" {
				return fmt.Errorf("Graylog URL is required (set --url or config)")
			}
			c, err := a.loginClient()
			if err != nil {
				return err
			}
			resp, err := c.CreateSession(cmd.Context(), user, pass)
			if err != nil {
				return err
			}

			profile := a.cfg.Profiles[a.runtime.Profile]
			profile.Auth.Session = resp.ID
			a.cfg.Profiles[a.runtime.Profile] = profile
			if err := config.SaveConfig(a.cfg); err != nil {
				return err
			}

			if a.runtime.Format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), map[string]any{"session_id": resp.ID, "profile": a.runtime.Profile})
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "session saved for profile %q\n", a.runtime.Profile)
			return err
		},
	}
	cmd.Flags().StringVar(&user, "user", "", "Graylog username")
	cmd.Flags().StringVar(&pass, "password", "", "Graylog password")
	_ = cmd.MarkFlagRequired("user")
	_ = cmd.MarkFlagRequired("password")
	return cmd
}

func (a *App) newAuthWhoAmICmd() *cobra.Command {
	return &cobra.Command{
		Use:   "whoami",
		Short: "Show authenticated identity info",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := a.mustAuth(); err != nil {
				return err
			}
			c, err := a.client()
			if err != nil {
				return err
			}

			var me map[string]any
			err = c.Do(cmd.Context(), http.MethodGet, "/users/me", nil, &me)
			if err != nil {
				apiErr, ok := err.(*graylog.APIError)
				if !ok || (apiErr.StatusCode != http.StatusNotFound && apiErr.StatusCode != http.StatusForbidden) {
					return err
				}
				me = map[string]any{}
				if err := c.Do(cmd.Context(), http.MethodGet, "/system", nil, &me); err != nil {
					return err
				}
			}

			if a.runtime.Format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), me)
			}
			return output.PrintKeyValueTable(cmd.OutOrStdout(), me)
		},
	}
}

func (a *App) newAuthLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Clear saved auth token/session for selected profile",
		RunE: func(cmd *cobra.Command, _ []string) error {
			profile := a.cfg.Profiles[a.runtime.Profile]
			profile.Auth.Token = ""
			profile.Auth.Session = ""
			a.cfg.Profiles[a.runtime.Profile] = profile
			if err := config.SaveConfig(a.cfg); err != nil {
				return err
			}
			if a.runtime.Format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), map[string]any{"profile": a.runtime.Profile, "logged_out": true})
			}
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "cleared token and session for profile %q\n", a.runtime.Profile)
			return err
		},
	}
}
