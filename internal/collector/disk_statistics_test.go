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

func TestCollectDiskStats(t *testing.T) {
	data, err := os.ReadFile("testdata/disk-statistics.json")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	c := NewME5Collector(&mockClient{
		responses: map[string][]byte{"/show/disk-statistics": data},
	}, onlyEnabled(CollectorDiskStats))

	expected := `
# HELP me5_disk_bad_blocks_total Total bad block count by controller path.
# TYPE me5_disk_bad_blocks_total counter
me5_disk_bad_blocks_total{location="0.0",path="1",serial="S0DUMMY0X000001"} 0
me5_disk_bad_blocks_total{location="0.0",path="2",serial="S0DUMMY0X000001"} 0
me5_disk_bad_blocks_total{location="0.1",path="1",serial="S0DUMMY0X000002"} 0
me5_disk_bad_blocks_total{location="0.1",path="2",serial="S0DUMMY0X000002"} 0
# HELP me5_disk_block_reassigns_total Total block reassignment count by controller path.
# TYPE me5_disk_block_reassigns_total counter
me5_disk_block_reassigns_total{location="0.0",path="1",serial="S0DUMMY0X000001"} 0
me5_disk_block_reassigns_total{location="0.0",path="2",serial="S0DUMMY0X000001"} 0
me5_disk_block_reassigns_total{location="0.1",path="1",serial="S0DUMMY0X000002"} 0
me5_disk_block_reassigns_total{location="0.1",path="2",serial="S0DUMMY0X000002"} 0
# HELP me5_disk_bytes_per_second The data transfer rate, in bytes per second, calculated over the interval since these statistics were last requested or reset. This value will be zero if it has not been requested or reset since a controller restart.
# TYPE me5_disk_bytes_per_second gauge
me5_disk_bytes_per_second{location="0.0",serial="S0DUMMY0X000001"} 6.8210688e+07
me5_disk_bytes_per_second{location="0.1",serial="S0DUMMY0X000002"} 7.2601088e+07
# HELP me5_disk_data_read_bytes_total Amount of data read since these statistics were last reset or since the controller was restarted, in bytes.
# TYPE me5_disk_data_read_bytes_total counter
me5_disk_data_read_bytes_total{location="0.0",serial="S0DUMMY0X000001"} 1.29817375425536e+14
me5_disk_data_read_bytes_total{location="0.1",serial="S0DUMMY0X000002"} 1.30940117951488e+14
# HELP me5_disk_data_written_bytes_total Amount of data written since these statistics were last reset or since the controller was restarted, in bytes.
# TYPE me5_disk_data_written_bytes_total counter
me5_disk_data_written_bytes_total{location="0.0",serial="S0DUMMY0X000001"} 9.0478084816896e+13
me5_disk_data_written_bytes_total{location="0.1",serial="S0DUMMY0X000002"} 9.0194983936e+13
# HELP me5_disk_io_timeout_count_total Total I/O timeout count by controller path.
# TYPE me5_disk_io_timeout_count_total counter
me5_disk_io_timeout_count_total{location="0.0",path="1",serial="S0DUMMY0X000001"} 0
me5_disk_io_timeout_count_total{location="0.0",path="2",serial="S0DUMMY0X000001"} 0
me5_disk_io_timeout_count_total{location="0.1",path="1",serial="S0DUMMY0X000002"} 0
me5_disk_io_timeout_count_total{location="0.1",path="2",serial="S0DUMMY0X000002"} 0
# HELP me5_disk_iops Input/output operations per second, calculated over the interval since these statistics were last requested or reset. This value will be zero if it has not been requested or reset since a controller restart.
# TYPE me5_disk_iops gauge
me5_disk_iops{location="0.0",serial="S0DUMMY0X000001"} 263
me5_disk_iops{location="0.1",serial="S0DUMMY0X000002"} 279
# HELP me5_disk_lifetime_data_read_bytes_total The amount of data read from the disk in its lifetime, in bytes.
# TYPE me5_disk_lifetime_data_read_bytes_total counter
me5_disk_lifetime_data_read_bytes_total{location="0.0",serial="S0DUMMY0X000001"} 4.93562850784768e+14
me5_disk_lifetime_data_read_bytes_total{location="0.1",serial="S0DUMMY0X000002"} 4.93984439021056e+14
# HELP me5_disk_lifetime_data_written_bytes_total The amount of data written to the disk in its lifetime, in bytes.
# TYPE me5_disk_lifetime_data_written_bytes_total counter
me5_disk_lifetime_data_written_bytes_total{location="0.0",serial="S0DUMMY0X000001"} 1.04776264498688e+14
me5_disk_lifetime_data_written_bytes_total{location="0.1",serial="S0DUMMY0X000002"} 1.0439850306304e+14
# HELP me5_disk_media_errors_total Total media error count by controller path.
# TYPE me5_disk_media_errors_total counter
me5_disk_media_errors_total{location="0.0",path="1",serial="S0DUMMY0X000001"} 0
me5_disk_media_errors_total{location="0.0",path="2",serial="S0DUMMY0X000001"} 0
me5_disk_media_errors_total{location="0.1",path="1",serial="S0DUMMY0X000002"} 0
me5_disk_media_errors_total{location="0.1",path="2",serial="S0DUMMY0X000002"} 0
# HELP me5_disk_no_response_count_total Total no-response count by controller path.
# TYPE me5_disk_no_response_count_total counter
me5_disk_no_response_count_total{location="0.0",path="1",serial="S0DUMMY0X000001"} 0
me5_disk_no_response_count_total{location="0.0",path="2",serial="S0DUMMY0X000001"} 0
me5_disk_no_response_count_total{location="0.1",path="1",serial="S0DUMMY0X000002"} 0
me5_disk_no_response_count_total{location="0.1",path="2",serial="S0DUMMY0X000002"} 0
# HELP me5_disk_nonmedia_errors_total Total non-media error count by controller path.
# TYPE me5_disk_nonmedia_errors_total counter
me5_disk_nonmedia_errors_total{location="0.0",path="1",serial="S0DUMMY0X000001"} 2
me5_disk_nonmedia_errors_total{location="0.0",path="2",serial="S0DUMMY0X000001"} 2
me5_disk_nonmedia_errors_total{location="0.1",path="1",serial="S0DUMMY0X000002"} 2
me5_disk_nonmedia_errors_total{location="0.1",path="2",serial="S0DUMMY0X000002"} 2
# HELP me5_disk_power_on_hours_total The total number of hours that the disk has been powered on since it was manufactured. This value is stored in disk metadata and is updated in 30- minute increments.
# TYPE me5_disk_power_on_hours_total counter
me5_disk_power_on_hours_total{location="0.0",serial="S0DUMMY0X000001"} 16172
me5_disk_power_on_hours_total{location="0.1",serial="S0DUMMY0X000002"} 16173
# HELP me5_disk_queue_depth Number of pending I/O operations currently being serviced.
# TYPE me5_disk_queue_depth gauge
me5_disk_queue_depth{location="0.0",serial="S0DUMMY0X000001"} 15
me5_disk_queue_depth{location="0.1",serial="S0DUMMY0X000002"} 0
# HELP me5_disk_reads_total Number of read operations since these statistics were last reset or since the controller was restarted.
# TYPE me5_disk_reads_total counter
me5_disk_reads_total{location="0.0",serial="S0DUMMY0X000001"} 6.60144488e+08
me5_disk_reads_total{location="0.1",serial="S0DUMMY0X000002"} 6.6437533e+08
# HELP me5_disk_smart_count SMART event count by controller path.
# TYPE me5_disk_smart_count gauge
me5_disk_smart_count{location="0.0",path="1",serial="S0DUMMY0X000001"} 0
me5_disk_smart_count{location="0.0",path="2",serial="S0DUMMY0X000001"} 0
me5_disk_smart_count{location="0.1",path="1",serial="S0DUMMY0X000002"} 0
me5_disk_smart_count{location="0.1",path="2",serial="S0DUMMY0X000002"} 0
# HELP me5_disk_spinup_retry_count_total Total spinup retry count by controller path.
# TYPE me5_disk_spinup_retry_count_total counter
me5_disk_spinup_retry_count_total{location="0.0",path="1",serial="S0DUMMY0X000001"} 0
me5_disk_spinup_retry_count_total{location="0.0",path="2",serial="S0DUMMY0X000001"} 0
me5_disk_spinup_retry_count_total{location="0.1",path="1",serial="S0DUMMY0X000002"} 0
me5_disk_spinup_retry_count_total{location="0.1",path="2",serial="S0DUMMY0X000002"} 0
# HELP me5_disk_writes_total Number of write operations since these statistics were last reset or since the controller was restarted.
# TYPE me5_disk_writes_total counter
me5_disk_writes_total{location="0.0",serial="S0DUMMY0X000001"} 4.05433383e+08
me5_disk_writes_total{location="0.1",serial="S0DUMMY0X000002"} 4.03991116e+08
`

	if err := testutil.CollectAndCompare(c, strings.NewReader(expected),
		"me5_disk_bad_blocks_total",
		"me5_disk_block_reassigns_total",
		"me5_disk_bytes_per_second",
		"me5_disk_data_read_bytes_total",
		"me5_disk_data_written_bytes_total",
		"me5_disk_io_timeout_count_total",
		"me5_disk_iops",
		"me5_disk_lifetime_data_read_bytes_total",
		"me5_disk_lifetime_data_written_bytes_total",
		"me5_disk_media_errors_total",
		"me5_disk_no_response_count_total",
		"me5_disk_nonmedia_errors_total",
		"me5_disk_power_on_hours_total",
		"me5_disk_queue_depth",
		"me5_disk_reads_total",
		"me5_disk_smart_count",
		"me5_disk_spinup_retry_count_total",
		"me5_disk_writes_total",
	); err != nil {
		t.Error(err)
	}
}

func TestCollectDiskStats_APIError(t *testing.T) {
	c := NewME5Collector(&mockClient{
		errs: map[string]error{"/show/disk-statistics": errors.New("connection refused")},
	}, onlyEnabled(CollectorDiskStats))

	ch := make(chan prometheus.Metric, 10)
	if err := c.CollectDiskStats(context.Background(), ch); err == nil {
		t.Error("expected error, got nil")
	}
	if len(ch) != 0 {
		t.Errorf("expected no metrics on error, got %d", len(ch))
	}
}
