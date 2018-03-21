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

#!/bin/bash
# $1 path to file containing the hosts on which Kubernetes will be deployed.
# $2 path to where Kubernetes will be downloaded

deploydir=$(dirname "$0")

sudo apt-get install golang-1.6-go golang-1.6 pssh
export PATH=$PATH:/usr/lib/go-1.6/bin

# Clone Kubernetes repository
git clone --depth 1 --no-single-branch -b release-1.5 https://github.com/kubernetes/kubernetes.git

# Apply patches
cd $2
git apply $deploydir/k8s_build_only_amd64.diff

# Build Kubernetes
cd $2
make
sudo cp $2/_output/local/bin/linux/amd64/kubectl /usr/local/bin/kubectl
sudo chmod +x /usr/local/bin/kubectl

# Install required packages on all machines
parallel-ssh -i -h $1 sudo apt-get -y install bridge-utils
