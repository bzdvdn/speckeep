#!/usr/bin/env sh
# speckeep release gate.
#
# Runs the checks that must pass before tagging a release:
#   L0 static analysis (gofmt, go vet, go test [-race])
#   L1/L3 Go suites (agent-artifact matrix + upgrade-from-legacy)
#   build matrix (linux/darwin/windows x amd64/arm64)   -- skipped with --quick
#   L2 black-box e2e against the freshly built binary (demo -> readiness ->
#      refresh idempotency -> doctor)
#   L4 packaging smoke (install.sh, npm launcher)
#
# Usage: sh scripts/release-check.sh [--quick]
#   --quick  skip `go test -race` and the cross-platform build matrix
set -u

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$ROOT_DIR" || exit 1

QUICK=0
for arg in "$@"; do
	case "$arg" in
	--quick) QUICK=1 ;;
	-h | --help)
		sed -n '2,16p' "$0"
		exit 0
		;;
	*)
		echo "unknown option: $arg" >&2
		exit 2
		;;
	esac
done

WORK=$(mktemp -d "${TMPDIR:-/tmp}/speckeep-release-check.XXXXXX")
trap 'rm -rf "$WORK"' EXIT INT TERM

PASS=0
FAIL=0
if [ -t 1 ]; then
	G=$(printf '\033[32m')
	R=$(printf '\033[31m')
	N=$(printf '\033[0m')
else
	G=
	R=
	N=
fi
ok() { PASS=$((PASS + 1)); printf '  %sPASS%s %s\n' "$G" "$N" "$1"; }
bad() { FAIL=$((FAIL + 1)); printf '  %sFAIL%s %s\n' "$R" "$N" "$1"; }
note() { printf '  ---- %s\n' "$1"; }
section() { printf '\n== %s ==\n' "$1"; }

section "L0 static analysis"
GOFMT_OUT=$(gofmt -l $(git ls-files '*.go' 2>/dev/null) 2>&1 || true)
if [ -z "$GOFMT_OUT" ]; then
	ok "gofmt"
else
	bad "gofmt: $GOFMT_OUT"
fi

if go vet ./... >"$WORK/vet.log" 2>&1; then
	ok "go vet"
else
	bad "go vet"
	tail -20 "$WORK/vet.log"
fi

if [ "$QUICK" -eq 0 ]; then
	if go test -race ./... >"$WORK/test-race.log" 2>&1; then
		ok "go test -race"
	else
		bad "go test -race (log: $WORK/test-race.log)"
		tail -20 "$WORK/test-race.log"
	fi
else
	if go test ./... >"$WORK/test.log" 2>&1; then
		ok "go test"
	else
		bad "go test (log: $WORK/test.log)"
		tail -20 "$WORK/test.log"
	fi
fi

section "L1/L3 Go suites"
if go test ./... -run 'TestAgentArtifactMatrix|TestUpgradeFromLegacyLayout' >"$WORK/matrix.log" 2>&1; then
	ok "artifact matrix + upgrade tests"
else
	bad "artifact matrix + upgrade tests"
	tail -30 "$WORK/matrix.log"
fi

if [ "$QUICK" -eq 0 ]; then
	section "build matrix"
	for goos in linux darwin windows; do
		for goarch in amd64 arm64; do
			if GOOS=$goos GOARCH=$goarch CGO_ENABLED=0 go build -o "$WORK/speckeep-$goos-$goarch" ./src/cmd/speckeep >/dev/null 2>&1; then
				ok "build $goos/$goarch"
			else
				bad "build $goos/$goarch"
			fi
		done
	done
fi

section "host binary"
BIN="$WORK/speckeep"
if CGO_ENABLED=0 go build -ldflags "-X speckeep/src/internal/cli.Version=v0.0.0-release-check" -o "$BIN" ./src/cmd/speckeep >/dev/null 2>&1; then
	ok "build host binary"
else
	bad "build host binary"
	printf '\n%s passed, %s failed\n' "$PASS" "$FAIL"
	exit 1
fi

# Pin the freshly built binary for generated wrapper scripts
# (`.speckeep/scripts/run-speckeep.sh` resolves SPECKEEP_BIN, else `speckeep`
# on PATH). Without this, gated readiness scripts fail in CI where the binary
# is not installed on PATH.
export SPECKEEP_BIN="$BIN"

section "L2 e2e: demo workspace (all targets)"
DEMO="$WORK/demo"
if "$BIN" demo "$DEMO" --agents all --shell sh >"$WORK/demo.log" 2>&1; then
	ok "demo workspace"
else
	bad "demo workspace"
	tail -10 "$WORK/demo.log"
fi

check_file() { if [ -f "$DEMO/$1" ]; then ok "generated $1"; else bad "missing $1"; fi; }
check_absent() { if [ -e "$DEMO/$1" ]; then bad "unexpected $1"; else ok "absent $1"; fi; }
check_file ".claude/skills/spk-spec/SKILL.md"
check_file ".opencode/skills/spk-spec/SKILL.md"
check_file ".windsurf/workflows/spk-spec.md"
check_file ".clinerules/workflows/spk-spec.md"
check_file ".gemini/commands/spk-spec.toml"
check_file ".aider/CONVENTIONS.md"
check_absent ".opencode/commands"

section "L2 e2e: readiness scripts"
if (cd "$DEMO" && ./.speckeep/scripts/check-ready.sh repo-map >"$WORK/ready-aux.log" 2>&1); then
	if grep -q "no readiness gate" "$WORK/ready-aux.log"; then
		ok "check-ready repo-map no-ops"
	else
		bad "check-ready repo-map must report no gate"
	fi
else
	bad "check-ready repo-map exited non-zero"
fi
if (cd "$DEMO" && ./.speckeep/scripts/check-ready.sh spec >/dev/null 2>&1); then
	ok "check-ready spec"
else
	bad "check-ready spec"
fi

section "L2 e2e: refresh idempotency + doctor"
"$BIN" refresh "$DEMO" >/dev/null 2>&1 || true
if "$BIN" refresh "$DEMO" >"$WORK/refresh2.log" 2>&1; then
	if grep -Eq '^(create|update|remove|rewrite) ' "$WORK/refresh2.log"; then
		bad "second refresh is not idempotent"
		grep -E '^(create|update|remove|rewrite) ' "$WORK/refresh2.log" | head
	else
		ok "refresh is idempotent"
	fi
else
	bad "refresh failed"
fi
if "$BIN" doctor "$DEMO" >/dev/null 2>&1; then
	ok "doctor clean"
else
	bad "doctor reports errors"
fi

section "L4 packaging smoke"
if command -v bash >/dev/null 2>&1; then
	if bash -n scripts/install.sh >/dev/null 2>&1; then ok "install.sh syntax"; else bad "install.sh syntax"; fi
else
	note "skip install.sh syntax (bash not found)"
fi

# npm launcher: must be tracked (a root `bin/` gitignore previously dropped it
# from the published tarball, so `npm install -g speckeep` created no shim),
# must parse, and must actually end up in `npm pack`.
if command -v node >/dev/null 2>&1 && command -v npm >/dev/null 2>&1; then
	if git ls-files --error-unmatch contrib/packaging/npm/bin/speckeep.js >/dev/null 2>&1; then
		ok "npm launcher tracked"
	else
		bad "npm launcher not tracked (is contrib/packaging/npm/bin ignored?)"
	fi
	if node --check contrib/packaging/npm/bin/speckeep.js >/dev/null 2>&1; then
		ok "npm launcher syntax"
	else
		bad "npm launcher syntax"
	fi
	if (cd contrib/packaging/npm && npm pack --dry-run --json 2>/dev/null | node -e '
let d = "";
process.stdin.on("data", (c) => (d += c));
process.stdin.on("end", () => {
  try {
    const packs = JSON.parse(d);
    const files = (packs[0] && packs[0].files) || [];
    process.exit(files.some((f) => f.path === "bin/speckeep.js") ? 0 : 1);
  } catch (e) {
    process.exit(1);
  }
});
'); then
		ok "npm pack includes bin/speckeep.js"
	else
		bad "npm pack is missing bin/speckeep.js"
	fi
else
	note "skip npm checks (node/npm not found)"
fi

printf '\n== summary ==\n%s passed, %s failed\n' "$PASS" "$FAIL"
[ "$FAIL" -eq 0 ]
