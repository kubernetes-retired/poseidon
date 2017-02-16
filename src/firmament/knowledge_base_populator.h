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

#ifndef POSEIDON_FIRMAMENT_KNOWLEDGE_BASED_POPULATOR_H
#define POSEIDON_FIRMAMENT_KNOWLEDGE_BASED_POPULATOR_H

#include "base/machine_perf_statistics_sample.pb.h"
#include "base/task_final_report.pb.h"
#include "base/task_perf_statistics_sample.pb.h"
#include "misc/wall_time.h"
#include "scheduling/knowledge_base.h"

#include "apiclient/utils.h"

using firmament::KnowledgeBase;
using firmament::MachinePerfStatisticsSample;
using firmament::ResourceID_t;
using firmament::TaskDescriptor;
using firmament::TaskFinalReport;
using firmament::TaskID_t;
using firmament::TaskPerfStatisticsSample;
using firmament::WallTime;
using poseidon::apiclient::NodeStatistics;
using poseidon::apiclient::PodStatistics;

namespace poseidon {

class KnowledgeBasePopulator {
 public:
  KnowledgeBasePopulator(boost::shared_ptr<KnowledgeBase> knowledge_base);
  void AddMachineCpuUsage(MachinePerfStatisticsSample* machine_sample,
                          const NodeStatistics& node_stats);
  void PopulateNodeStats(const string& res_id,
                         const NodeStatistics& node_stats);
  void PopulatePodStats(TaskID_t task_id, const string& node,
                        const PodStatistics& pod_stats);
  void PopulateTaskFinalReport(TaskDescriptor* td_ptr, TaskFinalReport* report);
 private:
  boost::shared_ptr<KnowledgeBase> knowledge_base_;
  WallTime time_manager_;
};

}  // namespace poseidon

#endif  // POSEIDON_FIRMAMENT_KNOWLEDGE_BASED_POPULATOR_H
