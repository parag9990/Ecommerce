# Merge Search Service Work Into Dev (Partial)

Use this runbook to bring Search Service work from `feature/search-service`
into `dev` without merging the whole feature branch and without deleting
existing files from `dev`.

Do not run:

```bash
git merge feature/search-service
```

The full branch comparison contains hundreds of unrelated deletions. This
runbook restores only added and modified Search Service files, handles shared
files carefully, and never modifies `docs/01-micro-tasks.md`.

## Rules

- Add or update only Search Service related files from
  `feature/search-service`.
- Keep every file that exists on `dev` but is deleted or missing on
  `feature/search-service`.
- Do not apply deletion diffs.
- Do not restore, modify, stage, or commit `docs/01-micro-tasks.md`.
- Do not replace shared files with the feature branch version.
- Leave `api/master-api.json` and `backend/go.work.sum` unchanged. Their
  `dev..feature/search-service` differences come from newer `dev` work, not
  Search Service feature-side changes.
- Update `backend/go.work` by adding the Search Service module while preserving
  every existing `dev` workspace entry.
- Commit with this exact message:

```text
Merge Search Service work into dev
```

Current Search Service scope:

- `backend/services/search-service`
- `TaskImplementation/Search Service`
- `infra/k8s/core/search-service`
- `infra/k8s/data/typesense`
- `infra/k8s/jobs/search-reindex`

Current shared files that require manual or guarded handling:

- `backend/go.work`
- `infra/compose/.env.local.example`
- `infra/compose/docker-compose.local.yml`
- `infra/k8s/jobs/namespace.yaml`

## 1. Precheck

```bash
set -euo pipefail

git status --short --branch
git rev-parse --verify dev
git rev-parse --verify feature/search-service

test -z "$(git status --porcelain)" || {
  echo "Working tree is dirty. Commit or stash first."
  exit 1
}

if git show-ref --verify --quiet refs/heads/chore/Search-Service-selective; then
  echo "Integration branch already exists. Stop and inspect it."
  exit 1
fi
```

What this does:

- Confirms both required branches exist.
- Confirms the working tree is clean before changing branches.
- Prevents accidentally reusing an existing integration branch.

## 2. Create A Safe Integration Branch From Dev

```bash
git switch dev

test -z "$(git status --porcelain)" || {
  echo "dev is dirty. Stop."
  exit 1
}

git switch -c chore/Search-Service-selective

test "$(git rev-parse HEAD)" = "$(git rev-parse dev)"
```

What this does:

- Starts from the current `dev` tip.
- Creates a temporary branch for selective integration.
- Keeps `dev` unchanged until the final fast-forward step.

## 3. Review The Full Branch Difference

```bash
git rev-list --left-right --count dev...feature/search-service

git diff --name-status --no-renames dev..feature/search-service
git diff --name-status --no-renames --diff-filter=D dev..feature/search-service
git diff --name-status --no-renames --diff-filter=AM dev..feature/search-service

git diff --name-status --no-renames dev...feature/search-service
```

What this does:

- `dev..feature/search-service` compares the complete branch states and exposes
  all unrelated deletions that must not be applied.
- `dev...feature/search-service` shows feature-side changes since the merge
  base and helps confirm Search Service scope.
- `--no-renames` avoids misclassifying new Search Service files as renames of
  unrelated files that are missing from the feature branch.

Review the important shared differences:

```bash
git diff dev..feature/search-service -- \
  api/master-api.json \
  backend/go.work \
  backend/go.work.sum \
  docs/01-micro-tasks.md \
  infra/compose/.env.local.example \
  infra/compose/docker-compose.local.yml \
  infra/k8s/jobs/namespace.yaml

git diff --name-status --no-renames dev...feature/search-service -- \
  api/master-api.json \
  backend/go.work \
  backend/go.work.sum \
  docs/01-micro-tasks.md \
  infra/compose \
  infra/k8s/jobs/namespace.yaml
```

In the current repository state:

- `backend/go.work` removes existing `dev` workspace entries in the feature
  version, so it must be updated manually.
- `infra/compose/.env.local.example`,
  `infra/compose/docker-compose.local.yml`, and
  `infra/k8s/jobs/namespace.yaml` are feature-side additions in shared
  locations, so they require guarded restore and review.
- `api/master-api.json` and `backend/go.work.sum` differ only because of newer
  `dev` work. Do not touch them.
- `docs/01-micro-tasks.md` contains unrelated status changes and is explicitly
  excluded from this runbook.

## 4. Build The Auto-Restore File List

```bash
git diff --name-only --no-renames --diff-filter=AM dev..feature/search-service -- \
  backend/services/search-service \
  'TaskImplementation/Search Service' \
  infra/k8s/core/search-service \
  infra/k8s/data/typesense \
  infra/k8s/jobs/search-reindex \
  ':!docs/01-micro-tasks.md' \
  > /tmp/search-service-auto-restore.txt

sed -n '1,240p' /tmp/search-service-auto-restore.txt

if grep -Fxq 'docs/01-micro-tasks.md' /tmp/search-service-auto-restore.txt; then
  echo "Protected docs/01-micro-tasks.md entered the restore list. Stop."
  exit 1
fi
```

What this does:

- Builds an allowlist from added and modified files only.
- Includes Search Service code, implementation notes, and Search-owned
  Kubernetes resources.
- Excludes deletion diffs and `docs/01-micro-tasks.md`.
- Excludes shared files that need manual or guarded handling.

In the current repository state, this list contains 109 added Search Service
files. If that count or scope changes, inspect the list before continuing.

## 5. Restore Added And Modified Search Service Files

```bash
if [ -s /tmp/search-service-auto-restore.txt ]; then
  git restore --overlay \
    --source=feature/search-service \
    --worktree \
    --pathspec-from-file=/tmp/search-service-auto-restore.txt
else
  echo "No Search Service-owned files to auto-restore."
fi
```

What this does:

- Copies only allowlisted added and modified Search Service files.
- Uses `--overlay` so Git does not remove existing `dev` files.
- Does not restore any shared file or `docs/01-micro-tasks.md`.

## 6. Handle Shared Files Manually

### Update `backend/go.work` Safely

```bash
git diff dev..feature/search-service -- backend/go.work

(
  cd backend
  go work edit -use=./services/search-service
)

git diff -- backend/go.work
test "$(git diff --numstat -- backend/go.work | awk '{print $2}')" = "0" || {
  echo "backend/go.work lost existing lines. Stop."
  exit 1
}

git diff --quiet -- backend/go.work.sum || {
  echo "backend/go.work.sum changed unexpectedly. Stop."
  exit 1
}
```

What this does:

- Adds `./services/search-service` to the existing `dev` workspace.
- Preserves existing entries such as API Gateway, other services, generated
  modules, and shared validation.
- Stops if the `backend/go.work` diff removes any existing line.
- Refuses to continue if `backend/go.work.sum` changes unexpectedly.

The `backend/go.work` diff must add Search Service without removing any
existing `dev` entry.

### Restore Current Shared Infrastructure Additions With Guards

```bash
git diff --name-status --no-renames dev..feature/search-service -- \
  infra/compose/.env.local.example \
  infra/compose/docker-compose.local.yml \
  infra/k8s/jobs/namespace.yaml

test ! -e infra/compose/.env.local.example || {
  echo "infra/compose/.env.local.example now exists on dev. Merge it manually."
  exit 1
}

test ! -e infra/compose/docker-compose.local.yml || {
  echo "infra/compose/docker-compose.local.yml now exists on dev. Merge it manually."
  exit 1
}

test ! -e infra/k8s/jobs/namespace.yaml || {
  echo "infra/k8s/jobs/namespace.yaml now exists on dev. Merge it manually."
  exit 1
}

git restore --overlay --source=feature/search-service --worktree -- \
  infra/compose/.env.local.example \
  infra/compose/docker-compose.local.yml \
  infra/k8s/jobs/namespace.yaml

git diff -- \
  infra/compose/.env.local.example \
  infra/compose/docker-compose.local.yml \
  infra/k8s/jobs/namespace.yaml
```

What this does:

- Confirms these shared-path files are additions relative to the current
  `dev`.
- Stops instead of overwriting them if they have appeared on `dev`.
- Restores the current feature-side additions only after those guards pass.

If any guard fails, do not restore the whole file. Manually add only the Search
Service, Typesense, Redis, RabbitMQ, or search-reindex content needed from the
feature diff while preserving all existing `dev` content.

### Confirm Protected Shared Files Are Untouched

```bash
git diff --quiet -- api/master-api.json || {
  echo "api/master-api.json changed. Stop."
  exit 1
}

git diff --quiet -- backend/go.work.sum || {
  echo "backend/go.work.sum changed. Stop."
  exit 1
}

git diff --quiet -- docs/01-micro-tasks.md || {
  echo "docs/01-micro-tasks.md changed. Stop."
  exit 1
}
```

## 7. Confirm No Deletions Were Applied

```bash
git diff --name-status --diff-filter=D --exit-code
git diff --check
```

What this does:

- Confirms the working tree has no deleted files.
- Checks the restored and manually handled changes for whitespace errors.

The deletion command must print nothing. If it prints any `D` entry, stop
before staging.

## 8. Show Status And Diff Stat Before Staging

Mark new files as intent-to-add so the unstaged diff and stat include them:

```bash
if [ -s /tmp/search-service-auto-restore.txt ]; then
  git add -N --pathspec-from-file=/tmp/search-service-auto-restore.txt
fi

git add -N -- \
  infra/compose/.env.local.example \
  infra/compose/docker-compose.local.yml \
  infra/k8s/jobs/namespace.yaml

{
  cat /tmp/search-service-auto-restore.txt
  printf '%s\n' \
    backend/go.work \
    infra/compose/.env.local.example \
    infra/compose/docker-compose.local.yml \
    infra/k8s/jobs/namespace.yaml
} | LC_ALL=C sort -u > /tmp/search-service-allowed-files.txt

git diff --name-only | LC_ALL=C sort -u > /tmp/search-service-working-files.txt
diff -u /tmp/search-service-allowed-files.txt /tmp/search-service-working-files.txt

git status
git diff --stat
git diff --name-status
git diff --name-status --diff-filter=D --exit-code
git diff --quiet -- docs/01-micro-tasks.md
```

What this does:

- Makes new files visible in `git diff --stat` without staging their content.
- Confirms every working-tree change is on the Search Service allowlist.
- Shows status and diff details before staging.
- Confirms again that no deletion or protected micro-task change is present.

## 9. Stage Only The Selected Changes

```bash
if [ -s /tmp/search-service-auto-restore.txt ]; then
  git add --pathspec-from-file=/tmp/search-service-auto-restore.txt
fi

git add -- \
  backend/go.work \
  infra/compose/.env.local.example \
  infra/compose/docker-compose.local.yml \
  infra/k8s/jobs/namespace.yaml

git diff --cached --name-only | LC_ALL=C sort -u > /tmp/search-service-staged-files.txt
diff -u /tmp/search-service-allowed-files.txt /tmp/search-service-staged-files.txt

git status
git diff --cached --stat
git diff --cached --name-status
git diff --cached --name-status --diff-filter=D --exit-code
git diff --cached --check
git diff --cached --quiet -- docs/01-micro-tasks.md
git diff --exit-code
```

What this does:

- Stages only the generated Search Service allowlist and reviewed shared
  changes.
- Verifies the staged file list exactly matches the approved file list.
- Confirms no deletion and no `docs/01-micro-tasks.md` change is staged.
- Confirms no selected or unselected working-tree changes remain unstaged.

The staged deletion check and the protected-doc check must print nothing.

## 10. Commit

Run the final pre-commit checks, then commit with the exact required message:

```bash
git diff --cached --name-status --diff-filter=D --exit-code
git diff --cached --quiet -- docs/01-micro-tasks.md

git commit -m "Merge Search Service work into dev"
```

Verify the commit immediately:

```bash
test "$(git log -1 --format=%s)" = "Merge Search Service work into dev"

git diff-tree --no-commit-id --name-only -r HEAD \
  | LC_ALL=C sort -u > /tmp/search-service-committed-files.txt
diff -u /tmp/search-service-allowed-files.txt /tmp/search-service-committed-files.txt

test -z "$(git diff-tree --no-commit-id --name-only --diff-filter=D -r HEAD)"
test -z "$(git diff-tree --no-commit-id --name-only -r HEAD -- docs/01-micro-tasks.md)"

git status --short --branch
git show --stat --oneline HEAD
```

What this does:

- Creates the required selective integration commit.
- Confirms the exact commit message and exact approved file set.
- Confirms the commit contains no deletions and no
  `docs/01-micro-tasks.md` change.

## 11. Fast-Forward Dev

```bash
git switch dev
git merge --ff-only chore/Search-Service-selective

test "$(git rev-parse dev)" = "$(git rev-parse chore/Search-Service-selective)"
```

What this does:

- Moves `dev` to the reviewed selective integration commit.
- Does not merge `feature/search-service`.
- Fails safely if `dev` changed and cannot fast-forward.

## 12. Final Verification

```bash
test "$(git branch --show-current)" = "dev"
test "$(git log -1 --format=%s)" = "Merge Search Service work into dev"
test "$(git rev-parse dev)" = "$(git rev-parse chore/Search-Service-selective)"

test -z "$(git diff-tree --no-commit-id --name-only --diff-filter=D -r HEAD)"
test -z "$(git diff-tree --no-commit-id --name-only -r HEAD -- docs/01-micro-tasks.md)"

git status --short --branch
git log --oneline -1
git show --stat --oneline HEAD
git show --name-status --diff-filter=D --format= HEAD

(
  cd backend/services/search-service
  go test ./...
)
```

What this does:

- Confirms `dev` points to the selective integration commit.
- Confirms the latest commit message is exactly:

```text
Merge Search Service work into dev
```

- Confirms the committed change has no deletions and does not modify
  `docs/01-micro-tasks.md`.
- Shows the final committed file summary.
- Runs the Search Service test suite from the fast-forwarded `dev` branch.
