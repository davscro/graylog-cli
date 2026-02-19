package cli

import (
	"net/http"

	"github.com/spf13/cobra"

	"github.com/dsantic/graylog-cli/internal/output"
)

func (a *App) newClusterCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "cluster", Short: "Cluster commands"}
	cmd.AddCommand(a.clusterInfoCmd())
	return cmd
}

func (a *App) clusterInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Show cluster info",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := a.mustAuth(); err != nil {
				return err
			}
			c, err := a.client()
			if err != nil {
				return err
			}
			var data map[string]any
			if err := c.Do(cmd.Context(), http.MethodGet, "/cluster", nil, &data); err != nil {
				return err
			}
			if a.runtime.Format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), data)
			}
			return output.PrintKeyValueTable(cmd.OutOrStdout(), data)
		},
	}
}
