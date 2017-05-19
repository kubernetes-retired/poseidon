// Poseidon
// Copyright (c) The Poseidon Authors.
// All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// THIS CODE IS PROVIDED ON AN *AS IS* BASIS, WITHOUT WARRANTIES OR
// CONDITIONS OF ANY KIND, EITHER EXPRESS OR IMPLIED, INCLUDING WITHOUT
// LIMITATION ANY IMPLIED WARRANTIES OR CONDITIONS OF TITLE, FITNESS FOR
// A PARTICULAR PURPOSE, MERCHANTABLITY OR NON-INFRINGEMENT.
//
// See the Apache Version 2.0 License for specific language governing
// permissions and limitations under the License.

package k8sclient

import (
	"fmt"
	"time"

	"github.com/ICGog/poseidongo/pkg/firmament"
	"github.com/golang/glog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

func NewPodWatcher(client kubernetes.Interface, firmamentAddress string) *PodWatcher {
	glog.Info("Starting PodWatcher...")
	PodToTD = make(map[PodIdentifier]*firmament.TaskDescriptor)
	TaskIDToPod = make(map[uint64]PodIdentifier)
	jobIDToJD = make(map[string]*firmament.JobDescriptor)
	jobNumIncompleteTasks = make(map[string]int)
	// TODO(ionel): Close connection.
	fc, _, err := firmament.New(firmamentAddress)
	if err != nil {
		panic(err)
	}
	podWatcher := &PodWatcher{
		clientset: client,
		fc:        fc,
	}
	podStatuSelector := fields.ParseSelectorOrDie("spec.nodeName==")
	_, controller := cache.NewInformer(
		&cache.ListWatch{
			ListFunc: func(alo metav1.ListOptions) (runtime.Object, error) {
				alo.FieldSelector = podStatuSelector.String()
				return client.CoreV1().Pods("").List(alo)
			},
			WatchFunc: func(alo metav1.ListOptions) (watch.Interface, error) {
				alo.FieldSelector = podStatuSelector.String()
				return client.CoreV1().Pods("").Watch(alo)
			},
		},
		&v1.Pod{},
		0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				podWatcher.enqueuePodAddition(obj)
			},
			UpdateFunc: func(old, new interface{}) {
				podWatcher.enqueuePodUpdate(old, new)
			},
			DeleteFunc: func(obj interface{}) {
				podWatcher.enqueuePodDeletion(obj)
			},
		},
	)
	podWatcher.controller = controller
	podWatcher.podWorkQueue = workqueue.NewNamedDelayingQueue("PodQueue")
	return podWatcher
}

func (this *PodWatcher) enqueuePodAddition(obj interface{}) {
	pod := obj.(*v1.Pod)
	cpuReq := int64(0)
	memReq := int64(0)
	for _, container := range pod.Spec.Containers {
		request := container.Resources.Requests
		cpuReqQuantity := request["cpu"]
		cpuReqCont, _ := cpuReqQuantity.AsInt64()
		cpuReq += cpuReqCont
		memReqQuantity := request["memory"]
		memReqCont, _ := memReqQuantity.AsInt64()
		memReq += memReqCont
	}
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
	addedPod := &Pod{
		Identifier: PodIdentifier{
			Name:      pod.Name,
			Namespace: pod.Namespace,
		},
		State:        podPhase,
		CpuRequest:   cpuReq,
		MemRequestKb: memReq / bytesToKb,
		Labels:       pod.Labels,
		Annotations:  pod.Annotations,
		NodeSelector: pod.Spec.NodeSelector,
	}
	this.podWorkQueue.Add(addedPod)
	glog.Info("enqueuePodAddition: Added pod ", addedPod.Identifier)
}

func (this *PodWatcher) enqueuePodDeletion(obj interface{}) {
	// TODO(ionel): This method is called when the pod is scheduled!
	pod := obj.(*v1.Pod)
	deletedPod := &Pod{
		Identifier: PodIdentifier{
			Name:      pod.Name,
			Namespace: pod.Namespace,
		},
		State: PodDeleted,
	}
	//	this.podWorkQueue.Add(deletedPod)
	glog.Info("enqueuePodDeletion: Added pod ", deletedPod.Identifier)
}

func (this *PodWatcher) enqueuePodUpdate(oldObj, newObj interface{}) {
	// TODO(ionel): Implement!
}

func (this *PodWatcher) Run(stopCh <-chan struct{}, nWorkers int) {
	defer utilruntime.HandleCrash()

	// The workers can stop when we are done.
	defer this.podWorkQueue.ShutDown()
	defer glog.Info("Shutting down PodWatcher")
	glog.Info("Getting pod updates...")

	go this.controller.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, this.controller.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	glog.Info("Starting pod watching workers")
	for i := 0; i < nWorkers; i++ {
		go wait.Until(this.podWorker, time.Second, stopCh)
	}

	<-stopCh
	glog.Info("Stopping pod watcher")
}

func (this *PodWatcher) podWorker() {
	for {
		func() {
			key, quit := this.podWorkQueue.Get()
			if quit {
				return
			}
			pod := key.(*Pod)
			switch pod.State {
			case PodPending:
				// TODO(ionel): We generate a job per pod. Add a field to the Pod struct that uniquely identifies jobs/daemon sets and use that one instead to group pods into Firmament jobs.
				jobId := this.generateJobID(pod.Identifier.Name)
				jd, ok := jobIDToJD[jobId]
				if !ok {
					jd = this.createNewJob(pod.Identifier.Name)
					jobIDToJD[jobId] = jd
					jobNumIncompleteTasks[jobId] = 0
				}
				td := this.addTaskToJob(pod, jd)
				jobNumIncompleteTasks[jobId]++
				PodToTD[pod.Identifier] = td
				TaskIDToPod[td.GetUid()] = pod.Identifier
				taskDescription := &firmament.TaskDescription{
					TaskDescriptor: td,
					JobDescriptor:  jd,
				}
				firmament.TaskSubmitted(this.fc, taskDescription)
			case PodSucceeded:
				td, ok := PodToTD[pod.Identifier]
				if !ok {
					glog.Fatalf("Pod %v does not exist", pod.Identifier)
				}
				firmament.TaskCompleted(this.fc, &firmament.TaskUID{TaskUid: td.Uid})
				delete(PodToTD, pod.Identifier)
				delete(TaskIDToPod, td.GetUid())
				jobId := this.generateJobID(pod.Identifier.Name)
				jobNumIncompleteTasks[jobId]--
				if jobNumIncompleteTasks[jobId] == 0 {
					// All tasks completed => Job is completed.
					delete(jobNumIncompleteTasks, jobId)
					delete(jobIDToJD, jobId)
				} else {
					// XXX(ionel): We currently leave the deleted task in the job's
					// root/spawned task list. We do not delete it in case we need
					// to consistently regenerate taskUids out of job name and tasks'
					// position within the Spawned list.
				}
			case PodDeleted:
				td, ok := PodToTD[pod.Identifier]
				if !ok {
					glog.Fatalf("Pod %s does not exist", pod.Identifier)
				}
				firmament.TaskRemoved(this.fc, &firmament.TaskUID{TaskUid: td.Uid})
				delete(PodToTD, pod.Identifier)
				delete(TaskIDToPod, td.GetUid())
				jobId := this.generateJobID(pod.Identifier.Name)
				jobNumIncompleteTasks[jobId]--
				if jobNumIncompleteTasks[jobId] == 0 {
					// Clean state because the job doesn't have any tasks left.
					delete(jobNumIncompleteTasks, jobId)
					delete(jobIDToJD, jobId)
				} else {
					// XXX(ionel): We currently leave the deleted task in the job's
					// root/spawned task list. We do not delete it in case we need
					// to consistently regenerate taskUids out of job name and tasks'
					// position within the Spawned list.
				}
			case PodFailed:
				td, ok := PodToTD[pod.Identifier]
				if !ok {
					glog.Fatalf("Pod %s does not exist", pod.Identifier)
				}
				firmament.TaskFailed(this.fc, &firmament.TaskUID{TaskUid: td.Uid})
				// TODO(ionel): We do not delete the task from podToTD and taskIDToPod in case the task may be rescheduled. Check how K8s restart policies work and decide what to do here.
				// TODO(ionel): Should we delete the task from JD's spawned field?
			case PodRunning:
				// We don't have to do anything.
			case PodUnknown:
				glog.Errorf("Pod %s in unknown state", pod.Identifier)
				// TODO(ionel): Handle Unknown case.
			default:
				glog.Fatalf("Pod %v in unexpected state %v", pod.Identifier, pod.State)
			}
			glog.Info("Pod data received from the queue", pod)
			defer this.podWorkQueue.Done(key)
		}()
	}
}

func (this *PodWatcher) createNewJob(jobName string) *firmament.JobDescriptor {
	jobDesc := &firmament.JobDescriptor{
		Uuid:  this.generateJobID(jobName),
		Name:  jobName,
		State: firmament.JobDescriptor_CREATED,
	}
	return jobDesc
}

func (this *PodWatcher) addTaskToJob(pod *Pod, jd *firmament.JobDescriptor) *firmament.TaskDescriptor {
	task := &firmament.TaskDescriptor{
		Name:  pod.Identifier.UniqueName(),
		State: firmament.TaskDescriptor_CREATED,
		JobId: jd.Uuid,
		ResourceRequest: &firmament.ResourceVector{
			CpuCores: float32(pod.CpuRequest),
			RamCap:   uint64(pod.MemRequestKb),
		},
		// TODO(ionel): Populate LabelSelector.
	}
	// Add labels.
	for label, value := range pod.Labels {
		task.Labels = append(task.Labels,
			&firmament.Label{
				Key:   label,
				Value: value,
			})
	}
	if jd.RootTask == nil {
		task.Uid = this.generateTaskID(jd.Name, 0)
		jd.RootTask = task
	} else {
		task.Uid = this.generateTaskID(jd.Name, len(jd.RootTask.Spawned)+1)
		jd.RootTask.Spawned = append(jd.RootTask.Spawned, task)
	}
	return task
}

func (this *PodWatcher) generateJobID(seed string) string {
	return GenerateUUID(seed)
}

func (this *PodWatcher) generateTaskID(jdUid string, taskNum int) uint64 {
	return HashCombine(jdUid, taskNum)
}
