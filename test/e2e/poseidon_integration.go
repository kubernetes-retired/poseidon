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
	"time"
)

var _ = Describe("Poseidon", func() {
	var clientset kubernetes.Interface
	var ns string // namespace string
	hostname, _ := os.Hostname()
	f := framework.NewDefaultFramework("sched-poseidon")

	glog.Info("Inside Poseidon tests for k8s:", hostname)

	BeforeEach(func() {
		clientset = f.ClientSet
		ns = f.Namespace.Name
		// This will list all pods in the namespace
		f.ListPodsInNamespace(ns)
	})

	AfterEach(func() {
		// fetch poseidon and firmament logs after each test case run
		f.FetchLogsFromFirmament(f.Namespace.Name)
		f.FetchLogsFromPoseidon(f.Namespace.Name)
		// This will list all pods in the namespace
		f.ListPodsInNamespace(f.Namespace.Name)
	})

	Describe("Add Pod using Poseidon scheduler", func() {
		glog.Info("Inside Check for adding pod using Poseidon scheduler")
		Context("using firmament for configuring pod", func() {
			name := fmt.Sprintf("test-nginx-pod-%d", rand.Uint32())

			It("should succeed deploying pod using firmament scheduler", func() {
				labels := make(map[string]string)
				labels["schedulerName"] = "poseidon"
				// Create a K8s Pod with poseidon
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
				err = f.WaitForPodRunning(pod.Name)
				Expect(err).NotTo(HaveOccurred())
				// This will list all pods in the namespace
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
				err = f.WaitForDeploymentComplete(deployment)
				Expect(err).NotTo(HaveOccurred())
				// This will list all pods in the namespace
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
				// Create a K8s ReplicaSet with poseidon scheduler
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
				err = f.WaitForReadyReplicaSet(replicaSet.Name)
				Expect(err).NotTo(HaveOccurred())
				// This will list all pods in the namespace
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
				// Create a K8s Job with poseidon scheduler
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
				err = f.WaitForAllJobPodsRunning(job.Name, parallelism)
				Expect(err).NotTo(HaveOccurred())
				// This will list all pods in the namespace
				f.ListPodsInNamespace(f.Namespace.Name)

				job, err = clientset.BatchV1().Jobs(ns).Get(name, metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred())

				By(fmt.Sprintf("Creation of Jobs %q in namespace %q succeeded.  Deleting Job.", job.Name, ns))
				Expect(job.Status.Active).To(Equal(parallelism))

				By("Job was in Running state... Time to delete the Job now...")
				err = f.WaitForJobDelete(job)
				Expect(err).NotTo(HaveOccurred())
				By("Check for Job deletion")
				_, err = clientset.BatchV1().Jobs(ns).Get(name, metav1.GetOptions{})
				if err != nil {
					Expect(errors.IsNotFound(err)).To(Equal(true))
				}
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
		// 1. Find the amount CPU/Memory resources on each node.
		// 2. Create one pod with affinity to each node that uses 70% of the node CPU/Memory.
		// 3. Wait for the pods to be scheduled.
		// 4. Create another pod with no affinity to any node that need 50% of the largest node CPU/Memory.
		// 5. Make sure this additional pod is not scheduled.

		//	    Testname: scheduler-resource-limits
		//	    Description: Ensure that scheduler accounts node resources correctly
		//		and respects pods' resource requirements during scheduling.

		It("should validates resource limits of pods that are allowed to run ", func() {
			nodeMaxAllocatable := int64(0)
			nodeMaxAllocatableMem := int64(0)
			nodeToAllocatableMap := make(map[string]int64)
			nodeToAllocatableMemMap := make(map[string]int64)
			nodeList, err := clientset.CoreV1().Nodes().List(metav1.ListOptions{})
			Expect(err).NotTo(HaveOccurred())
			for _, node := range nodeList.Items {
				// Skip the master node.
				if framework.IsMasterNode(node.Name) {
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

				// Find allocatable amount of Memory.
				allocatableMem, found := node.Status.Allocatable[v1.ResourceMemory]
				Expect(found).To(Equal(true))
				nodeToAllocatableMemMap[node.Name] = allocatableMem.Value()
				if nodeMaxAllocatableMem < allocatableMem.Value() {
					nodeMaxAllocatableMem = allocatableMem.Value()
				}
			}

			pods, err := clientset.CoreV1().Pods(metav1.NamespaceAll).List(metav1.ListOptions{})
			framework.ExpectNoError(err)
			for _, pod := range pods.Items {
				_, found := nodeToAllocatableMap[pod.Spec.NodeName]
				_, foundMem := nodeToAllocatableMemMap[pod.Spec.NodeName]
				if found && foundMem && pod.Status.Phase != v1.PodSucceeded && pod.Status.Phase != v1.PodFailed {
					framework.Logf("Pod %v requesting resource cpu=%vm memory=%vKi on Node %v", pod.Name, getRequestedCPU(pod), getRequestedMemory(pod), pod.Spec.NodeName)
					nodeToAllocatableMap[pod.Spec.NodeName] -= getRequestedCPU(pod)
					nodeToAllocatableMemMap[pod.Spec.NodeName] -= getRequestedMemory(pod)
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
							v1.ResourceCPU: *resource.NewMilliQuantity(requestedCPU, resource.DecimalSI),
						},
						Requests: v1.ResourceList{
							v1.ResourceCPU: *resource.NewMilliQuantity(requestedCPU, resource.DecimalSI),
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

			By("Starting Pods to consume most of the cluster Memory.")
			// Create one pod per node that requires 70% of the node remaining Memory.
			fillerPodsMem := []*v1.Pod{}
			for nodeName, memory := range nodeToAllocatableMemMap {
				requestedMem := memory * 7 / 10
				framework.Logf("node [%s] memory [%v] request [%v]", nodeName, memory, requestedMem)
				fillerPodsMem = append(fillerPodsMem, createTestPod(f, testPodConfig{
					Name: fmt.Sprintf("filler-pod-m-%d", rand.Uint32()),
					Resources: &v1.ResourceRequirements{
						Limits: v1.ResourceList{
							v1.ResourceMemory: *resource.NewQuantity(requestedMem, resource.BinarySI),
						},
						Requests: v1.ResourceList{
							v1.ResourceMemory: *resource.NewQuantity(requestedMem, resource.BinarySI),
						},
					},
					// TODO: We need to set node level affinity.
					SchedulerName: "poseidon",
				}))
			}
			// Wait for filler pods to be scheduled.
			for _, pod := range fillerPodsMem {
				framework.ExpectNoError(framework.WaitForPodRunningInNamespace(clientset, pod))
			}

			// Clean up filler pods after this test.
			defer func() {
				for _, pod := range fillerPodsMem {
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
						v1.ResourceCPU: *resource.NewMilliQuantity(nodeMaxAllocatable*5/10, resource.DecimalSI),
					},
				},
				SchedulerName: "poseidon",
			}
			additionalPod := createTestPod(f, conf)
			defer func() {
				framework.Logf("Time to clean up the pod [%s] now...", additionalPod.Name)
				err = clientset.CoreV1().Pods(ns).Delete(additionalPod.Name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred())
			}()
			// we wait only for 2 minuted to check if the pod can be scheduled here, since this pod will never be scheduled
			err = framework.WaitTimeoutForPodRunningInNamespace(clientset, additionalPod.Name, additionalPod.Namespace, time.Minute*2)
			Expect(err).To(HaveOccurred())

			By("Creating another pod that requires unavailable amount of Memory.")
			// Create another pod that requires 50% of the largest node Memory resources.
			// This pod should remain pending as at least 70% of Memory of other nodes in
			// the cluster are already consumed.
			podNameMem := "additional-pod-m"
			confMem := testPodConfig{
				Name:   podNameMem,
				Labels: map[string]string{"name": "additional-m"},
				Resources: &v1.ResourceRequirements{
					Limits: v1.ResourceList{
						v1.ResourceMemory: *resource.NewQuantity(nodeMaxAllocatableMem*5/10, resource.BinarySI),
					},
				},
				SchedulerName: "poseidon",
			}
			additionalPodMem := createTestPod(f, confMem)
			// Clean up additional pod after this test.
			defer func() {
				framework.Logf("Time to clean up the pod [%s] now...", additionalPodMem.Name)
				err = clientset.CoreV1().Pods(ns).Delete(additionalPodMem.Name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred())
			}()
			// we wait only for 2 minuted to check if the pod can be scheduled here, since this pod will never be scheduled
			err = framework.WaitTimeoutForPodRunningInNamespace(clientset, additionalPodMem.Name, additionalPodMem.Namespace, time.Minute*2)
			Expect(err).To(HaveOccurred())
		})

		// Test Nodes does not have any label, hence it should be impossible to schedule Pod with
		// nonempty Selector set.
		//
		//	    Testname: scheduler-node-selector-not-matching
		//	    Description: Ensure that scheduler respects the NodeSelector field of
		//		PodSpec during scheduling (when it does not match any node).
		//
		It("validates that NodeSelector is respected if not matching ", func() {
			By("Trying to schedule Pod with nonempty NodeSelector.")
			podName := "restricted-pod"
			conf := testPodConfig{
				Name:   podName,
				Labels: map[string]string{"name": "restricted"},
				NodeSelector: map[string]string{
					"label": "nonempty",
				},
				SchedulerName: "poseidon",
			}
			testPod := createTestPod(f, conf)
			// Clean up additional pod after this test.
			defer func() {
				framework.Logf("Time to clean up the pod [%s] now...", testPod.Name)
				err := clientset.CoreV1().Pods(ns).Delete(testPod.Name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred())
			}()
			// wait only for 2 minutes as we know this pod wont be scheduled.
			err := framework.WaitTimeoutForPodNotPending(clientset, ns, testPod.Name, time.Minute*2)
			Expect(err).To(HaveOccurred())
		})

		//
		//	    Testname: scheduler-node-selector-matching
		//	    Description: Ensure that scheduler respects the NodeSelector field
		//		of PodSpec during scheduling (when it matches).
		//
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
			testPod := createTestPod(f, testPodConfig{
				Name: labelPodName,
				NodeSelector: map[string]string{
					k: v,
				},
				SchedulerName: "poseidon",
			})

			// Clean up additional pod after this test.
			defer func() {
				framework.Logf("Time to clean up the pod [%s] now...", testPod.Name)
				err := clientset.CoreV1().Pods(ns).Delete(testPod.Name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred())
			}()

			// check that pod got scheduled. We intentionally DO NOT check that the
			// pod is running because this will create a race condition with the
			// kubelet and the scheduler: the scheduler might have scheduled a pod
			// already when the kubelet does not know about its new label yet. The
			// kubelet will then refuse to launch the pod.
			By("validate if pods match the node label")
			framework.ExpectNoError(framework.WaitForPodNotPending(clientset, ns, labelPodName))
			labelPod, err := clientset.CoreV1().Pods(ns).Get(labelPodName, metav1.GetOptions{})
			framework.ExpectNoError(err)
			Expect(labelPod.Spec.NodeName).To(Equal(nodeName))

			By("Remove the node label")
			framework.RemoveLabelOffNode(clientset, nodeName, k)
			err = framework.VerifyLabelsRemoved(clientset, nodeName, []string{k})
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("Poseidon [NodeAffinity hard-constraint]", func() {
		labelPodName := "with-nodeaffinity-hard"
		testpod := testPodConfig{
			Name: labelPodName,
			Affinity: &v1.Affinity{
				NodeAffinity: &v1.NodeAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
						NodeSelectorTerms: []v1.NodeSelectorTerm{
							{
								MatchExpressions: []v1.NodeSelectorRequirement{
									{
										Key:      "gpu",
										Operator: v1.NodeSelectorOpIn,
										Values: []string{
											"nvidia",
											"zotac-nvidia",
										},
									},
								},
							},
						},
					},
				},
			},
			SchedulerName: "poseidon",
		}
		It("validates scheduler respect's a pod with requiredDuringSchedulingIgnoredDuringExecution constraint", func() {

			By("Trying to get a schedulable node")
			schedulableNodes := framework.ListSchedulableNodes(clientset)
			if len(schedulableNodes) < 1 {
				Skip(fmt.Sprintf("Skipping this test case as the required minimum nodes not avaliable "))
			}
			nodeOne := schedulableNodes[0]

			By("Trying to apply a label on the found node.")
			k := "gpu"
			v := "nvidia"
			framework.AddOrUpdateLabelOnNode(clientset, nodeOne.Name, k, v)
			framework.ExpectNodeHasLabel(clientset, nodeOne.Name, k, v)

			By("Trying to launch the pod, now with requiredDuringSchedulingIgnoredDuringExecution constraint")
			createTestPod(f, testpod)

			By("Validate if the pod is running in the node having the label")
			framework.ExpectNoError(framework.WaitForPodNotPending(clientset, ns, labelPodName))
			labelPod, err := clientset.CoreV1().Pods(ns).Get(labelPodName, metav1.GetOptions{})
			framework.ExpectNoError(err)
			Expect(labelPod.Spec.NodeName).To(Equal(nodeOne.Name))

			By("Remove the node label")
			framework.RemoveLabelOffNode(clientset, nodeOne.Name, k)
			err = framework.VerifyLabelsRemoved(clientset, nodeOne.Name, []string{k})
			Expect(err).NotTo(HaveOccurred())

			By("Delete the pod")
			err = clientset.CoreV1().Pods(ns).Delete(labelPod.Name, &metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred())
			err = f.WaitForPodNotFound(labelPodName, 2*time.Minute)
			Expect(err).NotTo(HaveOccurred())

		})

		It("validates that a pod with requiredDuringSchedulingIgnoredDuringExecution stays pending if no node match the constraint", func() {

			By("Re-create a pod with requiredDuringSchedulingIgnoredDuringExecution constraint")

			createTestPod(f, testpod)
			// wait only for 2 minutes, as we know this pod wont be scheduled
			err := framework.WaitTimeoutForPodNotPending(clientset, ns, labelPodName, time.Minute*2)
			Expect(err).To(HaveOccurred())

			By("Delete the pod")
			err = clientset.CoreV1().Pods(ns).Delete(labelPodName, &metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred())
			err = f.WaitForPodNotFound(labelPodName, 2*time.Minute)
			Expect(err).NotTo(HaveOccurred())

		})
	})

	Describe("Poseidon [NodeAffinity soft-constraint]", func() {
		labelPodName := "with-nodeaffinity-soft"
		k := "drive"
		v := "ssd"

		testpod := testPodConfig{
			Name: labelPodName,
			Affinity: &v1.Affinity{
				NodeAffinity: &v1.NodeAffinity{
					PreferredDuringSchedulingIgnoredDuringExecution: []v1.PreferredSchedulingTerm{
						{
							Weight: 10,
							Preference: v1.NodeSelectorTerm{
								MatchExpressions: []v1.NodeSelectorRequirement{
									{
										Key:      k,
										Operator: v1.NodeSelectorOpIn,
										Values: []string{
											v,
										},
									},
								},
							},
						},
					},
				},
			},
			SchedulerName: "poseidon",
		}

		It("validates scheduler respect's a pod with preferredDuringSchedulingIgnoredDuringExecution constraint if a node satisfy ", func() {
			By("Trying to get a schedulable node")
			schedulableNodes := framework.ListSchedulableNodes(clientset)
			if len(schedulableNodes) < 1 {
				Skip(fmt.Sprintf("Skipping this test case as the required minimum nodes not avaliable "))
			}
			nodeOne := schedulableNodes[0]

			By("Trying to apply a label on the found node.")

			framework.AddOrUpdateLabelOnNode(clientset, nodeOne.Name, k, v)
			framework.ExpectNodeHasLabel(clientset, nodeOne.Name, k, v)

			By("Trying to launch the pod, now with requiredDuringSchedulingIgnoredDuringExecution constraint.")
			createTestPod(f, testpod)

			By("Validate if the pod is running in the node having the preferred label")
			framework.ExpectNoError(framework.WaitForPodNotPending(clientset, ns, labelPodName))
			labelPod, err := clientset.CoreV1().Pods(ns).Get(labelPodName, metav1.GetOptions{})
			framework.ExpectNoError(err)
			Expect(labelPod.Spec.NodeName).To(Equal(nodeOne.Name))

			By("Remove the node label")
			framework.RemoveLabelOffNode(clientset, nodeOne.Name, k)
			err = framework.VerifyLabelsRemoved(clientset, nodeOne.Name, []string{k})
			Expect(err).NotTo(HaveOccurred())

			By("Delete the pod")
			err = clientset.CoreV1().Pods(ns).Delete(labelPodName, &metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred())
			err = f.WaitForPodNotFound(labelPodName, 2*time.Minute)
			Expect(err).NotTo(HaveOccurred())

		})

		It("validates scheduler schedules a pod even if preferredDuringSchedulingIgnoredDuringExecution constraint is not satisfied by any node", func() {

			By("re-create a pod, now with requiredDuringSchedulingIgnoredDuringExecution constraint.")
			createTestPod(f, testpod)

			By("Validate if the pod is running")
			framework.ExpectNoError(framework.WaitForPodNotPending(clientset, ns, labelPodName))
			_, err := clientset.CoreV1().Pods(ns).Get(labelPodName, metav1.GetOptions{})
			framework.ExpectNoError(err)

			By("Delete the pod")
			err = clientset.CoreV1().Pods(ns).Delete(labelPodName, &metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred())
			err = f.WaitForPodNotFound(labelPodName, 2*time.Minute)
			Expect(err).NotTo(HaveOccurred())

		})
	})

	Describe("Poseidon [NodeAffinity hard and soft constraint]", func() {
		var nodeOne, nodeTwo v1.Node
		labelPodName := "with-nodeaffinity-hard-soft"
		testpod := testPodConfig{
			Name: labelPodName,
			Affinity: &v1.Affinity{
				NodeAffinity: &v1.NodeAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
						NodeSelectorTerms: []v1.NodeSelectorTerm{
							{
								MatchExpressions: []v1.NodeSelectorRequirement{
									{
										Key:      "gpu",
										Operator: v1.NodeSelectorOpIn,
										Values: []string{
											"nvidia",
											"zotac-nvidia",
										},
									},
								},
							},
						},
					},
					PreferredDuringSchedulingIgnoredDuringExecution: []v1.PreferredSchedulingTerm{
						{
							Weight: 10,
							Preference: v1.NodeSelectorTerm{
								MatchExpressions: []v1.NodeSelectorRequirement{
									{
										Key:      "drive",
										Operator: v1.NodeSelectorOpIn,
										Values: []string{
											"ssd",
										},
									},
								},
							},
						},
					},
				},
			},
			SchedulerName: "poseidon",
		}
		It("validates if scheduler respect's a pod's with both hard and soft constraint", func() {

			By("Trying to get a schedulable node")
			schedulableNodes := framework.ListSchedulableNodes(clientset)
			if len(schedulableNodes) < 2 {
				Skip(fmt.Sprintf("Skipping this test case as this requires minimum of two node and only %d nodes avaliable", len(schedulableNodes)))
			}
			nodeOne = schedulableNodes[0]
			nodeTwo = schedulableNodes[1]

			By("Trying to apply a label on the node one.")
			framework.AddOrUpdateLabelOnNode(clientset, nodeOne.Name, "gpu", "zotac-nvidia")
			framework.ExpectNodeHasLabel(clientset, nodeOne.Name, "gpu", "zotac-nvidia")

			By("Trying to apply two label on the node two")
			framework.AddOrUpdateLabelOnNode(clientset, nodeTwo.Name, "gpu", "nvidia")
			framework.ExpectNodeHasLabel(clientset, nodeTwo.Name, "gpu", "nvidia")
			framework.AddOrUpdateLabelOnNode(clientset, nodeTwo.Name, "drive", "ssd")
			framework.ExpectNodeHasLabel(clientset, nodeTwo.Name, "drive", "ssd")

			By("Trying to launch the pod, now with requiredDuringSchedulingIgnoredDuringExecution and preferredDuringSchedulingIgnoredDuringExecution constraints")
			createTestPod(f, testpod)

			By("Validate if the pod is running in node-two matching both the hard and soft constraints")
			framework.ExpectNoError(framework.WaitForPodNotPending(clientset, ns, labelPodName))
			labelPod, err := clientset.CoreV1().Pods(ns).Get(labelPodName, metav1.GetOptions{})
			framework.ExpectNoError(err)
			Expect(labelPod.Spec.NodeName).To(Equal(nodeTwo.Name))

			By("Remove the soft(drive) label from node-two")
			framework.RemoveLabelOffNode(clientset, nodeTwo.Name, "drive")
			err = framework.VerifyLabelsRemoved(clientset, nodeTwo.Name, []string{"drive"})
			Expect(err).NotTo(HaveOccurred())

			By("Delete the pod")
			err = clientset.CoreV1().Pods(ns).Delete(labelPodName, &metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred())
			err = f.WaitForPodNotFound(labelPodName, 2*time.Minute)
			Expect(err).NotTo(HaveOccurred())

		})

		It("validates scheduler schedules a pod on any node matching the hard constraints", func() {

			By("re-create a pod, now with requiredDuringSchedulingIgnoredDuringExecution and preferredDuringSchedulingIgnoredDuringExecution constraint.")
			createTestPod(f, testpod)

			By("Validate if the pod is running in either node-one or node-two")
			framework.ExpectNoError(framework.WaitForPodNotPending(clientset, ns, labelPodName))
			labelPod, err := clientset.CoreV1().Pods(ns).Get(labelPodName, metav1.GetOptions{})
			framework.ExpectNoError(err)
			Expect(labelPod.Spec.NodeName).To(SatisfyAny(Equal(nodeOne.Name), Equal(nodeTwo.Name)))

			By("Delete the pod")
			err = clientset.CoreV1().Pods(ns).Delete(labelPodName, &metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred())
			err = f.WaitForPodNotFound(labelPodName, 2*time.Minute)
			Expect(err).NotTo(HaveOccurred())

			By("Remove the from node-two")
			framework.RemoveLabelOffNode(clientset, nodeTwo.Name, "gpu")
			err = framework.VerifyLabelsRemoved(clientset, nodeTwo.Name, []string{"gpu"})
			Expect(err).NotTo(HaveOccurred())
			framework.RemoveLabelOffNode(clientset, nodeOne.Name, "gpu")
			err = framework.VerifyLabelsRemoved(clientset, nodeOne.Name, []string{"gpu"})
			Expect(err).NotTo(HaveOccurred())

		})
	})

	Describe("Poseidon [Anti-NodeAffinity hard and soft constraint]", func() {
		var nodeOne, nodeTwo, nodeThree v1.Node
		labelPodName := "with-anti-affinity-hard-soft"
		testpod := testPodConfig{
			Name: labelPodName,
			Affinity: &v1.Affinity{
				NodeAffinity: &v1.NodeAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
						NodeSelectorTerms: []v1.NodeSelectorTerm{
							{
								MatchExpressions: []v1.NodeSelectorRequirement{
									{
										Key:      "gpu",
										Operator: v1.NodeSelectorOpNotIn,
										Values: []string{
											"nvidia",
											"zotac-nvidia",
										},
									},
								},
							},
						},
					},
					PreferredDuringSchedulingIgnoredDuringExecution: []v1.PreferredSchedulingTerm{
						{
							Weight: 10,
							Preference: v1.NodeSelectorTerm{
								MatchExpressions: []v1.NodeSelectorRequirement{
									{
										Key:      "drive",
										Operator: v1.NodeSelectorOpIn,
										Values: []string{
											"ssd",
										},
									},
								},
							},
						},
					},
				},
			},
			SchedulerName: "poseidon",
		}
		It("validates scheduler respect's a pod's anti-affinity constraint", func() {

			By("Trying to get a schedulable node")
			schedulableNodes := framework.ListSchedulableNodes(clientset)
			if len(schedulableNodes) < 3 {
				Skip(fmt.Sprintf("Skipping this test case as this requires minimum of three node and only %d nodes avaliable", len(schedulableNodes)))
			}
			nodeOne = schedulableNodes[0]
			nodeTwo = schedulableNodes[1]
			nodeThree = schedulableNodes[2]

			By(fmt.Sprintf("Trying to apply a label on %s", nodeOne.Name))
			framework.AddOrUpdateLabelOnNode(clientset, nodeOne.Name, "gpu", "zotac-nvidia")
			framework.ExpectNodeHasLabel(clientset, nodeOne.Name, "gpu", "zotac-nvidia")

			By(fmt.Sprintf("Trying to apply a label on %s.", nodeTwo.Name))
			framework.AddOrUpdateLabelOnNode(clientset, nodeTwo.Name, "gpu", "nvidia")
			framework.ExpectNodeHasLabel(clientset, nodeTwo.Name, "gpu", "nvidia")

			By(fmt.Sprintf("Trying to apply a label on %s.", nodeThree.Name))
			framework.AddOrUpdateLabelOnNode(clientset, nodeThree.Name, "drive", "ssd")
			framework.ExpectNodeHasLabel(clientset, nodeThree.Name, "drive", "ssd")

			By(fmt.Sprintf("Trying to launch the pod, with anti-affinity constraints on nodes %s and %s", nodeOne.Name, nodeTwo.Name))
			createTestPod(f, testpod)

			By(fmt.Sprintf("Validate if the pod is running on %s satisfying soft constraints as it has anti-affinity contraint's to other two nodes", nodeThree.Name))
			framework.ExpectNoError(framework.WaitForPodNotPending(clientset, ns, labelPodName))
			labelPod, err := clientset.CoreV1().Pods(ns).Get(labelPodName, metav1.GetOptions{})
			framework.ExpectNoError(err)
			Expect(labelPod.Spec.NodeName).To(Equal(nodeThree.Name))

			By(fmt.Sprintf("Remove the label from %s", nodeOne.Name))
			framework.RemoveLabelOffNode(clientset, nodeOne.Name, "gpu")
			err = framework.VerifyLabelsRemoved(clientset, nodeOne.Name, []string{"gpu"})
			Expect(err).NotTo(HaveOccurred())

			By(fmt.Sprintf("Remove the label from %s", nodeTwo.Name))
			framework.RemoveLabelOffNode(clientset, nodeTwo.Name, "gpu")
			err = framework.VerifyLabelsRemoved(clientset, nodeTwo.Name, []string{"gpu"})
			Expect(err).NotTo(HaveOccurred())

			By(fmt.Sprintf("Remove the label from %s", nodeThree.Name))
			framework.RemoveLabelOffNode(clientset, nodeTwo.Name, "drive")
			err = framework.VerifyLabelsRemoved(clientset, nodeTwo.Name, []string{"drive"})
			Expect(err).NotTo(HaveOccurred())

			By("Delete the pod")
			err = clientset.CoreV1().Pods(ns).Delete(labelPodName, &metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred())
			err = f.WaitForPodNotFound(labelPodName, 2*time.Minute)
			Expect(err).NotTo(HaveOccurred())

		})
	})

	Describe("Poseidon [Nodeselector with node Anti-Affinity hard and soft constraint]", func() {
		var nodeOne, nodeTwo v1.Node
		labelPodName := "node-affinity-nodesel-hard-and-soft"
		testpod := testPodConfig{
			Name: labelPodName,
			NodeSelector: map[string]string{
				"disktype": "ssd",
			},
			Affinity: &v1.Affinity{
				NodeAffinity: &v1.NodeAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
						NodeSelectorTerms: []v1.NodeSelectorTerm{
							{
								MatchExpressions: []v1.NodeSelectorRequirement{
									{
										Key:      "gpu",
										Operator: v1.NodeSelectorOpNotIn,
										Values: []string{
											"nvidia",
											"zotac-nvidia",
										},
									},
								},
							},
						},
					},
					PreferredDuringSchedulingIgnoredDuringExecution: []v1.PreferredSchedulingTerm{
						{
							Weight: 10,
							Preference: v1.NodeSelectorTerm{
								MatchExpressions: []v1.NodeSelectorRequirement{
									{
										Key:      "drive",
										Operator: v1.NodeSelectorOpIn,
										Values: []string{
											"ssd",
										},
									},
								},
							},
						},
					},
				},
			},
			SchedulerName: "poseidon",
		}
		It("validates scheduler respect's a pod's anti-affinity constraint", func() {

			By("Trying to get a schedulable node")
			schedulableNodes := framework.ListSchedulableNodes(clientset)
			if len(schedulableNodes) < 2 {
				Skip(fmt.Sprintf("Skipping this test case as this requires minimum of three node and only %d nodes avaliable", len(schedulableNodes)))
			}
			nodeOne = schedulableNodes[0]
			nodeTwo = schedulableNodes[1]

			By(fmt.Sprintf("Trying to apply a label on %s", nodeOne.Name))
			framework.AddOrUpdateLabelOnNode(clientset, nodeOne.Name, "gpu", "nvidia")
			framework.ExpectNodeHasLabel(clientset, nodeOne.Name, "gpu", "nvidia")

			By(fmt.Sprintf("Trying to apply two label on %s.", nodeTwo.Name))
			framework.AddOrUpdateLabelOnNode(clientset, nodeTwo.Name, "gpu", "amd")
			framework.ExpectNodeHasLabel(clientset, nodeTwo.Name, "gpu", "amd")
			framework.AddOrUpdateLabelOnNode(clientset, nodeTwo.Name, "disktype", "ssd")
			framework.ExpectNodeHasLabel(clientset, nodeTwo.Name, "disktype", "ssd")

			By(fmt.Sprintf("Trying to launch the pod, with anti-affinity constraints on nodes %s", nodeOne.Name))
			createTestPod(f, testpod)

			By(fmt.Sprintf("Validate if the pod is running on %s satisfying soft constraints as it has anti-affinity contraint's to other two nodes", nodeTwo.Name))
			framework.ExpectNoError(framework.WaitForPodNotPending(clientset, ns, labelPodName))
			labelPod, err := clientset.CoreV1().Pods(ns).Get(labelPodName, metav1.GetOptions{})
			framework.ExpectNoError(err)
			Expect(labelPod.Spec.NodeName).To(Equal(nodeTwo.Name))

			By(fmt.Sprintf("Remove the label from %s", nodeOne.Name))
			framework.RemoveLabelOffNode(clientset, nodeOne.Name, "gpu")
			err = framework.VerifyLabelsRemoved(clientset, nodeOne.Name, []string{"gpu"})
			Expect(err).NotTo(HaveOccurred())

			By(fmt.Sprintf("Remove the label from %s", nodeTwo.Name))
			framework.RemoveLabelOffNode(clientset, nodeTwo.Name, "gpu")
			err = framework.VerifyLabelsRemoved(clientset, nodeTwo.Name, []string{"gpu"})
			Expect(err).NotTo(HaveOccurred())

			framework.RemoveLabelOffNode(clientset, nodeTwo.Name, "disktype")
			err = framework.VerifyLabelsRemoved(clientset, nodeTwo.Name, []string{"disktype"})
			Expect(err).NotTo(HaveOccurred())

			By("Delete the pod")
			err = clientset.CoreV1().Pods(ns).Delete(labelPodName, &metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred())
			err = f.WaitForPodNotFound(labelPodName, 2*time.Minute)
			Expect(err).NotTo(HaveOccurred())

		})
	})

	Describe("Poseidon [Pod Affinity and Anti-Affinity]", func() {
		var setupPodOnesNodeName, setupPodTwosNodeName string

		setupPodOneName := "setup-pod-one"
		setupPodTwoName := "setup-pod-two"
		setupPodOne := testPodConfig{
			Name: setupPodOneName,
			Labels: map[string]string{
				"security": "S1",
			},
			SchedulerName: "poseidon",
		}

		setupPodTwo := testPodConfig{
			Name: setupPodTwoName,
			Labels: map[string]string{
				"security": "S2",
			},
			SchedulerName: "poseidon",
		}

		It("validates scheduler respect's a pod-affinity with hard constraint", func() {
			labelPodName := "pod-affinity-hard"
			testpod := testPodConfig{
				Name: labelPodName,
				Affinity: &v1.Affinity{
					PodAffinity: &v1.PodAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: []v1.PodAffinityTerm{
							{
								LabelSelector: &metav1.LabelSelector{
									MatchExpressions: []metav1.LabelSelectorRequirement{
										{
											Key:      "security",
											Operator: metav1.LabelSelectorOpIn,
											Values: []string{
												"S1",
											},
										},
									},
								},
								TopologyKey: "kubernetes.io/hostname",
							},
						},
					},
				},
				SchedulerName: "poseidon",
			}

			By("Trying to launch the setup pod one with label security S1 to test pod affinity/anti-affinity")
			createTestPod(f, setupPodOne)

			By("validate if setup pod one is running and get the node of the setup pod")
			framework.ExpectNoError(framework.WaitForPodNotPending(clientset, ns, setupPodOneName))
			setupLabelPod, err := clientset.CoreV1().Pods(ns).Get(setupPodOneName, metav1.GetOptions{})
			framework.ExpectNoError(err)
			setupPodOnesNodeName = setupLabelPod.Spec.NodeName

			By("Deploy the pod with pod affinity hard constraint")
			createTestPod(f, testpod)

			By("validate if test pod is running on the right node")
			framework.ExpectNoError(framework.WaitForPodNotPending(clientset, ns, labelPodName))
			labelPod, err := clientset.CoreV1().Pods(ns).Get(labelPodName, metav1.GetOptions{})
			framework.ExpectNoError(err)
			Expect(labelPod.Spec.NodeName).To(Equal(setupPodOnesNodeName))

			By("Delete the test pod")
			err = clientset.CoreV1().Pods(ns).Delete(labelPod.Name, &metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred())
			err = f.WaitForPodNotFound(labelPodName, 2*time.Minute)
			Expect(err).NotTo(HaveOccurred())
		})

		It("validates scheduler respect's a pod-anti-affinity with hard constraint", func() {
			labelPodName := "pod-anti-affinity-hard"
			testpod := testPodConfig{
				Name: labelPodName,
				Affinity: &v1.Affinity{
					PodAntiAffinity: &v1.PodAntiAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: []v1.PodAffinityTerm{
							{
								LabelSelector: &metav1.LabelSelector{
									MatchExpressions: []metav1.LabelSelectorRequirement{
										{
											Key:      "security",
											Operator: metav1.LabelSelectorOpIn,
											Values: []string{
												"S1",
											},
										},
									},
								},
								TopologyKey: "kubernetes.io/hostname",
							},
						},
					},
				},
				SchedulerName: "poseidon",
			}

			// Note: We don't need to create the setup pod one for this test case
			// it already running and we use the same

			By("Deploy the pod with pod anti-affinity hard constraint")
			createTestPod(f, testpod)

			By("validate if test pod is running on the right node")
			framework.ExpectNoError(framework.WaitForPodNotPending(clientset, ns, labelPodName))
			labelPod, err := clientset.CoreV1().Pods(ns).Get(labelPodName, metav1.GetOptions{})
			framework.ExpectNoError(err)
			Expect(labelPod.Spec.NodeName).NotTo(Equal(setupPodOnesNodeName))

			By("Delete the test pod")
			err = clientset.CoreV1().Pods(ns).Delete(labelPod.Name, &metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred())
			err = f.WaitForPodNotFound(labelPodName, 2*time.Minute)
			Expect(err).NotTo(HaveOccurred())
		})

		It("validates scheduler respect's a pod-affinity with soft constraint", func() {
			labelPodName := "pod-affinity-soft"
			testpod := testPodConfig{
				Name: labelPodName,
				Affinity: &v1.Affinity{
					PodAffinity: &v1.PodAffinity{
						PreferredDuringSchedulingIgnoredDuringExecution: []v1.WeightedPodAffinityTerm{
							{
								Weight: 100,
								PodAffinityTerm: v1.PodAffinityTerm{
									LabelSelector: &metav1.LabelSelector{
										MatchExpressions: []metav1.LabelSelectorRequirement{
											{
												Key:      "security",
												Operator: metav1.LabelSelectorOpIn,
												Values: []string{
													"S1",
												},
											},
										},
									},
									TopologyKey: "kubernetes.io/hostname",
								},
							},
						},
					},
				},
				SchedulerName: "poseidon",
			}

			// Note: We don't need to create the setup pod for this test case
			// it already running and we use the same

			By("Deploy the pod with pod-affinity soft constraint")
			createTestPod(f, testpod)

			By("validate if test pod is running on the right node")
			framework.ExpectNoError(framework.WaitForPodNotPending(clientset, ns, labelPodName))
			labelPod, err := clientset.CoreV1().Pods(ns).Get(labelPodName, metav1.GetOptions{})
			framework.ExpectNoError(err)
			Expect(labelPod.Spec.NodeName).To(Equal(setupPodOnesNodeName))

			By("Delete the test pod")
			err = clientset.CoreV1().Pods(ns).Delete(labelPod.Name, &metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred())
			err = f.WaitForPodNotFound(labelPodName, 2*time.Minute)
			Expect(err).NotTo(HaveOccurred())
		})

		It("validates scheduler respect's a pod-anti-affinity with soft constraint", func() {
			labelPodName := "pod-anti-affinity-soft"
			testpod := testPodConfig{
				Name: labelPodName,
				Affinity: &v1.Affinity{
					PodAntiAffinity: &v1.PodAntiAffinity{
						PreferredDuringSchedulingIgnoredDuringExecution: []v1.WeightedPodAffinityTerm{
							{
								Weight: 100,
								PodAffinityTerm: v1.PodAffinityTerm{
									LabelSelector: &metav1.LabelSelector{
										MatchExpressions: []metav1.LabelSelectorRequirement{
											{
												Key:      "security",
												Operator: metav1.LabelSelectorOpIn,
												Values: []string{
													"S1",
												},
											},
										},
									},
									TopologyKey: "kubernetes.io/hostname",
								},
							},
						},
					},
				},
				SchedulerName: "poseidon",
			}

			// Note: We don't need to create the setup pod one for this test case
			// it already running and we use the same

			By("Deploy the pod with pod-anti-affinity soft constraint")
			createTestPod(f, testpod)

			By("validate if test pod is running on the right node")
			framework.ExpectNoError(framework.WaitForPodNotPending(clientset, ns, labelPodName))
			labelPod, err := clientset.CoreV1().Pods(ns).Get(labelPodName, metav1.GetOptions{})
			framework.ExpectNoError(err)
			Expect(labelPod.Spec.NodeName).NotTo(Equal(setupPodOnesNodeName))

			By("Delete the test pod")
			err = clientset.CoreV1().Pods(ns).Delete(labelPod.Name, &metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred())
			err = f.WaitForPodNotFound(labelPodName, 2*time.Minute)
			Expect(err).NotTo(HaveOccurred())
		})

		It("validates scheduler respect's a pod-affinity with both hard and soft constraint", func() {
			labelPodName := "pod-affinity-hard-soft"
			testpod := testPodConfig{
				Name: labelPodName,
				Affinity: &v1.Affinity{
					PodAffinity: &v1.PodAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: []v1.PodAffinityTerm{
							{
								LabelSelector: &metav1.LabelSelector{
									MatchExpressions: []metav1.LabelSelectorRequirement{
										{
											Key:      "security",
											Operator: metav1.LabelSelectorOpIn,
											Values: []string{
												"S1",
											},
										},
									},
								},
								TopologyKey: "kubernetes.io/hostname",
							},
						},
						PreferredDuringSchedulingIgnoredDuringExecution: []v1.WeightedPodAffinityTerm{
							{
								Weight: 100,
								PodAffinityTerm: v1.PodAffinityTerm{
									LabelSelector: &metav1.LabelSelector{
										MatchExpressions: []metav1.LabelSelectorRequirement{
											{
												Key:      "security",
												Operator: metav1.LabelSelectorOpIn,
												Values: []string{
													"S2",
												},
											},
										},
									},
									TopologyKey: "kubernetes.io/hostname",
								},
							},
						},
					},
				},
				SchedulerName: "poseidon",
			}

			By("Trying to launch the setup pod two with label security S2 to test pod affinity/anti-affinity")
			createTestPod(f, setupPodTwo)

			By("validate if setup pod two is running and get the node of the setup pod two")
			framework.ExpectNoError(framework.WaitForPodNotPending(clientset, ns, setupPodTwoName))
			setupLabelPod, err := clientset.CoreV1().Pods(ns).Get(setupPodTwoName, metav1.GetOptions{})
			framework.ExpectNoError(err)
			setupPodTwosNodeName = setupLabelPod.Spec.NodeName

			// Note: We don't need to create the setup pod one for this test case
			// it already running and we use the same
			By("Deploy the pod with pod-affinity hard and soft constraint")
			createTestPod(f, testpod)

			By("validate if test pod is running on the right node")
			framework.ExpectNoError(framework.WaitForPodNotPending(clientset, ns, labelPodName))
			labelPod, err := clientset.CoreV1().Pods(ns).Get(labelPodName, metav1.GetOptions{})
			framework.ExpectNoError(err)
			Expect(labelPod.Spec.NodeName).To(Equal(setupPodOnesNodeName))

			By("Delete the test pod")
			err = clientset.CoreV1().Pods(ns).Delete(labelPod.Name, &metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred())
			err = f.WaitForPodNotFound(labelPodName, 2*time.Minute)
			Expect(err).NotTo(HaveOccurred())
		})

		It("validates scheduler respect's a pod-anti-affinity with both hard and soft constraint", func() {
			labelPodName := "pod-anti-affinity-hard-soft"
			testpod := testPodConfig{
				Name: labelPodName,
				Affinity: &v1.Affinity{
					PodAntiAffinity: &v1.PodAntiAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: []v1.PodAffinityTerm{
							{
								LabelSelector: &metav1.LabelSelector{
									MatchExpressions: []metav1.LabelSelectorRequirement{
										{
											Key:      "security",
											Operator: metav1.LabelSelectorOpIn,
											Values: []string{
												"S1",
											},
										},
									},
								},
								TopologyKey: "kubernetes.io/hostname",
							},
						},
						PreferredDuringSchedulingIgnoredDuringExecution: []v1.WeightedPodAffinityTerm{
							{
								Weight: 100,
								PodAffinityTerm: v1.PodAffinityTerm{
									LabelSelector: &metav1.LabelSelector{
										MatchExpressions: []metav1.LabelSelectorRequirement{
											{
												Key:      "security",
												Operator: metav1.LabelSelectorOpIn,
												Values: []string{
													"S2",
												},
											},
										},
									},
									TopologyKey: "kubernetes.io/hostname",
								},
							},
						},
					},
				},
				SchedulerName: "poseidon",
			}

			// Note: We don't need to create the setup pod one and setup pod two for this test case
			// it already running and we use the same

			By("Deploy the pod with pod-anti-affinity hard and soft constraint")
			createTestPod(f, testpod)

			By("validate if test pod is running on the right node")
			framework.ExpectNoError(framework.WaitForPodNotPending(clientset, ns, labelPodName))
			labelPod, err := clientset.CoreV1().Pods(ns).Get(labelPodName, metav1.GetOptions{})
			framework.ExpectNoError(err)
			Expect(labelPod.Spec.NodeName).ShouldNot(SatisfyAll(Equal(setupPodOnesNodeName), Equal(setupPodTwosNodeName)))

			By("Delete the test pod")
			err = clientset.CoreV1().Pods(ns).Delete(labelPod.Name, &metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred())
			err = f.WaitForPodNotFound(labelPodName, 2*time.Minute)
			Expect(err).NotTo(HaveOccurred())
		})

		It("validates scheduler respect's a pod-affinity with hard constraint and anti-affinity soft constraint", func() {
			labelPodName := "pod-affinity-hard-anti-affinity-soft"
			testpod := testPodConfig{
				Name: labelPodName,
				Affinity: &v1.Affinity{
					PodAffinity: &v1.PodAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: []v1.PodAffinityTerm{
							{
								LabelSelector: &metav1.LabelSelector{
									MatchExpressions: []metav1.LabelSelectorRequirement{
										{
											Key:      "security",
											Operator: metav1.LabelSelectorOpIn,
											Values: []string{
												"S1",
											},
										},
									},
								},
								TopologyKey: "kubernetes.io/hostname",
							},
						},
					},
					PodAntiAffinity: &v1.PodAntiAffinity{
						PreferredDuringSchedulingIgnoredDuringExecution: []v1.WeightedPodAffinityTerm{
							{
								Weight: 100,
								PodAffinityTerm: v1.PodAffinityTerm{
									LabelSelector: &metav1.LabelSelector{
										MatchExpressions: []metav1.LabelSelectorRequirement{
											{
												Key:      "security",
												Operator: metav1.LabelSelectorOpIn,
												Values: []string{
													"S2",
												},
											},
										},
									},
									TopologyKey: "kubernetes.io/hostname",
								},
							},
						},
					},
				},
				SchedulerName: "poseidon",
			}

			// Note: We don't need to create the setup pod one and setup pod two for this test case
			// it already running and we use the same

			By("Deploy the pod with pod-affinity hard and anti-affinity soft constraint")
			createTestPod(f, testpod)

			By("validate if test pod is running on the right node")
			framework.ExpectNoError(framework.WaitForPodNotPending(clientset, ns, labelPodName))
			labelPod, err := clientset.CoreV1().Pods(ns).Get(labelPodName, metav1.GetOptions{})
			framework.ExpectNoError(err)
			Expect(labelPod.Spec.NodeName).To(Equal(setupPodOnesNodeName))

			By("Delete the test pod")
			err = clientset.CoreV1().Pods(ns).Delete(labelPod.Name, &metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred())
			err = f.WaitForPodNotFound(labelPodName, 2*time.Minute)
			Expect(err).NotTo(HaveOccurred())
		})

		It("validates scheduler respect's a pod-anti-affinity with hard constraint and affinity soft constraint", func() {
			labelPodName := "pod-anti-affinity-hard-affinity-soft"
			testpod := testPodConfig{
				Name: labelPodName,
				Affinity: &v1.Affinity{
					PodAntiAffinity: &v1.PodAntiAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: []v1.PodAffinityTerm{
							{
								LabelSelector: &metav1.LabelSelector{
									MatchExpressions: []metav1.LabelSelectorRequirement{
										{
											Key:      "security",
											Operator: metav1.LabelSelectorOpIn,
											Values: []string{
												"S1",
											},
										},
									},
								},
								TopologyKey: "kubernetes.io/hostname",
							},
						},
					},
					PodAffinity: &v1.PodAffinity{
						PreferredDuringSchedulingIgnoredDuringExecution: []v1.WeightedPodAffinityTerm{
							{
								Weight: 100,
								PodAffinityTerm: v1.PodAffinityTerm{
									LabelSelector: &metav1.LabelSelector{
										MatchExpressions: []metav1.LabelSelectorRequirement{
											{
												Key:      "security",
												Operator: metav1.LabelSelectorOpIn,
												Values: []string{
													"S2",
												},
											},
										},
									},
									TopologyKey: "kubernetes.io/hostname",
								},
							},
						},
					},
				},
				SchedulerName: "poseidon",
			}

			// Note: We don't need to create the setup pod one and setup pod two for this test case
			// it already running and we use the same

			By("Deploy the pod with pod-affinity hard and anti-affinity soft constraint")
			createTestPod(f, testpod)

			By("validate if test pod is running on the right node")
			framework.ExpectNoError(framework.WaitForPodNotPending(clientset, ns, labelPodName))
			labelPod, err := clientset.CoreV1().Pods(ns).Get(labelPodName, metav1.GetOptions{})
			framework.ExpectNoError(err)

			if setupPodTwosNodeName == setupPodOnesNodeName {
				Expect(labelPod.Spec.NodeName).NotTo(Equal(setupPodTwosNodeName))
			} else {
				Expect(labelPod.Spec.NodeName).To(Equal(setupPodTwosNodeName))
			}

			By("Delete the test pod")
			err = clientset.CoreV1().Pods(ns).Delete(labelPod.Name, &metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred())
			err = f.WaitForPodNotFound(labelPodName, 2*time.Minute)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("Poseidon [Pod Affinity and Anti-Affinity symmetry case]", func() {

		It("Hard constraint symmetry behavior should consider pod anti-affinity requirement of already running pod", func() {
			setupPodName := "setup-podantisymmetry"
			setupPod := testPodConfig{
				Name: setupPodName,
				Labels: map[string]string{
					"nkey": "nvalue",
				},
				Affinity: &v1.Affinity{
					PodAntiAffinity: &v1.PodAntiAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: []v1.PodAffinityTerm{
							{
								LabelSelector: &metav1.LabelSelector{
									MatchExpressions: []metav1.LabelSelectorRequirement{
										{
											Key:      "security",
											Operator: metav1.LabelSelectorOpIn,
											Values: []string{
												"S1",
											},
										},
									},
								},
								TopologyKey: "kubernetes.io/hostname",
							},
						},
					},
				},
				SchedulerName: "poseidon",
			}

			By("Trying to launch the setup pod one with anti-affinity requirement with security,S1 as key and value respectively")
			createTestPod(f, setupPod)

			By("validate if setup pod one is running and get the node of the setup pod")
			framework.ExpectNoError(framework.WaitForPodNotPending(clientset, ns, setupPodName))
			setupLabelPod, err := clientset.CoreV1().Pods(ns).Get(setupPodName, metav1.GetOptions{})
			framework.ExpectNoError(err)
			setupPodOnesNodeName := setupLabelPod.Spec.NodeName

			antiSymPodName := "podantisymmetry"
			antiSymPodPod := testPodConfig{
				Name: antiSymPodName,
				Labels: map[string]string{
					"security": "S1",
				},
				SchedulerName: "poseidon",
			}

			By("Deploy the pod which conflicts with the pod anti-affinity hard requirement of setup pod")
			createTestPod(f, antiSymPodPod)

			By("validate if pod is running on the right node")
			framework.ExpectNoError(framework.WaitForPodNotPending(clientset, ns, antiSymPodName))
			labelPod, err := clientset.CoreV1().Pods(ns).Get(antiSymPodName, metav1.GetOptions{})
			framework.ExpectNoError(err)
			Expect(labelPod.Spec.NodeName).NotTo(Equal(setupPodOnesNodeName))

			By("Delete the anti symmetry pod")
			err = clientset.CoreV1().Pods(ns).Delete(labelPod.Name, &metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred())
			err = f.WaitForPodNotFound(labelPod.Name, 2*time.Minute)
			Expect(err).NotTo(HaveOccurred())

			By("Delete the setup test pod")
			err = clientset.CoreV1().Pods(ns).Delete(setupPodName, &metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred())
			err = f.WaitForPodNotFound(setupPodName, 2*time.Minute)
			Expect(err).NotTo(HaveOccurred())

		})

		It("Soft constraint symmetry behavior should consider pod affinity hard constraint of already running pod", func() {
			setupPodName := "setup-podaffinity-one"
			setupPod := testPodConfig{
				Name: setupPodName,
				Labels: map[string]string{
					"nkey": "nvalue",
				},
				Affinity: &v1.Affinity{
					PodAffinity: &v1.PodAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: []v1.PodAffinityTerm{
							{
								LabelSelector: &metav1.LabelSelector{
									MatchExpressions: []metav1.LabelSelectorRequirement{
										{
											Key:      "security",
											Operator: metav1.LabelSelectorOpIn,
											Values: []string{
												"S1",
											},
										},
									},
								},
								TopologyKey: "kubernetes.io/hostname",
							},
						},
					},
				},
				SchedulerName: "poseidon",
			}

			By("Trying to launch the setup pod one with affinity requirement with security,S1 as key and value respectively")
			createTestPod(f, setupPod)

			By("validate if setup pod one is running and get the node of the setup pod")
			framework.ExpectNoError(framework.WaitForPodNotPending(clientset, ns, setupPodName))
			setupLabelPod, err := clientset.CoreV1().Pods(ns).Get(setupPodName, metav1.GetOptions{})
			framework.ExpectNoError(err)
			setupPodOnesNodeName := setupLabelPod.Spec.NodeName

			symPodName := "podsymmetry"
			symPodPod := testPodConfig{
				Name: symPodName,
				Labels: map[string]string{
					"security": "S1",
				},
				SchedulerName: "poseidon",
			}

			By("Deploy a pod with the same label as the affinity pod with hard requirement")
			createTestPod(f, symPodPod)

			By("validate if pod is running on the right node")
			framework.ExpectNoError(framework.WaitForPodNotPending(clientset, ns, symPodName))
			labelPod, err := clientset.CoreV1().Pods(ns).Get(symPodName, metav1.GetOptions{})
			framework.ExpectNoError(err)
			Expect(labelPod.Spec.NodeName).To(Equal(setupPodOnesNodeName))

			By("Delete the symmetry pod")
			err = clientset.CoreV1().Pods(ns).Delete(labelPod.Name, &metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred())
			err = f.WaitForPodNotFound(labelPod.Name, 2*time.Minute)
			Expect(err).NotTo(HaveOccurred())

			By("Delete the setup test pod")
			err = clientset.CoreV1().Pods(ns).Delete(setupPodName, &metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred())
			err = f.WaitForPodNotFound(setupPodName, 2*time.Minute)
			Expect(err).NotTo(HaveOccurred())

		})

		It("Soft constraint symmetry behavior should consider pod affinity soft constraint of already running pod", func() {
			setupPodName := "setup-podaffinity-two"
			setupPod := testPodConfig{
				Name: setupPodName,
				Labels: map[string]string{
					"nkey": "nvalue",
				},
				Affinity: &v1.Affinity{
					PodAffinity: &v1.PodAffinity{
						PreferredDuringSchedulingIgnoredDuringExecution: []v1.WeightedPodAffinityTerm{
							{
								Weight: 100,
								PodAffinityTerm: v1.PodAffinityTerm{
									LabelSelector: &metav1.LabelSelector{
										MatchExpressions: []metav1.LabelSelectorRequirement{
											{
												Key:      "security",
												Operator: metav1.LabelSelectorOpIn,
												Values: []string{
													"S1",
												},
											},
										},
									},
									TopologyKey: "kubernetes.io/hostname",
								},
							},
						},
					},
				},
				SchedulerName: "poseidon",
			}

			By("Trying to launch the setup pod one with affinity requirement with security,S1 as key and value respectively")
			createTestPod(f, setupPod)

			By("validate if setup pod one is running and get the node of the setup pod")
			framework.ExpectNoError(framework.WaitForPodNotPending(clientset, ns, setupPodName))
			setupLabelPod, err := clientset.CoreV1().Pods(ns).Get(setupPodName, metav1.GetOptions{})
			framework.ExpectNoError(err)
			setupPodOnesNodeName := setupLabelPod.Spec.NodeName

			symPodName := "podsymmetry"
			symPodPod := testPodConfig{
				Name: symPodName,
				Labels: map[string]string{
					"security": "S1",
				},
				SchedulerName: "poseidon",
			}

			By("Deploy a pod with the same label as the affinity pod with soft requirement")
			createTestPod(f, symPodPod)

			By("validate if pod is running on the right node")
			framework.ExpectNoError(framework.WaitForPodNotPending(clientset, ns, symPodName))
			labelPod, err := clientset.CoreV1().Pods(ns).Get(symPodName, metav1.GetOptions{})
			framework.ExpectNoError(err)
			Expect(labelPod.Spec.NodeName).To(Equal(setupPodOnesNodeName))

			By("Delete the symmetry pod")
			err = clientset.CoreV1().Pods(ns).Delete(labelPod.Name, &metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred())
			err = f.WaitForPodNotFound(labelPod.Name, 2*time.Minute)
			Expect(err).NotTo(HaveOccurred())

			By("Delete the setup test pod")
			err = clientset.CoreV1().Pods(ns).Delete(setupPodName, &metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred())
			err = f.WaitForPodNotFound(setupPodName, 2*time.Minute)
			Expect(err).NotTo(HaveOccurred())

		})

		It("Soft constraint symmetry behavior should consider pod anti-affinity soft constraint of already running pod", func() {
			setupPodName := "setup-pod-antiaffinity"
			setupPod := testPodConfig{
				Name: setupPodName,
				Labels: map[string]string{
					"nkey": "nvalue",
				},
				Affinity: &v1.Affinity{
					PodAntiAffinity: &v1.PodAntiAffinity{
						PreferredDuringSchedulingIgnoredDuringExecution: []v1.WeightedPodAffinityTerm{
							{
								Weight: 100,
								PodAffinityTerm: v1.PodAffinityTerm{
									LabelSelector: &metav1.LabelSelector{
										MatchExpressions: []metav1.LabelSelectorRequirement{
											{
												Key:      "security",
												Operator: metav1.LabelSelectorOpIn,
												Values: []string{
													"S1",
												},
											},
										},
									},
									TopologyKey: "kubernetes.io/hostname",
								},
							},
						},
					},
				},
				SchedulerName: "poseidon",
			}

			By("Trying to launch the setup pod one with affinity requirement with security,S1 as key and value respectively")
			createTestPod(f, setupPod)

			By("validate if setup pod one is running and get the node of the setup pod")
			framework.ExpectNoError(framework.WaitForPodNotPending(clientset, ns, setupPodName))
			setupLabelPod, err := clientset.CoreV1().Pods(ns).Get(setupPodName, metav1.GetOptions{})
			framework.ExpectNoError(err)
			setupPodOnesNodeName := setupLabelPod.Spec.NodeName

			symPodName := "podantisymmetry"
			symPodPod := testPodConfig{
				Name: symPodName,
				Labels: map[string]string{
					"security": "S1",
				},
				SchedulerName: "poseidon",
			}

			By("Deploy a pod with the same label as the anti affinity pod with soft requirement")
			createTestPod(f, symPodPod)

			By("validate if pod is running on the right node")
			framework.ExpectNoError(framework.WaitForPodNotPending(clientset, ns, symPodName))
			labelPod, err := clientset.CoreV1().Pods(ns).Get(symPodName, metav1.GetOptions{})
			framework.ExpectNoError(err)
			Expect(labelPod.Spec.NodeName).NotTo(Equal(setupPodOnesNodeName))

			By("Delete the symmetry pod")
			err = clientset.CoreV1().Pods(ns).Delete(labelPod.Name, &metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred())
			err = f.WaitForPodNotFound(labelPod.Name, 2*time.Minute)
			Expect(err).NotTo(HaveOccurred())

			By("Delete the setup test pod")
			err = clientset.CoreV1().Pods(ns).Delete(setupPodName, &metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred())
			err = f.WaitForPodNotFound(setupPodName, 2*time.Minute)
			Expect(err).NotTo(HaveOccurred())

		})
	})

	Describe("Poseidon [Taints and Tolerations hard and soft constraint]", func() {
		var nodeOne, nodeTwo v1.Node

		It("validates scheduler respect's a taints and tolerations constraint ie a pod having no tolerations can't be scheduled onto a node with nonempty taints.", func() {
			labelPodName := "nginx-with-no-tolerations"
			testpod := testPodConfig{
				Name:          labelPodName,
				SchedulerName: "poseidon",
			}

			By("Trying to get a schedulable node")
			schedulableNodes := framework.ListSchedulableNodes(clientset)
			if len(schedulableNodes) < 2 {
				Skip(fmt.Sprintf("Skipping this test case as this requires minimum of two node and only %d nodes avaliable", len(schedulableNodes)))
			}
			nodeOne = schedulableNodes[0]
			nodeTwo = schedulableNodes[1]

			By(fmt.Sprintf("Trying to apply a taint on %s", nodeOne.Name))
			taint := v1.Taint{
				Key:    "dedicated",
				Value:  "user1",
				Effect: "NoSchedule",
			}

			defer func() {
				By(fmt.Sprintf("Remove the taint from %s", nodeOne.Name))
				framework.RemoveTaintOffNode(clientset, nodeOne.Name, taint)
				framework.VerifyThatTaintIsGone(clientset, nodeOne.Name, &taint)
				By("Delete the pod")
				err := clientset.CoreV1().Pods(ns).Delete(labelPodName, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred())
				err = f.WaitForPodNotFound(labelPodName, 2*time.Minute)
				Expect(err).NotTo(HaveOccurred())

			}()

			framework.AddOrUpdateTaintOnNode(clientset, nodeOne.Name, taint)
			framework.ExpectNodeHasTaint(clientset, nodeOne.Name, &taint)

			By(fmt.Sprintf("Trying to launch the pod, with taints on nodes %s as it has intolerable taints on other nodes", nodeTwo.Name))
			createTestPod(f, testpod)

			By(fmt.Sprintf("Validate if the pod is running on %s ", nodeTwo.Name))
			framework.ExpectNoError(framework.WaitForPodNotPending(clientset, ns, labelPodName))
			labelPod, err := clientset.CoreV1().Pods(ns).Get(labelPodName, metav1.GetOptions{})
			framework.ExpectNoError(err)
			Expect(labelPod.Spec.NodeName).NotTo(Equal(nodeOne.Name))

		})

		It("a pod has a toleration that keys and values match the taint on the node, but (non-empty) effect doesn't match, can't be scheduled onto the node", func() {
			labelPodName := "nginx-with-no-match-effect"
			testpod := testPodConfig{
				Name: labelPodName,
				Tolerations: []v1.Toleration{
					{
						Key:      "foo",
						Operator: "Equal",
						Value:    "bar",
						Effect:   "PreferNoSchedule",
					},
				},
				SchedulerName: "poseidon",
			}

			By("Trying to get a schedulable node")
			schedulableNodes := framework.ListSchedulableNodes(clientset)
			if len(schedulableNodes) < 2 {
				Skip(fmt.Sprintf("Skipping this test case as this requires minimum of two node and only %d nodes avaliable", len(schedulableNodes)))
			}
			nodeOne = schedulableNodes[0]
			nodeTwo = schedulableNodes[1]

			By(fmt.Sprintf("Trying to apply a taint on %s", nodeOne.Name))
			taint := v1.Taint{
				Key:    "foo",
				Value:  "bar",
				Effect: "NoSchedule",
			}
			defer func() {
				By(fmt.Sprintf("Remove the taint from %s", nodeOne.Name))
				framework.RemoveTaintOffNode(clientset, nodeOne.Name, taint)
				framework.VerifyThatTaintIsGone(clientset, nodeOne.Name, &taint)
				By("Delete the pod")
				err := clientset.CoreV1().Pods(ns).Delete(labelPodName, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred())
				err = f.WaitForPodNotFound(labelPodName, 2*time.Minute)
				Expect(err).NotTo(HaveOccurred())

			}()
			framework.AddOrUpdateTaintOnNode(clientset, nodeOne.Name, taint)
			framework.ExpectNodeHasTaint(clientset, nodeOne.Name, &taint)

			By(fmt.Sprintf("Trying to launch the pod, with taints on nodes %s as it has intolerable taints on other nodes", nodeTwo.Name))
			createTestPod(f, testpod)

			By(fmt.Sprintf("Validate if the pod is running on %s ", nodeTwo.Name))
			framework.ExpectNoError(framework.WaitForPodNotPending(clientset, ns, labelPodName))
			labelPod, err := clientset.CoreV1().Pods(ns).Get(labelPodName, metav1.GetOptions{})
			framework.ExpectNoError(err)
			Expect(labelPod.Spec.NodeName).NotTo(Equal(nodeOne.Name))
		})

		It("a node with taints tolerated by the pod, gets a higher score or is more preferred than those node with intolerable taints", func() {
			labelPodName := "nginx-with-prefernoschedule"
			testpod := testPodConfig{
				Name: labelPodName,
				Tolerations: []v1.Toleration{
					{
						Key:      "foo",
						Operator: "Equal",
						Value:    "bar",
						Effect:   "PreferNoSchedule",
					},
				},

				SchedulerName: "poseidon",
			}

			By("Trying to get a schedulable node")
			schedulableNodes := framework.ListSchedulableNodes(clientset)
			if len(schedulableNodes) < 2 {
				Skip(fmt.Sprintf("Skipping this test case as this requires minimum of two node and only %d nodes avaliable", len(schedulableNodes)))
			}
			nodeOne = schedulableNodes[0]
			nodeTwo = schedulableNodes[1]

			taint := v1.Taint{
				Key:    "foo",
				Value:  "bar",
				Effect: "PreferNoSchedule",
			}

			taint1 := v1.Taint{
				Key:    "foo",
				Value:  "blah",
				Effect: "PreferNoSchedule",
			}

			defer func() {
				By(fmt.Sprintf("Remove the taint from %s", nodeOne.Name))
				framework.RemoveTaintOffNode(clientset, nodeOne.Name, taint)
				framework.VerifyThatTaintIsGone(clientset, nodeOne.Name, &taint)
				By(fmt.Sprintf("Remove the taint from %s", nodeTwo.Name))
				framework.RemoveTaintOffNode(clientset, nodeTwo.Name, taint1)
				framework.VerifyThatTaintIsGone(clientset, nodeTwo.Name, &taint1)
				By("Delete the pod")
				err := clientset.CoreV1().Pods(ns).Delete(labelPodName, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred())
				err = f.WaitForPodNotFound(labelPodName, 2*time.Minute)
				Expect(err).NotTo(HaveOccurred())

			}()

			By(fmt.Sprintf("Trying to apply a taint on %s", nodeOne.Name))
			framework.AddOrUpdateTaintOnNode(clientset, nodeOne.Name, taint)
			framework.ExpectNodeHasTaint(clientset, nodeOne.Name, &taint)

			By(fmt.Sprintf("Trying to apply a taint on %s", nodeTwo.Name))
			framework.AddOrUpdateTaintOnNode(clientset, nodeTwo.Name, taint1)
			framework.ExpectNodeHasTaint(clientset, nodeTwo.Name, &taint1)

			By(fmt.Sprintf("Trying to launch the pod, with taints on nodes %s %s ", nodeOne.Name, nodeTwo.Name))
			createTestPod(f, testpod)

			By(fmt.Sprintf("Validate if the pod is running on %s ", nodeOne.Name))
			framework.ExpectNoError(framework.WaitForPodNotPending(clientset, ns, labelPodName))
			labelPod, err := clientset.CoreV1().Pods(ns).Get(labelPodName, metav1.GetOptions{})
			framework.ExpectNoError(err)
			Expect(labelPod.Spec.NodeName).NotTo(Equal(nodeTwo.Name))

		})

		It("a pod without tolerations land on a node without taints", func() {
			labelPodName := "nginx-with-no-taints-no-tolerations"
			testpod := testPodConfig{
				Name:          labelPodName,
				SchedulerName: "poseidon",
			}

			By("Trying to get a schedulable node")
			schedulableNodes := framework.ListSchedulableNodes(clientset)
			if len(schedulableNodes) < 2 {
				Skip(fmt.Sprintf("Skipping this test case as this requires minimum of two node and only %d nodes avaliable", len(schedulableNodes)))
			}
			nodeOne = schedulableNodes[0]
			nodeTwo = schedulableNodes[1]
			By(fmt.Sprintf("Trying to apply a taint on %s", nodeOne.Name))
			taint := v1.Taint{
				Key:    "cpu-type",
				Value:  "arm64",
				Effect: "PreferNoSchedule",
			}
			framework.AddOrUpdateTaintOnNode(clientset, nodeOne.Name, taint)
			framework.ExpectNodeHasTaint(clientset, nodeOne.Name, &taint)

			defer func() {
				By(fmt.Sprintf("Remove the taint from %s", nodeOne.Name))
				framework.RemoveTaintOffNode(clientset, nodeOne.Name, taint)
				framework.VerifyThatTaintIsGone(clientset, nodeOne.Name, &taint)
				By("Delete the pod")
				err := clientset.CoreV1().Pods(ns).Delete(labelPodName, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred())
				err = f.WaitForPodNotFound(labelPodName, 2*time.Minute)
				Expect(err).NotTo(HaveOccurred())

			}()

			By(fmt.Sprintf("Trying to launch the pod, with taints on nodes %s ", nodeTwo.Name))
			createTestPod(f, testpod)

			By(fmt.Sprintf("Validate if the pod is running on %s as it has no taint", nodeOne.Name))
			framework.ExpectNoError(framework.WaitForPodNotPending(clientset, ns, labelPodName))
			labelPod, err := clientset.CoreV1().Pods(ns).Get(labelPodName, metav1.GetOptions{})
			framework.ExpectNoError(err)
			Expect(labelPod.Spec.NodeName).NotTo(Equal(nodeOne.Name))

		})

	})
})

func getNodeThatCanRunPodWithoutToleration(f *framework.Framework) string {
	By("Trying to launch a pod without a toleration to get a node which can launch it.")
	return runPodAndGetNodeName(f, testPodConfig{Name: "without-toleration", SchedulerName: "poseidon"})
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
