#!/bin/bash
# Hacky script that deploys Poseidon locally.

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
sudo cp ${DIR}/../build/poseidon /usr/bin/poseidon
sudo cp ${DIR}/../build/third_party/ms-cpprestsdk/src/cpp-rest-sdk-build/Binaries/libcpprest.so.2.7 /usr/lib/
sudo cp ${DIR}/../build/firmament/src/firmament-build/src/libfirmament_scheduling.so /usr/lib/
sudo cp ${DIR}/../build/firmament/src/firmament-build/third_party/cs2/src/cs2/cs2.exe /usr/bin/
