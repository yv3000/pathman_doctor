try {
    $installDir = "$env:USERPROFILE\.pathman\bin"
    $exePath = "$installDir\pathman.exe"

    if (Test-Path $exePath) {
        Write-Host "pathman is already installed. Run 'pathman --version' to check."
        exit
    }

    $arch = $env:PROCESSOR_ARCHITECTURE
    $assetName = ""
    if ($arch -eq "AMD64") {
        $assetName = "pathman-windows-amd64.exe"
    } elseif ($arch -eq "ARM64") {
        $assetName = "pathman-windows-arm64.exe"
    } else {
        Write-Host "Unsupported architecture: $arch"
        exit
    }

    try {
        $releaseUrl = "https://api.github.com/repos/yv3000/pathman_doctor/releases/latest"
        $release = Invoke-RestMethod -Uri $releaseUrl
        $version = $release.tag_name
        
        $downloadUrl = ""
        foreach ($asset in $release.assets) {
            if ($asset.name -eq $assetName) {
                $downloadUrl = $asset.browser_download_url
                break
            }
        }
        
        if ($downloadUrl -eq "") {
            Write-Host "Failed to find release asset for $arch"
            exit
        }
    } catch {
        Write-Host "Failed to fetch latest release. Check your internet connection."
        exit
    }

    if (!(Test-Path $installDir)) {
        New-Item -ItemType Directory -Force -Path $installDir | Out-Null
    }

    Write-Host "Downloading pathman $version..."
    try {
        Invoke-WebRequest -Uri $downloadUrl -OutFile $exePath
    } catch {
        Write-Host "Download failed. Please try again."
        exit
    }

    try {
        $currentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
        if ($currentPath -eq $null) {
            $currentPath = ""
        }
        
        $pathArray = $currentPath -split ';'
        $alreadyInPath = $false
        foreach ($p in $pathArray) {
            if ($p.TrimEnd('\') -eq $installDir.TrimEnd('\')) {
                $alreadyInPath = $true
                break
            }
        }

        if (-not $alreadyInPath) {
            $newPath = $currentPath
            if ($newPath -and -not $newPath.EndsWith(";")) {
                $newPath += ";"
            }
            $newPath += $installDir
            [Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
            Write-Host "Added to PATH: $installDir"
        }
    } catch {
        Write-Host "Could not update PATH automatically. Manually add $installDir to your PATH."
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

    Write-Host ""
    Write-Host "✅ pathman $version installed successfully!"
    Write-Host ""
    Write-Host "Restart your terminal, then run:"
    Write-Host "  pathman doctor"

} catch {
    Write-Host "An unexpected error occurred during installation: $_"
}
