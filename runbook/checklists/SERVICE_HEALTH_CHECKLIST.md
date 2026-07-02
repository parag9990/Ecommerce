# Service Health Checklist

Use after `docker compose up -d --build`. These commands match the root `docker-compose.yml` healthcheck paths where compose defines one.

## Gateway

- [ ] `curl http://localhost:8080/health/live`
- [ ] `curl http://localhost:8080/health/ready`

## Backend Services

- [ ] `curl http://localhost:8081/readyz`
- [ ] `docker compose exec user-service wget -qO- http://localhost:9091/health/ready`
- [ ] `curl http://localhost:8082/readyz`
- [ ] `curl http://localhost:8083/readyz`
- [ ] `curl http://localhost:8084/readyz`
- [ ] `curl http://localhost:8085/readyz`
- [ ] `curl http://localhost:8086/healthz`
- [ ] `curl http://localhost:8087/healthz`
- [ ] `curl http://localhost:8089/healthz`
- [ ] `curl http://localhost:8090/readyz`
- [ ] `curl http://localhost:8091/healthz`
- [ ] `curl http://localhost:8092/readyz`
- [ ] `curl http://localhost:8093/readyz`

## Frontend Apps

- [ ] `curl http://localhost:3000/healthz`
- [ ] `curl http://localhost:3001/healthz`
- [ ] `curl http://localhost:3002/healthz`
- [ ] `curl http://localhost:3003/healthz`

## Infrastructure

- [ ] `docker compose exec mysql mysqladmin ping -h localhost -plocal_password`
- [ ] `docker compose exec mongodb mongosh --quiet --username local_admin --password local_mongo_password --authenticationDatabase admin --eval "db.adminCommand('ping')"`
- [ ] `docker compose exec redis redis-cli -a local_redis_password ping`
- [ ] `docker compose exec kafka /opt/kafka/bin/kafka-topics.sh --bootstrap-server localhost:29092 --list`
- [ ] `curl http://localhost:8108/health`
- [ ] Open `http://localhost:8025`
- [ ] Open `http://localhost:15672`
