try {
    $installDir = "$env:USERPROFILE\.pathman\bin"
    $exePath = "$installDir\pathman.exe"

    if (-not (Test-Path $exePath)) {
        Write-Host "pathman does not appear to be installed."
        exit
    }

    Remove-Item -Path $exePath -Force

    if ((Get-ChildItem -Path $installDir | Measure-Object).Count -eq 0) {
        Remove-Item -Path $installDir -Recurse -Force
    }

    try {
        $currentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
        if ($currentPath) {
            $pathArray = $currentPath -split ';'
            $newPathArray = @()
            foreach ($p in $pathArray) {
                if ($p.TrimEnd('\') -ne $installDir.TrimEnd('\') -and $p.Trim() -ne "") {
                    $newPathArray += $p
                }
            }
            $newPath = $newPathArray -join ';'
            if ($newPath -ne $currentPath) {
                [Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
            }
        }
    } catch {
        Write-Host "Warning: Could not remove $installDir from PATH automatically."
    }

    try {
        $code = @'
using System;
using System.Runtime.InteropServices;
public class PathBroadcast {
    [DllImport("user32.dll", SetLastError=true, CharSet=CharSet.Auto)]
    public static extern IntPtr SendMessageTimeout(
        IntPtr hWnd, uint Msg, UIntPtr wParam,
        string lParam, uint fuFlags, uint uTimeout, out UIntPtr lpdwResult);
}
'@
        Add-Type -TypeDefinition $code -ErrorAction SilentlyContinue
        $result = [UIntPtr]::Zero
        [PathBroadcast]::SendMessageTimeout([IntPtr]0xffff, 0x001A, [UIntPtr]::Zero, "Environment", 2, 5000, [ref]$result) | Out-Null
    } catch {
        # ignore broadcast error
    }

    Write-Host "✅ pathman uninstalled successfully."
    Write-Host "Restart your terminal to complete removal."

} catch {
    Write-Host "An unexpected error occurred during uninstallation: $_"
}
