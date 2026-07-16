# Migrating a HiClaw fork to AgentTeams contracts

This release changes the active Kubernetes API group from
hiclaw.io/v1beta1 to agentteams.io/v1beta1. The migration is deliberately
maintenance-mode and reversible: old resources and CRDs are retained until a
later release.

## Before the maintenance window

1. Pin the old controller and runtime image tags.
2. Export all legacy Manager, Worker, Team, Human, and Project resources to
   hiclaw-resources-backup.yaml.
3. Back up the object-storage bucket containing agents/ and shared/.
4. Configure the dashboard password Secret before exposing its Service.

## Cutover

1. Stop the old controller so it cannot reconcile while converted resources
   are created.
2. Install the agentteams.io CRDs without deleting the legacy CRDs.
3. Preview the conversion and retain the JSON report:

       hiclaw migrate-api-group --dry-run --output migration-report.json

4. Apply the same conversion:

       hiclaw migrate-api-group --apply --output migration-report.json

5. Start the AgentTeams controller and verify that Manager, Worker, Team,
   Human, and Project counts match the report. Verify Project manifests and a
   sample of agents/ and shared/ objects before enabling the dashboard.

The migration command is idempotent. It creates or updates the
agentteams.io counterpart, records conflicts in the report, and never deletes
the legacy object.

## Compatibility window and rollback

AGENTTEAMS_* variables take precedence. Deprecated HICLAW_* aliases remain
accepted for one release and emit warnings.

To roll back during the compatibility window, stop the new controller, restore
the previous images, and restart the old controller against the retained
hiclaw.io resources. Restore object storage only if post-cutover validation
found data corruption; normal rollback does not require deleting new objects.

Do not delete legacy CRDs, legacy resources, previous images, or the storage
backup until the compatibility release has completed successfully.
