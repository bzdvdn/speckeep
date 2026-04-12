#!/bin/sh

set -eu

if [ "${SPECKEEP_BIN:-}" != "" ]; then
  if command -v "$SPECKEEP_BIN" >/dev/null 2>&1; then
    exec "$SPECKEEP_BIN" "$@"
  fi
  if [ -x "$SPECKEEP_BIN" ]; then
    exec "$SPECKEEP_BIN" "$@"
  fi
  echo "ERROR: SPECKEEP_BIN is set but could not be resolved: $SPECKEEP_BIN" >&2
  echo "Set SPECKEEP_BIN to an executable path or command name, or add speckeep to PATH." >&2
  exit 1
fi

if [ "${DRAFTSPEC_BIN:-}" != "" ]; then
  if command -v "$DRAFTSPEC_BIN" >/dev/null 2>&1; then
    exec "$DRAFTSPEC_BIN" "$@"
  fi
  if [ -x "$DRAFTSPEC_BIN" ]; then
    exec "$DRAFTSPEC_BIN" "$@"
  fi
  echo "ERROR: DRAFTSPEC_BIN is set but could not be resolved: $DRAFTSPEC_BIN" >&2
  echo "Set DRAFTSPEC_BIN to an executable path or command name, or add speckeep to PATH." >&2
  exit 1
fi

if command -v speckeep >/dev/null 2>&1; then
  exec speckeep "$@"
fi

if command -v draftspec >/dev/null 2>&1; then
  exec draftspec "$@"
fi

echo "ERROR: speckeep CLI not found." >&2
echo "Set SPECKEEP_BIN to an executable path or add speckeep to PATH." >&2
exit 1
