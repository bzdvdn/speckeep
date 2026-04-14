$ErrorActionPreference = "Stop"

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$RootDir = (Resolve-Path (Join-Path $ScriptDir "..\\..")).Path

function Get-YamlFirstValue {
  param(
    [Parameter(Mandatory = $true)][string]$Path,
    [Parameter(Mandatory = $true)][string]$Key
  )
  if (-not (Test-Path -LiteralPath $Path)) {
    return $null
  }
  $pattern = "^\s*$([Regex]::Escape($Key))\s*:\s*(.+)\s*$"
  foreach ($line in Get-Content -LiteralPath $Path) {
    $m = [Regex]::Match($line, $pattern)
    if ($m.Success) {
      $value = $m.Groups[1].Value.Trim()
      if (($value.StartsWith("'") -and $value.EndsWith("'")) -or ($value.StartsWith('"') -and $value.EndsWith('"'))) {
        $value = $value.Substring(1, $value.Length - 2)
      }
      if ($value -ne "") {
        return $value
      }
    }
  }
  return $null
}

if ($args.Count -lt 1) {
  Write-Error "Usage: inspect-spec.ps1 <spec-file|slug> [tasks-file]"
  exit 2
}

$configPath = Join-Path $RootDir ".speckeep\\speckeep.yaml"
$specsDir = (Get-YamlFirstValue -Path $configPath -Key "specs_dir")
if ([string]::IsNullOrWhiteSpace($specsDir)) { $specsDir = ".speckeep/specs" }
$specFile = (Get-YamlFirstValue -Path $configPath -Key "spec")
if ([string]::IsNullOrWhiteSpace($specFile)) { $specFile = "spec.md" }
$tasksFile = (Get-YamlFirstValue -Path $configPath -Key "tasks")
if ([string]::IsNullOrWhiteSpace($tasksFile)) { $tasksFile = "tasks.md" }

$input = [string]$args[0]
$tasksInput = $null
if ($args.Count -ge 2) { $tasksInput = [string]$args[1] }

$slug = $null
$specPath = $input
if (Test-Path -LiteralPath $specPath) {
  # ok: literal path
} elseif (Test-Path -LiteralPath (Join-Path $RootDir $specPath)) {
  # ok: relative to root
} else {
  $slug = $input
  $specPath = "$specsDir/$slug/$specFile"
}

if (-not (Test-Path -LiteralPath $specPath) -and -not (Test-Path -LiteralPath (Join-Path $RootDir $specPath))) {
  Write-Error "ERROR: spec file not found: $specPath"
  exit 1
}

$tasksPath = $null
if (-not [string]::IsNullOrWhiteSpace($tasksInput)) {
  $tasksPath = $tasksInput
} elseif (-not [string]::IsNullOrWhiteSpace($slug)) {
  $candidate = "$specsDir/$slug/plan/$tasksFile"
  if (Test-Path -LiteralPath (Join-Path $RootDir $candidate)) {
    $tasksPath = $candidate
  }
}

if ([string]::IsNullOrWhiteSpace($tasksPath)) {
  & (Join-Path $ScriptDir "run-speckeep.ps1") __internal inspect-spec --root $RootDir $specPath
  exit $LASTEXITCODE
}
& (Join-Path $ScriptDir "run-speckeep.ps1") __internal inspect-spec --root $RootDir $specPath $tasksPath
exit $LASTEXITCODE
