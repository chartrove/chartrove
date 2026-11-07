#!/bin/sh
# Exercise Docker's actual context matcher using synthetic files only.
set -eu
repo_root=$(CDPATH='' cd -- "$(dirname -- "$0")/.." && pwd)
case "${DOCKER_HOST:-unix://local}" in
  unix://*) ;;
  *) echo 'This fixture requires a local Docker Unix socket.' >&2; exit 1 ;;
esac
docker_context=$(docker context show)
endpoint=$(docker context inspect "$docker_context" --format '{{(index .Endpoints "docker").Host}}')
case "$endpoint" in
  unix://*) ;;
  *) echo 'This fixture refuses a remote Docker context.' >&2; exit 1 ;;
esac
driver=$(docker buildx inspect "$docker_context" | awk '$1 == "Driver:" { print $2 }')
[ "$driver" = docker ] || { echo 'This fixture requires the local Docker-context builder.' >&2; exit 1; }
test_root=$(mktemp -d "${TMPDIR:-/tmp}/chartrove-context-test.XXXXXX")
trap 'rm -rf "$test_root"' EXIT
trap 'exit 130' HUP INT TERM
mkdir -p "$test_root/context/backend" "$test_root/context/frontend"
printf 'FROM scratch\nCOPY . /snapshot/\n' > "$test_root/context/Dockerfile"
public_paths='backend/source.go frontend/package.json config.yaml'
private_paths='.env .env.local .direnv/state .couchdb/state config.dev.yaml healthit.yaml certs/private.key db/attachment.bin cache/record.bin .cache/record.bin fasten.db fasten.db-shm fasten.db-wal fasten-household.db fasten-household.db-shm fasten-household.db-wal nested/fasten.db nested/fasten-family.db-wal'
for path in $public_paths $private_paths; do
  mkdir -p "$test_root/context/$(dirname "$path")"
  printf 'synthetic build-context marker\n' > "$test_root/context/$path"
done
# Frozen accepted-main 6a5260bffcb1 context rules are a positive leak control.
# These are patterns only; no existing private file is read or sent to Docker.
cat > "$test_root/original.dockerignore" <<'ORIGINAL'
/vendor
/.idea
/.github
/.git
/frontend/node_modules
/frontend/.angular
/dist
ORIGINAL
# Exercise the same synthetic fixture first against the original rules, then
# against the current source. Only the second build must exclude private paths.
for variant in original current; do
  if [ "$variant" = original ]; then
    cp "$test_root/original.dockerignore" "$test_root/context/.dockerignore"
  else
    cp "$repo_root/.dockerignore" "$test_root/context/.dockerignore"
  fi
  # No RUN, base image, registry, runtime container, tag or production context.
  DOCKER_BUILDKIT=1 docker --context "$docker_context" build --builder "$docker_context" \
    --pull=false --network=none --no-cache --progress=plain \
    --output "type=local,dest=$test_root/$variant" "$test_root/context"
  failed=0
  for path in $public_paths; do
    if [ ! -f "$test_root/$variant/snapshot/$path" ]; then
      echo "Missing public fixture ($variant): $path" >&2
      failed=1
    fi
  done
  for path in $private_paths; do
    if [ "$variant" = original ]; then
      if [ ! -f "$test_root/$variant/snapshot/$path" ]; then
        echo "Original-rule control did not retain fixture: $path" >&2
        failed=1
      fi
    elif [ -e "$test_root/$variant/snapshot/$path" ]; then
      echo "Private fixture entered build context: $path" >&2
      failed=1
    fi
  done
  [ "$failed" -eq 0 ] || exit 1
done
echo 'Synthetic private-path build-context regression passed (original/current controls).'
