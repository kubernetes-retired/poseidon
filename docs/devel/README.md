# Developer Setup

This document show how to build and run Poseidon and other components on a dev setup.

* Dependency 
   * Kubernetes :- Running instance of a [kubernetes cluster](https://kubernetes.io/docs/setup/) is required. 
   * Firmament  :- For Firmament build info please refer [here](https://github.com/camsas/firmament/blob/master/README.md#building-instructions).
   * Heapster   :- For deploying heapster with Poseidon sink. please use these [instructions](https://github.com/kubernetes-sigs/poseidon/tree/master/docs/install#steps).

Before running Poseidon all the above three components must be running.

**Note:** 

   Heapster sink for Poseidon is not yet merged in the heapster repo.
   We will be doing that shortly. Please refer to the deployment scripts already created for [heapster](https://raw.githubusercontent.com/kubernetes-sigs/poseidon/master/deploy/heapster-poseidon.yaml). 
   
   For more info on the Heapster sink for Poseidon please refer [here](https://github.com/camsas/heapster).
   
   
# System requirements
  * Go 1.9+
  * Ubuntu 16.04
  * Kubernetes v1.5+
  * Docker 1.7+

# Build

  * **Building Firmament:**
  
       For completeness, we have included the Firmament build steps here.
     
     
```
$ git clone -b dev https://github.com/Huawei-PaaS/firmament
$ cd firmament
$ mkdir build
$ cd build
$ cmake ..
$ make
```

**Note:**
Currently the Firmament repo referred in this document is our dev repo.
We will be soon pointing it toward the main [repo](https://github.com/camsas/firmament) after our features are merged.

  * **Building Poseidon:**
  
  
 ```
 $ mkdir -p $GOPATH/src/k8s.io
 $ cd $GOPATH/src/k8s.io
 $ git clone https://github.com/kubernetes-sigs/poseidon
 # cd poseidon
 $ go build .
 
 ```
  
 # Docker container Build
 
   * **Building Firmament docker container:**
```
$ git clone -b dev https://github.com/Huawei-PaaS/firmament
$ cd firmament/contrib
$ ./docker-build.sh
```
This will create a container and push it in the local registry.


   * **Building Poseidon docker container:**
   
```
$ git clone https://github.com/kubernetes-sigs/poseidon
$ cd poseidon/deploy
$ ./build_docker_image.sh
```
This will create a container and push it in the local registry.


# Running
  * **Running Firmament as a process:**
  
```
$ cd firmament
$ ./build/src/firmament_scheduler --flagfile=config/firmament_scheduler.cfg

```
For more information on the arguments that could be passed to Firmament [please refer](https://github.com/Huawei-PaaS/firmament#using-the-flow-scheduler)

One can also use the below to get the list of supported arguments by Firmament.

```
./build/src/firmament_scheduler --help
```

  * **Running Poseidon as a process:**
      
      To run Poseidon as an independent process, it requires the kubeconfig (file) and Firmament's endpoint to be supplied as arguments.

 ```
 $ ./poseidon --logtostderr \
    --kubeConfig=<path_kubeconfig_file> \
    --firmamentAddress=<host>:<port> \
    --statsServerAddress=<host>:<port> \
    --kubeVersion=<Major.Minor>
 ```

  * **Running Firmament as docker container:**
    
```
sudo docker run --net=host firmament:dev /firmament/build/src/firmament_scheduler \
--flagfile=/firmament/config/firmament_scheduler_cpu_mem.cfg

```

  * **Running Poseidon as docker container:**
```
sudo docker run --net=host --volume=$GOPATH/src/github.com/kubernetes-sigs/poseidon/kubeconfig.cfg:/config/kubeconfig.cfg \
gcr.io/poseidon-173606/poseidon:latest \
--logtostderr \
--kubeConfig=/config/kubeconfig.cfg \
--firmamentAddress=<host>:<port> \
--statsServerAddress=<host>:<port> \ 
--kubeVersion=<Major.Minor>
```

**Note:**
The order of execution is, first Firmament has to be started and then Poseidon is started with the Firmament's address.
The order is required only when we run Poseidon and Firmament manually.
This order is not required for installation methods, since the Poseidon service will not start-up till Firmament service is available.


# Testing the setup
Run the below script and check if the pods are scheduled.
```
kubectl create -f https://raw.githubusercontent.com/kubernetes-sigs/poseidon/master/deploy/configs/cpu_spin.yaml
```

Few test scripts are available [here](https://github.com/kubernetes-sigs/poseidon/tree/master/deploy/configs).

# Local Cluster E2E test
To run E2E test on a local cluster.

```
cd poseidon/pkg/test
go test -args --testKubeConfig=<path to the kubeconfig file>
```
The local cluster should have the Poseidon and Firmament already deployed and running.
Only very basic check is supported currenlty.
We will be adding more test cases to this E2E.



