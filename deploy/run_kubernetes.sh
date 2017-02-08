#!/bin/bash
# $1 path to file containing the hosts on which Kubernetes will be deployed.
# $2 path to where Kubernetes is installed

export PATH=$PATH:/usr/lib/go-1.6/bin

NUM_HOSTS=`wc -l $1 | cut -d' ' -f1`
HOSTS=`cat $1 | tr '\n' ' '`
echo "Please update nodes, roles and NUM_NODES in ${2}/cluster/ubuntu/config-default.sh"
echo "Cluster nodes: ${HOSTS}"
echo "Number nodes: ${NUM_HOSTS}"

# # Update  nodes="vcap@10.10.103.250 vcap@10.10.103.162 vcap@10.10.103.223"
# # role="ai i i" and NUM_NODES=${NUM_NODES:-3} in cluster/ubuntu/config-default.sh

read -p "Have you updated ${2}/cluster/ubuntu/config-default.sh? " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]
then
    cd $2/cluster/
    KUBERNETES_PROVIDER=ubuntu ./kube-up.sh
    cd $2/cluster/ubuntu/
    KUBERNETES_PROVIDER=ubuntu ./deployAddons.sh
fi
