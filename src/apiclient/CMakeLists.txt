add_executable(demo
  demo.cpp
  )

target_link_libraries(demo m dl crypto ssl boost_system boost_thread
  ${cpp-rest-sdk_BINARY_DIR}/libcpprest.so)
