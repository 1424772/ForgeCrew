# ForgeCrew installer for Windows PowerShell
# Downloads the latest forgecrew.exe from GitHub Releases.
# Installs to user-local directory, no admin required.

param(
    [string]$Version = "latest",
    [string]$InstallDir = ""
)

$Repo = "1424772/ForgeCrew"
$Binary = "forgecrew.exe"

if ($InstallDir -eq "") {
    $InstallDir = "$env:USERPROFILE\.forgecrew\bin"
}

# Detect architecture using RuntimeInformation.OSArchitecture.
$Arch = switch ([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture) {
    "X64"   { "amd64" }
    "Arm64" { "arm64" }
    default {
        Write-Host "[forgecrew] Unsupported architecture: $([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture)" -ForegroundColor Red
        Write-Host "[forgecrew] This installer supports x64 and arm64 only." -ForegroundColor Red
        exit 1
    }
}
$OS = "windows"

Write-Host "[forgecrew] Installing ForgeCrew $Version for $OS/$Arch..." -ForegroundColor Green
Write-Host "[forgecrew] Install directory: $InstallDir" -ForegroundColor Green

# Build download URL (GitHub Releases placeholder until releases are tagged).
$ReleaseUrl = if ($Version -eq "latest") {
    "https://github.com/${Repo}/releases/latest/download/forgecrew_${OS}_${Arch}.exe"
} else {
    "https://github.com/${Repo}/releases/download/${Version}/forgecrew_${OS}_${Arch}.exe"
}

# Create install directory.
if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
}

$DestFile = Join-Path $InstallDir $Binary
$TmpFile = "$DestFile.tmp.$PID"

# Clean up any leftover temp file from a previous failed run.
Remove-Item -Force $TmpFile -ErrorAction SilentlyContinue | Out-Null

# Download to temp file first, validate, then move to final path.
$DownloadOk = $false
try {
    Write-Host "[forgecrew] Downloading $ReleaseUrl ..." -ForegroundColor Green
    $ProgressPreference = 'SilentlyContinue'
    Invoke-WebRequest -Uri $ReleaseUrl -OutFile $TmpFile -ErrorAction Stop
    $DownloadOk = $true
} catch {
    Write-Host "[forgecrew] Download failed: $_" -ForegroundColor Yellow
}

# Validate the downloaded temp file before moving it.
if ($DownloadOk) {
    if (-not (Test-Path $TmpFile)) {
        Write-Host "[forgecrew] Downloaded temp file not found." -ForegroundColor Red
        $DownloadOk = $false
    } elseif ((Get-Item $TmpFile).Length -eq 0) {
        Write-Host "[forgecrew] Downloaded file is empty." -ForegroundColor Red
        Remove-Item -Force $TmpFile -ErrorAction SilentlyContinue | Out-Null
        $DownloadOk = $false
    }
}

if (-not $DownloadOk) {
    Remove-Item -Force $TmpFile -ErrorAction SilentlyContinue | Out-Null
    Write-Host ""
    Write-Host "[forgecrew] Release not found at $ReleaseUrl" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "[forgecrew] ForgeCrew binaries are not yet published to GitHub Releases." -ForegroundColor Yellow
    Write-Host "[forgecrew] You can build from source:" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "  git clone https://github.com/${Repo}.git" -ForegroundColor Yellow
    Write-Host "  cd ForgeCrew" -ForegroundColor Yellow
    Write-Host "  go build -o ${Binary} .\cmd\forgecrew" -ForegroundColor Yellow
    Write-Host "  move ${Binary} ${InstallDir}\" -ForegroundColor Yellow
    Write-Host ""

    $GoCmd = Get-Command go -ErrorAction SilentlyContinue
    if ($GoCmd) {
        Write-Host "[forgecrew] Go detected. You can also run:" -ForegroundColor Green
        Write-Host "  go install github.com/${Repo}/cmd/forgecrew@latest" -ForegroundColor Green
    }
    exit 1
}

# Download validated — move temp file to final path.
Move-Item -Force $TmpFile $DestFile

Write-Host "[forgecrew] ForgeCrew installed to $DestFile" -ForegroundColor Green
Write-Host ""

# Check if install dir is in PATH.
$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($currentPath -notlike "*$InstallDir*") {
    Write-Host "[forgecrew] NOTE: $InstallDir is not in your PATH." -ForegroundColor Yellow
    Write-Host "[forgecrew] To add it, run the following in PowerShell:" -ForegroundColor Yellow
    Write-Host ""
    Write-Host '  [Environment]::SetEnvironmentVariable("Path", $env:Path + ";' + "$InstallDir" + '", "User")' -ForegroundColor Yellow
    Write-Host ""
    Write-Host "[forgecrew] Then restart your terminal, or run for this session:" -ForegroundColor Yellow
    Write-Host '  $env:Path += ";' + "$InstallDir" + '"' -ForegroundColor Yellow
}

Write-Host "[forgecrew] Run 'forgecrew init' in your project to get started." -ForegroundColor Green
