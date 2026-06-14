# Merge Notification Service Work Into Dev

Use this runbook to bring Notification Service work from
`feature/Notification-service` into `dev` without merging the whole feature
branch and without deleting existing files from `dev`.

Do not merge the source branch wholesale.

This branch has many unrelated deletions. We only apply added and modified
Notification Service related files, and we handle shared files carefully.

## Rules

- Add new Notification Service related files from `feature/Notification-service`.
- If a file already exists on `dev`, update it only when it is modified in
  `feature/Notification-service`.
- If a file exists on `dev` but is deleted or missing on
  `feature/Notification-service`, keep the `dev` file.
- Do not apply deletion diffs.
- Do not modify `docs/01-micro-tasks.md`.
- Do not restore, stage, or commit `docs/01-micro-tasks.md`.
- Handle shared files manually so existing `dev` content is not removed.
- Commit with this exact message:

```text
Merge Notification Service work into dev
```

## 1. Precheck

```bash
set -euo pipefail

git status --short --branch
git rev-parse --verify dev
git rev-parse --verify feature/Notification-service
test -z "$(git status --porcelain)" || { echo "Working tree is dirty. Commit/stash first."; exit 1; }

if git show-ref --verify --quiet refs/heads/chore/Notification-Service-selective; then
  echo "Integration branch already exists. Stop or choose a new branch name."
  exit 1
fi
```

What this does:

- Confirms both branches exist.
- Confirms the current working tree is clean before changing branches.
- Avoids accidentally reusing an existing integration branch.

## 2. Create A Safe Integration Branch From Dev

```bash
git switch dev
test -z "$(git status --porcelain)" || { echo "dev is dirty. Stop."; exit 1; }

git switch -c chore/Notification-Service-selective
```

What this does:

- Starts from `dev`.
- Creates a temporary branch for the selective merge work.
- Keeps `dev` safe until the final fast-forward step.

## 3. Review The Full Branch Difference

```bash
git diff --name-status dev..feature/Notification-service
git diff --name-status --diff-filter=D dev..feature/Notification-service
git diff --name-status --diff-filter=AM dev..feature/Notification-service
```

What this does:

- Shows the complete difference from `dev` to `feature/Notification-service`.
- Shows deletions separately. Do not apply those files.
- Shows only added and modified files. These are the only candidates for this
  runbook.

Current important paths to check:

```bash
git diff --name-status --diff-filter=AM dev..feature/Notification-service -- \
  backend/services/notification-service \
  proto/ecommerce/notification/v1 \
  'TaskImplementation/Notification Service'

git diff dev..feature/Notification-service -- backend/go.work
git diff dev..feature/Notification-service -- backend/go.work.sum
git diff dev..feature/Notification-service -- api/master-api.json
git diff dev..feature/Notification-service -- docs/01-micro-tasks.md
```

In the current repo state, the Notification Service branch adds:

- `backend/services/notification-service`
- `proto/ecommerce/notification/v1/notification.proto`
- `TaskImplementation/Notification Service`

It also modifies shared files:

- `backend/go.work`
- `backend/go.work.sum`
- `api/master-api.json`
- `docs/01-micro-tasks.md`

Do not restore shared files wholesale. In particular,
`docs/01-micro-tasks.md` is protected and must remain unchanged.

## 4. Build The Auto-Restore File List

```bash
git diff --name-only --diff-filter=AM dev..feature/Notification-service -- \
  backend/services/notification-service \
  proto/ecommerce/notification/v1 \
  'TaskImplementation/Notification Service' \
  ':(exclude)api/master-api.json' \
  ':(exclude)backend/go.work' \
  ':(exclude)backend/go.work.sum' \
  ':(exclude)docs/01-micro-tasks.md' \
  > /tmp/notification-service-auto-restore.txt

sed -n '1,240p' /tmp/notification-service-auto-restore.txt

if grep -E '^(api/master-api\.json|backend/go\.work|backend/go\.work\.sum|docs/01-micro-tasks\.md)$' /tmp/notification-service-auto-restore.txt; then
  echo "Protected shared file entered auto-restore list. Stop."
  exit 1
fi

test -s /tmp/notification-service-auto-restore.txt || { echo "No Notification Service files found. Stop and re-check branch names."; exit 1; }
```

What this does:

- Lists Notification Service related files that are added or modified.
- Excludes deleted files.
- Excludes shared files that need manual handling.
- Excludes `docs/01-micro-tasks.md` from the restore list.

## 5. Restore Added And Modified Notification Service Files

```bash
git restore --overlay --source=feature/Notification-service --worktree --pathspec-from-file=/tmp/notification-service-auto-restore.txt
```

What this does:

- Copies selected new files from `feature/Notification-service`.
- Updates selected existing files only when they differ from `dev`.
- Uses `--overlay` so Git does not remove dev files.
- Does not touch `docs/01-micro-tasks.md`.

## 6. Handle Shared Files Manually

First update `backend/go.work` safely:

```bash
(
  cd backend
  go work use ./services/notification-service
)

git diff -- backend/go.work
```

What this does:

- Adds the Notification Service workspace module.
- Keeps existing dev workspace entries such as `./services/api-gateway`,
  `./services/auth-service`, `./services/user-service`, `./shared/gen/go`, and
  `./shared/validation`.
- Avoids replacing `backend/go.work` with the feature branch version, because
  the feature version removes existing dev entries.

Now update `backend/go.work.sum` without removing existing sums:

```bash
git show dev:backend/go.work.sum > /tmp/dev-notification-go.work.sum
git show feature/Notification-service:backend/go.work.sum > /tmp/feature-notification-go.work.sum
LC_ALL=C sort -u /tmp/dev-notification-go.work.sum /tmp/feature-notification-go.work.sum > backend/go.work.sum

git diff -- backend/go.work.sum
```

What this does:

- Keeps checksum lines that already exist on `dev`.
- Adds checksum lines introduced by `feature/Notification-service`, if any.
- Prevents the feature branch from deleting existing `dev` checksum lines.

Review `api/master-api.json` but do not stage it for the current branch diff:

```bash
git diff dev..feature/Notification-service -- api/master-api.json
git diff -- api/master-api.json
```

In the current repo state, `api/master-api.json` changes product and payment API
definitions, not Notification Service entries. Leave this file unchanged for
this runbook.

Review `docs/01-micro-tasks.md` only:

```bash
git diff dev..feature/Notification-service -- docs/01-micro-tasks.md
git diff -- docs/01-micro-tasks.md
```

The current branch changes the Notification Service task statuses, but it also
changes many unrelated task statuses. This runbook intentionally does not apply
any `docs/01-micro-tasks.md` changes.

## 7. Confirm No Deletions Were Applied

```bash
git diff --name-status --diff-filter=D --exit-code
git diff --quiet -- docs/01-micro-tasks.md || { echo "docs/01-micro-tasks.md changed. Stop."; exit 1; }
git diff --quiet -- api/master-api.json || { echo "api/master-api.json changed. Stop."; exit 1; }
```

What this does:

- Confirms the working tree has no deletions.
- Confirms the protected micro-task file is unchanged.
- Confirms `api/master-api.json` was not accidentally changed.

The deletion check must print nothing.

## 8. Show Status And Diff Stat Before Staging

```bash
git add -N --pathspec-from-file=/tmp/notification-service-auto-restore.txt

git status
git diff --stat
git diff --name-status --diff-filter=D --exit-code
git diff --quiet -- docs/01-micro-tasks.md || { echo "docs/01-micro-tasks.md changed. Stop."; exit 1; }
git diff --quiet -- api/master-api.json || { echo "api/master-api.json changed. Stop."; exit 1; }
```

What this does:

- Marks new auto-restored files as intent-to-add so `git diff --stat` shows
  them.
- Shows the required `git status`.
- Shows the required `git diff --stat`.
- Confirms again that no deletion is present.
- Confirms `docs/01-micro-tasks.md` is still unchanged before staging.

## 9. Stage Only The Selected Changes

```bash
git add --pathspec-from-file=/tmp/notification-service-auto-restore.txt
git add backend/go.work backend/go.work.sum

git status
git diff --cached --stat
git diff --cached --name-status
git diff --cached --name-status --diff-filter=D --exit-code
git diff --cached --quiet -- docs/01-micro-tasks.md || { echo "docs/01-micro-tasks.md is staged. Stop."; exit 1; }
git diff --cached --quiet -- api/master-api.json || { echo "api/master-api.json is staged. Stop."; exit 1; }
```

What this does:

- Stages selected Notification Service added and modified files.
- Stages the reviewed `backend/go.work` and `backend/go.work.sum` files.
- Does not stage `docs/01-micro-tasks.md`.
- Does not stage unrelated API changes from `api/master-api.json`.
- Confirms no deletions are staged.

The staged deletion check must print nothing.

## 10. Commit

```bash
git commit -m "Merge Notification Service work into dev"
```

What this does:

- Creates the required selective integration commit.

After the commit, verify the commit is clean:

```bash
git status --short --branch
git log --oneline -1
git show --stat --oneline HEAD
git diff-tree --no-commit-id --name-status -r HEAD

if git diff-tree --no-commit-id --name-status -r HEAD | grep '^D'; then
  echo "Deletion was committed. Stop."
  exit 1
fi

if git diff-tree --no-commit-id --name-only -r HEAD | grep -Fx 'docs/01-micro-tasks.md'; then
  echo "docs/01-micro-tasks.md was committed. Stop."
  exit 1
fi
```

What this verifies:

- The latest commit message is the required message.
- The commit summary contains only selected Notification Service work and
  reviewed workspace files.
- The commit contains no deletions.
- The commit does not contain `docs/01-micro-tasks.md`.

## 11. Fast-Forward Dev

```bash
git switch dev
git merge --ff-only chore/Notification-Service-selective
```

What this does:

- Moves `dev` to the selective integration commit.
- Fails safely if `dev` changed and cannot fast-forward.

## 12. Final Verification

```bash
git status --short --branch
git log --oneline -1
git show --stat --oneline HEAD
git diff-tree --no-commit-id --name-status -r HEAD

if git diff-tree --no-commit-id --name-status -r HEAD | grep '^D'; then
  echo "Deletion is present in the final commit. Stop."
  exit 1
fi

if git diff-tree --no-commit-id --name-only -r HEAD | grep -Fx 'docs/01-micro-tasks.md'; then
  echo "docs/01-micro-tasks.md is present in the final commit. Stop."
  exit 1
fi
```

What this does:

- Confirms `dev` is on the new selective integration commit.
- Shows the final committed file summary.
- Confirms no deletion was included.
- Confirms `docs/01-micro-tasks.md` was not included.
- Confirms the latest commit message is:

```text
Merge Notification Service work into dev
```
