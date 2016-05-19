#include <iostream>
#include <fstream>
#include <map>
#include <sstream>
#include <streambuf>
#include <string>
#include <exception>
#include <vector>

#include <glog/logging.h>
#include <gflags/gflags.h>

#include "cpprest/http_client.h"
#include "cpprest/json.h"

#include "apiclient/utils.h"

#include "apiclient/k8s_api_client.h"

using namespace std;
using namespace web;
using namespace json;
using namespace utility;
using namespace http;
using namespace http::client;

namespace poseidon {
namespace apiclient {

// Given base URI and label selector, fetch list of nodes that match.
// Returns a task of json::value of node data
// JSON result Format:
// {"items":["metadata": { "name": "..." }, ... ]}
pplx::task<json::value> K8sApiClient::get_nodes_task(
    const utility::string_t& base_uri,
    const utility::string_t& label_selector) {
  uri_builder ub(base_uri);

  ub.append_path(U("/api/v1/nodes"));
  if (!label_selector.empty()) {
    ub.append_query("labelSelector", label_selector);
  }
  http::uri node_uri = ub.to_uri();

  http_client node_client(node_uri);
  return node_client.request(methods::GET).then([=](http_response resp) {
    return resp.extract_json();
  }).then([=](json::value nodes_json) {
    VLOG(3) << "Parsing response: " << nodes_json;
    json::value nodes_result_node = json::value::object();

    if (nodes_json.is_object() &&
        !nodes_json.as_object()[U("items")].is_null()) {
      auto& nList = nodes_json[U("items")].as_array();
      nodes_result_node[U("nodes")] = json::value::array(nList.size());

      int i = 0;
      for (auto& iter : nList) {
        auto& node = iter.as_object();
        auto& nStatus = node[U("status")].as_object();
        auto& nInfo = nStatus[U("nodeInfo")].as_object();
        const auto& nName = nInfo.find(U("machineID"));
        if (nName == nInfo.end()) {
          LOG(ERROR) << "Failed to find machineID for node!";
        }
        nodes_result_node[U("nodes")][i][U("id")] = nName->second;
        nodes_result_node[U("nodes")][i][U("hostname")] = node[U("metadata")][U("name")];
        ++i;
      }
    } else {
      LOG(ERROR) << "No nodes found in API server response for label selector "
                 << label_selector;
      // Node data is null, we hit an error, so return empty list.
      nodes_result_node[U("nodes")] = json::value::array(0);
    }

    return nodes_result_node;
  }).then([=](pplx::task<json::value> t) {
    // If there was an exception, modify the response to contain an error
    return HandleTaskException(t, U("status"));
  });
}

K8sApiClient::K8sApiClient(const string& host, const string& port) {
  utility::string_t address = U("http://" + U(host) + ":" + U(port));

  base_uri_ = http::uri(address);

  LOG(INFO) << "Starting K8sApiClient for API server at "
            << base_uri_.to_string();
}

vector<string> K8sApiClient::AllNodes(void) {
  return NodesWithLabel("");
}

vector<string> K8sApiClient::NodesWithLabel(const string& label) {
  vector<string> nodes;
  pplx::task<json::value> t = get_nodes_task(base_uri_.to_string(), U(label));

  try {
    t.wait();

    json::value jval = t.get();

    if (jval[U("status")].is_null() ||
        jval[U("status")].as_object()[U("error")].is_null()) {
      for (auto& iter : jval["nodes"].as_array()) {
        nodes.push_back(iter["id"].as_string());
      }
    } else {
      LOG(ERROR) << "Failed to get nodes: " << jval[U("status")][U("error")];
    }
  } catch (const std::exception &e) {
    LOG(ERROR) << "Exception while waiting for node list: " << e.what();
  }

  return nodes;
}

}  // namespace apiclient
}  // namespace poseidon
