# Protobuf Dependency - Superadmin Panel

## 1. What is this dependency?

Protobuf gRPC ke request/response contracts define karta hai. `.proto` files se Go generated code banta hai.

## 2. Why this service uses it

`api/master-api.json` me `ecommerce.superadmin.v1.SuperadminService` package and methods documented hain. In methods ke liye `.proto` contracts needed hain.

## 3. Required or optional

Required for backend gRPC implementation. Current repo me missing.

## 4. Where it is used in project

| Path | Use |
|------|-----|
| `api/master-api.json` | Method names and schema names |
| `docs/03-folder-structure.md` | Expected `proto/superadmin/v1/superadmin.proto` path |
| `docs/04-microservice-design.md` | Superadmin gRPC service list |

Not clearly found in project files:

- `proto/superadmin/v1/superadmin.proto`
- generated Go files like `superadmin.pb.go`
- generated gRPC files like `superadmin_grpc.pb.go`

## 5. Installation steps

Suggested tools:

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

This is suggested only because proto source is missing.

## 6. Docker setup, if possible

No proto generation Docker setup found. Agar Docker use karna ho, `protoc` image se generation run kiya ja sakta hai.

Suggested command based on common proto setup:

```bash
docker run --rm -v "$PWD":/workspace -w /workspace namely/protoc-all
```

This is suggested only and may need project-specific changes.

## 7. Local setup without Docker

Expected flow after proto files are added:

```bash
protoc \
  --go_out=. \
  --go-grpc_out=. \
  proto/superadmin/v1/superadmin.proto
```

This command is suggested. It needs actual proto layout and `go_package` options.

## 8. Required environment variables

Protobuf generation usually does not require runtime env vars. gRPC runtime uses:

| Variable | Purpose |
|----------|---------|
| `SUPERADMIN_GRPC_ADDR` | Gateway target for generated client |
| `GRPC_TLS_ENABLED` | TLS behavior |

## 9. Start commands

Protobuf itself does not start. Generate code, then start backend services.

Suggested command based on missing proto:

```bash
protoc --version
```

## 10. Verify running commands

```bash
find . -name "*.proto"
```

Current status from project scan: no `.proto` files clearly found.

After generation:

```bash
find backend -name "*superadmin*pb.go"
```

## 11. Common errors and fixes

| Error | Reason | Fix |
|-------|--------|-----|
| `protoc: command not found` | Protobuf compiler missing | Install protoc |
| `protoc-gen-go: program not found` | Go plugin missing | Install `protoc-gen-go` |
| Bad import paths | Missing `go_package` | Add correct `option go_package` |
| Gateway/client mismatch | Proto not synced with API contract | Keep `api/master-api.json` and proto aligned |

## 12. Security notes

- Do not add secret defaults inside proto comments/examples.
- Avoid exposing internal-only admin fields to public contracts.
- Use explicit auth annotations/docs for admin-only methods.

## 13. Final checklist

- [ ] `superadmin.proto` exists.
- [ ] Request/response messages defined.
- [ ] Generated Go code committed or reproducible.
- [ ] Gateway client uses generated code.
- [ ] Server registers generated service interface.

