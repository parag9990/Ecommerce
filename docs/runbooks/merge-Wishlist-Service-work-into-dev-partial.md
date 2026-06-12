# Merge Wishlist Service Work Into Dev

Use this runbook to bring Wishlist Service work from
`feature/wishlist-service` into `dev` without merging the whole feature branch,
without deleting existing files from `dev`, and without modifying
`docs/01-micro-tasks.md`.

Do not run:

```bash
git merge feature/wishlist-service
```

This branch has many unrelated deletions. We only apply added and modified
Wishlist Service related files, and we handle shared files carefully.

## Rules

- Add new Wishlist Service related files from `feature/wishlist-service`.
- If a Wishlist Service file already exists on `dev`, update it only when it is
  modified in `feature/wishlist-service`.
- If a file exists on `dev` but is deleted or missing on
  `feature/wishlist-service`, keep the `dev` file.
- Do not apply deletion diffs.
- Do not restore, stage, or commit `docs/01-micro-tasks.md`.
- Do not use blanket commands such as `git restore .`, `git add .`, or
  `git add -A`.
- Commit with this exact message:

```text
Merge Wishlist Service work into dev
```

## 1. Precheck

```bash
set -euo pipefail

git status --short --branch
git rev-parse --verify dev
git rev-parse --verify feature/wishlist-service
test -z "$(git status --porcelain)" || { echo "Working tree is dirty. Commit/stash first."; exit 1; }
```

What this does:

- Confirms both branches exist.
- Confirms the current working tree and index are clean before changing
  branches.
- Prevents existing local changes, including changes to
  `docs/01-micro-tasks.md`, from being mixed into the selective integration.

## 2. Create A Safe Integration Branch From Dev

```bash
git switch dev
test -z "$(git status --porcelain)" || { echo "dev is dirty. Stop."; exit 1; }

git switch -c chore/Wishlist-Service-selective
```

What this does:

- Starts from `dev`.
- Creates a temporary branch for the selective Wishlist Service work.
- Keeps `dev` safe until the final fast-forward step.

## 3. Review The Full Branch Difference

```bash
git diff --name-status --no-renames dev..feature/wishlist-service
git diff --name-status --no-renames --diff-filter=D dev..feature/wishlist-service
git diff --name-status --no-renames --diff-filter=AM dev..feature/wishlist-service
git diff --name-status --no-renames dev...feature/wishlist-service
```

What this does:

- Shows the complete endpoint difference from `dev` to
  `feature/wishlist-service`.
- Shows deletions separately. Do not apply those files.
- Shows added and modified files separately.
- Shows changes introduced on the feature side since the common ancestor.
- Uses `--no-renames` because Git currently detects Wishlist Service
  `go.mod` and `go.sum` as renames from unrelated Product Service files.
  Treating them as added files prevents them from being omitted.

Current important shared and protected files to check:

```bash
git diff dev..feature/wishlist-service -- api/master-api.json
git diff dev..feature/wishlist-service -- backend/go.work
git diff dev..feature/wishlist-service -- backend/go.work.sum
git diff dev..feature/wishlist-service -- docs/01-micro-tasks.md
git diff --name-status dev...feature/wishlist-service -- api/master-api.json
```

In the current repo state:

- Wishlist Service work is under `backend/services/wishlist-service` and
  `TaskImplementation/Wishlist Service`.
- There are currently 67 added Wishlist Service files in those two paths.
- The branch difference contains many unrelated deletions across existing
  services, shared modules, task documents, proto files, and reports. Do not
  apply any of them.
- `backend/go.work` and `backend/go.work.sum` need manual handling because the
  feature versions remove existing `dev` content.
- `api/master-api.json` appears in the endpoint difference because `dev`
  changed after the branches diverged. It has no feature-side change in
  `dev...feature/wishlist-service`, so leave it untouched.
- `docs/01-micro-tasks.md` contains Wishlist status updates and unrelated status
  reversions, but it is protected by this runbook and must remain untouched.

## 4. Build The Auto-Restore File List

```bash
git diff --name-only --no-renames --diff-filter=AM dev..feature/wishlist-service -- \
  backend/services/wishlist-service \
  'TaskImplementation/Wishlist Service' \
  ':(exclude)docs/01-micro-tasks.md' \
  > /tmp/wishlist-service-auto-restore.txt

sed -n '1,240p' /tmp/wishlist-service-auto-restore.txt
wc -l /tmp/wishlist-service-auto-restore.txt

if grep -Fxq 'docs/01-micro-tasks.md' /tmp/wishlist-service-auto-restore.txt; then
  echo "Protected docs/01-micro-tasks.md entered the restore list. Stop."
  exit 1
fi

if grep -Ev '^(backend/services/wishlist-service/|TaskImplementation/Wishlist Service/)' \
  /tmp/wishlist-service-auto-restore.txt; then
  echo "Unexpected path entered the restore list. Stop."
  exit 1
fi
```

What this does:

- Lists only added and modified Wishlist Service source and implementation
  documentation files.
- Excludes deleted files.
- Excludes all shared files and explicitly excludes
  `docs/01-micro-tasks.md`.
- Verifies that every auto-restore path is inside an approved Wishlist Service
  directory.
- Uses `--no-renames` so Wishlist Service `go.mod` and `go.sum` are included as
  added files.

The list should currently contain 67 files. If the count changes, review every
listed path before continuing.

## 5. Restore Added And Modified Wishlist Service Files

```bash
if [ -s /tmp/wishlist-service-auto-restore.txt ]; then
  git restore --overlay --source=feature/wishlist-service --worktree \
    --pathspec-from-file=/tmp/wishlist-service-auto-restore.txt
else
  echo "No Wishlist Service source/doc files to auto-restore."
fi

git diff --quiet -- docs/01-micro-tasks.md || {
  echo "Protected docs/01-micro-tasks.md changed. Stop."
  exit 1
}
```

What this does:

- Copies selected new files from `feature/wishlist-service`.
- Updates selected existing files only when they differ from `dev`.
- Uses `--overlay` so Git does not remove `dev` files.
- Confirms that the protected micro-task file remains unchanged.

## 6. Handle Shared Files Manually

### Update `backend/go.work` Safely

```bash
(
  cd backend
  go work use ./services/wishlist-service
)

git diff -- backend/go.work
```

What this does:

- Adds `./services/wishlist-service` to the existing `dev` workspace.
- Keeps all existing `dev` workspace entries.
- Avoids restoring the feature branch version, which currently removes every
  existing `dev` workspace entry.

### Update `backend/go.work.sum` Without Removing Existing Sums

```bash
git show dev:backend/go.work.sum > /tmp/dev-go.work.sum
git show feature/wishlist-service:backend/go.work.sum > /tmp/feature-wishlist-go.work.sum
LC_ALL=C sort -u \
  /tmp/dev-go.work.sum \
  /tmp/feature-wishlist-go.work.sum \
  > backend/go.work.sum

git diff -- backend/go.work.sum
```

What this does:

- Keeps checksum lines that already exist on `dev`.
- Adds checksum lines introduced by `feature/wishlist-service`.
- Prevents the feature branch from deleting existing `dev` checksum lines.

### Leave Other Shared And Protected Files Untouched

Do not restore or edit:

- `api/master-api.json`
- `docs/01-micro-tasks.md`

Verify both remain unchanged:

```bash
git diff --quiet -- api/master-api.json || {
  echo "api/master-api.json changed. Stop and review."
  exit 1
}

git diff --quiet -- docs/01-micro-tasks.md || {
  echo "Protected docs/01-micro-tasks.md changed. Stop."
  exit 1
}
```

## 7. Confirm No Deletions Were Applied

```bash
git diff --name-status --diff-filter=D --exit-code
git diff --quiet -- docs/01-micro-tasks.md || {
  echo "Protected docs/01-micro-tasks.md changed. Stop."
  exit 1
}
```

What this does:

- Confirms the working tree has no deletions.
- Confirms `docs/01-micro-tasks.md` remains unchanged.

The deletion check must print nothing. If it prints any `D` entry, stop before
staging.

## 8. Show Status And Diff Stat Before Staging

```bash
if [ -s /tmp/wishlist-service-auto-restore.txt ]; then
  git add -N --pathspec-from-file=/tmp/wishlist-service-auto-restore.txt
fi

git status
git diff --stat
git diff --name-status --diff-filter=D --exit-code
git diff --quiet -- docs/01-micro-tasks.md || {
  echo "Protected docs/01-micro-tasks.md changed. Stop."
  exit 1
}
```

What this does:

- Marks only the approved new Wishlist Service files as intent-to-add so
  `git diff --stat` shows them.
- Shows the required status and diff stat before full staging.
- Confirms again that no deletion is present and the protected file is
  unchanged.

## 9. Stage Only The Selected Changes

```bash
if [ -s /tmp/wishlist-service-auto-restore.txt ]; then
  git add --pathspec-from-file=/tmp/wishlist-service-auto-restore.txt
fi

git add -- backend/go.work backend/go.work.sum

git status
git diff --cached --stat
git diff --cached --name-status
git diff --cached --name-status --diff-filter=D --exit-code
git diff --cached --quiet -- docs/01-micro-tasks.md || {
  echo "Protected docs/01-micro-tasks.md is staged. Stop."
  exit 1
}
```

What this does:

- Stages only the approved Wishlist Service files and the two manually handled
  Go workspace files.
- Does not stage `api/master-api.json` or `docs/01-micro-tasks.md`.
- Shows all staged paths for review.
- Confirms no deletion is staged.

The staged deletion check must print nothing. Review the staged name-status
list and stop if it contains any path outside:

- `backend/services/wishlist-service`
- `TaskImplementation/Wishlist Service`
- `backend/go.work`
- `backend/go.work.sum`

## 10. Commit With The Exact Message

```bash
git diff --cached --name-status --diff-filter=D --exit-code
git diff --cached --quiet -- docs/01-micro-tasks.md || {
  echo "Protected docs/01-micro-tasks.md is staged. Stop."
  exit 1
}

git commit -m "Merge Wishlist Service work into dev" -- \
  backend/services/wishlist-service \
  'TaskImplementation/Wishlist Service' \
  backend/go.work \
  backend/go.work.sum \
  ':(exclude)docs/01-micro-tasks.md'
```

What this does:

- Rechecks the staged deletion and protected-file guards immediately before
  committing.
- Commits only the selected Wishlist Service paths and reviewed shared files.
- Explicitly excludes `docs/01-micro-tasks.md` from the commit command.
- Creates the selective integration commit with the exact required message.

## 11. Verify After Commit

```bash
git status --short --branch
git log --oneline -1
git show --stat --oneline HEAD
git show --name-status --format= HEAD

test "$(git log -1 --pretty=%s)" = "Merge Wishlist Service work into dev" || {
  echo "Unexpected commit message. Stop."
  exit 1
}

test -z "$(git show --format= --name-only --diff-filter=D HEAD)" || {
  echo "The selective commit contains deletions. Stop."
  exit 1
}

test -z "$(git show --format= --name-only HEAD -- docs/01-micro-tasks.md)" || {
  echo "The selective commit contains docs/01-micro-tasks.md. Stop."
  exit 1
}
```

What this does:

- Verifies the commit message and committed file summary.
- Confirms the selective commit contains no deletions.
- Confirms the selective commit does not contain
  `docs/01-micro-tasks.md`.

## 12. Fast-Forward Dev

```bash
git switch dev
git merge --ff-only chore/Wishlist-Service-selective
```

What this does:

- Moves `dev` to the reviewed selective integration commit.
- Does not merge `feature/wishlist-service`.
- Fails safely if `dev` changed and cannot fast-forward.

## 13. Final Verification

```bash
git status --short --branch
git log --oneline -1
git show --stat --oneline HEAD
git show --name-status --format= HEAD

test "$(git rev-parse dev)" = "$(git rev-parse chore/Wishlist-Service-selective)" || {
  echo "dev does not match the selective integration branch. Stop."
  exit 1
}

test "$(git log -1 --pretty=%s)" = "Merge Wishlist Service work into dev" || {
  echo "Unexpected final commit message. Stop."
  exit 1
}

test -z "$(git show --format= --name-only --diff-filter=D HEAD)" || {
  echo "Final commit contains deletions. Stop."
  exit 1
}

test -z "$(git show --format= --name-only HEAD -- docs/01-micro-tasks.md)" || {
  echo "Final commit contains docs/01-micro-tasks.md. Stop."
  exit 1
}

grep -F './services/wishlist-service' backend/go.work
test -f backend/services/wishlist-service/go.mod
```

What this does:

- Confirms `dev` points to the selective integration commit.
- Confirms the latest commit message is:

```text
Merge Wishlist Service work into dev
```

- Confirms the final commit has no deletions and does not contain
  `docs/01-micro-tasks.md`.
- Confirms the Wishlist Service module exists and is registered in the Go
  workspace.
