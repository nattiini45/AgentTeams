# Upstream integration: rename acceptance + fork overlay

This document describes how the fork absorbs `agentscope-ai/AgentTeams` after the
upstream hard-cut rename (`hiclaw-controller` → `agentteams-controller`,
`helm/hiclaw` → `helm/agentteams`, CLI `hiclaw` → `agt`).

**Stale claims removed:** dual `hiclaw.io` + `agentteams.io` CRDs and live
`HICLAW_*` runtime aliases are **not** the current DoD. Upstream hard-cut
removed those aliases; both sides of this merge are `agentteams.io` /
`AGENTTEAMS_*` only.

## Target baseline

| Side | Layout |
|------|--------|
| Upstream | `agentteams-controller/`, `helm/agentteams/`, CLI `agt` |
| Fork overlay | Dashboard, Project CRD, Gastown skills, `shared/python/`, remediation gates, quiet-rooms, Higress extra-provider prefix strip, SoloOperator, status overview CLI |

Merge method: staged **merge** (not rebase) onto `sync/upstream-main`, accepting
upstream path layout first, then replaying fork-only product value.

## Operator notes (existing clusters)

Fresh installs use AgentTeams names end-to-end. Clusters that still carry
historical `hiclaw.io` objects from older forks should treat conversion as a
one-time ops exercise outside this branch’s default path — do not reintroduce
dual-CRD templates into `helm/agentteams/crds/`.

CLI: use `agt` (not `hiclaw`). Environment: `AGENTTEAMS_*` only.

## Developer definition of done

Before merging the sync branch, every item below must hold:

- **Rename acceptance**: canonical trees are `agentteams-controller/` and
  `helm/agentteams/`. No active source path depends on `hiclaw-controller/` or
  `helm/hiclaw/`. CLI binary and agent-facing skill docs use `agt`.
- **Fork overlay present**:
  - `dashboard/` + `helm/agentteams/templates/dashboard/`
  - Project CRD (`projects.agentteams.io`) + reconciler + REST routes
  - Health monitor, message injection, manager-tasks APIs
  - `shared/python/agentteams_*` + remediation-gates path updates
  - Gastown skills (escalation / session-recovery / provider-management)
  - CoPaw quiet-rooms (`AGENTTEAMS_QUIET_ROOMS`) + upstream Team Leader → Team
    Room routing
  - Higress `AGENTTEAMS_EXTRA_LLM_PROVIDERS` modelMapping prefix strip
- **Upstream security / coordination fixes retained**: Team Leader room routing,
  migrate/CLI hardenings from upstream `main`.
- **Controller builds**: `CGO_ENABLED=1` for `cmd/controller`; `CGO_ENABLED=0`
  for `cmd/agt`. Project types register and deepcopy.
- **Gates**: remediation-gates jobs pass against `agentteams-controller` /
  `helm/agentteams` (controller, python-runtimes, shared-python, dashboard,
  helm, shared-shell).

## Deferred: non-root Manager images

Default installs still run Manager/Worker as root. Full non-root remains
deferred — see [development.md](development.md). Do not ship a partial UID
change that fails at first `mkdir`.
