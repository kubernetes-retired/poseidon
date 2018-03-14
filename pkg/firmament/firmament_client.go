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
	"fmt"

	"github.com/golang/glog"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

func Schedule(client FirmamentSchedulerClient) *SchedulingDeltas {
	scheduleResp, err := client.Schedule(context.Background(), &ScheduleRequest{})
	if err != nil {
		grpclog.Fatalf("%v.Schedule(_) = _, %v: ", client, err)
	}
	return scheduleResp
}

func TaskCompleted(client FirmamentSchedulerClient, tuid *TaskUID) {
	tCompletedResp, err := client.TaskCompleted(context.Background(), tuid)
	if err != nil {
		grpclog.Fatalf("%v.TaskCompleted(_) = _, %v: ", client, err)
	}
	switch tCompletedResp.Type {
	case TaskReplyType_TASK_NOT_FOUND:
		glog.Fatalf("Task %d not found", tuid.TaskUid)
	case TaskReplyType_TASK_JOB_NOT_FOUND:
		glog.Fatalf("Task's %d job not found", tuid.TaskUid)
	case TaskReplyType_TASK_COMPLETED_OK:
	default:
		panic(fmt.Sprintf("Unexpected TaskCompleted response %v for task %v", tCompletedResp, tuid.TaskUid))
	}
}

func TaskFailed(client FirmamentSchedulerClient, tuid *TaskUID) {
	tFailedResp, err := client.TaskFailed(context.Background(), tuid)
	if err != nil {
		grpclog.Fatalf("%v.TaskFailed(_) = _, %v: ", client, err)
	}
	switch tFailedResp.Type {
	case TaskReplyType_TASK_NOT_FOUND:
		glog.Fatalf("Task %d not found", tuid.TaskUid)
	case TaskReplyType_TASK_JOB_NOT_FOUND:
		glog.Fatalf("Task's %d job not found", tuid.TaskUid)
	case TaskReplyType_TASK_FAILED_OK:
	default:
		panic(fmt.Sprintf("Unexpected TaskFailed response %v for task %v", tFailedResp, tuid.TaskUid))
	}
}

func TaskRemoved(client FirmamentSchedulerClient, tuid *TaskUID) {
	tRemovedResp, err := client.TaskRemoved(context.Background(), tuid)
	if err != nil {
		grpclog.Fatalf("%v.TaskRemoved(_) = _, %v: ", client, err)
	}
	switch tRemovedResp.Type {
	case TaskReplyType_TASK_NOT_FOUND:
		glog.Fatalf("Task %d not found", tuid.TaskUid)
	case TaskReplyType_TASK_JOB_NOT_FOUND:
		glog.Fatalf("Task's %d job not found", tuid.TaskUid)
	case TaskReplyType_TASK_REMOVED_OK:
	default:
		panic(fmt.Sprintf("Unexpected TaskRemoved response %v for task %v", tRemovedResp, tuid.TaskUid))
	}
}

func TaskSubmitted(client FirmamentSchedulerClient, td *TaskDescription) {
	tSubmittedResp, err := client.TaskSubmitted(context.Background(), td)
	if err != nil {
		grpclog.Fatalf("%v.TaskSubmitted(_) = _, %v: ", client, err)
	}
	switch tSubmittedResp.Type {
	case TaskReplyType_TASK_ALREADY_SUBMITTED:
		glog.Fatalf("Task (%s,%d) already submitted", td.JobDescriptor.Uuid, td.TaskDescriptor.Uid)
	case TaskReplyType_TASK_STATE_NOT_CREATED:
		glog.Fatalf("Task (%s,%d) not in created state", td.JobDescriptor.Uuid, td.TaskDescriptor.Uid)
	case TaskReplyType_TASK_SUBMITTED_OK:
	default:
		panic(fmt.Sprintf("Unexpected TaskSubmitted response %v for task (%v,%v)", tSubmittedResp, td.JobDescriptor.Uuid, td.TaskDescriptor.Uid))
	}
}

func TaskUpdated(client FirmamentSchedulerClient, td *TaskDescription) {
	tUpdatedResp, err := client.TaskUpdated(context.Background(), td)
	if err != nil {
		grpclog.Fatalf("%v.TaskUpdated(_) = _, %v: ", client, err)
	}
	switch tUpdatedResp.Type {
	case TaskReplyType_TASK_NOT_FOUND:
		glog.Fatalf("Task (%s,%d) not found", td.JobDescriptor.Uuid, td.TaskDescriptor.Uid)
	case TaskReplyType_TASK_JOB_NOT_FOUND:
		glog.Fatalf("Task's (%s,%d) job not found", td.JobDescriptor.Uuid, td.TaskDescriptor.Uid)
	case TaskReplyType_TASK_UPDATED_OK:
	default:
		panic(fmt.Sprintf("Unexpected TaskUpdated response %v for task (%v,%v)", tUpdatedResp, td.JobDescriptor.Uuid, td.TaskDescriptor.Uid))
	}
}

func NodeAdded(client FirmamentSchedulerClient, rtnd *ResourceTopologyNodeDescriptor) {
	nAddedResp, err := client.NodeAdded(context.Background(), rtnd)
	if err != nil {
		grpclog.Fatalf("%v.NodeAdded(_) = _, %v: ", client, err)
	}
	switch nAddedResp.Type {
	case NodeReplyType_NODE_ALREADY_EXISTS:
		glog.Fatalf("Tried to add existing node %s", rtnd.ResourceDesc.Uuid)
	case NodeReplyType_NODE_ADDED_OK:
	default:
		panic(fmt.Sprintf("Unexpected NodeAdded response %v for node %v", nAddedResp, rtnd.ResourceDesc.Uuid))
	}
}

func NodeFailed(client FirmamentSchedulerClient, ruid *ResourceUID) {
	nFailedResp, err := client.NodeFailed(context.Background(), ruid)
	if err != nil {
		grpclog.Fatalf("%v.NodeFailed(_) = _, %v: ", client, err)
	}
	switch nFailedResp.Type {
	case NodeReplyType_NODE_NOT_FOUND:
		glog.Fatalf("Tried to fail non-existing node %s", ruid.ResourceUid)
	case NodeReplyType_NODE_FAILED_OK:
	default:
		panic(fmt.Sprintf("Unexpected NodeFailed response %v for node %v", nFailedResp, ruid.ResourceUid))
	}
}

func NodeRemoved(client FirmamentSchedulerClient, ruid *ResourceUID) {
	nRemovedResp, err := client.NodeRemoved(context.Background(), ruid)
	if err != nil {
		grpclog.Fatalf("%v.NodeRemoved(_) = _, %v: ", client, err)
	}
	switch nRemovedResp.Type {
	case NodeReplyType_NODE_NOT_FOUND:
		glog.Fatalf("Tried to remove non-existing node %s", ruid.ResourceUid)
	case NodeReplyType_NODE_REMOVED_OK:
	default:
		panic(fmt.Sprintf("Unexpected NodeRemoved response %v for node %v", nRemovedResp, ruid.ResourceUid))
	}
}

func NodeUpdated(client FirmamentSchedulerClient, rtnd *ResourceTopologyNodeDescriptor) {
	nUpdatedResp, err := client.NodeUpdated(context.Background(), rtnd)
	if err != nil {
		grpclog.Fatalf("%v.NodeUpdated(_) = _, %v: ", client, err)
	}
	switch nUpdatedResp.Type {
	case NodeReplyType_NODE_NOT_FOUND:
		glog.Fatalf("Tried to updated non-existing node %s", rtnd.ResourceDesc.Uuid)
	case NodeReplyType_NODE_UPDATED_OK:
	default:
		panic(fmt.Sprintf("Unexpected NodeUpdated response %v for node %v", nUpdatedResp, rtnd.ResourceDesc.Uuid))
	}
}

func AddTaskStats(client FirmamentSchedulerClient, ts *TaskStats) {
	_, err := client.AddTaskStats(context.Background(), ts)
	if err != nil {
		grpclog.Fatalf("%v.AddTaskStats(_) = _, %v: ", client, err)
	}
}

func AddNodeStats(client FirmamentSchedulerClient, rs *ResourceStats) {
	_, err := client.AddNodeStats(context.Background(), rs)
	if err != nil {
		grpclog.Fatalf("%v.AddNodeStats(_) = _, %v: ", client, err)
	}
}

func New(address string) (FirmamentSchedulerClient, *grpc.ClientConn, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(address, opts...)
	if err != nil {
		glog.Errorf("Did not connect to Firmament scheduler: %v", err)
		return nil, nil, err
	}
	fc := NewFirmamentSchedulerClient(conn)
	return fc, conn, nil
}
