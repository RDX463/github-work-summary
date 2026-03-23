#!/usr/bin/env bash
set -euo pipefail

TOOL_NAME="github-work-summary"
SHORT_NAME="gws"
REPO="${GWS_REPO:-RDX463/github-work-summary}"
VERSION_INPUT="${GWS_VERSION:-latest}"

log() {
  printf '[install] %s\n' "$*"
}

fail() {
  printf '[install] error: %s\n' "$*" >&2
  exit 1
}

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || fail "missing required command: $1"
}

detect_os() {
  local os
  os="$(uname -s | tr '[:upper:]' '[:lower:]')"
  case "$os" in
    linux) printf 'linux' ;;
    *) fail "this installer currently supports Linux only (detected: $os)" ;;
  esac
}

detect_arch() {
  local arch
  arch="$(uname -m)"
  case "$arch" in
    x86_64|amd64) printf 'amd64' ;;
    aarch64|arm64) printf 'arm64' ;;
    *) fail "unsupported architecture: $arch" ;;
  esac
}

resolve_install_dir() {
  if [[ -n "${GWS_INSTALL_DIR:-}" ]]; then
    printf '%s' "$GWS_INSTALL_DIR"
    return
  fi

  if [[ -d /usr/local/bin && -w /usr/local/bin ]]; then
    printf '/usr/local/bin'
    return
  fi

  mkdir -p "${HOME}/.local/bin"
  printf '%s' "${HOME}/.local/bin"
}

resolve_version() {
  if [[ "$VERSION_INPUT" != "latest" ]]; then
    if [[ "$VERSION_INPUT" == v* ]]; then
      printf '%s' "$VERSION_INPUT"
    else
      printf 'v%s' "$VERSION_INPUT"
    fi
    return
  fi

  local tag
  tag="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | sed -n 's/.*"tag_name":[[:space:]]*"\([^"]*\)".*/\1/p' | head -n1)"
  [[ -n "$tag" ]] || fail "unable to resolve latest release tag. You can set GWS_VERSION=vX.Y.Z"
  printf '%s' "$tag"
}

main() {
  require_cmd curl
  require_cmd tar
  require_cmd install
  require_cmd ln

  local os arch install_dir version asset_name download_url tmp_dir binary_path
  os="$(detect_os)"
  arch="$(detect_arch)"
  install_dir="$(resolve_install_dir)"
  version="$(resolve_version)"

  asset_name="${TOOL_NAME}-${os}-${arch}.tar.gz"
  download_url="https://github.com/${REPO}/releases/download/${version}/${asset_name}"

  log "repo: ${REPO}"
  log "version: ${version}"
  log "asset: ${asset_name}"
  log "install dir: ${install_dir}"

  tmp_dir="$(mktemp -d)"
  trap 'rm -rf "$tmp_dir"' EXIT

  log "downloading release..."
  curl -fL "$download_url" -o "${tmp_dir}/${asset_name}" || fail "download failed: ${download_url}"

  log "extracting archive..."
  tar -xzf "${tmp_dir}/${asset_name}" -C "${tmp_dir}"
  binary_path="${tmp_dir}/${TOOL_NAME}"
  [[ -f "$binary_path" ]] || fail "archive did not contain ${TOOL_NAME}"

  mkdir -p "$install_dir"

  if [[ -w "$install_dir" ]]; then
    install -m 0755 "$binary_path" "${install_dir}/${TOOL_NAME}"
    ln -sf "${install_dir}/${TOOL_NAME}" "${install_dir}/${SHORT_NAME}"
  else
    require_cmd sudo
    sudo install -m 0755 "$binary_path" "${install_dir}/${TOOL_NAME}"
    sudo ln -sf "${install_dir}/${TOOL_NAME}" "${install_dir}/${SHORT_NAME}"
  fi

  log "installed ${TOOL_NAME} to ${install_dir}/${TOOL_NAME}"
  log "shortcut available as ${SHORT_NAME}"

  if ! command -v "$TOOL_NAME" >/dev/null 2>&1; then
    log "note: ${install_dir} is not in your PATH for current shell."
    log "add this line to your shell profile:"
    log "export PATH=\"${install_dir}:\$PATH\""
  fi

  log "done. Run: ${SHORT_NAME} --help"
}

main "$@"
