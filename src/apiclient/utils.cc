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

http_response PrintHTTPResponse(const std::string &url,
                                const http_response &response) {
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
  } catch (const std::exception& ex) {
    json::value error_json = json::value::object();
    error_json[field_name] = json::value::object();
    error_json[field_name][U("error")] =
      json::value::string(utility::conversions::to_string_t(ex.what()));
    return pplx::task_from_result<json::value>(error_json);
  }

  return task;
}

}  // namespace apiclient
}  // namespace poseidon
