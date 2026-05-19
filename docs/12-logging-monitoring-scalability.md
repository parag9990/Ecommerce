# Step 14 and Step 16 - Logging, Monitoring, and Scalability

## Centralized Logging

Use structured JSON logs.

Required fields:

- timestamp
- level
- service
- environment
- request_id
- trace_id
- user_id_hash
- route or grpc_method
- message
- error_code
- latency_ms

Recommended stack:

- Loki + Grafana for logs.
- Fluent Bit or OpenTelemetry Collector to ship logs.
- Sentry-like service for frontend/backend error tracking.

Do not log:

- passwords
- OTP
- full JWT
- refresh token
- card data
- secrets

## Metrics Collection

Use Prometheus metrics.

Service metrics:

- request count
- request latency p50/p95/p99
- error rate
- gRPC status codes
- DB query latency
- Redis latency
- queue publish/consume count
- consumer lag

Business metrics:

- signup rate
- login failures
- product views
- add to cart rate
- checkout started
- payment success rate
- order cancellation rate
- refund rate

## Distributed Tracing

Use OpenTelemetry.

Trace flows:

- search request
- checkout
- payment webhook
- seller product publish
- admin refund review

Every service propagates:

- `traceparent`
- `x-request-id`
- auth context metadata where allowed

## Alerts

| Alert | Example Threshold |
|---|---|
| Gateway 5xx | > 2 percent for 5 min |
| Checkout p95 latency | > 2 seconds for 5 min |
| Payment webhook failures | > 1 percent for 10 min |
| Queue lag | > 10,000 messages or growing for 15 min |
| DB CPU | > 80 percent for 10 min |
| Redis memory | > 85 percent |
| Typesense latency | p95 > 500 ms |
| Auth login failures | sudden spike by IP/region |

## Error Tracking

Backend:

- capture panics
- capture unhandled errors
- include request id and trace id
- redact PII

Frontend:

- capture runtime errors
- capture route and release version
- capture API error code
- redact user input

## High Traffic Handling

Traffic strategy:

- CDN for static assets and product images.
- API Gateway horizontal scaling.
- Product read cache.
- Search handled by Typesense cluster.
- Cart hot state in Redis.
- Async events for notifications, recommendations, analytics.
- Read replicas for MySQL read-heavy paths.

## Caching Strategy

| Data | Cache | TTL |
|---|---|---:|
| Product summary | Redis/CDN edge where safe | 5 to 15 min |
| Category tree | Redis | 30 min |
| Cart count | Redis | 5 min |
| Search autocomplete | Redis | 1 to 5 min |
| Seller settings | Redis | 10 min |
| Platform settings | Redis | 5 min |
| Recommendations | Redis | 15 min |

Invalidation:

- Product update event invalidates product cache.
- Category update invalidates category tree.
- Seller settings update invalidates seller cache.
- Platform settings update invalidates gateway/admin cache.

## Horizontal Scaling

Stateless services scale with replicas.

HPA examples:

- CPU target 60 percent.
- Memory target 70 percent.
- Custom RPS metric.
- Queue lag for consumers.

Stateful components:

- Managed MySQL with read replicas.
- Managed MongoDB replica set/sharding.
- Managed Redis cluster.
- Typesense multi-node cluster.
- Kafka managed cluster or durable RabbitMQ cluster.

## Checkout Scalability

Checkout is sensitive path.

Rules:

- Keep synchronous calls minimal.
- Use deadlines for Product, Cart, Payment.
- Idempotency mandatory.
- Use inventory reservation TTL.
- Payment finalization async via webhook.
- Graceful degradation: recommendations/analytics failure must not block checkout.

## Failure Isolation

If Recommendation Service down:

- Product page still loads.
- Show fallback trending from cache or hide section.

If Notification Service down:

- Order still succeeds.
- Notification command retries later.

If Session Service down:

- User app still works.
- SDK buffers or drops analytics events depending policy.

If Search down:

- Category browse can fall back to Product Service list endpoint.

## Future Scaling Plan

Phase 1:

- Single region.
- Managed DBs.
- Basic HPA.
- Redis cache.
- Kafka/RabbitMQ.

Phase 2:

- Read replicas.
- Product cache.
- Session event sharding.
- Search cluster.
- Canary deploy.

Phase 3:

- Multi-region read replicas.
- Global CDN.
- Event-driven read models.
- Service mesh mTLS.
- ML recommendation pipeline.

