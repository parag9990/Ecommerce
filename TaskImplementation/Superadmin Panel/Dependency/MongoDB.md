# MongoDB Dependency - Superadmin Panel

## 1. What is this dependency?

MongoDB document database hai. Flexible/high-volume session events ke liye use hota hai.

## 2. Why this service uses it

Superadmin Panel ka Session Oversight module `/api/v1/analytics/live`, `/api/v1/analytics/sessions`, and `/api/v1/analytics/sessions/{id}/journey` endpoints call karta hai. Docs ke hisaab se Session Service MongoDB me sessions and session events store karta hai.

Frontend direct MongoDB access nahi karta.

## 3. Required or optional

Indirect required for Session Oversight real data. If sessions feature disabled ho, Superadmin core shell still run kar sakta hai.

## 4. Where it is used in project

| Path | Use |
|------|-----|
| `database/mongodb-schema-design.md` | Session DB schema |
| `docs/05-database-design.md` | Session uses MongoDB + Redis |
| `frontend/superadmin-panel/src/features/sessions/api/sessions-api.ts` | Consumes session analytics REST endpoints |
| `TaskImplementation/Superadmin Panel/task6.md` | Session oversight dependency |

## 5. Installation steps

Install MongoDB locally or use managed MongoDB.

Suggested Docker command is easiest for local development.

## 6. Docker setup, if possible

No compose file found.

Suggested command based on common MongoDB setup:

```bash
docker run --name ecommerce-mongodb -p 27017:27017 -d mongo:7
```

## 7. Local setup without Docker

Start local MongoDB service:

```bash
mongod --dbpath /tmp/ecommerce-mongodb
```

This is suggested only. Create the dbpath first if needed.

## 8. Required environment variables

Superadmin Panel does not read Mongo env directly. Session Service should have Mongo env, but Session Service source/env details for Mongo were not clearly found in this scan.

Expected variables if backend exists:

| Variable | Purpose |
|----------|---------|
| `MONGODB_URI` | Mongo connection string |
| `SESSION_DATABASE_NAME` | `session_db` or equivalent |

These are suggested names, not clearly found in project files.

## 9. Start commands

Docker:

```bash
docker start ecommerce-mongodb
```

Local:

```bash
mongod --dbpath /tmp/ecommerce-mongodb
```

## 10. Verify running commands

```bash
mongosh --eval "db.adminCommand('ping')"
```

Schema/index verification based on docs:

```bash
mongosh session_db --eval "db.sessions.getIndexes(); db.session_events.getIndexes();"
```

## 11. Common errors and fixes

| Error | Reason | Fix |
|-------|--------|-----|
| Session page empty | Session Service not writing events | Start/fix Session Service ingestion |
| Journey API slow | Missing indexes | Add indexes from `database/mongodb-schema-design.md` |
| Connection refused | MongoDB not running | Start MongoDB |
| High storage growth | Session event volume high | Use TTL index and retention policy |

## 12. Security notes

- Do not expose MongoDB publicly.
- Use auth/TLS in production.
- Session data can contain device/IP/user traces. Mask PII before returning admin UI responses.
- TTL retention should match privacy policy.

## 13. Final checklist

- [x] Session Mongo schema docs found.
- [x] Frontend session API module found.
- [ ] Session Service Mongo env clearly found.
- [ ] Session Service source code found.
- [ ] Docker Compose Mongo service found.

