# Step 11 - Frontend Implementation

## Frontend Stack

- React.js + TypeScript.
- Tailwind CSS.
- React Query for server state.
- Zustand for lightweight client UI/session state.
- Redux Toolkit only if complex cross-module workflows grow later.
- gRPC-Web for typed selected service calls through Envoy/gateway.
- REST for most public browser APIs through API Gateway.

## State Management Rule

Use the right state tool:

| State Type | Tool | Example |
|---|---|---|
| Server data | React Query | product list, cart, orders |
| Local UI state | Zustand | drawer open, filters, selected seller |
| Form state | React Hook Form | product editor, login |
| URL state | Router/search params | search query, category filters |
| Generated clients | gRPC-Web | typed analytics/admin calls |

## Component Structure

```text
src/
  features/
    feature-name/
      pages/
      components/
      api/
      hooks/
      types.ts
  components/
    ui/
    layout/
  lib/
    http.ts
    query-client.ts
    grpc-client.ts
  stores/
  routes/
```

Rules:

- Feature code feature folder me rahe.
- Shared UI components `packages/ui` me rahe.
- API calls `api` folder me typed functions ke saath.
- React Query hooks feature ke paas.
- Business calculations duplicate na karo, backend source of truth.

## API Layer

HTTP client responsibilities:

- Base URL from env.
- Attach access token.
- Refresh token flow.
- Request id header.
- Error normalization.
- Timeout handling.

Example response envelope:

```ts
type ApiResponse<T> = {
  data: T;
  request_id: string;
  error?: {
    code: string;
    message: string;
    details?: unknown;
  };
};
```

## gRPC-Web Integration

Recommended:

- Use `buf` to generate TypeScript clients.
- Browser talks to Envoy/gRPC-Web proxy.
- Envoy forwards to API Gateway or directly to allowed gRPC services.

Use cases:

- Analytics dashboard live metrics.
- Admin operational APIs.
- Internal typed calls where REST shape is not ideal.

## Protected Routes

Route guards:

- `RequireAuth`: user must be logged in.
- `RequireRole`: user must have role.
- `RequireSeller`: seller profile must exist and be active.
- `RequireAdmin`: admin/superadmin role.

Behavior:

- Unauthenticated: redirect to login.
- Unauthorized: show permission denied page.
- Token expired: refresh silently, then retry.
- Refresh failed: logout and redirect.

## User App Modules

### Auth

Pages:

- login
- signup
- OTP verify
- forgot password
- reset password

### Product

Pages:

- home
- category listing
- search results
- product detail

Features:

- Filters and facets.
- Sort.
- Pagination/infinite loading.
- Product recommendations.
- Recently viewed.

### Cart

Features:

- Add item.
- Quantity update.
- Remove item.
- Coupon preview.
- Cart total.
- Stock warning.

### Checkout

Steps:

1. Address.
2. Order review.
3. Payment.
4. Success/failure.

Rules:

- Amount displayed from server response.
- Checkout call uses idempotency key.
- Payment status confirmed by backend.

### Profile

Pages:

- profile details
- address book
- orders
- wishlist
- notification preferences

## Seller Dashboard Modules

- Overview metrics.
- Product manager.
- Variant editor.
- Image uploader.
- Order manager.
- Coupons/campaigns.
- Revenue analytics.
- Team permissions.
- Audit activity.

Design direction:

- Quiet, dense, operational UI.
- Tables with filters and bulk actions.
- Clear status badges.
- No marketing-style hero sections.

## Session Analytics Dashboard Modules

- Live sessions.
- Journey explorer.
- Funnel analysis.
- Heatmap.
- Device/browser reports.
- Cohorts.
- Privacy controls.

Design direction:

- Data-heavy dashboard.
- Date range and segment filters visible.
- Charts readable and exportable.

## Superadmin Panel Modules

- User management.
- Seller management.
- Order operations.
- Payment operations.
- Session oversight.
- Search/settings.
- Audit logs.

Security:

- Role-based menu.
- MFA recommended.
- Short session TTL.
- Audit note for risky actions.

## Testing

Frontend test layers:

- Unit tests for pure helpers.
- Component tests for forms and UI states.
- API mock tests with MSW.
- E2E tests for login, product search, cart, checkout, seller product create, admin refund review.

## Performance

- Route-level code splitting.
- Image lazy loading.
- CDN for static assets.
- React Query caching.
- Debounced search autocomplete.
- Avoid rendering huge tables without virtualization.

