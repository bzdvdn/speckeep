#!/bin/sh

set -eu

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"

ROOT_DIR="$(CDPATH= cd -- "$SCRIPT_DIR/../.." && pwd)"

CONFIG_PATH="$ROOT_DIR/.speckeep/speckeep.yaml"

yaml_first_value() {
  key="$1"
  if [ ! -f "$CONFIG_PATH" ]; then
    return 1
  fi
  # naive YAML extraction: first matching key in the file
  value="$(sed -n "s/^[[:space:]]*$key:[[:space:]]*//p" "$CONFIG_PATH" | head -n 1 | tr -d '\r')"
  value="$(printf "%s" "$value" | sed -e "s/^['\\\"]//" -e "s/['\\\"]$//")"
  if [ "$value" = "" ]; then
    return 1
  fi
  printf "%s" "$value"
  return 0
}

if [ $# -lt 1 ]; then
  echo "Usage: inspect-spec.sh <spec-file|slug> [tasks-file]" >&2
  exit 2
fi

input="$1"
tasks_input="${2:-}"

specs_dir="$(yaml_first_value specs_dir 2>/dev/null || true)"
if [ "$specs_dir" = "" ]; then
  specs_dir=".speckeep/specs"
fi

spec_file="$(yaml_first_value spec 2>/dev/null || true)"
if [ "$spec_file" = "" ]; then
  spec_file="spec.md"
fi

tasks_file="$(yaml_first_value tasks 2>/dev/null || true)"
if [ "$tasks_file" = "" ]; then
  tasks_file="tasks.md"
fi

spec_path="$input"
slug=""

if [ -f "$input" ]; then
  spec_path="$input"
elif [ -f "$ROOT_DIR/$input" ]; then
  spec_path="$input"
else
  slug="$input"
  spec_path="$specs_dir/$slug/$spec_file"
fi

if [ -f "$ROOT_DIR/$spec_path" ]; then
  # keep relative path
  :
elif [ -f "$spec_path" ]; then
  # absolute or relative to CWD, keep as-is
  :
else
  echo "ERROR: spec file not found: $spec_path" >&2
  exit 1
fi

tasks_path=""
if [ "$tasks_input" != "" ]; then
  tasks_path="$tasks_input"
elif [ "$slug" != "" ]; then
  candidate="$specs_dir/$slug/plan/$tasks_file"
  if [ -f "$ROOT_DIR/$candidate" ]; then
    tasks_path="$candidate"
  fi
fi

if [ "$tasks_path" = "" ]; then
  exec "$SCRIPT_DIR/run-speckeep.sh" __internal inspect-spec --root "$ROOT_DIR" "$spec_path"
fi
exec "$SCRIPT_DIR/run-speckeep.sh" __internal inspect-spec --root "$ROOT_DIR" "$spec_path" "$tasks_path"
