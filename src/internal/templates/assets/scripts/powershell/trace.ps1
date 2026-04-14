$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$RootDir = (Resolve-Path (Join-Path $ScriptDir "..\\..")).Path

if ($args.Count -eq 0) {
  & "$ScriptDir/run-speckeep.ps1" trace $RootDir @args
  exit $LASTEXITCODE
}

# If the first arg is a path, keep it; otherwise treat it as a slug and inject RootDir.
if (Test-Path -LiteralPath $args[0]) {
  & "$ScriptDir/run-speckeep.ps1" trace @args
  exit $LASTEXITCODE
}
& "$ScriptDir/run-speckeep.ps1" trace $args[0] $RootDir
exit $LASTEXITCODE
