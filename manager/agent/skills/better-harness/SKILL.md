---
name: better-harness
description: Use when running your weekly better-harness self-review of your own workspace/harness, or when reporting harness findings to your coordinator (Manager). Covers the weekly evidence-collection command, the 7-day cadence guard, and the report-to-Manager flow.
---

# Better Harness Weekly Self-Review

You run a weekly [better-harness](https://github.com/QoderAI/better-harness) review of your own workspace to surface how your harness (AGENTS.md, skills, memory, config) could improve. You report findings to your **coordinator (Manager)** — never directly to the human admin — and you never apply hooks/loops/skills changes yourself. The Manager aggregates all agents' reports and executes the integration fleet-wide.

## Gotchas

- **Report to your coordinator, not the user** — the Manager consumes your report and drives any change. Do not @mention the human admin for this.
- **Never self-apply hooks/loops/skill edits** — builtin `skills/` and manager-owned config are overwritten on upgrade/sync, so local edits there are lost. Durable changes are written by the Manager into the builtin templates.
- **Run at most once per 7 days** — the helper script enforces this; do not run the review on every wake.
- **Evidence is read-only** — the weekly collection gathers workspace/repo evidence only. Session evidence stays an explicit boundary.

## Run the weekly review

Once per session wake, run the cadence-guarded evidence collector:

```bash
bash skills/better-harness/scripts/run-weekly.sh
```

The script exits `0` with `SKIP` if fewer than 7 days have elapsed. When due, it collects a deterministic evidence envelope to `$HOME/.better-harness/report.source.json` and prints `READY`.

When you see `READY`, author your findings: read `$HOME/.better-harness/report.source.json`, evaluate your own harness against the Agent Work Loop (task understanding, controlled execution, change validation, reliable delivery, learning capture), and write a compact findings file:

```
$HOME/.better-harness/findings.json   — an array of { "title": ..., "summary": ... } objects
```

Keep findings honest: only include gaps you actually observed in the evidence; leave unobserved dimensions out rather than guessing.

## Report findings to your coordinator

After writing `findings.json`:

1. Copy it to the shared harness-reports prefix and push it:

   ```bash
   mkdir -p /root/agentteams-fs/shared/harness-reports/${AGENTTEAMS_WORKER_NAME}
   cp "$HOME/.better-harness/findings.json" \
      /root/agentteams-fs/shared/harness-reports/${AGENTTEAMS_WORKER_NAME}/$(date -u +%Y%m%d).json
   mc cp /root/agentteams-fs/shared/harness-reports/${AGENTTEAMS_WORKER_NAME}/$(date -u +%Y%m%d).json \
      ${AGENTTEAMS_STORAGE_PREFIX}/shared/harness-reports/${AGENTTEAMS_WORKER_NAME}/$(date -u +%Y%m%d).json
   ```

2. @mention your coordinator (Manager) with a one-line summary:

   ```
   @{coordinator}:{domain} BETTER_HARNESS_REPORT: <n> findings — <top finding title>
   ```

The Manager picks up the report from `shared/harness-reports/` and decides what to integrate. You do not need to do anything further.

## What gets integrated (by the Manager, not you)

- **Skills** — improved procedural knowledge is written into the builtin templates and pushed to every agent.
- **Hooks** — runtime-native hooks (e.g. QwenPaw plugin/startup hooks) are owned by the Manager/controller so they apply uniformly. (OpenClaw has no agent lifecycle-hook block — its `.hooks` key is the webhook ingress receiver, so OpenClaw reactive behavior is expressed as heartbeat steps or infinite tasks.)
- **Loops** — recurring work is scheduled by the Manager via the infinite-task mechanism (`manage-state.sh add-infinite --schedule CRON`).

Your job is only to run the review and report. Integration is the Manager's job.
