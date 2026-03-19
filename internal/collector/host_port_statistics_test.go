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

func TestCollectHostPortStats(t *testing.T) {
	data, err := os.ReadFile("testdata/host-port-statistics.json")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	c := NewME5Collector(&mockClient{
		responses: map[string][]byte{"/show/host-port-statistics": data},
	}, onlyEnabled(CollectorHostPortStats))

	expected := `
# HELP me5_host_port_avg_rsp_time_microseconds Average response time in microseconds for read and write operations, calculated over the interval since these statistics were last requested or reset.
# TYPE me5_host_port_avg_rsp_time_microseconds gauge
me5_host_port_avg_rsp_time_microseconds{port="hostport_A0"} 0
me5_host_port_avg_rsp_time_microseconds{port="hostport_A1"} 0
me5_host_port_avg_rsp_time_microseconds{port="hostport_A2"} 0
me5_host_port_avg_rsp_time_microseconds{port="hostport_A3"} 3748
me5_host_port_avg_rsp_time_microseconds{port="hostport_B0"} 0
me5_host_port_avg_rsp_time_microseconds{port="hostport_B1"} 0
me5_host_port_avg_rsp_time_microseconds{port="hostport_B2"} 0
me5_host_port_avg_rsp_time_microseconds{port="hostport_B3"} 0
# HELP me5_host_port_bytes_per_second The data transfer rate, in bytes per second, calculated over the interval since these statistics were last requested or reset. This value will be zero if it has not been requested or reset since a controller restart.
# TYPE me5_host_port_bytes_per_second gauge
me5_host_port_bytes_per_second{port="hostport_A0"} 0
me5_host_port_bytes_per_second{port="hostport_A1"} 0
me5_host_port_bytes_per_second{port="hostport_A2"} 0
me5_host_port_bytes_per_second{port="hostport_A3"} 1.693449728e+09
me5_host_port_bytes_per_second{port="hostport_B0"} 0
me5_host_port_bytes_per_second{port="hostport_B1"} 0
me5_host_port_bytes_per_second{port="hostport_B2"} 0
me5_host_port_bytes_per_second{port="hostport_B3"} 0
# HELP me5_host_port_data_read_bytes_total Amount of data read since these statistics were last reset or since the controller was restarted.
# TYPE me5_host_port_data_read_bytes_total counter
me5_host_port_data_read_bytes_total{port="hostport_A0"} 0
me5_host_port_data_read_bytes_total{port="hostport_A1"} 0
me5_host_port_data_read_bytes_total{port="hostport_A2"} 0
me5_host_port_data_read_bytes_total{port="hostport_A3"} 1.21646556022784e+15
me5_host_port_data_read_bytes_total{port="hostport_B0"} 0
me5_host_port_data_read_bytes_total{port="hostport_B1"} 0
me5_host_port_data_read_bytes_total{port="hostport_B2"} 0
me5_host_port_data_read_bytes_total{port="hostport_B3"} 3.5461449216e+10
# HELP me5_host_port_data_written_bytes_total Amount of data written since these statistics were last reset or since the controller was restarted.
# TYPE me5_host_port_data_written_bytes_total counter
me5_host_port_data_written_bytes_total{port="hostport_A0"} 0
me5_host_port_data_written_bytes_total{port="hostport_A1"} 0
me5_host_port_data_written_bytes_total{port="hostport_A2"} 0
me5_host_port_data_written_bytes_total{port="hostport_A3"} 2.124499708350464e+15
me5_host_port_data_written_bytes_total{port="hostport_B0"} 0
me5_host_port_data_written_bytes_total{port="hostport_B1"} 0
me5_host_port_data_written_bytes_total{port="hostport_B2"} 0
me5_host_port_data_written_bytes_total{port="hostport_B3"} 1.199013888e+09
# HELP me5_host_port_iops Input/output operations per second, calculated over the interval since these statistics were last requested or reset. This value will be zero if it has not been requested or reset since a controller restart.
# TYPE me5_host_port_iops gauge
me5_host_port_iops{port="hostport_A0"} 0
me5_host_port_iops{port="hostport_A1"} 0
me5_host_port_iops{port="hostport_A2"} 0
me5_host_port_iops{port="hostport_A3"} 10533
me5_host_port_iops{port="hostport_B0"} 0
me5_host_port_iops{port="hostport_B1"} 0
me5_host_port_iops{port="hostport_B2"} 0
me5_host_port_iops{port="hostport_B3"} 0
# HELP me5_host_port_queue_depth The number of pending I/O operations currently being serviced.
# TYPE me5_host_port_queue_depth gauge
me5_host_port_queue_depth{port="hostport_A0"} 0
me5_host_port_queue_depth{port="hostport_A1"} 0
me5_host_port_queue_depth{port="hostport_A2"} 0
me5_host_port_queue_depth{port="hostport_A3"} 30
me5_host_port_queue_depth{port="hostport_B0"} 0
me5_host_port_queue_depth{port="hostport_B1"} 0
me5_host_port_queue_depth{port="hostport_B2"} 0
me5_host_port_queue_depth{port="hostport_B3"} 0
# HELP me5_host_port_reads_total Number of read operations since these statistics were last reset or since the controller was restarted.
# TYPE me5_host_port_reads_total counter
me5_host_port_reads_total{port="hostport_A0"} 0
me5_host_port_reads_total{port="hostport_A1"} 0
me5_host_port_reads_total{port="hostport_A2"} 0
me5_host_port_reads_total{port="hostport_A3"} 4.726480591e+09
me5_host_port_reads_total{port="hostport_B0"} 0
me5_host_port_reads_total{port="hostport_B1"} 0
me5_host_port_reads_total{port="hostport_B2"} 0
me5_host_port_reads_total{port="hostport_B3"} 1.156238e+06
# HELP me5_host_port_writes_total Number of write operations since these statistics were last reset or since the controller was restarted.
# TYPE me5_host_port_writes_total counter
me5_host_port_writes_total{port="hostport_A0"} 0
me5_host_port_writes_total{port="hostport_A1"} 0
me5_host_port_writes_total{port="hostport_A2"} 0
me5_host_port_writes_total{port="hostport_A3"} 3.2917348112e+10
me5_host_port_writes_total{port="hostport_B0"} 0
me5_host_port_writes_total{port="hostport_B1"} 0
me5_host_port_writes_total{port="hostport_B2"} 0
me5_host_port_writes_total{port="hostport_B3"} 38410
`

	if err := testutil.CollectAndCompare(c, strings.NewReader(expected),
		"me5_host_port_avg_rsp_time_microseconds",
		"me5_host_port_bytes_per_second",
		"me5_host_port_data_read_bytes_total",
		"me5_host_port_data_written_bytes_total",
		"me5_host_port_iops",
		"me5_host_port_queue_depth",
		"me5_host_port_reads_total",
		"me5_host_port_writes_total",
	); err != nil {
		t.Error(err)
	}
}

func TestCollectHostPortStats_APIError(t *testing.T) {
	c := NewME5Collector(&mockClient{
		errs: map[string]error{"/show/host-port-statistics": errors.New("connection refused")},
	}, onlyEnabled(CollectorHostPortStats))

	ch := make(chan prometheus.Metric, 10)
	if err := c.CollectHostPortStats(context.Background(), ch); err == nil {
		t.Error("expected error, got nil")
	}
	if len(ch) != 0 {
		t.Errorf("expected no metrics on error, got %d", len(ch))
	}
}
