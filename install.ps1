#Requires -Version 5.1
<#
.SYNOPSIS
    GitZen Installer for Windows

.DESCRIPTION
    Auto-detect architecture, download from GitHub Releases, verify checksum, and install GitZen.

.PARAMETER Version
    Specific version to install (e.g., v0.1.0). If not specified, installs latest.

.PARAMETER InstallDir
    Installation directory. Defaults to $env:LOCALAPPDATA\gitzen

.PARAMETER Uninstall
    Uninstall GitZen

.EXAMPLE
    # Install latest version
    irm https://quanghai2k4.github.io/gitzen/install.ps1 | iex

    # Install specific version
    & { $v="v0.1.0"; irm https://quanghai2k4.github.io/gitzen/install.ps1 | iex }

    # Uninstall
    gitzen --uninstall
#>

[CmdletBinding()]
param(
    [string]$Version,
    [string]$InstallDir,
    [switch]$Uninstall
)

$ErrorActionPreference = "Stop"

# Configuration
$Repo = "quanghai2k4/gitzen"
$BinaryName = "gitzen.exe"
$GitHubApi = "https://api.github.com/repos/$Repo/releases"
$GitHubDownload = "https://github.com/$Repo/releases/download"

# Colors
function Write-Info { Write-Host "[INFO] $args" -ForegroundColor Cyan }
function Write-Success { Write-Host "[OK] $args" -ForegroundColor Green }
function Write-Warn { Write-Host "[WARN] $args" -ForegroundColor Yellow }
function Write-Err { Write-Host "[ERROR] $args" -ForegroundColor Red }

# Get latest version from GitHub
function Get-LatestVersion {
    try {
        $release = Invoke-RestMethod -Uri "$GitHubApi/latest" -UseBasicParsing
        return $release.tag_name
    }
    catch {
        throw "Failed to get latest version from GitHub: $_"
    }
}

# Detect architecture
function Get-Arch {
    $arch = [System.Environment]::GetEnvironmentVariable("PROCESSOR_ARCHITECTURE")
    switch ($arch) {
        "AMD64" { return "amd64" }
        "x86" { throw "32-bit Windows is not supported" }
        "ARM64" { throw "ARM64 Windows is not supported yet" }
        default { throw "Unsupported architecture: $arch" }
    }
}

# Get default install directory
function Get-DefaultInstallDir {
    if ($InstallDir) { return $InstallDir }
    return Join-Path $env:LOCALAPPDATA "gitzen"
}

# Verify checksum
function Test-Checksum {
    param(
        [string]$FilePath,
        [string]$ChecksumsFile
    )
    
    $fileName = Split-Path $FilePath -Leaf
    $checksums = Get-Content $ChecksumsFile
    
    $expectedLine = $checksums | Where-Object { $_ -match $fileName }
    if (-not $expectedLine) {
        Write-Warn "Checksum not found for $fileName, skipping verification"
        return $true
    }
    
    $expectedHash = ($expectedLine -split '\s+')[0]
    $actualHash = (Get-FileHash -Path $FilePath -Algorithm SHA256).Hash.ToLower()
    
    if ($expectedHash -ne $actualHash) {
        throw "Checksum verification failed!`nExpected: $expectedHash`nActual: $actualHash"
    }
    
    Write-Success "Checksum verified"
    return $true
}

# Add to PATH
function Add-ToPath {
    param([string]$Dir)
    
    $currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($currentPath -notlike "*$Dir*") {
        $newPath = "$currentPath;$Dir"
        [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
        $env:Path = "$env:Path;$Dir"
        Write-Success "Added $Dir to PATH"
    }
}

# Remove from PATH
function Remove-FromPath {
    param([string]$Dir)
    
    $currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
    $newPath = ($currentPath -split ';' | Where-Object { $_ -ne $Dir }) -join ';'
    [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
}

# Uninstall function
function Invoke-Uninstall {
    Write-Info "Uninstalling GitZen..."
    
    $locations = @(
        (Join-Path $env:LOCALAPPDATA "gitzen"),
        (Join-Path $env:ProgramFiles "gitzen"),
        (Join-Path ${env:ProgramFiles(x86)} "gitzen")
    )
    
    $found = $false
    foreach ($loc in $locations) {
        $binaryPath = Join-Path $loc $BinaryName
        if (Test-Path $binaryPath) {
            Remove-Item -Path $loc -Recurse -Force
            Remove-FromPath $loc
            Write-Success "Removed $loc"
            $found = $true
        }
    }
    
    if (-not $found) {
        Write-Warn "GitZen is not installed"
    }
    else {
        Write-Success "GitZen has been uninstalled successfully!"
    }
}

# Main install function
function Install-Gitzen {
    Write-Host ""
    Write-Host "╔═══════════════════════════════════════╗" -ForegroundColor Green
    Write-Host "║       GitZen Installer (Windows)      ║" -ForegroundColor Green
    Write-Host "╚═══════════════════════════════════════╝" -ForegroundColor Green
    Write-Host ""
    
    # Detect architecture
    $arch = Get-Arch
    Write-Info "Detected architecture: $arch"
    
    # Get version
    $ver = $Version
    if (-not $ver) {
        Write-Info "Fetching latest version..."
        $ver = Get-LatestVersion
    }
    Write-Info "Version: $ver"
    
    # Prepare download
    $versionNum = $ver -replace '^v', ''
    $archiveName = "gitzen_${versionNum}_windows_${arch}.zip"
    $downloadUrl = "$GitHubDownload/$ver/$archiveName"
    $checksumsUrl = "$GitHubDownload/$ver/checksums.txt"
    
    # Create temp directory
    $tempDir = Join-Path $env:TEMP "gitzen-install-$(Get-Random)"
    New-Item -ItemType Directory -Path $tempDir -Force | Out-Null
    
    try {
        # Download archive
        Write-Info "Downloading $archiveName..."
        $archivePath = Join-Path $tempDir $archiveName
        Invoke-WebRequest -Uri $downloadUrl -OutFile $archivePath -UseBasicParsing
        Write-Success "Downloaded $archiveName"
        
        # Download checksums
        Write-Info "Downloading checksums..."
        $checksumsPath = Join-Path $tempDir "checksums.txt"
        try {
            Invoke-WebRequest -Uri $checksumsUrl -OutFile $checksumsPath -UseBasicParsing
            Test-Checksum -FilePath $archivePath -ChecksumsFile $checksumsPath
        }
        catch {
            Write-Warn "Failed to download or verify checksums"
        }
        
        # Extract archive
        Write-Info "Extracting..."
        $extractDir = Join-Path $tempDir "extracted"
        Expand-Archive -Path $archivePath -DestinationPath $extractDir -Force
        
        # Get install directory
        $installDir = Get-DefaultInstallDir
        
        # Create install directory
        if (-not (Test-Path $installDir)) {
            New-Item -ItemType Directory -Path $installDir -Force | Out-Null
        }
        
        # Install binary
        Write-Info "Installing to $installDir..."
        $sourceBinary = Get-ChildItem -Path $extractDir -Filter "*.exe" -Recurse | Select-Object -First 1
        if (-not $sourceBinary) {
            $sourceBinary = Join-Path $extractDir $BinaryName
        }
        Copy-Item -Path $sourceBinary.FullName -Destination (Join-Path $installDir $BinaryName) -Force
        
        Write-Success "Installed GitZen to $installDir"
        
        # Add to PATH
        Add-ToPath $installDir
        
        Write-Host ""
        Write-Success "Installation complete!"
        Write-Host ""
        Write-Host "Run 'gitzen' to start using GitZen" -ForegroundColor Cyan
        Write-Host "You may need to restart your terminal for PATH changes to take effect." -ForegroundColor Yellow
        Write-Host ""
    }
    finally {
        # Cleanup
        if (Test-Path $tempDir) {
            Remove-Item -Path $tempDir -Recurse -Force -ErrorAction SilentlyContinue
        }
    }
}

# Main entry point
if ($Uninstall) {
    Invoke-Uninstall
}
else {
    Install-Gitzen
}
