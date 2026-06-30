# ForgeCrew installer for Windows PowerShell
# Compatible with PowerShell 5.1 and PowerShell 7.
# Downloads the latest forgecrew.exe from GitHub Releases,
# verifies SHA256 checksum, and adds to user PATH.

param(
    [string]$Version = "latest",
    [string]$InstallDir = ""
)

$ErrorActionPreference = "Stop"

$Repo = "1424772/ForgeCrew"
$Binary = "forgecrew.exe"

if ($InstallDir -eq "") {
    $InstallDir = "$env:USERPROFILE\.forgecrew\bin"
}

# ── Architecture detection ──
# Tries RuntimeInformation first (PS 7 / .NET Core),
# falls back to PROCESSOR_ARCHITEW6432 and PROCESSOR_ARCHITECTURE (PS 5.1).
$Arch = $null
try {
    $runtimeArch = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture
    $Arch = switch ($runtimeArch) {
        "X64"   { "amd64" }
        "Arm64" { "arm64" }
    }
} catch {
    # RuntimeInformation not available — fall back to env vars.
}

if (-not $Arch) {
    $envArch = $env:PROCESSOR_ARCHITEW6432
    if (-not $envArch) { $envArch = $env:PROCESSOR_ARCHITECTURE }
    $Arch = switch -Wildcard ($envArch) {
        "AMD64"   { "amd64" }
        "ARM64"   { "arm64" }
        "x86_64"  { "amd64" }
    }
}

if (-not $Arch) {
    Write-Host "[forgecrew] Unsupported architecture. Detected: $envArch" -ForegroundColor Red
    Write-Host "[forgecrew] This installer supports x64 and arm64 only." -ForegroundColor Red
    exit 1
}

$OS = "windows"

Write-Host "[forgecrew] Installing ForgeCrew $Version for $OS/$Arch..." -ForegroundColor Green
Write-Host "[forgecrew] Install directory: $InstallDir" -ForegroundColor Green

# ── Download URLs ──
$ReleaseBase = if ($Version -eq "latest") {
    "https://github.com/${Repo}/releases/latest/download"
} else {
    "https://github.com/${Repo}/releases/download/${Version}"
}

$BinaryUrl = "${ReleaseBase}/forgecrew_${OS}_${Arch}.exe"
$ChecksumUrl = "${ReleaseBase}/checksums.txt"

# ── Create install directory ──
if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
}

$DestFile = Join-Path $InstallDir $Binary
$TmpBinary = "$DestFile.tmp.$PID"
$TmpChecksum = "$DestFile.checksums.tmp.$PID"

# Clean up any leftover temp files from a previous failed run.
Remove-Item -Force $TmpBinary -ErrorAction SilentlyContinue | Out-Null
Remove-Item -Force $TmpChecksum -ErrorAction SilentlyContinue | Out-Null

# ── Download binary ──
$DownloadOk = $false
try {
    Write-Host "[forgecrew] Downloading $BinaryUrl ..." -ForegroundColor Green
    $ProgressPreference = 'SilentlyContinue'
    Invoke-WebRequest -Uri $BinaryUrl -OutFile $TmpBinary -ErrorAction Stop
    $DownloadOk = $true
} catch {
    Write-Host "[forgecrew] Binary download failed: $_" -ForegroundColor Yellow
}

# Validate binary temp file.
if ($DownloadOk) {
    if (-not (Test-Path $TmpBinary)) {
        Write-Host "[forgecrew] Downloaded binary file not found." -ForegroundColor Red
        $DownloadOk = $false
    } elseif ((Get-Item $TmpBinary).Length -eq 0) {
        Write-Host "[forgecrew] Downloaded binary file is empty." -ForegroundColor Red
        Remove-Item -Force $TmpBinary -ErrorAction SilentlyContinue | Out-Null
        $DownloadOk = $false
    }
}

# ── Download checksums.txt ──
$ChecksumOk = $false
try {
    Write-Host "[forgecrew] Downloading checksums.txt ..." -ForegroundColor Green
    Invoke-WebRequest -Uri $ChecksumUrl -OutFile $TmpChecksum -ErrorAction Stop
    $ChecksumOk = $true
} catch {
    Write-Host "[forgecrew] Checksum download failed: $_" -ForegroundColor Yellow
}

if ($ChecksumOk) {
    if (-not (Test-Path $TmpChecksum) -or (Get-Item $TmpChecksum).Length -eq 0) {
        Write-Host "[forgecrew] Checksum file empty or missing." -ForegroundColor Yellow
        $ChecksumOk = $false
    }
}

if (-not $DownloadOk) {
    Cleanup-Failure $TmpBinary $TmpChecksum
    Write-Host ""
    Write-Host "[forgecrew] Release not found at $BinaryUrl" -ForegroundColor Yellow
    Write-BuildInstructions $Repo $Binary $InstallDir
    exit 1
}

# ── SHA256 verification ──
if ($ChecksumOk) {
    $ExpectedHash = $null
    $BinaryBaseName = "forgecrew_${OS}_${Arch}.exe"
    foreach ($line in (Get-Content $TmpChecksum)) {
        if ($line -match "^\s*([0-9a-fA-F]{64})\s+.*${BinaryBaseName}") {
            $ExpectedHash = $Matches[1].ToLower()
            break
        }
    }
    if ($ExpectedHash) {
        Write-Host "[forgecrew] Verifying SHA256 checksum..." -ForegroundColor Green
        $ActualHash = (Get-FileHash -Path $TmpBinary -Algorithm SHA256).Hash.ToLower()
        if ($ActualHash -ne $ExpectedHash) {
            Write-Host "[forgecrew] SHA256 checksum mismatch!" -ForegroundColor Red
            Write-Host "[forgecrew] Expected: $ExpectedHash" -ForegroundColor Red
            Write-Host "[forgecrew] Actual:   $ActualHash" -ForegroundColor Red
            Cleanup-Failure $TmpBinary $TmpChecksum
            exit 1
        }
        Write-Host "[forgecrew] Checksum verified." -ForegroundColor Green
    } else {
        Write-Host "[forgecrew] Could not find checksum entry for $BinaryBaseName — skipping verification." -ForegroundColor Yellow
    }
    Remove-Item -Force $TmpChecksum -ErrorAction SilentlyContinue | Out-Null
} else {
    Write-Host "[forgecrew] Skipping checksum verification (checksums.txt not available)." -ForegroundColor Yellow
}

# ── Install binary ──
# Don't overwrite an existing working binary if verification fails later.
Move-Item -Force $TmpBinary $DestFile

# ── Self-test ──
Write-Host "[forgecrew] Running self-test: $DestFile version" -ForegroundColor Green
$selfTest = & $DestFile version 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "[forgecrew] Self-test failed: $selfTest" -ForegroundColor Red
    Write-Host "[forgecrew] The binary may be corrupted or incompatible with this system." -ForegroundColor Red
    Remove-Item -Force $DestFile -ErrorAction SilentlyContinue | Out-Null
    exit 1
}
Write-Host "[forgecrew] Self-test passed: $selfTest" -ForegroundColor Green

Write-Host "[forgecrew] ForgeCrew installed to $DestFile" -ForegroundColor Green
Write-Host ""

# ── Add to PATH ──
$p = [Environment]::GetEnvironmentVariable("Path", "User")
if ($p -notlike "*$InstallDir*") {
    Write-Host "[forgecrew] Adding $InstallDir to user PATH..." -ForegroundColor Green
    [Environment]::SetEnvironmentVariable("Path", "${p};${InstallDir}", "User")
    # Also add to current session PATH.
    $env:Path = "${env:Path};${InstallDir}"
    Write-Host "[forgecrew] Updated PATH for current session and user profile." -ForegroundColor Green
} else {
    # Ensure current session also has it.
    if ($env:Path -notlike "*$InstallDir*") {
        $env:Path = "${env:Path};${InstallDir}"
    }
}

Write-Host "[forgecrew] Run 'forgecrew init' in your project to get started." -ForegroundColor Green
Write-Host "[forgecrew] If the command is not found, restart your terminal or run:" -ForegroundColor Yellow
Write-Host "  `$env:Path += `";$InstallDir`"" -ForegroundColor Yellow

# ── Helpers ──

function Cleanup-Failure {
    param($f1, $f2)
    Remove-Item -Force $f1 -ErrorAction SilentlyContinue | Out-Null
    Remove-Item -Force $f2 -ErrorAction SilentlyContinue | Out-Null
}

function Write-BuildInstructions {
    param($repo, $binary, $dir)
    Write-Host ""
    Write-Host "[forgecrew] ForgeCrew binaries are not yet published to GitHub Releases." -ForegroundColor Yellow
    Write-Host "[forgecrew] You can build from source:" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "  git clone https://github.com/${repo}.git" -ForegroundColor Yellow
    Write-Host "  cd ForgeCrew" -ForegroundColor Yellow
    Write-Host "  go build -o ${binary} .\cmd\forgecrew" -ForegroundColor Yellow
    Write-Host "  move ${binary} ${dir}\" -ForegroundColor Yellow
    Write-Host ""

    $GoCmd = Get-Command go -ErrorAction SilentlyContinue
    if ($GoCmd) {
        Write-Host "[forgecrew] Go detected. You can also run:" -ForegroundColor Green
        Write-Host "  go install github.com/${repo}/cmd/forgecrew@latest" -ForegroundColor Green
    }
}
