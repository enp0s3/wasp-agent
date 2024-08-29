#!/usr/bin/env bash

#Copyright 2023 The WASP Authors.
#
#Licensed under the Apache License, Version 2.0 (the "License");
#you may not use this file except in compliance with the License.
#You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
#Unless required by applicable law or agreed to in writing, software
#distributed under the License is distributed on an "AS IS" BASIS,
#WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#See the License for the specific language governing permissions and
#limitations under the License.

set -eo pipefail

readonly MAX_WASP_WAIT_RETRY=30
readonly WASP_WAIT_TIME=10

script_dir="$(cd "$(dirname "$0")" && pwd -P)"
source hack/build/config.sh
source hack/build/common.sh
source cluster-up/hack/common.sh

KUBEVIRTCI_CONFIG_PATH="$(
    cd "$(dirname "$BASH_SOURCE[0]")/../../"
    echo "$(pwd)/_ci-configs"
)"

# functional testing
BASE_PATH=${KUBEVIRTCI_CONFIG_PATH:-$PWD}
KUBECONFIG=${KUBECONFIG:-$BASE_PATH/$KUBEVIRT_PROVIDER/.kubeconfig}
GOCLI=${GOCLI:-${WASP_DIR}/cluster-up/cli.sh}
KUBE_URL=${KUBE_URL:-""}
WASP_NAMESPACE=${WASP_NAMESPACE:-wasp}


OPERATOR_CONTAINER_IMAGE=$(./cluster-up/kubectl.sh get ds -n $WASP_NAMESPACE wasp-agent -o'custom-columns=spec:spec.template.spec.containers[0].image' --no-headers)
DOCKER_PREFIX=${OPERATOR_CONTAINER_IMAGE%/*}
DOCKER_TAG=${OPERATOR_CONTAINER_IMAGE##*:}

if [ -z "${KUBECTL+x}" ]; then
    kubevirtci_kubectl="${BASE_PATH}/${KUBEVIRT_PROVIDER}/.kubectl"
    if [ -e ${kubevirtci_kubectl} ]; then
        KUBECTL=${kubevirtci_kubectl}
    else
        KUBECTL=$(which kubectl)
    fi
fi

# parsetTestOpts sets 'pkgs' and test_args
parseTestOpts "${@}"

arg_kubeurl="${KUBE_URL:+-kubeurl=$KUBE_URL}"
arg_namespace="${WASP_NAMESPACE:+-wasp-namespace=$WASP_NAMESPACE}"
arg_kubeconfig_wasp="${KUBECONFIG:+-kubeconfig-wasp=$KUBECONFIG}"
arg_kubeconfig="${KUBECONFIG:+-kubeconfig=$KUBECONFIG}"
arg_kubectl="${KUBECTL:+-kubectl-path-wasp=$KUBECTL}"
arg_oc="${KUBECTL:+-oc-path-wasp=$KUBECTL}"
arg_gocli="${GOCLI:+-gocli-path-wasp=$GOCLI}"
arg_docker_prefix="${DOCKER_PREFIX:+-docker-prefix=$DOCKER_PREFIX}"
arg_docker_tag="${DOCKER_TAG:+-docker-tag=$DOCKER_TAG}"

test_args="${test_args}  -ginkgo.v  ${arg_kubeurl} ${arg_namespace} ${arg_kubeconfig} ${arg_kubeconfig_wasp} ${arg_kubectl} ${arg_oc} ${arg_gocli} ${arg_docker_prefix} ${arg_docker_tag}"

echo 'Wait until all wasp Pods are ready'
retry_counter=0
while [ $retry_counter -lt $MAX_WASP_WAIT_RETRY ] && [ -n "$(./cluster-up/kubectl.sh get pods -n $WASP_NAMESPACE -o'custom-columns=status:status.containerStatuses[*].ready' --no-headers | grep false)" ]; do
    retry_counter=$((retry_counter + 1))
    sleep $WASP_WAIT_TIME
    echo "Checking wasp pods again, count $retry_counter"
    if [ $retry_counter -gt 1 ] && [ "$((retry_counter % 6))" -eq 0 ]; then
        ./cluster-up/kubectl.sh get pods -n $WASP_NAMESPACE
    fi
done

if [ $retry_counter -eq $MAX_WASP_WAIT_RETRY ]; then
    echo "Not all wasp pods became ready"
    ./cluster-up/kubectl.sh get pods -n $WASP_NAMESPACE
    ./cluster-up/kubectl.sh get pods -n $WASP_NAMESPACE -o yaml
    ./cluster-up/kubectl.sh describe pods -n $WASP_NAMESPACE
    exit 1
fi

test_command="${TESTS_OUT_DIR}/tests.test -test.timeout 360m ${test_args}"
echo "$test_command"
(
    cd ${WASP_DIR}/tests
    ${test_command}
)
