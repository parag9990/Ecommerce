# Frontend Dependency - Seller Dashboard (CMS)

## 1. What is this dependency?

Frontend dependency ka matlab Seller Dashboard ka browser app hai. Ye React + TypeScript + Vite app hai jo seller ko dashboard UI provide karta hai.

## 2. Why this service uses it

Seller Dashboard (CMS) ka main user-facing surface frontend hi hai. Seller products, orders, offers, analytics, team, aur audit modules isi app se use karta hai.

## 3. Required or optional

Required. Frontend ke bina Seller Dashboard UI available nahi hoga.

## 4. Where it is used in project

| Path | Use |
|---|---|
| `frontend/seller-dashboard/package.json` | React/Vite/TanStack Query/Zustand dependencies |
| `frontend/seller-dashboard/vite.config.ts` | Dev server `5174`, preview `4174` |
| `frontend/seller-dashboard/src/routes/seller-routes.tsx` | `/seller` routes |
| `frontend/seller-dashboard/src/lib/http.ts` | API Gateway HTTP client |
| `frontend/seller-dashboard/src/features/` | Products, orders, offers, analytics, team, audit modules |

## 5. Installation steps

```bash
cd frontend
pnpm install
```

Project package scripts:

```bash
pnpm --filter seller-dashboard dev
pnpm --filter seller-dashboard build
pnpm --filter seller-dashboard test
pnpm --filter seller-dashboard typecheck
```

## 6. Docker setup, if possible

Not clearly found in project files.

No `frontend/seller-dashboard/Dockerfile` or compose service was found. Suggested Docker setup can be added later after production build requirements are finalized.

## 7. Local setup without Docker

```bash
cd frontend
pnpm install
pnpm --filter seller-dashboard dev
```

Open:

```text
http://localhost:5174
```

## 8. Required environment variables

| Variable | Required | Default/Notes |
|---|---|---|
| `VITE_API_BASE_URL` | Optional | Defaults to `http://localhost:8080` |
| `VITE_API_TIMEOUT_MS` | Optional | Defaults to `15000` |
| `VITE_LOGIN_URL` | Optional | Used by login redirect page |

`.env.example` for frontend was not clearly found in project files.

## 9. Start commands

```bash
cd frontend
pnpm --filter seller-dashboard dev
```

Build command:

```bash
cd frontend
pnpm --filter seller-dashboard build
```

## 10. Verify running commands

```bash
cd frontend
pnpm --filter seller-dashboard typecheck
pnpm --filter seller-dashboard test
```

Manual verify:

```bash
curl http://localhost:5174
```

## 11. Common errors and fixes

| Error | Reason | Fix |
|---|---|---|
| `pnpm: command not found` | pnpm installed nahi hai | Enable corepack or install pnpm |
| API calls fail with network error | API Gateway `8080` running nahi hai | Start gateway or set `VITE_API_BASE_URL` |
| 401 redirect to login | Seller session missing/expired | Login/session backend verify karo |
| 403/permission denied | Seller inactive or role missing | Backend RBAC and seller status check karo |
| Blank data in analytics | Backend optional fields missing | UI unavailable state expected hai; fake data mat add karo |

## 12. Security notes

- Frontend permission checks sirf UX ke liye hain. Real security API Gateway + backend services enforce karega.
- Cookies are sent with `credentials: include`; CORS and same-site cookie setup backend side correct hona chahiye.
- `x-request-id` request tracing ke liye useful hai; logs me secrets include mat karo.
- Never put private tokens in `VITE_*` vars because Vite env browser bundle me expose hota hai.

## 13. Final checklist

- [x] React/Vite app found.
- [x] Seller routes found.
- [x] API client found.
- [x] Port `5174` confirmed.
- [x] Frontend env vars identified.
- [ ] Frontend `.env.example` not found.
- [ ] Frontend Dockerfile not found.

