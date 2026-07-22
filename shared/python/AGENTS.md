# Shared Python Libraries Navigation Guide

This file helps AI Agents (and human developers) quickly understand the shared Python domain libraries and find relevant code. It complements the root [AGENTS.md](../../AGENTS.md); read that first for overall project structure.

## What are the Shared Python Libraries

Five independently-installable Python packages providing domain logic shared across all AgentTeams worker runtimes (CoPaw, Hermes, QwenPaw, OpenHuman, OpenClaw). Each package is consumed as a build-context dependency in Dockerfiles and installed via `pip install -e`.

## Scope of this guide

| In scope | Out of scope |
|---|---|
| `shared/python/agentteams_*/` (5 packages) | Runtime-specific packages (`copaw/`, `hermes/`, `qwenpaw/`) |
| `shared/tests/` (shell tests for shared/lib) | `shared/lib/` (shell libraries — separate concern) |

## Package table

| Package | Purpose | Key modules |
|---------|---------|-------------|
| `agentteams_protocol` | Domain models, task/project DAG validation | `task.py` (1300+ lines), `errors.py` |
| `agentteams_sync` | MinIO file sync, daemon, per-runtime push policies | `filesync.py`, `daemon.py`, `contract.py`, `policy.py`, `openclaw.py` |
| `agentteams_openclaw_merge` | Canonical openclaw.json merge logic | `merge.py`, `__main__.py` (CLI wrapper) |
| `agentteams_matrix_format` | Markdown-it rendering for Matrix messages | `__init__.py` |
| `agentteams_matrix_policies` | Matrix channel allow-list policy builder | `policies.py` |

## Test command

```bash
# All shared packages
for pkg in shared/python/agentteams_*; do
  [ -d "$pkg/tests" ] && python -m pytest -q "$pkg/tests"
done

# Or via Makefile (includes runtime tests too)
make test-python
```

## Conventions

- **Namespace**: all packages use `agentteams_*` prefix
- **Layout**: setuptools src-layout (`src/agentteams_<name>/`)
- **Independence**: each package is independently installable via `pip install -e ./shared/python/agentteams_<name>`
- **Dependency rule**: no cross-package imports except `agentteams_protocol` ← others may import it
- **Versioning**: all start at `>=0.1.0`; no published PyPI releases (installed from source)

## Consumer note

Changes to these packages trigger rebuilds of dependent runtime images (copaw, hermes, qwenpaw, openhuman, worker, manager-copaw). The `remediation-gates.yml` CI job auto-discovers and tests all `shared/python/agentteams_*` packages on every PR.
