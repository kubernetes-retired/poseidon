# To enable Gang Scheduling 

Supply the below annotation in JOB's, Deployments and Replicaset to enable Gang Scheduling.
The value "80" in the below annotation is a sample which tell firmament the minimum pods required 
by this JOB. To specify strict or relaxed type gang scheduling use the value to specify the same.
A value of "100" means 100 percent or strict gang scheduling requirement.
Any value less than "100" specifies relaxed gang scheduling.

**Note**: The value can't be more than "100".

```
 annotations:
        "firmament-gang-scheduling" : "80"
```

### Note:
We currently support only jobs. Deployment and replicasets are not fully tested.

# Sample

* **Assumption**: Three worker nodes and one master node with 1200 MiliCPU available.
 

## Case One
Available [here](https://raw.githubusercontent.com/kubernetes-sigs/poseidon/master/deploy/gang-scheduling/gang_schedule_test_case_one.yaml).

 
This test case will try to deploy four pods with 1200 MiliCPU and with the gang scheduling requirement of "75%" it will try to 
schedule three pods. Since the minimum requirement is met firmament will schedule three pods.
It will not be able to fit the fourth pod since the resource requirement for the fourth pod is not met.

**Expected output:** Three pods in running state with one pod still in pending state.


```$json
apiVersion: batch/v1
kind: Job
metadata:
  name: 1-test-gang-job
spec:
  parallelism: 4
  template:
    metadata:
      name: test-1
      annotations:
        "firmament-gang-scheduling" : "75"
    spec:
      schedulerName: poseidon
      containers:
      - name: nginx
        image: "nginx:1.11.1-alpine"
        resources:
          requests:
           memory: "120Mi"
           cpu: "1200m"
          limits:
            memory: "130Mi"
            cpu: "1300m"
      restartPolicy: Never
```

## Case Two

Available [here](https://raw.githubusercontent.com/kubernetes-sigs/poseidon/master/deploy/gang-scheduling/gang_schedule_test_case_two.yaml).

This test case will try to deploy Five pods with 1200 MiliCPU and with the gang scheduling requirement of "80%" it will try to 
schedule four pods. 
This will not be able to fit the fourth pod since the resource requirement for the fourth pod is not met.
Hence none of the pods will be scheduled, since the minimum requirement is to schedule four pods.

**Expected output:** None of the pods will be scheduled.


```$json
apiVersion: batch/v1
kind: Job
metadata:
  name: 2-test-gang-job
spec:
  parallelism: 5
  template:
    metadata:
      name: test-2
      annotations:
        "firmament-gang-scheduling" : "80"
    spec:
      schedulerName: poseidon
      containers:
      - name: nginx
        image: "nginx:1.11.1-alpine"
        resources:
          requests:
           memory: "120Mi"
           cpu: "1200m"
          limits:
            memory: "130Mi"
            cpu: "1300m"
      restartPolicy: Never
```

