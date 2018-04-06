/*
Copyright 2015 The Kubernetes Authors.

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

package framework

import (
	"time"

	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	//. "github.com/onsi/ginkgo"
)

// Framework supports common operations used by e2e tests; it will keep a client & a namespace for you.
// Eventual goal is to merge this with integration test framework.
type Framework struct {
	BaseName string

	ClientSet clientset.Interface
	Namespace *v1.Namespace
	Options   FrameworkOptions
}

type FrameworkOptions struct {
	ClientQPS   float32
	ClientBurst int
}

// NewFramework makes a new framework and sets up a BeforeEach/AfterEach for
// you (you can write additional before/after each functions).
func NewDefaultFramework(baseName string, client clientset.Interface) *Framework {
	options := FrameworkOptions{
		ClientQPS:   20,
		ClientBurst: 50,
	}
	return NewFramework(baseName, options, client)
}

func NewFramework(baseName string, options FrameworkOptions, client clientset.Interface) *Framework {
	f := &Framework{
		BaseName:  baseName,
		Options:   options,
		ClientSet: client,
	}

	//BeforeEach(f.BeforeEach)
	//AfterEach(f.AfterEach)

	return f
}

// BeforeEach gets a client and makes a namespace.
func (f *Framework) BeforeEach() {
	if f.ClientSet == nil {
		glog.Fatal("ClientSet is nil")
	}

}

// AfterEach deletes the namespace, after reading its events.
func (f *Framework) AfterEach() {
	//delete ns

}

// WaitForPodNotFound waits for the pod to be completely terminated (not "Get-able").
func (f *Framework) WaitForPodNotFound(podName string, timeout time.Duration) error {
	return waitForPodNotFoundInNamespace(f.ClientSet, podName, f.Namespace.Name, timeout)
}

// WaitForPodRunning waits for the pod to run in the namespace.
func (f *Framework) WaitForPodRunning(podName string) error {
	return WaitForPodNameRunningInNamespace(f.ClientSet, podName, f.Namespace.Name)
}

// WaitForPodRunningSlow waits for the pod to run in the namespace.
// It has a longer timeout then WaitForPodRunning (util.slowPodStartTimeout).
func (f *Framework) WaitForPodRunningSlow(podName string) error {
	return waitForPodRunningInNamespaceSlow(f.ClientSet, podName, f.Namespace.Name)
}

// WaitForPodNoLongerRunning waits for the pod to no longer be running in the namespace, for either
// success or failure.
func (f *Framework) WaitForPodNoLongerRunning(podName string) error {
	return WaitForPodNoLongerRunningInNamespace(f.ClientSet, podName, f.Namespace.Name)
}
