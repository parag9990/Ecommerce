# Merge Payment Service Work Into Dev

Use this runbook to bring Payment Service work from `feature/payment-service`
into `dev` without merging the whole feature branch and without deleting
existing files from `dev`.

Never merge `feature/payment-service` directly into `dev`.

This branch has unrelated deletions. We only apply added and modified Payment
Service related files, and we handle shared files carefully.

## Rules

- Add new Payment Service related files from `feature/payment-service`.
- If a file already exists on `dev`, update it only when it is modified in
  `feature/payment-service`.
- If a file exists on `dev` but is deleted or missing on
  `feature/payment-service`, keep the `dev` file.
- Do not apply deletion diffs.
- Update only the Payment Service related part in `docs/01-micro-tasks.md` so
  Payment Service tasks are marked correctly.
- Commit with this exact message:

```text
Merge Payment Service work into dev
```

## 1. Precheck

```bash
set -euo pipefail

git status --short --branch
git rev-parse --verify dev
git rev-parse --verify feature/payment-service
test -z "$(git status --porcelain)" || { echo "Working tree is dirty. Commit/stash first."; exit 1; }
```

What this does:

- Confirms both branches exist.
- Confirms the current working tree is clean before changing branches.

## 2. Create A Safe Integration Branch From Dev

```bash
git switch dev
test -z "$(git status --porcelain)" || { echo "dev is dirty. Stop."; exit 1; }

git switch -c chore/Payment-Service-selective
```

What this does:

- Starts from `dev`.
- Creates a temporary branch for the selective merge work.
- Keeps `dev` safe until the final fast-forward step.

## 3. Review The Full Branch Difference

```bash
git diff --name-status dev..feature/payment-service
git diff --name-status --diff-filter=D dev..feature/payment-service
git diff --name-status --no-renames --diff-filter=AM dev..feature/payment-service
```

What this does:

- Shows the complete difference from `dev` to `feature/payment-service`.
- Shows deletions separately. Do not apply those files.
- Shows only added and modified files. These are the only candidates for this
  runbook.
- Uses `--no-renames` for the added/modified review so
  `backend/services/payment-service/internal/transport/http/routes.go` is shown
  as an added Payment Service file instead of as a rename from another service.

Current important shared files to check:

```bash
git diff dev..feature/payment-service -- api/master-api.json
git diff dev..feature/payment-service -- backend/go.work
git diff dev..feature/payment-service -- backend/go.work.sum
git diff dev..feature/payment-service -- docs/01-micro-tasks.md
```

In the current repo state, the Payment Service branch updates
`api/master-api.json`, `backend/go.work`, `backend/go.work.sum`, and
`docs/01-micro-tasks.md`. These shared files need manual handling so existing
`dev` entries are not removed.

Also note that the branch contains many unrelated deletions for existing service
folders, shared generated code, proto files, reports, and docs. Those deletions
must not be applied.

## 4. Build The Auto-Restore File List

```bash
git diff --name-only --no-renames --diff-filter=AM dev..feature/payment-service -- \
  backend/services/payment-service \
  'TaskImplementation/Payment Service' \
  > /tmp/payment-service-auto-restore.txt

sed -n '1,240p' /tmp/payment-service-auto-restore.txt
```

What this does:

- Lists Payment Service related files that are added or modified.
- Excludes deleted files.
- Treats renamed files as add/delete pairs so only the new Payment Service path
  is selected.
- Excludes protected shared files that need special handling:
  `api/master-api.json`, `backend/go.work`, `backend/go.work.sum`, and
  `docs/01-micro-tasks.md`.

If this file is empty, that is okay. It means those files are already updated on
`dev`, and you should continue with the shared-file steps.

## 5. Restore Added And Modified Payment Service Files

```bash
if [ -s /tmp/payment-service-auto-restore.txt ]; then
  git restore --overlay --source=feature/payment-service --worktree --pathspec-from-file=/tmp/payment-service-auto-restore.txt
else
  echo "No Payment Service source/doc files to auto-restore."
fi
```

What this does:

- Copies selected new files from `feature/payment-service`.
- Updates selected existing files only when they differ from `dev`.
- Uses `--overlay` so Git does not remove dev files.
- Adds the Payment Service implementation under `backend/services/payment-service`
  and the Payment Service task docs under `TaskImplementation/Payment Service`.

## 6. Update `backend/go.work` Safely

```bash
(
  cd backend
  go work use ./services/payment-service
)

git diff -- backend/go.work
```

What this does:

- Adds the Payment Service workspace module.
- Keeps existing dev workspace entries such as `./services/api-gateway`,
  `./services/auth-service`, `./services/order-service`,
  `./services/product-service`, `./services/user-service`,
  `./shared/gen/go`, and `./shared/validation`.
- Avoids replacing `backend/go.work` with the feature branch version, because
  the feature version removes existing dev entries and keeps only
  `./services/payment-service`.

If `go` is not available, edit `backend/go.work` manually and add only this
entry inside the existing `use (...)` block:

```text
./services/payment-service
```

Then verify the diff again:

```bash
git diff -- backend/go.work
```

## 7. Update `backend/go.work.sum` Without Removing Existing Sums

```bash
git show dev:backend/go.work.sum > /tmp/dev-go.work.sum
git show feature/payment-service:backend/go.work.sum > /tmp/feature-payment-go.work.sum
LC_ALL=C sort -u /tmp/dev-go.work.sum /tmp/feature-payment-go.work.sum > backend/go.work.sum

git diff -- backend/go.work.sum
```

What this does:

- Keeps checksum lines that already exist on `dev`.
- Adds checksum lines introduced by `feature/payment-service`, if any exist.
- Prevents the feature branch from deleting existing `dev` checksum lines.

In the current repo state, the Payment Service branch mostly removes
`backend/go.work.sum` lines. If this step produces no diff, that is safe and
expected.

## 8. Update `api/master-api.json` Safely

First review the complete difference for the shared API catalog:

```bash
git diff dev..feature/payment-service -- api/master-api.json
```

Now apply only the Payment Service related changes using patch mode:

```bash
git restore -p --source=feature/payment-service --worktree -- api/master-api.json
```

When Git asks this question:

```text
Apply this hunk to worktree [y,n,q,a,d,s,e,p,?]?
```

Use these options carefully:

- Press `y` only for hunks that belong to Payment Service API/schema changes.
- Press `n` for hunks that remove or simplify Product Service, Inventory, or
  other unrelated API entries.
- Press `s` to split a large hunk into smaller hunks.
- Press `e` to manually edit a hunk if Payment Service changes and unrelated
  changes are mixed together.
- Press `q` to stop patch mode if the diff looks unsafe or confusing.

In the current repo state, the Payment Service related API catalog changes are
the retry payment schema changes, the `PaymentIntentResponse.retry` field, and
the `PaymentRetryMetadata` schema. Do not accept hunks that remove seller
product endpoints, remove `CategoryListRequest`, or simplify inventory request
validation.

After patch mode finishes, verify that only Payment Service API/schema changes
were applied:

```bash
git diff -- api/master-api.json
```

Do not use this command for the API catalog:

```bash
git restore --source=feature/payment-service --worktree -- api/master-api.json
```

That command can replace the whole shared file and remove existing `dev`
content. Use patch mode only.

If you already replaced the whole API catalog by mistake, reset that file and
apply only the Payment Service hunk again:

```bash
git restore --source=HEAD --worktree -- api/master-api.json
git restore -p --source=feature/payment-service --worktree -- api/master-api.json
git diff -- api/master-api.json
```

## 9. Update Only The Payment Service Part In `docs/01-micro-tasks.md`

First review the complete difference for the micro task file:

```bash
git diff dev..feature/payment-service -- docs/01-micro-tasks.md
```

Now apply only the Payment Service related change using patch mode:

```bash
git restore -p --source=feature/payment-service --worktree -- docs/01-micro-tasks.md
```

When Git asks this question:

```text
Apply this hunk to worktree [y,n,q,a,d,s,e,p,?]?
```

Use these options carefully:

- Press `y` only for hunks that belong to the Payment Service section.
- Press `n` for unrelated hunks from other services or unrelated documentation.
- Press `s` to split a large hunk into smaller hunks.
- Press `e` to manually edit a hunk if Payment Service changes and unrelated
  changes are mixed together.
- Press `q` to stop patch mode if the diff looks unsafe or confusing.

After patch mode finishes, verify that only the Payment Service part changed:

```bash
git diff -- docs/01-micro-tasks.md
```

What this does:

- Shows the micro task file changes before applying them.
- Updates only the Payment Service task statuses, for example from `Pending` to
  `Completed`.
- Prevents the whole `docs/01-micro-tasks.md` file from being replaced by the
  feature branch version.
- Keeps unrelated `dev` content in the same file unchanged.

In the current repo state, the feature branch also changes User Service, Auth
Service, Product Service, and Order Service task statuses from `Completed` back
to `Pending`. Reject those unrelated status changes.

Do not use this command for the micro task file:

```bash
git restore --source=feature/payment-service --worktree -- docs/01-micro-tasks.md
```

That command can replace the whole file. Use patch mode only.

If you already replaced the whole micro task file by mistake, reset that file
and apply only the Payment Service hunk again:

```bash
git restore --source=HEAD --worktree -- docs/01-micro-tasks.md
git restore -p --source=feature/payment-service --worktree -- docs/01-micro-tasks.md
git diff -- docs/01-micro-tasks.md
```

## 10. Confirm No Deletions Were Applied

```bash
git diff --name-status --diff-filter=D --exit-code
```

What this does:

- Confirms the working tree has no deletions.
- This command must print nothing.

If it prints any `D` entry, stop before staging.

## 11. Show Status And Diff Stat Before Staging

```bash
if [ -s /tmp/payment-service-auto-restore.txt ]; then
  git add -N --pathspec-from-file=/tmp/payment-service-auto-restore.txt
fi

git status
git diff --stat
git diff --name-status --diff-filter=D --exit-code
```

What this does:

- Marks new auto-restored files as intent-to-add so `git diff --stat` shows
  them.
- Shows the required `git status`.
- Shows the required `git diff --stat`.
- Confirms again that no deletion is present.

## 12. Stage Only The Selected Changes

```bash
if [ -s /tmp/payment-service-auto-restore.txt ]; then
  git add --pathspec-from-file=/tmp/payment-service-auto-restore.txt
fi

git add backend/go.work backend/go.work.sum
git add -p api/master-api.json
git add -p docs/01-micro-tasks.md

git status
git diff --cached --stat
git diff --cached --name-status --diff-filter=D --exit-code
```

What this does:

- Stages selected Payment Service added and modified files.
- Stages the reviewed `backend/go.work` and `backend/go.work.sum` files.
- Stages only the reviewed Payment Service related hunks from
  `api/master-api.json` and `docs/01-micro-tasks.md`.
- Confirms no deletions are staged.

The staged deletion check must print nothing.

## 13. Commit

```bash
git commit -m "Merge Payment Service work into dev"

git status --short --branch
git log --oneline -1
git show --stat --oneline HEAD
git show --name-status --format= --diff-filter=D HEAD
test -z "$(git show --name-only --format= --diff-filter=D HEAD)" || { echo "Commit contains deletions. Stop."; exit 1; }
test "$(git log -1 --pretty=%B)" = "Merge Payment Service work into dev" || { echo "Unexpected commit message. Stop."; exit 1; }
```

What this does:

- Creates the required selective integration commit.
- Verifies the latest commit and its file summary.
- Confirms the commit contains no deletion entries.
- Confirms the latest commit message is exactly:

```text
Merge Payment Service work into dev
```

## 14. Fast-Forward Dev

```bash
git switch dev
git merge --ff-only chore/Payment-Service-selective
```

What this does:

- Moves `dev` to the selective integration commit.
- Does not merge `feature/payment-service`.
- Fails safely if `dev` changed and cannot fast-forward.

## 15. Final Verification

```bash
git status --short --branch
git log --oneline -1
git show --stat --oneline HEAD
git show --name-status --format= --diff-filter=D HEAD
test -z "$(git show --name-only --format= --diff-filter=D HEAD)" || { echo "Final dev commit contains deletions. Stop."; exit 1; }
test "$(git log -1 --pretty=%B)" = "Merge Payment Service work into dev" || { echo "Unexpected final commit message. Stop."; exit 1; }
```

What this does:

- Confirms `dev` is on the new commit.
- Shows the final committed file summary.
- Confirms the final commit contains no deletion entries.
- Confirms the latest commit message is:

```text
Merge Payment Service work into dev
```
