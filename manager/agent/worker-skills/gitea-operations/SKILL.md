---
name: gitea-operations
description: Manage Gitea pull requests, issues, repos, and forge metadata via your per-worker gitea-mcp server. Use for PR/issue CRUD and review state — not for native git clone/push (use your own PAT helper for that).
assign_when: You are a Hermes or coding Worker with your own Gitea identity (per-worker mcp-gitea-* server and scoped PAT), and you need forge CRUD — PRs, issues, repos, labels — on git.pawcommit.com or the team Gitea instance (not GitHub)
---

# Gitea Forge Operations

## Overview

Use this skill to manage **Gitea forge objects** (pull requests, issues, repos, branches, releases, and related metadata) through your dedicated MCP server. For native git operations (clone, pull, commit, push in a working tree), use your **own scoped PAT** credential helper (S-GIT) — gitea-mcp is API-only and has no working tree.

Do **not** use the shared `mcp-github` server for Gitea work. Your server selector is **`mcp-gitea-<your-worker-name>`** (from your `mcporter-servers.json` / `mcporter.json`).

## Split vs Other Skills

| Task | Use This Skill | Use git-delegation / own PAT |
|------|----------------|------------------------------|
| Create/update PR on Gitea | ✅ | |
| Post Gitea review state | ✅ | |
| Create/update/close Issue | ✅ | |
| Repo/branch/tag/release CRUD | ✅ | |
| Clone Gitea repository | | ✅ (own PAT helper) |
| Pull/fetch/commit/push locally | | ✅ (own PAT helper) |

CoPaw Workers on GitHub may still use `github-operations` + `git-delegation`. Hermes coding Workers on Gitea use **this skill** for forge CRUD and **own PAT** for native checkout.

## How to Call Gitea Tools

Use `mcporter` with your per-worker server name:

```bash
# Method 1: key=value (simple args)
mcporter call mcp-gitea-<your-name>.<TOOL_NAME> key1=value1 key2=value2

# Method 2: JSON via --args (complex objects)
mcporter call mcp-gitea-<your-name>.<TOOL_NAME> --args '{"key1":"value1"}'
```

**IMPORTANT:**
- Replace `<your-name>` with your Worker name (e.g. `mcp-gitea-engineer-backend-architect`)
- Use `mcp-gitea-<your-name>.<tool_name>` selector format
- Auth is per-worker PAT via Higress — your coordinator provisions it at create time

Parse JSON output with `jq` when you need structured fields.

## Tool Catalog (53 tools — names only)

All tools use method-dispatch names below. This section lists **names only** — per-tool JSON schemas are deferred to a follow-up.

### User / organization

- `get_me`
- `get_user_orgs`
- `search_users`
- `search_org_teams`
- `list_org_repos`

### Repository

- `create_repo`
- `fork_repo`
- `list_my_repos`
- `search_repos`
- `get_repository_tree`

### Branch, tag, commit

- `create_branch`
- `delete_branch`
- `list_branches`
- `create_tag`
- `delete_tag`
- `get_tag`
- `list_tags`
- `list_commits`
- `get_commit`

### File (API)

- `get_file_contents`
- `get_dir_contents`
- `create_or_update_file`
- `delete_file`

### Issue

- `issue_read`
- `issue_write`
- `list_issues`
- `search_issues`

### Pull request

- `pull_request_read`
- `pull_request_write`
- `pull_request_review_write`
- `list_pull_requests`

### Label / milestone

- `label_read`
- `label_write`
- `milestone_read`
- `milestone_write`

### Release

- `create_release`
- `delete_release`
- `get_release`
- `get_latest_release`
- `list_releases`

### Actions (CI config / runs)

- `actions_config_read`
- `actions_config_write`
- `actions_run_read`
- `actions_run_write`

### Notifications

- `notification_read`
- `notification_write`

### Packages

- `package_read`
- `package_write`

### Time tracking

- `timetracking_read`
- `timetracking_write`

### Wiki

- `wiki_read`
- `wiki_write`

### Meta

- `get_gitea_mcp_server_version`

## Typical Worker Workflow

1. **Implement locally** — clone/checkout with your own PAT helper (S-GIT), commit, push branch.
2. **Open PR** — `pull_request_write` (this skill). Note `pr_number` and `head_sha` from the response.
3. **Report to coordinator** — include PR URL/number and review state: `open`, `changes-requested`, or `merged`.
4. **After review feedback** — push fixes (PAT helper), update PR if needed via `pull_request_write`.

Your Team Leader owns the **rich review verdict** via lifecycle-mcp (`set_review_verdict` / `get_review_verdict`). You report forge-visible state; the lead stores orchestration verdict and posts Gitea review via `pull_request_review_write`.

## Result Contract

When you complete a coding task that produces a PR, report:

- Repo (`owner/name`)
- PR number
- Current `head_sha` (if known)
- Review state: `open` | `changes-requested` | `merged`

Do not claim merge approval yourself — the lead runs `safe_merge` on lifecycle-mcp after verdict + CI gates.

## Important Notes

- **Transport**: HTTP MCP via Higress; Bearer token is your scoped PAT (auto-configured in mcporter).
- **Isolation**: Each Worker has its own Gitea user/PAT — PRs you open appear under your identity.
- **No forge CRUD on lifecycle-mcp** — use this skill (gitea-mcp) for all Gitea API operations.
- **Rate limits**: If you get 403/429, wait and retry; ask your coordinator if auth fails persistently.
