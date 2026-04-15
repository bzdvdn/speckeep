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

if ($args.Count -lt 1) {
  Write-Error "Usage: archive-feature.ps1 <slug> [path] [--status <status>] [--reason \"...\"] [--copy] [--restore]"
  exit 2
}

# The speckeep `archive` command does not accept `--root`. Instead, pass the project
# root as the optional [path] argument so this wrapper can run from any cwd.
$slug = [string]$args[0]
$rest = @()
if ($args.Count -gt 1) {
  $rest = $args[1..($args.Count - 1)]
}

if ($rest.Count -ge 1 -and -not $rest[0].StartsWith("-")) {
  & (Join-Path $ScriptDir "run-speckeep.ps1") archive $slug @rest
  exit $LASTEXITCODE
}

& (Join-Path $ScriptDir "run-speckeep.ps1") archive $slug $RootDir @rest
exit $LASTEXITCODE
