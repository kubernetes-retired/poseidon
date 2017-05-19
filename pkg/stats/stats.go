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

	"github.com/camsas/poseidon/pkg/firmament"
	"github.com/camsas/poseidon/pkg/k8sclient"
	"github.com/golang/glog"
	"google.golang.org/grpc"
)

type poseidonStatsServer struct {
	firmamentClient firmament.FirmamentSchedulerClient
}

func convertPodStatsToTaskStats(podStats *PodStats) *firmament.TaskStats {
	return &firmament.TaskStats{
		Hostname:            podStats.GetHostname(),
		CpuLimit:            podStats.GetCpuLimit(),
		CpuRequest:          podStats.GetCpuRequest(),
		CpuUsage:            podStats.GetCpuUsage(),
		MemLimit:            podStats.GetMemLimit(),
		MemRequest:          podStats.GetMemRequest(),
		MemUsage:            podStats.GetMemUsage(),
		MemRss:              podStats.GetMemRss(),
		MemCache:            podStats.GetMemCache(),
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
			glog.Info("Consumed all node stats from client")
			return nil
		}
		if err != nil {
			glog.Errorln("Stream error in node stats receive ", err)
			return err
		}
		resourceStats := convertNodeStatsToResourceStats(nodeStats)
		k8sclient.NodesCond.L.Lock()
		rtnd, ok := k8sclient.NodeToRTND[nodeStats.GetHostname()]
		k8sclient.NodesCond.L.Unlock()
		if !ok {
			sendErr := stream.Send(&NodeStatsResponse{
				Type:     NodeStatsResponseType_NODE_NOT_FOUND,
				Hostname: nodeStats.GetHostname(),
			})
			if sendErr != nil {
				glog.Error("Stream send error in node stats receive ", sendErr)
				return sendErr
			}
			continue
		}
		resourceStats.ResourceUid = &firmament.ResourceUID{
			ResourceUid: rtnd.GetResourceDesc().GetUuid(),
		}
		firmament.AddNodeStats(s.firmamentClient, resourceStats)
		sendErr := stream.Send(&NodeStatsResponse{
			Type:     NodeStatsResponseType_NODE_STATS_OK,
			Hostname: nodeStats.GetHostname(),
		})
		if sendErr != nil {
			glog.Error("Stream send error in node stats receive ", sendErr)
			return sendErr
		}
	}
}

func (s *poseidonStatsServer) ReceivePodStats(stream PoseidonStats_ReceivePodStatsServer) error {
	for {
		podStats, err := stream.Recv()
		if err == io.EOF {
			glog.Info("Consumed all pod stats from client")
			return nil
		}
		if err != nil {
			glog.Error("Stream receive error in pod stats receive ", err)
			return err
		}
		taskStats := convertPodStatsToTaskStats(podStats)
		podIdentifier := k8sclient.PodIdentifier{
			Name:      podStats.Name,
			Namespace: podStats.Namespace,
		}
		k8sclient.PodsCond.L.Lock()
		td, ok := k8sclient.PodToTD[podIdentifier]
		k8sclient.PodsCond.L.Unlock()
		if !ok {
			sendErr := stream.Send(&PodStatsResponse{
				Type:      PodStatsResponseType_POD_NOT_FOUND,
				Name:      podStats.GetName(),
				Namespace: podStats.GetNamespace(),
			})
			if sendErr != nil {
				glog.Error("Stream send error in pod stats receive ", sendErr)
				return sendErr
			}
			continue
		}
		taskStats.TaskId = td.GetUid()
		firmament.AddTaskStats(s.firmamentClient, taskStats)
		sendErr := stream.Send(&PodStatsResponse{
			Type:      PodStatsResponseType_POD_STATS_OK,
			Name:      podStats.GetName(),
			Namespace: podStats.GetNamespace(),
		})
		if sendErr != nil {
			glog.Error("Stream send error in pod stats receive ", sendErr)
			return sendErr
		}
	}
}

func StartgRPCStatsServer(statsServerAddress, firmamentAddress string) {
	glog.Info("Starting stats server...")
	listen, err := net.Listen("tcp", statsServerAddress)
	if err != nil {
		glog.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	fc, conn, err := firmament.New(firmamentAddress)
	if err != nil {
		glog.Fatalln("Unable to initialze Firmament client", err)

	}
	defer conn.Close()
	RegisterPoseidonStatsServer(grpcServer, &poseidonStatsServer{firmamentClient: fc})
	grpcServer.Serve(listen)
}
