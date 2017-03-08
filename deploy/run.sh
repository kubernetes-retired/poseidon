#!/bin/bash
DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
echo {$DIR}


sudo kubectl delete deployments poseidon-scheduler -n kube-system
sudo docker ps -a | grep poseidon-scheduler | awk '{print $1}' | xargs sudo docker rm -f
sudo docker images -a | grep poseidon-test | awk '{print $3}' | xargs sudo docker rmi -f


cp ${DIR}/../build/poseidon poseidon
cp ${DIR}/../build/third_party/ms-cpprestsdk/src/cpp-rest-sdk-build/Binaries/libcpprest.so.2.7 libcpprest.so.2.7
cp ${DIR}/../build/firmament/src/firmament-build/src/libfirmament_scheduling.so libfirmament_scheduling.so
cp ${DIR}/../build/firmament/src/firmament-build/third_party/cs2/src/cs2/cs2.exe cs2.exe

docker build --tag poseidon-test ${DIR}
docker tag poseidon-test localhost:5000/poseidon-test
docker push localhost:5000/poseidon-test

sudo kubectl create -f /home/ubuntu/poseidon-test.yaml
sleep 10
sudo kubectl get pods --all-namespaces
