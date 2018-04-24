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

# Use the 0.0 tag for testing, it shouldn't clobber any release builds
TAG?=0.1.0
REGISTRY?=k8s.gcr.io
GOOS?=linux
DOCKER?=docker

ARCH ?= amd64
GOARCH = ${ARCH}

ALL_ARCH = amd64 # TODO: arm arm64 ppc64le s390x

IMGNAME = poseidon
IMAGE = $(REGISTRY)/$(IMGNAME)
MULTI_ARCH_IMG = $(IMAGE)-$(ARCH)

TEMP_DIR := $(shell mktemp -d)

.PHONY: container
container: .container-$(ARCH)

.PHONY: .container-$(ARCH)
.container-$(ARCH): clean build
	$(DOCKER) build -t $(MULTI_ARCH_IMG):$(TAG) $(TEMP_DIR)/deploy

.PHONY: push
push: .push-$(ARCH)

.PHONY: .push-$(ARCH)
.push-$(ARCH):
	$(DOCKER) push $(MULTI_ARCH_IMG):$(TAG)

.PHONY: clean
clean:
	$(DOCKER) rmi -f $(MULTI_ARCH_IMG):$(TAG) || true

.PHONY: build
build:	
	cp -RP ./* $(TEMP_DIR)
	CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} go build -a -installsuffix cgo \
		-o ${TEMP_DIR}/deploy/poseidon github.com/kubernetes-sigs/poseidon/cmd/poseidon

.PHONY: release
release: container push
	echo "done"
