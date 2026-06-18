# Protobuf Dependency - Seller Dashboard (CMS)

## 1. What is this dependency?

Protobuf, ya Protocol Buffers, gRPC messages and services define karne ka schema format hai. Ye typed contracts generate karta hai so Go backend and optional TypeScript clients consistent rahen.

## 2. Why this service uses it

Seller Dashboard backend routes gRPC services par map hote hain. CMS Service methods jaise `CreateCoupon`, `ListCoupons`, `GetSellerAnalytics` protobuf contract me define hone chahiye.

## 3. Required or optional

Required for gRPC-based backend architecture.

## 4. Where it is used in project

| Path | Use |
|---|---|
| `docs/03-folder-structure.md` | Expected `proto/ecommerce/cms/v1/cms.proto` |
| `docs/02-system-architecture.md` | Protobuf versioning rules |
| `api/master-api.json` | Method names and schemas used as contract reference |

Actual `proto/` directory was not clearly found in project files.

## 5. Installation steps

Suggested tools after proto files are added:

```bash
# Suggested command based on project structure
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

If Buf is chosen:

```bash
# Suggested command based on project structure
buf generate
```

## 6. Docker setup, if possible

Not clearly found in project files.

Proto generation usually runs in local dev or CI. Docker can be used later with a generator image, but no current setup was found.

## 7. Local setup without Docker

Expected flow:

1. Add proto files under `proto/ecommerce/cms/v1/`.
2. Add shared proto files for pagination, money, errors, auth context.
3. Run generator.
4. Commit generated Go code or generate during build, depending project policy.
5. Update API Gateway/CMS imports.

## 8. Required environment variables

Protobuf generation usually does not need runtime env vars.

Runtime gRPC uses:

- `CMS_GRPC_ADDR`
- `GRPC_TLS_ENABLED`
- `GRPC_DIAL_TIMEOUT`
- downstream `*_GRPC_ADDR`

## 9. Start commands

Protobuf does not start as a service.

Generate command:

```bash
# Suggested command based on project structure
buf generate
```

or:

```bash
# Suggested command based on project structure
protoc --go_out=. --go-grpc_out=. proto/ecommerce/cms/v1/cms.proto
```

## 10. Verify running commands

```bash
# Suggested command based on project structure
buf lint
buf breaking --against '.git#branch=main'
```

If Buf config is missing, these commands will not work until `buf.yaml` exists.

## 11. Common errors and fixes

| Error | Reason | Fix |
|---|---|---|
| `proto file not found` | `proto/` directory missing | Add proto files |
| `protoc-gen-go not found` | Go generator not installed | Install generator |
| Breaking change detected | Field/method changed incompatibly | Add new field/version instead of modifying existing contract |
| Generated code missing | Generator not run | Run generation and update imports |

## 12. Security notes

- Do not place secrets in proto messages unless absolutely required.
- Auth context should be explicit and minimal.
- Use stable field numbers; never reuse deleted field numbers.
- Keep seller-scoped APIs clear so ownership validation is not ambiguous.

## 13. Final checklist

- [x] Protobuf requirement identified from docs.
- [x] CMS methods identified from API/design docs.
- [ ] `proto/` directory not found.
- [ ] `cms.proto` not found.
- [ ] Generated protobuf code not found.
- [ ] Buf/protoc config not found.
