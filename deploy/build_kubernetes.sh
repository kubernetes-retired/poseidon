#!/bin/bash
# $1 path to file containing the hosts on which Kubernetes will be deployed.
# $2 path to where Kubernetes will be downloaded

deploydir=$(dirname "$0")

sudo apt-get install golang-1.6-go golang-1.6
export PATH=$PATH:/usr/lib/go-1.6/bin

# Clone Kubernetes repository
git clone --depth 1 --no-single-branch https://github.com/kubernetes/kubernetes.git $2
cd $2
git checkout -b release-1.5 remotes/origin/release-1.5

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
