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
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	"time"
)

func (f *Framework) createNamespace(c clientset.Interface) (*v1.Namespace, error) {
	var got *v1.Namespace
	if err := wait.PollImmediate(2*time.Second, 5*time.Minute, func() (bool, error) {
		var err error
		Logf("Trying to create namespace %v", f.TestingNS)
		got, err = c.CoreV1().Namespaces().Create(&v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: f.TestingNS},
		})
		if errors.IsAlreadyExists(err) {
			Logf("%v namespace already exist or is still terminating wait and create again %v", f.TestingNS, err)
			return false, nil
		}
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

func (f *Framework) deleteNamespaceIfExist(nsName string) error {
	if _, err := f.ClientSet.CoreV1().Namespaces().Get(nsName, metav1.GetOptions{}); err != nil {
		if errors.IsNotFound(err) {
			Logf("%v dosent not exist, no need to delete non existing namespace", nsName)
			return nil
		} else {
			Logf("error occurred while fetching %v for deleting", nsName)
			return err
		}
	} else {
		//delete the namespace as it exist
		Logf("Deleting %v namespace as it exists", nsName)
		if err = f.deleteNamespace(nsName); err != nil {
			Logf("Unable to delete %v namespace, error %v occurred", nsName, err)
			return err
		}
	}
	return nil
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
