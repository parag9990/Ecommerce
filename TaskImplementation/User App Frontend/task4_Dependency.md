# Project Dependency & Setup Guide

## Document Variables

| Variable | Value |
|---|---|
| `SERVICE_NAME` | `User App Frontend` |
| `TASK_FILE_NAME` | `task4.md` |
| `INPUT_FILE_PATH` | `TaskImplementation/{SERVICE_NAME}/{TASK_FILE_NAME}` |
| `OUTPUT_FILE_NAME` | `task4_Dependency.md` |
| `OUTPUT_FILE_PATH` | `TaskImplementation/{SERVICE_NAME}/{OUTPUT_FILE_NAME}` |

This guide explains only the dependency, setup, environment, and DevOps impact for `TASK_FILE_NAME`.
Business logic and implementation steps are intentionally not rewritten here.

## 1. Project Overview

`TASK_FILE_NAME` covers product browsing:

- Home product listing.
- Category listing.
- Search results page.
- Filters, sort, pagination, loading, empty, and error states.
- Product detail page.
- Public Product/Search API integration through API Gateway.

Simple Hinglish goal:

"Buyer ko products browse karne ka complete frontend experience milna chahiye. Local UI run karne ke liye Node/pnpm enough hai, but real product data ke liye backend Product Service, Search Service, API Gateway, and search/index dependencies running hone chahiye."

### Task-specific files checked

| File | Why checked |
|---|---|
| `INPUT_FILE_PATH` | Task 4 scope, boundaries, routes, and API expectations. |
| `frontend/user-app/package.json` | Confirmed no new Task 4 package dependency is required beyond existing workspace dependencies. |
| `frontend/user-app/.env.example` | Existing public frontend env variables checked. |
| `frontend/user-app/src/features/product/api/product.api.ts` | Product, category, search, and autocomplete API calls checked. |
| `frontend/user-app/src/features/product/product-url-state.ts` | URL query params, default page size, filters, and sort handling checked. |
| `frontend/user-app/src/features/product/pages/home-page.tsx` | Home listing behavior checked. |
| `frontend/user-app/src/features/product/pages/category-page.tsx` | Category filters/sort/pagination behavior checked. |
| `frontend/user-app/src/features/search/pages/search-page.tsx` | Search route behavior checked. |
| `frontend/user-app/src/features/product/pages/product-detail-page.tsx` | Product detail and later-task side effects checked. |
| `frontend/user-app/src/routes/index.tsx` | Routes for home, categories, category, search, and product detail checked. |
| `api/master-api.json` | Product/Search REST contract checked. |
| `backend/services/product-service/.env` | Product backend runtime dependencies checked. |
| `backend/services/search-service/.env` | Search backend runtime dependencies checked. |
| `backend/services/api-gateway/.env` | API Gateway port and downstream service assumptions checked. |

### What is new for this task

| Area | Status | Beginner explanation |
|---|---|---|
| Public Product APIs | New Task 4 runtime dependency | Product list, detail, and categories real data ke liye required hain. |
| Public Search APIs | New Task 4 runtime dependency | Search result, filters, sort, and autocomplete ke liye required hain. |
| Product Service | Backend dependency for real data | Catalog, categories, variants, inventory snapshot, and product detail source hai. |
| Search Service | Backend dependency for real search | Typesense se ranked results/facets/autocomplete laata hai. |
| Typesense | Backend search engine dependency | Fast product search and autocomplete ke liye use hota hai. Installation details reused. |
| MongoDB | Backend Product Service database | Product/category data store karta hai. Frontend direct use nahi karta. |
| RabbitMQ | Backend event dependency | Product events search indexer tak pahunchane ke liye configured hai. |
| Redis | Backend Search/API cache/rate-limit dependency | Search autocomplete cache and gateway rate limit me use ho sakta hai. |
| Frontend env | Reused | No new frontend env variable was added for Task 4. |
| Frontend package dependency | Reused | No new package install needed if previous tasks are installed. |
| Docker setup | Reused | No new frontend Dockerfile or compose service detected for this task. |

## 2. Tech Stack

Base frontend stack is already explained in previous dependency docs:

Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Sections:
`1. Project Tech Stack Analysis`
`2. Node.js Dependency System`

Also refer:
`TaskImplementation/{SERVICE_NAME}/task2_Dependency.md`

Section:
`2. Tech Stack`

### Task 4 specific technologies

| Technology / Service | Required? | Used by | Hinglish + simple English explanation |
|---|---:|---|---|
| React Router | Required | Product routes and URL query state | Already explained in `task2_Dependency.md`. Task 4 uses it for `/`, `/categories`, `/category/:categoryId`, `/search`, and `/products/:productId`. |
| TanStack React Query | Required by current code | Product list/detail/category/autocomplete hooks | Already explained in `task1_Dependency.md`. Task 4 uses query cache, loading state, errors, and `keepPreviousData` during pagination/filter changes. |
| Fetch API through `lib/http.ts` | Required | `product.api.ts` | Browser ka built-in HTTP client wrapper hai. Ye `VITE_API_BASE_URL` ke saath API Gateway ko call karta hai. |
| URLSearchParams | Required | `product-url-state.ts`, `product.api.ts` | Query params ko safely read/write karta hai. Example: `q`, `category_id`, `page`, `page_size`, `sort`. |
| REST API Gateway | Required for live data | All Product/Search REST calls | Browser direct microservices ko call nahi karta. API Gateway single public backend entrypoint hai. |
| Product Service | Required for live catalog data | `/api/v1/products`, `/api/v1/products/{product_id}`, `/api/v1/categories` | Catalog service product, category, price/variant, and publish status ka source of truth hai. |
| Search Service | Required for search/filter/autocomplete | `/api/v1/search`, `/api/v1/search/autocomplete` | Search Service Typesense se product ranking, facets, sorting, and autocomplete return karta hai. |
| Typesense | Required for real search | Search Service backend | Typesense ek fast search engine hai. Is project me product search index ke liye use hota hai. Setup explanation already exists in previous docs. |
| MongoDB | Required for Product Service | Product backend only | MongoDB products/categories store karta hai kyunki product attributes dynamic ho sakte hain. Frontend direct connect nahi karta. |
| RabbitMQ | Required if product event indexing is enabled | Product Service and Search indexer | Product publish/update event Search Service indexer tak bhejne ke liye queue/broker use hota hai. |
| Redis | Required if backend cache/rate-limit enabled | API Gateway/Search Service backend | Fast cache/rate-limit/autocomplete prefix cache ke liye use ho sakta hai. |

### API endpoints used by Task 4

| Feature | Frontend function | REST endpoint | Backend service | Auth |
|---|---|---|---|---|
| Home product listing | `listProducts` | `GET /api/v1/products?page=1&page_size=24&status=published` | Product Service | Public |
| Product detail | `getProduct` | `GET /api/v1/products/{product_id}` | Product Service | Public |
| Categories | `listCategories` | `GET /api/v1/categories` | Product Service | Public |
| Search result page | `searchProducts` | `GET /api/v1/search?q=...&filter=...&sort=...` | Search Service | Public |
| Header autocomplete | `autocompleteProducts` | `GET /api/v1/search/autocomplete?q=...&limit=6` | Search Service | Public |

Important note:

`api/master-api.json` describes `SearchRequest.filters` as an object, but current frontend sends repeated query params like `filter=brand:Nike` and `filter=price:>=100`. Gateway/Search Service must support this query format, or Task 4 search filters will not work correctly.

## 3. Required Software

Do not repeat full installation steps here. They are already documented.

Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Sections:
`2. Node.js Dependency System`
`3. Database Analysis`
`5. External Services Analysis`
`7. Docker and DevOps Setup`

### Required for frontend-only UI

| Software | Required? | Version / Port | Status |
|---|---:|---|---|
| Node.js | Yes | `>=22.13.0` | Reused from previous docs |
| pnpm | Yes | `>=11.5.0` | Reused from previous docs |
| Vite dev server | Yes | Usually `5173` | Reused |
| Browser | Yes | Chrome/Edge/Firefox | Reused |

### Required for real Product/Search data

| Software / Service | Required for Task 4 live data? | Default / detected config | Status |
|---|---:|---|---|
| API Gateway | Yes | `http://localhost:8080` | Reused entrypoint, Task 4 depends on Product/Search routes |
| Product Service | Yes | Gateway target: `PRODUCT_GRPC_ADDR=localhost:50053` | Task-specific backend dependency |
| Search Service | Yes | Gateway target: `SEARCH_GRPC_ADDR=localhost:50058`; service HTTP env has `:8085` | Task-specific backend dependency |
| MongoDB | Yes for Product Service | `PRODUCT_MONGO_URI=mongodb://localhost:27017` | Reused DB setup |
| Typesense | Yes for Search Service | `localhost:8108`, key `dev-typesense-key` | Reused search setup |
| RabbitMQ | Yes if event indexer enabled | Product/Search env both define `RABBITMQ_URL` | Reused queue setup, config must be aligned |
| Redis | Yes if cache/rate-limit enabled | Gateway `REDIS_ADDR=localhost:6379`; Search `SEARCH_REDIS_ADDR=localhost:6379` | Reused cache setup |
| Docker | Optional but recommended | For MongoDB/Redis/RabbitMQ/Typesense local services | Reused setup |

## 4. Dependency Management

This is still a Node.js + pnpm workspace frontend.

Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`2. Node.js Dependency System`

### Package files involved

| File | Purpose |
|---|---|
| `frontend/package.json` | Workspace-level scripts like `dev:user`, `build:user`, `typecheck:user`. |
| `frontend/pnpm-workspace.yaml` | Defines `user-app` and shared packages. |
| `frontend/pnpm-lock.yaml` | Locks exact dependency versions. |
| `frontend/user-app/package.json` | User app dependencies and scripts. |

### Task 4 dependency status

No new npm/pnpm dependency is required specifically for product browsing.

Existing packages already used by Task 4:

| Package | Why used |
|---|---|
| `react` / `react-dom` | Render product browsing UI. |
| `react-router-dom` | Routes and URL query params. |
| `@tanstack/react-query` | Product/search/category request caching and loading states. |
| `lucide-react` | Icons in current UI components if needed. |
| `typescript` | Type-safe product/search contracts. |
| `vitest` / Testing Library | Product card and URL state tests. |

### Install command

Only run the normal workspace install if dependencies are not installed:

```bash
cd frontend
pnpm install
```

No Task 4-specific `pnpm add` command is needed.

### Quality checks

Use existing scripts:

```bash
cd frontend
pnpm --filter user-app typecheck
pnpm --filter user-app test
pnpm --filter user-app build
```

## 5. Database Setup

### Direct frontend database status

| Database | Directly used by browser frontend? | Required to start Vite? | Required for real Task 4 data? | Status |
|---|---:|---:|---:|---|
| MongoDB | No | No | Yes, through Product Service | Reused setup |
| Typesense | No | No | Yes, through Search Service | Reused setup |
| Redis | No | No | Backend-dependent, Search/Gateway config uses it | Reused setup |
| MySQL | No | No | Not directly for Task 4 product browsing | Reused/general backend setup |
| PostgreSQL/SQLite/Cassandra | No | No | Not detected for Task 4 | Not introduced |

Beginner explanation:

"Frontend browser app database username/password directly use nahi karta. Product data backend Product Service se aata hai. Search result backend Search Service se aata hai. Isliye MongoDB, Typesense, Redis, RabbitMQ credentials frontend `.env` me kabhi nahi daalne."

### MongoDB for Product Service

Product Service uses MongoDB for:

- `products`
- `categories`
- `brands`
- `inventory_snapshots`
- `price_books`

Task-specific detected env:

```env
PRODUCT_MONGO_URI=mongodb://localhost:27017
PRODUCT_MONGO_DATABASE=product_db
PRODUCT_MONGO_AUTO_CREATE_COLLECTIONS=false
```

Setup instructions are already explained in:

Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`3. Database Analysis`

Important Task 4 note:

`PRODUCT_MONGO_AUTO_CREATE_COLLECTIONS=false` means beginner developer ko collections/indexes/seed data setup ka separate backend process follow karna pad sakta hai. Agar categories/products seed nahi hue, frontend empty state dikhayega.

### Typesense for Search Service

Search Service uses Typesense for product search, sorting, facets, and autocomplete.

Task-specific detected env:

```env
TYPESENSE_HOST=localhost
TYPESENSE_PORT=8108
TYPESENSE_PROTOCOL=http
TYPESENSE_API_KEY=dev-typesense-key
TYPESENSE_PRODUCTS_COLLECTION=products
TYPESENSE_POPULAR_QUERIES_COLLECTION=popular_queries
```

Setup instructions are already explained in:

Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`5. External Services Analysis`

Task-specific warning:

Even if Product Service has products in MongoDB, search page can still show empty results if Typesense index is not populated or indexer is not consuming product events.

### Where database credentials belong

| Credential/config | Correct location | Never put in |
|---|---|---|
| `PRODUCT_MONGO_URI` | `backend/services/product-service/.env`, Docker Compose env, or secret manager | `frontend/user-app/.env.local` |
| `TYPESENSE_API_KEY` | `backend/services/search-service/.env`, Docker secret, or secret manager | Frontend env |
| `SEARCH_REDIS_PASSWORD` | `backend/services/search-service/.env`, Docker secret, or secret manager | Frontend env |
| `RABBITMQ_URL` | Backend service `.env` files or Docker secret | Frontend env |

## 6. Redis / Queue / External Services

Generic Redis, RabbitMQ, Kafka, Typesense, Docker, API Gateway, and external service explanations are already covered.

Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Sections:
`5. External Services Analysis`
`7. Docker and DevOps Setup`

### Task 4 live backend services

| Service | Used directly by frontend? | Required for real browsing? | Why needed |
|---|---:|---:|---|
| API Gateway | Yes, through `VITE_API_BASE_URL` | Yes | Browser calls only gateway URLs. |
| Product Service | No, behind gateway | Yes | Product list, product detail, categories. |
| Search Service | No, behind gateway | Yes | Search results, filters, sort, autocomplete. |
| Typesense | No, behind Search Service | Yes for search | Fast product search index. |
| MongoDB | No, behind Product Service | Yes | Product/category source data. |
| RabbitMQ | No, backend only | Yes if event indexer enabled | Product update events to search indexer. |
| Redis | No, backend only | Config-dependent | Search cache and gateway rate limiting. |
| Kafka | No | No new Task 4 usage detected | General architecture may mention it, but Task 4 env uses RabbitMQ. |

### Task-specific service checks

After backend stack is running, these checks help verify Task 4 dependencies:

```bash
curl http://localhost:8080/api/v1/categories
curl "http://localhost:8080/api/v1/products?page=1&page_size=24&status=published"
curl "http://localhost:8080/api/v1/search?q=shoes&page=1&page_size=24"
curl "http://localhost:8080/api/v1/search/autocomplete?q=sh&limit=6"
```

Expected beginner interpretation:

- `200` with categories/products means Gateway + Product Service are working.
- `200` with search response means Gateway + Search Service + Typesense are working.
- Empty `products: []` is not always an error; it may mean no seed data or no search index data.
- `404` usually means gateway route mismatch.
- `502/503` usually means downstream Product/Search Service is not reachable.

### RabbitMQ config alignment warning

Detected values:

```env
# backend/services/product-service/.env
RABBITMQ_URL=amqp://ecommerce:ecommerce@localhost:5672/

# backend/services/search-service/.env
RABBITMQ_URL=amqp://ecommerce:ecommerce_password@localhost:5672/ecommerce
```

These are different credentials/vhost paths. Agar broker me dono users/vhosts configured nahi hain, Product Service events publish karega but Search indexer consume nahi kar paayega. Result: product list may work, but search results/autocomplete may remain stale or empty.

Recommendation:

- Local Docker RabbitMQ credentials ko product-service and search-service env ke saath align karo.
- Same exchange/topic routing keys confirm karo.
- Search indexer logs check karo for authentication, queue declaration, and consumer errors.

## 7. Environment Variables

Frontend env loading and security rules are already documented.

Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`4. Environment Variables`

Also refer:
`TaskImplementation/{SERVICE_NAME}/task3_Dependency.md`

Section:
`7. Environment Variables`

### New frontend env variables

No new frontend env variable was introduced for Task 4.

Task 4 uses existing:

```env
VITE_API_BASE_URL=http://localhost:8080
```

Create frontend local env here when testing real APIs:

```text
frontend/user-app/.env.local
```

Minimal Task 4 frontend env:

```env
VITE_API_BASE_URL=http://localhost:8080
```

Notes:

- `VITE_API_BASE_URL` is public browser config, not a secret.
- Vite server restart required after env changes.
- Product/Search secrets stay in backend `.env` files only.
- `VITE_GRPC_WEB_BASE_URL`, `VITE_GRPC_WEB_TIMEOUT_MS`, and `VITE_PAYMENT_PROVIDERS` exist in `.env.example`, but Task 4 Product/Search browsing uses REST API Gateway paths, not gRPC-Web directly.

### Backend env variables relevant to Task 4

These are not frontend variables. They are listed so a beginner knows which backend config affects live product browsing.

| Variable | Location | Purpose | Required? | Security note |
|---|---|---|---:|---|
| `HTTP_ADDR=:8080` | `backend/services/api-gateway/.env` | Gateway public REST port. | Yes | Public port config, not secret. |
| `PRODUCT_GRPC_ADDR=localhost:50053` | `backend/services/api-gateway/.env` | Gateway target for Product Service. | Yes | Internal address. |
| `SEARCH_GRPC_ADDR=localhost:50058` | `backend/services/api-gateway/.env` | Gateway target for Search Service. | Yes | Internal address. |
| `PRODUCT_MONGO_URI` | `backend/services/product-service/.env` | Product DB connection. | Yes | Treat as secret in non-local env. |
| `PRODUCT_MONGO_DATABASE` | `backend/services/product-service/.env` | Product database name. | Yes | Not secret by itself. |
| `PRODUCT_EVENTS_ENABLED` | `backend/services/product-service/.env` | Enables product event publishing. | Yes for live indexing | Keep enabled if search index updates are expected. |
| `RABBITMQ_URL` | Product/Search backend env files | Broker URL for product events/indexer. | Yes if indexer enabled | Contains credentials; do not expose to browser. |
| `TYPESENSE_*` | `backend/services/search-service/.env` | Typesense connection and collection names. | Yes for search | API key is secret. |
| `SEARCH_REDIS_*` | `backend/services/search-service/.env` | Search cache connection. | Config-dependent | Redis password is secret. |
| `SEARCH_INDEXER_ENABLED` | `backend/services/search-service/.env` | Enables product event consumer/indexer. | Yes for automatic index updates | Check logs if search stale. |
| `PRODUCT_SERVICE_URL` | `backend/services/search-service/.env` | Search Service calls Product Service internal endpoints. | Yes for reindex/export | Must match Product Service internal HTTP address. |

## 8. Docker Setup

No new frontend Dockerfile, volume, network, or container was introduced by Task 4.

Do not duplicate Docker installation/setup here.

Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`7. Docker and DevOps Setup`

### Docker services needed for real Task 4 data

| Container / Service | Port | Why needed | Status |
|---|---:|---|---|
| MongoDB | `27017` | Product/category source data for Product Service | Reused |
| Typesense | `8108` | Product search index | Reused |
| RabbitMQ | `5672`, management usually `15672` | Product events to search indexer | Reused, config alignment needed |
| Redis | `6379` | Gateway rate limit and Search cache if enabled | Reused |
| API Gateway | `8080` | Browser-facing REST entrypoint | Reused |
| Product Service | `50053` via gateway config | Product gRPC backend | Task-specific runtime |
| Search Service | `50058` via gateway config, `8085` in service env | Search gRPC/HTTP backend | Task-specific runtime |

### Docker network notes

- Host machine URLs usually use `localhost`.
- Containers in same compose network should use service names like `mongodb`, `typesense`, `rabbitmq`, `redis`.
- If backend service runs inside Docker but env says `localhost`, it may point to the container itself, not host machine. Use compose service names in container env.
- Keep RabbitMQ credentials/vhost same for Product Service publisher and Search Service consumer.

## 9. Local Development Setup

### Step 1: Read previous dependency documentation first

Follow these first to avoid duplicate setup:

1. `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`
2. `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md`
3. `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md`

They already explain:

- Node.js and pnpm setup.
- Vite, React, Tailwind, TypeScript setup.
- `.env.local` location and Vite env rules.
- Generic Docker, MongoDB, Redis, Typesense, RabbitMQ setup.
- Generic CORS, network, and dependency troubleshooting.

### Step 2: Go to frontend workspace

```bash
cd frontend
```

### Step 3: Install dependencies

```bash
pnpm install
```

No new Task 4 package install is required.

### Step 4: Configure frontend env for live APIs

Create or update:

```text
frontend/user-app/.env.local
```

Use:

```env
VITE_API_BASE_URL=http://localhost:8080
```

Restart Vite after changing env.

### Step 5: Start backend services only for real product browsing

For static UI rendering, backend is optional.

For real products/search:

- Start MongoDB.
- Start Typesense.
- Start Redis if gateway/search cache or rate-limit is enabled.
- Start RabbitMQ if product event indexing is enabled.
- Start Product Service with product DB config.
- Start Search Service with Typesense, Redis, RabbitMQ, and Product Service config.
- Start API Gateway with Product/Search downstream addresses.

Exact backend startup commands are backend-service specific and are not clearly centralized in Task 4. Use backend service docs/scripts where available.

### Step 6: Seed data and index search

Task 4 UI needs data:

- At least one category.
- At least one published product.
- Product variants with price and stock.
- Product images if visual cards/detail should look complete.
- Typesense product index populated for search results.

If Product Service has MongoDB products but Search Service returns empty:

- Check RabbitMQ event flow.
- Check `SEARCH_INDEXER_ENABLED=true`.
- Run backend reindex command/admin endpoint if available.
- Check Search Service logs for Typesense import/index errors.

### Step 7: Start frontend

```bash
pnpm --filter user-app dev
```

Alternative workspace script:

```bash
pnpm dev:user
```

Open the Vite URL printed in terminal, usually:

```text
http://localhost:5173
```

## 10. Running the Project

### Task 4 routes to verify

| Route | What to check | Backend needed for real data? |
|---|---|---:|
| `/` | Home product listing and category strip. | Yes |
| `/categories` | Category browsing page. | Yes |
| `/category/:categoryId` | Category-specific listing, filters, sort, pagination. | Yes |
| `/search?q=shoes` | Search results, filters, sort, pagination. | Yes |
| `/products/:productId` | Product detail page. | Yes |

### API verification from terminal

```bash
curl http://localhost:8080/api/v1/categories
curl "http://localhost:8080/api/v1/products?page=1&page_size=24&status=published"
curl "http://localhost:8080/api/v1/search?q=shoes&page=1&page_size=24"
curl "http://localhost:8080/api/v1/search/autocomplete?q=sh&limit=6"
```

### Quality verification

```bash
cd frontend
pnpm --filter user-app typecheck
pnpm --filter user-app test
pnpm --filter user-app build
```

Task 4 has useful tests around:

- Product card detail link/render behavior.
- Product URL state parsing/writing.

## 11. Common Errors & Fixes

Generic Node, pnpm, Docker, env, CORS, and database errors are already documented in:

Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`9. Common Errors and Fixes`

Also refer:
`TaskImplementation/{SERVICE_NAME}/task2_Dependency.md`

Section:
`11. Common Errors & Fixes`

### Task 4 specific errors

| Error / Symptom | Likely cause | Fix | Prevention |
|---|---|---|---|
| Product grid shows network error | API Gateway stopped or `VITE_API_BASE_URL` wrong | Start gateway and verify `curl http://localhost:8080/api/v1/products` | Keep `.env.local` aligned with gateway port. |
| `/api/v1/products` returns `404` | Gateway route missing or API contract not loaded | Check `api/master-api.json` and gateway route config | Keep gateway contract and frontend API paths synced. |
| Categories load but product list fails | Product Service route/detail/list issue or MongoDB unavailable | Check Product Service logs and MongoDB connection | Start Product Service only after MongoDB is ready. |
| Product list works but search results are empty | Typesense index missing/stale, RabbitMQ/indexer not working, or no matching products | Check Search Service logs, Typesense collection, RabbitMQ consumer, and reindex process | Seed products and run reindex after local DB reset. |
| Filters do nothing | Gateway/Search Service does not understand repeated `filter=key:value` query params | Align API contract with frontend format or update frontend/backend parser together | Document one canonical filter query format. |
| Sort ignored | Search Service does not support sort value like `price:asc` or `popularity_score:desc` | Compare supported Search Service sort fields | Keep sort options synced with search index schema. |
| Autocomplete not showing | Query shorter than 2 chars, Search Service down, Typesense unavailable, or cache issue | Type at least 2 chars and verify autocomplete endpoint | Add smoke test for `/api/v1/search/autocomplete`. |
| Product detail says not found | Invalid product id, product unpublished/deleted, or backend returned empty response | Open a known published product id from list response | Use links generated from API data, not manually typed ids. |
| Product cards show blank image | API does not return `images`, image URL broken, or CORS/CDN issue | Check product response and image URL in browser Network tab | Store valid HTTPS image URLs and use fallback UI. |
| Rating not visible | API does not return optional `rating` field | Confirm product response schema | Treat rating as optional until backend supports it. |
| Add to cart/wishlist button causes auth/cart/wishlist errors | Current product detail/card includes later-task buttons | Login and start cart/wishlist backend, or ignore when verifying pure Task 4 browsing | Remember Task 4 browsing is public; mutations belong to later tasks. |
| `page_size` mismatch | Frontend default is `24`, backend default may be `20`; frontend max is `60`, backend max may be `100` | Use explicit `page_size` and keep max values agreed | Share pagination constants between docs/API/frontend where possible. |
| CORS error in browser | Gateway does not allow Vite origin | Allow `http://localhost:5173` with credentials if cookies are used | Maintain local/staging/prod origin allowlist. |

## 12. Security & Best Practices

### Security audit for Task 4

| Area | Status / Risk | Recommendation |
|---|---|---|
| Frontend secrets | No Task 4 secret should be in frontend env. | Keep only `VITE_API_BASE_URL` in frontend. |
| Typesense API key | Backend env contains `TYPESENSE_API_KEY`. | Never expose Typesense admin/search key to browser. |
| MongoDB URI | Backend env contains MongoDB URI. | Use secret manager or Docker secrets outside local. |
| RabbitMQ URL mismatch | Product and Search env use different RabbitMQ URLs. | Align local broker credentials/vhost before testing indexer. |
| Public Product/Search APIs | Public routes can receive arbitrary query params. | Gateway/Search Service should validate query size, page size, sort allowlist, and filter fields. |
| CORS with credentials | `lib/http.ts` uses `credentials: 'include'`. | Gateway must allow only trusted origins when credentials are enabled. |
| Product image URLs | Browser loads image URLs from product data. | Prefer HTTPS CDN URLs and safe fallback images. |
| Search index freshness | Product data and search index can drift. | Monitor indexer lag and expose reindex/admin tooling. |

### Task-specific best practices

- Keep product/search API paths centralized in `features/product/api/product.api.ts`.
- Keep filter names synced across frontend, gateway, and Search Service.
- Keep sort options synced with Typesense schema.
- Store product browsing state in URL query params so refresh/share/back button work.
- Use `URLSearchParams`, not manual query string concatenation.
- Keep `page_size` bounded to avoid heavy API calls.
- Treat Product/Search APIs as public read APIs, but still rate-limit them.
- Seed local categories/products before asking a beginner to verify UI.
- After DB reset, run search reindex or confirm RabbitMQ indexer consumed product events.
- Do not add database, Redis, RabbitMQ, or Typesense credentials to browser env.

## 13. Missing or Misconfigured Things

| Item | Detected issue | Why it matters | Suggested fix |
|---|---|---|---|
| Search filter contract | `api/master-api.json` says `filters` object, frontend sends repeated `filter=key:value`. | Search filters may silently fail. | Document and implement one gateway query format. |
| RabbitMQ credentials/vhost | Product Service and Search Service env values differ. | Product events may not reach indexer. | Align `RABBITMQ_URL` values or provision both users/vhosts intentionally. |
| Product Service internal URL for Search | Search env expects `PRODUCT_SERVICE_URL=http://localhost:8082`, but Product Service env shown does not clearly expose matching HTTP config. | Reindex/export may fail. | Confirm Product Service internal HTTP port and paths. |
| Seed/reindex commands | No clear Task 4 command found for seed data or search reindex. | Beginner can run frontend but see empty products/search. | Add backend seed and reindex docs/scripts. |
| Frontend Dockerfile | No user-app Dockerfile detected for Task 4. | Production packaging is not one-command from this task. | Add frontend Dockerfile/Nginx config when deployment packaging starts. |
| Local compose file | No single compose file detected for full Task 4 backend stack. | Beginner setup remains multi-step. | Add local compose for API Gateway, Product, Search, MongoDB, Typesense, Redis, RabbitMQ. |
| Product response schema | Current UI can use `images`, `rating`, and variant stock; master `Product` schema is narrower. | UI may show fallback/blank values. | Align API schema with frontend display fields or update UI expectations. |
| Later-task buttons on product UI | Product card/detail can include cart/wishlist actions. | Pure browsing verification may trigger auth/cart/wishlist dependencies. | Document as later-task side effect or feature-flag mutations during Task 4 testing. |

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `1. Project Tech Stack Analysis` | Base React, TypeScript, Vite, Tailwind, React Query, Zustand, Fetch, and testing stack already explained. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `2. Node.js Dependency System` | Node.js, pnpm, workspace install, lockfile, scripts, and dependency troubleshooting already documented. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `3. Database Analysis` | Frontend has no direct DB connection; MongoDB/Redis and backend DB setup already covered. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `4. Environment Variables` | `.env.local`, Vite `VITE_` prefix, env restart, and frontend env security already documented. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `5. External Services Analysis` | API Gateway, Typesense, Redis, RabbitMQ, Kafka, Docker, and external service explanations already exist. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `6. Ports and Networking` | Existing ports, port conflicts, CORS, host/container networking already documented. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `7. Docker and DevOps Setup` | Docker recommendations, database containers, volumes, and backend service container guidance already covered. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `9. Common Errors and Fixes` | Generic Node, pnpm, Docker, DB, Redis, env, and CORS troubleshooting already documented. |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | `2. Tech Stack` | React Router and App Shell route/provider setup already explained. |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | `6. Redis / Queue / External Services` | Header search/autocomplete and gateway side effects already introduced. |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | `11. Common Errors & Fixes` | Router/provider/search-shell issues already covered. |
| `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` | `7. Environment Variables` | `VITE_API_BASE_URL` for real API calls already documented. |
| `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` | `11. Common Errors & Fixes` | API Gateway down, CORS, and auth-related browser request errors already covered. |

## 15. Final Checklist

- [ ] Previous dependency documentation checked before using this file.
- [ ] No duplicate Node.js, pnpm, Vite, Tailwind, generic Docker, or generic DB setup copied.
- [ ] `frontend/user-app/.env.local` created only if real backend APIs are being tested.
- [ ] `VITE_API_BASE_URL` points to the running API Gateway.
- [ ] API Gateway running on `http://localhost:8080`.
- [ ] Product Service reachable through gateway Product routes.
- [ ] Search Service reachable through gateway Search routes.
- [ ] MongoDB running and Product Service connected.
- [ ] Product categories/products seeded for local testing.
- [ ] Typesense running and product index populated.
- [ ] RabbitMQ Product/Search URLs aligned if event indexing is enabled.
- [ ] Redis running if gateway/search config requires it.
- [ ] Search filter query format confirmed between frontend and backend.
- [ ] Product list route `/` verified.
- [ ] Category route `/category/:categoryId` verified.
- [ ] Search route `/search?q=...` verified.
- [ ] Product detail route `/products/:productId` verified.
- [ ] Autocomplete endpoint verified with query length 2+.
- [ ] Later-task cart/wishlist button errors not confused with Task 4 browsing errors.
- [ ] Browser console and Network tab checked.
- [ ] `pnpm --filter user-app typecheck` passes.
- [ ] `pnpm --filter user-app test` passes.
- [ ] `pnpm --filter user-app build` passes.
- [ ] No backend secrets added to frontend env.
- [ ] No original implementation task file modified.
