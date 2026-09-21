#!/bin/sh

set -eu

if [ $# -lt 1 ]; then
  echo "Usage: check-ready.sh <phase> [slug]" >&2
  echo "Phases: constitution, spec, propose, inspect, plan, tasks, implement, converge, verify, archive" >&2
  exit 2
fi

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"

ROOT_DIR="$(CDPATH= cd -- "$SCRIPT_DIR/../.." && pwd)"

PHASE="$1"
shift

case "$PHASE" in
  constitution|spec|propose|inspect|plan|tasks|implement|verify|converge|archive)
    exec "$SCRIPT_DIR/run-speckeep.sh" __internal "check-$PHASE-ready" --root "$ROOT_DIR" "$@"
    ;;
  *)
    echo "OK: no readiness gate for '$PHASE' (auxiliary command) - nothing to check"
    exit 0
    ;;
esac
