#!/bin/bash
# aggregate-reports.sh - Aggregate weekly better-harness findings from all agents.
#
# Pulls the shared harness-reports prefix from object storage, then merges each
# agent's findings.json into a single consolidated summary on stdout (JSON):
# how many agents reported, the distinct findings (by title), and which agents
# raised each. The Manager reads this to draft one fleet-wide integration plan.
#
# Environment:
#   AGENTTEAMS_STORAGE_PREFIX  - storage prefix (required)
#   AGENTTEAMS_FS              - local mirror root (default: /root/agentteams-fs)

set -u

log() {
    echo "[harness-integration $(date '+%Y-%m-%d %H:%M:%S')] $1" >&2
}

PREFIX="${AGENTTEAMS_STORAGE_PREFIX:-}"
if [ -z "${PREFIX}" ]; then
    log "ERROR: AGENTTEAMS_STORAGE_PREFIX is not set"
    exit 1
fi

if ! command -v jq >/dev/null 2>&1; then
    log "ERROR: jq is required but not found on PATH"
    exit 1
fi

FS_ROOT="${AGENTTEAMS_FS:-/root/agentteams-fs}"
LOCAL_DIR="${FS_ROOT}/shared/harness-reports"
mkdir -p "${LOCAL_DIR}"

# Best-effort pull of the shared reports prefix.
if command -v mc >/dev/null 2>&1; then
    mc mirror "${PREFIX}/shared/harness-reports/" "${LOCAL_DIR}/" --overwrite >/dev/null 2>&1 || true
fi

# Collect every findings file (any agent, any date).
mapfile -t FILES < <(find "${LOCAL_DIR}" -mindepth 2 -maxdepth 2 -name '*.json' 2>/dev/null | sort)
if [ "${#FILES[@]}" -eq 0 ]; then
    jq -cn '{agents:0, reports:0, findings:[]}'
    exit 0
fi

# Merge findings. Each file is expected to be a better-harness findings.json:
# either an array of finding objects or an object with a "findings" array.
# We normalize to {title, agents:[...]} rows, deduping by title.
tmp="$(mktemp)"
trap 'rm -f "${tmp}"' EXIT

printf '[' > "${tmp}"
first=1
for f in "${FILES[@]}"; do
    agent="$(basename "$(dirname "${f}")")"
    # Extract finding titles; tolerate both top-level array and {findings:[...]}.
    titles="$(jq -r 'if type=="array" then .[] elif .findings then .findings[] else empty end | (.title // .summary // .id // "untitled") | @text' "${f}" 2>/dev/null || true)"
    [ -z "${titles}" ] && continue
    while IFS= read -r title; do
        [ -z "${title}" ] && continue
        [ "${first}" -eq 0 ] && printf ',' >> "${tmp}"
        first=0
        jq -cn --arg t "${title}" --arg a "${agent}" '{title:$t, agent:$a}' >> "${tmp}"
    done <<< "${titles}"
done
printf ']' >> "${tmp}"

# Aggregate: distinct titles with the set of agents that raised each.
jq -s '
  (add // []) as $rows
  | ($rows | map(.agent) | unique) as $agents
  | {
      agents: ($agents | length),
      reports: ($rows | length),
      findings: (
        $rows
        | group_by(.title)
        | map({ title: .[0].title, agents: (map(.agent) | unique), count: length })
        | sort_by(-.count)
      )
    }
' "${tmp}"
