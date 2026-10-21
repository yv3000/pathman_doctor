# Roadmap

This file tracks the work triggered by three datafactor repository reports: the
first scored 40 (grade D) and drove the v1.2.x pass; the second, on 2026-10-20,
scored 68 (grade C); the third, on 2026-10-21, scored 77 (grade B).

## 0. Third audit: score 77

Seven of eight dimensions are at or above 75. This pass targets the two that are
not, plus the CI visibility items in the report.

| Dimension | Score | Change in this pass |
|---|---|---|
| Architecture & robustness | 72 | `internal.Logf` gives diagnostics a leveled sink instead of only colored writes |
| Test coverage | 78 | backup fallback and write failure, `-v`/`-h`, doctor help covered; profile uploaded as an artifact; total 89.4% |
| Security hygiene | 80 | `--only` moved next to `ValidateEntry`, so both flags share one validation boundary |
| Docs & onboarding | 85 | README documents `PATHMAN_LOG` and the uploaded profile |
| Code cleanliness | 85 | unchanged, already above target |
| History & maintenance | 55 | `v1.3.0` tagged with release notes; sustained cadence is still real work over weeks (section 4) |
| Dependency health | 80 | unchanged, above target |
| CI/CD maturity | 85 | coverage artifact on every run; weekly cold-cache run with module caching disabled |

Report fix list from the third audit:

| # | Fix | Status |
|---|---|---|
| 03 | Add a second contributor or a sustained cadence signal | Partial — `v1.3.0` has release notes and a changelog; a second human has to contribute under their own identity |
| 04 | Make input validation explicit and consistently checked | Done — `ValidateEntry` and `ValidateOnlyFlag` live in `internal/validate.go`, are called from `parseFixArgs`, and have table tests |
| 05 | Raise coverage gate visibility and extend breadth | Done — `coverage.out` is uploaded as an artifact; the dispatch and backup branches are covered. The only thin function left is the detached `cmd.exe` cleanup, which a unit test cannot safely spawn |
| 06 | Add a structured logging seam | Done — `internal.Logf` plus `App.Log`, opt-in with `PATHMAN_LOG=1`; tests assert the level tags and that the console stays clean |
| 07 | Confirm fresh-clone reproducibility end to end | Done — the CI schedule runs the same jobs weekly with caching disabled |
| 01 | Ship features with their tests in small focused commits | Applied to every commit in this pass |
| 02 | Keep developing over time, in real increments | Cannot be faked; see section 4 |

## 0.1 Second audit: score 68

Five dimensions were below 75. This pass targets them with code and CI changes,
not with cosmetic commits.

| Dimension | Score | Change in this pass |
|---|---|---|
| Architecture & robustness | 62 | `internal.ValidateEntry` gives one validation rule; `openRegistryKey` makes the registry layer injectable |
| Test coverage | 62 | coverage gate in CI; fake registry key covers reads, writes, missing values and errors; uninstall branches covered |
| Security hygiene | 68 | `--entry` is validated before PATH is read; malformed values exit 2 |
| Docs & onboarding | 82 | README and CONTRIBUTING document the gate and the opt-in live check |
| Code cleanliness | 80 | unchanged, already above target |
| History & maintenance | 48 | real work split into focused, tested commits; sustained history cannot be created in one session (section 4) |
| Dependency health | 72 | Dependabot groups routine updates; `govulncheck` pinned instead of `@latest` |
| CI/CD maturity | 78 | the test step now enforces coverage instead of only reporting it |

Report fix list from the second audit:

| # | Fix | Status |
|---|---|---|
| 03 | Enforce a coverage threshold in CI | Done — `scripts/coverage-gate.ps1` runs the suite and fails below 80% for `cmd`, `internal` and the module total; `ci.yml` calls it |
| 04 | Broaden integration-style tests for registry and uninstall paths | Done — fake registry key covers `ReadRaw`/`ReadPATH`/`WritePATH` success, missing value, open error and write error; uninstall covers cleanup ordering, a failed PATH write and an absent entry |
| 05 | Add input validation at the flag-parsing boundary | Done — `internal.ValidateEntry`, called by `parseFixArgs`, with table-driven tests |
| 06 | Grow commit history with paired feature and test commits | Ongoing — each change in this pass is its own commit with its tests; real increments over time remain the only honest fix |
| 07 | Document and verify a fully offline test run | Done — default tests use a fake store and fake registry key; `TestReadPATHUserIsReadable` is opt-in with `PATHMAN_LIVE_REGISTRY=1` |
| 01 | Ship features with their tests in small focused commits | Applied to every commit in this pass |
| 02 | Keep developing the repository over time, in real increments | Cannot be faked; see section 4 |

## 1. First report items and status

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
- [x] End-to-end run on a real machine: registry write with `--entry`
      filters, uninstall and installer (see the verification table).
- [ ] Re-run the datafactor audit after this pass.
- [x] Replace the `@latest` govulncheck install in CI with a pinned version:
      v1.9.0 is installed and reports no vulnerabilities on Go 1.26.1.
- [x] Make the registry layer testable without the live hive: `openRegistryKey`
      is a seam and the registry tests use a fake key.
- [x] Tag `v1.3.0` with release notes after the third audit.
- [x] Upload the coverage profile and run a weekly cold-cache build in CI.
- [ ] Optional: `doctor --json` for scripting and `fix --dry-run` for CI use.

## 6. Verification performed

| Check | Command | Result |
|---|---|---|
| Build | `go build ./...` | passes |
| Vet | `go vet ./...` | passes |
| Tests | `go test ./... -cover` | passes; `cmd` 86.0%, `internal` 89.4%, `main` 33.3% |
| Coverage gate | `./scripts/coverage-gate.ps1` | passes; total 87.0%, floors 80/80/80 |
| Negative gate check | `./scripts/coverage-gate.ps1 -MinimumPackage 95` | exits 1 with both packages listed |
| Fresh profile | `LOCALAPPDATA=<empty temp> go test ./... -cover` | passes; the live registry test skips, so no pre-existing PATH or backup state is needed |
| Vulnerability scan | `govulncheck v1.9.0 ./...` | 0 vulnerabilities reachable |
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
| Release pipeline | tags `v1.2.0`, `v1.2.1`, `v1.2.2` | tests ran, amd64/arm64 assets uploaded, `--version` reports the tag |
| Live registry write | `fix --only dup/dead --entry <temp entry>` | both filters removed only the test entry; 66 existing entries unchanged; backups written |
| Live uninstall | `pathman uninstall` on the installed binary | PATH entry removed with backup; the locked folder exposed a defect fixed in v1.2.1 |
| Deferred cleanup | same command string run directly with `cmd /c` | deletes the folder; the agent sandbox kills detached children, so the scheduled run is verified by unit tests plus the direct invocation |
| Installer | `install.ps1` via `irm \| iex` | downloaded v1.2.2 to `~\.pathman\bin`, added the User PATH entry, binary reports `pathman v1.2.2` |

Coverage percentages are from the last local run; CI records them on every push.

Remote checks are now verified as well: CI passed on `main`, the three jobs are
required by branch protection, and the release workflow published v1.2.0 and
v1.2.1. The live uninstall run surfaced the locked-executable defect that
v1.2.1 fixes; the folder deletion itself could only be observed when the exact
`cmd /c` command is run directly, because this agent sandbox terminates the
whole process tree of a finished command, including a detached cleanup child.
