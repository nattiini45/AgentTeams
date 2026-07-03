#!/bin/bash
# test-create-project.sh
# Static harness for create-project.sh's federation flags (--team / --repo,
# decision #16). PATH-shims curl, mc, AND hiclaw so nothing touches a real
# Matrix/MinIO/controller — create-project.sh curls Matrix and runs
# `mc mirror` unconditionally, and (when --team is given) shells out to
# `hiclaw apply -f`, so all three must be stubbed for the harness to run
# offline. Every shimmed call is appended to a call-log file the assertions
# inspect.
#
# Usage: bash manager/tests/test-create-project.sh

set -uo pipefail

PASS=0
FAIL=0
TMPDIR_ROOT=$(mktemp -d)
trap 'rm -rf "${TMPDIR_ROOT}"' EXIT

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
CREATE_PROJECT_SCRIPT="${PROJECT_ROOT}/manager/agent/skills/project-management/scripts/create-project.sh"

pass() { echo "  PASS: $1"; PASS=$((PASS + 1)); }
fail() { echo "  FAIL: $1"; echo "       expected: $2"; echo "       got:      $3"; FAIL=$((FAIL + 1)); }

assert_eq() {
    local desc="$1" expected="$2" actual="$3"
    if [ "${expected}" = "${actual}" ]; then
        pass "${desc}"
    else
        fail "${desc}" "${expected}" "${actual}"
    fi
}

assert_contains() {
    local desc="$1" needle="$2" haystack="$3"
    if printf '%s' "${haystack}" | grep -qF -- "${needle}"; then
        pass "${desc}"
    else
        fail "${desc}" "contains '${needle}'" "not found: ${haystack}"
    fi
}

assert_not_contains() {
    local desc="$1" needle="$2" haystack="$3"
    if ! printf '%s' "${haystack}" | grep -qF -- "${needle}"; then
        pass "${desc}"
    else
        fail "${desc}" "should NOT contain '${needle}'" "found it"
    fi
}

# ── CRLF strip (checkout is autocrlf=true) ────────────────────────────────────
strip_cr() {
    local f="$1"
    sed -i 's/\r$//' "${f}" 2>/dev/null || true
}
strip_cr "${CREATE_PROJECT_SCRIPT}"

# ── Stub /opt/hiclaw/scripts/lib/hiclaw-env.sh ────────────────────────────────
# create-project.sh sources this unconditionally (the real path only exists
# inside a built Manager container). Provide a harmless stub so the harness
# can run standalone; real containers always have the actual file.
HICLAW_ENV_STUB_DIR="/opt/hiclaw/scripts/lib"
if [ ! -f "${HICLAW_ENV_STUB_DIR}/hiclaw-env.sh" ]; then
    if mkdir -p "${HICLAW_ENV_STUB_DIR}" 2>/dev/null; then
        cat > "${HICLAW_ENV_STUB_DIR}/hiclaw-env.sh" << 'ENVSTUB'
# Test-harness stub — real containers ship the actual hiclaw-env.sh.
: "${HICLAW_STORAGE_PREFIX:=hiclaw/hiclaw-storage}"
: "${HICLAW_MATRIX_URL:=http://matrix.fake.local}"
ensure_mc_credentials() { :; }
log() { echo "[hiclaw-env-stub] $*" >&2; }
ENVSTUB
    else
        echo "SKIP: cannot write ${HICLAW_ENV_STUB_DIR}/hiclaw-env.sh (no permission) — run this harness as root/in a container" >&2
        exit 1
    fi
fi

# ── Fake bin dir: PATH-shimmed curl / mc / hiclaw / jq(passthrough) ───────────
# Each test case gets its own sandbox so shim state never leaks between cases.
new_sandbox() {
    local sandbox
    sandbox=$(mktemp -d "${TMPDIR_ROOT}/sandbox-XXXXXX")
    mkdir -p "${sandbox}/bin" "${sandbox}/home" "${sandbox}/root/hiclaw-fs/shared/projects" \
        "${sandbox}/root/hiclaw-fs/agents/manager" "${sandbox}/data/worker-creds"
    : > "${sandbox}/calls.log"
    echo "${sandbox}"
}

# The curl shim: logs every call ("METHOD URL BODY") and returns plausible
# canned JSON so the script's Matrix flow completes without a network.
write_curl_shim() {
    local sandbox="$1"
    cat > "${sandbox}/bin/curl" << 'CURLSHIM'
#!/bin/bash
CALL_LOG="${FAKE_CALL_LOG:?FAKE_CALL_LOG not set}"

METHOD="GET"
URL=""
BODY=""

args=("$@")
i=0
n=${#args[@]}
while [ $i -lt $n ]; do
    a="${args[$i]}"
    case "$a" in
        -X)
            i=$((i+1)); METHOD="${args[$i]}"
            ;;
        -d|--data)
            i=$((i+1)); BODY="${args[$i]}"
            ;;
        -H|-b)
            i=$((i+1))
            ;;
        -s|-sf|-f|-sS)
            ;;
        http*://*)
            URL="$a"
            ;;
        *)
            ;;
    esac
    i=$((i+1))
done

echo "${METHOD} ${URL} ${BODY}" >> "${CALL_LOG}"

case "${URL}" in
    */_matrix/client/v3/login)
        echo '{"access_token":"faketoken-123"}'
        ;;
    */_matrix/client/v3/createRoom)
        echo '{"room_id":"!fakeroom:matrix.fake.local"}'
        ;;
    *)
        echo '{}'
        ;;
esac
exit 0
CURLSHIM
    chmod +x "${sandbox}/bin/curl"
}

# The mc shim: cp/mirror/stat/cat are all no-op successes (stat always finds
# the meta.json the script just wrote, since create-project.sh never checks
# mc's own storage — only that the exit code is 0).
write_mc_shim() {
    local sandbox="$1"
    cat > "${sandbox}/bin/mc" << 'MCSHIM'
#!/bin/bash
CALL_LOG="${FAKE_CALL_LOG:?FAKE_CALL_LOG not set}"
echo "mc $*" >> "${CALL_LOG}"
exit 0
MCSHIM
    chmod +x "${sandbox}/bin/mc"
}

# The hiclaw shim: logs the full invocation plus the applied YAML's content
# (so assertions can inspect what create-project.sh generated), and always
# succeeds unless FAKE_HICLAW_FAIL=1 is set for a test case.
write_hiclaw_shim() {
    local sandbox="$1"
    cat > "${sandbox}/bin/hiclaw" << 'HICLAWSHIM'
#!/bin/bash
CALL_LOG="${FAKE_CALL_LOG:?FAKE_CALL_LOG not set}"
echo "hiclaw $*" >> "${CALL_LOG}"

# Find the -f/--file argument and dump the applied YAML into the log too.
prev=""
for a in "$@"; do
    if [ "${prev}" = "-f" ]; then
        echo "---APPLIED-YAML-START---" >> "${CALL_LOG}"
        cat "${a}" >> "${CALL_LOG}"
        echo "---APPLIED-YAML-END---" >> "${CALL_LOG}"
    fi
    prev="${a}"
done

if [ "${FAKE_HICLAW_FAIL:-0}" = "1" ]; then
    echo "hiclaw: simulated failure" >&2
    exit 1
fi
echo "  project/fake created"
exit 0
HICLAWSHIM
    chmod +x "${sandbox}/bin/hiclaw"
}

run_create_project() {
    # run_create_project <sandbox> <args...>
    local sandbox="$1"; shift
    HOME="${sandbox}/home" \
    PATH="${sandbox}/bin:${PATH}" \
    FAKE_CALL_LOG="${sandbox}/calls.log" \
    FAKE_HICLAW_FAIL="${FAKE_HICLAW_FAIL:-0}" \
    HICLAW_MATRIX_DOMAIN="matrix.fake.local" \
    HICLAW_ADMIN_USER="admin" \
    HICLAW_MATRIX_URL="http://matrix.fake.local" \
    HICLAW_STORAGE_PREFIX="hiclaw/hiclaw-storage" \
    MANAGER_MATRIX_TOKEN="preset-token" \
    bash "${CREATE_PROJECT_SCRIPT}" "$@" 2>&1
}

calls() {
    cat "$1/calls.log"
}

project_dir() {
    # project_dir <sandbox> <project-id> — create-project.sh hardcodes
    # /root/hiclaw-fs/shared/projects/<id>; on this box that resolves under
    # the real filesystem root, not the sandbox, so tests read from there and
    # clean up afterward.
    echo "/root/hiclaw-fs/shared/projects/$2"
}

cleanup_real_project_dir() {
    rm -rf "/root/hiclaw-fs/shared/projects/$1" 2>/dev/null || true
}

# ── Tests ─────────────────────────────────────────────────────────────────────

echo ""
echo "=== TC1: bash -n on create-project.sh ==="
{
    if bash -n "${CREATE_PROJECT_SCRIPT}"; then
        pass "create-project.sh parses (bash -n)"
    else
        fail "create-project.sh parses (bash -n)" "exit 0" "non-zero"
    fi
}

echo ""
echo "=== TC2: no-flags invocation — byte-identical to today (no hiclaw call at all) ==="
{
    s=$(new_sandbox)
    write_curl_shim "${s}"
    write_mc_shim "${s}"
    write_hiclaw_shim "${s}"
    pid="proj-tc2-$$"
    cleanup_real_project_dir "${pid}"

    out=$(run_create_project "${s}" --id "${pid}" --title "No Flags Project" --workers "alice,bob")
    log=$(calls "${s}")

    assert_contains "prints ---RESULT---" "---RESULT---" "${out}"
    assert_contains "result has project_id" "\"project_id\": \"${pid}\"" "${out}"
    assert_not_contains "hiclaw is NEVER invoked without --team" "hiclaw apply" "${log}"

    pd=$(project_dir "${s}" "${pid}")
    if [ -f "${pd}/meta.json" ]; then
        pass "meta.json written"
    else
        fail "meta.json written" "file exists" "missing"
    fi
    if [ -f "${pd}/project-cr.yaml" ]; then
        fail "no project-cr.yaml without --team" "absent" "present"
    else
        pass "no project-cr.yaml without --team"
    fi
    cleanup_real_project_dir "${pid}"
}

echo ""
echo "=== TC3: --team without --repo is rejected before any side effects ==="
{
    s=$(new_sandbox)
    write_curl_shim "${s}"
    write_mc_shim "${s}"
    write_hiclaw_shim "${s}"
    pid="proj-tc3-$$"
    cleanup_real_project_dir "${pid}"

    out=$(run_create_project "${s}" --id "${pid}" --title "Team No Repo" --workers "alice" --team "teamA")
    log=$(calls "${s}")

    assert_contains "error mentions --repo requirement" "--team requires at least one --repo" "${out}"
    assert_eq "no Matrix/MinIO/hiclaw calls happened" "" "${log}"
    if [ -f "$(project_dir "${s}" "${pid}")/meta.json" ]; then
        fail "no meta.json written on validation failure" "absent" "present"
    else
        pass "no meta.json written on validation failure"
    fi
    cleanup_real_project_dir "${pid}"
}

echo ""
echo "=== TC4: --repo without --team is rejected ==="
{
    s=$(new_sandbox)
    write_curl_shim "${s}"
    write_mc_shim "${s}"
    write_hiclaw_shim "${s}"
    pid="proj-tc4-$$"
    cleanup_real_project_dir "${pid}"

    out=$(run_create_project "${s}" --id "${pid}" --title "Repo No Team" --workers "alice" \
        --repo "https://git.fake.local/teamA/repo.git:rw")

    assert_contains "error mentions --team requirement" "--repo requires --team" "${out}"
    cleanup_real_project_dir "${pid}"
}

echo ""
echo "=== TC5: --repo access enum is enforced (rejects anything but rw/ro) ==="
{
    s=$(new_sandbox)
    write_curl_shim "${s}"
    write_mc_shim "${s}"
    write_hiclaw_shim "${s}"
    pid="proj-tc5-$$"
    cleanup_real_project_dir "${pid}"

    out=$(run_create_project "${s}" --id "${pid}" --title "Bad Access" --workers "alice" \
        --team "teamA" --repo "https://git.fake.local/teamA/repo.git:admin")

    assert_contains "error mentions rw or ro" "--repo access must be rw or ro" "${out}"
    cleanup_real_project_dir "${pid}"
}

echo ""
echo "=== TC6: --team + --repo emits a Project CR and calls hiclaw apply -f ==="
{
    s=$(new_sandbox)
    write_curl_shim "${s}"
    write_mc_shim "${s}"
    write_hiclaw_shim "${s}"
    pid="proj-tc6-$$"
    cleanup_real_project_dir "${pid}"

    out=$(run_create_project "${s}" --id "${pid}" --title "Federated Project" --workers "alice,bob" \
        --team "teamA" \
        --repo "https://git.fake.local/teamA/repo-rw.git:rw" \
        --repo "https://git.fake.local/teamA/repo-ro.git:ro")
    log=$(calls "${s}")

    assert_contains "hiclaw apply -f was invoked" "hiclaw apply -f" "${log}"
    assert_contains "applied YAML has apiVersion" "apiVersion: hiclaw.io/v1beta1" "${log}"
    assert_contains "applied YAML has kind: Project" "kind: Project" "${log}"
    assert_contains "applied YAML names the project" "name: ${pid}" "${log}"
    assert_contains "applied YAML sets team" "team: teamA" "${log}"
    assert_contains "applied YAML has rw repo" "url: https://git.fake.local/teamA/repo-rw.git" "${log}"
    assert_contains "applied YAML marks rw access" "access: rw" "${log}"
    assert_contains "applied YAML has ro repo" "url: https://git.fake.local/teamA/repo-ro.git" "${log}"
    assert_contains "applied YAML marks ro access" "access: ro" "${log}"
    assert_contains "applied YAML lists workers" "- alice" "${log}"

    # Chat-flow layer still written, untouched, alongside the CR.
    pd=$(project_dir "${s}" "${pid}")
    if [ -f "${pd}/meta.json" ]; then
        pass "meta.json still written alongside the CR"
    else
        fail "meta.json still written alongside the CR" "file exists" "missing"
    fi
    if [ -f "${pd}/project-cr.yaml" ]; then
        pass "project-cr.yaml written to the project directory"
    else
        fail "project-cr.yaml written to the project directory" "file exists" "missing"
    fi
    assert_contains "result JSON still has project_id (unchanged shape)" "\"project_id\": \"${pid}\"" "${out}"
    cleanup_real_project_dir "${pid}"
}

echo ""
echo "=== TC7: hiclaw apply failure does not fail the whole script (chat-flow project still created) ==="
{
    s=$(new_sandbox)
    write_curl_shim "${s}"
    write_mc_shim "${s}"
    write_hiclaw_shim "${s}"
    pid="proj-tc7-$$"
    cleanup_real_project_dir "${pid}"

    FAKE_HICLAW_FAIL=1 out=$(FAKE_HICLAW_FAIL=1 run_create_project "${s}" --id "${pid}" --title "Apply Fails" \
        --workers "alice" --team "teamA" --repo "https://git.fake.local/teamA/repo.git:rw")

    assert_contains "warns about the apply failure" "WARNING: hiclaw apply -f failed" "${out}"
    assert_contains "still prints the RESULT block" "---RESULT---" "${out}"
    pd=$(project_dir "${s}" "${pid}")
    if [ -f "${pd}/meta.json" ]; then
        pass "chat-flow meta.json unaffected by CR-apply failure"
    else
        fail "chat-flow meta.json unaffected by CR-apply failure" "file exists" "missing"
    fi
    cleanup_real_project_dir "${pid}"
}

# ── Summary ───────────────────────────────────────────────────────────────────
echo ""
echo "================================"
echo "Results: ${PASS} passed, ${FAIL} failed"
echo "================================"
[ "${FAIL}" -eq 0 ]
