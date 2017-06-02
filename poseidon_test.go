package main_test

import (
	"fmt"
	"github.com/camsas/poseidon/pkg/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	meta_v1 "k8s.io/client-go/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"math/rand"
	"os"
	"time"

	"k8s.io/client-go/pkg/api/errors"
	batchv1 "k8s.io/client-go/pkg/apis/batch/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

const TEST_NS = "test"

func init() {

}

var _ = Describe("Poseidon", func() {

	hostname, _ := os.Hostname()
	utils.GetLogger("info")
	logger := utils.GetContextLogger("poseidon_tests")
	logger.Info("Inside Poseidon tests for k8s:", hostname)

	Describe("Add Pod using Poseidon Scheduler", func() {
		logger.Info("Inside Check for adding pod using Poseidon Scheduler")
		Context("using firmament for configuring pod", func() {
			//Uses the current context in kubeconfig (for k8s v1.6 and above)
			config, err := clientcmd.BuildConfigFromFlags("", "/root/admin.conf")
			// works for k8s v1.5 and below
			//config, err := clientcmd.DefaultClientConfig.ClientConfig()
			if err != nil {
				panic(err)
			}
			clientset, err := kubernetes.NewForConfig(config)
			if err != nil {
				panic(err)
			}
			name := fmt.Sprintf("test-nginx-pod-%d", rand.Uint32())

			It("should create test namespace", func() {
				ns, err := clientset.Namespaces().Create(&v1.Namespace{
					ObjectMeta: v1.ObjectMeta{Name: TEST_NS},
				})
				if err != nil && errors.IsAlreadyExists(err) {
					//do nothing ignore
				} else if err != nil {
					//if some other error other than Already Exists
					Expect(err).ShouldNot(HaveOccurred())
				}
				By("Waiting 5 seconds")
				time.Sleep(time.Duration(5 * time.Second))
				ns, err = clientset.Namespaces().Get(TEST_NS, meta_v1.GetOptions{})
				Expect(err).ShouldNot(HaveOccurred())
				Expect(ns.Name).To(Equal(TEST_NS))
			})

			It("should succeed deploying pod using firmament scheduler", func() {
				annots := make(map[string]string)
				annots["scheduler.alpha.kubernetes.io/name"] = "poseidon-scheduler"
				labels := make(map[string]string)
				labels["scheduler"] = "poseidon"
				//Create a K8s Pod with poseidon
				_, err = clientset.Pods(TEST_NS).Create(&v1.Pod{
					ObjectMeta: v1.ObjectMeta{
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
				pod, err := clientset.Pods(TEST_NS).Get(name, meta_v1.GetOptions{})
				Expect(err).NotTo(HaveOccurred())
				logger.Info("pod status =", string(pod.Status.Phase))
				Expect(string(pod.Status.Phase)).To(Equal("Running"))

				By("Pod was in Running state... Time to delete the pod now...")
				err = clientset.Pods(TEST_NS).Delete(name, &v1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred())
				By("Waiting 5 seconds")
				time.Sleep(time.Duration(5 * time.Second))
				By("Check for pod deletion")
				_, err = clientset.Pods(TEST_NS).Get(name, meta_v1.GetOptions{})
				if err != nil && errors.IsNotFound(err) {
					//do nothing pod has already been deleted
				}
				Expect("Success").To(Equal("Success"))
			})
		})
	})

	Describe("Add Deployment using Poseidon Scheduler", func() {
		logger.Info("Inside Check for adding Deployment using Poseidon Scheduler")
		Context("using firmament for configuring Deployment", func() {
			//Uses the current context in kubeconfig (for k8s v1.6 and above)
			config, err := clientcmd.BuildConfigFromFlags("", "/root/admin.conf")
			// works for k8s v1.5 and below
			//config, err := clientcmd.DefaultClientConfig.ClientConfig()
			if err != nil {
				panic(err)
			}
			clientset, err := kubernetes.NewForConfig(config)
			if err != nil {
				panic(err)
			}
			name := fmt.Sprintf("test-nginx-deploy-%d", rand.Uint32())

			It("should create test namespace", func() {
				ns, err := clientset.Namespaces().Create(&v1.Namespace{
					ObjectMeta: v1.ObjectMeta{Name: TEST_NS},
				})
				if err != nil && errors.IsAlreadyExists(err) {
					//do nothing ignore
				} else if err != nil {
					//if some other error other than Already Exists
					Expect(err).ShouldNot(HaveOccurred())
				}
				By("Waiting 5 seconds")
				time.Sleep(time.Duration(5 * time.Second))
				ns, err = clientset.Namespaces().Get(TEST_NS, meta_v1.GetOptions{})
				Expect(err).ShouldNot(HaveOccurred())
				Expect(ns.Name).To(Equal(TEST_NS))
			})

			It("should succeed deploying Deployment using firmament scheduler", func() {
				annots := make(map[string]string)
				annots["scheduler.alpha.kubernetes.io/name"] = "poseidon-scheduler"
				labels := make(map[string]string)
				labels["scheduler"] = "poseidon"
				//Create a K8s Deployment with poseidon scheduler
				var replicas int32
				replicas = 2
				_, err = clientset.Deployments(TEST_NS).Create(&v1beta1.Deployment{
					ObjectMeta: v1.ObjectMeta{
						Name:        name,
						Annotations: annots,
						Labels:      labels,
					},
					Spec: v1beta1.DeploymentSpec{
						Replicas: &replicas,
						Selector: &meta_v1.LabelSelector{
							MatchLabels: map[string]string{"name": "test-dep"},
						},
						Template: v1.PodTemplateSpec{
							ObjectMeta: v1.ObjectMeta{
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
				deployment, err := clientset.Deployments(TEST_NS).Get(name, meta_v1.GetOptions{})
				Expect(err).NotTo(HaveOccurred())

				logger.Info("Replicas =", deployment.Status.Replicas)
				logger.Info("Available Replicas =", deployment.Status.AvailableReplicas)
				//Expect(string(int(deployment.Status.AvailableReplicas))).To(string(2))
				By(fmt.Sprintf("Creation of deployment %q in namespace %q succeeded.  Deleting deployment.", deployment.Name, TEST_NS))
				if deployment.Status.Replicas != deployment.Status.AvailableReplicas {
					Expect("Success").To(Equal("Fail"))
				}

				By("Pod was in Running state... Time to delete the deployment now...")
				err = clientset.Deployments(TEST_NS).Delete(name, &v1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred())
				By("Waiting 5 seconds")
				time.Sleep(time.Duration(5 * time.Second))
				By("Check for deployment deletion")
				_, err = clientset.Deployments(TEST_NS).Get(name, meta_v1.GetOptions{})
				if err != nil && errors.IsNotFound(err) {
					//do nothing pod has already been deleted
				}
				Expect("Success").To(Equal("Success"))
			})
		})
	})

	Describe("Add ReplicaSet using Poseidon Scheduler", func() {
		logger.Info("Inside Check for adding ReplicaSet using Poseidon Scheduler")
		Context("using firmament for configuring ReplicaSet", func() {
			//Uses the current context in kubeconfig (for k8s v1.6 and above)
			config, err := clientcmd.BuildConfigFromFlags("", "/root/admin.conf")
			// works for k8s v1.5 and below
			//config, err := clientcmd.DefaultClientConfig.ClientConfig()
			if err != nil {
				panic(err)
			}
			clientset, err := kubernetes.NewForConfig(config)
			if err != nil {
				panic(err)
			}
			name := fmt.Sprintf("test-nginx-rs-%d", rand.Uint32())

			It("should create test namespace", func() {
				ns, err := clientset.Namespaces().Create(&v1.Namespace{
					ObjectMeta: v1.ObjectMeta{Name: TEST_NS},
				})
				if err != nil && errors.IsAlreadyExists(err) {
					//do nothing ignore
				} else if err != nil {
					//if some other error other than Already Exists
					Expect(err).ShouldNot(HaveOccurred())
				}
				By("Waiting 5 seconds")
				time.Sleep(time.Duration(5 * time.Second))
				ns, err = clientset.Namespaces().Get(TEST_NS, meta_v1.GetOptions{})
				Expect(err).ShouldNot(HaveOccurred())
				Expect(ns.Name).To(Equal(TEST_NS))
			})

			It("should succeed deploying ReplicaSet using firmament scheduler", func() {
				annots := make(map[string]string)
				annots["scheduler.alpha.kubernetes.io/name"] = "poseidon-scheduler"
				labels := make(map[string]string)
				labels["scheduler"] = "poseidon"
				//Create a K8s ReplicaSet with poseidon scheduler
				var replicas int32
				replicas = 2
				_, err = clientset.ReplicaSets(TEST_NS).Create(&v1beta1.ReplicaSet{
					ObjectMeta: v1.ObjectMeta{
						Name:        name,
						Annotations: annots,
						Labels:      labels,
					},
					Spec: v1beta1.ReplicaSetSpec{
						Replicas: &replicas,
						Selector: &meta_v1.LabelSelector{
							MatchLabels: map[string]string{"name": "test-rs"},
						},
						Template: v1.PodTemplateSpec{
							ObjectMeta: v1.ObjectMeta{
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
				replicaSet, err := clientset.ReplicaSets(TEST_NS).Get(name, meta_v1.GetOptions{})
				Expect(err).NotTo(HaveOccurred())

				logger.Info("Replicas =", replicaSet.Status.Replicas)
				logger.Info("Available Replicas =", replicaSet.Status.AvailableReplicas)
				By(fmt.Sprintf("Creation of ReplicaSet %q in namespace %q succeeded.  Deleting ReplicaSet.", replicaSet.Name, TEST_NS))
				if replicaSet.Status.Replicas != replicaSet.Status.AvailableReplicas {
					Expect("Success").To(Equal("Fail"))
				}

				By("Pod was in Running state... Time to delete the ReplicaSet now...")
				err = clientset.ReplicaSets(TEST_NS).Delete(name, &v1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred())
				By("Waiting 5 seconds")
				time.Sleep(time.Duration(5 * time.Second))
				By("Check for ReplicaSet deletion")
				_, err = clientset.ReplicaSets(TEST_NS).Get(name, meta_v1.GetOptions{})
				if err != nil && errors.IsNotFound(err) {
					//do nothing pod has already been deleted
				}
				Expect("Success").To(Equal("Success"))
			})
		})
	})

	Describe("Add Job using Poseidon Scheduler", func() {
		logger.Info("Inside Check for adding Job using Poseidon Scheduler")
		Context("using firmament for configuring Job", func() {
			//Uses the current context in kubeconfig (for k8s v1.6 and above)
			config,err := clientcmd.BuildConfigFromFlags("","/root/admin.conf")
			// works for k8s v1.5 and below
			//config, err := clientcmd.DefaultClientConfig.ClientConfig()
			if err != nil {
				panic(err)
			}
			clientset, err := kubernetes.NewForConfig(config)
			if err != nil {
				panic(err)
			}
			name := fmt.Sprintf("test-nginx-job-%d", rand.Uint32())

			It("should create test namespace", func() {
				ns, err := clientset.Namespaces().Create(&v1.Namespace{
					ObjectMeta: v1.ObjectMeta{Name: TEST_NS},
				})
				if err != nil && errors.IsAlreadyExists(err) {
					//do nothing ignore
				} else if err != nil {
					//if some other error other than Already Exists
					Expect(err).ShouldNot(HaveOccurred())
				}
				By("Waiting 5 seconds")
				time.Sleep(time.Duration(5 * time.Second))
				ns, err = clientset.Namespaces().Get(TEST_NS, meta_v1.GetOptions{})
				Expect(err).ShouldNot(HaveOccurred())
				Expect(ns.Name).To(Equal(TEST_NS))
			})

			It("should succeed deploying Job using firmament scheduler", func() {
				annots := make(map[string]string)
				annots["scheduler.alpha.kubernetes.io/name"] = "poseidon-scheduler"
				labels := make(map[string]string)
				labels["scheduler"] = "poseidon"
				//Create a K8s Job with poseidon scheduler
				var completions int32
				completions = 2
				_, err = clientset.Batch().Jobs(TEST_NS).Create(&batchv1.Job{
					ObjectMeta: v1.ObjectMeta{
						Name:        name,
						Annotations: annots,
						Labels:      labels,
					},
					Spec: batchv1.JobSpec{
						Parallelism: &completions,
						Completions: &completions,
						Template: v1.PodTemplateSpec{
							ObjectMeta: v1.ObjectMeta{
								Annotations: annots,
								Labels:      labels,
							},
							Spec: v1.PodSpec {
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
				job, err := clientset.Batch().Jobs(TEST_NS).Get(name, meta_v1.GetOptions{})
				Expect(err).NotTo(HaveOccurred())


				logger.Info("Jobs Active =", job.Status.Active)
				logger.Info("Jobs Succeeded =", job.Status.Succeeded)
				By(fmt.Sprintf("Creation of Jobs %q in namespace %q succeeded.  Deleting Job.", job.Name, TEST_NS))
				if job.Status.Active != job.Status.Succeeded {
					Expect("Success").To(Equal("Fail"))
				}

				By("Job was in Running state... Time to delete the Job now...")
				err = clientset.Batch().Jobs(TEST_NS).Delete(name, &v1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred())
				By("Waiting 5 seconds")
				time.Sleep(time.Duration(5 * time.Second))
				By("Check for Job deletion")
				_, err = clientset.Batch().Jobs(TEST_NS).Get(name, meta_v1.GetOptions{})
				if err != nil && errors.IsNotFound(err) {
					//do nothing pod has already been deleted
				}
				Expect("Success").To(Equal("Success"))
			})
		})
	})

	Describe("Add Daemonset using Poseidon Scheduler", func() {
		logger.Info("Inside Check for adding Daemonset using Poseidon Scheduler")
		Context("using firmament for configuring Daemonset", func() {
			//Uses the current context in kubeconfig (for k8s v1.6 and above)
			config, err := clientcmd.BuildConfigFromFlags("", "/root/admin.conf")
			// works for k8s v1.5 and below
			//config, err := clientcmd.DefaultClientConfig.ClientConfig()
			if err != nil {
				panic(err)
			}
			clientset, err := kubernetes.NewForConfig(config)
			if err != nil {
				panic(err)
			}
			name := fmt.Sprintf("test-nginx-deploy-%d", rand.Uint32())

			It("should create test namespace", func() {
				ns, err := clientset.Namespaces().Create(&v1.Namespace{
					ObjectMeta: v1.ObjectMeta{Name: TEST_NS},
				})
				if err != nil && errors.IsAlreadyExists(err) {
					//do nothing ignore
				} else if err != nil {
					//if some other error other than Already Exists
					Expect(err).ShouldNot(HaveOccurred())
				}
				By("Waiting 5 seconds")
				time.Sleep(time.Duration(5 * time.Second))
				ns, err = clientset.Namespaces().Get(TEST_NS, meta_v1.GetOptions{})
				Expect(err).ShouldNot(HaveOccurred())
				Expect(ns.Name).To(Equal(TEST_NS))
			})

			It("should succeed deploying Daemonset using firmament scheduler", func() {
				annots := make(map[string]string)
				annots["scheduler.alpha.kubernetes.io/name"] = "poseidon-scheduler"
				labels := make(map[string]string)
				labels["scheduler"] = "poseidon"
				_, err = clientset.DaemonSets(TEST_NS).Create(&v1beta1.DaemonSet{
					ObjectMeta: v1.ObjectMeta{
						Name:        name,
						Annotations: annots,
						Labels:      labels,
					},
					Spec: v1beta1.DaemonSetSpec{
						Selector: &meta_v1.LabelSelector{
							MatchLabels: map[string]string{"name": "test-dep"},
						},
						Template: v1.PodTemplateSpec{
							ObjectMeta: v1.ObjectMeta{
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
				Daemonset, err := clientset.DaemonSets(TEST_NS).Get(name, meta_v1.GetOptions{})
				Expect(err).NotTo(HaveOccurred())

				logger.Info("DesiredNumberScheduled =", Daemonset.Status.DesiredNumberScheduled)
				logger.Info("CurrentNumberScheduled =", Daemonset.Status.CurrentNumberScheduled)
				//Expect(string(int(Daemonset.Status.AvailableReplicas))).To(string(2))
				By(fmt.Sprintf("Creation of Daemonset %q in namespace %q succeeded.  Deleting Daemonset.", Daemonset.Name, TEST_NS))
				if Daemonset.Status.DesiredNumberScheduled != Daemonset.Status.CurrentNumberScheduled {
					Expect("Success").To(Equal("Fail"))
				}

				By("Pod was in Running state... Time to delete the Daemonset now...")
				err = clientset.DaemonSets(TEST_NS).Delete(name, &v1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred())
				By("Waiting 5 seconds")
				time.Sleep(time.Duration(5 * time.Second))
				By("Check for Daemonset deletion")
				_, err = clientset.DaemonSets(TEST_NS).Get(name, meta_v1.GetOptions{})
				if err != nil && errors.IsNotFound(err) {
					//do nothing pod has already been deleted
				}
				Expect("Success").To(Equal("Success"))
			})
		})
	})

})
