package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/dsantic/graylog-cli/internal/graylog"
	"github.com/dsantic/graylog-cli/internal/output"
)

type searchCommon struct {
	Query     string
	Fields    string
	Offset    int
	Limit     int
	Streams   []string
	Sort      string
	SortOrder string
}

func (a *App) newSearchCmd() *cobra.Command {
	searchCmd := &cobra.Command{Use: "search", Short: "Search commands"}
	messagesCmd := &cobra.Command{Use: "messages", Short: "Search messages"}

	common := &searchCommon{}
	bindSearchCommonFlags(messagesCmd, common)

	messagesCmd.AddCommand(
		a.newSearchRelativeCmd(common),
		a.newSearchAbsoluteCmd(common),
		a.newSearchKeywordCmd(common),
	)
	searchCmd.AddCommand(messagesCmd)
	return searchCmd
}

func bindSearchCommonFlags(cmd *cobra.Command, common *searchCommon) {
	cmd.PersistentFlags().StringVar(&common.Query, "query", "", "Graylog query")
	cmd.PersistentFlags().StringVar(&common.Fields, "fields", "timestamp,source,message", "Comma-separated fields")
	cmd.PersistentFlags().IntVar(&common.Offset, "from", 0, "Result offset")
	cmd.PersistentFlags().IntVar(&common.Limit, "limit", 50, "Result size")
	cmd.PersistentFlags().StringSliceVar(&common.Streams, "stream", nil, "Restrict search to stream id (repeatable)")
	cmd.PersistentFlags().StringVar(&common.Sort, "sort", "", "Sort field")
	cmd.PersistentFlags().StringVar(&common.SortOrder, "sort-order", "desc", "Sort order asc|desc")
	_ = cmd.MarkPersistentFlagRequired("query")
}

func (a *App) newSearchRelativeCmd(common *searchCommon) *cobra.Command {
	var seconds int
	cmd := &cobra.Command{
		Use:   "relative",
		Short: "Relative time-range message search",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if seconds <= 0 {
				return fmt.Errorf("--seconds must be > 0")
			}
			req := buildSearchRequest(common)
			req.Timerange = graylog.SearchTimerange{Type: "relative", Range: seconds}
			return a.runSearch(cmd, req)
		},
	}
	cmd.Flags().IntVar(&seconds, "seconds", 300, "Relative timerange in seconds")
	return cmd
}

func (a *App) newSearchAbsoluteCmd(common *searchCommon) *cobra.Command {
	var from, to string
	cmd := &cobra.Command{
		Use:   "absolute",
		Short: "Absolute time-range message search",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if strings.TrimSpace(from) == "" || strings.TrimSpace(to) == "" {
				return fmt.Errorf("--from and --to are required")
			}
			req := buildSearchRequest(common)
			req.Timerange = graylog.SearchTimerange{Type: "absolute", From: strings.TrimSpace(from), To: strings.TrimSpace(to)}
			return a.runSearch(cmd, req)
		},
	}
	cmd.Flags().StringVar(&from, "from", "", "Start timestamp (ISO8601)")
	cmd.Flags().StringVar(&to, "to", "", "End timestamp (ISO8601)")
	return cmd
}

func (a *App) newSearchKeywordCmd(common *searchCommon) *cobra.Command {
	var keyword string
	cmd := &cobra.Command{
		Use:   "keyword",
		Short: "Keyword time-range message search",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if strings.TrimSpace(keyword) == "" {
				return fmt.Errorf("--keyword is required")
			}
			req := buildSearchRequest(common)
			req.Timerange = graylog.SearchTimerange{Type: "keyword", Keyword: strings.TrimSpace(keyword)}
			return a.runSearch(cmd, req)
		},
	}
	cmd.Flags().StringVar(&keyword, "keyword", "", "Keyword timerange (e.g. 'last five minutes')")
	return cmd
}

func (a *App) runSearch(cmd *cobra.Command, req graylog.SearchMessagesRequest) error {
	if err := a.mustAuth(); err != nil {
		return err
	}
	c, err := a.client()
	if err != nil {
		return err
	}
	resp, err := c.SearchMessages(cmd.Context(), req)
	if err != nil {
		return err
	}

	normalized := graylog.NormalizeSearchResponse(resp)
	if a.runtime.Format == "json" {
		return output.PrintJSON(cmd.OutOrStdout(), normalized)
	}
	return output.PrintSearchTable(cmd.OutOrStdout(), resp, a.runtime.MaxWidth)
}

func buildSearchRequest(common *searchCommon) graylog.SearchMessagesRequest {
	req := graylog.SearchMessagesRequest{
		Query:     strings.TrimSpace(common.Query),
		Fields:    parseFields(common.Fields),
		From:      common.Offset,
		Size:      common.Limit,
		Sort:      strings.TrimSpace(common.Sort),
		SortOrder: strings.ToLower(strings.TrimSpace(common.SortOrder)),
	}
	if req.SortOrder != "asc" && req.SortOrder != "desc" {
		req.SortOrder = "desc"
	}
	if len(common.Streams) > 0 {
		req.Streams = common.Streams
	}
	if len(req.Fields) == 0 {
		req.Fields = []string{"timestamp", "source", "message"}
	}
	return req
}

func parseFields(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
