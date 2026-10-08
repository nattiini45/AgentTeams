---
type: Reference
title: "Development: Build & Test"
description: Comprehensive guide to building, testing, and contributing to the AgentTeams project, including Makefile targets, CI/CD workflows, and development workflows.
tags: [development, build, test, ci-cd, makefile, docker, integration-tests, contribution]
verified:
  - by: openwiki/0.7.1
    at: 2026-10-08T08:19:13.018Z
sources:
  - id: openwiki-source-7a80b79a6fb3618cbfab08a2
    resource: repo://.github/workflows/build.yml
  - id: openwiki-source-9227e4899cd431c13fea0870
    resource: repo://.github/workflows/check-crd-sync.yml
  - id: openwiki-source-df90e26795bfe5de5cfe8e8f
    resource: repo://.github/workflows/helm-lint.yml
  - id: openwiki-source-6d4b4e707b8d60b6ccfa3425
    resource: repo://.github/workflows/openwiki-update.yml
  - id: openwiki-source-b85472c728afdd14c361dede
    resource: repo://.github/workflows/publish-copaw.yml
  - id: openwiki-source-4d1d392666be6dfdd7a91a2e
    resource: repo://.github/workflows/release.yml
  - id: openwiki-source-91097a2889d8ad93fdc4a691
    resource: repo://.github/workflows/remediation-gates.yml
  - id: openwiki-source-bb3b3c9f8b452d38353806f1
    resource: repo://.github/workflows/test-controller.yml
  - id: openwiki-source-072dbd7a5907d6c9101f8708
    resource: repo://.github/workflows/test-integration.yml
  - id: openwiki-source-4d1645cb6317345817452838
    resource: repo://.pre-commit-config.yaml
  - id: openwiki-source-82390634420bd9d77c608992
    resource: repo://agentteams-controller/Makefile
  - id: openwiki-source-16746e90993dfbd40cd0c46f
    resource: repo://changelog/current.md
  - id: openwiki-source-012f2c78e3b1446dfc35803f
    resource: repo://Makefile
  - id: openwiki-source-efad135382276c5ba1f0b093
    resource: repo://shared/python/agentteams_protocol/pyproject.toml
  - id: openwiki-source-614bc4d1f026a508ac3ccf53
    resource: repo://shared/python/agentteams_sync/pyproject.toml
  - id: openwiki-source-cc18f7c982b54a9643a61d74
    resource: repo://tests/README.md
  - id: openwiki-source-dce9e028b3aa5738d6aba3e3
    resource: repo://tests/run-all-tests.sh
generated: { by: "openwiki/0.7.1", at: "2026-10-08T08:19:13.018Z" }
---

# Development: Build & Test

This page covers the development workflow for AgentTeams: building images, running tests, CI/CD, and contribution guidelines.

## Makefile

The [`Makefile`](../../Makefile) is the unified build/test/push/install interface. Run `make help` for all targets.

### Build Targets

| Target | Description |
|--------|-------------|
| `make build` | Build all images (native arch, local) |
| `make build-openclaw-base` | Build base image (Ubuntu + Node.js + OpenClaw) |
| `make build-manager` | Build OpenClaw Manager image |
| `make build-manager-copaw` | Build CoPaw Manager image |
| `make build-worker` | Build OpenClaw Worker image |
| `make build-copaw-worker` | Build CoPaw Worker image |
| `make build-hermes-worker` | Build Hermes Worker image |
| `make build-openhuman-worker` | Build OpenHuman Worker image |
| `make build-qwenpaw-worker` | Build QwenPaw Worker image |
| `make build-agentteams-controller` | Build controller image |
| `make build-embedded` | Build embedded (all-in-one) image |

### Test Targets

| Target | Description |
|--------|-------------|
| `make test` | Build + run all integration tests |
| `make test SKIP_BUILD=1` | Run tests without rebuilding |
| `make test TEST_FILTER="01 02"` | Run specific tests |
| `make test-quick` | Run test-01 only (quick smoke test) |
| `make test-installed` | Run tests against an already-installed Manager |
| `make test-embedded` | Run integration tests in embedded mode |
| `make test-embedded SKIP_INSTALL=1` | Run tests against existing embedded installation |
| `make helm-lint` | Lint Helm chart |
| `make helm-template` | Render Helm templates |

### Push Targets

| Target | Description |
|--------|-------------|
| `make push` | Build + push multi-arch images (amd64 + arm64) |
| `make push-native` | Push native-arch images only (dev use) |
| `make push-openclaw-base` | Build + push multi-arch OpenClaw base image |
| `make push-agentteams-controller` | Build + push multi-arch agentteams-controller image |
| `make push-embedded` | Build + push multi-arch embedded image |
| `make push-manager` | Build + push multi-arch Manager image |
| `make push-manager-copaw` | Build + push multi-arch CoPaw Manager image |
| `make push-worker` | Build + push multi-arch Worker image |
| `make push-copaw-worker` | Build + push multi-arch CoPaw Worker image |
| `make push-hermes-worker` | Build + push multi-arch Hermes Worker image |
| `make push-openhuman-worker` | Build + push multi-arch OpenHuman Worker image |
| `make push-qwenpaw-worker` | Build + push multi-arch QwenPaw Worker image |

### Utility Targets

| Target | Description |
|--------|-------------|
| `make clean` | Remove local images and test containers |
| `make status` | Show status of Manager and Worker containers |
| `make logs` | Show recent logs (LINES=N to customize) |
| `make tag` | Tag images for registry push |
| `make buildx-setup` | Ensure multi-arch build prerequisites are met |
| `make mirror-images` | Mirror upstream images to Higress registry |
| `make local-k8s-up` | Create kind cluster and deploy AgentTeams via Helm |
| `make local-k8s-down` | Tear down the local AgentTeams kind cluster |
| `make generate` | Regenerate deepcopy functions and sync CRDs to Helm chart |
| `make sync-crds` | Sync CRDs from agentteams-controller/config/crd/ to helm/agentteams/crds/ |
| `make check-crd-sync` | Verify CRDs are in sync between controller and Helm chart |
| `make verify` | Run post-install verification against the running Manager container |
| `make replay` | Send a task to Manager (TASK="..." or interactive) |
| `make replay-log` | View the latest replay conversation log |

## Project Structure for Developers

### Controller Development (Go)

```
agentteams-controller/
├── api/v1beta1/           # CRD type definitions
├── cmd/agt/            # CLI entry point and commands
├── internal/
│   ├── controller/        # Reconciler implementations
│   ├── service/           # Provisioner and deployer logic
│   ├── gateway/           # Higress/AIGateway clients
│   ├── matrix/            # Matrix homeserver client
│   ├── server/            # REST API handlers
│   ├── backend/           # Container backend abstraction
│   ├── config/            # Configuration loading
│   ├── agentconfig/       # Agent config generation
│   ├── metrics/           # Prometheus metrics
│   └── managerstate/      # Task board state
├── config/crd/            # Generated CRD YAML manifests
├── hack/                  # Maintenance scripts
└── Dockerfile.embedded    # Embedded image build
```

**Key entry points:**
- [`agentteams-controller/cmd/agt/main.go`](../../agentteams-controller/cmd/agt/main.go) — CLI entry
- [`agentteams-controller/internal/app/app.go`](../../agentteams-controller/internal/app/app.go) — Application setup
<!-- openwiki: broken internal link [../../agentteams-controller/internal/AGENTS.md] file "../../agentteams-controller/internal/AGENTS.md" does not exist. Fix the href or restore the target, then delete this comment. -->
- [`agentteams-controller/internal/AGENTS.md`](../../agentteams-controller/internal/AGENTS.md) — Package routing map

### Worker Development (Python)

CoPaw, Hermes, and QwenPaw are Python packages with standard layouts:

```
copaw/
├── pyproject.toml         # Package definition
├── Dockerfile             # Image build
├── src/copaw_worker/      # Source package
└── tests/                 # Test suite
```

### Worker Development (Rust)

OpenHuman uses a Rust workspace:

```
openhuman/
├── Cargo.toml             # Workspace definition
├── Dockerfile             # Multi-stage build
├── src/                   # Rust source
└── tests/                 # Test suite
```

### Manager Development

```
manager/
├── Dockerfile             # OpenClaw Manager build
├── Dockerfile.copaw       # CoPaw Manager build
├── agent/                 # Agent-facing content (skills, prompts, templates)
├── configs/               # Configuration templates
├── scripts/               # Bootstrap and init scripts
└── tests/                 # Integration tests
```

### Shared Python Libraries

```
shared/python/
├── agentteams_matrix_format/   # Matrix message formatting
├── agentteams_matrix_policies/ # Matrix policy enforcement
├── agentteams_openclaw_merge/  # OpenClaw merge utilities
├── agentteams_protocol/        # Protocol definitions
└── agentteams_sync/            # File synchronization
```

Each shared Python package has its own test suite and can be installed with:

```bash
# Install all shared Python packages
for pkg in shared/python/agentteams_*; do
  [ -f "${pkg}/pyproject.toml" ] || continue
  python -m pip install -e "./${pkg}"
done
```

## Testing

### Integration Tests

Integration tests live in [`tests/`](../../tests/) and run against a full embedded stack. They verify:
- Container startup and configuration
- Manager-Worker communication via Matrix
- Task delegation and completion
- File sync with MinIO
- Gateway routing
- GitHub MCP operations
- Multi-worker collaboration
- Worker configuration management
- Team project DAG execution
- Skills management

### Test Cases

| Test | Description |
|------|-------------|
| test-01 | Manager boot, all services healthy, IM login |
| test-02 | Create Worker Alice via Matrix conversation |
| test-03 | Assign task, Worker completes |
| test-04 | Human intervenes with supplementary instructions |
| test-05 | Heartbeat triggers Manager inquiry |
| test-06 | Create Bob, collaborative task |
| test-08 | GitHub operations via MCP Server |
| test-09 | Multi-Worker GitHub collaboration |
| test-10 | MCP permission dynamic revoke/restore |
| test-11 | Multi-round GitHub PR collaboration |
| test-12 | GitHub MCP tools integration |
| test-13 | Git delegation workflow |
| test-14 | Git collaboration workflow |
| test-15 | Import worker from ZIP archive |
| test-16 | Import worker package default |
| test-17 | Worker configuration verification |
| test-18 | Team configuration verification |
| test-19 | Human and team admin operations |
| test-20 | Inline worker configuration |
| test-21 | Team project DAG execution |
| test-22 | Delete worker cleanup |
| test-23 | Runtime switch between OpenClaw and CoPaw |
| test-24 | Skills management |
| test-25 | Name validation |
| test-26 | QwenPaw teamharness plugin mode |
| test-100 | Cleanup and final verification |

### Running Tests

#### Via Makefile (Recommended)

```bash
# Full test flow (auto-creates and cleans up test container)
AGENTTEAMS_LLM_API_KEY=sk-xxx make test

# Skip image rebuild
make test SKIP_BUILD=1

# Run specific tests
make test TEST_FILTER="01 02"

# Test an existing Manager installation
make test SKIP_INSTALL=1

# Quick smoke test (test-01 only)
make test-quick

# Test embedded mode
make test-embedded

# Test embedded mode against existing installation
make test-embedded SKIP_INSTALL=1
```

#### Direct Script Execution

```bash
# Build + run all tests
./tests/run-all-tests.sh

# Use existing images
./tests/run-all-tests.sh --skip-build

# Run specific tests only
./tests/run-all-tests.sh --test-filter "01 02 03"

# Run against an already-installed Manager
./tests/run-all-tests.sh --use-existing

# Use a custom container name
./tests/run-all-tests.sh --container my-test-container
```

### Python Package Tests

Each shared Python package has its own test suite:

```bash
# Run all Python tests (via remediation-gates CI)
# Locally: install and run each package
cd shared/python/agentteams_protocol && python -m pytest
cd shared/python/agentteams_sync && python -m pytest
cd copaw && python -m pytest
cd hermes && python -m pytest
cd qwenpaw && python -m pytest
```

### Controller Tests

```bash
# Run Go unit tests
cd agentteams-controller && make test-unit

# Run Go integration tests (envtest)
cd agentteams-controller && make test-integration

# Run all Go tests
cd agentteams-controller && make test-all
```

### Helm Chart Validation

```bash
# Lint
make helm-lint

# Render templates
make helm-template

# Test rendered output
make helm-template | kubectl apply --dry-run=client -f -
```

## CI/CD

GitHub Actions workflows in [`.github/workflows/`](../../.github/workflows/):

### Core Workflows

| Workflow | Purpose |
|----------|---------|
| `test-integration.yml` | Full integration test suite with sharded parallel execution |
| `test-controller.yml` | Go unit and integration tests for controller |
| `remediation-gates.yml` | Security and quality gates for all runtimes |
| `build.yml` | Build and push multi-arch images on tag push |
| `build-rc.yml` | Build and push RC (release candidate) images |
| `build-base.yml` | Build and push OpenClaw base image |
| `release.yml` | Full release workflow with changelog and GitHub release |
| `helm-lint.yml` | Helm chart linting and template validation |
| `check-crd-sync.yml` | Verify CRDs are in sync between controller and Helm chart |
| `deploy-helm-to-oss.yml` | Deploy Helm chart to Alibaba Cloud OSS |
| `openwiki-update.yml` | Automated OpenWiki documentation refresh |
| `publish-copaw.yml` | Publish CoPaw Worker to PyPI |
| `translate.yml` | Translate GitHub content into English |
| `docs-links.yml` | Verify AGENTS.md relative links resolve |

### Integration Test Shards

The integration test suite is sharded for parallel execution:

- **Shard A**: LLM interaction tests (01-06) - sequential dependency chain
- **Shard B**: LLM interaction tests (14) - task assignment with LLM
- **Shard C**: Controller/CR tests (15, 17-20, 22-25, 100) - independent
- **Shard D**: Controller/CR tests (21) - team project DAG

### Authorization Gate

For security, integration tests run only for:
- Same-repo PRs
- Fork PRs with the `safe-to-test` label
- Push events to main
- Tag pushes
- Manual workflow_dispatch

### Pre-commit Hooks

[`.pre-commit-config.yaml`](../../.pre-commit-config.yaml) defines local hooks. Install with:

```bash
pre-commit install
```

Hooks include:
- Trailing whitespace removal
- End-of-file fixer
- YAML validation
- JSON validation
- Merge conflict detection
- Go formatting (for controller code)

## Changelog Policy

Any change that affects built image content **must** be recorded in [`changelog/current.md`](../../changelog/current.md) before committing.

Format:
```
- feat(manager): add task-management skill ([a1b2c3d](url))
- fix(controller): fix worker reconciliation loop ([e4f5g6h](url))
```

On release, the workflow renames `current.md` → `vX.Y.Z.md` and creates a fresh `current.md`.

## Key Design Patterns for Contributors

1. **All communication in Matrix rooms** — Human + Manager + Worker are all in the same room
2. **Centralized file system** — All agent configs and state stored in MinIO
3. **Unified credential management** — Workers use consumer tokens only
4. **Skills as documentation** — Each SKILL.md is self-contained
5. **Agent-facing content uses second-person voice** — "You are the Manager..."

## Common Development Tasks

### Adding a New Worker Runtime

1. Create runtime directory (e.g., `myruntime/`)
2. Implement Matrix connection, gateway auth, MinIO sync
3. Add Dockerfile
4. Create agent template in `manager/agent/myruntime-worker-agent/`
5. Add runtime option to controller CRD types
6. Update Helm chart defaults
7. Add Makefile targets

### Adding a Manager Skill

1. Create `manager/agent/skills/my-skill/SKILL.md`
2. Add optional `scripts/` and `references/`
3. Write SKILL.md in second-person voice
4. Test with a running Manager instance

### Modifying CRDs

1. Edit types in `agentteams-controller/api/v1beta1/`
2. Run `make generate` to regenerate deepcopy
3. Run `make manifests` to regenerate CRD YAML
4. Update Helm chart CRDs in `helm/agentteams/crds/`
5. Update reconcilers if needed

## Source References

- Makefile: [`Makefile`](../../Makefile)
- Controller Makefile: [`agentteams-controller/Makefile`](../../agentteams-controller/Makefile)
- Development guide: [`docs/development.md`](../../docs/development.md)
- CI workflows: [`.github/workflows/`](../../.github/workflows/)
- Pre-commit: [`.pre-commit-config.yaml`](../../.pre-commit-config.yaml)
- Changelog: [`changelog/current.md`](../../changelog/current.md)
- Integration tests: [`tests/`](../../tests/)
- Shared Python libraries: [`shared/python/`](../../shared/python/)
