package framework

import (
	"fmt"
	"strings"
	"strconv"


	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

// FetchLogsFromPoseidon will get the Poseidon and firmament deplyment logs
func (f *Framework) FetchLogsFromPoseidon(nsName string) {
	matchLabel := map[string]string{
		"component": "poseidon",
	}
	f.LogPodsWithOptionalLabels(nsName, matchLabel, Logf)
}

// FetchLogsFromPoseidon will get the Poseidon and firmament deplyment logs
func (f *Framework) FetchLogsFromFirmament(nsName string) {
	matchLabel := map[string]string{
		"scheduler": "firmament",
	}
	f.LogPodsWithOptionalLabels(nsName, matchLabel, Logf)
}

// FetchLogsForAllPodsInNamespace will fetch logs from all the pods in the namespace
// This can be called after each test is run
func (f* Framework) FetchLogsForAllPodsInNamespace(nsName string){

	// As poseidon and firmament run in kube-system namespaces we need to pass the correct ns here
	f.LogPodsWithOptionalLabels(nsName,nil , Logf)
}

// ListPodsInNamespace will list all the pods in a given namespaces
func (f *Framework) ListPodsInNamespace(ns string){
	podList, err := f.ClientSet.CoreV1().Pods(ns).List(metav1.ListOptions{})
	if err!=nil{
		Logf("Error listing pods in namespaces %v", err)
		return
	}
	logPodStates(podList.Items)
}
func (f *Framework) LogPodsWithOptionalLabels(ns string, match map[string]string, logFunc func(ftm string, args ...interface{})) {
	var podList *v1.PodList
	var err error
	if match!=nil || len(match)>0{
		podList, err = f.ClientSet.CoreV1().Pods(ns).List(metav1.ListOptions{LabelSelector: labels.SelectorFromSet(match).String()})

	}else{
		podList, err = f.ClientSet.CoreV1().Pods(ns).List(metav1.ListOptions{})
	}

	if err != nil {
		logFunc("Error getting pods in namespace %q: %v", ns, err)
		return
	}
	if len(podList.Items) > 0 {
		if match!=nil || len(match)>0 {
			logFunc("Running kubectl logs on pods with labels %v in %v namespace", match, ns)
		}else{
			logFunc("Running kubectl logs on pods in %v namespace",ns)
		}
		for _, pod := range podList.Items {
			kubectlLogPod(f.ClientSet, pod, "", logFunc)
		}
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
	maxPodW, maxNodeW, maxPhaseW, maxGraceW, maxNamespaceW := len("POD"), len("NODE"), len("PHASE"), len("GRACE"), len("Namespace")
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
		if len(pod.Namespace) > maxNamespaceW{
			maxNamespaceW =len(pod.Namespace)
		}
	}
	// Increase widths by one to separate by a single space.
	maxPodW++
	maxNodeW++
	maxPhaseW++
	maxGraceW++
	maxNamespaceW++

	// Log pod info. * does space padding, - makes them left-aligned.
	Logf("%-[1]*[2]s %-[3]*[4]s %-[5]*[6]s %-[7]*[8]s %-[9]*[10]s %[11]s",
		maxNamespaceW, "Namespace", maxPodW, "POD", maxNodeW, "NODE", maxPhaseW, "PHASE", maxGraceW, "GRACE", "CONDITIONS")
	for _, pod := range pods {
		grace := ""
		if pod.DeletionGracePeriodSeconds != nil {
			grace = fmt.Sprintf("%ds", *pod.DeletionGracePeriodSeconds)
		}
		Logf("%-[1]*[2]s %-[3]*[4]s %-[5]*[6]s %-[7]*[8]s %-[9]*[10]s %[11]s",
			maxNamespaceW,pod.Namespace,maxPodW, pod.ObjectMeta.Name, maxNodeW, pod.Spec.NodeName, maxPhaseW, pod.Status.Phase, maxGraceW, grace, pod.Status.Conditions)
	}
	Logf("") // Final empty line helps for readability.
}
