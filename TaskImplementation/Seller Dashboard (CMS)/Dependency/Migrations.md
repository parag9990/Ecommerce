# Migrations Dependency - Seller Dashboard (CMS)

## 1. What is this dependency?

Migrations database schema changes ko versioned and reversible banati hain. Example: `001_create_cms_tables.up.sql` create karega, and `001_create_cms_tables.down.sql` rollback karega.

## 2. Why this service uses it

CMS Service MySQL tables par depend karta hai. Team, coupons, campaigns, settings, and audit logs ke liye schema required hai. Production me raw `draw.sql` run karna enough nahi hota; controlled migrations chahiye.

## 3. Required or optional

Required for backend production readiness.

Current local schema source available hai, but proper migrations missing hain.

## 4. Where it is used in project

| Path | Use |
|---|---|
| `database/draw.sql` | Current full SQL DDL |
| `docs/13-developer-guide.md` | Migration guide says every MySQL schema change needs up/down migration |
| `docs/03-folder-structure.md` | Expected service `migrations/` folder |

Not clearly found:

- `database/migrations/`
- `backend/services/cms-service/migrations/`
- `.up.sql` files
- `.down.sql` files

## 5. Installation steps

Migration tool was not clearly found.

Suggested options:

- `golang-migrate/migrate`
- `goose`
- custom Go migration runner

Suggested command based on project structure:

```bash
# Suggested command based on project structure
migrate -path backend/services/cms-service/migrations -database "$CMS_MYSQL_DSN" up
```

## 6. Docker setup, if possible

Not clearly found in project files.

When compose is added, migration can run as:

- one-shot migration container
- app startup migration step for local only
- CI/CD migration job

Production me app container ke startup par destructive migrations avoid karo.

## 7. Local setup without Docker

Current fallback:

```bash
# Suggested command based on project structure
mysql -u root -p < database/draw.sql
```

This is not a rollback-safe migration workflow. Ye local bootstrap ke liye acceptable ho sakta hai, but production ke liye migration files chahiye.

## 8. Required environment variables

| Variable | Required | Purpose |
|---|---|---|
| `CMS_MYSQL_DSN` | Optional but recommended | Migration DB connection |
| `CMS_DB_HOST` | If DSN not used | MySQL host |
| `CMS_DB_PORT` | If DSN not used | MySQL port |
| `CMS_DB_NAME` | If DSN not used | Database name |
| `CMS_DB_USER` | If DSN not used | Migration DB user |
| `CMS_DB_PASSWORD` | If DSN not used | DB password |

## 9. Start commands

Migrations do not stay running as a service.

Suggested command based on project structure:

```bash
migrate -path backend/services/cms-service/migrations -database "$CMS_MYSQL_DSN" up
```

Rollback:

```bash
migrate -path backend/services/cms-service/migrations -database "$CMS_MYSQL_DSN" down 1
```

These commands require migration files and tool installation.

## 10. Verify running commands

Check tables:

```bash
mysql -u cms_user -p -D cms_db -e "SHOW TABLES;"
```

Check important indexes:

```bash
mysql -u cms_user -p -D cms_db -e "SHOW INDEX FROM coupons;"
mysql -u cms_user -p -D cms_db -e "SHOW INDEX FROM seller_staff;"
```

## 11. Common errors and fixes

| Error | Reason | Fix |
|---|---|---|
| Migration command missing | Tool not installed | Install selected migration tool |
| Dirty migration state | Migration failed halfway | Fix DB state carefully, then force version only after review |
| Rollback fails | `.down.sql` missing | Add rollback migration |
| Table already exists | Raw `draw.sql` already applied | Baseline migration version or recreate local DB |
| Foreign key failure | Migration order wrong | Create parent tables before child tables |

## 12. Security notes

- Migration DB user should have only required schema permissions.
- Review destructive migrations manually.
- Backup before production migration.
- Store migration history in DB.
- Never include sample real user/seller data in migrations.

## 13. Final checklist

- [x] CMS DDL found in `database/draw.sql`.
- [x] CMS tables identified.
- [x] Migration requirement found in docs.
- [ ] Dedicated CMS migration folder not found.
- [ ] Up migrations not found.
- [ ] Down migrations not found.
- [ ] Migration tool/config not found.
