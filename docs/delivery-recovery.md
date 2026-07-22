# Delivery Recovery Runbook

This runbook documents the recovery route for multi-arch image push failures in the AgentTeams CI/CD pipeline.

## Owner

The release engineer who triggered the build (or the on-call maintainer if the trigger was automated).

## Scenario: Partial Multi-Arch Push Failure

**Symptoms:** `build.yml` fails at step N. Images from steps 1..N-1 are already pushed to the registry; images from step N onward are missing or incomplete.

**Root cause:** Each `make push-<target>` step is independent. A failure in one step (network timeout, build error, registry auth expiry) does not roll back previously pushed images.

## Recovery Procedure

1. **Identify the failed step** from the GitHub Actions build summary (the step marked with a red ✗).

2. **Re-run only the failed targets** via `workflow_dispatch`:
   - Go to Actions → build.yml → "Run workflow"
   - Set `version` to the same tag that failed
   - Set `targets` to only the failed image names (e.g., `worker,copaw-worker`)

   Or locally:
   ```bash
   make push-<failed-target> VERSION=<tag> \
     REGISTRY=higress-registry.cn-hangzhou.cr.aliyuncs.com REPO=higress
   ```

3. **Verify the fix:**
   ```bash
   docker manifest inspect $REGISTRY/$REPO/<image>:$VERSION
   ```
   Confirm both `linux/amd64` and `linux/arm64` platforms are present.

## Postcondition

All images for `VERSION` exist in the registry with complete multi-arch manifests (`linux/amd64` + `linux/arm64`).

## Rollback: Bad Image Pushed

If a defective image is pushed to a stable tag:

1. **Restore the previous version:**
   ```bash
   make push-<target> VERSION=<previous-stable-tag> \
     REGISTRY=... REPO=...
   ```
   This rebuilds and re-pushes from the known-good revision (buildx `--push` overwrites the manifest atomically per image).

2. **Or delete the bad tag** from the registry console if no rebuild is desired.

3. **Re-trigger the release** from the corrected commit.

## Validation Command

```bash
# Check all images for a version
for img in agentteams-manager agentteams-manager-copaw agentteams-worker \
           agentteams-copaw-worker agentteams-hermes-worker agentteams-qwenpaw-worker \
           agentteams-openhuman-worker agentteams-controller agentteams-embedded; do
  echo "--- $img ---"
  docker manifest inspect "higress-registry.cn-hangzhou.cr.aliyuncs.com/higress/${img}:${VERSION}" \
    | jq -r '.manifests[].platform | "\(.os)/\(.architecture)"'
done
```

Expected output per image: `linux/amd64` and `linux/arm64`.

## CI Re-run

GitHub Actions → build.yml → **"Re-run failed jobs"** — all push steps are idempotent (buildx overwrites manifests atomically), so re-running a failed job is always safe.

## Image List

| Target | Image |
|--------|-------|
| `push-manager` | `agentteams-manager` |
| `push-manager-copaw` | `agentteams-manager-copaw` |
| `push-worker` | `agentteams-worker` |
| `push-copaw-worker` | `agentteams-copaw-worker` |
| `push-hermes-worker` | `agentteams-hermes-worker` |
| `push-qwenpaw-worker` | `agentteams-qwenpaw-worker` |
| `push-openhuman-worker` | `agentteams-openhuman-worker` |
| `push-hiclaw-controller` | `agentteams-controller` |
| `push-embedded` | `agentteams-embedded` |
