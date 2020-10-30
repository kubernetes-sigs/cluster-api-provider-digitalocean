#!/bin/bash

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

################################################################################
# usage: ci-conformance.sh
#  This program runs the conformance tests.
################################################################################

set -o nounset
set -o pipefail

export PATH=${PWD}/hack/tools/bin:${PATH}
REPO_ROOT=$(git rev-parse --show-toplevel)

# shellcheck source=../hack/ensure-go.sh
source "${REPO_ROOT}/hack/ensure-go.sh"
# shellcheck source=../hack/ensure-doctl.sh
source "${REPO_ROOT}/hack/ensure-doctl.sh"

export ARTIFACTS="${ARTIFACTS:-${REPO_ROOT}/_artifacts}"
export E2E_CONF_FILE="${REPO_ROOT}/test/e2e/config/digitalocean-ci.yaml"

SSH_KEY_NAME=capdo-conf-$(head /dev/urandom | tr -dc A-Za-z0-9 | head -c 12 ; echo '')
SSH_KEY_PATH=/tmp/${SSH_KEY_NAME}
create_ssh_key() {
    echo "generating new ssh key"
    ssh-keygen -t rsa -f ${SSH_KEY_PATH} -N '' 2>/dev/null <<< y >/dev/null
    echo "importing ssh key "
    doctl compute ssh-key import ${SSH_KEY_NAME} --public-key-file ${SSH_KEY_PATH}.pub
}

remove_ssh_key() {
    local ssh_fingerprint=$1
    echo "removing ssh key"
    doctl compute ssh-key delete ${ssh_fingerprint} --force
    rm -f ${SSH_KEY_PATH}
}

create_ssh_key
SSH_KEY_FINGERPRINT=$(ssh-keygen -E md5 -lf "${SSH_KEY_PATH}" | awk '{ print $2 }' | cut -c 5-)
trap 'remove_ssh_key ${SSH_KEY_FINGERPRINT}' EXIT

export DO_SSH_KEY_FINGERPRINT=${SSH_KEY_FINGERPRINT}

export GINKGO_FOCUS="Conformance Tests"
make test-conformance
test_status="${?}"
