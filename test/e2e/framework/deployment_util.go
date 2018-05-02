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
	"fmt"
	"time"

	extensions "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
)

// WaitForDeploymentComplete waits for the deployment to complete, and don't check if rolling update strategy is broken.
func (f *Framework) WaitForDeploymentComplete(d *extensions.Deployment) error {
	return waitForDeploymentCompleteNoRollingCheck(f.ClientSet, d, Poll, pollLongTimeout)
}

// Waits for the deployment to complete.
func waitForDeploymentCompleteNoRollingCheck(c clientset.Interface, d *extensions.Deployment, pollInterval, pollTimeout time.Duration) error {
	var (
		deployment *extensions.Deployment
		reason     string
	)

	err := wait.PollImmediate(pollInterval, pollTimeout, func() (bool, error) {
		var err error
		deployment, err = c.ExtensionsV1beta1().Deployments(d.Namespace).Get(d.Name, metav1.GetOptions{})
		if err != nil {
			Logf("Waiting for deployment error %v", err)
			return false, err
		}

		// When the deployment status and its underlying resources reach the desired state, we're done
		if deploymentComplete(d, &deployment.Status) {
			return true, nil
		}

		reason = fmt.Sprintf("deployment status: %#v", deployment.Status)
		Logf(reason)

		return false, nil
	})

	if err == wait.ErrWaitTimeout {
		err = fmt.Errorf("%s", reason)
	}
	if err != nil {
		return fmt.Errorf("error waiting for deployment %q status to match expectation: %v", d.Name, err)
	}
	return nil
}

// deploymentComplete considers a deployment to be complete once all of its desired replicas
// are updated and available, and no old pods are running.
func deploymentComplete(deployment *extensions.Deployment, newStatus *extensions.DeploymentStatus) bool {
	return newStatus.UpdatedReplicas == *(deployment.Spec.Replicas) &&
		newStatus.Replicas == *(deployment.Spec.Replicas) &&
		newStatus.AvailableReplicas == *(deployment.Spec.Replicas) &&
		newStatus.ObservedGeneration >= deployment.Generation
}

// WaitForDeploymentDelete waits for the Deployment to be removed
func (f *Framework) WaitForDeploymentDelete(d *extensions.Deployment) error {
	err := wait.Poll(Poll, JobTimeout, func() (bool, error) {
		err := f.ClientSet.ExtensionsV1beta1().Deployments(d.Namespace).Delete(d.Name, &metav1.DeleteOptions{})
		if err != nil {
			return false, err
		}
		return true, nil
	})
	if err == wait.ErrWaitTimeout {
		err = fmt.Errorf("Deployment %q not deleted", d.Name)
	}
	return err
}

// DeleteDeploymentIfExist deletes a deployment if it exists
func (f *Framework) DeleteDeploymentIfExist(nsName string, deploymentName string) error {

	if _, err := f.ClientSet.AppsV1beta1().Deployments(nsName).Get(deploymentName, metav1.GetOptions{}); err != nil {
		if errors.IsNotFound(err) {
			Logf("%v deployment doesn't exist", deploymentName)
			return nil
		} else {
			// error occurred while trying to fetch the deployment info
			Logf("Unable to fetch the deployment %v", deploymentName)
			return err
		}
	}
	// delete the deployment
	// Note: we use apps/v1beta1 since and test use extension/v1beta for deployments
	// apps/v1beta1 is used by poseidon and firmament deployment
	// TODO(shiv): Need to move the tests to also use apps v1beta1 api
	err := f.ClientSet.AppsV1beta1().Deployments(nsName).Delete(deploymentName, &metav1.DeleteOptions{})
	if errors.IsNotFound(err) {
		Logf("%v deployment doesn't exist", deploymentName)
		return nil
	} else {
		Logf("Unable to delete the deployment %v from namespace %v", deploymentName, nsName)
		return err
	}
}

// DeletePoseidonClusterRole deletes a cluster role and role binding
func (f *Framework) DeletePoseidonClusterRole(clusterRole string, nsName string) error {

	err := f.ClientSet.RbacV1().ClusterRoleBindings().Delete(clusterRole, &metav1.DeleteOptions{})
	if !errors.IsNotFound(err) {
		Logf("Error deleting cluster role binding %v", clusterRole)
	}

	// now deleting the cluster role
	err = f.ClientSet.RbacV1().ClusterRoles().Delete(clusterRole, &metav1.DeleteOptions{})
	if !errors.IsNotFound(err) {
		Logf("Error deleting cluster role %v", clusterRole)
	}

	// now deleting the service account
	err = f.ClientSet.CoreV1().ServiceAccounts(nsName).Delete(clusterRole, &metav1.DeleteOptions{})
	if !errors.IsNotFound(err) {
		Logf("Error deleting service account %v", clusterRole)
	}
	return nil
}

// DeletePoseidonClusterRole deletes a cluster role and role binding
func (f *Framework) DeleteService(nsName string, serviceName string) error {

	err := f.ClientSet.CoreV1().Services(nsName).Delete(serviceName, &metav1.DeleteOptions{})
	if !errors.IsNotFound(err) {
		Logf("Error deleting %v service", serviceName)
	}
	return nil
}
