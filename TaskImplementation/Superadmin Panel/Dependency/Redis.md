# Redis Dependency - Superadmin Panel

## 1. What is this dependency?

Redis ek in-memory data store hai. Ye fast counters, cache, rate limits, and short-lived data ke liye use hota hai.

## 2. Why this service uses it

Superadmin Panel direct Redis use nahi karta. Indirectly:

- API Gateway Redis use karta hai for rate limiting.
- Session Service docs Redis use karte hain for active sessions/live metrics.
- Superadmin session oversight screen `/api/v1/analytics/live` consume karta hai.

## 3. Required or optional

Required if API Gateway rate limiting or session live metrics enabled hain. Current status partial hai because env/docs found hue, but runtime code not found.

## 4. Where it is used in project

| Path | Use |
|------|-----|
| `backend/services/api-gateway/.env` | `REDIS_ADDR`, rate limit env |
| `docs/02-system-architecture.md` | Redis for caching/rate limiting/active sessions |
| `docs/04-microservice-design.md` | Gateway Redis and Session Redis responsibility |
| `TaskImplementation/Superadmin Panel/task6.md` | Session live metrics flow uses Redis |

## 5. Installation steps

Local Redis install:

```bash
sudo apt-get install redis-server
```

Or use Docker command below.

## 6. Docker setup, if possible

No compose file found.

Suggested command based on common Redis setup:

```bash
docker run --name ecommerce-redis -p 6379:6379 -d redis:7
```

## 7. Local setup without Docker

```bash
redis-server
```

Or service mode:

```bash
sudo service redis-server start
```

## 8. Required environment variables

| Variable | Purpose |
|----------|---------|
| `REDIS_ADDR` | Redis host/port, found `localhost:6379` |
| `REDIS_PASSWORD` | Redis password if enabled |
| `REDIS_DB` | Redis database number |
| `REDIS_TLS_ENABLED` | TLS toggle |
| `REDIS_DIAL_TIMEOUT` | Dial timeout |
| `RATE_LIMIT_ENABLED` | Rate limiter toggle |
| `RATE_LIMIT_KEY_PREFIX` | Prefix for rate limit keys |
| `RATE_LIMIT_FAIL_OPEN` | Behavior if Redis fails |

## 9. Start commands

Docker:

```bash
docker start ecommerce-redis
```

Local:

```bash
redis-server
```

## 10. Verify running commands

```bash
redis-cli -h localhost -p 6379 ping
```

Expected:

```text
PONG
```

## 11. Common errors and fixes

| Error | Reason | Fix |
|-------|--------|-----|
| `connection refused` | Redis not running | Start Redis |
| Rate limiter blocks too aggressively | Bad limits/window | Tune `RATE_LIMIT_*` env |
| Gateway startup fails | Redis required but unavailable | Start Redis or configure fail-open for local only |
| Live metrics empty | Session Service counters not writing | Check Session Service integration |

## 12. Security notes

- Do not expose Redis publicly.
- Use password/TLS in production.
- Keep key prefixes service-specific.
- Admin mutation rate limits should be stricter than read routes.

## 13. Final checklist

- [x] Redis env names found in API Gateway.
- [x] Session live metrics docs mention Redis.
- [ ] Redis code integration found.
- [ ] Docker Compose Redis service found.

