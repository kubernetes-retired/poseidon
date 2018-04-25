# Installation

## Prerequisite
   * Running kubernetes cluster :- These installation steps assume there is a running kubernetes cluster already setup.   
     Please refer [kubernetes setup](https://kubernetes.io/docs/setup/) for more info.
   
## Depends on 
   * Running Firmament scheduler ( refer step 1 )
   * Heapster instance running with Firmament sink. ( refer step 3 )
  

## Overview
   The [architecture diagram](https://github.com/kubernetes-sigs/poseidon/tree/script_changes#design) shows the various components of Posedion integration.
   
   Both Poseidon and Firmament run as deployment each exposed as a service to communicate with each other.
   Firmament's service is used by Poseidon to send nodes, pods and other information. 
   Poseidon's service is used by heapster sink to push the metrics info, which again is pushed to Firmament's knowledge base.
   
   For more detail info on the design please refer design docs.
   
   
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
  * Step 3:- Create the heapster deployment
 
```
kubectl create –f https://raw.githubusercontent.com/kubernetes-sigs/poseidon/master/deploy/heapster-poseidon.yaml

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
  
