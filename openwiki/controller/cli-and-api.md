---
type: "Reference"
title: "Controller: CLI & API"
description: "agt CLI commands and REST API endpoints for managing AgentTeams resources, worker lifecycle, gateway, credentials, and Matrix AppService."
tags: ["cli", "api", "rest", "controller", "endpoints", "http"]
verified:
  - by: openwiki/0.7.1
    at: 2026-10-08T08:19:13.018Z
sources:
  - id: openwiki-source-3ecfb07f5d28ee64bb19f746
    resource: repo://agentteams-controller/cmd/agt/client.go
  - id: openwiki-source-ce369c2f6a027106277214bf
    resource: repo://agentteams-controller/cmd/agt/create.go
  - id: openwiki-source-7f0cce94c10a98f9058ca05c
    resource: repo://agentteams-controller/cmd/agt/llm_preflight.go
  - id: openwiki-source-a190bc071e7857ac07de6bbe
    resource: repo://agentteams-controller/cmd/agt/main.go
  - id: openwiki-source-d02d737ed578e628d1bd0785
    resource: repo://agentteams-controller/cmd/agt/status_cmd.go
  - id: openwiki-source-3eeca766f3b3880bf7ed6cb5
    resource: repo://agentteams-controller/cmd/agt/worker_cmd.go
  - id: openwiki-source-a8b495510006b0d8e0e506c4
    resource: repo://agentteams-controller/internal/server/gateway_handler.go
  - id: openwiki-source-75f1314f7931e9e5735fe34b
    resource: repo://agentteams-controller/internal/server/http_metrics.go
  - id: openwiki-source-ed1f536c8229fc0a989186d8
    resource: repo://agentteams-controller/internal/server/http.go
  - id: openwiki-source-6d2c8cb46356fce3f82156eb
    resource: repo://agentteams-controller/internal/server/lifecycle_handler.go
  - id: openwiki-source-ce950cd2a4b4046e7cac7189
    resource: repo://agentteams-controller/internal/server/managertasks_handler.go
  - id: openwiki-source-4c5210c162efebfa20b8fd33
    resource: repo://agentteams-controller/internal/server/resource_handler.go
  - id: openwiki-source-fd29cad401408cfdf164f738
    resource: repo://agentteams-controller/internal/server/worker_events_test.go
generated: { by: "openwiki/0.7.1", at: "2026-10-08T08:19:13.018Z" }
---

# Controller: CLI & API

The `agt` CLI and REST API are the primary interfaces for managing AgentTeams resources. The CLI is built into Manager and Worker images and communicates with the controller REST API. The API runs on the controller at port **8090** (configurable via `AGENTTEAMS_CONTROLLER_URL`).

## Authentication

Both the CLI and API use bearer-token authentication:

| Variable | Purpose |
|----------|---------|
| `AGENTTEAMS_CONTROLLER_URL` | Controller base URL (default `http://localhost:8090`) |
| `AGENTTEAMS_AUTH_TOKEN` | Bearer token (inline) |
| `AGENTTEAMS_AUTH_TOKEN_FILE` | Path to a file containing the bearer token (K8s projected volume) |

When all three are empty the client operates unauthenticated (for controllers with auth disabled). The CLI token discovery logic is in [`agentteams-controller/cmd/agt/client.go`](../../agentteams-controller/cmd/agt/client.go).

## CLI Commands

The CLI is defined in [`agentteams-controller/cmd/agt/`](../../agentteams-controller/cmd/agt/) and built with Cobra.

### Resource CRUD

```bash
# Create resources
agt create worker --name my-worker --runtime openclaw --model qwen3.6-plus
agt create worker --name alice --soul-file /path/to/SOUL.md --skills github-operations
agt create team --name my-team --workers worker-a,worker-b
agt create human --name alice --email alice@example.com
agt create manager --name my-manager --runtime copaw

# List / get resources
agt get workers
agt get workers alice
agt get workers --team alpha-team
agt get teams
agt get humans
agt get managers
agt get workers -o json

# Update resources
agt update worker --name my-worker --model claude-sonnet-4-6
agt update team --name alpha --description "Updated description"
agt update team --name alpha --leader-model claude-sonnet-4-6

# Delete resources
agt delete worker my-worker
agt delete team my-team
agt delete human alice
agt delete manager my-manager

# Declarative apply (create or update from YAML)
agt apply -f resource.yaml

# Apply worker from CLI params (create-or-update)
agt apply worker --name alice --model qwen3.6-plus
agt apply worker --name alice --zip worker.zip
```

### Worker Lifecycle

```bash
# Wake a sleeping worker
agt worker wake --name my-worker

# Sleep a running worker
agt worker sleep --name my-worker

# Ensure worker is ready (start if sleeping, wait for phase)
agt worker ensure-ready --name my-worker

# Report worker ready (called by worker itself, with optional heartbeat)
agt worker report-ready --name my-worker
agt worker report-ready --heartbeat --interval 30s

# Get worker runtime status (single or all in a team)
agt worker status --name my-worker
agt worker status --team alpha-team
```

The `report-ready` subcommand reads the worker name from `--name`, `AGENTTEAMS_WORKER_CR_NAME`, or `AGENTTEAMS_WORKER_NAME` env var when `--name` is omitted. It retries up to 5 times on failure and supports periodic heartbeat mode.

### Status & Monitoring

```bash
# Show cluster overview (all resource types in one screen)
agt status

# Watch mode (auto-refresh every 3 seconds)
agt status --watch

# JSON output for scripting
agt status -o json

# Show controller version
agt version
```

### Credential Management

```bash
# Rotate Matrix AppService as_token (and optional hs_token)
agt rotate appservice-token --as-token NEW_TOKEN

# Test LLM provider connectivity
agt llm-preflight
agt llm-preflight --provider openai-compat --model gpt-4 --strict
```

### Manager State

```bash
# Initialize task board (state.json)
agt manager-state --action init

# Add finite task
agt manager-state --action add-finite --task-id T --title "Deploy service X" --assigned-to W --room-id R

# List tasks
agt manager-state --action list
```

The `manager-state` command operates on the local `state.json` file (path configured via `AGENTTEAMS_MANAGER_STATE_FILE`). It mirrors the shell-based `manage-state.sh` workflow so Manager skills can call either interface.

## REST API

The REST API is implemented in [`agentteams-controller/internal/server/`](../../agentteams-controller/internal/server/) and registered in [`http.go`](../../agentteams-controller/internal/server/http.go).

### Resource Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/workers` | Create worker |
| `GET` | `/api/v1/workers` | List workers (includes synthesized team members; filter `?team=`) |
| `GET` | `/api/v1/workers/{name}` | Get worker (falls back to team member synthesis) |
| `GET` | `/api/v1/workers/{name}/events` | Get bounded worker event history (newest first) |
| `PUT` | `/api/v1/workers/{name}` | Update worker |
| `DELETE` | `/api/v1/workers/{name}` | Delete worker |
| `POST` | `/api/v1/teams` | Create team |
| `GET` | `/api/v1/teams` | List teams |
| `GET` | `/api/v1/teams/{name}` | Get team |
| `PUT` | `/api/v1/teams/{name}` | Update team |
| `DELETE` | `/api/v1/teams/{name}` | Delete team |
| `POST` | `/api/v1/humans` | Create human |
| `GET` | `/api/v1/humans` | List humans |
| `GET` | `/api/v1/humans/{name}` | Get human |
| `DELETE` | `/api/v1/humans/{name}` | Delete human |
| `POST` | `/api/v1/managers` | Create manager |
| `GET` | `/api/v1/managers` | List managers |
| `GET` | `/api/v1/managers/{name}` | Get manager |
| `PUT` | `/api/v1/managers/{name}` | Update manager |
| `DELETE` | `/api/v1/managers/{name}` | Delete manager |
| `POST` | `/api/v1/projects` | Create project |
| `GET` | `/api/v1/projects` | List projects |
| `GET` | `/api/v1/projects/{name}` | Get project |
| `PUT` | `/api/v1/projects/{name}` | Update project |
| `DELETE` | `/api/v1/projects/{name}` | Delete project |

### Lifecycle Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/workers/{name}/wake` | Wake sleeping worker (sets spec.state=Running, starts backend) |
| `POST` | `/api/v1/workers/{name}/sleep` | Sleep running worker (sets spec.state=Sleeping, stops backend) |
| `POST` | `/api/v1/workers/{name}/ensure-ready` | Start if sleeping/stopped, return current phase |
| `POST` | `/api/v1/workers/{name}/ready` | Worker self-reports readiness (called by worker with optional heartbeat) |
| `GET` | `/api/v1/workers/{name}/status` | Get aggregated runtime status (CR + backend state + health checks) |

Lifecycle handlers set `spec.State` declaratively (triggering the reconciler) and also directly operate the backend for immediate response. The reconciler retries if the direct backend operation fails.

### Message Injection

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/managers/{name}/message` | Send system-level message to Manager's Admin DM room |
| `POST` | `/api/v1/teams/{name}/message` | Send system-level message to team leader (prefers leader↔admin DM) |

### Manager Tasks

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/manager-tasks` | Raw Manager state.json passthrough (embedded mode only; 404 in-cluster) |

### Gateway Management

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/gateway/consumers` | Create gateway consumer |
| `POST` | `/api/v1/gateway/consumers/{id}/bind` | Bind consumer to AI routes |
| `DELETE` | `/api/v1/gateway/consumers/{id}` | Delete gateway consumer |
| `GET` | `/api/v1/gateway/providers` | List AI providers |
| `POST` | `/api/v1/gateway/providers` | Register AI provider (creates service source + provider + route) |
| `DELETE` | `/api/v1/gateway/providers/{name}` | Delete AI provider |
| `GET` | `/api/v1/gateway/providers/{name}/models` | List models for a provider |

### Credentials & Auth

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/credentials/sts` | Refresh STS token (self-scoped to caller identity) |
| `POST` | `/api/v1/credentials/matrix-token` | Refresh Matrix access token (called by workers/managers on 401) |

### AppService Management

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/appservice/rotate-token` | Rotate Matrix AppService as_token/hs_token on homeserver |

### Packages & Status

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/packages` | Upload ZIP package to OSS (multipart/form-data; max 64 MB) |
| `GET` | `/api/v1/status` | Cluster status overview (resource counts) |
| `GET` | `/api/v1/version` | Controller version and kube mode |
| `GET` | `/healthz` | Health check (no auth) |

### Docker API Passthrough (Embedded Mode)

When running in embedded mode (`kubeMode=embedded`), the controller proxies Docker API calls:

| Method | Path | Description |
|--------|------|-------------|
| `*` | `/docker/*` | Proxied Docker socket API (authenticated, requires gateway action) |

### Matrix AppService Endpoints

The controller acts as a Matrix AppService for event push when `AppServiceEnabled` is true and an `hs_token` is configured:

| Method | Path | Description |
|--------|------|-------------|
| `PUT` | `/_matrix/app/v1/transactions/{txnId}` | Transaction push from homeserver (authenticated by hs_token) |
| `GET` | `/_matrix/app/v1/users/{userId}` | User query |
| `GET` | `/_matrix/app/v1/rooms/{roomAlias}` | Room query |

## Handler Architecture

API handlers are in [`agentteams-controller/internal/server/`](../../agentteams-controller/internal/server/):

| File | Purpose |
|------|---------|
| `http.go` | HTTP server setup, dependency injection, route registration |
| `http_metrics.go` | Prometheus HTTP request metrics middleware |
| `resource_handler.go` | CRUD handlers for all resource types (workers, teams, humans, managers, projects) |
| `worker_resource_service.go` | Worker-specific resource validation and construction |
| `lifecycle_handler.go` | Worker lifecycle (wake/sleep/ensure-ready/ready/status) |
| `worker_health.go` | Five-point worker health probing (container, heartbeat, LLM, git, sync) with short-TTL cache |
| `message_handler.go` | Message injection into Manager DM rooms and team leader channels |
| `managertasks_handler.go` | Manager state.json passthrough (embedded mode) |
| `gateway_handler.go` | Gateway consumer and AI provider management |
| `credentials_handler.go` | STS token issuance and Matrix token refresh |
| `package_handler.go` | ZIP package upload to OSS |
| `status_handler.go` | Health check, cluster status, version endpoints |
| `appservice_handler.go` | Matrix AppService transaction push handler |
| `appservice_mgmt_handler.go` | AppService token rotation |
| `types.go` | Request/response types for all endpoints |

## CLI Source Structure

The CLI is built with Cobra and lives in [`agentteams-controller/cmd/agt/`](../../agentteams-controller/cmd/agt/):

| File | Purpose |
|------|---------|
| `main.go` | Entry point, root command, subcommand registration |
| `create.go` | `create` subcommands (worker, team, human, manager) |
| `get.go` | `get` subcommands with table/JSON output |
| `update.go` | `update` subcommands (worker, team, manager) |
| `delete.go` | `delete` subcommands |
| `apply.go` | `apply -f` declarative YAML + `apply worker` CLI-param variant |
| `worker_cmd.go` | `worker` lifecycle subcommands (wake/sleep/ensure-ready/status/report-ready) |
| `status_cmd.go` | `status` cluster overview + `version` command |
| `llm_preflight.go` | `llm-preflight` LLM connectivity validation |
| `rotate_cmd.go` | `rotate appservice-token` |
| `manager_state_cmd.go` | `manager-state` task board operations |
| `client.go` | HTTP client, token discovery, `DoJSON`/`DoMultipart` helpers |
| `output.go` | Table, detail, and JSON formatting utilities |
| `team_flags.go` | Shared team flag parsing helpers |

## Source References

- CLI commands: [`agentteams-controller/cmd/agt/`](../../agentteams-controller/cmd/agt/)
- REST API server: [`agentteams-controller/internal/server/http.go`](../../agentteams-controller/internal/server/http.go)
- Resource CRUD: [`agentteams-controller/internal/server/resource_handler.go`](../../agentteams-controller/internal/server/resource_handler.go)
- Lifecycle handler: [`agentteams-controller/internal/server/lifecycle_handler.go`](../../agentteams-controller/internal/server/lifecycle_handler.go)
- Health probing: [`agentteams-controller/internal/server/worker_health.go`](../../agentteams-controller/internal/server/worker_health.go)
- Gateway handler: [`agentteams-controller/internal/server/gateway_handler.go`](../../agentteams-controller/internal/server/gateway_handler.go)
- Message handler: [`agentteams-controller/internal/server/message_handler.go`](../../agentteams-controller/internal/server/message_handler.go)
- Credentials handler: [`agentteams-controller/internal/server/credentials_handler.go`](../../agentteams-controller/internal/server/credentials_handler.go)
