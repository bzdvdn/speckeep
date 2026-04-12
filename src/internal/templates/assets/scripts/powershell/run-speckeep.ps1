$ErrorActionPreference = "Stop"

$SpeckeepBin = $env:SPECKEEP_BIN
if (-not [string]::IsNullOrWhiteSpace($SpeckeepBin)) {
  $configuredCommand = Get-Command -Name $SpeckeepBin -ErrorAction SilentlyContinue
  if ($null -ne $configuredCommand) {
    & $SpeckeepBin @args
    exit $LASTEXITCODE
  }
  if (Test-Path -LiteralPath $SpeckeepBin -PathType Leaf) {
    & $SpeckeepBin @args
    exit $LASTEXITCODE
  }
  Write-Error "SPECKEEP_BIN is set but could not be resolved: $SpeckeepBin. Set SPECKEEP_BIN to an executable path or command name, or add speckeep to PATH."
  exit 1
}

$DraftspecBin = $env:DRAFTSPEC_BIN
if (-not [string]::IsNullOrWhiteSpace($DraftspecBin)) {
  $configuredCommand = Get-Command -Name $DraftspecBin -ErrorAction SilentlyContinue
  if ($null -ne $configuredCommand) {
    & $DraftspecBin @args
    exit $LASTEXITCODE
  }
  if (Test-Path -LiteralPath $DraftspecBin -PathType Leaf) {
    & $DraftspecBin @args
    exit $LASTEXITCODE
  }
  Write-Error "DRAFTSPEC_BIN is set but could not be resolved: $DraftspecBin. Set DRAFTSPEC_BIN to an executable path or command name, or add speckeep to PATH."
  exit 1
}

$defaultCommand = Get-Command -Name "speckeep" -ErrorAction SilentlyContinue
if ($null -ne $defaultCommand) {
  & speckeep @args
  exit $LASTEXITCODE
}

$legacyCommand = Get-Command -Name "draftspec" -ErrorAction SilentlyContinue
if ($null -ne $legacyCommand) {
  & draftspec @args
  exit $LASTEXITCODE
}

Write-Error "speckeep CLI not found. Set SPECKEEP_BIN to an executable path or add speckeep to PATH."
exit 1
