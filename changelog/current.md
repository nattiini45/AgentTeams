# Changelog (Unreleased)

Record image-affecting changes to `manager/`, `worker/`, `copaw/`, `hermes/`,
`openclaw-base/`, `agentteams-controller/`, and release-facing install/chart
changes here before the next release.

---

**Bug Fixes**

- **QwenPaw file:// package refs on Windows**: Resolve `file://C:\\...` agent package refs correctly instead of treating an empty urlparse path as `.` (cwd).
- **CoPaw Team coordination routing**: Route Team Leader worker assignments sent through the `message` tool from Leader DM to Team Room, matching the Matrix channel send path. ([92c8145](https://github.com/agentscope-ai/AgentTeams/commit/92c8145))
- **Pinned OpenClaw source fetch**: Fetch the pinned OpenClaw commit directly so the base image build does not depend on a retired-brand external branch name. ([b0081c2](https://github.com/agentscope-ai/AgentTeams/commit/b0081c2))
- **Higress extra providers**: Restore OPT-IN `AGENTTEAMS_EXTRA_LLM_PROVIDERS` registration with `modelMapping` that strips the `<provider>/` prefix before upstream calls.
- **CoPaw quiet rooms**: Restore `AGENTTEAMS_QUIET_ROOMS` → `config.json` `show_tool_details: false` bridge and MatrixChannel read path.
- **CoPaw bridge/worker restore**: Restore fork `bridge_config` / `bootstrap_copaw_runtime` / `propagate_prompts` APIs and WorkspaceLayout-based worker startup so runtime re-bridge keeps local prompts and skips pruning self-installed skills.
- **Hermes sync semantics**: Restore thin `agentteams_sync` wrapper with byte-accurate push compare and keep local-only skills across pull.
- **Installer prompt safety**: Use `printf -v` for installer prompt helpers so preset/default values are never eval-executed.
- **OpenHuman gateway id**: Rename provider id to `agentteams-gateway` consistently in bridge and tests.

**Branding and Compatibility**

- **Complete AgentTeams runtime rename**: Rename installer and Helm entrypoints, the controller Go module and CLI, and container filesystem paths to AgentTeams. ([3121f5f](https://github.com/agentscope-ai/AgentTeams/commit/3121f5f))
- **Hard-cut AgentTeams naming**: Remove retired-brand installer wrappers, environment fallbacks, CLI aliases, Helm naming branches, runtime path migrations, and active source paths so fresh AgentTeams deployments use one canonical contract end to end. ([d20e606](https://github.com/agentscope-ai/AgentTeams/commit/d20e606617edefbbc42c28c1201c5629fa73fd88))

**Fork overlay (sync/upstream-main)**

- **Project CRD**: Port Project types, reconciler, Helm CRD/RBAC, and `/api/v1/projects` REST routes onto `agentteams-controller` / `helm/agentteams`.
- **Health + ops APIs**: Port health monitor, worker health probes, message injection, manager-tasks, and `agt manager-state`.
- **Status CLI**: Enhanced `agt status` overview (`--watch`, JSON, phase summaries).
- **SoloOperator / QwenPaw / Docker limits**: SoloOperator wiring, QwenPaw image builders, Docker CPU/memory limits, localhost console bind.
- **Dashboard Helm**: Restore operator dashboard templates under `helm/agentteams/templates/dashboard/`.
- **Remediation gates**: Point CI jobs at `agentteams-controller` and `helm/agentteams`.
- **Agent prompts**: Align agent-facing CLI examples from `hiclaw` to `agt` (hard-cut; no dual aliases).
