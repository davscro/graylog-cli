package cli

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"

	"github.com/dsantic/graylog-cli/internal/output"
)

func (a *App) newNodesCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "nodes", Short: "Node commands"}
	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List cluster nodes",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := a.mustAuth(); err != nil {
				return err
			}
			c, err := a.client()
			if err != nil {
				return err
			}
			var data map[string]map[string]any
			if err := c.Do(cmd.Context(), http.MethodGet, "/cluster/nodes", nil, &data); err != nil {
				return err
			}

			if a.runtime.Format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), data)
			}

			tw := table.NewWriter()
			tw.AppendHeader(table.Row{"NODE_ID", "TRANSPORT_ADDRESS", "LAST_SEEN"})

			ids := make([]string, 0, len(data))
			for id := range data {
				ids = append(ids, id)
			}
			sort.Strings(ids)

			for _, id := range ids {
				node := data[id]
				tw.AppendRow(table.Row{id, fmt.Sprintf("%v", node["transport_address"]), fmt.Sprintf("%v", node["last_seen"])})
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), tw.Render())
			return err
		},
	})
	return cmd
}
