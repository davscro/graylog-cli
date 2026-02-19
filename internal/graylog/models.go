package graylog

import "fmt"

type SessionRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Host     string `json:"host"`
}

type SessionResponse struct {
	ID         string `json:"session_id"`
	Username   string `json:"username"`
	ValidUntil string `json:"valid_until"`
}

type SearchTimerange struct {
	Type    string `json:"type"`
	Range   int    `json:"range,omitempty"`
	From    string `json:"from,omitempty"`
	To      string `json:"to,omitempty"`
	Keyword string `json:"keyword,omitempty"`
}

type SearchMessagesRequest struct {
	Query     string          `json:"query"`
	Fields    []string        `json:"fields"`
	From      int             `json:"from"`
	Size      int             `json:"size"`
	Timerange SearchTimerange `json:"timerange"`
	Streams   []string        `json:"streams,omitempty"`
	Sort      string          `json:"sort,omitempty"`
	SortOrder string          `json:"sort_order,omitempty"`
}

type SearchSchemaColumn struct {
	Name  string `json:"name"`
	Field string `json:"field"`
	Type  string `json:"type"`
}

type SearchMessagesResponse struct {
	Schema   []SearchSchemaColumn `json:"schema"`
	DataRows [][]any              `json:"datarows"`
	Metadata map[string]any       `json:"metadata"`
	Raw      map[string]any       `json:"-"`
}

type SearchNormalized struct {
	Schema   []SearchSchemaColumn `json:"schema"`
	Rows     []map[string]any     `json:"rows"`
	Metadata map[string]any       `json:"metadata"`
}

func NormalizeSearchResponse(resp SearchMessagesResponse) SearchNormalized {
	headers := make([]string, 0, len(resp.Schema))
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

	rows := make([]map[string]any, 0, len(resp.DataRows))
	for _, raw := range resp.DataRows {
		row := make(map[string]any, len(headers))
		for i, h := range headers {
			if i < len(raw) {
				row[h] = raw[i]
			} else {
				row[h] = nil
			}
		}
		rows = append(rows, row)
	}

	return SearchNormalized{Schema: resp.Schema, Rows: rows, Metadata: resp.Metadata}
}
