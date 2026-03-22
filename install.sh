#!/usr/bin/env sh
# SPDX-FileCopyrightText: 2026 Elwan Mayencourt <mayencourt@elwan.ch>
# SPDX-License-Identifier: MIT

set -eu

PROJECT="l1"
OWNER="YungBricoCoop"
REPO="l1"

BINDIR=""
VERSION_ARG="latest"
DEBUG=0

usage() {
  cat <<USAGE
Install ${PROJECT} from GitHub Releases.

Usage:
  $0 [-b <bindir>] [-d] [<version>]

Options:
  -b <bindir>  Install directory (default: /usr/local/bin on Unix, ~/bin on Windows Git Bash)
  -d           Enable debug logs
  -h           Show this help

Examples:
  curl -sSfL https://raw.githubusercontent.com/${OWNER}/${REPO}/main/install.sh | sh
  wget -qO- https://raw.githubusercontent.com/${OWNER}/${REPO}/main/install.sh | sh
  curl -sSfL https://raw.githubusercontent.com/${OWNER}/${REPO}/main/install.sh | sh -s -- v1.2.3
  curl -sSfL https://raw.githubusercontent.com/${OWNER}/${REPO}/main/install.sh | sh -s -- -b /usr/local/bin v1.2.3
USAGE
}

log() {
  printf '%s\n' "$*"
}

debug() {
  if [ "$DEBUG" -eq 1 ]; then
    log "[debug] $*"
  fi
}

fail() {
  log "[error] $*" >&2
  exit 1
}

has_cmd() {
  command -v "$1" >/dev/null 2>&1
}

http_get() {
  url="$1"
  if has_cmd curl; then
    curl -fsSL "$url"
    return
  fi
  if has_cmd wget; then
    wget -qO- "$url"
    return
  fi
  fail "curl or wget is required"
}

download_to() {
  url="$1"
  output="$2"
  if has_cmd curl; then
    curl -fsSL "$url" -o "$output"
    return
  fi
  if has_cmd wget; then
    wget -qO "$output" "$url"
    return
  fi
  fail "curl or wget is required"
}

detect_os() {
  os_name=$(uname -s | tr '[:upper:]' '[:lower:]')
  case "$os_name" in
    linux*) printf 'linux' ;;
    darwin*) printf 'darwin' ;;
    msys*|mingw*|cygwin*) printf 'windows' ;;
    *) fail "unsupported OS: $os_name" ;;
  esac
}

detect_arch() {
  arch_name=$(uname -m | tr '[:upper:]' '[:lower:]')
  case "$arch_name" in
    x86_64|amd64) printf 'amd64' ;;
    arm64|aarch64) printf 'arm64' ;;
    *) fail "unsupported architecture: $arch_name" ;;
  esac
}

resolve_tag() {
  requested="$1"
  if [ -z "$requested" ] || [ "$requested" = "latest" ]; then
    debug "resolving latest release tag"
    tag=$(http_get "https://api.github.com/repos/${OWNER}/${REPO}/releases/latest" | sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p')
    if [ -z "$tag" ]; then
      fail "unable to resolve latest release tag"
    fi
    printf '%s' "$tag"
    return
  fi

  case "$requested" in
    v*) printf '%s' "$requested" ;;
    *) printf 'v%s' "$requested" ;;
  esac
}

verify_checksum() {
  archive_path="$1"
  archive_name="$2"
  checksums_path="$3"

  expected=$(grep "  ${archive_name}$" "$checksums_path" | awk '{print $1}' || true)
  if [ -z "$expected" ]; then
    fail "checksum for ${archive_name} not found"
  fi

  if has_cmd sha256sum; then
    actual=$(sha256sum "$archive_path" | awk '{print $1}')
  elif has_cmd shasum; then
    actual=$(shasum -a 256 "$archive_path" | awk '{print $1}')
  else
    log "[warn] sha256sum/shasum not found; skipping checksum verification"
    return
  fi

  if [ "$expected" != "$actual" ]; then
    fail "checksum mismatch for ${archive_name}"
  fi
}

extract_archive() {
  archive_path="$1"
  ext="$2"
  extract_dir="$3"

  mkdir -p "$extract_dir"
  case "$ext" in
    tar.gz)
      tar -xzf "$archive_path" -C "$extract_dir"
      ;;
    zip)
      if ! has_cmd unzip; then
        fail "unzip is required to extract ${archive_path}"
      fi
      unzip -q "$archive_path" -d "$extract_dir"
      ;;
    *)
      fail "unsupported archive type: $ext"
      ;;
  esac
}

choose_default_bindir() {
  os="$1"
  if [ "$os" = "windows" ]; then
    printf '%s' "${HOME}/bin"
  else
    printf '%s' "/usr/local/bin"
  fi
}

install_binary() {
  source_path="$1"
  install_dir="$2"
  binary_name="$3"
  os="$4"

  use_sudo=0
  target_dir="$install_dir"

  if [ -d "$target_dir" ] && [ ! -w "$target_dir" ] && [ "$os" != "windows" ]; then
    if [ "$target_dir" = "/usr/local/bin" ] && has_cmd sudo; then
      use_sudo=1
    else
      target_dir="${HOME}/.local/bin"
      log "[warn] no write permission for ${install_dir}; installing to ${target_dir}"
    fi
  fi

  if [ "$use_sudo" -eq 1 ]; then
    sudo mkdir -p "$target_dir"
    if has_cmd install; then
      sudo install -m 0755 "$source_path" "${target_dir}/${binary_name}"
    else
      sudo cp "$source_path" "${target_dir}/${binary_name}"
      sudo chmod 0755 "${target_dir}/${binary_name}"
    fi
  else
    mkdir -p "$target_dir"
    if has_cmd install; then
      install -m 0755 "$source_path" "${target_dir}/${binary_name}"
    else
      cp "$source_path" "${target_dir}/${binary_name}"
      chmod 0755 "${target_dir}/${binary_name}"
    fi
  fi

  printf '%s' "$target_dir"
}

while getopts "b:dh" opt; do
  case "$opt" in
    b) BINDIR="$OPTARG" ;;
    d) DEBUG=1 ;;
    h)
      usage
      exit 0
      ;;
    *)
      usage
      exit 1
      ;;
  esac
done
shift $((OPTIND - 1))

if [ "$#" -gt 1 ]; then
  usage
  exit 1
fi

if [ "$#" -eq 1 ]; then
  VERSION_ARG="$1"
fi

OS=$(detect_os)
ARCH=$(detect_arch)
TAG=$(resolve_tag "$VERSION_ARG")
VERSION=${TAG#v}

if [ -z "$BINDIR" ]; then
  BINDIR=$(choose_default_bindir "$OS")
fi

if [ "$OS" = "windows" ]; then
  EXT="zip"
  BIN_NAME="${PROJECT}.exe"
else
  EXT="tar.gz"
  BIN_NAME="${PROJECT}"
fi

ARCHIVE="${PROJECT}_${VERSION}_${OS}_${ARCH}.${EXT}"
CHECKSUMS="checksums.txt"
RELEASE_BASE="https://github.com/${OWNER}/${REPO}/releases/download/${TAG}"
ARCHIVE_URL="${RELEASE_BASE}/${ARCHIVE}"
CHECKSUMS_URL="${RELEASE_BASE}/${CHECKSUMS}"

TMP_DIR=$(mktemp -d 2>/dev/null || mktemp -d -t "${PROJECT}-install")
trap 'rm -rf "$TMP_DIR"' EXIT INT TERM

debug "OS=${OS} ARCH=${ARCH} TAG=${TAG}"
debug "archive=${ARCHIVE}"

download_to "$ARCHIVE_URL" "$TMP_DIR/$ARCHIVE"
download_to "$CHECKSUMS_URL" "$TMP_DIR/$CHECKSUMS"

verify_checksum "$TMP_DIR/$ARCHIVE" "$ARCHIVE" "$TMP_DIR/$CHECKSUMS"
extract_archive "$TMP_DIR/$ARCHIVE" "$EXT" "$TMP_DIR/extract"

SOURCE_BIN="$TMP_DIR/extract/$BIN_NAME"
if [ ! -f "$SOURCE_BIN" ]; then
  fail "binary ${BIN_NAME} not found in archive"
fi

INSTALLED_DIR=$(install_binary "$SOURCE_BIN" "$BINDIR" "$BIN_NAME" "$OS")

log "${PROJECT} ${TAG} installed to ${INSTALLED_DIR}/${BIN_NAME}"
if [ "$OS" = "windows" ]; then
  log "On Windows with Git Bash, ensure ${INSTALLED_DIR} is in PATH."
  log "Example: echo 'export PATH=\"${INSTALLED_DIR}:\$PATH\"' >> ~/.bashrc"
elif [ "$INSTALLED_DIR" = "${HOME}/.local/bin" ] || [ "$INSTALLED_DIR" = "${HOME}/bin" ]; then
  log "Add ${INSTALLED_DIR} to your PATH if it is not already present."
fi
