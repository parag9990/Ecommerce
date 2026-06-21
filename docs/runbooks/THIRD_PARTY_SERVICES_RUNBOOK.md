# Third-party Services Runbook

| Service | Why used | Compose name | Port / UI | Env | Quick test |
|---|---|---|---|---|---|
| Redis | rate limits, OTP/cart/session/cache | `redis` | 6379 | `*_REDIS_*` | `docker compose exec redis redis-cli -a local_redis_password ping` |
| RabbitMQ | user/product/notification/search events | `rabbitmq` | 5672 / http://localhost:15672 | `RABBITMQ_URL` | `docker compose exec rabbitmq rabbitmq-diagnostics ping` |
| Kafka | order/wishlist/recommendation events | `kafka` | 9092 | `KAFKA_BROKERS` | `docker compose exec kafka /opt/kafka/bin/kafka-topics.sh --bootstrap-server localhost:29092 --list` |
| Typesense | product search | `typesense` | http://localhost:8108 | `TYPESENSE_*` | `curl http://localhost:8108/health` |
| Mailpit | captures local SMTP | `mailpit` | 1025 / http://localhost:8025 | `NOTIFICATION_EMAIL_SMTP_*` | open UI after sending OTP/email |
| Jaeger | OTLP traces | `jaeger` | 4317 / http://localhost:16686 | `TRACE_EXPORTER_OTLP_ENDPOINT` | open UI |
| Prometheus | metrics scraping | `prometheus` | http://localhost:9095 | metrics addresses | open `/targets` |

MySQL and MongoDB are covered in `DATABASE_LOCAL_RUNBOOK.md`. MinIO/S3 is not included because no runtime S3 client/config was found. Payment providers are disabled locally until real sandbox endpoints are supplied.

Common error: containers use service DNS names, while host-run applications use `localhost`. RabbitMQ and MongoDB also require their local credentials.
