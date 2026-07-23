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
| `make build-controller` | Build controller image |
| `make build-embedded` | Build embedded (all-in-one) image |

### Test Targets

| Target | Description |
|--------|-------------|
| `make test` | Build + run all integration tests |
| `make test SKIP_BUILD=1` | Run tests without rebuilding |
| `make test TEST_FILTER="01 02"` | Run specific tests |
| `make test-python` | Run Python package tests |
| `make test-embedded` | Run embedded controller tests |
| `make helm-lint` | Lint Helm chart |
| `make helm-template` | Render Helm templates |

### Push Targets

| Target | Description |
|--------|-------------|
| `make push` | Build + push multi-arch images (amd64 + arm64) |
| `make push-native` | Push native-arch images only (dev use) |

### Utility Targets

| Target | Description |
|--------|-------------|
| `make clean` | Remove local images and test containers |
| `make status` | Show status of Manager and Worker containers |
| `make logs` | Show recent logs (LINES=N to customize) |

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

## Testing

### Integration Tests

Integration tests live in [`tests/`](../../tests/) and run against a full embedded stack. They verify:
- Container startup and configuration
- Manager-Worker communication via Matrix
- Task delegation and completion
- File sync with MinIO
- Gateway routing

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

### Controller Tests

```bash
# Run Go tests
cd agentteams-controller && go test ./...

# Run specific test
cd agentteams-controller && go test ./internal/controller/...
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

| Workflow | Purpose |
|----------|---------|
| `test-integration.yml` | Integration test suite |
| `remediation-gates.yml` | Security and quality gates |
| `openwiki-update.yml` | OpenWiki documentation refresh |

### Pre-commit Hooks

[`.pre-commit-config.yaml`](../../.pre-commit-config.yaml) defines local hooks. Install with:

```bash
pre-commit install
```

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
- Development guide: [`docs/development.md`](../../docs/development.md)
- CI workflows: [`.github/workflows/`](../../.github/workflows/)
- Pre-commit: [`.pre-commit-config.yaml`](../../.pre-commit-config.yaml)
- Changelog: [`changelog/current.md`](../../changelog/current.md)
- Integration tests: [`tests/`](../../tests/)
