# Merge User Service Work Into Dev

Use this runbook to bring User Service work from `feature/user-service` into
`dev` without merging the whole feature branch and without deleting existing
files from `dev`.

Do not run:

```bash
git merge feature/user-service
```

This branch has unrelated deletions. We only apply added and modified User
Service related files, and we handle shared files carefully.

## Rules

- Add new User Service related files from `feature/user-service`.
- If a file already exists on `dev`, update it only when it is modified in
  `feature/user-service`.
- If a file exists on `dev` but is deleted or missing on `feature/user-service`,
  keep the `dev` file.
- Do not apply deletion diffs.
- Update `docs/01-micro-tasks.md` so User Service tasks are marked correctly.
- Commit with this exact message:

```text
Merge user service work into dev
```

## 1. Precheck

```bash
set -euo pipefail

git status --short --branch
git rev-parse --verify dev
git rev-parse --verify feature/user-service
test -z "$(git status --porcelain)" || { echo "Working tree is dirty. Commit/stash first."; exit 1; }
```

What this does:

- Confirms both branches exist.
- Confirms the current working tree is clean before changing branches.

## 2. Create A Safe Integration Branch From Dev

```bash
git switch dev
test -z "$(git status --porcelain)" || { echo "dev is dirty. Stop."; exit 1; }

git switch -c chore/user-service-selective
```

What this does:

- Starts from `dev`.
- Creates a temporary branch for the selective merge work.
- Keeps `dev` safe until the final fast-forward step.

## 3. Review The Full Branch Difference

```bash
git diff --name-status dev..feature/user-service
git diff --name-status --diff-filter=D dev..feature/user-service
git diff --name-status --diff-filter=AM dev..feature/user-service
```

What this does:

- Shows the complete difference from `dev` to `feature/user-service`.
- Shows deletions separately. Do not apply those files.
- Shows only added and modified files. These are the only candidates for this
  runbook.

Current important shared files to check:

```bash
git diff dev..feature/user-service -- backend/go.work
git diff dev..feature/user-service -- backend/go.work.sum
git diff dev..feature/user-service -- docs/01-micro-tasks.md
```

In the current repo state, these shared files can be the only remaining modified
files after User Service source files have already been brought into `dev`.

## 4. Build The Auto-Restore File List

```bash
git diff --name-only --diff-filter=AM dev..feature/user-service -- \
  backend/services/user-service \
  backend/services/api-gateway \
  backend/shared/gen/go/ecommerce/user/v1 \
  backend/shared/gen/go/go.mod \
  backend/shared/gen/go/go.sum \
  backend/shared/validation \
  proto/buf.yaml \
  proto/buf.gen.yaml \
  proto/ecommerce/user/v1/user.proto \
  'TaskImplementation/User Service' \
  > /tmp/user-service-auto-restore.txt

sed -n '1,240p' /tmp/user-service-auto-restore.txt
```

What this does:

- Lists User Service related files that are added or modified.
- Excludes deleted files.
- Excludes protected shared files that need special handling:
  `backend/go.work`, `backend/go.work.sum`, and `docs/01-micro-tasks.md`.

If this file is empty, that is okay. It means those files are already updated on
`dev`, and you should continue with the shared-file steps.

## 5. Restore Added And Modified User Service Files

```bash
if [ -s /tmp/user-service-auto-restore.txt ]; then
  git restore --overlay --source=feature/user-service --worktree --pathspec-from-file=/tmp/user-service-auto-restore.txt
else
  echo "No User Service source/doc files to auto-restore."
fi
```

What this does:

- Copies selected new files from `feature/user-service`.
- Updates selected existing files only when they differ from `dev`.
- Uses `--overlay` so Git does not remove dev files.

## 6. Update `backend/go.work` Safely

```bash
(
  cd backend
  go work use ./services/api-gateway ./services/user-service ./shared/gen/go ./shared/validation
)

git diff -- backend/go.work
```

What this does:

- Adds User Service workspace modules.
- Keeps existing dev workspace entries such as `./services/auth-service`.
- Avoids replacing `backend/go.work` with the feature branch version, because
  the feature version removes existing dev entries and changes the Go version.

## 7. Update `backend/go.work.sum` Without Removing Existing Sums

```bash
git show dev:backend/go.work.sum > /tmp/dev-go.work.sum
git show feature/user-service:backend/go.work.sum > /tmp/feature-user-go.work.sum
LC_ALL=C sort -u /tmp/dev-go.work.sum /tmp/feature-user-go.work.sum > backend/go.work.sum

git diff -- backend/go.work.sum
```

What this does:

- Keeps checksum lines that already exist on `dev`.
- Adds checksum lines introduced by `feature/user-service`.
- Prevents the feature branch from deleting existing `dev` checksum lines.

## 8. Update `docs/01-micro-tasks.md`

```bash
git diff dev..feature/user-service -- docs/01-micro-tasks.md
git restore --source=feature/user-service --worktree -- docs/01-micro-tasks.md
git diff -- docs/01-micro-tasks.md
```

What this does:

- Shows the `micro-tasks.md` change before applying it.
- Updates User Service task statuses from `Pending` to `Completed`.
- Ensures `dev` receives the `docs/01-micro-tasks.md` update.

If the first diff ever shows unrelated changes outside the User Service section,
use patch mode instead:

```bash
git restore -p --source=feature/user-service --worktree -- docs/01-micro-tasks.md
```

Accept the User Service task-status hunk and reject unrelated hunks.

## 9. Confirm No Deletions Were Applied

```bash
git diff --name-status --diff-filter=D --exit-code
```

What this does:

- Confirms the working tree has no deletions.
- This command must print nothing.

If it prints any `D` entry, stop before staging.

## 10. Show Status And Diff Stat Before Staging

```bash
if [ -s /tmp/user-service-auto-restore.txt ]; then
  git add -N --pathspec-from-file=/tmp/user-service-auto-restore.txt
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

## 11. Stage Only The Selected Changes

```bash
if [ -s /tmp/user-service-auto-restore.txt ]; then
  git add --pathspec-from-file=/tmp/user-service-auto-restore.txt
fi

git add backend/go.work backend/go.work.sum docs/01-micro-tasks.md

git status
git diff --cached --stat
git diff --cached --name-status --diff-filter=D --exit-code
```

What this does:

- Stages selected User Service added and modified files.
- Stages the reviewed shared files.
- Confirms no deletions are staged.

The staged deletion check must print nothing.

## 12. Commit

```bash
git commit -m "Merge user service work into dev"
```

What this does:

- Creates the required selective integration commit.

## 13. Fast-Forward Dev

```bash
git switch dev
git merge --ff-only chore/user-service-selective
```

What this does:

- Moves `dev` to the selective integration commit.
- Does not merge `feature/user-service`.
- Fails safely if `dev` changed and cannot fast-forward.

## 14. Final Verification

```bash
git status --short --branch
git log --oneline -1
git show --stat --oneline HEAD
```

What this does:

- Confirms `dev` is on the new commit.
- Shows the final committed file summary.
- Confirms the latest commit message is:

```text
Merge user service work into dev
```
