# Step 17 - Developer Guide

## Local Setup Guide

Prerequisites:

- Go 1.24+
- Node.js 22+
- pnpm
- Docker Desktop
- kubectl
- buf CLI
- golangci-lint

Setup steps:

1. Clone repository.
2. Copy `.env.example` files to `.env`.
3. Start local dependencies with Docker Compose.
4. Run database migrations.
5. Generate proto clients.
6. Start backend services.
7. Start frontend apps.

Example commands:

```bash
docker compose -f infra/compose/docker-compose.local.yml up -d
buf generate
go test ./...
pnpm install
pnpm dev
```

## Backend Development Guide

Service coding flow:

1. Update proto contract.
2. Generate Go and TS clients.
3. Add domain entity.
4. Add repository interface.
5. Add usecase.
6. Add repository implementation.
7. Add gRPC handler.
8. Add gateway route if public.
9. Add tests.
10. Add observability.

Layer rules:

- Domain must not import transport or DB packages.
- Usecase depends on interfaces.
- Repository implements interfaces.
- Transport only maps request/response and calls usecase.
- Config and infrastructure stay outside domain.

## API Docs

Master API file:

- [master-api.json](../api/master-api.json)

REST response convention:

```json
{
  "data": {},
  "request_id": "req_123",
  "error": null
}
```

Error convention:

```json
{
  "data": null,
  "request_id": "req_123",
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid request",
    "details": []
  }
}
```

## Testing Strategy

Backend:

- Unit tests for domain and usecase.
- Repository tests with test containers.
- gRPC handler tests.
- Contract tests for proto compatibility.
- Integration tests for checkout/payment.

Frontend:

- Component tests.
- API mock tests.
- E2E tests for critical flows.

Critical E2E flows:

- Signup with OTP.
- Login and refresh.
- Product search.
- Add to cart.
- Checkout payment success.
- Payment failure retry.
- Seller creates product.
- Superadmin approves seller.
- Refund review.

## Branching and Release

Recommended:

- `main`: production-ready.
- `develop`: integration branch if team prefers Git Flow.
- `feature/*`: feature branches.
- `release/*`: release stabilization.
- `hotfix/*`: production fix.

Versioning:

- APIs versioned as `/api/v1`.
- Protos versioned as package `ecommerce.service.v1`.
- Docker images tagged with commit SHA and semantic release tag.

## Migration Guide

Rules:

- Every MySQL schema change has up/down migration.
- Backward compatible DB changes first.
- App deploy after DB deploy.
- Remove old columns only after code no longer reads them.
- Large migrations run in chunks.

## Coding Standards

Go:

- Context first argument.
- No panic for normal errors.
- Wrap errors with domain code.
- Use structured logger.
- Add deadlines to outbound calls.

TypeScript:

- Strict mode.
- No `any` unless justified.
- API types generated or explicitly declared.
- Use React Query for async server data.
- Keep components focused.

## Production Readiness Checklist

- Health endpoints.
- Readiness and liveness probes.
- Graceful shutdown.
- Config validation at startup.
- Structured logs.
- Metrics.
- Traces.
- Rate limits.
- Auth and RBAC.
- Input validation.
- Idempotency for commands.
- Retry and DLQ for consumers.
- DB migrations.
- Secret management.
- Load test for critical paths.

## Future Scaling Plan

Near term:

- Complete MVP service skeleton.
- Implement auth/user/product/cart/order/payment.
- Add search and session ingestion.
- Build user app and seller dashboard.

Mid term:

- Add recommendation engine.
- Add superadmin workflows.
- Add analytics dashboard.
- Add production Kubernetes and monitoring.

Long term:

- Multi-region reads.
- Advanced ML recommendations.
- Fraud detection.
- Event-sourced read models for analytics.
- Service mesh and mTLS.

