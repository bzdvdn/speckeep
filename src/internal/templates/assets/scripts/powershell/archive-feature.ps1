$ErrorActionPreference = "Stop"

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$RootDir = (Resolve-Path (Join-Path $ScriptDir "..\\..")).Path

$hasRestore = $false
$hasStatus = $false
foreach ($arg in $args) {
  if ($arg -eq "--restore") { $hasRestore = $true }
  if ($arg -eq "--status") { $hasStatus = $true }
}
if (-not $hasRestore -and -not $hasStatus) {
  Write-Host "INFO: --status not provided; defaulting to completed (override via --status <status> [--reason \"...\"])."
}

& (Join-Path $ScriptDir "run-speckeep.ps1") archive --root $RootDir @args
exit $LASTEXITCODE
