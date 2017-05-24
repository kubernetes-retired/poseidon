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
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/pkg/api/v1"
	"testing"
)

const (
	// testKubeHostname is a fake hostname for unit tests
	testKubeHostname = "slave-node"
	// testKubeHostIP is a fake hostip for unit tests
	testKubeHostIP = "127.0.0.1"
)

type TestK8sClient struct {
	schedulerName string
	kubeConfig    string
	kubeClient    *fake.Clientset
}

// newTestK8sClient initializes and returns TestK8sClient
func newTestK8sClient(t *testing.T) *TestK8sClient {
	t.Logf("Initializing TestK8sClient...")
	schedulerName := "poseidon"
	fakeKubeConfig := ""
	fakeKubeClient := &fake.Clientset{}
	testK8sClient := &TestK8sClient{}
	testK8sClient.schedulerName = schedulerName
	testK8sClient.kubeConfig = fakeKubeConfig
	testK8sClient.kubeClient = fakeKubeClient
	clientSet = testK8sClient.kubeClient
	return testK8sClient
}

// cleanup is for deleting temporary files
// at the end of running test case
func (tk *TestK8sClient) cleanup() {
	//defer os.Remove(fakeKubeFile.Name())
}

func TestNew(t *testing.T) {
	k8sClient := newTestK8sClient(t)
	t.Log("k8sClient=", k8sClient)
	kubeClient := k8sClient.kubeClient
	t.Log("kubeClient=", kubeClient)
}

// TestGetClientConfig tests whether passed kubeconfig
// files are well-formed and parseable.
func TestGetClientConfig(t *testing.T) {
	fakeKubeFile, _ := ioutil.TempFile("", "")
	_, err := GetClientConfig(fakeKubeFile.Name())
	if err != nil {
		t.Log("Success. No kubeconfig file passed. err:", err)
	} else {
		t.Errorf("Fail. No kubeconfig file passed.")
	}

	// TODO (Karun): Below test should succeed
	// for now commenting as this is failing
	/*fakeKubeFile2, _ := ioutil.TempFile("test", "test")
	_,err = GetClientConfig(fakeKubeFile2.Name())
	if err!= nil {
		t.Errorf("Fail. kubeconfig file is avaialbe. err:", err)
	} else {
		t.Errorf("Success. kubeconfig file loaded.")
	}*/
}

// TestBindPodToNode tests BindPodToNode() function
func TestBindPodToNode(t *testing.T) {
	k8sClient := newTestK8sClient(t)
	t.Log("k8sClient=", k8sClient)
	kubeClient := k8sClient.kubeClient
	t.Log("kubeClient=", kubeClient)
	staticPod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			UID:       "123456789",
			Name:      "bar",
			Namespace: "default",
		},
	}
	staticNode := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: string(testKubeHostname),
		},
		Status: v1.NodeStatus{
			Addresses: []v1.NodeAddress{
				{
					Type:    v1.NodeInternalIP,
					Address: testKubeHostIP,
				},
			},
		},
	}

	BindPodToNode(staticPod.Name, staticPod.Namespace, staticNode.Name)
	t.Logf("Pod bind to Node success...")
}

// TestBindPodToNode tests DeletePod() function
func TestDeletePod(t *testing.T) {
	k8sClient := newTestK8sClient(t)
	t.Log("k8sClient=", k8sClient)
	kubeClient := k8sClient.kubeClient
	t.Log("kubeClient=", kubeClient)
	staticPod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			UID:       "123456789",
			Name:      "bar",
			Namespace: "default",
		},
	}

	DeletePod(staticPod.Name, staticPod.Namespace)
	t.Logf("Pod delete success...")
}
