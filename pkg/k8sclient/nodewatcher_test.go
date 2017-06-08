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
	"bytes"
	"encoding/json"
	//"os"
	"reflect"
	"testing"
	"time"

	"github.com/camsas/poseidon/pkg/firmament"
	"github.com/camsas/poseidon/pkg/mock_firmament"
	"github.com/golang/mock/gomock"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/tools/cache"
)

type TestNodeWatchObj struct {
	firmamentClient *mock_firmament.MockFirmamentSchedulerClient
	kubeClient      *fake.Clientset
	mockCtrl        *gomock.Controller
}

// initializeNodeObj initializes and returns TestNodeWatchObj
func initializeNodeObj(t *testing.T) *TestNodeWatchObj {
	mockCtrl := gomock.NewController(t)
	//defer mockCtrl.Finish()
	testObj := &TestNodeWatchObj{}
	testObj.firmamentClient = fakeFirmamentClient(mockCtrl)
	testObj.kubeClient = &fake.Clientset{}
	testObj.mockCtrl = mockCtrl
	return testObj
}

// fakeFirmamentClient creates a fake firmament client
func fakeFirmamentClient(t *gomock.Controller) *mock_firmament.MockFirmamentSchedulerClient {
	fc := mock_firmament.NewMockFirmamentSchedulerClient(t)
	return fc
}

// TestNewNodeWatcher tests for NewNodeWatcher() function
func TestNewNodeWatcher(t *testing.T) {
	testObj := initializeNodeObj(t)
	nodeWatch := NewNodeWatcher(testObj.kubeClient, testObj.firmamentClient)
	t.Logf("Node watcher=", nodeWatch)
}

func TestNodeWatchergetReadyAndOutOfDiskConditions(t *testing.T) {
	fakeNow := metav1.Date(2015, 1, 1, 12, 0, 0, 0, time.UTC)
	var testData = []struct {
		node     *v1.Node
		expected []bool
	}{
		{
			node: &v1.Node{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:              "node0",
					CreationTimestamp: fakeNow,
				},
				Spec: v1.NodeSpec{},
				Status: v1.NodeStatus{
					Conditions: []v1.NodeCondition{
						{
							Type:               v1.NodeReady,
							Status:             v1.ConditionFalse,
							LastHeartbeatTime:  metav1.Date(2015, 1, 1, 12, 0, 0, 0, time.UTC),
							LastTransitionTime: metav1.Date(2015, 1, 1, 12, 0, 0, 0, time.UTC),
						},
					},
				},
			},
			expected: []bool{false, false},
		},
		{
			node: &v1.Node{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:              "node0",
					CreationTimestamp: fakeNow,
				},
				Spec: v1.NodeSpec{},
				Status: v1.NodeStatus{
					Conditions: []v1.NodeCondition{
						{
							Type:               v1.NodeOutOfDisk,
							Status:             v1.ConditionFalse,
							LastHeartbeatTime:  metav1.Date(2015, 1, 1, 12, 0, 0, 0, time.UTC),
							LastTransitionTime: metav1.Date(2015, 1, 1, 12, 0, 0, 0, time.UTC),
						},
					},
				},
			},
			expected: []bool{false, false},
		},
		{
			node: &v1.Node{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:              "node0",
					CreationTimestamp: fakeNow,
				},
				Spec: v1.NodeSpec{},
				Status: v1.NodeStatus{
					Conditions: []v1.NodeCondition{
						{
							Type:               v1.NodeOutOfDisk,
							Status:             v1.ConditionTrue,
							LastHeartbeatTime:  metav1.Date(2015, 1, 1, 12, 0, 0, 0, time.UTC),
							LastTransitionTime: metav1.Date(2015, 1, 1, 12, 0, 0, 0, time.UTC),
						},
					},
				},
			},
			expected: []bool{false, true},
		},
		{
			node: &v1.Node{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:              "node0",
					CreationTimestamp: fakeNow,
				},
				Spec: v1.NodeSpec{},
				Status: v1.NodeStatus{
					Conditions: []v1.NodeCondition{
						{
							Type:               v1.NodeReady,
							Status:             v1.ConditionTrue,
							LastHeartbeatTime:  metav1.Date(2015, 1, 1, 12, 0, 0, 0, time.UTC),
							LastTransitionTime: metav1.Date(2015, 1, 1, 12, 0, 0, 0, time.UTC),
						},
					},
				},
			},
			expected: []bool{true, false},
		},
	}

	testObj := initializeNodeObj(t)
	nodeWatch := NewNodeWatcher(testObj.kubeClient, testObj.firmamentClient)

	for _, testValue := range testData {
		resultOne, resultTwo := nodeWatch.getReadyAndOutOfDiskConditions(testValue.node)
		if resultOne != testValue.expected[0] || resultTwo != testValue.expected[1] {
			t.Error("expected ", testValue.expected[0], testValue.expected[1], "got ", resultOne, resultTwo)
		}
	}

}

func TestNodeWatcherparseNode(t *testing.T) {

	fakeNow := metav1.Date(2015, 1, 1, 12, 0, 0, 0, time.UTC)
	var testData = []struct {
		node     *v1.Node
		phase    NodePhase
		expected *Node
	}{
		{
			node: &v1.Node{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:              "node0",
					CreationTimestamp: fakeNow,
				},
				Spec: v1.NodeSpec{},
				Status: v1.NodeStatus{
					Conditions: []v1.NodeCondition{
						{
							Type:               v1.NodeReady,
							Status:             v1.ConditionFalse,
							LastHeartbeatTime:  metav1.Date(2015, 1, 1, 12, 0, 0, 0, time.UTC),
							LastTransitionTime: metav1.Date(2015, 1, 1, 12, 0, 0, 0, time.UTC),
						},
					},
					Capacity: v1.ResourceList{
						v1.ResourceName(v1.ResourceCPU):    resource.MustParse("10"),
						v1.ResourceName(v1.ResourceMemory): resource.MustParse("10G"),
					},
				},
			},
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
	nodeWatch := NewNodeWatcher(testObj.kubeClient, testObj.firmamentClient)

	for _, testValue := range testData {
		result := nodeWatch.parseNode(testValue.node, testValue.phase)
		if !reflect.DeepEqual(result, testValue.expected) {
			t.Error("expected ", testValue.expected, "got ", result)
		}
	}

}

func TestNodeWatcher_enqueueNodeAddition(t *testing.T) {
	fakeNow := metav1.Date(2015, 1, 1, 12, 0, 0, 0, time.UTC)
	var testData = []struct {
		node     *v1.Node
		phase    NodePhase
		expected *Node
	}{
		{
			node: &v1.Node{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:              "node0",
					CreationTimestamp: fakeNow,
				},
				Spec: v1.NodeSpec{
					Unschedulable: false,
				},
				Status: v1.NodeStatus{
					Conditions: []v1.NodeCondition{
						{
							Type:               v1.NodeReady,
							Status:             v1.ConditionFalse,
							LastHeartbeatTime:  metav1.Date(2015, 1, 1, 12, 0, 0, 0, time.UTC),
							LastTransitionTime: metav1.Date(2015, 1, 1, 12, 0, 0, 0, time.UTC),
						},
					},
					Capacity: v1.ResourceList{
						v1.ResourceName(v1.ResourceCPU):    resource.MustParse("10"),
						v1.ResourceName(v1.ResourceMemory): resource.MustParse("10G"),
					},
				},
			},
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
			node: &v1.Node{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:              "node1",
					CreationTimestamp: fakeNow,
				},
				Spec: v1.NodeSpec{
					Unschedulable: true,
				},
				Status: v1.NodeStatus{
					Conditions: []v1.NodeCondition{
						{
							Type:               v1.NodeReady,
							Status:             v1.ConditionFalse,
							LastHeartbeatTime:  metav1.Date(2015, 1, 1, 12, 0, 0, 0, time.UTC),
							LastTransitionTime: metav1.Date(2015, 1, 1, 12, 0, 0, 0, time.UTC),
						},
					},
					Capacity: v1.ResourceList{
						v1.ResourceName(v1.ResourceCPU):    resource.MustParse("10"),
						v1.ResourceName(v1.ResourceMemory): resource.MustParse("10G"),
					},
				},
			},
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
	nodeWatch := NewNodeWatcher(testObj.kubeClient, testObj.firmamentClient)
	keychan := make(chan interface{})
	itemschan := make(chan interface{})

	waitTimer := time.NewTimer(time.Second * 2)

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

		select {
		case <-waitTimer.C:
		case newkey := <-keychan:
			newitems := <-itemschan
			if !reflect.DeepEqual(testValue.node, newitems) && !reflect.DeepEqual(newkey, key) {
				t.Error("expected ", testValue.node, "got ", newitems)
			}
		}
	}
}

func TestNodeWatcher_createResourceTopologyForNode(t *testing.T) {
	var testData = []struct {
		node     *Node
		expected string
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
			expected: `{
        "resource_desc": {
                "uuid": "e8107a51-344b-4946-963c-6a4f4eb35f0c",
                "friendly_name": "node0",
                "state": 1,
                "type": 6,
                "resource_capacity": {
                        "cpu_cores": 1000,
                        "ram_cap": 9765625
                }
        },
        "children": [
                {
                        "resource_desc": {
                                "uuid": "22e3e4c7-4c49-485e-8d41-7e83e2073a57",
                                "friendly_name": "node0_pu0",
                                "state": 1
                        },
                        "parent_id": "e8107a51-344b-4946-963c-6a4f4eb35f0c"
                }
        ]
}`,
		},
		{
			node: &Node{
				Hostname:         "node1",
				Phase:            NodeAdded,
				IsReady:          true,
				IsOutOfDisk:      false,
				CpuCapacity:      2000,
				CpuAllocatable:   0,
				MemCapacityKb:    2048,
				MemAllocatableKb: 0,
				Labels:           nil,
				Annotations:      nil,
			},
			expected: `{
        "resource_desc": {
                "uuid": "3e4eff18-02e5-4066-aac1-e49c5d4b9766",
                "friendly_name": "node1",
                "state": 1,
                "type": 6,
                "resource_capacity": {
                        "cpu_cores": 2000,
                        "ram_cap": 2048
                }
        },
        "children": [
                {
                        "resource_desc": {
                                "uuid": "c9fe9fd0-2a5b-4b47-a064-498da81c7aa1",
                                "friendly_name": "node1_pu0",
                                "state": 1
                        },
                        "parent_id": "3e4eff18-02e5-4066-aac1-e49c5d4b9766"
                },
                {
                        "resource_desc": {
                                "uuid": "6c810a7f-5d6b-4928-adaf-3a71d2694257",
                                "friendly_name": "node1_pu1",
                                "state": 1
                        },
                        "parent_id": "3e4eff18-02e5-4066-aac1-e49c5d4b9766"
                }
        ]
}`,
		},
	}

	testObj := initializeNodeObj(t)
	nodeWatch := NewNodeWatcher(testObj.kubeClient, testObj.firmamentClient)

	for _, testValue := range testData {

		got := nodeWatch.createResourceTopologyForNode(testValue.node)

		byt, _ := json.Marshal(got)
		var out bytes.Buffer
		json.Indent(&out, byt, "", "\t")
		newbyte := out.Bytes()

		//if !reflect.DeepEqual(string(newbyte[:len(newbyte)]), testValue.expected) {
		if string(newbyte[:]) == testValue.expected {
			t.Error("Error not same", string(newbyte[:]))
		}
	}
}

func TestNodeWatcher_nodeWorker(t *testing.T) {
	fakeNow := metav1.Date(2015, 1, 1, 12, 0, 0, 0, time.UTC)
	var testData = []struct {
		node     *v1.Node
		phase    NodePhase
		expected *Node
	}{
		{
			node: &v1.Node{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:              "node0",
					CreationTimestamp: fakeNow,
				},
				Spec: v1.NodeSpec{
					Unschedulable: false,
				},
				Status: v1.NodeStatus{
					Conditions: []v1.NodeCondition{
						{
							Type:               v1.NodeReady,
							Status:             v1.ConditionFalse,
							LastHeartbeatTime:  metav1.Date(2015, 1, 1, 12, 0, 0, 0, time.UTC),
							LastTransitionTime: metav1.Date(2015, 1, 1, 12, 0, 0, 0, time.UTC),
						},
					},
					Capacity: v1.ResourceList{
						v1.ResourceName(v1.ResourceCPU):    resource.MustParse("10"),
						v1.ResourceName(v1.ResourceMemory): resource.MustParse("10G"),
					},
				},
			},
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
			node: &v1.Node{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:              "node1",
					CreationTimestamp: fakeNow,
				},
				Spec: v1.NodeSpec{
					Unschedulable: false,
				},
				Status: v1.NodeStatus{
					Conditions: []v1.NodeCondition{
						{
							Type:               v1.NodeReady,
							Status:             v1.ConditionFalse,
							LastHeartbeatTime:  metav1.Date(2015, 1, 1, 12, 0, 0, 0, time.UTC),
							LastTransitionTime: metav1.Date(2015, 1, 1, 12, 0, 0, 0, time.UTC),
						},
					},
					Capacity: v1.ResourceList{
						v1.ResourceName(v1.ResourceCPU):    resource.MustParse("10"),
						v1.ResourceName(v1.ResourceMemory): resource.MustParse("10G"),
					},
				},
			},
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
	nodeWatch := NewNodeWatcher(testObj.kubeClient, testObj.firmamentClient)

	testObj.firmamentClient.EXPECT().NodeAdded(gomock.Any(), gomock.Any()).Return(&firmament.NodeAddedResponse{}, nil)
	testObj.firmamentClient.EXPECT().NodeAdded(gomock.Any(), gomock.Any()).Return(&firmament.NodeAddedResponse{}, nil)
	for _, testValue := range testData {

		key, err := cache.MetaNamespaceKeyFunc(testValue.node)
		if err != nil {
			t.Error("AddFunc: error getting key ", err)
		}
		nodeWatch.enqueueNodeAddition(key, testValue.node)

	}

	nodeWatch.nodeWorker()

}
