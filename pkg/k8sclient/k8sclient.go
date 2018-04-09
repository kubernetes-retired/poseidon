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
	"k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/poseidon/pkg/firmament"

	"github.com/golang/glog"
)

var clientSet kubernetes.Interface

// BindPodToNode call Kubernetes API to place a pod on a node.
func BindPodToNode(podName string, namespace string, nodeName string) {
	err := clientSet.CoreV1().Pods(namespace).Bind(&v1.Binding{
		TypeMeta: meta_v1.TypeMeta{},
		ObjectMeta: meta_v1.ObjectMeta{
			Name: podName,
		},
		Target: v1.ObjectReference{
			Namespace: namespace,
			Name:      nodeName,
		}})
	if err != nil {
		glog.Fatalf("Could not bind %v", err)
	}
}

// DeletePod calls Kubernetes API to delete a Pod by its namespace and name.
func DeletePod(podName string, namespace string) {
	clientSet.CoreV1().Pods(namespace).Delete(podName, &meta_v1.DeleteOptions{})
}

// GetClientConfig returns a kubeconfig object which to be passed to a Kubernetes client on initialization.
func GetClientConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}

// New initializes a firmament and Kubernetes client and starts watching Pod and Node.
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
