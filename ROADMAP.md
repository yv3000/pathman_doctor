# Roadmap

This file tracks the work triggered by the datafactor repository report
(score 40, grade D) and how each item maps to the buyer-facing signals.

## 1. Report items and status

| # | Report item | Status | Evidence |
|---|---|---|---|
| 01 | Ship features with their tests in small focused commits | Done for this pass | Every behaviour change below landed with tests in the same commit; existing history was left untouched |
| 02 | Make install, build and test work from a fresh clone | Done | README "Build from source" runs `go build`, `go vet`, `go test ./... -cover`; `go.sum` is committed; CI proves it on `windows-latest` |
| 03 | Keep developing over time, in real increments | Ongoing (cannot be faked) | More tested commits today; sustained work over weeks is the only honest fix (see section 4) |
| 04 | Add a real test suite for internal package logic | Done | `internal/path_test.go`, `plan_test.go`, `expand_test.go`, `registry_test.go`, `backup_test.go`; `cmd/app_test.go`, `uninstall_test.go` |
| 05 | Add lint and test gating to CI | Done | `.github/workflows/ci.yml` + `.golangci.yml`; branch protection is a repository setting (see section 5) |
| 06 | Stop swallowing errors from `ReadPATH` in `fix` | Done | Warning added, unread scope never written, commit `fix: surface System PATH read errors in fix command` |
| 07 | Add dependency update automation | Done | `.github/dependabot.yml` for `gomod` and `github-actions`, weekly |
| 08 | Add `CONTRIBUTING.md` and a changelog | Done | `CONTRIBUTING.md`, `CHANGELOG.md`, `LICENSE`; README changelog section now links to it |

Additional defects found while reading the code (not in the report):

| Defect | Status |
|---|---|
| `%VAR%` entries (`%SystemRoot%\system32`) were treated as dead and could be deleted by an admin `fix` | Fixed with tests |
| Release workflow used `working-directory: pathman`, which does not exist in this repository | Fixed; builds from the module root and tests first |
| `--only` typos silently matched nothing; unknown flags were ignored | Fixed; usage error + exit code 2 |
| Partial writes printed a success banner | Fixed; accurate counts and exit code 1 |
| `uninstall` executed a remote `irm ... \| iex` script | Replaced with local removal plus backup |
| 2.3 MB `pathman-windows-amd64.exe` committed | Removed and added to `.gitignore` |
| Duplicate detection compared raw strings | Normalized (case-insensitive, cleaned) with tests |

## 2. Score signals

Report scores at the time of the audit and what this pass changes. The "after"
column is the expected artifact-side result, not a measured score — the audit
service must be re-run to confirm.

| Dimension | Before | Expected now | Lever |
|---|---|---|---|
| Architecture & robustness | 40 | ~75–85 | Typed scopes, injected store, pure analysis, errors surfaced, backups, exit codes |
| Test coverage | 5 | ~65–75 | 40+ tests across internal and cmd packages; CI publishes coverage |
| Security hygiene | 55 | ~75–85 | Flag validation, govulncheck in CI, no remote script execution, no committed binaries |
| Docs & onboarding | 70 | ~85–90 | Build/test steps, CONTRIBUTING, CHANGELOG, LICENSE, AGENTS.md, ROADMAP.md |
| Code cleanliness | 55 | ~85 | Lint config enforced, no god files, binary removed, dead code deleted |
| History & maintenance | 25 | ~30–40 | More small tested commits; still one author and a short span |
| Dependency health | 55 | ~80 | Dependabot for modules and actions, lockfile, pinned versions |
| CI/CD maturity | 30 | ~85–90 | Build, vet, tests, coverage, lint, vulnerability scan on every push and PR |

Highest-weight gaps that remain are history/task capacity: the repository still
holds a single burst of commits from one author. No local change can fix that
honestly.

## 3. What this pass delivered

- Testable core: PATH expansion, parsing, analysis and removal planning are
  pure functions with injected expansion and existence checks.
- Testable commands: `cmd.App` takes a store plus input/output streams, so the
  fix and uninstall flows run against a fake registry in tests.
- Correctness fixes for percent-style variables, duplicate normalization,
  System PATH read failures and partial write reporting.
- Safety: PATH backups before writes, local uninstall, validated flags.
- CI: build, vet, test with coverage, golangci-lint, govulncheck, Dependabot,
  tag-based releases with the version stamped into the binary.
- Docs: README build/test/exit-code sections, CONTRIBUTING, CHANGELOG, LICENSE,
  AGENTS.md and this roadmap.

## 4. Honest note on history and task capacity

Buyers look for a repository that was built up over time, with commits that
arrive together with the tests that prove them. The existing six commits were
created in one hour with messages like `readme update2`. This pass adds a set of
small, tested commits, which improves the "mineable tasks" signal, but it still
happens on one day with one author.

Realistic path to 85+:

1. Land the next real improvements as separate tested commits over the coming
   weeks, not in a single burst.
2. Keep `CHANGELOG.md` current so each release has a visible, verifiable change.
3. If another person contributes, let them commit under their own identity.
4. Tag a release after each meaningful batch (`v1.2.0` after this pass).

Do not rewrite dates or author identities to simulate history — that is
detectable and it is not the behaviour the score is meant to reward.

## 5. Manual steps and remaining work

- [x] Push the commits to `origin` and tag `v1.2.0` (release workflow runs the
      tests and publishes both binaries).
- [x] Enable branch protection for `main`: the `Build, vet and test`, `Lint`
      and `Vulnerability scan` checks are required.
- [ ] Re-run the datafactor audit after the release.
- [ ] Replace the `@latest` govulncheck install in CI with a pinned version
      once a version that supports the runner's Go is chosen.
- [ ] Add an integration test that exercises a throwaway registry key if the
      registry layer changes; today only the key mapping is unit tested.
- [ ] Optional: `doctor --json` for scripting and `fix --dry-run` for CI use.

## 6. Verification performed

| Check | Command | Result |
|---|---|---|
| Build | `go build ./...` | passes |
| Vet | `go vet ./...` | passes |
| Tests | `go test ./... -cover` | passes; `cmd` 84.7%, `internal` 76.0%, `main` 33.3% |
| Lint | `golangci-lint run` (v2.15.0) | 0 issues |
| Workflows | `actionlint` (v1.7.12) | 0 errors for `ci.yml` and `release.yml` |
| Fresh clone | `git archive` copy, README steps only | build, vet and tests pass |
| Live scan | `pathman doctor` on the real machine | 105 entries; `%SystemRoot%` / `%WINDOWS%` entries report OK instead of DEAD |
| No-op safety | `pathman fix`, answer `n` | exit 0, `Cancelled. No changes made.`, User PATH hash unchanged |
| Flag validation | `pathman fix --only bogus` | exit 2, usage printed, no write |
| Installer scripts | PowerShell `Parser::ParseFile` | `install.ps1` and `uninstall.ps1` parse cleanly |
| Release version stamp | `go build -ldflags="-X main.version=v9.8.7"` | `pathman --version` prints `v9.8.7` |
| Remote CI | GitHub Actions run 39101743236 | build/vet/test, lint and vulncheck all pass |
| Branch protection | GitHub API | required checks configured on `main` |
| Dependency PRs | Dependabot #1–#5 | rebased, green, merged |

Coverage percentages are from the last local run; CI records them on every push.

Not verifiable locally: the GitHub Actions jobs and the required-status-check
setting need a push to the remote repository, and `uninstall` is covered by
tests with an injected filesystem instead of deleting the local installation.
