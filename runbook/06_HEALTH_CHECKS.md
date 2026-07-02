# Health Checks

These checks are validated against the root [`docker-compose.yml`](../docker-compose.yml). Use [09 Ports And Endpoints](09_PORTS_AND_ENDPOINTS.md) for the full port map.

## Core Commands

```powershell
docker compose ps
curl.exe http://localhost:8080/health/live
curl.exe http://localhost:8080/health/ready
```

Windows PowerShell aliases `curl` to `Invoke-WebRequest`. Use `curl.exe` or add `-UseBasicParsing` when running these checks in Windows PowerShell.

## Compose-Validated Application Checks

| Service | Command |
| --- | --- |
| API Gateway liveness | `curl http://localhost:8080/health/live` |
| API Gateway readiness | `curl http://localhost:8080/health/ready` |
| Auth | `curl http://localhost:8081/readyz` |
| User | `docker compose exec user-service wget -qO- http://localhost:9091/health/ready` |
| Product | `curl http://localhost:8082/readyz` |
| Cart | `curl http://localhost:8083/readyz` |
| Wishlist | `curl http://localhost:8084/readyz` |
| Search | `curl http://localhost:8085/readyz` |
| Session | `curl http://localhost:8086/healthz` |
| CMS | `curl http://localhost:8087/healthz` |
| Recommendation | `curl http://localhost:8089/healthz` |
| Order | `curl http://localhost:8090/readyz` |
| Payment | `curl http://localhost:8091/healthz` |
| Notification | `curl http://localhost:8092/readyz` |
| Superadmin | `curl http://localhost:8093/readyz` |
| User app | `curl http://localhost:3000/healthz` |
| Seller dashboard | `curl http://localhost:3001/healthz` |
| Session analytics dashboard | `curl http://localhost:3002/healthz` |
| Superadmin panel | `curl http://localhost:3003/healthz` |

The user service publishes gRPC on host port `50052`, but its admin HTTP health port `9091` is internal only in compose. Use `docker compose exec` for that health check.

## Infrastructure Checks

| Component | Command or URL |
| --- | --- |
| MySQL | `docker compose exec mysql mysqladmin ping -h localhost -plocal_password` |
| MongoDB | `docker compose exec mongodb mongosh --quiet --username local_admin --password local_mongo_password --authenticationDatabase admin --eval "db.adminCommand('ping')"` |
| Redis | `docker compose exec redis redis-cli -a local_redis_password ping` |
| RabbitMQ management | `http://localhost:15672` |
| Kafka | `docker compose exec kafka /opt/kafka/bin/kafka-topics.sh --bootstrap-server localhost:29092 --list` |
| Typesense | `curl http://localhost:8108/health` |
| Mailpit | `http://localhost:8025` |
| Jaeger | `http://localhost:16686` |
| Prometheus | `http://localhost:9095` |

## Expected Result

Health endpoints should return HTTP 2xx. Readiness can fail while dependencies are still starting; wait a short time and inspect `docker compose logs <service>` if a service stays unhealthy.
