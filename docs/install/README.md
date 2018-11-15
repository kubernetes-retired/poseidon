# Installation

## Prerequisite
   * Running kubernetes cluster :- These installation steps assume there is a running kubernetes cluster already setup.   
     Please refer [kubernetes setup](https://kubernetes.io/docs/setup/) for more info.
   
## Dependencies resolution
  The 'Poseidon' scheduler cannot run without Firmament being in a 'running' state.
  But this dependency is automatically resolved by 'Poseidon', there is no need for manual intervention like 
  before. Following is the mechanism used to resolve the dependency problem.
  
### Dependencies resolution approach

   The 'init-container' will check if the 'Firmament' service is available by doing a 'nslookup' on the 'Firmament' 
   service. Only when the 'Firmament' service is available it will start the 'Poseidon's' container.
   Sometimes the 'Firmament service' is registered and visible but the actual gRPC methods are not available yet.
   When 'Poseidon' starts, it will check if the 'Firmament' service is available by performing a gRPC health-check call.
   If the service is not up it will wait till the 'Firmament' gPRC methods are available. This is useful while running 
   'Poseidon' and 'Firmament' as a standalone process or as docker containers as well.

## Overview
   The [architecture diagram](https://github.com/kubernetes-sigs/poseidon#design) shows the various components of Poseidon integration.
   
   Both Poseidon and Firmament run as deployment each exposed as a service to communicate with each other.
   Firmament's service is used by Poseidon to send nodes, pods and other information.
   For more detail info on the design please refer design [docs](https://github.com/kubernetes-sigs/poseidon/blob/master/docs/design/README.md).
   
   
  The easiest way is to use the deployment [scripts](../../deploy/).
  
  ## Steps

  * Step 1:- Create the Firmament deployment
```
kubectl create -f https://raw.githubusercontent.com/kubernetes-sigs/poseidon/master/deploy/firmament-deployment.yaml

```
  * Step 2:- Create the Poseidon deployment
```
kubectl create -f https://raw.githubusercontent.com/kubernetes-sigs/poseidon/master/deploy/poseidon-deployment.yaml

```

# Testing the installation
  To check if the above setup works fine, deploy the below yaml.
  
  
```
kubectl create -f https://raw.githubusercontent.com/kubernetes-sigs/poseidon/master/deploy/configs/cpu_spin.yaml

```
 Check if the above JOB is running.

```
kubectl get pods -n default
```

# Running workloads using 'Poseidon/Firmament' scheduler:

Specify schedulerName as poseidon as part of the pod spec definition as shown below.

```$json
spec:
  schedulerName: poseidon

```