---
type: "Reference"
title: "Agent Content Model"
openwiki_generated: true
verified:
  - by: openwiki/0.7.0
    at: 2026-10-03T08:15:26.963Z
sources:
  - id: openwiki-source-4be8d89602ed92b69ba13da1
    resource: repo://manager/agent/AGENTS.md
  - id: openwiki-source-d0db781cb2f7c082c8289072
    resource: repo://manager/agent/copaw-manager-agent/AGENTS.md
  - id: openwiki-source-698383c1f268a67776b6bf34
    resource: repo://manager/agent/fragments/AGENTS/gotchas-openclaw.md
  - id: openwiki-source-172f1d38bc6a3a6ee6ef6619
    resource: repo://manager/agent/fragments/HEARTBEAT/openclaw-body.md
  - id: openwiki-source-9fa1c960450a88118c5cee89
    resource: repo://manager/agent/HEARTBEAT.md
  - id: openwiki-source-4b9e47a5943c8ec2ef5df0ce
    resource: repo://manager/agent/skills/task-management/references/finite-tasks.md
  - id: openwiki-source-4ab61b6b2d413b4e03d5b556
    resource: repo://manager/agent/skills/task-management/scripts/manage-state.sh
  - id: openwiki-source-f7351402557f85527745f95d
    resource: repo://manager/agent/skills/task-management/SKILL.md
  - id: openwiki-source-72a1d15365b85f6225ee157a
    resource: repo://manager/agent/SOUL.md
  - id: openwiki-source-27bf8e58b8d37073d91499a2
    resource: repo://manager/agent/TOOLS.md
  - id: openwiki-source-84339713f4aaf7691f639795
    resource: repo://manager/agent/worker-agent/AGENTS.md
  - id: openwiki-source-b28e752bb4e9a94966ef8c1d
    resource: repo://manager/scripts/init/upgrade-builtins.sh
  - id: openwiki-source-66196c857a5d6a61116a24f0
    resource: repo://manager/scripts/lib/render-manager-prompts.sh
  - id: openwiki-source-ac90dd232039b46523abdc8b
    resource: repo://shared/lib/render-skills.sh
generated: { by: "openwiki/0.7.0", at: "2026-10-03T08:15:26.963Z" }
---


# Agent Content Model

OpenClaw uses a structured set of Markdown files and directories to define agent behavior, capabilities, and operational rules. This page documents the conventions, structures, and pipelines that make up the agent content system.

## 1. File Types and Conventions

### AGENTS.md

The primary bootstrap instructions for an agent. Contains workspace layout, session startup steps, storage rules, gotchas, and memory management. Each runtime (OpenClaw, CoPaw) has its own version, assembled from composable fragments.

**Key responsibilities:**
- Define workspace paths (`~/`, shared storage, worker files)
- Host file access permissions and privacy rules
- Session initialization checklist (read SOUL.md, memory, YOLO mode)
- MinIO storage conventions and variable usage
- Critical operational gotchas (Worker creation, @mention rules, heartbeat)
- Controller API rules (use `agt` CLI, never raw curl)
- Memory management (daily notes, long-term MEMORY.md)

### SOUL.md

Defines the agent's identity, personality, and core nature. Contains AI identity statements, security rules, and behavioral guidelines.

**Structure:**
- **AI Identity** – clarifies the agent is not human, can work 24/7
- **Identity & Personality** – filled during onboarding with human admin
- **Core Nature** – delegation-first instinct, management vs execution boundaries
- **Security Rules** – response authorization, secret handling, file access

### HEARTBEAT.md

Periodic duties executed during heartbeat cycles. Contains a step-by-step checklist for monitoring tasks, checking worker status, handling escalations, and reporting to admin.

**Key steps:**
1. Read state.json, ensure admin notification channel
2. Check status of finite tasks
2b. Check team-delegated tasks
2c. Check escalation staleness
3. Check infinite task schedules
4. Check Worker health
5. Check pending tasks
6. Update shared knowledge
7. Report anomalies to admin

### TOOLS.md

Quick reference for management skills. Lists available skills, mandatory routing rules, skill boundaries, and cross-skill combos.

**Contains:**
- Available skills list with descriptions
- Mandatory routing rules (e.g., when to load `agentteams-find-worker`)
- Skill boundary definitions (primary vs auxiliary skills)
- Cross-skill combo table for common multi-skill workflows

### SKILL.md

Each skill directory contains a `SKILL.md` file that defines the skill's purpose, usage, and references. Skills are auto-loaded based on their description field when the agent encounters matching scenarios.

## 2. Skill Structure

Skills are organized in directories under `skills/` (for manager) and `worker-skills/` (for workers). Each skill follows a consistent layout:

```
skills/<name>/
├── SKILL.md          # Skill definition, frontmatter, gotchas, operation reference
├── scripts/          # Executable scripts (shell, Python) for skill operations
└── references/       # Detailed documentation for specific operations
```

### SKILL.md Format

The SKILL.md file uses YAML frontmatter and Markdown content:

```yaml
---
name: task-management
description: Use when admin gives a task to delegate to a Worker, when a Worker reports task completion, when managing recurring scheduled tasks, or when you need to check worker availability.
---

# Task Management

## Gotchas
- **Don't let Workers hallucinate on unfamiliar domains** — when a task involves niche frameworks...
- **Delegation-first** — always prefer assigning to a Worker over doing it yourself...

## Operation Reference
| Situation | Read |
|---|---|
| Admin gives task, no Worker specified | `references/worker-selection.md` |
| Assign a one-off task or handle completion | `references/finite-tasks.md` |
```

**Key elements:**
- **Frontmatter**: `name` (skill identifier), `description` (when to load the skill)
- **Gotchas**: Critical rules and warnings
- **Operation Reference**: Table mapping situations to reference docs

### Scripts Directory

Contains executable scripts for skill operations. Scripts are copied to workspace during upgrade and made executable. Example from task-management:

- `manage-state.sh` – manages state.json (atomic operations, deduplication)
- `resolve-notify-channel.sh` – resolves admin DM room for notifications

### References Directory

Detailed documentation for specific operations. Referenced from SKILL.md's operation reference table. Example references:

- `worker-selection.md` – how to choose a Worker for a task
- `finite-tasks.md` – one-off task assignment and completion flow
- `state-management.md` – state.json update procedures

## 3. Prompt Fragments System

Agent prompts are assembled from composable fragments stored in `fragments/`. This allows runtime-specific variations while sharing common content.

### Fragment Structure

```
fragments/
├── AGENTS/
│   ├── header-openclaw.md        # OpenClaw-specific header
│   ├── header-copaw.md          # CoPaw-specific header
│   ├── host-files.md            # Host file access rules (shared)
│   ├── every-session.md         # Session startup checklist
│   ├── minio.md                 # MinIO storage conventions
│   ├── gotchas-openclaw.md      # OpenClaw-specific gotchas
│   ├── gotchas-copaw.md         # CoPaw-specific gotchas
│   ├── controller-api.md        # Controller API rules
│   ├── memory.md                # Memory management
│   ├── tools.md                 # Tools reference
│   ├── management-skills.md     # Management skills list
│   ├── group-rooms-openclaw.md  # OpenClaw group room rules
│   ├── group-rooms-copaw.md     # CoPaw group room rules
│   ├── heartbeat-section.md     # Heartbeat reference
│   └── safety.md                # Safety rules
└── HEARTBEAT/
    ├── header-openclaw.md       # OpenClaw heartbeat header
    ├── header-copaw.md         # CoPaw heartbeat header
    ├── step-01-state.md        # State.json reading step
    ├── openclaw-body.md        # OpenClaw heartbeat steps
    ├── copaw-body.md           # CoPaw heartbeat steps
    └── copaw-cli-reference.md  # CoPaw CLI reference
```

### Assembly Process

The `render-manager-prompts.sh` script concatenates fragments in a specific order to produce runtime-specific AGENTS.md and HEARTBEAT.md files.

**OpenClaw AGENTS.md assembly order:**
1. `header-openclaw.md`
2. `host-files.md`
3. `every-session.md`
4. `minio.md`
5. `gotchas-openclaw.md`
6. `controller-api.md`
7. `memory.md`
8. `tools.md`
9. `management-skills.md`
10. `group-rooms-openclaw.md`
11. `heartbeat-section.md`
12. `safety.md`

**CoPaw AGENTS.md assembly order:**
1. `header-copaw.md`
2. `host-files.md`
3. `every-session.md`
4. `minio.md`
5. `gotchas-copaw.md`
6. `message-sending-copaw.md`
7. `controller-api.md`
8. `memory.md`
9. `tools.md`
10. `management-skills.md`
11. `group-rooms-copaw.md`
12. `heartbeat-section.md`
13. `safety.md`

## 4. Workspace Template Mapping per Runtime

Different runtimes use different workspace templates. The controller (or local registry) records `runtime` per worker, and init scripts materialize the matching template.

### Manager Runtimes

| Runtime | Template Source | Key Files |
|---------|----------------|-----------|
| `openclaw` (default) | `manager/agent/` | AGENTS.md, HEARTBEAT.md, SOUL.md, TOOLS.md, skills/ |
| `copaw` | `manager/agent/copaw-manager-agent/` | Copaw-specific AGENTS.md and HEARTBEAT.md, shares skills/ |

**Startup behavior:**
- **OpenClaw Manager**: workspace copies of AGENTS.md, HEARTBEAT.md, SOUL.md, plus everything under skills/
- **CoPaw Manager**: `copaw-manager-agent/AGENTS.md` and `copaw-manager-agent/HEARTBEAT.md` are merged into workspace copies during `upgrade-builtins.sh`

### Worker Runtimes

| Runtime | Template Source | Description |
|---------|----------------|-------------|
| `openclaw` (default) | `manager/agent/worker-agent/` | Primary OpenClaw worker template |
| `copaw` | `manager/agent/copaw-worker-agent/` | CoPaw Python worker template |
| `hermes` | `manager/agent/hermes-worker-agent/` | Hermes Matrix bridge worker |
| `openhuman` | `manager/agent/openhuman-worker-agent/` | Rust + native Matrix worker |
| `team_leader` | `manager/agent/team-leader-agent/` | Team Leader agent template |

**Worker template contents:**
- `AGENTS.md` – worker-specific workspace layout, session startup, gotchas
- `skills/` – builtin skills (file-sync, task-progress, find-skills, etc.)
- Optional `HEARTBEAT.md` for periodic duties

## 5. Render Pipeline

The agent content goes through a multi-stage rendering pipeline before agents read it.

### Pipeline Stages

1. **upgrade-builtins.sh** (first boot or image version change):
   - Step 0: Renders manager prompts from fragments using `render-manager-prompts.sh`
   - Step 1: Upgrades workspace .md files (merge builtin sections, preserve user content)
   - Step 2: Syncs scripts/ and references/ from image (always overwrite)
   - Step 3: Syncs builtins to registered workers' MinIO workspaces
   - Step 4: Writes installed version
   - Step 5: Marks workers needing builtin update notification

2. **render-skills.sh** (called from start-manager-agent.sh):
   - Replaces `${VAR}` placeholders in agent doc files with literal values
   - Only whitelisted variables are replaced (e.g., `${AGENTTEAMS_STORAGE_PREFIX}`, `${AGENTTEAMS_MATRIX_DOMAIN}`)
   - Leaves task-specific variables like `$task_id` untouched

3. **render-manager-prompts.sh** (called by upgrade-builtins.sh):
   - Assembles AGENTS.md and HEARTBEAT.md from fragments
   - Generates runtime-specific versions (openclaw, copaw)
   - Outputs to stdout or specified directory

### Variable Substitution

The `render-skills.sh` script uses `envsubst` with a whitelist of known variables:

```bash
VARS='${AGENTTEAMS_STORAGE_PREFIX} ${AGENTTEAMS_MATRIX_DOMAIN} ${AGENTTEAMS_MATRIX_URL}
${AGENTTEAMS_AI_GATEWAY_URL}
${AGENTTEAMS_ADMIN_USER} ${AGENTTEAMS_ADMIN_PASSWORD} ${AGENTTEAMS_REGISTRATION_TOKEN}
${AGENTTEAMS_DEFAULT_MODEL} ${AGENTTEAMS_AI_GATEWAY_DOMAIN} ${AGENTTEAMS_FS_DOMAIN}
${AGENTTEAMS_DEFAULT_WORKER_RUNTIME} ${AGENTTEAMS_WORKER_IMAGE} ${AGENTTEAMS_SKILLS_API_URL}
${AGENTTEAMS_CONTAINER_RUNTIME} ${AGENTTEAMS_GITHUB_TOKEN} ${AGENTTEAMS_WORKER_NAME}
${AGENTTEAMS_YOLO}
${MANAGER_MATRIX_TOKEN} ${MANAGER_TOKEN} ${HIGRESS_COOKIE_FILE}'
```

**Example transformation:**
- Before: `mc mirror ${AGENTTEAMS_STORAGE_PREFIX}/shared/tasks/{task-id}/ /root/agentteams-fs/shared/tasks/{task-id}/ --overwrite`
- After: `mc mirror agentteams-fs/mybucket/shared/tasks/123/ /root/agentteams-fs/shared/tasks/123/ --overwrite`

## 6. Second-Person Writing Convention

All agent-facing content uses second-person voice, addressing the agent directly. This creates clear, actionable instructions.

### Convention Rules

1. **Use "you" to address the agent**: "You are an AI Agent", "Your workspace"
2. **Use imperative mood for instructions**: "Read SOUL.md", "Don't ask permission"
3. **Avoid third-person descriptions**: Don't write "The agent should..." or "It is recommended..."
4. **Direct address for critical rules**: "You must receive explicit permission", "You can work continuously"
5. **Possessive pronouns for ownership**: "Your workspace", "Your skills", "Your memory"

### Examples

**Do:**
```markdown
## Every Session

Before doing anything:

1. Read `SOUL.md` — your identity and rules
2. Read `memory/YYYY-MM-DD.md` (today + yesterday) for recent context

Don't ask permission. Just do it.
```

**Don't:**
```markdown
## Session Initialization

The agent should read SOUL.md first. Then it should read memory files. This is mandatory.
```

**Do:**
```markdown
- **Your workspace:** `~/` (SOUL.md, openclaw.json, memory/, skills/, state.json, workers-registry.json — local only, host-mountable, never synced to MinIO)
- **Shared space:** `/root/agentteams-fs/shared/` (tasks, knowledge, collaboration data — synced with MinIO)
```

**Don't:**
```markdown
- The workspace directory is `~/`. It contains SOUL.md, openclaw.json, memory/, skills/, state.json, workers-registry.json. This is local only.
- The shared space is `/root/agentteams-fs/shared/`. It contains tasks, knowledge, collaboration data.
```

### Rationale

Second-person writing:
- Creates clear, actionable instructions
- Reduces ambiguity about who should perform actions
- Aligns with the agent's self-perception as an active entity
- Makes critical rules stand out (e.g., "You must...", "You can...")
- Matches the conversational nature of agent-human interaction

## 7. Controller Deployment to MinIO

The controller deploys agent content to MinIO for worker access during startup and updates.

### Deployment Process

1. **Image Build**: Agent content is copied to `/opt/agentteams/agent/` in the image
2. **Manager Startup**: `upgrade-builtins.sh` runs, syncing content to manager workspace
3. **Worker Sync**: Step 3 of `upgrade-builtins.sh` syncs builtins to registered workers' MinIO workspaces

### Worker Content Structure in MinIO

```
${AGENTTEAMS_STORAGE_PREFIX}/agents/<worker-name>/
├── AGENTS.md          # Worker-specific workspace instructions
├── HEARTBEAT.md       # Optional periodic duties
└── skills/            # Builtin and assigned skills
    ├── file-sync/     # File synchronization skill
    ├── task-progress/ # Task progress reporting
    ├── find-skills/   # Skill discovery
    └── <assigned>/    # Skills from workers-registry.json
```

### Sync Mechanisms

1. **Builtin Skills**: Copied from runtime-specific agent directory
   - OpenClaw: `manager/agent/worker-agent/skills/`
   - CoPaw: `manager/agent/copaw-worker-agent/skills/`
   - Hermes: `manager/agent/hermes-worker-agent/skills/`

2. **Assigned Skills**: From `workers-registry.json` `.workers.<name>.skills[]`
   - Source: `manager/agent/worker-skills/<skill-name>/`
   - Synced only to workers that have the skill assigned

3. **AGENTS.md Merge**: Uses `update_builtin_section_minio` to preserve user content after builtin-end marker

### Update Triggers

- **First boot**: Full sync of all builtins
- **Image version change**: Re-syncs builtins, preserving user content
- **Worker registration**: New worker gets builtins synced
- **Skill assignment**: `workers-registry.json` change triggers skill sync
- **Manual trigger**: `touch ${WORKSPACE}/.upgrade-pending-worker-notify`

## Auto-Loading Skills

OpenClaw automatically loads skills based on their description field in SKILL.md. When an agent encounters a situation matching a skill's description, it loads that skill's instructions.

### Auto-Loading Mechanism

1. **Description Matching**: Each SKILL.md has a `description` field in frontmatter
2. **Situation Recognition**: Agent recognizes when current task matches a skill description
3. **On-Demand Loading**: Agent reads the relevant SKILL.md when needed
4. **Progressive Disclosure**: Only loads full instructions when skill applies

### Example

```yaml
---
name: task-management
description: Use when admin gives a task to delegate to a Worker, when a Worker reports task completion, when managing recurring scheduled tasks, or when you need to check worker availability.
---
```

When the admin gives a task, the agent recognizes the match and loads `task-management/SKILL.md`.

## Related Pages

- [Manager Overview](../manager/overview.md) – Manager architecture and components
- [Worker Runtime Guide](../workers/runtime-guide.md) – Worker runtime specifics
- [Build and Test](../development/build-and-test.md) – Building and testing the system
- [Shared Libraries](../shared-libraries.md) – Shared shell libraries including render-skills.sh
- [TeamHarness and WorkerFlow](../plugins/teamharness-and-workerflow.md) – Team coordination plugins
