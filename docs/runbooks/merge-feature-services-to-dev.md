# Merge Feature Service Branches Into Dev Runbook

## 1. Purpose

This runbook explains how to safely bring microservice code, frontend apps, documentation, configuration, and related files from multiple `feature/*` branches into one common branch named `dev`.

Use this when service work was developed on separate branches like `feature/user-service`, `feature/order-service`, `feature/product-service`, `feature/auth-service`, `feature/user-app-frontend`, and similar service branches.

The main rule is simple:

- Add files or folders that do not exist in `dev`.
- Replace/update files in `dev` only when the same path belongs to the service branch being integrated.
- Do not delete unrelated files from `dev`.
- Do not blindly overwrite the whole project.
- Review shared files manually when multiple branches changed the same file.

## 2. Important Safety Rules

- Start only when your working tree is clean.
- Fetch the latest branch information before comparing or merging.
- Work on `dev`, not directly on `main` or `master`.
- Never force push.
- Never delete feature branches until the final code is verified.
- Create a backup branch before big integration work.
- Integrate one service branch at a time.
- Commit after each service integration when possible.
- If a file is changed by multiple service branches, review it manually before finalizing.
- Do not push `dev` until all services are integrated and verified.

## 3. Project Structure Found

This repository already has a `docs/` folder, so this runbook is stored under `docs/runbooks/`.

Current important folders and files:

| Path | Purpose |
|---|---|
| `docs/` | Main project documentation |
| `api/master-api.json` | Shared API reference |
| `database/` | Database design files |
| `backend/go.work` | Go workspace file |
| `backend/services/` | Backend microservice folders |
| `frontend/` | Frontend workspace and apps from frontend branches |
| `TaskImplementation/` | Task documents for each service |

Local `feature/*` branches found during inspection:

- `feature/CMS-service`
- `feature/Notification-service`
- `feature/api-gateway-service`
- `feature/auth-service`
- `feature/cart-service`
- `feature/order-service`
- `feature/payment-service`
- `feature/product-service`
- `feature/recommendation-service`
- `feature/search-service`
- `feature/seller-dashboard-cms`
- `feature/session-analytics-dashboard-service`
- `feature/session-management-service`
- `feature/superadmin-service`
- `feature/superadmin_panel`
- `feature/user-app-frontend`
- `feature/user-service`
- `feature/wishlist-service`

Note: a non-feature branch named `a` also exists, but it is not treated as a service branch in this runbook.

## 4. Check Current Git Status

Command:

```bash
git status
```

What this command is used for:

- Shows the current branch.
- Shows whether files are modified, staged, untracked, or conflicted.
- Helps confirm that you are starting from a clean working tree.

When to run it:

- Before fetching, merging, copying files, committing, or pushing.
- After every service integration.

Expected output/result:

```text
On branch dev
nothing to commit, working tree clean
```

What to do if something goes wrong:

- If you see modified files, either commit them, stash them, or stop and review them.
- If you see conflict messages, finish or abort the existing merge before starting new work.
- If you are not on `dev`, switch to `dev` before integrating services.

## 5. Fetch Latest Branches From Remote

Command:

```bash
git fetch --all --prune
```

What this command is used for:

- Downloads the latest branch and commit information from all remotes.
- It does not merge code into your current branch.
- `--all` means fetch from every configured remote.
- `--prune` removes stale remote-tracking references for branches that were deleted on remote.

When to run it:

- Before listing branches.
- Before comparing `dev` with any feature branch.
- Before merging or checking out files from a remote branch.

Expected output/result:

- Git may show updated remote branch names.
- If everything is already updated, it may print little or no output.

What to do if something goes wrong:

- If Git says the remote cannot be reached, check internet/VPN/SSH credentials.
- If Git asks for credentials, log in with the correct Git provider account.
- If fetch fails, do not continue with old branch information unless you intentionally want to work offline.

## 6. List All Branches

Command:

```bash
git branch
```

What this command is used for:

- Lists local branches only.
- The current branch is marked with `*`.

When to run it:

- After fetching.
- Before deciding whether a feature branch exists locally.

Expected output/result:

```text
* dev
  feature/user-service
  feature/order-service
  main
```

What to do if something goes wrong:

- If a feature branch is missing locally, check remote branches with `git branch -r`.
- If you are on the wrong branch, switch before merging or copying files.

Command:

```bash
git branch -r
```

What this command is used for:

- Lists remote-tracking branches like `origin/dev` and `origin/feature/user-service`.

When to run it:

- When a branch does not exist locally.
- After running `git fetch --all --prune`.

Expected output/result:

```text
origin/dev
origin/feature/user-service
origin/feature/order-service
origin/main
```

What to do if something goes wrong:

- If the expected remote branch is missing, confirm the branch name on GitHub/GitLab/remote.
- Run `git fetch --all --prune` again if branch information looks old.

Command:

```bash
git branch -a
```

What this command is used for:

- Lists both local and remote-tracking branches.

When to run it:

- When you want a complete branch view.

Expected output/result:

```text
* dev
  feature/user-service
  remotes/origin/dev
  remotes/origin/feature/user-service
```

What to do if something goes wrong:

- If you see only local branches and no remote branches, check that `origin` is configured.
- If branch names are different from this runbook, use the actual names shown by Git.

Command:

```bash
git for-each-ref --format='%(refname:short)' refs/heads/feature refs/remotes/origin/feature
```

What this command is used for:

- Lists only local and remote feature branches.
- This is useful when the repository has many non-feature branches.

When to run it:

- Before making the branch checklist.
- Before starting integration work.

Expected output/result:

```text
feature/user-service
feature/order-service
origin/feature/user-service
origin/feature/order-service
```

What to do if something goes wrong:

- If the output is empty, there may be no branches under `feature/*`.
- Run `git branch -a` and check if the project uses a different naming pattern.

## 7. Create or Switch to `dev` Branch

Use only one of these cases.

### Case A: `dev` already exists locally

Command:

```bash
git checkout dev
```

What this command is used for:

- Switches your working tree to the local `dev` branch.

When to run it:

- When `git branch` shows `dev`.

Expected output/result:

```text
Switched to branch 'dev'
```

What to do if something goes wrong:

- If Git refuses because of uncommitted changes, run `git status` and commit/stash/review those changes first.

Command:

```bash
git pull origin dev
```

What this command is used for:

- Updates local `dev` with the latest remote `origin/dev`.

When to run it:

- After switching to `dev`.
- Before creating the backup branch.

Expected output/result:

- Git may say `Already up to date.`
- Or Git may show files/commits that were updated.

What to do if something goes wrong:

- If there are conflicts, resolve them before starting service integration.
- If the remote branch does not exist, use Case C below.

### Case B: `dev` exists only on remote

Command:

```bash
git checkout -b dev origin/dev
```

What this command is used for:

- Creates a local `dev` branch from `origin/dev`.
- Switches to the new local `dev` branch.

When to run it:

- When `git branch -r` shows `origin/dev`, but `git branch` does not show local `dev`.

Expected output/result:

```text
Switched to a new branch 'dev'
branch 'dev' set up to track 'origin/dev'
```

What to do if something goes wrong:

- If `origin/dev` does not exist, use Case C below.
- If Git says local `dev` already exists, use Case A.

### Case C: `dev` does not exist anywhere

Command:

```bash
git checkout -b dev
```

What this command is used for:

- Creates a new local `dev` branch from your current branch.
- Switches to that new branch.

When to run it:

- Only when neither local `dev` nor `origin/dev` exists.
- Make sure you are starting from the correct base branch first, usually `main`.

Expected output/result:

```text
Switched to a new branch 'dev'
```

What to do if something goes wrong:

- If you created `dev` from the wrong branch, stop before committing and ask the team which branch should be the base.
- Do not push this new `dev` branch until it is reviewed.

## 8. Create a Backup Branch Before Integration

Command:

```bash
git checkout dev
```

What this command is used for:

- Makes sure the backup branch is created from `dev`.

When to run it:

- Immediately before creating the backup branch.

Expected output/result:

```text
Switched to branch 'dev'
```

What to do if something goes wrong:

- If uncommitted changes block checkout, run `git status` and clean up first.

Command:

```bash
git checkout -b backup/dev-before-service-merge
```

What this command is used for:

- Creates a local backup branch pointing to the current `dev` state.
- This gives you a safe restore point before large merge work.

When to run it:

- Once before integrating the first service branch.

Expected output/result:

```text
Switched to a new branch 'backup/dev-before-service-merge'
```

What to do if something goes wrong:

- If the branch already exists, create a dated backup name like `backup/dev-before-service-merge-2026-06-09`.

Command:

```bash
git checkout dev
```

What this command is used for:

- Returns you to `dev` after creating the backup branch.

When to run it:

- Immediately after creating the backup branch.

Expected output/result:

```text
Switched to branch 'dev'
```

What to do if something goes wrong:

- If Git refuses to switch, check `git status` and resolve the reason before continuing.

## 9. Compare Feature Branch With `dev`

Use this process for every feature branch before integrating it.

Command:

```bash
git diff --name-status dev..feature/user-service
```

What this command is used for:

- Shows which files are different in `feature/user-service` compared with `dev`.
- Shows whether files were added, modified, deleted, or renamed.

When to run it:

- Before merging or copying files from a feature branch.

Expected output/result:

```text
A       backend/services/user-service/go.mod
M       backend/go.work
D       old/file/path.go
R100    old-name.go    new-name.go
```

Meaning of status letters:

| Letter | Meaning | What to do |
|---|---|---|
| `A` | Added | Usually safe to bring if it belongs to that service |
| `M` | Modified | Review before replacing the `dev` version |
| `D` | Deleted | Do not delete from `dev` unless the deletion is intentional |
| `R` | Renamed | Review old and new path carefully |

What to do if something goes wrong:

- If Git says the branch does not exist, run `git branch -a` and use the correct branch name.
- If the output includes unrelated files, prefer selective checkout instead of full merge.
- If many shared files appear, review them manually.

Command:

```bash
git diff --name-status dev..origin/feature/user-service
```

What this command is used for:

- Compares `dev` with the remote-tracking branch.
- Useful when you do not have a local copy of the feature branch.

When to run it:

- After `git fetch --all --prune`.
- When local `feature/user-service` is missing or stale.

Expected output/result:

- Same type of output as the local branch compare command.

What to do if something goes wrong:

- If `origin/feature/user-service` does not exist, confirm the branch name on remote.

Command:

```bash
git log --oneline dev..feature/user-service
```

What this command is used for:

- Shows commits that exist in the feature branch but not in `dev`.

When to run it:

- Before integrating each feature branch.
- When you want to understand what work the branch contains.

Expected output/result:

```text
abc1234 Add user profile update usecase
def5678 Add user service migrations
```

What to do if something goes wrong:

- If there is no output, the branch may already be merged into `dev`.
- If commit messages show unrelated work, use selective checkout.

Command:

```bash
git diff --stat dev..feature/user-service
```

What this command is used for:

- Shows a short summary of changed files and line counts.

When to run it:

- Before deciding whether to merge the whole branch or copy selected paths.

Expected output/result:

```text
 backend/services/user-service/go.mod | 10 ++++++++++
 backend/go.work                      |  1 +
 2 files changed, 11 insertions(+)
```

What to do if something goes wrong:

- If the diff is unexpectedly large, inspect file paths with `git diff --name-status`.

## 10. Recommended Integration Strategy

There are two safe approaches.

### Approach A: Merge branch into `dev`

Use this when the feature branch contains only one service and has clean history.

Command:

```bash
git checkout dev
```

What this command is used for:

- Makes sure the merge happens into `dev`.

When to run it:

- Before running `git merge`.

Expected output/result:

```text
Switched to branch 'dev'
```

What to do if something goes wrong:

- If checkout fails, run `git status` and clean up local changes first.

Command:

```bash
git merge --no-ff feature/user-service
```

What this command is used for:

- Merges the full feature branch into `dev`.
- `--no-ff` creates a merge commit even when fast-forward is possible.
- This keeps a visible record that the service branch was integrated.

When to run it:

- After comparing the branch and confirming it contains only relevant service work.

Expected output/result:

```text
Merge made by the 'ort' strategy.
 backend/services/user-service/go.mod | 10 ++++++++++
```

What conflict output can look like:

```text
CONFLICT (content): Merge conflict in backend/go.work
Automatic merge failed; fix conflicts and then commit the result.
```

What to do if something goes wrong:

- If conflicts happen, follow Section 14.
- If the branch brings unrelated files, abort the merge with `git merge --abort` before committing.
- If you already committed the wrong merge, stop and review with the team before undoing anything.

### Approach B: Checkout specific files/folders from feature branch into `dev`

Use this when only service-specific code/docs should come into `dev`, or when the feature branch contains unrelated changes.

Command:

```bash
git checkout feature/user-service -- backend/services/user-service
```

What this command is used for:

- Copies only `backend/services/user-service` from the feature branch into the current `dev` working tree.
- It does not switch branches.

When to run it:

- When the service folder should be added or replaced in `dev`.

Expected output/result:

- Usually no output.
- `git status` will show added or modified files.

What to do if something goes wrong:

- If Git says the path does not exist, check the real path with `git diff --name-status dev..feature/user-service`.
- If Git refuses because local changes would be overwritten, stop and run `git status`.

Command:

```bash
git checkout feature/user-service -- "TaskImplementation/User Service"
```

What this command is used for:

- Copies the task documentation folder for the user service from the feature branch.
- Quotes are needed because the path contains spaces.

When to run it:

- When you want to bring service-specific task docs into `dev`.

Expected output/result:

- Usually no output.
- `git status` will show added or modified documentation files.

What to do if something goes wrong:

- If Git says the path is unknown, run `git diff --name-status dev..feature/user-service` and copy the exact path.

Command:

```bash
git checkout feature/user-service -- docs/01-micro-tasks.md
```

What this command is used for:

- Replaces `docs/01-micro-tasks.md` in your `dev` working tree with the version from `feature/user-service`.

When to run it:

- Only after manual review, because many feature branches changed this same file.

Expected output/result:

- Usually no output.
- `git diff -- docs/01-micro-tasks.md` will show the replacement.

What to do if something goes wrong:

- If more than one feature branch changed this file, do not choose a version blindly.
- Compare branch versions and manually combine the needed content.

## 11. Suggested Step-by-Step Service Integration Flow

Repeat this flow for every service branch.

Example branch: `feature/user-service`.

Command:

```bash
git checkout dev
```

What this command is used for:

- Ensures you are integrating into `dev`.

When to run it:

- At the start of each service integration.

Expected output/result:

```text
Switched to branch 'dev'
```

What to do if something goes wrong:

- If local changes block checkout, inspect with `git status`.

Command:

```bash
git status
```

What this command is used for:

- Confirms the working tree is clean before starting the next service.

When to run it:

- Before each compare, merge, checkout, and commit.

Expected output/result:

```text
nothing to commit, working tree clean
```

What to do if something goes wrong:

- If files are modified from a previous service, commit or review them before continuing.

Command:

```bash
git diff --name-status dev..feature/user-service
```

What this command is used for:

- Lists file-level changes in the service branch.

When to run it:

- Before choosing merge or selective checkout.

Expected output/result:

- A list of `A`, `M`, `D`, or `R` file changes.

What to do if something goes wrong:

- If unrelated files appear, use selective checkout.

Command:

```bash
git log --oneline dev..feature/user-service
```

What this command is used for:

- Shows commits in the feature branch that are not in `dev`.

When to run it:

- Before integrating the service.

Expected output/result:

- A short commit list.

What to do if something goes wrong:

- If there is no output, the branch may already be integrated.

Choose one method.

Merge method command:

```bash
git merge --no-ff feature/user-service
```

What this command is used for:

- Brings the whole service branch into `dev`.

When to run it:

- Only when the branch is clean and service-specific.

Expected output/result:

- A merge commit is created, or conflicts are shown.

What to do if something goes wrong:

- Resolve conflicts or abort with `git merge --abort`.

Selective checkout method command:

```bash
git checkout feature/user-service -- backend/services/user-service "TaskImplementation/User Service"
```

What this command is used for:

- Copies only the user service folder and its task docs into `dev`.

When to run it:

- When the branch contains unrelated or shared files that should not be copied blindly.

Expected output/result:

- Usually no output.
- `git status` will show added/modified files.

What to do if something goes wrong:

- If one path is wrong, copy paths one by one using the exact paths from `git diff --name-status`.

Command:

```bash
git status
```

What this command is used for:

- Shows which files are now staged, unstaged, untracked, or conflicted.

When to run it:

- After merge or selective checkout.

Expected output/result:

- For a clean merge, it may say nothing to commit.
- For selective checkout, it should show files ready to stage.

What to do if something goes wrong:

- If conflict files are listed, resolve them before staging.

Command:

```bash
git diff --stat
```

What this command is used for:

- Shows a summary of changes currently in your working tree.

When to run it:

- After checkout or conflict resolution.
- Before staging.

Expected output/result:

```text
 backend/services/user-service/go.mod | 10 ++++++++++
 1 file changed, 10 insertions(+)
```

What to do if something goes wrong:

- If the summary includes unrelated files, review before staging.

Command:

```bash
git add .
```

What this command is used for:

- Stages all current working tree changes.

When to run it:

- After verifying the files belong to the service being integrated.

Expected output/result:

- Usually no output.

What to do if something goes wrong:

- If you staged too much, use `git status` to inspect and unstage specific files before committing.

Command:

```bash
git commit -m "Merge user service into dev"
```

What this command is used for:

- Commits the integrated service changes into `dev`.

When to run it:

- After staging and verification.

Expected output/result:

```text
[dev abc1234] Merge user service into dev
```

What to do if something goes wrong:

- If Git says there is nothing to commit, the branch may already be merged or no files were staged.
- If hooks/tests fail, fix the issue before retrying the commit.

## 12. Handling Existing Files in `dev`

If the same file already exists in `dev` and the feature branch has the updated version, replace the `dev` working copy with the feature branch version only after review.

Command:

```bash
git diff dev..feature/user-service -- path/to/file
```

What this command is used for:

- Shows the difference for one specific file between `dev` and the feature branch.

When to run it:

- Before replacing an existing file in `dev`.

Expected output/result:

- A patch showing the exact line differences.

What to do if something goes wrong:

- If the file does not exist in one branch, Git may show it as added or deleted.
- Check the exact path with `git diff --name-status dev..feature/user-service`.

Command:

```bash
git checkout feature/user-service -- path/to/file
```

What this command is used for:

- Replaces the file in your current `dev` working tree with the version from `feature/user-service`.

When to run it:

- After deciding that the feature branch version should win.

Expected output/result:

- Usually no output.
- The file becomes modified in your working tree.

What to do if something goes wrong:

- If Git refuses to overwrite local changes, inspect with `git status`.
- Do not force it. Save or review local work first.

Command:

```bash
git diff -- path/to/file
```

What this command is used for:

- Shows what changed in your current working tree after replacement.

When to run it:

- After checking out the file from the feature branch.

Expected output/result:

- A patch showing the replacement now staged for review, but not committed.

What to do if something goes wrong:

- If the diff is not what you expected, restore the file from `dev` before committing or re-checkout the correct version.

## 13. Shared Files That Need Manual Review

During inspection, these files appeared in multiple feature branches. Do not blindly take the last branch version for these files.

| Shared path | Branches touching it | Recommendation |
|---|---|---|
| `docs/01-micro-tasks.md` | Most backend/service branches | Manually combine service task updates |
| `backend/go.work` | Most backend/service branches | Manually ensure every integrated Go module is listed once |
| `backend/go.work.sum` | Several backend/service branches | Regenerate/review after Go workspace integration |
| `api/master-api.json` | Payment, product, session analytics, superadmin | Manually combine API additions |
| `frontend/package.json` | User app, seller dashboard, superadmin panel | Manually combine scripts/dependencies |
| `frontend/pnpm-lock.yaml` | User app, seller dashboard, superadmin panel | Regenerate/review after final frontend package merge |
| `frontend/pnpm-workspace.yaml` | User app, seller dashboard, superadmin panel | Manually ensure all apps/packages are listed |
| `proto/buf.yaml` and `proto/buf.gen.yaml` | Order, recommendation, user | Manually combine proto configuration |
| `Project Work-Assignment 1.pdf` | Many service branches | Choose the correct final PDF intentionally |
| `reports/` files | Many service branches | Review final report content intentionally |
| `backend/services/auth-service` files | Auth, search, session management | Review carefully because non-auth branches also changed auth files |

Command:

```bash
git diff feature/payment-service..feature/product-service -- api/master-api.json
```

What this command is used for:

- Compares one shared file between two feature branches.

When to run it:

- When multiple branches modified the same file and you need to decide what final content should be kept.

Expected output/result:

- A diff showing how the file differs between the two branches.

What to do if something goes wrong:

- If one branch does not contain the file, compare with a branch that does.
- If the diff is large, open the file versions and manually combine the needed sections.

Command:

```bash
git show feature/payment-service:api/master-api.json
```

What this command is used for:

- Prints the version of a file from a specific branch without switching branches.

When to run it:

- When reviewing shared file content before deciding what to copy.

Expected output/result:

- The file content from that branch is printed to the terminal.

What to do if something goes wrong:

- If Git says the path does not exist, confirm the file path with `git diff --name-status dev..feature/payment-service`.

## 14. Handling Conflicts Safely

Conflicts can happen during `git merge`. Git will stop and ask you to resolve them.

Command:

```bash
git status
```

What this command is used for:

- Shows which files are conflicted.

When to run it:

- Immediately after Git reports a merge conflict.

Expected output/result:

```text
both modified:   backend/go.work
```

What to do if something goes wrong:

- Do not commit until all conflicted files are resolved.
- Open each conflicted file and decide the final content.

Conflict markers look like this:

```text
<<<<<<< HEAD
dev branch code
=======
feature branch code
>>>>>>> feature/user-service
```

Meaning:

- `HEAD` is the current `dev` version.
- The bottom section is the incoming feature branch version.
- The final file must not contain `<<<<<<<`, `=======`, or `>>>>>>>`.

Command:

```bash
git diff
```

What this command is used for:

- Shows unresolved conflict blocks and current file changes.

When to run it:

- While resolving conflicts.

Expected output/result:

- Diff output showing conflicted or edited files.

What to do if something goes wrong:

- If the diff is confusing, open the file in an editor and resolve one conflict at a time.

Command:

```bash
git add <resolved-file>
```

What this command is used for:

- Marks one resolved conflict file as fixed.

When to run it:

- After removing conflict markers and saving the correct final content.

Expected output/result:

- Usually no output.

What to do if something goes wrong:

- If Git says the file does not exist, check the path in `git status`.

Command:

```bash
git commit
```

What this command is used for:

- Completes the merge after all conflicts are resolved and staged.

When to run it:

- After `git status` no longer shows unmerged paths.

Expected output/result:

- Git opens the default merge commit message or creates the merge commit.

What to do if something goes wrong:

- If Git says there are unresolved conflicts, run `git status` and finish resolving them.

Command:

```bash
git merge --abort
```

What this command is used for:

- Cancels the current merge and returns the working tree to the state before the merge began.

When to run it:

- When the merge is too risky.
- When many unrelated files are coming in.
- When you want to switch to selective checkout instead.

Expected output/result:

- Usually no output.
- `git status` should return to the pre-merge state.

What to do if something goes wrong:

- If Git says there is no merge to abort, check `git status`.
- If local changes block abort, stop and ask for help before manually deleting files.

Binary files like PDFs cannot be resolved with text conflict markers. During a merge conflict, choose one side intentionally.

Command:

```bash
git checkout --ours -- "Project Work-Assignment 1.pdf"
```

What this command is used for:

- Keeps the current `dev` version of the conflicted PDF during a merge.

When to run it:

- Only during a merge conflict when the `dev` PDF should win.

Expected output/result:

- Usually no output.

What to do if something goes wrong:

- If Git says the file is not conflicted, check `git status`.

Command:

```bash
git checkout --theirs -- "Project Work-Assignment 1.pdf"
```

What this command is used for:

- Keeps the incoming feature branch version of the conflicted PDF during a merge.

When to run it:

- Only during a merge conflict when the feature branch PDF should win.

Expected output/result:

- Usually no output.

What to do if something goes wrong:

- If Git says the file is not conflicted, check `git status`.

## 15. Check Final Project Status

Command:

```bash
git status
```

What this command is used for:

- Confirms whether there are uncommitted changes or conflicts.

When to run it:

- After every service integration.
- Before final commit.
- Before push.

Expected output/result:

```text
On branch dev
nothing to commit, working tree clean
```

What to do if something goes wrong:

- If files are modified, decide whether to stage and commit them.
- If conflicts remain, resolve them before continuing.

Command:

```bash
git log --oneline --graph --decorate --all
```

What this command is used for:

- Shows branch and commit history as a graph.
- Helps verify that feature branch commits or merge commits are now reachable from `dev`.

When to run it:

- After integrating several branches.
- Before final verification.

Expected output/result:

```text
* abc1234 (HEAD -> dev) Merge user service into dev
* def5678 Merge auth service into dev
| * 123abcd (feature/user-service) Add user service
```

What to do if something goes wrong:

- If expected commits are missing, compare that branch again with `git log --oneline dev..feature/name`.
- If the graph is hard to read, inspect one branch at a time.

## 16. Run Project Verification Commands

Only run commands that match files actually present after integration.

At inspection time, current `dev` had `backend/go.work`, but service `go.mod` files live on feature branches. Frontend `package.json` files also live on frontend-related branches. After integrating those branches, run the matching checks below.

### Backend Go workspace checks

Command:

```bash
cd backend
```

What this command is used for:

- Moves into the backend workspace folder.

When to run it:

- Before running Go workspace commands.

Expected output/result:

- No output.

What to do if something goes wrong:

- If the folder does not exist, confirm the repository structure with `ls`.

Command:

```bash
go work sync
```

What this command is used for:

- Syncs workspace dependency versions across Go modules listed in `backend/go.work`.

When to run it:

- After integrating Go service modules and updating `backend/go.work`.

Expected output/result:

- Usually no output if successful.

What to do if something goes wrong:

- If Go says a module path does not exist, update `backend/go.work` or integrate the missing service folder.
- If dependencies cannot download, check network and module names.

Command:

```bash
go test ./...
```

What this command is used for:

- Runs Go tests for packages visible from the current module/workspace.

When to run it:

- After integrating backend services.

Expected output/result:

```text
ok      example/module/package    0.123s
```

What to do if something goes wrong:

- If a package fails to compile, fix the code or missing dependencies.
- If `go.work` references a missing module, add the missing folder or remove the bad workspace entry.

Command:

```bash
cd ..
```

What this command is used for:

- Returns from `backend/` to the repository root.

When to run it:

- After finishing backend verification.

Expected output/result:

- No output.

What to do if something goes wrong:

- Run `pwd` to confirm your current directory.

### Frontend workspace checks

Run these only after integrating frontend branches that add `frontend/package.json`, `frontend/pnpm-workspace.yaml`, and app folders.

Command:

```bash
cd frontend
```

What this command is used for:

- Moves into the frontend workspace folder.

When to run it:

- Before running frontend package commands.

Expected output/result:

- No output.

What to do if something goes wrong:

- If the folder does not exist, the frontend branches may not be integrated yet.

Command:

```bash
pnpm install
```

What this command is used for:

- Installs frontend dependencies from `package.json` and `pnpm-lock.yaml`.

When to run it:

- After integrating frontend package files.

Expected output/result:

- pnpm installs or confirms dependencies are already installed.

What to do if something goes wrong:

- If pnpm is missing, install the pnpm version expected by the project.
- If lockfile conflicts exist, resolve `package.json`, `pnpm-workspace.yaml`, and `pnpm-lock.yaml` first.

Command:

```bash
pnpm run
```

What this command is used for:

- Lists available frontend scripts.

When to run it:

- Before running build/test/typecheck scripts.

Expected output/result:

- A list of scripts from `frontend/package.json`.

What to do if something goes wrong:

- If there is no `package.json`, the frontend branches may not be integrated yet.

Commands found in frontend branch package files:

```bash
pnpm run build
pnpm run typecheck
pnpm run lint
pnpm run build:seller
pnpm run typecheck:seller
pnpm run test:seller
pnpm run superadmin:build
pnpm run superadmin:typecheck
pnpm run superadmin:test
```

What these commands are used for:

- `pnpm run build`: builds the user app when the user app script is present.
- `pnpm run typecheck`: type-checks the user app when the script is present.
- `pnpm run lint`: lints the user app when the script is present.
- `pnpm run build:seller`: builds the seller dashboard when the script is present.
- `pnpm run typecheck:seller`: type-checks the seller dashboard when the script is present.
- `pnpm run test:seller`: runs seller dashboard tests when the script is present.
- `pnpm run superadmin:build`: builds the superadmin panel when the script is present.
- `pnpm run superadmin:typecheck`: type-checks the superadmin panel when the script is present.
- `pnpm run superadmin:test`: runs superadmin panel tests when the script is present.

When to run them:

- After integrating the matching frontend app branch.
- Only if `pnpm run` shows that script in the final merged `frontend/package.json`.

Expected output/result:

- Build/typecheck/lint/test should finish successfully.

What to do if something goes wrong:

- If a script is missing, do not invent a new command. Check the final `package.json`.
- If dependencies are missing, run `pnpm install`.
- If TypeScript/tests fail, fix the code before final push.

Command:

```bash
cd ..
```

What this command is used for:

- Returns from `frontend/` to the repository root.

When to run it:

- After finishing frontend verification.

Expected output/result:

- No output.

What to do if something goes wrong:

- Run `pwd` to confirm your current directory.

### Session analytics dashboard checks

The `feature/session-analytics-dashboard-service` branch contains `frontend/session-analytics-dashboard/package.json` with `build`, `typecheck`, and `test` scripts.

Command:

```bash
cd frontend/session-analytics-dashboard
```

What this command is used for:

- Moves into the standalone session analytics dashboard app.

When to run it:

- After integrating `feature/session-analytics-dashboard-service`.

Expected output/result:

- No output.

What to do if something goes wrong:

- If the folder does not exist, confirm that branch was integrated.

Command:

```bash
pnpm install
```

What this command is used for:

- Installs dashboard dependencies.

When to run it:

- Before running dashboard build/test commands.

Expected output/result:

- pnpm installs dependencies or confirms they are current.

What to do if something goes wrong:

- Resolve dependency or lockfile issues before continuing.

Command:

```bash
pnpm run build
```

What this command is used for:

- Builds the session analytics dashboard.

When to run it:

- After dependencies are installed.

Expected output/result:

- TypeScript and Vite build should complete successfully.

What to do if something goes wrong:

- Fix TypeScript/build errors before final commit or push.

Command:

```bash
pnpm run typecheck
```

What this command is used for:

- Runs TypeScript type checking for the dashboard.

When to run it:

- After integrating or changing dashboard code.

Expected output/result:

- TypeScript should finish without errors.

What to do if something goes wrong:

- Fix reported type errors.

Command:

```bash
pnpm run test
```

What this command is used for:

- Runs the dashboard test suite.

When to run it:

- After typecheck/build passes.

Expected output/result:

- Vitest should report passing tests.

What to do if something goes wrong:

- Fix failing tests or update tests if requirements changed.

Command:

```bash
cd ../..
```

What this command is used for:

- Returns from `frontend/session-analytics-dashboard` to the repository root.

When to run it:

- After finishing session analytics dashboard verification.

Expected output/result:

- No output.

What to do if something goes wrong:

- Run `pwd` to confirm your current directory.

## 17. Final Commit and Push

If you committed after every service, the final commit may not be needed. If there are remaining verified changes, create a final commit.

Command:

```bash
git status
```

What this command is used for:

- Checks whether there are remaining uncommitted changes.

When to run it:

- Before final commit and push.

Expected output/result:

```text
On branch dev
nothing to commit, working tree clean
```

What to do if something goes wrong:

- If files are modified, review them before staging.
- If conflicts remain, resolve them first.

Command:

```bash
git add .
```

What this command is used for:

- Stages all remaining verified changes.

When to run it:

- Only after checking the changes with `git status` and `git diff --stat`.

Expected output/result:

- Usually no output.

What to do if something goes wrong:

- If too many files were staged, review with `git status` before committing.

Command:

```bash
git commit -m "Consolidate microservice branches into dev"
```

What this command is used for:

- Creates the final consolidation commit.

When to run it:

- After all integrations and verification are complete.

Expected output/result:

```text
[dev abc1234] Consolidate microservice branches into dev
```

What to do if something goes wrong:

- If Git says nothing to commit, all changes may already be committed.
- If hooks fail, fix the reported issue and retry.

Command:

```bash
git push origin dev
```

What this command is used for:

- Pushes the final `dev` branch to remote.

When to run it:

- Only after all service code/docs/config are integrated.
- Only after tests/build/typecheck checks are complete.
- Only when you are ready to update remote `dev`.

Expected output/result:

```text
To github.com:owner/repo.git
   oldsha..newsha  dev -> dev
```

What to do if something goes wrong:

- If push is rejected, run `git fetch --all --prune`, review remote updates, and merge/rebase safely.
- Do not force push unless the team explicitly agrees.
- If you are unsure, stop before pushing.

Important: do not push until all services are verified.

## 18. Branch Checklist Table

Fill `Status` as you integrate each branch.

| Service Branch | Service Folder | Docs Path | Integration Method | Status | Notes |
|---|---|---|---|---|---|
| `feature/CMS-service` | `backend/services/cms-service` | `TaskImplementation/CMS Service`, `docs/01-micro-tasks.md` | Merge / Selective Checkout | Pending | Review `backend/go.work` |
| `feature/Notification-service` | `backend/services/notification-service` | `TaskImplementation/Notification Service`, `docs/01-micro-tasks.md` | Merge / Selective Checkout | Pending | Review `backend/go.work` |
| `feature/api-gateway-service` | `backend/services/api-gateway` | `TaskImplementation/API Gateway Service`, `docs/01-micro-tasks.md` | Merge / Selective Checkout | Pending | Review `backend/go.work` and `backend/go.work.sum` |
| `feature/auth-service` | `backend/services/auth-service` | `TaskImplementation/Auth Service`, `TaskImplementation/Platform Foundation`, `docs/01-micro-tasks.md` | Merge / Selective Checkout | Pending | Auth files also appear in search/session branches |
| `feature/cart-service` | `backend/services/cart-service` | `TaskImplementation/Cart Service`, `docs/01-micro-tasks.md` | Merge / Selective Checkout | Pending | Review `backend/go.work` |
| `feature/order-service` | `backend/services/order-service`, `backend/shared/gen/go` | `TaskImplementation/Order Service`, `docs/01-micro-tasks.md` | Merge / Selective Checkout | Pending | Review proto files and `backend/go.work.sum` |
| `feature/payment-service` | `backend/services/payment-service` | `TaskImplementation/Payment Service`, `docs/01-micro-tasks.md`, `api/master-api.json` | Merge / Selective Checkout | Pending | Review shared API file |
| `feature/product-service` | `backend/services/product-service` | `TaskImplementation/Product Service`, `docs/01-micro-tasks.md`, `api/master-api.json` | Merge / Selective Checkout | Pending | Review shared API file |
| `feature/recommendation-service` | `backend/services/recommendation-service`, `backend/proto-gen` | `TaskImplementation/Recommendation Service`, `docs/01-micro-tasks.md` | Merge / Selective Checkout | Pending | Review proto/generated Go paths |
| `feature/search-service` | `backend/services/search-service`, `backend/services/auth-service` | `TaskImplementation/Search Service`, `docs/01-micro-tasks.md` | Selective Checkout recommended | Pending | Review why auth-service files are included |
| `feature/seller-dashboard-cms` | `frontend/seller-dashboard`, `frontend/package.json`, `frontend/pnpm-workspace.yaml` | `TaskImplementation/Seller Dashboard (CMS)` | Selective Checkout recommended | Pending | Merge frontend workspace files manually |
| `feature/session-analytics-dashboard-service` | `backend/services/session-service`, `frontend/session-analytics-dashboard` | `TaskImplementation/Session Analytics Dashboard Service`, `api/master-api.json` | Selective Checkout recommended | Pending | Includes backend and frontend app |
| `feature/session-management-service` | `backend/services/session-service`, `backend/services/auth-service` | `TaskImplementation/Session Management Service`, `docs/01-micro-tasks.md` | Selective Checkout recommended | Pending | Review why auth-service files are included |
| `feature/superadmin-service` | `backend/services/superadmin-service` | `TaskImplementation/Superadmin Service`, `docs/01-micro-tasks.md`, `api/master-api.json` | Merge / Selective Checkout | Pending | Review shared API file |
| `feature/superadmin_panel` | `frontend/superadmin-panel`, `frontend/package.json`, `frontend/pnpm-workspace.yaml` | `TaskImplementation/Superadmin Panel` | Selective Checkout recommended | Pending | Merge frontend workspace files manually |
| `feature/user-app-frontend` | `frontend/user-app`, `frontend/packages`, `frontend/package.json`, `frontend/pnpm-workspace.yaml` | `TaskImplementation/User App Frontend`, `TaskImplementation/Dependency`, `docs/01-micro-tasks.md` | Selective Checkout recommended | Pending | Merge frontend workspace files manually |
| `feature/user-service` | `backend/services/user-service`, `backend/services/api-gateway`, `backend/shared/gen/go`, `backend/shared/validation` | `TaskImplementation/User Service`, `TaskImplementation/Platform Foundation`, `docs/01-micro-tasks.md` | Selective Checkout recommended | Pending | Review api-gateway/shared paths |
| `feature/wishlist-service` | `backend/services/wishlist-service` | `TaskImplementation/Wishlist Service`, `docs/01-micro-tasks.md` | Merge / Selective Checkout | Pending | Review `backend/go.work` and `backend/go.work.sum` |

## 19. Copy-Paste Template For Each Branch

Replace `feature/user-service`, service paths, and docs paths with the branch you are integrating.

Command:

```bash
git checkout dev
```

What this command is used for:

- Switches to the integration branch.

When to run it:

- Before every service integration.

Expected output/result:

- You are on `dev`.

What to do if something goes wrong:

- Resolve local changes first.

Command:

```bash
git status
```

What this command is used for:

- Confirms the working tree is clean.

When to run it:

- Before comparing or copying.

Expected output/result:

- Clean working tree.

What to do if something goes wrong:

- Commit, stash, or review local changes.

Command:

```bash
git diff --name-status dev..feature/user-service
```

What this command is used for:

- Shows exactly what the branch changes.

When to run it:

- Before integration.

Expected output/result:

- File list with `A`, `M`, `D`, or `R`.

What to do if something goes wrong:

- Use the actual branch name from `git branch -a`.

Command:

```bash
git checkout feature/user-service -- backend/services/user-service "TaskImplementation/User Service"
```

What this command is used for:

- Copies selected service code and docs from the feature branch into `dev`.

When to run it:

- When selective checkout is safer than merging the whole branch.

Expected output/result:

- Files are copied into the working tree.

What to do if something goes wrong:

- Copy one path at a time and verify each path exists in the branch.

Command:

```bash
git status
```

What this command is used for:

- Shows copied files.

When to run it:

- After selective checkout.

Expected output/result:

- Added/modified files for that service.

What to do if something goes wrong:

- If unrelated files appear, review before staging.

Command:

```bash
git diff --stat
```

What this command is used for:

- Summarizes the working tree changes.

When to run it:

- Before staging.

Expected output/result:

- Summary should match the service being integrated.

What to do if something goes wrong:

- Review unexpected files before committing.

Command:

```bash
git add .
```

What this command is used for:

- Stages verified files.

When to run it:

- After reviewing changes.

Expected output/result:

- Usually no output.

What to do if something goes wrong:

- Use `git status` to inspect staged files before committing.

Command:

```bash
git commit -m "Merge user service into dev"
```

What this command is used for:

- Saves the service integration in Git history.

When to run it:

- After staging only the intended files.

Expected output/result:

- A new commit on `dev`.

What to do if something goes wrong:

- If there is nothing to commit, verify whether the service was already integrated.

## 20. Final Summary

Safest recommended process:

1. Fetch latest branches.
2. Checkout `dev`.
3. Create a backup branch from `dev`.
4. Compare one feature branch at a time.
5. Bring only service-specific files unless the whole branch is clean.
6. Manually review shared files like `backend/go.work`, `docs/01-micro-tasks.md`, `api/master-api.json`, and frontend workspace files.
7. Resolve conflicts openly and carefully.
8. Run backend/frontend verification commands that match the final project files.
9. Commit after each service when possible.
10. Push `dev` only after final verification.

Do not delete feature branches. Do not force push. Do not remove existing `dev` files unless the same path is intentionally replaced by the service branch version.
