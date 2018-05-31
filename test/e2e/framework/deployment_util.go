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
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
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
	replicas := int32(0)
	rhl := int32(0)
	d.Spec.Replicas = &replicas
	d.Spec.RevisionHistoryLimit = &rhl
	d.Spec.Paused = true
	err := wait.Poll(10*time.Millisecond, 1*time.Minute, func() (bool, error) {
		_, err := f.ClientSet.ExtensionsV1beta1().Deployments(d.Namespace).Update(d)
		if err == nil {
			return true, nil
		}
		// Retry only on update conflict.
		if errors.IsConflict(err) {
			return false, nil
		}
		return false, err
	})
	if err != nil {
		return err
	}
	replicaSets, err := f.listReplicaSets(d)
	if err != nil {
		return err
	}
	for _, rs := range replicaSets {
		if err := f.ClientSet.ExtensionsV1beta1().ReplicaSets(rs.Namespace).Delete(rs.Name, &metav1.DeleteOptions{}); err != nil {
			return err
		}
	}
	err = wait.Poll(Poll, JobTimeout, func() (bool, error) {
		err := f.ClientSet.ExtensionsV1beta1().Deployments(d.Namespace).Delete(d.Name, &metav1.DeleteOptions{})
		if err != nil {
			return false, err
		}
		return true, nil
	})
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
			return fmt.Errorf("unable to fetch the deployment %v. error: %v", deploymentName, err)
		}
	}
	// delete the deployment
	// Note: we use apps/v1beta1 since and test use extension/v1beta for deployments
	// apps/v1beta1 is used by poseidon and firmament deployment
	// TODO(shiv): Need to move the tests to also use apps v1beta1 api
	if err := f.ClientSet.AppsV1beta1().Deployments(nsName).Delete(deploymentName, &metav1.DeleteOptions{}); err != nil {
		if errors.IsNotFound(err) {
			Logf("Deployment %v doesn't exist", deploymentName)
			return nil
		} else {
			return fmt.Errorf("unable to delete the deployment %v from namespace %v. error: %v", deploymentName, nsName, err)
		}
	}
	return nil
}

// DeletePoseidonClusterRole deletes a cluster role and role binding
func (f *Framework) DeletePoseidonClusterRole(clusterRole string, nsName string) error {
	var errs []error
	err := f.ClientSet.RbacV1().ClusterRoleBindings().Delete(clusterRole, &metav1.DeleteOptions{})
	if !errors.IsNotFound(err) {
		errs = append(errs, fmt.Errorf("error deleting cluster role binding %v", clusterRole))
	}

	// now deleting the cluster role
	err = f.ClientSet.RbacV1().ClusterRoles().Delete(clusterRole, &metav1.DeleteOptions{})
	if !errors.IsNotFound(err) {
		errs = append(errs, fmt.Errorf("error deleting cluster role %v", clusterRole))
	}

	// now deleting the service account
	err = f.ClientSet.CoreV1().ServiceAccounts(nsName).Delete(clusterRole, &metav1.DeleteOptions{})
	if !errors.IsNotFound(err) {
		errs = append(errs, fmt.Errorf("error deleting service account %v", clusterRole))
	}
	return utilerrors.NewAggregate(errs)
}

// DeleteService deletes a service
func (f *Framework) DeleteService(nsName string, serviceName string) error {

	err := f.ClientSet.CoreV1().Services(nsName).Delete(serviceName, &metav1.DeleteOptions{})
	if !errors.IsNotFound(err) {
		return err
	}
	return nil
}

// ListReplicaSets returns a slice of RSes the given deployment targets.
// Note that this does NOT attempt to reconcile ControllerRef (adopt/orphan),
// because only the controller itself should do that.
// However, it does filter out anything whose ControllerRef doesn't match.
func (f *Framework) listReplicaSets(deployment *extensions.Deployment) ([]*extensions.ReplicaSet, error) {
	namespace := deployment.Namespace
	selector, err := metav1.LabelSelectorAsSelector(deployment.Spec.Selector)
	if err != nil {
		return nil, err
	}
	options := metav1.ListOptions{LabelSelector: selector.String()}
	rsList, err := f.ClientSet.ExtensionsV1beta1().ReplicaSets(namespace).List(options)
	if err != nil {
		return nil, err
	}
	var ret []*extensions.ReplicaSet
	for i := range rsList.Items {
		ret = append(ret, &rsList.Items[i])
	}
	if err != nil {
		return nil, err
	}
	// Only include those whose ControllerRef matches the Deployment.
	owned := make([]*extensions.ReplicaSet, 0, len(ret))
	for _, rs := range ret {
		if metav1.IsControlledBy(rs, deployment) {
			owned = append(owned, rs)
		}
	}
	return owned, nil
}
