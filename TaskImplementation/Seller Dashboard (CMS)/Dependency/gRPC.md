# gRPC Dependency - Seller Dashboard (CMS)

## 1. What is this dependency?

gRPC ek fast, typed service-to-service communication protocol hai. Is project design me browser REST call API Gateway tak jaati hai, phir Gateway internal services ko gRPC se call karta hai.

## 2. Why this service uses it

Seller Dashboard APIs multiple backend services par depend karte hain:

- `ProductService` for products
- `OrderService` for seller orders
- `CMSService` for coupons/campaigns/analytics/settings
- `UserService` and Auth boundary for seller session/profile

Gateway in services ko gRPC target addresses se call karega.

## 3. Required or optional

Required for current microservice architecture.

Frontend direct gRPC use nahi kar raha; frontend REST -> Gateway -> gRPC services flow follow karta hai.

## 4. Where it is used in project

| Path | Use |
|---|---|
| `backend/services/api-gateway/.env` | Downstream `*_GRPC_ADDR` values |
| `backend/services/cms-service/.env` | CMS gRPC server listen address |
| `api/master-api.json` | REST routes map to gRPC methods |
| `docs/02-system-architecture.md` | gRPC communication rules |
| `docs/04-microservice-design.md` | service gRPC method lists |

## 5. Installation steps

Go gRPC deps are not clearly installed because service `go.mod` files are missing.

Suggested command based on project structure:

```bash
cd backend/services/cms-service
go get google.golang.org/grpc
go get google.golang.org/protobuf
```

Run only after the service module exists.

## 6. Docker setup, if possible

Not clearly found in project files.

When Docker is added, expose internal gRPC ports and keep them private to backend network.

## 7. Local setup without Docker

Expected local flow:

1. Start CMS gRPC server.
2. Start Product/Order/User/Auth gRPC servers.
3. Start API Gateway.
4. Gateway dials the configured gRPC targets.

Current blocker:

- gRPC server/client implementation files were not clearly found.

## 8. Required environment variables

Gateway:

| Variable | Purpose |
|---|---|
| `GRPC_TLS_ENABLED` | TLS toggle |
| `GRPC_DIAL_TIMEOUT` | Dial timeout |
| `PRODUCT_GRPC_ADDR` | Product target |
| `ORDER_GRPC_ADDR` | Order target |
| `CMS_GRPC_ADDR` | CMS target |
| `USER_GRPC_ADDR` | User target |
| `AUTH_GRPC_ADDR` | Auth target |

CMS:

| Variable | Purpose |
|---|---|
| `CMS_GRPC_ADDR` | CMS gRPC listen address |
| `CMS_GRPC_MAX_RECV_MSG_BYTES` | Max receive size |
| `CMS_GRPC_MAX_SEND_MSG_BYTES` | Max send size |
| `CMS_GRPC_ALLOWED_INTERNAL_CALLERS` | Allowed caller allowlist |

## 9. Start commands

Suggested command based on project structure:

```bash
cd backend/services/cms-service
go run ./cmd/server
```

This should start gRPC after server registration exists.

## 10. Verify running commands

```bash
# Suggested command based on project structure
grpcurl -plaintext localhost:9098 list
```

If reflection is disabled, use generated proto descriptors/client tests instead.

## 11. Common errors and fixes

| Error | Reason | Fix |
|---|---|---|
| `connection refused` | Service not running | Start downstream service |
| `deadline exceeded` | Slow service or wrong network | Check timeout and service health |
| `unimplemented` | Method missing in server | Register service handler |
| `unauthenticated` | Internal metadata missing | Gateway should pass auth context/internal token |
| CMS unavailable | Port mismatch `50059` vs `9098` | Align Gateway and CMS env |

## 12. Security notes

- gRPC calls should have deadlines.
- Production internal calls should use TLS/mTLS.
- Pass only necessary auth context metadata.
- Never trust frontend-supplied seller IDs without backend ownership validation.
- Map gRPC errors to safe REST errors for frontend.

## 13. Final checklist

- [x] gRPC env config found.
- [x] REST-to-gRPC mapping found in `api/master-api.json`.
- [x] CMS gRPC listen config found.
- [ ] gRPC source/server registration not found.
- [ ] gRPC client code not found.
- [ ] Port mismatch needs fix.

