# Anticheat Audit Migration

This migration adds nullable compensation metadata to `cheat_audit_logs` and keeps legacy audit queries compatible.

## Added Columns

- `compensation_amount`
- `compensation_status`
- `compensation_message`
- `compensation_note`
- `compensation_date`
- `approval_note`

The migration also ensures indexes exist for compensation amount, status, and date.

## Apply

The application runs anticheat migrations during normal startup through `database.MigrateCheatTables`.

For local verification:

```bash
node scripts/run-backend-tests.js ./backend/database
```

## Backward Compatibility

Existing audit rows keep compensation columns as `NULL`. Existing time-range queries continue to work because the new fields are nullable and the query builder only filters compensation fields when a filter is supplied.

## Rollback

Preferred rollback is application-level:

1. Disable compensation in config with `unban.enabled: false`.
2. Redeploy the previous application version if needed.
3. Keep the new audit columns in place for forensic continuity.

Physical column rollback is not recommended after production writes. If a database rollback is mandatory, first export `cheat_audit_logs`, then drop the six added columns in a maintenance window, and finally redeploy a build that does not read or write compensation fields.
