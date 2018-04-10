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

package stats

import (
	"github.com/golang/mock/gomock"
	//"golang.org/x/net/context"
	"k8s.io/poseidon/pkg/firmament"
	"reflect"
	"time"

	"testing"
)

/*func Test_ReceiveNodeStats(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

}*/

func BuildPodStats(name, namespace, hostname string) *PodStats {

	return &PodStats{
		Name:                name,
		Namespace:           namespace,
		Hostname:            hostname,
		CpuLimit:            1000,
		CpuRequest:          10,
		CpuUsage:            30,
		MemLimit:            2000,
		MemRequest:          50,
		MemUsage:            500,
		MemRss:              0,
		MemCache:            0,
		MemWorkingSet:       0,
		MemPageFaults:       1,
		MemPageFaultsRate:   1.0,
		MajorPageFaults:     0,
		MajorPageFaultsRate: 0.0,
		NetRx:               1000,
		NetRxErrors:         0,
		NetRxErrorsRate:     0.0,
		NetRxRate:           0.0,
		NetTx:               10,
		NetTxErrors:         0,
		NetTxErrorsRate:     0.0,
		NetTxRate:           1.0,
	}

}

func BuildFirmamentTaskStats(hostname string) *firmament.TaskStats {

	return &firmament.TaskStats{
		//TaskId:   12345678,
		Hostname: hostname,
		//Timestamp:           uint64(time.Now().UnixNano()),
		CpuLimit:            1000,
		CpuRequest:          10,
		CpuUsage:            30,
		MemLimit:            2000,
		MemRequest:          50,
		MemUsage:            500,
		MemRss:              0,
		MemCache:            0,
		MemWorkingSet:       0,
		MemPageFaults:       1,
		MemPageFaultsRate:   1.0,
		MajorPageFaults:     0,
		MajorPageFaultsRate: 0.0,
		NetRx:               1000,
		NetRxErrors:         0,
		NetRxErrorsRate:     0.0,
		NetRxRate:           0.0,
		NetTx:               10,
		NetTxErrors:         0.0,
		NetTxErrorsRate:     0.0,
		NetTxRate:           1.0,
	}
}

func BuildNodeStats(hostname string, timestamp uint64) *NodeStats {

	return &NodeStats{
		Hostname:       hostname,
		Timestamp:      timestamp,
		CpuAllocatable: 7000,
		CpuCapacity:    10000,
		CpuReservation: 1000.0,
		CpuUtilization: 2000.0,
		MemAllocatable: 700000,
		MemCapacity:    1000000,
		MemReservation: 200000.0,
		MemUtilization: 100000.0,
	}
}

func BuildResourceStats(hostname string, timestamp uint64) *firmament.ResourceStats {

	cpuStats := &firmament.CpuStats{
		CpuAllocatable: 7000,
		CpuCapacity:    10000,
		CpuReservation: 1000.0,
		CpuUtilization: 2000.0,
	}
	return &firmament.ResourceStats{
		Timestamp:      timestamp,
		CpusStats:      []*firmament.CpuStats{cpuStats},
		MemAllocatable: 700000,
		MemCapacity:    1000000,
		MemReservation: 200000.0,
		MemUtilization: 100000.0,
	}
}

func Test_convertPodStatsToTaskStats(t *testing.T) {
	var testData = []struct {
		podStats *PodStats
		expected *firmament.TaskStats
	}{
		{
			podStats: BuildPodStats("TestPod_One", "TestSpace", "localhost"),
			expected: BuildFirmamentTaskStats("localhost"),
		},
		{
			podStats: BuildPodStats("TestPod_One", "TestSpaceTwo", "127.0.0.1"),
			expected: BuildFirmamentTaskStats("127.0.0.1"),
		},
	}

	for _, data := range testData {
		result := convertPodStatsToTaskStats(data.podStats)
		if !(reflect.DeepEqual(data.expected, result)) {
			t.Error("expected ", data.expected, "got ", result)
		}
	}
}

func Test_convertNodeStatsToResourceStats(t *testing.T) {

	fakeNow := uint64(time.Now().UnixNano())
	var testData = []struct {
		nodeStats *NodeStats
		expected  *firmament.ResourceStats
	}{
		{
			nodeStats: BuildNodeStats("localhost", fakeNow),
			expected:  BuildResourceStats("localhost", fakeNow),
		},
		{
			nodeStats: BuildNodeStats("127.0.0.1", fakeNow),
			expected:  BuildResourceStats("127.0.0.1", fakeNow),
		},
	}

	for _, data := range testData {
		result := convertNodeStatsToResourceStats(data.nodeStats)
		if !(reflect.DeepEqual(data.expected, result)) {
			t.Error("expected ", data.expected, "got ", result)
		}
	}
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

func Test_ReceiveNodeStats(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	//setup data to send and receive
	fakeNow := uint64(time.Now().UnixNano())
	nodeData := BuildNodeStats("localhost", fakeNow)

	//setup the client side stream with expectation
	clientStream := NewMockPoseidonStats_ReceiveNodeStatsServer(ctrl)
	clientStream.EXPECT().Recv().Return(nodeData, nil)
	clientStream.EXPECT().Send(&NodeStatsResponse{
		Type:     NodeStatsResponseType_NODE_STATS_OK,
		Hostname: "localhost",
	})

	//Mock the stats server and pass the stream object
	server := NewMockPoseidonStatsServer(ctrl)
	server.EXPECT().ReceiveNodeStats(clientStream).Return(nil)
	err := server.ReceiveNodeStats(clientStream)
	if err != nil {

		t.Fatalf("cannot setup server %v", err)
	}

	data, err := clientStream.Recv()
	if err != nil {
		t.Fatalf("recv error %v", err)
	}
	if err := clientStream.Send(&NodeStatsResponse{
		Type:     NodeStatsResponseType_NODE_STATS_OK,
		Hostname: "localhost",
	}); err != nil {
		t.Fatalf("cannot setup server %v", err)
	}
	if !(reflect.DeepEqual(nodeData, data)) {
		t.Error("expected ", nodeData, "got ", data)
	}

}

func Test_ReceivePodStats(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	//setup data to send and receive
	podData := BuildPodStats("TestPod_One", "TestSpace", "localhost")

	//setup the client side stream with expectation
	clientStream := NewMockPoseidonStats_ReceivePodStatsServer(ctrl)
	clientStream.EXPECT().Recv().Return(podData, nil)
	clientStream.EXPECT().Send(&PodStatsResponse{
		Type:      PodStatsResponseType_POD_STATS_OK,
		Name:      "TestPod_One",
		Namespace: "TestSpace",
	})

	//Mock the stats server and pass the stream object
	server := NewMockPoseidonStatsServer(ctrl)
	server.EXPECT().ReceivePodStats(clientStream).Return(nil)
	err := server.ReceivePodStats(clientStream)
	if err != nil {

		t.Fatalf("cannot setup server %v", err)
	}

	data, err := clientStream.Recv()
	if err != nil {
		t.Fatalf("recv error %v", err)
	}
	if err := clientStream.Send(&PodStatsResponse{
		Type:      PodStatsResponseType_POD_STATS_OK,
		Name:      "TestPod_One",
		Namespace: "TestSpace",
	}); err != nil {
		t.Fatalf("cannot setup server %v", err)
	}
	if !(reflect.DeepEqual(podData, data)) {
		t.Error("expected ", podData, "got ", data)
	}

}
