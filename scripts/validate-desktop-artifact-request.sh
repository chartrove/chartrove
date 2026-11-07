#!/bin/sh
set -eu

# Bind a separately authorized request; this script does not grant approval or
# prepare/upload files. The workflow repeats these conditions at the export step.
deny() {
  echo 'Desktop artifact export requires a first manual main run bound to its exact source SHA.' >&2
  exit 1
}

[ "${GITHUB_EVENT_NAME:-}" = workflow_dispatch ] || deny
[ "${GITHUB_REF:-}" = refs/heads/main ] || deny
[ "${GITHUB_RUN_ATTEMPT:-}" = 1 ] || deny
artifact_source_sha=${CHARTROVE_ARTIFACT_SOURCE_SHA:-}
[ "${#artifact_source_sha}" -eq 40 ] || deny
case "$artifact_source_sha" in *[!0-9a-f]*) deny ;; esac
[ "$artifact_source_sha" = "${GITHUB_SHA:-}" ] || deny
