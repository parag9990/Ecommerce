# Database Local Runbook

## Ownership

| Service | Store | Database | Migrations |
|---|---|---|---|
| Auth, User, Order, Payment, CMS, Superadmin | MySQL 8.4 | `<service>_db` | service `migrations/*.up.sql` |
| Product, Cart, Wishlist, Recommendation, Session, Notification | MongoDB 8 | `<service>_db` | service JS migrations |
| Auth, Cart, Recommendation, Search, Session, Gateway | Redis | numbered logical DBs | no schema migration |

## Apply migrations

```bash
docker compose run --rm migrate-auth
docker compose run --rm migrate-user
docker compose run --rm migrate-order
docker compose run --rm migrate-payment
docker compose run --rm migrate-cms
docker compose run --rm migrate-superadmin
docker compose run --rm migrate-mongodb
```

Compose runs these automatically before dependent applications.

## Verify

```bash
docker compose exec mysql mysql -uroot -plocal_password -e "SHOW DATABASES;"
docker compose exec mongodb mongosh --username local_admin --password local_mongo_password --authenticationDatabase admin --eval "db.adminCommand('listDatabases')"
docker compose exec redis redis-cli -a local_redis_password ping
```

## Seed data

Notification migration `002_seed_renderable_templates.up.js` seeds templates. No common product/user demo seed command exists; this is Missing / Required Work.

## Reset

```bash
docker compose down -v
docker compose up -d mysql mongodb redis
```

This deletes all local data. Connection refused normally means the DB is not healthy yet; authentication errors mean the DSN/URI does not match the local dummy credentials.
