# 🛍️ Product Service - Task 2: Choose MongoDB

![Task](https://img.shields.io/badge/Task-Product%20Service%20%232-blue)
![Priority](https://img.shields.io/badge/Priority-P0-red)
![Dependency](https://img.shields.io/badge/Dependency-Catalog%20Model-brightgreen)
![Status](https://img.shields.io/badge/Status-Documented-success)
![Guide](https://img.shields.io/badge/Guide-Hinglish-orange)
![Scope](https://img.shields.io/badge/Scope-Database%20Choice-purple)

## 📌 Task Summary

| Field | Detail |
|---|---|
| Service | Product Service |
| Task No. | 2 |
| Task Name | Choose MongoDB |
| Source | `docs/01-micro-tasks.md` -> `Product Service` -> Task 2 |
| Goal | Product attributes category-wise dynamic hote hain, isliye Product Service ke primary catalog store ke liye MongoDB choose karna |
| Dependency | Product Service Task 1: Define catalog model |
| Priority | P0 |
| Output Type | Documentation-only implementation guide |
| Final Decision | Product Service ka primary database **MongoDB** hoga |
| Not Included | Collections create karna, real Go repository code, migrations, seller CRUD, read APIs, inventory operations, search events |

> 🟢 **Simple Hinglish goal:** Is task me Product Service ke liye database technology choose ki gayi hai. Task 1 ne product/variant/category/attribute model define kiya tha. Us model me category-wise dynamic attributes, nested variants, images, prices, aur inventory fields hain. In requirements ke basis par Product Service ke catalog data ke liye MongoDB best fit hai.

---

## ✅ Final Output Created

```text
TaskImplementation/
└── Product Service/
    ├── task1.md
    └── task2.md
```

### Why this structure?

| Path | Purpose |
|---|---|
| `TaskImplementation/` | Saare task-wise implementation guides ka central folder |
| `TaskImplementation/Product Service/` | Product Service ke implementation guides ko group karne ke liye |
| `task1.md` | Product catalog model decision guide |
| `task2.md` | MongoDB database choice decision guide |

> 🔵 **Important:** Is task me sirf `task2.md` create kiya gaya hai. Product Service ka actual Go code, MongoDB collections, Docker Compose file, ya API endpoints create nahi kiye gaye.

---

## 🧭 Docs Studied Before Implementation

| Document | Kya use kiya gaya |
|---|---|
| `docs/01-micro-tasks.md` | Product Service Task 2 ka exact scope, dependency, priority identify kiya |
| `TaskImplementation/Product Service/task1.md` | Catalog model, variants, dynamic attributes, image metadata, inventory rules samjhe |
| `docs/02-system-architecture.md` | Product Service DB ownership, API Gateway, Search Service, Cart/Order relationship samjha |
| `docs/03-folder-structure.md` | Future Product Service folder and config shape align kiya |
| `docs/04-microservice-design.md` | Product Service responsibilities and MongoDB choice validate ki |
| `docs/05-database-design.md` | Service-wise DB strategy confirm ki: Product -> MongoDB |
| `database/mongodb-schema-design.md` | Future `product_db` and product-related collection direction samjha |
| `docs/11-devops-external-services.md` | Local Docker Compose dependency expectations samjhi |
| `docs/13-developer-guide.md` | Backend coding flow, config, testing, and production readiness rules align kiye |
| `api/master-api.json` | Product APIs and schemas ka high-level shape validate kiya |

---

## 🧱 Task Boundary

### ✅ Task 2 me kya implement/document kiya gaya

- MongoDB ko Product Service ka primary catalog database choose kiya
- MongoDB choice ke reasons explain kiye
- MySQL/PostgreSQL/Redis/Typesense ke against comparison diya
- Product Service database ownership rules define kiye
- Future local MongoDB setup guidance diya
- Future Go driver usage ka example diya
- Environment variables ka recommended shape diya
- Future repository integration ka clean folder structure diya
- Security, scalability, indexing, and validation notes document kiye
- Mermaid architecture and flow diagrams add kiye

### ❌ Task 2 me kya implement nahi kiya gaya

| Item | Reason |
|---|---|
| `products`, `categories`, `brands`, `inventory_snapshots`, `price_books` collections create karna | Product Service Task 3 ka scope hai |
| Real MongoDB server start karna | Platform Foundation Docker Compose task / local setup ka scope hai |
| Product Service Go folder create karna | Later Product Service implementation tasks ka scope hai |
| Mongo repository code likhna | Collections and service implementation ke baad hoga |
| Seller CRUD APIs | Product Service Task 4 ka scope hai |
| Product listing/detail APIs | Product Service Task 5 ka scope hai |
| Inventory reserve/release/commit code | Product Service Task 6 ka scope hai |
| Product search events | Product Service Task 7 ka scope hai |
| Media/CDN metadata implementation | Product Service Task 8 ka scope hai |

> 🔴 **Golden rule:** Task 2 ka kaam database choose karna aur us decision ko implementation-ready banana hai. Actual collections design Task 3 me aayega.

---

## 🏁 Final Decision

Product Service ke liye primary database:

```text
Database technology: MongoDB
Database name: product_db
Service owner: product-service
Access pattern: Product Service ke through only
Search engine: Typesense, MongoDB nahi
Cache: Redis optional future read cache, primary store nahi
```

### Decision in one line

> 🟢 Product catalog flexible, nested, category-specific, and read-heavy hai. Isliye MongoDB Product Service ke liye best primary datastore hai.

---

## 🧩 High-Level Architecture

```mermaid
flowchart TB
    Seller[Seller Dashboard] --> Gateway[API Gateway]
    Buyer[User App] --> Gateway

    Gateway --> Product[Product Service]
    Cart[Cart Service] --> Product
    Order[Order Service] --> Product
    CMS[CMS Service] --> Product

    Product --> Mongo[(MongoDB<br/>product_db)]

    Product -. ProductUpdated event<br/>future Task 7 .-> MQ[Kafka or RabbitMQ]
    MQ -. consume .-> Search[Search Service]
    Search --> Typesense[(Typesense)]

    Product -. never direct write .-> Typesense
    Cart -. no direct DB access .-> Mongo
    Order -. no direct DB access .-> Mongo
```

**Explanation:**  
Frontend REST call API Gateway pe aayegi. Gateway internal gRPC se Product Service ko call karega. Product Service hi `product_db` ko read/write karega. Cart, Order, Search, CMS kisi bhi case me Product MongoDB ko directly access nahi karenge.

---

## 🪜 Step-by-Step Implementation

## Step 1: Task requirement identify kiya

`docs/01-micro-tasks.md` me Product Service Task 2 ye define hai:

```text
Choose MongoDB:
Product attributes category-wise dynamic hote hain,
isliye MongoDB flexible schema ke liye better hai.
```

**How this part was built:**  
Pehle task source se exact requirement nikali gayi. Is task ka focus code likhna nahi, balki DB decision ko clear, justified, and future-ready banana hai.

---

## Step 2: Task 1 catalog model ko base banaya

Task 1 me Product model roughly aisa define hua tha:

```text
Product
├── product_id
├── seller_id
├── title
├── description
├── brand
├── category_id
├── status
├── attributes
├── images[]
├── variants[]
├── rating_summary
└── timestamps
```

Is model me important data characteristics:

| Characteristic | Meaning |
|---|---|
| Dynamic attributes | Shoes, phones, grocery, furniture sabke attributes alag honge |
| Nested variants | Same product ke multiple SKU, size, color, price, stock fields honge |
| Embedded images | Product detail page ko gallery metadata saath me chahiye |
| Read-heavy catalog | Product detail and listing frequently read honge |
| Seller-owned writes | Seller dashboard se product drafts and updates honge |
| Search projection | Search Service indexed copy rakhega, but canonical data Product Service me rahega |

**Hinglish explanation:**  
Agar hum Product ko rigid SQL columns me force karenge, to har category ke liye ya to new columns add karne padenge ya ugly JSON/EAV pattern banana padega. MongoDB me product document natural shape me store ho sakta hai.

---

## Step 3: Product Service DB ownership define ki

Microservice architecture ka rule:

```text
Har service apni database own karegi.
Dusri service direct DB read/write nahi karegi.
```

Product Service ownership:

| Area | Owner |
|---|---|
| Product catalog source of truth | Product Service |
| Product MongoDB database | Product Service |
| Product search index | Search Service |
| Cart product snapshot | Cart Service |
| Order item snapshot | Order Service |
| Product moderation workflow | CMS/Superadmin + Product Service APIs |

### Ownership rule

```text
Allowed:
API Gateway -> Product Service -> MongoDB
Order Service -> Product Service -> MongoDB
Search Service -> Product Service gRPC or product events

Not allowed:
Order Service -> Product MongoDB directly
Cart Service -> Product MongoDB directly
Search Service -> Product MongoDB directly in normal flow
```

**How this part was built:**  
`docs/02-system-architecture.md` and `docs/04-microservice-design.md` dono direct cross-service DB access ban karte hain. Isliye MongoDB choose karne ke saath access boundary bhi define ki gayi.

---

## Step 4: Database options compare kiye

| Option | Fit | Pros | Cons for Product Catalog |
|---|---:|---|---|
| MongoDB | ✅ Best | Flexible schema, nested documents, dynamic attributes, fast product detail reads | Joins weak hain; validation service layer me carefully karni padegi |
| PostgreSQL/MySQL | ⚠️ Good but rigid | Transactions, relational constraints, reporting queries strong | Dynamic attributes ke liye JSON/EAV pattern complex ho jayega |
| Redis | ❌ Not primary DB | Very fast cache | Durable catalog source of truth ke liye suitable nahi |
| Typesense | ❌ Search only | Full-text search, facets, typo tolerance | Source of truth DB nahi; product writes and lifecycle ke liye suitable nahi |
| Elasticsearch/OpenSearch | ⚠️ Search/index | Advanced search and analytics | Primary catalog DB ke liye operationally heavy |

### Decision matrix

| Requirement | MongoDB | SQL DB | Redis | Typesense |
|---|---:|---:|---:|---:|
| Category-wise dynamic attributes | ✅ Strong | ⚠️ JSON workaround | ❌ | ⚠️ Index copy only |
| Product with embedded variants/images | ✅ Strong | ⚠️ Multiple joins | ❌ | ⚠️ Read model only |
| Product detail self-contained read | ✅ Strong | ⚠️ Join-heavy | ✅ Cache only | ✅ Search result only |
| Seller CRUD source of truth | ✅ Good | ✅ Good | ❌ | ❌ |
| Full-text search and facets | ⚠️ Basic fallback | ⚠️ Basic | ❌ | ✅ Strong |
| Financial transaction consistency | ⚠️ Not main use case | ✅ Strong | ❌ | ❌ |

**Hinglish explanation:**  
SQL databases bad nahi hain. Orders, payments, auth jaise structured and transactional domains ke liye SQL better hai. But Product catalog ka shape dynamic hai, isliye MongoDB yahan natural fit hai.

---

## Step 5: MongoDB choose karne ke exact reasons finalize kiye

### 1. Flexible schema

Different category examples:

```json
{
  "category": "shoes",
  "attributes": {
    "size": "9",
    "color": "black",
    "material": "mesh",
    "sport_type": "running"
  }
}
```

```json
{
  "category": "mobile_phone",
  "attributes": {
    "ram": "8GB",
    "storage": "128GB",
    "processor": "Snapdragon",
    "battery_mah": 5000
  }
}
```

MongoDB me dono same `products` collection me cleanly fit ho sakte hain. SQL me ya to many nullable columns banenge, ya JSON columns, ya attribute tables ka complex EAV model.

### 2. Product document naturally nested hota hai

Product ke andar variants and images naturally child objects hain:

```json
{
  "_id": "prod_123",
  "title": "Running Shoes",
  "images": [
    {
      "url": "https://cdn.example.com/prod_123/main.jpg",
      "alt": "Running shoes side view",
      "position": 1
    }
  ],
  "variants": [
    {
      "variant_id": "var_1",
      "sku": "ACME-SHOE-9-BLK",
      "attributes": {
        "size": "9",
        "color": "black"
      },
      "price": {
        "amount": 299900,
        "currency": "INR"
      },
      "stock_quantity": 120,
      "reserved_quantity": 5
    }
  ]
}
```

**Note:** Ye example future collection design ko explain karta hai. Actual collection creation Task 3 me hoga.

### 3. Product detail page ko self-contained read chahiye

Product detail API ko mostly ye sab saath me chahiye:

- Title
- Description
- Brand
- Category
- Images
- Variants
- Price
- Availability
- Rating summary

MongoDB document model me one product detail read usually one document se complete ho sakta hai.

### 4. Catalog evolution easy hoti hai

Aaj attributes simple ho sakte hain:

```json
{
  "material": "cotton"
}
```

Kal category schema richer ho sakta hai:

```json
{
  "material": {
    "value": "cotton",
    "display_label": "100% Cotton",
    "source": "seller_input"
  }
}
```

MongoDB flexible schema future evolution ko easier banata hai, as long as service validation strong ho.

---

## Step 6: MongoDB ka role clearly define kiya

MongoDB Product Service me ye karega:

| Responsibility | Detail |
|---|---|
| Source of truth | Product canonical data store karega |
| Flexible catalog | Category-wise different attributes support karega |
| Product detail read | Self-contained product document read karega |
| Seller draft workflow | Draft/published/unpublished status store karega |
| Variant data | SKU, price, variant attributes, stock fields store karega |
| Future collection base | `products`, `categories`, `brands`, `inventory_snapshots`, `price_books` ke liye DB foundation dega |

MongoDB ye nahi karega:

| Not MongoDB Responsibility | Actual Owner |
|---|---|
| Full search ranking, typo tolerance, facets | Search Service + Typesense |
| Cart hot cache | Cart Service + Redis |
| Payment/order transactions | Order/Payment Service + MySQL |
| Cross-service analytics warehouse | Future analytics/read model |
| CDN image binary storage | Object storage/CDN |

---

## Step 7: Product database naming decide ki

Recommended database naming:

```text
product_db
```

Recommended future collection names from docs:

```text
products
categories
brands
inventory_snapshots
price_books
```

### Naming rules

| Item | Rule | Example |
|---|---|---|
| Database | snake_case service domain | `product_db` |
| Collection | lowercase plural noun | `products` |
| Document id | stable domain id | `prod_123` |
| Product id field | API-friendly string id | `product_id` or `_id` mapped carefully |
| Date fields | UTC timestamp | `created_at`, `updated_at` |

> 🟡 **Task 3 handoff:** Collection fields, indexes, validation schema, and exact `_id` mapping will be finalized in Product Service Task 3.

---

## Step 8: Future local setup plan document kiya

Task 2 me MongoDB run nahi kiya gaya, but future local setup ke liye recommended Docker Compose shape ye hai:

```yaml
services:
  product-mongo:
    image: mongo:8
    container_name: ecommerce-product-mongo
    restart: unless-stopped
    ports:
      - "27017:27017"
    environment:
      MONGO_INITDB_ROOT_USERNAME: root
      MONGO_INITDB_ROOT_PASSWORD: root_password
      MONGO_INITDB_DATABASE: product_db
    volumes:
      - product_mongo_data:/data/db

volumes:
  product_mongo_data:
```

### How to use in future

```bash
docker compose -f infra/compose/docker-compose.local.yml up -d product-mongo
```

```bash
docker exec -it ecommerce-product-mongo mongosh \
  -u root \
  -p root_password \
  --authenticationDatabase admin \
  product_db
```

### Why Docker?

| Reason | Benefit |
|---|---|
| Same setup for all developers | "Works on my machine" issues kam honge |
| Easy cleanup | Volumes remove karke local DB reset kar sakte hain |
| No manual Mongo install | New developer faster onboard hoga |
| CI-friendly | Integration tests me same image use ho sakti hai |

> 🔵 **Note:** Official `mongo` Docker image root user env vars support karta hai. Production me root credentials app ke liye use nahi karne chahiye; app-specific least-privilege user use hoga.

---

## Step 9: Future Product Service config shape define ki

Product Service ko MongoDB se connect karne ke liye future env variables:

```text
SERVICE_NAME=product-service
SERVICE_PORT=8080
GRPC_PORT=9090

MONGO_URI=mongodb://product_app:product_dev_password@product-mongo:27017/product_db?authSource=product_db
MONGO_DATABASE=product_db
MONGO_CONNECT_TIMEOUT=5s
MONGO_MAX_POOL_SIZE=50
MONGO_MIN_POOL_SIZE=5
MONGO_SERVER_SELECTION_TIMEOUT=5s
```

### Config rules

| Rule | Why |
|---|---|
| `MONGO_URI` secret/config se load karo | Credentials code me hardcode nahi honge |
| `MONGO_DATABASE` explicit rakho | Wrong DB writes avoid honge |
| Timeouts mandatory rakho | Service startup hang nahi karega |
| Pool size controlled rakho | MongoDB connection overload avoid hoga |
| Local `.env` git me commit mat karo | Secret leakage avoid hoga |

---

## Step 10: Future Go driver choose kiya

Product Service Go me hoga, isliye future MongoDB integration ke liye official MongoDB Go Driver use hoga.

| Tool/Library | What it is | Why used | Install |
|---|---|---|---|
| Official MongoDB Go Driver | Go applications ke liye official MongoDB client library | MongoDB connect, query, update, indexes, BSON mapping ke liye | `go get go.mongodb.org/mongo-driver/v2/mongo` |

### Future install command

```bash
cd backend/services/product-service
go get go.mongodb.org/mongo-driver/v2/mongo
```

### Future connection example

```go
package repository

import (
    "context"
    "time"

    "go.mongodb.org/mongo-driver/v2/mongo"
    "go.mongodb.org/mongo-driver/v2/mongo/options"
)

func NewMongoClient(ctx context.Context, uri string) (*mongo.Client, error) {
    connectCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    client, err := mongo.Connect(options.Client().ApplyURI(uri))
    if err != nil {
        return nil, err
    }

    if err := client.Ping(connectCtx, nil); err != nil {
        return nil, err
    }

    return client, nil
}
```

**Hinglish explanation:**  
Ye code abhi repo me add nahi hua. Ye future implementation example hai. Actual repository implementation Task 3/4 ke baad create hoga, jab collections and usecases finalized honge.

---

## Step 11: Future repository boundary sketch ki

Clean architecture me Product Service ka domain MongoDB driver ko directly import nahi karega.

```mermaid
flowchart LR
    Transport[gRPC Handler] --> Usecase[Usecase Layer]
    Usecase --> RepoInterface[Repository Interface]
    RepoInterface --> MongoRepo[Mongo Repository Implementation]
    MongoRepo --> Mongo[(product_db)]

    Domain[Domain Model] --> Usecase
    Domain -. no Mongo import .-> MongoRepo
```

### Future interface example

```go
package domain

import "context"

type ProductRepository interface {
    Save(ctx context.Context, product *Product) error
    FindByID(ctx context.Context, productID string) (*Product, error)
    ListBySeller(ctx context.Context, sellerID string, filter ProductFilter) ([]Product, error)
}
```

### Future implementation location

```text
backend/
└── services/
    └── product-service/
        └── internal/
            └── repository/
                ├── mongo_client.go
                ├── mongo_product_repository.go
                ├── mongo_category_repository.go
                └── mongo_inventory_repository.go
```

> 🔵 **Reminder:** Upar wali files future implementation guide hain. Task 2 me ye files create nahi ki gayi.

---

## Step 12: MongoDB query patterns identify kiye

Product Service ke future query patterns:

| Query Pattern | Example | Index Need |
|---|---|---|
| Product detail | Get product by `product_id` | `_id` or `product_id` unique |
| Seller catalog | List seller products by status | `seller_id + status + updated_at` |
| Category browse | List published products in category | `category_id + status + updated_at` |
| SKU lookup | Find variant by SKU | `variants.sku` unique |
| Admin moderation | List pending/draft products | `status + updated_at` |

### Future index examples

```javascript
db.products.createIndex({ seller_id: 1, status: 1, updated_at: -1 })
db.products.createIndex({ category_id: 1, status: 1, updated_at: -1 })
db.products.createIndex({ "variants.sku": 1 }, { unique: true })
db.products.createIndex({ title: "text", description: "text", brand: "text" })
```

**Important:**  
Ye indexes `database/mongodb-schema-design.md` me already suggested hain. Actual index creation and final index review Product Service Task 3 me hoga.

---

## Step 13: Inventory consistency approach document ki

Inventory product catalog ka sensitive part hai, kyunki checkout me race conditions aa sakti hain.

MongoDB future me atomic update support karega:

```javascript
db.products.updateOne(
  {
    _id: "prod_123",
    "variants.variant_id": "var_1",
    "variants.stock_quantity": { $gte: 2 }
  },
  {
    $inc: {
      "variants.$.reserved_quantity": 2
    }
  }
)
```

### Why this matters?

| Point | Explanation |
|---|---|
| Atomic update | Same variant stock race condition reduce hoti hai |
| Conditional filter | Stock available tabhi reserve hoga |
| Idempotency still needed | Retry se double reserve avoid karna hoga |
| TTL reservation needed | Payment fail/timeout pe reserved stock release hoga |

> 🔴 **Out of scope:** Actual `ReserveInventory`, `ReleaseInventory`, and `CommitInventory` implementation Product Service Task 6 me hoga.

---

## Step 14: Search boundary clarify ki

MongoDB primary catalog DB hai, but search engine nahi.

```mermaid
sequenceDiagram
    participant Seller as Seller Dashboard
    participant GW as API Gateway
    participant Product as Product Service
    participant Mongo as MongoDB product_db
    participant MQ as Kafka/RabbitMQ
    participant Search as Search Service
    participant Typesense as Typesense

    Seller->>GW: Update product
    GW->>Product: gRPC UpdateProduct
    Product->>Mongo: Save canonical product
    Product-->>GW: Product response
    Product-->>MQ: ProductUpdated event (future)
    MQ-->>Search: Consume event (future)
    Search->>Typesense: Upsert search document
```

### Search rule

| Concern | Owner |
|---|---|
| Canonical product data | Product Service + MongoDB |
| Fast text search | Search Service + Typesense |
| Search facets/sort/ranking | Typesense |
| Search data freshness | Product events |
| Search fallback/dev text index | MongoDB text index only if needed |

**Hinglish explanation:**  
User jab search karega to Typesense fast result dega. But product detail ka final truth Product Service se hi aayega. MongoDB text index sirf fallback/dev convenience ke liye socha ja sakta hai.

---

## Step 15: Security and production rules add kiye

| Rule | Why |
|---|---|
| App-specific MongoDB user use karo | Root user app runtime me unsafe hai |
| Least privilege role use karo | Product Service ko sirf product DB access mile |
| TLS enable karo in production | Network traffic secure rahe |
| Secrets env/config manager se load karo | Password code ya markdown me leak nahi honge |
| Backups enable karo | Catalog loss business-critical hai |
| Indexes monitor karo | Slow product listing avoid hogi |
| Query timeouts use karo | Service hang avoid hoga |
| Document size watch karo | Product document extremely large nahi hona chahiye |

### Production URI shape

```text
mongodb+srv://product_app:<password>@<cluster-host>/product_db?retryWrites=true&w=majority
```

### Local dev URI shape

```text
mongodb://product_app:product_dev_password@product-mongo:27017/product_db?authSource=product_db
```

> 🟡 **Beginner note:** `mongodb+srv://` usually managed MongoDB/Atlas style connection ke liye use hota hai. Local Docker me normally `mongodb://host:27017/db` use hota hai.

---

## Step 16: MongoDB limitations document kiye

MongoDB choose karna perfect magic solution nahi hai. Ye limitations consciously handle karni hongi:

| Limitation | Mitigation |
|---|---|
| Schema flexible hai, invalid data ka risk hota hai | Service-level validation + future collection JSON schema |
| Joins natural nahi hain | Product detail self-contained rakho, cross-service data snapshots use karo |
| Large documents slow ho sakte hain | Images/variants bounded rakho, media binary DB me store mat karo |
| Multi-document transaction heavier hoti hai | Product write model ko single aggregate-focused rakho |
| Text search advanced nahi | Typesense ko search owner rakho |
| Analytics scans expensive ho sakte hain | Events/read models/analytics pipeline later use karo |

**Hinglish explanation:**  
MongoDB flexible hai, but discipline chahiye. Agar validation weak hogi to collection me inconsistent data aa sakta hai. Isliye Product Service domain validation mandatory rahegi.

---

## 🧬 Decision Flow Diagram

```mermaid
flowchart TD
    A[Product Service data shape] --> B{Attributes fixed hain?}
    B -->|No, category-wise dynamic| C[Flexible schema needed]
    B -->|Yes| D[SQL can work]

    C --> E{Nested variants/images needed?}
    E -->|Yes| F[Document model fits well]
    E -->|No| D

    F --> G{Primary need search ranking hai?}
    G -->|No, canonical catalog store chahiye| H[Choose MongoDB]
    G -->|Yes| I[Use Typesense for Search Service]

    H --> J[Product Service owns product_db]
    I --> K[Search Service owns Typesense index]
```

---

## 🗂️ Clean Future Folder Structure

Task 2 me actual Product Service code create nahi hua, but future MongoDB integration ke liye recommended folder structure:

```text
backend/
└── services/
    └── product-service/
        ├── cmd/
        │   └── server/
        │       └── main.go
        ├── internal/
        │   ├── config/
        │   │   └── config.go
        │   ├── domain/
        │   │   ├── product.go
        │   │   ├── variant.go
        │   │   ├── category.go
        │   │   └── inventory.go
        │   ├── usecase/
        │   │   ├── create_product.go
        │   │   ├── publish_product.go
        │   │   └── list_products.go
        │   ├── repository/
        │   │   ├── mongo_client.go
        │   │   ├── mongo_product_repository.go
        │   │   ├── mongo_category_repository.go
        │   │   └── mongo_inventory_repository.go
        │   └── transport/
        │       └── grpc/
        │           └── server.go
        └── deploy/
            └── Dockerfile
```

### Real created structure for this task

```text
TaskImplementation/
└── Product Service/
    └── task2.md
```

> 🔵 **Reminder:** Future structure sirf guide hai. Is task ka actual created output sirf documentation file hai.

---

## 🔌 External Libraries / Tools Used

### Installed in this task

| Library/Tool | Installed? | Why |
|---|---:|---|
| MongoDB server | No | Task 2 documentation-only hai |
| MongoDB Go Driver | No | Future Product Service repository implementation me use hoga |
| Docker Compose file | No | Platform Foundation local stack scope hai |
| Go packages | No | Runtime code implement nahi hua |
| NPM packages | No | Frontend work nahi hai |

### Selected/recommended for future implementation

| Tool | What it is | Why used | Install/use |
|---|---|---|---|
| MongoDB | Document database | Dynamic product attributes and nested catalog documents ke liye | Local: Docker `mongo` image; Production: managed MongoDB/Atlas or self-managed cluster |
| Official `mongo` Docker image | MongoDB container image | Local dev stack quickly run karne ke liye | `docker compose up -d product-mongo` |
| Official MongoDB Go Driver | Go client library | Product Service se MongoDB connect/query/update karne ke liye | `go get go.mongodb.org/mongo-driver/v2/mongo` |
| `mongosh` | MongoDB shell | Local DB inspect/test karne ke liye | Docker container ke andar available hota hai, ya separately install kar sakte ho |
| Markdown | Documentation format | `task2.md` readable and repo-friendly banane ke liye | Install nahi chahiye |
| Mermaid | Markdown diagram syntax | Architecture and flow diagrams ke liye | GitHub/VS Code preview support |
| Shields.io badges | Badge image service | Task metadata visually clear dikhane ke liye | Markdown image URL se use hota hai |

### Helpful official references

| Reference | URL |
|---|---|
| MongoDB Go Driver docs | `https://www.mongodb.com/docs/drivers/go/current/` |
| MongoDB Go Driver package | `https://pkg.go.dev/go.mongodb.org/mongo-driver/v2/mongo` |
| Official Mongo Docker image | `https://hub.docker.com/_/mongo/` |

---

## 🧪 Beginner-Friendly Verification Checklist

Task 2 complete hai agar:

| Check | Status |
|---|---:|
| MongoDB final database choice documented hai | ✅ |
| Product catalog dynamic attributes ka reason clear hai | ✅ |
| SQL/Redis/Typesense comparison diya gaya hai | ✅ |
| Product Service DB ownership rule clear hai | ✅ |
| `product_db` database naming defined hai | ✅ |
| Future MongoDB local setup example included hai | ✅ |
| Future Go driver install/use example included hai | ✅ |
| Future repository folder structure included hai | ✅ |
| Architecture and flow Mermaid diagrams included hain | ✅ |
| External libraries/tools mention kiye gaye hain | ✅ |
| Task 3+ implementation avoid kiya gaya hai | ✅ |

---

## 🚫 Out of Scope for Product Service Task 2

```text
Do not create:
- backend/services/product-service/
- MongoDB collections
- MongoDB indexes
- Product Service Go repository files
- Product CRUD APIs
- Product read APIs
- Inventory operations
- Product search event publisher
- Docker Compose Mongo service file
- Kubernetes Mongo manifests
```

**Reason:**  
Ye sab later Product Service tasks ya Platform Foundation tasks me covered honge. Task 2 ka purpose database technology decision ko strong, understandable, and implementation-ready banana hai.

---

## 🧾 Final Task 2 Summary

Product Service Task 2 complete hai as a **MongoDB selection guide**:

- Product Service ke liye MongoDB choose kiya gaya.
- Reason: category-wise dynamic attributes, nested variants/images, flexible catalog evolution.
- `product_db` Product Service ka owned database hoga.
- Search ke liye Typesense separate rahega; MongoDB source of truth rahega.
- Future Go integration ke liye official MongoDB Go Driver recommend kiya gaya.
- Future local setup ke liye Docker-based MongoDB guidance diya gaya.
- Actual collections, indexes, and repository implementation intentionally Task 3+ ke liye leave kiye gaye.

> 🟢 **Final note:** Ab Product Service Task 3 me MongoDB collections design ki ja sakti hain: `products`, `categories`, `brands`, `inventory_snapshots`, and `price_books`.
