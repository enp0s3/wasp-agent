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

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd -P)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd -P)"

TESTDATA_DIR="${REPO_ROOT}/tools/cmd/oci-hook-render/testdata"
TEMPLATE="${REPO_ROOT}/OCI-hook/hookscript.template"
OUT_DIR=$(mktemp -d)
trap 'rm -rf "${OUT_DIR}"' EXIT

if ! command -v shellcheck &>/dev/null; then
    echo "Error: shellcheck is not installed"
    echo "Install it with: dnf install ShellCheck / apt install shellcheck / brew install shellcheck"
    exit 1
fi

failures=0

for config_dir in "${TESTDATA_DIR}"/*/crio.conf.d; do
    fixture=$(basename "$(dirname "${config_dir}")")
    output="${OUT_DIR}/hook-${fixture}.sh"

    echo "--- Rendering hook script for fixture: ${fixture}"
    go run "${REPO_ROOT}/tools/cmd/oci-hook-render/" \
        -crio-config-dir "${config_dir}" \
        -template "${TEMPLATE}" \
        -o "${output}" 2>/dev/null

    echo "--- Running shellcheck on: ${fixture}"
    if shellcheck -x "${output}"; then
        echo "    PASS: ${fixture}"
    else
        echo "    FAIL: ${fixture}"
        failures=$((failures + 1))
    fi
    echo
done

if [ "${failures}" -gt 0 ]; then
    echo "shellcheck failed for ${failures} fixture(s)"
    exit 1
fi

echo "All hook scripts passed shellcheck"
