# Merge Product Service Work Into Dev

Use this runbook to bring Product Service work from `feature/product-service`
into `dev` without direct-merging the feature branch and without deleting
existing files from `dev`.

This branch has many unrelated deletions. Only apply added and modified Product
Service related files, and handle shared files carefully.

## Rules

- Add new Product Service related files from `feature/product-service`.
- If a file already exists on `dev`, update it only when it is modified in
  `feature/product-service`.
- If a file exists on `dev` but is deleted or missing on
  `feature/product-service`, keep the `dev` file.
- Do not apply deletion diffs.
- Do not direct-merge `feature/product-service`.
- Update only the Product Service related part in `docs/01-micro-tasks.md` so
  Product Service tasks are marked correctly.
- Commit with this exact message:

```text
Merge Product Service work into dev
```

## 1. Precheck

```bash
set -euo pipefail

git status --short --branch
git rev-parse --verify dev
git rev-parse --verify feature/product-service
test -z "$(git status --porcelain)" || { echo "Working tree is dirty. Commit/stash first."; exit 1; }
```

What this does:

- Confirms both branches exist.
- Confirms the current working tree is clean before changing branches.

## 2. Create Safe Integration Branch From `dev`

Keep quotes around the integration branch name because it contains a space.

```bash
git switch dev
test -z "$(git status --porcelain)" || { echo "dev is dirty. Stop."; exit 1; }

git switch -c 'chore/Product Service-selective'
```

What this does:

- Starts from `dev`.
- Creates a temporary branch for the selective merge work.
- Keeps `dev` safe until the final fast-forward step.

## 3. Review Full Branch Difference

```bash
git diff --name-status dev..feature/product-service
git diff --name-status --diff-filter=D dev..feature/product-service
git diff --name-status --diff-filter=AM dev..feature/product-service
```

What this does:

- Shows the complete difference from `dev` to `feature/product-service`.
- Shows deletions separately. Do not apply those files.
- Shows only added and modified files. These are the only candidates for this
  runbook.

Current important shared files to check:

```bash
git diff dev..feature/product-service -- api/master-api.json
git diff dev..feature/product-service -- backend/go.work
git diff dev..feature/product-service -- backend/go.work.sum
git diff dev..feature/product-service -- docs/01-micro-tasks.md
```

In the current repo state, the Product Service branch adds
`backend/services/product-service` and `TaskImplementation/Product Service`.
It also modifies `api/master-api.json`, `backend/go.work`,
`backend/go.work.sum`, and `docs/01-micro-tasks.md`.

The branch also deletes unrelated Auth Service, User Service, API Gateway,
shared generated code, proto, report, and runbook files. Those deletion entries
must remain unapplied.

## 4. Build Auto-Restore File List

```bash
git diff --name-only --diff-filter=AM dev..feature/product-service -- \
  backend/services/product-service \
  'TaskImplementation/Product Service' \
  > /tmp/product-service-auto-restore.txt

sed -n '1,240p' /tmp/product-service-auto-restore.txt
```

What this does:

- Lists Product Service files that are added or modified.
- Excludes deleted files.
- Excludes protected shared files that need manual handling:
  `api/master-api.json`, `backend/go.work`, `backend/go.work.sum`, and
  `docs/01-micro-tasks.md`.

In the current repo state, this list should contain Product Service task docs
and backend service files only.

If this file is empty, that is okay. It means those files are already updated on
`dev`, and you should continue with the shared-file steps.

## 5. Restore Added And Modified Product Service Files Using `git restore --overlay`

```bash
if [ -s /tmp/product-service-auto-restore.txt ]; then
  git restore --overlay --source=feature/product-service --worktree --pathspec-from-file=/tmp/product-service-auto-restore.txt
else
  echo "No Product Service source/doc files to auto-restore."
fi
```

What this does:

- Copies selected new files from `feature/product-service`.
- Updates selected existing files only when they differ from `dev`.
- Uses `--overlay` so Git does not remove dev files.

## 6. Handle Shared Files Manually

### Update `api/master-api.json` Safely

First review the complete shared API diff:

```bash
git diff dev..feature/product-service -- api/master-api.json
```

Then apply only Product Service API changes using patch mode:

```bash
git restore -p --source=feature/product-service --worktree -- api/master-api.json
```

When Git asks this question:

```text
Apply this hunk to worktree [y,n,q,a,d,s,e,p,?]?
```

Use these options carefully:

- Press `y` only for hunks that belong to Product Service API behavior.
- Press `n` for unrelated shared API changes.
- Press `s` to split a large hunk into smaller hunks.
- Press `e` to manually edit a hunk if Product Service changes and unrelated
  changes are mixed together.
- Press `q` to stop patch mode if the diff looks unsafe or confusing.

In the current repo state, the expected Product Service API changes include:

- Seller product list and detail routes for `product-service`.
- `ProductService.ListCategories`.
- `ProductListRequest.sort`.
- `CategoryListRequest`.
- Inventory reservation item details and `idempotency_key`.

After patch mode finishes, verify the result:

```bash
git diff -- api/master-api.json
```

### Update `backend/go.work` Safely

```bash
(
  cd backend
  go work use ./services/product-service
)

git diff -- backend/go.work
```

What this does:

- Adds the Product Service workspace module.
- Keeps existing `dev` workspace entries such as `./services/api-gateway`,
  `./services/auth-service`, `./services/user-service`, `./shared/gen/go`, and
  `./shared/validation`.
- Avoids replacing `backend/go.work` with the feature branch version, because
  the feature version only contains `./services/product-service`.

### Update `backend/go.work.sum` Without Removing Existing Sums

```bash
git show dev:backend/go.work.sum > /tmp/dev-go.work.sum
git show feature/product-service:backend/go.work.sum > /tmp/feature-product-go.work.sum
LC_ALL=C sort -u /tmp/dev-go.work.sum /tmp/feature-product-go.work.sum > backend/go.work.sum

git diff -- backend/go.work.sum
```

What this does:

- Keeps checksum lines that already exist on `dev`.
- Adds checksum lines introduced by `feature/product-service`.
- Prevents the feature branch from deleting existing `dev` checksum lines.

## 7. Update Micro-Task Documentation Safely

First verify the file exists on both branches:

```bash
git cat-file -e dev:docs/01-micro-tasks.md
git cat-file -e feature/product-service:docs/01-micro-tasks.md
```

Now review the complete difference for the micro-task file:

```bash
git diff dev..feature/product-service -- docs/01-micro-tasks.md
```

Apply only the Product Service related change using patch mode:

```bash
git restore -p --source=feature/product-service --worktree -- docs/01-micro-tasks.md
```

When Git asks this question:

```text
Apply this hunk to worktree [y,n,q,a,d,s,e,p,?]?
```

Use these options carefully:

- Press `y` only for hunks in the Product Service section.
- Press `n` for Auth Service status reversions or other unrelated sections.
- Press `n` for blank-line-only cleanup unless it is clearly required for the
  Product Service section.
- Press `s` to split a large hunk into smaller hunks.
- Press `e` to manually edit a hunk if Product Service changes and unrelated
  changes are mixed together.
- Press `q` to stop patch mode if the diff looks unsafe or confusing.

After patch mode finishes, verify that only the Product Service part changed:

```bash
git diff -- docs/01-micro-tasks.md
```

What this does:

- Updates only the Product Service task statuses, for example from `Pending` to
  `Completed`.
- Prevents the whole `docs/01-micro-tasks.md` file from being replaced by the
  feature branch version.
- Keeps unrelated `dev` content in the same file unchanged.

If you already replaced the whole micro-task file by mistake, reset that file
and apply only the Product Service hunk again:

```bash
git restore --source=HEAD --worktree -- docs/01-micro-tasks.md
git restore -p --source=feature/product-service --worktree -- docs/01-micro-tasks.md
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
if [ -s /tmp/product-service-auto-restore.txt ]; then
  git add -N --pathspec-from-file=/tmp/product-service-auto-restore.txt
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

## 10. Stage Only Selected Changes

```bash
if [ -s /tmp/product-service-auto-restore.txt ]; then
  git add --pathspec-from-file=/tmp/product-service-auto-restore.txt
fi

git add backend/go.work backend/go.work.sum
git add -p api/master-api.json
git add -p docs/01-micro-tasks.md

git status
git diff --cached --stat
git diff --cached --name-status --diff-filter=D --exit-code
```

What this does:

- Stages selected Product Service added and modified files.
- Stages the reviewed `backend/go.work` and `backend/go.work.sum` files.
- Stages only reviewed Product Service related hunks from `api/master-api.json`
  and `docs/01-micro-tasks.md`.
- Confirms no deletions are staged.

The staged deletion check must print nothing.

## 11. Commit With The Exact Commit Message

```bash
git commit -m "Merge Product Service work into dev"
```

After the commit, verify the commit before touching `dev`:

```bash
git status --short --branch
git log --oneline -1
git show --stat --oneline HEAD
git diff-tree --no-commit-id --name-status -r --diff-filter=D --exit-code HEAD
```

What this does:

- Creates the required selective integration commit.
- Confirms the latest commit message is correct.
- Confirms the commit did not introduce deleted files.

## 12. Fast-Forward `dev`

```bash
git switch dev
git merge --ff-only 'chore/Product Service-selective'
```

What this does:

- Moves `dev` to the selective integration commit.
- Does not merge `feature/product-service`.
- Fails safely if `dev` changed and cannot fast-forward.

## 13. Final Verification

```bash
git status --short --branch
git log --oneline -1
git show --stat --oneline HEAD
git diff-tree --no-commit-id --name-status -r --diff-filter=D --exit-code HEAD
```

What this does:

- Confirms `dev` is on the new commit.
- Shows the final committed file summary.
- Confirms no deletion entries were introduced.
- Confirms the latest commit message is:

```text
Merge Product Service work into dev
```
