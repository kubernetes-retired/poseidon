#include <iostream>
#include <streambuf>
#include <sstream>
#include <fstream>

#include <map>
#include <vector>
#include <string>
#include <exception>

#include "cpprest/http_client.h"

#define iequals(x, y) boost::iequals((x), (y))

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

int main(int argc, char *argv[])
{
    utility::string_t port = U("8080");
    utility::string_t address = U("http://localhost:");
    address.append(port);

    http::uri base_uri = http::uri(address);
    http::uri full_uri = http::uri_builder(base_uri).append_path(U("/api/v1")).to_uri();

    http_client apiClient(full_uri);

    utility::ostringstream_t buf;

    buf << U("?foo=bar&baz=");

    PrintResponse(full_uri.to_string(), apiClient.request(methods::GET, buf.str()).get());

    return 0;
}

