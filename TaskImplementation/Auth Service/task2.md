# 🔐 Auth Service - Task 2: Create MySQL Schema

![Task](https://img.shields.io/badge/Task-Auth%20Service%20%232-blue)
![Priority](https://img.shields.io/badge/Priority-P0-red)
![Dependency](https://img.shields.io/badge/Dependency-Auth%20Flows-yellow)
![Status](https://img.shields.io/badge/Status-Documented-success)
![Guide](https://img.shields.io/badge/Guide-Hinglish-orange)
![Scope](https://img.shields.io/badge/Scope-MySQL%20Schema-purple)

## 📌 Task Summary

| Field | Detail |
|---|---|
| Task Name | Create MySQL schema |
| Source | `docs/01-micro-tasks.md` → `Auth Service` → Task 2 |
| Priority | `P0` security foundation |
| Dependency | Auth Service Task 1: Define auth flows |
| Main Goal | `auth_accounts`, `credentials`, `refresh_tokens`, `otp_challenges`, aur `role_assignments` tables banana |
| Core Boundary | Auth Service identity, credentials, tokens, OTP challenges, aur roles own karega |
| Output Type | Structured implementation guide |
| Not Included | Password hashing implementation, JWT signing, OTP delivery, gRPC server, REST handlers, RBAC middleware, Session Service integration |

> **Simple Hinglish goal:** Is task ka purpose Auth Service ke liye strong-consistency MySQL schema define karna hai. Auth flows Task 1 me clear ho chuke hain; ab un flows ko support karne ke liye tables, relationships, indexes, constraints, migration order, rollback, aur verification queries document kiye gaye hain.

---

## ✅ Final Output Created

```text
TaskImplementation/
├── Auth Service/
│   ├── task1.md
│   └── task2.md
├── Platform Foundation/
│   ├── task1.md
│   ├── task2.md
│   ├── task3.md
│   ├── task4.md
│   ├── task5.md
│   ├── task6.md
│   ├── task7.md
│   └── task8.md
└── User Service/
    └── task1.md
```

### Why this structure?

- `TaskImplementation/` project ke task-wise implementation guides ka central folder hai.
- `Auth Service/` folder already present tha, isliye usko keep kiya gaya.
- `task2.md` sirf **Auth Service - Task 2** ka guide hai.
- Existing `task1.md`, Platform Foundation guides, aur User Service guide untouched rakhe gaye.
- Actual `backend/services/auth-service/migrations/` files create nahi ki gayi, kyunki requested output sirf folder structure aur `task2.md` content generate karna hai.

---

## 🧭 Implementation Approach

Is guide ko banate time project ke official docs ko source of truth maana gaya:

| Document | Kya use kiya gaya |
|---|---|
| `docs/01-micro-tasks.md` | Auth Service Task 2 ka exact scope: MySQL schema |
| `TaskImplementation/Auth Service/task1.md` | Auth flows, ownership boundary, future table mapping |
| `docs/03-folder-structure.md` | Future `backend/services/auth-service/migrations/` folder convention |
| `docs/04-microservice-design.md` | Auth Service responsibilities, DB choice, table names |
| `docs/05-database-design.md` | Auth DB relationships, indexes, scalability notes |
| `database/draw.sql` | Existing detailed Auth Service DDL source |
| `docs/06-auth-security.md` | JWT, refresh token, OTP, RBAC, rate limit, secrets rules |
| `docs/08-session-management-system.md` | Refresh/session relationship and security touchpoints |

---

## 🧱 Task Boundary

### ✅ Included in Task 2

- Auth Service MySQL database schema design
- `auth_db` database creation strategy
- `auth_accounts` table design
- `credentials` table design
- `refresh_tokens` table design
- `otp_challenges` table design
- `role_assignments` table design
- Primary keys, unique keys, foreign keys, and indexes
- Migration up/down SQL examples
- Data relationship diagrams
- Beginner-friendly explanation of each table
- External tools/libraries explanation
- Verification queries and sanity checks
- Security checklist for schema-level sensitive data

### 🚫 Not Included in Task 2

- Password hashing code
- JWT signing/verification code
- Refresh token generation code
- OTP generation and SMS/email delivery code
- Redis rate limiter implementation
- Repository layer implementation
- gRPC/REST handlers
- Session Service event publishing
- Auth Service tests
- Production secret setup

> 🟢 **Rule:** Task 2 ka scope database foundation hai. Password hashing Task 3, JWT issuing Task 4, OTP verification Task 5, RBAC Task 6, Session link Task 7, aur security tests Task 8 me aayenge.

---

## 🗂️ Target Implementation Folder Structure

Future me actual Auth Service implementation ka schema part yahan rahega:

```text
backend/
└── services/
    └── auth-service/
        ├── cmd/
        │   └── server/
        │       └── main.go
        ├── internal/
        │   ├── domain/
        │   │   ├── account.go
        │   │   ├── credential.go
        │   │   ├── otp.go
        │   │   ├── token.go
        │   │   └── role.go
        │   ├── repository/
        │   │   ├── mysql_account_repository.go
        │   │   ├── mysql_credential_repository.go
        │   │   ├── mysql_token_repository.go
        │   │   ├── mysql_otp_repository.go
        │   │   └── mysql_role_repository.go
        │   └── config/
        │       └── config.go
        └── migrations/
            ├── 001_create_auth_tables.up.sql
            └── 001_create_auth_tables.down.sql
```

### Folder responsibility

| Path | Responsibility |
|---|---|
| `migrations/001_create_auth_tables.up.sql` | Auth DB tables create karega |
| `migrations/001_create_auth_tables.down.sql` | Rollback ke time tables drop karega |
| `internal/domain/` | Table-backed business entities define karega |
| `internal/repository/` | MySQL queries and transactions implement karega |
| `internal/config/` | MySQL DSN and migration config load karega |

> 🟡 **Important:** Ye target structure hai. Is Task 2 me sirf `TaskImplementation/Auth Service/task2.md` create kiya gaya.

---

## 🧩 Schema Architecture

```mermaid
erDiagram
    AUTH_ACCOUNTS ||--|| CREDENTIALS : has
    AUTH_ACCOUNTS ||--o{ REFRESH_TOKENS : owns
    AUTH_ACCOUNTS ||--o{ ROLE_ASSIGNMENTS : has
    AUTH_ACCOUNTS ||--o{ OTP_CHALLENGES : may_reference

    AUTH_ACCOUNTS {
        bigint id PK
        string account_id UK
        string email UK
        string phone UK
        bool email_verified
        bool phone_verified
        enum status
        timestamp created_at
        timestamp updated_at
    }

    CREDENTIALS {
        bigint id PK
        string account_id UK,FK
        string password_hash
        string password_algo
        timestamp password_changed_at
        int failed_attempts
        timestamp locked_until
        timestamp created_at
        timestamp updated_at
    }

    REFRESH_TOKENS {
        bigint id PK
        string token_id UK
        string account_id FK
        string session_id
        string token_hash UK
        string parent_token_id
        timestamp revoked_at
        timestamp expires_at
        timestamp created_at
    }

    OTP_CHALLENGES {
        bigint id PK
        string challenge_id UK
        string account_id
        string target
        enum channel
        enum purpose
        string otp_hash
        int attempts
        int max_attempts
        timestamp verified_at
        timestamp expires_at
        timestamp created_at
    }

    ROLE_ASSIGNMENTS {
        bigint id PK
        string account_id FK
        string role
        string scope_type
        string scope_id
        string assigned_by
        timestamp assigned_at
        timestamp revoked_at
    }
```

### Hinglish Explanation

`auth_accounts` root table hai. Har account ke paas exactly one `credentials` row hogi. Login ke baad multiple `refresh_tokens` ho sakte hain, kyunki user multiple devices pe logged in ho sakta hai. OTP challenges signup, login, password reset, email verify, ya phone verify ke liye create honge. Roles `role_assignments` me rahenge taaki JWT claims and service-side authorization dono consistent source se aa sakein.

---

## 🪜 Step-by-Step Implementation Guide

## Step 1: Auth DB ownership lock kiya

Auth Service ka database separate hoga. Dusri service direct `auth_db` read/write nahi karegi.

| Rule | Reason |
|---|---|
| Auth Service owns `auth_db` | Credentials and tokens high-risk security data hain |
| User Service direct Auth DB access nahi karega | Service boundary clean rahegi |
| API Gateway Auth DB access nahi karega | Gateway only JWT/JWKS ya Auth gRPC use karega |
| Admin tools direct token table edit nahi karenge | Revocation controlled usecase se hoga |

### Database name

```sql
CREATE DATABASE IF NOT EXISTS auth_db
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci;
```

**Why `utf8mb4`?**  
Email, names, metadata, aur future auth provider fields me full Unicode support chahiye. MySQL me `utf8mb4` modern safe default hai.

---

## Step 2: Migration strategy decide kiya

Auth schema ko migration file me versioned rakhna chahiye:

```text
backend/services/auth-service/migrations/
├── 001_create_auth_tables.up.sql
└── 001_create_auth_tables.down.sql
```

### Migration order

```mermaid
flowchart TD
    A[Start migration] --> B[Create auth_db]
    B --> C[Create auth_accounts]
    C --> D[Create credentials]
    D --> E[Create refresh_tokens]
    E --> F[Create otp_challenges]
    F --> G[Create role_assignments]
    G --> H[Verify indexes and foreign keys]
    H --> I[Done]
```

### Why this order?

- `auth_accounts` parent table hai.
- `credentials`, `refresh_tokens`, and `role_assignments` `auth_accounts(account_id)` ko reference karte hain.
- Parent table pehle create nahi hui to foreign key fail ho jayegi.
- `otp_challenges.account_id` nullable hai, kyunki signup/password reset jaise flows me account absent ya optional ho sakta hai.

---

## Step 3: `auth_accounts` table banaya

`auth_accounts` login identity and account status ka source hai.

### Table purpose

| Field | Purpose |
|---|---|
| `id` | Internal numeric primary key |
| `account_id` | Public/internal stable auth id, e.g. `auth_123` |
| `email` | Email login identity |
| `phone` | Phone login identity |
| `email_verified` | Email ownership verified hai ya nahi |
| `phone_verified` | Phone ownership verified hai ya nahi |
| `status` | `active`, `blocked`, ya `deleted` |
| `created_at`, `updated_at` | Audit timestamps |

### SQL

```sql
CREATE TABLE IF NOT EXISTS auth_accounts (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  account_id VARCHAR(64) NOT NULL,
  email VARCHAR(255) NULL,
  phone VARCHAR(32) NULL,
  email_verified BOOLEAN NOT NULL DEFAULT FALSE,
  phone_verified BOOLEAN NOT NULL DEFAULT FALSE,
  status ENUM('active', 'blocked', 'deleted') NOT NULL DEFAULT 'active',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_auth_accounts_account_id (account_id),
  UNIQUE KEY uk_auth_accounts_email (email),
  UNIQUE KEY uk_auth_accounts_phone (phone),
  KEY idx_auth_accounts_status (status)
) ENGINE=InnoDB;
```

### Hinglish Explanation

`account_id` ko stable external reference banaya gaya hai, numeric `id` ko internal DB key rakha gaya hai. Email and phone unique hain, taaki same identity se duplicate accounts na ban sakein. `email` and `phone` nullable hain because future me phone-only ya email-only signup support possible hai.

### Security note

```text
Email/phone PII hain. Table me store honge, but logs/events me masked ya hashed form prefer karo.
```

---

## Step 4: `credentials` table banaya

`credentials` password hash and lock metadata rakhega.

### Table purpose

| Field | Purpose |
|---|---|
| `account_id` | One credential row per auth account |
| `password_hash` | Argon2id/bcrypt generated hash |
| `password_algo` | Hash algorithm name |
| `password_changed_at` | Password rotation/reset tracking |
| `failed_attempts` | Login failures count |
| `locked_until` | Temporary lockout expiry |

### SQL

```sql
CREATE TABLE IF NOT EXISTS credentials (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  account_id VARCHAR(64) NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  password_algo VARCHAR(32) NOT NULL DEFAULT 'argon2id',
  password_changed_at TIMESTAMP NULL,
  failed_attempts INT NOT NULL DEFAULT 0,
  locked_until TIMESTAMP NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_credentials_account_id (account_id),
  CONSTRAINT fk_credentials_account
    FOREIGN KEY (account_id) REFERENCES auth_accounts(account_id)
) ENGINE=InnoDB;
```

### Hinglish Explanation

Password kabhi plain text store nahi hota. `password_hash` me only hashed value store hogi. `password_algo` future migrations ke liye useful hai, jaise bcrypt se Argon2id upgrade. `failed_attempts` and `locked_until` brute-force protection me help karte hain.

### Important rule

```text
Plain password:
- database me nahi
- logs me nahi
- events me nahi
- error messages me nahi
```

---

## Step 5: `refresh_tokens` table banaya

`refresh_tokens` rotating refresh token lifecycle track karega.

### Table purpose

| Field | Purpose |
|---|---|
| `token_id` | Stable refresh token id |
| `account_id` | Token owner |
| `session_id` | Device/browser session id |
| `token_hash` | Refresh token ka hash, plain token nahi |
| `parent_token_id` | Rotation chain tracking |
| `revoked_at` | Logout/reuse/password reset revocation |
| `expires_at` | Token expiry |

### SQL

```sql
CREATE TABLE IF NOT EXISTS refresh_tokens (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  token_id VARCHAR(64) NOT NULL,
  account_id VARCHAR(64) NOT NULL,
  session_id VARCHAR(64) NOT NULL,
  token_hash CHAR(64) NOT NULL,
  parent_token_id VARCHAR(64) NULL,
  revoked_at TIMESTAMP NULL,
  expires_at TIMESTAMP NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_refresh_tokens_token_id (token_id),
  UNIQUE KEY uk_refresh_tokens_hash (token_hash),
  KEY idx_refresh_tokens_account_session (account_id, session_id),
  KEY idx_refresh_tokens_expires_at (expires_at),
  CONSTRAINT fk_refresh_tokens_account
    FOREIGN KEY (account_id) REFERENCES auth_accounts(account_id)
) ENGINE=InnoDB;
```

### Refresh rotation flow

```mermaid
sequenceDiagram
    participant Client
    participant Auth as Auth Service
    participant DB as MySQL auth_db

    Client->>Auth: POST /auth/refresh with refresh_token
    Auth->>Auth: Hash incoming refresh token
    Auth->>DB: Find refresh_tokens by token_hash
    DB-->>Auth: Active token row
    Auth->>DB: Set old token revoked_at
    Auth->>DB: Insert new refresh token with parent_token_id
    Auth-->>Client: New access token + new refresh token
```

### Hinglish Explanation

Refresh token long-lived secret hai, isliye plain token DB me store nahi hoga. Sirf `token_hash` store hoga. Har refresh request pe old token revoke hoga and new token insert hoga. Agar old token dobara use hua, to possible theft signal hai.

### Hashing detail

```text
Recommended token hash:
SHA-256 or HMAC-SHA-256 hex string = 64 chars.

Plain refresh token user/client ko milega.
DB me only token_hash jayega.
```

---

## Step 6: `otp_challenges` table banaya

`otp_challenges` email/phone verification and password reset challenges track karega.

### Table purpose

| Field | Purpose |
|---|---|
| `challenge_id` | Client-facing challenge reference |
| `account_id` | Optional account link |
| `target` | Email or phone destination |
| `channel` | `email` or `phone` |
| `purpose` | `signup`, `login`, `password_reset`, `phone_verify`, `email_verify` |
| `otp_hash` | OTP hash, plain OTP nahi |
| `attempts` | Verify attempts count |
| `max_attempts` | Max allowed attempts |
| `verified_at` | Successful verification timestamp |
| `expires_at` | OTP expiry |

### SQL

```sql
CREATE TABLE IF NOT EXISTS otp_challenges (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  challenge_id VARCHAR(64) NOT NULL,
  account_id VARCHAR(64) NULL,
  target VARCHAR(255) NOT NULL,
  channel ENUM('email', 'phone') NOT NULL,
  purpose ENUM('signup', 'login', 'password_reset', 'phone_verify', 'email_verify') NOT NULL,
  otp_hash CHAR(64) NOT NULL,
  attempts INT NOT NULL DEFAULT 0,
  max_attempts INT NOT NULL DEFAULT 5,
  verified_at TIMESTAMP NULL,
  expires_at TIMESTAMP NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_otp_challenges_challenge_id (challenge_id),
  KEY idx_otp_target_purpose_expiry (target, purpose, expires_at),
  KEY idx_otp_account (account_id)
) ENGINE=InnoDB;
```

### OTP verification flow

```mermaid
sequenceDiagram
    participant Client
    participant Auth as Auth Service
    participant DB as MySQL auth_db

    Client->>Auth: VerifyOTP(challenge_id, otp)
    Auth->>DB: Load challenge by challenge_id
    Auth->>Auth: Check expires_at, verified_at, attempts
    Auth->>Auth: Hash submitted OTP
    Auth->>DB: Update attempts or verified_at
    Auth-->>Client: Success or generic failure
```

### Hinglish Explanation

OTP short-lived proof hai. User ko OTP dikhega, but DB me plain OTP kabhi nahi jayega. `attempts` and `max_attempts` brute-force prevent karte hain. `expires_at` normally 5 minutes ke around hoga, as security docs recommend.

### Why `account_id` nullable?

| Flow | Account known? | Reason |
|---|---:|---|
| Signup email verify | Sometimes no | Account create se pehle challenge ho sakta hai |
| Password reset | Maybe yes | Enumeration avoid karne ke liye generic handling |
| Phone verify | Yes | Existing account phone verify kar raha hai |
| Login step-up | Yes | Existing account extra verification |

---

## Step 7: `role_assignments` table banaya

`role_assignments` RBAC source of truth hoga.

### Table purpose

| Field | Purpose |
|---|---|
| `account_id` | Role owner |
| `role` | `buyer`, `seller`, `admin`, `superadmin`, etc. |
| `scope_type` | Optional role scope, e.g. `seller`, `tenant`, `global` |
| `scope_id` | Scope ka id, e.g. seller id |
| `assigned_by` | Actor account/admin id |
| `assigned_at` | Role grant timestamp |
| `revoked_at` | Soft revoke timestamp |

### SQL

```sql
CREATE TABLE IF NOT EXISTS role_assignments (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  account_id VARCHAR(64) NOT NULL,
  role VARCHAR(64) NOT NULL,
  scope_type VARCHAR(64) NULL,
  scope_id VARCHAR(64) NULL,
  assigned_by VARCHAR(64) NULL,
  assigned_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  revoked_at TIMESTAMP NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_role_active (account_id, role, scope_type, scope_id, revoked_at),
  KEY idx_role_assignments_account (account_id),
  KEY idx_role_assignments_role (role),
  CONSTRAINT fk_role_assignments_account
    FOREIGN KEY (account_id) REFERENCES auth_accounts(account_id)
) ENGINE=InnoDB;
```

### Supported role examples

| Role | Use |
|---|---|
| `buyer` | Default customer access |
| `seller` | Seller owner |
| `seller_manager` | Seller team manager |
| `seller_catalog_editor` | Product/catalog access |
| `seller_order_manager` | Seller order operations |
| `admin` | Internal admin |
| `finance_admin` | Refund/payment review |
| `catalog_admin` | Catalog moderation |
| `operations_admin` | Order/support operations |
| `superadmin` | Platform-wide high-risk actions |

### Hinglish Explanation

Role ko account se link kiya gaya hai, user profile se nahi. Reason simple hai: roles Auth boundary ka part hain. JWT claims roles se banenge, and service-level permission checks bhi same source se verify kar sakte hain.

### 🟡 MySQL uniqueness caution

MySQL me `UNIQUE` indexes nullable columns ke saath multiple `NULL` values allow karte hain. Isliye `uk_role_active(account_id, role, scope_type, scope_id, revoked_at)` duplicate active global roles ko fully prevent nahi karega agar nullable fields `NULL` hain.

Beginner-friendly solution options:

| Option | Use |
|---|---|
| Application transaction check | Role assign karte time active duplicate query karke reject karo |
| Generated normalized columns | `NULL` ko sentinel value me convert karke unique key banao |
| Non-null default scope | `scope_type='global'`, `scope_id='global'` store karo |

Task 2 base schema `database/draw.sql` ke saath aligned hai. Production implementation me role duplicate prevention ko repository/usecase layer me explicitly enforce karna recommended hai.

---

## Step 8: Complete `up` migration prepare ki

Recommended file:

```text
backend/services/auth-service/migrations/001_create_auth_tables.up.sql
```

### Full up migration

```sql
-- Auth Service Database
-- Target: MySQL 8+
-- Rule: Auth Service owns this schema. Other services must use gRPC/events.

SET NAMES utf8mb4;
SET time_zone = '+00:00';

CREATE DATABASE IF NOT EXISTS auth_db
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci;

USE auth_db;

CREATE TABLE IF NOT EXISTS auth_accounts (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  account_id VARCHAR(64) NOT NULL,
  email VARCHAR(255) NULL,
  phone VARCHAR(32) NULL,
  email_verified BOOLEAN NOT NULL DEFAULT FALSE,
  phone_verified BOOLEAN NOT NULL DEFAULT FALSE,
  status ENUM('active', 'blocked', 'deleted') NOT NULL DEFAULT 'active',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_auth_accounts_account_id (account_id),
  UNIQUE KEY uk_auth_accounts_email (email),
  UNIQUE KEY uk_auth_accounts_phone (phone),
  KEY idx_auth_accounts_status (status)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS credentials (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  account_id VARCHAR(64) NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  password_algo VARCHAR(32) NOT NULL DEFAULT 'argon2id',
  password_changed_at TIMESTAMP NULL,
  failed_attempts INT NOT NULL DEFAULT 0,
  locked_until TIMESTAMP NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_credentials_account_id (account_id),
  CONSTRAINT fk_credentials_account
    FOREIGN KEY (account_id) REFERENCES auth_accounts(account_id)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS refresh_tokens (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  token_id VARCHAR(64) NOT NULL,
  account_id VARCHAR(64) NOT NULL,
  session_id VARCHAR(64) NOT NULL,
  token_hash CHAR(64) NOT NULL,
  parent_token_id VARCHAR(64) NULL,
  revoked_at TIMESTAMP NULL,
  expires_at TIMESTAMP NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_refresh_tokens_token_id (token_id),
  UNIQUE KEY uk_refresh_tokens_hash (token_hash),
  KEY idx_refresh_tokens_account_session (account_id, session_id),
  KEY idx_refresh_tokens_expires_at (expires_at),
  CONSTRAINT fk_refresh_tokens_account
    FOREIGN KEY (account_id) REFERENCES auth_accounts(account_id)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS otp_challenges (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  challenge_id VARCHAR(64) NOT NULL,
  account_id VARCHAR(64) NULL,
  target VARCHAR(255) NOT NULL,
  channel ENUM('email', 'phone') NOT NULL,
  purpose ENUM('signup', 'login', 'password_reset', 'phone_verify', 'email_verify') NOT NULL,
  otp_hash CHAR(64) NOT NULL,
  attempts INT NOT NULL DEFAULT 0,
  max_attempts INT NOT NULL DEFAULT 5,
  verified_at TIMESTAMP NULL,
  expires_at TIMESTAMP NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_otp_challenges_challenge_id (challenge_id),
  KEY idx_otp_target_purpose_expiry (target, purpose, expires_at),
  KEY idx_otp_account (account_id)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS role_assignments (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  account_id VARCHAR(64) NOT NULL,
  role VARCHAR(64) NOT NULL,
  scope_type VARCHAR(64) NULL,
  scope_id VARCHAR(64) NULL,
  assigned_by VARCHAR(64) NULL,
  assigned_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  revoked_at TIMESTAMP NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_role_active (account_id, role, scope_type, scope_id, revoked_at),
  KEY idx_role_assignments_account (account_id),
  KEY idx_role_assignments_role (role),
  CONSTRAINT fk_role_assignments_account
    FOREIGN KEY (account_id) REFERENCES auth_accounts(account_id)
) ENGINE=InnoDB;
```

---

## Step 9: Rollback `down` migration prepare ki

Recommended file:

```text
backend/services/auth-service/migrations/001_create_auth_tables.down.sql
```

### Full down migration

```sql
-- Rollback Auth Service tables.
-- Drop child tables before parent tables because of foreign keys.

USE auth_db;

DROP TABLE IF EXISTS role_assignments;
DROP TABLE IF EXISTS otp_challenges;
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS credentials;
DROP TABLE IF EXISTS auth_accounts;
```

### Why this order?

`credentials`, `refresh_tokens`, and `role_assignments` `auth_accounts` ko reference karte hain. Agar parent table pehle drop karoge to foreign key error aa sakta hai. Isliye child tables first drop honge.

---

## Step 10: Indexes and constraints verify kiye

### Index summary

| Table | Index | Why needed |
|---|---|---|
| `auth_accounts` | `uk_auth_accounts_account_id` | Stable account lookup |
| `auth_accounts` | `uk_auth_accounts_email` | Email duplicate signup block |
| `auth_accounts` | `uk_auth_accounts_phone` | Phone duplicate signup block |
| `auth_accounts` | `idx_auth_accounts_status` | Blocked/deleted account filtering |
| `credentials` | `uk_credentials_account_id` | One credential per account |
| `refresh_tokens` | `uk_refresh_tokens_token_id` | Rotation chain lookup |
| `refresh_tokens` | `uk_refresh_tokens_hash` | Incoming refresh token verification |
| `refresh_tokens` | `idx_refresh_tokens_account_session` | Logout current/all device flow |
| `refresh_tokens` | `idx_refresh_tokens_expires_at` | Cleanup expired tokens |
| `otp_challenges` | `uk_otp_challenges_challenge_id` | OTP verify by challenge id |
| `otp_challenges` | `idx_otp_target_purpose_expiry` | Resend/rate/cooldown checks |
| `role_assignments` | `idx_role_assignments_account` | JWT role load |
| `role_assignments` | `idx_role_assignments_role` | Admin role search/reporting |

### Constraint summary

| Constraint | Meaning |
|---|---|
| `credentials.account_id → auth_accounts.account_id` | Credential must belong to valid account |
| `refresh_tokens.account_id → auth_accounts.account_id` | Token must belong to valid account |
| `role_assignments.account_id → auth_accounts.account_id` | Role must belong to valid account |

### Why no cross-service foreign key?

User Service has `users.auth_account_id`, but Auth DB must not create FK to User DB. Microservice rule:

```text
No cross-service foreign keys.
Use gRPC or events for cross-service consistency.
```

---

## Step 11: External tools/libraries note kiya

Task 2 documentation ke liye repo me koi external package install nahi kiya gaya. Actual implementation me ye tools useful honge:

| Tool/Library | What it is | Why used | Install | Basic use |
|---|---|---|---|---|
| MySQL 8+ | Relational database | Auth credentials, tokens, roles me strong consistency chahiye | Docker Compose ya package manager | `mysql -h 127.0.0.1 -P 3306 -u root -p` |
| Docker Compose | Local dependency runner | MySQL local run karne ke liye repeatable setup | Docker Desktop / Docker Engine | `docker compose up -d mysql` |
| golang-migrate | DB migration CLI/library | Versioned `up/down` migrations run karne ke liye | `go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest` | `migrate -path migrations -database "$MYSQL_URL" up` |
| Go MySQL driver | Go `database/sql` driver | Auth repository se MySQL connect karne ke liye | `go get github.com/go-sql-driver/mysql` | import driver as `_ "github.com/go-sql-driver/mysql"` |

### MySQL local install option

```bash
docker compose -f infra/compose/docker-compose.local.yml up -d mysql
```

### golang-migrate install

```bash
go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### Migration run example

```bash
migrate \
  -path backend/services/auth-service/migrations \
  -database "mysql://root:password@tcp(127.0.0.1:3306)/auth_db?multiStatements=true" \
  up
```

### Go driver install

```bash
go get github.com/go-sql-driver/mysql
```

### Go connection example

```go
package main

import (
    "database/sql"
    "time"

    _ "github.com/go-sql-driver/mysql"
)

func openAuthDB(dsn string) (*sql.DB, error) {
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        return nil, err
    }

    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(10)
    db.SetConnMaxLifetime(30 * time.Minute)

    if err := db.Ping(); err != nil {
        return nil, err
    }

    return db, nil
}
```

> 🟡 **Note:** Go repository implementation Task 3+ me useful hoga. Task 2 me sirf schema guide create ki gayi hai.

---

## Step 12: Signup write transaction design kiya

Signup ke time account and credential ek transaction me create honge.

```mermaid
sequenceDiagram
    participant Auth as Auth Service
    participant DB as MySQL auth_db

    Auth->>DB: BEGIN
    Auth->>DB: INSERT auth_accounts
    Auth->>DB: INSERT credentials
    Auth->>DB: INSERT role_assignments buyer
    Auth->>DB: COMMIT
```

### SQL transaction example

```sql
START TRANSACTION;

INSERT INTO auth_accounts (
  account_id,
  email,
  phone,
  email_verified,
  phone_verified,
  status
) VALUES (
  'auth_123',
  'user@example.com',
  NULL,
  FALSE,
  FALSE,
  'active'
);

INSERT INTO credentials (
  account_id,
  password_hash,
  password_algo,
  password_changed_at
) VALUES (
  'auth_123',
  '$argon2id$v=19$m=65536,t=3,p=2$example',
  'argon2id',
  CURRENT_TIMESTAMP
);

INSERT INTO role_assignments (
  account_id,
  role,
  scope_type,
  scope_id,
  assigned_by
) VALUES (
  'auth_123',
  'buyer',
  'global',
  'global',
  'system'
);

COMMIT;
```

### Hinglish Explanation

Signup partial create nahi hona chahiye. Agar account create ho gaya but credential fail ho gaya, user login nahi kar paayega. Isliye account, credential, and default buyer role ek transaction me write karna safer hai.

---

## Step 13: Login read path design kiya

Login me identifier se account find hoga, credential load hoga, password compare hoga, then token rows insert honge.

### Login lookup query

```sql
SELECT
  a.account_id,
  a.email,
  a.phone,
  a.email_verified,
  a.phone_verified,
  a.status,
  c.password_hash,
  c.password_algo,
  c.failed_attempts,
  c.locked_until
FROM auth_accounts a
JOIN credentials c ON c.account_id = a.account_id
WHERE a.email = 'user@example.com'
   OR a.phone = '+15551234567'
LIMIT 1;
```

### Active roles query

```sql
SELECT role, scope_type, scope_id
FROM role_assignments
WHERE account_id = 'auth_123'
  AND revoked_at IS NULL;
```

### Refresh token insert query

```sql
INSERT INTO refresh_tokens (
  token_id,
  account_id,
  session_id,
  token_hash,
  parent_token_id,
  expires_at
) VALUES (
  'rt_123',
  'auth_123',
  'sess_123',
  '7f83b1657ff1fc53b92dc18148a1d65dfa135818d208e29b',
  NULL,
  CURRENT_TIMESTAMP + INTERVAL 30 DAY
);
```

> 🔴 **Security note:** Example hash shortened-looking values documentation ke liye hain. Real `token_hash` exactly 64 hex chars hona chahiye if SHA-256 hex use kar rahe ho.

---

## Step 14: OTP queries design kiye

### Create OTP challenge

```sql
INSERT INTO otp_challenges (
  challenge_id,
  account_id,
  target,
  channel,
  purpose,
  otp_hash,
  max_attempts,
  expires_at
) VALUES (
  'otp_123',
  'auth_123',
  'user@example.com',
  'email',
  'email_verify',
  'e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934c',
  5,
  CURRENT_TIMESTAMP + INTERVAL 5 MINUTE
);
```

### Verify OTP challenge

```sql
SELECT
  challenge_id,
  account_id,
  target,
  channel,
  purpose,
  otp_hash,
  attempts,
  max_attempts,
  verified_at,
  expires_at
FROM otp_challenges
WHERE challenge_id = 'otp_123'
LIMIT 1;
```

### Mark OTP verified

```sql
UPDATE otp_challenges
SET verified_at = CURRENT_TIMESTAMP
WHERE challenge_id = 'otp_123'
  AND verified_at IS NULL
  AND expires_at > CURRENT_TIMESTAMP
  AND attempts < max_attempts;
```

### Increment failed attempt

```sql
UPDATE otp_challenges
SET attempts = attempts + 1
WHERE challenge_id = 'otp_123'
  AND verified_at IS NULL
  AND expires_at > CURRENT_TIMESTAMP
  AND attempts < max_attempts;
```

### Hinglish Explanation

OTP verify karte waqt row lock ya transaction use karna better hoga, taaki same OTP parallel requests me replay na ho. Exact repository locking Task 5 me implement hogi, but schema already `verified_at`, `attempts`, and `expires_at` support karta hai.

---

## Step 15: Token cleanup and account cleanup plan banaya

Auth DB me expired tokens and OTP challenges grow kar sakte hain. Cleanup job future me required hogi.

### Cleanup queries

```sql
DELETE FROM refresh_tokens
WHERE expires_at < CURRENT_TIMESTAMP - INTERVAL 30 DAY;
```

```sql
DELETE FROM otp_challenges
WHERE expires_at < CURRENT_TIMESTAMP - INTERVAL 7 DAY;
```

### Why delayed delete?

| Data | Delay reason |
|---|---|
| Refresh tokens | Security investigation/reuse detection ke liye short audit window useful hai |
| OTP challenges | Abuse/rate investigation ke liye limited retention useful hai |
| Deleted accounts | Legal/privacy policy ke according anonymize/delete hoga |

> 🟡 **Privacy note:** Retention policy final compliance decision ke saath align hogi. Default idea: secrets never stored, audit metadata limited time tak.

---

## Step 16: Verification checklist banaya

Migration run ke baad ye queries execute karke schema verify kar sakte ho.

### Tables exist

```sql
SHOW TABLES FROM auth_db;
```

Expected:

```text
auth_accounts
credentials
otp_challenges
refresh_tokens
role_assignments
```

### Table structure inspect

```sql
DESCRIBE auth_accounts;
DESCRIBE credentials;
DESCRIBE refresh_tokens;
DESCRIBE otp_challenges;
DESCRIBE role_assignments;
```

### Indexes inspect

```sql
SHOW INDEX FROM auth_accounts;
SHOW INDEX FROM credentials;
SHOW INDEX FROM refresh_tokens;
SHOW INDEX FROM otp_challenges;
SHOW INDEX FROM role_assignments;
```

### Foreign keys inspect

```sql
SELECT
  table_name,
  constraint_name,
  referenced_table_name
FROM information_schema.key_column_usage
WHERE table_schema = 'auth_db'
  AND referenced_table_name IS NOT NULL;
```

Expected FK references:

| Table | References |
|---|---|
| `credentials` | `auth_accounts` |
| `refresh_tokens` | `auth_accounts` |
| `role_assignments` | `auth_accounts` |

---

## Step 17: Common failure cases document kiye

| Problem | Likely reason | Fix |
|---|---|---|
| FK creation fails | Parent unique key missing ya table order wrong | `auth_accounts` pehle create karo |
| Duplicate email insert fails | `uk_auth_accounts_email` working | API me clean duplicate signup message map karo |
| Multiple active role duplicate possible | Nullable columns in MySQL unique key | Repository/usecase me active duplicate check add karo |
| Token lookup slow | Missing `token_hash` index | `uk_refresh_tokens_hash` verify karo |
| OTP resend lookup slow | Missing target/purpose/expiry index | `idx_otp_target_purpose_expiry` verify karo |
| Timezone mismatch | MySQL local timezone different | `SET time_zone = '+00:00'` and UTC app time use karo |

---

## Step 18: Security decisions lock kiye

### Sensitive data policy

| Data | Store in DB? | How |
|---|---:|---|
| Plain password | ❌ | Never |
| Password hash | ✅ | `credentials.password_hash` |
| Plain refresh token | ❌ | Never |
| Refresh token hash | ✅ | `refresh_tokens.token_hash` |
| Plain OTP | ❌ | Never |
| OTP hash | ✅ | `otp_challenges.otp_hash` |
| JWT private key | ❌ | Secret manager only |
| Full JWT | ❌ | Not in DB/logs |
| Email/phone | ✅ | Auth identity, mask in logs |

### Schema-level security rules

- Use `InnoDB` for transactions and foreign keys.
- Store timestamps in UTC.
- Keep token/OTP hashes fixed length.
- Keep account status explicit with enum.
- Use soft revoke for refresh tokens and roles.
- Avoid cross-service foreign keys.
- Do not add raw device/IP PII in Auth DB unless hashed and required.

---

## 🧪 Future Test Scenarios

Task 8 me security tests implement honge. Task 2 ke schema ke liye expected test cases:

| Test | Expected |
|---|---|
| Duplicate email account insert | Reject with unique key |
| Duplicate phone account insert | Reject with unique key |
| Credential without account | Reject with FK error |
| Second credential for same account | Reject with unique key |
| Refresh token duplicate hash | Reject with unique key |
| Refresh token without account | Reject with FK error |
| OTP challenge duplicate challenge id | Reject with unique key |
| Role assignment without account | Reject with FK error |
| Expired token cleanup query | Deletes old expired rows only |
| OTP expired verify update | Does not mark verified |

---

## 🧾 Final Schema Snapshot

| Table | Main responsibility | Strong consistency reason |
|---|---|---|
| `auth_accounts` | Login identity and account status | Duplicate identities must be prevented |
| `credentials` | Password hash and lock state | Password validation needs reliable state |
| `refresh_tokens` | Rotating refresh token audit | Reuse detection and logout need durable records |
| `otp_challenges` | OTP verification state | Attempts, expiry, replay prevention need consistency |
| `role_assignments` | RBAC source of truth | Permission claims must be controlled |

---

## ✅ Final Task 2 Standard

Auth Service Task 2 ke baad team ke paas ye clear decisions available hain:

- Auth DB ka root table `auth_accounts` hoga.
- Credential data `credentials` me one-to-one account relation ke saath rahega.
- Refresh tokens plain text me store nahi honge; only `token_hash` store hoga.
- OTP plain text me store nahi hoga; only `otp_hash` store hoga.
- Roles Auth boundary me `role_assignments` table se manage honge.
- MySQL foreign keys only Auth DB ke andar use honge, cross-service nahi.
- Migrations `up/down` files ke form me versioned honi chahiye.
- Schema MySQL 8+, InnoDB, `utf8mb4`, and UTC timestamps ke saath ready hai.
- Task 2 scope database schema tak limited rakha gaya hai.

Task 2 complete hai as a structured MySQL schema implementation guide. Ab Task 3 me isi schema ke basis par password hashing safely implement kiya ja sakta hai.
