#!/bin/sh

set -eu

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"

ROOT_DIR="$(CDPATH= cd -- "$SCRIPT_DIR/../.." && pwd)"

exec "$SCRIPT_DIR/run-speckeep.sh" __internal link-agents --root "$ROOT_DIR" "$@"
