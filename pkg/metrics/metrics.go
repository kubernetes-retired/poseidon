/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package metrics

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const schedulerSubsystem = "poseidon"

// All the histogram based metrics have 1ms as size for the smallest bucket.
var (
	SchedulingSubmitmLatency = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Subsystem: schedulerSubsystem,
			Name:      "scheduling_submit_latency_microseconds",
			Help:      "Scheduling submit latency",
			Buckets:   prometheus.ExponentialBuckets(1000, 2, 15),
		},
	)
	BindingLatency = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Subsystem: schedulerSubsystem,
			Name:      "binding_latency_microseconds",
			Help:      "Binding latency",
			Buckets:   prometheus.ExponentialBuckets(1000, 2, 15),
		},
	)
	SchedulingPremptionEvaluationDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Subsystem: schedulerSubsystem,
			Name:      "scheduling_preemption_evaluation",
			Help:      "Scheduling preemption evaluation duration",
			Buckets:   prometheus.ExponentialBuckets(1000, 2, 15),
		},
	)
	PreemptionVictims = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Subsystem: schedulerSubsystem,
			Name:      "pod_preemption_victims",
			Help:      "Number of selected preemption victims",
		})
	PreemptionAttempts = prometheus.NewCounter(
		prometheus.CounterOpts{
			Subsystem: schedulerSubsystem,
			Name:      "total_preemption_attempts",
			Help:      "Total preemption attempts in the cluster till now",
		})
)

var registerMetrics sync.Once

// Register all metrics.
func Register() {
	// Register the metrics.
	registerMetrics.Do(func() {
		prometheus.MustRegister(SchedulingSubmitmLatency)
		prometheus.MustRegister(BindingLatency)
		prometheus.MustRegister(SchedulingPremptionEvaluationDuration)
		prometheus.MustRegister(PreemptionVictims)
		prometheus.MustRegister(PreemptionAttempts)
	})
}

// SinceInMicroseconds gets the time since the specified start in microseconds.
func SinceInMicroseconds(start time.Time) float64 {
	return float64(time.Since(start).Nanoseconds() / time.Microsecond.Nanoseconds())
}
