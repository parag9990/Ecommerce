# 🧩 CMS Service - Task 4: Coupon Engine MVP

![Task](https://img.shields.io/badge/Task-CMS%20Service%20%234-blue)
![Priority](https://img.shields.io/badge/Priority-P1-orange)
![Dependency](https://img.shields.io/badge/Dependency-Cart%20%26%20Order-purple)
![Engine](https://img.shields.io/badge/Engine-Coupon%20MVP-2ea44f)
![Database](https://img.shields.io/badge/Database-MySQL%208%2B-4479A1)
![Status](https://img.shields.io/badge/Status-Documented-success)
![Guide](https://img.shields.io/badge/Guide-Hinglish-yellow)

## 📌 Task Summary

| Field | Detail |
|---|---|
| Task Name | Coupon engine MVP |
| Source | `docs/01-micro-tasks.md` -> `CMS Service` -> Task 4 |
| Goal | Fixed, percentage, min cart, category/product/seller scope rules design karna |
| Dependency | `Cart and Order` |
| Priority | `P1` |
| Scope | Coupon validation engine design, discount calculation rules, cart preview flow, checkout redemption flow, API/gRPC references, audit/idempotency expectations |
| Not Included | Actual backend code, new migrations, frontend coupon UI, campaign engine, seller analytics APIs, full CMS gRPC implementation |

> **Simple Hinglish goal:** CMS Service me coupon engine ka MVP define karna hai jisse Cart Service coupon preview kar sake aur Order Service successful payment ke baad redemption record kar sake. Engine fixed discount, percentage discount, minimum cart amount, aur category/product/seller scope rules validate karega.

---

## ✅ Final Output Created

```text
TaskImplementation/
└── CMS Service/
    ├── task1.md
    ├── task2.md
    ├── task3.md
    └── task4.md
```

### Why this structure?

- `TaskImplementation/` project me task-wise implementation guides ke liye existing folder hai.
- `CMS Service/` folder already present tha, isliye usko preserve kiya gaya.
- `task4.md` sirf **CMS Service - Task 4** ka coupon engine MVP guide document karta hai.
- Existing backend code, DB migrations, API contracts, ya frontend files modify nahi kiye gaye.

---

## 🧭 Requirement Source Mapping

| Document | Kya use kiya gaya |
|---|---|
| `docs/01-micro-tasks.md` | CMS Service Task 4 ka exact scope: `Coupon engine MVP` |
| `TaskImplementation/CMS Service/task1.md` | Coupon permissions: `cms:coupons:read`, `cms:coupons:create`, `cms:coupons:update`, `cms:coupons:disable` |
| `TaskImplementation/CMS Service/task2.md` | MySQL schema blueprint: `coupons`, `coupon_rules`, `coupon_redemptions` |
| `TaskImplementation/CMS Service/task3.md` | CMS service boundary and audit pattern reference |
| `docs/04-microservice-design.md` | CMS responsibility: coupon validation; validation side-effect free; redemption after payment success |
| `docs/05-database-design.md` | CMS DB indexes and scalability notes for coupon validation caching |
| `docs/09-cms-superadmin.md` | Supported MVP coupon rules and sample validation response |
| `database/draw.sql` | Existing CMS table DDL reference |
| `api/master-api.json` | `ValidateCoupon`, `CreateCoupon`, `UpdateCoupon`, `ListCoupons`, `RecordCouponRedemption` API references |

---

## 🚦 Scope Boundary

### ✅ In Scope

- Fixed amount coupon design.
- Percentage coupon design.
- Minimum cart amount eligibility rule.
- Category scope rule.
- Product scope rule.
- Seller scope rule.
- Active date window and status validation.
- Global usage limit and per-user usage limit validation.
- Cart preview flow with side-effect-free validation.
- Order redemption flow with idempotency.
- MySQL repository query examples.
- Future Go usecase/domain code examples.
- API/gRPC contract reference.
- External tools/libraries explanation.

### 🚫 Out of Scope

- Actual Go files create karna.
- Actual SQL migration files create karna.
- `api/master-api.json` update karna.
- Seller Dashboard coupon UI banana.
- Offer campaign scheduling and budget execution.
- Seller analytics APIs banana.
- Search, recommendation, notification, or payment logic implement karna.

> 🔴 **Reason:** Task 4 ka core goal coupon engine MVP design hai. Real service code, migrations, campaigns, analytics, aur UI later tasks me aayenge.

---

## 🪜 Step-by-Step Implementation

## Step 1: Task requirement identify kiya

`docs/01-micro-tasks.md` me CMS Service ka Task 4 ye hai:

| S.No | Task Name | Detail | Dependency |
|---:|---|---|---|
| 4 | Coupon engine MVP | Fixed, percentage, min cart, category/product/seller scope rules banao | Cart and Order |

**Implementation decision:**

- Coupon ka source of truth **CMS Service** rahega.
- Cart Service coupon preview ke liye CMS Service ko call karega.
- Order Service final successful payment ke baad coupon redemption record karega.
- Coupon validation preview side-effect free hogi, matlab preview ke time usage count increment nahi hoga.

> 🟢 **Why:** Cart me user multiple times coupon try kar sakta hai. Agar preview ke time redemption count badh gaya to coupon incorrectly exhaust ho sakta hai. Final redemption sirf successful order/payment ke baad honi chahiye.

---

## Step 2: Previous CMS tasks ke saath connection samjha

Coupon engine Task 4 previous CMS tasks par build hota hai:

| Previous Task | Coupon engine me use |
|---|---|
| Task 1: Seller permissions | Coupon create/update/disable actions seller role ke basis par control honge |
| Task 2: Choose MySQL | Coupon master, rules, aur redemptions MySQL tables me store honge |
| Task 3: Product moderation flow | Coupon product/category/seller scope Product Service data ke saath align rahega |

### Permission requirement

| Action | Required permission |
|---|---|
| List coupons | `cms:coupons:read` |
| Create coupon | `cms:coupons:create` |
| Update coupon | `cms:coupons:update` |
| Disable coupon | `cms:coupons:disable` |
| Validate coupon internally | Internal service auth |
| Record redemption internally | Internal service auth |

**Explanation:**  
Seller coupon manage kar sakta hai, but buyer/cart direct CMS DB ko touch nahi karega. Cart/Order internal gRPC method use karenge.

---

## Step 3: Coupon engine responsibilities define ki

CMS coupon engine ka MVP teen major workflows handle karega:

| Workflow | Trigger | Output |
|---|---|---|
| Manage coupon | Seller dashboard | Coupon create/update/list/disable |
| Validate coupon | Cart preview / checkout | Valid/invalid + discount amount |
| Record redemption | Order payment success | Coupon usage history stored |

### Responsibility split

| Area | Owner | Reason |
|---|---|---|
| Coupon rules | CMS Service | Business promotion domain CMS ka part hai |
| Cart item snapshots | Cart Service | Cart state Cart Service own karta hai |
| Final order amount | Order Service | Checkout/order source of truth Order Service hai |
| Payment success | Payment Service / Order Service | Payment webhook ke baad final status update hota hai |
| Redemption record | CMS Service | Coupon usage history CMS DB me rahegi |

> 🔵 **Important:** CMS Service Product, Cart, Order, ya Payment DB ko directly query nahi karega. Required cart/order data request payload ke through aayega ya service-to-service contract se milega.

---

## Step 4: Coupon types define kiye

Task 4 ke MVP me 2 discount types support honge:

| Discount Type | Meaning | Example |
|---|---|---|
| `fixed` | Flat amount discount | `SAVE250` gives INR 250 off |
| `percentage` | Cart eligible subtotal ka percent discount | `SAVE10` gives 10% off |

### Fixed discount rule

```text
discount = min(discount_value, eligible_subtotal)
```

Example:

| Eligible subtotal | Fixed discount value | Final discount |
|---:|---:|---:|
| INR 1,000 | INR 250 | INR 250 |
| INR 100 | INR 250 | INR 100 |

**Why:** Discount cart subtotal se zyada nahi hona chahiye, warna total negative ho sakta hai.

### Percentage discount rule

```text
raw_discount = eligible_subtotal * discount_value / 100
discount = min(raw_discount, max_discount_amount if present)
```

Example:

| Eligible subtotal | Coupon | Max cap | Final discount |
|---:|---:|---:|---:|
| INR 2,000 | 10% | No cap | INR 200 |
| INR 10,000 | 10% | INR 500 | INR 500 |

**Why:** Percentage coupon par max cap zaruri hota hai taaki high-value carts par discount uncontrolled na ho.

---

## Step 5: Money representation decide kiya

Money values floating point me store nahi karne.

```text
INR 250.00 => 25000 paise
INR 1,499.99 => 149999 paise
```

### Money model

```json
{
  "amount": 25000,
  "currency": "INR"
}
```

### Why minor units?

| Problem | Minor unit solution |
|---|---|
| Floating point rounding bug | `int64` exact value rakhta hai |
| Currency consistency | Currency explicit field me rahegi |
| DB compatibility | `BIGINT` MySQL me reliable hai |
| Calculation simplicity | Discount formula deterministic hota hai |

> 🟡 **Rule:** `discount_value`, `min_cart_amount`, `max_discount_amount`, `subtotal`, aur `discount_amount` sab minor units me handle honge.

---

## Step 6: Coupon database model use kiya

Task 2 aur `database/draw.sql` ke basis par coupon engine ye tables use karega:

```text
cms_db
├── coupons
├── coupon_rules
└── coupon_redemptions
```

### Coupons table role

`coupons` table master coupon data store karegi:

| Field | Purpose |
|---|---|
| `coupon_id` | Internal stable identifier |
| `seller_id` | Seller-owned coupon scope, nullable for platform-level coupon |
| `code` | Buyer-facing coupon code |
| `discount_type` | `fixed` or `percentage` |
| `discount_value` | Fixed amount in minor units or percentage integer |
| `max_discount_amount` | Percentage coupon cap in minor units |
| `min_cart_amount` | Minimum cart eligibility amount |
| `usage_limit` | Global max redemption count |
| `per_user_limit` | Single user max redemption count |
| `status` | `draft`, `active`, `paused`, `expired` |
| `starts_at`, `ends_at` | Validity window |

### Coupon rules table role

`coupon_rules` flexible scope rules store karegi:

| Rule Type | Example JSON | Meaning |
|---|---|---|
| `category_scope` | `{"category_ids":["cat_1","cat_2"]}` | Coupon sirf selected categories par lagega |
| `product_scope` | `{"product_ids":["prod_1","prod_2"]}` | Coupon sirf selected products par lagega |
| `seller_scope` | `{"seller_ids":["seller_1"]}` | Coupon sirf selected sellers ke items par lagega |

### Coupon redemptions table role

`coupon_redemptions` final usage history store karegi:

| Field | Purpose |
|---|---|
| `coupon_id` | Redeemed coupon |
| `order_id` | Final order reference |
| `user_id` | Buyer who used coupon |
| `discount_amount` | Applied discount |
| `currency` | Discount currency |
| `redeemed_at` | Final redemption time |

### Idempotency key

```sql
UNIQUE KEY uk_coupon_redemptions_coupon_order (coupon_id, order_id)
```

**Explanation:**  
Same order ke liye same coupon do baar record nahi hona chahiye. Payment webhook retry ho ya Order Service retry kare, duplicate redemption block ho jayega.

---

## Step 7: Coupon states define kiye

Coupon engine MVP me coupon status simple rakha gaya:

| Status | Meaning | Validatable? |
|---|---|---:|
| `draft` | Seller ne coupon banaya but live nahi kiya | ❌ |
| `active` | Coupon live hai | ✅ |
| `paused` | Seller ne temporarily stop kiya | ❌ |
| `expired` | Time window complete ho chuki hai | ❌ |

### State flow

```mermaid
stateDiagram-v2
    [*] --> draft: create coupon
    draft --> active: activate
    active --> paused: pause
    paused --> active: resume
    active --> expired: end time passed
    paused --> expired: end time passed
```

**Explanation:**  
MVP me delete operation avoid karna better hai. Coupon disable/pause karke audit history preserve hoti hai.

---

## Step 8: Coupon create/update rules design kiye

Coupon create/update seller dashboard se hoga.

### Create coupon validation

| Field | Rule |
|---|---|
| `code` | Required, uppercase normalized, unique |
| `discount_type` | `fixed` ya `percentage` |
| `discount_value` | Positive |
| `currency` | Valid ISO style currency, MVP default `INR` |
| `min_cart_amount` | Optional, zero ya positive |
| `max_discount_amount` | Percentage coupon ke liye optional cap |
| `starts_at` | Optional, `ends_at` se pehle |
| `ends_at` | Optional, `starts_at` ke baad |
| `usage_limit` | Optional, positive |
| `per_user_limit` | Optional, positive |
| `rules` | Supported rule types only |

### Normalization

```text
Input code: " save250 "
Stored code: "SAVE250"
```

**Why:** User lowercase/space ke saath coupon enter kare to bhi same coupon match ho.

### Coupon input example

```json
{
  "code": "SAVE250",
  "discount_type": "fixed",
  "discount_value": 25000,
  "min_cart_amount": {
    "amount": 99900,
    "currency": "INR"
  },
  "starts_at": "2026-06-01T00:00:00Z",
  "ends_at": "2026-06-30T23:59:59Z",
  "usage_limit": 10000
}
```

---

## Step 9: Validation request contract define kiya

`api/master-api.json` me existing reference:

```json
{
  "name": "ValidateCoupon",
  "input_schema": "CouponPreviewRequest",
  "output_schema": "CouponPreviewResponse",
  "auth": "internal"
}
```

### Current simple request

```json
{
  "coupon_code": "SAVE250",
  "cart_id": "cart_123",
  "order_id": "order_456"
}
```

### Recommended internal validation payload

MVP implementation ke time CMS ko cart/order details chahiye honge. Cart/Order Service request me ye normalized snapshot bhej sakte hain:

```json
{
  "coupon_code": "SAVE250",
  "user_id": "user_123",
  "cart_id": "cart_456",
  "currency": "INR",
  "subtotal_amount": 149900,
  "items": [
    {
      "product_id": "prod_1",
      "variant_id": "var_1",
      "seller_id": "seller_10",
      "category_ids": ["cat_shoes", "cat_sports"],
      "quantity": 1,
      "unit_amount": 149900,
      "line_subtotal_amount": 149900
    }
  ]
}
```

### Response

```json
{
  "valid": true,
  "coupon_id": "coupon_123",
  "discount": {
    "amount": 25000,
    "currency": "INR"
  },
  "reason": null
}
```

**Explanation:**  
Validation response simple rahega: coupon valid hai ya nahi, discount kitna hai, aur invalid hone par reason kya hai.

---

## Step 10: Validation pipeline banaya

Coupon validation deterministic sequence follow karegi:

```mermaid
flowchart TD
    A[ValidateCoupon request] --> B[Normalize coupon code]
    B --> C[Load active coupon by code]
    C --> D{Coupon found?}
    D -- no --> X[Invalid: coupon_not_found]
    D -- yes --> E[Check status and time window]
    E --> F{Currency match?}
    F -- no --> Y[Invalid: currency_mismatch]
    F -- yes --> G[Check minimum cart amount]
    G --> H[Load coupon rules]
    H --> I[Evaluate product/category/seller scope]
    I --> J{Eligible subtotal > 0?}
    J -- no --> Z[Invalid: scope_not_matched]
    J -- yes --> K[Check global and per-user usage]
    K --> L[Calculate discount]
    L --> M[Return valid response]
```

### Pipeline order kyun important hai?

| Order | Reason |
|---:|---|
| 1. Code normalize | DB lookup predictable hota hai |
| 2. Master coupon fetch | Missing coupon fast fail |
| 3. Status/time check | Expired/paused coupon par extra work avoid |
| 4. Currency/min cart | Basic eligibility clear hoti hai |
| 5. Scope rules | Eligible items calculate hote hain |
| 6. Usage limits | Abuse and exhausted coupons block hote hain |
| 7. Discount calc | Final amount only valid coupon par calculate hota hai |

---

## Step 11: Master coupon validation rules implement kiye

Master validation coupon row ke core fields par based hogi:

| Check | Failure reason |
|---|---|
| Coupon code exists | `coupon_not_found` |
| Status is `active` | `coupon_not_active` |
| Current time >= `starts_at` | `coupon_not_started` |
| Current time <= `ends_at` | `coupon_expired` |
| Currency matches cart currency | `currency_mismatch` |
| Cart subtotal >= `min_cart_amount` | `min_cart_not_met` |

### SQL lookup example

```sql
SELECT coupon_id,
       seller_id,
       code,
       discount_type,
       discount_value,
       max_discount_amount,
       min_cart_amount,
       currency,
       usage_limit,
       per_user_limit,
       status,
       starts_at,
       ends_at
FROM coupons
WHERE code = ?
LIMIT 1;
```

**Explanation:**  
Time window ko SQL me filter kar sakte hain, but service-level checks readable error reasons dene ke liye better hain. Agar SQL directly only active coupon fetch karega, then expired vs paused ka exact reason lose ho sakta hai.

---

## Step 12: Minimum cart rule define kiya

Minimum cart amount coupon master field me store hoga:

```text
min_cart_amount = 99900 means INR 999.00
```

### Rule

```text
if cart.subtotal_amount < coupon.min_cart_amount:
    invalid = min_cart_not_met
```

### Example

| Cart subtotal | Min cart | Result |
|---:|---:|---|
| INR 800 | INR 999 | Invalid |
| INR 999 | INR 999 | Valid |
| INR 1,499 | INR 999 | Valid |

**Explanation:**  
Minimum cart amount total cart subtotal par check hota hai. Scope rules ke baad discount sirf eligible items par calculate hoga, but basic eligibility cart subtotal level par rakhi gayi hai for MVP simplicity.

---

## Step 13: Product scope rule design kiya

Product scope coupon sirf selected products par apply hoga.

### Rule value

```json
{
  "product_ids": ["prod_101", "prod_202"]
}
```

### Evaluation

```text
eligible_items = cart.items where item.product_id in product_ids
eligible_subtotal = sum(eligible_items.line_subtotal_amount)
```

### Example

| Item | Product ID | Line subtotal | Eligible? |
|---|---|---:|---:|
| Shoe | `prod_101` | INR 1,000 | ✅ |
| Bag | `prod_999` | INR 500 | ❌ |

Discount calculation base:

```text
eligible_subtotal = INR 1,000
```

**Explanation:**  
Coupon cart total par visible ho sakta hai, but discount only matching products ke subtotal par calculate hoga.

---

## Step 14: Category scope rule design kiya

Category scope selected categories ke products par coupon allow karega.

### Rule value

```json
{
  "category_ids": ["cat_shoes", "cat_sports"]
}
```

### Evaluation

```text
eligible_items = cart.items where any item.category_ids exists in allowed category_ids
```

### Example

| Item | Categories | Line subtotal | Eligible? |
|---|---|---:|---:|
| Running shoe | `cat_shoes`, `cat_sports` | INR 2,000 | ✅ |
| Phone case | `cat_mobile_accessories` | INR 300 | ❌ |

**Explanation:**  
Cart Service request me item category ids snapshot bhejega. CMS Service Product DB direct query nahi karega.

---

## Step 15: Seller scope rule design kiya

Seller scope marketplace coupon ke liye useful hai jahan coupon selected seller stores par apply hota hai.

### Rule value

```json
{
  "seller_ids": ["seller_10", "seller_20"]
}
```

### Evaluation

```text
eligible_items = cart.items where item.seller_id in seller_ids
```

### Seller-owned coupon shortcut

If `coupons.seller_id` set hai, then coupon by default us seller ke items par scoped hoga.

```text
coupon.seller_id = seller_10
eligible_items = cart.items where item.seller_id == seller_10
```

**Explanation:**  
Seller-created coupon automatically own seller catalog par apply hona chahiye. Platform-level coupon ke liye `seller_id` null ho sakta hai and explicit seller/category/product rules apply ho sakte hain.

---

## Step 16: Multiple scope rules combine karne ka rule

MVP me multiple scope rules ko **AND between rule groups** aur **OR inside same group** ke tarah treat karna clean rahega.

### Example

Coupon rules:

```json
[
  {"rule_type": "seller_scope", "rule_value": {"seller_ids": ["seller_10"]}},
  {"rule_type": "category_scope", "rule_value": {"category_ids": ["cat_shoes", "cat_sports"]}}
]
```

Meaning:

```text
Item seller_10 ka hona chahiye
AND item category cat_shoes OR cat_sports me honi chahiye
```

### Rule combination table

| Rule group | Inside group | Across groups |
|---|---|---|
| Product IDs | OR | AND |
| Category IDs | OR | AND |
| Seller IDs | OR | AND |

**Why:**  
Is approach se seller + category jaise targeted coupon predictable bante hain. Agar future me complex boolean rules chahiye, rule engine later expand ho sakta hai.

---

## Step 17: Eligible subtotal calculation define kiya

Coupon discount hamesha eligible items ke subtotal par calculate hoga.

### Flow

```mermaid
flowchart LR
    Cart[Cart subtotal] --> Rules[Scope rules]
    Rules --> Eligible[Eligible items]
    Eligible --> Base[Eligible subtotal]
    Base --> Discount[Discount calculation]
```

### No scope rule case

If product/category/seller scope rules nahi hain:

```text
eligible_subtotal = cart.subtotal_amount
```

### With scope rules

If scope rules present hain:

```text
eligible_subtotal = sum(matching item subtotals)
```

### Invalid condition

```text
if eligible_subtotal <= 0:
    invalid = scope_not_matched
```

**Explanation:**  
Buyer coupon apply kare but cart me matching item hi nahi hai, to discount zero return karne ke bajay clear invalid reason dena better UX hai.

---

## Step 18: Usage limit validation design kiya

Coupon usage limits redemption table se check honge.

### Global usage limit

```sql
SELECT COUNT(*) AS used_count
FROM coupon_redemptions
WHERE coupon_id = ?;
```

Validation:

```text
if usage_limit is not null and used_count >= usage_limit:
    invalid = usage_limit_reached
```

### Per-user usage limit

```sql
SELECT COUNT(*) AS user_used_count
FROM coupon_redemptions
WHERE coupon_id = ?
  AND user_id = ?;
```

Validation:

```text
if per_user_limit is not null and user_used_count >= per_user_limit:
    invalid = per_user_limit_reached
```

### Race condition note

Preview ke time count check helpful hai, but final guarantee redemption transaction me hogi. High traffic coupons ke liye future me counter table or Redis atomic counters add kiye ja sakte hain.

> 🟡 **MVP decision:** Preview side-effect free rahega. Final redemption insert idempotent hoga.

---

## Step 19: Discount calculation rules implement kiye

### Fixed discount calculation

```go
func calculateFixedDiscount(eligibleSubtotal int64, discountValue int64) int64 {
	if discountValue <= 0 || eligibleSubtotal <= 0 {
		return 0
	}
	if discountValue > eligibleSubtotal {
		return eligibleSubtotal
	}
	return discountValue
}
```

### Percentage discount calculation

```go
func calculatePercentageDiscount(eligibleSubtotal int64, percent int64, maxCap *int64) int64 {
	if percent <= 0 || eligibleSubtotal <= 0 {
		return 0
	}

	discount := eligibleSubtotal * percent / 100
	if maxCap != nil && discount > *maxCap {
		return *maxCap
	}
	if discount > eligibleSubtotal {
		return eligibleSubtotal
	}
	return discount
}
```

### Why integer division acceptable?

Money minor units me hai. Percentage calculation floor rounding use karegi. Example:

```text
999 paise * 10 / 100 = 99 paise
```

MVP me floor rounding deterministic hai. Future me financial rounding policy central shared package me define ki ja sakti hai.

---

## Step 20: Validation result reasons standardize kiye

Frontend and Cart/Order services ko consistent reason code chahiye.

| Reason Code | Meaning |
|---|---|
| `coupon_not_found` | Code DB me nahi mila |
| `coupon_not_active` | Coupon draft/paused/expired status me hai |
| `coupon_not_started` | Start time future me hai |
| `coupon_expired` | End time pass ho chuki hai |
| `currency_mismatch` | Cart currency coupon currency se match nahi karti |
| `min_cart_not_met` | Cart subtotal minimum amount se kam hai |
| `scope_not_matched` | Cart me eligible product/category/seller item nahi |
| `usage_limit_reached` | Global usage limit complete ho chuki |
| `per_user_limit_reached` | User ne allowed limit use kar li |
| `invalid_discount_config` | Coupon values invalid/corrupt hain |

### Invalid response example

```json
{
  "valid": false,
  "coupon_id": "coupon_123",
  "discount": {
    "amount": 0,
    "currency": "INR"
  },
  "reason": "min_cart_not_met"
}
```

**Explanation:**  
Human-readable message frontend translate karega. CMS Service stable machine reason codes return karega.

---

## Step 21: Cart preview flow design kiya

Cart Service coupon preview request receive karega and CMS Service ko validate call karega.

```mermaid
sequenceDiagram
    participant Buyer as Buyer App
    participant GW as API Gateway
    participant Cart as Cart Service
    participant CMS as CMS Service
    participant DB as MySQL cms_db

    Buyer->>GW: POST /api/v1/cart/coupons/preview
    GW->>Cart: ApplyCouponPreview(coupon_code)
    Cart->>Cart: Load cart snapshot
    Cart->>CMS: ValidateCoupon(cart snapshot + coupon_code)
    CMS->>DB: Load coupon, rules, usage counts
    DB-->>CMS: Coupon data
    CMS-->>Cart: valid + discount + reason
    Cart-->>GW: Cart totals preview
    GW-->>Buyer: Updated cart preview
```

### Important rule

```text
Preview does not insert into coupon_redemptions.
```

**Explanation:**  
Preview bas discount estimate karega. Buyer checkout cancel kare to coupon usage count affect nahi hoga.

---

## Step 22: Checkout redemption flow design kiya

Final redemption Order Service payment success ke baad record karega.

```mermaid
sequenceDiagram
    participant Buyer as Buyer App
    participant GW as API Gateway
    participant Order as Order Service
    participant CMS as CMS Service
    participant Payment as Payment Service
    participant DB as MySQL cms_db

    Buyer->>GW: POST /api/v1/orders/checkout with coupon_code
    GW->>Order: CreateOrderFromCart
    Order->>CMS: ValidateCoupon(final order/cart snapshot)
    CMS-->>Order: valid + discount
    Order->>Order: Create pending order with discount snapshot
    Order->>Payment: CreatePaymentIntent
    Payment-->>Order: payment intent
    Payment->>Order: Webhook payment success
    Order->>CMS: RecordCouponRedemption(order_id, coupon_id, user_id, discount)
    CMS->>DB: INSERT coupon_redemptions
    CMS-->>Order: success/idempotent success
```

### Why redemption after payment success?

| Option | Problem |
|---|---|
| Redemption at preview | User may not checkout |
| Redemption at order created | Payment may fail |
| Redemption after payment success | ✅ Coupon usage matches real paid order |

---

## Step 23: Redemption transaction design kiya

Redemption write idempotent hona chahiye.

### SQL insert example

```sql
INSERT INTO coupon_redemptions (
  redemption_id,
  coupon_id,
  order_id,
  user_id,
  discount_amount,
  currency
) VALUES (?, ?, ?, ?, ?, ?);
```

### Duplicate handling

If duplicate key error on `(coupon_id, order_id)`:

```text
Treat as idempotent success if existing row matches same user and discount.
Return conflict if data mismatch.
```

### Pseudocode

```go
func (u *CouponUsecase) RecordRedemption(ctx context.Context, input RedemptionInput) error {
	err := u.repo.InsertRedemption(ctx, input)
	if err == nil {
		return nil
	}
	if IsDuplicateCouponOrder(err) {
		existing, findErr := u.repo.FindRedemption(ctx, input.CouponID, input.OrderID)
		if findErr != nil {
			return findErr
		}
		if existing.UserID == input.UserID && existing.DiscountAmount == input.DiscountAmount {
			return nil
		}
		return ErrRedemptionConflict
	}
	return err
}
```

**Explanation:**  
Retry safe hai. Same request repeat ho to success, but different amount/user ke saath same order redemption aaye to conflict.

---

## Step 24: Architecture diagram

```mermaid
flowchart TB
    Seller[Seller Dashboard] --> GW[API Gateway]
    Buyer[Buyer App] --> GW

    GW --> Cart[Cart Service]
    GW --> Order[Order Service]
    GW --> CMS[CMS Service]

    Cart -->|ValidateCoupon preview| CMS
    Order -->|ValidateCoupon checkout| CMS
    Order -->|RecordCouponRedemption| CMS

    CMS --> DB[(MySQL cms_db)]

    DB --> Coupons[(coupons)]
    DB --> Rules[(coupon_rules)]
    DB --> Redemptions[(coupon_redemptions)]
```

### Explanation

- Seller dashboard coupon create/update/list ke liye CMS Service use karega.
- Buyer app cart preview ke liye indirectly Cart Service -> CMS Service flow use karega.
- Order Service checkout ke time coupon discount snapshot store karega.
- Payment success ke baad Order Service CMS Service me final redemption record karega.

---

## Step 25: Future folder structure define kiya

Task 4 me actual backend files create nahi kiye gaye, but real implementation is structure me fit hogi:

```text
backend/
└── services/
    └── cms-service/
        ├── cmd/
        │   └── server/
        │       └── main.go
        ├── internal/
        │   ├── domain/
        │   │   ├── coupon.go
        │   │   └── coupon_rule.go
        │   ├── usecase/
        │   │   ├── validate_coupon.go
        │   │   ├── manage_coupon.go
        │   │   └── record_redemption.go
        │   ├── repository/
        │   │   └── mysql_coupon_repository.go
        │   └── transport/
        │       └── grpc/
        │           └── coupon_handler.go
        ├── migrations/
        │   ├── 001_create_coupon_tables.up.sql
        │   └── 001_create_coupon_tables.down.sql
        └── deploy/
            └── Dockerfile
```

### File responsibility

| File | Responsibility |
|---|---|
| `domain/coupon.go` | Coupon entity, status, discount type, validation reason constants |
| `domain/coupon_rule.go` | Rule payload structs and evaluator helpers |
| `usecase/validate_coupon.go` | Main validation pipeline |
| `usecase/manage_coupon.go` | Seller coupon create/update/list/disable workflow |
| `usecase/record_redemption.go` | Final redemption idempotency flow |
| `repository/mysql_coupon_repository.go` | SQL queries for coupons, rules, usage counts, redemptions |
| `transport/grpc/coupon_handler.go` | gRPC request/response mapping |
| `migrations/*.sql` | CMS coupon table creation |

> 🔵 **Note:** Ye folder structure reference hai. Is task me only documentation file create ki gayi hai.

---

## Step 26: Domain model code example

Future Go implementation ke liye sample domain model:

```go
package domain

import "time"

type DiscountType string

const (
	DiscountTypeFixed      DiscountType = "fixed"
	DiscountTypePercentage DiscountType = "percentage"
)

type CouponStatus string

const (
	CouponStatusDraft   CouponStatus = "draft"
	CouponStatusActive  CouponStatus = "active"
	CouponStatusPaused  CouponStatus = "paused"
	CouponStatusExpired CouponStatus = "expired"
)

type Coupon struct {
	CouponID          string
	SellerID          *string
	Code              string
	DiscountType      DiscountType
	DiscountValue     int64
	MaxDiscountAmount *int64
	MinCartAmount     *int64
	Currency          string
	UsageLimit        *int
	PerUserLimit      *int
	Status            CouponStatus
	StartsAt          *time.Time
	EndsAt            *time.Time
}
```

**Explanation:**  
Optional fields pointer rakhe gaye hain, taaki `NULL` and zero value ka difference clear rahe. Example: `MaxDiscountAmount == nil` means cap configured nahi hai.

---

## Step 27: Validation usecase code example

```go
func (u *CouponUsecase) ValidateCoupon(ctx context.Context, req ValidateCouponRequest) (ValidateCouponResult, error) {
	code := NormalizeCouponCode(req.CouponCode)

	coupon, err := u.repo.FindCouponByCode(ctx, code)
	if err == ErrNotFound {
		return Invalid(req.Currency, "coupon_not_found"), nil
	}
	if err != nil {
		return ValidateCouponResult{}, err
	}

	if reason := validateMasterCoupon(coupon, req); reason != "" {
		return Invalid(req.Currency, reason), nil
	}

	rules, err := u.repo.ListCouponRules(ctx, coupon.CouponID)
	if err != nil {
		return ValidateCouponResult{}, err
	}

	eligibleSubtotal := CalculateEligibleSubtotal(req.Items, coupon, rules)
	if eligibleSubtotal <= 0 {
		return Invalid(req.Currency, "scope_not_matched"), nil
	}

	if reason, err := u.checkUsageLimits(ctx, coupon, req.UserID); reason != "" || err != nil {
		if err != nil {
			return ValidateCouponResult{}, err
		}
		return Invalid(req.Currency, reason), nil
	}

	discount := CalculateDiscount(coupon, eligibleSubtotal)
	if discount <= 0 {
		return Invalid(req.Currency, "invalid_discount_config"), nil
	}

	return ValidateCouponResult{
		Valid:    true,
		CouponID: coupon.CouponID,
		Discount: Money{
			Amount:   discount,
			Currency: coupon.Currency,
		},
	}, nil
}
```

**Explanation:**  
Usecase clear stages me split hai: fetch, master validate, rules evaluate, usage check, discount calculate. Isse tests simple ho jate hain.

---

## Step 28: Scope evaluator code example

```go
func CalculateEligibleSubtotal(items []CartItem, coupon Coupon, rules []CouponRule) int64 {
	var total int64

	for _, item := range items {
		if !matchesSellerScope(item, coupon, rules) {
			continue
		}
		if !matchesProductScope(item, rules) {
			continue
		}
		if !matchesCategoryScope(item, rules) {
			continue
		}
		total += item.LineSubtotalAmount
	}

	return total
}
```

### Scope helper idea

```go
func matchesProductScope(item CartItem, rules []CouponRule) bool {
	allowedIDs := collectProductIDs(rules)
	if len(allowedIDs) == 0 {
		return true
	}
	return allowedIDs[item.ProductID]
}
```

**Explanation:**  
If product scope configured nahi hai, product check pass ho jayega. Agar configured hai, item ka product allowed list me hona chahiye.

---

## Step 29: Repository code examples

### Find coupon by code

```go
const findCouponByCodeQuery = `
SELECT coupon_id, seller_id, code, discount_type, discount_value,
       max_discount_amount, min_cart_amount, currency,
       usage_limit, per_user_limit, status, starts_at, ends_at
FROM coupons
WHERE code = ?
LIMIT 1`
```

### List coupon rules

```go
const listCouponRulesQuery = `
SELECT rule_id, coupon_id, rule_type, rule_value
FROM coupon_rules
WHERE coupon_id = ?
ORDER BY id ASC`
```

### Count usage

```go
const countCouponUsageQuery = `
SELECT COUNT(*)
FROM coupon_redemptions
WHERE coupon_id = ?`
```

### Count user usage

```go
const countUserCouponUsageQuery = `
SELECT COUNT(*)
FROM coupon_redemptions
WHERE coupon_id = ?
  AND user_id = ?`
```

**Explanation:**  
Repository layer SQL ko isolate karega. Usecase ko DB query details nahi pata honi chahiye.

---

## Step 30: API mapping define kiya

### Seller coupon APIs

| REST API | gRPC | Auth | Purpose |
|---|---|---|---|
| `GET /api/v1/seller/coupons` | `CMSService.ListCoupons` | Seller | Seller coupons list |
| `POST /api/v1/seller/coupons` | `CMSService.CreateCoupon` | Seller | Coupon create |
| `PATCH /api/v1/seller/coupons/{coupon_id}` | `CMSService.UpdateCoupon` | Seller | Coupon update |

### Internal coupon APIs

| Caller | gRPC | Purpose |
|---|---|---|
| Cart Service | `CMSService.ValidateCoupon` | Coupon preview |
| Order Service | `CMSService.ValidateCoupon` | Checkout final validation |
| Order Service | `CMSService.RecordCouponRedemption` | Payment success ke baad usage record |

### Cart coupon preview API

`api/master-api.json` me Cart route:

```text
POST /api/v1/cart/coupons/preview
```

Flow:

```text
Frontend -> API Gateway -> Cart Service -> CMS Service
```

**Explanation:**  
Buyer-facing route Cart Service ke through rahega. CMS Service buyer app ke liye direct public coupon validation endpoint expose nahi karega.

---

## Step 31: External libraries/tools

Is documentation-only task me koi package install nahi kiya gaya. Real CMS coupon engine implement karte time ye tools/libraries useful honge:

| Tool/Library | What it is | Why used | Install |
|---|---|---|---|
| MySQL 8+ | Relational database | Coupons, rules, redemptions, audit data structured and transactional hain | Docker ya local package manager |
| `github.com/go-sql-driver/mysql` | Go MySQL driver | Go service ko MySQL se connect karne ke liye | `go get github.com/go-sql-driver/mysql` |
| `golang-migrate/migrate` | DB migration tool | Versioned SQL migrations run karne ke liye | `go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest` |
| `google.golang.org/grpc` | gRPC framework | Cart/Order -> CMS internal calls typed banane ke liye | `go get google.golang.org/grpc` |
| `google.golang.org/protobuf` | Protobuf runtime | Generated proto messages use karne ke liye | `go get google.golang.org/protobuf` |
| `github.com/redis/go-redis/v9` | Redis client | Future short TTL coupon validation cache ke liye optional | `go get github.com/redis/go-redis/v9` |
| Mermaid | Markdown diagram syntax | Architecture and flow diagrams docs me render karne ke liye | GitHub/GitLab markdown me built-in support often available |

### MySQL quick start

```bash
docker run --name ecommerce-cms-mysql \
  -e MYSQL_ROOT_PASSWORD=root \
  -e MYSQL_DATABASE=cms_db \
  -p 3306:3306 \
  -d mysql:8
```

### MySQL connection string example

```text
cms_user:cms_password@tcp(localhost:3306)/cms_db?parseTime=true&charset=utf8mb4&loc=UTC
```

### Migration command example

```bash
migrate -path backend/services/cms-service/migrations \
  -database "mysql://cms_user:cms_password@tcp(localhost:3306)/cms_db" \
  up
```

### Redis optional cache example

```text
Key: coupon_validation:SAVE250:cart_hash_abc
TTL: 30 seconds
Value: validation response JSON
```

> 🟡 **Cache caution:** Redemption/usage final state MySQL se authoritative rahega. Cache sirf short-lived preview optimization ke liye use karna.

---

## Step 32: Observability and audit design

Coupon engine me business-impacting actions traceable hone chahiye.

### Audit actions

| Action | Resource type | Resource id |
|---|---|---|
| Coupon created | `coupon` | `coupon_id` |
| Coupon updated | `coupon` | `coupon_id` |
| Coupon paused | `coupon` | `coupon_id` |
| Coupon activated | `coupon` | `coupon_id` |
| Coupon redemption recorded | `coupon_redemption` | `redemption_id` |

### Structured log example

```json
{
  "level": "info",
  "event": "coupon_validated",
  "request_id": "req_123",
  "coupon_code": "SAVE250",
  "coupon_id": "coupon_123",
  "valid": true,
  "discount_amount": 25000,
  "currency": "INR"
}
```

### Metrics

| Metric | Type | Purpose |
|---|---|---|
| `cms_coupon_validate_total` | Counter | Validation volume |
| `cms_coupon_validate_invalid_total` | Counter | Invalid reason trend |
| `cms_coupon_validate_duration_ms` | Histogram | Latency monitoring |
| `cms_coupon_redemption_total` | Counter | Successful redemptions |
| `cms_coupon_redemption_duplicate_total` | Counter | Retry/idempotency visibility |

---

## Step 33: Security and abuse controls

Coupon endpoint abuse common hota hai, isliye basic controls important hain.

| Risk | Control |
|---|---|
| Coupon brute force | API Gateway rate limiting on preview route |
| Seller cross-scope update | Seller `seller_id` scope check |
| Coupon code enumeration | Generic frontend message, internal reason code |
| Duplicate redemption | Unique key `(coupon_id, order_id)` |
| Price tampering | Order Service final snapshot and validation use kare |
| Old cart replay | Checkout ke time final re-validation required |

### Seller scope rule

```text
Seller can create/update only coupons where coupon.seller_id == auth.seller_id.
```

### Internal auth rule

```text
ValidateCoupon and RecordCouponRedemption public browser ke liye direct expose nahi honge.
Only Cart and Order services internal auth ke saath call karenge.
```

---

## Step 34: Test plan define kiya

### Unit tests

| Test Case | Expected |
|---|---|
| Fixed coupon on higher subtotal | Full fixed discount |
| Fixed coupon greater than subtotal | Discount capped at subtotal |
| Percentage coupon without cap | Percent discount |
| Percentage coupon with cap | Discount capped |
| Min cart not met | Invalid `min_cart_not_met` |
| Product scope matched | Valid on eligible item subtotal |
| Product scope not matched | Invalid `scope_not_matched` |
| Category scope matched | Valid |
| Seller scope matched | Valid |
| Expired coupon | Invalid `coupon_expired` |
| Paused coupon | Invalid `coupon_not_active` |
| Per-user limit reached | Invalid `per_user_limit_reached` |

### Repository tests

| Test Case | Expected |
|---|---|
| Find coupon by normalized code | Coupon returned |
| Duplicate coupon code create | Unique constraint error |
| List rules by coupon id | Rules ordered by id |
| Count global usage | Correct count |
| Count per-user usage | Correct count |
| Duplicate redemption insert | Idempotent duplicate detected |

### Integration tests

| Flow | Expected |
|---|---|
| Cart preview valid coupon | Cart discount preview returned |
| Checkout re-validates coupon | Order discount snapshot matches CMS result |
| Payment success records redemption | `coupon_redemptions` row created |
| Payment webhook retry | No duplicate redemption |

---

## Step 35: Performance strategy

Coupon validation hot path ho sakta hai, especially checkout and cart preview.

### MVP indexes used

| Index | Purpose |
|---|---|
| `coupons(code)` | Fast coupon lookup |
| `coupons(seller_id, status)` | Seller coupon list |
| `coupon_rules(coupon_id)` | Fast rule load |
| `coupon_redemptions(user_id, coupon_id)` | Per-user usage check |
| `coupon_redemptions(coupon_id, order_id)` | Idempotent redemption |

### Optional cache

Short TTL cache only for validation response:

```text
TTL: 15-30 seconds
Use for: repeated cart preview
Do not use for: final redemption write
```

### Scaling note

High traffic platform coupons ke liye future optimization:

- `coupon_usage_counters` table.
- Redis atomic counters with reconciliation.
- Async analytics projection.
- Read replica for coupon preview reads.

---

## Step 36: Beginner-friendly end-to-end example

### Coupon setup

Seller creates coupon:

```json
{
  "code": "SHOE10",
  "discount_type": "percentage",
  "discount_value": 10,
  "max_discount_amount": 50000,
  "min_cart_amount": 100000,
  "rules": [
    {
      "rule_type": "category_scope",
      "rule_value": {
        "category_ids": ["cat_shoes"]
      }
    }
  ]
}
```

### Buyer cart

```json
{
  "subtotal_amount": 200000,
  "currency": "INR",
  "items": [
    {
      "product_id": "prod_shoe_1",
      "category_ids": ["cat_shoes"],
      "seller_id": "seller_10",
      "line_subtotal_amount": 150000
    },
    {
      "product_id": "prod_bag_1",
      "category_ids": ["cat_bags"],
      "seller_id": "seller_10",
      "line_subtotal_amount": 50000
    }
  ]
}
```

### Calculation

```text
Cart subtotal = INR 2,000
Min cart = INR 1,000, so valid
Eligible category = cat_shoes
Eligible subtotal = INR 1,500
Discount = 10% of INR 1,500 = INR 150
Max cap = INR 500, so final discount = INR 150
```

### Final response

```json
{
  "valid": true,
  "coupon_id": "coupon_shoe10",
  "discount": {
    "amount": 15000,
    "currency": "INR"
  },
  "reason": null
}
```

---

## ✅ Implementation Checklist

| Item | Status |
|---|---:|
| `TaskImplementation/` folder present | ✅ |
| `TaskImplementation/CMS Service/` folder preserved | ✅ |
| `task4.md` created | ✅ |
| Task source mapped to docs | ✅ |
| Scope limited to CMS Service Task 4 | ✅ |
| Fixed discount rule documented | ✅ |
| Percentage discount rule documented | ✅ |
| Minimum cart rule documented | ✅ |
| Product/category/seller scope rules documented | ✅ |
| Cart preview flow documented | ✅ |
| Order redemption flow documented | ✅ |
| MySQL table usage documented | ✅ |
| External tools/libraries explained | ✅ |
| Mermaid diagrams added | ✅ |
| Code examples added | ✅ |

---

## 🏁 Final Summary

CMS Service Task 4 ke liye Coupon Engine MVP ka final design:

| Area | Decision |
|---|---|
| Owner service | CMS Service |
| Source of truth | MySQL `cms_db` |
| Main tables | `coupons`, `coupon_rules`, `coupon_redemptions` |
| Discount types | `fixed`, `percentage` |
| Eligibility | Status, time window, currency, min cart, usage limits |
| Scope rules | Product, category, seller |
| Preview behavior | Side-effect free |
| Final redemption | Order Service payment success ke baad |
| Idempotency | Unique key on `(coupon_id, order_id)` |
| Public buyer flow | Buyer -> Gateway -> Cart -> CMS |
| Internal final flow | Order -> CMS `RecordCouponRedemption` |

> 🟢 **Task 4 complete:** Coupon engine MVP ka step-by-step design ready hai. Fixed/percentage discount, min cart, product/category/seller scope, validation pipeline, cart preview, checkout redemption, idempotency, tools, folder structure, diagrams, and code examples documented hain. No extra implementation beyond CMS Service Task 4 kiya gaya.
