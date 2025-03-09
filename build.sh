#!/usr/bin/env bash
set -euo pipefail

: "${MZ_CABI_INCLUDE_DIR:=include}"
: "${MZ_CABI_LIB_DIR:=lib}"

# Check that required commands are available.
function check_requirements() {
  if ! command -v cargo >/dev/null; then
    echo ">>> A Rust toolchain is required to build minijinja." >&2
    echo ">>> Please see https://www.rust-lang.org/tools/install for instructions." >&2
    return 1
  fi

  if ! command -v jq >/dev/null; then
    echo ">>> The jq command is required by the build script." >&2
    echo ">>> Please see https://jqlang.org/download for instructions." >&2
    return 1
  fi
}

# Ensure that the src/lib.rs file exists. Cargo complains if there is neither
# src/lib.rs nor src/main.rs file. In the case of a vendored Go module, only
# the module directory is present, everything else is omitted.
function ensure_src_exists() {
  mkdir -p src
  touch src/lib.rs
}

# Retrieve the source directory from `cargo metadata`.
function get_mj_cabi_source_dir() {
  local manifest_path
  manifest_path="$(
    cargo metadata --format-version=1 --locked \
      | jq --raw-output '.packages[] | select(.name == "minijinja-cabi") | .manifest_path'
  )"
  echo "${manifest_path%/*}"
}

# Get the dynamic library file extension for the current platform.
function get_lib_extension() {
  local platform
  platform="$(uname -s)"

  case "${platform}" in
    Darwin)
      echo 'dylib'
      ;;
    Linux)
      echo 'so'
      ;;
    *)
      printf '>>> Unsupported platform: %s\n' "${platform}" >&2
      return 1
      ;;
  esac
}

# Build the minijinja-cabi library and install the build output under the lib/
# directory.
function build_mj_cabi_lib() {
  cargo build --locked --package=minijinja-cabi --release
  rm -rf lib # do not use the var, only remove local dir
  mkdir -p "${MZ_CABI_LIB_DIR}"
  EXT="$(get_lib_extension)"
  mv "target/release/libminijinja_cabi.${EXT}" "${MZ_CABI_LIB_DIR}"
}

# Copy the minijinja-cabi include directory to the include/ directory.
function copy_mj_cabi_headers() {
  local source_dir="${1?source_dir must be provided}"
  rm -rf include # do not use the var, only remove local dir
  mkdir -p "${MZ_CABI_INCLUDE_DIR}"
  cp -a "${source_dir}"/include/minijinja.h "${MZ_CABI_INCLUDE_DIR}"
}

function main() {
  check_requirements

  ROOT_DIR="$(cd -P "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
  cd "${ROOT_DIR}"

  ensure_src_exists

  echo '>>> Getting minijinja-cabi source directory...'
  local mj_cabi_source_dir
  mj_cabi_source_dir="$(get_mj_cabi_source_dir)"

  echo ">>> Building minijinja-cabi..."
  build_mj_cabi_lib

  echo ">>> Copying minijinja-cabi C headers..."
  copy_mj_cabi_headers "${mj_cabi_source_dir}"

  echo '>>> Done!'
}

main "${@:-}"
