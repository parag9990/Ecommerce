# User App Frontend - Task 1 Dependency and Setup Documentation

Source task file:

```text
TaskImplementation/User App Frontend/task1.md
```

Generated dependency file:

```text
TaskImplementation/User App Frontend/task1_Dependency.md
```

This document explains dependencies, setup, environment, Docker, ports, backend runtime assumptions, common errors, and DevOps notes for the User App Frontend setup.

Important scope note:

- Task 1 ka original scope Vite + React + TypeScript + Tailwind + strict TypeScript + linting foundation tha.
- Current `frontend/user-app` package me later frontend dependencies bhi present hain, jaise React Router, React Query, Zustand, React Hook Form, Zod, ConnectRPC, and Vitest.
- Frontend direct database se connect nahi karta. Browser REST API Gateway and gRPC-Web bridge ko call karta hai.
- Database, Redis, queues, search engine, and migrations backend side ke dependencies hain. Frontend ko real data tab milega jab backend stack running hoga.

---

## 1. Project Tech Stack Analysis

### Main frontend stack

| Technology | Required? | Where Used | Beginner-Friendly Explanation |
|---|---:|---|---|
| Node.js `>=22.13.0` | Required | `frontend/package.json`, `frontend/user-app/package.json` | Node.js JavaScript runtime hai. Is project me Vite, TypeScript, ESLint, Vitest jaise tools chalane ke liye use hota hai. |
| pnpm `>=11.5.0` | Required | `frontend/pnpm-workspace.yaml`, `frontend/pnpm-lock.yaml` | pnpm package manager hai. Ye dependencies install karta hai aur monorepo workspace ko manage karta hai. |
| Vite `8.0.14` | Required | `frontend/user-app/vite.config.ts` | Vite fast frontend dev server and production build tool hai. Local development me hot reload provide karta hai. |
| React `19.2.6` | Required | `frontend/user-app/src/main.tsx`, components/pages | React UI library hai. Isse reusable components banakar user app ka interface render hota hai. |
| React DOM `19.2.6` | Required | `frontend/user-app/src/main.tsx` | React DOM browser ke real DOM me React app mount karta hai. |
| TypeScript `6.0.3` | Required | `tsconfig*.json`, `src/**/*.ts`, `src/**/*.tsx` | TypeScript JavaScript me types add karta hai. Strict mode se bugs compile time par pakde jate hain. |
| Tailwind CSS `4.3.0` | Required | `vite.config.ts`, `src/styles/globals.css` | Tailwind utility CSS framework hai. Isse classes se fast and consistent styling hoti hai. |
| `@tailwindcss/vite` | Required | `vite.config.ts` | Tailwind v4 ko Vite build pipeline ke saath connect karta hai. |
| ESLint `10.4.1` | Required for dev quality | `frontend/user-app/eslint.config.js` | ESLint code quality checker hai. Ye unsafe patterns, unused imports, and style issues catch karta hai. |
| `typescript-eslint` | Required for dev quality | `eslint.config.js` | TypeScript code ko ESLint rules ke saath lint karne me help karta hai. |
| Vitest `4.1.7` | Required for tests | `vitest.config.ts`, `*.test.ts`, `*.test.tsx` | Vitest test runner hai. Unit and component tests run karta hai. |
| jsdom | Required for component tests | `vitest.config.ts` | jsdom browser-like environment deta hai jisme React component tests run hote hain. |
| Testing Library | Required for component tests | `*.test.tsx` | React components ko user behavior ke perspective se test karne ke liye use hota hai. |

### App feature dependencies currently present

| Technology | Required? | Why Project Uses It | Simple Explanation |
|---|---:|---|---|
| React Router DOM | Required | App routes like home, product detail, cart, checkout, profile, login | React Router page navigation manage karta hai bina full page reload ke. |
| TanStack React Query | Required | Server data fetching, caching, mutation invalidation | React Query API se aane wale data ko cache and sync karta hai. |
| Zustand | Required | UI, auth, and session stores | Zustand lightweight state management library hai. Local app state store karne ke liye use hota hai. |
| React Hook Form | Required | Auth, address, profile, checkout forms | Form state and validation ko simple banata hai. |
| Zod | Required | Form and checkout schemas | Zod runtime validation library hai. User input ko validate karta hai. |
| `@hookform/resolvers` | Required | Zod + React Hook Form bridge | Zod schema ko React Hook Form validation me connect karta hai. |
| lucide-react | Optional but used | Icons in buttons, header, cart, profile, checkout | lucide-react icon library hai. Clean SVG icons provide karta hai. |
| `@connectrpc/connect` | Required for gRPC-Web | gRPC clients and error handling | ConnectRPC typed RPC calls ke liye use hota hai. |
| `@connectrpc/connect-web` | Required for gRPC-Web | Browser transport in `src/lib/grpc-client.ts` | Browser se gRPC-Web endpoint call karne ke liye transport provide karta hai. |
| `@bufbuild/protobuf` | Required for generated proto files | Runtime for generated protobuf TypeScript files | Protobuf messages ko TypeScript/browser me use karne ke liye runtime. |
| `@ecommerce/proto-client` | Required for gRPC-Web features | Workspace package with generated clients | Project ka internal generated TypeScript proto client package hai. |
| Fetch API | Required | `src/lib/http.ts` REST calls | Browser built-in API hai jo HTTP requests bhejta hai. |

### Why this stack is used

User App Frontend buyer-facing single page app hai. Isliye project ko ye capabilities chahiye:

- Fast local development: Vite.
- Component based UI: React.
- Type safety: TypeScript strict mode.
- Consistent styling: Tailwind CSS.
- Clean navigation: React Router.
- API data cache: React Query.
- Local UI/session state: Zustand.
- Form validation: React Hook Form + Zod.
- Typed gRPC-Web calls: ConnectRPC + generated proto client.
- Code quality: ESLint, TypeScript, Vitest.

---

## 2. Node.js Dependency System

This is a Node.js frontend project with pnpm workspace.

### Important files

| File | Purpose |
|---|---|
| `frontend/package.json` | Frontend workspace root scripts and engine requirements. |
| `frontend/user-app/package.json` | User app dependencies, devDependencies, and scripts. |
| `frontend/pnpm-workspace.yaml` | Tells pnpm which packages belong to workspace. |
| `frontend/pnpm-lock.yaml` | Exact dependency versions lock karta hai. Isse same install team machines par repeat hota hai. |
| `frontend/node_modules/` | Installed packages ka local folder. Commit nahi karna chahiye. |
| `frontend/tsconfig.base.json` | Shared strict TypeScript config. |
| `frontend/user-app/tsconfig*.json` | User app specific TypeScript configs. |
| `frontend/user-app/vite.config.ts` | Vite + React + Tailwind config. |
| `frontend/user-app/eslint.config.js` | ESLint config. |

### package.json kya hota hai?

`package.json` project ka dependency manifest hai. Simple words me:

- Kaunse libraries chahiye.
- Kaunse dev tools chahiye.
- Kaunse commands available hain.
- Node/pnpm versions kya expected hain.

Example scripts from current user app:

```json
{
  "scripts": {
    "dev": "vite",
    "build": "tsc -b && vite build",
    "preview": "vite preview",
    "lint": "eslint . --max-warnings=0",
    "test": "vitest run",
    "typecheck": "tsc -b --noEmit"
  }
}
```

### pnpm-lock.yaml kya hota hai?

`pnpm-lock.yaml` exact resolved dependency versions store karta hai.

Beginner tip:

- `package.json` bolta hai project ko kya chahiye.
- `pnpm-lock.yaml` bolta hai exact kaunsa version install hua.
- Team projects me lockfile commit karna important hai.

### node_modules kya hota hai?

`node_modules` installed dependencies ka folder hai. Ye heavy hota hai and generated hota hai.

Rules:

- Isko Git me commit mat karo.
- Agar dependency issue aaye, fresh install se recreate kar sakte ho.

### Install commands

Run from repository root:

```bash
cd frontend
corepack enable
corepack prepare pnpm@11.5.0 --activate
pnpm install
```

### Run commands

```bash
cd frontend
pnpm --filter user-app dev
```

Root shortcut:

```bash
cd frontend
pnpm dev:user
```

### Build, lint, typecheck, test

```bash
cd frontend
pnpm --filter user-app typecheck
pnpm --filter user-app lint
pnpm --filter user-app test
pnpm --filter user-app build
pnpm --filter user-app preview
```

### Common Node/pnpm dependency issues

| Problem | Cause | Fix |
|---|---|---|
| `pnpm: command not found` | Corepack/pnpm not enabled | Run `corepack enable` and `corepack prepare pnpm@11.5.0 --activate`. |
| `Unsupported engine` | Node version old hai | Install Node.js `>=22.13.0`. |
| Workspace package not found | Install wrong folder se hua | `cd frontend` then `pnpm install`. |
| Lockfile mismatch | `package.json` changed but lockfile old | Run `pnpm install` from `frontend`. |
| Vite import resolution error | `node_modules` incomplete | Re-run `pnpm install`. |
| TypeScript version mismatch | Global `tsc` use ho raha hai | Use `pnpm --filter user-app typecheck`, not global `tsc`. |

---

## 3. Database Analysis

### Direct database dependency

User App Frontend ka direct database dependency nahi hai.

Important:

- Browser app MySQL, MongoDB, Redis, PostgreSQL, SQLite, Cassandra etc. se direct connect nahi karta.
- Frontend REST calls `VITE_API_BASE_URL` par API Gateway ko bhejta hai.
- gRPC-Web calls `VITE_GRPC_WEB_BASE_URL` par bridge/gateway ko bhejta hai.
- Database credentials frontend `.env` me kabhi nahi dalne chahiye.

### Detected direct DB status

| Database | Directly Used by Frontend? | Required to Start Vite App? | Notes |
|---|---:|---:|---|
| MySQL | No | No | Backend auth/user/order/payment/admin services ke liye planned/required. |
| MongoDB | No | No | Backend product/cart/wishlist/session/recommendation/notification services ke liye planned/required. |
| Redis | No direct browser use | No | Backend cache/session/rate-limit/cart support ke liye planned/required. |
| PostgreSQL | No | No | Not detected. |
| SQLite | No | No | Not detected. |
| Cassandra | No | No | Not detected. |

### When do you need databases?

Only static frontend screen dekhna hai:

- Database not required.
- API Gateway not required.
- Run Vite app only.

Real login/product/cart/checkout flows test karne hain:

- Backend API Gateway running hona chahiye.
- Backend services running hone chahiye.
- Backend databases/caches/queues running hone chahiye.

### MySQL setup for backend services

#### A. What it is

MySQL ek relational database hai jisme data tables ke form me store hota hai. Auth, users, orders, payments jaise structured data ke liye useful hota hai.

#### B. Why project may use it

Project architecture docs me auth, user, order, payment, CMS, and admin services ke liye MySQL mention hai.

#### C. Required or optional

- Frontend start karne ke liye optional.
- Real backend flows ke liye required, depending on backend services.

#### D. Local installation

Windows:

```text
Install MySQL Installer from official MySQL website, then install MySQL Server and MySQL Shell.
```

Linux Ubuntu/Debian:

```bash
sudo apt update
sudo apt install mysql-server mysql-client
sudo systemctl enable mysql
sudo systemctl start mysql
```

macOS:

```bash
brew install mysql
brew services start mysql
```

#### E. Docker setup with persistent volume

```bash
docker volume create ecommerce_mysql_data
docker run -d \
  --name ecommerce-mysql \
  -e MYSQL_ROOT_PASSWORD=dev_root_password \
  -e MYSQL_DATABASE=ecommerce_dev \
  -e MYSQL_USER=ecommerce_user \
  -e MYSQL_PASSWORD=ecommerce_password \
  -p 3306:3306 \
  -v ecommerce_mysql_data:/var/lib/mysql \
  mysql:8.4
```

#### F. Start commands

Docker:

```bash
docker start ecommerce-mysql
```

Linux service:

```bash
sudo systemctl start mysql
```

macOS Homebrew:

```bash
brew services start mysql
```

#### G. Verify running

```bash
docker ps
docker logs ecommerce-mysql
docker exec -it ecommerce-mysql mysql -uecommerce_user -pecommerce_password ecommerce_dev
```

Local MySQL client:

```bash
mysql -h 127.0.0.1 -P 3306 -uecommerce_user -pecommerce_password ecommerce_dev
```

#### H. Default port

```text
3306
```

#### I. Connection string format

Backend Go services often use DSN format like:

```env
MYSQL_DSN=ecommerce_user:ecommerce_password@tcp(localhost:3306)/ecommerce_dev?parseTime=true
```

Inside Docker Compose network:

```env
MYSQL_DSN=ecommerce_user:ecommerce_password@tcp(mysql:3306)/ecommerce_dev?parseTime=true
```

#### J. Where to place credentials

Use backend service `.env` files or Docker Compose `.env`, not frontend env:

```text
backend/services/<service-name>/.env
infra/compose/.env
```

Never place MySQL username/password in:

```text
frontend/user-app/.env.local
```

### MongoDB setup for backend services

#### A. What it is

MongoDB ek document database hai. Isme JSON-like documents store hote hain. Product catalog, cart, wishlist, sessions jaise flexible data ke liye useful hota hai.

#### B. Why project may use it

Architecture docs me product, cart, wishlist, session, recommendation, and notification services ke liye MongoDB mention hai.

#### C. Required or optional

- Frontend Vite app ke liye optional.
- Real backend product/cart/search/session flows ke liye required, depending on service.

#### D. Local installation

Windows:

```text
Install MongoDB Community Server and MongoDB Shell from official MongoDB website.
```

Linux Ubuntu/Debian:

```bash
sudo apt update
sudo apt install -y mongodb-mongosh
```

macOS:

```bash
brew tap mongodb/brew
brew install mongodb-community
brew services start mongodb-community
```

#### E. Docker setup with persistent volume

```bash
docker volume create ecommerce_mongo_data
docker run -d \
  --name ecommerce-mongo \
  -e MONGO_INITDB_ROOT_USERNAME=ecommerce_root \
  -e MONGO_INITDB_ROOT_PASSWORD=ecommerce_password \
  -p 27017:27017 \
  -v ecommerce_mongo_data:/data/db \
  mongo:7
```

#### F. Start commands

```bash
docker start ecommerce-mongo
```

macOS Homebrew:

```bash
brew services start mongodb-community
```

#### G. Verify running

```bash
docker ps
docker exec -it ecommerce-mongo mongosh \
  "mongodb://ecommerce_root:ecommerce_password@localhost:27017/admin"
```

Local shell:

```bash
mongosh "mongodb://ecommerce_root:ecommerce_password@localhost:27017/admin"
```

#### H. Default port

```text
27017
```

#### I. Connection string format

Host machine:

```env
MONGO_URI=mongodb://ecommerce_root:ecommerce_password@localhost:27017/ecommerce_dev?authSource=admin
```

Docker Compose network:

```env
MONGO_URI=mongodb://ecommerce_root:ecommerce_password@mongo:27017/ecommerce_dev?authSource=admin
```

#### J. Where to place credentials

Use backend service `.env` or Compose `.env`:

```text
backend/services/<service-name>/.env
infra/compose/.env
```

Do not place MongoDB URI in frontend env.

### Redis setup for backend services

#### A. What it is

Redis in-memory data store/cache hai. Fast reads/writes ke liye use hota hai.

#### B. Why project may use it

Architecture docs me Redis rate limiting, sessions, cart cache, OTP retry counters, and gateway/cache support ke liye mention hai.

#### C. Required or optional

- Frontend Vite app ke liye optional.
- Backend gateway/auth/cart/session flows ke liye required, depending on services.

#### D. Local installation

Windows:

```text
Recommended: use Docker Desktop or WSL2, then run Redis container.
```

Linux Ubuntu/Debian:

```bash
sudo apt update
sudo apt install redis-server redis-tools
sudo systemctl enable redis-server
sudo systemctl start redis-server
```

macOS:

```bash
brew install redis
brew services start redis
```

#### E. Docker setup with persistent volume

```bash
docker volume create ecommerce_redis_data
docker run -d \
  --name ecommerce-redis \
  -p 6379:6379 \
  -v ecommerce_redis_data:/data \
  redis:7-alpine \
  redis-server --appendonly yes --requirepass dev_redis_password
```

#### F. Start commands

```bash
docker start ecommerce-redis
```

Linux service:

```bash
sudo systemctl start redis-server
```

macOS Homebrew:

```bash
brew services start redis
```

#### G. Verify running

```bash
docker exec -it ecommerce-redis redis-cli -a dev_redis_password ping
```

Expected:

```text
PONG
```

#### H. Default port

```text
6379
```

#### I. Connection string / config format

```env
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=dev_redis_password
```

Inside Docker Compose network:

```env
REDIS_ADDR=redis:6379
REDIS_PASSWORD=dev_redis_password
```

#### J. Where to place credentials

Backend `.env` or Compose `.env`, not frontend `.env.local`.

---

## 4. Environment Variables

### Where to create env file

For local frontend development:

```bash
cd frontend
cp user-app/.env.example user-app/.env.local
```

Recommended file:

```text
frontend/user-app/.env.local
```

Existing example file:

```text
frontend/user-app/.env.example
```

### Complete frontend `.env` example

```env
VITE_API_BASE_URL=http://localhost:8080
VITE_GRPC_WEB_BASE_URL=http://localhost:8082
VITE_GRPC_WEB_TIMEOUT_MS=5000
VITE_APP_ENV=local
VITE_PAYMENT_PROVIDERS=stripe,razorpay
```

### Environment variable details

| Variable | Required? | Example | Purpose | Security Notes |
|---|---:|---|---|---|
| `VITE_API_BASE_URL` | Required for real REST APIs | `http://localhost:8080` | REST API Gateway ka base URL. Product, auth, cart, checkout etc. REST calls yahan jayengi. | Public browser config hai. Secret mat rakho. |
| `VITE_GRPC_WEB_BASE_URL` | Required for gRPC-Web features | `http://localhost:8082` | gRPC-Web bridge/gateway URL. Search autocomplete, recommendation, session analytics typed RPC calls ke liye. | Public URL hai. Secret mat rakho. |
| `VITE_GRPC_WEB_TIMEOUT_MS` | Optional | `5000` | gRPC-Web request timeout milliseconds me. Invalid/empty ho to code default `5000` use karta hai. | Secret nahi hai. Positive integer rakho. |
| `VITE_APP_ENV` | Optional but validated | `local` | App environment name. Allowed: `local`, `development`, `staging`, `production`. | Wrong value app startup error throw kar sakta hai. |
| `VITE_PAYMENT_PROVIDERS` | Optional | `stripe,razorpay` | Checkout UI me provider names list karne ke liye. | Provider secret keys frontend me mat rakho. |

### How Vite loads env

Vite app root se env files load karta hai. Is project me app root:

```text
frontend/user-app
```

Vite only `VITE_` prefix wali variables browser bundle me expose karta hai.

Example:

```env
VITE_API_BASE_URL=http://localhost:8080
```

Works in browser code.

```env
API_BASE_URL=http://localhost:8080
```

Does not work in Vite browser code because prefix missing hai.

### Important frontend security rule

All `VITE_` variables browser me visible hoti hain. Isliye never add:

```env
MYSQL_PASSWORD=...
JWT_SECRET=...
STRIPE_SECRET_KEY=...
RAZORPAY_SECRET=...
AWS_SECRET_ACCESS_KEY=...
```

Frontend me only public config rakho. Secret values backend services or secret manager me rakho.

### Common env mistakes

| Mistake | Result | Fix |
|---|---|---|
| `.env.local` wrong folder me create karna | App defaults use karega | File `frontend/user-app/.env.local` me banao. |
| `API_BASE_URL` use karna instead of `VITE_API_BASE_URL` | `import.meta.env` me value nahi milegi | Always use `VITE_` prefix. |
| Invalid `VITE_APP_ENV=dev` | App error throw kar sakta hai | Use `local`, `development`, `staging`, or `production`. |
| Secret key in frontend env | Secret browser bundle me leak ho jayega | Secret backend env me rakho. |
| Env change ke baad Vite restart nahi kiya | Old values continue ho sakti hain | Stop dev server and start again. |

---

## 5. External Services Analysis

### Direct external services for frontend runtime

| Service | Required? | Used By | Default URL/Port | What It Does |
|---|---:|---|---|---|
| REST API Gateway | Required for real app data | `src/lib/http.ts`, feature `api/*.ts` | `http://localhost:8080` | Frontend REST requests ko backend services tak route karta hai. |
| gRPC-Web bridge / Envoy / gateway bridge | Required for gRPC-Web features | `src/lib/grpc-client.ts` | `http://localhost:8082` | Browser-compatible gRPC-Web requests ko backend gRPC services tak bridge karta hai. |
| Payment provider public flow | Optional/currently partial | Checkout UI provider list | `VITE_PAYMENT_PROVIDERS` | Frontend provider names show karta hai. Real secret/payment handling backend me hona chahiye. |

### API Gateway

What it is:

API Gateway ek entry service hota hai jo browser ke REST requests receive karta hai and internal backend services ko forward karta hai.

Why used:

Frontend ko har backend service ka direct URL ya internal topology know nahi karna padta. Sirf `VITE_API_BASE_URL` use hota hai.

Mandatory or optional:

- Static frontend open karne ke liye optional.
- Login, products, cart, checkout, orders, profile, wishlist ke real flows ke liye mandatory.

Start command:

Backend source command current task file me clearly defined nahi hai. Agar gateway service implementation available ho, typical command:

```bash
cd backend
go run ./services/api-gateway/cmd/server
```

Health check:

```bash
curl http://localhost:8080/health/live
curl http://localhost:8080/health/ready
```

Common issues:

| Issue | Cause | Fix |
|---|---|---|
| Frontend `NETWORK_ERROR` | Gateway not running | Start gateway or update `VITE_API_BASE_URL`. |
| CORS error | Gateway does not allow Vite origin | Allow `http://localhost:5173` and credentials. |
| 401 unauthorized | Missing/expired token | Login again and verify backend auth. |
| 404 endpoint | Gateway route missing | Compare frontend API path with API contract. |

### gRPC-Web bridge

What it is:

Browser native gRPC HTTP/2 directly use nahi kar sakta in the same way backend services do. gRPC-Web bridge browser requests ko backend gRPC services ke compatible format me convert karta hai.

Why used:

Current code uses ConnectRPC for:

- Recommendation service.
- Search autocomplete service.
- Session analytics ingestion service.

Mandatory or optional:

- Required only for gRPC-Web features.
- REST-only screens may still work without it.

Suggested Docker shape:

```yaml
grpc-web:
  image: envoyproxy/envoy:v1.31-latest
  ports:
    - "8082:8082"
  volumes:
    - ./envoy.yaml:/etc/envoy/envoy.yaml:ro
```

Health check:

```bash
curl http://localhost:8082
```

Common issues:

| Issue | Cause | Fix |
|---|---|---|
| gRPC `UNAVAILABLE` | Bridge/backend not running | Start bridge and target backend service. |
| CORS preflight failed | Bridge CORS not configured | Allow `http://localhost:5173`, methods, headers, credentials. |
| Deadline exceeded | Timeout low or backend slow | Increase `VITE_GRPC_WEB_TIMEOUT_MS` or fix backend latency. |
| Generated client import fails | Workspace package not installed | Run `pnpm install` from `frontend`. |

### Payment providers

Detected:

```env
VITE_PAYMENT_PROVIDERS=stripe,razorpay
```

What it is:

Stripe/Razorpay payment providers online payments process karte hain.

Current status:

- Frontend has provider names.
- No Stripe/Razorpay SDK dependency detected in `frontend/user-app/package.json`.
- Secret keys must be backend only.

Credential placement:

Frontend:

```env
VITE_PAYMENT_PROVIDERS=stripe,razorpay
```

Backend only:

```env
STRIPE_SECRET_KEY=...
STRIPE_WEBHOOK_SECRET=...
RAZORPAY_KEY_SECRET=...
```

Never put backend payment secrets in frontend env.

### Typesense search engine

What it is:

Typesense fast search engine hai. Product search/autocomplete ke liye useful hota hai.

Required?

- Frontend direct use nahi karta.
- Backend search service ke liye optional/planned based on docs.

Docker setup:

```bash
docker volume create ecommerce_typesense_data
docker run -d \
  --name ecommerce-typesense \
  -p 8108:8108 \
  -v ecommerce_typesense_data:/data \
  typesense/typesense:latest \
  --data-dir /data \
  --api-key dev-typesense-key \
  --enable-cors
```

Verify:

```bash
curl http://localhost:8108/health
```

Credentials:

```env
TYPESENSE_API_KEY=dev-typesense-key
TYPESENSE_HOST=localhost
TYPESENSE_PORT=8108
```

Place these in backend search service env, not frontend.

### RabbitMQ

What it is:

RabbitMQ message broker hai. Services ke beech async commands/events bhejne ke liye use hota hai.

Required?

- Frontend ke liye direct nahi.
- Backend notification/order/payment/search flows ke liye optional/planned.

Docker setup:

```bash
docker run -d \
  --name ecommerce-rabbitmq \
  -p 5672:5672 \
  -p 15672:15672 \
  -e RABBITMQ_DEFAULT_USER=ecommerce \
  -e RABBITMQ_DEFAULT_PASS=ecommerce_password \
  -e RABBITMQ_DEFAULT_VHOST=ecommerce \
  rabbitmq:3-management
```

Verify:

```bash
docker logs ecommerce-rabbitmq
```

Management UI:

```text
http://localhost:15672
```

Connection string:

```env
RABBITMQ_URL=amqp://ecommerce:ecommerce_password@localhost:5672/ecommerce
```

### Kafka

What it is:

Kafka event streaming platform hai. High-throughput events like product events, order events, payment events, session events ke liye use hota hai.

Required?

- Frontend ke liye direct nahi.
- Backend event streaming ke liye optional/planned.

Simple Docker Compose service example:

```yaml
kafka:
  image: bitnami/kafka:latest
  environment:
    KAFKA_ENABLE_KRAFT: "yes"
    KAFKA_CFG_NODE_ID: "1"
    KAFKA_CFG_PROCESS_ROLES: "broker,controller"
    KAFKA_CFG_CONTROLLER_LISTENER_NAMES: "CONTROLLER"
    KAFKA_CFG_LISTENERS: "PLAINTEXT://:9092,CONTROLLER://:9093"
    KAFKA_CFG_ADVERTISED_LISTENERS: "PLAINTEXT://kafka:9092"
    KAFKA_CFG_CONTROLLER_QUORUM_VOTERS: "1@kafka:9093"
    ALLOW_PLAINTEXT_LISTENER: "yes"
  ports:
    - "9092:9092"
```

Backend env:

```env
KAFKA_BROKERS=localhost:9092
```

### Docker

What it is:

Docker containers me dependencies run karne ka tool hai. Beginner ke liye MySQL/MongoDB/Redis locally install karne se easier ho sakta hai.

Current project status:

- Actual Dockerfile/docker-compose files were not clearly found in the repository root during inspection.
- Docs mention planned Docker setup.
- For frontend local development, Node + pnpm is enough.
- For backend dependencies, Docker is recommended.

---

## 6. Ports and Networking

| Service | Port | Purpose | Directly Needed by Frontend Start? |
|---|---:|---|---:|
| Vite User App dev server | `5173` | Local React dev server | Yes |
| Vite preview server | `4173` | Preview production build locally | Optional |
| REST API Gateway | `8080` | Frontend REST API target | Required for real data |
| gRPC-Web bridge | `8082` | Frontend gRPC-Web target | Required for gRPC features |
| MySQL | `3306` | Backend relational DB | No direct frontend use |
| MongoDB | `27017` | Backend document DB | No direct frontend use |
| Redis | `6379` | Backend cache/session/rate limit | No direct frontend use |
| Typesense | `8108` | Backend search engine | No direct frontend use |
| RabbitMQ AMQP | `5672` | Backend message broker | No direct frontend use |
| RabbitMQ UI | `15672` | Broker management UI | No direct frontend use |
| Kafka | `9092` | Backend event streaming | No direct frontend use |

### Port conflict explanation

Port conflict ka matlab same port par already koi process chal raha hai.

Check port on Linux/macOS:

```bash
lsof -i :5173
lsof -i :8080
```

Alternative:

```bash
ss -ltnp | grep 5173
```

Change Vite port:

```bash
cd frontend
pnpm --filter user-app dev -- --port 5174
```

Change Docker host port:

```yaml
ports:
  - "3307:3306"
```

Meaning:

- Host machine: `localhost:3307`
- Container network: `mysql:3306`

### Firewall/network issues

Common beginner problems:

- Browser cannot reach API because gateway is stopped.
- Docker container can reach service name `mysql`, but host machine must use `localhost`.
- CORS not configured for `http://localhost:5173`.
- WSL2/Docker Desktop networking may need restart.

---

## 7. Docker and DevOps Setup

### Current Docker status

| Item | Status |
|---|---|
| Frontend Dockerfile | Not clearly found |
| docker-compose.yml | Not clearly found |
| Nginx config for frontend static hosting | Not clearly found |
| Docker recommended for local backing services | Yes |

### Recommended beginner approach

Use local Node/pnpm for frontend:

```bash
cd frontend
pnpm install
pnpm --filter user-app dev
```

Use Docker for databases and external services:

```bash
docker ps
docker logs <container-name>
docker stop <container-name>
docker start <container-name>
```

### Suggested local docker-compose example

This compose file is a suggested DevOps setup. It is not currently detected as an existing file in the repo.

```yaml
services:
  mysql:
    image: mysql:8.4
    container_name: ecommerce-mysql
    environment:
      MYSQL_ROOT_PASSWORD: dev_root_password
      MYSQL_DATABASE: ecommerce_dev
      MYSQL_USER: ecommerce_user
      MYSQL_PASSWORD: ecommerce_password
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql
    restart: unless-stopped
    healthcheck:
      test: ["CMD-SHELL", "mysqladmin ping -h 127.0.0.1 -uroot -p$${MYSQL_ROOT_PASSWORD}"]
      interval: 10s
      timeout: 5s
      retries: 10

  mongo:
    image: mongo:7
    container_name: ecommerce-mongo
    environment:
      MONGO_INITDB_ROOT_USERNAME: ecommerce_root
      MONGO_INITDB_ROOT_PASSWORD: ecommerce_password
    ports:
      - "27017:27017"
    volumes:
      - mongo_data:/data/db
    restart: unless-stopped

  redis:
    image: redis:7-alpine
    container_name: ecommerce-redis
    command: ["redis-server", "--appendonly", "yes", "--requirepass", "dev_redis_password"]
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    restart: unless-stopped
    healthcheck:
      test: ["CMD-SHELL", "redis-cli -a dev_redis_password ping | grep PONG"]
      interval: 10s
      timeout: 5s
      retries: 10

  typesense:
    image: typesense/typesense:latest
    container_name: ecommerce-typesense
    command: ["--data-dir", "/data", "--api-key", "dev-typesense-key", "--enable-cors"]
    ports:
      - "8108:8108"
    volumes:
      - typesense_data:/data
    restart: unless-stopped

  rabbitmq:
    image: rabbitmq:3-management
    container_name: ecommerce-rabbitmq
    environment:
      RABBITMQ_DEFAULT_USER: ecommerce
      RABBITMQ_DEFAULT_PASS: ecommerce_password
      RABBITMQ_DEFAULT_VHOST: ecommerce
    ports:
      - "5672:5672"
      - "15672:15672"
    restart: unless-stopped

volumes:
  mysql_data:
  mongo_data:
  redis_data:
  typesense_data:
```

### Compose commands

```bash
docker compose up -d
docker compose ps
docker compose logs
docker compose logs mysql
docker compose down
```

Remove volumes only when you intentionally want to delete local data:

```bash
docker compose down -v
```

Warning:

`down -v` database data delete kar dega. Use carefully.

### Suggested frontend Dockerfile for production

This is a suggested Dockerfile, not detected as an existing repo file.

```dockerfile
FROM node:22-alpine AS builder
WORKDIR /app

RUN corepack enable

COPY frontend/package.json frontend/pnpm-lock.yaml frontend/pnpm-workspace.yaml ./
COPY frontend/user-app/package.json ./user-app/package.json
COPY frontend/packages/proto-client/package.json ./packages/proto-client/package.json

RUN pnpm install --frozen-lockfile

COPY frontend ./
RUN pnpm --filter user-app build

FROM nginx:alpine
COPY --from=builder /app/user-app/dist /usr/share/nginx/html
EXPOSE 80
```

Production note:

Vite `VITE_` variables build time par bundle me bake ho sakti hain. Production deployment me correct API URLs build-time ya runtime config strategy se set karo.

---

## 8. Complete Project Run Instructions

### Step 1: Clone repository

```bash
git clone <repository-url>
cd Ecommerce
```

### Step 2: Check runtime versions

```bash
node --version
corepack --version
```

Expected:

```text
Node.js >= 22.13.0
pnpm >= 11.5.0
```

Enable pnpm:

```bash
corepack enable
corepack prepare pnpm@11.5.0 --activate
```

### Step 3: Install frontend dependencies

```bash
cd frontend
pnpm install
```

### Step 4: Create frontend env file

```bash
cp user-app/.env.example user-app/.env.local
```

Open `frontend/user-app/.env.local` and confirm:

```env
VITE_API_BASE_URL=http://localhost:8080
VITE_GRPC_WEB_BASE_URL=http://localhost:8082
VITE_GRPC_WEB_TIMEOUT_MS=5000
VITE_APP_ENV=local
VITE_PAYMENT_PROVIDERS=stripe,razorpay
```

### Step 5: Start optional backend dependencies

For static frontend foundation check, skip this step.

For real APIs, start backend dependency containers:

```bash
docker compose up -d
docker compose ps
```

If compose file is not available, run individual Docker commands from database/external service sections above.

### Step 6: Run migrations

Frontend has no database migrations.

Backend services may require migrations. Typical backend flow:

```bash
cd backend
go run ./cmd/migrate up
```

If migration command is missing, check backend service docs for the exact command.

### Step 7: Start backend API Gateway

Only needed for real data:

```bash
cd backend
go run ./services/api-gateway/cmd/server
```

Verify:

```bash
curl http://localhost:8080/health/live
```

### Step 8: Start gRPC-Web bridge

Only needed for gRPC-Web features:

```bash
docker compose up -d grpc-web
```

Verify:

```bash
curl http://localhost:8082
```

### Step 9: Start User App Frontend

```bash
cd frontend
pnpm --filter user-app dev
```

Open:

```text
http://localhost:5173
```

### Step 10: Verify quality checks

```bash
cd frontend
pnpm --filter user-app typecheck
pnpm --filter user-app lint
pnpm --filter user-app test
pnpm --filter user-app build
```

Expected:

```text
typecheck: pass
lint: pass
test: pass
build: pass
```

---

## 9. Common Errors and Fixes

| Error | Cause | Fix | Prevention |
|---|---|---|---|
| `pnpm: command not found` | pnpm not enabled | Run `corepack enable` and `corepack prepare pnpm@11.5.0 --activate`. | Use Node 22+ with Corepack. |
| `Unsupported engine` | Old Node version | Install Node.js `>=22.13.0`. | Check `node --version` before setup. |
| `ERR_PNPM_WORKSPACE_PKG_NOT_FOUND` | Workspace dependency not linked | Run `pnpm install` from `frontend`. | Do installs from workspace root. |
| Vite port already in use | Another process using `5173` | Run `pnpm --filter user-app dev -- --port 5174`. | Keep one dev server per port. |
| Browser blank page | Build/runtime JS error | Check browser console and terminal logs. | Run typecheck/lint before dev. |
| Missing `.env.local` | Env file not created | `cp user-app/.env.example user-app/.env.local`. | Keep `.env.example` updated. |
| Env changed but app still old | Vite dev server not restarted | Stop and restart Vite. | Restart after env changes. |
| `Invalid VITE_APP_ENV` | Unsupported env value | Use `local`, `development`, `staging`, or `production`. | Copy from `.env.example`. |
| `NETWORK_ERROR` in frontend | API Gateway stopped or URL wrong | Start gateway or update `VITE_API_BASE_URL`. | Health check gateway first. |
| CORS error | Backend/bridge not allowing Vite origin | Allow `http://localhost:5173` and credentials. | Configure CORS in gateway/bridge. |
| 401 unauthorized | Missing/expired auth token | Login again or clear browser storage. | Keep auth flow and token refresh working. |
| gRPC `UNAVAILABLE` | gRPC-Web bridge/backend service stopped | Start bridge and target backend service. | Check `VITE_GRPC_WEB_BASE_URL`. |
| gRPC deadline exceeded | Slow backend or low timeout | Increase `VITE_GRPC_WEB_TIMEOUT_MS` or debug service latency. | Use sensible timeouts. |
| Docker daemon not running | Docker Desktop/service stopped | Start Docker Desktop or `sudo systemctl start docker`. | Start Docker before containers. |
| MySQL connection refused | MySQL container/service stopped | Start MySQL and verify port `3306`. | Use Docker healthchecks. |
| MongoDB auth failed | Wrong username/password/authSource | Check `MONGO_URI` and credentials. | Store credentials in backend `.env`. |
| Redis `NOAUTH` | Redis requires password | Add `REDIS_PASSWORD` in backend config. | Keep Redis config consistent. |
| RabbitMQ login failed | Wrong user/pass/vhost | Verify `RABBITMQ_DEFAULT_*` values. | Use one local `.env` source. |
| Kafka broker failure | Kafka not ready or advertised listeners wrong | Check Kafka logs and listener config. | Use tested compose config. |
| `go mod download failed` | Backend dependency/network issue | Run from backend module and check Go proxy/network. | Keep Go version and module files aligned. |
| Migration failed | DB unavailable or old schema state | Start DB, check credentials, rerun migration carefully. | Use up/down migrations and backups. |
| Permission denied | File/container permission issue | Check file owner, Docker volume permissions. | Avoid running mixed root/user writes. |
| Invalid payment credentials | Backend provider secret wrong | Update backend payment `.env`. | Never store secrets in frontend. |

---

## 10. Security and Configuration Audit

### What looks good

| Area | Finding |
|---|---|
| Frontend env prefix | Uses `VITE_` variables, which is correct for Vite public config. |
| No direct DB credentials in frontend example | `.env.example` contains only URLs/provider names/timeouts. |
| Strict TypeScript | `frontend/tsconfig.base.json` has strong strict options enabled. |
| Lint quality gate | `eslint . --max-warnings=0` is configured. |
| Build command | `tsc -b && vite build` catches TypeScript errors before bundle. |

### Risks / gaps found

| Risk | Impact | Suggested Fix |
|---|---|---|
| Frontend `VITE_` variables are public | Secrets can leak if added accidentally | Never add passwords, JWT secrets, payment secret keys, DB URIs in frontend env. |
| gRPC-Web bridge config not clearly found | gRPC features may fail locally | Add Envoy/gateway bridge config with CORS for `http://localhost:5173`. |
| Dockerfile/compose not clearly found | Beginner cannot start full stack easily | Add `infra/compose/docker-compose.local.yml` and frontend Dockerfile when deployment starts. |
| Backend startup/migration commands not clearly documented in Task 1 | Real API setup confusing | Add per-service backend setup docs and migration commands. |
| Payment provider env only lists names | Real payment integration incomplete | Keep secret keys backend-only and add provider SDK only when needed. |
| API URLs default to localhost | Production build may accidentally point to local APIs | Set production `VITE_API_BASE_URL` and `VITE_GRPC_WEB_BASE_URL` during build/deploy. |
| CORS/cookie requirements | Auth requests use credentials and bearer token pattern | Gateway must allow credentials and trusted origins only. |
| No detected frontend health endpoint | Static hosting health check may be missing | In Docker/Nginx, health check `/` or static file response. |

### Hardcoded credentials audit

Frontend code/env example:

- No DB credentials detected.
- No payment secret detected.
- No JWT secret detected.

Documentation examples:

- Example passwords like `ecommerce_password` and `dev_redis_password` are local placeholders only.
- Production me strong secrets and secret manager use karo.

---

## 11. Best Practices

### Frontend dependency best practices

- `pnpm install` hamesha `frontend` folder se run karo.
- `pnpm-lock.yaml` commit karo.
- `node_modules` commit mat karo.
- Node.js and pnpm versions package `engines` ke saath match rakho.
- Global `tsc`, `eslint`, or `vite` use karne se avoid karo; pnpm scripts use karo.
- Pull ke baad dependency errors aaye to `pnpm install` rerun karo.

### Environment best practices

- `.env.local` commit mat karo.
- `.env.example` updated rakho.
- Frontend env me only public config rakho.
- Secrets backend `.env`, Docker secrets, Kubernetes Secrets, or cloud secret manager me rakho.
- Env change ke baad Vite restart karo.

### Database and Docker best practices

- Docker volumes use karo taaki DB data container restart ke baad bhi rahe.
- `docker compose down -v` carefully use karo because data delete hota hai.
- Health checks add karo for MySQL, Redis, API Gateway.
- Local and production credentials alag rakho.
- DB backups production me mandatory rakho.

### API and networking best practices

- API Gateway ko single public backend entrypoint rakho.
- CORS sirf trusted origins ke liye allow karo.
- Cookies/credentials use karte time HTTPS production me mandatory rakho.
- gRPC-Web bridge me timeout, CORS, and headers properly configure karo.
- Frontend should never call internal service ports directly.

### Security best practices

- Payment secret keys frontend me kabhi nahi.
- JWT signing secrets frontend me kabhi nahi.
- Database passwords frontend me kabhi nahi.
- Keep dependencies updated and audit regularly.
- Use strong secrets in production.
- Use separate dev/staging/prod configs.

---

## 12. Final Beginner Checklist

### Local frontend checklist

- [ ] Repository cloned.
- [ ] Node.js `>=22.13.0` installed.
- [ ] Corepack enabled.
- [ ] pnpm `>=11.5.0` active.
- [ ] `cd frontend` done.
- [ ] `pnpm install` completed.
- [ ] `frontend/user-app/.env.local` created from `.env.example`.
- [ ] `VITE_API_BASE_URL` checked.
- [ ] `VITE_GRPC_WEB_BASE_URL` checked if gRPC-Web features are needed.
- [ ] `pnpm --filter user-app dev` running.
- [ ] Browser opens `http://localhost:5173`.
- [ ] Browser console checked for errors.

### Quality checklist

- [ ] `pnpm --filter user-app typecheck` passes.
- [ ] `pnpm --filter user-app lint` passes.
- [ ] `pnpm --filter user-app test` passes.
- [ ] `pnpm --filter user-app build` passes.
- [ ] `pnpm --filter user-app preview` works if production build preview is needed.

### Backend/runtime checklist for real flows

- [ ] Docker daemon running.
- [ ] MySQL running if backend relational services need it.
- [ ] MongoDB running if product/cart/session services need it.
- [ ] Redis running if gateway/auth/cart/session need it.
- [ ] Typesense running if search service needs it.
- [ ] RabbitMQ or Kafka running if event flows need it.
- [ ] Backend `.env` files created.
- [ ] Backend migrations completed.
- [ ] API Gateway running on `http://localhost:8080`.
- [ ] API Gateway health check working.
- [ ] gRPC-Web bridge running on `http://localhost:8082` if needed.
- [ ] CORS allows `http://localhost:5173`.
- [ ] Logs checked for frontend, gateway, bridge, and databases.

### Security checklist

- [ ] `.env.local` not committed.
- [ ] No secrets in `VITE_` variables.
- [ ] Payment secret keys backend-only.
- [ ] DB credentials backend-only.
- [ ] Production API URLs not pointing to localhost.
- [ ] Strong production secrets configured outside Git.

---

## Quick Command Summary

```bash
# From repo root
cd frontend

# Enable pnpm
corepack enable
corepack prepare pnpm@11.5.0 --activate

# Install dependencies
pnpm install

# Create local env
cp user-app/.env.example user-app/.env.local

# Start frontend
pnpm --filter user-app dev

# Verify quality
pnpm --filter user-app typecheck
pnpm --filter user-app lint
pnpm --filter user-app test
pnpm --filter user-app build
```

Open:

```text
http://localhost:5173
```

Final note:

For Task 1 foundation verification, Node.js + pnpm + Vite app setup is enough. For complete e-commerce flows, backend API Gateway, gRPC-Web bridge, and backend data services must also be running.
