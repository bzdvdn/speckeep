$ErrorActionPreference = "Stop"

if ($args.Count -lt 1) {
  Write-Host "Usage: check-ready.ps1 <phase> [slug]" -ForegroundColor Red
  Write-Host "Phases: constitution, spec, inspect, plan, tasks, implement, verify, archive" -ForegroundColor Red
  exit 2
}

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$RootDir = (Resolve-Path (Join-Path $ScriptDir "..\\..")).Path
$Phase = $args[0]
$PhaseArgs = $args[1..$args.Count]

& (Join-Path $ScriptDir "run-speckeep.ps1") __internal "check-$Phase-ready" --root $RootDir @PhaseArgs
exit $LASTEXITCODE
