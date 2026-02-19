package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"

	"github.com/dsantic/graylog-cli/internal/graylog"
)

func PrintJSON(w io.Writer, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}
	_, err = fmt.Fprintln(w, string(b))
	return err
}

func PrintKeyValueTable(w io.Writer, m map[string]any) error {
	tw := table.NewWriter()
	tw.AppendHeader(table.Row{"KEY", "VALUE"})
	for k, v := range m {
		tw.AppendRow(table.Row{k, fmt.Sprintf("%v", v)})
	}
	_, err := fmt.Fprintln(w, tw.Render())
	return err
}

func PrintSearchTable(w io.Writer, resp graylog.SearchMessagesResponse, maxWidth int) error {
	tw := table.NewWriter()

	headers := make(table.Row, 0, len(resp.Schema))
	for i, col := range resp.Schema {
		name := col.Name
		if name == "" {
			name = col.Field
		}
		if name == "" {
			name = fmt.Sprintf("col_%d", i)
		}
		headers = append(headers, name)
	}
	tw.AppendHeader(headers)

	for _, r := range resp.DataRows {
		row := make(table.Row, 0, len(headers))
		for i := range headers {
			var value any
			if i < len(r) {
				value = r[i]
			}
			cell := fmt.Sprintf("%v", value)
			if maxWidth > 0 {
				cell = truncate(cell, maxWidth)
			}
			row = append(row, cell)
		}
		tw.AppendRow(row)
	}

	if maxWidth > 0 {
		cfg := make([]table.ColumnConfig, 0, len(headers))
		for i := range headers {
			cfg = append(cfg, table.ColumnConfig{Number: i + 1, WidthMax: maxWidth})
		}
		tw.SetColumnConfigs(cfg)
	}
	_, err := fmt.Fprintln(w, tw.Render())
	return err
}

func truncate(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return strings.TrimSpace(s[:max-3]) + "..."
}
