// Copyright 2026 Cody Eding
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package collector

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestCollectControllerStats(t *testing.T) {
	data, err := os.ReadFile("testdata/controller-statistics.json")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	c := NewME5Collector(&mockClient{
		responses: map[string][]byte{"/show/controller-statistics": data},
	}, onlyEnabled(CollectorControllerStats))

	expected := `
# HELP me5_controller_bytes_per_second The data transfer rate, in bytes per second, calculated over the interval since these statistics were last requested or reset. This value will be zero if it has not been requested or reset since a controller restart.
# TYPE me5_controller_bytes_per_second gauge
me5_controller_bytes_per_second{id="controller_A"} 1.544183808e+09
me5_controller_bytes_per_second{id="controller_B"} 0
# HELP me5_controller_cpu_load_percent The percentage of time the CPU is busy, from 0 to 100.
# TYPE me5_controller_cpu_load_percent gauge
me5_controller_cpu_load_percent{id="controller_A"} 15
me5_controller_cpu_load_percent{id="controller_B"} 1
# HELP me5_controller_data_read_bytes_total The amount of data read since these statistics were last reset or since the controller was restarted.
# TYPE me5_controller_data_read_bytes_total counter
me5_controller_data_read_bytes_total{id="controller_A"} 1.21646556022784e+15
me5_controller_data_read_bytes_total{id="controller_B"} 3.5461449216e+10
# HELP me5_controller_data_written_bytes_total The amount of data written since these statistics were last reset or since the controller was restarted.
# TYPE me5_controller_data_written_bytes_total counter
me5_controller_data_written_bytes_total{id="controller_A"} 2.12449485156352e+15
me5_controller_data_written_bytes_total{id="controller_B"} 1.199013888e+09
# HELP me5_controller_iops The input/output operations per second, calculated over the interval since these statistics were last requested or reset. This value will be zero if it has not been requested or reset since a controller restart.
# TYPE me5_controller_iops gauge
me5_controller_iops{id="controller_A"} 5388
me5_controller_iops{id="controller_B"} 0
# HELP me5_controller_num_forwarded_cmds_total The current count of commands that are being forwarded or are queued to be forwarded to the partner controller for processing. This value will be zero if no commands are being forwarded or are queued to be forwarded.
# TYPE me5_controller_num_forwarded_cmds_total counter
me5_controller_num_forwarded_cmds_total{id="controller_A"} 0
me5_controller_num_forwarded_cmds_total{id="controller_B"} 0
# HELP me5_controller_power_on_hours_total The total amount of hours the controller has been powered on in its life time.
# TYPE me5_controller_power_on_hours_total counter
me5_controller_power_on_hours_total{id="controller_A"} 16157.74
me5_controller_power_on_hours_total{id="controller_B"} 16157.55
# HELP me5_controller_power_on_seconds_total The number of seconds since the controller was restarted.
# TYPE me5_controller_power_on_seconds_total counter
me5_controller_power_on_seconds_total{id="controller_A"} 7.690913e+06
me5_controller_power_on_seconds_total{id="controller_B"} 7.690886e+06
# HELP me5_controller_read_cache_hits_total For the controller that owns the volume, the number of times the block to be read is found in cache.
# TYPE me5_controller_read_cache_hits_total counter
me5_controller_read_cache_hits_total{id="controller_A"} 6.7641732982e+10
me5_controller_read_cache_hits_total{id="controller_B"} 0
# HELP me5_controller_read_cache_misses_total For the controller that owns the volume, the number of times the block to be read is not found in cache.
# TYPE me5_controller_read_cache_misses_total counter
me5_controller_read_cache_misses_total{id="controller_A"} 1.0137927851e+10
me5_controller_read_cache_misses_total{id="controller_B"} 0
# HELP me5_controller_reads_total The number of read operations since these statistics were last reset or since the controller was restarted.
# TYPE me5_controller_reads_total counter
me5_controller_reads_total{id="controller_A"} 4.726480591e+09
me5_controller_reads_total{id="controller_B"} 1.156238e+06
# HELP me5_controller_write_cache_hits_total For the controller that owns the volume, the number of times the block written to is found in cache.
# TYPE me5_controller_write_cache_hits_total counter
me5_controller_write_cache_hits_total{id="controller_A"} 7.821133655e+10
me5_controller_write_cache_hits_total{id="controller_B"} 0
# HELP me5_controller_write_cache_misses_total For the controller that owns the volume, the number of times the block written to is not found in cache.
# TYPE me5_controller_write_cache_misses_total counter
me5_controller_write_cache_misses_total{id="controller_A"} 1.29594701558e+11
me5_controller_write_cache_misses_total{id="controller_B"} 0
# HELP me5_controller_write_cache_used_percent Percentage of write cache in use, from 0 to 100.
# TYPE me5_controller_write_cache_used_percent gauge
me5_controller_write_cache_used_percent{id="controller_A"} 24
me5_controller_write_cache_used_percent{id="controller_B"} 0
# HELP me5_controller_writes_total The number of write operations since these statistics were last reset or since the controller was restarted.
# TYPE me5_controller_writes_total counter
me5_controller_writes_total{id="controller_A"} 3.291733005e+10
me5_controller_writes_total{id="controller_B"} 38410
`

	if err := testutil.CollectAndCompare(c, strings.NewReader(expected),
		"me5_controller_bytes_per_second",
		"me5_controller_cpu_load_percent",
		"me5_controller_data_read_bytes_total",
		"me5_controller_data_written_bytes_total",
		"me5_controller_iops",
		"me5_controller_num_forwarded_cmds_total",
		"me5_controller_power_on_hours_total",
		"me5_controller_power_on_seconds_total",
		"me5_controller_read_cache_hits_total",
		"me5_controller_read_cache_misses_total",
		"me5_controller_reads_total",
		"me5_controller_write_cache_hits_total",
		"me5_controller_write_cache_misses_total",
		"me5_controller_write_cache_used_percent",
		"me5_controller_writes_total",
	); err != nil {
		t.Error(err)
	}
}

func TestCollectControllerStats_APIError(t *testing.T) {
	c := NewME5Collector(&mockClient{
		errs: map[string]error{"/show/controller-statistics": errors.New("connection refused")},
	}, onlyEnabled(CollectorControllerStats))

	ch := make(chan prometheus.Metric, 10)
	if err := c.CollectControllerStats(context.Background(), ch); err == nil {
		t.Error("expected error, got nil")
	}
	if len(ch) != 0 {
		t.Errorf("expected no metrics on error, got %d", len(ch))
	}
}
