<p align="center">
  <a href="https://golang.org"><img src="https://img.shields.io/badge/built%20with-Go-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go" /></a>
  <a href="https://www.microsoft.com/windows"><img src="https://img.shields.io/badge/platform-Windows-0078D6?style=for-the-badge&logo=windows&logoColor=white" alt="Windows" /></a>
  <a href="https://github.com/yv3000/pathman_doctor/actions/workflows/ci.yml"><img src="https://github.com/yv3000/pathman_doctor/actions/workflows/ci.yml/badge.svg" alt="CI" /></a>
  <a href="#license"><img src="https://img.shields.io/badge/license-MIT-green?style=for-the-badge" alt="MIT License" /></a>
</p>

<h1 align="center">pathman</h1>
<p align="center"><b>Your Windows PATH, finally clean.</b></p>
<p align="center">Scan for dead folders and duplicate entries. Fix them in one command. No restart required.</p>

---

## What is pathman?

Over time, your Windows `PATH` environment variable silently accumulates garbage — folders from uninstalled programs that no longer exist, and duplicate entries added by multiple installers. Most tools never clean this up.

**pathman** reads your PATH directly from the Windows Registry, checks every entry, and tells you exactly what's broken:

- **Dead entries** — folders that no longer exist on disk (uninstalled tools, renamed directories)
- **Duplicate entries** — the same path added multiple times by different installers

One command to scan. One command to fix. No restarts required — pathman broadcasts the change to all open windows immediately.

> **Think of it as a doctor for your PATH.** It finds the problems, shows you exactly what's wrong, and cleans it up — without touching anything healthy.

---

## Install

Open **PowerShell** and run:

```powershell
irm https://raw.githubusercontent.com/yv3000/pathman_doctor/main/install.ps1 | iex
```

That's it. **One command. No admin required.**

The installer will:
1. Detect your architecture (AMD64 or ARM64)
2. Download the right binary from the latest GitHub release
3. Place it in `%USERPROFILE%\.pathman\bin\`
4. Add it to your User PATH automatically

Restart your terminal, then:

```powershell
pathman doctor   # scan your PATH for issues
```

---

## Build from source

Requires **Windows** and **Go 1.26 or newer**. No other tooling is needed.

```powershell
git clone https://github.com/yv3000/pathman_doctor
cd pathman_doctor

go build -o pathman.exe .
go vet ./...
go test ./... -cover

.\pathman.exe doctor
```

`go test ./...` runs the whole suite against a fake registry store and a fake
registry key, so it never reads or writes your live PATH. The coverage gate runs
the same tests and fails below 80% per package (`cmd`, `internal`) and overall:

```powershell
.\scripts\coverage-gate.ps1
```

The live-registry smoke test is opt-in: set `PATHMAN_LIVE_REGISTRY=1` before
`go test` if you want to read your own User PATH. The repository does not
contain a prebuilt binary; release executables are produced by CI for every
`v*` tag.

---

## Uninstall

```powershell
pathman uninstall
```

The command removes pathman's own User PATH entry (after saving a backup) and
deletes `%USERPROFILE%\.pathman\bin`. It needs no network access and never
touches your other PATH entries. Windows locks the running executable, so the
folder itself is deleted a moment after the process exits.

If the binary is already gone, run the script instead:

```powershell
irm https://raw.githubusercontent.com/yv3000/pathman_doctor/main/uninstall.ps1 | iex
```

---

## Proof it works

```powershell
pathman --version              # pathman v1.2.0

pathman doctor                 # scan and show all issues

pathman fix                    # fix everything (dead + duplicates)
pathman fix --only dead        # fix only dead entries
pathman fix --only dup         # fix only duplicates
pathman fix --entry "C:\some\old\tool\bin"   # fix one specific path
```

---

## Commands

| Command | What it does |
|---|---|
| `pathman doctor` | Scan all PATH entries and display their status |
| `pathman fix` | Remove all dead and duplicate entries |
| `pathman fix --only dead` | Remove only entries pointing to missing folders |
| `pathman fix --only dup` | Remove only duplicate entries (keeps first occurrence) |
| `pathman fix --entry "C:\path"` | Fix all issues for one specific path only |
| `pathman uninstall` | Remove pathman's files and its User PATH entry |
| `pathman --version` | Show installed version |
| `pathman --help` | Show usage |

Every command validates its flags. An unknown flag, a missing `--entry` value
or an invalid `--only` value prints the usage and changes nothing.

### Exit codes

| Code | Meaning |
|---|---|
| `0` | All requested work succeeded |
| `1` | Nothing changed, or only part of the work could be applied |
| `2` | Invalid command or flags — PATH was not touched |

`pathman fix` exits with `1` when, for example, it cleaned the User PATH but
lacked Administrator rights for the System PATH.

---

## How It Works

### What pathman scans

pathman reads both PATH scopes directly from the Windows Registry:

```
HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Control\Session Manager\Environment   ← System PATH
HKEY_CURRENT_USER\Environment                                                      ← User PATH
```

Each entry gets one of three statuses:

| Status | Meaning |
|---|---|
| `✅ OK` | Folder exists on disk |
| `❌ DEAD` | Folder is missing — safe to remove |
| `⚠️ DUP` | Same path already appeared earlier — safe to remove |

`%VAR%` entries are resolved with your environment before they are checked, so
`%SystemRoot%\system32` is verified as the real folder. A variable that is not
set is never treated as safe to delete.

### The Fix Flow

```
pathman fix
     |
     v
Read PATH from Registry (System + User)
     |
     v
Analyze: mark each entry as OK / DEAD / DUP
     |
     v
Show issues + ask for confirmation (y/n)
     |
     v
Back up the untouched PATH value
     |
     v
Write cleaned PATH back to Registry
     |
     v
Broadcast WM_SETTINGCHANGE to all open windows
     |
     v
✅ Done. No restart required.
```

### No Restart Needed

After fixing, pathman calls `SendMessageTimeout` with `WM_SETTINGCHANGE` — the same signal Windows itself uses when you change environment variables via System Properties. All open applications pick up the change immediately.

---

## Admin Rights

| Scope | Admin required? |
|---|---|
| User PATH | No — works as a normal user |
| System PATH | Yes — run PowerShell as Administrator |

If you run `pathman fix` without admin and there are System PATH issues, pathman fixes User PATH, warns you about the System entries, and exits with status `1` so the incomplete run is visible to scripts. To fix System PATH:

```powershell
# Open PowerShell as Administrator, then:
pathman fix
# or for a specific entry:
pathman fix --entry "C:\some\broken\path"
```

---

## What `--entry` Does

`pathman fix --entry "C:\some\path"` targets **all issues for that one path** — removes every occurrence that is DEAD or DUP, no matter how many times it appears. Other entries are never touched.

---

## Requirements

| Requirement | Details |
|---|---|
| **OS** | Windows 10 / Windows 11 |
| **Architecture** | AMD64 or ARM64 |
| **Admin** | Only for System PATH fixes (optional) |
| **Go** | 1.26+ if you build from source |
| **Dependencies** | None at runtime — single static binary |

---

## FAQ

**Will this delete folders from my disk?**
No. pathman only modifies PATH entries in the Registry. It never touches files or folders on disk.

**Can I preview what will be fixed before confirming?**
Yes — `pathman fix` always shows every entry it will remove and asks `Apply these changes? (y/n)` before doing anything.

**Does pathman keep a copy of my old PATH?**
Yes. Before every write, `fix` and `uninstall` save the untouched registry value to `%LOCALAPPDATA%\pathman\backups\` and print the file path. To restore, copy that file's content back into PATH from System Properties. If the backup cannot be written, nothing is changed.

**What if I accidentally remove something I needed?**
Restore the backup from the path printed by the command, or re-add the entry manually:
```powershell
$cur = [Environment]::GetEnvironmentVariable("PATH", "User")
[Environment]::SetEnvironmentVariable("PATH", "$cur;C:\the\path\you\want", "User")
```

**Why doesn't it fix System PATH without admin?**
Windows restricts writes to `HKEY_LOCAL_MACHINE` to Administrator accounts by design. This is a Windows security boundary, not a pathman limitation.

**Does pathman work inside a venv or conda environment?**
Yes. pathman operates on the system/user PATH in the Registry — it's fully independent of any Python virtual environment.

**Does it support ARM64?**
Yes. The installer detects your architecture and the release workflow publishes both AMD64 and ARM64 binaries.

---

## Tech Stack

- Go — single static binary, zero runtime dependencies
- Windows Registry API via `golang.org/x/sys/windows/registry`
- `SendMessageTimeout` (Win32) for no-restart PATH broadcasting
- PowerShell installer — no external dependencies
- Tests and lint run in CI on every push and pull request

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for the build, test and lint commands and
the rule that every behaviour change ships with tests. Bug reports and pull
requests are welcome.

---

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for the full release history.

---

## License <a name="license"></a>

MIT License — see [LICENSE](LICENSE).

Copyright (c) 2026 YV

---

<p align="center">
  <sub>YV 🖤 ~ I EXPECT NOTHING FROM YOU...</sub>
</p>
