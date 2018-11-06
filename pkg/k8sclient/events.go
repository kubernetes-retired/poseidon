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
	"github.com/golang/glog"
	"github.com/kubernetes-sigs/poseidon/pkg/firmament"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	"sync"
)

type PodEvents struct {
	Recorder    record.EventRecorder
	BroadCaster record.EventBroadcaster
}

type PoseidonEvents struct {
	podEvents *PodEvents
	k8sClient kubernetes.Interface
	sync.Mutex
}

var poseidonEvents *PoseidonEvents
var poseidonEventsLock sync.Mutex

// NewPoseidonEvents singleton Events object create function
func NewPoseidonEvents(coreEvent kubernetes.Interface) *PoseidonEvents {

	poseidonEventsLock.Lock()
	if poseidonEvents == nil {
		poseidonEvents = &PoseidonEvents{
			podEvents: NewPodEvents(coreEvent),
			k8sClient: coreEvent,
		}
	}
	poseidonEventsLock.Unlock()

	return poseidonEvents
}

func NewPodEvents(coreEvent kubernetes.Interface) *PodEvents {

	sch := legacyscheme.Scheme

	sch.AddKnownTypes(schema.GroupVersion{Group: "Poseidon", Version: "v1"}, &corev1.Pod{})
	objarr, ok, err := sch.ObjectKinds(&corev1.Pod{})

	glog.Info(objarr, ok, err)
	if err != nil {
		glog.Fatalf("legacyscheme err %v", err)
	}

	err = corev1.AddToScheme(legacyscheme.Scheme)
	//err=api.AddToScheme(legacyscheme.Scheme)
	if err != nil {
		glog.Fatalf("could not register schemes %v", err)
	}
	broadCaster := record.NewBroadcaster()
	recorder := broadCaster.NewRecorder(sch, corev1.EventSource{Component: "Poseidon"})
	broadCaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: coreEvent.CoreV1().Events("")})
	//recorder.Event()

	return &PodEvents{
		BroadCaster: broadCaster,
		Recorder:    recorder,
	}
}

func (posiedonEvents *PoseidonEvents) ProcessEvents(deltas *firmament.SchedulingDeltas) {

	//process in success events and failure events in prallel
	// Note we
	go posiedonEvents.ProcessFailureEvents(deltas.GetUnscheduledTasks())
	go poseidonEvents.ProcessSuccessEvents(deltas.GetDeltas())
}

// ProcessFailureEvents The failed/unscheduled task events are sent only once
func (posiedonEvents *PoseidonEvents) ProcessFailureEvents(unscheduledTasks []uint64) {
	//get the pod name from the unscheduled_tasks id
	for _, taskId := range unscheduledTasks {
		PodMux.RLock()
		podIdentifier, ok := TaskIDToPod[taskId]
		PodMux.RUnlock()
		if !ok {
			glog.Error("Task id to Pod mapping not found ", taskId)
			continue
		}

		ProcessedPodEventsLock.Lock()
		if _, ok := ProcessedPodEvents[podIdentifier]; !ok {
			//update the ProcessedPodEvents map first
			PodToK8sPodLock.Lock()
			if poseidonToK8sPod, ok := PodToK8sPod[podIdentifier]; ok {
				ProcessedPodEvents[podIdentifier] = poseidonToK8sPod
				// send the failure event and update the pods status
				posiedonEvents.podEvents.Recorder.Eventf(poseidonToK8sPod, corev1.EventTypeWarning, "FailedScheduling", "Firmament failed to schedule the pod %s in %s namespace", podIdentifier.Namespace, podIdentifier.Name)
				Update(posiedonEvents.k8sClient, poseidonToK8sPod, &corev1.PodCondition{
					Type:    corev1.PodScheduled,
					Status:  corev1.ConditionFalse,
					Reason:  corev1.PodReasonUnschedulable,
					Message: "Firmament unable to schedule the pod",
				})
			} else {
				glog.Error("k8s pod mapping for ", podIdentifier, " pod not found ")
			}
			PodToK8sPodLock.Unlock()
		} else {
			glog.V(2).Info("For ", podIdentifier, " already failure/unscheduled events sent")
		}
		ProcessedPodEventsLock.Unlock()
	}
}

// ProcessSuccessEvents send success event to api-server
func (posiedonEvents *PoseidonEvents) ProcessSuccessEvents(scheduledTasks []*firmament.SchedulingDelta) {
	//get the pod name from the unscheduled_tasks id
	for _, taskId := range scheduledTasks {
		// Note: We need to process only deltas of type PLACE
		if taskId.GetType() == firmament.SchedulingDelta_PLACE {
			PodMux.RLock()
			podIdentifier, ok := TaskIDToPod[taskId.GetTaskId()]
			PodMux.RUnlock()
			if !ok {
				glog.Errorf("Task id %v to Pod mapping not found ", taskId)
				continue
			}
			ProcessedPodEventsLock.Lock()
			if _, ok := ProcessedPodEvents[podIdentifier]; ok {
				// we remove this pod from  ProcessedPodEvents if it exists
				// this should be done when we get a delete on this pod also
				delete(ProcessedPodEvents, podIdentifier)
			}
			ProcessedPodEventsLock.Unlock()

			// Now send the success event
			PodToK8sPodLock.Lock()
			if poseidonToK8sPod, ok := PodToK8sPod[podIdentifier]; ok {
				NodeMux.RLock()
				nodeName, ok := ResIDToNode[taskId.GetResourceId()]
				NodeMux.RUnlock()
				if ok {
					posiedonEvents.podEvents.Recorder.Eventf(poseidonToK8sPod, corev1.EventTypeNormal, "Scheduled", "Successfully assigned %v/%v to %v", podIdentifier.Namespace, podIdentifier.Name, nodeName)
				} else {
					glog.Error("Node not found in Node to Resource Id mapping", taskId.GetResourceId())
				}
			} else {
				glog.Error("Pod mapping not found in PodToK8sPod map for Pod ", podIdentifier)
			}
			PodToK8sPodLock.Unlock()
		}
	}
}
