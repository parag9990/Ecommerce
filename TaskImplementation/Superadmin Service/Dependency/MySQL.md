# Superadmin Service - MySQL Dependency

## 1. What Is This Dependency?

MySQL ek relational database hai jisme data tables ke form me store hota hai. Superadmin Service ke liye MySQL source of truth hai for admin users, permissions, role mappings, review tasks, platform settings, and audit logs.

## 2. Why This Service Uses It

Superadmin data structured and audit-sensitive hai. Isliye MySQL useful hai because:

| Need | MySQL Benefit |
|------|---------------|
| RBAC permission lookup | Indexed relational joins |
| Audit logs | Durable append-style storage |
| Platform settings | Versioned rows with JSON values |
| Review tasks | Unique/indexed workflow rows |
| Consistency | InnoDB transactions and constraints |

## 3. Required Or Optional

| Environment | Required? | Detail |
|-------------|-----------|--------|
| Local quick boot | Optional | If DSN empty, service starts but RBAC/admin features deny or become unavailable |
| Local full testing | Required | Admin routes need permission tables |
| Staging/Production | Required | `SUPERADMIN_REQUIRE_DATABASE=true` recommended |

## 4. Where It Is Used In Project

| Path | Usage |
|------|-------|
| `backend/services/superadmin-service/cmd/server/main.go` | Opens DB with `sql.Open("mysql", dsn)` |
| `backend/services/superadmin-service/internal/repository/mysql_permission_repository.go` | RBAC permission lookup |
| `backend/services/superadmin-service/internal/repository/mysql_review_task_repository.go` | Admin review task storage |
| `backend/services/superadmin-service/internal/repository/mysql_platform_settings_repository.go` | Platform settings storage |
| `backend/services/superadmin-service/internal/repository/mysql_audit_log_repository.go` | Audit log insert/list |
| `backend/services/superadmin-service/migrations/*.sql` | Schema and seed migrations |
| `backend/services/superadmin-service/.env` | `SUPERADMIN_DATABASE_DSN` |

## 5. Installation Steps

Install MySQL 8.x locally or run it using Docker.

Suggested command based on project structure:

```bash
mysql --version
mysql -uroot -p -e "CREATE DATABASE IF NOT EXISTS superadmin_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
```

Create application user:

```sql
CREATE USER 'superadmin'@'%' IDENTIFIED BY 'change_me';
GRANT SELECT, INSERT, UPDATE, DELETE, CREATE, ALTER, INDEX, REFERENCES
ON superadmin_db.*
TO 'superadmin'@'%';
FLUSH PRIVILEGES;
```

Use a stronger password in real environments.

## 6. Docker Setup, If Possible

No project Dockerfile or docker-compose file was clearly found. Manual MySQL Docker run is possible.

Suggested command based on project structure:

```bash
docker run --name ecommerce-superadmin-mysql \
  -e MYSQL_ROOT_PASSWORD=localroot \
  -e MYSQL_DATABASE=superadmin_db \
  -e MYSQL_USER=superadmin \
  -e MYSQL_PASSWORD=Ecom8880Admin \
  -p 3306:3306 \
  -d mysql:8.0
```

Verify container:

```bash
docker exec ecommerce-superadmin-mysql mysql -usuperadmin -pEcom8880Admin -e "SHOW DATABASES;"
```

## 7. Local Setup Without Docker

1. Install MySQL from package manager.
2. Start MySQL service.
3. Create `superadmin_db`.
4. Create app user.
5. Apply migrations.
6. Export `SUPERADMIN_DATABASE_DSN`.

Suggested command based on project structure:

```bash
mysql -usuperadmin -pEcom8880Admin superadmin_db -e "SELECT 1;"
```

## 8. Required Environment Variables

| Variable | Example | Required? |
|----------|---------|-----------|
| `SUPERADMIN_DATABASE_DSN` | `superadmin:Ecom8880Admin@tcp(127.0.0.1:3306)/superadmin_db?parseTime=true&charset=utf8mb4&loc=UTC` | Yes for DB run |
| `MYSQL_DSN` | Same DSN | Optional fallback |
| `SUPERADMIN_DB_MAX_OPEN_CONNS` | `25` | Optional |
| `SUPERADMIN_DB_MAX_IDLE_CONNS` | `25` | Optional |
| `SUPERADMIN_DB_CONN_MAX_LIFETIME` | `5m` | Optional |
| `SUPERADMIN_DB_PING_TIMEOUT` | `5s` | Optional |
| `SUPERADMIN_REQUIRE_DATABASE` | `true` in staging/prod | Recommended |

## 9. Start Commands

Start MySQL with your local service manager or Docker.

Start Superadmin Service:

```bash
cd backend/services/superadmin-service
set -a
source .env
set +a
go run ./cmd/server
```

## 10. Verify Running Commands

```bash
mysql -usuperadmin -pEcom8880Admin -h127.0.0.1 -P3306 superadmin_db -e "SHOW TABLES;"
curl http://127.0.0.1:8088/readyz
```

Expected tables:

```text
admin_audit_logs
admin_permissions
admin_review_tasks
admin_role_permissions
admin_users
platform_settings
```

## 11. Common Errors And Fixes

| Error | Reason | Fix |
|-------|--------|-----|
| `SUPERADMIN_DATABASE_DSN is required` | Strict env requires DSN | Set DSN or set local `SUPERADMIN_REQUIRE_DATABASE=false` |
| `access denied for user` | Wrong user/password/grants | Recreate user or grant permissions |
| `/readyz` returns `not_ready` | DB ping failed | Check MySQL status and DSN |
| Admin API returns forbidden | No active `admin_users` row or missing permissions | Insert/provision active admin user with mapped role |
| Migration fails on FK | Tables applied in wrong order | Run migrations from `001` upward |
| Duplicate key on review task | Open task uniqueness is enforced | Close existing task before creating another open task |

## 12. Security Notes

- Real DB password should not be committed in `.env`.
- Use least-privilege DB user; avoid root in app DSN.
- Keep audit tables protected from manual update/delete.
- Use TLS for remote MySQL.
- Back up `admin_audit_logs` and `platform_settings`.
- Rotate DB passwords regularly.

## 13. Final Checklist

- [x] MySQL is an actual dependency.
- [x] DB schema migrations exist.
- [x] Go MySQL driver exists in `go.mod`.
- [x] DSN config exists.
- [ ] Safe `.env.example` is missing.
- [ ] Project Docker/compose setup is missing.
- [ ] Admin user provisioning is not clearly found in project files.
