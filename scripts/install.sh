#!/usr/bin/env bash
set -euo pipefail

REPO_OWNER="bzdvdn"
REPO_NAME="speckeep"

usage() {
  cat <<'EOF'
Install speckeep from GitHub Releases (Linux only).

Usage:
  install.sh [--version vX.Y.Z] [--bin-dir DIR] [--add-to-path]

Options:
  --version, -v   Release tag to install (default: latest)
  --bin-dir, -b   Install directory (default: ~/.local/bin)
  --add-to-path   Add install dir to PATH (writes to a profile file)
  --help, -h      Show help

Examples:
  ./install.sh
  ./install.sh --version v0.1.0 --bin-dir ~/.local/bin
  ./install.sh --version v0.1.0 --add-to-path
  sudo ./install.sh --version v0.1.0 --bin-dir /usr/local/bin
EOF
}

fail() {
  echo "error: $*" >&2
  exit 1
}

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || fail "missing required command: $1"
}

http_get() {
  local url="$1"
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$url"
    return 0
  fi
  if command -v wget >/dev/null 2>&1; then
    wget -qO- "$url"
    return 0
  fi
  fail "need curl or wget"
}

resolve_latest_tag() {
  local api="https://api.github.com/repos/${REPO_OWNER}/${REPO_NAME}/releases/latest"
  local json
  json="$(http_get "$api")"

  if command -v python3 >/dev/null 2>&1; then
    python3 - <<'PY' <<<"$json"
import json, sys
data = json.load(sys.stdin)
print(data["tag_name"])
PY
    return 0
  fi
  if command -v python >/dev/null 2>&1; then
    python - <<'PY' <<<"$json"
import json, sys
data = json.load(sys.stdin)
print(data["tag_name"])
PY
    return 0
  fi

  # Fallback: best-effort parsing without jq/python.
  echo "$json" | sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | head -n 1
}

detect_arch() {
  local machine
  machine="$(uname -m)"
  case "$machine" in
    x86_64|amd64) echo "amd64" ;;
    aarch64|arm64) echo "arm64" ;;
    *) fail "unsupported architecture: $machine (supported: x86_64, aarch64)" ;;
  esac
}

truthy() {
  case "${1:-}" in
    1|true|TRUE|yes|YES|y|Y|on|ON) return 0 ;;
    *) return 1 ;;
  esac
}

add_to_path_if_needed() {
  local bin_dir="$1"

  if [[ ":${PATH}:" == *":${bin_dir}:"* ]]; then
    return 0
  fi

  local shell_name profile_file
  shell_name="$(basename "${SHELL:-sh}")"
  if [[ "$shell_name" == "zsh" ]]; then
    profile_file="${HOME}/.zprofile"
  else
    profile_file="${HOME}/.profile"
  fi

  local export_line
  export_line="export PATH=\"${bin_dir}:\$PATH\""
  touch "$profile_file"

  if grep -Fq "$export_line" "$profile_file"; then
    return 0
  fi

  printf '\n# Added by speckeep installer\n%s\n' "$export_line" >>"$profile_file"
  echo "updated PATH in: ${profile_file}"
}

main() {
  local version="latest"
  local bin_dir="${SPECKEEP_INSTALL_DIR:-${DRAFTSPEC_INSTALL_DIR:-$HOME/.local/bin}}"
  local add_to_path="0"

  while [[ $# -gt 0 ]]; do
    case "$1" in
      -v|--version)
        [[ $# -ge 2 ]] || fail "--version requires a value"
        version="$2"
        shift 2
        ;;
      -b|--bin-dir)
        [[ $# -ge 2 ]] || fail "--bin-dir requires a value"
        bin_dir="$2"
        shift 2
        ;;
      --add-to-path)
        add_to_path="1"
        shift 1
        ;;
      -h|--help)
        usage
        exit 0
        ;;
      *)
        fail "unknown argument: $1 (use --help)"
        ;;
    esac
  done

  [[ "$(uname -s)" == "Linux" ]] || fail "this installer supports Linux only (use scripts/install.ps1 on Windows)"

  need_cmd tar
  need_cmd uname

  local arch
  arch="${SPECKEEP_ARCH:-${DRAFTSPEC_ARCH:-$(detect_arch)}}"

  if [[ "$version" == "latest" && -n "${SPECKEEP_VERSION:-}" ]]; then
    version="$SPECKEEP_VERSION"
  elif [[ "$version" == "latest" && -n "${DRAFTSPEC_VERSION:-}" ]]; then
    version="$DRAFTSPEC_VERSION"
  fi

  if [[ "$add_to_path" != "1" ]] && truthy "${SPECKEEP_ADD_TO_PATH:-}"; then
    add_to_path="1"
  elif [[ "$add_to_path" != "1" ]] && truthy "${DRAFTSPEC_ADD_TO_PATH:-}"; then
    add_to_path="1"
  fi

  if [[ "$version" == "latest" ]]; then
    version="$(resolve_latest_tag)"
    [[ -n "$version" ]] || fail "failed to resolve latest release tag (try --version vX.Y.Z)"
  fi

  local asset="speckeep_${version}_linux_${arch}.tar.gz"
  local url="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/${version}/${asset}"

  local tmpdir
  tmpdir="$(mktemp -d)"
  trap 'rm -rf "$tmpdir"' EXIT

  local archive="${tmpdir}/${asset}"
  if command -v curl >/dev/null 2>&1; then
    curl -fL --retry 3 --retry-delay 1 -o "$archive" "$url" || fail "download failed: $url"
  else
    wget -qO "$archive" "$url" || fail "download failed: $url"
  fi

  tar -C "$tmpdir" -xzf "$archive" || fail "failed to extract archive"
  [[ -f "${tmpdir}/speckeep" ]] || fail "archive did not contain expected 'speckeep' binary"

  mkdir -p "$bin_dir"
  install -m 0755 "${tmpdir}/speckeep" "${bin_dir}/speckeep"

  echo "installed: ${bin_dir}/speckeep"

  if [[ "$add_to_path" == "1" ]]; then
    add_to_path_if_needed "$bin_dir" || true
    echo "note: restart your shell (or source your profile) to pick up PATH changes"
  else
    if ! command -v speckeep >/dev/null 2>&1; then
      echo "note: '${bin_dir}' is not on PATH for this shell"
      echo "note: rerun with --add-to-path (or set SPECKEEP_ADD_TO_PATH=1) to update PATH"
    fi
  fi

  "${bin_dir}/speckeep" --version 2>/dev/null || true
}

main "$@"
