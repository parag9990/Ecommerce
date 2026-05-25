# 🛒 Cart Service - Task 2: Choose MongoDB Plus Redis

![Task](https://img.shields.io/badge/Task-Cart%20Service%20%232-blue)
![Priority](https://img.shields.io/badge/Priority-P0-red)
![Dependency](https://img.shields.io/badge/Dependency-Cart%20Rules-yellow)
![Status](https://img.shields.io/badge/Status-Documented-success)
![Guide](https://img.shields.io/badge/Guide-Hinglish-orange)
![Scope](https://img.shields.io/badge/Scope-DB%20%2B%20Cache-purple)

## 📌 Task Summary

| Field | Detail |
|---|---|
| Task Name | Choose MongoDB plus Redis |
| Source | `docs/01-micro-tasks.md` → `Cart Service` → Task 2 |
| Priority | `P0` foundation/blocker |
| Dependency | Cart Service Task 1: Define cart rules |
| Main Goal | Active cart fast mutable hot data hai, isliye Redis cache aur MongoDB durable snapshot ke liye use karna |
| Output Type | Documentation-only implementation guide |
| Not Included | Actual Mongo collections create karna, Redis code likhna, add/remove item usecase implement karna, gRPC/REST handlers banana |

> **Simple Hinglish goal:** Is task ka purpose Cart Service ke storage decision ko final karna hai. MongoDB durable source of truth hoga, aur Redis fast hot cache/summary path hoga. Is file me implementation blueprint diya gaya hai, but actual backend code ya database schema create nahi kiya gaya because woh next Cart Service tasks ka scope hai.

---

## ✅ Final Output Created

```text
TaskImplementation/
└── Cart Service/
    ├── task1.md
    └── task2.md
```

### Why this structure?

- `TaskImplementation/` project ke task-wise implementation guides ka central folder hai.
- `Cart Service/` Cart Service ke saare task guides ko group karta hai.
- `task2.md` sirf **Cart Service - Task 2** ka MongoDB plus Redis decision and implementation guide hai.

---

## 🧭 Implementation Approach

Is guide ko banane se pehle existing project documentation study ki gayi:

| Document | Kya use kiya gaya |
|---|---|
| `docs/01-micro-tasks.md` | Task 2 ka exact scope: MongoDB plus Redis choose karna |
| `TaskImplementation/Cart Service/task1.md` | Cart rules: guest/user cart, active status, quantity, price snapshots |
| `docs/02-system-architecture.md` | Cart Service ka API Gateway, Product Service, MongoDB, Redis ke saath relation |
| `docs/04-microservice-design.md` | Cart responsibilities, DB choice, gRPC methods, REST routes |
| `docs/05-database-design.md` | Service-wise DB strategy: Cart = MongoDB + Redis |
| `database/mongodb-schema-design.md` | Future `cart_db`, `carts` document, indexes, TTL reference |
| `api/master-api.json` | Cart API/gRPC contract mapping for future implementation |

---

## 🪜 Step-by-Step Implementation

## Step 1: Task Boundary Clear Kiya

Cart Service Task 2 ka kaam **database/cache technology choice and usage strategy define karna** hai.

### Included

- MongoDB ko durable cart source of truth choose karna
- Redis ko hot cart cache, cart summary, aur quick badge count ke liye choose karna
- MongoDB and Redis responsibilities split karna
- Read/write/cache-aside strategy define karna
- Redis key naming, TTL, invalidation, and fallback rules define karna
- Future code structure and config examples document karna
- Install/use commands document karna for future implementation

### Not Included

- `carts` collection ka final schema create karna
- `cart_items` ya embedded item snapshots ka detailed schema implement karna
- MongoDB indexes actually run karna
- Go repository/cache files create karna
- `AddItem`, `RemoveItem`, `MergeGuestCart`, `CouponPreview` flows implement karna
- API Gateway route ya gRPC service code banana

> 🟢 **Reason:** `docs/01-micro-tasks.md` ke according Task 2 sirf DB/cache choice hai. Collection design Task 3 me hoga, add item Task 4 me, remove item Task 5 me, coupon preview Task 6 me, merge Task 7 me, aur expiry Task 8 me.

---

## Step 2: Cart Data Nature Samjha

Cart data normal product catalog se different hota hai:

| Cart Behavior | Meaning |
|---|---|
| Fast mutable | Buyer quantity baar-baar change kar sakta hai |
| Session/user scoped | Ek cart mostly ek `user_id` ya `guest_session_id` se tied hota hai |
| UX sensitive | Cart page and header badge fast load hone chahiye |
| Durable enough | Browser close hone ke baad bhi logged-in cart recover hona chahiye |
| Stale possible | Product price/stock change ho sakta hai, checkout se pehle refresh required hai |

### Decision

Cart ke liye single storage technology enough nahi hai:

- Sirf MongoDB use karenge to durability milegi, but hot reads/badge count slower ho sakte hain.
- Sirf Redis use karenge to speed milegi, but durable source weak ho jayega.
- MongoDB + Redis together use karenge to durability + speed dono balance honge.

---

## Step 3: MongoDB Durable Store Choose Kiya

MongoDB Cart Service ka **source of truth** hoga.

### Why MongoDB?

| Reason | Explanation |
|---|---|
| Flexible cart document | Cart ke andar items, snapshots, totals, coupon preview fields naturally document form me fit hote hain |
| Embedded items possible | `items[]` array cart ke saath store ho sakta hai, simple cart read fast hota hai |
| Guest and user carts | `user_id` ya `guest_session_id` based lookup indexes easy hain |
| TTL support | Stale carts ke liye `expires_at` TTL index future task me use ho sakta hai |
| Service ownership | Cart Service apna `cart_db` own karega, dusri services direct DB access nahi karengi |

### MongoDB Role

| Responsibility | MongoDB me hoga? |
|---|---|
| Active cart durable snapshot | ✅ Yes |
| Cart item snapshots | ✅ Yes, Task 3 me detail |
| Totals snapshot | ✅ Yes |
| Guest/user ownership | ✅ Yes |
| Expiry timestamp | ✅ Yes |
| Header badge cache | ❌ No, Redis better |
| Per-request temporary locks | ❌ No, Redis better |

### Future MongoDB Database Name

```text
cart_db
```

### Future Collection Reference

```text
carts
cart_audit_events
```

> 🟡 **Note:** `database/mongodb-schema-design.md` already `cart_db` and `carts` ka reference deta hai. Is Task 2 me collection actually create nahi ki gayi. Task 3 me schema and indexes detail honge.

---

## Step 4: Redis Hot Cache Choose Kiya

Redis Cart Service ka **hot data cache** hoga.

### Why Redis?

| Reason | Explanation |
|---|---|
| Very low latency | Header cart badge, summary, and active cart reads fast honge |
| TTL native | Guest cart cache and active cart cache automatically expire kar sakte hain |
| Atomic operations | Future quantity counters, locks, idempotency helpers me useful |
| Shared infra already planned | Platform Foundation me Redis rate limit, sessions, cart cache ke liye planned hai |
| MongoDB load reduce | Repeated cart reads MongoDB hit ki jagah Redis se serve ho sakte hain |

### Redis Role

| Responsibility | Redis me hoga? |
|---|---|
| Full active cart cache | ✅ Yes |
| Cart summary / badge count | ✅ Yes |
| Guest cart short-lived cache | ✅ Yes |
| Mutation lock candidate | ✅ Future use |
| Durable source of truth | ❌ No |
| Final checkout authority | ❌ No, Order/Product validation required |

> 🔴 **Important:** Redis ko primary database ki tarah use nahi karna. Agar Redis flush/restart ho jaye, Cart Service MongoDB se cart reload karke cache rebuild karega.

---

## Step 5: Final Storage Responsibility Split

| Data / Operation | MongoDB | Redis |
|---|---|---|
| Full active cart document | Durable source | Cached copy |
| `user_id` active cart lookup | Source of truth | Optional cached key |
| `guest_session_id` active cart lookup | Source of truth | Optional cached key |
| Cart badge count | Recomputed source | Fast cached value |
| Cart totals | Durable snapshot | Fast cached summary |
| Price snapshot | Durable snapshot | Cached inside cart JSON |
| Cart expiry metadata | Durable `expires_at` | Cache TTL |
| Audit events | Durable collection | Not needed |
| Checkout validation | MongoDB read, then Product/CMS validation | Cache can speed read, not final authority |

### Core Rule

```text
MongoDB = truth
Redis   = speed
```

Beginner-friendly version: MongoDB me original cart rahega. Redis me us cart ki fast copy rahegi. Agar Redis me data missing ho, MongoDB se data milega. Agar MongoDB me data missing ho, cart truly missing/empty maana jayega.

---

## Step 6: Cache Strategy Define Ki

Cart Service ke liye **cache-aside with write-through refresh** strategy best rahegi.

### Read Flow

1. Request me `user_id` ya `guest_session_id` identify karo.
2. Redis me active cart key lookup karo.
3. Redis hit hai to cart return karo.
4. Redis miss hai to MongoDB se active cart load karo.
5. MongoDB result ko Redis me short TTL ke saath cache karo.
6. Cart response return karo.

```mermaid
sequenceDiagram
    participant Client as API Gateway / Client
    participant Cart as Cart Service
    participant Redis as Redis Hot Cache
    participant Mongo as MongoDB cart_db

    Client->>Cart: GetCart(user_id or guest_session_id)
    Cart->>Redis: GET cart active key
    alt Redis hit
        Redis-->>Cart: Cached cart JSON
        Cart-->>Client: Cart response
    else Redis miss
        Cart->>Mongo: Find active cart
        Mongo-->>Cart: Durable cart document
        Cart->>Redis: SET cart JSON with TTL
        Cart-->>Client: Cart response
    end
```

### Write Flow

1. Mutation request validate karo.
2. Current cart MongoDB se load/update karo.
3. MongoDB update successful hone ke baad Redis cache refresh karo.
4. Redis refresh fail ho to request fail karna zaruri nahi, but warning log karo.
5. Next read Redis miss/stale case me MongoDB se rebuild kar lega.

```mermaid
sequenceDiagram
    participant Client as API Gateway / Client
    participant Cart as Cart Service
    participant Mongo as MongoDB cart_db
    participant Redis as Redis Hot Cache

    Client->>Cart: Add/Update/Remove item
    Cart->>Mongo: Update durable cart
    Mongo-->>Cart: Updated cart
    Cart->>Redis: SET updated cart and summary TTL
    alt Redis ok
        Redis-->>Cart: OK
    else Redis unavailable
        Redis-->>Cart: Error
        Cart->>Cart: Log warning, keep Mongo success
    end
    Cart-->>Client: Updated cart
```

### Why this strategy?

- MongoDB truth safe rahega.
- Redis fast reads dega.
- Redis failure se cart mutation data lose nahi hoga.
- Cache rebuild simple rahega.

---

## Step 7: Redis Key Design Define Kiya

Future Cart Service me Redis keys predictable and namespaced honi chahiye.

| Key Pattern | Purpose | Suggested TTL |
|---|---|---|
| `cart:active:user:{user_id}` | Logged-in user ka active cart JSON | `15m` |
| `cart:active:guest:{guest_session_id}` | Guest session ka active cart JSON | `30m` |
| `cart:summary:user:{user_id}` | Header badge count and total | `5m` |
| `cart:summary:guest:{guest_session_id}` | Guest header badge count and total | `5m` |
| `cart:lock:{cart_id}` | Future mutation lock | `5s` to `15s` |

### TTL Decision

| Cache | TTL | Reason |
|---|---|---|
| Full logged-in cart | `15 minutes` | Fast UX, but stale price risk limited |
| Full guest cart | `30 minutes` | Guest browsing flow me useful, Mongo still durable |
| Cart summary | `5 minutes` | Header badge frequent read hota hai, stale count short window acceptable |
| Mutation lock | `5-15 seconds` | Long lock avoid karna hai |

> 🟡 **Note:** Real cart expiry, like 30 days, MongoDB `expires_at` and TTL index se handle hoga. Redis TTL sirf cache freshness ke liye hai.

### Redis Value Examples

Full cart cache:

```json
{
  "cart_id": "cart_123",
  "owner_type": "user",
  "owner_id": "user_123",
  "status": "active",
  "items": [
    {
      "item_id": "item_1",
      "product_id": "prod_123",
      "variant_id": "var_1",
      "quantity": 2,
      "unit_price": {
        "amount": 299900,
        "currency": "INR"
      }
    }
  ],
  "totals": {
    "subtotal": 599800,
    "discount": 0,
    "total": 599800,
    "currency": "INR"
  },
  "cached_at": "2026-05-24T00:00:00Z"
}
```

Summary cache:

```json
{
  "item_count": 2,
  "unique_items": 1,
  "total": {
    "amount": 599800,
    "currency": "INR"
  }
}
```

---

## Step 8: Cache Invalidation Rules Define Kiye

Cart mutation ke baad stale cache user ko galat data dikha sakta hai. Isliye invalidation rules clear hone chahiye.

| Event | Redis Action |
|---|---|
| Item added | Full cart and summary cache refresh |
| Quantity updated | Full cart and summary cache refresh |
| Item removed | Full cart and summary cache refresh |
| Cart cleared | Full cart and summary cache delete or set empty |
| Coupon preview changed | Full cart cache refresh |
| Guest cart merged | Guest keys delete, user keys refresh |
| Cart checked out | Active cart keys delete |
| Cart expired | Active cart keys delete |

### Safe Rule

```text
Mutation successful in MongoDB → Redis cache refresh/delete
Mutation failed in MongoDB     → Redis cache untouched
Redis failure after MongoDB    → log warning, future read rebuilds cache
```

---

## Step 9: Consistency Model Decide Kiya

Cart Service ke liye **eventual cache consistency** acceptable hai, but MongoDB consistency primary hai.

| Area | Consistency Decision |
|---|---|
| Cart mutation result | MongoDB update ke baad response return hoga |
| Redis cache | Best-effort refresh |
| Checkout | Redis-only data trusted nahi hoga |
| Price/stock | Product Service se fresh validation required |
| Coupon final apply | Order/CMS final validation required |

### Beginner Explanation

Agar Redis thoda stale ho bhi jaye, worst case user ko short time ke liye old cart badge ya old cart copy dikhegi. Lekin actual checkout ke time MongoDB cart + Product Service validation + CMS validation se final truth confirm hoga.

---

## Step 10: Failure Handling Define Kiya

| Failure | Expected Behavior |
|---|---|
| Redis down | MongoDB se cart read/write continue, cache skip |
| MongoDB down | Cart read/write fail, because source of truth unavailable |
| Redis stale | TTL expire hone do ya mutation par refresh karo |
| Redis key missing | MongoDB se reload and cache rebuild |
| MongoDB cart missing | Empty active cart create flow future Task 4/5 me define hoga |
| Partial cache update fail | Warning log, metrics increment, request usually successful if MongoDB success |

### Priority Rule

```text
MongoDB failure = user-facing error
Redis failure   = degraded mode, but cart can still work
```

---

## Step 11: Future Config Contract Define Kiya

Cart Service ko environment variables se MongoDB and Redis config milega.

```env
CART_SERVICE_NAME=cart-service
CART_GRPC_PORT=9090

MONGO_URI=mongodb://ecommerce_root:ecommerce_password@mongo:27017/cart_db?authSource=admin
MONGO_DATABASE=cart_db

REDIS_ADDR=redis:6379
REDIS_PASSWORD=dev_redis_password
REDIS_DB=0

CART_CACHE_TTL=15m
CART_SUMMARY_CACHE_TTL=5m
GUEST_CART_CACHE_TTL=30m
```

### Local vs Container Host

| Runtime | Mongo Host | Redis Host |
|---|---|---|
| Service inside Docker Compose | `mongo:27017` | `redis:6379` |
| Service running directly on host | `localhost:27017` | `localhost:6379` |

---

## Step 12: Future Go Interfaces Plan Kiye

Ye code examples future implementation ke liye reference hain. Is Task 2 me ye files create nahi ki gayi.

### Cart Store Interface

```go
package repository

import (
	"context"
	"time"
)

type Cart struct {
	ID             string
	UserID         string
	GuestSessionID string
	Status         string
	UpdatedAt      time.Time
}

type CartStore interface {
	FindActiveByUserID(ctx context.Context, userID string) (*Cart, error)
	FindActiveByGuestSessionID(ctx context.Context, guestSessionID string) (*Cart, error)
	Save(ctx context.Context, cart *Cart) error
}
```

### Cart Cache Interface

```go
package repository

import (
	"context"
	"time"
)

type CartCache interface {
	GetActiveCart(ctx context.Context, ownerType string, ownerID string) (*Cart, error)
	SetActiveCart(ctx context.Context, ownerType string, ownerID string, cart *Cart, ttl time.Duration) error
	DeleteActiveCart(ctx context.Context, ownerType string, ownerID string) error
}
```

### Config Example

```go
package config

import "time"

type CartStorageConfig struct {
	MongoURI             string
	MongoDatabase        string
	RedisAddr            string
	RedisPassword        string
	RedisDB              int
	CartCacheTTL         time.Duration
	CartSummaryCacheTTL  time.Duration
	GuestCartCacheTTL    time.Duration
}
```

> 🟣 **Note:** Actual `mongo_cart_repository.go` and `redis_cart_cache.go` files future backend implementation me banenge. Task 2 sirf contract and decision document karta hai.

---

## Step 13: External Libraries and Tools

Task 2 me koi dependency install nahi ki gayi, because user request ka output only `task2.md` documentation hai. Future implementation ke liye recommended tools/libraries:

| Tool / Library | What it is | Why used | Install / Use |
|---|---|---|---|
| MongoDB `mongo:7` | Document database server | Durable cart snapshots, flexible item documents | Docker Compose service from Platform Foundation Task 4 |
| Redis `redis:7.2-alpine` | In-memory cache server | Hot cart cache, summary cache, future locks | Docker Compose service from Platform Foundation Task 4 |
| `go.mongodb.org/mongo-driver` | Official MongoDB Go driver | Go Cart Service ko MongoDB se connect karne ke liye | `go get go.mongodb.org/mongo-driver/mongo` |
| `github.com/redis/go-redis/v9` | Redis Go client | Go Cart Service ko Redis GET/SET/DEL ke liye | `go get github.com/redis/go-redis/v9` |
| Docker Compose | Local infra runner | MongoDB and Redis local start karne ke liye | `docker compose -f infra/compose/docker-compose.local.yml up -d mongo redis` |
| `mongosh` | MongoDB shell | Health check, indexes, sample queries verify karne ke liye | MongoDB image ke andar available hota hai |
| `redis-cli` | Redis CLI | PING, GET, TTL, DEL debug karne ke liye | Redis image ke andar available hota hai |
| Mermaid | Markdown diagram syntax | Architecture/flow diagrams readable banane ke liye | GitHub/Markdown viewer usually render karta hai |

### Future Install Commands

```bash
go get go.mongodb.org/mongo-driver/mongo
go get github.com/redis/go-redis/v9
```

### Future Local Run Commands

```bash
docker compose -f infra/compose/docker-compose.local.yml up -d mongo redis
docker exec ecommerce-mongo mongosh --quiet --eval "db.adminCommand({ ping: 1 })"
docker exec ecommerce-redis redis-cli -a dev_redis_password ping
```

Expected health results:

```text
MongoDB: { ok: 1 }
Redis: PONG
```

---

## 🏗️ Architecture Diagram

```mermaid
flowchart TB
    Client[React App / API Client] --> GW[API Gateway]
    GW --> Cart[Cart Service]

    Cart --> Redis[(Redis Hot Cache)]
    Cart --> Mongo[(MongoDB cart_db)]
    Cart --> Product[Product Service]
    Cart --> CMS[CMS Service]
    Order[Order Service] --> Cart

    Redis -. cache miss fallback .-> Mongo

    classDef svc fill:#e8f2ff,stroke:#2563eb,color:#111827
    classDef db fill:#ecfdf5,stroke:#059669,color:#111827
    classDef cache fill:#fff7ed,stroke:#ea580c,color:#111827

    class GW,Cart,Product,CMS,Order svc
    class Mongo db
    class Redis cache
```

### Explanation

- Frontend direct Cart Service ko call nahi karega; API Gateway public REST route handle karega.
- Cart Service internal logic me MongoDB and Redis dono use karega.
- Product Service price/stock validation ke liye future add/update/checkout flows me use hoga.
- CMS Service coupon preview ke liye future Task 6 me use hoga.
- Order Service checkout ke time Cart Service se durable cart data lega.

---

## 🧱 Clean Folder Structure

### Created in this task

```text
TaskImplementation/
└── Cart Service/
    ├── task1.md
    └── task2.md
```

### Future Cart Service code structure

This is target structure from project docs. It is **not created in Task 2**.

```text
backend/
└── services/
    └── cart-service/
        ├── cmd/
        │   └── server/
        │       └── main.go
        ├── internal/
        │   ├── domain/
        │   │   └── cart.go
        │   ├── usecase/
        │   │   ├── add_item.go
        │   │   ├── remove_item.go
        │   │   ├── merge_cart.go
        │   │   └── price_cart.go
        │   ├── repository/
        │   │   ├── mongo_cart_repository.go
        │   │   └── redis_cart_cache.go
        │   └── transport/
        │       └── grpc/
        │           └── server.go
        └── deploy/
```

### Why this future structure?

- `domain/` cart entities and rules rakhega.
- `usecase/` add/remove/merge/price flows rakhega.
- `repository/mongo_cart_repository.go` durable MongoDB operations rakhega.
- `repository/redis_cart_cache.go` cache GET/SET/DEL operations rakhega.
- `transport/grpc/` CartService gRPC methods expose karega.

---

## 🔄 End-to-End Decision Flow

```mermaid
flowchart TD
    A[Cart Service Task 1 Rules] --> B{Cart data type?}
    B --> C[Fast mutable active cart]
    B --> D[Durable cart snapshot needed]
    C --> E[Choose Redis for hot cache]
    D --> F[Choose MongoDB for source of truth]
    E --> G[Define cache keys and TTL]
    F --> H[Use cart_db and carts collection in future Task 3]
    G --> I[Task 2 storage strategy complete]
    H --> I
```

---

## 🧪 Verification Checklist

Since Task 2 documentation-only hai, verification ka focus scope and completeness par hai:

| Check | Status |
|---|---|
| `TaskImplementation/Cart Service/task2.md` created | ✅ |
| Task 2 scope documented | ✅ |
| MongoDB source-of-truth decision documented | ✅ |
| Redis hot cache decision documented | ✅ |
| Cache read/write strategy documented | ✅ |
| Redis key naming and TTLs documented | ✅ |
| Failure handling documented | ✅ |
| External tools/libraries explained | ✅ |
| Future folder structure documented | ✅ |
| Mermaid diagrams added | ✅ |
| No backend code implemented beyond Task 2 | ✅ |

---

## 🚫 Out of Scope for Task 2

| Not Implemented | Belongs To |
|---|---|
| Final `carts` collection schema and indexes | Cart Service Task 3 |
| `AddItem` usecase | Cart Service Task 4 |
| `RemoveItem` and quantity cleanup | Cart Service Task 5 |
| Coupon preview logic | Cart Service Task 6 |
| Guest cart merge implementation | Cart Service Task 7 |
| Cart expiry cleanup job | Cart Service Task 8 |
| Checkout conversion and inventory reserve | Order Service Task 3 |
| API Gateway route implementation | API Gateway tasks |

> 🔴 **Reason:** Agar Task 2 me schema, repository code, ya add/remove cart logic implement kar diya jaye, to later tasks ka boundary break ho jayega.

---

## ✅ Final Task 2 Standard

Cart Service ke liye final DB/cache standard:

1. MongoDB Cart Service ka durable source of truth hoga.
2. Redis sirf hot cache, summary cache, badge count, and future short-lived locks ke liye use hoga.
3. Redis failure cart data loss nahi karega, because MongoDB primary truth hai.
4. MongoDB failure user-facing error hoga, because durable cart unavailable hai.
5. Reads Redis-first, MongoDB-fallback pattern follow karenge.
6. Writes MongoDB-first, Redis-refresh pattern follow karenge.
7. Checkout Redis-only data par trust nahi karega.
8. Cache keys namespaced and owner-based honge.
9. TTL cache freshness ke liye hoga, long-term expiry MongoDB `expires_at` se manage hogi.
10. Actual schema and repository implementation future Cart Service tasks me hoga.

Task 2 complete hai as a structured MongoDB plus Redis implementation guide. Future Cart Service tasks isi storage strategy ke basis par collection schema, usecases, repository code, cache invalidation, tests, and cleanup jobs implement karenge.
