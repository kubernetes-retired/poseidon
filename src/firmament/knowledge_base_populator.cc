// The Poseidon project
// Copyright (c) 2016 Ionel Gog <ionel.gog@cl.cam.ac.uk>
//

#include "firmament/knowledge_base_populator.h"

#include "base/units.h"

using firmament::KnowledgeBase;
using firmament::CpuUsage;

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
      node_stats.memory_capacity_kb_ / firmament::KB_TO_MB);
  machine_stats.set_free_ram(
      node_stats.memory_allocatable_kb_ / firmament::KB_TO_MB);
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

void KnowledgeBasePopulator::ProcessFinalPodReport() {
  // TODO(ionel): Implement!
  TaskFinalReport task_final_report;
  // task_final_report.set_task_id(task_id);
  // task_final_report.set_start_time();
  // task_final_report.set_finish_time();
  // task_final_report.set_instructions();
  // task_final_report.set_cycles();
  // task_final_report.set_llc_refs();
  // task_final_report.set_llc_misses();
  // task_final_report.set_runtime();
  // knowledge_base_->ProcessTaskFinalReport(equiv_classes, task_final_report);
}

}  // namespace poseidon
