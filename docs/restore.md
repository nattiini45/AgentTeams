# AgentTeams Snapshot Restore Runbook (S-BACKUP)

Checkout-only operator runbook for backing up and restoring an embedded or Docker Compose AgentTeams stack. Live VPS cron timing under `/root/backups/hiclaw-snapshot-*` still needs Phase 0 validation — **live timing TBD**.

## Backup Units

| Component | Backup | Restore | Notes |
|-----------|--------|---------|-------|
| **gitea-mcp** | None | N/A | Stateless; PAT/config live in Higress / provision script |
| **lifecycle-mcp** | `docker cp lifecycle-mcp:/data/ops.db ./ops.db.bak` (+ `-wal`/`-shm` if present) | Stop service → `docker cp` back → start | SQLite WAL; hot-copy of `ops.db` is acceptable for this runbook |
| **AgentTeams data volume** | `tar czf` of `${AGENTTEAMS_DATA_DIR:-agentteams-data}` (docs historically say `hiclaw-data`) | Fresh volume + `tar xzf` into `/data` | Holds Tuwunel + MinIO (`hiclaw-storage`) + kine + Higress |
| **Install env** | Copy `hiclaw-manager.env` / install env file beside snapshot | Restore before container recreate | Required so secrets/domains match |
| **Higress PATs** | Contained in data volume Higress config | Same as volume restore | Plaintext PAT risk in archives — see Security |

## 1. Preflight

Before backup or restore:

1. Stop Manager and Workers (embedded: `make uninstall-embedded` or stop containers; keep volumes unless doing a throwaway restore).
2. Note the data volume name:

   ```bash
   docker volume ls | grep -E 'hiclaw|agentteams'
   ```

3. Record Element Web URL, admin password, and Manager Matrix MXID (from install output or Element settings).
4. For lifecycle-mcp (if running on the host/VPS), note the container name (`lifecycle-mcp`).

## 2. Backup

Create a dated snapshot directory, e.g. `hiclaw-snapshot-YYYY-MM-DD`:

### Data volume

```bash
SNAPSHOT=hiclaw-snapshot-$(date +%Y-%m-%d)
mkdir -p "$SNAPSHOT"

# Replace hiclaw-data with your volume name if different
docker run --rm \
  -v hiclaw-data:/data:ro \
  -v "$(pwd)/$SNAPSHOT":/backup \
  ubuntu tar czf /backup/agentteams-data.tar.gz -C / data
```

### Install environment

```bash
cp /path/to/hiclaw-manager.env "$SNAPSHOT/"   # or your install env file path
```

### lifecycle-mcp SQLite (optional, if orchestration state matters)

```bash
docker cp lifecycle-mcp:/data/ops.db "$SNAPSHOT/ops.db.bak"
# If present:
docker cp lifecycle-mcp:/data/ops.db-wal "$SNAPSHOT/" 2>/dev/null || true
docker cp lifecycle-mcp:/data/ops.db-shm "$SNAPSHOT/" 2>/dev/null || true
```

**gitea-mcp**: no backup — redeploy/re-register via Higress provision scripts after restore.

## 3. Full Restore (throwaway volume)

Use this when replacing a corrupted volume or migrating to a fresh host.

1. Stop the stack and remove the old volume (or create a new volume name).
2. Create empty volume and extract:

   ```bash
   docker volume create hiclaw-data
   docker run --rm \
     -v hiclaw-data:/data \
     -v "$(pwd)/$SNAPSHOT":/backup \
     ubuntu tar xzf /backup/agentteams-data.tar.gz -C /
   ```

3. Restore install env file **before** recreating containers (`hiclaw-manager.env` or equivalent).
4. Start embedded stack (`make install-embedded` or your usual bring-up).
5. **Verify**:

   | Check | How |
   |-------|-----|
   | Manager MXID / login | Element Web |
   | Rooms intact | Element room list |
   | MinIO `agents/` + `shared/` | `mc ls` from Manager container |
   | Higress consumers + `mcp-gitea-*` | Higress Console or `setup-mcp-proxy.sh` audit |
   | lifecycle-mcp state | Restore `ops.db` if backed up; smoke `get_review_verdict` / `list_ops` via mcporter |

### lifecycle-mcp restore

```bash
docker stop lifecycle-mcp
docker cp "$SNAPSHOT/ops.db.bak" lifecycle-mcp:/data/ops.db
# Optional WAL/SHM if you copied them:
# docker cp "$SNAPSHOT/ops.db-wal" lifecycle-mcp:/data/ops.db-wal
docker start lifecycle-mcp
```

## 4. Selective Import

When migrating personas/workspaces without cloning controller/kine state:

- Restore **only** MinIO prefixes you need (e.g. `agents/<worker>/`, `shared/tasks/`, `shared/projects/`).
- Use `mc mirror` or selective `mc cp` into a fresh volume's MinIO data path.
- **Do not** blanket-restore an old kine DB when migrating teams (Phase 0 step 4) — CR/controller state may not match.

## 5. What Is Not Covered

- **gitea-mcp** — stateless; re-provision PAT routes after restore.
- **Live VPS automated backups** — cron format and retention under `/root/backups/hiclaw-snapshot-*` require Phase 0 live validation (checkout-only; timing TBD).
- **Per-tool MCP schema backup** — not applicable.

## 6. Security

- Snapshot archives contain Higress `defaultCredential` PATs and Matrix/MinIO secrets if the data volume and env file are included.
- Encrypt archives at rest or store offline with strict access control.
- Exclude env files from shared backups when possible; inject secrets via a secrets manager on restore.
- Treat `ops.db` as sensitive — it holds review verdicts and orchestration state.

## Quick Reference (embedded install)

Historical volume name: `hiclaw-data`. Newer installs may use `agentteams-data` — always confirm with `docker volume ls`.

Minimal one-liner backup (data volume only):

```bash
docker run --rm -v hiclaw-data:/data -v "$(pwd)":/backup ubuntu \
  tar czf /backup/hiclaw-backup-$(date +%Y%m%d).tar.gz /data
```

See also [manager-guide.md](manager-guide.md) for high-level backup pointers.
