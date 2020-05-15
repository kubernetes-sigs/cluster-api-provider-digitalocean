#!/usr/bin/env bash

# Copyright 2020 The Kubernetes Authors.
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

set -o errexit
set -o nounset
set -o pipefail

GOPATH_BIN="$(go env GOPATH)/bin"
MINIMUM_DOCTL_VERSION=1.43.0

# Ensure the doctl tool exists and is a viable version, or installs it
verify_doctl_version() {

  # If doctl is not available on the path, get it
  if ! [ -x "$(command -v doctl)" ]; then
    if [[ "${OSTYPE}" == "linux-gnu" ]]; then
      echo 'doctl not found, installing'
      if ! [ -d "${GOPATH_BIN}" ]; then
        mkdir -p "${GOPATH_BIN}"
      fi
      curl -sL https://github.com/digitalocean/doctl/releases/download/v${MINIMUM_DOCTL_VERSION}/doctl-${MINIMUM_DOCTL_VERSION}-linux-amd64.tar.gz | tar -C "${GOPATH_BIN}" -xzv
      chmod +x "${GOPATH_BIN}/doctl"
    else
      echo "Missing required binary in path: doctl"
      return 2
    fi
  fi

  local doctl_version
  # Format is 'doctl v0.6.1 go1.13.4 darwin/amd64'
  doctl_version=$(doctl version | grep -E -o '[0-9]+\.[0-9]+\.[0-9]+')
  if [[ "${MINIMUM_DOCTL_VERSION}" != $(echo -e "${MINIMUM_DOCTL_VERSION}\n${doctl_version}" | sort -s -t. -k 1,1 -k 2,2n -k 3,3n | head -n1) ]]; then
    cat <<EOF
Detected doctl version: ${doctl_version}.
Requires ${MINIMUM_DOCTL_VERSION} or greater.
Please install ${MINIMUM_DOCTL_VERSION} or later.
EOF
    return 2
  fi
}

verify_doctl_version
