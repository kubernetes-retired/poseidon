project(Poseidon)

cmake_minimum_required(VERSION 2.8)

set(Poseidon_ROOT_DIR ${CMAKE_SOURCE_DIR})
set(Poseidon_SOURCE_DIR ${CMAKE_SOURCE_DIR}/src)
set(Poseidon_BINARY_DIR ${CMAKE_SOURCE_DIR}/build)
set(CMAKE_MODULE_PATH ${Poseidon_ROOT_DIR}/cmake)
set(CMAKE_RUNTIME_OUTPUT_DIRECTORY ${Poseidon_BINARY_DIR})

# CMake libraries
include(CheckLibraryExists)
include(CheckFunctionExists)
include(ExternalProject)
include(CheckCXXSourceCompiles)
include(CheckCXXCompilerFlag)

# Build options
include(BuildOptions)

# Compiler config
include(CXXCompilerConfig)

# External/3rd party libraries
include(ExternalDependencies)

# Specify include directories
include_directories(${Poseidon_SOURCE_DIR}/src/include)
include_directories(${cpp-rest-sdk_SOURCE_DIR}/Release/include)

# TODO(malte): not needed?
#link_directories(third_party/ms-cpprestsdk/src/cpp-rest-sdk-build/Binaries)

# Set subdirectories
subdirs(src/apiclient)
subdirs(src/firmament)
