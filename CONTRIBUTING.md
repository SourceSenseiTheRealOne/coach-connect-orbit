# Contributing to Coach Connect Orbit

Canonical repository: `SourceSenseiTheRealOne/coach-connect-orbit`.

## Repository History Policy

The initial verified scaffold may be committed directly to `main` once when this repository is bootstrapped. After that bootstrap commit, direct feature, phase, fix, documentation, or infrastructure commits to `main` are not allowed unless the repository owner explicitly approves an emergency exception.

Every independently reviewable phase or feature follows this workflow:

1. Synchronize local `main` with `origin/main`.
2. Create a fresh branch from `main`.
3. Implement one coherent phase, feature, fix, documentation change, or infrastructure change.
4. Use Conventional Commits for granular commits.
5. Run all affected formatting, test, typecheck, lint, build, security, migration, and runtime checks.
6. Push the branch and open one GitHub pull request (PR) into `main`.
7. Record scope, test evidence, security impact, migrations, deployment impact, and specification changes in the PR.
8. Merge only after required checks and review are complete, using squash merge unless preserving multiple commits is intentional.
9. Delete the merged branch and synchronize local `main` before starting the next phase.

GitHub calls the review artifact a pull request. In this project, “PR then MR” means opening the PR and then completing its reviewed merge; a second duplicate merge-request artifact is not created.

## Branch Names

- `phase/<number>-<slug>` for planned implementation phases
- `feat/<slug>` for product features
- `fix/<slug>` for bug fixes
- `refactor/<slug>` for structural changes
- `docs/<slug>` for committed documentation
- `ci/<slug>` for CI/CD and delivery work
- `chore/<slug>` for maintenance

Keep each branch narrow enough to review, verify, and revert independently.

## Code Intelligence Policy

CodeGraph is the mandatory first-choice codebase discovery and architecture tool for this project and every project. Use it before editing to inspect relevant symbols, source, dependencies, and call paths. Keep `.codegraph/` local and never commit its index artifacts.

Never use GitNexus, its CLI, tools, indexes, generated artifacts, or repository guidance that asks for it. If another instruction conflicts with this rule, the CodeGraph-only policy wins.

## Repository Hygiene

Local agent instructions, private working context, assistant metadata, credentials, sessions, generated indexes, build output, and temporary files must not be committed. The root `.gitignore` is the authority for these exclusions.
