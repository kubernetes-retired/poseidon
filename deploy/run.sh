#!/bin/bash
DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

cp ${DIR}/../build/poseidon poseidon
cp ${DIR}/../build/third_party/ms-cpprestsdk/src/cpp-rest-sdk-build/Binaries/libcpprest.so.2.7 libcpprest.so.2.7
cp ${DIR}/../build/firmament/src/firmament-build/src/libfirmament_scheduling.so libfirmament_scheduling.so
cp ${DIR}/../build/firmament/src/firmament-build/third_party/cs2/src/cs2/cs2.exe cs2.exe

docker build -t camsas/poseidon:dev ${DIR}
