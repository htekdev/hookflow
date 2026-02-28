# hookflow install script for Windows
# Usage: iwr -useb https://raw.githubusercontent.com/htekdev/hookflow/main/scripts/install.ps1 | iex
# Or: Invoke-WebRequest -UseBasicParsing https://raw.githubusercontent.com/htekdev/hookflow/main/scripts/install.ps1 | Invoke-Expression

param(
    [string]$Version = "latest",
    [string]$InstallDir = ""
)

$ErrorActionPreference = "Stop"

$Repo = "htekdev/hookflow"

function Write-Info {
    param([string]$Message)
    Write-Host "[INFO] " -ForegroundColor Green -NoNewline
    Write-Host $Message
}

function Write-Warn {
    param([string]$Message)
    Write-Host "[WARN] " -ForegroundColor Yellow -NoNewline
    Write-Host $Message
}

function Write-Err {
    param([string]$Message)
    Write-Host "[ERROR] " -ForegroundColor Red -NoNewline
    Write-Host $Message
    exit 1
}

function Get-Architecture {
    if ([System.Runtime.InteropServices.RuntimeInformation]::ProcessArchitecture -eq "Arm64") {
        return "arm64"
    }
    return "amd64"
}

function Get-LatestVersion {
    try {
        $response = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest" -UseBasicParsing
        return $response.tag_name
    } catch {
        return $null
    }
}

function Install-AgenticOps {
    Write-Info "Installing hookflow CLI..."

    $Arch = Get-Architecture
    Write-Info "Detected: windows-$Arch"

    # Get version
    if ($Version -eq "latest") {
        $Version = Get-LatestVersion
        if (-not $Version) {
            Write-Warn "Could not fetch latest version, trying direct download..."
            $Version = "latest"
        }
    }

    # Build binary name and download URL
    $BinaryName = "hookflow-windows-$Arch.exe"
    
    if ($Version -eq "latest") {
        $DownloadUrl = "https://github.com/$Repo/releases/latest/download/$BinaryName"
    } else {
        $DownloadUrl = "https://github.com/$Repo/releases/download/$Version/$BinaryName"
    }

    Write-Info "Downloading from: $DownloadUrl"

    # Determine install location
    if (-not $InstallDir) {
        $InstallDir = Join-Path $env:LOCALAPPDATA "Programs\hookflow"
    }

    # Create install directory
    if (-not (Test-Path $InstallDir)) {
        New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    }

    $TargetPath = Join-Path $InstallDir "hookflow.exe"

    # Download
    try {
        Invoke-WebRequest -Uri $DownloadUrl -OutFile $TargetPath -UseBasicParsing
    } catch {
        Write-Err "Download failed: $_"
    }

    # Verify download
    if (-not (Test-Path $TargetPath) -or (Get-Item $TargetPath).Length -eq 0) {
        Write-Err "Downloaded file is empty. The release may not exist yet."
    }

    Write-Info "Installed hookflow to $TargetPath"

    # Add to PATH if not already there
    $UserPath = [Environment]::GetEnvironmentVariable("PATH", "User")
    if ($UserPath -notlike "*$InstallDir*") {
        Write-Info "Adding $InstallDir to PATH..."
        $NewPath = "$UserPath;$InstallDir"
        [Environment]::SetEnvironmentVariable("PATH", $NewPath, "User")
        $env:PATH = "$env:PATH;$InstallDir"
        Write-Info "Added to PATH. Restart your terminal to use 'hookflow' command."
    }

    # Verify installation
    try {
        $VersionOutput = & $TargetPath version 2>&1
        Write-Info $VersionOutput
    } catch {
        Write-Warn "Could not verify installation: $_"
    }

    Write-Info ""
    Write-Info "Get started:"
    Write-Info "  cd your-project"
    Write-Info "  hookflow init"
    Write-Info "  hookflow create `"block edits to .env files`""
}

Install-AgenticOps
