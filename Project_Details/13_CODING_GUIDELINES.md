# Coding Guidelines

This file documents style and conventions observed in the current codebase. It does not introduce new rules beyond what exists.

## Backend Style

| Area | Existing convention |
| --- | --- |
| Language | Go |
| Workspace | `backend/go.work` with service-specific modules |
| Formatting | `gofmt` enforced by CI |
| Static checks | `go vet` and `golangci-lint` in CI |
| Service layout | `cmd`, `internal/config`, `internal/domain`, `internal/usecase`, `internal/repository`, `internal/transport`, `migrations` |
| API contracts | Protobuf under `proto/ecommerce/*` and REST catalog in `api/master-api.json` |
| Logging | Structured logging with Go `slog` patterns |
| Context | Request-scoped context passed through usecases/repositories |

## Backend Naming And Folder Conventions

- Service code lives under `backend/services/<service-name>`.
- Public service entrypoints live under `cmd`.
- Business/domain types live under `internal/domain`.
- Usecases contain application workflows.
- Repositories encapsulate persistence.
- Transport packages own HTTP/gRPC request mapping.
- Service-owned migrations live under each service `migrations` folder.
- Internal HTTP routes use `/internal/v1/...` or `/internal/admin/...` patterns.
- Public REST routes use `/api/v1/...`.

## Frontend Style

| Area | Existing convention |
| --- | --- |
| Language | TypeScript |
| Framework | React with Vite |
| Routing | React Router |
| Server state | React Query |
| Client/session state | Zustand or session storage depending on app |
| Testing | Vitest and Testing Library |
| Icons | `lucide-react` is present in frontend apps |

## API Response And Error Patterns

Observed patterns:

- Auth service returns errors in an envelope like `{ "error": { "code", "message" } }`.
- Frontend HTTP helpers commonly unwrap `{ data }` envelopes when present.
- Gateway adds request IDs and maps catalog/service responses.
- gRPC handlers map domain errors to gRPC status codes.

A single platform-wide response envelope standard document was not found in current codebase.

## Error Handling Pattern

Observed patterns:

- Domain/usecase errors are mapped at the transport boundary.
- HTTP handlers validate request input before calling usecases.
- gRPC handlers validate metadata/actor context and translate to status errors.
- Payment/order/product internal routes use service tokens for trusted calls.
- Gateway enforces route auth before downstream dispatch.

## Logging Pattern

Observed patterns:

- Services initialize structured loggers.
- Gateway includes access logging middleware.
- Payment provider integrations log unsupported capture calls.
- Observability stack supports metrics/tracing through Prometheus and OpenTelemetry.

## Testing Pattern

| Area | Found |
| --- | --- |
| Backend | Go tests across service modules; CI runs module tests |
| Frontend | Vitest tests under frontend apps; CI runs recursive frontend checks |
| Proto | Buf lint and breaking checks |
| Containers | Docker build and Trivy scan in CI |

## How To Safely Add New Features

1. Update the relevant API contract first: protobuf for gRPC, `api/master-api.json` for REST Gateway routes.
2. Add or update service-owned domain/usecase/repository code within the owning service.
3. Add migrations in the owning service only.
4. Wire transport handlers at service boundary.
5. Update API Gateway route mapping/auth rules if the route is public.
6. Update frontend API clients and route usage if exposed to UI.
7. Add focused tests for domain/usecase and transport behavior.
8. Regenerate protobuf clients when proto files change.
9. Run service tests, frontend checks if touched, and relevant compose smoke checks.

## Dependency Guidelines Observed

- Prefer service APIs over cross-service database access.
- Use internal tokens for service-to-service internal HTTP/gRPC methods.
- Keep database schema ownership within the service.
- Use outbox/event publisher patterns for asynchronous propagation where already present.

## Missing Standards

- Dedicated style guide document: Not found in current codebase.
- Platform-wide API envelope specification: Not found in current codebase.
- Branching/release policy beyond existing docs and CI: Not fully found in current codebase.
- Security review checklist: Not found in current codebase.
