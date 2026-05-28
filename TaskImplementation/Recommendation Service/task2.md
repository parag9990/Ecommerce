# 🧠 Recommendation Service - Task 2: Choose MongoDB plus Redis

![Task](https://img.shields.io/badge/Task-Recommendation%20Service%20%232-blue)
![Priority](https://img.shields.io/badge/Priority-P2-orange)
![Dependency](https://img.shields.io/badge/Dependency-Recommendation%20Types-yellow)
![Status](https://img.shields.io/badge/Status-Documented-success)
![Guide](https://img.shields.io/badge/Guide-Hinglish-ff69b4)
![Storage](https://img.shields.io/badge/Storage-MongoDB%20%2B%20Redis-47A248)
![Scope](https://img.shields.io/badge/Scope-DB%20Choice%20%2B%20Cache%20Strategy-purple)

## 📌 Task Summary

| Field | Detail |
|---|---|
| Task Name | Choose MongoDB plus Redis |
| Source | `docs/01-micro-tasks.md` → `Recommendation Service` → Task 2 |
| Priority | `P2` optimization / advanced capability |
| Dependency | Recommendation Service Task 1: recommendation types |
| Main Goal | Recommendation documents ke liye MongoDB choose karna aur low-latency top lists ke liye Redis cache strategy define karna |
| Output Type | Structured implementation guide |
| Not Included | Event ingestion, final feature store schema, rule-based ranking engine, personalized ML ranking, gRPC endpoint, A/B testing implementation |

> **Simple Hinglish goal:** Is task ka kaam hai Recommendation Service ke storage stack ko clearly choose karna. MongoDB flexible recommendation data store karega, aur Redis fast cached recommendation lists serve karega. Is guide me actual backend code create nahi kiya gaya, kyunki Task 2 storage decision and setup strategy tak limited hai.

---

## ✅ Final Output Created

```text
TaskImplementation/
└── Recommendation Service/
    ├── task1.md
    └── task2.md
```

### Why this structure?

- `TaskImplementation/` project ke task-wise guides ka central folder hai.
- `Recommendation Service/` recommendation-service related tasks ko group karta hai.
- `task1.md` already existing recommendation types guide hai.
- `task2.md` sirf **Recommendation Service - Task 2** ka storage and cache guide hai.
- `backend/services/recommendation-service/` me koi new service code create nahi kiya gaya, taaki scope Task 2 se bahar na jaye.

---

## 🧭 Documents Studied

| Document | Kya use kiya gaya |
|---|---|
| `docs/01-micro-tasks.md` | Task 2 ka exact scope: MongoDB plus Redis choose karna |
| `docs/04-microservice-design.md` | Recommendation Service purpose, responsibilities, collections, APIs |
| `docs/05-database-design.md` | Service-wise database strategy and Recommendation DB choice |
| `database/mongodb-schema-design.md` | `recommendation_db`, `user_interactions`, `recommendation_sets`, indexes |
| `docs/03-folder-structure.md` | Future recommendation-service repository/cache files ka folder layout |
| `docs/12-logging-monitoring-scalability.md` | Redis recommendation cache TTL and graceful degradation rule |
| `TaskImplementation/Recommendation Service/task1.md` | Recommendation types and contexts from Task 1 |

---

## 🧱 Task Boundary

### ✅ Included in Task 2

- MongoDB plus Redis ko Recommendation Service ke storage stack ke roop me choose karna
- MongoDB ka purpose define karna: flexible recommendation documents and generated sets
- Redis ka purpose define karna: low-latency top lists and cached recommendation responses
- High-level database ownership rule define karna
- Recommended MongoDB collections list karna
- Basic indexes and TTL rules explain karna
- Redis key naming, TTL, and invalidation approach define karna
- Local setup commands and future Go client libraries mention karna
- Read/write/cache flow diagrams add karna
- Beginner-friendly code examples dena

### ❌ Not Included in Task 2

- Event consumer implement karna, kyunki wo Task 3 hai
- Detailed feature store schema finalize karna, kyunki wo Task 4 hai
- Trending/category/seller ranking engine banana, kyunki wo Task 5 hai
- Personalized ranking algorithm banana, kyunki wo Task 6 hai
- `GetRecommendations(user_id, context)` gRPC endpoint banana, kyunki wo Task 7 hai
- A/B testing hooks implement karna, kyunki wo Task 8 hai
- Production Kubernetes/managed database provisioning karna

> 🟢 **Rule:** Task 2 me hum storage architecture ready karte hain. Data ingest, scoring, API serving, aur analytics later tasks me aayenge.

---

## 🏗️ Storage Decision Overview

| Layer | Tool | Role | Why |
|---|---|---|---|
| Durable store | MongoDB | User interactions, generated recommendation sets, product feature snapshots | Recommendation data flexible hota hai; fields strategy ke hisaab se evolve kar sakte hain |
| Hot cache | Redis | Top/trending/personalized list cache | Recommendation response fast chahiye; Redis low-latency reads ke liye best fit hai |
| Source of truth | MongoDB | Cache miss ke baad fallback read | Redis volatile cache hai; durable data MongoDB me rahega |
| Future events | Kafka/RabbitMQ | Events se recommendation data update | Task 3 me implement hoga |

### Simple mental model

```text
MongoDB = memory bank
Redis   = fast serving counter
```

MongoDB me durable recommendation data rahega. Redis me wahi data short TTL ke saath cache hoga jo UI/API ko frequently chahiye.

---

## 🧩 Why MongoDB?

Recommendation data fixed relational shape me nahi hota. Aaj rule-based score ho sakta hai, kal ML embedding reference aa sakta hai, aur future me campaign boost ya seller boost add ho sakta hai.

### MongoDB benefits

| Benefit | Hinglish Explanation |
|---|---|
| Flexible documents | Recommendation sets, metadata, scoring reasons, aur model version fields easily add ho sakte hain |
| Natural JSON shape | API response aur stored document ka shape similar ho sakta hai |
| Fast indexed reads | `context_key`, `strategy_id`, `user_id`, `product_id` based reads efficient ho sakte hain |
| TTL indexes | Old interactions/recommendation sets automatically expire ho sakte hain |
| Independent ownership | Recommendation Service apna database own karega, dusri service ke DB ko direct read nahi karega |

### MongoDB me kya store hoga?

| Collection | Purpose | Task 2 Detail Level |
|---|---|---|
| `user_interactions` | Product views/clicks/cart/wishlist/purchase events ka durable log | Basic collection and indexes only |
| `product_features` | Recommendation ke liye product metadata snapshot/reference | High-level placeholder only |
| `recommendation_sets` | Generated top lists, similar lists, personalized list references | Basic document shape and indexes |
| `ab_test_assignments` | Future strategy experiment assignment | Listed only, implementation Task 8 me |

> 🟡 **Scope note:** `product_features` aur `ab_test_assignments` ki detailed schema Task 4/Task 8 me finalize hogi. Task 2 me sirf storage choice and collection boundary define hai.

---

## ⚡ Why Redis?

Recommendation sections mostly page load ke time render hoti hain. Agar har request MongoDB aggregation ya ranking calculation karegi, to latency badhegi. Redis me precomputed lists cache karne se API response fast milega.

### Redis benefits

| Benefit | Hinglish Explanation |
|---|---|
| Very low latency | Home page/product page pe recommendations quickly serve hongi |
| TTL support | Cached lists automatically expire ho sakti hain |
| Sorted sets | Trending score based lists naturally store ho sakti hain |
| Simple invalidation | Product update/event ke baad relevant keys delete/update ho sakti hain |
| Graceful fallback | MongoDB slow/down ho to last cached Redis data serve ho sakta hai |

### Redis me kya cache hoga?

| Cache Type | Example Key | Value Type | Suggested TTL |
|---|---|---|---:|
| Global trending | `reco:v1:trending:global` | JSON list or sorted set | 15 min |
| Category trending | `reco:v1:trending:category:cat_shoes` | JSON list or sorted set | 15 min |
| Seller popular | `reco:v1:trending:seller:seller_456` | JSON list or sorted set | 15 min |
| Similar products | `reco:v1:similar:product:prod_123` | JSON list | 15 min |
| Personalized feed | `reco:v1:personalized:user:user_123` | JSON list | 5 to 15 min |
| Guest personalized | `reco:v1:personalized:anon:anon_123` | JSON list | 5 min |

> `docs/12-logging-monitoring-scalability.md` me recommendation cache TTL `15 min` recommended hai. Personalized/guest cache zyada dynamic ho sakta hai, isliye uska TTL shorter rakhna better hai.

---

## 🪜 Step-by-Step Implementation Guide

### Step 1: Task 1 ke recommendation types ko base banaya

Task 1 me ye recommendation types define ho chuke hain:

| Type | Storage Need |
|---|---|
| `similar_products` | Product features and generated similar list |
| `trending` | Recent interactions, counters, and cached top lists |
| `personalized` | User/anonymous interaction history and generated user list |
| `frequently_bought_together` | Purchase/cart co-occurrence data and generated bundle list |

Task 2 me in types ke basis par decide kiya gaya ki:

- Flexible raw/generated docs MongoDB me rahenge.
- Hot response lists Redis me cache hongi.
- Long-running ranking calculation request path par nahi hogi.

---

### Step 2: Database ownership rule define kiya

Microservice docs ka golden rule hai:

```text
Har service apni database own karegi.
Dusri service ke DB ko directly read/write nahi karna.
Data chahiye to gRPC call ya event projection use karo.
```

Recommendation Service ke liye iska matlab:

- Product Service ke `products` collection ko direct query nahi karna.
- Session Service ke session DB ko direct read nahi karna.
- Product/session/order/wishlist events se apna recommendation projection build karna.
- Recommendation Service ka own DB hoga: `recommendation_db`.

### Ownership diagram

```mermaid
flowchart LR
    Product[Product Service] -->|Product events| Queue[(Kafka/RabbitMQ)]
    Session[Session Service] -->|View/click events| Queue
    Wishlist[Wishlist Service] -->|Wishlist events| Queue
    Order[Order Service] -->|Purchase events| Queue

    Queue --> Reco[Recommendation Service]
    Reco --> Mongo[(MongoDB: recommendation_db)]
    Reco --> Redis[(Redis: recommendation cache)]

    API[API Gateway / Product Page] --> Reco
```

> 🟣 **Task 2 view:** Queue/event consumer diagram me future integration dikhaya gaya hai, but consumer implementation Task 3 me hoga.

---

### Step 3: MongoDB ko durable store choose kiya

MongoDB `recommendation_db` me durable data store karega.

Recommended database name:

```text
recommendation_db
```

Recommended collections:

```text
recommendation_db
├── user_interactions
├── product_features
├── recommendation_sets
└── ab_test_assignments
```

### Collection responsibility

| Collection | Responsibility |
|---|---|
| `user_interactions` | User/session/product events ka event-level history |
| `product_features` | Product metadata/features ka recommendation-friendly snapshot |
| `recommendation_sets` | Precomputed recommendation results per context/strategy |
| `ab_test_assignments` | Future A/B test strategy assignment |

---

### Step 4: `user_interactions` collection ka base design define kiya

Ye collection raw interaction signals store karegi. Ye Task 3 event ingestion ke baad populate hogi.

### Example document

```json
{
  "_id": "interaction_123",
  "user_id": "user_123",
  "anonymous_id": "anon_123",
  "product_id": "prod_123",
  "category_id": "cat_shoes",
  "event_type": "product_view",
  "weight": 1,
  "occurred_at": "2026-05-18T00:00:00Z",
  "created_at": "2026-05-18T00:00:00Z"
}
```

### Important fields

| Field | Why Needed |
|---|---|
| `_id` | Interaction unique identify karne ke liye |
| `user_id` | Logged-in personalization ke liye |
| `anonymous_id` | Guest personalization/session continuity ke liye |
| `product_id` | Product-level scoring ke liye |
| `category_id` | Category trending/personalized preference ke liye |
| `event_type` | View/click/cart/wishlist/purchase signal distinguish karne ke liye |
| `weight` | Event strength define karne ke liye |
| `occurred_at` | Time-window ranking ke liye |

### Recommended indexes

```javascript
db.user_interactions.createIndex({ user_id: 1, occurred_at: -1 })
db.user_interactions.createIndex({ anonymous_id: 1, occurred_at: -1 })
db.user_interactions.createIndex({ product_id: 1, event_type: 1, occurred_at: -1 })
db.user_interactions.createIndex({ category_id: 1, event_type: 1, occurred_at: -1 })
db.user_interactions.createIndex({ occurred_at: 1 }, { expireAfterSeconds: 15552000 })
```

### Why these indexes?

| Index | Use |
|---|---|
| `{ user_id, occurred_at }` | User ke recent behavior se personalized recommendation banana |
| `{ anonymous_id, occurred_at }` | Guest user ke recent behavior se lightweight personalization |
| `{ product_id, event_type, occurred_at }` | Product popularity and co-occurrence calculations |
| `{ category_id, event_type, occurred_at }` | Category-level trending list |
| TTL on `occurred_at` | Old raw events auto cleanup; example TTL around 180 days |

> 🟡 **Beginner note:** TTL index MongoDB ko bolta hai ki given time ke baad document automatically delete kar do. Analytics aggregates long-term store ho sakte hain, but raw events forever store karna costly hota hai.

---

### Step 5: `recommendation_sets` collection ka base design define kiya

Ye collection generated recommendation lists store karegi. API/cache miss ke time yahi durable source hoga.

### Example document

```json
{
  "_id": "reco_set_home_trending_v1",
  "context_key": "home_feed:global",
  "recommendation_type": "trending",
  "strategy_id": "trending_v1_recent_activity",
  "items": [
    {
      "product_id": "prod_123",
      "score": 98.5,
      "rank": 1,
      "reason": "high_recent_activity"
    },
    {
      "product_id": "prod_456",
      "score": 91.2,
      "rank": 2,
      "reason": "high_add_to_cart_rate"
    }
  ],
  "metadata": {
    "window": "24h",
    "source": "rule_based_mvp",
    "generated_by": "recommendation-job"
  },
  "generated_at": "2026-05-18T00:00:00Z",
  "expires_at": "2026-05-18T00:15:00Z"
}
```

### Recommended indexes

```javascript
db.recommendation_sets.createIndex(
  { context_key: 1, strategy_id: 1 },
  { unique: true }
)

db.recommendation_sets.createIndex({ expires_at: 1 }, { expireAfterSeconds: 0 })
db.recommendation_sets.createIndex({ recommendation_type: 1, generated_at: -1 })
```

### Why these indexes?

| Index | Use |
|---|---|
| `{ context_key, strategy_id }` unique | Same context and strategy ke duplicate generated sets avoid karna |
| TTL on `expires_at` | Expired generated lists cleanup karna |
| `{ recommendation_type, generated_at }` | Latest sets inspect/debug karna |

---

### Step 6: `product_features` ka high-level boundary define kiya

`product_features` collection Product Service se received product events ka recommendation-friendly projection ho sakti hai.

### Example document

```json
{
  "_id": "prod_123",
  "product_id": "prod_123",
  "category_id": "cat_shoes",
  "seller_id": "seller_456",
  "brand": "Acme",
  "attributes": {
    "gender": "men",
    "material": "mesh",
    "color": "black"
  },
  "price_bucket": "2000_3000",
  "status": "published",
  "in_stock": true,
  "updated_at": "2026-05-18T00:00:00Z"
}
```

### Recommended indexes

```javascript
db.product_features.createIndex({ product_id: 1 }, { unique: true })
db.product_features.createIndex({ category_id: 1, status: 1, in_stock: 1 })
db.product_features.createIndex({ seller_id: 1, status: 1, in_stock: 1 })
```

> 🟡 **Scope note:** Ye final feature store schema nahi hai. Task 4 me counters, embeddings/reference IDs, aur richer feature model design hoga.

---

### Step 7: Redis key strategy define kiya

Redis keys predictable honi chahiye, taaki debug, invalidation, aur monitoring easy ho.

### Key naming pattern

```text
reco:{version}:{type}:{scope}:{id}
```

### Examples

```text
reco:v1:trending:global
reco:v1:trending:category:cat_shoes
reco:v1:trending:seller:seller_456
reco:v1:similar:product:prod_123
reco:v1:personalized:user:user_123
reco:v1:personalized:anon:anon_123
reco:v1:fbt:product:prod_123
```

### Why version prefix?

`v1` prefix future migration easy banata hai. Agar strategy or payload format change hota hai, `reco:v2:*` keys use ho sakti hain without old cache conflict.

---

### Step 8: Redis value shape define kiya

Simple JSON cache value beginner-friendly and API response friendly hai.

### Example Redis JSON value

```json
{
  "recommendation_id": "reco_20260518_home_global",
  "context_key": "home_feed:global",
  "recommendation_type": "trending",
  "strategy_id": "trending_v1_recent_activity",
  "items": [
    {
      "product_id": "prod_123",
      "score": 98.5,
      "rank": 1,
      "reason": "high_recent_activity"
    },
    {
      "product_id": "prod_456",
      "score": 91.2,
      "rank": 2,
      "reason": "high_add_to_cart_rate"
    }
  ],
  "cached_at": "2026-05-18T00:00:00Z",
  "expires_at": "2026-05-18T00:15:00Z"
}
```

### Redis type choice

| Need | Redis Type | Why |
|---|---|---|
| Full API-like cached response | `STRING` with JSON | Easy get/set and simple response hydration |
| Trending score ranking | `ZSET` | Score-based top-N reads efficient |
| Short lock during rebuild | `SET NX EX` | Prevent multiple workers from rebuilding same cache |

For Task 2, recommended default:

```text
Use STRING JSON for response cache.
Use ZSET later where score updates are frequent.
```

---

### Step 9: Cache TTL and invalidation rules define kiye

### Suggested TTLs

| Recommendation Type | TTL | Reason |
|---|---:|---|
| Global trending | 15 min | Stable enough, but should refresh often |
| Category trending | 15 min | Category activity changes throughout day |
| Seller popular | 15 min | Seller storefront should feel fresh |
| Similar products | 15 min | Product updates can affect similarity |
| Personalized user | 5 to 15 min | User actions change feed quickly |
| Guest personalized | 5 min | Anonymous behavior is session-heavy |

### Invalidation triggers

| Trigger | Cache Action |
|---|---|
| Product unpublished/deleted | Delete similar/category/seller keys containing that product |
| Product stock changed to out of stock | Refresh affected recommendation sets |
| Product category changed | Delete old/new category trending keys |
| New interaction burst | Let TTL expire or refresh async |
| Strategy version change | Write new `reco:v2:*` keys |

> 🟢 **Practical default:** Pehle TTL-based invalidation enough hai. Fine-grained invalidation later add karna better hai jab event ingestion and ranking jobs ready hon.

---

### Step 10: Read path design kiya

Recommendation read request me pehle Redis check hoga. Cache hit par fast response. Cache miss par MongoDB se latest generated recommendation set read hoga. Agar MongoDB me bhi data missing hai, fallback strategy use hogi.

### Read flow

```mermaid
flowchart TD
    A[Recommendation Request] --> B[Build cache key]
    B --> C{Redis cache hit?}
    C -->|Yes| D[Return cached recommendation list]
    C -->|No| E[Read recommendation_sets from MongoDB]
    E --> F{MongoDB set found and not expired?}
    F -->|Yes| G[Write Redis cache with TTL]
    G --> H[Return recommendation list]
    F -->|No| I[Use fallback: category popular or global trending]
    I --> J[Optional: cache fallback briefly]
    J --> K[Return fallback list]
```

### Hinglish explanation

1. Request aati hai, jaise home page ke liye `context=home_feed`.
2. Service Redis key banati hai, jaise `reco:v1:trending:global`.
3. Redis me data mil gaya to direct return.
4. Redis miss hua to MongoDB `recommendation_sets` check.
5. MongoDB me valid data mila to Redis me cache karke return.
6. Data missing/expired hua to fallback trending/category popular return.

---

### Step 11: Write/update path design kiya

Task 2 me actual write job implement nahi hota, but storage design ko support karna zaruri hai.

### Future write flow

```mermaid
sequenceDiagram
    participant E as Product/Session/Order Events
    participant C as Recommendation Consumer
    participant M as MongoDB
    participant J as Ranking Job
    participant R as Redis

    E->>C: product_view / add_to_cart / purchase
    C->>M: Store user_interactions
    J->>M: Read interactions/features
    J->>M: Upsert recommendation_sets
    J->>R: Set hot cache with TTL
```

### Task 2 responsibility

- MongoDB collections ready for future event writes.
- Redis key/value/TTL pattern ready for future cache writes.
- Actual consumer and ranking job Task 3/Task 5 me.

---

### Step 12: Failure behavior define kiya

Recommendation Service optional user-experience enhancement hai. Checkout ya product page ko block nahi karna chahiye.

| Failure | Behavior |
|---|---|
| Redis down | MongoDB se read try karo; response thoda slow ho sakta hai |
| MongoDB down | Redis me stale-but-available cached data ho to return karo |
| Both down | Empty recommendation list return karo ya UI section hide karne ke liye safe response do |
| Cache corrupt JSON | Cache delete karo and MongoDB fallback try karo |
| Recommendation expired | Fresh set read/rebuild later; request path me heavy calculation avoid karo |

### Safe empty response example

```json
{
  "products": [],
  "strategy_id": "fallback_empty_v1",
  "metadata": {
    "reason": "recommendation_unavailable"
  }
}
```

> 🟡 **Important:** Recommendation failure checkout/payment/order flow ko kabhi block nahi karega.

---

## 🗂️ Clean Folder Structure

### Created documentation structure

```text
TaskImplementation/
└── Recommendation Service/
    ├── task1.md
    └── task2.md
```

### Future backend structure recommended by docs

```text
backend/
└── services/
    └── recommendation-service/
        ├── cmd/
        │   └── server/
        │       └── main.go
        ├── internal/
        │   ├── domain/
        │   │   └── recommendation.go
        │   ├── usecase/
        │   │   ├── trending.go
        │   │   └── personalized.go
        │   ├── repository/
        │   │   ├── mongo_feature_repository.go
        │   │   └── redis_recommendation_cache.go
        │   ├── events/
        │   │   └── interaction_consumer.go
        │   └── transport/
        │       └── grpc/
        └── deploy/
```

### Task 2 future files mapping

| Future File | Purpose |
|---|---|
| `internal/repository/mongo_feature_repository.go` | MongoDB collections read/write wrapper |
| `internal/repository/redis_recommendation_cache.go` | Redis get/set/delete cache wrapper |
| `internal/domain/recommendation.go` | Recommendation set/cache models |
| `deploy/` | Future service deployment config |

> 🟢 **Scope note:** Ye files Task 2 me create nahi kiye gaye. Ye future implementation map hai.

---

## 🧪 MongoDB Setup Example

### Local Docker command

```bash
docker run --name ecommerce-mongo \
  -p 27017:27017 \
  -e MONGO_INITDB_ROOT_USERNAME=root \
  -e MONGO_INITDB_ROOT_PASSWORD=secret \
  -d mongo:7
```

### Connect with mongosh

```bash
mongosh "mongodb://root:secret@localhost:27017/admin"
```

### Create DB and indexes

```javascript
use recommendation_db

db.user_interactions.createIndex({ user_id: 1, occurred_at: -1 })
db.user_interactions.createIndex({ anonymous_id: 1, occurred_at: -1 })
db.user_interactions.createIndex({ product_id: 1, event_type: 1, occurred_at: -1 })
db.user_interactions.createIndex({ category_id: 1, event_type: 1, occurred_at: -1 })
db.user_interactions.createIndex({ occurred_at: 1 }, { expireAfterSeconds: 15552000 })

db.recommendation_sets.createIndex(
  { context_key: 1, strategy_id: 1 },
  { unique: true }
)
db.recommendation_sets.createIndex({ expires_at: 1 }, { expireAfterSeconds: 0 })
db.recommendation_sets.createIndex({ recommendation_type: 1, generated_at: -1 })

db.product_features.createIndex({ product_id: 1 }, { unique: true })
db.product_features.createIndex({ category_id: 1, status: 1, in_stock: 1 })
db.product_features.createIndex({ seller_id: 1, status: 1, in_stock: 1 })
```

> 🟡 **Note:** Commands future setup ke liye hain. Is Task 2 guide generation ke dauraan database start ya modify nahi kiya gaya.

---

## ⚡ Redis Setup Example

### Local Docker command

```bash
docker run --name ecommerce-redis \
  -p 6379:6379 \
  -d redis:7-alpine
```

### Connect with redis-cli

```bash
redis-cli -h localhost -p 6379
```

### Set and get recommendation cache

```bash
SETEX reco:v1:trending:global 900 '{"strategy_id":"trending_v1_recent_activity","items":[{"product_id":"prod_123","score":98.5}]}'
GET reco:v1:trending:global
TTL reco:v1:trending:global
```

### Sorted set example for trending

```bash
ZADD reco:v1:trending:zset:global 98.5 prod_123 91.2 prod_456
ZREVRANGE reco:v1:trending:zset:global 0 9 WITHSCORES
EXPIRE reco:v1:trending:zset:global 900
```

> `900` seconds = 15 minutes.

---

## 📦 External Libraries / Tools

### 1. MongoDB

| Field | Detail |
|---|---|
| What | Document database |
| Why used | Flexible recommendation documents, interaction logs, generated sets store karne ke liye |
| Install local | `docker run ... mongo:7` |
| Use | `recommendation_db` database and indexes create karna |

### 2. Redis

| Field | Detail |
|---|---|
| What | In-memory cache/data structure store |
| Why used | Low-latency cached recommendation lists and top-N reads ke liye |
| Install local | `docker run ... redis:7-alpine` |
| Use | `SETEX`, `GET`, `ZADD`, `ZREVRANGE`, `EXPIRE` |

### 3. MongoDB Go Driver

| Field | Detail |
|---|---|
| Package | `go.mongodb.org/mongo-driver/mongo` |
| Why used | Go service se MongoDB connect/query/upsert karne ke liye |
| Install | `go get go.mongodb.org/mongo-driver/mongo` |
| Future file | `internal/repository/mongo_feature_repository.go` |

### 4. go-redis

| Field | Detail |
|---|---|
| Package | `github.com/redis/go-redis/v9` |
| Why used | Go service se Redis cache read/write karne ke liye |
| Install | `go get github.com/redis/go-redis/v9` |
| Future file | `internal/repository/redis_recommendation_cache.go` |

### 5. Docker / Docker Compose

| Field | Detail |
|---|---|
| What | Local services run karne ka tool |
| Why used | MongoDB and Redis local development me quickly run karne ke liye |
| Install | Docker Desktop or Docker Engine |
| Use | Local stack, integration testing, development environment |

> 🟢 **Task 2 status:** Libraries/tools identify kiye gaye hain. Dependency install command future implementation ke liye documented hai; is guide generation me `go get` run nahi kiya gaya.

---

## 🔐 Environment Variables

Future Recommendation Service ke `.env` me ye values useful hongi:

```env
RECOMMENDATION_MONGO_URI=mongodb://root:secret@localhost:27017
RECOMMENDATION_MONGO_DATABASE=recommendation_db

RECOMMENDATION_REDIS_ADDR=localhost:6379
RECOMMENDATION_REDIS_PASSWORD=
RECOMMENDATION_REDIS_DB=0

RECOMMENDATION_CACHE_TTL_SECONDS=900
RECOMMENDATION_PERSONALIZED_CACHE_TTL_SECONDS=300
```

### Why separate env names?

- Service-specific prefix config conflicts avoid karta hai.
- Same Mongo/Redis cluster me different services alag DB/index use kar sakti hain.
- TTL environment se tune ho sakta hai without code deploy.

---

## 💻 Future Go Code Examples

> Ye examples implementation direction dikhane ke liye hain. Task 2 me actual Go files create nahi kiye gaye.

### MongoDB client setup example

```go
package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func NewMongoClient(ctx context.Context, uri string) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	return client, nil
}
```

### Redis client setup example

```go
package repository

import (
	"context"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(ctx context.Context, addr string, password string, db int) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return client, nil
}
```

### Cache repository example

```go
package repository

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RecommendationCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRecommendationCache(client *redis.Client, ttl time.Duration) *RecommendationCache {
	return &RecommendationCache{client: client, ttl: ttl}
}

func (c *RecommendationCache) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

func (c *RecommendationCache) Set(ctx context.Context, key string, payload string) error {
	return c.client.Set(ctx, key, payload, c.ttl).Err()
}

func (c *RecommendationCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}
```

### Cache key builder example

```go
package domain

import "fmt"

func TrendingGlobalKey() string {
	return "reco:v1:trending:global"
}

func TrendingCategoryKey(categoryID string) string {
	return fmt.Sprintf("reco:v1:trending:category:%s", categoryID)
}

func SimilarProductKey(productID string) string {
	return fmt.Sprintf("reco:v1:similar:product:%s", productID)
}

func PersonalizedUserKey(userID string) string {
	return fmt.Sprintf("reco:v1:personalized:user:%s", userID)
}
```

---

## 🧠 Data Flow by Recommendation Type

### `trending`

```mermaid
flowchart LR
    Events[Views Clicks Cart Purchases] --> Mongo[user_interactions]
    Mongo --> Job[Trending builder]
    Job --> Sets[recommendation_sets]
    Job --> Cache[Redis trending keys]
    Cache --> API[Recommendation response]
```

Storage choice:

- MongoDB stores interaction history and generated trending set.
- Redis stores top trending list for fast reads.

---

### `similar_products`

```mermaid
flowchart LR
    ProductEvents[Product updates] --> Features[product_features]
    Features --> SimilarJob[Similarity builder]
    SimilarJob --> Sets[recommendation_sets]
    Sets --> Cache[Redis similar product key]
    Cache --> ProductPage[Product detail page]
```

Storage choice:

- MongoDB stores product feature projection and generated similar list.
- Redis stores per-product similar list.

---

### `personalized`

```mermaid
flowchart LR
    UserEvents[User or anonymous behavior] --> Interactions[user_interactions]
    Interactions --> PersonalJob[Personalized builder]
    PersonalJob --> UserSet[recommendation_sets]
    UserSet --> UserCache[Redis personalized key]
    UserCache --> Home[Home feed]
```

Storage choice:

- MongoDB stores user/anonymous behavior history.
- Redis stores latest personalized result briefly.

---

### `frequently_bought_together`

```mermaid
flowchart LR
    Orders[Order/cart events] --> Interactions[user_interactions]
    Interactions --> CoOccurrence[Co-occurrence builder]
    CoOccurrence --> Sets[recommendation_sets]
    Sets --> Cache[Redis fbt key]
    Cache --> Cart[Cart / Product page]
```

Storage choice:

- MongoDB stores purchase/cart signals.
- Redis stores product-level bought-together list.

---

## 📊 Storage Responsibility Matrix

| Data | MongoDB | Redis | Why |
|---|---:|---:|---|
| Raw user interactions | ✅ | ❌ | Durable event history chahiye |
| Generated recommendation sets | ✅ | ✅ | Mongo source, Redis fast copy |
| Global trending top list | ✅ | ✅ | Redis serve, Mongo recover/debug |
| Category trending top list | ✅ | ✅ | Same category pages fast load |
| Similar product list | ✅ | ✅ | Product detail page cacheable |
| Personalized user feed | ✅ | ✅ | Fast home feed, short TTL |
| Temporary rebuild lock | ❌ | ✅ | Redis atomic `SET NX EX` useful |
| A/B assignment | ✅ | Optional | Durable assignment important |

---

## 🧾 Example Read Response

Task 1 me response metadata me `strategy_id` important define hua tha. Task 2 storage design isko preserve karega.

```json
{
  "recommendation_id": "reco_20260518_home_global",
  "strategy_id": "trending_v1_recent_activity",
  "recommendation_type": "trending",
  "items": [
    {
      "product_id": "prod_123",
      "rank": 1,
      "score": 98.5,
      "reason": "high_recent_activity"
    }
  ],
  "source": "redis_cache",
  "ttl_seconds": 900
}
```

### Why `source` useful hai?

- Debugging me pata chalega response Redis se aaya ya MongoDB se.
- Metrics me cache hit/miss track ho sakta hai.
- Performance tuning easy hoti hai.

---

## 📈 Observability Points

Future implementation me ye metrics/logs add hone chahiye:

| Signal | Why |
|---|---|
| `recommendation_cache_hit_total` | Redis cache effectiveness measure karna |
| `recommendation_cache_miss_total` | Mongo fallback frequency samajhna |
| `recommendation_mongo_query_duration_ms` | Slow Mongo reads detect karna |
| `recommendation_cache_set_errors_total` | Redis write failures track karna |
| `recommendation_empty_response_total` | Fallback failures monitor karna |
| `recommendation_set_age_seconds` | Stale generated sets detect karna |

### Log example

```json
{
  "level": "info",
  "service": "recommendation-service",
  "event": "recommendation_cache_hit",
  "cache_key": "reco:v1:trending:global",
  "strategy_id": "trending_v1_recent_activity",
  "ttl_seconds": 842
}
```

---

## 🔒 Security and Privacy Notes

| Rule | Explanation |
|---|---|
| Store IDs, not sensitive PII | Recommendation DB ko email/phone/address jaise raw PII ki zarurat nahi |
| TTL raw interactions | Old detailed behavior forever store nahi karna |
| Separate service DB | Cross-service accidental data access avoid hota hai |
| Do not cache secrets | Redis me sirf recommendation payload/counters/cache keys |
| User deletion support | Future me user privacy request par `user_interactions` cleanup possible hona chahiye |

---

## 🧯 Graceful Degradation Strategy

```mermaid
flowchart TD
    A[Need recommendations] --> B{Redis available?}
    B -->|Yes| C{Cached list exists?}
    C -->|Yes| D[Return cached list]
    C -->|No| E{MongoDB available?}
    B -->|No| E
    E -->|Yes| F[Return MongoDB generated set]
    E -->|No| G[Return empty safe response]
    F --> H[Try cache set async/best effort]
```

### Hinglish explanation

Recommendation Service helpful hai, critical nahi. Agar recommendation layer fail ho jaaye, product page/checkout still work karna chahiye. Frontend empty list receive karke section hide kar sakta hai.

---

## ✅ Validation Checklist

| Check | Status |
|---|---|
| Task 2 scope MongoDB plus Redis tak limited | ✅ Done |
| Recommendation Service folder preserved | ✅ Done |
| `task2.md` created | ✅ Done |
| MongoDB reason documented | ✅ Done |
| Redis reason documented | ✅ Done |
| Collections listed | ✅ Done |
| Index examples included | ✅ Done |
| Cache key strategy included | ✅ Done |
| TTL and invalidation strategy included | ✅ Done |
| External tools/libraries explained | ✅ Done |
| Mermaid diagrams added | ✅ Done |
| No backend implementation beyond Task 2 | ✅ Done |

---

## 🚦 Final Task 2 Decision

Recommendation Service ke liye final storage decision:

```text
Durable recommendation data: MongoDB
Fast recommendation serving: Redis
Primary DB name: recommendation_db
Default cache TTL: 15 minutes
Cache key prefix: reco:v1
```

### Final architecture

```mermaid
flowchart TB
    subgraph Sources[Future Event Sources]
        Product[Product Service]
        Session[Session Service]
        Wishlist[Wishlist Service]
        Order[Order Service]
    end

    Product --> Queue[(Kafka/RabbitMQ)]
    Session --> Queue
    Wishlist --> Queue
    Order --> Queue
    Queue --> Reco[Recommendation Service]

    Reco --> Mongo[(MongoDB recommendation_db)]
    Reco --> Redis[(Redis recommendation cache)]

    Gateway[API Gateway] --> Reco
    ProductPage[Product Detail / Home / Cart UI] --> Gateway

    Redis -. cache miss .-> Mongo
```

---

## 🏁 Conclusion

Recommendation Service Task 2 complete hai. Storage architecture clear ho gaya:

- MongoDB flexible, durable recommendation documents ke liye use hoga.
- Redis low-latency cached top lists and recommendation responses ke liye use hoga.
- Cache key naming, TTL, indexes, failure behavior, and future Go libraries define ho gaye.
- Actual event ingestion, feature store schema, ranking engine, gRPC endpoint, and A/B testing future tasks me implement honge.

**Next task after this:** Recommendation Service - Task 3: Event ingestion.
