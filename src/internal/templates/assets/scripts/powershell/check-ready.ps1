$ErrorActionPreference = "Stop"

if ($args.Count -lt 1) {
  Write-Host "Usage: check-ready.ps1 <phase> [slug]" -ForegroundColor Red
  Write-Host "Phases: constitution, spec, propose, inspect, plan, tasks, implement, converge, verify, archive" -ForegroundColor Red
  exit 2
}

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$RootDir = (Resolve-Path (Join-Path $ScriptDir "..\\..")).Path
$Phase = $args[0]
$PhaseArgs = $args[1..$args.Count]

switch ($Phase) {
  { $_ -in @("constitution", "spec", "propose", "inspect", "plan", "tasks", "implement", "verify", "converge", "archive") } {
    & (Join-Path $ScriptDir "run-speckeep.ps1") __internal "check-$Phase-ready" --root $RootDir @PhaseArgs
    exit $LASTEXITCODE
  }
  default {
    Write-Host "OK: no readiness gate for '$Phase' (auxiliary command) - nothing to check"
    exit 0
  }
}
