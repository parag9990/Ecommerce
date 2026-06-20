# Migrations Dependency - Superadmin Panel

## 1. What is this dependency?

Migrations database schema changes ko versioned tarike se apply/rollback karte hain. Production me plain one-shot SQL file enough nahi hota.

## 2. Why this service uses it

Superadmin backend ko `superadmin_db` tables chahiye. Admin audit/settings/review tables compliance-sensitive hain, so schema changes reviewed and rollbackable honi chahiye.

## 3. Required or optional

Required for production backend. Current status partial hai because schema exists in `database/draw.sql`, but migration folder/up-down files not found.

## 4. Where it is used in project

| Path | Use |
|------|-----|
| `database/draw.sql` | Full SQL schema including `superadmin_db` |
| `docs/03-folder-structure.md` | Expected service `migrations/` folder |
| `docs/13-developer-guide.md` | Every MySQL schema change should have up/down migration |

Not clearly found:

- `backend/services/superadmin-service/migrations/`
- migration runner for Superadmin
- rollback/down migration files

## 5. Installation steps

No migration tool found. Common options are `golang-migrate`, `goose`, or custom migration runner.

Suggested command based on common Go migration setup:

```bash
go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

This is suggested only.

## 6. Docker setup, if possible

No compose/migration runner found.

Suggested command with `migrate/migrate` image:

```bash
docker run --rm -v "$PWD/backend/services/superadmin-service/migrations:/migrations" migrate/migrate
```

This is suggested only and needs actual migration files.

## 7. Local setup without Docker

Current schema load:

```bash
mysql -u root -p < database/draw.sql
```

Suggested future migration layout:

```text
backend/services/superadmin-service/migrations/
  000001_create_superadmin_tables.up.sql
  000001_create_superadmin_tables.down.sql
```

## 8. Required environment variables

| Variable | Purpose |
|----------|---------|
| `SUPERADMIN_DATABASE_DSN` | Migration target DSN |

## 9. Start commands

Schema load command:

```bash
mysql -u root -p < database/draw.sql
```

Suggested future migration command:

```bash
migrate -path backend/services/superadmin-service/migrations -database "$SUPERADMIN_DATABASE_DSN" up
```

## 10. Verify running commands

```bash
mysql -u root -p -e "USE superadmin_db; SHOW TABLES;"
```

Expected:

```text
admin_users
admin_permissions
admin_role_permissions
platform_settings
admin_audit_logs
admin_review_tasks
```

## 11. Common errors and fixes

| Error | Reason | Fix |
|-------|--------|-----|
| Duplicate table errors | SQL already applied | Use idempotent SQL or versioned migrations |
| No rollback possible | Down migration missing | Add `.down.sql` files |
| Production drift | Manual DB changes | Use migration runner only |
| App starts before schema | Migration not run | Add startup/predeploy migration step |

## 12. Security notes

- Review migrations before production apply.
- Back up DB before destructive migrations.
- Avoid logging DSN credentials.
- Audit table migrations should not drop historical data casually.

## 13. Final checklist

- [x] Superadmin schema found in `database/draw.sql`.
- [ ] Service migration folder found.
- [ ] Up migrations found.
- [ ] Down migrations found.
- [ ] Migration runner command found.

