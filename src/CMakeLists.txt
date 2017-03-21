include_directories(${Firmament_SOURCE_DIR} ${Firmament_BINARY_DIR}/src
                    ${Firmament_BINARY_DIR}/third_party/gtest/src/gtest/include
                    ${Firmament_BINARY_DIR}/third_party/thread-safe-stl-containers/src/thread-safe-stl-containers)
link_directories(${Firmament_BINARY_DIR}/src/)

set(FIRMAMENT_SRCS
  firmament/knowledge_base_populator.cc
  firmament/scheduler_bridge.cc
  )

set(API_CLIENT_SRCS
  apiclient/k8s_api_client.cc
  apiclient/utils.cc
  )

add_executable(poseidon
  firmament/scheduler_integration.cc
  ${API_CLIENT_SRCS}
  ${FIRMAMENT_SRCS}
  )

add_dependencies(poseidon firmament)
add_definitions(-D__PLATFORM_HAS_BOOST__)

add_library(apiclient OBJECT ${API_CLIENT_SRCS})
add_dependencies(apiclient gflags glog ${cpp-rest-sdk_BINARY_DIR}/libcpprest.so)

add_library(firmament_lib OBJECT ${FIRMAMENT_SRCS})
add_dependencies(firmament_lib boost_filesystem boost_system boost_thread
  crypto ctemplate gflags glog ${protobuf3_LIBRARY} ssl
  firmament_scheduling ${cpp-rest-sdk_BINARY_DIR}/libcpprest.so)

target_link_libraries(poseidon
  boost_filesystem boost_system boost_thread
  crypto ctemplate gflags glog ${protobuf3_LIBRARY} ssl
  firmament_scheduling
  ${cpp-rest-sdk_BINARY_DIR}/libcpprest.so)

set(API_CLIENT_TESTS
  apiclient/k8s_api_client_test.cc
  )
set(FIRMAMENT_TESTS
  firmament/scheduler_bridge_test.cc
)

###############################################################################
# Unit tests

if (BUILD_TESTS)
  foreach(T IN ITEMS ${API_CLIENT_TESTS})
    get_filename_component(TEST_NAME ${T} NAME_WE)
    add_executable(src/${TEST_NAME} ${T}
      $<TARGET_OBJECTS:apiclient>)
    target_link_libraries(src/${TEST_NAME}
      boost_system boost_thread pthread
      crypto gflags glog ${protobuf3_LIBRARY} ssl
      ${gmock_LIBRARY} ${gmock_MAIN_LIBRARY}
      ${gtest_LIBRARY} ${gtest_MAIN_LIBRARY}
      ${cpp-rest-sdk_BINARY_DIR}/libcpprest.so)
    add_test(${TEST_NAME} ${TEST_NAME})
  endforeach(T)
  foreach(T IN ITEMS ${FIRMAMENT_TESTS})
    get_filename_component(TEST_NAME ${T} NAME_WE)
    add_executable(src/${TEST_NAME} ${T}
      $<TARGET_OBJECTS:apiclient>
      $<TARGET_OBJECTS:firmament_lib>)
    target_link_libraries(src/${TEST_NAME}
      boost_filesystem boost_system boost_thread pthread
      crypto ctemplate gflags glog ${protobuf3_LIBRARY} ssl
      firmament_scheduling
      ${gmock_LIBRARY} ${gmock_MAIN_LIBRARY}
      ${gtest_LIBRARY} ${gtest_MAIN_LIBRARY}
      ${cpp-rest-sdk_BINARY_DIR}/libcpprest.so)
    add_test(${TEST_NAME} ${TEST_NAME})
  endforeach(T)
endif (BUILD_TESTS)
