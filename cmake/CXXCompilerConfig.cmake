# CMake module for flag checking
include(CheckCXXCompilerFlag)

# Debug/Release mode toggle
if (_DEBUG)
  set(CMAKE_BUILD_TYPE Debug)
endif (_DEBUG)

# We require C++11 support
CHECK_CXX_COMPILER_FLAG("-std=c++11" COMPILER_SUPPORTS_CXX11)
if (COMPILER_SUPPORTS_CXX11)
  set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -std=c++11")
else (COMPILER_SUPPORTS_CXX11)
  message(
    FATAL_ERROR
    "The compiler ${CMAKE_CXX_COMPILER} does not support the `-std=c++11` "
    "flag for C++11 standards-compliant code. Please use a different C++ "
    "compiler.")
endif (COMPILER_SUPPORTS_CXX11)

# Optimization flags
if (${CMAKE_BUILD_TYPE} Debug)
  set(CMAKE_CXX_OPTFLAGS "-O3")
else (${CMAKE_BUILD_TYPE} Debug)
  set(CMAKE_CXX_OPTFLAGS "-O0 -g")
endif (${CMAKE_BUILD_TYPE} Debug)

# Shared compiler flags used by all builds
set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} ${CMAKE_CXX_OPTFLAGS} -fPIC -Wall -Wextra -pedantic")
# Ignore some non-fatal errors that occur in benign settings
set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -Wno-long-long -Wno-variadic-macros -Wno-deprecated -Wno-vla -Wno-unused-parameter")
# Downgrade some issues to warnings
set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -Wno-error=unused-parameter -Wno-error=unused-function")

# Compiler-specific flags
if (CMAKE_CXX_COMPILER_ID MATCHES "Clang")
  # using clang
  set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -Wno-error=language-extension-token")
else()
  # other compilers, usually g++
endif()
