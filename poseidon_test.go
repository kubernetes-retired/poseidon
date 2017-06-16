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

package main_test

import (
	"flag"
	"fmt"
	logger "github.com/golang/glog"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	appsv1beta1 "k8s.io/client-go/pkg/apis/apps/v1beta1"
	batchv1 "k8s.io/client-go/pkg/apis/batch/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"math/rand"
	"os"
	"time"
)

const TEST_NAMESPACE = "test"

var testKubeVersion string
var testKubeConfig string
var clientset *kubernetes.Clientset

func init() {
	// go test -args --testKubeVersion="1.6" --testKubeConfig="/root/admin.conf"
	// To override default values pass --testKubeVersion --testKubeConfig flags
	flag.StringVar(&testKubeVersion, "testKubeVersion", "1.5.6", "Specify kubernetes version eg: 1.5 or 1.6")
	flag.StringVar(&testKubeConfig, "testKubeConfig", "/root/admin.conf", "Specify testKubeConfig path eg: /root/kubeconfig")
}

var _ = Describe("Poseidon", func() {
	flag.Parse()
	var err error
	hostname, _ := os.Hostname()
	logger.Info("Inside Poseidon tests for k8s:", hostname)

	Describe("Add Pod using Poseidon scheduler", func() {
		logger.Info("Inside Check for adding pod using Poseidon scheduler")
		Context("using firmament for configuring pod", func() {
			name := fmt.Sprintf("test-nginx-pod-%d", rand.Uint32())

			It("should succeed deploying pod using firmament scheduler", func() {
				annots := make(map[string]string)
				annots["scheduler.alpha.kubernetes.io/name"] = "poseidon-scheduler"
				labels := make(map[string]string)
				labels["scheduler"] = "poseidon"
				//Create a K8s Pod with poseidon
				_, err = clientset.Pods(TEST_NAMESPACE).Create(&v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:        name,
						Annotations: annots,
						Labels:      labels,
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{{
							Name:            fmt.Sprintf("container-%s", name),
							Image:           "nginx:latest",
							ImagePullPolicy: "IfNotPresent",
						}}},
				})

				Expect(err).NotTo(HaveOccurred())

				By("Waiting for the pod to have running status")
				By("Waiting 10 seconds")
				time.Sleep(time.Duration(10 * time.Second))
				pod, err := clientset.Pods(TEST_NAMESPACE).Get(name, metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred())
				logger.Info("pod status =", string(pod.Status.Phase))
				Expect(string(pod.Status.Phase)).To(Equal("Running"))

				By("Pod was in Running state... Time to delete the pod now...")
				err = clientset.Pods(TEST_NAMESPACE).Delete(name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred())
				By("Waiting 5 seconds")
				time.Sleep(time.Duration(5 * time.Second))
				By("Check for pod deletion")
				_, err = clientset.Pods(TEST_NAMESPACE).Get(name, metav1.GetOptions{})
				if err != nil {
					Expect(errors.IsNotFound(err)).To(Equal(true))
				}
				Expect("Success").To(Equal("Success"))
			})
		})
	})

	Describe("Add Deployment using Poseidon scheduler", func() {
		logger.Info("Inside Check for adding Deployment using Poseidon scheduler")
		Context("using firmament for configuring Deployment", func() {
			name := fmt.Sprintf("test-nginx-deploy-%d", rand.Uint32())

			It("should succeed deploying Deployment using firmament scheduler", func() {
				annots := make(map[string]string)
				annots["scheduler.alpha.kubernetes.io/name"] = "poseidon-scheduler"
				labels := make(map[string]string)
				labels["scheduler"] = "poseidon"
				// Create a K8s Deployment with poseidon scheduler
				var replicas int32
				replicas = 2
				_, err = clientset.AppsV1beta1().Deployments(TEST_NAMESPACE).Create(&appsv1beta1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:        name,
						Annotations: annots,
						Labels:      labels,
					},
					Spec: appsv1beta1.DeploymentSpec{
						Replicas: &replicas,
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"name": "test-dep"},
						},
						Template: v1.PodTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Labels: map[string]string{"name": "test-dep"},
							},
							Spec: v1.PodSpec{
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
				By("Waiting 10 seconds")
				time.Sleep(time.Duration(10 * time.Second))
				deployment, err := clientset.AppsV1beta1().Deployments(TEST_NAMESPACE).Get(name, metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred())

				logger.Info("Replicas =", deployment.Status.Replicas)
				logger.Info("Available Replicas =", deployment.Status.AvailableReplicas)

				By(fmt.Sprintf("Creation of deployment %q in namespace %q succeeded.  Deleting deployment.", deployment.Name, TEST_NAMESPACE))
				if deployment.Status.Replicas != deployment.Status.AvailableReplicas {
					Expect("Success").To(Equal("Fail"))
				}

				By("Pod was in Running state... Time to delete the deployment now...")
				err = clientset.AppsV1beta1().Deployments(TEST_NAMESPACE).Delete(name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred())
				By("Waiting 5 seconds")
				time.Sleep(time.Duration(5 * time.Second))
				By("Check for deployment deletion")
				_, err = clientset.AppsV1beta1().Deployments(TEST_NAMESPACE).Get(name, metav1.GetOptions{})
				if err != nil {
					Expect(errors.IsNotFound(err)).To(Equal(true))
				}
				Expect("Success").To(Equal("Success"))
			})
		})
	})

	Describe("Add ReplicaSet using Poseidon scheduler", func() {
		logger.Info("Inside Check for adding ReplicaSet using Poseidon scheduler")
		Context("using firmament for configuring ReplicaSet", func() {
			name := fmt.Sprintf("test-nginx-rs-%d", rand.Uint32())

			It("should succeed deploying ReplicaSet using firmament scheduler", func() {
				annots := make(map[string]string)
				annots["scheduler.alpha.kubernetes.io/name"] = "poseidon-scheduler"
				labels := make(map[string]string)
				labels["scheduler"] = "poseidon"
				//Create a K8s ReplicaSet with poseidon scheduler
				var replicas int32
				replicas = 2
				_, err = clientset.ReplicaSets(TEST_NAMESPACE).Create(&v1beta1.ReplicaSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:        name,
						Annotations: annots,
						Labels:      labels,
					},
					Spec: v1beta1.ReplicaSetSpec{
						Replicas: &replicas,
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"name": "test-rs"},
						},
						Template: v1.PodTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Labels: map[string]string{"name": "test-rs"},
							},
							Spec: v1.PodSpec{
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
				By("Waiting 10 seconds")
				time.Sleep(time.Duration(10 * time.Second))
				replicaSet, err := clientset.ReplicaSets(TEST_NAMESPACE).Get(name, metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred())

				logger.Info("Replicas =", replicaSet.Status.Replicas)
				logger.Info("Available Replicas =", replicaSet.Status.AvailableReplicas)
				By(fmt.Sprintf("Creation of ReplicaSet %q in namespace %q succeeded.  Deleting ReplicaSet.", replicaSet.Name, TEST_NAMESPACE))
				if replicaSet.Status.Replicas != replicaSet.Status.AvailableReplicas {
					Expect("Success").To(Equal("Fail"))
				}

				By("Pod was in Running state... Time to delete the ReplicaSet now...")
				err = clientset.ReplicaSets(TEST_NAMESPACE).Delete(name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred())
				By("Waiting 5 seconds")
				time.Sleep(time.Duration(5 * time.Second))
				By("Check for ReplicaSet deletion")
				_, err = clientset.ReplicaSets(TEST_NAMESPACE).Get(name, metav1.GetOptions{})
				if err != nil {
					Expect(errors.IsNotFound(err)).To(Equal(true))
				}
				Expect("Success").To(Equal("Success"))
			})
		})
	})

	Describe("Add Job using Poseidon scheduler", func() {
		logger.Info("Inside Check for adding Job using Poseidon scheduler")
		Context("using firmament for configuring Job", func() {
			name := fmt.Sprintf("test-nginx-job-%d", rand.Uint32())

			It("should succeed deploying Job using firmament scheduler", func() {
				annots := make(map[string]string)
				annots["scheduler.alpha.kubernetes.io/name"] = "poseidon-scheduler"
				labels := make(map[string]string)
				labels["scheduler"] = "poseidon"
				//Create a K8s Job with poseidon scheduler
				var parallelism int32 = 2
				_, err = clientset.Batch().Jobs(TEST_NAMESPACE).Create(&batchv1.Job{
					ObjectMeta: metav1.ObjectMeta{
						Name:        name,
						Annotations: annots,
						Labels:      labels,
					},
					Spec: batchv1.JobSpec{
						Parallelism: &parallelism,
						Completions: &parallelism,
						Template: v1.PodTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Annotations: annots,
								Labels:      labels,
							},
							Spec: v1.PodSpec{
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
				By("Waiting 10 seconds")
				time.Sleep(time.Duration(10 * time.Second))
				job, err := clientset.Batch().Jobs(TEST_NAMESPACE).Get(name, metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred())

				logger.Info("Jobs Active =", job.Status.Active)
				logger.Info("Jobs Succeeded =", job.Status.Succeeded)
				By(fmt.Sprintf("Creation of Jobs %q in namespace %q succeeded.  Deleting Job.", job.Name, TEST_NAMESPACE))
				if job.Status.Active != parallelism {
					Expect("Success").To(Equal("Fail"))
				}

				By("Job was in Running state... Time to delete the Job now...")
				err = clientset.Batch().Jobs(TEST_NAMESPACE).Delete(name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred())
				By("Waiting 5 seconds")
				time.Sleep(time.Duration(5 * time.Second))
				By("Check for Job deletion")
				_, err = clientset.Batch().Jobs(TEST_NAMESPACE).Get(name, metav1.GetOptions{})
				if err != nil {
					Expect(errors.IsNotFound(err)).To(Equal(true))
				}
				Expect("Success").To(Equal("Success"))
			})
		})
	})

	Describe("Add Daemonset using Poseidon scheduler", func() {
		logger.Info("Inside Check for adding Daemonset using Poseidon scheduler")
		Context("using firmament for configuring Daemonset", func() {
			name := fmt.Sprintf("test-nginx-deploy-%d", rand.Uint32())

			It("should succeed deploying Daemonset using firmament scheduler", func() {
				annots := make(map[string]string)
				annots["scheduler.alpha.kubernetes.io/name"] = "poseidon-scheduler"
				labels := make(map[string]string)
				labels["scheduler"] = "poseidon"
				_, err = clientset.DaemonSets(TEST_NAMESPACE).Create(&v1beta1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:        name,
						Annotations: annots,
						Labels:      labels,
					},
					Spec: v1beta1.DaemonSetSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"name": "test-dep"},
						},
						Template: v1.PodTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Labels: map[string]string{"name": "test-dep"},
							},
							Spec: v1.PodSpec{
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

				By("Waiting for the Daemonset to have running status")
				By("Waiting 10 seconds")
				time.Sleep(time.Duration(10 * time.Second))
				Daemonset, err := clientset.DaemonSets(TEST_NAMESPACE).Get(name, metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred())

				logger.Info("DesiredNumberScheduled =", Daemonset.Status.DesiredNumberScheduled)
				logger.Info("CurrentNumberScheduled =", Daemonset.Status.CurrentNumberScheduled)
				By(fmt.Sprintf("Creation of Daemonset %q in namespace %q succeeded.  Deleting Daemonset.", Daemonset.Name, TEST_NAMESPACE))
				if Daemonset.Status.DesiredNumberScheduled != Daemonset.Status.CurrentNumberScheduled {
					Expect("Success").To(Equal("Fail"))
				}

				By("Pod was in Running state... Time to delete the Daemonset now...")
				err = clientset.DaemonSets(TEST_NAMESPACE).Delete(name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred())
				By("Waiting 5 seconds")
				time.Sleep(time.Duration(5 * time.Second))
				By("Check for Daemonset deletion")
				_, err = clientset.DaemonSets(TEST_NAMESPACE).Get(name, metav1.GetOptions{})
				if err != nil {
					Expect(errors.IsNotFound(err)).To(Equal(true))
				}
				Expect("Success").To(Equal("Success"))
			})
		})
	})

})

var _ = BeforeSuite(func() {
	var config *rest.Config
	var err error
	logger.Infof("Kube version %s", testKubeVersion)
	if testKubeVersion == "1.6" {
		config, err = clientcmd.BuildConfigFromFlags("", testKubeConfig)
	} else {
		config, err = clientcmd.DefaultClientConfig.ClientConfig()
	}
	if err != nil {
		panic(err)
	}
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	createNamespace(clientset)
})

var _ = AfterSuite(func() {
	// Delete namespace
	err := clientset.Namespaces().Delete(TEST_NAMESPACE, &metav1.DeleteOptions{})
	// Delete all pods
	err = clientset.Pods(TEST_NAMESPACE).DeleteCollection(&metav1.DeleteOptions{}, metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
})

func createNamespace(clientset *kubernetes.Clientset) {
	ns, err := clientset.Namespaces().Create(&v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: TEST_NAMESPACE},
	})
	if err != nil {
		if errors.IsAlreadyExists(err) {
			return
		} else {
			Expect(err).ShouldNot(HaveOccurred())
		}
	}
	By("Waiting 5 seconds")
	time.Sleep(time.Duration(5 * time.Second))
	ns, err = clientset.Namespaces().Get(TEST_NAMESPACE, metav1.GetOptions{})
	Expect(err).ShouldNot(HaveOccurred())
	Expect(ns.Name).To(Equal(TEST_NAMESPACE))
}
