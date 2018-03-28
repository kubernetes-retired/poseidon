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

package firmament

import (
	"github.com/golang/mock/gomock"

	"testing"
)

func Test_New(t *testing.T) {
	firClient, conn, err := New("127.0.0.1:6090")
	defer conn.Close()
	if firClient == nil || conn == nil || err != nil {
		t.Error("Failed to start the client")
	}
}

func Test_AddNodeStats(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	firmamentClient := NewMockFirmamentSchedulerClient(mockCtrl)
	firmamentClient.EXPECT().AddNodeStats(gomock.Any(), gomock.Any()).Return(
		&ResourceStatsResponse{Type: NodeReplyType_NODE_ADDED_OK}, nil)
	AddNodeStats(firmamentClient, nil)
}

func Test_AddTaskStats(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	firmamentClient := NewMockFirmamentSchedulerClient(mockCtrl)
	firmamentClient.EXPECT().AddTaskStats(gomock.Any(), gomock.Any()).Return(
		&TaskStatsResponse{Type: TaskReplyType_TASK_SUBMITTED_OK}, nil)
	AddTaskStats(firmamentClient, nil)
}

func Test_NodeUpdated(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	firmamentClient := NewMockFirmamentSchedulerClient(mockCtrl)
	firmamentClient.EXPECT().NodeUpdated(gomock.Any(), gomock.Any()).Return(
		&NodeUpdatedResponse{Type: NodeReplyType_NODE_UPDATED_OK}, nil)
	NodeUpdated(firmamentClient, nil)
}

func Test_NodeRemoved(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	firmamentClient := NewMockFirmamentSchedulerClient(mockCtrl)
	firmamentClient.EXPECT().NodeRemoved(gomock.Any(), gomock.Any()).Return(
		&NodeRemovedResponse{Type: NodeReplyType_NODE_REMOVED_OK}, nil)

	NodeRemoved(firmamentClient, nil)
}

func Test_NodeFailed(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	firmamentClient := NewMockFirmamentSchedulerClient(mockCtrl)
	firmamentClient.EXPECT().NodeFailed(gomock.Any(), gomock.Any()).Return(
		&NodeFailedResponse{Type: NodeReplyType_NODE_FAILED_OK}, nil)
	NodeFailed(firmamentClient, nil)
}

func Test_NodeAdded(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	firmamentClient := NewMockFirmamentSchedulerClient(mockCtrl)
	firmamentClient.EXPECT().NodeAdded(gomock.Any(), gomock.Any()).Return(
		&NodeAddedResponse{Type: NodeReplyType_NODE_ADDED_OK}, nil)
	NodeAdded(firmamentClient, nil)
}

func Test_TaskUpdated(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	firmamentClient := NewMockFirmamentSchedulerClient(mockCtrl)
	firmamentClient.EXPECT().TaskUpdated(gomock.Any(), gomock.Any()).Return(
		&TaskUpdatedResponse{Type: TaskReplyType_TASK_UPDATED_OK}, nil)
	TaskUpdated(firmamentClient, nil)
}

func Test_TaskSubmitted(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	firmamentClient := NewMockFirmamentSchedulerClient(mockCtrl)
	firmamentClient.EXPECT().TaskSubmitted(gomock.Any(), gomock.Any()).Return(
		&TaskSubmittedResponse{Type: TaskReplyType_TASK_SUBMITTED_OK}, nil)
	TaskSubmitted(firmamentClient, nil)
}

func Test_TaskRemoved(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	firmamentClient := NewMockFirmamentSchedulerClient(mockCtrl)
	firmamentClient.EXPECT().TaskRemoved(gomock.Any(), gomock.Any()).Return(
		&TaskRemovedResponse{Type: TaskReplyType_TASK_REMOVED_OK}, nil)
	TaskRemoved(firmamentClient, nil)
}

func Test_TaskFailed(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	firmamentClient := NewMockFirmamentSchedulerClient(mockCtrl)
	firmamentClient.EXPECT().TaskFailed(gomock.Any(), gomock.Any()).Return(
		&TaskFailedResponse{Type: TaskReplyType_TASK_FAILED_OK}, nil)
	TaskFailed(firmamentClient, nil)
}

func Test_TaskCompleted(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	firmamentClient := NewMockFirmamentSchedulerClient(mockCtrl)
	firmamentClient.EXPECT().TaskCompleted(gomock.Any(), gomock.Any()).Return(
		&TaskCompletedResponse{Type: TaskReplyType_TASK_COMPLETED_OK}, nil)
	TaskCompleted(firmamentClient, nil)
}

func Test_Schedule(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	firmamentClient := NewMockFirmamentSchedulerClient(mockCtrl)
	firmamentClient.EXPECT().Schedule(gomock.Any(), gomock.Any()).Return(
		&SchedulingDeltas{}, nil)
	Schedule(firmamentClient)
}
