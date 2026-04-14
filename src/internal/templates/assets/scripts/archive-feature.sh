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

exec "$SCRIPT_DIR/run-speckeep.sh" archive --root "$ROOT_DIR" "$@"
