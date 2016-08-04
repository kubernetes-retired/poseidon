#include <exception>
#include <fstream>
#include <iostream>
#include <map>
#include <sstream>
#include <streambuf>
#include <string>
#include <vector>

#include <gflags/gflags.h>
#include <glog/logging.h>

#include "cpprest/http_client.h"
#include "cpprest/json.h"

#include "apiclient/utils.h"

#include "apiclient/k8s_api_client.h"

DEFINE_string(k8s_apiserver_host, "localhost",
              "Hostname of the Kubernetes API server.");
DEFINE_string(k8s_apiserver_port, "8080",
              "Port number for Kubernetes API server.");
DEFINE_string(k8s_api_version, "v1", "Kubernetes API version to use.");

using namespace std;
using namespace web;
using namespace json;
using namespace utility;
using namespace http;
using namespace http::client;

namespace poseidon {
namespace apiclient {

inline string api_prefix(void) { return "/api/" + FLAGS_k8s_api_version + "/"; }

K8sApiClient::K8sApiClient() {
  utility::string_t address = U("http://" + U(FLAGS_k8s_apiserver_host) + ":" +
                                U(FLAGS_k8s_apiserver_port));

  base_uri_ = http::uri(address);

  LOG(INFO) << "Starting K8sApiClient for API server at "
            << base_uri_.to_string();
}

pplx::task<json::value> K8sApiClient::BindPodTask(
    const utility::string_t& base_uri, const string& k8s_namespace,
    const string& pod_name, const string& node_name) {
  uri_builder ub(base_uri);

  ub.append_path(
      U(api_prefix() + "namespaces/" + U(k8s_namespace) + "/bindings"));

  json::value body_json = json::value::object();
  body_json["target"] = json::value::object();
  body_json["target"]["name"] = json::value::string(node_name);
  body_json["metadata"] = json::value::object();
  body_json["metadata"]["name"] = json::value::string(pod_name);

  http_client node_client(ub.to_uri());
  return node_client.request(methods::POST, U(""), U(body_json))
      .then([=](http_response resp) { return resp.extract_json(); })
      .then([=](json::value resp) {
        LOG(INFO) << "Parsing binding response: " << resp;
        json::value result = json::value::object();

        return result;
      })
      .then([=](pplx::task<json::value> t) {
        // If there was an exception, modify the response to contain an error
        return HandleTaskException(t, U("status"));
      });
}

// Given base URI and label selector, fetch list of nodes that match.
// Returns a task of json::value of node data
// JSON result Format:
// {"items":["metadata": { "name": "..." }, ... ]}
pplx::task<json::value> K8sApiClient::GetNodesTask(
    const utility::string_t& base_uri,
    const utility::string_t& label_selector) {
  uri_builder ub(base_uri);

  ub.append_path(U(api_prefix() + "nodes"));
  if (!label_selector.empty()) {
    ub.append_query("labelSelector", label_selector);
  }

  http_client node_client(ub.to_uri());
  return node_client.request(methods::GET)
      .then([=](http_response resp) { return resp.extract_json(); })
      .then([=](json::value nodes_json) {
        VLOG(3) << "Parsing response: " << nodes_json;
        json::value nodes_result_node = json::value::object();

        if (nodes_json.is_object() &&
            !nodes_json.as_object()[U("items")].is_null()) {
          auto& nList = nodes_json[U("items")].as_array();
          nodes_result_node[U("nodes")] = json::value::array(nList.size());

          uint32_t index = 0;
          for (auto& iter : nList) {
            auto& node = iter.as_object();
            auto& nStatus = node[U("status")].as_object();
            auto& nInfo = nStatus[U("nodeInfo")].as_object();
            const auto& nName = nInfo.find(U("machineID"));
            if (nName == nInfo.end()) {
              LOG(ERROR) << "Failed to find machineID for node!";
            }
            nodes_result_node[U("nodes")][index][U("id")] = nName->second;
            nodes_result_node[U("nodes")][index][U("hostname")] =
                node[U("metadata")][U("name")];
            ++index;
          }
        } else {
          LOG(ERROR)
              << "No nodes found in API server response for label selector "
              << label_selector;
          // Node data is null, we hit an error, so return empty list.
          nodes_result_node[U("nodes")] = json::value::array(0);
        }

        return nodes_result_node;
      })
      .then([=](pplx::task<json::value> t) {
        // If there was an exception, modify the response to contain an error
        return HandleTaskException(t, U("status"));
      });
}

pplx::task<json::value> K8sApiClient::GetPodsTask(
    const utility::string_t& base_uri,
    const utility::string_t& label_selector) {
  uri_builder ub(base_uri);

  ub.append_path(U(api_prefix() + "pods"));
  if (!label_selector.empty()) {
    ub.append_query("labelSelector", label_selector);
  }

  http_client node_client(ub.to_uri());
  return node_client.request(methods::GET)
      .then([=](http_response resp) { return resp.extract_json(); })
      .then([=](json::value pods_json) {
        VLOG(3) << "Parsing response: " << pods_json;
        json::value result = json::value::object();

        if (pods_json.is_object() &&
            !pods_json.as_object()[U("items")].is_null()) {
          auto& pList = pods_json[U("items")].as_array();
          result[U("pods")] = json::value::array(pList.size());

          uint32_t index = 0;
          for (auto& iter : pList) {
            auto& pod = iter.as_object();
            auto& pName = pod[U("metadata")].as_object()[U("name")];
            result[U("pods")][index][U("name")] = pName;
            auto& pStatus = pod[U("status")].as_object();
            result[U("pods")][index][U("state")] = pStatus[U("phase")];
            ++index;
          }
        } else {
          LOG(ERROR)
              << "No pods found in API server response for label selector "
              << label_selector;
          // Node data is null, we hit an error, so return empty list.
          result[U("pods")] = json::value::array(0);
        }

        return result;
      })
      .then([=](pplx::task<json::value> t) {
        // If there was an exception, modify the response to contain an error
        return HandleTaskException(t, U("status"));
      });
}

vector<pair<string, string>> K8sApiClient::AllNodes(void) {
  return NodesWithLabel("");
}

vector<string> K8sApiClient::AllPods(void) { return PodsWithLabel(""); }

bool K8sApiClient::BindPodToNode(const string& pod_name,
                                 const string& node_name) {
  pplx::task<json::value> t =
      BindPodTask(base_uri_.to_string(), U("default"), pod_name, node_name);

  try {
    t.wait();

    json::value jval = t.get();

    if (jval[U("status")].is_null() ||
        jval[U("status")].as_object()[U("error")].is_null()) {
      LOG(INFO) << "Bound " << pod_name << " to " << node_name;
    } else {
      LOG(ERROR) << "Failed to bind pod: " << jval[U("status")][U("error")];
    }
  } catch (const std::exception& e) {
    LOG(ERROR) << "Exception while binding pod: " << e.what();
    return false;
  }
  return true;
}

vector<pair<string, string>> K8sApiClient::NodesWithLabel(const string& label) {
  vector<pair<string, string>> nodes;
  pplx::task<json::value> t = GetNodesTask(base_uri_.to_string(), U(label));

  try {
    t.wait();

    json::value jval = t.get();

    if (jval[U("status")].is_null() ||
        jval[U("status")].as_object()[U("error")].is_null()) {
      for (auto& iter : jval["nodes"].as_array()) {
        nodes.push_back(pair<string, string>(iter["id"].as_string(),
                                             iter["hostname"].as_string()));
      }
    } else {
      LOG(ERROR) << "Failed to get nodes: " << jval[U("status")][U("error")];
    }
  } catch (const std::exception& e) {
    LOG(ERROR) << "Exception while waiting for node list: " << e.what();
  }

  return nodes;
}

vector<string> K8sApiClient::PodsWithLabel(const string& label) {
  vector<string> pods;
  pplx::task<json::value> t = GetPodsTask(base_uri_.to_string(), U(label));

  try {
    t.wait();

    json::value jval = t.get();

    if (jval[U("status")].is_null() ||
        jval[U("status")].as_object()[U("error")].is_null()) {
      for (auto& iter : jval["pods"].as_array()) {
        pods.push_back(iter["name"].as_string());
      }
    } else {
      LOG(ERROR) << "Failed to get pods: " << jval[U("status")][U("error")];
    }
  } catch (const std::exception& e) {
    LOG(ERROR) << "Exception while waiting for pod list: " << e.what();
  }

  return pods;
}

}  // namespace apiclient
}  // namespace poseidon
