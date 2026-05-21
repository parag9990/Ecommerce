# Git Workflow & Repository Management Guide

Project context: ye repo e-commerce microservices platform hai. Current implementation me Go Auth Service hai, docs me future React/TypeScript frontend, Docker, Kubernetes, MySQL, MongoDB, Redis, Typesense, and Kafka/RabbitMQ planned hain. Isliye Git workflow code, docs, infra, env, and secrets sabko dhyan me rakh ke follow karo.

## 1. Recommended Git Branch Structure

| Branch | Use | Simple Rule |
|---|---|---|
| `main` / `master` | Production stable code | Direct push mat karo. Sirf tested release merge karo. |
| `dev` / `develop` | Active development integration | Team ka daily combined work yahan merge hota hai. |
| `feature/*` | New feature work | Har feature ke liye separate branch banao. |
| `hotfix/*` | Urgent production bug fix | `main` se banao, fix karo, phir `main` and `dev` dono me merge karo. |
| `release/*` | Release testing/stabilization | Final QA, bug fixes, version prep ke liye. |

```bash
main
dev
feature/auth-login
feature/user-service
hotfix/jwt-expiry
release/v1.0.0
```

Note: Is repo me current branches `main` and `dev` hain. `master` old/default branch naming convention hoti hai; yahan `main` use karo. Team ke liye `dev` ko active development branch use karna practical rahega.

## 2. How to Create Branches

Latest `dev` lo, phir feature branch banao:

```bash
git checkout dev
git pull origin dev
git checkout -b feature/auth-service
```

Branch name clear rakho:

```bash
feature/auth-rbac
feature/user-profile
fix/redis-connection
docs/git-workflow
```

## 3. Proper Git Commit Message Style

Commit message short and meaningful hona chahiye.

```bash
git commit -m "feat: added JWT authentication"
git commit -m "fix: resolved Redis connection issue"
git commit -m "docs: added Git workflow guide"
git commit -m "refactor: cleaned auth token usecase"
git commit -m "chore: updated Go dependencies"
```

| Prefix | Meaning | Example |
|---|---|---|
| `feat` | New feature | `feat: added OTP verification` |
| `fix` | Bug fix | `fix: handled expired refresh token` |
| `docs` | Documentation | `docs: updated developer guide` |
| `refactor` | Code cleanup without behavior change | `refactor: simplified role checks` |
| `chore` | Config, dependency, tooling | `chore: updated go.work` |
| `test` | Tests added/updated | `test: added RBAC security tests` |

## 4. Push & Pull Workflow

`pull` latest code laata hai and local branch me merge karta hai:

```bash
git pull origin dev
```

`push` apni branch GitHub pe bhejta hai:

```bash
git push origin feature/auth-service
```

`fetch` sirf remote updates check karta hai, merge nahi karta:

```bash
git fetch origin
```

Simple rule: kaam start karne se pehle `git pull origin dev` zaroor karo.

## 5. Merge Workflow

Feature complete hone ke baad branch ko `dev` me merge karo. Best way: Pull Request/Merge Request banao.

Local merge example:

```bash
git checkout dev
git pull origin dev
git merge feature/auth-service
git push origin dev
```

Important:

- Direct `main` me push avoid karo.
- Feature branch pe small commits rakho.
- Merge se pehle tests run karo.
- `main` me sirf reviewed and stable code merge karo.

Merge conflict ka matlab same file ki same lines do branches me change hui hain. File open karo, conflict markers remove karo, correct final code rakho:

```bash
git status
git add conflicted-file.go
git commit
```

## 6. .gitignore Setup

Root `.gitignore` add kiya gaya hai. Ye project ke Go backend, future Node/React frontend, env files, secrets, logs, cache, build output, and local tooling files ignore karta hai.

Important entries:

```gitignore
.env
.env.*
**/.env
**/secrets/*
*.pem
*.key
node_modules/
dist/
build/
coverage/
bin/
vendor/
*.test
*.out
logs/
*.log
tmp/
.codex/
.agents/
```

Why ignore?

| Item | Reason |
|---|---|
| `.env` | DB password, Redis password, JWT pepper, API keys ho sakte hain. |
| `secrets/`, `*.pem`, `*.key` | JWT private/public keys and certificates GitHub pe nahi jane chahiye. |
| `node_modules/` | Dependencies huge hoti hain, reinstall ho sakti hain. |
| `dist/`, `build/`, `bin/` | Generated build output hai. |
| `coverage/`, `*.out` | Test-generated files hain. |
| `logs/`, `*.log` | Runtime logs me sensitive data aa sakta hai. |
| `.codex/`, `.agents/` | Local tooling state hai, repo code nahi. |

Commit karne layak files:

```text
go.mod
go.sum
go.work
go.work.sum
docs/
api/master-api.json
database/draw.sql
database/mongodb-schema-design.md
migrations/*.sql
```

## 7. Files That Should NEVER Be Pushed

Kabhi push mat karo:

- `.env`
- JWT private keys: `jwt-private.pem`
- Secrets folder content: `backend/services/auth-service/secrets/`
- DB passwords, Redis passwords, token peppers
- API keys: payment, email, SMS, Typesense
- Logs: `*.log`
- Local build output: `bin/`, `dist/`, `build/`
- DB dumps: `*.dump`, `*.backup`, `*.sql.gz`

Security risk simple hai: agar secret GitHub pe chala gaya, koi bhi DB, tokens, ya external services misuse kar sakta hai.

## 8. Beginner-Friendly Git Workflow

Daily workflow:

### Step 1: Latest `dev` lo

```bash
git checkout dev
git pull origin dev
```

### Step 2: Feature branch banao

```bash
git checkout -b feature/auth-rbac
```

### Step 3: Code/docs change karo

Small focused changes karo. Ek branch = ek feature/fix.

### Step 4: Status check karo

```bash
git status
```

### Step 5: Files stage karo

```bash
git add backend/services/auth-service/internal/authorization/rbac.go
git add docs/14-git-workflow-repository-management.md
```

### Step 6: Commit karo

```bash
git commit -m "feat: added RBAC role checks"
```

### Step 7: Push branch

```bash
git push origin feature/auth-rbac
```

### Step 8: Pull Request banao

GitHub pe PR create karo: `feature/auth-rbac` → `dev`.

## 9. Common Git Mistakes & Fixes

| Mistake | Fix |
|---|---|
| `.env` accidentally staged | `git restore --staged backend/services/auth-service/.env` |
| Secret already committed | Secret rotate karo, file remove commit karo, history cleanup senior se karwao. |
| Wrong branch pe changes | `git stash`, correct branch checkout, `git stash pop` |
| Pull ke baad conflict | Conflict file edit karo, `git add`, then `git commit` |
| Detached HEAD | `git checkout dev` ya `git switch dev` |
| Force push mistake | Team ko immediately bolo, branch recover ke liye reflog/GitHub help lo. |

Useful commands:

```bash
git status
git log --oneline --decorate -5
git restore --staged <file>
git stash
git stash pop
```

## 10. Best Practices

- Small commits karo.
- Daily `git pull origin dev` lo.
- Direct `main` push avoid karo.
- `.env`, secrets, JWT keys GitHub pe mat push karo.
- Meaningful commit messages use karo.
- PR/MR review ke bina shared branch merge mat karo.
- Generated build folders commit mat karo.
- `go.mod`, `go.sum`, `go.work`, `go.work.sum` commit karo jab dependency/workspace change intentional ho.
- Migrations commit karo, local DB dump commit mat karo.
- Feature branch delete kar do jab PR merge ho jaye.
