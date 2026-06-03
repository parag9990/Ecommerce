# User Service - MySQL Dependency

## 1. What is this dependency?

MySQL ek relational database hai jisme data tables ke form me store hota hai. User Service ke structured profile data ke liye MySQL suitable hai because unique keys, foreign keys, indexes, and transactions strong support dete hain.

## 2. Why User Service uses it

User Service ko profile, address, seller profile, and KYC metadata durable store karna hota hai.

Current database:

```text
user_db
```

Current tables:

| Table | Purpose |
|---|---|
| `users` | User profile and status |
| `user_addresses` | Address book |
| `seller_profiles` | Seller profile and approval status |
| `seller_kyc_documents` | KYC metadata |

## 3. Required or Optional

| Item | Required? | Notes |
|---|---:|---|
| MySQL server | Yes at runtime | Service startup pings DB |
| MySQL client CLI | Recommended | Migration and verification ke liye |
| Docker MySQL | Optional | Native MySQL ke alternative ke roop me |
| `github.com/go-sql-driver/mysql` | Yes | Go app DB connect karne ke liye |

## 4. Where it is used in project

| Path | Purpose |
|---|---|
| `backend/services/user-service/internal/config/config.go` | DSN and DB pool config load karta hai |
| `backend/services/user-service/cmd/server/main.go` | `sql.Open`, DB ping, repositories wire karta hai |
| `backend/services/user-service/internal/repository/mysql_user_repository.go` | `users` queries |
| `backend/services/user-service/internal/repository/mysql_address_repository.go` | `user_addresses` queries |
| `backend/services/user-service/internal/repository/mysql_seller_repository.go` | `seller_profiles` and `seller_kyc_documents` queries |
| `backend/services/user-service/migrations/001_create_user_tables.up.sql` | Schema create |
| `backend/services/user-service/migrations/001_create_user_tables.down.sql` | Schema rollback |

## 5. Installation Steps

Native install:

```bash
mysql --version
```

If command missing, install MySQL 8 compatible server/client using your OS package manager.

Create runtime user manually if Docker is not creating it:

```sql
CREATE DATABASE IF NOT EXISTS user_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER IF NOT EXISTS 'ecommerce_user'@'%' IDENTIFIED BY 'Ecom8880User';
GRANT SELECT, INSERT, UPDATE, DELETE ON user_db.* TO 'ecommerce_user'@'%';
FLUSH PRIVILEGES;
```

Migration/admin user also needs create/alter/index/foreign-key permissions.

## 6. Docker Setup, If Possible

No checked-in docker-compose file was found. Suggested command based on project structure:

```bash
docker run --name ecommerce-user-mysql \
  -e MYSQL_ROOT_PASSWORD=root \
  -e MYSQL_DATABASE=user_db \
  -e MYSQL_USER=ecommerce_user \
  -e MYSQL_PASSWORD=Ecom8880User \
  -p 3306:3306 \
  -d mysql:8
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

Then use DSN port `3307`.

## 7. Local Setup Without Docker

1. Start local MySQL service.
2. Create DB/runtime user.
3. Apply migration.
4. Export DSN.
5. Start User Service.

Suggested command based on project structure:

```bash
mysql -h 127.0.0.1 -P 3306 -u root -p < backend/services/user-service/migrations/001_create_user_tables.up.sql
```

## 8. Required Environment Variables

| Variable | Required | Example |
|---|---:|---|
| `USER_SERVICE_DATABASE_DSN` | Yes | `ecommerce_user:Ecom8880User@tcp(127.0.0.1:3306)/user_db?parseTime=true&charset=utf8mb4&loc=UTC` |
| `MYSQL_DSN` | Optional fallback | Same format as above |
| `USER_SERVICE_DB_MAX_OPEN_CONNS` | Optional | `25` |
| `USER_SERVICE_DB_MAX_IDLE_CONNS` | Optional | `25` |
| `USER_SERVICE_DB_CONN_MAX_LIFETIME` | Optional | `5m` |
| `USER_SERVICE_DB_PING_TIMEOUT` | Optional | `5s` |

Important: `parseTime=true` required hai, warna timestamp scan errors aa sakte hain.

## 9. Start Commands

Start MySQL:

```bash
docker start ecommerce-user-mysql
```

Apply migration:

```bash
mysql -h 127.0.0.1 -P 3306 -u root -p < backend/services/user-service/migrations/001_create_user_tables.up.sql
```

Start backend:

```bash
cd backend/services/user-service
set -a
. ./.env
set +a
go run ./cmd/server
```

## 10. Verify Running Commands

Check DB is alive:

```bash
mysqladmin ping -h 127.0.0.1 -P 3306 -u root -p
```

Check tables:

```bash
mysql -h 127.0.0.1 -P 3306 -u ecommerce_user -p user_db -e "SHOW TABLES;"
```

Expected tables:

```text
users
user_addresses
seller_profiles
seller_kyc_documents
```

## 11. Common Errors and Fixes

| Error | Cause | Fix |
|---|---|---|
| `USER_SERVICE_DATABASE_DSN is required` | Env var not exported | Use `set -a; . ./.env; set +a` before `go run` |
| `ping mysql: connect: connection refused` | MySQL not running or wrong port | Start MySQL and verify port `3306`/`3307` |
| `Unknown database 'user_db'` | Migration/DB creation not applied | Run up migration or create DB |
| `Access denied for user` | Wrong password or grants | Recreate grants and confirm DSN |
| Timestamp scan error | DSN missing `parseTime=true` | Add `parseTime=true` |
| Duplicate user error | Unique key conflict on user/auth/email | Use new IDs/email or handle `AlreadyExists` |

## 12. Security Notes

- `.env` me DB password hota hai. Isko commit mat karo.
- Runtime DB user ko minimal permission do: usually SELECT, INSERT, UPDATE, DELETE.
- Migration/admin user alag rakho jisko CREATE/ALTER permissions mil sakte hain.
- Production me root MySQL user app DSN me use mat karo.
- `USER_SERVICE_GRPC_REFLECTION=true` local me ok hai, production exposure carefully control karo.

## 13. Final Checklist

| Check | Done |
|---|---|
| MySQL server running | [ ] |
| `user_db` exists | [ ] |
| Runtime user has grants | [ ] |
| Migration applied | [ ] |
| DSN has `parseTime=true` | [ ] |
| User Service can ping DB at startup | [ ] |
| Tables verified with `SHOW TABLES` | [ ] |

