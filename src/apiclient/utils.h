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

#ifndef POSEIDON_APICLIENT_UTILS_H
#define POSEIDON_APICLIENT_UTILS_H

#include <string>

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

const uint64_t MB_TO_KB = 1024;
const uint64_t GB_TO_KB = 1024 * 1024;
const double MCPU_TO_CPU_UNITS = 1000.0;

struct NodeStatistics {
  string hostname_;
  bool is_ready_;
  bool is_out_of_disk_;
  double cpu_capacity_;
  double cpu_allocatable_;
  uint64_t memory_capacity_kb_;
  uint64_t memory_allocatable_kb_;
};

struct PodStatistics {
  string name_;
  string state_;
  string controller_id_;
  double cpu_request_;
  uint64_t memory_request_kb_;
};

http_response PrintHTTPResponse(const string& url,
                                const http_response& response);

pplx::task<json::value> HandleTaskException(
    pplx::task<json::value>& task,
    const utility::string_t& field_name);

double StringRequestToCPU(string request);
uint64_t StringRequestToKB(string memory_request);

}  // namespace apiclient
}  // namespace poseidon

#endif  // POSEIDON_APICLIENT_UTILS_H
