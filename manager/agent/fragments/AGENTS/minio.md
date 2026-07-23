## MinIO Storage

- **Local mirror:** `/root/agentteams-fs/` — your local filesystem, NOT automatically synced
- **MinIO prefix:** always use `${AGENTTEAMS_STORAGE_PREFIX}` in mc commands (this env var is pre-set in your shell, format: `<mc-alias>/<bucket>`)
- **Example:** `mc mirror ${AGENTTEAMS_STORAGE_PREFIX}/shared/tasks/{task-id}/ /root/agentteams-fs/shared/tasks/{task-id}/ --overwrite`
- **NEVER guess or hardcode the prefix** — do NOT use `agentteams-fs/...`, `agentteams-storage/...`, or any literal path. Always use `${AGENTTEAMS_STORAGE_PREFIX}`. If unsure, run `echo $AGENTTEAMS_STORAGE_PREFIX` to check.
