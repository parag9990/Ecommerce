# Session Analytics Dashboard Local Runbook

## 1. Purpose

Admin analytics UI for live sessions, journeys, funnels, heatmaps, cohorts, reports, and privacy workflows.

## 2. Location

`frontend/session-analytics-dashboard`

## 3. Tech Stack

React, TypeScript, Vite, React Router, React Query, Recharts, pnpm workspace.

## 4. Required Dependencies

API Gateway and session management service.

## 5. Environment Variables

Use `frontend/session-analytics-dashboard/.env.example`.

Key vars: `VITE_API_BASE_URL`, `VITE_API_PROXY_TARGET`, `VITE_REQUEST_TIMEOUT_MS`.

## 6. Install Dependencies

```powershell
cd frontend
corepack pnpm install --frozen-lockfile
```

## 7. Database/Migration/Seed Setup

No database is owned by this frontend. The dashboard reads analytics data through API Gateway from `session-service`.

Required data stores:

- MongoDB `session_db`: `sessions`, `session_events`, `heatmap_points`, `analytics_aggregates`, retention/report/privacy collections.
- Redis DB 5: active session indexes, live user/session counters, and recent event counters.

Run `session-service` Mongo migrations before testing analytics pages. There is no standalone analytics seed script; generate data by using the user app or ingesting session events into `session-service`. Overview/live sessions require active Redis data, while journeys, funnels, heatmaps, cohorts, reports, and privacy pages read Mongo-backed session analytics data.

## 8. Run Command

Recommended:

```powershell
docker compose up -d --build session-analytics-dashboard
```

Manual:

```powershell
cd frontend
corepack pnpm dev:analytics -- --port 5175
```

The Vite config defaults to port `5174`, which can conflict with seller dashboard.

## 9. Health Check

- Docker app URL: `http://localhost:3002`
- Docker image health: `http://localhost:3002/healthz`

## 10. Logs

```powershell
docker compose logs -f session-analytics-dashboard
```

Manual Vite logs appear in the terminal.

## 11. Common Issues

- Dashboard is empty until session events exist.
- Overview and live sessions are empty until Redis has recent active session/event data.
- Funnels use `product_view`, `add_to_cart`, `checkout_step`, and `payment_result` event types.
- Heatmaps use the `heatmap_type` query value (`click` or `scroll`).
- Manual Vite port conflicts with seller dashboard.
- `VITE_API_PROXY_TARGET` must point to the API Gateway.
- Admin authorization is required for protected analytics endpoints.

## 12. Quick Verification

Open `http://localhost:3002` for Docker or the printed Vite URL for manual mode.
