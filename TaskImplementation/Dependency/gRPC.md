# User Service - gRPC Dependency

## 1. What is this dependency?

gRPC ek high-performance internal API framework hai. Isme services and methods protobuf contract se define hote hain, phir generated code se type-safe client/server ban jaate hain.

## 2. Why User Service uses it

User Service browser ko directly serve nahi karta. Internal services or API Gateway User Service ko gRPC se call karte hain.

Current implemented service:

```text
ecommerce.user.v1.UserService
```

Current implemented methods:

| Method | Purpose |
|---|---|
| `CreateUser` | Auth signup ke baad profile create |
| `GetUser` | Profile fetch |
| `UpdateUserProfile` | Profile update with field mask |
| `GetSellerProfile` | Seller profile fetch |

## 3. Required or Optional

| Item | Required? | Notes |
|---|---:|---|
| `google.golang.org/grpc` | Yes | Runtime server, status codes, interceptors |
| gRPC reflection | Optional | Local `grpcurl list/describe` ke liye useful |
| `grpcurl` CLI | Optional but recommended | Manual gRPC testing |
| gRPC health service | Missing | Not implemented currently |

## 4. Where it is used in project

| Path | Purpose |
|---|---|
| `backend/services/user-service/cmd/server/main.go` | gRPC server create/register/start |
| `backend/services/user-service/internal/transport/grpc/server.go` | gRPC method handlers |
| `backend/services/user-service/internal/transport/grpc/interceptors.go` | Logging and recovery unary interceptors |
| `backend/services/user-service/internal/transport/grpc/errors.go` | Domain error to gRPC status mapping |
| `backend/services/user-service/internal/transport/grpc/metadata.go` | Metadata extraction |
| `backend/services/user-service/internal/transport/grpc/mapper.go` | Domain to proto mapping |
| `proto/ecommerce/user/v1/user.proto` | gRPC contract |
| `backend/shared/gen/go/ecommerce/user/v1/user_grpc.pb.go` | Generated server/client interfaces |

## 5. Installation Steps

Go module dependency is already present in `go.mod`.

Install manual test tool:

```bash
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
```

Verify:

```bash
grpcurl -version
```

## 6. Docker Setup, If Possible

No service Dockerfile was found. If containerized later, expose gRPC port `50052` and pass env vars:

```text
EXPOSE 50052
USER_SERVICE_GRPC_ADDRESS=:50052
```

For local now, run service directly with Go.

## 7. Local Setup Without Docker

1. Start MySQL.
2. Apply migrations.
3. Export env vars.
4. Start backend.

Suggested command based on project structure:

```bash
cd backend/services/user-service
set -a
. ./.env
set +a
go run ./cmd/server
```

## 8. Required Environment Variables

| Variable | Required | Default | Purpose |
|---|---:|---|---|
| `USER_SERVICE_GRPC_ADDRESS` | No | `:50052` | gRPC listen address |
| `USER_SERVICE_GRPC_REFLECTION` | No | `true` | Enables reflection |
| `USER_SERVICE_SHUTDOWN_TIMEOUT` | No | `10s` | Graceful stop timeout |
| `USER_SERVICE_DATABASE_DSN` | Yes | None | Required before gRPC server starts, because DB ping happens first |
| `USER_SERVICE_LOG_LEVEL` | No | `info` | Affects gRPC request logs |

Metadata read by current gRPC handlers:

| Metadata | Purpose |
|---|---|
| `x-user-id` | Caller user id for self/permission check |
| `x-service-name` | Internal service caller name |
| `x-roles` | Comma-separated caller roles |
| `x-request-id` | Request id in logs |

## 9. Start Commands

Start:

```bash
cd backend/services/user-service
set -a
. ./.env
set +a
go run ./cmd/server
```

If port busy:

```bash
export USER_SERVICE_GRPC_ADDRESS=':50053'
go run ./cmd/server
```

## 10. Verify Running Commands

Reflection list:

```bash
grpcurl -plaintext localhost:50052 list
```

Describe service:

```bash
grpcurl -plaintext localhost:50052 describe ecommerce.user.v1.UserService
```

Example call:

```bash
grpcurl -plaintext \
  -H 'x-user-id: user_123' \
  -d '{"user_id":"user_123"}' \
  localhost:50052 ecommerce.user.v1.UserService/GetUser
```

This call requires matching data in MySQL.

## 11. Common Errors and Fixes

| Error | Cause | Fix |
|---|---|---|
| `connection refused` | Service not running or wrong port | Start service and verify `USER_SERVICE_GRPC_ADDRESS` |
| `server does not support reflection` | Reflection disabled | Set `USER_SERVICE_GRPC_REFLECTION=true` for local |
| `invalid request` | Missing required field or invalid field mask | Check request JSON |
| `permission denied` | `x-user-id` does not match requested user | Send correct metadata or internal service flow |
| `deadline exceeded` | DB/server slow or client timeout too low | Check MySQL health and client timeout |
| `listen tcp :50052 bind address already in use` | Port busy | Use another port like `:50053` |

## 12. Security Notes

- Public internet pe raw gRPC port expose mat karo.
- Reflection local dev me useful hai, production me internal-only or disabled rakho.
- Metadata from Gateway/Auth must be trusted only after authentication.
- Logs me request IDs and method names ok hain; PII fields log mat karo.
- Add gRPC health service before production readiness.

## 13. Final Checklist

| Check | Done |
|---|---|
| MySQL running and migrated | [ ] |
| Env vars exported | [ ] |
| Service listening on `50052` | [ ] |
| Reflection enabled locally | [ ] |
| `grpcurl list` works | [ ] |
| gRPC methods return expected status codes | [ ] |
| Port not publicly exposed | [ ] |

