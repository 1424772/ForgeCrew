# ForgeCrew installer for Windows PowerShell
# Downloads the latest forgecrew.exe from GitHub Releases.
# Falls back to go build if no release is found.

param(
    [string]$Version = "latest",
    [string]$InstallDir = ""
)

$Repo = "1424772/ForgeCrew"
$Binary = "forgecrew.exe"

if ($InstallDir -eq "") {
    $InstallDir = "$env:LOCALAPPDATA\forgecrew"
}

# Detect architecture.
$Arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }
$OS = "windows"

Write-Host "[forgecrew] Installing ForgeCrew $Version for $OS/$Arch..." -ForegroundColor Green

# Build download URL.
$ReleaseUrl = if ($Version -eq "latest") {
    "https://github.com/${Repo}/releases/latest/download/${Binary}-${OS}-${Arch}.exe"
} else {
    "https://github.com/${Repo}/releases/download/${Version}/${Binary}-${OS}-${Arch}.exe"
}

# Create install directory.
if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
}

$TempFile = Join-Path $env:TEMP $Binary

# Try to download from GitHub Releases.
try {
    Write-Host "[forgecrew] Downloading $ReleaseUrl ..." -ForegroundColor Green
    $ProgressPreference = 'SilentlyContinue'
    Invoke-WebRequest -Uri $ReleaseUrl -OutFile $TempFile -ErrorAction Stop
    $DownloadOk = $true
} catch {
    Write-Host "[forgecrew] Release download failed: $_" -ForegroundColor Yellow
    $DownloadOk = $false
}

if (-not $DownloadOk -or -not (Test-Path $TempFile)) {
    Write-Host ""
    Write-Host "[forgecrew] Release not found. Building from source or use manual install." -ForegroundColor Yellow
    Write-Host ""

    $GoCmd = Get-Command go -ErrorAction SilentlyContinue
    if ($GoCmd) {
        Write-Host "[forgecrew] Building from source with 'go install'..." -ForegroundColor Green
        go install "github.com/${Repo}/cmd/forgecrew@latest"
        Write-Host "[forgecrew] Installed via go install." -ForegroundColor Green
        exit 0
    }

    Write-Host "Option 1: Install Go (https://go.dev/dl/) and run:" -ForegroundColor Yellow
    Write-Host "  go install github.com/${Repo}/cmd/forgecrew@latest"
    Write-Host ""
    Write-Host "Option 2: Clone and build manually:" -ForegroundColor Yellow
    Write-Host "  git clone https://github.com/${Repo}.git"
    Write-Host "  cd ForgeCrew"
    Write-Host "  go build -o forgecrew.exe .\cmd\forgecrew"
    Write-Host "  move forgecrew.exe $InstallDir\"
    exit 1
}

# Install the binary.
$DestFile = Join-Path $InstallDir $Binary
Move-Item -Force -Path $TempFile -Destination $DestFile

# Add to PATH if not already there.
$UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($UserPath -notlike "*$InstallDir*") {
    Write-Host "[forgecrew] Adding $InstallDir to user PATH..." -ForegroundColor Green
    [Environment]::SetEnvironmentVariable("Path", "$UserPath;$InstallDir", "User")
    $env:Path = "$env:Path;$InstallDir"
}

Write-Host "[forgecrew] ForgeCrew installed successfully!" -ForegroundColor Green
Write-Host "[forgecrew] Run 'forgecrew init' in your project to get started." -ForegroundColor Green
Write-Host "[forgecrew] If the command is not found, restart your terminal or run:" -ForegroundColor Green
Write-Host "  `$env:Path += ';$InstallDir'" -ForegroundColor Green
