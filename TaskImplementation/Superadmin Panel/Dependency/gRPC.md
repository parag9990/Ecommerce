# gRPC Dependency - Superadmin Panel

## 1. What is this dependency?

gRPC internal service communication protocol hai. Browser frontend REST call karta hai, API Gateway internally gRPC services ko call karta hai.

## 2. Why this service uses it

Superadmin Panel ke admin REST endpoints API Gateway par aate hain. Contract ke hisaab se gateway `SuperadminService`, `SessionService`, `OrderService`, and `PaymentService` gRPC methods call karega.

## 3. Required or optional

Required for backend runtime. Frontend ke liye direct gRPC required nahi hai. Current repo me gRPC contract partial hai, implementation files missing hain.

## 4. Where it is used in project

| Path | Use |
|------|-----|
| `api/master-api.json` | REST to gRPC mapping |
| `backend/services/api-gateway/.env` | `SUPERADMIN_GRPC_ADDR` and other service addresses |
| `docs/02-system-architecture.md` | Gateway to service gRPC architecture |
| `docs/04-microservice-design.md` | Superadmin gRPC methods |

## 5. Installation steps

Backend code not found. When implemented, Go dependencies usually include:

Suggested command based on Go gRPC setup:

```bash
go get google.golang.org/grpc google.golang.org/protobuf
```

Run this only inside the future Superadmin backend module.

## 6. Docker setup, if possible

No Docker Compose found.

Suggested service wiring if compose is later added:

```yaml
environment:
  SUPERADMIN_GRPC_ADDR: superadmin-service:9090
```

This is suggested only.

## 7. Local setup without Docker

Expected local flow:

1. Start Superadmin Service gRPC server.
2. Ensure it listens on `localhost:50062` or update gateway env.
3. Start API Gateway.
4. Frontend calls REST gateway.

Suggested command based on current env:

```bash
grpcurl -plaintext localhost:50062 list
```

This needs `grpcurl` installed and a running server.

## 8. Required environment variables

| Variable | Component | Purpose |
|----------|-----------|---------|
| `SUPERADMIN_GRPC_ADDR` | API Gateway | Address for Superadmin Service, found `localhost:50062` |
| `GRPC_TLS_ENABLED` | API Gateway | Enable/disable TLS |
| `GRPC_DIAL_TIMEOUT` | API Gateway | Dial timeout |
| `SESSION_GRPC_ADDR` | API Gateway | Session analytics methods |
| `ORDER_GRPC_ADDR` | API Gateway | Order admin list/detail |
| `PAYMENT_GRPC_ADDR` | API Gateway | Payment/refund workflows |

## 9. Start commands

Not clearly found in project files.

Suggested command based on project structure:

```bash
cd backend/services/superadmin-service
go run ./cmd/superadmin-service
```

## 10. Verify running commands

```bash
grpcurl -plaintext localhost:50062 list ecommerce.superadmin.v1.SuperadminService
```

Expected methods:

- `ListUsersForAdmin`
- `UpdateUserStatus`
- `ListSellersForAdmin`
- `UpdateSellerStatus`
- `ReviewRefund`
- `GetPlatformSettings`
- `UpdatePlatformSetting`
- `ListAuditLogs`

## 11. Common errors and fixes

| Error | Reason | Fix |
|-------|--------|-----|
| `connection refused` | Superadmin gRPC server not running | Start service or correct address |
| `unknown service` | Proto/server registration missing | Register `SuperadminService` |
| `deadline exceeded` | Timeout too low or service slow | Check `GRPC_DIAL_TIMEOUT` and backend health |
| `permission denied` | Admin context missing | Pass validated admin claims from gateway |

## 12. Security notes

- Internal gRPC should use TLS/mTLS in production.
- Gateway must forward admin context safely, not raw untrusted headers.
- gRPC methods must enforce service-level authorization too.
- Avoid logging PII or secrets in gRPC metadata.

## 13. Final checklist

- [x] gRPC contract found.
- [x] Gateway gRPC address env found.
- [ ] `.proto` file found.
- [ ] Generated protobuf code found.
- [ ] Superadmin gRPC server code found.
- [ ] Gateway gRPC client code found.

