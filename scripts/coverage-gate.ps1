#requires -Version 5.1
<#
.SYNOPSIS
Runs the test suite and fails when coverage regresses.

.DESCRIPTION
`go test` reports coverage but does not enforce a floor, so a change could drop
the suite from 89% to 40% and still pass CI. This gate runs the same tests with a
coverage profile and fails when a required package or the whole module falls
below the configured minimum.

.PARAMETER MinimumPackage
Per-package floor for each package listed in -RequiredPackages. Defaults to 80.

.PARAMETER MinimumTotal
Floor for total statement coverage across the module. Defaults to 80.

.EXAMPLE
./scripts/coverage-gate.ps1

.EXAMPLE
./scripts/coverage-gate.ps1 -MinimumPackage 85 -MinimumTotal 85
#>
[CmdletBinding()]
param(
    [double]$MinimumPackage = 80,
    [double]$MinimumTotal = 80,
    [string[]]$RequiredPackages = @('cmd', 'internal')
)

$ErrorActionPreference = 'Stop'

$root = Split-Path -Parent $PSScriptRoot
$coverageFile = Join-Path $root 'coverage.out'

Push-Location $root
try {
    $output = & go test ./... -count=1 -coverprofile $coverageFile -covermode atomic 2>&1
    $testExit = $LASTEXITCODE
    foreach ($line in $output) {
        Write-Host $line
    }
    if ($testExit -ne 0) {
        Write-Host "coverage gate: tests failed with exit code $testExit"
        exit $testExit
    }

    $coverage = @{}
    foreach ($line in $output) {
        if ("$line" -match '^ok\s+(\S+)\s+.*coverage:\s+([0-9.]+)% of statements') {
            $package = $Matches[1]
            $short = $package.Substring($package.LastIndexOf('/') + 1)
            $coverage[$short] = [double]$Matches[2]
        }
    }

    $failures = New-Object System.Collections.Generic.List[string]

    foreach ($name in $RequiredPackages) {
        if (-not $coverage.ContainsKey($name)) {
            $failures.Add("package '$name' reported no coverage")
            continue
        }
        $value = $coverage[$name]
        if ($value -lt $MinimumPackage) {
            $failures.Add(('{0}: {1:N1}% is below the {2:N1}% floor' -f $name, $value, $MinimumPackage))
        } else {
            Write-Host ('coverage gate: {0} {1:N1}% >= {2:N1}%' -f $name, $value, $MinimumPackage)
        }
    }

    $totalLine = & go tool cover -func $coverageFile | Select-String -Pattern '^total:'
    $total = $null
    if ($null -ne $totalLine -and "$totalLine" -match '([0-9.]+)%\s*$') {
        $total = [double]$Matches[1]
    }

    if ($null -eq $total) {
        $failures.Add('could not read total coverage from the profile')
    } elseif ($total -lt $MinimumTotal) {
        $failures.Add(('total: {0:N1}% is below the {1:N1}% floor' -f $total, $MinimumTotal))
    } else {
        Write-Host ('coverage gate: total {0:N1}% >= {1:N1}%' -f $total, $MinimumTotal)
    }

    if ($failures.Count -gt 0) {
        foreach ($failure in $failures) {
            Write-Host "coverage gate failed: $failure"
        }
        exit 1
    }

    Write-Host 'coverage gate passed.'
    exit 0
}
finally {
    Pop-Location
}
