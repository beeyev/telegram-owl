#!/usr/bin/env bash
#
# Install the thongtech/go-legacy-win7 Go toolchain on a GitHub Actions runner
# and make it the active Go for later steps.
#
# Env: VERSION (upstream release without "v", e.g. 1.27.1-1) plus the standard
# RUNNER_* and GITHUB_* runner variables. Kept Bash 3.2 compatible for macOS.

set -Eeuo pipefail

readonly RELEASES_URL='https://github.com/thongtech/go-legacy-win7/releases/download'

# Script scope: the EXIT trap runs after main has returned.
workdir=''

die() {
  printf '::error::%s\n' "$*" >&2
  exit 1
}

cleanup() {
  if [[ -n ${workdir} && -d ${workdir} ]]; then
    rm -rf -- "${workdir}"
  fi
  return 0
}

main() {
  local var os arch archive toolchain_dir go_version actual

  for var in VERSION RUNNER_OS RUNNER_ARCH RUNNER_TEMP GITHUB_ENV GITHUB_PATH GITHUB_OUTPUT; do
    [[ -n ${!var:-} ]] || die "${var} is not set"
  done

  [[ ${VERSION} =~ ^[0-9]+\.[0-9]+(\.[0-9]+)?-[0-9]+$ ]] \
    || die "Invalid version '${VERSION}', expected e.g. 1.27.1-1"

  case "${RUNNER_OS}" in
    Linux) os=linux ;;
    macOS) os=darwin ;;
    *) die "Unsupported runner OS: ${RUNNER_OS}" ;;
  esac
  case "${RUNNER_ARCH}" in
    X64) arch=amd64 ;;
    X86) arch=386 ;;
    ARM64) arch=arm64 ;;
    ARM) arch=arm ;;
    *) die "Unsupported runner arch: ${RUNNER_ARCH}" ;;
  esac

  archive="go-legacy-win7-${VERSION}.${os}_${arch}.tar.gz"
  toolchain_dir="${RUNNER_TEMP}/go-legacy-win7"
  go_version="${VERSION%-*}"

  # Download and unpack in a private directory so a failed run leaves nothing
  # half-extracted under the final path.
  trap cleanup EXIT
  workdir="$(mktemp -d "${RUNNER_TEMP}/go-legacy-win7.XXXXXX")"

  curl --proto '=https' --tlsv1.2 --fail --silent --show-error --location --retry 3 \
    --output "${workdir}/${archive}" "${RELEASES_URL}/v${VERSION}/${archive}"

  # Archive entries are "./go-legacy-win7/...".
  tar -xzf "${workdir}/${archive}" -C "${workdir}"
  [[ -x ${workdir}/go-legacy-win7/bin/go ]] \
    || die "Unexpected archive layout: go-legacy-win7/bin/go not found in ${archive}"

  rm -rf -- "${toolchain_dir}"
  mv -- "${workdir}/go-legacy-win7" "${toolchain_dir}"

  export GOROOT="${toolchain_dir}"
  export GOTOOLCHAIN=local
  export PATH="${GOROOT}/bin:${PATH}"

  actual="$(go version)"
  [[ ${actual} == "go version go${go_version} ${os}/${arch}" ]] \
    || die "Unexpected toolchain: ${actual}"

  printf 'GOROOT=%s\nGOTOOLCHAIN=%s\n' "${GOROOT}" "${GOTOOLCHAIN}" >> "${GITHUB_ENV}"
  printf '%s\n' "${GOROOT}/bin" >> "${GITHUB_PATH}"
  printf 'go-version=%s\ngoroot=%s\n' "${go_version}" "${GOROOT}" >> "${GITHUB_OUTPUT}"
  printf '%s\n' "${actual}"
}

main "$@"
