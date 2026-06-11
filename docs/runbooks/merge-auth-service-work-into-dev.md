# Merge Auth Service Work Into Dev

Use this runbook to bring Auth Service work from `feature/auth-service` into
`dev` without merging the whole feature branch and without deleting existing
files from `dev`.

Do not run:

```bash
git merge feature/auth-service
```

This branch has unrelated deletions. We only apply added and modified Auth
Service related files, and we handle shared files carefully.

## Rules

- Add new Auth Service related files from `feature/auth-service`.
- If a file already exists on `dev`, update it only when it is modified in
  `feature/auth-service`.
- If a file exists on `dev` but is deleted or missing on `feature/auth-service`,
  keep the `dev` file.
- Do not apply deletion diffs.
- Update `docs/01-micro-tasks.md` so Auth Service tasks are marked correctly.
- Commit with this exact message:

```text
Merge auth service work into dev
```

## 1. Precheck

```bash
set -euo pipefail

git status --short --branch
git rev-parse --verify dev
git rev-parse --verify feature/auth-service
test -z "$(git status --porcelain)" || { echo "Working tree is dirty. Commit/stash first."; exit 1; }
```

What this does:

- Confirms both branches exist.
- Confirms the current working tree is clean before changing branches.

## 2. Create A Safe Integration Branch From Dev

```bash
git switch dev
test -z "$(git status --porcelain)" || { echo "dev is dirty. Stop."; exit 1; }

git switch -c chore/auth-service-selective
```

What this does:

- Starts from `dev`.
- Creates a temporary branch for the selective merge work.
- Keeps `dev` safe until the final fast-forward step.

## 3. Review The Full Branch Difference

```bash
git diff --name-status dev..feature/auth-service
git diff --name-status --diff-filter=D dev..feature/auth-service
git diff --name-status --diff-filter=AM dev..feature/auth-service
```

What this does:

- Shows the complete difference from `dev` to `feature/auth-service`.
- Shows deletions separately. Do not apply those files.
- Shows only added and modified files. These are the only candidates for this
  runbook.

Current important shared files to check:

```bash
git diff dev..feature/auth-service -- backend/go.work
git diff dev..feature/auth-service -- backend/go.work.sum
git diff dev..feature/auth-service -- docs/01-micro-tasks.md
```

In the current repo state, the Auth Service branch updates `backend/go.work`,
`backend/go.work.sum`, and `docs/01-micro-tasks.md`. These shared files need
manual handling so existing `dev` entries are not removed.

## 4. Build The Auto-Restore File List

```bash
git diff --name-only --diff-filter=AM dev..feature/auth-service -- \
  backend/services/auth-service \
  'TaskImplementation/Auth Service' \
  > /tmp/auth-service-auto-restore.txt

sed -n '1,240p' /tmp/auth-service-auto-restore.txt
```

What this does:

- Lists Auth Service related files that are added or modified.
- Excludes deleted files.
- Excludes protected shared files that need special handling:
  `backend/go.work`, `backend/go.work.sum`, and `docs/01-micro-tasks.md`.

If this file is empty, that is okay. It means those files are already updated on
`dev`, and you should continue with the shared-file steps.

## 5. Restore Added And Modified Auth Service Files

```bash
if [ -s /tmp/auth-service-auto-restore.txt ]; then
  git restore --overlay --source=feature/auth-service --worktree --pathspec-from-file=/tmp/auth-service-auto-restore.txt
else
  echo "No Auth Service source/doc files to auto-restore."
fi
```

What this does:

- Copies selected new files from `feature/auth-service`.
- Updates selected existing files only when they differ from `dev`.
- Uses `--overlay` so Git does not remove dev files.

## 6. Update `backend/go.work` Safely

```bash
(
  cd backend
  go work use ./services/auth-service
)

git diff -- backend/go.work
```

What this does:

- Adds the Auth Service workspace module.
- Keeps existing dev workspace entries such as `./services/api-gateway`,
  `./services/user-service`, `./shared/gen/go`, and `./shared/validation`.
- Avoids replacing `backend/go.work` with the feature branch version, because
  the feature version removes existing dev entries.

## 7. Update `backend/go.work.sum` Without Removing Existing Sums

```bash
git show dev:backend/go.work.sum > /tmp/dev-go.work.sum
git show feature/auth-service:backend/go.work.sum > /tmp/feature-auth-go.work.sum
LC_ALL=C sort -u /tmp/dev-go.work.sum /tmp/feature-auth-go.work.sum > backend/go.work.sum

git diff -- backend/go.work.sum
```

What this does:

- Keeps checksum lines that already exist on `dev`.
- Adds checksum lines introduced by `feature/auth-service`.
- Prevents the feature branch from deleting existing `dev` checksum lines.

## 8. Update `docs/01-micro-tasks.md`

```bash
git diff dev..feature/auth-service -- docs/01-micro-tasks.md
git restore --source=feature/auth-service --worktree -- docs/01-micro-tasks.md
git diff -- docs/01-micro-tasks.md
```

What this does:

- Shows the `micro-tasks.md` change before applying it.
- Updates Auth Service task statuses from `Pending` to `Completed`.
- Ensures `dev` receives the `docs/01-micro-tasks.md` update.

If the first diff ever shows unrelated changes outside the Auth Service section,
use patch mode instead:

```bash
git restore -p --source=feature/auth-service --worktree -- docs/01-micro-tasks.md
```

Accept the Auth Service task-status hunk and reject unrelated hunks.

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
if [ -s /tmp/auth-service-auto-restore.txt ]; then
  git add -N --pathspec-from-file=/tmp/auth-service-auto-restore.txt
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
if [ -s /tmp/auth-service-auto-restore.txt ]; then
  git add --pathspec-from-file=/tmp/auth-service-auto-restore.txt
fi

git add backend/go.work backend/go.work.sum docs/01-micro-tasks.md

git status
git diff --cached --stat
git diff --cached --name-status --diff-filter=D --exit-code
```

What this does:

- Stages selected Auth Service added and modified files.
- Stages the reviewed shared files.
- Confirms no deletions are staged.

The staged deletion check must print nothing.

## 12. Commit

```bash
git commit -m "Merge auth service work into dev"
```

What this does:

- Creates the required selective integration commit.

## 13. Fast-Forward Dev

```bash
git switch dev
git merge --ff-only chore/auth-service-selective
```

What this does:

- Moves `dev` to the selective integration commit.
- Does not merge `feature/auth-service`.
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
Merge auth service work into dev
```
