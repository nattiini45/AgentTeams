# QwenPaw Runtime Navigation Guide

This file helps AI Agents (and human developers) quickly understand the QwenPaw worker runtime package and find relevant code. It complements the root [AGENTS.md](../AGENTS.md); read that first for overall project structure.

## What is QwenPaw in AgentTeams

QwenPaw is a Python-based Worker runtime that wraps the upstream `qwenpaw` package (pinned at `1.1.11`) with AgentTeams-specific orchestration: config sync, heartbeat reporting, Matrix channel overlay, plugin bootstrap, and hot-reload updates. It is **not** the same as CoPaw — see root AGENTS.md runtime table.

## Scope of this guide

| In scope | Out of scope |
|---|---|
| `qwenpaw/` (Python package, Dockerfile, scripts, tests) | `plugins/teamharness/` MCP server and skills |
| `qwenpaw/src/matrix/` (Matrix channel overlay) | `manager/agent/qwenpaw-worker-agent/` (agent template) |
| `qwenpaw/scripts/` (entrypoint, gate, patches) | Upstream `qwenpaw` package source |

## Structure

```
qwenpaw/
├── Dockerfile                           # Worker image build (Python 3.12)
├── pyproject.toml                       # Package deps (qwenpaw-worker)
├── scripts/
│   ├── qwenpaw-worker-entrypoint.sh     # Container entrypoint
│   ├── qwenpaw_site_packages_gate.py    # Upstream version gate + overlay apply
│   ├── qwenpaw_upstream_manifest.json   # Checksums for gated upstream files
│   ├── patch-qwenpaw-defer-mcp-startup.py
│   ├── install-builtin-qwenpaw-plugins.py
│   └── qwenpaw-teamharness-plugin-reload.sh
├── src/
│   ├── qwenpaw_worker/                  # The Worker Python package
│   │   ├── cli.py                       # CLI entrypoint
│   │   ├── config.py                    # WorkerConfig dataclass
│   │   ├── worker.py                    # Worker orchestration + startup
│   │   ├── heartbeat.py                 # Heartbeat reporting loop
│   │   ├── sync.py                      # Storage sync (agentteams_sync shim)
│   │   ├── plugin_bootstrap.py          # Plugin installation at startup
│   │   ├── plugin_install.py            # Zip install / digest verification
│   │   ├── runtime_configurator.py      # runtime.yaml generation
│   │   ├── security_bootstrap.py        # Credential and auth setup
│   │   ├── log.py                       # Logging configuration
│   │   └── update/                      # Hot-reload update subpackage
│   │       ├── runtime_updater.py       # Update orchestrator
│   │       ├── runtime_config.py        # Config diff and apply
│   │       ├── agent_package.py         # Agent package sync
│   │       ├── model_sync.py            # Model parameter updates
│   │       ├── channel_writers.py       # Matrix channel config writers
│   │       ├── teams_prompt.py          # Team prompt injection
│   │       ├── constants.py             # Shared constants
│   │       └── utils.py                 # Helpers
│   └── matrix/
│       └── channel.py                   # Matrix channel overlay (4500+ lines)
└── tests/                               # Unit + integration tests (pytest)
```

## Entry points

- **CLI**: `qwenpaw_worker.cli:main` (invoked via `qwenpaw-worker` console script)
- **Worker start**: `qwenpaw_worker.worker:Worker.start()`
- **Container**: `scripts/qwenpaw-worker-entrypoint.sh`

## Test command

```bash
python -m pytest -q qwenpaw/tests
```

## Conventions

- Upstream pin: `qwenpaw==1.1.11` — do not upgrade without updating the gate manifest
- Site-packages gate (`qwenpaw_site_packages_gate.py`): applies Python overlays to the installed upstream package at build time; checksums in `qwenpaw_upstream_manifest.json`
- Matrix overlay (`src/matrix/channel.py`): replaces upstream Matrix channel; no direct Matrix API calls from worker code
- Storage sync delegates to `agentteams_sync` (PushPolicy.qwenpaw, byte-accurate compare ≤20 MiB)
- Matrix formatting delegates to `agentteams_matrix_format`

## Pitfalls

- After upstream qwenpaw changes, the overlay and gate checksums must be rebuilt
- `update/` subpackage modules are interdependent — changes to `runtime_config.py` often require `channel_writers.py` updates
- The Matrix overlay is large (4500+ lines); prefer behavior tests over source-lock tests
