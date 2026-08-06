# Postman Collections

Generated from `api/master-api.json`.

## Files

- `ecommerce-all-apis.postman_collection.json`
- `auth-service.postman_collection.json`
- `user-service.postman_collection.json`
- `api-gateway-service.postman_collection.json`
- `product-service.postman_collection.json`
- `search-service.postman_collection.json`
- `cart-service.postman_collection.json`
- `wishlist-service.postman_collection.json`
- `order-service.postman_collection.json`
- `payment-service.postman_collection.json`
- `cms-service.postman_collection.json`
- `session-service.postman_collection.json`
- `notification-service.postman_collection.json`
- `superadmin-service.postman_collection.json`
- `ecommerce-local.postman_environment.json`

## Usage

1. Import `ecommerce-local.postman_environment.json` into Postman.
2. Import either the all-in-one collection or individual service collections.
3. Keep `base_url` as `http://localhost:8080` for the API Gateway, or change it for another gateway environment.
4. Run `Auth Login` and copy the token into the role-specific token variable you need, for example `buyer_access_token`, `seller_access_token`, `admin_access_token`, or `superadmin_access_token`.

Request bodies include required and optional fields where the API schema or handlers define them. Optional query parameters are included in the requests and are disabled by default unless they are useful starter parameters.

Regenerate with:

```bash
node scripts/generate-postman-collections.mjs
```
