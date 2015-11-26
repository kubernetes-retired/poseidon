add_executable(demo
  demo.cpp
  )

target_link_libraries(demo dl crypto ssl boost_system ${Poseidon_SOURCE_DIR}/third_party/ms-cpprestsdk/src/cpp-rest-sdk-build/Binaries/libcpprest.so)
