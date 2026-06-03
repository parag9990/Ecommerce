# User Service - Docker Dependency

## 1. What is this dependency?

Docker containers run services like MySQL in isolated environments. For this User Service, Docker is not a current service runtime dependency because no User Service Dockerfile or docker-compose service was found.

Docker is useful as an optional local setup path for MySQL.

## 2. Why User Service uses it

Current project files do not clearly show User Service running inside Docker.

Docker can still help in local development:

| Use | Status |
|---|---|
| Run MySQL locally | Optional and useful |
| Run User Service container | Missing Dockerfile |
| Run full stack compose | Missing docker-compose |
| Add MySQL healthcheck | Missing compose |

## 3. Required or Optional

| Item | Required? | Notes |
|---|---:|---|
| Docker for MySQL | Optional | Native MySQL also works |
| User Service Dockerfile | Missing | Not required for current direct Go run |
| docker-compose | Missing | No compose file found |

## 4. Where it is used in project

| Path | Finding |
|---|---|
| `backend/services/user-service/` | No Dockerfile found |
| repository root | No docker-compose file found |
| `TaskImplementation/User Service/*_Dependency.md` | Older docs mention Docker as optional local MySQL path |
| `docs/11-devops-external-services.md` | General future Docker guidance |

## 5. Installation Steps

Install Docker using OS-specific installer.

Verify:

```bash
docker --version
docker ps
```

## 6. Docker Setup, If Possible

Suggested command based on project structure for MySQL:

```bash
docker run --name ecommerce-user-mysql \
  -e MYSQL_ROOT_PASSWORD=root \
  -e MYSQL_DATABASE=user_db \
  -e MYSQL_USER=ecommerce_user \
  -e MYSQL_PASSWORD=Ecom8880User \
  -p 3306:3306 \
  -d mysql:8
```

Apply migration:

```bash
mysql -h 127.0.0.1 -P 3306 -u root -p < backend/services/user-service/migrations/001_create_user_tables.up.sql
```

If `3306` is busy:

```bash
docker run --name ecommerce-user-mysql \
  -e MYSQL_ROOT_PASSWORD=root \
  -e MYSQL_DATABASE=user_db \
  -e MYSQL_USER=ecommerce_user \
  -e MYSQL_PASSWORD=Ecom8880User \
  -p 3307:3306 \
  -d mysql:8
```

Then update DSN:

```bash
export USER_SERVICE_DATABASE_DSN='ecommerce_user:Ecom8880User@tcp(127.0.0.1:3307)/user_db?parseTime=true&charset=utf8mb4&loc=UTC'
```

## 7. Local Setup Without Docker

Install native MySQL and run:

```bash
mysql -h 127.0.0.1 -P 3306 -u root -p < backend/services/user-service/migrations/001_create_user_tables.up.sql
```

Start service directly:

```bash
cd backend/services/user-service
set -a
. ./.env
set +a
go run ./cmd/server
```

## 8. Required Environment Variables

Docker MySQL values must match the service DSN:

| Docker env | DSN equivalent |
|---|---|
| `MYSQL_DATABASE=user_db` | `/user_db` |
| `MYSQL_USER=ecommerce_user` | DSN username |
| `MYSQL_PASSWORD=Ecom8880User` | DSN password |
| `-p 3306:3306` | `tcp(127.0.0.1:3306)` |
| `-p 3307:3306` | `tcp(127.0.0.1:3307)` |

User Service env:

```text
USER_SERVICE_DATABASE_DSN
USER_SERVICE_GRPC_ADDRESS
USER_SERVICE_GRPC_REFLECTION
```

## 9. Start Commands

Start container:

```bash
docker start ecommerce-user-mysql
```

Stop container:

```bash
docker stop ecommerce-user-mysql
```

Start User Service:

```bash
cd backend/services/user-service
set -a
. ./.env
set +a
go run ./cmd/server
```

## 10. Verify Running Commands

Check container:

```bash
docker ps --filter name=ecommerce-user-mysql
```

Check MySQL:

```bash
mysqladmin ping -h 127.0.0.1 -P 3306 -u root -p
```

Check tables:

```bash
mysql -h 127.0.0.1 -P 3306 -u ecommerce_user -p user_db -e "SHOW TABLES;"
```

## 11. Common Errors and Fixes

| Error | Cause | Fix |
|---|---|---|
| `port is already allocated` | Host port `3306` busy | Use `3307:3306` and update DSN |
| Container exits immediately | Bad env or existing volume state | Check `docker logs ecommerce-user-mysql` |
| `Access denied` | DSN password differs from Docker password | Align `MYSQL_PASSWORD` and DSN |
| Service cannot connect to DB | MySQL still starting | Wait and run `mysqladmin ping` |
| Migration runs on host but not container | Wrong port/container | Confirm `docker ps` port mapping |

## 12. Security Notes

- Docker command examples use local dev passwords only.
- Do not reuse sample passwords in production.
- Persisted Docker volumes can keep old passwords/users even if env changes.
- Do not expose MySQL port publicly.
- Add healthchecks in future compose file.

## 13. Final Checklist

| Check | Done |
|---|---|
| Docker installed if using container MySQL | [ ] |
| MySQL container running | [ ] |
| Correct host port selected | [ ] |
| DSN port matches Docker mapping | [ ] |
| Migration applied | [ ] |
| User Service still run directly with Go | [ ] |
| Dockerfile missing noted for future work | [ ] |
| docker-compose missing noted for future work | [ ] |

