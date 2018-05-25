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

package test

import (
	"github.com/kubernetes-sigs/poseidon/test/e2e/framework"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type testPodConfig struct {
	Name                              string
	Affinity                          *v1.Affinity
	Annotations, Labels, NodeSelector map[string]string
	Resources                         *v1.ResourceRequirements
	Tolerations                       []v1.Toleration
	NodeName                          string
	Ports                             []v1.ContainerPort
	OwnerReferences                   []metav1.OwnerReference
	PriorityClassName                 string
	SchedulerName                     string
}

func getRequestedCPU(pod v1.Pod) int64 {
	var result int64
	for _, container := range pod.Spec.Containers {
		result += container.Resources.Requests.Cpu().MilliValue()
	}
	return result
}

func createTestPod(f *framework.Framework, conf testPodConfig) *v1.Pod {
	pod, err := f.ClientSet.CoreV1().Pods(f.Namespace.Name).Create(initTestPod(f, conf))
	framework.ExpectNoError(err)
	return pod
}

func initTestPod(f *framework.Framework, conf testPodConfig) *v1.Pod {
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:            conf.Name,
			Labels:          conf.Labels,
			Annotations:     conf.Annotations,
			OwnerReferences: conf.OwnerReferences,
		},
		Spec: v1.PodSpec{
			NodeSelector: conf.NodeSelector,
			Affinity:     conf.Affinity,
			Containers: []v1.Container{
				{
					Name:  conf.Name,
					Image: "k8s.gcr.io/pause-amd64:3.0",
					Ports: conf.Ports,
				},
			},
			Tolerations:       conf.Tolerations,
			NodeName:          conf.NodeName,
			PriorityClassName: conf.PriorityClassName,
			SchedulerName:     conf.SchedulerName,
		},
	}
	if conf.Resources != nil {
		pod.Spec.Containers[0].Resources = *conf.Resources
	}
	return pod
}
