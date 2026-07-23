# Controller: CLI & API

The `agt` CLI and REST API are the primary interfaces for managing AgentTeams resources. The CLI is baked into Manager and Worker images. The REST API runs on the controller at port 8090.

## CLI Commands

The CLI is defined in [`agentteams-controller/cmd/agt/`](../../agentteams-controller/cmd/agt/). It communicates with the controller REST API.

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

# Update resources
agt update worker my-worker --model gpt-5
agt update team my-team --add-worker worker-c

# Delete resources
agt delete worker my-worker
agt delete team my-team

# Declarative apply (create or update from YAML)
agt apply -f worker.yaml
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

# Get worker runtime status
agt worker status my-worker
```

### Status & Monitoring

```bash
# Show cluster overview (all resources, one screen)
agt status

# Watch mode (auto-refresh)
agt status --watch
```

### Credential Management

```bash
# Rotate Matrix AppService token
agt rotate appservice-token

# Test LLM provider connectivity
agt llm-preflight
```

### Manager State

```bash
# Initialize task board
agt manager-state --action init

# Add finite task
agt manager-state --action add-finite --task "Deploy service X"

# List tasks
agt manager-state --action list
```

## REST API

The REST API is implemented in [`agentteams-controller/internal/server/`](../../agentteams-controller/internal/server/) and runs on port 8090.

### Resource Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/workers` | Create worker |
| `GET` | `/api/v1/workers` | List workers |
| `GET` | `/api/v1/workers/{name}` | Get worker |
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
| `POST` | `/api/v1/managers/{name}/message` | Send message to manager |
| `POST` | `/api/v1/teams/{name}/message` | Send message to team |

### Gateway Management

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/gateway/consumers` | Create gateway consumer |
| `POST` | `/api/v1/gateway/consumers/{id}/bind` | Bind consumer to routes |
| `DELETE` | `/api/v1/gateway/consumers/{id}` | Delete gateway consumer |

### Credentials & Auth

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/credentials/sts` | Refresh STS token |
| `POST` | `/api/v1/credentials/matrix-token` | Refresh Matrix token |
| `POST` | `/api/v1/appservice/rotate-token` | Rotate AppService token |

### Packages & Status

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/packages` | Upload package to OSS |
| `GET` | `/api/v1/status` | Cluster status overview |
| `GET` | `/api/v1/version` | Controller version |
| `GET` | `/healthz` | Health check |

### Matrix AppService Endpoints

The controller also acts as a Matrix AppService for event push:

| Method | Path | Description |
|--------|------|-------------|
| `PUT` | `/_matrix/app/v1/transactions/{txnId}` | Transaction push from homeserver |
| `GET` | `/_matrix/app/v1/users/{userId}` | User query |
| `GET` | `/_matrix/app/v1/rooms/{roomAlias}` | Room query |

## Handler Architecture

API handlers are in [`agentteams-controller/internal/server/`](../../agentteams-controller/internal/server/):

| File | Purpose |
|------|---------|
| `http.go` | HTTP server setup, middleware, routing |
| `resource_handler.go` | CRUD handlers for all resource types |
| `lifecycle_handler.go` | Worker lifecycle (wake/sleep/ensure-ready) |
| `worker_health.go` | Worker health monitoring endpoints |
| `worker_resource_service.go` | Worker-specific resource operations |
| `types.go` | Request/response types |

## CLI Source Structure

The CLI is built with Cobra and lives in [`agentteams-controller/cmd/agt/`](../../agentteams-controller/cmd/agt/):

| File | Purpose |
|------|---------|
| `main.go` | Entry point, root command |
| `create.go` | `create` subcommands |
| `apply.go` | `apply -f` declarative management |
| `update.go` | `update` subcommands |
| `status_cmd.go` | `status` cluster overview |
| `manager_state_cmd.go` | `manager-state` task board management |
| `team_flags.go` | Shared team flag parsing |

## Source References

- CLI commands: [`agentteams-controller/cmd/agt/`](../../agentteams-controller/cmd/agt/)
- REST handlers: [`agentteams-controller/internal/server/`](../../agentteams-controller/internal/server/)
- Lifecycle handler: [`agentteams-controller/internal/server/lifecycle_handler.go`](../../agentteams-controller/internal/server/lifecycle_handler.go)
- Health handler: [`agentteams-controller/internal/server/worker_health.go`](../../agentteams-controller/internal/server/worker_health.go)
