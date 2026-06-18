# MySQL Dependency - Seller Dashboard (CMS)

## 1. What is this dependency?

MySQL ek relational database hai jisme data tables ke form me store hota hai. CMS Service MySQL use karta hai because coupons, campaigns, seller staff, settings, and audit logs structured data hain.

## 2. Why this service uses it

Seller Dashboard (CMS) ko CMS-owned data persist karna hota hai:

- seller settings
- seller staff roles
- coupons
- coupon rules
- coupon redemptions
- campaigns
- CMS audit logs

## 3. Required or optional

Required for CMS Backend.

Frontend directly MySQL ko access nahi karta. Flow ye hai:

```text
Seller Dashboard -> API Gateway -> CMS Service -> MySQL
```

## 4. Where it is used in project

| Path | Use |
|---|---|
| `backend/services/cms-service/.env` | `CMS_MYSQL_DSN` and `CMS_DB_*` config |
| `database/draw.sql` | CMS DB/tables DDL |
| `docs/05-database-design.md` | CMS database strategy |
| `docs/04-microservice-design.md` | CMS Service database choice |

## 5. Installation steps

Install MySQL 8+ locally.

Ubuntu/Debian example:

```bash
sudo apt update
sudo apt install mysql-server
```

If MySQL is already installed, only schema setup is needed.

## 6. Docker setup, if possible

No docker-compose file was clearly found.

Suggested compose service based on project structure:

```yaml
mysql:
  image: mysql:8
  environment:
    MYSQL_DATABASE: cms_db
    MYSQL_USER: cms_user
    MYSQL_PASSWORD: change-this-locally
    MYSQL_ROOT_PASSWORD: change-this-locally
  ports:
    - "3306:3306"
```

This is a suggested example, not an existing project file.

## 7. Local setup without Docker

Create/apply schema:

```bash
# Suggested command based on project structure
mysql -u root -p < database/draw.sql
```

Create a CMS user if needed:

```sql
CREATE USER 'cms_user'@'localhost' IDENTIFIED BY 'change-this-locally';
GRANT ALL PRIVILEGES ON cms_db.* TO 'cms_user'@'localhost';
FLUSH PRIVILEGES;
```

## 8. Required environment variables

| Variable | Required | Purpose |
|---|---|---|
| `CMS_MYSQL_DSN` | Optional | Full DSN override |
| `CMS_DB_HOST` | Yes | MySQL host |
| `CMS_DB_PORT` | Yes | MySQL port |
| `CMS_DB_NAME` | Yes | DB name, expected `cms_db` |
| `CMS_DB_USER` | Yes | DB username |
| `CMS_DB_PASSWORD` | Yes | DB password secret |
| `CMS_DB_MAX_OPEN_CONNS` | Yes | Pool max open connections |
| `CMS_DB_MAX_IDLE_CONNS` | Yes | Pool max idle connections |
| `CMS_DB_CONN_MAX_LIFETIME_SECONDS` | Yes | Pool lifetime |
| `CMS_DB_TIMEZONE` | Yes | DB timezone |

## 9. Start commands

System service:

```bash
sudo service mysql start
```

or:

```bash
sudo systemctl start mysql
```

## 10. Verify running commands

```bash
mysqladmin ping -h 127.0.0.1 -P 3306
```

Check CMS tables:

```bash
mysql -u cms_user -p -D cms_db -e "SHOW TABLES;"
```

## 11. Common errors and fixes

| Error | Reason | Fix |
|---|---|---|
| `Access denied` | Wrong user/password | Verify `CMS_DB_USER` and `CMS_DB_PASSWORD` |
| `Unknown database cms_db` | Schema not applied | Run `database/draw.sql` |
| Time parsing issue | DSN missing parse/time location settings | Add proper DSN options when using `CMS_MYSQL_DSN` |
| Duplicate coupon code | `coupons.code` unique index | Return clean validation error to frontend |
| FK error on coupon rules/redemptions | Parent coupon missing | Create coupon before rule/redemption |

## 12. Security notes

- DB password ko git me commit mat karo.
- Use least privilege: CMS user ko sirf `cms_db` access do.
- Production me TLS-enabled MySQL connection use karo.
- Audit tables me sensitive secrets or raw tokens store mat karo.
- Backups and migration rollback plan maintain karo.

## 13. Final checklist

- [x] CMS DB schema found in `database/draw.sql`.
- [x] CMS DB env vars found.
- [x] Tables and indexes found for CMS.
- [ ] Dedicated migration files not found.
- [ ] Rollback migrations not found.
- [ ] Docker MySQL service not found.

