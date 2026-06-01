# Environment Dependency - User App Frontend

## 1. What Is This Dependency?

Environment dependency ka matlab hai runtime configuration jo app ko batata hai backend URLs, app mode, timeout, and payment provider options kya hain.

Simple Hinglish: Env vars app ke settings hain. Code same rehta hai, values local/staging/production me change hoti hain.

## 2. Why This Service Uses It

`User App Frontend` ko runtime pe ye know karna hota hai:

- REST API Gateway kaha running hai.
- gRPC-Web bridge kaha running hai.
- gRPC timeout kitna hoga.
- App local/development/staging/production me hai.
- Checkout screen me kaunse payment providers show karne hain.

## 3. Required Or Optional

| Variable | Required? | Default in code? |
|----------|-----------|------------------|
| `VITE_API_BASE_URL` | Required for real APIs | Yes, `http://localhost:8080` |
| `VITE_GRPC_WEB_BASE_URL` | Required for gRPC-Web | Yes, `http://localhost:8082` |
| `VITE_GRPC_WEB_TIMEOUT_MS` | Optional | Yes, `5000` |
| `VITE_APP_ENV` | Optional but validated | Yes, `local` |
| `VITE_PAYMENT_PROVIDERS` | Optional | Yes, `stripe,razorpay` |

## 4. Where It Is Used In Project

| Path | Usage |
|------|-------|
| `frontend/user-app/.env.example` | Example values. |
| `frontend/user-app/src/lib/env.ts` | Reads, validates, and defaults env vars. |
| `frontend/user-app/src/lib/http.ts` | Uses `env.apiBaseUrl`. |
| `frontend/user-app/src/lib/grpc-client.ts` | Uses `env.grpcWebBaseUrl` and timeout. |
| `frontend/user-app/src/features/checkout/pages/checkout-page.tsx` | Uses payment provider list. |
| `frontend/user-app/src/features/checkout/components/payment-step.tsx` | Shows payment providers. |

## 5. Installation Steps

No package install needed for env files.

Suggested command based on project structure:

```bash
cd frontend
cp user-app/.env.example user-app/.env.local
```

Then edit `frontend/user-app/.env.local` for your local backend URLs.

## 6. Docker Setup, If Possible

For Vite static builds, `VITE_` env vars are normally baked at build time. Docker build should pass build args or use an env file during build.

Suggested command based on project structure:

```bash
docker build --build-arg VITE_API_BASE_URL=http://localhost:8080 -t user-app-frontend .
```

Actual Dockerfile was not clearly found in project files.

## 7. Local Setup Without Docker

```bash
cd frontend
cp user-app/.env.example user-app/.env.local
pnpm --filter user-app dev
```

Example local `.env.local`:

```env
VITE_API_BASE_URL=http://localhost:8080
VITE_GRPC_WEB_BASE_URL=http://localhost:8082
VITE_GRPC_WEB_TIMEOUT_MS=5000
VITE_APP_ENV=local
VITE_PAYMENT_PROVIDERS=stripe,razorpay
```

## 8. Required Environment Variables

| Variable | Example | Notes |
|----------|---------|-------|
| `VITE_API_BASE_URL` | `http://localhost:8080` | No trailing slash required. Code normalizes paths. |
| `VITE_GRPC_WEB_BASE_URL` | `http://localhost:8082` | gRPC-Web bridge must support browser CORS. |
| `VITE_GRPC_WEB_TIMEOUT_MS` | `5000` | Must be positive integer. |
| `VITE_APP_ENV` | `local` | Allowed: `local`, `development`, `staging`, `production`. |
| `VITE_PAYMENT_PROVIDERS` | `stripe,razorpay` | Comma-separated list. |

## 9. Start Commands

```bash
cd frontend
pnpm --filter user-app dev
```

With a custom env file, Vite automatically reads `.env.local`.

## 10. Verify Running Commands

Verify env typing and validation through build:

```bash
cd frontend
pnpm --filter user-app typecheck
pnpm --filter user-app build
```

Verify API base URL manually:

```bash
curl http://localhost:8080/api/v1/categories
```

Suggested command based on project structure. It works only if API Gateway is running.

## 11. Common Errors And Fixes

| Error | Reason | Fix |
|-------|--------|-----|
| `Invalid VITE_APP_ENV` | Value not in allowed list | Use `local`, `development`, `staging`, or `production`. |
| API calls go to localhost in staging | Env not set during build | Set Vite env at build time. |
| gRPC timeout ignored | Non-positive timeout value | Use positive integer milliseconds. |
| Payment provider missing in UI | Empty provider list | Set `VITE_PAYMENT_PROVIDERS=stripe,razorpay`. |

## 12. Security Notes

- `VITE_` env vars public hote hain. Secret API keys, JWT secrets, DB passwords yahan mat rakho.
- Payment provider secret keys backend me rahenge, frontend me sirf public/client token allowed hai.
- Avoid committing `.env.local`.
- Keep `.env.example` secret-free and beginner-friendly.

## 13. Final Checklist

| Item | Status |
|------|--------|
| `.env.example` found | Completed |
| Env loader found | Completed |
| Allowed `VITE_APP_ENV` values documented | Completed |
| Security notes added | Completed |
| Backend env values not exposed as secrets | Completed |
