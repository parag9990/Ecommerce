# User Service - Protobuf and Buf Dependency

## 1. What is this dependency?

Protocol Buffers, ya protobuf, ek contract/schema format hai. Isme request, response, and service methods define hote hain. Buf ek proto tooling hai jo linting and code generation easy banata hai.

## 2. Why User Service uses it

User Service ka internal API `user.proto` me defined hai. Go generated files service server/client interfaces provide karte hain.

## 3. Required or Optional

| Item | Required? | Notes |
|---|---:|---|
| `.proto` contract | Yes | API source of truth |
| Generated Go code | Yes | User Service compile/runtime ke liye |
| Buf CLI | Required when proto changes | Current generated code already exists |
| `protoc-gen-go` | Required when generating locally without remote plugins | Buf config uses remote plugins |
| `protoc-gen-go-grpc` | Required when generating locally without remote plugins | Buf config uses remote plugins |

## 4. Where it is used in project

| Path | Purpose |
|---|---|
| `proto/ecommerce/user/v1/user.proto` | User Service protobuf contract |
| `proto/buf.yaml` | Buf module, lint, breaking config |
| `proto/buf.gen.yaml` | Go generation config |
| `backend/shared/gen/go/ecommerce/user/v1/user.pb.go` | Generated protobuf messages |
| `backend/shared/gen/go/ecommerce/user/v1/user_grpc.pb.go` | Generated gRPC client/server interfaces |
| `backend/shared/gen/go/go.mod` | Generated Go module |
| `backend/services/user-service/go.mod` | Replaces generated module to local path |

## 5. Installation Steps

Install Buf:

```bash
go install github.com/bufbuild/buf/cmd/buf@latest
```

Verify:

```bash
buf --version
```

Optional local generator binaries:

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

## 6. Docker Setup, If Possible

No proto generation Docker workflow was found.

Suggested future approach:

```bash
docker run --rm -v "$PWD":/workspace -w /workspace/proto bufbuild/buf generate
```

This is suggested only. Current repo has local Buf config but no Docker wrapper.

## 7. Local Setup Without Docker

Generate code after proto changes:

```bash
cd proto
buf generate
```

Then test User Service:

```bash
cd ../backend/services/user-service
go test ./...
```

## 8. Required Environment Variables

Protobuf generation does not require User Service runtime env vars.

Runtime still needs:

| Variable | Purpose |
|---|---|
| `USER_SERVICE_DATABASE_DSN` | Service startup DB connection |
| `USER_SERVICE_GRPC_ADDRESS` | gRPC listen address |

## 9. Start Commands

Proto generation command:

```bash
cd proto
buf generate
```

Backend start after generation:

```bash
cd backend/services/user-service
set -a
. ./.env
set +a
go run ./cmd/server
```

## 10. Verify Running Commands

Lint proto:

```bash
cd proto
buf lint
```

Verify generated files exist:

```bash
test -f ../backend/shared/gen/go/ecommerce/user/v1/user.pb.go
test -f ../backend/shared/gen/go/ecommerce/user/v1/user_grpc.pb.go
```

Verify service contract at runtime:

```bash
grpcurl -plaintext localhost:50052 describe ecommerce.user.v1.UserService
```

## 11. Common Errors and Fixes

| Error | Cause | Fix |
|---|---|---|
| `buf: command not found` | Buf not installed or PATH missing | Install Buf and add Go bin to PATH |
| Generated code unchanged | Ran command from wrong folder | Run from `proto/` |
| Go compile fails after proto change | Generated code stale | Run `buf generate`, then `go test ./...` |
| Method missing in grpcurl | Proto/server not updated or service not restarted | Regenerate, implement handler, restart service |
| Import path mismatch | `go_package` or output path wrong | Keep `go_package` aligned with `backend/shared/gen/go` |

## 12. Security Notes

- Proto contracts are API contracts. Breaking changes should be versioned, not silently changed.
- Do not put secrets or passwords in protobuf messages.
- Reflection is useful locally but can expose method names internally.
- Generated code should be reviewed after proto changes.

## 13. Final Checklist

| Check | Done |
|---|---|
| `user.proto` exists | [ ] |
| `buf.yaml` exists | [ ] |
| `buf.gen.yaml` exists | [ ] |
| Generated Go files exist | [ ] |
| `backend/shared/gen/go` module exists | [ ] |
| User Service `go.mod` has local replace | [ ] |
| `buf lint` passes after changes | [ ] |
| `go test ./...` passes after generation | [ ] |

