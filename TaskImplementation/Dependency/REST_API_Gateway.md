# REST API Gateway Dependency - User App Frontend

## 1. What Is This Dependency?

REST API Gateway public backend entry point hai. Browser REST JSON calls Gateway ko bhejta hai, phir Gateway internal services ko route karta hai.

Simple Hinglish: Frontend direct Auth/User/Product/Cart DB se baat nahi karta. Frontend Gateway ko call karta hai, Gateway backend services se baat karta hai.

## 2. Why This Service Uses It

User App Frontend REST APIs use karta hai for:

- Signup, login, logout, OTP, password reset.
- Product list/detail/category/search.
- Cart add/update/remove/coupon preview.
- Checkout, order detail/list/cancel.
- Profile, addresses, wishlist, notification preferences.
- Payment retry.

## 3. Required Or Optional

Required. Without API Gateway, most frontend pages load but data actions fail.

## 4. Where It Is Used In Project

| Path | Usage |
|------|-------|
| `frontend/user-app/src/lib/http.ts` | Shared REST client, auth headers, envelope parsing, error mapping. |
| `frontend/user-app/src/features/auth/api/auth.api.ts` | Auth endpoints. |
| `frontend/user-app/src/features/product/api/product.api.ts` | Product/category/search REST endpoints. |
| `frontend/user-app/src/features/cart/api/cart.api.ts` | Cart endpoints. |
| `frontend/user-app/src/features/checkout/api/*.api.ts` | Checkout, payment retry, addresses. |
| `frontend/user-app/src/features/profile/api/*.api.ts` | Profile and preferences. |
| `frontend/user-app/src/features/wishlist/api/wishlist.api.ts` | Wishlist endpoints. |
| `api/master-api.json` | Contract for REST route to backend service/gRPC mapping. |
| `backend/services/api-gateway/.env` | Gateway env keys and port detected. |

## 5. Installation Steps

Frontend side needs no extra package for REST because browser `fetch` is used.

Backend Gateway source/start command was not clearly found in project files.

Suggested command based on project structure, after backend source exists:

```bash
cd backend
go run ./services/api-gateway/cmd/server
```

## 6. Docker Setup, If Possible

Docker Compose for gateway was not clearly found.

Suggested compose shape:

```yaml
services:
  api-gateway:
    build:
      context: .
      dockerfile: backend/services/api-gateway/deploy/Dockerfile
    env_file:
      - backend/services/api-gateway/.env
    ports:
      - "8080:8080"
```

This is a suggested shape only. Actual Dockerfile/compose file is missing.

## 7. Local Setup Without Docker

Expected local URL:

```text
http://localhost:8080
```

Frontend env:

```env
VITE_API_BASE_URL=http://localhost:8080
```

Backend source was not clearly found, so exact backend local start is unclear.

## 8. Required Environment Variables

Frontend:

| Variable | Value |
|----------|-------|
| `VITE_API_BASE_URL` | `http://localhost:8080` |

Gateway env keys detected:

| Gateway Variable | Example Detected |
|------------------|------------------|
| `SERVICE_NAME` | `api-gateway` |
| `APP_ENV` | `local` |
| `HTTP_ADDR` | `:8080` |
| `API_BASE_PATH` | `/api/v1` |
| `AUTH_GRPC_ADDR` | `localhost:50051` |
| `USER_GRPC_ADDR` | `localhost:50052` |
| `CART_GRPC_ADDR` | `localhost:50054` |
| `SEARCH_GRPC_ADDR` | `localhost:50058` |
| `REDIS_ADDR` | `localhost:6379` |

## 9. Start Commands

Frontend command:

```bash
cd frontend
pnpm --filter user-app dev
```

Backend command:

```bash
cd backend
go run ./services/api-gateway/cmd/server
```

Suggested command based on project structure. Exact command is not clearly found in project files.

## 10. Verify Running Commands

Suggested checks:

```bash
curl http://localhost:8080/api/v1/categories
curl http://localhost:8080/api/v1/products
```

If protected endpoint is checked, auth token/cookie is required:

```bash
curl http://localhost:8080/api/v1/cart
```

## 11. Common Errors And Fixes

| Error | Reason | Fix |
|-------|--------|-----|
| `NETWORK_ERROR` in frontend | Gateway not running | Start gateway or update `VITE_API_BASE_URL`. |
| 401/403 | Missing/expired auth | Login again; backend must set/accept token/cookie. |
| CORS error | Gateway CORS not configured | Allow Vite dev origin and credentials. |
| Empty response | Backend response not matching envelope | Return `data` or valid JSON envelope. |
| Route 404 | Gateway route not registered | Implement route from `api/master-api.json`. |

## 12. Security Notes

- Gateway must validate JWT/cookies and not trust frontend route guards.
- Use `credentials: include` carefully with proper CORS and SameSite cookie settings.
- Keep `Authorization` token handling secure; frontend tokens are XSS-sensitive.
- Gateway should rate-limit auth/cart/checkout APIs.
- Gateway should normalize errors without leaking internal service details.

## 13. Final Checklist

| Item | Status |
|------|--------|
| REST base URL documented | Completed |
| Feature API wrapper paths documented | Completed |
| Gateway env keys detected | Completed |
| Backend source gap documented | Completed |
| Common errors documented | Completed |
