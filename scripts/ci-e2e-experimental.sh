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

export PATH=${PWD}/hack/tools/bin:${PATH}
REPO_ROOT=$(git rev-parse --show-toplevel)

# shellcheck source=../hack/ensure-go.sh
source "${REPO_ROOT}/hack/ensure-go.sh"
# shellcheck source=../hack/ensure-doctl.sh
source "${REPO_ROOT}/hack/ensure-doctl.sh"

export ARTIFACTS="${ARTIFACTS:-${REPO_ROOT}/_artifacts}"
export E2E_CONF_FILE="${REPO_ROOT}/test/e2e/config/digitalocean-ci.yaml"

SSH_KEY_NAME=capdo-e2e-$(head /dev/urandom | tr -dc A-Za-z0-9 | head -c 12 ; echo '')
SSH_KEY_PATH=/tmp/${SSH_KEY_NAME}

# our exit handler (trap)
cleanup() {
  remove_ssh_key "${SSH_KEY_FINGERPRINT}"
  # stop boskos heartbeat
  [[ -z ${HEART_BEAT_PID:-} ]] || kill -9 "${HEART_BEAT_PID}"
}
trap cleanup EXIT

create_ssh_key() {
    echo "generating new ssh key"
    ssh-keygen -t rsa -f "${SSH_KEY_PATH}" -N '' 2>/dev/null <<< y >/dev/null
    echo "importing ssh key "
    doctl compute ssh-key import "${SSH_KEY_NAME}" --public-key-file "${SSH_KEY_PATH}.pub"
}

remove_ssh_key() {
    local ssh_fingerprint=$1
    echo "removing ssh key"
    doctl compute ssh-key delete "${ssh_fingerprint}" --force
    rm -f "${SSH_KEY_PATH}"

    "${REPO_ROOT}"/hack/log/redact.sh || true
}

# If BOSKOS_HOST is set then acquire an GCP account from Boskos.
if [ -n "${BOSKOS_HOST:-}" ]; then
  # Check out the account from Boskos and store the produced environment
  # variables in a temporary file.
  account_env_var_file="$(mktemp)"
  python hack/checkout_account.py 1>"${account_env_var_file}"
  checkout_account_status="${?}"

  # If the checkout process was a success then load the account's
  # environment variables into this process.
  # shellcheck disable=SC1090
  [ "${checkout_account_status}" = "0" ] && . "${account_env_var_file}"

  # Always remove the account environment variable file. It contains
  # sensitive information.
  rm -f "${account_env_var_file}"

  if [ ! "${checkout_account_status}" = "0" ]; then
    echo "error getting account from boskos" 1>&2
    exit "${checkout_account_status}"
  fi

  # run the heart beat process to tell boskos that we are still
  # using the checked out account periodically
  python -u hack/heartbeat_account.py >> "$ARTIFACTS/logs/boskos.log" 2>&1 &
  HEART_BEAT_PID=$(echo $!)
fi

create_ssh_key
SSH_KEY_FINGERPRINT=$(ssh-keygen -E md5 -lf "${SSH_KEY_PATH}" | awk '{ print $2 }' | cut -c 5-)

export DO_SSH_KEY_FINGERPRINT=${SSH_KEY_FINGERPRINT}

defaultTag=$(date -u '+%Y%m%d%H%M%S')
export TAG="${defaultTag:-dev}"

gcloud auth configure-docker

make test-e2e
test_status="${?}"

# If Boskos is being used then release the GCP project back to Boskos.
[ -z "${BOSKOS_HOST:-}" ] || hack/checkin_account.py >> $ARTIFACTS/logs/boskos.log 2>&1

exit "${test_status}"
