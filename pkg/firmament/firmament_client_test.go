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

package firmament_test

import (
	"github.com/camsas/poseidon/pkg/firmament"
	"github.com/camsas/poseidon/pkg/mock_firmament"
	"github.com/golang/mock/gomock"

	"testing"
)

func Test_New(t *testing.T) {
	firClient, conn, err := firmament.New("127.0.0.1:6090")
	defer conn.Close()
	if firClient == nil || conn == nil || err != nil {

		t.Error("Failed to start the client")
	}
}

func Test_AddNodeStats(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	firmamentClient := mock_firmament.NewMockFirmamentSchedulerClient(mockCtrl)
	firmamentClient.EXPECT().AddNodeStats(gomock.Any(), gomock.Any()).Return(
		&firmament.ResourceStatsResponse{Type: firmament.NodeReplyType_NODE_ADDED_OK}, nil)

	firmament.AddNodeStats(firmamentClient, nil)
}

func Test_AddTaskStats(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	firmamentClient := mock_firmament.NewMockFirmamentSchedulerClient(mockCtrl)
	firmamentClient.EXPECT().AddTaskStats(gomock.Any(), gomock.Any()).Return(
		&firmament.TaskStatsResponse{Type: firmament.TaskReplyType_TASK_SUBMITTED_OK}, nil)

	firmament.AddTaskStats(firmamentClient, nil)
}

func Test_NodeUpdated(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	firmamentClient := mock_firmament.NewMockFirmamentSchedulerClient(mockCtrl)
	firmamentClient.EXPECT().NodeUpdated(gomock.Any(), gomock.Any()).Return(
		&firmament.NodeUpdatedResponse{Type: firmament.NodeReplyType_NODE_UPDATED_OK}, nil)

	firmament.NodeUpdated(firmamentClient, nil)
}

func Test_NodeRemoved(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	firmamentClient := mock_firmament.NewMockFirmamentSchedulerClient(mockCtrl)
	firmamentClient.EXPECT().NodeRemoved(gomock.Any(), gomock.Any()).Return(
		&firmament.NodeRemovedResponse{Type: firmament.NodeReplyType_NODE_REMOVED_OK}, nil)

	firmament.NodeRemoved(firmamentClient, nil)
}

func Test_NodeFailed(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	firmamentClient := mock_firmament.NewMockFirmamentSchedulerClient(mockCtrl)
	firmamentClient.EXPECT().NodeFailed(gomock.Any(), gomock.Any()).Return(
		&firmament.NodeFailedResponse{Type: firmament.NodeReplyType_NODE_FAILED_OK}, nil)
	firmament.NodeFailed(firmamentClient, nil)
}

func Test_NodeAdded(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	firmamentClient := mock_firmament.NewMockFirmamentSchedulerClient(mockCtrl)
	firmamentClient.EXPECT().NodeAdded(gomock.Any(), gomock.Any()).Return(
		&firmament.NodeAddedResponse{Type: firmament.NodeReplyType_NODE_ADDED_OK}, nil)
	firmament.NodeAdded(firmamentClient, nil)
}

func Test_TaskUpdated(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	firmamentClient := mock_firmament.NewMockFirmamentSchedulerClient(mockCtrl)
	firmamentClient.EXPECT().TaskUpdated(gomock.Any(), gomock.Any()).Return(
		&firmament.TaskUpdatedResponse{Type: firmament.TaskReplyType_TASK_UPDATED_OK}, nil)
	firmament.TaskUpdated(firmamentClient, nil)
}

func Test_TaskSubmitted(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	firmamentClient := mock_firmament.NewMockFirmamentSchedulerClient(mockCtrl)
	firmamentClient.EXPECT().TaskSubmitted(gomock.Any(), gomock.Any()).Return(
		&firmament.TaskSubmittedResponse{Type: firmament.TaskReplyType_TASK_SUBMITTED_OK}, nil)
	firmament.TaskSubmitted(firmamentClient, nil)
}

func Test_TaskRemoved(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	firmamentClient := mock_firmament.NewMockFirmamentSchedulerClient(mockCtrl)
	firmamentClient.EXPECT().TaskRemoved(gomock.Any(), gomock.Any()).Return(
		&firmament.TaskRemovedResponse{Type: firmament.TaskReplyType_TASK_REMOVED_OK}, nil)
	firmament.TaskRemoved(firmamentClient, nil)
}

func Test_TaskFailed(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	firmamentClient := mock_firmament.NewMockFirmamentSchedulerClient(mockCtrl)
	firmamentClient.EXPECT().TaskFailed(gomock.Any(), gomock.Any()).Return(
		&firmament.TaskFailedResponse{Type: firmament.TaskReplyType_TASK_FAILED_OK}, nil)
	firmament.TaskFailed(firmamentClient, nil)
}

func Test_TaskCompleted(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	firmamentClient := mock_firmament.NewMockFirmamentSchedulerClient(mockCtrl)
	firmamentClient.EXPECT().TaskCompleted(gomock.Any(), gomock.Any()).Return(
		&firmament.TaskCompletedResponse{Type: firmament.TaskReplyType_TASK_COMPLETED_OK}, nil)
	firmament.TaskCompleted(firmamentClient, nil)
}

func Test_Schedule(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	firmamentClient := mock_firmament.NewMockFirmamentSchedulerClient(mockCtrl)
	firmamentClient.EXPECT().Schedule(gomock.Any(), gomock.Any()).Return(
		&firmament.SchedulingDeltas{}, nil)
	firmament.Schedule(firmamentClient)
}
