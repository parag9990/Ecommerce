# User Service - Migrations Dependency

## 1. What is this dependency?

Migration SQL files database schema ko versioned tarike se create/update/rollback karne ke liye hote hain. User Service me current migration plain SQL files ke form me hai.

## 2. Why User Service uses it

Service startup se pehle `user_db` tables exist hone chahiye. Agar tables missing hain, repository queries fail hongi.

## 3. Required or Optional

| Item | Required? | Notes |
|---|---:|---|
| Up migration | Required | Tables create karta hai |
| Down migration | Required for rollback | Tables drop karta hai |
| Migration runner tool | Missing/Optional currently | No project runner found |
| MySQL CLI | Recommended | Manual migration apply/verify ke liye |

## 4. Where it is used in project

| Path | Purpose |
|---|---|
| `backend/services/user-service/migrations/001_create_user_tables.up.sql` | Creates `user_db` and User Service tables |
| `backend/services/user-service/migrations/001_create_user_tables.down.sql` | Drops User Service tables |
| `database/draw.sql` | Project-wide reference DDL |
| `internal/repository/*.go` | Queries depend on the tables and indexes |

## 5. Installation Steps

Install MySQL client:

```bash
mysql --version
```

No migration runner is configured in current project files.

Suggested future install if using `golang-migrate`:

```bash
go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

This is suggested only; current repo uses SQL files directly.

## 6. Docker Setup, If Possible

If MySQL is running in Docker:

```bash
docker exec -i ecommerce-user-mysql mysql -uroot -proot < backend/services/user-service/migrations/001_create_user_tables.up.sql
```

If the command cannot find the local SQL file from inside Docker, use shell redirection from host as shown above.

## 7. Local Setup Without Docker

Apply up migration:

```bash
mysql -h 127.0.0.1 -P 3306 -u root -p < backend/services/user-service/migrations/001_create_user_tables.up.sql
```

Rollback down migration in disposable/dev DB only:

```bash
mysql -h 127.0.0.1 -P 3306 -u root -p < backend/services/user-service/migrations/001_create_user_tables.down.sql
```

## 8. Required Environment Variables

Manual SQL migration does not read User Service env vars.

Runtime after migration needs:

| Variable | Purpose |
|---|---|
| `USER_SERVICE_DATABASE_DSN` | Must point to migrated `user_db` |
| `MYSQL_DSN` | Optional fallback |

## 9. Start Commands

Migration first, service second:

```bash
mysql -h 127.0.0.1 -P 3306 -u root -p < backend/services/user-service/migrations/001_create_user_tables.up.sql
cd backend/services/user-service
set -a
. ./.env
set +a
go run ./cmd/server
```

## 10. Verify Running Commands

Check table list:

```bash
mysql -h 127.0.0.1 -P 3306 -u ecommerce_user -p user_db -e "SHOW TABLES;"
```

Check important indexes:

```bash
mysql -h 127.0.0.1 -P 3306 -u ecommerce_user -p user_db -e "SHOW INDEX FROM users;"
mysql -h 127.0.0.1 -P 3306 -u ecommerce_user -p user_db -e "SHOW INDEX FROM user_addresses;"
mysql -h 127.0.0.1 -P 3306 -u ecommerce_user -p user_db -e "SHOW INDEX FROM seller_profiles;"
```

Check foreign keys:

```bash
mysql -h 127.0.0.1 -P 3306 -u ecommerce_user -p information_schema -e "SELECT table_name, constraint_name FROM referential_constraints WHERE constraint_schema='user_db';"
```

## 11. Common Errors and Fixes

| Error | Cause | Fix |
|---|---|---|
| `ERROR 1049 Unknown database 'user_db'` in down migration | DB was never created or already removed | Apply up migration first or recreate DB |
| `ERROR 1044 Access denied` | User lacks create/drop/alter permission | Use migration/admin user |
| `Table ... already exists` | Migration already applied | Current SQL uses `IF NOT EXISTS`, verify schema state |
| Foreign key creation fails | Referenced table/column missing or wrong collation | Run full up migration in order |
| Service starts but table missing | Migration not applied to same DB as DSN | Confirm DSN database and port |

## 12. Security Notes

- Down migration drops tables. Sirf dev/disposable DB me run karo unless rollback plan approved ho.
- Migration user and runtime app user alag rakho.
- Production migration se pehle backup mandatory hai.
- Plain SQL migration has no version table currently. Team ko manual state track karna padega.

## 13. Final Checklist

| Check | Done |
|---|---|
| MySQL running | [ ] |
| Up migration applied | [ ] |
| Down migration available | [ ] |
| Tables verified | [ ] |
| Indexes verified | [ ] |
| Foreign keys verified | [ ] |
| Runtime DSN points to same DB | [ ] |
| Migration state documented | [ ] |

