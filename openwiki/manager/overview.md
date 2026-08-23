---
type: "Reference"
title: "Manager"
description: "Coordinator agent in AgentTeams. Covers runtimes (OpenClaw, CoPaw), 19 skills, bootstrap chain, prompt fragments, and agent-facing content conventions."
tags: [manager, agent, coordination, skills]
openwiki:
  roles: [domain, architecture]
  change_kinds: [agent-config, skills]
  source_paths: [manager/agent/, manager/scripts/]
  symbols: [Manager, skills, AGENTS.md, HEARTBEAT.md, SOUL.md]
  test_paths: [manager/tests/]
  invariants: ["Manager orchestrates Workers via Matrix", "19 skill modules", "Skills shared across runtimes"]
  validation_commands: ["ls manager/agent/skills/"]
---

# Manager

The Manager is the coordinator agent in AgentTeams. It orchestrates Workers, manages Teams, handles Human interactions, and configures the Higress gateway. The Manager runs as a dedicated container and communicates with Workers and Humans through Matrix rooms.

## Manager Runtimes

The Manager supports two runtimes, selected via `AGENTTEAMS_MANAGER_RUNTIME` or Helm `manager.runtime`:

| Runtime | Image | Behavior |
|---------|-------|----------|
| `openclaw` (default) | `agentteams-manager` | OpenClaw gateway; primary Matrix channel uses the message tool pattern |
| `copaw` | `agentteams-manager-copaw` | Python CoPaw workspace; Matrix traffic uses `copaw channels send` CLI |

Hermes, OpenHuman, and QwenPaw are **Worker-only** runtimes — they cannot serve as Manager.

## Directory Structure

The Manager's agent-facing content lives in [`manager/agent/`](../../manager/agent/) and is copied to `/opt/agentteams/agent/` in the image. The `upgrade-builtins.sh` script syncs these files into the Manager workspace at startup.

```
manager/agent/
├── AGENTS.md                    # OpenClaw Manager bootstrap instructions
├── HEARTBEAT.md                 # OpenClaw Manager periodic duties
├── SOUL.md                      # Manager personality (filled by onboarding)
├── TOOLS.md                     # Available tools reference
├── skills/                      # 19 Manager skills (shared by both runtimes)
├── skills-alpha/                # Experimental skills
├── fragments/                   # Composable prompt fragments
│   ├── AGENTS/                  # AGENTS.md fragments by topic
│   └── HEARTBEAT/               # HEARTBEAT.md fragments by runtime
├── copaw-manager-agent/         # CoPaw Manager overrides
│   ├── AGENTS.md                # Replaces workspace AGENTS.md for CoPaw
│   └── HEARTBEAT.md             # Replaces workspace HEARTBEAT.md for CoPaw
├── worker-agent/                # OpenClaw Worker template
├── copaw-worker-agent/          # CoPaw Worker template
├── hermes-worker-agent/         # Hermes Worker template
├── openhuman-worker-agent/      # OpenHuman Worker template
├── qwenpaw-worker-agent/        # QwenPaw Worker template
├── team-leader-agent/           # Team Leader template
├── shared-worker-skills/        # Skills shared across runtimes
└── worker-skills/               # On-demand skills pushed to Workers
```

## Manager Skills

The Manager has 19 skill modules in [`manager/agent/skills/`](../../manager/agent/skills/). Each skill directory contains a `SKILL.md` (instructions for the agent), optional `scripts/` (executable helpers), and optional `references/` (detailed docs).

| Skill | Purpose |
|-------|---------|
| `channel-management` | Create and manage Matrix channels |
| `escalation-management` | Handle escalation protocols |
| `file-sync-management` | Manage MinIO file synchronization |
| `git-delegation-management` | Delegate git operations to workers |
| `agentteams-find-worker` | Find and query workers |
| `human-management` | Manage human participants |
| `matrix-server-management` | Matrix server administration |
| `mcp-server-management` | Create and configure MCP servers |
| `mcporter` | MCP tool calling via CLI |
| `model-switch` | Switch LLM models |
| `project-management` | Create and manage projects |
| `provider-management` | Manage LLM providers |
| `service-publishing` | Publish services |
| `session-recovery` | Recover from session interruptions |
| `task-coordination` | Coordinate tasks across workers |
| `task-management` | Create, assign, and track tasks |
| `team-management` | Create and manage teams |
| `worker-management` | Create, configure, and manage workers |
| `worker-model-switch` | Switch models for specific workers |

### Skill Structure Example

```
skills/task-management/
├── SKILL.md                        # Agent instructions
├── scripts/
│   ├── manage-state.sh             # Task state management
│   ├── send-task-message.sh        # Send task to worker via Matrix
│   └── verify-output.sh            # Verify task completion
└── references/
    ├── finite-tasks.md             # Finite task lifecycle
    ├── infinite-tasks.md           # Infinite task patterns
    ├── state-management.md         # State machine reference
    └── dispatch-gating.md          # Dispatch gating rules
```

## Prompt Fragments

The [`manager/agent/fragments/`](../../manager/agent/fragments/) directory contains composable prompt fragments that are assembled into the final AGENTS.md and HEARTBEAT.md:

**AGENTS fragments:**
- `header-openclaw.md` / `header-copaw.md` — Runtime-specific headers
- `every-session.md` — Per-session instructions
- `controller-api.md` — Controller API reference
- `management-skills.md` — Skill management
- `tools.md` — Available tools
- `memory.md` — Memory management
- `safety.md` — Safety guidelines
- `gotchas-openclaw.md` / `gotchas-copaw.md` — Runtime-specific gotchas
- `group-rooms-*.md` — Room management by runtime
- `message-sending-*.md` — Message sending by runtime
- `host-files.md` — File system access
- `minio.md` — MinIO integration

**HEARTBEAT fragments:**
- `header-openclaw.md` / `header-copaw.md` — Runtime-specific headers
- `openclaw-body.md` / `copaw-body.md` — Runtime-specific body
- `step-01-state.md` — State check step
- `copaw-cli-reference.md` — CoPaw CLI reference

## Manager Container

### Dockerfiles

- [`manager/Dockerfile`](../../manager/Dockerfile) — OpenClaw-based Manager (from `openclaw-base`; bundles `agt` CLI from controller image)
- [`manager/Dockerfile.copaw`](../../manager/Dockerfile.copaw) — CoPaw-based Manager (Python venv + CoPaw from PyPI)

### Bootstrap Chain

The Manager startup sequence is:

1. **Container entrypoint** starts supervisord (local) or direct process (K8s)
2. **`start-manager-agent.sh`** — Main bootstrap script
   - Detects runtime (`AGENTTEAMS_RUNTIME`: local/embedded/aliyun/k8s)
   - Runs bootstrap steps from [`manager/scripts/lib/bootstrap/`](../../manager/scripts/lib/bootstrap/)
3. **`upgrade-builtins.sh`** — Syncs built-in agent files to workspace
4. **`render-manager-prompts.sh`** — Renders prompt templates with variable substitution
5. **Runtime start** — Launches OpenClaw or CoPaw manager process

### Bootstrap Steps

Bootstrap scripts in [`manager/scripts/lib/bootstrap/`](../../manager/scripts/lib/bootstrap/):

| Script | Purpose |
|--------|---------|
| `pre-start.sh` | Pre-flight checks |
| `secrets.sh` | Secret loading and validation |
| `matrix-token.sh` | Matrix authentication |
| `workspace.sh` | Workspace directory setup |
| `local.sh` | Local-mode specific setup |
| `higress.sh` | Higress gateway configuration |
| `container-runtime.sh` | Docker/Podman runtime setup |
| `runtime.sh` | Runtime-specific initialization |
| `start-runtime.sh` | Launch the agent runtime |
| `admin-dm.sh` | Admin DM room setup |
| `cloud-sync.sh` | Cloud synchronization |
| `cloud-validate.sh` | Cloud config validation |
| `cms-plugin.sh` | CMS plugin setup |
| `workers.sh` | Worker initialization |
| `openclaw-config.sh` | OpenClaw configuration |

### Configuration

- [`manager/configs/`](../../manager/configs/) — Init-time configuration templates
- [`manager/configs/known-models.json`](../../manager/configs/known-models.json) — Known LLM model definitions
- [`manager/configs/manager-openclaw.json.tmpl`](../../manager/configs/manager-openclaw.json.tmpl) — OpenClaw Manager config template
- [`manager/configs/mcp-templates/`](../../manager/configs/mcp-templates/) — MCP server templates

## Agent-Facing Content Conventions

All files under `manager/agent/` are **read by the Agent at runtime**, not by human developers. Writing conventions:

- **Use second-person voice** — "You are the Manager...", "Your responsibilities include..."
- **SKILL.md** — Instruct the agent directly: "Use this script to...", "Run `mcporter list` to see..."
- **No third-person** — Never write "Manager can call..." or "This skill provides..."
- **Scripts** — Write log output from the system's perspective, with the agent as the implied operator

## Source References

- Agent configs: [`manager/agent/`](../../manager/agent/)
- Skills: [`manager/agent/skills/`](../../manager/agent/skills/)
- Fragments: [`manager/agent/fragments/`](../../manager/agent/fragments/)
- Scripts: [`manager/scripts/`](../../manager/scripts/)
- Bootstrap: [`manager/scripts/lib/bootstrap/`](../../manager/scripts/lib/bootstrap/)
- Dockerfiles: [`manager/Dockerfile`](../../manager/Dockerfile), [`manager/Dockerfile.copaw`](../../manager/Dockerfile.copaw)
- Configs: [`manager/configs/`](../../manager/configs/)
