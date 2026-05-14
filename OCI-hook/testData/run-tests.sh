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
TESTDATA_DIR="${SCRIPT_DIR}"
SCENARIOS_DIR="${TESTDATA_DIR}/scenarios"
CRIO_TESTDATA="${REPO_ROOT}/tools/cmd/oci-hook-render/testdata/crun/crio.conf.d"
TEMPLATE="${REPO_ROOT}/OCI-hook/hookscript.template"
RENDERED_SCRIPT="${TESTDATA_DIR}/rendered-hook.sh"

echo "=== Rendering OCI hook script ==="
go run "${REPO_ROOT}/tools/cmd/oci-hook-render/" \
    -crio-config-dir "${CRIO_TESTDATA}" \
    -template "${TEMPLATE}" \
    -o "${RENDERED_SCRIPT}" 2>/dev/null
chmod +x "${RENDERED_SCRIPT}"
echo "Rendered to ${RENDERED_SCRIPT}"
echo

failures=0
total=0

for scenario_dir in "${SCENARIOS_DIR}"/*/; do
    scenario=$(basename "${scenario_dir}")
    total=$((total + 1))

    expected_exit_code=$(tr -d '[:space:]' < "${scenario_dir}/expected_exit_code")
    use_crun=$(tr -d '[:space:]' < "${scenario_dir}/use_crun")
    crun_bin_dir="${TESTDATA_DIR}/bin/crun-${use_crun}"

    proc_dir="${scenario_dir}/proc"
    if [ ! -d "${proc_dir}" ]; then
        proc_dir=$(mktemp -d)
        trap_cleanup="rm -rf ${proc_dir}"
    else
        trap_cleanup=""
    fi

    echo "--- ${scenario}"
    echo "    crun: ${use_crun}, expected exit: ${expected_exit_code}, PROC_DIR: ${proc_dir}"

    state_file="${scenario_dir}/state.json"
    if [ -f "${state_file}" ]; then
        state_input="${state_file}"
    else
        state_input="/dev/null"
    fi

    set +e
    (
        cd "${scenario_dir}"
        PATH="${crun_bin_dir}:${PATH}" PROC_DIR="${proc_dir}" "${RENDERED_SCRIPT}" < "${state_input}"
    ) > /dev/null 2>&1
    actual_exit_code=$?
    set -e

    [ -n "${trap_cleanup}" ] && eval "${trap_cleanup}"

    if [ "${actual_exit_code}" -eq "${expected_exit_code}" ]; then
        echo "    PASS (exit code: ${actual_exit_code})"
    else
        echo "    FAIL (expected: ${expected_exit_code}, got: ${actual_exit_code})"
        failures=$((failures + 1))
    fi
    echo
done

echo "=== Results: $((total - failures))/${total} passed ==="

if [ "${failures}" -gt 0 ]; then
    echo "FAILED: ${failures} scenario(s) failed"
    exit 1
fi

echo "All scenarios passed"
