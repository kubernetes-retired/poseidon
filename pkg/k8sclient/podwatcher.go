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
	"github.com/kubernetes-sigs/poseidon/pkg/metrics"

	"github.com/golang/glog"
	"github.com/jinzhu/copier"
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

// SortNodeSelectorsKey sort node selectors keys and return an slice of sorted keys.
func SortNodeSelectorsKey(nodeSelector NodeSelectors) []string {
	var keyArray []string
	for key := range nodeSelector {
		// Skip key with networkRequirement
		if key == "networkRequirement" {
			continue
		}
		keyArray = append(keyArray, key)
	}
	sort.Strings(keyArray)

	return keyArray
}

// NewPodWatcher initialize a PodWatcher.
func NewPodWatcher(kubeVerMajor, kubeVerMinor int, schedulerName string, client kubernetes.Interface, fc firmament.FirmamentSchedulerClient) *PodWatcher {
	glog.V(2).Info("Starting PodWatcher...")
	PodMux = new(sync.RWMutex)
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

func (pw *PodWatcher) getCPUMemEphemeralRequest(pod *v1.Pod) (int64, int64, int64) {
	cpuReq := int64(0)
	memReq := int64(0)
	ephemeralReq := int64(0)
	for _, container := range pod.Spec.Containers {
		request := container.Resources.Requests
		cpuReqQuantity := request[v1.ResourceCPU]
		cpuReq += cpuReqQuantity.MilliValue()
		memReqQuantity := request[v1.ResourceMemory]
		memReqCont, _ := memReqQuantity.AsInt64()
		memReq += memReqCont
		ephemeralReqQuantity := request[v1.ResourceEphemeralStorage]
		ephemeralReqCont, _ := ephemeralReqQuantity.AsInt64()
		ephemeralReq += ephemeralReqCont
	}
	return cpuReq, memReq, ephemeralReq
}

func (pw *PodWatcher) getNodeSelectorTerm(pod *v1.Pod) []NodeSelectorTerm {
	var nodeSelTerm []NodeSelectorTerm
	if pod.Spec.Affinity != nil {
		if pod.Spec.Affinity.NodeAffinity != nil {
			if pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution != nil {
				err := copier.Copy(&nodeSelTerm, pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms)
				if err != nil {
					glog.Errorf("NodeSelectorTerm %v could not be copied, err: %v", pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms, err)
				}

			}
		}
	}
	return nodeSelTerm
}

func (pw *PodWatcher) getPreferredSchedulingTerm(pod *v1.Pod) []PreferredSchedulingTerm {
	var prefSchTerm []PreferredSchedulingTerm
	if pod.Spec.Affinity != nil {
		if pod.Spec.Affinity.NodeAffinity != nil {
			err := copier.Copy(&prefSchTerm, pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution)
			if err != nil {
				glog.Errorf("PreferredSchedulingTerm %v could not be copied, err: %v", pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution, err)
			}
		}

	}
	return prefSchTerm
}

func (pw *PodWatcher) getPodAffinityTerm(pod *v1.Pod) []PodAffinityTerm {
	var podAffTerm []PodAffinityTerm
	if pod.Spec.Affinity != nil {
		if pod.Spec.Affinity.PodAffinity != nil {
			err := copier.Copy(&podAffTerm, pod.Spec.Affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution)
			if err != nil {
				glog.Errorf("PodAffinityTerm %v could not be copied, err: %v", pod.Spec.Affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution, err)
			}
		}
	}
	return podAffTerm
}

func (pw *PodWatcher) getWgtPodAffinityTerm(pod *v1.Pod) []WeightedPodAffinityTerm {
	var wgtPodAffTerm []WeightedPodAffinityTerm
	if pod.Spec.Affinity != nil {
		if pod.Spec.Affinity.PodAffinity != nil {
			err := copier.Copy(&wgtPodAffTerm, pod.Spec.Affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution)
			if err != nil {
				glog.Errorf("WeightedPodAffinityTerm %v could not be copied, err: %v", pod.Spec.Affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution, err)
			}
		}
	}
	return wgtPodAffTerm
}

func (pw *PodWatcher) getPodAffinityTermforPodAntiAffinity(pod *v1.Pod) []PodAffinityTerm {
	var podAffTerm []PodAffinityTerm
	if pod.Spec.Affinity != nil {
		if pod.Spec.Affinity.PodAntiAffinity != nil {
			err := copier.Copy(&podAffTerm, pod.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution)
			if err != nil {
				glog.Errorf("PodAffinityTerm %v for PodAntiAffinity could not be copied, err: %v", pod.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution, err)
			}
		}
	}
	return podAffTerm
}

func (pw *PodWatcher) getWgtPodAffinityTermforPodAntiAffinity(pod *v1.Pod) []WeightedPodAffinityTerm {
	var wgtPodAffTerm []WeightedPodAffinityTerm
	if pod.Spec.Affinity != nil {
		if pod.Spec.Affinity.PodAntiAffinity != nil {
			err := copier.Copy(&wgtPodAffTerm, pod.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution)
			if err != nil {
				glog.Errorf("WeightedPodAffinityTerm %v for PodAntiAffinity could not be copied, err: %v", pod.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution, err)
			}
		}
	}
	return wgtPodAffTerm
}

func (pw *PodWatcher) getTolerations(pod *v1.Pod) []Toleration {
	var tolerations []Toleration

	copier.Copy(&tolerations, pod.Spec.Tolerations)

	return tolerations
}

func (pw *PodWatcher) parsePod(pod *v1.Pod) *Pod {
	cpuReq, memReq, ephemeralReq := pw.getCPUMemEphemeralRequest(pod)
	podPhase := PodUnknown
	switch pod.Status.Phase {
	case v1.PodPending:
		podPhase = PodPending
	case v1.PodRunning:
		podPhase = PodRunning
	case v1.PodSucceeded:
		podPhase = PodSucceeded
	case v1.PodFailed:
		podPhase = PodFailed
	}
	return &Pod{
		Identifier: PodIdentifier{
			Name:      pod.Name,
			Namespace: pod.Namespace,
		},
		State:          podPhase,
		CPURequest:     cpuReq,
		MemRequestKb:   memReq / bytesToKb,
		EphemeralReqKb: ephemeralReq / bytesToKb,
		Labels:         pod.Labels,
		Annotations:    pod.Annotations,
		NodeSelector:   pod.Spec.NodeSelector,
		OwnerRef:       GetOwnerReference(pod),
		Affinity: &Affinity{
			NodeAffinity: &NodeAffinity{
				HardScheduling: &NodeSelector{
					NodeSelectorTerms: pw.getNodeSelectorTerm(pod),
				},
				SoftScheduling: pw.getPreferredSchedulingTerm(pod),
			},
			PodAffinity: &PodAffinity{
				HardScheduling: pw.getPodAffinityTerm(pod),
				SoftScheduling: pw.getWgtPodAffinityTerm(pod),
			},
			PodAntiAffinity: &PodAffinity{
				HardScheduling: pw.getPodAffinityTermforPodAntiAffinity(pod),
				SoftScheduling: pw.getWgtPodAffinityTermforPodAntiAffinity(pod),
			},
		},
		CreateTimeStamp: pod.CreationTimestamp,
		Tolerations:     pw.getTolerations(pod),
	}
}

func (pw *PodWatcher) enqueuePodAddition(key interface{}, obj interface{}) {
	pod := obj.(*v1.Pod)
	addedPod := pw.parsePod(pod)

	// if the pod had volumes
	// check for the bounded volumes
	if len(pod.Spec.Volumes) > 0 {
		newPod := pod.DeepCopy()
		newPod, ok := pw.getPVNodeAffinity(pod.Spec.Volumes, newPod)
		if ok {
			addedPod = pw.parsePod(newPod)
		} else {
			// Note: after we ignore this pod, the same pod we could get updated
			// this case has to be handled
			// also we need to broadcast the pod failure event here.
			glog.Error("Falied to find the matching volumes for the pod", addedPod)
			return
		}
	}
	// update the pod
	// Note the sequence is importatnt
	PodToK8sPodLock.Lock()
	identifier := PodIdentifier{
		Name:      pod.Name,
		Namespace: pod.Namespace,
	}
	PodToK8sPod[identifier] = pod.DeepCopy()
	PodToK8sPodLock.Unlock()
	pw.podWorkQueue.Add(key, addedPod)
	glog.V(2).Info("enqueuePodAddition: Added pod ", addedPod.Identifier)
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
		ProcessedPodEventsLock.Lock()
		if _, ok := ProcessedPodEvents[deletedPod.Identifier]; ok {
			delete(ProcessedPodEvents, deletedPod.Identifier)
		}
		ProcessedPodEventsLock.Unlock()
		PodToK8sPodLock.Lock()
		if _, ok := PodToK8sPod[deletedPod.Identifier]; ok {
			// the only place where the pod is deleted from the map
			delete(PodToK8sPod, deletedPod.Identifier)
		}
		PodToK8sPodLock.Unlock()
		pw.podWorkQueue.Add(key, deletedPod)

		glog.V(2).Info("enqueuePodDeletion: Added pod ", deletedPod.Identifier)
	}
}

func (pw *PodWatcher) enqueuePodUpdate(key, oldObj, newObj interface{}) {
	oldPod := oldObj.(*v1.Pod)
	newPod := newObj.(*v1.Pod)
	if oldPod.Status.Phase != newPod.Status.Phase {
		// TODO(ionel): pw code assumes that if other fields changed as well then Firmament will automatically update them upon state transition. pw is currently not true.
		updatedPod := pw.parsePod(newPod)
		// update the pod
		PodToK8sPodLock.Lock()
		identifier := PodIdentifier{
			Name:      newPod.Name,
			Namespace: newPod.Namespace,
		}
		PodToK8sPod[identifier] = newPod.DeepCopy()
		PodToK8sPodLock.Unlock()
		pw.podWorkQueue.Add(key, updatedPod)
		glog.V(2).Infof("enqueuePodUpdate: Updated pod state change %v %s", updatedPod.Identifier, updatedPod.State)
		return
	}
	oldCPUReq, oldMemReq, oldEphemeralReq := pw.getCPUMemEphemeralRequest(oldPod)
	newCPUReq, newMemReq, newEphemeralReq := pw.getCPUMemEphemeralRequest(newPod)
	if oldCPUReq != newCPUReq || oldMemReq != newMemReq || oldEphemeralReq != newEphemeralReq ||
		!reflect.DeepEqual(oldPod.Labels, newPod.Labels) ||
		!reflect.DeepEqual(oldPod.Annotations, newPod.Annotations) ||
		!reflect.DeepEqual(oldPod.Spec.NodeSelector, newPod.Spec.NodeSelector) {
		if updatedPod := pw.parsePod(newPod); updatedPod != nil {
			// we need to change the state here
			updatedPod.State = PodUpdated
			pw.podWorkQueue.Add(key, updatedPod)
			glog.V(2).Infof("enqueuePodUpdate: Updated pod %v", updatedPod.Identifier)
		}
		return
	}
}

// Run starts a pod watcher.
func (pw *PodWatcher) Run(stopCh <-chan struct{}, nWorkers int) {
	defer utilruntime.HandleCrash()

	// The workers can stop when we are done.
	defer pw.podWorkQueue.ShutDown()
	defer glog.V(2).Info("Shutting down PodWatcher")
	glog.V(2).Info("Getting pod updates...")

	go pw.controller.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, pw.controller.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	glog.V(2).Info("Starting pod watching workers")
	for i := 0; i < nWorkers; i++ {
		go wait.Until(pw.podWorker, time.Second, stopCh)
	}

	<-stopCh
	glog.V(2).Info("Stopping pod watcher")
}

func (pw *PodWatcher) podWorker() {
	func() {
		wg := new(sync.WaitGroup)
		defer func() {
			wg.Wait()
		}()
		for {
			key, items, quit := pw.podWorkQueue.Get()
			if quit {
				return
			}
			wg.Add(1)
			go func(key interface{}, items []interface{}, wg *sync.WaitGroup) {
				defer func() {
					pw.podWorkQueue.Done(key)
					wg.Done()
				}()
				for _, item := range items {
					pod := item.(*Pod)
					switch pod.State {
					case PodPending:
						glog.V(2).Info("PodPending ", pod.Identifier)
						PodMux.Lock()

						// check if the pod already exists
						// this cases happened when Replicaset are used.
						// When a replicaset is delete it creates more pods with the same name
						_, ok := PodToTD[pod.Identifier]
						if ok {
							// we ignore this since the pod already exists
							// release the lock
							glog.V(2).Info("Pod already added", pod.Identifier.Name, pod.Identifier.Namespace)
							PodMux.Unlock()
							continue
						}
						jobID := pw.generateJobID(pod.OwnerRef)
						jd, ok := jobIDToJD[jobID]
						if !ok {
							jd = pw.createNewJob(pod.OwnerRef)
							jobIDToJD[jobID] = jd
							jobNumTasksToRemove[jobID] = 0
						}
						jobNumTasksToRemove[jobID]++
						taskCount := jobNumTasksToRemove[jobID]
						PodMux.Unlock()
						td := pw.addTaskToJob(pod, jd.Uuid, jd.Name, (taskCount))
						PodMux.Lock()
						// if taskCount is '1' it means root task, update the RootTask pointer in the JobDescriptor
						if taskCount == 1 {
							jd.RootTask = td
						}
						PodToTD[pod.Identifier] = td
						TaskIDToPod[td.GetUid()] = pod.Identifier
						taskDescription := &firmament.TaskDescription{
							TaskDescriptor: td,
							JobDescriptor:  jd,
						}
						PodMux.Unlock()
						metrics.SchedulingSubmitmLatency.Observe(metrics.SinceInMicroseconds(time.Time(pod.CreateTimeStamp.Time)))
						firmament.TaskSubmitted(pw.fc, taskDescription)
					case PodSucceeded:
						glog.V(2).Info("PodSucceeded ", pod.Identifier)
						PodMux.RLock()
						td, ok := PodToTD[pod.Identifier]
						PodMux.RUnlock()
						if !ok {
							glog.Fatalf("Pod %v does not exist", pod.Identifier)
						}
						firmament.TaskCompleted(pw.fc, &firmament.TaskUID{TaskUid: td.Uid})
					case PodDeleted:
						glog.V(2).Info("PodDeleted ", pod.Identifier)
						PodMux.RLock()
						td, ok := PodToTD[pod.Identifier]
						PodMux.RUnlock()
						if !ok {
							glog.Infof("Pod %s does not exist", pod.Identifier)
							continue
						}
						// TODO(jiaxuanzhou) need to metric the task remove latency ?
						firmament.TaskRemoved(pw.fc, &firmament.TaskUID{TaskUid: td.Uid})
						PodMux.Lock()
						delete(PodToTD, pod.Identifier)
						delete(TaskIDToPod, td.GetUid())
						// TODO(ionel): Should we delete the task from JD's spawned field?
						jobID := pw.generateJobID(pod.OwnerRef)
						jobNumTasksToRemove[jobID]--
						if jobNumTasksToRemove[jobID] == 0 {
							// Clean state because the job doesn't have any tasks left.
							delete(jobNumTasksToRemove, jobID)
							delete(jobIDToJD, jobID)
						}
						PodMux.Unlock()
					case PodFailed:
						glog.V(2).Info("PodFailed ", pod.Identifier)
						PodMux.RLock()
						td, ok := PodToTD[pod.Identifier]
						PodMux.RUnlock()
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
						PodMux.Lock()
						jobId := pw.generateJobID(pod.OwnerRef)
						jd, okJob := jobIDToJD[jobId]
						td, okPod := PodToTD[pod.Identifier]
						PodMux.Unlock()
						if !okJob {
							glog.Infof("Pod's %v job does not exist", pod.Identifier)
							continue
						}
						if !okPod {
							glog.Infof("Pod %v does not exist", pod.Identifier)
							continue
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
			}(key, items, wg)
		}
	}()
}

func (pw *PodWatcher) createNewJob(jobName string) *firmament.JobDescriptor {
	jobDesc := &firmament.JobDescriptor{
		Uuid:  pw.generateJobID(jobName),
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

	// update label selectors
	td.LabelSelectors = nil
	td.LabelSelectors = pw.getFirmamentLabelSelectorFromNodeSelectorMap(pod.NodeSelector, SortNodeSelectorsKey(pod.NodeSelector))

	//Add tolerations
	for _, tolerations := range pod.Tolerations {
		td.Toleration = append(td.Toleration,
			&firmament.Toleration{
				Key:      tolerations.Key,
				Value:    tolerations.Value,
				Operator: tolerations.Operator,
				Effect:   tolerations.Effect,
			})
	}

	nodeAffinity := len(pod.Affinity.NodeAffinity.HardScheduling.NodeSelectorTerms) > 0 || len(pod.Affinity.NodeAffinity.SoftScheduling) > 0
	podAffinity := len(pod.Affinity.PodAffinity.HardScheduling) > 0 || len(pod.Affinity.PodAffinity.SoftScheduling) > 0
	podAntiAffinity := len(pod.Affinity.PodAntiAffinity.HardScheduling) > 0 || len(pod.Affinity.PodAntiAffinity.SoftScheduling) > 0
	localAffinity := new(firmament.Affinity)

	if nodeAffinity {
		localAffinity.NodeAffinity = &firmament.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &firmament.NodeSelector{
				NodeSelectorTerms: pw.getFirmamentNodeSelTerm(pod),
			},
			PreferredDuringSchedulingIgnoredDuringExecution: pw.getFirmamentPreferredSchedulingTerm(pod),
		}
	}

	if podAffinity {
		localAffinity.PodAffinity = &firmament.PodAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution:  pw.getFirmamentPodAffinityTerm(pod),
			PreferredDuringSchedulingIgnoredDuringExecution: pw.getFirmamentWeightedPodAffinityTerm(pod),
		}
	}

	if podAntiAffinity {
		localAffinity.PodAntiAffinity = &firmament.PodAntiAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution:  pw.getFirmamentPodAffinityTermforPodAntiAffinity(pod),
			PreferredDuringSchedulingIgnoredDuringExecution: pw.getFirmamentWeightedPodAffinityTermforPodAntiAffinity(pod),
		}

	}

	td.Affinity = localAffinity
	if nodeAffinity == false && podAffinity == false && podAntiAffinity == false {
		td.Affinity = nil
	}
}

func (pw *PodWatcher) addTaskToJob(pod *Pod, jdUid string, jdName string, tdID int) *firmament.TaskDescriptor {
	task := &firmament.TaskDescriptor{
		Name:      pod.Identifier.UniqueName(),
		Namespace: pod.Identifier.Namespace,
		State:     firmament.TaskDescriptor_CREATED,
		JobId:     jdUid,
		ResourceRequest: &firmament.ResourceVector{
			// TODO(ionel): Update types so no cast is required.
			CpuCores:     float32(pod.CPURequest),
			RamCap:       uint64(pod.MemRequestKb),
			EphemeralCap: uint64(pod.EphemeralReqKb),
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

	//Add tolerations
	for _, tolerations := range pod.Tolerations {
		task.Toleration = append(task.Toleration,
			&firmament.Toleration{
				Key:      tolerations.Key,
				Value:    tolerations.Value,
				Operator: tolerations.Operator,
				Effect:   tolerations.Effect,
			})
	}
	// Get the network requirement from pods label, and set it in ResourceRequest of the TaskDescriptor
	setTaskNetworkRequirement(task, pod.Labels)
	task.LabelSelectors = pw.getFirmamentLabelSelectorFromNodeSelectorMap(pod.NodeSelector, SortNodeSelectorsKey(pod.NodeSelector))

	nodeAffinity := len(pod.Affinity.NodeAffinity.HardScheduling.NodeSelectorTerms) > 0 || len(pod.Affinity.NodeAffinity.SoftScheduling) > 0
	podAffinity := len(pod.Affinity.PodAffinity.HardScheduling) > 0 || len(pod.Affinity.PodAffinity.SoftScheduling) > 0
	podAntiAffinity := len(pod.Affinity.PodAntiAffinity.HardScheduling) > 0 || len(pod.Affinity.PodAntiAffinity.SoftScheduling) > 0
	localAffinity := new(firmament.Affinity)

	if nodeAffinity {
		localAffinity.NodeAffinity = &firmament.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &firmament.NodeSelector{
				NodeSelectorTerms: pw.getFirmamentNodeSelTerm(pod),
			},
			PreferredDuringSchedulingIgnoredDuringExecution: pw.getFirmamentPreferredSchedulingTerm(pod),
		}
	}

	if podAffinity {
		localAffinity.PodAffinity = &firmament.PodAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution:  pw.getFirmamentPodAffinityTerm(pod),
			PreferredDuringSchedulingIgnoredDuringExecution: pw.getFirmamentWeightedPodAffinityTerm(pod),
		}
	}

	if podAntiAffinity {
		localAffinity.PodAntiAffinity = &firmament.PodAntiAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution:  pw.getFirmamentPodAffinityTermforPodAntiAffinity(pod),
			PreferredDuringSchedulingIgnoredDuringExecution: pw.getFirmamentWeightedPodAffinityTermforPodAntiAffinity(pod),
		}

	}

	task.Affinity = localAffinity
	if nodeAffinity == false && podAffinity == false && podAntiAffinity == false {
		task.Affinity = nil
	}

	setTaskType(task)
	// No need to update the RootTask.Spawned here, it will be updated by firmament on processing the task submit call.
	task.Uid = pw.generateTaskID(jdName, tdID)
	return task
}

func (pw *PodWatcher) generateJobID(seed string) string {
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

func (pw *PodWatcher) getFirmamentLabelSelectorFromNodeSelectorMap(nodeSelector NodeSelectors, nodeSelectorKeys []string) []*firmament.LabelSelector {
	var firmamentLabelSelector []*firmament.LabelSelector
	for _, key := range nodeSelectorKeys {
		firmamentLabelSelector = append(firmamentLabelSelector, &firmament.LabelSelector{
			Type:   firmament.LabelSelector_IN_SET,
			Key:    key,
			Values: []string{nodeSelector[key]},
		})
	}
	return firmamentLabelSelector
}

func (pw *PodWatcher) getFirmamentNodeSelTerm(pod *Pod) []*firmament.NodeSelectorTerm {
	var fns []*firmament.NodeSelectorTerm
	err := copier.Copy(&fns, pod.Affinity.NodeAffinity.HardScheduling.NodeSelectorTerms)
	if err != nil {
		glog.Errorf("NodeSelectorTerm %v could not be copied to firmament, err: %v", pod.Affinity.NodeAffinity.HardScheduling.NodeSelectorTerms, err)
	}
	return fns
}

func (pw *PodWatcher) getFirmamentPreferredSchedulingTerm(pod *Pod) []*firmament.PreferredSchedulingTerm {
	var pst []*firmament.PreferredSchedulingTerm
	err := copier.Copy(&pst, pod.Affinity.NodeAffinity.SoftScheduling)
	if err != nil {
		glog.Errorf("PreferredSchedulingTerm %v could not be copied to firmament, err: %v", pod.Affinity.NodeAffinity.SoftScheduling, err)
	}
	return pst
}

func (pw *PodWatcher) getFirmamentPodAffinityTerm(pod *Pod) []*firmament.PodAffinityTerm {
	var pat []*firmament.PodAffinityTerm
	err := copier.Copy(&pat, pod.Affinity.PodAffinity.HardScheduling)
	if err != nil {
		glog.Errorf("PodAffinityTerm %v could not be copied to firmament, err: %v", pod.Affinity.PodAffinity.HardScheduling, err)
	}
	return pat
}

func (pw *PodWatcher) getFirmamentWeightedPodAffinityTerm(pod *Pod) []*firmament.WeightedPodAffinityTerm {
	var wpat []*firmament.WeightedPodAffinityTerm
	err := copier.Copy(&wpat, pod.Affinity.PodAffinity.SoftScheduling)
	if err != nil {
		glog.Errorf("WeightedPodAffinityTerm %v could not be copied to firmament, err: %v", pod.Affinity.PodAffinity.SoftScheduling, err)
	}
	return wpat
}

func (pw *PodWatcher) getFirmamentPodAffinityTermforPodAntiAffinity(pod *Pod) []*firmament.PodAffinityTermAntiAff {
	var pat []*firmament.PodAffinityTermAntiAff
	err := copier.Copy(&pat, pod.Affinity.PodAntiAffinity.HardScheduling)
	if err != nil {
		glog.Errorf("PodAffinityTerm %v for PodAntiAffinity could not be copied to firmament, err: %v", pod.Affinity.PodAntiAffinity.HardScheduling, err)
	}
	return pat
}

func (pw *PodWatcher) getFirmamentWeightedPodAffinityTermforPodAntiAffinity(pod *Pod) []*firmament.WeightedPodAffinityTermAntiAff {
	var wpat []*firmament.WeightedPodAffinityTermAntiAff
	err := copier.Copy(&wpat, pod.Affinity.PodAntiAffinity.SoftScheduling)
	if err != nil {
		glog.Errorf("WeightedPodAffinityTerm %v for PodAntiAffinity could not be copied to firmament, err: %v", pod.Affinity.PodAntiAffinity.SoftScheduling, err)
	}
	return wpat
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

// getPVNodeAffinity this method will check if the pod have any PVC's which are pre-bound to a PV and
// this will create an node selector term in the Pod's NodeAffinity section, based on the PV's node-selector and pass it on to firmamnet.
// When firmament see a node selector it will try to schdeule the task on that Node, provided all the other
// constraints are satisfied.
func (pw *PodWatcher) getPVNodeAffinity(volumes []v1.Volume, pod *v1.Pod) (*v1.Pod, bool) {

	var pvcName string
	if pod.Spec.Affinity != nil && pod.Spec.Affinity.NodeAffinity != nil {
		glog.V(2).Info("Pod already has an node affinity field ", pod.Spec.Affinity.NodeAffinity)
	}
	for _, v := range volumes {
		glog.V(2).Info(v, " is the volume name for pod ", pod.Name)
		if v.VolumeSource.PersistentVolumeClaim != nil {
			pvcName = v.VolumeSource.PersistentVolumeClaim.ClaimName
			glog.V(3).Info("Found a PV claim name ", pvcName, " for Pod", pod.Name)
		} else {
			glog.V(3).Info("No PV claim found for the Pod ", pod.Name)
			continue
		}
		// get the PVc object associated with the PV claim name from the api-server
		pvc, err := pw.clientset.CoreV1().PersistentVolumeClaims(pod.Namespace).Get(pvcName, metav1.GetOptions{})
		if err != nil {
			glog.Error(err, " Unable to retrieve the PVC object for PVC ", pvcName, " associated with Pod ", pod.Name)
			return nil, false
		}
		if pvc.Spec.VolumeName != "" {
			// search for PV associated with this PVC
			pv, err := pw.clientset.CoreV1().PersistentVolumes().Get(pvc.Spec.VolumeName, metav1.GetOptions{})
			if err != nil {
				glog.Error(err, "Unable to get the PV associated with the PVC volume name", pvc.Spec.VolumeName)
				return nil, false
			}
			var pvNodeSelector *v1.NodeSelector
			if pv.Spec.NodeAffinity != nil && pv.Spec.NodeAffinity.Required != nil {
				pvNodeSelector = pv.Spec.NodeAffinity.Required
			} else {
				glog.V(2).Info("No matching node selector for PV volume found", pvc.Spec.VolumeName)
				continue
			}
			// update the pods NodeAfiinity field with he PV's NodeAffinity term
			if pod.Spec.Affinity != nil {
				podNodeAffinity := pod.Spec.Affinity.NodeAffinity
				if podNodeAffinity != nil {
					// Check if the NodeAffinity Hard section is available, update only the RequiredDuringSchedulingIgnoredDuringExecution section
					// of the pod NodeAffinity
					if podNodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution != nil {
						podNodeSelector := podNodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution

						// check Pods NodeSelectorTerms and prepend PV's Node selector requirement to the first NodeSelectorTerms
						if len(podNodeSelector.NodeSelectorTerms) > 0 {
							//get the first nodeselector term and update the match expression with our term
							nodeSelectorTerm := podNodeSelector.NodeSelectorTerms[0]
							for _, pvNodeSelector := range pvNodeSelector.NodeSelectorTerms {
								//Note: we don't use the node fields terms
								nodeSelectorTerm.MatchExpressions = append(nodeSelectorTerm.MatchExpressions, pvNodeSelector.MatchExpressions...)
							}
							podNodeSelector.NodeSelectorTerms[0] = nodeSelectorTerm
						} else {
							podNodeSelector.NodeSelectorTerms = append(podNodeSelector.NodeSelectorTerms, pvNodeSelector.NodeSelectorTerms...)
						}
					} else {
						podNodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution = &v1.NodeSelector{
							NodeSelectorTerms: pvNodeSelector.NodeSelectorTerms,
						}
					}
				} else {
					pod.Spec.Affinity.NodeAffinity = &v1.NodeAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
							NodeSelectorTerms: pvNodeSelector.NodeSelectorTerms,
						}}
				}
			} else {
				pod.Spec.Affinity = &v1.Affinity{
					NodeAffinity: &v1.NodeAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
							NodeSelectorTerms: pvNodeSelector.NodeSelectorTerms,
						},
					},
				}
			}
		} else {
			// cannot find the right pv
			glog.V(2).Info("Cannot schedule this pod since no matchin PV found", pod.Name)
			return nil, false
		}
	}

	return pod, true
}

func Update(pw kubernetes.Interface, pod *v1.Pod, condition *v1.PodCondition) error {
	glog.V(1).Infof("Updating pod condition for %s/%s to (%s==%s)", pod.Namespace, pod.Name, condition.Type, condition.Status)
	if UpdatePodCondition(&pod.Status, condition) {
		_, err := pw.CoreV1().Pods(pod.Namespace).UpdateStatus(pod)
		return err
	}
	return nil
}

// Updates existing pod condition or creates a new one. Sets LastTransitionTime to now if the
// status has changed.
// Returns true if pod condition has changed or has been added.
func UpdatePodCondition(status *v1.PodStatus, condition *v1.PodCondition) bool {
	condition.LastTransitionTime = metav1.Now()
	// Try to find this pod condition.
	conditionIndex, oldCondition := GetPodCondition(status, condition.Type)

	if oldCondition == nil {
		// We are adding new pod condition.
		status.Conditions = append(status.Conditions, *condition)
		return true
	} else {
		// We are updating an existing condition, so we need to check if it has changed.
		if condition.Status == oldCondition.Status {
			condition.LastTransitionTime = oldCondition.LastTransitionTime
		}

		isEqual := condition.Status == oldCondition.Status &&
			condition.Reason == oldCondition.Reason &&
			condition.Message == oldCondition.Message &&
			condition.LastProbeTime.Equal(&oldCondition.LastProbeTime) &&
			condition.LastTransitionTime.Equal(&oldCondition.LastTransitionTime)

		status.Conditions[conditionIndex] = *condition
		// Return true if one of the fields have changed.
		return !isEqual
	}
}

// GetPodCondition extracts the provided condition from the given status and returns that.
// Returns nil and -1 if the condition is not present, and the index of the located condition.
func GetPodCondition(status *v1.PodStatus, conditionType v1.PodConditionType) (int, *v1.PodCondition) {
	if status == nil {
		return -1, nil
	}
	return GetPodConditionFromList(status.Conditions, conditionType)
}

// GetPodConditionFromList extracts the provided condition from the given list of condition and
// returns the index of the condition and the condition. Returns -1 and nil if the condition is not present.
func GetPodConditionFromList(conditions []v1.PodCondition, conditionType v1.PodConditionType) (int, *v1.PodCondition) {
	if conditions == nil {
		return -1, nil
	}
	for i := range conditions {
		if conditions[i].Type == conditionType {
			return i, &conditions[i]
		}
	}
	return -1, nil
}
