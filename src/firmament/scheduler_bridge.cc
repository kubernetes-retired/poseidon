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

namespace poseidon {

SchedulerBridge::SchedulerBridge() {
  job_map_.reset(new JobMap_t);
  task_map_.reset(new TaskMap_t);
  resource_map_.reset(new ResourceMap_t);
  knowledge_base_.reset(new KnowledgeBase);
  topology_manager_.reset(new TopologyManager);
  ResourceStatus* top_level_res_status = CreateTopLevelResource();
  top_level_res_id_ =
      firmament::ResourceIDFromString(
          top_level_res_status->descriptor().uuid());
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
  ResourceID_t rid = firmament::ResourceIDFromString(node_id);
  CHECK(ContainsKey(*resource_map_, rid));
  kb_populator_->PopulateNodeStats(firmament::to_string(rid), node_stats);
}

JobDescriptor* SchedulerBridge::CreateJobForPod(const string& pod) {
  // Fake out a job for this pod
  // XXX(malte): we should equate a Firmament "job" with a K8s
  // "deployment" and "job" for a more sane notion here.
  JobID_t job_id = firmament::GenerateJobID();
  JobDescriptor new_jd;
  CHECK(InsertIfNotPresent(job_map_.get(), job_id, new_jd));
  JobDescriptor* jd = FindOrNull(*job_map_, job_id);
  jd->set_uuid(firmament::to_string(job_id));
  jd->set_name(pod);
  jd->set_state(JobDescriptor::CREATED);
  TaskDescriptor* root_td = jd->mutable_root_task();
  root_td->set_uid(firmament::GenerateRootTaskID(*jd));
  root_td->set_name(pod);
  root_td->set_state(TaskDescriptor::CREATED);
  root_td->set_job_id(jd->uuid());
  CHECK(InsertIfNotPresent(task_map_.get(), root_td->uid(), root_td));
  return jd;
}

bool SchedulerBridge::CreateResourceForNode(const string& node_id,
                                            const string& node_name) {
  ResourceID_t rid = firmament::ResourceIDFromString(node_id);
  // Check if we know about this node already
  if (!ContainsKey(*resource_map_, rid)) {
    LOG(INFO) << "Adding new node's resource with RID " << rid;
    CHECK(InsertIfNotPresent(&node_map_, rid, node_name));
    // Create a new Firmament resource
    ResourceTopologyNodeDescriptor* rtnd_ptr =
      new ResourceTopologyNodeDescriptor();
    // Create and initialize RD
    ResourceDescriptor* rd_ptr = rtnd_ptr->mutable_resource_desc();
    rd_ptr->set_uuid(firmament::to_string(rid));
    rd_ptr->set_type(ResourceDescriptor::RESOURCE_PU);
    rd_ptr->set_state(ResourceDescriptor::RESOURCE_IDLE);
    rtnd_ptr->set_parent_id(firmament::to_string(top_level_res_id_));
    // Need to maintain a ResourceStatus for the resource map
    // TODO(malte): set hostname correctly
    ResourceStatus* rs = new ResourceStatus(rd_ptr, rtnd_ptr, "", 0);
    // Insert into resource map
    CHECK(InsertIfNotPresent(resource_map_.get(), rid, rs));
    // Register with the scheudler
    // TODO(malte): we use a hack here -- we pass simulated=true to
    // avoid Firmament instantiating an actual executor for this resource.
    // Instead, we rely on the no-op SimulatedExecutor. We should change
    // it such that Firmament does not mandatorily create an executor.
    flow_scheduler_->RegisterResource(rs->mutable_topology_node(), false, true);
    return true;
  }
  return false;
}

ResourceStatus* SchedulerBridge::CreateTopLevelResource() {
  ResourceID_t res_id = firmament::GenerateResourceID();
  ResourceTopologyNodeDescriptor* rtnd_ptr =
    new ResourceTopologyNodeDescriptor();
  // Set up the RD
  ResourceDescriptor* rd_ptr = rtnd_ptr->mutable_resource_desc();
  rd_ptr->set_uuid(firmament::to_string(res_id));
  rd_ptr->set_type(ResourceDescriptor::RESOURCE_COORDINATOR);
  // Need to maintain a ResourceStatus for the resource map
  // TODO(malte): don't pass localhost here
  ResourceStatus* rs_ptr = new ResourceStatus(rd_ptr, rtnd_ptr, "localhost", 0);
  // Insert into resource map
  CHECK(InsertIfNotPresent(resource_map_.get(), res_id, rs_ptr));
  return rs_ptr;
}

unordered_map<string, string>* SchedulerBridge::RunScheduler(
    const vector<PodStatistics>& pods) {
  bool found_new_pod = false;
  for (const PodStatistics& pod : pods) {
    if (pod.state_ == "Pending") {
      if (firmament::FindOrNull(pod_to_task_map_, pod.name_) == NULL) {
        LOG(INFO) << "New unscheduled pod: " << pod.name_;
        found_new_pod = true;
        JobDescriptor* jd_ptr = CreateJobForPod(pod.name_);
        CHECK(InsertIfNotPresent(&pod_to_task_map_, pod.name_,
                                 jd_ptr->root_task().uid()));
        CHECK(InsertIfNotPresent(&task_to_pod_map_,
                                 jd_ptr->root_task().uid(), pod.name_));
        flow_scheduler_->AddJob(jd_ptr);
      }
    } else if (pod.state_ == "Running") {
      // TODO(ionel): Update pod statistics.
      string* node = FindOrNull(pod_to_node_map_, pod.name_);
      CHECK_NOTNULL(node);
      TaskID_t* tid_ptr = FindOrNull(pod_to_task_map_, pod.name_);
      CHECK_NOTNULL(tid_ptr);
      kb_populator_->PopulatePodStats(*tid_ptr, *node, pod);
    } else if (pod.state_ == "Succeeded") {
      // TODO(ionel): Generate TaskFinalReport if were detecting
      // for the first time that the pod has succeeded.
      // kb_populator_->ProcessFinalPodReport();
    } else if (pod.state_ == "Failed" ||
               pod.state_ == "Unknown") {
      // We don't have to do anything in these cases.
    } else {
      LOG(ERROR) << "Pod " << pod.name_ << " is unexpected state "
                 << pod.state_;
    }
  }
  unordered_map<string, string>* pod_node_bindings =
    new unordered_map<string, string>();
  if (!found_new_pod) {
    // Do not run the scheduler if there are no new pods.
    return pod_node_bindings;
  }
  // Invoke Firmament scheduling
  firmament::scheduler::SchedulerStats sstat;
  vector<firmament::SchedulingDelta> deltas;
  flow_scheduler_->ScheduleAllJobs(&sstat, &deltas);

  // Extract results
  LOG(INFO) << "Got " << deltas.size() << " scheduling deltas";
  for (auto& d : deltas) {
    LOG(INFO) << "Delta: " << d.DebugString();
    if (d.type() == firmament::SchedulingDelta::PLACE) {
      const string* pod = FindOrNull(task_to_pod_map_, d.task_id());
      const string* node = FindOrNull(
          node_map_, firmament::ResourceIDFromString(d.resource_id()));
      CHECK_NOTNULL(pod);
      CHECK_NOTNULL(node);
      CHECK(InsertIfNotPresent(&pod_to_node_map_, *pod, *node));
      CHECK(InsertIfNotPresent(pod_node_bindings, *pod, *node));
    } else {
      LOG(WARNING) << "Encountered unsupported scheduling delta of type "
                   << firmament::to_string(d.type());
    }
  }
  return pod_node_bindings;
}

}  // namespace poseidon
