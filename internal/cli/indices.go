package cli

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"

	"github.com/dsantic/graylog-cli/internal/output"
)

func (a *App) newIndicesCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "indices", Short: "Indices commands"}
	cmd.AddCommand(&cobra.Command{
		Use:   "stats",
		Short: "Show index set stats",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := a.mustAuth(); err != nil {
				return err
			}
			c, err := a.client()
			if err != nil {
				return err
			}
			var data map[string]map[string]any
			if err := c.Do(cmd.Context(), http.MethodGet, "/system/indices/index_sets/stats", nil, &data); err != nil {
				return err
			}

			if a.runtime.Format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), data)
			}

			tw := table.NewWriter()
			tw.AppendHeader(table.Row{"INDEX_SET", "DOCS", "STORE_SIZE_BYTES"})

			keys := make([]string, 0, len(data))
			for k := range data {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			for _, k := range keys {
				row := data[k]
				tw.AppendRow(table.Row{k, fmt.Sprintf("%v", row["docs"]), fmt.Sprintf("%v", row["store_size_bytes"])})
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), tw.Render())
			return err
		},
	})
	return cmd
}
