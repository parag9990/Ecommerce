# Ports And Endpoints

This page is validated against the root [`docker-compose.yml`](../docker-compose.yml). Host ports and health commands below are for the Docker Compose local run.

Windows PowerShell aliases `curl` to `Invoke-WebRequest`; use `curl.exe` or `Invoke-WebRequest -UseBasicParsing` if a health command is copied into Windows PowerShell.

## Application Services

| Service | Host port(s) | Base URL | Compose-validated health check | Start command |
| --- | ---: | --- | --- | --- |
| API Gateway | 8080, 8099, 19090 | `http://localhost:8080` | `curl http://localhost:8080/health/live`; `curl http://localhost:8080/health/ready` | `docker compose up -d --build api-gateway` |
| Auth | 8081 | `http://localhost:8081` | `curl http://localhost:8081/readyz` | `docker compose up -d --build auth-service` |
| User | 50052 | `grpc://localhost:50052` | `docker compose exec user-service wget -qO- http://localhost:9091/health/ready` | `docker compose up -d --build user-service` |
| Product | 8082, 9093 | `http://localhost:8082`, `grpc://localhost:9093` | `curl http://localhost:8082/readyz` | `docker compose up -d --build product-service` |
| Cart | 8083 | `http://localhost:8083` | `curl http://localhost:8083/readyz` | `docker compose up -d --build cart-service` |
| Wishlist | 8084 | `http://localhost:8084` | `curl http://localhost:8084/readyz` | `docker compose up -d --build wishlist-service` |
| Search | 8085, 50055 | `http://localhost:8085`, `grpc://localhost:50055` | `curl http://localhost:8085/readyz` | `docker compose up -d --build search-service` |
| Session | 8086 | `http://localhost:8086` | `curl http://localhost:8086/healthz` | `docker compose up -d --build session-service` |
| CMS | 8087, 50057 | `http://localhost:8087`, `grpc://localhost:50057` | `curl http://localhost:8087/healthz` | `docker compose up -d --build cms-service` |
| Recommendation | 8089, 50058 | `http://localhost:8089`, `grpc://localhost:50058` | `curl http://localhost:8089/healthz` | `docker compose up -d --build recommendation-service` |
| Order | 8090, 9094 | `http://localhost:8090`, `grpc://localhost:9094` | `curl http://localhost:8090/readyz` | `docker compose up -d --build order-service` |
| Payment | 8091 | `http://localhost:8091` | `curl http://localhost:8091/healthz` | `docker compose up -d --build payment-service` |
| Notification | 8092, 50060 | `http://localhost:8092`, `grpc://localhost:50060` | `curl http://localhost:8092/readyz` | `docker compose up -d --build notification-service` |
| Superadmin | 8093 | `http://localhost:8093` | `curl http://localhost:8093/readyz` | `docker compose up -d --build superadmin-service` |
| User app | 3000 | `http://localhost:3000` | `curl http://localhost:3000/healthz` | `docker compose up -d --build user-app` |
| Seller dashboard | 3001 | `http://localhost:3001` | `curl http://localhost:3001/healthz` | `docker compose up -d --build seller-dashboard` |
| Session analytics dashboard | 3002 | `http://localhost:3002` | `curl http://localhost:3002/healthz` | `docker compose up -d --build session-analytics-dashboard` |
| Superadmin panel | 3003 | `http://localhost:3003` | `curl http://localhost:3003/healthz` | `docker compose up -d --build superadmin-panel` |

Notes:

- Gateway gRPC-Web is exposed on `http://localhost:8099`.
- Gateway metrics are exposed on `http://localhost:19090/metrics`.
- User service admin HTTP health runs on container port `9091`, but that port is not published to the host.

## Infrastructure

| Component | Host port(s) | Base URL / address | Health or verification |
| --- | ---: | --- | --- |
| MySQL | 3306 | `mysql://localhost:3306` | `docker compose exec mysql mysqladmin ping -h localhost -plocal_password` |
| MongoDB | 27017 | `mongodb://localhost:27017` | `docker compose exec mongodb mongosh --quiet --username local_admin --password local_mongo_password --authenticationDatabase admin --eval "db.adminCommand('ping')"` |
| Redis | 6379 | `redis://localhost:6379` | `docker compose exec redis redis-cli -a local_redis_password ping` |
| RabbitMQ | 5672, 15672 | `amqp://localhost:5672`, `http://localhost:15672` | Compose healthcheck |
| Kafka | 9092 | `localhost:9092` | `docker compose exec kafka /opt/kafka/bin/kafka-topics.sh --bootstrap-server localhost:29092 --list` |
| Typesense | 8108 | `http://localhost:8108` | `curl http://localhost:8108/health` |
| Mailpit | 1025, 8025 | `smtp://localhost:1025`, `http://localhost:8025` | Compose healthcheck |
| Jaeger | 16686, 4317 | `http://localhost:16686` | UI should load after compose start |
| OpenTelemetry Collector | Internal only | `otel-collector:4317` inside compose network | Check collector logs |
| Prometheus | 9095 | `http://localhost:9095` | UI should load after compose start |

## Compose Jobs And Workers

| Service | Port | Verification | Depends on | Start command |
| --- | ---: | --- | --- | --- |
| `init-mongodb-replica` | None | Job exit code | MongoDB | `docker compose up init-mongodb-replica` |
| `migrate-auth` | None | Job exit code | MySQL | `docker compose up migrate-auth` |
| `migrate-user` | None | Job exit code | MySQL | `docker compose up migrate-user` |
| `migrate-order` | None | Job exit code | MySQL | `docker compose up migrate-order` |
| `migrate-payment` | None | Job exit code | MySQL | `docker compose up migrate-payment` |
| `migrate-cms` | None | Job exit code | MySQL | `docker compose up migrate-cms` |
| `migrate-superadmin` | None | Job exit code | MySQL | `docker compose up migrate-superadmin` |
| `migrate-mongodb` | None | Job exit code | `init-mongodb-replica` | `docker compose up migrate-mongodb` |
| `migrate-session-mongodb` | None | Job exit code | `init-mongodb-replica` | `docker compose up migrate-session-mongodb` |
| `cart-expiry-worker` | None | Logs/job loop | MongoDB migrations, Redis | `docker compose --profile jobs up -d --build cart-expiry-worker` |
| `session-retention-worker` | None | Logs/job loop | Session Mongo migrations, Redis | `docker compose up -d --build session-retention-worker` |
| `payment-reconciliation` | None | Logs/job loop | Payment migrations | `docker compose --profile jobs up -d --build payment-reconciliation` |

## Makefile Shortcuts

| Command | What it starts |
| --- | --- |
| `make docker-up` | Full compose stack with build |
| `make infra-up` | `mysql`, `mongodb`, `redis`, `rabbitmq`, `kafka`, `typesense`, `mailpit`, `jaeger`, `prometheus` |
| `make backend-up` | Backend services and their compose dependencies |
| `make frontend-up` | Frontend apps |

Note: `make infra-up` does not directly start `init-mongodb-replica` or `otel-collector`. Use `docker compose up -d --build` for the complete normal local run.
