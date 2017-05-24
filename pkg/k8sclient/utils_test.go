// Poseidon
// Copyright (c) The Poseidon Authors.
// All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// THIS CODE IS PROVIDED ON AN *AS IS* BASIS, WITHOUT WARRANTIES OR
// CONDITIONS OF ANY KIND, EITHER EXPRESS OR IMPLIED, INCLUDING WITHOUT
// LIMITATION ANY IMPLIED WARRANTIES OR CONDITIONS OF TITLE, FITNESS FOR
// A PARTICULAR PURPOSE, MERCHANTABLITY OR NON-INFRINGEMENT.
//
// See the Apache Version 2.0 License for specific language governing
// permissions and limitations under the License.

package k8sclient

import "testing"

// TestGenerateUUID tests for GenerateUUID() for passed string
func TestGenerateUUID(t *testing.T) {
	seed := "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
	u1 := GenerateUUID(seed)
	t.Log("Returned UUID:", u1)
}

// TestHashCombine tests for HashCombine() function
func TestHashCombine(t *testing.T) {
	juid := "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
	juid2 := "6ba7b8109dad11d180b400c04fd430c8"
	u1 := HashCombine(juid, juid2)
	t.Log("Returned Hash combine:", u1)
}
