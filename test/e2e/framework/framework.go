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
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var kubeConfig = flag.String(clientcmd.RecommendedConfigPathFlag, os.Getenv(clientcmd.RecommendedConfigPathEnvVar), "Path to kubeconfig containing embedded authinfo.")
var kubectlPath = flag.String("kubectl-path", "kubectl", "The kubectl binary to use. For development, you might use 'cluster/kubectl.sh' here.")
var poseidonVersion = flag.String("poseidonVersion", "version", "The poseidon image version")
var gcrProject = flag.String("gcrProject", "google_containers", "The gcloud project")
var testNamespace = flag.String("testNamespace", "poseidon-test", "The namespace to use for test")
var clusterRole = flag.String("clusterRole", os.Getenv("CLUSTERROLE"), "The cluster role")

const (
	poseidonDeploymentName  = "poseidon"
	firmamentDeploymentName = "firmament-scheduler"
)

func init() {
	flag.Parse()
	getKubeConfigFromEnv()
}

// Framework supports common operations used by e2e tests; it will keep a client & a namespace for you.
// Eventual goal is to merge this with integration test framework.
type Framework struct {
	BaseName  string
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
	return NewFramework(baseName, options)
}

// NewFramework makes a new framework and sets up a BeforeEach/AfterEach
func NewFramework(baseName string, options FrameworkOptions) *Framework {
	f := &Framework{
		BaseName:  baseName,
		Options:   options,
		ClientSet: nil,
		TestingNS: *testNamespace,
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

	Logf("Poseidon test are pointing to %v", *kubeConfig)

	if err := f.DeleteService(f.TestingNS, "poseidon"); err != nil {
		Logf("Error deleting service poseidon: %v", err)
	}
	if err := f.DeleteService(f.TestingNS, "firmament-service"); err != nil {
		Logf("Error deleting service firmament-service: %v", err)
	}

	if len(*clusterRole) == 0 {
		*clusterRole = "poseidon"
	}
	if err := f.DeletePoseidonClusterRole(*clusterRole, f.TestingNS); err != nil {
		Logf("Error deleting cluster role: %v", err)
	}

	// This is needed if we have a dirty test run which leaves the pods and deployments hanging
	if err := f.deleteNamespaceIfExist(f.TestingNS); err != nil {
		Logf("Error occurred when delete namespace if exist: %v", err)
	}
	if err := f.DeleteDeploymentIfExist(f.TestingNS, poseidonDeploymentName); err != nil {
		Logf("Error occurred when delete deployment if exist: %v", err)
	}
	if err := f.DeleteDeploymentIfExist(f.TestingNS, firmamentDeploymentName); err != nil {
		Logf("Error occurred when delete deployment if exist: %v", err)
	}

	f.Namespace, err = f.createNamespace(f.ClientSet)
	Expect(err).NotTo(HaveOccurred())

	Logf("After namespace creation %v", f.Namespace)

	err = f.CreateFirmament()
	Expect(err).NotTo(HaveOccurred())

	err = f.CreatePoseidon()
	Expect(err).NotTo(HaveOccurred())

}

// AfterEach deletes the namespace, after reading its events.
func (f *Framework) AfterEach() {
	//delete ns
	var err error

	if f.ClientSet == nil {
		Expect(f.ClientSet).To(Not(Equal(nil)))
	}

	// Fetch Poseidon and Firmament logs before ending the test suite
	f.FetchLogsFromFirmament(f.TestingNS)
	f.FetchLogsFromPoseidon(f.TestingNS)
	Logf("Delete namespace called")
	err = f.deleteNamespace(f.TestingNS)
	Expect(err).NotTo(HaveOccurred())

	err = f.DeleteDeploymentIfExist(f.TestingNS, poseidonDeploymentName)
	Expect(err).NotTo(HaveOccurred())

	err = f.DeleteDeploymentIfExist(f.TestingNS, firmamentDeploymentName)
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

// CreateFirmament create firmament deployment
func (f *Framework) CreateFirmament() error {
	Logf("Create Firmament Deployment")
	deployment, err := f.createFirmamentDeployment()
	if err != nil {
		return fmt.Errorf("failed to create Firmament Deployment. error: %v", err)
	}

	Logf("Create Firmament Service")
	if err := f.createFirmamentService(); err != nil {
		return fmt.Errorf("failed to create Firmament Service. error: %v", err)
	}

	if err := f.WaitForDeploymentComplete(deployment); err != nil {
		return fmt.Errorf("timeout to wait deployment: %v to complete. error: %v", deployment, err)
	}
	return nil
}

// CreatePoseidon create firmament deployment
func (f *Framework) CreatePoseidon() error {
	Logf("Create Poseidon ClusterRoleBinding")
	if err := f.createPoseidonClusterRoleBinding(); err != nil {
		return fmt.Errorf("failed to create Poseidon ClusterRoleBinding. error: %v", err)
	}

	Logf("Create Poseidon ClusterRole")
	if err := f.createPoseidonClusterRole(); err != nil {
		return fmt.Errorf("failed to create Poseidon ClusterRole. error: %v", err)
	}

	Logf("Create Poseidon ServiceAccount")
	if err := f.createPoseidonServiceAccount(); err != nil {
		return fmt.Errorf("failed to create Poseidon ServiceAccount. error: %v", err)
	}

	Logf("Create Poseidon Service")
	if err := f.createPoseidonService(); err != nil {
		return fmt.Errorf("failed to create Poseidon Service. error: %v", err)
	}

	Logf("Create Poseidon Deployment")
	deployment, err := f.createPoseidonDeployment()
	if err != nil {
		return fmt.Errorf("failed to create Poseidon Deployment. error: %v", err)
	}

	if err := f.WaitForDeploymentComplete(deployment); err != nil {
		return fmt.Errorf("timeout to wait deployment: %v to complete. error: %v", deployment, err)
	}
	return nil
}

// KubectlCmd runs the kubectl executable through the wrapper script.
func KubectlCmd(args ...string) *exec.Cmd {
	defaultArgs := []string{}

	if kubeConfig != nil {
		defaultArgs = append(defaultArgs, "--"+clientcmd.RecommendedConfigPathFlag+"="+*kubeConfig)

	}
	Logf("kubeConfig file in KubectlCmd %v %v", *kubeConfig, defaultArgs)
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

func getKubeConfigFromEnv() {

	if *kubeConfig == "" {
		//read the config from the env
		*kubeConfig = path.Join(os.Getenv("HOME"), clientcmd.RecommendedHomeDir, clientcmd.RecommendedFileName)

	}
	Logf("Location of the kubeconfig file %v", *kubeConfig)
}

func (f *Framework) createFirmamentService() error {
	_, err := f.ClientSet.CoreV1().Services(f.TestingNS).Create(&v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "firmament-service",
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{{
				Protocol:   "TCP",
				Port:       9090,
				TargetPort: intstr.FromInt(9090),
			}},
			Selector: map[string]string{
				"scheduler": "firmament",
			},
		},
	})
	return err
}

func (f *Framework) createFirmamentDeployment() (*v1beta1.Deployment, error) {
	var replicas, revisionHistoryLimit int32
	replicas = 1
	revisionHistoryLimit = 10
	deployment, err := f.ClientSet.ExtensionsV1beta1().Deployments(f.TestingNS).Create(&v1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{"scheduler": "firmament"},
			Name:   "firmament-scheduler",
		},
		Spec: v1beta1.DeploymentSpec{
			Replicas:             &replicas,
			RevisionHistoryLimit: &revisionHistoryLimit,
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"scheduler": "firmament"},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:    "firmament-scheduler",
							Image:   "huaweifirmament/firmament:with_ephemeral_avoidpods",
							Command: []string{"/firmament/build/src/firmament_scheduler", "--flagfile=/firmament/config/firmament_scheduler_cpu_mem.cfg"},
							Ports:   []v1.ContainerPort{{ContainerPort: 9090}},
						},
					},
					HostNetwork: true,
					HostPID:     false,
					Volumes:     []v1.Volume{},
				},
			},
		},
	})
	return deployment, err
}

func (f *Framework) createPoseidonClusterRoleBinding() error {
	_, err := f.ClientSet.RbacV1().ClusterRoleBindings().Create(&rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "poseidon",
			Labels: map[string]string{"kubernetes.io/bootstrapping": "rbac-defaults"},
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "poseidon",
				Namespace: f.TestingNS,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     "poseidon",
			APIGroup: "rbac.authorization.k8s.io",
		},
	})
	return err
}

func (f *Framework) createPoseidonClusterRole() error {
	_, err := f.ClientSet.RbacV1().ClusterRoles().Create(&rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "poseidon",
			Labels: map[string]string{"kubernetes.io/bootstrapping": "rbac-defaults"},
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"events"},
				Verbs:     []string{"create", "patch", "update"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"endpoints"},
				Verbs:     []string{"create"},
			},
			{
				APIGroups:     []string{""},
				ResourceNames: []string{"poseidon"},
				Resources:     []string{"endpoints"},
				Verbs:         []string{"delete", "get", "patch", "update"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"nodes"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"pods"},
				Verbs:     []string{"delete", "get", "list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"bindings", "pods/binding"},
				Verbs:     []string{"create"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"pods/status"},
				Verbs:     []string{"patch", "update"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"replicationcontrollers", "services"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"apps", "extensions"},
				Resources: []string{"replicasets"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"apps"},
				Resources: []string{"statefulsets"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"policy"},
				Resources: []string{"poddisruptionbudgets"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"persistentvolumeclaims", "persistentvolumes"},
				Verbs:     []string{"get", "list", "watch"},
			},
		},
	})
	return err
}

func (f *Framework) createPoseidonServiceAccount() error {
	_, err := f.ClientSet.CoreV1().ServiceAccounts(f.TestingNS).Create(&v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{Name: "poseidon"},
	})
	return err
}

func (f *Framework) createPoseidonService() error {
	_, err := f.ClientSet.CoreV1().Services(f.TestingNS).Create(&v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "poseidon",
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{{
				Protocol:   "TCP",
				Port:       9091,
				TargetPort: intstr.FromInt(9091),
			}},
			Selector: map[string]string{
				"poseidonservice": "poseidon",
			},
		},
	})
	return err
}

func (f *Framework) createPoseidonDeployment() (*v1beta1.Deployment, error) {
	var replicas, revisionHistoryLimit int32
	replicas = 1
	revisionHistoryLimit = 10
	privileged := false
	poseidonImage := fmt.Sprintf("gcr.io/%s/poseidon-amd64:%s", *gcrProject, *poseidonVersion)
	deployment, err := f.ClientSet.ExtensionsV1beta1().Deployments(f.TestingNS).Create(&v1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{"component": "poseidon", "tier": "control-plane", "poseidonservice": "poseidon"},
			Name:   "poseidon",
		},
		Spec: v1beta1.DeploymentSpec{
			Replicas:             &replicas,
			RevisionHistoryLimit: &revisionHistoryLimit,
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"component": "poseidon", "tier": "control-plane", "poseidonservice": "poseidon", "version": "first"},
				},
				Spec: v1.PodSpec{
					ServiceAccountName: "poseidon",
					Containers: []v1.Container{
						{
							Name:    "poseidon",
							Image:   poseidonImage,
							Command: []string{"/poseidon", "--logtostderr", "--kubeConfig=", "--kubeVersion=1.6", "--firmamentAddress=firmament-service.poseidon-test", "--firmamentPort=9090"},
						},
					},
					InitContainers: []v1.Container{
						{
							Name:            "init-firmamentservice",
							Image:           "radial/busyboxplus:curl",
							Command:         []string{"sh", "-c", "until nslookup firmament-service.poseidon-test; do echo waiting for firmamentservice; sleep 1; done;"},
							SecurityContext: &v1.SecurityContext{Privileged: &privileged},
							VolumeMounts:    []v1.VolumeMount{},
						},
					},
					HostNetwork: false,
					HostPID:     false,
					Volumes:     []v1.Volume{},
				},
			},
		},
	})
	return deployment, err
}
