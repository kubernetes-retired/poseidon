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

#include "apiclient/k8s_api_client.h"

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

#include "misc/map-util.h"

#include "apiclient/utils.h"

DEFINE_string(k8s_apiserver_host, "localhost",
              "Hostname of the Kubernetes API server.");
DEFINE_string(k8s_apiserver_port, "8080",
              "Port number for Kubernetes API server.");
DEFINE_string(k8s_api_version, "v1", "Kubernetes API version to use.");

using firmament::ContainsKey;

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

pplx::task<json::value> K8sApiClient::DeletePod(
    const utility::string_t& base_uri,
    const string& k8s_namespace,
    const string& pod_name) {
  uri_builder ub(base_uri);
  ub.append_path(U(api_prefix() + "namespaces/" + k8s_namespace + "/pods/" +
                   pod_name));
  http_client client(ub.to_uri());
  return client.request(methods::DEL)
      .then([=](http_response resp) { return resp.extract_json(); })
      .then([=](json::value resp) {
        json::value result = json::value::object();
        CHECK(resp.is_object());
        auto& resp_object = resp.as_object();
        if (!resp_object[U("status")].is_null()) {
          if (resp_object[U("status")].is_string()) {
            result[U("status")] = resp_object[U("status")];
          } else if (resp_object[U("status")].is_object()) {
            // There's no explicit status to denote that a pod was deleted,
            // but if the reply contains a status object with a phase field
            // the the pod has been deleted.
            CHECK(!resp_object[U("status")][U("phase")].is_null());
            result[U("status")] = json::value("Deleted");
          } else {
            LOG(FATAL) << "Delete reply for pod " << pod_name
                     << " contain unknown type of status field";
          }
        } else {
          LOG(ERROR) << "Delete reply for pod " << pod_name
                     << " doesn't contain status field";
          result[U("status")] = json::value("Failure");
        }
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
            auto& nCapacity = nStatus[U("capacity")].as_object();
            auto& nAllocatable = nStatus[U("allocatable")].as_object();
            const auto& nName = nInfo.find(U("machineID"));
            if (nName == nInfo.end()) {
              LOG(ERROR) << "Failed to find machineID for node!";
            }
            nodes_result_node[U("nodes")][index][U("id")] = nName->second;
            nodes_result_node[U("nodes")][index][U("hostname")] =
                node[U("metadata")][U("name")];
            nodes_result_node[U("nodes")][index][U("cpu_allocatable")] =
              nAllocatable.find(U("cpu"))->second;
            nodes_result_node[U("nodes")][index][U("cpu_capacity")] =
              nCapacity.find(U("cpu"))->second;
            nodes_result_node[U("nodes")][index][U("mem_allocatable")] =
              nAllocatable.find(U("memory"))->second;
            nodes_result_node[U("nodes")][index][U("mem_capacity")] =
              nCapacity.find(U("memory"))->second;
            auto& conditions = nStatus[U("conditions")].as_array();
            for (auto& iter_condition : conditions) {
              auto& condition = iter_condition.as_object();
              auto& condition_type =
                condition.find(U("type"))->second.as_string();
              if (condition_type == "OutOfDisk") {
                nodes_result_node[U("nodes")][index][U("out_of_disk")] =
                  condition.find(U("status"))->second;
              } else if (condition_type == "Ready") {
                nodes_result_node[U("nodes")][index][U("ready")] =
                  condition.find(U("status"))->second;
              }
            }
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

  ub.append_path(U(api_prefix() + "namespaces/default/pods"));
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
            auto& metadata = pod[U("metadata")].as_object();
            auto& pName = metadata[U("name")];
            auto& pLabels = metadata[U("labels")].as_object();
            result[U("pods")][index][U("name")] = pName;
            result[U("pods")][index][U("controller-uid")] =
              pLabels[U("controller-uid")];
            auto& pStatus = pod[U("status")].as_object();
            result[U("pods")][index][U("state")] = pStatus[U("phase")];
            result[U("pods")][index][U("containers")] =
              pod[U("spec")].as_object()[U("containers")];
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

vector<pair<string, NodeStatistics>> K8sApiClient::AllNodes(void) {
  return NodesWithLabel("");
}

vector<PodStatistics> K8sApiClient::AllPods(void) {
  return PodsWithLabel("");
}

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

vector<pair<string, NodeStatistics>> K8sApiClient::NodesWithLabel(
    const string& label) {
  vector<pair<string, NodeStatistics>> nodes;
  pplx::task<json::value> t = GetNodesTask(base_uri_.to_string(), U(label));

  try {
    t.wait();

    json::value jval = t.get();

    if (jval[U("status")].is_null() ||
        jval[U("status")].as_object()[U("error")].is_null()) {
      for (auto& iter : jval["nodes"].as_array()) {
        NodeStatistics node_stats;
        node_stats.hostname_ = iter["hostname"].as_string();
        node_stats.cpu_capacity_ =
          StringRequestToCPU(iter["cpu_capacity"].as_string());
        node_stats.cpu_allocatable_ =
          StringRequestToCPU(iter["cpu_allocatable"].as_string());
        auto& mem_cap = iter["mem_capacity"].as_string();
        node_stats.memory_capacity_kb_ = StringRequestToKB(mem_cap);
        auto& mem_allocatable = iter["mem_allocatable"].as_string();
        node_stats.memory_allocatable_kb_ = StringRequestToKB(mem_allocatable);
        string node_out_of_disk = iter["out_of_disk"].as_string();
        if (node_out_of_disk == "False") {
          node_stats.is_out_of_disk_ = false;
        } else if (node_out_of_disk == "True" ||
                   node_out_of_disk == "Unknown") {
          node_stats.is_out_of_disk_ = true;
        } else {
          node_stats.is_out_of_disk_ = true;
          LOG(WARNING) << "Unexpected out of disk state: " << node_out_of_disk
                       << " for node " << node_stats.hostname_;
        }
        string node_ready = iter["ready"].as_string();
        if (node_ready == "True") {
          node_stats.is_ready_ = true;
        } else if (node_ready == "Unknown" || node_ready == "False") {
          node_stats.is_ready_ = false;
        } else {
          node_stats.is_ready_ = false;
          LOG(WARNING) << "Unexpected ready state: " << node_ready
                       << " for node " << node_stats.hostname_;
        }
        nodes.push_back(pair<string, NodeStatistics>(iter["id"].as_string(),
                                                     node_stats));
      }
    } else {
      LOG(ERROR) << "Failed to get nodes: " << jval[U("status")][U("error")];
    }
  } catch (const std::exception& e) {
    LOG(ERROR) << "Exception while waiting for node list: " << e.what();
  }

  return nodes;
}

bool K8sApiClient::DeletePod(const string& pod_name) {
  pplx::task<json::value> t =
    DeletePod(base_uri_.to_string(), U("default"), pod_name);
  try {
    t.wait();
    json::value jval = t.get();
    if (jval[U("status")].is_null() ||
        jval[U("status")].as_string() == "Failure") {
      return false;
    } else if (jval[U("status")].as_string() == "Deleted") {
      return true;
    }
    return false;
  } catch (const std::exception& e) {
    LOG(ERROR) << "Exception while deleting pod: " << e.what();
  }
  return false;
}

vector<PodStatistics> K8sApiClient::PodsWithLabel(
    const string& label) {
  vector<PodStatistics> pods;
  pplx::task<json::value> t = GetPodsTask(base_uri_.to_string(), U(label));

  try {
    t.wait();
    json::value jval = t.get();
    if (jval[U("status")].is_null() ||
        jval[U("status")].as_object()[U("error")].is_null()) {
      for (auto& iter : jval["pods"].as_array()) {
        auto& pContainerList = iter["containers"].as_array();
        double cpu_request = 0;
        uint64_t memory_request = 0;
        for (auto& iter : pContainerList) {
          auto& container = iter.as_object();
          if (ContainsKey(container, "resources")) {
            auto& container_res = container[U("resources")].as_object();
            if (ContainsKey(container_res, "requests")) {
              auto& container_req = container_res[U("requests")].as_object();
              cpu_request +=
                StringRequestToCPU(
                    container_req.find(U("cpu"))->second.as_string());
              memory_request +=
                StringRequestToKB(
                    container_req.find(U("memory"))->second.as_string());
            } else {
              LOG(INFO) << "Container " << container[U("name")].as_string() <<
                " json object does not have resource requests object";
              // TODO(ionel): Set resource requests default values.
            }
          } else {
            LOG(INFO) << "Container " << container[U("name")].as_string() <<
              " json object does not have resources object";
            // TODO(ionel): Set resource requests default values.
          }
        }
        PodStatistics pod_stats;
        pod_stats.name_ = iter["name"].as_string();
        pod_stats.state_ = iter["state"].as_string();
        pod_stats.controller_id_ = iter["controller-uid"].as_string();
        pod_stats.cpu_request_ = cpu_request;
        pod_stats.memory_request_kb_ = memory_request;
        pods.push_back(pod_stats);
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
