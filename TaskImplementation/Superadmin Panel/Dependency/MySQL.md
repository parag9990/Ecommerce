# MySQL Dependency - Superadmin Panel

## 1. What is this dependency?

MySQL ek relational database hai jisme data tables ke form me store hota hai. Superadmin backend ke liye admin users, permissions, settings, audit logs, and review tasks structured data hain, isliye MySQL suitable hai.

## 2. Why this service uses it

Docs ke hisaab se Superadmin Service MySQL use karta hai for:

- admin permissions
- platform settings
- immutable admin audit logs
- admin review tasks

Frontend direct MySQL ko access nahi karta. Browser hamesha API Gateway/backend ke through data leta hai.

## 3. Required or optional

Required for Superadmin backend. Current status partial hai because schema found hua but backend repository/migrations not found.

## 4. Where it is used in project

| Path | Use |
|------|-----|
| `database/draw.sql` | `superadmin_db` schema |
| `docs/05-database-design.md` | Superadmin DB tables/indexes |
| `docs/04-microservice-design.md` | Database choice and responsibilities |
| `backend/services/superadmin-service/.env` | `SUPERADMIN_DATABASE_DSN` and DB pool variables |

## 5. Tables detected

| Table | Purpose |
|-------|---------|
| `admin_users` | Admin identity, role, status |
| `admin_permissions` | Permission catalog |
| `admin_role_permissions` | Role to permission mapping |
| `platform_settings` | Platform settings JSON values |
| `admin_audit_logs` | Immutable admin action audit |
| `admin_review_tasks` | Maker-checker/review workflow tasks |

## 6. Installation steps

Install MySQL locally or use managed MySQL.

Ubuntu/Debian example:

```bash
sudo apt-get install mysql-server
```

This is system-level and may require admin permissions.

## 7. Docker setup, if possible

No docker-compose file found.

Suggested command based on common local MySQL setup:

```bash
docker run --name ecommerce-mysql -e MYSQL_ROOT_PASSWORD=root -p 3306:3306 -d mysql:8
```

This is suggested only. Change password for real usage.

## 8. Local setup without Docker

Suggested command based on project structure:

```bash
mysql -u root -p < database/draw.sql
```

This creates multiple project databases including `superadmin_db`.

## 9. Required environment variables

| Variable | Purpose |
|----------|---------|
| `SUPERADMIN_DATABASE_DSN` | MySQL DSN for Superadmin backend |
| `SUPERADMIN_DB_MAX_OPEN_CONNS` | DB pool max open connections |
| `SUPERADMIN_DB_MAX_IDLE_CONNS` | DB pool idle connections |
| `SUPERADMIN_DB_CONN_MAX_LIFETIME` | Connection lifetime |
| `SUPERADMIN_DB_PING_TIMEOUT` | Startup ping timeout |
| `SUPERADMIN_REQUIRE_DATABASE` | Whether service must fail if DB unavailable |

## 10. Start commands

Docker suggested:

```bash
docker start ecommerce-mysql
```

Local service suggested:

```bash
sudo service mysql start
```

## 11. Verify running commands

```bash
mysql -u root -p -e "SHOW DATABASES LIKE 'superadmin_db';"
mysql -u root -p -e "USE superadmin_db; SHOW TABLES;"
```

Expected tables include `admin_users`, `platform_settings`, and `admin_audit_logs`.

## 12. Common errors and fixes

| Error | Reason | Fix |
|-------|--------|-----|
| `Access denied` | Wrong user/password | Check DSN credentials |
| `Unknown database superadmin_db` | Schema not loaded | Run `database/draw.sql` |
| JSON column error | Old MySQL version | Use MySQL 8 compatible version |
| Backend starts but audit logs fail | Repository/migration missing | Implement backend repository and migrations |

## 13. Security notes

- Do not commit real DSN passwords.
- Admin audit logs should be append-only.
- Use least privilege DB user for Superadmin Service.
- Avoid storing raw IP/user PII if hashed values are enough.
- Backups and retention policy important hain because audit data compliance-sensitive hai.

## 14. Final checklist

- [x] Superadmin MySQL schema found.
- [x] Indexes found for audit/resource queries.
- [x] DB env variables found.
- [ ] Versioned migrations found.
- [ ] Rollback migrations found.
- [ ] Backend repository code found.

