#!/bin/sh

set -eu

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"

ROOT_DIR="$(CDPATH= cd -- "$SCRIPT_DIR/../.." && pwd)"

if [ $# -eq 0 ]; then
  exec "$SCRIPT_DIR/run-speckeep.sh" trace "$ROOT_DIR"
fi

# Flags and subcommands take the root as a trailing positional argument.
case "$1" in
  -*|init|rebuild)
    exec "$SCRIPT_DIR/run-speckeep.sh" trace "$@" "$ROOT_DIR"
    ;;
esac

# If the first arg is a path, keep it; otherwise treat it as a slug and inject ROOT_DIR.
if [ -e "$1" ]; then
  exec "$SCRIPT_DIR/run-speckeep.sh" trace "$@"
fi
SLUG="$1"
shift
exec "$SCRIPT_DIR/run-speckeep.sh" trace "$SLUG" "$ROOT_DIR" "$@"
