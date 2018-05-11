[![Build Status](https://travis-ci.org/kubernetes-sigs/poseidon.svg?branch=master)](https://travis-ci.org/kubernetes-sigs/poseidon)

# Introduction
The Poseidon/Firmament scheduler incubation project is to bring integration of Firmament Scheduler [OSDI paper](https://www.usenix.org/conference/osdi16/technical-sessions/presentation/gog) in Kubernetes.
At a very high level, Poseidon/Firmament scheduler augments the 
current Kubernetes scheduling capabilities by incorporating a new 
novel flow network graph based scheduling capabilities alongside the default Kubernetes Scheduler. 
Firmament models workloads on a cluster as flow networks and runs min-cost flow optimizations over these networks to make scheduling decisions.

Due to the inherent rescheduling capabilities, the new scheduler enables a globally optimal scheduling for a given policy that keeps on refining the dynamic placements of the workload.

As we all know that as part of the Kubernetes multiple schedulers support, each new pod is typically scheduled by the default scheduler, but Kubernetes can be instructed to use another scheduler by specifying the name of another custom scheduler (Poseidon in our case) at the time of pod deployment. In this case, the default scheduler will ignore that Pod and allow Poseidon scheduler to schedule the Pod on a relevant node. We plugin Poseidon as an add-on scheduler to K8s, by using the 'schedulerName' as Poseidon in the pod template this will by-pass the default-scheduler. We plugin Poseidon as an add-on scheduler to K8s, by using the 'schedulerName' as Poseidon in the pod template this will by-pass the default-scheduler.

# Key Advantages

* Flow graph scheduling provides the following 
  * Support for high-volume workloads placement.
  * Complex rule constraints. 
  * Globally optimal scheduling for a given policy.
  * Extremely high scalability. 
  
  **NOTE:** Additionally, it is also very important to highlight that Firmament scales much better than default scheduler as the number of nodes increase in a cluster.

# Current Project Stage
**Alpha Release**

# Design 

   <p align="center">
  <img src="docs/poseidon.png"> 
<p align="center"> <b>Poseidon/Firmament Integration architecture</b> </p>
</p>



For more details about the design of this project see the [design document](https://docs.google.com/document/d/1VNoaw1GoRK-yop_Oqzn7wZhxMxvN3pdNjuaICjXLarA/edit?usp=sharing) doc.



# Installation
  In-cluster installation of Poseidon, please start [here](https://github.com/kubernetes-sigs/poseidon/blob/master/docs/install/README.md).
  
  
  
# Development
  For developers please refer [here](https://github.com/kubernetes-sigs/poseidon/blob/master/docs/devel/README.md)

# Release Process
To view details related to coordinated release process between Firmament & Poseidon repos, refer [here](https://github.com/kubernetes-sigs/poseidon/blob/master/docs/releases/release-process.md).

# Roadmap
  * **Release 0.1** – Currently Available:
    * Baseline Poseidon/Firmament Scheduling capabilities using new multi-dimensional CPU/Memory cost model is part of 
      this release. Currently, this does not include node and pod level affinity/anti-affinity capabilities. 
      As shown below, we are building all this out as part of the upcoming releases.    
    * Entire test.infra BOT automation jobs are in place as part of this release.
    
  * **Release 0.2** – Target Date: 25th May 2018:
    * Node level Affinity and Anti-Affinity implementation.
  * **Release 0.3** – Target Date: 15th June 2018:
    * Pod level Affinity and Anti-Affinity implementation using multi-round scheduling based affinity and anti-affinity.
  * **Release 0.4** – Tentative Target Date: 29th June 2018:
    * Taints & Tolerations.
  * **Release 0.5** onwards:
    * Transitioning to Metrics server API – Our current work for upstreaming new Heapster sink is not a possibility as Heapster is getting deprecated.
    * Continuous running scheduling loop versus scheduling intervals mechanism.
    * Provide High Availability/Failover for in-memory Firmament/Poseidon processes.
    * Gang Scheduling.
    * Priority Pre-emption support.
    * Resource Utilization benchmark.