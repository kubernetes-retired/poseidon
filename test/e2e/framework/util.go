/*
Copyright 2014 The Kubernetes Authors.

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
	"fmt"
	"regexp"
	"time"

	. "github.com/onsi/ginkgo"

	"k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/poseidon/test/e2e/framework/ginkgowrapper"

	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/pkg/client/conditions"
)

const (
	// How long to wait for the pod to be listable
	PodListTimeout = time.Minute
	// Initial pod start can be delayed O(minutes) by slow docker pulls
	// TODO: Make this 30 seconds once #4566 is resolved.
	PodStartTimeout = 5 * time.Minute

	// If there are any orphaned namespaces to clean up, this test is running
	// on a long lived cluster. A long wait here is preferably to spurious test
	// failures caused by leaked resources from a previous test run.
	NamespaceCleanupTimeout = 15 * time.Minute

	// Some pods can take much longer to get ready due to volume attach/detach latency.
	slowPodStartTimeout = 15 * time.Minute

	// How long to wait for a service endpoint to be resolvable.
	ServiceStartTimeout = 1 * time.Minute

	// How often to Poll pods, nodes and claims.
	Poll = 2 * time.Second

	pollShortTimeout = 1 * time.Minute
	pollLongTimeout  = 5 * time.Minute

	// service accounts are provisioned after namespace creation
	// a service account is required to support pod creation in a namespace as part of admission control
	ServiceAccountProvisionTimeout = 2 * time.Minute

	// How long to try single API calls (like 'get' or 'list'). Used to prevent
	// transient failures from failing tests.
	// TODO: client should not apply this timeout to Watch calls. Increased from 30s until that is fixed.
	SingleCallTimeout = 5 * time.Minute

	// How long nodes have to be "ready" when a test begins. They should already
	// be "ready" before the test starts, so this is small.
	NodeReadyInitialTimeout = 20 * time.Second

	// How long pods have to be "ready" when a test begins.
	PodReadyBeforeTimeout = 5 * time.Minute

	// How long pods have to become scheduled onto nodes
	podScheduledBeforeTimeout = PodListTimeout + (20 * time.Second)

	podRespondingTimeout     = 15 * time.Minute
	ServiceRespondingTimeout = 2 * time.Minute
	EndpointRegisterTimeout  = time.Minute

	// How long claims have to become dynamically provisioned
	ClaimProvisionTimeout = 5 * time.Minute

	// How long claims have to become bound
	ClaimBindingTimeout = 3 * time.Minute

	// How long claims have to become deleted
	ClaimDeletingTimeout = 3 * time.Minute

	// How long PVs have to beome reclaimed
	PVReclaimingTimeout = 3 * time.Minute

	// How long PVs have to become bound
	PVBindingTimeout = 3 * time.Minute

	// How long PVs have to become deleted
	PVDeletingTimeout = 3 * time.Minute

	// How long a node is allowed to become "Ready" after it is restarted before
	// the test is considered failed.
	RestartNodeReadyAgainTimeout = 5 * time.Minute

	// How long a pod is allowed to become "running" and "ready" after a node
	// restart before test is considered failed.
	RestartPodReadyAgainTimeout = 5 * time.Minute

	// Number of objects that gc can delete in a second.
	// GC issues 2 requestes for single delete.
	gcThroughput = 10

	// Minimal number of nodes for the cluster to be considered large.
	largeClusterThreshold = 100

	// ssh port
	sshPort = "22"

	DefaultPodDeletionTimeout = 3 * time.Minute
)

var (
	BusyBoxImage = "busybox"
	// Label allocated to the image puller static pod that runs on each node
	// before e2es.
	ImagePullerLabels = map[string]string{"name": "e2e-image-puller"}

	// For parsing Kubectl version for version-skewed testing.
	gitVersionRegexp = regexp.MustCompile("GitVersion:\"(v.+?)\"")

	// Slice of regexps for names of pods that have to be running to consider a Node "healthy"
	requiredPerNodePods = []*regexp.Regexp{
		regexp.MustCompile(".*kube-proxy.*"),
		regexp.MustCompile(".*fluentd-elasticsearch.*"),
		regexp.MustCompile(".*node-problem-detector.*"),
	}
)

// Waits default amount of time (PodStartTimeout) for the specified pod to become running.
// Returns an error if timeout occurs first, or pod goes in to failed state.
func WaitForPodRunningInNamespace(c clientset.Interface, pod *v1.Pod) error {
	if pod.Status.Phase == v1.PodRunning {
		return nil
	}
	return WaitTimeoutForPodRunningInNamespace(c, pod.Name, pod.Namespace, PodStartTimeout)
}

// Waits default amount of time (PodStartTimeout) for the specified pod to become running.
// Returns an error if timeout occurs first, or pod goes in to failed state.
func WaitForPodNameRunningInNamespace(c clientset.Interface, podName, namespace string) error {
	return WaitTimeoutForPodRunningInNamespace(c, podName, namespace, PodStartTimeout)
}

// Waits an extended amount of time (slowPodStartTimeout) for the specified pod to become running.
// The resourceVersion is used when Watching object changes, it tells since when we care
// about changes to the pod. Returns an error if timeout occurs first, or pod goes in to failed state.
func waitForPodRunningInNamespaceSlow(c clientset.Interface, podName, namespace string) error {
	return WaitTimeoutForPodRunningInNamespace(c, podName, namespace, slowPodStartTimeout)
}

func WaitTimeoutForPodRunningInNamespace(c clientset.Interface, podName, namespace string, timeout time.Duration) error {
	return wait.PollImmediate(Poll, timeout, podRunning(c, podName, namespace))
}

func podRunning(c clientset.Interface, podName, namespace string) wait.ConditionFunc {
	return func() (bool, error) {
		pod, err := c.CoreV1().Pods(namespace).Get(podName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		switch pod.Status.Phase {
		case v1.PodRunning:
			return true, nil
		case v1.PodFailed, v1.PodSucceeded:
			return false, conditions.ErrPodCompleted
		}
		return false, nil
	}
}

// Waits default amount of time (DefaultPodDeletionTimeout) for the specified pod to stop running.
// Returns an error if timeout occurs first.
func WaitForPodNoLongerRunningInNamespace(c clientset.Interface, podName, namespace string) error {
	return WaitTimeoutForPodNoLongerRunningInNamespace(c, podName, namespace, DefaultPodDeletionTimeout)
}

func WaitTimeoutForPodNoLongerRunningInNamespace(c clientset.Interface, podName, namespace string, timeout time.Duration) error {
	return wait.PollImmediate(Poll, timeout, podCompleted(c, podName, namespace))
}

func podCompleted(c clientset.Interface, podName, namespace string) wait.ConditionFunc {
	return func() (bool, error) {
		pod, err := c.CoreV1().Pods(namespace).Get(podName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		switch pod.Status.Phase {
		case v1.PodFailed, v1.PodSucceeded:
			return true, nil
		}
		return false, nil
	}
}

// WaitForPodNotPending returns an error if it took too long for the pod to go out of pending state.
// The resourceVersion is used when Watching object changes, it tells since when we care
// about changes to the pod.
func WaitForPodNotPending(c clientset.Interface, ns, podName string) error {
	return wait.PollImmediate(Poll, PodStartTimeout, podNotPending(c, podName, ns))
}

func podNotPending(c clientset.Interface, podName, namespace string) wait.ConditionFunc {
	return func() (bool, error) {
		pod, err := c.CoreV1().Pods(namespace).Get(podName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		switch pod.Status.Phase {
		case v1.PodPending:
			return false, nil
		default:
			return true, nil
		}
	}
}

// waitForPodNotFoundInNamespace returns an error if it takes too long for the pod to fully terminate.
// Unlike `waitForPodTerminatedInNamespace`, the pod's Phase and Reason are ignored. If the pod Get
// api returns IsNotFound then the wait stops and nil is returned. If the Get api returns an error other
// than "not found" then that error is returned and the wait stops.
func waitForPodNotFoundInNamespace(c clientset.Interface, podName, ns string, timeout time.Duration) error {
	return wait.PollImmediate(Poll, timeout, func() (bool, error) {
		_, err := c.CoreV1().Pods(ns).Get(podName, metav1.GetOptions{})
		if apierrs.IsNotFound(err) {
			return true, nil // done
		}
		if err != nil {
			return true, err // stop wait with error
		}
		return false, nil
	})
}

// deleteNS deletes the provided namespace, waits for it to be completely deleted, and then checks
// whether there are any pods remaining in a non-terminating state.
func deleteNS(c clientset.Interface, namespace string, timeout time.Duration) error {
	startTime := time.Now()
	if err := c.CoreV1().Namespaces().Delete(namespace, nil); err != nil {
		return err
	}

	// wait for namespace to delete or timeout.
	err := wait.PollImmediate(2*time.Second, timeout, func() (bool, error) {
		if _, err := c.CoreV1().Namespaces().Get(namespace, metav1.GetOptions{}); err != nil {
			if apierrs.IsNotFound(err) {
				return true, nil
			}
			Logf("Error while waiting for namespace to be terminated: %v", err)
			return false, nil
		}
		return false, nil
	})

	if err != nil {
		return fmt.Errorf("namespace %v was not deleted with error: %v,", namespace, err)
	}

	Logf("namespace %v deletion completed in %s", namespace, time.Now().Sub(startTime))
	return nil
}

func Logf(format string, args ...interface{}) {
	log("INFO", format, args...)
}

func Failf(format string, args ...interface{}) {
	FailfWithOffset(1, format, args...)
}

// FailfWithOffset calls "Fail" and logs the error at "offset" levels above its caller
// (for example, for call chain f -> g -> FailfWithOffset(1, ...) error would be logged for "f").
func FailfWithOffset(offset int, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	log("INFO", msg)
	ginkgowrapper.Fail(nowStamp()+": "+msg, 1+offset)
}

func log(level string, format string, args ...interface{}) {
	fmt.Fprintf(GinkgoWriter, nowStamp()+": "+level+": "+format+"\n", args...)
}

func nowStamp() string {
	return time.Now().Format(time.StampMilli)
}
