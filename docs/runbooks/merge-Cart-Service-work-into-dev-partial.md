# Merge Cart Service Work Into Dev Selectively

Use this runbook to bring Cart Service work from `feature/cart-service` into
`dev` without merging the whole feature branch and without deleting existing
files from `dev`.

Do not run:

```bash
git merge feature/cart-service
```

The current feature branch contains many unrelated deletions. Restore only
added or modified Cart Service files, update shared files additively, and leave
`docs/01-micro-tasks.md` unchanged.

## Rules

- Add or update Cart Service files only from:
  `backend/services/cart-service` and `TaskImplementation/Cart Service`.
- Do not apply any deletion from `feature/cart-service`.
- Do not restore, stage, or commit `docs/01-micro-tasks.md`.
- Preserve every existing `dev` entry in shared files.
- Update `backend/go.work` manually so it gains the Cart Service module without
  losing existing workspace modules.
- Leave `api/master-api.json` and `backend/go.work.sum` unchanged. Their current
  `dev..feature/cart-service` differences are not Cart Service changes made on
  the feature branch.
- Commit with this exact message:

```text
Merge Cart Service work into dev
```

## Current Repository Findings

In the repository state used to prepare this runbook:

- The scoped auto-restore list contains 69 added Cart Service files.
- `backend/services/cart-service/go.sum` can appear as a rename from
  `backend/services/product-service/go.sum` when rename detection is enabled.
  The commands below use `--no-renames` so the Cart Service file is restored as
  an addition and the Product Service file is not deleted.
- The full `dev..feature/cart-service` difference contains many unrelated
  deletions across existing services, task documents, shared modules, proto
  files, reports, and other repository files. None of those deletions belong in
  this selective integration.
- Shared added/modified candidates outside the Cart Service paths are
  `api/master-api.json`, `backend/go.work`, and `backend/go.work.sum`.
- Comparing the feature branch to its merge base with `dev` shows that only
  `backend/go.work` and `docs/01-micro-tasks.md` were changed on the feature
  branch. Only `backend/go.work` should be handled, and it must be updated
  additively.

## 1. Precheck

```bash
set -euo pipefail

git status --short --branch
git rev-parse --verify dev
git rev-parse --verify feature/cart-service
test -z "$(git status --porcelain)" || {
  echo "Working tree is dirty. Commit or stash existing work first."
  exit 1
}

if git show-ref --verify --quiet refs/heads/chore/Cart-Service-selective; then
  echo "Integration branch chore/Cart-Service-selective already exists. Stop."
  exit 1
fi
```

What this does:

- Confirms the source and target branches exist.
- Requires a clean working tree before switching branches.
- Prevents accidentally reusing an existing integration branch.

## 2. Create A Safe Integration Branch From `dev`

```bash
git switch dev
test -z "$(git status --porcelain)" || { echo "dev is dirty. Stop."; exit 1; }

git switch -c chore/Cart-Service-selective
test "$(git branch --show-current)" = "chore/Cart-Service-selective"
```

What this does:

- Starts the selective work from the current `dev` tree.
- Keeps `dev` unchanged until the final fast-forward step.

## 3. Review The Full Branch Difference

```bash
git diff --no-renames --name-status dev..feature/cart-service
git diff --no-renames --name-status --diff-filter=D dev..feature/cart-service
git diff --no-renames --name-status --diff-filter=AM dev..feature/cart-service
```

What this does:

- Shows the complete source-versus-target difference.
- Shows deletions separately. Do not apply any file from that list.
- Shows added and modified files without rename detection obscuring the Cart
  Service `go.sum` addition.

Review the protected and shared candidates separately:

```bash
git diff dev..feature/cart-service -- \
  api/master-api.json \
  backend/go.work \
  backend/go.work.sum \
  docs/01-micro-tasks.md

MERGE_BASE="$(git merge-base dev feature/cart-service)"
git diff --name-status "$MERGE_BASE"..feature/cart-service -- \
  api/master-api.json \
  backend/go.work \
  backend/go.work.sum \
  docs/01-micro-tasks.md
```

The second command distinguishes changes actually made on the Cart Service
branch from files that differ only because `dev` moved forward. In the current
state, it shows feature-branch changes only for `backend/go.work` and the
protected `docs/01-micro-tasks.md`.

## 4. Build The Auto-Restore File List

```bash
git diff --no-renames --name-only --diff-filter=AM dev..feature/cart-service -- \
  backend/services/cart-service \
  'TaskImplementation/Cart Service' \
  > /tmp/cart-service-auto-restore.txt

sed -n '1,240p' /tmp/cart-service-auto-restore.txt
wc -l /tmp/cart-service-auto-restore.txt

if grep -Ev \
  '^(backend/services/cart-service/|TaskImplementation/Cart Service/)' \
  /tmp/cart-service-auto-restore.txt; then
  echo "Unexpected path in the auto-restore list. Stop."
  exit 1
fi

if grep -Fxq 'docs/01-micro-tasks.md' /tmp/cart-service-auto-restore.txt; then
  echo "Protected docs/01-micro-tasks.md entered the restore list. Stop."
  exit 1
fi
```

What this does:

- Lists only added and modified files under the two Cart Service-owned paths.
- Uses `--no-renames` so `backend/services/cart-service/go.sum` is included as
  an added Cart Service file.
- Excludes every deletion and every shared file.
- Verifies that `docs/01-micro-tasks.md` is not in the restore list.

The current list should contain 69 files. If the count differs, review the list
carefully before continuing because the source or target branch may have moved.

## 5. Restore Added And Modified Cart Service Files

```bash
if [ -s /tmp/cart-service-auto-restore.txt ]; then
  git restore --overlay --source=feature/cart-service --worktree \
    --pathspec-from-file=/tmp/cart-service-auto-restore.txt
else
  echo "No Cart Service-owned added or modified files to restore."
fi
```

What this does:

- Copies only the selected Cart Service files from `feature/cart-service`.
- Uses `--overlay` so Git does not remove files already present on `dev`.
- Does not restore any shared file or protected documentation file.

## 6. Handle Shared Files Manually

First, build and inspect the shared candidate list:

```bash
git diff --no-renames --name-only --diff-filter=AM dev..feature/cart-service -- \
  . \
  ':(exclude)backend/services/cart-service/**' \
  ':(exclude)TaskImplementation/Cart Service/**' \
  ':(exclude)docs/01-micro-tasks.md' \
  > /tmp/cart-service-shared-candidates.txt

sed -n '1,240p' /tmp/cart-service-shared-candidates.txt
```

In the current repository state, this prints:

```text
api/master-api.json
backend/go.work
backend/go.work.sum
```

Do not restore any of these shared files from the feature branch.

Update only `backend/go.work` additively:

```bash
(
  cd backend
  go work use ./services/cart-service
)

git diff -- backend/go.work
```

The `backend/go.work` diff must add `./services/cart-service` while preserving
all existing `dev` workspace entries, including the existing services and
shared modules. Stop if the diff removes an existing entry.

Confirm that the other shared files and the protected file remain unchanged:

```bash
git diff --exit-code -- \
  api/master-api.json \
  backend/go.work.sum \
  docs/01-micro-tasks.md
```

Why these files remain unchanged:

- `api/master-api.json` has no Cart Service change after the merge base.
  Restoring it would regress newer `dev` content.
- `backend/go.work.sum` has no Cart Service change after the merge base.
  Restoring it would remove existing `dev` checksum lines.
- `docs/01-micro-tasks.md` is explicitly protected and must not be modified,
  restored, staged, or committed.

## 7. Confirm No Deletions Were Applied

```bash
git diff --name-status --diff-filter=D --exit-code
```

What this does:

- Confirms the working tree contains no deleted files.
- Must print nothing and exit successfully.

If it prints any `D` entry, stop before staging.

## 8. Show Status And Diff Stat Before Staging

Mark only the selected new files as intent-to-add so they appear in the
unstaged diff:

```bash
if [ -s /tmp/cart-service-auto-restore.txt ]; then
  git add -N --pathspec-from-file=/tmp/cart-service-auto-restore.txt
fi

git status
git diff --stat
git diff --check
git diff --name-status --diff-filter=D --exit-code
```

Verify that the working diff contains only the allowed Cart Service paths and
the manually updated workspace file:

```bash
UNEXPECTED_FILES="$(
  git diff --name-only -- \
    . \
    ':(exclude)backend/services/cart-service/**' \
    ':(exclude)TaskImplementation/Cart Service/**' \
    ':(exclude)backend/go.work'
)"

test -z "$UNEXPECTED_FILES" || {
  printf 'Unexpected working-tree files:\n%s\n' "$UNEXPECTED_FILES"
  exit 1
}

test -z "$(git diff --name-only -- docs/01-micro-tasks.md)" || {
  echo "Protected docs/01-micro-tasks.md changed. Stop."
  exit 1
}
```

## 9. Stage Only The Selected Changes

```bash
if [ -s /tmp/cart-service-auto-restore.txt ]; then
  git add --pathspec-from-file=/tmp/cart-service-auto-restore.txt
fi

git add backend/go.work
```

No other shared file is staged.

## 10. Verify After Staging

```bash
git status
git diff --cached --stat
git diff --cached --check
git diff --cached --name-status --diff-filter=D --exit-code

test -z "$(git diff --cached --name-only -- docs/01-micro-tasks.md)" || {
  echo "Protected docs/01-micro-tasks.md is staged. Stop."
  exit 1
}

UNEXPECTED_STAGED_FILES="$(
  git diff --cached --name-only -- \
    . \
    ':(exclude)backend/services/cart-service/**' \
    ':(exclude)TaskImplementation/Cart Service/**' \
    ':(exclude)backend/go.work'
)"

test -z "$UNEXPECTED_STAGED_FILES" || {
  printf 'Unexpected staged files:\n%s\n' "$UNEXPECTED_STAGED_FILES"
  exit 1
}

test -n "$(git diff --cached --name-only)" || {
  echo "Nothing is staged. Stop."
  exit 1
}

git diff --exit-code
```

What this does:

- Shows the exact staged file summary.
- Confirms no deletion is staged.
- Confirms `docs/01-micro-tasks.md` and unrelated shared files are not staged.
- Confirms no unstaged change was left behind.

## 11. Commit

```bash
git commit -m "Merge Cart Service work into dev"
```

What this does:

- Creates the selective integration commit with the exact required message.
- Commits only the files verified in the staged checks.

## 12. Verify After Commit

```bash
git status --short --branch
git log --oneline -1
git show --stat --oneline HEAD

test "$(git log -1 --pretty=%s)" = "Merge Cart Service work into dev"

test -z "$(git diff --name-only --diff-filter=D HEAD^ HEAD)" || {
  echo "The selective commit contains a deletion. Stop."
  exit 1
}

test -z "$(git diff --name-only HEAD^ HEAD -- docs/01-micro-tasks.md)" || {
  echo "The selective commit changed protected docs/01-micro-tasks.md. Stop."
  exit 1
}

UNEXPECTED_COMMITTED_FILES="$(
  git diff --name-only HEAD^ HEAD -- \
    . \
    ':(exclude)backend/services/cart-service/**' \
    ':(exclude)TaskImplementation/Cart Service/**' \
    ':(exclude)backend/go.work'
)"

test -z "$UNEXPECTED_COMMITTED_FILES" || {
  printf 'Unexpected committed files:\n%s\n' "$UNEXPECTED_COMMITTED_FILES"
  exit 1
}

test -z "$(git status --porcelain)" || {
  echo "Integration branch is not clean after commit. Stop."
  exit 1
}
```

What this does:

- Verifies the exact commit message.
- Confirms the commit contains no deletions.
- Confirms the protected file and unrelated files are absent from the commit.
- Requires a clean integration branch before updating `dev`.

## 13. Fast-Forward `dev`

```bash
git switch dev
test -z "$(git status --porcelain)" || { echo "dev is dirty. Stop."; exit 1; }

git merge --ff-only chore/Cart-Service-selective
```

What this does:

- Moves `dev` to the already verified selective integration commit.
- Does not merge `feature/cart-service`.
- Fails safely if `dev` changed and can no longer fast-forward.

## 14. Final Verification

```bash
test "$(git branch --show-current)" = "dev"
test "$(git log -1 --pretty=%s)" = "Merge Cart Service work into dev"

git status --short --branch
git log --oneline -1
git show --stat --oneline HEAD

test -z "$(git diff --name-only --diff-filter=D HEAD^ HEAD)" || {
  echo "The final dev commit contains a deletion. Stop."
  exit 1
}

test -z "$(git diff --name-only HEAD^ HEAD -- docs/01-micro-tasks.md)" || {
  echo "The final dev commit changed protected docs/01-micro-tasks.md. Stop."
  exit 1
}

UNEXPECTED_FINAL_FILES="$(
  git diff --name-only HEAD^ HEAD -- \
    . \
    ':(exclude)backend/services/cart-service/**' \
    ':(exclude)TaskImplementation/Cart Service/**' \
    ':(exclude)backend/go.work'
)"

test -z "$UNEXPECTED_FINAL_FILES" || {
  printf 'Unexpected files in final dev commit:\n%s\n' "$UNEXPECTED_FINAL_FILES"
  exit 1
}

test -z "$(git status --porcelain)" || {
  echo "dev is not clean after fast-forward. Stop."
  exit 1
}
```

The latest `dev` commit must contain only:

- Added or modified files under `backend/services/cart-service`.
- Added or modified files under `TaskImplementation/Cart Service`.
- The additive `backend/go.work` update.

It must contain no deletions and no change to `docs/01-micro-tasks.md`.
