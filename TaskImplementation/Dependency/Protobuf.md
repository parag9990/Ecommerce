# Protobuf Dependency - User App Frontend

## 1. What Is This Dependency?

Protobuf structured contract format hai. Isse backend and frontend typed request/response messages share kar sakte hain.

Simple Hinglish: Proto ek contract hai. Generated TypeScript code frontend ko exact message and service types deta hai.

## 2. Why This Service Uses It

`User App Frontend` gRPC-Web selected calls ke liye generated TS protobuf client use karta hai:

- Recommendation service.
- Search autocomplete service.
- Session event ingestion service.

## 3. Required Or Optional

Required for gRPC-Web features. REST-only screens ke liye optional hai.

## 4. Where It Is Used In Project

| Path | Usage |
|------|-------|
| `frontend/packages/proto-client/package.json` | Shared generated client package. |
| `frontend/packages/proto-client/src/index.ts` | Re-exports service types. |
| `frontend/packages/proto-client/src/grpc-web.ts` | Re-exports service descriptors/types. |
| `frontend/packages/proto-client/src/gen/ecommerce/recommendation/v1/recommendation_pb.ts` | Recommendation generated code. |
| `frontend/packages/proto-client/src/gen/ecommerce/search/v1/search_pb.ts` | Search generated code. |
| `frontend/packages/proto-client/src/gen/ecommerce/session/v1/session_pb.ts` | Session generated code. |
| `frontend/user-app/src/lib/grpc-client.ts` | Imports service descriptors. |

## 5. Installation Steps

Current package dependencies:

```bash
cd frontend
pnpm --filter @ecommerce/proto-client add @bufbuild/protobuf @connectrpc/connect
pnpm --filter user-app add @ecommerce/proto-client@workspace:*
```

Suggested command based on project structure. Current package already has these dependencies.

Original proto generation setup is incomplete:

| Item | Status |
|------|--------|
| Generated TS files | Present |
| Original `.proto` files | Not clearly found in project files. |
| `buf.yaml` | Not clearly found in project files. |
| `buf.gen.yaml` | Not clearly found in project files. |

## 6. Docker Setup, If Possible

Proto generation usually does not need Docker. If team standardizes generation in CI, use Buf CLI or a containerized Buf image.

Suggested command based on project structure:

```bash
buf generate
```

But Buf config and original `.proto` files were not clearly found in project files.

## 7. Local Setup Without Docker

Current generated files can be consumed directly after install:

```bash
cd frontend
pnpm install
pnpm --filter user-app typecheck
```

If proto source files are added later:

```bash
buf generate
pnpm --filter @ecommerce/proto-client typecheck
```

Suggested command based on expected proto workflow.

## 8. Required Environment Variables

Protobuf generation does not need frontend env vars.

Runtime gRPC-Web needs:

| Variable | Purpose |
|----------|---------|
| `VITE_GRPC_WEB_BASE_URL` | gRPC-Web bridge URL. |
| `VITE_GRPC_WEB_TIMEOUT_MS` | RPC timeout. |

## 9. Start Commands

No separate start command for protobuf package.

Useful commands:

```bash
cd frontend
pnpm --filter @ecommerce/proto-client typecheck
pnpm --filter user-app typecheck
```

## 10. Verify Running Commands

Verify generated package is resolvable:

```bash
cd frontend
pnpm --filter user-app typecheck
```

Verify generated services are present:

```bash
rg "export const .*Service" frontend/packages/proto-client/src/gen/ecommerce
```

## 11. Common Errors And Fixes

| Error | Reason | Fix |
|-------|--------|-----|
| `Cannot find module @ecommerce/proto-client` | Workspace install missing | Run `pnpm install` from `frontend/`. |
| Missing generated type | Proto not generated for service | Add `.proto` and run generation. |
| Frontend and backend contract mismatch | Generated files stale | Regenerate TS and Go code from same proto source. |
| `buf generate` fails | Buf config missing | Add `buf.yaml` and `buf.gen.yaml`. |

## 12. Security Notes

- Proto contracts should not expose internal-only sensitive fields to browser clients.
- Version proto packages, for example `ecommerce.search.v1`.
- Do not break existing generated APIs without migration.
- Treat generated code as source artifact: review diffs when contracts change.

## 13. Final Checklist

| Item | Status |
|------|--------|
| Shared proto-client package found | Completed |
| Generated TS protobuf files found | Completed |
| Original `.proto` files found | Missing |
| Buf config found | Missing |
| Runtime env documented | Completed |
