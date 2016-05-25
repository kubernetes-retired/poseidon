# Add Swagger dependency
#ExternalProject_Add(
#    swagger-codegen
#    GIT_REPOSITORY https://github.com/swagger-api/swagger-codegen
#    GIT_TAG v2.1.4
#    TIMEOUT 10
#    PREFIX ${Poseidon_ROOT_DIR}/third_party/swagger-codegen
#   # no configure required
#    CONFIGURE_COMMAND ""
#    # This invokes Maven for building the Java code
#    BUILD_COMMAND mvn compile
#    BUILD_IN_SOURCE 1
#    # Installation is to generate the Maven package
#    INSTALL_COMMAND mvn package
#    # Wrap download, configure and build steps in a script to log output
#    LOG_DOWNLOAD ON
#    LOG_BUILD ON
#    LOG_INSTALL ON)

# Add MS CPP REST SDK dependency
ExternalProject_Add(
    cpp-rest-sdk
    GIT_REPOSITORY https://github.com/Microsoft/cpprestsdk
    GIT_TAG v2.7.0
    TIMEOUT 10
    PREFIX ${CMAKE_CURRENT_BINARY_DIR}/third_party/ms-cpprestsdk
    # N.B.: *must* use g++ here, the build fails with CXX=clang++!
    CONFIGURE_COMMAND cmake ../cpp-rest-sdk/Release -DCMAKE_BUILD_TYPE=Release -DBUILD_SAMPLES=false -DBUILD_TESTS=false
    BUILD_COMMAND make
    # no installation required
    INSTALL_COMMAND "")

ExternalProject_Get_Property(cpp-rest-sdk SOURCE_DIR)
ExternalProject_Get_Property(cpp-rest-sdk BINARY_DIR)
set(cpp-rest-sdk_SOURCE_DIR ${SOURCE_DIR})
set(cpp-rest-sdk_INCLUDE_DIR ${SOURCE_DIR}/Release/include)
set(cpp-rest-sdk_BINARY_DIR ${BINARY_DIR}/Binaries)

# Add Google Test dependency
ExternalProject_Add(
    gtest
    GIT_REPOSITORY https://github.com/google/googletest.git
    TIMEOUT 10
    PREFIX ${CMAKE_CURRENT_BINARY_DIR}/third_party/gtest
    # no install required, we link the library from the build tree
    INSTALL_COMMAND "")

ExternalProject_Get_Property(gtest BINARY_DIR)
ExternalProject_Get_Property(gtest SOURCE_DIR)
set(gtest_BINARY_DIR ${BINARY_DIR})
set(gtest_SOURCE_DIR ${SOURCE_DIR})
set(gtest_INCLUDE_DIR ${gtest_SOURCE_DIR}/googletest/include)
include_directories(${gtest_INCLUDE_DIR})
set(gtest_LIBRARY ${gtest_BINARY_DIR}/googlemock/gtest/libgtest.a)
set(gtest_MAIN_LIBRARY ${gtest_BINARY_DIR}/googlemock/gtest/libgtest_main.a)

set(gmock_INCLUDE_DIR ${gtest_SOURCE_DIR}/googlemock/include)
include_directories(${gmock_INCLUDE_DIR})
set(gmock_LIBRARY ${gtest_BINARY_DIR}/googlemock/libgmock.a)
set(gmock_MAIN_LIBRARY ${gtest_BINARY_DIR}/googlemock/libgmock_main.a)

# Add Firmament as a dependency
ExternalProject_Add(
    firmament
    GIT_REPOSITORY https://github.com/ms705/firmament
    TIMEOUT 10
    PREFIX ${CMAKE_CURRENT_BINARY_DIR}/firmament
    CMAKE_ARGS -DHTTP_UI=off -DENABLE_HDFS=off
    # N.B.: only build the integration library target
    BUILD_COMMAND make firmament_scheduling
    # no installation required
    INSTALL_COMMAND "")

ExternalProject_Get_Property(firmament SOURCE_DIR)
ExternalProject_Get_Property(firmament BINARY_DIR)
set(Firmament_ROOT_DIR ${SOURCE_DIR})
set(Firmament_SOURCE_DIR ${Firmament_ROOT_DIR}/src)
set(Firmament_BINARY_DIR ${BINARY_DIR})
