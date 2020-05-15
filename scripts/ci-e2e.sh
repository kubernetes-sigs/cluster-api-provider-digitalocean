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
# usage: ci-e2e.sh
#  This program runs the e2e tests.
################################################################################

set -o nounset
set -o pipefail

REPO_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
cd "${REPO_ROOT}" || exit 1

# shellcheck source=../hack/ensure-go.sh
source "${REPO_ROOT}/hack/ensure-go.sh"
# shellcheck source=../hack/ensure-kind.sh
source "${REPO_ROOT}/hack/ensure-kind.sh"
# shellcheck source=../hack/ensure-kubectl.sh
source "${REPO_ROOT}/hack/ensure-kubectl.sh"
# shellcheck source=../hack/ensure-doctl.sh
source "${REPO_ROOT}/hack/ensure-doctl.sh"

ARTIFACTS="${ARTIFACTS:-${PWD}/_artifacts}"
mkdir -p "${ARTIFACTS}/logs/"

SSH_KEY_NAME=capdo-e2e-test
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

export MACHINE_TYPE=${MACHINE_TYPE:-s-2vcpu-2gb}
export MACHINE_IMAGE=${MACHINE_IMAGE:-63624555} # default is capi do-default image (cluster-api-ubuntu-1804-v1.16.2) from image-builder https://github.com/kubernetes-sigs/image-builder/tree/master/images/capi/packer/digitalocean 
export MACHINE_SSHKEY=${SSH_KEY_FINGERPRINT}

make test-e2e ARTIFACTS=${ARTIFACTS}
test_status="${?}"
