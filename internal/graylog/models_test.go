package graylog

import "testing"

func TestNormalizeSearchResponse(t *testing.T) {
	t.Parallel()

	resp := SearchMessagesResponse{
		Schema: []SearchSchemaColumn{
			{Name: "timestamp", Field: "timestamp"},
			{Name: "", Field: "source"},
			{Name: "message", Field: "message"},
		},
		DataRows: [][]any{{"2026-02-18T10:00:00Z", "nginx-1", "error happened"}},
		Metadata: map[string]any{"effective_timerange": map[string]any{"type": "relative", "range": 300}},
	}

	n := NormalizeSearchResponse(resp)
	if len(n.Rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(n.Rows))
	}
	if n.Rows[0]["timestamp"] != "2026-02-18T10:00:00Z" {
		t.Fatalf("unexpected timestamp value: %v", n.Rows[0]["timestamp"])
	}
	if n.Rows[0]["source"] != "nginx-1" {
		t.Fatalf("unexpected source value: %v", n.Rows[0]["source"])
	}
}
