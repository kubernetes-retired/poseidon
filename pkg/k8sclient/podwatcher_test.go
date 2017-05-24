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
	"github.com/ICGog/poseidongo/pkg/firmament"
	"k8s.io/client-go/kubernetes/fake"
	"testing"
)

type TestPodWatchObj struct {
	firmamentClient firmament.FirmamentSchedulerClient
	kubeClient      *fake.Clientset
	kubeVerMajor    int
	kubeVerMinor    int
	schedulerName   string
}

// initializePodObj initializes and returns TestPodWatchObj
func initializePodObj() *TestPodWatchObj {
	testObj := &TestPodWatchObj{}
	testObj.firmamentClient = fakeFirmamentClient()
	testObj.kubeClient = &fake.Clientset{}
	testObj.kubeVerMajor = 1
	testObj.kubeVerMinor = 6
	testObj.schedulerName = "poseidon"
	return testObj
}

// TestNewPodWatcher tests for different k8s versions for NewPodWatcher()
func TestNewPodWatcher(t *testing.T) {
	testObj := initializePodObj()
	// for default k8s 1.6
	podWatch := NewPodWatcher(testObj.kubeVerMajor, testObj.kubeVerMinor, testObj.schedulerName, testObj.kubeClient, testObj.firmamentClient)
	t.Logf("Pod watcher for v1.6=", podWatch)

	// for k8s 1.5
	testObj.kubeVerMajor = 1
	testObj.kubeVerMinor = 5
	podWatch = NewPodWatcher(testObj.kubeVerMajor, testObj.kubeVerMinor, testObj.schedulerName, testObj.kubeClient, testObj.firmamentClient)
	t.Logf("Pod watcher for v1.5=", podWatch)

}
