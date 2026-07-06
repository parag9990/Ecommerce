# Session Analytics Dashboard Troubleshooting

## Common Error

- Empty charts or failed API requests.
- `400 Bad Request` on analytics endpoints.
- Manual dev server port conflict.
- Admin endpoints return 401/403.

## Possible Cause

- Session service has no ingested data.
- Dashboard and gateway API contracts are out of sync.
- Funnel events use old names instead of `checkout_step` / `payment_result`.
- Heatmap requests use `mode` instead of `heatmap_type`.
- Seller dashboard is already using port `5174`.
- Missing admin/superadmin auth.

## Fix

- Generate traffic in the user app or ingest session events.
- Confirm `api/master-api.json` routes analytics live/sessions/journey/funnels/heatmaps to `session-service`.
- Confirm gateway allows dashboard query params such as `status`, `page_size`, `heatmap_type`, `source`, and `user_type`.
- Run manual dashboard with `-- --port 5175`.
- Verify gateway/session readiness and admin credentials.

## Verification Command

```powershell
curl http://localhost:8086/healthz
curl http://localhost:8080/health/ready
```
