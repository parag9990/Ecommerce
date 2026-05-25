# 🛒 Cart Service - Task 4: Add Item Flow

![Task](https://img.shields.io/badge/Task-Cart%20Service%20%234-blue)
![Priority](https://img.shields.io/badge/Priority-P0-red)
![Dependency](https://img.shields.io/badge/Dependency-Product%20Service-yellow)
![Status](https://img.shields.io/badge/Status-Documented-success)
![Guide](https://img.shields.io/badge/Guide-Hinglish-orange)
![Scope](https://img.shields.io/badge/Scope-Add%20Item%20Flow-purple)

## 📌 Task Summary

| Field | Detail |
|---|---|
| Task Name | Add item flow |
| Source | `docs/01-micro-tasks.md` → `Cart Service` → Task 4 |
| Priority | `P0` foundation/blocker |
| Dependency | Product Service |
| Main Goal | Product validate karna, stock check karna, aur cart me item add/update karna |
| Output Type | Documentation-only implementation guide |
| Not Included | Remove item flow, coupon preview, guest cart merge, checkout inventory reservation, actual backend file creation |

> **Simple Hinglish goal:** Is task ka purpose Cart Service ke `AddItem` flow ko properly design karna hai. Buyer jab product add karega, Cart Service Product Service se product/variant validate karega, stock check karega, cart me item add ya quantity update karega, totals recalculate karega, MongoDB me save karega, Redis cache refresh karega, aur updated cart return karega.

---

## ✅ Final Output Created

```text
TaskImplementation/
└── Cart Service/
    ├── task1.md
    ├── task2.md
    ├── task3.md
    └── task4.md
```

### Why this structure?

- `TaskImplementation/` project ke task-wise implementation guides ka central folder hai.
- `Cart Service/` Cart Service ke saare task guides ko group karta hai.
- `task4.md` sirf **Cart Service - Task 4** ka Add Item flow guide hai.
- Is task me sirf documentation artifact create kiya gaya hai; backend source code create nahi kiya gaya.

---

## 🧭 Implementation Approach

Is guide ko banane se pehle existing project documentation study ki gayi:

| Document | Kya use kiya gaya |
|---|---|
| `docs/01-micro-tasks.md` | Task 4 ka exact scope: product validate, stock check, item add/update |
| `TaskImplementation/Cart Service/task1.md` | Cart rules: guest/user cart, quantity limit, price snapshot, one active cart |
| `TaskImplementation/Cart Service/task2.md` | MongoDB source of truth and Redis hot cache strategy |
| `TaskImplementation/Cart Service/task3.md` | `carts` collection, embedded `items[]`, totals, indexes, version field |
| `docs/04-microservice-design.md` | Cart Service responsibilities, gRPC methods, REST routes |
| `docs/03-folder-structure.md` | Future `backend/services/cart-service` code structure |
| `api/master-api.json` | `POST /api/v1/cart/items`, `CartService.AddItem`, `CartItemInput`, `Cart` response |
| `database/mongodb-schema-design.md` | `cart_db.carts` document reference and indexes |

---

## 🪜 Step-by-Step Implementation

## Step 1: Task Boundary Clear Kiya

Cart Service Task 4 ka kaam **AddItem mutation flow** define karna hai.

### Included

- `POST /api/v1/cart/items` and `CartService.AddItem` contract explain karna
- Request owner resolve karna: logged-in `user_id` ya guest `guest_session_id`
- Request validation: `product_id`, `variant_id`, `quantity`
- Product Service se product/variant validation
- Product publish/sellable status check
- Stock check before cart update
- Existing cart find/create karna
- Same `product_id + variant_id` line item ho to quantity increment/update karna
- New variant ho to embedded cart item add karna
- Price/title/image/variant snapshot store karna
- Cart totals recalculate karna
- MongoDB durable save with optimistic versioning
- Redis full cart and summary cache refresh
- Error handling, test cases, and future implementation structure

### Not Included

- Item remove karna: Task 5
- Quantity zero cleanup: Task 5
- Coupon validation/discount preview: Task 6
- Guest cart merge after login: Task 7
- Inactive cart cleanup scheduler: Task 8
- Checkout inventory reservation: Order Service task
- Actual `backend/services/cart-service` files create karna

> 🟢 **Reason:** `docs/01-micro-tasks.md` ke according Task 4 only Add Item flow hai. Remove, coupon, merge, expiry, aur checkout alag tasks me aayenge.

---

## Step 2: Public API Contract Samjha

Add item ka public REST route API Gateway ke through expose hoga.

| Layer | Contract |
|---|---|
| REST | `POST /api/v1/cart/items` |
| Internal gRPC | `CartService.AddItem` |
| Request schema | `CartItemInput` |
| Response schema | `Cart` |
| Auth context | Buyer/user context, ya guest session context |

### Request Body

```json
{
  "product_id": "prod_123",
  "variant_id": "var_black_9",
  "quantity": 2
}
```

### Response Body

```json
{
  "cart_id": "cart_user_123",
  "user_id": "user_123",
  "items": [
    {
      "item_id": "item_123",
      "product_id": "prod_123",
      "variant_id": "var_black_9",
      "title_snapshot": "Running Shoes",
      "image_url_snapshot": "https://cdn.example.com/prod_123/main.jpg",
      "quantity": 2,
      "unit_price": {
        "amount": 299900,
        "currency": "INR"
      },
      "line_subtotal": {
        "amount": 599800,
        "currency": "INR"
      }
    }
  ],
  "subtotal": {
    "amount": 599800,
    "currency": "INR"
  },
  "discount": {
    "amount": 0,
    "currency": "INR"
  },
  "total": {
    "amount": 599800,
    "currency": "INR"
  }
}
```

### Beginner Explanation

Frontend sirf product id, variant id, aur quantity bhejega. Cart Service khud Product Service se product ka latest data fetch karega. Isse user request ke andar fake price/title bhejne ka risk nahi rahega.

---

## Step 3: High-Level Add Item Flow Design Kiya

Add item flow me Cart Service three systems se kaam karega:

1. **Product Service**: product, variant, price, stock validate karne ke liye
2. **MongoDB**: durable cart save karne ke liye
3. **Redis**: updated cart and summary cache karne ke liye

```mermaid
sequenceDiagram
    participant Client as Frontend / API Gateway
    participant Cart as Cart Service
    participant Product as Product Service
    participant Mongo as MongoDB cart_db.carts
    participant Redis as Redis Hot Cache

    Client->>Cart: AddItem(product_id, variant_id, quantity)
    Cart->>Cart: Resolve owner + validate request
    Cart->>Product: GetProduct(product_id)
    Product-->>Cart: Product + variants + stock + price
    Cart->>Cart: Validate variant, status, stock, quantity limit
    Cart->>Mongo: Find or create active cart
    Cart->>Cart: Add new item or increment existing item
    Cart->>Cart: Recalculate totals
    Cart->>Mongo: Save updated cart with version guard
    Mongo-->>Cart: Updated cart saved
    Cart->>Redis: Refresh active cart + summary cache
    Redis-->>Cart: OK or warning
    Cart-->>Client: Updated Cart
```

### Core Decision

```text
AddItem does not trust client price.
AddItem trusts Product Service snapshot.
AddItem saves durable cart in MongoDB.
AddItem refreshes Redis after MongoDB success.
```

---

## Step 4: Request Validation Rules Banaye

Cart Service ko Product Service call karne se pehle basic request validate karna chahiye.

| Field | Rule | Error |
|---|---|---|
| `product_id` | Required, non-empty string | `INVALID_ARGUMENT` |
| `variant_id` | Required, non-empty string | `INVALID_ARGUMENT` |
| `quantity` | Required integer | `INVALID_ARGUMENT` |
| `quantity` minimum | `1` | `INVALID_ARGUMENT` |
| `quantity` maximum per request | `10` | `INVALID_ARGUMENT` |
| Owner | `user_id` or `guest_session_id` required | `UNAUTHENTICATED` / `PERMISSION_DENIED` |

### Quantity Rule

Task 1 ke rule ke according:

```text
Minimum quantity per item = 1
Maximum quantity per item = 10
Maximum unique items per cart = 100
```

### Validation Code Example

```go
const (
	MaxQuantityPerItem = 10
	MaxUniqueCartItems = 100
)

func ValidateAddItemInput(productID, variantID string, quantity int) error {
	if productID == "" {
		return ErrProductIDRequired
	}
	if variantID == "" {
		return ErrVariantIDRequired
	}
	if quantity < 1 {
		return ErrQuantityTooSmall
	}
	if quantity > MaxQuantityPerItem {
		return ErrQuantityTooLarge
	}
	return nil
}
```

> 🟡 **Note:** Ye example future `internal/usecase/add_item.go` ke liye reference hai. Is task me actual Go file create nahi ki gayi.

---

## Step 5: Owner Resolution Define Kiya

Add item request hamesha kisi cart owner se linked honi chahiye.

| Request Context | Owner Type | Lookup |
|---|---|---|
| Logged-in buyer | `user` | `user_id` |
| Guest buyer | `guest` | `guest_session_id` |
| Dono present | `user` primary, guest merge later Task 7 | User active cart |
| Dono missing | Invalid | Reject |

### Owner Model Example

```go
type CartOwner struct {
	Type           string // "user" or "guest"
	UserID         string
	GuestSessionID string
}

func ResolveCartOwner(userID, guestSessionID string) (CartOwner, error) {
	if userID != "" {
		return CartOwner{Type: "user", UserID: userID}, nil
	}
	if guestSessionID != "" {
		return CartOwner{Type: "guest", GuestSessionID: guestSessionID}, nil
	}
	return CartOwner{}, ErrCartOwnerMissing
}
```

### Beginner Explanation

Cart bina owner ke create nahi karna. Agar user login hai to cart user ke account se attach hoga. Agar login nahi hai to cart guest session se attach hoga.

---

## Step 6: Product Service Validation Design Kiya

Task 4 ka dependency Product Service hai, kyunki Cart Service ko product data khud se assume nahi karna.

### Product Service Se Required Data

| Data | Why needed |
|---|---|
| `product_id` | Requested product exist karta hai ya nahi |
| `seller_id` | Cart item snapshot and future seller/order grouping |
| `title` | Cart page display snapshot |
| `images[0]` | Cart item thumbnail snapshot |
| `status` | Sirf published/active product add hona chahiye |
| `variant_id` | Requested variant exist karta hai ya nahi |
| `sku` | Debug/display snapshot |
| `variant attributes` | Size/color jaise cart display fields |
| `price` | Client price trust nahi karna |
| `stock_quantity` | Add karne se pehle stock check |

### Product Validation Rules

| Check | Decision |
|---|---|
| Product missing | Reject with `NOT_FOUND` |
| Product unpublished/inactive | Reject with `FAILED_PRECONDITION` |
| Variant missing | Reject with `NOT_FOUND` |
| Variant disabled/unavailable | Reject with `FAILED_PRECONDITION` |
| Stock less than final cart quantity | Reject with `FAILED_PRECONDITION` |
| Price missing or currency missing | Reject with `FAILED_PRECONDITION` |

### Important Stock Decision

```text
AddItem stock check karega, but stock reserve nahi karega.
Final inventory reservation checkout/order flow me hoga.
```

Why? Agar add-to-cart par inventory reserve karenge, to carts inventory block kar sakte hain. Cart me stock warning/validation enough hai. Checkout ke time Order Service fresh price and stock validate karega.

### Product Client Interface Example

```go
type ProductClient interface {
	GetProductForCart(ctx context.Context, productID string) (*ProductForCart, error)
}

type ProductForCart struct {
	ID       string
	SellerID string
	Title    string
	ImageURL string
	Status   string
	Variants []ProductVariantForCart
}

type ProductVariantForCart struct {
	ID            string
	SKU           string
	Attributes    map[string]string
	PriceAmount   int64
	Currency      string
	StockQuantity int
	Available     bool
}
```

> 🔵 **Contract note:** `api/master-api.json` me Cart request `variant_id` expect karta hai. Product Service response/proto implementation ko variant identity expose karni hogi, chahe current high-level schema simplified ho.

---

## Step 7: Active Cart Find-Or-Create Flow Banaya

Add item ke time agar active cart already hai, usko update karna hai. Agar active cart nahi hai, new active cart create karna hai.

```mermaid
flowchart TD
    A[AddItem request] --> B{Owner type?}
    B -->|user| C[Find active cart by user_id]
    B -->|guest| D[Find active cart by guest_session_id]
    C --> E{Cart found?}
    D --> E
    E -->|Yes| F[Use existing active cart]
    E -->|No| G[Create empty active cart]
    F --> H[Apply item mutation]
    G --> H
    H --> I[Save cart]

    classDef start fill:#e0f2fe,stroke:#0284c7,color:#0f172a
    classDef decision fill:#fef3c7,stroke:#d97706,color:#0f172a
    classDef action fill:#dcfce7,stroke:#16a34a,color:#0f172a

    class A start
    class B,E decision
    class C,D,F,G,H,I action
```

### New Cart Defaults

```json
{
  "_id": "cart_generated_id",
  "user_id": "user_123",
  "guest_session_id": null,
  "status": "active",
  "items": [],
  "coupon_code": null,
  "coupon_preview": null,
  "totals": {
    "subtotal": 0,
    "discount": 0,
    "total": 0,
    "currency": "INR"
  },
  "version": 1,
  "created_at": "2026-05-24T00:00:00Z",
  "updated_at": "2026-05-24T00:00:00Z",
  "expires_at": "2026-08-22T00:00:00Z"
}
```

### MongoDB Safety

Task 3 me one active cart indexes define kiye gaye:

```javascript
db.carts.createIndex(
  { user_id: 1, status: 1 },
  {
    name: "uniq_active_cart_per_user",
    unique: true,
    partialFilterExpression: { status: "active", user_id: { $type: "string" } }
  }
)

db.carts.createIndex(
  { guest_session_id: 1, status: 1 },
  {
    name: "uniq_active_cart_per_guest_session",
    unique: true,
    partialFilterExpression: { status: "active", guest_session_id: { $type: "string" } }
  }
)
```

If two requests same time new cart create karne ki koshish karein, unique index duplicate active cart block karega. Implementation duplicate key error par existing active cart refetch karegi.

---

## Step 8: Add or Update Item Rule Implement Kiya

Same product variant ko cart me duplicate line item nahi banana.

| Condition | Action |
|---|---|
| Same `product_id + variant_id` already exists | Existing item quantity increment karo |
| Same `product_id`, different `variant_id` | New line item add karo |
| Cart has 100 unique items and new item add ho raha hai | Reject |
| Final quantity > 10 | Reject |
| Final quantity > stock | Reject |

### Final Quantity Formula

```text
existing_quantity = current cart line quantity, or 0
requested_quantity = AddItem request quantity
final_quantity = existing_quantity + requested_quantity
```

### Example

| Existing Quantity | Request Quantity | Final Quantity | Result |
|---:|---:|---:|---|
| 0 | 2 | 2 | New item add |
| 2 | 1 | 3 | Existing item update |
| 8 | 2 | 10 | Allowed |
| 8 | 3 | 11 | Reject: max quantity exceeded |

### Domain Code Example

```go
func (c *Cart) AddOrIncrementItem(snapshot ProductVariantSnapshot, addQty int, now time.Time, newItemID string) error {
	if c.Status != CartStatusActive {
		return ErrCartNotActive
	}

	for i := range c.Items {
		item := &c.Items[i]
		if item.ProductID == snapshot.ProductID && item.VariantID == snapshot.VariantID {
			finalQty := item.Quantity + addQty
			if finalQty > MaxQuantityPerItem {
				return ErrQuantityTooLarge
			}
			if finalQty > snapshot.StockQuantity {
				return ErrInsufficientStock
			}

			item.Quantity = finalQty
			item.UnitPrice = snapshot.UnitPrice
			item.LineSubtotal = Money{
				Amount:   snapshot.UnitPrice.Amount * int64(finalQty),
				Currency: snapshot.UnitPrice.Currency,
			}
			item.TitleSnapshot = snapshot.Title
			item.ImageURLSnapshot = snapshot.ImageURL
			item.VariantSnapshot = snapshot.Attributes
			item.PriceSnapshotAt = now
			item.UpdatedAt = now
			return nil
		}
	}

	if len(c.Items) >= MaxUniqueCartItems {
		return ErrTooManyCartItems
	}
	if addQty > snapshot.StockQuantity {
		return ErrInsufficientStock
	}

	c.Items = append(c.Items, CartItem{
		ItemID:           newItemID,
		ProductID:        snapshot.ProductID,
		VariantID:        snapshot.VariantID,
		SellerID:         snapshot.SellerID,
		SKUSnapshot:      snapshot.SKU,
		TitleSnapshot:    snapshot.Title,
		ImageURLSnapshot: snapshot.ImageURL,
		VariantSnapshot:  snapshot.Attributes,
		UnitPrice:        snapshot.UnitPrice,
		Quantity:         addQty,
		LineSubtotal: Money{
			Amount:   snapshot.UnitPrice.Amount * int64(addQty),
			Currency: snapshot.UnitPrice.Currency,
		},
		PriceSnapshotAt: now,
		AddedAt:         now,
		UpdatedAt:       now,
	})

	return nil
}
```

> 🟣 **Important:** Existing item add karne par quantity replace nahi hoti; increment hoti hai. Quantity replace/update separate `UpdateItemQuantity` flow hai.

---

## Step 9: Price Snapshot Build Kiya

Cart item me live product reference ke saath snapshot bhi store hoga.

### Why Snapshot?

| Snapshot Field | Why useful |
|---|---|
| `title_snapshot` | Product title later change ho to cart UI stable rahe |
| `image_url_snapshot` | Cart thumbnail fast render ho |
| `unit_price` | Add-time price visible rahe |
| `variant_snapshot` | Size/color display ho |
| `price_snapshot_at` | Checkout se pehle stale price detect kar sake |

### Snapshot Mapping

```go
func BuildProductVariantSnapshot(product ProductForCart, variant ProductVariantForCart) ProductVariantSnapshot {
	return ProductVariantSnapshot{
		ProductID:     product.ID,
		VariantID:     variant.ID,
		SellerID:      product.SellerID,
		SKU:           variant.SKU,
		Title:         product.Title,
		ImageURL:      product.ImageURL,
		Attributes:    variant.Attributes,
		StockQuantity: variant.StockQuantity,
		UnitPrice: Money{
			Amount:   variant.PriceAmount,
			Currency: variant.Currency,
		},
	}
}
```

### Money Rule

Money integer minor units me store hoga:

```text
INR 2999.00 => amount: 299900
```

Floating point use nahi karna, because money rounding bugs create ho sakte hain.

---

## Step 10: Totals Recalculate Kiya

Add item ke baad totals hamesha server side recalculate honge.

### Task 4 Totals Scope

| Field | Task 4 Behavior |
|---|---|
| `subtotal` | Sum of all `line_subtotal.amount` |
| `discount` | `0` because coupon preview Task 6 me hai |
| `total` | `subtotal - discount` |
| `currency` | Cart ke saare items same currency me hone chahiye |

### Totals Code Example

```go
func (c *Cart) RecalculateTotals(defaultCurrency string) error {
	var subtotal int64
	currency := defaultCurrency

	for _, item := range c.Items {
		if item.LineSubtotal.Currency == "" {
			return ErrCurrencyMissing
		}
		if currency == "" {
			currency = item.LineSubtotal.Currency
		}
		if item.LineSubtotal.Currency != currency {
			return ErrMixedCurrencyCart
		}
		subtotal += item.LineSubtotal.Amount
	}

	c.Totals = CartTotals{
		Subtotal: Money{Amount: subtotal, Currency: currency},
		Discount: Money{Amount: 0, Currency: currency},
		Total:    Money{Amount: subtotal, Currency: currency},
		Currency: currency,
	}
	return nil
}
```

> 🟡 **Future note:** Coupon preview Task 6 ke baad AddItem flow coupon preview ko clear ya revalidate karega. Task 4 me discount calculation implement nahi hota.

---

## Step 11: MongoDB Save Strategy Define Ki

MongoDB Cart Service ka durable source of truth hai.

### Recommended Repository Methods

```go
type CartRepository interface {
	FindActiveByOwner(ctx context.Context, owner CartOwner) (*Cart, error)
	CreateActiveCart(ctx context.Context, cart *Cart) error
	SaveWithVersion(ctx context.Context, cart *Cart, expectedVersion int64) error
}
```

### Optimistic Versioning

Task 3 me `version` field define hua tha. AddItem concurrent tabs se call ho sakta hai, isliye save ke time version guard use karna chahiye.

```javascript
filter := bson.M{
  "_id": cart.ID,
  "status": "active",
  "version": expectedVersion
}

update := bson.M{
  "$set": bson.M{
    "items": cart.Items,
    "totals": cart.Totals,
    "updated_at": cart.UpdatedAt,
    "expires_at": cart.ExpiresAt
  },
  "$inc": bson.M{
    "version": 1
  }
}
```

### Save Flow

```mermaid
flowchart TD
    A[Load active cart with version N] --> B[Apply AddItem domain change]
    B --> C[Save where _id and version N]
    C --> D{Matched one document?}
    D -->|Yes| E[Save success, version becomes N+1]
    D -->|No| F[Version conflict]
    F --> G[Reload latest cart]
    G --> H{Retry left?}
    H -->|Yes| B
    H -->|No| I[Return conflict error]

    classDef ok fill:#dcfce7,stroke:#16a34a,color:#0f172a
    classDef warn fill:#fef3c7,stroke:#d97706,color:#0f172a
    classDef err fill:#fee2e2,stroke:#dc2626,color:#0f172a

    class E ok
    class F,G,H warn
    class I err
```

### Why optimistic locking?

- Multiple browser tabs same cart mutate kar sakte hain.
- Same user mobile and desktop se item add kar sakta hai.
- Version guard lost updates prevent karta hai.
- Retry user ko unnecessary error dikhane se bachata hai.

---

## Step 12: Redis Cache Refresh Define Kiya

MongoDB save successful hone ke baad Redis refresh hoga.

### Redis Keys

| Key | Value | TTL |
|---|---|---|
| `cart:active:user:{user_id}` | Full cart JSON | `15m` |
| `cart:active:guest:{guest_session_id}` | Full cart JSON | `30m` |
| `cart:summary:user:{user_id}` | Item count + total | `5m` |
| `cart:summary:guest:{guest_session_id}` | Item count + total | `5m` |

### Cache Refresh Rule

```text
MongoDB save success => Redis refresh
Redis refresh failure => log warning, still return updated cart
```

### Cache Code Example

```go
type CartCache interface {
	SetActiveCart(ctx context.Context, owner CartOwner, cart *Cart, ttl time.Duration) error
	SetCartSummary(ctx context.Context, owner CartOwner, summary CartSummary, ttl time.Duration) error
}

func BuildCartSummary(cart *Cart) CartSummary {
	var itemCount int
	for _, item := range cart.Items {
		itemCount += item.Quantity
	}
	return CartSummary{
		ItemCount: itemCount,
		Total:     cart.Totals.Total,
	}
}
```

### Beginner Explanation

MongoDB me original cart save hota hai. Redis me fast copy save hoti hai. Agar Redis fail bhi ho gaya, cart data lose nahi hoga because MongoDB already update ho chuka hai.

---

## Step 13: AddItem Usecase Structure Banaya

Future backend implementation me `AddItemUsecase` orchestration layer hoga.

### Usecase Dependencies

```go
type AddItemUsecase struct {
	repo          CartRepository
	cache         CartCache
	productClient ProductClient
	ids           IDGenerator
	clock         Clock
	logger        Logger
}
```

### Command Model

```go
type AddItemCommand struct {
	UserID         string
	GuestSessionID string
	ProductID      string
	VariantID      string
	Quantity       int
}
```

### Full Usecase Pseudocode

```go
func (uc *AddItemUsecase) Execute(ctx context.Context, cmd AddItemCommand) (*Cart, error) {
	if err := ValidateAddItemInput(cmd.ProductID, cmd.VariantID, cmd.Quantity); err != nil {
		return nil, err
	}

	owner, err := ResolveCartOwner(cmd.UserID, cmd.GuestSessionID)
	if err != nil {
		return nil, err
	}

	product, err := uc.productClient.GetProductForCart(ctx, cmd.ProductID)
	if err != nil {
		return nil, err
	}

	variant, err := FindSellableVariant(product, cmd.VariantID)
	if err != nil {
		return nil, err
	}

	return uc.withRetry(ctx, func() (*Cart, error) {
		now := uc.clock.Now()
		cart, err := uc.repo.FindActiveByOwner(ctx, owner)
		if errors.Is(err, ErrCartNotFound) {
			cart = NewActiveCart(uc.ids.NewCartID(), owner, now)
			if err := uc.repo.CreateActiveCart(ctx, cart); err != nil {
				return nil, err
			}
		} else if err != nil {
			return nil, err
		}

		expectedVersion := cart.Version
		snapshot := BuildProductVariantSnapshot(*product, *variant)

		if err := cart.AddOrIncrementItem(snapshot, cmd.Quantity, now, uc.ids.NewItemID()); err != nil {
			return nil, err
		}
		if err := cart.RecalculateTotals(snapshot.UnitPrice.Currency); err != nil {
			return nil, err
		}

		cart.Touch(now)

		if err := uc.repo.SaveWithVersion(ctx, cart, expectedVersion); err != nil {
			return nil, err
		}

		summary := BuildCartSummary(cart)
		if err := uc.refreshCache(ctx, owner, cart, summary); err != nil {
			uc.logger.Warn("cart cache refresh failed", "error", err)
		}

		return cart, nil
	})
}
```

> 🟠 **Scope note:** Ye pseudocode implementation thinking clear karne ke liye hai. Is task me actual `AddItemUsecase` source file create nahi ki gayi.

---

## Step 14: Error Handling Matrix Banaya

| Scenario | gRPC Code | HTTP Status | User-facing Meaning |
|---|---|---:|---|
| Missing auth/session owner | `UNAUTHENTICATED` | `401` | Cart owner identify nahi hua |
| Missing `product_id` | `INVALID_ARGUMENT` | `400` | Product id required hai |
| Missing `variant_id` | `INVALID_ARGUMENT` | `400` | Variant id required hai |
| Quantity `< 1` | `INVALID_ARGUMENT` | `400` | Quantity at least 1 honi chahiye |
| Quantity `> 10` | `INVALID_ARGUMENT` | `400` | Per item max quantity exceeded |
| Product not found | `NOT_FOUND` | `404` | Product available nahi hai |
| Product unpublished | `FAILED_PRECONDITION` | `409` | Product sellable nahi hai |
| Variant not found | `NOT_FOUND` | `404` | Variant available nahi hai |
| Stock insufficient | `FAILED_PRECONDITION` | `409` | Requested quantity stock se zyada hai |
| Cart already has 100 unique items | `FAILED_PRECONDITION` | `409` | Cart item limit reached |
| MongoDB unavailable | `UNAVAILABLE` | `503` | Cart temporarily unavailable |
| Product Service unavailable | `UNAVAILABLE` | `503` | Product validation temporarily unavailable |
| Version conflict after retries | `ABORTED` | `409` | Concurrent cart update conflict |

### Error Response Example

```json
{
  "code": "INSUFFICIENT_STOCK",
  "message": "Requested quantity is not available for this variant.",
  "details": {
    "product_id": "prod_123",
    "variant_id": "var_black_9",
    "available_quantity": 3
  }
}
```

---

## Step 15: Architecture Diagram Banaya

```mermaid
flowchart LR
    UI[User App Frontend] --> GW[API Gateway]
    GW -->|POST /api/v1/cart/items| Cart[Cart Service]
    Cart -->|GetProduct| Product[Product Service]
    Cart -->|Read/Write active cart| Mongo[(MongoDB cart_db.carts)]
    Cart -->|Refresh active cart cache| Redis[(Redis)]

    Cart --> Rules[Cart Domain Rules]
    Rules --> Qty[Quantity Limit]
    Rules --> Snapshot[Price Snapshot]
    Rules --> Totals[Totals Recalculate]

    classDef client fill:#e0f2fe,stroke:#0284c7,color:#0f172a
    classDef svc fill:#f3e8ff,stroke:#7c3aed,color:#0f172a
    classDef db fill:#dcfce7,stroke:#16a34a,color:#0f172a
    classDef rule fill:#fef3c7,stroke:#d97706,color:#0f172a

    class UI,GW client
    class Cart,Product svc
    class Mongo,Redis db
    class Rules,Qty,Snapshot,Totals rule
```

### Architecture Explanation

- API Gateway public REST route receive karega.
- Gateway auth/session metadata Cart Service ko pass karega.
- Cart Service Product Service se trusted product snapshot lega.
- Cart Service MongoDB me durable cart update karega.
- Cart Service Redis me hot cache refresh karega.
- Response me updated cart frontend ko milega.

---

## Step 16: Folder Structure Define Kiya

### Actual TaskImplementation Structure

```text
TaskImplementation/
└── Cart Service/
    ├── task1.md
    ├── task2.md
    ├── task3.md
    └── task4.md
```

### Future Cart Service Code Structure

```text
backend/
└── services/
    └── cart-service/
        ├── cmd/
        │   └── server/
        │       └── main.go
        ├── internal/
        │   ├── domain/
        │   │   ├── cart.go
        │   │   ├── cart_item.go
        │   │   └── money.go
        │   ├── usecase/
        │   │   ├── add_item.go
        │   │   ├── remove_item.go
        │   │   ├── merge_cart.go
        │   │   └── price_cart.go
        │   ├── repository/
        │   │   ├── mongo_cart_repository.go
        │   │   └── redis_cart_cache.go
        │   ├── client/
        │   │   └── product_client.go
        │   └── transport/
        │       └── grpc/
        │           └── server.go
        └── deploy/
            └── Dockerfile
```

### Task 4 Files In Future Implementation

| File | Responsibility |
|---|---|
| `internal/usecase/add_item.go` | AddItem orchestration |
| `internal/domain/cart.go` | Cart aggregate rules |
| `internal/domain/cart_item.go` | Embedded item behavior |
| `internal/domain/money.go` | Money operations without float |
| `internal/client/product_client.go` | Product Service gRPC adapter |
| `internal/repository/mongo_cart_repository.go` | Active cart find/create/save |
| `internal/repository/redis_cart_cache.go` | Cache refresh |
| `internal/transport/grpc/server.go` | `CartService.AddItem` gRPC handler |

> 🔴 **Important:** Ye future code structure hai. Current task output sirf `task4.md` documentation file hai.

---

## Step 17: External Libraries and Tools

Task 4 ke actual future implementation me ye libraries/tools use honge:

| Tool / Library | What it is | Why used | Install / Use |
|---|---|---|---|
| MongoDB `mongo:7` | Document database | Durable `cart_db.carts` source of truth | Docker Compose local stack se run |
| Redis `redis:7.2-alpine` | In-memory cache | Active cart and summary cache | Docker Compose local stack se run |
| `go.mongodb.org/mongo-driver` | Official MongoDB Go driver | Cart repository read/write ke liye | `go get go.mongodb.org/mongo-driver/mongo` |
| `github.com/redis/go-redis/v9` | Redis Go client | Cache GET/SET/DEL ke liye | `go get github.com/redis/go-redis/v9` |
| `github.com/google/uuid` | UUID generator | `cart_id` and `item_id` generate karne ke liye | `go get github.com/google/uuid` |
| `google.golang.org/grpc` | gRPC framework | Product Service and Cart Service internal calls | `go get google.golang.org/grpc` |
| Protobuf generated code | Typed service contracts | `CartService.AddItem` and Product Service client stubs | Proto generation pipeline from Platform Foundation |

### Example Install Commands

```bash
cd backend/services/cart-service
go get go.mongodb.org/mongo-driver/mongo
go get github.com/redis/go-redis/v9
go get github.com/google/uuid
go get google.golang.org/grpc
```

### Example Environment Variables

```env
CART_SERVICE_NAME=cart-service
CART_GRPC_ADDR=:9090
MONGO_URI=mongodb://ecommerce_root:ecommerce_password@mongo:27017/cart_db?authSource=admin
MONGO_DATABASE=cart_db
REDIS_ADDR=redis:6379
REDIS_DB=0
PRODUCT_GRPC_ADDR=product-service:9090
CART_ACTIVE_CACHE_TTL=15m
CART_SUMMARY_CACHE_TTL=5m
GUEST_CART_CACHE_TTL=30m
```

---

## Step 18: Testing Strategy Banayi

Task 4 implementation high-risk hai because user-facing cart mutation hai. Tests focused hone chahiye.

### Unit Tests

| Test Case | Expected Result |
|---|---|
| Missing `product_id` | Validation error |
| Missing `variant_id` | Validation error |
| Quantity `0` | Validation error |
| Quantity `11` | Validation error |
| Product not found | Not found error |
| Product unpublished | Precondition error |
| Variant not found | Not found error |
| Stock less than requested | Insufficient stock error |
| Empty cart add item | New cart item created |
| Existing same variant add item | Quantity incremented |
| Different variant add item | New line item created |
| Final quantity above 10 | Reject |
| Cart has 100 unique items | Reject new line item |
| Mixed currency product | Reject |
| Redis refresh fails | Cart still returned after Mongo save |

### Integration Tests

| Test | Setup |
|---|---|
| Mongo find/create/save | MongoDB test container or local stack |
| Redis cache refresh | Redis test container or local stack |
| Product Service validation | Fake gRPC Product Service |
| Concurrent AddItem | Two goroutines adding same item, version retry expected |

### Example Unit Test Name List

```text
TestAddItemCreatesActiveCartWhenMissing
TestAddItemIncrementsExistingVariant
TestAddItemRejectsQuantityAboveLimit
TestAddItemRejectsOutOfStockVariant
TestAddItemRetriesOnVersionConflict
TestAddItemReturnsCartWhenRedisRefreshFails
```

---

## Step 19: Observability Notes Add Kiye

AddItem flow production me frequently call hoga, isliye metrics/logs useful rahenge.

### Logs

| Event | Fields |
|---|---|
| Add item started | `owner_type`, `product_id`, `variant_id`, `quantity` |
| Product validation failed | `reason`, `product_id`, `variant_id` |
| Stock failed | `requested_quantity`, `available_quantity` |
| Cart save conflict | `cart_id`, `expected_version` |
| Cache refresh failed | `cart_id`, `redis_error` |

### Metrics

| Metric | Meaning |
|---|---|
| `cart_add_item_total` | Add item attempts count |
| `cart_add_item_success_total` | Successful add item count |
| `cart_add_item_error_total` | Errors by reason |
| `cart_add_item_duration_ms` | End-to-end latency |
| `cart_cache_refresh_error_total` | Redis refresh failures |

### Trace Spans

```text
CartService.AddItem
├── validate_request
├── product_client.get_product
├── cart_repository.find_or_create_active
├── cart_domain.add_or_increment_item
├── cart_repository.save_with_version
└── cart_cache.refresh
```

---

## Step 20: Complete AddItem Flow Summary

```mermaid
flowchart TD
    A[Receive AddItem] --> B[Validate input]
    B --> C[Resolve user or guest owner]
    C --> D[Fetch product from Product Service]
    D --> E{Product sellable?}
    E -->|No| X[Reject]
    E -->|Yes| F{Variant exists and available?}
    F -->|No| X
    F -->|Yes| G[Find or create active cart]
    G --> H[Calculate final quantity]
    H --> I{Within qty and stock limits?}
    I -->|No| X
    I -->|Yes| J[Add new item or increment existing line]
    J --> K[Recalculate totals]
    K --> L[Save MongoDB with version guard]
    L --> M{Save conflict?}
    M -->|Yes| N[Reload and retry]
    N --> H
    M -->|No| O[Refresh Redis cache]
    O --> P[Return updated cart]

    classDef start fill:#e0f2fe,stroke:#0284c7,color:#0f172a
    classDef action fill:#dcfce7,stroke:#16a34a,color:#0f172a
    classDef decision fill:#fef3c7,stroke:#d97706,color:#0f172a
    classDef error fill:#fee2e2,stroke:#dc2626,color:#0f172a

    class A start
    class B,C,D,G,H,J,K,L,N,O,P action
    class E,F,I,M decision
    class X error
```

### Final Rules

1. Client price trusted nahi hoga.
2. Product Service validation mandatory hai.
3. Product unpublished/inactive ho to cart me add nahi hoga.
4. Variant missing/unavailable ho to add nahi hoga.
5. Stock check final cart quantity ke against hoga.
6. Same `product_id + variant_id` duplicate line nahi banegi.
7. Existing item par AddItem quantity increment karega.
8. Per item max quantity `10` rahegi.
9. New cart item limit `100` unique items rahegi.
10. MongoDB source of truth rahega.
11. Redis mutation ke baad refresh hoga.
12. Redis failure cart mutation ko fail nahi karega.
13. Checkout ke time final stock and price again validate honge.

---

## ✅ Task 4 Completion Checklist

| Requirement | Status |
|---|---|
| `TaskImplementation/` folder present | ✅ Done |
| `TaskImplementation/Cart Service/` folder present | ✅ Done |
| `task4.md` created | ✅ Done |
| Step-by-step implementation in Hinglish | ✅ Done |
| Product validation explained | ✅ Done |
| Stock check explained | ✅ Done |
| Add/update item rules explained | ✅ Done |
| MongoDB save strategy explained | ✅ Done |
| Redis cache refresh explained | ✅ Done |
| External tools/libraries listed | ✅ Done |
| Install/use commands added | ✅ Done |
| Folder structure added | ✅ Done |
| Code examples added | ✅ Done |
| Mermaid diagrams added | ✅ Done |
| Scope limited to Cart Service Task 4 | ✅ Done |

---

## 🚫 Explicitly Not Implemented in This Task

| Not Implemented | Reason |
|---|---|
| Actual Go backend files | User output requirement is TaskImplementation markdown artifact only |
| Remove item flow | Cart Service Task 5 |
| Coupon preview | Cart Service Task 6 |
| Guest cart merge | Cart Service Task 7 |
| Cart expiry job | Cart Service Task 8 |
| Inventory reservation | Order/checkout flow |
| Payment/order creation | Order and Payment services |
| Frontend cart page | User App Frontend task |

---

## 🧾 Final Outcome

Cart Service Task 4 ka final documented implementation standard:

```text
Frontend/API Gateway
  -> CartService.AddItem
  -> Validate owner and request
  -> Product Service validation
  -> Stock check
  -> Find/create active cart
  -> Add item or increment existing variant
  -> Recalculate totals
  -> Save MongoDB durable cart
  -> Refresh Redis hot cache
  -> Return updated cart
```

Task 4 complete hai as a structured, beginner-friendly Add Item flow guide. Future Cart Service implementation isi document ko usecase, repository, cache, and gRPC handler blueprint ki tarah follow karegi.
