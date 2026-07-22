# Controller: CRDs & Reconcilers

The `hiclaw-controller` is a Go-based Kubernetes operator that reconciles five Custom Resource Definitions. It runs as a standalone binary and can operate in two modes: as a Kubernetes Deployment (using CRDs) or as an embedded process in a local Docker container (using a simulated CRD store).

## CRD Types

All types are defined under `agentteams.io/v1beta1` in [`hiclaw-controller/api/v1beta1/`](../../hiclaw-controller/api/v1beta1/).

### Worker (`workers.agentteams.io`)

Represents an AI agent worker. The most feature-rich CRD.

**Key spec fields:**
- `runtime` — Agent framework: `openclaw` (default), `copaw`, `hermes`, `openhuman`, `qwenpaw`
- `model` / `modelProvider` — LLM configuration
- `image` — Container image override
- `workerName` — Display name in Matrix
- `identity` / `soul` — Agent personality and behavior
- `skills` — List of skill packages to deploy
- `remoteSkills` — Skills fetched from Nacos registry
- `mcpServers` — Declarative MCP server configuration
- `package` — Agent package URI (file/http/nacos)
- `expose` — Port exposure configuration
- `channelPolicy` — Matrix channel access rules
- `channels` — Specific Matrix channels to join
- `resources` — Container resource requests/limits
- `idleTimeout` — Auto-sleep after inactivity
- `state` — Lifecycle state: `Running`, `Sleeping`, `Stopped`
- `deployMode` — `Local` or `Edge`
- `env` — Custom environment variables
- `volumes` / `mounts` — Storage configuration

**Status fields:** `phase`, `matrixUserId`, `roomId`, `containerState`, `heartbeatInfo`, `healthState`

**Source:** [`worker_types.go`](../../hiclaw-controller/api/v1beta1/worker_types.go)

### Manager (`managers.agentteams.io`)

Represents the coordinator agent.

**Key spec fields:**
- `runtime` — `openclaw` (default) or `copaw`
- `model` / `modelProvider` — LLM configuration
- `image` — Container image
- `soul` / `agents` — Agent personality
- `skills` — Manager skills to deploy
- `mcpServers` — MCP server configuration
- `state` — `Running`, `Sleeping`, `Stopped`
- `accessEntries` — Access control

**Source:** [`manager_types.go`](../../hiclaw-controller/api/v1beta1/manager_types.go)

### Team (`teams.agentteams.io`)

Groups workers under a team with coordination rules.

**Key spec fields:**
- `teamName` — Team identifier
- `workerMembers` — References to Worker CRs with roles (leader/worker)
- `humanMembers` — Human participants
- `admin` — Team administrator
- `channelPolicy` — Team-wide channel rules
- `modelProvider` — Override LLM provider for team
- `heartbeatEvery` — Heartbeat interval
- `peerMentions` — Cross-worker mention rules

**Status fields:** `phase`, `leaderRoomId`, `teamRoomId`, `adminDmRoomId`, `memberStates`

**Source:** [`team_types.go`](../../hiclaw-controller/api/v1beta1/team_types.go)

### Human (`humans.agentteams.io`)

Represents a human participant.

**Key spec fields:**
- `username` — Matrix username
- `displayName` — Display name
- `email` — Email address
- `permissionLevel` — 1=Admin, 2=Team, 3=Worker
- `accessibleTeams` — Teams this human can access
- `accessibleWorkers` — Workers this human can access
- `identitySource` — Authentication source (legacy/external SSO)

**Source:** [`human_types.go`](../../hiclaw-controller/api/v1beta1/human_types.go)

### Project (`projects.agentteams.io`)

Represents a team-scoped project with repositories and worker assignments.

**Key spec fields:**
- `team` — Owning team (required)
- `projectName` — Project identifier
- `description` — Project description
- `repos` — Repository references with access level (rw/ro)
- `workers` — Assigned workers
- `dependsOn` — Project dependencies (for DAG ordering)

**Status fields:** `phase`, `storageKey`, conditions: `StorageIdentityReady`, `ReposResolved`, `WorkersRecorded`, `MinIOProjected`

**Source:** [`project_types.go`](../../hiclaw-controller/api/v1beta1/project_types.go)

## Shared Types

Common types defined in [`types_shared.go`](../../hiclaw-controller/api/v1beta1/types_shared.go):
- `ResourceRequirements` — CPU/memory requests and limits
- `MCPServerConfig` — MCP server connection configuration
- `AccessEntry` — Access control entries
- `ChannelPolicy` — Matrix channel access rules
- `ExposeSpec` — Port exposure configuration
- `WorkerHealthState` — Health monitoring states (healthy/stalled/zombie/idle)

## Reconcilers

All reconcilers live in [`hiclaw-controller/internal/controller/`](../../hiclaw-controller/internal/controller/).

### WorkerReconciler (`worker_controller.go`)

Reconciles standalone Worker resources (team members are handled by TeamReconciler).

**Key operations:**
1. Manages finalizers for cleanup on deletion
2. Provisions infrastructure: Matrix user, rooms, gateway consumers
3. Deploys packages and configs to MinIO
4. Handles lifecycle state transitions (Running/Sleeping/Stopped)
5. Manages edge worker heartbeat timeouts
6. Monitors health states via HealthMonitorController

### TeamReconciler (`team_controller.go`)

Reconciles Team resources that reference existing Worker CRs.

**Key operations:**
1. Creates team rooms (main room, leader DM, admin DM)
2. Provisions team members (leader + workers) into rooms
3. Injects coordination context and runtime configs
4. Handles team-wide channel policies
5. Manages team lifecycle and member states
6. Handles worker member decoupling (legacy vs decoupled mode)

**Supporting files:**
- `team_members_decoupled.go` — New member management via Worker CR references
- `team_members_legacy.go` — Legacy inline worker definitions
- `team_channel_policy.go` — Channel policy enforcement
- `team_rooms.go` — Room creation and management
- `team_runtime_config.go` — Runtime configuration injection
- `team_status.go` — Status updates

### ManagerReconciler (`manager_controller.go`)

Reconciles Manager resources.

**Key operations:**
1. Provisions manager containers
2. Manages admin DM rooms
3. Handles welcome messages and onboarding
4. Manages gateway auth and credentials

### HumanReconciler (`human_controller.go`)

Reconciles Human resources.

**Key operations:**
1. Creates/updates Matrix accounts for humans
2. Manages room memberships based on accessibleTeams/Workers
3. Syncs display names
4. Handles identity sources

### ProjectReconciler (`project_controller.go`)

Reconciles Project resources.

**Key operations:**
1. Ensures storage identity (MinIO paths)
2. Manages project `manifest.json`
3. Handles project dependencies
4. Manages archive/completion states

### AutoSleepController (`auto_sleep_controller.go`)

Monitors worker heartbeats and auto-sleeps idle workers.

### HealthMonitorController (`health_monitor_controller.go`)

Monitors worker health states: healthy, stalled, zombie, idle. Reports findings for the dashboard status view.

## Service Layers

Services in [`hiclaw-controller/internal/service/`](../../hiclaw-controller/internal/service/) provide the core logic used by reconcilers.

### Provisioner (`provisioner.go`)

Orchestrates infrastructure provisioning:
- `ProvisionWorker` / `DeprovisionWorker` — Full worker lifecycle
- `ProvisionManager` / `DeprovisionManager` — Manager lifecycle
- `EnsureWorkerGatewayAuth` / `EnsureManagerGatewayAuth` — Gateway consumer setup
- Credential management: Matrix tokens, gateway keys, MinIO passwords, STS tokens

**Specialized files:**
- `provisioner_credentials.go` — Credential generation and rotation
- `provisioner_manager.go` — Manager-specific provisioning
- `provisioner_team_rooms.go` — Team room provisioning
- `provisioner_worker_phases.go` — Worker lifecycle phases

### Deployer (`deployer.go`)

Handles config deployment and package management:
- `DeployPackage` — Deploy agent packages (file/http/nacos URIs)
- `WriteInlineConfigs` — Write inline configuration files
- `DeployMemberRuntimeConfig` — Per-member runtime configuration
- `InjectCoordinationContext` — Team coordination context injection
- `PushOnDemandSkills` — Push skills to workers on demand
- `PrepareWorkerDeps` — Prepare worker dependencies

**Specialized files:**
- `deployer_coordination.go` — Team coordination injection
- `deployer_manager.go` — Manager-specific deployment
- `deployer_merge.go` — Config merging logic
- `deployer_remote_skills.go` — Remote skill fetching from Nacos
- `deployer_worker_config.go` — Worker configuration deployment

### Interfaces (`interfaces.go`)

Defines service interfaces for testability:
- `WorkerProvisioner` / `WorkerDeployer`
- `ManagerProvisioner` / `ManagerDeployer`
- `HumanProvisioner`
- `WorkerEnvBuilderI` / `ManagerEnvBuilderI`

## Gateway Integration

Gateway clients in [`hiclaw-controller/internal/gateway/`](../../hiclaw-controller/internal/gateway/) abstract different gateway backends.

### HigressClient (`higress.go`)

Manages self-hosted Higress gateway via Console API:
- Session management (login, password changes)
- Consumer management (create/delete)
- AI route authorization
- Port exposure for workers
- Service source and route management

### AIGatewayClient (`aigateway.go`)

Manages Alibaba Cloud APIG:
- Consumer management via APIG SDK
- Authorization rules
- Model API binding

### Client Interface (`client.go`)

Unified interface:
- `ConsumerClient` — `EnsureConsumer`, `DeleteConsumer`, `AuthorizeAIRoutes`
- `PortExposeClient` — `ExposePort`, `UnexposePort`
- `InfrastructureClient` — `EnsureServiceSource`, `EnsureRoute`, `EnsureAIProvider`

## Matrix Integration

Matrix clients in [`hiclaw-controller/internal/matrix/`](../../hiclaw-controller/internal/matrix/):
- `client.go` — Base client setup
- `client_http.go` — HTTP-level Matrix API calls
- `client_messages.go` — Message sending
- `client_rooms.go` — Room creation and management
- `client_users.go` — User provisioning

## Package Organization

The internal package map is documented in [`hiclaw-controller/internal/AGENTS.md`](../../hiclaw-controller/internal/AGENTS.md). Key packages:

| Package | Purpose |
|---------|---------|
| `controller/` | Reconciler implementations |
| `service/` | Provisioner and deployer logic |
| `gateway/` | Higress/AIGateway client abstraction |
| `matrix/` | Matrix homeserver client |
| `server/` | REST API handlers |
| `backend/` | Container backend (Docker/Podman/K8s) |
| `config/` | Configuration loading and derivation |
| `agentconfig/` | Agent config file generation |
| `metrics/` | Prometheus metrics |
| `managerstate/` | Manager task board state |
| `workerdeps/` | Worker dependency manifests |

## Source References

- CRD types: [`hiclaw-controller/api/v1beta1/`](../../hiclaw-controller/api/v1beta1/)
- Reconcilers: [`hiclaw-controller/internal/controller/`](../../hiclaw-controller/internal/controller/)
- Services: [`hiclaw-controller/internal/service/`](../../hiclaw-controller/internal/service/)
- Gateway: [`hiclaw-controller/internal/gateway/`](../../hiclaw-controller/internal/gateway/)
- Matrix: [`hiclaw-controller/internal/matrix/`](../../hiclaw-controller/internal/matrix/)
- Package map: [`hiclaw-controller/internal/AGENTS.md`](../../hiclaw-controller/internal/AGENTS.md)
