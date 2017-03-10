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

#include "apiclient/utils.h"

#include <string>
#include <glog/logging.h>

#include "cpprest/http_client.h"
#include "cpprest/json.h"

using namespace std;
using namespace web;
using namespace json;
using namespace utility;
using namespace http;
using namespace http::client;

namespace poseidon {
namespace apiclient {

http_response PrintHTTPResponse(const string& url,
                                const http_response& response) {
    ucout << "Got response for '" << url << "': "
          << response.extract_json().get().serialize()
          << endl;
    return response;
}

// In case of any failure to fetch data from a service,
// create a json object with "error" key and value containing the exception
// details.
pplx::task<json::value> HandleTaskException(
    pplx::task<json::value>& task,
    const utility::string_t& field_name) {
  try {
    task.get();
  } catch (const exception& ex) {
    json::value error_json = json::value::object();
    error_json[field_name] = json::value::object();
    error_json[field_name][U("error")] =
      json::value(utility::conversions::to_string_t(ex.what()));
    return pplx::task_from_result<json::value>(error_json);
  }

  return task;
}

double StringRequestToCPU(string request) {
  if (request[request.size() - 1] == 'm') {
    return stod(request.substr(0, request.size() - 1)) / MCPU_TO_CPU_UNITS;
  } else {
    return stod(request);
  }
}

uint64_t StringRequestToKB(string memory_request) {
  uint64_t request =
    stoull(memory_request.substr(0, memory_request.size() - 2));
  string units =
    memory_request.substr(memory_request.size() - 2, memory_request.size());
  if (units == "Ki") {
    // Already in KB.
  } else if (units == "Mi") {
    request *= MB_TO_KB;
  } else if (units == "Gi") {
    request *= GB_TO_KB;
  } else {
    LOG(ERROR) << "Unexpected memory request: " << memory_request;
  }
  return request;
}

}  // namespace apiclient
}  // namespace poseidon
