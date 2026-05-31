# Superadmin Service - Migrations Dependency

## 1. What Is This Dependency?

Migrations versioned SQL files hote hain jo database schema create/update/rollback karte hain. Is service me migrations plain `.sql` files ke form me present hain.

## 2. Why This Service Uses It

Superadmin Service ko MySQL tables chahiye before DB-backed routes work correctly. Migrations ensure karti hain ki same schema local/staging/prod me apply ho.

## 3. Required Or Optional

Required for DB-backed run. Without migrations, DB connection ho sakta hai but permission/settings/audit queries fail karengi because tables missing honge.

## 4. Where It Is Used In Project

| Path | Purpose |
|------|---------|
| `backend/services/superadmin-service/migrations/001_create_superadmin_rbac.up.sql` | RBAC tables |
| `001_create_superadmin_rbac.down.sql` | RBAC rollback |
| `002_seed_admin_rbac.up.sql` | Permission and role-permission seed |
| `002_seed_admin_rbac.down.sql` | Seed rollback |
| `003_create_admin_review_tasks.up.sql` | Review task table |
| `003_create_admin_review_tasks.down.sql` | Review task rollback |
| `004_add_open_review_task_uniqueness.up.sql` | Unique open task generated column |
| `004_add_open_review_task_uniqueness.down.sql` | Rollback uniqueness |
| `005_create_platform_settings.up.sql` | Platform settings table + defaults |
| `005_create_platform_settings.down.sql` | Settings rollback |
| `006_create_admin_audit_logs.up.sql` | Audit log table |
| `006_create_admin_audit_logs.down.sql` | Audit rollback |

## 5. Installation Steps

Install a migration CLI if you want automated migration execution.

Suggested command based on project structure:

```bash
go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

Migration CLI is not declared in this service `go.mod`; it is a local/CI tool.

## 6. Docker Setup, If Possible

No Docker Compose migration runner is clearly found.

If using MySQL Docker container, run migration CLI from host:

```bash
migrate \
  -path backend/services/superadmin-service/migrations \
  -database "mysql://superadmin:Ecom8880Admin@tcp(127.0.0.1:3306)/superadmin_db?parseTime=true&charset=utf8mb4&loc=UTC" \
  up
```

## 7. Local Setup Without Docker

1. Start local MySQL.
2. Create `superadmin_db`.
3. Apply migrations in order.

Suggested command based on project structure:

```bash
mysql -usuperadmin -pEcom8880Admin superadmin_db < backend/services/superadmin-service/migrations/001_create_superadmin_rbac.up.sql
mysql -usuperadmin -pEcom8880Admin superadmin_db < backend/services/superadmin-service/migrations/002_seed_admin_rbac.up.sql
mysql -usuperadmin -pEcom8880Admin superadmin_db < backend/services/superadmin-service/migrations/003_create_admin_review_tasks.up.sql
mysql -usuperadmin -pEcom8880Admin superadmin_db < backend/services/superadmin-service/migrations/004_add_open_review_task_uniqueness.up.sql
mysql -usuperadmin -pEcom8880Admin superadmin_db < backend/services/superadmin-service/migrations/005_create_platform_settings.up.sql
mysql -usuperadmin -pEcom8880Admin superadmin_db < backend/services/superadmin-service/migrations/006_create_admin_audit_logs.up.sql
```

## 8. Required Environment Variables

Migrations need a DB connection string.

| Variable | Example |
|----------|---------|
| `SUPERADMIN_DATABASE_DSN` | `superadmin:Ecom8880Admin@tcp(127.0.0.1:3306)/superadmin_db?parseTime=true&charset=utf8mb4&loc=UTC` |

For `golang-migrate`, DSN format uses `mysql://` prefix:

```text
mysql://superadmin:Ecom8880Admin@tcp(127.0.0.1:3306)/superadmin_db?parseTime=true&charset=utf8mb4&loc=UTC
```

## 9. Start Commands

Run all migrations:

```bash
migrate \
  -path backend/services/superadmin-service/migrations \
  -database "mysql://superadmin:Ecom8880Admin@tcp(127.0.0.1:3306)/superadmin_db?parseTime=true&charset=utf8mb4&loc=UTC" \
  up
```

Rollback one migration:

```bash
migrate \
  -path backend/services/superadmin-service/migrations \
  -database "mysql://superadmin:Ecom8880Admin@tcp(127.0.0.1:3306)/superadmin_db?parseTime=true&charset=utf8mb4&loc=UTC" \
  down 1
```

## 10. Verify Running Commands

```bash
mysql -usuperadmin -pEcom8880Admin superadmin_db -e "SHOW TABLES;"
mysql -usuperadmin -pEcom8880Admin superadmin_db -e "SELECT COUNT(*) FROM admin_permissions;"
mysql -usuperadmin -pEcom8880Admin superadmin_db -e "SELECT setting_key, version FROM platform_settings;"
```

Expected permission count after seed: 20 permissions.

## 11. Common Errors And Fixes

| Error | Reason | Fix |
|-------|--------|-----|
| `table does not exist` | Migrations not applied | Run migrations up |
| `duplicate key` during seed | Seed already applied | `ON DUPLICATE KEY` should handle most seed rows; verify schema |
| `DROP KEY` fails on rollback | Migration order mismatch | Roll back in reverse order only |
| Generated column unsupported | Old MySQL version | Use MySQL 8.x |
| FK failure on audit log insert | `actor_admin_id` not present in `admin_users` | Provision admin user before mutation audit |

## 12. Security Notes

- Migration user can have elevated schema permissions; app user should be more restricted.
- Review down migrations carefully before production rollback.
- Audit log rollback drops audit data; avoid destructive rollback in production without backup.
- Keep backups before applying schema changes.

## 13. Final Checklist

- [x] Migration folder exists.
- [x] Up migrations exist.
- [x] Down migrations exist.
- [x] RBAC seed exists.
- [x] Platform settings defaults exist.
- [x] Audit table exists.
- [ ] Migration CLI/tooling is not included in repo.
- [ ] Admin user seed/provisioning not clearly found.
