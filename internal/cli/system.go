package cli

import (
	"net/http"

	"github.com/spf13/cobra"

	"github.com/dsantic/graylog-cli/internal/output"
)

func (a *App) newSystemCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "system", Short: "System commands"}
	cmd.AddCommand(&cobra.Command{
		Use:   "overview",
		Short: "Show system overview",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := a.mustAuth(); err != nil {
				return err
			}
			c, err := a.client()
			if err != nil {
				return err
			}
			var data map[string]any
			if err := c.Do(cmd.Context(), http.MethodGet, "/system", nil, &data); err != nil {
				return err
			}
			if a.runtime.Format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), data)
			}
			return output.PrintKeyValueTable(cmd.OutOrStdout(), data)
		},
	})
	return cmd
}
