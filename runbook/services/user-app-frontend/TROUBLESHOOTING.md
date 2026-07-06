# User App Frontend Troubleshooting

## Common Error

- Blank page or API requests fail.
- Login/signup request fails.
- gRPC-Web calls fail.
- Popular products fail with a validation error.
- Wishlist cards show `Saved product`, `N/A`, or no image.
- Checkout fails with a gateway route bridge or upstream error.

## Possible Cause

- Gateway is down.
- `VITE_API_BASE_URL` or `VITE_GRPC_WEB_BASE_URL` is wrong.
- CORS does not include the manual Vite origin.
- Backend signup route needs confirmation.
- Gateway/API contract validation does not allow the frontend's search sort.
- Search results or wishlist payloads only contain product summaries and need Product Service hydration.
- Order Service, User Service address lookup, Cart Service, Product Service, or Payment Service is not ready.

## Fix

- Start API Gateway and backend services.
- Use `http://localhost:8080` for REST and `http://localhost:8099` for gRPC-Web in Docker local mode.
- Add manual frontend origin to gateway `CORS_ALLOWED_ORIGINS`.
- Restart API Gateway after route/contract changes.
- Confirm Product Service returns published products before debugging category visibility.
- Confirm Search Service and Typesense are healthy before debugging popular/search-only filters.
- Confirm the cart has items and the selected delivery address exists before checkout.

## Verification Command

```powershell
curl http://localhost:8080/health/live
curl "http://localhost:8080/api/v1/products?page=1&page_size=12&status=published"
curl "http://localhost:8080/api/v1/search?sort=popular&page=1&page_size=12"
```
