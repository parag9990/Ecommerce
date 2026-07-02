# User App Frontend Troubleshooting

## Common Error

- Blank page or API requests fail.
- Login/signup request fails.
- gRPC-Web calls fail.

## Possible Cause

- Gateway is down.
- `VITE_API_BASE_URL` or `VITE_GRPC_WEB_BASE_URL` is wrong.
- CORS does not include the manual Vite origin.
- Backend signup route needs confirmation.

## Fix

- Start API Gateway and backend services.
- Use `http://localhost:8080` for REST and `http://localhost:8099` for gRPC-Web in Docker local mode.
- Add manual frontend origin to gateway `CORS_ALLOWED_ORIGINS`.

## Verification Command

```powershell
curl http://localhost:8080/health/live
```
