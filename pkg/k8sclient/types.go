/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package k8sclient

import (
	"sync"

	"github.com/kubernetes-sigs/poseidon/pkg/firmament"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

const bytesToKb = 1024

// PodMux is used to guard access to the pod, task and job related maps.
var PodMux *sync.RWMutex

// PodToTD maps Kubernetes pod identifier(namespace + name) to firmament task descriptor.
var PodToTD map[PodIdentifier]*firmament.TaskDescriptor

// TaskIDToPod maps firmament task ID to Kubernetes pod identifier(namespace + name).
var TaskIDToPod map[uint64]PodIdentifier
var jobIDToJD map[string]*firmament.JobDescriptor
var jobNumTasksToRemove map[string]int

// NodeMux is used to guard access to the node and resource related maps.
var NodeMux *sync.RWMutex

// NodeToRTND maps node name to firmament resource topology node descriptor.
var NodeToRTND map[string]*firmament.ResourceTopologyNodeDescriptor

// ResIDToNode maps resource ID to node name.
var ResIDToNode map[string]string

// NodePhase represents a node phase.
type NodePhase string

const (
	// NodeAdded represents a node added phase.
	NodeAdded NodePhase = "Added"
	// NodeDeleted represents a node deleted phase.
	NodeDeleted NodePhase = "Deleted"
	// NodeFailed represents a node failed phase.
	NodeFailed NodePhase = "Failed"
	// NodeUpdated represents a node updated phase.
	NodeUpdated NodePhase = "Updated"
)

// Node is an internal structure for a Kubernetes node.
type Node struct {
	Hostname         string
	Phase            NodePhase
	IsReady          bool
	IsOutOfDisk      bool
	CPUCapacity      int64
	CPUAllocatable   int64
	MemCapacityKb    int64
	MemAllocatableKb int64
	Labels           map[string]string
	Annotations      map[string]string
}

// PodPhase represents a pod phase.
type PodPhase string

const (
	// PodPending is an internal phase used for unscheduled pods.
	PodPending PodPhase = "Pending"
	// PodRunning is an internal phase used for running pods.
	PodRunning PodPhase = "Running"
	// PodSucceeded is an internal phase used for successfully existed pods.
	PodSucceeded PodPhase = "Succeeded"
	// PodFailed is an internal phase used for failed pods.
	PodFailed PodPhase = "Failed"
	// PodUnknown is an internal phase used for state unknown pods.
	PodUnknown PodPhase = "Unknown"
	// PodDeleted is an internal phase used for removed pods.
	PodDeleted PodPhase = "Deleted"
	// PodUpdated is an internal phase for pods that are externally updated.
	PodUpdated PodPhase = "Updated"
)

// PodIdentifier is used to identify a pod by its namespace and name.
type PodIdentifier struct {
	Name      string
	Namespace string
}

// UniqueName returns pod namespace/name.
func (this *PodIdentifier) UniqueName() string {
	return this.Namespace + "/" + this.Name
}

//Node Affinity Struct
type NodeSelectorRequirement struct {
	Key      string
	Operator string
	Values   []string
}

type NodeSelector struct {
	//Required. A list of node selector terms. The terms are ORed.
	NodeSelectorTerms []NodeSelectorTerm
}

// A null or empty node selector term matches no objects.
type NodeSelectorTerm struct {
	MatchExpressions []NodeSelectorRequirement
}
type PreferredSchedulingTerm struct {
	// Weight associated with matching the corresponding nodeSelectorTerm, in the range 1-100.
	Weight int32
	// A node selector term, associated with the corresponding weight.
	Preference NodeSelectorTerm
}
type NodeAffinity struct {
	HardScheduling *NodeSelector
	SoftScheduling []PreferredSchedulingTerm
}

//--------Pod Affinity -----//
// Pod affinity is a group of inter pod affinity scheduling rules.

type PodAffinityTerm struct {
	LabelSelector *metav1.LabelSelector
	Namespaces    []string
	TopologyKey   string
}

type WeightedPodAffinityTerm struct {
	Weight          int32
	PodAffinityTerm PodAffinityTerm
}

type PodAffinity struct {
	HardScheduling []PodAffinityTerm
	SoftScheduling []WeightedPodAffinityTerm
}

type Affinity struct {
	NodeAffinity    *NodeAffinity
	PodAffinity     *PodAffinity
	PodAntiAffinity *PodAffinity
}

// Pod is an internal structure for a Kubernetes pod.
type Pod struct {
	Identifier      PodIdentifier
	State           PodPhase
	CPURequest      int64
	MemRequestKb    int64
	Labels          map[string]string
	Annotations     map[string]string
	NodeSelector    map[string]string
	OwnerRef        string
	Affinity        *Affinity
	CreateTimeStamp metav1.Time
}

// NodeWatcher is a Kubernetes node watcher.
type NodeWatcher struct {
	//ID string
	clientset     kubernetes.Interface
	nodeWorkQueue Queue
	controller    cache.Controller
	fc            firmament.FirmamentSchedulerClient
}

// PodWatcher is a Kubernetes pod watcher.
type PodWatcher struct {
	//ID string
	clientset    kubernetes.Interface
	podWorkQueue Queue
	controller   cache.Controller
	fc           firmament.FirmamentSchedulerClient
}
