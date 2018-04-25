#!/usr/bin/env bash
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

set -euo pipefail

# The script will try to checkout the release version if specified.
# This is useful for manual or older tagged version releases.
# If a release version is not specified, it will check the latest commit
# for a tag, if a tag has appeared on the latest commit, it will build
# and push the binaries from the latest commit.
RELEASE_VERSION=${1:-}
RELEASE_TAG=${RELEASE_VERSION:-}
GCP_PROJECT=${GCP_PROJECT:-k8s-jkns-e2e-gce-federation}
GCS_BUCKET=${GCS_BUCKET:-kubernetes-federation-release}
GCR_REPO_PATH=${GCR_REPO_PATH:-"k8s-jkns-e2e-gce-federation/poseidon-amd64"}
TMPDIR="$(mktemp -d /tmp/k8s-fed-relXXXXXXXX)"
KUBE_ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )/.."

source "${KUBE_ROOT}/hack/lib/version.sh"

function clean_up() {
  if [[ "${TMPDIR}" == "/tmp/k8s-fed-rel"* ]]; then
    rm -rf "${TMPDIR}"
  fi
}
trap clean_up EXIT

pushd "${KUBE_ROOT}" 1>&2

if [[ -z ${RELEASE_VERSION} ]]; then
  echo "No release version specified, will check if the latest commit has a tag."
  tags=$(exec 2>&1; git describe --exact-match) >/dev/null 2>&1 || { echo >&2 "No tag found on latest commit; No release to be made."; exit 0; }
  # Just use the first one in the list if the commit has multiple tags.
  RELEASE_TAG=${tags[0]}
else
  git checkout ${RELEASE_VERSION}
fi

# Load the version vars.
kube::version::get_version_vars

# This is to prevent the script from starting if its on a dirty commit.
if [[ "${KUBE_GIT_VERSION}" != "${RELEASE_TAG}" ]]; then
  echo "Version being build: ${KUBE_GIT_VERSION} does not match the release version: ${RELEASE_TAG}, there probably is uncommited work."
  exit 1
else
  echo "Using ${RELEASE_TAG} for release push"
fi

# Check for and install (some) necessary dependencies.
command -v gcloud >/dev/null 2>&1 || { echo >&2 "Please install gcloud before running this script."; exit 1; }
gcloud components install gsutil docker-credential-gcr
docker-credential-gcr configure-docker 1>&2

RELEASE_TMPDIR="${TMPDIR}/${RELEASE_TAG}"

# Build the tarballs.
make bazel-release

# Copy the archives.
mkdir -p "${RELEASE_TMPDIR}"
cp \
  bazel-bin/build/release-tars/poseidon-server-linux-amd64.tar.gz \
  "${RELEASE_TMPDIR}"

# Create the `latest` file.
# We consider a version built and pushed in chronology as latest
echo "${RELEASE_TAG}" > "${TMPDIR}/latest"

pushd "${RELEASE_TMPDIR}" 1>&2

# Create the SHAs.
sha256sum poseidon-server-linux-amd64.tar.gz > poseidon-server-linux-amd64.tar.gz.sha

popd 1>&2

# Upload the files to GCS.
gsutil -m cp -r "${RELEASE_TMPDIR}" "gs://${GCS_BUCKET}/release"

# Push the server container image.
bazel run //build:push-poseidon-image --define repository="${GCR_REPO_PATH}" 1>&2

# Adjust the tags on the image. The `push-fcp-image` rule tags the
# pushed image with the `dev` tag by default. This consistent tag allows the
# tool to easily add other tags to the image. The tool then removes the `dev`
# tag since this is not a development image.
gcloud container images add-tag --quiet \
  "gcr.io/${GCR_REPO_PATH}:dev" \
  "gcr.io/${GCR_REPO_PATH}:${RELEASE_TAG}"
# We consider a version built and pushed in chronology as latest
gcloud container images add-tag --quiet \
  "gcr.io/${GCR_REPO_PATH}:dev" \
  "gcr.io/${GCR_REPO_PATH}:latest"
gcloud container images untag --quiet \
  "gcr.io/${GCR_REPO_PATH}:dev"

echo "Pushing the release completed for poseidon ${RELEASE_TAG} release"

