# Failed Compensation Runbook

Use this runbook when an appeal is approved but `compensation_status` is `failed`.

## Detect

1. Open `/admin/anticheat`.
2. Go to Audit.
3. Filter `compensation_status` to `failed`.
4. Confirm player ID, appeal ID, amount, message, and failure note.

## Triage

Check the failure note first:

- Amount validation failure: confirm `unban.min_amount`, `unban.max_amount`, and the submitted amount.
- User/account lookup failure: confirm the player account still exists and is not in an incompatible state.
- Database write failure: check application logs and database health.
- Redis issue: retry is allowed because failed compensation deletes the idempotency key.

## Recover

Preferred path:

1. Reopen or retry approval through the admin workflow with a valid amount and message.
2. Confirm the audit row changes to `ok`.
3. Confirm the player fuel balance changed once.

Manual path, only if the admin workflow cannot recover:

1. Add fuel through the existing repository/admin operational path with a unique compensation ID.
2. Record the manual action in the incident notes.
3. Do not edit old audit rows directly; preserve the failed record for traceability.

## Verify

Run or inspect:

```bash
node scripts/run-backend-tests.js ./backend/cache ./backend/anticheat ./backend/repository ./backend/database
```

For duplicate-payment concerns, check `fuel_compensation_records` for the appeal compensation ID, usually `appeal_{appeal_id}`.
