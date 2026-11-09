#!/usr/bin/env sh
# Derive the image's default config from the public standalone sample. Never use
# an image ENV default: it would override an operator's explicit mounted config.
set -eu

usage() {
  printf 'Usage: sh scripts/container-config.sh SOURCE NEW_DESTINATION\n'
}

die() {
  printf 'container-config: %s\n' "$1" >&2
  exit 1
}

main() {
  if [ "$#" -eq 1 ] && { [ "$1" = '-h' ] || [ "$1" = '--help' ]; }; then
    usage
    return 0
  fi
  [ "$#" -eq 2 ] || {
    usage >&2
    return 2
  }
  command -v grep > /dev/null 2>&1 || die 'grep is required'
  command -v sed > /dev/null 2>&1 || die 'sed is required'

  # A changed/ambiguous source must fail the build rather than silently produce
  # an unusable or differently exposed image. No output is created in this case.
  if count=$(grep -Fxc '    host: 127.0.0.1' < "$1"); then
    :
  else
    status=$?
    [ "$status" -eq 1 ] || die 'cannot read the standalone sample'
  fi
  [ "$count" = 1 ] || die 'expected exactly one standalone loopback host line'

  # Public sample only; retain the caller's umask. Noclobber refuses an existing
  # destination (including the source), preserving mounted/user configuration.
  set -C
  sed 's/^    host: 127\.0\.0\.1$/    host: 0.0.0.0/' < "$1" > "$2"
}

main "$@"
