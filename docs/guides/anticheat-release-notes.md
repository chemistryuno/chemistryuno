# Anticheat Panel Release Notes

## Summary

This release adds a unified anticheat admin panel, player-facing transparency stats, appeal compensation tracking, and audit export support.

## Admin Changes

- New `/admin/anticheat` panel with Detection, Appeals, Configuration, and Audit tabs.
- Appeal approvals can include fuel compensation amount and player-facing message.
- Audit log includes compensation amount, status, message, note, date, and approval note.
- Audit export includes compensation columns.

## Player Changes

- Profile dashboard shows:
  - `Bans Today`
  - `System Running`
- Stats refresh every 5 minutes and use cached values during rate limiting.

## Configuration

New `backend/config/anticheat.yaml` keys:

```yaml
unban:
  enabled: true
  compensation_amount: 100
  default_message: "由于反作弊系统将您误封，在此，ChemistryUNO开发组向受到影响的研究员提供燃素补偿，感谢研究员对维护纯净游戏环境做出的贡献"
  message_max_length: 500
  min_amount: 1
  max_amount: 10000
  idempotency_ttl: 60
```

## Deployment Notes

- Database migration is backward-compatible and adds nullable audit columns.
- Redis is used for approval idempotency and player stats caching. Database compensation records still prevent duplicate fuel issuance if Redis is unavailable.
- No `ENABLE_ANTICHEAT_PANEL` feature flag is currently wired in code; rollout is controlled by route permissions and deployment.
