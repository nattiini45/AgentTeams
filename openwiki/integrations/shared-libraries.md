---
type: Reference
title: Shared Libraries
description: Five shared Python packages and shell libraries providing domain logic across all AgentTeams runtimes.
tags: [shared-libraries, python, shell, packages, domain-logic, runtimes]
verified:
  - by: openwiki/0.7.1
    at: 2026-10-08T08:19:13.018Z
sources:
  - id: openwiki-source-ffd6aaeb62ab32468d8e6996
    resource: repo://copaw/Dockerfile
  - id: openwiki-source-6611a78e24321ddf152d2493
    resource: repo://hermes/Dockerfile
  - id: openwiki-source-ef6e752277a37cde760a5429
    resource: repo://shared/lib/agentteams-env.sh
  - id: openwiki-source-4ac88d6ddb278dc010745528
    resource: repo://shared/lib/mc-wrapper.sh
  - id: openwiki-source-fb7ee92bc7849ae57802387e
    resource: repo://shared/lib/oss-credentials.sh
  - id: openwiki-source-ac90dd232039b46523abdc8b
    resource: repo://shared/lib/render-skills.sh
  - id: openwiki-source-e2c41317a1ead02a54ef6b91
    resource: repo://shared/python/AGENTS.md
  - id: openwiki-source-9ce1b289bec6502ecb0b4701
    resource: repo://shared/python/agentteams_matrix_format/src/agentteams_matrix_format/__init__.py
  - id: openwiki-source-fbb27809b3f88c0f070536da
    resource: repo://shared/python/agentteams_matrix_policies/src/agentteams_matrix_policies/__init__.py
  - id: openwiki-source-ff6aa58e65d3fdc881863ebf
    resource: repo://shared/python/agentteams_openclaw_merge/src/agentteams_openclaw_merge/__init__.py
  - id: openwiki-source-1a7c0e9f5fdc08668214d066
    resource: repo://shared/python/agentteams_protocol/src/agentteams_protocol/__init__.py
  - id: openwiki-source-24dd44841985c2438213b390
    resource: repo://shared/python/agentteams_protocol/tests/test_verify_artifacts.py
  - id: openwiki-source-614bc4d1f026a508ac3ccf53
    resource: repo://shared/python/agentteams_sync/pyproject.toml
  - id: openwiki-source-2bd222fc7059b993942b23a4
    resource: repo://shared/python/agentteams_sync/src/agentteams_sync/__init__.py
  - id: openwiki-source-e465ef634a402b10354d1a39
    resource: repo://shared/tests/test-agentteams-env.sh
  - id: openwiki-source-c0cd2ef9b294d737f02ac874
    resource: repo://shared/tests/test-oss-credentials.sh
  - id: openwiki-source-57ee6679dfbabf27c935b1e4
    resource: repo://worker/Dockerfile
generated: { by: "openwiki/0.7.1", at: "2026-10-08T08:19:13.018Z" }
---

# Shared Libraries

AgentTeams provides a set of shared libraries—five Python packages and several shell scripts—that implement common domain logic used by all worker runtimes (CoPaw, Hermes, QwenPaw, OpenHuman, OpenClaw). These libraries ensure consistent behavior across runtimes while reducing code duplication and maintenance overhead.

## Overview

The shared libraries are located in two directories:

- **Shell libraries**: `shared/lib/` — Bootstrap scripts for environment setup, credential management, and utility wrappers
- **Python packages**: `shared/python/agentteams_*/` — Domain logic packages for file synchronization, protocol validation, Matrix integration, and configuration management

All worker container images include these libraries as build-context dependencies, and changes to shared libraries trigger rebuilds of dependent runtime images.

## Shell Libraries

The shell libraries provide environment bootstrapping, credential management, and utility functions for all container runtimes.

### Core Shell Libraries

| Library | Purpose | Key Functions |
|---------|---------|---------------|
| `agentteams-env.sh` | Unified environment bootstrap for Manager and Worker containers | `AGENTTEAMS_RUNTIME`, `AGENTTEAMS_MATRIX_URL`, `AGENTTEAMS_STORAGE_PREFIX`, `ensure_mc_credentials` |
| `oss-credentials.sh` | STS credential management for MinIO object storage | `ensure_mc_credentials`, `_oss_refresh_sts_via_controller`, credential caching |
| `mc-wrapper.sh` | Transparent STS credential refresh for MinIO Client (mc) | Auto-refreshes credentials before every mc invocation |
| `render-skills.sh` | Replace environment variable placeholders in agent documentation files | Renders `${AGENTTEAMS_*}` variables in markdown files |

### Additional Shell Scripts

| Script | Purpose |
|--------|---------|
| `merge-openclaw-config.sh` | Merge openclaw.json configuration files |
| `resolve-model-params.sh` | Resolve LLM model parameters |
| `sync-shared-worker-skills.sh` | Synchronize shared worker skills |

### Environment Bootstrap (`agentteams-env.sh`)

This script provides a single source of truth for environment variables across Manager and Worker containers. Key responsibilities:

1. **Runtime Detection**: Automatically detects runtime environment (`aliyun`, `k8s`, `docker`, `none`)
2. **Variable Normalization**: Sets default values for Matrix URLs, AI Gateway, and storage configuration
3. **Worker Environment Loading**: Handles dynamic environment injection for sandbox pool workers
4. **Credential Management**: Integrates with STS credential refresh for cloud deployments

**Key Variables Provided:**
- `AGENTTEAMS_RUNTIME` — Current runtime environment
- `AGENTTEAMS_MATRIX_URL` — Matrix server URL
- `AGENTTEAMS_AI_GATEWAY_URL` — AI Gateway base URL
- `AGENTTEAMS_STORAGE_PREFIX` — MinIO storage path prefix
- `AGENTTEAMS_EMBEDDING_MODEL` — Default embedding model (text-embedding-v4)

### Credential Management (`oss-credentials.sh`)

Manages STS (Security Token Service) credentials for MinIO object storage:

- **Cloud Mode**: Uses controller-mediated STS tokens via REST API
- **Local Mode**: No-op; uses static credentials configured in mc alias
- **Token Refresh**: Lazy-refreshes credentials with 10-minute margin before expiry
- **Security**: Validates environment files to prevent code injection

**Credential Flow:**
1. Resolve bearer token from `AGENTTEAMS_AUTH_TOKEN` or `AGENTTEAMS_AUTH_TOKEN_FILE`
2. Call controller `/api/v1/credentials/sts` endpoint
3. Parse STS credentials (access key, secret, security token)
4. Cache credentials in `/tmp/mc-oss-credentials.env`
5. Export `MC_HOST_<alias>` for mc commands

### MinIO Client Wrapper (`mc-wrapper.sh`)

Transparent wrapper that ensures STS credentials are refreshed before every mc invocation:

```bash
# Installed as /usr/local/bin/mc (symlink)
# Real binary at /usr/local/bin/mc.bin
# In cloud mode: refreshes STS credentials
# In local mode: no-op (near-zero overhead)
```

## Python Packages

Five independently-installable Python packages provide domain logic shared across all runtimes. Each uses setuptools src-layout (`src/agentteams_<name>/`) and starts at version `>=0.1.0`.

### Package Overview

| Package | Purpose | Key Modules |
|---------|---------|-------------|
| `agentteams_protocol` | Domain models, task/project DAG validation | `task.py`, `errors.py` |
| `agentteams_sync` | MinIO file sync, daemon, per-runtime push policies | `filesync.py`, `daemon.py`, `contract.py`, `policy.py` |
| `agentteams_openclaw_merge` | Canonical openclaw.json merge logic | `merge.py`, `__main__.py` |
| `agentteams_matrix_format` | Markdown-it rendering for Matrix messages | `__init__.py` |
| `agentteams_matrix_policies` | Matrix channel allow-list policy builder | `policies.py` |

### Dependency Rules

**Critical Constraint**: No cross-package imports except `agentteams_protocol`. All other packages may import `agentteams_protocol`, but must not import each other.

**Dependency Graph:**
```
agentteams_protocol (base)
├── agentteams_openclaw_merge
├── agentteams_sync
├── agentteams_matrix_format
└── agentteams_matrix_policies
```

### Package Details

#### `agentteams_protocol`

Core domain models and validation logic for task/project management.

**Key Exports:**
- `TaskMeta`, `TaskResult`, `ProjectMeta` — Domain models
- `DagTask`, `LoopPlan` — Task flow types
- `validate_dag`, `validate_task_result` — Validation functions
- `create_project`, `submit_task`, `complete_project` — Lifecycle operations
- `FileSystemTaskStore` — File-based task storage

**Usage**: All runtimes use this package for task management and project coordination.

#### `agentteams_sync`

MinIO file synchronization with per-runtime push policies.

**Key Components:**
- `FileSync` — Core synchronization class
- `SyncContract` — Runtime-specific sync contracts (CoPaw, Hermes, OpenClaw, etc.)
- `PushPolicy` — File push policies per runtime
- `push_loop`, `sync_loop` — Background synchronization loops

**Runtime Contracts**: Each runtime defines specific sync behaviors (e.g., `OPENCLAW`, `COPAW`, `HERMES`).

#### `agentteams_openclaw_merge`

Deep merge logic for OpenClaw configuration files.

**Key Functions:**
- `deep_merge(source, target)` — Recursive dictionary merging
- `merge_openclaw_config(config_path)` — Merge configuration with defaults

**CLI Usage**: Includes `__main__.py` for command-line invocation.

#### `agentteams_matrix_format`

Matrix message formatting with Markdown rendering.

**Key Functions:**
- `md_to_html(text)` — Convert Markdown to HTML for Matrix `formatted_body`
- `render_with_markdown_it(text)` — Render with markdown-it-py
- `edit_fallback_html(text)` — Format edit fallbacks

**Dependencies**: Requires `markdown-it-py>=3.0` and `linkify-it-py>=2.0`.

#### `agentteams_matrix_policies`

Matrix channel message filtering and mention handling.

**Key Components:**
- `DualAllowList` — Dual allow-list for message filtering
- `HistoryBuffer` — Message history management
- `should_suppress_outbound` — Message suppression logic
- `extract_mentions_from_text` — Mention extraction
- `apply_outbound_mentions` — Mention formatting

## Build Integration

### Docker Build Context

Shared libraries are included as named build contexts in Dockerfiles:

```dockerfile
# Named build context for shared libraries (requires BuildKit / Docker 23+)
SHARED_LIB_CTX = --build-context shared=./shared/lib
```

### Installation in Runtimes

Each runtime installs the required shared packages:

**OpenClaw Worker:**
```dockerfile
COPY --from=shared . /opt/agentteams/scripts/lib/
COPY --from=agentteams-openclaw-merge . /tmp/agentteams-openclaw-merge/
COPY --from=agentteams-sync . /tmp/agentteams-sync/
RUN python3 -m pip install --no-cache-dir /tmp/agentteams-openclaw-merge/ /tmp/agentteams-sync/
```

**CoPaw Worker:**
```dockerfile
COPY --from=agentteams-protocol . /tmp/agentteams-protocol/
COPY --from=agentteams-openclaw-merge . /tmp/agentteams-openclaw-merge/
COPY --from=agentteams-sync . /tmp/agentteams-sync/
RUN /opt/venv/standard/bin/pip install --no-cache-dir /tmp/agentteams-protocol/ /tmp/agentteams-openclaw-merge/ /tmp/agentteams-sync/
```

**Hermes Worker:**
```dockerfile
COPY --from=agentteams-matrix-policies . /tmp/agentteams-matrix-policies/
COPY --from=agentteams-protocol . /tmp/agentteams-protocol/
COPY --from=agentteams-openclaw-merge . /tmp/agentteams-openclaw-merge/
COPY --from=agentteams-sync . /tmp/agentteams-sync/
RUN /opt/venv/hermes/bin/pip install --no-cache-dir /tmp/agentteams-protocol/ /tmp/agentteams-openclaw-merge/ /tmp/agentteams-sync/ /tmp/agentteams-matrix-policies/
```

## Testing

### Shell Library Tests

Shell library tests are located in `shared/tests/` and can be run individually:

```bash
# Test agentteams-env.sh
bash shared/tests/test-agentteams-env.sh

# Test oss-credentials.sh
bash shared/tests/test-oss-credentials.sh

# Test merge-openclaw-config.sh
bash shared/tests/test-merge-openclaw-config.sh
```

### Python Package Tests

Each Python package has its own test suite:

```bash
# Test all shared packages
for pkg in shared/python/agentteams_*; do
  [ -d "$pkg/tests" ] && python -m pytest -q "$pkg/tests"
done

# Test specific package
python -m pytest -q shared/python/agentteams_protocol/tests

# Via Makefile (includes runtime tests)
make test-python
```

### CI/CD Integration

The `remediation-gates.yml` CI job automatically discovers and tests all `shared/python/agentteams_*` packages on every PR.

## Configuration

### Environment Variables

**Runtime Configuration:**
- `AGENTTEAMS_RUNTIME` — Runtime environment (`aliyun`, `k8s`, `docker`, `none`)
- `AGENTTEAMS_MATRIX_URL` — Matrix server URL
- `AGENTTEAMS_AI_GATEWAY_URL` — AI Gateway URL
- `AGENTTEAMS_STORAGE_PREFIX` — MinIO storage prefix

**Worker Configuration:**
- `AGENTTEAMS_WORKER_NAME` — Worker instance name
- `AGENTTEAMS_AUTH_TOKEN_FILE` — Path to authentication token file
- `AGENTTEAMS_WORKER_ENV_MOUNT_DIR` — Directory for dynamic environment injection

**Credential Configuration:**
- `AGENTTEAMS_CONTROLLER_URL` — Controller REST API URL
- `AGENTTEAMS_AUTH_TOKEN` — Bearer token for controller authentication
- `AGENTTEAMS_FS_ACCESS_KEY` / `AGENTTEAMS_FS_SECRET_KEY` — Static MinIO credentials

### Storage Configuration

The storage system uses a hierarchical configuration:

```bash
AGENTTEAMS_STORAGE_PREFIX="${AGENTTEAMS_STORAGE_PREFIX:-${AGENTTEAMS_STORAGE_ALIAS}/${AGENTTEAMS_FS_BUCKET}}"
```

Default values:
- `AGENTTEAMS_STORAGE_ALIAS` — `agentteams`
- `AGENTTEAMS_FS_BUCKET` — `agentteams-storage`
- `AGENTTEAMS_STORAGE_PREFIX` — `agentteams/agentteams-storage`

## Extension Points

### Adding New Shared Packages

To add a new shared Python package:

1. Create directory: `shared/python/agentteams_<name>/`
2. Add `pyproject.toml` with `agentteams-protocol` dependency (if needed)
3. Implement package in `src/agentteams_<name>/`
4. Add tests in `tests/`
5. Update Dockerfiles to install the new package

### Customizing Runtime Behavior

Each runtime defines its own `SyncContract` in `agentteams_sync`:

```python
# Example: Adding a new runtime
NEW_RUNTIME = "new-runtime"
RUNTIME_CONTRACTS[NEW_RUNTIME] = SyncContract(
    runtime=NEW_RUNTIME,
    # ... contract configuration
)
```

### Extending Credential Providers

The credential system supports extensible providers:

1. Implement `_oss_refresh_sts_via_<provider>()` function
2. Register in `ensure_mc_credentials()` logic
3. Update environment variable handling

## Failure Modes

### Credential Refresh Failures

- **Cached Fallback**: Uses cached credentials if refresh fails
- **Timeout Handling**: Configurable timeout for controller API calls
- **Security Validation**: Rejects environment files with command substitution

### File Sync Failures

- **Retry Logic**: Automatic retries for transient MinIO errors
- **Conflict Resolution**: Last-write-wins for concurrent modifications
- **Health Monitoring**: Heartbeat-based sync status reporting

### Configuration Merge Failures

- **Deep Merge Safety**: Recursive merging with type checking
- **Default Fallbacks**: Graceful degradation to default configuration
- **Validation**: Schema validation for critical configuration fields

## Performance Considerations

### Credential Caching

- **Lazy Refresh**: Only refreshes credentials when needed
- **Margin-Based**: Refreshes 10 minutes before expiry
- **File-Based Cache**: Persists across container restarts

### File Sync Optimization

- **Background Loops**: Non-blocking synchronization
- **Delta Detection**: Only syncs changed files
- **Compression**: Optional compression for large files

### Package Installation

- **Layer Caching**: Docker layer optimization for shared packages
- **Dependency Resolution**: Efficient pip dependency resolution
- **Virtual Environments**: Isolated environments per runtime

## Security

### Credential Security

- **No Command Substitution**: Rejects environment files with `$()` or backticks
- **File Permissions**: Sets 600 permissions on credential files
- **Token Isolation**: Bearer tokens never logged or exposed

### Environment Validation

- **Input Sanitization**: Validates all environment variables
- **Path Traversal Protection**: Prevents directory traversal attacks
- **Secure Defaults**: Conservative default configurations

### Network Security

- **TLS Enforcement**: HTTPS for all controller communications
- **Token Rotation**: Automatic STS token rotation
- **Connection Timeouts**: Configurable timeouts for all network calls

## Monitoring

### Health Checks

- **Credential Status**: Monitor STS token expiry
- **Sync Health**: File synchronization status
- **Runtime Detection**: Verify runtime environment

### Logging

- **Structured Logging**: JSON-formatted logs for analysis
- **Error Reporting**: Detailed error messages with context
- **Audit Trail**: Credential refresh and file sync logging

## Migration Guide

### From Static Credentials

1. Set `AGENTTEAMS_CONTROLLER_URL` and `AGENTTEAMS_AUTH_TOKEN`
2. Remove static MinIO credentials
3. Verify STS token refresh works

### Between Runtimes

1. Update `AGENTTEAMS_RUNTIME` environment variable
2. Verify package compatibility
3. Test file synchronization

## Related Pages

- [Architecture Overview](../architecture/overview.md) — System architecture and component relationships
- [Build & Test](../development/build-and-test.md) — Development workflow and testing
- [Plugin System](./plugin-system.md) — Plugin architecture and runtime extensions
- [Manager Overview](../manager/overview.md) — Manager component details
- [Worker Runtimes](../workers/runtime-guide.md) — Runtime-specific configurations
