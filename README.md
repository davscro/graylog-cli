# graylogctl

`graylogctl` is a production-focused CLI for Graylog v6.3.x (tested for 6.3.4), with first-class support for the Search Scripting API (`/api/search/messages`).

## Documentation

- Agent-focused operational manual: `docs/AGENT_OPERATIONS.md`

## Features

- Graylog REST API base path configurable (`--api-base`, default `/api`)
- Auth:
  - Access token via Basic Auth (`username=<token>`, `password=token`)
  - Session login via `POST /api/system/sessions` (`username=<session_id>`, `password=session`)
- Commands:
  - `auth login|whoami|logout`
  - `cluster info`
  - `system overview`
  - `nodes list`
  - `indices stats`
  - `search messages relative|absolute|keyword`
- Output formats: `table` or `json`
- Config profiles in `~/.config/graylogctl/config.yaml`
- Precedence: `flags > env > config > defaults`

## Installation

### Homebrew (recommended)

```bash
brew tap davscro/homebrew-tap
brew install davscro/homebrew-tap/graylogctl
graylogctl --help
```

### Build from source

```bash
make build
./bin/graylogctl --help
```

## Configuration

Path: `~/.config/graylogctl/config.yaml`

```yaml
profiles:
  default:
    url: https://graylog.example.com
    api_base: /api
    insecure: false
    auth:
      token: ""
      session: ""
```

Environment variables:

- `GRAYLOGCTL_URL`
- `GRAYLOGCTL_API_BASE`
- `GRAYLOGCTL_TOKEN`
- `GRAYLOGCTL_SESSION`
- `GRAYLOGCTL_INSECURE`
- `GRAYLOGCTL_TIMEOUT`
- `GRAYLOGCTL_FORMAT`
- `GRAYLOGCTL_PROFILE`

## Authentication Examples

### Token Auth (preferred)

```bash
export GRAYLOGCTL_URL='https://graylog.example.com'
export GRAYLOGCTL_TOKEN='your-graylog-token'
graylogctl system overview
```

### Session Login

```bash
graylogctl --url https://graylog.example.com auth login --user admin --password 'secret'
graylogctl auth whoami
```

### Logout

```bash
graylogctl auth logout
```

## Search Examples (Graylog 6.3 Search Scripting API)

### Relative

```bash
graylogctl search messages relative \
  --query 'source:nginx AND error' \
  --seconds 300 \
  --limit 50 \
  --fields 'timestamp,source,message'
```

### Absolute

```bash
graylogctl search messages absolute \
  --query 'source:nginx AND error' \
  --from '2026-02-18T10:00:00Z' \
  --to '2026-02-18T11:00:00Z' \
  --limit 50 \
  --fields 'timestamp,source,message'
```

### Keyword

```bash
graylogctl search messages keyword \
  --query 'source:nginx AND error' \
  --keyword 'last five minutes' \
  --limit 50 \
  --fields 'timestamp,source,message'
```

## Common Global Flags

- `--url`
- `--api-base`
- `--token`
- `--session`
- `--insecure`
- `--timeout` (Go duration format, default `30s`)
- `--format` (`table|json`)
- `--profile`
- `--max-width` (optional truncation for table cells)

## Testing

```bash
make test
```

In this environment, tests are run with `CGO_ENABLED=0` to avoid local dynamic loader issues.
