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

package main

import (
	"flag"
	"strconv"
	"strings"
	"time"

	"github.com/camsas/poseidon/pkg/firmament"
	"github.com/camsas/poseidon/pkg/k8sclient"
	"github.com/camsas/poseidon/pkg/stats"
	"github.com/golang/glog"
)

var (
	schedulerName      string
	firmamentAddress   string
	kubeConfig         string
	kubeVersion        string
	statsServerAddress string
	schedulingInterval int
)

func init() {
	flag.StringVar(&schedulerName, "schedulerName", "poseidon", "The scheduler name with which pods are labeled")
	flag.StringVar(&firmamentAddress, "firmamentAddress", "127.0.0.1:9090", "Firmament scheduler service address and port")
	flag.StringVar(&kubeConfig, "kubeConfig", "kubeconfig.cfg", "Path to the kubeconfig file")
	flag.StringVar(&kubeVersion, "kubeVersion", "1.5.6", "Kubernetes version")
	flag.StringVar(&statsServerAddress, "statsServerAddress", "127.0.0.1:9091", "Address on which the stats server listens")
	flag.IntVar(&schedulingInterval, "schedulingInterval", 10, "Time between scheduler runs (in seconds)")
	flag.Parse()
}

func schedule(fc firmament.FirmamentSchedulerClient) {
	for {
		deltas := firmament.Schedule(fc)
		glog.Infof("Scheduler returned %d deltas", len(deltas.GetDeltas()))
		for _, delta := range deltas.GetDeltas() {
			switch delta.GetType() {
			case firmament.SchedulingDelta_PLACE:
				k8sclient.PodsCond.L.Lock()
				podIdentifier, ok := k8sclient.TaskIDToPod[delta.GetTaskId()]
				k8sclient.PodsCond.L.Unlock()
				if !ok {
					glog.Fatalf("Placed task %d without pod pairing", delta.GetTaskId())
				}
				k8sclient.NodesCond.L.Lock()
				nodeName, ok := k8sclient.ResIDToNode[delta.GetResourceId()]
				k8sclient.NodesCond.L.Unlock()
				if !ok {
					glog.Fatalf("Placed task %d on resource %s without node pairing", delta.GetTaskId(), delta.GetResourceId())
				}
				k8sclient.BindPodToNode(podIdentifier.Name, podIdentifier.Namespace, nodeName)
			case firmament.SchedulingDelta_PREEMPT, firmament.SchedulingDelta_MIGRATE:
				k8sclient.PodsCond.L.Lock()
				podIdentifier, ok := k8sclient.TaskIDToPod[delta.GetTaskId()]
				k8sclient.PodsCond.L.Unlock()
				if !ok {
					glog.Fatalf("Preempted task %d without pod pairing", delta.GetTaskId())
				}
				// XXX(ionel): HACK! Kubernetes does not yet have support for preemption.
				// However, preemption can be achieved by deleting the preempted pod
				// and relying on the controller mechanism (e.g., job, replica set)
				// to submit another instance of this pod.
				k8sclient.DeletePod(podIdentifier.Name, podIdentifier.Namespace)
			case firmament.SchedulingDelta_NOOP:
			default:
				glog.Fatalf("Unexpected SchedulingDelta type %v", delta.GetType())
			}
		}
		// TODO(ionel): Temporary sleep statement because we currently call the scheduler even if there's no work do to.
		time.Sleep(time.Duration(schedulingInterval) * time.Second)
	}
}

func main() {
	glog.Info("Starting Poseidon...")
	fc, conn, err := firmament.New(firmamentAddress)
	defer conn.Close()
	if err != nil {
		panic(err)
	}
	go schedule(fc)
	go stats.StartgRPCStatsServer(statsServerAddress, firmamentAddress)
	kubeVer := strings.Split(kubeVersion, ".")
	kubeMajorVer, err := strconv.Atoi(kubeVer[0])
	if err != nil {
		glog.Fatalf("Incorrect content in --kubeVersion %s", kubeVersion)
	}
	kubeMinorVer, err := strconv.Atoi(kubeVer[1])
	if err != nil {
		glog.Fatalf("Incorrect content in --kubeVersion %s", kubeVersion)
	}
	k8sclient.New(schedulerName, kubeConfig, kubeMajorVer, kubeMinorVer, firmamentAddress)
}
