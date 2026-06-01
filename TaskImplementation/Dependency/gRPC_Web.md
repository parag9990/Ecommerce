# gRPC-Web Dependency - User App Frontend

## 1. What Is This Dependency?

gRPC-Web browser-friendly RPC protocol hai. Normal browser direct HTTP/2 gRPC call nahi kar sakta, isliye Envoy ya gateway bridge gRPC-Web request ko backend gRPC me convert karta hai.

Simple Hinglish: Frontend typed RPC call karta hai, bridge usko backend gRPC service tak pahuchata hai.

## 2. Why This Service Uses It

User App selected typed calls ke liye gRPC-Web use karta hai:

- Recommendations.
- Search autocomplete.
- Session event ingestion.

REST primary API path hai. gRPC-Web sirf selected calls ke liye use ho raha hai.

## 3. Required Or Optional

Required only for features using `.grpc.ts` wrappers. Agar gRPC-Web bridge down hai, REST pages still partially work, but recommendations/autocomplete/session event tracking fail ho sakte hain.

## 4. Where It Is Used In Project

| Path | Usage |
|------|-------|
| `frontend/user-app/src/lib/grpc-client.ts` | Connect-Web transport and service clients. |
| `frontend/user-app/src/lib/grpc-interceptors.ts` | Request id, source, session headers, auth header. |
| `frontend/user-app/src/lib/grpc-errors.ts` | Connect error to `ApiError` mapping. |
| `frontend/user-app/src/features/recommendation/api/recommendation.grpc.ts` | Recommendations RPC wrapper. |
| `frontend/user-app/src/features/search/api/autocomplete.grpc.ts` | Autocomplete RPC wrapper. |
| `frontend/user-app/src/features/analytics/api/session-events.grpc.ts` | Session event RPC wrapper. |
| `frontend/packages/proto-client/src/grpc-web.ts` | Shared service exports. |

## 5. Installation Steps

Dependencies already present in `frontend/user-app/package.json`:

```bash
cd frontend
pnpm --filter user-app add @connectrpc/connect @connectrpc/connect-web
pnpm --filter user-app add @ecommerce/proto-client@workspace:*
```

Suggested command based on project structure. Current package already has these dependencies.

## 6. Docker Setup, If Possible

gRPC-Web needs an Envoy or gateway bridge. Actual Envoy config/Docker Compose was not clearly found.

Suggested compose shape:

```yaml
services:
  grpc-web:
    image: envoyproxy/envoy:v1.31-latest
    ports:
      - "8082:8082"
    volumes:
      - ./infra/envoy/envoy.yaml:/etc/envoy/envoy.yaml:ro
```

This is suggested only. Actual `infra/envoy/envoy.yaml` was not clearly found in project files.

## 7. Local Setup Without Docker

Frontend env:

```env
VITE_GRPC_WEB_BASE_URL=http://localhost:8082
VITE_GRPC_WEB_TIMEOUT_MS=5000
```

Bridge start command is not clearly found in project files.

Suggested command based on project structure:

```bash
envoy -c infra/envoy/envoy.yaml
```

## 8. Required Environment Variables

| Variable | Required? | Example |
|----------|-----------|---------|
| `VITE_GRPC_WEB_BASE_URL` | Required for gRPC-Web | `http://localhost:8082` |
| `VITE_GRPC_WEB_TIMEOUT_MS` | Optional | `5000` |

Headers added by frontend interceptor:

| Header | Purpose |
|--------|---------|
| `x-request-id` | Trace one request. |
| `x-request-source` | Always `user-app`. |
| `x-anonymous-id` | Anonymous visitor id from session store. |
| `x-session-id` | Session id from session store. |
| `authorization` | Bearer access token if present. |

## 9. Start Commands

Frontend:

```bash
cd frontend
pnpm --filter user-app dev
```

Bridge:

```bash
envoy -c infra/envoy/envoy.yaml
```

Suggested command based on project structure. Exact bridge command/config is not clearly found.

## 10. Verify Running Commands

Check bridge responds:

```bash
curl http://localhost:8082
```

Run gRPC-related tests:

```bash
cd frontend
pnpm --filter user-app test
```

Relevant tests found:

- `frontend/user-app/src/features/recommendation/api/recommendation.grpc.test.ts`
- `frontend/user-app/src/features/search/api/autocomplete.grpc.test.ts`
- `frontend/user-app/src/features/analytics/api/session-events.grpc.test.ts`
- `frontend/user-app/src/lib/grpc-interceptors.test.ts`
- `frontend/user-app/src/lib/grpc-errors.test.ts`

## 11. Common Errors And Fixes

| Error | Reason | Fix |
|-------|--------|-----|
| Browser CORS error | Bridge not allowing Vite origin | Configure CORS for `http://localhost:5173` and credentials. |
| 501 `GRPC_UNIMPLEMENTED` | Backend method not registered | Register service/method on backend. |
| 503 `GRPC_UNAVAILABLE` | Bridge/backend down | Start bridge and backend service. |
| Deadline exceeded | Timeout too low or backend slow | Increase `VITE_GRPC_WEB_TIMEOUT_MS` or fix backend latency. |
| Missing auth/session context | Interceptor not running | Use `createUserGrpcTransport()` from `grpc-client.ts`. |

## 12. Security Notes

- gRPC-Web bridge must enforce CORS and not allow arbitrary origins in production.
- Do not put secrets in gRPC metadata from frontend.
- Auth must be enforced on backend. Frontend headers are user-controllable.
- Request ids are good for debugging but should not contain PII.
- Session analytics events should avoid raw sensitive data.

## 13. Final Checklist

| Item | Status |
|------|--------|
| Connect-Web client configured | Completed |
| gRPC interceptors configured | Completed |
| gRPC error mapping configured | Completed |
| Selected RPC wrappers found | Completed |
| Bridge config found | Missing |
| Backend gRPC server registration found | Missing |
