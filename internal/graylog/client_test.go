package graylog

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthHeaderSelectionTokenPreferred(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		token    string
		session  string
		expected string
	}{
		{name: "token", token: "tkn", expected: "tkn:token"},
		{name: "session", session: "sid", expected: "sid:session"},
		{name: "token over session", token: "tkn", session: "sid", expected: "tkn:token"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				auth := r.Header.Get("Authorization")
				want := "Basic " + base64.StdEncoding.EncodeToString([]byte(tc.expected))
				if auth != want {
					t.Fatalf("authorization mismatch: got %q want %q", auth, want)
				}
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"ok":true}`))
			}))
			defer srv.Close()

			c, err := NewClient(ClientConfig{BaseURL: srv.URL, APIBase: "/api", Token: tc.token, Session: tc.session})
			if err != nil {
				t.Fatalf("new client: %v", err)
			}
			if err := c.Do(context.Background(), http.MethodGet, "/system", nil, &map[string]any{}); err != nil {
				t.Fatalf("do request: %v", err)
			}
		})
	}
}

func TestHeaderInjection(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Accept") != "application/json" {
			t.Fatalf("missing Accept header")
		}
		if r.Method == http.MethodPost {
			if r.Header.Get("Content-Type") != "application/json" {
				t.Fatalf("missing Content-Type header")
			}
			if r.Header.Get("X-Requested-By") != "cli" {
				t.Fatalf("missing X-Requested-By header")
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c, err := NewClient(ClientConfig{BaseURL: srv.URL, APIBase: "/api"})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	if err := c.Do(context.Background(), http.MethodGet, "/system", nil, &map[string]any{}); err != nil {
		t.Fatalf("get request failed: %v", err)
	}
	if err := c.Do(context.Background(), http.MethodPost, "/system/sessions", map[string]string{"a": "b"}, &map[string]any{}); err != nil {
		t.Fatalf("post request failed: %v", err)
	}
}

func TestURLJoinBehavior(t *testing.T) {
	t.Parallel()

	c, err := NewClient(ClientConfig{BaseURL: "https://graylog.example.com/root/", APIBase: "/api"})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	got := c.URLFor("/cluster/nodes")
	want := "https://graylog.example.com/root/api/cluster/nodes"
	if got != want {
		t.Fatalf("url mismatch: got %q want %q", got, want)
	}
}
