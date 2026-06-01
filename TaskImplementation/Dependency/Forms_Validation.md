# Forms And Validation Dependency - User App Frontend

## 1. What Is This Dependency?

Forms and validation stack user input ko manage and validate karta hai.

This service uses:

- `react-hook-form` for form state.
- `zod` for schemas.
- `@hookform/resolvers` to connect Zod with React Hook Form.

Simple Hinglish: User ka input pehle browser me validate hota hai, phir API Gateway ko send hota hai.

## 2. Why This Service Uses It

Used in:

- Auth forms.
- OTP/password reset forms.
- Checkout address/payment selection.
- Profile form.
- Address form.
- Notification preference form.
- Order cancel reason.
- Coupon form.

## 3. Required Or Optional

Required. Current feature pages/components import these libraries.

## 4. Where It Is Used In Project

| Path | Usage |
|------|-------|
| `frontend/user-app/src/features/auth/schemas.ts` | Auth validation. |
| `frontend/user-app/src/features/checkout/checkout-schema.ts` | Checkout validation. |
| `frontend/user-app/src/features/profile/profile-schema.ts` | Profile validation. |
| `frontend/user-app/src/features/addresses/address-schema.ts` | Address validation. |
| `frontend/user-app/src/features/cart/components/coupon-box.tsx` | Coupon form. |
| `frontend/user-app/src/features/orders/components/cancel-order-dialog.tsx` | Cancel order form. |
| `frontend/user-app/src/features/*/pages/*.tsx` | Form screens. |

## 5. Installation Steps

Current dependencies already exist.

```bash
cd frontend
pnpm --filter user-app add react-hook-form zod @hookform/resolvers
```

## 6. Docker Setup, If Possible

No separate Docker setup. Validation code is bundled into frontend build.

Verify in Docker build by running:

```bash
pnpm --filter user-app build
```

## 7. Local Setup Without Docker

```bash
cd frontend
pnpm install
pnpm --filter user-app dev
```

## 8. Required Environment Variables

No direct env variables.

Forms that submit to backend depend indirectly on:

| Variable | Purpose |
|----------|---------|
| `VITE_API_BASE_URL` | Submit forms through REST API Gateway. |

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

- `features/auth/schemas.test.ts`
- `features/checkout/checkout-schema.test.ts`
- `features/profile/profile-schema.test.ts`
- `features/addresses/address-schema.test.ts`

## 11. Common Errors And Fixes

| Error | Reason | Fix |
|-------|--------|-----|
| Form submits invalid data | Missing Zod schema/resolver | Add schema and `zodResolver`. |
| Type mismatch between schema and API type | Schema/type duplicated incorrectly | Infer form type from schema when possible. |
| Backend rejects frontend-valid input | Contract mismatch | Align with `api/master-api.json` and backend validation. |
| Error messages not shown | Form field not wired | Pass field error to UI component. |

## 12. Security Notes

- Frontend validation UX ke liye hai. Backend validation mandatory hai.
- Never trust hidden fields or client-calculated amounts.
- Password rules should avoid logging raw values.
- Payment card details should not be stored in frontend state.

## 13. Final Checklist

| Item | Status |
|------|--------|
| React Hook Form found | Completed |
| Zod schemas found | Completed |
| Resolver dependency found | Completed |
| Validation tests found | Completed |
| Backend validation gap noted | Completed |
