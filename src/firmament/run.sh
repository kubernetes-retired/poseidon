#!/bin/bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

export LD_LIBRARY_PATH=${DIR}/../../firmament/src/firmament-build/src

${DIR}/test_integration
