---
type: Reference
title: "Controller: CLI & API"
description: "Reference for the agt CLI commands and REST API endpoints (port 8090) covering resource CRUD, lifecycle operations, status/monitoring, credential management, and message injection."
tags: [cli, api, rest, controller, agt, endpoints]
verified:
  - by: openwiki/0.7.0
    at: 2026-10-03T08:15:26.963Z
sources:
  - id: openwiki-source-8e2b1d60c720a31b97758ede
    resource: repo://agentteams-controller/cmd/agt/apply.go
  - id: openwiki-source-ce369c2f6a027106277214bf
    resource: repo://agentteams-controller/cmd/agt/create.go
  - id: openwiki-source-7c7a3f52fa6dc4a83cc198b8
    resource: repo://agentteams-controller/cmd/agt/delete.go
  - id: openwiki-source-c61b01bc9c1074afe25e4e51
    resource: repo://agentteams-controller/cmd/agt/get.go
  - id: openwiki-source-a190bc071e7857ac07de6bbe
    resource: repo://agentteams-controller/cmd/agt/main.go
  - id: openwiki-source-d02d737ed578e628d1bd0785
    resource: repo://agentteams-controller/cmd/agt/status_cmd.go
  - id: openwiki-source-b9a3c9819248e7b1fdd6fe49
    resource: repo://agentteams-controller/cmd/agt/update.go
  - id: openwiki-source-3eeca766f3b3880bf7ed6cb5
    resource: repo://agentteams-controller/cmd/agt/worker_cmd.go
  - id: openwiki-source-ed1f536c8229fc0a989186d8
    resource: repo://agentteams-controller/internal/server/http.go
  - id: openwiki-source-ce950cd2a4b4046e7cac7189
    resource: repo://agentteams-controller/internal/server/managertasks_handler.go
  - id: openwiki-source-8d668c55535579f7cb4e8920
    resource: repo://agentteams-controller/internal/server/message_handler.go
  - id: openwiki-source-311b5c5f3a83755b689fba6d
    resource: repo://agentteams-controller/internal/server/package_handler.go
generated: { by: "openwiki/0.7.0", at: "2026-10-03T08:15:26.963Z" }
---

# Controller: CLI & API

The `agt` CLI and REST API are the primary interfaces for managing AgentTeams resources. The CLI is baked into Manager and Worker images. The REST API runs on the controller at port 8090.

## CLI Commands

The CLI is defined in [`agentteams-controller/cmd/agt/`](../agentteams-controller/cmd/agt/). It communicates with the controller REST API. Authentication uses environment variables:

| Variable | Purpose |
|----------|---------|
| `AGENTTEAMS_CONTROLLER_URL` | Controller base URL (default: `http://localhost:8090`) |
| `AGENTTEAMS_AUTH_TOKEN` | Bearer token for authentication |
| `AGENTTEAMS_AUTH_TOKEN_FILE` | Path to a file containing the bearer token (K8s projected volume) |

### Resource CRUD

```bash
# Create resources
agt create worker --name my-worker --runtime openclaw --model gpt-4
agt create team --name my-team --workers worker-a,worker-b
agt create human --name alice --email alice@example.com
agt create manager --name my-manager --runtime copaw

# List resources
agt get workers
agt get teams
agt get humans
agt get managers

# Get a specific resource
agt get workers my-worker
agt get teams my-team -o json

# Update resources
agt update worker my-worker --model gpt-5
agt update team my-team --add-worker worker-c

# Delete resources
agt delete worker my-worker
agt delete team my-team

# Declarative apply (create or update from YAML)
agt apply -f worker.yaml

# Apply worker with ZIP package
agt apply worker --name alice --zip worker.zip

# Apply worker with model override
agt apply worker --name alice --model qwen3.6-plus
```

### Worker Lifecycle

```bash
# Wake a sleeping worker
agt worker wake my-worker

# Sleep a running worker
agt worker sleep my-worker

# Ensure worker is ready (wait for provisioning)
agt worker ensure-ready my-worker

# Report worker ready (called by worker itself)
agt worker report-ready my-worker

# Report worker ready with periodic heartbeat
agt worker report-ready --heartbeat --interval 60s

# Get worker runtime status
agt worker status my-worker
```

### Status & Monitoring

```bash
# Show cluster overview (all resources, one screen)
agt status

# Watch mode (auto-refresh every 3 seconds)
agt status --watch

# JSON output for scripting
agt status --output json

# Show controller version
agt version
```

#### Cluster Overview Output Format

The `agt status` command displays a one-screen cluster overview with phase breakdowns per resource type:

```
Mode:       embedded

Workers (3 total, 2 Ready, 1 Pending)
NAME        PHASE    STATE     RUNTIME   MODEL
alice       Ready    running   openclaw  qwen3.6-plus
bob         Ready    running   copaw     gpt-4
charlie     Pending  unknown   openclaw  -

Teams (1 total, 1 Ready)
NAME        PHASE   LEADER   READY
alpha-team  Ready   alice    2/2

Managers (1 total, 1 Ready)
NAME    PHASE   RUNTIME   MODEL
main    Ready   copaw     gpt-4

Humans (1 total, 1 Ready)
NAME   PHASE   DISPLAY-NAME
alice  Ready   Alice Smith

Tip: all resources healthy.
```

The output includes:
- **Mode**: Controller deployment mode (embedded, incluster)
- **Phase summary**: Count of resources in each phase (Ready, Failed, Pending)
- **Resource tables**: Sorted with non-Ready rows first, then alphabetically
- **Tip**: Contextual next-step hint (failed resource → inspect; provisioning → wait; all healthy → success)

### Credential Management

```bash
# Rotate Matrix AppService token
agt rotate appservice-token --as-token <new-token>

# Rotate AppService token with optional hs_token
agt rotate appservice-token --as-token <as-token> --hs-token <hs-token>

# Test LLM provider connectivity
agt llm-preflight

# Test LLM with specific provider
agt llm-preflight --provider openai-compat --api-key sk-xxx --model gpt-4
```

### Manager State

```bash
# Initialize task board
agt manager-state --action init

# Add finite task
agt manager-state --action add-finite --task-id T1 --title "Deploy service X" --assigned-to worker-a --room-id !room:example.com

# List tasks
agt manager-state --action list
```

## REST API

The REST API is implemented in [`agentteams-controller/internal/server/`](../agentteams-controller/internal/server/) and runs on port 8090. All endpoints require authentication (Bearer token) unless noted.

### Resource Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/workers` | Create worker |
| `GET` | `/api/v1/workers` | List workers (optional `?team=` filter) |
| `GET` | `/api/v1/workers/{name}` | Get worker |
| `GET` | `/api/v1/workers/{name}/events` | Get worker event history |
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
| `POST` | `/api/v1/workers/{name}/wake` | Wake sleeping worker |
| `POST` | `/api/v1/workers/{name}/sleep` | Sleep running worker |
| `POST` | `/api/v1/workers/{name}/ensure-ready` | Wait for worker provisioning |
| `POST` | `/api/v1/workers/{name}/ready` | Report worker ready |
| `GET` | `/api/v1/workers/{name}/status` | Get runtime status |

### Message Injection

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/managers/{name}/message` | Send message to manager's Admin DM room |
| `POST` | `/api/v1/teams/{name}/message` | Send message to team's leader channel |

**Request body:**
```json
{
  "body": "Please pause new task intake for the next hour."
}
```

**Response:**
```json
{
  "roomID": "!room:example.com",
  "sent": true
}
```

Message injection endpoints post a system-level message directly into a Manager's Admin DM room or a Team's leader channel, bypassing the normal agent chat loop. There is no explicit body size limit enforced; the standard Go `http.MaxBytesReader` default applies. For team messages, the endpoint prefers the leader↔admin DM room (`Status.LeaderDMRoomID`) and falls back to the shared team room (`Status.TeamRoomID`) when the DM hasn't been provisioned yet.

### Gateway Management

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/gateway/consumers` | Create gateway consumer |
| `POST` | `/api/v1/gateway/consumers/{id}/bind` | Bind consumer to routes |
| `DELETE` | `/api/v1/gateway/consumers/{id}` | Delete gateway consumer |
| `GET` | `/api/v1/gateway/providers` | List AI providers |
| `POST` | `/api/v1/gateway/providers` | Register AI provider |
| `DELETE` | `/api/v1/gateway/providers/{name}` | Delete AI provider |
| `GET` | `/api/v1/gateway/providers/{name}/models` | List provider's available models |

### Credentials & Auth

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/credentials/sts` | Refresh STS token (self-scoped to caller) |
| `POST` | `/api/v1/credentials/matrix-token` | Refresh Matrix token |
| `POST` | `/api/v1/appservice/rotate-token` | Rotate AppService token |

### Manager Tasks

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/manager-tasks` | Get Manager task board state (embedded mode only) |

### Packages & Status

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/packages` | Upload package to OSS (multipart/form-data, 64 MB max) |
| `GET` | `/api/v1/status` | Cluster status overview |
| `GET` | `/api/v1/version` | Controller version |
| `GET` | `/healthz` | Health check (no auth required) |

### Matrix AppService Endpoints

The controller also acts as a Matrix AppService for event push (when `AppServiceEnabled` is true):

| Method | Path | Description |
|--------|------|-------------|
| `PUT` | `/_matrix/app/v1/transactions/{txnId}` | Transaction push from homeserver |
| `GET` | `/_matrix/app/v1/users/{userId}` | User query |
| `GET` | `/_matrix/app/v1/rooms/{roomAlias}` | Room query |

## Handler Architecture

API handlers are in [`agentteams-controller/internal/server/`](../agentteams-controller/internal/server/):

| File | Purpose |
|------|---------|
| `http.go` | HTTP server setup, middleware, routing |
| `resource_handler.go` | CRUD handlers for all resource types |
| `lifecycle_handler.go` | Worker lifecycle (wake/sleep/ensure-ready) |
| `worker_health.go` | Worker health monitoring endpoints |
| `worker_resource_service.go` | Worker-specific resource operations |
| `message_handler.go` | Message injection endpoints |
| `gateway_handler.go` | Gateway consumer and provider management |
| `status_handler.go` | Health, status, and version endpoints |
| `credentials_handler.go` | STS and Matrix token refresh |
| `appservice_mgmt_handler.go` | AppService token rotation |
| `managertasks_handler.go` | Manager task board state |
| `package_handler.go` | Package upload to OSS |
| `http_metrics.go` | HTTP request metrics middleware |
| `types.go` | Request/response types |

## CLI Source Structure

The CLI is built with Cobra and lives in [`agentteams-controller/cmd/agt/`](../agentteams-controller/cmd/agt/):

| File | Purpose |
|------|---------|
| `main.go` | Entry point, root command |
| `create.go` | `create` subcommands (worker, team, human, manager) |
| `apply.go` | `apply -f` declarative management and `apply worker` subcommand |
| `update.go` | `update` subcommands (worker, team, manager) |
| `delete.go` | `delete` subcommands (worker, team, human, manager) |
| `get.go` | `get` subcommands (workers, teams, humans, managers) |
| `worker_cmd.go` | `worker` lifecycle subcommands (wake, sleep, ensure-ready, report-ready, status) |
| `status_cmd.go` | `status` cluster overview and `version` command |
| `manager_state_cmd.go` | `manager-state` task board management |
| `rotate_cmd.go` | `rotate appservice-token` command |
| `llm_preflight.go` | `llm-preflight` LLM connectivity test |
| `team_flags.go` | Shared team flag parsing |
| `client.go` | API client, HTTP helpers |
| `output.go` | Output formatting (tables, JSON, detail views) |

## YAML Apply Examples

### Worker YAML

```yaml
apiVersion: agentteams/v1beta1
kind: Worker
metadata:
  name: alice
spec:
  model: qwen3.6-plus
  runtime: openclaw
  identity: "A helpful coding assistant"
  skills:
    - github-operations
    - code-review
  mcpServers:
    - name: github
      url: https://mcp.github.com
  expose:
    - port: 8080
    - port: 3000
  resources:
    cpu: "1"
    memory: "2Gi"
```

### Team YAML

```yaml
apiVersion: agentteams/v1beta1
kind: Team
metadata:
  name: alpha-team
spec:
  description: "Backend development team"
  leader:
    name: team-leader
    model: gpt-4
    heartbeat:
      enabled: true
      every: "5m"
  workers:
    - name: worker-a
      model: qwen3.6-plus
    - name: worker-b
      model: claude-sonnet-4-6
  peerMentions: true
  channelPolicy:
    allowDirectMessages: true
```

### Manager YAML

```yaml
apiVersion: agentteams/v1beta1
kind: Manager
metadata:
  name: main
spec:
  model: gpt-4
  runtime: copaw
  config:
    maxConcurrentTasks: 5
    taskTimeout: "30m"
```

### Human YAML

```yaml
apiVersion: agentteams/v1beta1
kind: Human
metadata:
  name: alice
spec:
  displayName: "Alice Smith"
  email: alice@example.com
  permissionLevel: 2
  accessibleTeams:
    - alpha-team
  note: "Team lead"
```

### Multi-Document YAML

The `agt apply -f` command supports multi-document YAML files (separated by `---`):

```yaml
apiVersion: agentteams/v1beta1
kind: Worker
metadata:
  name: worker-a
spec:
  model: qwen3.6-plus
---
apiVersion: agentteams/v1beta1
kind: Worker
metadata:
  name: worker-b
spec:
  model: claude-sonnet-4-6
---
apiVersion: agentteams/v1beta1
kind: Team
metadata:
  name: alpha-team
spec:
  leader:
    name: team-leader
  workers:
    - name: worker-a
    - name: worker-b
```

Apply with:
```bash
agt apply -f resources.yaml
```

## Source References

- CLI commands: [`agentteams-controller/cmd/agt/`](../agentteams-controller/cmd/agt/)
- REST handlers: [`agentteams-controller/internal/server/`](../agentteams-controller/internal/server/)
<!-- openwiki: broken internal link [../agentteams-controller/internal/server/lifecycle_handler.go] file "../agentteams-controller/internal/server/lifecycle_handler.go" does not exist. Fix the href or restore the target, then delete this comment. -->
- Lifecycle handler: [`agentteams-controller/internal/server/lifecycle_handler.go`](../agentteams-controller/internal/server/lifecycle_handler.go)
<!-- openwiki: broken internal link [../agentteams-controller/internal/server/message_handler.go] file "../agentteams-controller/internal/server/message_handler.go" does not exist. Fix the href or restore the target, then delete this comment. -->
- Message handler: [`agentteams-controller/internal/server/message_handler.go`](../agentteams-controller/internal/server/message_handler.go)
<!-- openwiki: broken internal link [../agentteams-controller/internal/server/gateway_handler.go] file "../agentteams-controller/internal/server/gateway_handler.go" does not exist. Fix the href or restore the target, then delete this comment. -->
- Gateway handler: [`agentteams-controller/internal/server/gateway_handler.go`](../agentteams-controller/internal/server/gateway_handler.go)
<!-- openwiki: broken internal link [../agentteams-controller/internal/server/status_handler.go] file "../agentteams-controller/internal/server/status_handler.go" does not exist. Fix the href or restore the target, then delete this comment. -->
- Status handler: [`agentteams-controller/internal/server/status_handler.go`](../agentteams-controller/internal/server/status_handler.go)
<!-- openwiki: broken internal link [../agentteams-controller/internal/server/credentials_handler.go] file "../agentteams-controller/internal/server/credentials_handler.go" does not exist. Fix the href or restore the target, then delete this comment. -->
- Credentials handler: [`agentteams-controller/internal/server/credentials_handler.go`](../agentteams-controller/internal/server/credentials_handler.go)
<!-- openwiki: broken internal link [../agentteams-controller/internal/server/appservice_mgmt_handler.go] file "../agentteams-controller/internal/server/appservice_mgmt_handler.go" does not exist. Fix the href or restore the target, then delete this comment. -->
- AppService management: [`agentteams-controller/internal/server/appservice_mgmt_handler.go`](../agentteams-controller/internal/server/appservice_mgmt_handler.go)
<!-- openwiki: broken internal link [../agentteams-controller/internal/server/managertasks_handler.go] file "../agentteams-controller/internal/server/managertasks_handler.go" does not exist. Fix the href or restore the target, then delete this comment. -->
- Manager tasks handler: [`agentteams-controller/internal/server/managertasks_handler.go`](../agentteams-controller/internal/server/managertasks_handler.go)
<!-- openwiki: broken internal link [../agentteams-controller/internal/server/package_handler.go] file "../agentteams-controller/internal/server/package_handler.go" does not exist. Fix the href or restore the target, then delete this comment. -->
- Package handler: [`agentteams-controller/internal/server/package_handler.go`](../agentteams-controller/internal/server/package_handler.go)
<!-- openwiki: broken internal link [../agentteams-controller/internal/server/http_metrics.go] file "../agentteams-controller/internal/server/http_metrics.go" does not exist. Fix the href or restore the target, then delete this comment. -->
- HTTP metrics: [`agentteams-controller/internal/server/http_metrics.go`](../agentteams-controller/internal/server/http_metrics.go)
<!-- openwiki: broken internal link [../agentteams-controller/internal/server/types.go] file "../agentteams-controller/internal/server/types.go" does not exist. Fix the href or restore the target, then delete this comment. -->
- Request/response types: [`agentteams-controller/internal/server/types.go`](../agentteams-controller/internal/server/types.go)
