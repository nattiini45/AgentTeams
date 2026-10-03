---
type: reference
title: Helm Chart Reference
description: Detailed Helm chart values reference for AgentTeams deployment, covering every values.yaml section, chart structure, CRD manifests, and customization patterns.
tags: [helm, kubernetes, deployment, configuration, values, crd, templates]
verified:
  - by: openwiki/0.7.0
    at: 2026-10-03T08:15:26.963Z
sources:
  - id: openwiki-source-47c1eb8cedca9425dc039eae
    resource: repo://helm/agentteams/Chart.yaml
  - id: openwiki-source-8ec7bc8f5105ef7c5b8b0b5a
    resource: repo://helm/agentteams/crds/humans.agentteams.io.yaml
  - id: openwiki-source-85959c22bbf788647eccac38
    resource: repo://helm/agentteams/crds/managers.agentteams.io.yaml
  - id: openwiki-source-b1737fa90ff10e709bdea7a0
    resource: repo://helm/agentteams/crds/projects.agentteams.io.yaml
  - id: openwiki-source-ae899b06efb26639cc69b0a4
    resource: repo://helm/agentteams/crds/teams.agentteams.io.yaml
  - id: openwiki-source-daf493ad855a647cda263042
    resource: repo://helm/agentteams/crds/workers.agentteams.io.yaml
  - id: openwiki-source-d05dace50d6e6e04bd614aa4
    resource: repo://helm/agentteams/templates/_helpers.infra.tpl
  - id: openwiki-source-faea9fd9f6838fe95c99eefc
    resource: repo://helm/agentteams/templates/_helpers.tpl
  - id: openwiki-source-d724d19ec0dcf697acc9f5b5
    resource: repo://helm/agentteams/templates/controller/deployment.yaml
  - id: openwiki-source-30cb2681d4fb44518d02960e
    resource: repo://helm/agentteams/templates/controller/uninstall-hook.yaml
  - id: openwiki-source-357a42fbb67ba97649c8b98a
    resource: repo://helm/agentteams/templates/preflight/llm-job.yaml
  - id: openwiki-source-124a12417423aecf5b9b4da7
    resource: repo://helm/agentteams/values.yaml
generated: { by: "openwiki/0.7.0", at: "2026-10-03T08:15:26.963Z" }
---

# Helm Chart Reference

The AgentTeams Helm chart (`helm/agentteams`) packages the complete AgentTeams platform for Kubernetes deployment. This reference documents every configurable value, the chart's template structure, included CRDs, and customization patterns for both local and cloud deployments.

> For quick-start installation commands, see [Install & Deploy](./installation.md).

## Chart Metadata

| Field | Value |
|-------|-------|
| Name | `agentteams` |
| Type | `application` |
| Version | `1.1.1` |
| App Version | `1.1.1` |
| Keywords | `ai-agent`, `multi-agent`, `matrix`, `collaboration`, `llm` |

**Dependencies:**

| Dependency | Version | Repository | Condition |
|------------|---------|------------|-----------|
| `higress` | `2.2.1` | `https://higress.io/helm-charts` | `gateway.higress.enabled` |

The Higress subchart is deployed only when `gateway.provider=higress` and `gateway.mode=managed`. Its values are passed through the top-level `higress:` block because the dependency name remains `higress`.

## Component Dependency Tree

<!-- openwiki: mermaid parse failed and this diagram was converted to a text fence so it does not break rendering. Fix the diagram source and restore the mermaid fence. Parser error: Heuristic: an unescaped angle bracket inside a label breaks rendering; rephrase the label. -->
```text
flowchart TD
    subgraph chart["AgentTeams Helm Chart"]
        CRDs["CRDs<br/>humans, managers, projects, teams, workers"]
        Controller["Controller Deployment"]
        Secrets["Runtime Env Secret"]
        PreflightJob["Preflight LLM Check Job"]
        UninstallHook["Uninstall Cleanup Job"]
    end

    subgraph infra["Managed Infrastructure"]
        Tuwunel["Tuwunel<br/>(Matrix Homeserver)"]
        HigressGW["Higress Gateway<br/>(AI Gateway)"]
        MinIO["MinIO<br/>(Object Storage)"]
    end

    subgraph optional["Optional Components"]
        ElementWeb["Element Web<br/>(IM UI)"]
        Dashboard["Dashboard<br/>(Operator UI)"]
    end

    subgraph crd_driven["CRD-Driven Resources"]
        ManagerCR["Manager CR"]
        WorkerCRs["Worker CRs"]
        TeamCRs["Team CRs"]
        HumanCRs["Human CRs"]
    end

    chart --> CRDs
    chart --> Controller
    chart --> Secrets
    chart --> PreflightJob
    chart --> UninstallHook

    Controller --> Tuwunel
    Controller --> HigressGW
    Controller --> MinIO

    Tuwunel -.->|"matrix.provider=tuwunel"| Controller
    HigressGW -.->|"gateway.provider=higress"| Controller
    MinIO -.->|"storage.provider=minio"| Controller

    Controller --> ManagerCR
    ManagerCR --> WorkerCRs
    TeamCRs --> WorkerCRs
    HumanCRs --> WorkerCRs

    chart --> ElementWeb
    chart --> Dashboard
```

*Component dependency tree showing how the Helm chart's templates relate to managed infrastructure and CRD-driven resources.*

## Template Structure

The chart organizes templates into functional directories:

```
templates/
├── _helpers.tpl                 # Naming, labeling, and image helpers
├── _helpers.infra.tpl           # Infrastructure abstraction (matrix, gateway, storage URLs)
├── 00-validate.yaml             # Pre-install validation of required values
├── NOTES.txt                    # Post-install usage instructions
├── controller/
│   ├── deployment.yaml          # Controller Deployment with optional credential-provider sidecar
│   ├── rbac.yaml                # ClusterRole, ClusterRoleBinding for CRD management
│   ├── service.yaml             # Controller Service (metrics + webhook)
│   ├── serviceaccount.yaml      # Controller ServiceAccount
│   ├── servicemonitor.yaml      # Prometheus ServiceMonitor (optional)
│   └── uninstall-hook.yaml      # Pre-delete Job that cleans up CRs
├── dashboard/
│   └── deployment.yaml          # Operator Dashboard (optional)
├── element-web/
│   └── deployment.yaml          # Element Web IM UI (optional)
├── gateway/
│   └── _placeholder.tpl         # Gateway-specific helpers
├── matrix/
│   └── deployment.yaml          # Tuwunel Matrix homeserver (when managed)
├── preflight/
│   └── llm-check-job.yaml       # Pre-install/upgrade LLM connectivity probe
├── secrets/
│   └── runtime-env.yaml         # Shared runtime secret (credentials, endpoints)
├── storage/
│   └── minio.yaml               # MinIO StatefulSet (when managed)
```

## CRD Manifests

The chart includes five Custom Resource Definitions under `crds/`:

| CRD | Group | Version | Description |
|-----|-------|---------|-------------|
| `humans.agentteams.io` | `agentteams.io` | `v1beta1` | Human participant accounts with Matrix credentials |
| `managers.agentteams.io` | `agentteams.io` | `v1beta1` | Manager agent specifications (model, runtime, image) |
| `projects.agentteams.io` | `agentteams.io` | `v1beta1` | Project groupings for organizing teams and workers |
| `teams.agentteams.io` | `agentteams.io` | `v1beta1` | Team definitions with worker and human memberships |
| `workers.agentteams.io` | `agentteams.io` | `v1beta1` | Worker agent specifications (model, runtime, image, skills) |

**Worker CRD Key Fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `spec.model` | `string` | Yes | LLM model ID (e.g., `claude-sonnet-4-6`, `qwen3.6-plus`) |
| `spec.runtime` | `string` | No | Agent runtime: `openclaw`, `copaw`, `hermes`, `qwenpaw`, `openhuman` |
| `spec.image` | `string` | No | Custom Docker image (overrides default for runtime) |
| `spec.workerName` | `string` | No | Business/runtime identity for Matrix localpart and OSS path key |
| `spec.identity` | `string` | No | Worker public identity (generates `IDENTITY.md`) |
| `spec.soul` | `string` | No | Worker identity and role definition (generates `SOUL.md`) |
| `spec.agents` | `string` | No | Agent behavior rules (generates `AGENTS.md`) |
| `spec.skills` | `[]string` | No | List of skill identifiers for this worker |
| `spec.modelProvider` | `string` | No | Name of AI gateway Model API for LLM provider |

## Values Reference

### Global Settings

| Value Path | Type | Default | Description |
|------------|------|---------|-------------|
| `global.namespace` | `string` | `""` | Override namespace (defaults to `.Release.Namespace`) |
| `global.imageRegistry` | `string` | `higress-registry.cn-hangzhou.cr.aliyuncs.com/agentteams` | Default registry/namespace for AgentTeams images (controller, manager, workers). Component images under Higress org keep full repository paths. |
| `global.imageTag` | `string` | `""` | Global image tag override (defaults to `v{Chart.AppVersion}`) |
| `imagePullSecrets` | `[]string` | `[]` | Image pull secrets for private registries |

### Credentials

Shared across all components via the runtime-env Secret.

| Value Path | Type | Default | Description |
|------------|------|---------|-------------|
| `credentials.registrationToken` | `string` | `""` | Matrix registration token (auto-generated if empty) |
| `credentials.adminUser` | `string` | `"admin"` | Matrix admin username |
| `credentials.adminPassword` | `string` | `""` | Matrix admin password (auto-generated if empty) |
| `credentials.llmApiKey` | `string` | `""` | LLM API key (**required**) |
| `credentials.llmProvider` | `string` | `"openai-compat"` | LLM provider type |
| `credentials.defaultModel` | `string` | `"gpt-5.4"` | Default LLM model for manager and workers |
| `credentials.llmBaseUrl` | `string` | `""` | OpenAI-compatible base URL (e.g., `https://api.openai.com/v1`) |

### Preflight LLM Check

The chart runs a Helm pre-install/pre-upgrade Job that validates LLM connectivity before proceeding. This prevents deploying a non-functional cluster due to misconfigured credentials or network issues.

| Value Path | Type | Default | Description |
|------------|------|---------|-------------|
| `preflight.llm.enabled` | `bool` | `true` | Run LLM probe on install/upgrade |
| `preflight.llm.strict` | `bool` | `true` | Fail install/upgrade when probe fails |
| `preflight.llm.timeoutSeconds` | `int` | `30` | Per HTTP request timeout |
| `preflight.llm.retries` | `int` | `2` | Retry transient network/429/5xx failures |
| `preflight.llm.activeDeadlineSeconds` | `int` | `120` | Hard ceiling for the hook Job |
| `preflight.llm.resources` | `object` | `{}` | Resource requests/limits for the preflight Job |

### Matrix

| Value Path | Type | Default | Description |
|------------|------|---------|-------------|
| `matrix.provider` | `string` | `tuwunel` | Matrix homeserver provider: `tuwunel` or `synapse` |
| `matrix.mode` | `string` | `managed` | Deployment mode: `managed` (chart deploys homeserver) or `existing` (external) |
| `matrix.internalURL` | `string` | `""` | Existing mode only; managed mode auto-derives from Tuwunel service |
| `matrix.serverName` | `string` | `""` | Existing mode only; managed mode auto-derives from Tuwunel FQDN |
| `matrix.appservice.enabled` | `bool` | `true` | Enable Matrix AppService mode (controller registers as appservice to provision accounts) |
| `matrix.appservice.asToken` | `string` | `""` | Override AppService as_token (auto-generated if empty) |
| `matrix.appservice.hsToken` | `string` | `""` | Override AppService hs_token (auto-generated if empty) |

**Tuwunel Configuration (`matrix.tuwunel.*`):**

| Value Path | Type | Default | Description |
|------------|------|---------|-------------|
| `matrix.tuwunel.image.repository` | `string` | `higress-registry.cn-hangzhou.cr.aliyuncs.com/higress/tuwunel` | Tuwunel image repository |
| `matrix.tuwunel.image.tag` | `string` | `"20260216"` | Tuwunel image tag |
| `matrix.tuwunel.image.pullPolicy` | `string` | `IfNotPresent` | Image pull policy |
| `matrix.tuwunel.newUserDisplayNameSuffix` | `string` | `""` | Suffix for new user display names (empty disables Tuwunel's default) |
| `matrix.tuwunel.replicaCount` | `int` | `1` | Number of Tuwunel replicas |
| `matrix.tuwunel.resources` | `object` | See below | Resource requests and limits |
| `matrix.tuwunel.service.type` | `string` | `ClusterIP` | Service type |
| `matrix.tuwunel.service.port` | `int` | `6167` | Tuwunel service port |
| `matrix.tuwunel.persistence.enabled` | `bool` | `true` | Enable persistent storage |
| `matrix.tuwunel.persistence.size` | `string` | `10Gi` | Persistent volume size |
| `matrix.tuwunel.persistence.storageClassName` | `string` | `""` | Storage class name (empty = default) |
| `matrix.tuwunel.persistence.mountPath` | `string` | `/data/conduwuit` | Mount path for persistent volume |
| `matrix.tuwunel.extraEnv` | `object` | `{}` | Additional environment variables |

### Gateway

| Value Path | Type | Default | Description |
|------------|------|---------|-------------|
| `gateway.provider` | `string` | `higress` | Gateway provider: `higress` (local) or `ai-gateway` (Alibaba Cloud APIG) |
| `gateway.mode` | `string` | `managed` | Deployment mode: `managed` (chart deploys) or `existing` (external) |
| `gateway.publicURL` | `string` | `""` | **Required** browser/public URL (e.g., `http://localhost:18080`) |
| `gateway.higress.enabled` | `bool` | `true` | Controls Higress subchart deployment |

**AI Gateway Configuration (`gateway.aiGateway.*`):**

Required when `gateway.provider=ai-gateway`.

| Value Path | Type | Default | Description |
|------------|------|---------|-------------|
| `gateway.aiGateway.region` | `string` | `""` | Alibaba Cloud region (e.g., `cn-hangzhou`) |
| `gateway.aiGateway.endpoint` | `string` | `""` | Optional APIG OpenAPI endpoint override |
| `gateway.aiGateway.gatewayId` | `string` | `""` | APIG gateway instance ID (**required**) |
| `gateway.aiGateway.modelApiId` | `string` | `""` | LLM Model API ID for consumer authorization |
| `gateway.aiGateway.envId` | `string` | `""` | APIG environment ID |

### Storage

| Value Path | Type | Default | Description |
|------------|------|---------|-------------|
| `storage.provider` | `string` | `minio` | Storage provider: `minio` (local) or `oss` (Alibaba Cloud OSS) |
| `storage.mode` | `string` | `managed` | Deployment mode: `managed` (chart deploys MinIO) or `existing` (external) |
| `storage.bucket` | `string` | `"agentteams-storage"` | Canonical AgentTeams bucket name (does not auto-migrate older bucket names) |

**OSS Configuration (`storage.oss.*`):**

Required when `storage.provider=oss`.

| Value Path | Type | Default | Description |
|------------|------|---------|-------------|
| `storage.oss.region` | `string` | `""` | Alibaba Cloud region (e.g., `cn-hangzhou`) |
| `storage.oss.endpoint` | `string` | `""` | Explicit endpoint override (credential-provider returns correct endpoint normally) |

**MinIO Configuration (`storage.minio.*`):**

| Value Path | Type | Default | Description |
|------------|------|---------|-------------|
| `storage.minio.image.repository` | `string` | `higress-registry.cn-hangzhou.cr.aliyuncs.com/higress/minio` | MinIO image repository |
| `storage.minio.image.tag` | `string` | `"20260216"` | MinIO image tag |
| `storage.minio.image.pullPolicy` | `string` | `IfNotPresent` | Image pull policy |
| `storage.minio.resources` | `object` | See below | Resource requests and limits |
| `storage.minio.service.type` | `string` | `ClusterIP` | Service type |
| `storage.minio.service.apiPort` | `int` | `9000` | MinIO API port |
| `storage.minio.service.consolePort` | `int` | `9001` | MinIO console port |
| `storage.minio.persistence.enabled` | `bool` | `true` | Enable persistent storage |
| `storage.minio.persistence.size` | `string` | `10Gi` | Persistent volume size |
| `storage.minio.persistence.storageClassName` | `string` | `""` | Storage class name (empty = default) |
| `storage.minio.auth.rootUser` | `string` | `"minioadmin"` | MinIO root user |
| `storage.minio.auth.rootPassword` | `string` | `"minioadmin"` | MinIO root password |

### Higress Subchart Values

Passed directly to the Higress dependency. The condition flag is `gateway.higress.enabled`. See [Higress documentation](https://higress.io/en/docs/latest/user/configurations/) for full options.

| Value Path | Type | Default | Description |
|------------|------|---------|-------------|
| `higress.global.local` | `bool` | `true` | Local mode (kind/minikube): no cloud LoadBalancer |
| `higress.higress-core.gateway.replicas` | `int` | `1` | Gateway replicas |
| `higress.higress-core.gateway.httpPort` | `int` | `80` | HTTP port |
| `higress.higress-core.gateway.httpsPort` | `int` | `443` | HTTPS port |
| `higress.higress-core.gateway.service.type` | `string` | `ClusterIP` | Service type |
| `higress.higress-core.gateway.service.ports` | `list` | See below | Service port definitions |
| `higress.higress-core.controller.replicas` | `int` | `1` | Controller replicas |
| `higress.higress-core.controller.image` | `string` | `higress` | Controller image |
| `higress.higress-console.admin.password` | `string` | `""` | Console admin password (empty → controller initializes via `/system/init`) |

### Credential Provider Sidecar

The credential provider sidecar issues scoped STS tokens to the controller (for APIG/OSS SDK calls) and to workers (via `POST /api/v1/credentials/sts`). Required when `gateway.provider=ai-gateway` or `storage.provider=oss`. No default image is provided: real deployments use a customer-specific RAM-role-issuing service. For local development, provide a mock implementation of the same API.

| Value Path | Type | Default | Description |
|------------|------|---------|-------------|
| `credentialProvider.enabled` | `bool` | `false` | Auto-forced to `true` when any cloud provider is selected |
| `credentialProvider.image.repository` | `string` | `""` | Image repository (e.g., `registry.example.com/agentteams/credential-provider`) |
| `credentialProvider.image.tag` | `string` | `""` | Image tag |
| `credentialProvider.image.pullPolicy` | `string` | `IfNotPresent` | Image pull policy |
| `credentialProvider.port` | `int` | `17070` | Sidecar HTTP port |
| `credentialProvider.resources` | `object` | See below | Resource requests and limits |
| `credentialProvider.env` | `object` | `{}` | Additional environment variables as key/value map |
| `credentialProvider.envFrom` | `[]object` | `[]` | Raw `envFrom[]` entries (secretRef / configMapRef) |

**Cloud Deployment Pattern:**

When deploying with cloud providers, the credential provider sidecar is injected into the controller pod. It provides:

1. **Controller STS tokens** — For APIG and OSS SDK calls when `gateway.provider=ai-gateway` or `storage.provider=oss`
2. **Worker STS tokens** — Via `POST /api/v1/credentials/sts` endpoint for workers to access cloud resources

```yaml
# Example: Alibaba Cloud deployment with AI Gateway and OSS
gateway:
  provider: ai-gateway
  mode: existing
  publicURL: "https://api.example.com"
  aiGateway:
    region: "cn-hangzhou"
    gatewayId: "gw-12345"
    modelApiId: "api-67890"
    envId: "env-abcde"

storage:
  provider: oss
  mode: existing
  bucket: "my-agentteams-bucket"
  oss:
    region: "cn-hangzhou"

credentialProvider:
  enabled: true
  image:
    repository: "registry.example.com/agentteams/credential-provider"
    tag: "v1.0.0"
```

### Controller

| Value Path | Type | Default | Description |
|------------|------|---------|-------------|
| `controller.replicaCount` | `int` | `1` | Number of controller replicas (increase for HA with leader election) |
| `controller.image.repository` | `string` | `higress-registry.cn-hangzhou.cr.aliyuncs.com/agentteams/agentteams-controller` | Controller image repository |
| `controller.image.tag` | `string` | `""` | Image tag (defaults to `global.imageTag`) |
| `controller.image.pullPolicy` | `string` | `IfNotPresent` | Image pull policy |
| `controller.service.type` | `string` | `ClusterIP` | Service type |
| `controller.service.port` | `int` | `8090` | Controller service port |
| `controller.service.targetPort` | `int` | `8090` | Container target port |
| `controller.metrics.enabled` | `bool` | `true` | Enable metrics endpoint |
| `controller.metrics.bindAddress` | `string` | `":8080"` | Metrics bind address |
| `controller.metrics.port` | `int` | `8080` | Metrics service port |
| `controller.metrics.targetPort` | `int` | `8080` | Metrics container port |
| `controller.metrics.serviceMonitor.enabled` | `bool` | `false` | Create Prometheus ServiceMonitor |
| `controller.metrics.serviceMonitor.interval` | `string` | `30s` | Scrape interval |
| `controller.metrics.serviceMonitor.scrapeTimeout` | `string` | `10s` | Scrape timeout |
| `controller.metrics.serviceMonitor.labels` | `object` | `{}` | Additional labels for ServiceMonitor |
| `controller.resources` | `object` | See below | Resource requests and limits |
| `controller.workerBackend` | `string` | `"k8s"` | Worker provisioning backend |
| `controller.resourcePrefix` | `string` | `"agentteams-"` | Prefix for Kubernetes resource names |
| `controller.resourceAutoPrefix` | `bool` | `true` | Auto-add prefix to resource names |
| `controller.serviceAccount.create` | `bool` | `true` | Create ServiceAccount |
| `controller.serviceAccount.name` | `string` | `""` | ServiceAccount name (defaults to fullname) |
| `controller.serviceAccount.annotations` | `object` | `{}` | ServiceAccount annotations |
| `controller.env` | `object` | `{}` | Additional environment variables |
| `controller.timezone` | `string` | `"Asia/Shanghai"` | Timezone for controller pod |

**Uninstall Hook (`controller.uninstallHook.*`):**

When enabled, `helm uninstall` first runs a Job that deletes all Manager/Worker/Team/Human CRs while the controller is still alive, so the controller's finalizer logic can clean up Pods, Matrix users, and OSS data. Set to `false` to skip auto-cleanup (you must then delete those CRs manually before `helm uninstall`). The hook reuses the controller image (which ships kubectl) so it does not depend on Docker Hub being reachable from the cluster.

| Value Path | Type | Default | Description |
|------------|------|---------|-------------|
| `controller.uninstallHook.enabled` | `bool` | `true` | Enable uninstall cleanup hook |
| `controller.uninstallHook.timeoutSeconds` | `int` | `300` | Per-resource-group `kubectl delete --wait` timeout |
| `controller.uninstallHook.backoffLimit` | `int` | `1` | Job backoff limit |
| `controller.uninstallHook.activeDeadlineSeconds` | `int` | `1500` | Hard ceiling (4 groups × timeoutSeconds + headroom) |
| `controller.uninstallHook.resources` | `object` | `{}` | Resource requests/limits for cleanup Job |

### Manager Agent

When enabled, the controller creates a Manager CR at startup. The ManagerReconciler then provisions the Matrix account, Gateway consumer, and Pod automatically. No static Deployment is created by Helm.

| Value Path | Type | Default | Description |
|------------|------|---------|-------------|
| `manager.enabled` | `bool` | `true` | Create Manager CR during cluster initialization |
| `manager.model` | `string` | `""` | LLM model (defaults to `credentials.defaultModel`) |
| `manager.runtime` | `string` | `"openclaw"` | Runtime: `openclaw`, `copaw`, or `hermes` |
| `manager.image.repository` | `string` | `higress-registry.cn-hangzhou.cr.aliyuncs.com/agentteams/agentteams-manager` | Manager image repository |
| `manager.image.tag` | `string` | `""` | Image tag (defaults to `global.imageTag`) |
| `manager.resources` | `object` | See below | Resource requests and limits |

### Element Web (IM UI)

| Value Path | Type | Default | Description |
|------------|------|---------|-------------|
| `elementWeb.enabled` | `bool` | `true` | Deploy Element Web UI |
| `elementWeb.image.repository` | `string` | `higress-registry.cn-hangzhou.cr.aliyuncs.com/higress/element-web` | Element Web image repository |
| `elementWeb.image.tag` | `string` | `"20260216"` | Element Web image tag |
| `elementWeb.image.pullPolicy` | `string` | `IfNotPresent` | Image pull policy |
| `elementWeb.replicaCount` | `int` | `1` | Number of replicas |
| `elementWeb.resources` | `object` | See below | Resource requests and limits |
| `elementWeb.service.type` | `string` | `ClusterIP` | Service type |
| `elementWeb.service.port` | `int` | `8080` | Service port |

### Dashboard (Operator UI)

| Value Path | Type | Default | Description |
|------------|------|---------|-------------|
| `dashboard.enabled` | `bool` | `false` | Deploy operator dashboard |
| `dashboard.replicaCount` | `int` | `1` | Number of replicas |
| `dashboard.image.repository` | `string` | `higress-registry.cn-hangzhou.cr.aliyuncs.com/higress/agentteams-dashboard` | Dashboard image repository |
| `dashboard.image.tag` | `string` | `""` | Image tag (defaults to `global.imageTag`) |
| `dashboard.image.pullPolicy` | `string` | `IfNotPresent` | Image pull policy |
| `dashboard.bindHost` | `string` | `"0.0.0.0"` | Bind host |
| `dashboard.username` | `string` | `"admin"` | Dashboard username |
| `dashboard.password` | `string` | `""` | Dashboard password (empty → auto-generated and preserved in release Secret) |
| `dashboard.existingSecret` | `string` | `""` | Use existing secret for password |
| `dashboard.existingSecretKey` | `string` | `"password"` | Key in existing secret |
| `dashboard.publicOrigin` | `string` | `""` | Public origin URL |
| `dashboard.authDisabled` | `bool` | `false` | Disable authentication |
| `dashboard.serviceAccount.create` | `bool` | `true` | Create ServiceAccount |
| `dashboard.serviceAccount.name` | `string` | `""` | ServiceAccount name (defaults to `{controller.resourcePrefix}admin`) |
| `dashboard.service.type` | `string` | `ClusterIP` | Service type |
| `dashboard.service.port` | `int` | `8090` | Service port |
| `dashboard.upstreamTimeoutMs` | `int` | `15000` | Upstream request timeout (ms) |
| `dashboard.maxJsonBytes` | `int` | `2097152` | Max JSON payload size (bytes) |
| `dashboard.maxObjectBytes` | `int` | `16777216` | Max object payload size (bytes) |
| `dashboard.resources` | `object` | See below | Resource requests and limits |

### CMS Observability

Alibaba Cloud CMS 2.0 (ARMS) integration. Set `cms.enabled=true` and fill in credentials to enable OTLP trace/metric export from Manager and all Workers.

| Value Path | Type | Default | Description |
|------------|------|---------|-------------|
| `cms.enabled` | `bool` | `false` | Enable CMS observability |
| `cms.endpoint` | `string` | `""` | OTLP endpoint |
| `cms.licenseKey` | `string` | `""` | CMS license key |
| `cms.project` | `string` | `""` | CMS project name |
| `cms.workspace` | `string` | `""` | CMS workspace |
| `cms.serviceName` | `string` | `"agentteams-manager"` | Service name (Workers auto-derive `agentteams-worker-<name>`) |
| `cms.metricsEnabled` | `bool` | `false` | Enable metrics export |

### Worker Defaults

The `worker.defaultImage` section maps each runtime to its container image. When a Worker CR omits `spec.image`, the controller selects the image based on `spec.runtime` (or `worker.defaultRuntime` if unset).

| Value Path | Type | Default | Description |
|------------|------|---------|-------------|
| `worker.defaultImage.openclaw.repository` | `string` | `higress-registry.cn-hangzhou.cr.aliyuncs.com/agentteams/agentteams-worker` | OpenClaw runtime image |
| `worker.defaultImage.openclaw.tag` | `string` | `""` | Image tag (defaults to `global.imageTag`) |
| `worker.defaultImage.copaw.repository` | `string` | `higress-registry.cn-hangzhou.cr.aliyuncs.com/agentteams/agentteams-copaw-worker` | Copaw runtime image |
| `worker.defaultImage.copaw.tag` | `string` | `""` | Image tag (defaults to `global.imageTag`) |
| `worker.defaultImage.hermes.repository` | `string` | `higress-registry.cn-hangzhou.cr.aliyuncs.com/agentteams/agentteams-hermes-worker` | Hermes runtime image |
| `worker.defaultImage.hermes.tag` | `string` | `""` | Image tag (defaults to `global.imageTag`) |
| `worker.defaultImage.openhuman.repository` | `string` | `higress-registry.cn-hangzhou.cr.aliyuncs.com/agentteams/agentteams-openhuman-worker` | OpenHuman runtime image |
| `worker.defaultImage.openhuman.tag` | `string` | `""` | Image tag (defaults to `global.imageTag`) |
| `worker.defaultImage.qwenpaw.repository` | `string` | `higress-registry.cn-hangzhou.cr.aliyuncs.com/agentteams/agentteams-qwenpaw-worker` | QwenPaw runtime image |
| `worker.defaultImage.qwenpaw.tag` | `string` | `""` | Image tag (defaults to `global.imageTag`) |
| `worker.defaultRuntime` | `string` | `"openclaw"` | Default runtime when Worker CR omits `spec.runtime` |
| `worker.resources` | `object` | See below | Default resource requests and limits for workers |

**Runtime Image Mapping:**

| Runtime | Default Image Repository | Description |
|---------|--------------------------|-------------|
| `openclaw` | `agentteams-worker` | Primary OpenClaw-based agent runtime |
| `copaw` | `agentteams-copaw-worker` | Copaw agent runtime |
| `hermes` | `agentteams-hermes-worker` | Hermes agent runtime |
| `openhuman` | `agentteams-openhuman-worker` | OpenHuman agent runtime |
| `qwenpaw` | `agentteams-qwenpaw-worker` | QwenPaw agent runtime |

## Customization Patterns

### Local Development (kind/minikube)

```yaml
# values-local.yaml
gateway:
  publicURL: "http://localhost:18080"
  higress:
    enabled: true

higress:
  global:
    local: true
  higress-core:
    gateway:
      service:
        type: NodePort
```

### Alibaba Cloud ACK/ACS Deployment

```yaml
# values-aliyun.yaml
gateway:
  provider: ai-gateway
  mode: existing
  publicURL: "https://api.example.com"
  aiGateway:
    region: "cn-hangzhou"
    gatewayId: "gw-12345"
    modelApiId: "api-67890"
    envId: "env-abcde"

storage:
  provider: oss
  mode: existing
  bucket: "my-agentteams-bucket"
  oss:
    region: "cn-hangzhou"

credentialProvider:
  enabled: true
  image:
    repository: "registry.example.com/agentteams/credential-provider"
    tag: "v1.0.0"
```

### Custom Worker Image

```yaml
worker:
  defaultImage:
    openclaw:
      repository: "my-registry.com/custom-agent"
      tag: "v2.0.0"
```

### Disabling Optional Components

```yaml
# Minimal deployment
elementWeb:
  enabled: false

dashboard:
  enabled: false

matrix:
  appservice:
    enabled: false
```

### High-Availability Controller

```yaml
controller:
  replicaCount: 3
  resources:
    requests:
      cpu: 500m
      memory: 512Mi
    limits:
      cpu: "1"
      memory: 1Gi
```

## Resource Defaults

### Controller Resources

```yaml
controller.resources:
  requests:
    cpu: 100m
    memory: 256Mi
  limits:
    cpu: 500m
    memory: 512Mi
```

### Manager Resources

```yaml
manager.resources:
  requests:
    cpu: 500m
    memory: 1Gi
  limits:
    cpu: "2"
    memory: 4Gi
```

### Worker Resources

```yaml
worker.resources:
  requests:
    cpu: 100m
    memory: 256Mi
  limits:
    cpu: "2"
    memory: 2Gi
```

### Tuwunel Resources

```yaml
matrix.tuwunel.resources:
  requests:
    cpu: 100m
    memory: 256Mi
  limits:
    cpu: 500m
    memory: 512Mi
```

### MinIO Resources

```yaml
storage.minio.resources:
  requests:
    cpu: 250m
    memory: 512Mi
  limits:
    cpu: 500m
    memory: 1Gi
```

### Credential Provider Resources

```yaml
credentialProvider.resources:
  requests:
    cpu: 50m
    memory: 64Mi
  limits:
    cpu: 200m
    memory: 128Mi
```

### Dashboard Resources

```yaml
dashboard.resources:
  requests:
    cpu: 50m
    memory: 64Mi
  limits:
    cpu: 250m
    memory: 256Mi
```

## Installation Commands

For quick-start installation, see [Install & Deploy](./installation.md). Basic Helm commands:

```bash
# Install with default values
helm install agentteams ./helm/agentteams -n agentteams --create-namespace

# Install with custom values
helm install agentteams ./helm/agentteams -n agentteams -f values-custom.yaml

# Upgrade
helm upgrade agentteams ./helm/agentteams -n agentteams -f values-custom.yaml

# Uninstall (runs cleanup hook automatically)
helm uninstall agentteams -n agentteams
```

## Related Pages

- [Install & Deploy](./installation.md) — Quick-start installation commands and local Docker setup
- [Architecture Overview](../architecture/overview.md) — High-level system architecture
- [CRDs and Reconcilers](../controller/crds-and-reconcilers.md) — CRD definitions and controller logic
- [Higress, Matrix, MinIO Integration](../integrations/higress-matrix-minio.md) — Infrastructure component details
