# State Management Dependency - User App Frontend

## 1. What Is This Dependency?

State management dependencies frontend data ko predictable way me manage karte hain.

This service uses:

- TanStack React Query for server data cache.
- Zustand for small client-side UI/auth/session state.

Simple Hinglish: Backend se aaya data React Query me, browser ka chhota state Zustand me.

## 2. Why This Service Uses It

React Query:

- Product, cart, wishlist, profile, orders data fetch/cache.
- Loading/error/retry states.
- Cache invalidation after mutations.

Zustand:

- Mobile nav/account menu state.
- Toast state.
- Auth user/session snapshot.
- Anonymous/session ids for analytics/gRPC headers.

## 3. Required Or Optional

Required. Current source imports both React Query and Zustand.

## 4. Where It Is Used In Project

| Path | Usage |
|------|-------|
| `frontend/user-app/src/lib/query-client.ts` | Query client default retry/cache settings. |
| `frontend/user-app/src/providers/app-providers.tsx` | `QueryClientProvider` and devtools. |
| `frontend/user-app/src/features/*/hooks/*.ts` | Query/mutation hooks. |
| `frontend/user-app/src/stores/ui-store.ts` | UI store. |
| `frontend/user-app/src/stores/auth-store.ts` | Auth user/session store. |
| `frontend/user-app/src/stores/session-store.ts` | Anonymous/session analytics state. |
| `frontend/user-app/src/lib/grpc-interceptors.ts` | Reads session store for gRPC metadata. |

## 5. Installation Steps

Current dependencies already exist.

```bash
cd frontend
pnpm --filter user-app add @tanstack/react-query zustand
pnpm --filter user-app add -D @tanstack/react-query-devtools
```

Suggested command based on project structure.

## 6. Docker Setup, If Possible

No separate Docker dependency. State libraries are bundled into frontend build.

Build verification:

```bash
cd frontend
pnpm --filter user-app build
```

## 7. Local Setup Without Docker

```bash
cd frontend
pnpm install
pnpm --filter user-app dev
```

React Query Devtools load only in dev mode through `import.meta.env.DEV`.

## 8. Required Environment Variables

State management has no direct env var.

Indirect env values matter because queries call APIs:

| Variable | Purpose |
|----------|---------|
| `VITE_API_BASE_URL` | REST query/mutation base URL. |
| `VITE_GRPC_WEB_BASE_URL` | gRPC query base URL. |

## 9. Start Commands

```bash
cd frontend
pnpm --filter user-app dev
```

## 10. Verify Running Commands

```bash
cd frontend
pnpm --filter user-app test
pnpm --filter user-app typecheck
```

Relevant tests found:

- `frontend/user-app/src/stores/ui-store.test.ts`
- feature hook/API tests use React Query test wrapper.

## 11. Common Errors And Fixes

| Error | Reason | Fix |
|-------|--------|-----|
| Query retries protected 401 endpoint | Retry rules wrong | Keep `query-client.ts` no-retry for 401/403/404. |
| Stale UI after mutation | Missing invalidation | Invalidate matching query keys after mutation. |
| Zustand state persists bad old shape | Persisted localStorage mismatch | Add migration/version or clear local app storage. |
| Devtools in prod | Wrong env/build mode | Keep devtools behind `import.meta.env.DEV`. |

## 12. Security Notes

- Do not store long-lived secrets in Zustand localStorage.
- Access token is sessionStorage-based through `auth-session.ts`; XSS protection remains important.
- Avoid putting PII in toast messages or analytics session metadata.
- Server remains source of truth; frontend state is not authorization.

## 13. Final Checklist

| Item | Status |
|------|--------|
| React Query dependency found | Completed |
| Query client found | Completed |
| Zustand stores found | Completed |
| Devtools dev-only behavior found | Completed |
| Security notes added | Completed |
