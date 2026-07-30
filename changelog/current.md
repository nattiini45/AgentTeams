# Changelog (Unreleased)

Record image-affecting changes to `manager/`, `worker/`, `copaw/`, `hermes/`,
`openclaw-base/`, `agentteams-controller/`, and release-facing install/chart
changes here before the next release.

---

**What's New**

- **Better Harness weekly self-review**: Bake the pinned `QoderAI/better-harness` CLI into the `openclaw-base`, `hermes`, and `qwenpaw` images (Node 22 via NodeSource for the Python images; `npm install --omit=dev` since the project is npm-managed) and ship a builtin `better-harness` skill in every agent template (Manager, OpenClaw/CoPaw/Hermes/OpenHuman/QwenPaw workers, Team Leader). Each agent runs a deterministic, read-only `better-harness harness source` evidence collection at most once per 7 days (deterministic `run-weekly.sh` cadence guard), authors findings from that evidence, and reports them to the Manager — never the human admin. The weekly trigger is native to each runtime: the QwenPaw and Hermes worker daemons run a supervised `run_better_harness_weekly_loop` background task (state persisted to MinIO, so the cadence survives restarts), while OpenClaw/CoPaw/Manager/Team Leader invoke `run-weekly.sh` from their existing heartbeat / session-wake.
- **Manager-mediated harness integration**: Add a `harness-integration` Manager skill (`aggregate-reports.sh`) that pulls `shared/harness-reports/`, aggregates all agents' findings into one consolidated summary, and presents a single fleet-wide integration plan to the admin for approval. Approved changes are written to the manager-owned builtin templates / controller-owned config so they propagate to every agent via `upgrade-builtins.sh` / `pushBuiltinSkills`; agents never self-apply hooks/loops/skill edits.
- **`AGENTTEAMS_BETTER_HARNESS_ENABLED` kill switch**: New controller env gate (default on). Setting it to a falsy value propagates `AGENTTEAMS_BETTER_HARNESS_ENABLED=0` to every worker and manager container so the weekly review is skipped fleet-wide.
- **Worker overview dashboard page**: New **Workers** tab in the dashboard web UI listing every worker; clicking a worker opens a drill-down detail with current health (reusing the health-probe strip), an event/failure timeline, and the worker's better-harness weekly findings (read from the existing MinIO `shared/harness-reports/<name>/` space). Backed by a new controller event-history store: `WorkerStatus.RecentEvents` is a bounded ring buffer (newest first, capped at 50) recorded by the health monitor (health transitions, zombie failures) and the lifecycle handlers (wake/sleep/ensure-ready), exposed via `GET /api/v1/workers/{name}/events`. CRD schema updated and synced to Helm.

**Bug Fixes**

- **QwenPaw file:// package refs on Windows**: Resolve `file://C:\\...` agent package refs correctly instead of treating an empty urlparse path as `.` (cwd).
- **CoPaw Team coordination routing**: Route Team Leader worker assignments sent through the `message` tool from Leader DM to Team Room, matching the Matrix channel send path. ([92c8145](https://github.com/agentscope-ai/AgentTeams/commit/92c8145))
- **Pinned OpenClaw source fetch**: Fetch the pinned OpenClaw commit directly so the base image build does not depend on a retired-brand external branch name. ([b0081c2](https://github.com/agentscope-ai/AgentTeams/commit/b0081c2))
- **Higress extra providers**: Restore OPT-IN `AGENTTEAMS_EXTRA_LLM_PROVIDERS` registration with `modelMapping` that strips the `<provider>/` prefix before upstream calls.
- **CoPaw quiet rooms**: Restore `AGENTTEAMS_QUIET_ROOMS` → `config.json` `show_tool_details: false` bridge and MatrixChannel read path.
- **CoPaw Matrix channel restore**: Restore fork Matrix channel startup-replay/bare-mention APIs, Team Leader DM preamble suppression, and shared filesync/_toolhelpers wiring.
- **CoPaw bridge/worker restore**: Restore fork `bridge_config` / `bootstrap_copaw_runtime` / `propagate_prompts` APIs and WorkspaceLayout-based worker startup so runtime re-bridge keeps local prompts and skips pruning self-installed skills.
- **Hermes sync semantics**: Restore thin `agentteams_sync` wrapper with byte-accurate push compare and keep local-only skills across pull.
- **Installer prompt safety**: Use `printf -v` for installer prompt helpers so preset/default values are never eval-executed.
- **OpenHuman gateway id**: Rename provider id to `agentteams-gateway` consistently in bridge and tests.
- **Controller appservice auth**: Compare Matrix `hs_token` with `subtle.ConstantTimeCompare` and cap the appservice mention dedup map.
- **Controller proxy DoS**: Bound Docker container-create request bodies with `io.LimitReader`.
- **Worker lifecycle errors**: Log backend Start/Stop and status-update failures in wake/sleep/ensure-ready instead of discarding them.
- **Helm worker images**: Wire `AGENTTEAMS_QWENPAW_WORKER_IMAGE`, align openhuman/qwenpaw image repos to the `agentteams/` registry namespace, and drop StorageClass `delete` RBAC.
- **OpenHuman config.toml**: Escape Matrix bridge string values before TOML heredoc interpolation.
- **CoPaw bridge Matrix identity**: Remove duplicate env lookups and restore `COPAW_*` fallbacks for domain/worker name.

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
