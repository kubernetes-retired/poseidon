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
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
	"github.com/kubernetes-sigs/poseidon/pkg/firmament"
)

type TestPodWatchObj struct {
	firmamentClient *firmament.MockFirmamentSchedulerClient
	kubeClient      *fake.Clientset
	kubeVerMajor    int
	kubeVerMinor    int
	schedulerName   string
	mockCtrl        *gomock.Controller
}

// initializePodObj initializes and returns TestPodWatchObj
func initializePodObj(t *testing.T) *TestPodWatchObj {
	testObj := &TestPodWatchObj{}
	testObj.mockCtrl = gomock.NewController(t)
	testObj.firmamentClient = firmament.NewMockFirmamentSchedulerClient(testObj.mockCtrl)
	testObj.kubeClient = &fake.Clientset{}
	testObj.kubeVerMajor = 1
	testObj.kubeVerMinor = 6
	testObj.schedulerName = "poseidon"
	return testObj
}

func ChangePodPhase(pod *v1.Pod, newPhase string) *v1.Pod {
	newPod := *pod
	newPod.Status = v1.PodStatus{
		Phase: GetPodPhase(newPhase),
	}
	return &newPod
}

func ChangePodCpuAndMemRequest(pod *v1.Pod, newCpu, newMem string) *v1.Pod {
	newPod := *pod
	newPod.Spec.Containers = []v1.Container{
		{
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse(newCpu),
					v1.ResourceMemory: resource.MustParse(newMem),
				},
			},
		},
	}
	return &newPod
}

func BuildPod(namespace, podName string,
	podLabels map[string]string,
	phase v1.PodPhase,
	requestCPU, requestMem string,
	deletionTime *metav1.Time,
	ownerRef string) *v1.Pod {

	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              podName,
			Namespace:         namespace,
			Labels:            podLabels,
			DeletionTimestamp: deletionTime,
			UID:               types.UID(ownerRef),
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							v1.ResourceCPU:    resource.MustParse(requestCPU),
							v1.ResourceMemory: resource.MustParse(requestMem),
						},
					},
				},
			},
		},
		Status: v1.PodStatus{
			Phase: phase,
		},
	}
}

// This will return a valid pod phase
func GetPodPhase(strPhase string) v1.PodPhase {
	var podPhase v1.PodPhase
	switch strPhase {
	case "Pending":
		podPhase = v1.PodPending
	case "Running":
		podPhase = v1.PodRunning
	case "Succeeded":
		podPhase = v1.PodSucceeded
	case "Failed":
		podPhase = v1.PodFailed
	}
	return podPhase
}

//get the meta key for the pod
func GetKey(pod *v1.Pod, t *testing.T) string {
	key, err := cache.MetaNamespaceKeyFunc(pod)
	if err != nil {
		t.Fatal("Cannot get a valid key for pod")
	}
	return key
}

// TestNewPodWatcher tests for different k8s versions for NewPodWatcher()
func TestNewPodWatcher(t *testing.T) {
	testObj := initializePodObj(t)
	defer testObj.mockCtrl.Finish()

	// for default k8s 1.6
	podWatch := NewPodWatcher(testObj.kubeVerMajor, testObj.kubeVerMinor, testObj.schedulerName, testObj.kubeClient, testObj.firmamentClient)
	t.Logf("Pod watcher for v1.6=%v", podWatch)

	// for k8s 1.5
	testObj.kubeVerMajor = 1
	testObj.kubeVerMinor = 5
	podWatch = NewPodWatcher(testObj.kubeVerMajor, testObj.kubeVerMinor, testObj.schedulerName, testObj.kubeClient, testObj.firmamentClient)
	t.Logf("Pod watcher for v1.5=%v", podWatch)

}

func TestPodWatcher_enqueuePodAddition(t *testing.T) {
	var empty map[string]string
	fakeNow := metav1.Now()
	keychan := make(chan interface{})
	itemschan := make(chan []interface{})
	fakeOwnerRef := "abcdfe12345"

	var testData = []struct {
		pod      *v1.Pod
		expected *Pod
	}{
		{
			pod: BuildPod("Poseidon-Namespace", "Pod1", empty, GetPodPhase("Pending"), "2", "1024", &fakeNow, fakeOwnerRef),
			expected: &Pod{
				State: PodPending,
				Identifier: PodIdentifier{
					Name:      "Pod1",
					Namespace: "Poseidon-Namespace",
				},
				CpuRequest:   2000,
				MemRequestKb: 1,
				OwnerRef:     fakeOwnerRef,
			},
		},
		{
			pod: BuildPod("Poseidon-Namespace", "Pod2", empty, GetPodPhase("Running"), "2", "1024", &fakeNow, fakeOwnerRef),
			expected: &Pod{
				State: PodRunning,
				Identifier: PodIdentifier{
					Name:      "Pod2",
					Namespace: "Poseidon-Namespace",
				},
				CpuRequest:   2000,
				MemRequestKb: 1,
				OwnerRef:     fakeOwnerRef,
			},
		},
		{
			pod: BuildPod("Poseidon-Namespace", "Pod3", empty, GetPodPhase("Succeeded"), "2", "1024", &fakeNow, fakeOwnerRef),
			expected: &Pod{
				State: PodSucceeded,
				Identifier: PodIdentifier{
					Name:      "Pod3",
					Namespace: "Poseidon-Namespace",
				},
				CpuRequest:   2000,
				MemRequestKb: 1,
				OwnerRef:     fakeOwnerRef,
			},
		},
		{
			pod: BuildPod("Poseidon-Namespace", "Pod4", empty, GetPodPhase("Failed"), "2", "1024", &fakeNow, fakeOwnerRef),
			expected: &Pod{
				State: PodFailed,
				Identifier: PodIdentifier{
					Name:      "Pod4",
					Namespace: "Poseidon-Namespace",
				},
				CpuRequest:   2000,
				MemRequestKb: 1,
				OwnerRef:     fakeOwnerRef,
			},
		},
	}

	testObj := initializePodObj(t)
	defer testObj.mockCtrl.Finish()
	podWatch := NewPodWatcher(testObj.kubeVerMajor, testObj.kubeVerMinor, testObj.schedulerName, testObj.kubeClient, testObj.firmamentClient)

	for _, podData := range testData {
		key := GetKey(podData.pod, t)
		podWatch.enqueuePodAddition(key, podData.pod)
		go func() {
			newkey, newitems, _ := podWatch.podWorkQueue.Get()
			keychan <- newkey
			itemschan <- newitems
		}()
		waitTimer := time.NewTimer(time.Second * 2)
		select {
		case <-waitTimer.C:
		case newkey := <-keychan:
			newitems := <-itemschan
			for _, item := range newitems {
				if newItem, ok := item.(*Pod); ok {
					if !(reflect.DeepEqual(podData.expected, newItem) && reflect.DeepEqual(newkey, key)) {
						t.Error("expected ", podData.expected, "got ", newItem)
					}
				}
			}
		}
	}
}

func TestPodWatcher_podWorker(t *testing.T) {

	var empty map[string]string
	fakeNow := metav1.Now()
	fakeOwnerRef := "abcdfe12345"

	var testData = []struct {
		pod *v1.Pod
	}{
		{
			pod: BuildPod("Poseidon-Namespace", "Pod1", empty, GetPodPhase("Pending"), "2", "1024", &fakeNow, fakeOwnerRef),
		},
		{
			pod: BuildPod("Poseidon-Namespace", "Pod2", empty, GetPodPhase("Pending"), "2", "1024", &fakeNow, fakeOwnerRef),
		},
		{
			pod: BuildPod("Poseidon-Namespace", "Pod3", empty, GetPodPhase("Pending"), "2", "1024", &fakeNow, fakeOwnerRef),
		},
		{
			pod: BuildPod("Poseidon-Namespace", "Pod4", empty, GetPodPhase("Pending"), "2", "1024", &fakeNow, fakeOwnerRef),
		},
		{
			pod: BuildPod("Poseidon-Namespace", "Pod5", empty, GetPodPhase("Pending"), "2", "1024", &fakeNow, fakeOwnerRef),
		},
	}

	testObj := initializePodObj(t)
	defer testObj.mockCtrl.Finish()
	podWatch := NewPodWatcher(testObj.kubeVerMajor, testObj.kubeVerMinor, testObj.schedulerName, testObj.kubeClient, testObj.firmamentClient)
	for index, podData := range testData {
		switch index {
		case 0:
			//TaskSubmitted case
			key := GetKey(podData.pod, t)
			podWatch.enqueuePodAddition(key, podData.pod)
		case 1:
			//TaskRemoved case
			key := GetKey(podData.pod, t)
			podWatch.enqueuePodAddition(key, podData.pod)
			podData.pod = ChangePodPhase(testData[index].pod, "Failed")
			podWatch.enqueuePodDeletion(key, podData.pod)
		case 2:
			//TaskCompleted case
			key := GetKey(podData.pod, t)
			podWatch.enqueuePodAddition(key, podData.pod)
			newPod := ChangePodPhase(testData[index].pod, "Succeeded")
			podWatch.enqueuePodUpdate(key, podData.pod, newPod)
		case 3:
			//TaskSubmitted case
			key := GetKey(podData.pod, t)
			podWatch.enqueuePodAddition(key, podData.pod)
			newPod := ChangePodCpuAndMemRequest(testData[index].pod, "3", "3072")
			podWatch.enqueuePodUpdate(key, podData.pod, newPod)
		case 4:
			//TaskFailed case
			key := GetKey(podData.pod, t)
			podWatch.enqueuePodAddition(key, podData.pod)
			newPod := ChangePodPhase(testData[index].pod, "Failed")
			podWatch.enqueuePodUpdate(key, podData.pod, newPod)
		}
	}
	gomock.InOrder(
		//case 0
		testObj.firmamentClient.EXPECT().TaskSubmitted(gomock.Any(), gomock.Any()).Return(
			&firmament.TaskSubmittedResponse{Type: firmament.TaskReplyType_TASK_SUBMITTED_OK}, nil),

		//case 1
		testObj.firmamentClient.EXPECT().TaskSubmitted(gomock.Any(), gomock.Any()).Return(
			&firmament.TaskSubmittedResponse{Type: firmament.TaskReplyType_TASK_SUBMITTED_OK}, nil),
		testObj.firmamentClient.EXPECT().TaskRemoved(gomock.Any(), gomock.Any()).Return(
			&firmament.TaskRemovedResponse{Type: firmament.TaskReplyType_TASK_REMOVED_OK}, nil),

		//case 2
		testObj.firmamentClient.EXPECT().TaskSubmitted(gomock.Any(), gomock.Any()).Return(
			&firmament.TaskSubmittedResponse{Type: firmament.TaskReplyType_TASK_SUBMITTED_OK}, nil),
		testObj.firmamentClient.EXPECT().TaskCompleted(gomock.Any(), gomock.Any()).Return(
			&firmament.TaskCompletedResponse{Type: firmament.TaskReplyType_TASK_COMPLETED_OK}, nil),

		//case 3
		testObj.firmamentClient.EXPECT().TaskSubmitted(gomock.Any(), gomock.Any()).Return(
			&firmament.TaskSubmittedResponse{Type: firmament.TaskReplyType_TASK_SUBMITTED_OK}, nil),
		testObj.firmamentClient.EXPECT().TaskSubmitted(gomock.Any(), gomock.Any()).Return(
			&firmament.TaskSubmittedResponse{Type: firmament.TaskReplyType_TASK_SUBMITTED_OK}, nil),

		//case 4
		testObj.firmamentClient.EXPECT().TaskSubmitted(gomock.Any(), gomock.Any()).Return(
			&firmament.TaskSubmittedResponse{Type: firmament.TaskReplyType_TASK_SUBMITTED_OK}, nil),
		testObj.firmamentClient.EXPECT().TaskFailed(gomock.Any(), gomock.Any()).Return(
			&firmament.TaskFailedResponse{Type: firmament.TaskReplyType_TASK_FAILED_OK}, nil),

		//testObj.firmamentClient.EXPECT().TaskUpdated(gomock.Any(), gomock.Any()).Return(
		//	&firmament.TaskUpdatedResponse{Type: firmament.TaskReplyType_TASK_UPDATED_OK}, nil),
	)
	go podWatch.podWorker()
	newTimer := time.NewTimer(time.Second * 2)
	<-newTimer.C
}
