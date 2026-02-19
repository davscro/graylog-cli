# graylogctl Agent Operations Guide

This guide is for autonomous agents and scripted workflows that must use `graylogctl` without human interaction.

## Scope and Compatibility

- CLI: `graylogctl`
- Graylog target: v6.3.x (validated for 6.3.4)
- API base default: `/api`
- Primary search endpoint: `POST /api/search/messages`
- Legacy universal search endpoints are not used by this CLI.

## Non-Interactive Rules

- Always pass required values via flags or environment variables.
- Never rely on prompts (the CLI does not prompt).
- Prefer `--format json` for machine consumption.
- Treat any non-zero exit code as failure.
- Avoid embedding secrets directly in command history; use environment variables.

## Installation and Build

```bash
cd /path/to/graylog-cli
make build
```

Binary path:

```bash
./bin/graylogctl
```

## Configuration Model

Config file:

```text
~/.config/graylogctl/config.yaml
```

Format:

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

Precedence:

```text
flags > environment variables > config file > built-in defaults
```

Defaults:

- `api_base`: `/api`
- `timeout`: `30s`
- `format`: `table`
- `profile`: `default`

## Supported Environment Variables

- `GRAYLOGCTL_URL`
- `GRAYLOGCTL_API_BASE`
- `GRAYLOGCTL_TOKEN`
- `GRAYLOGCTL_SESSION`
- `GRAYLOGCTL_INSECURE`
- `GRAYLOGCTL_TIMEOUT`
- `GRAYLOGCTL_FORMAT`
- `GRAYLOGCTL_PROFILE`

## Authentication (Automated)

### Preferred: Access Token

Graylog token auth is Basic Auth with:

- username: `<token>`
- password: `token`

Set once per process:

```bash
export GRAYLOGCTL_URL='https://graylog.example.com'
export GRAYLOGCTL_TOKEN='your-graylog-token'
export GRAYLOGCTL_FORMAT='json'
```

Validation command:

```bash
./bin/graylogctl auth whoami
```

### Alternative: Session Token Login

Create and store session token in selected profile:

```bash
./bin/graylogctl \
  --url 'https://graylog.example.com' \
  --profile default \
  auth login --user 'admin' --password 'secret'
```

- Internally calls `POST /api/system/sessions` with `X-Requested-By: cli`.
- Stores returned `session_id` as profile session auth.

Logout (clear saved profile auth):

```bash
./bin/graylogctl --profile default auth logout
```

## Global Flags

- `--url`
- `--api-base`
- `--token`
- `--session`
- `--insecure`
- `--timeout` (Go duration, e.g. `30s`, `2m`)
- `--format` (`table` or `json`)
- `--profile`
- `--max-width` (table cell truncation; `0` means no truncation)

## Core Command Set

### Cluster/System

```bash
# GET /api/cluster
./bin/graylogctl cluster info

# GET /api/system
./bin/graylogctl system overview

# GET /api/cluster/nodes
./bin/graylogctl nodes list

# GET /api/system/indices/index_sets/stats
./bin/graylogctl indices stats
```

### Search Messages (Primary API)

Search endpoint used:

- `POST /api/search/messages`

Timerange modes:

- `relative` with `--seconds`
- `absolute` with `--from` and `--to` (ISO8601)
- `keyword` with `--keyword` (example: `last five minutes`)

#### Relative

```bash
./bin/graylogctl \
  --format json \
  search messages relative \
  --query 'level:3' \
  --seconds 300 \
  --limit 50 \
  --fields 'timestamp,source,message,level,error,service,operationName,attempt,maxRetries,streams,_id,gl2_message_id,gl2_remote_ip,gl2_remote_port,gl2_receive_timestamp,gl2_processing_timestamp,gl2_processing_duration_ms,gl2_source_input,gl2_source_node'
```

#### Absolute

```bash
./bin/graylogctl \
  --format json \
  search messages absolute \
  --query 'source:cff7b235f30c AND level:3' \
  --from '2026-02-19T22:00:00Z' \
  --to '2026-02-19T22:10:00Z' \
  --limit 100 \
  --fields 'timestamp,source,message,level,error,_id'
```

#### Keyword

```bash
./bin/graylogctl \
  --format json \
  search messages keyword \
  --query 'service:sync-service AND level:3' \
  --keyword 'last five minutes' \
  --limit 50 \
  --fields 'timestamp,source,message,level,error'
```

### Streams and Sorting in Search

```bash
./bin/graylogctl \
  --format json \
  search messages relative \
  --query 'level:3' \
  --seconds 900 \
  --stream '6900fa30becaa4ac09796c05' \
  --sort 'timestamp' \
  --sort-order 'desc' \
  --limit 200 \
  --fields 'timestamp,source,message,level,error,service'
```

## Output Contract

### `--format table`

- Human-readable table.
- Default search fields if omitted: `timestamp,source,message`.
- This is why only three columns appear unless `--fields` is explicitly set.

### `--format json`

For search commands, output is normalized JSON:

```json
{
  "schema": [...],
  "rows": [{"fieldA":"...","fieldB":"..."}],
  "metadata": {...}
}
```

For non-search commands, output is decoded endpoint JSON.

## Error Handling and Automation Behavior

### HTTP and API Errors

- Non-2xx responses return command failure.
- Error text includes status code and endpoint.
- If error payload is JSON with `message`, CLI surfaces that message.
- If response is not JSON, CLI includes a body snippet.

### Search-Specific Guidance

- `404` from `/search/messages` yields explicit guidance:
  - Graylog may not expose Search Scripting API.
  - Check Graylog version and permissions.
- `403` from `/search/messages` appends guidance about missing search permissions.

### Header Behavior

- `Accept: application/json` sent on all requests.
- For non-GET requests also sends:
  - `Content-Type: application/json`
  - `X-Requested-By: cli`

## Recommended Agent Playbooks

### Playbook A: Stateless Token-Based Collection

1. Export runtime env vars.
2. Run read-only commands with `--format json`.
3. Parse JSON and publish findings.
4. Unset secrets after completion.

Example:

```bash
export GRAYLOGCTL_URL='https://graylog.example.com'
export GRAYLOGCTL_TOKEN='***'
export GRAYLOGCTL_FORMAT='json'

./bin/graylogctl system overview
./bin/graylogctl nodes list
./bin/graylogctl indices stats
./bin/graylogctl search messages relative --query 'level:3' --seconds 300 --limit 100 --fields 'timestamp,source,message,level,error,service,_id'

unset GRAYLOGCTL_TOKEN
```

### Playbook B: Profile-Based Session Workflow

1. Perform `auth login` once.
2. Use `--profile` consistently.
3. Rotate/logout when done.

```bash
./bin/graylogctl --url 'https://graylog.example.com' --profile default auth login --user 'admin' --password 'secret'
./bin/graylogctl --profile default --format json auth whoami
./bin/graylogctl --profile default --format json search messages keyword --query 'service:sync-service' --keyword 'last five minutes' --fields 'timestamp,source,message,level,error'
./bin/graylogctl --profile default auth logout
```

## Troubleshooting

### Only `timestamp/source/message` visible

Cause: default search fields.

Fix: provide explicit `--fields` list.

### Command fails with auth error

- Verify token/session validity.
- Verify user permissions in Graylog.
- Run `auth whoami` with same profile/flags.

### Search returns 403

- Token/session user lacks search permissions.
- Use a principal allowed to execute searches.

### Search returns 404

- Endpoint not exposed or inaccessible.
- Verify Graylog version is 6.x and API availability.
- Confirm base URL and `--api-base` are correct.

### TLS certificate issues

- Prefer fixing CA trust.
- Temporary bypass: `--insecure` (risk accepted).

## Security Guidance for Agents

- Keep tokens in environment or secret manager, not hardcoded files.
- Avoid writing secrets to logs.
- Use least-privilege tokens.
- Clear sensitive env vars after use.

## Minimal Machine-First Command Reference

```bash
# verify auth identity
graylogctl --format json auth whoami

# system state
graylogctl --format json system overview
graylogctl --format json cluster info
graylogctl --format json nodes list
graylogctl --format json indices stats

# search relative
graylogctl --format json search messages relative --query 'level:3' --seconds 300 --limit 50 --fields 'timestamp,source,message,level,error,_id'

# search absolute
graylogctl --format json search messages absolute --query 'level:3' --from '2026-02-19T22:00:00Z' --to '2026-02-19T22:10:00Z' --limit 50 --fields 'timestamp,source,message,level,error,_id'

# search keyword
graylogctl --format json search messages keyword --query 'level:3' --keyword 'last five minutes' --limit 50 --fields 'timestamp,source,message,level,error,_id'
```
