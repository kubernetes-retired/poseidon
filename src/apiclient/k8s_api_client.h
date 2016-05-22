#ifndef POSEIDON_APICLIENT_K8S_API_CLIENT_H
#define POSEIDON_APICLIENT_K8S_API_CLIENT_H

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

class K8sApiClient {
 public:
  K8sApiClient();
  vector<pair<string, string>> AllNodes(void);
  vector<string> AllPods(void);
  vector<pair<string, string>> NodesWithLabel(const string& label);
  vector<string> PodsWithLabel(const string& label);

 private:
  pplx::task<json::value> GetNodesTask(
      const utility::string_t& base_uri,
      const utility::string_t& label_selector);
  pplx::task<json::value> GetPodsTask(
      const utility::string_t& base_uri,
      const utility::string_t& label_selector);

  // API server URI
  web::uri base_uri_;
};

}  // namespace apiclient
}  // namespace poseidon

#endif  // POSEIDON_APICLIENT_K8S_API_CLIENT_H
