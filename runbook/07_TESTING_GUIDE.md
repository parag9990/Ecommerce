# Testing Guide

Run [checklists/SERVICE_HEALTH_CHECKLIST.md](checklists/SERVICE_HEALTH_CHECKLIST.md) before functional testing.

## Automated Tests

Backend command from the root `Makefile`:

```powershell
make test-go
```

Frontend:

```powershell
make test-frontend
```

All tests:

```powershell
make test
```

If `make` is unavailable:

```powershell
Get-ChildItem backend/services,backend/shared,backend/proto-gen -Recurse -Filter go.mod -File | ForEach-Object {
  Push-Location $_.DirectoryName
  go test ./...
  Pop-Location
}

cd frontend
corepack pnpm install --frozen-lockfile
corepack pnpm -r --if-present test
```

Do not use `go test ./...` from `backend/`; the workspace root has `go.work` but no root `go.mod`, and that pattern fails for this repo layout.

## Normal User Flow

1. Open `http://localhost:3000`.
2. Register or log in.
3. Browse products.
4. Search/autocomplete products.
5. Add an item to cart.
6. Add/remove wishlist item.
7. Place an order.
8. Test payment flow.
9. Check notification email in Mailpit at `http://localhost:8025`.

Warnings:

- Seed users and products: Not found in codebase - please confirm.
- Auth signup route mismatch needs confirmation.
- Payment providers are disabled by default.
- Internal CMS/payment/event token and endpoint defaults are blank in examples.

## Seller/CMS Flow

1. Open `http://localhost:3001`.
2. Log in as a seller.
3. Create or update a product.
4. Publish or submit product workflow.
5. Manage orders.
6. Create coupons/campaigns.
7. Check seller analytics.

Seller credentials: Not found in codebase - please confirm.

## Superadmin Flow

1. Open `http://localhost:3003`.
2. Log in as superadmin.
3. Manage users and sellers.
4. Review orders, payments, refunds, and reconciliation views.
5. Manage platform settings.
6. Review audit logs and session analytics.

Superadmin credentials: Not found in codebase - please confirm.

## Session Analytics Flow

1. Open `http://localhost:3002`.
2. Confirm live sessions load.
3. Trigger user app events.
4. Check journeys, funnels, heatmaps, cohorts, reports, and privacy pages.
