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

	"k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/scheduling_poseidon/pkg/firmament"

	"github.com/golang/glog"
)

var clientSet kubernetes.Interface

func BindPodToNode(podName string, namespace string, nodeName string) {
	err := clientSet.CoreV1().Pods(namespace).Bind(&v1.Binding{
		meta_v1.TypeMeta{},
		meta_v1.ObjectMeta{
			Name: podName,
		},
		v1.ObjectReference{
			Namespace: namespace,
			Name:      nodeName,
		}})
	if err != nil {
		glog.Fatalf("Could not bind %v", err)
	}
}

func DeletePod(podName string, namespace string) {
	clientSet.CoreV1().Pods(namespace).Delete(podName, &meta_v1.DeleteOptions{})
}

func GetClientConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}

func New(schedulerName string, kubeConfig string, kubeVersionMajor, kubeVersionMinor int, firmamentAddress string) {
	config, err := GetClientConfig(kubeConfig)
	if err != nil {
		glog.Fatalf("Failed to load client config: %v", err)
	}
	clientSet, err = kubernetes.NewForConfig(config)
	if err != nil {
		glog.Fatalf("Failed to create connection: %v", err)
	}
	fc, conn, err := firmament.New(firmamentAddress)
	if err != nil {
		glog.Fatalf("Failed to connect to Firmament: %v", err)
	}
	defer conn.Close()
	glog.Info("k8s newclient called")
	stopCh := make(chan struct{})
	go NewPodWatcher(kubeVersionMajor, kubeVersionMinor, schedulerName, clientSet, fc).Run(stopCh, 10)
	go NewNodeWatcher(clientSet, fc).Run(stopCh, 10)

	// We block here.
	<-stopCh
}
