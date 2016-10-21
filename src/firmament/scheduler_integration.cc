#include "base/resource_status.h"
#include "base/resource_topology_node_desc.pb.h"
#include "misc/map-util.h"
#include "misc/trace_generator.h"
#include "misc/utils.h"
#include "misc/wall_time.h"
#include "platforms/sim/simulated_messaging_adapter.h"
#include "scheduling/flow/flow_scheduler.h"
#include "scheduling/scheduling_delta.pb.h"

#include "apiclient/k8s_api_client.h"

#include <gflags/gflags.h>
#include <glog/logging.h>

DEFINE_int64(polling_frequency, 10000000,
             "K8s API polling frequency, in microseconds");
// XXX(malte): hack to make things compile
DEFINE_string(listen_uri, "", "");

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
using firmament::TaskID_t;
using firmament::TaskDescriptor;
using firmament::TaskMap_t;
using firmament::KnowledgeBase;
using firmament::ResourceTopologyNodeDescriptor;
using firmament::TraceGenerator;
using firmament::WallTime;
using firmament::platform::sim::SimulatedMessagingAdapter;
using firmament::scheduler::FlowScheduler;
using firmament::scheduler::ObjectStoreInterface;
using firmament::scheduler::TopologyManager;

using poseidon::apiclient::K8sApiClient;

boost::shared_ptr<JobMap_t> job_map_;
boost::shared_ptr<KnowledgeBase> knowledge_base_;
boost::shared_ptr<ObjectStoreInterface> obj_store_;
boost::shared_ptr<ResourceMap_t> resource_map_;
boost::shared_ptr<TaskMap_t> task_map_;
boost::shared_ptr<TopologyManager> topology_manager_;

map<ResourceID_t, string> node_map_;
unordered_map<string, TaskID_t> pod_to_task_map_;
unordered_map<TaskID_t, string> task_to_pod_map_;

JobDescriptor* CreateJobForPod(const string& pod) {
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

ResourceStatus* CreateTopLevelResource(void) {
  ResourceID_t res_id = firmament::GenerateResourceID();
  ResourceTopologyNodeDescriptor* rtnd = new ResourceTopologyNodeDescriptor();
  // Set up the RD
  ResourceDescriptor* rd = rtnd->mutable_resource_desc();
  rd->set_uuid(firmament::to_string(res_id));
  rd->set_type(ResourceDescriptor::RESOURCE_COORDINATOR);
  // Need to maintain a ResourceStatus for the resource map
  // TODO(malte): don't pass localhost here
  ResourceStatus* rs = new ResourceStatus(rd, rtnd, "localhost", 0);
  // Insert into resource map
  CHECK(InsertIfNotPresent(resource_map_.get(), res_id, rs));
  return rs;
}

ResourceStatus* CreateResourceForNode(ResourceID_t node_id,
                                      ResourceID_t parent_id) {
  ResourceTopologyNodeDescriptor* r = new ResourceTopologyNodeDescriptor();
  // Create and initialize RD
  ResourceDescriptor* rd = r->mutable_resource_desc();
  rd->set_uuid(firmament::to_string(node_id));
  rd->set_type(ResourceDescriptor::RESOURCE_PU);
  rd->set_state(ResourceDescriptor::RESOURCE_IDLE);
  r->set_parent_id(firmament::to_string(parent_id));
  // Need to maintain a ResourceStatus for the resource map
  // TODO(malte): set hostname correctly
  ResourceStatus* rs = new ResourceStatus(rd, r, "", 0);
  // Insert into resource map
  CHECK(InsertIfNotPresent(resource_map_.get(), node_id, rs));
  return rs;
}

int main(int argc, char** argv) {
  google::ParseCommandLineFlags(&argc, &argv, false);
  google::InitGoogleLogging(argv[0]);

  // Kubernetes API client
  K8sApiClient api_client;

  job_map_.reset(new JobMap_t);
  task_map_.reset(new TaskMap_t);
  resource_map_.reset(new ResourceMap_t);

  ResourceStatus* toplevel_res_status = CreateTopLevelResource();
  ResourceID_t toplevel_res_id =
      firmament::ResourceIDFromString(toplevel_res_status->descriptor().uuid());

  SimulatedMessagingAdapter<BaseMessage> ma;
  WallTime wall_time;
  TraceGenerator tg(&wall_time);

  FlowScheduler fs(job_map_, resource_map_,
                   toplevel_res_status->mutable_topology_node(), obj_store_,
                   task_map_, knowledge_base_, topology_manager_, &ma, NULL,
                   toplevel_res_id, "", &wall_time, &tg);
  LOG(INFO) << "Firmament scheduler instantiated: " << fs;

  // main loop -- keep looking for nodes and pods
  while (true) {
    // Poll nodes
    vector<pair<string, string>> nodes = api_client.AllNodes();
    if (!nodes.empty()) {
      for (auto& n : nodes) {
        ResourceID_t rid = firmament::ResourceIDFromString(n.first);
        // Check if we know about this node already
        if (!ContainsKey(*resource_map_, rid)) {
          LOG(INFO) << "Adding new node's resource with RID " << rid;
          CHECK(InsertIfNotPresent(&node_map_, rid, n.second));
          // Create a new Firmament resource
          ResourceStatus* rs = CreateResourceForNode(rid, toplevel_res_id);
          // Register with the scheudler
          // TODO(malte): we use a hack here -- we pass simulated=true to
          // avoid Firmament instantiating an actual executor for this resource.
          // Instead, we rely on the no-op SimulatedExecutor. We should change
          // it such that Firmament does not mandatorily create an executor.
          fs.RegisterResource(rs->mutable_topology_node(), false, true);
        }
      }
    }

    // Poll pods
    vector<string> pods = api_client.AllPods();
    if (!pods.empty()) {
      vector<string>::iterator unscheduled_pods_iter =
          find_if(pods.begin(), pods.end(), [](const string& p) -> bool {
            // TODO(malte): should use p.state() == "Pending" or similar
            return firmament::FindOrNull(pod_to_task_map_, p) == NULL;
          });
      if (unscheduled_pods_iter != pods.end()) {
        for_each(unscheduled_pods_iter, next(unscheduled_pods_iter, 1),
                 [&](string p) {
                   LOG(INFO) << "New unscheduled pod: " << p;
                   JobDescriptor* jd = CreateJobForPod(p);
                   CHECK(InsertIfNotPresent(&pod_to_task_map_, p,
                                            jd->root_task().uid()));
                   CHECK(InsertIfNotPresent(&task_to_pod_map_,
                                            jd->root_task().uid(), p));
                   fs.AddJob(jd);
                 });
      }
    }

    // Invoke Firmament scheduling
    firmament::scheduler::SchedulerStats sstat;
    vector<firmament::SchedulingDelta> deltas;
    fs.ScheduleAllJobs(&sstat, &deltas);

    // Apply results
    LOG(INFO) << "Got " << deltas.size() << " scheduling deltas";
    for (auto& d : deltas) {
      LOG(INFO) << "Delta: " << d.DebugString();
      if (d.type() == firmament::SchedulingDelta::PLACE) {
        const string* pod = FindOrNull(task_to_pod_map_, d.task_id());
        const string* node = FindOrNull(
            node_map_, firmament::ResourceIDFromString(d.resource_id()));
        CHECK_NOTNULL(pod);
        CHECK_NOTNULL(node);
        api_client.BindPodToNode(*pod, *node);
      } else {
        LOG(WARNING) << "Encountered unsupported scheduling delta of type "
                     << firmament::to_string(d.type());
      }
    }

    // Sleep a bit until we poll again
    usleep(FLAGS_polling_frequency);
  }
}
