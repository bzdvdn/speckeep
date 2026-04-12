$ErrorActionPreference = "Stop"

$RepoOwner = "bzdvdn"
$RepoName = "speckeep"

param(
  [string]$Version = "latest",
  [string]$BinDir = "",
  [switch]$AddToPath
)

function Fail([string]$Message) {
  throw $Message
}

function Resolve-LatestTag {
  $api = "https://api.github.com/repos/$RepoOwner/$RepoName/releases/latest"
  $headers = @{ "User-Agent" = "speckeep-install" }
  $release = Invoke-RestMethod -Uri $api -Headers $headers -Method Get
  if (-not $release.tag_name) { Fail "Failed to resolve latest release tag (try -Version vX.Y.Z)" }
  return [string]$release.tag_name
}

function Detect-Arch {
  $arch = $env:PROCESSOR_ARCHITECTURE
  if ($env:PROCESSOR_ARCHITEW6432) { $arch = $env:PROCESSOR_ARCHITEW6432 }
  switch ($arch) {
    "AMD64" { return "amd64" }
    "ARM64" { return "arm64" }
    default { Fail "Unsupported architecture: $arch (supported: AMD64, ARM64)" }
  }
}

if (-not $BinDir -or $BinDir.Trim().Length -eq 0) {
  if ($env:SPECKEEP_INSTALL_DIR) {
    $BinDir = $env:SPECKEEP_INSTALL_DIR
  } elseif ($env:DRAFTSPEC_INSTALL_DIR) {
    $BinDir = $env:DRAFTSPEC_INSTALL_DIR
  } else {
    $BinDir = Join-Path $env:LOCALAPPDATA "Programs\\speckeep\\bin"
  }
}

if (-not $AddToPath -and $env:SPECKEEP_ADD_TO_PATH) {
  $v = [string]$env:SPECKEEP_ADD_TO_PATH
  if ($v -match "^(1|true|yes|y|on)$") { $AddToPath = $true }
} elseif (-not $AddToPath -and $env:DRAFTSPEC_ADD_TO_PATH) {
  $v = [string]$env:DRAFTSPEC_ADD_TO_PATH
  if ($v -match "^(1|true|yes|y|on)$") { $AddToPath = $true }
}

if ($Version -eq "latest" -and $env:SPECKEEP_VERSION) {
  $Version = [string]$env:SPECKEEP_VERSION
} elseif ($Version -eq "latest" -and $env:DRAFTSPEC_VERSION) {
  $Version = [string]$env:DRAFTSPEC_VERSION
}

if ($Version -eq "latest") {
  $Version = Resolve-LatestTag
}

$arch = if ($env:SPECKEEP_ARCH) { $env:SPECKEEP_ARCH } elseif ($env:DRAFTSPEC_ARCH) { $env:DRAFTSPEC_ARCH } else { Detect-Arch }
$asset = "speckeep_{0}_windows_{1}.zip" -f $Version, $arch
$url = "https://github.com/$RepoOwner/$RepoName/releases/download/$Version/$asset"

$tmpDir = Join-Path ([System.IO.Path]::GetTempPath()) ("speckeep-install-" + [System.Guid]::NewGuid().ToString("N"))
New-Item -ItemType Directory -Path $tmpDir | Out-Null

try {
  $zipPath = Join-Path $tmpDir $asset
  Invoke-WebRequest -Uri $url -OutFile $zipPath -UseBasicParsing

  $extractDir = Join-Path $tmpDir "extract"
  Expand-Archive -Path $zipPath -DestinationPath $extractDir -Force

  $exe = Join-Path $extractDir "speckeep.exe"
  if (-not (Test-Path $exe)) { Fail "Archive did not contain expected speckeep.exe" }

  New-Item -ItemType Directory -Path $BinDir -Force | Out-Null
  Copy-Item -Path $exe -Destination (Join-Path $BinDir "speckeep.exe") -Force

  Write-Host ("installed: {0}" -f (Join-Path $BinDir "speckeep.exe"))

  if ($AddToPath) {
    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    $parts = @()
    if ($userPath) { $parts = $userPath -split ";" }
    if (-not ($parts | Where-Object { $_ -eq $BinDir })) {
      $newPath = @($parts + $BinDir | Where-Object { $_ -and $_.Trim().Length -gt 0 } | Select-Object -Unique) -join ";"
      [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
      Write-Host ("updated user PATH to include: {0}" -f $BinDir)
    }
    if (-not ($env:PATH -split ";" | Where-Object { $_ -eq $BinDir })) {
      $env:PATH = "$env:PATH;$BinDir"
    }
    Write-Host "note: open a new terminal to pick up PATH changes everywhere"
  } else {
    if (-not ($env:PATH -split ";" | Where-Object { $_ -eq $BinDir })) {
      Write-Host ("note: '{0}' is not on PATH for this session" -f $BinDir)
      Write-Host "note: rerun with -AddToPath (or set SPECKEEP_ADD_TO_PATH=1) to update PATH"
    }
  }

  & (Join-Path $BinDir "speckeep.exe") --version 2>$null
} finally {
  Remove-Item -Recurse -Force $tmpDir -ErrorAction SilentlyContinue
}
