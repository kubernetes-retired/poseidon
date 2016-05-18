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

http_response PrintHTTPResponse(const string& url,
                                const http_response& response);

pplx::task<json::value> HandleTaskException(
    pplx::task<json::value>& task,
    const utility::string_t& field_name);

}  // namespace apiclient
}  // namespace poseidon

#endif  // POSEIDON_APICLIENT_UTILS_H
