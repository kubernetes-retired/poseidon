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

#include "firmament/knowledge_base_populator.h"

#include "base/units.h"

using firmament::CpuUsage;
using firmament::KB_TO_MB;
using firmament::KnowledgeBase;

namespace poseidon {

KnowledgeBasePopulator::KnowledgeBasePopulator(
    boost::shared_ptr<KnowledgeBase> knowledge_base) :
  knowledge_base_(knowledge_base) {
}

void KnowledgeBasePopulator::AddMachineCpuUsage(
    MachinePerfStatisticsSample* machine_stats,
    const NodeStatistics& node_stats) {
  for (uint64_t cpu_index = 0; cpu_index < node_stats.cpu_capacity_;
       cpu_index++) {
    CpuUsage* cpu_usage = machine_stats->add_cpus_usage();
    // TODO(ionel): Get more fine-grained info.
    if (cpu_index < node_stats.cpu_allocatable_) {
      if (cpu_index + 1 <= node_stats.cpu_allocatable_) {
        cpu_usage->set_idle(100.0);
      } else {
        cpu_usage->set_idle((node_stats.cpu_allocatable_ - cpu_index) * 100.0);
      }
    } else {
      cpu_usage->set_idle(0.0);
    }
    // TODO(ionel): Populate the other fields. They're currently not used,
    // but they may be used in the future.
    // cpu_usage->set_user();
    // cpu_usage->set_nice();
    // cpu_usage->set_system();
    // cpu_usage->set_iowait(it->iowait());
    // cpu_usage->set_irq(it->irq());
    // cpu_usage->set_soft_irq(it->soft_irq());
    // cpu_usage->set_steal(it->steal());
    // cpu_usage->set_guest(it->guest());
    // cpu_usage->set_guest_nice(it->guest_nice());
  }
}

void KnowledgeBasePopulator::PopulateNodeStats(
    const string& res_id,
    const NodeStatistics& node_stats) {
  MachinePerfStatisticsSample machine_stats;
  machine_stats.set_resource_id(res_id);
  machine_stats.set_timestamp(time_manager_.GetCurrentTimestamp());
  machine_stats.set_total_ram(
      node_stats.memory_capacity_kb_ / KB_TO_MB);
  machine_stats.set_free_ram(
      node_stats.memory_allocatable_kb_ / KB_TO_MB);
  // TODO(ionel): Get more accurate CPU values.
  AddMachineCpuUsage(&machine_stats, node_stats);
  // TODO(ionel): Get real disk and network values.
  machine_stats.set_disk_bw(50);
  machine_stats.set_net_tx_bw(1250);
  machine_stats.set_net_rx_bw(1250);
  knowledge_base_->AddMachineSample(machine_stats);
}

void KnowledgeBasePopulator::PopulatePodStats(TaskID_t task_id,
                                              const string& node,
                                              const PodStatistics& pod_stats) {
  TaskPerfStatisticsSample task_stats;
  task_stats.set_task_id(task_id);
  task_stats.set_timestamp(time_manager_.GetCurrentTimestamp());
  task_stats.set_hostname(node);
  // TODO(ionel): Populate the other fields. They're currently not used,
  // but they may be used in the future.
  // task_stats.set_vsize();
  // task_stats.set_rsize();
  // task_stats.set_sched_run();
  // task_stats.set_sched_wait();
  task_stats.set_completed(false);
  knowledge_base_->AddTaskSample(task_stats);
}

void KnowledgeBasePopulator::PopulateTaskFinalReport(const TaskDescriptor& td,
                                                     TaskFinalReport* report) {
  // TODO(ionel): Implement!
  // task_final_report.set_task_id(task_id);
  // task_final_report.set_start_time();
  // task_final_report.set_finish_time();
  // task_final_report.set_instructions();
  // task_final_report.set_cycles();
  // task_final_report.set_llc_refs();
  // task_final_report.set_llc_misses();
  // task_final_report.set_runtime();
}

}  // namespace poseidon
