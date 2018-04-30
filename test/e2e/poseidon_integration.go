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
	"github.com/golang/glog"
	"github.com/kubernetes-sigs/poseidon/test/e2e/framework"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"math/rand"
	"os"
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

})
