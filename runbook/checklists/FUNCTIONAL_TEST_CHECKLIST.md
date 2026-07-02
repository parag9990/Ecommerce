# Functional Test Checklist

Run this only after [SERVICE_HEALTH_CHECKLIST.md](SERVICE_HEALTH_CHECKLIST.md) passes and test credentials are confirmed.

## Normal User

- [ ] Register or log in. If signup fails, confirm the known auth signup route mismatch first.
- [ ] Browse products.
- [ ] Search products.
- [ ] Add product to cart.
- [ ] Add product to wishlist.
- [ ] Place order.
- [ ] Complete or simulate payment only after sandbox payment provider variables are configured.
- [ ] Verify notification in Mailpit.

## Seller/CMS

- [ ] Log in as seller.
- [ ] Create or update product.
- [ ] Publish or submit product for workflow.
- [ ] Manage orders.
- [ ] Create coupon or campaign.
- [ ] Check dashboard analytics.

## Superadmin

- [ ] Log in as superadmin.
- [ ] Manage users.
- [ ] Manage sellers.
- [ ] Review orders and payments.
- [ ] Review refunds/reconciliation.
- [ ] Manage platform settings.
- [ ] Review audit logs and dashboard.

## Notes

- Seed users and credentials: Not found in codebase - please confirm.
- Payment sandbox provider setup: Not found in codebase - please confirm.
- Auth signup route mismatch: catalog/frontend reference signup, but auth route registration was not confirmed.
- Internal CMS/payment/event token and endpoint defaults are blank in examples.
