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

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/poseidon/pkg/firmament"
)

const bytesToKb = 1024

// Used to guard access to the pod, task and job related maps.
var PodsCond *sync.Cond
var PodToTD map[PodIdentifier]*firmament.TaskDescriptor
var TaskIDToPod map[uint64]PodIdentifier
var jobIDToJD map[string]*firmament.JobDescriptor
var jobNumTasksToRemove map[string]int

// Used to guard access to the node and resource related maps.
var NodesCond *sync.Cond
var NodeToRTND map[string]*firmament.ResourceTopologyNodeDescriptor
var ResIDToNode map[string]string

type NodePhase string

const (
	NodeAdded   NodePhase = "Added"
	NodeDeleted NodePhase = "Deleted"
	NodeFailed  NodePhase = "Failed"
	NodeUpdated NodePhase = "Updated"
)

type Node struct {
	Hostname         string
	Phase            NodePhase
	IsReady          bool
	IsOutOfDisk      bool
	CpuCapacity      int64
	CpuAllocatable   int64
	MemCapacityKb    int64
	MemAllocatableKb int64
	Labels           map[string]string
	Annotations      map[string]string
}

type PodPhase string

const (
	PodPending   PodPhase = "Pending"
	PodRunning   PodPhase = "Running"
	PodSucceeded PodPhase = "Succeeded"
	PodFailed    PodPhase = "Failed"
	PodUnknown   PodPhase = "Unknown"
	// Internal phase used for removed pods.
	PodDeleted PodPhase = "Deleted"
	// Internal phase for pods that are externally updated.
	PodUpdated PodPhase = "Updated"
)

type PodIdentifier struct {
	Name      string
	Namespace string
}

func (this *PodIdentifier) UniqueName() string {
	return this.Namespace + "/" + this.Name
}

type Pod struct {
	Identifier   PodIdentifier
	State        PodPhase
	CpuRequest   int64
	MemRequestKb int64
	Labels       map[string]string
	Annotations  map[string]string
	NodeSelector map[string]string
	OwnerRef     string
}

type NodeWatcher struct {
	//ID string
	clientset     kubernetes.Interface
	nodeWorkQueue Queue
	controller    cache.Controller
	fc            firmament.FirmamentSchedulerClient
}

type PodWatcher struct {
	//ID string
	clientset    kubernetes.Interface
	podWorkQueue Queue
	controller   cache.Controller
	fc           firmament.FirmamentSchedulerClient
}
