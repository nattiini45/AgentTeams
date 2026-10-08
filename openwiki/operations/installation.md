---
type: Reference
title: "Operations: Install & Deploy"
description: "Docker/Podman local install, Kubernetes Helm install, upgrade, uninstall, and build-from-source procedures for AgentTeams."
tags: [operations, installation, deployment, docker, podman, helm, kubernetes, upgrade, build]
verified:
  - by: openwiki/0.7.1
    at: 2026-10-08T08:19:13.018Z
sources:
  - id: openwiki-source-dd071132c2f003dbe8cc41ab
    resource: repo://docs/embedded-docker-layout.md
  - id: openwiki-source-47c1eb8cedca9425dc039eae
    resource: repo://helm/agentteams/Chart.yaml
  - id: openwiki-source-124a12417423aecf5b9b4da7
    resource: repo://helm/agentteams/values.yaml
  - id: openwiki-source-067453d36925cd87a6260d7a
    resource: repo://install/agentteams-install.sh
  - id: openwiki-source-f7f2756ee27b954d149600e7
    resource: repo://install/defaults.env
generated: { by: "openwiki/0.7.1", at: "2026-10-08T08:19:13.018Z" }
---

# Operations: Install & Deploy

AgentTeams supports two deployment modes: **local Docker/Podman install** and **Kubernetes via Helm**. Both modes use the same container images but differ in how infrastructure components are orchestrated. For post-install monitoring and web UI access, see [Operations: Dashboard](./dashboard.md).

## Local Install (Docker/Podman)

### Requirements

- **Docker Desktop** (Windows/macOS) or **Docker Engine** (Linux) — must be installed and running
- **PowerShell 7+** (Windows only)

### Quick Start

**macOS / Linux:**
```bash
bash <(curl -fsSL https://higress.ai/agentteams/install.sh)
```

**Windows (PowerShell 7+):**
```powershell
Set-ExecutionPolicy Bypass -Scope Process -Force
$wc = New-Object Net.WebClient
$wc.Encoding = [Text.Encoding]::UTF8
iex $wc.DownloadString('https://higress.ai/agentteams/install.ps1')
```

The installer offers two onboarding modes:

1. **Quick Start** — Fast installation with all default values (recommended). Just provide your API key.
2. **Manual** — Customize LLM provider, admin credentials, port configuration, domain names, and data persistence options step by step.

### What Gets Installed

Since **v1.1.0**, the local install uses the **embedded stack** architecture. A single `agentteams-embedded` container runs Higress, Tuwunel, MinIO, Element Web, and the controller under a `supervisord` process. The Manager is a separate slim container that connects to infrastructure over the Docker network.

| Component | Container | Host Port (default) | Internal Port |
|-----------|-----------|---------------------|---------------|
| **Higress** (AI gateway) | `agentteams-embedded` | 18080 (HTTP), 18001 (console) | 8080, 8001 |
| **Tuwunel** (Matrix homeserver) | `agentteams-embedded` | — | 6167 |
| **MinIO** (object storage) | `agentteams-embedded` | — | 9000 |
| **Element Web** (browser Matrix client) | `agentteams-embedded` | 18088 | 8088 |
| **Controller** (Go operator) | `agentteams-embedded` | — | 8090 |
| **Manager** (coordinator agent) | `agentteams-manager` | 18888 (console) | 18888 |
| **Workers** (task executors) | `agentteams-worker`, etc. | — | — |

Workers are created on demand by the Manager through Docker socket. Each Worker is a stateless, disposable container; all persistent state lives in MinIO object storage and the Tuwunel Matrix homeserver.

### Access

Open `http://127.0.0.1:18088` in your browser to access Element Web. The Manager will greet you and explain how to create your first Worker. The Manager console is available at `http://127.0.0.1:18888`.

### Install Scripts

| Script | Purpose |
|--------|---------|
| [`install/agentteams-install.sh`](../../install/agentteams-install.sh) | Main installer (macOS/Linux) |
| [`install/agentteams-install.ps1`](../../install/agentteams-install.ps1) | Main installer (Windows) |
| [`install/agentteams-verify.sh`](../../install/agentteams-verify.sh) | Installation verification |
| [`install/defaults.env`](../../install/defaults.env) | Default configuration values |
| [`install/load-defaults.ps1`](../../install/load-defaults.ps1) | PowerShell defaults loader |

Shared defaults (ports, image names, version gates) are defined in [`install/defaults.env`](../../install/defaults.env). Both install scripts source this file; override any value via environment variable before running the installer.

### Environment Variables

The installers accept environment variables for automation:

| Variable | Default | Description |
|----------|---------|-------------|
| `AGENTTEAMS_NON_INTERACTIVE` | `0` | Skip all prompts, use defaults |
| `AGENTTEAMS_LLM_PROVIDER` | `openai-compat` (en) / `qwen` (zh Token Plan) | LLM provider name |
| `AGENTTEAMS_DEFAULT_MODEL` | `qwen3.6-plus` | Default LLM model |
| `AGENTTEAMS_OPENAI_BASE_URL` | *(empty)* | OpenAI-compatible base URL |
| `AGENTTEAMS_LLM_API_KEY` | *(required)* | LLM API key |
| `AGENTTEAMS_ADMIN_USER` | `admin` | Admin username |
| `AGENTTEAMS_ADMIN_PASSWORD` | *(auto-generated)* | Admin password (min 8 chars) |
| `AGENTTEAMS_MATRIX_DOMAIN` | `matrix-local.agentteams.io:18080` | Matrix domain |
| `AGENTTEAMS_MOUNT_SOCKET` | `1` | Mount container runtime socket |
| `AGENTTEAMS_DATA_DIR` | `agentteams-data` | Docker volume name for persistent data |
| `AGENTTEAMS_WORKSPACE_DIR` | `~/agentteams-manager` | Host directory for manager workspace |
| `AGENTTEAMS_VERSION` | `latest` | Image tag |
| `AGENTTEAMS_REGISTRY` | *(auto-detected by timezone)* | Image registry |
| `AGENTTEAMS_PORT_GATEWAY` | `18080` | Host port for Higress gateway |
| `AGENTTEAMS_PORT_CONSOLE` | `18001` | Host port for Higress console |
| `AGENTTEAMS_PORT_ELEMENT_WEB` | `18088` | Host port for Element Web |
| `AGENTTEAMS_PORT_MANAGER_CONSOLE` | `18888` | Host port for Manager console |
| `AGENTTEAMS_WORKER_IDLE_TIMEOUT` | `720` | Worker idle timeout in minutes (12 hours) |

### Upgrade

The installer performs a **keep-all upgrade** by default: it preserves all existing data (Docker volumes, workspace files, and configuration) while pulling updated images.

```bash
# Upgrade to latest (preserves all data)
bash <(curl -fsSL https://higress.ai/agentteams/install.sh)

# Upgrade to specific version
AGENTTEAMS_VERSION=v1.1.2 bash <(curl -fsSL https://higress.ai/agentteams/install.sh)
```

On Windows:
```powershell
Set-ExecutionPolicy Bypass -Scope Process -Force
$wc = New-Object Net.WebClient
$wc.Encoding = [Text.Encoding]::UTF8
iex $wc.DownloadString('https://higress.ai/agentteams/install.ps1')
```

### Uninstall

```bash
# macOS / Linux
bash <(curl -fsSL https://raw.githubusercontent.com/agentscope-ai/AgentTeams/main/install/agentteams-install.sh) uninstall

# Windows
Set-ExecutionPolicy Bypass -Scope Process -Force
$wc = New-Object Net.WebClient
$wc.Encoding = [Text.Encoding]::UTF8
$s = $wc.DownloadString('https://raw.githubusercontent.com/agentscope-ai/AgentTeams/main/install/agentteams-install.ps1')
& ([scriptblock]::Create($s)) uninstall
```

## Kubernetes Install (Helm)

### Prerequisites

- Kubernetes 1.24+ (kind / minikube / k3s / managed K8s)
- Helm 3.7+
- Default StorageClass (for Tuwunel + MinIO PVCs)

### Quick Install

```bash
helm repo add higress.io https://higress.io/helm-charts
helm repo update

helm install agt higress.io/agentteams \
  -n agentteams-system --create-namespace \
  --render-subchart-notes \
  --set credentials.llmApiKey=<your-api-key> \
  --set credentials.adminPassword=<your-admin-password> \
  --set gateway.publicURL=http://localhost:18080
```

### Key Helm Values

| Value | Required | Default | Description |
|-------|----------|---------|-------------|
| `credentials.llmApiKey` | yes | *(empty)* | LLM provider API key |
| `gateway.publicURL` | yes | *(empty)* | Public URL for Element Web access |
| `credentials.adminPassword` | recommended | *(auto-generated)* | Matrix admin password (auto-generated if empty) |
| `credentials.llmProvider` | no | `openai-compat` | LLM provider name |
| `credentials.defaultModel` | no | `gpt-5.4` | Default LLM model |
| `credentials.llmBaseUrl` | no | *(empty)* | OpenAI-compatible base URL |
| `manager.runtime` | no | `openclaw` | Manager runtime: `openclaw`, `copaw`, `hermes` |
| `manager.enabled` | no | `true` | Create Manager CR during cluster initialization |
| `worker.defaultRuntime` | no | `openclaw` | Default worker runtime: `openclaw`, `copaw`, `hermes`, `openhuman`, `qwenpaw` |
| `preflight.llm.enabled` | no | `true` | Run LLM connectivity preflight check |
| `preflight.llm.strict` | no | `true` | Fail install on preflight failure |
| `matrix.provider` | no | `tuwunel` | Matrix homeserver: `tuwunel` or `synapse` |
| `matrix.mode` | no | `managed` | `managed` (chart deploys) or `existing` (external) |
| `gateway.provider` | no | `higress` | Gateway: `higress` (subchart) or `ai-gateway` (external APIG) |
| `storage.provider` | no | `minio` | Object storage: `minio` (StatefulSet) or `oss` (Alibaba Cloud OSS) |
| `storage.bucket` | no | `agentteams-storage` | Canonical bucket name |
| `elementWeb.enabled` | no | `true` | Deploy Element Web UI |
| `dashboard.enabled` | no | `false` | Deploy operator dashboard |
| `controller.uninstallHook.enabled` | no | `true` | Auto-cleanup CRs on `helm uninstall` |

### Helm Chart Structure

The chart is at [`helm/agentteams/`](../../helm/agentteams/):

```
helm/agentteams/
├── Chart.yaml              # Chart metadata (v1.1.1)
├── values.yaml             # Default values
├── crds/                   # CRD manifests
│   ├── workers.agentteams.io.yaml
│   ├── managers.agentteams.io.yaml
│   ├── teams.agentteams.io.yaml
│   ├── projects.agentteams.io.yaml
│   └── humans.agentteams.io.yaml
├── templates/
│   ├── _helpers.tpl        # Template helpers
│   ├── controller/         # Controller deployment
│   ├── manager/            # Manager CR (optional)
│   ├── matrix/             # Tuwunel StatefulSet
│   ├── minio/              # MinIO StatefulSet
│   ├── gateway/            # Higress configuration
│   ├── element/            # Element Web deployment
│   └── dashboard/          # Operator dashboard (optional)
└── charts/                 # Subcharts (Higress)
```

The chart depends on the **Higress** subchart (`higress` v2.2.1) for gateway deployment when `gateway.higress.enabled=true`.

### Upgrading (Helm)

```bash
# Upgrade to latest chart version
helm repo update
helm upgrade agt higress.io/agentteams \
  -n agentteams-system \
  --render-subchart-notes

# Upgrade with specific values
helm upgrade agt higress.io/agentteams \
  -n agentteams-system \
  --set credentials.defaultModel=gpt-4 \
  --render-subchart-notes
```

The Helm chart includes a **preflight LLM check** (`preflight.llm.enabled=true`) that validates LLM connectivity during install and upgrade. When `preflight.llm.strict=true` (default), the operation fails if the check cannot reach the LLM endpoint.

### Uninstalling (Helm)

```bash
helm uninstall agt -n agentteams-system
```

When `controller.uninstallHook.enabled=true` (default), `helm uninstall` first runs a cleanup Job that deletes all Manager/Worker/Team/Human CRs while the controller is still alive, so the controller's finalizer logic can clean up Pods, Matrix users, and storage data.

## Building from Source

### Image Dependency Chain

```
openclaw-base → manager / worker
                copaw-worker (separate)
                hermes-worker (separate)
                openhuman-worker (separate)
                qwenpaw-worker (separate)
                controller (separate)
                embedded (controller + infra)
```

### Build Commands

```bash
# Build all images
make build

# Build specific images
make build-manager
make build-worker
make build-copaw-worker
make build-hermes-worker
make build-openhuman-worker
make build-qwenpaw-worker
make build-controller
make build-embedded
make build-openclaw-base

# Build using local base image (important!)
make build-manager build-worker \
    OPENCLAW_BASE_IMAGE=agt/openclaw-base \
    OPENCLAW_BASE_VERSION=latest
```

**Common pitfall:** Running `make build-manager` without `OPENCLAW_BASE_IMAGE=agt/openclaw-base` will pull the remote registry's image instead of using your locally-built base. Always set both variables together.

### Registry Configuration

Default registry: `higress-registry.cn-hangzhou.cr.aliyuncs.com/agentteams/`

Regional mirrors:
- **China:** `higress-registry.cn-hangzhou.cr.aliyuncs.com`
- **North America:** `higress-registry.us-west-1.cr.aliyuncs.com`
- **Southeast Asia:** `higress-registry.ap-southeast-7.cr.aliyuncs.com`

### Multi-Architecture Builds

```bash
# Build and push multi-arch (amd64 + arm64)
make push

# Push native-arch only (dev use)
make push-native
```

### Proxy Support

```bash
PROXY_ARGS="--build-arg HTTP_PROXY=http://host.containers.internal:1087 \
    --build-arg HTTPS_PROXY=http://host.containers.internal:1087"

make build-embedded build-manager build-worker DOCKER_BUILD_ARGS="${PROXY_ARGS}"
```

### China Build Acceleration

```bash
# APT mirror
make build-embedded DOCKER_BUILD_ARGS="--build-arg APT_MIRROR=mirrors.aliyun.com"

# PIP mirror (Python images)
make build-copaw-worker DOCKER_BUILD_ARGS="--build-arg PIP_INDEX_URL=https://mirrors.aliyun.com/pypi/simple/"

# NPM mirror (Node.js images)
make build-openclaw-base DOCKER_BUILD_ARGS="--build-arg NPM_REGISTRY=https://registry.npmmirror.com/"
```

## Declarative Resource Management

Workers, Teams, and Humans can be created from YAML files:

```yaml
# worker.yaml
apiVersion: agentteams.io/v1beta1
kind: Worker
metadata:
  name: my-worker
spec:
  runtime: openclaw
  model: gpt-4
  workerName: "My Worker"
  soul: "You are a helpful assistant."
  skills:
    - name: github-operations
```

```bash
agt apply -f worker.yaml
```

See [`docs/declarative-resource-management.md`](../../docs/declarative-resource-management.md) for full reference.

## Source References

- Install scripts: [`install/`](../../install/)
- Install README: [`install/README.md`](../../install/README.md)
- Embedded Docker layout: [`docs/embedded-docker-layout.md`](../../docs/embedded-docker-layout.md)
- Helm chart: [`helm/agentteams/`](../../helm/agentteams/)
- Makefile: [`Makefile`](../../Makefile)
- Development guide: [Development: Build & Test](../development/build-and-test.md)
- Dashboard: [Operations: Dashboard](./dashboard.md)
- Declarative management: [`docs/declarative-resource-management.md`](../../docs/declarative-resource-management.md)
