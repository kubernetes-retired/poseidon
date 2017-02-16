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

#ifndef POSEIDON_FIRMAMENT_SCHEDULER_BRIDGE_H
#define POSEIDON_FIRMAMENT_SCHEDULER_BRIDGE_H

#include <unordered_map>
#include <unordered_set>

#include "base/resource_status.h"
#include "base/resource_topology_node_desc.pb.h"
#include "base/units.h"
#include "misc/map-util.h"
#include "misc/trace_generator.h"
#include "misc/utils.h"
#include "misc/wall_time.h"
#include "platforms/sim/simulated_messaging_adapter.h"
#include "scheduling/flow/flow_scheduler.h"
#include "scheduling/scheduling_delta.pb.h"
#include "storage/simple_object_store.h"

#include "apiclient/utils.h"
#include "firmament/knowledge_base_populator.h"

using firmament::BaseMessage;
using firmament::ContainsKey;
using firmament::FindOrNull;
using firmament::InsertIfNotPresent;
using firmament::JobID_t;
using firmament::JobDescriptor;
using firmament::JobMap_t;
using firmament::ResourceDescriptor;
using firmament::ResourceMap_t;
using firmament::ResourceID_t;
using firmament::ResourceStatus;
using firmament::ResourceVector;
using firmament::TaskID_t;
using firmament::TaskDescriptor;
using firmament::TaskMap_t;
using firmament::KnowledgeBase;
using firmament::ResourceTopologyNodeDescriptor;
using firmament::TraceGenerator;
using firmament::WallTime;
using firmament::scheduler::FlowScheduler;
using firmament::scheduler::ObjectStoreInterface;
using firmament::scheduler::TopologyManager;
using firmament::platform::sim::SimulatedMessagingAdapter;
using poseidon::apiclient::NodeStatistics;
using poseidon::apiclient::PodStatistics;

using namespace std;

namespace poseidon {

class SchedulerBridge {
 public:
  SchedulerBridge();
  ~SchedulerBridge();
  void AddStatisticsForNode(const string& node_id,
                            const NodeStatistics& node_stats);
  JobDescriptor* CreateJobForPod(const string& pod);
  bool CreateResourceTopologyForNode(const string& node_id,
                                     const apiclient::NodeStatistics& node_stats);
  unordered_map<string, string>* RunScheduler(
      const vector<PodStatistics>& pods);

 private:
  ResourceStatus* CreateTopLevelResource();

  SimulatedMessagingAdapter<BaseMessage>* sim_messaging_adapter_;
  TraceGenerator* trace_generator_;
  WallTime wall_time_;
  FlowScheduler* flow_scheduler_;
  // Data structures thare are populated by Firmament. We should never have to
  // direclty insert values in these data structures. However, we can query them.
  boost::shared_ptr<JobMap_t> job_map_;
  boost::shared_ptr<KnowledgeBase> knowledge_base_;
  boost::shared_ptr<ObjectStoreInterface> obj_store_;
  boost::shared_ptr<TaskMap_t> task_map_;
  boost::shared_ptr<TopologyManager> topology_manager_;
  // Data structures that we populate in the scheduler_bridge.
  boost::shared_ptr<ResourceMap_t> resource_map_;
  // Mapping from ResourceID_t to K8s node hostname.
  map<ResourceID_t, string> resource_to_node_map_;
  // Mapping from PU ResourceID_t string to node ResourceID_t string.
  unordered_map<string, string> pu_to_node_map_;
  // Mapping from pod name to TaskID_t.
  unordered_map<string, TaskID_t> pod_to_task_map_;
  // Mapping from pod name to K8s node hostname.
  unordered_map<string, string> pod_to_node_map_;
  // Mapping from TaskID_t to pod name.
  unordered_map<TaskID_t, string> task_to_pod_map_;
  // Set of pods in Running state.
  unordered_set<string> running_pods_;
  // Set of pods in Pending state.
  unordered_set<string> pending_pods_;
  // Set of pods in Succeeded state.
  unordered_set<string> succeeded_pods_;
  // Set of pods in Failed state.
  unordered_set<string> failed_pods_;
  // Set of pods in Unknown state.
  unordered_set<string> unknown_pods_;
  ResourceID_t top_level_res_id_;
  KnowledgeBasePopulator* kb_populator_;
};

}  // namespace poseidon

#endif  // POSEIDON_FIRMAMENT_SCHEDULER_BRIDGE_H
