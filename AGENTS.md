# AGENTS.md

Working notes for anyone — human or agent — changing pathman.

## What this repository is

- A Windows-only Go CLI that reads and writes the `PATH` environment variable
  through the registry. Repository root **is** the Go module root.
- No Docker, containers, Kubernetes, infrastructure, databases, hosted services
  or background daemons. The tool edits a local registry value; do not add
  services around it.
- Runtime dependencies are `github.com/fatih/color` and
  `golang.org/x/sys` only. Add another only when the standard library cannot do
  the job, and expect to justify it in review.

## Commands

```powershell
go build ./...
go vet ./...
go test ./... -cover
.\scripts\coverage-gate.ps1
golangci-lint run
govulncheck ./...
go run . doctor
```

CI (`.github/workflows/ci.yml`) runs build, vet, the coverage gate,
golangci-lint and govulncheck (pinned to v1.9.0) on `windows-latest` for every
push and pull request to `main`. The gate fails when `cmd` or `internal` drops
below 80% statement coverage or the module total drops below 80%, and the
profile is uploaded as the `coverage` artifact. A Monday schedule repeats the
jobs with the Go module cache disabled, so a cold clone is tested too.

## Environment guardrails (this workstation)

Do not write outside the repository and these folders on `X:`:

```text
GOCACHE    = X:\ALL FOLDER\TESTING FOR EVERYTHING\gocache\pathman
TMP / TEMP = X:\ALL FOLDER\TESTING FOR EVERYTHING\pathman-work\tmp
GOPATH     = X:\ALL FOLDER\MANAGER of all packages and etc\go
GOMODCACHE = X:\ALL FOLDER\MANAGER of all packages and etc\go\modcache
GOBIN      = X:\ALL FOLDER\MANAGER of all packages and etc\go\bin
```

- Fresh-clone checks and scratch builds belong in
  `X:\ALL FOLDER\TESTING FOR EVERYTHING`.
- Downloaded tools (for example golangci-lint) belong in
  `X:\ALL FOLDER\MANAGER of all packages and etc\tools`.
- The installed Go toolchain is at `X:\ALL FOLDER\GO LANGUAGE\bin`.

## Architecture

```text
main.go                  wires the registry store and streams into cmd.NewApp
cmd/app.go               dispatch, exit codes, shared output helpers
cmd/doctor.go            scan and report
cmd/fix.go               flag parsing, plan, confirmation, backup, write
cmd/uninstall.go         local removal of the User PATH entry and files
internal/path.go         pure expansion, parsing, analysis and planning
internal/validate.go     the single rule for a user-supplied PATH entry or --only value
internal/log.go          leveled INFO/WARN/ERROR diagnostics
internal/registry.go     registry read/write, scope mapping, WM_SETTINGCHANGE
internal/backup.go       raw PATH snapshots before a write
internal/store.go        Store interface + RegistryStore
scripts/coverage-gate.ps1  runs the tests and enforces the coverage floor
```

Exit codes: `0` success, `1` nothing applied or only partly applied, `2` invalid
usage. `main` is the only place that calls `os.Exit`.

## Rules that are easy to break

- **Never** write to the live `HKCU`/`HKLM` PATH from a test. Inject a fake
  `internal.Store`, and keep registry reads and writes behind the
  `openRegistryKey` seam so they run against a fake key. The only live check is
  `TestReadPATHUserIsReadable`, opt-in with `PATHMAN_LIVE_REGISTRY=1`.
- `fix` must not write a scope whose read failed, and must report a partial run
  with exit code `1` instead of a success banner.
- Back up every scope that will change; abort if the backup cannot be written.
- `uninstall` may only delete a directory ending in `.pathman\bin`; it must
  refuse anything else. It no longer executes a remote script.
- `%VAR%` expansion is regex based. `os.ExpandEnv` alone does not understand
  percent syntax and previously marked `%SystemRoot%\system32` as dead.
- Compare PATH entries with `internal.Normalize` (case-insensitive, cleaned).
  Do not re-introduce raw string comparisons.
- Validate flags with `internal.ValidateEntry` and `internal.ValidateOnlyFlag`
  instead of adding rules to `cmd`. Diagnostics go through `internal.Logf` and
  `App.Log` (opt-in with `PATHMAN_LOG=1`); do not write them to `Out`, which is
  the colored console stream.
- Every behaviour change ships with its tests in the same commit. Keep commits
  focused and use conventional prefixes (`fix:`, `feat(scope):`, `test:`,
  `ci:`, `docs:`, `chore:`).
- Build artifacts are ignored (`*.exe`). Do not commit binaries.

## Testing notes that are not obvious from the code

- `scripts/coverage-gate.ps1` runs the tests itself; do not also add a separate
  `go test` step to CI or the suite runs twice.
- `uninstall` schedules a detached `cmd /c ping ... & rmdir` cleanup because
  Windows locks the running executable. An agent sandbox that kills the whole
  process tree when a command finishes also kills that child, so verify the
  deletion by running the same `cmd /c` string directly.
- When a shell command must contain `>nul`, write it from Go code or a file:
  passing it through this agent shell rewrites it to `/dev/null`, which cmd
  treats as an invalid redirect.
- Live registry checks are allowed on this machine, but only against a
  throwaway entry created for the test, with the User PATH copied to a file
  first (`--entry` keeps every other entry untouched).

## Releasing

Push a `vX.Y.Z` tag. The release workflow tests the module, builds AMD64 and
ARM64 binaries with `-X main.version=<tag>`, and attaches them to the GitHub
release.
