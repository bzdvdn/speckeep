$ErrorActionPreference = "Stop"

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$RootDir = (Resolve-Path (Join-Path $ScriptDir "..\\..")).Path
& (Join-Path $ScriptDir "run-speckeep.ps1") __internal check-spec-ready --root $RootDir @args
exit $LASTEXITCODE
