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

type TestNodeWatchObj struct {
	firmamentClient firmament.FirmamentSchedulerClient
	kubeClient      *fake.Clientset
}

// initializeNodeObj initializes and returns TestNodeWatchObj
func initializeNodeObj() *TestNodeWatchObj {
	testObj := &TestNodeWatchObj{}
	testObj.firmamentClient = fakeFirmamentClient()
	testObj.kubeClient = &fake.Clientset{}
	return testObj
}

// fakeFirmamentClient creates a fake firmament client
func fakeFirmamentClient() firmament.FirmamentSchedulerClient {
	fc := firmament.NewFirmamentSchedulerClient(nil)
	return fc
}

// TestNewNodeWatcher tests for NewNodeWatcher() function
func TestNewNodeWatcher(t *testing.T) {
	testObj := initializeNodeObj()
	nodeWatch := NewNodeWatcher(testObj.kubeClient, testObj.firmamentClient)
	t.Logf("Node watcher=", nodeWatch)
}

func TestNodeWatcher_Run(t *testing.T) {
	// No need of this test case here
	// We will test this in k8sclient_test.go
}
