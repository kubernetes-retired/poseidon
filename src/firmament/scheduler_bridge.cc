/*
 * Poseidon
 * Copyright (c) The Poseidon Authors.
 * All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * THIS CODE IS PROVIDED ON AN *AS IS* BASIS, WITHOUT WARRANTIES OR
 * CONDITIONS OF ANY KIND, EITHER EXPRESS OR IMPLIED, INCLUDING WITHOUT
 * LIMITATION ANY IMPLIED WARRANTIES OR CONDITIONS OF TITLE, FITNESS FOR
 * A PARTICULAR PURPOSE, MERCHANTABLITY OR NON-INFRINGEMENT.
 *
 * See the Apache Version 2.0 License for specific language governing
 * permissions and limitations under the License.
 */

#include "firmament/scheduler_bridge.h"

#include "misc/pb_utils.h"

using firmament::GenerateJobID;
using firmament::GenerateResourceID;
using firmament::JobIDFromString;
using firmament::KB_TO_MB;
using firmament::ResourceIDFromString;
using firmament::scheduler::SchedulerStats;
using firmament::SchedulingDelta;
using firmament::to_string;

namespace poseidon {

SchedulerBridge::SchedulerBridge() {
  job_map_.reset(new JobMap_t);
  task_map_.reset(new TaskMap_t);
  resource_map_.reset(new ResourceMap_t);
  knowledge_base_.reset(new KnowledgeBase);
  topology_manager_.reset(new TopologyManager);
  ResourceStatus* top_level_res_status = CreateTopLevelResource();
  top_level_res_id_ =
      ResourceIDFromString(top_level_res_status->descriptor().uuid());
  sim_messaging_adapter_ = new SimulatedMessagingAdapter<BaseMessage>();
  trace_generator_ = new TraceGenerator(&wall_time_);
  flow_scheduler_ =
    new FlowScheduler(job_map_, resource_map_,
                      top_level_res_status->mutable_topology_node(), obj_store_,
                      task_map_, knowledge_base_, topology_manager_,
                      sim_messaging_adapter_, NULL, top_level_res_id_, "",
                      &wall_time_, trace_generator_);
  kb_populator_ = new KnowledgeBasePopulator(knowledge_base_);
  LOG(INFO) << "Firmament scheduler instantiated: " << flow_scheduler_;
}

SchedulerBridge::~SchedulerBridge() {
  delete flow_scheduler_;
  delete sim_messaging_adapter_;
  delete trace_generator_;
  delete kb_populator_;
}

void SchedulerBridge::AddStatisticsForNode(const string& node_id,
                                           const NodeStatistics& node_stats) {
  ResourceID_t rid = ResourceIDFromString(node_id);
  CHECK(ContainsKey(*resource_map_, rid));
  kb_populator_->PopulateNodeStats(to_string(rid), node_stats);
}

JobDescriptor* SchedulerBridge::CreateJobForPod(const string& pod) {
  // Fake out a job for this pod
  // XXX(malte): we should equate a Firmament "job" with a K8s
  // "deployment" and "job" for a more sane notion here.
  JobID_t job_id = GenerateJobID();
  JobDescriptor new_jd;
  CHECK(InsertIfNotPresent(job_map_.get(), job_id, new_jd));
  JobDescriptor* jd = FindOrNull(*job_map_, job_id);
  jd->set_uuid(to_string(job_id));
  jd->set_name(pod);
  jd->set_state(JobDescriptor::CREATED);
  TaskDescriptor* root_td = jd->mutable_root_task();
  root_td->set_uid(GenerateRootTaskID(*jd));
  root_td->set_name(pod);
  root_td->set_state(TaskDescriptor::CREATED);
  root_td->set_job_id(jd->uuid());
  root_td->set_start_time(wall_time_.GetCurrentTimestamp());
  CHECK(InsertIfNotPresent(task_map_.get(), root_td->uid(), root_td));
  return jd;
}

void SchedulerBridge::CreateResourceTopologyForNode(
    const ResourceID_t& rid,
    const apiclient::NodeStatistics& node_stats) {
  LOG(INFO) << "Adding new node's resource with RID " << rid;
  CHECK(InsertIfNotPresent(&resource_to_node_map_, rid,
                           node_stats.hostname_));
  ResourceStatus* root_rs_ptr =
    FindPtrOrNull(*resource_map_, top_level_res_id_);
  CHECK_NOTNULL(root_rs_ptr);
  // Create a new Firmament resource
  ResourceTopologyNodeDescriptor* rtnd_ptr =
    root_rs_ptr->mutable_topology_node()->add_children();
  // Create and initialize machine RD
  ResourceDescriptor* rd_ptr = rtnd_ptr->mutable_resource_desc();
  rd_ptr->set_uuid(to_string(rid));
  rd_ptr->set_type(ResourceDescriptor::RESOURCE_MACHINE);
  rd_ptr->set_state(ResourceDescriptor::RESOURCE_IDLE);
  rd_ptr->set_friendly_name(node_stats.hostname_);
  ResourceVector* res_cap = rd_ptr->mutable_resource_capacity();
  res_cap->set_ram_cap(node_stats.memory_capacity_kb_ / KB_TO_MB);
  res_cap->set_cpu_cores(node_stats.cpu_capacity_);
  rtnd_ptr->set_parent_id(to_string(top_level_res_id_));
  // Connect PU RDs to the machine RD.
  // TODO(ionel): In the future, we want to get real node topology rather
  // than manually connecting PU RDs to the machine RD.
  for (uint32_t num_pu = 0; num_pu < node_stats.cpu_capacity_; num_pu++) {
    ResourceID_t pu_rid = GenerateResourceID();
    string pu_name = node_stats.hostname_ + "_pu" + to_string(num_pu);
    LOG(INFO) << "Adding new PU with RID " << pu_name << " " << pu_rid;
    ResourceTopologyNodeDescriptor* pu_rtnd_ptr = rtnd_ptr->add_children();
    ResourceDescriptor* pu_rd_ptr = pu_rtnd_ptr->mutable_resource_desc();
    pu_rd_ptr->set_uuid(to_string(pu_rid));
    pu_rd_ptr->set_type(ResourceDescriptor::RESOURCE_PU);
    pu_rd_ptr->set_state(ResourceDescriptor::RESOURCE_IDLE);
    pu_rd_ptr->set_friendly_name(pu_name);
    pu_rtnd_ptr->set_parent_id(to_string(rid));
    CHECK(InsertIfNotPresent(&pu_to_node_map_, pu_rid, rid));
    ResourceStatus* pu_rs = new ResourceStatus(pu_rd_ptr, pu_rtnd_ptr, "", 0);
    CHECK(InsertIfNotPresent(resource_map_.get(), pu_rid, pu_rs));
  }
  // TODO(malte): set hostname correctly
  ResourceStatus* rs = new ResourceStatus(rd_ptr, rtnd_ptr, "", 0);
  // Insert into resource map
  CHECK(InsertIfNotPresent(resource_map_.get(), rid, rs));
  // Register with the scheduler
  // TODO(malte): we use a hack here -- we pass simulated=true to
  // avoid Firmament instantiating an actual executor for this resource.
  // Instead, we rely on the no-op SimulatedExecutor. We should change
  // it such that Firmament does not mandatorily create an executor.
  flow_scheduler_->RegisterResource(rs->mutable_topology_node(), false, true);
}

ResourceStatus* SchedulerBridge::CreateTopLevelResource() {
  ResourceID_t res_id = GenerateResourceID();
  ResourceTopologyNodeDescriptor* rtnd_ptr =
    new ResourceTopologyNodeDescriptor();
  // Set up the RD
  ResourceDescriptor* rd_ptr = rtnd_ptr->mutable_resource_desc();
  rd_ptr->set_uuid(to_string(res_id));
  rd_ptr->set_type(ResourceDescriptor::RESOURCE_COORDINATOR);
  // Need to maintain a ResourceStatus for the resource map
  // TODO(malte): don't pass localhost here
  ResourceStatus* rs_ptr = new ResourceStatus(rd_ptr, rtnd_ptr, "localhost", 0);
  // Insert into resource map
  CHECK(InsertIfNotPresent(resource_map_.get(), res_id, rs_ptr));
  return rs_ptr;
}

bool SchedulerBridge::NodeAdded(const string& node_id,
                                const apiclient::NodeStatistics& node_stats) {
  ResourceID_t rid = ResourceIDFromString(node_id);
  // Check if we know about this node already
  if (!ContainsKey(*resource_map_, rid)) {
    CreateResourceTopologyForNode(rid, node_stats);
    return true;
  }
  return false;
}

void SchedulerBridge::NodeFailed(const string& node_id,
                                 const apiclient::NodeStatistics& node_stats) {
  ResourceID_t rid = ResourceIDFromString(node_id);
  ResourceStatus* rs_ptr = FindPtrOrNull(*resource_map_, rid);
  if (rs_ptr == NULL) {
    // We haven't yet seen this node in ready state.
    return;
  }
  LOG(INFO) << "Node " << node_stats.hostname_ << " failed";
  resource_to_node_map_.erase(rid);
  ResourceTopologyNodeDescriptor* rtnd_ptr = rs_ptr->mutable_topology_node();
  DFSTraversePostOrderResourceProtobufTreeReturnRTND(
      rtnd_ptr,
      boost::bind(&SchedulerBridge::CleanPUStateForDeregisteredResource,
                  this, _1));
  flow_scheduler_->DeregisterResource(rtnd_ptr);
}

void SchedulerBridge::CleanPUStateForDeregisteredResource(
    ResourceTopologyNodeDescriptor* rtnd_ptr) {
  const ResourceDescriptor& rd = rtnd_ptr->resource_desc();
  if (rd.type() == ResourceDescriptor::RESOURCE_PU) {
    ResourceID_t res_id = ResourceIDFromString(rd.uuid());
    pu_to_node_map_.erase(res_id);
  }
}

unordered_map<string, string>* SchedulerBridge::RunScheduler(
    const vector<PodStatistics>& pods) {
  bool found_pending_pod = false;
  for (const PodStatistics& pod : pods) {
    if (pod.state_ == "Pending") {
      found_pending_pod = true;
      if (pending_pods_.find(pod.name_) == pending_pods_.end()) {
        // First time we detect the pod in pending state.
        pending_pods_.insert(pod.name_);
        // Erase the pod from the other sets in case it previously was in a
        // failed or unknown state.
        failed_pods_.erase(pod.name_);
        unknown_pods_.erase(pod.name_);
        // Check if this is the truly first time we see it.
        if (FindOrNull(pod_to_task_map_, pod.name_) == NULL) {
          LOG(INFO) << "New unscheduled pod: " << pod.name_;
          JobDescriptor* jd_ptr = CreateJobForPod(pod.name_);
          CHECK(InsertIfNotPresent(&pod_to_task_map_, pod.name_,
                                   jd_ptr->root_task().uid()));
          CHECK(InsertIfNotPresent(&task_to_pod_map_,
                                   jd_ptr->root_task().uid(), pod.name_));
          CHECK(InsertIfNotPresent(&job_num_incomplete_tasks_,
                                   JobIDFromString(jd_ptr->uuid()), 1));
          flow_scheduler_->AddJob(jd_ptr);
        }
      }
    } else if (pod.state_ == "Running") {
      // TODO(ionel): Update pod statistics.
      if (running_pods_.find(pod.name_) == running_pods_.end()) {
        // First time we detect the pod in running state.
        running_pods_.insert(pod.name_);
        // Delete the pod from the collection corresponding to its prior state.
        uint32_t num_pending_erased = pending_pods_.erase(pod.name_);
        uint32_t num_failed_erased = failed_pods_.erase(pod.name_);
        uint32_t num_unknown_erased = unknown_pods_.erase(pod.name_);
        // The pod should have either been in pending, failed or unknown state.
        CHECK_EQ(num_pending_erased + num_failed_erased + num_unknown_erased,
                 1);
      }
      string* node = FindOrNull(pod_to_node_map_, pod.name_);
      CHECK_NOTNULL(node);
      TaskID_t* tid_ptr = FindOrNull(pod_to_task_map_, pod.name_);
      CHECK_NOTNULL(tid_ptr);
      kb_populator_->PopulatePodStats(*tid_ptr, *node, pod);
    } else if (pod.state_ == "Succeeded") {
      if (succeeded_pods_.find(pod.name_) == succeeded_pods_.end()) {
        // First time we detect the pod in succeeded state.
        succeeded_pods_.insert(pod.name_);
        // Delete the pod from the running pods or pending set. The pod
        // can be in the pending set if it got scheduled in the prior
        // scheduler run and completed before this run.
        uint32_t num_pending_erased = pending_pods_.erase(pod.name_);
        uint32_t num_running_erased = running_pods_.erase(pod.name_);
        CHECK_EQ(num_pending_erased + num_running_erased, 1);
        // Inform the scheduler that the pod has completed.
        TaskID_t* tid_ptr = FindOrNull(pod_to_task_map_, pod.name_);
        CHECK_NOTNULL(tid_ptr);
        TaskDescriptor* td_ptr = FindPtrOrNull(*task_map_, *tid_ptr);
        CHECK_NOTNULL(td_ptr);
        td_ptr->set_finish_time(wall_time_.GetCurrentTimestamp());
        TaskFinalReport report;
        flow_scheduler_->HandleTaskCompletion(td_ptr, &report);
        kb_populator_->PopulateTaskFinalReport(*td_ptr, &report);
        flow_scheduler_->HandleTaskFinalReport(report, td_ptr);
        // Delete the binding between the pod and the K8s node.
        pod_to_node_map_.erase(pod.name_);
        // Delete the both mappings between pod and task.
        task_to_pod_map_.erase(*tid_ptr);
        pod_to_task_map_.erase(pod.name_);
        JobID_t job_id = JobIDFromString(td_ptr->job_id());
        JobDescriptor* jd_ptr = FindOrNull(*job_map_, job_id);
        // Don't remove the root task so that tasks can still be appended to
        // the job. We only remove the root task when the job completes.
        if (td_ptr != jd_ptr->mutable_root_task()) {
          task_map_->erase(td_ptr->uid());
        }
        // Check if it was the last task of the job.
        uint64_t* num_incomplete_tasks =
          FindOrNull(job_num_incomplete_tasks_, job_id);
        CHECK_NOTNULL(num_incomplete_tasks);
        if (*num_incomplete_tasks == 0) {
          flow_scheduler_->HandleJobCompletion(job_id);
          job_num_incomplete_tasks_.erase(job_id);
          task_map_->erase(jd_ptr->root_task().uid());
          job_map_->erase(job_id);
        }
      }
    } else if (pod.state_ == "Failed") {
      if (failed_pods_.find(pod.name_) == failed_pods_.end()) {
        // First time we detect the pod in failed state.
        failed_pods_.insert(pod.name_);
        // Delete the pod from the running pods or pending set. The pod
        // can be in the pending set if it got scheduled in the prior
        // scheduler run and failed before this run.
        uint32_t num_running_erased = running_pods_.erase(pod.name_);
        uint32_t num_pending_erased = pending_pods_.erase(pod.name_);
        CHECK_EQ(num_pending_erased + num_running_erased, 1);
        // Inform the scheduler that the the pod has failed.
        TaskID_t* tid_ptr = FindOrNull(pod_to_task_map_, pod.name_);
        CHECK_NOTNULL(tid_ptr);
        TaskDescriptor* td_ptr = FindPtrOrNull(*task_map_, *tid_ptr);
        CHECK_NOTNULL(td_ptr);
        flow_scheduler_->HandleTaskFailure(td_ptr);
        // Delete the binding between the pod and the K8s node.
        pod_to_node_map_.erase(pod.name_);
        // We do not delete the mappings between pod and task in case the
        // task may be rescheduled.
      }
    } else if (pod.state_ == "Unknown") {
      // TODO(ionel): Handle this case.
    } else {
      LOG(ERROR) << "Pod " << pod.name_ << " is unexpected state "
                 << pod.state_;
    }
  }
  unordered_map<string, string>* pod_node_bindings =
    new unordered_map<string, string>();
  if (!found_pending_pod) {
    // Do not run the scheduler if there are no pending pods.
    return pod_node_bindings;
  }
  // Invoke Firmament scheduling
  SchedulerStats sstat;
  vector<SchedulingDelta> deltas;
  flow_scheduler_->ScheduleAllJobs(&sstat, &deltas);

  // Extract results
  LOG(INFO) << "Got " << deltas.size() << " scheduling deltas";
  for (auto& d : deltas) {
    LOG(INFO) << "Delta: " << d.DebugString();
    if (d.type() == SchedulingDelta::PLACE) {
      const string* pod = FindOrNull(task_to_pod_map_, d.task_id());
      const ResourceID_t* node_rid =
        FindOrNull(pu_to_node_map_, ResourceIDFromString(d.resource_id()));
      CHECK_NOTNULL(node_rid);
      const string* node_name = FindOrNull(resource_to_node_map_, *node_rid);
      CHECK_NOTNULL(pod);
      CHECK_NOTNULL(node_name);
      CHECK(InsertIfNotPresent(&pod_to_node_map_, *pod, *node_name));
      CHECK(InsertIfNotPresent(pod_node_bindings, *pod, *node_name));
    } else {
      // TODO(ionel): Handle SchedulingDelta::NOOP, PREEMPT, MIGRATE.
      LOG(WARNING) << "Encountered unsupported scheduling delta of type "
                   << to_string(d.type());
    }
  }
  return pod_node_bindings;
}

}  // namespace poseidon
