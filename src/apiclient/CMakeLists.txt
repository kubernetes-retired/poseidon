add_executable(demo
  demo.cpp
  )

target_link_libraries(demo dl crypto ssl boost_system ${cpp-rest-sdk_BINARY_DIR}/libcpprest.so)
