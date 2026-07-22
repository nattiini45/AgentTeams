# TeamHarness Plugin Navigation Guide

This file helps AI Agents (and human developers) quickly understand the TeamHarness plugin and find relevant code. It complements the root [AGENTS.md](../../AGENTS.md); read that first for overall project structure.

## What is TeamHarness

TeamHarness is the team-based task orchestration plugin for AgentTeams. It provides an MCP server (tools for project/task/file/artifact management), agent prompts and skills for team coordination, and runtime adapters that bridge the plugin into specific worker runtimes (currently QwenPaw).

## Scope of this guide

| In scope | Out of scope |
|---|---|
| `plugins/teamharness/` (MCP server, prompts, skills, adapters) | `plugins/workerflow/` (separate plugin) |
| `plugins/tests/teamharness/` (tests for this plugin) | `plugins/cli/` (plugin CLI tooling) |
| `plugins/adapters/qwenpaw/` (shared adapter scripts) | `qwenpaw/` runtime package itself |

## Structure

```
plugins/teamharness/
├── plugin.yaml                  # Plugin manifest (prompts, skills, MCP, adapters)
├── mcp/
│   ├── server.py                # MCP stdio server — tool registry and dispatch
│   ├── mcp_common.py            # Shared helpers (storage, config, auth context)
│   ├── protocol_bridge.py       # DAG validation bridge to agentteams_protocol
│   ├── message_tool.py          # Matrix message send/read tools
│   ├── roomflow_tool.py         # Room creation and membership tools
│   ├── _bootstrap.py            # Server startup wiring
│   └── tools/                   # Focused tool dispatch modules
│       ├── projectflow.py       # Project CRUD + DAG delegation
│       ├── taskflow.py          # Task lifecycle (create/assign/submit/review)
│       ├── filesync.py          # File pull/push/stat/list via agentteams_sync
│       ├── artifact.py          # Deliverable artifact verification
│       └── matrix_format.py     # Markdown-it rendering for Matrix messages
├── prompts/
│   ├── team/TEAMS.md            # Team coordination prompt
│   ├── agent/                   # Per-role agent prompts (leader, worker, remote-member)
│   └── manager/                 # Manager prompt overlays (AGENTS, TOOLS, HEARTBEAT)
├── skills/
│   ├── agent/                   # Agent skills (mcporter, find-skills)
│   └── team/                    # Team skills (communication, task-delegation, etc.)
├── adapters/
│   └── qwenpaw/                 # QwenPaw runtime adapter (plugin.py, matrix_channel.py)
│       └── scripts/             # Ruby build/validate scripts for adapter packaging
├── loongsuite/                   # LoongSuite agent descriptor
└── scripts/                     # Install/uninstall helpers
```

## Entry points

- **Plugin manifest**: `plugin.yaml` — declares prompts, skills, MCP server, and adapters
- **MCP server**: `python mcp/server.py` (stdio transport, registered tools: health, message, roomflow, filesync, artifact, projectflow, taskflow)

## Test command

```bash
python3 -m pytest plugins/tests/teamharness/ -q
```

## Conventions

- Tool dispatch modules live in `mcp/tools/`; shared helpers in `mcp_common.py`
- `protocol_bridge.py` delegates DAG validation to `agentteams_protocol`
- `filesync` tool delegates to `agentteams_sync.FileSync`
- Adapters generate runtime-specific configs via Ruby scripts (`scripts/build-qwenpaw-plugin.rb`)
- `plugin.yaml` is the single source of truth for what the plugin exposes

## Inter-module dependencies

- `agentteams_protocol` — domain models, task/project DAG validation
- `agentteams_sync` — MinIO file sync (filesync tool)
- `agentteams_matrix_format` — markdown rendering (matrix_format tool)
