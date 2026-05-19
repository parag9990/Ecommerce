# Step 12 and Step 13 - DevOps and External Services

## Docker Setup

Each Go service Dockerfile:

```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.work go.work.sum ./
COPY backend/shared ./backend/shared
COPY backend/services/auth-service ./backend/services/auth-service
WORKDIR /app/backend/services/auth-service
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/service ./cmd/server

FROM gcr.io/distroless/static-debian12
COPY --from=builder /out/service /service
EXPOSE 8080 9090
ENTRYPOINT ["/service"]
```

Frontend Dockerfile:

```dockerfile
FROM node:22-alpine AS builder
WORKDIR /app
COPY frontend/package.json frontend/pnpm-lock.yaml ./
RUN corepack enable && pnpm install --frozen-lockfile
COPY frontend ./
RUN pnpm --filter user-app build

FROM nginx:alpine
COPY --from=builder /app/user-app/dist /usr/share/nginx/html
COPY infra/docker/nginx.conf /etc/nginx/conf.d/default.conf
```

## Local Docker Compose

Local stack should include:

- MySQL
- MongoDB
- Redis
- Typesense
- Kafka or RabbitMQ
- Jaeger/Tempo
- Prometheus
- Grafana
- Mailpit for email testing

## Kubernetes Deployment

Each service needs:

- Deployment
- Service
- ConfigMap
- Secret reference
- HorizontalPodAutoscaler
- PodDisruptionBudget
- ServiceMonitor

Example deployment shape:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: product-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: product-service
  template:
    metadata:
      labels:
        app: product-service
    spec:
      containers:
        - name: product-service
          image: registry.example.com/product-service:latest
          ports:
            - containerPort: 8080
            - containerPort: 9090
          readinessProbe:
            httpGet:
              path: /health/ready
              port: 8080
          livenessProbe:
            httpGet:
              path: /health/live
              port: 8080
```

## Service Discovery

Kubernetes DNS:

- `auth-service.core.svc.cluster.local:9090`
- `product-service.core.svc.cluster.local:9090`
- `order-service.core.svc.cluster.local:9090`

Gateway config maps service names to gRPC targets.

## Scaling Strategy

| Component | Scaling Signal |
|---|---|
| API Gateway | CPU, RPS, latency p95 |
| Product Service | CPU, DB latency, read RPS |
| Search Service | search latency, Typesense CPU |
| Cart Service | Redis latency, mutation RPS |
| Order Service | checkout RPS, queue lag |
| Payment Service | webhook lag, provider latency |
| Session Service | event ingestion RPS, queue lag |
| Consumers | Kafka/RabbitMQ lag |

## CI/CD

Recommended pipeline:

1. Lint.
2. Unit tests.
3. Proto lint and breaking-change check.
4. Integration tests.
5. Build Docker images.
6. Scan images.
7. Push registry.
8. Deploy dev.
9. Smoke tests.
10. Deploy staging.
11. Manual production approval.
12. Canary production deploy.

## Typesense Setup

Local:

```yaml
typesense:
  image: typesense/typesense:latest
  command: ["--data-dir", "/data", "--api-key", "dev-typesense-key", "--enable-cors"]
  ports:
    - "8108:8108"
```

Integration:

- Search Service owns Typesense schema.
- Product events update index.
- Full reindex job reads Product Service pages and upserts to Typesense.
- API key stored in secret manager.

## Payment Gateway Setup

Provider config:

- public key for frontend payment SDK.
- secret key for backend provider API.
- webhook secret for signature verification.
- allowed currencies.
- capture mode.

Integration guide:

1. Payment Service creates intent.
2. Frontend uses provider SDK/client secret.
3. Provider redirects/callbacks frontend.
4. Provider webhook updates Payment Service.
5. Order Service reacts to payment event.

## Redis Setup

Use cases:

- rate limiting
- active sessions
- cart cache
- recommendation cache
- OTP retry counters
- distributed locks where needed

Production:

- Use managed Redis or Redis cluster.
- Enable auth/TLS.
- Set memory policy carefully.

## Message Queue Setup

Kafka recommended for high-throughput event streaming. RabbitMQ acceptable for simpler routing and command queues.

Topics/queues:

- `user.events`
- `auth.events`
- `product.events`
- `order.events`
- `payment.events`
- `session.events`
- `notification.commands`
- `search.indexing`
- `recommendation.events`

Event envelope:

```json
{
  "event_id": "evt_123",
  "event_type": "OrderPaid",
  "version": 1,
  "occurred_at": "2026-05-18T14:00:00Z",
  "producer": "order-service",
  "trace_id": "trace_123",
  "payload": {}
}
```

Consumer rules:

- Idempotent event handling.
- Dead-letter queue.
- Retry with backoff.
- Schema versioning.

