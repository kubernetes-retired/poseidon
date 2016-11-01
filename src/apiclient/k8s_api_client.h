#ifndef POSEIDON_APICLIENT_K8S_API_CLIENT_H
#define POSEIDON_APICLIENT_K8S_API_CLIENT_H

#include <string>

#include "cpprest/http_client.h"
#include "cpprest/json.h"

#include "apiclient/utils.h"

using namespace std;
using namespace web;
using namespace json;
using namespace utility;
using namespace http;
using namespace http::client;

namespace poseidon {
namespace apiclient {

class K8sApiClient {
 public:
  K8sApiClient();
  vector<pair<string, NodeStatistics>> AllNodes(void);
  vector<PodStatistics> AllPods(void);
  vector<pair<string, NodeStatistics>> NodesWithLabel(const string& label);
  vector<PodStatistics> PodsWithLabel(const string& label);
  bool BindPodToNode(const string& pod_name, const string& node_name);

 private:
  pplx::task<json::value> BindPodTask(const utility::string_t& base_uri,
                                      const string& k8s_namespace,
                                      const string& pod_name,
                                      const string& node_name);
  pplx::task<json::value> GetNodesTask(const utility::string_t& base_uri,
                                       const utility::string_t& label_selector);
  pplx::task<json::value> GetPodsTask(const utility::string_t& base_uri,
                                      const utility::string_t& label_selector);

  // API server URI
  web::uri base_uri_;
};

}  // namespace apiclient
}  // namespace poseidon

#endif  // POSEIDON_APICLIENT_K8S_API_CLIENT_H
