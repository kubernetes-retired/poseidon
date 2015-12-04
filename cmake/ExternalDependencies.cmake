# Add Swagger dependency
ExternalProject_Add(
    swagger-codegen
    GIT_REPOSITORY https://github.com/swagger-api/swagger-codegen
    GIT_TAG v2.1.4
    TIMEOUT 10
    PREFIX ${Poseidon_ROOT_DIR}/third_party/swagger-codegen
    # no configure required
    CONFIGURE_COMMAND ""
    # This invokes Maven for building the Java code
    BUILD_COMMAND mvn compile
    BUILD_IN_SOURCE 1
    # Installation is to generate the Maven package
    INSTALL_COMMAND mvn package
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
    PREFIX ${Poseidon_ROOT_DIR}/third_party/ms-cpprestsdk
    # N.B.: *must* use g++ here, the build fails with CXX=clang++!
    CONFIGURE_COMMAND CXX=g++-4.8 cmake ../cpp-rest-sdk/Release -DCMAKE_BUILD_TYPE=Release
    BUILD_COMMAND make
    # no installation required
    INSTALL_COMMAND ""
    # Wrap download, configure and build steps in a script to log output
    LOG_DOWNLOAD ON
    LOG_BUILD ON)

ExternalProject_Get_Property(cpp-rest-sdk SOURCE_DIR)
ExternalProject_Get_Property(cpp-rest-sdk BINARY_DIR)
set(cpp-rest-sdk_SOURCE_DIR ${SOURCE_DIR})
set(cpp-rest-sdk_BINARY_DIR ${BINARY_DIR}/Binaries)
