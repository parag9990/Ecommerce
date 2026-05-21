# 👤 User Service - Task 2: Create MySQL Schema

![Task](https://img.shields.io/badge/Task-User%20Service%20%232-blue)
![Priority](https://img.shields.io/badge/Priority-P0-red)
![Dependency](https://img.shields.io/badge/Dependency-User%20Domain-yellow)
![Status](https://img.shields.io/badge/Status-Documented-success)
![Guide](https://img.shields.io/badge/Guide-Hinglish-orange)
![Scope](https://img.shields.io/badge/Scope-MySQL%20Schema-purple)

## 📌 Task Summary

| Field | Detail |
|---|---|
| Task Name | Create MySQL schema |
| Source | `docs/01-micro-tasks.md` → `User Service` → Task 2 |
| Priority | `P0` core profile foundation |
| Dependency | User Service Task 1: Define user domain |
| Main Goal | `users`, `user_addresses`, `seller_profiles`, aur `seller_kyc_documents` tables design karna |
| Core Boundary | User Service sirf profile, address, seller profile, aur KYC metadata own karega |
| Output Type | Structured implementation guide |
| Not Included | Repository code, gRPC service, REST APIs, validation layer, events, Superadmin workflow implementation |

> **Simple Hinglish goal:** Is task ka purpose User Service ke domain ko MySQL schema me convert karna hai. Task 1 me profile ownership clear ho chuki hai; ab tables, relationships, indexes, constraints, migrations, rollback, aur verification queries document kiye gaye hain.

---

## ✅ Final Output Created

```text
TaskImplementation/
├── Auth Service/
├── Platform Foundation/
└── User Service/
    ├── task1.md
    └── task2.md
```

### Why this structure?

- `TaskImplementation/` project ke task-wise guides ka central folder hai.
- `User Service/` folder already present tha, isliye usko keep kiya gaya.
- `task2.md` sirf **User Service - Task 2** ka guide hai.
- Existing `task1.md`, Auth Service guides, Platform Foundation guides, aur backend code untouched rakhe gaye.
- Actual `backend/services/user-service/` migration files create nahi kiye gaye, kyunki requested output sirf folder structure aur `task2.md` content generate karna hai.

---

## 🧭 Implementation Approach

Is guide ko banate time project ke official docs ko source of truth maana gaya:

| Document | Kya use kiya gaya |
|---|---|
| `docs/01-micro-tasks.md` | User Service Task 2 ka exact scope: MySQL schema |
| `TaskImplementation/User Service/task1.md` | User domain boundary, entity ownership, auth split |
| `docs/03-folder-structure.md` | Future `backend/services/user-service/migrations/` folder convention |
| `docs/04-microservice-design.md` | User Service responsibilities, DB choice, table names |
| `docs/05-database-design.md` | User DB relationships, indexes, scalability notes |
| `database/draw.sql` | Existing detailed User Service DDL reference |
| `api/master-api.json` | User profile, address, seller profile DTO field expectations |
| `docs/09-cms-superadmin.md` | Seller KYC approval flow and Superadmin touchpoint |
| `docs/13-developer-guide.md` | Migration and backend layer rules |

---

## 🧱 Task Boundary

### ✅ Included in Task 2

- User Service MySQL database schema design
- `user_db` database creation strategy
- `users` table design
- `user_addresses` table design
- `seller_profiles` table design
- `seller_kyc_documents` table design
- Primary keys, unique keys, foreign keys, and indexes
- Migration up/down SQL examples
- ER diagram and request flow diagrams
- Beginner-friendly table explanations
- External tools/libraries explanation
- Verification queries and sanity checks
- Schema-level security and data ownership notes

### 🚫 Not Included in Task 2

- `backend/services/user-service/` code generation
- Repository implementation
- gRPC server implementation
- REST route implementation
- Runtime validation implementation
- Redis cache
- Kafka/RabbitMQ event publishing
- Seller approval workflow implementation
- KYC file upload implementation
- Notification integration

> 🟢 **Rule:** Task 2 ka scope database foundation hai. Repository Task 3, gRPC Task 4, REST APIs Task 5, validation Task 6, audit fields Task 7, aur user events Task 8 me aayenge.

---

## 🗂️ Target Implementation Folder Structure

Future me actual User Service implementation ka schema part yahan rahega:

```text
backend/
└── services/
    └── user-service/
        ├── cmd/
        │   └── server/
        │       └── main.go
        ├── internal/
        │   ├── domain/
        │   │   ├── user.go
        │   │   ├── address.go
        │   │   └── seller_profile.go
        │   ├── usecase/
        │   │   ├── create_user.go
        │   │   ├── update_profile.go
        │   │   ├── manage_address.go
        │   │   └── seller_kyc.go
        │   ├── repository/
        │   │   └── mysql_user_repository.go
        │   ├── transport/
        │   │   └── grpc/
        │   │       └── server.go
        │   └── config/
        │       └── config.go
        ├── migrations/
        │   ├── 001_create_user_tables.up.sql
        │   └── 001_create_user_tables.down.sql
        └── deploy/
            ├── Dockerfile
            └── k8s.yaml
```

### Folder responsibility

| Path | Responsibility |
|---|---|
| `migrations/001_create_user_tables.up.sql` | User DB tables create karega |
| `migrations/001_create_user_tables.down.sql` | Rollback ke time tables safely drop karega |
| `internal/domain/` | Table-backed business entities define karega |
| `internal/repository/` | MySQL queries and transactions implement karega |
| `internal/usecase/` | Create/update profile and address workflows rakhega |
| `internal/transport/grpc/` | Future gRPC request/response mapping karega |

> 🟡 **Important:** Ye target structure hai. Is Task 2 me sirf `TaskImplementation/User Service/task2.md` create kiya gaya.

---

## 🧩 Schema Architecture

```mermaid
erDiagram
    USERS ||--o{ USER_ADDRESSES : has
    USERS ||--o| SELLER_PROFILES : may_have
    SELLER_PROFILES ||--o{ SELLER_KYC_DOCUMENTS : has

    USERS {
        bigint id PK
        string user_id UK
        string auth_account_id UK
        string email UK
        string phone
        string full_name
        string avatar_url
        enum status
        timestamp created_at
        timestamp updated_at
    }

    USER_ADDRESSES {
        bigint id PK
        string address_id UK
        string user_id FK
        string name
        string phone
        string line1
        string line2
        string city
        string state
        string postal_code
        string country
        bool is_default
        timestamp created_at
        timestamp updated_at
        timestamp deleted_at
    }

    SELLER_PROFILES {
        bigint id PK
        string seller_id UK
        string user_id UK,FK
        string store_name
        string display_name
        string gst_number
        string support_email
        enum status
        string approved_by
        timestamp approved_at
        timestamp created_at
        timestamp updated_at
    }

    SELLER_KYC_DOCUMENTS {
        bigint id PK
        string document_id UK
        string seller_id FK
        string document_type
        string storage_url
        enum status
        string reviewed_by
        timestamp reviewed_at
        string rejection_reason
        timestamp created_at
    }
```

### Hinglish explanation

- Ek `user` ke multiple addresses ho sakte hain.
- Ek `user` seller ban sakta hai, but ek user ka sirf ek seller profile hoga.
- Ek seller profile ke multiple KYC documents ho sakte hain.
- Password, refresh token, OTP, aur roles yahan store nahi honge. Wo Auth Service ka domain hai.
- Actual KYC file binary DB me store nahi hogi. DB me sirf `storage_url` metadata rahega.

---

## 🪜 Step-by-Step Implementation

## Step 1: Database Choice Confirm Kiya

User Service ke liye MySQL choose kiya gaya.

### Why MySQL?

| Reason | Explanation |
|---|---|
| Structured data | User, address, seller profile, KYC records fixed fields wale structured entities hain |
| Relationships | User → addresses, user → seller profile, seller → documents clearly relational hain |
| Constraints | Unique email, unique auth account, one seller profile per user enforce karna easy hai |
| Transactions | Address default change ya seller profile update atomic karna possible hai |
| Indexes | Profile lookup, seller status filter, default address query fast ho sakti hai |

### Database name

```sql
CREATE DATABASE IF NOT EXISTS user_db
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci;
```

**Why `utf8mb4`?**  
Names, addresses, store names, aur multilingual text future me safely store ho sake. `utf8mb4` emoji and full Unicode support bhi deta hai.

---

## Step 2: `users` Table Design Kiya

`users` table User Service ka main profile table hai.

### Purpose

Ye table buyer, seller, aur admin ke base profile ko represent karega. Har user ka ek stable `user_id` hoga, aur Auth Service ke account se mapping ke liye `auth_account_id` hoga.

### SQL

```sql
CREATE TABLE IF NOT EXISTS users (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id VARCHAR(64) NOT NULL,
  auth_account_id VARCHAR(64) NOT NULL,
  email VARCHAR(255) NOT NULL,
  phone VARCHAR(32) NULL,
  full_name VARCHAR(255) NOT NULL,
  avatar_url VARCHAR(1024) NULL,
  status ENUM('active', 'blocked', 'deleted') NOT NULL DEFAULT 'active',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_users_user_id (user_id),
  UNIQUE KEY uk_users_auth_account_id (auth_account_id),
  UNIQUE KEY uk_users_email (email),
  KEY idx_users_status (status)
) ENGINE=InnoDB;
```

### Field explanation

| Field | Meaning | Why needed |
|---|---|---|
| `id` | Internal numeric primary key | MySQL clustered PK efficient rahega |
| `user_id` | Public/internal stable user identifier | Other services user ko isi id se refer karenge |
| `auth_account_id` | Auth Service account reference | Signup ke baad Auth account se profile link hoga |
| `email` | Email mirror | Profile display/contact ke liye; login ownership Auth Service me rahega |
| `phone` | Optional phone mirror | Delivery/contact use cases ke liye |
| `full_name` | User display name | Account, checkout, admin views me use hoga |
| `avatar_url` | Optional CDN/object URL | Profile image ke liye |
| `status` | User lifecycle | `active`, `blocked`, `deleted` states |
| `created_at` | Creation timestamp | Audit and sorting |
| `updated_at` | Last update timestamp | Profile freshness tracking |

### Important constraints

| Constraint | Reason |
|---|---|
| `uk_users_user_id` | Har profile ka stable unique ID |
| `uk_users_auth_account_id` | Ek Auth account ke against sirf ek User profile |
| `uk_users_email` | Duplicate email profile avoid karna |
| `idx_users_status` | Admin list/filter me status query fast karna |

> 🟠 **Note:** `email` and `phone` User Service me mirror fields hain. Login identity ka final source Auth Service hi rahega.

---

## Step 3: `user_addresses` Table Design Kiya

`user_addresses` table user ke shipping/billing address book ko store karega.

### Purpose

Ek user ke multiple addresses ho sakte hain. Checkout ke time default address fast fetch hona chahiye, aur deleted addresses ko soft delete karna useful hota hai.

### SQL

```sql
CREATE TABLE IF NOT EXISTS user_addresses (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  address_id VARCHAR(64) NOT NULL,
  user_id VARCHAR(64) NOT NULL,
  name VARCHAR(255) NOT NULL,
  phone VARCHAR(32) NULL,
  line1 VARCHAR(255) NOT NULL,
  line2 VARCHAR(255) NULL,
  city VARCHAR(128) NOT NULL,
  state VARCHAR(128) NOT NULL,
  postal_code VARCHAR(32) NOT NULL,
  country VARCHAR(64) NOT NULL,
  is_default BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_user_addresses_address_id (address_id),
  KEY idx_user_addresses_user_default (user_id, is_default),
  CONSTRAINT fk_user_addresses_user FOREIGN KEY (user_id) REFERENCES users(user_id)
) ENGINE=InnoDB;
```

### Field explanation

| Field | Meaning | Why needed |
|---|---|---|
| `address_id` | Public address identifier | REST/gRPC APIs address ko expose karenge |
| `user_id` | Address owner | User ke saath relation |
| `name` | Receiver name | Delivery ke time required |
| `phone` | Receiver phone | Delivery partner contact ke liye |
| `line1`, `line2` | Street address lines | Address detail |
| `city`, `state`, `postal_code`, `country` | Location fields | Delivery, tax, shipping zones ke liye |
| `is_default` | Default address flag | Checkout me preselected address |
| `deleted_at` | Soft delete marker | Old orders/address history safely preserve |

### Address behavior rules

| Rule | Explanation |
|---|---|
| User must exist first | Foreign key `users(user_id)` enforce karega |
| Address delete soft hoga | `deleted_at` set hoga, row remove nahi hogi |
| Default address one active honi chahiye | Repository/usecase transaction me enforce hoga |
| Address limit future me enforce hoga | Docs me example max 20 addresses per user diya gaya hai |

### Default address transaction idea

Default address set karte time two-step transaction useful hoga:

```sql
START TRANSACTION;

UPDATE user_addresses
SET is_default = FALSE
WHERE user_id = 'user_123'
  AND deleted_at IS NULL;

UPDATE user_addresses
SET is_default = TRUE
WHERE user_id = 'user_123'
  AND address_id = 'addr_123'
  AND deleted_at IS NULL;

COMMIT;
```

**Hinglish explanation:** Pehle user ke saare active addresses non-default karo, phir selected address default banao. Ye transaction me hoga, isliye beech me inconsistent state permanent nahi hogi.

---

## Step 4: `seller_profiles` Table Design Kiya

`seller_profiles` table marketplace seller ka business profile store karega.

### Purpose

Buyer user seller ban sakta hai. Seller profile separate table me rakha gaya hai kyunki seller fields buyer profile se alag hain: store name, GST, support email, approval status, etc.

### SQL

```sql
CREATE TABLE IF NOT EXISTS seller_profiles (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  seller_id VARCHAR(64) NOT NULL,
  user_id VARCHAR(64) NOT NULL,
  store_name VARCHAR(255) NOT NULL,
  display_name VARCHAR(255) NULL,
  gst_number VARCHAR(64) NULL,
  support_email VARCHAR(255) NULL,
  status ENUM('draft', 'pending_review', 'active', 'suspended', 'rejected') NOT NULL DEFAULT 'draft',
  approved_by VARCHAR(64) NULL,
  approved_at TIMESTAMP NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_seller_profiles_seller_id (seller_id),
  UNIQUE KEY uk_seller_profiles_user_id (user_id),
  KEY idx_seller_profiles_status (status),
  CONSTRAINT fk_seller_profiles_user FOREIGN KEY (user_id) REFERENCES users(user_id)
) ENGINE=InnoDB;
```

### Field explanation

| Field | Meaning | Why needed |
|---|---|---|
| `seller_id` | Stable seller identifier | Product/CMS/Order services seller ko isi id se refer karenge |
| `user_id` | Seller owner user | One user may have one seller profile |
| `store_name` | Legal/store name | Seller dashboard and admin review |
| `display_name` | Public display name | Product listing me seller name |
| `gst_number` | Tax/GST identifier | Seller compliance metadata |
| `support_email` | Seller support contact | Buyer support and seller operations |
| `status` | Seller lifecycle | Draft to active/suspended workflow |
| `approved_by` | Admin user id | Approval traceability |
| `approved_at` | Approval timestamp | Compliance and audit |

### Seller statuses

| Status | Meaning |
|---|---|
| `draft` | Seller profile incomplete hai |
| `pending_review` | Seller ne profile/KYC submit kar diya |
| `active` | Seller approved hai, products sell kar sakta hai |
| `suspended` | Seller temporarily blocked hai |
| `rejected` | Seller application reject hui |

### Why `UNIQUE KEY uk_seller_profiles_user_id`?

Ek user ke multiple seller profiles avoid karne ke liye. Marketplace MVP me one user = maximum one seller identity simple and safe rahega.

---

## Step 5: `seller_kyc_documents` Table Design Kiya

`seller_kyc_documents` seller verification documents ka metadata store karega.

### Purpose

Actual KYC document file object storage/CDN me rahegi. MySQL me document id, type, storage URL, review status, reviewer, aur rejection reason store hoga.

### SQL

```sql
CREATE TABLE IF NOT EXISTS seller_kyc_documents (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  document_id VARCHAR(64) NOT NULL,
  seller_id VARCHAR(64) NOT NULL,
  document_type VARCHAR(64) NOT NULL,
  storage_url VARCHAR(1024) NOT NULL,
  status ENUM('pending', 'approved', 'rejected') NOT NULL DEFAULT 'pending',
  reviewed_by VARCHAR(64) NULL,
  reviewed_at TIMESTAMP NULL,
  rejection_reason VARCHAR(512) NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_kyc_document_id (document_id),
  KEY idx_kyc_seller_status (seller_id, status),
  CONSTRAINT fk_kyc_seller FOREIGN KEY (seller_id) REFERENCES seller_profiles(seller_id)
) ENGINE=InnoDB;
```

### Field explanation

| Field | Meaning | Why needed |
|---|---|---|
| `document_id` | Stable KYC document id | Review/update APIs me use hoga |
| `seller_id` | Seller profile relation | KYC belongs to seller |
| `document_type` | PAN/GST/license/etc. | Verification category |
| `storage_url` | Object storage/CDN URL | Actual file DB me store nahi hogi |
| `status` | Document review state | Admin review workflow |
| `reviewed_by` | Admin/reviewer id | Audit and compliance |
| `reviewed_at` | Review timestamp | Review history |
| `rejection_reason` | Rejection explanation | Seller ko correction reason dikhane ke liye |
| `created_at` | Upload timestamp | Review ordering |

### KYC document statuses

| Status | Meaning |
|---|---|
| `pending` | Document uploaded, review baaki hai |
| `approved` | Document verified hai |
| `rejected` | Document invalid/unclear hai |

> 🟡 **Important:** KYC files sensitive hoti hain. DB me binary content mat rakho. `storage_url` bhi private/signed URL ya internal object key ho sakta hai, production security policy ke according.

---

## Step 6: Full Up Migration Example

Future file:

```text
backend/services/user-service/migrations/001_create_user_tables.up.sql
```

```sql
CREATE DATABASE IF NOT EXISTS user_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE user_db;

CREATE TABLE IF NOT EXISTS users (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id VARCHAR(64) NOT NULL,
  auth_account_id VARCHAR(64) NOT NULL,
  email VARCHAR(255) NOT NULL,
  phone VARCHAR(32) NULL,
  full_name VARCHAR(255) NOT NULL,
  avatar_url VARCHAR(1024) NULL,
  status ENUM('active', 'blocked', 'deleted') NOT NULL DEFAULT 'active',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_users_user_id (user_id),
  UNIQUE KEY uk_users_auth_account_id (auth_account_id),
  UNIQUE KEY uk_users_email (email),
  KEY idx_users_status (status)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS user_addresses (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  address_id VARCHAR(64) NOT NULL,
  user_id VARCHAR(64) NOT NULL,
  name VARCHAR(255) NOT NULL,
  phone VARCHAR(32) NULL,
  line1 VARCHAR(255) NOT NULL,
  line2 VARCHAR(255) NULL,
  city VARCHAR(128) NOT NULL,
  state VARCHAR(128) NOT NULL,
  postal_code VARCHAR(32) NOT NULL,
  country VARCHAR(64) NOT NULL,
  is_default BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_user_addresses_address_id (address_id),
  KEY idx_user_addresses_user_default (user_id, is_default),
  CONSTRAINT fk_user_addresses_user FOREIGN KEY (user_id) REFERENCES users(user_id)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS seller_profiles (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  seller_id VARCHAR(64) NOT NULL,
  user_id VARCHAR(64) NOT NULL,
  store_name VARCHAR(255) NOT NULL,
  display_name VARCHAR(255) NULL,
  gst_number VARCHAR(64) NULL,
  support_email VARCHAR(255) NULL,
  status ENUM('draft', 'pending_review', 'active', 'suspended', 'rejected') NOT NULL DEFAULT 'draft',
  approved_by VARCHAR(64) NULL,
  approved_at TIMESTAMP NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_seller_profiles_seller_id (seller_id),
  UNIQUE KEY uk_seller_profiles_user_id (user_id),
  KEY idx_seller_profiles_status (status),
  CONSTRAINT fk_seller_profiles_user FOREIGN KEY (user_id) REFERENCES users(user_id)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS seller_kyc_documents (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  document_id VARCHAR(64) NOT NULL,
  seller_id VARCHAR(64) NOT NULL,
  document_type VARCHAR(64) NOT NULL,
  storage_url VARCHAR(1024) NOT NULL,
  status ENUM('pending', 'approved', 'rejected') NOT NULL DEFAULT 'pending',
  reviewed_by VARCHAR(64) NULL,
  reviewed_at TIMESTAMP NULL,
  rejection_reason VARCHAR(512) NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_kyc_document_id (document_id),
  KEY idx_kyc_seller_status (seller_id, status),
  CONSTRAINT fk_kyc_seller FOREIGN KEY (seller_id) REFERENCES seller_profiles(seller_id)
) ENGINE=InnoDB;
```

### Migration order

| Order | Table | Why this order |
|---:|---|---|
| 1 | `users` | Parent table hai |
| 2 | `user_addresses` | `users(user_id)` FK use karta hai |
| 3 | `seller_profiles` | `users(user_id)` FK use karta hai |
| 4 | `seller_kyc_documents` | `seller_profiles(seller_id)` FK use karta hai |

---

## Step 7: Full Down Migration Example

Future file:

```text
backend/services/user-service/migrations/001_create_user_tables.down.sql
```

```sql
USE user_db;

DROP TABLE IF EXISTS seller_kyc_documents;
DROP TABLE IF EXISTS seller_profiles;
DROP TABLE IF EXISTS user_addresses;
DROP TABLE IF EXISTS users;
```

### Why reverse order?

Foreign key dependencies child → parent hoti hain. Rollback me pehle child tables drop karo, phir parent table. Agar `users` pehle drop karoge to `user_addresses` and `seller_profiles` FK error de sakte hain.

---

## Step 8: Relationship and Flow Samjha

### Signup to user profile creation

```mermaid
sequenceDiagram
    participant Client as User App
    participant GW as API Gateway
    participant Auth as Auth Service
    participant User as User Service
    participant DB as MySQL user_db

    Client->>GW: POST /api/v1/auth/signup
    GW->>Auth: Signup request
    Auth->>Auth: Create auth account and credentials
    Auth->>User: CreateUser(auth_account_id, email, phone, full_name)
    User->>DB: INSERT INTO users
    DB-->>User: user_id created
    User-->>Auth: User profile response
    Auth-->>GW: Auth session response
    GW-->>Client: access_token + user profile
```

### Seller onboarding and KYC metadata

```mermaid
sequenceDiagram
    participant Seller as Seller Dashboard
    participant GW as API Gateway
    participant User as User Service
    participant DB as MySQL user_db
    participant Admin as Superadmin Service

    Seller->>GW: Submit seller profile + KYC metadata
    GW->>User: UpdateSellerProfile / SubmitKYC
    User->>DB: INSERT seller_profiles
    User->>DB: INSERT seller_kyc_documents
    User-->>Admin: Review task event/API handoff
    Admin-->>User: Approve or reject seller
    User->>DB: UPDATE seller_profiles.status
```

### Hinglish explanation

Signup Auth se start hota hai, kyunki credentials Auth Service own karta hai. Auth account create hone ke baad User Service me profile row create hoti hai. Seller onboarding me User Service seller profile and KYC metadata store karta hai, lekin admin review Superadmin workflow se drive hota hai.

---

## Step 9: Index Strategy Define Kiya

| Index | Table | Query use case |
|---|---|---|
| `uk_users_user_id` | `users` | `GetUser(user_id)` |
| `uk_users_auth_account_id` | `users` | Auth signup/link lookup |
| `uk_users_email` | `users` | Email-based profile lookup/admin search |
| `idx_users_status` | `users` | Active/blocked/deleted user filters |
| `uk_user_addresses_address_id` | `user_addresses` | Address update/delete by id |
| `idx_user_addresses_user_default` | `user_addresses` | User address list and default address lookup |
| `uk_seller_profiles_seller_id` | `seller_profiles` | Product/CMS/Order seller lookup |
| `uk_seller_profiles_user_id` | `seller_profiles` | Current user's seller profile lookup |
| `idx_seller_profiles_status` | `seller_profiles` | Admin review queue/filter |
| `uk_kyc_document_id` | `seller_kyc_documents` | KYC document detail/update |
| `idx_kyc_seller_status` | `seller_kyc_documents` | Seller KYC review list |

### Example fast queries

```sql
-- Get user profile
SELECT user_id, email, phone, full_name, avatar_url, status, created_at, updated_at
FROM users
WHERE user_id = 'user_123';

-- Get active addresses for checkout/profile page
SELECT address_id, name, phone, line1, line2, city, state, postal_code, country, is_default
FROM user_addresses
WHERE user_id = 'user_123'
  AND deleted_at IS NULL
ORDER BY is_default DESC, updated_at DESC;

-- Get seller profile for seller dashboard
SELECT seller_id, user_id, store_name, display_name, gst_number, support_email, status
FROM seller_profiles
WHERE user_id = 'user_123';

-- Admin review queue
SELECT seller_id, user_id, store_name, gst_number, status, created_at
FROM seller_profiles
WHERE status = 'pending_review'
ORDER BY created_at ASC;
```

---

## Step 10: Data Integrity Rules Define Kiye

### Schema-level rules

| Rule | Enforced by |
|---|---|
| Every user has unique `user_id` | Unique key |
| Every Auth account maps to one User profile | Unique key on `auth_account_id` |
| Address cannot exist without user | Foreign key |
| Seller profile cannot exist without user | Foreign key |
| One user can have only one seller profile | Unique key on `seller_profiles.user_id` |
| KYC document cannot exist without seller | Foreign key |
| Seller/KYC status must be known value | MySQL `ENUM` |

### Usecase-level rules

| Rule | Where to enforce later |
|---|---|
| Max 20 active addresses per user | Repository/usecase transaction |
| Only one active default address | Repository/usecase transaction |
| Seller can move from `draft` to `pending_review` only after required KYC | Usecase validation |
| `approved_by` and `approved_at` required when seller becomes `active` | Usecase validation |
| Rejection reason required when document is `rejected` | Usecase validation |
| User status `blocked/deleted` should restrict profile updates | Usecase authorization |

> 🟢 **Reason:** Har business rule DB constraint me fit nahi hota. Simple uniqueness/FK DB me, workflow rules usecase layer me rakho.

---

## Step 11: External Tools/Libraries

Task 2 me actual external library install nahi ki gayi. Lekin future schema implementation ke liye ye tools recommended hain:

| Tool/Library | What it is | Why used | Install | Basic use |
|---|---|---|---|---|
| MySQL 8.x | Relational database | Strong consistency, FK, indexes, transactions | Docker ya native installer | `mysql -u root -p` |
| golang-migrate/migrate | DB migration CLI/library | Versioned up/down migrations run karne ke liye | `brew install golang-migrate` or download binary | `migrate -path migrations -database "$USER_DB_DSN" up` |
| Docker Compose | Local dependency runner | MySQL local easily start karne ke liye | Docker Desktop install | `docker compose up -d mysql` |
| MySQL client | CLI for SQL verification | Tables/indexes/query check karne ke liye | Comes with MySQL package | `SHOW TABLES;` |

### Example: MySQL via Docker

```bash
docker run --name ecommerce-user-mysql \
  -e MYSQL_ROOT_PASSWORD=root \
  -e MYSQL_DATABASE=user_db \
  -p 3307:3306 \
  -d mysql:8
```

### Example: Migration command

```bash
migrate \
  -path backend/services/user-service/migrations \
  -database "mysql://root:root@tcp(localhost:3307)/user_db?multiStatements=true" \
  up
```

### Example: Rollback command

```bash
migrate \
  -path backend/services/user-service/migrations \
  -database "mysql://root:root@tcp(localhost:3307)/user_db?multiStatements=true" \
  down 1
```

> 🟡 **Note:** Project docs mention Docker Compose local stack separately in Platform Foundation. Is task me Docker Compose file create nahi ki gayi.

---

## Step 12: Verification Queries

Migration run hone ke baad ye sanity checks useful honge.

### Check database and tables

```sql
SHOW DATABASES LIKE 'user_db';
USE user_db;
SHOW TABLES;
```

Expected tables:

```text
users
user_addresses
seller_profiles
seller_kyc_documents
```

### Check indexes

```sql
SHOW INDEX FROM users;
SHOW INDEX FROM user_addresses;
SHOW INDEX FROM seller_profiles;
SHOW INDEX FROM seller_kyc_documents;
```

### Check foreign keys

```sql
SELECT
  table_name,
  constraint_name,
  referenced_table_name
FROM information_schema.key_column_usage
WHERE table_schema = 'user_db'
  AND referenced_table_name IS NOT NULL;
```

Expected FK mapping:

| Table | References |
|---|---|
| `user_addresses` | `users` |
| `seller_profiles` | `users` |
| `seller_kyc_documents` | `seller_profiles` |

### Insert smoke test data

```sql
INSERT INTO users (
  user_id,
  auth_account_id,
  email,
  phone,
  full_name
) VALUES (
  'user_123',
  'auth_123',
  'buyer@example.com',
  '+919999999999',
  'Test Buyer'
);

INSERT INTO user_addresses (
  address_id,
  user_id,
  name,
  phone,
  line1,
  city,
  state,
  postal_code,
  country,
  is_default
) VALUES (
  'addr_123',
  'user_123',
  'Test Buyer',
  '+919999999999',
  '221B Demo Street',
  'Bengaluru',
  'Karnataka',
  '560001',
  'IN',
  TRUE
);

INSERT INTO seller_profiles (
  seller_id,
  user_id,
  store_name,
  display_name,
  gst_number,
  support_email,
  status
) VALUES (
  'seller_123',
  'user_123',
  'Demo Store Pvt Ltd',
  'Demo Store',
  '29ABCDE1234F1Z5',
  'support@demo-store.example',
  'pending_review'
);

INSERT INTO seller_kyc_documents (
  document_id,
  seller_id,
  document_type,
  storage_url
) VALUES (
  'doc_123',
  'seller_123',
  'gst_certificate',
  's3://private-kyc-bucket/seller_123/gst_certificate.pdf'
);
```

### Cleanup smoke test data

```sql
DELETE FROM seller_kyc_documents WHERE document_id = 'doc_123';
DELETE FROM seller_profiles WHERE seller_id = 'seller_123';
DELETE FROM user_addresses WHERE address_id = 'addr_123';
DELETE FROM users WHERE user_id = 'user_123';
```

---

## Step 13: Security and Privacy Notes

| Data | Risk | Schema decision |
|---|---|---|
| Password | Very sensitive | Not stored in User DB |
| OTP | Very sensitive and short-lived | Not stored in User DB |
| Refresh token | Sensitive auth secret | Not stored in User DB |
| KYC file | Sensitive document | Only `storage_url` metadata stored |
| Email/phone | PII | Stored as profile mirror, protect via service auth and logs masking |
| Address | PII | Soft delete plus access control required |
| Seller GST | Business-sensitive | Limit access to seller/admin workflows |

### Logging caution

Future repository/usecase code me ye fields logs me plain text me dump nahi karne:

- `email`
- `phone`
- address lines
- `postal_code`
- `storage_url`
- `gst_number`
- `rejection_reason` if it includes sensitive details

> 🟢 **Best practice:** Logs me `user_id`, `seller_id`, `document_id`, aur request id enough hote hain. PII ko mask karo.

---

## Step 14: Common Mistakes Avoid Kiye

| Mistake | Why problematic | Better approach |
|---|---|---|
| Password ko `users` table me rakhna | Auth/User boundary break hoti hai | Password Auth Service ke `credentials` me rahe |
| Address hard-delete karna | Old order/debug context lose hota hai | `deleted_at` soft delete use karo |
| KYC binary DB me store karna | DB bloat and security risk | Object storage URL/key store karo |
| Multiple seller profiles per user allow karna | MVP complexity badhti hai | `seller_profiles.user_id` unique rakho |
| No indexes on status columns | Admin review/list slow ho sakti hai | `idx_users_status`, `idx_seller_profiles_status` |
| Cross-service FK to Auth DB banana | Microservice DB ownership break hoti hai | `auth_account_id` as reference string only |

### Important microservice rule

User DB me `auth_account_id` store hoga, lekin MySQL foreign key Auth DB ke `auth_accounts` table par nahi banega.

**Reason:** Har microservice apni database own karegi. Cross-service DB foreign keys independent deployments and scaling ko tightly couple kar dete hain.

---

## Step 15: Future Repository Mapping

Task 3 me repository layer banegi. Ye schema uske liye base hoga.

### Expected repository methods

```go
type UserRepository interface {
    CreateUser(ctx context.Context, user User) error
    GetUserByID(ctx context.Context, userID string) (*User, error)
    GetUserByAuthAccountID(ctx context.Context, authAccountID string) (*User, error)
    UpdateUserProfile(ctx context.Context, user User) error
    UpdateUserStatus(ctx context.Context, userID string, status UserStatus) error

    ListAddresses(ctx context.Context, userID string) ([]Address, error)
    CreateAddress(ctx context.Context, address Address) error
    UpdateAddress(ctx context.Context, address Address) error
    SoftDeleteAddress(ctx context.Context, userID, addressID string) error
    SetDefaultAddress(ctx context.Context, userID, addressID string) error

    GetSellerProfileByUserID(ctx context.Context, userID string) (*SellerProfile, error)
    GetSellerProfileBySellerID(ctx context.Context, sellerID string) (*SellerProfile, error)
    UpsertSellerProfile(ctx context.Context, seller SellerProfile) error
    ListKYCDocuments(ctx context.Context, sellerID string) ([]KYCDocument, error)
}
```

> 🟡 **Reminder:** Ye code sirf future mapping example hai. Task 2 me actual Go repository file create nahi ki gayi.

---

## ✅ Acceptance Checklist

| Check | Status |
|---|---|
| `TaskImplementation/User Service/` folder exists | ✅ Done |
| `task2.md` created inside User Service folder | ✅ Done |
| Task 2 scope limited to MySQL schema documentation | ✅ Done |
| `users` table documented | ✅ Done |
| `user_addresses` table documented | ✅ Done |
| `seller_profiles` table documented | ✅ Done |
| `seller_kyc_documents` table documented | ✅ Done |
| Relationships and indexes explained | ✅ Done |
| Up/down migration examples included | ✅ Done |
| Mermaid ER and flow diagrams included | ✅ Done |
| External tools/libraries explained | ✅ Done |
| Hinglish beginner-friendly explanation included | ✅ Done |
| No backend service implementation created | ✅ Done |

---

## 🧾 Final Notes

User Service Task 2 ka output ek clean MySQL schema guide hai. Ye schema User domain ko relational form me convert karta hai:

- `users` base profile ke liye
- `user_addresses` address book ke liye
- `seller_profiles` seller onboarding/profile ke liye
- `seller_kyc_documents` KYC metadata ke liye

Is design me service boundary clean rakhi gayi hai: Auth secrets Auth Service me, profile data User Service me, actual document files object storage me, aur Superadmin approval workflow alag service se coordinate hoga.
