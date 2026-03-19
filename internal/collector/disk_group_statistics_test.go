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

func TestCollectDiskGroupStats(t *testing.T) {
	data, err := os.ReadFile("testdata/disk-group-statistics.json")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	c := NewME5Collector(&mockClient{
		responses: map[string][]byte{"/show/disk-group-statistics": data},
	}, onlyEnabled(CollectorDiskGroupStats))

	expected := `
# HELP me5_disk_group_avg_read_rsp_time_microseconds Average response time in microseconds for all read operations, calculated over the interval since these statistics were last requested or reset.
# TYPE me5_disk_group_avg_read_rsp_time_microseconds gauge
me5_disk_group_avg_read_rsp_time_microseconds{name="dgA01",serial="00c0fffa39f700001295066800000000"} 4439
# HELP me5_disk_group_avg_rsp_time_microseconds Average response time in microseconds for read and write operations, calculated over the interval since these statistics were last requested or reset.
# TYPE me5_disk_group_avg_rsp_time_microseconds gauge
me5_disk_group_avg_rsp_time_microseconds{name="dgA01",serial="00c0fffa39f700001295066800000000"} 4439
# HELP me5_disk_group_avg_write_rsp_time_microseconds Average response time in microseconds for all write operations, calculated over the interval since these statistics were last requested or reset.
# TYPE me5_disk_group_avg_write_rsp_time_microseconds gauge
me5_disk_group_avg_write_rsp_time_microseconds{name="dgA01",serial="00c0fffa39f700001295066800000000"} 4682
# HELP me5_disk_group_bytes_per_second The data transfer rate, in bytes per second, calculated over the interval since these statistics were last requested or reset. This value will be zero if it has not been requested or reset since a controller restart.
# TYPE me5_disk_group_bytes_per_second gauge
me5_disk_group_bytes_per_second{name="dgA01",serial="00c0fffa39f700001295066800000000"} 1.785451008e+09
# HELP me5_disk_group_data_read_bytes_total Amount of data read since these statistics were last reset or since the controller was restarted, in bytes.
# TYPE me5_disk_group_data_read_bytes_total counter
me5_disk_group_data_read_bytes_total{name="dgA01",serial="00c0fffa39f700001295066800000000"} 1.446882860496384e+15
# HELP me5_disk_group_data_written_bytes_total Amount of data written since these statistics were last reset or since the controller was restarted, in bytes.
# TYPE me5_disk_group_data_written_bytes_total counter
me5_disk_group_data_written_bytes_total{name="dgA01",serial="00c0fffa39f700001295066800000000"} 2.199975823384576e+15
# HELP me5_disk_group_iops Input/output operations per second, calculated over the interval since these statistics were last requested or reset. This value will be zero if it has not been requested or reset since a controller restart.
# TYPE me5_disk_group_iops gauge
me5_disk_group_iops{name="dgA01",serial="00c0fffa39f700001295066800000000"} 508
# HELP me5_disk_group_pages_allocated_per_minute The rate, in pages per minute, at which pages are allocated to volumes in the disk group because they need more space to store data.
# TYPE me5_disk_group_pages_allocated_per_minute gauge
me5_disk_group_pages_allocated_per_minute{name="dgA01",serial="00c0fffa39f700001295066800000000"} 0
# HELP me5_disk_group_pages_deallocated_per_minute The rate, in pages per minute, at which pages are deallocated from volumes in the disk group because they no longer need the space to store data.
# TYPE me5_disk_group_pages_deallocated_per_minute gauge
me5_disk_group_pages_deallocated_per_minute{name="dgA01",serial="00c0fffa39f700001295066800000000"} 0
# HELP me5_disk_group_reads_total Number of read operations since these statistics were last reset or since the controller was restarted.
# TYPE me5_disk_group_reads_total counter
me5_disk_group_reads_total{name="dgA01",serial="00c0fffa39f700001295066800000000"} 3.388180281e+09
# HELP me5_disk_group_writes_total Number of write operations since these statistics were last reset or since the controller was restarted.
# TYPE me5_disk_group_writes_total counter
me5_disk_group_writes_total{name="dgA01",serial="00c0fffa39f700001295066800000000"} 1.146491286e+09
`

	if err := testutil.CollectAndCompare(c, strings.NewReader(expected),
		"me5_disk_group_avg_read_rsp_time_microseconds",
		"me5_disk_group_avg_rsp_time_microseconds",
		"me5_disk_group_avg_write_rsp_time_microseconds",
		"me5_disk_group_bytes_per_second",
		"me5_disk_group_data_read_bytes_total",
		"me5_disk_group_data_written_bytes_total",
		"me5_disk_group_iops",
		"me5_disk_group_pages_allocated_per_minute",
		"me5_disk_group_pages_deallocated_per_minute",
		"me5_disk_group_reads_total",
		"me5_disk_group_writes_total",
	); err != nil {
		t.Error(err)
	}
}

func TestCollectDiskGroupStats_APIError(t *testing.T) {
	c := NewME5Collector(&mockClient{
		errs: map[string]error{"/show/disk-group-statistics": errors.New("connection refused")},
	}, onlyEnabled(CollectorDiskGroupStats))

	ch := make(chan prometheus.Metric, 10)
	if err := c.CollectDiskGroupStats(context.Background(), ch); err == nil {
		t.Error("expected error, got nil")
	}
	if len(ch) != 0 {
		t.Errorf("expected no metrics on error, got %d", len(ch))
	}
}
