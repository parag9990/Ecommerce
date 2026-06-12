# 🛒 Cart Service - Task 5: Remove Item Flow

![Task](https://img.shields.io/badge/Task-Cart%20Service%20%235-blue)
![Priority](https://img.shields.io/badge/Priority-P1-orange)
![Dependency](https://img.shields.io/badge/Dependency-Add%20Item-yellow)
![Status](https://img.shields.io/badge/Status-Documented-success)
![Guide](https://img.shields.io/badge/Guide-Hinglish-orange)
![Scope](https://img.shields.io/badge/Scope-Remove%20Item%20Flow-purple)

## 📌 Task Summary

| Field | Detail |
|---|---|
| Task Name | Remove item flow |
| Source | `docs/01-micro-tasks.md` → `Cart Service` → Task 5 |
| Priority | `P1` core business feature |
| Dependency | Cart Service Task 4: Add item flow |
| Main Goal | Cart se item remove karna, quantity zero cleanup karna, aur totals recalculate karna |
| Output Type | Documentation-only implementation guide |
| Not Included | Coupon preview, cart merge, cart expiry job, checkout inventory reservation, actual backend file creation |

> **Simple Hinglish goal:** Is task ka purpose Cart Service ke `RemoveItem` flow ko clearly design karna hai. Buyer jab cart se ek item remove karega, Cart Service active cart ownership validate karega, embedded `items[]` se matching `item_id` remove karega, zero/invalid quantity items cleanup karega, totals recalculate karega, MongoDB me updated cart save karega, Redis cache refresh karega, aur updated cart response return karega.

---

## ✅ Final Output Created

```text
TaskImplementation/
└── Cart Service/
    ├── task1.md
    ├── task2.md
    ├── task3.md
    ├── task4.md
    └── task5.md
```

### Why this structure?

- `TaskImplementation/` project ke task-wise implementation guides ka central folder hai.
- `Cart Service/` Cart Service ke saare task guides ko group karta hai.
- `task5.md` sirf **Cart Service - Task 5** ka Remove Item flow guide hai.
- Is task me sirf documentation artifact create kiya gaya hai; backend source code create nahi kiya gaya.

---

## 🧭 Implementation Approach

Is guide ko banane se pehle existing project documentation study ki gayi:

| Document | Kya use kiya gaya |
|---|---|
| `docs/01-micro-tasks.md` | Task 5 ka exact scope: item remove, quantity zero cleanup, totals recalculate |
| `TaskImplementation/Cart Service/task1.md` | Cart rules: active cart, owner validation, quantity rules, totals rules |
| `TaskImplementation/Cart Service/task2.md` | MongoDB source of truth and Redis hot cache strategy |
| `TaskImplementation/Cart Service/task3.md` | `cart_db.carts`, embedded `items[]`, totals, version field, indexes |
| `TaskImplementation/Cart Service/task4.md` | Add item flow, cache refresh, optimistic versioning, usecase structure |
| `docs/04-microservice-design.md` | Cart Service responsibilities, gRPC methods, REST routes |
| `docs/03-folder-structure.md` | Future `backend/services/cart-service` code structure |
| `api/master-api.json` | `DELETE /api/v1/cart/items/{item_id}`, `CartService.RemoveItem`, `Cart` response |
| `database/mongodb-schema-design.md` | Existing `cart_db.carts` collection reference |

---

## 🪜 Step-by-Step Implementation

## Step 1: Task Boundary Clear Kiya

Cart Service Task 5 ka kaam **RemoveItem mutation flow** define karna hai.

### Included

- `DELETE /api/v1/cart/items/{item_id}` and `CartService.RemoveItem` contract explain karna
- Request owner resolve karna: logged-in `user_id` ya guest `guest_session_id`
- `item_id` path parameter validate karna
- Active cart load karna
- Cart ownership and active status enforce karna
- Matching embedded item remove karna
- Quantity zero/negative cleanup rule define karna
- Empty cart ke totals zero karna
- Non-empty cart ke totals recalculate karna
- MongoDB durable save with optimistic versioning
- Redis active cart and summary cache refresh
- Error handling, tests, observability, and future code structure document karna

### Not Included

- Add item flow: already Task 4 me documented
- Product validation/stock check: RemoveItem me normally Product Service call required nahi hai
- Coupon preview validation: Task 6
- Guest cart merge after login: Task 7
- Inactive cart cleanup scheduler: Task 8
- Checkout inventory reservation: Order Service task
- Actual `backend/services/cart-service` source files create karna

> 🟢 **Reason:** `docs/01-micro-tasks.md` ke according Task 5 only remove flow, zero quantity cleanup, and totals recalculation hai. Coupon, merge, expiry, checkout, aur frontend implementation alag tasks me aayenge.

---

## Step 2: Public API Contract Samjha

Remove item ka public REST route API Gateway ke through expose hoga.

| Layer | Contract |
|---|---|
| REST | `DELETE /api/v1/cart/items/{item_id}` |
| Internal gRPC | `CartService.RemoveItem` |
| Request schema | `IdPathRequest` |
| Response schema | `Cart` |
| Auth context | Buyer/user context, ya guest session context |

### REST Request

```http
DELETE /api/v1/cart/items/item_123
Authorization: Bearer <access_token>
```

Guest cart ke case me Gateway guest session cookie/header se `guest_session_id` pass karega.

### gRPC Request Shape

```json
{
  "id": "item_123"
}
```

> 🔵 **Mapping note:** REST path ka `{item_id}` internal gRPC `IdPathRequest.id` me map hoga.

### Response Body

```json
{
  "cart_id": "cart_user_123",
  "user_id": "user_123",
  "items": [],
  "subtotal": {
    "amount": 0,
    "currency": "INR"
  },
  "discount": {
    "amount": 0,
    "currency": "INR"
  },
  "total": {
    "amount": 0,
    "currency": "INR"
  }
}
```

### Beginner Explanation

Frontend ko sirf `item_id` bhejna hai. Cart Service khud user/session ke active cart me item dhundhega. Agar item milta hai to remove hoga, totals update honge, aur updated cart return hoga.

---

## Step 3: High-Level Remove Item Flow Design Kiya

RemoveItem flow me Cart Service mainly two storage systems ke saath kaam karega:

1. **MongoDB**: durable active cart update karne ke liye
2. **Redis**: updated cart and summary fast cache karne ke liye

Product Service call normally required nahi hai, kyunki item remove karte time product details validate karne ki zarurat nahi hoti.

```mermaid
sequenceDiagram
    participant Client as Frontend / API Gateway
    participant Cart as Cart Service
    participant Mongo as MongoDB cart_db.carts
    participant Redis as Redis Hot Cache

    Client->>Cart: RemoveItem(item_id)
    Cart->>Cart: Resolve owner + validate item_id
    Cart->>Mongo: Find active cart by owner
    Mongo-->>Cart: Active cart with items[]
    Cart->>Cart: Remove matching item_id
    Cart->>Cart: Cleanup zero quantity items
    Cart->>Cart: Recalculate totals
    Cart->>Mongo: Save updated cart with version guard
    Mongo-->>Cart: Updated cart saved
    Cart->>Redis: Refresh active cart + summary cache
    Redis-->>Cart: OK or warning
    Cart-->>Client: Updated Cart
```

### Core Decision

```text
RemoveItem does not call Product Service.
RemoveItem only mutates the current owner's active cart.
RemoveItem keeps MongoDB as truth.
RemoveItem refreshes Redis after MongoDB success.
```

---

## Step 4: Request Validation Rules Banaye

Cart Service ko DB call se pehle basic request validate karna chahiye.

| Field | Rule | Error |
|---|---|---|
| `item_id` | Required, non-empty string | `INVALID_ARGUMENT` |
| Owner | `user_id` or `guest_session_id` required | `UNAUTHENTICATED` / `PERMISSION_DENIED` |
| HTTP method | `DELETE` only | Gateway route rule |
| Body | Not required | Ignore/reject unexpected mutation body as per Gateway policy |

### Validation Code Example

```go
func ValidateRemoveItemInput(itemID string) error {
	if itemID == "" {
		return ErrItemIDRequired
	}
	return nil
}
```

> 🟡 **Note:** Ye example future `internal/usecase/remove_item.go` ke liye reference hai. Is task me actual Go file create nahi ki gayi.

---

## Step 5: Owner Resolution Reuse Kiya

Task 4 me owner resolution define ho chuka hai. RemoveItem bhi same rule follow karega.

| Request Context | Owner Type | Lookup |
|---|---|---|
| Logged-in buyer | `user` | `user_id` |
| Guest buyer | `guest` | `guest_session_id` |
| Dono present | `user` primary | User active cart |
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

Remove request me user ko doosre user ke cart item ka access nahi milna chahiye. Isliye `item_id` se direct global delete nahi karna. Pehle owner ka active cart load hoga, phir usi cart ke andar item remove hoga.

---

## Step 6: Active Cart Load Rule Define Kiya

RemoveItem active cart ke against chalega. `merged`, `checked_out`, `expired`, ya `abandoned` cart mutate nahi hoga.

### Repository Lookup

```go
type CartRepository interface {
	FindActiveByOwner(ctx context.Context, owner CartOwner) (*Cart, error)
	SaveWithVersion(ctx context.Context, cart *Cart, expectedVersion int64) error
}
```

### Lookup Rules

| Case | Behavior |
|---|---|
| Active cart found | Continue remove flow |
| Active cart not found | Return `CART_NOT_FOUND` |
| Cart status not `active` | Return `CART_NOT_ACTIVE` |
| DB unavailable | Return `UNAVAILABLE` |

### MongoDB Query Example

```javascript
// Logged-in user cart
db.carts.findOne({
  user_id: "user_123",
  status: "active"
})

// Guest cart
db.carts.findOne({
  guest_session_id: "sess_abc123",
  status: "active"
})
```

> 🟣 **Important:** RemoveItem new cart create nahi karega. AddItem me find-or-create useful hai, but remove flow me cart missing hona stale/invalid state hai.

---

## Step 7: Item Remove Rule Implement Kiya

Cart items `carts.items[]` embedded array me stored hain. Remove operation `item_id` ke basis par hoga.

### Final Remove Rules

| Condition | Action |
|---|---|
| `item_id` current active cart me mil gaya | Item remove karo |
| `item_id` current active cart me nahi mila | Current cart return karo, mutation idempotent rakho |
| Cart empty ho gaya | Active empty cart keep karo, totals zero karo |
| Cart already empty hai | Empty cart return karo |
| Cart non-active hai | Reject |

### Why idempotent delete?

DELETE request double-click, network retry, ya stale frontend state ke wajah se repeat ho sakti hai. Agar item already removed hai, same user ko unnecessary error dikhane ki zarurat nahi. Isliye current active cart unchanged return karna beginner-friendly aur retry-safe behavior hai.

### Domain Code Example

```go
func (c *Cart) RemoveItem(itemID string, now time.Time) (bool, error) {
	if c.Status != CartStatusActive {
		return false, ErrCartNotActive
	}

	if itemID == "" {
		return false, ErrItemIDRequired
	}

	originalLen := len(c.Items)
	filtered := c.Items[:0]

	for _, item := range c.Items {
		if item.ItemID == itemID {
			continue
		}
		filtered = append(filtered, item)
	}

	c.Items = filtered
	removed := len(c.Items) != originalLen

	if removed {
		c.UpdatedAt = now
	}

	return removed, nil
}
```

> 🔵 **Code note:** `filtered := c.Items[:0]` Go me same backing array reuse karta hai. Ye simple and efficient hai, but implementation me readability prefer ho to new slice bhi use kar sakte hain.

---

## Step 8: Quantity Zero Cleanup Rule Banaya

Task 5 explicitly quantity zero cleanup mention karta hai. Cart document me kabhi bhi `quantity <= 0` item save nahi hona chahiye.

### Cleanup Rules

| Situation | Behavior |
|---|---|
| Item explicitly remove hua | Line item array se delete |
| Kisi internal update ke baad quantity `0` ho gayi | Item remove |
| Quantity negative milti hai | Item remove + warning log |
| Existing corrupted zero quantity item load hua | Save ke time cleanup |
| API `PATCH quantity=0` | Official API schema ke hisaab se reject; internal cleanup still defensive rahega |

### Why defensive cleanup?

API schema quantity minimum `1` rakhta hai, but bugs, old data, manual DB edits, ya future internal flows zero quantity create kar sakte hain. Save se pehle cleanup rule cart ko healthy state me rakhta hai.

### Cleanup Code Example

```go
func (c *Cart) CleanupZeroQuantityItems(now time.Time) int {
	cleaned := 0
	filtered := c.Items[:0]

	for _, item := range c.Items {
		if item.Quantity <= 0 {
			cleaned++
			continue
		}
		filtered = append(filtered, item)
	}

	c.Items = filtered
	if cleaned > 0 {
		c.UpdatedAt = now
	}

	return cleaned
}
```

### Beginner Explanation

Cart me quantity zero item rakhna confusing hota hai. User ko cart me product dikhe, but quantity zero ho, to total mismatch aur UI bugs aate hain. Isliye zero quantity ka meaning simple hai: item cart se remove.

---

## Step 9: Totals Recalculate Kiya

Remove ke baad totals hamesha server side recalculate honge.

### Task 5 Totals Scope

| Field | Task 5 Behavior |
|---|---|
| `subtotal` | Remaining items ke `line_subtotal.amount` ka sum |
| `discount` | `0` because coupon preview Task 6 me hai |
| `total` | `subtotal - discount` |
| `currency` | Remaining cart currency; empty cart me default currency |
| `item_count` | Remaining quantities ka sum |
| `unique_item_count` | Remaining `items.length` |

### Empty Cart Totals

```json
{
  "subtotal": {
    "amount": 0,
    "currency": "INR"
  },
  "discount": {
    "amount": 0,
    "currency": "INR"
  },
  "total": {
    "amount": 0,
    "currency": "INR"
  },
  "currency": "INR",
  "item_count": 0,
  "unique_item_count": 0
}
```

### Recalculate Code Example

```go
func (c *Cart) RecalculateTotals(defaultCurrency string) error {
	var subtotal int64
	var itemCount int
	currency := defaultCurrency

	for i := range c.Items {
		item := &c.Items[i]

		if item.UnitPrice.Currency == "" {
			return ErrCurrencyMissing
		}
		if currency == "" {
			currency = item.UnitPrice.Currency
		}
		if item.UnitPrice.Currency != currency {
			return ErrMixedCurrencyCart
		}

		item.LineSubtotal = Money{
			Amount:   item.UnitPrice.Amount * int64(item.Quantity),
			Currency: item.UnitPrice.Currency,
		}

		subtotal += item.LineSubtotal.Amount
		itemCount += item.Quantity
	}

	if currency == "" {
		currency = "INR"
	}

	c.Totals = CartTotals{
		Subtotal:        Money{Amount: subtotal, Currency: currency},
		Discount:        Money{Amount: 0, Currency: currency},
		Total:           Money{Amount: subtotal, Currency: currency},
		Currency:        currency,
		ItemCount:       itemCount,
		UniqueItemCount: len(c.Items),
	}

	return nil
}
```

> 🟡 **Future note:** Coupon preview Task 6 ke baad RemoveItem flow cart mutation ke baad coupon preview ko clear ya revalidate karega. Task 5 me actual coupon validation/discount calculation implement nahi hota.

---

## Step 10: Empty Cart Behavior Decide Kiya

RemoveItem ke baad agar last item remove ho jaye, cart document delete nahi hoga.

### Final Decision

```text
Last item remove => active empty cart keep karo
```

### Why?

| Reason | Explanation |
|---|---|
| UX simple | Frontend updated empty cart display kar sakta hai |
| Cart identity stable | Same cart id future add operations ke liye continue ho sakta hai |
| Audit/debug easy | Cart creation/removal history preserve hoti hai |
| Expiry separate task | Empty/inactive cart cleanup Task 8 me handle hoga |

### Empty Cart Document Example

```json
{
  "_id": "cart_user_123",
  "user_id": "user_123",
  "guest_session_id": null,
  "status": "active",
  "items": [],
  "coupon_code": null,
  "coupon_preview": null,
  "totals": {
    "subtotal": {
      "amount": 0,
      "currency": "INR"
    },
    "discount": {
      "amount": 0,
      "currency": "INR"
    },
    "total": {
      "amount": 0,
      "currency": "INR"
    },
    "currency": "INR",
    "item_count": 0,
    "unique_item_count": 0
  },
  "version": 7,
  "updated_at": "2026-05-24T00:00:00Z"
}
```

---

## Step 11: MongoDB Save Strategy Define Ki

MongoDB Cart Service ka durable source of truth hai. Remove ke baad updated `items[]`, `totals`, `updated_at`, `expires_at`, and `version` save honge.

### Optimistic Versioning

Task 3 me `version` field define hua tha. RemoveItem bhi AddItem ki tarah version guard use karega.

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
    "coupon_code": cart.CouponCode,
    "coupon_preview": cart.CouponPreview,
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
    A[Load active cart with version N] --> B[Remove item and cleanup zero quantity]
    B --> C[Recalculate totals]
    C --> D[Save where _id and version N]
    D --> E{Matched one document?}
    E -->|Yes| F[Save success, version becomes N+1]
    E -->|No| G[Version conflict]
    G --> H[Reload latest cart]
    H --> I{Retry left?}
    I -->|Yes| B
    I -->|No| J[Return conflict error]

    classDef ok fill:#dcfce7,stroke:#16a34a,color:#0f172a
    classDef warn fill:#fef3c7,stroke:#d97706,color:#0f172a
    classDef err fill:#fee2e2,stroke:#dc2626,color:#0f172a

    class F ok
    class G,H,I warn
    class J err
```

### Why optimistic locking?

- User two browser tabs se cart mutate kar sakta hai.
- One request item add kar rahi ho, another remove kar rahi ho.
- Version guard lost update avoid karta hai.
- Retry se user ko unnecessary conflict error kam dikhta hai.

---

## Step 12: Redis Cache Refresh Define Kiya

MongoDB save successful hone ke baad Redis cache refresh hoga.

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
Idempotent no-op delete => Redis refresh best-effort from current MongoDB cart
Redis refresh failure => log warning, still return updated cart
```

### Cache Code Example

```go
type CartCache interface {
	SetActiveCart(ctx context.Context, owner CartOwner, cart *Cart, ttl time.Duration) error
	SetCartSummary(ctx context.Context, owner CartOwner, summary CartSummary, ttl time.Duration) error
}

func BuildCartSummary(cart *Cart) CartSummary {
	return CartSummary{
		ItemCount: cart.Totals.ItemCount,
		Total:     cart.Totals.Total,
	}
}
```

### Beginner Explanation

MongoDB me original cart save hota hai. Redis me fast copy update hoti hai. Agar Redis fail bhi ho gaya, user ka remove operation successful maana jayega because durable cart MongoDB me correct state me hai.

---

## Step 13: RemoveItem Usecase Structure Banaya

Future backend implementation me `RemoveItemUsecase` orchestration layer hoga.

### Usecase Dependencies

```go
type RemoveItemUsecase struct {
	repo   CartRepository
	cache  CartCache
	clock  Clock
	logger Logger
}
```

### Command Model

```go
type RemoveItemCommand struct {
	UserID         string
	GuestSessionID string
	ItemID         string
}
```

### Full Usecase Pseudocode

```go
func (uc *RemoveItemUsecase) Execute(ctx context.Context, cmd RemoveItemCommand) (*Cart, error) {
	if err := ValidateRemoveItemInput(cmd.ItemID); err != nil {
		return nil, err
	}

	owner, err := ResolveCartOwner(cmd.UserID, cmd.GuestSessionID)
	if err != nil {
		return nil, err
	}

	return uc.withRetry(ctx, func() (*Cart, error) {
		now := uc.clock.Now()

		cart, err := uc.repo.FindActiveByOwner(ctx, owner)
		if err != nil {
			return nil, err
		}

		expectedVersion := cart.Version

		previousTotals := cart.Totals

		removed, err := cart.RemoveItem(cmd.ItemID, now)
		if err != nil {
			return nil, err
		}

		cleaned := cart.CleanupZeroQuantityItems(now)

		if err := cart.RecalculateTotals("INR"); err != nil {
			return nil, err
		}

		totalsChanged := !TotalsEqual(previousTotals, cart.Totals)
		mutated := removed || cleaned > 0 || totalsChanged

		if mutated {
			cart.ClearCouponPreview()
			cart.Touch(now)

			if err := uc.repo.SaveWithVersion(ctx, cart, expectedVersion); err != nil {
				return nil, err
			}
		} else {
			uc.logger.Info("remove item was idempotent", "item_id", cmd.ItemID)
		}

		summary := BuildCartSummary(cart)
		if err := uc.refreshCache(ctx, owner, cart, summary); err != nil {
			uc.logger.Warn("cart cache refresh failed", "error", err)
		}

		return cart, nil
	})
}
```

> 🟠 **Scope note:** Ye pseudocode implementation thinking clear karne ke liye hai. Is task me actual `RemoveItemUsecase` source file create nahi ki gayi.

---

## Step 14: Coupon Preview Boundary Clear Kiya

Task 5 ka scope coupon validation nahi hai, but cart item remove hone ke baad existing coupon preview stale ho sakta hai.

### Task 5 Decision

| Field | Behavior in Task 5 |
|---|---|
| `coupon_code` | Clear karna safe default hai |
| `coupon_preview` | Clear karna safe default hai |
| `discount` | `0` set karo |
| Coupon revalidation | Task 6 me implement hoga |

### Helper Example

```go
func (c *Cart) ClearCouponPreview() {
	c.CouponCode = ""
	c.CouponPreview = nil
}
```

### Why clear?

Cart content change hone ke baad old coupon discount valid nahi ho sakta. Example: coupon minimum cart amount `₹1000` hai, remove ke baad cart `₹700` ho gaya. Task 6 me proper revalidation aayegi; Task 5 me stale discount avoid karna safer hai.

---

## Step 15: Error Handling Matrix Banaya

| Scenario | gRPC Code | HTTP Status | User-facing Meaning |
|---|---|---:|---|
| Missing auth/session owner | `UNAUTHENTICATED` | `401` | Cart owner identify nahi hua |
| Missing `item_id` | `INVALID_ARGUMENT` | `400` | Item id required hai |
| Active cart not found | `NOT_FOUND` | `404` | Cart available nahi hai |
| Cart not active | `FAILED_PRECONDITION` | `409` | Cart editable state me nahi hai |
| Item already removed | `OK` | `200` | Updated/current cart return hoga |
| Mixed currency/corrupt totals | `FAILED_PRECONDITION` | `409` | Cart totals recalculate nahi ho paye |
| MongoDB unavailable | `UNAVAILABLE` | `503` | Cart temporarily unavailable |
| Redis refresh failed | `OK` with warning log | `200` | Mongo save successful, cache later rebuild hoga |
| Version conflict after retries | `ABORTED` | `409` | Concurrent cart update conflict |

### Error Response Example

```json
{
  "code": "CART_NOT_FOUND",
  "message": "Active cart was not found for this shopper.",
  "details": {
    "owner_type": "user"
  }
}
```

---

## Step 16: Architecture Diagram Banaya

```mermaid
flowchart LR
    UI[User App Frontend] --> GW[API Gateway]
    GW -->|DELETE /api/v1/cart/items/item_id| Cart[Cart Service]
    Cart -->|Read active cart| Mongo[(MongoDB cart_db.carts)]
    Cart -->|Save updated items and totals| Mongo
    Cart -->|Refresh active cart cache| Redis[(Redis)]

    Cart --> Rules[Cart Domain Rules]
    Rules --> Owner[Owner Validation]
    Rules --> Remove[Remove item_id]
    Rules --> Cleanup[Zero Quantity Cleanup]
    Rules --> Totals[Totals Recalculate]

    classDef client fill:#e0f2fe,stroke:#0284c7,color:#0f172a
    classDef svc fill:#f3e8ff,stroke:#7c3aed,color:#0f172a
    classDef db fill:#dcfce7,stroke:#16a34a,color:#0f172a
    classDef rule fill:#fef3c7,stroke:#d97706,color:#0f172a

    class UI,GW client
    class Cart svc
    class Mongo,Redis db
    class Rules,Owner,Remove,Cleanup,Totals rule
```

### Architecture Explanation

- API Gateway public REST route receive karega.
- Gateway auth/session metadata Cart Service ko pass karega.
- Cart Service active cart MongoDB se load karega.
- Cart Service embedded `items[]` array me se item remove karega.
- Cart Service zero quantity items cleanup karega.
- Cart Service totals recalculate karega.
- Cart Service MongoDB me durable update save karega.
- Cart Service Redis me hot cache refresh karega.
- Updated cart frontend ko return hoga.

---

## Step 17: Detailed Flow Diagram Banaya

```mermaid
flowchart TD
    A[Receive RemoveItem] --> B[Validate item_id]
    B --> C[Resolve user or guest owner]
    C --> D[Find active cart by owner]
    D --> E{Cart found and active?}
    E -->|No| X[Return error]
    E -->|Yes| F{Item exists in cart?}
    F -->|No| G[Keep cart unchanged]
    F -->|Yes| H[Remove item from items array]
    G --> I[Cleanup zero quantity items]
    H --> I
    I --> J[Recalculate totals]
    J --> K[Save MongoDB with version guard]
    K --> L{Save conflict?}
    L -->|Yes| M[Reload and retry]
    M --> F
    L -->|No| N[Refresh Redis cache]
    N --> O[Return updated cart]

    classDef start fill:#e0f2fe,stroke:#0284c7,color:#0f172a
    classDef action fill:#dcfce7,stroke:#16a34a,color:#0f172a
    classDef decision fill:#fef3c7,stroke:#d97706,color:#0f172a
    classDef error fill:#fee2e2,stroke:#dc2626,color:#0f172a

    class A start
    class B,C,D,G,H,I,J,K,M,N,O action
    class E,F,L decision
    class X error
```

### Final Remove Flow Summary

```text
Request
  -> Validate item_id
  -> Resolve owner
  -> Load active cart
  -> Remove matching item
  -> Cleanup quantity <= 0 items
  -> Recalculate totals
  -> Save MongoDB with version guard
  -> Refresh Redis cache
  -> Return updated cart
```

---

## Step 18: Folder Structure Define Kiya

### Actual TaskImplementation Structure

```text
TaskImplementation/
└── Cart Service/
    ├── task1.md
    ├── task2.md
    ├── task3.md
    ├── task4.md
    └── task5.md
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
        │   └── transport/
        │       └── grpc/
        │           └── server.go
        └── deploy/
            └── Dockerfile
```

### Task 5 Files In Future Implementation

| File | Responsibility |
|---|---|
| `internal/usecase/remove_item.go` | RemoveItem orchestration |
| `internal/domain/cart.go` | Active cart, status, owner, totals behavior |
| `internal/domain/cart_item.go` | Item remove and zero quantity cleanup helpers |
| `internal/domain/money.go` | Money operations without float |
| `internal/repository/mongo_cart_repository.go` | Active cart find/save with version |
| `internal/repository/redis_cart_cache.go` | Cache refresh after remove |
| `internal/transport/grpc/server.go` | `CartService.RemoveItem` gRPC handler |

> 🔴 **Important:** Ye future code structure hai. Current task output sirf `task5.md` documentation file hai.

---

## Step 19: External Libraries and Tools

Task 5 ke actual future implementation me ye libraries/tools use honge:

| Tool / Library | What it is | Why used | Install / Use |
|---|---|---|---|
| MongoDB `mongo:7` | Document database | Durable `cart_db.carts` source of truth | Docker Compose local stack se run |
| Redis `redis:7.2-alpine` | In-memory cache | Active cart and summary cache refresh ke liye | Docker Compose local stack se run |
| `go.mongodb.org/mongo-driver` | Official MongoDB Go driver | Active cart find/save ke liye | `go get go.mongodb.org/mongo-driver/mongo` |
| `github.com/redis/go-redis/v9` | Redis Go client | Cache SET/DEL/TTL operations ke liye | `go get github.com/redis/go-redis/v9` |
| `google.golang.org/grpc` | gRPC framework | `CartService.RemoveItem` internal service method ke liye | `go get google.golang.org/grpc` |
| Protobuf generated code | Typed service contracts | `IdPathRequest`, `Cart`, and `CartService.RemoveItem` stubs ke liye | Proto generation pipeline from Platform Foundation |

### Example Install Commands

```bash
cd backend/services/cart-service
go get go.mongodb.org/mongo-driver/mongo
go get github.com/redis/go-redis/v9
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
CART_ACTIVE_CACHE_TTL=15m
CART_SUMMARY_CACHE_TTL=5m
GUEST_CART_CACHE_TTL=30m
CART_DEFAULT_CURRENCY=INR
```

### How To Use In Future Code

```go
// Handler layer
func (s *Server) RemoveItem(ctx context.Context, req *cartv1.IdPathRequest) (*cartv1.Cart, error) {
	claims := authctx.FromContext(ctx)
	cart, err := s.removeItem.Execute(ctx, RemoveItemCommand{
		UserID:         claims.UserID,
		GuestSessionID: claims.GuestSessionID,
		ItemID:         req.Id,
	})
	if err != nil {
		return nil, mapCartError(err)
	}
	return mapCartToProto(cart), nil
}
```

---

## Step 20: Testing Strategy Banayi

Task 5 implementation user-facing mutation hai, so tests clear and focused hone chahiye.

### Unit Tests

| Test Case | Expected Result |
|---|---|
| Missing `item_id` | Validation error |
| Missing owner | Auth/owner error |
| Active cart not found | `CART_NOT_FOUND` |
| Cart not active | `CART_NOT_ACTIVE` |
| Remove existing item | Item array se delete |
| Remove last item | Empty active cart return with zero totals |
| Remove missing item | Current cart return, no hard error |
| Existing zero quantity item | Cleanup removes invalid item |
| Existing negative quantity item | Cleanup removes invalid item |
| Totals after remove | Subtotal, total, item counts correct |
| Mixed currency corrupted cart | Recalculate error |
| Redis refresh fails | Cart still returned after Mongo save |
| Version conflict | Reload and retry |

### Integration Tests

| Test | Setup |
|---|---|
| Mongo active cart load/save | MongoDB test container or local stack |
| Redis cache refresh | Redis test container or local stack |
| Remove item with concurrent add | Version retry expected |
| Remove item with stale frontend retry | Idempotent current cart response |

### Example Unit Test Name List

```text
TestRemoveItemDeletesExistingItem
TestRemoveItemReturnsEmptyCartAfterLastItem
TestRemoveItemCleansZeroQuantityItems
TestRemoveItemRecalculatesTotals
TestRemoveItemIsIdempotentWhenItemMissing
TestRemoveItemRetriesOnVersionConflict
TestRemoveItemReturnsCartWhenRedisRefreshFails
```

---

## Step 21: Observability Notes Add Kiye

RemoveItem flow production me frequent call hoga, isliye logs, metrics, and traces useful rahenge.

### Logs

| Event | Fields |
|---|---|
| Remove item started | `owner_type`, `item_id` |
| Active cart missing | `owner_type` |
| Item removed | `cart_id`, `item_id`, `remaining_unique_items` |
| Item already missing | `cart_id`, `item_id` |
| Zero quantity cleanup | `cart_id`, `cleaned_count` |
| Cart save conflict | `cart_id`, `expected_version` |
| Cache refresh failed | `cart_id`, `redis_error` |

### Metrics

| Metric | Meaning |
|---|---|
| `cart_remove_item_total` | Remove item attempts count |
| `cart_remove_item_success_total` | Successful remove item count |
| `cart_remove_item_idempotent_total` | Item already missing but request handled |
| `cart_remove_item_error_total` | Errors by reason |
| `cart_remove_item_duration_ms` | End-to-end latency |
| `cart_zero_quantity_cleanup_total` | Cleanup count |
| `cart_cache_refresh_error_total` | Redis refresh failures |

### Trace Spans

```text
CartService.RemoveItem
├── validate_request
├── cart_repository.find_active_by_owner
├── cart_domain.remove_item
├── cart_domain.cleanup_zero_quantity_items
├── cart_domain.recalculate_totals
├── cart_repository.save_with_version
└── cart_cache.refresh
```

---

## Step 22: Security and Ownership Rules

Cart data user-specific private data hai. RemoveItem me ownership bypass nahi hona chahiye.

### Security Rules

| Rule | Decision |
|---|---|
| Direct delete by global `item_id` | ❌ Not allowed |
| Delete inside owner active cart | ✅ Required |
| Logged-in user cart access | JWT/user claims se |
| Guest cart access | Guest session id se |
| Cross-user item remove | ❌ Blocked because lookup owner cart me scoped hai |
| Rate limit | Cart mutations around `60 per user per min` as per auth/security guide |

### Safe Delete Query Thinking

```text
Wrong thinking:
Delete item where item_id = X globally

Right thinking:
Load active cart for current owner
Then remove item_id = X only from that cart
```

---

## Step 23: Complete RemoveItem Flow Summary

```mermaid
flowchart TD
    A[Frontend clicks remove] --> B[API Gateway DELETE route]
    B --> C[CartService.RemoveItem]
    C --> D[Validate item_id]
    D --> E[Resolve owner from auth/session]
    E --> F[Load active cart]
    F --> G[Remove item_id if present]
    G --> H[Cleanup quantity <= 0]
    H --> I[Clear stale coupon preview]
    I --> J[Recalculate totals]
    J --> K[Save MongoDB]
    K --> L[Refresh Redis summary and active cart]
    L --> M[Return Cart response]

    classDef start fill:#e0f2fe,stroke:#0284c7,color:#0f172a
    classDef action fill:#dcfce7,stroke:#16a34a,color:#0f172a
    classDef cache fill:#f3e8ff,stroke:#7c3aed,color:#0f172a

    class A,B,C start
    class D,E,F,G,H,I,J,K,M action
    class L cache
```

### Final Rules

1. `item_id` required hai.
2. Owner required hai: `user_id` ya `guest_session_id`.
3. Remove direct global item delete nahi karega.
4. Remove sirf current owner ke active cart me hoga.
5. Missing item id current cart ke liye idempotent success hoga.
6. Last item remove hone par active empty cart keep hoga.
7. Quantity `0` ya negative items save se pehle cleanup honge.
8. Totals server side recalculate honge.
9. Discount Task 5 me `0` rahega because coupon preview Task 6 hai.
10. Cart mutation ke baad stale coupon preview clear karna safe default hai.
11. MongoDB source of truth rahega.
12. Redis mutation ke baad refresh hoga.
13. Redis failure cart mutation ko fail nahi karega.
14. Version guard concurrent updates se lost changes avoid karega.

---

## ✅ Task 5 Completion Checklist

| Requirement | Status |
|---|---|
| `TaskImplementation/` folder present | ✅ Done |
| `TaskImplementation/Cart Service/` folder present | ✅ Done |
| `task5.md` created | ✅ Done |
| Step-by-step implementation in Hinglish | ✅ Done |
| Remove item API/gRPC contract explained | ✅ Done |
| Owner validation explained | ✅ Done |
| Active cart lookup explained | ✅ Done |
| Item remove logic explained | ✅ Done |
| Quantity zero cleanup explained | ✅ Done |
| Empty cart behavior explained | ✅ Done |
| Totals recalculation explained | ✅ Done |
| MongoDB save strategy explained | ✅ Done |
| Redis cache refresh explained | ✅ Done |
| External tools/libraries listed | ✅ Done |
| Install/use commands added | ✅ Done |
| Folder structure added | ✅ Done |
| Code examples added | ✅ Done |
| Mermaid diagrams added | ✅ Done |
| Scope limited to Cart Service Task 5 | ✅ Done |

---

## 🚫 Explicitly Not Implemented in This Task

| Not Implemented | Reason |
|---|---|
| Actual Go backend files | User output requirement is TaskImplementation markdown artifact only |
| Add item flow | Already covered in Cart Service Task 4 |
| Product validation during remove | RemoveItem does not need Product Service call |
| Coupon preview calculation | Cart Service Task 6 |
| Guest cart merge | Cart Service Task 7 |
| Cart expiry job | Cart Service Task 8 |
| Inventory reservation | Order/checkout flow |
| Payment/order creation | Order and Payment services |
| Frontend cart page | User App Frontend task |

---

## 🧾 Final Outcome

Cart Service Task 5 ka final documented implementation standard:

```text
Frontend/API Gateway
  -> DELETE /api/v1/cart/items/{item_id}
  -> CartService.RemoveItem
  -> Validate item_id
  -> Resolve user/session owner
  -> Load active cart from MongoDB
  -> Remove matching embedded item
  -> Cleanup quantity <= 0 items
  -> Clear stale coupon preview
  -> Recalculate totals
  -> Save MongoDB durable cart with version guard
  -> Refresh Redis active cart and summary
  -> Return updated cart
```

Task 5 complete hai as a structured, beginner-friendly Remove Item flow guide. Future Cart Service implementation isi document ko usecase, repository, cache, and gRPC handler blueprint ki tarah follow karegi.
