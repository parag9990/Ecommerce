# 📊 CMS Service - Task 6: Seller Analytics APIs

![Task](https://img.shields.io/badge/Task-CMS%20Service%20%236-blue)
![Priority](https://img.shields.io/badge/Priority-P2-yellow)
![Dependency](https://img.shields.io/badge/Dependency-Order%20%2B%20Payment-purple)
![Feature](https://img.shields.io/badge/Feature-Seller%20Analytics-2ea44f)
![API](https://img.shields.io/badge/API-GetSellerAnalytics-orange)
![Status](https://img.shields.io/badge/Status-Documented-success)
![Guide](https://img.shields.io/badge/Guide-Hinglish-ff69b4)

## 📌 Task Summary

| Field | Detail |
|---|---|
| Task Name | Seller analytics APIs |
| Source | `docs/01-micro-tasks.md` -> `CMS Service` -> Task 6 |
| Goal | Revenue, orders, conversion, top products metrics expose karna |
| Dependencies | `Order Service` and `Payment Service` |
| Priority | `P2` |
| Main API | `GET /api/v1/seller/dashboard/summary` -> `CMSService.GetSellerAnalytics` |
| Permission | `cms:analytics:read` |
| Scope | Analytics API contract, metric formulas, aggregate read model design, service flow, SQL examples, Go usecase examples, cache/security/testing plan |
| Not Included | Actual backend code, SQL migration files, API spec edits, frontend dashboard UI, real event consumers, Task 7 full CMS gRPC implementation, Task 8 audit log implementation |

> **Simple Hinglish goal:** CMS Service me seller dashboard ke liye analytics API design karni hai jisse seller apna revenue, paid orders, conversion rate, aur top products dekh sake. Raw Order/Payment DB ko directly query nahi karna; CMS ko clean aggregate/read-model ke through metrics serve karne chahiye.

---

## ✅ Final Output Created

```text
TaskImplementation/
└── CMS Service/
    ├── task1.md
    ├── task2.md
    ├── task3.md
    ├── task4.md
    ├── task5.md
    └── task6.md
```

### Why this structure?

- `TaskImplementation/` existing project task documentation folder hai.
- `CMS Service/` folder already present tha, isliye usko preserve kiya gaya.
- `task6.md` sirf **CMS Service - Task 6** ka seller analytics API implementation guide document karta hai.
- Existing backend code, migrations, API contracts, frontend files, ya previous task docs modify nahi kiye gaye.

---

## 🧭 Requirement Source Mapping

| Document | Kya use kiya gaya |
|---|---|
| `docs/01-micro-tasks.md` | CMS Service Task 6 ka exact scope: `Seller analytics APIs` |
| `docs/03-folder-structure.md` | Future file placement: `backend/services/cms-service/internal/usecase/seller_analytics.go` |
| `docs/04-microservice-design.md` | CMS responsibility: seller analytics can query read model or pre-aggregated data |
| `docs/05-database-design.md` | Important rule: analytics should use aggregates, not scan raw orders |
| `docs/09-cms-superadmin.md` | Seller analytics metrics: revenue, orders, conversion, top products |
| `docs/12-logging-monitoring-scalability.md` | Observability, caching, and analytics failure isolation guidance |
| `api/master-api.json` | Existing route: `GET /api/v1/seller/dashboard/summary`, gRPC: `CMSService.GetSellerAnalytics` |
| `database/draw.sql` | Order/payment table references for future aggregate event payload understanding |
| `TaskImplementation/CMS Service/task1.md` | Permission reference: `cms:analytics:read` |
| `TaskImplementation/CMS Service/task2.md` | CMS uses MySQL and owns CMS DB only |
| `TaskImplementation/CMS Service/task5.md` | Campaign/coupon analytics are related but not part of Task 6 MVP |

---

## 🚦 Scope Boundary

### ✅ In Scope

- Seller analytics API design.
- Revenue metric definition.
- Orders metric definition.
- Conversion rate metric definition.
- Top products metric definition.
- Date range and currency handling.
- Seller permission and seller ownership rules.
- Aggregate/read-model table blueprint.
- SQL query examples.
- Go domain/usecase/repository code examples.
- REST and gRPC contract reference.
- Cache strategy.
- Observability metrics.
- Test plan.
- Mermaid architecture, sequence, and data-flow diagrams.

### 🚫 Out of Scope

- Actual `backend/services/cms-service/` code create karna.
- Actual MySQL migration file create karna.
- `api/master-api.json` update karna.
- Seller Dashboard React charts banana.
- Session Analytics Dashboard implement karna.
- Full CMS gRPC/proto implementation karna. Ye Task 7 me aayega.
- Seller dashboard action audit logs implement karna. Ye Task 8 me aayega.
- Campaign performance analytics, coupon usage analytics, refund analytics dashboard detail.

> 🔴 **Reason:** User ne required output me folder structure aur `task6.md` content manga hai. Isliye current implementation documentation-level guide hai, production code nahi.

---

## 🪜 Step-by-Step Implementation

## Step 1: Task requirement identify kiya

`docs/01-micro-tasks.md` me CMS Service ka Task 6 ye hai:

| S.No | Task Name | Detail | Dependencies |
|---:|---|---|---|
| 6 | Seller analytics APIs | Revenue, orders, conversion, top products metrics expose karo | Order and Payment |

**Implementation decision:**

- Analytics API ka owner **CMS Service** hoga, kyunki seller dashboard CMS domain ka part hai.
- Raw order/payment facts ka source of truth **Order Service** and **Payment Service** rahenge.
- CMS raw Order DB ya Payment DB ko directly read nahi karega.
- CMS analytics endpoint aggregate/read-model se data serve karega.
- Seller identity request ke auth context se aayegi, query/body se nahi.

> 🟢 **Why:** Seller dashboard ko fast summary chahiye. Agar har request par Order/Payment raw tables scan karenge to dashboard slow hoga and service ownership rule break hoga.

---

## Step 2: MVP metrics final kiye

Task 6 ke liye four core metrics expose karne hain:

| Metric | Meaning | Main Source |
|---|---|---|
| Revenue | Seller ke paid/captured orders se earned amount | Order + Payment aggregate |
| Orders | Seller ke paid orders ka count | Order aggregate |
| Conversion Rate | Seller traffic/view sessions me se paid order tak pahunchne ka rate | Session aggregate + Order aggregate |
| Top Products | Seller ke best-performing products by revenue/units | Order item aggregate |

### MVP response shape

Existing `api/master-api.json` me `SellerAnalyticsResponse` ye fields define karta hai:

```json
{
  "revenue": {
    "currency": "INR",
    "amount": 1250000
  },
  "orders": 84,
  "conversion_rate": 3.42,
  "top_products": [
    {
      "product_id": "prod_101",
      "title": "Cotton T-Shirt",
      "units_sold": 156,
      "revenue": {
        "currency": "INR",
        "amount": 312000
      }
    }
  ]
}
```

> 🔵 **Money rule:** Amounts minor units me store/return honge. Example: `INR 12,500.00` ko API me `1250000` paise ke form me return kar sakte hain.

---

## Step 3: Revenue metric define kiya

Revenue seller dashboard ka primary number hai, isliye iski definition clear honi chahiye.

### Recommended formula

```text
seller_revenue =
  sum(successfully_paid_seller_item_total)
  - sum(successfully_refunded_seller_item_amount)
```

### Revenue calculation fields

| Field | Source | Meaning |
|---|---|---|
| `order_items.seller_id` | Order Service | Item kis seller ka hai |
| `order_items.total_amount` | Order Service | Seller item ka final line amount |
| `orders.status = paid` | Order Service | Order successfully paid hai |
| `payments.status = captured` | Payment Service | Payment provider ne amount capture kar liya |
| `refunds.status = succeeded` | Payment Service | Refund successful hai |

### MVP simplification

Task 6 MVP me API `revenue` ko paid/captured seller item total ke basis par expose karegi. Refund adjustment aggregate table me ready ho to subtract hoga; warna refund analytics future extension me separate field ban sakta hai.

```text
MVP revenue = paid captured seller item total
Future net revenue = paid captured seller item total - succeeded refund allocation
```

> 🟡 **Why not raw payment amount?** Payment amount whole order ka hota hai. Marketplace order me multiple sellers ho sakte hain, isliye seller revenue item-level split se calculate hoga.

---

## Step 4: Orders metric define kiya

Seller ke orders ka count duplicate nahi hona chahiye.

### Recommended formula

```text
orders = count(distinct order_id)
where order has at least one item for seller
and order status is paid/captured
```

### Example

| Order | Seller Items | Count for Seller A? | Count for Seller B? |
|---|---:|---:|---:|
| `ord_1` | A: 2 items | ✅ 1 | ❌ 0 |
| `ord_2` | A: 1 item, B: 1 item | ✅ 1 | ✅ 1 |
| `ord_3` | B: 3 items | ❌ 0 | ✅ 1 |

> 🔵 **Important:** Same seller ke multiple items ek order me hon to bhi order count `1` hi hoga.

---

## Step 5: Conversion rate metric define kiya

Conversion rate ke liye traffic denominator clear hona important hai.

### MVP formula

```text
conversion_rate =
  paid_order_sessions / seller_product_view_sessions * 100
```

### Field meaning

| Field | Meaning |
|---|---|
| `seller_product_view_sessions` | Unique/aggregated sessions jinhone seller ke products dekhe |
| `paid_order_sessions` | Sessions jinke through seller ke item ka paid order bana |
| `conversion_rate` | Percentage value, for example `3.42` |

### Edge cases

| Case | Result |
|---|---|
| No product views | `0` conversion rate |
| Views hain but paid order nahi | `0` conversion rate |
| Orders hain but session data delayed | Return available aggregate and mark freshness internally |

> 🟡 **Note:** Session Service independent analytics owner hai. CMS conversion metric ke liye Session Service se aggregate event/read-model use karega. Direct Session DB read nahi karega.

---

## Step 6: Top products metric define kiya

Top products seller ko batate hain ki kaunse products revenue/orders drive kar rahe hain.

### Default sort

```text
top_products sorted by revenue_amount desc
```

### Product row shape

| Field | Meaning |
|---|---|
| `product_id` | Product Service ka stable product id |
| `title` | Order-time/product snapshot title |
| `units_sold` | Paid units sold |
| `orders` | Distinct paid orders containing product |
| `revenue` | Product-level paid seller item revenue |

### Example

```json
{
  "product_id": "prod_shoe_001",
  "title": "Running Shoes",
  "units_sold": 42,
  "orders": 36,
  "revenue": {
    "currency": "INR",
    "amount": 840000
  }
}
```

> 🟢 **Why title snapshot?** Product title future me change ho sakta hai. Historical analytics stable rakhne ke liye order item/product aggregate me title snapshot useful hota hai.

---

## Step 7: Service ownership boundary fix kiya

System architecture ka golden rule:

```text
Har service apni database own karegi.
Dusri service ke DB ko directly read/write nahi karna.
```

### Ownership table

| Data | Owner Service | CMS ka access pattern |
|---|---|---|
| Orders, order items | Order Service | Event projection or gRPC aggregate |
| Payments, refunds | Payment Service | Event projection or gRPC aggregate |
| Product title/status | Product Service | Snapshot/event/gRPC only |
| Sessions/conversion | Session Service | Aggregate event or gRPC report |
| Seller dashboard analytics read model | CMS Service | Own MySQL aggregate tables |

> 🔴 **Not allowed:** CMS Service -> `order_db` direct SQL query.  
> 🟢 **Allowed:** Order Service publishes `OrderPaid` event, CMS analytics projection consumes and updates CMS aggregate table.

---

## Step 8: High-level architecture design kiya

```mermaid
flowchart LR
    Seller[Seller Dashboard] --> GW[API Gateway]
    GW --> CMS[CMS Service]
    CMS --> Cache[(Redis Cache - optional)]
    CMS --> CMSDB[(MySQL cms_db aggregate tables)]

    Order[Order Service] -->|OrderPaid / OrderCancelled events| Bus[(Kafka/RabbitMQ)]
    Payment[Payment Service] -->|PaymentCaptured / RefundSucceeded events| Bus
    Session[Session Service] -->|Seller funnel aggregate events| Bus
    Product[Product Service] -->|Product snapshot updates| Bus

    Bus --> Projection[CMS Analytics Projector]
    Projection --> CMSDB
```

### Explanation

1. Seller dashboard API Gateway ko request bhejta hai.
2. Gateway auth validate karke request CMS Service ko forward karta hai.
3. CMS Service `cms:analytics:read` permission check karta hai.
4. CMS Redis cache check kar sakta hai.
5. Cache miss par CMS MySQL aggregate tables se metrics read karta hai.
6. Order/Payment/Session/Product events background me aggregate tables update karte hain.

---

## Step 9: API endpoint choose kiya

Existing API master file me route already present hai:

| Route | Method | Service | gRPC | Auth |
|---|---|---|---|---|
| `/api/v1/seller/dashboard/summary` | `GET` | `cms-service` | `CMSService.GetSellerAnalytics` | seller |

### Request query params

```http
GET /api/v1/seller/dashboard/summary?from=2026-05-01&to=2026-05-24&currency=INR&top_products_limit=5
Authorization: Bearer <seller-token>
```

### Request rules

| Param | Required | Default | Rule |
|---|---:|---|---|
| `from` | No | Today - 30 days | ISO date or datetime |
| `to` | No | Today | ISO date or datetime |
| `currency` | No | Seller default currency | 3-letter currency code |
| `top_products_limit` | No | `5` | Min `1`, max `20` |

> 🔵 **Seller id rule:** Client se `seller_id` accept nahi karna. Seller id authenticated token/context se derive hoga.

---

## Step 10: API response contract design kiya

### Successful response

```json
{
  "revenue": {
    "currency": "INR",
    "amount": 1250000
  },
  "orders": 84,
  "conversion_rate": 3.42,
  "top_products": [
    {
      "product_id": "prod_101",
      "title": "Cotton T-Shirt",
      "units_sold": 156,
      "orders": 72,
      "revenue": {
        "currency": "INR",
        "amount": 312000
      }
    },
    {
      "product_id": "prod_205",
      "title": "Slim Fit Jeans",
      "units_sold": 91,
      "orders": 64,
      "revenue": {
        "currency": "INR",
        "amount": 455000
      }
    }
  ]
}
```

### Empty data response

```json
{
  "revenue": {
    "currency": "INR",
    "amount": 0
  },
  "orders": 0,
  "conversion_rate": 0,
  "top_products": []
}
```

> 🟢 **Why empty response 200?** Seller ka new store ho sakta hai jahan data abhi nahi hai. Empty analytics error nahi hai.

---

## Step 11: Permission and access control add kiya

Task 1 me permission already define hai:

```text
cms:analytics:read
```

### Access rules

| Rule | Behavior |
|---|---|
| Seller token required | Anonymous/buyer request reject |
| `cms:analytics:read` required | Staff role ke basis par allow/deny |
| Seller scope mandatory | Data only authenticated seller ka |
| No user PII in response | Buyer names/emails/phones expose nahi honge |
| Date range bounded | Huge scans avoid honge |

### Role permission reference

| Role | Can read analytics? |
|---|---:|
| `seller` | ✅ |
| `seller_manager` | ✅ |
| `seller_catalog_editor` | ✅ |
| `seller_order_manager` | ✅ |

> 🟡 **Why order manager ko analytics read?** Task 1 role matrix me `cms:analytics:read` sab seller workspace roles ko diya gaya hai. Production me finance/revenue metrics ke liye stricter role split future me add ho sakta hai.

---

## Step 12: Date range validation design kiya

Analytics APIs me date range guard zaruri hai.

### Validation rules

| Check | Rule |
|---|---|
| `from` format | Valid date/datetime |
| `to` format | Valid date/datetime |
| Range order | `from <= to` |
| Max range | Suggested max `366` days |
| Future date | Allow only up to current day, or clamp |
| Timezone | Store/query UTC; display timezone frontend handle karega |

### Error examples

| Case | Error Code | HTTP |
|---|---|---:|
| Invalid date | `invalid_date_range` | `400` |
| Range too large | `date_range_too_large` | `400` |
| Missing seller context | `seller_context_required` | `403` |
| Permission missing | `permission_denied` | `403` |

---

## Step 13: Aggregate read model choose kiya

`docs/05-database-design.md` me clear note hai:

```text
Analytics should use aggregates, not scan raw orders.
```

Isliye CMS Service ke liye read model aggregate tables recommended hain.

### Future aggregate table group

```text
cms_db
├── seller_analytics_daily
├── seller_product_analytics_daily
├── seller_conversion_daily
└── seller_analytics_event_dedupe
```

### Why aggregate tables?

| Reason | Benefit |
|---|---|
| Fast dashboard load | Single seller/date range aggregation quick |
| Service ownership safe | Raw Order/Payment DB access avoid |
| Event-driven updates | Checkout path block nahi hota |
| Historical snapshots | Product title/revenue history stable |
| Cache-friendly | Same date range repeatedly read hota hai |

---

## Step 14: Aggregate table blueprint banaya

> 🔵 **Note:** Ye SQL migration example future implementation ke liye hai. Current task me migration file create nahi ki gayi.

### `seller_analytics_daily`

```sql
CREATE TABLE IF NOT EXISTS seller_analytics_daily (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  seller_id VARCHAR(64) NOT NULL,
  metric_date DATE NOT NULL,
  currency CHAR(3) NOT NULL,
  gross_revenue_amount BIGINT NOT NULL DEFAULT 0,
  refund_amount BIGINT NOT NULL DEFAULT 0,
  net_revenue_amount BIGINT NOT NULL DEFAULT 0,
  paid_order_count BIGINT NOT NULL DEFAULT 0,
  paid_item_count BIGINT NOT NULL DEFAULT 0,
  cancelled_order_count BIGINT NOT NULL DEFAULT 0,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_seller_analytics_day_currency (seller_id, metric_date, currency),
  KEY idx_seller_analytics_seller_date (seller_id, metric_date)
) ENGINE=InnoDB;
```

### `seller_product_analytics_daily`

```sql
CREATE TABLE IF NOT EXISTS seller_product_analytics_daily (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  seller_id VARCHAR(64) NOT NULL,
  product_id VARCHAR(64) NOT NULL,
  metric_date DATE NOT NULL,
  currency CHAR(3) NOT NULL,
  title_snapshot VARCHAR(512) NULL,
  units_sold BIGINT NOT NULL DEFAULT 0,
  order_count BIGINT NOT NULL DEFAULT 0,
  revenue_amount BIGINT NOT NULL DEFAULT 0,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_seller_product_day_currency (seller_id, product_id, metric_date, currency),
  KEY idx_seller_product_top_revenue (seller_id, metric_date, revenue_amount),
  KEY idx_seller_product_top_units (seller_id, metric_date, units_sold)
) ENGINE=InnoDB;
```

### `seller_conversion_daily`

```sql
CREATE TABLE IF NOT EXISTS seller_conversion_daily (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  seller_id VARCHAR(64) NOT NULL,
  metric_date DATE NOT NULL,
  product_view_sessions BIGINT NOT NULL DEFAULT 0,
  add_to_cart_sessions BIGINT NOT NULL DEFAULT 0,
  checkout_started_sessions BIGINT NOT NULL DEFAULT 0,
  paid_order_sessions BIGINT NOT NULL DEFAULT 0,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_seller_conversion_day (seller_id, metric_date),
  KEY idx_seller_conversion_seller_date (seller_id, metric_date)
) ENGINE=InnoDB;
```

### `seller_analytics_event_dedupe`

```sql
CREATE TABLE IF NOT EXISTS seller_analytics_event_dedupe (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  event_id VARCHAR(128) NOT NULL,
  event_type VARCHAR(128) NOT NULL,
  source_service VARCHAR(64) NOT NULL,
  processed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_seller_analytics_event_id (event_id)
) ENGINE=InnoDB;
```

> 🟢 **Why dedupe table?** Message queues at-least-once delivery kar sakte hain. Same `OrderPaid` event dobara aaye to revenue double count nahi hona chahiye.

---

## Step 15: Event projection flow design kiya

Analytics read model async events se update hoga.

```mermaid
flowchart TD
    A[OrderPaid event] --> B[CMS Analytics Consumer]
    B --> C{event_id processed?}
    C -- yes --> D[Skip duplicate]
    C -- no --> E[Start DB transaction]
    E --> F[Insert event_id into dedupe table]
    F --> G[Upsert seller_analytics_daily]
    G --> H[Upsert seller_product_analytics_daily]
    H --> I[Commit transaction]

    J[Payment RefundSucceeded event] --> K[Update refund/net revenue aggregate]
    L[Session seller funnel aggregate] --> M[Upsert seller_conversion_daily]
```

### Event types needed

| Event | Source | Use |
|---|---|---|
| `OrderPaid` | Order Service | Order count, item units, product revenue |
| `OrderCancelled` | Order Service | Cancellation counters |
| `PaymentCaptured` | Payment Service | Revenue recognition confirmation |
| `RefundSucceeded` | Payment Service | Refund adjustment |
| `SellerFunnelDailyAggregated` | Session Service | Conversion denominator/numerator |
| `ProductUpdated` | Product Service | Product title snapshot refresh, optional |

> 🟡 **Checkout rule:** Analytics projection failure checkout ko block nahi karegi. Event retry/DLQ se projection later recover ho sakti hai.

---

## Step 16: Request sequence design kiya

```mermaid
sequenceDiagram
    participant Seller as Seller Dashboard
    participant GW as API Gateway
    participant CMS as CMS Service
    participant Cache as Redis Cache
    participant DB as CMS MySQL Aggregates

    Seller->>GW: GET /api/v1/seller/dashboard/summary?from&to
    GW->>GW: Verify JWT + inject auth context
    GW->>CMS: GetSellerAnalytics(auth_context, date_range)
    CMS->>CMS: Check cms:analytics:read
    CMS->>CMS: Resolve seller_id from auth context
    CMS->>Cache: Get analytics cache key
    alt Cache hit
        Cache-->>CMS: Cached response
    else Cache miss
        CMS->>DB: Query summary + conversion + top products
        DB-->>CMS: Aggregated rows
        CMS->>CMS: Compute conversion and response
        CMS->>Cache: Store short TTL response
    end
    CMS-->>GW: SellerAnalyticsResponse
    GW-->>Seller: JSON response
```

---

## Step 17: Repository SQL examples likhe

### Summary query

```sql
SELECT
  currency,
  COALESCE(SUM(net_revenue_amount), 0) AS revenue_amount,
  COALESCE(SUM(paid_order_count), 0) AS paid_orders,
  COALESCE(SUM(paid_item_count), 0) AS paid_items
FROM seller_analytics_daily
WHERE seller_id = ?
  AND metric_date >= ?
  AND metric_date <= ?
  AND currency = ?
GROUP BY currency;
```

### Conversion query

```sql
SELECT
  COALESCE(SUM(product_view_sessions), 0) AS product_view_sessions,
  COALESCE(SUM(paid_order_sessions), 0) AS paid_order_sessions
FROM seller_conversion_daily
WHERE seller_id = ?
  AND metric_date >= ?
  AND metric_date <= ?;
```

### Conversion formula in service

```text
if product_view_sessions == 0:
    conversion_rate = 0
else:
    conversion_rate = paid_order_sessions * 100 / product_view_sessions
```

### Top products query

```sql
SELECT
  product_id,
  COALESCE(MAX(title_snapshot), '') AS title,
  COALESCE(SUM(units_sold), 0) AS units_sold,
  COALESCE(SUM(order_count), 0) AS orders,
  COALESCE(SUM(revenue_amount), 0) AS revenue_amount
FROM seller_product_analytics_daily
WHERE seller_id = ?
  AND metric_date >= ?
  AND metric_date <= ?
  AND currency = ?
GROUP BY product_id
ORDER BY revenue_amount DESC, units_sold DESC
LIMIT ?;
```

> 🔵 **Why query aggregate table only?** Ye dashboard read path ko predictable rakhta hai. Raw `orders` and `order_items` scan karna future high traffic me expensive hoga.

---

## Step 18: Domain model example banaya

> 🔵 **Note:** Ye future Go implementation example hai. Current task me Go file create nahi ki gayi.

```go
package domain

import "time"

type Money struct {
    Currency string
    Amount   int64
}

type AnalyticsQuery struct {
    SellerID         string
    From             time.Time
    To               time.Time
    Currency         string
    TopProductsLimit int
}

type SellerAnalytics struct {
    Revenue       Money
    Orders        int64
    ConversionRate float64
    TopProducts   []TopProductMetric
}

type TopProductMetric struct {
    ProductID string
    Title     string
    UnitsSold int64
    Orders    int64
    Revenue   Money
}
```

### Why this model?

| Type | Purpose |
|---|---|
| `Money` | Currency + minor amount together rakhta hai |
| `AnalyticsQuery` | Seller/date/currency filters one object me |
| `SellerAnalytics` | API response ka core domain shape |
| `TopProductMetric` | Top products table/cards ke liye row model |

---

## Step 19: Repository interface define kiya

```go
package usecase

import (
    "context"

    "ecommerce/cms-service/internal/domain"
)

type SellerAnalyticsRepository interface {
    GetSellerSummary(ctx context.Context, query domain.AnalyticsQuery) (SellerSummaryRow, error)
    GetSellerConversion(ctx context.Context, query domain.AnalyticsQuery) (SellerConversionRow, error)
    ListTopProducts(ctx context.Context, query domain.AnalyticsQuery) ([]domain.TopProductMetric, error)
}

type SellerSummaryRow struct {
    Currency      string
    RevenueAmount int64
    PaidOrders    int64
}

type SellerConversionRow struct {
    ProductViewSessions int64
    PaidOrderSessions   int64
}
```

### Why interface?

- Usecase DB driver se independent rahega.
- Unit tests fake repository se easy ho jayenge.
- Future me MySQL read replica ya cached repository swap kar sakte hain.

---

## Step 20: Usecase flow implement kiya

```go
package usecase

import (
    "context"
    "errors"
    "time"

    "ecommerce/cms-service/internal/domain"
)

var (
    ErrPermissionDenied = errors.New("permission denied")
    ErrInvalidDateRange = errors.New("invalid date range")
)

type AuthContext struct {
    UserID      string
    SellerID    string
    Permissions map[string]bool
}

func (a AuthContext) HasPermission(permission string) bool {
    return a.Permissions[permission]
}

type SellerAnalyticsUsecase struct {
    repo SellerAnalyticsRepository
    now  func() time.Time
}

func (u *SellerAnalyticsUsecase) GetSellerAnalytics(
    ctx context.Context,
    auth AuthContext,
    query domain.AnalyticsQuery,
) (domain.SellerAnalytics, error) {
    if !auth.HasPermission("cms:analytics:read") {
        return domain.SellerAnalytics{}, ErrPermissionDenied
    }

    if auth.SellerID == "" {
        return domain.SellerAnalytics{}, ErrPermissionDenied
    }

    query.SellerID = auth.SellerID

    if err := validateAnalyticsRange(query.From, query.To, u.now()); err != nil {
        return domain.SellerAnalytics{}, err
    }

    if query.TopProductsLimit <= 0 {
        query.TopProductsLimit = 5
    }
    if query.TopProductsLimit > 20 {
        query.TopProductsLimit = 20
    }

    summary, err := u.repo.GetSellerSummary(ctx, query)
    if err != nil {
        return domain.SellerAnalytics{}, err
    }

    conversion, err := u.repo.GetSellerConversion(ctx, query)
    if err != nil {
        return domain.SellerAnalytics{}, err
    }

    topProducts, err := u.repo.ListTopProducts(ctx, query)
    if err != nil {
        return domain.SellerAnalytics{}, err
    }

    return domain.SellerAnalytics{
        Revenue: domain.Money{
            Currency: query.Currency,
            Amount:   summary.RevenueAmount,
        },
        Orders:         summary.PaidOrders,
        ConversionRate: calculateConversionRate(conversion),
        TopProducts:    topProducts,
    }, nil
}

func calculateConversionRate(row SellerConversionRow) float64 {
    if row.ProductViewSessions <= 0 {
        return 0
    }
    return float64(row.PaidOrderSessions) * 100 / float64(row.ProductViewSessions)
}
```

### Date validation helper

```go
func validateAnalyticsRange(from, to, now time.Time) error {
    if from.IsZero() || to.IsZero() {
        return ErrInvalidDateRange
    }
    if from.After(to) {
        return ErrInvalidDateRange
    }
    if to.After(now.Add(24 * time.Hour)) {
        return ErrInvalidDateRange
    }
    if to.Sub(from) > 366*24*time.Hour {
        return ErrInvalidDateRange
    }
    return nil
}
```

> 🟢 **Why seller id overwrite?** Client request me seller id spoof ho sakta hai. Usecase auth context ka `SellerID` final source maanta hai.

---

## Step 21: gRPC contract reference diya

`api/master-api.json` already `CMSService.GetSellerAnalytics` define karta hai.

### Proto-style reference

```proto
service CMSService {
  rpc GetSellerAnalytics(DateRangeRequest) returns (SellerAnalyticsResponse);
}

message DateRangeRequest {
  string from = 1;
  string to = 2;
  string currency = 3;
  int32 top_products_limit = 4;
}

message SellerAnalyticsResponse {
  Money revenue = 1;
  int64 orders = 2;
  double conversion_rate = 3;
  repeated TopProductMetric top_products = 4;
}

message TopProductMetric {
  string product_id = 1;
  string title = 2;
  int64 units_sold = 3;
  int64 orders = 4;
  Money revenue = 5;
}
```

> 🟡 **Important:** Full proto file generation and CMS gRPC server implementation Task 7 ka scope hai. Task 6 sirf analytics method behavior and data contract define karta hai.

---

## Step 22: REST to gRPC mapping samjha

```mermaid
flowchart LR
    REST[GET /api/v1/seller/dashboard/summary] --> Gateway[API Gateway]
    Gateway -->|DateRangeRequest + auth metadata| GRPC[CMSService.GetSellerAnalytics]
    GRPC --> Usecase[SellerAnalyticsUsecase]
    Usecase --> Repo[Analytics Repository]
    Repo --> DB[(CMS aggregate tables)]
```

### Mapping table

| REST | gRPC |
|---|---|
| Query `from` | `DateRangeRequest.from` |
| Query `to` | `DateRangeRequest.to` |
| Query `currency` | `DateRangeRequest.currency` |
| Query `top_products_limit` | `DateRangeRequest.top_products_limit` |
| Auth token seller id | gRPC metadata/auth context |
| JSON response | `SellerAnalyticsResponse` |

---

## Step 23: Cache strategy add kiya

Seller dashboard repeatedly same range load karega, isliye short TTL cache useful hai.

### Cache key

```text
cms:seller_analytics:{seller_id}:{from}:{to}:{currency}:{top_products_limit}
```

### TTL recommendation

| Range | TTL |
|---|---:|
| Today | 30-60 seconds |
| Last 7 days | 2 minutes |
| Last 30 days | 5 minutes |
| Historical closed range | 15-60 minutes |

### Cache invalidation

| Event | Action |
|---|---|
| `OrderPaid` | Delete seller analytics keys for affected seller |
| `RefundSucceeded` | Delete seller analytics keys for affected seller |
| `SellerFunnelDailyAggregated` | Delete conversion-related seller analytics keys |
| Product title update | Optional top product cache delete |

> 🟡 **MVP option:** Redis cache optional hai. Agar cache unavailable ho, CMS aggregate DB se response serve karega.

---

## Step 24: Error behavior define kiya

### Error response examples

#### Invalid date range

```json
{
  "error": {
    "code": "invalid_date_range",
    "message": "from date must be before to date"
  }
}
```

#### Permission denied

```json
{
  "error": {
    "code": "permission_denied",
    "message": "analytics read permission required"
  }
}
```

### Behavior table

| Scenario | HTTP/gRPC | Response |
|---|---:|---|
| No analytics rows | `200 / OK` | Zero values |
| Invalid date | `400 / InvalidArgument` | `invalid_date_range` |
| Range too large | `400 / InvalidArgument` | `date_range_too_large` |
| Missing seller context | `403 / PermissionDenied` | `seller_context_required` |
| DB timeout | `503 / Unavailable` | `analytics_temporarily_unavailable` |

---

## Step 25: Observability add kiya

Task 6 analytics API ko monitor karna important hai.

### Logs

Structured JSON log fields:

```json
{
  "service": "cms-service",
  "event": "seller_analytics_requested",
  "seller_id_hash": "hash_seller_123",
  "from": "2026-05-01",
  "to": "2026-05-24",
  "currency": "INR",
  "request_id": "req_abc",
  "latency_ms": 42
}
```

### Metrics

| Metric | Type | Meaning |
|---|---|---|
| `cms_seller_analytics_requests_total` | Counter | Total analytics requests |
| `cms_seller_analytics_errors_total` | Counter | Error count by reason |
| `cms_seller_analytics_duration_ms` | Histogram | API latency |
| `cms_seller_analytics_cache_hits_total` | Counter | Redis cache hits |
| `cms_seller_analytics_cache_misses_total` | Counter | Redis cache misses |
| `cms_seller_analytics_projection_lag_seconds` | Gauge | Latest event projection delay |

### Trace span names

```text
CMSService.GetSellerAnalytics
SellerAnalyticsUsecase.GetSellerAnalytics
SellerAnalyticsRepository.GetSellerSummary
SellerAnalyticsRepository.ListTopProducts
```

> 🔴 **Privacy rule:** Logs me product title ok ho sakta hai, but buyer PII, emails, phone, payment provider secrets, JWT, ya raw auth token kabhi log nahi karna.

---

## Step 26: Security checklist banaya

| Check | Required |
|---|---:|
| JWT/Auth context required | ✅ |
| `cms:analytics:read` permission check | ✅ |
| Seller id auth context se derive | ✅ |
| Cross-seller query blocked | ✅ |
| Date range max limit | ✅ |
| No buyer PII in response | ✅ |
| No direct Order/Payment DB access | ✅ |
| Request/response logs redacted | ✅ |
| Cache key seller-scoped | ✅ |
| Cache data does not leak across sellers | ✅ |

---

## Step 27: Clean future folder structure define kiya

Current task me sirf documentation file create hui hai. Future actual implementation ke liye recommended structure:

```text
backend/
└── services/
    └── cms-service/
        ├── cmd/
        │   └── server/
        │       └── main.go
        ├── internal/
        │   ├── domain/
        │   │   ├── seller_analytics.go
        │   │   ├── money.go
        │   │   └── analytics_metric.go
        │   ├── usecase/
        │   │   └── seller_analytics.go
        │   ├── repository/
        │   │   ├── mysql_analytics_repository.go
        │   │   └── redis_analytics_cache.go
        │   ├── events/
        │   │   ├── order_analytics_consumer.go
        │   │   ├── payment_analytics_consumer.go
        │   │   └── session_analytics_consumer.go
        │   └── transport/
        │       └── grpc/
        │           └── analytics_handler.go
        ├── migrations/
        │   ├── 006_create_seller_analytics_tables.up.sql
        │   └── 006_create_seller_analytics_tables.down.sql
        └── deploy/
```

### File responsibility

| File | Responsibility |
|---|---|
| `domain/seller_analytics.go` | Seller analytics response and top product structs |
| `domain/money.go` | Money representation in minor units |
| `usecase/seller_analytics.go` | Permission, date validation, aggregate query orchestration |
| `repository/mysql_analytics_repository.go` | Summary/conversion/top-product SQL queries |
| `repository/redis_analytics_cache.go` | Optional response cache |
| `events/order_analytics_consumer.go` | `OrderPaid` and order lifecycle analytics projection |
| `events/payment_analytics_consumer.go` | `PaymentCaptured` and refund projection |
| `events/session_analytics_consumer.go` | Conversion/funnel aggregate projection |
| `transport/grpc/analytics_handler.go` | `GetSellerAnalytics` request/response mapping |
| `migrations/006_*` | Aggregate table creation |

> 🔵 **Note:** Ye future implementation placement hai. Current Task 6 me ye files create nahi ki gayi.

---

## Step 28: External libraries/tools document kiye

Is documentation-only task me koi package install nahi kiya gaya. Real CMS Service Task 6 implement karte time ye tools useful honge:

| Tool/Library | What it is | Why used | Install / Usage |
|---|---|---|---|
| MySQL 8+ | Relational database | CMS aggregate tables store karne ke liye | Docker/managed MySQL |
| InnoDB | MySQL storage engine | Transactions and row-level upserts ke liye | MySQL ke saath built-in |
| `github.com/go-sql-driver/mysql` | Go MySQL driver | Go CMS service ko MySQL se connect karne ke liye | `go get github.com/go-sql-driver/mysql` |
| `github.com/golang-migrate/migrate/v4` | Migration tool | Analytics aggregate table migrations run karne ke liye | `go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest` |
| `google.golang.org/grpc` | gRPC framework | Gateway -> CMS typed calls ke liye | `go get google.golang.org/grpc` |
| `google.golang.org/protobuf` | Proto runtime | Generated proto messages ke liye | `go get google.golang.org/protobuf` |
| `github.com/redis/go-redis/v9` | Redis client | Optional analytics response cache ke liye | `go get github.com/redis/go-redis/v9` |
| OpenTelemetry | Tracing toolkit | Analytics API traces collect karne ke liye | `go get go.opentelemetry.io/otel` |
| Prometheus client | Metrics library | Request/cache/projection metrics expose karne ke liye | `go get github.com/prometheus/client_golang` |
| Mermaid | Markdown diagram syntax | Architecture/flow diagrams ke liye | GitHub/Markdown viewer me usually built-in |
| Shields.io badges | Visual badges | Document readability ke liye | No install required |

### Local MySQL example

```bash
docker run --name ecommerce-cms-mysql \
  -e MYSQL_ROOT_PASSWORD=root \
  -e MYSQL_DATABASE=cms_db \
  -e MYSQL_USER=cms_user \
  -e MYSQL_PASSWORD=cms_password \
  -p 3306:3306 \
  -d mysql:8
```

### Future migration command example

```bash
migrate -path backend/services/cms-service/migrations \
  -database "mysql://cms_user:cms_password@tcp(localhost:3306)/cms_db" \
  up
```

### Future Go package install example

```bash
go get github.com/go-sql-driver/mysql
go get google.golang.org/grpc
go get google.golang.org/protobuf
go get github.com/redis/go-redis/v9
```

> 🟡 **Note:** Commands reference ke liye hain. Current task me koi package install ya service run nahi kiya gaya.

---

## Step 29: Projection upsert examples diye

### Daily seller summary upsert

```sql
INSERT INTO seller_analytics_daily (
  seller_id,
  metric_date,
  currency,
  gross_revenue_amount,
  net_revenue_amount,
  paid_order_count,
  paid_item_count
) VALUES (?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  gross_revenue_amount = gross_revenue_amount + VALUES(gross_revenue_amount),
  net_revenue_amount = net_revenue_amount + VALUES(net_revenue_amount),
  paid_order_count = paid_order_count + VALUES(paid_order_count),
  paid_item_count = paid_item_count + VALUES(paid_item_count),
  updated_at = CURRENT_TIMESTAMP;
```

### Daily top product upsert

```sql
INSERT INTO seller_product_analytics_daily (
  seller_id,
  product_id,
  metric_date,
  currency,
  title_snapshot,
  units_sold,
  order_count,
  revenue_amount
) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  title_snapshot = VALUES(title_snapshot),
  units_sold = units_sold + VALUES(units_sold),
  order_count = order_count + VALUES(order_count),
  revenue_amount = revenue_amount + VALUES(revenue_amount),
  updated_at = CURRENT_TIMESTAMP;
```

### Daily conversion upsert

```sql
INSERT INTO seller_conversion_daily (
  seller_id,
  metric_date,
  product_view_sessions,
  add_to_cart_sessions,
  checkout_started_sessions,
  paid_order_sessions
) VALUES (?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  product_view_sessions = VALUES(product_view_sessions),
  add_to_cart_sessions = VALUES(add_to_cart_sessions),
  checkout_started_sessions = VALUES(checkout_started_sessions),
  paid_order_sessions = VALUES(paid_order_sessions),
  updated_at = CURRENT_TIMESTAMP;
```

> 🔵 **Difference:** Order/payment events increment counters. Session daily aggregate can replace daily counters because Session Service may publish full daily aggregate snapshot.

---

## Step 30: Analytics freshness define kiya

Analytics data real-time nahi bhi ho sakta. Seller ko near-real-time dashboard enough hota hai.

| Data | Freshness Target |
|---|---:|
| Revenue/orders after payment success | 1-5 minutes |
| Top products | 1-5 minutes |
| Conversion/session aggregates | 5-15 minutes |
| Historical closed days | Stable |

### Freshness handling

```text
If projection lag is low:
    serve normal response
If projection lag is high:
    serve last known aggregate
    log warning metric
If aggregate table unavailable:
    return analytics_temporarily_unavailable
```

> 🟡 **UX note:** Frontend can show "Last updated" later, but Task 6 backend MVP response in `api/master-api.json` does not require this field.

---

## Step 31: Test plan banaya

### Unit tests

| Test | Expected |
|---|---|
| Permission missing | `ErrPermissionDenied` |
| Seller id missing | `ErrPermissionDenied` |
| Invalid date range | `ErrInvalidDateRange` |
| Range over max limit | Validation error |
| Top products limit missing | Default `5` |
| Top products limit too high | Clamp to `20` |
| Product views zero | Conversion `0` |
| Product views positive | Conversion calculated correctly |

### Repository tests

| Test | Expected |
|---|---|
| Summary rows exist | Correct sum returned |
| No summary rows | Zero revenue/orders |
| Multiple days | Date range sum correct |
| Different currency | Only requested currency returned |
| Top products sorted | Revenue desc, units desc |
| Cross seller data | Not returned |

### Projection tests

| Test | Expected |
|---|---|
| `OrderPaid` event | Summary and product aggregates increment |
| Duplicate event id | No double count |
| Multi-seller order | Each seller aggregate updated separately |
| Refund event | Refund/net revenue adjusted |
| Session aggregate event | Conversion row replaced/upserted |

### Security tests

| Test | Expected |
|---|---|
| Seller A asks for Seller B data | Impossible because seller id from auth |
| Buyer token calls endpoint | `403` |
| Staff without analytics permission | `403` |
| Huge date range | `400` |
| Cache key isolation | Seller A cache never returned to Seller B |

---

## Step 32: API examples add kiye

### cURL request

```bash
curl -X GET \
  "https://api.example.com/api/v1/seller/dashboard/summary?from=2026-05-01&to=2026-05-24&currency=INR&top_products_limit=5" \
  -H "Authorization: Bearer <seller_access_token>"
```

### JavaScript frontend call example

```ts
type Money = {
  currency: string;
  amount: number;
};

type TopProductMetric = {
  product_id: string;
  title: string;
  units_sold: number;
  orders: number;
  revenue: Money;
};

type SellerAnalyticsResponse = {
  revenue: Money;
  orders: number;
  conversion_rate: number;
  top_products: TopProductMetric[];
};

export async function getSellerDashboardSummary(params: {
  from: string;
  to: string;
  currency?: string;
  topProductsLimit?: number;
}): Promise<SellerAnalyticsResponse> {
  const search = new URLSearchParams({
    from: params.from,
    to: params.to,
    currency: params.currency ?? "INR",
    top_products_limit: String(params.topProductsLimit ?? 5),
  });

  const response = await fetch(`/api/v1/seller/dashboard/summary?${search}`);

  if (!response.ok) {
    throw new Error("Failed to load seller analytics");
  }

  return response.json();
}
```

> 🔵 **Frontend note:** Amount format frontend locale ke according karega. Backend minor units return karega.

---

## Step 33: Mermaid ER diagram add kiya

```mermaid
erDiagram
    SELLER_ANALYTICS_DAILY {
        bigint id PK
        varchar seller_id
        date metric_date
        char currency
        bigint gross_revenue_amount
        bigint refund_amount
        bigint net_revenue_amount
        bigint paid_order_count
        bigint paid_item_count
    }

    SELLER_PRODUCT_ANALYTICS_DAILY {
        bigint id PK
        varchar seller_id
        varchar product_id
        date metric_date
        char currency
        varchar title_snapshot
        bigint units_sold
        bigint order_count
        bigint revenue_amount
    }

    SELLER_CONVERSION_DAILY {
        bigint id PK
        varchar seller_id
        date metric_date
        bigint product_view_sessions
        bigint add_to_cart_sessions
        bigint checkout_started_sessions
        bigint paid_order_sessions
    }

    SELLER_ANALYTICS_EVENT_DEDUPE {
        bigint id PK
        varchar event_id UK
        varchar event_type
        varchar source_service
        timestamp processed_at
    }

    SELLER_ANALYTICS_DAILY ||--o{ SELLER_PRODUCT_ANALYTICS_DAILY : "same seller/date"
    SELLER_ANALYTICS_DAILY ||--o{ SELLER_CONVERSION_DAILY : "same seller/date"
```

---

## Step 34: Metric formula cheat sheet banaya

| Metric | Formula | Zero Handling |
|---|---|---|
| Revenue | `sum(net_revenue_amount)` | Return `0` |
| Orders | `sum(paid_order_count)` | Return `0` |
| Conversion | `paid_order_sessions * 100 / product_view_sessions` | If views `0`, return `0` |
| Top Products | `group by product_id order by revenue desc limit N` | Return `[]` |

### Example calculation

```text
Seller product view sessions = 2,000
Paid order sessions = 68

conversion_rate = 68 * 100 / 2000
conversion_rate = 3.4
```

---

## Step 35: Performance considerations likhe

| Concern | Design |
|---|---|
| Dashboard latency | Query aggregate tables, not raw orders |
| Large date ranges | Limit to 366 days |
| Hot sellers | Redis cache with short TTL |
| High event volume | Batch event consumers |
| Duplicate events | `seller_analytics_event_dedupe` unique key |
| Multi-currency | Group/filter by currency |
| Read scale | MySQL read replica for aggregate reads |

### Useful indexes

| Index | Why |
|---|---|
| `(seller_id, metric_date)` | Seller date range summary |
| `(seller_id, metric_date, currency)` | Currency-specific dashboard |
| `(seller_id, metric_date, revenue_amount)` | Top product ranking |
| `(event_id)` unique | Idempotent event processing |

---

## Step 36: Failure handling define kiya

### Projection failure

```text
Event consumer fails:
    message retry
    if max retry exceeded -> DLQ
    alert projection lag
    API serves last successful aggregate
```

### Cache failure

```text
Redis down:
    log warning
    query MySQL aggregate tables
    do not fail API only because cache failed
```

### Aggregate DB failure

```text
MySQL down:
    return 503 analytics_temporarily_unavailable
    emit metric
    trace error
```

> 🟢 **Principle:** Analytics failure checkout/order/payment flows ko block nahi karegi. Seller dashboard temporarily stale/empty ho sakta hai, but business transaction path safe rahega.

---

## Step 37: Implementation checklist banaya

| Item | Status |
|---|---:|
| `TaskImplementation/` folder exists | ✅ |
| `TaskImplementation/CMS Service/` folder preserved | ✅ |
| `task6.md` created | ✅ |
| Task source mapped | ✅ |
| Scope limited to CMS Service Task 6 | ✅ |
| Seller analytics metrics defined | ✅ |
| API route documented | ✅ |
| Permission `cms:analytics:read` documented | ✅ |
| Aggregate table blueprint included | ✅ |
| SQL examples included | ✅ |
| Go examples included | ✅ |
| Mermaid diagrams included | ✅ |
| External tools/libraries documented | ✅ |
| No backend code created | ✅ |
| No migration files created | ✅ |
| No API spec edited | ✅ |

---

## 🧾 Final Summary

CMS Service Task 6 ke liye Seller Analytics APIs ka final design:

| Area | Decision |
|---|---|
| Owner service | CMS Service |
| Main endpoint | `GET /api/v1/seller/dashboard/summary` |
| gRPC method | `CMSService.GetSellerAnalytics` |
| Required permission | `cms:analytics:read` |
| Metrics | Revenue, orders, conversion rate, top products |
| Data source | Aggregated read model from Order/Payment/Session events |
| Storage | Future MySQL aggregate tables in `cms_db` |
| Cache | Optional Redis short TTL |
| Cross-service DB access | Not allowed |
| Implementation type in this task | Documentation-only guide |

> 🟢 **Task 6 complete:** Seller analytics APIs ke liye revenue, orders, conversion, top products, permission rules, read model design, API contract, SQL examples, Go examples, diagrams, tools, cache, observability, and tests documented hain. No implementation beyond CMS Service Task 6 kiya gaya.
