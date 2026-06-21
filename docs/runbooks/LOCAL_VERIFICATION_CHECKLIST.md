# Local Verification Checklist

## Containers and logs

```bash
docker compose config
docker compose ps
docker compose logs --tail=100 api-gateway
```

## Backend health

```bash
curl http://localhost:8080/health/live
curl http://localhost:8081/healthz
curl http://localhost:8082/readyz
curl http://localhost:8083/readyz
curl http://localhost:8084/readyz
curl http://localhost:8085/healthz
curl http://localhost:8086/readyz
curl http://localhost:8087/healthz
curl http://localhost:8089/healthz
curl http://localhost:8090/readyz
curl http://localhost:8091/healthz
curl http://localhost:8093/readyz
```

Expected caveat: `GET http://localhost:8080/health/ready` returns `503` while Pending downstream gRPC bridges remain incomplete.

## Infrastructure

```bash
docker compose exec redis redis-cli -a local_redis_password ping
docker compose exec rabbitmq rabbitmq-diagnostics ping
curl http://localhost:8108/health
```

Open frontend URLs `3000` through `3003`, Mailpit `8025`, RabbitMQ `15672`, Jaeger `16686`, and Prometheus `9095`.

Failure signs are restart loops, migration containers exiting non-zero, gateway liveness not `200`, or DB/queue health marked unhealthy. Check the named container logs first.
