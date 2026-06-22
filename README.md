<p align="center">
  <a href="https://golang.org"><img src="https://img.shields.io/badge/built%20with-Go-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go" /></a>
  <a href="https://www.microsoft.com/windows"><img src="https://img.shields.io/badge/platform-Windows-0078D6?style=for-the-badge&logo=windows&logoColor=white" alt="Windows" /></a>
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

## Uninstall

```powershell
irm https://raw.githubusercontent.com/yv3000/pathman_doctor/main/uninstall.ps1 | iex
```

This removes the `~\.pathman\` folder and cleans the entry from your User PATH — completely. Your original PATH entries are never modified by the uninstaller.

---

## Proof it works

```powershell
pathman --version              # pathman v1.1.0

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
| `pathman --version` | Show installed version |
| `pathman --help` | Show usage |

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

If you run `pathman fix` without admin and there are System PATH issues, pathman fixes User PATH and warns you about the System entries. To fix System PATH:

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
| **Dependencies** | None — single static binary |

---

## FAQ

**Will this delete folders from my disk?**
No. pathman only modifies PATH entries in the Registry. It never touches files or folders on disk.

**Can I preview what will be fixed before confirming?**
Yes — `pathman fix` always shows every entry it will remove and asks `Apply these changes? (y/n)` before doing anything.

**What if I accidentally remove something I needed?**
Re-add the path manually:
```powershell
$cur = [Environment]::GetEnvironmentVariable("PATH", "User")
[Environment]::SetEnvironmentVariable("PATH", "$cur;C:\the\path\you\want", "User")
```

**Why doesn't it fix System PATH without admin?**
Windows restricts writes to `HKEY_LOCAL_MACHINE` to Administrator accounts by design. This is a Windows security boundary, not a pathman limitation.

**Does pathman work inside a venv or conda environment?**
Yes. pathman operates on the system/user PATH in the Registry — it's fully independent of any Python virtual environment.

---

## Tech Stack

- Go — single static binary, zero runtime dependencies
- Windows Registry API via `golang.org/x/sys/windows/registry`
- `SendMessageTimeout` (Win32) for no-restart PATH broadcasting
- PowerShell installer / uninstaller — no external dependencies

---

## Changelog

- **v1.1.0** — Added `--only dead`, `--only dup`, and `--entry` flags; fixed PowerShell window closing on install/uninstall
- **v1.0.0** — Initial release: `doctor`, `fix`, install/uninstall scripts, no-restart broadcasting

---

## License <a name="license"></a>

MIT License

Copyright (c) 2026 YV

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

---

<p align="center">
  <sub>YV 🖤 ~ I EXPECT NOTHING FROM YOU...</sub>
</p>
