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

#include <gtest/gtest.h>

#include "firmament/scheduler_bridge.h"

DEFINE_string(listen_uri, "", "");

namespace poseidon {

class SchedulerBridgeTest : public ::testing::Test {
 protected:
  // You can remove any or all of the following functions if its body is empty.

  SchedulerBridgeTest() {
    // You can do set-up work for each test here.
    FLAGS_v = 3;
  }

  virtual ~SchedulerBridgeTest() {
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

TEST_F(SchedulerBridgeTest, CreateTopLevelResourceTest) {
  SchedulerBridge scheduler_bridge;
  CHECK_NOTNULL(scheduler_bridge.CreateTopLevelResource());
}

}  // namespace poseidon

int main(int argc, char **argv) {
  ::testing::InitGoogleTest(&argc, argv);
  return RUN_ALL_TESTS();
}
