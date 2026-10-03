---
type: "Reference"
title: "Development: Build & Test"
description: "Developer workflow: Makefile targets, project structure for each language (Go, Python, Rust, Shell), testing strategy (integration, Python package, Go tests, Helm validation), CI/CD workflows, pre-commit hooks, changelog policy, and common development tasks."
tags: ["development", "build", "test", "ci-cd", "makefile", "integration-tests", "pre-commit"]
verified:
  - by: openwiki/0.7.0
    at: 2026-10-03T08:15:26.963Z
sources:
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
  - id: openwiki-source-e2c41317a1ead02a54ef6b91
    resource: repo://shared/python/AGENTS.md
  - id: openwiki-source-cc18f7c982b54a9643a61d74
    resource: repo://tests/README.md
  - id: openwiki-source-dce9e028b3aa5738d6aba3e3
    resource: repo://tests/run-all-tests.sh
generated: { by: "openwiki/0.7.0", at: "2026-10-03T08:15:26.963Z" }
---

# Development: Build & Test

This page covers the development workflow for AgentTeams: building images, running tests, CI/CD, and contribution guidelines.

## Makefile Targets

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
| `make build-openhuman-worker` | Build OpenHuman Worker image (Rust + native Matrix) |
| `make build-qwenpaw-worker` | Build QwenPaw Worker image |
| `make build-controller` | Build controller image |
| `make build-embedded` | Build embedded (all-in-one) image |

### Test Targets

| Target | Description |
|--------|-------------|
| `make test` | Build + run all integration tests |
| `make test SKIP_BUILD=1` | Run tests without rebuilding |
| `make test TEST_FILTER="01 02"` | Run specific tests |
| `make test SKIP_INSTALL=1` | Run tests against existing Manager |
| `make test-quick` | Run test-01 only (quick smoke test) |
| `make test-installed` | Run tests against an already-installed Manager (no container lifecycle) |
| `make test-embedded` | Run embedded controller tests |
| `make test-python` | Run Python package tests |
| `make helm-lint` | Lint Helm chart |
| `make helm-template` | Render Helm templates |

### Push Targets

| Target | Description |
|--------|-------------|
| `make push` | Build + push multi-arch images (amd64 + arm64) |
| `make push-native` | Push native-arch images only (dev use) |
| `make push-openclaw-base` | Build + push multi-arch OpenClaw base image |
| `make push-controller` | Build + push multi-arch agentteams-controller image |
| `make push-embedded` | Build + push multi-arch embedded all-in-one image |
| `make push-manager` | Build + push multi-arch Manager image (OpenClaw) |
| `make push-manager-copaw` | Build + push multi-arch Manager CoPaw image |
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
| `make install` | Install Manager locally (non-interactive, set AGENTTEAMS_LLM_API_KEY) |
| `make install-interactive` | Install Manager interactively (prompts for config) |
| `make install-embedded` | Install in embedded mode (dual-container: controller + agent) |
| `make uninstall` | Stop and remove Manager + all Worker containers |
| `make uninstall-embedded` | Stop and remove embedded containers |
| `make verify` | Run post-install verification against the running Manager container |
| `make replay` | Send a task to Manager (TASK="..." or interactive, YOLO mode auto-enabled) |
| `make replay-log` | View the latest replay conversation log |
| `make local-k8s-up` | Create kind cluster and deploy AgentTeams via Helm |
| `make local-k8s-down` | Tear down the local AgentTeams kind cluster |
| `make generate` | Regenerate deepcopy functions and sync CRDs to Helm chart |
| `make sync-crds` | Sync CRDs from agentteams-controller/config/crd/ to helm/agentteams/crds/ |
| `make check-crd-sync` | Verify CRDs are in sync between controller and Helm chart |
| `make mirror-images` | Mirror upstream images to Higress registry (multi-arch, via skopeo) |
| `make buildx-setup` | Ensure multi-arch build prerequisites are met |

### Key Makefile Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `VERSION` | `latest` | Image tag |
| `REGISTRY` | `higress-registry.cn-hangzhou.cr.aliyuncs.com` | Container registry |
| `REPO` | `agentteams` | Repository namespace |
| `HIGRESS_REGISTRY` | `higress-registry.cn-hangzhou.cr.aliyuncs.com` | Base image registry |
| `SKIP_BUILD` | (empty) | Skip build in `install` (set to 1 to skip) |
| `SKIP_INSTALL` | (empty) | Skip install in `test` (set to 1 to test existing) |
| `TEST_FILTER` | (empty) | Test numbers to run (e.g., '01 02 03') |
| `DOCKER_PLATFORM` | (empty) | Build platform (e.g., linux/amd64) |
| `MULTIARCH_PLATFORMS` | `linux/amd64,linux/arm64` | Multi-arch platforms |
| `BUILDX_BUILDER` | `agentteams-multiarch` | Buildx builder name |

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

Five independently-installable Python packages providing domain logic shared across all AgentTeams worker runtimes:

| Package | Purpose | Key modules |
|---------|---------|-------------|
| `agentteams_protocol` | Domain models, task/project DAG validation | `task.py` (1300+ lines), `errors.py` |
| `agentteams_sync` | MinIO file sync, daemon, per-runtime push policies | `filesync.py`, `daemon.py`, `contract.py`, `policy.py`, `openclaw.py` |
| `agentteams_openclaw_merge` | Canonical openclaw.json merge logic | `merge.py`, `__main__.py` (CLI wrapper) |
| `agentteams_matrix_format` | Markdown-it rendering for Matrix messages | `__init__.py` |
| `agentteams_matrix_policies` | Matrix channel allow-list policy builder | `policies.py` |

**Namespace**: all packages use `agentteams_*` prefix  
**Layout**: setuptools src-layout (`src/agentteams_<name>/`)  
**Independence**: each package is independently installable via `pip install -e ./shared/python/agentteams_<name>`  
**Dependency rule**: no cross-package imports except `agentteams_protocol` ← others may import it  
**Versioning**: all start at `>=0.1.0`; no published PyPI releases (installed from source)

### Shell Libraries

Shared shell libraries in `shared/lib/` provide utilities for Docker operations, environment management, and common scripts used across the codebase.

## Testing

### Integration Tests

Integration tests live in [`tests/`](../../tests/) and run against a full embedded stack. They verify:
- Container startup and configuration
- Manager-Worker communication via Matrix
- Task delegation and completion
- File sync with MinIO
- Gateway routing

**Architecture:**

Tests simulate human interaction by calling the Matrix API directly, then verify system responses and side effects:

```
Test Script                     AgentTeams System
    |                               |
    ├── Matrix API: send message ──>| Manager Agent processes
    |                               │ (creates Worker, assigns task, etc.)
    ├── poll Matrix API for reply <─|
    ├── verify reply content        |
    ├── verify Higress Console ────>| (Consumer created? Route updated?)
    ├── verify MinIO files ────────>| (SOUL.md written? task/spec.md?)
    └── PASS / FAIL                 |
```

**Test Cases:**

| Test | POC Case | Description |
|------|----------|-------------|
| test-01 | Case 1 | Manager boot, all services healthy, IM login |
| test-02 | Case 2 | Create Worker Alice via Matrix conversation |
| test-03 | Case 3 | Assign task, Worker completes |
| test-04 | Case 4 | Human intervenes with supplementary instructions |
| test-05 | Case 5 | Heartbeat triggers Manager inquiry |
| test-06 | Case 6 | Create Bob, collaborative task |
| test-07 | Case 7 | Credential smooth rotation (TODO) |
| test-08 | Case 8 | GitHub operations via MCP Server |
| test-09 | Case 9 | Multi-Worker GitHub collaboration |
| test-10 | Case 10 | MCP permission dynamic revoke/restore |
| test-11 | Feature | Multi-round GitHub PR collaboration |
| test-12 | Feature | GitHub MCP tools |
| test-13 | Feature | Git delegation |
| test-14 | Feature | Git collaboration |
| test-15 | Feature | Import worker zip |
| test-16 | Feature | Import worker package default |
| test-17 | Feature | Worker config verify |
| test-18 | Feature | Team config verify |
| test-19 | Feature | Human and team admin |
| test-20 | Feature | Inline worker config |
| test-21 | Feature | Team project DAG |
| test-22 | Feature | Delete worker cleanup |
| test-23 | Feature | Runtime switch |
| test-24 | Feature | Skills management |
| test-25 | Feature | Name validation |
| test-26 | Feature | QwenPaw teamharness plugin mode |
| test-100 | Cleanup | Cleanup test environment |

**Required Environment Variables:**

| Variable | Required | Description |
|----------|----------|-------------|
| `AGENTTEAMS_LLM_API_KEY` | Yes | LLM API key for Agent behavior |
| `AGENTTEAMS_GITHUB_TOKEN` | No | GitHub PAT for tests 08-11 |

**Helper Libraries:**

- `lib/test-helpers.sh`: Assertions, lifecycle, logging, Docker helpers
- `lib/matrix-client.sh`: Matrix API wrapper (register, login, send/read messages)
- `lib/higress-client.sh`: Higress Console API wrapper (consumers, routes, MCP)
- `lib/minio-client.sh`: MinIO verification (file existence, content, listing)

**Running Tests:**

```bash
# Via Makefile (Recommended)
AGENTTEAMS_LLM_API_KEY=sk-xxx make test

# Skip image rebuild
make test SKIP_BUILD=1

# Run specific tests
make test TEST_FILTER="01 02"

# Test an existing Manager installation
make test SKIP_INSTALL=1

# Quick smoke test (test-01 only)
make test-quick
```

### Python Package Tests

Each shared Python package has its own test suite:

```bash
# Run all Python tests
make test-python

# Run specific package tests
cd shared/python/agentteams_protocol && python -m pytest
cd shared/python/agentteams_sync && python -m pytest
cd copaw && python -m pytest
```

**Test strategy for shared packages:**

Changes to shared Python packages trigger rebuilds of dependent runtime images (copaw, hermes, qwenpaw, openhuman, worker, manager-copaw). The `remediation-gates.yml` CI job auto-discovers and tests all `shared/python/agentteams_*` packages on every PR.

### Controller Tests (Go)

```bash
# Run Go tests
cd agentteams-controller && go test ./...

# Run specific test
cd agentteams-controller && go test ./internal/controller/...

# Run integration tests (requires envtest setup)
cd agentteams-controller && make test-integration
```

### Helm Chart Validation

```bash
# Lint
make helm-lint

# Render templates
make helm-template

# Test rendered output
make helm-template | kubectl apply --dry-run=client -f -

# Check CRD sync between controller and Helm chart
make check-crd-sync
```

## CI/CD Workflows

GitHub Actions workflows in [`.github/workflows/`](../../.github/workflows/):

| Workflow | Purpose |
|----------|---------|
| `test-integration.yml` | Integration test suite (builds images, runs all tests in shards) |
| `remediation-gates.yml` | Security and quality gates (Go tests, Python tests, Helm lint) |
| `build.yml` | Build and push multi-arch images |
| `build-base.yml` | Build and push OpenClaw base image |
| `build-rc.yml` | Build release candidate images |
| `release.yml` | Release workflow |
| `helm-lint.yml` | Helm chart linting |
| `check-crd-sync.yml` | Verify CRDs are in sync between controller and Helm chart |
| `test-controller.yml` | Controller-specific tests |
| `openwiki-update.yml` | OpenWiki documentation refresh |
| `docs-links.yml` | Documentation link validation |
| `translate.yml` | Translation workflow |
| `deploy-helm-to-oss.yml` | Deploy Helm chart to OSS |
| `publish-copaw.yml` | Publish CoPaw package |

**Integration Test Shards:**

The integration test workflow uses four parallel shards:

- **Shard A**: LLM interaction tests (sequential dependency: 02 creates alice → 03-06 use alice)
- **Shard B**: LLM interaction tests 2 (task assignment with LLM)
- **Shard C**: Controller/CR tests (independent, no cross-test dependencies)
- **Shard D**: Controller/CR tests 2 (team project DAG)

**Authorization Gate:**

The workflow uses an authorization gate for fork PRs to prevent untrusted code from running with repository secrets. Fork PRs require the `safe-to-test` label from a maintainer.

### Pre-commit Hooks

[`.pre-commit-config.yaml`](../../.pre-commit-config.yaml) defines local hooks. Install with:

```bash
pre-commit install
```

**Available hooks:**

| Hook | Source | Description |
|------|--------|-------------|
| `trailing-whitespace` | pre-commit-hooks | Remove trailing whitespace |
| `end-of-file-fixer` | pre-commit-hooks | Ensure files end with newline |
| `check-yaml` | pre-commit-hooks | Validate YAML syntax |
| `check-json` | pre-commit-hooks | Validate JSON syntax |
| `check-merge-conflict` | pre-commit-hooks | Check for merge conflict markers |
| `go-fmt` | pre-commit-golang | Format Go code |

**Note:** Hooks exclude `manager/agent/` and `.qoder/` directories, and Helm templates from YAML validation.

## Changelog Policy

Any change that affects built image content **must** be recorded in [`changelog/current.md`](../../changelog/current.md) before committing.

**Format:**
```
<!-- openwiki: broken internal link [url] file "url" does not exist. Fix the href or restore the target, then delete this comment. -->
- feat(manager): add task-management skill ([a1b2c3d](url))
<!-- openwiki: broken internal link [url] file "url" does not exist. Fix the href or restore the target, then delete this comment. -->
- fix(controller): fix worker reconciliation loop ([e4f5g6h](url))
```

**What to record:**
- Changes to `manager/`, `worker/`, `copaw/`, `hermes/`, `openclaw-base/`, `agentteams-controller/`
- Release-facing install/chart changes
- New features, bug fixes, and improvements

**Release process:**
On release, the workflow renames `current.md` → `vX.Y.Z.md` and creates a fresh `current.md`.

## Common Development Tasks

### Adding a New Worker Runtime

1. Create runtime directory (e.g., `myruntime/`)
2. Implement Matrix connection, gateway auth, MinIO sync
3. Add Dockerfile
4. Create agent template in `manager/agent/myruntime-worker-agent/`
5. Add runtime option to controller CRD types
6. Update Helm chart defaults
7. Add Makefile targets (`build-myruntime-worker`, `push-myruntime-worker`, `push-native-myruntime-worker`)
8. Update CI workflows to include new runtime in test matrix

### Adding a Manager Skill

1. Create `manager/agent/skills/my-skill/SKILL.md`
2. Add optional `scripts/` and `references/`
3. Write SKILL.md in second-person voice (see [Agent Content](../agent-content/skills-and-prompts.md))
4. Test with a running Manager instance
5. Update changelog if skill affects built images

### Modifying CRDs

1. Edit types in `agentteams-controller/api/v1beta1/`
2. Run `make generate` to regenerate deepcopy
3. Run `make manifests` to regenerate CRD YAML
4. Update Helm chart CRDs in `helm/agentteams/crds/`
5. Update reconcilers if needed
6. Run `make check-crd-sync` to verify synchronization
7. Update tests if CRD changes affect existing behavior

### Adding a New Shared Python Package

1. Create package directory in `shared/python/agentteams_<name>/`
2. Add `pyproject.toml` with `agentteams_*` namespace
3. Implement package in `src/agentteams_<name>/`
4. Add tests in `tests/`
5. Update `shared/python/AGENTS.md` with package documentation
6. Ensure no cross-package imports except `agentteams_protocol`

### Updating Agent-Facing Content

1. Follow second-person voice conventions (see [Agent Content](../agent-content/skills-and-prompts.md))
2. Test skills with running Manager instance
3. Update relevant worker templates if needed
4. Record changes in changelog if affecting built images

## Source References

- Makefile: [`Makefile`](../../Makefile)
- Controller Makefile: [`agentteams-controller/Makefile`](../../agentteams-controller/Makefile)
- Development guide: [`docs/development.md`](../../docs/development.md)
- CI workflows: [`.github/workflows/`](../../.github/workflows/)
- Pre-commit: [`.pre-commit-config.yaml`](../../.pre-commit-config.yaml)
- Changelog: [`changelog/current.md`](../../changelog/current.md)
- Integration tests: [`tests/`](../../tests/)
- Shared Python libraries: [`shared/python/AGENTS.md`](../../shared/python/AGENTS.md)
- Shared shell libraries: [`shared/lib/`](../../shared/lib/)
