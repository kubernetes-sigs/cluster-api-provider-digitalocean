#!/usr/bin/env bash

# Copyright 2026 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Copy DigitalOcean snapshots to other regions. See docs/images.md.

set -o nounset
set -o pipefail

usage() {
  cat <<EOF
Usage: $(basename "$0") [-r region,region,...] <snapshot-id>...

Copies each snapshot to every available region (default), or to the given regions.
Regions a snapshot is already in are skipped. Transfers run in parallel.

Options:
  -r <regions>  Comma-separated region slugs, e.g. nyc3,ams3. (default: all available regions)
  -h            Show this help.

Environment:
  DIGITALOCEAN_ACCESS_TOKEN   Required. API token of the account that owns the snapshots.
EOF
}

TARGET_REGIONS=""
while getopts ":r:h" opt; do
  case "${opt}" in
    r) TARGET_REGIONS="${OPTARG//,/ }" ;;
    h) usage; exit 0 ;;
    *) usage >&2; exit 2 ;;
  esac
done
shift $((OPTIND - 1))

if [[ $# -eq 0 ]]; then
  usage >&2
  exit 2
fi
if [[ -z "${DIGITALOCEAN_ACCESS_TOKEN:-}" ]]; then
  echo "DIGITALOCEAN_ACCESS_TOKEN must be set" >&2
  exit 1
fi
command -v doctl >/dev/null || { echo "doctl is required" >&2; exit 1; }

if [[ -z "${TARGET_REGIONS}" ]]; then
  TARGET_REGIONS="$(doctl compute region list --format Slug,Available --no-header | awk '$2 == "true" {print $1}')"
fi

LOGDIR="$(mktemp -d)"
pids=()
for id in "$@"; do
  current="$(doctl compute snapshot get "${id}" --format Regions --no-header | tr -d '[]' | tr -s ' ,' '\n')"
  for region in ${TARGET_REGIONS}; do
    if grep -qx "${region}" <<<"${current}"; then
      continue
    fi
    (
      if doctl compute image-action transfer "${id}" --region "${region}" --wait >"${LOGDIR}/${id}-${region}.log" 2>&1; then
        echo "OK      ${id} -> ${region}"
      else
        echo "FAILED  ${id} -> ${region}: $(tail -n 1 "${LOGDIR}/${id}-${region}.log")"
      fi
    ) &
    pids+=("$!")
  done
done

# ${pids[@]+...} keeps bash 3.2 (macOS) happy when nothing needed transferring.
for pid in ${pids[@]+"${pids[@]}"}; do
  wait "${pid}"
done

echo
echo "=== Regions per snapshot (transfer logs: ${LOGDIR})"
for id in "$@"; do
  doctl compute snapshot get "${id}" --format ID,Name,Regions --no-header
done
