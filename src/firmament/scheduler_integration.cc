#include <unordered_map>

#include <gflags/gflags.h>
#include <glog/logging.h>

#include "apiclient/k8s_api_client.h"
#include "apiclient/utils.h"
#include "firmament/scheduler_bridge.h"

DEFINE_int64(polling_frequency, 10000000,
             "K8s API polling frequency, in microseconds");
// XXX(malte): hack to make things compile
DEFINE_string(listen_uri, "", "");

using poseidon::apiclient::K8sApiClient;

int main(int argc, char** argv) {
  google::ParseCommandLineFlags(&argc, &argv, false);
  google::InitGoogleLogging(argv[0]);

  poseidon::SchedulerBridge scheduler_bridge;
  K8sApiClient api_client;

  // main loop -- keep looking for nodes and pods
  while (true) {
    // Poll nodes
    vector<pair<string, poseidon::apiclient::NodeStatistics>> nodes =
      api_client.AllNodes();
    if (!nodes.empty()) {
      for (auto& n : nodes) {
        // node_id, hostname
        scheduler_bridge.CreateResourceForNode(n.first, n.second.hostname_);
        scheduler_bridge.AddStatisticsForNode(n.first, n.second);
      }
    }

    // Poll pods
    vector<poseidon::apiclient::PodStatistics> pods = api_client.AllPods();
    unordered_map<string, string>* pod_node_bindings =
      scheduler_bridge.RunScheduler(pods);
    for (auto& pod_node : *pod_node_bindings) {
      api_client.BindPodToNode(pod_node.first, pod_node.second);
    }
    delete pod_node_bindings;
    // Sleep a bit until we poll again
    usleep(FLAGS_polling_frequency);
  }
  return 0;
}
