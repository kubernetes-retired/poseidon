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
	"fmt"
	"math/rand"
	"os"

	"github.com/golang/glog"
	"github.com/kubernetes-sigs/poseidon/test/e2e/framework"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var _ = Describe("Poseidon", func() {
	var clientset kubernetes.Interface
	var ns string //namespace string
	hostname, _ := os.Hostname()
	f := framework.NewDefaultFramework("sched-poseidon")

	glog.Info("Inside Poseidon tests for k8s:", hostname)

	BeforeEach(func() {
		clientset = f.ClientSet
		ns = f.Namespace.Name
		//This will list all pods in the namespace
		f.ListPodsInNamespace(ns)
	})

	AfterEach(func() {
		//fetch poseidon and firmament logs after each test case run
		f.FetchLogsFromFirmament(f.Namespace.Name)
		f.FetchLogsFromPoseidon(f.Namespace.Name)
		//This will list all pods in the namespace
		f.ListPodsInNamespace(f.Namespace.Name)
	})

	Describe("Add Pod using Poseidon scheduler", func() {
		glog.Info("Inside Check for adding pod using Poseidon scheduler")
		Context("using firmament for configuring pod", func() {
			name := fmt.Sprintf("test-nginx-pod-%d", rand.Uint32())

			It("should succeed deploying pod using firmament scheduler", func() {
				labels := make(map[string]string)
				labels["schedulerName"] = "poseidon"
				//Create a K8s Pod with poseidon
				pod, err := clientset.CoreV1().Pods(ns).Create(&v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:   name,
						Labels: labels,
					},
					Spec: v1.PodSpec{
						SchedulerName: "poseidon",
						Containers: []v1.Container{{
							Name:            fmt.Sprintf("container-%s", name),
							Image:           "nginx:latest",
							ImagePullPolicy: "IfNotPresent",
						}}},
				})

				Expect(err).NotTo(HaveOccurred())

				By("Waiting for the pod to have running status")
				f.WaitForPodRunning(pod.Name)
				//This will list all pods in the namespace
				f.ListPodsInNamespace(f.Namespace.Name)
				pod, err = clientset.CoreV1().Pods(ns).Get(name, metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred())
				glog.Info("pod status =", string(pod.Status.Phase))
				Expect(string(pod.Status.Phase)).To(Equal("Running"))

				By("Pod was in Running state... Time to delete the pod now...")
				err = clientset.CoreV1().Pods(ns).Delete(name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred())
				By("Check for pod deletion")
				_, err = clientset.CoreV1().Pods(ns).Get(name, metav1.GetOptions{})
				if err != nil {
					Expect(errors.IsNotFound(err)).To(Equal(true))
				}
			})
		})
	})

	Describe("Add Deployment using Poseidon scheduler", func() {
		glog.Info("Inside Check for adding Deployment using Poseidon scheduler")
		Context("using firmament for configuring Deployment", func() {
			name := fmt.Sprintf("test-nginx-deploy-%d", rand.Uint32())

			It("should succeed deploying Deployment using firmament scheduler", func() {
				// Create a K8s Deployment with poseidon scheduler
				var replicas int32
				replicas = 2
				deployment, err := clientset.ExtensionsV1beta1().Deployments(ns).Create(&v1beta1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{"app": "nginx"},
						Name:   name,
					},
					Spec: v1beta1.DeploymentSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"app": "nginx", "name": "test-dep"},
						},
						Replicas: &replicas,
						Template: v1.PodTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Labels: map[string]string{"name": "test-dep", "app": "nginx", "schedulerName": "poseidon"},
								Name:   name,
							},
							Spec: v1.PodSpec{
								SchedulerName: "poseidon",
								Containers: []v1.Container{
									{
										Name:            fmt.Sprintf("container-%s", name),
										Image:           "nginx:latest",
										ImagePullPolicy: "IfNotPresent",
									},
								},
							},
						},
					},
				})

				Expect(err).NotTo(HaveOccurred())

				By("Waiting for the Deployment to have running status")
				f.WaitForDeploymentComplete(deployment)
				//This will list all pods in the namespace
				f.ListPodsInNamespace(f.Namespace.Name)
				deployment, err = clientset.ExtensionsV1beta1().Deployments(ns).Get(name, metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred())

				By(fmt.Sprintf("Creation of deployment %q in namespace %q succeeded.  Deleting deployment.", deployment.Name, ns))
				Expect(deployment.Status.Replicas).To(Equal(deployment.Status.AvailableReplicas))

				By("Pod was in Running state... Time to delete the deployment now...")
				err = f.WaitForDeploymentDelete(deployment)
				Expect(err).NotTo(HaveOccurred())
				By("Waiting 5 seconds")
				By("Check for deployment deletion")
				_, err = clientset.ExtensionsV1beta1().Deployments(ns).Get(name, metav1.GetOptions{})
				if err != nil {
					Expect(errors.IsNotFound(err)).To(Equal(true))
				}
			})
		})
	})

	Describe("Add ReplicaSet using Poseidon scheduler", func() {
		glog.Info("Inside Check for adding ReplicaSet using Poseidon scheduler")
		Context("using firmament for configuring ReplicaSet", func() {
			name := fmt.Sprintf("test-nginx-rs-%d", rand.Uint32())

			It("should succeed deploying ReplicaSet using firmament scheduler", func() {
				//Create a K8s ReplicaSet with poseidon scheduler
				var replicas int32
				replicas = 3
				replicaSet, err := clientset.ExtensionsV1beta1().ReplicaSets(ns).Create(&v1beta1.ReplicaSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: name,
					},
					Spec: v1beta1.ReplicaSetSpec{
						Replicas: &replicas,
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"name": "test-rs"},
						},
						Template: v1.PodTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Labels: map[string]string{"name": "test-rs", "schedulerName": "poseidon"},
								Name:   name,
							},
							Spec: v1.PodSpec{
								SchedulerName: "poseidon",
								Containers: []v1.Container{
									{
										Name:            fmt.Sprintf("container-%s", name),
										Image:           "nginx:latest",
										ImagePullPolicy: "IfNotPresent",
									},
								},
							},
						},
					},
				})

				Expect(err).NotTo(HaveOccurred())

				By("Waiting for the ReplicaSet to have running status")
				f.WaitForReadyReplicaSet(replicaSet.Name)
				//This will list all pods in the namespace
				f.ListPodsInNamespace(f.Namespace.Name)
				replicaSet, err = clientset.ExtensionsV1beta1().ReplicaSets(ns).Get(replicaSet.Name, metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred())

				By(fmt.Sprintf("Creation of ReplicaSet %q in namespace %q succeeded.  Deleting ReplicaSet.", replicaSet.Name, ns))
				Expect(replicaSet.Status.Replicas).To(Equal(replicaSet.Status.AvailableReplicas))
				By("Pod was in Running state... Time to delete the ReplicaSet now...")
				err = f.WaitForReplicaSetDelete(replicaSet)
				Expect(err).NotTo(HaveOccurred())
				By("Waiting 5 seconds")
				By("Check for ReplicaSet deletion")
				_, err = clientset.ExtensionsV1beta1().ReplicaSets(ns).Get(name, metav1.GetOptions{})
				if err != nil {
					Expect(errors.IsNotFound(err)).To(Equal(true))
				}
			})
		})
	})

	Describe("Add Job using Poseidon scheduler", func() {
		glog.Info("Inside Check for adding Job using Poseidon scheduler")
		Context("using firmament for configuring Job", func() {
			name := fmt.Sprintf("test-nginx-job-%d", rand.Uint32())

			It("should succeed deploying Job using firmament scheduler", func() {
				labels := make(map[string]string)
				labels["name"] = "test-job"
				//Create a K8s Job with poseidon scheduler
				var parallelism int32 = 2
				job, err := clientset.BatchV1().Jobs(ns).Create(&batchv1.Job{
					ObjectMeta: metav1.ObjectMeta{
						Name:   name,
						Labels: labels,
					},
					Spec: batchv1.JobSpec{
						Parallelism: &parallelism,
						Completions: &parallelism,
						Template: v1.PodTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Labels: labels,
							},
							Spec: v1.PodSpec{
								SchedulerName: "poseidon",
								Containers: []v1.Container{
									{
										Name:            fmt.Sprintf("container-%s", name),
										Image:           "nginx:latest",
										ImagePullPolicy: "IfNotPresent",
									},
								},
								RestartPolicy: "Never",
							},
						},
					},
				})

				Expect(err).NotTo(HaveOccurred())

				By("Waiting for the Job to have running status")
				f.WaitForAllJobPodsRunning(job.Name, parallelism)
				//This will list all pods in the namespace
				f.ListPodsInNamespace(f.Namespace.Name)

				job, err = clientset.BatchV1().Jobs(ns).Get(name, metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred())

				By(fmt.Sprintf("Creation of Jobs %q in namespace %q succeeded.  Deleting Job.", job.Name, ns))
				Expect(job.Status.Active).To(Equal(parallelism))

				By("Job was in Running state... Time to delete the Job now...")
				err = f.WaitForJobDelete(job.Name)
				Expect(err).NotTo(HaveOccurred())
				By("Check for Job deletion")
				_, err = clientset.BatchV1().Jobs(ns).Get(name, metav1.GetOptions{})
				if err != nil {
					Expect(errors.IsNotFound(err)).To(Equal(true))
				}
				Expect("Success").To(Equal("Success"))
			})
		})
	})

	Describe("Poseidon [Predicates]", func() {
		// This test verifies we don't allow scheduling of pods in a way that sum of
		// limits of pods is greater than machines capacity.
		// It assumes that cluster add-on pods stay stable and cannot be run in parallel
		// with any other test that touches Nodes or Pods.
		// It is so because we need to have precise control on what's running in the cluster.
		// Test scenario:
		// 1. Find the amount CPU resources on each node.
		// 2. Create one pod with affinity to each node that uses 70% of the node CPU.
		// 3. Wait for the pods to be scheduled.
		// 4. Create another pod with no affinity to any node that need 50% of the largest node CPU.
		// 5. Make sure this additional pod is not scheduled.
		/*
			    Testname: scheduler-resource-limits
			    Description: Ensure that scheduler accounts node resources correctly
				and respects pods' resource requirements during scheduling.
		*/
		It("should validates resource limits of pods that are allowed to run ", func() {
			nodeMaxAllocatable := int64(0)
			nodeToAllocatableMap := make(map[string]int64)
			nodeList, err := clientset.CoreV1().Nodes().List(metav1.ListOptions{})
			Expect(err).NotTo(HaveOccurred())
			for _, node := range nodeList.Items {
				// Skip the master node.
				if IsMasterNode(node.Name) {
					continue
				}
				nodeReady := false
				for _, condition := range node.Status.Conditions {
					if condition.Type == v1.NodeReady && condition.Status == v1.ConditionTrue {
						nodeReady = true
						break
					}
				}
				// Skip the unready node.
				if !nodeReady {
					continue
				}
				// Find allocatable amount of CPU.
				allocatable, found := node.Status.Allocatable[v1.ResourceCPU]
				Expect(found).To(Equal(true))
				nodeToAllocatableMap[node.Name] = allocatable.MilliValue()
				if nodeMaxAllocatable < allocatable.MilliValue() {
					nodeMaxAllocatable = allocatable.MilliValue()
				}
			}

			pods, err := clientset.CoreV1().Pods(metav1.NamespaceAll).List(metav1.ListOptions{})
			framework.ExpectNoError(err)
			for _, pod := range pods.Items {
				_, found := nodeToAllocatableMap[pod.Spec.NodeName]
				if found && pod.Status.Phase != v1.PodSucceeded && pod.Status.Phase != v1.PodFailed {
					framework.Logf("Pod %v requesting resource cpu=%vm on Node %v", pod.Name, getRequestedCPU(pod), pod.Spec.NodeName)
					nodeToAllocatableMap[pod.Spec.NodeName] -= getRequestedCPU(pod)
				}
			}

			By("Starting Pods to consume most of the cluster CPU.")
			// Create one pod per node that requires 70% of the node remaining CPU.
			fillerPods := []*v1.Pod{}
			for nodeName, cpu := range nodeToAllocatableMap {
				requestedCPU := cpu * 7 / 10
				framework.Logf("node [%s] cpu [%v] request [%v]", nodeName, cpu, requestedCPU)
				fillerPods = append(fillerPods, createTestPod(f, testPodConfig{
					Name: fmt.Sprintf("filler-pod-%d", rand.Uint32()),
					Resources: &v1.ResourceRequirements{
						Limits: v1.ResourceList{
							v1.ResourceCPU: *resource.NewMilliQuantity(requestedCPU, "DecimalSI"),
						},
						Requests: v1.ResourceList{
							v1.ResourceCPU: *resource.NewMilliQuantity(requestedCPU, "DecimalSI"),
						},
					},
					// TODO: We need to set node level affinity.
					SchedulerName: "poseidon",
				}))
			}
			// Wait for filler pods to be scheduled.
			for _, pod := range fillerPods {
				framework.ExpectNoError(framework.WaitForPodRunningInNamespace(clientset, pod))
			}

			// Clean up filler pods after this test.
			defer func() {
				for _, pod := range fillerPods {
					framework.Logf("Time to clean up the pod [%s] now...", pod.Name)
					err = clientset.CoreV1().Pods(ns).Delete(pod.Name, &metav1.DeleteOptions{})
					Expect(err).NotTo(HaveOccurred())
				}
			}()

			By("Creating another pod that requires unavailable amount of CPU.")
			// Create another pod that requires 50% of the largest node CPU resources.
			// This pod should remain pending as at least 70% of CPU of other nodes in
			// the cluster are already consumed.
			podName := "additional-pod"
			conf := testPodConfig{
				Name:   podName,
				Labels: map[string]string{"name": "additional"},
				Resources: &v1.ResourceRequirements{
					Limits: v1.ResourceList{
						v1.ResourceCPU: *resource.NewMilliQuantity(nodeMaxAllocatable*5/10, "DecimalSI"),
					},
				},
				SchedulerName: "poseidon",
			}
			additionalPod := createTestPod(f, conf)
			// Clean up additional pod after this test.
			defer func() {
				framework.Logf("Time to clean up the pod [%s] now...", additionalPod.Name)
				err = clientset.CoreV1().Pods(ns).Delete(additionalPod.Name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred())
			}()
			err = framework.WaitForPodRunningInNamespace(clientset, additionalPod)
			Expect(err).To(HaveOccurred())
		})

		// Test Nodes does not have any label, hence it should be impossible to schedule Pod with
		// nonempty Selector set.
		/*
			    Testname: scheduler-node-selector-not-matching
			    Description: Ensure that scheduler respects the NodeSelector field of
				PodSpec during scheduling (when it does not match any node).
		*/
		It("validates that NodeSelector is respected if not matching ", func() {
			By("Trying to schedule Pod with nonempty NodeSelector.")
			podName := "restricted-pod"
			conf := testPodConfig{
				Name:   podName,
				Labels: map[string]string{"name": "restricted"},
				NodeSelector: map[string]string{
					"label": "nonempty",
				},
			}
			testPod := createTestPod(f, conf)
			// Clean up additional pod after this test.
			defer func() {
				framework.Logf("Time to clean up the pod [%s] now...", testPod.Name)
				err := clientset.CoreV1().Pods(ns).Delete(testPod.Name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred())
			}()
			err := framework.WaitForPodRunningInNamespace(clientset, testPod)
			Expect(err).To(HaveOccurred())
		})

		/*
			    Testname: scheduler-node-selector-matching
			    Description: Ensure that scheduler respects the NodeSelector field
				of PodSpec during scheduling (when it matches).
		*/
		It("validates that NodeSelector is respected if matching ", func() {
			// Randomly pick a node
			nodeName := getNodeThatCanRunPodWithoutToleration(f)

			By("Trying to apply a random label on the found node.")
			k := fmt.Sprintf("kubernetes.io/e2e-%d", rand.Uint32())
			v := "42"

			framework.AddOrUpdateLabelOnNode(clientset, nodeName, k, v)
			framework.ExpectNodeHasLabel(clientset, nodeName, k, v)

			By("Trying to relaunch the pod, now with labels.")
			labelPodName := "with-labels"
			createTestPod(f, testPodConfig{
				Name: labelPodName,
				NodeSelector: map[string]string{
					k: v,
				},
			})

			// check that pod got scheduled. We intentionally DO NOT check that the
			// pod is running because this will create a race condition with the
			// kubelet and the scheduler: the scheduler might have scheduled a pod
			// already when the kubelet does not know about its new label yet. The
			// kubelet will then refuse to launch the pod.
			framework.ExpectNoError(framework.WaitForPodNotPending(clientset, ns, labelPodName))
			labelPod, err := clientset.CoreV1().Pods(ns).Get(labelPodName, metav1.GetOptions{})
			framework.ExpectNoError(err)
			Expect(labelPod.Spec.NodeName).To(Equal(nodeName))
		})
	})
})

func getNodeThatCanRunPodWithoutToleration(f *framework.Framework) string {
	By("Trying to launch a pod without a toleration to get a node which can launch it.")
	return runPodAndGetNodeName(f, testPodConfig{Name: "without-toleration"})
}

func runPodAndGetNodeName(f *framework.Framework, conf testPodConfig) string {
	// launch a pod to find a node which can launch a pod. We intentionally do
	// not just take the node list and choose the first of them. Depending on the
	// cluster and the scheduler it might be that a "normal" pod cannot be
	// scheduled onto it.
	pod := runPausePod(f, conf)

	By("Explicitly delete pod here to free the resource it takes.")
	err := f.ClientSet.CoreV1().Pods(f.Namespace.Name).Delete(pod.Name, metav1.NewDeleteOptions(0))
	framework.ExpectNoError(err)

	return pod.Spec.NodeName
}

func runPausePod(f *framework.Framework, conf testPodConfig) *v1.Pod {
	pod := createTestPod(f, conf)
	framework.ExpectNoError(framework.WaitForPodRunningInNamespace(f.ClientSet, pod))
	pod, err := f.ClientSet.CoreV1().Pods(f.Namespace.Name).Get(conf.Name, metav1.GetOptions{})
	framework.ExpectNoError(err)
	return pod
}
