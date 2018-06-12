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

package main

import (
	"time"

	"github.com/kubernetes-sigs/poseidon/pkg/config"
	"github.com/kubernetes-sigs/poseidon/pkg/firmament"
	"github.com/kubernetes-sigs/poseidon/pkg/k8sclient"
	"github.com/kubernetes-sigs/poseidon/pkg/metrics"
	"github.com/kubernetes-sigs/poseidon/pkg/poseidonhttp"
	"github.com/kubernetes-sigs/poseidon/pkg/stats"

	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/golang/glog"
)

func schedule(fc firmament.FirmamentSchedulerClient) {
	for {
		deltas := firmament.Schedule(fc)
		glog.Infof("Scheduler returned %d deltas", len(deltas.GetDeltas()))
		for _, delta := range deltas.GetDeltas() {
			switch delta.GetType() {
			case firmament.SchedulingDelta_PLACE:
				k8sclient.PodMux.RLock()
				podIdentifier, ok := k8sclient.TaskIDToPod[delta.GetTaskId()]
				k8sclient.PodMux.RUnlock()
				if !ok {
					glog.Fatalf("Placed task %d without pod pairing", delta.GetTaskId())
				}
				k8sclient.NodeMux.RLock()
				nodeName, ok := k8sclient.ResIDToNode[delta.GetResourceId()]
				k8sclient.NodeMux.RUnlock()
				if !ok {
					glog.Fatalf("Placed task %d on resource %s without node pairing", delta.GetTaskId(), delta.GetResourceId())
				}
				// TODO(jiaxuanzhou): Metric the latency of binding one node when client provided to get the desc of the task(pod)
				// metrics.BindingLatency.Observe(metrics.SinceInMicroseconds(time.Time(task.SubmitTime)))
				k8sclient.BindPodToNode(podIdentifier.Name, podIdentifier.Namespace, nodeName)
			case firmament.SchedulingDelta_PREEMPT, firmament.SchedulingDelta_MIGRATE:
				k8sclient.PodMux.RLock()
				preemptionStartTime := time.Now()
				podIdentifier, ok := k8sclient.TaskIDToPod[delta.GetTaskId()]
				k8sclient.PodMux.RUnlock()
				if !ok {
					glog.Fatalf("Preempted task %d without pod pairing", delta.GetTaskId())
				}
				metrics.PreemptionAttempts.Inc()
				// XXX(ionel): HACK! Kubernetes does not yet have support for preemption.
				// However, preemption can be achieved by deleting the preempted pod
				// and relying on the controller mechanism (e.g., job, replica set)
				// to submit another instance of this pod.
				k8sclient.DeletePod(podIdentifier.Name, podIdentifier.Namespace)
				metrics.SchedulingPremptionEvaluationDuration.Observe(metrics.SinceInMicroseconds(preemptionStartTime))
			case firmament.SchedulingDelta_NOOP:
			default:
				glog.Fatalf("Unexpected SchedulingDelta type %v", delta.GetType())
			}
		}
		// TODO(ionel): Temporary sleep statement because we currently call the scheduler even if there's no work do to.
		time.Sleep(time.Duration(config.GetSchedulingInterval()) * time.Second)
	}
}

// WaitForFirmamentService blocks till the Firmament service is available
func WaitForFirmamentService(fc firmament.FirmamentSchedulerClient) {
	// TODO(jiaxuanzhou): Need to metric the wait latency of firmament service?
	serviceReq := new(firmament.HealthCheckRequest)
	err := wait.PollImmediate(2*time.Second, 10*time.Minute, func() (bool, error) {
		ok, _ := firmament.Check(fc, serviceReq)
		if !ok {
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		glog.Fatalf("Timed-out waiting for firmament service %v", err)
	}
}

func main() {
	glog.Info("Starting Poseidon...", config.GetFirmamentAddress())
	fc, conn, err := firmament.New(config.GetFirmamentAddress())
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	// Check if firmament grpc service is available and then proceed
	WaitForFirmamentService(fc)
	go schedule(fc)
	go stats.StartgRPCStatsServer(config.GetStatsServerAddress(), config.GetFirmamentAddress())
	go poseidonhttp.Serve(fc)
	kubeMajorVer, kubeMinorVer := config.GetKubeVersion()
	k8sclient.New(config.GetSchedulerName(), config.GetKubeConfig(), kubeMajorVer, kubeMinorVer, config.GetFirmamentAddress())
}
