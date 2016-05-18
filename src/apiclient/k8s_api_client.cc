#include <iostream>
#include <streambuf>
#include <sstream>
#include <fstream>

#include <map>
#include <vector>
#include <string>
#include <exception>

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
  return node_client.request(methods::GET).then([](http_response resp) {
    return resp.extract_json();
  }).then([](json::value nodes_json) {
    json::value nodes_result_node = json::value::object();

    if (!nodes_json[U("items")].is_null()) {
      nodes_result_node[U("nodes")] = json::value::array();

      int i = 0;
      for (auto& iter : nodes_json[U("items")].as_array()) {
        ucout << "node!" << endl;
        auto& node = iter.as_object();
        const auto& nName = node[U("status")].as_object()[U("nodeInfo")].as_object().find(U("machineID"));
        if (nName == node[U("status")].as_object()[U("nodeInfo")].as_object().end()) {
          throw web::json::json_exception(U("name key not found"));
        }
        nodes_result_node[U("nodes")][i][U("id")] = nName->second;
        ++i;
      }
    } else {
      // Node data is null, we hit an error.
      nodes_result_node[U("nodes")] = json::value::object();
      nodes_result_node[U("nodes")][U("error")] =
        json::value::string(U("no nodes!"));
    }

    return nodes_result_node;
  }).then([=](pplx::task<json::value> t) {
    return HandleTaskException(t, U("nodes"));
  });
}

K8sApiClient::K8sApiClient(const string& host, const string& port) {
  utility::string_t address = U("http://" + U(host) + ":" + U(port));
  address.append(port);

  base_uri_ = http::uri(address);
}

int K8sApiClient::AllNodes(void) {
  pplx::task<json::value> t = get_nodes_task(base_uri_.to_string(), "");

  t.wait();

  json::value jval = t.get();
  ucout << jval.to_string() << endl;

  return 0;
}

int K8sApiClient::NodesWithLabel(const string& label) {
  pplx::task<json::value> t = get_nodes_task(base_uri_.to_string(), U(label));

  t.wait();

  json::value jval = t.get();
  ucout << jval.to_string() << endl;

  return 0;
}

}  // namespace apiclient
}  // namespace poseidon
