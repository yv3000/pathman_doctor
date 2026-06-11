# pathman

A fast, no-nonsense Windows CLI tool to diagnose and auto-fix your PATH environment variable.

Over time, the Windows PATH gets cluttered with dead folders (uninstalled software) and duplicate entries. `pathman` reads directly from the registry to show you what's broken and safely cleans it up with a single command, without requiring a system restart.

[![Built with Go](https://img.shields.io/badge/Built_with-Go-00ADD8?style=flat-square&logo=go)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg?style=flat-square)](https://opensource.org/licenses/MIT)

## Installation

1. Download `pathman.exe` from the [Releases](../../releases) page.
2. Drop it into any folder that's already in your PATH (e.g., `C:\Windows\System32` or your preferred bin folder).
3. Open a new terminal and type `pathman`.

## Usage

### Scan for Issues
Run `pathman doctor` to see all entries in your System and User PATH variables. It will highlight dead directories and duplicates.

```
> pathman doctor
pathman doctor — scanning PATH entries...

SCOPE    STATUS   PATH
------   ------   ----
System   ✅ OK    C:\Windows\System32
System   ✅ OK    C:\Windows
User     ✅ OK    C:\Program Files\nodejs
User     ❌ DEAD  C:\Users\YV\AppData\Local\Programs\oldpython   (folder missing)
User     ⚠️  DUP   C:\Program Files\Git\cmd   (duplicate — appears 2x)
User     ✅ OK    C:\Users\YV\.bun\bin

Summary: 6 entries | 1 dead | 1 duplicate | action needed
Run `pathman fix` to auto-clean.
```

### Auto-Fix
Run `pathman fix` to safely remove dead and duplicate entries. It will ask for confirmation before applying any changes. Run your terminal as Administrator to allow fixing System PATH entries.

```
> pathman fix
pathman fix — found 2 issues

  ❌ Removing DEAD:  C:\Users\YV\AppData\Local\Programs\oldpython   (folder missing)
  ⚠️  Removing DUP:   C:\Program Files\Git\cmd   (keeping first occurrence)

Apply these changes? (y/n): y

✅ Done. PATH cleaned. 2 entries removed.
No restart required.
```

## Author
[YV](https://github.com/yv3000)

## License
MIT License
