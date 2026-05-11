#!/usr/bin/env bash

set -e

source hack/build/common.sh

hack/build/build-ginkgo.sh
GINKGO="${BIN_DIR}/ginkgo"

# Find every folder containing tests
for dir in $(find ${WASP_DIR}/pkg/ ${WASP_DIR}/cmd/ -type f -name '*_test.go' -printf '%h\n' | sort -u); do
    # If there is no file ending with _suite_test.go, bootstrap ginkgo
    SUITE_FILE=$(find $dir -maxdepth 1 -type f -name '*_suite_test.go')
    if [ -z "$SUITE_FILE" ]; then
        echo "Missing test suite entrypoint in ${dir}; creating one automatically"
        (cd $dir && ${GINKGO} bootstrap)
    fi
done
