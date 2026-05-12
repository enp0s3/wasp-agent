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

set -euo pipefail

source hack/build/config.sh
source hack/build/common.sh

hack/build/build-ginkgo.sh
GINKGO="${BIN_DIR}/ginkgo"

# parsetTestOpts sets 'pkgs' and test_args
parseTestOpts "${@}"
export KUBEBUILDER_CONTROLPLANE_START_TIMEOUT=120s

# ginkgo requires directory paths, not import paths
ginkgo_dirs=$(go list -f '{{.Dir}}' ${pkgs} 2>/dev/null)

test_command="env OPERATOR_DIR=${WASP_DIR} ${GINKGO} -v -coverprofile=.coverprofile ${ginkgo_dirs} ${test_args:+-args $test_args}"
echo "${test_command}"
${test_command}

echo ""
echo "=== Running OCI hook script tests ==="
"${WASP_DIR}/OCI-hook/testData/run-tests.sh"
