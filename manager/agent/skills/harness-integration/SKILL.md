---
name: harness-integration
description: Use when aggregating weekly better-harness reports from your agents, drafting a single fleet-wide harness integration plan, and applying approved changes to the builtin templates and controller-owned config. Agents report findings to you; you present one consolidated plan to the admin.
---

# Harness Integration (Manager)

Your agents each run a weekly better-harness self-review and report their findings to **you** (not to the human admin). You aggregate those reports, decide what is worth integrating, and apply approved changes to the **manager-owned builtin templates and controller-owned config** so they propagate to every agent. You are the single integration point — agents never self-apply hooks/loops/skill changes.

## Gotchas

- **Agents report to you, never to the admin** — you present one consolidated plan to the admin, not a stream of per-agent messages.
- **Approve before integrating** — draft a scoped plan and get explicit admin approval before writing to the builtin templates or controller config. This is the fleet-wide change gate.
- **Write only to manager-owned sources** — builtin `manager/agent/*` templates and the controller config generator. Never edit an individual agent's local workspace; that state is overwritten on sync/upgrade.
- **Durable hooks/loops live in the templates/config, not in agents** — see "Where each change lands" below.

## Aggregate agent reports

Run the aggregation pass (scans the shared harness-reports prefix, dedupes, summarizes):

```bash
bash /opt/agentteams/agent/skills/harness-integration/scripts/aggregate-reports.sh
```

It pulls `${AGENTTEAMS_STORAGE_PREFIX}/shared/harness-reports/` into a local dir and prints a consolidated JSON summary: how many agents reported, the distinct findings (by title), and which agents raised each. Read its output, then read the individual `findings.json` files it lists for detail.

## Draft and present the plan

1. Group the findings by theme (skills gaps, missing hooks, missing loops, config drift).
2. Decide the smallest durable change per theme and where it lands (see below).
3. Draft a scoped integration plan: per change — impact, expected output, repair boundary, acceptance checks.
4. Present it to the admin once via the resolved notification channel (see `task-management` → `resolve-notify-channel.sh`):

   ```
   [Harness Integration] <N> findings across <M> agents. Proposed: <one-line per change>. Approve?
   ```

5. Apply only after explicit approval.

## Where each change lands (so it persists and propagates)

- **Skills** — write the improved skill into `manager/agent/skills/<name>/` (Manager) and/or `manager/agent/<runtime>-worker-agent/skills/<name>/`. Propagation: `upgrade-builtins.sh` on your next boot/upgrade mirrors them to every registered worker; the controller's `pushBuiltinSkills` does the same on worker create.
- **QwenPaw hooks/loops** — add a workspace `crons/` entry or plugin behavior to the `qwenpaw-worker-agent` template (or the plugin package). Propagation: the skills mirror / plugin install path.
- **OpenClaw** — openclaw.json has **no lifecycle-hook block**; its `.hooks` key is the Gmail/webhook ingress receiver (`.hooks.token`), not an agent hook. Recurring or reactive behavior for OpenClaw agents is expressed as a heartbeat step (edit the template `HEARTBEAT.md`) or an infinite task — not a `.hooks` edit. Do not invent a `.hooks` lifecycle schema.
- **Loops (recurring work)** — schedule via the infinite-task mechanism: `bash /opt/agentteams/agent/skills/task-management/scripts/manage-state.sh --action add-infinite --task-id <id> --title <t> --assigned-to <worker> --room-id <room> --schedule <CRON> --timezone <tz> --next-scheduled-at <ISO>`. Your heartbeat already triggers and records these.

## Record the integration

After applying, log what changed and when to `memory/YYYY-MM-DD.md`, and bump the relevant template so the next upgrade propagates it.
