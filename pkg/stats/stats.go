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

package stats

import (
	"io"
	"net"

	"github.com/ICGog/poseidongo/pkg/firmament"
	"github.com/ICGog/poseidongo/pkg/k8sclient"
	"github.com/golang/glog"
	"google.golang.org/grpc"
)

type poseidonStatsServer struct {
	firmamentClient firmament.FirmamentSchedulerClient
}

func convertPodStatsToTaskStats(podStats *PodStats) *firmament.TaskStats {
	return &firmament.TaskStats{
		CpuLimit:            podStats.GetCpuLimit(),
		CpuRequest:          podStats.GetCpuRequest(),
		CpuUsage:            podStats.GetCpuUsage(),
		MemLimit:            podStats.GetMemLimit(),
		MemRequest:          podStats.GetMemRequest(),
		MemUsage:            podStats.GetMemUsage(),
		MemWorkingSet:       podStats.GetMemWorkingSet(),
		MemPageFaults:       podStats.GetMemPageFaults(),
		MemPageFaultsRate:   podStats.GetMemPageFaultsRate(),
		MajorPageFaults:     podStats.GetMajorPageFaults(),
		MajorPageFaultsRate: podStats.GetMajorPageFaultsRate(),
		NetRx:               podStats.GetNetRx(),
		NetRxErrors:         podStats.GetNetRxErrors(),
		NetRxErrorsRate:     podStats.GetNetRxErrorsRate(),
		NetRxRate:           podStats.GetNetRxRate(),
		NetTx:               podStats.GetNetTx(),
		NetTxErrors:         podStats.GetNetTxErrors(),
		NetTxErrorsRate:     podStats.GetNetTxErrorsRate(),
		NetTxRate:           podStats.GetNetTxRate(),
	}
}

func convertNodeStatsToResourceStats(nodeStats *NodeStats) *firmament.ResourceStats {
	return &firmament.ResourceStats{
		CpuAllocatable: nodeStats.GetCpuAllocatable(),
		CpuCapacity:    nodeStats.GetCpuCapacity(),
		CpuReservation: nodeStats.GetCpuReservation(),
		CpuUtilization: nodeStats.GetCpuUtilization(),
		MemAllocatable: nodeStats.GetMemAllocatable(),
		MemCapacity:    nodeStats.GetMemCapacity(),
		MemReservation: nodeStats.GetMemReservation(),
		MemUtilization: nodeStats.GetMemUtilization(),
	}

}

func (s *poseidonStatsServer) ReceiveNodeStats(stream PoseidonStats_ReceiveNodeStatsServer) error {
	for {
		nodeStats, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		resourceStats := convertNodeStatsToResourceStats(nodeStats)
		rtnd, ok := k8sclient.NodeToRTND[nodeStats.GetHostname()]
		if !ok {
			// TODO(ionel): Handle case when node does not have an associated resource id.
		}
		resourceStats.ResourceUid = &firmament.ResourceUID{
			ResourceUid: rtnd.GetResourceDesc().GetUuid(),
		}
		firmament.AddNodeStats(s.firmamentClient, resourceStats)
	}
}

func (s *poseidonStatsServer) ReceivePodStats(stream PoseidonStats_ReceivePodStatsServer) error {
	for {
		podStats, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		taskStats := convertPodStatsToTaskStats(podStats)
		podIdentifier := k8sclient.PodIdentifier{
			Name:      podStats.Name,
			Namespace: podStats.Namespace,
		}
		td, ok := k8sclient.PodToTD[podIdentifier]
		if !ok {
			// TODO(ionel): Handle case when pod does not have an associated task id.
		}
		taskStats.TaskUid = &firmament.TaskUID{
			TaskUid: td.GetUid(),
		}
		firmament.AddTaskStats(s.firmamentClient, taskStats)
	}
}

func newposeidonStatsServer(firmamentAddress string) *poseidonStatsServer {
	newfirmamentClient, _, err := firmament.New(firmamentAddress)
	if err != nil {
		glog.Fatalln("Unable to initialze Firmament client", err)

	}
	return &poseidonStatsServer{firmamentClient: newfirmamentClient}
}

func StartgRPCStatsServer(statsServerAddress, firmamentAddress string) {
	glog.Info("Starting stats server...")
	listen, err := net.Listen("tcp", statsServerAddress)
	if err != nil {
		glog.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	RegisterPoseidonStatsServer(grpcServer, newposeidonStatsServer(firmamentAddress))
	grpcServer.Serve(listen)
}
