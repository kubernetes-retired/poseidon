include_directories(${Firmament_SOURCE_DIR} ${Firmament_BINARY_DIR}/src
                    ${Firmament_BINARY_DIR}/third_party/gtest/src/gtest/include
                    ${Firmament_BINARY_DIR}/third_party/thread-safe-stl-containers/src/thread-safe-stl-containers)
link_directories(${Firmament_BINARY_DIR}/src/)

add_executable(poseidon
  apiclient/k8s_api_client.cc
  apiclient/utils.cc
  firmament/scheduler_integration.cc
  )

add_dependencies(poseidon firmament)
add_definitions(-D__PLATFORM_HAS_BOOST__)

target_link_libraries(poseidon
  boost_filesystem boost_system boost_thread
  crypto ctemplate gflags glog ${protobuf3_LIBRARY} ssl
  firmament_scheduling
  ${cpp-rest-sdk_BINARY_DIR}/libcpprest.so)
