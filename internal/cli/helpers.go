package cli

import "github.com/dsantic/graylog-cli/internal/graylog"

func (a *App) loginClient() (*graylog.Client, error) {
	return graylog.NewClient(graylog.ClientConfig{
		BaseURL:  a.runtime.URL,
		APIBase:  a.runtime.APIBase,
		Insecure: a.runtime.Insecure,
		Timeout:  a.runtime.Timeout,
	})
}
