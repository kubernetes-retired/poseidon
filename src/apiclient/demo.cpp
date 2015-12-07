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

using namespace std;
using namespace web;
using namespace json;
using namespace utility;
using namespace http;
using namespace http::client;

http_response PrintResponse(const std::string &url,
                            const http_response &response)
{
    ucout << "Got response for '" << url << "': "
          << response.extract_json().get().serialize()
          << endl;
    return response;
}

// In case of any failure to fetch data from a service,
// create a json object with "error" key and value containing the exception
// details.
pplx::task<json::value> handle_exception(pplx::task<json::value>& t,
                                         const utility::string_t& field_name)
{
  try
  {
    t.get();
  }
  catch(const std::exception& ex)
  {
    json::value error_json = json::value::object();
    error_json[field_name] = json::value::object();
    error_json[field_name][U("error")] =
      json::value::string(utility::conversions::to_string_t(ex.what()));
    return pplx::task_from_result<json::value>(error_json);
  }

  return t;
}

// Given base URI and label selector, fetch list of nodes that match.
// Returns a task of json::value of node data
// JSON result Format:
// {"items":["metadata": { "name": "..." }, ... ]}
pplx::task<json::value> GetNodes(const utility::string_t& base_uri,
                                 const utility::string_t& label_selector)
{
  uri_builder ub(base_uri);

  ub.append_path(U("/api/v1/nodes"));
  if (!label_selector.empty()) {
    ub.append_query("labelSelector", label_selector);
  }
  http::uri node_uri = ub.to_uri();

  http_client node_client(node_uri);
  return node_client.request(methods::GET).then([](http_response resp)
  {
    //PrintResponse(full_uri.to_string(), apiClient.request(methods::GET, buf.str()).get());
    return resp.extract_json();
  }).then([](json::value nodes_json)
  {
    json::value nodes_result_node = json::value::object();

    if (!nodes_json[U("items")].is_null())
    {
      nodes_result_node[U("nodes")] = json::value::array();

      int i = 0;
      for(auto& iter : nodes_json[U("items")].as_array())
      {
        ucout << "node!" << endl;
        auto& node = iter.as_object();
        const auto& nName = node[U("status")].as_object()[U("nodeInfo")].as_object().find(U("machineID"));
        if (nName == node[U("status")].as_object()[U("nodeInfo")].as_object().end())
        {
          throw web::json::json_exception(U("name key not found"));
        }
        nodes_result_node[U("nodes")][i][U("id")] = nName->second;
      }
    }
    else
    {
      // Node data is null, we hit an error.
      nodes_result_node[U("nodes")] = json::value::object();
      nodes_result_node[U("nodes")][U("error")] =
        json::value::string(U("no nodes!"));
    }

    return nodes_result_node;
  }).then([=](pplx::task<json::value> t)
  {
    return handle_exception(t, U("nodes"));
  });
}

int main(int argc, char *argv[])
{
  utility::string_t port = U("8080");
  utility::string_t address = U("http://localhost:");
  address.append(port);

  http::uri base_uri = http::uri(address);

  pplx::task<json::value> t = GetNodes(base_uri.to_string(), "");

  t.wait();

  json::value jval = t.get();
  ucout << jval.to_string() << endl;

  return 0;
}

