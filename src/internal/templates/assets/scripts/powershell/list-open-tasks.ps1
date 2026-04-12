$ErrorActionPreference = "Stop"

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$RootDir = (Resolve-Path (Join-Path $ScriptDir "..\\..")).Path
& (Join-Path $ScriptDir "run-speckeep.ps1") __internal list-open-tasks --root $RootDir @args
exit $LASTEXITCODE
