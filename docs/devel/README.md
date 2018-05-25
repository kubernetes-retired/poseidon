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

  * **Building Poseidon without Bazel:**
  
  
 ```
 $ mkdir -p $GOPATH/src/github.com/kubernetes-sigs
 $ cd $GOPATH/src/github.com/kubernetes-sigs
 $ git clone https://github.com/kubernetes-sigs/poseidon
 $ cd poseidon
 $ cd cmd/poseidon
 $ go build .
 ```

 * **Building Poseidon using Bazel:**
   * Refer [Bazel](https://docs.bazel.build/versions/master/install.html) on how to install Bazel.
 ```
 $ mkdir -p $GOPATH/src/github.com/kubernetes-sigs
 $ cd $GOPATH/src/github.com/kubernetes-sigs
 $ git clone https://github.com/kubernetes-sigs/poseidon
 $ cd poseidon
 $ bazel build //cmd/poseidon
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
    --firmamentAddress=<host> \
    --firmamentPort=<port> \
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
--firmamentAddress=<host> \
--firmamentPort=<port> \
--statsServerAddress=<host>:<port> \ 
--kubeVersion=<Major.Minor>
```

**Note:**
The order of execution is, first Firmament has to be started and then Poseidon is started with the Firmament's address 
and Firmament Port.
The order is required only when we run Poseidon and Firmament manually.
This order is not required for installation methods, since the Poseidon service will not start-up till Firmament service is available.

# Running Unit Tests
Using Bazel
```
$ cd $GOPATH/src/github.com/kubernetes-sigs/poseidon
$ bazel test -- //... -//hack/... -//vendor/... -//test/e2e/...
```

Using go Test
```
$ cd $GOPATH/src/github.com/kubernetes-sigs/poseidon
$ go test $(go list ./... | grep -v /vendor/ | grep -v /test/ | grep -v /hack/)

```

# Testing the setup
Run the below script and check if the pods are scheduled.
```
kubectl create -f https://raw.githubusercontent.com/kubernetes-sigs/poseidon/master/deploy/configs/cpu_spin.yaml
```

Few test scripts are available [here](https://github.com/kubernetes-sigs/poseidon/tree/master/deploy/configs).

# Local Cluster E2E test
To run E2E test on a local cluster.

```
$ cd $GOPATH/src/github.com/kubernetes-sigs/poseidon/test/e2e
$ go test -v . -ginkgo.v \
-args -kubeconfig=/home/ubuntu/.kube/config \ 
-poseidonVersion=${BUILD_VERSION} \
-gcrProject="google_containers"
```
You can get ```${BUILD_VERSION}``` by ```BUILD_VERSION=$(git rev-parse HEAD)```
```kubeconfig``` should point to the running local k8s cluster.

***Note***
You need to have a working kubernetes cluster to run the 
above test. You can optionally try ```kubetest``` , to deploy a kubernetes
cluster on your gce account. Please refer the doc [here](https://github.com/kubernetes/test-infra/tree/master/kubetest).

# Building Release packages locally

```
$ cd $GOPATH/src/github.com/kubernetes-sigs/poseidon
$ make release
```

# Testing release packages
The best way to test the release packages locally, is to run the
below script. It will build the release tar push it to docker locally and run the e2e tests.

***Note***

The ```'kubeconfig'``` path should be ```$HOME/.kube/config```.
The below script run based on the above assumptions.
And it should point to a running k8s cluster.
If your running k8s cluster that is started by local-up-cluster.sh, you should ```export HOSTNAME_OVERRIDE=$master-ip``` before running local-up-cluster.sh.
```$master-ip``` is the non-loopback IP of your machine where running the k8s cluster.
And copy ```KUBECONFIG```(such as ```/var/run/kubernetes/admin.kubeconfig```) to ```$HOME/.kube/config``` before running test/e2e-poseidon-local.sh.

```
$ $GOPATH/src/github.com/kubernetes-sigs/poseidon/test/e2e
$ test/e2e-poseidon-local.sh
```

# Code contribution
We recommend running the following, before raising a PR.

This will test all the essential checks. 

```
$ make verify
```

All the existing unit tests should [pass](https://github.com/kubernetes-sigs/poseidon/tree/master/docs/devel#running-unit-tests).
Also recommend to run the local release test mentioned [here](https://github.com/kubernetes-sigs/poseidon/tree/master/docs/devel#testing-release-packages).

