# Redis Dependency - Seller Dashboard (CMS)

## 1. What is this dependency?

Redis ek in-memory data store hai. Ye fast counters, cache, rate limiting, sessions, and temporary data ke liye use hota hai.

## 2. Why this service uses it

Seller Dashboard frontend directly Redis use nahi karta. API Gateway Redis use karta hai rate limiting ke liye, and Seller Dashboard ki API calls Gateway se pass hoti hain.

Gateway env me:

- `RATE_LIMIT_ENABLED=true`
- `REDIS_ADDR=localhost:6379`

## 3. Required or optional

Required if API Gateway rate limiting enabled hai.

Optional only if `RATE_LIMIT_ENABLED=false` set kiya jaye for local development.

## 4. Where it is used in project

| Path | Use |
|---|---|
| `backend/services/api-gateway/.env` | Redis rate limit config |
| `docs/02-system-architecture.md` | Redis platform dependency |
| `docs/11-devops-external-services.md` | Redis setup guidance |
| `docs/12-logging-monitoring-scalability.md` | Seller settings cache mentioned in docs |

CMS Service `.env` me direct Redis config clearly found nahi hua.

## 5. Installation steps

Ubuntu/Debian example:

```bash
sudo apt update
sudo apt install redis-server
```

## 6. Docker setup, if possible

No docker-compose file was clearly found.

Suggested compose service based on project structure:

```yaml
redis:
  image: redis:7-alpine
  ports:
    - "6379:6379"
```

Production me password/TLS configure karo.

## 7. Local setup without Docker

```bash
redis-server
```

or as service:

```bash
sudo service redis-server start
```

## 8. Required environment variables

| Variable | Required | Purpose |
|---|---|---|
| `RATE_LIMIT_ENABLED` | Yes | Enables/disables gateway rate limiting |
| `REDIS_ADDR` | If rate limit enabled | Redis address |
| `REDIS_PASSWORD` | Optional local, required production | Redis password |
| `REDIS_DB` | Yes | Redis logical DB |
| `REDIS_TLS_ENABLED` | Production recommended | TLS toggle |
| `REDIS_DIAL_TIMEOUT` | Yes | Connection timeout |
| `RATE_LIMIT_KEY_PREFIX` | Yes | Key namespace |
| `RATE_LIMIT_FAIL_OPEN` | Yes | Fail behavior if Redis down |

## 9. Start commands

```bash
redis-server
```

## 10. Verify running commands

```bash
redis-cli ping
```

Expected:

```text
PONG
```

## 11. Common errors and fixes

| Error | Reason | Fix |
|---|---|---|
| Gateway startup fails | Redis not running and fail-open false | Start Redis or disable rate limit locally |
| `NOAUTH Authentication required` | Redis password required | Set `REDIS_PASSWORD` |
| Too many 429 responses | Rate limit too strict | Tune gateway route/user/IP limits |
| Redis connection timeout | Wrong host/port | Verify `REDIS_ADDR` |

## 12. Security notes

- Production Redis public internet par expose mat karo.
- Password/TLS enable karo.
- Rate limit keys me raw PII store mat karo; hash IDs if needed.
- Memory policy and maxmemory configure karo.

## 13. Final checklist

- [x] Redis env found in API Gateway config.
- [x] Rate limiting dependency identified.
- [ ] Redis Docker Compose service not found.
- [ ] CMS direct Redis config not found.
- [ ] Gateway rate limiter code not clearly found.

