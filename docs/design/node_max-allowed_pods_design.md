# Node Max-allowed Pods Design for Firmament/Poseidon Scheduler

- [Motivation](#motivation)
    - [Goals](#goals)
    - [User Stories](#user-stories)
        - [Story 1](#story-1)

- [Proposed Design](#proposed-design)
    - [Firmament Design Details](#firmament-design-details)
        - [CPU-Memory Cost Model Enhancements](#cpu-memory-cost-model-enhancements)
        - [Key Processing Design Details](#key-processing-design-details)
    - [Poseidon Design Details](#key-processing-design-details)

# Motivation

There are some circumstances where one may want to control the maximum number of pods running on a given node.

Firmament CPU/Memory cost model (scheduling policy) typically does a task placement based on resource dimensions (e.g. do not place the pod on a node with insufficient free resources, etc.).  We think the limit number of pods which are allowed to run is a resource of node as well.

Purpose of this design document is to enable such validation functionality within Poseidon/Firmament scheduling environment by extending Firmament CPU/Memory cost model.

## Goals

Kubernetes default scheduler validates maximum number of pod that are allowed to run using “pods” field as part of the `NodeStatus.Capacity`. Goal is to leverage these kubernetes node definitions in order to enable pod limit number validation functionality within Firmament scheduler.

## User Stories

### Story 1
Suppose node A at maximum can run 10 pods and already has 10 pods running on it. Then, any pod can't be scheduled to node A until the number of pods running on node A decreases to less than 10.


# Proposed Design

## Firmament Design Details

### CPU-Memory Cost Model Enhancements
Firstly, a brief description of underlying new CPU/Memory cost model. In this cost model each machine will have a set of predefined number of machine ECs (M0EC1,M0EC2,..,M2EC2) in order to do load distribution across filtered machines during each scheduling iteration. In practice, firmament limit the number of machine ECs for a given machine by setting the flag `--max-multi_arcs_for_cpu`, which is default to 50.

Let us take an example where there is a machine M0, and it has a capacity of 2 pods. Load distribution is achieved by assigning two arcs via corresponding machine ECs. Capacity on these arcs would be 1 each.

### Key Processing Design Details

The flag `--max-multi_arcs_for_cpu` within firmament has global effect - the values for each node are always the same. However, the max-allowed pods may differ from node to node.

The `num_slots_below` field of `ResourceDescriptor` in firmament provides us a way to dynamically configure the number of max-allowed pods running on a given node. Therefore, when a machine is added to the system, firmament will create `${num_slots_below}` (instead of `${FLAGS_max_multi_arcs_for_cpu}`) machine ECs connecting to the machine node. The cost of each arc is set by other hard/soft requirements and the cap of each arc is always 1. As the number of running tasks increases, we will keep decreasing the number of machine ECs. The formula would be: (number_of_machines_ECs = Max_pods - Number of current running tasks)

Currently firmament code is not updating current running tasks properly. Because we are using static resource calculation instead of heapster. So we need to modify `GatherStats()` function in cpu_cost_model.cc, which should update the current running tasks for each PU and accumulate each PU's running tasks in machine's resource descriptor.

Replacing static number of slots of PU.

Current firmament code fixed the maximum number of pods that can be scheduled on PU by using constant number `FLAGS_max_tasks_per_pu`. So because of this when `running_tasks` on PU equals `FLAGS_max_tasks_per_pu`, capacity of arc from Machine to PU is being set to zero. So we are restricted to schedule only `FLAGS_max_tasks_per_pu` number of tasks/pods on the PU. So in order to remove this restriction we need to set this `rd.num_slots_below()` to maximum pods that can be scheduled on that PU, that too only for PU. Code changes can be like,
Add new field `max_pods` in `ResourceDescriptor`, which gets value from kubelet parameter ‘max-pods’ only once when machine is added.
For PU node only, while updating the resource descriptor, update `num_slots_below` to `max_pods` like below.

`rd.set_num_slots_below(machine_rd.max_pods());`

In this way, we can schedule `max_pods` number of pods on that PU not just `FLAGS_max_tasks_per_pu`.

## Poseidon Design Details

Poseidson will watch Kubernetes nodes changes and parse the value of nodeStatus.Allocatable. Eventually, `nodeStatus.Allocatable` will be passed to firmament as the value of `ResourceDescriptior.num_slots_below`.
