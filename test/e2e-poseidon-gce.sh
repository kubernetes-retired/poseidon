#!/bin/bash

# Copyright 2018 The Kubernetes Authors.
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

set -o errexit
set -o nounset
set -o pipefail

TEST_NAMESPACE="poseidon-test"

SCRIPT_ROOT=$(dirname ${BASH_SOURCE[0]})/..
echo $SCRIPT_ROOT # ../test/..

#Set environment variables
BUILD_VERSION=$(git rev-parse HEAD)
POSEIDON_ROOT_DIR=${SCRIPT_ROOT}
FIRMAMENT_MANIFEST_FILE_PATH=../../deploy/firmament-deployment-e2e.yaml
POSEIDON_MANIFEST_FILE_PATH=../../deploy/poseidon-deployment-e2e.yaml

# Get the compute project
project=$(gcloud info --format='value(config.project)')
if [[ $project == "" ]]; then
  echo "Could not find gcloud project"
  exit 1
fi

# setup gcr project registry 
kube_registry="${KUBE_REGISTRY:-gcr.io/${project}}"


# work from the correct path
cd $(dirname ${BASH_SOURCE[0]})/..
#Create a poseidon release and extract images and packages in the  _output folder 
make quick-release

#Push to Registry
gcloud docker -- load -i _output/release-images/amd64/poseidon-amd64.tar
gcloud docker -- tag "gcr.io/google_containers/poseidon-amd64:${BUILD_VERSION}" "${kube_registry}/poseidon-amd64:${BUILD_VERSION}"
gcloud docker -- push "${kube_registry}/poseidon-amd64:${BUILD_VERSION}"

#Extract Deployment files and place in the deploy folder
tar -xzf _output/release-tars/poseidon-src.tar.gz -C /tmp/
cp /tmp/deploy/*.yaml deploy/.


# setup the env and correct test directory
cd test/e2e

# update the poseidon deployment yaml with the correct image
sed -i "s/gcr.io\/poseidon-173606\/poseidon:latest/gcr.io\/$project\/poseidon-amd64:${BUILD_VERSION}/" $POSEIDON_MANIFEST_FILE_PATH

#Run e2e test
go test -v . -ginkgo.v -args -testNamespace=${TEST_NAMESPACE} -firmamentManifestPath=${FIRMAMENT_MANIFEST_FILE_PATH} -poseidonManifestPath=${POSEIDON_MANIFEST_FILE_PATH}

