# Merge User App Frontend Work Into Dev

Use this runbook to bring User App Frontend work from
`feature/user-app-frontend` into `dev` without merging the whole feature branch
and without deleting existing files from `dev`.

Do not merge the source branch directly. This branch has broad unrelated
deletions across backend, infra, proto, and task documentation. We only apply
added and modified User App Frontend related files, and we handle shared files
carefully.

## Rules

- Add new User App Frontend related files from `feature/user-app-frontend`.
- If a file already exists on `dev`, update it only when it is modified in
  `feature/user-app-frontend` and is part of the User App Frontend scope.
- If a file exists on `dev` but is deleted or missing on
  `feature/user-app-frontend`, keep the `dev` file.
- Do not apply deletion diffs.
- Do not modify `docs/01-micro-tasks.md`.
- Exclude `docs/01-micro-tasks.md` from restore, staging, and commit work.
- Commit with this exact message:

```text
Merge User App Frontend work into dev
```

## 1. Precheck

```bash
set -euo pipefail

git status --short --branch
git rev-parse --verify dev
git rev-parse --verify feature/user-app-frontend
test -z "$(git status --porcelain)" || { echo "Working tree is dirty. Commit/stash first."; exit 1; }
```

What this does:

- Confirms both branches exist.
- Confirms the current working tree is clean before changing branches.
- Prevents accidental mixing with unrelated local work.

## 2. Create A Safe Integration Branch From Dev

```bash
git switch dev
test -z "$(git status --porcelain)" || { echo "dev is dirty. Stop."; exit 1; }

git switch -c chore/User-App-Frontend-selective
```

What this does:

- Starts from `dev`.
- Creates a temporary branch for the selective merge work.
- Keeps `dev` safe until the final fast-forward step.

## 3. Review The Full Branch Difference

```bash
git diff --name-status dev..feature/user-app-frontend
git diff --name-status --diff-filter=D dev..feature/user-app-frontend
git diff --name-status --diff-filter=AM dev..feature/user-app-frontend
```

What this does:

- Shows the complete difference from `dev` to `feature/user-app-frontend`.
- Shows deletions separately. Do not apply those files.
- Shows only added and modified files. These are the only candidates for this
  runbook.

Current User App Frontend candidates to review:

```bash
git diff --name-status --diff-filter=AM dev..feature/user-app-frontend -- \
  frontend \
  'TaskImplementation/User App Frontend'

git diff --stat dev..feature/user-app-frontend -- \
  frontend \
  'TaskImplementation/User App Frontend'
```

Current important shared files to check:

```bash
git diff --stat dev..feature/user-app-frontend -- \
  api/master-api.json \
  backend/go.work \
  backend/go.work.sum \
  docs/01-micro-tasks.md

git diff dev..feature/user-app-frontend -- api/master-api.json
git diff dev..feature/user-app-frontend -- backend/go.work
git diff dev..feature/user-app-frontend -- backend/go.work.sum
git diff dev..feature/user-app-frontend -- docs/01-micro-tasks.md
```

In the current repo state:

- User App Frontend files are added under `frontend`.
- User App Frontend task notes are added under
  `TaskImplementation/User App Frontend`.
- `TaskImplementation/Dependency/*` is also added, but those are shared
  dependency docs rather than service-specific User App Frontend files. Leave
  them out unless a separate documentation task owns them.
- `api/master-api.json` removes existing Superadmin/Search API content from
  `dev`. Do not restore it wholesale.
- `backend/go.work` and `backend/go.work.sum` remove existing backend workspace
  content from `dev`. They are not needed for the frontend restore.
- `docs/01-micro-tasks.md` changes the User App Frontend section to
  `Completed`, but also flips many existing completed services back to
  `Pending`. Do not modify this file in this runbook.

## 4. Build The Auto-Restore File List

```bash
git diff --name-only --diff-filter=AM dev..feature/user-app-frontend -- \
  frontend \
  'TaskImplementation/User App Frontend' \
  > /tmp/user-app-frontend-auto-restore.txt

test -z "$(grep -Fx 'docs/01-micro-tasks.md' /tmp/user-app-frontend-auto-restore.txt || true)" || { echo "Protected docs/01-micro-tasks.md leaked into restore list. Stop."; exit 1; }

sed -n '1,260p' /tmp/user-app-frontend-auto-restore.txt
```

What this does:

- Lists User App Frontend related files that are added or modified.
- Excludes deleted files.
- Includes the frontend workspace, proto-client package, user app source, and
  User App Frontend task notes.
- Excludes protected or shared files:
  `docs/01-micro-tasks.md`, `api/master-api.json`, `backend/go.work`,
  `backend/go.work.sum`, and `TaskImplementation/Dependency/*`.

If this file is empty, stop and re-check the path filters before continuing.

## 5. Restore Added And Modified User App Frontend Files

```bash
if [ -s /tmp/user-app-frontend-auto-restore.txt ]; then
  git restore --overlay --source=feature/user-app-frontend --worktree --pathspec-from-file=/tmp/user-app-frontend-auto-restore.txt
else
  echo "No User App Frontend files to auto-restore. Stop and re-check the file list."
  exit 1
fi
```

What this does:

- Copies selected new files from `feature/user-app-frontend`.
- Updates selected existing files only when they differ from `dev`.
- Uses `--overlay` so Git does not remove `dev` files.
- Does not restore `docs/01-micro-tasks.md`.

## 6. Handle Shared Files Manually

Do not auto-restore shared files from the feature branch.

Review the shared files after the auto-restore:

```bash
git diff -- api/master-api.json
git diff -- backend/go.work
git diff -- backend/go.work.sum
git diff -- docs/01-micro-tasks.md
```

Expected result in the current repo state:

- `api/master-api.json` should have no working-tree diff.
- `backend/go.work` should have no working-tree diff.
- `backend/go.work.sum` should have no working-tree diff.
- `docs/01-micro-tasks.md` should have no working-tree diff.

If `api/master-api.json` has a required User App Frontend addition in a future
branch revision, apply only that addition manually and keep every existing
`dev` entry. Do not accept hunks that remove Superadmin, Search, or other
existing API definitions.

If `backend/go.work` or `backend/go.work.sum` changed, stop and review why.
User App Frontend files under `frontend` should not need backend Go workspace
changes.

If `docs/01-micro-tasks.md` changed, stop before staging. This runbook must not
modify that file.

## 7. Confirm No Deletions Were Applied

```bash
git diff --name-status --diff-filter=D --exit-code
test -z "$(git diff --name-only -- docs/01-micro-tasks.md)" || { echo "docs/01-micro-tasks.md changed. Stop."; exit 1; }
```

What this does:

- Confirms the working tree has no deletions.
- Confirms the protected micro-task file is unchanged.

The deletion check must print nothing. If it prints any `D` entry, stop before
staging.

## 8. Show Status And Diff Stat Before Staging

```bash
if [ -s /tmp/user-app-frontend-auto-restore.txt ]; then
  git add -N --pathspec-from-file=/tmp/user-app-frontend-auto-restore.txt
fi

git status
git diff --stat
git diff --name-status --diff-filter=D --exit-code
test -z "$(git diff --name-only -- docs/01-micro-tasks.md)" || { echo "docs/01-micro-tasks.md changed. Stop."; exit 1; }
```

What this does:

- Marks new auto-restored files as intent-to-add so `git diff --stat` shows
  them.
- Shows the required `git status`.
- Shows the required `git diff --stat`.
- Confirms again that no deletion is present.
- Confirms again that `docs/01-micro-tasks.md` is unchanged.

## 9. Stage Only The Selected Changes

```bash
test -z "$(git diff --name-only -- docs/01-micro-tasks.md)" || { echo "docs/01-micro-tasks.md changed. Stop."; exit 1; }

if [ -s /tmp/user-app-frontend-auto-restore.txt ]; then
  git add --pathspec-from-file=/tmp/user-app-frontend-auto-restore.txt
fi

git status
git diff --cached --stat
git diff --cached --name-status --diff-filter=D --exit-code
test -z "$(git diff --cached --name-only -- docs/01-micro-tasks.md)" || { echo "docs/01-micro-tasks.md is staged. Stop."; exit 1; }
```

What this does:

- Stages only the User App Frontend auto-restore file list.
- Does not stage `docs/01-micro-tasks.md`.
- Does not stage `api/master-api.json`, `backend/go.work`, or
  `backend/go.work.sum` unless they were intentionally handled in a separate,
  reviewed command.
- Confirms no deletions are staged.

The staged deletion check must print nothing.

## 10. Commit

Before committing, run one final staged verification:

```bash
git diff --cached --name-status
git diff --cached --name-status --diff-filter=D --exit-code
test -z "$(git diff --cached --name-only -- docs/01-micro-tasks.md)" || { echo "docs/01-micro-tasks.md is staged. Stop."; exit 1; }
```

Now commit with the exact required message:

```bash
git commit -m "Merge User App Frontend work into dev"
```

After committing, verify the commit:

```bash
git status --short --branch
git log --oneline -1
git show --stat --oneline HEAD
test -z "$(git diff-tree --no-commit-id --name-only -r HEAD -- docs/01-micro-tasks.md)" || { echo "docs/01-micro-tasks.md was committed. Stop."; exit 1; }
git diff-tree --no-commit-id --name-status -r HEAD --diff-filter=D --exit-code
```

What this does:

- Creates the required selective integration commit.
- Confirms the commit message is correct.
- Confirms the protected micro-task file was not committed.
- Confirms the commit contains no deletions.

## 11. Fast-Forward Dev

```bash
git switch dev
git merge --ff-only chore/User-App-Frontend-selective
```

What this does:

- Moves `dev` to the selective integration commit.
- Does not merge the source branch.
- Fails safely if `dev` changed and cannot fast-forward.

## 12. Final Verification

```bash
git status --short --branch
git log --oneline -1
git show --stat --oneline HEAD
test -z "$(git show --name-only --format= HEAD -- docs/01-micro-tasks.md)" || { echo "docs/01-micro-tasks.md is present in the final commit. Investigate."; exit 1; }
git diff --name-status HEAD^..HEAD --diff-filter=D --exit-code
```

What this does:

- Confirms `dev` is on the new selective integration commit.
- Shows the final committed file summary.
- Confirms the latest commit message is:

```text
Merge User App Frontend work into dev
```

- Confirms `docs/01-micro-tasks.md` was not included.
- Confirms the final commit contains no deletions.
