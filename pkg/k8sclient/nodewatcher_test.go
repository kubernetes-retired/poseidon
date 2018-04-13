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

	"github.com/kubernetes-sigs/poseidon/pkg/firmament"

	"github.com/golang/mock/gomock"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
)

type TestNodeWatchObj struct {
	firmamentClient *firmament.MockFirmamentSchedulerClient
	kubeClient      *fake.Clientset
	mockCtrl        *gomock.Controller
}

// initializeNodeObj initializes and returns TestNodeWatchObj
func initializeNodeObj(t *testing.T) *TestNodeWatchObj {
	mockCtrl := gomock.NewController(t)
	testObj := &TestNodeWatchObj{}
	testObj.firmamentClient = firmament.NewMockFirmamentSchedulerClient(mockCtrl)
	testObj.kubeClient = &fake.Clientset{}
	testObj.mockCtrl = mockCtrl
	return testObj
}

// BuildFirmamentResourceDescriptor constructs a ResourceTopologyNodeDescriptor
func BuildFirmamentResourceDescriptor(
	uuid, friendlyName string,
	cpuCores float32,
	ramCap uint64,
	coreOneUuid, coreOnefriendlyName string) *firmament.ResourceTopologyNodeDescriptor {

	return &firmament.ResourceTopologyNodeDescriptor{
		ResourceDesc: &firmament.ResourceDescriptor{
			Uuid:         uuid,
			FriendlyName: friendlyName,
			State:        firmament.ResourceDescriptor_RESOURCE_IDLE,
			Type:         firmament.ResourceDescriptor_RESOURCE_MACHINE,
			ResourceCapacity: &firmament.ResourceVector{
				CpuCores: cpuCores,
				RamCap:   ramCap,
			},
		},
		Children: []*firmament.ResourceTopologyNodeDescriptor{
			{
				ResourceDesc: &firmament.ResourceDescriptor{
					Uuid:         coreOneUuid,
					FriendlyName: coreOnefriendlyName,
					State:        firmament.ResourceDescriptor_RESOURCE_IDLE,
					ResourceCapacity: &firmament.ResourceVector{
						CpuCores: cpuCores,
						RamCap:   ramCap,
					},
				},
				ParentId: uuid,
			},
		},
	}
}

//BuildNode build a v1.Node struct to test
func BuildNode(name, requestCPU, requestMem string,
	nodeLabel map[string]string,
	nodeConditions []v1.NodeCondition,
	nodeIsSchedulable bool) *v1.Node {
	fakeNow := metav1.Now()
	newnode := &v1.Node{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			CreationTimestamp: fakeNow,
			Labels:            nodeLabel,
		},
		Spec: v1.NodeSpec{
			Unschedulable: nodeIsSchedulable,
		},
		Status: v1.NodeStatus{
			Conditions: nodeConditions,
			Capacity: v1.ResourceList{
				v1.ResourceName(v1.ResourceCPU):    resource.MustParse(requestCPU),
				v1.ResourceName(v1.ResourceMemory): resource.MustParse(requestMem),
			},
		},
	}
	return newnode
}

// TestNewNodeWatcher tests for NewNodeWatcher() function
func TestNewNodeWatcher(t *testing.T) {
	testObj := initializeNodeObj(t)
	defer testObj.mockCtrl.Finish()
	nodeWatch := NewNodeWatcher(testObj.kubeClient, testObj.firmamentClient)
	t.Logf("Node watcher=%v", nodeWatch)
}

func TestNodeWatcher_getReadyAndOutOfDiskConditions(t *testing.T) {
	var testData = []struct {
		node     *v1.Node
		expected []bool
	}{
		{
			node: BuildNode("node0", "10", "1024", nil, []v1.NodeCondition{
				{
					Type:   v1.NodeReady,
					Status: v1.ConditionFalse,
				},
			}, false),
			expected: []bool{false, false},
		},
		{
			node: BuildNode("node0", "10", "1024", nil, []v1.NodeCondition{
				{
					Type:   v1.NodeReady,
					Status: v1.ConditionFalse,
				},
			}, false),
			expected: []bool{false, false},
		},
		{
			node: BuildNode("node0", "10", "1024", nil, []v1.NodeCondition{
				{
					Type:   v1.NodeOutOfDisk,
					Status: v1.ConditionTrue,
				},
			}, false),
			expected: []bool{false, true},
		},
		{
			node: BuildNode("node0", "10", "1024", nil, []v1.NodeCondition{
				{
					Type:   v1.NodeReady,
					Status: v1.ConditionTrue,
				},
			}, false),
			expected: []bool{true, false},
		},
	}
	testObj := initializeNodeObj(t)
	defer testObj.mockCtrl.Finish()

	nodeWatch := NewNodeWatcher(testObj.kubeClient, testObj.firmamentClient)

	for _, testValue := range testData {
		resultOne, resultTwo := nodeWatch.getReadyAndOutOfDiskConditions(testValue.node)
		if resultOne != testValue.expected[0] || resultTwo != testValue.expected[1] {
			t.Error("expected ", testValue.expected[0], testValue.expected[1], "got ", resultOne, resultTwo)
		}
	}

}

func TestNodeWatcher_parseNode(t *testing.T) {

	var testData = []struct {
		node     *v1.Node
		phase    NodePhase
		expected *Node
	}{
		{
			node: BuildNode("node0", "10", "10000000000", nil, []v1.NodeCondition{
				{
					Type:   v1.NodeOutOfDisk,
					Status: v1.ConditionFalse,
				},
			}, false),
			phase: NodeAdded,
			expected: &Node{
				Hostname:         "node0",
				Phase:            NodeAdded,
				IsReady:          false,
				IsOutOfDisk:      false,
				CpuCapacity:      10000,
				CpuAllocatable:   0,
				MemCapacityKb:    9765625,
				MemAllocatableKb: 0,
				Labels:           nil,
				Annotations:      nil,
			},
		},
	}

	testObj := initializeNodeObj(t)
	defer testObj.mockCtrl.Finish()
	nodeWatch := NewNodeWatcher(testObj.kubeClient, testObj.firmamentClient)

	for _, testValue := range testData {
		result := nodeWatch.parseNode(testValue.node, testValue.phase)
		if !reflect.DeepEqual(result, testValue.expected) {
			t.Error("expected ", testValue.expected, "got ", result)
		}
	}
}

func TestNodeWatcher_enqueueNodeAddition(t *testing.T) {
	var testData = []struct {
		node     *v1.Node
		phase    NodePhase
		expected *Node
	}{
		{
			node: BuildNode("node0", "10", "10000000000", nil, []v1.NodeCondition{
				{
					Type:   v1.NodeOutOfDisk,
					Status: v1.ConditionFalse,
				},
			}, false),
			phase: NodeAdded,
			expected: &Node{
				Hostname:         "node0",
				Phase:            NodeAdded,
				IsReady:          false,
				IsOutOfDisk:      false,
				CpuCapacity:      10000,
				CpuAllocatable:   0,
				MemCapacityKb:    9765625,
				MemAllocatableKb: 0,
				Labels:           nil,
				Annotations:      nil,
			},
		},
		{
			node: BuildNode("node1", "10", "10000000000", nil, []v1.NodeCondition{
				{
					Type:   v1.NodeOutOfDisk,
					Status: v1.ConditionFalse,
				},
			}, true),
			phase: NodeAdded,
			expected: &Node{
				Hostname:         "node1",
				Phase:            NodeAdded,
				IsReady:          false,
				IsOutOfDisk:      false,
				CpuCapacity:      10000,
				CpuAllocatable:   0,
				MemCapacityKb:    9765625,
				MemAllocatableKb: 0,
				Labels:           nil,
				Annotations:      nil,
			},
		},
	}
	testObj := initializeNodeObj(t)
	defer testObj.mockCtrl.Finish()

	nodeWatch := NewNodeWatcher(testObj.kubeClient, testObj.firmamentClient)
	keychan := make(chan interface{})
	itemschan := make(chan []interface{})

	for _, testValue := range testData {
		key, err := cache.MetaNamespaceKeyFunc(testValue.node)
		if err != nil {
			t.Error("AddFunc: error getting key ", err)
		}
		nodeWatch.enqueueNodeAddition(key, testValue.node)
		go func() {
			newkey, newitems, _ := nodeWatch.nodeWorkQueue.Get()
			keychan <- newkey
			itemschan <- newitems
		}()
		waitTimer := time.NewTimer(time.Second * 5)
		select {
		case <-waitTimer.C:
		case newkey := <-keychan:
			newitem := <-itemschan

			for _, item := range newitem {
				if nextItem, ok := item.(*Node); ok {
					if !(reflect.DeepEqual(testValue.expected, nextItem) && reflect.DeepEqual(newkey, key)) {
						t.Error("expected ", testValue.node, "got ", nextItem, "\n", newkey)
					}
				} else {
					t.Error("Type assertion failed", reflect.TypeOf(newitem))
				}
			}
		}
	}
}

func TestNodeWatcher_createResourceTopologyForNode(t *testing.T) {
	var testData = []struct {
		node     *Node
		expected *firmament.ResourceTopologyNodeDescriptor
	}{
		{
			node: &Node{
				Hostname:         "node0",
				Phase:            NodeAdded,
				IsReady:          false,
				IsOutOfDisk:      false,
				CpuCapacity:      1000,
				CpuAllocatable:   0,
				MemCapacityKb:    9765625,
				MemAllocatableKb: 0,
				Labels:           nil,
				Annotations:      nil,
			},
			expected: BuildFirmamentResourceDescriptor("e8107a51-344b-4946-963c-6a4f4eb35f0c",
				"node0",
				1000,
				9765625,
				"a03e6733-dc70-4da5-91ea-bded4bbee951",
				"node0_PU #0"),
		},
		{
			node: &Node{
				Hostname:         "node1",
				Phase:            NodeAdded,
				IsReady:          true,
				IsOutOfDisk:      false,
				CpuCapacity:      1000,
				CpuAllocatable:   0,
				MemCapacityKb:    2048,
				MemAllocatableKb: 0,
				Labels:           nil,
				Annotations:      nil,
			},
			expected: BuildFirmamentResourceDescriptor("3e4eff18-02e5-4066-aac1-e49c5d4b9766",
				"node1",
				1000,
				2048,
				"6b45f181-d411-4bf1-9755-89a35115cdd7",
				"node1_PU #0"),
		},
	}
	testObj := initializeNodeObj(t)
	nodeWatch := NewNodeWatcher(testObj.kubeClient, testObj.firmamentClient)
	for _, testValue := range testData {
		got := nodeWatch.createResourceTopologyForNode(testValue.node)
		if !reflect.DeepEqual(got, testValue.expected) {
			t.Error("Error expected ", testValue.expected, " got ", got)
		}
	}
}

func TestNodeWatcher_nodeWorker(t *testing.T) {
	fakeNow := metav1.Now()
	var testData = []struct {
		node  *v1.Node
		phase NodePhase
	}{
		{
			node: BuildNode("node0", "1", "10000000000", nil, []v1.NodeCondition{
				{
					Type:               v1.NodeReady,
					Status:             v1.ConditionFalse,
					LastHeartbeatTime:  fakeNow,
					LastTransitionTime: fakeNow,
				},
			}, false),
			phase: NodeAdded,
		},
		{
			node: BuildNode("node0", "1", "10000000000", nil, []v1.NodeCondition{
				{
					Type:               v1.NodeReady,
					Status:             v1.ConditionFalse,
					LastHeartbeatTime:  fakeNow,
					LastTransitionTime: fakeNow,
				},
			}, false),
			phase: NodeDeleted,
		},
		{
			node: BuildNode("node0", "1", "10000000000", nil, []v1.NodeCondition{
				{
					Type:               v1.NodeReady,
					Status:             v1.ConditionFalse,
					LastHeartbeatTime:  metav1.Now(),
					LastTransitionTime: metav1.Now(),
				},
				{
					Type:               v1.NodeOutOfDisk,
					Status:             v1.ConditionTrue,
					LastHeartbeatTime:  metav1.Now(),
					LastTransitionTime: metav1.Now(),
				},
			}, false),
			phase: NodeFailed,
		},
		{
			node: BuildNode("node0", "1", "10000000000", map[string]string{
				"name": "foo",
				"type": "production",
			}, []v1.NodeCondition{
				{
					Type:               v1.NodeReady,
					Status:             v1.ConditionFalse,
					LastHeartbeatTime:  metav1.Now(),
					LastTransitionTime: metav1.Now(),
				},
			}, false),
			phase: NodeUpdated,
		},
	}

	testObj := initializeNodeObj(t)
	defer testObj.mockCtrl.Finish()

	gomock.InOrder(
		testObj.firmamentClient.EXPECT().NodeAdded(gomock.Any(), gomock.Any()).Return(
			&firmament.NodeAddedResponse{Type: firmament.NodeReplyType_NODE_ADDED_OK}, nil),
		testObj.firmamentClient.EXPECT().NodeRemoved(gomock.Any(), gomock.Any()).Return(
			&firmament.NodeRemovedResponse{Type: firmament.NodeReplyType_NODE_REMOVED_OK}, nil),

		testObj.firmamentClient.EXPECT().NodeAdded(gomock.Any(), gomock.Any()).Return(
			&firmament.NodeAddedResponse{Type: firmament.NodeReplyType_NODE_ADDED_OK}, nil),
		testObj.firmamentClient.EXPECT().NodeFailed(gomock.Any(), gomock.Any()).Return(
			&firmament.NodeFailedResponse{Type: firmament.NodeReplyType_NODE_FAILED_OK}, nil),

		testObj.firmamentClient.EXPECT().NodeAdded(gomock.Any(), gomock.Any()).Return(
			&firmament.NodeAddedResponse{Type: firmament.NodeReplyType_NODE_ADDED_OK}, nil),
		testObj.firmamentClient.EXPECT().NodeUpdated(gomock.Any(), gomock.Any()).Return(
			&firmament.NodeUpdatedResponse{Type: firmament.NodeReplyType_NODE_UPDATED_OK}, nil),
	)
	nodeWatch := NewNodeWatcher(testObj.kubeClient, testObj.firmamentClient)
	for index, testValue := range testData {

		key, err := cache.MetaNamespaceKeyFunc(testValue.node)
		if err != nil {
			t.Error("AddFunc: error getting key ", err)
		}

		switch index {
		case 0:
			nodeWatch.enqueueNodeAddition(key, testValue.node)
		case 1:
			nodeWatch.enqueueNodeDeletion(key, testValue.node)
		case 2:
			//testValue.node.Status.Conditions =
			//add a new node and then update the node
			// Adding a new node
			key, err := cache.MetaNamespaceKeyFunc(testData[0].node)
			if err != nil {
				t.Error("AddFunc: error getting key ", err)
			}
			nodeWatch.enqueueNodeAddition(key, testData[0].node)
			nodeWatch.enqueueNodeUpdate(key, testData[0].node, testValue.node)
		case 3:
			//add a new node and then update the node

			// Adding a new node
			key, err := cache.MetaNamespaceKeyFunc(testData[0].node)
			if err != nil {
				t.Error("AddFunc: error getting key ", err)
			}
			nodeWatch.enqueueNodeAddition(key, testData[0].node)
			nodeWatch.enqueueNodeUpdate(key, testData[0].node, testValue.node)
		default:
		}
	}
	go nodeWatch.nodeWorker()
	timer1 := time.NewTimer(time.Second * 2)
	<-timer1.C
	nodeWatch.nodeWorkQueue.ShutDown()
}
