// The Poseidon project
// Copyright (c) 2016 Ionel Gog <ionel.gog@cl.cam.ac.uk>
//

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
  void ProcessFinalPodReport();
 private:
  boost::shared_ptr<KnowledgeBase> knowledge_base_;
  WallTime time_manager_;
};

}  // namespace poseidon

#endif  // POSEIDON_FIRMAMENT_KNOWLEDGE_BASED_POPULATOR_H
