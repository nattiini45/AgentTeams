#!/bin/bash
# run-weekly.sh - better-harness weekly self-review evidence collection with a
# 7-day cadence guard.
#
# Exits 0 with "SKIP" when fewer than 7 days have elapsed since the last run,
# when the feature is disabled, or when the CLI is unavailable. Otherwise it
# collects a deterministic better-harness evidence envelope (no LLM) into the
# agent's output dir and stamps state.json with last_run_at. The agent then
# reads the evidence and its better-harness skill to author findings and report
# them to the Manager.
#
# Environment:
#   AGENTTEAMS_BETTER_HARNESS_ENABLED  - "0"/"false" disables the review (default: on)
#   HOME                               - agent workspace (state + evidence output live here)
#   BETTER_HARNESS_BIN                 - override path to the better-harness CLI
#   AGENT_WORKSPACE                    - workspace to scan (default: HOME)

set -u

log() {
    echo "[better-harness $(date '+%Y-%m-%d %H:%M:%S')] $1"
}

# Feature gate (fleet-wide kill switch from the controller / Helm).
case "${AGENTTEAMS_BETTER_HARNESS_ENABLED:-1}" in
    0|false|False|FALSE|no|NO)
        log "SKIP: better-harness disabled via AGENTTEAMS_BETTER_HARNESS_ENABLED"
        exit 0
        ;;
esac

BH_BIN="${BETTER_HARNESS_BIN:-better-harness}"
if ! command -v "${BH_BIN}" >/dev/null 2>&1; then
    log "SKIP: better-harness CLI not found on PATH"
    exit 0
fi

OUT_DIR="${HOME:-/root}/.better-harness"
STATE_FILE="${OUT_DIR}/state.json"
SOURCE_FILE="${OUT_DIR}/report.source.json"
WORKSPACE="${AGENT_WORKSPACE:-${HOME:-/root}}"
mkdir -p "${OUT_DIR}"

# 7-day cadence guard (deterministic; does not rely on the agent to compute dates).
NOW=$(date -u +%s)
LAST=0
if [ -f "${STATE_FILE}" ]; then
    LAST=$(sed -n 's/.*"last_run_at"[[:space:]]*:[[:space:]]*\([0-9][0-9]*\).*/\1/p' "${STATE_FILE}" | head -1)
    LAST="${LAST:-0}"
fi
ELAPSED=$((NOW - LAST))
SEVEN_DAYS=$((7 * 24 * 3600))
if [ "${ELAPSED}" -lt "${SEVEN_DAYS}" ]; then
    log "SKIP: last run $((ELAPSED / 86400)) day(s) ago (< 7)"
    exit 0
fi

# Collect the deterministic evidence envelope (read-only, no LLM). Unobserved
# dimensions stay explicit in the source; the agent resolves them with its skill.
log "Collecting better-harness evidence for workspace ${WORKSPACE} -> ${SOURCE_FILE}"
if "${BH_BIN}" harness source --workspace "${WORKSPACE}" --source "${SOURCE_FILE}" --language en --json >/dev/null 2>&1 \
   && [ -f "${SOURCE_FILE}" ]; then
    printf '{\n  "last_run_at": %s,\n  "last_run_at_iso": "%s",\n  "source": "%s"\n}\n' \
        "${NOW}" "$(date -u -d "@${NOW}" '+%Y-%m-%dT%H:%M:%SZ' 2>/dev/null || date -u '+%Y-%m-%dT%H:%M:%SZ')" \
        "${SOURCE_FILE}" > "${STATE_FILE}"
    log "READY: evidence collected at ${SOURCE_FILE}. Run your better-harness review and report findings to your coordinator (Manager)."
else
    log "WARNING: better-harness evidence collection failed; state not stamped (will retry next cycle)"
fi
