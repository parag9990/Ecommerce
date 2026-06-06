# Step 1 - Micro Tasks

Status values: `Pending`, `WIP`, `Completed`.

Priority values:
- `P0`: foundation or blocker
- `P1`: core business feature
- `P2`: optimization, analytics, or advanced capability

## Platform Foundation

| S.No | Task Name | Task Detail (simple Hinglish + deep explanation) | Dependencies | Priority | Status |
|---:|---|---|---|---|---|
| 1 | Define repo standards | Naming, folder, branching, commit rules decide karo. Ye future team ko same style me kaam karne me help karega. | None | P0 | Pending |
| 2 | Create proto strategy | Har service ke gRPC contract ke liye `proto` folder aur versioning rule define karo. Contract pehle stable hoga to services loosely coupled rahengi. | Repo standards | P0 | Pending |
| 3 | Create shared Go libs | Logger, config, errors, middleware, tracing, validation jaise reusable packages banao. Har service me duplicate code kam hoga. | Repo standards | P0 | Pending |
| 4 | Docker Compose local stack | MySQL, MongoDB, Redis, Typesense, Kafka/RabbitMQ, Jaeger, Prometheus local run karne ke liye compose banao. | Repo standards | P0 | Pending |
| 5 | API gateway base | Gateway REST request receive karega, auth check karega, gRPC service call karega. Ye public entry point hoga. | Proto strategy | P0 | Pending |
| 6 | Observability baseline | Logs, metrics, traces ka format decide karo. Pehle din se monitoring ready rahegi. | Shared Go libs | P1 | Pending |
| 7 | CI pipeline skeleton | Lint, test, build, Docker image scan automatic banao. Team ke merge se pehle quality check hoga. | Repo standards | P1 | Pending |
| 8 | Kubernetes base manifests | Namespace, config map, secret, deployment, service, ingress templates banao. Later services fast deploy hongi. | Docker setup | P1 | Pending |

## User Service

| S.No | Task Name | Task Detail (simple Hinglish + deep explanation) | Dependencies | Priority | Status |
|---:|---|---|---|---|---|
| 1 | Define user domain | Buyer, seller, admin profile fields final karo. User Service sirf profile ownership rakhega, password Auth Service me rahega. | Platform foundation | P0 | Pending |
| 2 | Create MySQL schema | `users`, `addresses`, `seller_profiles`, `kyc_documents` tables banao. Structured relationships ke liye MySQL best hai. | User domain | P0 | Pending |
| 3 | Implement repository | DB queries ko repository layer me rakho. Business logic DB details se independent rahegi. | MySQL schema | P0 | Pending |
| 4 | Implement gRPC service | `GetUser`, `CreateUser`, `UpdateUser`, `GetSellerProfile` methods banao. Internal services direct profile data le sakenge. | Proto strategy | P0 | Pending |
| 5 | Add REST profile APIs | Gateway ke through profile view/update, address CRUD expose karo. Frontend simple REST use karega. | gRPC service | P1 | Pending |
| 6 | Add validation | Email, phone, address, GST, seller details validate karo. Bad data DB me enter nahi hoga. | REST APIs | P1 | Pending |
| 7 | Add audit fields | Created by, updated by, status, timestamps maintain karo. Compliance aur debugging easy hogi. | Schema | P1 | Pending |
| 8 | Add user events | `UserCreated`, `SellerApproved`, `AddressUpdated` publish karo. Notification, analytics, recommendation consume karenge. | Message queue | P2 | Pending |

## Auth Service

| S.No | Task Name | Task Detail (simple Hinglish + deep explanation) | Dependencies | Priority | Status |
|---:|---|---|---|---|---|
| 1 | Define auth flows | Signup, login, OTP, refresh token, logout, password reset, MFA flows document karo. Ye security ka core hai. | Platform foundation | P0 | Pending |
| 2 | Create MySQL schema | `auth_accounts`, `credentials`, `refresh_tokens`, `otp_challenges`, `role_assignments` tables banao. Strong consistency chahiye. | Auth flows | P0 | Pending |
| 3 | Password hashing | Argon2id or bcrypt with strong cost use karo. Plain password kabhi store nahi hoga. | Schema | P0 | Pending |
| 4 | JWT issuing | Access token short lived, refresh token rotating banao. Claims me user id, tenant/seller id, roles, session id rakho. | Credentials | P0 | Pending |
| 5 | OTP verification | Email and phone OTP generate, hash, expire, retry limit implement karo. Brute force se bachna zaruri hai. | Notification Service | P1 | Pending |
| 6 | RBAC middleware | Buyer, Seller, Admin, Superadmin role checks centralize karo. Gateway aur services dono enforce karenge. | JWT | P0 | Pending |
| 7 | Session link | Login ke time Session Management Service ko session start event bhejo. Analytics and fraud detection possible hoga. | Session Service | P1 | Pending |
| 8 | Security tests | Token expiry, refresh reuse, OTP replay, role bypass test cases likho. Auth bugs high risk hote hain. | Implementation | P1 | Pending |

## Product Service

| S.No | Task Name | Task Detail (simple Hinglish + deep explanation) | Dependencies | Priority | Status |
|---:|---|---|---|---|---|
| 1 | Define catalog model | Product, variant, category, attribute, image, inventory rules finalize karo. E-commerce catalog flexible hona chahiye. | Platform foundation | P0 | Pending |
| 2 | Choose MongoDB | Product attributes category-wise dynamic hote hain, isliye MongoDB flexible schema ke liye better hai. | Catalog model | P0 | Pending |
| 3 | Create collections | `products`, `categories`, `brands`, `inventory_snapshots`, `price_books` collections design karo. | MongoDB choice | P0 | Pending |
| 4 | Implement seller CRUD | Seller apne products create/update/publish kar sake. Draft and published workflow banao. | CMS Service | P1 | Pending |
| 5 | Implement read APIs | Product listing, detail, category browse, seller catalog APIs banao. Fast read path ke liye indexes zaruri hain. | Collections | P1 | Pending |
| 6 | Inventory operations | Reserve, release, decrement stock methods implement karo. Checkout race condition avoid hogi. | Order Service | P0 | Pending |
| 7 | Publish search events | Product changes ko Search Service me index karne ke liye events publish karo. Search data async sync hoga. | Message queue | P1 | Pending |
| 8 | Add media metadata | Images CDN URLs, alt text, order, status maintain karo. Frontend clean gallery render karega. | Collections | P2 | Pending |

## Order Service

| S.No | Task Name | Task Detail (simple Hinglish + deep explanation) | Dependencies | Priority | Status |
|---:|---|---|---|---|---|
| 1 | Define order lifecycle | Created, pending payment, paid, packed, shipped, delivered, cancelled, refunded states define karo. | Platform foundation | P0 | Pending |
| 2 | Create MySQL schema | Orders transactional hote hain, isliye MySQL with relational tables best hai. | Lifecycle | P0 | Pending |
| 3 | Cart to order flow | Cart validate karo, price snapshot lo, inventory reserve karo, order create karo. | Cart and Product | P0 | Pending |
| 4 | Payment coordination | Payment Service ko payment intent request bhejo, result ke basis pe order status update karo. | Payment Service | P0 | Pending |
| 5 | Implement order gRPC | `CreateOrder`, `GetOrder`, `ListOrders`, `UpdateFulfillment` methods banao. | Proto strategy | P1 | Pending |
| 6 | Seller order view | Seller apne items ke orders dekh sake. Multi-seller order splitting handle karo. | User and Product | P1 | Pending |
| 7 | Idempotency | Checkout retry pe duplicate order na bane. Idempotency key store and enforce karo. | Schema | P0 | Pending |
| 8 | Emit order events | `OrderCreated`, `OrderPaid`, `OrderCancelled`, `OrderDelivered` events publish karo. | Message queue | P1 | Pending |

## Payment Service

| S.No | Task Name | Task Detail (simple Hinglish + deep explanation) | Dependencies | Priority | Status |
|---:|---|---|---|---|---|
| 1 | Define payment state machine | Initiated, authorized, captured, failed, refunded, partially refunded states clear karo. | Order lifecycle | P0 | Pending |
| 2 | Create MySQL schema | Payment financial record hai, consistency and audit ke liye MySQL use karo. | State machine | P0 | Pending |
| 3 | Gateway abstraction | Stripe/Razorpay-like providers ke liye interface banao. Provider swap karna easy hoga. | Schema | P0 | Pending |
| 4 | Create payment intent | Order amount, currency, customer, idempotency key ke saath provider intent create karo. | Order Service | P0 | Pending |
| 5 | Webhook handler | Provider webhook signature verify karo, payment status update karo. Webhook source of truth hoga. | Gateway provider | P0 | Pending |
| 6 | Refund flow | Full and partial refunds support karo. Refund records immutable rakho. | Paid payments | P1 | Pending |
| 7 | Retry handling | Failed payment retry allowed karo without duplicate charge. Idempotency mandatory hai. | Intent flow | P1 | Pending |
| 8 | Reconciliation job | Provider settlement report compare karo. Financial mismatch alert karo. | Monitoring | P2 | Pending |

## Cart Service

| S.No | Task Name | Task Detail (simple Hinglish + deep explanation) | Dependencies | Priority | Status |
|---:|---|---|---|---|---|
| 1 | Define cart rules | Guest cart, logged-in cart, merge, quantity limit, price refresh rules define karo. | Platform foundation | P0 | Pending |
| 2 | Choose MongoDB plus Redis | Active cart fast mutable hot data hai. Redis cache, Mongo durable snapshot ke liye use karo. | Cart rules | P0 | Pending |
| 3 | Cart collections | `carts`, `cart_items`, embedded item snapshots design karo. | DB choice | P0 | Pending |
| 4 | Add item flow | Product validate karo, stock check karo, item add/update karo. | Product Service | P0 | Pending |
| 5 | Remove item flow | Item remove, quantity zero cleanup, totals recalculate karo. | Add item | P1 | Pending |
| 6 | Coupon preview | CMS coupon rules validate karke cart discount preview karo. Final apply Order Service karega. | CMS Service | P1 | Pending |
| 7 | Cart merge | Guest session cart ko login user cart me merge karo. Duplicate variants quantity combine hongi. | Auth and Session | P1 | Pending |
| 8 | Cart expiry | Inactive carts cleanup job banao. Storage and stale prices control me rahenge. | Scheduler | P2 | Pending |

## Wishlist Service

| S.No | Task Name | Task Detail (simple Hinglish + deep explanation) | Dependencies | Priority | Status |
|---:|---|---|---|---|---|
| 1 | Define wishlist model | User multiple wishlists rakhe ya single list, sharing allowed hai ya nahi decide karo. | Platform foundation | P1 | Pending |
| 2 | Choose MongoDB | Wishlist item list flexible aur read-heavy hai, MongoDB simple document model fit hai. | Model | P1 | Pending |
| 3 | Collections design | `wishlists` collection with user id, items, visibility, timestamps banao. | DB choice | P1 | Pending |
| 4 | Add/remove item APIs | Product id validate karke wishlist item add/remove karo. Duplicate item block karo. | Product Service | P1 | Pending |
| 5 | Move to cart | Wishlist item ko Cart Service me add karne ka flow banao. UX friction kam hoga. | Cart Service | P1 | Pending |
| 6 | Availability sync | Product deleted/out-of-stock hone pe wishlist status update karo. | Product events | P2 | Pending |
| 7 | Price drop events | Price change pe interested users ko notification trigger karo. | Notification Service | P2 | Pending |
| 8 | Analytics events | Wishlist add/remove events Recommendation Service ko bhejo. Personalization improve hoga. | Message queue | P2 | Pending |

## Recommendation Service

| S.No | Task Name | Task Detail (simple Hinglish + deep explanation) | Dependencies | Priority | Status |
|---:|---|---|---|---|---|
| 1 | Define recommendation types | Similar products, trending, personalized, frequently bought together define karo. | Product and Session | P2 | Pending |
| 2 | Choose MongoDB plus Redis | Recommendation documents flexible hain; Redis top lists and low-latency cache ke liye use hoga. | Types | P2 | Pending |
| 3 | Event ingestion | Product views, add-to-cart, wishlist, purchase events consume karo. | Kafka/RabbitMQ | P1 | Pending |
| 4 | Feature store schema | User-product interaction counters and embeddings/reference ids store karo. | Event ingestion | P2 | Pending |
| 5 | Rule-based MVP | Trending, category popular, seller popular recommendations pehle banao. ML later add karna easier hoga. | Product data | P1 | Pending |
| 6 | Personalized ranking | User behavior based scoring implement karo. Cold-start fallback mandatory hai. | Feature store | P2 | Pending |
| 7 | gRPC endpoint | `GetRecommendations(user_id, context)` method banao. Frontend and Product Service use karenge. | Proto strategy | P1 | Pending |
| 8 | A/B testing hooks | Strategy id return karo so analytics conversion compare kar sake. | Session analytics | P2 | Pending |

## Search Service (Typesense)

| S.No | Task Name | Task Detail (simple Hinglish + deep explanation) | Dependencies | Priority | Status |
|---:|---|---|---|---|---|
| 1 | Define search schema | Product searchable fields, facets, sorting, typo tolerance, synonyms define karo. | Product catalog | P0 | Pending |
| 2 | Setup Typesense | Local and Kubernetes Typesense cluster configure karo. Search data yahi index hoga. | External services | P0 | Pending |
| 3 | Product indexer | Product events consume karke Typesense document upsert/delete karo. | Product events | P0 | Pending |
| 4 | Search API | Query, filter, facet, sort, pagination endpoint banao. | Typesense setup | P1 | Pending |
| 5 | Autocomplete API | Prefix search and popular queries return karo. Header search fast feel karega. | Search schema | P1 | Pending |
| 6 | Synonym management | CMS/Superadmin synonyms add/update kar sake. Merchandising improve hogi. | CMS/Superadmin | P2 | Pending |
| 7 | Zero-result tracking | No result queries analytics me bhejo. Catalog and synonyms improve karne me help hoga. | Session Service | P2 | Pending |
| 8 | Reindex job | Full catalog reindex command and job banao. Schema change safe hoga. | Product Service | P1 | Pending |

## CMS Service

| S.No | Task Name | Task Detail (simple Hinglish + deep explanation) | Dependencies | Priority | Status |
|---:|---|---|---|---|---|
| 1 | Define seller permissions | Seller manager, catalog editor, order manager roles define karo. | Auth RBAC | P1 | Pending |
| 2 | Choose MySQL | Coupons, offers, seller settings, workflows structured hain, relational model fit hai. | Permissions | P1 | Pending |
| 3 | Product moderation flow | Seller draft, submit, approve/reject, publish states design karo. | Product Service | P1 | Pending |
| 4 | Coupon engine MVP | Fixed, percentage, min cart, category/product/seller scope rules banao. | Cart and Order | P1 | Pending |
| 5 | Offer campaigns | Start/end time, budget, usage limits, seller ownership implement karo. | Coupon engine | P2 | Pending |
| 6 | Seller analytics APIs | Revenue, orders, conversion, top products metrics expose karo. | Order and Payment | P2 | Pending |
| 7 | CMS gRPC | Validate coupon, get seller settings, get campaign methods banao. | Proto strategy | P1 | Pending |
| 8 | Audit logs | Seller dashboard actions audit karo. Disputes me traceability milegi. | Schema | P1 | Pending |

## Session Management Service (Independent)

| S.No | Task Name | Task Detail (simple Hinglish + deep explanation) | Dependencies | Priority | Status |
|---:|---|---|---|---|---|
| 1 | Define session model | Anonymous id, user id, device, IP, user agent, channel, started/ended timestamps define karo. | Platform foundation | P0 | Pending |
| 2 | Choose MongoDB plus Redis | Events high-volume flexible hote hain. Mongo store, Redis active sessions ke liye use karo. | Session model | P0 | Pending |
| 3 | Event ingestion API | Frontend SDK page view, click, scroll, cart action events bhej sake. | API Gateway | P0 | Pending |
| 4 | Journey tracking | Session ke andar ordered events store karo. User journey reconstruct ho sakegi. | Event API | P1 | Pending |
| 5 | Device tracking | Browser, OS, device, location approximation parse karo. Fraud and analytics me useful hai. | Event API | P1 | Pending |
| 6 | Heatmap concept | Click and scroll coordinates aggregate karo. UI optimization ke liye dashboard me show hoga. | Event ingestion | P2 | Pending |
| 7 | Analytics APIs | Active users, funnels, conversion, retention, session replay metadata expose karo. | Aggregations | P1 | Pending |
| 8 | Data retention policy | Raw events TTL, aggregated metrics long-term store rules define karo. Cost control ke liye zaruri hai. | MongoDB | P1 | Pending |

## Notification Service

| S.No | Task Name | Task Detail (simple Hinglish + deep explanation) | Dependencies | Priority | Status |
|---:|---|---|---|---|---|
| 1 | Define channels | Email, SMS, push, WhatsApp-like provider abstraction define karo. | Platform foundation | P1 | Completed |
| 2 | Choose MongoDB | Notification templates and delivery logs flexible hain, MongoDB suitable hai. | Channels | P1 | Completed |
| 3 | Template engine | OTP, order updates, payment, promotional templates with variables banao. | DB choice | P1 | Completed |
| 4 | Send OTP | Auth Service ke OTP request ke liye email/SMS send method implement karo. | Auth Service | P0 | Completed |
| 5 | Event consumers | Order/payment/user events consume karke notifications trigger karo. | Message queue | P1 | Completed |
| 6 | Retry and DLQ | Provider failure pe retry, final failure pe dead-letter queue. Reliability improve hogi. | Queue | P1 | Completed |
| 7 | User preferences | User opt-in/opt-out and channel preference respect karo. Compliance ke liye important hai. | User Service | P2 | Completed |
| 8 | Delivery analytics | Sent, delivered, failed, opened metrics collect karo. Campaign quality measure hogi. | Monitoring | P2 | Completed |

## API Gateway

| S.No | Task Name | Task Detail (simple Hinglish + deep explanation) | Dependencies | Priority | Status |
|---:|---|---|---|---|---|
| 1 | Define public routes | Frontend ke REST endpoints map karo: auth, products, cart, checkout, seller, admin. | Platform foundation | P0 | Pending |
| 2 | gRPC clients setup | Gateway ke andar har microservice ka gRPC client configure karo. | Proto strategy | P0 | Pending |
| 3 | Auth middleware | JWT validate, user context inject, RBAC enforce karo. Public vs protected routes clear rakho. | Auth Service | P0 | Pending |
| 4 | Rate limiting | IP, user, route based limits Redis se implement karo. Abuse and bots control honge. | Redis | P0 | Pending |
| 5 | Request validation | DTO validation, size limits, content-type checks add karo. Service ko clean input milega. | Route definitions | P1 | Pending |
| 6 | Error mapping | gRPC errors ko REST error format me convert karo. Frontend consistent error handle karega. | Shared errors | P1 | Pending |
| 7 | Observability | Request id, logs, metrics, traces har request me add karo. Production debugging easy hogi. | Logging baseline | P1 | Pending |
| 8 | gRPC-Web bridge | Browser clients ke liye Envoy or gateway bridge configure karo. Real-time-like typed APIs possible hongi. | Frontend | P2 | Pending |

## Superadmin Service

| S.No | Task Name | Task Detail (simple Hinglish + deep explanation) | Dependencies | Priority | Status |
|---:|---|---|---|---|---|
| 1 | Define admin domain | Users, sellers, orders, payments, sessions, platform settings ke control rules define karo. | Auth RBAC | P1 | Pending |
| 2 | Choose MySQL | Admin actions and permissions highly structured and auditable hain, MySQL suitable hai. | Admin domain | P1 | Pending |
| 3 | Admin RBAC | Superadmin, operations admin, finance admin, catalog admin permissions define karo. | Auth Service | P0 | Pending |
| 4 | User/seller controls | Block user, approve seller, suspend seller, verify KYC flows banao. | User Service | P1 | Pending |
| 5 | Order/payment controls | Refund approve, manual order status review, dispute view APIs banao. | Order and Payment | P1 | Pending |
| 6 | Session visibility | Session analytics dashboard ko admin level access control do. | Session Service | P2 | Pending |
| 7 | Platform settings | Search synonyms, commission, feature flags, maintenance mode settings banao. | CMS/Search | P2 | Pending |
| 8 | Admin audit logs | Har admin action immutable audit log me store karo. Compliance and security ke liye must hai. | Schema | P0 | Pending |

## User App Frontend

| S.No | Task Name | Task Detail (simple Hinglish + deep explanation) | Dependencies | Priority | Status |
|---:|---|---|---|---|---|
| 1 | Setup React TS app | Vite/React/TypeScript/Tailwind setup karo. Strict TS and linting enable karo. | Repo foundation | P0 | Pending |
| 2 | App shell | Header, nav, search bar, account menu, cart badge layout banao. | Design system | P0 | Pending |
| 3 | Auth screens | Signup, login, OTP, forgot password screens banao. Form validation strong rakho. | Auth APIs | P0 | Pending |
| 4 | Product browsing | Home listing, category listing, filters, sort, product detail pages banao. | Product/Search APIs | P1 | Pending |
| 5 | Cart and checkout | Cart, address select, coupon, payment intent, order success/failure flows banao. | Cart/Order/Payment | P0 | Pending |
| 6 | Profile module | Profile, addresses, orders, wishlist pages banao. | User/Order/Wishlist | P1 | Pending |
| 7 | State management | Zustand for UI/session state, React Query for server cache. Global state minimal rakho. | App shell | P1 | Pending |
| 8 | gRPC-Web client | Proto generated TS client integrate karo for typed internal calls where needed. | Proto and Envoy | P2 | Pending |

## Seller Dashboard (CMS)

| S.No | Task Name | Task Detail (simple Hinglish + deep explanation) | Dependencies | Priority | Status |
|---:|---|---|---|---|---|
| 1 | Dashboard shell | Sidebar, topbar, seller switcher, protected layout banao. Operational UI compact rakho. | React setup | P1 | Pending |
| 2 | Product manager | Product list, create/edit form, variants, images, publish status banao. | CMS/Product APIs | P1 | Pending |
| 3 | Order manager | Seller order list, status updates, shipment info, refunds view banao. | Order APIs | P1 | Pending |
| 4 | Offers and coupons | Coupon create/edit, campaign calendar, usage stats UI banao. | CMS APIs | P1 | Pending |
| 5 | Revenue analytics | Revenue, GMV, orders, conversion, top products charts banao. | CMS analytics | P2 | Pending |
| 6 | Team permissions | Seller staff invite, roles, access control UI banao. | Auth/CMS | P2 | Pending |
| 7 | Audit activity | Recent seller actions timeline banao. Debug and trust improve hota hai. | CMS audit APIs | P2 | Pending |
| 8 | Error states | Empty, loading, failed, permission denied states polish karo. Production UX reliable lagega. | All modules | P1 | Pending |

## Session Analytics Dashboard

| S.No | Task Name | Task Detail (simple Hinglish + deep explanation) | Dependencies | Priority | Status |
|---:|---|---|---|---|---|
| 1 | Analytics shell | Date range, segment filters, metric cards layout banao. Data-heavy UI clean and scannable rakho. | Session APIs | P1 | Pending |
| 2 | Active sessions | Live active users, devices, locations, entry pages view banao. | Session Service | P1 | Pending |
| 3 | Journey explorer | Session timeline, page views, clicks, cart events sequence display karo. | Journey APIs | P1 | Pending |
| 4 | Funnel analysis | Product view to cart to checkout to paid funnel chart banao. | Aggregation APIs | P2 | Pending |
| 5 | Heatmap view | Conceptual click/scroll heatmap overlay render karo. Exact replay later phase me add hoga. | Heatmap APIs | P2 | Pending |
| 6 | Retention reports | New vs returning, cohort retention charts banao. | Aggregates | P2 | Pending |
| 7 | Export reports | CSV download and scheduled report option banao. | Analytics APIs | P2 | Pending |
| 8 | Privacy controls | PII masking, user deletion, retention settings UI banao. | Security policy | P1 | Pending |

## Superadmin Panel

| S.No | Task Name | Task Detail (simple Hinglish + deep explanation) | Dependencies | Priority | Status |
|---:|---|---|---|---|---|
| 1 | Admin shell | Secure admin layout, navigation, role-based menu banao. Sirf allowed modules visible honge. | Auth RBAC | P0 | Pending |
| 2 | User management | User search, block/unblock, profile view, session view banao. | User/Superadmin APIs | P1 | Pending |
| 3 | Seller management | Seller KYC approval, suspension, catalog review UI banao. | User/CMS/Product | P1 | Pending |
| 4 | Order operations | Order search, detail, dispute, manual review UI banao. | Order APIs | P1 | Pending |
| 5 | Payment operations | Payment status, refunds, reconciliation alerts UI banao. | Payment APIs | P1 | Pending |
| 6 | Session oversight | High-risk sessions, live traffic, suspicious activity view banao. | Session APIs | P2 | Pending |
| 7 | Platform settings | Commission, search synonyms, feature flags, maintenance mode UI banao. | Superadmin APIs | P2 | Pending |
| 8 | Audit log viewer | Admin actions filterable table with export banao. Security review ke liye important hai. | Audit APIs | P0 | Pending |

