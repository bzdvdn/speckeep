#!/bin/sh

set -eu

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"

ROOT_DIR="$(CDPATH= cd -- "$SCRIPT_DIR/../.." && pwd)"

case " $* " in
  *" --restore "*) ;;
  *" --status "*) ;;
  *)
    echo "INFO: --status not provided; defaulting to completed (override via --status <status> [--reason \"...\"])." >&2
    ;;
esac

if [ $# -lt 1 ]; then
  echo "Usage: archive-feature.sh <slug> [path] [--status <status>] [--reason \"...\"] [--copy] [--restore]" >&2
  exit 2
fi

# The speckeep `archive` command does not accept `--root`. Instead, pass the project
# root as the optional [path] argument so this wrapper can run from any cwd.
slug="$1"
shift

if [ $# -ge 1 ] && [ "${1#-}" = "$1" ]; then
  # path explicitly provided by caller
  exec "$SCRIPT_DIR/run-speckeep.sh" archive "$slug" "$@"
fi

exec "$SCRIPT_DIR/run-speckeep.sh" archive "$slug" "$ROOT_DIR" "$@"
