# 💖 Wishlist Service - Task 2: Choose MongoDB

![Task](https://img.shields.io/badge/Task-Wishlist%20Service%20%232-blue)
![Priority](https://img.shields.io/badge/Priority-P1-orange)
![Dependency](https://img.shields.io/badge/Dependency-Wishlist%20Model-brightgreen)
![Status](https://img.shields.io/badge/Status-Documented-success)
![Guide](https://img.shields.io/badge/Guide-Hinglish-yellow)
![Scope](https://img.shields.io/badge/Scope-DB%20Choice-purple)

---

## 📌 Task Summary

| Field | Detail |
|---|---|
| Service | Wishlist Service |
| Task | Task 2 - Choose MongoDB |
| Source | `docs/01-micro-tasks.md` -> `Wishlist Service` -> Task 2 |
| Requirement | Wishlist item list flexible aur read-heavy hai, MongoDB simple document model fit hai |
| Dependency | Wishlist Service Task 1: Define wishlist model |
| Priority | P1 |
| Final Decision | **MongoDB as primary Wishlist Service database** |
| Output | Documentation-only database choice guide |

> **Simple Hinglish goal:** Is task ka kaam Wishlist Service ke liye database technology decide karna hai. Task 1 me model decide hua tha: one buyer = one private wishlist. Task 2 me hum decide kar rahe hain ki is model ko store karne ke liye **MongoDB** best fit hai, kyun best fit hai, kaise use hoga, aur future implementation me developer ko kya setup chahiye.

---

## ✅ Final Output Created

```text
TaskImplementation/
└── Wishlist Service/
    ├── task1.md
    └── task2.md
```

### Why this structure?

| Folder/File | Purpose |
|---|---|
| `TaskImplementation/` | Saare task-wise implementation guides ka central folder |
| `Wishlist Service/` | Wishlist service ke task guides ko group karta hai |
| `task1.md` | Wishlist model decision guide |
| `task2.md` | Sirf **Wishlist Service - Task 2** ka MongoDB choice guide |

> 🟢 **Note:** Is task me actual backend code, MongoDB collection migration, repository implementation, API, ya gRPC server create nahi kiya gaya. Ye Task 2 ka database-choice implementation guide hai.

---

## 🧭 Documents Studied

| Document | Kya use kiya gaya |
|---|---|
| `docs/01-micro-tasks.md` | Task 2 ka exact scope, dependency, priority |
| `TaskImplementation/Wishlist Service/task1.md` | Existing model decision: single private wishlist per buyer |
| `docs/04-microservice-design.md` | Wishlist Service purpose, MongoDB direction, expected APIs |
| `docs/05-database-design.md` | Wishlist high-value indexes: user id, product id |
| `database/mongodb-schema-design.md` | Existing `wishlist_db`, `wishlists` collection, example document, indexes |
| `docs/03-folder-structure.md` | Future `backend/services/wishlist-service` folder layout reference |
| `api/master-api.json` | Future wishlist REST/gRPC contract names |

---

## 🧱 Task Boundary

### ✅ Included in Task 2

| Included | Explanation |
|---|---|
| Database choice | Wishlist Service ke liye MongoDB choose karna |
| Reasoning | MongoDB kyun fit hai aur alternatives kyun less suitable hain |
| Access pattern analysis | Wishlist read/write pattern samajhna |
| Tooling guide | MongoDB, MongoDB Go Driver, `mongosh`, Docker Compose ka role explain karna |
| Config examples | Future service ke liye env variable examples |
| High-level data direction | `wishlist_db` aur `wishlists` direction confirm karna |
| Diagrams | Architecture aur flow ko Mermaid me document karna |

### ❌ Not Included in Task 2

| Not Included | Future Task |
|---|---|
| Actual `wishlists` collection design finalize karna | Wishlist Service Task 3 |
| Mongo indexes actually create karna | Wishlist Service Task 3 |
| Add/remove item APIs implement karna | Wishlist Service Task 4 |
| Product validation call implement karna | Wishlist Service Task 4 |
| Move-to-cart flow implement karna | Wishlist Service Task 5 |
| Product availability sync events implement karna | Wishlist Service Task 6 |
| Price drop notifications implement karna | Wishlist Service Task 7 |
| Analytics events publish karna | Wishlist Service Task 8 |

> 🔵 **Reason:** Task 2 ka scope technology decision hai. Actual schema aur code implementation next tasks me cleanly handle honge.

---

## 🪜 Step-by-Step Implementation

## Step 1: Task 1 Model Ko Base Banaya

Task 1 me final model ye decide hua:

```text
One authenticated buyer -> One private wishlist -> Many wishlist items
```

Wishlist item ka primary identifier:

```text
product_id
```

Optional identifier:

```text
variant_id
```

### Iska DB choice pe impact

| Model Property | DB Requirement |
|---|---|
| One user ki one wishlist | `user_id` se fast lookup chahiye |
| Items array ho sakti hai | Embedded list-friendly database chahiye |
| Product snapshots optional hain | Flexible fields chahiye |
| Read-heavy flow hai | Single document read fast hona chahiye |
| Future price/availability snapshots | Schema evolution easy chahiye |

**Conclusion:** Wishlist ka shape naturally ek document jaisa hai, relational rows jaisa nahi. Isliye MongoDB strong candidate hai.

---

## Step 2: Wishlist Access Patterns Identify Kiye

Wishlist Service ka database decision sirf data shape se nahi, usage pattern se bhi decide hota hai.

### Main read patterns

| Query | Example | Frequency |
|---|---|---|
| Get user's wishlist | `GET /api/v1/wishlist` | High |
| Check product in wishlist | Product page heart icon status | High |
| Show wishlist count | Header/profile badge | Medium |
| Filter item by product | Remove/check status | Medium |

### Main write patterns

| Write | Example | Frequency |
|---|---|---|
| Add item | Product detail page se heart click | Medium |
| Remove item | Wishlist page/product page se remove | Medium |
| Update item status | Product out-of-stock event ke baad | Low/async |
| Update price snapshot | Product price change event ke baad | Low/async |

### Access pattern conclusion

```text
Wishlist read-heavy hai.
Most common query user_id se complete wishlist read karna hai.
```

MongoDB me ek user ki wishlist ek document me store karne se full wishlist read simple aur fast hota hai.

---

## Step 3: MongoDB Select Kiya

### Final decision

```text
Wishlist Service primary database = MongoDB
Database name = wishlist_db
Primary collection = wishlists
```

### Why MongoDB?

| Reason | Hinglish Explanation |
|---|---|
| Document model fit | Wishlist naturally ek user document hai jisme items embedded array hota hai |
| Flexible schema | Future me `last_known_price`, `availability`, `source`, `notes`, `priority` jaise fields add karna easy |
| Fast user lookup | `user_id` unique index se user's wishlist quickly fetch hogi |
| Read-heavy optimized | One read me complete wishlist aa sakti hai |
| Microservice ownership | Wishlist Service apna independent Mongo database own karega |
| Existing architecture aligned | Project docs already Wishlist ke liye MongoDB recommend karte hain |
| Future event updates friendly | Product events ke basis par embedded item snapshots update kiye ja sakte hain |

> 🟢 **Final rule:** Wishlist Service kisi dusre service ke database ko directly read/write nahi karega. Product details chahiye to Product Service gRPC/API ya events use honge.

---

## Step 4: Alternatives Compare Kiye

MongoDB choose karne se pehle alternatives ko compare karna zaruri hai.

| Option | Pros | Cons | Decision |
|---|---|---|---|
| MongoDB | Flexible document, embedded items, fast user wishlist reads | Very large arrays ko control karna hoga | ✅ Chosen |
| MySQL | Strong relational constraints, transactions | Wishlist items ke liye joins/tables extra complexity create karenge | ❌ Not chosen |
| PostgreSQL JSONB | Relational + JSON flexibility | Project docs Mongo stack already use kar rahe hain for wishlist-like data | ❌ Not chosen |
| Redis only | Very fast hot reads | Durable source of truth ke liye enough nahi | ❌ Not chosen |
| Elasticsearch/Typesense | Search ke liye strong | Wishlist source-of-truth database nahi | ❌ Not chosen |

### Beginner explanation

Wishlist ek aisi cheez hai jo user ke profile me saved products ki list jaisi behave karti hai. Hume usually poori list ek saath chahiye hoti hai. MongoDB me ye list ek document ke andar natural tareeke se store hoti hai. MySQL me same thing ke liye `wishlists` table + `wishlist_items` table + joins chahiye hote.

---

## Step 5: Database Ownership Rule Define Kiya

Microservice architecture ka golden rule:

```text
Har service apni database own karegi.
Dusri service ke DB ko directly read/write nahi karna.
```

Wishlist Service ke liye:

```text
Wishlist Service -> wishlist_db -> wishlists collection
```

Product Service ke liye:

```text
Product Service -> product_db -> products collection
```

### Important boundary

Wishlist Service `product_db.products` ko direct query nahi karega.

Product validation ke liye future task me ye approach use hogi:

```text
Wishlist Service -> Product Service gRPC/API -> Product validation
```

Ya async sync ke liye:

```text
Product Service -> ProductUpdated event -> Wishlist Service consumer
```

---

## Step 6: High-Level Data Direction Confirm Ki

Task 2 actual collection design nahi banata, lekin project docs me existing direction confirm karta hai.

### Planned database

```text
wishlist_db
```

### Planned primary collection

```text
wishlists
```

### Conceptual document shape

> ⚠️ Ye final schema implementation nahi hai. Final schema Wishlist Service Task 3 me lock hoga.

```json
{
  "_id": "wish_123",
  "user_id": "user_123",
  "visibility": "private",
  "items": [
    {
      "product_id": "prod_123",
      "variant_id": "var_1",
      "added_at": "2026-05-18T00:00:00Z",
      "last_known_price": {
        "amount": 299900,
        "currency": "INR"
      },
      "availability": "in_stock"
    }
  ],
  "created_at": "2026-05-18T00:00:00Z",
  "updated_at": "2026-05-18T00:00:00Z"
}
```

### Why embedded items?

| Design | Reason |
|---|---|
| `items` embedded array | User ki complete wishlist ek document me read hogi |
| `last_known_price` snapshot | UI display aur future price-drop logic ke liye useful |
| `availability` snapshot | Product deleted/out-of-stock case me user ko status dikhega |
| `visibility` field | Task 1 ke private model ke saath future public/shared support possible |

---

## Step 7: Index Direction Note Kiya

Task 2 me indexes actually create nahi honge, but docs ke basis par required index direction clear hai.

### Planned indexes for Task 3

```javascript
db.wishlists.createIndex({ user_id: 1 }, { unique: true })
db.wishlists.createIndex({ "items.product_id": 1 })
```

### Explanation

| Index | Why needed |
|---|---|
| `{ user_id: 1 }` unique | One buyer ke liye one wishlist enforce karne me help |
| `{ "items.product_id": 1 }` | Product wishlist status check aur remove flow fast karne ke liye |

> 🟡 **Scope note:** Ye commands yahan documentation ke liye hain. Actual index creation Wishlist Service Task 3 me hoga.

---

## Step 8: External Libraries / Tools Document Kiye

Task 2 me koi package install nahi kiya gaya, but Wishlist Service ke future implementation ke liye required tools yahan clear kiye gaye.

## 🧰 External Tool 1: MongoDB

| Field | Detail |
|---|---|
| What | Document database |
| Why used | Wishlist document flexible, read-heavy, embedded item list ke liye best fit |
| Used in Task 2? | Decision documented only |
| Future use | `wishlist_db` database and `wishlists` collection |

### Local install options

#### Option A: Docker Compose

Project foundation me local stack ke liye Docker Compose use karne ka direction hai. Future implementation me MongoDB service aise run ho sakta hai:

```yaml
services:
  mongo:
    image: mongo:7
    ports:
      - "27017:27017"
    environment:
      MONGO_INITDB_ROOT_USERNAME: root
      MONGO_INITDB_ROOT_PASSWORD: root
    volumes:
      - mongo_data:/data/db

volumes:
  mongo_data:
```

Run command:

```bash
docker compose up -d mongo
```

#### Option B: Direct MongoDB install

MongoDB Community Server local machine par install kiya ja sakta hai. Future dev command:

```bash
mongod --dbpath ./data/mongo
```

> 🟢 Recommended for this repo: Docker Compose, because project local stack me multiple services rahenge.

---

## 🧰 External Tool 2: MongoDB Go Driver

| Field | Detail |
|---|---|
| What | Official Go driver for MongoDB |
| Why used | Go-based Wishlist Service ko MongoDB se connect karne ke liye |
| Package | `go.mongodb.org/mongo-driver/mongo` |
| Used in Task 2? | Install nahi kiya; future implementation guide me documented |

### Future installation command

Wishlist Service module ke andar future me:

```bash
go get go.mongodb.org/mongo-driver/mongo
```

### Future usage example

> ⚠️ Ye conceptual example hai. Actual repository code Task 3/4 ke baad implement hoga.

```go
package repository

import (
    "context"
    "time"

    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

func NewMongoClient(ctx context.Context, uri string) (*mongo.Client, error) {
    connectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()

    return mongo.Connect(connectCtx, options.Client().ApplyURI(uri))
}
```

### Why official driver?

| Reason | Explanation |
|---|---|
| Maintained by MongoDB | Production support aur compatibility better |
| Context support | Request timeouts/cancellation handle kar sakte hain |
| BSON support | MongoDB document mapping clean hota hai |
| Index API | Future Task 3 me indexes create/check kar sakte hain |

---

## 🧰 External Tool 3: mongosh

| Field | Detail |
|---|---|
| What | MongoDB shell |
| Why used | Local DB inspect, test queries, indexes verify karne ke liye |
| Used in Task 2? | No command run; future developer workflow documented |

### Future usage

```bash
mongosh "mongodb://root:root@localhost:27017"
```

Database select:

```javascript
use wishlist_db
```

Collection inspect:

```javascript
db.wishlists.findOne({ user_id: "user_123" })
```

Indexes inspect:

```javascript
db.wishlists.getIndexes()
```

---

## 🧰 External Tool 4: Docker Compose

| Field | Detail |
|---|---|
| What | Local multi-service runner |
| Why used | MongoDB, Redis, Kafka, API Gateway, services ko local machine par run karne ke liye |
| Used in Task 2? | Existing platform direction referenced only |

### Future use

```bash
docker compose up -d mongo
docker compose logs -f mongo
docker compose down
```

> 🔵 **Note:** Docker Compose file ka actual creation Platform Foundation tasks me hota hai, Wishlist Task 2 me nahi.

---

## Step 9: Future Config Contract Define Kiya

Wishlist Service ko MongoDB se connect karne ke liye future env variables clear hone chahiye.

### Recommended `.env` fields

```env
WISHLIST_SERVICE_PORT=50054
WISHLIST_MONGO_URI=mongodb://root:root@localhost:27017
WISHLIST_MONGO_DATABASE=wishlist_db
WISHLIST_MONGO_COLLECTION=wishlists
WISHLIST_MONGO_CONNECT_TIMEOUT_SECONDS=10
```

### Explanation

| Env var | Purpose |
|---|---|
| `WISHLIST_SERVICE_PORT` | Wishlist gRPC server port |
| `WISHLIST_MONGO_URI` | MongoDB connection string |
| `WISHLIST_MONGO_DATABASE` | Database name |
| `WISHLIST_MONGO_COLLECTION` | Primary collection name |
| `WISHLIST_MONGO_CONNECT_TIMEOUT_SECONDS` | Startup connection timeout |

### Production note

Production me credentials hardcode nahi honge. Kubernetes Secret / cloud secret manager se inject honge.

```text
Local .env -> development only
Kubernetes Secret -> staging/production
```

---

## Step 10: Repository Boundary Plan Kiya

Future code structure `docs/03-folder-structure.md` ke according aise rahega:

```text
backend/
└── services/
    └── wishlist-service/
        ├── cmd/
        │   └── server/
        │       └── main.go
        ├── internal/
        │   ├── domain/
        │   │   └── wishlist.go
        │   ├── usecase/
        │   ├── repository/
        │   │   └── mongo_wishlist_repository.go
        │   └── transport/
        │       └── grpc/
        └── deploy/
```

### How MongoDB will fit later

| Layer | Responsibility |
|---|---|
| `domain/` | Wishlist entity and business rules |
| `usecase/` | Add/remove/get wishlist orchestration |
| `repository/` | MongoDB queries hidden behind interface |
| `transport/grpc/` | gRPC request/response handling |
| `cmd/server/` | Config load, Mongo connection, server start |

### Repository interface idea

> ⚠️ Conceptual only. Actual implementation Task 3/4 me hoga.

```go
package usecase

import (
    "context"
)

type WishlistRepository interface {
    GetByUserID(ctx context.Context, userID string) (*Wishlist, error)
    CreateForUser(ctx context.Context, userID string) (*Wishlist, error)
    ExistsProduct(ctx context.Context, userID string, productID string) (bool, error)
}
```

### Why interface?

| Benefit | Explanation |
|---|---|
| Usecase Mongo se independent | Business logic DB details nahi jaanta |
| Testing easy | Mock repository use ho sakta hai |
| Future DB swap possible | Repository implementation change hogi, usecase stable rahega |

---

## 🏗️ Architecture Diagram

```mermaid
flowchart LR
    Buyer[Buyer App] --> Gateway[API Gateway]
    Gateway -->|REST to gRPC mapping| WishlistService[Wishlist Service]
    WishlistService --> Repo[Wishlist Repository]
    Repo --> Mongo[(MongoDB: wishlist_db)]
    Mongo --> Wishlists[(wishlists collection)]

    WishlistService -. future validation .-> ProductService[Product Service]
    ProductService -. product events .-> EventBus[(Message Queue)]
    EventBus -. future sync .-> WishlistService
```

### Diagram explanation

- Buyer public REST API call karega.
- API Gateway auth check karke Wishlist Service ko call karega.
- Wishlist Service apne repository layer ke through MongoDB use karega.
- Product data direct DB se nahi liya jayega.
- Future me Product Service events se availability/price update ho sakte hain.

---

## 🔁 Data Read Flow

```mermaid
sequenceDiagram
    actor Buyer
    participant Gateway as API Gateway
    participant Wishlist as Wishlist Service
    participant Mongo as MongoDB wishlist_db

    Buyer->>Gateway: GET /api/v1/wishlist
    Gateway->>Gateway: Validate JWT and extract user_id
    Gateway->>Wishlist: GetWishlist(user_id)
    Wishlist->>Mongo: findOne({ user_id })
    Mongo-->>Wishlist: wishlist document
    Wishlist-->>Gateway: Wishlist response
    Gateway-->>Buyer: 200 OK
```

### Read flow explanation

1. Buyer wishlist page open karta hai.
2. Gateway JWT se `user_id` identify karta hai.
3. Wishlist Service `user_id` ke basis par MongoDB se document fetch karta hai.
4. Complete wishlist ek response me return hoti hai.

---

## ✍️ Data Write Flow

```mermaid
sequenceDiagram
    actor Buyer
    participant Gateway as API Gateway
    participant Wishlist as Wishlist Service
    participant Product as Product Service
    participant Mongo as MongoDB wishlist_db

    Buyer->>Gateway: POST /api/v1/wishlist/items
    Gateway->>Wishlist: AddWishlistItem(user_id, product_id)
    Wishlist->>Product: Future: validate product exists
    Product-->>Wishlist: Product valid
    Wishlist->>Mongo: Future: add product to items array
    Mongo-->>Wishlist: Updated wishlist
    Wishlist-->>Gateway: Wishlist response
    Gateway-->>Buyer: 200 OK
```

### Write flow explanation

Task 2 me write code implement nahi hua. Diagram future Task 4 flow ko show karta hai, taaki MongoDB decision clear ho:

- Product validation Product Service se hoga.
- Wishlist update MongoDB me hoga.
- Duplicate product block karna usecase/repository level par hoga.

---

## 🧩 MongoDB Fit With Wishlist Model

```mermaid
classDiagram
    class Wishlist {
        string _id
        string user_id
        string visibility
        WishlistItem[] items
        datetime created_at
        datetime updated_at
    }

    class WishlistItem {
        string product_id
        string variant_id
        datetime added_at
        Money last_known_price
        string availability
    }

    class Money {
        int amount
        string currency
    }

    Wishlist "1" --> "0..many" WishlistItem
    WishlistItem "0..1" --> "1" Money
```

### Why this is document-friendly

MongoDB ek `Wishlist` ko ek document ki tarah store kar sakta hai. `items` array nested hai, aur `last_known_price` nested object hai. Ye shape MongoDB ke BSON document model se naturally match hoti hai.

---

## 🧠 Decision Flow

```mermaid
flowchart TD
    A[Start Task 2] --> B[Read Task 1 model]
    B --> C{Wishlist data flexible hai?}
    C -->|Yes| D{Read-heavy by user_id?}
    D -->|Yes| E{Embedded item list useful hai?}
    E -->|Yes| F[Choose MongoDB]
    F --> G[Document tools and config]
    G --> H[Defer schema to Task 3]
```

---

## 🚦 MongoDB Usage Rules For Wishlist Service

| Rule | Explanation |
|---|---|
| One service owns one DB | Wishlist Service `wishlist_db` own karega |
| No cross-service DB query | Product data direct Product DB se nahi read hogi |
| Use auth context user id | `user_id` request body se trust nahi hoga |
| Keep wishlist item bounded | Wishlist item count limit future business rule me define hoga |
| Use indexes for hot paths | `user_id` and `items.product_id` indexes Task 3 me |
| Store snapshots carefully | Price/availability source of truth Product Service hi rahega |
| Handle duplicate product | Same wishlist me same product duplicate nahi hoga |
| Plan for schema evolution | Optional fields backward compatible hone chahiye |

---

## ⚖️ Data Consistency Notes

Wishlist is not an order or payment record. Iska consistency requirement moderate hai.

| Data | Source of Truth | Wishlist Role |
|---|---|---|
| Product title | Product Service | Optional display snapshot |
| Product price | Product Service | Last known snapshot only |
| Inventory/availability | Product Service | Cached status only |
| Wishlist ownership | Wishlist Service | Source of truth |
| Wishlist item added time | Wishlist Service | Source of truth |

### Important principle

```text
Checkout ke time price/stock fresh validate hoga.
Wishlist ka snapshot final truth nahi hoga.
```

---

## 📈 Scalability Considerations

### Why MongoDB scales well here

| Concern | MongoDB Support |
|---|---|
| Many users | Documents naturally partition by `user_id` |
| Read-heavy access | Single document read is efficient |
| Flexible item metadata | Nested fields easy to add |
| Product event updates | Array element updates possible |
| Future sharding | `user_id` can be considered as shard key if needed |

### Watch-outs

| Risk | Mitigation |
|---|---|
| Very large wishlist arrays | Add max item limit in future business rules |
| Duplicate items | Unique logic in service layer, possible index strategy later |
| Stale price snapshot | Product events and fresh checkout validation |
| Hot users | Cache summary/count later if needed |
| Concurrent add/remove | Use atomic MongoDB updates in future implementation |

---

## 🔒 Security Considerations

MongoDB choice ke saath security rules bhi clear hone chahiye.

| Security Rule | Explanation |
|---|---|
| Authenticated buyer only | Wishlist APIs buyer auth require karenge |
| `user_id` from token | Request body ka `user_id` ignore/avoid hoga |
| Private wishlist only | Task 1 ke according MVP me sharing nahi |
| Least privilege DB user | Wishlist service DB user sirf `wishlist_db` access kare |
| Secrets in env/secret store | Mongo URI hardcode nahi hogi |
| No direct DB exposure | Mongo port production me public nahi hoga |

### Future Mongo user example

> ⚠️ Conceptual command only. Actual DB provisioning DevOps/Foundation tasks me hoga.

```javascript
use wishlist_db

db.createUser({
  user: "wishlist_service",
  pwd: "change-me",
  roles: [
    { role: "readWrite", db: "wishlist_db" }
  ]
})
```

---

## 🧪 Future Verification Checklist

Task 2 me tests run nahi hote kyunki code/schema implement nahi hua. Future tasks me ye checks useful honge.

| Check | Future Task | Expected Result |
|---|---|---|
| Mongo connection works | Task 3 | Service startup Mongo ping pass |
| `wishlist_db` exists | Task 3 | DB reachable |
| `wishlists` collection exists | Task 3 | Collection created or lazy-created |
| `user_id` unique index exists | Task 3 | Duplicate wishlist blocked |
| `items.product_id` index exists | Task 3 | Product status lookup fast |
| Get wishlist by user id | Task 4 | Fast and correct response |
| Duplicate add blocked | Task 4 | Same product duplicate nahi |
| Remove product from wishlist | Task 4 | Item removed correctly |

---

## 🧾 Example Future Commands

These commands are not executed in Task 2, but beginner developers ke liye future workflow clear karte hain.

### Start MongoDB locally

```bash
docker compose up -d mongo
```

### Connect using mongosh

```bash
mongosh "mongodb://root:root@localhost:27017"
```

### Select Wishlist DB

```javascript
use wishlist_db
```

### Insert sample wishlist manually

```javascript
db.wishlists.insertOne({
  _id: "wish_123",
  user_id: "user_123",
  visibility: "private",
  items: [],
  created_at: new Date(),
  updated_at: new Date()
})
```

### Read sample wishlist

```javascript
db.wishlists.findOne({ user_id: "user_123" })
```

### Clean sample data

```javascript
db.wishlists.deleteOne({ user_id: "user_123" })
```

> 🟡 **Note:** Real application me repository layer ye operations karega. Manual commands sirf local debugging ke liye hain.

---

## 📚 Code Examples

## Example 1: Future Go Config Struct

> ⚠️ Conceptual only. Actual file create nahi ki gayi.

```go
package config

type MongoConfig struct {
    URI            string
    Database       string
    Collection     string
    ConnectTimeout int
}
```

### Explanation

- `URI`: MongoDB connection string.
- `Database`: `wishlist_db`.
- `Collection`: `wishlists`.
- `ConnectTimeout`: startup ke time Mongo connect timeout.

---

## Example 2: Future Repository Constructor

```go
package repository

import "go.mongodb.org/mongo-driver/mongo"

type WishlistRepository struct {
    collection *mongo.Collection
}

func NewWishlistRepository(db *mongo.Database, collectionName string) *WishlistRepository {
    return &WishlistRepository{
        collection: db.Collection(collectionName),
    }
}
```

### Explanation

Repository ko direct URI nahi diya jayega. App startup Mongo client banayega, database select karega, aur repository ko collection dega. Isse testing aur dependency injection clean rahega.

---

## Example 3: Future Find Query

```go
filter := bson.M{
    "user_id": userID,
}

err := collection.FindOne(ctx, filter).Decode(&wishlist)
```

### Explanation

Wishlist read ka hot path simple hai:

```text
user_id -> wishlist document
```

Isi reason se `{ user_id: 1 }` index important hai.

---

## 🗂️ Clean Folder Structure Of This Implementation

Task 2 ke liye sirf documentation output create hua:

```text
TaskImplementation/
└── Wishlist Service/
    ├── task1.md
    └── task2.md
```

Future actual service folder expected structure:

```text
backend/
└── services/
    └── wishlist-service/
        ├── cmd/
        │   └── server/
        │       └── main.go
        ├── internal/
        │   ├── domain/
        │   │   └── wishlist.go
        │   ├── usecase/
        │   │   ├── get_wishlist.go
        │   │   ├── add_item.go
        │   │   └── remove_item.go
        │   ├── repository/
        │   │   └── mongo_wishlist_repository.go
        │   └── transport/
        │       └── grpc/
        └── deploy/
```

> 🔴 **Important:** Future folder structure yahan reference ke liye hai. Task 2 me `backend/services/wishlist-service` create nahi kiya gaya.

---

## ✅ Final Decision Record

| Decision | Value |
|---|---|
| Database | MongoDB |
| DB name | `wishlist_db` |
| Primary collection | `wishlists` |
| Main lookup key | `user_id` |
| Main nested field | `items` |
| Item key | `product_id` |
| DB ownership | Wishlist Service owns Wishlist DB |
| Cross-service access | gRPC/events only, no direct DB access |
| Schema implementation | Deferred to Task 3 |
| API implementation | Deferred to Task 4+ |

---

## ✅ Completion Checklist

| Requirement | Status |
|---|---:|
| `TaskImplementation/` folder available | ✅ |
| `TaskImplementation/Wishlist Service/` folder kept | ✅ |
| `task2.md` created | ✅ |
| Step-by-step implementation in Hinglish | ✅ |
| MongoDB decision clearly explained | ✅ |
| External tools/libraries documented | ✅ |
| Install/use examples included | ✅ |
| Folder structure included | ✅ |
| Code examples included | ✅ |
| Mermaid diagrams included | ✅ |
| Nothing beyond Wishlist Service Task 2 implemented | ✅ |

---

## 🎯 Final Summary

Wishlist Service Task 2 ka final outcome:

```text
Wishlist Service ke liye MongoDB choose kiya gaya hai.
```

Reason simple hai:

```text
Wishlist ek read-heavy, flexible, user-owned document hai.
MongoDB is model ko naturally support karta hai.
```

Task 2 complete hai as a documentation-first database decision. Next logical task Wishlist Service Task 3 hoga: `wishlists` collection design with user id, items, visibility, and timestamps.
