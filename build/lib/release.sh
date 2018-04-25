#!/bin/bash

# Copyright 2016 The Kubernetes Authors.
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

# This file creates release artifacts (tar files, container images) that are
# ready to distribute to install or distribute to end users.

###############################################################################
# Most of the ::release:: namespace functions have been moved to
# github.com/kubernetes/release.  Have a look in that repo and specifically in
# lib/releaselib.sh for ::release::-related functionality.
###############################################################################

# This is where the final release artifacts are created locally
readonly RELEASE_STAGE="${LOCAL_OUTPUT_ROOT}/release-stage"
readonly RELEASE_TARS="${LOCAL_OUTPUT_ROOT}/release-tars"
readonly RELEASE_IMAGES="${LOCAL_OUTPUT_ROOT}/release-images"

KUBE_BUILD_HYPERKUBE=${KUBE_BUILD_HYPERKUBE:-n}
if [[ -n "${KUBE_DOCKER_IMAGE_TAG-}" && -n "${KUBE_DOCKER_REGISTRY-}" ]]; then
  # retain legacy behavior of automatically building hyperkube during releases
  KUBE_BUILD_HYPERKUBE=y
fi

# Validate a ci version
#
# Globals:
#   None
# Arguments:
#   version
# Returns:
#   If version is a valid ci version
# Sets:                    (e.g. for '1.2.3-alpha.4.56+abcdef12345678')
#   VERSION_MAJOR          (e.g. '1')
#   VERSION_MINOR          (e.g. '2')
#   VERSION_PATCH          (e.g. '3')
#   VERSION_PRERELEASE     (e.g. 'alpha')
#   VERSION_PRERELEASE_REV (e.g. '4')
#   VERSION_BUILD_INFO     (e.g. '.56+abcdef12345678')
#   VERSION_COMMITS        (e.g. '56')
function kube::release::parse_and_validate_ci_version() {
  # Accept things like "v1.2.3-alpha.4.56+abcdef12345678" or "v1.2.3-beta.4"
  local -r version_regex="^v(0|[1-9][0-9]*)\\.(0|[1-9][0-9]*)\\.(0|[1-9][0-9]*)-([a-zA-Z0-9]+)\\.(0|[1-9][0-9]*)(\\.(0|[1-9][0-9]*)\\+[0-9a-f]{7,40})?$"
  local -r version="${1-}"
  [[ "${version}" =~ ${version_regex} ]] || {
    kube::log::error "Invalid ci version: '${version}', must match regex ${version_regex}"
    return 1
  }
  VERSION_MAJOR="${BASH_REMATCH[1]}"
  VERSION_MINOR="${BASH_REMATCH[2]}"
  VERSION_PATCH="${BASH_REMATCH[3]}"
  VERSION_PRERELEASE="${BASH_REMATCH[4]}"
  VERSION_PRERELEASE_REV="${BASH_REMATCH[5]}"
  VERSION_BUILD_INFO="${BASH_REMATCH[6]}"
  VERSION_COMMITS="${BASH_REMATCH[7]}"
}

# ---------------------------------------------------------------------------
# Build final release artifacts
function kube::release::clean_cruft() {
  # Clean out cruft
  find ${RELEASE_STAGE} -name '*~' -exec rm {} \;
  find ${RELEASE_STAGE} -name '#*#' -exec rm {} \;
  find ${RELEASE_STAGE} -name '.DS*' -exec rm {} \;
}

function kube::release::package_tarballs() {
  # Clean out any old releases
  rm -rf "${RELEASE_STAGE}" "${RELEASE_TARS}" "${RELEASE_IMAGES}"
  mkdir -p "${RELEASE_TARS}"

  kube::release::package_src_tarball &
  kube::util::wait-for-jobs || { kube::log::error "previous tarball phase failed"; return 1; }

  kube::release::package_server_tarballs &
  kube::util::wait-for-jobs || { kube::log::error "previous tarball phase failed"; return 1; }

  kube::release::package_final_tarball & # _final depends on some of the previous phases
  kube::release::package_test_tarball & # _test doesn't depend on anything
  kube::log::status "Completed poseidon release build."
}

# Package the source code we built, for compliance/licensing/audit/yadda.
function kube::release::package_src_tarball() {
  kube::log::status "Building tarball: src"
  local source_files=(
    $(cd "${KUBE_ROOT}" && find . -mindepth 1 -maxdepth 1 \
      -not \( \
        \( -path ./_\*        -o \
           -path ./.git\*     -o \
           -path ./.config\* -o \
           -path ./.gsutil\*    \
        \) -prune \
      \))
  )
  "${TAR}" czf "${RELEASE_TARS}/poseidon-src.tar.gz" -C "${KUBE_ROOT}" "${source_files[@]}"
}

# Package up all of the server binaries
function kube::release::package_server_tarballs() {
  local platform
  for platform in "${KUBE_SERVER_PLATFORMS[@]}"; do
    local platform_tag=${platform/\//-} # Replace a "/" for a "-"
    local arch=$(basename ${platform})
    local images_dir="${RELEASE_IMAGES}/${arch}"
    mkdir -p "${images_dir}"

    local -r docker_registry="gcr.io/google_containers"
    # Clean out any old releases
    kube::log::status "Building tarball: server $platform_tag"

    # Docker tags cannot contain '+'
    local docker_tag=$(git rev-parse HEAD)
    if [[ -z "${docker_tag}" ]]; then
      kube::log::error "git version information missing; cannot create Docker tag"
      return 1
    fi

    fed::release::build_poseidon_image "${arch}" "${docker_registry}" \
        "${docker_tag}" "${images_dir}"

    local release_stage="${RELEASE_STAGE}/server/${platform_tag}/poseidon"
    rm -rf "${release_stage}"
    mkdir -p "${release_stage}/server/bin"

    # This fancy expression will expand to prepend a path
    # (${LOCAL_OUTPUT_BINPATH}/${platform}/) to every item in the
    # KUBE_SERVER_BINARIES array.
    cp "${KUBE_SERVER_BINARIES[@]/#/${LOCAL_OUTPUT_BINPATH}/${platform}/}" \
      "${release_stage}/server/bin/"

    #TODO (irfan) Do we need this license file and any other src file/script in release.
    #cp "${KUBE_ROOT}/Godeps/LICENSES" "${release_stage}/"
    cp "${RELEASE_TARS}/poseidon-src.tar.gz" "${release_stage}/"

    #copy out fcp image
    cp "${images_dir}/poseidon-${arch}.tar" "${release_stage}/"

    kube::release::clean_cruft

    local package_name="${RELEASE_TARS}/poseidon-server-${platform_tag}.tar.gz"
    kube::release::create_tarball "${package_name}" "${release_stage}/.."
  done
}

function kube::release::md5() {
  if which md5 >/dev/null 2>&1; then
    md5 -q "$1"
  else
    md5sum "$1" | awk '{ print $1 }'
  fi
}

function kube::release::sha1() {
  if which sha1sum >/dev/null 2>&1; then
    sha1sum "$1" | awk '{ print $1 }'
  else
    shasum -a1 "$1" | awk '{ print $1 }'
  fi
}

function kube::release::build_hyperkube_image() {
  local -r arch="$1"
  local -r registry="$2"
  local -r version="$3"
  local -r save_dir="${4-}"
  kube::log::status "Building hyperkube image for arch: ${arch}"
  ARCH="${arch}" REGISTRY="${registry}" VERSION="${version}" \
    make -C cluster/images/hyperkube/ build >/dev/null

  local hyperkube_tag="${registry}/hyperkube-${arch}:${version}"
  if [[ -n "${save_dir}" ]]; then
    "${DOCKER[@]}" save "${hyperkube_tag}" > "${save_dir}/hyperkube-${arch}.tar"
  fi
  if [[ -z "${KUBE_DOCKER_IMAGE_TAG-}" || -z "${KUBE_DOCKER_REGISTRY-}" ]]; then
    # not a release
    kube::log::status "Deleting hyperkube image ${hyperkube_tag}"
    "${DOCKER[@]}" rmi "${hyperkube_tag}" &>/dev/null || true
  fi
}

function fed::release::build_poseidon_image() {
  local -r arch="$1"
  local -r registry="$2"
  local -r version="$3"
  local -r save_dir="${4-}"
  kube::log::status "Building poseidon image for arch: ${arch}"
  ARCH="${arch}" REGISTRY="${registry}" VERSION="${version}" \
    make -C image build >/dev/null

  local poseidon_tag="${registry}/poseidon-${arch}:${version}"
  if [[ -n "${save_dir}" ]]; then
    "${DOCKER[@]}" save "${poseidon_tag}" > "${save_dir}/poseidon-${arch}.tar"
  fi
  if [[ -z "${KUBE_DOCKER_IMAGE_TAG-}" || -z "${KUBE_DOCKER_REGISTRY-}" ]]; then
    # not a release
    kube::log::status "Deleting poseidon image ${poseidon_tag}"
    "${DOCKER[@]}" rmi "${poseidon_tag}" &>/dev/null || true
  fi
}

# This will take binaries that run on master and creates Docker images
# that wrap the binary in them. (One docker image per binary)
# Args:
#  $1 - binary_dir, the directory to save the tared images to.
#  $2 - arch, architecture for which we are building docker images.
function kube::release::create_docker_images_for_server() {
  # Create a sub-shell so that we don't pollute the outer environment
  (
    local binary_dir="$1"
    local arch="$2"
    local binary_name
    local binaries=($(kube::build::get_docker_wrapped_binaries ${arch}))
    local images_dir="${RELEASE_IMAGES}/${arch}"
    mkdir -p "${images_dir}"

    local -r docker_registry="gcr.io/google_containers"
    # Docker tags cannot contain '+'
    local docker_tag=$(git rev-parse HEAD)
    if [[ -z "${docker_tag}" ]]; then
      kube::log::error "git version information missing; cannot create Docker tag"
      return 1
    fi

    for wrappable in "${binaries[@]}"; do

      local oldifs=$IFS
      IFS=","
      set $wrappable
      IFS=$oldifs

      local binary_name="$1"
      local base_image="$2"
      local docker_build_path="${binary_dir}/${binary_name}.dockerbuild"
      local docker_file_path="${docker_build_path}/Dockerfile"
      local binary_file_path="${binary_dir}/${binary_name}"
      local docker_image_tag="${docker_registry}"
      if [[ ${arch} == "amd64" ]]; then
        # If we are building a amd64 docker image, preserve the original
        # image name
        docker_image_tag+="/${binary_name}:${docker_tag}"
      else
        # If we are building a docker image for another architecture,
        # append the arch in the image tag
        docker_image_tag+="/${binary_name}-${arch}:${docker_tag}"
      fi


      kube::log::status "Starting docker build for image: ${binary_name}-${arch}"
      (
        rm -rf ${docker_build_path}
        mkdir -p ${docker_build_path}
        ln ${binary_dir}/${binary_name} ${docker_build_path}/${binary_name}
        printf " FROM ${base_image} \n ADD ${binary_name} /usr/local/bin/${binary_name}\n" > ${docker_file_path}

        "${DOCKER[@]}" build --pull -q -t "${docker_image_tag}" ${docker_build_path} >/dev/null
        "${DOCKER[@]}" save "${docker_image_tag}" > "${binary_dir}/${binary_name}.tar"
        echo "${docker_tag}" > ${binary_dir}/${binary_name}.docker_tag
        rm -rf ${docker_build_path}
        ln "${binary_dir}/${binary_name}.tar" "${images_dir}/"

        # If we are building an official/alpha/beta release we want to keep
        # docker images and tag them appropriately.
        if [[ -n "${KUBE_DOCKER_IMAGE_TAG-}" && -n "${KUBE_DOCKER_REGISTRY-}" ]]; then
          local release_docker_image_tag="${KUBE_DOCKER_REGISTRY}/${binary_name}-${arch}:${KUBE_DOCKER_IMAGE_TAG}"
          # Only rmi and tag if name is different
          if [[ $docker_image_tag != $release_docker_image_tag ]]; then
            kube::log::status "Tagging docker image ${docker_image_tag} as ${release_docker_image_tag}"
            "${DOCKER[@]}" rmi "${release_docker_image_tag}" 2>/dev/null || true
            "${DOCKER[@]}" tag "${docker_image_tag}" "${release_docker_image_tag}" 2>/dev/null
          fi
        else
          # not a release
          kube::log::status "Deleting docker image ${docker_image_tag}"
          "${DOCKER[@]}" rmi ${docker_image_tag} &>/dev/null || true
        fi
      ) &
    done

    if [[ "${KUBE_BUILD_HYPERKUBE}" =~ [yY] ]]; then
      kube::release::build_hyperkube_image "${arch}" "${docker_registry}" \
        "${docker_tag}" "${images_dir}" &
    fi

    kube::util::wait-for-jobs || { kube::log::error "previous Docker build failed"; return 1; }
    kube::log::status "Docker builds done"
  )

}

# This is the stuff you need to run tests from the binary distribution.
function kube::release::package_test_tarball() {
  kube::log::status "Building tarball: test"

  local release_stage="${RELEASE_STAGE}/test/poseidon"
  rm -rf "${release_stage}"
  mkdir -p "${release_stage}"

  local platform
  for platform in "${KUBE_TEST_PLATFORMS[@]}"; do
    local test_bins=("${KUBE_TEST_BINARIES[@]}")
    if [[ "${platform%/*}" == "windows" ]]; then
      test_bins=("${KUBE_TEST_BINARIES_WIN[@]}")
    fi
    mkdir -p "${release_stage}/platforms/${platform}"
    cp "${test_bins[@]/#/${LOCAL_OUTPUT_BINPATH}/${platform}/}" \
      "${release_stage}/platforms/${platform}"
  done
  for platform in "${KUBE_TEST_SERVER_PLATFORMS[@]}"; do
    mkdir -p "${release_stage}/platforms/${platform}"
    cp "${KUBE_TEST_SERVER_BINARIES[@]/#/${LOCAL_OUTPUT_BINPATH}/${platform}/}" \
      "${release_stage}/platforms/${platform}"
  done

  tar c ${KUBE_TEST_PORTABLE[@]} | tar x -C ${release_stage}

  kube::release::clean_cruft

  local package_name="${RELEASE_TARS}/poseidon-test.tar.gz"
  kube::release::create_tarball "${package_name}" "${release_stage}/.."
}

# This is all the platform-independent stuff you need to run/install kubernetes.
# Arch-specific binaries will need to be downloaded separately (possibly by
# using the bundled cluster/get-kube-binaries.sh script).
# Included in this tarball:
#   - Cluster spin up/down scripts and configs for various cloud providers
#     to master by whatever means appropriate.
#   - Examples (which may or may not still work)
#   - The remnants of the docs/ directory
function kube::release::package_final_tarball() {
  kube::log::status "Building tarball: final"

  local release_stage="${RELEASE_STAGE}/poseidon"
  rm -rf "${release_stage}"
  mkdir -p "${release_stage}"

  # Copy the last know good k8s version (which carries a proper tag)
  cp "${KUBE_ROOT}/k8s-version" "${release_stage}/"

  # Include hack/lib as a dependency for the cluster/ scripts
  mkdir -p "${release_stage}/hack"
  cp -R "${KUBE_ROOT}/hack/lib" "${release_stage}/hack/"
  echo "${KUBE_GIT_VERSION}" > "${release_stage}/version"

  kube::release::clean_cruft

  local package_name="${RELEASE_TARS}/poseidon.tar.gz"
  kube::release::create_tarball "${package_name}" "${release_stage}/.."
}

# Build a release tarball.  $1 is the output tar name.  $2 is the base directory
# of the files to be packaged.  This assumes that ${2}/kubernetes is what is
# being packaged.
function kube::release::create_tarball() {
  kube::build::ensure_tar

  local tarfile=$1
  local stagingdir=$2

  "${TAR}" czf "${tarfile}" -C "${stagingdir}" poseidon --owner=0 --group=0
}
