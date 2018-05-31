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

	batch "k8s.io/api/batch/v1"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
)

const (
	// How long to wait for a job to finish.
	JobTimeout = 15 * time.Minute

	// Job selector name
	JobSelectorKey = "name"
)

// GetJob uses c to get the Job in namespace ns named name. If the returned error is nil, the returned Job is valid.
func GetJob(c clientset.Interface, ns, name string) (*batch.Job, error) {
	return c.BatchV1().Jobs(ns).Get(name, metav1.GetOptions{})
}

// GetJobPods returns a list of Pods belonging to a Job.
func GetJobPods(c clientset.Interface, ns, jobName string) (*v1.PodList, error) {
	label := labels.SelectorFromSet(labels.Set(map[string]string{JobSelectorKey: "test-job"}))
	options := metav1.ListOptions{LabelSelector: label.String()}
	return c.CoreV1().Pods(ns).List(options)
}

// WaitForAllJobPodsRunning wait for all pods for the Job named JobName in namespace ns to become Running.  Only use
// when pods will run for a long time, or it will be racy.
func (f *Framework) WaitForAllJobPodsRunning(jobName string, parallelism int32) error {
	return wait.Poll(Poll, JobTimeout, func() (bool, error) {
		pods, err := GetJobPods(f.ClientSet, f.Namespace.Name, jobName)
		if err != nil {
			return false, err
		}
		count := int32(0)
		for _, p := range pods.Items {
			if p.Status.Phase == v1.PodRunning {
				count++
			}
		}
		return count == parallelism, nil
	})
}

// WaitForJobDelete waits for job to be removed
func (f *Framework) WaitForJobDelete(job *batchv1.Job) error {
	// at this point only dead pods are left, that should be removed
	selector, _ := metav1.LabelSelectorAsSelector(job.Spec.Selector)
	options := metav1.ListOptions{LabelSelector: selector.String()}
	podList, err := f.ClientSet.CoreV1().Pods(job.Namespace).List(options)
	if err != nil {
		return err
	}

	err = wait.Poll(Poll, JobTimeout, func() (bool, error) {
		err := f.ClientSet.BatchV1().Jobs(f.Namespace.Name).Delete(job.Name, &metav1.DeleteOptions{})
		if err != nil {
			return false, err
		}
		return true, nil
	})
	if err == wait.ErrWaitTimeout {
		return fmt.Errorf("job %q not deleted", job.Name)
	}

	errList := []error{}
	for _, pod := range podList.Items {
		if err := f.ClientSet.CoreV1().Pods(job.Namespace).Delete(pod.Name, &metav1.DeleteOptions{}); err != nil {
			// ignores the error when the pod isn't found
			if !errors.IsNotFound(err) {
				errList = append(errList, err)
			}
		}
	}
	if len(errList) > 0 {
		return utilerrors.NewAggregate(errList)
	}
	return err
}
