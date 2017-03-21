/*
 * Poseidon
 * Copyright (c) The Poseidon Authors.
 * All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * THIS CODE IS PROVIDED ON AN *AS IS* BASIS, WITHOUT WARRANTIES OR
 * CONDITIONS OF ANY KIND, EITHER EXPRESS OR IMPLIED, INCLUDING WITHOUT
 * LIMITATION ANY IMPLIED WARRANTIES OR CONDITIONS OF TITLE, FITNESS FOR
 * A PARTICULAR PURPOSE, MERCHANTABLITY OR NON-INFRINGEMENT.
 *
 * See the Apache Version 2.0 License for specific language governing
 * permissions and limitations under the License.
 */

#include <glog/logging.h>
#include <gflags/gflags.h>
#include <gtest/gtest.h>

#include "apiclient/k8s_api_client.h"

namespace poseidon {
namespace apiclient {

class K8sApiClientTest : public ::testing::Test {
 protected:
  // You can remove any or all of the following functions if its body is empty.

  K8sApiClientTest() {
    // You can do set-up work for each test here.
    FLAGS_v = 3;
  }

  virtual ~K8sApiClientTest() {
    // You can do clean-up work that doesn't throw exceptions here.
  }

  // If the constructor and destructor are not enough for setting up
  // and cleaning up each test, you can define the following methods:
  virtual void SetUp() {
    // Code here will be called immediately after the constructor (right
    // before each test).
  }

  virtual void TearDown() {
    // Code here will be called immediately after each test (right
    // before the destructor).
  }

  // Objects declared here can be used by all tests in the test case for
  // LocalExecutor.
};

TEST_F(K8sApiClientTest, PodsWithLabelTest) {
  K8sApiClient api_client;
  vector<PodStatistics> all_pods = api_client.PodsWithLabel("");
  CHECK_EQ(all_pods.size(), 0);
}

}  // namespace apiclient
}  // namespace poseidon

int main(int argc, char **argv) {
  ::testing::InitGoogleTest(&argc, argv);
  return RUN_ALL_TESTS();
}
