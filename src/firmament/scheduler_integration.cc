#include "base/resource_topology_node_desc.pb.h"
#include "misc/utils.h"
#include "misc/trace_generator.h"
#include "misc/wall_time.h"
#include "platforms/sim/simulated_messaging_adapter.h"
#include "scheduling/flow/flow_scheduler.h"
#include "storage/stub_object_store.h"

#include <glog/logging.h>
#include <gflags/gflags.h>

DEFINE_string(listen_uri, "", "");

using boost::shared_ptr;
using firmament::BaseMessage;
using firmament::JobMap_t;
using firmament::ResourceDescriptor;
using firmament::ResourceMap_t;
using firmament::ResourceID_t;
using firmament::TaskMap_t;
using firmament::KnowledgeBase;
using firmament::ResourceTopologyNodeDescriptor;
using firmament::TraceGenerator;
using firmament::WallTime;
using firmament::platform::sim::SimulatedMessagingAdapter;
using firmament::scheduler::FlowScheduler;
using firmament::scheduler::ObjectStoreInterface;
using firmament::scheduler::TopologyManager;

shared_ptr<JobMap_t> job_map_;
shared_ptr<KnowledgeBase> knowledge_base_;
shared_ptr<ObjectStoreInterface> obj_store_;
shared_ptr<ResourceMap_t> resource_map_;
shared_ptr<TaskMap_t> task_map_;
shared_ptr<TopologyManager> topology_manager_;

int main(int argc, char** argv) {

  google::ParseCommandLineFlags(&argc, &argv, false);
  google::InitGoogleLogging(argv[0]);

  job_map_.reset(new JobMap_t);

  ResourceID_t res_id = firmament::GenerateResourceID();

  ResourceTopologyNodeDescriptor rtnd;
  rtnd.mutable_resource_desc()->set_uuid(
    firmament::to_string(res_id));
  rtnd.mutable_resource_desc()->set_type(
      ResourceDescriptor::RESOURCE_COORDINATOR);

  SimulatedMessagingAdapter<BaseMessage> ma;
  WallTime wall_time;
  TraceGenerator tg(&wall_time);

  FlowScheduler fs(job_map_, resource_map_, &rtnd, obj_store_,
                   task_map_, knowledge_base_, topology_manager_,
                   &ma, NULL, res_id, "", &wall_time, &tg);

  LOG(INFO) << fs;
}
