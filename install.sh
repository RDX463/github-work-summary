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

resolve_asset_name() {
  local version="$1"
  local os="$2"
  local arch="$3"
  local release_api desired aliases release_json asset_urls

  release_api="https://api.github.com/repos/${REPO}/releases/tags/${version}"
  release_json="$(curl -fsSL "$release_api")" || fail "unable to read release metadata for ${version}"
  asset_urls="$(printf '%s' "$release_json" | grep -oE '"browser_download_url":[[:space:]]*"[^"]+"' | sed -E 's/"browser_download_url":[[:space:]]*"([^"]+)"/\1/')"
  [[ -n "$asset_urls" ]] || fail "release ${version} has no downloadable assets"

  desired="${TOOL_NAME}-${os}-${arch}.tar.gz"
  while IFS= read -r url; do
    local name
    name="${url##*/}"
    if [[ "$name" == "$desired" ]]; then
      printf '%s' "$name"
      return
    fi
  done <<< "$asset_urls"

  case "$arch" in
    amd64) aliases="amd64 x86_64" ;;
    arm64) aliases="arm64 aarch64" ;;
    *) aliases="$arch" ;;
  esac

  local alias
  for alias in $aliases; do
    while IFS= read -r url; do
      local name lower_name
      name="${url##*/}"
      lower_name="${name,,}"
      if [[ "$lower_name" == *"${TOOL_NAME}"* && "$lower_name" == *"${os}"* && "$lower_name" == *"${alias}"* && "$lower_name" == *.tar.gz ]]; then
        printf '%s' "$name"
        return
      fi
    done <<< "$asset_urls"
  done

  log "available assets in ${version}:"
  while IFS= read -r url; do
    log " - ${url##*/}"
  done <<< "$asset_urls"

  fail "no ${os}/${arch} .tar.gz asset found for ${version}"
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
  asset_name="$(resolve_asset_name "$version" "$os" "$arch")"
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

  local tool_path short_path
  tool_path="$(command -v "$TOOL_NAME" || true)"
  short_path="$(command -v "$SHORT_NAME" || true)"

  if [[ -z "$tool_path" || -z "$short_path" ]]; then
    log "note: ${install_dir} is not in your PATH for current shell."
    log "add this line to your shell profile:"
    log "export PATH=\"${install_dir}:\$PATH\""
  else
    log "verified in PATH: ${TOOL_NAME} and ${SHORT_NAME}"
  fi

  log "done. Run: ${SHORT_NAME} --help"
}

main "$@"
