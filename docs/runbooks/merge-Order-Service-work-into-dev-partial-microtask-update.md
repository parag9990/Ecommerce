# Merge Order Service Work Into Dev

Use this runbook to bring Order Service work from `feature/order-service` into
`dev` without merging the whole feature branch and without deleting existing
files from `dev`.

No direct source-branch merge is part of this runbook.

This branch has many unrelated deletions. We only apply added and modified Order
Service related files, treat Git rename targets as new Order Service files, and
handle shared files carefully.

## Rules

- Add new Order Service related files from `feature/order-service`.
- If a file already exists on `dev`, update it only when it is modified in
  `feature/order-service` and belongs to Order Service.
- If a file exists on `dev` but is deleted or missing on `feature/order-service`,
  keep the `dev` file.
- Do not apply deletion diffs.
- For rename-looking diffs, restore only the target Order Service path. Do not
  stage the old source path deletion.
- Update only the Order Service related part in `docs/01-micro-tasks.md`, if
  that file exists.
- Commit with this exact message:

```text
Merge Order Service work into dev
```

## 1. Precheck

```bash
set -euo pipefail

git status --short --branch
git rev-parse --verify dev
git rev-parse --verify feature/order-service
test -z "$(git status --porcelain)" || { echo "Working tree is dirty. Commit/stash first."; exit 1; }

if git show-ref --verify --quiet refs/heads/chore/Order-Service-selective; then
  echo "Integration branch already exists. Delete or rename it before continuing."
  exit 1
fi
```

What this does:

- Confirms both branches exist.
- Confirms the current working tree is clean before changing branches.
- Avoids accidentally reusing an old integration branch.

## 2. Create A Safe Integration Branch From Dev

```bash
git switch dev
test -z "$(git status --porcelain)" || { echo "dev is dirty. Stop."; exit 1; }

git switch -c chore/Order-Service-selective
```

What this does:

- Starts from `dev`.
- Creates a temporary branch for the selective merge work.
- Keeps `dev` safe until the final fast-forward step.

## 3. Review The Full Branch Difference

```bash
git diff --name-status dev..feature/order-service
git diff --name-status --diff-filter=D dev..feature/order-service
git diff --name-status --diff-filter=AMR dev..feature/order-service
```

What this does:

- Shows the complete difference from `dev` to `feature/order-service`.
- Shows deletions separately. Do not apply those files.
- Shows added, modified, and rename target candidates. For this runbook, rename
  targets under Order Service paths are treated as new Order Service files, but
  the old source path deletion must not be staged.

Current important Order Service paths to check:

```bash
git diff --name-status --diff-filter=AMR dev..feature/order-service -- \
  backend/services/order-service \
  'TaskImplementation/Order Service' \
  proto/ecommerce/order/v1 \
  backend/shared/gen/go/ecommerce/order/v1
```

Current important shared files to check:

```bash
git diff dev..feature/order-service -- backend/go.work
git diff dev..feature/order-service -- backend/go.work.sum
git diff dev..feature/order-service -- backend/shared/gen/go/go.mod
git diff dev..feature/order-service -- backend/shared/gen/go/go.sum
git diff dev..feature/order-service -- proto/buf.gen.yaml
git diff dev..feature/order-service -- api/master-api.json
git diff dev..feature/order-service -- docs/01-micro-tasks.md
```

In the current repo state, the Order Service branch adds Order Service backend
files, Order Service task documentation, Order protobuf files, and generated
Order protobuf Go files.

The full branch diff also deletes unrelated Auth Service, Product Service, User
Service, API Gateway, shared validation, reports, and other files. Those
deletions must not be applied.

The shared files need manual handling:

- `backend/go.work` removes existing `dev` workspace entries if restored
  wholesale.
- `backend/go.work.sum` removes existing `dev` checksum lines if restored
  wholesale.
- `backend/shared/gen/go/go.mod` changes the shared generated module path and
  dependency versions globally.
- `backend/shared/gen/go/go.sum` changes shared checksum lines globally.
- `proto/buf.gen.yaml` switches protobuf generation from remote plugins to local
  plugins globally.
- `api/master-api.json` contains non-Order API/schema edits in the current diff.
- `docs/01-micro-tasks.md` marks Order Service complete but also regresses other
  service statuses if restored wholesale.

## 4. Build The Auto-Restore File List

```bash
git diff --name-only --diff-filter=AMR dev..feature/order-service -- \
  backend/services/order-service \
  'TaskImplementation/Order Service' \
  proto/ecommerce/order/v1 \
  backend/shared/gen/go/ecommerce/order/v1 \
  > /tmp/order-service-auto-restore.txt

LC_ALL=C sort -u /tmp/order-service-auto-restore.txt -o /tmp/order-service-auto-restore.txt
sed -n '1,260p' /tmp/order-service-auto-restore.txt

grep -Ev '^(backend/services/order-service/|TaskImplementation/Order Service/|proto/ecommerce/order/v1/|backend/shared/gen/go/ecommerce/order/v1/)' /tmp/order-service-auto-restore.txt \
  && { echo "Unexpected path in auto-restore list. Stop."; exit 1; } || true
```

What this does:

- Lists Order Service related files that are added, modified, or reported as
  rename targets.
- Excludes deleted files.
- Excludes protected shared files that need manual handling.
- Confirms the list contains only Order Service backend files, Order Service task
  docs, Order protobuf files, and generated Order protobuf Go files.

If this file is empty, that is okay. It means those files are already updated on
`dev`, and you should continue with the shared-file steps.

## 5. Restore Added And Modified Order Service Files

```bash
if [ -s /tmp/order-service-auto-restore.txt ]; then
  git restore --overlay --source=feature/order-service --worktree --pathspec-from-file=/tmp/order-service-auto-restore.txt
else
  echo "No Order Service source/doc/proto files to auto-restore."
fi
```

What this does:

- Copies selected new Order Service files from `feature/order-service`.
- Updates selected existing files only when they differ from `dev`.
- Uses `--overlay` so Git does not remove `dev` files.
- Restores rename target files only by their Order Service target paths.

## 6. Handle Shared Files Manually

First review the shared-file diffs again:

```bash
git diff dev..feature/order-service -- \
  backend/go.work \
  backend/go.work.sum \
  backend/shared/gen/go/go.mod \
  backend/shared/gen/go/go.sum \
  proto/buf.gen.yaml \
  api/master-api.json
```

### Update `backend/go.work` Safely

```bash
(
  cd backend
  go work use ./services/order-service
)

git diff -- backend/go.work
```

What this does:

- Adds the Order Service workspace module.
- Keeps existing `dev` workspace entries such as `./services/api-gateway`,
  `./services/auth-service`, `./services/product-service`,
  `./services/user-service`, `./shared/gen/go`, and `./shared/validation`.
- Avoids replacing `backend/go.work` with the feature branch version, because
  the feature version removes existing `dev` entries.

### Update `backend/go.work.sum` Without Removing Existing Sums

```bash
git show dev:backend/go.work.sum > /tmp/dev-go.work.sum
git show feature/order-service:backend/go.work.sum > /tmp/feature-order-go.work.sum
LC_ALL=C sort -u /tmp/dev-go.work.sum /tmp/feature-order-go.work.sum > backend/go.work.sum

git diff -- backend/go.work.sum
```

What this does:

- Keeps checksum lines that already exist on `dev`.
- Adds checksum lines introduced by `feature/order-service`.
- Prevents the feature branch from deleting existing `dev` checksum lines.

### Review Shared Generated Go Module Files

```bash
git diff dev..feature/order-service -- backend/shared/gen/go/go.mod backend/shared/gen/go/go.sum
```

In the current repo state, `backend/shared/gen/go/go.mod` changes the module path
from the existing `dev` module to the feature branch module path. Do not restore
that file wholesale.

If the generated Order protobuf files require shared generated module dependency
updates, apply only those lines manually:

```bash
git restore -p --source=feature/order-service --worktree -- \
  backend/shared/gen/go/go.mod \
  backend/shared/gen/go/go.sum
```

When Git asks this question:

```text
Apply this hunk to worktree [y,n,q,a,d,s,e,p,?]?
```

Use these options carefully:

- Press `y` only for hunks clearly required by the Order generated code.
- Press `n` for module path changes that replace the existing `dev` module path.
- Press `n` for hunks that remove existing `dev` dependency or checksum lines.
- Press `s` to split a large hunk into smaller hunks.
- Press `e` to manually edit a hunk if required Order additions and unrelated
  changes are mixed together.
- Press `q` if the diff looks unsafe.

After patch mode finishes:

```bash
git diff -- backend/shared/gen/go/go.mod backend/shared/gen/go/go.sum
```

If no Order-specific dependency update is needed, leave these files unchanged.

### Review `proto/buf.gen.yaml`

```bash
git diff dev..feature/order-service -- proto/buf.gen.yaml
```

In the current repo state, this file changes protobuf generation globally from
remote plugins to local plugins. That is not an Order Service-only change. Leave
it unchanged unless you intentionally want that global generator change and have
verified it with the rest of `dev`.

If you must apply a small safe hunk:

```bash
git restore -p --source=feature/order-service --worktree -- proto/buf.gen.yaml
git diff -- proto/buf.gen.yaml
```

### Review `api/master-api.json`

```bash
git diff dev..feature/order-service -- api/master-api.json
```

In the current repo state, this file contains non-Order API/schema edits. Do not
restore it wholesale. Apply only explicit Order Service route, schema, or RPC
hunks if review finds any:

```bash
git restore -p --source=feature/order-service --worktree -- api/master-api.json
git diff -- api/master-api.json
```

If the patch mode only shows product, category, inventory, or unrelated schema
changes, press `n` for those hunks and leave `api/master-api.json` unchanged.

## 7. Update Micro-Task Documentation Safely

First confirm the file exists on both branches and review the complete diff:

```bash
if git cat-file -e dev:docs/01-micro-tasks.md && git cat-file -e feature/order-service:docs/01-micro-tasks.md; then
  git diff dev..feature/order-service -- docs/01-micro-tasks.md
else
  echo "docs/01-micro-tasks.md is absent on one side; skip this step."
fi
```

Now apply only the Order Service related status changes using patch mode:

```bash
if git cat-file -e dev:docs/01-micro-tasks.md && git cat-file -e feature/order-service:docs/01-micro-tasks.md; then
  git restore -p --source=feature/order-service --worktree -- docs/01-micro-tasks.md
fi
```

When Git asks this question:

```text
Apply this hunk to worktree [y,n,q,a,d,s,e,p,?]?
```

Use these options carefully:

- Press `y` only for hunks that belong to the Order Service section.
- Press `n` for User Service, Auth Service, Product Service, or unrelated
  documentation hunks.
- Press `s` to split a large hunk into smaller hunks.
- Press `e` to manually edit a hunk if Order Service changes and unrelated
  changes are mixed together.
- Press `q` to stop patch mode if the diff looks unsafe or confusing.

In the current repo state, the safe documentation change is the Order Service
task status update from `Pending` to `Completed`. Do not apply the hunks that
change User Service, Auth Service, or Product Service statuses back to
`Pending`.

After patch mode finishes, verify that only the Order Service part changed:

```bash
git diff -- docs/01-micro-tasks.md
```

Do not restore the whole micro-task file from the feature branch. That can
replace unrelated `dev` content.

If you already replaced the whole micro-task file by mistake, reset that file
and apply only the Order Service hunk again:

```bash
git restore --source=HEAD --worktree -- docs/01-micro-tasks.md
git restore -p --source=feature/order-service --worktree -- docs/01-micro-tasks.md
git diff -- docs/01-micro-tasks.md
```

## 8. Confirm No Deletions Were Applied

```bash
git diff --name-status --diff-filter=D --exit-code
```

What this does:

- Confirms the working tree has no deletions.
- This command must print nothing.

If it prints any `D` entry, stop before staging.

## 9. Show Status And Diff Stat Before Staging

```bash
if [ -s /tmp/order-service-auto-restore.txt ]; then
  git add -N --pathspec-from-file=/tmp/order-service-auto-restore.txt
fi

git status
git diff --stat
git diff --name-status --diff-filter=D --exit-code
git diff -- backend/go.work docs/01-micro-tasks.md api/master-api.json proto/buf.gen.yaml
```

What this does:

- Marks new auto-restored files as intent-to-add so `git diff --stat` shows
  them.
- Shows the required `git status`.
- Shows the required `git diff --stat`.
- Confirms again that no deletion is present.
- Re-checks high-risk shared files before anything is staged.

## 10. Stage Only The Selected Changes

```bash
if [ -s /tmp/order-service-auto-restore.txt ]; then
  git add --pathspec-from-file=/tmp/order-service-auto-restore.txt
fi

git add backend/go.work backend/go.work.sum

git add -p \
  backend/shared/gen/go/go.mod \
  backend/shared/gen/go/go.sum \
  proto/buf.gen.yaml \
  api/master-api.json \
  docs/01-micro-tasks.md

git status
git diff --cached --stat
git diff --cached --name-status
git diff --cached --name-status --diff-filter=D --exit-code
```

What this does:

- Stages selected Order Service added and modified files.
- Stages the reviewed `backend/go.work` and `backend/go.work.sum` files.
- Uses patch staging for shared files so only reviewed hunks are staged.
- Confirms no deletions are staged.

The staged deletion check must print nothing.

Before committing, inspect the staged shared-file changes one more time:

```bash
git diff --cached -- \
  backend/go.work \
  backend/go.work.sum \
  backend/shared/gen/go/go.mod \
  backend/shared/gen/go/go.sum \
  proto/buf.gen.yaml \
  api/master-api.json \
  docs/01-micro-tasks.md
```

## 11. Commit With The Exact Message

```bash
git commit -m "Merge Order Service work into dev"
```

What this does:

- Creates the required selective integration commit.

## 12. Verify After Commit

```bash
git status --short --branch
git log --oneline -1
git show --stat --oneline HEAD
test "$(git log -1 --pretty=%s)" = "Merge Order Service work into dev"
test -z "$(git diff-tree --no-commit-id --name-only --diff-filter=D -r HEAD)" || { echo "Commit contains deletions. Stop."; exit 1; }
```

What this does:

- Confirms the integration branch is clean after the commit.
- Confirms the latest commit has the required message.
- Shows the committed file summary.
- Confirms the commit contains no deletions.

## 13. Fast-Forward Dev

```bash
git switch dev
git merge --ff-only chore/Order-Service-selective
```

What this does:

- Moves `dev` to the selective integration commit.
- Does not merge `feature/order-service`.
- Fails safely if `dev` changed and cannot fast-forward.

## 14. Verify After Fast-Forwarding Dev

```bash
git status --short --branch
git log --oneline -1
git show --stat --oneline HEAD
test "$(git log -1 --pretty=%s)" = "Merge Order Service work into dev"
test -z "$(git diff-tree --no-commit-id --name-only --diff-filter=D -r HEAD)" || { echo "Fast-forwarded commit contains deletions. Stop."; exit 1; }
```

What this does:

- Confirms `dev` points at the selective integration commit.
- Confirms the latest commit message is still exact.
- Confirms no deletion slipped into the fast-forwarded commit.

## 15. Final Verification

```bash
git status --short --branch
git log --oneline -1
git show --stat --oneline HEAD
git diff --name-status HEAD^..HEAD --diff-filter=D --exit-code

git ls-tree -r --name-only HEAD -- \
  backend/services/order-service \
  'TaskImplementation/Order Service' \
  proto/ecommerce/order/v1 \
  backend/shared/gen/go/ecommerce/order/v1 \
  | sed -n '1,240p'

(
  cd backend/services/order-service
  go test ./...
)
```

What this does:

- Confirms `dev` is clean after the fast-forward.
- Shows the final committed file summary.
- Confirms the final commit contains no deletions.
- Confirms the expected Order Service paths exist in `HEAD`.
- Runs the Order Service Go test suite.

The latest commit message must be:

```text
Merge Order Service work into dev
```
