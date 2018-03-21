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
# $2 path to where Kubernetes is installed

export PATH=$PATH:/usr/lib/go-1.6/bin

NUM_HOSTS=$(wc -l $1 | cut -d' ' -f1)
HOSTS=$(cat $1 | tr '\n' ' ')
echo "Please update nodes, roles and NUM_NODES in ${2}/cluster/ubuntu/config-default.sh"
echo "Cluster nodes: ${HOSTS}"
echo "Number nodes: ${NUM_HOSTS}"

read -p "Have you updated ${2}/cluster/ubuntu/config-default.sh? " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]
then
    cd $2/cluster/
    KUBERNETES_PROVIDER=ubuntu ./kube-up.sh
    cd $2/cluster/ubuntu/
    KUBERNETES_PROVIDER=ubuntu ./deployAddons.sh
fi
