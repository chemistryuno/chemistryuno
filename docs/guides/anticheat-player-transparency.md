# Anticheat Player Transparency

Players see a compact anticheat transparency widget on the profile dashboard.

## Visible Metrics

- `Bans Today`: total bans issued in the last 24 hours.
- `System Running`: whole days since the anticheat service startup timestamp.

The widget calls `GET /api/player/anticheat/stats` after the profile loads and refreshes every 5 minutes.

## API Contract

Authenticated players receive:

```json
{
  "bans_today": 2,
  "system_uptime_days": 9
}
```

The endpoint does not expose detection rules, thresholds, player identities, appeal records, or admin-only policy details.

## Freshness And Limits

- Stats are cached for up to 5 minutes.
- Requests are rate-limited to 1 request per second per player.
- If a refresh is rate-limited, the frontend keeps the last successful value and marks it as cached.

## Support Notes

When answering player questions, describe the widget as a transparency signal, not as evidence about a specific account. Detailed enforcement decisions remain private and should be handled through the appeal workflow.
