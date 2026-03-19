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

func TestCollectPoolStats(t *testing.T) {
	data, err := os.ReadFile("testdata/pool-statistics.json")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	c := NewME5Collector(&mockClient{
		responses: map[string][]byte{"/show/pool-statistics": data},
	}, onlyEnabled(CollectorPoolStats))

	expected := `
# HELP me5_pool_avg_read_rsp_time_microseconds The average response time, in microseconds, for read operations since the last sampling time.
# TYPE me5_pool_avg_read_rsp_time_microseconds gauge
me5_pool_avg_read_rsp_time_microseconds{pool="A",serial="00c0fffa39f700001395066801000000"} 1966
# HELP me5_pool_avg_rsp_time_microseconds The average response time, in microseconds, for read and write operations since the last sampling time.
# TYPE me5_pool_avg_rsp_time_microseconds gauge
me5_pool_avg_rsp_time_microseconds{pool="A",serial="00c0fffa39f700001395066801000000"} 12001
# HELP me5_pool_avg_write_rsp_time_microseconds The average response time, in microseconds, for write operations since the last sampling time.
# TYPE me5_pool_avg_write_rsp_time_microseconds gauge
me5_pool_avg_write_rsp_time_microseconds{pool="A",serial="00c0fffa39f700001395066801000000"} 12983
# HELP me5_pool_bytes_per_second The data transfer rate, in bytes per second, calculated over the interval since these statistics were last requested or reset. This value will be zero if it has not been requested or reset since a controller restart.
# TYPE me5_pool_bytes_per_second gauge
me5_pool_bytes_per_second{pool="A",serial="00c0fffa39f700001395066801000000"} 7.34060544e+08
# HELP me5_pool_data_read_bytes_total The amount of data read since these statistics were last reset or since the controller was restarted.
# TYPE me5_pool_data_read_bytes_total counter
me5_pool_data_read_bytes_total{pool="A",serial="00c0fffa39f700001395066801000000"} 1.446909605729792e+15
# HELP me5_pool_data_written_bytes_total The amount of data written since these statistics were last reset or since the controller was restarted.
# TYPE me5_pool_data_written_bytes_total counter
me5_pool_data_written_bytes_total{pool="A",serial="00c0fffa39f700001395066801000000"} 2.200189004988416e+15
# HELP me5_pool_iops The number of input/output operations per second, calculated over the interval since these statistics were last requested or reset. This value will be zero if it has not been requested or reset since a controller restart.
# TYPE me5_pool_iops gauge
me5_pool_iops{pool="A",serial="00c0fffa39f700001395066801000000"} 192
# HELP me5_pool_num_cold_page_moves The number of 'cold' pages promoted from lower tiers to higher tiers since statistics were last reset.
# TYPE me5_pool_num_cold_page_moves counter
me5_pool_num_cold_page_moves{pool="A",serial="00c0fffa39f700001395066801000000"} 0
# HELP me5_pool_num_hot_page_moves The number of 'hot' pages promoted from lower tiers to higher tiers since statistics were last reset.
# TYPE me5_pool_num_hot_page_moves counter
me5_pool_num_hot_page_moves{pool="A",serial="00c0fffa39f700001395066801000000"} 0
# HELP me5_pool_pages_alloc_per_minute The rate, in pages per minute, at which pages are allocated to volumes in the pool because they need more space to store data.
# TYPE me5_pool_pages_alloc_per_minute gauge
me5_pool_pages_alloc_per_minute{pool="A",serial="00c0fffa39f700001395066801000000"} 0
# HELP me5_pool_pages_dealloc_per_minute The rate, in pages per minute, at which pages are deallocated from volumes in the pool because they no longer need the space to store data.
# TYPE me5_pool_pages_dealloc_per_minute gauge
me5_pool_pages_dealloc_per_minute{pool="A",serial="00c0fffa39f700001395066801000000"} 0
# HELP me5_pool_pages_unmap_per_minute The number of 4 MB pages that host systems have unmapped per minute, through use of the SCSI UNMAP command, to free storage space as a result of deleting files or formatting volumes on the host.
# TYPE me5_pool_pages_unmap_per_minute gauge
me5_pool_pages_unmap_per_minute{pool="A",serial="00c0fffa39f700001395066801000000"} 0
# HELP me5_pool_reads_total The number of read operations since these statistics were last reset or since the controller was restarted.
# TYPE me5_pool_reads_total counter
me5_pool_reads_total{pool="A",serial="00c0fffa39f700001395066801000000"} 3.388193652e+09
# HELP me5_pool_writes_total The number of write operations since these statistics were last reset or since the controller was restarted.
# TYPE me5_pool_writes_total counter
me5_pool_writes_total{pool="A",serial="00c0fffa39f700001395066801000000"} 1.146544637e+09
`

	if err := testutil.CollectAndCompare(c, strings.NewReader(expected),
		"me5_pool_avg_read_rsp_time_microseconds",
		"me5_pool_avg_rsp_time_microseconds",
		"me5_pool_avg_write_rsp_time_microseconds",
		"me5_pool_bytes_per_second",
		"me5_pool_data_read_bytes_total",
		"me5_pool_data_written_bytes_total",
		"me5_pool_iops",
		"me5_pool_num_cold_page_moves",
		"me5_pool_num_hot_page_moves",
		"me5_pool_pages_alloc_per_minute",
		"me5_pool_pages_dealloc_per_minute",
		"me5_pool_pages_unmap_per_minute",
		"me5_pool_reads_total",
		"me5_pool_writes_total",
	); err != nil {
		t.Error(err)
	}
}

func TestCollectPoolStats_APIError(t *testing.T) {
	c := NewME5Collector(&mockClient{
		errs: map[string]error{"/show/pool-statistics": errors.New("connection refused")},
	}, onlyEnabled(CollectorPoolStats))

	ch := make(chan prometheus.Metric, 10)
	if err := c.CollectPoolStats(context.Background(), ch); err == nil {
		t.Error("expected error, got nil")
	}
	if len(ch) != 0 {
		t.Errorf("expected no metrics on error, got %d", len(ch))
	}
}
