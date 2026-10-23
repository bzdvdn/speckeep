$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$RootDir = (Resolve-Path (Join-Path $ScriptDir "..\\..")).Path

if ($args.Count -eq 0) {
  & "$ScriptDir/run-speckeep.ps1" trace $RootDir
  exit $LASTEXITCODE
}

# Flags and subcommands take the root as a trailing positional argument.
if ($args[0].StartsWith("-") -or $args[0] -eq "init" -or $args[0] -eq "rebuild") {
  & "$ScriptDir/run-speckeep.ps1" trace @args $RootDir
  exit $LASTEXITCODE
}

# If the first arg is a path, keep it; otherwise treat it as a slug and inject RootDir.
if (Test-Path -LiteralPath $args[0]) {
  & "$ScriptDir/run-speckeep.ps1" trace @args
  exit $LASTEXITCODE
}
$Slug = $args[0]
$Rest = @()
if ($args.Count -gt 1) {
  $Rest = $args[1..($args.Count - 1)]
}
& "$ScriptDir/run-speckeep.ps1" trace $Slug $RootDir @Rest
exit $LASTEXITCODE
