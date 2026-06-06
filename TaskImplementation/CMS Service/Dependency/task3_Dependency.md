# Dependency + Setup Documentation

## Document Variables

| Variable | Value |
|---|---|
| `SERVICE_NAME` | `CMS Service` |
| `TASK_FILE_NAME` | `task3.md` |
| `INPUT_FILE_PATH` | `TaskImplementation/${SERVICE_NAME}/${TASK_FILE_NAME}` |
| `OUTPUT_FILE_NAME` | `task3_Dependency.md` |
| `OUTPUT_FILE_PATH` | `TaskImplementation/${SERVICE_NAME}/${OUTPUT_FILE_NAME}` |

This file documents dependency, setup, environment, database, and DevOps requirements for `INPUT_FILE_PATH`.

Simple goal: beginner developer ko clear ho ki product moderation flow run karne ke liye kaunsi common setup reuse karni hai, kaunsi task-specific table/env/service dependency check karni hai, aur local debugging kaise karni hai.

Important: original implementation task file is not modified. This is a new setup/dependency document saved at `OUTPUT_FILE_PATH`.

## 1. Previous Dependency File Reuse

Previous dependency files in `TaskImplementation/${SERVICE_NAME}/` were checked first:

| Previous file | Reused topic | What to do here |
|---|---|---|
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | Go setup, Go modules, MySQL install, MySQL Docker command, full `.env`, service run commands, generic troubleshooting | Follow that file first. Do not duplicate those steps here. |
| `TaskImplementation/${SERVICE_NAME}/task2_Dependency.md` | MySQL 8+ decision, schema ownership, `cms_db`, InnoDB, UTF-8, migration order, DB verification | Reuse the same DB setup. This file only explains the product moderation table and checks. |

Refer:
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`

Sections:
`2. Tech Stack Analysis`
`3. Go Dependency System`
`4. Database Analysis`
`5. Environment Variables`
`8. Docker and DevOps Setup`
`9. Complete Project Run Instructions`
`11. Common Errors and Fixes`

Refer:
`TaskImplementation/${SERVICE_NAME}/task2_Dependency.md`

Sections:
`5. Database Setup`
`7. Environment Variables`
`10. Running the Project`

## 2. What `TASK_FILE_NAME` Adds

`TASK_FILE_NAME` designs the product moderation flow: seller draft, submit review, moderator approve/reject, publish, and unpublish.

Current backend implementation under `backend/services/cms-service` contains task-specific support for this flow:

| Area | File/path | Setup impact |
|---|---|---|
| Moderation domain rules | `internal/domain/product_moderation.go` | No external install. Defines states, transitions, validation, major/minor edit logic. |
| Moderation usecase | `internal/usecase/product_moderation.go` | Needs Product Service client, MySQL review repository, and audit recorder. |
| Review repository | `internal/repository/mysql_product_moderation_repository.go` | Requires MySQL table `product_moderation_reviews`. |
| Migration | `migrations/002_create_product_moderation_reviews.up.sql` | Must be applied with all other service migrations. |
| Product client | `internal/clients/product_client.go` | Requires a valid Product Service base URL and optional internal token. |
| HTTP routes | `internal/transport/http/handler.go` | Product moderation routes require internal auth headers and actor headers. |
| Audit log writes | `internal/domain/audit.go` and MySQL audit repository | Requires shared audit table setup from previous tasks. |

No new Go package, database engine, Docker container, Redis cache, queue, or port is introduced by `TASK_FILE_NAME`.

## 3. Tech Stack Analysis

Most stack explanation is already covered in previous dependency files. This section only highlights what matters for product moderation.

| Technology/service | Status | Why it matters for `TASK_FILE_NAME` | Beginner explanation |
|---|---|---|---|
| Go | Reused | Moderation business rules, HTTP handlers, Product Service client, and MySQL repository are implemented in Go. | Go backend language hai. Fast service logic aur APIs banane ke liye use hota hai. |
| Go modules | Reused | `go.mod` has no new direct dependency for this task. | `go.mod` dependency list hai; `go.sum` checksum safety file hai. |
| `net/http` | Reused | Product moderation HTTP routes are registered on the existing service HTTP server. | Ye Go ka built-in HTTP server hai. |
| MySQL 8+ | Reused, task-specific table added | `product_moderation_reviews` stores review queue and decisions. | MySQL table-based database hai. Review records rows ke form me store hote hain. |
| Product Service HTTP API | Task-specific runtime dependency | Product status and product details remain owned by Product Service. | Product Service catalog ka owner hai. Moderation ke time ye service product DB direct access nahi karti. |
| API Gateway/Auth headers | Reused | Actor context comes from trusted headers like `X-User-ID`, `X-Roles`, and `X-Seller-ID`. | Gateway JWT validate karke service ko identity headers deta hai. |
| Audit logging | Reused, task-specific actions added | Submit, approve, reject, publish, and unpublish actions write audit entries. | Sensitive action ka record banta hai, taaki later debugging/compliance easy ho. |
| MongoDB | Not directly used by this service | Product Service may own product MongoDB, but this service does not connect to MongoDB. | Is task ke liye MongoDB container start karna zaruri nahi, unless Product Service itself needs it. |
| Kafka/RabbitMQ/Typesense/Search | Design reference only | `TASK_FILE_NAME` mentions future product events/search indexing, but current code does not connect to them. | Abhi local setup me queue/search run karna required nahi hai. |

## 4. Go Dependency System

No new direct Go dependency is added by `TASK_FILE_NAME`.

Current direct dependencies remain:

| Dependency | Used for |
|---|---|
| `github.com/go-sql-driver/mysql` | MySQL connection and repository queries. |
| `google.golang.org/grpc` | Existing internal gRPC server. Product moderation itself is exposed over HTTP in current code. |
| `google.golang.org/protobuf` | Existing generated proto support. |

Dependency install, `go mod download`, `go mod tidy`, `go build`, and `go run` are already explained in:

```md
Refer:
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`

Sections:
`3. Go Dependency System`
`9. Complete Project Run Instructions`
`15. Minimal Local Command Flow`
```

Task-specific note: do not run `go get` for product moderation. The code uses existing standard library, MySQL driver, and service-local packages.

Optional validation from service folder:

```bash
cd backend/services/cms-service
go test ./...
```

## 5. Database Analysis

### 5.1 Database used

| Database | Required? | Default port | Status |
|---|---:|---:|---|
| MySQL 8+ | Yes | 3306 | Reused setup with task-specific `product_moderation_reviews` table. |

MySQL installation, Docker run command, DB user creation, and full migration flow are already explained in previous dependency files. Use those steps first.

### 5.2 Task-specific table

`TASK_FILE_NAME` requires this table:

```text
product_moderation_reviews
```

Purpose: seller jab product review ke liye submit karta hai, CMS ek review record create karta hai. Moderator approve/reject karta hai, to wahi record decision, reviewer, reason, and timestamps store karta hai.

Important columns:

| Column | Purpose |
|---|---|
| `review_id` | Public/internal review identifier returned by CMS. |
| `product_id` | Product Service product id. No cross-service foreign key. |
| `seller_id` | Seller owner for filtering and audit scope. |
| `status` | Review status: `submitted`, `approved`, `rejected`, `cancelled`. |
| `submitted_by` | Seller actor user id. |
| `reviewed_by` | Catalog admin or superadmin actor user id. |
| `rejection_reason` | Required when review is rejected. |
| `submitted_at`, `reviewed_at`, `created_at`, `updated_at` | Timeline/debugging fields. |
| `active_submitted_product_id` | Generated helper column to enforce one active submitted review per product. |

Important keys and constraints:

| Key/constraint | Why it matters |
|---|---|
| `uk_product_moderation_review_id` | Prevents duplicate `review_id`. |
| `uk_product_moderation_active_submitted` | Allows only one active `submitted` review per product. |
| `idx_product_moderation_seller_status` | Admin/seller filtering by seller and status is faster. |
| `idx_product_moderation_product` | Product-specific lookup is faster. |
| `idx_product_moderation_status_submitted` | Review queue sorted by submitted time is faster. |
| `chk_product_moderation_rejection_reason` | Rejected reviews must have a rejection reason. |

### 5.3 Migration requirement

Task-specific migration:

```text
backend/services/cms-service/migrations/002_create_product_moderation_reviews.up.sql
```

Important: full service startup initializes many repositories, not only product moderation. For local service run, apply all service migrations in numeric order as documented in `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`.

If you already ran every migration from the previous dependency guide, do not rerun this migration manually.

Task-specific verification:

```bash
cd backend/services/cms-service
mysql -h 127.0.0.1 -P 3306 -u cms_user -p -e "SHOW TABLES LIKE 'product_moderation_reviews';" cms_db
mysql -h 127.0.0.1 -P 3306 -u cms_user -p -e "SHOW INDEX FROM product_moderation_reviews;" cms_db
```

Expected:

```text
product_moderation_reviews
uk_product_moderation_review_id
uk_product_moderation_active_submitted
idx_product_moderation_seller_status
idx_product_moderation_product
idx_product_moderation_status_submitted
```

Check table DDL when debugging:

```bash
mysql -h 127.0.0.1 -P 3306 -u cms_user -p -e "SHOW CREATE TABLE product_moderation_reviews\\G" cms_db
```

### 5.4 Connection string and credentials

No new DB credentials are introduced by `TASK_FILE_NAME`.

Use the DB variables already documented in:

```md
Refer:
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`

Section:
`5. Environment Variables`
```

Task-specific reminder:

| Variable | Why it matters here |
|---|---|
| `CMS_MYSQL_DSN` | If set, this must point to the DB where `product_moderation_reviews` exists. |
| `CMS_DB_NAME` | Should be `cms_db` for the documented local setup. |
| `CMS_DB_TIMEZONE` | Use `UTC` so review timestamps are predictable. |
| `CMS_STAFF_STATUS_SOURCE` | `gateway` trusts headers; `mysql` requires staff table data from previous setup. |

## 6. Environment Variables

No new environment variable is introduced by `TASK_FILE_NAME`.

Do not duplicate the full `.env`. Use the complete `.env` from:

```md
Refer:
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`

Section:
`5. Environment Variables`
```

Task-specific variables to verify:

| Variable | Required? | Example | Purpose |
|---|---:|---|---|
| `CMS_PRODUCT_SERVICE_BASE_URL` | Required for real moderation calls | `http://localhost:8080` | Base URL used by the Product Service HTTP client. |
| `CMS_PRODUCT_SERVICE_TIMEOUT` | Optional | `5s` | Timeout for Product Service calls. |
| `CMS_PRODUCT_SERVICE_AUTH_HEADER` | Required when Product Service expects internal auth | `X-Internal-Token` | Header name sent from this service to Product Service. |
| `CMS_PRODUCT_SERVICE_AUTH_TOKEN` | Strongly recommended | `local-product-internal-token` | Token value sent to Product Service. Do not commit real token. |
| `CMS_INTERNAL_AUTH_HEADER` | Required | `X-Internal-Token` | Header name required by this service on internal HTTP routes. |
| `CMS_INTERNAL_AUTH_TOKEN` | Required in production, recommended locally | `local-cms-internal-token` | If set, every moderation route must include this token. |
| `CMS_DB_*` / `CMS_MYSQL_DSN` | Required for startup | See previous dependency file | MySQL connection where moderation table exists. |

Where to create `.env`:

```text
backend/services/cms-service/.env
```

Important beginner note: current code uses `os.Getenv`. It does not auto-load `.env`. Source it before running the service, exactly as shown in the previous dependency file.

## 7. External Services Analysis

### 7.1 Product Service

Product Service is required for actual submit, approve/reject, publish, and unpublish flows.

Why: Product data and product status are owned by Product Service. This service stores review records and policy/audit history, but it does not directly read or write the Product Service database.

Expected Product Service internal endpoints used by the current client:

| Method | Path | Used for |
|---|---|---|
| `GET` | `/internal/v1/products/{product_id}` | Fetch product detail before validation/decision. |
| `PATCH` | `/internal/v1/products/{product_id}/status` | Move product to `submitted`, `approved`, or `rejected`. |
| `POST` | `/internal/v1/products/{product_id}/publish` | Publish an approved product. |
| `POST` | `/internal/v1/products/{product_id}/unpublish` | Unpublish a product. |

The Product Service response must include enough fields for validation:

```json
{
  "product": {
    "id": "prod_123",
    "seller_id": "seller_456",
    "title": "Running Shoes",
    "description": "Comfortable running shoes for daily use.",
    "category_id": "cat_shoes",
    "brand": "Acme",
    "status": "draft",
    "images": [{"url": "https://example.com/image.jpg"}],
    "variants": [
      {
        "sku": "SHOE-001",
        "price": {"amount": 199900, "currency": "INR"}
      }
    ]
  }
}
```

Current local repo observation: `backend/services/product-service` only contains an `.env` file, so a runnable Product Service implementation was not found in the current workspace. That means:

| Scenario | Result |
|---|---|
| Start this service and call `/healthz` | Works if MySQL/config are correct. |
| List moderation reviews | Works without Product Service because it only reads MySQL. |
| Submit/approve/reject/publish/unpublish a product | Needs real Product Service or a local stub at `CMS_PRODUCT_SERVICE_BASE_URL`. |

### 7.2 API Gateway/Auth context

All moderation HTTP routes are internal routes and use trusted headers.

Required/common headers:

| Header | Purpose |
|---|---|
| `X-Internal-Token` | Internal route protection when `CMS_INTERNAL_AUTH_TOKEN` is set. |
| `X-User-ID` | Actor user id. |
| `X-Roles` | Comma-separated roles such as `seller`, `seller_catalog_editor`, `catalog_admin`, `superadmin`. |
| `X-Seller-ID` | Seller workspace id for seller actions. |
| `X-Staff-Status` | Use `active` for local testing when gateway status source is used. |
| `X-Request-ID` | Trace/audit correlation id. |

Security note: in production, these headers must come from a trusted API Gateway/Auth layer. Do not expose internal routes directly to public traffic.

### 7.3 Services not required by this task

| Service | Required? | Reason |
|---|---:|---|
| Redis | No | No cache client is used by current moderation code. |
| Kafka/RabbitMQ | No | Product/search events are design references, not current CMS runtime dependencies. |
| MongoDB | No for this service | Product Service may use MongoDB, but this service does not connect to it. |
| Typesense/Search Service | No | Search indexing is outside current moderation setup. |
| MinIO/S3/SMTP/Stripe/Twilio/Firebase | No | Not used by this task. |

## 8. Ports and Networking

No new port is introduced by `TASK_FILE_NAME`.

| Service | Port | Purpose | Status |
|---|---:|---|---|
| `${SERVICE_NAME}` HTTP API | 8087 | Health and internal moderation routes | Reused |
| `${SERVICE_NAME}` gRPC API | 9098 | Existing internal RPC routes | Reused |
| MySQL | 3306 | `cms_db` and moderation review table | Reused |
| Product Service | 8080 | Internal product detail/status calls | Reused/default dependency |

How to change task-specific networking:

```env
CMS_PRODUCT_SERVICE_BASE_URL=http://localhost:18080
```

Docker networking reminder:

| Where this service runs | Product Service URL example |
|---|---|
| Host machine | `http://localhost:8080` |
| Docker Compose same network | `http://product-service:8080` |

For generic port conflict fixes, refer to `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`, section `7. Ports and Networking`.

## 9. Docker and DevOps Setup

No new Dockerfile, compose service, volume, network, Kubernetes manifest, Redis container, queue container, or search container is introduced by `TASK_FILE_NAME`.

Reuse:

```md
Refer:
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`

Sections:
`4. Database Analysis` -> MySQL Docker setup
`8. Docker and DevOps Setup`
```

Recommended beginner setup remains:

| Component | Run method |
|---|---|
| MySQL | Docker container from previous dependency file. |
| `${SERVICE_NAME}` | `go run ./cmd/server` from host. |
| Product Service | Real service when available, or a local stub for moderation endpoint testing. |

If Product Service/stub runs in Docker and this service runs on host, expose the Product Service port to host and keep `CMS_PRODUCT_SERVICE_BASE_URL=http://localhost:<mapped-port>`.

## 10. Local Development Flow

### Step 1: Follow shared setup

Complete shared setup first:

```text
TaskImplementation/${SERVICE_NAME}/task1_Dependency.md
TaskImplementation/${SERVICE_NAME}/task2_Dependency.md
```

Minimum shared setup needed:

| Setup item | Why |
|---|---|
| Go installed | Service runtime. |
| MySQL running | Service startup requires DB connection. |
| All service migrations applied | Server initializes all repositories. |
| `.env` sourced | Config comes from environment variables. |
| `CMS_INTERNAL_AUTH_TOKEN` known | Needed for internal moderation routes. |

### Step 2: Verify moderation table

```bash
cd backend/services/cms-service
mysql -h 127.0.0.1 -P 3306 -u cms_user -p -e "SHOW TABLES LIKE 'product_moderation_reviews';" cms_db
```

If table is missing, run all migrations in numeric order using the previous dependency guide.

### Step 3: Verify task-specific env

```bash
cd backend/services/cms-service
set -a
source .env
set +a
printf '%s\n' "$CMS_PRODUCT_SERVICE_BASE_URL"
```

Expected local value is usually:

```text
http://localhost:8080
```

### Step 4: Start the service

Use the run command from the previous dependency file:

```bash
cd backend/services/cms-service
go run ./cmd/server
```

Expected logs include:

```text
cms.mysql.connected
cms.http.started
cms.grpc.started
```

### Step 5: Basic health check

```bash
curl http://localhost:8087/healthz
```

Expected:

```json
{"success":true}
```

### Step 6: Task-specific route smoke test without Product Service

This checks internal auth, actor headers, route wiring, permission, and MySQL read path.

```bash
curl -s \
  -H "X-Internal-Token: local-cms-internal-token" \
  -H "X-User-ID: admin_1" \
  -H "X-Roles: catalog_admin" \
  -H "X-Request-ID: local-review-list-1" \
  "http://localhost:8087/internal/v1/cms/admin/catalog/reviews?status=submitted&limit=10"
```

Expected on empty DB:

```json
{"reviews":[],"limit":10,"offset":0}
```

If you get `401`, internal token/header is wrong. If you get `403`, role headers are missing or wrong.

### Step 7: Full moderation flow test with Product Service

Only run this after a real Product Service or local stub is available at `CMS_PRODUCT_SERVICE_BASE_URL`.

Submit review:

```bash
curl -s -X POST \
  -H "X-Internal-Token: local-cms-internal-token" \
  -H "X-User-ID: seller_user_1" \
  -H "X-Seller-ID: seller_456" \
  -H "X-Roles: seller_catalog_editor" \
  -H "X-Staff-Status: active" \
  -H "X-Request-ID: local-submit-review-1" \
  "http://localhost:8087/internal/v1/cms/seller/products/prod_123/submit-review"
```

Approve review:

```bash
curl -s -X POST \
  -H "X-Internal-Token: local-cms-internal-token" \
  -H "X-User-ID: admin_1" \
  -H "X-Roles: catalog_admin" \
  -H "Content-Type: application/json" \
  -H "X-Request-ID: local-approve-review-1" \
  -d '{"decision":"approved","reason":"Product details are valid."}' \
  "http://localhost:8087/internal/v1/cms/admin/catalog/reviews/review_123/decision"
```

Reject review:

```bash
curl -s -X POST \
  -H "X-Internal-Token: local-cms-internal-token" \
  -H "X-User-ID: admin_1" \
  -H "X-Roles: catalog_admin" \
  -H "Content-Type: application/json" \
  -H "X-Request-ID: local-reject-review-1" \
  -d '{"decision":"rejected","reason":"Product image is unclear."}' \
  "http://localhost:8087/internal/v1/cms/admin/catalog/reviews/review_123/decision"
```

Note: replace `prod_123` and `review_123` with actual ids returned by your Product Service/CMS flow.

## 11. Product Moderation Runtime Rules

### States

| Product status | Meaning |
|---|---|
| `draft` | Seller can prepare product before review. |
| `submitted` | Product is waiting for moderation. |
| `approved` | Moderator approved it, but it is not public until publish. |
| `rejected` | Moderator rejected it and seller must fix issues. |
| `published` | Product is live/public. |
| `unpublished` | Product was removed from public listing. |

### Allowed transitions

| From | To |
|---|---|
| `draft` | `submitted` |
| `submitted` | `approved`, `rejected` |
| `rejected` | `draft`, `submitted` |
| `approved` | `published` |
| `published` | `unpublished`, `submitted` |
| `unpublished` | `published`, `draft` |

Setup meaning: if Product Service returns an unexpected status, CMS may return `FAILED_PRECONDITION` or validation errors.

### Submit-review validation

Before product can move to `submitted`, CMS validates Product Service response:

| Field | Requirement |
|---|---|
| `id` / `product_id` | Required. |
| `seller_id` | Required and must match `X-Seller-ID`. |
| `title` | Required, max 180 characters. |
| `description` | At least meaningful content; current minimum is 20 characters. |
| `category_id` | Required. |
| `brand` or generic flag | Brand required unless product is marked generic. |
| `images` | At least one image with URL. |
| `variants` | At least one variant. |
| `variants[].sku` | Required and unique within product. |
| `variants[].price.amount` | Must be greater than zero. |
| `variants[].price.currency` | Valid currencies include `INR`, `USD`, `EUR`, `GBP`, `AED`, `SGD`. |

## 12. Common Task-Specific Errors and Fixes

Generic MySQL, Docker, Go module, and port errors are already documented in the previous dependency files. This table focuses only on product moderation.

| Error/symptom | Likely cause | Fix |
|---|---|---|
| `UNAVAILABLE: Product service is temporarily unavailable` | Product Service or stub is not running at `CMS_PRODUCT_SERVICE_BASE_URL`. | Start Product Service/stub or update `CMS_PRODUCT_SERVICE_BASE_URL`. |
| Startup fails with `product service base url...` | `CMS_PRODUCT_SERVICE_BASE_URL` is invalid, missing scheme, or missing host. | Use a valid URL like `http://localhost:8080`. |
| `401 INTERNAL_AUTH_REQUIRED` | Missing/wrong `X-Internal-Token` when `CMS_INTERNAL_AUTH_TOKEN` is set. | Send header matching `CMS_INTERNAL_AUTH_HEADER` and `CMS_INTERNAL_AUTH_TOKEN`. |
| `403 PERMISSION_DENIED` on seller action | Missing product permission, wrong role, inactive staff header, or seller mismatch. | Use role `seller`, `seller_manager`, or `seller_catalog_editor`; set matching `X-Seller-ID`; use `X-Staff-Status: active`. |
| `403 PERMISSION_DENIED` on admin review action | Actor lacks `admin:catalog:moderate`. | Use role `catalog_admin` or `superadmin`. |
| `422 VALIDATION_FAILED` on submit review | Product Service response is missing required product fields. | Fix product title, description, category, brand/generic flag, images, and variants. |
| `422 invalid moderation decision` | Decision body is not `approved` or `rejected`. | Send `{"decision":"approved"}` or `{"decision":"rejected","reason":"..."}`. |
| `422 rejection reason required` | Reject decision missing reason. | Add clear rejection reason. |
| `412 FAILED_PRECONDITION` | Invalid product state transition or review already decided. | Check current product status and review status before retrying. |
| Duplicate active review error | Product already has a `submitted` review. | Reuse existing review or complete/cancel it before new submission. |
| `Table 'cms_db.product_moderation_reviews' doesn't exist` | Migration `002` or full migrations were not applied. | Run all service migrations in numeric order. |
| Audit write failed | `cms_audit_logs` missing/misconfigured or DB write failed. | Run migration `001` and `007`, verify MySQL access. |

## 13. Security and Best Practices

Shared security guidance is reused from:

```md
Refer:
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`

Sections:
`12. Security and Configuration Audit`
`13. Best Practices`
```

Task-specific practices:

| Practice | Why |
|---|---|
| Do not query Product Service database directly | Service boundary rule: product catalog source of truth stays in Product Service. |
| Keep review records in `cms_db` only | Moderation audit/workflow data belongs to this service. |
| Use strong internal tokens | Moderation routes can approve/reject/publish products, so they must be protected. |
| Trust actor headers only from API Gateway | Direct public access could fake roles/seller id. |
| Always provide rejection/force-unpublish reason | Helps seller correction, audit, and support debugging. |
| Keep audit writes enabled | Approval/rejection history is sensitive and should be traceable. |
| Store timestamps in UTC | Review queue and decision timelines stay consistent. |
| Do not make Redis/cache the source of truth | Review decision state must be durable in MySQL. |

## 14. Missing or Misconfigured Things to Watch

| Finding | Impact | Suggested fix |
|---|---|---|
| No runnable Product Service implementation found in current workspace | Full moderation flow cannot be exercised end-to-end locally. | Add/run Product Service or use a local stub for the expected internal endpoints. |
| No compose stack wires this service plus Product Service | Beginners must start dependencies manually. | Add compose later with MySQL healthcheck and Product Service networking. |
| No migration tracking tool is wired into service startup | Manual migration order can be skipped or repeated. | Add migration runner/deployment migration job later. |
| Product events/search indexing are design references only | Developers may assume Kafka/RabbitMQ/Typesense are required now. | Do not start those services for this task until implementation adds clients/consumers. |
| Product create/update draft routes are not implemented in this service | Submit/publish routes assume product already exists in Product Service. | Implement Product Service/Gateway routes in the appropriate service/task. |

## 15. References to Previous Dependency Files

| Previous dependency file | Section/topic | Reason reused |
|---|---|---|
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `2. Tech Stack Analysis` | Same Go, HTTP, gRPC, MySQL, Product Service, and Gateway context. |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `3. Go Dependency System` | No new Go module or package for `TASK_FILE_NAME`. |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `4. Database Analysis` | MySQL install, Docker, user creation, and full migration commands already exist. |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `5. Environment Variables` | Complete `.env` is already documented; this task only verifies Product Service/internal auth values. |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `6. External Services Analysis` | Product Service and API Gateway/Auth concepts are already introduced. |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `7. Ports and Networking` | No new ports are introduced. |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `8. Docker and DevOps Setup` | No new Docker service is introduced. |
| `TaskImplementation/${SERVICE_NAME}/task2_Dependency.md` | `5. Database Setup` | MySQL 8+, `cms_db`, migration order, and schema ownership are reused. |

## 16. Final Checklist

- [ ] Previous dependency files checked first.
- [ ] Shared Go/MySQL/Docker/env setup followed from previous dependency files.
- [ ] No duplicate installation instructions added here.
- [ ] MySQL is running and reachable.
- [ ] All service migrations applied in numeric order.
- [ ] `product_moderation_reviews` table exists.
- [ ] Table indexes include `uk_product_moderation_review_id` and `uk_product_moderation_active_submitted`.
- [ ] `.env` exists at `backend/services/cms-service/.env`.
- [ ] `.env` is sourced before `go run`.
- [ ] `CMS_PRODUCT_SERVICE_BASE_URL` is a valid `http` or `https` URL.
- [ ] `CMS_INTERNAL_AUTH_TOKEN` and request `X-Internal-Token` match.
- [ ] Service starts and logs `cms.mysql.connected`.
- [ ] `/healthz` returns success.
- [ ] Admin review-list smoke test returns `reviews` JSON.
- [ ] Real submit/approve/publish tests are run only after Product Service or a stub is available.
- [ ] Redis/Kafka/RabbitMQ/Typesense are not added for this task.

Final simple summary: `TASK_FILE_NAME` reuses the existing Go, MySQL, Docker, and `.env` setup. The only task-specific setup focus is the MySQL `product_moderation_reviews` table plus Product Service connectivity for actual moderation actions.
