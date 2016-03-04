include_directories(${Firmament_SOURCE_DIR} ${Firmament_BINARY_DIR}/src
                    ${Firmament_BINARY_DIR}/third_party/gtest/src/gtest/include
                    ${Firmament_BINARY_DIR}/third_party/thread-safe-stl-containers/src/thread-safe-stl-containers)
link_directories(${Firmament_BINARY_DIR}/src/)

add_executable(test_integration
  test_integration.cc
  )

add_dependencies(test_integration firmament)
add_definitions(-D__PLATFORM_HAS_BOOST__)

target_link_libraries(test_integration ctemplate gflags glog boost_system
  boost_filesystem boost_thread protobuf firmament_scheduling)
