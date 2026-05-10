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

source hack/build/common.sh

GINKGO_BIN="${BIN_DIR}/ginkgo"

if [ -f "${GINKGO_BIN}" ]; then
    echo "ginkgo binary already exists at ${GINKGO_BIN}, skipping build"
    exit 0
fi

mkdir -p "${BIN_DIR}"
echo "Building ginkgo from vendored source..."
go build -o "${GINKGO_BIN}" ./vendor/github.com/onsi/ginkgo/v2/ginkgo
echo "ginkgo binary built at ${GINKGO_BIN}"
