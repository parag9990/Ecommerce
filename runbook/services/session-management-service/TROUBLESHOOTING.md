# Session Management Service Troubleshooting

## Common Error

- `/healthz` fails.
- Analytics dashboard is empty.
- Retention worker fails.

## Possible Cause

- MongoDB or Redis is unavailable.
- No session events have been ingested.
- Worker env still points at Docker DNS during host run.

## Fix

- Run `docker compose up migrate-session-mongodb`.
- Start Redis and MongoDB.
- Trigger user-app traffic or send session events through the gateway.

## Verification Command

```powershell
curl http://localhost:8086/healthz
docker compose logs session-service
```
