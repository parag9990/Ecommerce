# 🛒 Cart Service - Task 6: Coupon Preview

![Task](https://img.shields.io/badge/Task-Cart%20Service%20%236-blue)
![Priority](https://img.shields.io/badge/Priority-P1-orange)
![Dependency](https://img.shields.io/badge/Dependency-CMS%20Service-yellow)
![Status](https://img.shields.io/badge/Status-Documented-success)
![Guide](https://img.shields.io/badge/Guide-Hinglish-orange)
![Scope](https://img.shields.io/badge/Scope-Coupon%20Preview-purple)

## 📌 Task Summary

| Field | Detail |
|---|---|
| Task Name | Coupon preview |
| Source | `docs/01-micro-tasks.md` → `Cart Service` → Task 6 |
| Priority | `P1` buyer experience feature |
| Dependency | CMS Service coupon validation |
| Main Goal | CMS coupon rules validate karke cart discount preview dikhana |
| Output Type | Documentation-only implementation guide |
| Not Included | Final coupon apply, coupon redemption write, payment/order confirmation, CMS coupon engine implementation, actual backend source file creation |

> **Simple Hinglish goal:** Is task ka purpose Cart Service ke `ApplyCouponPreview` flow ko design karna hai. Buyer coupon code enter karega, Cart Service active cart load karega, cart context CMS Service ko bhejega, CMS coupon rules validate karega, Cart Service discount preview ko cart totals me reflect karega, MongoDB + Redis update karega, aur frontend ko preview response return karega. Final coupon apply checkout ke time Order Service fresh validate karega.

---

## ✅ Final Output Created

```text
TaskImplementation/
└── Cart Service/
    ├── task1.md
    ├── task2.md
    ├── task3.md
    ├── task4.md
    ├── task5.md
    └── task6.md
```

### Why this structure?

- `TaskImplementation/` project ke task-wise implementation guides ka central folder hai.
- `Cart Service/` Cart Service ke saare task guides ko group karta hai.
- `task6.md` sirf **Cart Service - Task 6** ka Coupon Preview guide hai.
- Is task me sirf documentation artifact create kiya gaya hai; backend source code create nahi kiya gaya.

---

## 🧭 Implementation Approach

Is guide ko banane se pehle existing project documentation study ki gayi:

| Document | Kya use kiya gaya |
|---|---|
| `docs/01-micro-tasks.md` | Task 6 ka exact scope: CMS coupon rules validate karke discount preview; final apply Order Service karega |
| `TaskImplementation/Cart Service/task1.md` | Coupon boundary: Cart preview, CMS validation, Order final apply |
| `TaskImplementation/Cart Service/task2.md` | MongoDB source of truth and Redis hot cache strategy |
| `TaskImplementation/Cart Service/task3.md` | `coupon_code`, `coupon_preview`, and `totals.discount` schema fields |
| `TaskImplementation/Cart Service/task4.md` | Add item ke baad coupon preview stale ho sakta hai |
| `TaskImplementation/Cart Service/task5.md` | Remove item ke baad stale coupon preview clear/revalidate boundary |
| `docs/04-microservice-design.md` | Cart `ApplyCouponPreview` and CMS `ValidateCoupon` service responsibilities |
| `docs/05-database-design.md` | CMS coupon tables and short TTL validation cache note |
| `docs/09-cms-superadmin.md` | Supported coupon rules: fixed, percentage, min cart, cap, scope, usage limits, time window |
| `api/master-api.json` | `POST /api/v1/cart/coupons/preview`, `CouponPreviewRequest`, `CouponPreviewResponse` |
| `backend/services/cart-service/internal/domain/cart.go` | Existing future domain fields: `CouponCode`, `CouponPreview`, `CartTotals.Discount` |

---

## 🪜 Step-by-Step Implementation

## Step 1: Task Boundary Clear Kiya

Cart Service Task 6 ka kaam **coupon preview** define karna hai, final coupon redemption nahi.

### Included

- `POST /api/v1/cart/coupons/preview` and `CartService.ApplyCouponPreview` contract explain karna
- Buyer/guest owner resolve karna
- Active cart load karna
- Coupon code sanitize and validate karna
- Cart context CMS Service ko bhejna
- CMS `ValidateCoupon` result ko Cart Service me safely apply karna
- Valid coupon ke liye discount preview persist karna
- Invalid coupon ke liye old discount clear karke reason return karna
- Cart totals recalculate karna
- MongoDB me updated cart save karna
- Redis cart and summary cache refresh karna
- Error handling, tests, observability, and future file structure document karna

### Not Included

- CMS coupon engine create/update/list implementation
- Coupon usage/redemption count increment karna
- Payment success ke baad coupon redemption record karna
- Final checkout coupon apply guarantee dena
- Order Service `CreateOrderFromCart` implementation
- Product inventory reservation
- Actual backend source files create karna

> 🟢 **Reason:** `docs/01-micro-tasks.md` clearly bolta hai: "Coupon preview ... Final apply Order Service karega." Isliye Cart Service buyer ko estimate dikhayega, but checkout ke time Order Service fresh validation karega.

---

## Step 2: Public API Contract Samjha

Coupon preview API Gateway ke through buyer-facing route hoga.

| Layer | Contract |
|---|---|
| REST | `POST /api/v1/cart/coupons/preview` |
| Internal gRPC | `CartService.ApplyCouponPreview` |
| Request schema | `CouponPreviewRequest` |
| Response schema | `CouponPreviewResponse` |
| Auth context | Buyer/user context, ya guest session context |

### REST Request

```http
POST /api/v1/cart/coupons/preview
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "coupon_code": "SAVE10"
}
```

Guest cart ke case me Gateway guest session cookie/header se `guest_session_id` Cart Service ko pass karega.

### Request Fields

| Field | Required? | Meaning |
|---|---:|---|
| `coupon_code` | ✅ Yes | Buyer ne jo code enter kiya |
| `cart_id` | ❌ Optional | Agar frontend send kare to Cart Service ownership validate karega |
| `order_id` | ❌ Optional | Cart preview flow me normally use nahi hoga; final apply Order Service side par hoga |

### Response Body

```json
{
  "valid": true,
  "coupon_id": "coupon_123",
  "discount": {
    "amount": 2500,
    "currency": "INR"
  },
  "reason": ""
}
```

### Invalid Coupon Response

```json
{
  "valid": false,
  "coupon_id": "",
  "discount": {
    "amount": 0,
    "currency": "INR"
  },
  "reason": "MIN_CART_AMOUNT_NOT_MET"
}
```

> 🔵 **Mapping note:** API response only preview result return karta hai. Cart Service internally cart document bhi update karega, taaki next `GET /api/v1/cart` me `discount` and `total` updated dikhe.

---

## Step 3: High-Level Coupon Preview Flow Design Kiya

Coupon preview flow me Cart Service three systems ke saath kaam karega:

1. **MongoDB**: Active cart durable source of truth
2. **CMS Service**: Coupon rules validate karne ka owner
3. **Redis**: Updated cart and summary cache

```mermaid
sequenceDiagram
    participant Client as Frontend / API Gateway
    participant Cart as Cart Service
    participant Mongo as MongoDB cart_db.carts
    participant CMS as CMS Service
    participant Redis as Redis Hot Cache

    Client->>Cart: ApplyCouponPreview(coupon_code)
    Cart->>Cart: Resolve owner + sanitize coupon code
    Cart->>Mongo: Find active cart by owner
    Mongo-->>Cart: Active cart with items and totals
    Cart->>Cart: Validate cart is previewable
    Cart->>CMS: ValidateCoupon(cart context + coupon_code)
    CMS-->>Cart: valid / coupon_id / discount / reason
    Cart->>Cart: Apply preview discount safely
    Cart->>Cart: Recalculate totals
    Cart->>Mongo: Save updated cart with version guard
    Mongo-->>Cart: Updated cart saved
    Cart->>Redis: Refresh active cart + summary cache
    Redis-->>Cart: OK or warning
    Cart-->>Client: CouponPreviewResponse
```

### Beginner Explanation

Frontend coupon code bhejta hai, but frontend se amount trust nahi hota. Cart Service active cart khud load karta hai. CMS Service coupon rules check karta hai. Cart Service CMS ke result ko validate karke cart totals update karta hai. Isse buyer ko fast preview milta hai, but final checkout still fresh validation se safe rehta hai.

---

## Step 4: Coupon Ownership Boundary Define Kiya

Coupon logic do services me split hai:

| Service | Responsibility |
|---|---|
| Cart Service | Active cart load, cart context banana, preview result cart totals me apply karna |
| CMS Service | Coupon code lookup, rule validation, discount calculation, usage/time/scope checks |
| Order Service | Checkout ke time final coupon validation and final apply |
| Payment/Order success flow | Coupon redemption record karna |

### Golden Rule

```text
Cart Service discount preview de sakta hai.
Order Service final payable amount decide karega.
CMS Service coupon rule truth own karega.
```

> 🟡 **Important:** Preview valid hone ka matlab final checkout guarantee nahi hai. Coupon expire ho sakta hai, usage limit khatam ho sakti hai, ya cart content change ho sakta hai.

---

## Step 5: Existing Cart Schema Use Kiya

Task 3 ke schema me coupon preview ke liye fields already reserved hain:

```json
{
  "coupon_code": "SAVE10",
  "coupon_preview": {
    "coupon_id": "coupon_123",
    "code": "SAVE10",
    "valid": true,
    "discount": {
      "amount": 2500,
      "currency": "INR"
    },
    "reason": ""
  },
  "totals": {
    "subtotal": {
      "amount": 25000,
      "currency": "INR"
    },
    "discount": {
      "amount": 2500,
      "currency": "INR"
    },
    "total": {
      "amount": 22500,
      "currency": "INR"
    },
    "currency": "INR",
    "item_count": 3,
    "unique_item_count": 2
  }
}
```

### Field Meaning

| Field | Meaning |
|---|---|
| `coupon_code` | Latest previewed coupon code |
| `coupon_preview.coupon_id` | CMS coupon ID |
| `coupon_preview.valid` | CMS validation result |
| `coupon_preview.discount` | Preview discount in smallest currency unit |
| `coupon_preview.reason` | Invalid reason or warning |
| `totals.discount` | Cart total discount used for preview |
| `totals.total` | `subtotal - discount`, never below zero |

### Existing Domain Shape

```go
type Cart struct {
	ID            string
	CouponCode    *string
	CouponPreview *CouponView
	Totals        CartTotals
	Version       int64
}

type CouponView struct {
	CouponID string
	Code     string
	Valid    bool
	Discount Money
	Reason   string
}

type CartTotals struct {
	Subtotal        Money
	Discount        Money
	Total           Money
	Currency        string
	ItemCount       int
	UniqueItemCount int
}
```

> 🟢 **Good news:** Task 6 ko new cart document shape invent nahi karna. Existing schema fields use honge.

---

## Step 6: Request Validation Rules Banaye

Coupon preview start hone se pehle request validate hogi.

| Rule | Behavior |
|---|---|
| Missing owner | Reject with auth/owner error |
| Empty `coupon_code` | Reject with validation error |
| Too long code | Reject; suggested max `64` chars |
| Leading/trailing spaces | Trim |
| Lowercase code | Normalize to uppercase |
| Cart not found | Return cart not found / empty cart error |
| Cart not active | Reject mutation |
| Empty cart | Reject or return invalid preview with reason `EMPTY_CART` |
| Optional `cart_id` mismatch | Reject with ownership error |

### Coupon Code Sanitizer

```go
const MaxCouponCodeLength = 64

func NormalizeCouponCode(code string) (string, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	if code == "" {
		return "", ErrCouponCodeRequired
	}
	if len(code) > MaxCouponCodeLength {
		return "", ErrCouponCodeTooLong
	}
	return code, nil
}
```

### Why uppercase?

Buyer `save10`, `SAVE10`, ya ` Save10 ` likh sakta hai. Coupon matching predictable rakhne ke liye Cart Service normalized code CMS ko bhejega.

---

## Step 7: Active Cart Load Strategy Banayi

Coupon preview always active cart par apply hoga.

```mermaid
flowchart TD
    A[ApplyCouponPreview request] --> B{user_id available?}
    B -- Yes --> C[Find active cart by user_id]
    B -- No --> D{guest_session_id available?}
    D -- Yes --> E[Find active cart by guest_session_id]
    D -- No --> F[Reject: owner missing]
    C --> G{active cart found?}
    E --> G
    G -- No --> H[Reject or invalid EMPTY_CART]
    G -- Yes --> I{optional cart_id matches?}
    I -- No --> J[Reject ownership mismatch]
    I -- Yes --> K[Proceed to CMS validation]
```

### Repository Interface

```go
type CartRepository interface {
	FindActiveByOwner(ctx context.Context, owner CartOwner) (*Cart, error)
	SaveWithVersion(ctx context.Context, cart *Cart, expectedVersion int64) error
}
```

### Owner Safety

Cart Service ko request body ke `cart_id` par blindly trust nahi karna. Owner context JWT/session se aayega. Agar `cart_id` optional hai to sirf extra verification ke liye use hoga.

---

## Step 8: Cart Preview Eligibility Check Kiya

Har cart coupon preview ke liye eligible nahi hota.

| Cart State | Preview Allowed? | Reason |
|---|---:|---|
| `active` with items | ✅ Yes | Normal cart |
| `active` but empty | ❌ No | Discount ka base amount nahi hai |
| `merged` | ❌ No | Old guest cart editable nahi |
| `checked_out` | ❌ No | Checkout ho chuka hai |
| `expired` | ❌ No | Stale cart |
| `abandoned` | ❌ No | Not active |

### Example Check

```go
func EnsureCouponPreviewAllowed(cart *Cart) error {
	if cart.Status != CartStatusActive {
		return ErrCartNotActive
	}
	if len(cart.Items) == 0 || cart.Totals.Subtotal.Amount == 0 {
		return ErrEmptyCart
	}
	return nil
}
```

> 🟣 **Why:** Empty cart par coupon preview karna confusing UX hai. Buyer ko pehle item add karna chahiye.

---

## Step 9: CMS Validation Context Build Kiya

CMS Service ko coupon validate karne ke liye cart ka trusted snapshot chahiye.

### Validation Context

```go
type CouponValidationRequest struct {
	Code     string
	UserID   string
	CartID   string
	Currency string
	Subtotal Money
	Items    []CouponCartItem
	Now      time.Time
}

type CouponCartItem struct {
	ProductID    string
	VariantID    string
	SellerID     string
	UnitPrice    Money
	Quantity     int
	LineSubtotal Money
}
```

### Cart Se CMS Ko Kya Bhejna Hai?

| Data | Why needed |
|---|---|
| `coupon_code` | Coupon lookup |
| `user_id` | Per-user usage limit, new user rule |
| `cart_id` | Audit/debug and optional validation cache key |
| `subtotal` | Minimum cart amount and percentage discount |
| `currency` | Coupon currency compatibility |
| `items.product_id` | Product-scoped coupons |
| `items.seller_id` | Seller-scoped coupons |
| `items.quantity` | Item-level calculations |
| `items.line_subtotal` | Eligible subtotal calculation |

### Category Scope Note

Docs me category scope bhi mentioned hai. Current cart item snapshot me `category_id` field nahi hai. Isliye MVP ke liye recommended boundary:

- Cart Service product IDs and seller IDs CMS ko bheje.
- CMS Service apne read model ya Product Service lookup se category scope validate kare.
- Future schema extension me cart item snapshot me `category_ids` add kiye ja sakte hain, but Task 6 me schema change force nahi karna.

> 🔵 **Contract note:** `api/master-api.json` me CMS `ValidateCoupon` high-level `CouponPreviewRequest` reuse karta hai. Real internal proto ko richer cart context include karna chahiye, ya CMS ko cart/product read model access chahiye. Public frontend request simple hi rahega.

---

## Step 10: CMS Service Client Interface Define Kiya

Cart Service CMS coupon logic khud implement nahi karega. Ye sirf CMS validator ko call karega.

```go
type CouponValidator interface {
	ValidateCoupon(ctx context.Context, req CouponValidationRequest) (*CouponValidationResult, error)
}

type CouponValidationResult struct {
	Valid    bool
	CouponID string
	Code     string
	Discount Money
	Reason   string
}
```

### CMS Result Rules

| CMS Result | Cart Behavior |
|---|---|
| `valid=true`, discount positive/zero | Apply preview |
| `valid=false` | Clear discount, store reason |
| CMS timeout/unavailable | Do not change cart; return service unavailable |
| Currency mismatch | Reject CMS result as invalid |
| Discount greater than subtotal | Clamp or reject; recommended reject and log |
| Negative discount | Reject CMS result |

### Why Cart Should Validate CMS Result?

Service-to-service result trusted hota hai, but defensive validation zaruri hai. Agar CMS bug ya stale version wrong currency/negative amount return kare, Cart totals corrupt nahi hone chahiye.

---

## Step 11: Discount Safety Rules Banaye

Cart Service ko final discount amount CMS se milega, but apply karne se pehle safety checks honge.

```go
func ValidatePreviewDiscount(subtotal Money, discount Money) error {
	if err := discount.Validate(); err != nil {
		return err
	}
	if !subtotal.SameCurrency(discount) {
		return ErrCouponCurrencyMismatch
	}
	if discount.Amount > subtotal.Amount {
		return ErrCouponDiscountTooLarge
	}
	return nil
}
```

### Money Handling

| Rule | Reason |
|---|---|
| Amount smallest currency unit me store hoga | Floating point rounding avoid hota hai |
| INR `2500` means `₹25.00` if paise model use ho | Consistent money math |
| Discount subtotal se zyada nahi hoga | Negative total avoid |
| Total minimum `0` hoga | Buyer payable amount negative nahi ho sakta |

> 🟢 **Important:** Cart Service percentage formula khud calculate nahi karega. CMS coupon engine percentage/fixed/min/cart/scope logic calculate karega and final preview discount amount return karega.

---

## Step 12: Valid Coupon Apply Flow Banaya

Jab CMS valid coupon return kare, Cart Service cart preview update karega.

```go
func (c *Cart) ApplyValidCouponPreview(result CouponValidationResult, now time.Time) error {
	if err := ValidatePreviewDiscount(c.Totals.Subtotal, result.Discount); err != nil {
		return err
	}

	code := strings.ToUpper(strings.TrimSpace(result.Code))
	c.CouponCode = &code
	c.CouponPreview = &CouponView{
		CouponID: result.CouponID,
		Code:     code,
		Valid:    true,
		Discount: result.Discount,
		Reason:   "",
	}

	totals, err := CalculateTotals(c.Items, c.Totals.Currency, result.Discount)
	if err != nil {
		return err
	}
	c.Totals = totals
	c.UpdatedAt = now
	c.Version++
	return nil
}
```

### Result

```text
subtotal = old subtotal
discount = CMS returned preview discount
total    = subtotal - discount
```

---

## Step 13: Invalid Coupon Flow Banaya

Invalid coupon par old valid discount continue nahi rehna chahiye, warna buyer confused hoga.

```go
func (c *Cart) ApplyInvalidCouponPreview(code string, reason string, now time.Time) error {
	code = strings.ToUpper(strings.TrimSpace(code))
	zero := NewMoney(0, c.Totals.Currency)

	c.CouponCode = &code
	c.CouponPreview = &CouponView{
		CouponID: "",
		Code:     code,
		Valid:    false,
		Discount: zero,
		Reason:   reason,
	}

	totals, err := CalculateTotals(c.Items, c.Totals.Currency, zero)
	if err != nil {
		return err
	}
	c.Totals = totals
	c.UpdatedAt = now
	c.Version++
	return nil
}
```

### Invalid Reasons Examples

| Reason | Meaning |
|---|---|
| `COUPON_NOT_FOUND` | Code exist nahi karta |
| `COUPON_INACTIVE` | Coupon active nahi |
| `COUPON_EXPIRED` | Time window khatam |
| `COUPON_NOT_STARTED` | Start time future me hai |
| `MIN_CART_AMOUNT_NOT_MET` | Cart subtotal minimum se kam hai |
| `SELLER_SCOPE_NOT_MATCHED` | Cart me eligible seller item nahi |
| `PRODUCT_SCOPE_NOT_MATCHED` | Eligible product cart me nahi |
| `USER_LIMIT_REACHED` | User ne max usage kar liya |
| `GLOBAL_LIMIT_REACHED` | Global usage limit khatam |

> 🟡 **UX note:** Invalid preview cart me zero discount set karega and reason store karega. Frontend clear message dikha sakta hai.

---

## Step 14: Totals Recalculation Rule Define Kiya

Coupon preview ke baad totals same domain function se recalculate honge.

```go
func RecalculateCartTotalsWithCoupon(cart *Cart, discount Money) error {
	totals, err := CalculateTotals(cart.Items, cart.Totals.Currency, discount)
	if err != nil {
		return err
	}
	cart.Totals = totals
	return nil
}
```

### Example

| Field | Before Coupon | After `SAVE10` |
|---|---:|---:|
| Subtotal | `25000 INR` | `25000 INR` |
| Discount | `0 INR` | `2500 INR` |
| Total | `25000 INR` | `22500 INR` |
| Item count | `3` | `3` |

### Why item count same rahega?

Coupon preview price/totals change karta hai. Cart items add/remove/update nahi karta.

---

## Step 15: MongoDB Save Strategy Define Kiya

Coupon preview cart mutation hai, so MongoDB durable update required hai.

### Save Rules

| Rule | Why |
|---|---|
| Save with version guard | Concurrent add/remove/update se lost update avoid hota hai |
| Update `coupon_code` | Latest user-entered code persist hota hai |
| Update `coupon_preview` | UI reason/discount show kar sakta hai |
| Update `totals.discount` and `totals.total` | Cart summary accurate dikhta hai |
| Increment `version` | Optimistic locking work karta hai |
| Update `updated_at` | Cart activity track hoti hai |

### Optimistic Save Example

```go
func (uc *ApplyCouponPreviewUsecase) saveWithRetry(
	ctx context.Context,
	owner CartOwner,
	apply func(*Cart) error,
) (*Cart, error) {
	for attempt := 0; attempt < 3; attempt++ {
		cart, err := uc.repo.FindActiveByOwner(ctx, owner)
		if err != nil {
			return nil, err
		}

		expectedVersion := cart.Version
		if err := apply(cart); err != nil {
			return nil, err
		}

		err = uc.repo.SaveWithVersion(ctx, cart, expectedVersion)
		if errors.Is(err, ErrVersionConflict) {
			continue
		}
		return cart, err
	}
	return nil, ErrCartUpdateConflict
}
```

> 🔵 **Why retry:** Buyer same time quantity update aur coupon apply kar sakta hai. Version conflict par latest cart reload karke discount fresh context par apply karna safer hai.

---

## Step 16: Redis Cache Refresh Strategy Define Kiya

MongoDB save successful hone ke baad Redis refresh hoga.

| Cache Key | Value | TTL |
|---|---|---|
| `cart:active:user:{user_id}` | Full active cart | `15m` |
| `cart:active:guest:{guest_session_id}` | Full guest cart | `30m` |
| `cart:summary:user:{user_id}` | Badge count + total | `5m` |
| `cart:summary:guest:{guest_session_id}` | Badge count + total | `5m` |

### Cache Interface

```go
type CartCache interface {
	SetActiveCart(ctx context.Context, owner CartOwner, cart *Cart, ttl time.Duration) error
	SetCartSummary(ctx context.Context, owner CartOwner, summary CartSummary, ttl time.Duration) error
}
```

### Cache Failure Rule

```text
MongoDB save success + Redis refresh fail = request still successful.
Log warning and allow cache to heal on next read/write.
```

Cart mutation ka source of truth MongoDB hai. Redis performance layer hai, correctness layer nahi.

---

## Step 17: Complete Usecase Flow Likha

```go
type ApplyCouponPreviewCommand struct {
	UserID         string
	GuestSessionID string
	CartID         string
	CouponCode     string
}

type ApplyCouponPreviewUsecase struct {
	repo       CartRepository
	cache      CartCache
	cms        CouponValidator
	clock      Clock
	logger     Logger
}

func (uc *ApplyCouponPreviewUsecase) Execute(
	ctx context.Context,
	cmd ApplyCouponPreviewCommand,
) (*CouponPreviewResult, error) {
	owner, err := ResolveCartOwner(cmd.UserID, cmd.GuestSessionID)
	if err != nil {
		return nil, err
	}

	code, err := NormalizeCouponCode(cmd.CouponCode)
	if err != nil {
		return nil, err
	}

	var preview *CouponPreviewResult
	cart, err := uc.saveWithRetry(ctx, owner, func(cart *Cart) error {
		if cmd.CartID != "" && cmd.CartID != cart.ID {
			return ErrCartOwnershipMismatch
		}
		if err := EnsureCouponPreviewAllowed(cart); err != nil {
			return err
		}

		cmsReq := BuildCouponValidationRequest(cart, owner, code, uc.clock.Now())
		cmsResult, err := uc.cms.ValidateCoupon(ctx, cmsReq)
		if err != nil {
			return err
		}

		if cmsResult.Valid {
			if err := cart.ApplyValidCouponPreview(*cmsResult, uc.clock.Now()); err != nil {
				return err
			}
		} else {
			if err := cart.ApplyInvalidCouponPreview(code, cmsResult.Reason, uc.clock.Now()); err != nil {
				return err
			}
		}

		preview = MapCouponPreviewResult(cart.CouponPreview)
		return nil
	})
	if err != nil {
		return nil, err
	}

	if err := uc.refreshCache(ctx, owner, cart); err != nil {
		uc.logger.Warn("cart coupon preview cache refresh failed", "error", err)
	}

	return preview, nil
}
```

### Flow Explanation

- Owner resolve karta hai taaki buyer apna hi cart mutate kare.
- Coupon code normalize hota hai taaki CMS lookup consistent rahe.
- Active cart latest version ke saath load hota hai.
- CMS validation fresh cart context par hoti hai.
- Valid/invalid preview cart totals me reflect hota hai.
- MongoDB save version guard ke saath hota hai.
- Redis refresh best-effort hota hai.

---

## Step 18: gRPC Handler Mapping Define Kiya

```go
func (s *Server) ApplyCouponPreview(
	ctx context.Context,
	req *cartv1.CouponPreviewRequest,
) (*cartv1.CouponPreviewResponse, error) {
	claims := authctx.FromContext(ctx)

	result, err := s.applyCouponPreview.Execute(ctx, ApplyCouponPreviewCommand{
		UserID:         claims.UserID,
		GuestSessionID: claims.GuestSessionID,
		CartID:         req.CartId,
		CouponCode:     req.CouponCode,
	})
	if err != nil {
		return nil, mapCartError(err)
	}

	return &cartv1.CouponPreviewResponse{
		Valid:    result.Valid,
		CouponId: result.CouponID,
		Discount: mapMoneyToProto(result.Discount),
		Reason:   result.Reason,
	}, nil
}
```

### Handler Responsibility

Handler sirf transport mapping karega:

- gRPC request se command banana
- auth/session context attach karna
- usecase call karna
- domain error ko gRPC status me map karna
- response DTO/proto return karna

Business logic handler me nahi aayega.

---

## Step 19: Error Handling Map Banaya

| Scenario | gRPC Code | HTTP Code | Reason |
|---|---|---:|---|
| Missing auth/session owner | `UNAUTHENTICATED` | `401` | Cart owner identify nahi hua |
| Empty coupon code | `INVALID_ARGUMENT` | `400` | Input invalid |
| Coupon code too long | `INVALID_ARGUMENT` | `400` | Input unsafe/invalid |
| Active cart not found | `NOT_FOUND` | `404` | Cart missing |
| Empty cart | `FAILED_PRECONDITION` | `409` | Coupon empty cart par valid nahi |
| Cart not active | `FAILED_PRECONDITION` | `409` | Mutations allowed nahi |
| Cart ID ownership mismatch | `PERMISSION_DENIED` | `403` | User dusre cart ko mutate nahi kar sakta |
| CMS says coupon invalid | `OK` | `200` | Business invalid response, transport error nahi |
| CMS unavailable | `UNAVAILABLE` | `503` | Preview temporarily unavailable |
| MongoDB unavailable | `UNAVAILABLE` | `503` | Cart temporarily unavailable |
| Version conflict after retries | `ABORTED` | `409` | Concurrent update conflict |

### Invalid Coupon Is Not Transport Error

```json
{
  "valid": false,
  "discount": {
    "amount": 0,
    "currency": "INR"
  },
  "reason": "COUPON_EXPIRED"
}
```

Buyer ne wrong/expired coupon enter kiya to API fail nahi hoti. Response successful hota hai with `valid=false`.

---

## Step 20: Architecture Diagram

```mermaid
flowchart LR
    FE[Frontend Cart Page] -->|POST /api/v1/cart/coupons/preview| GW[API Gateway]
    GW -->|gRPC ApplyCouponPreview| Cart[Cart Service]

    Cart -->|Find active cart| Mongo[(MongoDB cart_db.carts)]
    Cart -->|ValidateCoupon| CMS[CMS Service]
    CMS -->|Read rules| MySQL[(CMS MySQL coupons)]
    Cart -->|Refresh cache| Redis[(Redis)]

    Cart --> Rules[Cart Domain Rules]
    Rules --> Totals[Recalculate totals]
    Totals --> Mongo

    classDef svc fill:#e8f1ff,stroke:#2f66d0,color:#10213d
    classDef db fill:#eef9f1,stroke:#2c7a3f,color:#12351d
    classDef rule fill:#fff5df,stroke:#b7791f,color:#3d2c10

    class GW,Cart,CMS svc
    class Mongo,MySQL,Redis db
    class Rules,Totals rule
```

### Architecture Explanation

- Frontend only coupon code bhejta hai.
- Gateway auth/session metadata Cart Service ko pass karta hai.
- Cart Service cart data MongoDB se read karta hai.
- CMS Service coupon rules MySQL se validate karta hai.
- Cart Service CMS result ko cart totals me apply karta hai.
- Redis updated cart/summary cache rakhta hai.

---

## Step 21: Coupon Preview State Diagram

```mermaid
stateDiagram-v2
    [*] --> NoCoupon
    NoCoupon --> ValidPreview: valid coupon applied
    NoCoupon --> InvalidPreview: invalid coupon attempted

    ValidPreview --> NoCoupon: cart item changed and preview cleared
    ValidPreview --> ValidPreview: another valid coupon applied
    ValidPreview --> InvalidPreview: invalid coupon attempted

    InvalidPreview --> ValidPreview: valid coupon applied
    InvalidPreview --> NoCoupon: coupon cleared / cart changed

    ValidPreview --> CheckoutRevalidation: buyer checks out
    CheckoutRevalidation --> FinalApplied: Order Service validates again
    CheckoutRevalidation --> FinalRejected: expired/limit/cart changed
```

### State Meaning

| State | Meaning |
|---|---|
| `NoCoupon` | Cart discount zero hai |
| `ValidPreview` | Cart me preview discount visible hai |
| `InvalidPreview` | Last attempted code invalid tha |
| `CheckoutRevalidation` | Order Service fresh validation kar raha hai |
| `FinalApplied` | Coupon final order me apply ho gaya |
| `FinalRejected` | Preview checkout me invalid ho gaya |

---

## Step 22: Folder Structure

### Documentation Output

```text
TaskImplementation/
└── Cart Service/
    ├── task1.md
    ├── task2.md
    ├── task3.md
    ├── task4.md
    ├── task5.md
    └── task6.md
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
        │   │   ├── coupon.go
        │   │   └── money.go
        │   ├── usecase/
        │   │   ├── add_item.go
        │   │   ├── remove_item.go
        │   │   └── apply_coupon_preview.go
        │   ├── client/
        │   │   └── cms_coupon_client.go
        │   ├── repository/
        │   │   ├── mongo_cart_repository.go
        │   │   └── redis_cart_cache.go
        │   └── transport/
        │       └── grpc/
        │           └── server.go
        └── deploy/
            └── Dockerfile
```

### Task 6 Files In Future Implementation

| File | Responsibility |
|---|---|
| `internal/usecase/apply_coupon_preview.go` | Coupon preview orchestration |
| `internal/domain/coupon.go` | Coupon preview domain helpers |
| `internal/domain/cart.go` | Apply/clear coupon preview on cart aggregate |
| `internal/domain/money.go` | Money validation and currency checks |
| `internal/client/cms_coupon_client.go` | CMS `ValidateCoupon` gRPC client |
| `internal/repository/mongo_cart_repository.go` | Active cart find/save with version |
| `internal/repository/redis_cart_cache.go` | Cache refresh after preview |
| `internal/transport/grpc/server.go` | `CartService.ApplyCouponPreview` handler |

> 🔴 **Important:** Ye future code structure hai. Current task output sirf `task6.md` documentation file hai.

---

## Step 23: External Libraries and Tools

Task 6 documentation create karne ke liye koi external library install nahi ki gayi. Future backend implementation me existing platform stack ke tools use honge.

| Tool / Library | What it is | Why used | Install / Use |
|---|---|---|---|
| MongoDB `mongo:7` | Document database | `cart_db.carts` me coupon preview and totals persist karne ke liye | Docker Compose local stack se run |
| Redis `redis:7.2-alpine` | In-memory cache | Active cart and cart summary refresh ke liye | Docker Compose local stack se run |
| MySQL | Relational database | CMS coupon rules, coupon usage limits, seller ownership store karne ke liye | CMS Service stack me run |
| `go.mongodb.org/mongo-driver` | Official MongoDB Go driver | Cart read/save with optimistic versioning | `go get go.mongodb.org/mongo-driver/mongo` |
| `github.com/redis/go-redis/v9` | Redis Go client | Cart cache SET/TTL operations ke liye | `go get github.com/redis/go-redis/v9` |
| `google.golang.org/grpc` | gRPC framework | Cart Service ↔ CMS Service internal call ke liye | `go get google.golang.org/grpc` |
| Protobuf generated code | Typed service contracts | `CartService.ApplyCouponPreview` and `CMSService.ValidateCoupon` stubs ke liye | Platform proto generation pipeline se generate |

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
CMS_GRPC_ADDR=cms-service:9090
CART_ACTIVE_CACHE_TTL=15m
CART_SUMMARY_CACHE_TTL=5m
GUEST_CART_CACHE_TTL=30m
COUPON_PREVIEW_TIMEOUT=500ms
CART_DEFAULT_CURRENCY=INR
```

### How To Use In Future Code

```go
cmsConn, err := grpc.DialContext(ctx, cfg.CMSGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
if err != nil {
	return err
}

couponValidator := client.NewCMSCouponClient(cmsv1.NewCMSServiceClient(cmsConn))
usecase := usecase.NewApplyCouponPreviewUsecase(repo, cache, couponValidator, clock, logger)
```

> 🟡 **Note:** Real production code me TLS/mTLS, deadlines, retries, and service discovery Platform Foundation ke standards ke according configure honge.

---

## Step 24: Timeout and Resilience Rules

Coupon preview user-facing hai, isliye slow CMS call cart page ko block nahi karni chahiye.

| Rule | Recommended Value |
|---|---:|
| CMS validation timeout | `500ms` to `1s` |
| Retry count | `0` or `1` max for transient unavailable |
| Cache refresh timeout | `100ms` to `250ms` |
| Mongo save timeout | request deadline ke andar |

### Deadline Example

```go
ctx, cancel := context.WithTimeout(ctx, 700*time.Millisecond)
defer cancel()

result, err := cms.ValidateCoupon(ctx, req)
if err != nil {
	return nil, ErrCouponPreviewUnavailable
}
```

### Why short timeout?

Coupon preview helpful feature hai, but cart page core mutation nahi rukni chahiye. Agar CMS down hai, buyer cart still use kar sakta hai. Checkout final validation baad me Order Service karega.

---

## Step 25: Observability Plan Banaya

Task 6 me future implementation ke liye logs, metrics, and traces define karna important hai.

### Logs

```json
{
  "level": "info",
  "service": "cart-service",
  "event": "coupon_preview_applied",
  "cart_id": "cart_123",
  "user_id": "user_123",
  "coupon_code": "SAVE10",
  "valid": true,
  "discount_amount": 2500,
  "currency": "INR",
  "duration_ms": 82
}
```

### Metrics

| Metric | Type | Meaning |
|---|---|---|
| `cart_coupon_preview_requests_total` | Counter | Total preview attempts |
| `cart_coupon_preview_valid_total` | Counter | Valid coupon previews |
| `cart_coupon_preview_invalid_total` | Counter | Invalid coupon previews by reason |
| `cart_coupon_preview_duration_ms` | Histogram | End-to-end latency |
| `cart_coupon_preview_cms_errors_total` | Counter | CMS call failures |
| `cart_coupon_preview_cache_refresh_errors_total` | Counter | Redis refresh failures |

### Trace Spans

```text
CartService.ApplyCouponPreview
├── Mongo.FindActiveByOwner
├── CMS.ValidateCoupon
├── Cart.ApplyCouponPreview
├── Mongo.SaveWithVersion
└── Redis.RefreshCartCache
```

> 🟢 **Privacy note:** Coupon code log karna generally okay ho sakta hai, but user/session identifiers should follow platform logging policy. Payment or personal address data log nahi karna.

---

## Step 26: Testing Strategy Banayi

Coupon preview cross-service flow hai, so tests unit + integration dono chahiye.

### Unit Tests

| Test Case | Expected Result |
|---|---|
| Empty coupon code | Validation error |
| Coupon code with spaces/lowercase | Normalized uppercase code |
| Missing owner | Owner/auth error |
| Cart not found | Not found |
| Cart not active | Precondition error |
| Empty cart | `EMPTY_CART` / precondition error |
| Optional `cart_id` mismatch | Permission error |
| CMS valid fixed discount | Discount applied to totals |
| CMS valid percentage result | CMS returned amount applied |
| CMS invalid coupon | Discount zero, reason stored |
| CMS discount currency mismatch | Error; cart not saved |
| CMS discount greater than subtotal | Error; cart not saved |
| CMS unavailable | Error; cart not mutated |
| Mongo version conflict | Retry with latest cart |
| Redis refresh fails | Response still succeeds after Mongo save |

### Integration Tests

| Test | Setup |
|---|---|
| Valid coupon preview persists cart | MongoDB test container/local stack |
| Redis cache updates after preview | Redis test container/local stack |
| CMS timeout returns unavailable | Fake CMS server with delayed response |
| Concurrent add item + coupon preview | Version retry expected |
| Invalid coupon clears previous discount | Start with valid preview, then invalid result |

### Example Unit Test Names

```text
TestApplyCouponPreviewNormalizesCouponCode
TestApplyCouponPreviewRejectsEmptyCart
TestApplyCouponPreviewAppliesValidDiscount
TestApplyCouponPreviewClearsDiscountWhenCouponInvalid
TestApplyCouponPreviewRejectsCurrencyMismatch
TestApplyCouponPreviewDoesNotMutateCartWhenCMSUnavailable
TestApplyCouponPreviewRetriesOnVersionConflict
TestApplyCouponPreviewReturnsSuccessWhenRedisRefreshFails
```

---

## Step 27: Security and Abuse Protection

Coupon preview endpoint abuse-prone ho sakta hai because users random coupon codes try kar sakte hain.

| Protection | Why |
|---|---|
| Auth/session owner required | Anonymous abuse and ownership bypass avoid |
| Rate limit per user/session | Coupon brute force reduce |
| Coupon code length limit | Payload abuse avoid |
| CMS timeout | Resource exhaustion avoid |
| No redemption write in preview | Usage counters fake increment nahi honge |
| Cart ownership validation | User dusre cart ka preview mutate nahi karega |
| No frontend discount trust | Buyer amount tamper nahi kar sakta |

### Suggested Rate Limit

```text
POST /api/v1/cart/coupons/preview
Limit: 20 attempts per user/session per minute
```

> 🔵 **Note:** Exact rate limit API Gateway/Auth security policy me tune ho sakta hai. Cart Service should still be defensive.

---

## Step 28: API Examples

### Valid Fixed Coupon

```http
POST /api/v1/cart/coupons/preview
Content-Type: application/json

{
  "coupon_code": "FLAT25"
}
```

```json
{
  "valid": true,
  "coupon_id": "coupon_flat25",
  "discount": {
    "amount": 2500,
    "currency": "INR"
  },
  "reason": ""
}
```

### Valid Percentage Coupon

```http
POST /api/v1/cart/coupons/preview
Content-Type: application/json

{
  "coupon_code": "SAVE10"
}
```

```json
{
  "valid": true,
  "coupon_id": "coupon_save10",
  "discount": {
    "amount": 1800,
    "currency": "INR"
  },
  "reason": ""
}
```

### Invalid Minimum Cart Coupon

```http
POST /api/v1/cart/coupons/preview
Content-Type: application/json

{
  "coupon_code": "BIGSAVE"
}
```

```json
{
  "valid": false,
  "coupon_id": "",
  "discount": {
    "amount": 0,
    "currency": "INR"
  },
  "reason": "MIN_CART_AMOUNT_NOT_MET"
}
```

---

## Step 29: Checkout Boundary Clearly Define Kiya

Cart coupon preview and checkout final apply different phases hain.

```mermaid
sequenceDiagram
    participant Buyer as Buyer
    participant Cart as Cart Service
    participant CMS as CMS Service
    participant Order as Order Service
    participant Payment as Payment Service

    Buyer->>Cart: ApplyCouponPreview(SAVE10)
    Cart->>CMS: ValidateCoupon(cart context)
    CMS-->>Cart: valid discount preview
    Cart-->>Buyer: Show estimated total

    Buyer->>Order: CreateOrderFromCart(coupon_code)
    Order->>Cart: GetCart
    Order->>CMS: ValidateCoupon(fresh checkout context)
    CMS-->>Order: final validation result
    Order->>Payment: Create payment intent with final amount
    Payment-->>Order: payment success
    Order->>CMS: RecordCouponRedemption
```

### Why Final Validation Required?

| Risk | Example |
|---|---|
| Coupon expired after preview | Buyer preview karta hai 11:59 PM, checkout 12:01 AM |
| Usage limit changed | Dusre users ne coupon use kar liya |
| Cart changed | Buyer ne eligible product remove kar diya |
| Seller/product status changed | Seller coupon pause kar deta hai |
| Price changed | Product price refresh checkout me change ho gaya |

> 🔴 **Final rule:** Cart Service preview is informational. Order Service final payable amount ka source hoga.

---

## Step 30: Acceptance Checklist

| Requirement | Status |
|---|---|
| `TaskImplementation/` folder present | ✅ Done |
| `TaskImplementation/Cart Service/` folder present | ✅ Done |
| `task6.md` created | ✅ Done |
| Step-by-step implementation in Hinglish | ✅ Done |
| Clear explanation of each part | ✅ Done |
| External libraries/tools mentioned | ✅ Done |
| Install/use instructions included | ✅ Done |
| Clean folder structure included | ✅ Done |
| Code examples included | ✅ Done |
| Mermaid diagrams included | ✅ Done |
| Proper formatting with headings, badges, emojis | ✅ Done |
| Scope limited to Cart Service Task 6 | ✅ Done |
| No backend implementation beyond documentation | ✅ Done |

---

## 🧾 Final Summary

Cart Service Task 6 ka final documented implementation standard:

```text
Frontend
  -> POST /api/v1/cart/coupons/preview
  -> CartService.ApplyCouponPreview
  -> Active cart load from MongoDB
  -> CMSService.ValidateCoupon
  -> Apply valid/invalid preview to cart totals
  -> Save cart with version guard
  -> Refresh Redis cache
  -> Return CouponPreviewResponse
```

### Key Decisions

1. Coupon preview Cart Service me dikhega, but final apply Order Service karega.
2. Coupon rules ka owner CMS Service hai.
3. Cart Service frontend-provided discount amount kabhi trust nahi karega.
4. Valid preview `coupon_code`, `coupon_preview`, and `totals.discount` update karega.
5. Invalid preview old discount clear karega and reason return karega.
6. MongoDB source of truth rahega; Redis best-effort cache rahega.
7. CMS unavailable hone par cart mutate nahi hoga.
8. Checkout ke time fresh CMS validation mandatory hoga.

> ✅ **Task 6 complete:** Coupon Preview ke liye documentation-level implementation guide ready hai. Actual backend, CMS engine, Order Service final apply, redemption recording, ya infrastructure files intentionally create nahi kiye gaye, kyunki current requested output sirf required folder structure aur `task6.md` content hai.
