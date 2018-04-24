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
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/kubernetes-sigs/poseidon/pkg/firmament"

	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

// NodeSelectors stores Kubernetes node selectors.
type NodeSelectors map[string]string

// Redefine below Annotation key as that is deprecated from original Kubernetes.
const (
	// CreatedByAnnotation represents the original Kubernetes `kubernetes.io/created-by` annotation.
	CreatedByAnnotation = "kubernetes.io/created-by"
)

// SortNodeSelectors sort node selectors by its key name.
func SortNodeSelectors(nodeSelector NodeSelectors) NodeSelectors {
	newSortedNodeSelectors := make(NodeSelectors)
	var keyArray []string
	for key := range nodeSelector {
		// Skip key with networkRequirement
		if key == "networkRequirement" {
			continue
		}
		keyArray = append(keyArray, key)
	}
	sort.Strings(keyArray)
	for _, key := range keyArray {
		newSortedNodeSelectors[key] = nodeSelector[key]
	}
	return newSortedNodeSelectors
}

// NewPodWatcher initialize a PodWatcher.
func NewPodWatcher(kubeVerMajor, kubeVerMinor int, schedulerName string, client kubernetes.Interface, fc firmament.FirmamentSchedulerClient) *PodWatcher {
	glog.Info("Starting PodWatcher...")
	PodsCond = sync.NewCond(&sync.Mutex{})
	PodToTD = make(map[PodIdentifier]*firmament.TaskDescriptor)
	TaskIDToPod = make(map[uint64]PodIdentifier)
	jobIDToJD = make(map[string]*firmament.JobDescriptor)
	jobNumTasksToRemove = make(map[string]int)
	podWatcher := &PodWatcher{
		clientset: client,
		fc:        fc,
	}
	schedulerSelector := fields.Everything()
	podSelector := labels.Everything()
	if kubeVerMajor >= 1 && kubeVerMinor >= 6 {
		// schedulerName is only available in Kubernetes >= 1.6.
		schedulerSelector = fields.ParseSelectorOrDie("spec.schedulerName==" + schedulerName)
	} else {
		var err error
		podSelector, err = labels.Parse("scheduler in (" + schedulerName + ")")
		if err != nil {
			glog.Fatal("Failed to parse scheduler label selector")
		}
	}
	_, controller := cache.NewInformer(
		&cache.ListWatch{
			ListFunc: func(alo metav1.ListOptions) (runtime.Object, error) {
				alo.FieldSelector = schedulerSelector.String()
				alo.LabelSelector = podSelector.String()
				return client.CoreV1().Pods("").List(alo)
			},
			WatchFunc: func(alo metav1.ListOptions) (watch.Interface, error) {
				alo.FieldSelector = schedulerSelector.String()
				alo.LabelSelector = podSelector.String()
				return client.CoreV1().Pods("").Watch(alo)
			},
		},
		&v1.Pod{},
		0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				key, err := cache.MetaNamespaceKeyFunc(obj)
				if err != nil {
					glog.Errorf("AddFunc: error getting key %v", err)
				}
				podWatcher.enqueuePodAddition(key, obj)
			},
			UpdateFunc: func(old, new interface{}) {
				key, err := cache.MetaNamespaceKeyFunc(new)
				if err != nil {
					glog.Errorf("UpdateFunc: error getting key %v", err)
				}
				podWatcher.enqueuePodUpdate(key, old, new)
			},
			DeleteFunc: func(obj interface{}) {
				key, err := cache.MetaNamespaceKeyFunc(obj)
				if err != nil {
					glog.Errorf("DeleteFunc: error getting key %v", err)
				}
				podWatcher.enqueuePodDeletion(key, obj)
			},
		},
	)
	podWatcher.controller = controller
	podWatcher.podWorkQueue = NewKeyedQueue()
	return podWatcher
}

func (pw *PodWatcher) getCPUMemRequest(pod *v1.Pod) (int64, int64) {
	cpuReq := int64(0)
	memReq := int64(0)
	for _, container := range pod.Spec.Containers {
		request := container.Resources.Requests
		cpuReqQuantity := request["cpu"]
		cpuReq += cpuReqQuantity.MilliValue()
		memReqQuantity := request["memory"]
		memReqCont, _ := memReqQuantity.AsInt64()
		memReq += memReqCont
	}
	return cpuReq, memReq
}

func (pw *PodWatcher) parsePod(pod *v1.Pod) *Pod {
	cpuReq, memReq := pw.getCPUMemRequest(pod)
	podPhase := PodPhase("Unknown")
	switch pod.Status.Phase {
	case "Pending":
		podPhase = "Pending"
	case "Running":
		podPhase = "Running"
	case "Succeeded":
		podPhase = "Succeeded"
	case "Failed":
		podPhase = "Failed"
	}
	return &Pod{
		Identifier: PodIdentifier{
			Name:      pod.Name,
			Namespace: pod.Namespace,
		},
		State:        podPhase,
		CPURequest:   cpuReq,
		MemRequestKb: memReq / bytesToKb,
		Labels:       pod.Labels,
		Annotations:  pod.Annotations,
		NodeSelector: pod.Spec.NodeSelector,
		OwnerRef:     GetOwnerReference(pod),
	}
}

func (pw *PodWatcher) enqueuePodAddition(key interface{}, obj interface{}) {
	pod := obj.(*v1.Pod)
	addedPod := pw.parsePod(pod)
	pw.podWorkQueue.Add(key, addedPod)
	glog.Info("enqueuePodAddition: Added pod ", addedPod.Identifier)
}

func (pw *PodWatcher) enqueuePodDeletion(key interface{}, obj interface{}) {
	pod := obj.(*v1.Pod)
	if pod.DeletionTimestamp != nil {
		// Only delete pods if they have a DeletionTimestamp.
		deletedPod := &Pod{
			Identifier: PodIdentifier{
				Name:      pod.Name,
				Namespace: pod.Namespace,
			},
			State:    PodDeleted,
			OwnerRef: GetOwnerReference(pod),
		}
		pw.podWorkQueue.Add(key, deletedPod)
		glog.Info("enqueuePodDeletion: Added pod ", deletedPod.Identifier)
	}
}

func (pw *PodWatcher) enqueuePodUpdate(key, oldObj, newObj interface{}) {
	oldPod := oldObj.(*v1.Pod)
	newPod := newObj.(*v1.Pod)
	if oldPod.Status.Phase != newPod.Status.Phase {
		// TODO(ionel): pw code assumes that if other fields changed as well then Firmament will automatically update them upon state transition. pw is currently not true.
		updatedPod := pw.parsePod(newPod)
		pw.podWorkQueue.Add(key, updatedPod)
		glog.Infof("enqueuePodUpdate: Updated pod state change %v %s", updatedPod.Identifier, updatedPod.State)
		return
	}
	oldCPUReq, oldMemReq := pw.getCPUMemRequest(oldPod)
	newCPUReq, newMemReq := pw.getCPUMemRequest(newPod)
	if oldCPUReq != newCPUReq || oldMemReq != newMemReq ||
		!reflect.DeepEqual(oldPod.Labels, newPod.Labels) ||
		!reflect.DeepEqual(oldPod.Annotations, newPod.Annotations) ||
		!reflect.DeepEqual(oldPod.Spec.NodeSelector, newPod.Spec.NodeSelector) {
		updatedPod := pw.parsePod(newPod)
		pw.podWorkQueue.Add(key, updatedPod)
		glog.Info("enqueuePodUpdate: Updated pod ", updatedPod.Identifier)
		return
	}
}

// Run starts a pod watcher.
func (pw *PodWatcher) Run(stopCh <-chan struct{}, nWorkers int) {
	defer utilruntime.HandleCrash()

	// The workers can stop when we are done.
	defer pw.podWorkQueue.ShutDown()
	defer glog.Info("Shutting down PodWatcher")
	glog.Info("Getting pod updates...")

	go pw.controller.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, pw.controller.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	glog.Info("Starting pod watching workers")
	for i := 0; i < nWorkers; i++ {
		go wait.Until(pw.podWorker, time.Second, stopCh)
	}

	<-stopCh
	glog.Info("Stopping pod watcher")
}

func (pw *PodWatcher) podWorker() {
	for {
		func() {
			key, items, quit := pw.podWorkQueue.Get()
			if quit {
				return
			}
			for _, item := range items {
				pod := item.(*Pod)
				switch pod.State {
				case PodPending:
					glog.V(2).Info("PodPending ", pod.Identifier)
					PodsCond.L.Lock()
					jobID := pw.generatejobID(pod.OwnerRef)
					jd, ok := jobIDToJD[jobID]
					if !ok {
						jd = pw.createNewJob(pod.OwnerRef)
						jobIDToJD[jobID] = jd
						jobNumTasksToRemove[jobID] = 0
					}
					td := pw.addTaskToJob(pod, jd)
					jobNumTasksToRemove[jobID]++
					PodToTD[pod.Identifier] = td
					TaskIDToPod[td.GetUid()] = pod.Identifier
					taskDescription := &firmament.TaskDescription{
						TaskDescriptor: td,
						JobDescriptor:  jd,
					}
					PodsCond.L.Unlock()
					firmament.TaskSubmitted(pw.fc, taskDescription)
				case PodSucceeded:
					glog.V(2).Info("PodSucceeded ", pod.Identifier)
					PodsCond.L.Lock()
					td, ok := PodToTD[pod.Identifier]
					PodsCond.L.Unlock()
					if !ok {
						glog.Fatalf("Pod %v does not exist", pod.Identifier)
					}
					firmament.TaskCompleted(pw.fc, &firmament.TaskUID{TaskUid: td.Uid})
				case PodDeleted:
					glog.V(2).Info("PodDeleted ", pod.Identifier)
					PodsCond.L.Lock()
					td, ok := PodToTD[pod.Identifier]
					PodsCond.L.Unlock()
					if !ok {
						glog.Fatalf("Pod %s does not exist", pod.Identifier)
					}
					firmament.TaskRemoved(pw.fc, &firmament.TaskUID{TaskUid: td.Uid})
					PodsCond.L.Lock()
					delete(PodToTD, pod.Identifier)
					delete(TaskIDToPod, td.GetUid())
					// TODO(ionel): Should we delete the task from JD's spawned field?
					jobID := pw.generatejobID(pod.OwnerRef)
					jobNumTasksToRemove[jobID]--
					if jobNumTasksToRemove[jobID] == 0 {
						// Clean state because the job doesn't have any tasks left.
						delete(jobNumTasksToRemove, jobID)
						delete(jobIDToJD, jobID)
					}
					PodsCond.L.Unlock()
				case PodFailed:
					glog.V(2).Info("PodFailed ", pod.Identifier)
					PodsCond.L.Lock()
					td, ok := PodToTD[pod.Identifier]
					PodsCond.L.Unlock()
					if !ok {
						glog.Fatalf("Pod %s does not exist", pod.Identifier)
					}
					firmament.TaskFailed(pw.fc, &firmament.TaskUID{TaskUid: td.Uid})
				case PodRunning:
					glog.V(2).Info("PodRunning ", pod.Identifier)
					// We don't have to do anything.
				case PodUnknown:
					glog.Errorf("Pod %s in unknown state", pod.Identifier)
					// TODO(ionel): Handle Unknown case.
				case PodUpdated:
					glog.V(2).Info("PodUpdated ", pod.Identifier)
					PodsCond.L.Lock()
					jobID := pw.generatejobID(pod.OwnerRef)
					jd, okJob := jobIDToJD[jobID]
					td, okPod := PodToTD[pod.Identifier]
					PodsCond.L.Unlock()
					if !okJob {
						glog.Fatalf("Pod's %v job does not exist", pod.Identifier)
					}
					if !okPod {
						glog.Fatalf("Pod %v does not exist", pod.Identifier)
					}
					pw.updateTask(pod, td)
					taskDescription := &firmament.TaskDescription{
						TaskDescriptor: td,
						JobDescriptor:  jd,
					}
					firmament.TaskUpdated(pw.fc, taskDescription)
				default:
					glog.Fatalf("Pod %v in unexpected state %v", pod.Identifier, pod.State)
				}
			}
			defer pw.podWorkQueue.Done(key)
		}()
	}
}

func (pw *PodWatcher) createNewJob(jobName string) *firmament.JobDescriptor {
	jobDesc := &firmament.JobDescriptor{
		Uuid:  pw.generatejobID(jobName),
		Name:  jobName,
		State: firmament.JobDescriptor_CREATED,
	}
	return jobDesc
}

func (pw *PodWatcher) updateTask(pod *Pod, td *firmament.TaskDescriptor) {
	// TODO(ionel): Update LabelSelector!
	td.ResourceRequest.CpuCores = float32(pod.CPURequest)
	td.ResourceRequest.RamCap = uint64(pod.MemRequestKb)
	// Update labels.
	td.Labels = nil
	for label, value := range pod.Labels {
		td.Labels = append(td.Labels,
			&firmament.Label{
				Key:   label,
				Value: value,
			})
	}
}

func (pw *PodWatcher) addTaskToJob(pod *Pod, jd *firmament.JobDescriptor) *firmament.TaskDescriptor {
	task := &firmament.TaskDescriptor{
		Name:  pod.Identifier.UniqueName(),
		State: firmament.TaskDescriptor_CREATED,
		JobId: jd.Uuid,
		ResourceRequest: &firmament.ResourceVector{
			// TODO(ionel): Update types so no cast is required.
			CpuCores: float32(pod.CPURequest),
			RamCap:   uint64(pod.MemRequestKb),
		},
	}

	// Add labels.
	for label, value := range pod.Labels {
		task.Labels = append(task.Labels,
			&firmament.Label{
				Key:   label,
				Value: value,
			})
	}
	// Get the network requirement from pods label, and set it in ResourceRequest of the TaskDescriptor
	setTaskNetworkRequirement(task, pod.Labels)
	task.LabelSelectors = pw.getFirmamentLabelSelectorFromNodeSelectorMap(SortNodeSelectors(pod.NodeSelector))
	setTaskType(task)

	if jd.RootTask == nil {
		task.Uid = pw.generateTaskID(jd.Name, 0)
		jd.RootTask = task
	} else {
		task.Uid = pw.generateTaskID(jd.Name, len(jd.RootTask.Spawned)+1)
		jd.RootTask.Spawned = append(jd.RootTask.Spawned, task)
	}
	return task
}

func (pw *PodWatcher) generatejobID(seed string) string {
	if seed == "" {
		glog.Fatal("Seed value is nil")
	}

	return GenerateUUID(seed)
}

func (pw *PodWatcher) generateTaskID(jdUID string, taskNum int) uint64 {
	return HashCombine(jdUID, taskNum)
}

// GetOwnerReference to get the parent object reference
func GetOwnerReference(pod *v1.Pod) string {
	// Return if owner reference exists.
	ownerRefs := pod.GetObjectMeta().GetOwnerReferences()
	if len(ownerRefs) != 0 {
		for x := range ownerRefs {
			ref := &ownerRefs[x]
			if ref.Controller != nil && *ref.Controller {
				return string(ref.UID)
			}
		}
	}

	// Return the controller-uid label if it exists.
	if controllerID := pod.GetObjectMeta().GetLabels()["controller-uid"]; controllerID != "" {
		return controllerID
	}

	// Return 'kubernetes.io/created-by' if it exists.
	if createdByAnnotation, ok := pod.GetObjectMeta().GetAnnotations()[CreatedByAnnotation]; ok {
		var serialCreatedBy v1.SerializedReference
		err := json.Unmarshal([]byte(createdByAnnotation), &serialCreatedBy)
		if err == nil {
			return string(serialCreatedBy.Reference.UID)
		}
	}

	// Return the uid of the ObjectMeta if none from the above is present.
	return string(pod.GetObjectMeta().GetUID())
}

func (pw *PodWatcher) getFirmamentLabelSelectorFromNodeSelectorMap(nodeSelector NodeSelectors) []*firmament.LabelSelector {
	var firmamentLabelSelector []*firmament.LabelSelector
	for key, value := range nodeSelector {
		firmamentLabelSelector = append(firmamentLabelSelector, &firmament.LabelSelector{
			Type:   firmament.LabelSelector_IN_SET,
			Values: []string{value},
			Key:    key,
		})
	}
	return firmamentLabelSelector
}

func setTaskNetworkRequirement(td *firmament.TaskDescriptor, nodeSelectors NodeSelectors) {
	if val, ok := nodeSelectors["networkRequirement"]; ok {
		res, err := strconv.ParseUint(val, 10, 64)
		if err == nil {
			td.ResourceRequest.NetRxBw = res
		} else {
			glog.Errorf("Failed to parse networkRequirement %v", err)
		}
	}
}

func setTaskType(td *firmament.TaskDescriptor) {
	for _, label := range td.Labels {
		if label.Key == "taskType" {
			switch label.Value {
			case "Sheep":
				td.TaskType = firmament.TaskDescriptor_SHEEP
			case "Rabbit":
				td.TaskType = firmament.TaskDescriptor_RABBIT
			case "Devil":
				td.TaskType = firmament.TaskDescriptor_DEVIL
			case "Turtle":
				td.TaskType = firmament.TaskDescriptor_TURTLE
			default:
				glog.Errorf("Unexpected task type %s for task %s", label.Value, td.Name)
			}
		}
	}
}
