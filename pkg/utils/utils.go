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

package utils

import (
	"strings"
	"os"

	log "github.com/Sirupsen/logrus"
)

// GetLogger returns logrus log of given logLevel
func GetLogger(logLevel string) {
	if strings.EqualFold(logLevel, "debug") {
		log.SetLevel(log.DebugLevel)
	} else if strings.EqualFold(logLevel, "info") {
		log.SetLevel(log.InfoLevel)
	} else {
		// Default level
		log.SetLevel(log.WarnLevel)
	}

	log.SetOutput(os.Stderr)
}

// GetContextLogger creates a logger with common fields
func GetContextLogger(workload string) *log.Entry {
	contextLogger := log.WithFields(log.Fields{
		"Workload": workload,
	})

	return contextLogger
}