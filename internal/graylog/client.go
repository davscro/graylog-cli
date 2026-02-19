package graylog

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

type ClientConfig struct {
	BaseURL  string
	APIBase  string
	Token    string
	Session  string
	Insecure bool
	Timeout  time.Duration
}

type Client struct {
	baseURL string
	apiBase string
	token   string
	session string
	http    *http.Client
}

type APIError struct {
	StatusCode int
	Endpoint   string
	Message    string
	Body       string
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("graylog API %d at %s: %s", e.StatusCode, e.Endpoint, e.Message)
	}
	if e.Body != "" {
		return fmt.Sprintf("graylog API %d at %s: %s", e.StatusCode, e.Endpoint, e.Body)
	}
	return fmt.Sprintf("graylog API %d at %s", e.StatusCode, e.Endpoint)
}

func NewClient(cfg ClientConfig) (*Client, error) {
	if strings.TrimSpace(cfg.BaseURL) == "" {
		return nil, errors.New("graylog URL is required (set --url, GRAYLOGCTL_URL, or config profile)")
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 30 * time.Second
	}
	if cfg.APIBase == "" {
		cfg.APIBase = "/api"
	}
	if !strings.HasPrefix(cfg.APIBase, "/") {
		cfg.APIBase = "/" + cfg.APIBase
	}

	if _, err := url.ParseRequestURI(cfg.BaseURL); err != nil {
		return nil, fmt.Errorf("invalid Graylog URL %q: %w", cfg.BaseURL, err)
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: cfg.Insecure} //nolint:gosec

	return &Client{
		baseURL: strings.TrimRight(cfg.BaseURL, "/"),
		apiBase: cfg.APIBase,
		token:   cfg.Token,
		session: cfg.Session,
		http: &http.Client{
			Timeout:   cfg.Timeout,
			Transport: transport,
		},
	}, nil
}

func (c *Client) URLFor(apiPath string) string {
	base, _ := url.Parse(c.baseURL)
	base.Path = path.Join(base.Path, c.apiBase, strings.TrimLeft(apiPath, "/"))
	return base.String()
}

func (c *Client) Do(ctx context.Context, method, apiPath string, reqBody any, out any) error {
	endpoint := c.URLFor(apiPath)
	var body io.Reader
	if reqBody != nil {
		b, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("marshal request to %s: %w", endpoint, err)
		}
		body = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return fmt.Errorf("create request %s %s: %w", method, endpoint, err)
	}

	req.Header.Set("Accept", "application/json")
	if method != http.MethodGet {
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Requested-By", "cli")
	}

	if c.token != "" {
		req.SetBasicAuth(c.token, "token")
	} else if c.session != "" {
		req.SetBasicAuth(c.session, "session")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("request %s %s: %w", method, endpoint, err)
	}
	defer resp.Body.Close()

	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response %s: %w", endpoint, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		apiErr := parseAPIError(resp.StatusCode, endpoint, payload)
		if strings.HasSuffix(apiPath, "/search/messages") && resp.StatusCode == http.StatusNotFound {
			apiErr.Message = "Your Graylog may not expose Search Scripting API. Check version >= 6.x and permissions. Consider using views search or legacy endpoints."
		}
		if strings.HasSuffix(apiPath, "/search/messages") && resp.StatusCode == http.StatusForbidden {
			apiErr.Message = apiErr.Message + " Guidance: token/session user must have permission to run searches."
		}
		return apiErr
	}

	if out == nil || len(payload) == 0 {
		return nil
	}

	if err := json.Unmarshal(payload, out); err != nil {
		return fmt.Errorf("decode response from %s: %w", endpoint, err)
	}
	return nil
}

func (c *Client) CreateSession(ctx context.Context, username, password string) (SessionResponse, error) {
	body := SessionRequest{Username: username, Password: password, Host: ""}
	var resp SessionResponse
	if err := c.Do(ctx, http.MethodPost, "/system/sessions", body, &resp); err != nil {
		return SessionResponse{}, err
	}
	if resp.ID == "" {
		return SessionResponse{}, errors.New("session response missing session_id")
	}
	return resp, nil
}

func (c *Client) SearchMessages(ctx context.Context, req SearchMessagesRequest) (SearchMessagesResponse, error) {
	var raw map[string]any
	if err := c.Do(ctx, http.MethodPost, "/search/messages", req, &raw); err != nil {
		return SearchMessagesResponse{}, err
	}
	b, _ := json.Marshal(raw)
	var parsed SearchMessagesResponse
	if err := json.Unmarshal(b, &parsed); err != nil {
		return SearchMessagesResponse{}, fmt.Errorf("parse search response: %w", err)
	}
	parsed.Raw = raw
	return parsed, nil
}

func parseAPIError(status int, endpoint string, body []byte) *APIError {
	errResp := struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	}{}
	if json.Unmarshal(body, &errResp) == nil && errResp.Message != "" {
		return &APIError{StatusCode: status, Endpoint: endpoint, Message: errResp.Message}
	}
	snippet := strings.TrimSpace(string(body))
	if len(snippet) > 300 {
		snippet = snippet[:300] + "..."
	}
	return &APIError{StatusCode: status, Endpoint: endpoint, Body: snippet}
}
