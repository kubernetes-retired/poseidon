project(Poseidon)

cmake_minimum_required(VERSION 2.8)

set(CMAKE_MODULE_PATH ${Poseidon_SOURCE_DIR}/cmake)

include(CheckLibraryExists)
include(CheckFunctionExists)
include(ExternalProject)
include(CheckCXXSourceCompiles)

# Set include path
include_directories(${Poseidon_SOURCE_DIR}/src/include)

# Shared compiler flags used by all builds
set(CMAKE_CXX_FLAGS "-O3 -fPIC -std=c++11 -Wall -Wextra -Wno-long-long -Wno-variadic-macros -Wno-deprecated -Wno-vla -Wno-unused-parameter -Wno-error=unused-parameter -Wno-error=unused-function")

# Compiler-specific flags
if (CMAKE_CXX_COMPILER_ID MATCHES "Clang")
  # using clang
  set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -Wno-error=language-extension-token")
else()
  # other compilers, usually g++
endif()

# Add Swagger dependency
ExternalProject_Add(
    swagger-codegen
    GIT_REPOSITORY https://github.com/swagger-api/swagger-codegen
    GIT_TAG v2.1.4
    TIMEOUT 10
    PREFIX ${CMAKE_CURRENT_BINARY_DIR}/third_party/swagger-codegen
    # no configure required
    CONFIGURE_COMMAND ""
    BUILD_COMMAND mvn package
    BUILD_IN_SOURCE 1
    # no installation required
    INSTALL_COMMAND ""
    # Wrap download, configure and build steps in a script to log output
    LOG_DOWNLOAD ON
    LOG_BUILD ON
    LOG_INSTALL ON)

# Add MS CPP REST SDK dependency
ExternalProject_Add(
    cpp-rest-sdk
    GIT_REPOSITORY https://github.com/Microsoft/cpprestsdk
    GIT_TAG v2.7.0
    TIMEOUT 10
    PREFIX ${CMAKE_CURRENT_BINARY_DIR}/third_party/ms-cpprestsdk
    # N.B.: *must* use g++ here, the build fails with CXX=clang++!
    CONFIGURE_COMMAND CXX=g++-4.8 cmake ../cpp-rest-sdk/Release -DCMAKE_BUILD_TYPE=Release
    BUILD_COMMAND make
    # no installation required
    INSTALL_COMMAND ""
    # Wrap download, configure and build steps in a script to log output
    LOG_DOWNLOAD ON
    LOG_BUILD ON)

#
# Specify include dir
#ExternalProject_Get_Property(cpp-rest-sdk source_dir)
include_directories(third_party/ms-cpprestsdk/src/cpp-rest-sdk/Release/include)
#link_directories(third_party/ms-cpprestsdk/src/cpp-rest-sdk-build/Binaries)

# Set subdirectories
subdirs(src/apiclient)
