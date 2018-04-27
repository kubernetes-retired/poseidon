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

package framework

import (
	"flag"
	"os"
	"strconv"
	"time"

	"bytes"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os/exec"
	"strings"
)

var kubeConfig = flag.String(clientcmd.RecommendedConfigPathFlag, os.Getenv(clientcmd.RecommendedConfigPathEnvVar), "Path to kubeconfig containing embedded authinfo.")
var kubectlPath = flag.String("kubectl-path", "kubectl", "The kubectl binary to use. For development, you might use 'cluster/kubectl.sh' here.")
var poseidonManifestPath = flag.String("poseidonManifestPath", "github.com/kubernetes-sigs/poseidon/deploy/poseidon-deployment.yaml", "The Poseidon deployment manifest to use.")
var firmamentManifestPath = flag.String("firmamentManifestPath", "github.com/kubernetes-sigs/poseidon/deploy/firmament-deployment.yaml", "The Firmament deployment manifest to use.")

func init() {
	flag.Parse()
	fmt.Println(*kubeConfig, *kubectlPath, *poseidonManifestPath, *firmamentManifestPath)
}

// Framework supports common operations used by e2e tests; it will keep a client & a namespace for you.
// Eventual goal is to merge this with integration test framework.
type Framework struct {
	BaseName string

	ClientSet clientset.Interface
	Namespace *v1.Namespace
	TestingNS string
	Options   FrameworkOptions
}

type FrameworkOptions struct {
	ClientQPS   float32
	ClientBurst int
}

// NewDefaultFramework makes a new framework and sets up a BeforeEach/AfterEach for
// you (you can write additional before/after each functions).
func NewDefaultFramework(baseName string) *Framework {
	options := FrameworkOptions{
		ClientQPS:   20,
		ClientBurst: 50,
	}
	return NewFramework(baseName, options, nil)
}

// NewFramework makes a new framework and sets up a BeforeEach/AfterEach
func NewFramework(baseName string, options FrameworkOptions, client clientset.Interface) *Framework {
	f := &Framework{
		BaseName:  baseName,
		Options:   options,
		ClientSet: nil,
		TestingNS: "poseidon-test",
	}

	BeforeSuite(f.BeforeEach)
	AfterSuite(f.AfterEach)

	return f
}

// BeforeEach gets a client and makes a namespace.
func (f *Framework) BeforeEach() {
	var err error
	if f.ClientSet == nil {
		var config *rest.Config
		var err error
		config, err = clientcmd.BuildConfigFromFlags("", *kubeConfig)
		if err != nil {
			panic(err)
		}
		cs, err := clientset.NewForConfig(config)
		if err != nil {
			panic(err)
		}
		f.ClientSet = cs

	}

	_ = f.deleteNamespace(f.TestingNS)
	_ = f.DeletePoseidon()
	_ = f.DeleteFirmament()

	f.Namespace, err = f.createNamespace(f.ClientSet)
	Expect(err).NotTo(HaveOccurred())

	err = f.CreateFirmament()
	Expect(err).NotTo(HaveOccurred())

	err = f.CreatePoseidon()
	Expect(err).NotTo(HaveOccurred())

	go f.FetchLogsforFirmament()
	go f.FetchLogsforPoseidon()

}

// AfterEach deletes the namespace, after reading its events.
func (f *Framework) AfterEach() {
	//delete ns
	var err error

	if f.ClientSet == nil {
		Expect(f.ClientSet).To(Not(Equal(nil)))
	}
	Logf("Delete namespace called")
	err = f.deleteNamespace(f.TestingNS)
	Expect(err).NotTo(HaveOccurred())

	err = f.DeletePoseidon()
	Expect(err).NotTo(HaveOccurred())

	err = f.DeleteFirmament()
	Expect(err).NotTo(HaveOccurred())

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

func (f *Framework) deleteNamespace(nsName string) error {

	if err := f.ClientSet.CoreV1().Namespaces().Delete(nsName, nil); err != nil {
		return err
	}

	// wait for namespace to delete or timeout.
	err := wait.PollImmediate(2*time.Second, 10*time.Minute, func() (bool, error) {
		if _, err := f.ClientSet.CoreV1().Namespaces().Get(nsName, metav1.GetOptions{}); err != nil {
			if errors.IsNotFound(err) {
				return false, nil
			}
			Logf("Error while waiting for namespace to be terminated: %v", err)
			return false, err
		}
		return true, nil
	})

	remainingPods, missingTimestamp, _ := countRemainingPods(f.ClientSet, nsName)

	Logf("Total of %v pods available in the namespace after namespace deletion and %v pods dont have deletiontimestamp", remainingPods, missingTimestamp)

	return err

}

func (f *Framework) createNamespace(c clientset.Interface) (*v1.Namespace, error) {

	var got *v1.Namespace
	if err := wait.PollImmediate(2*time.Second, 30*time.Second, func() (bool, error) {
		var err error
		got, err = c.CoreV1().Namespaces().Create(&v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: f.TestingNS},
		})
		if err != nil {
			Logf("Unexpected error while creating namespace: %v", err)
			return false, nil
		}
		return true, nil
	}); err != nil {
		return nil, err
	}

	return got, nil
}

func (f *Framework) CreateFirmament() error {
	outputStr, errorStr, err := f.KubectlExecCreate(*firmamentManifestPath)
	if err != nil {
		Logf("Command error string %v", errorStr)
		Logf("Command output string %v", outputStr)
		Logf("%v", err)
	}
	return err
}

func (f *Framework) CreatePoseidon() error {
	outputStr, errorStr, err := f.KubectlExecCreate(*poseidonManifestPath)
	if err != nil {
		Logf("Command error string %v", errorStr)
		Logf("Command output string %v", outputStr)
		Logf("%v", err)
	}
	return err
}

func (f *Framework) DeleteFirmament() error {
	outputStr, errorStr, err := f.KubectlExecDelete(*firmamentManifestPath)
	if err != nil {
		Logf("Command error string %v", errorStr)
		Logf("Command output string %v", outputStr)
		Logf("%v", err)
	}
	return err
}

func (f *Framework) DeletePoseidon() error {
	outputStr, errorStr, err := f.KubectlExecDelete(*poseidonManifestPath)
	if err != nil {
		Logf("Command error string %v", errorStr)
		Logf("Command output string %v", outputStr)
		Logf("%v", err)
	}
	return err
}

// KubectlCmd runs the kubectl executable through the wrapper script.
func KubectlCmd(args ...string) *exec.Cmd {
	defaultArgs := []string{}

	if kubeConfig != nil {
		defaultArgs = append(defaultArgs, "--"+clientcmd.RecommendedConfigPathFlag+"="+*kubeConfig)

	}
	kubectlArgs := append(defaultArgs, args...)
	cmd := exec.Command(*kubectlPath, kubectlArgs...)
	return cmd
}

func (f *Framework) KubectlExecCreate(manifestPath string) (string, string, error) {
	var stdout, stderr bytes.Buffer
	cmdArgs := []string{
		fmt.Sprintf("create"),
		fmt.Sprintf("-f"),
		fmt.Sprintf("%v", manifestPath),
	}
	cmd := KubectlCmd(cmdArgs...)
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	Logf("Running '%s %s'", cmd.Path, strings.Join(cmdArgs, " "))
	err := cmd.Run()

	if err != nil {
		Logf("Unable to deploy %v %v", stdout.String(), stderr.String())
	}

	return stdout.String(), stderr.String(), err
}

func (f *Framework) KubectlExecDelete(manifestPath string) (string, string, error) {
	var stdout, stderr bytes.Buffer
	cmdArgs := []string{
		fmt.Sprintf("delete"),
		fmt.Sprintf("-f"),
		fmt.Sprintf("%v", manifestPath),
	}
	cmd := KubectlCmd(cmdArgs...)
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	Logf("Running '%s %s'", cmd.Path, strings.Join(cmdArgs, " "))
	err := cmd.Run()

	if err != nil {
		Logf("Unable to deploy %v %v", stdout.String(), stderr.String())
	}

	return stdout.String(), stderr.String(), err
}

// FetchLogsforPoseidon will get the Poseidon and firmament deplyment logs
func (f *Framework) FetchLogsforPoseidon() {
	matchLabel := map[string]string{
		"component": "poseidon",
	}

	// wait for namespace to delete or timeout.
	err := wait.PollImmediate(2*time.Second, 10*time.Minute, func() (bool, error) {
		// As poseidon and firmament run in kube-system namespaces we need to pass the correct ns here
		f.LogPodsWithLabels("kube-system", matchLabel, Logf)
		return false, nil
	})

	if err != nil {
		Logf("FetchLogsforPoseidon error : %v,", err)
	}

}

// FetchLogsforPoseidon will get the Poseidon and firmament deplyment logs
func (f *Framework) FetchLogsforFirmament() {
	matchLabel := map[string]string{
		"scheduler": "firmament",
	}

	// wait for namespace to delete or timeout.
	err := wait.PollImmediate(2*time.Second, 10*time.Minute, func() (bool, error) {
		// As poseidon and firmament run in kube-system namespaces we need to pass the correct ns here
		f.LogPodsWithLabels("kube-system", matchLabel, Logf)
		return false, nil
	})

	if err != nil {
		Logf("FetchLogsforFrimament error : %v,", err)
	}

}

func (f *Framework) LogPodsWithLabels(ns string, match map[string]string, logFunc func(ftm string, args ...interface{})) {
	podList, err := f.ClientSet.CoreV1().Pods(ns).List(metav1.ListOptions{LabelSelector: labels.SelectorFromSet(match).String()})
	if err != nil {
		logFunc("Error getting pods in namespace %q: %v", ns, err)
		return
	}
	logFunc("Running kubectl logs on pods with labels %v in %v", match, ns)
	for _, pod := range podList.Items {
		kubectlLogPod(f.ClientSet, pod, "", logFunc)
	}
}

func kubectlLogPod(c clientset.Interface, pod v1.Pod, containerNameSubstr string, logFunc func(ftm string, args ...interface{})) {
	for _, container := range pod.Spec.Containers {
		if strings.Contains(container.Name, containerNameSubstr) {
			logFunc("fetching logs for %v in pod %v", container.Name, pod.Name)
			// Contains() matches all strings if substr is empty
			logs, err := GetPodLogs(c, pod.Namespace, pod.Name, container.Name)
			if err != nil {
				logs, err = getPreviousPodLogs(c, pod.Namespace, pod.Name, container.Name)
				if err != nil {
					logFunc("Failed to get logs of pod %v, container %v, err: %v", pod.Name, container.Name, err)
				}
			}
			logFunc("Logs of %v/%v:%v on node %v", pod.Namespace, pod.Name, container.Name, pod.Spec.NodeName)
			logFunc("%s : STARTLOG\n%s\nENDLOG for container %v:%v:%v", containerNameSubstr, logs, pod.Namespace, pod.Name, container.Name)
		}
	}
}

func GetPodLogs(c clientset.Interface, namespace, podName, containerName string) (string, error) {
	return getPodLogsInternal(c, namespace, podName, containerName, false)
}

// utility function for gomega Eventually
func getPodLogsInternal(c clientset.Interface, namespace, podName, containerName string, previous bool) (string, error) {
	logs, err := c.CoreV1().RESTClient().Get().
		Resource("pods").
		Namespace(namespace).
		Name(podName).SubResource("log").
		Param("container", containerName).
		Param("previous", strconv.FormatBool(previous)).
		Do().
		Raw()
	if err != nil {
		return "", err
	}
	if err == nil && strings.Contains(string(logs), "Internal Error") {
		return "", fmt.Errorf("Fetched log contains \"Internal Error\": %q.", string(logs))
	}
	return string(logs), err
}

func getPreviousPodLogs(c clientset.Interface, namespace, podName, containerName string) (string, error) {
	return getPodLogsInternal(c, namespace, podName, containerName, true)
}

// logPodStates logs basic info of provided pods for debugging.
func logPodStates(pods []v1.Pod) {
	// Find maximum widths for pod, node, and phase strings for column printing.
	maxPodW, maxNodeW, maxPhaseW, maxGraceW := len("POD"), len("NODE"), len("PHASE"), len("GRACE")
	for i := range pods {
		pod := &pods[i]
		if len(pod.ObjectMeta.Name) > maxPodW {
			maxPodW = len(pod.ObjectMeta.Name)
		}
		if len(pod.Spec.NodeName) > maxNodeW {
			maxNodeW = len(pod.Spec.NodeName)
		}
		if len(pod.Status.Phase) > maxPhaseW {
			maxPhaseW = len(pod.Status.Phase)
		}
	}
	// Increase widths by one to separate by a single space.
	maxPodW++
	maxNodeW++
	maxPhaseW++
	maxGraceW++

	// Log pod info. * does space padding, - makes them left-aligned.
	Logf("%-[1]*[2]s %-[3]*[4]s %-[5]*[6]s %-[7]*[8]s %[9]s",
		maxPodW, "POD", maxNodeW, "NODE", maxPhaseW, "PHASE", maxGraceW, "GRACE", "CONDITIONS")
	for _, pod := range pods {
		grace := ""
		if pod.DeletionGracePeriodSeconds != nil {
			grace = fmt.Sprintf("%ds", *pod.DeletionGracePeriodSeconds)
		}
		Logf("%-[1]*[2]s %-[3]*[4]s %-[5]*[6]s %-[7]*[8]s %[9]s",
			maxPodW, pod.ObjectMeta.Name, maxNodeW, pod.Spec.NodeName, maxPhaseW, pod.Status.Phase, maxGraceW, grace, pod.Status.Conditions)
	}
	Logf("") // Final empty line helps for readability.
}

func countRemainingPods(c clientset.Interface, namespace string) (int, int, error) {
	// check for remaining pods
	pods, err := c.CoreV1().Pods(namespace).List(metav1.ListOptions{})
	if err != nil {
		return 0, 0, err
	}

	// nothing remains!
	if len(pods.Items) == 0 {
		return 0, 0, nil
	}

	// stuff remains, log about it
	logPodStates(pods.Items)

	// check if there were any pods with missing deletion timestamp
	numPods := len(pods.Items)
	missingTimestamp := 0
	for _, pod := range pods.Items {
		if pod.DeletionTimestamp == nil {
			missingTimestamp++
		}
	}
	return numPods, missingTimestamp, nil
}
