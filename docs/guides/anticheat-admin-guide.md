# Anticheat Admin Guide

This guide covers the operator workflow for the unified anticheat panel at `/admin/anticheat`.

## Access

- Required role: `admin` or `co-worker`.
- Route: `/admin/anticheat`.
- Main tabs: Detection, Appeals, Configuration, Audit.

## Detection

Use the Detection tab to review recent risk scoring output and enforcement decisions.

- Search by player ID or room ID when investigating a report.
- Filter by sanction type: observe, warning, mute, ban.
- Open a detection record to inspect risk score dimensions before confirming or overriding the decision.

## Appeals

Use the Appeals tab for false-positive review and compensation issuance.

1. Filter or search for the player appeal.
2. Click Approve on a pending appeal.
3. Confirm the default compensation amount and message.
4. Expand Adjust Compensation only when the case requires an override.
5. Add an approval note for audit context.
6. Confirm approval.

The approval request submits:

```json
{
  "note": "Reviewed clean replay",
  "compensation_amount": 100,
  "compensation_message": "message sent to player"
}
```

Successful compensation writes `compensation_status: "ok"` to the audit trail. Duplicate approval retries are guarded by Redis key `unban_compensation:{user_id}:{event_id}` and the database compensation record.

## Configuration

Use the Configuration tab to tune compensation defaults without a code change.

- `unban.enabled`: enables compensation behavior.
- `unban.compensation_amount`: default fuel amount.
- `unban.default_message`: default player-facing message, max 500 characters.
- `unban.min_amount` and `unban.max_amount`: accepted range.
- `unban.idempotency_ttl`: Redis duplicate-approval window in minutes.

Changes take effect immediately for newly opened approval modals and new approval requests.

## Audit

Use the Audit tab as the read-only ledger for enforcement and compensation outcomes.

- Filter by player ID, date range, action type, and compensation status.
- Export CSV when handing off an investigation or preparing release/support evidence.
- Key compensation states:
  - `pending`: approval recorded before issuance result.
  - `ok`: fuel issued or duplicate request treated as idempotent success.
  - `failed`: appeal approval completed, but compensation needs recovery.

## Recovery

If an approval shows `failed`, follow [Failed Compensation Runbook](failed-compensation-runbook.md). Do not manually edit audit rows; append recovery context through the supported retry or operational notes.
