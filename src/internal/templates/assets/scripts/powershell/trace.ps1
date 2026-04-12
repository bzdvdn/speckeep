$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
& "$ScriptDir/run-speckeep.ps1" trace $args
