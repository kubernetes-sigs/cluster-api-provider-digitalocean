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

# Build Cluster API machine images (DigitalOcean snapshots) with image-builder.
# See docs/images.md for prerequisites and the token scopes it needs.

set -o errexit
set -o nounset
set -o pipefail

usage() {
  cat <<EOF
Usage: $(basename "$0") [options] <kubernetes-version>...

Builds one DigitalOcean snapshot per Kubernetes version (for example 1.35.11) with
kubernetes-sigs/image-builder, and prints the snapshot IDs.

Options:
  -r <region>   Region to build in. The snapshot is stored there. (default: \$REGION or nyc1)
  -o <os>       image-builder DigitalOcean OS target, e.g. ubuntu-2404, ubuntu-2204.
                (default: \$IMAGE_OS or ubuntu-2404)
  -d <dir>      Existing image-builder checkout to use. If not set, image-builder is cloned
                into a temporary directory at \$IMAGE_BUILDER_REF (default: main).
  -h            Show this help.

Environment:
  DIGITALOCEAN_ACCESS_TOKEN   Required. API token of the account the snapshots go to.
  IMAGE_BUILDER_REF           Branch, tag or commit of image-builder to clone (default: main).
  IMAGE_BUILDER_REPO          Repository to clone (default: https://github.com/kubernetes-sigs/image-builder.git).
EOF
}

REGION="${REGION:-nyc1}"
IMAGE_OS="${IMAGE_OS:-ubuntu-2404}"
IMAGE_BUILDER_DIR="${IMAGE_BUILDER_DIR:-}"
IMAGE_BUILDER_REF="${IMAGE_BUILDER_REF:-main}"
IMAGE_BUILDER_REPO="${IMAGE_BUILDER_REPO:-https://github.com/kubernetes-sigs/image-builder.git}"

while getopts ":r:o:d:h" opt; do
  case "${opt}" in
    r) REGION="${OPTARG}" ;;
    o) IMAGE_OS="${OPTARG}" ;;
    d) IMAGE_BUILDER_DIR="${OPTARG}" ;;
    h) usage; exit 0 ;;
    *) usage >&2; exit 2 ;;
  esac
done
shift $((OPTIND - 1))

if [[ $# -eq 0 ]]; then
  usage >&2
  exit 2
fi

for v in "$@"; do
  if [[ ! "${v}" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "invalid Kubernetes version '${v}': use MAJOR.MINOR.PATCH, e.g. 1.35.11" >&2
    exit 2
  fi
done

if [[ -z "${DIGITALOCEAN_ACCESS_TOKEN:-}" ]]; then
  echo "DIGITALOCEAN_ACCESS_TOKEN must be set (see docs/images.md for the scopes it needs)" >&2
  exit 1
fi
export DIGITALOCEAN_ACCESS_TOKEN

for tool in doctl jq git make; do
  command -v "${tool}" >/dev/null || { echo "${tool} is required" >&2; exit 1; }
done

# image-builder's scripts need GNU sed. On macOS, use Homebrew's gnu-sed if it is installed.
if ! sed --version >/dev/null 2>&1; then
  if [[ -d /opt/homebrew/opt/gnu-sed/libexec/gnubin ]]; then
    export PATH="/opt/homebrew/opt/gnu-sed/libexec/gnubin:${PATH}"
  elif [[ -d /usr/local/opt/gnu-sed/libexec/gnubin ]]; then
    export PATH="/usr/local/opt/gnu-sed/libexec/gnubin:${PATH}"
  else
    echo "GNU sed is required (on macOS: brew install gnu-sed)" >&2
    exit 1
  fi
fi

# ansible is installed with `pip --user`; make sure the user scripts directory is on PATH.
if command -v python3 >/dev/null; then
  PYTHON_USER_BIN="$(python3 -c 'import site, os; print(os.path.join(site.getuserbase(), "bin"))')"
  export PATH="${PYTHON_USER_BIN}:${PATH}"
fi

if [[ -z "${IMAGE_BUILDER_DIR}" ]]; then
  IMAGE_BUILDER_DIR="$(mktemp -d)/image-builder"
  echo "Cloning ${IMAGE_BUILDER_REPO} (${IMAGE_BUILDER_REF}) into ${IMAGE_BUILDER_DIR}"
  git clone --quiet "${IMAGE_BUILDER_REPO}" "${IMAGE_BUILDER_DIR}"
  git -C "${IMAGE_BUILDER_DIR}" checkout --quiet "${IMAGE_BUILDER_REF}"
fi

CAPI_DIR="${IMAGE_BUILDER_DIR}/images/capi"
OS_VARS="${CAPI_DIR}/packer/digitalocean/${IMAGE_OS}.json"
if [[ ! -f "${OS_VARS}" ]]; then
  echo "unknown OS target '${IMAGE_OS}': ${OS_VARS} does not exist" >&2
  exit 1
fi
NAME_SUFFIX="$(jq -r '.snapshot_name_suffix' "${OS_VARS}")"

export PATH="${CAPI_DIR}/.local/bin:${PATH}"
make -C "${CAPI_DIR}" deps-do

results=()
for v in "$@"; do
  series="v${v%.*}"
  echo "=== Building Kubernetes v${v} (${IMAGE_OS}) in ${REGION}"
  # pkgs.k8s.io package versions are <version>-1.1.
  PACKER_FLAGS="--var region=${REGION} --var kubernetes_semver=v${v} --var kubernetes_series=${series} --var kubernetes_deb_version=${v}-1.1 --var kubernetes_rpm_version=${v}" \
    make -C "${CAPI_DIR}" "build-do-${IMAGE_OS}"

  name="Cluster API Kubernetes v${v} ${NAME_SUFFIX}"
  id="$(doctl compute snapshot list --resource droplet --output json \
    | jq -r --arg name "${name}" '[.[] | select(.name == $name)] | sort_by(.created_at) | last | .id // empty')"
  if [[ -z "${id}" ]]; then
    echo "build finished but no snapshot named '${name}' was found" >&2
    exit 1
  fi
  results+=("${id}  ${name}")
done

echo
echo "=== Snapshots"
printf '%s\n' "${results[@]}"
echo
echo "To copy them to other regions: hack/images/transfer-do-images.sh <id>..."
