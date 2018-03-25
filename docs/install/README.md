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

## RBAC cluster

   For RBAC enabled clusters one must update the ```system:kube-scheduler clusterrole``` and add the Poseidon service-account to 
   the ```system:kube-scheduler clusterrolebinding ```
   
   
   * Add the Poseidon service-account to the cluster-rolebinding as shown below: 
   
```yaml
- kind: ServiceAccount
  name: poseidon
  namespace: kube-system
  
```  
This service account is created by the poseidon-deployment.yaml file.
Here we add it to the appropriate cluster-rolebinding.
     
```yaml
$ kubectl edit clusterrolebinding system:kube-scheduler
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  annotations:
    rbac.authorization.kubernetes.io/autoupdate: "true"
  creationTimestamp: 2018-01-08T11:38:48Z
  labels:
    kubernetes.io/bootstrapping: rbac-defaults
  name: system:kube-scheduler
  resourceVersion: "781144"
  selfLink: /apis/rbac.authorization.k8s.io/v1/clusterrolebindings/system%3Akube-scheduler
  uid: 7db4b6e2-f468-11e7-a7c5-fa163ee4f284
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:kube-scheduler
subjects:
- apiGroup: rbac.authorization.k8s.io
  kind: User
  name: system:kube-scheduler
- kind: ServiceAccount
  name: poseidon
  namespace: kube-system

```

  * Add the Poseidon scheduler name under the ‘resourceNames’ as shown below.
```yaml
 resourceNames:
  - poseidon
  
```
```yaml
$ kubectl edit clusterrole system:kube-scheduler
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  annotations:
    rbac.authorization.kubernetes.io/autoupdate: "true"
  creationTimestamp: 2018-01-08T11:38:47Z
  labels:
    kubernetes.io/bootstrapping: rbac-defaults
  name: system:kube-scheduler
  resourceVersion: "811176"
  selfLink: /apis/rbac.authorization.k8s.io/v1/clusterroles/system%3Akube-scheduler
  uid: 7d82c981-f468-11e7-a7c5-fa163ee4f284
rules:
- apiGroups:
  - ""
  resources:
  - events
  verbs:
  - create
  - patch
  - update
- apiGroups:
  - ""
  resources:
  - endpoints
  verbs:
  - create
- apiGroups:
  - ""
  resourceNames:
  - kube-scheduler
  - poseidon

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
  
